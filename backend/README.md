# Accounting Backend

Backend API untuk sistem akuntansi dengan fitur lengkap termasuk SSOT (Single Source of Truth) Journal System.

## 🚀 Quick Start

### Prerequisites
- Go 1.19+
- PostgreSQL 13+
- Database `sistem_akuntans_test` sudah dibuat

### 1. Setup Environment (Untuk PC Baru)

Setelah `git clone` atau `git pull` di PC baru, pilih salah satu cara:

#### Opsi A: Script Otomatis (Recommended)

**Windows (PowerShell):**
```powershell
# Masuk ke direktori backend
cd backend

# Jalankan setup script
.\setup_environment.ps1
```

**Linux/Mac (Bash):**
```bash
# Masuk ke direktori backend
cd backend

# Jalankan setup script
./setup_environment.sh
```

#### Opsi B: Manual Step-by-Step

```bash
# Masuk ke direktori backend
cd backend

# Jalankan migration fixes (WAJIB untuk PC baru)
go run cmd/fix_migrations.go
go run cmd/fix_remaining_migrations.go

# Verifikasi setup berhasil
go run cmd/final_verification.go
```

### 2. Jalankan Backend

```bash
go run cmd/main.go
```

Backend akan berjalan di:
- **API**: http://localhost:8080/api/v1
- **Swagger Docs**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/api/v1/health

## 🔧 Database Configuration

Pastikan PostgreSQL connection string sudah benar:
```
postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable
```

## 📝 Migration Scripts

### Apa itu Migration Fixes?

Migration fixes adalah script untuk mengatasi masalah kompatibilitas database dan memastikan SSOT Journal System berjalan dengan baik. Script ini:

- ✅ Membuat tabel `purchase_payments` yang missing
- ✅ Membuat materialized view `account_balances` untuk SSOT
- ✅ Membuat functions untuk sync balance (`sync_account_balance_from_ssot`)
- ✅ Memperbaiki index dan constraint yang bermasalah

### Kapan Perlu Menjalankan?

**WAJIB dijalankan di:**
- ✅ PC baru setelah git clone
- ✅ Environment baru (development/staging/production)
- ✅ Setelah database reset/restore
- ✅ Jika muncul error SSOT Journal System

**TIDAK perlu dijalankan jika:**
- ❌ Sudah pernah dijalankan di PC yang sama
- ❌ Backend sudah berjalan normal tanpa error

### Troubleshooting

Jika backend masih error setelah migration fixes:

```bash
# Cek status database
go run cmd/final_verification.go

# Jika masih ada masalah, coba jalankan ulang
go run cmd/fix_remaining_migrations.go
```

## 🏗️ Build Backend

```bash
docker build --push --platform linux/amd64 -t registry.digitalocean.com/registry-tigapilar/dbm/account-backend:latest .
```

## 📚 API Documentation

Setelah backend running, akses dokumentasi lengkap di:
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **API Endpoints**: 400+ endpoint tersedia
- **Authentication**: JWT-based dengan role permission

## 🛡️ Features

- ✅ **SSOT Journal System** - Single source of truth untuk semua transaksi
- ✅ **Account Balance Sync** - Automatic balance synchronization
- ✅ **Purchase Payment Integration** - Complete purchase-to-payment workflow
- ✅ **Sales Management** - Full sales cycle management
- ✅ **Financial Reporting** - Trial balance, P&L, Balance sheet
- ✅ **Asset Management** - Fixed asset tracking dengan depreciation
- ✅ **Cash Bank Management** - Multi-currency, multi-account
- ✅ **Approval Workflow** - Configurable approval processes
- ✅ **Audit Trail** - Complete transaction logging

---

> **💡 Tips**: Jika mengalami masalah, jalankan `go run cmd/final_verification.go` untuk memastikan semua komponen berjalan dengan benar.
