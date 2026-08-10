package service

import (
	"fmt"
	"log"
	"time"

	"bot_be/internal/config"
	"bot_be/internal/model"
	"bot_be/internal/provider"

	"gorm.io/gorm"
)

type BotService struct {
	Config   *config.Config
	Sheets   *provider.SheetsProvider
	DB       *gorm.DB
}

func NewBotService(cfg *config.Config, sheetsProvider *provider.SheetsProvider, db *gorm.DB) *BotService {
	return &BotService{
		Config: cfg,
		Sheets: sheetsProvider,
		DB:     db,
	}
}

// HandleIncomingMessage memproses pesan masuk dari WhatsApp dan menyimpannya ke Sheets
func (s *BotService) HandleIncomingMessage(sender string, message string) error {
	// Contoh sederhana: hanya simpan pesan yang formatnya tertentu,
	// atau di contoh ini, simpan semua pesan teks ke Google Sheets.
	log.Printf("Memproses pesan dari %s: %s", sender, message)

	if s.Sheets == nil {
		return fmt.Errorf("Google Sheets provider tidak tersedia (credentials.json tidak ditemukan)")
	}

	var appConfig model.AppConfig
	if err := s.DB.Where("key = ?", "spreadsheet_id").First(&appConfig).Error; err != nil {
		return fmt.Errorf("Spreadsheet ID belum dikonfigurasi di Web Dashboard")
	}
	
	if appConfig.Value == "" {
		return fmt.Errorf("Spreadsheet ID kosong")
	}

	spreadsheetId := appConfig.Value

	// Data yang akan ditulis ke Sheets
	// Asumsi urutan kolom: [Waktu, Pengirim, Pesan]
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	rowData := []interface{}{currentTime, sender, message}

	// Tulis ke Sheet1 kolom A:C
	err := s.Sheets.AppendData(spreadsheetId, "Sheet1!A:C", rowData)
	if err != nil {
		log.Printf("Gagal menulis ke Sheets: %v\n", err)
	} else {
		log.Println("Pesan berhasil dicatat ke Google Sheets!")
	}

	// Simpan juga ke Database lokal sebagai ActivityLog
	activity := model.ActivityLog{
		Type:        "INCOMING_MSG",
		Sender:      sender,
		Description: message,
	}
	if err := s.DB.Create(&activity).Error; err != nil {
		log.Printf("Gagal menyimpan activity log ke database: %v\n", err)
	}

	return nil
}
