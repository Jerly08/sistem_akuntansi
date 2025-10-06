package main

import (
	"fmt"
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
	
	fmt.Println("🧪 TESTING FIXED PURCHASE APPROVAL FLOW")
	
	// Test dengan purchase ID 1 (yang sudah approved tapi belum ada cash bank transaction)
	purchaseID := uint(1)
	
	// Get current purchase status
	fmt.Printf("\n📋 Current status of Purchase %d:\n", purchaseID)
	var currentPurchase struct {
		ID     uint   `json:"id"`
		Code   string `json:"code"`
		Status string `json:"status"`
		TotalAmount float64 `json:"total_amount"`
		PaymentMethod string `json:"payment_method"`
		BankAccountID *uint `json:"bank_account_id"`
	}
	
	db.Raw(`
		SELECT id, code, status, total_amount, payment_method, bank_account_id
		FROM purchases WHERE id = ?
	`, purchaseID).Scan(&currentPurchase)
	
	fmt.Printf("   Purchase: %s, Status: %s, Amount: %.2f\n", 
		currentPurchase.Code, currentPurchase.Status, currentPurchase.TotalAmount)
	fmt.Printf("   Payment Method: %s, Bank Account ID: %v\n", 
		currentPurchase.PaymentMethod, currentPurchase.BankAccountID)
	
	// Check cash bank transactions before
	var cbTxCountBefore int64
	db.Raw("SELECT COUNT(*) FROM cash_bank_transactions WHERE reference_type = 'PURCHASE' AND reference_id = ?", purchaseID).Scan(&cbTxCountBefore)
	fmt.Printf("   Cash Bank Transactions Before: %d\n", cbTxCountBefore)
	
	// Check bank balance before
	if currentPurchase.BankAccountID != nil {
		var balanceBefore float64
		db.Raw("SELECT balance FROM cash_banks WHERE id = ?", *currentPurchase.BankAccountID).Scan(&balanceBefore)
		fmt.Printf("   Bank Balance Before: %.2f\n", balanceBefore)
	} else {
		fmt.Printf("   Bank Balance Before: N/A (no bank account set)\n")
	}
	
	// If purchase is already APPROVED, let's manually trigger OnPurchaseApproved
	if currentPurchase.Status == "APPROVED" {
		fmt.Printf("\n🔄 Purchase already APPROVED, manually triggering OnPurchaseApproved callback...\n")
		
		// Initialize service
		purchaseService := services.NewPurchaseService(db, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		
		// Call OnPurchaseApproved directly
		err := purchaseService.OnPurchaseApproved(purchaseID)
		if err != nil {
			fmt.Printf("❌ OnPurchaseApproved callback failed: %v\n", err)
			return
		} else {
			fmt.Printf("✅ OnPurchaseApproved callback completed successfully\n")
		}
	} else if currentPurchase.Status == "DRAFT" {
		fmt.Printf("\n🔄 Purchase is DRAFT, testing complete approval flow...\n")
		
		// Submit for approval first
		purchaseService := services.NewPurchaseService(db, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		
		userID := uint(1) // Admin user
		err := purchaseService.SubmitForApproval(purchaseID, userID)
		if err != nil {
			fmt.Printf("❌ Submit for approval failed: %v\n", err)
			return
		}
		fmt.Printf("✅ Purchase submitted for approval\n")
		
		// Get approval request ID
		db.Raw("SELECT id, code, status, approval_request_id FROM purchases WHERE id = ?", purchaseID).Scan(&currentPurchase)
		
		if currentPurchase.BankAccountID == nil {
			// Get the approval request ID after submit
			var approvalRequestID *uint
			db.Raw("SELECT approval_request_id FROM purchases WHERE id = ?", purchaseID).Scan(&approvalRequestID)
			
			if approvalRequestID != nil {
				// Process approval
				approvalService := services.NewApprovalService(db)
				
				action := models.ApprovalActionDTO{
					Action:   "APPROVE",
					Comments: "Test approval via fixed flow",
				}
				
				err = approvalService.ProcessApprovalAction(*approvalRequestID, userID, action)
				if err != nil {
					fmt.Printf("❌ Approval processing failed: %v\n", err)
					return
				} else {
					fmt.Printf("✅ Purchase approved via approval workflow\n")
				}
			}
		}
	} else {
		fmt.Printf("\n⚠️ Purchase status is %s - cannot process approval\n", currentPurchase.Status)
		return
	}
	
	// Check results after
	fmt.Printf("\n📊 RESULTS AFTER PROCESSING:\n")
	
	// Check cash bank transactions after
	var cbTxCountAfter int64
	db.Raw("SELECT COUNT(*) FROM cash_bank_transactions WHERE reference_type = 'PURCHASE' AND reference_id = ?", purchaseID).Scan(&cbTxCountAfter)
	fmt.Printf("   Cash Bank Transactions After: %d\n", cbTxCountAfter)
	
	if cbTxCountAfter > cbTxCountBefore {
		fmt.Printf("   ✅ NEW cash bank transaction(s) created!\n")
		
		// Show the new transactions
		var newTx []struct {
			ID     uint    `json:"id"`
			Amount float64 `json:"amount"`
			Notes  string  `json:"notes"`
		}
		
		db.Raw(`
			SELECT id, amount, notes
			FROM cash_bank_transactions 
			WHERE reference_type = 'PURCHASE' AND reference_id = ?
		`, purchaseID).Scan(&newTx)
		
		for _, tx := range newTx {
			fmt.Printf("     Transaction %d: %.2f - %s\n", tx.ID, tx.Amount, tx.Notes)
		}
	} else {
		fmt.Printf("   ❌ NO new cash bank transactions created\n")
	}
	
	// Check bank balance after (if bank account exists)
	if currentPurchase.BankAccountID != nil {
		var balanceAfter float64
		db.Raw("SELECT balance FROM cash_banks WHERE id = ?", *currentPurchase.BankAccountID).Scan(&balanceAfter)
		fmt.Printf("   Bank Balance After: %.2f\n", balanceAfter)
	} else {
		// Check if purchase now has bank account assigned
		db.Raw("SELECT bank_account_id FROM purchases WHERE id = ?", purchaseID).Scan(&currentPurchase.BankAccountID)
		if currentPurchase.BankAccountID != nil {
			var balanceAfter float64
			db.Raw("SELECT balance FROM cash_banks WHERE id = ?", *currentPurchase.BankAccountID).Scan(&balanceAfter)
			fmt.Printf("   Bank Balance After: %.2f (Bank ID: %d)\n", balanceAfter, *currentPurchase.BankAccountID)
		} else {
			fmt.Printf("   Bank Balance After: N/A (no bank account assigned)\n")
		}
	}
	
	// Check journal entries
	var journalCount int64
	db.Raw("SELECT COUNT(*) FROM journal_entries WHERE reference LIKE ? OR (reference_type = 'PURCHASE' AND reference_id = ?)", 
		"%"+currentPurchase.Code+"%", purchaseID).Scan(&journalCount)
	fmt.Printf("   Journal Entries: %d\n", journalCount)
	
	fmt.Printf("\n🎯 SUMMARY:\n")
	if cbTxCountAfter > cbTxCountBefore {
		fmt.Printf("   ✅ SUCCESS: Cash bank transactions were created!\n")
	} else {
		fmt.Printf("   ❌ ISSUE: Cash bank transactions were NOT created\n")
	}
	
	if journalCount > 0 {
		fmt.Printf("   ✅ SUCCESS: Journal entries exist\n")
	} else {
		fmt.Printf("   ❌ ISSUE: No journal entries found\n")
	}
	
	fmt.Printf("\n🔧 If issues remain, check:\n")
	fmt.Printf("   1. Purchase payment method (should be BANK_TRANSFER for immediate payment)\n")
	fmt.Printf("   2. Bank account assignment on purchase\n")
	fmt.Printf("   3. OnPurchaseApproved callback execution\n")
}