# Restart backend with latest code from GitHub

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "RESTART BACKEND WITH LATEST CODE" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan

# 1. Stop backend if running
Write-Host ""
Write-Host "1️⃣  Stopping backend server..." -ForegroundColor Yellow
$processes = Get-Process | Where-Object {$_.ProcessName -like "*go*" -or $_.CommandLine -like "*main.go*"}
if ($processes) {
    $processes | Stop-Process -Force
    Write-Host "   ✅ Stopped backend process" -ForegroundColor Green
    Start-Sleep -Seconds 2
} else {
    Write-Host "   ℹ️  Backend not running" -ForegroundColor Gray
}

# 2. Pull latest code
Write-Host ""
Write-Host "2️⃣  Pulling latest code from GitHub..." -ForegroundColor Yellow
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location "$scriptDir\.."

git pull origin main

if ($LASTEXITCODE -ne 0) {
    Write-Host "   ❌ Failed to pull latest code" -ForegroundColor Red
    Write-Host "   Please check your git configuration" -ForegroundColor Red
    exit 1
}

Write-Host "   ✅ Code updated successfully" -ForegroundColor Green

# 3. Run database fix script
Write-Host ""
Write-Host "3️⃣  Running database cleanup/fix..." -ForegroundColor Yellow
Set-Location "backend"
go run cmd/verify_and_fix_pc.go

if ($LASTEXITCODE -ne 0) {
    Write-Host "   ⚠️  Warning: Database fix had issues" -ForegroundColor Yellow
}

# 4. Start backend
Write-Host ""
Write-Host "4️⃣  Starting backend server..." -ForegroundColor Yellow
Write-Host "   Press Ctrl+C to stop" -ForegroundColor Gray
Write-Host ""

go run main.go
