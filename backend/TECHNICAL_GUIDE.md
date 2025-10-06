# 🔧 Technical Guide - Sistema Akuntansi

Dokumentasi teknis lengkap untuk engineers, covering deployment, architecture, code structure, dan troubleshooting.

## 📋 Table of Contents
- [Architecture Overview](#-architecture-overview)
- [Development Setup](#-development-setup)
- [Code Structure](#-code-structure)
- [Database Schema](#-database-schema)
- [Deployment](#-deployment)
- [Security](#-security)
- [Performance](#-performance)
- [Troubleshooting](#-troubleshooting)
- [API Reference](#-api-reference)

## 🏗️ Architecture Overview

### System Architecture
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend API   │    │   Database      │
│   React/Next.js │◄──►│   Go/Gin        │◄──►│   PostgreSQL    │
│   Port: 3000    │    │   Port: 8080    │    │   Port: 5432    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                       ┌─────────────────┐
                       │   WebSocket     │
                       │   Real-time     │
                       │   Updates       │
                       └─────────────────┘
```

### Technology Stack
- **Backend**: Go 1.19+ with Gin framework
- **Database**: PostgreSQL 13+ with GORM ORM
- **Authentication**: JWT with refresh tokens
- **Documentation**: Swagger/OpenAPI 3.0
- **Real-time**: WebSocket for live updates
- **Caching**: In-memory caching for reports
- **Security**: Role-based access control (RBAC)

### Key Components
- **SSOT Journal System**: Single source of truth untuk semua transaksi
- **Balance Protection System**: Real-time balance monitoring dan sync
- **Enhanced Security Middleware**: Multi-layer security protection
- **Unified Report Service**: Centralized reporting dengan caching
- **Approval Workflow Engine**: Configurable approval processes

## 🚀 Development Setup

### Prerequisites
```bash
# Required software
Go 1.19+
PostgreSQL 13+
Git
Make (optional)
```

### 1. Environment Setup

**Clone Repository:**
```bash
git clone <repository_url>
cd accounting_proj/backend
```

**Database Setup:**
```bash
# Create database
createdb sistem_akuntans_test

# Or via psql
psql -U postgres
CREATE DATABASE sistem_akuntans_test;
```

**Environment Configuration:**
```bash
# Copy environment template
cp .env.example .env

# Edit .env file
DATABASE_URL=postgres://postgres:password@localhost/sistem_akuntans_test?sslmode=disable
JWT_ACCESS_SECRET=your-access-secret-key
JWT_REFRESH_SECRET=your-refresh-secret-key
SERVER_PORT=8080
ENVIRONMENT=development
```

### 2. Dependencies & Build

**Install Dependencies:**
```bash
go mod download
go mod verify
```

**Build Application:**
```bash
# Development build
go build -o bin/server cmd/main.go

# Production build with optimizations
go build -ldflags="-w -s" -o bin/server cmd/main.go
```

### 3. Critical Setup (MANDATORY)

**⚠️ Balance Protection Setup:**
```bash
# Windows
setup_balance_protection.bat

# Linux/Mac
chmod +x setup_balance_protection.sh
./setup_balance_protection.sh

# Manual fallback
go run cmd/scripts/setup_balance_sync_auto.go
```

**Migration & Database Setup:**
```bash
# Run migration fixes
go run cmd/fix_migrations.go
go run cmd/fix_remaining_migrations.go

# Verify setup
go run cmd/final_verification.go
```

### 4. Run Application

**Development Mode:**
```bash
# Direct run
go run cmd/main.go

# With hot reload (install air)
go install github.com/cosmtrek/air@latest
air
```

**Production Mode:**
```bash
# Build and run
make build
./bin/server

# With environment
ENVIRONMENT=production ./bin/server
```

### 5. Verification

**Health Checks:**
```bash
# Basic health
curl http://localhost:8080/api/v1/health

# Database health
curl http://localhost:8080/api/v1/monitoring/status

# Balance system health
curl http://localhost:8080/api/v1/admin/balance-health/check
```

## 📁 Code Structure

### Directory Layout
```
backend/
├── cmd/                    # Entry points
│   ├── main.go            # Main application
│   ├── fix_migrations.go  # Migration fixes
│   └── scripts/           # Utility scripts
├── config/                # Configuration
│   ├── config.go          # App configuration
│   └── swagger.go         # Swagger setup
├── controllers/           # HTTP handlers
│   ├── auth_controller.go
│   ├── sales_controller.go
│   └── enhanced_report_controller.go
├── middleware/            # HTTP middleware
│   ├── auth_middleware.go
│   ├── permission_middleware.go
│   └── security_middleware.go
├── models/                # Data models
│   ├── user.go
│   ├── account.go
│   └── journal_entry.go
├── repositories/          # Data access layer
│   ├── account_repository.go
│   └── journal_repository.go
├── services/              # Business logic
│   ├── auth_service.go
│   ├── sales_service.go
│   └── report_service.go
├── handlers/              # Request handlers
├── database/              # Database setup
│   ├── connection.go
│   ├── migrations/
│   └── seeds/
├── routes/                # Route definitions
│   └── routes.go
├── utils/                 # Utilities
├── docs/                  # Generated swagger docs
└── templates/             # PDF templates
```

### Key Files Explained

**cmd/main.go**: Application entry point
- Configuration loading
- Database connection
- Middleware setup
- Route registration
- Server startup

**config/config.go**: Configuration management
- Environment variables
- JWT settings
- Database configuration
- Security settings

**routes/routes.go**: Route definitions
- API endpoint mapping
- Middleware application
- Permission setup
- Swagger documentation

### Design Patterns

**Repository Pattern:**
```go
type AccountRepository interface {
    Create(account *models.Account) error
    FindByID(id uint) (*models.Account, error)
    Update(account *models.Account) error
    Delete(id uint) error
}
```

**Service Layer Pattern:**
```go
type SalesService struct {
    repo *repositories.SalesRepository
    journalService *services.UnifiedJournalService
}

func (s *SalesService) CreateSale(sale *models.Sale) error {
    // Business logic here
    return s.repo.Create(sale)
}
```

**Middleware Chain:**
```go
protected := v1.Group("")
protected.Use(jwtManager.AuthRequired())
protected.Use(permMiddleware.CanView("sales"))
protected.GET("/sales", salesController.GetSales)
```

## 🗃️ Database Schema

### Core Tables

**accounts**: Chart of accounts
```sql
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    account_type VARCHAR(50) NOT NULL,
    parent_id INTEGER REFERENCES accounts(id),
    is_header BOOLEAN DEFAULT FALSE,
    balance DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**journal_entries**: SSOT Journal system
```sql
CREATE TABLE journal_entries (
    id SERIAL PRIMARY KEY,
    journal_code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    total_amount DECIMAL(15,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'DRAFT',
    posted_at TIMESTAMP,
    created_by INTEGER REFERENCES users(id)
);
```

**journal_items**: Journal entry details
```sql
CREATE TABLE journal_items (
    id SERIAL PRIMARY KEY,
    journal_entry_id INTEGER REFERENCES journal_entries(id),
    account_id INTEGER REFERENCES accounts(id),
    debit_amount DECIMAL(15,2) DEFAULT 0,
    credit_amount DECIMAL(15,2) DEFAULT 0,
    description TEXT
);
```

### Key Indexes
```sql
-- Performance indexes
CREATE INDEX idx_accounts_code ON accounts(code);
CREATE INDEX idx_accounts_type ON accounts(account_type);
CREATE INDEX idx_journal_entries_status ON journal_entries(status);
CREATE INDEX idx_journal_items_account ON journal_items(account_id);
CREATE INDEX idx_journal_items_journal ON journal_items(journal_entry_id);

-- Balance monitoring
CREATE INDEX idx_accounts_balance ON accounts(balance);
CREATE INDEX idx_journal_entries_posted_at ON journal_entries(posted_at);
```

### Materialized Views

**account_balances**: Real-time balance calculation
```sql
CREATE MATERIALIZED VIEW account_balances AS
SELECT 
    a.id,
    a.code,
    a.name,
    a.account_type,
    COALESCE(SUM(ji.debit_amount - ji.credit_amount), 0) as calculated_balance,
    a.balance as stored_balance,
    a.updated_at
FROM accounts a
LEFT JOIN journal_items ji ON a.id = ji.account_id
LEFT JOIN journal_entries je ON ji.journal_entry_id = je.id
WHERE je.status = 'POSTED'
GROUP BY a.id, a.code, a.name, a.account_type, a.balance, a.updated_at;

-- Refresh strategy
CREATE UNIQUE INDEX ON account_balances (id);
```

### Database Functions

**Balance Sync Function:**
```sql
CREATE OR REPLACE FUNCTION sync_account_balance_from_ssot(account_id INTEGER)
RETURNS VOID AS $$
DECLARE
    calculated_balance DECIMAL(15,2);
BEGIN
    -- Calculate balance from journal items
    SELECT COALESCE(SUM(debit_amount - credit_amount), 0)
    INTO calculated_balance
    FROM journal_items ji
    JOIN journal_entries je ON ji.journal_entry_id = je.id
    WHERE ji.account_id = sync_account_balance_from_ssot.account_id
    AND je.status = 'POSTED';
    
    -- Update account balance
    UPDATE accounts 
    SET balance = calculated_balance, 
        updated_at = NOW()
    WHERE id = sync_account_balance_from_ssot.account_id;
END;
$$ LANGUAGE plpgsql;
```

### Triggers

**Auto Balance Sync:**
```sql
CREATE OR REPLACE FUNCTION trigger_balance_sync()
RETURNS TRIGGER AS $$
BEGIN
    -- Refresh materialized view
    REFRESH MATERIALIZED VIEW CONCURRENTLY account_balances;
    
    -- Sync specific account balance
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        PERFORM sync_account_balance_from_ssot(NEW.account_id);
    END IF;
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER balance_sync_trigger
    AFTER INSERT OR UPDATE OR DELETE ON journal_items
    FOR EACH ROW EXECUTE FUNCTION trigger_balance_sync();
```

## 🚀 Deployment

### Production Deployment

**1. Build Preparation:**
```bash
# Set production environment
export ENVIRONMENT=production
export GO_ENV=production

# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o accounting-server cmd/main.go

# Verify binary
./accounting-server --version
```

**2. Docker Deployment:**
```dockerfile
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Jakarta

# Create app directory
WORKDIR /app

# Copy binary
COPY accounting-server .
COPY templates/ templates/
COPY .env.production .env

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Run application
CMD ["./accounting-server"]
```

**3. Database Migration in Production:**
```bash
# Pre-deployment database backup
pg_dump -h $DB_HOST -U $DB_USER $DB_NAME > backup_$(date +%Y%m%d_%H%M%S).sql

# Run migrations
./accounting-server migrate

# Verify migration
./accounting-server migrate --status
```

**4. Environment Variables (Production):**
```bash
# Database
DATABASE_URL=postgres://user:pass@prod-db:5432/accounting_prod?sslmode=require

# JWT
JWT_ACCESS_SECRET=complex-secret-key-64-chars
JWT_REFRESH_SECRET=another-complex-secret-key-64-chars

# Server
SERVER_PORT=8080
ENVIRONMENT=production

# Security
ENABLE_HTTPS=true
ENABLE_RATE_LIMIT=true
ENABLE_SECURITY_HEADERS=true

# CORS
ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
```

### Kubernetes Deployment

**deployment.yaml:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: accounting-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: accounting-backend
  template:
    metadata:
      labels:
        app: accounting-backend
    spec:
      containers:
      - name: accounting-backend
        image: your-registry/accounting-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## 🔒 Security

### Authentication & Authorization

**JWT Implementation:**
```go
type JWTManager struct {
    accessSecret  string
    refreshSecret string
    accessExpiry  time.Duration
    refreshExpiry time.Duration
}

func (j *JWTManager) GenerateTokenPair(userID uint) (*TokenPair, error) {
    // Access token (short-lived)
    accessClaims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(j.accessExpiry).Unix(),
        "iat":     time.Now().Unix(),
        "type":    "access",
    }
    
    // Refresh token (long-lived)
    refreshClaims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(j.refreshExpiry).Unix(),
        "iat":     time.Now().Unix(),
        "type":    "refresh",
    }
    
    // Generate tokens...
}
```

**Permission Middleware:**
```go
func (p *PermissionMiddleware) CanView(module string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := getUserID(c)
        if !p.hasPermission(userID, module, "view") {
            c.JSON(403, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### Security Middleware Stack

**Enhanced Security Middleware:**
```go
type EnhancedSecurityMiddleware struct {
    db           *gorm.DB
    rateLimiter  *RateLimiter
    ipWhitelist  []string
}

func (m *EnhancedSecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Security headers
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        
        // CSP header
        csp := "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'"
        c.Header("Content-Security-Policy", csp)
        
        c.Next()
    }
}
```

**Rate Limiting:**
```go
type RateLimiter struct {
    requests map[string]*limiter.Limiter
    mutex    sync.RWMutex
}

func (r *RateLimiter) Limit(key string, rate int) gin.HandlerFunc {
    return func(c *gin.Context) {
        limiter := r.getLimiter(c.ClientIP())
        if !limiter.Allow() {
            c.JSON(429, gin.H{"error": "Rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### Data Protection

**Sensitive Data Handling:**
```go
// Password hashing
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// Audit logging
type AuditLog struct {
    UserID    uint      `json:"user_id"`
    Action    string    `json:"action"`
    Resource  string    `json:"resource"`
    Details   string    `json:"details"`
    IPAddress string    `json:"ip_address"`
    Timestamp time.Time `json:"timestamp"`
}
```

## ⚡ Performance

### Database Optimization

**Connection Pooling:**
```go
func ConnectDB() *gorm.DB {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    
    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    return db
}
```

**Query Optimization:**
```go
// Preloading relations
func (r *SalesRepository) FindWithDetails(id uint) (*models.Sale, error) {
    var sale models.Sale
    return &sale, r.db.
        Preload("Customer").
        Preload("Items").
        Preload("Items.Product").
        First(&sale, id).Error
}

// Selective fields
func (r *AccountRepository) GetAccountSummary() ([]AccountSummary, error) {
    var summaries []AccountSummary
    return summaries, r.db.
        Model(&models.Account{}).
        Select("id, code, name, account_type, balance").
        Where("is_active = ?", true).
        Find(&summaries).Error
}
```

### Caching Strategy

**Report Caching:**
```go
type ReportCacheService struct {
    cache map[string]*CacheEntry
    mutex sync.RWMutex
    ttl   time.Duration
}

func (c *ReportCacheService) Get(key string) (interface{}, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    entry, exists := c.cache[key]
    if !exists || time.Now().After(entry.ExpireAt) {
        return nil, false
    }
    
    return entry.Data, true
}
```

### Ultra-Fast Endpoints

**Optimized Payment Processing:**
```go
// Ultra-fast payment endpoint
func (s *UltraFastPaymentService) RecordPaymentUltraFast(req *UltraFastPaymentRequest) error {
    // Minimal validation
    if req.Amount <= 0 {
        return errors.New("invalid amount")
    }
    
    // Direct database insert with minimal overhead
    return s.db.Create(&models.Payment{
        Amount:      req.Amount,
        PaymentDate: time.Now(),
        Status:      "PENDING_JOURNAL",
    }).Error
}

// Async journal creation
func (s *UltraFastPaymentService) CreateJournalEntryAsync(req *UltraFastPaymentRequest) {
    go func() {
        // Create journal entry in background
        s.createJournalEntry(req)
    }()
}
```

## 🐛 Troubleshooting

### Common Issues

**1. Database Connection Issues**

*Symptoms:*
```
FATAL: database connection failed
Error: dial tcp 127.0.0.1:5432: connect: connection refused
```

*Solutions:*
```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Check connection string
echo $DATABASE_URL

# Test connection
psql $DATABASE_URL -c "SELECT 1;"

# Check firewall
sudo ufw status
```

**2. Balance Mismatch Errors**

*Symptoms:*
```
ERROR: Balance mismatch detected for account 1001
Expected: 1000.00, Actual: 995.00
```

*Solutions:*
```bash
# Run balance health check
curl http://localhost:8080/api/v1/admin/balance-health/check

# Auto-heal balance issues
curl -X POST http://localhost:8080/api/v1/admin/balance-health/auto-heal

# Manual balance sync
go run cmd/scripts/fix_balance_sync.go
```

**3. JWT Token Issues**

*Symptoms:*
```
HTTP 401: Invalid token
Token has expired
```

*Solutions:*
```bash
# Check JWT configuration
grep JWT .env

# Verify token expiry settings
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/auth/validate-token

# Use refresh token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -d '{"refresh_token": "your-refresh-token"}'
```

**4. Migration Failures**

*Symptoms:*
```
ERROR: Migration failed: table "accounts" already exists
SQLSTATE 42P07
```

*Solutions:*
```bash
# Check migration status
go run cmd/check_migration_table.go

# Reset migrations (DANGER: data loss)
go run cmd/reset_migrations.go

# Fix specific migrations
go run cmd/fix_remaining_migrations.go
```

**5. Performance Issues**

*Symptoms:*
```
HTTP 504: Gateway timeout
Slow query performance
High memory usage
```

*Solutions:*
```bash
# Check performance metrics
curl http://localhost:8080/api/v1/monitoring/performance/metrics

# Database analysis
EXPLAIN ANALYZE SELECT * FROM accounts WHERE balance > 1000;

# Clear performance metrics
curl -X POST http://localhost:8080/api/v1/monitoring/performance/metrics/clear

# Check connection pool
curl http://localhost:8080/api/v1/monitoring/status
```

### Debug Tools

**Debug Endpoints (Development Only):**
```bash
# System information
curl http://localhost:8080/api/v1/debug/auth/context \
  -H "Authorization: Bearer $TOKEN"

# Test permissions
curl http://localhost:8080/api/v1/debug/auth/test-cashbank-permission \
  -H "Authorization: Bearer $TOKEN"

# JWT context
curl http://localhost:8080/api/v1/debug/auth/role \
  -H "Authorization: Bearer $TOKEN"
```

**Logging Configuration:**
```go
// Enable detailed logging in development
if gin.Mode() == gin.DebugMode {
    db.Logger = logger.Default.LogMode(logger.Info)
}

// Custom log format
log.SetFlags(log.LstdFlags | log.Lshortfile)
log.Printf("DEBUG: User %d accessed %s", userID, endpoint)
```

### Monitoring & Alerts

**Health Check Endpoints:**
```bash
# Basic health
curl http://localhost:8080/api/v1/health

# Detailed system status
curl http://localhost:8080/api/v1/monitoring/status

# Balance system health
curl http://localhost:8080/api/v1/admin/balance-health/detailed-report

# Performance bottlenecks
curl http://localhost:8080/api/v1/monitoring/performance/bottlenecks
```

**Log Analysis:**
```bash
# Error pattern analysis
grep -i "error\|fail\|panic" logs/app.log | tail -50

# Performance issues
grep -i "slow query\|timeout" logs/app.log

# Security incidents
grep -i "401\|403\|security" logs/app.log
```

## 📚 API Reference

### Base Configuration
```
Base URL: http://localhost:8080/api/v1
Content-Type: application/json
Authorization: Bearer <jwt_token>
```

### Core Endpoints

**Authentication:**
```bash
POST /api/v1/auth/login
POST /api/v1/auth/refresh
GET  /api/v1/auth/validate-token
GET  /api/v1/profile
```

**Master Data:**
```bash
GET    /api/v1/accounts
POST   /api/v1/accounts
GET    /api/v1/accounts/:code
PUT    /api/v1/accounts/:code
DELETE /api/v1/accounts/:code

GET    /api/v1/contacts
POST   /api/v1/contacts
GET    /api/v1/contacts/:id

GET    /api/v1/products
POST   /api/v1/products
PUT    /api/v1/products/:id
```

**Transactions:**
```bash
GET  /api/v1/sales
POST /api/v1/sales
GET  /api/v1/sales/:id
PUT  /api/v1/sales/:id

GET  /api/v1/purchases
POST /api/v1/purchases
GET  /api/v1/purchases/:id

GET  /api/v1/payments
POST /api/v1/payments
```

**Reports:**
```bash
GET /api/v1/reports/balance-sheet?start_date=2024-01-01&end_date=2024-12-31
GET /api/v1/reports/ssot-profit-loss?start_date=2024-01-01&end_date=2024-12-31
GET /api/v1/ssot-reports/trial-balance?as_of_date=2024-12-31
GET /api/v1/reports/ssot/cash-flow?start_date=2024-01-01&end_date=2024-12-31
```

**Monitoring:**
```bash
GET /api/v1/health
GET /api/v1/monitoring/status
GET /api/v1/monitoring/performance/metrics
GET /api/v1/admin/balance-health/check
```

### Error Response Format
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": {
      "field": "amount",
      "issue": "must be greater than zero"
    }
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Response Format
```json
{
  "success": true,
  "data": {
    // Response data
  },
  "metadata": {
    "total": 100,
    "page": 1,
    "limit": 20
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

## 📞 Support & Maintenance

### Regular Maintenance Tasks

**Daily:**
```bash
# Check system health
curl http://localhost:8080/api/v1/health

# Monitor balance health
curl http://localhost:8080/api/v1/admin/balance-health/check
```

**Weekly:**
```bash
# Database maintenance
go run cmd/maintenance/db_cleanup.go

# Performance analysis
curl http://localhost:8080/api/v1/monitoring/performance/report
```

**Monthly:**
```bash
# Security audit
go run cmd/security/audit_report.go

# Backup verification
pg_dump $DATABASE_URL > backup_verification.sql
```

### Getting Help

**Internal Support:**
- Check logs: `tail -f logs/app.log`
- Run diagnostics: `go run cmd/diagnostics.go`
- Check documentation: `/docs` folder

**External Resources:**
- Go Documentation: https://golang.org/doc/
- Gin Framework: https://gin-gonic.com/docs/
- GORM Documentation: https://gorm.io/docs/
- PostgreSQL Manual: https://www.postgresql.org/docs/

---

**🎯 Next Steps:**
- Review [API Documentation](API_DOCUMENTATION.md) for detailed endpoint specs
- Check [Features Guide](FEATURES.md) for functional capabilities  
- Refer to [User Manual](README_COMPREHENSIVE.md) for business process guides