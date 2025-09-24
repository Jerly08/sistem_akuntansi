# 🔧 Fix Login Swagger - Credentials yang Benar

## ❌ **Masalah Anda**
Dari screenshot, Anda menggunakan:
```json
{
  "username": "admin@company.com",  // ❌ INI EMAIL, BUKAN USERNAME
  "password": "password123"         // ✅ Password benar
}
```

## ✅ **Solusi - Gunakan Salah Satu Dari Ini**

### **Opsi 1: Gunakan Username**
```json
{
  "username": "admin",
  "password": "password123"
}
```

### **Opsi 2: Gunakan Email (jika endpoint mendukung)**
```json
{
  "email": "admin@company.com",
  "password": "password123"
}
```

## 🎯 **Langkah Perbaikan di Swagger**

1. **Buka kembali `POST /auth/login`**
2. **Klik "Try it out"**  
3. **Hapus data yang lama**
4. **Masukkan data yang benar:**
   ```json
   {
     "username": "admin",
     "password": "password123"
   }
   ```
5. **Klik "Execute"**

## 📋 **Semua User Default dari Seed**

Berdasarkan file `seed.go`, user yang tersedia:

### 👤 **Admin**
- Username: `admin`
- Email: `admin@company.com`
- Password: `password123`
- Role: `admin`

### 💰 **Finance**
- Username: `finance`
- Email: `finance@company.com`  
- Password: `password123`
- Role: `finance`

### 📦 **Inventory Manager**
- Username: `inventory`
- Email: `inventory@company.com`
- Password: `password123`
- Role: `inventory_manager`

### 👔 **Director**
- Username: `director`
- Email: `director@company.com`
- Password: `password123`
- Role: `director`

### 👨‍💼 **Employee**
- Username: `employee`
- Email: `employee@company.com`
- Password: `password123`
- Role: `employee`

## 🎯 **Test Login Sekarang**

**Coba login dengan:**
```json
{
  "username": "admin",
  "password": "password123"
}
```

Seharusnya akan berhasil dan mendapat response seperti:
```json
{
  "status": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@company.com",
      "role": "admin",
      "first_name": "Admin",
      "last_name": "User"
    }
  }
}
```

## 🔑 **Setelah Berhasil Login**

1. **Copy token** dari response
2. **Klik "Authorize" (🔒) di bagian atas Swagger**
3. **Masukkan:** `Bearer [paste_token_disini]`
4. **Klik "Authorize"**
5. **Sekarang bisa coba endpoint lain!**

---

**Masalah Anda hanya salah credentials - gunakan `username: "admin"` bukan email sebagai username!** 🎉