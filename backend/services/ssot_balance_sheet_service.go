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
	// Default to end of current fiscal year if parameter is empty
	if strings.TrimSpace(asOfDate) == "" {
		settingsSvc := NewSettingsService(s.db)
		_, fyEnd, err := settingsSvc.GetCurrentFiscalYearRange()
		if err == nil { asOfDate = fyEnd }
	}
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

	// First, get all accounts with their direct balances to ensure we capture all accounts
	allAccountsQuery := `
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			0 as debit_total,
			0 as credit_total,
			a.balance as net_balance
		FROM accounts a
		WHERE COALESCE(a.is_header, false) = false
		  AND UPPER(a.type) IN ('ASSET', 'LIABILITY', 'EQUITY')
		  AND a.is_active = true
		ORDER BY a.code
	`

	if err := s.db.Raw(allAccountsQuery).Scan(&balances).Error; err != nil {
		return nil, fmt.Errorf("error executing all accounts query: %v", err)
	}

	fmt.Printf("[DEBUG] Retrieved %d accounts with direct balances\n", len(balances))
	
	// Log all liability accounts for debugging
	for _, balance := range balances {
		if strings.ToUpper(balance.AccountType) == "LIABILITY" {
			fmt.Printf("[DEBUG] Liability Account - ID: %d, Code: %s, Name: %s, Type: %s, Balance: %.2f\n",
				balance.AccountID, balance.AccountCode, balance.AccountName, balance.AccountType, balance.NetBalance)
		}
	}
	
	// Create a map of accounts by ID for quick lookup
	accountMap := make(map[uint]*SSOTAccountBalance)
	for i := range balances {
		accountMap[balances[i].AccountID] = &balances[i]
	}

	// Then, get transaction data from unified journal ledger
	transactionQuery := `
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
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id AND uje.status = 'POSTED' AND uje.deleted_at IS NULL AND (uje.entry_date <= ? OR uje.entry_date IS NULL)
		WHERE COALESCE(a.is_header, false) = false
		  AND UPPER(a.type) IN ('ASSET', 'LIABILITY', 'EQUITY')
		  AND a.is_active = true
		GROUP BY a.id, a.code, a.name, a.type
		ORDER BY a.code
	`

	var transactionBalances []SSOTAccountBalance
	if err := s.db.Raw(transactionQuery, asOfDate).Scan(&transactionBalances).Error; err != nil {
		return nil, fmt.Errorf("error executing transaction balances query: %v", err)
	}
	
	// Log transaction data for all liability accounts
	fmt.Printf("[DEBUG] Retrieved %d accounts with transaction data\n", len(transactionBalances))
	for _, tb := range transactionBalances {
		if strings.ToUpper(tb.AccountType) == "LIABILITY" {
			fmt.Printf("[DEBUG] Transaction Data - ID: %d, Code: %s, Name: %s, Type: %s, Debit: %.2f, Credit: %.2f, Net: %.2f\n",
				tb.AccountID, tb.AccountCode, tb.AccountName, tb.AccountType, tb.DebitTotal, tb.CreditTotal, tb.NetBalance)
		}
	}

	// Update accounts with transaction data where available
	// Only update accounts that have actual transaction activity
	for _, tb := range transactionBalances {
		// Log transaction data for UTANG USAHA account for debugging
		if tb.AccountCode == "2101" || strings.Contains(strings.ToUpper(tb.AccountName), "UTANG") {
			fmt.Printf("[DEBUG] Transaction data for %s (%s): Debit=%.2f, Credit=%.2f, Net=%.2f, AccountID=%d\n",
				tb.AccountCode, tb.AccountName, tb.DebitTotal, tb.CreditTotal, tb.NetBalance, tb.AccountID)
		}
		
		// Only update if the account has actual transaction activity
		// (either debit or credit amount is non-zero)
		if tb.DebitTotal != 0 || tb.CreditTotal != 0 {
			if account, exists := accountMap[tb.AccountID]; exists {
				// Log when we're updating an account with transaction data
				if tb.AccountCode == "2101" || strings.Contains(strings.ToUpper(tb.AccountName), "UTANG") {
					fmt.Printf("[DEBUG] Updating %s (%s) with transaction data: %.2f -> %.2f\n",
						tb.AccountCode, tb.AccountName, account.NetBalance, tb.NetBalance)
				}
				account.DebitTotal = tb.DebitTotal
				account.CreditTotal = tb.CreditTotal
				account.NetBalance = tb.NetBalance
			}
		} else {
			// If account has no transaction activity but direct balance is zero, try to calculate it
			if account, exists := accountMap[tb.AccountID]; exists {
				if account.NetBalance == 0 {
					// Try to calculate balance from transactions
					calculatedBalance, err := s.calculateAccountBalanceFromTransactions(tb.AccountID, tb.AccountType)
					if err == nil && calculatedBalance != 0 {
						fmt.Printf("[DEBUG] Calculated balance for %s (%s) from transactions: %.2f\n",
							tb.AccountCode, tb.AccountName, calculatedBalance)
						account.NetBalance = calculatedBalance
					} else if err == nil && calculatedBalance == 0 {
						// If still zero, try to calculate from child accounts
						childBalance, err := s.calculateAccountBalanceFromChildren(tb.AccountID, tb.AccountType)
						if err == nil && childBalance != 0 {
							fmt.Printf("[DEBUG] Calculated balance for %s (%s) from children: %.2f\n",
								tb.AccountCode, tb.AccountName, childBalance)
							account.NetBalance = childBalance
						}
					}
				}
				
				// Log when an account has no transaction activity
				if tb.AccountCode == "2101" || strings.Contains(strings.ToUpper(tb.AccountName), "UTANG") {
					fmt.Printf("[DEBUG] Account %s (%s has no transaction activity, final balance: %.2f\n",
						tb.AccountCode, tb.AccountName, account.NetBalance)
				}
			}
		}
		// If the account has no transactions, it will retain its direct balance from the accounts table
	}

	// If no balances found, try legacy journals fallback
	if len(balances) == 0 {
		legacy, lerr := s.getAccountBalancesFromLegacy(asOfDate)
		if lerr == nil && len(legacy) > 0 {
			fmt.Printf("[DEBUG] SSOT returned no data, using legacy: %d accounts\n", len(legacy))
			return legacy, nil
		}
	}

	fmt.Printf("[DEBUG] Retrieved %d account balances from SSOT for balance sheet\n", len(balances))
	
	// Log some sample balances for debugging
	for i, balance := range balances {
		if i < 10 || balance.AccountCode == "2101" || strings.Contains(strings.ToUpper(balance.AccountName), "UTANG") { // Always log UTANG USAHA for debugging
			fmt.Printf("[DEBUG] Final Account %s (%s) Type: %s, Balance: %.2f, AccountID: %d\n", 
				balance.AccountCode, balance.AccountName, balance.AccountType, balance.NetBalance, balance.AccountID)
		}
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
		WHERE UPPER(a.type) IN ('ASSET', 'LIABILITY', 'EQUITY')
		  AND COALESCE(a.is_header, false) = false
		ORDER BY a.code
	`
	
	if err := s.db.Raw(query).Scan(&balances).Error; err != nil {
		return nil, fmt.Errorf("error executing fallback account balances query: %v", err)
	}
	
	fmt.Printf("[DEBUG] Fallback method: Found %d accounts with direct balances\n", len(balances))
	for i, balance := range balances {
		if i < 10 && balance.NetBalance != 0 { // Log first 10 non-zero accounts
			fmt.Printf("[DEBUG] Fallback Account %s (%s): Type=%s, Balance=%.2f\n", 
				balance.AccountCode, balance.AccountName, balance.AccountType, balance.NetBalance)
		}
	}
	
	return balances, nil
}

// getAccountBalancesFromLegacy retrieves account balances using legacy journal tables up to an as-of date
func (s *SSOTBalanceSheetService) getAccountBalancesFromLegacy(asOfDate string) ([]SSOTAccountBalance, error) {
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
		WHERE (je.entry_date <= ? OR je.entry_date IS NULL)
		  AND COALESCE(a.is_header, false) = false
		  AND UPPER(a.type) IN ('ASSET', 'LIABILITY', 'EQUITY')
		GROUP BY a.id, a.code, a.name, a.type
		-- Include accounts with either transaction activity OR non-zero balance
		HAVING (COALESCE(SUM(jl.debit_amount), 0) <> 0 OR COALESCE(SUM(jl.credit_amount), 0) <> 0 OR a.balance <> 0)
		ORDER BY a.code`
	if err := s.db.Raw(legacyQuery, asOfDate).Scan(&balances).Error; err != nil {
		return nil, fmt.Errorf("legacy account balances query failed: %v", err)
	}
	
	// If no balances found, try to get all accounts with their direct balances
	if len(balances) == 0 {
		fmt.Printf("[DEBUG] No legacy balances found, trying direct account balances\n")
		allAccountsQuery := `
			SELECT 
				a.id as account_id,
				a.code as account_code,
				a.name as account_name,
				a.type as account_type,
				0 as debit_total,
				0 as credit_total,
				a.balance as net_balance
			FROM accounts a
			WHERE COALESCE(a.is_header, false) = false
			  AND UPPER(a.type) IN ('ASSET', 'LIABILITY', 'EQUITY')
			ORDER BY a.code
		`
		if err := s.db.Raw(allAccountsQuery).Scan(&balances).Error; err != nil {
			return nil, fmt.Errorf("error executing all accounts query: %v", err)
		}
		fmt.Printf("[DEBUG] Retrieved %d accounts with direct balances (legacy fallback)\n", len(balances))
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
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id AND uje.status = 'POSTED' AND uje.deleted_at IS NULL
		WHERE (uje.entry_date <= ? OR uje.entry_date IS NULL)
		AND UPPER(a.type) IN ('REVENUE', 'EXPENSE')
		AND COALESCE(a.is_header, false) = false
	`

	if err := s.db.Raw(query, asOfDate).Scan(&row).Error; err == nil {
		netIncome = row.NetIncome
		fmt.Printf("[DEBUG] Net Income calculated from SSOT: %.2f\n", netIncome)
	} else {
		fmt.Printf("[DEBUG] Failed to get net income from SSOT, falling back to accounts.balance: %v\n", err)
	}

	// If SSOT returned zero and tables might be empty, try legacy journals fallback
	if netIncome == 0 {
		var legacy niRow
		legacyQuery := `
			SELECT 
				COALESCE(SUM(
					CASE 
						WHEN UPPER(a.type) = 'REVENUE' THEN 
							COALESCE(jl.credit_amount, 0) - COALESCE(jl.debit_amount, 0)
						WHEN UPPER(a.type) = 'EXPENSE' THEN 
							COALESCE(jl.debit_amount, 0) - COALESCE(jl.credit_amount, 0)
						ELSE 0
					END
				), 0) as net_income
			FROM accounts a
			LEFT JOIN journal_lines jl ON jl.account_id = a.id
			LEFT JOIN journal_entries je ON je.id = jl.journal_entry_id AND je.status = 'POSTED' AND je.deleted_at IS NULL
			WHERE (je.entry_date <= ? OR je.entry_date IS NULL)
			AND UPPER(a.type) IN ('REVENUE', 'EXPENSE')
			AND COALESCE(a.is_header, false) = false`
		if err := s.db.Raw(legacyQuery, asOfDate).Scan(&legacy).Error; err == nil {
			netIncome = legacy.NetIncome
			fmt.Printf("[DEBUG] Net Income from legacy journals: %.2f\n", netIncome)
		} else {
			// Fallback to accounts table balances as last resort
			var revenue, expense float64
			s.db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE UPPER(type) = 'REVENUE'`).Scan(&revenue)
			s.db.Raw(`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE UPPER(type) = 'EXPENSE'`).Scan(&expense)
			netIncome = revenue - expense
			fmt.Printf("[DEBUG] Net Income from fallback - Revenue: %.2f, Expense: %.2f, Net: %.2f\n", revenue, expense, netIncome)
		}
	}

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
		
		// For liability accounts, ensure the amount is positive for display
		// In accounting, liability accounts normally have credit balances (positive values)
		if strings.ToUpper(balance.AccountType) == "LIABILITY" {
			// Log liability accounts for debugging
			if amount != 0 || strings.Contains(strings.ToUpper(balance.AccountName), "UTANG") {
				fmt.Printf("[DEBUG] Liability Account Found: %s - %s (Type: %s, Balance: %.2f)\n", 
					code, balance.AccountName, balance.AccountType, amount)
			}
			
			// Log specific liability account for debugging
			if code == "2101" {
				fmt.Printf("[DEBUG] Processing UTANG USAHA (2101): Amount=%.2f, Type=%s\n", amount, balance.AccountType)
			}
			
			// Convert negative liability balances to positive for proper display
			// This handles cases where the database might store negative values incorrectly
			if amount < 0 {
				amount = -amount // Make it positive
				fmt.Printf("[DEBUG] Converted negative liability %s (%s) from %.2f to %.2f\n",
					code, balance.AccountName, balance.NetBalance, amount)
			}
		}
		
		item := BSAccountItem{
			AccountCode: balance.AccountCode,
			AccountName: balance.AccountName,
			Amount:      amount,
			AccountID:   balance.AccountID,
		}
		
		// Special handling for PPN accounts
		if strings.Contains(strings.ToLower(item.AccountName), "ppn masukan") {
			fmt.Printf("[DEBUG] Special handling for PPN Masukan: %s - %s (%.2f)\n", code, item.AccountName, amount)
			s.categorizeAssetAccount(bsData, item, code)
		} else if strings.Contains(strings.ToLower(item.AccountName), "ppn keluaran") {
			fmt.Printf("[DEBUG] Special handling for PPN Keluaran: %s - %s (%.2f)\n", code, item.AccountName, amount)
			// PPN Keluaran should be handled separately, not categorized as regular liability
			// We'll add it to a temporary list and handle it in the netting function
			bsData.Liabilities.CurrentLiabilities.Items = append(bsData.Liabilities.CurrentLiabilities.Items, item)
		} else {
			// Convert account type to uppercase for comparison
			accountType := strings.ToUpper(balance.AccountType)
			switch accountType {
			case "ASSET":
				s.categorizeAssetAccount(bsData, item, code)
			case "LIABILITY":
				s.categorizeLiabilityAccount(bsData, item, code)
			case "EQUITY":
				s.categorizeEquityAccount(bsData, item, code)
			default:
				// Log unknown account types
				fmt.Printf("[DEBUG] Unknown account type for %s - %s: %s\n", code, item.AccountName, balance.AccountType)
				// Try to categorize based on code prefixes as fallback
				if strings.HasPrefix(code, "1") {
					s.categorizeAssetAccount(bsData, item, code)
				} else if strings.HasPrefix(code, "2") {
					s.categorizeLiabilityAccount(bsData, item, code)
				} else if strings.HasPrefix(code, "3") {
					s.categorizeEquityAccount(bsData, item, code)
				}
			}
		}
	}
	
	// Net PPN accounts - adjust PPN Keluaran by PPN Masukan amounts to ensure proper netting
	s.netPPNAccounts(bsData)
	
	// Add consolidated PPN line item for better presentation
	s.addNetPPNToBalanceSheet(bsData)
	
	// Remove individual PPN accounts from display to avoid confusion
	s.removePPNAccountsFromDisplay(bsData)
	
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
	
	// Log totals for debugging
	fmt.Printf("[DEBUG] Balance Sheet Totals - Assets: %.2f, Liabilities: %.2f, Equity: %.2f\n",
		bsData.Assets.TotalAssets, bsData.Liabilities.TotalLiabilities, bsData.Equity.TotalEquity)
	
	return bsData
}

// netPPNAccounts adjusts PPN Keluaran by PPN Masukan amounts to ensure proper netting
func (s *SSOTBalanceSheetService) netPPNAccounts(bsData *SSOTBalanceSheetData) {
	// Log items before netting
	fmt.Printf("[DEBUG] Before PPN netting - Liability items:\n")
	for _, item := range bsData.Liabilities.CurrentLiabilities.Items {
		fmt.Printf("[DEBUG]   %s - %s: %.2f\n", item.AccountCode, item.AccountName, item.Amount)
	}
	
	// Calculate total PPN Masukan (from assets)
	var totalPPNMasukan float64
	for _, item := range bsData.Assets.CurrentAssets.Items {
		if strings.Contains(strings.ToLower(item.AccountName), "ppn masukan") {
			totalPPNMasukan += item.Amount
		}
	}
	
	// Calculate total PPN Keluaran (from liabilities items)
	var totalPPNKeluaran float64
	for _, item := range bsData.Liabilities.CurrentLiabilities.Items {
		if strings.Contains(strings.ToLower(item.AccountName), "ppn keluaran") {
			totalPPNKeluaran += item.Amount
		}
	}
	
	// Calculate net PPN liability
	netPPN := totalPPNKeluaran - totalPPNMasukan
	
	// Note: We don't adjust the liability totals here because we handle that in removePPNAccountsFromDisplay
	// This ensures that the calculations are consistent
	
	fmt.Printf("[DEBUG] PPN Netting - Masukan: %.2f, Keluaran: %.2f, Net: %.2f\n",
		totalPPNMasukan, totalPPNKeluaran, netPPN)
}

// addNetPPNToBalanceSheet adds a consolidated PPN line item to show the net PPN liability
func (s *SSOTBalanceSheetService) addNetPPNToBalanceSheet(bsData *SSOTBalanceSheetData) {
	// Calculate total PPN Masukan (from assets)
	var totalPPNMasukan float64
	for _, item := range bsData.Assets.CurrentAssets.Items {
		if strings.Contains(strings.ToLower(item.AccountName), "ppn masukan") {
			totalPPNMasukan += item.Amount
		}
	}
	
	// Calculate total PPN Keluaran (from liabilities)
	var totalPPNKeluaran float64
	for _, item := range bsData.Liabilities.CurrentLiabilities.Items {
		if strings.Contains(strings.ToLower(item.AccountName), "ppn keluaran") {
			totalPPNKeluaran += item.Amount
		}
	}
	
	// Only add net PPN line if there are PPN accounts
	if totalPPNMasukan > 0 || totalPPNKeluaran > 0 {
		netPPN := totalPPNKeluaran - totalPPNMasukan
		ppnItem := BSAccountItem{
			AccountCode: "PPN_NET",
			AccountName: fmt.Sprintf("PPN (Keluaran %.0f - Masukan %.0f)", totalPPNKeluaran, totalPPNMasukan),
			Amount:      netPPN,
		}
		
		// Add to liabilities if positive, or assets if negative
		if netPPN >= 0 {
			bsData.Liabilities.CurrentLiabilities.Items = append(bsData.Liabilities.CurrentLiabilities.Items, ppnItem)
		} else {
			// If net PPN is negative (PPN Masukan > PPN Keluaran), it's a receivable (asset)
			ppnItem.Amount = -netPPN // Make it positive for display
			bsData.Assets.CurrentAssets.Items = append(bsData.Assets.CurrentAssets.Items, ppnItem)
		}
		
		fmt.Printf("[DEBUG] Added Net PPN Item - Keluaran: %.2f, Masukan: %.2f, Net: %.2f\n",
			totalPPNKeluaran, totalPPNMasukan, netPPN)
	}
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
	// Log liability categorization for debugging
	fmt.Printf("[DEBUG] Categorizing Liability: %s - %s (Amount: %.2f, Code: %s)\n", 
		item.AccountCode, item.AccountName, item.Amount, code)
	
	// Ensure liability amounts are positive for proper categorization
	amount := item.Amount
	if amount < 0 {
		amount = -amount // Make it positive
		fmt.Printf("[DEBUG] Converted negative liability amount from %.2f to %.2f\n", item.Amount, amount)
	}
	
	switch {
	// Current Liabilities (21xx)
	case strings.HasPrefix(code, "210"): // Accounts Payable
		bsData.Liabilities.CurrentLiabilities.AccountsPayable += amount
		// Update the item amount to be positive for consistency
		item.Amount = amount
		bsData.Liabilities.CurrentLiabilities.Items = append(bsData.Liabilities.CurrentLiabilities.Items, item)
		fmt.Printf("[DEBUG] Added to AccountsPayable: %.2f\n", amount)
	
	case strings.HasPrefix(code, "211"): // Short-term debt
		bsData.Liabilities.CurrentLiabilities.ShortTermDebt += amount
		// Update the item amount to be positive for consistency
		item.Amount = amount
		bsData.Liabilities.CurrentLiabilities.Items = append(bsData.Liabilities.CurrentLiabilities.Items, item)
		fmt.Printf("[DEBUG] Added to ShortTermDebt: %.2f\n", amount)
	
	case strings.HasPrefix(code, "212"), strings.HasPrefix(code, "213"): // Accrued liabilities and taxes
		if strings.Contains(strings.ToLower(item.AccountName), "tax") || strings.Contains(strings.ToLower(item.AccountName), "pajak") {
			bsData.Liabilities.CurrentLiabilities.TaxPayable += amount
			fmt.Printf("[DEBUG] Added to TaxPayable: %.2f\n", amount)
		} else {
			bsData.Liabilities.CurrentLiabilities.AccruedLiabilities += amount
			fmt.Printf("[DEBUG] Added to AccruedLiabilities: %.2f\n", amount)
		}
		// Update the item amount to be positive for consistency
		item.Amount = amount
		bsData.Liabilities.CurrentLiabilities.Items = append(bsData.Liabilities.CurrentLiabilities.Items, item)
	
	case strings.HasPrefix(code, "21"): // Other current liabilities
		bsData.Liabilities.CurrentLiabilities.OtherCurrentLiabilities += amount
		// Update the item amount to be positive for consistency
		item.Amount = amount
		bsData.Liabilities.CurrentLiabilities.Items = append(bsData.Liabilities.CurrentLiabilities.Items, item)
		fmt.Printf("[DEBUG] Added to OtherCurrentLiabilities: %.2f\n", amount)
	
	// Non-Current Liabilities (22xx, 23xx)
	case strings.HasPrefix(code, "22"): // Long-term debt
		bsData.Liabilities.NonCurrentLiabilities.LongTermDebt += amount
		// Update the item amount to be positive for consistency
		item.Amount = amount
		bsData.Liabilities.NonCurrentLiabilities.Items = append(bsData.Liabilities.NonCurrentLiabilities.Items, item)
		fmt.Printf("[DEBUG] Added to LongTermDebt: %.2f\n", amount)
	
	case strings.HasPrefix(code, "23"): // Other non-current liabilities
		if strings.Contains(strings.ToLower(item.AccountName), "tax") || strings.Contains(strings.ToLower(item.AccountName), "pajak") {
			bsData.Liabilities.NonCurrentLiabilities.DeferredTax += amount
			fmt.Printf("[DEBUG] Added to DeferredTax: %.2f\n", amount)
		} else {
			bsData.Liabilities.NonCurrentLiabilities.OtherNonCurrentLiabilities += amount
			fmt.Printf("[DEBUG] Added to OtherNonCurrentLiabilities: %.2f\n", amount)
		}
		// Update the item amount to be positive for consistency
		item.Amount = amount
		bsData.Liabilities.NonCurrentLiabilities.Items = append(bsData.Liabilities.NonCurrentLiabilities.Items, item)
	
	default: // Other liabilities
		bsData.Liabilities.NonCurrentLiabilities.OtherNonCurrentLiabilities += amount
		// Update the item amount to be positive for consistency
		item.Amount = amount
		bsData.Liabilities.NonCurrentLiabilities.Items = append(bsData.Liabilities.NonCurrentLiabilities.Items, item)
		fmt.Printf("[DEBUG] Added to OtherNonCurrentLiabilities (default): %.2f\n", amount)
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
	
	// Log assets calculation
	fmt.Printf("[DEBUG] Assets Calculation - Current: %.2f, Non-Current: %.2f, Total: %.2f\n",
		bsData.Assets.CurrentAssets.TotalCurrentAssets,
		bsData.Assets.NonCurrentAssets.TotalNonCurrentAssets,
		bsData.Assets.TotalAssets)
	
	// Calculate current liabilities total
	bsData.Liabilities.CurrentLiabilities.TotalCurrentLiabilities = 
		bsData.Liabilities.CurrentLiabilities.AccountsPayable +
		bsData.Liabilities.CurrentLiabilities.ShortTermDebt +
		bsData.Liabilities.CurrentLiabilities.AccruedLiabilities +
		bsData.Liabilities.CurrentLiabilities.TaxPayable +
		bsData.Liabilities.CurrentLiabilities.OtherCurrentLiabilities
	
	// Log liabilities calculation
	fmt.Printf("[DEBUG] Liabilities Calculation - AccountsPayable: %.2f, Other: %.2f, Total: %.2f\n",
		bsData.Liabilities.CurrentLiabilities.AccountsPayable,
		bsData.Liabilities.CurrentLiabilities.ShortTermDebt +
		bsData.Liabilities.CurrentLiabilities.AccruedLiabilities +
		bsData.Liabilities.CurrentLiabilities.TaxPayable +
		bsData.Liabilities.CurrentLiabilities.OtherCurrentLiabilities,
		bsData.Liabilities.CurrentLiabilities.TotalCurrentLiabilities)
	
	// Calculate total liabilities
	bsData.Liabilities.TotalLiabilities = 
		bsData.Liabilities.CurrentLiabilities.TotalCurrentLiabilities +
		bsData.Liabilities.NonCurrentLiabilities.TotalNonCurrentLiabilities
	
	// Log liabilities calculation
	fmt.Printf("[DEBUG] Liabilities Calculation - Current: %.2f, Non-Current: %.2f, Total: %.2f\n",
		bsData.Liabilities.CurrentLiabilities.TotalCurrentLiabilities,
		bsData.Liabilities.NonCurrentLiabilities.TotalNonCurrentLiabilities,
		bsData.Liabilities.TotalLiabilities)
	
	// Calculate total equity
	bsData.Equity.TotalEquity = 
		bsData.Equity.ShareCapital +
		bsData.Equity.RetainedEarnings +
		bsData.Equity.OtherEquity
	
	// Log equity calculation
	fmt.Printf("[DEBUG] Equity Calculation - ShareCapital: %.2f, RetainedEarnings: %.2f, OtherEquity: %.2f, Total: %.2f\n",
		bsData.Equity.ShareCapital,
		bsData.Equity.RetainedEarnings,
		bsData.Equity.OtherEquity,
		bsData.Equity.TotalEquity)
	
	// Calculate total liabilities and equity
	bsData.TotalLiabilitiesAndEquity = bsData.Liabilities.TotalLiabilities + bsData.Equity.TotalEquity
	
	// Check if balance sheet is balanced
	tolerance := 0.01 // 1 cent tolerance
	bsData.BalanceDifference = bsData.Assets.TotalAssets - bsData.TotalLiabilitiesAndEquity
	bsData.IsBalanced = (bsData.BalanceDifference >= -tolerance && bsData.BalanceDifference <= tolerance)
	
	// Log final balance check
	fmt.Printf("[DEBUG] Balance Check - Assets: %.2f, Liabilities+Equity: %.2f, Difference: %.2f, Balanced: %t\n",
		bsData.Assets.TotalAssets,
		bsData.TotalLiabilitiesAndEquity,
		bsData.BalanceDifference,
		bsData.IsBalanced)
}

// calculateAccountBalanceFromTransactions calculates an account's balance from its transaction data
func (s *SSOTBalanceSheetService) calculateAccountBalanceFromTransactions(accountID uint, accountType string) (float64, error) {
	query := `
		SELECT 
			CASE 
				WHEN UPPER(?) IN ('ASSET', 'EXPENSE') THEN 
					COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
				ELSE 
					COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
			END as net_balance
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id AND uje.status = 'POSTED' AND uje.deleted_at IS NULL
		WHERE a.id = ?
		  AND COALESCE(a.is_header, false) = false
		GROUP BY a.type
	`
	
	type result struct {
		NetBalance float64 `gorm:"column:net_balance"`
	}
	
	var row result
	if err := s.db.Raw(query, accountType, accountID).Scan(&row).Error; err != nil {
		return 0, fmt.Errorf("error calculating account balance from transactions: %v", err)
	}
	
	return row.NetBalance, nil
}

// calculateAccountBalanceFromChildren calculates an account's balance from its child accounts
func (s *SSOTBalanceSheetService) calculateAccountBalanceFromChildren(accountID uint, accountType string) (float64, error) {
	// First, get direct children
	query := `
		SELECT COALESCE(SUM(
			CASE 
				WHEN UPPER(?) IN ('ASSET', 'EXPENSE') THEN 
					-- For assets and expenses, debit increases balance
					COALESCE(balance, 0)
				ELSE 
					-- For liabilities, equity, and revenue, credit increases balance
					COALESCE(balance, 0)
			END
		), 0) as total_balance
		FROM accounts 
		WHERE parent_id = ?
		  AND COALESCE(is_header, false) = false
	`
	
	type result struct {
		TotalBalance float64 `gorm:"column:total_balance"`
	}
	
	var row result
	if err := s.db.Raw(query, accountType, accountID).Scan(&row).Error; err != nil {
		return 0, fmt.Errorf("error calculating account balance from children: %v", err)
	}
	
	return row.TotalBalance, nil
}

// removePPNAccountsFromDisplay removes individual PPN accounts from display since we're showing the net amount
func (s *SSOTBalanceSheetService) removePPNAccountsFromDisplay(bsData *SSOTBalanceSheetData) {
	// Log items before removal
	fmt.Printf("[DEBUG] Before PPN removal - Liability items:\n")
	for _, item := range bsData.Liabilities.CurrentLiabilities.Items {
		fmt.Printf("[DEBUG]   %s - %s: %.2f\n", item.AccountCode, item.AccountName, item.Amount)
	}
	
	// Remove PPN Masukan from assets display and adjust totals
	var filteredAssets []BSAccountItem
	var totalPPNMasukanRemoved float64
	for _, item := range bsData.Assets.CurrentAssets.Items {
		if strings.Contains(strings.ToLower(item.AccountName), "ppn masukan") {
			totalPPNMasukanRemoved += item.Amount
		} else {
			filteredAssets = append(filteredAssets, item)
		}
	}
	bsData.Assets.CurrentAssets.Items = filteredAssets
	// Adjust the asset totals to remove the PPN Masukan amounts
	bsData.Assets.CurrentAssets.OtherCurrentAssets -= totalPPNMasukanRemoved
	bsData.Assets.CurrentAssets.TotalCurrentAssets -= totalPPNMasukanRemoved
	bsData.Assets.TotalAssets -= totalPPNMasukanRemoved
	
	fmt.Printf("[DEBUG] Before removing PPN Keluaran - AccountsPayable: %.2f\n", bsData.Liabilities.CurrentLiabilities.AccountsPayable)
	
	// Remove PPN Keluaran from liabilities display and adjust totals
	var filteredLiabilities []BSAccountItem
	var totalPPNKeluaranRemoved float64
	for _, item := range bsData.Liabilities.CurrentLiabilities.Items {
		if strings.Contains(strings.ToLower(item.AccountName), "ppn keluaran") && item.AccountCode != "PPN_NET" {
			totalPPNKeluaranRemoved += item.Amount
		} else if item.AccountCode != "PPN_NET" {
			filteredLiabilities = append(filteredLiabilities, item)
		}
	}
	
	// Add back the PPN_NET item if it exists and calculate its amount
	var ppnNetAmount float64
	for _, item := range bsData.Liabilities.CurrentLiabilities.Items {
		if item.AccountCode == "PPN_NET" {
			ppnNetAmount = item.Amount
			filteredLiabilities = append(filteredLiabilities, item)
		}
	}
	
	bsData.Liabilities.CurrentLiabilities.Items = filteredLiabilities
	// Adjust the liability totals to remove the PPN Keluaran amounts and add the net PPN amount
	bsData.Liabilities.CurrentLiabilities.AccountsPayable -= totalPPNKeluaranRemoved
	bsData.Liabilities.CurrentLiabilities.AccountsPayable += ppnNetAmount // Add net PPN to AccountsPayable
	bsData.Liabilities.CurrentLiabilities.TotalCurrentLiabilities -= totalPPNKeluaranRemoved
	bsData.Liabilities.CurrentLiabilities.TotalCurrentLiabilities += ppnNetAmount // Add net PPN to total
	bsData.Liabilities.TotalLiabilities -= totalPPNKeluaranRemoved
	bsData.Liabilities.TotalLiabilities += ppnNetAmount // Add net PPN to total liabilities
	
	fmt.Printf("[DEBUG] Removed PPN accounts from display - Masukan: %.2f, Keluaran: %.2f, Net PPN added: %.2f\n", totalPPNMasukanRemoved, totalPPNKeluaranRemoved, ppnNetAmount)
	fmt.Printf("[DEBUG] After removing PPN Keluaran - AccountsPayable: %.2f\n", bsData.Liabilities.CurrentLiabilities.AccountsPayable)
	
	// Log items after removal
	fmt.Printf("[DEBUG] After PPN removal - Liability items:\n")
	for _, item := range bsData.Liabilities.CurrentLiabilities.Items {
		fmt.Printf("[DEBUG]   %s - %s: %.2f\n", item.AccountCode, item.AccountName, item.Amount)
	}
}
