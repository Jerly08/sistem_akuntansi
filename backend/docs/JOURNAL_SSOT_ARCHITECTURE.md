# Single Source of Truth (SSOT) untuk Sistem Jurnal

## Executive Summary

Dokumen ini merancang arsitektur Single Source of Truth (SSOT) untuk sistem jurnal akuntansi yang mengatasi fragmentasi data, inkonsistensi, dan kompleksitas koordinasi yang ada pada sistem saat ini.

## 1. PROBLEMATIKA SISTEM SAAT INI

### 1.1 Fragmentasi Data Jurnal
- `Journal` dan `JournalEntry` terpisah dengan relasi kompleks
- Logic pembuatan jurnal tersebar di multiple service
- Koordinasi antar transaksi yang rawan error dan race condition

### 1.2 Sumber Data Yang Tersebar
- **Sales**: `UnifiedSalesJournalService`, `SalesAccountingService`
- **Purchase**: `PurchaseAccountingService`
- **Cash/Bank**: `CashBankAccountingService`
- **Manual**: `JournalEntryController`
- **Asset**: `AssetService`

### 1.3 Sinkronisasi Balance Kompleks
- Service terpisah (`JournalBalanceSyncService`) untuk validasi
- Race condition antara posting jurnal dan update balance
- Validasi konsistensi yang berat dan tidak realtime

### 1.4 Duplikasi dan Inkonsistensi
- Auto-generation logic duplikat
- Validation rules tidak konsisten
- Complex locking mechanism (`JournalCoordinator`)

## 2. ARSITEKTUR SSOT YANG DIUSULKAN

### 2.1 Unified Journal Ledger (Tabel Utama)

```sql
-- Tabel utama jurnal yang menyatukan semua transaksi
CREATE TABLE unified_journal_ledger (
    id BIGSERIAL PRIMARY KEY,
    
    -- Transaction Identity
    transaction_uuid UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    entry_number VARCHAR(50) UNIQUE NOT NULL, -- JE-2024-01-0001
    
    -- Source Transaction
    source_type VARCHAR(30) NOT NULL, -- SALE, PURCHASE, PAYMENT, CASH_BANK, ASSET, MANUAL
    source_id BIGINT, -- Reference ke tabel sumber
    source_code VARCHAR(100), -- Code dari transaksi sumber
    
    -- Journal Entry Details
    entry_date DATE NOT NULL,
    description TEXT NOT NULL,
    reference VARCHAR(200),
    notes TEXT,
    
    -- Amounts (Always Balanced)
    total_debit DECIMAL(20,2) NOT NULL DEFAULT 0,
    total_credit DECIMAL(20,2) NOT NULL DEFAULT 0,
    
    -- Status & Control
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT', -- DRAFT, POSTED, REVERSED, CANCELLED
    is_balanced BOOLEAN NOT NULL DEFAULT TRUE,
    is_auto_generated BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Posting Information
    posted_at TIMESTAMPTZ,
    posted_by BIGINT REFERENCES users(id),
    
    -- Reversal Information
    reversed_by BIGINT, -- Points to reversing entry
    reversed_from BIGINT, -- Points to original entry
    reversal_reason TEXT,
    
    -- Audit Fields
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT chk_balanced CHECK (total_debit = total_credit),
    CONSTRAINT chk_amounts_positive CHECK (total_debit >= 0 AND total_credit >= 0),
    CONSTRAINT chk_status_valid CHECK (status IN ('DRAFT', 'POSTED', 'REVERSED', 'CANCELLED'))
);
```

### 2.2 Unified Journal Lines (Detail Transaksi)

```sql
-- Detail lines untuk setiap jurnal entry
CREATE TABLE unified_journal_lines (
    id BIGSERIAL PRIMARY KEY,
    
    -- Parent Journal
    journal_id BIGINT NOT NULL REFERENCES unified_journal_ledger(id) ON DELETE CASCADE,
    
    -- Account Information
    account_id BIGINT NOT NULL REFERENCES accounts(id),
    
    -- Line Details
    line_number SMALLINT NOT NULL,
    description TEXT,
    
    -- Amounts (Mutually Exclusive)
    debit_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    credit_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    
    -- Additional Information
    quantity DECIMAL(15,4), -- For inventory-related entries
    unit_price DECIMAL(15,4), -- For inventory-related entries
    
    -- Audit Fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT chk_amounts_not_both CHECK (
        NOT (debit_amount > 0 AND credit_amount > 0)
    ),
    CONSTRAINT chk_amounts_not_zero CHECK (
        debit_amount > 0 OR credit_amount > 0
    ),
    CONSTRAINT chk_amounts_positive CHECK (
        debit_amount >= 0 AND credit_amount >= 0
    ),
    
    -- Unique line numbering per journal
    UNIQUE(journal_id, line_number)
);
```

### 2.3 Account Balance Materialized View

```sql
-- Materialized view untuk real-time account balances
CREATE MATERIALIZED VIEW account_balances AS
WITH journal_totals AS (
    SELECT 
        jl.account_id,
        SUM(jl.debit_amount) as total_debits,
        SUM(jl.credit_amount) as total_credits,
        COUNT(*) as transaction_count
    FROM unified_journal_lines jl
    JOIN unified_journal_ledger jd ON jl.journal_id = jd.id
    WHERE jd.status = 'POSTED' 
      AND jd.deleted_at IS NULL
    GROUP BY jl.account_id
)
SELECT 
    a.id as account_id,
    a.code as account_code,
    a.name as account_name,
    a.type as account_type,
    a.normal_balance,
    
    COALESCE(jt.total_debits, 0) as total_debits,
    COALESCE(jt.total_credits, 0) as total_credits,
    COALESCE(jt.transaction_count, 0) as transaction_count,
    
    -- Calculate balance based on normal balance
    CASE 
        WHEN a.normal_balance = 'DEBIT' THEN 
            COALESCE(jt.total_debits, 0) - COALESCE(jt.total_credits, 0)
        WHEN a.normal_balance = 'CREDIT' THEN 
            COALESCE(jt.total_credits, 0) - COALESCE(jt.total_debits, 0)
        ELSE 0
    END as current_balance,
    
    NOW() as last_updated
FROM accounts a
LEFT JOIN journal_totals jt ON a.id = jt.account_id
WHERE a.deleted_at IS NULL;

-- Index untuk performance
CREATE UNIQUE INDEX idx_account_balances_account_id ON account_balances(account_id);
```

### 2.4 Journal Event Log (Audit Trail)

```sql
-- Event sourcing untuk audit trail
CREATE TABLE journal_event_log (
    id BIGSERIAL PRIMARY KEY,
    
    -- Event Identity
    event_uuid UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    
    -- Related Journal
    journal_id BIGINT REFERENCES unified_journal_ledger(id),
    
    -- Event Details
    event_type VARCHAR(50) NOT NULL, -- CREATED, POSTED, REVERSED, UPDATED, DELETED
    event_data JSONB NOT NULL, -- Full event payload
    event_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- User Context
    user_id BIGINT REFERENCES users(id),
    user_role VARCHAR(50),
    ip_address INET,
    user_agent TEXT,
    
    -- Additional Context
    source_system VARCHAR(50) DEFAULT 'ACCOUNTING_SYSTEM',
    correlation_id UUID, -- For tracing related events
    
    -- Constraints
    CONSTRAINT chk_event_type_valid CHECK (
        event_type IN ('CREATED', 'POSTED', 'REVERSED', 'UPDATED', 'DELETED', 'BALANCED', 'VALIDATED')
    )
);
```

## 3. INDEKS DAN PERFORMA

### 3.1 Primary Indexes

```sql
-- Journal Ledger Indexes
CREATE INDEX idx_journal_source ON unified_journal_ledger(source_type, source_id);
CREATE INDEX idx_journal_date_status ON unified_journal_ledger(entry_date, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_journal_posted ON unified_journal_ledger(posted_at) WHERE status = 'POSTED';
CREATE INDEX idx_journal_user_date ON unified_journal_ledger(created_by, entry_date DESC);

-- Journal Lines Indexes
CREATE INDEX idx_journal_lines_account ON unified_journal_lines(account_id, journal_id);
CREATE INDEX idx_journal_lines_amounts ON unified_journal_lines(account_id) WHERE debit_amount > 0 OR credit_amount > 0;

-- Event Log Indexes
CREATE INDEX idx_event_log_journal ON journal_event_log(journal_id, event_timestamp DESC);
CREATE INDEX idx_event_log_user ON journal_event_log(user_id, event_timestamp DESC);
CREATE INDEX idx_event_log_type_time ON journal_event_log(event_type, event_timestamp DESC);
```

### 3.2 Partitioning Strategy

```sql
-- Partition by date for large datasets
CREATE TABLE unified_journal_ledger_y2024m01 PARTITION OF unified_journal_ledger
FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Partition event log by date
CREATE TABLE journal_event_log_y2024m01 PARTITION OF journal_event_log
FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

## 4. DATABASE TRIGGERS & CONSTRAINTS

### 4.1 Balance Maintenance Trigger

```sql
-- Function to update materialized view when journal changes
CREATE OR REPLACE FUNCTION refresh_account_balances()
RETURNS TRIGGER AS $$
BEGIN
    -- Refresh only affected accounts for performance
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        REFRESH MATERIALIZED VIEW CONCURRENTLY account_balances;
    ELSIF TG_OP = 'DELETE' THEN
        REFRESH MATERIALIZED VIEW CONCURRENTLY account_balances;
    END IF;
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Trigger to maintain balance consistency
CREATE TRIGGER trg_refresh_account_balances
    AFTER INSERT OR UPDATE OR DELETE ON unified_journal_lines
    FOR EACH STATEMENT
    EXECUTE FUNCTION refresh_account_balances();
```

### 4.2 Validation Trigger

```sql
-- Function to validate journal entry balance
CREATE OR REPLACE FUNCTION validate_journal_balance()
RETURNS TRIGGER AS $$
DECLARE
    calculated_debit DECIMAL(20,2);
    calculated_credit DECIMAL(20,2);
BEGIN
    -- Calculate totals from lines
    SELECT 
        COALESCE(SUM(debit_amount), 0),
        COALESCE(SUM(credit_amount), 0)
    INTO calculated_debit, calculated_credit
    FROM unified_journal_lines
    WHERE journal_id = NEW.id;
    
    -- Update totals and validate balance
    NEW.total_debit := calculated_debit;
    NEW.total_credit := calculated_credit;
    NEW.is_balanced := (calculated_debit = calculated_credit);
    
    -- Prevent posting unbalanced entries
    IF NEW.status = 'POSTED' AND NOT NEW.is_balanced THEN
        RAISE EXCEPTION 'Cannot post unbalanced journal entry. Debit: %, Credit: %', 
                       calculated_debit, calculated_credit;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for validation
CREATE TRIGGER trg_validate_journal_balance
    BEFORE INSERT OR UPDATE ON unified_journal_ledger
    FOR EACH ROW
    EXECUTE FUNCTION validate_journal_balance();
```

## 5. APPLICATION LAYER DESIGN

### 5.1 Unified Journal Service

```go
// UnifiedJournalService - Single service for all journal operations
type UnifiedJournalService struct {
    db          *gorm.DB
    validator   *JournalValidator
    eventStore  *EventStore
    balanceView *BalanceViewManager
}

// TransactionRequest - Unified request for all transaction types
type TransactionRequest struct {
    TransactionUUID string                  `json:"transaction_uuid,omitempty"`
    SourceType      string                  `json:"source_type"`        // SALE, PURCHASE, etc
    SourceID        uint64                  `json:"source_id,omitempty"`
    SourceCode      string                  `json:"source_code,omitempty"`
    
    EntryDate       time.Time               `json:"entry_date"`
    Description     string                  `json:"description"`
    Reference       string                  `json:"reference,omitempty"`
    Notes           string                  `json:"notes,omitempty"`
    
    Lines           []JournalLineRequest    `json:"lines"`
    UserID          uint64                  `json:"user_id"`
    IsAutoGenerated bool                    `json:"is_auto_generated"`
}

// JournalLineRequest - Unified line request
type JournalLineRequest struct {
    AccountID     uint64          `json:"account_id"`
    Description   string          `json:"description,omitempty"`
    DebitAmount   decimal.Decimal `json:"debit_amount,omitempty"`
    CreditAmount  decimal.Decimal `json:"credit_amount,omitempty"`
    Quantity      *decimal.Decimal `json:"quantity,omitempty"`
    UnitPrice     *decimal.Decimal `json:"unit_price,omitempty"`
}

// CreateJournalEntry - Single method for all journal creation
func (ujs *UnifiedJournalService) CreateJournalEntry(ctx context.Context, req *TransactionRequest) (*UnifiedJournalEntry, error) {
    // 1. Validate request
    if err := ujs.validator.ValidateRequest(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // 2. Check for duplicates
    if existing, err := ujs.findExistingEntry(req.SourceType, req.SourceID); err == nil && existing != nil {
        return existing, nil // Return existing entry
    }
    
    // 3. Begin transaction
    return ujs.db.Transaction(func(tx *gorm.DB) (*UnifiedJournalEntry, error) {
        // 4. Create journal entry
        entry := &UnifiedJournalEntry{
            TransactionUUID: req.TransactionUUID,
            SourceType:      req.SourceType,
            SourceID:        req.SourceID,
            SourceCode:      req.SourceCode,
            EntryDate:       req.EntryDate,
            Description:     req.Description,
            Reference:       req.Reference,
            Notes:           req.Notes,
            Status:          "DRAFT",
            IsAutoGenerated: req.IsAutoGenerated,
            CreatedBy:       req.UserID,
        }
        
        if err := tx.Create(entry).Error; err != nil {
            return nil, fmt.Errorf("failed to create journal entry: %w", err)
        }
        
        // 5. Create journal lines
        totalDebit, totalCredit := decimal.Zero, decimal.Zero
        for i, line := range req.Lines {
            journalLine := &UnifiedJournalLine{
                JournalID:     entry.ID,
                AccountID:     line.AccountID,
                LineNumber:    i + 1,
                Description:   line.Description,
                DebitAmount:   line.DebitAmount,
                CreditAmount:  line.CreditAmount,
                Quantity:      line.Quantity,
                UnitPrice:     line.UnitPrice,
            }
            
            if err := tx.Create(journalLine).Error; err != nil {
                return nil, fmt.Errorf("failed to create journal line: %w", err)
            }
            
            totalDebit = totalDebit.Add(line.DebitAmount)
            totalCredit = totalCredit.Add(line.CreditAmount)
        }
        
        // 6. Update totals and validate balance
        entry.TotalDebit = totalDebit
        entry.TotalCredit = totalCredit
        entry.IsBalanced = totalDebit.Equal(totalCredit)
        
        if !entry.IsBalanced {
            return nil, fmt.Errorf("journal entry is not balanced: debit=%s, credit=%s", 
                                 totalDebit.String(), totalCredit.String())
        }
        
        if err := tx.Save(entry).Error; err != nil {
            return nil, fmt.Errorf("failed to update journal totals: %w", err)
        }
        
        // 7. Log event
        ujs.eventStore.LogEvent(ctx, "CREATED", entry.ID, req.UserID, map[string]interface{}{
            "source_type": req.SourceType,
            "source_id":   req.SourceID,
            "total_debit": totalDebit.String(),
            "total_credit": totalCredit.String(),
        })
        
        return entry, nil
    })
}
```

### 5.2 Transaction Factory Pattern

```go
// TransactionFactory - Factory untuk generate journal dari berbagai sumber
type TransactionFactory struct {
    accountResolver *AccountResolver
}

// CreateSaleTransaction - Generate journal request dari Sale
func (tf *TransactionFactory) CreateSaleTransaction(sale *Sale, userID uint64) (*TransactionRequest, error) {
    // Get required accounts
    arAccount, _ := tf.accountResolver.GetAccountByType("ACCOUNTS_RECEIVABLE")
    salesAccount, _ := tf.accountResolver.GetAccountByType("SALES_REVENUE")
    taxAccount, _ := tf.accountResolver.GetAccountByType("TAX_PAYABLE")
    
    req := &TransactionRequest{
        SourceType:      "SALE",
        SourceID:        sale.ID,
        SourceCode:      sale.Code,
        EntryDate:       sale.Date,
        Description:     fmt.Sprintf("Sales Invoice %s - %s", sale.Code, sale.Customer.Name),
        Reference:       sale.Code,
        UserID:          userID,
        IsAutoGenerated: true,
        Lines:           []JournalLineRequest{},
    }
    
    // 1. Debit: Accounts Receivable
    if sale.TotalAmount.GreaterThan(decimal.Zero) {
        req.Lines = append(req.Lines, JournalLineRequest{
            AccountID:    arAccount.ID,
            Description:  fmt.Sprintf("A/R - %s", sale.Customer.Name),
            DebitAmount:  sale.TotalAmount,
        })
    }
    
    // 2. Credit: Sales Revenue
    if sale.SubtotalAmount.GreaterThan(decimal.Zero) {
        req.Lines = append(req.Lines, JournalLineRequest{
            AccountID:    salesAccount.ID,
            Description:  "Sales Revenue",
            CreditAmount: sale.SubtotalAmount,
        })
    }
    
    // 3. Credit: Tax Payable (if applicable)
    if sale.TaxAmount.GreaterThan(decimal.Zero) {
        req.Lines = append(req.Lines, JournalLineRequest{
            AccountID:    taxAccount.ID,
            Description:  "VAT Payable",
            CreditAmount: sale.TaxAmount,
        })
    }
    
    return req, nil
}
```

## 6. MIGRATION STRATEGY

### 6.1 Phase 1: Schema Creation

```sql
-- Create new tables alongside existing ones
-- Run migrations incrementally during low-traffic periods
```

### 6.2 Phase 2: Data Migration

```go
// MigrationService - Service untuk migrasi data lama ke SSOT baru
type MigrationService struct {
    oldDB *gorm.DB
    newDB *gorm.DB
}

func (ms *MigrationService) MigrateJournalEntries() error {
    // 1. Migrate existing journal_entries
    var oldEntries []models.JournalEntry
    ms.oldDB.Preload("JournalLines").Find(&oldEntries)
    
    for _, oldEntry := range oldEntries {
        newEntry := &UnifiedJournalEntry{
            EntryNumber:      oldEntry.Code,
            SourceType:       oldEntry.ReferenceType,
            SourceID:         oldEntry.ReferenceID,
            SourceCode:       oldEntry.Reference,
            EntryDate:        oldEntry.EntryDate,
            Description:      oldEntry.Description,
            Reference:        oldEntry.Reference,
            Notes:            oldEntry.Notes,
            TotalDebit:       decimal.NewFromFloat(oldEntry.TotalDebit),
            TotalCredit:      decimal.NewFromFloat(oldEntry.TotalCredit),
            Status:           oldEntry.Status,
            IsBalanced:       oldEntry.IsBalanced,
            IsAutoGenerated:  oldEntry.IsAutoGenerated,
            PostedAt:         oldEntry.PostingDate,
            PostedBy:         oldEntry.PostedBy,
            CreatedBy:        oldEntry.UserID,
            CreatedAt:        oldEntry.CreatedAt,
            UpdatedAt:        oldEntry.UpdatedAt,
        }
        
        // Save to new table
        if err := ms.newDB.Create(newEntry).Error; err != nil {
            return fmt.Errorf("failed to migrate entry %s: %w", oldEntry.Code, err)
        }
        
        // Migrate lines
        for _, oldLine := range oldEntry.JournalLines {
            newLine := &UnifiedJournalLine{
                JournalID:     newEntry.ID,
                AccountID:     oldLine.AccountID,
                LineNumber:    oldLine.LineNumber,
                Description:   oldLine.Description,
                DebitAmount:   decimal.NewFromFloat(oldLine.DebitAmount),
                CreditAmount:  decimal.NewFromFloat(oldLine.CreditAmount),
                CreatedAt:     oldLine.CreatedAt,
                UpdatedAt:     oldLine.UpdatedAt,
            }
            
            if err := ms.newDB.Create(newLine).Error; err != nil {
                return fmt.Errorf("failed to migrate line: %w", err)
            }
        }
    }
    
    return nil
}
```

### 6.3 Phase 3: Application Update

```go
// Gradual replacement of existing services
// Update each module one by one to use UnifiedJournalService
```

### 6.4 Phase 4: Cleanup

```sql
-- Remove old tables after successful migration and validation
-- DROP TABLE journal_entries CASCADE;
-- DROP TABLE journal_lines CASCADE;
-- DROP TABLE journals CASCADE;
```

## 7. BENEFITS & ADVANTAGES

### 7.1 Data Consistency
- ✅ Single source of truth untuk semua transaksi jurnal
- ✅ Atomic transactions dengan ACID compliance
- ✅ Real-time balance consistency melalui materialized views
- ✅ Comprehensive audit trail dengan event sourcing

### 7.2 Performance
- ✅ Optimized indexing strategy
- ✅ Partitioning untuk scalability
- ✅ Materialized views untuk fast reporting
- ✅ Reduced complexity dalam balance calculations

### 7.3 Maintainability
- ✅ Unified API untuk semua journal operations
- ✅ Consistent validation rules
- ✅ Simplified testing dan debugging
- ✅ Single point of customization untuk business rules

### 7.4 Scalability
- ✅ Table partitioning untuk data besar
- ✅ Event sourcing untuk audit compliance
- ✅ Horizontal scaling capability
- ✅ Optimized for high-volume transactions

## 8. MONITORING & ALERTS

### 8.1 Health Checks
```sql
-- Balance consistency check
CREATE VIEW v_balance_health_check AS
SELECT 
    COUNT(*) FILTER (WHERE current_balance != 0) as accounts_with_balance,
    COUNT(*) FILTER (WHERE transaction_count = 0 AND current_balance != 0) as orphaned_balances,
    COUNT(*) FILTER (WHERE transaction_count > 0 AND current_balance = 0) as zero_balance_with_transactions
FROM account_balances;
```

### 8.2 Performance Monitoring
```sql
-- Journal performance metrics
CREATE VIEW v_journal_performance AS
SELECT 
    DATE_TRUNC('hour', created_at) as hour,
    source_type,
    COUNT(*) as entries_count,
    AVG(total_debit) as avg_amount,
    COUNT(*) FILTER (WHERE status = 'POSTED') as posted_count,
    AVG(EXTRACT(EPOCH FROM (posted_at - created_at))) as avg_posting_time_seconds
FROM unified_journal_ledger 
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY DATE_TRUNC('hour', created_at), source_type
ORDER BY hour DESC;
```

## 9. IMPLEMENTATION TIMELINE

### Phase 1 (Week 1-2): Core Infrastructure
- [ ] Create new database schema
- [ ] Setup materialized views
- [ ] Create database triggers
- [ ] Setup monitoring views

### Phase 2 (Week 3-4): Application Layer
- [ ] Implement UnifiedJournalService
- [ ] Create TransactionFactory
- [ ] Setup event sourcing
- [ ] Create migration tools

### Phase 3 (Week 5-6): Integration
- [ ] Update Sales module
- [ ] Update Purchase module
- [ ] Update Cash/Bank module
- [ ] Update Asset module

### Phase 4 (Week 7-8): Testing & Deployment
- [ ] Comprehensive testing
- [ ] Performance testing
- [ ] Data migration
- [ ] Production deployment

### Phase 5 (Week 9-10): Cleanup & Documentation
- [ ] Remove old code
- [ ] Update documentation
- [ ] Training and handover
- [ ] Post-deployment monitoring

---

**Document Version**: 1.0  
**Last Updated**: 2024-01-18  
**Author**: System Architect  
**Review Status**: Draft - Pending Review