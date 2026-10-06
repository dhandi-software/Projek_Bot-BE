package model

import (
	"time"

	"gorm.io/gorm"
)

type BrowsingHistory struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserID    uint           `gorm:"index;default:0" json:"user_id"`
	UserEmail string         `gorm:"index;size:255;default:''" json:"user_email"`
	SessionID string         `gorm:"index;size:255;default:''" json:"session_id"`
	ProductID uint           `gorm:"index;not null" json:"product_id"`
	Product   Product        `gorm:"foreignKey:ProductID" json:"product"`
	ViewedAt  time.Time      `gorm:"index" json:"viewed_at"`
}
