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
	
	// Update existing purchase data with new payment tracking fields
	UpdateExistingPurchaseData(db)
	
	log.Println("Database initialization completed")
}

// RunMigrations creates all tables based on models
func RunMigrations(db *gorm.DB) {
	log.Println("Running database migrations...")
	
		err := db.AutoMigrate(
			// Core models
			&models.User{},
			&models.CompanyProfile{},
			
			// Auth models
			&models.RefreshToken{},
			&models.UserSession{},
			&models.BlacklistedToken{},
			&models.AuthAttempt{},
			&models.RateLimitRecord{},
			&models.Permission{},
			&models.RolePermission{},
			
			// Approval models
			&models.ApprovalWorkflow{},
			&models.ApprovalStep{},
			&models.ApprovalRequest{},
			&models.ApprovalAction{},
			&models.ApprovalHistory{},
			
			// Accounting models
			&models.Account{},
			&models.Transaction{},
			&models.Journal{},
			&models.JournalEntry{},
			
			// Product & Inventory models
			&models.ProductCategory{},
			&models.ProductUnit{},
			&models.Product{},
			&models.Inventory{},
			
			// Contact models
			&models.Contact{},
			&models.ContactAddress{},
			&models.ContactHistory{},
			&models.CommunicationLog{},
			
			// Sales & Purchase models
			&models.Sale{},
			&models.SaleItem{},
			&models.SalePayment{},
			&models.SaleReturn{},
			&models.SaleReturnItem{},
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
	db.Exec("CREATE INDEX IF NOT EXISTS idx_contact_addresses_type_default ON contact_addresses(contact_id, type, is_default)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_contact_history_contact_user ON contact_histories(contact_id, user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_contact_history_action_date ON contact_histories(action, created_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_communication_logs_contact_type ON communication_logs(contact_id, type)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_communication_logs_status_date ON communication_logs(status, created_at)")
	
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

// UpdateExistingPurchaseData updates existing purchase records with new payment tracking fields
func UpdateExistingPurchaseData(db *gorm.DB) {
	log.Println("Updating existing purchase data with payment tracking fields...")
	
	// Update existing purchases where outstanding_amount is 0 or null
	// Set outstanding_amount to total_amount for unpaid purchases
	result := db.Exec(`
		UPDATE purchases 
		SET outstanding_amount = total_amount,
			paid_amount = 0,
			matching_status = 'PENDING'
		WHERE (outstanding_amount IS NULL OR outstanding_amount = 0)
			AND total_amount > 0
	`)
	
	if result.Error != nil {
		log.Printf("Error updating existing purchase data: %v", result.Error)
	} else {
		log.Printf("Updated %d purchase records with payment tracking fields", result.RowsAffected)
	}
	
	log.Println("Existing purchase data update completed")
}
