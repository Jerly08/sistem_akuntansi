package services

import (
	"fmt"
	"time"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

type InvoiceService struct {
	db              *gorm.DB
	settingsService *SettingsService
}

func NewInvoiceService(db *gorm.DB) *InvoiceService {
	settingsService := NewSettingsService(db)
	return &InvoiceService{
		db:              db,
		settingsService: settingsService,
	}
}

// GenerateInvoiceNumber generates next invoice number using settings
func (s *InvoiceService) GenerateInvoiceNumber() (string, error) {
	// Get settings
	_, err := s.settingsService.GetSettings()
	if err != nil {
		return "", fmt.Errorf("failed to get settings: %v", err)
	}

	var invoiceNumber string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Lock settings for update to prevent race conditions
		var settingsForUpdate models.Settings
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&settingsForUpdate).Error; err != nil {
			return err
		}

		// Generate invoice number using current next number
		invoiceNumber = fmt.Sprintf("%s-%05d", settingsForUpdate.InvoicePrefix, settingsForUpdate.InvoiceNextNumber)
		
		// Increment next number for future use
		settingsForUpdate.InvoiceNextNumber++
		
		// Save updated settings
		if err := tx.Save(&settingsForUpdate).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate invoice number: %v", err)
	}

	return invoiceNumber, nil
}

// GenerateQuoteNumber generates next quote number using settings
func (s *InvoiceService) GenerateQuoteNumber() (string, error) {
	// Get settings
	_, err := s.settingsService.GetSettings()
	if err != nil {
		return "", fmt.Errorf("failed to get settings: %v", err)
	}

	var quoteNumber string
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Lock settings for update to prevent race conditions
		var settingsForUpdate models.Settings
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&settingsForUpdate).Error; err != nil {
			return err
		}

		// Generate quote number using current next number
		quoteNumber = fmt.Sprintf("%s-%05d", settingsForUpdate.QuotePrefix, settingsForUpdate.QuoteNextNumber)
		
		// Increment next number for future use
		settingsForUpdate.QuoteNextNumber++
		
		// Save updated settings
		if err := tx.Save(&settingsForUpdate).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate quote number: %v", err)
	}

	return quoteNumber, nil
}

// FormatCurrency formats amount according to system settings
func (s *InvoiceService) FormatCurrency(amount float64) (string, error) {
	settings, err := s.settingsService.GetSettings()
	if err != nil {
		return "", err
	}

	// Format with decimal places from settings
	formatStr := fmt.Sprintf("%%.%df", settings.DecimalPlaces)
	formatted := fmt.Sprintf(formatStr, amount)
	
	// Apply thousand separator (simplified implementation)
	// In production, you might want to use a more sophisticated number formatting library
	
	return fmt.Sprintf("%s %s", settings.Currency, formatted), nil
}

// FormatDate formats date according to system settings
func (s *InvoiceService) FormatDate(date time.Time) (string, error) {
	settings, err := s.settingsService.GetSettings()
	if err != nil {
		return "", err
	}

	switch settings.DateFormat {
	case "DD/MM/YYYY":
		return date.Format("02/01/2006"), nil
	case "MM/DD/YYYY":
		return date.Format("01/02/2006"), nil
	case "DD-MM-YYYY":
		return date.Format("02-01-2006"), nil
	case "YYYY-MM-DD":
		return date.Format("2006-01-02"), nil
	default:
		return date.Format("02/01/2006"), nil // Default to DD/MM/YYYY
	}
}