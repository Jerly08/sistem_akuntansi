# Script Reset Database - Manual Penggunaan

## 🎯 Tujuan

Script `reset_transaction_data_gorm.go` telah ditingkatkan dengan 3 mode operasi:

1. **Mode 1**: Reset Transaksi (Hard Delete) - Default
2. **Mode 2**: Soft Delete Semua Data - ⭐ **FITUR BARU**
3. **Mode 3**: Recovery Semua Soft Deleted Data - ⭐ **FITUR BARU**

---

## 🚀 Cara Menjalankan

```bash
cd D:\Project\app_sistem_akuntansi\backend
go run scripts/maintenance/reset_transaction_data_gorm.go
```

Script akan menanyakan mode operasi yang diinginkan.

---

## 📋 Mode Operasi

### Mode 1: Reset Transaksi (Hard Delete)
**Default mode - sama seperti sebelumnya**

✅ **Yang Dipertahankan:**
- Chart of Accounts (COA)
- Master data produk
- Data kontak/customer/vendor
- Data user dan permission
- Master data cash bank

❌ **Yang Dihapus Permanen:**
- Semua transaksi penjualan/pembelian
- Semua journals (classic & SSOT)
- Payments, inventory movements, expenses
- Notifications, stock alerts
- Balance & stock direset ke 0
- Sequence ID direset

---

### Mode 2: Soft Delete Semua Data ⭐
**Menandai semua data sebagai deleted tanpa menghapus permanen**

✅ **Yang Terjadi:**
- Semua record di tabel dengan kolom `deleted_at` akan diset `deleted_at = NOW()`
- Data **TIDAK** dihapus dari database
- Data tidak akan muncul di aplikasi (karena ORM mengabaikan soft deleted records)
- Dapat dipulihkan kapan saja dengan Mode 3

❌ **Yang Dikecualikan:**
- Tabel backup sistem (`accounts_backup`, dll)
- Tabel migrations

**Keuntungan:**
- **AMAN** - tidak ada data yang hilang permanen
- **REVERSIBLE** - dapat dibatalkan dengan Mode 3
- **CEPAT** - hanya mengupdate kolom `deleted_at`

---

### Mode 3: Recovery Semua Data ⭐
**Memulihkan semua soft deleted data**

✅ **Yang Terjadi:**
- Set `deleted_at = NULL` untuk semua record yang soft deleted
- Data kembali muncul di aplikasi
- Tidak ada data yang hilang

📊 **Fitur:**
- Menampilkan statistik data yang dapat dipulihkan sebelum recovery
- Menghitung jumlah record yang dipulihkan per tabel
- Memberikan laporan lengkap

---

## 🔄 Workflow Umum

### Skenario 1: Testing & Development
```bash
# 1. Soft delete semua data untuk testing
go run scripts/maintenance/reset_transaction_data_gorm.go
# Pilih: 2

# 2. Testing aplikasi dengan data kosong

# 3. Recovery data setelah testing selesai
go run scripts/maintenance/reset_transaction_data_gorm.go
# Pilih: 3
```

### Skenario 2: Reset untuk Demo
```bash
# 1. Hard delete transaksi tapi pertahankan master data
go run scripts/maintenance/reset_transaction_data_gorm.go
# Pilih: 1
```

### Skenario 3: Emergency Recovery
```bash
# Jika ada masalah dan butuh recovery data
go run scripts/maintenance/reset_transaction_data_gorm.go
# Pilih: 3
```

---

## ⚙️ Technical Details

### Tabel yang Memiliki Soft Delete
Script otomatis mendeteksi tabel yang memiliki kolom `deleted_at`:

```sql
SELECT table_name 
FROM information_schema.columns
WHERE table_schema = 'public' 
AND column_name = 'deleted_at'
```

### Query yang Dijalankan

**Soft Delete (Mode 2):**
```sql
UPDATE {table_name} 
SET deleted_at = NOW() 
WHERE deleted_at IS NULL
```

**Recovery (Mode 3):**
```sql
UPDATE {table_name} 
SET deleted_at = NULL 
WHERE deleted_at IS NOT NULL
```

---

## 🛡️ Safety Features

1. **Konfirmasi Ganda**: Script meminta konfirmasi sebelum eksekusi
2. **Transaksi Database**: Semua operasi dalam transaksi (rollback jika ada error)
3. **Exclude Sistem Tables**: Tabel sistem/backup otomatis dikecualikan
4. **Error Handling**: Warning untuk tabel yang gagal diproses
5. **Logging**: Mencatat aktivitas reset ke `audit_logs`

---

## 📊 Output Example

### Mode 2 (Soft Delete)
```
⏳ Menandai soft delete semua data...
   ✅ Soft-deleted: sales (150 records)
   ✅ Soft-deleted: purchases (89 records)
   ✅ Soft-deleted: payments (45 records)
   ...
✅ Soft delete selesai. 15 tabel diproses dalam 2.3s
```

### Mode 3 (Recovery)
```
📊 Data yang dapat dipulihkan:
   - sales: 150 record(s) dapat dipulihkan
   - purchases: 89 record(s) dapat dipulihkan
   ...
🔢 Total: 284 record dapat dipulihkan

⏳ Memulihkan semua soft deleted data...
   ✅ Recovered: sales (150 records)
   ✅ Recovered: purchases (89 records)
   ...
✅ Recovery selesai. 15 tabel diproses, 284 record dipulihkan dalam 1.8s
```

---

## ⚠️ Peringatan & Tips

1. **Backup Database**: Selalu backup database sebelum menjalankan Mode 1 (hard delete)

2. **Mode 2 vs Mode 1**: 
   - Gunakan Mode 2 untuk testing/development
   - Gunakan Mode 1 hanya untuk reset permanen

3. **Konsistensi Data**: Setelah recovery (Mode 3), periksa konsistensi data seperti:
   - Balance accounts
   - Stock products
   - Sequence numbers

4. **Permission**: Pastikan user database memiliki permission untuk UPDATE semua tabel

---

## 🚨 Troubleshooting

### Error: "relation does not exist"
- Tabel belum ada, skip saja (normal)

### Error: "permission denied"
- User database tidak punya akses UPDATE
- Gunakan user dengan privilege lebih tinggi

### Error: "column 'deleted_at' does not exist"
- Tabel tidak support soft delete
- Script otomatis skip tabel ini

---

## 📝 Changelog

### v2.0 (New Features)
- ➕ Mode 2: Soft Delete All
- ➕ Mode 3: Recovery All
- ➕ Auto-detect tables dengan soft delete
- ➕ Statistik recovery
- ➕ Enhanced error handling
- ➕ Exclude sistem tables

### v1.0 (Original)
- Mode 1: Hard Delete Transaksi
- Backup COA
- Reset sequences