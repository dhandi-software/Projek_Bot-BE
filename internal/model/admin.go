package model

import "gorm.io/gorm"

// Admin merepresentasikan tabel admin di database
type Admin struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"` // Disimpan dalam bentuk hash (bcrypt)
}
