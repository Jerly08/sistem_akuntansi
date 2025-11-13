# Fix Period Closing Issue

## Problem
Period closing fails with error: **"closing validation failed: 1 revenue/expense accounts still have non-zero balance"**

## Root Cause
The balance calculation formula was using incorrect sign convention. Revenue accounts were showing **POSITIVE** balance instead of **NEGATIVE** (credit balance).

## Solution

### Step 1: Pull Latest Code
```bash
cd /path/to/accounting_proj
git pull origin main
```

### Step 2: Run Verification & Fix Script
```bash
cd backend
go run cmd/verify_and_fix_pc.go
```

This script will:
1. ✅ Check if balances have wrong signs
2. ✅ Remove corrupt closing entries
3. ✅ Recalculate all balances with correct formula
4. ✅ Verify the fix

### Step 3: Restart Backend
Stop the backend server (Ctrl+C) and restart:
```bash
go run main.go
```

### Step 4: Try Period Closing Again
Go to the frontend and execute period closing. It should now succeed!

## Expected Results

### Before Fix
```
PENDAPATAN PENJUALAN (4101): +59,400,000  ❌ WRONG (positive)
```

### After Fix
```
PENDAPATAN PENJUALAN (4101): -59,400,000  ✅ CORRECT (negative)
```

## Manual Verification

You can manually verify balances using SQL:
```sql
SELECT code, name, type, balance
FROM accounts
WHERE type IN ('REVENUE', 'EXPENSE')
  AND is_header = false
  AND deleted_at IS NULL
ORDER BY type, code;
```

**Expected:**
- REVENUE accounts: **negative** balance (credit)
- EXPENSE accounts: **positive** balance (debit)

## Technical Details

### What Changed
1. **Balance recalculation query** in `unified_period_closing_service.go`:
   - OLD: Used different formulas for different account types
   - NEW: Uses `Debit - Credit` for ALL account types

2. **Sign convention**:
   - ASSET/EXPENSE: Positive = normal debit balance
   - LIABILITY/EQUITY/REVENUE: Negative = normal credit balance

### Files Modified
- `backend/services/unified_period_closing_service.go`
- Added cleanup scripts in `backend/cmd/`

## Troubleshooting

### Issue: "Still getting error after fix"
**Solution:** Make sure you pulled latest code AND restarted backend server

### Issue: "Balance still has wrong sign"
**Solution:** Run the fix script again:
```bash
go run cmd/verify_and_fix_pc.go
```

### Issue: "Database connection error"
**Solution:** Check your DATABASE_URL environment variable or .env file

## Contact
If you still have issues, check the Git commit history for the fix:
```bash
git log --oneline -5
```

Look for: "Fix period closing balance recalculation and add cleanup scripts"
