# Script Reset Transaksi dan Jurnal

Script ini digunakan untuk mereset semua data transaksi dan jurnal dalam sistem akuntansi, namun **mempertahankan data seed** yang didefinisikan dalam `database/seed.go`.

## File yang Tersedia

### 1. `reset_transactions_comprehensive.go` (Versi Terbaik) ⭐
- **RECOMMENDED**: Script yang menangani semua foreign key dependencies
- Tanpa konfirmasi interaktif untuk kemudahan penggunaan
- Menangani purchase receipt items yang sering menyebabkan error
- Memberikan laporan lengkap sebelum dan sesudah reset
- Verifikasi hasil reset untuk memastikan keberhasilan
- **SUDAH TERUJI** bekerja dengan baik

### 2. `reset_transactions.go` (Versi dengan Konfirmasi)
- Memiliki konfirmasi interaktif sebelum menjalankan reset
- Memberikan peringatan lengkap sebelum eksekusi
- Lebih aman untuk environment production
- **Note**: Mungkin error karena masalah AccountingPeriod

### 3. `reset_transactions_simple.go` (Versi Sederhana)
- Tanpa konfirmasi interaktif
- Langsung menjalankan reset
- **Note**: Mungkin error karena masalah AccountingPeriod

### 4. `reset_transactions_minimal.go` (Versi Minimal)
- Koneksi database langsung tanpa inisialisasi kompleks
- **Note**: Mungkin gagal karena foreign key constraint

## Cara Penggunaan

### Method 1: Menggunakan go run (Recommended)
```bash
# Versi terbaik yang sudah teruji (RECOMMENDED)
go run cmd/reset_transactions_comprehensive.go

# Alternatif lain (mungkin error)
go run cmd/reset_transactions.go
go run cmd/reset_transactions_simple.go
go run cmd/reset_transactions_minimal.go
```

### Method 2: Build kemudian jalankan
```bash
# Build versi comprehensive (RECOMMENDED)
go build -o bin/reset_transactions_comprehensive cmd/reset_transactions_comprehensive.go

# Build versi lain
go build -o bin/reset_transactions cmd/reset_transactions.go
go build -o bin/reset_transactions_simple cmd/reset_transactions_simple.go

# Jalankan (pilih yang sudah di-build)
./bin/reset_transactions_comprehensive
```

### Method 3: Menjalankan dari direktori root
```bash
# Dari direktori backend
cd /path/to/app_sistem_akuntansi/backend
go run cmd/reset_transactions_comprehensive.go
```

## Data yang Akan Dihapus

Script ini akan menghapus **SEMUA** data transaksi dan jurnal:

### Jurnal & Transaksi
- ✅ Journal Lines (semua)
- ✅ Journal Entries (semua)
- ✅ Journals (semua)

### Penjualan (kecuali seed)
- ✅ Sale Payments (semua)
- ✅ Sale Return Items (semua)
- ✅ Sale Returns (semua)
- ✅ Sale Items (kecuali yang terkait seed sales)
- ✅ Sales (kecuali SAL-2024-001, SAL-2024-002, SAL-2024-003)

### Pembelian (kecuali seed)
- ✅ Purchase Payments (semua)
- ✅ Purchase Items (kecuali yang terkait seed purchases)
- ✅ Purchases (kecuali PUR-2024-001, PUR-2024-002, PUR-2024-003)

### Pembayaran & Kas Bank
- ✅ Payment Allocations (semua)
- ✅ Payments (semua)
- ✅ Cash Bank Transactions (semua)
- ✅ Cash Bank balances akan direset ke 0

### Pengeluaran
- ✅ Expenses (semua)

## Data yang Dipertahankan (Seed Data)

### Penjualan Seed
- **SAL-2024-001**: Status PAID, Paid Amount: 15,817,500
- **SAL-2024-002**: Status INVOICED, Outstanding: 9,435,000
- **SAL-2024-003**: Status PAID, Paid Amount: 12,920,400

### Pembelian Seed
- **PUR-2024-001**: Status COMPLETED, Paid Amount: 10,878,000
- **PUR-2024-002**: Status APPROVED, Outstanding: 7,215,000
- **PUR-2024-003**: Status COMPLETED, Paid Amount: 8,436,000

### Produk Seed
- **PRD001**: Stock 10 (Laptop Dell XPS 13)
- **PRD002**: Stock 25 (Mouse Wireless Logitech)
- **PRD003**: Stock 100 (Kertas A4 80gsm)

### Master Data (Semua Dipertahankan)
- ✅ Users
- ✅ Accounts (Chart of Accounts)
- ✅ Contacts (Customers & Vendors)
- ✅ Products (dengan stock direset ke nilai seed)
- ✅ Product Categories
- ✅ Product Units
- ✅ Expense Categories
- ✅ Cash Bank Accounts (struktur dipertahankan, balance direset ke 0)
- ✅ Company Profile
- ✅ Report Templates
- ✅ Permissions & Role Permissions
- ✅ Approval Workflows

## Keamanan & Backup

⚠️ **PERINGATAN**: Script ini akan **PERMANENT** menghapus data transaksi!

### Sebelum Menjalankan:
1. **BACKUP DATABASE** terlebih dahulu
2. Pastikan Anda berada di environment yang benar
3. Verifikasi koneksi database dalam `config/config.go`

### Contoh Backup Database:
```bash
# MySQL/MariaDB
mysqldump -u username -p database_name > backup_$(date +%Y%m%d_%H%M%S).sql

# PostgreSQL
pg_dump -U username -d database_name > backup_$(date +%Y%m%d_%H%M%S).sql
```

## Troubleshooting

### Error: "Failed to connect to database"
- Periksa file `.env` atau `config/config.go`
- Pastikan database service berjalan
- Verifikasi username, password, dan nama database

### Error: "Foreign key constraint fails"
- Script sudah menangani foreign key dependencies
- Jika masih terjadi, coba jalankan dengan mode maintenance
- Periksa apakah ada custom foreign key yang tidak tercovered

### Error: Model tidak ditemukan
- Pastikan semua import model sudah benar
- Verifikasi nama model sesuai dengan yang ada di `models/` folder

## Contoh Output (Script Comprehensive)

```
Starting comprehensive transaction and journal reset...
Starting comprehensive reset operation...

Records to be processed:
- Journal Lines: 45
- Journal Entries: 31
- Journals: 0
- Sale Payments: 2
- Sale Return Items: 0
- Sale Returns: 0
- Purchase Payments: 6
- Purchase Receipt Items: 5
- Purchase Receipts: 5
- Payment Allocations: 10
- Payments: 13
- Expenses: 0
- Cash Bank Transactions: 16
- Non-seed Sale Items: 3
- Non-seed Sales: 4
- Non-seed Purchase Items: 11
- Non-seed Purchases: 11

1. Deleting journal lines...
2. Deleting journal entries...
3. Deleting journals...
4. Deleting sale return items...
5. Deleting sale returns...
6. Deleting sale payments...
7. Deleting purchase receipt items (that reference purchase_items)...
8. Deleting purchase receipts...
9. Deleting purchase payments...
10. Deleting payment allocations...
11. Deleting payments...
12. Deleting expenses...
13. Deleting cash bank transactions...
14. Resetting cash bank balances to 0...
15. Deleting sale items (excluding seed data)...
16. Deleting sales (excluding seed data)...
17. Deleting purchase items (excluding seed data)...
18. Deleting purchases (excluding seed data)...
19. Resetting seed sales to original state...
20. Resetting seed purchases to original state...
21. Resetting seed product stocks to original values...
22. Verifying reset results...
23. Reset operation completed!

=== RESET SUMMARY ===
✓ 45 Journal Lines processed
✓ 31 Journal Entries processed
✓ 0 Journals processed
✓ 2 Sale Payments processed
✓ 0 Sale Return Items processed
✓ 0 Sale Returns processed
✓ 6 Purchase Payments processed
✓ 5 Purchase Receipt Items processed
✓ 5 Purchase Receipts processed
✓ 10 Payment Allocations processed
✓ 13 Payments processed
✓ 0 Expenses processed
✓ 16 Cash Bank Transactions processed
✓ Cash Bank balances reset to 0
✓ 3 Non-seed Sale Items processed
✓ 4 Non-seed Sales processed
✓ 11 Non-seed Purchase Items processed
✓ 11 Non-seed Purchases processed

=== VERIFICATION RESULTS ===
Remaining Journal Lines: 0 (should be 0)
Remaining Journal Entries: 0 (should be 0)
Remaining Sale Payments: 0 (should be 0)
Remaining Purchase Payments: 0 (should be 0)
Remaining Payments: 0 (should be 0)
Remaining Cash Bank Transactions: 0 (should be 0)
Remaining Non-seed Sales: 0 (should be 0)
Remaining Non-seed Purchases: 0 (should be 0)

✓ Seed Sales reset to original state
✓ Seed Purchases reset to original state
✓ Seed Product stocks reset to original values

PRESERVED SEED DATA:
- Sales: SAL-2024-001, SAL-2024-002, SAL-2024-003
- Purchases: PUR-2024-001, PUR-2024-002, PUR-2024-003
- Products: PRD001, PRD002, PRD003
- All master data (accounts, contacts, categories, etc.)

Comprehensive transaction and journal reset completed successfully!
```

## Use Cases

### Development Environment (Recommended)
```bash
# Reset untuk testing baru - script terbaik
go run cmd/reset_transactions_comprehensive.go
```

### Staging Environment
```bash
# Script comprehensive juga baik untuk staging
go run cmd/reset_transactions_comprehensive.go

# Atau dengan konfirmasi jika diperlukan (bisa error)
go run cmd/reset_transactions.go
```

### Production Environment
⚠️ **TIDAK DISARANKAN** untuk production. Jika diperlukan:
1. Lakukan full backup database terlebih dahulu
2. Set aplikasi ke maintenance mode  
3. Koordinasi dengan tim
4. Gunakan script comprehensive yang sudah teruji:
```bash
go run cmd/reset_transactions_comprehensive.go
```

## Notes

- Script menggunakan database transaction untuk memastikan atomicity
- Semua operasi menggunakan `Unscoped()` untuk permanent delete (termasuk soft-deleted records)
- Reset akan mempertahankan struktur tabel dan indexes
- Auto-increment IDs tidak direset (bergantung pada database engine)