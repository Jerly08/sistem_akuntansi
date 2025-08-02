# Database Schema Documentation

## Overview
Sistem Akuntansi Database Schema menggunakan PostgreSQL dengan ORM GORM untuk Go. Database ini dirancang untuk mendukung aplikasi akuntansi lengkap dengan fitur-fitur modern.

## Core Entities

### 1. Users & Authentication
- **users**: Pengguna sistem dengan role-based access
  - Roles: admin, finance, director, inventory_manager, employee, auditor
  - Fields: username, email, password, role, personal info, employment details

### 2. Company Profile
- **company_profiles**: Profil perusahaan
  - Company information, legal details, fiscal year settings

### 3. Chart of Accounts
- **accounts**: Struktur chart of accounts dengan hierarchy
  - Support untuk parent-child relationships
  - Account types: ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
  - Categories untuk setiap type (CURRENT_ASSET, FIXED_ASSET, etc.)

### 4. Transactions & Journals
- **transactions**: Transaksi individual per account
- **journals**: Jurnal umum dengan multiple entries
- **journal_entries**: Detail entry jurnal (debit/credit)

## Business Entities

### 5. Products & Inventory
- **product_categories**: Kategori produk dengan hierarchy
- **products**: Master produk dengan pricing, stock, dan specs
- **inventories**: Movement inventory dengan FIFO tracking

### 6. Contacts Management
- **contacts**: Customer, vendor, dan employee contacts
- **contact_addresses**: Multiple addresses per contact (billing, shipping, mailing)

### 7. Sales & Purchases
- **sales**: Header transaksi penjualan
- **sale_items**: Detail item penjualan
- **purchases**: Header transaksi pembelian  
- **purchase_items**: Detail item pembelian

### 8. Payments & Cash Management
- **payments**: Transaksi pembayaran
- **cash_banks**: Master cash dan bank accounts
- **cash_bank_transactions**: Movement kas dan bank

### 9. Expenses
- **expense_categories**: Kategori biaya dengan hierarchy
- **expenses**: Transaksi biaya/expense

### 10. Fixed Assets
- **assets**: Fixed assets dengan depreciation tracking
- **Features**: Multiple depreciation methods, asset tracking

## Financial Management

### 11. Budgeting
- **budgets**: Master budget per tahun
- **budget_items**: Detail budget per account per bulan
- **budget_comparisons**: Perbandingan budget vs actual

### 12. Financial Reporting
- **reports**: Generated reports dengan metadata
- **report_templates**: Template untuk berbagai jenis laporan
- **financial_ratios**: Calculated financial ratios
- **account_balances**: Historical account balances per period

### 13. Audit & Compliance
- **audit_logs**: Complete audit trail untuk semua perubahan data
- **Features**: User tracking, IP logging, old/new values

## Key Features

### Database Constraints
- Foreign key relationships properly defined
- Unique constraints pada codes dan identifiers
- Check constraints untuk data validation
- NOT NULL constraints untuk required fields

### Indexing Strategy
- Primary keys dan foreign keys terindex otomatis
- Composite indexes untuk query optimization:
  - `(type, category)` pada accounts
  - `(transaction_date, account_id)` pada transactions
  - `(period, status)` pada journals
  - `(date, customer_id)` pada sales
  - Dan masih banyak lagi

### Data Types
- `decimal(20,2)` untuk monetary values
- `decimal(15,4)` untuk ratios dan percentages
- `text` untuk long descriptions
- `time.Time` untuk timestamps dengan timezone support

## Relationships

### One-to-Many Relationships
- User → Sales, Purchases, Journals, etc.
- Account → Transactions, JournalEntries
- Product → SaleItems, PurchaseItems, Inventories
- Contact → Sales (as Customer), Purchases (as Vendor)

### Many-to-Many Relationships
- Implemented through junction tables where needed

### Self-Referential Relationships
- Accounts (parent-child hierarchy)
- ProductCategories (parent-child hierarchy)
- ExpenseCategories (parent-child hierarchy)

## Data Integrity

### Soft Deletes
- Semua entities menggunakan GORM soft delete
- Data tidak pernah benar-benar dihapus untuk audit trail

### Audit Trail
- Semua perubahan data tercatat di audit_logs
- Track user, timestamp, old values, new values

### Referential Integrity
- Foreign key constraints dijaga oleh database
- Cascade delete policies defined untuk data consistency

## Performance Optimization

### Indexing
- Strategic indexes untuk common query patterns
- Composite indexes untuk complex queries

### Database Design
- Normalized design untuk data consistency
- Denormalized fields where needed untuk performance (calculated totals, etc.)

## Migration & Seeding

### Auto Migration
- GORM AutoMigrate untuk schema updates
- Safe schema evolution

### Initial Data
- Company profile setup
- Default chart of accounts (Indonesian standard)
- Default users dengan roles
- Product categories
- Report templates

## Usage

```go
// Initialize database
database.InitializeDatabase(db)

// Create additional indexes
database.CreateIndexes(db)
```

## Notes
- Database schema mengikuti standar akuntansi Indonesia
- Support untuk multi-currency (IDR default)
- Fiscal year support (default: January-December)
- Role-based access control terintegrasi dalam schema
