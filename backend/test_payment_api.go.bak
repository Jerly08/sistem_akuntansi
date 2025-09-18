package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type PaymentRequest struct {
	Amount     float64   `json:"amount"`
	Date       time.Time `json:"date"`
	Method     string    `json:"method"`
	CashBankID uint      `json:"cash_bank_id"`
	Reference  string    `json:"reference"`
	Notes      string    `json:"notes"`
}

func main() {
	fmt.Println("🧪 Testing Payment API Response...")

	// Prepare payment request (same as frontend)
	paymentReq := PaymentRequest{
		Amount:     5550000,
		Date:       time.Date(2025, 9, 8, 0, 0, 0, 0, time.UTC),
		Method:     "BANK_TRANSFER",
		CashBankID: 8,
		Reference:  "test",
		Notes:      "test",
	}

	// Convert to JSON
	jsonData, err := json.Marshal(paymentReq)
	if err != nil {
		log.Fatalf("Error marshaling request: %v", err)
	}

	fmt.Printf("📤 Request payload:\n%s\n\n", string(jsonData))

	// Make HTTP request to the API
	url := "http://localhost:8080/api/v1/sales/16/integrated-payment"
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Add headers (you might need to add authentication headers)
	req.Header.Set("Content-Type", "application/json")
	// Add any authentication headers if needed
	// req.Header.Set("Authorization", "Bearer YOUR_TOKEN")

	// Make the request
	fmt.Println("📡 Making API request...")
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	fmt.Printf("📥 Response Status: %s (%d)\n", resp.Status, resp.StatusCode)
	fmt.Printf("📥 Response Headers:\n")
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	fmt.Printf("\n📥 Response Body:\n%s\n\n", string(body))

	// Try to parse JSON response
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("❌ Could not parse JSON response: %v\n", err)
		fmt.Println("Raw response:")
		fmt.Println(string(body))
		return
	}

	// Analyze response structure
	fmt.Println("📊 Response Analysis:")
	
	if success, ok := response["success"].(bool); ok {
		if success {
			fmt.Println("✅ Success field: true")
		} else {
			fmt.Println("❌ Success field: false")
		}
	} else {
		fmt.Println("⚠️  No 'success' field found in response")
	}

	if message, ok := response["message"].(string); ok {
		fmt.Printf("💬 Message: %s\n", message)
	}

	if status, ok := response["status"].(string); ok {
		fmt.Printf("📊 Status: %s\n", status)
	}

	if errorMsg, ok := response["error"].(string); ok {
		fmt.Printf("❌ Error: %s\n", errorMsg)
		if details, ok := response["details"].(string); ok {
			fmt.Printf("📝 Details: %s\n", details)
		}
	}

	if payment, ok := response["payment"].(map[string]interface{}); ok {
		fmt.Println("💰 Payment created:")
		if id, ok := payment["id"].(float64); ok {
			fmt.Printf("  ID: %.0f\n", id)
		}
		if code, ok := payment["code"].(string); ok {
			fmt.Printf("  Code: %s\n", code)
		}
		if amount, ok := payment["amount"].(float64); ok {
			fmt.Printf("  Amount: %.2f\n", amount)
		}
	}

	// Determine if this should be considered success or failure
	fmt.Println("\n🎯 Final Assessment:")
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if success, ok := response["success"].(bool); ok && success {
			fmt.Println("✅ This should be treated as SUCCESS by frontend")
		} else if _, hasError := response["error"]; hasError {
			fmt.Println("❌ This should be treated as ERROR by frontend")
		} else {
			fmt.Println("⚠️  Ambiguous response - frontend may be confused")
		}
	} else {
		fmt.Println("❌ HTTP error status - this should be treated as ERROR")
	}

	fmt.Println("\n💡 Recommendations:")
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && response["payment"] != nil {
		fmt.Println("- Payment was created successfully")
		fmt.Println("- Frontend should show success message")
		fmt.Println("- If frontend shows error, check frontend error handling logic")
	} else {
		fmt.Println("- Check server logs for actual error")
		fmt.Println("- Verify request payload and authentication")
	}
}
