-- Migration: Add snapshot columns to accounting_periods table for archive tracking
-- Purpose: Store Balance Sheet and P&L snapshots for historical reporting
-- Date: 2025-01-10

DO $$
BEGIN
    -- Add balance_sheet_snapshot column (JSONB for better querying)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                  WHERE table_name = 'accounting_periods' AND column_name = 'balance_sheet_snapshot') THEN
        RAISE NOTICE '🔧 Adding balance_sheet_snapshot column';
        ALTER TABLE accounting_periods ADD COLUMN balance_sheet_snapshot JSONB;
        COMMENT ON COLUMN accounting_periods.balance_sheet_snapshot IS 'Snapshot of Balance Sheet at period end date';
    END IF;
    
    -- Add profit_loss_snapshot column (JSONB for better querying)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                  WHERE table_name = 'accounting_periods' AND column_name = 'profit_loss_snapshot') THEN
        RAISE NOTICE '🔧 Adding profit_loss_snapshot column';
        ALTER TABLE accounting_periods ADD COLUMN profit_loss_snapshot JSONB;
        COMMENT ON COLUMN accounting_periods.profit_loss_snapshot IS 'Snapshot of Profit & Loss for the period';
    END IF;
    
    -- Add financial_metrics column for quick summary access
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                  WHERE table_name = 'accounting_periods' AND column_name = 'financial_metrics') THEN
        RAISE NOTICE '🔧 Adding financial_metrics column';
        ALTER TABLE accounting_periods ADD COLUMN financial_metrics JSONB;
        COMMENT ON COLUMN accounting_periods.financial_metrics IS 'Key financial metrics (gross profit, operating income, ratios, etc.)';
    END IF;
    
    -- Add snapshot_generated_at timestamp
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                  WHERE table_name = 'accounting_periods' AND column_name = 'snapshot_generated_at') THEN
        RAISE NOTICE '🔧 Adding snapshot_generated_at column';
        ALTER TABLE accounting_periods ADD COLUMN snapshot_generated_at TIMESTAMP;
        COMMENT ON COLUMN accounting_periods.snapshot_generated_at IS 'When the snapshot was generated';
    END IF;
    
    -- Add period_type column (MONTHLY, QUARTERLY, ANNUAL, CUSTOM)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                  WHERE table_name = 'accounting_periods' AND column_name = 'period_type') THEN
        RAISE NOTICE '🔧 Adding period_type column';
        ALTER TABLE accounting_periods ADD COLUMN period_type VARCHAR(20) DEFAULT 'CUSTOM';
        COMMENT ON COLUMN accounting_periods.period_type IS 'Type of period: MONTHLY, QUARTERLY, SEMESTER, ANNUAL, CUSTOM';
    END IF;
    
    -- Add fiscal_year column for grouping
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                  WHERE table_name = 'accounting_periods' AND column_name = 'fiscal_year') THEN
        RAISE NOTICE '🔧 Adding fiscal_year column';
        ALTER TABLE accounting_periods ADD COLUMN fiscal_year INTEGER;
        CREATE INDEX IF NOT EXISTS idx_accounting_periods_fiscal_year ON accounting_periods(fiscal_year);
        COMMENT ON COLUMN accounting_periods.fiscal_year IS 'Fiscal year this period belongs to';
    END IF;
    
    -- Add account_count for summary
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                  WHERE table_name = 'accounting_periods' AND column_name = 'account_count') THEN
        RAISE NOTICE '🔧 Adding account_count column';
        ALTER TABLE accounting_periods ADD COLUMN account_count INTEGER DEFAULT 0;
        COMMENT ON COLUMN accounting_periods.account_count IS 'Number of active accounts during the period';
    END IF;
    
    -- Add transaction_count for summary
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                  WHERE table_name = 'accounting_periods' AND column_name = 'transaction_count') THEN
        RAISE NOTICE '🔧 Adding transaction_count column';
        ALTER TABLE accounting_periods ADD COLUMN transaction_count INTEGER DEFAULT 0;
        COMMENT ON COLUMN accounting_periods.transaction_count IS 'Number of transactions during the period';
    END IF;
    
    -- Create indexes for JSONB columns (GIN indexes for fast querying)
    CREATE INDEX IF NOT EXISTS idx_accounting_periods_bs_snapshot ON accounting_periods USING GIN (balance_sheet_snapshot);
    CREATE INDEX IF NOT EXISTS idx_accounting_periods_pl_snapshot ON accounting_periods USING GIN (profit_loss_snapshot);
    CREATE INDEX IF NOT EXISTS idx_accounting_periods_metrics ON accounting_periods USING GIN (financial_metrics);
    
    -- Create composite index for common queries
    CREATE INDEX IF NOT EXISTS idx_accounting_periods_type_year ON accounting_periods(period_type, fiscal_year);
    
    RAISE NOTICE '✅ Migration 050 completed successfully';
END $$;
