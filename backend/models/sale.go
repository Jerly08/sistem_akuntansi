package models

import (
	"time"
	"gorm.io/gorm"
)

type Sale struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	InvoiceNo    string         `json:"invoice_no" gorm:"unique;not null"`
	CustomerName string         `json:"customer_name"`
	CustomerEmail string        `json:"customer_email"`
	CustomerPhone string        `json:"customer_phone"`
	SaleDate     time.Time      `json:"sale_date" gorm:"not null"`
	SubTotal     float64        `json:"sub_total" gorm:"not null"`
	Tax          float64        `json:"tax" gorm:"default:0"`
	Discount     float64        `json:"discount" gorm:"default:0"`
	Total        float64        `json:"total" gorm:"not null"`
	PaymentMethod string        `json:"payment_method"` // cash, transfer, credit
	PaymentStatus string        `json:"payment_status" gorm:"default:'pending'"` // pending, paid, partial
	Notes        string         `json:"notes"`
	UserID       uint           `json:"user_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	User      User       `json:"user,omitempty"`
	SaleItems []SaleItem `json:"sale_items,omitempty"`
}

type SaleItem struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	SaleID    uint           `json:"sale_id"`
	ProductID uint           `json:"product_id"`
	Quantity  int            `json:"quantity" gorm:"not null"`
	Price     float64        `json:"price" gorm:"not null"`
	Total     float64        `json:"total" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	Sale    Sale    `json:"sale,omitempty"`
	Product Product `json:"product,omitempty"`
}
