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
	err := db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Sale{},
		&models.SaleItem{},
		&models.Purchase{},
		&models.PurchaseItem{},
		&models.Expense{},
		&models.Asset{},
		&models.CashBank{},
		&models.Account{},
		&models.Inventory{},
	)
	
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	
	log.Println("Database migration completed successfully")
}
