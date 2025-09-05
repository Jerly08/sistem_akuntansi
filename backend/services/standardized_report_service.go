package services

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type StandardizedReportService struct {
	db              *gorm.DB
	accountRepo     repositories.AccountRepository
	salesRepo       *repositories.SalesRepository
	purchaseRepo    *repositories.PurchaseRepository
	productRepo     *repositories.ProductRepository
	contactRepo     repositories.ContactRepository
	paymentRepo     *repositories.PaymentRepository
	cashBankRepo    *repositories.CashBankRepository
	companyProfile  *models.CompanyProfile
}

type ReportStyle struct {
	CompanyName        string
	CompanyAddress     string
	CompanyPhone       string
	CompanyEmail       string
	LogoPath          string
	ReportTitle       string
	ReportSubtitle    string
	Currency          string
	DateFormat        string
	NumberFormat      string
	HeaderColor       [3]int
	AlternateRowColor [3]int
	FontFamily        string
	ShowPageNumbers   bool
	ShowWatermark     bool
	WatermarkText     string
}

type FinancialStatement struct {
	Header         StatementHeader                `json:"header"`
	Sections       []StatementSection             `json:"sections"`
	Totals         map[string]float64             `json:"totals"`
	Ratios         map[string]float64             `json:"ratios"`
	Notes          []string                       `json:"notes"`
	Comparatives   map[string]interface{}         `json:"comparatives,omitempty"`
	Metadata       StatementMetadata              `json:"metadata"`
}

type StatementHeader struct {
	CompanyName    string    `json:"company_name"`
	StatementType  string    `json:"statement_type"`
	PeriodStart    time.Time `json:"period_start,omitempty"`
	PeriodEnd      time.Time `json:"period_end"`
	AsOfDate       time.Time `json:"as_of_date,omitempty"`
	GeneratedAt    time.Time `json:"generated_at"`
	Currency       string    `json:"currency"`
	PreparedBy     string    `json:"prepared_by"`
}

type StatementSection struct {
	Name           string                 `json:"name"`
	Type           string                 `json:"type"` // ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
	Items          []StatementItem        `json:"items"`
	Subtotal       float64               `json:"subtotal"`
	Order          int                   `json:"order"`
	IsCollapsible  bool                  `json:"is_collapsible"`
	Subsections    []StatementSection    `json:"subsections,omitempty"`
}

type StatementItem struct {
	AccountID      uint      `json:"account_id"`
	AccountCode    string    `json:"account_code"`
	AccountName    string    `json:"account_name"`
	Balance        float64   `json:"balance"`
	PreviousBalance float64   `json:"previous_balance,omitempty"`
	Variance       float64   `json:"variance,omitempty"`
	VariancePercent float64   `json:"variance_percent,omitempty"`
	Level          int       `json:"level"`
	IsHeader       bool      `json:"is_header"`
	IsBold         bool      `json:"is_bold"`
	IsTotal        bool      `json:"is_total"`
	ParentID       *uint     `json:"parent_id,omitempty"`
	Notes          string    `json:"notes,omitempty"`
}

type StatementMetadata struct {
	GenerationTime  time.Duration          `json:"generation_time"`
	RecordCount     int                    `json:"record_count"`
	Filters         map[string]interface{} `json:"filters"`
	Version         string                 `json:"version"`
	Signature       string                 `json:"signature,omitempty"`
}

// Enhanced Report Response with standard compliance
type StandardizedReportResponse struct {
	ID              string                         `json:"id"`
	Title           string                         `json:"title"`
	Type            string                         `json:"type"`
	Period          string                         `json:"period"`
	GeneratedAt     time.Time                      `json:"generated_at"`
	Statement       FinancialStatement             `json:"statement"`
	FileData        []byte                         `json:"-"`
	Summary         map[string]float64             `json:"summary"`
	Parameters      map[string]interface{}         `json:"parameters"`
	Compliance      ComplianceInfo                 `json:"compliance"`
	Audit           AuditTrail                     `json:"audit"`
}

type ComplianceInfo struct {
	Standard        string    `json:"standard"` // GAAP, IFRS, SAK
	LastReviewed    time.Time `json:"last_reviewed"`
	ReviewedBy      string    `json:"reviewed_by"`
	Certifications  []string  `json:"certifications"`
	Warnings        []string  `json:"warnings,omitempty"`
}

type AuditTrail struct {
	GeneratedBy     string    `json:"generated_by"`
	GeneratedAt     time.Time `json:"generated_at"`
	DataSources     []string  `json:"data_sources"`
	LastDataUpdate  time.Time `json:"last_data_update"`
	Hash            string    `json:"hash"`
}

func NewStandardizedReportService(
	db *gorm.DB,
	accountRepo repositories.AccountRepository,
	salesRepo *repositories.SalesRepository,
	purchaseRepo *repositories.PurchaseRepository,
	productRepo *repositories.ProductRepository,
	contactRepo repositories.ContactRepository,
	paymentRepo *repositories.PaymentRepository,
	cashBankRepo *repositories.CashBankRepository,
) *StandardizedReportService {
	
	service := &StandardizedReportService{
		db:           db,
		accountRepo:  accountRepo,
		salesRepo:    salesRepo,
		purchaseRepo: purchaseRepo,
		productRepo:  productRepo,
		contactRepo:  contactRepo,
		paymentRepo:  paymentRepo,
		cashBankRepo: cashBankRepo,
	}
	
	// Load company profile
	service.loadCompanyProfile()
	
	return service
}

func (srs *StandardizedReportService) loadCompanyProfile() {
	var profile models.CompanyProfile
	if err := srs.db.First(&profile).Error; err != nil {
		// Create default profile if none exists
		profile = models.CompanyProfile{
			Name:     "Your Company Name",
			Currency: "IDR",
			Country:  "Indonesia",
		}
		srs.db.Create(&profile)
	}
	srs.companyProfile = &profile
}

// Generate standardized Balance Sheet following Indonesian accounting standards (SAK)
func (srs *StandardizedReportService) GenerateStandardBalanceSheet(asOfDate time.Time, format string, comparative bool) (*StandardizedReportResponse, error) {
	startTime := time.Now()
	
	// Get company profile and accounting context
	ctx := context.Background()
	accounts, err := srs.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}

	statement := FinancialStatement{
		Header: StatementHeader{
			CompanyName:   srs.companyProfile.Name,
			StatementType: "Balance Sheet / Neraca",
			AsOfDate:      asOfDate,
			GeneratedAt:   time.Now(),
			Currency:      srs.companyProfile.Currency,
			PreparedBy:    "System Generated",
		},
		Sections: []StatementSection{},
		Totals:   make(map[string]float64),
		Ratios:   make(map[string]float64),
		Notes:    []string{},
	}

	// Build hierarchical account structure
	accountHierarchy := srs.buildAccountHierarchy(accounts, asOfDate, comparative)
	
	// Create standardized sections following SAK structure
	assetSection := srs.createAssetSection(accountHierarchy, comparative)
	liabilitySection := srs.createLiabilitySection(accountHierarchy, comparative)
	equitySection := srs.createEquitySection(accountHierarchy, comparative)

	statement.Sections = append(statement.Sections, assetSection, liabilitySection, equitySection)

	// Calculate totals and verify balance equation
	statement.Totals["total_assets"] = assetSection.Subtotal
	statement.Totals["total_liabilities"] = liabilitySection.Subtotal
	statement.Totals["total_equity"] = equitySection.Subtotal
	statement.Totals["total_liabilities_equity"] = liabilitySection.Subtotal + equitySection.Subtotal

	// Calculate financial ratios
	statement.Ratios = srs.calculateBalanceSheetRatios(statement.Totals, accountHierarchy)

	// Add compliance notes
	statement.Notes = srs.generateBalanceSheetNotes(statement.Totals)

	// Verify accounting equation
	isBalanced := math.Abs(statement.Totals["total_assets"]-statement.Totals["total_liabilities_equity"]) < 0.01
	if !isBalanced {
		statement.Notes = append(statement.Notes, 
			fmt.Sprintf("WARNING: Balance sheet does not balance. Assets: %.2f, Liabilities + Equity: %.2f", 
				statement.Totals["total_assets"], statement.Totals["total_liabilities_equity"]))
	}

	// Create response
	response := &StandardizedReportResponse{
		ID:          fmt.Sprintf("balance-sheet-%d", time.Now().Unix()),
		Title:       "Balance Sheet / Neraca",
		Type:        models.ReportTypeBalanceSheet,
		Period:      asOfDate.Format("2006-01-02"),
		GeneratedAt: time.Now(),
		Statement:   statement,
		Summary:     statement.Totals,
		Parameters: map[string]interface{}{
			"as_of_date":  asOfDate.Format("2006-01-02"),
			"format":      format,
			"comparative": comparative,
		},
		Compliance: ComplianceInfo{
			Standard:     "SAK (Indonesian GAAP)",
			LastReviewed: time.Now(),
			ReviewedBy:   "System",
		},
		Audit: AuditTrail{
			GeneratedBy:    "System",
			GeneratedAt:    time.Now(),
			DataSources:    []string{"accounts", "journal_entries"},
			LastDataUpdate: time.Now(),
		},
	}

	// Generate file data if requested
	if format == "pdf" {
		pdfData, err := srs.generateStandardizedBalanceSheetPDF(&statement, asOfDate, comparative)
		if err != nil {
			return nil, fmt.Errorf("failed to generate PDF: %v", err)
		}
		response.FileData = pdfData
	} else if format == "excel" {
		excelData, err := srs.generateStandardizedBalanceSheetExcel(&statement, asOfDate, comparative)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Excel: %v", err)
		}
		response.FileData = excelData
	}

	// Set metadata
	response.Statement.Metadata = StatementMetadata{
		GenerationTime: time.Since(startTime),
		RecordCount:    len(accounts),
		Version:        "2.0",
		Filters: map[string]interface{}{
			"as_of_date":  asOfDate,
			"comparative": comparative,
		},
	}

	return response, nil
}

// Generate standardized Profit & Loss Statement following SAK
func (srs *StandardizedReportService) GenerateStandardProfitLoss(startDate, endDate time.Time, format string, comparative bool) (*StandardizedReportResponse, error) {
	startTime := time.Now()
	
	ctx := context.Background()
	accounts, err := srs.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}

	statement := FinancialStatement{
		Header: StatementHeader{
			CompanyName:   srs.companyProfile.Name,
			StatementType: "Profit & Loss Statement / Laporan Laba Rugi",
			PeriodStart:   startDate,
			PeriodEnd:     endDate,
			GeneratedAt:   time.Now(),
			Currency:      srs.companyProfile.Currency,
			PreparedBy:    "System Generated",
		},
		Sections: []StatementSection{},
		Totals:   make(map[string]float64),
		Ratios:   make(map[string]float64),
		Notes:    []string{},
	}

	// Build P&L specific account hierarchy
	plHierarchy := srs.buildPLAccountHierarchy(accounts, startDate, endDate, comparative)

	// Create standardized P&L sections
	revenueSection := srs.createRevenueSection(plHierarchy, comparative)
	cogsSection := srs.createCOGSSection(plHierarchy, comparative)
	operatingExpenseSection := srs.createOperatingExpenseSection(plHierarchy, comparative)
	nonOperatingSection := srs.createNonOperatingSection(plHierarchy, comparative)

	statement.Sections = append(statement.Sections, 
		revenueSection, cogsSection, operatingExpenseSection, nonOperatingSection)

	// Calculate comprehensive P&L totals
	grossRevenue := revenueSection.Subtotal
	costOfGoodsSold := cogsSection.Subtotal
	operatingExpenses := operatingExpenseSection.Subtotal
	nonOperatingNet := nonOperatingSection.Subtotal

	statement.Totals["gross_revenue"] = grossRevenue
	statement.Totals["cost_of_goods_sold"] = costOfGoodsSold
	statement.Totals["gross_profit"] = grossRevenue - costOfGoodsSold
	statement.Totals["operating_expenses"] = operatingExpenses
	statement.Totals["operating_income"] = statement.Totals["gross_profit"] - operatingExpenses
	statement.Totals["non_operating_income"] = nonOperatingNet
	statement.Totals["income_before_tax"] = statement.Totals["operating_income"] + nonOperatingNet
	
	// Calculate tax (simplified - would need actual tax calculation)
	taxRate := 0.25 // 25% corporate tax rate for Indonesia
	statement.Totals["income_tax"] = statement.Totals["income_before_tax"] * taxRate
	statement.Totals["net_income"] = statement.Totals["income_before_tax"] - statement.Totals["income_tax"]

	// Calculate profitability ratios
	statement.Ratios = srs.calculateProfitabilityRatios(statement.Totals)

	// Add P&L notes
	statement.Notes = srs.generateProfitLossNotes(statement.Totals)

	// Create response
	response := &StandardizedReportResponse{
		ID:          fmt.Sprintf("profit-loss-%d", time.Now().Unix()),
		Title:       "Profit & Loss Statement / Laporan Laba Rugi",
		Type:        models.ReportTypeIncomeStatement,
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		GeneratedAt: time.Now(),
		Statement:   statement,
		Summary:     statement.Totals,
		Parameters: map[string]interface{}{
			"start_date":  startDate.Format("2006-01-02"),
			"end_date":    endDate.Format("2006-01-02"),
			"format":      format,
			"comparative": comparative,
		},
		Compliance: ComplianceInfo{
			Standard:     "SAK (Indonesian GAAP)",
			LastReviewed: time.Now(),
			ReviewedBy:   "System",
		},
		Audit: AuditTrail{
			GeneratedBy:    "System",
			GeneratedAt:    time.Now(),
			DataSources:    []string{"accounts", "journal_entries", "sales", "purchases"},
			LastDataUpdate: time.Now(),
		},
	}

	// Generate file data if requested
	if format == "pdf" {
		pdfData, err := srs.generateStandardizedProfitLossPDF(&statement, startDate, endDate, comparative)
		if err != nil {
			return nil, fmt.Errorf("failed to generate PDF: %v", err)
		}
		response.FileData = pdfData
	} else if format == "excel" {
		excelData, err := srs.generateStandardizedProfitLossExcel(&statement, startDate, endDate, comparative)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Excel: %v", err)
		}
		response.FileData = excelData
	}

	// Set metadata
	response.Statement.Metadata = StatementMetadata{
		GenerationTime: time.Since(startTime),
		RecordCount:    len(accounts),
		Version:        "2.0",
		Filters: map[string]interface{}{
			"start_date":  startDate,
			"end_date":    endDate,
			"comparative": comparative,
		},
	}

	return response, nil
}

// Helper methods for account hierarchy and calculations

func (srs *StandardizedReportService) buildAccountHierarchy(accounts []models.Account, asOfDate time.Time, comparative bool) map[string][]StatementItem {
	hierarchy := make(map[string][]StatementItem)
	
	for _, account := range accounts {
		balance := srs.calculateAccountBalanceStandardized(account.ID, asOfDate)
		var previousBalance float64
		
		if comparative {
			// Calculate balance for previous period (same date last year)
			previousDate := asOfDate.AddDate(-1, 0, 0)
			previousBalance = srs.calculateAccountBalanceStandardized(account.ID, previousDate)
		}
		
		variance := balance - previousBalance
		variancePercent := 0.0
		if previousBalance != 0 {
			variancePercent = (variance / previousBalance) * 100
		}
		
		item := StatementItem{
			AccountID:       account.ID,
			AccountCode:     account.Code,
			AccountName:     account.Name,
			Balance:         balance,
			PreviousBalance: previousBalance,
			Variance:        variance,
			VariancePercent: variancePercent,
			Level:           account.Level,
			IsHeader:        account.IsHeader,
			IsBold:          account.IsHeader,
			ParentID:        account.ParentID,
		}
		
		hierarchy[account.Type] = append(hierarchy[account.Type], item)
	}
	
	// Sort by account code within each type
	for accountType := range hierarchy {
		sort.Slice(hierarchy[accountType], func(i, j int) bool {
			return hierarchy[accountType][i].AccountCode < hierarchy[accountType][j].AccountCode
		})
	}
	
	return hierarchy
}

func (srs *StandardizedReportService) buildPLAccountHierarchy(accounts []models.Account, startDate, endDate time.Time, comparative bool) map[string][]StatementItem {
	hierarchy := make(map[string][]StatementItem)
	
	for _, account := range accounts {
		// Only include P&L accounts (Revenue and Expense)
		if account.Type != models.AccountTypeRevenue && account.Type != models.AccountTypeExpense {
			continue
		}
		
		balance := srs.calculateAccountBalanceForPeriodStandardized(account.ID, startDate, endDate)
		var previousBalance float64
		
		if comparative {
			// Calculate balance for previous period
			prevStart := startDate.AddDate(-1, 0, 0)
			prevEnd := endDate.AddDate(-1, 0, 0)
			previousBalance = srs.calculateAccountBalanceForPeriodStandardized(account.ID, prevStart, prevEnd)
		}
		
		variance := balance - previousBalance
		variancePercent := 0.0
		if previousBalance != 0 {
			variancePercent = (variance / previousBalance) * 100
		}
		
		item := StatementItem{
			AccountID:       account.ID,
			AccountCode:     account.Code,
			AccountName:     account.Name,
			Balance:         balance,
			PreviousBalance: previousBalance,
			Variance:        variance,
			VariancePercent: variancePercent,
			Level:           account.Level,
			IsHeader:        account.IsHeader,
			IsBold:          account.IsHeader,
			ParentID:        account.ParentID,
		}
		
		hierarchy[account.Type] = append(hierarchy[account.Type], item)
	}
	
	// Sort by account code within each type
	for accountType := range hierarchy {
		sort.Slice(hierarchy[accountType], func(i, j int) bool {
			return hierarchy[accountType][i].AccountCode < hierarchy[accountType][j].AccountCode
		})
	}
	
	return hierarchy
}

func (srs *StandardizedReportService) calculateAccountBalanceStandardized(accountID uint, asOfDate time.Time) float64 {
	// Get account details
	var account models.Account
	if err := srs.db.First(&account, accountID).Error; err != nil {
		return 0
	}

	// Calculate balance from journal entries up to asOfDate with proper handling
	var totalDebits, totalCredits float64
	srs.db.Table("journal_entries").
		Joins("JOIN journals ON journal_entries.journal_id = journals.id").
		Where("journal_entries.account_id = ? AND journals.date <= ? AND journals.status = 'POSTED'", accountID, asOfDate).
		Select("COALESCE(SUM(journal_entries.debit_amount), 0) as total_debits, COALESCE(SUM(journal_entries.credit_amount), 0) as total_credits").
		Row().Scan(&totalDebits, &totalCredits)

	// Apply proper accounting equation rules
	switch account.Type {
	case models.AccountTypeAsset, models.AccountTypeExpense:
		// Assets and Expenses: Debit increases, Credit decreases
		return account.Balance + totalDebits - totalCredits
	case models.AccountTypeLiability, models.AccountTypeEquity, models.AccountTypeRevenue:
		// Liabilities, Equity, Revenue: Credit increases, Debit decreases
		return account.Balance + totalCredits - totalDebits
	default:
		return account.Balance
	}
}

func (srs *StandardizedReportService) calculateAccountBalanceForPeriodStandardized(accountID uint, startDate, endDate time.Time) float64 {
	// Get account details
	var account models.Account
	if err := srs.db.First(&account, accountID).Error; err != nil {
		return 0
	}

	var totalDebits, totalCredits float64
	srs.db.Table("journal_entries").
		Joins("JOIN journals ON journal_entries.journal_id = journals.id").
		Where("journal_entries.account_id = ? AND journals.date BETWEEN ? AND ? AND journals.status = 'POSTED'", accountID, startDate, endDate).
		Select("COALESCE(SUM(journal_entries.debit_amount), 0) as total_debits, COALESCE(SUM(journal_entries.credit_amount), 0) as total_credits").
		Row().Scan(&totalDebits, &totalCredits)

	// For P&L accounts, we want the period activity
	switch account.Type {
	case models.AccountTypeRevenue:
		// Revenue: Credit is positive
		return totalCredits - totalDebits
	case models.AccountTypeExpense:
		// Expenses: Debit is positive
		return totalDebits - totalCredits
	default:
		// For balance sheet accounts, return cumulative balance
		return srs.calculateAccountBalanceStandardized(accountID, endDate)
	}
}

// createAssetSection creates the asset section of the balance sheet
func (srs *StandardizedReportService) createAssetSection(accountHierarchy map[string][]StatementItem, comparative bool) StatementSection {
	assetSection := StatementSection{
		Name:        "ASSETS / ASET",
		Type:        models.AccountTypeAsset,
		Items:       []StatementItem{},
		Subtotal:    0,
		Order:       1,
		Subsections: []StatementSection{},
	}

	// Current Assets Subsection
	currentAssets := StatementSection{
		Name:        "Current Assets / Aset Lancar",
		Type:        "CURRENT_ASSET",
		Items:       []StatementItem{},
		Subtotal:    0,
		Order:       1,
	}

	// Non-Current Assets Subsection
	nonCurrentAssets := StatementSection{
		Name:        "Non-Current Assets / Aset Tidak Lancar",
		Type:        "NON_CURRENT_ASSET",
		Items:       []StatementItem{},
		Subtotal:    0,
		Order:       2,
	}

	// Categorize asset accounts into current and non-current
	for _, item := range accountHierarchy[models.AccountTypeAsset] {
		// Skip header accounts that don't have balances
		if item.IsHeader && len(item.AccountCode) < 6 {
			continue
		}

		// Determine if current or non-current based on account code
		// Common practice: 1-1-XXX for current assets, 1-2-XXX for non-current
		if strings.HasPrefix(item.AccountCode, "1-1") {
			currentAssets.Items = append(currentAssets.Items, item)
			currentAssets.Subtotal += item.Balance
		} else if strings.HasPrefix(item.AccountCode, "1-2") {
			nonCurrentAssets.Items = append(nonCurrentAssets.Items, item)
			nonCurrentAssets.Subtotal += item.Balance
		} else {
			// Default to non-current if code doesn't match patterns
			nonCurrentAssets.Items = append(nonCurrentAssets.Items, item)
			nonCurrentAssets.Subtotal += item.Balance
		}
	}

	// Add subsections to asset section
	assetSection.Subsections = append(assetSection.Subsections, currentAssets, nonCurrentAssets)
	assetSection.Subtotal = currentAssets.Subtotal + nonCurrentAssets.Subtotal

	return assetSection
}

// createLiabilitySection creates the liability section of the balance sheet
func (srs *StandardizedReportService) createLiabilitySection(accountHierarchy map[string][]StatementItem, comparative bool) StatementSection {
	liabilitySection := StatementSection{
		Name:        "LIABILITIES / KEWAJIBAN",
		Type:        models.AccountTypeLiability,
		Items:       []StatementItem{},
		Subtotal:    0,
		Order:       2,
		Subsections: []StatementSection{},
	}

	// Current Liabilities Subsection
	currentLiabilities := StatementSection{
		Name:        "Current Liabilities / Kewajiban Lancar",
		Type:        "CURRENT_LIABILITY",
		Items:       []StatementItem{},
		Subtotal:    0,
		Order:       1,
	}

	// Non-Current Liabilities Subsection
	nonCurrentLiabilities := StatementSection{
		Name:        "Non-Current Liabilities / Kewajiban Tidak Lancar",
		Type:        "NON_CURRENT_LIABILITY",
		Items:       []StatementItem{},
		Subtotal:    0,
		Order:       2,
	}

	// Categorize liability accounts into current and non-current
	for _, item := range accountHierarchy[models.AccountTypeLiability] {
		// Skip header accounts that don't have balances
		if item.IsHeader && len(item.AccountCode) < 6 {
			continue
		}

		// Determine if current or non-current based on account code
		// Common practice: 2-1-XXX for current liabilities, 2-2-XXX for non-current
		if strings.HasPrefix(item.AccountCode, "2-1") {
			currentLiabilities.Items = append(currentLiabilities.Items, item)
			currentLiabilities.Subtotal += item.Balance
		} else if strings.HasPrefix(item.AccountCode, "2-2") {
			nonCurrentLiabilities.Items = append(nonCurrentLiabilities.Items, item)
			nonCurrentLiabilities.Subtotal += item.Balance
		} else {
			// Default to current if code doesn't match patterns
			currentLiabilities.Items = append(currentLiabilities.Items, item)
			currentLiabilities.Subtotal += item.Balance
		}
	}

	// Add subsections to liability section
	liabilitySection.Subsections = append(liabilitySection.Subsections, currentLiabilities, nonCurrentLiabilities)
	liabilitySection.Subtotal = currentLiabilities.Subtotal + nonCurrentLiabilities.Subtotal

	return liabilitySection
}

// createEquitySection creates the equity section of the balance sheet
func (srs *StandardizedReportService) createEquitySection(accountHierarchy map[string][]StatementItem, comparative bool) StatementSection {
	equitySection := StatementSection{
		Name:        "EQUITY / EKUITAS",
		Type:        models.AccountTypeEquity,
		Items:       []StatementItem{},
		Subtotal:    0,
		Order:       3,
		Subsections: []StatementSection{},
	}

	// Capital & Reserves Subsection
	capitalReserves := StatementSection{
		Name:        "Capital & Reserves / Modal & Cadangan",
		Type:        "CAPITAL",
		Items:       []StatementItem{},
		Subtotal:    0,
		Order:       1,
	}

	// Retained Earnings Subsection
	retainedEarnings := StatementSection{
		Name:        "Retained Earnings / Laba Ditahan",
		Type:        "RETAINED_EARNINGS",
		Items:       []StatementItem{},
		Subtotal:    0,
		Order:       2,
	}

	// Categorize equity accounts
	for _, item := range accountHierarchy[models.AccountTypeEquity] {
		// Skip header accounts that don't have balances
		if item.IsHeader && len(item.AccountCode) < 6 {
			continue
		}

		// Determine the appropriate subsection
		// Common practice: 3-1-XXX for capital, 3-2-XXX for retained earnings
		if strings.HasPrefix(item.AccountCode, "3-1") {
			capitalReserves.Items = append(capitalReserves.Items, item)
			capitalReserves.Subtotal += item.Balance
		} else if strings.HasPrefix(item.AccountCode, "3-2") || 
				strings.Contains(strings.ToLower(item.AccountName), "retained") || 
				strings.Contains(strings.ToLower(item.AccountName), "laba ditahan") {
			retainedEarnings.Items = append(retainedEarnings.Items, item)
			retainedEarnings.Subtotal += item.Balance
		} else {
			// Default to capital if code doesn't match patterns
			capitalReserves.Items = append(capitalReserves.Items, item)
			capitalReserves.Subtotal += item.Balance
		}
	}

	// Add subsections to equity section
	equitySection.Subsections = append(equitySection.Subsections, capitalReserves, retainedEarnings)
	equitySection.Subtotal = capitalReserves.Subtotal + retainedEarnings.Subtotal

	return equitySection
}

// calculateBalanceSheetRatios calculates key financial ratios from the balance sheet
func (srs *StandardizedReportService) calculateBalanceSheetRatios(totals map[string]float64, accountHierarchy map[string][]StatementItem) map[string]float64 {
	ratios := make(map[string]float64)

	// Extract required totals
	totalAssets := totals["total_assets"]
	totalLiabilities := totals["total_liabilities"]
	totalEquity := totals["total_equity"]

	// Get current assets and current liabilities
	var currentAssets, currentLiabilities float64
	for _, item := range accountHierarchy[models.AccountTypeAsset] {
		if strings.HasPrefix(item.AccountCode, "1-1") { // Current assets typically start with 1-1
			currentAssets += item.Balance
		}
	}

	for _, item := range accountHierarchy[models.AccountTypeLiability] {
		if strings.HasPrefix(item.AccountCode, "2-1") { // Current liabilities typically start with 2-1
			currentLiabilities += item.Balance
		}
	}

	// Calculate ratios
	// 1. Current Ratio = Current Assets / Current Liabilities
	if currentLiabilities > 0 {
		ratios["current_ratio"] = currentAssets / currentLiabilities
	} else {
		ratios["current_ratio"] = 0
	}

	// 2. Debt to Equity Ratio = Total Liabilities / Total Equity
	if totalEquity > 0 {
		ratios["debt_to_equity"] = totalLiabilities / totalEquity
	} else {
		ratios["debt_to_equity"] = 0
	}

	// 3. Debt Ratio = Total Liabilities / Total Assets
	if totalAssets > 0 {
		ratios["debt_ratio"] = totalLiabilities / totalAssets
	} else {
		ratios["debt_ratio"] = 0
	}

	// 4. Equity Ratio = Total Equity / Total Assets
	if totalAssets > 0 {
		ratios["equity_ratio"] = totalEquity / totalAssets
	} else {
		ratios["equity_ratio"] = 0
	}

	// 5. Working Capital = Current Assets - Current Liabilities
	ratios["working_capital"] = currentAssets - currentLiabilities

	return ratios
}

// generateBalanceSheetNotes generates notes for the balance sheet
func (srs *StandardizedReportService) generateBalanceSheetNotes(totals map[string]float64) []string {
	notes := []string{}

	// Note about accounting standards
	notes = append(notes, "This balance sheet is prepared in accordance with SAK (Indonesian GAAP)")

	// Note about rounding
	notes = append(notes, "All figures are rounded to the nearest whole unit of currency")

	// Add verification of accounting equation
	totalAssets := totals["total_assets"]
	totalLiabilitiesEquity := totals["total_liabilities_equity"]
	difference := math.Abs(totalAssets - totalLiabilitiesEquity)

	if difference < 0.01 {
		notes = append(notes, "The balance sheet is in balance (Assets = Liabilities + Equity)")
	} else {
		notes = append(notes, fmt.Sprintf("Warning: The balance sheet is out of balance by %s %0.2f", 
			srs.companyProfile.Currency, difference))
	}

	// Add note about comparative figures if needed
	notes = append(notes, "Comparative figures, where presented, may be restated to conform with current year presentation")

	return notes
}

// createRevenueSection creates the revenue section of the P&L statement
func (srs *StandardizedReportService) createRevenueSection(accountHierarchy map[string][]StatementItem, comparative bool) StatementSection {
	revenueSection := StatementSection{
		Name:        "REVENUE / PENDAPATAN",
		Type:        models.AccountTypeRevenue,
		Items:       accountHierarchy[models.AccountTypeRevenue],
		Subtotal:    0,
		Order:       1,
		Subsections: []StatementSection{},
	}

	// Calculate total revenue
	for _, item := range revenueSection.Items {
		if !item.IsHeader {
			revenueSection.Subtotal += item.Balance
		}
	}

	return revenueSection
}

// createCOGSSection creates the cost of goods sold section
func (srs *StandardizedReportService) createCOGSSection(accountHierarchy map[string][]StatementItem, comparative bool) StatementSection {
	cogsSection := StatementSection{
		Name:        "COST OF GOODS SOLD / HARGA POKOK PENJUALAN",
		Type:        models.AccountTypeExpense,
		Items:       []StatementItem{},
		Subtotal:    0,
		Order:       2,
		Subsections: []StatementSection{},
	}

	// Filter COGS-related expense accounts (typically starting with 5-1)
	for _, item := range accountHierarchy[models.AccountTypeExpense] {
		if strings.HasPrefix(item.AccountCode, "5-1") || 
			strings.Contains(strings.ToLower(item.AccountName), "cost of goods") ||
			strings.Contains(strings.ToLower(item.AccountName), "harga pokok") {
			cogsSection.Items = append(cogsSection.Items, item)
			if !item.IsHeader {
				cogsSection.Subtotal += item.Balance
			}
		}
	}

	return cogsSection
}

// createOperatingExpenseSection creates the operating expense section
func (srs *StandardizedReportService) createOperatingExpenseSection(accountHierarchy map[string][]StatementItem, comparative bool) StatementSection {
	operatingExpenseSection := StatementSection{
		Name:        "OPERATING EXPENSES / BIAYA OPERASIONAL",
		Type:        models.AccountTypeExpense,
		Items:       []StatementItem{},
		Subtotal:    0,
		Order:       3,
		Subsections: []StatementSection{},
	}

	// Filter operating expense accounts (typically starting with 6-1)
	for _, item := range accountHierarchy[models.AccountTypeExpense] {
		if strings.HasPrefix(item.AccountCode, "6-1") || 
			(!strings.HasPrefix(item.AccountCode, "5-1") && !strings.HasPrefix(item.AccountCode, "7-")) {
			operatingExpenseSection.Items = append(operatingExpenseSection.Items, item)
			if !item.IsHeader {
				operatingExpenseSection.Subtotal += item.Balance
			}
		}
	}

	return operatingExpenseSection
}

// createNonOperatingSection creates the non-operating income/expense section
func (srs *StandardizedReportService) createNonOperatingSection(accountHierarchy map[string][]StatementItem, comparative bool) StatementSection {
	nonOperatingSection := StatementSection{
		Name:        "NON-OPERATING INCOME / PENDAPATAN LAIN-LAIN",
		Type:        "NON_OPERATING",
		Items:       []StatementItem{},
		Subtotal:    0,
		Order:       4,
		Subsections: []StatementSection{},
	}

	// Filter non-operating accounts
	for _, item := range accountHierarchy[models.AccountTypeRevenue] {
		if strings.HasPrefix(item.AccountCode, "7-") ||
			strings.Contains(strings.ToLower(item.AccountName), "other income") ||
			strings.Contains(strings.ToLower(item.AccountName), "pendapatan lain") {
			nonOperatingSection.Items = append(nonOperatingSection.Items, item)
			if !item.IsHeader {
				nonOperatingSection.Subtotal += item.Balance
			}
		}
	}

	// Add non-operating expenses
	for _, item := range accountHierarchy[models.AccountTypeExpense] {
		if strings.HasPrefix(item.AccountCode, "7-") ||
			strings.Contains(strings.ToLower(item.AccountName), "interest expense") ||
			strings.Contains(strings.ToLower(item.AccountName), "biaya bunga") {
			// For expenses, subtract from the subtotal
			item.Balance = -item.Balance // Make expense negative for non-operating
			nonOperatingSection.Items = append(nonOperatingSection.Items, item)
			if !item.IsHeader {
				nonOperatingSection.Subtotal += item.Balance
			}
		}
	}

	return nonOperatingSection
}

// calculateProfitabilityRatios calculates profitability ratios from P&L totals
func (srs *StandardizedReportService) calculateProfitabilityRatios(totals map[string]float64) map[string]float64 {
	ratios := make(map[string]float64)

	grossRevenue := totals["gross_revenue"]
	grossProfit := totals["gross_profit"]
	operatingIncome := totals["operating_income"]
	netIncome := totals["net_income"]

	// 1. Gross Profit Margin = Gross Profit / Revenue
	if grossRevenue > 0 {
		ratios["gross_profit_margin"] = (grossProfit / grossRevenue) * 100
	} else {
		ratios["gross_profit_margin"] = 0
	}

	// 2. Operating Margin = Operating Income / Revenue
	if grossRevenue > 0 {
		ratios["operating_margin"] = (operatingIncome / grossRevenue) * 100
	} else {
		ratios["operating_margin"] = 0
	}

	// 3. Net Profit Margin = Net Income / Revenue
	if grossRevenue > 0 {
		ratios["net_profit_margin"] = (netIncome / grossRevenue) * 100
	} else {
		ratios["net_profit_margin"] = 0
	}

	return ratios
}

// generateProfitLossNotes generates notes for the P&L statement
func (srs *StandardizedReportService) generateProfitLossNotes(totals map[string]float64) []string {
	notes := []string{}

	// Note about accounting standards
	notes = append(notes, "This profit & loss statement is prepared in accordance with SAK (Indonesian GAAP)")

	// Note about rounding
	notes = append(notes, "All figures are rounded to the nearest whole unit of currency")

	// Note about tax calculation
	notes = append(notes, "Income tax is calculated at the standard corporate rate of 25%")

	return notes
}

// generateStandardizedBalanceSheetPDF generates PDF for balance sheet
func (srs *StandardizedReportService) generateStandardizedBalanceSheetPDF(statement *FinancialStatement, asOfDate time.Time, comparative bool) ([]byte, error) {
	// Create PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set up fonts
	pdf.SetFont("Arial", "B", 16)

	// Company header
	pdf.Cell(190, 10, statement.Header.CompanyName)
	pdf.Ln(7)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 5, "Balance Sheet / Neraca")
	pdf.Ln(5)
	pdf.Cell(190, 5, fmt.Sprintf("As of %s", asOfDate.Format("January 2, 2006")))
	pdf.Ln(10)

	// Add sections
	for _, section := range statement.Sections {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 8, section.Name)
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 9)
		for _, item := range section.Items {
			pdf.Cell(120, 6, "  "+item.AccountName)
			pdf.Cell(70, 6, fmt.Sprintf("%s %.2f", statement.Header.Currency, item.Balance))
			pdf.Ln(6)
		}

		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(120, 6, fmt.Sprintf("Total %s", section.Name))
		pdf.Cell(70, 6, fmt.Sprintf("%s %.2f", statement.Header.Currency, section.Subtotal))
		pdf.Ln(10)
	}

	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// generateStandardizedBalanceSheetExcel generates Excel for balance sheet
func (srs *StandardizedReportService) generateStandardizedBalanceSheetExcel(statement *FinancialStatement, asOfDate time.Time, comparative bool) ([]byte, error) {
	f := excelize.NewFile()
	sheetName := "Balance Sheet"

	// Create sheet and make it active
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create Excel sheet: %v", err)
	}
	f.SetActiveSheet(index)

	// Set headers
	f.SetCellValue(sheetName, "A1", statement.Header.CompanyName)
	f.SetCellValue(sheetName, "A2", "Balance Sheet / Neraca")
	f.SetCellValue(sheetName, "A3", fmt.Sprintf("As of %s", asOfDate.Format("January 2, 2006")))

	row := 5
	for _, section := range statement.Sections {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), section.Name)
		row++

		for _, item := range section.Items {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "  "+item.AccountName)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.Balance)
			row++
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("Total %s", section.Name))
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), section.Subtotal)
		row += 2
	}

	// Delete the default sheet
	f.DeleteSheet("Sheet1")

	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %v", err)
	}

	return buf.Bytes(), nil
}

// generateStandardizedProfitLossPDF generates PDF for P&L statement
func (srs *StandardizedReportService) generateStandardizedProfitLossPDF(statement *FinancialStatement, startDate, endDate time.Time, comparative bool) ([]byte, error) {
	// Similar implementation to balance sheet PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set up fonts
	pdf.SetFont("Arial", "B", 16)

	// Company header
	pdf.Cell(190, 10, statement.Header.CompanyName)
	pdf.Ln(7)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 5, "Profit & Loss Statement / Laporan Laba Rugi")
	pdf.Ln(5)
	pdf.Cell(190, 5, fmt.Sprintf("For the period %s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	pdf.Ln(10)

	// Add sections similar to balance sheet
	for _, section := range statement.Sections {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 8, section.Name)
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 9)
		for _, item := range section.Items {
			pdf.Cell(120, 6, "  "+item.AccountName)
			pdf.Cell(70, 6, fmt.Sprintf("%s %.2f", statement.Header.Currency, item.Balance))
			pdf.Ln(6)
		}

		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(120, 6, fmt.Sprintf("Total %s", section.Name))
		pdf.Cell(70, 6, fmt.Sprintf("%s %.2f", statement.Header.Currency, section.Subtotal))
		pdf.Ln(10)
	}

	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// generateStandardizedProfitLossExcel generates Excel for P&L statement
func (srs *StandardizedReportService) generateStandardizedProfitLossExcel(statement *FinancialStatement, startDate, endDate time.Time, comparative bool) ([]byte, error) {
	f := excelize.NewFile()
	sheetName := "Profit Loss"

	// Create sheet and make it active
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create Excel sheet: %v", err)
	}
	f.SetActiveSheet(index)

	// Set headers
	f.SetCellValue(sheetName, "A1", statement.Header.CompanyName)
	f.SetCellValue(sheetName, "A2", "Profit & Loss Statement / Laporan Laba Rugi")
	f.SetCellValue(sheetName, "A3", fmt.Sprintf("For the period %s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))

	row := 5
	for _, section := range statement.Sections {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), section.Name)
		row++

		for _, item := range section.Items {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "  "+item.AccountName)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.Balance)
			row++
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("Total %s", section.Name))
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), section.Subtotal)
		row += 2
	}

	// Delete the default sheet
	f.DeleteSheet("Sheet1")

	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %v", err)
	}

	return buf.Bytes(), nil
}
