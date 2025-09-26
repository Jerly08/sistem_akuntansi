# 🚀 Quick Fix Script untuk Login Error
Write-Host "🔧 Quick Fix untuk Login Error" -ForegroundColor Green
Write-Host "=================================" -ForegroundColor Green

# 1. Check if backend is running
Write-Host "`n1. Checking backend..." -ForegroundColor Yellow
$backendRunning = $false
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -TimeoutSec 3
    Write-Host "✅ Backend running on port 8080" -ForegroundColor Green
    $backendRunning = $true
} catch {
    Write-Host "❌ Backend not running on port 8080" -ForegroundColor Red
    Write-Host "   Start backend dengan: cd backend; go run ./cmd"
}

# 2. Check if frontend is running  
Write-Host "`n2. Checking frontend..." -ForegroundColor Yellow
$frontendRunning = $false
try {
    $response = Invoke-WebRequest -Uri "http://localhost:3000" -TimeoutSec 3 -UseBasicParsing
    Write-Host "✅ Frontend running on port 3000" -ForegroundColor Green
    $frontendRunning = $true
} catch {
    Write-Host "❌ Frontend not running on port 3000" -ForegroundColor Red  
    Write-Host "   Start frontend dengan: cd frontend; npm run dev"
}

# 3. Check frontend .env.local
Write-Host "`n3. Checking frontend configuration..." -ForegroundColor Yellow
$frontendEnvPath = "frontend\.env.local"
if (Test-Path $frontendEnvPath) {
    $envContent = Get-Content $frontendEnvPath
    $apiUrlFound = $envContent | Where-Object { $_ -match "NEXT_PUBLIC_API_URL" }
    
    if ($apiUrlFound -and $apiUrlFound -match "localhost:8080") {
        Write-Host "✅ Frontend .env.local configured correctly" -ForegroundColor Green
    } else {
        Write-Host "❌ Frontend .env.local needs fixing" -ForegroundColor Red
        Write-Host "   Create/update frontend/.env.local with:"
        Write-Host "   NEXT_PUBLIC_API_URL=http://localhost:8080" -ForegroundColor Cyan
    }
} else {
    Write-Host "❌ Frontend .env.local missing" -ForegroundColor Red
    Write-Host "   Create frontend/.env.local with:"
    Write-Host "   NEXT_PUBLIC_API_URL=http://localhost:8080" -ForegroundColor Cyan
}

# 4. Test login API if backend is running
if ($backendRunning) {
    Write-Host "`n4. Testing login API..." -ForegroundColor Yellow
    try {
        $body = @{
            email = "admin@company.com"
            password = "password123"
        } | ConvertTo-Json
        
        $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
            -Method Post `
            -ContentType "application/json" `
            -Body $body
            
        if ($response.success -and $response.access_token) {
            Write-Host "✅ Login API works perfectly!" -ForegroundColor Green
            Write-Host "   Response includes: success, access_token, token, user" -ForegroundColor Green
        } else {
            Write-Host "⚠️  Login API responded but format might be wrong" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "❌ Login API test failed" -ForegroundColor Red
        Write-Host "   Error: $($_.Exception.Message)"
    }
}

# 5. Summary and next steps
Write-Host "`n" + "="*50 -ForegroundColor Green
Write-Host "📋 SUMMARY" -ForegroundColor Green
Write-Host "="*50 -ForegroundColor Green

if ($backendRunning -and $frontendRunning) {
    Write-Host "✅ Both services running" -ForegroundColor Green
    Write-Host "`n🎯 NEXT STEPS:" -ForegroundColor Cyan
    Write-Host "1. Open browser: http://localhost:3000/login"
    Write-Host "2. Clear browser cache (F12 → Application → Clear storage)"  
    Write-Host "3. Try login with:"
    Write-Host "   Email: admin@company.com"
    Write-Host "   Password: password123"
    Write-Host "`n4. If still error, check browser console (F12) for details"
    
} else {
    Write-Host "❌ Services not running properly" -ForegroundColor Red
    Write-Host "`n🚀 START SERVICES:" -ForegroundColor Cyan
    if (!$backendRunning) {
        Write-Host "Terminal 1: cd backend && go run ./cmd"
    }
    if (!$frontendRunning) {
        Write-Host "Terminal 2: cd frontend && npm run dev"  
    }
}

Write-Host "`nFor detailed guide, see: LOGIN_ERROR_FIX_GUIDE.md" -ForegroundColor Yellow
