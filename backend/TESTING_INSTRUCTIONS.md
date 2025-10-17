# 🧪 Testing Instructions - Revenue Duplication

## 📊 Current Status

| Metric | Expected | Actual | Status |
|--------|----------|--------|--------|
| Account Balance (4101) | Rp 10,000,000 | - | ✅ Correct |
| P&L Report Revenue | Rp 10,000,000 | Rp 20,000,000 | ❌ Duplicated |
| Discrepancy | - | Rp 10,000,000 | 🚨 100% error! |

---

## 🎯 Critical Query to Run

**This is the MOST IMPORTANT query to identify the issue:**

```sql
SELECT 
    je.account_code,
    je.account_name,
    SUM(je.credit - je.debit) as amount,
    COUNT(*) as entry_count
FROM journal_entries je
INNER JOIN journals j ON je.journal_id = j.id
WHERE je.account_code LIKE '4%'
  AND j.status = 'POSTED'
  AND j.date BETWEEN '2025-01-01' AND '2025-12-31'
GROUP BY je.account_code, je.account_name
ORDER BY je.account_code, je.account_name;
```

### Expected Results (if no duplication):
```
account_code | account_name           | amount     | entry_count
-------------|------------------------|------------|------------
4101         | PENDAPATAN PENJUALAN   | 10000000   | X
```

### Actual Results (if duplication exists):
```
account_code | account_name           | amount     | entry_count
-------------|------------------------|------------|------------
4101         | PENDAPATAN PENJUALAN   | 10000000   | X
4101         | Pendapatan Penjualan   | 10000000   | Y
                                      -----------
                                      20000000  ← PROBLEM!
```

---

## 🔍 How to Run Tests

### Method 1: Run Simple PowerShell Script ✅ EASIEST
```powershell
cd backend
.\test_revenue_simple.ps1
```
This shows you all the queries to run.

### Method 2: Direct Database Query
1. Open **HeidiSQL**, **phpMyAdmin**, or **MySQL Workbench**
2. Connect to database: `accounting_db`
3. Copy and paste the critical query above
4. Run it
5. Take screenshot of results

### Method 3: MySQL Command Line
```bash
mysql -u root -p accounting_db
# Then paste and run the query
```

---

## 🎯 What to Check

### ✅ Things to Verify

1. **Number of Rows Returned**
   - ✅ **1 row** for account 4101 = Good (no duplication)
   - ❌ **2+ rows** for account 4101 = PROBLEM FOUND!

2. **Account Name Variations**
   - Check if `account_name` column has variations:
     - "PENDAPATAN PENJUALAN" (uppercase)
     - "Pendapatan Penjualan" (mixed case)
     - "pendapatan penjualan" (lowercase)
   - **Different names = Backend is grouping separately**

3. **Total Amount**
   - Sum of all amounts should be **Rp 10,000,000**
   - If total is **Rp 20,000,000** = duplication confirmed

---

## 🔧 Quick Fixes Based on Results

### Scenario 1: Multiple Rows with Different Names
**Problem**: Backend groups by `account_code` AND `account_name`

**Fix Option A - Database Update**:
```sql
-- Standardize account names to match accounts table
UPDATE journal_entries je
INNER JOIN accounts a ON a.code = je.account_code
SET je.account_name = a.name
WHERE je.account_code = '4101';
```

**Fix Option B - Backend Update**:
```go
// File: backend/services/ssot_profit_loss_service.go
// Change GROUP BY to use only account_code, not account_name
```

### Scenario 2: Both SSOT and Legacy Systems Active
**Problem**: Both `unified_journal_lines` and `journal_lines` are counted

**Fix**: Update backend fallback logic (see REVENUE_FIX_QUICK_START.md)

---

## 📝 Test Results Template

After running the critical query, fill this out:

```
=== TEST RESULTS ===
Date: _____________
Tester: _____________

QUERY RESULTS:
- Number of rows returned: _______
- Account codes found: _______
- Account names found: _______
- Total amount: Rp _______

SCREENSHOTS:
[ ] Query result screenshot attached
[ ] P&L report screenshot attached

DIAGNOSIS:
[ ] Multiple rows for same account_code (different account_name)
[ ] SSOT and Legacy both have data
[ ] Duplicate journal entries
[ ] Other: _______

NEXT ACTION:
[ ] Apply database fix
[ ] Apply backend fix
[ ] Need more investigation
```

---

## 🚀 After Fix - Verification

Run these to verify the fix worked:

### 1. Re-run Critical Query
Should return **1 row** with **Rp 10,000,000**

### 2. Test P&L API
```powershell
cd backend
.\test_profit_loss_summary_fix.ps1
```

### 3. Test Frontend
1. Navigate to `http://localhost:3000/reports`
2. Generate Profit & Loss Report
3. Verify **Total Revenue = Rp 10,000,000** ✅

---

## 📞 Support

If still duplicated after fixes:

1. Run all queries in `investigate_revenue_duplication.sql`
2. Check `unified_journal_lines` vs `journal_lines` tables
3. Review backend logs for "SSOT" vs "LEGACY" data source
4. Share screenshots from critical query

---

## 📁 Related Files

- `test_revenue_simple.ps1` - Simple testing guide
- `test_profit_loss_database_validation.ps1` - Full API testing
- `debug_revenue_duplication.go` - Database analysis tool
- `investigate_revenue_duplication.sql` - Complete SQL queries
- `REVENUE_FIX_QUICK_START.md` - Quick fix guide

---

**Created**: 2025-10-17  
**Priority**: 🔴 HIGH (100% revenue discrepancy)  
**Status**: Ready for Testing

