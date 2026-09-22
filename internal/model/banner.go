package model

import "gorm.io/gorm"

type Banner struct {
	gorm.Model
	Title      string `gorm:"not null" json:"title"`
	Subtitle   string `json:"subtitle"`
	Tagline    string `json:"tagline"`
	Image      string `gorm:"not null" json:"image"`
	PriceBadge string `json:"price_badge"`
	LinkUrl    string `json:"link_url"`
	ButtonText string `json:"button_text"`
	BgColor    string `json:"bg_color"`
	IsActive   bool   `gorm:"default:true" json:"is_active"`
	SortOrder  int    `gorm:"default:0" json:"sort_order"`
}
