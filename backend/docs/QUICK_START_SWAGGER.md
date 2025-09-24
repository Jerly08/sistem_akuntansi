# 🚀 Quick Start: Cara Pakai Swagger

## ✅ Backend Sudah Berjalan!

Backend Anda sudah running di `http://localhost:8080` ✅

## 🌐 Langkah 1: Buka Swagger UI

**Buka browser dan pergi ke:**
```
http://localhost:8080/swagger/index.html
```

**Atau alternatif:**
```
http://localhost:8080/docs/index.html
```

## 🎯 Langkah 2: Yang Akan Anda Lihat

Anda akan melihat halaman Swagger dengan:
- **Judul**: "Sistema Akuntansi API v1.0"
- **Daftar endpoint** yang sudah dibersihkan (hanya 36 endpoint yang aktif dipakai)
- **Beberapa kategori utama**:
  - 🔐 Authentication 
  - 💰 CashBank
  - 💳 Payments
  - 📊 Dashboard
  - 📋 Journal

## 🔥 Langkah 3: Coba Sekarang (5 Menit)

### A. Login Dulu
1. **Scroll ke section "Authentication"**
2. **Klik `POST /auth/login`**
3. **Klik "Try it out"**
4. **Ganti example data dengan:**
   ```json
   {
     "username": "admin",
     "password": "admin123"
   }
   ```
   *(Sesuaikan dengan user yang ada di database Anda)*
5. **Klik "Execute"**
6. **Copy token dari response**

### B. Authorize
1. **Klik tombol hijau "Authorize" di atas**
2. **Paste token dengan format:**
   ```
   Bearer [token_anda_disini]
   ```
3. **Klik "Authorize" dan "Close"**

### C. Coba API
Sekarang coba endpoint lain seperti:
- `GET /dashboard/summary` - Lihat ringkasan data
- `GET /api/cashbank/accounts` - Lihat rekening
- `GET /api/payments` - Lihat pembayaran

## 🎨 Interface Swagger Explained

### Setiap endpoint menampilkan:
```
🔵 GET    - Ambil data (read)
🟢 POST   - Buat data baru (create) 
🟡 PUT    - Update data (update)
🔴 DELETE - Hapus data (delete)
```

### Parameter Types:
- **Path**: Di URL seperti `/users/{id}`
- **Query**: Setelah ? seperti `?page=1&limit=10`
- **Body**: Data JSON yang dikirim

### Response Codes:
- **200** ✅ = Berhasil
- **401** 🔒 = Perlu login
- **400** ❌ = Data salah
- **500** 💥 = Error server

## 📚 API yang Paling Sering Dipakai

### 🔐 Authentication:
```
POST /auth/login     → Login
GET  /profile        → Lihat profile
```

### 💰 CashBank (Kas & Bank):
```
GET  /api/cashbank/accounts         → Lihat semua rekening
POST /api/cashbank/deposit          → Setor uang
POST /api/cashbank/withdrawal       → Tarik uang
GET  /api/cashbank/balance-summary  → Ringkasan saldo
```

### 💳 Payments:
```
GET  /api/payments              → Lihat pembayaran
POST /api/payments/receivable   → Terima pembayaran
POST /api/payments/payable      → Bayar tagihan
```

### 📊 Dashboard:
```
GET /dashboard/summary    → Data ringkasan
```

## ⚡ Tips Cepat

1. **Selalu login dulu** sebelum coba endpoint lain
2. **Format tanggal**: `2025-01-15` (YYYY-MM-DD)
3. **Angka**: Pakai titik untuk desimal `100.50`
4. **Jika error 401**: Login ulang dan update token

## 🎯 Latihan 5 Menit

Coba urutan ini:
1. Login → Dapat token
2. Authorize dengan token
3. `GET /dashboard/summary` → Lihat data dashboard
4. `GET /api/cashbank/accounts` → Lihat rekening
5. `GET /api/payments` → Lihat pembayaran

## 🆘 Troubleshooting

**Swagger tidak bisa dibuka?**
- Pastikan backend running
- Coba refresh browser
- Coba URL alternatif: `http://localhost:8080/docs/index.html`

**Error 401 terus?**
- Periksa username/password
- Pastikan format token: `Bearer [token]`
- Token mungkin expire, login ulang

**API response kosong?**
- Database mungkin kosong
- Coba endpoint lain dulu

---

## 🎉 Selamat!

Swagger sudah siap digunakan! File ini ada di:
`D:\Project\app_sistem_akuntansi\QUICK_START_SWAGGER.md`

**Bookmark URL ini untuk akses cepat:**
`http://localhost:8080/swagger/index.html`