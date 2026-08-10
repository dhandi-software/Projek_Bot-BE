package model

import "gorm.io/gorm"

// AppConfig merepresentasikan tabel konfigurasi dinamis aplikasi
type AppConfig struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex;not null"`
	Value string `gorm:"not null"`
}
