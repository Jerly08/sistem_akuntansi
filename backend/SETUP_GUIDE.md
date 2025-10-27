# Backend Setup Guide - Accounting System

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 14 or higher
- Git

## Initial Setup

### 1. Clone & Install Dependencies

```bash
cd backend
go mod download
```

### 2. Database Setup

```sql
-- Create database
CREATE DATABASE accounting_db;

-- Create user (optional)
CREATE USER accounting_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE accounting_db TO accounting_user;
```

### 3. Environment Configuration

Create `.env` file di root folder backend:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=accounting_user
DB_PASSWORD=your_password
DB_NAME=accounting_db
DB_SSLMODE=disable

JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
PORT=8080
```

### 4. Run Backend

```bash
go run main.go
```

**PENTING**: Saat pertama kali run, backend akan otomatis:
- ✅ Create tables (auto-migration)
- ✅ Seed Chart of Accounts (COA) yang diperlukan
- ✅ Seed default data (contacts, invoice types, dll)

## What Gets Seeded Automatically

### 1. Chart of Accounts (COA)

Backend akan create **33 accounts** otomatis, termasuk:

**Asset Accounts:**
- 1101 - KAS
- 1102 - BANK
- 1114 - PPh 21 DIBAYAR DIMUKA *(baru ditambahkan)*
- 1115 - PPh 23 DIBAYAR DIMUKA *(baru ditambahkan)*
- 1116 - POTONGAN PAJAK LAINNYA DIBAYAR DIMUKA *(baru ditambahkan)*
- 1201 - PIUTANG USAHA
- 1301 - PERSEDIAAN BARANG DAGANGAN
- Dan lainnya...

**Liability Accounts:**
- 2103 - PPN KELUARAN
- 2108 - PENAMBAHAN PAJAK LAINNYA
- 292 - PENAMBAHAN PAJAK LAINNYA (SALES) *(baru ditambahkan)*
- Dan lainnya...

**Revenue Accounts:**
- 4101 - PENDAPATAN PENJUALAN
- 4102 - PENDAPATAN JASA/ONGKIR
- 293 - PENDAPATAN ONGKIR (SHIPPING) *(baru ditambahkan)*
- Dan lainnya...

**Expense Accounts:**
- 5101 - HARGA POKOK PENJUALAN
- Dan lainnya...

> **Lihat detail lengkap di**: `docs/COA_SALES_REQUIREMENTS.md`

### 2. Default Data

- Invoice Types (5 types)
- Sample Contacts (customers, suppliers, employees)
- Default module permissions

## Performance Optimizations (NEW!)

Backend sudah dilengkapi dengan optimasi performance:

### Database Index
```sql
-- Jalankan migration untuk index permission
-- File: migrations/add_permission_index.sql

CREATE INDEX idx_module_permission_user_module 
ON module_permission_records(user_id, module);

CREATE INDEX idx_module_permission_user_module_active 
ON module_permission_records(user_id, module) 
WHERE deleted_at IS NULL;
```

**Impact**: Permission check **60-95% lebih cepat** (dari 300-400ms → <10ms)

### In-Memory Permission Cache

- Cache duration: 5 minutes
- Auto-cleanup setiap 5 menit
- Thread-safe dengan sync.RWMutex

## Verification

### 1. Check COA Seeding

```sql
SELECT COUNT(*) FROM accounts WHERE deleted_at IS NULL;
-- Expected: >= 33 accounts
```

### 2. Verify Sales-Required Accounts

```sql
SELECT code, name, type 
FROM accounts 
WHERE code IN ('1114', '1115', '1116', '292', '293', '4101', '5101', '1301', '2103')
  AND deleted_at IS NULL
ORDER BY code;
-- Expected: 9 rows
```

### 3. Test API

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status": "ok"}
```

## Troubleshooting

### "Account XXX not found" saat create sales

**Penyebab**: Seeding belum jalan atau gagal

**Solusi**:
1. Restart backend (seeding akan retry)
2. Atau manual run seed function
3. Check log saat startup untuk error

### Slow Permission Check (>100ms)

**Solusi**: Apply database index

```bash
# Di PostgreSQL
psql -U accounting_user -d accounting_db -f migrations/add_permission_index.sql
```

### Database Connection Error

**Check**:
1. PostgreSQL service running
2. Credentials di `.env` benar
3. Database `accounting_db` sudah dibuat
4. User punya permission ke database

## Development Mode vs Production

### Development (default)
- Auto-migration enabled
- Auto-seeding enabled
- Detailed debug logs
- CORS allow all origins

### Production Setup

Update `.env`:
```env
GIN_MODE=release
DB_SSLMODE=require
JWT_SECRET=use-strong-secret-here-at-least-32-chars
ALLOWED_ORIGINS=https://yourdomain.com
```

Disable auto-seeding di production (optional):
```go
// main.go
if os.Getenv("ENV") != "production" {
    database.SeedAccountsImproved(db)
}
```

## API Documentation

- Base URL: `http://localhost:8080/api/v1`
- Auth: JWT Bearer token
- Content-Type: `application/json`

### Key Endpoints:

**Authentication:**
- `POST /auth/login` - Login
- `POST /auth/register` - Register

**Sales:**
- `POST /api/v1/sales` - Create sale
- `POST /api/v1/sales/:id/invoice` - Confirm invoice
- `GET /api/v1/sales` - List sales

**Accounts:**
- `GET /api/v1/accounts` - List COA
- `POST /api/v1/accounts` - Create account

## Support

Untuk pertanyaan atau issue:
1. Check `docs/COA_SALES_REQUIREMENTS.md`
2. Check log files
3. Raise issue di repository

## Updates Log

**2025-10-27:**
- ✅ Added missing accounts: 1116, 292, 293
- ✅ Performance optimization (permission caching)
- ✅ Database index for permission checks
- ✅ Improved documentation
