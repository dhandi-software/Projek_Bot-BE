package model

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	OrderID         string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"order_id"`
	OrderNumber     string         `gorm:"type:varchar(100);default:''" json:"order_number"`
	IdempotencyKey  string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"idempotency_key"`
	CustomerID      uint           `gorm:"index" json:"customer_id"`
	UserID          uint           `gorm:"index;default:0" json:"user_id"`
	CustomerName    string         `gorm:"type:varchar(255);default:''" json:"customer_name"`
	CustomerEmail   string         `gorm:"type:varchar(255);default:''" json:"customer_email"`
	CustomerPhone   string         `gorm:"type:varchar(50);default:''" json:"customer_phone"`
	ShippingAddress string         `gorm:"type:text;default:''" json:"shipping_address"`
	TotalAmount     float64        `gorm:"type:numeric(15,2);not null" json:"total_amount"`
	TotalPrice      float64        `gorm:"type:numeric(15,2);default:0" json:"total_price"`
	Status          string         `gorm:"type:varchar(50);default:'pending';index" json:"status"`
	SnapToken       string         `gorm:"type:text;default:''" json:"snap_token"`
	SnapRedirectURL string         `gorm:"type:text;default:''" json:"snap_redirect_url"`
	PaymentType     string         `gorm:"type:varchar(50);default:''" json:"payment_type"`
	PaidAt          *time.Time     `json:"paid_at,omitempty"`
	OrderItems      []OrderItem    `gorm:"foreignKey:OrderID;references:OrderID" json:"items"`
}

type OrderItem struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	OrderID   string  `gorm:"type:varchar(100);index;not null" json:"order_id"`
	ProductID uint    `gorm:"index;not null" json:"product_id"`
	Title     string  `gorm:"type:varchar(255);not null" json:"title"`
	Quantity  int     `gorm:"not null" json:"quantity"`
	Price     float64 `gorm:"type:numeric(15,2);not null" json:"price"`
}
