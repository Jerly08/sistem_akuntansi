package services

import (
	"fmt"
	"log"
	"strings"
	"time"
	"app-sistem-akuntansi/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SalesJournalServiceSSOT handles sales journal entries with CORRECT unified_journal_ledger integration
// This service writes to unified_journal_ledger which is read by Balance Sheet service
type SalesJournalServiceSSOT struct {
	db         *gorm.DB
	coaService *COAService
}

// NewSalesJournalServiceSSOT creates a new instance
func NewSalesJournalServiceSSOT(db *gorm.DB, coaService *COAService) *SalesJournalServiceSSOT {
	return &SalesJournalServiceSSOT{
		db:         db,
		coaService: coaService,
	}
}

// ShouldPostToJournal checks if a status should create journal entries
func (s *SalesJournalServiceSSOT) ShouldPostToJournal(status string) bool {
	allowedStatuses := []string{"INVOICED", "PAID"}
	for _, allowed := range allowedStatuses {
		if status == allowed {
			return true
		}
	}
	return false
}

// syncCashBankBalance syncs cash_banks.balance with linked accounts.balance
// This ensures Cash & Bank Management page always shows same balance as COA
func (s *SalesJournalServiceSSOT) syncCashBankBalance(tx *gorm.DB, accountID uint64) error {
	// Find if this account is linked to a cash_bank record
	var cashBank models.CashBank
	if err := tx.Where("account_id = ?", accountID).First(&cashBank).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Not a cash/bank account, skip sync (normal for revenue, expense accounts)
			return nil
		}
		return err
	}
	
	// Get current COA account balance
	var account models.Account
	if err := tx.First(&account, accountID).Error; err != nil {
		return err
	}
	
	// Check if sync needed
	if cashBank.Balance == account.Balance {
		// Already in sync, no update needed
		return nil
	}
	
	// Sync balance
	oldBalance := cashBank.Balance
	cashBank.Balance = account.Balance
	
	if err := tx.Save(&cashBank).Error; err != nil {
		return err
	}
	
	log.Printf("🔄 [SYNC] CashBank #%d '%s' synced: %.2f → %.2f (from COA account #%d)", 
		cashBank.ID, cashBank.Name, oldBalance, cashBank.Balance, accountID)
	
	return nil
}

// CreateSalesJournal creates journal entries in unified_journal_ledger for Balance Sheet integration
func (s *SalesJournalServiceSSOT) CreateSalesJournal(sale *models.Sale, tx *gorm.DB) error {
	// VALIDASI STATUS - HANYA INVOICED/PAID YANG BOLEH POSTING
	if !s.ShouldPostToJournal(sale.Status) {
		log.Printf("⚠️ [SSOT] Skipping journal creation for Sale #%d with status: %s (only INVOICED/PAID allowed)", sale.ID, sale.Status)
		return nil
	}

	log.Printf("📝 [SSOT] Creating unified journal entries for Sale #%d (Status: %s, Payment Method: '%s')", 
		sale.ID, sale.Status, sale.PaymentMethodType)
	
	// ✅ FIX: Don't fail if payment method type is empty, use default CREDIT
	// This prevents blocking journal creation for valid sales
	if strings.TrimSpace(sale.PaymentMethodType) == "" {
		log.Printf("⚠️ [SSOT] Warning: Sale #%d has empty PaymentMethodType, defaulting to CREDIT", sale.ID)
		// Don't return error, allow journal creation with default
	}

	// Tentukan database yang akan digunakan
	dbToUse := s.db
	if tx != nil {
		dbToUse = tx
	}

	// Check if journal already exists
	var existingCount int64
	if err := dbToUse.Model(&models.SSOTJournalEntry{}).
		Where("source_type = ? AND source_id = ?", "SALE", sale.ID).
		Count(&existingCount).Error; err == nil && existingCount > 0 {
		log.Printf("⚠️ [SSOT] Journal already exists for Sale #%d (found %d entries), skipping", 
			sale.ID, existingCount)
		return nil
	}

	// Helper to resolve account by code
	resolveByCode := func(code string) (*models.Account, error) {
		var acc models.Account
		if err := dbToUse.Where("code = ?", code).First(&acc).Error; err != nil {
			return nil, fmt.Errorf("account code %s not found: %v", code, err)
		}
		return &acc, nil
	}

	// Prepare journal lines
	var lines []SalesJournalLineRequest

	// 1. DEBIT side - based on payment method
	var debitAccount *models.Account
	var err error
	
	switch strings.ToUpper(strings.TrimSpace(sale.PaymentMethodType)) {
	case "TUNAI", "CASH":
		debitAccount, err = resolveByCode("1101")
		if err != nil {
			return fmt.Errorf("cash account not found: %v", err)
		}
	case "TRANSFER", "BANK":
		// Try to get from cash/bank account if set
		if sale.CashBankID != nil && *sale.CashBankID > 0 {
			if err := dbToUse.First(&debitAccount, *sale.CashBankID).Error; err != nil {
				log.Printf("⚠️ Cash/Bank account ID %d not found, using default BANK account", *sale.CashBankID)
				debitAccount, err = resolveByCode("1102")
				if err != nil {
					return fmt.Errorf("bank account not found: %v", err)
				}
			}
		} else {
			debitAccount, err = resolveByCode("1102")
			if err != nil {
				return fmt.Errorf("bank account not found: %v", err)
			}
		}
	case "CREDIT", "PIUTANG":
		debitAccount, err = resolveByCode("1201")
		if err != nil {
			return fmt.Errorf("receivables account not found: %v", err)
		}
	default:
		// ✅ FIX: Use CREDIT as safe default for unknown payment methods
		// This allows journal creation even if payment method type doesn't match
		// Better to record the transaction than to fail completely
		log.Printf("⚠️ [SSOT] Warning: Unknown payment method type '%s' for Sale #%d, defaulting to CREDIT (Piutang)", 
			sale.PaymentMethodType, sale.ID)
		debitAccount, err = resolveByCode("1201")
		if err != nil {
			return fmt.Errorf("receivables account not found: %v", err)
		}
	}

	// Add DEBIT line
	lines = append(lines, SalesJournalLineRequest{
		AccountID:    uint64(debitAccount.ID),
		DebitAmount:  decimal.NewFromFloat(sale.TotalAmount),
		CreditAmount: decimal.Zero,
		Description:  fmt.Sprintf("Penjualan - %s", sale.InvoiceNumber),
	})

	// 2. CREDIT side - Revenue
	revenueAccount, err := resolveByCode("4101")
	if err != nil {
		return fmt.Errorf("revenue account not found: %v", err)
	}

	lines = append(lines, SalesJournalLineRequest{
		AccountID:    uint64(revenueAccount.ID),
		DebitAmount:  decimal.Zero,
		CreditAmount: decimal.NewFromFloat(sale.Subtotal),
		Description:  fmt.Sprintf("Pendapatan Penjualan - %s", sale.InvoiceNumber),
	})

	// 3. PPN if exists
	if sale.PPN > 0 {
		ppnAccount, err := resolveByCode("2103")
		if err != nil {
			log.Printf("⚠️ PPN account not found, skipping PPN entry: %v", err)
		} else {
			lines = append(lines, SalesJournalLineRequest{
				AccountID:    uint64(ppnAccount.ID),
				DebitAmount:  decimal.Zero,
				CreditAmount: decimal.NewFromFloat(sale.PPN),
				Description:  fmt.Sprintf("PPN Keluaran - %s", sale.InvoiceNumber),
			})
		}
	}

	// 4. PPh if exists (liability - reduces receivable/cash on debit side)
	// ✅ FIX: PPh is a LIABILITY (utang pajak) so it should be CREDITED
	// The receivable/cash on debit side will be NET of PPh (Total - PPh)
	if sale.PPh > 0 {
		pphAccount, err := resolveByCode("2104")
		if err != nil {
			log.Printf("⚠️ PPh account not found, skipping PPh entry: %v", err)
		} else {
			lines = append(lines, SalesJournalLineRequest{
				AccountID:    uint64(pphAccount.ID),
				DebitAmount:  decimal.Zero,
				CreditAmount: decimal.NewFromFloat(sale.PPh),
				Description:  fmt.Sprintf("PPh Dipotong - %s", sale.InvoiceNumber),
			})
			log.Printf("💰 [PPh] Recorded PPh liability: Rp %.2f", sale.PPh)
		}
	}

	// ========================================
	// 🔥 FIX CRITICAL: ADD COGS JOURNAL ENTRY
	// ========================================
	// 5. COGS Recording - Cost of Goods Sold
	// Load sale items with products to calculate COGS
	var saleWithItems models.Sale
	if err := dbToUse.Preload("SaleItems.Product").First(&saleWithItems, sale.ID).Error; err != nil {
		log.Printf("⚠️ Failed to load sale items for COGS calculation: %v", err)
	} else {
		var totalCOGS decimal.Decimal
		var cogsDetails []string
		
		// Calculate total COGS from all sale items
		for _, item := range saleWithItems.SaleItems {
			// Check if product is loaded
			if item.Product.ID == 0 {
				log.Printf("⚠️ Sale item #%d has no product loaded, skipping COGS", item.ID)
				continue
			}
			
			// Calculate COGS: Quantity × Cost Price
			itemCOGS := decimal.NewFromFloat(float64(item.Quantity)).
				Mul(decimal.NewFromFloat(item.Product.CostPrice))
			
			if itemCOGS.IsZero() {
				log.Printf("⚠️ Product '%s' (ID: %d) has zero cost price, COGS = 0", 
					item.Product.Name, item.Product.ID)
			} else {
				totalCOGS = totalCOGS.Add(itemCOGS)
				cogsDetails = append(cogsDetails, fmt.Sprintf("%s(Qty:%d×Rp%.0f)", 
					item.Product.Name, item.Quantity, item.Product.CostPrice))
			}
		}
		
		// Only create COGS entry if total COGS > 0
		if !totalCOGS.IsZero() {
			// DEBIT: 5101 - Harga Pokok Penjualan (COGS Expense)
			cogsAccount, err := resolveByCode("5101")
			if err != nil {
				log.Printf("⚠️ COGS account (5101) not found, skipping COGS entry: %v", err)
			} else {
				lines = append(lines, SalesJournalLineRequest{
					AccountID:    uint64(cogsAccount.ID),
					DebitAmount:  totalCOGS,
					CreditAmount: decimal.Zero,
					Description:  fmt.Sprintf("HPP - %s", sale.InvoiceNumber),
				})
				
				// CREDIT: 1301 - Persediaan Barang (Inventory)
				inventoryAccount, err := resolveByCode("1301")
				if err != nil {
					log.Printf("⚠️ Inventory account (1301) not found, skipping inventory credit: %v", err)
				} else {
					lines = append(lines, SalesJournalLineRequest{
						AccountID:    uint64(inventoryAccount.ID),
						DebitAmount:  decimal.Zero,
						CreditAmount: totalCOGS,
						Description:  fmt.Sprintf("Pengurangan Persediaan - %s", sale.InvoiceNumber),
					})
					
					log.Printf("💰 [COGS] Calculated COGS for Sale #%d: Rp %.2f (%d items: %s)", 
						sale.ID, totalCOGS.InexactFloat64(), len(cogsDetails), 
						strings.Join(cogsDetails, ", "))
				}
			}
		} else {
			log.Printf("⚠️ [COGS] No COGS calculated for Sale #%d (all items have zero cost price)", sale.ID)
		}
	}
	// ========================================
	// END COGS FIX
	// ========================================

	// Calculate totals
	var totalDebit, totalCredit decimal.Decimal
	for _, line := range lines {
		totalDebit = totalDebit.Add(line.DebitAmount)
		totalCredit = totalCredit.Add(line.CreditAmount)
	}

	// Verify balanced
	if !totalDebit.Equal(totalCredit) {
		return fmt.Errorf("journal entry not balanced: debit=%.2f, credit=%.2f", 
			totalDebit.InexactFloat64(), totalCredit.InexactFloat64())
	}

	// Create journal entry
	sourceID := uint64(sale.ID)
	now := time.Now()
	postedBy := uint64(sale.UserID)
	
	journalEntry := &models.SSOTJournalEntry{
		EntryNumber:     fmt.Sprintf("SALE-%d-%d", sale.ID, now.Unix()),
		SourceType:      "SALE",
		SourceID:        &sourceID,
		SourceCode:      sale.InvoiceNumber,
		EntryDate:       sale.Date,
		Description:     fmt.Sprintf("Sales Invoice #%s - %s", sale.InvoiceNumber, sale.Customer.Name),
		Reference:       sale.InvoiceNumber,
		TotalDebit:      totalDebit,
		TotalCredit:     totalCredit,
		Status:          "POSTED",
		IsBalanced:      true,
		IsAutoGenerated: true,
		PostedAt:        &now,
		PostedBy:        &postedBy,
		CreatedBy:       uint64(sale.UserID),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := dbToUse.Create(journalEntry).Error; err != nil {
		return fmt.Errorf("failed to create SSOT journal entry: %v", err)
	}

	// Create journal lines
	for i, lineReq := range lines {
		journalLine := &models.SSOTJournalLine{
			JournalID:    journalEntry.ID,
			AccountID:    lineReq.AccountID,
			LineNumber:   i + 1,
			Description:  lineReq.Description,
			DebitAmount:  lineReq.DebitAmount,
			CreditAmount: lineReq.CreditAmount,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := dbToUse.Create(journalLine).Error; err != nil {
			return fmt.Errorf("failed to create SSOT journal line: %v", err)
		}

		// ✅ RE-ENABLED: Update account balance for COA tree view
		// P&L uses journal entries (correct), but COA Tree View uses account.balance field
		if err := s.updateAccountBalance(dbToUse, lineReq.AccountID, lineReq.DebitAmount, lineReq.CreditAmount); err != nil {
			log.Printf("⚠️ Warning: Failed to update account balance for account %d: %v", lineReq.AccountID, err)
			// Continue processing - don't fail transaction for balance update issues
		}
		
		// ✅ CRITICAL: Sync cash_banks.balance with accounts.balance after COA update
		// This ensures Cash & Bank Management page shows same balance as COA
		if err := s.syncCashBankBalance(dbToUse, lineReq.AccountID); err != nil {
			log.Printf("⚠️ Warning: Failed to sync cash/bank balance for account %d: %v", lineReq.AccountID, err)
			// Continue processing - don't fail transaction for sync issues
		}
	}

	log.Printf("✅ [SSOT] Created journal entry #%d with %d lines (Debit: %.2f, Credit: %.2f)", 
		journalEntry.ID, len(lines), totalDebit.InexactFloat64(), totalCredit.InexactFloat64())

	return nil
}

// updateAccountBalance updates account.balance field for COA tree view display
// RE-ENABLED: COA Tree View needs this field updated
// Note: P&L Report calculates balance from journal entries (real-time, always correct)
//       COA Tree View reads from account.balance field (needs manual update)
func (s *SalesJournalServiceSSOT) updateAccountBalance(db *gorm.DB, accountID uint64, debitAmount, creditAmount decimal.Decimal) error {
	var account models.Account
	if err := db.First(&account, accountID).Error; err != nil {
		return fmt.Errorf("account %d not found: %v", accountID, err)
	}

	// Calculate net change: debit - credit
	debit := debitAmount.InexactFloat64()
	credit := creditAmount.InexactFloat64()
	netChange := debit - credit
	
	oldBalance := account.Balance

	// ✅ VALIDATION: Check account type correctness for critical accounts
	accountType := strings.ToUpper(account.Type)
	if account.Code == "1301" && accountType != "ASSET" {
		log.Printf("❌ [BUG] Account 1301 (Persediaan) has WRONG type '%s', should be 'ASSET'!", accountType)
		log.Printf("❌ This will cause INCORRECT balance calculation!")
	}
	if account.Code == "5101" && accountType != "EXPENSE" {
		log.Printf("❌ [BUG] Account 5101 (COGS) has WRONG type '%s', should be 'EXPENSE'!", accountType)
	}

	// Update balance based on account type
	switch accountType {
	case "ASSET", "EXPENSE":
		// Assets and Expenses: debit increases balance
		account.Balance += netChange
		log.Printf("📊 [SSOT] Account %s (%s) TYPE=%s: Balance %.2f + netChange(%.2f) = %.2f", 
			account.Code, account.Name, accountType, oldBalance, netChange, account.Balance)
	case "LIABILITY", "EQUITY", "REVENUE":
		// Liabilities, Equity, Revenue: credit increases balance (so debit decreases)
		account.Balance -= netChange
		log.Printf("📊 [SSOT] Account %s (%s) TYPE=%s: Balance %.2f - netChange(%.2f) = %.2f", 
			account.Code, account.Name, accountType, oldBalance, netChange, account.Balance)
	default:
		log.Printf("⚠️ [SSOT] Unknown account type '%s' for account %s (%s)", accountType, account.Code, account.Name)
		// Fallback: treat as ASSET
		account.Balance += netChange
	}

	if err := db.Save(&account).Error; err != nil {
		return fmt.Errorf("failed to save account balance: %v", err)
	}

	// ✅ DETAILED LOGGING for debugging
	balanceChange := account.Balance - oldBalance
	if account.Code == "1301" {
		if credit > 0 && balanceChange > 0 {
			log.Printf("❌ [BUG DETECTED] Account 1301 CREDIT %.2f but balance INCREASED by %.2f! (Should DECREASE)", 
				credit, balanceChange)
			log.Printf("❌ Old Balance: %.2f, New Balance: %.2f, Type: %s, netChange: %.2f", 
				oldBalance, account.Balance, accountType, netChange)
		} else if credit > 0 && balanceChange < 0 {
			log.Printf("✅ [CORRECT] Account 1301 CREDIT %.2f and balance DECREASED by %.2f (correct!)", 
				credit, -balanceChange)
		}
	}

	log.Printf("💰 [SSOT] Updated account %s (%s) TYPE=%s balance: Dr=%.2f, Cr=%.2f, netChange=%.2f, Old=%.2f, New=%.2f", 
		account.Code, account.Name, accountType, debit, credit, netChange, oldBalance, account.Balance)

	return nil
}

// UpdateSalesJournal updates journal entries based on status change
func (s *SalesJournalServiceSSOT) UpdateSalesJournal(sale *models.Sale, oldStatus string, tx *gorm.DB) error {
	dbToUse := s.db
	if tx != nil {
		dbToUse = tx
	}

	oldShouldPost := s.ShouldPostToJournal(oldStatus)
	newShouldPost := s.ShouldPostToJournal(sale.Status)

	if !oldShouldPost && newShouldPost {
		// Create journal
		log.Printf("📈 [SSOT] Status changed from %s to %s - Creating journal entries", oldStatus, sale.Status)
		return s.CreateSalesJournal(sale, dbToUse)
	} else if oldShouldPost && !newShouldPost {
		// Delete journal
		log.Printf("📉 [SSOT] Status changed from %s to %s - Removing journal entries", oldStatus, sale.Status)
		return s.DeleteSalesJournal(sale.ID, dbToUse)
	} else if oldShouldPost && newShouldPost {
		// Update existing
		log.Printf("🔄 [SSOT] Updating journal entries for Sale #%d", sale.ID)
		
		if err := s.DeleteSalesJournal(sale.ID, dbToUse); err != nil {
			return err
		}
		
		return s.CreateSalesJournal(sale, dbToUse)
	}

	log.Printf("ℹ️ [SSOT] No journal update needed for Sale #%d (Status: %s)", sale.ID, sale.Status)
	return nil
}

// DeleteSalesJournal deletes all journal entries for a sale
func (s *SalesJournalServiceSSOT) DeleteSalesJournal(saleID uint, tx *gorm.DB) error {
	dbToUse := s.db
	if tx != nil {
		dbToUse = tx
	}

	// Find journal entry
	var entry models.SSOTJournalEntry
	if err := dbToUse.Where("source_type = ? AND source_id = ?", "SALE", saleID).First(&entry).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("⚠️ [SSOT] No journal found for Sale #%d, nothing to delete", saleID)
			return nil
		}
		return fmt.Errorf("failed to find journal entry: %v", err)
	}

	// Delete lines first (FK constraint)
	if err := dbToUse.Where("journal_id = ?", entry.ID).Delete(&models.SSOTJournalLine{}).Error; err != nil {
		return fmt.Errorf("failed to delete journal lines: %v", err)
	}

	// Delete entry
	if err := dbToUse.Delete(&entry).Error; err != nil {
		return fmt.Errorf("failed to delete journal entry: %v", err)
	}

	log.Printf("✅ [SSOT] Deleted journal entry #%d and its lines for Sale #%d", entry.ID, saleID)
	return nil
}

// SalesJournalLineRequest represents a request to create a sales journal line
type SalesJournalLineRequest struct {
	AccountID    uint64
	DebitAmount  decimal.Decimal
	CreditAmount decimal.Decimal
	Description  string
}

