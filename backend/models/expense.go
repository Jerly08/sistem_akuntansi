package models

import (
	"time"
	"gorm.io/gorm"
)

type Expense struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	ExpenseNo   string         `json:"expense_no" gorm:"unique;not null"`
	Category    string         `json:"category" gorm:"not null"` // operational, marketing, etc
	Description string         `json:"description" gorm:"not null"`
	Amount      float64        `json:"amount" gorm:"not null"`
	ExpenseDate time.Time      `json:"expense_date" gorm:"not null"`
	PaymentMethod string       `json:"payment_method"` // cash, transfer
	Notes       string         `json:"notes"`
	UserID      uint           `json:"user_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	User User `json:"user,omitempty"`
}

type Asset struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	AssetNo       string         `json:"asset_no" gorm:"unique;not null"`
	Name          string         `json:"name" gorm:"not null"`
	Category      string         `json:"category" gorm:"not null"` // building, equipment, vehicle, etc
	Description   string         `json:"description"`
	PurchasePrice float64        `json:"purchase_price" gorm:"not null"`
	CurrentValue  float64        `json:"current_value"`
	PurchaseDate  time.Time      `json:"purchase_date" gorm:"not null"`
	Condition     string         `json:"condition" gorm:"default:'good'"` // good, fair, poor
	Location      string         `json:"location"`
	IsActive      bool           `json:"is_active" gorm:"default:true"`
	UserID        uint           `json:"user_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	User User `json:"user,omitempty"`
}

type CashBank struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	AccountName string         `json:"account_name" gorm:"not null"`
	AccountType string         `json:"account_type" gorm:"not null"` // cash, bank
	AccountNo   string         `json:"account_no"`
	BankName    string         `json:"bank_name"`
	Balance     float64        `json:"balance" gorm:"default:0"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type Account struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"unique;not null"`
	Name        string         `json:"name" gorm:"not null"`
	Type        string         `json:"type" gorm:"not null"` // asset, liability, equity, revenue, expense
	ParentID    *uint          `json:"parent_id"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	Parent   *Account `json:"parent,omitempty"`
	Children []Account `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

type Inventory struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	ProductID      uint           `json:"product_id"`
	TransactionType string        `json:"transaction_type"` // in, out, adjustment
	Quantity       int            `json:"quantity"`
	ReferenceType  string         `json:"reference_type"` // sale, purchase, adjustment
	ReferenceID    uint           `json:"reference_id"`
	Notes          string         `json:"notes"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	Product Product `json:"product,omitempty"`
}
