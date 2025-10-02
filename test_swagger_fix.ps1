# 🔧 Quick Fix Test Script for Swagger AUTH_HEADER_MISSING
# This script will restart the backend and test the fixes

Write-Host "🚀 Starting Swagger Authentication Fix Test..." -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Cyan

# Function to test if backend is running
function Test-Backend {
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -Method Get -TimeoutSec 5
        return $true
    } catch {
        return $false
    }
}

# Stop any running backend processes
Write-Host "🛑 Stopping existing backend processes..." -ForegroundColor Yellow
Get-Process -Name "main" -ErrorAction SilentlyContinue | Stop-Process -Force
Get-Process -Name "go" -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Seconds 2

# Navigate to backend directory
Set-Location "D:\Project\clone_app_akuntansi\accounting_proj\backend"

# Build and start backend
Write-Host "🔨 Building backend..." -ForegroundColor Yellow
go build -o main.exe .

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Build successful!" -ForegroundColor Green
    
    # Start backend in background
    Write-Host "🚀 Starting backend server..." -ForegroundColor Yellow
    Start-Process -FilePath ".\main.exe" -WindowStyle Minimized
    
    # Wait for backend to start
    Write-Host "⏳ Waiting for backend to start..." -ForegroundColor Yellow
    $timeout = 30
    $elapsed = 0
    
    while ($elapsed -lt $timeout) {
        if (Test-Backend) {
            Write-Host "✅ Backend is running!" -ForegroundColor Green
            break
        }
        Start-Sleep -Seconds 1
        $elapsed++
        Write-Progress -Activity "Starting Backend" -Status "Elapsed: $elapsed seconds" -PercentComplete (($elapsed / $timeout) * 100)
    }
    
    if ($elapsed -ge $timeout) {
        Write-Host "❌ Backend failed to start within timeout!" -ForegroundColor Red
        exit 1
    }
    
    # Test authentication
    Write-Host "🧪 Testing authentication..." -ForegroundColor Yellow
    
    try {
        # Test login
        $loginBody = @{
            email = "admin@company.com"
            password = "admin123"
        } | ConvertTo-Json
        
        $headers = @{
            "Content-Type" = "application/json"
        }
        
        $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Body $loginBody -Headers $headers
        
        if ($loginResponse.access_token -or $loginResponse.token) {
            $token = if ($loginResponse.access_token) { $loginResponse.access_token } else { $loginResponse.token }
            Write-Host "✅ Login successful! Token received: $($token.Substring(0, 20))..." -ForegroundColor Green
            
            # Test protected endpoint
            $authHeaders = @{
                "Authorization" = "Bearer $token"
                "Content-Type" = "application/json"
            }
            
            try {
                $profileResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/profile" -Method Get -Headers $authHeaders
                Write-Host "✅ Protected endpoint test successful!" -ForegroundColor Green
                Write-Host "   User: $($profileResponse.user.username)" -ForegroundColor Cyan
                Write-Host "   Role: $($profileResponse.user.role)" -ForegroundColor Cyan
            } catch {
                $errorDetails = $_.Exception.Response.GetResponseStream()
                $reader = New-Object System.IO.StreamReader($errorDetails)
                $errorBody = $reader.ReadToEnd()
                Write-Host "❌ Protected endpoint test failed!" -ForegroundColor Red
                Write-Host "   Error: $errorBody" -ForegroundColor Red
            }
        } else {
            Write-Host "❌ Login failed - No token received!" -ForegroundColor Red
        }
        
    } catch {
        Write-Host "❌ Authentication test failed!" -ForegroundColor Red
        Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
    }
    
    # Show available URLs
    Write-Host "`n🌐 Available URLs:" -ForegroundColor Green
    Write-Host "   Backend API: http://localhost:8080/api/v1/health" -ForegroundColor Cyan
    Write-Host "   Enhanced Swagger: http://localhost:8080/enhanced-swagger/" -ForegroundColor Cyan
    Write-Host "   Debug Tool: http://localhost:8080/swagger-debug.html" -ForegroundColor Cyan
    
    # Open browser to diagnostic tool
    Write-Host "`n🔍 Opening diagnostic tool in browser..." -ForegroundColor Yellow
    Start-Process "http://localhost:8080/swagger-debug.html"
    
    Write-Host "`n✅ Test completed! Check the diagnostic tool for detailed analysis." -ForegroundColor Green
    Write-Host "   Press Ctrl+C to stop the backend server when done." -ForegroundColor Yellow
    
} else {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "🔧 Fix Summary:" -ForegroundColor Green
Write-Host "   ✅ Enhanced authentication helper with timing fixes"
Write-Host "   ✅ Fixed request interceptor conflicts"
Write-Host "   ✅ Improved JWT middleware debugging"
Write-Host "   ✅ Enhanced auto-authorization with retry"
Write-Host "   ✅ Added comprehensive diagnostic tool"
Write-Host "=========================================" -ForegroundColor Cyan