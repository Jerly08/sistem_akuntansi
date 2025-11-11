# Test Closing History API
# Script untuk test endpoint closing history setelah perbaikan

Write-Host "=== Testing Closing History API ===" -ForegroundColor Cyan
Write-Host ""

# Configuration
$baseUrl = "http://localhost:8080"
$endpoint = "/api/v1/fiscal-closing/history"

Write-Host "Testing endpoint: $baseUrl$endpoint" -ForegroundColor Yellow
Write-Host ""

# Test 1: Check if backend is running
Write-Host "1. Checking if backend is running..." -ForegroundColor Green
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/health" -Method GET -ErrorAction Stop
    Write-Host "   ✅ Backend is running" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Backend is not running!" -ForegroundColor Red
    Write-Host "   Please start backend first: go run main.go" -ForegroundColor Yellow
    exit 1
}

Write-Host ""

# Test 2: Test closing history endpoint
Write-Host "2. Testing closing history endpoint..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl$endpoint" -Method GET -ErrorAction Stop
    
    if ($response.success -eq $true) {
        Write-Host "   ✅ API call successful" -ForegroundColor Green
        
        $dataCount = $response.data.Count
        Write-Host "   ✅ Found $dataCount closing entries" -ForegroundColor Green
        
        if ($dataCount -gt 0) {
            Write-Host ""
            Write-Host "   Latest closing entry:" -ForegroundColor Cyan
            $latest = $response.data[0]
            Write-Host "   - ID: $($latest.id)" -ForegroundColor White
            Write-Host "   - Code: $($latest.code)" -ForegroundColor White
            Write-Host "   - Description: $($latest.description)" -ForegroundColor White
            Write-Host "   - Entry Date: $($latest.entry_date)" -ForegroundColor White
            Write-Host "   - Total Debit: $($latest.total_debit)" -ForegroundColor White
            Write-Host ""
            Write-Host "   🎉 SUKSES! Closing history dapat terbaca dengan benar!" -ForegroundColor Green
        } else {
            Write-Host "   ⚠️  No closing entries found in database" -ForegroundColor Yellow
            Write-Host "   This might be expected if no fiscal year closing has been performed yet" -ForegroundColor Yellow
        }
    } else {
        Write-Host "   ❌ API returned success=false" -ForegroundColor Red
        Write-Host "   Error: $($response.error)" -ForegroundColor Red
    }
} catch {
    Write-Host "   ❌ API call failed!" -ForegroundColor Red
    Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Test Complete ===" -ForegroundColor Cyan
