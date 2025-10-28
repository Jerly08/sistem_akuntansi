# Quick Fix untuk Activity Logs User ID Error

## Masalah
```
ERROR: null value in column "user_id" of relation "activity_logs" violates not-null constraint
```

## Penyebab
Kolom `user_id` di tabel `activity_logs` masih memiliki constraint `NOT NULL`, sehingga tidak bisa menyimpan log untuk anonymous users (users yang belum login).

## Solusi

### Opsi 1: Jalankan via Docker (RECOMMENDED)

Copy dan jalankan command ini di PowerShell:

```powershell
docker exec -i postgres_db psql -U postgres -d accounting_db << 'EOF'
BEGIN;

-- Drop existing foreign key constraint
ALTER TABLE activity_logs DROP CONSTRAINT IF EXISTS fk_activity_logs_user;

-- Make user_id nullable
ALTER TABLE activity_logs ALTER COLUMN user_id DROP NOT NULL;

-- Re-add foreign key constraint that allows NULL values
ALTER TABLE activity_logs ADD CONSTRAINT fk_activity_logs_user 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add comment
COMMENT ON COLUMN activity_logs.user_id IS 'ID of the user who performed the action (NULL for anonymous/unauthenticated users)';

COMMIT;

-- Verify the change
\d activity_logs
EOF
```

### Opsi 2: Jalankan via SQL File

1. Pastikan Anda berada di folder `backend`
2. Jalankan:

```powershell
docker cp fix_activity_logs_user_id_constraint.sql postgres_db:/tmp/fix.sql
docker exec postgres_db psql -U postgres -d accounting_db -f /tmp/fix.sql
```

### Opsi 3: Manual via psql

1. Connect ke database:
```powershell
docker exec -it postgres_db psql -U postgres -d accounting_db
```

2. Jalankan SQL berikut:
```sql
BEGIN;

ALTER TABLE activity_logs DROP CONSTRAINT IF EXISTS fk_activity_logs_user;
ALTER TABLE activity_logs ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE activity_logs ADD CONSTRAINT fk_activity_logs_user 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

COMMENT ON COLUMN activity_logs.user_id IS 'ID of the user who performed the action (NULL for anonymous/unauthenticated users)';

COMMIT;
```

3. Verify dengan:
```sql
\d activity_logs
```

## Verifikasi

Setelah menjalankan fix, cek kolom `user_id`:

```powershell
docker exec postgres_db psql -U postgres -d accounting_db -c "\d activity_logs" | Select-String "user_id"
```

Output seharusnya:
```
user_id | integer |  |  |
```

**TIDAK** ada tulisan `not null` di kolom user_id.

## Restart Backend

Setelah fix diterapkan, restart backend service:

```powershell
# Stop jika masih running
# Kemudian start lagi
go run cmd/main.go
```

## Testing

Test dengan mengakses endpoint tanpa login:
```
GET http://localhost:8080/api/v1/accounts/catalog
```

Error seharusnya sudah hilang dan activity log akan tersimpan dengan `user_id = NULL`, `username = "anonymous"`, dan `role = "guest"`.
