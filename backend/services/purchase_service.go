package services

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math"
	"os"
	"time"
)

type PurchaseService struct {
	db              *gorm.DB
	purchaseRepo    *repositories.PurchaseRepository
	productRepo     *repositories.ProductRepository
	contactRepo     repositories.ContactRepository
	accountRepo     repositories.AccountRepository
	approvalService *ApprovalService
	journalService  JournalServiceInterface
	journalRepo     repositories.JournalEntryRepository
	pdfService      PDFServiceInterface
}

type PurchaseResult struct {
	Data       []models.Purchase `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

func NewPurchaseService(
	db *gorm.DB,
	purchaseRepo *repositories.PurchaseRepository,
	productRepo *repositories.ProductRepository,
	contactRepo repositories.ContactRepository,
	accountRepo repositories.AccountRepository,
	approvalService *ApprovalService,
	journalService JournalServiceInterface,
	journalRepo repositories.JournalEntryRepository,
	pdfService PDFServiceInterface,
) *PurchaseService {
	return &PurchaseService{
		db:              db,
		purchaseRepo:    purchaseRepo,
		productRepo:     productRepo,
		contactRepo:     contactRepo,
		accountRepo:     accountRepo,
		approvalService: approvalService,
		journalService:  journalService,
		journalRepo:     journalRepo,
		pdfService:      pdfService,
	}
}

// Purchase CRUD Operations

func (s *PurchaseService) GetPurchases(filter models.PurchaseFilter) (*PurchaseResult, error) {
	purchases, total, err := s.purchaseRepo.FindWithFilter(filter)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))

	return &PurchaseResult{
		Data:       purchases,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *PurchaseService) GetPurchaseByID(id uint) (*models.Purchase, error) {
	return s.purchaseRepo.FindByID(id)
}

func (s *PurchaseService) CreatePurchase(request models.PurchaseCreateRequest, userID uint) (*models.Purchase, error) {
	// Validate vendor exists
	_, err := s.contactRepo.GetByID(request.VendorID)
	if err != nil {
		return nil, errors.New("vendor not found")
	}

	// Generate purchase code
	code, err := s.generatePurchaseCode()
	if err != nil {
		return nil, err
	}

	// Create purchase entity
	purchase := &models.Purchase{
		Code:     code,
		VendorID: request.VendorID,
		UserID:   userID,
		Date:     request.Date,
		DueDate:  request.DueDate,
		Discount: request.Discount,
		// Payment method fields
		PaymentMethod:     getPaymentMethod(request.PaymentMethod),
		BankAccountID:     request.BankAccountID,
		CreditAccountID:   request.CreditAccountID,
		PaymentReference:  request.PaymentReference,
		// Tax rates from request (don't use legacy tax field directly)
		PPNRate:            request.PPNRate,
		OtherTaxAdditions:  request.OtherTaxAdditions,
		PPh21Rate:          request.PPh21Rate,
		PPh23Rate:          request.PPh23Rate,
		OtherTaxDeductions: request.OtherTaxDeductions,
		Status:             models.PurchaseStatusDraft,
		Notes:              request.Notes,
		ApprovalStatus:     models.PurchaseApprovalNotStarted,
		RequiresApproval:   false,
		// Initialize payment tracking fields
		PaidAmount:        0,
		OutstandingAmount: 0, // Will be set after total calculation
		MatchingStatus:    models.PurchaseMatchingPending,
	}

	// Calculate totals and create purchase items
	err = s.calculatePurchaseTotals(purchase, request.Items)
	if err != nil {
		return nil, err
	}

	// Determine approval basis and base amount for later use
	if s.approvalService != nil {
		s.setApprovalBasisAndBase(purchase)
	} else {
		// For testing purposes, set default values
		purchase.RequiresApproval = false
		purchase.ApprovalStatus = models.PurchaseApprovalNotRequired
	}

	// Save purchase, status will remain DRAFT
	createdPurchase, err := s.purchaseRepo.Create(purchase)
	if err != nil {
		return nil, err
	}

	// For credit purchases that don't require approval, create journal entries immediately
	// This ensures COA is updated correctly without waiting for approval
	// If approval is required, journal entries will be created after approval
	if purchase.PaymentMethod == models.PurchasePaymentCredit && !purchase.RequiresApproval {
		fmt.Printf("Creating immediate journal entries for credit purchase %s (no approval required)\n", purchase.Code)
		err = s.createAndPostPurchaseJournalEntries(createdPurchase, userID)
		if err != nil {
			fmt.Printf("Warning: Failed to create journal entries for credit purchase %d: %v\n", createdPurchase.ID, err)
			// Don't fail the purchase creation, but log the issue
		}
	}

	return s.GetPurchaseByID(createdPurchase.ID)
}

func (s *PurchaseService) UpdatePurchase(id uint, request models.PurchaseUpdateRequest, userID uint) (*models.Purchase, error) {
	purchase, err := s.purchaseRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check if purchase can be updated
	if purchase.Status != models.PurchaseStatusDraft && purchase.Status != models.PurchaseStatusPending {
		return nil, errors.New("purchase cannot be updated in current status")
	}

	// Update fields if provided
	if request.VendorID != nil {
		purchase.VendorID = *request.VendorID
	}
	if request.Date != nil {
		purchase.Date = *request.Date
	}
	if request.DueDate != nil {
		purchase.DueDate = *request.DueDate
	}
	if request.Discount != nil {
		purchase.Discount = *request.Discount
	}
	// Update tax rates from request (don't use legacy tax field directly)
	if request.PPNRate != nil {
		purchase.PPNRate = *request.PPNRate
	}
	if request.OtherTaxAdditions != nil {
		purchase.OtherTaxAdditions = *request.OtherTaxAdditions
	}
	if request.PPh21Rate != nil {
		purchase.PPh21Rate = *request.PPh21Rate
	}
	if request.PPh23Rate != nil {
		purchase.PPh23Rate = *request.PPh23Rate
	}
	if request.OtherTaxDeductions != nil {
		purchase.OtherTaxDeductions = *request.OtherTaxDeductions
	}
	if request.Notes != nil {
		purchase.Notes = *request.Notes
	}
	// Update payment method fields
	if request.PaymentMethod != nil {
		purchase.PaymentMethod = *request.PaymentMethod
	}
	if request.BankAccountID != nil {
		purchase.BankAccountID = request.BankAccountID
	}
	if request.CreditAccountID != nil {
		purchase.CreditAccountID = request.CreditAccountID
	}
	if request.PaymentReference != nil {
		purchase.PaymentReference = *request.PaymentReference
	}

	// Update items if provided
	if len(request.Items) > 0 {
		err = s.updatePurchaseItems(purchase, request.Items)
		if err != nil {
			return nil, err
		}
	}

	// Recalculate totals
	err = s.recalculatePurchaseTotals(purchase)
	if err != nil {
		return nil, err
	}
	// Re-evaluate approval base
	s.setApprovalBasisAndBase(purchase)

	// Save updated purchase
	updatedPurchase, err := s.purchaseRepo.Update(purchase)
	if err != nil {
		return nil, err
	}

	return s.GetPurchaseByID(updatedPurchase.ID)
}

func (s *PurchaseService) DeletePurchase(id uint) error {
	purchase, err := s.purchaseRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Allow deletion of draft purchases by all authorized roles
	// Allow deletion of non-draft purchases only by admin (validation handled at controller level)
	if purchase.Status != models.PurchaseStatusDraft {
		// This will require role-based validation in the controller
		// For now, we'll allow deletion and let controller handle admin check
	}

	return s.purchaseRepo.Delete(id)
}

// Approval Integration

func (s *PurchaseService) SubmitForApproval(id uint, userID uint) error {
	purchase, err := s.purchaseRepo.FindByID(id)
	if err != nil {
		return err
	}

	if purchase.Status != models.PurchaseStatusDraft {
		return errors.New("only draft purchases can be submitted for approval")
	}

	// Ensure approval base up-to-date
	s.setApprovalBasisAndBase(purchase)
	// Check if approval is required
	requiresApproval := s.checkIfApprovalRequired(purchase.ApprovalBaseAmount)
	if !requiresApproval {
		// No approval required, move directly to approved
		purchase.Status = models.PurchaseStatusApproved
		purchase.ApprovalStatus = models.PurchaseApprovalNotRequired
		_, err = s.purchaseRepo.Update(purchase)
		return err
	}

	// Create approval request
	err = s.createApprovalRequest(purchase, models.ApprovalPriorityNormal, userID)
	if err != nil {
		return err
	}

	// The approval workflow now starts from Employee step (step 1)
	// When Employee submits, we immediately progress to the next step (Finance/Manager)
	// This mimics the Employee "submitting" the purchase for approval
	if purchase.ApprovalRequestID != nil {
		// Automatically approve the Employee step since the Employee is submitting
		action := models.ApprovalActionDTO{
			Action:   "APPROVE",
			Comments: "Purchase submitted by Employee for approval",
		}
		err = s.approvalService.ProcessApprovalAction(*purchase.ApprovalRequestID, userID, action)
		if err != nil {
			return fmt.Errorf("failed to process employee submission step: %v", err)
		}
	}

	// Update purchase status
	now := time.Now()
	purchase.Status = models.PurchaseStatusPending // Change to PENDING instead of PENDING_APPROVAL
	purchase.ApprovalStatus = models.PurchaseApprovalPending
	purchase.RequiresApproval = true
	purchase.UpdatedAt = now

	_, err = s.purchaseRepo.Update(purchase)
	return err
}

func (s *PurchaseService) ProcessPurchaseApproval(purchaseID uint, approved bool, userID uint) error {
	purchase, err := s.purchaseRepo.FindByID(purchaseID)
	if err != nil {
		return err
	}

	if purchase.ApprovalStatus != models.PurchaseApprovalPending {
		return errors.New("purchase is not pending approval")
	}

	now := time.Now()
	if approved {
		// Purchase approved
		purchase.Status = models.PurchaseStatusApproved
		purchase.ApprovalStatus = models.PurchaseApprovalApproved
		purchase.ApprovedAt = &now
		purchase.ApprovedBy = &userID
	} else {
		// Purchase rejected
		purchase.Status = models.PurchaseStatusCancelled
		purchase.ApprovalStatus = models.PurchaseApprovalRejected
	}

	purchase.UpdatedAt = now
	_, err = s.purchaseRepo.Update(purchase)
	return err
}

// ProcessPurchaseApprovalWithEscalation processes purchase approval with escalation logic
func (s *PurchaseService) ProcessPurchaseApprovalWithEscalation(purchaseID uint, approved bool, userID uint, userRole, comments string, escalateToDirector bool) (map[string]interface{}, error) {
	purchase, err := s.purchaseRepo.FindByID(purchaseID)
	if err != nil {
		return nil, err
	}

	// Allow approval/rejection of DRAFT purchases (Finance approving new purchases)
	// and PENDING purchases (Director approving escalated purchases)
	// Also allow NOT_STARTED for rejection
	if purchase.Status != models.PurchaseStatusDraft &&
		purchase.Status != models.PurchaseStatusPending &&
		purchase.ApprovalStatus != models.PurchaseApprovalPending &&
		purchase.ApprovalStatus != models.PurchaseApprovalNotStarted {
		return nil, errors.New("purchase cannot be approved in current status")
	}

	now := time.Now()
	result := make(map[string]interface{})

	if !approved {
		// Purchase rejected
		purchase.Status = models.PurchaseStatusCancelled
		purchase.ApprovalStatus = models.PurchaseApprovalRejected
		purchase.UpdatedAt = now

		// If no approval request exists (DRAFT status), create one for history tracking
		if purchase.ApprovalRequestID == nil {
			// Create a minimal approval request for history tracking (without workflow dependency)
			err = s.createMinimalApprovalRequestForRejection(purchase, userID)
			if err != nil {
				// Continue even if this fails - the rejection should still proceed
			}
		}

		_, err = s.purchaseRepo.Update(purchase)
		if err != nil {
			return nil, err
		}

		// Create approval history record for rejection
		if purchase.ApprovalRequestID != nil {
			// First update the approval request status to rejected
			if approvalReq, err := s.approvalService.GetApprovalRequest(*purchase.ApprovalRequestID); err == nil {
				approvalReq.Status = models.ApprovalStatusRejected
				approvalReq.CompletedAt = &now
				approvalReq.RejectReason = comments
				s.approvalService.UpdateApprovalRequest(approvalReq)
			}

			// Ensure comments are not empty for rejection history
			historyComments := comments
			if historyComments == "" {
				historyComments = "Purchase rejected without comment"
			}

			historyErr := s.approvalService.CreateApprovalHistory(*purchase.ApprovalRequestID, userID, models.ApprovalActionRejected, historyComments)
			if historyErr != nil {
				// Log error but don't fail the entire operation
				fmt.Printf("Failed to create approval history for rejection: %v\n", historyErr)
				// Continue with fallback - directly insert into approval_histories table if needed
			}
		} else {
			fmt.Printf("Warning: Purchase %d rejected but no approval request ID found\n", purchaseID)
		}

		result["message"] = "Purchase rejected"
		result["purchase_id"] = purchaseID
		result["status"] = "REJECTED"
		result["rejected_by"] = userID
		result["rejected_at"] = now.Format(time.RFC3339)
		result["rejection_reason"] = comments
		return result, nil
	}

	// Purchase is approved, check for escalation
	if userRole == "finance" && escalateToDirector {
		// If no approval request exists (DRAFT status), create one
		if purchase.ApprovalRequestID == nil {
			err = s.createApprovalRequest(purchase, models.ApprovalPriorityNormal, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to create approval request: %v", err)
			}
			// Reload purchase to get the ApprovalRequestID
			purchase, err = s.purchaseRepo.FindByID(purchaseID)
			if err != nil {
				return nil, err
			}
		}

		// IMPORTANT: Escalate to director FIRST before processing approval
		// This ensures the request stays PENDING for director approval
		if purchase.ApprovalRequestID != nil {
			// First escalate to director to add director step
			err = s.approvalService.EscalateToDirector(*purchase.ApprovalRequestID, userID, "Requires Director approval as requested by Finance")
			if err != nil {
				return nil, fmt.Errorf("failed to escalate to director: %v", err)
			}

			// Then process the finance approval
			action := models.ApprovalActionDTO{
				Action:   "APPROVE",
				Comments: fmt.Sprintf("%s (Escalated to Director for final approval)", comments),
			}
			err = s.approvalService.ProcessApprovalAction(*purchase.ApprovalRequestID, userID, action)
			if err != nil {
				return nil, fmt.Errorf("failed to process finance approval: %v", err)
			}
		}

		// Purchase stays PENDING for director review
		purchase.Status = models.PurchaseStatusPending           // Keep as PENDING
		purchase.ApprovalStatus = models.PurchaseApprovalPending // Set to PENDING for director
		purchase.RequiresApproval = true                         // Mark as requiring approval

		purchase.UpdatedAt = now
		_, err = s.purchaseRepo.Update(purchase)
		if err != nil {
			return nil, err
		}

		result["message"] = "Purchase approved by Finance and escalated to Director for final approval"
		result["purchase_id"] = purchaseID
		result["escalated"] = true
		result["status"] = "PENDING"          // Status is PENDING
		result["approval_status"] = "PENDING" // But approval status is PENDING
		return result, nil
	}

	// Direct approval (no escalation needed)
	// If no approval request exists (DRAFT status), create one for history tracking
	if purchase.ApprovalRequestID == nil && purchase.Status == models.PurchaseStatusDraft {
		err = s.createApprovalRequest(purchase, models.ApprovalPriorityNormal, userID)
		if err != nil {
			fmt.Printf("Failed to create approval request: %v\n", err)
			// Continue even if this fails - the approval should still proceed
		}
		// Reload purchase to get the ApprovalRequestID
		purchase, err = s.purchaseRepo.FindByID(purchaseID)
		if err != nil {
			return nil, err
		}
	}

	// Create approval history
	if purchase.ApprovalRequestID != nil {
		// Update the approval request status to approved
		if approvalReq, err := s.approvalService.GetApprovalRequest(*purchase.ApprovalRequestID); err == nil {
			approvalReq.Status = models.ApprovalStatusApproved
			approvalReq.CompletedAt = &now
			s.approvalService.UpdateApprovalRequest(approvalReq)
		}

		historyErr := s.approvalService.CreateApprovalHistory(*purchase.ApprovalRequestID, userID, models.ApprovalActionApproved, comments)
		if historyErr != nil {
			fmt.Printf("Failed to create approval history: %v\n", historyErr)
		}
	}

	purchase.Status = models.PurchaseStatusApproved
	purchase.ApprovalStatus = models.PurchaseApprovalApproved
	purchase.ApprovedAt = &now
	purchase.ApprovedBy = &userID
	purchase.UpdatedAt = now

	_, err = s.purchaseRepo.Update(purchase)
	if err != nil {
		return nil, err
	}

	// Update product stock when purchase is approved
	fmt.Printf("🔄 Updating product stock for approved purchase %d\n", purchaseID)
	err = s.updateProductStockOnApproval(purchase)
	if err != nil {
		fmt.Printf("⚠️ Warning: Failed to update product stock for purchase %d: %v\n", purchaseID, err)
		// Don't fail the approval process, but log the issue
	} else {
		fmt.Printf("✅ Successfully updated product stock for purchase %d\n", purchaseID)
	}

	// Create and post journal entries for the approved purchase
	// Skip if journal entries were already created (e.g., for credit purchases without approval)
	hasExistingJournalEntries, checkErr := s.purchaseHasJournalEntries(purchaseID)
	if checkErr != nil {
		fmt.Printf("Warning: Failed to check existing journal entries for purchase %d: %v\n", purchaseID, checkErr)
		// Continue with journal creation to be safe
		hasExistingJournalEntries = false
	}

	if !hasExistingJournalEntries {
		fmt.Printf("Creating journal entries for approved purchase %d\n", purchaseID)
		err = s.createAndPostPurchaseJournalEntries(purchase, userID)
		if err != nil {
			fmt.Printf("Warning: Failed to create/post journal entries for purchase %d: %v\n", purchaseID, err)
			// Don't fail the approval process, but log the issue
		}
	} else {
		fmt.Printf("Journal entries already exist for purchase %d, skipping creation\n", purchaseID)
	}

	result["message"] = "Purchase approved successfully"
	result["purchase_id"] = purchaseID
	result["escalated"] = false
	result["status"] = "APPROVED"
	result["approval_status"] = "APPROVED"
	return result, nil
}

// Receipt Management

func (s *PurchaseService) CreatePurchaseReceipt(request models.PurchaseReceiptRequest, userID uint) (*models.PurchaseReceipt, error) {
	purchase, err := s.purchaseRepo.FindByID(request.PurchaseID)
	if err != nil {
		return nil, err
	}

	if purchase.Status != models.PurchaseStatusApproved && purchase.Status != models.PurchaseStatusPending {
		return nil, errors.New("can only receive items for approved or pending purchases")
	}

	// Generate receipt number
	receiptNumber := s.generateReceiptNumber()

	// Create receipt
	receipt := &models.PurchaseReceipt{
		PurchaseID:    request.PurchaseID,
		ReceiptNumber: receiptNumber,
		ReceivedDate:  request.ReceivedDate,
		ReceivedBy:    userID,
		Status:        models.ReceiptStatusPending,
		Notes:         request.Notes,
	}

	// Validate receipt items
	err = s.validateReceiptItems(request.ReceiptItems, purchase.PurchaseItems)
	if err != nil {
		return nil, err
	}

	// Create receipt with items
	createdReceipt, err := s.purchaseRepo.CreateReceipt(receipt)
	if err != nil {
		return nil, err
	}

	// Process receipt items and check completion
	allReceived := true
	for _, itemReq := range request.ReceiptItems {
		purchaseItem, err := s.purchaseRepo.GetPurchaseItemByID(itemReq.PurchaseItemID)
		if err != nil {
			return nil, err
		}
		
		if purchaseItem.PurchaseID != request.PurchaseID {
			return nil, errors.New("purchase item does not belong to this purchase")
		}

		// Create receipt item
		receiptItem := &models.PurchaseReceiptItem{
			ReceiptID:        createdReceipt.ID,
			PurchaseItemID:   itemReq.PurchaseItemID,
			QuantityReceived: itemReq.QuantityReceived,
			Condition:        s.getDefaultCondition(itemReq.Condition),
			Notes:            itemReq.Notes,
		}

		err = s.purchaseRepo.CreateReceiptItem(receiptItem)
		if err != nil {
			return nil, err
		}

		// Check if all items are fully received
		if itemReq.QuantityReceived < purchaseItem.Quantity {
			allReceived = false
		}
	}

	// Update receipt and purchase status directly
	if allReceived {
		createdReceipt.Status = models.ReceiptStatusComplete
		purchase.Status = models.PurchaseStatusCompleted
	} else {
		createdReceipt.Status = models.ReceiptStatusPartial
	}

	// Update receipt status
	createdReceipt, err = s.purchaseRepo.UpdateReceipt(createdReceipt)
	if err != nil {
		return nil, err
	}

	// Update purchase status if completed
	if allReceived {
		_, err = s.purchaseRepo.Update(purchase)
		if err != nil {
			return nil, err
		}
	}

	return s.purchaseRepo.FindReceiptByID(createdReceipt.ID)
}

// Document Management

func (s *PurchaseService) UploadDocument(purchaseID uint, documentType, fileName, filePath string, fileSize int64, mimeType string, userID uint) error {
	_, err := s.purchaseRepo.FindByID(purchaseID)
	if err != nil {
		return err
	}

	document := &models.PurchaseDocument{
		PurchaseID:   purchaseID,
		DocumentType: documentType,
		FileName:     fileName,
		FilePath:     filePath,
		FileSize:     fileSize,
		MimeType:     mimeType,
		UploadedBy:   userID,
	}

	return s.purchaseRepo.CreateDocument(document)
}

func (s *PurchaseService) GetPurchaseDocuments(purchaseID uint) ([]models.PurchaseDocument, error) {
	return s.purchaseRepo.FindDocumentsByPurchaseID(purchaseID)
}

func (s *PurchaseService) DeleteDocument(documentID uint) error {
	return s.purchaseRepo.DeleteDocument(documentID)
}

// Three-way Matching

func (s *PurchaseService) GetPurchaseMatching(purchaseID uint) (*models.PurchaseMatchingData, error) {
	return s.purchaseRepo.GetPurchaseForMatching(purchaseID)
}

func (s *PurchaseService) ValidateThreeWayMatching(purchaseID uint) (bool, error) {
	matching, err := s.purchaseRepo.GetPurchaseForMatching(purchaseID)
	if err != nil {
		return false, err
	}

	// Validate Purchase Order exists
	if matching.Purchase.ID == 0 {
		return false, errors.New("purchase order not found")
	}

	// Validate Receipt exists and is complete
	hasCompleteReceipt := false
	for _, receipt := range matching.Receipts {
		if receipt.Status == models.ReceiptStatusComplete {
			hasCompleteReceipt = true
			break
		}
	}

	if !hasCompleteReceipt {
		return false, errors.New("complete receipt required")
	}

	// Validate Invoice exists
	hasInvoice := false
	for _, doc := range matching.Documents {
		if doc.DocumentType == models.PurchaseDocumentInvoice {
			hasInvoice = true
			break
		}
	}

	if !hasInvoice {
		return false, errors.New("invoice document required")
	}

	// Update matching status
	err = s.purchaseRepo.UpdateMatchingStatus(purchaseID, "MATCHED")
	if err != nil {
		return false, err
	}

	return true, nil
}

// Analytics and Reporting

func (s *PurchaseService) GetPurchasesSummary(startDate, endDate string) (*models.PurchaseSummary, error) {
	return s.purchaseRepo.GetPurchasesSummary(startDate, endDate)
}

func (s *PurchaseService) GetVendorPurchaseSummary(vendorID uint) (*models.VendorPurchaseSummary, error) {
	return s.purchaseRepo.GetVendorPurchaseSummary(vendorID)
}

// Receipt Management - Additional methods

// GetPurchaseReceipts returns all receipts for a purchase
func (s *PurchaseService) GetPurchaseReceipts(purchaseID uint) ([]models.PurchaseReceipt, error) {
	// Verify purchase exists
	_, err := s.purchaseRepo.FindByID(purchaseID)
	if err != nil {
		return nil, err
	}

	// Get receipts from repository
	return s.purchaseRepo.FindReceiptsByPurchaseID(purchaseID)
}

// GetCompletedPurchaseReceipts returns only completed receipts for a purchase
func (s *PurchaseService) GetCompletedPurchaseReceipts(purchaseID uint) ([]models.PurchaseReceipt, error) {
	// Verify purchase exists
	_, err := s.purchaseRepo.FindByID(purchaseID)
	if err != nil {
		return nil, err
	}

	// Get completed receipts from repository
	return s.purchaseRepo.FindCompletedReceiptsByPurchaseID(purchaseID)
}

// GenerateReceiptPDF generates PDF for a specific receipt
func (s *PurchaseService) GenerateReceiptPDF(receiptID uint) ([]byte, *models.PurchaseReceipt, error) {
	// Get receipt with all related data
	receipt, err := s.purchaseRepo.FindReceiptByID(receiptID)
	if err != nil {
		return nil, nil, err
	}

	// Check if PDF service is available
	if s.pdfService == nil {
		return nil, nil, errors.New("PDF service not available")
	}

	// Generate PDF using the PDF service
	pdfBytes, err := s.pdfService.GenerateReceiptPDF(receipt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate receipt PDF: %v", err)
	}

	return pdfBytes, receipt, nil
}

// GenerateAllReceiptsPDF generates combined PDF for all receipts of a purchase
func (s *PurchaseService) GenerateAllReceiptsPDF(purchaseID uint) ([]byte, *models.Purchase, error) {
	// Get purchase with all related data
	purchase, err := s.purchaseRepo.FindByID(purchaseID)
	if err != nil {
		return nil, nil, err
	}

	// Get all receipts for this purchase
	receipts, err := s.GetPurchaseReceipts(purchaseID)
	if err != nil {
		return nil, nil, err
	}

	if len(receipts) == 0 {
		return nil, nil, errors.New("no receipts found for this purchase")
	}

	// Check if PDF service is available
	if s.pdfService == nil {
		return nil, nil, errors.New("PDF service not available")
	}

	// Generate combined PDF using the PDF service
	pdfBytes, err := s.pdfService.GenerateAllReceiptsPDF(purchase, receipts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate combined receipts PDF: %v", err)
	}

	return pdfBytes, purchase, nil
}

// Private helper methods

func (s *PurchaseService) checkIfApprovalRequired(amount float64) bool {
	// Check if there's an active workflow for this amount
	workflow, err := s.approvalService.GetWorkflowByAmount(models.ApprovalModulePurchase, amount)
	return err == nil && workflow != nil
}

func (s *PurchaseService) createApprovalRequest(purchase *models.Purchase, priority string, userID uint) error {
	// Ensure vendor is loaded
	vendorName := "Unknown"
	if purchase.Vendor.ID != 0 {
		// Vendor is already loaded
		vendorName = purchase.Vendor.Name
	} else {
		// Try to load vendor if not already loaded
		vendor, err := s.contactRepo.GetByID(purchase.VendorID)
		if err == nil {
			vendorName = vendor.Name
		}
	}

	// Create approval request
	approvalReq := models.CreateApprovalRequestDTO{
		EntityType:     models.EntityTypePurchase,
		EntityID:       purchase.ID,
		Amount:         purchase.ApprovalBaseAmount,
		Priority:       priority,
		RequestTitle:   fmt.Sprintf("Purchase Approval - %s (Vendor: %s)", purchase.Code, vendorName),
		RequestMessage: fmt.Sprintf("Approval request for purchase %s with base amount %.2f (basis: %s)", purchase.Code, purchase.ApprovalBaseAmount, purchase.ApprovalAmountBasis),
	}

	// Determine priority based on amount
	if purchase.TotalAmount > 50000000 { // 50M IDR
		approvalReq.Priority = models.ApprovalPriorityUrgent
	} else if purchase.TotalAmount > 25000000 { // 25M IDR
		approvalReq.Priority = models.ApprovalPriorityHigh
	} else {
		approvalReq.Priority = models.ApprovalPriorityNormal
	}

	approvalRequest, err := s.approvalService.CreateApprovalRequest(approvalReq, userID)
	if err != nil {
		// Log the error for debugging
		fmt.Printf("Failed to create approval request: %v\n", err)
		return fmt.Errorf("failed to create approval request: %v", err)
	}

	// Check if approvalRequest is nil
	if approvalRequest == nil {
		return errors.New("approval request creation returned nil")
	}

	// Update purchase with approval request ID
	purchase.ApprovalRequestID = &approvalRequest.ID
	_, err = s.purchaseRepo.Update(purchase)
	return err
}

func (s *PurchaseService) validateReceiptItems(receiptItems []models.PurchaseReceiptItemRequest, purchaseItems []models.PurchaseItem) error {
	for _, receiptItem := range receiptItems {
		found := false
		for _, purchaseItem := range purchaseItems {
			if purchaseItem.ID == receiptItem.PurchaseItemID {
				if receiptItem.QuantityReceived > purchaseItem.Quantity {
					return fmt.Errorf("received quantity cannot exceed ordered quantity for item %d", receiptItem.PurchaseItemID)
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("purchase item %d not found", receiptItem.PurchaseItemID)
		}
	}
	return nil
}

func (s *PurchaseService) updateReceiptStatus(receiptID uint) error {
	receipt, err := s.purchaseRepo.FindReceiptByID(receiptID)
	if err != nil {
		return err
	}

	// Get the purchase to check completion status
	purchase, err := s.purchaseRepo.FindByID(receipt.PurchaseID)
	if err != nil {
		return err
	}

	// Get all receipt items for this receipt
	receiptItems, err := s.purchaseRepo.GetReceiptItems(receiptID)
	if err != nil {
		return err
	}

	// Check if all purchase items are fully received
	allReceived := true
	for _, purchaseItem := range purchase.PurchaseItems {
		totalReceived := 0
		// Sum up all received quantities for this purchase item across all receipts
		for _, receiptItem := range receiptItems {
			if receiptItem.PurchaseItemID == purchaseItem.ID {
				totalReceived += receiptItem.QuantityReceived
			}
		}
		if totalReceived < purchaseItem.Quantity {
			allReceived = false
			break
		}
	}

	// Update receipt status
	if allReceived {
		receipt.Status = models.ReceiptStatusComplete
		// Update purchase status to COMPLETED
		purchase.Status = models.PurchaseStatusCompleted
	} else {
		receipt.Status = models.ReceiptStatusPartial
	}

	// Update both receipt and purchase
	_, err = s.purchaseRepo.UpdateReceipt(receipt)
	if err != nil {
		return err
	}

	if allReceived {
		_, err = s.purchaseRepo.Update(purchase)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PurchaseService) getDefaultCondition(condition string) string {
	if condition == "" {
		return models.ReceiptConditionGood
	}
	return condition
}


// createMinimalApprovalRequestForRejection creates a minimal approval request for rejection tracking without workflow dependency
func (s *PurchaseService) createMinimalApprovalRequestForRejection(purchase *models.Purchase, userID uint) error {
	// Ensure vendor is loaded
	vendorName := "Unknown"
	if purchase.Vendor.ID != 0 {
		// Vendor is already loaded
		vendorName = purchase.Vendor.Name
	} else {
		// Try to load vendor if not already loaded
		vendor, err := s.contactRepo.GetByID(purchase.VendorID)
		if err == nil {
			vendorName = vendor.Name
		}
	}

	// Create approval request directly in approval service without workflow dependency
	return s.approvalService.CreateMinimalApprovalRequestForRejection(
		models.EntityTypePurchase,
		purchase.ID,
		purchase.ApprovalBaseAmount,
		fmt.Sprintf("Purchase Rejection Tracking - %s (Vendor: %s)", purchase.Code, vendorName),
		userID,
		purchase,
	)
}

// getApprovalBasis reads basis from env var APPROVAL_AMOUNT_BASIS
func getApprovalBasis() string {
	basis := os.Getenv("APPROVAL_AMOUNT_BASIS")
	if basis == "" {
		return "SUBTOTAL_BEFORE_DISCOUNT"
	}
	return basis
}

// getPaymentMethod returns default payment method if empty
func getPaymentMethod(paymentMethod string) string {
	if paymentMethod == "" {
		return models.PurchasePaymentCredit // Default to credit
	}
	return paymentMethod
}

// purchaseHasJournalEntries checks if a purchase already has associated journal entries
func (s *PurchaseService) purchaseHasJournalEntries(purchaseID uint) (bool, error) {
	if s.journalRepo == nil {
		return false, errors.New("journal repository not available")
	}
	
	// Use FindByReferenceID which is specifically designed for finding entries by reference
	ctx := context.Background()
	existingEntry, err := s.journalRepo.FindByReferenceID(ctx, models.JournalRefPurchase, purchaseID)
	if err != nil {
		return false, err
	}
	
	return existingEntry != nil, nil
}

// isImmediatePayment checks if payment method requires immediate payment
func isImmediatePayment(paymentMethod string) bool {
	return paymentMethod == models.PurchasePaymentCash ||
		paymentMethod == models.PurchasePaymentTransfer ||
		paymentMethod == models.PurchasePaymentCheck
}

// updateProductStockOnApproval updates product stock when purchase is approved
func (s *PurchaseService) updateProductStockOnApproval(purchase *models.Purchase) error {
	fmt.Printf("📦 Starting stock update for purchase %s with %d items\n", purchase.Code, len(purchase.PurchaseItems))
	
	for _, item := range purchase.PurchaseItems {
		// Get current product data
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			fmt.Printf("❌ Error finding product %d: %v\n", item.ProductID, err)
			continue // Skip this item but continue with others
		}
		
		fmt.Printf("📋 Product %d (%s): Current stock = %d, Adding quantity = %d\n", 
			product.ID, product.Name, product.Stock, item.Quantity)
		
		// Update stock quantity (add purchased quantity)
		oldStock := product.Stock
		product.Stock += item.Quantity
		
		// Update cost price using weighted average if we have existing stock
		if oldStock > 0 {
			// Weighted average: (old_stock * old_price + new_qty * new_price) / total_qty
			totalValue := (float64(oldStock) * product.PurchasePrice) + (float64(item.Quantity) * item.UnitPrice)
			totalQuantity := oldStock + item.Quantity
			product.PurchasePrice = totalValue / float64(totalQuantity)
			fmt.Printf("💰 Updated weighted average price: %.2f (was %.2f)\n", 
				product.PurchasePrice, (float64(oldStock) * product.PurchasePrice) / float64(oldStock))
		} else {
			// If no existing stock, use new price
			product.PurchasePrice = item.UnitPrice
			fmt.Printf("💰 Set new purchase price: %.2f\n", product.PurchasePrice)
		}
		
		// Save updated product
		err = s.productRepo.Update(context.Background(), product)
		if err != nil {
			fmt.Printf("❌ Failed to update product %d stock: %v\n", product.ID, err)
			return fmt.Errorf("failed to update stock for product %d: %v", product.ID, err)
		}
		
		fmt.Printf("✅ Product %d stock updated: %d → %d\n", product.ID, oldStock, product.Stock)
	}
	
	fmt.Printf("🎉 Stock update completed for purchase %s\n", purchase.Code)
	return nil
}
