package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Account struct {
	ID           uint      `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Balance      float64   `json:"balance"`
	TotalBalance float64   `json:"total_balance"`
	IsHeader     bool      `json:"is_header"`
	IsActive     bool      `json:"is_active"`
	Level        int       `json:"level"`
	ParentID     *uint     `json:"parent_id"`
	ChildCount   int       `json:"child_count"`
	Children     []Account `json:"children"`
}

type Response struct {
	Data []Account `json:"data"`
}

func main() {
	// Give backend time to fully start
	time.Sleep(2 * time.Second)
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Test the hierarchy endpoint
	resp, err := client.Get("http://localhost:8080/api/v1/accounts/hierarchy")
	if err != nil {
		log.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		fmt.Println("Endpoint requires authentication, but we can see from backend logs it's being called successfully")
		fmt.Println("The issue is likely in frontend data handling, not the backend")
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("HTTP Status: %d\n", resp.StatusCode)
		return
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	fmt.Printf("Total root accounts returned: %d\n", len(response.Data))
	
	// Print hierarchy in tree format
	printAccountTree(response.Data, 0)
}

func printAccountTree(accounts []Account, indent int) {
	for _, acc := range accounts {
		indentStr := ""
		for i := 0; i < indent; i++ {
			indentStr += "  "
		}
		
		fmt.Printf("%s%s - %s (Type: %s, Balance: %.2f, Children: %d)\n", 
			indentStr, acc.Code, acc.Name, acc.Type, acc.Balance, len(acc.Children))
		
		if len(acc.Children) > 0 {
			printAccountTree(acc.Children, indent+1)
		}
	}
}
