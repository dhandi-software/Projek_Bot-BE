package service

import (
	"fmt"
	"strings"
	"time"

	"bot_be/internal/model"
	"bot_be/internal/wshub"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderService interface {
	GetOrders() ([]model.Order, error)
	GetOrderByID(orderID string) (*model.Order, error)
	CancelOrder(orderID string) (*model.Order, error)
	UpdateOrderStatus(orderID string, status string) (*model.Order, error)
}

type orderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) OrderService {
	return &orderService{db: db}
}

func (s *orderService) GetOrders() ([]model.Order, error) {
	var orders []model.Order
	err := s.db.Preload("OrderItems").Order("created_at desc").Find(&orders).Error
	return orders, err
}

func (s *orderService) GetOrderByID(orderID string) (*model.Order, error) {
	var order model.Order
	err := s.db.Where("order_id = ?", orderID).Preload("OrderItems").First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *orderService) UpdateOrderStatus(orderID string, status string) (*model.Order, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var order model.Order
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_id = ?", orderID).Preload("OrderItems").First(&order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Order %s tidak ditemukan", orderID)
	}

	validStatuses := map[string]bool{
		"pending":     true,
		"paid":        true,
		"packaging":   true,
		"on_the_road": true,
		"shipped":     true,
		"delivered":   true,
		"completed":   true,
		"cancelled":   true,
		"failed":      true,
	}

	statusLower := strings.ToLower(strings.TrimSpace(status))
	if !validStatuses[statusLower] {
		tx.Rollback()
		return nil, fmt.Errorf("Status '%s' tidak valid", status)
	}

	updateFields := map[string]interface{}{
		"status": statusLower,
	}

	if (statusLower == "packaging" || statusLower == "paid" || statusLower == "shipped" || statusLower == "on_the_road" || statusLower == "delivered" || statusLower == "completed") && order.PaidAt == nil {
		now := time.Now()
		updateFields["paid_at"] = &now
	}

	if err := tx.Model(&model.Order{}).Where("order_id = ?", orderID).Updates(updateFields).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Gagal meng-update status order: %w", err)
	}

	order.Status = statusLower

	activity := model.ActivityLog{
		Type:        "ORDER_STATUS_UPDATED",
		Sender:      "ADMIN",
		Description: fmt.Sprintf("Admin memperbarui status order %s menjadi %s", orderID, statusLower),
	}
	_ = tx.Create(&activity).Error

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	wshub.BroadcastMessage(map[string]interface{}{
		"event":    "payment_status_updated",
		"order_id": orderID,
		"status":   statusLower,
	})

	return &order, nil
}

func (s *orderService) CancelOrder(orderID string) (*model.Order, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var order model.Order
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_id = ?", orderID).Preload("OrderItems").First(&order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Order %s tidak ditemukan", orderID)
	}

	statusLower := strings.ToLower(order.Status)
	if statusLower == "paid" || statusLower == "settlement" || statusLower == "completed" || statusLower == "shipped" {
		tx.Rollback()
		return nil, fmt.Errorf("Order %s sudah dibayar/diproses dan tidak dapat dibatalkan", orderID)
	}

	if statusLower == "cancelled" || statusLower == "cancel" || statusLower == "failed" || statusLower == "expire" {
		tx.Rollback()
		return &order, nil
	}

	order.Status = "cancelled"
	if err := tx.Model(&model.Order{}).Where("order_id = ?", orderID).Update("status", "cancelled").Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("Gagal membatalkan order: %w", err)
	}

	for _, item := range order.OrderItems {
		if item.ProductID > 0 {
			_ = tx.Model(&model.Product{}).
				Where("id = ?", item.ProductID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error
		}
	}

	activity := model.ActivityLog{
		Type:        "ORDER_CANCELLED",
		Sender:      "CUSTOMER",
		Description: fmt.Sprintf("Order %s telah dibatalkan oleh pelanggan", orderID),
	}
	_ = tx.Create(&activity).Error

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	wshub.BroadcastMessage(map[string]interface{}{
		"event":    "payment_status_updated",
		"order_id": orderID,
		"status":   "cancelled",
	})

	return &order, nil
}
