package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	ClientURL          string
	CORSAllowedOrigins string
	SpreadsheetID      string
	DBHost             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBPort             string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan environment default")
	}

	return &Config{
		Port:               getEnv("PORT", "8080"),
		ClientURL:          getEnv("CLIENT_URL", "http://localhost:5173"),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000,http://127.0.0.1:5173,http://localhost:5174"),
		SpreadsheetID:      getEnv("SPREADSHEET_ID", ""),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", "your_db_password"),
		DBName:             getEnv("DB_NAME", "your_db_name"),
		DBPort:             getEnv("DB_PORT", "5432"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
