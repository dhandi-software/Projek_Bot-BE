package model

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	SKU         string  `gorm:"default:''" json:"sku"`
	Title       string  `gorm:"default:''" json:"title"`
	Category    string  `gorm:"default:''" json:"category"`
	Price       float64 `gorm:"default:0" json:"price"`
	Stock       int     `gorm:"default:0" json:"stock"`
	Materials   string  `gorm:"default:''" json:"materials"`
	Brand       string  `gorm:"default:''" json:"brand"`
	Description string  `gorm:"default:''" json:"description"`
	Image       string  `gorm:"default:''" json:"image"`
	IsFeatured  bool    `gorm:"default:false" json:"is_featured"`
	IsActive    bool    `gorm:"default:true" json:"is_active"`
}
