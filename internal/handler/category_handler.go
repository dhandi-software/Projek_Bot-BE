package handler

import (
	"strings"

	"bot_be/internal/model"
	"bot_be/internal/service"

	"github.com/gofiber/fiber/v2"
)

type CategoryHandler struct {
	categoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

func (h *CategoryHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.categoryService.GetCategories()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data kategori"})
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

	if existing, _ := h.categoryService.GetCategoryByName(name); existing != nil {
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

	if err := h.categoryService.CreateCategory(&category); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan kategori: " + err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Kategori berhasil dibuat",
		"data":    category,
	})
}

func (h *CategoryHandler) UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	category, err := h.categoryService.GetCategoryByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Kategori tidak ditemukan"})
	}

	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	if name, ok := body["name"].(string); ok && strings.TrimSpace(name) != "" {
		newName := strings.TrimSpace(name)
		if existing, _ := h.categoryService.GetCategoryByName(newName); existing != nil && existing.ID != category.ID {
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

	if err := h.categoryService.UpdateCategory(category); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memperbarui kategori: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Kategori berhasil diperbarui",
		"data":    category,
	})
}

func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.categoryService.DeleteCategory(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus kategori"})
	}

	return c.JSON(fiber.Map{"message": "Kategori berhasil dihapus"})
}
