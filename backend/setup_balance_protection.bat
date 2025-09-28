@echo off
echo ================================================================
echo 🛡️  BALANCE PROTECTION SETUP
echo ================================================================
echo.
echo This script will setup automatic balance synchronization system
echo to prevent balance mismatch issues in the accounting system.
echo.
echo What this does:
echo   ✅ Install database triggers for auto-sync
echo   ✅ Install monitoring system  
echo   ✅ Install manual sync functions
echo   ✅ Fix any existing balance issues
echo.
echo ================================================================
echo.

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Go is not installed or not in PATH
    echo Please install Go first: https://golang.org/dl/
    pause
    exit /b 1
)

REM Check if .env file exists
if not exist ".env" (
    echo ❌ .env file not found
    echo Please make sure you're in the backend directory with .env file
    pause
    exit /b 1
)

echo 🚀 Running balance protection setup...
echo.

REM Run the setup script
go run cmd/scripts/setup_balance_sync_auto.go

if %errorlevel% equ 0 (
    echo.
    echo ================================================================
    echo ✅ SUCCESS: Balance Protection System Installed!
    echo ================================================================
    echo.
    echo Your accounting system is now protected against balance mismatches.
    echo.
    echo 💡 What's installed:
    echo   • Automatic balance sync triggers
    echo   • Real-time monitoring system
    echo   • Manual sync functions
    echo   • Performance optimizations
    echo.
    echo 🔧 Manual commands available:
    echo   • Health check: psql -d DATABASE_URL -c "SELECT * FROM account_balance_monitoring WHERE status='MISMATCH';"
    echo   • Manual sync:  psql -d DATABASE_URL -c "SELECT * FROM sync_account_balances();"
    echo.
    echo 📚 For more info, read: BALANCE_PREVENTION_GUIDE.md
    echo.
) else (
    echo.
    echo ❌ FAILED: Setup encountered errors
    echo Please check the error messages above and try again.
    echo.
)

pause