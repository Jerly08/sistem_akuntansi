package database

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

// RunAutoMigrations runs all pending SQL migrations automatically
func RunAutoMigrations(db *gorm.DB) error {
	log.Println("🔄 Starting auto-migrations...")
	
	// Create migration_logs table first if it doesn't exist
	if err := createMigrationLogsTable(db); err != nil {
		return fmt.Errorf("failed to create migration_logs table: %v", err)
	}
	
	// Get migration files
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %v", err)
	}
	
	// Run each migration
	for _, file := range migrationFiles {
		if err := runMigration(db, file); err != nil {
			log.Printf("❌ Migration failed: %s - %v", file, err)
			continue
		}
	}
	
	// Check and create Standard Purchase Approval workflow
	log.Println("============================================")
	log.Println("🔍 STARTING STANDARD PURCHASE APPROVAL WORKFLOW CHECK")
	log.Println("============================================")
	if err := ensureStandardPurchaseApprovalWorkflow(db); err != nil {
		log.Printf("⚠️  WORKFLOW AUTO-MIGRATION FAILED: %v", err)
	} else {
		log.Println("✅ WORKFLOW AUTO-MIGRATION COMPLETED SUCCESSFULLY")
	}
	log.Println("============================================")

	log.Println("✅ Auto-migrations completed")
	return nil
}

// createMigrationLogsTable creates the migration_logs table if it doesn't exist
func createMigrationLogsTable(db *gorm.DB) error {
	// Execute CREATE TABLE statement
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS migration_logs (
		id SERIAL PRIMARY KEY,
		migration_name VARCHAR(255) NOT NULL UNIQUE,
		status VARCHAR(20) NOT NULL DEFAULT 'SUCCESS' CHECK (status IN ('SUCCESS', 'FAILED', 'SKIPPED')),
		message TEXT,
		executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		execution_time_ms INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	
	if err := db.Exec(createTableSQL).Error; err != nil {
		return err
	}
	
	// Execute INDEX statements separately
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_migration_logs_name ON migration_logs(migration_name)",
		"CREATE INDEX IF NOT EXISTS idx_migration_logs_status ON migration_logs(status)", 
		"CREATE INDEX IF NOT EXISTS idx_migration_logs_executed_at ON migration_logs(executed_at)",
	}
	
	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			// Index creation failure is not critical
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}
	
	return nil
}

// getMigrationFiles gets all SQL migration files sorted by name
func getMigrationFiles() ([]string, error) {
	migrationDir, err := findMigrationDir()
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(migrationDir)
	if err != nil {
		return nil, err
	}

	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	// Sort files to ensure proper execution order
	sort.Strings(migrationFiles)

	// Log found migration files for debugging
	log.Printf("Using migration dir: %s", migrationDir)
	log.Printf("Found %d migration files: %v", len(migrationFiles), migrationFiles)

	return migrationFiles, nil
}

// findMigrationDir tries multiple locations to locate the migrations folder
func findMigrationDir() (string, error) {
	candidates := []string{}

	// 0) Explicit environment override
	if envDir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR")); envDir != "" {
		candidates = append(candidates, filepath.Clean(envDir))
	}

	// 1) Current working directory and its parents
	cwd, _ := os.Getwd()
	if cwd != "" {
		candidates = append(candidates,
			filepath.Clean(filepath.Join(cwd, "migrations")),
			filepath.Clean(filepath.Join(cwd, "backend", "migrations")),
			filepath.Clean(filepath.Join(cwd, "..", "migrations")),
			filepath.Clean(filepath.Join(cwd, "..", "backend", "migrations")),
			filepath.Clean(filepath.Join(cwd, "..", "..", "migrations")),
			filepath.Clean(filepath.Join(cwd, "..", "..", "backend", "migrations")),
		)
	}

	// 2) Relative to process (works when running from repo root or backend dir)
	candidates = append(candidates,
		filepath.Clean("./migrations"),
		filepath.Clean("backend/migrations"),
		filepath.Clean("../migrations"),
		filepath.Clean("../backend/migrations"),
		filepath.Clean("../../migrations"),
		filepath.Clean("../../backend/migrations"),
	)

	// 3) Directory next to the executable and its parents
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "migrations"),
			filepath.Join(exeDir, "backend", "migrations"),
			filepath.Join(exeDir, "..", "migrations"),
			filepath.Join(exeDir, "..", "backend", "migrations"),
			filepath.Join(exeDir, "..", "..", "migrations"),
			filepath.Join(exeDir, "..", "..", "backend", "migrations"),
		)
	}

	// Deduplicate while preserving order
	seen := map[string]struct{}{}
	unique := make([]string, 0, len(candidates))
	for _, dir := range candidates {
		if _, ok := seen[dir]; ok {
			continue
		}
		seen[dir] = struct{}{}
		unique = append(unique, dir)
	}

	for _, dir := range unique {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
	}
	return "", fmt.Errorf("migrations directory not found. Tried: %v", unique)
}

// runMigration runs a single migration file
func runMigration(db *gorm.DB, filename string) error {
	// Check if migration already ran
	var count int64
	db.Raw("SELECT COUNT(*) FROM migration_logs WHERE migration_name = ?", filename).Scan(&count)
	if count > 0 {
		log.Printf("⏭️  Migration already ran: %s", filename)
		return nil
	}
	
	startTime := time.Now()
	log.Printf("🔄 Running migration: %s", filename)
	
	// Read migration file
	migrationDir, dirErr := findMigrationDir()
	if dirErr != nil {
		logMigrationResult(db, filename, "FAILED", fmt.Sprintf("Failed to locate migrations dir: %v", dirErr), 0)
		return dirErr
	}
	migrationPath := filepath.Join(migrationDir, filename)
	content, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		logMigrationResult(db, filename, "FAILED", fmt.Sprintf("Failed to read file: %v", err), 0)
		return err
	}
	
	// Special handling for SSOT migration (contains complex SQL structures)
	if strings.Contains(filename, "unified_journal_ssot") {
		return runComplexMigration(db, filename, string(content), startTime)
	}
	
	// Split SQL statements by semicolon (simple approach)
	sqlStatements := strings.Split(string(content), ";")
	
	// Execute in transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	for _, stmt := range sqlStatements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") || strings.HasPrefix(stmt, "/*") {
			continue
		}
		
		if err := tx.Exec(stmt).Error; err != nil {
			tx.Rollback()
			executionTime := int(time.Since(startTime).Milliseconds())
			logMigrationResult(db, filename, "FAILED", fmt.Sprintf("SQL error: %v", err), executionTime)
			return err
		}
	}
	
	if err := tx.Commit().Error; err != nil {
		executionTime := int(time.Since(startTime).Milliseconds())
		logMigrationResult(db, filename, "FAILED", fmt.Sprintf("Commit error: %v", err), executionTime)
		return err
	}
	
	executionTime := int(time.Since(startTime).Milliseconds())
	logMigrationResult(db, filename, "SUCCESS", "Migration completed successfully", executionTime)
	
	log.Printf("✅ Migration completed: %s (%dms)", filename, executionTime)
	return nil
}

// runComplexMigration runs complex migrations with better SQL parsing
func runComplexMigration(db *gorm.DB, filename, content string, startTime time.Time) error {
	log.Printf("🏠 Running complex migration (SSOT): %s", filename)
	
	// Parse SQL more intelligently for complex structures
	statements := parseComplexSQL(content)
	
	// Execute each parsed statement
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		
		log.Printf("🔧 Executing statement %d/%d...", i+1, len(statements))
		
		// Execute statement (not in transaction for DDL operations)
		if err := db.Exec(stmt).Error; err != nil {
			executionTime := int(time.Since(startTime).Milliseconds())
			logMigrationResult(db, filename, "FAILED", fmt.Sprintf("SQL error at statement %d: %v", i+1, err), executionTime)
			return fmt.Errorf("SQL error at statement %d: %v", i+1, err)
		}
	}
	
	executionTime := int(time.Since(startTime).Milliseconds())
	logMigrationResult(db, filename, "SUCCESS", "Complex migration completed successfully", executionTime)
	
	log.Printf("✅ Complex migration completed: %s (%dms)", filename, executionTime)
	return nil
}

// parseComplexSQL parses SQL with complex structures like functions, triggers, etc.
func parseComplexSQL(content string) []string {
	var statements []string
	current := ""
	lines := strings.Split(content, "\n")
	
	inFunction := false
	inBlock := false
	blockDepth := 0
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		
		// Handle multi-line comments
		if strings.HasPrefix(line, "/*") {
			continue
		}
		if strings.HasSuffix(line, "*/") {
			continue
		}
		
		// Handle function/procedure blocks
		if strings.Contains(strings.ToUpper(line), "CREATE OR REPLACE FUNCTION") || 
		   strings.Contains(strings.ToUpper(line), "CREATE FUNCTION") ||
		   strings.Contains(strings.ToUpper(line), "CREATE TRIGGER") ||
		   strings.Contains(strings.ToUpper(line), "CREATE MATERIALIZED VIEW") {
			inFunction = true
		}
		
		// Handle DO blocks
		if strings.HasPrefix(strings.ToUpper(line), "DO $$") {
			inBlock = true
			blockDepth = 1
		}
		
		// Count BEGIN/END pairs in functions
		if inFunction || inBlock {
			if strings.Contains(strings.ToUpper(line), "BEGIN") {
				blockDepth++
			}
			if strings.Contains(strings.ToUpper(line), "END") {
				blockDepth--
			}
		}
		
		current += line + "\n"
		
		// Check for statement end
		if strings.HasSuffix(line, ";") {
			// For functions/triggers, only end when we're back to depth 0
			if (inFunction || inBlock) && blockDepth > 0 {
				continue
			}
			
			// Complete statement found
			stmt := strings.TrimSpace(current)
			if stmt != "" {
				statements = append(statements, stmt)
			}
			
			current = ""
			inFunction = false
			inBlock = false
			blockDepth = 0
		}
	}
	
	// Add any remaining content
	if current != "" {
		stmt := strings.TrimSpace(current)
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}
	
	return statements
}

// logMigrationResult logs the result of a migration
func logMigrationResult(db *gorm.DB, migrationName, status, message string, executionTimeMs int) {
	sql := `
	INSERT INTO migration_logs (migration_name, status, message, execution_time_ms)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (migration_name) DO UPDATE SET
		status = EXCLUDED.status,
		message = EXCLUDED.message,
		execution_time_ms = EXCLUDED.execution_time_ms,
		executed_at = CURRENT_TIMESTAMP
	`
	
	if err := db.Exec(sql, migrationName, status, message, executionTimeMs).Error; err != nil {
		log.Printf("⚠️  Failed to log migration result: %v", err)
	}
}

// GetMigrationStatus returns the status of all migrations
func GetMigrationStatus(db *gorm.DB) ([]MigrationLog, error) {
	var logs []MigrationLog
	err := db.Order("executed_at DESC").Find(&logs).Error
	return logs, err
}

// MigrationLog represents a migration log entry
type MigrationLog struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	MigrationName   string    `json:"migration_name" gorm:"size:255;uniqueIndex"`
	Status          string    `json:"status" gorm:"size:20"`
	Message         string    `json:"message" gorm:"type:text"`
	ExecutedAt      time.Time `json:"executed_at"`
	ExecutionTimeMs int       `json:"execution_time_ms"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ApprovalWorkflow represents the approval_workflows table for auto-migration
type ApprovalWorkflow struct {
	ID              uint    `gorm:"primaryKey"`
	Name            string  `gorm:"not null;size:100"`
	Module          string  `gorm:"not null;size:50"`
	MinAmount       float64 `gorm:"type:decimal(15,2);default:0"`
	MaxAmount       float64 `gorm:"type:decimal(15,2)"`
	IsActive        bool    `gorm:"default:true"`
	RequireDirector bool    `gorm:"default:false"`
	RequireFinance  bool    `gorm:"default:false"`
}

// ApprovalStep represents the approval_steps table for auto-migration
type ApprovalStep struct {
	ID           uint   `gorm:"primaryKey"`
	WorkflowID   uint   `gorm:"not null;index"`
	StepOrder    int    `gorm:"not null"`
	StepName     string `gorm:"not null;size:100"`
	ApproverRole string `gorm:"not null;size:50"`
	IsOptional   bool   `gorm:"default:false"`
	TimeLimit    int    `gorm:"default:24"`
}

// ensureStandardPurchaseApprovalWorkflow checks and creates Standard Purchase Approval workflow if it doesn't exist
func ensureStandardPurchaseApprovalWorkflow(db *gorm.DB) error {
	log.Println("🔍 Checking Standard Purchase Approval workflow...")
	
	// Check if Standard Purchase Approval workflow exists
	var existingWorkflow ApprovalWorkflow
	result := db.Where("name = ? AND module = ?", "Standard Purchase Approval", "PURCHASE").First(&existingWorkflow)
	
	if result.Error == nil {
		log.Println("✅ Standard Purchase Approval workflow found")
		
		// Check if workflow has steps
		var stepCount int64
		db.Model(&ApprovalStep{}).Where("workflow_id = ?", existingWorkflow.ID).Count(&stepCount)
		
		if stepCount == 0 {
			log.Println("⚠️  Workflow exists but has no steps - creating steps...")
			// Create steps for existing workflow
			return createWorkflowSteps(db, existingWorkflow.ID)
		} else {
			log.Printf("✅ Workflow has %d steps - no action needed", stepCount)
			return nil
		}
	}
	
	// If not found, create it
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("📝 Creating Standard Purchase Approval workflow...")
		
		// Create workflow
		workflow := ApprovalWorkflow{
			Name:            "Standard Purchase Approval",
			Module:          "PURCHASE",
			MinAmount:       0,
			MaxAmount:       999999999999,
			IsActive:        true,
			RequireDirector: true,
			RequireFinance:  true,
		}
		
		if err := db.Create(&workflow).Error; err != nil {
			return fmt.Errorf("failed to create Standard Purchase Approval workflow: %v", err)
		}
		
		log.Printf("✅ Created Standard Purchase Approval workflow with ID: %d", workflow.ID)
		
		// Create workflow steps
		return createWorkflowSteps(db, workflow.ID)
		
		return nil
	}
	
	// Other database errors
	return fmt.Errorf("failed to check existing workflow: %v", result.Error)
}

// createWorkflowSteps creates the standard approval workflow steps
func createWorkflowSteps(db *gorm.DB, workflowID uint) error {
	steps := []ApprovalStep{
		{
			WorkflowID:   workflowID,
			StepOrder:    1,
			StepName:     "Employee Submission",
			ApproverRole: "employee",
			IsOptional:   false,
			TimeLimit:    24,
		},
		{
			WorkflowID:   workflowID,
			StepOrder:    2,
			StepName:     "Finance Approval",
			ApproverRole: "finance",
			IsOptional:   false,
			TimeLimit:    48,
		},
		{
			WorkflowID:   workflowID,
			StepOrder:    3,
			StepName:     "Director Approval",
			ApproverRole: "director",
			IsOptional:   true,
			TimeLimit:    72,
		},
	}
	
	for _, step := range steps {
		if err := db.Create(&step).Error; err != nil {
			return fmt.Errorf("failed to create workflow step '%s': %v", step.StepName, err)
		}
	}
	
	log.Printf("✅ Created %d workflow steps for Standard Purchase Approval", len(steps))
	log.Println("🎯 Standard Purchase Approval workflow setup completed!")
	
	return nil
}
