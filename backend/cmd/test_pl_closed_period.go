package main

import (
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/services"
	"fmt"
	"log"
	"strings"
)

func main() {
	fmt.Println("=== P&L Closed Period Historical Data Test ===\n")

	// Initialize database
	db := database.DB
	if db == nil {
		log.Fatalf("Database not initialized")
	}

	// Create P&L service
	plService := services.NewSSOTProfitLossService(db)

	// Test Case 1: First Closed Period (2025-01-01 to 2025-12-01)
	fmt.Println("📊 TEST 1: First Closed Period (2025-01-01 to 2025-12-01)")
	fmt.Println(strings.Repeat("-", 80))
	
	plData1, err := plService.GenerateSSOTProfitLoss("2025-01-01", "2025-12-01")
	if err != nil {
		log.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Data Source: %s\n", plData1.DataSource)
		fmt.Printf("📈 Total Revenue: Rp %.2f\n", plData1.Revenue.TotalRevenue)
		fmt.Printf("📉 Total COGS: Rp %.2f\n", plData1.COGS.TotalCOGS)
		fmt.Printf("💰 Gross Profit: Rp %.2f\n", plData1.GrossProfit)
		fmt.Printf("📊 Total Operating Expenses: Rp %.2f\n", plData1.OperatingExpenses.TotalOpEx)
		fmt.Printf("🎯 Operating Income: Rp %.2f\n", plData1.OperatingIncome)
		fmt.Printf("💵 Net Income: Rp %.2f\n", plData1.NetIncome)
		
		// Show revenue items
		if len(plData1.Revenue.Items) > 0 {
			fmt.Println("\n📋 Revenue Items:")
			for _, item := range plData1.Revenue.Items {
				fmt.Printf("  - %s (%s): Rp %.2f\n", item.AccountName, item.AccountCode, item.Amount)
			}
		}
		
		// Show COGS items
		if len(plData1.COGS.Items) > 0 {
			fmt.Println("\n📋 COGS Items:")
			for _, item := range plData1.COGS.Items {
				fmt.Printf("  - %s (%s): Rp %.2f\n", item.AccountName, item.AccountCode, item.Amount)
			}
		}
	}

	// Test Case 2: Second Closed Period (2025-01-01 to 2026-12-31)
	fmt.Println("\n\n📊 TEST 2: Second Closed Period (2025-01-01 to 2026-12-31)")
	fmt.Println(strings.Repeat("-", 80))
	
	plData2, err := plService.GenerateSSOTProfitLoss("2025-01-01", "2026-12-31")
	if err != nil {
		log.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Data Source: %s\n", plData2.DataSource)
		fmt.Printf("📈 Total Revenue: Rp %.2f\n", plData2.Revenue.TotalRevenue)
		fmt.Printf("📉 Total COGS: Rp %.2f\n", plData2.COGS.TotalCOGS)
		fmt.Printf("💰 Gross Profit: Rp %.2f\n", plData2.GrossProfit)
		fmt.Printf("📊 Total Operating Expenses: Rp %.2f\n", plData2.OperatingExpenses.TotalOpEx)
		fmt.Printf("🎯 Operating Income: Rp %.2f\n", plData2.OperatingIncome)
		fmt.Printf("💵 Net Income: Rp %.2f\n", plData2.NetIncome)
	}

	// Test Case 3: Third Closed Period (2025-01-01 to 2027-12-31)
	fmt.Println("\n\n📊 TEST 3: Third Closed Period (2025-01-01 to 2027-12-31)")
	fmt.Println(strings.Repeat("-", 80))
	
	plData3, err := plService.GenerateSSOTProfitLoss("2025-01-01", "2027-12-31")
	if err != nil {
		log.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Data Source: %s\n", plData3.DataSource)
		fmt.Printf("📈 Total Revenue: Rp %.2f\n", plData3.Revenue.TotalRevenue)
		fmt.Printf("📉 Total COGS: Rp %.2f\n", plData3.COGS.TotalCOGS)
		fmt.Printf("💰 Gross Profit: Rp %.2f\n", plData3.GrossProfit)
		fmt.Printf("📊 Total Operating Expenses: Rp %.2f\n", plData3.OperatingExpenses.TotalOpEx)
		fmt.Printf("🎯 Operating Income: Rp %.2f\n", plData3.OperatingIncome)
		fmt.Printf("💵 Net Income: Rp %.2f\n", plData3.NetIncome)
	}

	// Summary
	fmt.Println("\n\n" + strings.Repeat("=", 80))
	fmt.Println("📊 SUMMARY")
	fmt.Println(strings.Repeat("=", 80))
	
	fmt.Println("\n✅ Expected Results (BEFORE closing):")
	fmt.Println("   Period 1 (2025-12-01): Revenue = 7,000,000 | COGS = 3,500,000 | Net = 3,500,000")
	fmt.Println("   Period 2 (2026-12-31): Revenue = 14,000,000 | COGS = 7,000,000 | Net = 7,000,000")
	fmt.Println("   Period 3 (2027-12-31): Revenue = 21,000,000 | COGS = 10,500,000 | Net = 10,500,000")
	
	fmt.Println("\n📝 Actual Results:")
	fmt.Printf("   Period 1: Revenue = %.0f | COGS = %.0f | Net = %.0f\n", 
		plData1.Revenue.TotalRevenue, plData1.COGS.TotalCOGS, plData1.NetIncome)
	fmt.Printf("   Period 2: Revenue = %.0f | COGS = %.0f | Net = %.0f\n", 
		plData2.Revenue.TotalRevenue, plData2.COGS.TotalCOGS, plData2.NetIncome)
	fmt.Printf("   Period 3: Revenue = %.0f | COGS = %.0f | Net = %.0f\n", 
		plData3.Revenue.TotalRevenue, plData3.COGS.TotalCOGS, plData3.NetIncome)
	
	// Validation
	fmt.Println("\n🔍 Validation:")
	
	isValid := true
	
	if plData1.Revenue.TotalRevenue != 7000000 {
		fmt.Println("   ❌ Period 1 Revenue mismatch")
		isValid = false
	} else {
		fmt.Println("   ✅ Period 1 Revenue correct")
	}
	
	if plData2.Revenue.TotalRevenue != 14000000 {
		fmt.Println("   ❌ Period 2 Revenue mismatch")
		isValid = false
	} else {
		fmt.Println("   ✅ Period 2 Revenue correct")
	}
	
	if plData3.Revenue.TotalRevenue != 21000000 {
		fmt.Println("   ❌ Period 3 Revenue mismatch")
		isValid = false
	} else {
		fmt.Println("   ✅ Period 3 Revenue correct")
	}
	
	fmt.Println("\n" + strings.Repeat("=", 80))
	if isValid {
		fmt.Println("✅ ALL TESTS PASSED! P&L correctly shows historical data for closed periods.")
	} else {
		fmt.Println("❌ SOME TESTS FAILED! P&L data does not match expected historical values.")
	}
	fmt.Println(strings.Repeat("=", 80))
}
