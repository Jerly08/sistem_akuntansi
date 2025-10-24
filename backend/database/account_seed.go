package database

import (
	"fmt"
	"log"
	"strings"
	
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

// SeedAccounts creates initial chart of accounts
func SeedAccounts(db *gorm.DB) error {
	log.Println("🌱 Starting account seeding (idempotent mode)...")
	log.Println("   Note: Accounts from migrations will be preserved, not recreated")
		accounts := []models.Account{
		// ASSETS (1xxx)
		{Code: "1000", Name: "ASSETS", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 1, IsHeader: true, IsActive: true},
		{Code: "1100", Name: "CURRENT ASSETS", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 2, IsHeader: true, IsActive: true},
		{Code: "1101", Name: "KAS", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: true, IsActive: true, Balance: 0},
		{Code: "1102", Name: "BANK", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: true, IsActive: true, Balance: 0},
		{Code: "1200", Name: "ACCOUNTS RECEIVABLE", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 2, IsHeader: true, IsActive: true},
		{Code: "1201", Name: "PIUTANG USAHA", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		
		// Tax Prepaid Accounts (Prepaid taxes/Input VAT)
		{Code: "1114", Name: "PPh 21 DIBAYAR DIMUKA", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1115", Name: "PPh 23 DIBAYAR DIMUKA", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1240", Name: "PPN MASUKAN", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		
		// Inventory
		{Code: "1301", Name: "PERSEDIAAN BARANG DAGANGAN", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},

		{Code: "1500", Name: "FIXED ASSETS", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 2, IsHeader: true, IsActive: true},
		{Code: "1501", Name: "PERALATAN KANTOR", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1502", Name: "KENDARAAN", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1503", Name: "BANGUNAN", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "1509", Name: "TRUK", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3, IsHeader: false, IsActive: true, Balance: 0},

		// LIABILITIES (2xxx)
		{Code: "2000", Name: "LIABILITIES", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 1, IsHeader: true, IsActive: true},
		{Code: "2100", Name: "CURRENT LIABILITIES", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 2, IsHeader: true, IsActive: true},
		{Code: "2101", Name: "UTANG USAHA", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "2103", Name: "PPN KELUARAN", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "2104", Name: "PPh YANG DIPOTONG", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3, IsHeader: false, IsActive: true, Balance: 0},

		// EQUITY (3xxx)
		{Code: "3000", Name: "EQUITY", Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 1, IsHeader: true, IsActive: true},
		{Code: "3101", Name: "MODAL PEMILIK", Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "3201", Name: "LABA DITAHAN", Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 2, IsHeader: false, IsActive: true, Balance: 0},

		// REVENUE (4xxx)
		{Code: "4000", Name: "REVENUE", Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 1, IsHeader: true, IsActive: true},
		{Code: "4101", Name: "PENDAPATAN PENJUALAN", Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "4201", Name: "PENDAPATAN LAIN-LAIN", Type: models.AccountTypeRevenue, Category: models.CategoryOtherIncome, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "4900", Name: "OTHER INCOME", Type: models.AccountTypeRevenue, Category: models.CategoryOtherIncome, Level: 2, IsHeader: false, IsActive: true, Balance: 0},

		// EXPENSES (5xxx)
		{Code: "5000", Name: "EXPENSES", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 1, IsHeader: true, IsActive: true},
		{Code: "5101", Name: "HARGA POKOK PENJUALAN", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5201", Name: "BEBAN GAJI", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5202", Name: "BEBAN LISTRIK", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5203", Name: "BEBAN TELEPON", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5204", Name: "BEBAN TRANSPORTASI", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
		{Code: "5900", Name: "GENERAL EXPENSE", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2, IsHeader: false, IsActive: true, Balance: 0},
	}

	// Set parent relationships based on account hierarchy
	accountMap := make(map[string]uint)

	// First pass: create accounts to get IDs (using atomic FirstOrCreate)
	for _, account := range accounts {
		var existingAccount models.Account
		
		// Use FirstOrCreate for atomic operation (prevents race conditions)
		result := db.Where("code = ? AND deleted_at IS NULL", account.Code).
			Assign(models.Account{
				Name:        account.Name,
				Type:        account.Type,
				Category:    account.Category,
				Level:       account.Level,
				IsHeader:    account.IsHeader,
				IsActive:    account.IsActive,
				Description: account.Description,
			}).
			FirstOrCreate(&existingAccount, models.Account{
				Code:        account.Code,
				Name:        account.Name,
				Type:        account.Type,
				Category:    account.Category,
				Level:       account.Level,
				IsHeader:    account.IsHeader,
				IsActive:    account.IsActive,
				Balance:     account.Balance,
				Description: account.Description,
			})

		if result.Error != nil {
			errMsg := strings.ToLower(result.Error.Error())
			// Check if it's a duplicate/unique constraint error (SQLSTATE 23505)
			if strings.Contains(errMsg, "duplicate") || 
			   strings.Contains(errMsg, "unique") ||
			   strings.Contains(errMsg, "already exists") ||
			   strings.Contains(errMsg, "23505") {
				// Account already exists - this is EXPECTED and NORMAL
				// Migrations often create accounts before seeding runs
				if err := db.Where("code = ? AND deleted_at IS NULL", account.Code).First(&existingAccount).Error; err != nil {
					return fmt.Errorf("failed to fetch existing account %s: %v", account.Code, err)
				}
				accountMap[account.Code] = existingAccount.ID
				// Silent skip - no log needed for normal condition
				continue
			}
			return fmt.Errorf("failed to seed account %s: %v", account.Code, result.Error)
		}

		// Check if account was created or already existed
		if result.RowsAffected > 0 {
			log.Printf("✅ Created account: %s - %s", account.Code, account.Name)
		}
		// Silent if already existed - reduces noise in logs
		
		accountMap[account.Code] = existingAccount.ID
	}

	// Define parent-child relationships
	parentChildMap := map[string]string{
		"1100": "1000", // CURRENT ASSETS -> ASSETS
		"1101": "1100", // Kas -> CURRENT ASSETS
		"1102": "1100", // Bank -> CURRENT ASSETS
		"1200": "1100", // ACCOUNTS RECEIVABLE -> CURRENT ASSETS
		"1201": "1200", // Piutang Usaha -> ACCOUNTS RECEIVABLE
		"1114": "1200", // PPh 21 Dibayar Dimuka -> ACCOUNTS RECEIVABLE
		"1115": "1200", // PPh 23 Dibayar Dimuka -> ACCOUNTS RECEIVABLE
		"1240": "1100", // PPN Masukan -> CURRENT ASSETS
		"1301": "1100", // Persediaan Barang Dagangan -> CURRENT ASSETS
		"1500": "1000", // FIXED ASSETS -> ASSETS
		"1501": "1500", // Peralatan Kantor -> FIXED ASSETS
		"1502": "1500", // Kendaraan -> FIXED ASSETS
		"1503": "1500", // Bangunan -> FIXED ASSETS
		"1509": "1500", // TRUK -> FIXED ASSETS
		"2100": "2000", // CURRENT LIABILITIES -> LIABILITIES
		"2101": "2100", // Utang Usaha -> CURRENT LIABILITIES
		"2103": "2100", // PPN Keluaran -> CURRENT LIABILITIES
		"2104": "2100", // PPh Yang Dipotong -> CURRENT LIABILITIES
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

	// Count how many accounts exist vs created
	var totalAccounts int64
	db.Model(&models.Account{}).Where("deleted_at IS NULL").Count(&totalAccounts)
	
	log.Printf("✅ Account seeding completed: %d accounts ready", totalAccounts)
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
		{
			Code:        "1240",
			ParentCode:  "1100",
			Description: "Fix PPN Masukan (ASSET) to be under CURRENT ASSETS",
		},
		{
			Code:        "1114",
			ParentCode:  "1200",
			Description: "Fix PPh 21 Dibayar Dimuka to be under ACCOUNTS RECEIVABLE",
		},
		{
			Code:        "1115",
			ParentCode:  "1200",
			Description: "Fix PPh 23 Dibayar Dimuka to be under ACCOUNTS RECEIVABLE",
		},
		{
			Code:        "2104",
			ParentCode:  "2100",
			Description: "Fix PPh Yang Dipotong to be under CURRENT LIABILITIES",
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
