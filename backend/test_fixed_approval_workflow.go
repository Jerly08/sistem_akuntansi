package main

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/models"
)

func main() {
	// Load configuration  
	_ = config.LoadConfig()
	
	// Connect to database
	db := database.ConnectDB()
	
	fmt.Println("🧪 Testing fixed approval workflow...")
	
	// Initialize approval service
	approvalService := services.NewApprovalService(db)
	
	// Test the fixed workflow with the pending requests
	requestIDs := []uint{24, 25}
	
	for _, requestID := range requestIDs {
		fmt.Printf("\n📋 Testing approval workflow for request %d:\n", requestID)
		
		// Get current state
		var purchase struct {
			ID uint `json:"id"`
			Code string `json:"code"`
			Status string `json:"status"`
		}
		
		err := db.Raw(`
			SELECT p.id, p.code, p.status 
			FROM purchases p 
			JOIN approval_requests ar ON p.approval_request_id = ar.id 
			WHERE ar.id = ?
		`, requestID).Scan(&purchase).Error
		
		if err != nil {
			log.Printf("Error getting purchase for request %d: %v", requestID, err)
			continue
		}
		
		fmt.Printf("  Purchase: %s, Status: %s\n", purchase.Code, purchase.Status)
		
		if purchase.Status != "PENDING" {
			fmt.Printf("  ℹ️  Purchase is not PENDING, skipping test\n")
			continue
		}
		
		// Simulate finance user approving the request
		financeUserID := uint(2) // Finance user ID from logs
		
		fmt.Printf("  🚀 Finance user approving request %d...\n", requestID)
		
		action := models.ApprovalActionDTO{
			Action:   "APPROVE",
			Comments: "Approved by Finance - Testing fixed workflow",
		}
		
		err = approvalService.ProcessApprovalAction(requestID, financeUserID, action)
		if err != nil {
			fmt.Printf("  ❌ Approval failed: %v\n", err)
			continue
		}
		
		fmt.Printf("  ✅ Approval action completed\n")
		
		// Check the result
		var updatedPurchase struct {
			ID uint `json:"id"`
			Code string `json:"code"`
			Status string `json:"status"`
			ApprovalStatus string `json:"approval_status"`
		}
		
		err = db.Raw(`
			SELECT p.id, p.code, p.status, p.approval_status
			FROM purchases p 
			JOIN approval_requests ar ON p.approval_request_id = ar.id 
			WHERE ar.id = ?
		`, requestID).Scan(&updatedPurchase).Error
		
		if err != nil {
			log.Printf("Error getting updated purchase: %v", err)
			continue
		}
		
		fmt.Printf("  📊 Result: %s, Status: %s, Approval: %s\n", 
			updatedPurchase.Code, updatedPurchase.Status, updatedPurchase.ApprovalStatus)
		
		// Check if cash bank transaction was created (should be if approved)
		if updatedPurchase.Status == "APPROVED" {
			var txCount int64
			db.Raw("SELECT COUNT(*) FROM cash_bank_transactions WHERE reference_type = 'PURCHASE' AND reference_id = ?", 
				updatedPurchase.ID).Scan(&txCount)
				
			if txCount > 0 {
				fmt.Printf("  ✅ Cash bank transaction created: %d records\n", txCount)
			} else {
				fmt.Printf("  ⏳ Cash bank transaction may be processing asynchronously...\n")
			}
		}
		
		// Check approval actions consistency  
		var actions []struct {
			StepOrder int `json:"step_order"`
			ApproverRole string `json:"approver_role"`
			Status string `json:"status"`
			IsActive bool `json:"is_active"`
		}
		
		err = db.Raw(`
			SELECT astp.step_order, astp.approver_role, aa.status, aa.is_active
			FROM approval_actions aa
			JOIN approval_steps astp ON aa.step_id = astp.id
			WHERE aa.request_id = ?
			ORDER BY astp.step_order
		`, requestID).Scan(&actions).Error
		
		if err == nil {
			fmt.Printf("  📋 Approval actions state:\n")
			for _, action := range actions {
				fmt.Printf("    Step %d (%s): %s, Active: %t\n", 
					action.StepOrder, action.ApproverRole, action.Status, action.IsActive)
			}
		}
		
		// Check approval request status
		var requestStatus struct {
			Status string `json:"status"`
			CompletedAt *string `json:"completed_at"`
		}
		
		db.Raw("SELECT status, completed_at FROM approval_requests WHERE id = ?", requestID).Scan(&requestStatus)
		fmt.Printf("  📋 Request status: %s, Completed: %v\n", requestStatus.Status, requestStatus.CompletedAt != nil)
	}
	
	fmt.Println("\n📊 SUMMARY:")
	fmt.Println("  If the approval workflow is fixed correctly:")
	fmt.Println("  1. ✅ Purchase status should be APPROVED")
	fmt.Println("  2. ✅ approval_actions should show finance step as APPROVED")
	fmt.Println("  3. ✅ approval_requests should be APPROVED with completed_at set")
	fmt.Println("  4. ✅ Cash bank transactions should be created")
	fmt.Println("  5. ✅ No inconsistency between request status and action status")
	
	fmt.Println("\n🎯 Fixed approval workflow test completed!")
}