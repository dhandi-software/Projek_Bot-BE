package handler

import (
	"encoding/json"
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
	var bodyMap map[string]interface{}
	if err := c.BodyParser(&bodyMap); err != nil {
		log.Printf("[Checkout Error] Body parser failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}

	// Smart check: If Midtrans sends a Notification Webhook Callback directly to /payment/checkout
	if _, hasSig := bodyMap["signature_key"]; hasSig || bodyMap["transaction_status"] != nil || bodyMap["order_id"] != nil && bodyMap["items"] == nil {
		log.Printf("[Checkout Webhook] Midtrans callback detected on /payment/checkout")
		if err := h.paymentService.HandleNotification(bodyMap); err != nil {
			log.Printf("[Checkout Webhook Warning] HandleNotification error: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Notifikasi berhasil diproses",
		})
	}

	var req service.CreateCheckoutRequest
	reqBytes, _ := json.Marshal(bodyMap)
	_ = json.Unmarshal(reqBytes, &req)

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

	log.Printf("[Checkout Success] OrderID: %s", resp.OrderID)

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

	if err := h.paymentService.HandleNotification(payload); err != nil {
		log.Printf("[Notification Warning] HandleNotification error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Notifikasi berhasil diproses",
	})
}
