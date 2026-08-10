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

// Login memproses permintaan login dari admin
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}

	var admin model.Admin
	if err := h.DB.Where("username = ?", req.Username).First(&admin).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Username atau password salah",
		})
	}

	// Bandingkan hash password di database dengan password inputan
	err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Username atau password salah",
		})
	}

	// Dalam skenario nyata, kembalikan JWT Token di sini.
	// Untuk saat ini, kembalikan pesan sukses beserta data user.
	return c.JSON(fiber.Map{
		"message": "Login berhasil",
		"token":   "dummy-jwt-token-123", // TODO: Implement JWT
		"user": fiber.Map{
			"id":       admin.ID,
			"username": admin.Username,
			"name":     "Administrator",
			"role":     "admin",
		},
	})
}
