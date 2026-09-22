package model

import "gorm.io/gorm"

// Customer merepresentasikan tabel customer di database
type Customer struct {
	gorm.Model
	Name     string `gorm:"not null" json:"name"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Phone    string `json:"phone"`
	Password string `gorm:"not null" json:"-"` // Hash bcrypt
	Address  string `json:"address"`
	Bio      string `json:"bio"`
	Photo    string `json:"photo"`
}
