# CashBank-COA Synchronization Phase 1 Deployment Guide

## Overview
Phase 1 implementasi CashBank-COA synchronization telah selesai dengan fitur-fitur berikut:
- ✅ CashBankAccountingService - Automatic journal entry creation
- ✅ CashBankValidationService - Discrepancy detection and fixing  
- ✅ Database triggers - Safety net synchronization
- ✅ CashBankEnhancedService - Integration wrapper
- ✅ Validation middleware - Health checks and API endpoints

## Files Created

### Services
1. `services/cashbank_accounting_service.go` - Core accounting service
2. `services/cashbank_validation_service.go` - Validation and fixing service  
3. `services/cashbank_enhanced_service.go` - Integration wrapper service

### Database
4. `database/migrations/cashbank_coa_sync_trigger.sql` - Database triggers

### Middleware
5. `middleware/cashbank_validation_middleware.go` - API validation middleware

### Documentation & Testing
6. `docs/CASHBANK_COA_SYNC_IMPLEMENTATION.md` - Technical implementation details
7. `scripts/test_cashbank_sync_phase1.go` - Test script

## Deployment Steps

### 1. Database Setup

**Install Database Triggers** (Critical for safety net):
```sql
-- Run the SQL commands in database/migrations/cashbank_coa_sync_trigger.sql
-- This creates triggers and audit log table
```

**Verify Installation**:
```sql
-- Check if triggers are installed
SELECT trigger_name, event_object_table, action_timing, event_manipulation 
FROM information_schema.triggers 
WHERE trigger_name LIKE '%sync_cashbank%';

-- Check if audit_logs table exists
SELECT table_name FROM information_schema.tables WHERE table_name = 'audit_logs';

-- Test trigger validation function
SELECT * FROM validate_cashbank_coa_integrity();
```

### 2. Service Integration

**Update Dependency Injection** (example in main.go or service container):
```go
// Add to your service initialization
cashBankRepo := repositories.NewCashBankRepository(db)
accountRepo := repositories.NewAccountRepository(db)

// Create new enhanced service
cashBankService := services.NewCashBankEnhancedService(db, cashBankRepo, accountRepo)

// Create validation middleware
accountingService := services.NewCashBankAccountingService(db)
validationService := services.NewCashBankValidationService(db, accountingService)
validationMiddleware := middleware.NewCashBankValidationMiddleware(validationService, cashBankService)
```

**Update Controllers** (gradually migrate):
```go
// Replace old service calls with enhanced service
// Old: cashBankService.ProcessDeposit(...)
// New: cashBankService.ProcessDepositV2(..., sourceAccountID)
```

### 3. API Routes Setup

**Add Validation Routes**:
```go
// Add to your router setup
api := router.Group("/api")
validationMiddleware.AddRoutes(api)

// Apply validation middleware to cash bank operations
cashbank := api.Group("/cashbank")
cashbank.Use(validationMiddleware.ValidateCashBankSync())
{
    // Your existing cashbank routes here
}
```

### 4. Initial Data Setup

**Fix Existing Discrepancies**:
```bash
# Run the test script to identify issues
go run scripts/test_cashbank_sync_phase1.go

# Or use API endpoints:
# GET /api/health/cashbank/sync - Check status
# POST /api/cashbank/sync/fix - Auto-fix issues
```

**Link Unlinked Accounts**:
```sql
-- Find unlinked cash banks
SELECT id, name, code, account_id FROM cash_banks WHERE account_id IS NULL OR account_id = 0;

-- Link manually if needed, or use API:
-- POST /api/cashbank/sync/link with {"cash_bank_id": 1, "account_id": 1105}
```

### 5. Testing & Verification

**Run Phase 1 Tests**:
```bash
go run scripts/test_cashbank_sync_phase1.go
```

**API Health Checks**:
```bash
# Check overall health
curl GET http://localhost:8080/api/health/cashbank

# Detailed sync status
curl GET http://localhost:8080/api/cashbank/sync/status

# Auto-fix any issues
curl -X POST http://localhost:8080/api/cashbank/sync/fix
```

## Configuration

### Environment Variables (if needed)
```env
# Add to your .env file if you want to disable/enable features
CASHBANK_SYNC_ENABLED=true
CASHBANK_VALIDATION_STRICT=true
CASHBANK_AUTO_FIX_ENABLED=true
```

### Logging
Ensure your logging captures:
- Sync validation failures
- Auto-fix actions
- Database trigger activities (check audit_logs table)

## Monitoring

### Health Check Endpoints
- `GET /api/health/cashbank` - Overall health status
- `GET /api/health/cashbank/sync` - Detailed sync status
- `GET /api/cashbank/sync/status` - Full validation report

### Key Metrics to Monitor
- Discrepancy count: Should be 0 for healthy system
- Linked account ratio: Should be close to 100%
- Auto-fix success rate
- Trigger execution frequency (from audit_logs)

### Alerting Thresholds
- **Critical**: Any BALANCE_MISMATCH or TRANSACTION_SUM_MISMATCH
- **Warning**: More than 10% accounts NOT_LINKED  
- **Info**: Auto-fix operations completed

## Rollback Plan

If issues occur, you can safely rollback:

1. **Disable Validation Middleware**:
```go
// Comment out validation middleware temporarily
// cashbank.Use(validationMiddleware.ValidateCashBankSync())
```

2. **Use Legacy Services**:
```go
// Revert to original CashBankService
// legacyCashBankService := services.NewCashBankService(db, cashBankRepo, accountRepo)
```

3. **Database Triggers** (keep enabled for safety):
```sql
-- Triggers provide safety net and should remain active
-- Only disable if they cause performance issues
-- DROP TRIGGER IF EXISTS trg_sync_cashbank_coa ON cash_bank_transactions;
```

## Performance Considerations

### Expected Impact
- **Database**: Minimal impact from triggers (< 1ms per transaction)
- **API**: Validation middleware adds ~10-50ms for write operations
- **Memory**: New services add ~5-10MB memory usage

### Optimization Options
- Cache validation results for read operations
- Batch sync operations during low traffic periods
- Use database connection pooling for heavy validation queries

## Next Steps (Future Phases)

### Phase 2: Event-Driven Architecture
- Implement event bus for real-time sync
- Add message queue for async processing
- Enhanced monitoring dashboard

### Phase 3: Advanced Features  
- Automated reconciliation workflows
- Integration with external bank APIs
- Advanced reporting and analytics

## Troubleshooting

### Common Issues

**"Cash bank account must be linked to COA before processing"**
```bash
# Solution: Link the account
curl -X POST http://localhost:8080/api/cashbank/sync/link \
  -H "Content-Type: application/json" \
  -d '{"cash_bank_id": 1, "account_id": 1105}'
```

**"Sync validation failed"**  
```bash
# Check status and auto-fix
curl GET http://localhost:8080/api/cashbank/sync/status
curl -X POST http://localhost:8080/api/cashbank/sync/fix
```

**Database trigger not working**
```sql
-- Verify trigger installation
SELECT * FROM pg_trigger WHERE tgname LIKE '%sync_cashbank%';

-- Check audit logs for trigger activity
SELECT * FROM audit_logs WHERE table_name = 'cashbank_coa_sync' ORDER BY created_at DESC LIMIT 10;
```

**Performance issues**
```sql
-- Check for missing indexes
CREATE INDEX IF NOT EXISTS idx_cash_bank_transactions_cash_bank_id ON cash_bank_transactions(cash_bank_id);
CREATE INDEX IF NOT EXISTS idx_cash_banks_account_id ON cash_banks(account_id);
```

## Support

For issues or questions:
1. Check the test script output: `go run scripts/test_cashbank_sync_phase1.go`
2. Review health check endpoints
3. Examine audit_logs table for detailed trigger activity
4. Check application logs for service-level errors

---

**Status**: ✅ Ready for Production Deployment  
**Last Updated**: 2025-09-06  
**Phase**: 1 - Core Implementation Complete
