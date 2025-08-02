package models

import (
	"time"
	"gorm.io/gorm"
)

type Product struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Code          string         `json:"code" gorm:"unique;not null;size:20"`
	Name          string         `json:"name" gorm:"not null;size:100"`
	Description   string         `json:"description" gorm:"type:text"`
	CategoryID    *uint          `json:"category_id" gorm:"index"`
	Brand         string         `json:"brand" gorm:"size:50"`
	Model         string         `json:"model" gorm:"size:50"`
	Unit          string         `json:"unit" gorm:"not null;size:20"` // pcs, kg, liter, etc
	PurchasePrice float64        `json:"purchase_price" gorm:"type:decimal(15,2);default:0"`
	SalePrice     float64        `json:"sale_price" gorm:"type:decimal(15,2);default:0"`
	Stock         int            `json:"stock" gorm:"default:0"`
	MinStock      int            `json:"min_stock" gorm:"default:0"`
	MaxStock      int            `json:"max_stock" gorm:"default:0"`
	ReorderLevel  int            `json:"reorder_level" gorm:"default:0"`
	Barcode       string         `json:"barcode" gorm:"size:50"`
	SKU           string         `json:"sku" gorm:"size:50"`
	Weight        float64        `json:"weight" gorm:"type:decimal(10,3);default:0"`
	Dimensions    string         `json:"dimensions" gorm:"size:100"`
	IsActive      bool           `json:"is_active" gorm:"default:true"`
	IsService     bool           `json:"is_service" gorm:"default:false"`
	Taxable       bool           `json:"taxable" gorm:"default:true"`
	Notes         string         `json:"notes" gorm:"type:text"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	Category      *ProductCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	SaleItems     []SaleItem       `json:"-" gorm:"foreignKey:ProductID"`
	PurchaseItems []PurchaseItem   `json:"-" gorm:"foreignKey:ProductID"`
	Inventories   []Inventory      `json:"-" gorm:"foreignKey:ProductID"`
}

type ProductCategory struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"unique;not null;size:20"`
	Name        string         `json:"name" gorm:"not null;size:100"`
	Description string         `json:"description" gorm:"type:text"`
	ParentID    *uint          `json:"parent_id" gorm:"index"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Parent   *ProductCategory   `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []ProductCategory  `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Products []Product          `json:"-" gorm:"foreignKey:CategoryID"`
}

type Inventory struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	ProductID     uint           `json:"product_id" gorm:"not null;index"`
	ReferenceType string         `json:"reference_type" gorm:"size:50"` // SALE, PURCHASE, ADJUSTMENT, etc.
	ReferenceID   uint           `json:"reference_id" gorm:"index"`
	Type          string         `json:"type" gorm:"not null;size:20"` // IN, OUT
	Quantity      int            `json:"quantity" gorm:"not null"`
	UnitCost      float64        `json:"unit_cost" gorm:"type:decimal(15,2);default:0"`
	TotalCost     float64        `json:"total_cost" gorm:"type:decimal(15,2);default:0"`
	RemainingQty  int            `json:"remaining_qty" gorm:"default:0"`
	Notes         string         `json:"notes" gorm:"type:text"`
	TransactionDate time.Time    `json:"transaction_date"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Product Product `json:"product" gorm:"foreignKey:ProductID"`
}

// Inventory Types Constants
const (
	InventoryTypeIn  = "IN"
	InventoryTypeOut = "OUT"
)
