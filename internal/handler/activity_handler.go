package handler

import (
	"bot_be/internal/model"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ActivityHandler struct {
	DB *gorm.DB
}

func NewActivityHandler(db *gorm.DB) *ActivityHandler {
	return &ActivityHandler{DB: db}
}

func (h *ActivityHandler) GetActivities(c *fiber.Ctx) error {
	var activities []model.ActivityLog
	// Ambil 50 aktivitas terakhir
	if err := h.DB.Order("created_at desc").Limit(50).Find(&activities).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil aktivitas"})
	}

	return c.JSON(fiber.Map{
		"data": activities,
	})
}
