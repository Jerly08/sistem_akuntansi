package database

import (
	"log"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

// InitializeDatabase runs migrations and seeds initial data
func InitializeDatabase(db *gorm.DB) {
	log.Println("Initializing database...")
	
	// Run migrations
	RunMigrations(db)
	
	// Seed initial data
	SeedData(db)
	
	log.Println("Database initialization completed")
}

// RunMigrations creates all tables based on models
func RunMigrations(db *gorm.DB) {
	log.Println("Running database migrations...")
	
	err := db.AutoMigrate(
		// Core models
		&models.User{},
		&models.CompanyProfile{},
		
		// Accounting models
		&models.Account{},
		&models.Transaction{},
		&models.Journal{},
		&models.JournalEntry{},
		
		// Product & Inventory models
		&models.ProductCategory{},
		&models.Product{},
		&models.Inventory{},
		
		// Contact models
		&models.Contact{},
		&models.ContactAddress{},
		
		// Sales & Purchase models
		&models.Sale{},
		&models.SaleItem{},
		&models.Purchase{},
		&models.PurchaseItem{},
		
		// Payment & Cash Bank models
		&models.Payment{},
		&models.CashBank{},
		&models.CashBankTransaction{},
		
		// Expense models
		&models.ExpenseCategory{},
		&models.Expense{},
		
		// Asset models
		&models.Asset{},
		
		// Budget models
		&models.Budget{},
		&models.BudgetItem{},
		&models.BudgetComparison{},
		
		// Report models
		&models.Report{},
		&models.ReportTemplate{},
		&models.FinancialRatio{},
		&models.AccountBalance{},
		
		// Audit models
		&models.AuditLog{},
	)

	if err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}
	
	log.Println("Database migrations completed successfully")
}

// CreateIndexes creates additional database indexes for performance optimization
func CreateIndexes(db *gorm.DB) {
	log.Println("Creating additional database indexes...")
	
	// Account indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_accounts_type_category ON accounts(type, category)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_accounts_parent_level ON accounts(parent_id, level)")
	
	// Transaction indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_transactions_date_account ON transactions(transaction_date, account_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_transactions_reference ON transactions(reference_type, reference_id)")
	
	// Journal indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_journals_period_status ON journals(period, status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_journal_entries_amounts ON journal_entries(debit_amount, credit_amount)")
	
	// Sales & Purchase indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sales_date_customer ON sales(date, customer_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_purchases_date_vendor ON purchases(date, vendor_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_sale_items_product ON sale_items(product_id, sale_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_purchase_items_product ON purchase_items(product_id, purchase_id)")
	
	// Inventory indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_inventory_product_date ON inventories(product_id, transaction_date)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_inventory_reference ON inventories(reference_type, reference_id)")
	
	// Contact indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_contacts_type_category ON contacts(type, category)")
	
	// Payment indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_payments_date_contact ON payments(date, contact_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_cash_bank_transactions_date ON cash_bank_transactions(transaction_date, cash_bank_id)")
	
	// Expense indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_expenses_date_category ON expenses(date, category_id)")
	
	// Budget indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_budget_items_budget_month ON budget_items(budget_id, month)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_budget_items_account ON budget_items(account_id, budget_id)")
	
	// Report indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_reports_type_period ON reports(type, period)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_account_balances_period ON account_balances(period, account_id)")
	
	// Audit Log indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_table_record ON audit_logs(table_name, record_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_user_action ON audit_logs(user_id, action)")
	
	log.Println("Additional database indexes created successfully")
}
