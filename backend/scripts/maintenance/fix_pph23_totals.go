// fix_pph23_totals.go
// Script to fix old sales data where PPh23 was not properly deducted from TotalAmount
// Run with: go run backend/scripts/maintenance/fix_pph23_totals.go

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Sale struct {
	ID                 uint    `gorm:"primaryKey"`
	Code               string
	InvoiceNumber      string
	Subtotal           float64
	DiscountPercent    float64
	DiscountAmount     float64
	TaxableAmount      float64
	PPNRate            float64
	PPNPercent         float64
	PPN                float64
	PPNAmount          float64
	PPh21Rate          float64
	PPh21Amount        float64
	PPh23Rate          float64
	PPh23Amount        float64
	OtherTaxAdditions  float64
	OtherTaxDeductions float64
	TotalTaxDeductions float64
	ShippingCost       float64
	TotalAmount        float64
	PaidAmount         float64
	OutstandingAmount  float64
}

func main() {
	// Load environment variables
	if err := godotenv.Load("backend/.env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Get database connection string
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "accounting_db")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Println("🔧 PPh23 Total Amount Fix Script")
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Println()

	// Find all sales with PPh23 rate > 0
	var salesWithPPh23 []Sale
	if err := db.Table("sales").
		Where("pph23_rate > 0 AND deleted_at IS NULL").
		Find(&salesWithPPh23).Error; err != nil {
		log.Fatalf("Failed to query sales: %v", err)
	}

	fmt.Printf("📊 Found %d sales with PPh23\n\n", len(salesWithPPh23))

	if len(salesWithPPh23) == 0 {
		fmt.Println("✅ No sales with PPh23 found. Nothing to fix.")
		return
	}

	// Analyze and fix each sale
	fixedCount := 0
	for _, sale := range salesWithPPh23 {
		// Recalculate what the total should be
		taxableAmount := sale.Subtotal - sale.DiscountAmount
		if taxableAmount == 0 && sale.TaxableAmount > 0 {
			taxableAmount = sale.TaxableAmount
		}

		// Calculate PPN
		ppnRate := sale.PPNRate
		if ppnRate == 0 {
			ppnRate = sale.PPNPercent
		}
		ppnAmount := taxableAmount * ppnRate / 100

		// Calculate PPh21
		pph21Amount := float64(0)
		if sale.PPh21Rate > 0 {
			pph21Amount = taxableAmount * sale.PPh21Rate / 100
		}

		// Calculate PPh23
		pph23Amount := float64(0)
		if sale.PPh23Rate > 0 {
			pph23Amount = taxableAmount * sale.PPh23Rate / 100
		}

		// Calculate total deductions
		totalDeductions := pph21Amount + pph23Amount + sale.OtherTaxDeductions

		// Calculate correct total
		correctTotal := taxableAmount + ppnAmount + sale.OtherTaxAdditions + sale.ShippingCost - totalDeductions

		// Check if current total is wrong
		diff := sale.TotalAmount - correctTotal
		if diff > 1 { // Allow small rounding differences
			fmt.Printf("❌ Sale %s (Invoice: %s)\n", sale.Code, sale.InvoiceNumber)
			fmt.Printf("   Subtotal: Rp %.0f\n", sale.Subtotal)
			fmt.Printf("   Taxable Amount: Rp %.0f\n", taxableAmount)
			fmt.Printf("   PPN (%.1f%%): Rp %.0f\n", ppnRate, ppnAmount)
			fmt.Printf("   PPh23 (%.1f%%): Rp %.0f\n", sale.PPh23Rate, pph23Amount)
			fmt.Printf("   Current Total: Rp %.0f\n", sale.TotalAmount)
			fmt.Printf("   Correct Total: Rp %.0f\n", correctTotal)
			fmt.Printf("   Difference: Rp %.0f (PPh23 not deducted)\n", diff)

			// Update the sale
			updates := map[string]interface{}{
				"taxable_amount":      taxableAmount,
				"ppn":                 ppnAmount,
				"ppn_amount":          ppnAmount,
				"pph21_amount":        pph21Amount,
				"pph23_amount":        pph23Amount,
				"total_tax_deductions": totalDeductions,
				"total_amount":        correctTotal,
				"outstanding_amount":  correctTotal - sale.PaidAmount,
			}

			if err := db.Table("sales").Where("id = ?", sale.ID).Updates(updates).Error; err != nil {
				fmt.Printf("   ⚠️ Failed to update: %v\n", err)
			} else {
				fmt.Printf("   ✅ Fixed! New Total: Rp %.0f\n", correctTotal)
				fixedCount++
			}
			fmt.Println()
		} else {
			fmt.Printf("✅ Sale %s (Invoice: %s) - Total is correct: Rp %.0f\n", 
				sale.Code, sale.InvoiceNumber, sale.TotalAmount)
		}
	}

	fmt.Println()
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Printf("📊 Summary: Fixed %d out of %d sales with PPh23\n", fixedCount, len(salesWithPPh23))
	fmt.Println("=" + string(make([]byte, 60)))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
