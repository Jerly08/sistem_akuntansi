package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	log.Println("🚨 EMERGENCY TRIGGER REMOVAL SCRIPT")
	log.Println("===================================")
	
	// Get database connection from environment
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "accounting_db")
	
	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	
	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}
	
	log.Printf("✅ Connected to database: %s\n", dbName)
	log.Println()
	
	// Check if trigger exists
	log.Println("🔍 Checking for problematic trigger...")
	var triggerExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_trigger 
			WHERE tgname = 'trg_refresh_account_balances'
		)
	`).Scan(&triggerExists)
	
	if err != nil {
		log.Fatalf("❌ Failed to check trigger: %v", err)
	}
	
	if !triggerExists {
		log.Println("✅ Trigger does not exist - nothing to do")
		log.Println("✅ Your database is clean!")
		return
	}
	
	log.Println("⚠️  Found problematic trigger: trg_refresh_account_balances")
	log.Println()
	
	// Drop the trigger
	log.Println("🗑️  Removing trigger...")
	_, err = db.Exec("DROP TRIGGER IF EXISTS trg_refresh_account_balances ON unified_journal_lines CASCADE")
	if err != nil {
		log.Fatalf("❌ Failed to drop trigger: %v", err)
	}
	
	log.Println("✅ Trigger removed successfully")
	log.Println()
	
	// Verify removal
	log.Println("🔍 Verifying removal...")
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_trigger 
			WHERE tgname = 'trg_refresh_account_balances'
		)
	`).Scan(&triggerExists)
	
	if err != nil {
		log.Fatalf("❌ Failed to verify: %v", err)
	}
	
	if triggerExists {
		log.Fatalf("❌ Trigger still exists after removal!")
	}
	
	log.Println("✅ Verification passed - trigger is gone")
	log.Println()
	
	// Create helper functions
	log.Println("🔧 Creating helper functions...")
	
	// Manual refresh function
	_, err = db.Exec(`
		CREATE OR REPLACE FUNCTION manual_refresh_account_balances()
		RETURNS TABLE(
			success BOOLEAN,
			message TEXT,
			refreshed_at TIMESTAMPTZ
		) AS $$
		DECLARE
			start_time TIMESTAMPTZ;
			end_time TIMESTAMPTZ;
			duration INTERVAL;
		BEGIN
			start_time := clock_timestamp();
			REFRESH MATERIALIZED VIEW account_balances;
			end_time := clock_timestamp();
			duration := end_time - start_time;
			
			RETURN QUERY SELECT 
				TRUE as success,
				format('Account balances refreshed in %s', duration) as message,
				end_time as refreshed_at;
				
		EXCEPTION WHEN OTHERS THEN
			RETURN QUERY SELECT 
				FALSE as success,
				format('Refresh failed: %s', SQLERRM) as message,
				clock_timestamp() as refreshed_at;
		END;
		$$ LANGUAGE plpgsql;
	`)
	
	if err != nil {
		log.Printf("⚠️  Failed to create manual_refresh_account_balances: %v", err)
	} else {
		log.Println("✅ Created function: manual_refresh_account_balances()")
	}
	
	// Freshness check function
	_, err = db.Exec(`
		CREATE OR REPLACE FUNCTION check_account_balances_freshness()
		RETURNS TABLE(
			last_updated TIMESTAMPTZ,
			age_minutes INTEGER,
			needs_refresh BOOLEAN
		) AS $$
		DECLARE
			last_update TIMESTAMPTZ;
			age_mins INTEGER;
		BEGIN
			SELECT MAX(ab.last_updated) INTO last_update FROM account_balances ab;
			age_mins := EXTRACT(EPOCH FROM (NOW() - last_update)) / 60;
			
			RETURN QUERY SELECT 
				last_update,
				age_mins,
				age_mins > 60 as needs_refresh;
		END;
		$$ LANGUAGE plpgsql;
	`)
	
	if err != nil {
		log.Printf("⚠️  Failed to create check_account_balances_freshness: %v", err)
	} else {
		log.Println("✅ Created function: check_account_balances_freshness()")
	}
	
	log.Println()
	log.Println("=========================================")
	log.Println("✅ EMERGENCY FIX COMPLETED SUCCESSFULLY")
	log.Println("=========================================")
	log.Println()
	log.Println("What was done:")
	log.Println("  ✅ Removed trigger: trg_refresh_account_balances")
	log.Println("  ✅ Created helper functions for manual refresh")
	log.Println()
	log.Println("Next steps:")
	log.Println("  1. Test your transactions (deposit, sales, etc.)")
	log.Println("  2. Error SQLSTATE 55000 should be gone")
	log.Println("  3. No need to restart backend")
	log.Println()
	log.Println("For manual refresh:")
	log.Println("  SELECT * FROM manual_refresh_account_balances();")
	log.Println()
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
