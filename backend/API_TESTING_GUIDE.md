# API Testing Guide - Sistem Akuntansi

Panduan lengkap untuk testing API sistem akuntansi menggunakan berbagai tools seperti Postman, cURL, atau REST client lainnya.

## 📋 Daftar Isi

- [Prerequisite](#prerequisite)
- [Authentication](#authentication)
- [Journal Entries](#journal-entries)
- [Admin](#admin)
- [CashBank](#cashbank)
- [Balance Monitoring](#balance-monitoring)
- [Payments](#payments)
- [Purchases](#purchases)
- [Enhanced Reports](#enhanced-reports)
- [Security](#security)
- [Dashboard](#dashboard)
- [Journal Drilldown](#journal-drilldown)
- [Testing Tools](#testing-tools)
- [Environment Variables](#environment-variables)

## 🔧 Prerequisite

Sebelum melakukan testing, pastikan:
1. Server backend sudah berjalan
2. Database sudah ter-setup dengan benar
3. Memiliki akses ke tools testing (Postman, Thunder Client, atau cURL)
4. Base URL sudah dikonfigurasi (contoh: `http://localhost:3000` atau `https://your-domain.com`)

## 🔐 Authentication

### 1. User Registration
```http
POST /auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123",
  "full_name": "Test User"
}
```

### 2. User Login
```http
POST /auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}
```

**Response akan berisi access_token yang diperlukan untuk endpoint lainnya**

### 3. Refresh Token
```http
POST /auth/refresh
Content-Type: application/json
Authorization: Bearer {refresh_token}
```

### 4. Validate Token
```http
GET /auth/validate-token
Authorization: Bearer {access_token}
```

### 5. Get User Profile
```http
GET /profile
Authorization: Bearer {access_token}
```

## 📊 Journal Entries

### 1. List Journal Entries
```http
GET /journal-entries
Authorization: Bearer {access_token}
```

### 2. Get Account Journal Entries
```http
GET /accounts/{account_id}/journal-entries
Authorization: Bearer {access_token}
```

### 3. Create Journal Entry
```http
POST /journal-entries
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "date": "2024-01-15",
  "description": "Test journal entry",
  "reference": "JE001",
  "entries": [
    {
      "account_id": 1,
      "debit": 1000,
      "credit": 0,
      "description": "Debit entry"
    },
    {
      "account_id": 2,
      "debit": 0,
      "credit": 1000,
      "description": "Credit entry"
    }
  ]
}
```

### 4. Auto-generate from Purchase
```http
POST /journal-entries/auto-generate/purchase
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "purchase_id": 1,
  "auto_post": false
}
```

### 5. Auto-generate from Sale
```http
POST /journal-entries/auto-generate/sale
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "sale_id": 1,
  "auto_post": false
}
```

### 6. Get Journal Entry Summary
```http
GET /journal-entries/summary
Authorization: Bearer {access_token}
```

### 7. Get Specific Journal Entry
```http
GET /journal-entries/{id}
Authorization: Bearer {access_token}
```

### 8. Update Journal Entry
```http
PUT /journal-entries/{id}
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "description": "Updated description",
  "entries": [
    {
      "id": 1,
      "account_id": 1,
      "debit": 1500,
      "credit": 0
    }
  ]
}
```

### 9. Delete Journal Entry
```http
DELETE /journal-entries/{id}
Authorization: Bearer {access_token}
```

### 10. Post Journal Entry
```http
POST /journal-entries/{id}/post
Authorization: Bearer {access_token}
```

### 11. Reverse Journal Entry
```http
POST /journal-entries/{id}/reverse
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "reason": "Error in original entry"
}
```

## ⚙️ Admin

### 1. Check Cash Bank GL Links
```http
GET /api/admin/check-cashbank-gl-links
Authorization: Bearer {access_token}
```

### 2. Fix Cash Bank GL Links
```http
POST /api/admin/fix-cashbank-gl-links
Authorization: Bearer {access_token}
```

## 💰 CashBank

### 1. Get Cash and Bank Accounts
```http
GET /api/cashbank/accounts
Authorization: Bearer {access_token}
```

### 2. Create Cash/Bank Account
```http
POST /api/cashbank/accounts
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "name": "Bank BCA",
  "type": "bank",
  "account_number": "1234567890",
  "initial_balance": 50000000,
  "gl_account_id": 1
}
```

### 3. Get Account by ID
```http
GET /api/cashbank/accounts/{id}
Authorization: Bearer {access_token}
```

### 4. Update Cash/Bank Account
```http
PUT /api/cashbank/accounts/{id}
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "name": "Bank BCA - Updated",
  "account_number": "1234567890"
}
```

### 5. Delete Cash/Bank Account
```http
DELETE /api/cashbank/accounts/{id}
Authorization: Bearer {access_token}
```

### 6. Reconcile Bank Account
```http
POST /api/cashbank/accounts/{id}/reconcile
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "statement_date": "2024-01-31",
  "statement_balance": 48500000,
  "reconciliation_items": []
}
```

### 7. Get Account Transactions
```http
GET /api/cashbank/accounts/{id}/transactions
Authorization: Bearer {access_token}
```

### 8. Get Balance Summary
```http
GET /api/cashbank/balance-summary
Authorization: Bearer {access_token}
```

### 9. Process Deposit
```http
POST /api/cashbank/deposit
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "account_id": 1,
  "amount": 1000000,
  "source_account_id": 2,
  "description": "Customer payment",
  "date": "2024-01-15"
}
```

### 10. Get Deposit Source Accounts
```http
GET /api/cashbank/deposit-source-accounts
Authorization: Bearer {access_token}
```

### 11. Get Payment Accounts
```http
GET /api/cashbank/payment-accounts
Authorization: Bearer {access_token}
```

### 12. Get Revenue Accounts
```http
GET /api/cashbank/revenue-accounts
Authorization: Bearer {access_token}
```

### 13. Process Transfer
```http
POST /api/cashbank/transfer
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "from_account_id": 1,
  "to_account_id": 2,
  "amount": 500000,
  "description": "Transfer funds",
  "date": "2024-01-15"
}
```

### 14. Process Withdrawal
```http
POST /api/cashbank/withdrawal
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "account_id": 1,
  "amount": 300000,
  "expense_account_id": 3,
  "description": "Office supplies",
  "date": "2024-01-15"
}
```

## 📊 Balance Monitoring

### 1. Get Balance Health Metrics
```http
GET /api/monitoring/balance-health
Authorization: Bearer {access_token}
```

### 2. Check Balance Synchronization
```http
GET /api/monitoring/balance-sync
Authorization: Bearer {access_token}
```

### 3. Get Current Balance Discrepancies
```http
GET /api/monitoring/discrepancies
Authorization: Bearer {access_token}
```

### 4. Fix Balance Discrepancies
```http
POST /api/monitoring/fix-discrepancies
Authorization: Bearer {access_token}
```

### 5. Get Synchronization Status Summary
```http
GET /api/monitoring/sync-status
Authorization: Bearer {access_token}
```

## 💳 Payments

### 1. Get Payments List
```http
GET /api/payments
Authorization: Bearer {access_token}
```

### 2. Get Payment Analytics
```http
GET /api/payments/analytics
Authorization: Bearer {access_token}
```

### 3. Get Recent Payments (Debug)
```http
GET /api/payments/debug/recent
Authorization: Bearer {access_token}
```

### 4. Export Payment Report (Excel)
```http
GET /api/payments/export/excel
Authorization: Bearer {access_token}
```

### 5. Create Payable Payment
```http
POST /api/payments/payable
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "vendor_id": 1,
  "amount": 2000000,
  "payment_date": "2024-01-15",
  "payment_method": "bank_transfer",
  "reference": "PAY001",
  "bills": [
    {
      "bill_id": 1,
      "amount": 2000000
    }
  ]
}
```

### 6. Create Receivable Payment
```http
POST /api/payments/receivable
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "customer_id": 1,
  "amount": 3000000,
  "payment_date": "2024-01-15",
  "payment_method": "cash",
  "reference": "REC001",
  "invoices": [
    {
      "invoice_id": 1,
      "amount": 3000000
    }
  ]
}
```

### 7. Export Payment Report (PDF)
```http
GET /api/payments/report/pdf
Authorization: Bearer {access_token}
```

### 8. Get Payment Summary
```http
GET /api/payments/summary
Authorization: Bearer {access_token}
```

### 9. Get Unpaid Bills
```http
GET /api/payments/unpaid-bills/{vendor_id}
Authorization: Bearer {access_token}
```

### 10. Get Unpaid Invoices
```http
GET /api/payments/unpaid-invoices/{customer_id}
Authorization: Bearer {access_token}
```

### 11. Get Payment by ID
```http
GET /api/payments/{id}
Authorization: Bearer {access_token}
```

### 12. Delete Payment
```http
DELETE /api/payments/{id}
Authorization: Bearer {access_token}
```

### 13. Cancel Payment
```http
POST /api/payments/{id}/cancel
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "reason": "Payment error"
}
```

### 14. Export Payment Detail (PDF)
```http
GET /api/payments/{id}/pdf
Authorization: Bearer {access_token}
```

## 🛒 Purchases

### 1. Get Purchase for Payment
```http
GET /api/purchases/{id}/for-payment
Authorization: Bearer {access_token}
```

### 2. Create Integrated Payment for Purchase
```http
POST /api/purchases/{id}/integrated-payment
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "payment_amount": 1500000,
  "payment_date": "2024-01-15",
  "payment_method": "bank_transfer"
}
```

### 3. Get Purchase Payments
```http
GET /api/purchases/{id}/payments
Authorization: Bearer {access_token}
```

## 📈 Enhanced Reports

### 1. Get Key Financial Metrics
```http
GET /api/reports/enhanced/financial-metrics
Authorization: Bearer {access_token}
```

### 2. Generate Enhanced Profit & Loss Statement
```http
GET /api/reports/enhanced/profit-loss
Authorization: Bearer {access_token}
Query Parameters: ?start_date=2024-01-01&end_date=2024-01-31
```

### 3. Compare P&L Between Periods
```http
GET /api/reports/enhanced/profit-loss-comparison
Authorization: Bearer {access_token}
Query Parameters: ?period1_start=2024-01-01&period1_end=2024-01-31&period2_start=2023-01-01&period2_end=2023-01-31
```

## 🔒 Security

### 1. Get System Alerts
```http
GET /api/v1/admin/security/alerts
Authorization: Bearer {access_token}
```

### 2. Acknowledge System Alert
```http
PUT /api/v1/admin/security/alerts/{id}/acknowledge
Authorization: Bearer {access_token}
```

### 3. Cleanup Old Security Logs
```http
POST /api/v1/admin/security/cleanup
Authorization: Bearer {access_token}
```

### 4. Get Security Configuration
```http
GET /api/v1/admin/security/config
Authorization: Bearer {access_token}
```

### 5. Get Security Incidents
```http
GET /api/v1/admin/security/incidents
Authorization: Bearer {access_token}
```

### 6. Get Security Incident Details
```http
GET /api/v1/admin/security/incidents/{id}
Authorization: Bearer {access_token}
```

### 7. Resolve Security Incident
```http
PUT /api/v1/admin/security/incidents/{id}/resolve
Authorization: Bearer {access_token}
```

### 8. Get IP Whitelist
```http
GET /api/v1/admin/security/ip-whitelist
Authorization: Bearer {access_token}
```

### 9. Add IP to Whitelist
```http
POST /api/v1/admin/security/ip-whitelist
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "ip_address": "192.168.1.100",
  "description": "Office IP"
}
```

### 10. Get Security Metrics
```http
GET /api/v1/admin/security/metrics
Authorization: Bearer {access_token}
```

## 📊 Dashboard

### 1. Get Analytics Data
```http
GET /dashboard/analytics
Authorization: Bearer {access_token}
```

### 2. Get Quick Statistics
```http
GET /dashboard/quick-stats
Authorization: Bearer {access_token}
```

### 3. Get Dashboard Summary
```http
GET /dashboard/summary
Authorization: Bearer {access_token}
```

## 🔍 Journal Drilldown

### 1. Journal Entry Drill-down (POST)
```http
POST /journal-drilldown
Content-Type: application/json
Authorization: Bearer {access_token}

{
  "account_id": 1,
  "start_date": "2024-01-01",
  "end_date": "2024-01-31"
}
```

### 2. Get Active Accounts for Period
```http
GET /journal-drilldown/accounts
Authorization: Bearer {access_token}
Query Parameters: ?start_date=2024-01-01&end_date=2024-01-31
```

### 3. Journal Entry Drill-down (GET)
```http
GET /journal-drilldown/entries
Authorization: Bearer {access_token}
Query Parameters: ?account_id=1&start_date=2024-01-01&end_date=2024-01-31
```

### 4. Get Journal Entry Detail
```http
GET /journal-drilldown/entries/{id}
Authorization: Bearer {access_token}
```

## 🛠️ Testing Tools

### Postman Collection
Untuk kemudahan testing, buat Postman Collection dengan:

1. **Environment Variables:**
   - `base_url`: URL server (contoh: `http://localhost:3000`)
   - `access_token`: Token dari login response

2. **Pre-request Scripts** untuk authentication:
```javascript
// Set token from login response
pm.globals.set("access_token", pm.response.json().access_token);
```

### cURL Examples

#### Login dengan cURL:
```bash
curl -X POST \
  http://localhost:3000/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

#### Get Journal Entries dengan cURL:
```bash
curl -X GET \
  http://localhost:3000/journal-entries \
  -H 'Authorization: Bearer YOUR_ACCESS_TOKEN'
```

## 🌍 Environment Variables

Buat file `.env.test` untuk environment testing:

```env
# Server Configuration
PORT=3000
BASE_URL=http://localhost:3000

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=accounting_test
DB_USER=test_user
DB_PASSWORD=test_password

# JWT
JWT_SECRET=your_test_jwt_secret
JWT_EXPIRES_IN=24h
JWT_REFRESH_EXPIRES_IN=7d
```

## 📝 Testing Scenarios

### Scenario 1: Complete Journal Entry Flow
1. Login user
2. Create journal entry
3. Get journal entry details
4. Post journal entry
5. Verify in journal entries list

### Scenario 2: Cash Bank Management
1. Create cash/bank account
2. Process deposit
3. Process withdrawal
4. Check balance summary
5. Reconcile account

### Scenario 3: Payment Processing
1. Create payable payment
2. Create receivable payment
3. Get payment analytics
4. Export payment report

### Scenario 4: Security Monitoring
1. Check security alerts
2. Review security incidents
3. Monitor system metrics

## ⚠️ Important Notes

1. **Authentication**: Semua endpoint (kecuali `/auth/*`) memerlukan Bearer token
2. **Error Handling**: Perhatikan response code dan error messages
3. **Data Validation**: Pastikan data yang dikirim sesuai dengan format yang diperlukan
4. **Rate Limiting**: Beberapa endpoint mungkin memiliki rate limiting
5. **Permissions**: Pastikan user memiliki permission yang sesuai untuk endpoint tertentu

## 🔧 Troubleshooting

### Common Issues:

1. **401 Unauthorized**: Token expired atau tidak valid
   - Solution: Login ulang untuk mendapatkan token baru

2. **403 Forbidden**: User tidak memiliki permission
   - Solution: Gunakan user dengan role yang sesuai

3. **422 Unprocessable Entity**: Data validation error
   - Solution: Periksa format data yang dikirim

4. **500 Internal Server Error**: Server error
   - Solution: Periksa log server dan database connection

## 📊 Testing Checklist

- [ ] Authentication flow (register, login, refresh, validate)
- [ ] Journal entries CRUD operations
- [ ] Cash bank account management
- [ ] Payment processing
- [ ] Balance monitoring
- [ ] Report generation
- [ ] Security features
- [ ] Dashboard data
- [ ] Error handling
- [ ] Permission checks

---

**Happy Testing! 🚀**

Untuk pertanyaan atau issues, silakan buka issue di repository ini atau hubungi tim development.