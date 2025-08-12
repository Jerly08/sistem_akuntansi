package services

import (
	"errors"
	"fmt"
	"math"
	"os"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
)

type PurchaseService struct {
	purchaseRepo    *repositories.PurchaseRepository
	productRepo     *repositories.ProductRepository
	contactRepo     repositories.ContactRepository
	accountRepo     repositories.AccountRepository
	approvalService *ApprovalService
	journalService  JournalServiceInterface
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
	purchaseRepo *repositories.PurchaseRepository,
	productRepo *repositories.ProductRepository,
	contactRepo repositories.ContactRepository,
	accountRepo repositories.AccountRepository,
	approvalService *ApprovalService,
	journalService JournalServiceInterface,
	pdfService PDFServiceInterface,
) *PurchaseService {
	return &PurchaseService{
		purchaseRepo:    purchaseRepo,
		productRepo:     productRepo,
		contactRepo:     contactRepo,
		accountRepo:     accountRepo,
		approvalService: approvalService,
		journalService:  journalService,
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
		Code:             code,
		VendorID:         request.VendorID,
		UserID:           userID,
		Date:             request.Date,
		DueDate:          request.DueDate,
		Discount:         request.Discount,
		TaxAmount:        request.Tax,
		Status:           models.PurchaseStatusDraft,
		Notes:            request.Notes,
		ApprovalStatus:   models.PurchaseApprovalNotStarted,
		RequiresApproval: false,
	}

	// Calculate totals and create purchase items
	err = s.calculatePurchaseTotals(purchase, request.Items)
	if err != nil {
		return nil, err
	}

	// Determine approval basis and base amount for later use
	s.setApprovalBasisAndBase(purchase)

	// Save purchase, status will remain DRAFT
	createdPurchase, err := s.purchaseRepo.Create(purchase)
	if err != nil {
		return nil, err
	}

	return s.GetPurchaseByID(createdPurchase.ID)
}

func (s *PurchaseService) UpdatePurchase(id uint, request models.PurchaseUpdateRequest, userID uint) (*models.Purchase, error) {
	purchase, err := s.purchaseRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check if purchase can be updated
	if purchase.Status != models.PurchaseStatusDraft && purchase.Status != models.PurchaseStatusPendingApproval {
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
	if request.Tax != nil {
		purchase.TaxAmount = *request.Tax
	}
	if request.Notes != nil {
		purchase.Notes = *request.Notes
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

	// Update purchase status
	now := time.Now()
	purchase.Status = models.PurchaseStatusPendingApproval
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
		
		// Create approval history for finance approval with escalation note
		if purchase.ApprovalRequestID != nil {
			historyErr := s.approvalService.CreateApprovalHistory(*purchase.ApprovalRequestID, userID, models.ApprovalActionApproved, 
				fmt.Sprintf("%s (Escalated to Director)", comments))
			if historyErr != nil {
				fmt.Printf("Failed to create approval history: %v\n", historyErr)
			}
		}
		
		// Finance approved but escalated to director
		// Purchase stays DRAFT, but approval_status becomes PENDING for director review
		purchase.Status = models.PurchaseStatusDraft  // Keep as DRAFT
		purchase.ApprovalStatus = models.PurchaseApprovalPending  // Set to PENDING for director
		purchase.RequiresApproval = true  // Mark as requiring approval
		
		purchase.UpdatedAt = now
		_, err = s.purchaseRepo.Update(purchase)
		if err != nil {
			return nil, err
		}
		
		result["message"] = "Purchase approved by Finance and escalated to Director for final approval"
		result["purchase_id"] = purchaseID
		result["escalated"] = true
		result["status"] = "DRAFT"  // Status remains DRAFT
		result["approval_status"] = "PENDING"  // But approval status is PENDING
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

	if purchase.Status != models.PurchaseStatusApproved {
		return nil, errors.New("can only receive items for approved purchases")
	}

	// Generate receipt number
	receiptNumber, err := s.generateReceiptNumber()
	if err != nil {
		return nil, err
	}

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

	// Create receipt items
	for _, itemReq := range request.ReceiptItems {
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
	}

	// Update receipt status based on quantities
	err = s.updateReceiptStatus(createdReceipt.ID)
	if err != nil {
		return nil, err
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

// Private helper methods

func (s *PurchaseService) generatePurchaseCode() (string, error) {
	now := time.Now()
	year := now.Year()
	month := now.Month()
	
	// Use microsecond timestamp for uniqueness
	microseconds := now.UnixMicro()
	timestampSuffix := microseconds % 100000 // Get last 5 digits
	
	// Generate code with timestamp to ensure uniqueness
	code := fmt.Sprintf("PO-%04d-%02d-%05d", year, month, timestampSuffix)
	
	// Double-check uniqueness
	for attempt := 0; attempt < 5; attempt++ {
		exists, err := s.purchaseRepo.CodeExists(code)
		if err != nil {
			return "", err
		}
		
		if !exists {
			return code, nil
		}
		
		// If exists, add attempt number and retry
		code = fmt.Sprintf("PO-%04d-%02d-%05d%d", year, month, timestampSuffix, attempt+1)
	}
	
	return code, nil
}

func (s *PurchaseService) generateReceiptNumber() (string, error) {
	year := time.Now().Year()
	month := time.Now().Month()
	count, err := s.purchaseRepo.CountByMonth(year, int(month))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("RCP-%04d-%02d-%04d", year, month, count+1), nil
}

func (s *PurchaseService) checkIfApprovalRequired(amount float64) bool {
	// Check if there's an active workflow for this amount
	workflow, err := s.approvalService.GetWorkflowByAmount(models.ApprovalModulePurchase, amount)
	return err == nil && workflow != nil
}

// setApprovalBasisAndBase determines approval basis from env/config and sets base amount
func (s *PurchaseService) setApprovalBasisAndBase(p *models.Purchase) {
	basis := getApprovalBasis()
	p.ApprovalAmountBasis = basis
	switch basis {
	case "SUBTOTAL_BEFORE_DISCOUNT":
		p.ApprovalBaseAmount = p.SubtotalBeforeDiscount
	case "NET_AFTER_DISCOUNT_BEFORE_TAX":
		p.ApprovalBaseAmount = p.NetBeforeTax
	case "GRAND_TOTAL_AFTER_TAX":
		p.ApprovalBaseAmount = p.TotalAmount
	default:
		p.ApprovalBaseAmount = p.SubtotalBeforeDiscount
	}
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

func (s *PurchaseService) calculatePurchaseTotals(purchase *models.Purchase, items []models.PurchaseItemRequest) error {
	subtotalRaw := 0.0
	itemDiscountTotal := 0.0
	totalItemTax := 0.0

	// Clear existing items
	purchase.PurchaseItems = []models.PurchaseItem{}

	for _, itemReq := range items {
		// Validate product exists
		_, err := s.productRepo.FindByID(itemReq.ProductID)
		if err != nil {
			return fmt.Errorf("product %d not found", itemReq.ProductID)
		}

		// Get product details for auto expense account assignment
		product, err := s.productRepo.FindByID(itemReq.ProductID)
		if err != nil {
			return fmt.Errorf("product %d not found", itemReq.ProductID)
		}

		// Determine expense account ID with priority:
		// 1. Explicitly provided in request
		// 2. Product's default expense account
		// 3. Product category's default expense account
		// 4. Vendor's default expense account
		var expenseAccountID uint
		if itemReq.ExpenseAccountID != 0 {
			// Priority 1: Explicitly provided
			expenseAccountID = itemReq.ExpenseAccountID
		} else if product.DefaultExpenseAccountID != nil {
			// Priority 2: Product's default
			expenseAccountID = *product.DefaultExpenseAccountID
		} else if product.Category != nil && product.Category.DefaultExpenseAccountID != nil {
			// Priority 3: Product category's default
			expenseAccountID = *product.Category.DefaultExpenseAccountID
		} else {
			// Priority 4: Get vendor's default expense account
			vendor, err := s.contactRepo.GetByID(purchase.VendorID)
			if err == nil && vendor.DefaultExpenseAccountID != nil {
				expenseAccountID = *vendor.DefaultExpenseAccountID
			}
			// If still no account found, validation will catch this later
		}

		// Create purchase item
		item := models.PurchaseItem{
			ProductID:        itemReq.ProductID,
			Quantity:         itemReq.Quantity,
			UnitPrice:        itemReq.UnitPrice,
			Discount:         itemReq.Discount,
			Tax:              itemReq.Tax,
			ExpenseAccountID: expenseAccountID,
		}

		// Calculate line totals
		lineSubtotal := float64(item.Quantity) * item.UnitPrice
		discountAmount := lineSubtotal * (item.Discount / 100)
		item.TotalPrice = (lineSubtotal - discountAmount) // exclude tax here; tax handled separately

		subtotalRaw += lineSubtotal
		itemDiscountTotal += discountAmount
		totalItemTax += item.Tax

		purchase.PurchaseItems = append(purchase.PurchaseItems, item)
	}

	// Calculate purchase totals
	purchase.SubtotalBeforeDiscount = subtotalRaw
	purchase.ItemDiscountAmount = itemDiscountTotal
	// Subtotal after item-level discount
	subtotalAfterItem := subtotalRaw - itemDiscountTotal
	globalDiscountAmount := subtotalAfterItem * (purchase.Discount / 100)
	purchase.OrderDiscountAmount = globalDiscountAmount
	purchase.NetBeforeTax = subtotalAfterItem - globalDiscountAmount
	// Total tax = item taxes + order-level tax amount (TaxAmount treated as absolute)
	purchase.TaxAmount = totalItemTax + purchase.TaxAmount
	purchase.TotalAmount = purchase.NetBeforeTax + purchase.TaxAmount

	return nil
}

func (s *PurchaseService) recalculatePurchaseTotals(purchase *models.Purchase) error {
	subtotalRaw := 0.0
	itemDiscountTotal := 0.0
	totalItemTax := 0.0

	for i := range purchase.PurchaseItems {
		item := &purchase.PurchaseItems[i]

		// Calculate line totals
		lineSubtotal := float64(item.Quantity) * item.UnitPrice
		discountAmount := lineSubtotal * (item.Discount / 100)
		item.TotalPrice = (lineSubtotal - discountAmount)

		subtotalRaw += lineSubtotal
		itemDiscountTotal += discountAmount
		totalItemTax += item.Tax
	}

	// Calculate purchase totals
	purchase.SubtotalBeforeDiscount = subtotalRaw
	purchase.ItemDiscountAmount = itemDiscountTotal
	subtotalAfterItem := subtotalRaw - itemDiscountTotal
	globalDiscountAmount := subtotalAfterItem * (purchase.Discount / 100)
	purchase.OrderDiscountAmount = globalDiscountAmount
	purchase.NetBeforeTax = subtotalAfterItem - globalDiscountAmount
	purchase.TaxAmount = totalItemTax + purchase.TaxAmount
	purchase.TotalAmount = purchase.NetBeforeTax + purchase.TaxAmount

	return nil
}

func (s *PurchaseService) updatePurchaseItems(purchase *models.Purchase, items []models.PurchaseItemRequest) error {
	// Clear existing items and recreate
	purchase.PurchaseItems = []models.PurchaseItem{}

	for _, itemReq := range items {
		item := models.PurchaseItem{
			ProductID:        itemReq.ProductID,
			Quantity:         itemReq.Quantity,
			UnitPrice:        itemReq.UnitPrice,
			Discount:         itemReq.Discount,
			Tax:              itemReq.Tax,
			ExpenseAccountID: itemReq.ExpenseAccountID,
		}

		purchase.PurchaseItems = append(purchase.PurchaseItems, item)
	}

	return nil
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

	// Logic to determine if receipt is complete or partial
	// This is simplified - in reality you'd compare received vs ordered quantities
	receipt.Status = models.ReceiptStatusComplete

	_, err = s.purchaseRepo.Update(&models.Purchase{}) // Update receipt status
	return err
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
