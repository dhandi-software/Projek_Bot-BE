package handler

import (
	"bot_be/internal/model"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type BannerHandler struct {
	db *gorm.DB
}

func NewBannerHandler(db *gorm.DB) *BannerHandler {
	return &BannerHandler{db: db}
}

func (h *BannerHandler) GetBanners(c *fiber.Ctx) error {
	var banners []model.Banner
	query := h.db.Model(&model.Banner{})

	if activeOnly := c.Query("active"); activeOnly == "true" {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("sort_order asc, id desc").Find(&banners).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data banner"})
	}

	return c.JSON(fiber.Map{
		"total": len(banners),
		"data":  banners,
	})
}

func (h *BannerHandler) CreateBanner(c *fiber.Ctx) error {
	var banner model.Banner
	if err := c.BodyParser(&banner); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format request tidak valid"})
	}

	if banner.Title == "" || banner.Image == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Judul dan gambar banner wajib diisi"})
	}

	if err := h.db.Create(&banner).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan banner"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Banner berhasil ditambahkan",
		"data":    banner,
	})
}

func (h *BannerHandler) UpdateBanner(c *fiber.Ctx) error {
	id := c.Params("id")
	var banner model.Banner

	if err := h.db.First(&banner, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Banner tidak ditemukan"})
	}

	var payload model.Banner
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	banner.Title = payload.Title
	banner.Subtitle = payload.Subtitle
	banner.Tagline = payload.Tagline
	banner.Image = payload.Image
	banner.PriceBadge = payload.PriceBadge
	banner.LinkUrl = payload.LinkUrl
	banner.ButtonText = payload.ButtonText
	banner.BgColor = payload.BgColor
	banner.IsActive = payload.IsActive
	banner.SortOrder = payload.SortOrder

	if err := h.db.Save(&banner).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memperbarui banner"})
	}

	return c.JSON(fiber.Map{
		"message": "Banner berhasil diperbarui",
		"data":    banner,
	})
}

func (h *BannerHandler) DeleteBanner(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.db.Delete(&model.Banner{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus banner"})
	}

	return c.JSON(fiber.Map{"message": "Banner berhasil dihapus"})
}
