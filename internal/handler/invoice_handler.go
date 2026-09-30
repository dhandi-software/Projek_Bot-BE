package handler

import (
	"log"

	"bot_be/internal/service"

	"github.com/gofiber/fiber/v2"
)

type InvoiceHandler struct {
	invoiceService service.InvoiceService
}

func NewInvoiceHandler(invoiceService service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: invoiceService}
}

func (h *InvoiceHandler) GetOrderInvoice(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID wajib diisi",
		})
	}

	pdfBytes, err := h.invoiceService.GenerateInvoicePDF(orderID)
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
