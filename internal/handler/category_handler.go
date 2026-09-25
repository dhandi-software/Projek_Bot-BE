package handler

import (
	"strings"

	"bot_be/internal/model"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CategoryHandler struct {
	db *gorm.DB
}

func NewCategoryHandler(db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

func (h *CategoryHandler) GetCategories(c *fiber.Ctx) error {
	var categories []model.Category
	if err := h.db.Order("name asc").Find(&categories).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data kategori"})
	}

	if len(categories) == 0 {
		defaultCategories := []model.Category{
			{Name: "Computer & Laptop", Slug: "computer-laptop", Description: "Perangkat Komputer dan Laptop", Icon: "Laptop", IsActive: true},
			{Name: "Gaming Console", Slug: "gaming-console", Description: "Konsol Game dan Aksesoris", Icon: "Gaming", IsActive: true},
			{Name: "Smartphone", Slug: "smartphone", Description: "Ponsel Pintar dan Tablet", Icon: "Smartphone", IsActive: true},
			{Name: "Headphone", Slug: "headphone", Description: "Headphone, Earphone, Audio", Icon: "Headphone", IsActive: true},
			{Name: "Computer Accessories", Slug: "computer-accessories", Description: "Aksesoris Komputer", Icon: "Plug", IsActive: true},
			{Name: "Umum", Slug: "umum", Description: "Kategori Umum", Icon: "Package", IsActive: true},
		}
		for _, cat := range defaultCategories {
			h.db.Create(&cat)
		}
		h.db.Order("name asc").Find(&categories)
	}

	return c.JSON(fiber.Map{
		"total": len(categories),
		"data":  categories,
	})
}

func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	name, _ := body["name"].(string)
	name = strings.TrimSpace(name)
	if name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Nama kategori wajib diisi"})
	}

	slug, _ := body["slug"].(string)
	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	}

	desc, _ := body["description"].(string)
	icon, _ := body["icon"].(string)
	isActive := true
	if val, ok := body["is_active"].(bool); ok {
		isActive = val
	}

	var existing model.Category
	if err := h.db.Where("LOWER(name) = LOWER(?)", name).First(&existing).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Kategori dengan nama ini sudah ada",
		})
	}

	category := model.Category{
		Name:        name,
		Slug:        slug,
		Description: desc,
		Icon:        icon,
		IsActive:    isActive,
	}

	if err := h.db.Create(&category).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan kategori: " + err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Kategori berhasil dibuat",
		"data":    category,
	})
}

func (h *CategoryHandler) UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	var category model.Category

	if err := h.db.First(&category, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Kategori tidak ditemukan"})
	}

	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	if name, ok := body["name"].(string); ok && strings.TrimSpace(name) != "" {
		newName := strings.TrimSpace(name)
		var existing model.Category
		if err := h.db.Where("LOWER(name) = LOWER(?) AND id != ?", newName, category.ID).First(&existing).Error; err == nil {
			return c.Status(400).JSON(fiber.Map{"error": "Kategori dengan nama ini sudah ada"})
		}
		category.Name = newName
		category.Slug = strings.ToLower(strings.ReplaceAll(category.Name, " ", "-"))
	}
	if desc, ok := body["description"].(string); ok {
		category.Description = desc
	}
	if icon, ok := body["icon"].(string); ok {
		category.Icon = icon
	}
	if val, ok := body["is_active"].(bool); ok {
		category.IsActive = val
	}

	if err := h.db.Save(&category).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memperbarui kategori: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Kategori berhasil diperbarui",
		"data":    category,
	})
}

func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.db.Delete(&model.Category{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus kategori"})
	}

	return c.JSON(fiber.Map{"message": "Kategori berhasil dihapus"})
}
