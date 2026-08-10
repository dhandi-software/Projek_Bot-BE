package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bot_be/internal/config"
	"bot_be/internal/handler"
	"bot_be/internal/provider"
	"bot_be/internal/server"
	"bot_be/internal/service"
)

func main() {
	// 1. Load Config dari .env
	cfg := config.LoadConfig()

	// 2. Init Context
	ctx := context.Background()

	// 3. Init Database PostgreSQL
	db, err := provider.InitDatabase(cfg)
	if err != nil {
		log.Fatalf("Gagal inisialisasi Database: %v", err)
	}

	// 4. Init Google Sheets API (Optional)
	var sheetsProvider *provider.SheetsProvider
	if _, err := os.Stat("credentials.json"); err == nil {
		sheetsProvider, err = provider.InitGoogleSheets(ctx, "credentials.json")
		if err != nil {
			log.Printf("WARNING: Gagal inisialisasi Google Sheets (%v). Bot akan berjalan tanpa fitur Sheets.", err)
			sheetsProvider = nil
		}
	} else {
		log.Println("WARNING: File 'credentials.json' tidak ditemukan. Bot akan berjalan tanpa fitur Google Sheets.")
	}

	// 5. Init Service & Handler
	botService := service.NewBotService(cfg, sheetsProvider, db)
	messageHandler := handler.NewMessageHandler(botService)

	// 6. Init WhatsApp Client
	client, err := provider.InitWhatsApp(messageHandler.EventHandler)
	if err != nil {
		log.Fatalf("Gagal inisialisasi WhatsApp: %v", err)
	}

	// 7. Jalankan Web Server di Goroutine terpisah agar tidak memblokir WhatsApp
	go func() {
		server.StartWebServer(cfg, db, client, sheetsProvider)
	}()

	// 8. Listen to OS signals to gracefully stop the bot
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Mematikan bot...")
	client.Disconnect()
}
