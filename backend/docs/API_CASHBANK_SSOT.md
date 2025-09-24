# Cash-Bank SSOT API Documentation

## Overview
API endpoints for Cash-Bank transactions with SSOT (Single Source of Truth) journal integration. All cash-bank transactions will automatically create unified journal entries.

**Base URL:** `/api/v1/cash-bank`

## Authentication
All endpoints require Bearer token authentication.

**Header:**
```
Authorization: Bearer <JWT_TOKEN>
```

## Account Management

### GET /accounts
Get all cash-bank accounts

**Permissions:** `cash_bank:view`

**Response:**
```json
{
  "status": "success",
  "message": "Cash-bank accounts retrieved successfully",
  "data": [
    {
      "id": 1,
      "code": "CSH-2024-0001",
      "name": "Kas Utama",
      "type": "CASH",
      "account_id": 5,
      "balance": 5000000.00,
      "currency": "IDR",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### GET /accounts/:id
Get cash-bank account by ID

**Permissions:** `cash_bank:view`

**Parameters:**
- `id` (path): Account ID

**Response:**
```json
{
  "status": "success", 
  "message": "Cash-bank account retrieved successfully",
  "data": {
    "id": 1,
    "code": "CSH-2024-0001",
    "name": "Kas Utama",
    "type": "CASH",
    "account_id": 5,
    "balance": 5000000.00,
    "currency": "IDR",
    "is_active": true
  }
}
```

### POST /accounts
Create new cash-bank account with opening balance

**Permissions:** `cash_bank:create`

**Request Body:**
```json
{
  "name": "Kas Utama",
  "type": "CASH",
  "currency": "IDR",
  "opening_balance": 5000000,
  "opening_date": "2024-01-01",
  "description": "Kas utama untuk operasional"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Cash-bank account created successfully",
  "data": {
    "id": 1,
    "code": "CSH-2024-0001", 
    "name": "Kas Utama",
    "type": "CASH",
    "balance": 5000000.00
  }
}
```

**SSOT Integration:**
- Automatically creates GL account if not specified
- Creates opening balance SSOT journal entry if opening_balance > 0
- Journal format: Dr. Cash Account, Cr. Owner Equity

### PUT /accounts/:id
Update cash-bank account

**Permissions:** `cash_bank:edit`

**Request Body:**
```json
{
  "name": "Kas Utama Updated",
  "description": "Updated description",
  "is_active": true
}
```

### DELETE /accounts/:id
Delete cash-bank account

**Permissions:** `cash_bank:delete`

**Response:**
```json
{
  "status": "success",
  "message": "Cash-bank account deleted successfully"
}
```

## Transaction Processing

### POST /transactions/deposit
Process deposit transaction with SSOT journal integration

**Permissions:** `cash_bank:create`

**Request Body:**
```json
{
  "account_id": 1,
  "date": "2024-01-15",
  "amount": 1000000,
  "reference": "DEP001",
  "notes": "Deposit tunai dari penjualan",
  "source_account_id": 12
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Deposit processed successfully",
  "data": {
    "transaction": {
      "id": 1,
      "amount": 1000000,
      "balance_after": 6000000,
      "transaction_date": "2024-01-15T00:00:00Z"
    },
    "message": "Deposit processed successfully with SSOT journal entry"
  }
}
```

**SSOT Integration:**
- Automatically creates journal entry: Dr. Cash Account, Cr. Revenue Account
- Updates GL account balances in real-time
- Reference format: `DEP-{ACCOUNT_CODE}-{TRANSACTION_ID}`

### POST /transactions/withdrawal
Process withdrawal transaction with SSOT journal integration

**Permissions:** `cash_bank:create`

**Request Body:**
```json
{
  "account_id": 1,
  "date": "2024-01-15",
  "amount": 500000,
  "reference": "WTH001", 
  "notes": "Penarikan untuk operasional",
  "target_account_id": 15
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Withdrawal processed successfully",
  "data": {
    "transaction": {
      "id": 2,
      "amount": -500000,
      "balance_after": 5500000,
      "transaction_date": "2024-01-15T00:00:00Z"
    },
    "message": "Withdrawal processed successfully with SSOT journal entry"
  }
}
```

**SSOT Integration:**
- Automatically creates journal entry: Dr. Expense Account, Cr. Cash Account
- Reference format: `WTH-{ACCOUNT_CODE}-{TRANSACTION_ID}`

### POST /transactions/transfer
Process transfer between cash-bank accounts with SSOT journal integration

**Permissions:** `cash_bank:create`

**Request Body:**
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

**Response:**
```json
{
  "status": "success",
  "message": "Transfer processed successfully", 
  "data": {
    "transfer": {
      "id": 1,
      "transfer_number": "TRF/2024/01/0001",
      "amount": 750000,
      "converted_amount": 750000
    },
    "message": "Transfer processed successfully with SSOT journal entry"
  }
}
```

**SSOT Integration:**
- Automatically creates journal entry: Dr. Destination Account, Cr. Source Account
- Reference format: `TRF-{FROM_CODE}-TO-{TO_CODE}-{TRANSACTION_ID}`

## Reporting

### GET /accounts/:id/transactions
Get transaction history for specific account

**Permissions:** `cash_bank:view`

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)
- `start_date` (optional): Start date filter (YYYY-MM-DD)
- `end_date` (optional): End date filter (YYYY-MM-DD)
- `type` (optional): Transaction type filter

**Response:**
```json
{
  "status": "success",
  "message": "Transactions retrieved successfully",
  "data": {
    "data": [
      {
        "id": 1,
        "amount": 1000000,
        "balance_after": 6000000,
        "transaction_date": "2024-01-15T00:00:00Z",
        "reference_type": "DEPOSIT",
        "notes": "Deposit tunai"
      }
    ],
    "total": 50,
    "page": 1,
    "limit": 20,
    "total_pages": 3
  }
}
```

### GET /reports/balance-summary
Get balance summary for all accounts

**Permissions:** `cash_bank:view`

**Response:**
```json
{
  "status": "success",
  "message": "Balance summary retrieved successfully",
  "data": {
    "total_cash": 5000000,
    "total_bank": 15000000,
    "total_balance": 20000000,
    "by_account": [
      {
        "account_id": 1,
        "account_name": "Kas Utama",
        "account_type": "CASH",
        "balance": 5000000,
        "currency": "IDR"
      }
    ],
    "by_currency": {
      "IDR": 20000000,
      "USD": 0
    }
  }
}
```

### GET /reports/payment-accounts
Get active payment accounts

**Permissions:** `cash_bank:view`

**Response:**
```json
{
  "status": "success",
  "message": "Payment accounts retrieved successfully", 
  "data": [
    {
      "id": 1,
      "name": "Kas Utama",
      "type": "CASH",
      "balance": 5000000,
      "currency": "IDR"
    }
  ]
}
```

### POST /accounts/:id/reconcile
Bank account reconciliation

**Permissions:** `cash_bank:edit`

**Request Body:**
```json
{
  "date": "2024-01-31",
  "statement_balance": 4950000,
  "items": [
    {
      "transaction_id": 1,
      "is_cleared": true,
      "notes": "Cleared in bank statement"
    }
  ]
}
```

## SSOT Integration

### GET /ssot/journals
Get SSOT journal entries for cash-bank transactions

**Permissions:** `reports:view`

**Response:**
```json
{
  "status": "success",
  "message": "SSOT journal entries endpoint",
  "data": {
    "message": "This endpoint will show SSOT journal entries for cash-bank transactions",
    "note": "Implementation would query unified_journal_ledger where source_type = 'CASH_BANK'"
  }
}
```

### POST /ssot/validate-integrity
Validate SSOT integration integrity

**Permissions:** `reports:view`

**Response:**
```json
{
  "status": "success",
  "message": "SSOT integration validation",
  "data": {
    "message": "This endpoint validates cash-bank SSOT integration integrity",
    "note": "Implementation would use CashBankSSOTJournalAdapter.ValidateJournalIntegrity()"
  }
}
```

## Error Handling

### Common Error Responses

**400 Bad Request:**
```json
{
  "status": "error",
  "message": "Invalid request data",
  "error": "Amount must be greater than 0"
}
```

**401 Unauthorized:**
```json
{
  "status": "error",
  "message": "User not authenticated"
}
```

**403 Forbidden:**
```json
{
  "status": "error", 
  "message": "Insufficient permissions"
}
```

**404 Not Found:**
```json
{
  "status": "error",
  "message": "Cash-bank account not found"
}
```

**500 Internal Server Error:**
```json
{
  "status": "error",
  "message": "Failed to process transaction",
  "error": "Database connection failed"
}
```

### Business Logic Errors

**Insufficient Balance:**
```json
{
  "status": "error",
  "message": "Insufficient balance. Available: 1000000.00"
}
```

**Journal Creation Failed:**
```json
{
  "status": "error",
  "message": "Failed to create SSOT journal entry",
  "error": "Account mapping not found"
}
```

## SSOT Journal Entry Format

### Deposit Transaction
```
Entry Number: JE-2024-01-000001
Reference: DEP-CSH-2024-0001-1
Date: 2024-01-15

Dr. 1101 - Kas Utama           Rp 1,000,000
    Cr. 4900 - Other Income                    Rp 1,000,000
```

### Withdrawal Transaction  
```
Entry Number: JE-2024-01-000002
Reference: WTH-CSH-2024-0001-2
Date: 2024-01-15

Dr. 5900 - General Expense     Rp 500,000
    Cr. 1101 - Kas Utama                      Rp 500,000
```

### Transfer Transaction
```
Entry Number: JE-2024-01-000003  
Reference: TRF-CSH-0001-TO-BNK-0002-3
Date: 2024-01-15

Dr. 1110 - Bank BCA            Rp 750,000
    Cr. 1101 - Kas Utama                      Rp 750,000
```

### Opening Balance
```
Entry Number: JE-2024-01-000004
Reference: OPN-CSH-2024-0001-4  
Date: 2024-01-01

Dr. 1101 - Kas Utama           Rp 5,000,000
    Cr. 3101 - Modal Pemilik                  Rp 5,000,000
```

## Testing

### Test with cURL

**Create Cash Account:**
```bash
curl -X POST http://localhost:8080/api/v1/cash-bank/accounts \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Kas Test",
    "type": "CASH", 
    "currency": "IDR",
    "opening_balance": 1000000,
    "opening_date": "2024-01-01"
  }'
```

**Process Deposit:**
```bash
curl -X POST http://localhost:8080/api/v1/cash-bank/transactions/deposit \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": 1,
    "date": "2024-01-15",
    "amount": 500000,
    "reference": "DEP001",
    "notes": "Test deposit"
  }'
```

### Test with Integration Script
```bash
cd backend
go run scripts/test_cashbank_ssot_integration.go
```

## Monitoring & Validation

### Balance Consistency Check
```sql
SELECT 
  cb.name,
  cb.balance as cashbank_balance,
  acc.balance as gl_balance,
  (cb.balance - acc.balance) as difference
FROM cash_banks cb
JOIN accounts acc ON cb.account_id = acc.id
WHERE cb.balance != acc.balance;
```

### Journal Integrity Check
```sql
SELECT cbt.id, cbt.amount, ujl.entry_number
FROM cash_bank_transactions cbt
LEFT JOIN unified_journal_ledger ujl ON 
  ujl.source_type = 'CASH_BANK' AND ujl.source_id = cbt.id
WHERE ujl.id IS NULL;
```