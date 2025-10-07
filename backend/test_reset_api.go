package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID    uint   `json:"id"`
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"user"`
	} `json:"data"`
}

type ResetCounterRequest struct {
	Year    int `json:"year"`
	Counter int `json:"counter"`
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run test_reset_api.go <invoice_type_id> <year> <new_counter_value>")
		fmt.Println("Example: go run test_reset_api.go 1 2025 200")
		os.Exit(1)
	}

	invoiceTypeID := os.Args[1]
	year, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("Invalid year: %v", err)
	}

	counter, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatalf("Invalid counter: %v", err)
	}

	baseURL := "http://localhost:8080/api/v1"

	// Step 1: Login to get token
	fmt.Println("🔐 Logging in...")
	token, err := login(baseURL)
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	fmt.Println("✅ Login successful")

	// Step 2: Reset counter
	fmt.Printf("🔄 Resetting counter for Invoice Type %s to %d for year %d...\n", invoiceTypeID, counter, year)
	err = resetCounter(baseURL, token, invoiceTypeID, year, counter)
	if err != nil {
		log.Fatalf("Failed to reset counter: %v", err)
	}
	fmt.Println("✅ Counter reset successful!")
}

func login(baseURL string) (string, error) {
	loginReq := LoginRequest{
		Email:    "admin@test.com", // Change this to your admin email
		Password: "password123",    // Change this to your admin password
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var loginResp LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	if err != nil {
		return "", err
	}

	if !loginResp.Success {
		return "", fmt.Errorf("login failed: %s", loginResp.Message)
	}

	return loginResp.Data.AccessToken, nil
}

func resetCounter(baseURL, token, invoiceTypeID string, year, counter int) error {
	resetReq := ResetCounterRequest{
		Year:    year,
		Counter: counter,
	}

	jsonData, err := json.Marshal(resetReq)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/invoice-types/%s/reset-counter", baseURL, invoiceTypeID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}

	fmt.Printf("📊 API Response: %+v\n", result)

	if resp.StatusCode != 200 {
		return fmt.Errorf("API call failed with status %d: %v", resp.StatusCode, result)
	}

	return nil
}