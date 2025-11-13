package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	_ = godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		host := getEnvOrDefault("DB_HOST", "localhost")
		port := getEnvOrDefault("DB_PORT", "5432")
		user := getEnvOrDefault("DB_USER", "postgres")
		password := getEnvOrDefault("DB_PASSWORD", "postgres")
		dbname := getEnvOrDefault("DB_NAME", "sistem_akuntansi")
		sslmode := getEnvOrDefault("DB_SSLMODE", "disable")
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, password, host, port, dbname, sslmode)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	log.Println("============================================")
	log.Println("Investigating PENDAPATAN PENJUALAN Balance")
	log.Println("============================================")

	// Check all journal lines for account 4101
	log.Println("\n📊 All journal lines for account 4101 (PENDAPATAN PENJUALAN):")
	rows, err := db.Query(`
		SELECT 
			uje.id,
			uje.source_type,
			uje.entry_date,
			uje.description,
			ujl.debit_amount,
			ujl.credit_amount,
			uje.status,
			uje.created_at
		FROM unified_journal_lines ujl
		INNER JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE ujl.account_id = (SELECT id FROM accounts WHERE code = '4101')
		ORDER BY uje.entry_date, uje.id
	`)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer rows.Close()

	log.Println("\nID | Source     | Date       | Debit      | Credit     | Status | Description")
	log.Println("---|------------|------------|------------|------------|--------|---------------------------")
	
	totalDebit := 0.0
	totalCredit := 0.0
	for rows.Next() {
		var id int
		var sourceType, description, status string
		var entryDate, createdAt string
		var debit, credit float64
		
		if err := rows.Scan(&id, &sourceType, &entryDate, &description, &debit, &credit, &status, &createdAt); err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		
		log.Printf("%-3d| %-10s | %s | %10.2f | %10.2f | %-6s | %s", 
			id, sourceType, entryDate[:10], debit, credit, status, description)
		
		if status == "POSTED" {
			totalDebit += debit
			totalCredit += credit
		}
	}
	
	// Use correct formula: Debit - Credit (will be negative for revenue)
	balance := totalDebit - totalCredit
	log.Printf("\n📈 TOTALS (POSTED only):")
	log.Printf("   Total Debit:  %12.2f", totalDebit)
	log.Printf("   Total Credit: %12.2f", totalCredit)
	log.Printf("   Balance:      %12.2f (Debit - Credit)", balance)
	log.Printf("   Expected:     Negative for Revenue (credit balance)")
	
	if balance < 0 {
		log.Printf("   ✅ Balance is correct (Negative = Credit balance)")
	} else {
		log.Printf("   ⚠️  WARNING: Balance is positive (should be negative)")
	}

	// Check existing CLOSING entries
	log.Println("\n\n📋 Existing CLOSING journal entries:")
	closingRows, err := db.Query(`
		SELECT 
			id,
			entry_date,
			description,
			total_debit,
			total_credit,
			status,
			created_at
		FROM unified_journal_ledger
		WHERE source_type = 'CLOSING'
		ORDER BY entry_date
	`)
	if err != nil {
		log.Fatalf("Failed to query closings: %v", err)
	}
	defer closingRows.Close()

	log.Println("\nID | Date       | Total Debit | Total Credit | Status | Created At          | Description")
	log.Println("---|------------|-------------|--------------|--------|---------------------|--------------------")
	
	for closingRows.Next() {
		var id int
		var entryDate, description, status, createdAt string
		var totalDebit, totalCredit float64
		
		if err := closingRows.Scan(&id, &entryDate, &description, &totalDebit, &totalCredit, &status, &createdAt); err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		
		log.Printf("%-3d| %s | %11.2f | %12.2f | %-6s | %s | %s", 
			id, entryDate[:10], totalDebit, totalCredit, status, createdAt[:19], description)
	}

	log.Println("\n============================================")
}
