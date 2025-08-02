package models

import (
	"time"
	"gorm.io/gorm"
)

type Sale struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Code         string         `json:"code" gorm:"unique;not null;size:20"`
	CustomerID   uint           `json:"customer_id" gorm:"not null;index"`
	UserID       uint           `json:"user_id" gorm:"not null;index"`
	Date         time.Time      `json:"date"`
	DueDate      time.Time      `json:"due_date"`
	TotalAmount  float64        `json:"total_amount" gorm:"type:decimal(15,2);default:0"`
	Discount     float64        `json:"discount" gorm:"type:decimal(8,2);default:0"`
	Tax          float64        `json:"tax" gorm:"type:decimal(8,2);default:0"`
	Status       string         `json:"status" gorm:"size:20"` // PENDING, COMPLETED, CANCELLED
	Notes        string         `json:"notes" gorm:"type:text"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Customer  Contact    `json:"customer" gorm:"foreignKey:CustomerID"`
	User      User       `json:"user" gorm:"foreignKey:UserID"`
	SaleItems []SaleItem `json:"sale_items" gorm:"foreignKey:SaleID"`
}

type SaleItem struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	SaleID         uint           `json:"sale_id" gorm:"not null;index"`
	ProductID      uint           `json:"product_id" gorm:"not null;index"`
	Quantity       int            `json:"quantity" gorm:"not null"`
	UnitPrice      float64        `json:"unit_price" gorm:"type:decimal(15,2);default:0"`
	TotalPrice     float64        `json:"total_price" gorm:"type:decimal(15,2);default:0"`
	Discount       float64        `json:"discount" gorm:"type:decimal(8,2);default:0"`
	Tax            float64        `json:"tax" gorm:"type:decimal(8,2);default:0"`
	RevenueAccountID uint         `json:"revenue_account_id" gorm:"index"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Sale          Sale     `json:"sale" gorm:"foreignKey:SaleID"`
	Product       Product  `json:"product" gorm:"foreignKey:ProductID"`
	RevenueAccount Account `json:"revenue_account" gorm:"foreignKey:RevenueAccountID"`
}

// Sale Status Constants
const (
	SaleStatusPending   = "PENDING"
	SaleStatusCompleted = "COMPLETED"
	SaleStatusCancelled = "CANCELLED"
)
