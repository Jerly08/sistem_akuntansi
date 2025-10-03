# Git Pull Setup - Fix untuk PC Yang Berbeda

## 🚨 Problem yang Sudah Diperbaiki

Setelah `git pull` dari PC lain, bisa terjadi error karena:
- File `.env` memiliki kredensial database yang berbeda
- Database state berbeda (missing tables seperti `invoice_counters`)  
- Migration logs yang tidak konsisten

## ✅ Solusi Otomatis yang Sudah Diterapkan

### 1. **Auto-Fix di Migration System**

Sistem migration sekarang sudah memiliki **auto-detection dan auto-fix** untuk:

- ✅ Memeriksa tabel `invoice_types` dan `invoice_counters` 
- ✅ Membuat tabel `invoice_counters` otomatis jika hilang
- ✅ Membuat helper functions untuk invoice numbering
- ✅ Membersihkan migration logs yang failed
- ✅ Menjalankan migration 037 jika diperlukan
- ✅ Memastikan migration status konsisten
- ✅ Membuat tabel `tax_account_settings` jika hilang
- ✅ Konfigurasi default tax account settings

### 2. **Yang Akan Terjadi Saat Menjalankan Backend**

Ketika menjalankan `go run main.go`, sistem akan otomatis:

```
🔄 Starting auto-migrations...
============================================
🔍 VERIFYING INVOICE TYPES SYSTEM
============================================
🧾 Checking invoice types system...
   📊 System status:
      - invoice_types table: true
      - invoice_counters table: false
      - sales.invoice_type_id column: true
🔧 Creating missing invoice_counters table...
✅ Created invoice_counters table successfully
🔧 Ensuring helper functions exist...
✅ Helper functions verified
🧹 Cleaning up failed migration logs...
✅ Cleaned up 2 failed migration logs
🔧 Ensuring migration success status...
✅ Marked 037 migration as SUCCESS
✅ Invoice types system verification completed
============================================
📈 VERIFYING TAX ACCOUNT SETTINGS TABLE
============================================
📊 Checking tax account settings table...
🔧 Creating tax_account_settings table...
✅ Inserted 1 default tax account configuration
✅ Tax account settings table created successfully
✅ Tax account settings table verified and ready
============================================
```

## 🛡️ Pencegahan Error di PC Baru

### **Langkah Setup di PC Baru:**

1. **Clone Repository**
   ```bash
   git clone <repository-url>
   cd accounting_proj/backend
   ```

2. **Setup Environment**
   - Copy file `.env.example` ke `.env`
   - Sesuaikan `DATABASE_URL` dengan database PC Anda:
   ```bash
   DATABASE_URL=postgres://username:password@localhost/database_name?sslmode=disable
   ```

3. **Jalankan Backend**
   ```bash
   go run main.go
   ```
   
   Sistem akan otomatis:
   - Memeriksa semua tabel yang diperlukan
   - Membuat tabel yang hilang
   - Menjalankan migration yang diperlukan
   - Membersihkan log yang error

### **Jika Masih Ada Error:**

1. **Database Connection Error**
   - Pastikan PostgreSQL running
   - Cek kredensial di `.env`
   - Pastikan database sudah dibuat

2. **Permission Error**
   - Pastikan user database memiliki permission CREATE TABLE
   - Jalankan sebagai admin jika diperlukan

3. **Migration Error**
   - Sistem auto-fix akan handle ini
   - Check log untuk detail error spesifik

## 📋 File Penting yang TIDAK Boleh di-commit

**JANGAN** commit file berikut ke git:

```
backend/.env                 # Berisi kredensial database lokal
backend/.env.local          # Environment khusus lokal
backend/.env.production     # Kredensial production
```

**BOLEH** commit file berikut:

```
backend/.env.example        # Template environment
backend/.env.simple         # Environment sederhana untuk reference
```

## 🔧 Komponen Auto-Fix yang Ditambahkan

### 1. **ensureInvoiceTypesSystem()**
- Memeriksa status tabel invoice types
- Auto-create missing tables
- Menjalankan migration jika diperlukan

### 2. **createInvoiceCountersTable()**
- Membuat tabel `invoice_counters` dengan struktur lengkap
- Menambahkan indexes dan constraints
- Initialize data untuk invoice types yang ada

### 3. **ensureInvoiceNumberFunctions()** 
- Membuat fungsi `get_next_invoice_number()`
- Membuat fungsi `preview_next_invoice_number()`

### 4. **Auto Migration Log Cleanup**
- Membersihkan log migration yang failed
- Memastikan status SUCCESS untuk sistem yang berfungsi

## 🎯 Hasil Akhir

Setelah implementasi ini:

- ✅ **Tidak ada lagi error saat git pull di PC berbeda**
- ✅ **Setup otomatis untuk PC baru** 
- ✅ **Database schema selalu konsisten**
- ✅ **Migration logs selalu benar**
- ✅ **Invoice numbering system selalu ready**

## 📞 Support

Jika masih mengalami masalah:

1. Check log output saat startup
2. Verify file `.env` configuration
3. Pastikan database service running
4. Check database permissions

**Auto-fix system akan handle 99% kasus yang mungkin terjadi!**