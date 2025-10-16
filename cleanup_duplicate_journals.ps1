# Cleanup Duplicate Sales Journals Script
# This script removes duplicate journal entries and recalculates COA balances

Write-Host "=======================================" -ForegroundColor Cyan
Write-Host "  Duplicate Journal Cleanup Tool" -ForegroundColor Cyan
Write-Host "=======================================" -ForegroundColor Cyan
Write-Host ""

# Check if backend directory exists
if (-not (Test-Path ".\backend")) {
    Write-Host "❌ Error: backend directory not found!" -ForegroundColor Red
    Write-Host "Please run this script from the project root directory." -ForegroundColor Yellow
    exit 1
}

# Navigate to backend directory
Set-Location backend

Write-Host "🔧 Running cleanup script..." -ForegroundColor Yellow
Write-Host ""

# Run the cleanup script
go run scripts/cleanup_duplicate_sales_journals.go

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "=======================================" -ForegroundColor Green
    Write-Host "  ✅ Cleanup Completed Successfully!" -ForegroundColor Green
    Write-Host "=======================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "1. Restart backend: .\restart_backend.ps1" -ForegroundColor White
    Write-Host "2. Verify COA balances in frontend" -ForegroundColor White
    Write-Host "3. Test new sales transactions" -ForegroundColor White
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "=======================================" -ForegroundColor Red
    Write-Host "  ❌ Cleanup Failed!" -ForegroundColor Red
    Write-Host "=======================================" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please check the error messages above." -ForegroundColor Yellow
    Write-Host ""
}

# Return to project root
Set-Location ..

# Wait for user input
Write-Host "Press any key to continue..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

