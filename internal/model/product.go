package model

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	SKU               string         `gorm:"default:''" json:"sku"`
	Title             string         `gorm:"default:''" json:"title"`
	Category          string         `gorm:"default:''" json:"category"`
	Price             float64        `gorm:"default:0" json:"price"`
	DiscountPrice     float64        `gorm:"default:0" json:"discount_price"`
	Stock             int            `gorm:"default:0" json:"stock"`
	LowStockThreshold int            `gorm:"default:10" json:"low_stock_threshold"`
	Weight            float64        `gorm:"default:0" json:"weight"`
	Materials         string         `gorm:"default:''" json:"materials"`
	Brand             string         `gorm:"default:''" json:"brand"`
	ShortDescription  string         `gorm:"default:''" json:"short_description"`
	Description       string         `gorm:"default:''" json:"description"`
	Image             string         `gorm:"default:''" json:"image"`
	Status            string         `gorm:"default:'active'" json:"status"`
	IsFeatured        bool           `gorm:"default:false" json:"is_featured"`
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	IsBestDeal        bool           `gorm:"default:false" json:"is_best_deal"`
	BestDealStartedAt *time.Time     `json:"best_deal_started_at"`
	BestDealExpiresAt *time.Time     `json:"best_deal_expires_at"`
	Features          string         `gorm:"type:text;default:''" json:"features"`
	Colors            string         `gorm:"type:text;default:''" json:"colors"`
	ShippingInfo      string         `gorm:"type:text;default:''" json:"shipping_info"`
	AdditionalInfo    string         `gorm:"type:text;default:''" json:"additional_info"`
	Specifications    string         `gorm:"type:text;default:''" json:"specifications"`
}



