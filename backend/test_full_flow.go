package main

import (
	"fmt"
	"time"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/models"
)

func main() {
	// Load configuration  
	_ = config.LoadConfig()
	
	// Connect to database
	db := database.ConnectDB()
	
	fmt.Println("🧪 TESTING COMPLETE PURCHASE APPROVAL FLOW")
	
	// Initialize required repositories and services
	purchaseRepo := repositories.NewPurchaseRepository(db)
	productRepo := repositories.NewProductRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	approvalService := services.NewApprovalService(db)
	journalRepo := repositories.NewJournalEntryRepository(db)
	pdfService := services.NewPDFService(db)
	unifiedJournalService := services.NewUnifiedJournalService(db)
	coaService := services.NewCOAService(db)
	
	// Initialize purchase service with all dependencies
	purchaseService := services.NewPurchaseService(
		db,
		purchaseRepo,
		productRepo, 
		contactRepo,
		accountRepo,
		approvalService,
		nil, // journal service - can be nil for now
		journalRepo,
		pdfService,
		unifiedJournalService,
		coaService,
	)
	
	// Create a new test purchase
	fmt.Printf("\n📋 Creating a new test purchase...\n")
	
	// Get the first vendor and product for testing
	var vendorID uint
	db.Raw("SELECT id FROM contacts WHERE contact_type = 'VENDOR' LIMIT 1").Scan(&vendorID)
	
	var productID uint
	db.Raw("SELECT id FROM products WHERE stock > 0 LIMIT 1").Scan(&productID)
	
	if vendorID == 0 || productID == 0 {
		fmt.Printf("❌ No vendor or product found for testing\n")
		return
	}
	
	// Create purchase request
	purchaseReq := models.CreatePurchaseRequest{
		VendorID:      vendorID,
		Date:          time.Now().Format("2006-01-02"),
		DueDate:       time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
		PaymentMethod: models.PurchasePaymentBankTransfer, // Immediate payment method
		BankAccountID: 7,                                 // Assign bank account directly
		PPNRate:       11.0,
		PPHRate:       0.0,
		Items: []models.CreatePurchaseItemRequest{
			{
				ProductID: productID,
				Quantity:  2,
				UnitPrice: 1500000.0, // 1.5M per item
			},
		},
		Notes: "Test purchase for end-to-end approval flow",
	}
	
	userID := uint(1) // Admin user
	
	// Create the purchase
	purchase, err := purchaseService.CreatePurchase(purchaseReq, userID)
	if err != nil {
		fmt.Printf("❌ Failed to create purchase: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Purchase created: %s (ID: %d, Amount: %.2f)\n", 
		purchase.Code, purchase.ID, purchase.TotalAmount)
	
	// Step 1: Submit for approval
	fmt.Printf("\n🔄 Step 1: Submitting purchase for approval...\n")
	err = purchaseService.SubmitForApproval(purchase.ID, userID)
	if err != nil {
		fmt.Printf("❌ Submit for approval failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Purchase submitted for approval\n")
	
	// Get approval request ID
	var approvalRequestID *uint
	db.Raw("SELECT approval_request_id FROM purchases WHERE id = ?", purchase.ID).Scan(&approvalRequestID)
	
	if approvalRequestID == nil {
		fmt.Printf("❌ No approval request created\n")
		return
	}
	
	fmt.Printf("   Approval Request ID: %d\n", *approvalRequestID)
	
	// Check initial bank balance
	var initialBalance float64
	db.Raw("SELECT balance FROM cash_banks WHERE id = 7").Scan(&initialBalance)
	fmt.Printf("   Initial Bank Balance: %.2f\n", initialBalance)
	
	// Step 2: Finance approves the purchase (using new fixed method)
	fmt.Printf("\n🔄 Step 2: Finance user approving purchase...\n")
	financeUserID := uint(2) // Finance user
	
	result, err := purchaseService.ProcessPurchaseApprovalWithEscalation(
		purchase.ID, 
		true,        // approved
		financeUserID, 
		"finance",  // user role
		"Test approval via fixed flow",
		false,      // no escalation to director
	)
	if err != nil {
		fmt.Printf("❌ Approval processing failed: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Purchase approval processed: %v\n", result["message"])
	
	// Check final results
	fmt.Printf("\n📊 FINAL RESULTS:\n")
	
	// Check purchase status
	var finalPurchase struct {
		ID     uint   `json:"id"`
		Code   string `json:"code"`
		Status string `json:"status"`
		TotalAmount float64 `json:"total_amount"`
	}
	
	db.Raw("SELECT id, code, status, total_amount FROM purchases WHERE id = ?", purchase.ID).Scan(&finalPurchase)
	fmt.Printf("   Purchase Status: %s\n", finalPurchase.Status)
	
	// Check cash bank transactions
	var cbTxCount int64
	db.Raw("SELECT COUNT(*) FROM cash_bank_transactions WHERE reference_type = 'PURCHASE' AND reference_id = ?", purchase.ID).Scan(&cbTxCount)
	fmt.Printf("   Cash Bank Transactions: %d\n", cbTxCount)
	
	if cbTxCount > 0 {
		var newTx []struct {
			ID     uint    `json:"id"`
			Amount float64 `json:"amount"`
			Notes  string  `json:"notes"`
		}
		
		db.Raw(`
			SELECT id, amount, notes
			FROM cash_bank_transactions 
			WHERE reference_type = 'PURCHASE' AND reference_id = ?
		`, purchase.ID).Scan(&newTx)
		
		for _, tx := range newTx {
			fmt.Printf("     Transaction %d: %.2f - %s\n", tx.ID, tx.Amount, tx.Notes)
		}
	}
	
	// Check final bank balance
	var finalBalance float64
	db.Raw("SELECT balance FROM cash_banks WHERE id = 7").Scan(&finalBalance)
	fmt.Printf("   Final Bank Balance: %.2f\n", finalBalance)
	fmt.Printf("   Balance Change: %.2f\n", finalBalance - initialBalance)
	
	// Check journal entries
	var journalCount int64
	db.Raw("SELECT COUNT(*) FROM journal_entries WHERE reference LIKE ?", "%"+finalPurchase.Code+"%").Scan(&journalCount)
	fmt.Printf("   Journal Entries: %d\n", journalCount)
	
	// Summary
	fmt.Printf("\n🎯 END-TO-END TEST SUMMARY:\n")
	
	success := true
	
	if finalPurchase.Status == "APPROVED" {
		fmt.Printf("   ✅ Purchase Status: APPROVED\n")
	} else {
		fmt.Printf("   ❌ Purchase Status: %s (expected: APPROVED)\n", finalPurchase.Status)
		success = false
	}
	
	if cbTxCount > 0 {
		fmt.Printf("   ✅ Cash Bank Transactions: Created (%d)\n", cbTxCount)
	} else {
		fmt.Printf("   ❌ Cash Bank Transactions: None created\n")
		success = false
	}
	
	if finalBalance != initialBalance {
		fmt.Printf("   ✅ Bank Balance: Updated (change: %.2f)\n", finalBalance - initialBalance)
	} else {
		fmt.Printf("   ❌ Bank Balance: Unchanged\n")
		success = false
	}
	
	if journalCount > 0 {
		fmt.Printf("   ✅ Journal Entries: Created (%d)\n", journalCount)
	} else {
		fmt.Printf("   ❌ Journal Entries: None created\n")
		success = false
	}
	
	if success {
		fmt.Printf("\n🎉 SUCCESS: End-to-end purchase approval flow is working correctly!\n")
		fmt.Printf("   The fix for OnPurchaseApproved callback is permanent and complete.\n")
	} else {
		fmt.Printf("\n❌ ISSUES FOUND: Some aspects of the flow are not working correctly.\n")
		fmt.Printf("   Please check the logs above for specific problems.\n")
	}
}