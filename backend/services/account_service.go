package services

import (
	"context"
	"fmt"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/utils"
	"strconv"
	"strings"
)

// AccountService handles account business logic
type AccountService interface {
	CreateAccount(ctx context.Context, req *models.AccountCreateRequest) (*models.Account, error)
	UpdateAccount(ctx context.Context, code string, req *models.AccountUpdateRequest) (*models.Account, error)
	DeleteAccount(ctx context.Context, code string) error
	GetAccount(ctx context.Context, code string) (*models.Account, error)
	ListAccounts(ctx context.Context, accountType string) ([]models.Account, error)
	GetAccountHierarchy(ctx context.Context) ([]models.Account, error)
	GetBalanceSummary(ctx context.Context) ([]models.AccountSummaryResponse, error)
	BulkImportAccounts(ctx context.Context, accounts []models.AccountImportRequest) error
	GenerateAccountCode(ctx context.Context, accountType, parentCode string) (string, error)
	ValidateAccountHierarchy(ctx context.Context, parentID *uint, accountType string) error
}

// AccountServiceImpl implements AccountService
type AccountServiceImpl struct {
	accountRepo repositories.AccountRepository
}

// NewAccountService creates a new account service
func NewAccountService(accountRepo repositories.AccountRepository) AccountService {
	return &AccountServiceImpl{
		accountRepo: accountRepo,
	}
}

// CreateAccount creates a new account with validation
func (s *AccountServiceImpl) CreateAccount(ctx context.Context, req *models.AccountCreateRequest) (*models.Account, error) {
	// Validate account type
	if !models.IsValidAccountType(string(req.Type)) {
		return nil, utils.NewValidationError("Invalid account type", map[string]string{
			"type": "Must be one of: ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE",
		})
	}

	// Validate parent hierarchy if parent is specified
	if req.ParentID != nil {
		if err := s.ValidateAccountHierarchy(ctx, req.ParentID, string(req.Type)); err != nil {
			return nil, err
		}
	}

	// Generate account code if not provided
	if req.Code == "" {
		var parentCode string
		if req.ParentID != nil {
			parent, err := s.accountRepo.FindByID(ctx, *req.ParentID)
			if err != nil {
				return nil, err
			}
			parentCode = parent.Code
		}
		
		code, err := s.GenerateAccountCode(ctx, string(req.Type), parentCode)
		if err != nil {
			return nil, err
		}
		req.Code = code
	}

	return s.accountRepo.Create(ctx, req)
}

// UpdateAccount updates an account
func (s *AccountServiceImpl) UpdateAccount(ctx context.Context, code string, req *models.AccountUpdateRequest) (*models.Account, error) {
	// Check if account exists
	existingAccount, err := s.accountRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Prevent deactivating accounts with children or transactions
	if req.IsActive != nil && !*req.IsActive {
		// Check for child accounts
		if len(existingAccount.Children) > 0 {
			return nil, utils.NewBadRequestError("Cannot deactivate account that has child accounts")
		}

		// Check for transactions (this would need to be implemented based on your transaction model)
		// For now, we'll skip this check
	}

	return s.accountRepo.Update(ctx, code, req)
}

// DeleteAccount deletes an account with validation
func (s *AccountServiceImpl) DeleteAccount(ctx context.Context, code string) error {
	return s.accountRepo.Delete(ctx, code)
}

// GetAccount gets a single account
func (s *AccountServiceImpl) GetAccount(ctx context.Context, code string) (*models.Account, error) {
	return s.accountRepo.FindByCode(ctx, code)
}

// ListAccounts lists accounts with optional filtering
func (s *AccountServiceImpl) ListAccounts(ctx context.Context, accountType string) ([]models.Account, error) {
	if accountType != "" {
		return s.accountRepo.FindByType(ctx, accountType)
	}
	return s.accountRepo.FindAll(ctx)
}

// GetAccountHierarchy gets account hierarchy
func (s *AccountServiceImpl) GetAccountHierarchy(ctx context.Context) ([]models.Account, error) {
	return s.accountRepo.GetHierarchy(ctx)
}

// GetBalanceSummary gets balance summary
func (s *AccountServiceImpl) GetBalanceSummary(ctx context.Context) ([]models.AccountSummaryResponse, error) {
	return s.accountRepo.GetBalanceSummary(ctx)
}

// BulkImportAccounts imports multiple accounts with validation
func (s *AccountServiceImpl) BulkImportAccounts(ctx context.Context, accounts []models.AccountImportRequest) error {
	// Validate all accounts before importing
	codeMap := make(map[string]bool)
	
	for i, account := range accounts {
		// Check for duplicate codes in the import
		if codeMap[account.Code] {
			return utils.NewBadRequestError(fmt.Sprintf("Duplicate account code in import: %s at row %d", account.Code, i+1))
		}
		codeMap[account.Code] = true

		// Validate account type
		if !models.IsValidAccountType(string(account.Type)) {
			return utils.NewValidationError(fmt.Sprintf("Invalid account type at row %d", i+1), map[string]string{
				"type": "Must be one of: ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE",
			})
		}
	}

	return s.accountRepo.BulkImport(ctx, accounts)
}

// GenerateAccountCode generates account code based on type and parent
func (s *AccountServiceImpl) GenerateAccountCode(ctx context.Context, accountType, parentCode string) (string, error) {
	var prefix string
	
	// Define account type prefixes
	switch accountType {
	case models.AccountTypeAsset:
		prefix = "1"
	case models.AccountTypeLiability:
		prefix = "2"
	case models.AccountTypeEquity:
		prefix = "3"
	case models.AccountTypeRevenue:
		prefix = "4"
	case models.AccountTypeExpense:
		prefix = "5"
	default:
		return "", utils.NewValidationError("Invalid account type for code generation", nil)
	}

	// If parent code exists, use it as base
	baseCode := prefix
	if parentCode != "" {
		baseCode = parentCode
	}

	// Find next available code
	accounts, err := s.accountRepo.FindByType(ctx, accountType)
	if err != nil {
		return "", err
	}

	// Find the highest existing code number for this prefix
	maxNumber := 0
	for _, account := range accounts {
		if strings.HasPrefix(account.Code, baseCode) {
			// Extract the number part
			numberPart := strings.TrimPrefix(account.Code, baseCode)
			if len(numberPart) > 0 {
				if num, err := strconv.Atoi(numberPart); err == nil && num > maxNumber {
					maxNumber = num
				}
			}
		}
	}

	// Generate next code
	nextNumber := maxNumber + 1
	return fmt.Sprintf("%s%02d", baseCode, nextNumber), nil
}

// ValidateAccountHierarchy validates parent-child account relationships
func (s *AccountServiceImpl) ValidateAccountHierarchy(ctx context.Context, parentID *uint, accountType string) error {
	if parentID == nil {
		return nil
	}

	parent, err := s.accountRepo.FindByID(ctx, *parentID)
	if err != nil {
		return err
	}

	// Validate that parent and child have compatible types
	// Asset accounts can have asset children
	// Liability accounts can have liability children
	// etc.
	if parent.Type != accountType {
		return utils.NewValidationError("Parent and child accounts must be of the same type", map[string]string{
			"parent_type": parent.Type,
			"child_type":  accountType,
		})
	}

	// Prevent creating deep hierarchies (max 4 levels)
	if parent.Level >= 4 {
		return utils.NewValidationError("Maximum account hierarchy depth exceeded", map[string]string{
			"max_depth": "4",
			"parent_level": fmt.Sprintf("%d", parent.Level),
		})
	}

	return nil
}
