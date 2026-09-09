package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ProductItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Price    string `json:"price"`
	Brand    string `json:"brand"`
	Image    string `json:"image"`
}

var sampleProducts = []ProductItem{
	{
		ID:       "1",
		Title:    "MacBook Pro M3 Max 16-inch",
		Category: "Computer & Laptop",
		Price:    "$2,499",
		Brand:    "Apple",
		Image:    "https://images.unsplash.com/photo-1517336714731-489689fd1ca8?auto=format&fit=crop&w=300&q=80",
	},
	{
		ID:       "2",
		Title:    "Dell XPS 15 OLED Touch Laptop",
		Category: "Computer & Laptop",
		Price:    "$1,899",
		Brand:    "Dell",
		Image:    "https://images.unsplash.com/photo-1593642632823-8f785ba67e45?auto=format&fit=crop&w=300&q=80",
	},
	{
		ID:       "3",
		Title:    "Sony WH-1000XM5 Wireless Headphones",
		Category: "Headphone",
		Price:    "$399",
		Brand:    "Sony",
		Image:    "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?auto=format&fit=crop&w=300&q=80",
	},
	{
		ID:       "4",
		Title:    "Samsung Galaxy S24 Ultra 512GB",
		Category: "Smartphone",
		Price:    "$1,299",
		Brand:    "Samsung",
		Image:    "https://images.unsplash.com/photo-1610945265064-0e34e5519bbf?auto=format&fit=crop&w=300&q=80",
	},
	{
		ID:       "5",
		Title:    "LG UltraGear 32-inch Gaming Monitor",
		Category: "Computer Accessories",
		Price:    "$699",
		Brand:    "LG",
		Image:    "https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?auto=format&fit=crop&w=300&q=80",
	},
	{
		ID:       "6",
		Title:    "Google Pixel 8 Pro AI Smartphone",
		Category: "Smartphone",
		Price:    "$999",
		Brand:    "Google",
		Image:    "https://images.unsplash.com/photo-1598327105666-5b89351aff97?auto=format&fit=crop&w=300&q=80",
	},
	{
		ID:       "7",
		Title:    "Logitech MX Master 3S Wireless Mouse",
		Category: "Computer Accessories",
		Price:    "$99",
		Brand:    "Logitech",
		Image:    "https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?auto=format&fit=crop&w=300&q=80",
	},
	{
		ID:       "8",
		Title:    "Keychron K2 Mechanical Keyboard RGB",
		Category: "Computer Accessories",
		Price:    "$89",
		Brand:    "Keychron",
		Image:    "https://images.unsplash.com/photo-1587829741301-dc798b83add3?auto=format&fit=crop&w=300&q=80",
	},
	{
		ID:       "9",
		Title:    "Asus ROG Strix Gaming Laptop 17-inch",
		Category: "Computer & Laptop",
		Price:    "$2,199",
		Brand:    "Asus",
		Image:    "https://images.unsplash.com/photo-1603302576837-37561b2e2302?auto=format&fit=crop&w=300&q=80",
	},
	{
		ID:       "10",
		Title:    "Apple Airpods Pro 2nd Gen Wireless",
		Category: "Headphone",
		Price:    "$249",
		Brand:    "Apple",
		Image:    "https://images.unsplash.com/photo-1600294037681-c80b4cb5b434?auto=format&fit=crop&w=300&q=80",
	},
}

type ProductHandler struct{}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

// SearchProducts handles GET /api/products/search?q=...
func (h *ProductHandler) SearchProducts(c *fiber.Ctx) error {
	query := strings.TrimSpace(strings.ToLower(c.Query("q")))

	if query == "" {
		return c.JSON(fiber.Map{
			"query":   "",
			"total":   0,
			"results": []ProductItem{},
		})
	}

	var results []ProductItem
	for _, p := range sampleProducts {
		if strings.Contains(strings.ToLower(p.Title), query) ||
			strings.Contains(strings.ToLower(p.Category), query) ||
			strings.Contains(strings.ToLower(p.Brand), query) {
			results = append(results, p)
		}
	}

	return c.JSON(fiber.Map{
		"query":   query,
		"total":   len(results),
		"results": results,
	})
}
