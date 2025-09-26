-- Migration: 030_create_account_balances_materialized_view.sql
-- Purpose: Create missing account_balances materialized view for SSOT system
-- Date: 2025-09-26

BEGIN;

-- Drop materialized view if it exists to recreate safely
DROP MATERIALIZED VIEW IF EXISTS account_balances;

-- Create account_balances materialized view
CREATE MATERIALIZED VIEW account_balances AS
SELECT 
    a.id as account_id,
    a.code as account_code,
    a.name as account_name,
    a.type as account_type,
    a.category as account_category,
    
    -- Current balance from accounts table
    a.balance as current_balance,
    
    -- Calculate balance from SSOT journal system (if tables exist)
    CASE 
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_lines') THEN
            COALESCE((
                SELECT 
                    CASE 
                        WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                            SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
                        ELSE 
                            SUM(ujl.credit_amount) - SUM(ujl.debit_amount)
                    END
                FROM unified_journal_lines ujl
                JOIN unified_journal_ledger ujd ON ujl.journal_id = ujd.id
                WHERE ujl.account_id = a.id 
                  AND ujd.status = 'POSTED'
                  AND ujd.deleted_at IS NULL
            ), 0)
        ELSE 
            -- Fallback to traditional journal system
            COALESCE((
                SELECT 
                    CASE 
                        WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                            SUM(jl.debit_amount) - SUM(jl.credit_amount)
                        ELSE 
                            SUM(jl.credit_amount) - SUM(jl.debit_amount)
                    END
                FROM journal_lines jl
                JOIN journal_entries je ON jl.journal_entry_id = je.id
                WHERE jl.account_id = a.id 
                  AND je.status = 'POSTED'
                  AND je.deleted_at IS NULL
            ), 0)
    END as calculated_balance,
    
    -- Balance difference for reconciliation
    a.balance - CASE 
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_lines') THEN
            COALESCE((
                SELECT 
                    CASE 
                        WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                            SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
                        ELSE 
                            SUM(ujl.credit_amount) - SUM(ujl.debit_amount)
                    END
                FROM unified_journal_lines ujl
                JOIN unified_journal_ledger ujd ON ujl.journal_id = ujd.id
                WHERE ujl.account_id = a.id 
                  AND ujd.status = 'POSTED'
                  AND ujd.deleted_at IS NULL
            ), 0)
        ELSE 
            COALESCE((
                SELECT 
                    CASE 
                        WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                            SUM(jl.debit_amount) - SUM(jl.credit_amount)
                        ELSE 
                            SUM(jl.credit_amount) - SUM(jl.debit_amount)
                    END
                FROM journal_lines jl
                JOIN journal_entries je ON jl.journal_entry_id = je.id
                WHERE jl.account_id = a.id 
                  AND je.status = 'POSTED'
                  AND je.deleted_at IS NULL
            ), 0)
    END as balance_difference,
    
    -- Metadata
    a.is_active,
    a.created_at,
    a.updated_at,
    NOW() as last_refresh

FROM accounts a
WHERE a.deleted_at IS NULL;

-- Create indexes on materialized view for better performance
CREATE INDEX IF NOT EXISTS idx_account_balances_account_id ON account_balances(account_id);
CREATE INDEX IF NOT EXISTS idx_account_balances_account_type ON account_balances(account_type);
-- Parent ID index commented out as accounts table may not have parent_id column
-- CREATE INDEX IF NOT EXISTS idx_account_balances_parent_id ON account_balances(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_account_balances_difference ON account_balances(balance_difference) WHERE ABS(balance_difference) > 0.01;

-- Create function to refresh account balances materialized view
CREATE OR REPLACE FUNCTION refresh_account_balances()
RETURNS VOID AS $$
BEGIN
    REFRESH MATERIALIZED VIEW account_balances;
    
    -- Log the refresh
    RAISE NOTICE 'Account balances materialized view refreshed at %', NOW();
END;
$$ LANGUAGE plpgsql;

-- Create function to get balance summary
CREATE OR REPLACE FUNCTION get_balance_summary()
RETURNS TABLE(
    total_assets DECIMAL(20,2),
    total_liabilities DECIMAL(20,2),
    total_equity DECIMAL(20,2),
    total_revenue DECIMAL(20,2),
    total_expenses DECIMAL(20,2),
    balancing_check DECIMAL(20,2)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COALESCE(SUM(CASE WHEN account_type = 'ASSET' THEN calculated_balance ELSE 0 END), 0) as total_assets,
        COALESCE(SUM(CASE WHEN account_type = 'LIABILITY' THEN calculated_balance ELSE 0 END), 0) as total_liabilities,
        COALESCE(SUM(CASE WHEN account_type = 'EQUITY' THEN calculated_balance ELSE 0 END), 0) as total_equity,
        COALESCE(SUM(CASE WHEN account_type = 'REVENUE' THEN calculated_balance ELSE 0 END), 0) as total_revenue,
        COALESCE(SUM(CASE WHEN account_type = 'EXPENSE' THEN calculated_balance ELSE 0 END), 0) as total_expenses,
        -- Assets should equal Liabilities + Equity (Revenue - Expenses contribute to Equity)
        COALESCE(SUM(CASE WHEN account_type = 'ASSET' THEN calculated_balance ELSE 0 END), 0) - 
        (COALESCE(SUM(CASE WHEN account_type = 'LIABILITY' THEN calculated_balance ELSE 0 END), 0) + 
         COALESCE(SUM(CASE WHEN account_type = 'EQUITY' THEN calculated_balance ELSE 0 END), 0) +
         COALESCE(SUM(CASE WHEN account_type = 'REVENUE' THEN calculated_balance ELSE 0 END), 0) -
         COALESCE(SUM(CASE WHEN account_type = 'EXPENSE' THEN calculated_balance ELSE 0 END), 0)) as balancing_check
    FROM account_balances;
END;
$$ LANGUAGE plpgsql;

-- Initial refresh of the materialized view
SELECT refresh_account_balances();

-- Log completion
RAISE NOTICE '✅ Account balances materialized view created and populated';
RAISE NOTICE '📊 SSOT Journal System should now work properly';

COMMIT;