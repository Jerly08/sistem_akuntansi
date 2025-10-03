package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	log.Println("🧪 Starting endpoint verification...")

	baseURL := "http://localhost:8080"
	
	// Test endpoints without auth (should return auth error, not 404)
	endpointsToTest := []string{
		"/api/v1/health",
		"/api/v1/auth/login",
		"/api/v1/accounts",
		"/api/v1/contacts", 
		"/api/v1/cash-bank/accounts",
		"/api/v1/receipts",
		"/api/v1/sales",
		"/api/v1/purchases",
		"/docs/index.html",
		"/docs/swagger.json",
	}

	log.Println("🔍 Testing endpoints...")
	
	for _, endpoint := range endpointsToTest {
		testEndpoint(baseURL + endpoint)
		time.Sleep(100 * time.Millisecond) // Small delay between requests
	}

	log.Println("✅ Endpoint verification completed!")
}

func testEndpoint(url string) {
	client := &http.Client{Timeout: 10 * time.Second}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("❌ %s: Failed to create request - %v", url, err)
		return
	}

	// Add headers for API endpoints
	if strings.Contains(url, "/api/v1") && !strings.Contains(url, "/health") && !strings.Contains(url, "/auth/login") {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		// Note: Not adding Authorization header to test auth requirement
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ %s: Request failed - %v", url, err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)

	switch resp.StatusCode {
	case 200:
		log.Printf("✅ %s: OK (200) - Endpoint working", url)
	case 401:
		if strings.Contains(bodyStr, "Authorization") || strings.Contains(bodyStr, "AUTH_") {
			log.Printf("🔒 %s: Requires auth (401) - Endpoint exists and protected", url)
		} else {
			log.Printf("⚠️  %s: Unauthorized (401) - %s", url, bodyStr)
		}
	case 400:
		log.Printf("📝 %s: Bad Request (400) - Endpoint exists, needs valid data", url)
	case 404:
		if strings.Contains(bodyStr, "page not found") {
			log.Printf("❌ %s: NOT FOUND (404) - Endpoint missing!", url)
		} else {
			log.Printf("❓ %s: Not Found (404) - %s", url, bodyStr)
		}
	case 500:
		log.Printf("💥 %s: Server Error (500) - %s", url, bodyStr[:min(100, len(bodyStr))])
	default:
		log.Printf("❓ %s: Status %d - %s", url, resp.StatusCode, bodyStr[:min(100, len(bodyStr))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}