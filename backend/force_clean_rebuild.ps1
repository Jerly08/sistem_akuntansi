# Force Clean and Rebuild Script for Windows

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "🔧 FORCE CLEAN & REBUILD" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan

# 1. Kill ALL Go processes
Write-Host ""
Write-Host "1️⃣  Killing ALL Go processes..." -ForegroundColor Yellow
Get-Process | Where-Object {$_.ProcessName -like "*go*" -or $_.ProcessName -like "*main*"} | Stop-Process -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 2

# 2. Clean Go cache
Write-Host ""
Write-Host "2️⃣  Cleaning Go build cache..." -ForegroundColor Yellow
go clean -cache
go clean -modcache  
go clean -testcache

# 3. Remove binary files
Write-Host ""
Write-Host "3️⃣  Removing old binaries..." -ForegroundColor Yellow
Remove-Item -Path "main.exe" -Force -ErrorAction SilentlyContinue
Remove-Item -Path "main" -Force -ErrorAction SilentlyContinue
Remove-Item -Path "*.exe" -Force -ErrorAction SilentlyContinue
Remove-Item -Path "tmp" -Recurse -Force -ErrorAction SilentlyContinue

# 4. Force git reset to remote
Write-Host ""
Write-Host "4️⃣  Force resetting to remote version..." -ForegroundColor Yellow
git fetch origin main
git reset --hard origin/main

# 5. Verify critical file
Write-Host ""
Write-Host "5️⃣  Verifying critical file update..." -ForegroundColor Yellow
$fileContent = Get-Content "services\unified_period_closing_service.go" -Raw
if ($fileContent -match "absBalance := netBalance\.Abs\(\)") {
    Write-Host "   ✅ Code is updated correctly" -ForegroundColor Green
} else {
    Write-Host "   ❌ Code is NOT updated!" -ForegroundColor Red
    Write-Host "   Checking line 140-150..." -ForegroundColor Yellow
    $lines = Get-Content "services\unified_period_closing_service.go"
    $lines[139..149] | ForEach-Object {"   Line $($lines.IndexOf($_) + 1): $_"}
}

# 6. Re-download dependencies
Write-Host ""
Write-Host "6️⃣  Re-downloading dependencies..." -ForegroundColor Yellow
go mod download
go mod tidy

# 7. Fix database
Write-Host ""
Write-Host "7️⃣  Running database fix..." -ForegroundColor Yellow
go run cmd/fix_period_closing_comprehensive.go

# 8. Build fresh
Write-Host ""
Write-Host "8️⃣  Building fresh binary..." -ForegroundColor Yellow
go build -a -o main.exe main.go

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "✅ Clean rebuild completed!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Now run: .\main.exe" -ForegroundColor Yellow
Write-Host "Or: go run main.go" -ForegroundColor Yellow