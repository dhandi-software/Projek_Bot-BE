package handler

import (
	"strconv"

	"bot_be/internal/model"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ProductHandler struct {
	db *gorm.DB
}

func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{db: db}
}

func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
	var products []model.Product

	query := h.db.Model(&model.Product{})

	if cat := c.Query("category"); cat != "" {
		query = query.Where("LOWER(category) = LOWER(?)", cat)
	}
	if search := c.Query("q"); search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("LOWER(title) LIKE LOWER(?) OR LOWER(sku) LIKE LOWER(?) OR LOWER(brand) LIKE LOWER(?) OR LOWER(materials) LIKE LOWER(?)", searchTerm, searchTerm, searchTerm, searchTerm)
	}
	if featured := c.Query("featured"); featured == "true" {
		query = query.Where("is_featured = ?", true)
	}
	if activeOnly := c.Query("active"); activeOnly == "true" {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("id desc").Find(&products).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data produk"})
	}

	return c.JSON(fiber.Map{
		"total": len(products),
		"data":  products,
	})
}

func (h *ProductHandler) GetProductByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var product model.Product

	if err := h.db.First(&product, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Produk tidak ditemukan"})
	}

	return c.JSON(product)
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var product model.Product
	if err := c.BodyParser(&product); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format request tidak valid"})
	}

	if product.Title == "" || product.Category == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Nama produk dan kategori wajib diisi"})
	}

	if product.SKU == "" {
		product.SKU = "PROD-" + strconv.FormatInt(c.Context().Time().UnixNano(), 36)
	}

	if err := h.db.Create(&product).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan produk: " + err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Produk berhasil ditambahkan",
		"data":    product,
	})
}

func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	var product model.Product

	if err := h.db.First(&product, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Produk tidak ditemukan"})
	}

	var payload model.Product
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	product.Title = payload.Title
	product.Category = payload.Category
	product.Price = payload.Price
	product.Stock = payload.Stock
	product.Materials = payload.Materials
	product.Brand = payload.Brand
	product.Description = payload.Description
	product.Image = payload.Image
	product.IsFeatured = payload.IsFeatured
	product.IsActive = payload.IsActive
	if payload.SKU != "" {
		product.SKU = payload.SKU
	}

	if err := h.db.Save(&product).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memperbarui produk"})
	}

	return c.JSON(fiber.Map{
		"message": "Produk berhasil diperbarui",
		"data":    product,
	})
}

func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.db.Delete(&model.Product{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus produk"})
	}

	return c.JSON(fiber.Map{"message": "Produk berhasil dihapus"})
}

func (h *ProductHandler) BulkCreateProducts(c *fiber.Ctx) error {
	var products []model.Product
	if err := c.BodyParser(&products); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data import tidak valid"})
	}

	if len(products) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Data produk kosong"})
	}

	now := c.Context().Time().UnixNano()
	for i := range products {
		if products[i].Title == "" {
			continue
		}
		if products[i].Category == "" {
			products[i].Category = "Umum"
		}
		if products[i].SKU == "" {
			products[i].SKU = "IMP-" + strconv.FormatInt(now+int64(i), 36)
		}
		products[i].IsActive = true
	}

	if err := h.db.Create(&products).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal melakukan import produk: " + err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Import produk berhasil",
		"total":   len(products),
		"data":    products,
	})
}

func (h *ProductHandler) SearchProducts(c *fiber.Ctx) error {
	return h.GetProducts(c)
}
