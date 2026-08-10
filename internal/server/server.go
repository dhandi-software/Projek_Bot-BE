package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	
	"bot_be/internal/config"
	"bot_be/internal/handler"
	"bot_be/internal/provider"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/gorm"
	"go.mau.fi/whatsmeow"
)

// StartWebServer menginisialisasi dan menjalankan server Fiber
func StartWebServer(cfg *config.Config, db *gorm.DB, waClient *whatsmeow.Client, sheetsProvider *provider.SheetsProvider) {
	app := fiber.New()

	// Middleware
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return true // Mengizinkan origin apapun secara dinamis
		},
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))
	app.Use(logger.New())

	// Init Handlers
	authHandler := handler.NewAuthHandler(db)
	configHandler := handler.NewConfigHandler(db, sheetsProvider)
	activityHandler := handler.NewActivityHandler(db)

	// Routes
	api := app.Group("/api")
	
	// API Auth
	api.Post("/login", authHandler.Login)

	// API Activities
	api.Get("/activities", activityHandler.GetActivities)

	// API Config
	api.Get("/config/spreadsheet", configHandler.GetSpreadsheetID)
	api.Post("/config/spreadsheet", configHandler.SaveSpreadsheetID)

	// API WhatsApp
	api.Get("/wa/qr", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"qr": provider.WAQRString,
		})
	})

	api.Get("/wa/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"is_logged_in": waClient.IsLoggedIn(),
		})
	})

	api.Post("/wa/logout", func(c *fiber.Ctx) error {
		if waClient.IsLoggedIn() {
			waClient.Logout(context.Background())
		} else {
			waClient.Disconnect()
		}
		// Reset QR
		provider.WAQRString = ""
		
		// Force restart to trigger new QR code generation
		go func() {
			time.Sleep(500 * time.Millisecond)
			os.RemoveAll("data")
			os.Exit(0)
		}()
		
		return c.JSON(fiber.Map{"message": "Berhasil logout WhatsApp"})
	})
	
	// Simple Health Check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("Server Bot OK!")
	})

	fmt.Printf("\n🚀 Berhasil! Aplikasi Backend berjalan di: http://localhost:%s\n", cfg.Port)
	fmt.Printf("🌐 Silakan buka Frontend di: http://localhost:3000\n\n")
	
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
