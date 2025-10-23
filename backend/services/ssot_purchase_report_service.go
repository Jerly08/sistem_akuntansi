package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// SSOTPurchaseReportService generates purchase reports from SSOT journal data
type SSOTPurchaseReportService struct {
	db *gorm.DB
}

// NewSSOTPurchaseReportService creates a new SSOT purchase report service
func NewSSOTPurchaseReportService(db *gorm.DB) *SSOTPurchaseReportService {
	return &SSOTPurchaseReportService{
		db: db,
	}
}

// PurchaseReportData represents comprehensive purchase analysis
type PurchaseReportData struct {
	Company              CompanyInfo              `json:"company"`
	StartDate            time.Time               `json:"start_date"`
	EndDate              time.Time               `json:"end_date"`
	Currency             string                  `json:"currency"`
	TotalPurchases       int64                   `json:"total_purchases"`
	CompletedPurchases   int64                   `json:"completed_purchases"`
	TotalAmount          float64                 `json:"total_amount"`
	TotalPaid            float64                 `json:"total_paid"`
	OutstandingPayables  float64                 `json:"outstanding_payables"`
	PurchasesByVendor    []VendorPurchaseSummary `json:"purchases_by_vendor"`
	PurchasesByMonth     []MonthlyPurchaseSummary `json:"purchases_by_month"`
	PurchasesByCategory  []CategoryPurchaseSummary `json:"purchases_by_category"`
	PaymentAnalysis      PurchasePaymentAnalysis  `json:"payment_analysis"`
	TaxAnalysis          PurchaseTaxAnalysis      `json:"tax_analysis"`
	GeneratedAt          time.Time               `json:"generated_at"`
}

// VendorPurchaseSummary represents purchase summary by vendor
type VendorPurchaseSummary struct {
	VendorID        uint64               `json:"vendor_id"`
	VendorName      string               `json:"vendor_name"`
	TotalPurchases  int64                `json:"total_purchases"`
	TotalAmount     float64              `json:"total_amount"`
	TotalPaid       float64              `json:"total_paid"`
	Outstanding     float64              `json:"outstanding"`
	LastPurchaseDate time.Time           `json:"last_purchase_date"`
	PaymentMethod   string               `json:"payment_method"`
	Status          string               `json:"status"`
	Items           []PurchaseItemDetail `json:"items,omitempty"`
}

// PurchaseItemDetail represents individual item purchased
type PurchaseItemDetail struct {
	ProductID     uint64    `json:"product_id"`
	ProductCode   string    `json:"product_code"`
	ProductName   string    `json:"product_name"`
	Quantity      float64   `json:"quantity"`
	UnitPrice     float64   `json:"unit_price"`
	TotalPrice    float64   `json:"total_price"`
	Unit          string    `json:"unit"`
	PurchaseDate  time.Time `json:"purchase_date"`
	InvoiceNumber string    `json:"invoice_number,omitempty"`
}

// MonthlyPurchaseSummary represents purchase summary by month
type MonthlyPurchaseSummary struct {
	Year            int     `json:"year"`
	Month           int     `json:"month"`
	MonthName       string  `json:"month_name"`
	TotalPurchases  int64   `json:"total_purchases"`
	TotalAmount     float64 `json:"total_amount"`
	TotalPaid       float64 `json:"total_paid"`
	AverageAmount   float64 `json:"average_amount"`
}

// CategoryPurchaseSummary represents purchase summary by category
type CategoryPurchaseSummary struct {
	CategoryName    string  `json:"category_name"`
	AccountCode     string  `json:"account_code"`
	AccountName     string  `json:"account_name"`
	TotalPurchases  int64   `json:"total_purchases"`
	TotalAmount     float64 `json:"total_amount"`
	Percentage      float64 `json:"percentage"`
}

// PurchasePaymentAnalysis represents payment pattern analysis
type PurchasePaymentAnalysis struct {
	CashPurchases     int64   `json:"cash_purchases"`
	CreditPurchases   int64   `json:"credit_purchases"`
	CashAmount        float64 `json:"cash_amount"`
	CreditAmount      float64 `json:"credit_amount"`
	CashPercentage    float64 `json:"cash_percentage"`
	CreditPercentage  float64 `json:"credit_percentage"`
	AverageOrderValue float64 `json:"average_order_value"`
}

// PurchaseTaxAnalysis represents tax analysis for purchases
type PurchaseTaxAnalysis struct {
	TotalTaxableAmount     float64 `json:"total_taxable_amount"`
	TotalTaxAmount         float64 `json:"total_tax_amount"`
	AverageTaxRate         float64 `json:"average_tax_rate"`
	TaxReclaimableAmount   float64 `json:"tax_reclaimable_amount"`
	TaxByMonth             []MonthlyTaxSummary `json:"tax_by_month"`
}

// MonthlyTaxSummary represents tax summary by month
type MonthlyTaxSummary struct {
	Year      int     `json:"year"`
	Month     int     `json:"month"`
	MonthName string  `json:"month_name"`
	TaxAmount float64 `json:"tax_amount"`
}

// GeneratePurchaseReport generates comprehensive purchase report from SSOT data
func (s *SSOTPurchaseReportService) GeneratePurchaseReport(ctx context.Context, startDate, endDate time.Time) (*PurchaseReportData, error) {
result := &PurchaseReportData{
		Company:     s.getCompanyInfo(),
		StartDate:   startDate,
		EndDate:     endDate,
		Currency:    s.getCurrencyFromSettings(),
		GeneratedAt: time.Now(),
	}

	// Get purchase transactions from SSOT journal
	purchaseSummary, err := s.getPurchaseSummary(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase summary: %w", err)
	}

	// Populate basic statistics
	result.TotalPurchases = purchaseSummary.TotalCount
	result.CompletedPurchases = purchaseSummary.CompletedCount
	result.TotalAmount = purchaseSummary.TotalAmount
	result.TotalPaid = purchaseSummary.TotalPaid
	result.OutstandingPayables = purchaseSummary.TotalAmount - purchaseSummary.TotalPaid

	// Get detailed analyses
	result.PurchasesByVendor, err = s.getPurchasesByVendor(ctx, startDate, endDate)
	if err != nil {
		log.Printf("Error getting purchases by vendor: %v", err)
		result.PurchasesByVendor = []VendorPurchaseSummary{}
	}

	result.PurchasesByMonth, err = s.getPurchasesByMonth(ctx, startDate, endDate)
	if err != nil {
		log.Printf("Error getting purchases by month: %v", err)
		result.PurchasesByMonth = []MonthlyPurchaseSummary{}
	}

	result.PurchasesByCategory, err = s.getPurchasesByCategory(ctx, startDate, endDate)
	if err != nil {
		log.Printf("Error getting purchases by category: %v", err)
		result.PurchasesByCategory = []CategoryPurchaseSummary{}
	}

	result.PaymentAnalysis, err = s.getPaymentAnalysis(ctx, startDate, endDate)
	if err != nil {
		log.Printf("Error getting payment analysis: %v", err)
		result.PaymentAnalysis = PurchasePaymentAnalysis{}
	}

	result.TaxAnalysis, err = s.getTaxAnalysis(ctx, startDate, endDate)
	if err != nil {
		log.Printf("Error getting tax analysis: %v", err)
		result.TaxAnalysis = PurchaseTaxAnalysis{}
	}

	return result, nil
}

// Helper struct for purchase summary
type purchaseBaseSummary struct {
	TotalCount     int64
	CompletedCount int64
	TotalAmount    float64
	TotalPaid      float64
}

// getPurchaseSummary gets basic purchase statistics from SSOT journal
func (s *SSOTPurchaseReportService) getPurchaseSummary(ctx context.Context, startDate, endDate time.Time) (*purchaseBaseSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_count,
			COUNT(CASE WHEN status = 'POSTED' THEN 1 END) as completed_count,
			COALESCE(SUM(total_debit), 0) as total_amount,
			-- For cash purchases, total_paid = total_amount (fully paid)
			-- Check both main description and if there are cash account debits
			COALESCE(SUM(CASE 
				WHEN description ILIKE '%cash%' OR description ILIKE '%kas%' OR
				     EXISTS(SELECT 1 FROM unified_journal_lines ujl 
				            JOIN accounts a ON ujl.account_id = a.id 
				            WHERE ujl.journal_id = unified_journal_ledger.id 
				              AND a.code IN ('1101', '1102', '1103', '1104', '1105') -- Cash/Bank accounts
				              AND ujl.credit_amount > 0) -- Cash account credited (cash out)
				THEN total_debit  -- Cash: use debit amount (full purchase amount)
				ELSE 0           -- Credit: not paid yet
			END), 0) as total_paid
		FROM unified_journal_ledger 
		WHERE source_type = 'PURCHASE'
		  AND entry_date BETWEEN ? AND ?
		  AND deleted_at IS NULL
	`

	var summary purchaseBaseSummary
	err := s.db.WithContext(ctx).Raw(query, startDate, endDate).Scan(&summary).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query purchase summary: %w", err)
	}

	return &summary, nil
}

// getPurchasesByVendor gets purchase summary grouped by vendor with correct outstanding logic
func (s *SSOTPurchaseReportService) getPurchasesByVendor(ctx context.Context, startDate, endDate time.Time) ([]VendorPurchaseSummary, error) {
	// Updated query with better vendor name extraction and cash/credit logic
	query := `
		SELECT 
			COALESCE(sje.source_id, 0) as vendor_id,
			-- Extract vendor name from description using multiple patterns
			CASE 
				WHEN sje.source_id IS NOT NULL AND sje.source_id > 0 
				THEN COALESCE((
					-- Try multiple patterns to extract vendor name
					SELECT CASE
						WHEN description ~ 'Purchase from (.+) - ' 
							THEN TRIM(SUBSTRING(description FROM 'Purchase from (.+) - '))
						WHEN description ~ 'Purchase Order [^-]+ - (.+)'
							THEN TRIM(SUBSTRING(description FROM 'Purchase Order [^-]+ - (.+)'))
						WHEN description ~ '- (.+)$'
							THEN TRIM(SUBSTRING(description FROM '- (.+)$'))
						ELSE NULL
					END
					FROM unified_journal_ledger
					WHERE source_id = sje.source_id 
					  AND source_type = 'PURCHASE'
					LIMIT 1
				), 'Vendor ID: ' || sje.source_id::text)
				ELSE 'Unknown Vendor'
			END as vendor_name,
			COUNT(*) as total_purchases,
			COALESCE(SUM(sje.total_debit), 0) as total_amount,
			-- For cash purchases, total_paid = total_amount (fully paid)
			-- Check if any purchase in this vendor group has cash payment
			CASE 
				WHEN bool_or(sje.description ILIKE '%cash%' OR sje.description ILIKE '%kas%' OR
					     EXISTS(SELECT 1 FROM unified_journal_lines ujl 
					            JOIN accounts a ON ujl.account_id = a.id 
					            WHERE ujl.journal_id = sje.id
					              AND a.code IN ('1101', '1102', '1103', '1104', '1105') 
					              AND ujl.credit_amount > 0))
				THEN COALESCE(SUM(sje.total_debit), 0)  -- Cash = fully paid
				ELSE 0  -- Credit = not paid yet (simplified)
			END as total_paid,
			MAX(sje.entry_date) as last_purchase_date,
			CASE 
				WHEN bool_or(sje.description ILIKE '%cash%' OR sje.description ILIKE '%kas%' OR
					     EXISTS(SELECT 1 FROM unified_journal_lines ujl 
					            JOIN accounts a ON ujl.account_id = a.id 
					            WHERE ujl.journal_id = sje.id
					              AND a.code IN ('1101', '1102', '1103', '1104', '1105') 
					              AND ujl.credit_amount > 0))
				THEN 'CASH'
				ELSE 'CREDIT' 
			END as payment_method,
			CASE WHEN bool_and(sje.status = 'POSTED') THEN 'COMPLETED' ELSE 'PENDING' END as status,
			-- Get sample descriptions for debugging
			string_agg(DISTINCT sje.description, ', ') as descriptions
		FROM unified_journal_ledger sje
		WHERE sje.source_type = 'PURCHASE'
		  AND sje.entry_date BETWEEN ? AND ?
		  AND sje.deleted_at IS NULL
		GROUP BY sje.source_id
		ORDER BY total_amount DESC
	`

	var vendors []struct {
		VendorID         *uint64    `json:"vendor_id"`
		VendorName       string     `json:"vendor_name"`
		TotalPurchases   int64      `json:"total_purchases"`
		TotalAmount      float64    `json:"total_amount"`
		TotalPaid        float64    `json:"total_paid"`
		LastPurchaseDate time.Time  `json:"last_purchase_date"`
		PaymentMethod    string     `json:"payment_method"`
		Status           string     `json:"status"`
		Descriptions     string     `json:"descriptions"`
	}

	err := s.db.WithContext(ctx).Raw(query, startDate, endDate).Scan(&vendors).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query purchases by vendor: %w", err)
	}

	// Add logging for debugging
	log.Printf("Found %d vendor groups from SSOT journal", len(vendors))
	for i, v := range vendors {
		log.Printf("Vendor %d: ID=%v, Name=%s, Purchases=%d, Amount=%.2f, Descriptions=%s", 
			i+1, v.VendorID, v.VendorName, v.TotalPurchases, v.TotalAmount, v.Descriptions)
	}

	// Convert to result format
	var result []VendorPurchaseSummary
	validVendorCount := 0
	
	for _, v := range vendors {
		vendorID := uint64(0)
		if v.VendorID != nil {
			vendorID = *v.VendorID
		}

		// Only include vendors that have meaningful data
		if v.TotalAmount > 0 && v.VendorName != "Unknown Vendor" {
			validVendorCount++
		}

		summary := VendorPurchaseSummary{
			VendorID:         vendorID,
			VendorName:       v.VendorName,
			TotalPurchases:   v.TotalPurchases,
			TotalAmount:      v.TotalAmount,
			TotalPaid:        v.TotalPaid,
			Outstanding:      v.TotalAmount - v.TotalPaid,
			LastPurchaseDate: v.LastPurchaseDate,
			PaymentMethod:    v.PaymentMethod,
			Status:           v.Status,
		}

		result = append(result, summary)
	}

	log.Printf("Valid vendors found: %d out of %d total groups", validVendorCount, len(vendors))
	
	// Fetch items for each vendor using SSOT journal source_id
	for i := range result {
		items, err := s.getPurchaseItemsFromSSOT(ctx, startDate, endDate, result[i].VendorName)
		if err != nil {
			log.Printf("Warning: Failed to get items for vendor %s: %v", result[i].VendorName, err)
			continue
		}
		result[i].Items = items
		if len(items) > 0 {
			log.Printf("✅ Loaded %d items for vendor: %s", len(items), result[i].VendorName)
		}
	}
	
	return result, nil
}

// getPurchaseItemsFromSSOT gets detailed items purchased from SSOT journal
func (s *SSOTPurchaseReportService) getPurchaseItemsFromSSOT(ctx context.Context, startDate, endDate time.Time, vendorName string) ([]PurchaseItemDetail, error) {
	// Query items using SSOT journal source_id to link to purchases
	query := `
		SELECT 
			COALESCE(pi.product_id, 0) as product_id,
			COALESCE(p.code, 'N/A') as product_code,
			COALESCE(p.name, 'Unknown Product') as product_name,
			pi.quantity,
			pi.unit_price,
			pi.total_price,
			COALESCE(p.unit, 'pcs') as unit,
			pur.date as purchase_date,
			COALESCE(pur.code, '') as invoice_number
		FROM unified_journal_ledger ujl
		INNER JOIN purchases pur ON pur.id = ujl.source_id
		INNER JOIN purchase_items pi ON pi.purchase_id = pur.id
		INNER JOIN contacts v ON v.id = pur.vendor_id
		LEFT JOIN products p ON p.id = pi.product_id
		WHERE ujl.source_type = 'PURCHASE'
		  AND ujl.entry_date BETWEEN ? AND ?
		  AND ujl.deleted_at IS NULL
		  AND pur.deleted_at IS NULL
		  AND pi.deleted_at IS NULL
		  AND v.name = ?
		ORDER BY pur.date DESC, pi.id
	`
	
	var items []PurchaseItemDetail
	err := s.db.WithContext(ctx).Raw(query, startDate, endDate, vendorName).Scan(&items).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query purchase items from SSOT: %w", err)
	}
	
	log.Printf("🔍 Query items for vendor '%s': found %d items", vendorName, len(items))
	return items, nil
}

// getPurchasesByMonth gets purchase summary grouped by month
func (s *SSOTPurchaseReportService) getPurchasesByMonth(ctx context.Context, startDate, endDate time.Time) ([]MonthlyPurchaseSummary, error) {
	query := `
		SELECT 
			EXTRACT(YEAR FROM entry_date) as year,
			EXTRACT(MONTH FROM entry_date) as month,
			COUNT(*) as total_purchases,
			COALESCE(SUM(total_debit), 0) as total_amount,
			COALESCE(SUM(CASE 
				WHEN description ILIKE '%cash%' OR description ILIKE '%kas%' 
				THEN total_credit 
				ELSE 0 
			END), 0) as total_paid
		FROM unified_journal_ledger
		WHERE source_type = 'PURCHASE'
		  AND entry_date BETWEEN ? AND ?
		  AND deleted_at IS NULL
		GROUP BY EXTRACT(YEAR FROM entry_date), EXTRACT(MONTH FROM entry_date)
		ORDER BY year, month
	`

	var months []struct {
		Year           int     `json:"year"`
		Month          int     `json:"month"`
		TotalPurchases int64   `json:"total_purchases"`
		TotalAmount    float64 `json:"total_amount"`
		TotalPaid      float64 `json:"total_paid"`
	}

	err := s.db.WithContext(ctx).Raw(query, startDate, endDate).Scan(&months).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query purchases by month: %w", err)
	}

	// Convert to result format with month names
	monthNames := []string{
		"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}

	var result []MonthlyPurchaseSummary
	for _, m := range months {
		avgAmount := float64(0)
		if m.TotalPurchases > 0 {
			avgAmount = m.TotalAmount / float64(m.TotalPurchases)
		}

		summary := MonthlyPurchaseSummary{
			Year:           m.Year,
			Month:          m.Month,
			MonthName:      monthNames[m.Month],
			TotalPurchases: m.TotalPurchases,
			TotalAmount:    m.TotalAmount,
			TotalPaid:      m.TotalPaid,
			AverageAmount:  avgAmount,
		}

		result = append(result, summary)
	}

	return result, nil
}

// getPurchasesByCategory gets purchase summary grouped by account category
func (s *SSOTPurchaseReportService) getPurchasesByCategory(ctx context.Context, startDate, endDate time.Time) ([]CategoryPurchaseSummary, error) {
	query := `
		SELECT 
			a.code as account_code,
			a.name as account_name,
			CASE 
				WHEN a.code LIKE '13%' THEN 'Inventory'
				WHEN a.code LIKE '15%' THEN 'Fixed Assets'
				WHEN a.code LIKE '6%' THEN 'Expenses'
				ELSE 'Other'
			END as category_name,
			COUNT(*) as total_purchases,
			COALESCE(SUM(sjl.debit_amount), 0) as total_amount
		FROM unified_journal_lines sjl
		JOIN unified_journal_ledger sje ON sje.id = sjl.journal_id
		LEFT JOIN accounts a ON a.id = sjl.account_id
		WHERE sje.source_type = 'PURCHASE'
		  AND sje.entry_date BETWEEN ? AND ?
		  AND sje.deleted_at IS NULL
		  AND sjl.debit_amount > 0
		  AND a.code NOT LIKE '21%'  -- Exclude tax accounts
		GROUP BY a.code, a.name
		ORDER BY total_amount DESC
	`

	var categories []struct {
		AccountCode    string  `json:"account_code"`
		AccountName    string  `json:"account_name"`
		CategoryName   string  `json:"category_name"`
		TotalPurchases int64   `json:"total_purchases"`
		TotalAmount    float64 `json:"total_amount"`
	}

	err := s.db.WithContext(ctx).Raw(query, startDate, endDate).Scan(&categories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query purchases by category: %w", err)
	}

	// Calculate total for percentage calculation
	var totalAmount float64
	for _, cat := range categories {
		totalAmount += cat.TotalAmount
	}

	// Convert to result format with percentages
	var result []CategoryPurchaseSummary
	for _, cat := range categories {
		percentage := float64(0)
		if totalAmount > 0 {
			percentage = (cat.TotalAmount / totalAmount) * 100
		}

		summary := CategoryPurchaseSummary{
			CategoryName:   cat.CategoryName,
			AccountCode:    cat.AccountCode,
			AccountName:    cat.AccountName,
			TotalPurchases: cat.TotalPurchases,
			TotalAmount:    cat.TotalAmount,
			Percentage:     percentage,
		}

		result = append(result, summary)
	}

	return result, nil
}

// getPaymentAnalysis analyzes payment patterns in purchases
func (s *SSOTPurchaseReportService) getPaymentAnalysis(ctx context.Context, startDate, endDate time.Time) (PurchasePaymentAnalysis, error) {
	query := `
		SELECT 
			-- Count cash purchases (either by description or by cash account usage)
			COUNT(CASE 
				WHEN unified_journal_ledger.description ILIKE '%cash%' OR unified_journal_ledger.description ILIKE '%kas%' OR
				     EXISTS(SELECT 1 FROM unified_journal_lines ujl 
				            JOIN accounts a ON ujl.account_id = a.id 
				            WHERE ujl.journal_id = unified_journal_ledger.id
				              AND a.code IN ('1101', '1102', '1103', '1104', '1105') -- Cash/Bank accounts
				              AND ujl.credit_amount > 0) -- Cash account credited (cash out)
				THEN 1 END) as cash_purchases,
			-- Count credit purchases (not cash)
			COUNT(CASE 
				WHEN NOT (unified_journal_ledger.description ILIKE '%cash%' OR unified_journal_ledger.description ILIKE '%kas%' OR
				          EXISTS(SELECT 1 FROM unified_journal_lines ujl 
				                 JOIN accounts a ON ujl.account_id = a.id 
				                 WHERE ujl.journal_id = unified_journal_ledger.id 
				                   AND a.code IN ('1101', '1102', '1103', '1104', '1105')
				                   AND ujl.credit_amount > 0))
				THEN 1 END) as credit_purchases,
			-- Cash amount: sum of cash purchases
			COALESCE(SUM(CASE 
				WHEN unified_journal_ledger.description ILIKE '%cash%' OR unified_journal_ledger.description ILIKE '%kas%' OR
				     EXISTS(SELECT 1 FROM unified_journal_lines ujl 
				            JOIN accounts a ON ujl.account_id = a.id 
				            WHERE ujl.journal_id = unified_journal_ledger.id 
				              AND a.code IN ('1101', '1102', '1103', '1104', '1105')
				              AND ujl.credit_amount > 0)
				THEN unified_journal_ledger.total_debit 
				ELSE 0 
			END), 0) as cash_amount,
			-- Credit amount: sum of credit purchases
			COALESCE(SUM(CASE 
				WHEN NOT (unified_journal_ledger.description ILIKE '%cash%' OR unified_journal_ledger.description ILIKE '%kas%' OR
				          EXISTS(SELECT 1 FROM unified_journal_lines ujl 
				                 JOIN accounts a ON ujl.account_id = a.id 
				                 WHERE ujl.journal_id = unified_journal_ledger.id 
				                   AND a.code IN ('1101', '1102', '1103', '1104', '1105')
				                   AND ujl.credit_amount > 0))
				THEN unified_journal_ledger.total_debit 
				ELSE 0 
			END), 0) as credit_amount,
			COALESCE(AVG(unified_journal_ledger.total_debit), 0) as average_order_value
		FROM unified_journal_ledger
		WHERE source_type = 'PURCHASE'
		  AND entry_date BETWEEN ? AND ?
		  AND deleted_at IS NULL
	`

	var analysis struct {
		CashPurchases    int64   `json:"cash_purchases"`
		CreditPurchases  int64   `json:"credit_purchases"`
		CashAmount       float64 `json:"cash_amount"`
		CreditAmount     float64 `json:"credit_amount"`
		AverageOrderValue float64 `json:"average_order_value"`
	}

	err := s.db.WithContext(ctx).Raw(query, startDate, endDate).Scan(&analysis).Error
	if err != nil {
		return PurchasePaymentAnalysis{}, fmt.Errorf("failed to analyze payment patterns: %w", err)
	}

	totalAmount := analysis.CashAmount + analysis.CreditAmount
	cashPercentage := float64(0)
	creditPercentage := float64(0)

	if totalAmount > 0 {
		cashPercentage = (analysis.CashAmount / totalAmount) * 100
		creditPercentage = (analysis.CreditAmount / totalAmount) * 100
	}

	return PurchasePaymentAnalysis{
		CashPurchases:     analysis.CashPurchases,
		CreditPurchases:   analysis.CreditPurchases,
		CashAmount:        analysis.CashAmount,
		CreditAmount:      analysis.CreditAmount,
		CashPercentage:    cashPercentage,
		CreditPercentage:  creditPercentage,
		AverageOrderValue: analysis.AverageOrderValue,
	}, nil
}

// getTaxAnalysis analyzes tax information in purchases
func (s *SSOTPurchaseReportService) getTaxAnalysis(ctx context.Context, startDate, endDate time.Time) (PurchaseTaxAnalysis, error) {
	// Get tax summary
	taxQuery := `
		SELECT 
			COALESCE(SUM(CASE 
				WHEN a.code NOT LIKE '21%' 
				THEN sjl.debit_amount 
				ELSE 0 
			END), 0) as total_taxable_amount,
			COALESCE(SUM(CASE 
				WHEN a.code LIKE '21%' AND a.name ILIKE '%ppn%'
				THEN sjl.debit_amount 
				ELSE 0 
			END), 0) as total_tax_amount
		FROM unified_journal_lines sjl
		JOIN unified_journal_ledger sje ON sje.id = sjl.journal_id
		LEFT JOIN accounts a ON a.id = sjl.account_id
		WHERE sje.source_type = 'PURCHASE'
		  AND sje.entry_date BETWEEN ? AND ?
		  AND sje.deleted_at IS NULL
		  AND sjl.debit_amount > 0
	`

	var taxSummary struct {
		TotalTaxableAmount float64 `json:"total_taxable_amount"`
		TotalTaxAmount     float64 `json:"total_tax_amount"`
	}

	err := s.db.WithContext(ctx).Raw(taxQuery, startDate, endDate).Scan(&taxSummary).Error
	if err != nil {
		return PurchaseTaxAnalysis{}, fmt.Errorf("failed to analyze tax: %w", err)
	}

	averageTaxRate := float64(0)
	if taxSummary.TotalTaxableAmount > 0 {
		averageTaxRate = (taxSummary.TotalTaxAmount / taxSummary.TotalTaxableAmount) * 100
	}

	// Get monthly tax breakdown
	monthlyTaxQuery := `
		SELECT 
			EXTRACT(YEAR FROM sje.entry_date) as year,
			EXTRACT(MONTH FROM sje.entry_date) as month,
			COALESCE(SUM(sjl.debit_amount), 0) as tax_amount
		FROM unified_journal_lines sjl
		JOIN unified_journal_ledger sje ON sje.id = sjl.journal_id
		LEFT JOIN accounts a ON a.id = sjl.account_id
		WHERE sje.source_type = 'PURCHASE'
		  AND sje.entry_date BETWEEN ? AND ?
		  AND sje.deleted_at IS NULL
		  AND sjl.debit_amount > 0
		  AND a.code LIKE '21%' AND a.name ILIKE '%ppn%'
		GROUP BY EXTRACT(YEAR FROM sje.entry_date), EXTRACT(MONTH FROM sje.entry_date)
		ORDER BY year, month
	`

	var monthlyTax []struct {
		Year      int     `json:"year"`
		Month     int     `json:"month"`
		TaxAmount float64 `json:"tax_amount"`
	}

	err = s.db.WithContext(ctx).Raw(monthlyTaxQuery, startDate, endDate).Scan(&monthlyTax).Error
	if err != nil {
		log.Printf("Error getting monthly tax data: %v", err)
		monthlyTax = []struct {
			Year      int     `json:"year"`
			Month     int     `json:"month"`
			TaxAmount float64 `json:"tax_amount"`
		}{}
	}

	// Convert monthly tax to result format
	monthNames := []string{
		"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}

	var taxByMonth []MonthlyTaxSummary
	for _, mt := range monthlyTax {
		summary := MonthlyTaxSummary{
			Year:      mt.Year,
			Month:     mt.Month,
			MonthName: monthNames[mt.Month],
			TaxAmount: mt.TaxAmount,
		}
		taxByMonth = append(taxByMonth, summary)
	}

	return PurchaseTaxAnalysis{
		TotalTaxableAmount:   taxSummary.TotalTaxableAmount,
		TotalTaxAmount:       taxSummary.TotalTaxAmount,
		AverageTaxRate:       averageTaxRate,
		TaxReclaimableAmount: taxSummary.TotalTaxAmount, // Input tax is reclaimable
		TaxByMonth:           taxByMonth,
	}, nil
}

// getCompanyInfo returns company information for reports
func (s *SSOTPurchaseReportService) getCompanyInfo() CompanyInfo {
	// Prefer Settings table (admin-configured company information)
	var settings models.Settings
	if err := s.db.First(&settings).Error; err == nil {
		return CompanyInfo{
			Name:      settings.CompanyName,
			Address:   settings.CompanyAddress,
			City:      "", // City may be embedded in the address field
			State:     "",
			Phone:     settings.CompanyPhone,
			Email:     settings.CompanyEmail,
			Website:   settings.CompanyWebsite,
			TaxNumber: settings.TaxNumber,
		}
	}
	// Fallback defaults
	return CompanyInfo{
		Name:      "PT. Default Company",
		Address:   "Jalan Default No. 1",
		City:      "Jakarta",
		State:     "DKI Jakarta",
		Phone:     "+62-21-12345678",
		Email:     "info@defaultcompany.com",
		TaxNumber: "01.234.567.8-901.000",
	}
}

// getCurrencyFromSettings returns the configured currency or IDR as fallback
func (s *SSOTPurchaseReportService) getCurrencyFromSettings() string {
	var settings models.Settings
	if err := s.db.First(&settings).Error; err == nil && settings.Currency != "" {
		return settings.Currency
	}
	return "IDR"
}
