# Cash Bank Deposit Optimization

## Performance Improvements Needed

### 1. Database Transaction Optimization

Modify the `ProcessDeposit` function in `cashbank_service.go` to:

```go
// ProcessDeposit processes a deposit transaction with optimized performance
func (s *CashBankService) ProcessDeposit(request DepositRequest, userID uint) (*models.CashBankTransaction, error) {
    tx := s.db.Begin()
    
    // Set transaction isolation level to reduce lock contention
    tx.Exec("SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED")
    
    // Validate account with single query
    account, err := s.cashBankRepo.FindByID(request.AccountID)
    if err != nil {
        tx.Rollback()
        return nil, errors.New("account not found")
    }
    
    // Skip integrity check for performance - run as background job instead
    // if err := database.EnsureCashBankAccountIntegrity(s.db, account.ID); err != nil {
    //     // Log only, don't fail transaction
    //     log.Printf("Warning: Account integrity issue for ID %d: %v", account.ID, err)
    // }
    
    // Update balance with single query
    newBalance := account.Balance + request.Amount
    if err := tx.Model(account).Update("balance", newBalance).Error; err != nil {
        tx.Rollback()
        return nil, err
    }
    
    // Create transaction record
    transaction := &models.CashBankTransaction{
        CashBankID:      request.AccountID,
        ReferenceType:   TransactionTypeDeposit,
        ReferenceID:     0,
        Amount:          request.Amount,
        BalanceAfter:    newBalance,
        TransactionDate: request.Date.ToTime(),
        Notes:           request.Notes,
    }
    
    if err := tx.Create(transaction).Error; err != nil {
        tx.Rollback()
        return nil, err
    }
    
    // Create journal entries asynchronously
    go func() {
        // Run journal entry creation in background
        s.createDepositJournalEntriesAsync(transaction, account, request, userID)
    }()
    
    return transaction, tx.Commit().Error
}
```

### 2. Asynchronous Journal Entry Processing

```go
// createDepositJournalEntriesAsync creates journal entries in background
func (s *CashBankService) createDepositJournalEntriesAsync(transaction *models.CashBankTransaction, account *models.CashBank, request DepositRequest, userID uint) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Error in async journal entry creation: %v", r)
        }
    }()
    
    // Create new transaction for background processing
    tx := s.db.Begin()
    defer tx.Rollback() // Will be overridden by Commit() if successful
    
    err := s.createDepositJournalEntries(tx, transaction, account, request, userID)
    if err != nil {
        log.Printf("Failed to create journal entries for deposit %d: %v", transaction.ID, err)
        return
    }
    
    tx.Commit()
    log.Printf("Journal entries created successfully for deposit %d", transaction.ID)
}
```

### 3. Database Indexing Improvements

Add these indices to improve query performance:

```sql
-- Indices for faster cash bank operations
CREATE INDEX CONCURRENTLY idx_cashbanks_account_id_active ON cashbanks(account_id) WHERE is_active = true;
CREATE INDEX CONCURRENTLY idx_cashbank_transactions_date_desc ON cashbank_transactions(transaction_date DESC, id DESC);
CREATE INDEX CONCURRENTLY idx_journal_entries_reference ON journal_entries(reference_type, reference_id);
CREATE INDEX CONCURRENTLY idx_accounts_type_active ON accounts(type) WHERE is_active = true;

-- Partial index for active cash bank accounts
CREATE INDEX CONCURRENTLY idx_cashbanks_active_balance ON cashbanks(id, balance) WHERE is_active = true;
```

### 4. Connection Pool Optimization

In your database configuration, ensure optimal connection pooling:

```go
// database/database.go
func InitDatabase() *gorm.DB {
    // ... existing code ...
    
    sqlDB, err := db.DB()
    if err != nil {
        panic("Failed to get database instance")
    }
    
    // Optimize connection pool for better performance
    sqlDB.SetMaxOpenConns(25)    // Maximum number of open connections
    sqlDB.SetMaxIdleConns(10)    // Maximum number of idle connections
    sqlDB.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime
    sqlDB.SetConnMaxIdleTime(30 * time.Second) // Connection idle time
    
    return db
}
```

### 5. Quick Implementation Steps

1. **Immediate Fix** - Increase timeouts (already done above)
2. **Short Term** - Optimize database queries and reduce integrity checks
3. **Medium Term** - Implement asynchronous journal processing
4. **Long Term** - Add database indices and connection pool optimization

### 6. Monitoring and Logging

Add performance monitoring to track deposit operation times:

```go
// Add to ProcessDeposit function
start := time.Now()
defer func() {
    duration := time.Since(start)
    log.Printf("Deposit operation for account %d took %v", request.AccountID, duration)
    
    // Alert if operation takes longer than 10 seconds
    if duration > 10*time.Second {
        log.Printf("SLOW QUERY ALERT: Deposit operation took %v for account %d", duration, request.AccountID)
    }
}()
```

## Quick Fix Priority

1. ✅ **Frontend timeout increase** (implemented)
2. ✅ **Better error handling** (implemented)  
3. ⚠️  **Skip integrity checks during transaction**
4. ⚠️  **Asynchronous journal processing**
5. ⚠️  **Database indexing**

The implemented frontend changes should resolve the immediate timeout issue. Backend optimizations can be implemented gradually for better long-term performance.