# 🔧 Fix Guide: Concurrent Materialized View Refresh Error

## Problem Description
Error yang terjadi saat create invoice dan deposit cash & bank:
```
ERROR: cannot refresh materialized view "public.account_balances" concurrently (SQLSTATE 55000)
```

## Root Cause
Trigger `trg_refresh_account_balances` mencoba melakukan refresh materialized view secara concurrent pada setiap insert ke `unified_journal_lines`, yang menyebabkan konflik saat multiple transactions berjalan bersamaan.

## Solution Steps

### Step 1: Install UUID Extension (Required)
UUID extension diperlukan untuk beberapa migration. Jalankan script ini:

```bash
go run add_uuid_extension.go
```

Atau manual via SQL:
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

### Step 2: Fix Concurrent Refresh Error
Jalankan script untuk menghapus trigger yang bermasalah:

```bash
go run fix_concurrent_refresh_error.go
```

Script ini akan:
1. Menghapus trigger `trg_refresh_account_balances` 
2. Membuat function `manual_refresh_account_balances()` untuk refresh manual
3. Membuat function `check_account_balances_freshness()` untuk cek status

### Step 3: Verify the Fix
Test apakah error sudah teratasi:

1. **Test Deposit Cash & Bank:**
   - Login ke aplikasi
   - Buka menu Cash & Bank
   - Buat transaksi deposit
   - Seharusnya berhasil tanpa error

2. **Test Sales Invoice:**
   - Buka menu Sales
   - Create invoice untuk sale yang ada
   - Seharusnya berhasil tanpa error

### Step 4: Optional - Setup Scheduled Refresh
Untuk reporting yang selalu up-to-date, setup scheduled refresh:

**Option A: Cron Job (Linux/Mac):**
```bash
# Add to crontab -e
0 * * * * psql -U postgres -d sistem_akuntansi -c "SELECT * FROM manual_refresh_account_balances();"
```

**Option B: Windows Task Scheduler:**
Create scheduled task yang run command:
```powershell
psql -U postgres -d sistem_akuntansi -c "SELECT * FROM manual_refresh_account_balances();"
```

**Option C: Backend Service (Recommended):**
Add ke backend untuk auto refresh setiap 1 jam:

```go
// Add to main.go or separate service
func scheduleAccountBalanceRefresh(db *gorm.DB) {
    ticker := time.NewTicker(1 * time.Hour)
    go func() {
        for range ticker.C {
            var result struct {
                Success bool
                Message string
            }
            db.Raw("SELECT success, message FROM manual_refresh_account_balances()").Scan(&result)
            log.Printf("Account balance refresh: %v - %s", result.Success, result.Message)
        }
    }()
}
```

## What Changed

### Before (Problematic):
- Trigger `trg_refresh_account_balances` runs on EVERY insert to `unified_journal_lines`
- Causes concurrent refresh conflicts when multiple transactions happen
- Results in SQLSTATE 55000 error

### After (Fixed):
- No automatic trigger on journal line inserts
- Real-time balance sync handled by existing `setup_automatic_balance_sync.sql` triggers
- Materialized view refreshed manually or via schedule
- No more concurrent refresh errors

## Technical Details

### Balance Synchronization Flow:
1. **Real-time sync** (Always accurate):
   - `accounts.balance` field updated via triggers
   - Used by Cash & Bank Management page
   - No materialized view needed

2. **Reporting view** (Refreshed periodically):
   - `account_balances` materialized view
   - Used for complex reports
   - Refresh manually or scheduled

### Functions Available:

```sql
-- Manual refresh (run anytime)
SELECT * FROM manual_refresh_account_balances();

-- Check if refresh needed
SELECT * FROM check_account_balances_freshness();
```

## Troubleshooting

### If error still occurs:
1. Check if trigger was successfully removed:
```sql
SELECT tgname FROM pg_trigger WHERE tgname = 'trg_refresh_account_balances';
```

2. If trigger still exists, manually drop it:
```sql
DROP TRIGGER IF EXISTS trg_refresh_account_balances ON unified_journal_lines;
```

3. Verify functions exist:
```sql
\df manual_refresh_account_balances
\df check_account_balances_freshness
```

### If UUID error occurs:
Run:
```bash
go run add_uuid_extension.go
```

### Port 8080 already in use:
Windows:
```powershell
netstat -ano | findstr :8080
taskkill /F /PID <PID_NUMBER>
```

Linux/Mac:
```bash
lsof -i :8080
kill -9 <PID_NUMBER>
```

## Verification Checklist

✅ UUID extension installed  
✅ Trigger `trg_refresh_account_balances` removed  
✅ Functions `manual_refresh_account_balances` and `check_account_balances_freshness` created  
✅ Deposit Cash & Bank works without error  
✅ Sales Invoice creation works without error  
✅ Account balances show correctly in COA  

## Notes for Production

1. **Database Backup**: Always backup database before applying fixes
2. **Testing**: Test in staging environment first
3. **Monitoring**: Monitor for any balance discrepancies after fix
4. **Schedule**: Set up refresh schedule based on reporting needs (hourly recommended)

## Contact

If issues persist after following this guide, check:
- PostgreSQL version (should be 12+)
- Database permissions
- Network connectivity between backend and database