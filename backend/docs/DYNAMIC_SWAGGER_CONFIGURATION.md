# 🔧 Dynamic Swagger Configuration - No More Hardcode!

## 🎯 Problem Solved

**Before:** Swagger configuration was hardcoded dengan `localhost:8080` dan CORS origins `localhost:3000`

**After:** Sepenuhnya dinamis berdasarkan environment variables dan environment mode!

## ✅ Yang Diperbaiki

### 1. **Hardcoded Host Address** ❌ → **Dynamic Host** ✅
```go
// SEBELUM (hardcode)
// @host localhost:8080

// SESUDAH (dynamic runtime)  
// Host ditentukan berdasarkan:
// - SWAGGER_HOST environment variable
// - SERVER_PORT untuk development  
// - DOMAIN/APP_URL untuk production
```

### 2. **Hardcoded CORS Origins** ❌ → **Dynamic CORS** ✅
```go
// SEBELUM (hardcode)
AllowOrigins: []string{"http://localhost:3000", "http://localhost:3001"}

// SESUDAH (dynamic)
allowedOrigins := config.GetAllowedOrigins(cfg) // Dynamic based on env
```

### 3. **Static Scheme** ❌ → **Environment-Aware Scheme** ✅
```go
// Development: http (default)
// Production: https (when ENABLE_HTTPS=true)
// Manual override: SWAGGER_SCHEME environment variable
```

## 🛠️ Implementasi Detail

### 1. **Dynamic Swagger Configuration**

#### File: `config/swagger.go`
```go
// GetSwaggerConfig() menentukan konfigurasi berdasarkan environment:
// - Development: localhost:{SERVER_PORT}
// - Production: {DOMAIN} atau {SWAGGER_HOST}
// - Scheme: http/https berdasarkan ENABLE_HTTPS
```

#### File: `config/swagger_updater.go`  
```go
// UpdateSwaggerDocs() mengupdate swagger.json secara runtime
// - Membaca file swagger.json yang di-generate
// - Update host, scheme, basePath secara dynamic
// - Support untuk production deployment
```

### 2. **Environment Variables untuk Konfigurasi**

#### **Development** (.env)
```env
# Auto-detection berdasarkan SERVER_PORT
SWAGGER_HOST=                    # Empty = auto-detect
SWAGGER_SCHEME=http             # Default untuk dev
SERVER_PORT=8080                # Akan jadi localhost:8080
ALLOWED_ORIGINS=                # Empty = development defaults
```

#### **Production** (.env.production)
```env
# Explicit configuration untuk production
SWAGGER_HOST=api.yourdomain.com
SWAGGER_SCHEME=https  
ENABLE_HTTPS=true
ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
DOMAIN=yourdomain.com
APP_URL=https://yourdomain.com
```

### 3. **Dynamic CORS Origins**

#### **Development Mode (Auto)**
```go
// Otomatis includes:
// - http://localhost:3000, :3001, :3002
// - http://127.0.0.1:3000, :3001, :3002
```

#### **Production Mode (Configuration-Based)**
```go
// Berdasarkan environment variables:
// 1. ALLOWED_ORIGINS (highest priority)
// 2. FRONTEND_URL  
// 3. DOMAIN + ENABLE_HTTPS
// 4. APP_URL
// 5. Warning + fallback jika tidak ada
```

## 🚀 Cara Kerja

### 1. **Application Startup**
```go
func main() {
    cfg := config.LoadConfig()                    // Load env config
    
    // Dynamic CORS
    allowedOrigins := config.GetAllowedOrigins(cfg) // ✅ No hardcode
    
    // Dynamic Swagger Update (runtime)
    if cfg.Environment == "development" || os.Getenv("ENABLE_SWAGGER") == "true" {
        config.UpdateSwaggerDocs()                // ✅ Update docs dynamic
        config.PrintSwaggerInfo()                 // ✅ Print helpful info
    }
}
```

### 2. **Runtime Swagger Update**
```go
// Proses update swagger.json:
// 1. Read existing swagger.json
// 2. Parse JSON structure  
// 3. Update: host, schemes, basePath, title, description
// 4. Write back to file
// 5. Also update docs.go if needed
```

### 3. **Smart Environment Detection**
```go
// Development Detection:
host = fmt.Sprintf("localhost:%s", serverPort)

// Production Detection:  
if domain := os.Getenv("DOMAIN"); domain != "" {
    host = domain
} else if appURL := os.Getenv("APP_URL"); appURL != "" {
    host = extractHostFromURL(appURL)
} else {
    host = "api.yourdomain.com" // fallback
}
```

## 📋 Environment Variables Reference

### **Required for Production**
```env
ENVIRONMENT=production
SWAGGER_HOST=api.yourdomain.com     # Your API domain
SWAGGER_SCHEME=https                # Use HTTPS
ALLOWED_ORIGINS=https://app.com     # Frontend URL(s)
ENABLE_HTTPS=true                   # Enable HTTPS mode
```

### **Optional (Smart Defaults)**
```env
SWAGGER_BASE_PATH=/api/v1           # API base path
SWAGGER_TITLE=Your API Title        # Custom title
SWAGGER_DESCRIPTION=Your API Desc   # Custom description
DOMAIN=yourdomain.com               # Main domain
APP_URL=https://yourdomain.com      # Full app URL
FRONTEND_URL=https://app.com        # Frontend specific URL
```

### **Development (Auto-Detected)**
```env
SERVER_PORT=8080                    # Will become localhost:8080
# All other values auto-detected or use sensible defaults
```

## 🌟 Features

### ✅ **Smart Environment Detection**
- Deteksi otomatis development vs production
- Auto-configure berdasarkan environment variables
- Fallback values yang masuk akal

### ✅ **Multiple Configuration Sources**
1. Explicit environment variables (highest priority)
2. Derived from DOMAIN/APP_URL  
3. Auto-detection dari SERVER_PORT
4. Sensible defaults

### ✅ **Runtime Updates**
- Update swagger.json saat aplikasi start
- Tidak perlu rebuild untuk ganti host
- Support untuk containerized deployment

### ✅ **Development Friendly**
- Zero configuration untuk development
- Auto-deteksi port dari SERVER_PORT
- Multiple localhost variations untuk CORS

### ✅ **Production Ready**
- HTTPS support otomatis
- Security warnings jika misconfigured  
- Multiple frontend URL support

## 📞 Testing & Usage

### **Development Testing**
```bash
# Default configuration (auto-detect)
./main.exe

# Custom port
SERVER_PORT=3000 ./main.exe

# Custom host override
SWAGGER_HOST=192.168.1.100:8080 ./main.exe
```

### **Production Testing** 
```bash
# Full production config
ENVIRONMENT=production \
SWAGGER_HOST=api.company.com \
SWAGGER_SCHEME=https \
ALLOWED_ORIGINS=https://app.company.com \
ENABLE_HTTPS=true \
./main.exe
```

### **Docker/Container**
```dockerfile
ENV ENVIRONMENT=production
ENV SWAGGER_HOST=api.company.com
ENV SWAGGER_SCHEME=https  
ENV ALLOWED_ORIGINS=https://app.company.com
ENV ENABLE_HTTPS=true
```

## 🎯 Output Examples

### **Development Output**
```
🚀 Swagger Configuration:
   Environment: development
   Swagger URL: http://localhost:8080/swagger/index.html
   API Base URL: http://localhost:8080/api/v1
   Host: localhost:8080
   Scheme: http
   CORS Origins: [http://localhost:3000 http://localhost:3001 ...]

💡 Development Mode - Dynamic Configuration Active
   To override: Set SWAGGER_HOST, SWAGGER_SCHEME, ALLOWED_ORIGINS
```

### **Production Output**
```
🚀 Swagger Configuration:
   Environment: production  
   Swagger URL: https://api.company.com/swagger/index.html
   API Base URL: https://api.company.com/api/v1
   Host: api.company.com
   Scheme: https
   CORS Origins: [https://app.company.com https://admin.company.com]

💡 Production Environment Variables:
   SWAGGER_HOST: Set your production domain (e.g., api.yourdomain.com)
   SWAGGER_SCHEME: https (recommended for production)
   ALLOWED_ORIGINS: Your frontend URL(s)
   DOMAIN or APP_URL: Your main domain
   ENABLE_HTTPS: true (recommended for production)
```

## 🔧 Migration Guide

### **From Hardcoded to Dynamic**

1. **Update .env file:**
```env
# Add these lines (empty values = auto-detect)
SWAGGER_HOST=
SWAGGER_SCHEME=http
ALLOWED_ORIGINS=
```

2. **For Production deployment:**
```env
# Create .env.production with:
ENVIRONMENT=production
SWAGGER_HOST=your-api-domain.com
SWAGGER_SCHEME=https
ALLOWED_ORIGINS=https://your-frontend.com
ENABLE_HTTPS=true
```

3. **Rebuild application:**
```bash
go build -o main.exe cmd/main.go
./main.exe
```

4. **Verify configuration:**
- Check startup logs untuk configuration info
- Test Swagger UI pada URL yang ditampilkan
- Verify CORS dengan frontend

## 🎉 Benefits

### **For Development**
- ✅ Zero configuration needed
- ✅ Works dengan any port  
- ✅ Support multiple frontend ports
- ✅ Auto-deteksi semua environment

### **For Production** 
- ✅ Proper HTTPS support
- ✅ Custom domain support
- ✅ Multiple frontend URLs
- ✅ Security best practices

### **For DevOps**
- ✅ Container-friendly
- ✅ Environment-specific configuration
- ✅ No hardcoded values
- ✅ Easy deployment automation

---

## 🎯 Kesimpulan

✅ **Tidak ada lagi hardcode!**
- Host address: Dynamic berdasarkan environment
- CORS origins: Dynamic berdasarkan configuration  
- Scheme: Otomatis http/https berdasarkan environment
- Titles & descriptions: Configurable via environment

✅ **Production-ready deployment!**
- Support untuk custom domains
- HTTPS automatic detection
- Multiple frontend URL support
- Security warnings & best practices

✅ **Developer-friendly experience!**
- Zero config untuk development
- Helpful startup information
- Easy testing dengan different ports
- Smart fallbacks untuk semua scenarios

**No more localhost hardcoding!** 🎉