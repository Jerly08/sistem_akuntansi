package services

import (
	"errors"
	"fmt"
	"log"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

type SettingsService struct {
	db *gorm.DB
}

// NewSettingsService creates a new instance of SettingsService
func NewSettingsService(db *gorm.DB) *SettingsService {
	return &SettingsService{db: db}
}

// GetSettings retrieves the system settings
func (s *SettingsService) GetSettings() (*models.Settings, error) {
	var settings models.Settings
	
	// First, try to get existing settings
	err := s.db.First(&settings).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no settings exist, create default settings
			settings = s.createDefaultSettings()
			if err := s.db.Create(&settings).Error; err != nil {
				log.Printf("Error creating default settings: %v", err)
				return nil, err
			}
			return &settings, nil
		}
		log.Printf("Error fetching settings: %v", err)
		return nil, err
	}
	
	return &settings, nil
}

// UpdateSettings updates the system settings
func (s *SettingsService) UpdateSettings(updates map[string]interface{}, userID uint) error {
	var settings models.Settings
	
	// Get existing settings or create if not exists
	err := s.db.First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			settings = s.createDefaultSettings()
			if err := s.db.Create(&settings).Error; err != nil {
				log.Printf("Error creating settings: %v", err)
				return err
			}
		} else {
			log.Printf("Error fetching settings: %v", err)
			return err
		}
	}
	
	// Add the user who is updating
	updates["updated_by"] = userID
	
	// Validate certain fields before updating
	if err := s.validateSettings(updates); err != nil {
		return err
	}
	
	// Update settings
	if err := s.db.Model(&settings).Updates(updates).Error; err != nil {
		log.Printf("Error updating settings: %v", err)
		return err
	}
	
	return nil
}

// validateSettings validates the settings before saving
func (s *SettingsService) validateSettings(updates map[string]interface{}) error {
	// Validate email if present
	if email, ok := updates["company_email"].(string); ok {
		if email != "" && !isValidEmail(email) {
			return errors.New("invalid email format")
		}
	}
	
	// Validate tax rate if present
	if taxRate, ok := updates["default_tax_rate"].(float64); ok {
		if taxRate < 0 || taxRate > 100 {
			return errors.New("tax rate must be between 0 and 100")
		}
	}
	
	// Validate decimal places
	if decimalPlaces, ok := updates["decimal_places"].(int); ok {
		if decimalPlaces < 0 || decimalPlaces > 4 {
			return errors.New("decimal places must be between 0 and 4")
		}
	}
	
	// Validate language
	if language, ok := updates["language"].(string); ok {
		if language != "id" && language != "en" {
			return errors.New("language must be 'id' or 'en'")
		}
	}
	
	
	return nil
}

// createDefaultSettings creates default system settings
func (s *SettingsService) createDefaultSettings() models.Settings {
	return models.Settings{
		CompanyName:        "PT. Sistem Akuntansi Indonesia",
		CompanyAddress:     "Jl. Sudirman Kav. 45-46, Jakarta Pusat 10210, Indonesia",
		CompanyPhone:       "+62-21-5551234",
		CompanyEmail:       "info@sistemakuntansi.co.id",
		Currency:           "IDR",
		DateFormat:         "DD/MM/YYYY",
		FiscalYearStart:    "January 1",
		Language:           "id",
		Timezone:           "Asia/Jakarta",
		ThousandSeparator:  ".",
		DecimalSeparator:   ",",
		DecimalPlaces:      2,
		DefaultTaxRate:     11.0,
		InvoicePrefix:      "INV",
		InvoiceNextNumber:  1,
		QuotePrefix:        "QT",
		QuoteNextNumber:    1,
		PurchasePrefix:     "PO",
		PurchaseNextNumber: 1,
	}
}

// GetNextInvoiceNumber gets and increments the next invoice number
func (s *SettingsService) GetNextInvoiceNumber() (string, error) {
	var settings models.Settings
	
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Lock the settings row for update
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&settings).Error; err != nil {
		tx.Rollback()
		return "", err
	}
	
	// Get current number and format
	invoiceNumber := formatInvoiceNumber(settings.InvoicePrefix, settings.InvoiceNextNumber)
	
	// Increment the number
	settings.InvoiceNextNumber++
	if err := tx.Save(&settings).Error; err != nil {
		tx.Rollback()
		return "", err
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return "", err
	}
	
	return invoiceNumber, nil
}

// GetNextQuoteNumber gets and increments the next quote number
func (s *SettingsService) GetNextQuoteNumber() (string, error) {
	var settings models.Settings
	
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&settings).Error; err != nil {
		tx.Rollback()
		return "", err
	}
	
	quoteNumber := formatInvoiceNumber(settings.QuotePrefix, settings.QuoteNextNumber)
	settings.QuoteNextNumber++
	
	if err := tx.Save(&settings).Error; err != nil {
		tx.Rollback()
		return "", err
	}
	
	if err := tx.Commit().Error; err != nil {
		return "", err
	}
	
	return quoteNumber, nil
}

// GetNextPurchaseNumber gets and increments the next purchase order number
func (s *SettingsService) GetNextPurchaseNumber() (string, error) {
	var settings models.Settings
	
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&settings).Error; err != nil {
		tx.Rollback()
		return "", err
	}
	
	purchaseNumber := formatInvoiceNumber(settings.PurchasePrefix, settings.PurchaseNextNumber)
	settings.PurchaseNextNumber++
	
	if err := tx.Save(&settings).Error; err != nil {
		tx.Rollback()
		return "", err
	}
	
	if err := tx.Commit().Error; err != nil {
		return "", err
	}
	
	return purchaseNumber, nil
}

// Helper functions

func isValidEmail(email string) bool {
	// Simple email validation
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	
	atIndex := -1
	for i, char := range email {
		if char == '@' {
			if atIndex != -1 {
				return false // Multiple @ symbols
			}
			atIndex = i
		}
	}
	
	if atIndex <= 0 || atIndex >= len(email)-1 {
		return false
	}
	
	return true
}

func formatInvoiceNumber(prefix string, number int) string {
	// Format: PREFIX-YYYY-MM-00001
	// You can customize this format as needed
	return prefix + "-" + formatNumberWithPadding(number, 5)
}

func formatNumberWithPadding(number int, padding int) string {
	format := "%0" + string(rune(padding+'0')) + "d"
	return fmt.Sprintf(format, number)
}
