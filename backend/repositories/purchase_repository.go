package repositories

import (
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
)

type PurchaseRepository struct {
	db *gorm.DB
}

func NewPurchaseRepository(db *gorm.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

// Purchase CRUD Operations

func (r *PurchaseRepository) Create(purchase *models.Purchase) (*models.Purchase, error) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(purchase).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return r.FindByID(purchase.ID)
}

func (r *PurchaseRepository) FindByID(id uint) (*models.Purchase, error) {
	var purchase models.Purchase
	err := r.db.Preload("Vendor").
		Preload("User").
		Preload("PurchaseItems.Product").
		Preload("PurchaseItems.ExpenseAccount").
		Preload("ApprovalRequest.Workflow").
		Preload("ApprovalRequest.Requester").
		Preload("ApprovalRequest.ApprovalSteps.Step").
		Preload("ApprovalRequest.ApprovalSteps.Approver").
		Preload("Approver").
		First(&purchase, id).Error

	if err != nil {
		return nil, err
	}
	return &purchase, nil
}

func (r *PurchaseRepository) FindAll() ([]models.Purchase, error) {
	var purchases []models.Purchase
	err := r.db.Preload("Vendor").
		Preload("User").
		Preload("PurchaseItems").
		Preload("ApprovalRequest").
		Find(&purchases).Error

	return purchases, err
}

func (r *PurchaseRepository) FindWithFilter(filter models.PurchaseFilter) ([]models.Purchase, int64, error) {
	var purchases []models.Purchase
	var total int64

	query := r.db.Model(&models.Purchase{}).
		Preload("Vendor").
		Preload("User").
		Preload("PurchaseItems.Product").
		// Load approval request with nested workflow and steps for role-based filtering on server side
		Preload("ApprovalRequest.Workflow").
		Preload("ApprovalRequest.ApprovalSteps.Step").
		Preload("ApprovalRequest.ApprovalSteps.Approver")

	// Apply filters
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.VendorID != "" {
		query = query.Where("vendor_id = ?", filter.VendorID)
	}

	if filter.StartDate != "" {
		query = query.Where("date >= ?", filter.StartDate)
	}

	if filter.EndDate != "" {
		query = query.Where("date <= ?", filter.EndDate)
	}

	if filter.Search != "" {
		query = query.Where("code ILIKE ? OR notes ILIKE ?", 
			"%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if filter.ApprovalStatus != "" {
		query = query.Where("approval_status = ?", filter.ApprovalStatus)
	}

	if filter.RequiresApproval != nil {
		query = query.Where("requires_approval = ?", *filter.RequiresApproval)
	}

	// Get total count
	query.Count(&total)

	// Apply pagination
	offset := (filter.Page - 1) * filter.Limit
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&purchases).Error

	return purchases, total, err
}

func (r *PurchaseRepository) Update(purchase *models.Purchase) (*models.Purchase, error) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Save(purchase).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return r.FindByID(purchase.ID)
}

func (r *PurchaseRepository) Delete(id uint) error {
	return r.db.Delete(&models.Purchase{}, id).Error
}

// Purchase Item Operations

func (r *PurchaseRepository) CreateItem(item *models.PurchaseItem) error {
	return r.db.Create(item).Error
}

func (r *PurchaseRepository) UpdateItem(item *models.PurchaseItem) error {
	return r.db.Save(item).Error
}

func (r *PurchaseRepository) DeleteItem(id uint) error {
	return r.db.Delete(&models.PurchaseItem{}, id).Error
}

func (r *PurchaseRepository) FindItemsByPurchaseID(purchaseID uint) ([]models.PurchaseItem, error) {
	var items []models.PurchaseItem
	err := r.db.Preload("Product").
		Preload("ExpenseAccount").
		Where("purchase_id = ?", purchaseID).
		Find(&items).Error

	return items, err
}

// Purchase Document Operations

func (r *PurchaseRepository) CreateDocument(document *models.PurchaseDocument) error {
	return r.db.Create(document).Error
}

func (r *PurchaseRepository) FindDocumentsByPurchaseID(purchaseID uint) ([]models.PurchaseDocument, error) {
	var documents []models.PurchaseDocument
	err := r.db.Where("purchase_id = ?", purchaseID).
		Find(&documents).Error

	return documents, err
}

func (r *PurchaseRepository) DeleteDocument(id uint) error {
	return r.db.Delete(&models.PurchaseDocument{}, id).Error
}

// Purchase Receipt Operations

func (r *PurchaseRepository) CreateReceipt(receipt *models.PurchaseReceipt) (*models.PurchaseReceipt, error) {
	if err := r.db.Create(receipt).Error; err != nil {
		return nil, err
	}

	return r.FindReceiptByID(receipt.ID)
}

func (r *PurchaseRepository) FindReceiptByID(id uint) (*models.PurchaseReceipt, error) {
	var receipt models.PurchaseReceipt
	err := r.db.Preload("Purchase").
		Preload("Receiver").
		Preload("ReceiptItems.PurchaseItem.Product").
		First(&receipt, id).Error

	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

func (r *PurchaseRepository) FindReceiptsByPurchaseID(purchaseID uint) ([]models.PurchaseReceipt, error) {
	var receipts []models.PurchaseReceipt
	err := r.db.Preload("User").
		Preload("ReceiptItems").
		Where("purchase_id = ?", purchaseID).
		Find(&receipts).Error

	return receipts, err
}

func (r *PurchaseRepository) CreateReceiptItem(item *models.PurchaseReceiptItem) error {
	return r.db.Create(item).Error
}

// Statistics and Analytics

func (r *PurchaseRepository) GetPurchasesSummary(startDate, endDate string) (*models.PurchaseSummary, error) {
	var summary models.PurchaseSummary

	query := r.db.Model(&models.Purchase{})

	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	// Get basic counts and totals
	var totalCount int64
	var totalAmount float64

	query.Count(&totalCount)
	// Only include non-rejected purchases in total amount calculation
	query.Where("status != ? AND approval_status != ?", models.PurchaseStatusCancelled, models.PurchaseApprovalRejected).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&totalAmount)

	// Get status counts
	var statusCounts []struct {
		Status string
		Count  int64
	}
	r.db.Model(&models.Purchase{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts)

	// Get approval status counts
	var approvalStatusCounts []struct {
		ApprovalStatus string
		Count          int64
	}
	r.db.Model(&models.Purchase{}).
		Select("approval_status, COUNT(*) as count").
		Group("approval_status").
		Find(&approvalStatusCounts)

	summary.TotalPurchases = totalCount
	summary.TotalAmount = totalAmount
	summary.StatusCounts = make(map[string]int64)
	summary.ApprovalStatusCounts = make(map[string]int64)

	for _, sc := range statusCounts {
		summary.StatusCounts[sc.Status] = sc.Count
	}

	for _, asc := range approvalStatusCounts {
		summary.ApprovalStatusCounts[asc.ApprovalStatus] = asc.Count
	}

	return &summary, nil
}

func (r *PurchaseRepository) GetVendorPurchaseSummary(vendorID uint) (*models.VendorPurchaseSummary, error) {
	var summary models.VendorPurchaseSummary

	err := r.db.Model(&models.Purchase{}).
		Where("vendor_id = ?", vendorID).
		Select("COUNT(*) as total_orders, COALESCE(SUM(total_amount), 0) as total_amount").
		Scan(&summary).Error

	return &summary, err
}

// Count operations for code generation

func (r *PurchaseRepository) CountByMonth(year, month int) (int64, error) {
	var count int64
	err := r.db.Model(&models.Purchase{}).
		Where("EXTRACT(year FROM created_at) = ? AND EXTRACT(month FROM created_at) = ?", 
			year, month).
		Count(&count).Error

	return count, err
}

func (r *PurchaseRepository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Purchase{}).
		Where("status = ?", status).
		Count(&count).Error

	return count, err
}

func (r *PurchaseRepository) CountPendingApproval() (int64, error) {
	var count int64
	err := r.db.Model(&models.Purchase{}).
		Where("approval_status = ?", models.PurchaseApprovalPending).
		Count(&count).Error

	return count, err
}

// Three-way matching operations

func (r *PurchaseRepository) GetPurchaseForMatching(id uint) (*models.PurchaseMatchingData, error) {
	var purchase models.Purchase
	err := r.db.Preload("PurchaseItems.Product").
		Preload("Vendor").
		First(&purchase, id).Error

	if err != nil {
		return nil, err
	}

	// Get receipts for this purchase
	var receipts []models.PurchaseReceipt
	err = r.db.Where("purchase_id = ?", id).
		Preload("ReceiptItems").
		Find(&receipts).Error
	if err != nil {
		return nil, err
	}

	// Get documents for this purchase
	var documents []models.PurchaseDocument
	err = r.db.Where("purchase_id = ?", id).Find(&documents).Error
	if err != nil {
		return nil, err
	}

	matching := &models.PurchaseMatchingData{
		Purchase:  purchase,
		Receipts:  receipts,
		Documents: documents,
	}

	return matching, nil
}

func (r *PurchaseRepository) UpdateMatchingStatus(purchaseID uint, status string) error {
	return r.db.Model(&models.Purchase{}).
		Where("id = ?", purchaseID).
		Update("matching_status", status).Error
}

// CodeExists checks if a purchase with the given code already exists
func (r *PurchaseRepository) CodeExists(code string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Purchase{}).Where("code = ?", code).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateJournal creates journal entry
func (r *PurchaseRepository) CreateJournal(journal *models.Journal) error {
	return r.db.Create(journal).Error
}

// GetPurchaseItemByID gets purchase item by ID
func (r *PurchaseRepository) GetPurchaseItemByID(id uint) (*models.PurchaseItem, error) {
	var item models.PurchaseItem
	err := r.db.Preload("Product").
		Preload("ExpenseAccount").
		First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateReceipt updates receipt
func (r *PurchaseRepository) UpdateReceipt(receipt *models.PurchaseReceipt) (*models.PurchaseReceipt, error) {
	err := r.db.Save(receipt).Error
	if err != nil {
		return nil, err
	}
	return r.FindReceiptByID(receipt.ID)
}

// CountReceiptsByMonth counts receipts by month
func (r *PurchaseRepository) CountReceiptsByMonth(year, month int) (int64, error) {
	var count int64
	err := r.db.Model(&models.PurchaseReceipt{}).
		Where("EXTRACT(year FROM received_date) = ? AND EXTRACT(month FROM received_date) = ?", 
			year, month).
		Count(&count).Error
	return count, err
}

// CountJournalsByMonth counts journals by month
func (r *PurchaseRepository) CountJournalsByMonth(year, month int) (int64, error) {
	var count int64
	err := r.db.Model(&models.Journal{}).
		Where("reference_type = ? AND EXTRACT(year FROM date) = ? AND EXTRACT(month FROM date) = ?", 
			models.JournalRefTypePurchase, year, month).
		Count(&count).Error
	return count, err
}

// GetPurchaseSummary gets purchase summary
func (r *PurchaseRepository) GetPurchaseSummary(startDate, endDate string) (*models.PurchaseSummary, error) {
	return r.GetPurchasesSummary(startDate, endDate)
}

// GetPayablesReport gets payables report  
func (r *PurchaseRepository) GetPayablesReport() (*models.PayablesReportResponse, error) {
	// This is a placeholder implementation - you may need to adjust based on your models
	var payables []models.PayablesReportData
	
	err := r.db.Model(&models.Purchase{}).
		Select(`
			purchases.id as purchase_id,
			purchases.code as purchase_code,
			contacts.name as vendor_name,
			purchases.date,
			purchases.due_date,
			purchases.total_amount,
			purchases.paid_amount,
			purchases.outstanding_amount,
			purchases.status
		`).
		Joins("JOIN contacts ON contacts.id = purchases.vendor_id").
		Where("purchases.outstanding_amount > 0 AND purchases.status IN (?)", 
			[]string{models.PurchaseStatusApproved, models.PurchaseStatusCompleted}).
		Order("purchases.due_date ASC").
		Scan(&payables).Error

	if err != nil {
		return nil, err
	}

	// Calculate totals
	var totalOutstanding float64
	for _, item := range payables {
		totalOutstanding += item.OutstandingAmount
	}

	return &models.PayablesReportResponse{
		TotalOutstanding: totalOutstanding,
		Payables:         payables,
	}, nil
}
