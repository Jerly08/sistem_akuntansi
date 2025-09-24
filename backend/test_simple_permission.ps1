#!/usr/bin/env pwsh

# Test script simple untuk permission employee
Write-Host "Testing Employee Permission untuk Purchase Module" -ForegroundColor Cyan

$baseUrl = "http://localhost:8080/api/v1"

# Test health check
Write-Host "`n1. Testing Health Check..." -ForegroundColor Yellow
try {
    $healthResponse = Invoke-WebRequest -Uri "$baseUrl/health" -Method GET
    Write-Host "Server is running" -ForegroundColor Green
} catch {
    Write-Host "Server tidak berjalan" -ForegroundColor Red
    exit 1
}

# Test login admin (karena employee mungkin belum ada)
Write-Host "`n2. Testing Admin Login..." -ForegroundColor Yellow
$adminLoginBody = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

try {
    $adminLoginResponse = Invoke-WebRequest -Uri "$baseUrl/auth/login" -Method POST -Body $adminLoginBody -ContentType "application/json"
    $adminLoginData = $adminLoginResponse.Content | ConvertFrom-Json
    $token = $adminLoginData.token
    Write-Host "Admin login berhasil" -ForegroundColor Green
    Write-Host "User: $($adminLoginData.user.username) | Role: $($adminLoginData.user.role)" -ForegroundColor Cyan
} catch {
    Write-Host "Admin login gagal" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Test user permissions
Write-Host "`n3. Testing User Permissions..." -ForegroundColor Yellow
try {
    $permResponse = Invoke-WebRequest -Uri "$baseUrl/permissions/me" -Method GET -Headers $headers
    $permData = $permResponse.Content | ConvertFrom-Json
    Write-Host "User permissions retrieved" -ForegroundColor Green
    
    if ($permData.permissions.purchases) {
        $purchasePerm = $permData.permissions.purchases
        Write-Host "Purchase permissions:" -ForegroundColor Cyan
        Write-Host "- View: $($purchasePerm.can_view)" -ForegroundColor Cyan
        Write-Host "- Create: $($purchasePerm.can_create)" -ForegroundColor Cyan
        Write-Host "- Edit: $($purchasePerm.can_edit)" -ForegroundColor Cyan
        Write-Host "- Approve: $($purchasePerm.can_approve)" -ForegroundColor Cyan
    }
} catch {
    Write-Host "User permissions error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test purchase endpoint
Write-Host "`n4. Testing Purchase Endpoint..." -ForegroundColor Yellow
try {
    $purchaseResponse = Invoke-WebRequest -Uri "$baseUrl/purchases" -Method GET -Headers $headers
    Write-Host "Purchases endpoint accessible" -ForegroundColor Green
} catch {
    Write-Host "Purchases endpoint error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`nTest completed!" -ForegroundColor Cyan