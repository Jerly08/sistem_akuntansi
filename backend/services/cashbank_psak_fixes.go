package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

// PSAKCashBankService provides PSAK-compliant cash/bank account code generation
type PSAKCashBankService struct {
	db          *gorm.DB
	accountRepo repositories.AccountRepository
}

// NewPSAKCashBankService creates a new PSAK-compliant service
func NewPSAKCashBankService(db *gorm.DB, accountRepo repositories.AccountRepository) *PSAKCashBankService {
	return &PSAKCashBankService{
		db:          db,
		accountRepo: accountRepo,
	}
}

// generatePSAKCompliantAccountCode generates account code following PSAK standards
func (s *PSAKCashBankService) generatePSAKCompliantAccountCode(cashBankType string, accountName string) (string, error) {
	var parentCode string
	var accountPrefix string
	
	// Determine parent account based on cash/bank type and name
	switch cashBankType {
	case models.CashBankTypeCash:
		parentCode = "1101" // KAS
		accountPrefix = "1101"
	case models.CashBankTypeBank:
		// Determine bank type based on name
		bankName := strings.ToUpper(accountName)
		switch {
		case strings.Contains(bankName, "BCA"):
			parentCode = "1102"
			accountPrefix = "1102"
		case strings.Contains(bankName, "MANDIRI"):
			parentCode = "1103" 
			accountPrefix = "1103"
		case strings.Contains(bankName, "UOB"):
			parentCode = "1104"
			accountPrefix = "1104"
		case strings.Contains(bankName, "BRI"):
			parentCode = "1105"
			accountPrefix = "1105"
		case strings.Contains(bankName, "BNI"):
			parentCode = "1106"
			accountPrefix = "1106"
		default:
			// For unknown banks, use generic bank parent
			parentCode = "1110" // BANK LAIN-LAIN
			accountPrefix = "1110"
		}
	default:
		return "", fmt.Errorf("unsupported cash bank type: %s", cashBankType)
	}
	
	// Ensure parent account exists
	if err := s.ensureParentAccountExists(parentCode, cashBankType); err != nil {
		return "", fmt.Errorf("failed to ensure parent account exists: %w", err)
	}
	
	// Generate next child code
	childCode, err := s.generateNextChildCode(accountPrefix)
	if err != nil {
		return "", fmt.Errorf("failed to generate child code: %w", err)
	}
	
	return childCode, nil
}

// ensureParentAccountExists creates parent account if it doesn't exist
func (s *PSAKCashBankService) ensureParentAccountExists(parentCode, cashBankType string) error {
	// Check if parent exists
	_, err := s.accountRepo.FindByCode(context.Background(), parentCode)
	if err == nil {
		return nil // Parent exists
	}
	
	// Create parent account based on code
	var parentAccount *models.Account
	
	switch parentCode {
	case "1100":
		// Main cash/bank header
		parentAccount = &models.Account{
			Code:        "1100",
			Name:        "KAS DAN BANK",
			Type:        models.AccountTypeAsset,
			Category:    models.CategoryCurrentAsset,
			Level:       2,
			IsHeader:    true,
			IsActive:    true,
			Description: "Auto-created parent account for cash and bank accounts",
		}
		// Set parent to 1000 (ASSETS) if exists
		if assetsParent, err := s.accountRepo.FindByCode(context.Background(), "1000"); err == nil {
			parentAccount.ParentID = &assetsParent.ID
		}
		
	case "1101":
		// Cash parent
		parentAccount = &models.Account{
			Code:        "1101",
			Name:        "KAS",
			Type:        models.AccountTypeAsset,
			Category:    models.CategoryCurrentAsset,
			Level:       3,
			IsHeader:    false, // Will be set to true when first child is added
			IsActive:    true,
			Description: "Auto-created cash parent account",
		}
		// Set parent to 1100 
		if cashBankParent, err := s.accountRepo.FindByCode(context.Background(), "1100"); err == nil {
			parentAccount.ParentID = &cashBankParent.ID
		}
		
	case "1102":
		// BCA parent
		parentAccount = &models.Account{
			Code:        "1102",
			Name:        "BANK BCA",
			Type:        models.AccountTypeAsset,
			Category:    models.CategoryCurrentAsset,
			Level:       3,
			IsHeader:    false,
			IsActive:    true,
			Description: "Auto-created Bank BCA parent account",
		}
		if cashBankParent, err := s.accountRepo.FindByCode(context.Background(), "1100"); err == nil {
			parentAccount.ParentID = &cashBankParent.ID
		}
		
	case "1103":
		// Mandiri parent
		parentAccount = &models.Account{
			Code:        "1103",
			Name:        "BANK MANDIRI",
			Type:        models.AccountTypeAsset,
			Category:    models.CategoryCurrentAsset,
			Level:       3,
			IsHeader:    false,
			IsActive:    true,
			Description: "Auto-created Bank Mandiri parent account",
		}
		if cashBankParent, err := s.accountRepo.FindByCode(context.Background(), "1100"); err == nil {
			parentAccount.ParentID = &cashBankParent.ID
		}
		
	case "1104":
		// UOB parent
		parentAccount = &models.Account{
			Code:        "1104",
			Name:        "BANK UOB",
			Type:        models.AccountTypeAsset,
			Category:    models.CategoryCurrentAsset,
			Level:       3,
			IsHeader:    false,
			IsActive:    true,
			Description: "Auto-created Bank UOB parent account",
		}
		if cashBankParent, err := s.accountRepo.FindByCode(context.Background(), "1100"); err == nil {
			parentAccount.ParentID = &cashBankParent.ID
		}
		
	case "1105":
		// BRI parent
		parentAccount = &models.Account{
			Code:        "1105",
			Name:        "BANK BRI",
			Type:        models.AccountTypeAsset,
			Category:    models.CategoryCurrentAsset,
			Level:       3,
			IsHeader:    false,
			IsActive:    true,
			Description: "Auto-created Bank BRI parent account",
		}
		if cashBankParent, err := s.accountRepo.FindByCode(context.Background(), "1100"); err == nil {
			parentAccount.ParentID = &cashBankParent.ID
		}
		
	case "1106":
		// BNI parent
		parentAccount = &models.Account{
			Code:        "1106",
			Name:        "BANK BNI",
			Type:        models.AccountTypeAsset,
			Category:    models.CategoryCurrentAsset,
			Level:       3,
			IsHeader:    false,
			IsActive:    true,
			Description: "Auto-created Bank BNI parent account",
		}
		if cashBankParent, err := s.accountRepo.FindByCode(context.Background(), "1100"); err == nil {
			parentAccount.ParentID = &cashBankParent.ID
		}
		
	case "1110":
		// Other banks
		parentAccount = &models.Account{
			Code:        "1110",
			Name:        "BANK LAIN-LAIN",
			Type:        models.AccountTypeAsset,
			Category:    models.CategoryCurrentAsset,
			Level:       3,
			IsHeader:    false,
			IsActive:    true,
			Description: "Auto-created other banks parent account",
		}
		if cashBankParent, err := s.accountRepo.FindByCode(context.Background(), "1100"); err == nil {
			parentAccount.ParentID = &cashBankParent.ID
		}
		
	default:
		return fmt.Errorf("unsupported parent code: %s", parentCode)
	}
	
	// Create parent account
	if err := s.db.Create(parentAccount).Error; err != nil {
		return fmt.Errorf("failed to create parent account %s: %w", parentCode, err)
	}
	
	fmt.Printf("✅ Created parent account: %s - %s\n", parentAccount.Code, parentAccount.Name)
	return nil
}

// generateNextChildCode generates the next sequential child code for a parent
func (s *PSAKCashBankService) generateNextChildCode(parentCode string) (string, error) {
	// Get all existing child accounts for this parent
	var accounts []models.Account
	err := s.db.Where("code LIKE ? AND type = ?", parentCode+"-%", models.AccountTypeAsset).Find(&accounts).Error
	if err != nil {
		return "", fmt.Errorf("failed to query existing child accounts: %w", err)
	}
	
	// Find the highest existing child number
	maxChildNumber := 0
	childPrefix := parentCode + "-"
	
	for _, account := range accounts {
		if strings.HasPrefix(account.Code, childPrefix) {
			// Extract child number (e.g., "001" from "1101-001")
			childPart := strings.TrimPrefix(account.Code, childPrefix)
			if len(childPart) == 3 {
				if num, err := strconv.Atoi(childPart); err == nil && num > maxChildNumber {
					maxChildNumber = num
				}
			}
		}
	}
	
	// Generate next child number
	nextChildNumber := maxChildNumber + 1
	if nextChildNumber > 999 {
		return "", fmt.Errorf("child account limit reached for parent %s (max: 999)", parentCode)
	}
	
	childCode := fmt.Sprintf("%s-%03d", parentCode, nextChildNumber)
	
	// Mark parent as header if it's not already and this is the first child
	if maxChildNumber == 0 {
		if err := s.markParentAsHeader(parentCode); err != nil {
			fmt.Printf("Warning: failed to mark parent %s as header: %v\n", parentCode, err)
		}
	}
	
	return childCode, nil
}

// markParentAsHeader marks the parent account as a header account
func (s *PSAKCashBankService) markParentAsHeader(parentCode string) error {
	return s.db.Model(&models.Account{}).
		Where("code = ?", parentCode).
		Update("is_header", true).Error
}

// CreatePSAKCompliantGLAccount creates a GL account following PSAK standards
func (s *PSAKCashBankService) CreatePSAKCompliantGLAccount(cashBankType, accountName string) (*models.Account, error) {
	// Generate PSAK-compliant code
	accountCode, err := s.generatePSAKCompliantAccountCode(cashBankType, accountName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate account code: %w", err)
	}
	
	// Find parent account for proper hierarchy
	parentCode := accountCode[:4] // Get parent part (e.g., "1101" from "1101-001")
	parent, err := s.accountRepo.FindByCode(context.Background(), parentCode)
	if err != nil {
		return nil, fmt.Errorf("parent account %s not found: %w", parentCode, err)
	}
	
	// Create GL account
	glAccount := &models.Account{
		Code:        accountCode,
		Name:        accountName,
		Type:        models.AccountTypeAsset,
		Category:    models.CategoryCurrentAsset,
		ParentID:    &parent.ID,
		Level:       parent.Level + 1,
		IsHeader:    false,
		IsActive:    true,
		Description: fmt.Sprintf("Auto-created GL account for %s: %s", cashBankType, accountName),
	}
	
	if err := s.db.Create(glAccount).Error; err != nil {
		return nil, fmt.Errorf("failed to create GL account: %w", err)
	}
	
	fmt.Printf("✅ Created PSAK-compliant GL account: %s - %s\n", glAccount.Code, glAccount.Name)
	return glAccount, nil
}

// MigrateCashBankAccountsToPSAK migrates existing cash/bank accounts to PSAK-compliant structure
func (s *PSAKCashBankService) MigrateCashBankAccountsToPSAK() error {
	fmt.Println("🔄 Starting Cash/Bank accounts migration to PSAK compliance...")
	
	// Get all cash/bank accounts
	var cashBanks []models.CashBank
	if err := s.db.Preload("Account").Find(&cashBanks).Error; err != nil {
		return fmt.Errorf("failed to load cash/bank accounts: %w", err)
	}
	
	fmt.Printf("Found %d cash/bank accounts to migrate\n", len(cashBanks))
	
	// Process each account
	for i, cb := range cashBanks {
		fmt.Printf("Processing %d/%d: %s (%s)\n", i+1, len(cashBanks), cb.Name, cb.Type)
		
		// Check if already has proper PSAK-compliant GL account
		if cb.AccountID > 0 && cb.Account.Code != "" {
			// Check if the linked account follows PSAK format
			if s.isPSAKCompliantCode(cb.Account.Code) {
				fmt.Printf("  ✅ Already PSAK compliant: %s\n", cb.Account.Code)
				continue
			}
		}
		
		// Create new PSAK-compliant GL account
		newGLAccount, err := s.CreatePSAKCompliantGLAccount(cb.Type, cb.Name)
		if err != nil {
			fmt.Printf("  ❌ Failed to create GL account: %v\n", err)
			continue
		}
		
		// Update cash/bank account to link to new GL account
		cb.AccountID = newGLAccount.ID
		if err := s.db.Save(&cb).Error; err != nil {
			fmt.Printf("  ❌ Failed to update cash/bank link: %v\n", err)
			continue
		}
		
		fmt.Printf("  ✅ Migrated to GL account: %s\n", newGLAccount.Code)
	}
	
	fmt.Println("🎉 Cash/Bank accounts migration completed!")
	return nil
}

// isPSAKCompliantCode checks if an account code follows PSAK standards
func (s *PSAKCashBankService) isPSAKCompliantCode(code string) bool {
	// Check main account format (4 digits)
	if len(code) == 4 {
		if _, err := strconv.Atoi(code); err != nil {
			return false // Must be numeric
		}
		return strings.HasPrefix(code, "1") // Asset accounts start with 1
	}
	
	// Check child account format (XXXX-XXX)
	if strings.Contains(code, "-") {
		parts := strings.Split(code, "-")
		if len(parts) != 2 {
			return false // Must have exactly one dash
		}
		
		parentPart := parts[0]
		childPart := parts[1]
		
		// Validate parent part (4 digits starting with 1)
		if len(parentPart) != 4 || !strings.HasPrefix(parentPart, "1") {
			return false
		}
		if _, err := strconv.Atoi(parentPart); err != nil {
			return false
		}
		
		// Validate child part (3 digits)
		if len(childPart) != 3 {
			return false
		}
		if childNum, err := strconv.Atoi(childPart); err != nil || childNum < 1 {
			return false
		}
		
		return true
	}
	
	return false
}

// FixAllCashBankGLLinks fixes all cash/bank accounts to have proper GL links
func (s *PSAKCashBankService) FixAllCashBankGLLinks() error {
	fmt.Println("🔧 Fixing all Cash/Bank GL account links...")
	
	// Get cash/bank accounts that don't have proper GL links
	var cashBanks []models.CashBank
	if err := s.db.Where("account_id = 0 OR account_id IS NULL").Find(&cashBanks).Error; err != nil {
		return fmt.Errorf("failed to find unlinked cash/bank accounts: %w", err)
	}
	
	fmt.Printf("Found %d unlinked cash/bank accounts\n", len(cashBanks))
	
	for _, cb := range cashBanks {
		fmt.Printf("Fixing link for: %s (%s)\n", cb.Name, cb.Type)
		
		// Create GL account
		glAccount, err := s.CreatePSAKCompliantGLAccount(cb.Type, cb.Name)
		if err != nil {
			fmt.Printf("  ❌ Error creating GL account: %v\n", err)
			continue
		}
		
		// Link to cash/bank account
		cb.AccountID = glAccount.ID
		if err := s.db.Save(&cb).Error; err != nil {
			fmt.Printf("  ❌ Error linking accounts: %v\n", err)
			continue
		}
		
		fmt.Printf("  ✅ Created and linked GL account: %s\n", glAccount.Code)
	}
	
	fmt.Println("🎉 All Cash/Bank GL links fixed!")
	return nil
}

// ValidatePSAKCompliance validates all cash/bank accounts for PSAK compliance
func (s *PSAKCashBankService) ValidatePSAKCompliance() ([]string, error) {
	var issues []string
	
	// Get all cash/bank accounts
	var cashBanks []models.CashBank
	if err := s.db.Preload("Account").Find(&cashBanks).Error; err != nil {
		return nil, fmt.Errorf("failed to load cash/bank accounts: %w", err)
	}
	
	for _, cb := range cashBanks {
		// Check GL link
		if cb.AccountID == 0 {
			issues = append(issues, fmt.Sprintf("CashBank %s (%s) not linked to GL account", cb.Code, cb.Name))
			continue
		}
		
		// Check GL account code format
		if !s.isPSAKCompliantCode(cb.Account.Code) {
			issues = append(issues, fmt.Sprintf("CashBank %s linked to non-PSAK compliant GL account: %s", cb.Code, cb.Account.Code))
		}
		
		// Check account type
		if cb.Account.Type != models.AccountTypeAsset {
			issues = append(issues, fmt.Sprintf("CashBank %s linked to non-asset GL account: %s (%s)", cb.Code, cb.Account.Code, cb.Account.Type))
		}
	}
	
	return issues, nil
}