package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

type ReportService struct {
	DB              *gorm.DB
	AccountRepo     repositories.AccountRepository
	SalesRepo       *repositories.SalesRepository
	PurchaseRepo    *repositories.PurchaseRepository
	ProductRepo     *repositories.ProductRepository
	ContactRepo     repositories.ContactRepository
	PaymentRepo     *repositories.PaymentRepository
	CashBankRepo    *repositories.CashBankRepository
	EnhancedService *EnhancedReportService
}

func NewReportService(
	db *gorm.DB,
	accountRepo repositories.AccountRepository,
	salesRepo *repositories.SalesRepository,
	purchaseRepo *repositories.PurchaseRepository,
	productRepo *repositories.ProductRepository,
	contactRepo repositories.ContactRepository,
	paymentRepo *repositories.PaymentRepository,
	cashBankRepo *repositories.CashBankRepository,
) *ReportService {
	// Create enhanced service for comprehensive reporting
	enhancedService := NewEnhancedReportService(
		db, accountRepo, salesRepo, purchaseRepo, productRepo, contactRepo, paymentRepo, cashBankRepo,
	)

	return &ReportService{
		DB:              db,
		AccountRepo:     accountRepo,
		SalesRepo:       salesRepo,
		PurchaseRepo:    purchaseRepo,
		ProductRepo:     productRepo,
		ContactRepo:     contactRepo,
		PaymentRepo:     paymentRepo,
		CashBankRepo:    cashBankRepo,
		EnhancedService: enhancedService,
	}
}

func (rs *ReportService) GetAvailableReports() []models.ReportMetadata {
reports := []models.ReportMetadata{
		{
			ReportType:          "balance-sheet",
			Name:                "Balance Sheet",
			Description:         "Statement of financial position showing assets, liabilities, and equity",
			SupportsComparative: false,
			RequiredParams:      []string{"as_of_date", "format"},
			OptionalParams:      []string{"show_zero"},
		},
		{
			ReportType:          "profit-loss",
			Name:                "Profit & Loss Statement",
			Description:         "Statement showing revenues, expenses, and net income for a period",
			SupportsComparative: true,
			RequiredParams:      []string{"start_date", "end_date", "format"},
			OptionalParams:      []string{"comparative", "show_zero"},
		},
		{
			ReportType:          "cash-flow",
			Name:                "Cash Flow Statement",
			Description:         "Statement showing cash receipts and payments for a period",
			SupportsComparative: true,
			RequiredParams:      []string{"start_date", "end_date", "format"},
			OptionalParams:      []string{"comparative", "method"},
		},
		{
			ReportType:          "sales-summary",
			Name:                "Sales Summary Report",
			Description:         "Summary of sales transactions and analytics for a period",
			SupportsComparative: false,
			RequiredParams:      []string{"start_date", "end_date", "group_by", "format"},
			OptionalParams:      []string{"customer_id", "product_id"},
		},
		{
			ReportType:          "purchase-summary",
			Name:                "Purchase Summary Report",
			Description:         "Summary of purchase transactions and analytics for a period",
			SupportsComparative: false,
			RequiredParams:      []string{"start_date", "end_date", "group_by", "format"},
			OptionalParams:      []string{"vendor_id", "product_id"},
		},
		{
			ReportType:          "accounts-receivable",
			Name:                "Accounts Receivable Report",
			Description:         "Detailed aging analysis of customer receivables",
			SupportsComparative: false,
			RequiredParams:      []string{"as_of_date", "format"},
			OptionalParams:      []string{"customer_id"},
		},
		{
			ReportType:          "accounts-payable",
			Name:                "Accounts Payable Report",
			Description:         "Detailed aging analysis of vendor payables",
			SupportsComparative: false,
			RequiredParams:      []string{"as_of_date", "format"},
			OptionalParams:      []string{"vendor_id"},
		},
		{
			ReportType:          "inventory-report",
			Name:                "Inventory Valuation Report",
			Description:         "Inventory quantities, costs, and valuations",
			SupportsComparative: false,
			RequiredParams:      []string{"as_of_date", "format"},
			OptionalParams:      []string{"include_valuation"},
		},
		{
			ReportType:          "trial-balance",
			Name:                "Trial Balance",
			Description:         "Summary of all account balances to verify debits equal credits",
			SupportsComparative: false,
			RequiredParams:      []string{"as_of_date", "format"},
			OptionalParams:      []string{"show_zero"},
		},
		{
			ReportType:          "financial-ratios",
			Name:                "Financial Ratios Analysis",
			Description:         "Key financial ratios and performance indicators",
			SupportsComparative: true,
			RequiredParams:      []string{"as_of_date", "format"},
			OptionalParams:      []string{"period", "comparative"},
		},
	}
	return reports
}

func (rs *ReportService) GenerateBalanceSheet(asOfDate time.Time, format string) (*models.ReportResponse, error) {
	// Use enhanced service for comprehensive balance sheet
	balanceSheetData, err := rs.EnhancedService.GenerateBalanceSheet(asOfDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate balance sheet: %v", err)
	}

	// Create report response
	reportResponse := &models.ReportResponse{
		ID:          0, // Will be set if saved
		Title:       "Balance Sheet",
		Type:        "FINANCIAL",
		Period:      fmt.Sprintf("As of %s", asOfDate.Format("2006-01-02")),
		Format:      format,
		Data:        balanceSheetData,
		GeneratedAt: time.Now(),
		Status:      "success",
		Metadata: map[string]interface{}{
			"total_assets":            balanceSheetData.TotalAssets,
			"total_liabilities":       balanceSheetData.Liabilities.Total,
			"total_equity":            balanceSheetData.Equity.Total,
			"is_balanced":             balanceSheetData.IsBalanced,
			"balance_difference":      balanceSheetData.Difference,
			"parameters": map[string]interface{}{
				"as_of_date": asOfDate,
				"format":     format,
			},
		},
	}

	return reportResponse, nil
}

func (rs *ReportService) GenerateProfitLoss(startDate, endDate time.Time, format string) (*models.ReportResponse, error) {
	// Use enhanced service for comprehensive P&L
	profitLossData, err := rs.EnhancedService.GenerateProfitLoss(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate profit & loss statement: %v", err)
	}

	// Create report response
	reportResponse := &models.ReportResponse{
		ID:          0, // Will be set if saved
		Title:       "Profit & Loss Statement",
		Type:        "FINANCIAL",
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Format:      format,
		Data:        profitLossData,
		GeneratedAt: time.Now(),
		Status:      "success",
		Metadata: map[string]interface{}{
			"total_revenue":     profitLossData.Revenue.Subtotal,
			"total_expenses":    profitLossData.CostOfGoodsSold.Subtotal + profitLossData.OperatingExpenses.Subtotal,
			"gross_profit":      profitLossData.GrossProfit,
			"operating_income":  profitLossData.OperatingIncome,
			"net_income":        profitLossData.NetIncome,
			"net_margin":        profitLossData.NetIncomeMargin,
			"parameters": map[string]interface{}{
				"start_date": startDate,
				"end_date":   endDate,
				"format":     format,
			},
		},
	}

	return reportResponse, nil
}

func (rs *ReportService) GenerateCashFlow(startDate, endDate time.Time, format string) (*models.ReportResponse, error) {
	// Use enhanced service for comprehensive cash flow
	cashFlowData, err := rs.EnhancedService.GenerateCashFlow(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cash flow statement: %v", err)
	}

	// Create report response
	reportResponse := &models.ReportResponse{
		ID:          0, // Will be set if saved
		Title:       "Cash Flow Statement",
		Type:        "FINANCIAL",
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Format:      format,
		Data:        cashFlowData,
		GeneratedAt: time.Now(),
		Status:      "success",
		Metadata: map[string]interface{}{
			"beginning_cash":        cashFlowData.BeginningCash,
			"ending_cash":          cashFlowData.EndingCash,
			"net_cash_flow":        cashFlowData.NetCashFlow,
			"operating_cash_flow":  cashFlowData.OperatingActivities.Total,
			"investing_cash_flow":  cashFlowData.InvestingActivities.Total,
			"financing_cash_flow":  cashFlowData.FinancingActivities.Total,
			"parameters": map[string]interface{}{
				"start_date": startDate,
				"end_date":   endDate,
				"format":     format,
			},
		},
	}

	return reportResponse, nil
}

func (rs *ReportService) GenerateTrialBalance(asOfDate time.Time, format string) (*models.ReportResponse, error) {
	// Get all accounts with their balances
	ctx := context.Background()
	accounts, err := rs.AccountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}

	// Calculate trial balance
	type TrialBalanceEntry struct {
		AccountID   uint    `json:"account_id"`
		Code        string  `json:"code"`
		Name        string  `json:"name"`
		DebitAmount float64 `json:"debit_amount"`
		CreditAmount float64 `json:"credit_amount"`
	}

	var entries []TrialBalanceEntry
	var totalDebits, totalCredits float64

	for _, account := range accounts {
		if !account.IsActive {
			continue
		}

		balance := rs.EnhancedService.calculateAccountBalance(account.ID, asOfDate)
		if balance == 0 {
			continue
		}

		entry := TrialBalanceEntry{
			AccountID: account.ID,
			Code:      account.Code,
			Name:      account.Name,
		}

		// Determine if balance should be in debit or credit column
		switch account.Type {
		case models.AccountTypeAsset, models.AccountTypeExpense:
			if balance > 0 {
				entry.DebitAmount = balance
				totalDebits += balance
			} else {
				entry.CreditAmount = -balance
				totalCredits += -balance
			}
		case models.AccountTypeLiability, models.AccountTypeEquity, models.AccountTypeRevenue:
			if balance > 0 {
				entry.CreditAmount = balance
				totalCredits += balance
			} else {
				entry.DebitAmount = -balance
				totalDebits += -balance
			}
		}

		entries = append(entries, entry)
	}

	// Create trial balance data
	trialBalanceData := map[string]interface{}{
		"as_of_date":     asOfDate,
		"entries":        entries,
		"total_debits":   totalDebits,
		"total_credits":  totalCredits,
		"difference":     totalDebits - totalCredits,
		"is_balanced":    totalDebits == totalCredits,
		"entry_count":    len(entries),
	}

	reportResponse := &models.ReportResponse{
		ID:          0, // Will be set if saved
		Title:       "Trial Balance",
		Type:        "FINANCIAL",
		Period:      fmt.Sprintf("As of %s", asOfDate.Format("2006-01-02")),
		Format:      format,
		Data:        trialBalanceData,
		GeneratedAt: time.Now(),
		Status:      "success",
		Metadata: map[string]interface{}{
			"total_debits":  totalDebits,
			"total_credits": totalCredits,
			"is_balanced":   totalDebits == totalCredits,
			"entry_count":   len(entries),
			"parameters": map[string]interface{}{
				"as_of_date": asOfDate,
				"format":     format,
			},
		},
	}

	return reportResponse, nil
}

func (rs *ReportService) GenerateGeneralLedger(startDate, endDate time.Time, accountCode string, format string) (*models.ReportResponse, error) {
	// Get specific account if code provided, otherwise get all accounts
	ctx := context.Background()
	var accounts []models.Account
	var err error

	if accountCode != "" {
		// Get specific account
		account, findErr := rs.AccountRepo.FindByCode(ctx, accountCode)
		if findErr != nil {
			return nil, fmt.Errorf("account not found: %v", findErr)
		}
		accounts = []models.Account{*account}
	} else {
		// Get all accounts
		accounts, err = rs.AccountRepo.FindAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch accounts: %v", err)
		}
	}

	// Get journal entries for accounts within date range
	type GLEntry struct {
		Date        time.Time `json:"date"`
		JournalID   uint      `json:"journal_id"`
		Description string    `json:"description"`
		Debit       float64   `json:"debit"`
		Credit      float64   `json:"credit"`
		Balance     float64   `json:"balance"`
		Reference   string    `json:"reference"`
	}

	type GLAccount struct {
		AccountID       uint      `json:"account_id"`
		Code            string    `json:"code"`
		Name            string    `json:"name"`
		OpeningBalance  float64   `json:"opening_balance"`
		ClosingBalance  float64   `json:"closing_balance"`
		TotalDebits     float64   `json:"total_debits"`
		TotalCredits    float64   `json:"total_credits"`
		Entries         []GLEntry `json:"entries"`
	}

	var glAccounts []GLAccount

	for _, account := range accounts {
		if !account.IsActive {
			continue
		}

		// Get opening balance
		openingBalance := rs.EnhancedService.calculateAccountBalance(account.ID, startDate.AddDate(0, 0, -1))

		// Get journal entries for this account
		var journalEntries []struct {
			Date          time.Time
			JournalID     uint
			Description   string
			DebitAmount   float64
			CreditAmount  float64
			Reference     string
		}

		rs.DB.Table("journal_entries je").
			Joins("JOIN journals j ON je.journal_id = j.id").
			Where("je.account_id = ? AND j.date BETWEEN ? AND ? AND j.status = ?", 
				account.ID, startDate, endDate, models.JournalStatusPosted).
			Select("j.date, j.id as journal_id, j.description, je.debit_amount, je.credit_amount, j.reference").
			Order("j.date ASC, j.id ASC").
			Scan(&journalEntries)

		// Convert to GL entries and calculate running balance
		var glEntries []GLEntry
		runningBalance := openingBalance
		var totalDebits, totalCredits float64

		for _, entry := range journalEntries {
			// Update running balance based on account type
			if account.Type == models.AccountTypeAsset || account.Type == models.AccountTypeExpense {
				runningBalance += entry.DebitAmount - entry.CreditAmount
			} else {
				runningBalance += entry.CreditAmount - entry.DebitAmount
			}

			totalDebits += entry.DebitAmount
			totalCredits += entry.CreditAmount

			glEntries = append(glEntries, GLEntry{
				Date:        entry.Date,
				JournalID:   entry.JournalID,
				Description: entry.Description,
				Debit:       entry.DebitAmount,
				Credit:      entry.CreditAmount,
				Balance:     runningBalance,
				Reference:   entry.Reference,
			})
		}

		if len(glEntries) > 0 || openingBalance != 0 {
			glAccounts = append(glAccounts, GLAccount{
				AccountID:      account.ID,
				Code:           account.Code,
				Name:           account.Name,
				OpeningBalance: openingBalance,
				ClosingBalance: runningBalance,
				TotalDebits:    totalDebits,
				TotalCredits:   totalCredits,
				Entries:        glEntries,
			})
		}
	}

	generalLedgerData := map[string]interface{}{
		"start_date":     startDate,
		"end_date":       endDate,
		"account_code":   accountCode,
		"accounts":       glAccounts,
		"account_count":  len(glAccounts),
	}

	reportResponse := &models.ReportResponse{
		ID:          0, // Will be set if saved
		Title:       "General Ledger",
		Type:        "FINANCIAL",
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Format:      format,
		Data:        generalLedgerData,
		GeneratedAt: time.Now(),
		Status:      "success",
		Metadata: map[string]interface{}{
			"account_count": len(glAccounts),
			"account_code":  accountCode,
			"parameters": map[string]interface{}{
				"start_date":   startDate,
				"end_date":     endDate,
				"account_code": accountCode,
				"format":       format,
			},
		},
	}

	return reportResponse, nil
}

func (rs *ReportService) GenerateAccountsReceivable(asOfDate time.Time, customerID *uint, format string) (*models.ReportResponse, error) {
	// Get outstanding sales (unpaid or partially paid)
	query := rs.DB.Preload("Customer").Preload("SaleItems").Preload("SalePayments")
	
	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}

	var sales []models.Sale
	if err := query.Where("date <= ? AND outstanding_amount > 0", asOfDate).Find(&sales).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch receivables: %v", err)
	}

	type AREntry struct {
		SaleID          uint      `json:"sale_id"`
		SaleCode        string    `json:"sale_code"`
		CustomerID      uint      `json:"customer_id"`
		CustomerName    string    `json:"customer_name"`
		InvoiceDate     time.Time `json:"invoice_date"`
		DueDate         time.Time `json:"due_date"`
		OriginalAmount  float64   `json:"original_amount"`
		PaidAmount      float64   `json:"paid_amount"`
		Balance         float64   `json:"balance"`
		DaysOverdue     int       `json:"days_overdue"`
		AgingCategory   string    `json:"aging_category"`
	}

	var entries []AREntry
	var totalOriginal, totalPaid, totalBalance float64
	var current, days30, days60, days90, over90 float64

	for _, sale := range sales {
		balance := sale.OutstandingAmount
		if balance <= 0 {
			continue
		}

		// Calculate days overdue
		daysOverdue := int(asOfDate.Sub(sale.DueDate).Hours() / 24)
		if daysOverdue < 0 {
			daysOverdue = 0
		}

		// Determine aging category
		var agingCategory string
		if daysOverdue == 0 {
			agingCategory = "Current"
			current += balance
		} else if daysOverdue <= 30 {
			agingCategory = "1-30 days"
			days30 += balance
		} else if daysOverdue <= 60 {
			agingCategory = "31-60 days"
			days60 += balance
		} else if daysOverdue <= 90 {
			agingCategory = "61-90 days"
			days90 += balance
		} else {
			agingCategory = "Over 90 days"
			over90 += balance
		}

		entries = append(entries, AREntry{
			SaleID:          sale.ID,
			SaleCode:        sale.Code,
			CustomerID:      sale.CustomerID,
			CustomerName:    sale.Customer.Name,
			InvoiceDate:     sale.Date,
			DueDate:         sale.DueDate,
			OriginalAmount:  sale.TotalAmount,
			PaidAmount:      sale.PaidAmount,
			Balance:         balance,
			DaysOverdue:     daysOverdue,
			AgingCategory:   agingCategory,
		})

		totalOriginal += sale.TotalAmount
		totalPaid += sale.PaidAmount
		totalBalance += balance
	}

	// Create AR aging summary
	arData := map[string]interface{}{
		"as_of_date":     asOfDate,
		"customer_id":    customerID,
		"entries":        entries,
		"total_original": totalOriginal,
		"total_paid":     totalPaid,
		"total_balance":  totalBalance,
		"aging_summary": map[string]float64{
			"current":      current,
			"1_30_days":    days30,
			"31_60_days":   days60,
			"61_90_days":   days90,
			"over_90_days": over90,
		},
		"entry_count":    len(entries),
	}

	reportResponse := &models.ReportResponse{
		ID:          0, // Will be set if saved
		Title:       "Accounts Receivable",
		Type:        "FINANCIAL",
		Period:      fmt.Sprintf("As of %s", asOfDate.Format("2006-01-02")),
		Format:      format,
		Data:        arData,
		GeneratedAt: time.Now(),
		Status:      "success",
		Metadata: map[string]interface{}{
			"total_balance": totalBalance,
			"entry_count":   len(entries),
			"overdue_amount": days30 + days60 + days90 + over90,
			"parameters": map[string]interface{}{
				"as_of_date":  asOfDate,
				"customer_id": customerID,
				"format":      format,
			},
		},
	}

	return reportResponse, nil
}

func (rs *ReportService) GenerateAccountsPayable(asOfDate time.Time, vendorID *uint, format string) (*models.ReportResponse, error) {
	// Get outstanding purchases (unpaid or partially paid)
	query := rs.DB.Preload("Vendor")
	
	if vendorID != nil {
		query = query.Where("vendor_id = ?", *vendorID)
	}

	var purchases []models.Purchase
	if err := query.Where("date <= ? AND outstanding_amount > 0", asOfDate).Find(&purchases).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch payables: %v", err)
	}

	type APEntry struct {
		PurchaseID      uint      `json:"purchase_id"`
		PurchaseCode    string    `json:"purchase_code"`
		VendorID        uint      `json:"vendor_id"`
		VendorName      string    `json:"vendor_name"`
		InvoiceDate     time.Time `json:"invoice_date"`
		DueDate         time.Time `json:"due_date"`
		OriginalAmount  float64   `json:"original_amount"`
		PaidAmount      float64   `json:"paid_amount"`
		Balance         float64   `json:"balance"`
		DaysOverdue     int       `json:"days_overdue"`
		AgingCategory   string    `json:"aging_category"`
	}

	var entries []APEntry
	var totalOriginal, totalPaid, totalBalance float64
	var current, days30, days60, days90, over90 float64

	for _, purchase := range purchases {
		balance := purchase.OutstandingAmount
		if balance <= 0 {
			continue
		}

		// Calculate days overdue
		daysOverdue := int(asOfDate.Sub(purchase.DueDate).Hours() / 24)
		if daysOverdue < 0 {
			daysOverdue = 0
		}

		// Determine aging category
		var agingCategory string
		if daysOverdue == 0 {
			agingCategory = "Current"
			current += balance
		} else if daysOverdue <= 30 {
			agingCategory = "1-30 days"
			days30 += balance
		} else if daysOverdue <= 60 {
			agingCategory = "31-60 days"
			days60 += balance
		} else if daysOverdue <= 90 {
			agingCategory = "61-90 days"
			days90 += balance
		} else {
			agingCategory = "Over 90 days"
			over90 += balance
		}

		entries = append(entries, APEntry{
			PurchaseID:      purchase.ID,
			PurchaseCode:    purchase.Code,
			VendorID:        purchase.VendorID,
			VendorName:      purchase.Vendor.Name,
			InvoiceDate:     purchase.Date,
			DueDate:         purchase.DueDate,
			OriginalAmount:  purchase.TotalAmount,
			PaidAmount:      purchase.PaidAmount,
			Balance:         balance,
			DaysOverdue:     daysOverdue,
			AgingCategory:   agingCategory,
		})

		totalOriginal += purchase.TotalAmount
		totalPaid += purchase.PaidAmount
		totalBalance += balance
	}

	// Create AP aging summary
	apData := map[string]interface{}{
		"as_of_date":     asOfDate,
		"vendor_id":      vendorID,
		"entries":        entries,
		"total_original": totalOriginal,
		"total_paid":     totalPaid,
		"total_balance":  totalBalance,
		"aging_summary": map[string]float64{
			"current":      current,
			"1_30_days":    days30,
			"31_60_days":   days60,
			"61_90_days":   days90,
			"over_90_days": over90,
		},
		"entry_count":    len(entries),
	}

	reportResponse := &models.ReportResponse{
		ID:          0, // Will be set if saved
		Title:       "Accounts Payable",
		Type:        "FINANCIAL",
		Period:      fmt.Sprintf("As of %s", asOfDate.Format("2006-01-02")),
		Format:      format,
		Data:        apData,
		GeneratedAt: time.Now(),
		Status:      "success",
		Metadata: map[string]interface{}{
			"total_balance": totalBalance,
			"entry_count":   len(entries),
			"overdue_amount": days30 + days60 + days90 + over90,
			"parameters": map[string]interface{}{
				"as_of_date": asOfDate,
				"vendor_id":  vendorID,
				"format":     format,
			},
		},
	}

	return reportResponse, nil
}

func (rs *ReportService) GenerateSalesSummary(startDate, endDate time.Time, groupBy string, format string) (*models.ReportResponse, error) {
	// Use enhanced service for comprehensive sales summary
	salesData, err := rs.EnhancedService.GenerateSalesSummary(startDate, endDate, groupBy)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sales summary: %v", err)
	}

	reportResponse := &models.ReportResponse{
		ID:          0, // Will be set if saved
		Title:       "Sales Summary Report",
		Type:        "OPERATIONAL",
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Format:      format,
		Data:        salesData,
		GeneratedAt: time.Now(),
		Status:      "success",
		Metadata: map[string]interface{}{
			"total_revenue":      salesData.TotalRevenue,
			"total_transactions": salesData.TotalTransactions,
			"average_order_value": salesData.AverageOrderValue,
			"total_customers":    salesData.TotalCustomers,
			"new_customers":      salesData.NewCustomers,
			"parameters": map[string]interface{}{
				"start_date": startDate,
				"end_date":   endDate,
				"group_by":   groupBy,
				"format":     format,
			},
		},
	}

	return reportResponse, nil
}

func (rs *ReportService) GeneratePurchaseSummary(startDate, endDate time.Time, groupBy string, format string) (*models.ReportResponse, error) {
	// Use enhanced service for comprehensive purchase summary
	purchaseData, err := rs.EnhancedService.GeneratePurchaseSummary(startDate, endDate, groupBy)
	if err != nil {
		return nil, fmt.Errorf("failed to generate purchase summary: %v", err)
	}

	reportResponse := &models.ReportResponse{
		ID:          0, // Will be set if saved
		Title:       "Purchase Summary Report",
		Type:        "OPERATIONAL",
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Format:      format,
		Data:        purchaseData,
		GeneratedAt: time.Now(),
		Status:      "success",
		Metadata: map[string]interface{}{
			"total_purchases":        purchaseData.TotalPurchases,
			"total_transactions":     purchaseData.TotalTransactions,
			"average_purchase_value": purchaseData.AveragePurchaseValue,
			"total_vendors":          purchaseData.TotalVendors,
			"new_vendors":            purchaseData.NewVendors,
			"parameters": map[string]interface{}{
				"start_date": startDate,
				"end_date":   endDate,
				"group_by":   groupBy,
				"format":     format,
			},
		},
	}

	return reportResponse, nil
}

func (rs *ReportService) GenerateVendorAnalysis(startDate, endDate time.Time, groupBy string, format string) (*models.ReportResponse, error) {
	// Use enhanced service for comprehensive vendor analysis based on purchase data
	purchaseData, err := rs.EnhancedService.GeneratePurchaseSummary(startDate, endDate, groupBy)
	if err != nil {
		return nil, fmt.Errorf("failed to generate vendor analysis: %v", err)
	}

	// Get vendor performance metrics
	var vendors []models.Contact
	if err := rs.DB.Where("type = 'vendor' AND is_active = true").Find(&vendors).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch vendors: %v", err)
	}

	// Get purchase transactions grouped by vendor for the period
	var purchases []models.Purchase
	if err := rs.DB.Preload("Vendor").Where("date BETWEEN ? AND ?", startDate, endDate).Find(&purchases).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch purchases: %v", err)
	}

	// Calculate vendor metrics
	vendorMetrics := make(map[uint]VendorMetric)
	for _, purchase := range purchases {
		metric := vendorMetrics[purchase.VendorID]
		metric.VendorID = purchase.VendorID
		metric.VendorName = purchase.Vendor.Name
		metric.TotalPurchases += purchase.TotalAmount
		metric.TransactionCount++
		metric.OutstandingAmount += purchase.OutstandingAmount
		vendorMetrics[purchase.VendorID] = metric
	}

	// Convert to slice and calculate average values
	var vendorList []VendorMetric
	totalVendorValue := 0.0
	for _, metric := range vendorMetrics {
		if metric.TransactionCount > 0 {
			metric.AveragePurchaseValue = metric.TotalPurchases / float64(metric.TransactionCount)
		}
		vendorList = append(vendorList, metric)
		totalVendorValue += metric.TotalPurchases
	}

	// Create vendor analysis data compatible with frontend expectations
	vendorAnalysisData := map[string]interface{}{
		"company": map[string]interface{}{
			"name": "PT. Example Company", // This should come from company settings
		},
		"start_date": startDate,
		"end_date": endDate,
		"group_by": groupBy,
		"purchases_by_period": purchaseData.PurchasesByPeriod,
		"total_purchases": purchaseData.TotalPurchases,
		"vendor_metrics": vendorList,
		"vendor_count": len(vendorList),
		"average_vendor_value": func() float64 {
			if len(vendorList) > 0 {
				return totalVendorValue / float64(len(vendorList))
			}
			return 0
		}(),
	}

	reportResponse := &models.ReportResponse{
		ID:          0, // Will be set if saved
		Title:       "Vendor Analysis Report",
		Type:        "OPERATIONAL",
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Format:      format,
		Data:        vendorAnalysisData,
		GeneratedAt: time.Now(),
		Status:      "success",
		Metadata: map[string]interface{}{
			"total_purchases":        purchaseData.TotalPurchases,
			"total_transactions":     purchaseData.TotalTransactions,
			"average_purchase_value": purchaseData.AveragePurchaseValue,
			"total_vendors":          purchaseData.TotalVendors,
			"vendor_count":           len(vendorList),
			"parameters": map[string]interface{}{
				"start_date": startDate,
				"end_date":   endDate,
				"group_by":   groupBy,
				"format":     format,
			},
		},
	}

	return reportResponse, nil
}

// VendorMetric represents vendor performance metrics
type VendorMetric struct {
	VendorID              uint    `json:"vendor_id"`
	VendorName            string  `json:"vendor_name"`
	TotalPurchases        float64 `json:"total_purchases"`
	TransactionCount      int     `json:"transaction_count"`
	AveragePurchaseValue  float64 `json:"average_purchase_value"`
	OutstandingAmount     float64 `json:"outstanding_amount"`
}

func (rs *ReportService) GenerateInventoryReport(asOfDate time.Time, includeValuation bool, format string) (*models.ReportResponse, error) {
	// Get all active products with their current stock levels
	ctx := context.Background()
	products, err := rs.ProductRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: %v", err)
	}


	var entries []InventoryEntry
	var totalQuantity int
	var totalValue float64
	var lowStockCount, outOfStockCount int

	for _, product := range products {
		if !product.IsActive {
			continue
		}

		// Determine status based on stock levels
		status := "Normal"
		if product.Stock == 0 {
			status = "Out of Stock"
			outOfStockCount++
		} else if product.MinStock > 0 && product.Stock <= product.MinStock {
			status = "Low Stock"
			lowStockCount++
		}

		// Calculate total value if valuation is included
		var totalItemValue float64
		if includeValuation {
			totalItemValue = product.CostPrice * float64(product.Stock)
			totalValue += totalItemValue
		}

		// Get last stock movement date (simplified - would need stock movement tracking)
		var lastMovement *time.Time
		// This would require a stock movements table to track properly

		entry := InventoryEntry{
			ProductID:       product.ID,
			Code:            product.Code,
			Name:            product.Name,
			Category:        product.Category.Name,
			Unit:            product.Unit,
			CurrentStock:    product.Stock,
			MinimumStock:    product.MinStock,
			CostPerUnit:     product.CostPrice,
			TotalValue:      totalItemValue,
			Status:          status,
			LastMovement:    lastMovement,
		}

		entries = append(entries, entry)
		totalQuantity += product.Stock
	}

	// Create inventory summary
	inventoryData := map[string]interface{}{
		"as_of_date":         asOfDate,
		"include_valuation":  includeValuation,
		"entries":            entries,
		"total_items":        len(entries),
		"total_quantity":     totalQuantity,
		"total_value":        totalValue,
		"low_stock_count":    lowStockCount,
		"out_of_stock_count": outOfStockCount,
		"summary_by_category": rs.calculateInventoryByCategory(entries),
		"stock_status_summary": map[string]interface{}{
			"normal":       len(entries) - lowStockCount - outOfStockCount,
			"low_stock":    lowStockCount,
			"out_of_stock": outOfStockCount,
		},
	}

	reportResponse := &models.ReportResponse{
		ID:          0, // Will be set if saved
		Title:       "Inventory Valuation Report",
		Type:        "OPERATIONAL",
		Period:      fmt.Sprintf("As of %s", asOfDate.Format("2006-01-02")),
		Format:      format,
		Data:        inventoryData,
		GeneratedAt: time.Now(),
		Status:      "success",
		Metadata: map[string]interface{}{
			"total_items":        len(entries),
			"total_quantity":     totalQuantity,
			"total_value":        totalValue,
			"low_stock_count":    lowStockCount,
			"out_of_stock_count": outOfStockCount,
			"parameters": map[string]interface{}{
				"as_of_date":        asOfDate,
				"include_valuation": includeValuation,
				"format":            format,
			},
		},
	}

	return reportResponse, nil
}

func (rs *ReportService) GenerateFinancialRatios(asOfDate time.Time, period string, format string) (*models.ReportResponse, error) {
	// Generate balance sheet for ratio calculations
	balanceSheet, err := rs.EnhancedService.GenerateBalanceSheet(asOfDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate balance sheet for ratios: %v", err)
	}

	// Generate P&L for the period (assume current year)
	startOfYear := time.Date(asOfDate.Year(), 1, 1, 0, 0, 0, 0, asOfDate.Location())
	profitLoss, err := rs.EnhancedService.GenerateProfitLoss(startOfYear, asOfDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate P&L for ratios: %v", err)
	}

	// Calculate financial ratios
	ratios := rs.calculateFinancialRatios(balanceSheet, profitLoss)

	// Create ratios data
	ratiosData := map[string]interface{}{
		"as_of_date":       asOfDate,
		"period":           period,
		"liquidity_ratios": ratios["liquidity"],
		"efficiency_ratios": ratios["efficiency"],
		"leverage_ratios":   ratios["leverage"],
		"profitability_ratios": ratios["profitability"],
		"market_ratios":     ratios["market"],
		"balance_sheet_data": map[string]interface{}{
			"total_assets":      balanceSheet.TotalAssets,
			"total_liabilities": balanceSheet.Liabilities.Total,
			"total_equity":      balanceSheet.Equity.Total,
		},
		"income_statement_data": map[string]interface{}{
			"total_revenue": profitLoss.Revenue.Subtotal,
			"net_income":    profitLoss.NetIncome,
			"gross_profit":  profitLoss.GrossProfit,
		},
	}

	reportResponse := &models.ReportResponse{
		ID:          0, // Will be set if saved
		Title:       "Financial Ratios Analysis",
		Type:        "ANALYTICAL",
		Period:      fmt.Sprintf("As of %s", asOfDate.Format("2006-01-02")),
		Format:      format,
		Data:        ratiosData,
		GeneratedAt: time.Now(),
		Status:      "success",
		Metadata: map[string]interface{}{
			"current_ratio":  ratios["liquidity"].(map[string]float64)["current_ratio"],
			"debt_to_equity": ratios["leverage"].(map[string]float64)["debt_to_equity"],
			"roa":            ratios["profitability"].(map[string]float64)["return_on_assets"],
			"roe":            ratios["profitability"].(map[string]float64)["return_on_equity"],
			"parameters": map[string]interface{}{
				"as_of_date": asOfDate,
				"period":     period,
				"format":     format,
			},
		},
	}

	return reportResponse, nil
}

func (rs *ReportService) SaveReportTemplate(request *models.ReportTemplateRequest, userID uint) (*models.ReportTemplate, error) {
	// Implementation for saving custom report templates
	template := &models.ReportTemplate{
		Name:        request.Name,
		Type:        request.Type,
		Description: request.Description,
		Template:    request.Template,
		IsDefault:   request.IsDefault,
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := rs.DB.Create(template).Error; err != nil {
		return nil, fmt.Errorf("failed to save report template: %v", err)
	}

	return template, nil
}

func (rs *ReportService) GetReportTemplates() ([]models.ReportTemplate, error) {
	var templates []models.ReportTemplate
	if err := rs.DB.Where("is_active = ?", true).Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch report templates: %v", err)
	}
	return templates, nil
}

// Helper methods

// InventoryEntry represents an inventory item for reporting purposes
type InventoryEntry struct {
	ProductID       uint    `json:"product_id"`
	Code            string  `json:"code"`
	Name            string  `json:"name"`
	Category        string  `json:"category"`
	Unit            string  `json:"unit"`
	CurrentStock    int     `json:"current_stock"`
	MinimumStock    int     `json:"minimum_stock"`
	CostPerUnit     float64 `json:"cost_per_unit"`
	TotalValue      float64 `json:"total_value"`
	Status          string  `json:"status"`
	LastMovement    *time.Time `json:"last_movement"`
}

func (rs *ReportService) calculateInventoryByCategory(entries []InventoryEntry) map[string]interface{} {
	categoryMap := make(map[string]struct {
		Count    int     `json:"count"`
		Quantity int     `json:"quantity"`
		Value    float64 `json:"value"`
	})

	for _, entry := range entries {
		cat := entry.Category
		if cat == "" {
			cat = "Uncategorized"
		}

		data := categoryMap[cat]
		data.Count++
		data.Quantity += entry.CurrentStock
		data.Value += entry.TotalValue
		categoryMap[cat] = data
	}

	result := make(map[string]interface{})
	for category, data := range categoryMap {
		result[category] = data
	}

	return result
}

func (rs *ReportService) calculateFinancialRatios(balanceSheet *BalanceSheetData, profitLoss *ProfitLossData) map[string]interface{} {
	// Get specific balance sheet components
	currentAssets := rs.getSubtotalByCategory(balanceSheet.Assets.Subtotals, "CURRENT_ASSET")
	currentLiabilities := rs.getSubtotalByCategory(balanceSheet.Liabilities.Subtotals, "CURRENT_LIABILITY")

	// Liquidity Ratios
	liquidityRatios := map[string]float64{
		"current_ratio": rs.safeDivide(currentAssets, currentLiabilities),
		"quick_ratio":   rs.safeDivide(currentAssets*0.8, currentLiabilities), // Simplified quick ratio
		"cash_ratio":    rs.safeDivide(currentAssets*0.3, currentLiabilities), // Simplified cash ratio
	}

	// Efficiency Ratios
	efficiencyRatios := map[string]float64{
		"asset_turnover":        rs.safeDivide(profitLoss.Revenue.Subtotal, balanceSheet.TotalAssets),
		"inventory_turnover":    rs.safeDivide(profitLoss.CostOfGoodsSold.Subtotal, currentAssets*0.4), // Simplified
		"receivables_turnover":  rs.safeDivide(profitLoss.Revenue.Subtotal, currentAssets*0.3),         // Simplified
	}

	// Leverage Ratios
	totalLiabilities := balanceSheet.Liabilities.Total
	leverageRatios := map[string]float64{
		"debt_to_equity":  rs.safeDivide(totalLiabilities, balanceSheet.Equity.Total),
		"debt_to_assets":  rs.safeDivide(totalLiabilities, balanceSheet.TotalAssets),
		"equity_ratio":    rs.safeDivide(balanceSheet.Equity.Total, balanceSheet.TotalAssets),
		"times_interest": rs.safeDivide(profitLoss.EBIT, profitLoss.OtherExpenses.Subtotal*0.3), // Simplified interest expense
	}

	// Profitability Ratios
	profitabilityRatios := map[string]float64{
		"gross_margin":       rs.safeDivide(profitLoss.GrossProfit, profitLoss.Revenue.Subtotal) * 100,
		"operating_margin":   rs.safeDivide(profitLoss.OperatingIncome, profitLoss.Revenue.Subtotal) * 100,
		"net_margin":         profitLoss.NetIncomeMargin,
		"return_on_assets":   rs.safeDivide(profitLoss.NetIncome, balanceSheet.TotalAssets) * 100,
		"return_on_equity":   rs.safeDivide(profitLoss.NetIncome, balanceSheet.Equity.Total) * 100,
	}

	// Market Ratios (simplified - would need stock data)
	marketRatios := map[string]float64{
		"earnings_per_share": profitLoss.EarningsPerShare,
		"book_value_per_share": rs.safeDivide(balanceSheet.Equity.Total, profitLoss.SharesOutstanding),
	}

	return map[string]interface{}{
		"liquidity":     liquidityRatios,
		"efficiency":    efficiencyRatios,
		"leverage":      leverageRatios,
		"profitability": profitabilityRatios,
		"market":        marketRatios,
	}
}

func (rs *ReportService) getSubtotalByCategory(subtotals []BalanceSheetSubtotal, category string) float64 {
	for _, subtotal := range subtotals {
		if subtotal.Category == category {
			return subtotal.Amount
		}
	}
	return 0
}

func (rs *ReportService) safeDivide(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

// ========== VALIDATION AND ERROR HANDLING FUNCTIONS ==========

// ValidateReportParameters validates common report parameters
func (rs *ReportService) ValidateReportParameters(reportType string, params map[string]interface{}) error {
	switch reportType {
	case "balance-sheet":
		return rs.validateBalanceSheetParams(params)
	case "profit-loss", "cash-flow":
		return rs.validatePeriodParams(params)
	case "sales-summary", "purchase-summary":
		return rs.validateSummaryParams(params)
	case "accounts-receivable", "accounts-payable":
		return rs.validateAgingParams(params)
	case "inventory-report":
		return rs.validateInventoryParams(params)
	case "financial-ratios":
		return rs.validateRatiosParams(params)
	default:
		return fmt.Errorf("unsupported report type: %s", reportType)
	}
}

// validateBalanceSheetParams validates balance sheet parameters
func (rs *ReportService) validateBalanceSheetParams(params map[string]interface{}) error {
	asOfDate, exists := params["as_of_date"]
	if !exists {
		return errors.New("as_of_date is required for balance sheet")
	}

	date, ok := asOfDate.(time.Time)
	if !ok {
		return errors.New("as_of_date must be a valid date")
	}

	if date.After(time.Now()) {
		return errors.New("as_of_date cannot be in the future")
	}

	// Check if date is too far in the past (configurable)
	minDate := time.Now().AddDate(-10, 0, 0) // 10 years ago
	if date.Before(minDate) {
		return errors.New("as_of_date cannot be more than 10 years ago")
	}

	return nil
}

// validatePeriodParams validates period-based report parameters
func (rs *ReportService) validatePeriodParams(params map[string]interface{}) error {
	startDate, hasStart := params["start_date"]
	endDate, hasEnd := params["end_date"]

	if !hasStart || !hasEnd {
		return errors.New("both start_date and end_date are required")
	}

	startTime, okStart := startDate.(time.Time)
	endTime, okEnd := endDate.(time.Time)

	if !okStart || !okEnd {
		return errors.New("start_date and end_date must be valid dates")
	}

	if startTime.After(endTime) {
		return errors.New("start_date must be before or equal to end_date")
	}

	if endTime.After(time.Now()) {
		return errors.New("end_date cannot be in the future")
	}

	// Check if period is too long (configurable)
	maxPeriod := 365 * 5 // 5 years
	if endTime.Sub(startTime).Hours() > float64(maxPeriod*24) {
		return errors.New("report period cannot exceed 5 years")
	}

	// Check if dates are too far in the past
	minDate := time.Now().AddDate(-10, 0, 0)
	if startTime.Before(minDate) {
		return errors.New("start_date cannot be more than 10 years ago")
	}

	return nil
}

// validateSummaryParams validates sales/purchase summary parameters
func (rs *ReportService) validateSummaryParams(params map[string]interface{}) error {
	if err := rs.validatePeriodParams(params); err != nil {
		return err
	}

	groupBy, exists := params["group_by"]
	if exists {
		groupByStr, ok := groupBy.(string)
		if !ok {
			return errors.New("group_by must be a string")
		}

		validGroupBy := []string{"day", "week", "month", "quarter", "year"}
		valid := false
		for _, validValue := range validGroupBy {
			if groupByStr == validValue {
				valid = true
				break
			}
		}

		if !valid {
			return fmt.Errorf("group_by must be one of: %s", strings.Join(validGroupBy, ", "))
		}
	}

	return nil
}

// validateAgingParams validates aging report parameters
func (rs *ReportService) validateAgingParams(params map[string]interface{}) error {
	if err := rs.validateBalanceSheetParams(params); err != nil {
		return err
	}

	// Validate customer_id if provided
	if customerID, exists := params["customer_id"]; exists && customerID != nil {
		if _, ok := customerID.(uint); !ok {
			return errors.New("customer_id must be a valid integer")
		}
	}

	// Validate vendor_id if provided
	if vendorID, exists := params["vendor_id"]; exists && vendorID != nil {
		if _, ok := vendorID.(uint); !ok {
			return errors.New("vendor_id must be a valid integer")
		}
	}

	return nil
}

// validateInventoryParams validates inventory report parameters
func (rs *ReportService) validateInventoryParams(params map[string]interface{}) error {
	if err := rs.validateBalanceSheetParams(params); err != nil {
		return err
	}

	includeVal, exists := params["include_valuation"]
	if exists {
		if _, ok := includeVal.(bool); !ok {
			return errors.New("include_valuation must be a boolean")
		}
	}

	return nil
}

// validateRatiosParams validates financial ratios parameters
func (rs *ReportService) validateRatiosParams(params map[string]interface{}) error {
	if err := rs.validateBalanceSheetParams(params); err != nil {
		return err
	}

	period, exists := params["period"]
	if exists {
		periodStr, ok := period.(string)
		if !ok {
			return errors.New("period must be a string")
		}

		validPeriods := []string{"current", "ytd", "comparative"}
		valid := false
		for _, validValue := range validPeriods {
			if periodStr == validValue {
				valid = true
				break
			}
		}

		if !valid {
			return fmt.Errorf("period must be one of: %s", strings.Join(validPeriods, ", "))
		}
	}

	return nil
}

// ValidateFormat validates report format parameter
func (rs *ReportService) ValidateFormat(format string) error {
	if format == "" {
		return nil // Default format will be used
	}

	validFormats := []string{"json", "pdf", "csv", "xlsx"}
	for _, validFormat := range validFormats {
		if format == validFormat {
			return nil
		}
	}

	return fmt.Errorf("format must be one of: %s", strings.Join(validFormats, ", "))
}

// ValidateDataIntegrity validates the integrity of report data
func (rs *ReportService) ValidateDataIntegrity(reportType string, data interface{}) error {
	switch reportType {
	case "balance-sheet":
		return rs.validateBalanceSheetData(data)
	case "profit-loss":
		return rs.validateProfitLossData(data)
	case "cash-flow":
		return rs.validateCashFlowData(data)
	default:
		return nil // Generic validation passed
	}
}

// validateBalanceSheetData validates balance sheet data integrity
func (rs *ReportService) validateBalanceSheetData(data interface{}) error {
	balanceSheet, ok := data.(*BalanceSheetData)
	if !ok {
		return errors.New("invalid balance sheet data structure")
	}

	if balanceSheet == nil {
		return errors.New("balance sheet data is nil")
	}

	// Check basic required fields
	if balanceSheet.Company.Name == "" {
		return errors.New("company name is required in balance sheet")
	}

	if balanceSheet.Currency == "" {
		return errors.New("currency is required in balance sheet")
	}

	// Validate balance sheet equation: Assets = Liabilities + Equity
	if !balanceSheet.IsBalanced {
		// Allow for small rounding differences
		if balanceSheet.Difference > 0.01 || balanceSheet.Difference < -0.01 {
			return fmt.Errorf("balance sheet is not balanced: difference of %.2f", balanceSheet.Difference)
		}
	}

	// Validate that totals are reasonable
	if balanceSheet.TotalAssets < 0 {
		return errors.New("total assets cannot be negative")
	}

	if balanceSheet.Assets.Total != balanceSheet.TotalAssets {
		return errors.New("assets section total does not match total assets")
	}

	return nil
}

// validateProfitLossData validates P&L data integrity
func (rs *ReportService) validateProfitLossData(data interface{}) error {
	profitLoss, ok := data.(*ProfitLossData)
	if !ok {
		return errors.New("invalid profit & loss data structure")
	}

	if profitLoss == nil {
		return errors.New("profit & loss data is nil")
	}

	// Check basic required fields
	if profitLoss.Company.Name == "" {
		return errors.New("company name is required in P&L")
	}

	if profitLoss.Currency == "" {
		return errors.New("currency is required in P&L")
	}

	// Validate date range
	if profitLoss.EndDate.Before(profitLoss.StartDate) {
		return errors.New("end date must be after start date in P&L")
	}

	// Validate calculations
	expectedGrossProfit := profitLoss.Revenue.Subtotal - profitLoss.CostOfGoodsSold.Subtotal
	if math.Abs(profitLoss.GrossProfit-expectedGrossProfit) > 0.01 {
		return fmt.Errorf("gross profit calculation error: expected %.2f, got %.2f", expectedGrossProfit, profitLoss.GrossProfit)
	}

	// Validate profit margins
	if profitLoss.Revenue.Subtotal > 0 {
		expectedGrossMargin := (profitLoss.GrossProfit / profitLoss.Revenue.Subtotal) * 100
		if math.Abs(profitLoss.GrossProfitMargin-expectedGrossMargin) > 0.01 {
			return fmt.Errorf("gross profit margin calculation error: expected %.2f%%, got %.2f%%", expectedGrossMargin, profitLoss.GrossProfitMargin)
		}
	}

	return nil
}

// validateCashFlowData validates cash flow data integrity
func (rs *ReportService) validateCashFlowData(data interface{}) error {
	cashFlow, ok := data.(*CashFlowData)
	if !ok {
		return errors.New("invalid cash flow data structure")
	}

	if cashFlow == nil {
		return errors.New("cash flow data is nil")
	}

	// Check basic required fields
	if cashFlow.Company.Name == "" {
		return errors.New("company name is required in cash flow")
	}

	if cashFlow.Currency == "" {
		return errors.New("currency is required in cash flow")
	}

	// Validate date range
	if cashFlow.EndDate.Before(cashFlow.StartDate) {
		return errors.New("end date must be after start date in cash flow")
	}

	// Validate cash flow equation: Beginning + Net Cash Flow = Ending
	expectedEnding := cashFlow.BeginningCash + cashFlow.NetCashFlow
	if math.Abs(cashFlow.EndingCash-expectedEnding) > 0.01 {
		return fmt.Errorf("cash flow balance error: expected ending cash %.2f, got %.2f", expectedEnding, cashFlow.EndingCash)
	}

	// Validate net cash flow calculation
	expectedNetCashFlow := cashFlow.OperatingActivities.Total + cashFlow.InvestingActivities.Total + cashFlow.FinancingActivities.Total
	if math.Abs(cashFlow.NetCashFlow-expectedNetCashFlow) > 0.01 {
		return fmt.Errorf("net cash flow calculation error: expected %.2f, got %.2f", expectedNetCashFlow, cashFlow.NetCashFlow)
	}

	return nil
}

// CheckPermissions validates user permissions for report access
func (rs *ReportService) CheckPermissions(userID uint, reportType string) error {
	// This would integrate with your authentication/authorization system
	// For now, we'll implement basic role-based checks
	
	var user models.User
	if err := rs.DB.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %v", err)
	}

	// Define permission requirements for each report type
	permissionMap := map[string][]string{
		"balance-sheet":       {"admin", "finance", "director"},
		"profit-loss":         {"admin", "finance", "director"},
		"cash-flow":           {"admin", "finance", "director"},
		"trial-balance":       {"admin", "finance"},
		"general-ledger":      {"admin", "finance"},
		"accounts-receivable": {"admin", "finance", "sales"},
		"accounts-payable":    {"admin", "finance", "purchasing"},
		"sales-summary":       {"admin", "finance", "sales", "director"},
		"purchase-summary":    {"admin", "finance", "purchasing", "director"},
		"inventory-report":    {"admin", "finance", "inventory", "director"},
		"financial-ratios":    {"admin", "finance", "director"},
	}

	requiredRoles, exists := permissionMap[reportType]
	if !exists {
		return fmt.Errorf("unknown report type for permission check: %s", reportType)
	}

	// Check if user has required role
	userRole := strings.ToLower(user.Role)
	for _, role := range requiredRoles {
		if userRole == role {
			return nil
		}
	}

	return fmt.Errorf("insufficient permissions to access %s report", reportType)
}

// ValidateSystemState validates the system state before generating reports
func (rs *ReportService) ValidateSystemState() error {
	// Check database connectivity
	db, err := rs.DB.DB()
	if err != nil {
		return fmt.Errorf("database connection error: %v", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	// Check if essential tables exist
	essentialTables := []string{"accounts", "journals", "journal_entries", "sales", "purchases"}
	for _, table := range essentialTables {
		if !rs.DB.Migrator().HasTable(table) {
			return fmt.Errorf("essential table missing: %s", table)
		}
	}

	// Check if chart of accounts is set up
	var accountCount int64
	rs.DB.Model(&models.Account{}).Where("is_active = ?", true).Count(&accountCount)
	if accountCount == 0 {
		return errors.New("chart of accounts not set up - no active accounts found")
	}

	// Check if company profile exists
	var profileCount int64
	rs.DB.Model(&models.CompanyProfile{}).Count(&profileCount)
	if profileCount == 0 {
		return errors.New("company profile not configured")
	}

	return nil
}

// LogReportGeneration logs report generation for audit purposes
func (rs *ReportService) LogReportGeneration(userID uint, reportType string, parameters map[string]interface{}, success bool, duration time.Duration) {
	// This would typically log to a dedicated audit table
	// For now, we'll use basic logging
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}

	// Create audit log entry
	auditLog := map[string]interface{}{
		"user_id":     userID,
		"report_type": reportType,
		"parameters":  parameters,
		"status":      status,
		"duration_ms": duration.Milliseconds(),
		"timestamp":   time.Now(),
	}

	// In a production system, this would be saved to an audit table
	// For now, we'll just log it
	fmt.Printf("AUDIT LOG: %+v\n", auditLog)

	// Optionally save to database audit table
	// rs.DB.Create(&models.AuditLog{...})
}

