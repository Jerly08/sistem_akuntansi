package models

import (
	"time"
	"gorm.io/gorm"
)

type Product struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"unique;not null"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Category    string         `json:"category"`
	Unit        string         `json:"unit" gorm:"not null"` // pcs, kg, liter, etc
	PurchasePrice float64      `json:"purchase_price" gorm:"not null"`
	SalePrice     float64      `json:"sale_price" gorm:"not null"`
	Stock         int          `json:"stock" gorm:"default:0"`
	MinStock      int          `json:"min_stock" gorm:"default:0"`
	IsActive      bool         `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	SaleItems     []SaleItem     `json:"sale_items,omitempty"`
	PurchaseItems []PurchaseItem `json:"purchase_items,omitempty"`
	Inventories   []Inventory    `json:"inventories,omitempty"`
}
