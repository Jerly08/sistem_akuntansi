# Emergency Fix Scripts - Trigger Removal

Jika masih mendapat error `SQLSTATE 55000` setelah restart backend, gunakan salah satu script berikut untuk manual fix.

## 🚨 Option 1: Quick Fix (Recommended)

**Paling cepat dan simple - hanya 1 command**

### Via psql:
```bash
psql -U postgres -d accounting_db -f backend/quick_fix_trigger.sql
```

### Via pgAdmin / DBeaver / Postico:
Copy paste dan execute:
```sql
DROP TRIGGER IF EXISTS trg_refresh_account_balances ON unified_journal_lines CASCADE;
```

---

## 🔧 Option 2: Emergency Script dengan Verification

**Lengkap dengan testing dan verification**

### Via psql:
```bash
psql -U postgres -d accounting_db -f backend/emergency_fix_trigger.sql
```

### Via pgAdmin / DBeaver:
1. Open `backend/emergency_fix_trigger.sql`
2. Copy all content
3. Execute di query window

Output yang diharapkan:
```
🔧 Starting emergency trigger fix...
📋 Current triggers on unified_journal_lines:
...
🗑️  Dropping trg_refresh_account_balances...
✅ Verifying trigger removal...
✅ Trigger successfully removed
🔧 Creating manual refresh helper functions...
🧪 Testing manual refresh function...
success | message                                  | refreshed_at
--------+------------------------------------------+-------------------------
true    | Account balances refreshed in 00:00:01.2 | 2025-10-26 11:00:00+07
✅ EMERGENCY FIX COMPLETED
```

---

## 🛠️ Option 3: Go Script (Tanpa Restart Backend)

**Jalankan langsung dari command line tanpa perlu restart backend**

### Setup:
```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=accounting_db
```

### Windows (PowerShell):
```powershell
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="postgres"
$env:DB_PASSWORD="your_password"
$env:DB_NAME="accounting_db"

go run backend/cmd/scripts/emergency_remove_trigger.go
```

### Linux/Mac:
```bash
DB_HOST=localhost \
DB_PORT=5432 \
DB_USER=postgres \
DB_PASSWORD=your_password \
DB_NAME=accounting_db \
go run backend/cmd/scripts/emergency_remove_trigger.go
```

Output:
```
🚨 EMERGENCY TRIGGER REMOVAL SCRIPT
===================================
✅ Connected to database: accounting_db

🔍 Checking for problematic trigger...
⚠️  Found problematic trigger: trg_refresh_account_balances

🗑️  Removing trigger...
✅ Trigger removed successfully

🔍 Verifying removal...
✅ Verification passed - trigger is gone

🔧 Creating helper functions...
✅ Created function: manual_refresh_account_balances()
✅ Created function: check_account_balances_freshness()

=========================================
✅ EMERGENCY FIX COMPLETED SUCCESSFULLY
=========================================

What was done:
  ✅ Removed trigger: trg_refresh_account_balances
  ✅ Created helper functions for manual refresh

Next steps:
  1. Test your transactions (deposit, sales, etc.)
  2. Error SQLSTATE 55000 should be gone
  3. No need to restart backend

For manual refresh:
  SELECT * FROM manual_refresh_account_balances();
```

---

## ✅ Verification

Setelah run salah satu script di atas, verify dengan:

### 1. Check trigger sudah dihapus:
```sql
SELECT tgname 
FROM pg_trigger 
WHERE tgname = 'trg_refresh_account_balances';
```
Hasil harusnya **kosong** (0 rows)

### 2. Check helper functions ada:
```sql
SELECT proname 
FROM pg_proc 
WHERE proname IN ('manual_refresh_account_balances', 'check_account_balances_freshness');
```
Hasil harusnya ada **2 functions**

### 3. Test manual refresh:
```sql
SELECT * FROM manual_refresh_account_balances();
```

### 4. Test freshness check:
```sql
SELECT * FROM check_account_balances_freshness();
```

---

## 🧪 Testing

Setelah fix, test dengan membuat transaksi:

```bash
# Test deposit
curl -X POST http://localhost:8080/api/v1/cash-bank/transactions/deposit \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": 1,
    "amount": 100000,
    "date": "2025-10-26",
    "description": "Test deposit after fix"
  }'

# Harusnya tidak ada error SQLSTATE 55000
```

---

## 🆘 Troubleshooting

### Error: "relation unified_journal_lines does not exist"
- Database belum running migration SSOT
- Run: `backend/migrations/020_create_unified_journal_ssot.sql`

### Error: "materialized view account_balances does not exist"
- Materialized view belum dibuat
- Run: `REFRESH MATERIALIZED VIEW account_balances;` (akan error jika belum ada)
- Check migration 020 sudah dijalankan

### Script berhasil tapi masih error SQLSTATE 55000
1. Restart backend application
2. Check auto-fix di startup log
3. Verify trigger benar-benar sudah dihapus (query di atas)

---

## 📝 Summary Comparison

| Method | Speed | Verification | Restart Required | Difficulty |
|--------|-------|--------------|------------------|------------|
| Quick Fix SQL | ⚡⚡⚡ | ❌ | ✅ | Easy |
| Emergency SQL | ⚡⚡ | ✅ | ✅ | Easy |
| Go Script | ⚡⚡ | ✅ | ❌ | Medium |
| Auto-fix (startup) | ⚡ | ✅ | ✅ | Easy |

**Recommendation:**
- **First time error**: Use **Quick Fix SQL** 
- **Need verification**: Use **Emergency SQL**
- **Production without downtime**: Use **Go Script**
- **Long term**: Let **Auto-fix** handle it (already implemented)

---

## 📚 Related Files

- `backend/fix_concurrent_refresh_error.sql` - Original comprehensive fix script
- `backend/database/fix_concurrent_refresh.go` - Auto-fix on backend startup
- `backend/FIX_CONCURRENT_REFRESH_README.md` - Detailed documentation
- `backend/migrations/020_create_unified_journal_ssot.sql` - Migration file (trigger disabled)
