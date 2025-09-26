package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("❌ DATABASE_URL environment variable is required")
	}

	// Connect to database
	fmt.Printf("🔗 Connecting to database: %s\n", databaseURL)
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("❌ Failed to ping database:", err)
	}
	fmt.Println("✅ Database connection successful!")

	fmt.Println("\n🔧 Simulating client git pull scenario...")
	fmt.Println("📝 Dropping 'description' column to simulate old database state...")

	// Drop description column to simulate client's old state
	_, err = db.Exec(`ALTER TABLE migration_logs DROP COLUMN IF EXISTS description;`)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to drop description column: %v", err)
	} else {
		fmt.Println("✅ Successfully dropped 'description' column")
	}

	// Reset some problematic migration statuses to FAILED to simulate issues
	fmt.Println("📝 Setting some migrations to FAILED status to simulate migration problems...")
	problematicMigrations := []string{
		"020_add_sales_data_integrity_constraints.sql",
		"022_comprehensive_model_updates.sql",
		"025_safe_ssot_journal_migration_fix.sql",
	}

	for _, migration := range problematicMigrations {
		_, err = db.Exec(`
			UPDATE migration_logs 
			SET status = 'FAILED', message = 'Simulated failure for testing'
			WHERE migration_name = $1
		`, migration)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to update %s: %v", migration, err)
		} else {
			fmt.Printf("✅ Set %s to FAILED status\n", migration)
		}
	}

	// Check current state
	fmt.Println("\n🔍 Checking current database state:")
	
	// Check description column
	var columnExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'migration_logs' 
			AND column_name = 'description'
		);
	`).Scan(&columnExists)
	
	if err != nil {
		log.Printf("⚠️  Failed to check description column: %v", err)
	} else {
		fmt.Printf("📊 Description column exists: %v\n", columnExists)
	}

	// Check problematic migration statuses
	fmt.Println("\n📊 Current migration statuses:")
	rows, err := db.Query(`
		SELECT migration_name, status 
		FROM migration_logs 
		WHERE migration_name IN ($1, $2, $3)
		ORDER BY migration_name
	`, problematicMigrations[0], problematicMigrations[1], problematicMigrations[2])
	
	if err != nil {
		log.Printf("⚠️  Failed to query migration statuses: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var name, status string
			rows.Scan(&name, &status)
			fmt.Printf("  - %s: %s\n", name, status)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎭 Simulation completed!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("✅ Database is now in 'client post-git-pull' state")
	fmt.Println("📝 Description column: REMOVED")
	fmt.Println("⚠️  Some migrations: FAILED status")
	fmt.Println("")
	fmt.Println("🚀 Now run the backend to test auto-fix functionality!")
	fmt.Println("   Expected behavior:")
	fmt.Println("   1. Auto-add description column")
	fmt.Println("   2. Fix problematic migration statuses")
	fmt.Println("   3. Backend starts without errors")
}