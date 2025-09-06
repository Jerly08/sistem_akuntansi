# CashBank-COA Synchronization Implementation Plan

## Overview
Implementasi untuk memastikan Cash & Bank selalu sync dengan Chart of Accounts (COA) melalui automatic journal entries dan validation mechanisms.

## Phase 1: Core Implementation (Critical)

### 1. Automatic Journal Entry Creation

**File**: `services/cashbank_accounting_service.go`
```go
package services

import (
    "fmt"
    "app-sistem-akuntansi/models"
    "gorm.io/gorm"
)

type CashBankAccountingService struct {
    db             *gorm.DB
    journalService *JournalService
}

func NewCashBankAccountingService(db *gorm.DB, journalService *JournalService) *CashBankAccountingService {
    return &CashBankAccountingService{
        db:             db,
        journalService: journalService,
    }
}

// CreateTransactionWithJournal creates CashBank transaction and corresponding journal entry
func (s *CashBankAccountingService) CreateTransactionWithJournal(
    cashBankID uint, 
    amount float64, 
    referenceType string, 
    referenceID uint, 
    notes string,
    counterAccountID uint, // Account for the other side of the transaction
) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 1. Get CashBank with linked COA account
        var cashBank models.CashBank
        if err := tx.Preload("Account").First(&cashBank, cashBankID).Error; err != nil {
            return fmt.Errorf("cash bank not found: %v", err)
        }
        
        if cashBank.AccountID == 0 {
            return fmt.Errorf("cash bank %s not linked to COA account", cashBank.Name)
        }

        // 2. Create CashBank Transaction
        cashBankTx := &models.CashBankTransaction{
            CashBankID:      cashBankID,
            Amount:          amount,
            BalanceAfter:    cashBank.Balance + amount,
            TransactionDate: time.Now(),
            ReferenceType:   referenceType,
            ReferenceID:     referenceID,
            Notes:          notes,
        }
        
        if err := tx.Create(cashBankTx).Error; err != nil {
            return fmt.Errorf("failed to create cash bank transaction: %v", err)
        }

        // 3. Update CashBank Balance
        if err := tx.Model(&cashBank).Update("balance", cashBank.Balance + amount).Error; err != nil {
            return fmt.Errorf("failed to update cash bank balance: %v", err)
        }

        // 4. Create Journal Entry
        journal := &models.Journal{
            Date:        cashBankTx.TransactionDate,
            Code:        s.generateJournalCode("CB", cashBankTx.ID),
            Description: fmt.Sprintf("CashBank: %s - %s", cashBank.Name, notes),
            ReferenceType: "CASHBANK_TRANSACTION",
            ReferenceID:   cashBankTx.ID,
            Status:      models.JournalStatusPosted,
            UserID:      1, // TODO: Get from context
        }

        if err := tx.Create(journal).Error; err != nil {
            return fmt.Errorf("failed to create journal: %v", err)
        }

        // 5. Create Journal Entries (Double Entry)
        entries := []models.JournalEntry{
            // Debit Cash/Bank Account (if amount > 0) or Credit (if amount < 0)
            {
                JournalID:    journal.ID,
                AccountID:    cashBank.AccountID,
                DebitAmount:  math.Max(amount, 0),
                CreditAmount: math.Max(-amount, 0),
                Description:  notes,
            },
            // Credit/Debit Counter Account
            {
                JournalID:    journal.ID,
                AccountID:    counterAccountID,
                DebitAmount:  math.Max(-amount, 0),
                CreditAmount: math.Max(amount, 0),
                Description:  notes,
            },
        }

        for _, entry := range entries {
            if err := tx.Create(&entry).Error; err != nil {
                return fmt.Errorf("failed to create journal entry: %v", err)
            }
        }

        // 6. Update COA Account Balances
        if err := s.updateAccountBalance(tx, cashBank.AccountID, amount); err != nil {
            return fmt.Errorf("failed to update cash bank COA balance: %v", err)
        }
        
        if err := s.updateAccountBalance(tx, counterAccountID, -amount); err != nil {
            return fmt.Errorf("failed to update counter account balance: %v", err)
        }

        return nil
    })
}

func (s *CashBankAccountingService) updateAccountBalance(tx *gorm.DB, accountID uint, amount float64) error {
    var account models.Account
    if err := tx.First(&account, accountID).Error; err != nil {
        return err
    }

    newBalance := account.Balance
    // Apply accounting rules based on account type
    switch account.Type {
    case models.AccountTypeAsset:
        newBalance += amount // Debit increases assets
    case models.AccountTypeLiability, models.AccountTypeEquity, models.AccountTypeRevenue:
        newBalance -= amount // Credit increases liabilities/equity/revenue
    case models.AccountTypeExpense:
        newBalance += amount // Debit increases expenses
    }

    return tx.Model(&account).Update("balance", newBalance).Error
}

func (s *CashBankAccountingService) generateJournalCode(prefix string, id uint) string {
    return fmt.Sprintf("%s-%04d-%d", prefix, time.Now().Year(), id)
}
```

### 2. Database Trigger for Safety

**File**: `database/cashbank_coa_sync_trigger.sql`
```sql
-- Function to sync CashBank balance changes to COA
CREATE OR REPLACE FUNCTION sync_cashbank_balance_to_coa()
RETURNS TRIGGER AS $$
DECLARE
    coa_account_id INTEGER;
    transaction_sum DECIMAL(15,2);
BEGIN
    -- Get the linked COA account ID
    IF TG_OP = 'DELETE' THEN
        SELECT account_id INTO coa_account_id FROM cash_banks WHERE id = OLD.cash_bank_id;
    ELSE
        SELECT account_id INTO coa_account_id FROM cash_banks WHERE id = NEW.cash_bank_id;
    END IF;
    
    -- Skip if no linked COA account
    IF coa_account_id IS NULL THEN
        RETURN COALESCE(NEW, OLD);
    END IF;
    
    -- Calculate total transaction sum for this CashBank
    SELECT COALESCE(SUM(amount), 0) INTO transaction_sum
    FROM cash_bank_transactions
    WHERE cash_bank_id = COALESCE(NEW.cash_bank_id, OLD.cash_bank_id)
    AND deleted_at IS NULL;
    
    -- Update CashBank balance to match transaction sum
    UPDATE cash_banks 
    SET balance = transaction_sum 
    WHERE id = COALESCE(NEW.cash_bank_id, OLD.cash_bank_id);
    
    -- Update linked COA account balance to match CashBank balance
    UPDATE accounts 
    SET balance = transaction_sum 
    WHERE id = coa_account_id;
    
    -- Log the sync action
    INSERT INTO audit_logs (
        table_name, 
        action, 
        record_id, 
        old_values, 
        new_values,
        created_at
    ) VALUES (
        'cashbank_coa_sync',
        'AUTO_SYNC',
        coa_account_id,
        '{}',
        json_build_object('balance', transaction_sum, 'cash_bank_id', COALESCE(NEW.cash_bank_id, OLD.cash_bank_id)),
        NOW()
    );
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Create trigger on cash_bank_transactions
DROP TRIGGER IF EXISTS trg_sync_cashbank_coa ON cash_bank_transactions;
CREATE TRIGGER trg_sync_cashbank_coa
    AFTER INSERT OR UPDATE OR DELETE ON cash_bank_transactions
    FOR EACH ROW
    EXECUTE FUNCTION sync_cashbank_balance_to_coa();
```

### 3. Validation Service

**File**: `services/cashbank_validation_service.go`
```go
package services

import (
    "fmt"
    "app-sistem-akuntansi/models"
    "gorm.io/gorm"
)

type CashBankValidationService struct {
    db *gorm.DB
}

func NewCashBankValidationService(db *gorm.DB) *CashBankValidationService {
    return &CashBankValidationService{db: db}
}

type SyncDiscrepancy struct {
    CashBankID     uint    `json:"cash_bank_id"`
    CashBankName   string  `json:"cash_bank_name"`
    CashBankCode   string  `json:"cash_bank_code"`
    COAAccountID   uint    `json:"coa_account_id"`
    COAAccountCode string  `json:"coa_account_code"`
    CashBankBalance float64 `json:"cash_bank_balance"`
    COABalance     float64 `json:"coa_balance"`
    TransactionSum float64 `json:"transaction_sum"`
    Discrepancy    float64 `json:"discrepancy"`
}

func (s *CashBankValidationService) FindSyncDiscrepancies() ([]SyncDiscrepancy, error) {
    var discrepancies []SyncDiscrepancy
    
    err := s.db.Raw(`
        SELECT 
            cb.id as cash_bank_id,
            cb.name as cash_bank_name,
            cb.code as cash_bank_code,
            a.id as coa_account_id,
            a.code as coa_account_code,
            cb.balance as cash_bank_balance,
            a.balance as coa_balance,
            COALESCE(tx_sum.transaction_sum, 0) as transaction_sum,
            (cb.balance - a.balance) as discrepancy
        FROM cash_banks cb
        JOIN accounts a ON cb.account_id = a.id
        LEFT JOIN (
            SELECT 
                cash_bank_id,
                SUM(amount) as transaction_sum
            FROM cash_bank_transactions 
            WHERE deleted_at IS NULL 
            GROUP BY cash_bank_id
        ) tx_sum ON cb.id = tx_sum.cash_bank_id
        WHERE cb.deleted_at IS NULL 
          AND a.deleted_at IS NULL
          AND cb.balance != a.balance
    `).Scan(&discrepancies).Error
    
    return discrepancies, err
}

func (s *CashBankValidationService) ValidateAllSync() error {
    discrepancies, err := s.FindSyncDiscrepancies()
    if err != nil {
        return err
    }
    
    if len(discrepancies) > 0 {
        return fmt.Errorf("found %d cash bank/COA sync discrepancies", len(discrepancies))
    }
    
    return nil
}

func (s *CashBankValidationService) AutoFixDiscrepancies() (int, error) {
    discrepancies, err := s.FindSyncDiscrepancies()
    if err != nil {
        return 0, err
    }
    
    fixedCount := 0
    for _, d := range discrepancies {
        // Use transaction sum as source of truth
        correctBalance := d.TransactionSum
        
        err := s.db.Transaction(func(tx *gorm.DB) error {
            // Update CashBank balance
            if err := tx.Model(&models.CashBank{}).
                Where("id = ?", d.CashBankID).
                Update("balance", correctBalance).Error; err != nil {
                return err
            }
            
            // Update COA balance
            if err := tx.Model(&models.Account{}).
                Where("id = ?", d.COAAccountID).
                Update("balance", correctBalance).Error; err != nil {
                return err
            }
            
            return nil
        })
        
        if err != nil {
            return fixedCount, fmt.Errorf("failed to fix discrepancy for %s: %v", d.CashBankName, err)
        }
        
        fixedCount++
    }
    
    return fixedCount, nil
}
```

## Integration Points

### 1. Update CashBank Service
Modify existing `services/cashbank_service.go` to use the new accounting service:

```go
// In existing methods, replace direct balance updates with:
func (s *CashBankService) ProcessPayment(paymentID uint, cashBankID uint, amount float64) error {
    // Get payment details for counter account
    var payment models.Payment
    if err := s.db.First(&payment, paymentID).Error; err != nil {
        return err
    }
    
    // Use accounting service instead of direct update
    return s.accountingService.CreateTransactionWithJournal(
        cashBankID,
        amount,
        "PAYMENT",
        paymentID,
        fmt.Sprintf("Payment %s", payment.Code),
        payment.AccountID, // Counter account from payment
    )
}
```

### 2. Add Validation Middleware
```go
func (m *ValidationMiddleware) ValidateCashBankSync(c *gin.Context) {
    if err := m.validationService.ValidateAllSync(); err != nil {
        c.JSON(500, gin.H{"error": "CashBank-COA sync validation failed", "details": err.Error()})
        c.Abort()
        return
    }
    c.Next()
}
```

## Testing Strategy

### 1. Unit Tests
- Test automatic journal entry creation
- Test balance synchronization
- Test validation service

### 2. Integration Tests  
- Test complete payment flow
- Test trigger functionality
- Test auto-fix mechanisms

### 3. Load Tests
- Test performance with large transaction volumes
- Test trigger performance impact

## Monitoring

### 1. Health Check Endpoint
```go
// GET /api/health/cashbank-sync
func (c *HealthController) CashBankSyncStatus(ctx *gin.Context) {
    discrepancies, _ := c.validationService.FindSyncDiscrepancies()
    
    status := "healthy"
    if len(discrepancies) > 0 {
        status = "unhealthy"
    }
    
    ctx.JSON(200, gin.H{
        "status": status,
        "discrepancies_count": len(discrepancies),
        "discrepancies": discrepancies,
    })
}
```

### 2. Scheduled Jobs
```go
// Daily reconciliation job
func (s *SchedulerService) RunDailyCashBankReconciliation() {
    fixed, err := s.validationService.AutoFixDiscrepancies()
    if err != nil {
        s.logger.Error("Daily reconciliation failed", err)
        s.alertService.SendAlert("CashBank reconciliation failed")
        return
    }
    
    if fixed > 0 {
        s.logger.Info(fmt.Sprintf("Daily reconciliation fixed %d discrepancies", fixed))
    }
}
```

## Implementation Timeline

**Week 1**: Core Implementation
- CashBankAccountingService
- Database triggers  
- Basic validation

**Week 2**: Integration & Testing
- Update existing services
- Comprehensive testing
- Performance optimization

**Week 3**: Monitoring & Production
- Health checks
- Scheduled jobs
- Production deployment

This implementation will ensure that Cash & Bank always stays synchronized with COA through automatic journal entries and robust validation mechanisms.
