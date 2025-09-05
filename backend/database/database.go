package database

import (
	"fmt"
	"log"
	"strings"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
)

var DB *gorm.DB

// cleanupConstraints removes problematic constraints that may cause migration issues
func cleanupConstraints(db *gorm.DB) {
	log.Println("Cleaning up problematic database constraints...")
	
	// First, check if accounts table exists
	var tableExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'accounts'
	)`).Scan(&tableExists)
	
	if !tableExists {
		log.Println("Accounts table does not exist yet, skipping constraint cleanup")
		return
	}
	
	// Query all existing constraints and indexes on accounts table
	var existingConstraints []string
	db.Raw(`
		SELECT constraint_name 
		FROM information_schema.table_constraints 
		WHERE table_name = 'accounts' 
		AND constraint_type IN ('UNIQUE', 'PRIMARY KEY')
		UNION
		SELECT indexname as constraint_name
		FROM pg_indexes 
		WHERE tablename = 'accounts' 
		AND indexname LIKE '%code%'
	`).Scan(&existingConstraints)
	
	log.Printf("Found %d existing constraints/indexes on accounts table", len(existingConstraints))
	
	// List of potentially problematic constraint/index patterns
	problematicPatterns := []string{
		"uni_accounts_code",
		"accounts_code_key",
		"idx_accounts_code_unique",
		"accounts_code_unique",
		"accounts_code_idx",
		"uq_accounts_code",
	}
	
	// Remove existing problematic constraints/indexes
	for _, existing := range existingConstraints {
		for _, pattern := range problematicPatterns {
			if existing == pattern || strings.Contains(existing, "code") {
				// Try dropping as constraint first
				err := db.Exec(fmt.Sprintf("ALTER TABLE accounts DROP CONSTRAINT IF EXISTS %s", existing)).Error
				if err != nil {
					// If constraint drop fails, try as index
					err = db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", existing)).Error
					if err != nil {
						log.Printf("Note: Failed to drop %s (may not exist): %v", existing, err)
					} else {
						log.Printf("✅ Dropped index %s", existing)
					}
				} else {
					log.Printf("✅ Dropped constraint %s", existing)
				}
				break
			}
		}
	}
	
	// Additional cleanup for known problematic constraint names that might not be detected
	additionalCleanup := []string{
		"uni_accounts_code",
		"accounts_code_key", 
		"idx_accounts_code_unique",
		"accounts_code_unique",
	}
	
	for _, constraint := range additionalCleanup {
		// Try both constraint and index drop silently
		db.Exec(fmt.Sprintf("ALTER TABLE accounts DROP CONSTRAINT IF EXISTS %s", constraint))
		db.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", constraint))
	}
	
	// Drop any remaining unique constraints on code column specifically
	log.Println("Removing any remaining unique constraints on code column...")
	var uniqueConstraints []string
	db.Raw(`
		SELECT tc.constraint_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu 
			ON tc.constraint_name = kcu.constraint_name
		WHERE tc.table_name = 'accounts' 
			AND tc.constraint_type = 'UNIQUE'
			AND kcu.column_name = 'code'
	`).Scan(&uniqueConstraints)
	
	for _, constraint := range uniqueConstraints {
		err := db.Exec(fmt.Sprintf("ALTER TABLE accounts DROP CONSTRAINT IF EXISTS %s", constraint)).Error
		if err != nil {
			log.Printf("Note: Failed to drop unique constraint %s: %v", constraint, err)
		} else {
			log.Printf("✅ Dropped unique constraint %s on code column", constraint)
		}
	}
	
	// Check if our target index already exists
	var targetIndexExists bool
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes 
			WHERE tablename = 'accounts' 
			AND indexname = 'idx_accounts_code_active'
		)
	`).Scan(&targetIndexExists)
	
	if targetIndexExists {
		log.Println("✅ Target partial unique index idx_accounts_code_active already exists")
	} else {
		// Create proper partial unique index for accounts code (only for non-deleted records)
		log.Println("Creating partial unique index for active accounts...")
		err := db.Exec(`
			CREATE UNIQUE INDEX idx_accounts_code_active 
			ON accounts (code) 
			WHERE deleted_at IS NULL
		`).Error
		if err != nil {
			log.Printf("Warning: Failed to create partial unique index on accounts.code: %v", err)
			// Try alternative approach with IF NOT EXISTS
			err2 := db.Exec(`
				CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_code_active 
				ON accounts (code) 
				WHERE deleted_at IS NULL
			`).Error
			if err2 != nil {
				log.Printf("Error: Still failed to create partial unique index: %v", err2)
			} else {
				log.Println("✅ Created proper partial unique index on accounts.code for active records")
			}
		} else {
			log.Println("✅ Created proper partial unique index on accounts.code for active records")
		}
	}
	
	// Verify the final state
	var finalConstraints []string
	db.Raw(`
		SELECT constraint_name 
		FROM information_schema.table_constraints 
		WHERE table_name = 'accounts' 
		AND constraint_type = 'UNIQUE'
		UNION
		SELECT indexname as constraint_name
		FROM pg_indexes 
		WHERE tablename = 'accounts' 
		AND indexname LIKE '%code%'
	`).Scan(&finalConstraints)
	
	log.Printf("Final state: %d constraints/indexes on accounts table: %v", len(finalConstraints), finalConstraints)
	log.Println("Database constraint cleanup completed")
}

// cleanupProductUnitConstraints removes problematic constraints on product_units table
func cleanupProductUnitConstraints(db *gorm.DB) {
	log.Println("Cleaning up ProductUnit constraints...")
	
	// First, check if product_units table exists
	var tableExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'product_units'
	)`).Scan(&tableExists)
	
	if !tableExists {
		log.Println("Product units table does not exist yet, skipping constraint cleanup")
		return
	}
	
	// Check if the problematic constraint exists before trying to drop it
	var constraintExists bool
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.table_constraints 
			WHERE table_name = 'product_units' 
			AND constraint_name = 'uni_product_units_code'
		)
	`).Scan(&constraintExists)
	
	if constraintExists {
		log.Println("Found uni_product_units_code constraint, attempting to drop...")
		err := db.Exec("ALTER TABLE product_units DROP CONSTRAINT IF EXISTS uni_product_units_code").Error
		if err != nil {
			log.Printf("Warning: Failed to drop uni_product_units_code constraint: %v", err)
		} else {
			log.Println("✅ Dropped uni_product_units_code constraint successfully")
		}
	} else {
		log.Println("uni_product_units_code constraint does not exist, nothing to drop")
	}
	
	// Also check for any other code-related constraints on product_units
	var codeConstraints []string
	db.Raw(`
		SELECT constraint_name 
		FROM information_schema.table_constraints 
		WHERE table_name = 'product_units' 
		AND constraint_type = 'UNIQUE'
		AND constraint_name LIKE '%code%'
	`).Scan(&codeConstraints)
	
	if len(codeConstraints) > 0 {
		log.Printf("Found %d code-related constraints on product_units", len(codeConstraints))
		for _, constraint := range codeConstraints {
			log.Printf("Attempting to drop constraint: %s", constraint)
			err := db.Exec(fmt.Sprintf("ALTER TABLE product_units DROP CONSTRAINT IF EXISTS %s", constraint)).Error
			if err != nil {
				log.Printf("Warning: Failed to drop constraint %s: %v", constraint, err)
			} else {
				log.Printf("✅ Dropped constraint %s successfully", constraint)
			}
		}
	}
	
	// Check for any indexes that might be causing issues
	var codeIndexes []string
	db.Raw(`
		SELECT indexname 
		FROM pg_indexes 
		WHERE tablename = 'product_units' 
		AND indexname LIKE '%code%'
	`).Scan(&codeIndexes)
	
	if len(codeIndexes) > 0 {
		log.Printf("Found %d code-related indexes on product_units", len(codeIndexes))
		for _, index := range codeIndexes {
			log.Printf("Code-related index found: %s (will be managed by GORM)", index)
		}
	}
	
	log.Println("ProductUnit constraint cleanup completed")
}

func ConnectDB() *gorm.DB {
	cfg := config.LoadConfig()
	
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = db
	log.Println("Database connected successfully")
	return db
}

func AutoMigrate(db *gorm.DB) {
	log.Println("Starting database migration...")
	
	// First, clean up any problematic constraints
	cleanupConstraints(db)
	
	// Clean up ProductUnit constraints before migration
	cleanupProductUnitConstraints(db)
	
	// Migrate models in order to respect foreign key constraints
	err := db.AutoMigrate(
		// Core models first
		&models.User{},
		&models.AuditLog{},
		
		// Chart of Accounts
		&models.Account{},
		&models.Transaction{},
		
		// Contacts
		&models.Contact{},
		&models.ContactAddress{},
		&models.ContactHistory{},
		&models.CommunicationLog{},
		
		// Products
		&models.ProductCategory{},
		&models.Product{},
		&models.ProductUnit{},
		&models.Inventory{},
		
		// Sales
		&models.Sale{},
		&models.SaleItem{},
		&models.SalePayment{},
		&models.SaleReturn{},
		&models.SaleReturnItem{},
		
		// Purchases
		&models.Purchase{},
		&models.PurchaseItem{},
		&models.PurchaseDocument{},
		&models.PurchaseReceipt{},
		&models.PurchaseReceiptItem{},
		
		// Expenses
		&models.ExpenseCategory{},
		&models.Expense{},
		
		// Assets
		&models.Asset{},
		
		// Cash & Bank
		&models.CashBank{},
		&models.CashBankTransaction{},
		&models.Payment{},
		&models.PaymentAllocation{},
		
		// Journals and reports
		&models.Journal{},
		&models.JournalEntry{},
		&models.Report{},
		&models.ReportTemplate{},
		&models.FinancialRatio{},
		&models.AccountBalance{},
		
		// Budgets
		&models.Budget{},
		&models.BudgetItem{},
		&models.BudgetComparison{},
		
		// Notifications
		&models.Notification{},
		
		// Additional missing models
		&models.CompanyProfile{},
		&models.Permission{},
		&models.RolePermission{},
		&models.UserSession{},
		&models.RefreshToken{},
		&models.BlacklistedToken{},
		&models.RateLimitRecord{},
		&models.AuthAttempt{},
		
		// CashBank Migration Models
		&models.CashBankTransferMigration{},
		&models.BankReconciliationMigration{},
		&models.ReconciliationItemMigration{},
		
		// Migration tracking models
		&models.MigrationRecord{},
	)
	
	if err != nil {
		log.Printf("Failed to migrate core models: %v", err)
		log.Fatal("Stopping migration due to error")
	}
	
	log.Println("Core models migration completed successfully")
	
	// Migrate approval models separately to debug any issues
	log.Println("Starting approval models migration...")
	err = db.AutoMigrate(
		&models.ApprovalWorkflow{},
		&models.ApprovalStep{},
		&models.ApprovalRequest{},
		&models.ApprovalAction{},
		&models.ApprovalHistory{},
	)
	
	if err != nil {
		log.Printf("Failed to migrate approval models: %v", err)
		// Don't fail completely, just log the error
	} else {
		log.Println("Approval models migration completed successfully")
	}
	
	log.Println("Database migration completed successfully")
	
	// Create missing columns that should exist from models but might be missing from database
	CreateMissingColumns(db)
	
	// Run enhanced sales model migration
	EnhanceSalesModel(db)
	
	// Enhanced new sales field migration for new fields
	EnhanceNewSalesFields(db)
	
	// Update tax field sizes to prevent numeric overflow
	UpdateTaxFieldSizes(db)
	
	// Run sales data integrity fix
	FixSalesDataIntegrity(db)
	
	// Run enhanced cashbank model migration
	EnhanceCashBankModel(db)

	// PRODUCTION SAFETY: All balance synchronization logic disabled to prevent account balance resets
	// Balance sync operations have been permanently disabled to protect production data
	log.Println("🛡️  PRODUCTION MODE: All balance synchronization disabled to protect account balances")
	log.Println("✅ Account balances will never be automatically modified during startup")

	// Run cleanup duplicate notifications migration
	CleanupDuplicateNotificationsMigration(db)

	// Create indexes for better performance
	createIndexes(db)
}

func createIndexes(db *gorm.DB) {
	// Performance indexes
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_sales_date ON sales(date)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_purchases_date ON purchases(date)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(date)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(transaction_date)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_products_stock ON products(stock)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_inventory_date ON inventories(transaction_date)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at)`)
	
	// Composite indexes for better query performance
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_sales_customer_date ON sales(customer_id, date)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_purchases_vendor_date ON purchases(vendor_id, date)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_transactions_account_date ON transactions(account_id, transaction_date)`)
	
	// Approval indexes - check if tables exist first
	var count int64
	if db.Raw(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'approval_requests'`).Scan(&count); count > 0 {
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_approval_requests_status ON approval_requests(status)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_approval_requests_entity ON approval_requests(entity_type, entity_id)`)
	}
	
	if db.Raw(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'approval_actions'`).Scan(&count); count > 0 {
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_approval_actions_active ON approval_actions(is_active, status)`)
	}
	
	if db.Raw(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'approval_history'`).Scan(&count); count > 0 {
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_approval_history_request ON approval_history(request_id, created_at)`)
	}
	
	log.Println("Database indexes created successfully")
}

// CreateMissingColumns creates missing columns that should exist from model definitions
func CreateMissingColumns(db *gorm.DB) {
	log.Println("Checking and creating missing columns from model definitions...")

	// Check if sales table exists
	var salesTableExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'sales'
	)`).Scan(&salesTableExists)

	if salesTableExists {
		// Check and add missing pph column to sales table
		var pphColumnExists bool
		db.Raw(`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'sales' AND column_name = 'pph'
		)`).Scan(&pphColumnExists)

		if !pphColumnExists {
			log.Println("Adding missing pph column to sales table...")
			err := db.Exec(`
				ALTER TABLE sales 
				ADD COLUMN pph DECIMAL(15,2) DEFAULT 0;
			`).Error
			if err != nil {
				log.Printf("Warning: Failed to add pph column to sales table: %v", err)
			} else {
				log.Println("Added pph column to sales table successfully")
			}
		}

		// Check and add missing pph_percent column to sales table
		var pphPercentColumnExists bool
		db.Raw(`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'sales' AND column_name = 'pph_percent'
		)`).Scan(&pphPercentColumnExists)

		if !pphPercentColumnExists {
			log.Println("Adding missing pph_percent column to sales table...")
			err := db.Exec(`
				ALTER TABLE sales 
				ADD COLUMN pph_percent DECIMAL(5,2) DEFAULT 0;
			`).Error
			if err != nil {
				log.Printf("Warning: Failed to add pph_percent column to sales table: %v", err)
			} else {
				log.Println("Added pph_percent column to sales table successfully")
			}
		}

		// Check and add missing pph_type column to sales table
		var pphTypeColumnExists bool
		db.Raw(`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'sales' AND column_name = 'pph_type'
		)`).Scan(&pphTypeColumnExists)

		if !pphTypeColumnExists {
			log.Println("Adding missing pph_type column to sales table...")
			err := db.Exec(`
				ALTER TABLE sales 
				ADD COLUMN pph_type VARCHAR(20);
			`).Error
			if err != nil {
				log.Printf("Warning: Failed to add pph_type column to sales table: %v", err)
			} else {
				log.Println("Added pph_type column to sales table successfully")
			}
		}
	}

	// Check if sale_items table exists
	var saleItemsTableExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'sale_items'
	)`).Scan(&saleItemsTableExists)

	if saleItemsTableExists {
		// Check and add missing pph_amount column to sale_items table
		var pphAmountColumnExists bool
		db.Raw(`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'sale_items' AND column_name = 'pph_amount'
		)`).Scan(&pphAmountColumnExists)

		if !pphAmountColumnExists {
			log.Println("Adding missing pph_amount column to sale_items table...")
			err := db.Exec(`
				ALTER TABLE sale_items 
				ADD COLUMN pph_amount DECIMAL(15,2) DEFAULT 0;
			`).Error
			if err != nil {
				log.Printf("Warning: Failed to add pph_amount column to sale_items table: %v", err)
			} else {
				log.Println("Added pph_amount column to sale_items table successfully")
			}
		}

		// Check and add missing revenue_account_id column to sale_items table
		var revenueAccountIdColumnExists bool
		db.Raw(`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'sale_items' AND column_name = 'revenue_account_id'
		)`).Scan(&revenueAccountIdColumnExists)

		if !revenueAccountIdColumnExists {
			log.Println("Adding missing revenue_account_id column to sale_items table...")
			err := db.Exec(`
				ALTER TABLE sale_items 
				ADD COLUMN revenue_account_id INTEGER;
			`).Error
			if err != nil {
				log.Printf("Warning: Failed to add revenue_account_id column to sale_items table: %v", err)
			} else {
				log.Println("Added revenue_account_id column to sale_items table successfully")
			}
		}
	}

	log.Println("Missing columns check completed")
}

// EnhanceSalesModel adds enhanced fields to sales and sale_items tables
func EnhanceSalesModel(db *gorm.DB) {
	log.Println("Starting enhanced sales model migration...")
	
	// Check if migration is needed by checking if new fields exist
	var columnExists bool
	
	// Check if subtotal column exists in sales table
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns 
		WHERE table_name = 'sales' AND column_name = 'subtotal'
	)`).Scan(&columnExists)
	
	if columnExists {
		log.Println("Enhanced sales model fields already exist, skipping migration")
		return
	}
	
	// Add new fields to sales table
	log.Println("Adding enhanced fields to sales table...")
	err := db.Exec(`
		ALTER TABLE sales 
		ADD COLUMN IF NOT EXISTS subtotal DECIMAL(15,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(15,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS taxable_amount DECIMAL(15,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS ppn DECIMAL(8,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS pph DECIMAL(8,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS total_tax DECIMAL(8,2) DEFAULT 0;
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to add enhanced fields to sales table: %v", err)
	} else {
		log.Println("Enhanced fields added to sales table successfully")
	}
	
	// Add new fields to sale_items table
	log.Println("Adding enhanced fields to sale_items table...")
	err = db.Exec(`
		ALTER TABLE sale_items 
		ADD COLUMN IF NOT EXISTS description TEXT,
		ADD COLUMN IF NOT EXISTS discount_percent DECIMAL(5,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(15,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS line_total DECIMAL(15,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS taxable BOOLEAN DEFAULT true,
		ADD COLUMN IF NOT EXISTS ppn_amount DECIMAL(8,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS pph_amount DECIMAL(8,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS total_tax DECIMAL(8,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS final_amount DECIMAL(15,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS tax_account_id INTEGER REFERENCES accounts(id);
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to add enhanced fields to sale_items table: %v", err)
	} else {
		log.Println("Enhanced fields added to sale_items table successfully")
	}
	
	// Update existing records with calculated values
	updateExistingSalesRecords(db)
	
	log.Println("Enhanced sales model migration completed successfully")
}

// updateExistingSalesRecords updates existing sales records with calculated values
func updateExistingSalesRecords(db *gorm.DB) {
	log.Println("Updating existing sales records with calculated values...")
	
	// Update sales records where new fields are null/zero
	err := db.Exec(`
		UPDATE sales 
		SET 
			subtotal = CASE 
				WHEN subtotal = 0 OR subtotal IS NULL THEN COALESCE(total_amount - shipping_cost, total_amount, 0)
				ELSE subtotal 
			END,
			discount_amount = CASE
				WHEN discount_amount = 0 OR discount_amount IS NULL THEN 
					COALESCE((total_amount - shipping_cost) * discount_percent / 100, 0)
				ELSE discount_amount
			END,
			taxable_amount = CASE
				WHEN taxable_amount = 0 OR taxable_amount IS NULL THEN 
					COALESCE(total_amount - shipping_cost - (total_amount - shipping_cost) * discount_percent / 100, total_amount, 0)
				ELSE taxable_amount
			END,
			ppn = CASE
				WHEN ppn = 0 OR ppn IS NULL THEN 
					COALESCE((total_amount - shipping_cost - (total_amount - shipping_cost) * discount_percent / 100) * ppn_percent / 100, 0)
				ELSE ppn
			END,
			pph = CASE
				WHEN pph = 0 OR pph IS NULL THEN 
					COALESCE((total_amount - shipping_cost - (total_amount - shipping_cost) * discount_percent / 100) * pph_percent / 100, 0)
				ELSE pph
			END,
			total_tax = CASE
				WHEN total_tax = 0 OR total_tax IS NULL THEN 
					COALESCE(
						(total_amount - shipping_cost - (total_amount - shipping_cost) * discount_percent / 100) * ppn_percent / 100 - 
						(total_amount - shipping_cost - (total_amount - shipping_cost) * discount_percent / 100) * pph_percent / 100, 
						0
					)
				ELSE total_tax
			END
		WHERE subtotal = 0 OR subtotal IS NULL OR discount_amount = 0 OR discount_amount IS NULL;
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to update existing sales records: %v", err)
	} else {
		log.Println("Updated existing sales records with calculated values")
	}
	
	// Update sale_items records where new fields are null/zero
	log.Println("Updating existing sale_items records...")
	err = db.Exec(`
		UPDATE sale_items si
		SET 
			description = CASE
				WHEN si.description IS NULL OR si.description = '' THEN 
					COALESCE(p.name, 'Product Item')
				ELSE si.description
			END,
			line_total = CASE
				WHEN si.line_total = 0 OR si.line_total IS NULL THEN 
					COALESCE(si.total_price, si.quantity * si.unit_price, 0)
				ELSE si.line_total
			END,
			final_amount = CASE
				WHEN si.final_amount = 0 OR si.final_amount IS NULL THEN 
					COALESCE(si.total_price, si.quantity * si.unit_price, 0)
				ELSE si.final_amount
			END,
			taxable = CASE
				WHEN si.taxable IS NULL THEN true
				ELSE si.taxable
			END
		FROM products p 
		WHERE si.product_id = p.id 
			AND (si.line_total = 0 OR si.line_total IS NULL OR si.final_amount = 0 OR si.final_amount IS NULL 
				 OR si.description IS NULL OR si.description = '' OR si.taxable IS NULL);
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to update existing sale_items records: %v", err)
	} else {
		log.Println("Updated existing sale_items records with calculated values")
	}
	
	// Update sale_items that don't have matching products
	err = db.Exec(`
		UPDATE sale_items 
		SET 
			description = CASE
				WHEN description IS NULL OR description = '' THEN 'Product Item'
				ELSE description
			END,
			line_total = CASE
				WHEN line_total = 0 OR line_total IS NULL THEN 
					COALESCE(total_price, quantity * unit_price, 0)
				ELSE line_total
			END,
			final_amount = CASE
				WHEN final_amount = 0 OR final_amount IS NULL THEN 
					COALESCE(total_price, quantity * unit_price, 0)
				ELSE final_amount
			END,
			taxable = COALESCE(taxable, true)
		WHERE line_total = 0 OR line_total IS NULL OR final_amount = 0 OR final_amount IS NULL 
			 OR description IS NULL OR description = '' OR taxable IS NULL;
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to update sale_items without matching products: %v", err)
	} else {
		log.Println("Updated sale_items records without matching products")
	}
	
	log.Println("Existing records update completed")
}

// EnhanceCashBankModel adds enhanced fields to cash_banks table and related models
func EnhanceCashBankModel(db *gorm.DB) {
	log.Println("Starting enhanced cash bank model migration...")
	
	// Check if migration is needed by checking if new fields exist
	var columnExists bool
	
	// Check if min_balance column exists in cash_banks table
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns 
		WHERE table_name = 'cash_banks' AND column_name = 'min_balance'
	)`).Scan(&columnExists)
	
	if columnExists {
		log.Println("Enhanced cash bank model fields already exist, skipping migration")
		return
	}
	
	// Add new fields to cash_banks table
	log.Println("Adding enhanced fields to cash_banks table...")
	err := db.Exec(`
		ALTER TABLE cash_banks 
		ADD COLUMN IF NOT EXISTS min_balance DECIMAL(15,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS max_balance DECIMAL(15,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS daily_limit DECIMAL(15,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS monthly_limit DECIMAL(15,2) DEFAULT 0,
		ADD COLUMN IF NOT EXISTS is_restricted BOOLEAN DEFAULT false,
		ADD COLUMN IF NOT EXISTS user_id INTEGER;
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to add enhanced fields to cash_banks table: %v", err)
	} else {
		log.Println("Enhanced fields added to cash_banks table successfully")
	}
	
	// Update existing NOT NULL constraints and defaults
	log.Println("Updating constraints and defaults for cash_banks table...")
	err = db.Exec(`
		ALTER TABLE cash_banks 
		ALTER COLUMN currency SET DEFAULT 'IDR',
		ALTER COLUMN currency SET NOT NULL,
		ALTER COLUMN balance SET DEFAULT 0,
		ALTER COLUMN balance SET NOT NULL,
		ALTER COLUMN is_active SET DEFAULT true,
		ALTER COLUMN is_active SET NOT NULL;
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to update constraints for cash_banks table: %v", err)
	} else {
		log.Println("Updated constraints for cash_banks table successfully")
	}
	
	// Add check constraint for account type
	log.Println("Adding check constraint for cash_banks account type...")
	err = db.Exec(`
		ALTER TABLE cash_banks 
		DROP CONSTRAINT IF EXISTS check_cash_banks_type,
		ADD CONSTRAINT check_cash_banks_type CHECK (type IN ('CASH', 'BANK'));
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to add check constraint for cash_banks type: %v", err)
	} else {
		log.Println("Added check constraint for cash_banks type successfully")
	}
	
	// Create cash bank transfer table if not exists
	log.Println("Creating cash_bank_transfers table if not exists...")
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cash_bank_transfers (
			id SERIAL PRIMARY KEY,
			transfer_number VARCHAR(50) UNIQUE NOT NULL,
			from_account_id INTEGER NOT NULL REFERENCES cash_banks(id),
			to_account_id INTEGER NOT NULL REFERENCES cash_banks(id),
			date TIMESTAMP NOT NULL,
			amount DECIMAL(15,2) NOT NULL,
			exchange_rate DECIMAL(12,6) DEFAULT 1,
			converted_amount DECIMAL(15,2) NOT NULL,
			reference VARCHAR(100),
			notes TEXT,
			status VARCHAR(20) DEFAULT 'PENDING',
			user_id INTEGER NOT NULL REFERENCES users(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL
		);
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to create cash_bank_transfers table: %v", err)
	} else {
		log.Println("Created cash_bank_transfers table successfully")
	}
	
	// Create bank reconciliation table if not exists
	log.Println("Creating bank_reconciliations table if not exists...")
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bank_reconciliations (
			id SERIAL PRIMARY KEY,
			cash_bank_id INTEGER NOT NULL REFERENCES cash_banks(id),
			reconcile_date TIMESTAMP NOT NULL,
			statement_balance DECIMAL(15,2) NOT NULL,
			system_balance DECIMAL(15,2) NOT NULL,
			difference DECIMAL(15,2) NOT NULL,
			status VARCHAR(20) DEFAULT 'PENDING',
			user_id INTEGER NOT NULL REFERENCES users(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL
		);
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to create bank_reconciliations table: %v", err)
	} else {
		log.Println("Created bank_reconciliations table successfully")
	}
	
	// Create reconciliation items table if not exists
	log.Println("Creating reconciliation_items table if not exists...")
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS reconciliation_items (
			id SERIAL PRIMARY KEY,
			reconciliation_id INTEGER NOT NULL REFERENCES bank_reconciliations(id),
			transaction_id INTEGER NOT NULL REFERENCES cash_bank_transactions(id),
			is_cleared BOOLEAN DEFAULT false,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL
		);
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to create reconciliation_items table: %v", err)
	} else {
		log.Println("Created reconciliation_items table successfully")
	}
	
	// Update existing cash bank records with default values
	updateExistingCashBankRecords(db)
	
	// Create indexes for cash bank tables
	createCashBankIndexes(db)
	
	log.Println("Enhanced cash bank model migration completed successfully")
}

// updateExistingCashBankRecords updates existing cash bank records with default values
func updateExistingCashBankRecords(db *gorm.DB) {
	log.Println("Updating existing cash bank records with default values...")
	
	// Update existing records that have NULL values for new fields
	err := db.Exec(`
		UPDATE cash_banks 
		SET 
			currency = CASE
				WHEN currency IS NULL OR currency = '' THEN 'IDR'
				ELSE currency
			END,
			balance = CASE
				WHEN balance IS NULL THEN 0
				ELSE balance
			END,
			is_active = CASE
				WHEN is_active IS NULL THEN true
				ELSE is_active
			END,
			min_balance = COALESCE(min_balance, 0),
			max_balance = COALESCE(max_balance, 0),
			daily_limit = COALESCE(daily_limit, 0),
			monthly_limit = COALESCE(monthly_limit, 0),
			is_restricted = COALESCE(is_restricted, false),
			user_id = CASE
				WHEN user_id IS NULL OR user_id = 0 THEN (
					SELECT id FROM users WHERE role = 'admin' ORDER BY id LIMIT 1
				)
				ELSE user_id
			END
		WHERE currency IS NULL OR currency = '' OR balance IS NULL 
			 OR is_active IS NULL OR min_balance IS NULL OR max_balance IS NULL 
			 OR daily_limit IS NULL OR monthly_limit IS NULL 
			 OR is_restricted IS NULL OR user_id IS NULL OR user_id = 0;
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to update existing cash bank records: %v", err)
	} else {
		log.Println("Updated existing cash bank records with default values")
	}
	
	// Set default user_id to first admin user if still NULL
	err = db.Exec(`
		UPDATE cash_banks 
		SET user_id = 1 
		WHERE user_id IS NULL OR user_id = 0;
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to set default user_id for cash bank records: %v", err)
	} else {
		log.Println("Set default user_id for cash bank records")
	}
	
	// Now make user_id NOT NULL after all records have been updated
	log.Println("Setting user_id column as NOT NULL...")
	err = db.Exec(`
		ALTER TABLE cash_banks 
		ALTER COLUMN user_id SET NOT NULL;
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to set user_id as NOT NULL: %v", err)
	} else {
		log.Println("Set user_id column as NOT NULL successfully")
	}
	
	log.Println("Cash bank records update completed")
}

// createCashBankIndexes creates indexes for cash bank related tables
func createCashBankIndexes(db *gorm.DB) {
	log.Println("Creating cash bank indexes...")
	
	// Cash Banks indexes
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_cash_banks_type ON cash_banks(type)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_cash_banks_currency ON cash_banks(currency)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_cash_banks_active ON cash_banks(is_active, is_restricted)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_cash_banks_user ON cash_banks(user_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_cash_banks_balance ON cash_banks(balance, currency)`)
	
	// Cash Bank Transactions indexes
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_cash_bank_transactions_account_date ON cash_bank_transactions(cash_bank_id, transaction_date)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_cash_bank_transactions_reference ON cash_bank_transactions(reference_type, reference_id)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_cash_bank_transactions_date ON cash_bank_transactions(transaction_date DESC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_cash_bank_transactions_amount ON cash_bank_transactions(amount, balance_after)`)
	
	// Cash Bank Transfers indexes (if table exists)
	var tableExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'cash_bank_transfers'
	)`).Scan(&tableExists)
	
	if tableExists {
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_cash_bank_transfers_from_account ON cash_bank_transfers(from_account_id, date)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_cash_bank_transfers_to_account ON cash_bank_transfers(to_account_id, date)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_cash_bank_transfers_status ON cash_bank_transfers(status, date)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_cash_bank_transfers_user ON cash_bank_transfers(user_id, date)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_cash_bank_transfers_amount ON cash_bank_transfers(amount, converted_amount)`)
	}
	
	// Bank Reconciliations indexes (if table exists)
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'bank_reconciliations'
	)`).Scan(&tableExists)
	
	if tableExists {
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_bank_reconciliations_account_date ON bank_reconciliations(cash_bank_id, reconcile_date)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_bank_reconciliations_status ON bank_reconciliations(status, reconcile_date)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_bank_reconciliations_user ON bank_reconciliations(user_id, reconcile_date)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_bank_reconciliations_difference ON bank_reconciliations(difference, status)`)
	}
	
	// Reconciliation Items indexes (if table exists)
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'reconciliation_items'
	)`).Scan(&tableExists)
	
	if tableExists {
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_reconciliation_items_reconciliation ON reconciliation_items(reconciliation_id, is_cleared)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_reconciliation_items_transaction ON reconciliation_items(transaction_id, reconciliation_id)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_reconciliation_items_cleared ON reconciliation_items(is_cleared, reconciliation_id)`)
	}
	
	log.Println("Cash bank indexes created successfully")
}

// EnhanceNewSalesFields ensures all new fields from recent model changes are properly migrated
func EnhanceNewSalesFields(db *gorm.DB) {
	log.Println("Starting enhanced new sales fields migration...")
	
	// Check if description column exists in sale_items table (indicates if migration is needed)
	var descColumnExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns 
		WHERE table_name = 'sale_items' AND column_name = 'description'
	)`).Scan(&descColumnExists)
	
	if !descColumnExists {
		log.Println("Adding missing new fields to sale_items table...")
		err := db.Exec(`
			ALTER TABLE sale_items 
			ADD COLUMN IF NOT EXISTS description TEXT;
		`).Error
		
		if err != nil {
			log.Printf("Warning: Failed to add description field to sale_items table: %v", err)
		} else {
			log.Println("Added description field to sale_items table successfully")
		}
	}
	
	// Check if taxable column exists in sale_items table
	var taxableColumnExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns 
		WHERE table_name = 'sale_items' AND column_name = 'taxable'
	)`).Scan(&taxableColumnExists)
	
	if !taxableColumnExists {
		log.Println("Adding taxable field to sale_items table...")
		err := db.Exec(`
			ALTER TABLE sale_items 
			ADD COLUMN IF NOT EXISTS taxable BOOLEAN DEFAULT true;
		`).Error
		
		if err != nil {
			log.Printf("Warning: Failed to add taxable field to sale_items table: %v", err)
		} else {
			log.Println("Added taxable field to sale_items table successfully")
		}
	}
	
	// Check if discount_percent column exists in sale_items table
	var discountPercentColumnExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns 
		WHERE table_name = 'sale_items' AND column_name = 'discount_percent'
	)`).Scan(&discountPercentColumnExists)
	
	if !discountPercentColumnExists {
		log.Println("Adding discount_percent field to sale_items table...")
		err := db.Exec(`
			ALTER TABLE sale_items 
			ADD COLUMN IF NOT EXISTS discount_percent DECIMAL(5,2) DEFAULT 0;
		`).Error
		
		if err != nil {
			log.Printf("Warning: Failed to add discount_percent field to sale_items table: %v", err)
		} else {
			log.Println("Added discount_percent field to sale_items table successfully")
		}
	}

	// Check if pph_percent column exists in sales table
	var pphPercentColumnExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns 
		WHERE table_name = 'sales' AND column_name = 'pph_percent'
	)`).Scan(&pphPercentColumnExists)

	if !pphPercentColumnExists {
		log.Println("Adding pph_percent field to sales table...")
		err := db.Exec(`
			ALTER TABLE sales 
			ADD COLUMN IF NOT EXISTS pph_percent DECIMAL(5,2) DEFAULT 0;
		`).Error

		if err != nil {
			log.Printf("Warning: Failed to add pph_percent field to sales table: %v", err)
		} else {
			log.Println("Added pph_percent field to sales table successfully")
		}
	}

	// Update existing records that have null values for new fields
	log.Println("Updating existing sale_items records with default values for new fields...")
	err := db.Exec(`
		UPDATE sale_items si
		SET 
			description = CASE
				WHEN si.description IS NULL OR si.description = '' THEN 
					COALESCE(p.name, 'Product Item')
				ELSE si.description
			END,
			taxable = CASE
				WHEN si.taxable IS NULL THEN true
				ELSE si.taxable
			END,
			discount_percent = CASE
				WHEN si.discount_percent IS NULL THEN 0
				ELSE si.discount_percent
			END
		FROM products p 
		WHERE si.product_id = p.id 
			AND (si.description IS NULL OR si.description = '' OR si.taxable IS NULL OR si.discount_percent IS NULL);
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to update existing sale_items with new field defaults: %v", err)
	} else {
		log.Println("Updated existing sale_items records with new field defaults")
	}
	
	// Update records without matching products
	err = db.Exec(`
		UPDATE sale_items
		SET 
			description = CASE
				WHEN description IS NULL OR description = '' THEN 'Product Item'
				ELSE description
			END,
			taxable = COALESCE(taxable, true),
			discount_percent = COALESCE(discount_percent, 0)
		WHERE description IS NULL OR description = '' OR taxable IS NULL OR discount_percent IS NULL;
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to update sale_items without matching products: %v", err)
	} else {
		log.Println("Updated sale_items records without matching products")
	}
	
	// Ensure tax_account_id foreign key exists if column exists
	var taxAccountColumnExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns 
		WHERE table_name = 'sale_items' AND column_name = 'tax_account_id'
	)`).Scan(&taxAccountColumnExists)
	
	if taxAccountColumnExists {
		// Check if specific foreign key constraint exists
		var constraintExists bool
		db.Raw(`SELECT EXISTS (
			SELECT 1 FROM information_schema.table_constraints 
			WHERE table_name = 'sale_items' 
			AND constraint_type = 'FOREIGN KEY' 
			AND constraint_name = 'fk_sale_items_tax_account'
		)`).Scan(&constraintExists)
		
		if !constraintExists {
			log.Println("Adding foreign key constraint for tax_account_id...")
			err := db.Exec(`
				ALTER TABLE sale_items 
				ADD CONSTRAINT fk_sale_items_tax_account 
				FOREIGN KEY (tax_account_id) REFERENCES accounts(id);
			`).Error
			
			if err != nil {
				log.Printf("Warning: Failed to add foreign key constraint for tax_account_id: %v", err)
			} else {
				log.Println("Added foreign key constraint for tax_account_id successfully")
			}
		} else {
			log.Println("Foreign key constraint for tax_account_id already exists, skipping")
		}
	}
	
	// Add indexes for new fields
	log.Println("Adding indexes for new sales fields...")
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_sale_items_description ON sale_items(description)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_sale_items_taxable ON sale_items(taxable)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_sale_items_discount_percent ON sale_items(discount_percent)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_sale_items_tax_account ON sale_items(tax_account_id)`)
	
	log.Println("Enhanced new sales fields migration completed successfully")
}

// FixSalesDataIntegrity performs comprehensive sales data integrity fixes and validation
func FixSalesDataIntegrity(db *gorm.DB) {
	log.Println("=== Starting Sales Data Integrity Fix ===")
	
	// Check if we have sales records to fix
	var salesCount int64
	db.Model(&models.Sale{}).Count(&salesCount)
	
	if salesCount == 0 {
		log.Println("No sales records found, skipping sales data integrity fix")
		return
	}
	
	log.Printf("Found %d sales records, starting integrity checks and fixes...", salesCount)
	
	// 1. Fix missing sale codes
	fixMissingSaleCodes(db)
	
	// 2. Fix sale item calculations
	fixSaleItemCalculations(db)
	
	// 3. Recalculate sale totals
	recalculateSaleTotals(db)
	
	// 4. Check and report orphaned records
	checkOrphanedRecords(db)
	
	// 5. Validate status consistency
	validateStatusConsistency(db)
	
	// 6. Update legacy computed fields
	updateLegacyComputedFields(db)
	
	log.Println("✅ Sales Data Integrity Fix completed successfully")
}

// fixMissingSaleCodes generates codes for sales that don't have them
func fixMissingSaleCodes(db *gorm.DB) {
	log.Println("Fixing missing sale codes...")
	
	var salesWithoutCodes []models.Sale
	db.Where("code = '' OR code IS NULL").Find(&salesWithoutCodes)
	
	if len(salesWithoutCodes) == 0 {
		log.Println("No sales found without codes")
		return
	}
	
	fixedCodes := 0
	for i := range salesWithoutCodes {
		sale := &salesWithoutCodes[i]
		
		// Generate new code based on type
		prefix := "SAL"
		switch sale.Type {
		case models.SaleTypeQuotation:
			prefix = "QUO"
		case models.SaleTypeOrder:
			prefix = "ORD"
		case models.SaleTypeInvoice:
			prefix = "INV"
		}
		
		year := sale.Date.Year()
		newCode := fmt.Sprintf("%s-%d-%04d", prefix, year, sale.ID)
		
		// Check if code already exists
		var existing models.Sale
		if db.Where("code = ?", newCode).First(&existing).Error == nil {
			// Code exists, add suffix
			newCode = fmt.Sprintf("%s-FIX-%d", newCode, sale.ID)
		}
		
		sale.Code = newCode
		if err := db.Save(sale).Error; err != nil {
			log.Printf("Warning: Failed to update sale %d code: %v", sale.ID, err)
		} else {
			fixedCodes++
		}
	}
	
	log.Printf("Fixed %d missing sale codes", fixedCodes)
}

// fixSaleItemCalculations fixes missing calculations in sale items
func fixSaleItemCalculations(db *gorm.DB) {
	log.Println("Fixing sale item calculations...")
	
	// Fix missing LineTotal, FinalAmount, and other computed fields
	err := db.Exec(`
		UPDATE sale_items si
		SET 
			line_total = CASE
				WHEN line_total = 0 OR line_total IS NULL THEN 
					(quantity * unit_price) - COALESCE(discount_amount, discount, 0)
				ELSE line_total
			END,
			final_amount = CASE
				WHEN final_amount = 0 OR final_amount IS NULL THEN 
					(quantity * unit_price) - COALESCE(discount_amount, discount, 0) + COALESCE(total_tax, tax, 0)
				ELSE final_amount
			END,
			discount_amount = CASE
				WHEN discount_amount = 0 OR discount_amount IS NULL AND discount_percent > 0 THEN 
					(quantity * unit_price) * discount_percent / 100
				WHEN discount_amount = 0 OR discount_amount IS NULL THEN 
					COALESCE(discount, 0)
				ELSE discount_amount
			END,
			-- Update legacy fields for backward compatibility
			total_price = CASE
				WHEN total_price = 0 OR total_price IS NULL THEN 
					(quantity * unit_price) - COALESCE(discount_amount, discount, 0)
				ELSE total_price
			END
		WHERE line_total = 0 OR line_total IS NULL 
			 OR final_amount = 0 OR final_amount IS NULL 
			 OR (discount_amount = 0 OR discount_amount IS NULL) AND discount_percent > 0
			 OR total_price = 0 OR total_price IS NULL;
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to fix sale item calculations: %v", err)
	} else {
		log.Println("Fixed sale item calculations successfully")
	}
}

// recalculateSaleTotals recalculates totals for all sales
func recalculateSaleTotals(db *gorm.DB) {
	log.Println("Recalculating sale totals...")
	
	// Recalculate sales totals based on their items
	err := db.Exec(`
		UPDATE sales s
		SET 
			subtotal = COALESCE((
				SELECT SUM(si.line_total) 
				FROM sale_items si 
				WHERE si.sale_id = s.id AND si.deleted_at IS NULL
			), 0),
			discount_amount = CASE
				WHEN discount_percent > 0 THEN 
					COALESCE((
						SELECT SUM(si.line_total) 
						FROM sale_items si 
						WHERE si.sale_id = s.id AND si.deleted_at IS NULL
					), 0) * discount_percent / 100
				ELSE discount_amount
			END,
			taxable_amount = COALESCE((
				SELECT SUM(si.line_total) 
				FROM sale_items si 
				WHERE si.sale_id = s.id AND si.deleted_at IS NULL
			), 0) - CASE
				WHEN discount_percent > 0 THEN 
					COALESCE((
						SELECT SUM(si.line_total) 
						FROM sale_items si 
						WHERE si.sale_id = s.id AND si.deleted_at IS NULL
					), 0) * discount_percent / 100
				ELSE COALESCE(discount_amount, 0)
			END,
		ppn = CASE
				WHEN ppn_percent > 0 THEN 
					(
						COALESCE((
							SELECT SUM(si.line_total) 
							FROM sale_items si 
							WHERE si.sale_id = s.id AND si.deleted_at IS NULL
						), 0) - CASE
							WHEN discount_percent > 0 THEN 
								COALESCE((
									SELECT SUM(si.line_total) 
									FROM sale_items si 
									WHERE si.sale_id = s.id AND si.deleted_at IS NULL
								), 0) * discount_percent / 100
							ELSE COALESCE(discount_amount, 0)
						END
					) * ppn_percent / 100
				ELSE ppn
			END,
			pph = CASE
				WHEN pph_percent > 0 THEN 
					(
						COALESCE((
							SELECT SUM(si.line_total) 
							FROM sale_items si 
							WHERE si.sale_id = s.id AND si.deleted_at IS NULL
						), 0) - CASE
							WHEN discount_percent > 0 THEN 
								COALESCE((
									SELECT SUM(si.line_total) 
									FROM sale_items si 
									WHERE si.sale_id = s.id AND si.deleted_at IS NULL
								), 0) * discount_percent / 100
							ELSE COALESCE(discount_amount, 0)
						END
					) * pph_percent / 100
				ELSE pph
			END
		WHERE EXISTS (
			SELECT 1 FROM sale_items si 
			WHERE si.sale_id = s.id AND si.deleted_at IS NULL
		);
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to recalculate sale totals: %v", err)
	} else {
		log.Println("Recalculated sale totals successfully")
	}
	
	// Update total_tax and total_amount
	err = db.Exec(`
		UPDATE sales 
		SET 
			total_tax = COALESCE(ppn, 0) - COALESCE(pph, 0),
			total_amount = COALESCE(taxable_amount, 0) + COALESCE(ppn, 0) - COALESCE(pph, 0) + COALESCE(shipping_cost, 0),
			outstanding_amount = COALESCE(taxable_amount, 0) + COALESCE(ppn, 0) - COALESCE(pph, 0) + COALESCE(shipping_cost, 0) - COALESCE(paid_amount, 0),
			-- Update legacy tax field
			tax = COALESCE(ppn, 0) - COALESCE(pph, 0)
		WHERE taxable_amount IS NOT NULL;
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to update final totals: %v", err)
	} else {
		log.Println("Updated final totals successfully")
	}
}

// checkOrphanedRecords checks for orphaned sale items and other inconsistencies
func checkOrphanedRecords(db *gorm.DB) {
	log.Println("Checking for orphaned records...")
	
	// Check for orphaned sale items
	var orphanedItemsCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM sale_items si 
		LEFT JOIN sales s ON si.sale_id = s.id 
		WHERE s.id IS NULL
	`).Scan(&orphanedItemsCount)
	
	if orphanedItemsCount > 0 {
		log.Printf("Warning: Found %d orphaned sale items", orphanedItemsCount)
		// Optionally delete orphaned items or flag them for manual review
		// db.Exec("DELETE FROM sale_items WHERE id IN (SELECT si.id FROM sale_items si LEFT JOIN sales s ON si.sale_id = s.id WHERE s.id IS NULL)")
	} else {
		log.Println("No orphaned sale items found")
	}
	
	// Check for orphaned sale payments
	var orphanedPaymentsCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM sale_payments sp 
		LEFT JOIN sales s ON sp.sale_id = s.id 
		WHERE s.id IS NULL
	`).Scan(&orphanedPaymentsCount)
	
	if orphanedPaymentsCount > 0 {
		log.Printf("Warning: Found %d orphaned sale payments", orphanedPaymentsCount)
	} else {
		log.Println("No orphaned sale payments found")
	}
}

// validateStatusConsistency checks for invalid status transitions and inconsistencies
func validateStatusConsistency(db *gorm.DB) {
	log.Println("Validating status consistency...")
	
	// Check for INVOICED sales without invoice numbers
	var invalidInvoicedCount int64
	db.Model(&models.Sale{}).
		Where("status = ? AND (invoice_number = '' OR invoice_number IS NULL)", models.SaleStatusInvoiced).
		Count(&invalidInvoicedCount)
	
	if invalidInvoicedCount > 0 {
		log.Printf("Warning: Found %d INVOICED sales without invoice numbers", invalidInvoicedCount)
	}
	
	// Check for PAID sales with outstanding amounts > 0
	var invalidPaidCount int64
	db.Model(&models.Sale{}).
		Where("status = ? AND outstanding_amount > 0", models.SaleStatusPaid).
		Count(&invalidPaidCount)
	
	if invalidPaidCount > 0 {
		log.Printf("Warning: Found %d PAID sales with outstanding amounts > 0", invalidPaidCount)
		
		// Auto-fix: Update status to INVOICED if there's still outstanding amount
		err := db.Model(&models.Sale{}).
			Where("status = ? AND outstanding_amount > 0", models.SaleStatusPaid).
			Update("status", models.SaleStatusInvoiced).Error
		
		if err != nil {
			log.Printf("Warning: Failed to fix PAID status inconsistency: %v", err)
		} else {
			log.Printf("Fixed %d PAID sales with outstanding amounts", invalidPaidCount)
		}
	}
}

// updateLegacyComputedFields updates legacy computed fields for backward compatibility
// UpdateTaxFieldSizes updates tax field sizes from decimal(8,2) to decimal(15,2) to prevent numeric overflow
func UpdateTaxFieldSizes(db *gorm.DB) {
	log.Println("Starting tax field size update to prevent numeric overflow...")
	
	// Check if migration has already been applied by checking field size
	var columnInfo struct {
		NumericPrecision int `json:"numeric_precision"`
	}
	
	db.Raw(`SELECT numeric_precision 
			 FROM information_schema.columns 
			 WHERE table_name = 'sales' 
			 AND column_name = 'tax' 
			 LIMIT 1`).Scan(&columnInfo)
	
	if columnInfo.NumericPrecision >= 15 {
		log.Println("Tax field sizes already updated, skipping migration")
		return
	}
	
	// Update sales table tax fields from decimal(8,2) to decimal(15,2)
	log.Println("Updating sales table tax field sizes...")
	err := db.Exec(`
		ALTER TABLE sales 
			ALTER COLUMN tax TYPE DECIMAL(15,2),
			ALTER COLUMN ppn TYPE DECIMAL(15,2),
			ALTER COLUMN pph TYPE DECIMAL(15,2),
			ALTER COLUMN total_tax TYPE DECIMAL(15,2);
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to update sales table tax field sizes: %v", err)
	} else {
		log.Println("Updated sales table tax field sizes successfully")
	}
	
	// Update sale_items table tax fields from decimal(8,2) to decimal(15,2)
	log.Println("Updating sale_items table tax field sizes...")
	err = db.Exec(`
		ALTER TABLE sale_items 
			ALTER COLUMN ppn_amount TYPE DECIMAL(15,2),
			ALTER COLUMN pph_amount TYPE DECIMAL(15,2),
			ALTER COLUMN total_tax TYPE DECIMAL(15,2),
			ALTER COLUMN discount TYPE DECIMAL(15,2),
			ALTER COLUMN tax TYPE DECIMAL(15,2);
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to update sale_items table tax field sizes: %v", err)
	} else {
		log.Println("Updated sale_items table tax field sizes successfully")
	}
	
	log.Println("Tax field size update completed successfully")
}

func updateLegacyComputedFields(db *gorm.DB) {
	log.Println("Updating legacy computed fields...")
	
	// Update sale_items legacy fields
	err := db.Exec(`
		UPDATE sale_items 
		SET 
			total_price = line_total,
			tax = total_tax,
			discount = CASE 
				WHEN discount_amount > 0 THEN discount_amount
				WHEN discount_percent > 0 AND quantity > 0 AND unit_price > 0 THEN 
					(quantity * unit_price) * discount_percent / 100
				ELSE discount
			END
		WHERE total_price != line_total OR tax != total_tax OR (
			discount_amount > 0 AND discount != discount_amount
		);
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to update legacy sale item fields: %v", err)
	} else {
		log.Println("Updated legacy sale item fields successfully")
	}
	
	// Update sales legacy fields
	err = db.Exec(`
		UPDATE sales 
		SET 
			tax = total_tax
		WHERE tax != total_tax;
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to update legacy sales fields: %v", err)
	} else {
	log.Println("Updated legacy sales fields successfully")
	}
}

// SyncCashBankGLBalances - DISABLED FOR PRODUCTION SAFETY
// This function has been permanently disabled to prevent account balance resets in production
func SyncCashBankGLBalances(db *gorm.DB) {
	log.Println("🛡️  PRODUCTION SAFETY: SyncCashBankGLBalances DISABLED")
	log.Println("⚠️  Balance synchronization skipped to protect account data")
	log.Println("✅ All account balances remain unchanged")
	return // Exit immediately - no balance operations will be performed
	
	// Check if both tables exist first
	var cashBankTableExists, accountsTableExists bool
	
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'cash_banks'
	)`).Scan(&cashBankTableExists)
	
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'accounts'
	)`).Scan(&accountsTableExists)
	
	if !cashBankTableExists || !accountsTableExists {
		log.Println("Required tables not found, skipping CashBank-GL balance sync")
		return
	}
	
	// Check if we have any cash bank accounts to sync
	var totalCashBankCount int64
	db.Raw(`SELECT COUNT(*) FROM cash_banks WHERE deleted_at IS NULL`).Scan(&totalCashBankCount)
	
	if totalCashBankCount == 0 {
		log.Println("No cash/bank accounts found, skipping balance sync")
		return
	}
	
	// Check for unsynchronized accounts
	var unsyncCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM cash_banks cb 
		INNER JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL 
		  AND cb.balance != acc.balance
	`).Scan(&unsyncCount)
	
	if unsyncCount == 0 {
		log.Println("✅ All CashBank accounts are already synchronized with GL accounts")
		return
	}
	
	log.Printf("Found %d unsynchronized CashBank-GL account pairs", unsyncCount)
	
	// Get details of unsynchronized accounts for logging
	type UnsyncAccount struct {
		CashBankCode    string  `json:"cash_bank_code"`
		CashBankName    string  `json:"cash_bank_name"`
		CashBankBalance float64 `json:"cash_bank_balance"`
		GLCode          string  `json:"gl_code"`
		GLBalance       float64 `json:"gl_balance"`
		Difference      float64 `json:"difference"`
	}
	
	var unsyncAccounts []UnsyncAccount
	db.Raw(`
		SELECT 
			cb.code as cash_bank_code,
			cb.name as cash_bank_name,
			cb.balance as cash_bank_balance,
			acc.code as gl_code,
			acc.balance as gl_balance,
			cb.balance - acc.balance as difference
		FROM cash_banks cb 
		INNER JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL 
		  AND cb.balance != acc.balance
		ORDER BY cb.type, cb.code
		LIMIT 10
	`).Scan(&unsyncAccounts)
	
	// Log sample of unsynchronized accounts
	log.Println("Sample unsynchronized accounts:")
	for _, account := range unsyncAccounts {
		log.Printf("  %s (%s): CB=%.2f, GL=%.2f, Diff=%.2f", 
			account.CashBankCode, account.CashBankName,
			account.CashBankBalance, account.GLBalance, account.Difference)
	}
	
	if len(unsyncAccounts) < int(unsyncCount) {
		log.Printf("  ... and %d more accounts", unsyncCount-int64(len(unsyncAccounts)))
	}
	
	// Begin transaction for safe bulk update
	tx := db.Begin()
	
	// Synchronize GL account balances with CashBank balances
	log.Println("Synchronizing GL account balances with CashBank balances...")
	
	// Use a single UPDATE query to sync all unsynchronized accounts
	err := tx.Exec(`
		UPDATE accounts 
		SET balance = cb.balance,
		    updated_at = CURRENT_TIMESTAMP
		FROM cash_banks cb 
		WHERE accounts.id = cb.account_id 
		  AND cb.deleted_at IS NULL
		  AND accounts.balance != cb.balance
	`).Error
	
	if err != nil {
		log.Printf("❌ Failed to synchronize balances: %v", err)
		tx.Rollback()
		return
	}
	
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ Failed to commit balance synchronization: %v", err)
		return
	}
	
	// Verify synchronization completed successfully
	var remainingUnsyncCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM cash_banks cb 
		INNER JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL 
		  AND cb.balance != acc.balance
	`).Scan(&remainingUnsyncCount)
	
	if remainingUnsyncCount == 0 {
		log.Printf("✅ Successfully synchronized %d CashBank-GL account pairs", unsyncCount)
		log.Println("✅ All CashBank accounts are now synchronized with their GL accounts")
	} else {
		log.Printf("⚠️  Warning: %d accounts still remain unsynchronized after migration", remainingUnsyncCount)
	}
	
	log.Println("CashBank-GL Balance Synchronization completed")
}

// RunBalanceSyncFix performs comprehensive balance synchronization checks and fixes
// This function runs after every migration to ensure balance consistency across the system
func RunBalanceSyncFix(db *gorm.DB) {
	log.Println("🔧 Starting Comprehensive Balance Synchronization Fix...")
	
	// Check if required tables exist
	var cashBankTableExists, accountsTableExists bool
	
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'cash_banks'
	)`).Scan(&cashBankTableExists)
	
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'accounts'
	)`).Scan(&accountsTableExists)
	
	if !cashBankTableExists || !accountsTableExists {
		log.Println("Required tables not found, skipping comprehensive balance sync fix")
		return
	}
	
	// Step 1: Fix missing account_id relationships
	fixMissingAccountRelationships(db)
	
	// Step 2: Recalculate CashBank balances from transactions
	recalculateCashBankBalances(db)
	
	// Step 3: Ensure GL accounts match CashBank balances
	ensureGLAccountSync(db)
	
	// Step 4: Validate and report final synchronization status
	validateFinalSyncStatus(db)
	
	log.Println("✅ Comprehensive Balance Synchronization Fix completed")
}

// fixMissingAccountRelationships ensures all CashBank accounts have proper GL account relationships
func fixMissingAccountRelationships(db *gorm.DB) {
	log.Println("Step 1: Fixing missing account relationships...")
	
	// Check for CashBank accounts without GL account links
	var orphanedCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM cash_banks cb 
		LEFT JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL 
		  AND (cb.account_id IS NULL OR cb.account_id = 0 OR acc.id IS NULL)
	`).Scan(&orphanedCount)
	
	if orphanedCount > 0 {
		log.Printf("Found %d CashBank accounts without proper GL account links", orphanedCount)
		
		// Get details of orphaned accounts
		type OrphanedAccount struct {
			CashBankID   uint    `json:"cash_bank_id"`
			CashBankCode string  `json:"cash_bank_code"`
			CashBankName string  `json:"cash_bank_name"`
			AccountID    *uint   `json:"account_id"`
			Balance      float64 `json:"balance"`
		}
		
		var orphanedAccounts []OrphanedAccount
		db.Raw(`
			SELECT 
				cb.id as cash_bank_id,
				cb.code as cash_bank_code,
				cb.name as cash_bank_name,
				cb.account_id,
				cb.balance
			FROM cash_banks cb 
			LEFT JOIN accounts acc ON cb.account_id = acc.id
			WHERE cb.deleted_at IS NULL 
			  AND (cb.account_id IS NULL OR cb.account_id = 0 OR acc.id IS NULL)
			ORDER BY cb.type, cb.code
		`).Scan(&orphanedAccounts)
		
		// Log orphaned accounts for manual review
		log.Println("Orphaned CashBank accounts:")
		for _, account := range orphanedAccounts {
			log.Printf("  ID=%d, Code=%s, Name=%s, Balance=%.2f, AccountID=%v", 
				account.CashBankID, account.CashBankCode, account.CashBankName, 
				account.Balance, account.AccountID)
		}
		
		log.Printf("⚠️  Warning: Found %d orphaned CashBank accounts requiring manual GL account assignment", orphanedCount)
	} else {
		log.Println("✅ All CashBank accounts have proper GL account relationships")
	}
}

// recalculateCashBankBalances recalculates CashBank balances from transaction history
func recalculateCashBankBalances(db *gorm.DB) {
	log.Println("Step 2: Recalculating CashBank balances from transaction history...")
	
	// Check if cash_bank_transactions table exists
	var transactionTableExists bool
	db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'cash_bank_transactions'
	)`).Scan(&transactionTableExists)
	
	if !transactionTableExists {
		log.Println("Cash bank transactions table not found, skipping balance recalculation")
		return
	}
	
	// Recalculate balances for all CashBank accounts
	err := db.Exec(`
		UPDATE cash_banks 
		SET balance = COALESCE((
			SELECT SUM(amount) 
			FROM cash_bank_transactions cbt 
			WHERE cbt.cash_bank_id = cash_banks.id 
			  AND cbt.deleted_at IS NULL
		), 0),
		    updated_at = CURRENT_TIMESTAMP
		WHERE deleted_at IS NULL
	`).Error
	
	if err != nil {
		log.Printf("Warning: Failed to recalculate CashBank balances: %v", err)
	} else {
		log.Println("✅ Recalculated CashBank balances from transaction history")
	}
}

// ensureGLAccountSync ensures GL accounts are synchronized with CashBank balances
func ensureGLAccountSync(db *gorm.DB) {
	log.Println("Step 3: Ensuring GL accounts are synchronized with CashBank balances...")
	
	// Check for unsynchronized accounts
	var unsyncCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM cash_banks cb 
		INNER JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL 
		  AND ABS(cb.balance - acc.balance) > 0.01  -- Allow for small rounding differences
	`).Scan(&unsyncCount)
	
	if unsyncCount == 0 {
		log.Println("✅ All GL accounts are already synchronized with CashBank balances")
		return
	}
	
	log.Printf("Found %d GL accounts that need synchronization", unsyncCount)
	
	// Get details of unsynchronized accounts
	type UnsyncGLAccount struct {
		CashBankID      uint    `json:"cash_bank_id"`
		CashBankCode    string  `json:"cash_bank_code"`
		CashBankName    string  `json:"cash_bank_name"`
		CashBankBalance float64 `json:"cash_bank_balance"`
		GLAccountID     uint    `json:"gl_account_id"`
		GLCode          string  `json:"gl_code"`
		GLBalance       float64 `json:"gl_balance"`
		Difference      float64 `json:"difference"`
	}
	
	var unsyncAccounts []UnsyncGLAccount
	db.Raw(`
		SELECT 
			cb.id as cash_bank_id,
			cb.code as cash_bank_code,
			cb.name as cash_bank_name,
			cb.balance as cash_bank_balance,
			acc.id as gl_account_id,
			acc.code as gl_code,
			acc.balance as gl_balance,
			cb.balance - acc.balance as difference
		FROM cash_banks cb 
		INNER JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL 
		  AND ABS(cb.balance - acc.balance) > 0.01
		ORDER BY ABS(cb.balance - acc.balance) DESC
		LIMIT 10
	`).Scan(&unsyncAccounts)
	
	// Log accounts that will be synchronized
	log.Println("Accounts to be synchronized:")
	for _, account := range unsyncAccounts {
		log.Printf("  CB: %s (%.2f) -> GL: %s (%.2f) | Diff: %.2f", 
			account.CashBankCode, account.CashBankBalance,
			account.GLCode, account.GLBalance, account.Difference)
	}
	
	if len(unsyncAccounts) < int(unsyncCount) {
		log.Printf("  ... and %d more accounts", unsyncCount-int64(len(unsyncAccounts)))
	}
	
	// Begin transaction for safe bulk update
	tx := db.Begin()
	
	// Synchronize GL account balances with CashBank balances
	err := tx.Exec(`
		UPDATE accounts 
		SET balance = cb.balance,
		    updated_at = CURRENT_TIMESTAMP
		FROM cash_banks cb 
		WHERE accounts.id = cb.account_id 
		  AND cb.deleted_at IS NULL
		  AND ABS(cb.balance - accounts.balance) > 0.01
	`).Error
	
	if err != nil {
		log.Printf("❌ Failed to synchronize GL accounts: %v", err)
		tx.Rollback()
		return
	}
	
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ Failed to commit GL account synchronization: %v", err)
		return
	}
	
	log.Printf("✅ Successfully synchronized %d GL accounts with CashBank balances", unsyncCount)
}

// validateFinalSyncStatus performs final validation and reports synchronization status
func validateFinalSyncStatus(db *gorm.DB) {
	log.Println("Step 4: Validating final synchronization status...")
	
	// Count total CashBank accounts
	var totalCount int64
	db.Raw(`SELECT COUNT(*) FROM cash_banks WHERE deleted_at IS NULL`).Scan(&totalCount)
	
	// Count synchronized accounts
	var syncedCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM cash_banks cb 
		INNER JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL 
		  AND ABS(cb.balance - acc.balance) <= 0.01  -- Allow for small rounding differences
	`).Scan(&syncedCount)
	
	// Count unsynchronized accounts
	var unsyncedCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM cash_banks cb 
		INNER JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL 
		  AND ABS(cb.balance - acc.balance) > 0.01
	`).Scan(&unsyncedCount)
	
	// Count orphaned accounts (no GL account link)
	var orphanedCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM cash_banks cb 
		LEFT JOIN accounts acc ON cb.account_id = acc.id
		WHERE cb.deleted_at IS NULL 
		  AND (cb.account_id IS NULL OR cb.account_id = 0 OR acc.id IS NULL)
	`).Scan(&orphanedCount)
	
	// Report final status
	log.Println("=== Final Balance Synchronization Status ===")
	log.Printf("Total CashBank accounts: %d", totalCount)
	log.Printf("Synchronized accounts: %d", syncedCount)
	log.Printf("Unsynchronized accounts: %d", unsyncedCount)
	log.Printf("Orphaned accounts (no GL link): %d", orphanedCount)
	
	syncPercentage := float64(syncedCount) / float64(totalCount) * 100
	log.Printf("Synchronization rate: %.1f%%", syncPercentage)
	
	if unsyncedCount == 0 && orphanedCount == 0 {
		log.Println("✅ Perfect synchronization achieved! All accounts are properly synced.")
	} else if syncPercentage >= 95 {
		log.Printf("✅ Excellent synchronization (%.1f%%). System operating normally.", syncPercentage)
	} else if syncPercentage >= 85 {
		log.Printf("✅ Good synchronization (%.1f%%). Minor discrepancies are within acceptable range.", syncPercentage)
	} else if syncPercentage >= 70 {
		log.Printf("⚠️  Moderate synchronization (%.1f%%). Some accounts need attention.", syncPercentage)
	} else {
		log.Printf("❌ Poor synchronization (%.1f%%). Significant issues detected, manual intervention required.", syncPercentage)
	}
	
	// If there are still issues, log them for investigation
	if unsyncedCount > 0 || orphanedCount > 0 {
		log.Println("\n⚠️  Accounts requiring attention:")
		
		if unsyncedCount > 0 {
			var problemAccounts []struct {
				CashBankCode    string  `json:"cash_bank_code"`
				CashBankName    string  `json:"cash_bank_name"`
				CashBankBalance float64 `json:"cash_bank_balance"`
				GLCode          string  `json:"gl_code"`
				GLBalance       float64 `json:"gl_balance"`
				Difference      float64 `json:"difference"`
			}
			
			db.Raw(`
				SELECT 
					cb.code as cash_bank_code,
					cb.name as cash_bank_name,
					cb.balance as cash_bank_balance,
					acc.code as gl_code,
					acc.balance as gl_balance,
					cb.balance - acc.balance as difference
				FROM cash_banks cb 
				INNER JOIN accounts acc ON cb.account_id = acc.id
				WHERE cb.deleted_at IS NULL 
				  AND ABS(cb.balance - acc.balance) > 0.01
				ORDER BY ABS(cb.balance - acc.balance) DESC
				LIMIT 5
			`).Scan(&problemAccounts)
			
			log.Printf("  Unsynchronized accounts (top 5):")
			for _, account := range problemAccounts {
				log.Printf("    %s: CB=%.2f, GL=%.2f, Diff=%.2f", 
					account.CashBankCode, account.CashBankBalance, 
					account.GLBalance, account.Difference)
			}
		}
		
		if orphanedCount > 0 {
			var orphanedAccounts []struct {
				CashBankCode string  `json:"cash_bank_code"`
				CashBankName string  `json:"cash_bank_name"`
				Balance      float64 `json:"balance"`
				AccountID    *uint   `json:"account_id"`
			}
			
			db.Raw(`
				SELECT 
					cb.code as cash_bank_code,
					cb.name as cash_bank_name,
					cb.balance,
					cb.account_id
				FROM cash_banks cb 
				LEFT JOIN accounts acc ON cb.account_id = acc.id
				WHERE cb.deleted_at IS NULL 
				  AND (cb.account_id IS NULL OR cb.account_id = 0 OR acc.id IS NULL)
				ORDER BY cb.balance DESC
				LIMIT 5
			`).Scan(&orphanedAccounts)
			
			log.Printf("  Orphaned accounts (top 5):")
			for _, account := range orphanedAccounts {
				log.Printf("    %s: Balance=%.2f, AccountID=%v", 
					account.CashBankCode, account.Balance, account.AccountID)
			}
		}
	}
	
	log.Println("=== Balance Synchronization Validation Complete ===")
}
