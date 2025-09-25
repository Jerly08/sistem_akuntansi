# 📋 Panduan Script Maintenance Database

Repository ini berisi script-script untuk maintenance dan reset database sistem akuntansi.

## 📁 Daftar Script

### 1. 🔧 `create_account_balances_materialized_view.go`
**Fungsi:** Membuat materialized view `account_balances` yang kompatibel dengan SSOT Journal System.

### 2. 🔄 `reset_transaction_data_gorm.go` 
**Fungsi:** Reset data transaksi dengan berbagai mode (hard delete, soft delete, recovery).

### 3. 🆘 `fix_fresh_database.go`
**Fungsi:** Fix database setelah fresh install - complete migration dan setup semua tabel.

---

## 🚀 Cara Menjalankan Script

### Prerequisite
Pastikan Anda berada di direktori `backend/`:
```bash
cd backend/
```

### A. Script Materialized View

#### **Kapan perlu dijalankan:**
- ✅ Setelah fresh install database
- ✅ Ketika mendapat error: `"account_balances" does not exist`
- ✅ Sebelum menjalankan financial reports
- ✅ Setelah migrasi database besar

#### **Cara menjalankan:**
```bash
go run scripts/maintenance/create_account_balances_materialized_view.go
```

#### **Output yang diharapkan:**
```
🔧 Creating Account Balances Materialized View (SSOT Compatible)
================================================================

🔗 Berhasil terhubung ke database

🗑️ Step 1: Menghapus account_balances yang sudah ada (jika ada)...
   ✅ Cleanup selesai

🔍 Step 2: Memeriksa tabel SSOT...
   ✅ SSOT tables ditemukan - membuat materialized view SSOT

🏗️ Step 3a: Membuat SSOT Materialized View...
   ✅ SSOT Materialized View berhasil dibuat

🔧 Step 4: Membuat index untuk performance...
   ✅ Index berhasil dibuat

🔄 Step 5: Initial refresh materialized view...
   ✅ Materialized view berhasil di-refresh

🧪 Step 6: Testing materialized view...
   📊 Total accounts in view: 34
   💰 Accounts with transactions: 0
   💼 Balance Summary:
      Assets: 0.00
      Liabilities: 0.00
      Equity: 0.00
      Revenue: 0.00
      Expenses: 0.00
   ✅ Balance sheet is balanced (diff: 0.00)
   ✅ Testing selesai

🎉 MATERIALIZED VIEW ACCOUNT_BALANCES BERHASIL DIBUAT!
✅ View sekarang kompatibel dengan SSOT Journal System
✅ Dapat digunakan untuk financial reports
✅ Script reset_transaction_data_gorm.go sekarang akan berfungsi
```

---

### B. Script Fresh Database Fix

#### **Kapan perlu dijalankan:**
- ✅ Setelah client drop dan create ulang database
- ✅ Ketika error: `"column debit_amount does not exist"`
- ✅ Fresh install yang migration belum lengkap
- ✅ Database struktur tidak sesuai dengan code

#### **Cara menjalankan:**
```bash
go run scripts/maintenance/fix_fresh_database.go
```

#### **Output yang diharapkan:**
```
🔧 DATABASE FRESH INSTALL FIX
============================

⚠️  PERINGATAN: Script ini akan memperbaiki database yang baru dibuat.
✅ Yang akan dilakukan:
   - Jalankan complete database migrations
   - Buat SSOT journal system tables
   - Setup materialized views
   - Seed initial data

Lanjutkan? (ketik 'ya' untuk konfirmasi): ya

🔗 Berhasil terhubung ke database

📋 Step 1: Menjalankan database initialization...
   ✅ Database initialization selesai

🔄 Step 2: Menjalankan SSOT migration...
   ✅ SSOT migration berhasil

🏩️ Step 3: Membuat materialized view...
   ✅ Materialized view berhasil dibuat

📊 Step 4: Membuat additional indexes...
   ✅ Indexes berhasil dibuat

🧪 Step 5: Verifikasi struktur database...
   🔧 Adding missing columns to transactions table...
   ✅ Kolom debit_amount dan credit_amount berhasil ditambahkan
   ✅ Materialized view account_balances: 34 records
   ✅ SSOT journal system: 0 entries
   ✅ Verifikasi berhasil - Database siap digunakan

🎉 DATABASE FRESH INSTALL FIX SELESAI!
✅ Database sudah lengkap dan siap digunakan
✅ Semua tabel dan views sudah tersedia
✅ Error 'column does not exist' sudah teratasi
```

---

### C. Script Reset Database

#### **⚠️ PERINGATAN PENTING:**
- Script ini akan **MENGHAPUS DATA TRANSAKSI**
- Pastikan Anda sudah **BACKUP DATABASE** terlebih dahulu
- Jangan jalankan di **PRODUCTION** tanpa persetujuan tim

#### **Mode operasi yang tersedia:**

##### **Mode 1: Reset TRANSAKSI (Hard Delete) - DEFAULT**
- ✅ **DIPERTAHANKAN:** COA, Master produk, Kontak, User, Cash Bank
- ❌ **DIHAPUS:** Semua transaksi, journals, payments, inventory movements

##### **Mode 2: Soft Delete SEMUA data**
- Data ditandai `deleted_at = NOW()` (tidak dihapus permanen)
- Dapat dipulihkan dengan Mode 3

##### **Mode 3: RECOVERY**
- Mengembalikan semua soft deleted data
- Set `deleted_at = NULL`

#### **Cara menjalankan:**

```bash
go run scripts/maintenance/reset_transaction_data_gorm.go
```

#### **Proses interaktif:**

1. **Pilih mode operasi:**
   ```
   Pilih mode operasi yang diinginkan:
     1) Reset TRANSAKSI (hard delete) — mempertahankan master (DEFAULT)
     2) Soft Delete SEMUA data — menandai semua record (deleted_at)
     3) RECOVERY — kembalikan semua soft deleted data

   Masukkan pilihan [1/2/3] (default 1): 1
   ```

2. **Konfirmasi pertama:**
   ```
   Apakah Anda yakin ingin melanjutkan? (ketik 'ya' untuk konfirmasi): ya
   ```

3. **Review data yang akan diproses:**
   ```
   📊 Data saat ini:
      COA Accounts: 34 (akan DIPERTAHANKAN)
      Products: 3 (akan DIPERTAHANKAN, stock direset)
      Sales: 0 (akan DIHAPUS)
      Purchases: 0 (akan DIHAPUS)
   ```

4. **Konfirmasi final:**
   ```
   ⚠️  KONFIRMASI TERAKHIR:
   Ketik 'RESET SEKARANG' untuk melanjutkan: RESET SEKARANG
   ```

#### **Output sukses:**
```
🎉 HARD DELETE RESET SELESAI!
Database siap digunakan dengan COA yang bersih.
Anda bisa mulai input transaksi baru dari 0.
```

---

## 🔧 Troubleshooting

### Error: `"account_balances" does not exist`
**Solusi:** Jalankan script materialized view terlebih dahulu:
```bash
go run scripts/maintenance/create_account_balances_materialized_view.go
```

### Error: `database connection failed`
**Solusi:** 
1. Pastikan PostgreSQL service berjalan
2. Check file `.env` untuk konfigurasi database
3. Pastikan database `sistem_akuntans_test` sudah dibuat

### Error: `package not found`
**Solusi:**
```bash
# Pastikan berada di direktori backend/
cd backend/

# Update dependencies
go mod tidy
go mod download
```

### Error: Permission denied / Access denied
**Solusi:**
1. Jalankan terminal sebagai Administrator (Windows)
2. Pastikan user database memiliki privilege CREATE VIEW

---

## 📊 File Backup yang Dibuat

Setelah menjalankan script reset, file backup akan dibuat:

- `accounts_backup` - Backup tabel accounts
- `accounts_hierarchy_backup` - Backup struktur hierarki COA  
- `accounts_original_balances` - Backup balance asli

### Restore backup (jika diperlukan):
```bash
go run cmd/restore_coa_from_backup.go
```

---

## 💡 Tips & Best Practices

### 1. **Sebelum Reset Database:**
```bash
# 1. Backup database
pg_dump sistem_akuntans_test > backup_$(date +%Y%m%d).sql

# 2. Pastikan materialized view sudah ada
go run scripts/maintenance/create_account_balances_materialized_view.go

# 3. Baru jalankan reset
go run scripts/maintenance/reset_transaction_data_gorm.go
```

### 2. **Setelah Reset Database:**
```bash
# 1. Verify materialized view
psql -d sistem_akuntans_test -c "SELECT COUNT(*) FROM account_balances;"

# 2. Test financial reports di frontend
# 3. Input sample transactions untuk testing
```

### 3. **Development Workflow:**
```bash
# Reset untuk testing
go run scripts/maintenance/reset_transaction_data_gorm.go

# Input test data
# ... (input transactions via frontend/API)

# Reset lagi jika perlu
go run scripts/maintenance/reset_transaction_data_gorm.go
```

---

## 🚨 Peringatan Keamanan

### ❌ **JANGAN LAKUKAN di PRODUCTION:**
- Script reset tanpa backup lengkap
- Mode hard delete tanpa persetujuan tim
- Reset di jam kerja/operasional

### ✅ **LAKUKAN di PRODUCTION:**
- Backup database terlebih dahulu
- Koordinasi dengan tim
- Testing di environment staging dulu
- Dokumentasi perubahan
- Monitoring setelah reset

---

## 📞 Support

Jika mengalami kendala:

1. **Check logs:** Output script memberikan informasi detail
2. **Database logs:** Check PostgreSQL logs untuk error
3. **Contact team:** Koordinasi dengan database administrator
4. **Documentation:** Baca file README dan migration scripts

---

## 📝 Changelog

### v1.0.0 (2025-09-25)
- ✅ Add materialized view creation script
- ✅ Support both SSOT and classic journal systems
- ✅ Comprehensive balance validation
- ✅ Auto-detection of available journal tables
- ✅ Enhanced database connection logging

### Previous versions
- Reset transaction data script
- Backup and restore functionality
- Multi-mode operations (hard delete, soft delete, recovery)