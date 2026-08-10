package provider

import (
	"fmt"
	"log"

	"bot_be/internal/config"
	"bot_be/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDatabase menginisialisasi koneksi PostgreSQL dan menjalankan AutoMigrate
func InitDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, fmt.Errorf("gagal terhubung ke database postgres: %w", err)
	}

	log.Println("Berhasil terhubung ke database PostgreSQL")

	// Jalankan Auto Migrate
	err = db.AutoMigrate(&model.Admin{}, &model.AppConfig{}, &model.ActivityLog{})
	if err != nil {
		return nil, fmt.Errorf("gagal melakukan migrasi tabel: %w", err)
	}

	log.Println("Migrasi tabel berhasil")

	// Buat akun admin default jika belum ada
	seedAdmin(db)

	return db, nil
}

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&model.Admin{}).Count(&count)
	if count == 0 {
		// Gunakan default password "admin123" untuk akun pertama kali
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		
		admin := model.Admin{
			Username: "admin",
			Password: string(hashedPassword),
		}
		db.Create(&admin)
		log.Println("Akun admin default berhasil dibuat (Username: admin, Password: admin123)")
	}
}
