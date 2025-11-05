-- Migration: Increase payment code size from VARCHAR(20) to VARCHAR(30)
-- Reason: SETOR-PPN code format can exceed 20 characters
-- Format: SETOR-PPN-YYMM-NNNN = up to 19 chars, need buffer for future

ALTER TABLE payments ALTER COLUMN code TYPE VARCHAR(30);

-- Update model constraint if any
COMMENT ON COLUMN payments.code IS 'Payment code (max 30 chars)';
