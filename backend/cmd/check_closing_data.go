package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully!")
	fmt.Println("==================================================")

	// 1. Check journal_entries table for closing entries
	fmt.Println("\n1. CHECKING JOURNAL ENTRIES FOR CLOSING ENTRIES:")
	fmt.Println("--------------------------------------------------")
	
	query := `
		SELECT id, code, description, reference_type, entry_date, total_debit, created_at
		FROM journal_entries
		WHERE code LIKE 'CLO-%' 
		   OR description LIKE '%closing%' 
		   OR description LIKE '%tutup buku%'
		   OR reference_type IN ('FISCAL_CLOSING', 'PERIOD_CLOSING', 'CLOSING')
		ORDER BY entry_date DESC
		LIMIT 10
	`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying journal_entries: %v", err)
	} else {
		defer rows.Close()
		
		count := 0
		for rows.Next() {
			var id int
			var code, description, referenceType string
			var entryDate time.Time
			var totalDebit float64
			var createdAt time.Time
			
			err := rows.Scan(&id, &code, &description, &referenceType, &entryDate, &totalDebit, &createdAt)
			if err != nil {
				log.Printf("Error scanning row: %v", err)
				continue
			}
			
			fmt.Printf("ID: %d | Code: %s | RefType: %s | Date: %s\n", id, code, referenceType, entryDate.Format("2006-01-02"))
			fmt.Printf("  Description: %s\n", description)
			fmt.Printf("  Total Debit: %.2f | Created: %s\n", totalDebit, createdAt.Format("2006-01-02 15:04:05"))
			fmt.Println()
			count++
		}
		
		if count == 0 {
			fmt.Println("No closing entries found in journal_entries table")
		} else {
			fmt.Printf("Found %d closing entries\n", count)
		}
	}

	// 2. Check all reference_types in journal_entries
	fmt.Println("\n2. ALL REFERENCE TYPES IN JOURNAL_ENTRIES:")
	fmt.Println("--------------------------------------------------")
	
	query = `
		SELECT DISTINCT reference_type, COUNT(*) as count
		FROM journal_entries
		WHERE reference_type IS NOT NULL AND reference_type != ''
		GROUP BY reference_type
		ORDER BY count DESC
	`
	
	rows, err = db.Query(query)
	if err != nil {
		log.Printf("Error querying reference types: %v", err)
	} else {
		defer rows.Close()
		
		for rows.Next() {
			var refType string
			var count int
			
			if err := rows.Scan(&refType, &count); err != nil {
				log.Printf("Error scanning: %v", err)
				continue
			}
			
			fmt.Printf("Reference Type: %-20s | Count: %d\n", refType, count)
		}
	}

	// 3. Check accounting_periods table if exists
	fmt.Println("\n3. CHECKING ACCOUNTING_PERIODS TABLE:")
	fmt.Println("--------------------------------------------------")
	
	// First check if table exists
	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'accounting_periods'
		)
	`).Scan(&tableExists)
	
	if err != nil {
		log.Printf("Error checking table existence: %v", err)
	} else if !tableExists {
		fmt.Println("Table accounting_periods does not exist")
	} else {
		query = `
			SELECT id, period_code, description, start_date, end_date, status, is_locked, created_at
			FROM accounting_periods
			ORDER BY end_date DESC
			LIMIT 10
		`
		
		rows, err = db.Query(query)
		if err != nil {
			log.Printf("Error querying accounting_periods: %v", err)
		} else {
			defer rows.Close()
			
			count := 0
			for rows.Next() {
				var id int
				var periodCode, description, status string
				var startDate, endDate time.Time
				var isLocked bool
				var createdAt time.Time
				
				err := rows.Scan(&id, &periodCode, &description, &startDate, &endDate, &status, &isLocked, &createdAt)
				if err != nil {
					log.Printf("Error scanning: %v", err)
					continue
				}
				
				fmt.Printf("ID: %d | Code: %s | Status: %s | Locked: %v\n", id, periodCode, status, isLocked)
				fmt.Printf("  Period: %s to %s\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
				fmt.Printf("  Description: %s\n", description)
				fmt.Printf("  Created: %s\n", createdAt.Format("2006-01-02 15:04:05"))
				fmt.Println()
				count++
			}
			
			if count == 0 {
				fmt.Println("No periods found in accounting_periods table")
			} else {
				fmt.Printf("Found %d periods\n", count)
			}
		}
	}

	// 4. Check for recent journal entries to see what's being created
	fmt.Println("\n4. RECENT JOURNAL ENTRIES (Last 5):")
	fmt.Println("--------------------------------------------------")
	
	query = `
		SELECT id, code, description, reference_type, entry_date, total_debit
		FROM journal_entries
		ORDER BY created_at DESC
		LIMIT 5
	`
	
	rows, err = db.Query(query)
	if err != nil {
		log.Printf("Error querying recent entries: %v", err)
	} else {
		defer rows.Close()
		
		for rows.Next() {
			var id int
			var code, description string
			var referenceType sql.NullString
			var entryDate time.Time
			var totalDebit float64
			
			err := rows.Scan(&id, &code, &description, &referenceType, &entryDate, &totalDebit)
			if err != nil {
				log.Printf("Error scanning: %v", err)
				continue
			}
			
			refType := "NULL"
			if referenceType.Valid {
				refType = referenceType.String
			}
			
			fmt.Printf("ID: %d | Code: %s | RefType: %s | Date: %s\n", id, code, refType, entryDate.Format("2006-01-02"))
			fmt.Printf("  Description: %s | Total: %.2f\n", description, totalDebit)
			fmt.Println()
		}
	}

	fmt.Println("\nDiagnostic complete!")
}