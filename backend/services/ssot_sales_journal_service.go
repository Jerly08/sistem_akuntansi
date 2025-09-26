package services

import (
	"fmt"
	"log"
	"app-sistem-akuntansi/models"
	"gorm.io/gorm"
	"github.com/shopspring/decimal"
)

// SSOTSalesJournalService handles sales journal entries using the SSOT system
type SSOTSalesJournalService struct {
	db                *gorm.DB
	unifiedJournal    *UnifiedJournalService
	accountResolver   *AccountResolver
}

func NewSSOTSalesJournalService(db *gorm.DB) *SSOTSalesJournalService {
	return &SSOTSalesJournalService{
		db:                db,
		unifiedJournal:    NewUnifiedJournalService(db),
		accountResolver:   NewAccountResolver(db),
	}
}

// CreateSaleJournalEntry creates a journal entry for a sale using SSOT
func (s *SSOTSalesJournalService) CreateSaleJournalEntry(sale *models.Sale, userID uint) (*models.SSOTJournalEntry, error) {
	log.Printf("📝 Creating SSOT journal entry for sale %d", sale.ID)
	
	// Resolve required accounts via AccountResolver (more robust than hard-coded codes)
	arAccount, err := s.accountResolver.GetAccount(AccountTypeAccountsReceivable)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve AR account: %v", err)
	}
	
	salesAccount, err := s.accountResolver.GetAccount(AccountTypeSalesRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve sales revenue account: %v", err)
	}
	
	// Create journal lines
	lines := []JournalLineRequest{
		{
			AccountID:    uint64(arAccount.ID),
			Description:  fmt.Sprintf("Sales to %s", sale.Customer.Name),
			DebitAmount:  decimal.NewFromFloat(sale.TotalAmount),
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    uint64(salesAccount.ID),
			Description:  "Sales Revenue",
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromFloat(sale.TotalAmount),
		},
	}
	
	// Add PPN line if applicable
	if sale.PPN > 0 {
		ppnAccount, err := s.accountResolver.GetAccount(AccountTypePPNPayable)
		if err == nil && ppnAccount != nil {
			lines = append(lines, JournalLineRequest{
				AccountID:    uint64(ppnAccount.ID),
				Description:  "PPN Keluaran",
				DebitAmount:  decimal.Zero,
				CreditAmount: decimal.NewFromFloat(sale.PPN),
			})
			// Adjust amounts for tax
			lines[0].DebitAmount = decimal.NewFromFloat(sale.TotalAmount) // AR = total incl. tax
			lines[1].CreditAmount = decimal.NewFromFloat(sale.TotalAmount - sale.PPN) // Sales = net
		}
	}
	
	// Create journal entry request
	journalRequest := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypeSale,
		SourceID:    func() *uint64 { id := uint64(sale.ID); return &id }(),
		Reference:   sale.Code,
		EntryDate:   sale.Date,
		Description: fmt.Sprintf("Sales Invoice %s - %s", sale.Code, sale.Customer.Name),
		CreatedBy:   uint64(userID),
		AutoPost:    true,  // ✅ Auto-post to update account balances
		Lines:       lines,
	}
	
	// Create using SSOT unified journal service
	journalResponse, err := s.unifiedJournal.CreateJournalEntry(journalRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSOT journal entry: %v", err)
	}
	
	// Retrieve the created SSOT journal entry
	var ssotEntry models.SSOTJournalEntry
	if err := s.db.First(&ssotEntry, journalResponse.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve created SSOT entry: %v", err)
	}
	
	log.Printf("✅ Created SSOT journal entry %d for sale %d", ssotEntry.ID, sale.ID)
	return &ssotEntry, nil
}

// CreatePaymentJournalEntry creates a journal entry for payment using SSOT
func (s *SSOTSalesJournalService) CreatePaymentJournalEntry(payment *models.SalePayment, userID uint) (*models.SSOTJournalEntry, error) {
	log.Printf("💰 Creating SSOT journal entry for payment %d", payment.ID)
	
	// Resolve required accounts
	arAccount, err := s.accountResolver.GetAccount(AccountTypeAccountsReceivable)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve AR account: %v", err)
	}
	
	cashAccount, err := s.accountResolver.GetBankAccountForPaymentMethod(payment.PaymentMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve cash/bank account: %v", err)
	}
	
	// Create journal entry request
	journalRequest := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypePayment,
		SourceID:    func() *uint64 { id := uint64(payment.ID); return &id }(),
		Reference:   fmt.Sprintf("PAY-%d", payment.ID),
		EntryDate:   payment.PaymentDate,
		Description: fmt.Sprintf("Payment for Sale %d - %s", payment.SaleID, payment.PaymentMethod),
		CreatedBy:   uint64(userID),
		AutoPost:    true,  // ✅ Auto-post to update account balances
		Lines: []JournalLineRequest{
			{
				AccountID:    uint64(cashAccount.ID),
				Description:  fmt.Sprintf("Payment received - %s", payment.PaymentMethod),
				DebitAmount:  decimal.NewFromFloat(payment.Amount),
				CreditAmount: decimal.Zero,
			},
			{
				AccountID:    uint64(arAccount.ID),
				Description:  "Payment against receivables",
				DebitAmount:  decimal.Zero,
				CreditAmount: decimal.NewFromFloat(payment.Amount),
			},
		},
	}
	
	// Create using SSOT unified journal service
	journalResponse, err := s.unifiedJournal.CreateJournalEntry(journalRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSOT payment journal entry: %v", err)
	}
	
	// Retrieve the created SSOT journal entry
	var ssotEntry models.SSOTJournalEntry
	if err := s.db.First(&ssotEntry, journalResponse.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve created SSOT payment entry: %v", err)
	}
	
	log.Printf("✅ Created SSOT payment journal entry %d for payment %d", ssotEntry.ID, payment.ID)
	return &ssotEntry, nil
}

// Helper methods
func (s *SSOTSalesJournalService) getAccountByCode(code string) (*models.Account, error) {
	var account models.Account
	if err := s.db.Where("code = ?", code).First(&account).Error; err != nil {
		return nil, fmt.Errorf("account %s not found: %v", code, err)
	}
	return &account, nil
}

func (s *SSOTSalesJournalService) getCashAccountForPayment(method string) (*models.Account, error) {
	// Map payment methods to account codes (legacy fallback)
	accountCodeMap := map[string]string{
		"CASH":        "1101", // Kas
		"BANK":        "1104", // Bank Mandiri
		"TRANSFER":    "1104", // Bank Mandiri
		"CREDIT_CARD": "1102", // Bank BCA
		"DEBIT_CARD":  "1102", // Bank BCA
	}
	
	code, exists := accountCodeMap[method]
	if !exists {
		code = "1104" // Default to Bank Mandiri
	}
	
	return s.getAccountByCode(code)
}
