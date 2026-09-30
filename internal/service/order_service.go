package service

import (
	"bot_be/internal/model"

	"gorm.io/gorm"
)

type OrderService interface {
	GetOrders() ([]model.Order, error)
	GetOrderByID(orderID string) (*model.Order, error)
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
