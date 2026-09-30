package service

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
	"gorm.io/gorm"
)

type InvoiceService interface {
	GenerateInvoicePDF(orderID string) ([]byte, error)
}

type invoiceService struct {
	db           *gorm.DB
	orderService OrderService
}

func NewInvoiceService(db *gorm.DB, orderService OrderService) InvoiceService {
	return &invoiceService{
		db:           db,
		orderService: orderService,
	}
}

func (s *invoiceService) GenerateInvoicePDF(orderID string) ([]byte, error) {
	order, err := s.orderService.GetOrderByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("order tidak ditemukan: %w", err)
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(14, 14, 14)
	pdf.SetAutoPageBreak(true, 14)
	pdf.AddPage()

	// 1. Top Header Banner (#191C1F)
	pdf.SetFillColor(25, 28, 31)
	pdf.Rect(0, 0, 210, 36, "F")

	// Blue Accent Bar (#2DA5F3)
	pdf.SetFillColor(45, 165, 243)
	pdf.Rect(0, 35, 210, 1.5, "F")

	// Store Brand Logo Box
	pdf.SetFillColor(45, 165, 243)
	pdf.RoundedRect(14, 8, 11, 11, 2, "1234", "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 14)
	pdf.Text(17.5, 16, "D")

	// Brand Title
	pdf.SetFont("Arial", "B", 15)
	pdf.Text(28, 14.5, "DHANDI ECOMMERCE")
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(180, 190, 200)
	pdf.Text(28, 19, "Platform E-Commerce & Transaksi Pembayaran Resmi")

	// Right Header: Invoice Badge & Order ID
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 11)
	pdf.Text(132, 14.5, "INVOICE PEMBAYARAN RESMI")
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(200, 210, 225)
	pdf.Text(132, 19, fmt.Sprintf("No: %s", order.OrderID))

	// 2. Status & Dates Section
	pdf.SetY(42)
	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(226, 232, 240)
	pdf.Rect(14, 40, 182, 14, "FD")

	statusUpper := strings.ToUpper(order.Status)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(100, 116, 139)
	pdf.Text(18, 48.5, "STATUS PESANAN:")

	if statusUpper == "PAID" || statusUpper == "SETTLEMENT" {
		pdf.SetFillColor(220, 252, 231)
		pdf.SetTextColor(22, 101, 52)
		pdf.RoundedRect(53, 44, 32, 6, 1, "1234", "F")
		pdf.SetFont("Arial", "B", 8)
		pdf.Text(55.5, 48.2, "LUNAS / PAID")
	} else if statusUpper == "EXPIRE" || statusUpper == "EXPIRED" || statusUpper == "CANCEL" || statusUpper == "DENY" {
		pdf.SetFillColor(254, 226, 226)
		pdf.SetTextColor(153, 27, 27)
		pdf.RoundedRect(53, 44, 34, 6, 1, "1234", "F")
		pdf.SetFont("Arial", "B", 8)
		pdf.Text(55.5, 48.2, "GAGAL / EXPIRED")
	} else {
		pdf.SetFillColor(254, 243, 199)
		pdf.SetTextColor(146, 64, 14)
		pdf.RoundedRect(53, 44, 38, 6, 1, "1234", "F")
		pdf.SetFont("Arial", "B", 8)
		pdf.Text(55.5, 48.2, "MENUNGGU BAYAR")
	}

	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(100, 116, 139)
	pdf.Text(125, 48.5, fmt.Sprintf("Tanggal Diterbitkan: %s", order.CreatedAt.Format("02 Jan 2006 15:04 WIB")))

	// 3. Sender & Customer Columns
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

	pdf.SetFont("Arial", "B", 9.5)
	pdf.SetTextColor(15, 23, 42)
	pdf.Text(14, 62, "DITERBITKAN OLEH:")
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(30, 41, 59)
	pdf.Text(14, 67, "Dhandi Ecommerce Store")
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(100, 116, 139)
	pdf.Text(14, 71.5, "Jl. Jendral Sudirman No. 123, Jakarta")
	pdf.Text(14, 76, "Email: support@dhandiecommerce.com")
	pdf.Text(14, 80.5, "Telp: +62 813-1924-0256")

	pdf.SetFont("Arial", "B", 9.5)
	pdf.SetTextColor(15, 23, 42)
	pdf.Text(110, 62, "DITUJUKAN KEPADA (CUSTOMER):")
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(30, 41, 59)
	pdf.Text(110, 67, custName)
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(100, 116, 139)
	pdf.Text(110, 71.5, fmt.Sprintf("Email: %s", custEmail))
	pdf.Text(110, 76, fmt.Sprintf("No. Telp: %s", custPhone))
	pdf.Text(110, 80.5, fmt.Sprintf("Alamat: %s", truncateString(custAddress, 42)))

	// 4. Payment Info Box
	pdf.SetFillColor(240, 249, 255)
	pdf.SetDrawColor(186, 230, 253)
	pdf.Rect(14, 86, 182, 13, "FD")

	paymentTypeDisplay := strings.ToUpper(order.PaymentType)
	if paymentTypeDisplay == "" {
		paymentTypeDisplay = "QRIS / VIRTUAL ACCOUNT"
	} else if paymentTypeDisplay == "BANK_TRANSFER" || paymentTypeDisplay == "BANK" {
		if order.VABank != "" {
			paymentTypeDisplay = fmt.Sprintf("VIRTUAL ACCOUNT (%s)", strings.ToUpper(order.VABank))
		} else {
			paymentTypeDisplay = "VIRTUAL ACCOUNT"
		}
	} else if paymentTypeDisplay == "GOPAY" || paymentTypeDisplay == "QRIS" {
		paymentTypeDisplay = "QRIS / E-WALLET"
	}

	pdf.SetFont("Arial", "B", 8.5)
	pdf.SetTextColor(30, 41, 59)
	pdf.Text(18, 93, fmt.Sprintf("Metode Pembayaran: %s", paymentTypeDisplay))
	if order.VANumber != "" {
		pdf.Text(110, 93, fmt.Sprintf("No. VA: %s", order.VANumber))
	} else if order.PaidAt != nil {
		pdf.Text(110, 93, fmt.Sprintf("Waktu Bayar: %s", order.PaidAt.Format("02 Jan 2006 15:04 WIB")))
	}

	// 5. Items Table
	pdf.SetY(104)
	pdf.SetFont("Arial", "B", 8.5)
	pdf.SetFillColor(45, 165, 243)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetDrawColor(45, 165, 243)

	pdf.CellFormat(12, 7.5, " No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(96, 7.5, "  Nama Produk / Items", "1", 0, "L", true, 0, "")
	pdf.CellFormat(18, 7.5, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(28, 7.5, "Harga Satuan  ", "1", 0, "R", true, 0, "")
	pdf.CellFormat(28, 7.5, "Total Harga  ", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 8.5)
	pdf.SetDrawColor(226, 232, 240)

	for i, item := range order.OrderItems {
		if i%2 == 1 {
			pdf.SetFillColor(248, 250, 252)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.SetTextColor(51, 65, 85)

		pdf.CellFormat(12, 7.5, fmt.Sprintf("%d", i+1), "1", 0, "C", true, 0, "")
		pdf.CellFormat(96, 7.5, fmt.Sprintf("  %s", truncateString(item.Title, 52)), "1", 0, "L", true, 0, "")
		pdf.CellFormat(18, 7.5, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", true, 0, "")
		pdf.CellFormat(28, 7.5, fmt.Sprintf("Rp %s  ", formatRupiahInt(int64(item.Price))), "1", 0, "R", true, 0, "")
		pdf.CellFormat(28, 7.5, fmt.Sprintf("Rp %s  ", formatRupiahInt(int64(item.Price*float64(item.Quantity)))), "1", 1, "R", true, 0, "")
	}

	// 6. Total Summary Box
	currY := pdf.GetY() + 4

	pdf.SetY(currY)
	pdf.SetFont("Arial", "I", 7.5)
	pdf.SetTextColor(100, 116, 139)
	pdf.Text(14, currY+4, "Dokumen ini diterbitkan secara sah oleh sistem Dhandi Ecommerce.")
	pdf.Text(14, currY+8, "Berlaku sebagai bukti pembayaran & transaksi resmi.")

	pdf.SetY(currY)
	pdf.SetX(110)
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(71, 85, 105)
	pdf.SetFillColor(248, 250, 252)

	pdf.CellFormat(46, 7, "  Subtotal Produk:", "1", 0, "L", true, 0, "")
	pdf.SetFont("Arial", "B", 8.5)
	pdf.CellFormat(40, 7, fmt.Sprintf("Rp %s  ", formatRupiahInt(int64(order.TotalAmount))), "1", 1, "R", true, 0, "")

	pdf.SetX(110)
	pdf.SetFont("Arial", "", 8.5)
	pdf.CellFormat(46, 7, "  Ongkos Kirim:", "1", 0, "L", true, 0, "")
	pdf.SetFont("Arial", "B", 8.5)
	pdf.SetTextColor(22, 101, 52)
	pdf.CellFormat(40, 7, "GRATIS  ", "1", 1, "R", true, 0, "")

	pdf.SetX(110)
	pdf.SetFont("Arial", "B", 9.5)
	pdf.SetFillColor(240, 249, 255)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(46, 8.5, "  TOTAL BAYAR:", "1", 0, "L", true, 0, "")
	pdf.SetTextColor(45, 165, 243)
	pdf.CellFormat(40, 8.5, fmt.Sprintf("Rp %s  ", formatRupiahInt(int64(order.TotalAmount))), "1", 1, "R", true, 0, "")

	// 7. Footer
	pdf.SetY(272)
	pdf.SetDrawColor(226, 232, 240)
	pdf.Line(14, 270, 196, 270)
	pdf.SetFont("Arial", "", 7.5)
	pdf.SetTextColor(148, 163, 184)
	pdf.CellFormat(182, 5, "Terima kasih atas kepercayaan Anda berbelanja di Dhandi Ecommerce Store.", "0", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("gagal merender PDF: %w", err)
	}

	return buf.Bytes(), nil
}

func formatRupiahInt(amount int64) string {
	s := fmt.Sprintf("%d", amount)
	n := len(s)
	if n <= 3 {
		return s
	}
	var result []string
	remainder := n % 3
	if remainder > 0 {
		result = append(result, s[:remainder])
	}
	for i := remainder; i < n; i += 3 {
		result = append(result, s[i:i+3])
	}
	return strings.Join(result, ".")
}
