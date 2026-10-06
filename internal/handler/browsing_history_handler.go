package handler

import (
	"strconv"

	"bot_be/internal/service"

	"github.com/gofiber/fiber/v2"
)

type BrowsingHistoryHandler struct {
	historyService service.BrowsingHistoryService
}

func NewBrowsingHistoryHandler(historyService service.BrowsingHistoryService) *BrowsingHistoryHandler {
	return &BrowsingHistoryHandler{
		historyService: historyService,
	}
}

type RecordBrowsingHistoryRequest struct {
	ProductID uint   `json:"product_id"`
	UserID    uint   `json:"user_id"`
	UserEmail string `json:"user_email"`
	SessionID string `json:"session_id"`
}

func (h *BrowsingHistoryHandler) RecordBrowsingHistory(c *fiber.Ctx) error {
	var req RecordBrowsingHistoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}

	if req.ProductID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product_id wajib diisi",
		})
	}

	history, err := h.historyService.RecordView(req.UserID, req.UserEmail, req.SessionID, req.ProductID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Riwayat penelusuran berhasil disimpan",
		"data":    history,
	})
}

func (h *BrowsingHistoryHandler) GetBrowsingHistory(c *fiber.Ctx) error {
	userEmail := c.Query("user_email")
	sessionID := c.Query("session_id")
	query := c.Query("q")
	date := c.Query("date")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	filter := service.BrowsingHistoryFilter{
		UserEmail: userEmail,
		SessionID: sessionID,
		Query:     query,
		Date:      date,
		Page:      page,
		Limit:     limit,
	}

	histories, total, err := h.historyService.GetHistory(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil riwayat penelusuran",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"total": total,
		"page":  page,
		"limit": limit,
		"data":  histories,
	})
}

func (h *BrowsingHistoryHandler) DeleteBrowsingHistoryItem(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID riwayat tidak valid",
		})
	}

	userEmail := c.Query("user_email")
	sessionID := c.Query("session_id")

	if err := h.historyService.DeleteHistoryItem(uint(id), userEmail, sessionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menghapus riwayat penelusuran",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Item riwayat berhasil dihapus",
	})
}

func (h *BrowsingHistoryHandler) ClearBrowsingHistory(c *fiber.Ctx) error {
	userEmail := c.Query("user_email")
	sessionID := c.Query("session_id")

	if userEmail == "" && sessionID == "" {
		var body struct {
			UserEmail string `json:"user_email"`
			SessionID string `json:"session_id"`
		}
		_ = c.BodyParser(&body)
		userEmail = body.UserEmail
		sessionID = body.SessionID
	}

	if userEmail == "" && sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user_email atau session_id wajib disertakan",
		})
	}

	if err := h.historyService.ClearHistory(userEmail, sessionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal membersihkan seluruh riwayat",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Seluruh riwayat penelusuran berhasil dibersihkan",
	})
}
