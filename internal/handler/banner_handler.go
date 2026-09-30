package handler

import (
	"bot_be/internal/model"
	"bot_be/internal/service"

	"github.com/gofiber/fiber/v2"
)

type BannerHandler struct {
	bannerService service.BannerService
}

func NewBannerHandler(bannerService service.BannerService) *BannerHandler {
	return &BannerHandler{bannerService: bannerService}
}

func (h *BannerHandler) GetBanners(c *fiber.Ctx) error {
	activeOnly := c.Query("active") == "true"
	banners, err := h.bannerService.GetBanners(activeOnly)
	if err != nil {
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

	if err := h.bannerService.CreateBanner(&banner); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan banner"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Banner berhasil ditambahkan",
		"data":    banner,
	})
}

func (h *BannerHandler) UpdateBanner(c *fiber.Ctx) error {
	id := c.Params("id")
	banner, err := h.bannerService.GetBannerByID(id)
	if err != nil {
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

	if err := h.bannerService.UpdateBanner(banner); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memperbarui banner"})
	}

	return c.JSON(fiber.Map{
		"message": "Banner berhasil diperbarui",
		"data":    banner,
	})
}

func (h *BannerHandler) DeleteBanner(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.bannerService.DeleteBanner(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus banner"})
	}

	return c.JSON(fiber.Map{"message": "Banner berhasil dihapus"})
}
