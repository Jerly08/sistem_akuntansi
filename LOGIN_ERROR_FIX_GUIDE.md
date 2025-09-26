# 🔧 Panduan Mengatasi Error Login "Unexpected non-whitespace character after JSON"

## 📋 Langkah-langkah Perbaikan

### 1. ✅ **Konfigurasi Environment Variables**

#### **Backend (.env)**
```bash
# Database
DATABASE_URL=postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable

# Server
SERVER_PORT=8080
ENVIRONMENT=development

# CORS - PENTING untuk frontend
ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000

# JWT
JWT_SECRET=your-secret-key-here
JWT_ACCESS_EXPIRY=90m
JWT_REFRESH_EXPIRY=7d
```

#### **Frontend (.env.local)**
```bash
# API URL - HARUS sesuai dengan backend
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### 2. 🚀 **Menjalankan Aplikasi**

#### **Backend (Terminal 1):**
```powershell
cd backend
go run ./cmd
```

#### **Frontend (Terminal 2):**
```powershell
cd frontend
npm run dev
```

### 3. 🧪 **Test Koneksi**

#### **Test Backend API:**
```powershell
# Di folder backend
./test_full_login.ps1
```

**Expected Output:**
```
✅ Backend API works!
✅ access_token present
✅ token (compatibility) present
✅ user object present
✅ success flag present
✅ CORS preflight works!
✅ Frontend is running!
```

### 4. 🧹 **Clear Browser Data**

**Sebelum test login, clear semua data:**

#### **Manual (Chrome/Edge):**
1. Tekan `F12` (Developer Tools)
2. Klik tab **Application/Storage**
3. Klik **Clear storage** atau **Storage** → **Clear site data**
4. Refresh page (`Ctrl + F5`)

#### **Console Commands:**
```javascript
// Di browser console (F12):
localStorage.clear();
sessionStorage.clear();
document.cookie.split(";").forEach(c => {
    document.cookie = c.replace(/^ +/, "").replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/");
});
location.reload(true);
```

### 5. 🔐 **Test Login**

**Credentials:**
- Email: `admin@company.com`  
- Password: `password123`

### 6. 🔍 **Troubleshooting**

#### **Jika masih error, cek:**

1. **Port conflicts:**
   ```powershell
   netstat -ano | findstr :8080  # Backend
   netstat -ano | findstr :3000  # Frontend
   ```

2. **CORS errors di browser console:**
   - Pastikan `ALLOWED_ORIGINS` di backend benar
   - Pastikan `NEXT_PUBLIC_API_URL` di frontend benar

3. **API endpoint check:**
   ```powershell
   # Test manual
   curl -X POST http://localhost:8080/api/v1/auth/login `
     -H "Content-Type: application/json" `
     -d '{"email":"admin@company.com","password":"password123"}'
   ```

#### **Expected API Response:**
```json
{
  "success": true,
  "access_token": "eyJ...",
  "token": "eyJ...",
  "refresh_token": "eyJ...",
  "refreshToken": "eyJ...",
  "user": {
    "id": 1,
    "email": "admin@company.com",
    "role": "admin"
  },
  "message": "Login successful"
}
```

### 7. ⚠️ **Common Issues**

| Error | Solution |
|-------|----------|
| `Unable to connect to server` | Check backend is running on port 8080 |
| `CORS error` | Verify `ALLOWED_ORIGINS` includes `http://localhost:3000` |
| `Invalid credentials` | Use `admin@company.com` / `password123` |
| `JSON parse error` | Clear browser cache and localStorage |
| `Network error` | Check firewall/antivirus blocking ports |

### 8. 🔧 **Reset Complete**

**Jika semua gagal, reset total:**

```powershell
# Stop all processes
taskkill /f /im node.exe 2>$null
taskkill /f /im go.exe 2>$null

# Clear all browser data
# (Use incognito/private browsing)

# Restart backend
cd backend
go run ./cmd

# Restart frontend  
cd ../frontend
npm run dev
```

### 9. ✅ **Verifikasi Final**

**Login berhasil jika:**
- ✅ No console errors
- ✅ Redirect ke `/dashboard`
- ✅ Token tersimpan di localStorage
- ✅ User data loaded

### 10. 📞 **Support**

**Jika masih bermasalah, kirim:**
1. Screenshot browser console (F12)
2. Network tab response dari login API
3. Backend terminal logs
4. Frontend terminal logs

---

## 🎯 **Quick Fix Checklist**

- [ ] Backend running on port 8080
- [ ] Frontend running on port 3000  
- [ ] `.env.local` contains `NEXT_PUBLIC_API_URL=http://localhost:8080`
- [ ] Browser cache cleared
- [ ] Try login dengan `admin@company.com` / `password123`
- [ ] Check browser console for errors

**Error sudah diperbaiki di backend, tinggal setup environment yang benar!** 🚀