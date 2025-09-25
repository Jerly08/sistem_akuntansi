# 🚀 Quick Reference - Script Maintenance

## 📋 Command Cheat Sheet

### Pastikan di direktori yang benar:
```bash
cd backend/
```

### 1. 🔧 Buat/Fix Materialized View
```bash
go run scripts/maintenance/create_account_balances_materialized_view.go
```

### 2. 🔄 Reset Database (Interactive)
```bash
go run scripts/maintenance/reset_transaction_data_gorm.go
```

### 3. 🆘 Fix Fresh Database (Complete Migration)
```bash
go run scripts/maintenance/fix_fresh_database.go
```

### 4. 🔧 Fix Migration Issues (Clean Error Logs)
```bash
go run scripts/maintenance/fix_migration_issues.go
```

### 3. 💾 Backup Database (Manual)
```bash
# Windows PowerShell
pg_dump -h localhost -U postgres sistem_akuntans_test > "backup_$(Get-Date -Format 'yyyyMMdd').sql"

# Linux/Mac
pg_dump sistem_akuntans_test > backup_$(date +%Y%m%d).sql
```

### 4. 🔍 Check Materialized View
```bash
# Via psql
psql -d sistem_akuntans_test -c "SELECT COUNT(*) FROM account_balances;"

# Via Go (create and run test script)
echo "SELECT COUNT(*) FROM account_balances;" > test.sql
psql -d sistem_akuntans_test -f test.sql
```

---

## ⚡ One-liner Commands

### Complete Reset Flow:
```bash
# 1. Backup, 2. Create view, 3. Reset
pg_dump sistem_akuntans_test > backup.sql && go run scripts/maintenance/create_account_balances_materialized_view.go && go run scripts/maintenance/reset_transaction_data_gorm.go
```

### Quick Fix for "account_balances does not exist":
```bash
go run scripts/maintenance/create_account_balances_materialized_view.go
```

### Development Cycle:
```bash
# Reset -> Test -> Reset
go run scripts/maintenance/reset_transaction_data_gorm.go
# (input test data via frontend)
go run scripts/maintenance/reset_transaction_data_gorm.go
```

---

## 🔧 Common Issues & Solutions

|| Error | Command |
||-------|---------|
|| `column "debit_amount" does not exist` | `go run scripts/maintenance/fix_fresh_database.go` |
|| `relation "unified_journal_ledger" already exists` | `go run scripts/maintenance/fix_migration_issues.go` |
|| `current transaction is aborted` | `go run scripts/maintenance/fix_migration_issues.go` |
|| `account_balances does not exist` | `go run scripts/maintenance/create_account_balances_materialized_view.go` |
|| Fresh database after DROP/CREATE | `go run scripts/maintenance/fix_fresh_database.go` |
|| Migration errors in logs | `go run scripts/maintenance/fix_migration_issues.go` |
|| `database connection failed` | Check `.env` file and PostgreSQL service |
|| `package not found` | `go mod tidy && go mod download` |
|| `permission denied` | Run terminal as Administrator |

---

## 📊 Reset Modes Quick Guide

| Mode | Command Input | Effect |
|------|---------------|---------|
| **1** | `1` + `ya` + `RESET SEKARANG` | Hard delete transactions, keep master data |
| **2** | `2` + `ya` + `RESET SEKARANG` | Soft delete all data (recoverable) |  
| **3** | `3` + `ya` + `RESET SEKARANG` | Recover all soft deleted data |

---

## ⚠️ Safety Checklist

Before running reset scripts:

- [ ] ✅ Database backup created
- [ ] ✅ Team notified (if production)
- [ ] ✅ Materialized view exists
- [ ] ✅ Know which mode to use
- [ ] ✅ Have restore plan ready

---

**💡 Pro Tip:** Keep this file open while running maintenance scripts!