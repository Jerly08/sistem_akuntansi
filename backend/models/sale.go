package models

import (
	"time"
	"gorm.io/gorm"
)

type Sale struct {
	ID                 uint            `json:"id" gorm:"primaryKey"`
	Code               string          `json:"code" gorm:"unique;not null;size:20"`
	CustomerID         uint            `json:"customer_id" gorm:"not null;index"`
	UserID             uint            `json:"user_id" gorm:"not null;index"`
	SalesPersonID      *uint           `json:"sales_person_id" gorm:"index"`
	Type               string          `json:"type" gorm:"size:20;default:'INVOICE'"` // QUOTATION, ORDER, INVOICE
	Date               time.Time       `json:"date"`
	DueDate            time.Time       `json:"due_date"`
	ValidUntil         *time.Time      `json:"valid_until"`
	QuotationNumber    string          `json:"quotation_number" gorm:"size:50"`
	InvoiceNumber      string          `json:"invoice_number" gorm:"size:50"`
	Currency           string          `json:"currency" gorm:"size:5;default:'IDR'"`
	ExchangeRate       float64         `json:"exchange_rate" gorm:"type:decimal(12,6);default:1"`
	TotalAmount        float64         `json:"total_amount" gorm:"type:decimal(15,2);default:0"`
	PaidAmount         float64         `json:"paid_amount" gorm:"type:decimal(15,2);default:0"`
	OutstandingAmount  float64         `json:"outstanding_amount" gorm:"type:decimal(15,2);default:0"`
	Subtotal           float64         `json:"subtotal" gorm:"type:decimal(15,2);default:0"`
	SubTotal           float64         `json:"sub_total" gorm:"-"` // Read-only alias for frontend consistency
	DiscountPercent    float64         `json:"discount_percent" gorm:"type:decimal(5,2);default:0"`
	DiscountAmount     float64         `json:"discount_amount" gorm:"type:decimal(15,2);default:0"`
	TaxableAmount      float64         `json:"taxable_amount" gorm:"type:decimal(15,2);default:0"`
	Tax                float64         `json:"tax" gorm:"type:decimal(15,2);default:0"`
	PPN                float64         `json:"ppn" gorm:"type:decimal(15,2);default:0"`
	PPNPercent         float64         `json:"ppn_percent" gorm:"type:decimal(5,2);default:11"`
	PPh                float64         `json:"pph" gorm:"type:decimal(15,2);default:0"`
	PPhPercent         float64         `json:"pph_percent" gorm:"type:decimal(5,2);default:0"`
	PPhType            string          `json:"pph_type" gorm:"size:20"`
	TotalTax           float64         `json:"total_tax" gorm:"type:decimal(15,2);default:0"`
	PaymentTerms       string          `json:"payment_terms" gorm:"size:50"`
	PaymentMethod      string          `json:"payment_method" gorm:"size:50"`
	ShippingMethod     string          `json:"shipping_method" gorm:"size:50"`
	ShippingCost       float64         `json:"shipping_cost" gorm:"type:decimal(15,2);default:0"`
	ShippingTaxable    bool            `json:"shipping_taxable" gorm:"default:false"`
	BillingAddress     string          `json:"billing_address" gorm:"type:text"`
	ShippingAddress    string          `json:"shipping_address" gorm:"type:text"`
	Status             string          `json:"status" gorm:"size:20"` // DRAFT, PENDING, CONFIRMED, CANCELLED
	Notes              string          `json:"notes" gorm:"type:text"`
	InternalNotes      string          `json:"internal_notes" gorm:"type:text"`
	Reference          string          `json:"reference" gorm:"size:100"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	DeletedAt          gorm.DeletedAt  `json:"-" gorm:"index"`

	// Relations
	Customer     Contact       `json:"customer" gorm:"foreignKey:CustomerID"`
	User         User          `json:"user" gorm:"foreignKey:UserID"`
	SalesPerson  *Contact      `json:"sales_person" gorm:"foreignKey:SalesPersonID"`
	SaleItems    []SaleItem    `json:"sale_items" gorm:"foreignKey:SaleID"`
	SalePayments []SalePayment `json:"sale_payments" gorm:"foreignKey:SaleID"`
	SaleReturns  []SaleReturn  `json:"sale_returns" gorm:"foreignKey:SaleID"`
}

type SaleItem struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	SaleID           uint           `json:"sale_id" gorm:"not null;index"`
	ProductID        uint           `json:"product_id" gorm:"not null;index"`
	Description      string         `json:"description" gorm:"type:text"`
	Quantity         int            `json:"quantity" gorm:"not null"`
	UnitPrice        float64        `json:"unit_price" gorm:"type:decimal(15,2);default:0"`
	DiscountPercent  float64        `json:"discount_percent" gorm:"type:decimal(5,2);default:0"`
	DiscountAmount   float64        `json:"discount_amount" gorm:"type:decimal(15,2);default:0"`
	LineTotal        float64        `json:"line_total" gorm:"type:decimal(15,2);default:0"`
	Taxable          bool           `json:"taxable" gorm:"default:true"`
	PPNAmount        float64        `json:"ppn_amount" gorm:"type:decimal(15,2);default:0"`
	PPhAmount        float64        `json:"pph_amount" gorm:"type:decimal(15,2);default:0"`
	TotalTax         float64        `json:"total_tax" gorm:"type:decimal(15,2);default:0"`
	FinalAmount      float64        `json:"final_amount" gorm:"type:decimal(15,2);default:0"`
	// Computed fields - these are calculated and should not be set directly
	TotalPrice       float64        `json:"total_price" gorm:"type:decimal(15,2);default:0;->"` // Read-only: Same as LineTotal
	Discount         float64        `json:"discount" gorm:"type:decimal(15,2);default:0;->"`      // Read-only: Legacy field
	Tax              float64        `json:"tax" gorm:"type:decimal(15,2);default:0;->"`           // Read-only: Legacy field
	RevenueAccountID uint           `json:"revenue_account_id" gorm:"index"`
	TaxAccountID     *uint          `json:"tax_account_id" gorm:"index"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Sale          Sale     `json:"-" gorm:"foreignKey:SaleID"` // Exclude to prevent circular reference
	Product       Product  `json:"product" gorm:"foreignKey:ProductID"`
	RevenueAccount Account `json:"revenue_account" gorm:"foreignKey:RevenueAccountID"`
	TaxAccount    *Account `json:"tax_account" gorm:"foreignKey:TaxAccountID"`
}

// Sale Type Constants
const (
	SaleTypeQuotation = "QUOTATION"
	SaleTypeOrder     = "ORDER"
	SaleTypeInvoice   = "INVOICE"
)

// Sale Status Constants
const (
	SaleStatusDraft     = "DRAFT"
	SaleStatusPending   = "PENDING"
	SaleStatusConfirmed = "CONFIRMED"
	SaleStatusCompleted = "COMPLETED"
	SaleStatusCancelled = "CANCELLED"
	SaleStatusInvoiced  = "INVOICED"
	SaleStatusOverdue   = "OVERDUE"
	SaleStatusPaid      = "PAID"
)

// Filter and Request DTOs
type SalesFilter struct {
	Status     string `json:"status"`
	CustomerID string `json:"customer_id"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Search     string `json:"search"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

type SaleCreateRequest struct {
	CustomerID       uint                `json:"customer_id" binding:"required"`
	SalesPersonID    *uint               `json:"sales_person_id"`
	Type             string              `json:"type" binding:"required"`
	Date             time.Time           `json:"date" binding:"required"`
	DueDate          time.Time           `json:"due_date"`
	ValidUntil       *time.Time          `json:"valid_until"`
	Currency         string              `json:"currency"`
	ExchangeRate     *float64            `json:"exchange_rate"`
	DiscountPercent  float64             `json:"discount_percent"`
	PPNPercent       *float64            `json:"ppn_percent"`
	PPhPercent       float64             `json:"pph_percent"`
	PPhType          string              `json:"pph_type"`
	PaymentTerms     string              `json:"payment_terms"`
	PaymentMethod    string              `json:"payment_method"`
	ShippingMethod   string              `json:"shipping_method"`
	ShippingCost     float64             `json:"shipping_cost"`
	ShippingTaxable  bool                `json:"shipping_taxable"`
	BillingAddress   string              `json:"billing_address"`
	ShippingAddress  string              `json:"shipping_address"`
	Notes            string              `json:"notes"`
	InternalNotes    string              `json:"internal_notes"`
	Reference        string              `json:"reference"`
	Items            []SaleItemRequest   `json:"items" binding:"required,min=1"`
}

type SaleUpdateRequest struct {
	CustomerID       *uint               `json:"customer_id"`
	SalesPersonID    *uint               `json:"sales_person_id"`
	Date             *time.Time          `json:"date"`
	DueDate          *time.Time          `json:"due_date"`
	ValidUntil       *time.Time          `json:"valid_until"`
	DiscountPercent  *float64            `json:"discount_percent"`
	PPNPercent       *float64            `json:"ppn_percent"`
	PPhPercent       *float64            `json:"pph_percent"`
	PPhType          *string             `json:"pph_type"`
	PaymentTerms     *string             `json:"payment_terms"`
	PaymentMethod    *string             `json:"payment_method"`
	ShippingMethod   *string             `json:"shipping_method"`
	ShippingCost     *float64            `json:"shipping_cost"`
	ShippingTaxable  *bool               `json:"shipping_taxable"`
	BillingAddress   *string             `json:"billing_address"`
	ShippingAddress  *string             `json:"shipping_address"`
	Notes            *string             `json:"notes"`
	InternalNotes    *string             `json:"internal_notes"`
	Reference        *string             `json:"reference"`
	Items            []SaleItemRequest   `json:"items"`
}

type SaleItemRequest struct {
	ProductID        uint    `json:"product_id" binding:"required"`
	Description      string  `json:"description"`
	Quantity         int     `json:"quantity" binding:"required,min=1"`
	UnitPrice        float64 `json:"unit_price" binding:"required,min=0"`
	Discount         float64 `json:"discount"`          // Legacy field - will be mapped to discount_percent
	DiscountPercent  float64 `json:"discount_percent"`  // New field to match frontend
	Tax              float64 `json:"tax"`               // Legacy field
	Taxable          bool    `json:"taxable"`           // New field to match frontend
	RevenueAccountID uint    `json:"revenue_account_id"`
}

// Payment and Return DTOs
type SalePaymentRequest struct {
	SaleID        uint      `json:"sale_id"` // Make optional, will be set by controller
	Amount        float64   `json:"amount" binding:"required,min=0"`
	PaymentDate   time.Time `json:"payment_date" binding:"required"`
	PaymentMethod string    `json:"payment_method" binding:"required"`
	Reference     string    `json:"reference"`
	Notes         string    `json:"notes"`
	CashBankID    *uint     `json:"cash_bank_id"` // Add cashbank integration
	AccountID     *uint     `json:"account_id"`   // Add account integration
}

type SaleReturnRequest struct {
	SaleID      uint                    `json:"sale_id" binding:"required"`
	ReturnDate  time.Time               `json:"return_date" binding:"required"`
	Reason      string                  `json:"reason" binding:"required"`
	Notes       string                  `json:"notes"`
	ReturnItems []SaleReturnItemRequest `json:"return_items" binding:"required,min=1"`
}

type SaleReturnItemRequest struct {
	SaleItemID uint `json:"sale_item_id" binding:"required"`
	Quantity   int  `json:"quantity" binding:"required,min=1"`
	Reason     string `json:"reason"`
}

// Item Create/Update DTOs
type SaleItemCreateRequest struct {
	SaleID           uint    `json:"sale_id" binding:"required"`
	ProductID        uint    `json:"product_id" binding:"required"`
	Quantity         int     `json:"quantity" binding:"required,min=1"`
	UnitPrice        float64 `json:"unit_price" binding:"required,min=0"`
	Discount         float64 `json:"discount"`
	Tax              float64 `json:"tax"`
	RevenueAccountID uint    `json:"revenue_account_id"`
}

type SaleItemUpdateRequest struct {
	Quantity         *int     `json:"quantity"`
	UnitPrice        *float64 `json:"unit_price"`
	Discount         *float64 `json:"discount"`
	Tax              *float64 `json:"tax"`
	RevenueAccountID *uint    `json:"revenue_account_id"`
}

// GORM hooks untuk update computed fields
func (s *Sale) AfterFind(tx *gorm.DB) (err error) {
	// Update SubTotal alias to match Subtotal
	s.SubTotal = s.Subtotal
	return
}

func (s *Sale) BeforeCreate(tx *gorm.DB) (err error) {
	// Update SubTotal alias to match Subtotal
	s.SubTotal = s.Subtotal
	return
}

func (s *Sale) BeforeUpdate(tx *gorm.DB) (err error) {
	// Update SubTotal alias to match Subtotal
	s.SubTotal = s.Subtotal
	return
}

func (s *Sale) AfterCreate(tx *gorm.DB) (err error) {
	// Update SubTotal alias after create
	s.SubTotal = s.Subtotal
	return
}

// GORM hooks untuk SaleItem computed fields
func (si *SaleItem) AfterFind(tx *gorm.DB) (err error) {
	// Update legacy fields to match new fields
	si.TotalPrice = si.LineTotal
	si.Discount = si.DiscountAmount
	si.Tax = si.TotalTax
	return
}
