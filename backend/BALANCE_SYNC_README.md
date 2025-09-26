# 🔄 Automatic Balance Synchronization System

## 📋 **Overview**

Sistem ini memastikan bahwa account balances di Chart of Accounts (COA) selalu sinkron dengan SSOT (Single Source of Truth) journal entries, mencegah masalah balance yang tidak update setelah transaksi sales invoicing.

## 🎯 **Masalah yang Diselesaikan**

### Masalah Sebelumnya:
- ✅ SSOT journal entries dibuat dengan benar saat sales invoicing
- ❌ Account balances di COA tidak terupdate otomatis
- ❌ Frontend menampilkan balance Rp 0 meskipun ada transaksi
- ❌ PPN account salah hierarchy (di Assets instead of Liabilities)
- ❌ Revenue di-post ke header account bukan detail account

### Solusi Sekarang:
- ✅ Account balances sinkron otomatis dengan SSOT journal entries
- ✅ Database triggers untuk real-time balance updates
- ✅ Periodic integrity checking dan auto-repair
- ✅ Account structure fixes (PPN hierarchy, revenue posting)
- ✅ Robust account resolution via AccountResolver

## 🏗️ **Komponen Sistem**

### 1. **BalanceSyncService** (`services/balance_sync_service.go`)
```go
// Fungsi utama untuk sinkronisasi balance
func (s *BalanceSyncService) SyncAccountBalancesFromSSOT() error
func (s *BalanceSyncService) AutoSyncAfterJournalPost(journalID uint) error
func (s *BalanceSyncService) VerifyBalanceIntegrity() (bool, error)
func (s *BalanceSyncService) SchedulePeriodicSync(intervalMinutes int)
```

### 2. **Account Structure Migration** (`database/fix_account_structure_migration.go`)
- Fixes PPN account hierarchy (moves from Assets to Liabilities)
- Ensures correct account types for all accounts  
- Creates database triggers for automatic balance sync
- Runs initial balance synchronization

### 3. **Database Triggers**
```sql
-- Auto-update account balance saat journal lines berubah
CREATE TRIGGER trigger_update_balance_on_journal_line_insert
AFTER INSERT ON unified_journal_lines
FOR EACH ROW EXECUTE FUNCTION update_account_balance_from_ssot();
```

### 4. **Enhanced SSOT Sales Journal Service**
- Terintegrasi dengan BalanceSyncService
- Auto-sync setelah journal posting
- Uses AccountResolver untuk robust account mapping

## 🚀 **Cara Kerja**

### 1. **Saat Sales Invoice Dibuat:**
```
Sale Created → SSOT Journal Entry → Database Trigger → Account Balance Updated → Parent Balances Updated
```

### 2. **Real-time Synchronization:**
- Database triggers update balances otomatis saat journal lines berubah
- Parent account balances dihitung rekursif dari child accounts
- Balance inconsistencies detected dan diperbaiki otomatis

### 3. **Periodic Monitoring:**
- Integrity check setiap 30 menit
- Auto-repair jika ditemukan inconsistencies
- Logging lengkap untuk troubleshooting

## 📦 **Installation & Setup**

### 1. **Database Migration**
```bash
# Migration akan berjalan otomatis saat startup application
# Atau manual dengan:
go run database/migration.go
```

### 2. **Startup Integration**
```bash
# Menjalankan balance sync service di background
go run startup_integration.go
```

### 3. **Manual Balance Sync** (jika diperlukan)
```bash
# Fix balance inconsistencies secara manual
go run comprehensive_account_fix.go
```

## 🔧 **Configuration**

### Environment Variables
```env
# Database connection (sudah ada)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=accounting_system
DB_USER=postgres
DB_PASSWORD=password

# Balance sync settings (optional)
BALANCE_SYNC_INTERVAL_MINUTES=30
BALANCE_SYNC_ENABLED=true
```

### Account Resolution
```go
// Account types yang di-resolve otomatis
AccountTypeAccountsReceivable = "AR"
AccountTypeSalesRevenue = "REVENUE"  
AccountTypePPNPayable = "PPN_PAYABLE"
```

## 🛠️ **Troubleshooting**

### **Balance Masih Tidak Sinkron?**
```bash
# 1. Check integrity
go run -c "balanceSync.VerifyBalanceIntegrity()"

# 2. Manual sync
go run comprehensive_account_fix.go

# 3. Check database triggers
psql -d accounting_system -c "\df update_account_balance_from_ssot"
```

### **Database Triggers Tidak Berfungsi?**
```sql
-- Re-create triggers
DROP TRIGGER IF EXISTS trigger_update_balance_on_journal_line_insert ON unified_journal_lines;
-- (akan otomatis dibuat ulang oleh migration)
```

### **Account Structure Bermasalah?**
```bash
# Check account hierarchy
go run check_account_hierarchy.go

# Fix structure
go run comprehensive_account_fix.go
```

## 🎯 **Pencegahan Masalah di Masa Depan**

### 1. **Automatic Prevention**
- ✅ Database triggers mencegah balance desync
- ✅ Periodic integrity checks (30 menit)
- ✅ Auto-repair inconsistencies
- ✅ Robust account resolution

### 2. **Git Pull di PC Lain**
- ✅ Migration akan run otomatis saat startup
- ✅ Account structure fix included
- ✅ Balance sync akan jalan otomatis
- ✅ No manual intervention needed

### 3. **Development Environment**
```bash
# Setup baru / fresh database
go run startup_integration.go

# Akan otomatis:
# - Run migration
# - Fix account structure  
# - Sync balances
# - Start monitoring service
```

### 4. **Production Environment**
- Database triggers ensure real-time accuracy
- Periodic monitoring catches any edge cases
- Auto-repair maintains system integrity
- Comprehensive logging for audit trail

## 📊 **Monitoring & Logs**

### Log Examples:
```
🔄 Starting automatic SSOT balance synchronization...
✅ SSOT balance sync completed: 3 accounts updated
🔍 Verifying balance integrity...
✅ All balances are consistent
📝 Created SSOT journal entry 1 for sale 1 with auto-sync
```

### Health Check:
```bash
# Check system health
curl http://localhost:8080/api/balance-sync/health

# Manual integrity check  
curl http://localhost:8080/api/balance-sync/verify
```

## 🎉 **Expected Results**

Setelah implementasi sistem ini:

1. **✅ Sales Invoicing** → COA balances update real-time
2. **✅ Balance Consistency** → SSOT = COA always
3. **✅ Account Structure** → Proper hierarchy & types
4. **✅ Zero Manual Intervention** → Everything automatic
5. **✅ Cross-Environment** → Works on any PC after git pull

## 🤝 **Support**

Jika masih ada issues:
1. Check logs di application startup
2. Run `go run comprehensive_account_fix.go`
3. Verify database triggers exist
4. Contact development team dengan logs

---
*Sistem ini memastikan accounting integrity dan eliminates manual balance corrections permanently.*