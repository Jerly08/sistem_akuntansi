# Backend - Sistem Akuntansi

Backend API untuk aplikasi sistem akuntansi menggunakan Go, Gin, GORM, dan PostgreSQL.

## 🚀 Teknologi yang Digunakan

- **Go 1.21+** - Programming language
- **Gin** - Web framework
- **GORM** - ORM untuk database
- **PostgreSQL** - Database
- **JWT** - Authentication
- **bcrypt** - Password hashing

## 📁 Struktur Folder

```
backend/
├── cmd/
│   └── main.go              # Entry point aplikasi
├── config/
│   └── config.go            # Konfigurasi aplikasi
├── controllers/
│   ├── auth_controller.go   # Controller untuk autentikasi
│   └── product_controller.go # Controller untuk produk
├── database/
│   └── database.go          # Koneksi dan migrasi database
├── middleware/
│   └── auth.go              # Middleware autentikasi JWT
├── models/
│   ├── user.go              # Model User
│   ├── product.go           # Model Product
│   ├── sale.go              # Model Sale & SaleItem
│   ├── purchase.go          # Model Purchase & PurchaseItem
│   └── expense.go           # Model lainnya (Expense, Asset, etc.)
├── routes/
│   └── routes.go            # Routing API
├── services/                # Business logic (akan digunakan nanti)
├── dto/                     # Data Transfer Objects (akan digunakan nanti)
├── utils/                   # Utility functions (akan digunakan nanti)
├── .env                     # Environment variables
├── go.mod                   # Go modules
└── README.md               # Dokumentasi ini
```

## 🛠️ Setup & Installation

### Prerequisites
- Go 1.21 atau lebih baru
- PostgreSQL 12+
- Git

### 1. Clone dan Setup
```bash
cd backend
```

### 2. Install Dependencies
```bash
go mod tidy
```

### 3. Setup Database
Buat database PostgreSQL:
```sql
CREATE DATABASE sistem_akuntansi;
```

### 4. Konfigurasi Environment
Copy dan edit file `.env`:
```bash
cp .env.example .env
```

Edit file `.env` sesuai dengan konfigurasi Anda:
```env
DATABASE_URL=postgres://username:password@localhost:5432/sistem_akuntansi?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key
SERVER_PORT=8080
ENVIRONMENT=development
```

### 5. Jalankan Aplikasi
```bash
go run cmd/main.go
```

Server akan berjalan di `http://localhost:8080`

## 📚 API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register user baru
- `POST /api/v1/auth/login` - Login user
- `GET /api/v1/profile` - Get user profile (perlu auth)

### Products
- `GET /api/v1/products` - Get semua produk
- `GET /api/v1/products/:id` - Get produk by ID
- `POST /api/v1/products` - Create produk baru (admin/inventory_manager)
- `PUT /api/v1/products/:id` - Update produk (admin/inventory_manager)
- `DELETE /api/v1/products/:id` - Delete produk (admin only)

### Coming Soon
- `GET /api/v1/sales` - Sales management
- `GET /api/v1/purchases` - Purchase management
- `GET /api/v1/expenses` - Expense management
- `GET /api/v1/assets` - Asset management
- `GET /api/v1/cash-bank` - Cash & Bank management
- `GET /api/v1/inventory` - Inventory management
- `GET /api/v1/reports` - Reports (admin/director/finance)

### Health Check
- `GET /health` - Check server status

## 🔐 Authentication

API menggunakan JWT (JSON Web Token) untuk autentikasi. Setelah login, Anda akan mendapat token yang harus disertakan di header:

```
Authorization: Bearer YOUR_JWT_TOKEN
```

## 👥 User Roles

- `admin` - Full access ke semua fitur
- `director` - Access ke reports dan read operations
- `finance` - Access ke financial data dan reports
- `inventory_manager` - Manage products dan inventory
- `employee` - Basic access

## 🗃️ Database Models

### User
```go
type User struct {
    ID        uint
    Username  string
    Email     string
    Password  string
    Role      string // admin, director, finance, employee, inventory_manager
    FirstName string
    LastName  string
    IsActive  bool
}
```

### Product
```go
type Product struct {
    ID            uint
    Code          string
    Name          string
    Description   string
    Category      string
    Unit          string
    PurchasePrice float64
    SalePrice     float64
    Stock         int
    MinStock      int
    IsActive      bool
}
```

Dan model lainnya untuk Sale, Purchase, Expense, Asset, CashBank, Account, dan Inventory.

## 🚦 Development

### Running in Development Mode
```bash
go run cmd/main.go
```

### Building for Production
```bash
go build -o app cmd/main.go
./app
```

### Database Migration
Migration akan berjalan otomatis saat aplikasi pertama kali dijalankan.

## 📝 TODO

- [ ] Implement Sales controller
- [ ] Implement Purchase controller  
- [ ] Implement Expense controller
- [ ] Implement Asset controller
- [ ] Implement Cash/Bank controller
- [ ] Implement Inventory controller
- [ ] Implement Reports controller
- [ ] Add validation
- [ ] Add logging
- [ ] Add unit tests
- [ ] Add API documentation (Swagger)
- [ ] Add Docker support

## 🤝 Contributing

1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License.
