package models

import (
    "time"
    "gorm.io/gorm"
)

type Payment struct {
    ID           uint           `json:"id" gorm:"primaryKey"`
    Code         string         `json:"code" gorm:"unique;not null;size:20"`
    ContactID    uint           `json:"contact_id" gorm:"not null;index"`
    UserID       uint           `json:"user_id" gorm:"not null;index"`
    Date         time.Time      `json:"date"`
    Amount       float64        `json:"amount" gorm:"type:decimal(15,2);default:0"`
    Method       string         `json:"method" gorm:"size:20"` // CASH, BANK
    Reference    string         `json:"reference" gorm:"size:50"`
    Status       string         `json:"status" gorm:"size:20"` // PENDING, COMPLETED, FAILED
    Notes        string         `json:"notes" gorm:"type:text"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

    // Relations
    Contact  Contact `json:"contact" gorm:"foreignKey:ContactID"`
    User     User    `json:"user" gorm:"foreignKey:UserID"`
}

type CashBank struct {
    ID           uint           `json:"id" gorm:"primaryKey"`
    Code         string         `json:"code" gorm:"unique;not null;size:20"`
    Name         string         `json:"name" gorm:"not null;size:100"`
    Type         string         `json:"type" gorm:"not null;size:20"` // CASH, BANK
    Balance      float64        `json:"balance" gorm:"type:decimal(20,2);default:0"`
    IsActive     bool           `json:"is_active" gorm:"default:true"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

    // Relations
    Transactions []CashBankTransaction `json:"transactions" gorm:"foreignKey:CashBankID"`
}

type CashBankTransaction struct {
    ID           uint           `json:"id" gorm:"primaryKey"`
    CashBankID   uint           `json:"cash_bank_id" gorm:"not null;index"`
    ReferenceType string        `json:"reference_type" gorm:"size:50"` // PAYMENT, TRANSFER
    ReferenceID   uint          `json:"reference_id" gorm:"index"`
    Amount        float64       `json:"amount" gorm:"type:decimal(20,2);default:0"`
    BalanceAfter  float64       `json:"balance_after" gorm:"type:decimal(20,2);default:0"`
    TransactionDate time.Time   `json:"transaction_date"`
    Notes         string        `json:"notes" gorm:"type:text"`
    CreatedAt     time.Time     `json:"created_at"`
    UpdatedAt     time.Time     `json:"updated_at"`
    DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

    // Relations
    CashBank CashBank `json:"cash_bank" gorm:"foreignKey:CashBankID"`
}

// Payment Status Constants
const (
    PaymentStatusPending   = "PENDING"
    PaymentStatusCompleted = "COMPLETED"
    PaymentStatusFailed    = "FAILED"
)

// CashBank Types Constants
const (
    CashBankTypeCash = "CASH"
    CashBankTypeBank = "BANK"
)
