package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔧 Sales Accounting Structure Fix")
	fmt.Println("=================================")

	// Database connection
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = ""
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "sistem_akuntansi"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Step 1: Check current status
	fmt.Println("📊 Step 1: Current Account Status")
	checkCurrentAccounts(db)

	// Step 2: Add missing PPN Keluaran account
	fmt.Println("\n🏗️  Step 2: Adding PPN Keluaran Account")
	addPPNKeluaranAccount(db)

	// Step 3: Analyze and fix accounting inconsistencies
	fmt.Println("\n🔍 Step 3: Analyzing Accounting Logic")
	analyzeAccountingLogic(db)

	// Step 4: Provide recommendations
	fmt.Println("\n💡 Step 4: Recommendations")
	provideRecommendations(db)
}

func checkCurrentAccounts(db *gorm.DB) {
	var accounts []struct {
		Code    string
		Name    string
		Type    string
		Balance float64
	}

	query := `
		SELECT code, name, type, balance 
		FROM accounts 
		WHERE code IN ('1201', '4101', '2102', '2103', '5101', '1301')
		AND deleted_at IS NULL
		ORDER BY code
	`

	err := db.Raw(query).Scan(&accounts).Error
	if err != nil {
		log.Printf("Failed to fetch accounts: %v", err)
		return
	}

	fmt.Println("Key accounts for sales:")
	for _, account := range accounts {
		fmt.Printf("  %s - %s (%s): Rp %.2f\n", 
			account.Code, account.Name, account.Type, account.Balance)
	}
}

func addPPNKeluaranAccount(db *gorm.DB) {
	// Check if PPN Keluaran account exists
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM accounts WHERE code = '2103' AND deleted_at IS NULL").Scan(&count).Error
	if err != nil {
		log.Printf("Error checking PPN Keluaran account: %v", err)
		return
	}

	if count > 0 {
		fmt.Println("✅ PPN Keluaran account (2103) already exists")
		return
	}

	// Get parent account ID for Current Liabilities
	var parentID uint
	err = db.Raw("SELECT id FROM accounts WHERE code = '2100' AND deleted_at IS NULL").Scan(&parentID).Error
	if err != nil {
		log.Printf("Error finding parent account: %v", err)
		return
	}

	// Add PPN Keluaran account
	insertQuery := `
		INSERT INTO accounts (code, name, type, parent_id, balance, status, created_at, updated_at)
		VALUES ('2103', 'PPN Keluaran', 'Liability', ?, 0, 'ACTIVE', NOW(), NOW())
	`

	err = db.Exec(insertQuery, parentID).Error
	if err != nil {
		log.Printf("❌ Failed to add PPN Keluaran account: %v", err)
		return
	}

	fmt.Println("✅ Added PPN Keluaran account (2103)")
}

func analyzeAccountingLogic(db *gorm.DB) {
	// Get current balances
	var arBalance, salesRevenue, taxPayable float64

	db.Raw("SELECT balance FROM accounts WHERE code = '1201'").Scan(&arBalance)
	db.Raw("SELECT balance FROM accounts WHERE code = '4101'").Scan(&salesRevenue)
	db.Raw("SELECT balance FROM accounts WHERE code = '2102'").Scan(&taxPayable)

	fmt.Printf("Current balances:\n")
	fmt.Printf("  Accounts Receivable (1201): Rp %.2f\n", arBalance)
	fmt.Printf("  Sales Revenue (4101):       Rp %.2f\n", salesRevenue)
	fmt.Printf("  Tax Payable (2102):         Rp %.2f\n", taxPayable)

	// Expected logic validation
	expectedTotal := -salesRevenue + (-taxPayable) // Both should be negative (credit balances)
	difference := arBalance - expectedTotal

	fmt.Printf("\n🧮 Accounting Logic Check:\n")
	fmt.Printf("  AR Balance:           Rp %.2f (Debit)\n", arBalance)
	fmt.Printf("  Sales + Tax:          Rp %.2f (Credit total)\n", expectedTotal)
	fmt.Printf("  Difference:           Rp %.2f\n", difference)

	if abs(difference) < 0.01 {
		fmt.Println("✅ Accounting equation balances!")
	} else {
		fmt.Printf("❌ Accounting imbalance of Rp %.2f\n", difference)
		
		if abs(difference - 1100.0) < 0.01 {
			fmt.Println("💡 This looks like 11% PPN on Rp 10,000,000 sales")
			fmt.Println("   Recommend: Move PPN amount from Tax Payable to PPN Keluaran")
		}
	}
}

func provideRecommendations(db *gorm.DB) {
	fmt.Println("📋 Recommended Actions:")

	// Check if journal entries exist for sales
	var journalCount int64
	db.Raw("SELECT COUNT(*) FROM journal_entries WHERE reference_type = 'SALE'").Scan(&journalCount)

	fmt.Printf("1. Journal Entries: Found %d sales journal entries\n", journalCount)
	
	if journalCount == 0 {
		fmt.Println("   ❌ No sales journal entries found!")
		fmt.Println("   → Need to implement proper journal entry creation for sales")
	}

	// Check sales transactions
	var salesCount int64
	var totalSales, totalOutstanding float64
	db.Raw("SELECT COUNT(*), COALESCE(SUM(total_amount), 0), COALESCE(SUM(outstanding_amount), 0) FROM sales WHERE deleted_at IS NULL").
		Scan(&salesCount, &totalSales, &totalOutstanding)

	fmt.Printf("2. Sales Data: %d transactions, Total: Rp %.2f, Outstanding: Rp %.2f\n", 
		salesCount, totalSales, totalOutstanding)

	// Recommendations based on analysis
	fmt.Println("\n🎯 Specific Recommendations:")
	
	fmt.Println("1. **Journal Entry Structure** - Ensure each sale creates:")
	fmt.Println("   ```")
	fmt.Println("   Debit:  1201 Accounts Receivable    [Total Amount]")
	fmt.Println("   Credit: 4101 Sales Revenue          [Amount - PPN]")  
	fmt.Println("   Credit: 2103 PPN Keluaran           [PPN Amount]")
	fmt.Println("   ```")

	fmt.Println("\n2. **Account Mapping** - Verify these accounts exist:")
	checkRequiredAccounts(db)

	fmt.Println("\n3. **PPN Handling** - Consider:")
	fmt.Println("   - Separate PPN Keluaran (2103) from general Tax Payable (2102)")
	fmt.Println("   - Calculate PPN as percentage of sales amount")
	fmt.Println("   - Include PPN in invoice total but track separately")

	fmt.Println("\n4. **Inventory Integration** (if applicable):")
	fmt.Println("   - When sale is made: COGS (5101) Dr, Inventory (1301) Cr")
	fmt.Println("   - Ensure perpetual inventory system if using real-time COGS")
}

func checkRequiredAccounts(db *gorm.DB) {
	requiredAccounts := map[string]string{
		"1201": "Accounts Receivable",
		"4101": "Sales Revenue", 
		"2103": "PPN Keluaran",
		"5101": "Cost of Goods Sold",
		"1301": "Inventory",
		"1101": "Cash",
		"1102": "Bank BCA",
		"1103": "Bank Mandiri",
	}

	var existingCodes []string
	db.Raw("SELECT code FROM accounts WHERE code IN ('1201','4101','2103','5101','1301','1101','1102','1103') AND deleted_at IS NULL").
		Pluck("code", &existingCodes)

	existingMap := make(map[string]bool)
	for _, code := range existingCodes {
		existingMap[code] = true
	}

	for code, name := range requiredAccounts {
		if existingMap[code] {
			fmt.Printf("   ✅ %s - %s\n", code, name)
		} else {
			fmt.Printf("   ❌ %s - %s (MISSING)\n", code, name)
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}