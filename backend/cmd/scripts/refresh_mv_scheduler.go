package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// RefreshScheduler handles periodic refresh of account_balances materialized view
type RefreshScheduler struct {
	db       *sql.DB
	interval time.Duration
	stopChan chan bool
}

// RefreshResult represents the result of a materialized view refresh
type RefreshResult struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	RefreshedAt time.Time `json:"refreshed_at"`
}

// FreshnessCheck represents the freshness status of the materialized view
type FreshnessCheck struct {
	LastUpdated   time.Time `json:"last_updated"`
	AgeMinutes    int       `json:"age_minutes"`
	NeedsRefresh  bool      `json:"needs_refresh"`
}

func main() {
	// Get database connection string from environment
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "accounting_db")
	refreshInterval := getEnv("REFRESH_INTERVAL", "1h") // Default 1 hour

	// Parse refresh interval
	interval, err := time.ParseDuration(refreshInterval)
	if err != nil {
		log.Fatalf("Invalid REFRESH_INTERVAL: %v", err)
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Printf("🔗 Connected to database: %s", dbName)
	log.Printf("⏰ Refresh interval: %s", interval)

	// Create scheduler
	scheduler := NewRefreshScheduler(db, interval)

	// Start scheduler
	scheduler.Start()
}

// NewRefreshScheduler creates a new refresh scheduler
func NewRefreshScheduler(db *sql.DB, interval time.Duration) *RefreshScheduler {
	return &RefreshScheduler{
		db:       db,
		interval: interval,
		stopChan: make(chan bool),
	}
}

// Start begins the scheduled refresh process
func (s *RefreshScheduler) Start() {
	log.Println("🚀 Starting materialized view refresh scheduler...")

	// Run initial refresh
	s.checkAndRefresh()

	// Schedule periodic refreshes
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkAndRefresh()
		case <-s.stopChan:
			log.Println("🛑 Stopping refresh scheduler...")
			return
		}
	}
}

// Stop stops the scheduler
func (s *RefreshScheduler) Stop() {
	s.stopChan <- true
}

// checkAndRefresh checks if refresh is needed and performs it
func (s *RefreshScheduler) checkAndRefresh() {
	log.Println("🔍 Checking materialized view freshness...")

	// Check freshness
	freshness, err := s.checkFreshness()
	if err != nil {
		log.Printf("❌ Failed to check freshness: %v", err)
		// Attempt refresh anyway
		s.refresh()
		return
	}

	log.Printf("📊 View age: %d minutes (last updated: %s)", 
		freshness.AgeMinutes, freshness.LastUpdated.Format(time.RFC3339))

	if freshness.NeedsRefresh {
		log.Println("⚠️ Refresh needed!")
		s.refresh()
	} else {
		log.Println("✅ View is fresh, no refresh needed")
	}
}

// checkFreshness checks the freshness of the materialized view
func (s *RefreshScheduler) checkFreshness() (*FreshnessCheck, error) {
	query := "SELECT last_updated, age_minutes, needs_refresh FROM check_account_balances_freshness()"
	
	var fc FreshnessCheck
	err := s.db.QueryRow(query).Scan(&fc.LastUpdated, &fc.AgeMinutes, &fc.NeedsRefresh)
	if err != nil {
		return nil, fmt.Errorf("failed to check freshness: %v", err)
	}

	return &fc, nil
}

// refresh performs the materialized view refresh
func (s *RefreshScheduler) refresh() {
	log.Println("🔄 Starting materialized view refresh...")
	startTime := time.Now()

	query := "SELECT success, message, refreshed_at FROM manual_refresh_account_balances()"
	
	var result RefreshResult
	err := s.db.QueryRow(query).Scan(&result.Success, &result.Message, &result.RefreshedAt)
	if err != nil {
		log.Printf("❌ Refresh failed: %v", err)
		return
	}

	duration := time.Since(startTime)

	if result.Success {
		log.Printf("✅ %s (took %v)", result.Message, duration)
	} else {
		log.Printf("❌ Refresh failed: %s", result.Message)
	}
}

// getEnv gets environment variable with fallback default
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
