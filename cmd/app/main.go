package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bot_be/internal/config"
	"bot_be/internal/provider"
	"bot_be/internal/server"
)

func main() {
	cfg := config.LoadConfig()
	ctx := context.Background()

	db, err := provider.InitDatabase(cfg)
	if err != nil {
		log.Fatalf("Gagal inisialisasi Database: %v", err)
	}

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

	go func() {
		server.StartWebServer(cfg, db, sheetsProvider)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Mematikan bot...")
}
