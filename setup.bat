@echo off
echo ============================================
echo    SETUP SISTEM AKUNTANSI - BACKEND
echo ============================================

:: Check if we're in the right directory
if not exist "backend\go.mod" (
    echo Error: Script harus dijalankan di root directory project
    echo Current directory: %CD%
    echo Please navigate to the correct project directory
    pause
    exit /b 1
)

echo.
echo [1/5] Checking Go installation...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Go tidak terinstall atau tidak ada di PATH
    echo Silakan install Go terlebih dahulu dari https://golang.org
    pause
    exit /b 1
) else (
    go version
    echo ✅ Go sudah terinstall
)

echo.
echo [2/5] Setting up environment file...
cd backend
if not exist ".env" (
    echo Creating .env file from template...
    copy ".env.example" ".env" >nul
    echo ✅ .env file created - silakan edit sesuai konfigurasi database Anda
) else (
    echo ✅ .env file sudah ada
)

echo.
echo [3/5] Installing Go dependencies...
go mod tidy
if %errorlevel% neq 0 (
    echo Error: Gagal menginstall dependencies
    pause
    exit /b 1
)
echo ✅ Dependencies berhasil diinstall

echo.
echo [4/5] Running database constraint fix...
echo Menjalankan fix untuk constraint database yang bermasalah...
go run cmd/fix_accounts_constraint.go
if %errorlevel% neq 0 (
    echo Warning: Database constraint fix gagal - mungkin database belum dibuat
    echo Silakan buat database 'sistem_akuntansi' di PostgreSQL terlebih dahulu
) else (
    echo ✅ Database constraint fix berhasil
)

echo.
echo [5/5] Testing server startup...
echo Testing backend server...
timeout /t 2 /nobreak >nul
go run cmd/main.go &
set SERVER_PID=%!
timeout /t 5 /nobreak >nul
taskkill /F /PID %SERVER_PID% >nul 2>&1

echo.
echo ============================================
echo           SETUP COMPLETED!
echo ============================================
echo.
echo Setup selesai! Untuk menjalankan backend:
echo   1. Pastikan PostgreSQL sudah running
echo   2. Database 'sistem_akuntansi' sudah dibuat
echo   3. Edit file backend\.env sesuai konfigurasi database
echo   4. Jalankan: cd backend && go run cmd/main.go
echo.
echo Untuk frontend (di terminal terpisah):
echo   1. cd frontend
echo   2. npm install
echo   3. npm run dev
echo.
pause
