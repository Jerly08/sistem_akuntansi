# Fix Current Assets Dashboard Issue

## Problem Description

Dashboard admin menampilkan nilai "CURRENT ASSETS" sebesar Rp 555.000 pada chart pie, namun:
1. Di Chart of Accounts (COA), balance current assets tidak tampil dengan nilai yang benar
2. Current assets detail accounts (kas, bank, dll) memiliki balance 0
3. Tidak ada transaksi purchase/sales yang berkontribusi pada balance tersebut

## Root Cause Analysis

**MASALAH DITEMUKAN**: Account header 1100 "CURRENT ASSETS" memiliki balance 555.000, padahal ini adalah account header yang seharusnya tidak memiliki balance langsung.

### Detail Temuan:
- Account 1100 (IsHeader: true) memiliki balance 555.000
- Semua current assets detail accounts (1101, 1102, 1103, 1105, 1106) memiliki balance 0
- Account header tidak memiliki journal entries apapun
- Query dashboard `getTopAccounts` mengambil semua accounts dengan balance != 0, termasuk header accounts

## Solutions Implemented

### 1. Fix Header Account Balance
```go
// Set balance account header 1100 menjadi 0
UPDATE accounts SET balance = 0 WHERE code = '1100' AND is_header = true;
```

### 2. Update Dashboard Queries
Modified queries in:
- `services/dashboard_service.go` - getTopAccounts()
- `controllers/dashboard_controller.go` - getTopAccounts()

Added filter `AND is_header = false` to exclude header accounts:
```sql
SELECT 
    name,
    ABS(balance) as balance,
    type
FROM accounts 
WHERE deleted_at IS NULL 
    AND is_active = true
    AND balance != 0
    AND is_header = false  -- ✅ Added this line
ORDER BY ABS(balance) DESC
LIMIT 5
```

### 3. Prevention Mechanism
Created database trigger to prevent header accounts from having non-zero balance:
```sql
CREATE OR REPLACE FUNCTION prevent_header_account_balance()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_header = true AND NEW.balance != 0 THEN
        RAISE EXCEPTION 'Header accounts cannot have non-zero balance. Account: % (%), Balance: %', 
            NEW.code, NEW.name, NEW.balance;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_header_balance
    BEFORE INSERT OR UPDATE ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION prevent_header_account_balance();
```

## Verification Results

**Before Fix:**
- Account 1100: Balance 555.000 (INCORRECT)
- Dashboard shows "CURRENT ASSETS: 555.000"
- COA shows 0 balance for current assets detail accounts

**After Fix:**
- Account 1100: Balance 0.00 (CORRECT)
- Dashboard: No accounts with non-zero balance found
- COA: Consistent with dashboard
- Total Current Assets: 0.00 (calculated correctly from detail accounts)

## Files Modified

1. **backend/services/dashboard_service.go**
   - Added `AND is_header = false` to getTopAccounts query

2. **backend/controllers/dashboard_controller.go** 
   - Added `AND is_header = false` to getTopAccounts query

## Scripts Created

1. **scripts/debug_current_assets.go** - For debugging current assets calculation
2. **scripts/fix_header_account_balance.go** - For fixing header account balances
3. **scripts/prevent_header_balance.go** - For creating prevention trigger

## Best Practices Established

1. **Header accounts should never have direct balance** - Balance should be calculated from child accounts
2. **Dashboard queries must exclude header accounts** - Only detail accounts should appear in financial summaries
3. **Database constraints prevent future issues** - Trigger ensures header accounts can't have balance
4. **Proper account hierarchy** - Header accounts (IsHeader: true) vs Detail accounts (IsHeader: false)

## Testing

Run debug script to verify fix:
```bash
cd backend
go run scripts/debug_current_assets.go
```

Expected output:
- All header accounts have balance 0.00
- Total Current Assets calculated correctly from detail accounts
- Dashboard top accounts exclude header accounts

## Impact

✅ **RESOLVED**: Current assets balance consistency between dashboard and COA
✅ **PREVENTED**: Future occurrences of header accounts having direct balance  
✅ **IMPROVED**: Data integrity and accounting accuracy
