# 📚 Tutorial Swagger untuk Pemula

## 🎯 Apa itu Swagger?

Swagger adalah dokumentasi API yang interaktif. Seperti "buku petunjuk" yang bisa langsung dipraktekkan untuk:
- Melihat semua API yang tersedia
- Mencoba API langsung dari browser
- Melihat format data yang diperlukan
- Melihat contoh response

## 🚀 Langkah 1: Menjalankan Backend

Pertama, pastikan backend aplikasi berjalan:

```bash
# Masuk ke folder backend
cd backend

# Jalankan aplikasi Go
go run cmd/main.go
```

Atau jika sudah ada binary:
```bash
./main
```

Backend akan berjalan di: `http://localhost:8080`

## 🌐 Langkah 2: Mengakses Swagger UI

Buka browser dan kunjungi salah satu URL ini:

**Opsi 1 (Utama):**
```
http://localhost:8080/swagger/index.html
```

**Opsi 2 (Alternatif):**
```
http://localhost:8080/docs/index.html
```

## 📖 Langkah 3: Memahami Interface Swagger

### Bagian-bagian Swagger UI:

1. **Header Info**: 
   - Nama API: "Sistema Akuntansi API"
   - Versi: "1.0"
   - Deskripsi project

2. **Base URL**: 
   - Server: `localhost:8080`
   - Base path: `/api/v1`

3. **Categories (Tags)**:
   - 🔐 **Authentication**: Login, register, profile
   - 💰 **CashBank**: Manajemen kas & bank
   - 💳 **Payments**: Pembayaran
   - 📊 **Dashboard**: Dashboard data
   - 📋 **Journal**: Jurnal akuntansi

## 🔧 Langkah 4: Mencoba API (Hands-on)

### A. Login Dulu (Authentication)

1. **Cari section "Authentication"**
2. **Klik endpoint `POST /auth/login`**
3. **Klik tombol "Try it out"**
4. **Isi data login:**
   ```json
   {
     "username": "admin",
     "password": "password123"
   }
   ```
5. **Klik "Execute"**
6. **Copy token dari response** (akan seperti ini):
   ```json
   {
     "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
     "user": {...}
   }
   ```

### B. Authorize Swagger

1. **Klik tombol "Authorize" (🔒) di bagian atas**
2. **Paste token tadi dengan format:**
   ```
   Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```
3. **Klik "Authorize"**
4. **Sekarang Anda sudah bisa mengakses API yang perlu login**

### C. Mencoba API Lain

**Contoh: Melihat Dashboard**
1. Cari section "Dashboard"
2. Klik `GET /dashboard/summary`
3. Klik "Try it out"
4. Klik "Execute"
5. Lihat hasilnya!

## 📝 Langkah 5: Memahami API Documentation

### Untuk setiap endpoint, Anda akan melihat:

1. **HTTP Method**: GET, POST, PUT, DELETE
2. **URL Path**: Alamat endpoint
3. **Parameters**: Data yang perlu dikirim
   - **Path Parameters**: Di URL (contoh: `/users/{id}`)
   - **Query Parameters**: Di URL setelah ? (contoh: `?page=1&limit=10`)
   - **Body Parameters**: Data JSON di request body
4. **Responses**: Contoh response dan status code
5. **Models**: Format data yang digunakan

### Contoh Membaca Endpoint:

```
POST /api/cashbank/deposit

Parameters:
- account_id (integer, required): ID rekening
- amount (number, required): Jumlah uang
- date (string, required): Tanggal (YYYY-MM-DD)
- notes (string, optional): Catatan
```

## 🎯 Tips untuk Pemula

### 1. **Urutan Belajar yang Disarankan:**
   1. Authentication (login dulu)
   2. Dashboard (lihat ringkasan)
   3. CashBank (operasi kas/bank)
   4. Payments (pembayaran)
   5. Journal (jurnal)

### 2. **Yang Perlu Diperhatikan:**
   - Selalu login dulu sebelum mencoba endpoint lain
   - Perhatikan format tanggal: `YYYY-MM-DD` (contoh: `2025-01-15`)
   - Perhatikan tipe data: `integer` untuk angka bulat, `number` untuk decimal
   - Status code 200 = berhasil, 401 = tidak ada akses, 500 = error server

### 3. **Format Data Umum:**
   ```json
   {
     "status": "success",
     "message": "Operation successful",
     "data": { ... }
   }
   ```

### 4. **Troubleshooting Umum:**
   - **401 Unauthorized**: Token expired/invalid → Login ulang
   - **400 Bad Request**: Format data salah → Cek parameter
   - **404 Not Found**: URL salah → Cek endpoint
   - **500 Internal Server Error**: Error server → Cek log backend

## 📋 Cheat Sheet - API yang Sering Dipakai

### Authentication:
```
POST /auth/login          # Login
GET  /profile            # Lihat profile user
```

### Dashboard:
```
GET /dashboard/summary    # Ringkasan dashboard
GET /dashboard/analytics  # Data analytics
```

### CashBank:
```
GET  /api/cashbank/accounts              # Lihat semua rekening
POST /api/cashbank/deposit               # Setor uang
POST /api/cashbank/withdrawal            # Tarik uang
GET  /api/cashbank/balance-summary       # Ringkasan saldo
```

### Payments:
```
GET  /api/payments                       # Lihat pembayaran
POST /api/payments/receivable           # Terima pembayaran
POST /api/payments/payable              # Bayar tagihan
```

## 🔍 Latihan untuk Pemula

1. **Login dan dapat token**
2. **Lihat dashboard summary**
3. **Lihat daftar rekening cashbank**
4. **Coba buat deposit (setor uang)**
5. **Lihat balance summary**
6. **Lihat daftar payments**

## 💡 Pro Tips

1. **Bookmark URL Swagger** untuk akses cepat
2. **Simpan token login** sementara untuk testing
3. **Gunakan Postman** untuk testing lebih advanced
4. **Pelajari satu section dulu** sebelum lanjut ke section lain
5. **Selalu cek response** untuk memahami format data

## 🆘 Butuh Bantuan?

Jika ada error atau bingung:
1. Cek console browser (F12)
2. Lihat log backend
3. Pastikan backend running
4. Cek format data yang dikirim
5. Pastikan sudah login dan authorize

---

**Selamat Belajar! 🎉**
Swagger adalah tool yang sangat powerful untuk memahami dan testing API. Practice makes perfect!