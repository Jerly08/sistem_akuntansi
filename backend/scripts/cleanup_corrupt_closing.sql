-- ============================================
-- Cleanup Script for Corrupt Period Closing
-- ============================================
-- This script removes failed/corrupt closing entries and recalculates account balances

BEGIN;

-- 1. Delete failed accounting periods (yang error pas closing)
DELETE FROM accounting_periods 
WHERE is_closed = true 
  AND closed_at >= '2025-11-13 00:00:00';

-- 2. Delete corrupt closing journal entries
DELETE FROM unified_journal_lines 
WHERE journal_id IN (
    SELECT id FROM unified_journal_ledger 
    WHERE source_type = 'CLOSING' 
      AND created_at >= '2025-11-13 00:00:00'
);

DELETE FROM unified_journal_ledger 
WHERE source_type = 'CLOSING' 
  AND created_at >= '2025-11-13 00:00:00';

-- 3. Recalculate ALL account balances from POSTED journals (SSOT)
UPDATE accounts a
SET balance = COALESCE((
    SELECT 
        CASE 
            WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
            ELSE 
                COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
        END
    FROM unified_journal_lines ujl
    INNER JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
    WHERE ujl.account_id = a.id 
      AND uje.status = 'POSTED'
), 0)
WHERE a.deleted_at IS NULL;

-- 4. Verify: Check revenue and expense accounts (should have non-zero before closing)
SELECT 
    code,
    name,
    type,
    balance,
    CASE 
        WHEN type = 'REVENUE' AND balance < 0 THEN 'OK - Credit balance'
        WHEN type = 'EXPENSE' AND balance > 0 THEN 'OK - Debit balance'
        WHEN type IN ('REVENUE', 'EXPENSE') AND ABS(balance) < 0.01 THEN 'ZERO - Ready for closing'
        ELSE '⚠️ WARNING - Unusual balance'
    END as status
FROM accounts
WHERE type IN ('REVENUE', 'EXPENSE')
  AND is_header = false
  AND deleted_at IS NULL
ORDER BY type, code;

-- 5. Verify: Show journal entries count by source type
SELECT 
    source_type,
    COUNT(*) as count,
    SUM(CASE WHEN status = 'POSTED' THEN 1 ELSE 0 END) as posted_count
FROM unified_journal_ledger
WHERE deleted_at IS NULL
GROUP BY source_type
ORDER BY source_type;

COMMIT;

-- After running this script:
-- 1. All corrupt closing data will be removed
-- 2. Account balances will be recalculated from POSTED journals
-- 3. You can safely attempt period closing again
