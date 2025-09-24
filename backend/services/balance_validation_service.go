package services

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

// BalanceValidationService validates accounting equation after each transaction
type BalanceValidationService struct {
	db *gorm.DB
}

// NewBalanceValidationService creates balance validation service
func NewBalanceValidationService(db *gorm.DB) *BalanceValidationService {
	return &BalanceValidationService{db: db}
}

// BalanceValidationResult represents balance validation result
type BalanceValidationResult struct {
	IsValid           bool      `json:"is_valid"`
	TotalAssets      float64   `json:"total_assets"`
	TotalLiabilities float64   `json:"total_liabilities"`
	TotalEquity      float64   `json:"total_equity"`
	NetIncome        float64   `json:"net_income"`
	AdjustedEquity   float64   `json:"adjusted_equity"`  // Equity + Net Income
	BalanceDiff      float64   `json:"balance_diff"`
	ValidationTime   time.Time `json:"validation_time"`
	Errors          []string   `json:"errors,omitempty"`
}

// ValidateRealTimeBalance validates accounting equation in real-time
func (s *BalanceValidationService) ValidateRealTimeBalance() (*BalanceValidationResult, error) {
	result := &BalanceValidationResult{
		ValidationTime: time.Now(),
		Errors:        []string{},
	}
	
	// Get current balances for all account types
	var assets, liabilities, equity, revenue, expenses float64
	
	// Assets (should be positive)
	if err := s.db.Raw(`
		SELECT COALESCE(SUM(balance), 0) 
		FROM accounts 
		WHERE type = 'ASSET' AND is_active = 1
	`).Scan(&assets).Error; err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get assets: %v", err))
		return result, err
	}
	
	// Liabilities (should be positive)
	if err := s.db.Raw(`
		SELECT COALESCE(SUM(balance), 0) 
		FROM accounts 
		WHERE type = 'LIABILITY' AND is_active = 1
	`).Scan(&liabilities).Error; err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get liabilities: %v", err))
		return result, err
	}
	
	// Equity (should be positive)  
	if err := s.db.Raw(`
		SELECT COALESCE(SUM(balance), 0) 
		FROM accounts 
		WHERE type = 'EQUITY' AND is_active = 1
	`).Scan(&equity).Error; err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get equity: %v", err))
		return result, err
	}
	
	// Revenue (should be positive)
	if err := s.db.Raw(`
		SELECT COALESCE(SUM(balance), 0) 
		FROM accounts 
		WHERE type = 'REVENUE' AND is_active = 1
	`).Scan(&revenue).Error; err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get revenue: %v", err))
		return result, err
	}
	
	// Expenses (should be positive)
	if err := s.db.Raw(`
		SELECT COALESCE(SUM(balance), 0) 
		FROM accounts 
		WHERE type = 'EXPENSE' AND is_active = 1
	`).Scan(&expenses).Error; err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get expenses: %v", err))
		return result, err
	}
	
	// Calculate Net Income
	netIncome := revenue - expenses
	
	// Calculate Adjusted Equity (includes current period net income)
	adjustedEquity := equity + netIncome
	
	// Check accounting equation: Assets = Liabilities + (Equity + Net Income)
	tolerance := 0.01 // 1 cent tolerance
	balanceDiff := assets - (liabilities + adjustedEquity)
	isValid := (balanceDiff >= -tolerance && balanceDiff <= tolerance)
	
	result.IsValid = isValid
	result.TotalAssets = assets
	result.TotalLiabilities = liabilities
	result.TotalEquity = equity
	result.NetIncome = netIncome
	result.AdjustedEquity = adjustedEquity
	result.BalanceDiff = balanceDiff
	
	// Add validation warnings
	if !isValid {
		result.Errors = append(result.Errors, 
			fmt.Sprintf("Accounting equation not balanced: Assets (%.2f) != Liabilities + Equity + Net Income (%.2f). Difference: %.2f", 
				assets, liabilities + adjustedEquity, balanceDiff))
	}
	
	if netIncome < 0 {
		result.Errors = append(result.Errors, 
			fmt.Sprintf("Warning: Net Loss detected: %.2f (Revenue: %.2f, Expenses: %.2f)", 
				netIncome, revenue, expenses))
	}
	
	return result, nil
}

// ValidateAfterTransaction validates balance after a specific transaction
func (s *BalanceValidationService) ValidateAfterTransaction(transactionID uint, transactionType string) error {
	validation, err := s.ValidateRealTimeBalance()
	if err != nil {
		return fmt.Errorf("validation failed after %s (ID: %d): %v", transactionType, transactionID, err)
	}
	
	if !validation.IsValid {
		return fmt.Errorf("accounting equation violated after %s (ID: %d): %s", 
			transactionType, transactionID, validation.Errors[0])
	}
	
	// Log successful validation
	fmt.Printf("✅ Balance validation passed after %s (ID: %d): Assets=%.2f, L+E=%.2f\n", 
		transactionType, transactionID, validation.TotalAssets, validation.TotalLiabilities + validation.AdjustedEquity)
	
	return nil
}

// GetDetailedValidationReport provides detailed validation report
func (s *BalanceValidationService) GetDetailedValidationReport() (map[string]interface{}, error) {
	validation, err := s.ValidateRealTimeBalance()
	if err != nil {
		return nil, err
	}
	
	// Get account details for troubleshooting
	var accountDetails []struct {
		AccountCode string  `json:"account_code"`
		AccountName string  `json:"account_name"`
		AccountType string  `json:"account_type"`
		Balance     float64 `json:"balance"`
	}
	
	s.db.Raw(`
		SELECT code as account_code, name as account_name, type as account_type, balance
		FROM accounts 
		WHERE is_active = 1 AND balance != 0
		ORDER BY type, code
	`).Scan(&accountDetails)
	
	return map[string]interface{}{
		"validation_summary": validation,
		"account_details":   accountDetails,
		"recommendations":   s.getRecommendations(validation),
	}, nil
}

// getRecommendations provides fix recommendations based on validation result
func (s *BalanceValidationService) getRecommendations(validation *BalanceValidationResult) []string {
	recommendations := []string{}
	
	if !validation.IsValid {
		if validation.BalanceDiff > 0 {
			recommendations = append(recommendations, 
				"Assets exceed Liabilities + Equity. Check for missing liabilities or understated equity.")
		} else {
			recommendations = append(recommendations, 
				"Liabilities + Equity exceed Assets. Check for missing assets or overstated liabilities.")
		}
	}
	
	if validation.NetIncome != 0 {
		recommendations = append(recommendations, 
			"Consider running period-end closing entries to move net income to retained earnings.")
	}
	
	return recommendations
}