# 🚀 Swagger API Documentation - Quick Start Guide

## ⚡ Quick Setup

### 1. **Start the Application**
```bash
# Build & Run
go build -o main.exe cmd/main.go
./main.exe
```

### 2. **Access Swagger UI**
Open browser and navigate to:
```
http://localhost:8080/swagger/index.html
```

### 3. **Test API (5 minutes)**

#### Step 1: Login
1. Find `POST /api/v1/auth/login`
2. Click "Try it out"
3. Enter credentials:
```json
{
  "username": "admin", 
  "password": "your_password"
}
```
4. Copy the `access_token` from response

#### Step 2: Authorize
1. Click "Authorize" button (🔓)
2. Enter: `Bearer YOUR_ACCESS_TOKEN`
3. Click "Authorize"

#### Step 3: Test Protected Endpoints
Now you can test any protected endpoint like:
- `GET /api/v1/profile` - Your profile
- `GET /api/v1/dashboard/summary` - Dashboard data
- `GET /api/v1/dashboard/quick-stats` - Quick stats

## 📋 Available Documentation

### **Authentication & Security**
- User login/logout
- Token management
- Profile management

### **Dashboard & Analytics** 
- Dashboard summary
- Analytics data
- Quick statistics

### **Financial Management**
- Cash & Bank accounts
- Payment processing
- Balance monitoring

### **System Administration**
- Balance synchronization
- System health checks
- Error monitoring

## 🎯 Key Features

✅ **Interactive API Testing**
- Test API directly from browser
- Real-time request/response
- Authentication integrated

✅ **Comprehensive Documentation**  
- All endpoints documented
- Request/response examples
- Error codes & descriptions

✅ **Security Ready**
- JWT authentication
- Role-based access control
- Environment-aware activation

## 🔧 Configuration

### Development (Auto-enabled)
```env
ENV=development
```

### Production  
```env
ENV=production
ENABLE_SWAGGER=true
```

## 📚 Full Documentation

For complete implementation details, see:
`docs/SWAGGER_IMPLEMENTATION_GUIDE.md`

---

**🎉 Ready to explore your API!** 
Navigate to `http://localhost:8080/swagger/index.html` and start testing!