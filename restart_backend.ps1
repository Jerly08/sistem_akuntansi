# Script to restart backend server
Write-Host "Stopping any running backend processes..." -ForegroundColor Yellow

# Kill any existing backend processes
Get-Process -Name "main" -ErrorAction SilentlyContinue | Stop-Process -Force
Get-Process -Name "go" -ErrorAction SilentlyContinue | Stop-Process -Force

# Wait a moment
Start-Sleep -Seconds 2

Write-Host "Building backend..." -ForegroundColor Green
cd backend
go build -o main cmd/main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful! Starting server..." -ForegroundColor Green
    ./main
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
