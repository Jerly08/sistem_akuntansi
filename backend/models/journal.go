package models

import (
	"time"
	"gorm.io/gorm"
)

type Journal struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Code         string         `json:"code" gorm:"unique;not null;size:20"`
	Date         time.Time      `json:"date"`
	Description  string         `json:"description" gorm:"not null;type:text"`
	ReferenceType string        `json:"reference_type" gorm:"size:50"` // MANUAL, SALE, PURCHASE, PAYMENT, etc.
	ReferenceID   *uint         `json:"reference_id" gorm:"index"`
	UserID       uint           `json:"user_id" gorm:"not null;index"`
	Status       string         `json:"status" gorm:"not null;size:20;default:'PENDING'"` // PENDING, POSTED, CANCELLED
	TotalDebit   float64        `json:"total_debit" gorm:"type:decimal(20,2);default:0"`
	TotalCredit  float64        `json:"total_credit" gorm:"type:decimal(20,2);default:0"`
	IsAdjusting  bool           `json:"is_adjusting" gorm:"default:false"`
	Period       string         `json:"period" gorm:"size:7"` // YYYY-MM format
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User         User            `json:"user" gorm:"foreignKey:UserID"`
	JournalEntries []JournalEntry `json:"journal_entries" gorm:"foreignKey:JournalID"`
}

type JournalEntry struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	JournalID   uint           `json:"journal_id" gorm:"not null;index"`
	AccountID   uint           `json:"account_id" gorm:"not null;index"`
	Description string         `json:"description" gorm:"type:text"`
	DebitAmount  float64       `json:"debit_amount" gorm:"type:decimal(20,2);default:0"`
	CreditAmount float64       `json:"credit_amount" gorm:"type:decimal(20,2);default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Journal Journal `json:"journal" gorm:"foreignKey:JournalID"`
	Account Account `json:"account" gorm:"foreignKey:AccountID"`
}

// Journal Status Constants
const (
	JournalStatusPending   = "PENDING"
	JournalStatusPosted    = "POSTED"
	JournalStatusCancelled = "CANCELLED"
)

// Journal Reference Types Constants
const (
	JournalRefTypeManual   = "MANUAL"
	JournalRefTypeSale     = "SALE"
	JournalRefTypePurchase = "PURCHASE"
	JournalRefTypePayment  = "PAYMENT"
	JournalRefTypeExpense  = "EXPENSE"
	JournalRefTypeAsset    = "ASSET"
	JournalRefTypeAdjustment = "ADJUSTMENT"
)
