package model

import "gorm.io/gorm"

type ActivityLog struct {
	gorm.Model
	Type        string `gorm:"not null"` // "INCOMING_MSG", "SYSTEM"
	Sender      string `gorm:"not null"`
	Description string `gorm:"not null"`
}
