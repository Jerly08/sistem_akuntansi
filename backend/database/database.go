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
		
		// Expenses
		&models.ExpenseCategory{},
		&models.Expense{},
		
		// Assets
		&models.Asset{},
		
		// Cash & Bank
		&models.CashBank{},
		&models.CashBankTransaction{},
		&models.Payment{},
	)
	
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
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
	
	log.Println("Database indexes created successfully")
}
