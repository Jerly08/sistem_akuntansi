# Fix: SSOT Profit/Loss Report 404 Error

## Masalah yang Ditemukan

### 1. Double Slash di API Endpoint (404 Error)
**Error Log:**
```
GET "//api/v1/reports/ssot-profit-loss?start_date=2025-01-01&end_date=2025-12-31&format=json"
404 Not Found
```

**Root Cause:**
- Frontend membuat URL dengan double slash (`//`) karena concatenation issue
- Backend route: `/api/v1/reports/ssot-profit-loss`
- Frontend generated: `//api/v1/reports/ssot-profit-loss`
- Double slash menyebabkan route tidak match

**Solusi:**
File: `frontend/src/services/ssotProfitLossService.ts`
```typescript
// Sebelum:
const url = `${API_BASE_URL}${API_V1_BASE}/reports/ssot-profit-loss${queryString ? '?' + queryString : ''}`;

// Sesudah:
const baseUrl = API_BASE_URL.endsWith('/') ? API_BASE_URL.slice(0, -1) : API_BASE_URL;
const url = `${baseUrl}${API_V1_BASE}/reports/ssot-profit-loss${queryString ? '?' + queryString : ''}`;
```

### 2. Activity Logger Error (NULL Constraint Violation)
**Error Log:**
```
ERROR: null value in column "user_id" of relation "activity_logs" violates not-null constraint (SQLSTATE 23502)
```

**Root Cause:**
- Request dari endpoint yang tidak ter-autentikasi (anonymous user)
- Activity logger middleware mencoba insert dengan `user_id = NULL`
- Database constraint: `user_id NOT NULL`
- Model Go sudah benar (`*uint` = nullable) tapi database constraint belum diupdate

**Solusi:**

#### Opsi 1: Via SQL Script (Tercepat)
```bash
# Connect to PostgreSQL and run:
psql -U your_username -d accounting_db -f fix_activity_logs_quick.sql
```

Atau langsung via psql:
```sql
ALTER TABLE activity_logs ALTER COLUMN user_id DROP NOT NULL;
```

#### Opsi 2: Via Go Script
```bash
go run cmd/scripts/fix_activity_logs_constraint.go
```

#### Opsi 3: Via PowerShell Script
```powershell
.\fix_activity_logs.ps1
```

## Verifikasi Fix

### 1. Cek Frontend Fix
Restart frontend dan coba generate P/L report:
```bash
cd frontend
npm run dev
```

Check browser console - URL harus benar tanpa double slash:
```
Making SSOT Profit Loss request to: http://localhost:8080/api/v1/reports/ssot-profit-loss?start_date=...
```

### 2. Cek Database Fix
```sql
-- Verify column is now nullable
SELECT column_name, is_nullable, data_type 
FROM information_schema.columns 
WHERE table_name = 'activity_logs' 
  AND column_name = 'user_id';

-- Should return: is_nullable = 'YES'
```

### 3. Test End-to-End
1. Login ke aplikasi
2. Buka halaman Reports
3. Klik "Generate SSOT P/L Report"
4. Pilih date range
5. Klik "Generate"
6. Report harus berhasil di-generate tanpa error

## Files Changed

### Frontend
- `frontend/src/services/ssotProfitLossService.ts` - Fixed double slash in URL construction

### Backend
- ✅ Model sudah benar (`models/activity_log.go` - UserID is `*uint`)
- ✅ Migration files sudah ada:
  - `migrations/039_fix_nullable_user_id_logs.sql`
  - `migrations/fix_activity_logs_user_id_nullable.sql`
- ✅ Fix script sudah ada: `cmd/scripts/fix_activity_logs_constraint.go`

### New Files
- `backend/fix_activity_logs.ps1` - PowerShell script untuk fix constraint
- `backend/fix_activity_logs_quick.sql` - SQL script untuk fix constraint
- `backend/FIX_PL_REPORT_ERROR.md` - Dokumentasi ini

## Langkah Deployment

1. **Deploy Frontend Fix:**
   ```bash
   cd frontend
   git pull
   npm install
   npm run build
   # Restart frontend service
   ```

2. **Fix Database:**
   ```bash
   cd backend
   # Option A: Run Go script
   go run cmd/scripts/fix_activity_logs_constraint.go
   
   # OR Option B: Run SQL directly
   psql -U your_username -d accounting_db -c "ALTER TABLE activity_logs ALTER COLUMN user_id DROP NOT NULL;"
   ```

3. **Restart Backend:**
   ```bash
   # Restart backend service (method depends on your deployment)
   systemctl restart accounting-backend
   # or
   pm2 restart accounting-backend
   ```

4. **Test:**
   - Try generating P/L report
   - Check logs for any errors
   - Verify anonymous user logging works

## Prevention

Untuk mencegah masalah serupa di masa depan:

1. **URL Construction:**
   - Selalu normalize base URL sebelum concatenation
   - Use helper function untuk URL building
   - Add unit tests untuk URL construction

2. **Database Migrations:**
   - Ensure migrations are run during deployment
   - Add migration status check to startup
   - Document required manual steps

3. **Testing:**
   - Add E2E test untuk anonymous user scenarios
   - Test report generation in CI/CD
   - Monitor activity logger errors

## Status
- ✅ Root cause identified
- ✅ Frontend fix applied
- ⏳ Database fix pending (run migration)
- ⏳ Testing pending
