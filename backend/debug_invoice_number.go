package main

import (
	"fmt"
	"log"
	"os"
	"time"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
)

func main() {
	// Database connection (adjust according to your config)
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=your_password dbname=accounting_db port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create invoice number service
	invoiceService := services.NewInvoiceNumberService(db)

	// Get all active invoice types
	var invoiceTypes []models.InvoiceType
	if err := db.Where("is_active = ?", true).Find(&invoiceTypes).Error; err != nil {
		log.Fatal("Failed to get invoice types:", err)
	}

	fmt.Println("=== DEBUGGING INVOICE NUMBER GENERATION ===")
	fmt.Println("Current time:", time.Now())
	fmt.Printf("Current month: %d (%s)\n", time.Now().Month(), models.GetRomanMonth(int(time.Now().Month())))
	fmt.Println()

	// Test for each invoice type
	for _, invType := range invoiceTypes {
		fmt.Printf("Invoice Type: %s (%s)\n", invType.Name, invType.Code)
		
		// Test with today's date
		testDate := time.Now()
		fmt.Printf("Test date: %s (month: %d - %s)\n", 
			testDate.Format("2006-01-02"), 
			testDate.Month(), 
			models.GetRomanMonth(int(testDate.Month())))

		// Preview next number
		preview, err := invoiceService.PreviewInvoiceNumber(invType.ID, testDate)
		if err != nil {
			log.Printf("Error previewing for type %d: %v", invType.ID, err)
			continue
		}
		
		fmt.Printf("Preview: %s\n", preview.InvoiceNumber)
		fmt.Printf("Counter: %d, Year: %d, Month: %s, Code: %s\n", 
			preview.Counter, preview.Year, preview.Month, preview.TypeCode)
		
		// Test with September 2025 (your example date)
		septDate := time.Date(2025, 9, 3, 12, 0, 0, 0, time.Local)
		septPreview, err := invoiceService.PreviewInvoiceNumber(invType.ID, septDate)
		if err != nil {
			log.Printf("Error previewing Sept 2025 for type %d: %v", invType.ID, err)
		} else {
			fmt.Printf("September 2025 preview: %s\n", septPreview.InvoiceNumber)
		}
		
		fmt.Println("---")
	}

	// Check current counters
	fmt.Println("\n=== CURRENT COUNTERS ===")
	var counters []models.InvoiceCounter
	if err := db.Preload("InvoiceType").Find(&counters).Error; err != nil {
		log.Fatal("Failed to get counters:", err)
	}

	for _, counter := range counters {
		fmt.Printf("Type: %s (%s), Year: %d, Counter: %d\n", 
			counter.InvoiceType.Name, counter.InvoiceType.Code, 
			counter.Year, counter.Counter)
	}
}