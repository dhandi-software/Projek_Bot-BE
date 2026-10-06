package service

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"bot_be/internal/config"
	"bot_be/internal/model"
	"bot_be/internal/wshub"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CheckoutItemRequest struct {
	ProductID uint    `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Image     string  `json:"image"`
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
	PaymentMethod  string                  `json:"payment_method"`
	Bank           string                  `json:"bank"`
	Items          []CheckoutItemRequest   `json:"items"`
	Customer       CheckoutCustomerRequest `json:"customer"`
}

type CheckoutResponse struct {
	OrderID         string  `json:"order_id"`
	SnapToken       string  `json:"snap_token"`
	SnapRedirectURL string  `json:"snap_redirect_url"`
	QRISURL         string  `json:"qris_url,omitempty"`
	QRISString      string  `json:"qris_string,omitempty"`
	VANumber        string  `json:"va_number,omitempty"`
	VABank          string  `json:"va_bank,omitempty"`
	TotalAmount     float64 `json:"total_amount"`
	Status          string  `json:"status"`
	IsReused        bool    `json:"is_reused"`
}

type PaymentService interface {
	CreateCheckoutTransaction(req CreateCheckoutRequest) (*CheckoutResponse, error)
	HandleNotification(payload map[string]interface{}) error
	GetOrders() ([]model.Order, error)
	GetOrderByID(orderID string) (*model.Order, error)
	GenerateInvoicePDF(orderID string) ([]byte, error)
	CancelOrder(orderID string) (*model.Order, error)
}

type paymentService struct {
	cfg            *config.Config
	db             *gorm.DB
	snapClient     snap.Client
	coreClient     coreapi.Client
	orderService   OrderService
	invoiceService InvoiceService
}

func NewPaymentService(cfg *config.Config, db *gorm.DB, orderService OrderService, invoiceService InvoiceService) PaymentService {
	var s snap.Client
	var c coreapi.Client
	env := midtrans.Sandbox
	if cfg.MidtransIsProduction {
		env = midtrans.Production
	}

	s.New(cfg.MidtransServerKey, env)
	c.New(cfg.MidtransServerKey, env)

	if orderService == nil {
		orderService = NewOrderService(db)
	}
	if invoiceService == nil {
		invoiceService = NewInvoiceService(db, orderService)
	}

	return &paymentService{
		cfg:            cfg,
		db:             db,
		snapClient:     s,
		coreClient:     c,
		orderService:   orderService,
		invoiceService: invoiceService,
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

	paymentMethod := strings.ToLower(strings.TrimSpace(req.PaymentMethod))

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
		if err := s.db.Where("id IN ?", productIDs).Find(&products).Error; err != nil {
			return nil, fmt.Errorf("gagal mengambil data produk: %w", err)
		}
	}

	productMap := make(map[uint]model.Product)
	for _, prod := range products {
		productMap[prod.ID] = prod
	}

	var totalAmount float64
	var orderItems []model.OrderItem
	var midtransItems []midtrans.ItemDetails

	suffix := req.IdempotencyKey
	if len(suffix) > 16 {
		suffix = suffix[len(suffix)-16:]
	}
	orderID := fmt.Sprintf("ORD-%d-%s", time.Now().UnixMilli(), suffix)

	for _, item := range req.Items {
		qty := item.Quantity

		var title string
		var price float64
		var prodID uint

		if item.ProductID > 0 {
			prod, exists := productMap[item.ProductID]
			if exists {
				prodID = prod.ID
				title = prod.Title
				price = prod.Price
				if prod.DiscountPrice > 0 && prod.DiscountPrice < prod.Price {
					price = prod.DiscountPrice
				}
			} else {
				var foundProd model.Product
				if err := s.db.Where("LOWER(title) = ?", strings.ToLower(strings.TrimSpace(item.Title))).First(&foundProd).Error; err == nil {
					prodID = foundProd.ID
					title = foundProd.Title
					price = foundProd.Price
					if foundProd.DiscountPrice > 0 && foundProd.DiscountPrice < foundProd.Price {
						price = foundProd.DiscountPrice
					}
				} else {
					prodID = 0
					title = strings.TrimSpace(item.Title)
					if title == "" {
						title = fmt.Sprintf("Produk #%d", item.ProductID)
					}
					price = item.Price
					if price <= 0 {
						price = 100000
					}
				}
			}
		} else {
			title = strings.TrimSpace(item.Title)
			if title == "" {
				title = "Produk Dhandi Ecommerce"
			}
			price = item.Price
			if price <= 0 {
				price = 100000
			}
		}

		var image string
		if prodID > 0 {
			if prod, exists := productMap[prodID]; exists {
				image = prod.Image
			}
		}
		if image == "" && item.Image != "" {
			image = item.Image
		}

		itemTotal := price * float64(qty)
		totalAmount += itemTotal

		orderItems = append(orderItems, model.OrderItem{
			OrderID:   orderID,
			ProductID: prodID,
			Title:     title,
			Quantity:  qty,
			Price:     price,
			Image:     image,
			ImageURL:  image,
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

	var existingOrder model.Order
	err := s.db.Where("idempotency_key = ?", req.IdempotencyKey).Preload("OrderItems").First(&existingOrder).Error
	if err == nil {
		if math.Abs(existingOrder.TotalAmount-totalAmount) < 0.01 && len(existingOrder.OrderItems) == len(req.Items) {
			if (paymentMethod == "bank" || paymentMethod == "va") && (existingOrder.VANumber == "" || (req.Bank != "" && !strings.EqualFold(existingOrder.VABank, req.Bank))) {
				bankName := strings.ToLower(strings.TrimSpace(req.Bank))
				if bankName == "" {
					bankName = "bca"
				}
				bankReq := &coreapi.ChargeReq{
					PaymentType: coreapi.PaymentTypeBankTransfer,
					TransactionDetails: midtrans.TransactionDetails{
						OrderID:  existingOrder.OrderID,
						GrossAmt: int64(existingOrder.TotalAmount),
					},
					CustomerDetails: &midtrans.CustomerDetails{
						FName: existingOrder.CustomerName,
						Email: existingOrder.CustomerEmail,
						Phone: existingOrder.CustomerPhone,
					},
					BankTransfer: &coreapi.BankTransferDetails{
						Bank: midtrans.Bank(bankName),
					},
				}
				coreResp, coreErr := s.coreClient.ChargeTransaction(bankReq)
				if coreErr == nil && coreResp != nil && len(coreResp.VaNumbers) > 0 {
					existingOrder.VANumber = coreResp.VaNumbers[0].VANumber
					existingOrder.VABank = coreResp.VaNumbers[0].Bank
					existingOrder.PaymentType = "bank"
					s.db.Model(&existingOrder).Updates(map[string]interface{}{
						"va_number":    existingOrder.VANumber,
						"va_bank":      existingOrder.VABank,
						"payment_type": "bank",
					})
				}
			} else if (paymentMethod == "wallet" || paymentMethod == "qris" || paymentMethod == "") && existingOrder.QRISURL == "" {
				qrisReq := &coreapi.ChargeReq{
					PaymentType: coreapi.PaymentTypeQris,
					TransactionDetails: midtrans.TransactionDetails{
						OrderID:  existingOrder.OrderID,
						GrossAmt: int64(existingOrder.TotalAmount),
					},
					CustomerDetails: &midtrans.CustomerDetails{
						FName: existingOrder.CustomerName,
						Email: existingOrder.CustomerEmail,
						Phone: existingOrder.CustomerPhone,
					},
					Qris: &coreapi.QrisDetails{
						Acquirer: "gopay",
					},
				}
				coreResp, coreErr := s.coreClient.ChargeTransaction(qrisReq)
				if coreErr == nil && coreResp != nil {
					existingOrder.QRISString = coreResp.QRString
					baseURL := "https://api.sandbox.midtrans.com"
					if s.cfg.MidtransIsProduction {
						baseURL = "https://api.midtrans.com"
					}
					if coreResp.TransactionID != "" {
						existingOrder.QRISURL = fmt.Sprintf("%s/v2/qris/%s/qr-code", baseURL, coreResp.TransactionID)
					}
					if existingOrder.QRISURL == "" {
						for _, act := range coreResp.Actions {
							if act.Name == "generate-qr-code" {
								existingOrder.QRISURL = act.URL
								break
							}
						}
					}
					existingOrder.PaymentType = "wallet"
					s.db.Model(&existingOrder).Updates(map[string]interface{}{
						"qris_url":     existingOrder.QRISURL,
						"qris_string":  existingOrder.QRISString,
						"payment_type": "wallet",
					})
				}
			}

			return &CheckoutResponse{
				OrderID:         existingOrder.OrderID,
				SnapToken:       existingOrder.SnapToken,
				SnapRedirectURL: existingOrder.SnapRedirectURL,
				QRISURL:         existingOrder.QRISURL,
				QRISString:      existingOrder.QRISString,
				VANumber:        existingOrder.VANumber,
				VABank:          existingOrder.VABank,
				TotalAmount:     existingOrder.TotalAmount,
				Status:          existingOrder.Status,
				IsReused:        true,
			}, nil
		}
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("gagal mengecek idempotency key: %w", err)
	}

	var qrisURL string
	var qrisString string
	var vaNumber string
	var vaBank string

	paymentMethod = strings.ToLower(strings.TrimSpace(req.PaymentMethod))

	if paymentMethod == "wallet" || paymentMethod == "qris" || paymentMethod == "" {
		qrisReq := &coreapi.ChargeReq{
			PaymentType: coreapi.PaymentTypeQris,
			TransactionDetails: midtrans.TransactionDetails{
				OrderID:  orderID,
				GrossAmt: itemsSum,
			},
			Items: &midtransItems,
			CustomerDetails: &midtrans.CustomerDetails{
				FName: req.Customer.Name,
				Email: req.Customer.Email,
				Phone: req.Customer.Phone,
			},
			Qris: &coreapi.QrisDetails{
				Acquirer: "gopay",
			},
		}

		coreResp, coreErr := s.coreClient.ChargeTransaction(qrisReq)
		if coreErr == nil && coreResp != nil {
			qrisString = coreResp.QRString
			baseURL := "https://api.sandbox.midtrans.com"
			if s.cfg.MidtransIsProduction {
				baseURL = "https://api.midtrans.com"
			}
			if coreResp.TransactionID != "" {
				qrisURL = fmt.Sprintf("%s/v2/qris/%s/qr-code", baseURL, coreResp.TransactionID)
			}
			if qrisURL == "" {
				for _, act := range coreResp.Actions {
					if act.Name == "generate-qr-code" {
						qrisURL = act.URL
						break
					}
				}
			}
		} else if coreErr != nil {
			log.Printf("[MIDTRANS QRIS CHARGE WARNING] %v", coreErr)
		}
	} else if paymentMethod == "bank" || paymentMethod == "va" {
		bankName := strings.ToLower(strings.TrimSpace(req.Bank))
		if bankName == "" {
			bankName = "bca"
		}
		bankReq := &coreapi.ChargeReq{
			PaymentType: coreapi.PaymentTypeBankTransfer,
			TransactionDetails: midtrans.TransactionDetails{
				OrderID:  orderID,
				GrossAmt: itemsSum,
			},
			Items: &midtransItems,
			CustomerDetails: &midtrans.CustomerDetails{
				FName: req.Customer.Name,
				Email: req.Customer.Email,
				Phone: req.Customer.Phone,
			},
			BankTransfer: &coreapi.BankTransferDetails{
				Bank: midtrans.Bank(bankName),
			},
		}

		coreResp, coreErr := s.coreClient.ChargeTransaction(bankReq)
		if coreErr == nil && coreResp != nil {
			if len(coreResp.VaNumbers) > 0 {
				vaNumber = coreResp.VaNumbers[0].VANumber
				vaBank = coreResp.VaNumbers[0].Bank
			} else if coreResp.BillKey != "" && coreResp.BillerCode != "" {
				vaNumber = fmt.Sprintf("%s%s", coreResp.BillerCode, coreResp.BillKey)
				vaBank = "mandiri"
			} else if coreResp.PermataVaNumber != "" {
				vaNumber = coreResp.PermataVaNumber
				vaBank = "permata"
			}
		} else if coreErr != nil {
			log.Printf("[MIDTRANS VA CHARGE WARNING] %v", coreErr)
		}
	}

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

	var snapToken string
	var snapRedirectURL string
	snapResp, snapErr := s.snapClient.CreateTransaction(snapReq)
	if snapErr == nil && snapResp != nil {
		snapToken = snapResp.Token
		snapRedirectURL = snapResp.RedirectURL
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
		SnapToken:       snapToken,
		SnapRedirectURL: snapRedirectURL,
		QRISURL:         qrisURL,
		QRISString:      qrisString,
		VANumber:        vaNumber,
		VABank:          vaBank,
		PaymentType:     paymentMethod,
		OrderItems:      orderItems,
	}

	tx := s.db.Begin()
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("gagal menyimpan order ke database: %w", err)
	}

	// Immediate stock deduction when product is purchased
	for _, item := range orderItems {
		if item.ProductID > 0 {
			if err := tx.Model(&model.Product{}).
				Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
				UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("stok produk ID %d tidak mencukupi untuk dibeli", item.ProductID)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("gagal commit order: %w", err)
	}

	return &CheckoutResponse{
		OrderID:         orderID,
		SnapToken:       snapToken,
		SnapRedirectURL: snapRedirectURL,
		QRISURL:         qrisURL,
		QRISString:      qrisString,
		VANumber:        vaNumber,
		VABank:          vaBank,
		TotalAmount:     totalAmount,
		Status:          "pending",
		IsReused:        false,
	}, nil
}

func (s *paymentService) HandleNotification(payload map[string]interface{}) error {
	orderID, _ := payload["order_id"].(string)
	statusCode, _ := payload["status_code"].(string)
	grossAmountStr, _ := payload["gross_amount"].(string)
	signatureKey, _ := payload["signature_key"].(string)
	transactionStatus, _ := payload["transaction_status"].(string)
	fraudStatus, _ := payload["fraud_status"].(string)
	paymentType, _ := payload["payment_type"].(string)

	log.Printf("[MIDTRANS NOTIFICATION] order_id: %s | status: %s", orderID, transactionStatus)

	if orderID == "" || strings.HasPrefix(strings.ToLower(orderID), "test") || strings.Contains(strings.ToLower(orderID), "dummy") {
		return nil
	}

	if statusCode == "" || grossAmountStr == "" || signatureKey == "" {
		return nil
	}

	rawSignature := orderID + statusCode + grossAmountStr + s.cfg.MidtransServerKey
	hasher := sha512.New()
	hasher.Write([]byte(rawSignature))
	expectedSignature := hex.EncodeToString(hasher.Sum(nil))

	if !strings.EqualFold(signatureKey, expectedSignature) {
		log.Printf("[MIDTRANS SECURITY] Signature key mismatch untuk order_id: %s", orderID)
		return errors.New("invalid signature key")
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var order model.Order
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_id = ?", orderID).Preload("OrderItems").First(&order).Error; err != nil {
		tx.Rollback()
		log.Printf("[MIDTRANS ERROR] order_id %s tidak ditemukan: %v", orderID, err)
		return nil
	}

	if grossAmt, parseErr := strconv.ParseFloat(grossAmountStr, 64); parseErr == nil {
		if grossAmt != order.TotalAmount {
			tx.Rollback()
			log.Printf("[MIDTRANS SECURITY] Gross amount mismatch untuk order_id %s (diterima: %.2f, expected: %.2f)", orderID, grossAmt, order.TotalAmount)
			return errors.New("gross amount mismatch")
		}
	}

	statusLower := strings.ToLower(order.Status)
	if statusLower == "paid" || statusLower == "settlement" || statusLower == "packaging" || statusLower == "shipped" || statusLower == "on_the_road" || statusLower == "delivered" || statusLower == "completed" {
		tx.Rollback()
		log.Printf("[MIDTRANS IDEMPOTENT] order_id %s sudah berstatus '%s'", orderID, order.Status)
		return nil
	}

	isPaid := false
	newStatus := order.Status

	switch transactionStatus {
	case "capture":
		if fraudStatus == "challenge" {
			newStatus = "challenge"
		} else if fraudStatus == "accept" || fraudStatus == "" {
			newStatus = "packaging"
			isPaid = true
		}
	case "settlement":
		newStatus = "packaging"
		isPaid = true
	case "cancel", "deny", "expire":
		newStatus = "failed"
	case "pending":
		newStatus = "pending"
	}

	now := time.Now()
	updateFields := map[string]interface{}{
		"status":       newStatus,
		"payment_type": paymentType,
	}

	if isPaid {
		updateFields["paid_at"] = &now

		for _, item := range order.OrderItems {
			if item.ProductID > 0 {
				if err := tx.Model(&model.Product{}).
					Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
					UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
					tx.Rollback()
					log.Printf("[PAYMENT ERROR] Gagal mengurangi stok produk ID %d: %v", item.ProductID, err)
					return fmt.Errorf("gagal mengurangi stok produk ID %d: %w", item.ProductID, err)
				}
			}
		}
	}

	if err := tx.Model(&model.Order{}).Where("order_id = ?", orderID).Updates(updateFields).Error; err != nil {
		tx.Rollback()
		log.Printf("[PAYMENT ERROR] Gagal mengupdate status database order_id %s: %v", orderID, err)
		return fmt.Errorf("gagal mengupdate status order: %w", err)
	}

	activity := model.ActivityLog{
		Type:        "PAYMENT_UPDATE",
		Sender:      "SYSTEM",
		Description: fmt.Sprintf("Order %s berubah status menjadi %s via %s", orderID, newStatus, paymentType),
	}
	_ = tx.Create(&activity).Error

	if err := tx.Commit().Error; err != nil {
		log.Printf("[PAYMENT ERROR] Gagal commit transaksi database: %v", err)
		return err
	}

	log.Printf("[PAYMENT SUCCESS] Status order_id=%s diperbarui menjadi %s", orderID, newStatus)

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
	return s.orderService.GetOrders()
}

func (s *paymentService) GetOrderByID(orderID string) (*model.Order, error) {
	return s.orderService.GetOrderByID(orderID)
}

func (s *paymentService) GenerateInvoicePDF(orderID string) ([]byte, error) {
	return s.invoiceService.GenerateInvoicePDF(orderID)
}

func (s *paymentService) CancelOrder(orderID string) (*model.Order, error) {
	return s.orderService.CancelOrder(orderID)
}
