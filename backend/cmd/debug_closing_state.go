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
	log.Println("DEBUG CLOSING STATE")
	log.Println("============================================")

	// 1. Check all CLOSING journal entries
	log.Println("\n1️⃣  All CLOSING journal entries:")
	rows, err := db.Query(`
		SELECT 
			id, entry_date, description, 
			total_debit, total_credit, status,
			created_at
		FROM unified_journal_ledger
		WHERE source_type = 'CLOSING'
		  AND deleted_at IS NULL
		ORDER BY entry_date, id
	`)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	log.Println("\n   ID | Date       | Debit       | Credit      | Status | Created At          | Description")
	log.Println("   ---|------------|-------------|-------------|--------|---------------------|---------------------------")
	
	closingCount := 0
	for rows.Next() {
		var id int
		var date, desc, status, createdAt string
		var debit, credit float64
		rows.Scan(&id, &date, &desc, &debit, &credit, &status, &createdAt)
		log.Printf("   %-3d| %s | %11.2f | %11.2f | %-6s | %s | %s",
			id, date[:10], debit, credit, status, createdAt[:19], desc)
		closingCount++
	}
	rows.Close()
	
	if closingCount == 0 {
		log.Println("   (No closing entries found)")
	} else {
		log.Printf("\n   Total: %d closing entries", closingCount)
	}

	// 2. Check revenue account journal lines (exclude CLOSING)
	log.Println("\n2️⃣  Revenue journal lines (excluding CLOSING):")
	rows, err = db.Query(`
		SELECT 
			uje.id, uje.source_type, uje.entry_date,
			ujl.account_id, a.code, a.name,
			ujl.debit_amount, ujl.credit_amount,
			uje.status
		FROM unified_journal_lines ujl
		INNER JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		INNER JOIN accounts a ON a.id = ujl.account_id
		WHERE a.type = 'REVENUE'
		  AND uje.source_type != 'CLOSING'
		  AND uje.status = 'POSTED'
		  AND uje.deleted_at IS NULL
		ORDER BY a.code, uje.entry_date, uje.id
	`)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	log.Println("\n   JID | Source | Date       | Account Code | Debit       | Credit      ")
	log.Println("   ----|--------|------------|--------------|-------------|-------------")
	
	revenueLineCount := 0
	totalRevenueDebit := 0.0
	totalRevenueCredit := 0.0
	for rows.Next() {
		var jid, accId int
		var sourceType, date, code, name, status string
		var debit, credit float64
		rows.Scan(&jid, &sourceType, &date, &accId, &code, &name, &debit, &credit, &status)
		log.Printf("   %-4d| %-6s | %s | %-12s | %11.2f | %11.2f",
			jid, sourceType, date[:10], code, debit, credit)
		revenueLineCount++
		totalRevenueDebit += debit
		totalRevenueCredit += credit
	}
	rows.Close()
	
	log.Printf("\n   Total: %d lines", revenueLineCount)
	log.Printf("   Sum Debit:  %12.2f", totalRevenueDebit)
	log.Printf("   Sum Credit: %12.2f", totalRevenueCredit)
	log.Printf("   Balance:    %12.2f (Debit - Credit)", totalRevenueDebit-totalRevenueCredit)

	// 3. Check current account balances
	log.Println("\n3️⃣  Current account balances:")
	rows, err = db.Query(`
		SELECT code, name, type, balance
		FROM accounts
		WHERE type IN ('REVENUE', 'EXPENSE')
		  AND is_header = false
		  AND deleted_at IS NULL
		ORDER BY type, code
	`)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	log.Println("\n   Code    | Name                        | Type    | Balance      ")
	log.Println("   --------|-----------------------------|---------|--------------")
	
	for rows.Next() {
		var code, name, accType string
		var balance float64
		rows.Scan(&code, &name, &accType, &balance)
		log.Printf("   %-7s | %-27s | %-7s | %12.2f", code, name, accType, balance)
	}
	rows.Close()

	log.Println("\n============================================")
	log.Println("✅ Debug completed!")
	log.Println("============================================")
	log.Println("\n📊 Analysis:")
	if closingCount > 0 {
		log.Printf("   ⚠️  Found %d closing entries - these should be deleted before retry", closingCount)
		log.Println("   Run: go run cmd/verify_and_fix_pc.go")
	} else {
		log.Println("   ✅ No closing entries found")
	}
	
	expectedBalance := totalRevenueDebit - totalRevenueCredit
	log.Printf("\n   Expected revenue balance (from journals): %.2f", expectedBalance)
	log.Println("   This should match the account balance shown above")
}
