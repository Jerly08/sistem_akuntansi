package database

import (
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
)

var DB *gorm.DB

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
		&models.Inventory{},
		
		// Sales
		&models.Sale{},
		&models.SaleItem{},
		
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
