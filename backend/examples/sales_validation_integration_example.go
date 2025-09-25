package services

// EXAMPLE: How to integrate balance validation into existing services

import (
	"fmt"
	"log"
)

// Enhanced Sales Service with Balance Validation
type EnhancedSalesService struct {
	db                *gorm.DB
	balanceValidator *BalanceValidationService
	// ... other dependencies
}

// CreateSaleWithValidation creates sale and validates accounting balance
func (s *EnhancedSalesService) CreateSaleWithValidation(saleRequest *SaleRequest) (*Sale, error) {
	// 1. Create sale transaction
	sale, err := s.createSaleTransaction(saleRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create sale: %v", err)
	}
	
	// 2. VALIDATE BALANCE after sale creation
	if err := s.balanceValidator.ValidateAfterTransaction(sale.ID, "SALE_CREATION"); err != nil {
		log.Printf("⚠️ Balance validation failed after sale creation: %v", err)
		// Option 1: Rollback transaction
		// s.rollbackSale(sale.ID)
		// return nil, err
		
		// Option 2: Log warning but continue (for now)
		log.Printf("🔄 Continuing with sale creation despite balance warning...")
	}
	
	return sale, nil
}

// RecordPaymentWithValidation records payment and validates balance
func (s *EnhancedSalesService) RecordPaymentWithValidation(paymentRequest *PaymentRequest) (*Payment, error) {
	// 1. Record payment transaction
	payment, err := s.recordPaymentTransaction(paymentRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to record payment: %v", err)
	}
	
	// 2. VALIDATE BALANCE after payment
	if err := s.balanceValidator.ValidateAfterTransaction(payment.ID, "PAYMENT_RECORDING"); err != nil {
		log.Printf("🚨 CRITICAL: Balance validation failed after payment: %v", err)
		
		// For payments, this is more critical - consider rollback
		// Get detailed validation report for debugging
		report, _ := s.balanceValidator.GetDetailedValidationReport()
		log.Printf("📋 Detailed validation report: %+v", report)
		
		// Could implement automatic correction or alert admin
		s.alertAdministrator("Balance validation failed after payment", payment.ID, err)
	}
	
	return payment, nil
}

// PREVENTION MECHANISMS:

// 1. Daily Balance Check (runs automatically)
func (s *EnhancedSalesService) RunDailyBalanceCheck() {
	validation, err := s.balanceValidator.ValidateRealTimeBalance()
	if err != nil {
		log.Printf("❌ Daily balance check failed: %v", err)
		return
	}
	
	if !validation.IsValid {
		log.Printf("🚨 DAILY ALERT: Accounting equation not balanced!")
		log.Printf("   Assets: %.2f", validation.TotalAssets)
		log.Printf("   Liabilities + Equity: %.2f", validation.TotalLiabilities + validation.AdjustedEquity)
		log.Printf("   Difference: %.2f", validation.BalanceDiff)
		
		// Send alert to admin
		s.alertAdministrator("Daily balance check failed", 0, fmt.Errorf("Balance difference: %.2f", validation.BalanceDiff))
	} else {
		log.Printf("✅ Daily balance check passed - accounting equation balanced")
	}
}

// 2. Period-End Closing Entries
func (s *EnhancedSalesService) RunPeriodEndClosingEntries(periodEndDate string) error {
	// Move net income to retained earnings
	validation, err := s.balanceValidator.ValidateRealTimeBalance()
	if err != nil {
		return fmt.Errorf("failed to get current balance: %v", err)
	}
	
	if validation.NetIncome != 0 {
		log.Printf("🔄 Creating closing entries for Net Income: %.2f", validation.NetIncome)
		
		// Create closing journal entry
		closingEntry := JournalEntry{
			Date:        periodEndDate,
			Reference:   fmt.Sprintf("CLOSING-%s", periodEndDate),
			Description: "Period-end closing entries",
			Lines: []JournalLine{
				{
					AccountCode: "4101", // Revenue
					DebitAmount: validation.NetIncome > 0 ? validation.NetIncome : 0,
					CreditAmount: validation.NetIncome < 0 ? -validation.NetIncome : 0,
				},
				{
					AccountCode: "3201", // Retained Earnings
					DebitAmount: validation.NetIncome < 0 ? -validation.NetIncome : 0,
					CreditAmount: validation.NetIncome > 0 ? validation.NetIncome : 0,
				},
			},
		}
		
		// Post closing entry
		if err := s.postJournalEntry(&closingEntry); err != nil {
			return fmt.Errorf("failed to post closing entry: %v", err)
		}
		
		// Validate after closing
		if err := s.balanceValidator.ValidateAfterTransaction(closingEntry.ID, "CLOSING_ENTRY"); err != nil {
			log.Printf("⚠️ Balance validation warning after closing entry: %v", err)
		}
		
		log.Printf("✅ Period-end closing entries completed successfully")
	}
	
	return nil
}

// 3. Admin Alert System
func (s *EnhancedSalesService) alertAdministrator(message string, transactionID uint, err error) {
	// Could send email, SMS, Slack notification, etc.
	log.Printf("🚨 ADMIN ALERT: %s (Transaction: %d) - Error: %v", message, transactionID, err)
	
	// Example: Store alert in database for admin dashboard
	alert := AdminAlert{
		Type:          "BALANCE_VALIDATION_FAILED",
		Message:       message,
		TransactionID: transactionID,
		ErrorDetails:  err.Error(),
		CreatedAt:     time.Now(),
		Resolved:      false,
	}
	
	s.db.Create(&alert)
}

// 4. Automated Fix Suggestions
func (s *EnhancedSalesService) SuggestBalanceFixes() ([]string, error) {
	report, err := s.balanceValidator.GetDetailedValidationReport()
	if err != nil {
		return nil, err
	}
	
	validation := report["validation_summary"].(*ValidationResult)
	
	suggestions := []string{}
	
	if !validation.IsValid {
		if validation.BalanceDiff > 0 {
			suggestions = append(suggestions, "Assets exceed L+E: Check for missing liabilities or journal entry errors")
		} else {
			suggestions = append(suggestions, "L+E exceed Assets: Check for missing assets or double-posted entries")
		}
	}
	
	if validation.NetIncome != 0 {
		suggestions = append(suggestions, fmt.Sprintf("Run period-end closing to move Net Income (%.2f) to Retained Earnings", validation.NetIncome))
	}
	
	return suggestions, nil
}