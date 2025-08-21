package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

type CashBankService struct {
	db           *gorm.DB
	cashBankRepo *repositories.CashBankRepository
	accountRepo  repositories.AccountRepository
}

func NewCashBankService(
	db *gorm.DB,
	cashBankRepo *repositories.CashBankRepository,
	accountRepo repositories.AccountRepository,
) *CashBankService {
	return &CashBankService{
		db:           db,
		cashBankRepo: cashBankRepo,
		accountRepo:  accountRepo,
	}
}

// Transaction Types
const (
	TransactionTypeDeposit     = "DEPOSIT"
	TransactionTypeWithdrawal  = "WITHDRAWAL"
	TransactionTypeTransfer    = "TRANSFER"
	TransactionTypeAdjustment  = "ADJUSTMENT"
	TransactionTypeOpeningBalance = "OPENING_BALANCE"
)

// GetCashBankAccounts retrieves all cash and bank accounts
func (s *CashBankService) GetCashBankAccounts() ([]models.CashBank, error) {
	return s.cashBankRepo.FindAll()
}

// GetCashBankByID retrieves cash/bank account by ID
func (s *CashBankService) GetCashBankByID(id uint) (*models.CashBank, error) {
	return s.cashBankRepo.FindByID(id)
}

// CreateCashBankAccount creates a new cash or bank account
func (s *CashBankService) CreateCashBankAccount(request CashBankCreateRequest, userID uint) (*models.CashBank, error) {
	// Start transaction
	tx := s.db.Begin()
	
	// Validate GL account if provided
	var glAccount *models.Account
	if request.AccountID > 0 {
		account, err := s.accountRepo.FindByID(context.Background(), request.AccountID)
		if err != nil {
			tx.Rollback()
			return nil, errors.New("GL account not found")
		}
		glAccount = account
	} else {
		// Create default GL account
		accountCode := s.generateAccountCode(request.Type)
		newAccount := &models.Account{
			Code:        accountCode,
			Name:        request.Name,
			Type:        "ASSET",
			Category:    s.getAccountCategory(request.Type),
			IsActive:    true,
			Description: fmt.Sprintf("Auto-created %s account", request.Type),
		}
		
		if err := tx.Create(newAccount).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		glAccount = newAccount
	}
	
	// Generate code
	code := s.generateCashBankCode(request.Type)
	
	// Create cash/bank account
	cashBank := &models.CashBank{
		Code:        code,
		Name:        request.Name,
		Type:        request.Type,
		AccountID:   glAccount.ID,
		BankName:    request.BankName,
		AccountNo:   request.AccountNo,
		Currency:    request.Currency,
		Balance:     0, // Will be set via opening balance transaction
		IsActive:    true,
		Description: request.Description,
	}
	
	if err := tx.Create(cashBank).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Create opening balance transaction if provided
	if request.OpeningBalance > 0 {
		openingDate := request.OpeningDate.ToTime()
		if openingDate.IsZero() {
			openingDate = time.Now() // Use current time if no date provided
		}
		err := s.createOpeningBalanceTransaction(tx, cashBank, request.OpeningBalance, openingDate, userID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	
	return cashBank, tx.Commit().Error
}

// UpdateCashBankAccount updates cash/bank account details
func (s *CashBankService) UpdateCashBankAccount(id uint, request CashBankUpdateRequest) (*models.CashBank, error) {
	cashBank, err := s.cashBankRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	
	// Update fields
	if request.Name != "" {
		cashBank.Name = request.Name
	}
	if request.BankName != "" {
		cashBank.BankName = request.BankName
	}
	if request.AccountNo != "" {
		cashBank.AccountNo = request.AccountNo
	}
	if request.Description != "" {
		cashBank.Description = request.Description
	}
	if request.IsActive != nil {
		cashBank.IsActive = *request.IsActive
	}
	
	return s.cashBankRepo.Update(cashBank)
}

// DeleteCashBankAccount deletes (soft delete) cash/bank account
func (s *CashBankService) DeleteCashBankAccount(id uint) error {
	// Check if account exists
	cashBank, err := s.cashBankRepo.FindByID(id)
	if err != nil {
		return errors.New("account not found")
	}
	
	// Check if account has balance
	if cashBank.Balance != 0 {
		return fmt.Errorf("cannot delete account with non-zero balance: %.2f", cashBank.Balance)
	}
	
	// Check if account has transactions
	// In a real implementation, you might want to check for recent transactions
	// For now, we'll allow deletion if balance is zero
	
	return s.cashBankRepo.Delete(id)
}

// ProcessTransfer processes transfer between cash/bank accounts
func (s *CashBankService) ProcessTransfer(request TransferRequest, userID uint) (*CashBankTransfer, error) {
	// Start transaction
	tx := s.db.Begin()
	
	// Validate source account
	sourceAccount, err := s.cashBankRepo.FindByID(request.FromAccountID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("source account not found")
	}
	
	// Validate destination account
	destAccount, err := s.cashBankRepo.FindByID(request.ToAccountID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("destination account not found")
	}
	
	// Check balance
	if sourceAccount.Balance < request.Amount {
		tx.Rollback()
		return nil, fmt.Errorf("insufficient balance. Available: %.2f", sourceAccount.Balance)
	}
	
	// Apply exchange rate if different currencies
	transferAmount := request.Amount
	if sourceAccount.Currency != destAccount.Currency {
		if request.ExchangeRate <= 0 {
			tx.Rollback()
			return nil, errors.New("exchange rate required for different currencies")
		}
		transferAmount = request.Amount * request.ExchangeRate
	}
	
	// Create transfer record
	transfer := &CashBankTransfer{
		TransferNumber: s.generateTransferNumber(),
		FromAccountID:  request.FromAccountID,
		ToAccountID:    request.ToAccountID,
		Date:           request.Date.ToTime(),
		Amount:         request.Amount,
		ExchangeRate:   request.ExchangeRate,
		ConvertedAmount: transferAmount,
		Reference:      request.Reference,
		Notes:          request.Notes,
		Status:         "COMPLETED",
		UserID:         userID,
	}
	
	if err := tx.Create(transfer).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Update source account balance
	sourceAccount.Balance -= request.Amount
	if err := tx.Save(sourceAccount).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Create source transaction record
	sourceTx := &models.CashBankTransaction{
		CashBankID:      request.FromAccountID,
		ReferenceType:   "TRANSFER",
		ReferenceID:     transfer.ID,
		Amount:          -request.Amount,
		BalanceAfter:    sourceAccount.Balance,
		TransactionDate: request.Date.ToTime(),
		Notes:           fmt.Sprintf("Transfer to %s", destAccount.Name),
	}
	
	if err := tx.Create(sourceTx).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Update destination account balance
	destAccount.Balance += transferAmount
	if err := tx.Save(destAccount).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Create destination transaction record
	destTx := &models.CashBankTransaction{
		CashBankID:      request.ToAccountID,
		ReferenceType:   "TRANSFER",
		ReferenceID:     transfer.ID,
		Amount:          transferAmount,
		BalanceAfter:    destAccount.Balance,
		TransactionDate: request.Date.ToTime(),
		Notes:           fmt.Sprintf("Transfer from %s", sourceAccount.Name),
	}
	
	if err := tx.Create(destTx).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Create journal entries
	err = s.createTransferJournalEntries(tx, transfer, sourceAccount, destAccount, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	
	return transfer, tx.Commit().Error
}

// ProcessDeposit processes a deposit transaction
func (s *CashBankService) ProcessDeposit(request DepositRequest, userID uint) (*models.CashBankTransaction, error) {
	tx := s.db.Begin()
	
	// Validate account
	account, err := s.cashBankRepo.FindByID(request.AccountID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("account not found")
	}
	
	// Update balance
	account.Balance += request.Amount
	if err := tx.Save(account).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Create transaction record
	transaction := &models.CashBankTransaction{
		CashBankID:      request.AccountID,
		ReferenceType:   TransactionTypeDeposit,
		ReferenceID:     0, // No specific reference for direct deposit
		Amount:          request.Amount,
		BalanceAfter:    account.Balance,
		TransactionDate: request.Date.ToTime(),
		Notes:           request.Notes,
	}
	
	if err := tx.Create(transaction).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Create journal entries
	err = s.createDepositJournalEntries(tx, transaction, account, request, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	
	return transaction, tx.Commit().Error
}

// ProcessWithdrawal processes a withdrawal transaction
func (s *CashBankService) ProcessWithdrawal(request WithdrawalRequest, userID uint) (*models.CashBankTransaction, error) {
	tx := s.db.Begin()
	
	// Validate account
	account, err := s.cashBankRepo.FindByID(request.AccountID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("account not found")
	}
	
	// Check balance
	if account.Balance < request.Amount {
		tx.Rollback()
		return nil, fmt.Errorf("insufficient balance. Available: %.2f", account.Balance)
	}
	
	// Update balance
	account.Balance -= request.Amount
	if err := tx.Save(account).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Create transaction record
	transaction := &models.CashBankTransaction{
		CashBankID:      request.AccountID,
		ReferenceType:   TransactionTypeWithdrawal,
		ReferenceID:     0,
		Amount:          -request.Amount,
		BalanceAfter:    account.Balance,
		TransactionDate: request.Date.ToTime(),
		Notes:           request.Notes,
	}
	
	if err := tx.Create(transaction).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Create journal entries
	err = s.createWithdrawalJournalEntries(tx, transaction, account, request, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	
	return transaction, tx.Commit().Error
}

// GetTransactions retrieves transactions for a cash/bank account
func (s *CashBankService) GetTransactions(accountID uint, filter TransactionFilter) (*TransactionResult, error) {
	// Convert service filter to repository filter
	repoFilter := repositories.TransactionFilter{
		StartDate: filter.StartDate,
		EndDate:   filter.EndDate,
		Type:      filter.Type,
		Page:      filter.Page,
		Limit:     filter.Limit,
	}
	
	// Get transactions from repository
	result, err := s.cashBankRepo.GetTransactions(accountID, repoFilter)
	if err != nil {
		return nil, err
	}
	
	// Convert repository result to service result
	return &TransactionResult{
		Data:       result.Data,
		Total:      result.Total,
		Page:       result.Page,
		Limit:      result.Limit,
		TotalPages: result.TotalPages,
	}, nil
}

// GetBalanceSummary gets balance summary for all accounts
func (s *CashBankService) GetBalanceSummary() (*BalanceSummary, error) {
	accounts, err := s.cashBankRepo.FindAll()
	if err != nil {
		return nil, err
	}
	
	summary := &BalanceSummary{
		TotalCash:     0,
		TotalBank:     0,
		TotalBalance:  0,
		ByAccount:     []AccountBalance{},
		ByCurrency:    make(map[string]float64),
	}
	
	for _, account := range accounts {
		if account.Type == models.CashBankTypeCash {
			summary.TotalCash += account.Balance
		} else {
			summary.TotalBank += account.Balance
		}
		
		summary.TotalBalance += account.Balance
		
		// Group by currency
		summary.ByCurrency[account.Currency] += account.Balance
		
		// Add to account list
		summary.ByAccount = append(summary.ByAccount, AccountBalance{
			AccountID:   account.ID,
			AccountName: account.Name,
			AccountType: account.Type,
			Balance:     account.Balance,
			Currency:    account.Currency,
		})
	}
	
	return summary, nil
}

// GetPaymentAccounts gets active cash and bank accounts for payment processing
func (s *CashBankService) GetPaymentAccounts() ([]models.CashBank, error) {
	accounts, err := s.cashBankRepo.FindAll()
	if err != nil {
		return nil, err
	}
	
	// Filter only active accounts
	var paymentAccounts []models.CashBank
	for _, account := range accounts {
		if account.IsActive {
			paymentAccounts = append(paymentAccounts, account)
		}
	}
	
	return paymentAccounts, nil
}

// ReconcileAccount reconciles bank account with statement
func (s *CashBankService) ReconcileAccount(accountID uint, request ReconciliationRequest, userID uint) (*BankReconciliation, error) {
	tx := s.db.Begin()
	
	account, err := s.cashBankRepo.FindByID(accountID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	
	if account.Type != models.CashBankTypeBank {
		tx.Rollback()
		return nil, errors.New("reconciliation only for bank accounts")
	}
	
	// Create reconciliation record
	reconciliation := &BankReconciliation{
		CashBankID:       accountID,
		ReconcileDate:    request.Date.ToTime(),
		StatementBalance: request.StatementBalance,
		SystemBalance:    account.Balance,
		Difference:       request.StatementBalance - account.Balance,
		Status:           "PENDING",
		UserID:           userID,
	}
	
	if err := tx.Create(reconciliation).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Process reconciliation items
	for _, item := range request.Items {
		recItem := &ReconciliationItem{
			ReconciliationID: reconciliation.ID,
			TransactionID:    item.TransactionID,
			IsCleared:        item.IsCleared,
			Notes:            item.Notes,
		}
		
		if err := tx.Create(recItem).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	
	// If difference is zero and all items cleared, mark as completed
	if reconciliation.Difference == 0 {
		reconciliation.Status = "COMPLETED"
		if err := tx.Save(reconciliation).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	
	return reconciliation, tx.Commit().Error
}

// Helper functions

func (s *CashBankService) generateCashBankCode(accountType string) string {
	prefix := "CSH"
	if accountType == models.CashBankTypeBank {
		prefix = "BNK"
	}
	
	year := time.Now().Year()
	count, _ := s.cashBankRepo.CountByType(accountType)
	return fmt.Sprintf("%s-%04d-%04d", prefix, year, count+1)
}

func (s *CashBankService) generateAccountCode(accountType string) string {
	if accountType == models.CashBankTypeCash {
		return fmt.Sprintf("1100-%03d", time.Now().Unix()%1000)
	}
	return fmt.Sprintf("1110-%03d", time.Now().Unix()%1000)
}

func (s *CashBankService) getAccountCategory(cashBankType string) string {
	if cashBankType == models.CashBankTypeCash {
		return "CURRENT_ASSET"
	}
	return "CURRENT_ASSET"
}

func (s *CashBankService) generateTransferNumber() string {
	year := time.Now().Year()
	month := time.Now().Month()
	// Would need proper counting
	count := 1
	return fmt.Sprintf("TRF/%04d/%02d/%04d", year, month, count)
}

func (s *CashBankService) createOpeningBalanceTransaction(tx *gorm.DB, cashBank *models.CashBank, amount float64, date time.Time, userID uint) error {
	// Update balance
	cashBank.Balance = amount
	if err := tx.Save(cashBank).Error; err != nil {
		return err
	}
	
	// Create transaction record
	transaction := &models.CashBankTransaction{
		CashBankID:      cashBank.ID,
		ReferenceType:   TransactionTypeOpeningBalance,
		ReferenceID:     0,
		Amount:          amount,
		BalanceAfter:    amount,
		TransactionDate: date,
		Notes:           "Opening Balance",
	}
	
	return tx.Create(transaction).Error
}

func (s *CashBankService) createTransferJournalEntries(tx *gorm.DB, transfer *CashBankTransfer, source, dest *models.CashBank, userID uint) error {
	// Create journal entry for transfer
	journal := &models.Journal{
		Code:          fmt.Sprintf("TRF-JV/%s", time.Now().Format("20060102-150405")),
		Date:          transfer.Date,
		Description:   fmt.Sprintf("Transfer from %s to %s", source.Name, dest.Name),
		ReferenceType: "TRANSFER",
		ReferenceID:   &transfer.ID,
		UserID:        userID,
		Status:        models.JournalStatusPosted,
		Period:        transfer.Date.Format("2006-01"),
	}
	
	// Journal entries
	entries := []models.JournalEntry{
		// Debit: Destination account
		{
			AccountID:    dest.AccountID,
			Description:  fmt.Sprintf("Transfer from %s", source.Name),
			DebitAmount:  transfer.ConvertedAmount,
			CreditAmount: 0,
		},
		// Credit: Source account
		{
			AccountID:    source.AccountID,
			Description:  fmt.Sprintf("Transfer to %s", dest.Name),
			DebitAmount:  0,
			CreditAmount: transfer.Amount,
		},
	}
	
	// Handle exchange rate difference if applicable
	if transfer.ExchangeRate > 0 && transfer.ExchangeRate != 1 {
		exchangeDiff := transfer.ConvertedAmount - transfer.Amount
		if exchangeDiff != 0 {
			// Get exchange gain/loss account
			var exchangeAccountID uint
			// TODO: Implement GetAccountByCode or FindByCode
			// For now, skip exchange rate handling
			if exchangeDiff > 0 {
				// Exchange gain - would need account lookup
				// account, _ := s.accountRepo.FindByCode(context.Background(), "7100")
			} else {
				// Exchange loss - would need account lookup
				// account, _ := s.accountRepo.FindByCode(context.Background(), "8100")
			}
			
			if exchangeAccountID > 0 {
				entry := models.JournalEntry{
					AccountID:   exchangeAccountID,
					Description: "Exchange rate difference",
				}
				
				if exchangeDiff > 0 {
					entry.CreditAmount = exchangeDiff
				} else {
					entry.DebitAmount = -exchangeDiff
				}
				
				entries = append(entries, entry)
			}
		}
	}
	
	journal.JournalEntries = entries
	
	// Calculate totals
	totalDebit := 0.0
	totalCredit := 0.0
	for _, entry := range entries {
		totalDebit += entry.DebitAmount
		totalCredit += entry.CreditAmount
	}
	
	journal.TotalDebit = totalDebit
	journal.TotalCredit = totalCredit
	
	return tx.Create(journal).Error
}

func (s *CashBankService) createDepositJournalEntries(tx *gorm.DB, transaction *models.CashBankTransaction, account *models.CashBank, request DepositRequest, userID uint) error {
	// Simple deposit journal - would need more details for proper categorization
	return nil
}

func (s *CashBankService) createWithdrawalJournalEntries(tx *gorm.DB, transaction *models.CashBankTransaction, account *models.CashBank, request WithdrawalRequest, userID uint) error {
	// Simple withdrawal journal - would need more details for proper categorization
	return nil
}

// DTOs and Models

// CustomDate for handling date-only formats from frontend
type CustomDate time.Time

// UnmarshalJSON handles multiple date formats
func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "null" || s == "" {
		return nil
	}
	
	// Try multiple date formats
	formats := []string{
		"2006-01-02",           // YYYY-MM-DD from frontend
		"2006-01-02T15:04:05Z", // Full ISO format
		"2006-01-02T15:04:05Z07:00", // RFC3339
		"2006-01-02 15:04:05",  // MySQL datetime
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			*cd = CustomDate(t)
			return nil
		}
	}
	
	return fmt.Errorf("cannot parse date: %s", s)
}

// MarshalJSON converts to JSON
func (cd CustomDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(cd).Format("2006-01-02"))
}

// ToTime converts to time.Time
func (cd CustomDate) ToTime() time.Time {
	return time.Time(cd)
}

type CashBankCreateRequest struct {
	Name           string     `json:"name" binding:"required"`
	Type           string     `json:"type" binding:"required,oneof=CASH BANK"`
	AccountID      uint       `json:"account_id"`
	BankName       string     `json:"bank_name"`
	AccountNo      string     `json:"account_no"`
	Currency       string     `json:"currency"`
	OpeningBalance float64    `json:"opening_balance"`
	OpeningDate    CustomDate `json:"opening_date"`
	Description    string     `json:"description"`
}

type CashBankUpdateRequest struct {
	Name        string `json:"name"`
	BankName    string `json:"bank_name"`
	AccountNo   string `json:"account_no"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type TransferRequest struct {
	FromAccountID uint       `json:"from_account_id" binding:"required"`
	ToAccountID   uint       `json:"to_account_id" binding:"required"`
	Date          CustomDate `json:"date" binding:"required"`
	Amount        float64    `json:"amount" binding:"required,min=0"`
	ExchangeRate  float64    `json:"exchange_rate"`
	Reference     string     `json:"reference"`
	Notes         string     `json:"notes"`
}

type DepositRequest struct {
	AccountID uint       `json:"account_id" binding:"required"`
	Date      CustomDate `json:"date" binding:"required"`
	Amount    float64    `json:"amount" binding:"required,min=0"`
	Reference string     `json:"reference"`
	Notes     string     `json:"notes"`
}

type WithdrawalRequest struct {
	AccountID uint       `json:"account_id" binding:"required"`
	Date      CustomDate `json:"date" binding:"required"`
	Amount    float64    `json:"amount" binding:"required,min=0"`
	Reference string     `json:"reference"`
	Notes     string     `json:"notes"`
}

type ReconciliationRequest struct {
	Date             CustomDate                  `json:"date" binding:"required"`
	StatementBalance float64                     `json:"statement_balance" binding:"required"`
	Items            []ReconciliationItemRequest `json:"items"`
}

type ReconciliationItemRequest struct {
	TransactionID uint   `json:"transaction_id"`
	IsCleared     bool   `json:"is_cleared"`
	Notes         string `json:"notes"`
}

type TransactionFilter struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Type      string    `json:"type"`
	Page      int       `json:"page"`
	Limit     int       `json:"limit"`
}

type TransactionResult struct {
	Data       []models.CashBankTransaction `json:"data"`
	Total      int64                        `json:"total"`
	Page       int                          `json:"page"`
	Limit      int                          `json:"limit"`
	TotalPages int                          `json:"total_pages"`
}

type BalanceSummary struct {
	TotalCash    float64                `json:"total_cash"`
	TotalBank    float64                `json:"total_bank"`
	TotalBalance float64                `json:"total_balance"`
	ByAccount    []AccountBalance       `json:"by_account"`
	ByCurrency   map[string]float64     `json:"by_currency"`
}

type AccountBalance struct {
	AccountID   uint    `json:"account_id"`
	AccountName string  `json:"account_name"`
	AccountType string  `json:"account_type"`
	Balance     float64 `json:"balance"`
	Currency    string  `json:"currency"`
}

type CashBankTransfer struct {
	ID              uint      `gorm:"primaryKey"`
	TransferNumber  string    `gorm:"unique;not null;size:50"`
	FromAccountID   uint      `gorm:"not null;index"`
	ToAccountID     uint      `gorm:"not null;index"`
	Date            time.Time
	Amount          float64   `gorm:"type:decimal(15,2)"`
	ExchangeRate    float64   `gorm:"type:decimal(12,6);default:1"`
	ConvertedAmount float64   `gorm:"type:decimal(15,2)"`
	Reference       string    `gorm:"size:100"`
	Notes           string    `gorm:"type:text"`
	Status          string    `gorm:"size:20"`
	UserID          uint      `gorm:"not null;index"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type BankReconciliation struct {
	ID               uint      `gorm:"primaryKey"`
	CashBankID       uint      `gorm:"not null;index"`
	ReconcileDate    time.Time
	StatementBalance float64   `gorm:"type:decimal(15,2)"`
	SystemBalance    float64   `gorm:"type:decimal(15,2)"`
	Difference       float64   `gorm:"type:decimal(15,2)"`
	Status           string    `gorm:"size:20"`
	UserID           uint      `gorm:"not null;index"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ReconciliationItem struct {
	ID               uint   `gorm:"primaryKey"`
	ReconciliationID uint   `gorm:"not null;index"`
	TransactionID    uint   `gorm:"not null;index"`
	IsCleared        bool
	Notes            string `gorm:"type:text"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
