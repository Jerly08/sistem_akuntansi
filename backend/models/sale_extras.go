package models

import (
	"time"
	"gorm.io/gorm"
)

// Payment related to a Sale
type SalePayment struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	SaleID        uint           `json:"sale_id" gorm:"not null;index"`
	PaymentID     *uint          `json:"payment_id" gorm:"index"` // Cross-reference to Payment Management
	PaymentNumber string         `json:"payment_number" gorm:"size:50"`
	Date          time.Time      `json:"date"`
	Amount        float64        `json:"amount" gorm:"type:decimal(15,2);default:0"`
	Method        string         `json:"method" gorm:"size:20"`
	Reference     string         `json:"reference" gorm:"size:50"`
	Notes         string         `json:"notes" gorm:"type:text"`
	CashBankID    *uint          `json:"cash_bank_id" gorm:"index"`
	AccountID     *uint          `json:"account_id" gorm:"index"`
	UserID        uint           `json:"user_id" gorm:"not null;index"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Sale     Sale     `json:"sale" gorm:"foreignKey:SaleID"`
	CashBank CashBank `json:"cash_bank" gorm:"foreignKey:CashBankID"`
	Account  Account  `json:"account" gorm:"foreignKey:AccountID"`
	User     User     `json:"user" gorm:"foreignKey:UserID"`
}

// Return related to a Sale
type SaleReturn struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	SaleID           uint           `json:"sale_id" gorm:"not null;index"`
	UserID           uint           `json:"user_id" gorm:"not null;index"`
	ApproverID       *uint          `json:"approver_id" gorm:"index"`
	ReturnNumber     string         `json:"return_number" gorm:"size:50"`
	Type             string         `json:"type" gorm:"size:20"`
	Date             time.Time      `json:"date"`
	Reason           string         `json:"reason" gorm:"type:text"`
	CreditNoteNumber string         `json:"credit_note_number" gorm:"size:50"`
	TotalAmount      float64        `json:"total_amount" gorm:"type:decimal(15,2);default:0"`
	Status           string         `json:"status" gorm:"size:20;default:'PENDING'"`
	Notes            string         `json:"notes" gorm:"type:text"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Sale        Sale            `json:"sale" gorm:"foreignKey:SaleID"`
	User        User            `json:"user" gorm:"foreignKey:UserID"`
	Approver    *User           `json:"approver" gorm:"foreignKey:ApproverID"`
	ReturnItems []SaleReturnItem `json:"return_items" gorm:"foreignKey:SaleReturnID"`
}

type SaleReturnItem struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	SaleReturnID uint           `json:"sale_return_id" gorm:"not null;index"`
	SaleItemID   uint           `json:"sale_item_id" gorm:"not null;index"`
	Quantity     int            `json:"quantity" gorm:"not null"`
	Reason       string         `json:"reason" gorm:"size:255"`
	UnitPrice    float64        `json:"unit_price" gorm:"type:decimal(15,2);default:0"`
	TotalAmount  float64        `json:"total_amount" gorm:"type:decimal(15,2);default:0"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	SaleReturn SaleReturn `json:"sale_return" gorm:"foreignKey:SaleReturnID"`
	SaleItem   SaleItem   `json:"sale_item" gorm:"foreignKey:SaleItemID"`
}

// Return types and statuses
const (
	ReturnTypeCreditNote = "CREDIT_NOTE"
	ReturnTypeRefund     = "REFUND"
)

const (
	ReturnStatusPending  = "PENDING"
	ReturnStatusApproved = "APPROVED"
	ReturnStatusRejected = "REJECTED"
)

// Reporting/Analytics DTOs

type CustomerSales struct {
	CustomerID   uint    `json:"customer_id"`
	CustomerName string  `json:"customer_name"`
	TotalAmount  float64 `json:"total_amount"`
	TotalOrders  int64   `json:"total_orders"`
}

type SalesSummaryResponse struct {
	TotalSales       int64           `json:"total_sales"`
	TotalAmount      float64         `json:"total_amount"`
	TotalPaid        float64         `json:"total_paid"`
	TotalOutstanding float64         `json:"total_outstanding"`
	AvgOrderValue    float64         `json:"avg_order_value"`
	TopCustomers     []CustomerSales `json:"top_customers"`
}

type SalesAnalyticsData struct {
	Period       string  `json:"period"`
	TotalSales   int64   `json:"total_sales"`
	TotalAmount  float64 `json:"total_amount"`
	GrowthRate   float64 `json:"growth_rate"`
}

type SalesAnalyticsResponse struct {
	Period string               `json:"period"`
	Data   []SalesAnalyticsData `json:"data"`
}

type ReceivableItem struct {
	SaleID            uint      `json:"sale_id"`
	InvoiceNumber     string    `json:"invoice_number"`
	CustomerName      string    `json:"customer_name"`
	Date              time.Time `json:"date"`
	DueDate           time.Time `json:"due_date"`
	TotalAmount       float64   `json:"total_amount"`
	PaidAmount        float64   `json:"paid_amount"`
	OutstandingAmount float64   `json:"outstanding_amount"`
	DaysOverdue       int       `json:"days_overdue"`
	Status            string    `json:"status"`
}

type ReceivablesReportResponse struct {
	TotalOutstanding float64          `json:"total_outstanding"`
	OverdueAmount    float64          `json:"overdue_amount"`
	Receivables      []ReceivableItem `json:"receivables"`
}
