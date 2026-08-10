package handler

import (
	"regexp"
	"bot_be/internal/model"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"bot_be/internal/provider"
)

type ConfigHandler struct {
	DB             *gorm.DB
	SheetsProvider *provider.SheetsProvider
}

func NewConfigHandler(db *gorm.DB, sheetsProvider *provider.SheetsProvider) *ConfigHandler {
	return &ConfigHandler{
		DB:             db,
		SheetsProvider: sheetsProvider,
	}
}

// GetSpreadsheetID mengambil konfigurasi ID Spreadsheet
func (h *ConfigHandler) GetSpreadsheetID(c *fiber.Ctx) error {
	var appConfig model.AppConfig
	err := h.DB.Where("key = ?", "spreadsheet_id").First(&appConfig).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(fiber.Map{"spreadsheet_id": ""})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil konfigurasi"})
	}

	if appConfig.Value != "" {
		fullURL := "https://docs.google.com/spreadsheets/d/" + appConfig.Value + "/edit"
		return c.JSON(fiber.Map{"spreadsheet_id": fullURL})
	}
	return c.JSON(fiber.Map{"spreadsheet_id": appConfig.Value})
}

// SaveSpreadsheetID menyimpan konfigurasi ID Spreadsheet
func (h *ConfigHandler) SaveSpreadsheetID(c *fiber.Ctx) error {
	var req struct {
		SpreadsheetID string `json:"spreadsheet_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format input salah"})
	}

	// Ekstrak ID jika user memasukkan URL lengkap
	// Contoh: https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	finalID := req.SpreadsheetID
	re := regexp.MustCompile(`/d/([a-zA-Z0-9-_]+)`)
	matches := re.FindStringSubmatch(req.SpreadsheetID)
	if len(matches) > 1 {
		finalID = matches[1]
	}

	// Validasi Spreadsheet ID ke Google Sheets API
	if h.SheetsProvider != nil && h.SheetsProvider.Service != nil {
		_, err := h.SheetsProvider.Service.Spreadsheets.Get(finalID).Do()
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Spreadsheet ID tidak valid atau tidak dapat diakses bot",
			})
		}
	} else {
		return c.Status(500).JSON(fiber.Map{
			"error": "Sistem Bot belum memiliki credentials.json Google Sheets",
		})
	}

	var appConfig model.AppConfig
	err := h.DB.Where("key = ?", "spreadsheet_id").First(&appConfig).Error

	if err == gorm.ErrRecordNotFound {
		// Create new
		appConfig = model.AppConfig{
			Key:   "spreadsheet_id",
			Value: finalID,
		}
		if err := h.DB.Create(&appConfig).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan konfigurasi"})
		}
	} else if err == nil {
		// Update existing
		appConfig.Value = finalID
		if err := h.DB.Save(&appConfig).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal memperbarui konfigurasi"})
		}
	} else {
		return c.Status(500).JSON(fiber.Map{"error": "Kesalahan database"})
	}

	return c.JSON(fiber.Map{
		"message": "Konfigurasi berhasil disimpan",
		"spreadsheet_id": appConfig.Value,
	})
}
