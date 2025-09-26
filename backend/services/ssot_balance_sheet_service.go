package services

import (
	"fmt"
	"strings"
	"time"
	"gorm.io/gorm"
)

// SSOTBalanceSheetService generates Balance Sheet reports from SSOT Journal System
type SSOTBalanceSheetService struct {
	db *gorm.DB
}

// NewSSOTBalanceSheetService creates a new SSOT Balance Sheet service
func NewSSOTBalanceSheetService(db *gorm.DB) *SSOTBalanceSheetService {
	return &SSOTBalanceSheetService{
		db: db,
	}
}

// SSOTBalanceSheetData represents the comprehensive Balance Sheet structure for SSOT
type SSOTBalanceSheetData struct {
	Company               CompanyInfo              `json:"company"`
	AsOfDate              time.Time                `json:"as_of_date"`
	Currency              string                   `json:"currency"`
	
	// Assets Section
	Assets struct {
		CurrentAssets struct {
			Cash          float64              `json:"cash"`
			Receivables   float64              `json:"receivables"`
			Inventory     float64              `json:"inventory"`
			PrepaidExpenses float64            `json:"prepaid_expenses"`
			OtherCurrentAssets float64         `json:"other_current_assets"`
			TotalCurrentAssets float64         `json:"total_current_assets"`
			Items         []BSAccountItem      `json:"items"`
		} `json:"current_assets"`
		
		NonCurrentAssets struct {
			FixedAssets   float64              `json:"fixed_assets"`
			IntangibleAssets float64           `json:"intangible_assets"`
			Investments   float64              `json:"investments"`
			OtherNonCurrentAssets float64      `json:"other_non_current_assets"`
			TotalNonCurrentAssets float64      `json:"total_non_current_assets"`
			Items         []BSAccountItem      `json:"items"`
		} `json:"non_current_assets"`
		
		TotalAssets   float64                `json:"total_assets"`
	} `json:"assets"`
	
	// Liabilities Section
	Liabilities struct {
		CurrentLiabilities struct {
			AccountsPayable    float64           `json:"accounts_payable"`
			ShortTermDebt      float64           `json:"short_term_debt"`
			AccruedLiabilities float64           `json:"accrued_liabilities"`
			TaxPayable         float64           `json:"tax_payable"`
			OtherCurrentLiabilities float64      `json:"other_current_liabilities"`
			TotalCurrentLiabilities float64      `json:"total_current_liabilities"`
			Items         []BSAccountItem       `json:"items"`
		} `json:"current_liabilities"`
		
		NonCurrentLiabilities struct {
			LongTermDebt       float64           `json:"long_term_debt"`
			DeferredTax        float64           `json:"deferred_tax"`
			OtherNonCurrentLiabilities float64  `json:"other_non_current_liabilities"`
			TotalNonCurrentLiabilities float64  `json:"total_non_current_liabilities"`
			Items         []BSAccountItem       `json:"items"`
		} `json:"non_current_liabilities"`
		
		TotalLiabilities float64              `json:"total_liabilities"`
	} `json:"liabilities"`
	
	// Equity Section
	Equity struct {
		ShareCapital       float64            `json:"share_capital"`
		RetainedEarnings   float64            `json:"retained_earnings"`
		OtherEquity        float64            `json:"other_equity"`
		TotalEquity        float64            `json:"total_equity"`
		Items              []BSAccountItem    `json:"items"`
	} `json:"equity"`
	
	// Balance Check
	TotalLiabilitiesAndEquity float64        `json:"total_liabilities_and_equity"`
	IsBalanced                 bool          `json:"is_balanced"`
	BalanceDifference          float64       `json:"balance_difference"`
	
	GeneratedAt               time.Time      `json:"generated_at"`
	Enhanced                  bool           `json:"enhanced"`
	
	// Account Details for Drilldown
	AccountDetails            []SSOTAccountBalance `json:"account_details,omitempty"`
}

// BSAccountItem represents an account item within a Balance Sheet section
type BSAccountItem struct {
	AccountCode   string  `json:"account_code"`
	AccountName   string  `json:"account_name"`
	Amount        float64 `json:"amount"`
	AccountID     uint    `json:"account_id,omitempty"`
}

// GenerateSSOTBalanceSheet generates Balance Sheet statement from SSOT journal system
func (s *SSOTBalanceSheetService) GenerateSSOTBalanceSheet(asOfDate string) (*SSOTBalanceSheetData, error) {
	// Parse date
	asOf, err := time.Parse("2006-01-02", asOfDate)
	if err != nil {
		return nil, fmt.Errorf("invalid as of date format: %v", err)
	}
	
	// Get account balances from SSOT journal entries up to the specified date
	accountBalances, err := s.getAccountBalancesFromSSOT(asOfDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get account balances: %v", err)
	}
	
	// Generate Balance Sheet data structure
	bsData := s.generateBalanceSheetFromBalances(accountBalances, asOf)
	
	return bsData, nil
}

// getAccountBalancesFromSSOT retrieves account balances from SSOT journal system up to a specific date
func (s *SSOTBalanceSheetService) getAccountBalancesFromSSOT(asOfDate string) ([]SSOTAccountBalance, error) {
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
		WHERE ((uje.status = 'POSTED' AND uje.entry_date <= ?) OR uje.status IS NULL)
		  AND COALESCE(a.is_header, false) = false
		GROUP BY a.id, a.code, a.name, a.type
		HAVING a.type IN ('ASSET', 'LIABILITY', 'EQUITY')
		ORDER BY a.code
	`
	
	if err := s.db.Raw(query, asOfDate).Scan(&balances).Error; err != nil {
		return nil, fmt.Errorf("error executing account balances query: %v", err)
	}
	
	// Debug logging
	fmt.Printf("[DEBUG] Balance Sheet Service: Found %d account balances for date %s\n", len(balances), asOfDate)
	
	// Check if we have any actual journal activity (non-zero balances)
	hasJournalActivity := false
	for _, balance := range balances {
		if balance.DebitTotal != 0 || balance.CreditTotal != 0 {
			hasJournalActivity = true
			break
		}
	}
	
	for i, balance := range balances {
		if i < 5 { // Log first 5 accounts
			fmt.Printf("[DEBUG] Account %s (%s): Debit=%.2f, Credit=%.2f, Net=%.2f\n", 
				balance.AccountCode, balance.AccountName, balance.DebitTotal, balance.CreditTotal, balance.NetBalance)
		}
	}
	
	// If no journal activity found for this date, fall back to account balances
	if !hasJournalActivity {
		fmt.Printf("[DEBUG] No journal activity found for date %s, falling back to account.balance\n", asOfDate)
		return s.getAccountBalancesFromAccountTable()
	}
	
	return balances, nil
}

// getAccountBalancesFromAccountTable gets account balances directly from accounts.balance when SSOT data is not available
func (s *SSOTBalanceSheetService) getAccountBalancesFromAccountTable() ([]SSOTAccountBalance, error) {
	var balances []SSOTAccountBalance
	
	query := `
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			0 as debit_total,
			0 as credit_total,
			a.balance as net_balance
		FROM accounts a
		WHERE a.type IN ('ASSET', 'LIABILITY', 'EQUITY')
		  AND COALESCE(a.is_header, false) = false
		ORDER BY a.code
	`
	
	if err := s.db.Raw(query).Scan(&balances).Error; err != nil {
		return nil, fmt.Errorf("error executing fallback account balances query: %v", err)
	}
	
	fmt.Printf("[DEBUG] Fallback method: Found %d accounts with direct balances\n", len(balances))
	for i, balance := range balances {
		if i < 5 && balance.NetBalance != 0 { // Log first 5 non-zero accounts
			fmt.Printf("[DEBUG] Fallback Account %s (%s): Balance=%.2f\n", 
				balance.AccountCode, balance.AccountName, balance.NetBalance)
		}
	}
	
	return balances, nil
}

// calculateNetIncome calculates net income (Revenue - Expenses) from SSOT journal system
func (s *SSOTBalanceSheetService) calculateNetIncome(asOfDate string) float64 {
	// NOTE: Scanning a single aggregate value into a basic type with GORM Scan can be unreliable.
	// To be robust, scan into a struct with a named column and then read the field.
	type niRow struct{ NetIncome float64 `gorm:"column:net_income"` }
	var row niRow
	var netIncome float64

	query := `
		SELECT 
			COALESCE(SUM(
				CASE 
					WHEN UPPER(a.type) = 'REVENUE' THEN 
						COALESCE(ujl.credit_amount, 0) - COALESCE(ujl.debit_amount, 0)
					WHEN UPPER(a.type) = 'EXPENSE' THEN 
						COALESCE(ujl.debit_amount, 0) - COALESCE(ujl.credit_amount, 0)
					ELSE 0
				END
			), 0) as net_income
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE ((uje.status = 'POSTED' AND uje.entry_date <= ?) OR uje.status IS NULL)
		AND UPPER(a.type) IN ('REVENUE', 'EXPENSE')
		AND COALESCE(a.is_header, false) = false
	`

	if err := s.db.Raw(query, asOfDate).Scan(&row).Error; err == nil {
		netIncome = row.NetIncome
		fmt.Printf("[DEBUG] Net Income calculated from SSOT: %.2f\n", netIncome)
	} else {
		fmt.Printf("[DEBUG] Failed to get net income from SSOT, falling back to accounts.balance: %v\n", err)
		// If query fails, try to get from account balances directly (fallback for environments without SSOT data)
		var revenue, expense float64
		s.db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE UPPER(type) = 'REVENUE'`).Scan(&revenue)
		s.db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE UPPER(type) = 'EXPENSE'`).Scan(&expense)
		netIncome = revenue - expense
		fmt.Printf("[DEBUG] Net Income from fallback - Revenue: %.2f, Expense: %.2f, Net: %.2f\n", revenue, expense, netIncome)
	}

	// Net Income calculation: Revenue - Expenses
	// For Revenue accounts: Credit increases balance (positive net income)
	// For Expense accounts: Debit increases balance (negative net income)
	return netIncome
}

// generateBalanceSheetFromBalances creates the Balance Sheet structure from account balances
func (s *SSOTBalanceSheetService) generateBalanceSheetFromBalances(balances []SSOTAccountBalance, asOf time.Time) *SSOTBalanceSheetData {
	bsData := &SSOTBalanceSheetData{
		Company: CompanyInfo{
			Name: "PT. Sistem Akuntansi",
		},
		AsOfDate:    asOf,
		Currency:    "IDR",
		Enhanced:    true,
		GeneratedAt: time.Now(),
	}
	
	// Initialize sections
	bsData.Assets.CurrentAssets.Items = []BSAccountItem{}
	bsData.Assets.NonCurrentAssets.Items = []BSAccountItem{}
	bsData.Liabilities.CurrentLiabilities.Items = []BSAccountItem{}
	bsData.Liabilities.NonCurrentLiabilities.Items = []BSAccountItem{}
	bsData.Equity.Items = []BSAccountItem{}
	bsData.AccountDetails = balances
	
	// Calculate Net Income from Revenue and Expense accounts
	netIncome := s.calculateNetIncome(asOf.Format("2006-01-02"))
	
	// Process each account balance
	for _, balance := range balances {
		code := balance.AccountCode
		amount := balance.NetBalance
		
		// Abaikan akun dengan saldo nol agar hanya data relevan yang tampil
		if amount == 0 {
			continue
		}
		
		item := BSAccountItem{
			AccountCode: balance.AccountCode,
			AccountName: balance.AccountName,
			Amount:      amount,
			AccountID:   balance.AccountID,
		}
		
		// Categorize accounts based on code ranges (following Indonesian chart of accounts)
		// Special handling for PPN Masukan (2102) - should be treated as asset regardless of type in database
		if code == "2102" || strings.Contains(strings.ToLower(item.AccountName), "ppn masukan") {
			fmt.Printf("[DEBUG] Special handling for PPN Masukan: %s - %s (%.2f)\n", code, item.AccountName, amount)
			s.categorizeAssetAccount(bsData, item, code)
		} else {
			switch balance.AccountType {
			case "ASSET":
				s.categorizeAssetAccount(bsData, item, code)
			case "LIABILITY":
				s.categorizeLiabilityAccount(bsData, item, code)
			case "EQUITY":
				s.categorizeEquityAccount(bsData, item, code)
			}
		}
	}
	
		// Tambahkan Net Income ke Retained Earnings dan tampilkan sebagai baris khusus
		if netIncome != 0 {
			bsData.Equity.RetainedEarnings += netIncome
			netIncomeItem := BSAccountItem{
				AccountCode: "NET_INCOME",
				AccountName: "Laba/Rugi Berjalan",
				Amount:      netIncome,
			}
			bsData.Equity.Items = append(bsData.Equity.Items, netIncomeItem)
		}
	
	// Calculate totals and check balance
	s.calculateBalanceSheetTotals(bsData)
	
	return bsData
}

// categorizeAssetAccount categorizes asset accounts into current and non-current assets
func (s *SSOTBalanceSheetService) categorizeAssetAccount(bsData *SSOTBalanceSheetData, item BSAccountItem, code string) {
	switch {
	// Current Assets (11xx)
	case strings.HasPrefix(code, "110"): // Cash accounts
		bsData.Assets.CurrentAssets.Cash += item.Amount
		bsData.Assets.CurrentAssets.Items = append(bsData.Assets.CurrentAssets.Items, item)
	
	case strings.HasPrefix(code, "112"), strings.HasPrefix(code, "120"): // Accounts Receivable
		bsData.Assets.CurrentAssets.Receivables += item.Amount
		bsData.Assets.CurrentAssets.Items = append(bsData.Assets.CurrentAssets.Items, item)
	
	case strings.HasPrefix(code, "113"), strings.HasPrefix(code, "130"): // Inventory
		bsData.Assets.CurrentAssets.Inventory += item.Amount
		bsData.Assets.CurrentAssets.Items = append(bsData.Assets.CurrentAssets.Items, item)
	
	case strings.HasPrefix(code, "114"), strings.HasPrefix(code, "115"): // Prepaid expenses
		bsData.Assets.CurrentAssets.PrepaidExpenses += item.Amount
		bsData.Assets.CurrentAssets.Items = append(bsData.Assets.CurrentAssets.Items, item)
	
	case strings.HasPrefix(code, "11"): // Other current assets
		bsData.Assets.CurrentAssets.OtherCurrentAssets += item.Amount
		bsData.Assets.CurrentAssets.Items = append(bsData.Assets.CurrentAssets.Items, item)
		
	// Special case: PPN Masukan should be current asset (input VAT)
	case code == "2102" || strings.Contains(strings.ToLower(item.AccountName), "ppn masukan"):
		bsData.Assets.CurrentAssets.OtherCurrentAssets += item.Amount
		bsData.Assets.CurrentAssets.Items = append(bsData.Assets.CurrentAssets.Items, item)
	
	// Non-Current Assets (12xx, 13xx, 14xx, 15xx)
	case strings.HasPrefix(code, "12"), strings.HasPrefix(code, "16"), strings.HasPrefix(code, "17"): // Fixed Assets
		bsData.Assets.NonCurrentAssets.FixedAssets += item.Amount
		bsData.Assets.NonCurrentAssets.Items = append(bsData.Assets.NonCurrentAssets.Items, item)
	
	case strings.HasPrefix(code, "14"): // Intangible Assets
		bsData.Assets.NonCurrentAssets.IntangibleAssets += item.Amount
		bsData.Assets.NonCurrentAssets.Items = append(bsData.Assets.NonCurrentAssets.Items, item)
	
	case strings.HasPrefix(code, "15"): // Investments
		bsData.Assets.NonCurrentAssets.Investments += item.Amount
		bsData.Assets.NonCurrentAssets.Items = append(bsData.Assets.NonCurrentAssets.Items, item)
	
	default: // Other non-current assets
		bsData.Assets.NonCurrentAssets.OtherNonCurrentAssets += item.Amount
		bsData.Assets.NonCurrentAssets.Items = append(bsData.Assets.NonCurrentAssets.Items, item)
	}
}

// categorizeLiabilityAccount categorizes liability accounts into current and non-current liabilities
func (s *SSOTBalanceSheetService) categorizeLiabilityAccount(bsData *SSOTBalanceSheetData, item BSAccountItem, code string) {
	switch {
	// Current Liabilities (21xx)
	case strings.HasPrefix(code, "210"): // Accounts Payable
		bsData.Liabilities.CurrentLiabilities.AccountsPayable += item.Amount
		bsData.Liabilities.CurrentLiabilities.Items = append(bsData.Liabilities.CurrentLiabilities.Items, item)
	
	case strings.HasPrefix(code, "211"): // Short-term debt
		bsData.Liabilities.CurrentLiabilities.ShortTermDebt += item.Amount
		bsData.Liabilities.CurrentLiabilities.Items = append(bsData.Liabilities.CurrentLiabilities.Items, item)
	
	case strings.HasPrefix(code, "212"), strings.HasPrefix(code, "213"): // Accrued liabilities and taxes
		if strings.Contains(strings.ToLower(item.AccountName), "tax") || strings.Contains(strings.ToLower(item.AccountName), "pajak") {
			bsData.Liabilities.CurrentLiabilities.TaxPayable += item.Amount
		} else {
			bsData.Liabilities.CurrentLiabilities.AccruedLiabilities += item.Amount
		}
		bsData.Liabilities.CurrentLiabilities.Items = append(bsData.Liabilities.CurrentLiabilities.Items, item)
	
	case strings.HasPrefix(code, "21"): // Other current liabilities
		bsData.Liabilities.CurrentLiabilities.OtherCurrentLiabilities += item.Amount
		bsData.Liabilities.CurrentLiabilities.Items = append(bsData.Liabilities.CurrentLiabilities.Items, item)
	
	// Non-Current Liabilities (22xx, 23xx)
	case strings.HasPrefix(code, "22"): // Long-term debt
		bsData.Liabilities.NonCurrentLiabilities.LongTermDebt += item.Amount
		bsData.Liabilities.NonCurrentLiabilities.Items = append(bsData.Liabilities.NonCurrentLiabilities.Items, item)
	
	case strings.HasPrefix(code, "23"): // Other non-current liabilities
		if strings.Contains(strings.ToLower(item.AccountName), "tax") || strings.Contains(strings.ToLower(item.AccountName), "pajak") {
			bsData.Liabilities.NonCurrentLiabilities.DeferredTax += item.Amount
		} else {
			bsData.Liabilities.NonCurrentLiabilities.OtherNonCurrentLiabilities += item.Amount
		}
		bsData.Liabilities.NonCurrentLiabilities.Items = append(bsData.Liabilities.NonCurrentLiabilities.Items, item)
	
	default: // Other liabilities
		bsData.Liabilities.NonCurrentLiabilities.OtherNonCurrentLiabilities += item.Amount
		bsData.Liabilities.NonCurrentLiabilities.Items = append(bsData.Liabilities.NonCurrentLiabilities.Items, item)
	}
}

// categorizeEquityAccount categorizes equity accounts
func (s *SSOTBalanceSheetService) categorizeEquityAccount(bsData *SSOTBalanceSheetData, item BSAccountItem, code string) {
	switch {
	case strings.HasPrefix(code, "31"): // Share Capital
		bsData.Equity.ShareCapital += item.Amount
		bsData.Equity.Items = append(bsData.Equity.Items, item)
	
	case strings.HasPrefix(code, "32"): // Retained Earnings
		bsData.Equity.RetainedEarnings += item.Amount
		bsData.Equity.Items = append(bsData.Equity.Items, item)
	
	default: // Other Equity
		bsData.Equity.OtherEquity += item.Amount
		bsData.Equity.Items = append(bsData.Equity.Items, item)
	}
}

// calculateBalanceSheetTotals calculates all totals and checks if the balance sheet is balanced
func (s *SSOTBalanceSheetService) calculateBalanceSheetTotals(bsData *SSOTBalanceSheetData) {
	// Calculate current assets total
	bsData.Assets.CurrentAssets.TotalCurrentAssets = 
		bsData.Assets.CurrentAssets.Cash +
		bsData.Assets.CurrentAssets.Receivables +
		bsData.Assets.CurrentAssets.Inventory +
		bsData.Assets.CurrentAssets.PrepaidExpenses +
		bsData.Assets.CurrentAssets.OtherCurrentAssets
	
	// Calculate non-current assets total
	bsData.Assets.NonCurrentAssets.TotalNonCurrentAssets = 
		bsData.Assets.NonCurrentAssets.FixedAssets +
		bsData.Assets.NonCurrentAssets.IntangibleAssets +
		bsData.Assets.NonCurrentAssets.Investments +
		bsData.Assets.NonCurrentAssets.OtherNonCurrentAssets
	
	// Calculate total assets
	bsData.Assets.TotalAssets = 
		bsData.Assets.CurrentAssets.TotalCurrentAssets +
		bsData.Assets.NonCurrentAssets.TotalNonCurrentAssets
	
	// Calculate current liabilities total
	bsData.Liabilities.CurrentLiabilities.TotalCurrentLiabilities = 
		bsData.Liabilities.CurrentLiabilities.AccountsPayable +
		bsData.Liabilities.CurrentLiabilities.ShortTermDebt +
		bsData.Liabilities.CurrentLiabilities.AccruedLiabilities +
		bsData.Liabilities.CurrentLiabilities.TaxPayable +
		bsData.Liabilities.CurrentLiabilities.OtherCurrentLiabilities
	
	// Calculate non-current liabilities total
	bsData.Liabilities.NonCurrentLiabilities.TotalNonCurrentLiabilities = 
		bsData.Liabilities.NonCurrentLiabilities.LongTermDebt +
		bsData.Liabilities.NonCurrentLiabilities.DeferredTax +
		bsData.Liabilities.NonCurrentLiabilities.OtherNonCurrentLiabilities
	
	// Calculate total liabilities
	bsData.Liabilities.TotalLiabilities = 
		bsData.Liabilities.CurrentLiabilities.TotalCurrentLiabilities +
		bsData.Liabilities.NonCurrentLiabilities.TotalNonCurrentLiabilities
	
	// Calculate total equity
	bsData.Equity.TotalEquity = 
		bsData.Equity.ShareCapital +
		bsData.Equity.RetainedEarnings +
		bsData.Equity.OtherEquity
	
	// Calculate total liabilities and equity
	bsData.TotalLiabilitiesAndEquity = bsData.Liabilities.TotalLiabilities + bsData.Equity.TotalEquity
	
	// Check if balance sheet is balanced
	tolerance := 0.01 // 1 cent tolerance
	bsData.BalanceDifference = bsData.Assets.TotalAssets - bsData.TotalLiabilitiesAndEquity
	bsData.IsBalanced = (bsData.BalanceDifference >= -tolerance && bsData.BalanceDifference <= tolerance)
}