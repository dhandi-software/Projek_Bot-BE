package server

import (
	"fmt"
	"log"
	
	"bot_be/internal/config"
	"bot_be/internal/handler"
	"bot_be/internal/provider"
	"bot_be/internal/service"
	"bot_be/internal/wshub"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

// StartWebServer menginisialisasi dan menjalankan server Fiber
func StartWebServer(cfg *config.Config, db *gorm.DB, sheetsProvider *provider.SheetsProvider) {
	app := fiber.New()

	// Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSAllowedOrigins,
		AllowOriginsFunc: func(origin string) bool {
			if cfg.CORSAllowedOrigins == "*" || cfg.CORSAllowedOrigins == "" {
				return true
			}
			return false
		},
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		AllowMethods:     "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
		AllowCredentials: true,
	}))
	app.Use(logger.New())

	// Init Services & Handlers
	botService := service.NewBotService(cfg, sheetsProvider, db)
	messageHandler := handler.NewMessageHandler(botService)
	authHandler := handler.NewAuthHandler(db)
	configHandler := handler.NewConfigHandler(db, sheetsProvider)
	activityHandler := handler.NewActivityHandler(db)
	customerHandler := handler.NewCustomerHandler(db)

	// Start WAHA Session automatically
	provider.StartSession()

	// Routes
	api := app.Group("/api")
	
	// Middleware upgrade WebSocket
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	wshub.InitWSHub()

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		wshub.Hub.Register <- c
		defer func() {
			wshub.Hub.Unregister <- c
		}()

		// Keep connection alive and read messages if needed
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				break
			}
		}
	}))

	// API Auth (Admin & Customer)
	api.Post("/auth/login", authHandler.Login)
	api.Post("/login", authHandler.Login)
	api.Post("/auth/check-email", authHandler.CheckEmail)
	api.Post("/auth/reset-password", authHandler.ResetPassword)
	api.Post("/auth/change-password", authHandler.ChangePassword)
	api.Post("/customer/register", customerHandler.RegisterCustomer)
	api.Post("/customer/login", customerHandler.LoginCustomer)
	api.Get("/customer/profile", customerHandler.GetCustomerProfile)
	api.Put("/customer/profile", customerHandler.UpdateCustomerProfile)
	api.Post("/customer/profile", customerHandler.UpdateCustomerProfile)
	api.Get("/admin/customers", customerHandler.GetAdminCustomers)
	api.Get("/admin/dashboard/stats", customerHandler.GetDashboardStats)

	// API Activities
	api.Get("/activities", activityHandler.GetActivities)

	// API Config
	api.Get("/config/spreadsheet", configHandler.GetSpreadsheetID)
	api.Post("/config/spreadsheet", configHandler.SaveSpreadsheetID)

	// API WhatsApp WAHA endpoints
	api.Get("/wa/qr", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"qr": handler.FetchQR(),
		})
	})

	api.Get("/wa/status", func(c *fiber.Ctx) error {
		status, _ := provider.GetSessionStatus()
		return c.JSON(fiber.Map{
			"is_logged_in": status == "WORKING",
		})
	})

	api.Post("/wa/logout", func(c *fiber.Ctx) error {
		provider.LogoutSession()
		return c.JSON(fiber.Map{"message": "Berhasil logout WhatsApp"})
	})

	// Webhook from WAHA
	api.Post("/wa/webhook", messageHandler.HandleWebhook)

	// Add WhatsApp Chat API endpoints
	chatHandler := handler.NewChatHandler()
	api.Get("/chat/contacts", chatHandler.GetContacts)
	api.Get("/chat/history/:jid", chatHandler.GetChatHistory)
	api.Post("/chat/send", chatHandler.SendMessage)

	// API Products CRUD
	productHandler := handler.NewProductHandler(db)
	api.Get("/products", productHandler.GetProducts)
	api.Get("/products/search", productHandler.SearchProducts)
	api.Get("/products/:id", productHandler.GetProductByID)
	api.Post("/products", productHandler.CreateProduct)
	api.Post("/products/bulk", productHandler.BulkCreateProducts)
	api.Put("/products/:id", productHandler.UpdateProduct)
	api.Delete("/products/:id", productHandler.DeleteProduct)

	// API Banners CRUD
	bannerHandler := handler.NewBannerHandler(db)
	api.Get("/banners", bannerHandler.GetBanners)
	api.Post("/banners", bannerHandler.CreateBanner)
	api.Put("/banners/:id", bannerHandler.UpdateBanner)
	api.Delete("/banners/:id", bannerHandler.DeleteBanner)
	
	// Simple Health Check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("Server Bot OK!")
	})

	fmt.Printf("\n🚀 Berhasil! Aplikasi Backend berjalan di: http://localhost:%s\n", cfg.Port)
	fmt.Printf("🌐 Silakan buka Frontend di: %s\n\n", cfg.ClientURL)
	
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}

