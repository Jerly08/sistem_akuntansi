package services

import (
	"fmt"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"log"
)

// CashBankSSOTJournalAdapter handles SSOT journal integration for Cash-Bank transactions
type CashBankSSOTJournalAdapter struct {
	db                    *gorm.DB
	unifiedJournalService *UnifiedJournalService
	accountRepo           repositories.AccountRepository
}

// NewCashBankSSOTJournalAdapter creates a new adapter instance
func NewCashBankSSOTJournalAdapter(
	db *gorm.DB, 
	unifiedJournalService *UnifiedJournalService,
	accountRepo repositories.AccountRepository,
) *CashBankSSOTJournalAdapter {
	return &CashBankSSOTJournalAdapter{
		db:                    db,
		unifiedJournalService: unifiedJournalService,
		accountRepo:           accountRepo,
	}
}

// CashBankJournalRequest represents a request to create SSOT journal for cash-bank transaction
type CashBankJournalRequest struct {
	TransactionType string             // DEPOSIT, WITHDRAWAL, TRANSFER
	CashBankID      uint64            
	Amount          decimal.Decimal    
	Date            time.Time          
	Reference       string             
	Description     string             
	Notes           string             
	CounterAccountID *uint64           // Source/Target account for double entry
	CreatedBy       uint64             
	
	// For transfer transactions
	FromCashBankID  *uint64           
	ToCashBankID    *uint64           
}

// CashBankJournalResult represents the result of SSOT journal creation
type CashBankJournalResult struct {
	JournalEntry   *JournalResponse  `json:"journal_entry"`
	Success        bool              `json:"success"`
	Message        string            `json:"message"`
	TransactionRef string            `json:"transaction_ref"`
}

// CreateDepositJournalEntry creates SSOT journal entry for cash/bank deposit
func (adapter *CashBankSSOTJournalAdapter) CreateDepositJournalEntry(
	cashBank *models.CashBank,
	transaction *models.CashBankTransaction,
	request *CashBankJournalRequest,
) (*CashBankJournalResult, error) {
	return adapter.CreateDepositJournalEntryWithTx(adapter.db, cashBank, transaction, request)
}

// CreateDepositJournalEntryWithTx creates SSOT journal entry for cash/bank deposit with existing transaction
func (adapter *CashBankSSOTJournalAdapter) CreateDepositJournalEntryWithTx(
	tx *gorm.DB,
	cashBank *models.CashBank,
	transaction *models.CashBankTransaction,
	request *CashBankJournalRequest,
) (*CashBankJournalResult, error) {
	
	// Get counter account (source of funds - Equity for capital deposits)
	var counterAccountID uint64
	if request.CounterAccountID != nil {
		counterAccountID = *request.CounterAccountID
	} else {
		// Default to "Modal Pemilik" (Owner Equity) account for deposits
		// This is more appropriate for capital deposits vs operational revenue
		account, err := adapter.getOwnerEquityAccountWithTx(tx)
		if err != nil {
			// Fallback to revenue account if no equity account found
			account, err = adapter.getDefaultRevenueAccountWithTx(tx)
			if err != nil {
				return nil, fmt.Errorf("failed to get default equity or revenue account: %v", err)
			}
		}
		counterAccountID = uint64(account.ID)
	}
	
	// Create journal lines
	lines := []JournalLineRequest{
		{
			AccountID:    uint64(cashBank.AccountID),
			Description:  fmt.Sprintf("Deposit to %s", cashBank.Name),
			DebitAmount:  request.Amount,
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    counterAccountID,
			Description:  fmt.Sprintf("Capital deposit to %s", cashBank.Name),
			DebitAmount:  decimal.Zero,
			CreditAmount: request.Amount,
		},
	}
	
	// Create SSOT journal entry request
	journalRequest := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypeCashBank,
		SourceID:    func() *uint64 { id := uint64(transaction.ID); return &id }(),
		Reference:   fmt.Sprintf("DEP-%s-%d", cashBank.Code, transaction.ID),
		EntryDate:   request.Date,
		Description: fmt.Sprintf("Capital Deposit - %s: %s", cashBank.Name, request.Description),
		Lines:       lines,
		AutoPost:    true,
		CreatedBy:   request.CreatedBy,
	}
	
	// Create journal entry via SSOT using existing transaction to avoid deadlock
	journalResponse, err := adapter.unifiedJournalService.CreateJournalEntryWithTx(tx, journalRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSOT deposit journal entry: %v", err)
	}
	
	log.Printf("✅ Created SSOT deposit journal entry %s for cash-bank transaction %d", 
		journalResponse.EntryNumber, transaction.ID)
	
	return &CashBankJournalResult{
		JournalEntry:   journalResponse,
		Success:        true,
		Message:        fmt.Sprintf("Deposit journal entry %s created successfully", journalResponse.EntryNumber),
		TransactionRef: fmt.Sprintf("DEP-%s-%d", cashBank.Code, transaction.ID),
	}, nil
}

// CreateWithdrawalJournalEntry creates SSOT journal entry for cash/bank withdrawal
func (adapter *CashBankSSOTJournalAdapter) CreateWithdrawalJournalEntry(
	cashBank *models.CashBank,
	transaction *models.CashBankTransaction,
	request *CashBankJournalRequest,
) (*CashBankJournalResult, error) {
	return adapter.CreateWithdrawalJournalEntryWithTx(adapter.db, cashBank, transaction, request)
}

// CreateWithdrawalJournalEntryWithTx creates SSOT journal entry for cash/bank withdrawal with existing transaction
func (adapter *CashBankSSOTJournalAdapter) CreateWithdrawalJournalEntryWithTx(
	tx *gorm.DB,
	cashBank *models.CashBank,
	transaction *models.CashBankTransaction,
	request *CashBankJournalRequest,
) (*CashBankJournalResult, error) {
	
	// Get counter account (destination of funds - Expense account)
	var counterAccountID uint64
	if request.CounterAccountID != nil {
		counterAccountID = *request.CounterAccountID
	} else {
		// Default to "General Expense" account
		account, err := adapter.getDefaultExpenseAccountWithTx(tx)
		if err != nil {
			return nil, fmt.Errorf("failed to get default expense account: %v", err)
		}
		counterAccountID = uint64(account.ID)
	}
	
	// Create journal lines
	lines := []JournalLineRequest{
		{
			AccountID:    counterAccountID,
			Description:  fmt.Sprintf("Expense from %s withdrawal", cashBank.Name),
			DebitAmount:  request.Amount,
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    uint64(cashBank.AccountID),
			Description:  fmt.Sprintf("Withdrawal from %s", cashBank.Name),
			DebitAmount:  decimal.Zero,
			CreditAmount: request.Amount,
		},
	}
	
	// Create SSOT journal entry request
	journalRequest := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypeCashBank,
		SourceID:    func() *uint64 { id := uint64(transaction.ID); return &id }(),
		Reference:   fmt.Sprintf("WTH-%s-%d", cashBank.Code, transaction.ID),
		EntryDate:   request.Date,
		Description: fmt.Sprintf("Cash/Bank Withdrawal - %s: %s", cashBank.Name, request.Description),
		Lines:       lines,
		AutoPost:    true,
		CreatedBy:   request.CreatedBy,
	}
	
	// Create journal entry via SSOT using existing transaction to avoid deadlock
	journalResponse, err := adapter.unifiedJournalService.CreateJournalEntryWithTx(tx, journalRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSOT withdrawal journal entry: %v", err)
	}
	
	log.Printf("✅ Created SSOT withdrawal journal entry %s for cash-bank transaction %d", 
		journalResponse.EntryNumber, transaction.ID)
	
	return &CashBankJournalResult{
		JournalEntry:   journalResponse,
		Success:        true,
		Message:        fmt.Sprintf("Withdrawal journal entry %s created successfully", journalResponse.EntryNumber),
		TransactionRef: fmt.Sprintf("WTH-%s-%d", cashBank.Code, transaction.ID),
	}, nil
}

// CreateTransferJournalEntry creates SSOT journal entry for cash/bank transfer
func (adapter *CashBankSSOTJournalAdapter) CreateTransferJournalEntry(
	fromCashBank *models.CashBank,
	toCashBank *models.CashBank,
	transaction *models.CashBankTransaction,
	request *CashBankJournalRequest,
) (*CashBankJournalResult, error) {
	return adapter.CreateTransferJournalEntryWithTx(adapter.db, fromCashBank, toCashBank, transaction, request)
}

// CreateTransferJournalEntryWithTx creates SSOT journal entry for cash/bank transfer with existing transaction
func (adapter *CashBankSSOTJournalAdapter) CreateTransferJournalEntryWithTx(
	tx *gorm.DB,
	fromCashBank *models.CashBank,
	toCashBank *models.CashBank,
	transaction *models.CashBankTransaction,
	request *CashBankJournalRequest,
) (*CashBankJournalResult, error) {
	
	// Create journal lines for transfer
	lines := []JournalLineRequest{
		{
			AccountID:    uint64(toCashBank.AccountID),
			Description:  fmt.Sprintf("Transfer from %s to %s", fromCashBank.Name, toCashBank.Name),
			DebitAmount:  request.Amount,
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    uint64(fromCashBank.AccountID),
			Description:  fmt.Sprintf("Transfer from %s to %s", fromCashBank.Name, toCashBank.Name),
			DebitAmount:  decimal.Zero,
			CreditAmount: request.Amount,
		},
	}
	
	// Create SSOT journal entry request
	journalRequest := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypeCashBank,
		SourceID:    func() *uint64 { id := uint64(transaction.ID); return &id }(),
		Reference:   fmt.Sprintf("TRF-%s-TO-%s-%d", fromCashBank.Code, toCashBank.Code, transaction.ID),
		EntryDate:   request.Date,
		Description: fmt.Sprintf("Cash/Bank Transfer - From %s to %s: %s", fromCashBank.Name, toCashBank.Name, request.Description),
		Lines:       lines,
		AutoPost:    true,
		CreatedBy:   request.CreatedBy,
	}
	
	// Create journal entry via SSOT using existing transaction to avoid deadlock
	journalResponse, err := adapter.unifiedJournalService.CreateJournalEntryWithTx(tx, journalRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSOT transfer journal entry: %v", err)
	}
	
	log.Printf("✅ Created SSOT transfer journal entry %s for cash-bank transaction %d", 
		journalResponse.EntryNumber, transaction.ID)
	
	return &CashBankJournalResult{
		JournalEntry:   journalResponse,
		Success:        true,
		Message:        fmt.Sprintf("Transfer journal entry %s created successfully", journalResponse.EntryNumber),
		TransactionRef: fmt.Sprintf("TRF-%s-TO-%s-%d", fromCashBank.Code, toCashBank.Code, transaction.ID),
	}, nil
}

// CreateOpeningBalanceJournalEntry creates SSOT journal entry for opening balance
func (adapter *CashBankSSOTJournalAdapter) CreateOpeningBalanceJournalEntry(
	cashBank *models.CashBank,
	transaction *models.CashBankTransaction,
	request *CashBankJournalRequest,
) (*CashBankJournalResult, error) {
	return adapter.CreateOpeningBalanceJournalEntryWithTx(adapter.db, cashBank, transaction, request)
}

// CreateOpeningBalanceJournalEntryWithTx creates SSOT journal entry for opening balance with existing transaction
func (adapter *CashBankSSOTJournalAdapter) CreateOpeningBalanceJournalEntryWithTx(
	tx *gorm.DB,
	cashBank *models.CashBank,
	transaction *models.CashBankTransaction,
	request *CashBankJournalRequest,
) (*CashBankJournalResult, error) {
	
	// Get owner's equity account for opening balance
	equityAccount, err := adapter.getOwnerEquityAccountWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner equity account: %v", err)
	}
	
	// Create journal lines
	lines := []JournalLineRequest{
		{
			AccountID:    uint64(cashBank.AccountID),
			Description:  fmt.Sprintf("Opening balance for %s", cashBank.Name),
			DebitAmount:  request.Amount,
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    uint64(equityAccount.ID),
			Description:  fmt.Sprintf("Owner equity - Opening balance %s", cashBank.Name),
			DebitAmount:  decimal.Zero,
			CreditAmount: request.Amount,
		},
	}
	
	// Create SSOT journal entry request
	journalRequest := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypeCashBank,
		SourceID:    func() *uint64 { id := uint64(transaction.ID); return &id }(),
		Reference:   fmt.Sprintf("OPN-%s-%d", cashBank.Code, transaction.ID),
		EntryDate:   request.Date,
		Description: fmt.Sprintf("Opening Balance - %s: %s", cashBank.Name, request.Description),
		Lines:       lines,
		AutoPost:    true,
		CreatedBy:   request.CreatedBy,
	}
	
	// Create journal entry via SSOT using existing transaction to avoid deadlock
	journalResponse, err := adapter.unifiedJournalService.CreateJournalEntryWithTx(tx, journalRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSOT opening balance journal entry: %v", err)
	}
	
	log.Printf("✅ Created SSOT opening balance journal entry %s for cash-bank account %s", 
		journalResponse.EntryNumber, cashBank.Name)
	
	return &CashBankJournalResult{
		JournalEntry:   journalResponse,
		Success:        true,
		Message:        fmt.Sprintf("Opening balance journal entry %s created successfully", journalResponse.EntryNumber),
		TransactionRef: fmt.Sprintf("OPN-%s-%d", cashBank.Code, transaction.ID),
	}, nil
}

// Helper methods to get default accounts

// getDefaultRevenueAccount returns the default revenue account for deposits
func (adapter *CashBankSSOTJournalAdapter) getDefaultRevenueAccount() (*models.Account, error) {
	return adapter.getDefaultRevenueAccountWithTx(adapter.db)
}

// getDefaultRevenueAccountWithTx returns the default revenue account for deposits with transaction
func (adapter *CashBankSSOTJournalAdapter) getDefaultRevenueAccountWithTx(tx *gorm.DB) (*models.Account, error) {
	// Look for "Other Revenue" or create if not exists
	var account models.Account
	
	// Try to find existing "Other Revenue" account
	err := tx.Where("code = ? OR name ILIKE ?", "4900", "%Other Income%").First(&account).Error
	if err == nil {
		return &account, nil
	}
	
	// Try to find any revenue account (case insensitive)
	err = tx.Where("type ILIKE ? AND is_active = ?", "%REVENUE%", true).First(&account).Error
	if err == nil {
		return &account, nil
	}
	
	return nil, fmt.Errorf("no active revenue account found for deposits")
}

// getDefaultExpenseAccount returns the default expense account for withdrawals
func (adapter *CashBankSSOTJournalAdapter) getDefaultExpenseAccount() (*models.Account, error) {
	return adapter.getDefaultExpenseAccountWithTx(adapter.db)
}

// getDefaultExpenseAccountWithTx returns the default expense account for withdrawals with transaction
func (adapter *CashBankSSOTJournalAdapter) getDefaultExpenseAccountWithTx(tx *gorm.DB) (*models.Account, error) {
	// Look for "General Expense" or create if not exists
	var account models.Account
	
	// Try to find existing "General Expense" account
	err := tx.Where("code = ? OR name ILIKE ?", "5900", "%General Expense%").First(&account).Error
	if err == nil {
		return &account, nil
	}
	
	// Try to find any expense account (case insensitive)
	err = tx.Where("type ILIKE ? AND is_active = ?", "%EXPENSE%", true).First(&account).Error
	if err == nil {
		return &account, nil
	}
	
	return nil, fmt.Errorf("no active expense account found for withdrawals")
}

// getOwnerEquityAccount returns the owner equity account for opening balances
func (adapter *CashBankSSOTJournalAdapter) getOwnerEquityAccount() (*models.Account, error) {
	return adapter.getOwnerEquityAccountWithTx(adapter.db)
}

// getOwnerEquityAccountWithTx returns the owner equity account for opening balances with transaction
func (adapter *CashBankSSOTJournalAdapter) getOwnerEquityAccountWithTx(tx *gorm.DB) (*models.Account, error) {
	// Look for "Modal Pemilik" or owner equity account
	var account models.Account
	
	// Try to find existing owner equity account
	err := tx.Where("code = ? OR name ILIKE ?", "3101", "%Modal Pemilik%").First(&account).Error
	if err == nil {
		return &account, nil
	}
	
	// Try to find any equity account (case insensitive)
	err = tx.Where("type ILIKE ? AND is_active = ?", "%EQUITY%", true).First(&account).Error
	if err == nil {
		return &account, nil
	}
	
	return nil, fmt.Errorf("no active equity account found for opening balance")
}

// ReverseJournalEntry creates a reversal journal entry for cash-bank transaction
func (adapter *CashBankSSOTJournalAdapter) ReverseJournalEntry(
	transactionID uint64,
	reason string,
	createdBy uint64,
) (*CashBankJournalResult, error) {
	
	// Find the original SSOT journal entry
	var originalJournal models.SSOTJournalEntry
	err := adapter.db.Where("source_type = ? AND source_id = ?", 
		models.SSOTSourceTypeCashBank, transactionID).First(&originalJournal).Error
	if err != nil {
		return nil, fmt.Errorf("original cash-bank journal entry not found: %v", err)
	}
	
	// Use SSOT service to create reversal
	reversalResponse, err := adapter.unifiedJournalService.ReverseJournalEntry(
		originalJournal.ID, reason, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create reversal journal entry: %v", err)
	}
	
	log.Printf("✅ Created SSOT reversal journal entry %s for cash-bank transaction %d", 
		reversalResponse.EntryNumber, transactionID)
	
	return &CashBankJournalResult{
		JournalEntry:   reversalResponse,
		Success:        true,
		Message:        fmt.Sprintf("Reversal journal entry %s created successfully", reversalResponse.EntryNumber),
		TransactionRef: fmt.Sprintf("REV-CASHBANK-%d", transactionID),
	}, nil
}

// ValidateJournalIntegrity validates the integrity of SSOT journal entries for cash-bank transactions
func (adapter *CashBankSSOTJournalAdapter) ValidateJournalIntegrity() error {
	// Get all cash-bank transactions
	var transactions []models.CashBankTransaction
	err := adapter.db.Find(&transactions).Error
	if err != nil {
		return fmt.Errorf("failed to get cash-bank transactions: %v", err)
	}
	
	for _, tx := range transactions {
		// Check if SSOT journal entry exists
		var count int64
		adapter.db.Model(&models.SSOTJournalEntry{}).
			Where("source_type = ? AND source_id = ?", models.SSOTSourceTypeCashBank, tx.ID).
			Count(&count)
		
		if count == 0 {
			log.Printf("⚠️ Warning: Cash-bank transaction %d has no SSOT journal entry", tx.ID)
		}
	}
	
	return nil
}