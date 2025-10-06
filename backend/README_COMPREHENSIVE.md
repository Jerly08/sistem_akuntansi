# 📊 Sistema Akuntansi - Panduan Lengkap

Aplikasi akuntansi komprehensif dengan fitur lengkap untuk manajemen keuangan, inventory, sales, purchases, dan reporting.

## 📋 Daftar Isi
- [Panduan Umum](#-panduan-umum)
- [Setup Pertama Kali](#-setup-pertama-kali)
- [Alur Bisnis](#-alur-bisnis)
- [Modul Utama](#-modul-utama)
- [FAQ](#-faq)
- [Troubleshooting](#-troubleshooting)

## 🚀 Panduan Umum

### Cara Menggunakan Aplikasi

**Sistem ini terdiri dari:**
- **Backend API**: Server aplikasi (Go/Gin) pada port 8080
- **Frontend**: Interface pengguna (React/Next.js) pada port 3000  
- **Database**: PostgreSQL untuk penyimpanan data
- **SSOT Journal System**: Single Source of Truth untuk semua transaksi

### Konsep Dasar

**Chart of Accounts (COA)**
- Semua akun keuangan disimpan dalam struktur hierarkis
- Kode akun menggunakan format standar akuntansi
- Setiap transaksi harus balance (Debit = Credit)

**SSOT Journal System**
- Semua transaksi dicatat dalam journal entries
- Real-time balance calculation
- Automatic balance sync protection

## 🎯 Setup Pertama Kali

### Prerequisites
```bash
# System Requirements
- Go 1.19+
- PostgreSQL 13+
- Node.js 16+ (untuk frontend)
- Database: sistem_akuntans_test
```

### 1. Clone & Setup Database
```bash
git clone [repository_url]
cd accounting_proj/backend

# Buat database PostgreSQL
createdb sistem_akuntans_test
```

### 2. Environment Setup
```bash
# Copy environment file
cp .env.example .env

# Edit .env dengan konfigurasi Anda:
DATABASE_URL=postgres://postgres:password@localhost/sistem_akuntans_test?sslmode=disable
JWT_SECRET=your-secret-key
SERVER_PORT=8080
```

### 3. ⚠️ WAJIB: Balance Protection Setup

**PENTING:** Sistem ini mencegah balance mismatch yang dapat merusak laporan keuangan.

**Windows:**
```bash
setup_balance_protection.bat
```

**Linux/Mac:**
```bash
chmod +x setup_balance_protection.sh
./setup_balance_protection.sh
```

**Manual (jika script gagal):**
```bash
go run cmd/scripts/setup_balance_sync_auto.go
```

### 4. Migration & Startup
```bash
# Jalankan migration fixes
go run cmd/fix_migrations.go
go run cmd/fix_remaining_migrations.go

# Verifikasi setup
go run cmd/final_verification.go

# Jalankan aplikasi
go run cmd/main.go
```

### 5. Akses Aplikasi
- **Backend API**: http://localhost:8080/api/v1
- **Swagger Documentation**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/api/v1/health

## 💼 Alur Bisnis

### 1. Alur Penjualan (Sales)

```mermaid
graph LR
    A[Quotation] --> B[Sales Order]
    B --> C[Stock Validation]
    C --> D[Invoice Generation]
    D --> E[Payment Recording]
    E --> F[Journal Posting]
    F --> G[Receivable Management]
```

**Langkah-langkah:**

1. **Buat Quotation (Opsional)**
   - Input customer dan produk
   - Set harga dan terms
   - Kirim ke customer

2. **Sales Order**
   - Convert quotation atau buat baru
   - Validasi stock availability
   - Status: Draft → Confirmed

3. **Invoice Generation**
   - Generate invoice dari sales order
   - Pilih invoice type dan numbering
   - Automatic journal entry creation

4. **Payment Recording**
   - Record customer payment
   - Multiple payment methods support
   - Auto-update receivables

5. **Journal Integration**
   - Automatic journal entries
   - Real-time balance updates
   - SSOT compliance

### 2. Alur Pembelian (Purchases)

```mermaid
graph LR
    A[Purchase Request] --> B[Vendor Selection]
    B --> C[Purchase Order]
    C --> D[Approval Workflow]
    D --> E[Goods Receipt]
    E --> F[Invoice Matching]
    F --> G[Payment Processing]
    G --> H[Journal Posting]
```

**Langkah-langkah:**

1. **Purchase Request**
   - Input kebutuhan barang/jasa
   - Set budget dan specifications
   - Submit for approval

2. **Approval Workflow**
   - Multi-level approval system
   - Role-based permissions
   - Email notifications

3. **Purchase Order**
   - Create PO after approval
   - Send to vendor
   - Track delivery status

4. **Three-Way Matching**
   - PO vs Goods Receipt vs Vendor Invoice
   - Automated matching process
   - Exception handling

5. **Payment Processing**
   - Integrated payment to purchases
   - Multiple payment terms
   - Vendor payment tracking

### 3. Alur Pembayaran (Payments)

```mermaid
graph LR
    A[Payment Request] --> B[Payment Method]
    B --> C[Approval]
    C --> D[Payment Execution]
    D --> E[Journal Entry]
    E --> F[Reconciliation]
```

**Metode Pembayaran:**

**Receivable Payments (Piutang):**
- Customer payments untuk sales invoices
- Multiple payment methods (Cash, Bank, Card)
- Partial payments support
- Auto-journal posting

**Payable Payments (Hutang):**
- Vendor payments untuk purchase invoices  
- Batch payment processing
- Payment terms tracking
- Cash flow planning

**Cash & Bank Management:**
- Multi-currency support
- Bank reconciliation
- Cash flow reporting
- GL account integration

## 🏗️ Modul Utama

### 1. Chart of Accounts (COA)
- **Fungsi**: Master data akun keuangan
- **Fitur**: Hierarchical structure, account validation, balance tracking
- **Import/Export**: Excel, CSV support

### 2. Master Data
- **Contacts**: Customer, Vendor, Employee management
- **Products**: Inventory management, pricing, categories
- **Locations**: Warehouse, storage locations

### 3. Transactions
- **Sales**: Quotation, Sales Order, Invoice, Returns
- **Purchases**: Purchase Order, Receipts, Returns  
- **Payments**: Receivables, Payables, Cash management

### 4. Reporting & Analytics
- **Financial Reports**: Balance Sheet, P&L, Cash Flow, Trial Balance
- **Operational Reports**: Sales, Purchase, Inventory reports
- **Analytics**: Financial ratios, trends, forecasting

### 5. System Features
- **User Management**: Role-based access control
- **Approval Workflows**: Configurable approval processes
- **Audit Trail**: Complete transaction logging
- **Real-time Updates**: WebSocket-based notifications

## 📊 Fitur Reporting

### Financial Reports
1. **Balance Sheet (Neraca)**
   - Assets, Liabilities, Equity
   - Comparative periods
   - PDF/Excel export

2. **Profit & Loss (Laba Rugi)**
   - Revenue, COGS, Expenses
   - Financial metrics (margins, ratios)
   - Period comparisons

3. **Cash Flow (Arus Kas)**
   - Operating, Investing, Financing activities
   - Direct & Indirect methods
   - Cash position analysis

4. **Trial Balance (Neraca Saldo)**
   - All account balances
   - Debit/Credit validation
   - Period-end closing

### Operational Reports
- Sales Summary & Analytics
- Purchase Analysis  
- Inventory Valuation
- Vendor Performance
- Customer Analysis

### Real-time Dashboard
- Financial KPIs
- Cash position
- Pending approvals
- System alerts

## ❓ FAQ

### Q: Bagaimana cara reset password?
A: Hubungi administrator untuk reset password. Fitur self-reset akan ditambahkan pada update mendatang.

### Q: Apakah mendukung multi-currency?
A: Ya, sistem mendukung multiple currencies dengan exchange rates yang dapat dikonfigurasi.

### Q: Bagaimana cara backup data?
A: Gunakan PostgreSQL backup tools atau hubungi administrator sistem.

### Q: Apakah ada mobile app?
A: Saat ini hanya tersedia web-based application yang responsive untuk mobile.

### Q: Bagaimana cara menambah user baru?
A: Admin dapat menambah user melalui User Management di menu System Settings.

## 🔧 Troubleshooting

### Masalah Umum

**1. Backend tidak bisa start**
```bash
# Cek database connection
go run cmd/final_verification.go

# Jalankan migration fixes
go run cmd/fix_migrations.go
```

**2. Balance mismatch pada laporan**
```bash
# Cek balance health
curl http://localhost:8080/api/v1/admin/balance-health/check

# Auto-heal balance issues
curl -X POST http://localhost:8080/api/v1/admin/balance-health/auto-heal
```

**3. Journal entries tidak balance**
- Pastikan setiap journal entry memiliki total debit = total credit
- Cek SSOT Journal System status
- Gunakan balance monitoring untuk deteksi

**4. Report generation failed**
- Pastikan semua journal entries sudah POSTED
- Cek date range parameters
- Verifikasi user permissions

**5. Performance issues**
```bash
# Cek performance metrics
curl http://localhost:8080/api/v1/monitoring/performance/metrics

# Clear metrics jika perlu
curl -X POST http://localhost:8080/api/v1/monitoring/performance/metrics/clear
```

### Error Codes

| Code | Description | Solution |
|------|-------------|-----------|
| AUTH_001 | Invalid credentials | Check username/password |
| BAL_001 | Balance mismatch | Run auto-heal process |
| PERM_001 | Insufficient permissions | Check user role |
| DB_001 | Database connection failed | Check database status |
| SSOT_001 | Journal system error | Run SSOT verification |

### Kontak Support

Untuk bantuan lebih lanjut:
- **Technical Issues**: Hubungi IT Administrator
- **Business Process**: Hubungi Finance Team
- **System Enhancement**: Submit feature request

---

## 📈 Roadmap & Updates

### Versi Terbaru (v1.0)
- ✅ SSOT Journal System
- ✅ Balance Protection System  
- ✅ Enhanced Reporting
- ✅ Real-time Updates

### Upcoming Features
- 🔄 Mobile Application
- 🔄 Advanced Analytics
- 🔄 API Integrations
- 🔄 Multi-company Support

---

**💡 Tips**: Selalu jalankan Balance Protection Setup setelah git clone atau environment baru untuk mencegah masalah data inconsistency.

**📚 Dokumentasi Tambahan**: 
- [Technical Guide](TECHNICAL_GUIDE.md)
- [API Documentation](API_DOCUMENTATION.md) 
- [Features Guide](FEATURES.md)