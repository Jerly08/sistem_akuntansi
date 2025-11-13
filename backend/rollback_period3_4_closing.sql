-- Rollback incorrect closing entries for periods 3 & 4
-- These closing entries were created with fixed amounts instead of cumulative totals

BEGIN;

-- 1. Find the incorrect closing journal entries for period 3 (2027-02-02) and 4 (2027-12-31)
SELECT 
    id, 
    entry_date, 
    description, 
    total_debit, 
    total_credit,
    source_type
FROM unified_journal_ledger
WHERE source_type = 'CLOSING' 
    AND entry_date IN ('2027-02-02', '2027-12-31')
ORDER BY entry_date;

-- 2. Store the IDs for reference
-- Period 3: 2027-02-02
-- Period 4: 2027-12-31

-- 3. Get the journal lines before deletion (for audit)
SELECT 
    ujl.id as line_id,
    uje.entry_date,
    uje.description as journal_desc,
    a.code,
    a.name,
    a.type,
    ujl.debit_amount,
    ujl.credit_amount,
    ujl.description as line_desc
FROM unified_journal_lines ujl
INNER JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
INNER JOIN accounts a ON a.id = ujl.account_id
WHERE uje.source_type = 'CLOSING' 
    AND uje.entry_date IN ('2027-02-02', '2027-12-31')
ORDER BY uje.entry_date, ujl.line_number;

-- 4. Before deletion, calculate the current balance impact
-- Get account balances that will be affected
SELECT 
    a.code,
    a.name,
    a.type,
    a.balance as current_balance
FROM accounts a
WHERE a.code IN ('3201', '4101', '5101')
ORDER BY a.code;

-- 5. Delete the journal lines first (child records)
DELETE FROM unified_journal_lines
WHERE journal_id IN (
    SELECT id 
    FROM unified_journal_ledger 
    WHERE source_type = 'CLOSING' 
        AND entry_date IN ('2027-02-02', '2027-12-31')
);

-- 6. Delete the journal entries (parent records)
DELETE FROM unified_journal_ledger
WHERE source_type = 'CLOSING' 
    AND entry_date IN ('2027-02-02', '2027-12-31');

-- 7. Delete the accounting period records
DELETE FROM accounting_periods
WHERE end_date IN ('2027-02-02', '2027-12-31')
    AND is_closed = true;

-- 8. Restore account balances to their pre-closing state
-- We need to reverse the closing entries impact
-- For period 3 closing (2027-02-02):
--   Revenue (4101): was 7M closed, should restore to cumulative 14M
--   Expense (5101): was 3.5M closed, should restore to cumulative 7M
--   Retained Earnings (3201): remove incorrect 3.5M net income

-- For period 4 closing (2027-12-31):
--   Revenue (4101): was 7M closed, should restore to cumulative 21M
--   Expense (5101): was 3.5M closed, should restore to cumulative 10.5M
--   Retained Earnings (3201): remove incorrect 3.5M net income

-- IMPORTANT: We need to recalculate from scratch
-- Option 1: Manually set balances
-- Option 2: Let the system recalculate when we regenerate closing entries

-- For now, let's reset to state before period 3 closing
-- After period 2 (2026-12-31) closed:
--   Revenue (4101): should be 0 (was closed)
--   Expense (5101): should be 0 (was closed)
--   Retained Earnings (3201): should be 7M (net income from period 1+2)

-- But we have transactions after period 2:
-- Period 3 transactions: +7M revenue, +3.5M expense
-- Period 4 transactions: +7M revenue, +3.5M expense

-- So current unclosed state should be:
UPDATE accounts SET balance = 0 WHERE code = '4101'; -- Revenue will be recalculated from journal lines
UPDATE accounts SET balance = 0 WHERE code = '5101'; -- Expense will be recalculated from journal lines
UPDATE accounts SET balance = 7000000 WHERE code = '3201'; -- Retained Earnings from period 1+2

-- Actually, better approach: recalculate balances from ALL journal lines except closing entries
-- This is complex in SQL, better to let the application handle it

-- 9. Verify deletion
SELECT 
    'Closing journals after deletion' as check_point,
    COUNT(*) as count
FROM unified_journal_ledger
WHERE source_type = 'CLOSING' 
    AND entry_date IN ('2027-02-02', '2027-12-31');

SELECT 
    'Accounting periods after deletion' as check_point,
    COUNT(*) as count
FROM accounting_periods
WHERE end_date IN ('2027-02-02', '2027-12-31')
    AND is_closed = true;

-- 10. Show remaining closed periods
SELECT 
    id,
    start_date,
    end_date,
    description,
    total_revenue,
    total_expense,
    net_income,
    is_closed
FROM accounting_periods
WHERE is_closed = true
ORDER BY end_date;

-- COMMIT only if everything looks correct
-- Otherwise ROLLBACK
ROLLBACK; -- Change to COMMIT when ready
