package model

import "time"

type Category struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	Name        string     `gorm:"uniqueIndex;not null" json:"name"`
	Slug        string     `gorm:"default:''" json:"slug"`
	Description string     `gorm:"default:''" json:"description"`
	Icon        string     `gorm:"default:''" json:"icon"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
}
