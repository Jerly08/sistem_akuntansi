package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
)

func main() {
	// Load environment variables or use defaults
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5432")
	dbUser := getEnvOrDefault("DB_USER", "postgres")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "password")
	dbName := getEnvOrDefault("DB_NAME", "accounting_db")
	
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		dbHost, dbUser, dbPassword, dbName, dbPort)
	
	fmt.Println("Connecting to database...")
	fmt.Printf("DSN: host=%s user=%s dbname=%s port=%s\n", dbHost, dbUser, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("✅ Connected to database successfully!")
	fmt.Println("\n" + strings.Repeat("=", 60))
	
	// Test 1: Check if invoice types exist
	fmt.Println("TEST 1: Checking Invoice Types")
	fmt.Println(strings.Repeat("-", 40))
	
	var invoiceTypes []models.InvoiceType
	if err := db.Find(&invoiceTypes).Error; err != nil {
		log.Fatal("Failed to get invoice types:", err)
	}
	
	fmt.Printf("Found %d invoice types:\n", len(invoiceTypes))
	for _, invType := range invoiceTypes {
		fmt.Printf("  - ID: %d, Name: %s, Code: %s, Active: %t\n", 
			invType.ID, invType.Name, invType.Code, invType.IsActive)
	}
	
	if len(invoiceTypes) == 0 {
		fmt.Println("❌ NO INVOICE TYPES FOUND!")
		fmt.Println("Run the migration first: 037_add_invoice_types_system.sql")
		return
	}
	
	// Test 2: Check counters
	fmt.Println("\nTEST 2: Checking Invoice Counters")
	fmt.Println(strings.Repeat("-", 40))
	
	var counters []models.InvoiceCounter
	if err := db.Preload("InvoiceType").Find(&counters).Error; err != nil {
		log.Fatal("Failed to get counters:", err)
	}
	
	fmt.Printf("Found %d counters:\n", len(counters))
	for _, counter := range counters {
		fmt.Printf("  - Type: %s (%s), Year: %d, Counter: %d\n", 
			counter.InvoiceType.Name, counter.InvoiceType.Code, 
			counter.Year, counter.Counter)
	}
	
	// Test 3: Test invoice number generation
	fmt.Println("\nTEST 3: Testing Invoice Number Generation")
	fmt.Println(strings.Repeat("-", 40))
	
	invoiceService := services.NewInvoiceNumberService(db)
	
	// Test with different dates and invoice types
	testDates := []time.Time{
		time.Now(),                                    // Today
		time.Date(2025, 9, 3, 12, 0, 0, 0, time.UTC), // September 2025 (your example)
		time.Date(2025, 10, 3, 12, 0, 0, 0, time.UTC), // October 2025 (current screenshot date)
	}
	
	for _, testDate := range testDates {
		fmt.Printf("\nTesting with date: %s (Month: %d - %s)\n", 
			testDate.Format("2006-01-02"), 
			testDate.Month(), 
			models.GetRomanMonth(int(testDate.Month())))
		
		for _, invType := range invoiceTypes {
			if !invType.IsActive {
				continue
			}
			
			// Preview (without incrementing counter)
			preview, err := invoiceService.PreviewInvoiceNumber(invType.ID, testDate)
			if err != nil {
				fmt.Printf("  ❌ Error previewing %s: %v\n", invType.Code, err)
				continue
			}
			
			fmt.Printf("  %s: %s (Counter: %d)\n", 
				invType.Code, preview.InvoiceNumber, preview.Counter)
		}
	}
	
	// Test 4: Simulate actual generation (this will increment counters)
	fmt.Println("\nTEST 4: Simulating Actual Generation")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("⚠️ This will increment the counters!")
	
	// Only test with first active invoice type to avoid incrementing all
	var activeType *models.InvoiceType
	for _, invType := range invoiceTypes {
		if invType.IsActive {
			activeType = &invType
			break
		}
	}
	
	if activeType != nil {
		testDate := time.Date(2025, 9, 3, 12, 0, 0, 0, time.UTC) // September 2025
		fmt.Printf("Generating invoice number for type %s (%s) on %s:\n", 
			activeType.Name, activeType.Code, testDate.Format("2006-01-02"))
		
		// Generate actual invoice number (increments counter)
		result, err := invoiceService.GenerateInvoiceNumber(activeType.ID, testDate)
		if err != nil {
			fmt.Printf("❌ Error generating: %v\n", err)
		} else {
			fmt.Printf("✅ Generated: %s\n", result.InvoiceNumber)
			fmt.Printf("   Details: Counter=%d, Year=%d, Month=%s, Code=%s\n",
				result.Counter, result.Year, result.Month, result.TypeCode)
		}
		
		// Generate another one to test counter increment
		result2, err := invoiceService.GenerateInvoiceNumber(activeType.ID, testDate)
		if err != nil {
			fmt.Printf("❌ Error generating second: %v\n", err)
		} else {
			fmt.Printf("✅ Generated second: %s (should be counter+1)\n", result2.InvoiceNumber)
		}
	}
	
	// Test 5: Check format matches expected
	fmt.Println("\nTEST 5: Format Validation")
	fmt.Println(strings.Repeat("-", 40))
	
	// Expected format: {4 digit}/{code}/{month roman}-{year}
	// Example: 0120/STA-C/IX-2025
	expectedPattern := `^\d{4}/[A-Z0-9\-]+/[IVX]+-\d{4}$`
	fmt.Printf("Expected pattern: %s\n", expectedPattern)
	fmt.Println("Example: 0120/STA-C/IX-2025")
	
	// Test if current generation matches
	if activeType != nil {
		testDate := time.Date(2025, 9, 3, 12, 0, 0, 0, time.UTC)
		preview, err := invoiceService.PreviewInvoiceNumber(activeType.ID, testDate)
		if err == nil {
			fmt.Printf("Current format: %s\n", preview.InvoiceNumber)
			// Simple validation
			if len(preview.InvoiceNumber) > 10 && 
			   preview.InvoiceNumber[4] == '/' && 
			   preview.InvoiceNumber[len(preview.InvoiceNumber)-5] == '-' {
				fmt.Println("✅ Format looks correct!")
			} else {
				fmt.Println("❌ Format doesn't match expected pattern")
			}
		}
	}
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Testing complete!")
}

// Helper function
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}