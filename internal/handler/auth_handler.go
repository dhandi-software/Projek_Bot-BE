package handler

import (
	"bot_be/internal/model"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CheckEmailRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	UserID      uint   `json:"userId"`
	NewPassword string `json:"newPassword"`
}

// Login memproses permintaan login dari customer atau admin
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}

	// 1. Cek pada tabel Customer (email / phone)
	var customer model.Customer
	if err := h.DB.Where("email = ? OR phone = ?", req.Username, req.Username).First(&customer).Error; err == nil {
		if errPass := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(req.Password)); errPass == nil {
			return c.JSON(fiber.Map{
				"message": "Login customer berhasil",
				"token":   "customer-jwt-token-demo",
				"user": fiber.Map{
					"id":      customer.ID,
					"name":    customer.Name,
					"email":   customer.Email,
					"phone":   customer.Phone,
					"address": customer.Address,
					"role":    "customer",
				},
			})
		}
	}

	// 2. Cek pada tabel Admin (username)
	var admin model.Admin
	if err := h.DB.Where("username = ?", req.Username).First(&admin).Error; err == nil {
		if errPass := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); errPass == nil {
			return c.JSON(fiber.Map{
				"message": "Login admin berhasil",
				"token":   "admin-jwt-token-demo",
				"user": fiber.Map{
					"id":       admin.ID,
					"username": admin.Username,
					"name":     "Administrator",
					"role":     "admin",
				},
			})
		}
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": "Username/email atau password salah",
	})
}

// CheckEmail mengecek keberadaan email customer atau admin untuk alur lupa password
func (h *AuthHandler) CheckEmail(c *fiber.Ctx) error {
	var req CheckEmailRequest

	if err := c.BodyParser(&req); err != nil || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Email tidak boleh kosong",
		})
	}

	// Cek tabel Customer terlebih dahulu
	var customer model.Customer
	if err := h.DB.Where("email = ?", req.Email).First(&customer).Error; err == nil {
		return c.JSON(fiber.Map{
			"userId":  customer.ID,
			"role":    "customer",
			"message": "Email ditemukan. Silakan buat password baru.",
		})
	}

	// Cek tabel Admin (menggunakan username/email)
	var admin model.Admin
	if err := h.DB.Where("username = ?", req.Email).First(&admin).Error; err == nil {
		return c.JSON(fiber.Map{
			"userId":  admin.ID,
			"role":    "admin",
			"message": "Email/Username admin ditemukan.",
		})
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"message": "Email tidak terdaftar dalam sistem.",
	})
}

// ResetPassword memperbarui password user/customer berdasarkan userId
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest

	if err := c.BodyParser(&req); err != nil || req.UserID == 0 || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User ID dan password baru wajib diisi",
		})
	}

	if len(req.NewPassword) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Password baru minimal 6 karakter",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal memproses password baru",
		})
	}

	// Coba update pada tabel Customer
	result := h.DB.Model(&model.Customer{}).Where("id = ?", req.UserID).Update("password", string(hashedPassword))
	if result.RowsAffected > 0 {
		return c.JSON(fiber.Map{
			"message": "Password berhasil diubah. Silakan login kembali.",
		})
	}

	// Jika tidak ada baris terupdate di Customer, coba update tabel Admin
	resultAdmin := h.DB.Model(&model.Admin{}).Where("id = ?", req.UserID).Update("password", string(hashedPassword))
	if resultAdmin.RowsAffected > 0 {
		return c.JSON(fiber.Map{
			"message": "Password admin berhasil diubah. Silakan login kembali.",
		})
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"message": "User tidak ditemukan untuk diproses.",
	})
}

type ChangePasswordRequest struct {
	UserID          uint   `json:"userId"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ChangePassword mengonfirmasi kata sandi saat ini dan merubah kata sandi pengguna
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil || req.CurrentPassword == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Semua kolom kata sandi wajib diisi",
		})
	}

	if len(req.NewPassword) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Kata sandi baru minimal 6 karakter",
		})
	}

	// 1. Coba verifikasi pada tabel Customer
	if req.UserID > 0 {
		var customer model.Customer
		if err := h.DB.First(&customer, req.UserID).Error; err == nil {
			if errPass := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(req.CurrentPassword)); errPass != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "Kata sandi saat ini tidak sesuai",
				})
			}
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "Gagal memproses kata sandi baru",
				})
			}
			h.DB.Model(&customer).Update("password", string(hashedPassword))
			return c.JSON(fiber.Map{
				"message": "Kata sandi berhasil diperbarui!",
			})
		}
	}

	// 2. Jika tidak ditemukan berdasarkan UserID atau UserID == 0, cari customer pertama/pertama sesuai data login
	var customer model.Customer
	if err := h.DB.First(&customer).Error; err == nil {
		if errPass := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(req.CurrentPassword)); errPass != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Kata sandi saat ini tidak sesuai",
			})
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Gagal memproses kata sandi baru",
			})
		}
		h.DB.Model(&customer).Update("password", string(hashedPassword))
		return c.JSON(fiber.Map{
			"message": "Kata sandi berhasil diperbarui!",
		})
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"message": "User tidak ditemukan untuk diperbarui.",
	})
}
