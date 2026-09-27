package handler

import (
	"log"
	"strings"

	"bot_be/internal/service"

	"github.com/gofiber/fiber/v2"
)

type PaymentHandler struct {
	paymentService service.PaymentService
}

func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

func (h *PaymentHandler) CreateCheckoutTransaction(c *fiber.Ctx) error {
	var req service.CreateCheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[Checkout Error] Body parser failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}

	if req.IdempotencyKey == "" {
		req.IdempotencyKey = strings.TrimSpace(c.Get("Idempotency-Key"))
	}

	resp, err := h.paymentService.CreateCheckoutTransaction(req)
	if err != nil {
		log.Printf("[Checkout Error] CreateCheckoutTransaction failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Printf("[Checkout Success] OrderID: %s, SnapToken: %s", resp.OrderID, resp.SnapToken)

	return c.JSON(fiber.Map{
		"message": "Transaksi berhasil dibuat",
		"data":    resp,
	})
}

func (h *PaymentHandler) HandleNotification(c *fiber.Ctx) error {
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		log.Printf("[Notification Error] Body parser failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Payload notifikasi tidak valid",
		})
	}

	log.Printf("[Notification Incoming] Payload: %v", payload)

	if err := h.paymentService.HandleNotification(payload); err != nil {
		log.Printf("[Notification Warning] HandleNotification: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Notifikasi berhasil diproses",
	})
}

func (h *PaymentHandler) GetOrders(c *fiber.Ctx) error {
	orders, err := h.paymentService.GetOrders()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil daftar order",
		})
	}

	return c.JSON(fiber.Map{
		"data": orders,
	})
}

func (h *PaymentHandler) GetOrderByID(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID wajib diisi",
		})
	}

	order, err := h.paymentService.GetOrderByID(orderID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order tidak ditemukan",
		})
	}

	return c.JSON(fiber.Map{
		"data": order,
	})
}

func (h *PaymentHandler) GetOrderInvoice(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID wajib diisi",
		})
	}

	pdfBytes, err := h.paymentService.GenerateInvoicePDF(orderID)
	if err != nil {
		log.Printf("[Invoice Error] %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Invoice tidak ditemukan atau gagal diproses",
		})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=\"Invoice_"+orderID+".pdf\"")
	return c.Send(pdfBytes)
}
