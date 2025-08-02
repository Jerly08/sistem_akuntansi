# 📊 Sistem Akuntansi

Aplikasi sistem akuntansi lengkap yang terdiri dari backend API (Go) dan frontend web (Next.js) untuk mengelola keuangan dan operasional bisnis.

## 🏗️ Arsitektur Aplikasi

```
accounting_proj/
├── backend/          # Go REST API dengan Gin & GORM
│   ├── cmd/
│   ├── controllers/
│   ├── models/
│   ├── routes/
│   └── ...
├── frontend/         # Next.js React App dengan Chakra UI
│   ├── app/
│   ├── src/
│   └── ...
├── dummy_data_summary.md
├── Noted.txt
└── README.md
```

## 🚀 Teknologi yang Digunakan

### Backend
- **Go 1.21+** - Programming language
- **Gin** - Web framework
- **GORM** - ORM untuk database
- **PostgreSQL** - Database
- **JWT** - Authentication
- **bcrypt** - Password hashing

### Frontend
- **Next.js 14** - React framework
- **TypeScript** - Type-safe JavaScript
- **Chakra UI** - Component library
- **React Context** - State management
- **Recharts** - Data visualization

## 🌟 Fitur Utama

### 🔐 Authentication & Authorization
- Login/Register sistem
- Role-based access control
- JWT authentication
- Protected routes

### 👥 User Roles
- **Admin** - Full access ke semua fitur
- **Director** - Access ke reports dan read operations
- **Finance** - Access ke financial data dan reports
- **Inventory Manager** - Manage products dan inventory
- **Employee** - Basic access

### 📊 Modul Bisnis
- **Dashboard** - Overview dan statistik
- **Products** - Manajemen produk dan inventori
- **Sales** - Penjualan dan invoicing
- **Purchases** - Pembelian dari supplier
- **Expenses** - Pengeluaran operasional
- **Assets** - Manajemen aset perusahaan
- **Cash & Bank** - Manajemen kas dan bank
- **Reports** - Laporan keuangan
- **Users** - Manajemen pengguna

## 🛠️ Quick Start

### Prerequisites
- **Node.js 18+** dan npm/yarn
- **Go 1.21+**
- **PostgreSQL 12+**
- **Git**

### 1. Clone Repository
```bash
git clone https://github.com/dbm-main/accounting_proj.git
cd accounting_proj
```

### 2. Setup Backend
```bash
cd backend

# Install dependencies
go mod tidy

# Setup database PostgreSQL
createdb sistem_akuntansi

# Copy dan edit environment variables
cp .env.example .env
# Edit .env sesuai konfigurasi database Anda

# Jalankan server
go run cmd/main.go
```
Backend akan berjalan di `http://localhost:8080`

### 3. Setup Frontend
```bash
cd frontend

# Install dependencies
npm install
# atau
yarn install

# Jalankan development server
npm run dev
# atau
yarn dev
```
Frontend akan berjalan di `http://localhost:3000`

## 📚 API Documentation

### Authentication Endpoints
```
POST /api/v1/auth/register    # Register user baru
POST /api/v1/auth/login       # Login user
GET  /api/v1/profile          # Get user profile
```

### Business Endpoints
```
GET    /api/v1/products       # Get semua produk
POST   /api/v1/products       # Create produk baru
PUT    /api/v1/products/:id   # Update produk
DELETE /api/v1/products/:id   # Delete produk

GET    /api/v1/sales          # Sales management
GET    /api/v1/purchases      # Purchase management
GET    /api/v1/expenses       # Expense management
GET    /api/v1/assets         # Asset management
GET    /api/v1/cash-bank      # Cash & Bank management
GET    /api/v1/inventory      # Inventory management
GET    /api/v1/reports        # Reports
```

### Health Check
```
GET /health                   # Server status
```

## 🗃️ Database Schema

### Core Tables
- **users** - User authentication dan roles
- **products** - Master data produk
- **sales** & **sale_items** - Transaksi penjualan
- **purchases** & **purchase_items** - Transaksi pembelian
- **expenses** - Pengeluaran operasional
- **assets** - Aset perusahaan
- **cash_banks** - Akun kas dan bank
- **accounts** - Chart of accounts
- **inventories** - Log pergerakan inventori

## 🔧 Development

### Backend Development
```bash
cd backend
go run cmd/main.go          # Development mode
go build -o app cmd/main.go # Build for production
```

### Frontend Development
```bash
cd frontend
npm run dev         # Development mode
npm run build       # Build for production
npm run start       # Production mode
```

## 📦 Deployment

### Backend Deployment
1. Build aplikasi: `go build -o app cmd/main.go`
2. Setup PostgreSQL di server
3. Set environment variables
4. Jalankan: `./app`

### Frontend Deployment
1. Build aplikasi: `npm run build`
2. Deploy ke platform seperti Vercel, Netlify, atau server
3. Set environment variables untuk API URL

## 🧪 Testing

### Backend Testing
```bash
cd backend
go test ./...
```

### Frontend Testing
```bash
cd frontend
npm run test
```

## 📝 TODO

### Backend
- [ ] Implement remaining controllers (Sales, Purchases, etc.)
- [ ] Add input validation
- [ ] Add logging system
- [ ] Add unit tests
- [ ] Add Swagger documentation
- [ ] Add Docker support

### Frontend
- [ ] Complete all pages implementation
- [ ] Add form validations
- [ ] Add loading states
- [ ] Add error handling
- [ ] Add unit tests
- [ ] Add PWA support

## 🤝 Contributing

1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📞 Support

Jika Anda mengalami masalah atau memiliki pertanyaan, silakan buat issue di repository ini.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Made with ❤️ for efficient business management**
