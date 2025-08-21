-- Fix cash_banks account relationships
-- Match cash banks to appropriate accounts

-- Update CASH001 (Kas Besar) to use account 1101 (Kas)
UPDATE cash_banks 
SET account_id = 3
WHERE code = 'CASH001' AND account_id IS NULL;

-- Update BANK001 (Bank BCA - Operasional) to use account 1102 (Bank BCA)
UPDATE cash_banks 
SET account_id = 4
WHERE code = 'BANK001' AND account_id IS NULL;

-- Update BANK002 (Bank Mandiri - Payroll) to use account 1103 (Bank Mandiri)
UPDATE cash_banks 
SET account_id = 5
WHERE code = 'BANK002' AND account_id IS NULL;

-- Verify the updates
SELECT 
    cb.id,
    cb.code,
    cb.name,
    cb.type,
    cb.account_id,
    a.code as account_code,
    a.name as account_name,
    cb.is_active
FROM cash_banks cb 
LEFT JOIN accounts a ON cb.account_id = a.id 
ORDER BY cb.id;
