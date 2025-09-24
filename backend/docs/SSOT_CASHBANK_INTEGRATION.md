# SSOT Cash-Bank Integration Documentation

## Overview

Integrasi SSOT (Single Source of Truth) untuk modul Cash-Bank memungkinkan pencatatan transaksi kas dan bank secara terpadu dengan sistem jurnal yang konsisten. Setiap transaksi cash-bank akan otomatis membuat jurnal entry yang mengikuti prinsip double-entry bookkeeping.

## Architecture

```
Cash-Bank Transaction
         ↓
CashBankService
         ↓
CashBankSSOTJournalAdapter
         ↓
UnifiedJournalService
         ↓
SSOTJournalEntry & SSOTJournalLine
```

## Components

### 1. CashBankSSOTJournalAdapter
Adapter yang menghubungkan transaksi cash-bank dengan sistem SSOT journal.

**File:** `backend/services/cashbank_ssot_journal_adapter.go`

**Fungsi utama:**
- `CreateDepositJournalEntry()` - Membuat jurnal untuk deposit
- `CreateWithdrawalJournalEntry()` - Membuat jurnal untuk withdrawal 
- `CreateTransferJournalEntry()` - Membuat jurnal untuk transfer
- `CreateOpeningBalanceJournalEntry()` - Membuat jurnal untuk opening balance
- `ReverseJournalEntry()` - Membuat jurnal reversal
- `ValidateJournalIntegrity()` - Validasi integritas jurnal

### 2. Modified CashBankService
Service yang telah dimodifikasi untuk menggunakan SSOT adapter.

**File:** `backend/services/cashbank_service.go`

**Perubahan utama:**
- Inisialisasi `CashBankSSOTJournalAdapter`
- Penggantian fungsi jurnal manual dengan panggilan adapter
- Integrasi dengan `UnifiedJournalService`

### 3. CashBankHandler 
Handler API untuk endpoint cash-bank dengan integrasi SSOT.

**File:** `backend/handlers/cash_bank_handler.go`

## Transaction Types & Journal Entries

### 1. Deposit (Deposit)
**Jurnal Entry:**
```
Dr. Cash/Bank Account     Rp X
  Cr. Other Income              Rp X
```

**Request Example:**
```json
{
  "account_id": 1,
  "date": "2024-01-15",
  "amount": 1000000,
  "reference": "DEP001",
  "notes": "Deposit tunai dari penjualan",
  "source_account_id": 12  // Optional revenue account
}
```

### 2. Withdrawal (Penarikan)
**Jurnal Entry:**
```
Dr. General Expense       Rp X
  Cr. Cash/Bank Account         Rp X
```

**Request Example:**
```json
{
  "account_id": 1,
  "date": "2024-01-15", 
  "amount": 500000,
  "reference": "WTH001",
  "notes": "Penarikan untuk operasional",
  "target_account_id": 15  // Optional expense account
}
```

### 3. Transfer (Transfer)
**Jurnal Entry:**
```
Dr. Destination Account   Rp X
  Cr. Source Account            Rp X
```

**Request Example:**
```json
{
  "from_account_id": 1,
  "to_account_id": 2,
  "date": "2024-01-15",
  "amount": 750000,
  "exchange_rate": 1.0,
  "reference": "TRF001", 
  "notes": "Transfer antar rekening"
}
```

### 4. Opening Balance (Saldo Awal)
**Jurnal Entry:**
```
Dr. Cash/Bank Account     Rp X
  Cr. Owner Equity              Rp X
```

## API Endpoints

### Cash-Bank Accounts
- `GET /api/cash-bank/accounts` - Get all accounts
- `GET /api/cash-bank/accounts/:id` - Get account by ID
- `POST /api/cash-bank/accounts` - Create new account
- `PUT /api/cash-bank/accounts/:id` - Update account
- `DELETE /api/cash-bank/accounts/:id` - Delete account

### Transactions
- `POST /api/cash-bank/deposit` - Process deposit
- `POST /api/cash-bank/withdrawal` - Process withdrawal
- `POST /api/cash-bank/transfer` - Process transfer
- `GET /api/cash-bank/accounts/:id/transactions` - Get transactions

### Reporting & Management
- `GET /api/cash-bank/balance-summary` - Get balance summary
- `GET /api/cash-bank/payment-accounts` - Get active payment accounts
- `POST /api/cash-bank/accounts/:id/reconcile` - Bank reconciliation

### SSOT Integration
- `GET /api/cash-bank/ssot-journals` - View SSOT journal entries
- `POST /api/cash-bank/validate-integrity` - Validate SSOT integrity

## Database Schema

### SSOT Tables Used
- `unified_journal_ledger` - Journal entries
- `unified_journal_lines` - Journal lines  
- `journal_event_log` - Audit trail
- `account_balances` - Real-time balances

### Source Type
All cash-bank transactions use source type: `CASH_BANK`

### Reference Format
- Deposit: `DEP-{ACCOUNT_CODE}-{TRANSACTION_ID}`
- Withdrawal: `WTH-{ACCOUNT_CODE}-{TRANSACTION_ID}`
- Transfer: `TRF-{FROM_CODE}-TO-{TO_CODE}-{TRANSACTION_ID}`
- Opening: `OPN-{ACCOUNT_CODE}-{TRANSACTION_ID}`

## Testing

### Manual Testing
Run the test script:
```bash
go run backend/scripts/test_cashbank_ssot_integration.go
```

### Test Coverage
✅ Cash account creation with opening balance
✅ Bank account creation with opening balance  
✅ Deposit transaction processing
✅ Withdrawal transaction processing
✅ Transfer transaction processing
✅ SSOT journal entries verification
✅ Balance consistency verification
✅ SSOT adapter validation
✅ Journal reversal functionality

## Default Accounts

### Revenue Account for Deposits
- **Code:** 4900
- **Name:** Other Income
- **Type:** REVENUE

### Expense Account for Withdrawals  
- **Code:** 5900
- **Name:** General Expense
- **Type:** EXPENSE

### Equity Account for Opening Balance
- **Code:** 3101
- **Name:** Modal Pemilik
- **Type:** EQUITY

## Error Handling

### Common Errors
1. **Insufficient Balance:** When withdrawal/transfer amount > available balance
2. **Account Not Found:** Invalid account IDs
3. **Journal Creation Failed:** Issues with SSOT journal entry creation
4. **Validation Failed:** Journal entry not balanced or invalid data

### Rollback Mechanism
Semua transaksi menggunakan database transactions. Jika terjadi error pada tahap manapun (update balance, create transaction, create journal), maka seluruh proses akan di-rollback.

## Monitoring & Maintenance

### Balance Consistency Check
```sql
-- Check if cash-bank balances match GL account balances
SELECT 
  cb.name as account_name,
  cb.balance as cashbank_balance,
  acc.balance as gl_balance,
  (cb.balance - acc.balance) as difference
FROM cash_banks cb
JOIN accounts acc ON cb.account_id = acc.id
WHERE cb.balance != acc.balance;
```

### Journal Integrity Check
```sql
-- Check for cash-bank transactions without SSOT journal entries
SELECT cbt.*
FROM cash_bank_transactions cbt
LEFT JOIN unified_journal_ledger ujl ON 
  ujl.source_type = 'CASH_BANK' AND ujl.source_id = cbt.id
WHERE ujl.id IS NULL;
```

### Balance Validation
```sql
-- Verify journal entries are balanced
SELECT 
  entry_number,
  SUM(debit_amount) as total_debits,
  SUM(credit_amount) as total_credits,
  (SUM(debit_amount) - SUM(credit_amount)) as difference
FROM unified_journal_ledger ujl
JOIN unified_journal_lines ujline ON ujl.id = ujline.journal_id
WHERE ujl.source_type = 'CASH_BANK'
GROUP BY ujl.id, ujl.entry_number
HAVING SUM(debit_amount) != SUM(credit_amount);
```

## Implementation Notes

1. **Auto-Post:** Semua jurnal cash-bank di-post otomatis (`auto_post = true`)
2. **GL Sync:** Balance GL account akan tersinkronisasi melalui SSOT system
3. **Audit Trail:** Semua perubahan tercatat di `journal_event_log`
4. **Reversal Support:** Mendukung reversal journal entries untuk koreksi
5. **Multi-Currency:** Dukungan exchange rate untuk transfer multi-currency (future enhancement)

## Future Enhancements

1. **Bank Statement Import:** Import dan rekonsiliasi otomatis dari bank statement
2. **Multi-Currency Support:** Dukungan penuh untuk transaksi multi-currency
3. **Automated Reconciliation:** Otomasi proses rekonsiliasi bank
4. **Cash Flow Reporting:** Laporan cash flow berdasarkan SSOT data
5. **Integration with External Banks:** Integrasi dengan API bank untuk real-time balance
6. **Advanced Validation Rules:** Business rules validation yang lebih kompleks
7. **Workflow Approval:** Approval workflow untuk transaksi tertentu

## Troubleshooting

### 1. Journal Creation Failed
**Symptoms:** Transaksi berhasil tetapi journal tidak terbuat
**Solution:** 
- Check SSOT service status
- Verify account mapping
- Check audit logs

### 2. Balance Mismatch
**Symptoms:** Cash-bank balance != GL account balance
**Solution:**
- Run integrity validation
- Check for missing journal entries
- Manual balance adjustment if necessary

### 3. Validation Errors
**Symptoms:** Transaction validation failed
**Solution:**
- Verify account exists and active
- Check amount format and constraints
- Validate date format

## Contact & Support

Untuk pertanyaan terkait integrasi SSOT Cash-Bank:
- Review audit logs di `journal_event_log`
- Run validation script
- Check balance consistency
- Contact development team untuk troubleshooting advanced issues