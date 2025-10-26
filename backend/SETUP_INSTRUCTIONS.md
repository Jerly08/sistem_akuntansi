# 🚀 Setup Instructions for New Environment

## Prerequisites
- Go 1.19+ installed
- PostgreSQL 12+ running
- Git installed

## Quick Setup After Git Pull

### 1. Clone/Pull the Repository
```bash
git pull origin main
cd accounting_proj/backend
```

### 2. Configure Database Connection
Create or update `.env` file with your database configuration:

```env
# Option A: Using DATABASE_URL (recommended)
DATABASE_URL=postgres://username:password@localhost:5432/your_db_name?sslmode=disable

# Option B: Using individual variables
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=your_database_name

# Other required settings
SERVER_PORT=8080
JWT_SECRET=your-secret-key-here
```

**Note:** The scripts will automatically detect and use your database configuration from:
1. First priority: `DATABASE_URL` if present
2. Second priority: Individual `DB_*` variables
3. Fallback: Default values (localhost, postgres/postgres, accounting_db)

### 3. Install Go Dependencies
```bash
go mod download
go mod tidy
```

### 4. Apply Database Fixes (IMPORTANT!)
Run this single command to fix all known database issues:

```bash
go run apply_database_fixes.go
```

This script will:
- ✅ Add UUID extension (fixes `uuid_generate_v4()` errors)
- ✅ Remove problematic concurrent refresh trigger (fixes SQLSTATE 55000)
- ✅ Create manual refresh functions for materialized views
- ✅ Automatically detect your database configuration from `.env`

### 5. Run the Application
```bash
go run main.go
```

## Troubleshooting

### If you get "connection refused" error:
1. Check PostgreSQL is running:
   ```bash
   # Windows
   net start postgresql-x64-14
   
   # Linux/Mac
   sudo systemctl start postgresql
   ```

2. Verify your `.env` database settings match your PostgreSQL configuration

### If you get "authentication failed" error:
1. Double-check your database credentials in `.env`
2. Make sure the database user has proper permissions

### If port 8080 is already in use:
```bash
# Windows
netstat -ano | findstr :8080
taskkill /F /PID <PID_NUMBER>

# Linux/Mac
lsof -i :8080
kill -9 <PID_NUMBER>
```

Or change the port in `.env`:
```env
SERVER_PORT=8081
```

### Manual Fix Options
If the combined script fails, you can run fixes individually:

```bash
# Fix 1: Add UUID extension only
go run add_uuid_extension.go

# Fix 2: Fix concurrent refresh error only
go run fix_concurrent_refresh_error.go
```

## Database Configuration Examples

### Local Development
```env
DATABASE_URL=postgres://postgres:postgres@localhost/accounting_dev?sslmode=disable
```

### Docker PostgreSQL
```env
DATABASE_URL=postgres://postgres:admin123@localhost:5432/accounting_prod?sslmode=disable
```

### Remote Database
```env
DATABASE_URL=postgres://user:password@remote-host.com:5432/dbname?sslmode=require
```

## Verification Checklist

After running the setup, verify everything works:

- [ ] Backend starts without errors: `go run main.go`
- [ ] Can create deposit in Cash & Bank module
- [ ] Can create sales invoice without errors
- [ ] No SQLSTATE 55000 errors in logs
- [ ] Frontend connects to backend successfully

## Common Database Names

The scripts will work with any PostgreSQL database. Common names used:
- `sistem_akuntansi` (original)
- `accounting_prod` (production)
- `accounting_dev` (development)
- `accounting_db` (generic)

The script automatically uses the database specified in your `.env` file.

## Support

If you encounter issues after following these steps:
1. Check the error logs carefully
2. Ensure PostgreSQL version is 12 or higher
3. Verify all environment variables are set correctly
4. Try running fixes individually to isolate the problem

## Notes

- The fixes are idempotent (safe to run multiple times)
- Database connection info is read from `.env` automatically
- No hardcoded credentials in the scripts
- Scripts work across Windows, Linux, and macOS