package models

import (
	"time"
	"gorm.io/gorm"
)

type Purchase struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	PurchaseNo   string         `json:"purchase_no" gorm:"unique;not null"`
	SupplierName string         `json:"supplier_name" gorm:"not null"`
	SupplierEmail string        `json:"supplier_email"`
	SupplierPhone string        `json:"supplier_phone"`
	PurchaseDate time.Time      `json:"purchase_date" gorm:"not null"`
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
	User          User           `json:"user,omitempty"`
	PurchaseItems []PurchaseItem `json:"purchase_items,omitempty"`
}

type PurchaseItem struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	PurchaseID uint           `json:"purchase_id"`
	ProductID  uint           `json:"product_id"`
	Quantity   int            `json:"quantity" gorm:"not null"`
	Price      float64        `json:"price" gorm:"not null"`
	Total      float64        `json:"total" gorm:"not null"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	Purchase Purchase `json:"purchase,omitempty"`
	Product  Product  `json:"product,omitempty"`
}
