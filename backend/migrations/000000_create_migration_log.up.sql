-- ================================================================
-- Initial Migration: Create migration log table
-- ================================================================
-- This table is optional but useful for tracking custom migrations
-- golang-migrate already uses schema_migrations table
-- ================================================================

BEGIN;

-- Create migration_log table for additional tracking
CREATE TABLE IF NOT EXISTS migration_log (
    id SERIAL PRIMARY KEY,
    version INTEGER NOT NULL,
    description TEXT NOT NULL,
    applied_at TIMESTAMP NOT NULL DEFAULT NOW(),
    execution_time_ms INTEGER,
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    applied_by VARCHAR(255),
    UNIQUE(version)
);

-- Add index for querying
CREATE INDEX IF NOT EXISTS idx_migration_log_version ON migration_log(version);
CREATE INDEX IF NOT EXISTS idx_migration_log_applied_at ON migration_log(applied_at DESC);

-- Add comment
COMMENT ON TABLE migration_log IS 'Tracks custom migration history with additional metadata';

COMMIT;
