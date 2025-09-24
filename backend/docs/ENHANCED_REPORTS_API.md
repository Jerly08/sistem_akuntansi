# Enhanced Financial Reports API Documentation

## 🚀 Overview

Sistem laporan keuangan yang telah diperbaiki dan dikembangkan sesuai dengan UI mockup, dengan integrasi penuh terhadap journal entries dan data akuntansi yang ada. API ini menyediakan laporan keuangan yang comprehensive, accurate, dan terintegrasi dengan sistem double-entry bookkeeping.

## 📊 Available Reports

### 1. **Balance Sheet** - Neraca
**Endpoint:** `GET /api/reports/comprehensive/balance-sheet`
- **Purpose:** Laporan posisi keuangan (aset, kewajiban, dan ekuitas)
- **Integration:** Terintegrasi dengan journal entries dan account balances
- **Parameters:**
  - `as_of_date` (optional): Format YYYY-MM-DD, default = today
  - `format` (optional): json|pdf|excel, default = json

### 2. **Profit & Loss Statement** - Laporan Laba Rugi  
**Endpoint:** `GET /api/reports/comprehensive/profit-loss`
- **Purpose:** Laporan laba rugi comprehensive dengan analisis margin
- **Integration:** Data dari journal entries revenue & expense accounts
- **Parameters:**
  - `start_date` (required): Format YYYY-MM-DD
  - `end_date` (required): Format YYYY-MM-DD
  - `format` (optional): json|pdf|excel, default = json

### 3. **Cash Flow Statement** - Laporan Arus Kas
**Endpoint:** `GET /api/reports/comprehensive/cash-flow`
- **Purpose:** Laporan arus kas dari aktivitas operasi, investasi, dan pendanaan
- **Integration:** Data dari cash accounts dan journal entries
- **Parameters:**
  - `start_date` (required): Format YYYY-MM-DD
  - `end_date` (required): Format YYYY-MM-DD
  - `format` (optional): json|pdf, default = json

### 4. **Sales Summary** - Ringkasan Penjualan
**Endpoint:** `GET /api/reports/comprehensive/sales-summary`
- **Purpose:** Analisis penjualan dengan customer dan product performance
- **Integration:** Data dari sales transactions dan journal entries
- **Parameters:**
  - `start_date` (required): Format YYYY-MM-DD
  - `end_date` (required): Format YYYY-MM-DD
  - `group_by` (optional): day|week|month|quarter|year, default = month
  - `format` (optional): json|pdf|excel, default = json

### 5. **Purchase Summary** - Ringkasan Pembelian
**Endpoint:** `GET /api/reports/comprehensive/purchase-summary`
- **Purpose:** Analisis pembelian dengan vendor dan category performance
- **Integration:** Data dari purchase transactions dan journal entries
- **Parameters:**
  - `start_date` (required): Format YYYY-MM-DD
  - `end_date` (required): Format YYYY-MM-DD
  - `group_by` (optional): day|week|month|quarter|year, default = month
  - `format` (optional): json|pdf, default = json

## 🆕 New Enhanced Reports

### 6. **Vendor Analysis** - Analisis Vendor ⭐ NEW
**Endpoint:** `GET /api/reports/comprehensive/vendor-analysis`
- **Purpose:** Analisis performa vendor, payment analysis, dan outstanding payables
- **Features:**
  - Vendor performance scoring
  - Payment efficiency analysis
  - Outstanding payables tracking
  - Top vendors by spend
- **Parameters:**
  - `start_date` (required): Format YYYY-MM-DD
  - `end_date` (required): Format YYYY-MM-DD
  - `format` (optional): json|pdf, default = json

### 7. **Trial Balance** - Neraca Saldo ⭐ NEW
**Endpoint:** `GET /api/reports/comprehensive/trial-balance`
- **Purpose:** Trial balance dengan summary per account type
- **Features:**
  - Account balance verification
  - Account type summaries (Asset, Liability, Equity, Revenue, Expense)
  - Balance validation
  - Debit/Credit totals
- **Parameters:**
  - `as_of_date` (optional): Format YYYY-MM-DD, default = today
  - `format` (optional): json|pdf|excel, default = json

### 8. **General Ledger** - Buku Besar ⭐ NEW
**Endpoint:** `GET /api/reports/comprehensive/general-ledger`
- **Purpose:** Detailed general ledger untuk specific account
- **Features:**
  - Complete transaction history
  - Running balance calculation
  - Monthly summaries
  - Journal entry references
- **Parameters:**
  - `account_id` (required): Account ID
  - `start_date` (required): Format YYYY-MM-DD
  - `end_date` (required): Format YYYY-MM-DD
  - `format` (optional): json|pdf, default = json

### 9. **Journal Entry Analysis** - Analisis Jurnal ⭐ NEW
**Endpoint:** `GET /api/reports/comprehensive/journal-analysis`
- **Purpose:** Comprehensive journal entry analysis dan compliance check
- **Features:**
  - Journal entry statistics
  - Balance compliance check
  - Entry breakdown by type, status, user
  - Largest and unbalanced entries identification
- **Parameters:**
  - `start_date` (required): Format YYYY-MM-DD
  - `end_date` (required): Format YYYY-MM-DD
  - `format` (optional): json|pdf, default = json

### 10. **Financial Dashboard** - Dashboard Keuangan
**Endpoint:** `GET /api/reports/financial-dashboard`
- **Purpose:** Real-time financial metrics dan key performance indicators
- **Integration:** Comprehensive data dari semua modules
- **Parameters:**
  - `start_date` (optional): Format YYYY-MM-DD, default = first day of current month
  - `end_date` (optional): Format YYYY-MM-DD, default = today

## 🔗 Journal Entry Integration Features

### ✅ **Double-Entry Bookkeeping Compliance**
- Semua reports menggunakan data dari `journal_entries` dan `journal_lines` tables
- Balance validation untuk memastikan Debit = Credit
- Account balance calculation berdasarkan normal balance type

### ✅ **Real-time Data Integration**
- Reports menggunakan posted journal entries (`status = 'POSTED'`)
- Account balances calculated real-time dari journal lines
- No dependency pada cached balances

### ✅ **Account Classification**
- Proper account type classification (Asset, Liability, Equity, Revenue, Expense)
- Category-based grouping (Current Asset, Fixed Asset, Operating Expense, etc.)
- Support untuk account hierarchy

### ✅ **Reference Tracking**
- Journal entry references ke source transactions (Sales, Purchase, Payment)
- Transaction traceability dari reports ke original documents
- User audit trail dalam journal entries

## 📈 Sample API Calls

### Balance Sheet Example
```bash
curl -X GET "http://localhost:8080/api/reports/comprehensive/balance-sheet?as_of_date=2024-12-31&format=json" \
     -H "Authorization: Bearer {token}" \
     -H "Content-Type: application/json"
```

### Profit & Loss Example
```bash
curl -X GET "http://localhost:8080/api/reports/comprehensive/profit-loss?start_date=2024-01-01&end_date=2024-12-31&format=json" \
     -H "Authorization: Bearer {token}" \
     -H "Content-Type: application/json"
```

### Vendor Analysis Example
```bash
curl -X GET "http://localhost:8080/api/reports/comprehensive/vendor-analysis?start_date=2024-01-01&end_date=2024-12-31&format=json" \
     -H "Authorization: Bearer {token}" \
     -H "Content-Type: application/json"
```

### Trial Balance Example
```bash
curl -X GET "http://localhost:8080/api/reports/comprehensive/trial-balance?as_of_date=2024-12-31&format=json" \
     -H "Authorization: Bearer {token}" \
     -H "Content-Type: application/json"
```

### General Ledger Example
```bash
curl -X GET "http://localhost:8080/api/reports/comprehensive/general-ledger?account_id=1101&start_date=2024-01-01&end_date=2024-12-31&format=json" \
     -H "Authorization: Bearer {token}" \
     -H "Content-Type: application/json"
```

## 🛡️ Authentication & Authorization

All report endpoints require:
- **Authentication:** JWT Bearer token
- **Authorization:** User must have role: `finance`, `admin`, `director`, or `auditor`

## 📊 Response Structure

### Standard Success Response
```json
{
  "status": "success",
  "data": {
    "company": {
      "name": "PT. Sistema Akuntansi Digital",
      "address": "Jl. Teknologi Digital No. 123",
      "city": "Jakarta",
      "currency": "IDR"
    },
    "generated_at": "2024-12-31T23:59:59Z",
    // Report-specific data...
  }
}
```

### Error Response
```json
{
  "status": "error",
  "message": "Error description",
  "error": "Technical error details (if available)"
}
```

## 🎯 Key Features & Improvements

### ✅ **Accounting Standards Compliance**
- Indonesian accounting standards compliance
- Proper financial statement structure
- Double-entry bookkeeping validation

### ✅ **Data Accuracy & Integrity**
- Real-time calculation dari journal entries
- Balance verification dan validation
- Data consistency checks

### ✅ **Performance Optimization**
- Efficient database queries dengan proper indexing
- Optimized account balance calculations
- Minimal data processing overhead

### ✅ **User Experience**
- Comprehensive error handling
- Clear parameter validation
- Consistent response formats

### ✅ **Audit Trail**
- Complete transaction traceability
- Journal entry references
- User activity tracking

### ✅ **Extensibility**
- Modular report structure
- Easy addition of new report types
- Configurable company profiles

## 🔧 Configuration

### Company Profile Setup
The system will automatically create a default company profile, but you can customize it via environment variables:

```env
COMPANY_NAME="Your Company Name"
COMPANY_ADDRESS="Your Company Address"
COMPANY_CITY="Your City"
COMPANY_PHONE="+62-21-1234567"
COMPANY_EMAIL="info@company.com"
DEFAULT_CURRENCY="IDR"
COMPANY_TAX_NUMBER="12.345.678.9-012.000"
```

## 🚀 Development & Testing

### Testing Reports
1. Ensure you have journal entries in your database
2. Use the provided API endpoints with proper authentication
3. Start with simple date ranges for testing
4. Verify balance calculations manually for accuracy

### Adding New Reports
1. Define new data structures in `enhanced_report_service.go`
2. Implement service methods for data calculation
3. Add controller endpoints in `enhanced_report_controller.go`
4. Register routes in `enhanced_report_routes.go`
5. Add helper methods in `enhanced_report_helpers.go`

## 💡 Best Practices

### For Frontend Integration
- Always handle loading states untuk report generation
- Implement proper error handling untuk failed reports
- Cache report data appropriately untuk better UX
- Use proper date validations before API calls

### For Backend Development
- Always validate account balance calculations
- Ensure proper journal entry status checks
- Implement proper error handling dan logging
- Follow double-entry bookkeeping principles

## 📞 Support

For technical support or questions regarding the Enhanced Reports API:
- Check the application logs for detailed error information
- Verify journal entry data integrity
- Ensure proper account setup dan configuration
- Contact development team for advanced troubleshooting

---

**Note:** This enhanced reporting system fully integrates with your existing journal entry system and follows proper accounting principles. All reports are generated in real-time from posted journal entries, ensuring data accuracy and consistency.