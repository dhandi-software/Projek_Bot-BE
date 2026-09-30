package handler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"bot_be/internal/model"
	"bot_be/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func stringifyFlexJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func parseFlexFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		cleaned := strings.TrimSpace(val)
		cleaned = strings.ReplaceAll(cleaned, "Rp", "")
		cleaned = strings.ReplaceAll(cleaned, ".", "")
		cleaned = strings.ReplaceAll(cleaned, ",", ".")
		f, _ := strconv.ParseFloat(cleaned, 64)
		return f
	}
	return 0
}

func parseFlexInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	case string:
		cleaned := strings.TrimSpace(val)
		i, _ := strconv.Atoi(cleaned)
		return i
	}
	return 0
}

func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
	filter := service.ProductFilter{
		Category:  c.Query("category"),
		Query:     c.Query("q"),
		Featured:  c.Query("featured") == "true",
		Active:    c.Query("active") == "true",
		Status:    c.Query("status"),
		BestDeals: c.Query("best_deals") == "true",
		Now:       c.Context().Time(),
	}

	products, err := h.productService.GetProducts(filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data produk"})
	}

	return c.JSON(fiber.Map{
		"total": len(products),
		"data":  products,
	})
}

func (h *ProductHandler) GetProductByID(c *fiber.Ctx) error {
	param := c.Params("id")
	product, err := h.productService.GetProductByID(param)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Produk tidak ditemukan"})
	}
	return c.JSON(product)
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format request tidak valid: " + err.Error()})
	}

	title, _ := body["title"].(string)
	if title == "" {
		if nameVal, ok := body["name"].(string); ok {
			title = nameVal
		}
	}

	category, _ := body["category"].(string)
	if title == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Nama produk wajib diisi"})
	}
	if category == "" {
		category = "Umum"
	}

	sku, _ := body["sku"].(string)
	if sku == "" {
		sku = fmt.Sprintf("SKU-%s", strings.ToUpper(uuid.New().String()[:8]))
	}

	brand, _ := body["brand"].(string)
	materials, _ := body["materials"].(string)
	shortDesc, _ := body["short_description"].(string)
	desc, _ := body["description"].(string)
	image, _ := body["image"].(string)
	status, _ := body["status"].(string)

	if status == "" {
		status = "active"
	}

	price := parseFlexFloat(body["price"])
	discountPrice := parseFlexFloat(body["discount_price"])
	stock := parseFlexInt(body["stock"])
	lowStockThreshold := parseFlexInt(body["low_stock_threshold"])
	if lowStockThreshold == 0 && body["low_stock_threshold"] == nil {
		lowStockThreshold = 10
	}
	weight := parseFlexFloat(body["weight"])

	isFeatured, _ := body["is_featured"].(bool)
	isActive := true
	if val, ok := body["is_active"].(bool); ok {
		isActive = val
	} else if strings.ToLower(status) == "draft" || strings.ToLower(status) == "archived" {
		isActive = false
	}

	isBestDeal, _ := body["is_best_deal"].(bool)
	var bestDealStartedAt *time.Time
	var bestDealExpiresAt *time.Time

	if isBestDeal {
		now := c.Context().Time()
		durationHours := parseFlexInt(body["best_deal_duration"])
		if durationHours <= 0 {
			durationHours = parseFlexInt(body["best_deal_duration_hours"])
		}
		if durationHours <= 0 {
			durationHours = 6
		}
		expires := now.Add(time.Duration(durationHours) * time.Hour)
		bestDealStartedAt = &now
		bestDealExpiresAt = &expires
	}

	featuresStr := stringifyFlexJSON(body["features"])
	colorsStr := stringifyFlexJSON(body["colors"])
	shippingInfoStr := stringifyFlexJSON(body["shipping_info"])
	additionalInfoStr := stringifyFlexJSON(body["additional_info"])
	specificationsStr := stringifyFlexJSON(body["specifications"])

	product := model.Product{
		SKU:               sku,
		Title:             title,
		Category:          category,
		Price:             price,
		DiscountPrice:     discountPrice,
		Stock:             stock,
		LowStockThreshold: lowStockThreshold,
		Weight:            weight,
		Materials:         materials,
		Brand:             brand,
		ShortDescription:  shortDesc,
		Description:       desc,
		Image:             image,
		Status:            status,
		IsFeatured:        isFeatured,
		IsActive:          isActive,
		IsBestDeal:        isBestDeal,
		BestDealStartedAt: bestDealStartedAt,
		BestDealExpiresAt: bestDealExpiresAt,
		Features:          featuresStr,
		Colors:            colorsStr,
		ShippingInfo:      shippingInfoStr,
		AdditionalInfo:    additionalInfoStr,
		Specifications:    specificationsStr,
	}

	if err := h.productService.CreateProduct(&product); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan produk: " + err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Produk berhasil ditambahkan",
		"data":    product,
	})
}

func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	product, err := h.productService.GetProductByID(idStr)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Produk tidak ditemukan"})
	}

	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	if title, ok := body["title"].(string); ok && title != "" {
		product.Title = title
	} else if nameVal, ok := body["name"].(string); ok && nameVal != "" {
		product.Title = nameVal
	}

	if cat, ok := body["category"].(string); ok && cat != "" {
		product.Category = cat
	}
	if sku, ok := body["sku"].(string); ok && sku != "" {
		product.SKU = sku
	}
	if brand, ok := body["brand"].(string); ok {
		product.Brand = brand
	}
	if mat, ok := body["materials"].(string); ok {
		product.Materials = mat
	}
	if sDesc, ok := body["short_description"].(string); ok {
		product.ShortDescription = sDesc
	}
	if desc, ok := body["description"].(string); ok {
		product.Description = desc
	}
	if img, ok := body["image"].(string); ok {
		product.Image = img
	}
	if st, ok := body["status"].(string); ok && st != "" {
		product.Status = st
		if strings.ToLower(st) == "draft" || strings.ToLower(st) == "archived" {
			product.IsActive = false
		} else if strings.ToLower(st) == "active" {
			product.IsActive = true
		}
	}

	if val, ok := body["price"]; ok {
		product.Price = parseFlexFloat(val)
	}
	if val, ok := body["discount_price"]; ok {
		product.DiscountPrice = parseFlexFloat(val)
	}
	if val, ok := body["stock"]; ok {
		product.Stock = parseFlexInt(val)
	}
	if val, ok := body["low_stock_threshold"]; ok {
		product.LowStockThreshold = parseFlexInt(val)
	}
	if val, ok := body["weight"]; ok {
		product.Weight = parseFlexFloat(val)
	}
	if val, ok := body["is_featured"].(bool); ok {
		product.IsFeatured = val
	}
	if val, ok := body["is_active"].(bool); ok {
		product.IsActive = val
	}

	if val, ok := body["features"]; ok {
		product.Features = stringifyFlexJSON(val)
	}
	if val, ok := body["colors"]; ok {
		product.Colors = stringifyFlexJSON(val)
	}
	if val, ok := body["shipping_info"]; ok {
		product.ShippingInfo = stringifyFlexJSON(val)
	}
	if val, ok := body["additional_info"]; ok {
		product.AdditionalInfo = stringifyFlexJSON(val)
	}
	if val, ok := body["specifications"]; ok {
		product.Specifications = stringifyFlexJSON(val)
	}

	if isBD, ok := body["is_best_deal"].(bool); ok {
		product.IsBestDeal = isBD
		if isBD {
			now := c.Context().Time()
			durationHours := parseFlexInt(body["best_deal_duration"])
			if durationHours <= 0 {
				durationHours = parseFlexInt(body["best_deal_duration_hours"])
			}
			if durationHours <= 0 {
				durationHours = 6
			}
			expires := now.Add(time.Duration(durationHours) * time.Hour)
			product.BestDealStartedAt = &now
			product.BestDealExpiresAt = &expires
		} else {
			product.BestDealStartedAt = nil
			product.BestDealExpiresAt = nil
		}
	} else if durationVal := parseFlexInt(body["best_deal_duration"]); durationVal > 0 {
		product.IsBestDeal = true
		now := c.Context().Time()
		expires := now.Add(time.Duration(durationVal) * time.Hour)
		product.BestDealStartedAt = &now
		product.BestDealExpiresAt = &expires
	}

	if err := h.productService.UpdateProduct(product); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memperbarui produk: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Produk berhasil diperbarui",
		"data":    product,
	})
}

func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID produk tidak valid"})
	}

	if err := h.productService.DeleteProduct(uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus produk: " + err.Error()})
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
			products[i].SKU = fmt.Sprintf("IMP-%d-%d", now, i)
		}
		if products[i].Status == "" {
			products[i].Status = "active"
		}
		products[i].IsActive = true
	}

	if err := h.productService.BulkCreateProducts(products); err != nil {
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
