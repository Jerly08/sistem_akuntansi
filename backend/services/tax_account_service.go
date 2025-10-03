package services

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// TaxAccountService handles tax account settings management
type TaxAccountService struct {
	db       *gorm.DB
	cache    *models.TaxAccountSettings
	cacheMux sync.RWMutex
	cacheLoaded bool
}

// NewTaxAccountService creates a new TaxAccountService instance
func NewTaxAccountService(db *gorm.DB) *TaxAccountService {
	service := &TaxAccountService{
		db: db,
	}
	
	// Load current settings on initialization
	if err := service.loadSettings(); err != nil {
		log.Printf("Warning: Failed to load tax account settings: %v", err)
	}
	
	return service
}

// loadSettings loads the current tax account settings into cache
func (s *TaxAccountService) loadSettings() error {
	s.cacheMux.Lock()
	defer s.cacheMux.Unlock()

	var settings models.TaxAccountSettings
	err := s.db.Preload("SalesReceivableAccount").
		Preload("SalesCashAccount").
		Preload("SalesBankAccount").
		Preload("SalesRevenueAccount").
		Preload("SalesOutputVATAccount").
		Preload("PurchasePayableAccount").
		Preload("PurchaseCashAccount").
		Preload("PurchaseBankAccount").
		Preload("PurchaseInputVATAccount").
		Preload("PurchaseExpenseAccount").
		Preload("WithholdingTax21Account").
		Preload("WithholdingTax23Account").
		Preload("WithholdingTax25Account").
		Preload("TaxPayableAccount").
		Preload("InventoryAccount").
		Preload("COGSAccount").
		Preload("UpdatedByUser").
		Where("is_active = ?", true).
		First(&settings).Error

	if err == nil {
		s.cache = &settings
		s.cacheLoaded = true
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create default settings if none exist
		defaultSettings := models.GetDefaultTaxAccountSettings()
		defaultSettings.UpdatedBy = 1 // System user ID
		
		if createErr := s.db.Create(defaultSettings).Error; createErr != nil {
			return fmt.Errorf("failed to create default settings: %v", createErr)
		}
		
		// Load the created settings with relations
		return s.loadSettings()
	}

	return fmt.Errorf("failed to load tax account settings: %v", err)
}

// GetSettings returns the current tax account settings
func (s *TaxAccountService) GetSettings() (*models.TaxAccountSettings, error) {
	s.cacheMux.RLock()
	defer s.cacheMux.RUnlock()

	if !s.cacheLoaded || s.cache == nil {
		// Reload if cache is empty
		s.cacheMux.RUnlock()
		if err := s.loadSettings(); err != nil {
			s.cacheMux.RLock()
			return nil, err
		}
		s.cacheMux.RLock()
	}

	return s.cache, nil
}

// CreateSettings creates new tax account settings
func (s *TaxAccountService) CreateSettings(req *models.TaxAccountSettingsCreateRequest, userID uint) (*models.TaxAccountSettings, error) {
	// Validate that accounts exist
	if err := s.validateAccounts(req); err != nil {
		return nil, err
	}

	// Deactivate existing active settings
	if err := s.db.Model(&models.TaxAccountSettings{}).
		Where("is_active = ?", true).
		Update("is_active", false).Error; err != nil {
		return nil, fmt.Errorf("failed to deactivate existing settings: %v", err)
	}

	// Create new settings
	settings := &models.TaxAccountSettings{
		SalesReceivableAccountID:   req.SalesReceivableAccountID,
		SalesCashAccountID:         req.SalesCashAccountID,
		SalesBankAccountID:         req.SalesBankAccountID,
		SalesRevenueAccountID:      req.SalesRevenueAccountID,
		SalesOutputVATAccountID:    req.SalesOutputVATAccountID,
		PurchasePayableAccountID:   req.PurchasePayableAccountID,
		PurchaseCashAccountID:      req.PurchaseCashAccountID,
		PurchaseBankAccountID:      req.PurchaseBankAccountID,
		PurchaseInputVATAccountID:  req.PurchaseInputVATAccountID,
		PurchaseExpenseAccountID:   req.PurchaseExpenseAccountID,
		WithholdingTax21AccountID:  req.WithholdingTax21AccountID,
		WithholdingTax23AccountID:  req.WithholdingTax23AccountID,
		WithholdingTax25AccountID:  req.WithholdingTax25AccountID,
		TaxPayableAccountID:        req.TaxPayableAccountID,
		InventoryAccountID:         req.InventoryAccountID,
		COGSAccountID:              req.COGSAccountID,
		IsActive:                   true,
		ApplyToAllCompanies:        req.ApplyToAllCompanies != nil && *req.ApplyToAllCompanies,
		UpdatedBy:                  userID,
		Notes:                      req.Notes,
	}

	// Validate settings
	if err := settings.ValidateAccountSettings(); err != nil {
		return nil, err
	}

	if err := s.db.Create(settings).Error; err != nil {
		return nil, fmt.Errorf("failed to create tax account settings: %v", err)
	}

	// Reload cache
	if err := s.loadSettings(); err != nil {
		log.Printf("Warning: Failed to reload cache after creating settings: %v", err)
	}

	// Load relations for response
	err := s.db.Preload("SalesReceivableAccount").
		Preload("SalesCashAccount").
		Preload("SalesBankAccount").
		Preload("SalesRevenueAccount").
		Preload("SalesOutputVATAccount").
		Preload("PurchasePayableAccount").
		Preload("PurchaseCashAccount").
		Preload("PurchaseBankAccount").
		Preload("PurchaseInputVATAccount").
		Preload("PurchaseExpenseAccount").
		Preload("WithholdingTax21Account").
		Preload("WithholdingTax23Account").
		Preload("WithholdingTax25Account").
		Preload("TaxPayableAccount").
		Preload("InventoryAccount").
		Preload("COGSAccount").
		Preload("UpdatedByUser").
		First(settings, settings.ID).Error

	return settings, err
}

// UpdateSettings updates existing tax account settings
func (s *TaxAccountService) UpdateSettings(id uint, req *models.TaxAccountSettingsUpdateRequest, userID uint) (*models.TaxAccountSettings, error) {
	var settings models.TaxAccountSettings
	if err := s.db.First(&settings, id).Error; err != nil {
		return nil, fmt.Errorf("tax account settings not found: %v", err)
	}

	// Update fields if provided
	if req.SalesReceivableAccountID != nil {
		settings.SalesReceivableAccountID = *req.SalesReceivableAccountID
	}
	if req.SalesCashAccountID != nil {
		settings.SalesCashAccountID = *req.SalesCashAccountID
	}
	if req.SalesBankAccountID != nil {
		settings.SalesBankAccountID = *req.SalesBankAccountID
	}
	if req.SalesRevenueAccountID != nil {
		settings.SalesRevenueAccountID = *req.SalesRevenueAccountID
	}
	if req.SalesOutputVATAccountID != nil {
		settings.SalesOutputVATAccountID = *req.SalesOutputVATAccountID
	}
	if req.PurchasePayableAccountID != nil {
		settings.PurchasePayableAccountID = *req.PurchasePayableAccountID
	}
	if req.PurchaseCashAccountID != nil {
		settings.PurchaseCashAccountID = *req.PurchaseCashAccountID
	}
	if req.PurchaseBankAccountID != nil {
		settings.PurchaseBankAccountID = *req.PurchaseBankAccountID
	}
	if req.PurchaseInputVATAccountID != nil {
		settings.PurchaseInputVATAccountID = *req.PurchaseInputVATAccountID
	}
	if req.PurchaseExpenseAccountID != nil {
		settings.PurchaseExpenseAccountID = *req.PurchaseExpenseAccountID
	}
	if req.WithholdingTax21AccountID != nil {
		settings.WithholdingTax21AccountID = req.WithholdingTax21AccountID
	}
	if req.WithholdingTax23AccountID != nil {
		settings.WithholdingTax23AccountID = req.WithholdingTax23AccountID
	}
	if req.WithholdingTax25AccountID != nil {
		settings.WithholdingTax25AccountID = req.WithholdingTax25AccountID
	}
	if req.TaxPayableAccountID != nil {
		settings.TaxPayableAccountID = req.TaxPayableAccountID
	}
	if req.InventoryAccountID != nil {
		settings.InventoryAccountID = req.InventoryAccountID
	}
	if req.COGSAccountID != nil {
		settings.COGSAccountID = req.COGSAccountID
	}
	if req.IsActive != nil {
		settings.IsActive = *req.IsActive
	}
	if req.ApplyToAllCompanies != nil {
		settings.ApplyToAllCompanies = *req.ApplyToAllCompanies
	}
	if req.Notes != nil {
		settings.Notes = *req.Notes
	}

	settings.UpdatedBy = userID
	settings.UpdatedAt = time.Now()

	// Validate settings
	if err := settings.ValidateAccountSettings(); err != nil {
		return nil, err
	}

	if err := s.db.Save(&settings).Error; err != nil {
		return nil, fmt.Errorf("failed to update tax account settings: %v", err)
	}

	// Reload cache
	if err := s.loadSettings(); err != nil {
		log.Printf("Warning: Failed to reload cache after updating settings: %v", err)
	}

	// Load relations for response
	err := s.db.Preload("SalesReceivableAccount").
		Preload("SalesCashAccount").
		Preload("SalesBankAccount").
		Preload("SalesRevenueAccount").
		Preload("SalesOutputVATAccount").
		Preload("PurchasePayableAccount").
		Preload("PurchaseCashAccount").
		Preload("PurchaseBankAccount").
		Preload("PurchaseInputVATAccount").
		Preload("PurchaseExpenseAccount").
		Preload("WithholdingTax21Account").
		Preload("WithholdingTax23Account").
		Preload("WithholdingTax25Account").
		Preload("TaxPayableAccount").
		Preload("InventoryAccount").
		Preload("COGSAccount").
		Preload("UpdatedByUser").
		First(&settings, settings.ID).Error

	return &settings, err
}

// validateAccounts validates that all specified accounts exist and are active
func (s *TaxAccountService) validateAccounts(req *models.TaxAccountSettingsCreateRequest) error {
	accountIDs := []uint{
		req.SalesReceivableAccountID,
		req.SalesCashAccountID,
		req.SalesBankAccountID,
		req.SalesRevenueAccountID,
		req.SalesOutputVATAccountID,
		req.PurchasePayableAccountID,
		req.PurchaseCashAccountID,
		req.PurchaseBankAccountID,
		req.PurchaseInputVATAccountID,
		req.PurchaseExpenseAccountID,
	}

	// Add optional account IDs if provided
	if req.WithholdingTax21AccountID != nil {
		accountIDs = append(accountIDs, *req.WithholdingTax21AccountID)
	}
	if req.WithholdingTax23AccountID != nil {
		accountIDs = append(accountIDs, *req.WithholdingTax23AccountID)
	}
	if req.WithholdingTax25AccountID != nil {
		accountIDs = append(accountIDs, *req.WithholdingTax25AccountID)
	}
	if req.TaxPayableAccountID != nil {
		accountIDs = append(accountIDs, *req.TaxPayableAccountID)
	}
	if req.InventoryAccountID != nil {
		accountIDs = append(accountIDs, *req.InventoryAccountID)
	}
	if req.COGSAccountID != nil {
		accountIDs = append(accountIDs, *req.COGSAccountID)
	}

	var count int64
	if err := s.db.Model(&models.Account{}).
		Where("id IN ? AND is_active = ?", accountIDs, true).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to validate accounts: %v", err)
	}

	if int(count) != len(accountIDs) {
		return fmt.Errorf("some accounts are not found or inactive")
	}

	return nil
}

// GetAccountID returns the account ID for a specific account type
func (s *TaxAccountService) GetAccountID(accountType string) (uint, error) {
	settings, err := s.GetSettings()
	if err != nil {
		return 0, err
	}

	switch accountType {
	// Sales accounts
	case "sales_receivable":
		return settings.SalesReceivableAccountID, nil
	case "sales_cash":
		return settings.SalesCashAccountID, nil
	case "sales_bank":
		return settings.SalesBankAccountID, nil
	case "sales_revenue":
		return settings.SalesRevenueAccountID, nil
	case "sales_output_vat":
		return settings.SalesOutputVATAccountID, nil

	// Purchase accounts
	case "purchase_payable":
		return settings.PurchasePayableAccountID, nil
	case "purchase_cash":
		return settings.PurchaseCashAccountID, nil
	case "purchase_bank":
		return settings.PurchaseBankAccountID, nil
	case "purchase_input_vat":
		return settings.PurchaseInputVATAccountID, nil
	case "purchase_expense":
		return settings.PurchaseExpenseAccountID, nil

	// Tax accounts
	case "withholding_tax21":
		if settings.WithholdingTax21AccountID != nil {
			return *settings.WithholdingTax21AccountID, nil
		}
		return 0, fmt.Errorf("withholding tax 21 account not configured")
	case "withholding_tax23":
		if settings.WithholdingTax23AccountID != nil {
			return *settings.WithholdingTax23AccountID, nil
		}
		return 0, fmt.Errorf("withholding tax 23 account not configured")
	case "withholding_tax25":
		if settings.WithholdingTax25AccountID != nil {
			return *settings.WithholdingTax25AccountID, nil
		}
		return 0, fmt.Errorf("withholding tax 25 account not configured")
	case "tax_payable":
		if settings.TaxPayableAccountID != nil {
			return *settings.TaxPayableAccountID, nil
		}
		return 0, fmt.Errorf("tax payable account not configured")

	// Inventory accounts
	case "inventory":
		if settings.InventoryAccountID != nil {
			return *settings.InventoryAccountID, nil
		}
		return 0, fmt.Errorf("inventory account not configured")
	case "cogs":
		if settings.COGSAccountID != nil {
			return *settings.COGSAccountID, nil
		}
		return 0, fmt.Errorf("COGS account not configured")

	default:
		return 0, fmt.Errorf("unknown account type: %s", accountType)
	}
}

// GetAccountByCode returns account ID by account code (fallback method)
func (s *TaxAccountService) GetAccountByCode(code string) (uint, error) {
	var account models.Account
	err := s.db.Where("code = ? AND is_active = ?", code, true).First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("account with code %s not found", code)
		}
		return 0, fmt.Errorf("failed to find account: %v", err)
	}
	return account.ID, nil
}

// RefreshCache forces a reload of the settings cache
func (s *TaxAccountService) RefreshCache() error {
	return s.loadSettings()
}

// GetAllSettings returns all tax account settings (for admin purposes)
func (s *TaxAccountService) GetAllSettings() ([]models.TaxAccountSettings, error) {
	var settings []models.TaxAccountSettings
	err := s.db.Preload("SalesReceivableAccount").
		Preload("SalesCashAccount").
		Preload("SalesBankAccount").
		Preload("SalesRevenueAccount").
		Preload("SalesOutputVATAccount").
		Preload("PurchasePayableAccount").
		Preload("PurchaseCashAccount").
		Preload("PurchaseBankAccount").
		Preload("PurchaseInputVATAccount").
		Preload("PurchaseExpenseAccount").
		Preload("WithholdingTax21Account").
		Preload("WithholdingTax23Account").
		Preload("WithholdingTax25Account").
		Preload("TaxPayableAccount").
		Preload("InventoryAccount").
		Preload("COGSAccount").
		Preload("UpdatedByUser").
		Order("created_at DESC").
		Find(&settings).Error

	return settings, err
}

// ActivateSettings activates specific tax account settings by ID
func (s *TaxAccountService) ActivateSettings(id uint, userID uint) error {
	// Deactivate all other settings
	if err := s.db.Model(&models.TaxAccountSettings{}).
		Where("id != ?", id).
		Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate other settings: %v", err)
	}

	// Activate the specified settings
	if err := s.db.Model(&models.TaxAccountSettings{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":  true,
			"updated_by": userID,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to activate settings: %v", err)
	}

	// Reload cache
	return s.loadSettings()
}