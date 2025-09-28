-- CashBank ⇄ COA Sync Trigger (Assets only)
-- This trigger keeps cash_banks.balance and linked accounts.balance equal to the
-- sum of cash_bank_transactions for the given cash_bank_id.
-- Safe to run multiple times: DROP TRIGGER guards are included.

-- Function
CREATE OR REPLACE FUNCTION sync_cashbank_balance_to_coa()
RETURNS TRIGGER AS $$
DECLARE
    coa_account_id INTEGER;
    transaction_sum DECIMAL(18,2);
    acct_type TEXT;
BEGIN
    -- Identify the CashBank id affected
    -- and its linked COA account
    IF TG_OP = 'DELETE' THEN
        SELECT account_id INTO coa_account_id FROM cash_banks WHERE id = OLD.cash_bank_id;
    ELSE
        SELECT account_id INTO coa_account_id FROM cash_banks WHERE id = NEW.cash_bank_id;
    END IF;

    -- Skip if no linked COA account
    IF coa_account_id IS NULL THEN
        RETURN COALESCE(NEW, OLD);
    END IF;

    -- Ensure COA account is ASSET
    SELECT type INTO acct_type FROM accounts WHERE id = coa_account_id;
    IF acct_type IS NULL OR acct_type <> 'ASSET' THEN
        RETURN COALESCE(NEW, OLD);
    END IF;

    -- Calculate transaction sum as source of truth
    SELECT COALESCE(SUM(amount), 0) INTO transaction_sum
      FROM cash_bank_transactions
     WHERE cash_bank_id = COALESCE(NEW.cash_bank_id, OLD.cash_bank_id)
       AND deleted_at IS NULL;

    -- Update CashBank balance
    UPDATE cash_banks
       SET balance = transaction_sum,
           updated_at = NOW()
     WHERE id = COALESCE(NEW.cash_bank_id, OLD.cash_bank_id);

    -- Update linked COA balance
    UPDATE accounts
       SET balance = transaction_sum,
           updated_at = NOW()
     WHERE id = coa_account_id;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Trigger on cash_bank_transactions
DROP TRIGGER IF EXISTS trg_sync_cashbank_coa ON cash_bank_transactions;
CREATE TRIGGER trg_sync_cashbank_coa
    AFTER INSERT OR UPDATE OR DELETE ON cash_bank_transactions
    FOR EACH ROW
    EXECUTE FUNCTION sync_cashbank_balance_to_coa();