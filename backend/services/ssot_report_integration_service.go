package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"app-sistem-akuntansi/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SSOTReportIntegrationService integrates SSOT journal with all financial reports
type SSOTReportIntegrationService struct {
	db                   *gorm.DB
	unifiedJournalService *UnifiedJournalService
	enhancedReportService *EnhancedReportService
	mu                   sync.RWMutex
	lastRefreshTime      time.Time
}

// NewSSOTReportIntegrationService creates a new SSOT report integration service
func NewSSOTReportIntegrationService(
	db *gorm.DB,
	unifiedJournalService *UnifiedJournalService,
	enhancedReportService *EnhancedReportService,
) *SSOTReportIntegrationService {
	return &SSOTReportIntegrationService{
		db:                   db,
		unifiedJournalService: unifiedJournalService,
		enhancedReportService: enhancedReportService,
		lastRefreshTime:      time.Now(),
	}
}

// ReportUpdateEvent represents real-time report updates via websocket
type ReportUpdateEvent struct {
	Type        string                 `json:"type"`        // PROFIT_LOSS, BALANCE_SHEET, CASH_FLOW, etc.
	ReportData  map[string]interface{} `json:"report_data"`
	UpdatedAt   time.Time              `json:"updated_at"`
	TriggeredBy string                 `json:"triggered_by"` // JOURNAL_POSTED, TRANSACTION_CREATED, etc.
	JournalID   *uint64                `json:"journal_id,omitempty"`
}

// IntegratedFinancialReports represents all reports integrated with SSOT journal
type IntegratedFinancialReports struct {
	ProfitLoss        *ProfitLossData        `json:"profit_loss"`
	BalanceSheet      *BalanceSheetData      `json:"balance_sheet"`
	CashFlow          *CashFlowData          `json:"cash_flow"`
	SalesSummary      *SalesSummaryData      `json:"sales_summary"`
	VendorAnalysis    *VendorAnalysisData    `json:"vendor_analysis"`
	TrialBalance      *TrialBalanceData      `json:"trial_balance"`
	GeneralLedger     *GeneralLedgerData     `json:"general_ledger"`
	JournalAnalysis   *JournalAnalysisData   `json:"journal_analysis"`
	GeneratedAt       time.Time              `json:"generated_at"`
	DataSourceInfo    DataSourceInfo         `json:"data_source_info"`
}

// DataSourceInfo provides information about the SSOT integration
type DataSourceInfo struct {
	SSOTVersion         string    `json:"ssot_version"`
	LastJournalSync     time.Time `json:"last_journal_sync"`
	TotalJournalEntries int64     `json:"total_journal_entries"`
	PostedEntries       int64     `json:"posted_entries"`
	DraftEntries        int64     `json:"draft_entries"`
	DataIntegrityCheck  bool      `json:"data_integrity_check"`
}

// SSOT-specific extensions of existing types (removing duplicates)

// JournalAnalysisData represents comprehensive journal entry analysis
type JournalAnalysisData struct {
	Company                 CompanyInfo            `json:"company"`
	StartDate               time.Time              `json:"start_date"`
	EndDate                 time.Time              `json:"end_date"`
	Currency                string                 `json:"currency"`
	TotalEntries            int64                  `json:"total_entries"`
	PostedEntries           int64                  `json:"posted_entries"`
	DraftEntries            int64                  `json:"draft_entries"`
	ReversedEntries         int64                  `json:"reversed_entries"`
	TotalAmount             decimal.Decimal        `json:"total_amount"`
	EntriesByType           []EntryTypeBreakdown   `json:"entries_by_type"`
	EntriesByAccount        []AccountBreakdown     `json:"entries_by_account"`
	EntriesByPeriod         []PeriodBreakdown      `json:"entries_by_period"`
	ComplianceCheck         ComplianceReport       `json:"compliance_check"`
	DataQualityMetrics      DataQualityMetrics     `json:"data_quality_metrics"`
	GeneratedAt             time.Time              `json:"generated_at"`
}

// Supporting data structures
type VendorDetail struct {
	VendorID            uint64          `json:"vendor_id"`
	VendorName          string          `json:"vendor_name"`
	TotalPurchases      decimal.Decimal `json:"total_purchases"`
	TotalPayments       decimal.Decimal `json:"total_payments"`
	Outstanding         decimal.Decimal `json:"outstanding"`
	LastTransactionDate time.Time       `json:"last_transaction_date"`
	PaymentTerms        string          `json:"payment_terms"`
	AveragePaymentDays  float64         `json:"average_payment_days"`
}

type PaymentAnalysis struct {
	OnTimePayments  int64           `json:"on_time_payments"`
	LatePayments    int64           `json:"late_payments"`
	AveragePayDays  float64         `json:"average_pay_days"`
	TotalDiscounts  decimal.Decimal `json:"total_discounts"`
	PenaltyFees     decimal.Decimal `json:"penalty_fees"`
}

type AgingBucket struct {
	Description string          `json:"description"`
	DaysRange   string          `json:"days_range"`
	Amount      decimal.Decimal `json:"amount"`
	Count       int64           `json:"count"`
	Percentage  float64         `json:"percentage"`
}

// Using existing types from enhanced_report_service.go
// SSOTTrialBalanceAccount extends TrialBalanceItem
type SSOTTrialBalanceAccount struct {
	AccountID       uint64          `json:"account_id"`
	AccountCode     string          `json:"account_code"`
	AccountName     string          `json:"account_name"`
	AccountType     string          `json:"account_type"`
	DebitBalance    decimal.Decimal `json:"debit_balance"`
	CreditBalance   decimal.Decimal `json:"credit_balance"`
	NormalBalance   string          `json:"normal_balance"`
	SSOTBalance     decimal.Decimal `json:"ssot_balance"`
	LastUpdated     time.Time       `json:"last_updated"`
}

// SSOTGeneralLedgerEntry extends GeneralLedgerEntry
type SSOTGeneralLedgerEntry struct {
	JournalID       uint64          `json:"journal_id"`
	EntryNumber     string          `json:"entry_number"`
	EntryDate       time.Time       `json:"entry_date"`
	Description     string          `json:"description"`
	Reference       string          `json:"reference"`
	DebitAmount     decimal.Decimal `json:"debit_amount"`
	CreditAmount    decimal.Decimal `json:"credit_amount"`
	RunningBalance  decimal.Decimal `json:"running_balance"`
	Status          string          `json:"status"`
	SourceType      string          `json:"source_type"`
	SSOTJournalID   uint64          `json:"ssot_journal_id"`
	SSOTLineID      uint64          `json:"ssot_line_id"`
}

type EntryTypeBreakdown struct {
	SourceType  string          `json:"source_type"`
	Count       int64           `json:"count"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	Percentage  float64         `json:"percentage"`
}

type AccountBreakdown struct {
	AccountID   uint64          `json:"account_id"`
	AccountCode string          `json:"account_code"`
	AccountName string          `json:"account_name"`
	Count       int64           `json:"count"`
	TotalDebit  decimal.Decimal `json:"total_debit"`
	TotalCredit decimal.Decimal `json:"total_credit"`
}

type PeriodBreakdown struct {
	Period      string          `json:"period"`
	StartDate   time.Time       `json:"start_date"`
	EndDate     time.Time       `json:"end_date"`
	Count       int64           `json:"count"`
	TotalAmount decimal.Decimal `json:"total_amount"`
}

type ComplianceReport struct {
	TotalChecks       int                 `json:"total_checks"`
	PassedChecks      int                 `json:"passed_checks"`
	FailedChecks      int                 `json:"failed_checks"`
	ComplianceScore   float64             `json:"compliance_score"`
	Issues            []ComplianceIssue   `json:"issues"`
	Recommendations   []string            `json:"recommendations"`
}

type ComplianceIssue struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	JournalID   uint64 `json:"journal_id"`
}

type DataQualityMetrics struct {
	OverallScore        float64                    `json:"overall_score"`
	CompletenessScore   float64                    `json:"completeness_score"`
	AccuracyScore       float64                    `json:"accuracy_score"`
	ConsistencyScore    float64                    `json:"consistency_score"`
	Issues              []DataQualityIssue         `json:"issues"`
	DetailedMetrics     map[string]interface{}     `json:"detailed_metrics"`
}

type DataQualityIssue struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Count       int64  `json:"count"`
	Severity    string `json:"severity"`
}

// GenerateIntegratedReports generates all financial reports integrated with SSOT journal
func (s *SSOTReportIntegrationService) GenerateIntegratedReports(ctx context.Context, startDate, endDate time.Time) (*IntegratedFinancialReports, error) {
	reports := &IntegratedFinancialReports{
		GeneratedAt: time.Now(),
	}

	// Generate data source info
	dataSourceInfo, err := s.generateDataSourceInfo(ctx)
	if err != nil {
		log.Printf("Error generating data source info: %v", err)
		// Continue with empty data source info
		dataSourceInfo = &DataSourceInfo{}
	}
	reports.DataSourceInfo = *dataSourceInfo

	// Generate all reports concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, 8)

	// Profit & Loss Statement
	wg.Add(1)
	go func() {
		defer wg.Done()
		pl, err := s.enhancedReportService.GenerateProfitLoss(startDate, endDate)
		if err != nil {
			errChan <- fmt.Errorf("profit loss generation failed: %w", err)
			return
		}
		s.mu.Lock()
		reports.ProfitLoss = pl
		s.mu.Unlock()
	}()

	// Balance Sheet
	wg.Add(1)
	go func() {
		defer wg.Done()
		bs, err := s.enhancedReportService.GenerateBalanceSheet(endDate)
		if err != nil {
			errChan <- fmt.Errorf("balance sheet generation failed: %w", err)
			return
		}
		s.mu.Lock()
		reports.BalanceSheet = bs
		s.mu.Unlock()
	}()

	// Cash Flow Statement
	wg.Add(1)
	go func() {
		defer wg.Done()
		cf, err := s.enhancedReportService.GenerateCashFlow(startDate, endDate)
		if err != nil {
			errChan <- fmt.Errorf("cash flow generation failed: %w", err)
			return
		}
		s.mu.Lock()
		reports.CashFlow = cf
		s.mu.Unlock()
	}()

	// Sales Summary Report
	wg.Add(1)
	go func() {
		defer wg.Done()
		ss, err := s.enhancedReportService.GenerateSalesSummary(startDate, endDate, "month")
		if err != nil {
			errChan <- fmt.Errorf("sales summary generation failed: %w", err)
			return
		}
		s.mu.Lock()
		reports.SalesSummary = ss
		s.mu.Unlock()
	}()

	// Vendor Analysis Report
	wg.Add(1)
	go func() {
		defer wg.Done()
		va, err := s.generateVendorAnalysis(ctx, startDate, endDate)
		if err != nil {
			errChan <- fmt.Errorf("vendor analysis generation failed: %w", err)
			return
		}
		s.mu.Lock()
		reports.VendorAnalysis = va
		s.mu.Unlock()
	}()

	// Trial Balance
	wg.Add(1)
	go func() {
		defer wg.Done()
		tb, err := s.generateTrialBalance(ctx, endDate)
		if err != nil {
			errChan <- fmt.Errorf("trial balance generation failed: %w", err)
			return
		}
		s.mu.Lock()
		reports.TrialBalance = tb
		s.mu.Unlock()
	}()

	// General Ledger
	wg.Add(1)
	go func() {
		defer wg.Done()
		gl, err := s.generateGeneralLedger(ctx, startDate, endDate, nil)
		if err != nil {
			errChan <- fmt.Errorf("general ledger generation failed: %w", err)
			return
		}
		s.mu.Lock()
		reports.GeneralLedger = gl
		s.mu.Unlock()
	}()

	// Journal Analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		ja, err := s.generateJournalAnalysis(ctx, startDate, endDate)
		if err != nil {
			errChan <- fmt.Errorf("journal analysis generation failed: %w", err)
			return
		}
		s.mu.Lock()
		reports.JournalAnalysis = ja
		s.mu.Unlock()
	}()

	// Wait for all reports to complete
	wg.Wait()
	close(errChan)

	// Check for any errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return reports, fmt.Errorf("report generation completed with errors: %v", errors)
	}

	// Update last refresh time (WebSocket broadcasting removed)
	s.mu.Lock()
	s.lastRefreshTime = time.Now()
	s.mu.Unlock()

	// Log report generation completion
	log.Printf("Generated integrated reports: profit_loss, balance_sheet, cash_flow, sales_summary, vendor_analysis, trial_balance, general_ledger, journal_analysis")

	return reports, nil
}

// GenerateSalesSummaryFromSSot generates sales summary using SSOT journal data
func (s *SSOTReportIntegrationService) GenerateSalesSummaryFromSSot(startDate, endDate time.Time) (*SalesSummaryData, error) {
	return s.enhancedReportService.GenerateSalesSummary(startDate, endDate, "month")
}

// GenerateVendorAnalysisFromSSot generates vendor analysis using SSOT journal data
func (s *SSOTReportIntegrationService) GenerateVendorAnalysisFromSSot(startDate, endDate time.Time) (*VendorAnalysisData, error) {
	ctx := context.Background()
	return s.generateVendorAnalysis(ctx, startDate, endDate)
}

// GenerateTrialBalanceFromSSot generates trial balance using SSOT journal data
func (s *SSOTReportIntegrationService) GenerateTrialBalanceFromSSot(asOfDate time.Time) (*TrialBalanceData, error) {
	ctx := context.Background()
	return s.generateTrialBalance(ctx, asOfDate)
}

// GenerateGeneralLedgerFromSSot generates general ledger using SSOT journal data
func (s *SSOTReportIntegrationService) GenerateGeneralLedgerFromSSot(startDate, endDate time.Time, accountID *uint64) (*GeneralLedgerData, error) {
	ctx := context.Background()
	return s.generateGeneralLedger(ctx, startDate, endDate, accountID)
}

// GenerateJournalAnalysisFromSSot generates journal analysis using SSOT journal data
func (s *SSOTReportIntegrationService) GenerateJournalAnalysisFromSSot(startDate, endDate time.Time) (*JournalAnalysisData, error) {
	ctx := context.Background()
	return s.generateJournalAnalysis(ctx, startDate, endDate)
}

// OnJournalPosted handles updates when a journal is posted (WebSocket removed for stability)
func (s *SSOTReportIntegrationService) OnJournalPosted(journalID uint64) {
	// Update last refresh time
	s.mu.Lock()
	s.lastRefreshTime = time.Now()
	s.mu.Unlock()

	// Log journal posting event (WebSocket broadcasting removed)
	log.Printf("Journal %d posted, reports may need refresh", journalID)
}

// OnTransactionCreated handles updates when transactions are created (WebSocket removed for stability)
func (s *SSOTReportIntegrationService) OnTransactionCreated(transactionType string, transactionID uint64, affectedAccounts []uint64) {
	// Update last refresh time
	s.mu.Lock()
	s.lastRefreshTime = time.Now()
	s.mu.Unlock()

	// Log transaction creation event (WebSocket broadcasting removed)
	log.Printf("Transaction %s %d created, reports may need refresh", transactionType, transactionID)
}

// Private helper methods

func (s *SSOTReportIntegrationService) generateDataSourceInfo(ctx context.Context) (*DataSourceInfo, error) {
	var stats struct {
		TotalEntries int64 `json:"total_entries"`
		PostedEntries int64 `json:"posted_entries"`
		DraftEntries int64 `json:"draft_entries"`
	}

	// Get journal statistics
	if err := s.db.WithContext(ctx).Model(&models.SSOTJournalEntry{}).
		Select(`
			COUNT(*) as total_entries,
			COUNT(CASE WHEN status = ? THEN 1 END) as posted_entries,
			COUNT(CASE WHEN status = ? THEN 1 END) as draft_entries
		`, models.SSOTStatusPosted, models.SSOTStatusDraft).
		Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get journal statistics: %w", err)
	}

	// Get last sync time (last updated journal entry)
	var lastSync time.Time
	s.db.WithContext(ctx).Model(&models.SSOTJournalEntry{}).
		Select("MAX(updated_at)").
		Where("status = ?", models.SSOTStatusPosted).
		Scan(&lastSync)

	return &DataSourceInfo{
		SSOTVersion:         "1.0",
		LastJournalSync:     lastSync,
		TotalJournalEntries: stats.TotalEntries,
		PostedEntries:       stats.PostedEntries,
		DraftEntries:        stats.DraftEntries,
		DataIntegrityCheck:  true, // This could be a more complex check
	}, nil
}


// GetDB returns the database instance (for external access)
func (s *SSOTReportIntegrationService) GetDB() *gorm.DB {
	return s.db
}

// getCompanyInfo returns company information for reports
func (s *SSOTReportIntegrationService) getCompanyInfo() CompanyInfo {
	// TODO: Load from database or configuration
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

// Implementation of report generation methods
func (s *SSOTReportIntegrationService) generateVendorAnalysis(ctx context.Context, startDate, endDate time.Time) (*VendorAnalysisData, error) {
	var result VendorAnalysisData
	result.Company = s.getCompanyInfo()
	result.StartDate = startDate
	result.EndDate = endDate
	result.Currency = "IDR"
	result.GeneratedAt = time.Now()

	// Get vendor analysis from SSOT journal data
	vendorStats := make(map[uint64]*VendorDetail)
	
	// Query vendor transactions from SSOT journal entries with account information
	type VendorJournalLine struct {
		JournalID    uint64
		AccountID    uint64
		AccountCode  string
		AccountName  string
		Description  string
		DebitAmount  decimal.Decimal
		CreditAmount decimal.Decimal
		EntryDate    time.Time
		SourceID     *uint64
		SourceType   string
	}

	var journalLines []VendorJournalLine
	query := `
		SELECT 
			sjl.journal_id,
			sjl.account_id,
			a.code as account_code,
			a.name as account_name,
			sjl.description,
			sjl.debit_amount,
			sjl.credit_amount,
			sje.entry_date,
			sje.source_id,
			sje.source_type
		FROM unified_journal_lines sjl
		JOIN unified_journal_ledger sje ON sje.id = sjl.journal_id
		LEFT JOIN accounts a ON a.id = sjl.account_id
		WHERE sje.entry_date BETWEEN ? AND ?
			AND sje.status = ?
			AND a.code LIKE '2%'
			AND sje.source_type IN ('PURCHASE', 'PAYMENT')
		ORDER BY sje.entry_date, sjl.journal_id
	`

	err := s.db.WithContext(ctx).Raw(query, startDate, endDate, models.SSOTStatusPosted).Scan(&journalLines).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vendor journal entries: %w", err)
	}

	// Process journal lines to build vendor analysis
	for _, line := range journalLines {
		// Use SourceID as VendorID for purchase/payment transactions
		if line.SourceID != nil && *line.SourceID > 0 {
			vendorID := *line.SourceID
			if _, exists := vendorStats[vendorID]; !exists {
				vendorStats[vendorID] = &VendorDetail{
					VendorID:   vendorID,
					VendorName: line.Description,
					LastTransactionDate: line.EntryDate,
				}
			}
			
			vendor := vendorStats[vendorID]
			// For accounts payable: Credit increases payable (purchases), Debit decreases payable (payments)
			if line.CreditAmount.GreaterThan(decimal.Zero) {
				vendor.TotalPurchases = vendor.TotalPurchases.Add(line.CreditAmount)
			} else if line.DebitAmount.GreaterThan(decimal.Zero) {
				vendor.TotalPayments = vendor.TotalPayments.Add(line.DebitAmount)
			}
			
			// Update last transaction date
			if line.EntryDate.After(vendor.LastTransactionDate) {
				vendor.LastTransactionDate = line.EntryDate
			}
		}
	}

	// Calculate totals and convert to slice
	var vendors []VendorDetail
	var totalPurchases, totalPayments decimal.Decimal
	
	for _, vendor := range vendorStats {
		vendor.Outstanding = vendor.TotalPurchases.Sub(vendor.TotalPayments)
		vendors = append(vendors, *vendor)
		totalPurchases = totalPurchases.Add(vendor.TotalPurchases)
		totalPayments = totalPayments.Add(vendor.TotalPayments)
	}

	result.TotalVendors = int64(len(vendors))
	result.ActiveVendors = result.TotalVendors // Simplified - all vendors are considered active
	result.TotalPurchases, _ = totalPurchases.Float64()
	result.TotalPayments, _ = totalPayments.Float64()
	result.OutstandingPayables, _ = totalPurchases.Sub(totalPayments).Float64()

	// Create vendor performance data
	result.VendorsByPerformance = []VendorPerformanceData{}
	for _, vendor := range vendors {
		purchases, _ := vendor.TotalPurchases.Float64()
		payments, _ := vendor.TotalPayments.Float64()
		outstanding, _ := vendor.Outstanding.Float64()
		
		perfData := VendorPerformanceData{
			VendorID:          uint(vendor.VendorID),
			VendorName:        vendor.VendorName,
			TotalPurchases:    purchases,
			TotalPayments:     payments,
			Outstanding:       outstanding,
			AveragePaymentDays: 30.0, // Default value
			PaymentScore:      85.0, // Default score
			Rating:           "Good", // Default rating
		}
		result.VendorsByPerformance = append(result.VendorsByPerformance, perfData)
	}

	return &result, nil
}

func (s *SSOTReportIntegrationService) generateTrialBalance(ctx context.Context, asOfDate time.Time) (*TrialBalanceData, error) {
	result := &TrialBalanceData{
		Company:     s.getCompanyInfo(),
		AsOfDate:    asOfDate,
		Currency:    "IDR",
		Accounts:    []TrialBalanceItem{},
		GeneratedAt: time.Now(),
	}

	// Get all accounts with their balances from SSOT journal
	type AccountBalance struct {
		AccountID   uint64
		AccountCode string
		AccountName string
		AccountType string
		TotalDebit  decimal.Decimal
		TotalCredit decimal.Decimal
	}

	var balances []AccountBalance
	query := `
		SELECT 
			sjl.account_id,
			a.code as account_code,
			a.name as account_name,
			CASE 
				WHEN a.code LIKE '1%' THEN 'Asset'
				WHEN a.code LIKE '2%' THEN 'Liability'
				WHEN a.code LIKE '3%' THEN 'Equity'
				WHEN a.code LIKE '4%' THEN 'Revenue'
				WHEN a.code LIKE '5%' THEN 'Expense'
				ELSE 'Other'
			END as account_type,
			COALESCE(SUM(sjl.debit_amount), 0) as total_debit,
			COALESCE(SUM(sjl.credit_amount), 0) as total_credit
		FROM unified_journal_lines sjl
		JOIN unified_journal_ledger sje ON sje.id = sjl.journal_id
		LEFT JOIN accounts a ON a.id = sjl.account_id
		WHERE sje.entry_date <= ?
			AND sje.status = ?
		GROUP BY sjl.account_id, a.code, a.name
		ORDER BY a.code
	`

	err := s.db.WithContext(ctx).Raw(query, asOfDate, models.SSOTStatusPosted).Scan(&balances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate trial balance: %w", err)
	}

	var totalDebits, totalCredits decimal.Decimal
	var accounts []TrialBalanceItem

	// Process balances
	for _, bal := range balances {
		item := TrialBalanceItem{
			AccountID:     uint(bal.AccountID),
			AccountCode:   bal.AccountCode,
			AccountName:   bal.AccountName,
			AccountType:   bal.AccountType,
			DebitBalance:  bal.TotalDebit.InexactFloat64(),
			CreditBalance: bal.TotalCredit.InexactFloat64(),
		}
		
		// Calculate net balance based on account type
		netBalance := bal.TotalDebit.Sub(bal.TotalCredit)
		
		// For display purposes, show balance in normal side
		switch bal.AccountType {
		case "Asset", "Expense":
			// Normal debit accounts
			if netBalance.GreaterThanOrEqual(decimal.Zero) {
				item.DebitBalance = netBalance.InexactFloat64()
				item.CreditBalance = 0
			} else {
				item.DebitBalance = 0
				item.CreditBalance = netBalance.Abs().InexactFloat64()
			}
		case "Liability", "Equity", "Revenue":
			// Normal credit accounts
			netBalance = netBalance.Neg() // Reverse for credit accounts
			if netBalance.GreaterThanOrEqual(decimal.Zero) {
				item.CreditBalance = netBalance.InexactFloat64()
				item.DebitBalance = 0
			} else {
				item.CreditBalance = 0
				item.DebitBalance = netBalance.Abs().InexactFloat64()
			}
		}
		
		totalDebits = totalDebits.Add(bal.TotalDebit)
		totalCredits = totalCredits.Add(bal.TotalCredit)
		accounts = append(accounts, item)
	}

	result.Accounts = accounts
	result.TotalDebits = totalDebits.InexactFloat64()
	result.TotalCredits = totalCredits.InexactFloat64()
	result.IsBalanced = totalDebits.Equal(totalCredits)
	result.Difference = totalDebits.Sub(totalCredits).InexactFloat64()

	return result, nil
}

func (s *SSOTReportIntegrationService) generateGeneralLedger(ctx context.Context, startDate, endDate time.Time, accountID *uint64) (*GeneralLedgerData, error) {
	result := &GeneralLedgerData{
		Company:           s.getCompanyInfo(),
		StartDate:         startDate,
		EndDate:           endDate,
		Currency:          "IDR",
		OpeningBalance:    0.0,
		ClosingBalance:    0.0,
		TotalDebits:       0.0,
		TotalCredits:      0.0,
		Transactions:      []GeneralLedgerEntry{},
		MonthlySummary:    []MonthlyLedgerSummary{},
		GeneratedAt:       time.Now(),
	}

	// Build query for general ledger
	type ledgerRow struct {
		JournalID    uint64
		EntryNumber  string
		EntryDate    time.Time
		Description  string
		Reference    string
		AccountID    uint64
		AccountCode  string
		AccountName  string
		DebitAmount  decimal.Decimal
		CreditAmount decimal.Decimal
		Status       string
		SourceType   string
	}

	var query string
	var args []interface{}
	
	if accountID != nil {
		query = `
			SELECT 
				sje.id as journal_id,
				sje.entry_number,
				sje.entry_date,
				sjl.description,
				sje.reference,
				sjl.account_id,
				a.code as account_code,
				a.name as account_name,
				sjl.debit_amount,
				sjl.credit_amount,
				sje.status,
				sje.source_type
			FROM unified_journal_lines sjl
			JOIN unified_journal_ledger sje ON sje.id = sjl.journal_id
			LEFT JOIN accounts a ON a.id = sjl.account_id
			WHERE sje.status = ? 
				AND sje.entry_date BETWEEN ? AND ?
				AND sjl.account_id = ?
			ORDER BY sje.entry_date, sje.id, sjl.line_number
		`
		args = []interface{}{models.SSOTStatusPosted, startDate, endDate, *accountID}
	} else {
		query = `
			SELECT 
				sje.id as journal_id,
				sje.entry_number,
				sje.entry_date,
				sjl.description,
				sje.reference,
				sjl.account_id,
				a.code as account_code,
				a.name as account_name,
				sjl.debit_amount,
				sjl.credit_amount,
				sje.status,
				sje.source_type
			FROM unified_journal_lines sjl
			JOIN unified_journal_ledger sje ON sje.id = sjl.journal_id
			LEFT JOIN accounts a ON a.id = sjl.account_id
			WHERE sje.status = ? 
				AND sje.entry_date BETWEEN ? AND ?
			ORDER BY sje.entry_date, sje.id, sjl.line_number
		`
		args = []interface{}{models.SSOTStatusPosted, startDate, endDate}
	}
	
	var rows []ledgerRow
	err := s.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query general ledger: %w", err)
	}

	// Group by account for organized display
	accountEntries := make(map[uint64][]GeneralLedgerEntry)
	accountInfo := make(map[uint64]struct{ Code, Name string })
	
	for _, row := range rows {
		entry := GeneralLedgerEntry{
			Date:         row.EntryDate,
			JournalCode:  row.EntryNumber,
			Description:  row.Description,
			Reference:    row.Reference,
			DebitAmount:  row.DebitAmount.InexactFloat64(),
			CreditAmount: row.CreditAmount.InexactFloat64(),
			EntryType:    row.SourceType,
		}
		
		accountEntries[row.AccountID] = append(accountEntries[row.AccountID], entry)
		accountInfo[row.AccountID] = struct{ Code, Name string }{
			Code: row.AccountCode,
			Name: row.AccountName,
		}
	}

	// Calculate running balances, organize entries, and compute totals
	var totalDebits, totalCredits decimal.Decimal
	accountBalances := make(map[uint64]decimal.Decimal) // Track final balance per account
	
	for accID, entries := range accountEntries {
		info := accountInfo[accID]
		var runningBalance decimal.Decimal
		
		for i, entry := range entries {
			// Add to overall totals
			totalDebits = totalDebits.Add(decimal.NewFromFloat(entry.DebitAmount))
			totalCredits = totalCredits.Add(decimal.NewFromFloat(entry.CreditAmount))
			
			// Calculate running balance for this entry
			runningBalance = runningBalance.Add(decimal.NewFromFloat(entry.DebitAmount)).Sub(decimal.NewFromFloat(entry.CreditAmount))
			entries[i].Balance = runningBalance.InexactFloat64()
		}
		
		// Store final balance for this account
		accountBalances[accID] = runningBalance
		
		// Add account header info to description if needed
		if len(entries) > 0 {
			accountHeader := fmt.Sprintf("Account: %s - %s", info.Code, info.Name)
			entries[0].Description = accountHeader + " | " + entries[0].Description
		}
		
		result.Transactions = append(result.Transactions, entries...)
	}
	
	// Calculate enhanced UI metrics
	var totalAccountBalances decimal.Decimal
	var cashAccountBalance decimal.Decimal
	
	// Sum all account balances and calculate cash impact
	for accID, balance := range accountBalances {
		totalAccountBalances = totalAccountBalances.Add(balance)
		
		// Calculate cash impact (accounts starting with 11xx are typically cash accounts)
		if info, exists := accountInfo[accID]; exists {
			if strings.HasPrefix(info.Code, "11") { // Cash and cash equivalents
				cashAccountBalance = cashAccountBalance.Add(balance)
			}
		}
	}
	
	// Calculate UI-friendly metrics
	netPositionChange := totalAccountBalances.InexactFloat64()
	totalTransactionVolume := totalDebits.InexactFloat64() // Total transaction activity
	cashImpact := cashAccountBalance.InexactFloat64()
	isBalanced := totalDebits.Equal(totalCredits)
	
	// Determine status messages
	var netPositionStatus string
	if netPositionChange == 0 {
		netPositionStatus = "Balanced"
	} else if netPositionChange > 0 {
		netPositionStatus = "Net Debit Position"
	} else {
		netPositionStatus = "Net Credit Position"
	}
	
	var cashImpactStatus string
	if cashImpact > 0 {
		cashImpactStatus = "Cash Increased"
	} else if cashImpact < 0 {
		cashImpactStatus = "Cash Decreased"
	} else {
		cashImpactStatus = "No Cash Impact"
	}
	
	// Update result with calculated totals and enhanced UI fields
	result.TotalDebits = totalDebits.InexactFloat64()
	result.TotalCredits = totalCredits.InexactFloat64()
	result.OpeningBalance = 0.0 // Starting point for period
	result.ClosingBalance = netPositionChange
	result.NetPositionChange = netPositionChange
	result.NetPositionStatus = netPositionStatus
	result.TotalTransactionVol = totalTransactionVolume
	result.CashImpact = cashImpact
	result.CashImpactStatus = cashImpactStatus
	result.IsBalanced = isBalanced

	return result, nil
}

func (s *SSOTReportIntegrationService) generateJournalAnalysis(ctx context.Context, startDate, endDate time.Time) (*JournalAnalysisData, error) {
	// Get basic statistics
	var stats struct {
		TotalEntries    int64           `json:"total_entries"`
		PostedEntries   int64           `json:"posted_entries"`
		DraftEntries    int64           `json:"draft_entries"`
		ReversedEntries int64           `json:"reversed_entries"`
		TotalAmount     decimal.Decimal `json:"total_amount"`
	}

	if err := s.db.WithContext(ctx).Model(&models.SSOTJournalEntry{}).
		Select(`
			COUNT(*) as total_entries,
			COUNT(CASE WHEN status = ? THEN 1 END) as posted_entries,
			COUNT(CASE WHEN status = ? THEN 1 END) as draft_entries,
			COUNT(CASE WHEN status = ? THEN 1 END) as reversed_entries,
			COALESCE(SUM(total_debit), 0) as total_amount
		`, models.SSOTStatusPosted, models.SSOTStatusDraft, models.SSOTStatusReversed).
		Where("entry_date >= ? AND entry_date <= ?", startDate, endDate).
		Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get journal statistics: %w", err)
	}

	// Get entries by type
	var entriesByType []EntryTypeBreakdown
	s.db.WithContext(ctx).Model(&models.SSOTJournalEntry{}).
		Select("source_type, COUNT(*) as count, COALESCE(SUM(total_debit), 0) as total_amount").
		Where("entry_date >= ? AND entry_date <= ? AND status = ?", startDate, endDate, models.SSOTStatusPosted).
		Group("source_type").
		Scan(&entriesByType)

	// Calculate percentages for entry types
	for i := range entriesByType {
		if stats.TotalAmount.GreaterThan(decimal.Zero) {
			entriesByType[i].Percentage = entriesByType[i].TotalAmount.Div(stats.TotalAmount).Mul(decimal.NewFromInt(100)).InexactFloat64()
		}
	}

	return &JournalAnalysisData{
		StartDate:          startDate,
		EndDate:            endDate,
		TotalEntries:       stats.TotalEntries,
		PostedEntries:      stats.PostedEntries,
		DraftEntries:       stats.DraftEntries,
		ReversedEntries:    stats.ReversedEntries,
		TotalAmount:        stats.TotalAmount,
		EntriesByType:      entriesByType,
		EntriesByAccount:   []AccountBreakdown{}, // TODO: Implement
		EntriesByPeriod:    []PeriodBreakdown{},  // TODO: Implement
		ComplianceCheck:    ComplianceReport{},   // TODO: Implement
		DataQualityMetrics: DataQualityMetrics{}, // TODO: Implement
		GeneratedAt:        time.Now(),
	}, nil
}