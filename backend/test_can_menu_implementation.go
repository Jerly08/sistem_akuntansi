package main

import (
	"fmt"
	"log"
	"encoding/json"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Testing CanMenu Permission Implementation")
	fmt.Println("========================================")

	// Test 1: Check if employee role has correct default permissions
	fmt.Println("\n1. Testing Employee Default Permissions:")
	employeePerms := models.GetDefaultPermissions("employee")
	
	// Employee should have view access to contacts but NO menu access
	if contactPerm, exists := employeePerms["contacts"]; exists {
		fmt.Printf("   Contacts - CanView: %v, CanMenu: %v", contactPerm.CanView, contactPerm.CanMenu)
		if contactPerm.CanView && !contactPerm.CanMenu {
			fmt.Println(" ✅ CORRECT: Can view data but not access menu")
		} else {
			fmt.Println(" ❌ WRONG: Should have view=true, menu=false")
		}
	}

	// Employee should have both view and menu access to products
	if productPerm, exists := employeePerms["products"]; exists {
		fmt.Printf("   Products - CanView: %v, CanMenu: %v", productPerm.CanView, productPerm.CanMenu)
		if productPerm.CanView && productPerm.CanMenu {
			fmt.Println(" ✅ CORRECT: Can view data and access menu")
		} else {
			fmt.Println(" ❌ WRONG: Should have view=true, menu=true")
		}
	}

	// Employee should have both view and menu access to purchases
	if purchasePerm, exists := employeePerms["purchases"]; exists {
		fmt.Printf("   Purchases - CanView: %v, CanMenu: %v", purchasePerm.CanView, purchasePerm.CanMenu)
		if purchasePerm.CanView && purchasePerm.CanMenu {
			fmt.Println(" ✅ CORRECT: Can view data and access menu")
		} else {
			fmt.Println(" ❌ WRONG: Should have view=true, menu=true")
		}
	}

	// Test 2: Check Admin permissions (should have all access)
	fmt.Println("\n2. Testing Admin Default Permissions:")
	adminPerms := models.GetDefaultPermissions("admin")
	
	testModules := []string{"contacts", "products", "purchases", "accounts"}
	allCorrect := true
	
	for _, module := range testModules {
		if perm, exists := adminPerms[module]; exists {
			fmt.Printf("   %s - CanView: %v, CanMenu: %v", module, perm.CanView, perm.CanMenu)
			if perm.CanView && perm.CanMenu {
				fmt.Println(" ✅")
			} else {
				fmt.Println(" ❌")
				allCorrect = false
			}
		}
	}
	
	if allCorrect {
		fmt.Println("   Admin permissions: ALL CORRECT")
	}

	// Test 3: Check specific employee user if exists
	fmt.Println("\n3. Testing Actual Employee User (if exists):")
	
	var employee models.User
	result := db.Where("role = ?", "employee").First(&employee)
	
	if result.Error == nil {
		fmt.Printf("   Found employee: %s (ID: %d)\n", employee.Username, employee.ID)
		
		// Check their actual permissions in database
		var permissions []models.ModulePermissionRecord
		db.Where("user_id = ?", employee.ID).Find(&permissions)
		
		if len(permissions) > 0 {
			fmt.Println("   Current database permissions:")
			for _, perm := range permissions {
				fmt.Printf("   - %s: CanView=%v, CanMenu=%v\n", 
					perm.Module, perm.CanView, perm.CanMenu)
			}
		} else {
			fmt.Println("   No custom permissions set, would use defaults")
		}
	} else {
		fmt.Println("   No employee users found in database")
	}

	// Test 4: Verify database schema
	fmt.Println("\n4. Testing Database Schema:")
	
	// Try to query a permission record to see if can_menu column exists
	var testPerm models.ModulePermissionRecord
	result = db.Select("can_menu").First(&testPerm)
	
	if result.Error != nil && result.Error.Error() != "record not found" {
		fmt.Printf("   ❌ Database schema issue: %v\n", result.Error)
		fmt.Println("   You may need to run the migration script first!")
	} else {
		fmt.Println("   ✅ Database schema supports can_menu column")
	}

	fmt.Println("\n========================================")
	fmt.Println("Test Summary:")
	fmt.Println("- Employee should see contacts data in purchase forms")
	fmt.Println("- Employee should NOT see Contacts in navigation menu")
	fmt.Println("- Employee should see Products and Purchases in menu")
	fmt.Println("- Admin should have full access to everything")
}