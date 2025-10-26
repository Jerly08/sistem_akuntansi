# Changelog - COA Capitalization Update

## Tanggal: 2025-10-25

### 🔄 Perubahan

#### 1. **Semua Nama COA Default Dijadikan CAPITAL**
- Semua account names pada seed data sekarang menggunakan huruf **CAPITAL** (uppercase)
- Menggunakan `strings.ToUpper()` untuk memastikan konsistensi

#### 2. **Auto-Update Existing Accounts**
- Fungsi `CapitalizeExistingAccounts()` ditambahkan untuk mengupdate existing accounts yang belum capitalize
- Fungsi ini akan otomatis dijalankan setiap kali seeding dilakukan
- Hanya mengupdate nama (name field), tidak mengubah balance atau data lainnya

#### 3. **Pencegahan Duplikasi**
- Sistem tetap menggunakan validasi berdasarkan `code` untuk mencegah duplikasi
- `FirstOrCreate` dengan filter `code` dan `deleted_at IS NULL`
- Jika account sudah ada (berdasarkan code), hanya update metadata (name, type, category, dll) tanpa mengubah balance

### 📋 File yang Diubah

1. **backend/database/account_seed.go**
   - Ditambahkan `strings.ToUpper()` pada semua account names
   - Ditambahkan normalisasi nama di `FirstOrCreate`
   - Ditambahkan fungsi `CapitalizeExistingAccounts()`

2. **backend/database/account_seed_improved.go**
   - Ditambahkan `strings.ToUpper()` pada semua account names
   - Ditambahkan normalisasi nama di fungsi `upsertAccount()`

### 🚀 Cara Kerja

#### Saat Backend Dijalankan:
1. Seeding akan membuat accounts baru dengan nama CAPITAL
2. Jika account sudah ada, akan diupdate nama-nya menjadi CAPITAL
3. Tidak ada duplikasi karena validasi berdasarkan `code`
4. Balance dan data transaksi tetap terjaga

#### Contoh Output Log:
```
🌱 Starting account seeding (idempotent mode)...
✅ Created account: 1000 - ASSETS
✅ Created account: 1100 - CURRENT ASSETS
🔤 Capitalizing existing account names to ensure consistency...
✅ Capitalized: 1101 - Kas → KAS
✅ Capitalized: 1102 - Bank → BANK
✅ Capitalized 2 account names
```

### ✅ Keuntungan

1. **Konsistensi UI**: Semua nama account tampil dengan format yang sama (CAPITAL)
2. **Tidak Ada Data Loss**: Balance dan transaksi tetap aman
3. **Idempoten**: Aman dijalankan berulang kali tanpa membuat duplikasi
4. **Auto-Migration**: Existing database akan otomatis terupdate

### ⚠️ Catatan

- Fungsi ini hanya mengubah **nama account** menjadi CAPITAL
- Tidak mengubah `code`, `balance`, atau data lainnya
- Jika ada account yang sudah CAPITAL, tidak akan diupdate lagi (efisien)
- Duplikasi berdasarkan `code` tetap dicegah seperti sebelumnya

### 🔍 Testing

Untuk memverifikasi tidak ada duplikasi:
```sql
-- Cek duplikasi berdasarkan code
SELECT code, COUNT(*) as count 
FROM accounts 
WHERE deleted_at IS NULL 
GROUP BY code 
HAVING COUNT(*) > 1;

-- Cek nama yang belum capitalize (seharusnya kosong)
SELECT code, name 
FROM accounts 
WHERE deleted_at IS NULL 
AND name != UPPER(name);
```

### 📝 Maintenance

Jika ingin menambah account baru di seed:
1. Gunakan `strings.ToUpper()` untuk nama account
2. Pastikan `code` unik (tidak duplikat)
3. Contoh:
   ```go
   {Code: "1103", Name: strings.ToUpper("Kas Kecil"), Type: models.AccountTypeAsset, ...}
   ```
