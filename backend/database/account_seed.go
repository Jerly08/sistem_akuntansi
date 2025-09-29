package database

import (
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
	"log"
)

// SeedAccounts creates initial chart of accounts
func SeedAccounts(db *gorm.DB) error {
	log.Println("🔒 PRODUCTION MODE: Seeding accounts while preserving existing balances...")
	accounts := []models.Account{
		// ASSETS (1xxx)
		{Code: "1000", Name: "ASSETS", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 1, IsHeader: true, IsActive: true},
		{Code: "1100", Name: "CURRENT ASSETS", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 2, IsHeader: true, IsActive: true},
		{Code: "1101", Name: "Kas", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		// Removed default bank accounts from seed per request (1102-1105)
		{Code: "1200", Name: "ACCOUNTS RECEIVABLE", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 2, IsHeader: true, IsActive: true},
		{Code: "1201", Name: "Piutang Usaha", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1301", Name: "Persediaan Barang Dagangan", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},

		{Code: "1500", Name: "FIXED ASSETS", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 2, IsHeader: true, IsActive: true},
		{Code: "1501", Name: "Peralatan Kantor", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1502", Name: "Kendaraan", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1503", Name: "Bangunan", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1509", Name: "TRUK", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},

		// LIABILITIES (2xxx)
		{Code: "2000", Name: "LIABILITIES", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 1, IsHeader: true, IsActive: true},
		{Code: "2100", Name: "CURRENT LIABILITIES", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 2, IsHeader: true, IsActive: true},
		{Code: "2101", Name: "Utang Usaha", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "2102", Name: "PPN Masukan", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "2103", Name: "PPN Keluaran", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},

		// EQUITY (3xxx)
		{Code: "3000", Name: "EQUITY", Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 1, IsHeader: true, IsActive: true},
		{Code: "3101", Name: "Modal Pemilik", Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "3201", Name: "Laba Ditahan", Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 2, IsHeader: false, IsActive: true, Balance: 0},

		// REVENUE (4xxx)
		{Code: "4000", Name: "REVENUE", Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 1, IsHeader: true, IsActive: true},
		{Code: "4101", Name: "Pendapatan Penjualan", Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "4201", Name: "Pendapatan Lain-lain", Type: models.AccountTypeRevenue, Category: models.CategoryOtherIncome, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "4900", Name: "Other Income", Type: models.AccountTypeRevenue, Category: models.CategoryOtherIncome, Level: 2, IsHeader: false, IsActive: true, Balance: 0},

		// EXPENSES (5xxx)
		{Code: "5000", Name: "EXPENSES", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 1, IsHeader: true, IsActive: true},
		{Code: "5101", Name: "Harga Pokok Penjualan", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5201", Name: "Beban Gaji", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5202", Name: "Beban Listrik", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5203", Name: "Beban Telepon", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5204", Name: "Beban Transportasi", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5900", Name: "General Expense", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
	}

	// Set parent relationships based on account hierarchy
	accountMap := make(map[string]uint)

	// First pass: create accounts to get IDs
	for _, account := range accounts {
		var existingAccount models.Account
		result := db.Where("code = ?", account.Code).First(&existingAccount)

		if result.Error != nil {
			// Account doesn't exist, create it
			if err := db.Create(&account).Error; err != nil {
				return err
			}
			accountMap[account.Code] = account.ID
		} else {
			// Account exists, update it but PRESERVE EXISTING BALANCE
			log.Printf("🔒 Preserving balance for account %s (%s): %.2f", existingAccount.Code, existingAccount.Name, existingAccount.Balance)
			existingAccount.Name = account.Name
			existingAccount.Type = account.Type
			existingAccount.Category = account.Category
			existingAccount.Level = account.Level
			existingAccount.IsHeader = account.IsHeader
			existingAccount.IsActive = account.IsActive
			existingAccount.Description = account.Description
			// REMOVED: existingAccount.Balance = account.Balance (preserving existing balances)

			if err := db.Save(&existingAccount).Error; err != nil {
				return err
			}
			accountMap[account.Code] = existingAccount.ID
		}
	}

	// Define parent-child relationships
	parentChildMap := map[string]string{
		"1100": "1000", // CURRENT ASSETS -> ASSETS
		"1101": "1100", // Kas -> CURRENT ASSETS
		// Removed default bank accounts (1102-1105) from seed mapping per request
		"1200": "1100", // ACCOUNTS RECEIVABLE -> CURRENT ASSETS
		"1201": "1200", // Piutang Usaha -> ACCOUNTS RECEIVABLE
		"1301": "1100", // Persediaan Barang Dagangan -> CURRENT ASSETS
		"1500": "1000", // FIXED ASSETS -> ASSETS
		"1501": "1500", // Peralatan Kantor -> FIXED ASSETS
		"1502": "1500", // Kendaraan -> FIXED ASSETS
		"1503": "1500", // Bangunan -> FIXED ASSETS
		"1509": "1500", // TRUK -> FIXED ASSETS
		"2100": "2000", // CURRENT LIABILITIES -> LIABILITIES
		"2101": "2100", // Utang Usaha -> CURRENT LIABILITIES
		"2102": "1100", // PPN Masukan -> CURRENT ASSETS
		"2103": "2100", // PPN Keluaran -> CURRENT LIABILITIES
		"3101": "3000", // Modal Pemilik -> EQUITY
		"3201": "3000", // Laba Ditahan -> EQUITY
		"4101": "4000", // Pendapatan Penjualan -> REVENUE
		"4201": "4000", // Pendapatan Lain-lain -> REVENUE
		"4900": "4000", // Other Income -> REVENUE
		"5101": "5000", // Harga Pokok Penjualan -> EXPENSES
		"5201": "5000", // Beban Gaji -> EXPENSES
		"5202": "5000", // Beban Listrik -> EXPENSES
		"5203": "5000", // Beban Telepon -> EXPENSES
		"5204": "5000", // Beban Transportasi -> EXPENSES
		"5900": "5000", // General Expense -> EXPENSES
	}

	// Second pass: set parent relationships
	for childCode, parentCode := range parentChildMap {
		if childID, childExists := accountMap[childCode]; childExists {
			if parentID, parentExists := accountMap[parentCode]; parentExists {
				if err := db.Model(&models.Account{}).Where("id = ?", childID).Update("parent_id", parentID).Error; err != nil {
					return err
				}
			}
		}
	}

	log.Println("✅ Account seeding completed - all existing balances preserved")
	return nil
}

// FixAccountHierarchies fixes incorrect account hierarchies in existing databases
func FixAccountHierarchies(db *gorm.DB) error {
	log.Println("🔧 Fixing account hierarchies for existing databases...")
	
	// Define fixes needed for incorrect hierarchies
	hierarchyFixes := []struct {
		Code        string
		ParentCode  string
		Description string
	}{
		{
			Code:        "2103",
			ParentCode:  "2100",
			Description: "Fix PPN Keluaran (LIABILITY) to be under CURRENT LIABILITIES",
		},
	}
	
	for _, fix := range hierarchyFixes {
		log.Printf("🔧 Processing fix: %s", fix.Description)
		
		// Find the account to fix
		var account models.Account
		result := db.Where("code = ?", fix.Code).First(&account)
		if result.Error != nil {
			log.Printf("⚠️  Account %s not found, skipping fix", fix.Code)
			continue
		}
		
		// Find the target parent
		var parent models.Account
		result = db.Where("code = ?", fix.ParentCode).First(&parent)
		if result.Error != nil {
			log.Printf("⚠️  Parent account %s not found, skipping fix", fix.ParentCode)
			continue
		}
		
		// Check if fix is needed
		if account.ParentID != nil && *account.ParentID == parent.ID {
			log.Printf("✅ Account %s (%s) already has correct parent %s", 
				account.Code, account.Name, parent.Code)
			continue
		}
		
		// Apply the fix
		oldParentID := account.ParentID
		newLevel := parent.Level + 1
		
		// Update account with correct parent and level
		result = db.Model(&account).Updates(map[string]interface{}{
			"parent_id": parent.ID,
			"level":     newLevel,
		})
		
		if result.Error != nil {
			log.Printf("❌ Failed to fix account %s: %v", fix.Code, result.Error)
			continue
		}
		
		// Ensure parent is marked as header
		if !parent.IsHeader {
			db.Model(&parent).Update("is_header", true)
		}
		
		log.Printf("✅ Fixed: %s (%s) moved from parent %v to %s (level %d)", 
			account.Code, account.Name, oldParentID, parent.Code, newLevel)
	}
	
	log.Println("✅ Account hierarchy fixes completed")
	return nil
}
