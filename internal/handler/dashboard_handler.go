package handler

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"bot_be/internal/model"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	DB *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{DB: db}
}

func formatRupiah(val float64) string {
	in := int64(val)
	str := strconv.FormatInt(in, 10)
	n := len(str)
	if n <= 3 {
		return "Rp " + str
	}
	var result []byte
	rem := n % 3
	if rem > 0 {
		result = append(result, str[:rem]...)
		result = append(result, '.')
	}
	for i := rem; i < n; i += 3 {
		result = append(result, str[i:i+3]...)
		if i+3 < n {
			result = append(result, '.')
		}
	}
	return "Rp " + string(result)
}

func formatShortRupiah(val float64) string {
	if val >= 1_000_000_000 {
		return fmt.Sprintf("Rp %.1f M", val/1_000_000_000.0)
	} else if val >= 1_000_000 {
		return fmt.Sprintf("Rp %.1f jt", val/1_000_000.0)
	} else if val >= 1_000 {
		return fmt.Sprintf("Rp %.1fk", val/1_000.0)
	}
	return fmt.Sprintf("Rp %.0f", val)
}

// GetDashboardStats mengembalikan statistik dashboard admin berdasarkan data transaksi dan produk
func (h *DashboardHandler) GetDashboardStats(c *fiber.Ctx) error {
	var orders []model.Order
	if err := h.DB.Preload("OrderItems").Order("created_at desc").Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil data transaksi",
		})
	}

	var products []model.Product
	h.DB.Find(&products)
	productMap := make(map[uint]model.Product)
	for _, p := range products {
		productMap[p.ID] = p
	}

	now := time.Now()
	yearParam := c.Query("year")
	monthParam := c.Query("month")

	selectedYear := now.Year()
	if y, err := strconv.Atoi(yearParam); err == nil && y > 2000 {
		selectedYear = y
	}

	selectedMonth := 0
	if m, err := strconv.Atoi(monthParam); err == nil && m >= 1 && m <= 12 {
		selectedMonth = m
	}

	yearSet := map[int]bool{
		now.Year():     true,
		now.Year() - 1: true,
	}
	for _, ord := range orders {
		yearSet[ord.CreatedAt.Year()] = true
	}
	var availableYears []int
	for y := range yearSet {
		availableYears = append(availableYears, y)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(availableYears)))

	totalOrderCount := len(orders)
	var totalOmset float64
	for _, ord := range orders {
		totalOmset += ord.TotalAmount
	}
	avgOrderVal := 0.0
	if totalOrderCount > 0 {
		avgOrderVal = totalOmset / float64(totalOrderCount)
	}

	var salesTrend []fiber.Map
	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

	if selectedMonth == 0 {
		monthMap := make(map[int]float64)
		for _, ord := range orders {
			if ord.CreatedAt.Year() == selectedYear {
				m := int(ord.CreatedAt.Month())
				monthMap[m] += ord.TotalAmount
			}
		}
		for m := 1; m <= 12; m++ {
			valM := monthMap[m] / 1_000_000.0
			salesTrend = append(salesTrend, fiber.Map{
				"day":   monthNames[m-1],
				"value": math.Round(valM*100) / 100,
			})
		}
	} else {
		daysInMonth := time.Date(selectedYear, time.Month(selectedMonth)+1, 0, 0, 0, 0, 0, time.UTC).Day()
		dayMap := make(map[int]float64)
		monthName := monthNames[selectedMonth-1]

		for _, ord := range orders {
			if ord.CreatedAt.Year() == selectedYear && int(ord.CreatedAt.Month()) == selectedMonth {
				d := ord.CreatedAt.Day()
				dayMap[d] += ord.TotalAmount
			}
		}
		for d := 1; d <= daysInMonth; d++ {
			valD := dayMap[d] / 1_000_000.0
			salesTrend = append(salesTrend, fiber.Map{
				"day":   fmt.Sprintf("%02d %s", d, monthName),
				"value": math.Round(valD*100) / 100,
			})
		}
	}

	type prodSummary struct {
		ID            string
		ProductID     uint
		Title         string
		SoldCount     int
		TotalAmount   float64
		LastOrderTime time.Time
		Image         string
	}

	summaryMap := make(map[string]*prodSummary)

	for _, ord := range orders {
		for _, item := range ord.OrderItems {
			key := fmt.Sprintf("%d", item.ProductID)
			if item.ProductID == 0 {
				key = item.Title
			}

			title := item.Title
			img := item.Image
			if img == "" {
				img = item.ImageURL
			}

			if item.ProductID > 0 {
				if realProd, exists := productMap[item.ProductID]; exists {
					if realProd.Title != "" {
						title = realProd.Title
					}
					if realProd.Image != "" {
						img = realProd.Image
					}
				}
			}

			s, exists := summaryMap[key]
			if !exists {
				s = &prodSummary{
					ID:            key,
					ProductID:     item.ProductID,
					Title:         title,
					SoldCount:     0,
					TotalAmount:   0,
					LastOrderTime: ord.CreatedAt,
					Image:         img,
				}
				summaryMap[key] = s
			}

			s.SoldCount += item.Quantity
			s.TotalAmount += item.Price * float64(item.Quantity)
			if ord.CreatedAt.After(s.LastOrderTime) {
				s.LastOrderTime = ord.CreatedAt
			}
		}
	}

	summaries := make([]*prodSummary, 0, len(summaryMap))
	for _, s := range summaryMap {
		if s.SoldCount > 0 {
			summaries = append(summaries, s)
		}
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].SoldCount == summaries[j].SoldCount {
			return summaries[i].LastOrderTime.After(summaries[j].LastOrderTime)
		}
		return summaries[i].SoldCount > summaries[j].SoldCount
	})

	topProducts := make([]fiber.Map, 0, 5)
	for i, s := range summaries {
		if i >= 5 {
			break
		}
		topProducts = append(topProducts, fiber.Map{
			"rank":       i + 1,
			"title":      s.Title,
			"soldCount":  s.SoldCount,
			"change":     "+10.0%",
			"isPositive": true,
			"image":      s.Image,
		})
	}

	produkKeluar := make([]fiber.Map, 0, len(summaries))
	for _, s := range summaries {
		timeStr := s.LastOrderTime.Format("02 Jan, 15:04")
		if s.LastOrderTime.Format("2006-01-02") == now.Format("2006-01-02") {
			timeStr = "Hari ini, " + s.LastOrderTime.Format("15:04")
		}

		idDisplay := s.ID
		if s.ProductID > 0 {
			idDisplay = fmt.Sprintf("ID: %d", s.ProductID)
		}

		produkKeluar = append(produkKeluar, fiber.Map{
			"id":            idDisplay,
			"title":         s.Title,
			"soldQty":       s.SoldCount,
			"totalAmount":   formatRupiah(s.TotalAmount),
			"lastOrderDate": timeStr,
			"status":        "Terjual & Stok Berkurang",
			"image":         s.Image,
		})
	}

	latestOrders := make([]fiber.Map, 0, len(orders))
	for _, ord := range orders {
		var items []string
		for _, it := range ord.OrderItems {
			title := it.Title
			if it.ProductID > 0 {
				if realProd, exists := productMap[it.ProductID]; exists && realProd.Title != "" {
					title = realProd.Title
				}
			}
			items = append(items, fmt.Sprintf("%s ×%d", title, it.Quantity))
		}
		itemsStr := strings.Join(items, ", ")
		if itemsStr == "" {
			itemsStr = "Detail Pesanan"
		}

		custName := ord.CustomerName
		if custName == "" {
			custName = "Pelanggan"
		}

		st := ord.Status
		switch strings.ToLower(st) {
		case "paid", "settlement":
			st = "Selesai"
		case "pending":
			st = "Diproses"
		default:
			if st == "" {
				st = "Dibayar"
			}
		}

		orderLabel := ord.OrderID
		if ord.OrderNumber != "" {
			orderLabel = ord.OrderNumber
		}

		pm := ord.PaymentType
		if pm == "" {
			pm = "Transfer Bank"
		} else {
			pm = strings.Title(pm)
		}

		latestOrders = append(latestOrders, fiber.Map{
			"id":            orderLabel,
			"customer":      custName,
			"items":         itemsStr,
			"total":         formatRupiah(ord.TotalAmount),
			"paymentMethod": pm,
			"status":        st,
			"date":          ord.CreatedAt.Format("02 Jan 2006"),
			"time":          ord.CreatedAt.Format("15:04"),
		})
	}

	omsetStr := formatShortRupiah(totalOmset)
	if totalOmset == 0 {
		omsetStr = "Rp 0"
	}
	avgStr := formatShortRupiah(avgOrderVal)
	if avgOrderVal == 0 {
		avgStr = "Rp 0"
	}

	stats := fiber.Map{
		"omset": fiber.Map{
			"amount":     omsetStr,
			"change":     "+0.0%",
			"isPositive": true,
			"comparison": "vs bulan lalu",
		},
		"totalOrder": fiber.Map{
			"count":      totalOrderCount,
			"change":     "+0.0%",
			"isPositive": true,
			"comparison": "vs bulan lalu",
		},
		"averageOrderValue": fiber.Map{
			"value":      avgStr,
			"change":     "+0.0%",
			"isPositive": true,
			"comparison": "vs bulan lalu",
		},
		"returnRate": fiber.Map{
			"rate":       "0.0%",
			"change":     "0.0pp",
			"isPositive": true,
			"comparison": "Target: di bawah 3%",
		},
		"pencapaianTarget": fiber.Map{
			"percentage": 100,
			"status":     "tercapai",
			"realisasi":  omsetStr,
			"target":     omsetStr,
			"surplus":    "Rp 0",
		},
		"salesTrend":     salesTrend,
		"topProducts":    topProducts,
		"latestOrders":   latestOrders,
		"produkKeluar":   produkKeluar,
		"availableYears": availableYears,
		"selectedYear":   selectedYear,
		"selectedMonth":  selectedMonth,
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   stats,
	})
}
