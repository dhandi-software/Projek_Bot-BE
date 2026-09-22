package handler

import (
	"bot_be/internal/model"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CustomerHandler struct {
	DB *gorm.DB
}

func NewCustomerHandler(db *gorm.DB) *CustomerHandler {
	return &CustomerHandler{DB: db}
}

type CustomerRegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Address  string `json:"address"`
	Bio      string `json:"bio"`
	Photo    string `json:"photo"`
}

type CustomerUpdateRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	Bio     string `json:"bio"`
	Photo   string `json:"photo"`
}

type CustomerLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterCustomer memproses pendaftaran customer baru
func (h *CustomerHandler) RegisterCustomer(c *fiber.Ctx) error {
	var req CustomerRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Nama, email, dan password wajib diisi",
		})
	}

	// Cek apakah email sudah terdaftar
	var existing model.Customer
	if err := h.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Email sudah terdaftar. Silakan login.",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal memproses password",
		})
	}

	customer := model.Customer{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: string(hashedPassword),
		Address:  req.Address,
		Bio:      req.Bio,
		Photo:    req.Photo,
	}

	if err := h.DB.Create(&customer).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menyimpan akun customer",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Pendaftaran akun customer berhasil",
		"token":   "customer-jwt-token-demo",
		"user": fiber.Map{
			"id":       customer.ID,
			"name":     customer.Name,
			"email":    customer.Email,
			"phone":    customer.Phone,
			"address":  customer.Address,
			"bio":      customer.Bio,
			"photo":    customer.Photo,
			"role":     "customer",
		},
	})
}

// LoginCustomer memproses login customer
func (h *CustomerHandler) LoginCustomer(c *fiber.Ctx) error {
	var req CustomerLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}

	var customer model.Customer
	if err := h.DB.Where("email = ? OR phone = ?", req.Email, req.Email).First(&customer).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Email/Nomor HP atau password salah",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Email/Nomor HP atau password salah",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Login customer berhasil",
		"token":   "customer-jwt-token-demo",
		"user": fiber.Map{
			"id":       customer.ID,
			"name":     customer.Name,
			"email":    customer.Email,
			"phone":    customer.Phone,
			"address":  customer.Address,
			"bio":      customer.Bio,
			"photo":    customer.Photo,
			"role":     "customer",
		},
	})
}

// GetCustomerProfile mengembalikan profil customer
func (h *CustomerHandler) GetCustomerProfile(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		email = "user@example.com"
	}

	var customer model.Customer
	if err := h.DB.Where("email = ?", email).First(&customer).Error; err != nil {
		return c.JSON(fiber.Map{
			"message": "Berhasil mendapatkan profil customer",
			"user": fiber.Map{
				"name":    "Pengguna Customer",
				"email":   email,
				"phone":   "081234567890",
				"address": "Jakarta, Indonesia",
				"bio":     "Customer aktif eCommerce",
				"photo":   "/images/avatar.svg",
				"role":    "customer",
			},
		})
	}

	return c.JSON(fiber.Map{
		"message": "Berhasil mendapatkan profil customer",
		"user": fiber.Map{
			"id":      customer.ID,
			"name":    customer.Name,
			"email":   customer.Email,
			"phone":   customer.Phone,
			"address": customer.Address,
			"bio":     customer.Bio,
			"photo":   customer.Photo,
			"role":    "customer",
		},
	})
}

// UpdateCustomerProfile memperbarui data dan foto profil customer
func (h *CustomerHandler) UpdateCustomerProfile(c *fiber.Ctx) error {
	var req CustomerUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format data profil tidak valid",
		})
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email customer wajib diisi",
		})
	}

	var customer model.Customer
	if err := h.DB.Where("email = ?", req.Email).First(&customer).Error; err != nil {
		// Jika belum ada di DB, buat baru
		customer = model.Customer{
			Name:    req.Name,
			Email:   req.Email,
			Phone:   req.Phone,
			Address: req.Address,
			Bio:     req.Bio,
			Photo:   req.Photo,
		}
		if err := h.DB.Create(&customer).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal menyimpan profil customer baru",
			})
		}
	} else {
		// Update data terdaftar
		updates := map[string]interface{}{}
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if req.Phone != "" {
			updates["phone"] = req.Phone
		}
		if req.Address != "" {
			updates["address"] = req.Address
		}
		if req.Bio != "" {
			updates["bio"] = req.Bio
		}
		if req.Photo != "" {
			updates["photo"] = req.Photo
		}

		if err := h.DB.Model(&customer).Updates(updates).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal mengupdate profil customer",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Profil customer berhasil diperbarui",
		"user": fiber.Map{
			"id":      customer.ID,
			"name":    customer.Name,
			"email":   customer.Email,
			"phone":   customer.Phone,
			"address": customer.Address,
			"bio":     customer.Bio,
			"photo":   customer.Photo,
			"role":    "customer",
		},
	})
}

// GetAdminCustomers mengembalikan daftar customer untuk admin
func (h *CustomerHandler) GetAdminCustomers(c *fiber.Ctx) error {
	var customers []model.Customer
	if err := h.DB.Find(&customers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil data customer",
		})
	}

	return c.JSON(fiber.Map{
		"data": customers,
	})
}

// GetDashboardStats mengembalikan statistik dashboard admin (Figma design node 78:15427)
func (h *CustomerHandler) GetDashboardStats(c *fiber.Ctx) error {
	stats := fiber.Map{
		"omset": fiber.Map{
			"amount":      "Rp 91,2 jt",
			"change":      "+18.4%",
			"isPositive":  true,
			"comparison":  "vs Sep 2025: Rp 76,9 jt",
		},
		"totalOrder": fiber.Map{
			"count":      643,
			"change":     "+26.9%",
			"isPositive": true,
			"comparison": "vs bulan lalu: 487 order",
		},
		"averageOrderValue": fiber.Map{
			"value":      "Rp 141,8k",
			"change":     "+5.2%",
			"isPositive": true,
			"comparison": "Bulan lalu: Rp 134,8k",
		},
		"returnRate": fiber.Map{
			"rate":       "2.3%",
			"change":     "-0.4pp",
			"isPositive": true,
			"comparison": "Target: di bawah 3%",
		},
		"pencapaianTarget": fiber.Map{
			"percentage": 107,
			"status":     "tercapai",
			"realisasi":  "Rp 91,2 jt",
			"target":     "Rp 85 jt",
			"surplus":    "+Rp 6,2 jt",
		},
		"monthlyTrends": []fiber.Map{
			{"month": "Mar", "omset": 42, "order": 310, "target": 50},
			{"month": "Apr", "omset": 48, "order": 340, "target": 55},
			{"month": "Mei", "omset": 58, "order": 410, "target": 60},
			{"month": "Jun", "omset": 52, "order": 380, "target": 65},
			{"month": "Jul", "omset": 74, "order": 520, "target": 70},
			{"month": "Agu", "omset": 70, "order": 490, "target": 75},
			{"month": "Sep", "omset": 91.2, "order": 643, "target": 85},
		},
		"trafficSources": []fiber.Map{
			{"name": "Organik", "percentage": 38, "color": "#00a884"},
			{"name": "Iklan Meta", "percentage": 27, "color": "#3b82f6"},
			{"name": "Tokopedia", "percentage": 19, "color": "#10b981"},
			{"name": "WhatsApp", "percentage": 11, "color": "#8b5cf6"},
			{"name": "Lainnya", "percentage": 5, "color": "#6b7280"},
		},
		"weeklyOrders": []fiber.Map{
			{"day": "Sen", "orders": 85, "returns": 3},
			{"day": "Sel", "orders": 110, "returns": 5},
			{"day": "Rab", "orders": 95, "returns": 2},
			{"day": "Kam", "orders": 140, "returns": 8},
			{"day": "Jum", "orders": 165, "returns": 6},
			{"day": "Sab", "orders": 201, "returns": 11},
			{"day": "Min", "orders": 145, "returns": 4},
		},
		"latestOrders": []fiber.Map{
			{
				"id": "ORD-9021",
				"customer": "Budi Santoso",
				"items": "Headphone Pro X1 ×2",
				"total": "Rp 398k",
				"time": "5 mnt yang lalu",
				"status": "Selesai",
			},
			{
				"id": "ORD-9020",
				"customer": "Siti Rahma",
				"items": "Smartwatch Sport V2 ×1",
				"total": "Rp 549k",
				"time": "18 mnt yang lalu",
				"status": "Diproses",
			},
			{
				"id": "ORD-9019",
				"customer": "Andi Wijaya",
				"items": "Keyboard Mechanical RGB ×1",
				"total": "Rp 720k",
				"time": "42 mnt yang lalu",
				"status": "Selesai",
			},
			{
				"id": "ORD-9018",
				"customer": "Dewi Lestari",
				"items": "Mouse Wireless Silent ×3",
				"total": "Rp 285k",
				"time": "1 jam yang lalu",
				"status": "Selesai",
			},
		},
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   stats,
	})
}
