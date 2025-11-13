-- Check closing journal entries
SELECT 
    ujl.id,
    ujl.code,
    ujl.entry_date,
    ujl.description,
    ujl.status,
    ujl.created_at
FROM unified_journal_ledger ujl
WHERE ujl.code LIKE 'CLO-%'
ORDER BY ujl.entry_date;

-- Check journal lines for account 3201 (Laba Ditahan)
SELECT 
    ujl.id as journal_id,
    ujl.code,
    ujl.entry_date,
    a.code as account_code,
    a.name as account_name,
    jl.debit_amount,
    jl.credit_amount
FROM unified_journal_ledger ujl
JOIN unified_journal_lines jl ON jl.journal_id = ujl.id
JOIN accounts a ON a.id = jl.account_id
WHERE a.code = '3201'
AND ujl.status = 'POSTED'
ORDER BY ujl.entry_date;
