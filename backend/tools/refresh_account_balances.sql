-- Refresh Materialized View account_balances
-- This will update all account balances based on current journal data

BEGIN;

-- Refresh the materialized view to get latest account balances
REFRESH MATERIALIZED VIEW account_balances;

-- Check the refresh result
SELECT 
    'REFRESH COMPLETED' as status,
    COUNT(*) as total_accounts,
    COUNT(CASE WHEN current_balance != 0 THEN 1 END) as accounts_with_balance,
    SUM(CASE WHEN current_balance > 0 THEN current_balance ELSE 0 END) as total_debit_balances,
    SUM(CASE WHEN current_balance < 0 THEN ABS(current_balance) ELSE 0 END) as total_credit_balances,
    NOW() as refreshed_at
FROM account_balances;

-- Show sample of updated balances
SELECT 
    account_code,
    account_name,
    account_type,
    current_balance,
    transaction_count,
    last_updated
FROM account_balances 
WHERE current_balance != 0
ORDER BY ABS(current_balance) DESC
LIMIT 20;

COMMIT;