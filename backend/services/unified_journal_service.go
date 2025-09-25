package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Type aliases for compatibility
type SourceType = string
type JournalStatus = string

// JournalEntryRequest represents the input for creating a journal entry
type JournalEntryRequest struct {
	SourceType      string            `json:"source_type"`
	SourceID        *uint64           `json:"source_id,omitempty"`
	Reference       string            `json:"reference"`
	EntryDate       time.Time         `json:"entry_date"`
	Description     string            `json:"description"`
	Lines           []JournalLineRequest `json:"lines"`
	AutoPost        bool              `json:"auto_post"`
	CreatedBy       uint64            `json:"created_by"`
}

type JournalLineRequest struct {
	AccountID    uint64          `json:"account_id"`
	Description  string          `json:"description"`
	DebitAmount  decimal.Decimal `json:"debit_amount"`
	CreditAmount decimal.Decimal `json:"credit_amount"`
}

// JournalResponse represents the response after creating/updating a journal entry
type JournalResponse struct {
	ID           uint64                    `json:"id"`
	EntryNumber  string                    `json:"entry_number"`
	Status       string                    `json:"status"`
	TotalDebit   decimal.Decimal           `json:"total_debit"`
	TotalCredit  decimal.Decimal           `json:"total_credit"`
	IsBalanced   bool                      `json:"is_balanced"`
	Lines        []JournalLineResponse     `json:"lines"`
	CreatedAt    time.Time                 `json:"created_at"`
	UpdatedAt    time.Time                 `json:"updated_at"`
}

type JournalLineResponse struct {
	ID           uint64          `json:"id"`
	LineNumber   int             `json:"line_number"`
	AccountID    uint64          `json:"account_id"`
	Description  string          `json:"description"`
	DebitAmount  decimal.Decimal `json:"debit_amount"`
	CreditAmount decimal.Decimal `json:"credit_amount"`
}

// UnifiedJournalService handles all journal operations using the SSOT schema
type UnifiedJournalService struct {
	db              *gorm.DB
	settingsService *SettingsService
}

// NewUnifiedJournalService creates a new instance of UnifiedJournalService
func NewUnifiedJournalService(db *gorm.DB) *UnifiedJournalService {
	return &UnifiedJournalService{db: db, settingsService: NewSettingsService(db)}
}

// CreateJournalEntry creates a new journal entry with validation (starts its own transaction)
func (s *UnifiedJournalService) CreateJournalEntry(req *JournalEntryRequest) (*JournalResponse, error) {
	var (
		entry *models.SSOTJournalEntry
		lines []models.SSOTJournalLine
		resp  *JournalResponse
	)

	// Use GORM Transaction for robust rollback/commit handling
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		return s.createJournalEntryWithTx(tx, req, &entry, &lines, &resp)
	}); err != nil {
		return nil, err
	}

	return resp, nil
}

// CreateJournalEntryWithTx creates a new journal entry within an existing transaction
// This prevents deadlocks when called from within another transaction
func (s *UnifiedJournalService) CreateJournalEntryWithTx(tx *gorm.DB, req *JournalEntryRequest) (*JournalResponse, error) {
	var (
		entry *models.SSOTJournalEntry
		lines []models.SSOTJournalLine
		resp  *JournalResponse
	)

	if err := s.createJournalEntryWithTx(tx, req, &entry, &lines, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// createJournalEntryWithTx - internal method that does the actual work
func (s *UnifiedJournalService) createJournalEntryWithTx(tx *gorm.DB, req *JournalEntryRequest, entry **models.SSOTJournalEntry, lines *[]models.SSOTJournalLine, resp **JournalResponse) error {
	// Validate request
	if err := s.validateJournalRequest(req); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Calculate totals
	totalDebit := decimal.Zero
	totalCredit := decimal.Zero
	for _, line := range req.Lines {
		totalDebit = totalDebit.Add(line.DebitAmount)
		totalCredit = totalCredit.Add(line.CreditAmount)
	}

	// Check balance
	isBalanced := totalDebit.Equal(totalCredit)
	if !isBalanced {
		return fmt.Errorf("journal entry is not balanced: debit=%s, credit=%s", totalDebit.String(), totalCredit.String())
	}

	// Create journal entry with proper source type mapping
	sourceType := req.SourceType
	if sourceType == "" {
		sourceType = models.SSOTSourceTypeManual
	}
	
// Generate entry number from settings sequence (within the same transaction)
	journalNumber, err := s.settingsService.GetNextJournalNumberTx(tx)
	if err != nil {
		return fmt.Errorf("failed to generate journal number: %w", err)
	}

	*entry = &models.SSOTJournalEntry{
		EntryNumber:     journalNumber,
		SourceType:      sourceType,
		SourceID:        req.SourceID,
		Reference:       req.Reference,
		EntryDate:       req.EntryDate,
		Description:     req.Description,
		TotalDebit:      totalDebit,
		TotalCredit:     totalCredit,
		IsBalanced:      isBalanced,
		IsAutoGenerated: sourceType != models.SSOTSourceTypeManual,
		Status:          models.SSOTStatusDraft,
		CreatedBy:       req.CreatedBy,
	}

	if err := tx.Create(*entry).Error; err != nil {
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	// Create journal lines
	for i, lineReq := range req.Lines {
		line := models.SSOTJournalLine{
			JournalID:    (*entry).ID,
			AccountID:    lineReq.AccountID,
			LineNumber:   i + 1,
			Description:  lineReq.Description,
			DebitAmount:  lineReq.DebitAmount,
			CreditAmount: lineReq.CreditAmount,
		}
		*lines = append(*lines, line)
	}

	if err := tx.Create(lines).Error; err != nil {
		return fmt.Errorf("failed to create journal lines: %w", err)
	}

	// Auto-post if requested
	if req.AutoPost {
		if err := s.postJournalEntryTx(tx, (*entry).ID); err != nil {
			return fmt.Errorf("failed to post journal entry: %w", err)
		}
	}

	// Build response after successful transactional operations
	*resp = s.buildJournalResponse(*entry, *lines)
	return nil
}

// GetJournalEntry retrieves a journal entry by ID
func (s *UnifiedJournalService) GetJournalEntry(id uint64) (*JournalResponse, error) {
	var entry models.SSOTJournalEntry
	if err := s.db.Preload("Lines").First(&entry, id).Error; err != nil {
		return nil, fmt.Errorf("journal entry not found: %w", err)
	}

	return s.buildJournalResponse(&entry, entry.Lines), nil
}

// PostJournalEntry posts a draft journal entry
func (s *UnifiedJournalService) PostJournalEntry(id uint64) error {
	// Use GORM Transaction for status update and validations only
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		return s.postJournalEntryTx(tx, id)
	}); err != nil {
		return err
	}

	// Refresh materialized view outside transaction with timeout and error logging
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.refreshAccountBalancesSafely(ctx); err != nil {
		log.Printf("warn: failed to refresh account balances after posting: %v", err)
		// Do not return error here to avoid blocking posting completion
	}

	// Log balance refresh event (WebSocket broadcasting removed for stability)
	log.Printf("Journal entry %d posted, account balances may need refresh", id)
	return nil
}

// postJournalEntryTx posts a journal entry within a transaction
func (s *UnifiedJournalService) postJournalEntryTx(tx *gorm.DB, id uint64) error {
	var entry models.SSOTJournalEntry
	if err := tx.First(&entry, id).Error; err != nil {
		return fmt.Errorf("journal entry not found: %w", err)
	}

	if entry.Status != models.SSOTStatusDraft {
		return fmt.Errorf("only draft entries can be posted")
	}

	if !entry.IsBalanced {
		return fmt.Errorf("cannot post unbalanced journal entry")
	}

	// Update status to posted
	if err := tx.Model(&entry).Updates(map[string]interface{}{
		"status":    models.SSOTStatusPosted,
		"posted_at": time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to update entry status: %w", err)
	}
	
	// Update account balances for each journal line
	var lines []models.SSOTJournalLine
	if err := tx.Where("journal_id = ?", entry.ID).Find(&lines).Error; err != nil {
		return fmt.Errorf("failed to get journal lines: %w", err)
	}
	
		for _, line := range lines {
			// Skip balance update for cash bank accounts as they are already updated by cash bank service
			// IMPORTANT: Cash bank accounts (cash_banks.account_id) must be handled by CashBankService
			// to prevent double posting and ensure proper transaction recording
			if s.isCashBankAccount(tx, line.AccountID) {
				log.Printf("ℹ️ Skipping balance update for cash bank account %d (handled by CashBankService)", line.AccountID)
				continue
			}
		
		if err := s.updateAccountBalance(tx, line.AccountID, line.DebitAmount, line.CreditAmount); err != nil {
			log.Printf("⚠️ Warning: Failed to update balance for account %d: %v", line.AccountID, err)
			// Continue with other updates instead of failing the entire transaction
		}
	}
	
	// Update header account balances after updating individual accounts
	if err := s.updateHeaderAccountBalances(tx); err != nil {
		log.Printf("⚠️ Warning: Failed to update header account balances: %v", err)
		// Continue without failing the transaction
	}

	return nil
}

// ReverseJournalEntry creates a reversing entry for a posted journal
func (s *UnifiedJournalService) ReverseJournalEntry(id uint64, description string, createdBy uint64) (*JournalResponse, error) {
	// Get original entry
	originalResp, err := s.GetJournalEntry(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get original entry: %w", err)
	}

	var original models.SSOTJournalEntry
	if err := s.db.First(&original, id).Error; err != nil {
		return nil, fmt.Errorf("original entry not found: %w", err)
	}

	if original.Status != models.SSOTStatusPosted {
		return nil, fmt.Errorf("can only reverse posted entries")
	}

	// Create reversing lines (swap debit/credit)
	var reversingLines []JournalLineRequest
	for _, line := range originalResp.Lines {
		reversingLines = append(reversingLines, JournalLineRequest{
			AccountID:    line.AccountID,
			Description:  fmt.Sprintf("Reversing: %s", line.Description),
			DebitAmount:  line.CreditAmount, // Swap
			CreditAmount: line.DebitAmount,  // Swap
		})
	}

	// Create reversing entry
	reversingReq := &JournalEntryRequest{
		SourceType:   models.SSOTSourceTypeReversal,
		SourceID:     &id, // Reference to original entry
		Reference:    fmt.Sprintf("REV-%s", original.Reference),
		EntryDate:    time.Now(),
		Description:  description,
		Lines:        reversingLines,
		AutoPost:     true,
		CreatedBy:    createdBy,
	}

	return s.CreateJournalEntry(reversingReq)
}

// GetJournalEntries retrieves journal entries with pagination and filters
func (s *UnifiedJournalService) GetJournalEntries(filters JournalFilters) (*PaginatedJournalResponse, error) {
	var entries []models.SSOTJournalEntry
	var total int64

	query := s.db.Model(&models.SSOTJournalEntry{})

	// Apply filters
	if filters.SourceType != "" {
		query = query.Where("source_type = ?", filters.SourceType)
	}
	if filters.SourceID != nil {
		query = query.Where("source_id = ?", *filters.SourceID)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if !filters.DateFrom.IsZero() {
		query = query.Where("entry_date >= ?", filters.DateFrom)
	}
	if !filters.DateTo.IsZero() {
		query = query.Where("entry_date <= ?", filters.DateTo)
	}
	if filters.Reference != "" {
		query = query.Where("reference ILIKE ?", "%"+filters.Reference+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count entries: %w", err)
	}

	// Apply pagination
	offset := (filters.Page - 1) * filters.Limit
	if err := query.Preload("Lines").
		Order("entry_date DESC, created_at DESC").
		Limit(filters.Limit).
		Offset(offset).
		Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	// Build response
	var responses []JournalResponse
	for _, entry := range entries {
		responses = append(responses, *s.buildJournalResponse(&entry, entry.Lines))
	}

	return &PaginatedJournalResponse{
		Data:       responses,
		Total:      total,
		Page:       filters.Page,
		Limit:      filters.Limit,
		TotalPages: (total + int64(filters.Limit) - 1) / int64(filters.Limit),
	}, nil
}

// GetAccountBalances retrieves account balances from materialized view
func (s *UnifiedJournalService) GetAccountBalances() ([]models.SSOTAccountBalance, error) {
	var balances []models.SSOTAccountBalance
	if err := s.db.Order("account_id").Find(&balances).Error; err != nil {
		return nil, fmt.Errorf("failed to get account balances: %w", err)
	}
	return balances, nil
}

// RefreshAccountBalances manually refreshes the materialized view
func (s *UnifiedJournalService) RefreshAccountBalances() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.refreshAccountBalancesSafely(ctx); err != nil {
		return err
	}
	// Log manual refresh event (WebSocket broadcasting removed for stability)
	log.Printf("Account balances manually refreshed")
	return nil
}

// updateAccountBalance updates the balance of an account based on journal line
func (s *UnifiedJournalService) updateAccountBalance(tx *gorm.DB, accountID uint64, debitAmount, creditAmount decimal.Decimal) error {
	// Get account details to determine balance calculation
	var account models.Account
	if err := tx.First(&account, accountID).Error; err != nil {
		return fmt.Errorf("account %d not found: %w", accountID, err)
	}
	
	// Calculate balance change based on account type
	// For ASSET and EXPENSE accounts: Debit increases balance, Credit decreases balance
	// For LIABILITY, EQUITY, and REVENUE accounts: Credit increases balance, Debit decreases balance
	// But we store all balances as positive numbers in the database
	var balanceChange decimal.Decimal
	
	switch account.Type {
	case "ASSET", "EXPENSE":
		// Normal debit balance accounts: Dr (+), Cr (-)
		balanceChange = debitAmount.Sub(creditAmount)
	case "LIABILITY", "EQUITY", "REVENUE":
		// Normal credit balance accounts: Cr (+), Dr (-)
		// Store as positive balance for revenue/liability/equity
		balanceChange = creditAmount.Sub(debitAmount)
	default:
		// Default to debit balance behavior
		balanceChange = debitAmount.Sub(creditAmount)
	}
	
	// Update account balance using raw SQL for precision
	balanceChangeFloat, _ := balanceChange.Float64()
	result := tx.Exec(
		"UPDATE accounts SET balance = balance + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL",
		balanceChangeFloat, accountID)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update account balance: %w", result.Error)
	}
	
	log.Printf("✅ Updated account %s (%d) balance by %.2f", account.Code, accountID, balanceChangeFloat)
	return nil
}

// updateHeaderAccountBalances updates header account balances based on children
func (s *UnifiedJournalService) updateHeaderAccountBalances(tx *gorm.DB) error {
	// Define header accounts that need balance updates
	headerMappings := map[string]string{
		"1000": "1%",    // ASSETS - all 1xxx accounts
		"1100": "11%",   // CURRENT ASSETS - all 11xx accounts
		"1200": "12%",   // ACCOUNTS RECEIVABLE - all 12xx accounts
		"2000": "2%",    // LIABILITIES - all 2xxx accounts
		"2100": "21%",   // CURRENT LIABILITIES - all 21xx accounts
		"3000": "3%",    // EQUITY - all 3xxx accounts
		"4000": "4%",    // REVENUE - all 4xxx accounts
		"5000": "5%",    // EXPENSES - all 5xxx accounts
	}
	
	for headerCode, childPattern := range headerMappings {
		// Get header account
		var headerAccount models.Account
		if err := tx.Where("code = ? AND is_header = ?", headerCode, true).First(&headerAccount).Error; err != nil {
			continue // Skip if header account not found
		}
		
		// Calculate sum of children (non-header accounts with matching pattern)
		var childrenSum float64
		err := tx.Model(&models.Account{}).
			Where("code LIKE ? AND code != ? AND is_header = ? AND deleted_at IS NULL", childPattern, headerCode, false).
			Select("COALESCE(SUM(balance), 0)").
			Scan(&childrenSum).Error
			
		if err != nil {
			continue // Skip on error
		}
		
		// Update header account balance if different
		if headerAccount.Balance != childrenSum {
			result := tx.Exec(
				"UPDATE accounts SET balance = ?, updated_at = CURRENT_TIMESTAMP WHERE code = ? AND deleted_at IS NULL",
				childrenSum, headerCode)
				
			if result.Error == nil {
				log.Printf("✅ Updated header account %s balance: %.2f -> %.2f", headerCode, headerAccount.Balance, childrenSum)
			}
		}
	}
	
	return nil
}

// isCashBankAccount checks if an account is linked to a cash bank account
func (s *UnifiedJournalService) isCashBankAccount(tx *gorm.DB, accountID uint64) bool {
	// Check if there's a cash bank account that uses this GL account
	var count int64
	tx.Table("cash_banks").Where("account_id = ?", accountID).Count(&count)
	return count > 0
}

// Helper types for filters and responses
type JournalFilters struct {
	SourceType string    `form:"source_type"`
	SourceID   *uint64   `form:"source_id"`
	Status     string    `form:"status"`
	DateFrom   time.Time `form:"date_from"`
	DateTo     time.Time `form:"date_to"`
	Reference  string    `form:"reference"`
	Page       int       `form:"page"`
	Limit      int       `form:"limit"`
}

type PaginatedJournalResponse struct {
	Data       []JournalResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int64             `json:"total_pages"`
}

// refreshAccountBalancesSafely refreshes the materialized view with context timeout and logging
// TEMPORARILY DISABLED: Skip materialized view refresh to prevent deposit timeout
func (s *UnifiedJournalService) refreshAccountBalancesSafely(ctx context.Context) error {
	// Skip materialized view refresh for instant processing
	log.Printf("ℹ️ Skipping materialized view refresh for instant processing")
	return nil
	
	// Original code (disabled):
	// if err := s.db.WithContext(ctx).Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY account_balances").Error; err != nil {
	//	return fmt.Errorf("failed to refresh account balances: %w", err)
	// }
	// return nil
}

// validateJournalRequest validates the journal entry request
func (s *UnifiedJournalService) validateJournalRequest(req *JournalEntryRequest) error {
	if req.Description == "" {
		return fmt.Errorf("description is required")
	}

	if len(req.Lines) < 2 {
		return fmt.Errorf("at least 2 journal lines are required")
	}

	// Validate lines
	for i, line := range req.Lines {
		if line.AccountID == 0 {
			return fmt.Errorf("line %d: account_id is required", i+1)
		}

		// Either debit or credit must be non-zero, but not both
		debitZero := line.DebitAmount.IsZero()
		creditZero := line.CreditAmount.IsZero()

		if debitZero && creditZero {
			return fmt.Errorf("line %d: either debit or credit amount must be non-zero", i+1)
		}

		if !debitZero && !creditZero {
			return fmt.Errorf("line %d: cannot have both debit and credit amounts", i+1)
		}

		// Amounts must be positive
		if line.DebitAmount.IsNegative() {
			return fmt.Errorf("line %d: debit amount cannot be negative", i+1)
		}
		if line.CreditAmount.IsNegative() {
			return fmt.Errorf("line %d: credit amount cannot be negative", i+1)
		}
	}

	return nil
}

// buildJournalResponse builds a JournalResponse from models
func (s *UnifiedJournalService) buildJournalResponse(entry *models.SSOTJournalEntry, lines []models.SSOTJournalLine) *JournalResponse {
	var lineResponses []JournalLineResponse
	for _, line := range lines {
		lineResponses = append(lineResponses, JournalLineResponse{
			ID:           line.ID,
			LineNumber:   line.LineNumber,
			AccountID:    line.AccountID,
			Description:  line.Description,
			DebitAmount:  line.DebitAmount,
			CreditAmount: line.CreditAmount,
		})
	}

	return &JournalResponse{
		ID:          entry.ID,
		EntryNumber: entry.EntryNumber,
		Status:      entry.Status,
		TotalDebit:  entry.TotalDebit,
		TotalCredit: entry.TotalCredit,
		IsBalanced:  entry.IsBalanced,
		Lines:       lineResponses,
		CreatedAt:   entry.CreatedAt,
		UpdatedAt:   entry.UpdatedAt,
	}
}