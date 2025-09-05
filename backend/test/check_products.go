package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"app-sistem-akuntansi/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Use environment variables or hardcoded values for testing
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "sistem_akuntansi"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	
	// Connect to database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost,
		dbUser,
		dbPassword,
		dbName,
		dbPort,
	)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	// Query all products
	var products []models.Product
	err = db.Find(&products).Error
	if err != nil {
		log.Fatal("Failed to query products:", err)
	}
	
	fmt.Println("=== PRODUCTS IN DATABASE ===")
	fmt.Printf("Total products: %d\n\n", len(products))
	
	for _, p := range products {
		fmt.Printf("ID: %d | Code: %s | Name: %s | Unit: %s | Purchase Price: %.2f | Stock: %d\n",
			p.ID, p.Code, p.Name, p.Unit, p.PurchasePrice, p.Stock)
	}
	
	// Check if product with ID 19 exists
	var product19 models.Product
	err = db.First(&product19, 19).Error
	if err != nil {
		fmt.Printf("\n❌ Product with ID 19 NOT FOUND: %v\n", err)
	} else {
		fmt.Printf("\n✅ Product with ID 19 FOUND:\n")
		data, _ := json.MarshalIndent(product19, "", "  ")
		fmt.Println(string(data))
	}
}
