# Sales Invoice Chart of Accounts Fix Guide

## Problem Summary

Your current system has incorrect Chart of Accounts balances:
1. **Piutang Usaha (Account Receivable)** shows negative balance (-Rp 3,330,000) instead of positive debit balance
2. **Pendapatan Penjualan (Sales Revenue)** shows zero balance instead of negative credit balance reflecting sales
3. **Bank accounts** are updated but not properly coordinated with sales invoicing

## Root Cause Analysis

### 1. Incorrect Journal Entry Logic
Your existing `SSOTSalesJournalService` may have incorrect debit/credit logic or balance calculation.

### 2. Double Journal Entry Creation
Both `SalesService` and `PaymentService` may be creating journal entries, causing double-posting.

### 3. Wrong Balance Update Logic
The account balance updates don't follow normal balance rules (Assets/Expenses increase with debits, Liabilities/Equity/Revenue increase with credits).

## Solution Implementation

### Step 1: Run Diagnostic Tool

First, run the diagnostic tool to understand current issues:

```bash
# Update database connection in the file first
cd backend/tools
go run diagnose_coa_balances.go
```

### Step 2: Update Your Sales Service

Replace your journal entry creation in `sales_service.go`:

```go
func (s *SalesService) createJournalEntriesForSale(sale *models.Sale, userID uint) error {
	// Use the corrected SSOT Sales Journal Service
	correctedService := services.NewCorrectedSSOTSalesJournalService(s.db)
	
	// Create the journal entry using corrected service
	_, err := correctedService.CreateSaleJournalEntry(sale, userID)
	if err != nil {
		log.Printf("❌ Failed to create corrected journal entry for sale %d: %v", sale.ID, err)
		return fmt.Errorf("failed to create journal entries: %v", err)
	}
	
	log.Printf("✅ Successfully created corrected journal entries for sale %d", sale.ID)
	return nil
}
```

### Step 3: Update Payment Processing

Update your payment service to use the corrected logic:

```go
func (s *PaymentService) CreateSalePayment(saleID uint, request models.SalePaymentRequest, userID uint) (*models.SalePayment, error) {
	// ... existing payment creation logic ...
	
	// Use corrected SSOT service for journal entries
	correctedService := services.NewCorrectedSSOTSalesJournalService(s.db)
	_, err = correctedService.CreatePaymentJournalEntry(createdPayment, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment journal entries: %v", err)
	}
	
	return createdPayment, nil
}
```

### Step 4: Fix Existing Account Balances

Run the balance correction script:

```bash
# Update database connection in the file first
cd backend/tools
go run correct_account_balances.go
```

### Step 5: Verify Account Mappings

Ensure your Chart of Accounts has these account codes:

| Account Code | Account Name | Type | Expected Balance |
|--------------|--------------|------|------------------|
| 1201 | Piutang Usaha | ASSET | Positive (Debit) |
| 4101 | Pendapatan Penjualan | REVENUE | Negative (Credit) |
| 1102 | Bank BCA | ASSET | Positive (Debit) |
| 1104 | Bank Mandiri | ASSET | Positive (Debit) |
| 2103 | PPN Keluaran | LIABILITY | Negative (Credit) |

## Correct Accounting Logic

### When Sale is INVOICED:
```
Dr. Piutang Usaha (1201)         3,330,000
    Cr. Pendapatan Penjualan (4101)    3,000,000
    Cr. PPN Keluaran (2103)               330,000
```

### When Payment is RECEIVED:
```
Dr. Bank UOB (1103)              3,330,000
    Cr. Piutang Usaha (1201)           3,330,000
```

## Expected Account Balances After Correction

| Account | Before Fix | After Fix | Explanation |
|---------|------------|-----------|-------------|
| Piutang Usaha (1201) | -3,330,000 | +3,330,000 | Asset account should have positive balance |
| Pendapatan Penjualan (4101) | 0 | -3,000,000 | Revenue account should have negative balance |
| PPN Keluaran (2103) | 0 | -330,000 | Tax liability should have negative balance |
| Bank UOB (1103) | +3,330,000 | +3,330,000 | Correct - asset account with positive balance |

## Testing and Verification

### 1. Create Test Sale
```go
// Create a test sale and verify journal entries
testSale := &models.Sale{
	// ... sale data ...
	Status: models.SaleStatusInvoiced,
	TotalAmount: 1000000,
	PPN: 110000,
}

correctedService := services.NewCorrectedSSOTSalesJournalService(db)
entry, err := correctedService.CreateSaleJournalEntry(testSale, userID)

// Verify entry is balanced
assert.Equal(t, entry.TotalDebit, entry.TotalCredit)
```

### 2. Test Payment Processing
```go
payment := &models.SalePayment{
	Amount: 1000000,
	PaymentMethod: "BANK_TRANSFER",
	// ... payment data ...
}

entry, err := correctedService.CreatePaymentJournalEntry(payment, userID)
// Verify correct account updates
```

## Migration Steps

### 1. Backup Current Data
```sql
-- Create backup tables
CREATE TABLE accounts_backup_YYYYMMDD AS SELECT * FROM accounts;
CREATE TABLE journal_entries_backup_YYYYMMDD AS SELECT * FROM journal_entries;
CREATE TABLE ssot_journal_entries_backup_YYYYMMDD AS SELECT * FROM ssot_journal_entries;
```

### 2. Deploy New Services
1. Deploy the corrected services
2. Update sales service integration
3. Update payment service integration

### 3. Fix Historical Data
1. Run balance correction script
2. Verify account balances
3. Test new transactions

### 4. Monitor and Validate
1. Monitor new sales transactions
2. Verify balance updates are correct
3. Run periodic balance reconciliation

## Preventing Future Issues

### 1. Add Balance Validation
```go
func (s *SSOTService) ValidateAccountBalance(accountID uint, expectedNormalBalance models.NormalBalanceType) error {
	var account models.Account
	s.db.First(&account, accountID)
	
	actualNormalBalance := account.GetNormalBalance()
	if actualNormalBalance != expectedNormalBalance {
		return fmt.Errorf("account %s has wrong normal balance type", account.Code)
	}
	
	// Check balance sign
	if expectedNormalBalance == models.NormalBalanceDebit && account.Balance < 0 {
		return fmt.Errorf("debit account %s has negative balance", account.Code)
	}
	if expectedNormalBalance == models.NormalBalanceCredit && account.Balance > 0 {
		return fmt.Errorf("credit account %s has positive balance", account.Code)
	}
	
	return nil
}
```

### 2. Add Unit Tests
```go
func TestSalesJournalEntry(t *testing.T) {
	// Test that sales journal entries are created correctly
	// Test that balances are updated correctly
	// Test that entries are balanced
}

func TestPaymentJournalEntry(t *testing.T) {
	// Test payment journal entries
	// Test AR reduction
	// Test bank balance increase
}
```

### 3. Add Balance Monitoring
Create a scheduled job to check for balance inconsistencies:

```go
func MonitorAccountBalances() {
	// Check that AR balance matches outstanding invoices
	// Check that revenue balance matches total invoiced sales
	// Alert if balances don't make sense
}
```

## Support and Troubleshooting

### Common Issues:

1. **"Account not found" errors**: Verify AccountResolver mappings match your Chart of Accounts
2. **"Unbalanced journal entry" errors**: Check that all amounts are calculated correctly
3. **Wrong balance signs**: Ensure you understand normal balance rules (Assets/Expenses = Debit, Others = Credit)

### Debug Commands:

```bash
# Check account mappings
go run diagnose_coa_balances.go

# Verify journal entries
SELECT * FROM ssot_journal_entries WHERE reference_type = 'SALES_INVOICE' ORDER BY created_at DESC LIMIT 10;

# Check account balances
SELECT code, name, type, balance FROM accounts WHERE code IN ('1201', '4101', '1102', '1103', '1104');
```

This comprehensive fix should resolve your Chart of Accounts balance issues and ensure proper accounting treatment of sales invoices and payments.