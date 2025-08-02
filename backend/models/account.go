package models

import (
	"time"
	"gorm.io/gorm"
)

type Account struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"unique;not null;size:20"`
	Name        string         `json:"name" gorm:"not null;size:100"`
	Description string         `json:"description" gorm:"type:text"`
	Type        string         `json:"type" gorm:"not null;size:20"` // ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
	Category    string         `json:"category" gorm:"size:50"`      // CURRENT_ASSET, FIXED_ASSET, etc.
	ParentID    *uint          `json:"parent_id" gorm:"index"`
	Level       int            `json:"level" gorm:"default:1"`
	IsHeader    bool           `json:"is_header" gorm:"default:false"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	Balance     float64        `json:"balance" gorm:"type:decimal(20,2);default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Parent       *Account          `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children     []Account         `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Transactions []Transaction     `json:"-" gorm:"foreignKey:AccountID"`
	SaleItems    []SaleItem        `json:"-" gorm:"foreignKey:RevenueAccountID"`
	PurchaseItems []PurchaseItem   `json:"-" gorm:"foreignKey:ExpenseAccountID"`
	Assets       []Asset           `json:"-" gorm:"foreignKey:AssetAccountID"`
}

type Transaction struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	AccountID     uint           `json:"account_id" gorm:"not null;index"`
	JournalID     *uint          `json:"journal_id" gorm:"index"`
	ReferenceType string         `json:"reference_type" gorm:"size:50"` // SALE, PURCHASE, PAYMENT, etc.
	ReferenceID   uint           `json:"reference_id" gorm:"index"`
	Description   string         `json:"description" gorm:"type:text"`
	DebitAmount   float64        `json:"debit_amount" gorm:"type:decimal(20,2);default:0"`
	CreditAmount  float64        `json:"credit_amount" gorm:"type:decimal(20,2);default:0"`
	TransactionDate time.Time    `json:"transaction_date"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Account Account `json:"account" gorm:"foreignKey:AccountID"`
	Journal *Journal `json:"journal,omitempty" gorm:"foreignKey:JournalID"`
}

// Account Types Constants
const (
	AccountTypeAsset     = "ASSET"
	AccountTypeLiability = "LIABILITY"
	AccountTypeEquity    = "EQUITY"
	AccountTypeRevenue   = "REVENUE"
	AccountTypeExpense   = "EXPENSE"
)

// Account Categories Constants
const (
	CategoryCurrentAsset    = "CURRENT_ASSET"
	CategoryFixedAsset      = "FIXED_ASSET"
	CategoryCurrentLiability = "CURRENT_LIABILITY"
	CategoryLongTermLiability = "LONG_TERM_LIABILITY"
	CategoryEquity          = "EQUITY"
	CategoryOperatingRevenue = "OPERATING_REVENUE"
	CategoryOtherRevenue    = "OTHER_REVENUE"
	CategoryOperatingExpense = "OPERATING_EXPENSE"
	CategoryOtherExpense    = "OTHER_EXPENSE"
)
