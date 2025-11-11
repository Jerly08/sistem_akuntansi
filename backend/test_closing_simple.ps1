# Simple Test Script for Closing Period Data
Write-Host "========================================" -ForegroundColor Cyan
Write-Host " Testing Closed Period Dropdown Issue" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080"

# Test 1: Check backend
Write-Host "[1] Checking backend..." -ForegroundColor Yellow
try {
    $null = Invoke-WebRequest -Uri "$baseUrl/health" -Method GET -ErrorAction Stop -TimeoutSec 5
    Write-Host "    OK - Backend is running" -ForegroundColor Green
} catch {
    Write-Host "    FAIL - Backend not running" -ForegroundColor Red
    Write-Host "    Start with: go run main.go" -ForegroundColor White
    exit 1
}

Write-Host ""

# Test 2: Check API endpoint
Write-Host "[2] Testing API: /api/v1/fiscal-closing/history" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/fiscal-closing/history" -Method GET -ErrorAction Stop
    
    if ($response.success) {
        $count = 0
        if ($response.data) {
            $count = $response.data.Count
        }
        
        Write-Host "    OK - API works" -ForegroundColor Green
        Write-Host "    Found: $count closing entries" -ForegroundColor Cyan
        Write-Host ""
        
        if ($count -eq 0) {
            Write-Host "========================================" -ForegroundColor Red
            Write-Host " PROBLEM: NO CLOSING DATA IN DATABASE" -ForegroundColor Red
            Write-Host "========================================" -ForegroundColor Red
            Write-Host ""
            Write-Host "This is why dropdown is empty!" -ForegroundColor Yellow
            Write-Host ""
            Write-Host "SOLUTION:" -ForegroundColor Green
            Write-Host "1. Open the application" -ForegroundColor White
            Write-Host "2. Go to Period Closing page" -ForegroundColor White
            Write-Host "3. Select fiscal year end date" -ForegroundColor White
            Write-Host "4. Click Execute Closing" -ForegroundColor White
            Write-Host "5. Test dropdown again" -ForegroundColor White
            Write-Host ""
        } else {
            Write-Host "Closing entries found:" -ForegroundColor Cyan
            $counter = 1
            foreach ($entry in $response.data) {
                Write-Host ""
                Write-Host "  Entry $counter" -ForegroundColor White
                Write-Host "  Code: $($entry.code)" -ForegroundColor Gray
                Write-Host "  Date: $($entry.entry_date)" -ForegroundColor Gray
                Write-Host "  Desc: $($entry.description)" -ForegroundColor Gray
                $counter++
                if ($counter -gt 3) { break }
            }
            Write-Host ""
            Write-Host "========================================" -ForegroundColor Green
            Write-Host " DATA EXISTS IN BACKEND" -ForegroundColor Green
            Write-Host "========================================" -ForegroundColor Green
            Write-Host ""
            Write-Host "The problem might be in FRONTEND." -ForegroundColor Yellow
            Write-Host "Check browser console (F12) for errors." -ForegroundColor White
            Write-Host ""
        }
    } else {
        Write-Host "    FAIL - API returned success=false" -ForegroundColor Red
        Write-Host "    Error: $($response.error)" -ForegroundColor Red
    }
} catch {
    Write-Host "    FAIL - API request failed" -ForegroundColor Red
    Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "Test completed: $(Get-Date -Format 'HH:mm:ss')" -ForegroundColor Gray
