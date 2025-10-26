# Fix: Concurrent Materialized View Refresh Error

## Problem

Error terjadi di PC lain:
```
ERROR: cannot refresh materialized view "public.account_balances" concurrently (SQLSTATE 55000)
```

**Root Cause:**
- Trigger `trg_refresh_account_balances` di `020_create_unified_journal_ssot.sql` otomatis refresh materialized view setiap INSERT/UPDATE/DELETE di `unified_journal_lines`
- `REFRESH MATERIALIZED VIEW CONCURRENTLY` tidak bisa jalan simultan
- Di environment dengan multiple concurrent requests, trigger ini menyebabkan conflict

## Solution Implemented

### 1. **Hapus Trigger Bermasalah** ✅
```sql
-- Run this fix script
psql -U postgres -d accounting_db -f backend/fix_concurrent_refresh_error.sql
```

### 2. **Balance Sync Strategy** ✅

**Real-time Balance Sync:**
- Handled oleh `setup_automatic_balance_sync.sql` 
- Trigger `trg_sync_account_balance_on_line_change` update `accounts.balance` per transaction
- Trigger `trg_sync_account_balance_on_status_change` sync saat journal status berubah
- **No concurrent conflicts** karena update per account, bukan full view refresh

**Materialized View (untuk reporting):**
- Digunakan untuk aggregated reports (balance sheet, income statement)
- Refresh via scheduled job atau manual API call
- **No automatic trigger** untuk mencegah concurrent conflicts

### 3. **New Functions**

#### Manual Refresh
```sql
-- Refresh materialized view (gunakan di scheduled job atau API)
SELECT * FROM manual_refresh_account_balances();
```

Output:
```
success | message                                  | refreshed_at
--------+------------------------------------------+-------------------------
true    | Account balances refreshed in 00:00:02.5 | 2025-10-26 10:45:30+07
```

#### Check Freshness
```sql
-- Cek apakah perlu di-refresh
SELECT * FROM check_account_balances_freshness();
```

Output:
```
last_updated            | age_minutes | needs_refresh
------------------------+-------------+---------------
2025-10-26 09:30:00+07  | 75          | true
```

## Implementation Steps

### Untuk Database yang Sudah Error:

1. **Apply Fix SQL**
   ```bash
   psql -U postgres -d accounting_db -f backend/fix_concurrent_refresh_error.sql
   ```

2. **Verify Fix**
   ```sql
   -- Check trigger sudah dihapus
   SELECT tgname FROM pg_trigger WHERE tgname = 'trg_refresh_account_balances';
   -- Harusnya kosong
   
   -- Test manual refresh
   SELECT * FROM manual_refresh_account_balances();
   ```

3. **Setup Scheduled Refresh (Optional)**
   
   **Option A: Via Go Scheduler**
   ```bash
   # Set environment variables
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=your_password
   export DB_NAME=accounting_db
   export REFRESH_INTERVAL=1h  # Refresh every 1 hour
   
   # Run scheduler
   go run backend/cmd/scripts/refresh_mv_scheduler.go
   ```
   
   **Option B: Via Cron Job**
   ```bash
   # Add to crontab (refresh every hour)
   0 * * * * psql -U postgres -d accounting_db -c "SELECT * FROM manual_refresh_account_balances();"
   ```
   
   **Option C: Via API Endpoint** (Recommended)
   Add endpoint di controller:
   ```go
   // controllers/system_controller.go
   func (c *SystemController) RefreshMaterializedView(ctx *gin.Context) {
       var result struct {
           Success     bool      `json:"success"`
           Message     string    `json:"message"`
           RefreshedAt time.Time `json:"refreshed_at"`
       }
       
       err := c.db.Raw("SELECT success, message, refreshed_at FROM manual_refresh_account_balances()").Scan(&result).Error
       if err != nil {
           ctx.JSON(500, gin.H{"error": err.Error()})
           return
       }
       
       ctx.JSON(200, result)
   }
   ```

### Untuk Fresh Install:

Migration `020_create_unified_journal_ssot.sql` sudah di-update:
- Trigger `trg_refresh_account_balances` sudah di-disable (commented out)
- Function `refresh_account_balances()` hanya untuk dokumentasi
- Fresh install tidak akan mengalami masalah ini

## Verification

### Check Balance Sync Working
```sql
-- Insert deposit transaction
-- Check accounts.balance langsung updated
SELECT id, name, balance FROM accounts WHERE id = 40;

-- Check materialized view (mungkin belum updated)
SELECT account_name, current_balance FROM account_balances WHERE account_id = 40;

-- Manual refresh
SELECT * FROM manual_refresh_account_balances();

-- Check again - sekarang sudah sync
SELECT account_name, current_balance FROM account_balances WHERE account_id = 40;
```

### Monitor Scheduler (if using Go scheduler)
```bash
# Check logs
tail -f scheduler.log

# Expected output:
# 🔗 Connected to database: accounting_db
# ⏰ Refresh interval: 1h0m0s
# 🚀 Starting materialized view refresh scheduler...
# 🔍 Checking materialized view freshness...
# 📊 View age: 5 minutes (last updated: 2025-10-26T10:40:00+07:00)
# ✅ View is fresh, no refresh needed
```

## Architecture

```
Transaction Flow:
┌─────────────────┐
│ Cash Bank API   │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│ Create SSOT Journal Entry               │
│ - unified_journal_ledger (DRAFT)        │
│ - unified_journal_lines                 │
│ - Update status to POSTED               │
└────────┬────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│ Trigger: trg_sync_account_balance      │  ✅ Real-time
│ - Update accounts.balance per account   │
└─────────────────────────────────────────┘
         
         
Reporting Flow:
┌─────────────────┐
│ Financial Report│
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│ Query: account_balances (MV)            │  📊 Cached
└────────┬────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────┐
│ If stale: manual_refresh()              │  ⏰ Scheduled
│ - Via API call                          │
│ - Via scheduled job                     │
│ - Via cron                              │
└─────────────────────────────────────────┘
```

## Benefits

1. ✅ **No more concurrent refresh errors** - Trigger dihapus
2. ✅ **Real-time balance sync** - Via per-account update triggers
3. ✅ **Better performance** - Materialized view refresh on-demand
4. ✅ **Scalability** - No locking conflicts di high-concurrency environments
5. ✅ **Control** - Manual control kapan refresh dilakukan
6. ✅ **Monitoring** - Freshness check untuk mengetahui status view

## Rollback (if needed)

Jika ingin kembali ke auto-refresh (NOT RECOMMENDED):
```sql
-- Enable trigger kembali (akan menyebabkan concurrent errors)
CREATE TRIGGER trg_refresh_account_balances
    AFTER INSERT OR UPDATE OR DELETE ON unified_journal_lines
    FOR EACH STATEMENT
    EXECUTE FUNCTION refresh_account_balances();
```

## Support

Jika masih ada masalah:
1. Check trigger list: `SELECT * FROM pg_trigger WHERE tgrelid = 'unified_journal_lines'::regclass;`
2. Check function exists: `SELECT * FROM pg_proc WHERE proname LIKE '%refresh%';`
3. Test manual refresh: `SELECT * FROM manual_refresh_account_balances();`
4. Check balance sync trigger: `SELECT * FROM pg_trigger WHERE tgname LIKE '%sync_account_balance%';`
