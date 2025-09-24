package services

import (
	"fmt"
	"strings"
	"time"
	"gorm.io/gorm"
)

// SSOTProfitLossService generates P&L reports from SSOT Journal System
type SSOTProfitLossService struct {
	db *gorm.DB
}

// NewSSOTProfitLossService creates a new SSOT P&L service
func NewSSOTProfitLossService(db *gorm.DB) *SSOTProfitLossService {
	return &SSOTProfitLossService{
		db: db,
	}
}

// SSOTAccountBalance represents account balance from journal entries (renamed to avoid conflict)
type SSOTAccountBalance struct {
	AccountID   uint    `json:"account_id"`
	AccountCode string  `json:"account_code"`
	AccountName string  `json:"account_name"`
	AccountType string  `json:"account_type"`
	DebitTotal  float64 `json:"debit_total"`
	CreditTotal float64 `json:"credit_total"`
	NetBalance  float64 `json:"net_balance"`
}

// SSOTProfitLossData represents the comprehensive P&L structure for SSOT
type SSOTProfitLossData struct {
	Company               CompanyInfo            `json:"company"`
	StartDate             time.Time              `json:"start_date"`
	EndDate               time.Time              `json:"end_date"`
	Currency              string                 `json:"currency"`
	
	// Revenue Section
	Revenue struct {
		SalesRevenue    float64                `json:"sales_revenue"`
		ServiceRevenue  float64                `json:"service_revenue"`
		OtherRevenue    float64                `json:"other_revenue"`
		TotalRevenue    float64                `json:"total_revenue"`
		Items           []PLSectionItem        `json:"items"`
	} `json:"revenue"`
	
	// Cost of Goods Sold
	COGS struct {
		DirectMaterials float64                `json:"direct_materials"`
		DirectLabor     float64                `json:"direct_labor"`
		Manufacturing   float64                `json:"manufacturing"`
		OtherCOGS       float64                `json:"other_cogs"`
		TotalCOGS       float64                `json:"total_cogs"`
		Items           []PLSectionItem        `json:"items"`
	} `json:"cost_of_goods_sold"`
	
	GrossProfit       float64                `json:"gross_profit"`
	GrossProfitMargin float64                `json:"gross_profit_margin"`
	
	// Operating Expenses
	OperatingExpenses struct {
		Administrative struct {
			Subtotal float64        `json:"subtotal"`
			Items    []PLSectionItem `json:"items"`
		} `json:"administrative"`
		SellingMarketing struct {
			Subtotal float64        `json:"subtotal"`
			Items    []PLSectionItem `json:"items"`
		} `json:"selling_marketing"`
		General struct {
			Subtotal float64        `json:"subtotal"`
			Items    []PLSectionItem `json:"items"`
		} `json:"general"`
		TotalOpEx float64 `json:"total_opex"`
	} `json:"operating_expenses"`
	
	OperatingIncome   float64                `json:"operating_income"`
	OperatingMargin   float64                `json:"operating_margin"`
	
	// Other Income/Expenses
	OtherIncome       float64                `json:"other_income"`
	OtherExpenses     float64                `json:"other_expenses"`
	
	// Tax and Final Results
	EBITDA            float64                `json:"ebitda"`
	EBITDAMargin      float64                `json:"ebitda_margin"`
	IncomeBeforeTax   float64                `json:"income_before_tax"`
	TaxExpense        float64                `json:"tax_expense"`
	NetIncome         float64                `json:"net_income"`
	NetIncomeMargin   float64                `json:"net_income_margin"`
	
	GeneratedAt       time.Time              `json:"generated_at"`
	Enhanced          bool                   `json:"enhanced"`
	
	// Account Details for Drilldown
	AccountDetails    []SSOTAccountBalance       `json:"account_details,omitempty"`
}

// PLSectionItem represents an item within a P&L section
type PLSectionItem struct {
	AccountCode   string  `json:"account_code"`
	AccountName   string  `json:"account_name"`
	Amount        float64 `json:"amount"`
	AccountID     uint    `json:"account_id,omitempty"`
}

// GenerateSSOTProfitLoss generates P&L statement from SSOT journal system
func (s *SSOTProfitLossService) GenerateSSOTProfitLoss(startDate, endDate string) (*SSOTProfitLossData, error) {
	// Parse dates
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %v", err)
	}
	
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %v", err)
	}
	
	// Get account balances from SSOT journal entries
	accountBalances, err := s.getAccountBalancesFromSSOT(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get account balances: %v", err)
	}
	
	// Generate P&L data structure
	plData := s.generateProfitLossFromBalances(accountBalances, start, end)
	
	return plData, nil
}

// getAccountBalancesFromSSOT retrieves account balances from SSOT journal system
func (s *SSOTProfitLossService) getAccountBalancesFromSSOT(startDate, endDate string) ([]SSOTAccountBalance, error) {
	var balances []SSOTAccountBalance
	
	query := `
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			COALESCE(SUM(ujl.debit_amount), 0) as debit_total,
			COALESCE(SUM(ujl.credit_amount), 0) as credit_total,
			CASE 
				WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
					COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
				ELSE 
					COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
			END as net_balance
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE uje.status = 'POSTED' 
			AND uje.entry_date >= ? 
			AND uje.entry_date <= ?
		GROUP BY a.id, a.code, a.name, a.type
		HAVING COALESCE(SUM(ujl.debit_amount), 0) > 0 OR COALESCE(SUM(ujl.credit_amount), 0) > 0
		ORDER BY a.code
	`
	
	if err := s.db.Raw(query, startDate, endDate).Scan(&balances).Error; err != nil {
		return nil, fmt.Errorf("error executing account balances query: %v", err)
	}
	
	return balances, nil
}

// generateProfitLossFromBalances creates the P&L structure from account balances
func (s *SSOTProfitLossService) generateProfitLossFromBalances(balances []SSOTAccountBalance, start, end time.Time) *SSOTProfitLossData {
	plData := &SSOTProfitLossData{
		Company: CompanyInfo{
			Name: "PT. Sistem Akuntansi",
		},
		StartDate:   start,
		EndDate:     end,
		Currency:    "IDR",
		Enhanced:    true,
		GeneratedAt: time.Now(),
	}
	
	// Initialize sections
	plData.Revenue.Items = []PLSectionItem{}
	plData.COGS.Items = []PLSectionItem{}
	plData.OperatingExpenses.Administrative.Items = []PLSectionItem{}
	plData.OperatingExpenses.SellingMarketing.Items = []PLSectionItem{}
	plData.OperatingExpenses.General.Items = []PLSectionItem{}
	plData.AccountDetails = balances
	
	// Process each account balance
	for _, balance := range balances {
		code := balance.AccountCode
		amount := balance.NetBalance
		
		// Skip if amount is zero
		if amount == 0 {
			continue
		}
		
		item := PLSectionItem{
			AccountCode: balance.AccountCode,
			AccountName: balance.AccountName,
			Amount:      amount,
			AccountID:   balance.AccountID,
		}
		
		// Categorize accounts based on code ranges (following Indonesian chart of accounts)
		switch {
		// REVENUE ACCOUNTS (4xxx)
		case strings.HasPrefix(code, "40") || strings.HasPrefix(code, "41"):
			// Sales Revenue
			plData.Revenue.SalesRevenue += amount
			plData.Revenue.Items = append(plData.Revenue.Items, item)
			
		case strings.HasPrefix(code, "42"):
			// Service Revenue  
			plData.Revenue.ServiceRevenue += amount
			plData.Revenue.Items = append(plData.Revenue.Items, item)
			
		case strings.HasPrefix(code, "49"):
			// Other Revenue
			plData.Revenue.OtherRevenue += amount
			plData.Revenue.Items = append(plData.Revenue.Items, item)
			
		// COST OF GOODS SOLD (51xx)
		case strings.HasPrefix(code, "510"):
			// Direct materials, direct COGS
			plData.COGS.DirectMaterials += amount
			plData.COGS.Items = append(plData.COGS.Items, item)
			
		case strings.HasPrefix(code, "511"):
			// Direct labor
			plData.COGS.DirectLabor += amount
			plData.COGS.Items = append(plData.COGS.Items, item)
			
		case strings.HasPrefix(code, "512"):
			// Manufacturing overhead
			plData.COGS.Manufacturing += amount
			plData.COGS.Items = append(plData.COGS.Items, item)
			
		case strings.HasPrefix(code, "513"), strings.HasPrefix(code, "514"), strings.HasPrefix(code, "519"):
			// Other COGS
			plData.COGS.OtherCOGS += amount
			plData.COGS.Items = append(plData.COGS.Items, item)

		// OPERATING EXPENSES
		case strings.HasPrefix(code, "52"):
			// Administrative expenses (520x-529x)
			plData.OperatingExpenses.Administrative.Subtotal += amount
			plData.OperatingExpenses.Administrative.Items = append(plData.OperatingExpenses.Administrative.Items, item)
			
		case strings.HasPrefix(code, "53"):
			// Selling & Marketing expenses (530x-539x)
			plData.OperatingExpenses.SellingMarketing.Subtotal += amount
			plData.OperatingExpenses.SellingMarketing.Items = append(plData.OperatingExpenses.SellingMarketing.Items, item)
			
		case strings.HasPrefix(code, "54"), strings.HasPrefix(code, "55"), strings.HasPrefix(code, "56"), 
			 strings.HasPrefix(code, "57"), strings.HasPrefix(code, "58"), strings.HasPrefix(code, "59"):
			// General expenses (540x-599x)
			plData.OperatingExpenses.General.Subtotal += amount
			plData.OperatingExpenses.General.Items = append(plData.OperatingExpenses.General.Items, item)

		// OTHER INCOME/EXPENSES
		case strings.HasPrefix(code, "6"):
			// Other expenses (6xxx)
			plData.OtherExpenses += amount
			
		case strings.HasPrefix(code, "7"):
			// Other income (7xxx)
			plData.OtherIncome += amount
		}
	}
	
	// Calculate totals and ratios
	s.calculatePLTotalsAndRatios(plData)
	
	return plData
}

// calculatePLTotalsAndRatios calculates all totals, subtotals, and financial ratios
func (s *SSOTProfitLossService) calculatePLTotalsAndRatios(plData *SSOTProfitLossData) {
	// Calculate revenue totals
	plData.Revenue.TotalRevenue = plData.Revenue.SalesRevenue + plData.Revenue.ServiceRevenue + plData.Revenue.OtherRevenue
	
	// Calculate COGS totals
	plData.COGS.TotalCOGS = plData.COGS.DirectMaterials + plData.COGS.DirectLabor + plData.COGS.Manufacturing + plData.COGS.OtherCOGS
	
	// Calculate gross profit and margin
	plData.GrossProfit = plData.Revenue.TotalRevenue - plData.COGS.TotalCOGS
	if plData.Revenue.TotalRevenue > 0 {
		plData.GrossProfitMargin = (plData.GrossProfit / plData.Revenue.TotalRevenue) * 100
	}
	
	// Calculate operating expense totals
	plData.OperatingExpenses.TotalOpEx = plData.OperatingExpenses.Administrative.Subtotal + 
		plData.OperatingExpenses.SellingMarketing.Subtotal + 
		plData.OperatingExpenses.General.Subtotal
	
	// Calculate operating income and margin
	plData.OperatingIncome = plData.GrossProfit - plData.OperatingExpenses.TotalOpEx
	if plData.Revenue.TotalRevenue > 0 {
		plData.OperatingMargin = (plData.OperatingIncome / plData.Revenue.TotalRevenue) * 100
	}
	
	// Calculate EBITDA (assume no depreciation/amortization for now)
	plData.EBITDA = plData.OperatingIncome
	if plData.Revenue.TotalRevenue > 0 {
		plData.EBITDAMargin = (plData.EBITDA / plData.Revenue.TotalRevenue) * 100
	}
	
	// Calculate income before tax
	plData.IncomeBeforeTax = plData.OperatingIncome + plData.OtherIncome - plData.OtherExpenses
	
	// Estimate tax expense (assume 25% rate if income is positive)
	if plData.IncomeBeforeTax > 0 {
		plData.TaxExpense = plData.IncomeBeforeTax * 0.25
	}
	
	// Calculate net income and margin
	plData.NetIncome = plData.IncomeBeforeTax - plData.TaxExpense
	if plData.Revenue.TotalRevenue > 0 {
		plData.NetIncomeMargin = (plData.NetIncome / plData.Revenue.TotalRevenue) * 100
	}
}