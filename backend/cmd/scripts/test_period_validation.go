package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL = "http://localhost:8080/api/v1"
	// Replace with your actual JWT token
	authToken = "YOUR_JWT_TOKEN_HERE"
)

type TestResult struct {
	TestName     string
	Endpoint     string
	Expected     string
	Status       int
	Success      bool
	ErrorDetails map[string]interface{}
	Message      string
}

var testResults []TestResult

func main() {
	fmt.Println("🧪 Period Validation Testing Suite")
	fmt.Println("===================================\n")

	// Get auth token first
	token := getAuthToken()
	if token == "" {
		fmt.Println("❌ Failed to authenticate. Please update credentials.")
		return
	}

	fmt.Printf("✅ Authenticated successfully\n\n")

	// Run all test scenarios
	runAllTests(token)

	// Print summary
	printSummary()
}

func getAuthToken() string {
	// Try to login with different possible admin accounts
	credentialsList := []map[string]string{
		{"email": "admin@company.com", "password": "admin123"},
		{"email": "admin@admin.com", "password": "admin123"},
		{"email": "admin@example.com", "password": "admin123"},
	}

	for _, creds := range credentialsList {
		fmt.Printf("Trying login with: %s...\n", creds["email"])
		
		loginData := map[string]interface{}{
			"email":    creds["email"],
			"password": creds["password"],
		}

		body, _ := json.Marshal(loginData)
		resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Printf("  ❌ Connection error: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		
		if resp.StatusCode == 200 {
			var result map[string]interface{}
			json.Unmarshal(bodyBytes, &result)

			if token, ok := result["token"].(string); ok && token != "" {
				fmt.Printf("  ✅ Login successful!\n")
				return token
			}
		} else {
			fmt.Printf("  ❌ Status %d: %s\n", resp.StatusCode, string(bodyBytes))
		}
	}

	fmt.Println("\n⚠️  All login attempts failed.")
	fmt.Println("Please create an admin user first or update credentials in the script.")
	fmt.Println("\nTo create admin user, run:")
	fmt.Println("  POST /api/v1/auth/register")
	fmt.Println("  {\"email\": \"admin@admin.com\", \"password\": \"admin123\", \"name\": \"Admin\", \"role\": \"admin\"}")
	
	return ""
}

func runAllTests(token string) {
	fmt.Println("📋 Running Test Scenarios...")
	fmt.Println("----------------------------\n")

	// Test 1: Create transaction to OPEN period (should succeed)
	testOpenPeriod(token)

	// Test 2: Close a period
	closePeriod(token, 2025, 1)

	// Test 3: Create transaction to CLOSED period (should fail)
	testClosedPeriod(token)

	// Test 4: Reopen the period
	reopenPeriod(token, 2025, 1)

	// Test 5: Create transaction to REOPENED period (should succeed)
	testReopenedPeriod(token)

	// Test 6: Try old date (> 2 years, should fail)
	testOldDate(token)

	// Test 7: Try future date (> 7 days, should fail)
	testFutureDate(token)

	// Test 8: Auto-create period (should succeed)
	testAutoCreatePeriod(token)

	// Test 9: Test different transaction types
	testDifferentTransactionTypes(token)
}

func testOpenPeriod(token string) {
	fmt.Println("🟢 Test 1: Create Sale to OPEN Period")

	// Use current month
	now := time.Now()
	saleData := map[string]interface{}{
		"customer_id": 1,
		"type":        "INVOICE",
		"date":        now.Format("2006-01-02"),
		"due_date":    now.AddDate(0, 0, 30).Format("2006-01-02"),
		"items": []map[string]interface{}{
			{
				"product_id": 1,
				"quantity":   5,
				"unit_price": 100000,
			},
		},
	}

	result := makeRequest(token, "POST", "/sales", saleData)
	testResults = append(testResults, TestResult{
		TestName: "Open Period - Create Sale",
		Endpoint: "POST /sales",
		Expected: "200 OK - Transaction allowed",
		Status:   result.Status,
		Success:  result.Status == 200 || result.Status == 201,
		Message:  result.Message,
	})

	if result.Status == 200 || result.Status == 201 {
		fmt.Printf("   ✅ SUCCESS: Sale created to open period\n")
		fmt.Printf("   Response: %s\n\n", result.Message)
	} else {
		fmt.Printf("   ❌ FAILED: %s\n\n", result.Message)
	}
}

func closePeriod(token string, year, month int) {
	fmt.Printf("🔒 Closing Period: %d-%02d\n", year, month)

	result := makeRequest(token, "POST", fmt.Sprintf("/periods/%d/%d/close", year, month), nil)
	
	if result.Status == 200 {
		fmt.Printf("   ✅ Period %d-%02d closed successfully\n\n", year, month)
	} else {
		fmt.Printf("   ⚠️  Period close status: %d\n\n", result.Status)
	}
}

func testClosedPeriod(token string) {
	fmt.Println("🔴 Test 2: Create Sale to CLOSED Period")

	// Try to create sale to January 2025 (now closed)
	saleData := map[string]interface{}{
		"customer_id": 1,
		"type":        "INVOICE",
		"date":        "2025-01-15",
		"due_date":    "2025-02-15",
		"items": []map[string]interface{}{
			{
				"product_id": 1,
				"quantity":   5,
				"unit_price": 100000,
			},
		},
	}

	result := makeRequest(token, "POST", "/sales", saleData)
	testResults = append(testResults, TestResult{
		TestName:     "Closed Period - Create Sale",
		Endpoint:     "POST /sales",
		Expected:     "403 Forbidden - Period closed",
		Status:       result.Status,
		Success:      result.Status == 403,
		ErrorDetails: result.ErrorDetails,
		Message:      result.Message,
	})

	if result.Status == 403 {
		fmt.Printf("   ✅ CORRECT: Transaction blocked\n")
		fmt.Printf("   Error Code: %v\n", result.ErrorDetails["code"])
		fmt.Printf("   Details: %v\n", result.ErrorDetails["details"])
		fmt.Printf("   Period: %v\n\n", result.ErrorDetails["period"])
	} else {
		fmt.Printf("   ❌ UNEXPECTED: Status %d (expected 403)\n", result.Status)
		fmt.Printf("   Response: %s\n\n", result.Message)
	}
}

func reopenPeriod(token string, year, month int) {
	fmt.Printf("🔓 Reopening Period: %d-%02d\n", year, month)

	reopenData := map[string]interface{}{
		"reason": "Testing period validation - need to add correction entry",
	}

	result := makeRequest(token, "POST", fmt.Sprintf("/periods/%d/%d/reopen", year, month), reopenData)
	
	if result.Status == 200 {
		fmt.Printf("   ✅ Period %d-%02d reopened successfully\n\n", year, month)
	} else {
		fmt.Printf("   ⚠️  Period reopen status: %d - %s\n\n", result.Status, result.Message)
	}
}

func testReopenedPeriod(token string) {
	fmt.Println("🟢 Test 3: Create Sale to REOPENED Period")

	saleData := map[string]interface{}{
		"customer_id": 1,
		"type":        "INVOICE",
		"date":        "2025-01-20",
		"due_date":    "2025-02-20",
		"items": []map[string]interface{}{
			{
				"product_id": 1,
				"quantity":   3,
				"unit_price": 150000,
			},
		},
	}

	result := makeRequest(token, "POST", "/sales", saleData)
	testResults = append(testResults, TestResult{
		TestName: "Reopened Period - Create Sale",
		Endpoint: "POST /sales",
		Expected: "200 OK - Transaction allowed after reopen",
		Status:   result.Status,
		Success:  result.Status == 200 || result.Status == 201,
		Message:  result.Message,
	})

	if result.Status == 200 || result.Status == 201 {
		fmt.Printf("   ✅ SUCCESS: Sale created to reopened period\n\n")
	} else {
		fmt.Printf("   ❌ FAILED: %s\n\n", result.Message)
	}
}

func testOldDate(token string) {
	fmt.Println("🔴 Test 4: Create Sale with OLD Date (> 2 years)")

	// Date more than 2 years ago
	oldDate := time.Now().AddDate(-3, 0, 0).Format("2006-01-02")
	
	saleData := map[string]interface{}{
		"customer_id": 1,
		"type":        "INVOICE",
		"date":        oldDate,
		"due_date":    oldDate,
		"items": []map[string]interface{}{
			{
				"product_id": 1,
				"quantity":   2,
				"unit_price": 50000,
			},
		},
	}

	result := makeRequest(token, "POST", "/sales", saleData)
	testResults = append(testResults, TestResult{
		TestName:     "Old Date - Create Sale",
		Endpoint:     "POST /sales",
		Expected:     "403 Forbidden - Date too old",
		Status:       result.Status,
		Success:      result.Status == 403,
		ErrorDetails: result.ErrorDetails,
		Message:      result.Message,
	})

	if result.Status == 403 {
		fmt.Printf("   ✅ CORRECT: Old date rejected\n")
		fmt.Printf("   Details: %v\n\n", result.ErrorDetails["details"])
	} else {
		fmt.Printf("   ❌ UNEXPECTED: Status %d (expected 403)\n\n", result.Status)
	}
}

func testFutureDate(token string) {
	fmt.Println("🔴 Test 5: Create Sale with FUTURE Date (> 7 days)")

	// Date more than 7 days in future
	futureDate := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	
	saleData := map[string]interface{}{
		"customer_id": 1,
		"type":        "INVOICE",
		"date":        futureDate,
		"due_date":    futureDate,
		"items": []map[string]interface{}{
			{
				"product_id": 1,
				"quantity":   2,
				"unit_price": 50000,
			},
		},
	}

	result := makeRequest(token, "POST", "/sales", saleData)
	testResults = append(testResults, TestResult{
		TestName:     "Future Date - Create Sale",
		Endpoint:     "POST /sales",
		Expected:     "403 Forbidden - Date too far in future",
		Status:       result.Status,
		Success:      result.Status == 403,
		ErrorDetails: result.ErrorDetails,
		Message:      result.Message,
	})

	if result.Status == 403 {
		fmt.Printf("   ✅ CORRECT: Future date rejected\n")
		fmt.Printf("   Details: %v\n\n", result.ErrorDetails["details"])
	} else {
		fmt.Printf("   ❌ UNEXPECTED: Status %d (expected 403)\n\n", result.Status)
	}
}

func testAutoCreatePeriod(token string) {
	fmt.Println("🟢 Test 6: Auto-Create Period (within ±2 years)")

	// Use a date 6 months in future (within range)
	futureDate := time.Now().AddDate(0, 6, 0).Format("2006-01-02")
	
	saleData := map[string]interface{}{
		"customer_id": 1,
		"type":        "INVOICE",
		"date":        futureDate,
		"due_date":    futureDate,
		"items": []map[string]interface{}{
			{
				"product_id": 1,
				"quantity":   1,
				"unit_price": 25000,
			},
		},
	}

	result := makeRequest(token, "POST", "/sales", saleData)
	testResults = append(testResults, TestResult{
		TestName: "Auto-Create Period - Create Sale",
		Endpoint: "POST /sales",
		Expected: "200 OK - Period auto-created",
		Status:   result.Status,
		Success:  result.Status == 200 || result.Status == 201,
		Message:  result.Message,
	})

	if result.Status == 200 || result.Status == 201 {
		fmt.Printf("   ✅ SUCCESS: Period auto-created and sale created\n")
		fmt.Printf("   Date: %s\n\n", futureDate)
	} else {
		fmt.Printf("   ❌ FAILED: %s\n\n", result.Message)
	}
}

func testDifferentTransactionTypes(token string) {
	fmt.Println("🔵 Test 7: Different Transaction Types")
	
	now := time.Now()
	currentDate := now.Format("2006-01-02")

	// Test Purchase
	fmt.Println("   Testing Purchase...")
	purchaseData := map[string]interface{}{
		"vendor_id":      1,
		"date":           currentDate,
		"due_date":       now.AddDate(0, 0, 30).Format("2006-01-02"),
		"payment_method": "CREDIT",
		"items": []map[string]interface{}{
			{
				"product_id": 1,
				"quantity":   10,
				"unit_price": 75000,
			},
		},
	}
	
	result := makeRequest(token, "POST", "/purchases", purchaseData)
	fmt.Printf("   Purchase: Status %d\n", result.Status)

	// Test Journal Entry
	fmt.Println("   Testing Journal Entry...")
	journalData := map[string]interface{}{
		"entry_date":  currentDate,
		"description": "Test period validation",
		"lines": []map[string]interface{}{
			{
				"account_code": "1111",
				"debit":        100000,
				"credit":       0,
			},
			{
				"account_code": "4111",
				"debit":        0,
				"credit":       100000,
			},
		},
	}
	
	result = makeRequest(token, "POST", "/journals", journalData)
	fmt.Printf("   Journal: Status %d\n", result.Status)

	// Test Stock Adjustment
	fmt.Println("   Testing Stock Adjustment...")
	stockData := map[string]interface{}{
		"product_id": 1,
		"quantity":   5,
		"type":       "IN",
		"notes":      "Period validation test",
	}
	
	result = makeRequest(token, "POST", "/products/adjust-stock", stockData)
	fmt.Printf("   Stock Adjustment: Status %d\n\n", result.Status)
}

type RequestResult struct {
	Status       int
	Message      string
	ErrorDetails map[string]interface{}
}

func makeRequest(token, method, endpoint string, data interface{}) RequestResult {
	var body io.Reader
	if data != nil {
		jsonData, _ := json.Marshal(data)
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+endpoint, body)
	if err != nil {
		return RequestResult{Status: 0, Message: err.Error()}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return RequestResult{Status: 0, Message: err.Error()}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	
	var result map[string]interface{}
	json.Unmarshal(bodyBytes, &result)

	return RequestResult{
		Status:       resp.StatusCode,
		Message:      string(bodyBytes),
		ErrorDetails: result,
	}
}

func printSummary() {
	fmt.Println("\n📊 Test Summary")
	fmt.Println("===============")
	
	passed := 0
	failed := 0
	
	for _, result := range testResults {
		status := "❌ FAILED"
		if result.Success {
			status = "✅ PASSED"
			passed++
		} else {
			failed++
		}
		
		fmt.Printf("%s - %s (Status: %d)\n", status, result.TestName, result.Status)
		fmt.Printf("   Expected: %s\n", result.Expected)
		if result.ErrorDetails != nil && result.ErrorDetails["code"] != nil {
			fmt.Printf("   Error Code: %v\n", result.ErrorDetails["code"])
		}
	}
	
	fmt.Printf("\n📈 Results: %d passed, %d failed (Total: %d)\n", passed, failed, len(testResults))
	
	if failed == 0 {
		fmt.Println("🎉 All tests passed!")
	} else {
		fmt.Println("⚠️  Some tests failed. Please review.")
	}
}
