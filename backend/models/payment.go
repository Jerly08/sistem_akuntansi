package models

import (
    "time"
    "gorm.io/gorm"
)

type Payment struct {
    ID              uint           `json:"id" gorm:"primaryKey"`
    Code            string         `json:"code" gorm:"unique;not null;size:20"`
    ContactID       uint           `json:"contact_id" gorm:"not null;index"`
    UserID          uint           `json:"user_id" gorm:"not null;index"`
    Date            time.Time      `json:"date"`
    Amount          float64        `json:"amount" gorm:"type:decimal(15,2);default:0"`
    Method          string         `json:"method" gorm:"size:20"` // CASH, BANK_TRANSFER, CHECK, etc.
    Reference       string         `json:"reference" gorm:"size:50"`
    Status          string         `json:"status" gorm:"size:20"` // PENDING, COMPLETED, FAILED, REVERSED
    Notes           string         `json:"notes" gorm:"type:text"`
    JournalEntryID  *uint          `json:"journal_entry_id" gorm:"index"`  // Link to SSOT journal entry
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`
    DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

    // Relations
    Contact  Contact `json:"contact" gorm:"foreignKey:ContactID"`
    User     User    `json:"user" gorm:"foreignKey:UserID"`
}

// PaymentAllocation represents payment allocation to invoices or bills
type PaymentAllocation struct {
    ID              uint           `json:"id" gorm:"primaryKey"`
    PaymentID       uint64         `json:"payment_id" gorm:"not null;index"`
    InvoiceID       *uint          `json:"invoice_id" gorm:"index"`
    BillID          *uint          `json:"bill_id" gorm:"index"`
    AllocatedAmount float64        `json:"allocated_amount" gorm:"type:decimal(15,2);default:0"`
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`

    // Relations
    Payment Payment `json:"payment" gorm:"foreignKey:PaymentID"`
    Sale    *Sale   `json:"sale,omitempty" gorm:"foreignKey:InvoiceID"`
}

// Payment Status Constants
const (
    PaymentStatusPending   = "PENDING"
    PaymentStatusCompleted = "COMPLETED"
    PaymentStatusFailed    = "FAILED"
    PaymentStatusReversed  = "REVERSED"
)

// Payment Method Constants
const (
    PaymentMethodCash         = "CASH"
    PaymentMethodBankTransfer = "BANK_TRANSFER"
    PaymentMethodCheck        = "CHECK"
    PaymentMethodCreditCard   = "CREDIT_CARD"
    PaymentMethodDebitCard    = "DEBIT_CARD"
    PaymentMethodDigitalWallet = "DIGITAL_WALLET"
)
