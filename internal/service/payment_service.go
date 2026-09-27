package service

import (
	"bytes"
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

	"github.com/go-pdf/fpdf"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
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
}

type paymentService struct {
	cfg        *config.Config
	db         *gorm.DB
	snapClient snap.Client
	coreClient coreapi.Client
}

func NewPaymentService(cfg *config.Config, db *gorm.DB) PaymentService {
	var s snap.Client
	var c coreapi.Client
	env := midtrans.Sandbox
	if cfg.MidtransIsProduction {
		env = midtrans.Production
	}

	s.New(cfg.MidtransServerKey, env)
	c.New(cfg.MidtransServerKey, env)

	return &paymentService{
		cfg:        cfg,
		db:         db,
		snapClient: s,
		coreClient: c,
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
			QRISURL:         existingOrder.QRISURL,
			QRISString:      existingOrder.QRISString,
			VANumber:        existingOrder.VANumber,
			VABank:          existingOrder.VABank,
			TotalAmount:     existingOrder.TotalAmount,
			Status:          existingOrder.Status,
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

	var qrisURL string
	var qrisString string
	var vaNumber string
	var vaBank string

	paymentMethod := strings.ToLower(strings.TrimSpace(req.PaymentMethod))

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
			for _, act := range coreResp.Actions {
				if act.Name == "generate-qr-code" {
					qrisURL = act.URL
					break
				}
			}
			if qrisURL == "" && coreResp.TransactionID != "" {
				qrisURL = fmt.Sprintf("https://api.sandbox.midtrans.com/v2/qris/%s/qr-code", coreResp.TransactionID)
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
		if coreErr == nil && coreResp != nil && len(coreResp.VaNumbers) > 0 {
			vaNumber = coreResp.VaNumbers[0].VANumber
			vaBank = coreResp.VaNumbers[0].Bank
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
	tx.Commit()

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

func (s *paymentService) GenerateInvoicePDF(orderID string) ([]byte, error) {
	order, err := s.GetOrderByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("order tidak ditemukan: %w", err)
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFillColor(27, 99, 146)
	pdf.Rect(0, 0, 210, 30, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 16)
	pdf.Text(14, 18, "DHANDI ECOMMERCE")
	pdf.SetFont("Arial", "", 10)
	pdf.Text(140, 18, "INVOICE PEMBAYARAN RESMI")

	pdf.SetTextColor(30, 30, 30)
	pdf.SetFont("Arial", "B", 11)
	pdf.Text(14, 42, "Rincian Pesanan & Pembayaran:")

	pdf.SetFont("Arial", "", 9)
	pdf.Text(14, 50, fmt.Sprintf("Nomor Pesanan (Order ID): %s", order.OrderID))
	pdf.Text(14, 56, fmt.Sprintf("Tanggal: %s", order.CreatedAt.Format("02 January 2006 15:04 WIB")))
	pdf.Text(14, 62, fmt.Sprintf("Status Pembayaran: %s", strings.ToUpper(order.Status)))
	paymentTypeDisplay := order.PaymentType
	if paymentTypeDisplay == "" {
		paymentTypeDisplay = "QRIS / VA"
	}
	pdf.Text(14, 68, fmt.Sprintf("Metode Pembayaran: Midtrans (%s)", paymentTypeDisplay))

	custName := order.CustomerName
	if custName == "" {
		custName = "Customer"
	}
	custEmail := order.CustomerEmail
	if custEmail == "" {
		custEmail = "-"
	}
	custPhone := order.CustomerPhone
	if custPhone == "" {
		custPhone = "-"
	}
	custAddress := order.ShippingAddress
	if custAddress == "" {
		custAddress = "Indonesia"
	}

	pdf.Text(120, 50, fmt.Sprintf("Nama Pembeli: %s", custName))
	pdf.Text(120, 56, fmt.Sprintf("Email: %s", custEmail))
	pdf.Text(120, 62, fmt.Sprintf("No. Telepon: %s", custPhone))
	pdf.Text(120, 68, fmt.Sprintf("Alamat: %s", truncateString(custAddress, 35)))

	pdf.SetY(78)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(45, 165, 243)
	pdf.SetTextColor(255, 255, 255)

	pdf.CellFormat(90, 8, " Nama Produk", "1", 0, "L", true, 0, "")
	pdf.CellFormat(20, 8, " Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(36, 8, " Harga Satuan", "1", 0, "R", true, 0, "")
	pdf.CellFormat(36, 8, " Total", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(30, 30, 30)

	for _, item := range order.OrderItems {
		pdf.CellFormat(90, 8, fmt.Sprintf(" %s", truncateString(item.Title, 45)), "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 8, fmt.Sprintf("%d ", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(36, 8, fmt.Sprintf("Rp %.0f ", item.Price), "1", 0, "R", false, 0, "")
		pdf.CellFormat(36, 8, fmt.Sprintf("Rp %.0f ", item.Price*float64(item.Quantity)), "1", 1, "R", false, 0, "")
	}

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(146, 8, "Total Pembayaran: ", "1", 0, "R", false, 0, "")
	pdf.CellFormat(36, 8, fmt.Sprintf("Rp %.0f ", order.TotalAmount), "1", 1, "R", false, 0, "")

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("gagal merender PDF: %w", err)
	}

	return buf.Bytes(), nil
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
