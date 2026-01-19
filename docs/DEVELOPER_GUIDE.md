# 📘 Developer Guide - Sistem Akuntansi Modern

Dokumen ini ditujukan untuk developer yang ingin melakukan deployment, konfigurasi, troubleshooting, atau pengembangan lebih lanjut pada aplikasi ini.

## 🚀 1. Deployment & Setup

### Prasyarat
- **Go** 1.23+
- **Node.js** 18+
- **PostgreSQL** 12+
- **Git**

### A. Backend Setup (Go)

1.  **Masuk ke direktori backend:**
    ```bash
    cd backend
    ```

2.  **Install Dependencies:**
    ```bash
    go mod tidy
    ```

3.  **Setup Database:**
    Buat database PostgreSQL baru.
    ```bash
    createdb sistem_akuntansi
    ```

4.  **Konfigurasi Environment:**
    Copy file contoh `.env` dan sesuaikan.
    ```bash
    cp .env.example .env
    ```
    Edit `.env` dan pastikan koneksi database benar:
    ```env
    DATABASE_URL=postgres://username:password@localhost:5432/sistem_akuntansi?sslmode=disable
    # ... konfigurasi lain seperti JWT_SECRET
    ```

5.  **Jalankan Aplikasi (Development):**
    ```bash
    go run cmd/main.go
    ```
    *Catatan: Saat pertama kali dijalankan, backend akan otomatis melakukan migrasi database dan seeding data awal (User Admin, Chart of Accounts, dll).*

6.  **Build untuk Production:**
    ```bash
    go build -o sistem-akuntansi cmd/main.go
    ./sistem-akuntansi
    ```

### B. Frontend Setup (Next.js)

1.  **Masuk ke direktori frontend:**
    ```bash
    cd frontend
    ```

2.  **Install Dependencies:**
    ```bash
    npm install
    ```

3.  **Konfigurasi Environment:**
    Buat file `.env.local` jika perlu mengubah URL API backend.
    ```bash
    echo "NEXT_PUBLIC_API_URL=http://localhost:8080/" > .env.local
    ```

4.  **Jalankan Development Server:**
    ```bash
    npm run dev
    ```
    Akses di `http://localhost:3000`.

5.  **Build untuk Production:**
    ```bash
    npm run build
    npm start
    ```

---

## ⚙️ 2. Konfigurasi Database & Environment

### Backend (`backend/.env`)
Variabel penting yang perlu diperhatikan:

| Variable | Deskripsi | Contoh |
| :--- | :--- | :--- |
| `DATABASE_URL` | Connection string PostgreSQL | `postgres://user:pass@localhost:5432/db?sslmode=disable` |
| `SERVER_PORT` | Port server backend | `8080` |
| `JWT_SECRET` | Secret key untuk token JWT | `rahasia-super-aman-min-32-karakter` |
| `ENVIRONMENT` | Mode aplikasi | `development` atau `production` |
| `ENABLE_SWAGGER`| Mengaktifkan dokumentasi API | `true` |

### Frontend (`frontend/.env.local`)

| Variable | Deskripsi | Default |
| :--- | :--- | :--- |
| `NEXT_PUBLIC_API_URL` | Base URL API Backend | `http://localhost:8080/` |

---

## 🛠️ 3. Troubleshooting & Maintenance Scripts

Jika terjadi error atau masalah data, gunakan script berikut yang tersedia di folder `backend/`:

### Masalah Database / Migrasi
*   **Fix UUID Extension Error:**
    ```bash
    go run apply_database_fixes.go
    ```
*   **Cek Log Migrasi:**
    Jika migrasi gagal, cek tabel `migration_logs` di database.

### Masalah Data Akuntansi
*   **Fix Account Balances:**
    Jika neraca tidak seimbang atau ada anomali saldo.
    ```bash
    go run scripts/maintenance/run_balance_monitor.go
    ```
*   **Fix Duplicate Accounts:**
    ```bash
    go run scripts/maintenance/fix_accounts.go
    ```
*   **Reset Data Transaksi (Hati-hati!):**
    ```bash
    go run scripts/maintenance/reset_transaction_data_gorm.go
    ```

### Masalah Security
*   **Test Security System:**
    ```bash
    go run scripts/test_security_system.go
    ```

---

## 🏗️ 4. Struktur Kode & Pengembangan

### Backend Structure (`backend/`)
Menggunakan **Clean Architecture**.

*   `cmd/main.go`: Entry point aplikasi. Setup server, database, dan middleware.
*   `controllers/`: **Handler HTTP**. Menerima request, validasi input, panggil service, kembalikan response.
    *   *Contoh: Ingin nambah API baru? Buat controller baru di sini.*
*   `services/`: **Business Logic**. Tempat logika utama aplikasi.
    *   *Contoh: Logika hitung pajak, validasi stok, approval workflow ada di sini.*
*   `repositories/`: **Data Access**. Interaksi langsung dengan database (GORM).
    *   *Contoh: Query SQL custom atau CRUD operation.*
*   `models/`: **Struct Database**. Definisi tabel dan relasi.
*   `middleware/`: Auth, Logging, CORS, Security headers.
*   `routes/`: Definisi URL endpoint dan mapping ke controller.

### Frontend Structure (`frontend/`)
Menggunakan **Next.js App Router**.

*   `app/`: Halaman-halaman aplikasi (Routes).
    *   `layout.tsx`: Layout utama (Sidebar, Header).
    *   `globals.css`: Global styles (Tailwind).
*   `src/components/`: Komponen React reusable.
    *   `common/`: Button, Input, Modal, dll.
    *   `reports/`: Komponen khusus laporan.
*   `src/services/`: API Client functions (Axios) untuk panggil backend.
    *   *Contoh: `authService.ts`, `salesService.ts`.*
*   `src/contexts/`: Global state (Auth, Theme, Language).
*   `src/hooks/`: Custom hooks (e.g., `useAuth`, `useTranslation`).

### Panduan Pengembangan Fitur Baru

1.  **Backend:**
    *   Buat **Model** di `models/` (jika perlu tabel baru).
    *   Buat **Repository** interface & implementation di `repositories/`.
    *   Buat **Service** untuk logic di `services/`.
    *   Buat **Controller** di `controllers/`.
    *   Daftarkan **Route** di `routes/`.

2.  **Frontend:**
    *   Buat **Service** di `src/services/` untuk panggil API baru.
    *   Buat **Component** UI di `src/components/`.
    *   Buat **Page** baru di `app/` (jika perlu halaman baru).

---
