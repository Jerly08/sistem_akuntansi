package repositories

import (
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
	"math"
	"time"
)

type CashBankRepository struct {
	db *gorm.DB
}

func NewCashBankRepository(db *gorm.DB) *CashBankRepository {
	return &CashBankRepository{db: db}
}

// FindAll retrieves all cash and bank accounts
func (r *CashBankRepository) FindAll() ([]models.CashBank, error) {
	var accounts []models.CashBank
	err := r.db.Preload("Account").Where("is_active = ?", true).Find(&accounts).Error
	return accounts, err
}

// FindByID retrieves account by ID
func (r *CashBankRepository) FindByID(id uint) (*models.CashBank, error) {
	var account models.CashBank
	err := r.db.Preload("Account").First(&account, id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// FindByCode retrieves account by code
func (r *CashBankRepository) FindByCode(code string) (*models.CashBank, error) {
	var account models.CashBank
	err := r.db.Preload("Account").Where("code = ?", code).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// Create creates a new cash/bank account
func (r *CashBankRepository) Create(account *models.CashBank) (*models.CashBank, error) {
	if err := r.db.Create(account).Error; err != nil {
		return nil, err
	}
	
	return r.FindByID(account.ID)
}

// Update updates a cash/bank account
func (r *CashBankRepository) Update(account *models.CashBank) (*models.CashBank, error) {
	// Avoid overwriting immutable/linked fields like AccountID
	// Only update allowed columns explicitly
	updateTx := r.db.Model(&models.CashBank{}).
		Where("id = ?", account.ID).
		Select("Code", "Name", "Type", "BankName", "AccountNo", "Currency", "Balance", "MinBalance", "MaxBalance", "DailyLimit", "MonthlyLimit", "IsActive", "IsRestricted", "Description")

	if err := updateTx.Updates(account).Error; err != nil {
		return nil, err
	}

	return r.FindByID(account.ID)
}

// Delete soft deletes a cash/bank account
func (r *CashBankRepository) Delete(id uint) error {
	return r.db.Delete(&models.CashBank{}, id).Error
}

// GetTransactions retrieves transactions for an account
func (r *CashBankRepository) GetTransactions(accountID uint, filter TransactionFilter) (*TransactionResult, error) {
	query := r.db.Model(&models.CashBankTransaction{}).Where("cash_bank_id = ?", accountID)
	
	// Apply filters
	if !filter.StartDate.IsZero() {
		query = query.Where("transaction_date >= ?", filter.StartDate)
	}
	
	if !filter.EndDate.IsZero() {
		query = query.Where("transaction_date <= ?", filter.EndDate)
	}
	
	if filter.Type != "" {
		query = query.Where("reference_type = ?", filter.Type)
	}
	
	// Count total records
	var total int64
	query.Count(&total)
	
	// Apply pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	
	offset := (filter.Page - 1) * filter.Limit
	
	// Get paginated results
	var transactions []models.CashBankTransaction
	err := query.Order("transaction_date DESC, id DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&transactions).Error
	
	if err != nil {
		return nil, err
	}
	
	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))
	
	return &TransactionResult{
		Data:       transactions,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

// CreateTransaction creates a new transaction record
func (r *CashBankRepository) CreateTransaction(tx *models.CashBankTransaction) error {
	return r.db.Create(tx).Error
}

// GetBalanceSummary gets summary of all account balances directly from cash_banks table
// This fixes the issue where COA (accounts table) was not synchronized with cash_banks balance
// IMPORTANT: Now respects soft delete (deleted_at IS NULL) to match FindAll() behavior
func (r *CashBankRepository) GetBalanceSummary() (*BalanceSummary, error) {
	var cashTotal, bankTotal float64
	
	// Get balance directly from cash_banks table for CASH accounts (exclude soft deleted)
	r.db.Table("cash_banks").
		Where("type = ? AND is_active = ? AND deleted_at IS NULL", models.CashBankTypeCash, true).
		Select("COALESCE(SUM(balance), 0)").
		Scan(&cashTotal)
	
	// Get balance directly from cash_banks table for BANK accounts (exclude soft deleted)
	r.db.Table("cash_banks").
		Where("type = ? AND is_active = ? AND deleted_at IS NULL", models.CashBankTypeBank, true).
		Select("COALESCE(SUM(balance), 0)").
		Scan(&bankTotal)
	
	// Get by currency from cash_banks table (exclude soft deleted)
	var currencySums []struct {
		Currency string
		Total    float64
	}
	
	r.db.Table("cash_banks").
		Where("is_active = ? AND deleted_at IS NULL", true).
		Select("currency, SUM(balance) as total").
		Group("currency").
		Scan(&currencySums)
	
	byCurrency := make(map[string]float64)
	for _, cs := range currencySums {
		byCurrency[cs.Currency] = cs.Total
	}
	
	return &BalanceSummary{
		TotalCash:    cashTotal,
		TotalBank:    bankTotal,
		TotalBalance: cashTotal + bankTotal,
		ByCurrency:   byCurrency,
	}, nil
}

// CountByType counts accounts by type for code generation
func (r *CashBankRepository) CountByType(accountType string) (int64, error) {
	var count int64
	err := r.db.Model(&models.CashBank{}).
		Where("type = ?", accountType).
		Count(&count).Error
	
	return count, err
}

// GetCashAccounts retrieves all cash accounts
func (r *CashBankRepository) GetCashAccounts() ([]models.CashBank, error) {
	var accounts []models.CashBank
	err := r.db.Where("type = ? AND is_active = ?", models.CashBankTypeCash, true).
		Find(&accounts).Error
	return accounts, err
}

// GetBankAccounts retrieves all bank accounts
func (r *CashBankRepository) GetBankAccounts() ([]models.CashBank, error) {
	var accounts []models.CashBank
	err := r.db.Where("type = ? AND is_active = ?", models.CashBankTypeBank, true).
		Find(&accounts).Error
	return accounts, err
}

// UpdateBalance updates account balance
func (r *CashBankRepository) UpdateBalance(id uint, amount float64) error {
	return r.db.Model(&models.CashBank{}).
		Where("id = ?", id).
		Update("balance", gorm.Expr("balance + ?", amount)).Error
}

// GetMonthlyTransactionSummary gets monthly transaction summary
func (r *CashBankRepository) GetMonthlyTransactionSummary(accountID uint, year, month int) (*MonthlyTransactionSummary, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
	
	summary := &MonthlyTransactionSummary{
		Month: startDate.Format("2006-01"),
	}
	
	// Get opening balance (last transaction before start date)
	var lastTx models.CashBankTransaction
	r.db.Where("cash_bank_id = ? AND transaction_date < ?", accountID, startDate).
		Order("transaction_date DESC, id DESC").
		First(&lastTx)
	
	if lastTx.ID > 0 {
		summary.OpeningBalance = lastTx.BalanceAfter
	}
	
	// Sum deposits
	r.db.Model(&models.CashBankTransaction{}).
		Where("cash_bank_id = ? AND transaction_date BETWEEN ? AND ? AND amount > 0", 
			accountID, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&summary.TotalDeposits)
	
	// Sum withdrawals
	r.db.Model(&models.CashBankTransaction{}).
		Where("cash_bank_id = ? AND transaction_date BETWEEN ? AND ? AND amount < 0", 
			accountID, startDate, endDate).
		Select("COALESCE(SUM(ABS(amount)), 0)").
		Scan(&summary.TotalWithdrawals)
	
	// Get closing balance
	var currentAccount models.CashBank
	r.db.First(&currentAccount, accountID)
	summary.ClosingBalance = currentAccount.Balance
	
	return summary, nil
}

// DTOs
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
	TotalCash    float64            `json:"total_cash"`
	TotalBank    float64            `json:"total_bank"`
	TotalBalance float64            `json:"total_balance"`
	ByCurrency   map[string]float64 `json:"by_currency"`
}

type MonthlyTransactionSummary struct {
	Month            string  `json:"month"`
	OpeningBalance   float64 `json:"opening_balance"`
	TotalDeposits    float64 `json:"total_deposits"`
	TotalWithdrawals float64 `json:"total_withdrawals"`
	ClosingBalance   float64 `json:"closing_balance"`
}
