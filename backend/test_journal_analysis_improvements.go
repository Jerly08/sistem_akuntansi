package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"github.com/shopspring/decimal"
)

func main() {
	// Load configuration
	config.Load()
	
	// Connect to database
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("=== Testing Journal Analysis Improvements ===")

	// Initialize services
	enhancedReportService := services.NewEnhancedReportService(db)
	unifiedJournalService := services.NewUnifiedJournalService(db)
	ssotReportService := services.NewSSOTReportIntegrationService(db, unifiedJournalService, enhancedReportService)

	ctx := context.Background()

	// Test date range - last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	fmt.Printf("Testing journal analysis for period: %s to %s\n", 
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Test 1: Check database content first
	fmt.Println("\n1. Checking SSOT Journal Entries in Database...")
	var totalEntries int64
	var postedEntries int64
	var sourceTypes []string

	// Count total entries
	db.Model(&models.SSOTJournalEntry{}).
		Where("entry_date >= ? AND entry_date <= ?", startDate, endDate).
		Count(&totalEntries)

	// Count posted entries  
	db.Model(&models.SSOTJournalEntry{}).
		Where("entry_date >= ? AND entry_date <= ? AND status = ?", startDate, endDate, models.SSOTStatusPosted).
		Count(&postedEntries)

	// Get unique source types
	db.Model(&models.SSOTJournalEntry{}).
		Where("entry_date >= ? AND entry_date <= ? AND status = ?", startDate, endDate, models.SSOTStatusPosted).
		Distinct("source_type").
		Pluck("source_type", &sourceTypes)

	fmt.Printf("  Total entries in period: %d\n", totalEntries)
	fmt.Printf("  Posted entries: %d\n", postedEntries)
	fmt.Printf("  Source types found: %v\n", sourceTypes)

	if totalEntries == 0 {
		fmt.Println("  ⚠️  No journal entries found in the specified period!")
		fmt.Println("     This could explain why the report shows no data.")
		
		// Check if there are any entries at all
		var anyEntries int64
		db.Model(&models.SSOTJournalEntry{}).Count(&anyEntries)
		fmt.Printf("     Total entries in database: %d\n", anyEntries)
		
		if anyEntries == 0 {
			fmt.Println("     ❌ No SSOT journal entries exist in the database!")
			return
		}
		
		// Check date range of existing entries
		var minDate, maxDate time.Time
		db.Model(&models.SSOTJournalEntry{}).
			Select("MIN(entry_date) as min_date, MAX(entry_date) as max_date").
			Scan(&struct{
				MinDate time.Time `json:"min_date"`
				MaxDate time.Time `json:"max_date"`
			}{MinDate: minDate, MaxDate: maxDate})
			
		fmt.Printf("     Date range of existing entries: %s to %s\n", 
			minDate.Format("2006-01-02"), maxDate.Format("2006-01-02"))
		
		// Use the actual date range for testing
		startDate = minDate
		endDate = maxDate
		fmt.Printf("     Adjusting test period to: %s to %s\n",
			startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	}

	// Test 2: Generate journal analysis
	fmt.Println("\n2. Generating Enhanced Journal Analysis...")
	result, err := ssotReportService.GenerateJournalAnalysisFromSSot(startDate, endDate)
	if err != nil {
		log.Fatalf("Failed to generate journal analysis: %v", err)
	}

	// Test 3: Display results
	fmt.Println("\n3. Analysis Results:")
	fmt.Printf("  Company: %s\n", result.Company.Name)
	fmt.Printf("  Currency: %s\n", result.Currency)
	fmt.Printf("  Total Entries: %d\n", result.TotalEntries)
	fmt.Printf("  Posted Entries: %d\n", result.PostedEntries)
	fmt.Printf("  Draft Entries: %d\n", result.DraftEntries)
	fmt.Printf("  Reversed Entries: %d\n", result.ReversedEntries)
	fmt.Printf("  Total Amount: %s\n", formatIDR(result.TotalAmount))

	fmt.Println("\n4. Entry Type Breakdown:")
	if len(result.EntriesByType) == 0 {
		fmt.Println("  ❌ No entry type breakdown found!")
	} else {
		for i, entry := range result.EntriesByType {
			fmt.Printf("  %d. %s\n", i+1, entry.SourceType)
			fmt.Printf("     Count: %d entries\n", entry.Count)
			fmt.Printf("     Amount: %s\n", formatIDR(entry.TotalAmount))
			fmt.Printf("     Percentage: %.1f%%\n", entry.Percentage)
		}
	}

	fmt.Println("\n5. Account Breakdown:")
	if len(result.EntriesByAccount) == 0 {
		fmt.Println("  ❌ No account breakdown found!")
	} else {
		fmt.Printf("  Found %d accounts:\n", len(result.EntriesByAccount))
		for i, account := range result.EntriesByAccount {
			if i >= 5 { // Show only first 5
				fmt.Printf("  ... and %d more accounts\n", len(result.EntriesByAccount)-5)
				break
			}
			fmt.Printf("  %d. %s - %s\n", i+1, account.AccountCode, account.AccountName)
			fmt.Printf("     Entries: %d\n", account.Count)
			fmt.Printf("     Total Debit: %s\n", formatIDR(account.TotalDebit))
			fmt.Printf("     Total Credit: %s\n", formatIDR(account.TotalCredit))
		}
	}

	fmt.Println("\n6. Period Breakdown:")
	if len(result.EntriesByPeriod) == 0 {
		fmt.Println("  ❌ No period breakdown found!")
	} else {
		fmt.Printf("  Found %d periods:\n", len(result.EntriesByPeriod))
		for i, period := range result.EntriesByPeriod {
			if i >= 10 { // Show only first 10
				fmt.Printf("  ... and %d more periods\n", len(result.EntriesByPeriod)-10)
				break
			}
			fmt.Printf("  %d. %s\n", i+1, period.Period)
			fmt.Printf("     Entries: %d\n", period.Count)
			fmt.Printf("     Amount: %s\n", formatIDR(period.TotalAmount))
		}
	}

	fmt.Println("\n7. Compliance Check:")
	fmt.Printf("  Compliance Score: %.1f%%\n", result.ComplianceCheck.ComplianceScore)
	fmt.Printf("  Total Checks: %d\n", result.ComplianceCheck.TotalChecks)
	fmt.Printf("  Passed Checks: %d\n", result.ComplianceCheck.PassedChecks)
	fmt.Printf("  Failed Checks: %d\n", result.ComplianceCheck.FailedChecks)
	if len(result.ComplianceCheck.Issues) > 0 {
		fmt.Println("  Issues Found:")
		for _, issue := range result.ComplianceCheck.Issues {
			fmt.Printf("    - %s: %s [%s]\n", issue.Type, issue.Description, issue.Severity)
		}
	}

	fmt.Println("\n8. Data Quality Metrics:")
	fmt.Printf("  Overall Score: %.1f%%\n", result.DataQualityMetrics.OverallScore)
	fmt.Printf("  Completeness: %.1f%%\n", result.DataQualityMetrics.CompletenessScore)
	fmt.Printf("  Accuracy: %.1f%%\n", result.DataQualityMetrics.AccuracyScore)
	fmt.Printf("  Consistency: %.1f%%\n", result.DataQualityMetrics.ConsistencyScore)
	if len(result.DataQualityMetrics.Issues) > 0 {
		fmt.Println("  Quality Issues:")
		for _, issue := range result.DataQualityMetrics.Issues {
			fmt.Printf("    - %s: %s [%s] (%d entries affected)\n", 
				issue.Type, issue.Description, issue.Severity, issue.Count)
		}
	}

	fmt.Println("\n=== Test Summary ===")
	if result.TotalEntries == 0 {
		fmt.Println("❌ No journal entries found - this explains the empty report")
		fmt.Println("   Action needed: Create some journal entries or check the SSOT migration")
	} else {
		fmt.Printf("✅ Found %d journal entries\n", result.TotalEntries)
		
		if len(result.EntriesByType) == 0 {
			fmt.Println("❌ Entry type breakdown is empty")
		} else {
			fmt.Printf("✅ Entry type breakdown shows %d types\n", len(result.EntriesByType))
		}
		
		if len(result.EntriesByAccount) == 0 {
			fmt.Println("❌ Account breakdown is empty")
		} else {
			fmt.Printf("✅ Account breakdown shows %d accounts\n", len(result.EntriesByAccount))
		}
		
		if len(result.EntriesByPeriod) == 0 {
			fmt.Println("❌ Period breakdown is empty")
		} else {
			fmt.Printf("✅ Period breakdown shows %d periods\n", len(result.EntriesByPeriod))
		}
		
		fmt.Printf("✅ Compliance score: %.1f%%\n", result.ComplianceCheck.ComplianceScore)
		fmt.Printf("✅ Data quality score: %.1f%%\n", result.DataQualityMetrics.OverallScore)
	}

	fmt.Println("\n🎉 Journal Analysis Improvements Test Complete!")
}

func formatIDR(amount decimal.Decimal) string {
	// Simple IDR formatting
	f, _ := amount.Float64()
	return fmt.Sprintf("Rp %.0f", f)
}