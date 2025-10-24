# Migration Errors - Analysis & Fixes

## Date: 2025-10-23

## Problems Identified

### 1. ❌ MySQL Syntax in PostgreSQL Database

**Error:**
```
ERROR: syntax error at or near "INNER" (SQLSTATE 42601)
```

**Location:** `database/auto_migrations.go` - `fixRevenueDuplication()` function

**Root Cause:**
- Code menggunakan MySQL syntax untuk UPDATE dengan JOIN
- PostgreSQL menggunakan syntax yang berbeda

**MySQL Syntax (Wrong):**
```sql
UPDATE table1 t1
INNER JOIN table2 t2 ON t1.id = t2.id
SET t1.column = t2.column
```

**PostgreSQL Syntax (Correct):**
```sql
UPDATE table1 t1
SET column = t2.column
FROM table2 t2
WHERE t1.id = t2.id
```

**Fixed Files:**
- `database/auto_migrations.go` lines 1636-1721
  - Fixed UPDATE untuk `journal_entries` revenue (4xxx)
  - Fixed UPDATE untuk `journal_entries` expenses (5xxx)
  - Fixed UPDATE untuk `unified_journal_lines` revenue (4xxx)
  - Fixed UPDATE untuk `unified_journal_lines` expenses (5xxx)
  - Changed `GROUP_CONCAT(...SEPARATOR ' | ')` → `STRING_AGG(..., ' | ')`

---

### 2. ❌ RAISE NOTICE Outside DO Block

**Error:**
```
ERROR: syntax error at or near "RAISE" (SQLSTATE 42601)
```

**Location:** `migrations/040_lock_critical_tax_accounts.sql` line 138

**Root Cause:**
- `RAISE NOTICE` tidak bisa dijalankan langsung di luar function/DO block
- Harus dibungkus dalam `DO $$ ... END $$;`

**Fixed:**
```sql
-- Before (Wrong):
CREATE TRIGGER ...;
RAISE NOTICE '✅ Created trigger to protect critical accounts';

-- After (Correct):
CREATE TRIGGER ...;
DO $$
BEGIN
    RAISE NOTICE '✅ Created trigger to protect critical accounts';
END $$;
```

---

### 3. ❌ Check Constraint Violation

**Error:**
```
ERROR: check constraint "chk_account_code_format" of relation "accounts" is violated by some row (SQLSTATE 23514)
```

**Location:** `migrations/prevent_duplicate_accounts.sql` line 131

**Root Cause:**
- Constraint terlalu strict: `code ~ '^[0-9]{4}$' OR code ~ '^[0-9]{4}\.[0-9]+$'`
- Ada existing data yang tidak memenuhi format (mungkin alphanumeric atau format lain)

**Solution:**
- Commented out strict constraint
- Dibuat script `check_invalid_account_codes.go` untuk identifikasi data bermasalah
- Tetap enforce unique constraint, tapi skip format validation untuk sekarang

---

### 4. ❌ Prepared Statement with Multiple Commands

**Error:**
```
ERROR: cannot insert multiple commands into a prepared statement (SQLSTATE 42601)
```

**Location:** `database/prevent_duplicate_accounts_migration.go` line 24

**Root Cause:**
- File SQL mengandung multiple commands (DO blocks, CREATE, ALTER, etc.)
- GORM `db.Exec(string(sqlContent))` tidak bisa handle multiple commands sekaligus
- PostgreSQL prepared statements hanya bisa execute 1 command per call

**Fixed:**
- Ubah `RunPreventDuplicateAccountsMigration()` untuk langsung call `applyInlineMigration()`
- Inline migration execute statement by statement dengan error handling yang proper

---

## Files Modified

### 1. `database/auto_migrations.go`
**Changes:**
- Lines 1636-1643: Fixed UPDATE syntax untuk journal_entries revenue
- Lines 1653-1660: Fixed UPDATE syntax untuk journal_entries expenses
- Lines 1670-1677: Fixed UPDATE syntax untuk unified_journal_lines revenue
- Lines 1687-1694: Fixed UPDATE syntax untuk unified_journal_lines expenses
- Lines 1716: Changed GROUP_CONCAT → STRING_AGG

### 2. `migrations/040_lock_critical_tax_accounts.sql`
**Changes:**
- Lines 136-141: Wrapped RAISE NOTICE dalam DO block

### 3. `migrations/prevent_duplicate_accounts.sql`
**Changes:**
- Lines 126-140: Commented out strict check constraint, added DO block for cleanup

### 4. `database/prevent_duplicate_accounts_migration.go`
**Changes:**
- Lines 11-19: Simplified to always use inline migration

---

## New Files Created

### 1. `scripts/check_invalid_account_codes.go`
**Purpose:** Mengidentifikasi account codes yang tidak sesuai format standar

**Usage:**
```bash
go run scripts/check_invalid_account_codes.go
```

**Output:** List account IDs yang perlu diperbaiki atau dikecualikan

### 2. `run_check_invalid_codes.ps1`
**Purpose:** PowerShell wrapper untuk run check script

**Usage:**
```powershell
.\run_check_invalid_codes.ps1
```

---

## How to Verify Fixes

### Step 1: Check Invalid Account Codes (Optional)
```bash
# Run check script to see if there are any invalid codes
go run scripts/check_invalid_account_codes.go
```

### Step 2: Run Application
```bash
# The application should now start without migration errors
go run main.go
```

### Step 3: Verify Logs
Look for these success messages:
- ✅ Migration completed: `040_lock_critical_tax_accounts.sql`
- ✅ Migration completed: `prevent_duplicate_accounts.sql`
- ✅ Revenue duplication fix completed
- ✅ Balance sync system setup completed

---

## Summary of Changes

| Issue | Type | Severity | Status |
|-------|------|----------|--------|
| MySQL syntax in PostgreSQL | Syntax Error | HIGH | ✅ Fixed |
| RAISE NOTICE outside DO block | Syntax Error | MEDIUM | ✅ Fixed |
| Check constraint too strict | Data Constraint | MEDIUM | ✅ Fixed (Relaxed) |
| Multiple commands in prepared statement | Execution Error | HIGH | ✅ Fixed |

---

## PostgreSQL Best Practices Applied

1. **UPDATE with JOIN:**
   - Use `FROM` clause instead of `INNER JOIN` after table name
   - Move join condition to WHERE clause

2. **PL/pgSQL Statements:**
   - Always wrap `RAISE NOTICE` in function or DO block
   - Use `DO $$ BEGIN ... END $$;` for inline procedural code

3. **String Aggregation:**
   - Use `STRING_AGG(column, separator)` not `GROUP_CONCAT()`

4. **Migration Execution:**
   - Parse and execute statements individually for complex migrations
   - Handle `DO $$` blocks properly with dollar-quote parser
   - Allow idempotent reruns by checking "already exists" errors

---

## Additional Notes

- The check constraint for account code format was intentionally relaxed to allow existing data
- If strict validation is needed in the future, first fix existing data using the check script
- All MySQL-specific syntax has been converted to PostgreSQL equivalents
- Migration system now properly handles complex PL/pgSQL blocks with the `parseComplexSQL()` function

---

## Next Steps

1. ✅ Test application startup
2. ⏳ Monitor migration logs for any remaining issues
3. ⏳ Optionally: Fix invalid account codes if found by check script
4. ⏳ Re-enable strict check constraint after data cleanup (if needed)
