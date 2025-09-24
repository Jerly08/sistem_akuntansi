package services

import (
	"fmt"
	"time"

	"app-sistem-akuntansi/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PaymentJournalFactory creates SSOT journal entries from payment transactions
type PaymentJournalFactory struct {
	db             *gorm.DB
	journalService *UnifiedJournalService
}

// NewPaymentJournalFactory creates a new instance of PaymentJournalFactory
func NewPaymentJournalFactory(db *gorm.DB, journalService *UnifiedJournalService) *PaymentJournalFactory {
	return &PaymentJournalFactory{
		db:             db,
		journalService: journalService,
	}
}

// PaymentJournalRequest represents the request for creating payment journal entry
type PaymentJournalRequest struct {
	PaymentID     uint64
	ContactID     uint64
	Amount        decimal.Decimal
	Date          time.Time
	Method        string
	Reference     string
	Notes         string
	CashBankID    uint64
	CreatedBy     uint64
	ContactType   string // CUSTOMER or VENDOR
	ContactName   string
}

// PaymentJournalResult represents the result of payment journal creation
type PaymentJournalResult struct {
	JournalEntry    *JournalResponse            `json:"journal_entry"`
	AccountUpdates  []AccountBalanceUpdate      `json:"account_updates"`
	Success         bool                        `json:"success"`
	Message         string                      `json:"message"`
}

// AccountBalanceUpdate represents account balance change
type AccountBalanceUpdate struct {
	AccountID    uint64          `json:"account_id"`
	AccountCode  string          `json:"account_code"`
	AccountName  string          `json:"account_name"`
	OldBalance   decimal.Decimal `json:"old_balance"`
	NewBalance   decimal.Decimal `json:"new_balance"`
	Change       decimal.Decimal `json:"change"`
	ChangeType   string          `json:"change_type"` // INCREASE, DECREASE
}

// CreatePaymentJournalEntry creates a SSOT journal entry for a payment transaction
func (pjf *PaymentJournalFactory) CreatePaymentJournalEntry(req *PaymentJournalRequest) (*PaymentJournalResult, error) {
	// Validate request
	if err := pjf.validateRequest(req); err != nil {
		return nil, fmt.Errorf("payment journal validation failed: %w", err)
	}

	// Determine journal entry type based on payment method and contact type
	sourceType := pjf.determineSourceType(req.Method, req.ContactType)
	
	// Get account balances before transaction
	accountUpdates, err := pjf.getAccountUpdatesPreview(req)
	if err != nil {
		return nil, fmt.Errorf("failed to preview account updates: %w", err)
	}

	// Create journal lines based on payment type
	journalLines, err := pjf.createJournalLines(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create journal lines: %w", err)
	}

	// Create journal entry request
	journalReq := &JournalEntryRequest{
		SourceType:  sourceType,
		SourceID:    &req.PaymentID,
		Reference:   req.Reference,
		EntryDate:   req.Date,
		Description: pjf.generateDescription(req),
		Lines:       journalLines,
		AutoPost:    true,
		CreatedBy:   req.CreatedBy,
	}

	// Create the journal entry via SSOT Journal Service
	journalResp, err := pjf.journalService.CreateJournalEntry(journalReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSOT journal entry: %w", err)
	}

	return &PaymentJournalResult{
		JournalEntry:   journalResp,
		AccountUpdates: accountUpdates,
		Success:        true,
		Message:        fmt.Sprintf("Journal entry %s created successfully", journalResp.EntryNumber),
	}, nil
}

// CreateReceivablePaymentJournal creates journal entry for receivable payment (customer pays us)
func (pjf *PaymentJournalFactory) CreateReceivablePaymentJournal(payment *models.Payment, contact *models.Contact, cashBankAccount *models.CashBank) (*PaymentJournalResult, error) {
	req := &PaymentJournalRequest{
		PaymentID:   uint64(payment.ID),
		ContactID:   uint64(payment.ContactID),
		Amount:      decimal.NewFromFloat(payment.Amount),
		Date:        payment.Date,
		Method:      payment.Method,
		Reference:   payment.Reference,
		Notes:       payment.Notes,
		CashBankID:  uint64(cashBankAccount.ID),
		CreatedBy:   uint64(payment.UserID),
		ContactType: contact.Type,
		ContactName: contact.Name,
	}

	return pjf.CreatePaymentJournalEntry(req)
}

// CreatePayablePaymentJournal creates journal entry for payable payment (we pay vendor)
func (pjf *PaymentJournalFactory) CreatePayablePaymentJournal(payment *models.Payment, contact *models.Contact, cashBankAccount *models.CashBank) (*PaymentJournalResult, error) {
	req := &PaymentJournalRequest{
		PaymentID:   uint64(payment.ID),
		ContactID:   uint64(payment.ContactID),
		Amount:      decimal.NewFromFloat(payment.Amount),
		Date:        payment.Date,
		Method:      payment.Method,
		Reference:   payment.Reference,
		Notes:       payment.Notes,
		CashBankID:  uint64(cashBankAccount.ID),
		CreatedBy:   uint64(payment.UserID),
		ContactType: contact.Type,
		ContactName: contact.Name,
	}

	return pjf.CreatePaymentJournalEntry(req)
}

// determineSourceType determines the SSOT journal source type based on payment method and contact type
func (pjf *PaymentJournalFactory) determineSourceType(method, contactType string) string {
	if contactType == "CUSTOMER" {
		return models.SSOTSourceTypePayment + "_RECEIVABLE"
	} else if contactType == "VENDOR" {
		return models.SSOTSourceTypePayment + "_PAYABLE"
	}
	return models.SSOTSourceTypePayment
}

// createJournalLines creates the appropriate journal lines for the payment
func (pjf *PaymentJournalFactory) createJournalLines(req *PaymentJournalRequest) ([]JournalLineRequest, error) {
	var lines []JournalLineRequest

	// Get required accounts
	accounts, err := pjf.getRequiredAccounts(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get required accounts: %w", err)
	}

	if req.ContactType == "CUSTOMER" {
		// Receivable Payment: Dr. Cash/Bank, Cr. Accounts Receivable
		lines = []JournalLineRequest{
			{
				AccountID:    accounts.CashBankAccountID,
				Description:  fmt.Sprintf("Payment received from %s", req.ContactName),
				DebitAmount:  req.Amount,
				CreditAmount: decimal.Zero,
			},
			{
				AccountID:    accounts.AccountsReceivableID,
				Description:  fmt.Sprintf("Payment against receivables - %s", req.ContactName),
				DebitAmount:  decimal.Zero,
				CreditAmount: req.Amount,
			},
		}
	} else if req.ContactType == "VENDOR" {
		// Payable Payment: Dr. Accounts Payable, Cr. Cash/Bank
		lines = []JournalLineRequest{
			{
				AccountID:    accounts.AccountsPayableID,
				Description:  fmt.Sprintf("Payment to %s", req.ContactName),
				DebitAmount:  req.Amount,
				CreditAmount: decimal.Zero,
			},
			{
				AccountID:    accounts.CashBankAccountID,
				Description:  fmt.Sprintf("Payment made - %s", req.ContactName),
				DebitAmount:  decimal.Zero,
				CreditAmount: req.Amount,
			},
		}
	} else {
		return nil, fmt.Errorf("unsupported contact type: %s", req.ContactType)
	}

	return lines, nil
}

// RequiredAccounts represents the accounts needed for payment journal entries
type RequiredAccounts struct {
	CashBankAccountID      uint64
	AccountsReceivableID   uint64
	AccountsPayableID      uint64
}

// getRequiredAccounts retrieves the necessary account IDs for payment journal entries
func (pjf *PaymentJournalFactory) getRequiredAccounts(req *PaymentJournalRequest) (*RequiredAccounts, error) {
	accounts := &RequiredAccounts{}

	// Get Cash/Bank account from CashBank model
	var cashBank models.CashBank
	if err := pjf.db.Preload("Account").First(&cashBank, req.CashBankID).Error; err != nil {
		return nil, fmt.Errorf("cash/bank account not found: %w", err)
	}
	accounts.CashBankAccountID = uint64(cashBank.AccountID)

	// Get Accounts Receivable account
	var arAccount models.Account
	if err := pjf.db.Where("code = ? AND type = ?", "1201", "ASSET").First(&arAccount).Error; err != nil {
		return nil, fmt.Errorf("accounts receivable account (1201) not found: %w", err)
	}
	accounts.AccountsReceivableID = uint64(arAccount.ID)

	// Get Accounts Payable account
	var apAccount models.Account
	if err := pjf.db.Where("code = ? AND type = ?", "2101", "LIABILITY").First(&apAccount).Error; err != nil {
		return nil, fmt.Errorf("accounts payable account (2101) not found: %w", err)
	}
	accounts.AccountsPayableID = uint64(apAccount.ID)

	return accounts, nil
}

// getAccountUpdatesPreview previews the account balance changes that will occur
func (pjf *PaymentJournalFactory) getAccountUpdatesPreview(req *PaymentJournalRequest) ([]AccountBalanceUpdate, error) {
	var updates []AccountBalanceUpdate

	// Get required accounts
	accounts, err := pjf.getRequiredAccounts(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get required accounts: %w", err)
	}

	// Get current balances from SSOT materialized view
	balances, err := pjf.journalService.GetAccountBalances()
	if err != nil {
		return nil, fmt.Errorf("failed to get current account balances: %w", err)
	}

	// Create balance map for easy lookup
	balanceMap := make(map[uint64]models.SSOTAccountBalance)
	for _, balance := range balances {
		balanceMap[balance.AccountID] = balance
	}

	if req.ContactType == "CUSTOMER" {
		// Cash/Bank account increases (Debit)
		cashBalance := balanceMap[accounts.CashBankAccountID]
		updates = append(updates, AccountBalanceUpdate{
			AccountID:   accounts.CashBankAccountID,
			AccountCode: cashBalance.AccountCode,
			AccountName: cashBalance.AccountName,
			OldBalance:  cashBalance.CurrentBalance,
			NewBalance:  cashBalance.CurrentBalance.Add(req.Amount),
			Change:      req.Amount,
			ChangeType:  "INCREASE",
		})

		// Accounts Receivable decreases (Credit)
		arBalance := balanceMap[accounts.AccountsReceivableID]
		updates = append(updates, AccountBalanceUpdate{
			AccountID:   accounts.AccountsReceivableID,
			AccountCode: arBalance.AccountCode,
			AccountName: arBalance.AccountName,
			OldBalance:  arBalance.CurrentBalance,
			NewBalance:  arBalance.CurrentBalance.Sub(req.Amount),
			Change:      req.Amount.Neg(),
			ChangeType:  "DECREASE",
		})
	} else if req.ContactType == "VENDOR" {
		// Accounts Payable decreases (Debit)
		apBalance := balanceMap[accounts.AccountsPayableID]
		updates = append(updates, AccountBalanceUpdate{
			AccountID:   accounts.AccountsPayableID,
			AccountCode: apBalance.AccountCode,
			AccountName: apBalance.AccountName,
			OldBalance:  apBalance.CurrentBalance,
			NewBalance:  apBalance.CurrentBalance.Sub(req.Amount),
			Change:      req.Amount.Neg(),
			ChangeType:  "DECREASE",
		})

		// Cash/Bank account decreases (Credit)
		cashBalance := balanceMap[accounts.CashBankAccountID]
		updates = append(updates, AccountBalanceUpdate{
			AccountID:   accounts.CashBankAccountID,
			AccountCode: cashBalance.AccountCode,
			AccountName: cashBalance.AccountName,
			OldBalance:  cashBalance.CurrentBalance,
			NewBalance:  cashBalance.CurrentBalance.Sub(req.Amount),
			Change:      req.Amount.Neg(),
			ChangeType:  "DECREASE",
		})
	}

	return updates, nil
}

// generateDescription generates a descriptive text for the journal entry
func (pjf *PaymentJournalFactory) generateDescription(req *PaymentJournalRequest) string {
	if req.ContactType == "CUSTOMER" {
		return fmt.Sprintf("Payment received from %s - %s", req.ContactName, req.Reference)
	} else if req.ContactType == "VENDOR" {
		return fmt.Sprintf("Payment to %s - %s", req.ContactName, req.Reference)
	}
	return fmt.Sprintf("Payment transaction - %s", req.Reference)
}

// validateRequest validates the payment journal request
func (pjf *PaymentJournalFactory) validateRequest(req *PaymentJournalRequest) error {
	if req.PaymentID == 0 {
		return fmt.Errorf("payment ID is required")
	}

	if req.ContactID == 0 {
		return fmt.Errorf("contact ID is required")
	}

	if req.Amount.IsZero() || req.Amount.IsNegative() {
		return fmt.Errorf("payment amount must be positive")
	}

	if req.Date.IsZero() {
		return fmt.Errorf("payment date is required")
	}

	if req.Method == "" {
		return fmt.Errorf("payment method is required")
	}

	if req.ContactType != "CUSTOMER" && req.ContactType != "VENDOR" {
		return fmt.Errorf("contact type must be CUSTOMER or VENDOR")
	}

	if req.CashBankID == 0 {
		return fmt.Errorf("cash/bank account ID is required")
	}

	if req.CreatedBy == 0 {
		return fmt.Errorf("created by user ID is required")
	}

	return nil
}

// PreviewPaymentJournalEntry previews what journal entry would be created without actually creating it
func (pjf *PaymentJournalFactory) PreviewPaymentJournalEntry(req *PaymentJournalRequest) (*PaymentJournalResult, error) {
	// Validate request
	if err := pjf.validateRequest(req); err != nil {
		return nil, fmt.Errorf("payment journal validation failed: %w", err)
	}

	// Get account balances preview
	accountUpdates, err := pjf.getAccountUpdatesPreview(req)
	if err != nil {
		return nil, fmt.Errorf("failed to preview account updates: %w", err)
	}

	// Create journal lines preview
	journalLines, err := pjf.createJournalLines(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create journal lines preview: %w", err)
	}

	// Create preview response (without actually creating the journal entry)
	sourceType := pjf.determineSourceType(req.Method, req.ContactType)
	
	previewResponse := &JournalResponse{
		EntryNumber: fmt.Sprintf("PREVIEW-%s", sourceType),
		Status:      "PREVIEW",
		TotalDebit:  req.Amount,
		TotalCredit: req.Amount,
		IsBalanced:  true,
		Lines:       make([]JournalLineResponse, len(journalLines)),
	}

	// Convert journal lines to response format
	for i, line := range journalLines {
		previewResponse.Lines[i] = JournalLineResponse{
			LineNumber:   i + 1,
			AccountID:    line.AccountID,
			Description:  line.Description,
			DebitAmount:  line.DebitAmount,
			CreditAmount: line.CreditAmount,
		}
	}

	return &PaymentJournalResult{
		JournalEntry:   previewResponse,
		AccountUpdates: accountUpdates,
		Success:        true,
		Message:        "Journal entry preview generated successfully",
	}, nil
}

// ReversePaymentJournal creates a reversal journal entry for a payment
func (pjf *PaymentJournalFactory) ReversePaymentJournal(paymentID uint64, reason string, createdBy uint64) (*PaymentJournalResult, error) {
	// Find the original journal entry
	var originalJournal models.SSOTJournalEntry
	if err := pjf.db.Where("source_type LIKE ? AND source_id = ?", "%PAYMENT%", paymentID).
		Preload("Lines").First(&originalJournal).Error; err != nil {
		return nil, fmt.Errorf("original payment journal entry not found: %w", err)
	}

	// Use SSOT Journal Service to create reversal
	reversalResp, err := pjf.journalService.ReverseJournalEntry(originalJournal.ID, reason, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create reversal journal entry: %w", err)
	}

	return &PaymentJournalResult{
		JournalEntry: reversalResp,
		Success:      true,
		Message:      fmt.Sprintf("Reversal journal entry %s created successfully", reversalResp.EntryNumber),
	}, nil
}