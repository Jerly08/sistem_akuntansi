package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"gorm.io/gorm"
)

// BalanceReconciliationJob handles scheduled balance reconciliation
type BalanceReconciliationJob struct {
	db               *gorm.DB
	autoSyncService  *services.AutoBalanceSyncService
	config           *ReconciliationConfig
	isRunning        bool
	mutex            sync.Mutex
	stopChannel      chan bool
	statusChannel    chan ReconciliationStatus
	lastRunTime      time.Time
	lastRunResult    *ReconciliationResult
}

// ReconciliationConfig configuration for balance reconciliation job
type ReconciliationConfig struct {
	// Enabled whether the reconciliation job is enabled
	Enabled bool `json:"enabled"`
	
	// Interval how often to run reconciliation (e.g., "5m", "1h", "24h")
	Interval time.Duration `json:"interval"`
	
	// AutoFix automatically fix detected issues
	AutoFix bool `json:"auto_fix"`
	
	// MaxRetries maximum number of retry attempts for failed reconciliations
	MaxRetries int `json:"max_retries"`
	
	// RetryInterval interval between retry attempts
	RetryInterval time.Duration `json:"retry_interval"`
	
	// AlertThreshold threshold for balance differences that should trigger alerts
	AlertThreshold float64 `json:"alert_threshold"`
	
	// RunOnStartup run reconciliation immediately when the job starts
	RunOnStartup bool `json:"run_on_startup"`
	
	// QuietHours hours during which reconciliation should not run (e.g., [22, 6] means 10 PM to 6 AM)
	QuietHours []int `json:"quiet_hours"`
	
	// MaxExecutionTime maximum time allowed for a single reconciliation run
	MaxExecutionTime time.Duration `json:"max_execution_time"`
}

// ReconciliationStatus represents the current status of reconciliation job
type ReconciliationStatus struct {
	IsRunning     bool                                  `json:"is_running"`
	LastRunTime   time.Time                             `json:"last_run_time"`
	NextRunTime   time.Time                             `json:"next_run_time"`
	LastResult    *ReconciliationResult                 `json:"last_result"`
	TotalRuns     int                                   `json:"total_runs"`
	SuccessfulRuns int                                  `json:"successful_runs"`
	FailedRuns    int                                   `json:"failed_runs"`
}

// ReconciliationResult represents the result of a reconciliation run
type ReconciliationResult struct {
	StartTime               time.Time                             `json:"start_time"`
	EndTime                 time.Time                             `json:"end_time"`
	Duration                time.Duration                         `json:"duration"`
	Success                 bool                                  `json:"success"`
	Error                   string                                `json:"error,omitempty"`
	ValidationReport        *services.BalanceConsistencyReport   `json:"validation_report"`
	IssuesFound             int                                   `json:"issues_found"`
	IssuesFixed             int                                   `json:"issues_fixed"`
	AutoFixEnabled          bool                                  `json:"auto_fix_enabled"`
	AlertsTriggered         []ReconciliationAlert                 `json:"alerts_triggered"`
}

// ReconciliationAlert represents an alert triggered during reconciliation
type ReconciliationAlert struct {
	Type        string    `json:"type"`
	Message     string    `json:"message"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
	Data        interface{} `json:"data,omitempty"`
}

// NewBalanceReconciliationJob creates a new balance reconciliation job
func NewBalanceReconciliationJob(db *gorm.DB, config *ReconciliationConfig) *BalanceReconciliationJob {
	if config == nil {
		config = &ReconciliationConfig{
			Enabled:           true,
			Interval:          30 * time.Minute, // Default: every 30 minutes
			AutoFix:           false,            // Default: safe mode
			MaxRetries:        3,
			RetryInterval:     5 * time.Minute,
			AlertThreshold:    1000.0, // Alert for differences > 1000
			RunOnStartup:      true,
			QuietHours:        []int{}, // No quiet hours by default
			MaxExecutionTime:  10 * time.Minute,
		}
	}

	accountRepo := repositories.NewAccountRepository(db)
	autoSyncService := services.NewAutoBalanceSyncService(db, accountRepo)

	return &BalanceReconciliationJob{
		db:              db,
		autoSyncService: autoSyncService,
		config:          config,
		stopChannel:     make(chan bool, 1),
		statusChannel:   make(chan ReconciliationStatus, 10),
	}
}

// Start starts the balance reconciliation job
func (j *BalanceReconciliationJob) Start() error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if j.isRunning {
		return fmt.Errorf("balance reconciliation job is already running")
	}

	if !j.config.Enabled {
		log.Println("⚠️ Balance reconciliation job is disabled")
		return nil
	}

	j.isRunning = true
	log.Println("🚀 Starting balance reconciliation job...")
	log.Printf("   ⏰ Interval: %v", j.config.Interval)
	log.Printf("   🔧 Auto-fix: %t", j.config.AutoFix)
	log.Printf("   🚨 Alert threshold: %.2f", j.config.AlertThreshold)

	// Run immediately on startup if configured
	if j.config.RunOnStartup {
		go func() {
			time.Sleep(5 * time.Second) // Give the system time to start
			j.runReconciliation()
		}()
	}

	// Start the main job loop
	go j.jobLoop()

	return nil
}

// Stop stops the balance reconciliation job
func (j *BalanceReconciliationJob) Stop() error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if !j.isRunning {
		return fmt.Errorf("balance reconciliation job is not running")
	}

	log.Println("⏹️ Stopping balance reconciliation job...")
	j.stopChannel <- true
	j.isRunning = false

	return nil
}

// GetStatus returns the current status of the reconciliation job
func (j *BalanceReconciliationJob) GetStatus() ReconciliationStatus {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	nextRunTime := j.lastRunTime.Add(j.config.Interval)
	if j.lastRunTime.IsZero() {
		nextRunTime = time.Now().Add(j.config.Interval)
	}

	// Calculate run statistics (this would typically be persisted)
	totalRuns := 0
	successfulRuns := 0
	failedRuns := 0

	if j.lastRunResult != nil {
		totalRuns = 1
		if j.lastRunResult.Success {
			successfulRuns = 1
		} else {
			failedRuns = 1
		}
	}

	return ReconciliationStatus{
		IsRunning:      j.isRunning,
		LastRunTime:    j.lastRunTime,
		NextRunTime:    nextRunTime,
		LastResult:     j.lastRunResult,
		TotalRuns:      totalRuns,
		SuccessfulRuns: successfulRuns,
		FailedRuns:     failedRuns,
	}
}

// UpdateConfig updates the reconciliation job configuration
func (j *BalanceReconciliationJob) UpdateConfig(newConfig *ReconciliationConfig) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	oldConfig := *j.config
	j.config = newConfig

	log.Printf("🔄 Updated balance reconciliation job configuration:")
	log.Printf("   Enabled: %t -> %t", oldConfig.Enabled, newConfig.Enabled)
	log.Printf("   Interval: %v -> %v", oldConfig.Interval, newConfig.Interval)
	log.Printf("   Auto-fix: %t -> %t", oldConfig.AutoFix, newConfig.AutoFix)

	return nil
}

// jobLoop main loop for the reconciliation job
func (j *BalanceReconciliationJob) jobLoop() {
	ticker := time.NewTicker(j.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-j.stopChannel:
			log.Println("📤 Received stop signal, terminating balance reconciliation job")
			return

		case <-ticker.C:
			if j.shouldRunNow() {
				j.runReconciliation()
			} else {
				log.Println("⏸️ Skipping reconciliation due to quiet hours")
			}

		case status := <-j.statusChannel:
			// Handle status updates if needed
			log.Printf("📊 Reconciliation status update: Running=%t", status.IsRunning)
		}
	}
}

// shouldRunNow checks if reconciliation should run based on quiet hours
func (j *BalanceReconciliationJob) shouldRunNow() bool {
	if len(j.config.QuietHours) == 0 {
		return true
	}

	currentHour := time.Now().Hour()
	for _, quietHour := range j.config.QuietHours {
		if currentHour == quietHour {
			return false
		}
	}
	return true
}

// runReconciliation performs a single reconciliation run
func (j *BalanceReconciliationJob) runReconciliation() {
	log.Println("🔄 Starting balance reconciliation run...")
	
	result := &ReconciliationResult{
		StartTime:      time.Now(),
		AutoFixEnabled: j.config.AutoFix,
		AlertsTriggered: []ReconciliationAlert{},
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), j.config.MaxExecutionTime)
	defer cancel()

	// Perform validation
	report, err := j.performValidationWithContext(ctx)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		j.finalizeResult(result)
		return
	}

	result.ValidationReport = report
	result.IssuesFound = len(report.CashBankIssues) + len(report.ParentChildIssues)
	
	// Check for balance equation issues
	if report.BalanceEquationDifference > j.config.AlertThreshold || 
	   report.BalanceEquationDifference < -j.config.AlertThreshold {
		alert := ReconciliationAlert{
			Type:      "balance_sheet_imbalance",
			Message:   fmt.Sprintf("Balance sheet imbalance detected: %.2f", report.BalanceEquationDifference),
			Severity:  "high",
			Timestamp: time.Now(),
			Data:      report.BalanceEquationDifference,
		}
		result.AlertsTriggered = append(result.AlertsTriggered, alert)
	}

	// Check for cash bank sync issues
	for _, issue := range report.CashBankIssues {
		if issue.Difference > j.config.AlertThreshold || issue.Difference < -j.config.AlertThreshold {
			alert := ReconciliationAlert{
				Type:      "cash_bank_sync_issue",
				Message:   fmt.Sprintf("Cash bank %s has sync issue: %.2f", issue.Code, issue.Difference),
				Severity:  "medium",
				Timestamp: time.Now(),
				Data:      issue,
			}
			result.AlertsTriggered = append(result.AlertsTriggered, alert)
		}
	}

	// Auto-fix if enabled and issues found
	if j.config.AutoFix && !report.IsConsistent {
		log.Println("🔧 Auto-fixing detected balance issues...")
		
		if err := j.performAutoFixWithContext(ctx); err != nil {
			alert := ReconciliationAlert{
				Type:      "auto_fix_failed",
				Message:   fmt.Sprintf("Auto-fix failed: %v", err),
				Severity:  "high",
				Timestamp: time.Now(),
				Data:      err.Error(),
			}
			result.AlertsTriggered = append(result.AlertsTriggered, alert)
		} else {
			result.IssuesFixed = result.IssuesFound
			log.Println("✅ Auto-fix completed successfully")
		}
	}

	result.Success = true
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	j.finalizeResult(result)
}

// performValidationWithContext performs validation with context timeout
func (j *BalanceReconciliationJob) performValidationWithContext(ctx context.Context) (*services.BalanceConsistencyReport, error) {
	reportChan := make(chan *services.BalanceConsistencyReport, 1)
	errorChan := make(chan error, 1)

	go func() {
		report, err := j.autoSyncService.ValidateBalanceConsistency()
		if err != nil {
			errorChan <- err
			return
		}
		reportChan <- report
	}()

	select {
	case report := <-reportChan:
		return report, nil
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("validation timed out after %v", j.config.MaxExecutionTime)
	}
}

// performAutoFixWithContext performs auto-fix with context timeout
func (j *BalanceReconciliationJob) performAutoFixWithContext(ctx context.Context) error {
	errorChan := make(chan error, 1)

	go func() {
		err := j.autoSyncService.FixAllBalanceIssues()
		errorChan <- err
	}()

	select {
	case err := <-errorChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("auto-fix timed out after %v", j.config.MaxExecutionTime)
	}
}

// finalizeResult finalizes the reconciliation result
func (j *BalanceReconciliationJob) finalizeResult(result *ReconciliationResult) {
	j.mutex.Lock()
	j.lastRunTime = result.StartTime
	j.lastRunResult = result
	j.mutex.Unlock()

	// Log results
	if result.Success {
		log.Printf("✅ Reconciliation completed successfully in %v", result.Duration)
		if result.IssuesFound > 0 {
			log.Printf("   🔍 Issues found: %d", result.IssuesFound)
			if result.IssuesFixed > 0 {
				log.Printf("   🔧 Issues fixed: %d", result.IssuesFixed)
			}
		}
		if len(result.AlertsTriggered) > 0 {
			log.Printf("   🚨 Alerts triggered: %d", len(result.AlertsTriggered))
		}
	} else {
		log.Printf("❌ Reconciliation failed after %v: %s", result.Duration, result.Error)
	}

	// Send status update
	select {
	case j.statusChannel <- j.GetStatus():
	default:
		// Channel full, skip status update
	}

	// Log alerts
	for _, alert := range result.AlertsTriggered {
		log.Printf("🚨 ALERT [%s] %s: %s", alert.Severity, alert.Type, alert.Message)
	}
}

// ManualRun triggers a manual reconciliation run
func (j *BalanceReconciliationJob) ManualRun() (*ReconciliationResult, error) {
	if !j.isRunning {
		return nil, fmt.Errorf("reconciliation job is not running")
	}

	log.Println("🔄 Manual reconciliation run requested")
	
	// Run in a separate goroutine and wait for completion
	resultChan := make(chan *ReconciliationResult, 1)
	
	go func() {
		j.runReconciliation()
		resultChan <- j.lastRunResult
	}()

	// Wait for completion or timeout
	select {
	case result := <-resultChan:
		return result, nil
	case <-time.After(j.config.MaxExecutionTime):
		return nil, fmt.Errorf("manual reconciliation timed out")
	}
}