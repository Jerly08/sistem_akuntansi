# Rencana Migrasi Sistem Jurnal ke Single Source of Truth (SSOT)

## Executive Summary

Dokumen ini menyediakan rencana migrasi yang detail dari sistem jurnal yang ada saat ini ke arsitektur Single Source of Truth (SSOT) yang telah dirancang. Migrasi dilakukan secara bertahap untuk meminimalkan downtime dan risiko.

## 1. COMPATIBILITY MAPPING

### 1.1 Table Mapping

| **Tabel Lama** | **Tabel SSOT** | **Status** | **Notes** |
|---|---|---|---|
| `journals` | `unified_journal_ledger` | MIGRATE | Header journal akan di-merge dengan entry |
| `journal_entries` | `unified_journal_ledger` | MIGRATE | Tabel utama untuk semua jurnal |
| `journal_lines` | `unified_journal_lines` | MIGRATE | Structure hampir sama |
| N/A | `journal_event_log` | NEW | Audit trail baru |
| N/A | `account_balances` (View) | NEW | Materialized view untuk balance |

### 1.2 Field Mapping

#### journals → unified_journal_ledger
```sql
-- Field mapping untuk tabel journals lama
SELECT 
    j.id as old_journal_id,
    gen_random_uuid() as transaction_uuid,
    j.code as entry_number,
    
    -- Determine source type from reference_type or set to MANUAL
    COALESCE(j.reference_type, 'MANUAL') as source_type,
    j.reference_id as source_id,
    j.code as source_code,
    
    j.date as entry_date,
    j.description,
    j.code as reference,
    '' as notes,
    
    j.total_debit,
    j.total_credit,
    j.status,
    true as is_balanced, -- Assume existing journals are balanced
    false as is_auto_generated, -- Old journals are mostly manual
    
    null as posted_at, -- Will be updated based on status
    null as posted_by,
    
    j.user_id as created_by,
    j.created_at,
    j.updated_at,
    j.deleted_at
FROM journals j;
```

#### journal_entries → unified_journal_ledger
```sql
-- Field mapping untuk tabel journal_entries lama
SELECT 
    je.id as old_journal_entry_id,
    gen_random_uuid() as transaction_uuid,
    je.code as entry_number,
    
    COALESCE(je.reference_type, 'MANUAL') as source_type,
    je.reference_id as source_id,
    je.reference as source_code,
    
    je.entry_date,
    je.description,
    je.reference,
    je.notes,
    
    je.total_debit,
    je.total_credit,
    je.status,
    je.is_balanced,
    je.is_auto_generated,
    
    je.posting_date as posted_at,
    je.posted_by,
    
    je.user_id as created_by,
    je.created_at,
    je.updated_at,
    je.deleted_at
FROM journal_entries je;
```

#### journal_lines → unified_journal_lines
```sql
-- Field mapping untuk journal_lines (minimal changes)
SELECT 
    jl.id as old_line_id,
    -- journal_id will be mapped to new unified_journal_ledger ID
    jl.account_id,
    jl.line_number,
    jl.description,
    jl.debit_amount,
    jl.credit_amount,
    null as quantity, -- New field
    null as unit_price, -- New field
    jl.created_at,
    jl.updated_at
FROM journal_lines jl;
```

### 1.3 Service Layer Mapping

| **Service Lama** | **Service SSOT** | **Migration Strategy** |
|---|---|---|
| `JournalEntryRepository` | `UnifiedJournalService` | Replace gradually per module |
| `JournalBatchService` | `UnifiedJournalService.BatchCreate` | Merge functionality |
| `JournalCoordinator` | Built-in duplicate prevention | Remove complex locking |
| `JournalBalanceSyncService` | Real-time materialized view | Replace with automated sync |
| `UnifiedSalesJournalService` | `TransactionFactory.CreateSale` | Simplify and standardize |
| `PurchaseAccountingService` | `TransactionFactory.CreatePurchase` | Simplify and standardize |
| `CashBankAccountingService` | `TransactionFactory.CreateCashBank` | Simplify and standardize |

## 2. MIGRATION PHASES

### Phase 1: Infrastructure Setup (Week 1-2)

#### 2.1 Database Schema Creation
```sql
-- Create new tables in parallel with existing ones
-- Add "_v2" suffix to avoid conflicts during migration

CREATE TABLE unified_journal_ledger_v2 (
    -- Full schema as defined in SSOT architecture
    id BIGSERIAL PRIMARY KEY,
    transaction_uuid UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    entry_number VARCHAR(50) UNIQUE NOT NULL,
    -- ... rest of schema
);

CREATE TABLE unified_journal_lines_v2 (
    -- Full schema as defined in SSOT architecture
    id BIGSERIAL PRIMARY KEY,
    journal_id BIGINT NOT NULL REFERENCES unified_journal_ledger_v2(id) ON DELETE CASCADE,
    -- ... rest of schema
);

CREATE TABLE journal_event_log_v2 (
    -- Full schema as defined in SSOT architecture
    -- ... schema
);
```

#### 2.2 Materialized Views Setup
```sql
-- Create materialized view for real-time balance tracking
CREATE MATERIALIZED VIEW account_balances_v2 AS
-- Full view definition as per SSOT architecture
-- ...

-- Setup refresh function and triggers
CREATE OR REPLACE FUNCTION refresh_account_balances_v2()
RETURNS TRIGGER AS $$
-- Trigger function implementation
-- ...
```

#### 2.3 Indexes and Performance Optimization
```sql
-- Create all necessary indexes for performance
CREATE INDEX CONCURRENTLY idx_journal_v2_source ON unified_journal_ledger_v2(source_type, source_id);
CREATE INDEX CONCURRENTLY idx_journal_v2_date_status ON unified_journal_ledger_v2(entry_date, status) WHERE deleted_at IS NULL;
-- ... all other indexes from architecture document
```

### Phase 2: Data Migration (Week 3)

#### 2.1 Migration Scripts
```go
// DataMigrationService handles the migration from old to new schema
type DataMigrationService struct {
    oldDB *gorm.DB
    newDB *gorm.DB
    logger *log.Logger
}

func (dms *DataMigrationService) MigrateAllJournalData() error {
    return dms.newDB.Transaction(func(tx *gorm.DB) error {
        // Step 1: Migrate journal entries (both journals and journal_entries tables)
        if err := dms.migrateJournalEntries(tx); err != nil {
            return fmt.Errorf("failed to migrate journal entries: %w", err)
        }
        
        // Step 2: Migrate journal lines
        if err := dms.migrateJournalLines(tx); err != nil {
            return fmt.Errorf("failed to migrate journal lines: %w", err)
        }
        
        // Step 3: Validate migration
        if err := dms.validateMigration(tx); err != nil {
            return fmt.Errorf("migration validation failed: %w", err)
        }
        
        // Step 4: Refresh materialized views
        if err := dms.refreshMaterializedViews(tx); err != nil {
            return fmt.Errorf("failed to refresh views: %w", err)
        }
        
        return nil
    })
}

func (dms *DataMigrationService) migrateJournalEntries(tx *gorm.DB) error {
    // Create mapping table to track old -> new ID relationships
    idMapping := make(map[uint]uint)
    
    // Step 1: Migrate from journal_entries table
    var journalEntries []models.JournalEntry
    if err := dms.oldDB.Find(&journalEntries).Error; err != nil {
        return fmt.Errorf("failed to fetch journal entries: %w", err)
    }
    
    for _, oldEntry := range journalEntries {
        newEntry := &UnifiedJournalEntry{
            TransactionUUID:  uuid.New(),
            EntryNumber:      oldEntry.Code,
            SourceType:       dms.normalizeSourceType(oldEntry.ReferenceType),
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
            DeletedAt:        oldEntry.DeletedAt,
        }
        
        if err := tx.Create(newEntry).Error; err != nil {
            return fmt.Errorf("failed to create new journal entry: %w", err)
        }
        
        idMapping[oldEntry.ID] = newEntry.ID
        dms.logger.Printf("Migrated journal entry %s (ID: %d -> %d)", oldEntry.Code, oldEntry.ID, newEntry.ID)
    }
    
    // Step 2: Migrate from journals table (if different structure)
    var journals []models.Journal
    if err := dms.oldDB.Find(&journals).Error; err != nil {
        return fmt.Errorf("failed to fetch journals: %w", err)
    }
    
    for _, oldJournal := range journals {
        // Check if already migrated from journal_entries
        if _, exists := idMapping[oldJournal.ID]; exists {
            continue
        }
        
        newEntry := &UnifiedJournalEntry{
            TransactionUUID:  uuid.New(),
            EntryNumber:      oldJournal.Code,
            SourceType:       dms.normalizeSourceType(oldJournal.ReferenceType),
            SourceID:         oldJournal.ReferenceID,
            EntryDate:        oldJournal.Date,
            Description:      oldJournal.Description,
            Reference:        oldJournal.Code,
            TotalDebit:       decimal.NewFromFloat(oldJournal.TotalDebit),
            TotalCredit:      decimal.NewFromFloat(oldJournal.TotalCredit),
            Status:           oldJournal.Status,
            IsBalanced:       oldJournal.TotalDebit == oldJournal.TotalCredit,
            IsAutoGenerated:  false, // Assume manual for old journals
            CreatedBy:        oldJournal.UserID,
            CreatedAt:        oldJournal.CreatedAt,
            UpdatedAt:        oldJournal.UpdatedAt,
            DeletedAt:        oldJournal.DeletedAt,
        }
        
        if err := tx.Create(newEntry).Error; err != nil {
            return fmt.Errorf("failed to create new journal entry from journal: %w", err)
        }
        
        idMapping[oldJournal.ID] = newEntry.ID
        dms.logger.Printf("Migrated journal %s (ID: %d -> %d)", oldJournal.Code, oldJournal.ID, newEntry.ID)
    }
    
    // Store mapping for line migration
    dms.journalIDMapping = idMapping
    return nil
}

func (dms *DataMigrationService) migrateJournalLines(tx *gorm.DB) error {
    var journalLines []models.JournalLine
    if err := dms.oldDB.Find(&journalLines).Error; err != nil {
        return fmt.Errorf("failed to fetch journal lines: %w", err)
    }
    
    for _, oldLine := range journalLines {
        // Map old journal_entry_id to new unified journal ID
        newJournalID, exists := dms.journalIDMapping[oldLine.JournalEntryID]
        if !exists {
            dms.logger.Printf("Warning: No mapping found for journal entry ID %d, skipping line %d", 
                oldLine.JournalEntryID, oldLine.ID)
            continue
        }
        
        newLine := &UnifiedJournalLine{
            JournalID:    newJournalID,
            AccountID:    oldLine.AccountID,
            LineNumber:   oldLine.LineNumber,
            Description:  oldLine.Description,
            DebitAmount:  decimal.NewFromFloat(oldLine.DebitAmount),
            CreditAmount: decimal.NewFromFloat(oldLine.CreditAmount),
            CreatedAt:    oldLine.CreatedAt,
            UpdatedAt:    oldLine.UpdatedAt,
        }
        
        if err := tx.Create(newLine).Error; err != nil {
            return fmt.Errorf("failed to create new journal line: %w", err)
        }
    }
    
    dms.logger.Printf("Migrated %d journal lines", len(journalLines))
    return nil
}

func (dms *DataMigrationService) normalizeSourceType(oldType string) string {
    switch oldType {
    case "SALE", "SALES":
        return "SALE"
    case "PURCHASE", "PURCHASES":
        return "PURCHASE"  
    case "PAYMENT", "PAYMENTS":
        return "PAYMENT"
    case "CASH_BANK", "CASHBANK":
        return "CASH_BANK"
    case "ASSET", "ASSETS":
        return "ASSET"
    case "MANUAL", "":
        return "MANUAL"
    default:
        return "MANUAL"
    }
}

func (dms *DataMigrationService) validateMigration(tx *gorm.DB) error {
    // Validation 1: Count comparison
    var oldCount, newCount int64
    
    dms.oldDB.Model(&models.JournalEntry{}).Count(&oldCount)
    tx.Model(&UnifiedJournalEntry{}).Count(&newCount)
    
    if oldCount != newCount {
        return fmt.Errorf("entry count mismatch: old=%d, new=%d", oldCount, newCount)
    }
    
    // Validation 2: Balance totals comparison
    var oldTotalDebit, oldTotalCredit float64
    var newTotalDebit, newTotalCredit decimal.Decimal
    
    dms.oldDB.Model(&models.JournalEntry{}).
        Select("SUM(total_debit), SUM(total_credit)").
        Row().Scan(&oldTotalDebit, &oldTotalCredit)
    
    tx.Model(&UnifiedJournalEntry{}).
        Select("SUM(total_debit), SUM(total_credit)").
        Row().Scan(&newTotalDebit, &newTotalCredit)
    
    if !decimal.NewFromFloat(oldTotalDebit).Equal(newTotalDebit) {
        return fmt.Errorf("total debit mismatch: old=%.2f, new=%s", 
            oldTotalDebit, newTotalDebit.String())
    }
    
    if !decimal.NewFromFloat(oldTotalCredit).Equal(newTotalCredit) {
        return fmt.Errorf("total credit mismatch: old=%.2f, new=%s", 
            oldTotalCredit, newTotalCredit.String())
    }
    
    dms.logger.Printf("✅ Migration validation passed - counts match: %d entries", oldCount)
    return nil
}
```

#### 2.2 Migration Verification
```go
// VerificationService validates the migrated data
type VerificationService struct {
    oldDB *gorm.DB
    newDB *gorm.DB
}

func (vs *VerificationService) VerifyMigration() error {
    checks := []func() error{
        vs.verifyJournalCounts,
        vs.verifyBalances,
        vs.verifyAccountIntegrity,
        vs.verifyUniqueConstraints,
    }
    
    for _, check := range checks {
        if err := check(); err != nil {
            return fmt.Errorf("verification failed: %w", err)
        }
    }
    
    return nil
}

func (vs *VerificationService) verifyJournalCounts() error {
    // Detailed verification logic
    // Compare counts, totals, balances between old and new systems
    return nil
}
```

### Phase 3: Application Layer Migration (Week 4-6)

#### 3.1 Service Layer Updates
```go
// UnifiedJournalService implementation
func NewUnifiedJournalService(db *gorm.DB) *UnifiedJournalService {
    return &UnifiedJournalService{
        db: db,
        validator: NewJournalValidator(),
        eventStore: NewEventStore(db),
        balanceView: NewBalanceViewManager(db),
    }
}

// Backward compatibility wrappers
type LegacyJournalService struct {
    unifiedService *UnifiedJournalService
}

func (ljs *LegacyJournalService) CreateJournalEntry(req *models.JournalEntryCreateRequest) (*models.JournalEntry, error) {
    // Convert legacy request to unified format
    unifiedReq := &TransactionRequest{
        SourceType:      req.ReferenceType,
        SourceID:        req.ReferenceID,
        EntryDate:       req.EntryDate,
        Description:     req.Description,
        Reference:       req.Reference,
        Notes:           req.Notes,
        UserID:          req.UserID,
        IsAutoGenerated: req.IsAutoGenerated,
    }
    
    // Use unified service
    unifiedEntry, err := ljs.unifiedService.CreateJournalEntry(context.Background(), unifiedReq)
    if err != nil {
        return nil, err
    }
    
    // Convert back to legacy format for backward compatibility
    return ljs.convertToLegacyFormat(unifiedEntry), nil
}
```

#### 3.2 Module-by-Module Updates
```go
// Sales module update
type SalesService struct {
    unifiedJournal *UnifiedJournalService
    factory        *TransactionFactory
}

func (ss *SalesService) CreateSaleWithJournal(sale *Sale, userID uint64) error {
    // Create sale record
    if err := ss.db.Create(sale).Error; err != nil {
        return fmt.Errorf("failed to create sale: %w", err)
    }
    
    // Generate journal transaction
    journalReq, err := ss.factory.CreateSaleTransaction(sale, userID)
    if err != nil {
        return fmt.Errorf("failed to create journal request: %w", err)
    }
    
    // Create journal entry using unified service
    _, err = ss.unifiedJournal.CreateJournalEntry(context.Background(), journalReq)
    if err != nil {
        // Rollback sale creation if journal fails
        ss.db.Delete(sale)
        return fmt.Errorf("failed to create journal entry: %w", err)
    }
    
    return nil
}
```

### Phase 4: Testing & Validation (Week 7)

#### 4.1 Data Consistency Tests
```go
func TestDataConsistency(t *testing.T) {
    // Test 1: Balance consistency
    t.Run("Balance Consistency", func(t *testing.T) {
        // Verify account balances match between old and new systems
        oldBalances := getOldAccountBalances()
        newBalances := getNewAccountBalances()
        
        for accountID, oldBalance := range oldBalances {
            newBalance := newBalances[accountID]
            assert.Equal(t, oldBalance, newBalance, 
                "Balance mismatch for account %d", accountID)
        }
    })
    
    // Test 2: Transaction totals
    t.Run("Transaction Totals", func(t *testing.T) {
        oldTotal := getOldTransactionTotals()
        newTotal := getNewTransactionTotals()
        
        assert.Equal(t, oldTotal.TotalDebit, newTotal.TotalDebit)
        assert.Equal(t, oldTotal.TotalCredit, newTotal.TotalCredit)
    })
}
```

#### 4.2 Performance Tests
```go
func BenchmarkJournalCreation(b *testing.B) {
    service := NewUnifiedJournalService(testDB)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        req := &TransactionRequest{
            SourceType:  "MANUAL",
            EntryDate:   time.Now(),
            Description: fmt.Sprintf("Test Entry %d", i),
            Lines: []JournalLineRequest{
                {AccountID: 1, DebitAmount: decimal.NewFromInt(100)},
                {AccountID: 2, CreditAmount: decimal.NewFromInt(100)},
            },
            UserID: 1,
        }
        
        _, err := service.CreateJournalEntry(context.Background(), req)
        if err != nil {
            b.Fatalf("Failed to create journal entry: %v", err)
        }
    }
}
```

### Phase 5: Deployment & Cutover (Week 8)

#### 5.1 Blue-Green Deployment Strategy
```yaml
# Deployment configuration
deployment:
  strategy: blue-green
  phases:
    1_schema_deployment:
      - Create new tables with "_v2" suffix
      - Deploy triggers and functions
      - Test schema in staging
      
    2_data_migration:
      - Run migration scripts during maintenance window
      - Verify data consistency
      - Keep old tables for rollback
      
    3_application_deployment:
      - Deploy new application version
      - Switch traffic gradually (10% -> 50% -> 100%)
      - Monitor performance and errors
      
    4_finalization:
      - Drop old tables after successful operation
      - Remove compatibility layers
      - Clean up temporary artifacts
```

#### 5.2 Rollback Plan
```sql
-- Rollback procedures if migration fails
-- Step 1: Revert application to use old tables
UPDATE application_config SET use_legacy_journal = true;

-- Step 2: Drop new tables if necessary
DROP TABLE IF EXISTS unified_journal_ledger_v2 CASCADE;
DROP TABLE IF EXISTS unified_journal_lines_v2 CASCADE;
DROP TABLE IF EXISTS journal_event_log_v2 CASCADE;

-- Step 3: Verify old system functionality
SELECT COUNT(*) FROM journal_entries; -- Should return original count
SELECT COUNT(*) FROM journal_lines;   -- Should return original count
```

## 3. RISK MITIGATION

### 3.1 Data Loss Prevention
- ✅ **Backup Strategy**: Full database backup before migration
- ✅ **Parallel Tables**: Keep old tables during migration period
- ✅ **Transaction Safety**: All migrations in database transactions
- ✅ **Verification Steps**: Multiple validation checkpoints

### 3.2 Downtime Minimization
- ✅ **Schema First**: Create new schema without affecting existing operations
- ✅ **Gradual Cutover**: Module-by-module migration approach
- ✅ **Blue-Green**: Parallel deployment with traffic switching
- ✅ **Instant Rollback**: Capability to revert quickly if issues arise

### 3.3 Performance Monitoring
- ✅ **Benchmark Tests**: Performance comparison before/after
- ✅ **Real-time Monitoring**: Database and application metrics
- ✅ **Load Testing**: Stress testing with production-like load
- ✅ **Query Analysis**: Optimization of slow queries

## 4. VALIDATION CHECKLIST

### 4.1 Pre-Migration Checklist
- [ ] Full database backup completed
- [ ] New schema created and tested
- [ ] Migration scripts tested in staging
- [ ] Rollback procedures verified
- [ ] Performance benchmarks established
- [ ] Team trained on new architecture

### 4.2 Post-Migration Checklist
- [ ] Data counts match exactly
- [ ] Balance totals consistent
- [ ] All unique constraints working
- [ ] Performance within acceptable limits
- [ ] No data corruption detected
- [ ] Application functions correctly
- [ ] Monitoring systems operational
- [ ] Documentation updated

## 5. SUCCESS CRITERIA

### 5.1 Data Integrity
- ✅ 100% data preservation (zero data loss)
- ✅ All balance calculations accurate
- ✅ Referential integrity maintained
- ✅ No duplicate entries created

### 5.2 Performance
- ✅ Journal creation ≤ 200ms (95th percentile)
- ✅ Balance queries ≤ 50ms (95th percentile)
- ✅ Reporting queries ≤ 2s (95th percentile)
- ✅ Database space usage reduced by 20%

### 5.3 Operational
- ✅ Zero downtime during cutover
- ✅ All integrations working correctly
- ✅ Error rates within normal limits
- ✅ Team comfortable with new system

---

**Document Version**: 1.0  
**Last Updated**: 2024-01-18  
**Author**: System Architect  
**Review Status**: Draft - Ready for Review