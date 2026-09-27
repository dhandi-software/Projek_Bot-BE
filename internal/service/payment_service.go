package service

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"bot_be/internal/config"
	"bot_be/internal/model"
	"bot_be/internal/wshub"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"gorm.io/gorm"
)

type CheckoutItemRequest struct {
	ProductID uint    `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
}

type CheckoutCustomerRequest struct {
	CustomerID uint   `json:"customer_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
}

type CreateCheckoutRequest struct {
	IdempotencyKey string                  `json:"idempotency_key"`
	Items          []CheckoutItemRequest   `json:"items"`
	Customer       CheckoutCustomerRequest `json:"customer"`
}

type CheckoutResponse struct {
	OrderID         string  `json:"order_id"`
	SnapToken       string  `json:"snap_token"`
	SnapRedirectURL string  `json:"snap_redirect_url"`
	TotalAmount     float64 `json:"total_amount"`
	IsReused        bool    `json:"is_reused"`
}

type PaymentService interface {
	CreateCheckoutTransaction(req CreateCheckoutRequest) (*CheckoutResponse, error)
	HandleNotification(payload map[string]interface{}) error
	GetOrders() ([]model.Order, error)
	GetOrderByID(orderID string) (*model.Order, error)
}

type paymentService struct {
	cfg        *config.Config
	db         *gorm.DB
	snapClient snap.Client
}

func NewPaymentService(cfg *config.Config, db *gorm.DB) PaymentService {
	var s snap.Client
	env := midtrans.Sandbox
	if cfg.MidtransIsProduction {
		env = midtrans.Production
	}

	s.New(cfg.MidtransServerKey, env)

	return &paymentService{
		cfg:        cfg,
		db:         db,
		snapClient: s,
	}
}

func (s *paymentService) CreateCheckoutTransaction(req CreateCheckoutRequest) (*CheckoutResponse, error) {
	req.IdempotencyKey = strings.TrimSpace(req.IdempotencyKey)
	if req.IdempotencyKey == "" {
		return nil, errors.New("idempotency_key wajib diisi untuk mencegah transaksi ganda")
	}

	if len(req.Items) == 0 {
		return nil, errors.New("keranjang belanja tidak boleh kosong")
	}

	var existingOrder model.Order
	err := s.db.Where("idempotency_key = ?", req.IdempotencyKey).Preload("OrderItems").First(&existingOrder).Error
	if err == nil {
		return &CheckoutResponse{
			OrderID:         existingOrder.OrderID,
			SnapToken:       existingOrder.SnapToken,
			SnapRedirectURL: existingOrder.SnapRedirectURL,
			TotalAmount:     existingOrder.TotalAmount,
			IsReused:        true,
		}, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("gagal mengecek idempotency key: %w", err)
	}

	productIDs := make([]uint, 0, len(req.Items))
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("jumlah produk ID %d tidak valid", item.ProductID)
		}
		if item.ProductID > 0 {
			productIDs = append(productIDs, item.ProductID)
		}
	}

	var products []model.Product
	if len(productIDs) > 0 {
		_ = s.db.Where("id IN ?", productIDs).Find(&products).Error
	}

	productMap := make(map[uint]model.Product)
	for _, prod := range products {
		productMap[prod.ID] = prod
	}

	var totalAmount float64
	var orderItems []model.OrderItem
	var midtransItems []midtrans.ItemDetails

	for _, item := range req.Items {
		qty := item.Quantity
		prod, exists := productMap[item.ProductID]

		var title string
		var price float64
		var prodID uint

		if exists {
			prodID = prod.ID
			title = prod.Title
			price = prod.Price
			if prod.DiscountPrice > 0 && prod.DiscountPrice < prod.Price {
				price = prod.DiscountPrice
			}
			if prod.Stock < qty {
				return nil, fmt.Errorf("stok produk '%s' tidak mencukupi (tersedia: %d, diminta: %d)", prod.Title, prod.Stock, qty)
			}
		} else {
			prodID = item.ProductID
			if prodID == 0 {
				prodID = 1
			}
			title = strings.TrimSpace(item.Title)
			if title == "" {
				title = fmt.Sprintf("Produk #%d", prodID)
			}
			price = item.Price
			if price <= 0 {
				price = 100000
			}
		}

		itemTotal := price * float64(qty)
		totalAmount += itemTotal

		orderItems = append(orderItems, model.OrderItem{
			ProductID: prodID,
			Title:     title,
			Quantity:  qty,
			Price:     price,
		})

		midtransItems = append(midtransItems, midtrans.ItemDetails{
			ID:    fmt.Sprintf("PROD-%d", prodID),
			Name:  truncateString(title, 50),
			Price: int64(price),
			Qty:   int32(qty),
		})
	}

	var itemsSum int64
	for _, mItem := range midtransItems {
		itemsSum += mItem.Price * int64(mItem.Qty)
	}
	totalAmount = float64(itemsSum)

	orderID := fmt.Sprintf("ORDER-%d-%s", time.Now().UnixNano(), req.IdempotencyKey[:minInt(8, len(req.IdempotencyKey))])

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: itemsSum,
		},
		Items: &midtransItems,
		CustomerDetail: &midtrans.CustomerDetails{
			FName: req.Customer.Name,
			Email: req.Customer.Email,
			Phone: req.Customer.Phone,
		},
	}

	snapResp, snapErr := s.snapClient.CreateTransaction(snapReq)
	if snapErr != nil {
		if snapErr.StatusCode == 401 || strings.Contains(snapErr.Message, "Access denied") || strings.Contains(snapErr.Message, "unauthorized") {
			return nil, errors.New("Midtrans Server Key tidak valid atau belum terdaftar di Midtrans Dashboard (HTTP 401 Unauthorized). Silakan periksa MIDTRANS_SERVER_KEY di file .env")
		}
		errMsg := snapErr.Message
		if errMsg == "" && snapErr.RawError != nil {
			errMsg = snapErr.RawError.Error()
		}
		if errMsg == "" {
			errMsg = snapErr.Error()
		}
		return nil, fmt.Errorf("gagal membuat transaksi di Midtrans: %s", errMsg)
	}

	order := model.Order{
		OrderID:         orderID,
		OrderNumber:     orderID,
		IdempotencyKey:  req.IdempotencyKey,
		CustomerID:      req.Customer.CustomerID,
		UserID:          req.Customer.CustomerID,
		CustomerName:    req.Customer.Name,
		CustomerEmail:   req.Customer.Email,
		CustomerPhone:   req.Customer.Phone,
		ShippingAddress: req.Customer.Address,
		TotalAmount:     totalAmount,
		TotalPrice:      totalAmount,
		Status:          "pending",
		SnapToken:       snapResp.Token,
		SnapRedirectURL: snapResp.RedirectURL,
		OrderItems:      orderItems,
	}

	tx := s.db.Begin()
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("gagal menyimpan order ke database: %w", err)
	}
	tx.Commit()

	return &CheckoutResponse{
		OrderID:         orderID,
		SnapToken:       snapResp.Token,
		SnapRedirectURL: snapResp.RedirectURL,
		TotalAmount:     totalAmount,
		IsReused:        false,
	}, nil
}

func (s *paymentService) HandleNotification(payload map[string]interface{}) error {
	orderID, _ := payload["order_id"].(string)
	transactionID, _ := payload["transaction_id"].(string)
	statusCode, _ := payload["status_code"].(string)
	grossAmountStr, _ := payload["gross_amount"].(string)
	signatureKey, _ := payload["signature_key"].(string)
	transactionStatus, _ := payload["transaction_status"].(string)
	fraudStatus, _ := payload["fraud_status"].(string)
	paymentType, _ := payload["payment_type"].(string)

	log.Printf("[MIDTRANS] notification received")
	log.Printf("[MIDTRANS] order_id: %s | transaction_id: %s | transaction_status: %s | status_code: %s | gross_amount: %s",
		orderID, transactionID, transactionStatus, statusCode, grossAmountStr)

	// Midtrans Dashboard Test Ping or empty test payload handler
	if orderID == "" || strings.HasPrefix(strings.ToLower(orderID), "test") || strings.Contains(strings.ToLower(orderID), "dummy") {
		log.Printf("[MIDTRANS] Notifikasi tes Midtrans diterima (order_id: %s)", orderID)
		return nil
	}

	if statusCode == "" || grossAmountStr == "" || signatureKey == "" {
		log.Printf("[MIDTRANS] Notifikasi Midtrans parsial/tes diterima (order_id: %s)", orderID)
		return nil
	}

	rawSignature := orderID + statusCode + grossAmountStr + s.cfg.MidtransServerKey
	hasher := sha512.New()
	hasher.Write([]byte(rawSignature))
	expectedSignature := hex.EncodeToString(hasher.Sum(nil))

	if !strings.EqualFold(signatureKey, expectedSignature) {
		log.Printf("[MIDTRANS] WARNING: Signature key mismatch untuk order_id: %s", orderID)
		return nil
	}

	var order model.Order
	if err := s.db.Where("order_id = ?", orderID).Preload("OrderItems").First(&order).Error; err != nil {
		log.Printf("[MIDTRANS] ERROR: order_id %s tidak ditemukan di database: %v", orderID, err)
		return nil
	}

	log.Printf("[ORDER] matched internal order: ID=%d, OrderID=%s, CurrentStatus=%s", order.ID, order.OrderID, order.Status)

	if order.Status == "paid" || order.Status == "settlement" {
		log.Printf("[MIDTRANS] Notifikasi diabaikan (Idempotent): order_id %s sudah berstatus '%s'", orderID, order.Status)
		return nil
	}

	isPaid := false
	newStatus := order.Status

	switch transactionStatus {
	case "capture":
		if fraudStatus == "challenge" {
			newStatus = "challenge"
		} else if fraudStatus == "accept" || fraudStatus == "" {
			newStatus = "paid"
			isPaid = true
		}
	case "settlement":
		newStatus = "paid"
		isPaid = true
	case "cancel", "deny", "expire":
		newStatus = "failed"
	case "pending":
		newStatus = "pending"
	}

	tx := s.db.Begin()

	now := time.Now()
	updateFields := map[string]interface{}{
		"status":       newStatus,
		"payment_type": paymentType,
	}

	if isPaid {
		updateFields["paid_at"] = &now

		for _, item := range order.OrderItems {
			if err := tx.Model(&model.Product{}).
				Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
				UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
				tx.Rollback()
				log.Printf("[PAYMENT] ERROR: Gagal mengurangi stok produk ID %d: %v", item.ProductID, err)
				return fmt.Errorf("gagal mengurangi stok produk ID %d: %w", item.ProductID, err)
			}
		}
	}

	if err := tx.Model(&model.Order{}).Where("order_id = ?", orderID).Updates(updateFields).Error; err != nil {
		tx.Rollback()
		log.Printf("[PAYMENT] ERROR: Gagal mengupdate status database untuk order_id %s: %v", orderID, err)
		return fmt.Errorf("gagal mengupdate status order: %w", err)
	}

	logMsg := fmt.Sprintf("Order %s berubah status menjadi %s via %s", orderID, newStatus, paymentType)
	activity := model.ActivityLog{
		Type:        "PAYMENT_UPDATE",
		Sender:      "SYSTEM",
		Description: logMsg,
	}
	_ = tx.Create(&activity).Error

	if err := tx.Commit().Error; err != nil {
		log.Printf("[PAYMENT] ERROR: Gagal commit transaksi database: %v", err)
		return err
	}

	log.Printf("[PAYMENT] status updated: order_id=%s, new_status=%s", orderID, newStatus)

	wshub.BroadcastMessage(map[string]interface{}{
		"event":        "payment_status_updated",
		"order_id":     orderID,
		"status":       newStatus,
		"total_amount": order.TotalAmount,
		"paid_at":      now,
	})

	return nil
}

func (s *paymentService) GetOrders() ([]model.Order, error) {
	var orders []model.Order
	err := s.db.Preload("OrderItems").Order("created_at desc").Find(&orders).Error
	return orders, err
}

func (s *paymentService) GetOrderByID(orderID string) (*model.Order, error) {
	var order model.Order
	err := s.db.Where("order_id = ?", orderID).Preload("OrderItems").First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncateString(s string, maxLen int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= maxLen {
		return string(runes)
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}
