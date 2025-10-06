package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/handlers"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration  
	_ = config.LoadConfig()
	
	// Connect to database
	db := database.ConnectDB()
	
	fmt.Println("🔍 TESTING API HIERARCHY RESPONSE DIRECTLY")
	fmt.Println("===========================================")
	
	// First check database directly
	fmt.Printf("\n1️⃣ DIRECT DATABASE CHECK:\n")
	var bankMandiri struct {
		ID      int     `json:"id"`
		Code    string  `json:"code"`
		Name    string  `json:"name"`
		Balance float64 `json:"balance"`
	}
	
	db.Raw("SELECT id, code, name, balance FROM accounts WHERE code = '1103'").Scan(&bankMandiri)
	fmt.Printf("   Database Bank Mandiri (1103): %.2f\n", bankMandiri.Balance)
	
	// Test repository GetHierarchy directly
	fmt.Printf("\n2️⃣ TESTING REPOSITORY GetHierarchy():\n")
	
	accountRepo := repositories.NewAccountRepository(db)
	
	hierarchy, err := accountRepo.GetHierarchy(nil)
	if err != nil {
		fmt.Printf("   ❌ Repository error: %v\n", err)
		return
	}
	
	fmt.Printf("   ✅ Repository returned %d root accounts\n", len(hierarchy))
	
	// Find Bank Mandiri in repository result
	var findBankMandiri func(accounts []interface{}) *interface{}
	findBankMandiri = func(accounts []interface{}) *interface{} {
		for i, acc := range accounts {
			account := acc.(map[string]interface{})
			if code, ok := account["code"].(string); ok && code == "1103" {
				return &accounts[i]
			}
			if children, ok := account["children"].([]interface{}); ok {
				if found := findBankMandiri(children); found != nil {
					return found
				}
			}
		}
		return nil
	}
	
	// Convert to interface{} for searching
	hierarchyInterface := make([]interface{}, len(hierarchy))
	hierarchyBytes, _ := json.Marshal(hierarchy)
	json.Unmarshal(hierarchyBytes, &hierarchyInterface)
	
	if found := findBankMandiri(hierarchyInterface); found != nil {
		foundAccount := (*found).(map[string]interface{})
		fmt.Printf("   🎯 Repository Bank Mandiri (1103):\n")
		fmt.Printf("      Balance: %.2f\n", foundAccount["balance"].(float64))
		fmt.Printf("      Total Balance: %.2f\n", foundAccount["total_balance"].(float64))
		fmt.Printf("      Is Header: %v\n", foundAccount["is_header"])
	} else {
		fmt.Printf("   ❌ Bank Mandiri not found in repository result\n")
	}
	
	// Test API Handler directly
	fmt.Printf("\n3️⃣ TESTING API HANDLER GetAccountHierarchy():\n")
	
	exportService := services.NewExportService(accountRepo, db)
	accountHandler := handlers.NewAccountHandler(accountRepo, exportService)
	
	// Create a test HTTP request
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	req := httptest.NewRequest("GET", "/api/v1/accounts/hierarchy", nil)
	c.Request = req
	
	// Call the handler
	accountHandler.GetAccountHierarchy(c)
	
	fmt.Printf("   Response Status: %d\n", w.Code)
	
	if w.Code == http.StatusOK {
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			fmt.Printf("   ❌ Error parsing JSON: %v\n", err)
			fmt.Printf("   Raw response: %s\n", w.Body.String()[:500])
		} else {
			fmt.Printf("   ✅ JSON parsed successfully\n")
			
			// Look for Bank Mandiri in API response
			if data, ok := response["data"].([]interface{}); ok {
				fmt.Printf("   📊 API returned %d root accounts\n", len(data))
				
				if found := findBankMandiri(data); found != nil {
					foundAccount := (*found).(map[string]interface{})
					fmt.Printf("   🎯 API Bank Mandiri (1103):\n")
					fmt.Printf("      Balance: %.2f\n", foundAccount["balance"].(float64))
					if totalBalance, ok := foundAccount["total_balance"]; ok {
						fmt.Printf("      Total Balance: %.2f\n", totalBalance.(float64))
					}
					fmt.Printf("      Is Header: %v\n", foundAccount["is_header"])
					
					// Check if it's correct
					actualBalance := foundAccount["balance"].(float64)
					if actualBalance == 44450000 {
						fmt.Printf("      ✅ API Balance is CORRECT!\n")
					} else {
						fmt.Printf("      ❌ API Balance is WRONG! Expected: 44,450,000, Got: %.0f\n", actualBalance)
					}
				} else {
					fmt.Printf("   ❌ Bank Mandiri not found in API response\n")
				}
			} else {
				fmt.Printf("   ❌ No data array in API response\n")
				fmt.Printf("   Response keys: %v\n", getKeys(response))
			}
		}
	} else {
		fmt.Printf("   ❌ API call failed with status: %d\n", w.Code)
		fmt.Printf("   Response: %s\n", w.Body.String())
	}
	
	fmt.Printf("\n💡 DEBUGGING HINTS:\n")
	if bankMandiri.Balance == 44450000 {
		fmt.Printf("✅ Database is correct (44,450,000)\n")
		fmt.Printf("❓ Check if repository or handler is modifying the balance\n")
		fmt.Printf("❓ Look for any SSOT balance calculation logic\n")
	} else {
		fmt.Printf("❌ Database itself might be wrong\n")
	}
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}