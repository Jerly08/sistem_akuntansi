package repositories

import (
	"context"
	"errors"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/utils"
	"gorm.io/gorm"
)

// AccountRepository defines account-related database operations
type AccountRepository interface {
	Create(ctx context.Context, req *models.AccountCreateRequest) (*models.Account, error)
	Update(ctx context.Context, code string, req *models.AccountUpdateRequest) (*models.Account, error)
	Delete(ctx context.Context, code string) error
	FindByCode(ctx context.Context, code string) (*models.Account, error)
	FindByID(ctx context.Context, id uint) (*models.Account, error)
	FindAll(ctx context.Context) ([]models.Account, error)
	FindByType(ctx context.Context, accountType string) ([]models.Account, error)
	GetHierarchy(ctx context.Context) ([]models.Account, error)
	BulkImport(ctx context.Context, accounts []models.AccountImportRequest) error
	CalculateBalance(ctx context.Context, accountID uint) (float64, error)
	UpdateBalance(ctx context.Context, accountID uint, debitAmount, creditAmount float64) error
	GetBalanceSummary(ctx context.Context) ([]models.AccountSummaryResponse, error)
}

// AccountRepo implements AccountRepository
type AccountRepo struct {
	*BaseRepo
}

// NewAccountRepository creates a new account repository
func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &AccountRepo{
		BaseRepo: &BaseRepo{DB: db},
	}
}

// Create creates a new account
func (r *AccountRepo) Create(ctx context.Context, req *models.AccountCreateRequest) (*models.Account, error) {
	// Validate account type
	if !models.IsValidAccountType(string(req.Type)) {
		return nil, utils.NewValidationError("Invalid account type", map[string]string{
			"type": "Must be one of: ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE",
		})
	}

	// Check if code already exists
	var existingAccount models.Account
	if err := r.DB.WithContext(ctx).Where("code = ?", req.Code).First(&existingAccount).Error; err == nil {
		return nil, utils.NewConflictError("Account code already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.NewDatabaseError("check existing code", err)
	}

	// Calculate level if parent exists
	level := 1
	if req.ParentID != nil {
		var parent models.Account
		if err := r.DB.WithContext(ctx).First(&parent, *req.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, utils.NewNotFoundError("Parent account")
			}
			return nil, utils.NewDatabaseError("find parent account", err)
		}
		level = parent.Level + 1
	}

	account := &models.Account{
		Code:        req.Code,
		Name:        req.Name,
		Type:        string(req.Type),
		Category:    req.Category,
		ParentID:    req.ParentID,
		Level:       level,
		Description: req.Description,
		Balance:     req.OpeningBalance,
		IsActive:    true,
		IsHeader:    false,
	}

	if err := r.DB.WithContext(ctx).Create(account).Error; err != nil {
		return nil, utils.NewDatabaseError("create account", err)
	}

	return account, nil
}

// Update updates an account
func (r *AccountRepo) Update(ctx context.Context, code string, req *models.AccountUpdateRequest) (*models.Account, error) {
	var account models.Account
	if err := r.DB.WithContext(ctx).Where("code = ?", code).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFoundError("Account")
		}
		return nil, utils.NewDatabaseError("find account", err)
	}

	// Update fields
	account.Name = req.Name
	account.Description = req.Description
	account.Category = req.Category
	if req.IsActive != nil {
		account.IsActive = *req.IsActive
	}

	if err := r.DB.WithContext(ctx).Save(&account).Error; err != nil {
		return nil, utils.NewDatabaseError("update account", err)
	}

	return &account, nil
}

// Delete deletes an account
func (r *AccountRepo) Delete(ctx context.Context, code string) error {
	var account models.Account
	if err := r.DB.WithContext(ctx).Where("code = ?", code).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.NewNotFoundError("Account")
		}
		return utils.NewDatabaseError("find account", err)
	}

	// Check if account has children
	var childrenCount int64
	if err := r.DB.WithContext(ctx).Model(&models.Account{}).Where("parent_id = ?", account.ID).Count(&childrenCount).Error; err != nil {
		return utils.NewDatabaseError("count children", err)
	}

	if childrenCount > 0 {
		return utils.NewBadRequestError("Cannot delete account that has child accounts")
	}

	// Check if account has transactions
	var transactionCount int64
	if err := r.DB.WithContext(ctx).Model(&models.Transaction{}).Where("account_id = ?", account.ID).Count(&transactionCount).Error; err != nil {
		return utils.NewDatabaseError("count transactions", err)
	}

	if transactionCount > 0 {
		return utils.NewBadRequestError("Cannot delete account that has transactions")
	}

	if err := r.DB.WithContext(ctx).Delete(&account).Error; err != nil {
		return utils.NewDatabaseError("delete account", err)
	}

	return nil
}

// FindByCode finds account by code
func (r *AccountRepo) FindByCode(ctx context.Context, code string) (*models.Account, error) {
	var account models.Account
	if err := r.DB.WithContext(ctx).Preload("Parent").Preload("Children").Where("code = ?", code).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFoundError("Account")
		}
		return nil, utils.NewDatabaseError("find account", err)
	}
	return &account, nil
}

// FindByID finds account by ID
func (r *AccountRepo) FindByID(ctx context.Context, id uint) (*models.Account, error) {
	var account models.Account
	if err := r.DB.WithContext(ctx).Preload("Parent").Preload("Children").First(&account, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewNotFoundError("Account")
		}
		return nil, utils.NewDatabaseError("find account", err)
	}
	return &account, nil
}

// FindAll finds all accounts
func (r *AccountRepo) FindAll(ctx context.Context) ([]models.Account, error) {
	var accounts []models.Account
	if err := r.DB.WithContext(ctx).Preload("Parent").Order("code").Find(&accounts).Error; err != nil {
		return nil, utils.NewDatabaseError("find all accounts", err)
	}
	return accounts, nil
}

// FindByType finds accounts by type
func (r *AccountRepo) FindByType(ctx context.Context, accountType string) ([]models.Account, error) {
	var accounts []models.Account
	if err := r.DB.WithContext(ctx).Where("type = ?", accountType).Order("code").Find(&accounts).Error; err != nil {
		return nil, utils.NewDatabaseError("find accounts by type", err)
	}
	return accounts, nil
}

// GetHierarchy gets account hierarchy
func (r *AccountRepo) GetHierarchy(ctx context.Context) ([]models.Account, error) {
	var accounts []models.Account
	if err := r.DB.WithContext(ctx).Preload("Children").Where("parent_id IS NULL").Order("code").Find(&accounts).Error; err != nil {
		return nil, utils.NewDatabaseError("get account hierarchy", err)
	}
	return accounts, nil
}

// BulkImport imports multiple accounts
func (r *AccountRepo) BulkImport(ctx context.Context, accounts []models.AccountImportRequest) error {
	tx := r.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, req := range accounts {
		var parentID *uint
		if req.ParentCode != "" {
			var parent models.Account
			if err := tx.Where("code = ?", req.ParentCode).First(&parent).Error; err != nil {
				tx.Rollback()
				return utils.NewBadRequestError("Parent account not found: " + req.ParentCode)
			}
			parentID = &parent.ID
		}

		level := 1
		if parentID != nil {
			var parent models.Account
			if err := tx.First(&parent, *parentID).Error; err != nil {
				tx.Rollback()
				return utils.NewDatabaseError("find parent for level calculation", err)
			}
			level = parent.Level + 1
		}

		account := models.Account{
			Code:        req.Code,
			Name:        req.Name,
			Type:        string(req.Type),
			Category:    req.Category,
			ParentID:    parentID,
			Level:       level,
			Description: req.Description,
			Balance:     req.OpeningBalance,
			IsActive:    true,
			IsHeader:    false,
		}

		if err := tx.Create(&account).Error; err != nil {
			tx.Rollback()
			return utils.NewDatabaseError("create account in bulk import", err)
		}
	}

	return tx.Commit().Error
}

// CalculateBalance calculates account balance
func (r *AccountRepo) CalculateBalance(ctx context.Context, accountID uint) (float64, error) {
	var result struct {
		DebitSum  float64
		CreditSum float64
	}

	if err := r.DB.WithContext(ctx).Model(&models.Transaction{}).
		Select("COALESCE(SUM(debit_amount), 0) as debit_sum, COALESCE(SUM(credit_amount), 0) as credit_sum").
		Where("account_id = ?", accountID).
		Scan(&result).Error; err != nil {
		return 0, utils.NewDatabaseError("calculate balance", err)
	}

	// Get account to determine if it's a debit or credit account
	var account models.Account
	if err := r.DB.WithContext(ctx).First(&account, accountID).Error; err != nil {
		return 0, utils.NewDatabaseError("find account for balance calculation", err)
	}

	// Calculate balance based on account type
	if account.Type == models.AccountTypeAsset || account.Type == models.AccountTypeExpense {
		return result.DebitSum - result.CreditSum, nil
	}
	return result.CreditSum - result.DebitSum, nil
}

// UpdateBalance updates account balance
func (r *AccountRepo) UpdateBalance(ctx context.Context, accountID uint, debitAmount, creditAmount float64) error {
	balance, err := r.CalculateBalance(ctx, accountID)
	if err != nil {
		return err
	}

	if err := r.DB.WithContext(ctx).Model(&models.Account{}).
		Where("id = ?", accountID).
		Update("balance", balance).Error; err != nil {
		return utils.NewDatabaseError("update balance", err)
	}

	return nil
}

// GetBalanceSummary gets balance summary by account type
func (r *AccountRepo) GetBalanceSummary(ctx context.Context) ([]models.AccountSummaryResponse, error) {
	var summaries []models.AccountSummaryResponse

	accountTypes := []string{
		models.AccountTypeAsset,
		models.AccountTypeLiability,
		models.AccountTypeEquity,
		models.AccountTypeRevenue,
		models.AccountTypeExpense,
	}

	for _, accountType := range accountTypes {
		var result struct {
			TotalAccounts  int64
			ActiveAccounts int64
			TotalBalance   float64
		}

		if err := r.DB.WithContext(ctx).Model(&models.Account{}).
			Select("COUNT(*) as total_accounts, SUM(CASE WHEN is_active THEN 1 ELSE 0 END) as active_accounts, COALESCE(SUM(balance), 0) as total_balance").
			Where("type = ?", accountType).
			Scan(&result).Error; err != nil {
			return nil, utils.NewDatabaseError("get balance summary", err)
		}

		summaries = append(summaries, models.AccountSummaryResponse{
			Type:           models.AccountType(accountType),
			TotalAccounts:  result.TotalAccounts,
			ActiveAccounts: result.ActiveAccounts,
			TotalBalance:   result.TotalBalance,
		})
	}

	return summaries, nil
}
