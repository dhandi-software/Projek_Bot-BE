package handler

import (
	"strings"

	"bot_be/internal/model"
	"bot_be/internal/service"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CustomerHandler struct {
	customerService service.CustomerService
	DB              *gorm.DB
}

func NewCustomerHandler(customerService service.CustomerService, db *gorm.DB) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
		DB:              db,
	}
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

	customer := model.Customer{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: req.Address,
		Bio:     req.Bio,
		Photo:   req.Photo,
	}

	if err := h.customerService.RegisterCustomer(&customer, req.Password); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Pendaftaran akun customer berhasil",
		"token":   "customer-jwt-token-demo",
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

// LoginCustomer memproses login customer
func (h *CustomerHandler) LoginCustomer(c *fiber.Ctx) error {
	var req CustomerLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}

	email := strings.TrimSpace(req.Email)
	password := req.Password

	if email == "" || password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email/Nomor HP dan password wajib diisi",
		})
	}

	customer, err := h.customerService.GetCustomerByEmail(email)
	if err != nil {
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

// GetCustomerProfile mengembalikan profil customer
func (h *CustomerHandler) GetCustomerProfile(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		email = "user@example.com"
	}

	customer, err := h.customerService.GetCustomerByEmail(email)
	if err != nil {
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

	customer, err := h.customerService.GetCustomerByEmail(req.Email)
	if err != nil {
		newCust := model.Customer{
			Name:    req.Name,
			Email:   req.Email,
			Phone:   req.Phone,
			Address: req.Address,
			Bio:     req.Bio,
			Photo:   req.Photo,
		}
		if err := h.customerService.RegisterCustomer(&newCust, "defaultPass123"); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal menyimpan profil customer baru",
			})
		}
		customer = &newCust
	} else {
		if req.Name != "" {
			customer.Name = req.Name
		}
		if req.Phone != "" {
			customer.Phone = req.Phone
		}
		if req.Address != "" {
			customer.Address = req.Address
		}
		if req.Bio != "" {
			customer.Bio = req.Bio
		}
		if req.Photo != "" {
			customer.Photo = req.Photo
		}

		if err := h.customerService.UpdateCustomer(customer); err != nil {
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
	customers, err := h.customerService.GetAllCustomers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil data customer",
		})
	}

	return c.JSON(fiber.Map{
		"data": customers,
	})
}

// GetDashboardStats mengdelegasikan statistik ke DashboardHandler untuk pemisahan kode yang bersih
func (h *CustomerHandler) GetDashboardStats(c *fiber.Ctx) error {
	dh := NewDashboardHandler(h.DB)
	return dh.GetDashboardStats(c)
}
