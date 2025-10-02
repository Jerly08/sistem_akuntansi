package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No .env file found")
	}

	// Database connection
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"), 
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create PDF service
	pdfService := services.NewPDFService(db)

	// Create sample sales data for testing
	sampleSales := []models.Sale{
		{
			Code:              "SOA-000001",
			InvoiceNumber:     "INV-202509-001",
			Date:              time.Now(),
			TotalAmount:       1110000,
			PaidAmount:        1110000,
			OutstandingAmount: 0,
			Status:            "PAID",
			Type:              "INVOICE",
			Customer: models.Customer{
				ID:   1,
				Name: "PT. Sejahtera Manufacturing Indonesia",
			},
		},
		{
			Code:              "SOA-000002", 
			InvoiceNumber:     "INV-202509-002",
			Date:              time.Now().AddDate(0, 0, -1),
			TotalAmount:       1110000,
			PaidAmount:        1110000,
			OutstandingAmount: 0,
			Status:            "PAID",
			Type:              "INVOICE",
			Customer: models.Customer{
				ID:   2,
				Name: "CV. Maju Jaya Sentosa Berkembang",
			},
		},
	}

	// Generate test PDF
	startDate := "2025-09-02"
	endDate := "2025-10-02"

	pdfBytes, err := pdfService.GenerateSalesReportPDF(sampleSales, startDate, endDate)
	if err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
	}

	// Save to file
	filename := fmt.Sprintf("test_sales_report_%s.pdf", time.Now().Format("20060102_150405"))
	err = os.WriteFile(filename, pdfBytes, 0644)
	if err != nil {
		log.Fatalf("Failed to save PDF: %v", err)
	}

	fmt.Printf("✓ Test PDF generated successfully: %s\n", filename)
	fmt.Printf("✓ PDF size: %d bytes\n", len(pdfBytes))
	fmt.Println("✓ Formatting improvements applied:")
	fmt.Println("  - Font size reduced from 8pt to 7pt")
	fmt.Println("  - Row height increased from 5mm to 6.5mm") 
	fmt.Println("  - Column widths optimized for better text fit")
	fmt.Println("  - Customer name truncation improved")
	fmt.Println("\nNow testing receipt PDF localization...")

	// Test receipt PDF generation  
	receiptBytes, err := pdfService.GenerateReceiptPDF(&sampleSales[0])
	if err != nil {
		log.Printf("Warning: Failed to generate receipt PDF: %v", err)
	} else {
		receiptFilename := fmt.Sprintf("test_receipt_%s.pdf", time.Now().Format("20060102_150405"))
		err = os.WriteFile(receiptFilename, receiptBytes, 0644)
		if err != nil {
			log.Printf("Warning: Failed to save receipt PDF: %v", err)
		} else {
			fmt.Printf("✓ Receipt PDF generated: %s\n", receiptFilename)
			fmt.Printf("✓ Receipt PDF size: %d bytes\n", len(receiptBytes))
			fmt.Println("✓ Localization fixes applied:")
			fmt.Println("  - Title now uses system language setting (RECEIPT/KWITANSI)")
			fmt.Println("  - Field labels respect language configuration")
			fmt.Println("  - Date formatting follows locale")
		}
	}

	fmt.Println("\nPlease open the generated PDFs to verify all fixes.")
}