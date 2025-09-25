package services

import (
	"errors"
	"fmt"
	"log"
	"math"
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"time"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

type SettingsService struct {
	db *gorm.DB
}

// Simple in-memory cache for settings (process-local)
var (
	settingsCache      *models.Settings
	settingsCacheAt    time.Time
	settingsCacheMutex sync.RWMutex
	settingsCacheTTL   = 5 * time.Minute
)

// NewSettingsService creates a new instance of SettingsService
func NewSettingsService(db *gorm.DB) *SettingsService {
	return &SettingsService{db: db}
}

// GetSettings retrieves the system settings (with caching)
func (s *SettingsService) GetSettings() (*models.Settings, error) {
	// Try cache first
	settingsCacheMutex.RLock()
	if settingsCache != nil && time.Since(settingsCacheAt) < settingsCacheTTL {
		defer settingsCacheMutex.RUnlock()
		return settingsCache, nil
	}
	settingsCacheMutex.RUnlock()

	var settings models.Settings
	// First, try to get existing settings
	err := s.db.First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no settings exist, create default settings (from accounting_config)
			settings = s.createDefaultSettings()
			if err := s.db.Create(&settings).Error; err != nil {
				log.Printf("Error creating default settings: %v", err)
				return nil, err
			}
		} else {
			log.Printf("Error fetching settings: %v", err)
			return nil, err
		}
	}

	// Update cache
	settingsCacheMutex.Lock()
	settingsCache = &settings
	settingsCacheAt = time.Now()
	settingsCacheMutex.Unlock()

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

	// Normalize fields (frontend-backend sync)
	if v, ok := updates["fiscal_year_start"].(string); ok && v != "" {
		updates["fiscal_year_start"] = s.normalizeFiscalYearStart(v)
	}

	// Capture old values for audit logging
	oldValues := s.captureCurrentValues(&settings, updates)
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
	// Invalidate cache after update
	settingsCacheMutex.Lock()
	settingsCache = nil
	settingsCacheMutex.Unlock()
	// Log changes to history
	go s.logSettingsChanges(settings.ID, oldValues, updates, userID, "UPDATE")
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
	// Validate date format
	if dateFormat, ok := updates["date_format"].(string); ok {
		validFormats := []string{"DD/MM/YYYY", "MM/DD/YYYY", "YYYY-MM-DD", "DD-MM-YYYY"}
		isValid := false
		for _, format := range validFormats {
			if dateFormat == format {
				isValid = true
				break
			}
		}
		if !isValid {
			return errors.New("invalid date format")
		}
	}
	// Validate currency
	if currency, ok := updates["currency"].(string); ok {
		validCurrencies := []string{"IDR", "USD", "EUR", "SGD", "MYR"}
		isValid := false
		for _, curr := range validCurrencies {
			if currency == curr {
				isValid = true
				break
			}
		}
		if !isValid {
			return errors.New("invalid currency code")
		}
	}
	// Validate prefixes (non-empty and length)
	prefixes := []string{"invoice_prefix", "quote_prefix", "purchase_prefix", "journal_prefix"}
	for _, prefix := range prefixes {
		if value, ok := updates[prefix].(string); ok {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("%s cannot be empty", prefix)
			}
			if len(value) > 10 {
				return fmt.Errorf("%s cannot be longer than 10 characters", prefix)
			}
		}
	}
	// Validate next numbers (must be positive)
	nextNumbers := []string{"invoice_next_number", "quote_next_number", "purchase_next_number", "journal_next_number"}
	for _, nextNum := range nextNumbers {
		if value, ok := updates[nextNum].(int); ok {
			if value < 1 {
				return fmt.Errorf("%s must be at least 1", nextNum)
			}
		}
	}
	return nil
}

// createDefaultSettings creates default system settings
func (s *SettingsService) createDefaultSettings() models.Settings {
	// Load unified config for defaults
	cfg := config.GetAccountingConfig()
	journalPrefix := "JE"
	requireApproval := false
	if cfg != nil {
		if cfg.JournalSettings.CodePrefix != "" { journalPrefix = cfg.JournalSettings.CodePrefix }
		requireApproval = cfg.JournalSettings.RequireApproval
	}
	return models.Settings{
		CompanyName:        "PT. Sistem Akuntansi Indonesia",
		CompanyAddress:     "Jl. Sudirman Kav. 45-46, Jakarta Pusat 10210, Indonesia",
		CompanyPhone:       "+62-21-5551234",
		CompanyEmail:       "info@sistemakuntansi.co.id",
		Currency:           ifThen(cfg != nil && cfg.CurrencySettings.BaseCurrency != "", cfg.CurrencySettings.BaseCurrency, "IDR"),
		DateFormat:         "DD/MM/YYYY",
		FiscalYearStart:    "January 1",
		Language:           "id",
		Timezone:           "Asia/Jakarta",
		ThousandSeparator:  ".",
		DecimalSeparator:   ",",
		DecimalPlaces:      ifThenInt(cfg != nil && cfg.CurrencySettings.DecimalPlaces != 0, cfg.CurrencySettings.DecimalPlaces, 2),
		DefaultTaxRate:     ifThenFloat(cfg != nil && cfg.TaxRates.DefaultPPN != 0, cfg.TaxRates.DefaultPPN, 11.0),
		InvoicePrefix:      "INV",
		InvoiceNextNumber:  1,
		QuotePrefix:        "QT",
		QuoteNextNumber:    1,
		PurchasePrefix:     "PO",
		PurchaseNextNumber: 1,
		JournalPrefix:          journalPrefix,
		JournalNextNumber:      1,
		RequireJournalApproval: requireApproval,
	}
}

// Utility inline helpers for default selection
func ifThen(cond bool, a string, b string) string { if cond { return a }; return b }
func ifThenInt(cond bool, a int, b int) int { if cond { return a }; return b }
func ifThenFloat(cond bool, a float64, b float64) float64 { if cond { return a }; return b }

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

// ResetToDefaults resets all settings to default values
func (s *SettingsService) ResetToDefaults(userID uint) error {
	var settings models.Settings
	
	// Get existing settings or create if not exists
	err := s.db.First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			settings = s.createDefaultSettings()
			if err := s.db.Create(&settings).Error; err != nil {
				log.Printf("Error creating default settings: %v", err)
				return err
			}
			return nil
		}
		log.Printf("Error fetching settings: %v", err)
		return err
	}
	
	// Reset to default values
	defaultSettings := s.createDefaultSettings()
	
	// Keep the ID and preserve existing next numbers to avoid conflicts
	defaultSettings.ID = settings.ID
	defaultSettings.CreatedAt = settings.CreatedAt
	defaultSettings.InvoiceNextNumber = settings.InvoiceNextNumber
	defaultSettings.QuoteNextNumber = settings.QuoteNextNumber
	defaultSettings.PurchaseNextNumber = settings.PurchaseNextNumber
	defaultSettings.UpdatedBy = userID
	
	// Save the reset settings
	if err := s.db.Save(&defaultSettings).Error; err != nil {
		log.Printf("Error resetting settings to defaults: %v", err)
		return err
	}
	
	log.Printf("Settings reset to defaults by user %d", userID)
return nil
}

// GetNextJournalNumberTx generates next journal number within a transaction (safe sequence)
func (s *SettingsService) GetNextJournalNumberTx(tx *gorm.DB) (string, error) {
	var settingsForUpdate models.Settings
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&settingsForUpdate).Error; err != nil {
		return "", err
	}
	code := fmt.Sprintf("%s-%05d", settingsForUpdate.JournalPrefix, settingsForUpdate.JournalNextNumber)
	settingsForUpdate.JournalNextNumber++
	if err := tx.Save(&settingsForUpdate).Error; err != nil {
		return "", err
	}
	// Invalidate cache since sequence changed
	settingsCacheMutex.Lock()
	settingsCache = nil
	settingsCacheMutex.Unlock()
	return code, nil
}

// normalizeFiscalYearStart normalizes various inputs to a standard format like "January 1"
func (s *SettingsService) normalizeFiscalYearStart(input string) string {
	in := strings.TrimSpace(strings.ToLower(input))
	// Accept formats: "mm/dd/yyyy", "mm/dd", "dd/mm/yyyy", month name + day
	months := []string{"january","february","march","april","may","june","july","august","september","october","november","december"}
for _, m := range months {
		if strings.Contains(in, m) {
			// e.g., "january 1" -> capitalize
			return strings.Title(m) + " " + extractDay(in)
		}
	}
	// Try MM/DD or DD/MM
	parts := strings.FieldsFunc(in, func(r rune) bool { return r == '/' || r == '-' })
	if len(parts) >= 2 {
		mm := parts[0]
		dd := parts[1]
		// If year exists, ignore
		monthIdx := toInt(mm)
		day := toInt(dd)
		if monthIdx >= 1 && monthIdx <= 12 && day >= 1 && day <= 31 {
			return strings.Title(months[monthIdx-1]) + fmt.Sprintf(" %d", day)
		}
	}
	// Fallback
	return "January 1"
}

func extractDay(s string) string {
	for i := 1; i <= 31; i++ {
		if strings.Contains(s, fmt.Sprintf("%d", i)) {
			return fmt.Sprintf("%d", i)
		}
	}
	return "1"
}

func toInt(s string) int {
	var n int
	for _, r := range s {
		if r < '0' || r > '9' { return -1 }
		n = n*10 + int(r-'0')
	}
	return n
}

// GetValidationRules returns validation rules for settings
func (s *SettingsService) GetValidationRules() map[string]interface{} {
	return map[string]interface{}{
		"date_formats":     []string{"DD/MM/YYYY", "MM/DD/YYYY", "YYYY-MM-DD", "DD-MM-YYYY"},
		"currencies":       []string{"IDR", "USD", "EUR", "SGD", "MYR"},
		"languages":        []string{"id", "en"},
		"tax_rate_range":   map[string]float64{"min": 0, "max": 100},
		"decimal_places_range": map[string]int{"min": 0, "max": 4},
		"prefix_max_length": 10,
		"min_next_number":  1,
	}
}

type SettingsHistoryResult struct {
	Data       []models.SettingsHistoryResponse `json:"data"`
	Total      int64                            `json:"total"`
	Page       int                              `json:"page"`
	Limit      int                              `json:"limit"`
	TotalPages int                              `json:"total_pages"`
}

// GetSettingsHistory retrieves paginated settings history
func (s *SettingsService) GetSettingsHistory(filter models.SettingsHistoryFilter) (*SettingsHistoryResult, error) {
	var history []models.SettingsHistory
	var total int64

	query := s.db.Model(&models.SettingsHistory{}).Preload("User")

	// Apply filters
	if filter.Field != "" {
		query = query.Where("field = ?", filter.Field)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.ChangedBy != "" {
		query = query.Where("changed_by = ?", filter.ChangedBy)
	}
	if filter.StartDate != "" {
		query = query.Where("created_at >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("created_at <= ?", filter.EndDate)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count settings history: %v", err)
	}

	// Apply pagination
	offset := (filter.Page - 1) * filter.Limit
	if err := query.Offset(offset).Limit(filter.Limit).Order("created_at DESC").Find(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve settings history: %v", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))

	// Convert to response format
	responses := make([]models.SettingsHistoryResponse, len(history))
	for i, h := range history {
		responses[i] = h.ToResponse()
	}

	return &SettingsHistoryResult{
		Data:       responses,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

// Helper methods for audit logging

// captureCurrentValues captures current values of fields that will be updated
func (s *SettingsService) captureCurrentValues(settings *models.Settings, updates map[string]interface{}) map[string]interface{} {
	oldValues := make(map[string]interface{})
	
	v := reflect.ValueOf(*settings)
	t := reflect.TypeOf(*settings)
	
	for key := range updates {
		// Skip meta fields
		if key == "updated_by" || key == "updated_at" {
			continue
		}
		
		// Find field in struct
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			
			// Parse json tag to get field name
			fieldName := jsonTag
			if commaIdx := len(jsonTag); commaIdx > 0 {
				if idx := len(jsonTag); idx > 0 {
					for j, char := range jsonTag {
						if char == ',' {
							fieldName = jsonTag[:j]
							break
						}
					}
				}
			}
			
			if fieldName == key {
				oldValues[key] = v.Field(i).Interface()
				break
			}
		}
	}
	
	return oldValues
}

// logSettingsChanges logs changes to settings_history table
func (s *SettingsService) logSettingsChanges(settingsID uint, oldValues, newValues map[string]interface{}, userID uint, action string) {
	for field, newValue := range newValues {
		// Skip meta fields
		if field == "updated_by" || field == "updated_at" {
			continue
		}
		
		oldValue := oldValues[field]
		
		// Skip if values are the same
		if reflect.DeepEqual(oldValue, newValue) {
			continue
		}
		
		// Convert values to JSON strings
		oldValueJSON, _ := json.Marshal(oldValue)
		newValueJSON, _ := json.Marshal(newValue)
		
		// Create history record
		history := models.SettingsHistory{
			SettingsID: settingsID,
			Field:      field,
			OldValue:   string(oldValueJSON),
			NewValue:   string(newValueJSON),
			Action:     action,
			ChangedBy:  userID,
		}
		
		// Save history record
		if err := s.db.Create(&history).Error; err != nil {
			log.Printf("Error logging settings change for field %s: %v", field, err)
		}
	}
}
