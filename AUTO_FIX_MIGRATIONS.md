# Auto Fix Migrations System

## Deskripsi

Sistem auto-migration ini dirancang untuk menangani masalah-masalah umum yang sering terjadi setelah `git pull`, seperti:

1. **Upload gambar produk tidak terupdate**
2. **Error sales journal entry tidak balance**
3. **Error cash_banks dengan account_id = 0 atau NULL**
4. **Missing columns atau schema issues**

## Fitur Auto Fix

### 1. Auto Fix Migration (`auto_fix_migration_v2.0`)
- ✅ Memastikan kolom `products.image_path` ada dengan ukuran yang tepat
- ✅ Memperbaiki cash_banks records dengan `account_id = 0` atau NULL
- ✅ Membuat account GL otomatis untuk cash_banks yang belum punya
- ✅ Menambahkan kolom yang hilang pada tabel cash_banks

### 2. Sales Balance Fix Migration (`sales_balance_fix_v1.0`)
- ✅ Memperbaiki perhitungan tax yang tidak balance pada sales
- ✅ Memastikan account default untuk sales (Revenue, AR, Tax Payable)
- ✅ Memperbaiki rounding issues pada sales calculations

### 3. Product Image Fix Migration (`product_image_fix_v1.0`)
- ✅ Memastikan kolom `image_path` memiliki spesifikasi yang benar
- ✅ Membersihkan path gambar yang tidak valid
- ✅ Membuat struktur direktori uploads otomatis

### 4. Component Conflict Resolution
- ✅ Menghapus duplikasi `EnhancedPurchaseTable` 
- ✅ Konsolidasi ke `frontend/src/components/purchase/`
- ✅ Update theme compatibility dengan Chakra UI tokens

## Cara Kerja

Semua migration akan berjalan otomatis saat aplikasi startup melalui `database/init.go`. Setiap migration dicatat dalam tabel `migration_records` untuk mencegah eksekusi berulang.

## Keamanan

- ✅ Database transactions untuk semua operasi
- ✅ Rollback otomatis jika ada error  
- ✅ Idempotent (aman dijalankan berulang kali)
- ✅ Detailed logging untuk audit trail
