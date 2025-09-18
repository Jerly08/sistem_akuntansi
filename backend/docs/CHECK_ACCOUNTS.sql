-- Query untuk check account IDs yang tersedia
-- Jalankan di database untuk lihat account IDs yang valid

-- 1. Check Fixed Asset Accounts
SELECT id, code, name, type 
FROM accounts 
WHERE type = 'ASSET' 
  AND (code LIKE '15%' OR name ILIKE '%asset%' OR name ILIKE '%fixed%')
  AND is_active = true
ORDER BY code;

-- 2. Check Liability Accounts  
SELECT id, code, name, type
FROM accounts 
WHERE type = 'LIABILITY'
  AND is_active = true
ORDER BY code;

-- 3. Check Expense Accounts (for depreciation)
SELECT id, code, name, type
FROM accounts 
WHERE type = 'EXPENSE' 
  AND (code LIKE '62%' OR name ILIKE '%depreciation%')
  AND is_active = true
ORDER BY code;

-- 4. Check ALL accounts untuk debugging
SELECT id, code, name, type, is_active
FROM accounts 
ORDER BY type, code;
