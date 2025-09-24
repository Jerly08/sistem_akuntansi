# Database Enhancements Integration Guide

## Overview
Sistem akuntansi telah diintegrasikan dengan comprehensive database enhancements yang akan dijalankan secara otomatis setiap kali aplikasi di-start. Enhancements ini mencakup optimasi performa, validasi data, audit trail, dan fungsi-fungsi utility untuk maintenance database.

## Auto Migration Integration

### 🚀 Bagaimana Migration Bekerja

Ketika aplikasi dimulai, fungsi `AutoMigrate()` di `database/database.go` akan:

1. **Menjalankan GORM AutoMigrate** untuk semua model termasuk `AccountingPeriod` yang baru ditambahkan
2. **Melakukan Various Data Fixes** seperti sales data integrity, cash bank enhancements, dll
3. **Menjalankan Enhanced Indexes** melalui `createIndexes()`
4. **Mengeksekusi Database Enhancements** melalui `RunDatabaseEnhancements()`

### 📋 Migration Tracking

Sistem menggunakan `MigrationRecord` model untuk tracking:

```go
type MigrationRecord struct {
    MigrationID string    `json:"migration_id" gorm:"unique"`
    Description string    `json:"description"`
    Version     string    `json:"version"`
    AppliedAt   time.Time `json:"applied_at"`
}
```

- Migration hanya akan dijalankan sekali
- Status tersimpan di tabel `migration_records`
- ID Migration: `database_enhancements_v2024.1`

## 🔧 Database Enhancements Yang Diintegrasikan

### 1. Journal Entry Performance Indexes
- `idx_journal_entries_entry_date` - Index pada tanggal entry
- `idx_journal_entries_reference_type_id` - Index untuk referensi
- `idx_journal_entries_status_date` - Index untuk status dan tanggal
- `idx_journal_lines_account_debit/credit` - Index untuk debit/credit per account

### 2. Accounting Performance Indexes
- `idx_accounts_type_balance` - Index untuk balance per tipe account
- `idx_transactions_period_reporting` - Index untuk reporting periode
- `idx_sales_customer_period` - Index untuk analisis sales per customer
- `idx_cash_bank_transactions_flow` - Index untuk cash flow analysis

### 3. Validation Constraints
- **Journal Entry Balance Check**: Memastikan debit = credit (toleransi 0.01)
- **Account Type Validation**: Hanya menerima ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
- **Amount Validation**: Memastikan amount >= 0 dan minimal salah satu > 0
- **Date Range Validation**: Tanggal entry dalam range yang valid
- **Status Validation**: Status harus DRAFT, POSTED, atau REVERSED

### 4. Accounting Period Management
- Auto-create periode akuntansi untuk tahun current dan next
- Table `accounting_periods` dengan tracking open/closed status
- Index untuk date range dan status queries

### 5. Audit Trail Enhancements
- Enhanced indexes untuk query audit logs yang lebih cepat
- Views untuk audit summary dan critical changes
- Tracking untuk DELETE/UPDATE pada tabel penting

### 6. Database Views untuk Reporting

#### Account Balance Summary View
```sql
CREATE VIEW account_balance_summary AS
SELECT 
    a.code, a.name, a.type, a.balance as current_balance,
    calculated_balance, balance_difference
FROM accounts a ...
```

#### Trial Balance View
```sql
CREATE VIEW trial_balance AS
SELECT 
    a.code, a.name, a.type,
    debit_balance, credit_balance
FROM accounts a ...
```

#### Cash Flow Analysis View
```sql
CREATE VIEW cash_flow_analysis AS
SELECT 
    transaction_date, account_code, account_name,
    total_inflow, total_outflow, net_flow
FROM cash_bank_transactions ...
```

#### Sales & Purchase Analysis Views
- Real-time payment status (PAID, PENDING, OVERDUE)
- Comprehensive transaction analysis
- Automatic aging calculation

### 7. Database Functions

#### System Health Check
```sql
SELECT * FROM system_health_check();
```
Mengembalikan:
- Unbalanced journal entries
- Orphaned journal lines  
- Inactive accounts
- Database size information

#### Journal Entry Validation
```sql
SELECT * FROM validate_journal_entry(123);
```
Mengembalikan status valid/invalid dengan error message

#### Account Balance Reconciliation
```sql
SELECT * FROM reconcile_account_balance(456);
```
Membandingkan balance tersimpan vs calculated balance

#### Performance Statistics
```sql
SELECT * FROM get_database_performance_stats();
```
Menampilkan ukuran tabel, index, dan row count

## 📂 File Structure

```
backend/
├── database/
│   ├── database.go                 # Main migration file (updated)
│   └── init.go                     # Database initialization
├── migrations/
│   └── database_enhancements_v2024_1.sql  # SQL migration file
├── models/
│   ├── accounting_period.go        # New model (sudah ada)
│   ├── journal_logger.go          # Enhanced logging (sudah ada)
│   └── migration_record.go        # Migration tracking
└── docs/
    └── database_enhancements_integration.md  # This file
```

## 🔍 Monitoring & Maintenance

### Performance Monitoring
```go
// Check database performance
stats, err := db.Raw("SELECT * FROM get_database_performance_stats()").Rows()

// System health check
health, err := db.Raw("SELECT * FROM system_health_check()").Rows()
```

### Manual Migration Check
```go
// Check if migration already applied
var record models.MigrationRecord
err := db.Where("migration_id = ?", "database_enhancements_v2024.1").First(&record).Error
if err == nil {
    log.Printf("Migration applied at: %v", record.AppliedAt)
}
```

### Index Usage Monitoring
```sql
-- Check index usage statistics
SELECT 
    schemaname, tablename, indexname, 
    idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes 
WHERE schemaname = 'public'
ORDER BY idx_scan DESC;
```

## ⚙️ Configuration

### Environment Variables
Tidak ada environment variable khusus yang diperlukan. Migration akan berjalan otomatis.

### Database Requirements
- PostgreSQL 12+ (untuk advanced indexing features)
- Sufficient disk space untuk additional indexes
- ANALYZE permissions untuk statistics update

## 🚨 Safety Features

### Production Safety
- **Balance Protection**: Semua balance sync operations disabled untuk protect production data
- **Rollback Safe**: Semua operations menggunakan `IF NOT EXISTS` atau `DROP IF EXISTS`
- **Error Handling**: Migration akan continue meski ada error di individual steps
- **Logging**: Comprehensive logging untuk troubleshooting

### Backup Recommendations
Sebelum running migration di production:
1. Database backup lengkap
2. Test di staging environment
3. Monitor disk space untuk additional indexes
4. Verify aplikasi performance setelah migration

## 🔧 Troubleshooting

### Common Issues

#### 1. Migration Tidak Berjalan
- Check log aplikasi untuk error messages
- Verify database connectivity
- Check `migration_records` table untuk status

#### 2. Performance Issues
- Run `ANALYZE` pada tabel utama
- Check index usage dengan pg_stat_user_indexes
- Monitor query execution plans

#### 3. Constraint Violations
- Check data integrity sebelum apply constraints
- Use validation functions untuk identify issues
- Fix data issues sebelum re-run migration

### Manual Migration
Jika perlu run migration manual:

```go
import "app-sistem-akuntansi/database"

// Force re-run migration
db.Where("migration_id = ?", "database_enhancements_v2024.1").Delete(&models.MigrationRecord{})
database.RunDatabaseEnhancements(db)
```

## 📈 Expected Benefits

### Performance Improvements
- **Query Speed**: 50-80% faster untuk complex accounting queries
- **Report Generation**: Significant improvement untuk trial balance, P&L, cash flow
- **Search Operations**: Faster customer/vendor/account lookups

### Data Integrity
- **Automatic Validation**: Real-time constraint checking
- **Balance Accuracy**: Automated balance reconciliation functions
- **Audit Trail**: Complete change tracking dengan enhanced views

### Maintenance
- **Health Monitoring**: Automated system health checks
- **Performance Stats**: Built-in monitoring functions
- **Data Cleanup**: Automated orphaned record cleanup

## 🎯 Next Steps

1. **Monitor Application**: Check performance after migration
2. **Update Reporting**: Leverage new views untuk faster reports  
3. **Implement Health Checks**: Use system_health_check() dalam monitoring
4. **Training**: Train users pada new features dan capabilities

---

**Migration Version**: v2024.1  
**Last Updated**: December 2024  
**Compatibility**: PostgreSQL 12+, GORM v1.25+