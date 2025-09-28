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
	DataSource            string                 `json:"data_source"`
	
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
	
// Get account balances from SSOT journal entries (with data source flag)
	accountBalances, source, err := s.getAccountBalancesFromSSOT(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get account balances: %v", err)
	}
	
	// Generate P&L data structure
	plData := s.generateProfitLossFromBalances(accountBalances, start, end)
	plData.DataSource = source
	
return plData, nil
}

// getAccountBalancesFromLegacy retrieves account balances using legacy journal tables as fallback
func (s *SSOTProfitLossService) getAccountBalancesFromLegacy(startDate, endDate string) ([]SSOTAccountBalance, error) {
	var balances []SSOTAccountBalance
	legacyQuery := `
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			COALESCE(SUM(jl.debit_amount), 0) as debit_total,
			COALESCE(SUM(jl.credit_amount), 0) as credit_total,
			CASE 
				WHEN UPPER(a.type) IN ('ASSET', 'EXPENSE') THEN 
					COALESCE(SUM(jl.debit_amount), 0) - COALESCE(SUM(jl.credit_amount), 0)
				ELSE 
					COALESCE(SUM(jl.credit_amount), 0) - COALESCE(SUM(jl.debit_amount), 0)
			END as net_balance
		FROM accounts a
		LEFT JOIN journal_lines jl ON jl.account_id = a.id
		LEFT JOIN journal_entries je ON je.id = jl.journal_entry_id AND je.status = 'POSTED' AND je.deleted_at IS NULL
		WHERE je.entry_date >= ? AND je.entry_date <= ?
		  AND COALESCE(a.is_header, false) = false
		GROUP BY a.id, a.code, a.name, a.type
		HAVING (COALESCE(SUM(jl.debit_amount), 0) <> 0 OR COALESCE(SUM(jl.credit_amount), 0) <> 0)
		ORDER BY a.code`
	if err := s.db.Raw(legacyQuery, startDate, endDate).Scan(&balances).Error; err != nil {
		return nil, fmt.Errorf("legacy account balances query failed: %v", err)
	}
	return balances, nil
}

// getAccountBalancesFromSSOT retrieves account balances from SSOT journal system, with automatic fallbacks
func (s *SSOTProfitLossService) getAccountBalancesFromSSOT(startDate, endDate string) ([]SSOTAccountBalance, string, error) {
	var balances []SSOTAccountBalance
	source := "SSOT"
	
	query := `
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			COALESCE(SUM(ujl.debit_amount), 0) as debit_total,
			COALESCE(SUM(ujl.credit_amount), 0) as credit_total,
			CASE 
				WHEN UPPER(a.type) IN ('ASSET', 'EXPENSE') THEN 
					COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
				ELSE 
					COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
			END as net_balance
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id AND uje.status = 'POSTED' AND uje.deleted_at IS NULL
		WHERE uje.entry_date >= ? AND uje.entry_date <= ?
		  AND COALESCE(a.is_header, false) = false
		GROUP BY a.id, a.code, a.name, a.type
		HAVING (COALESCE(SUM(ujl.debit_amount), 0) <> 0 OR COALESCE(SUM(ujl.credit_amount), 0) <> 0)
		ORDER BY a.code
	`
	
if err := s.db.Raw(query, startDate, endDate).Scan(&balances).Error; err != nil {
		return nil, source, fmt.Errorf("error executing account balances query: %v", err)
	}
	
	// Fallback to legacy journals if SSOT returns no activity OR no PL activity (no 4xxx/5xxx)
if len(balances) == 0 || !hasPLActivity(balances) {
		legacy, lerr := s.getAccountBalancesFromLegacy(startDate, endDate)
		if lerr == nil && len(legacy) > 0 && hasPLActivity(legacy) {
			source = "LEGACY"
			return legacy, source, nil
		}
		// Final fallback: derive from accounts table balances for 4xxx/5xxx if journals have no PL activity
		acctFallback, aerr := s.getPLFromAccountsBalance()
		if aerr == nil && len(acctFallback) > 0 && hasPLActivity(acctFallback) {
			source = "ACCOUNTS"
			return acctFallback, source, nil
		}
	}
	
	return balances, source, nil
}

// hasPLActivity returns true if there is any revenue or expense activity in the balances
func hasPLActivity(balances []SSOTAccountBalance) bool {
	for _, b := range balances {
		if strings.HasPrefix(b.AccountCode, "4") || strings.HasPrefix(b.AccountCode, "5") ||
			strings.EqualFold(b.AccountType, "REVENUE") || strings.EqualFold(b.AccountType, "EXPENSE") {
			if b.DebitTotal != 0 || b.CreditTotal != 0 || b.NetBalance != 0 {
				return true
			}
		}
	}
return false
}

// getPLFromAccountsBalance builds minimal PL balances using accounts.balance when neither SSOT nor legacy journals provide PL activity
func (s *SSOTProfitLossService) getPLFromAccountsBalance() ([]SSOTAccountBalance, error) {
	var balances []SSOTAccountBalance
	query := `
		SELECT 
			id as account_id,
			code as account_code,
			name as account_name,
			type as account_type,
			0 as debit_total,
			0 as credit_total,
			CASE WHEN UPPER(type) = 'EXPENSE' THEN balance ELSE balance END as net_balance
		FROM accounts
		WHERE (code LIKE '4%%' OR code LIKE '5%%')
		  AND COALESCE(is_header,false) = false
		ORDER BY code`
	if err := s.db.Raw(query).Scan(&balances).Error; err != nil {
		return nil, fmt.Errorf("accounts balance PL fallback failed: %v", err)
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