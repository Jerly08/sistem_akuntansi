# Quick Test Script for Closing Period Data
# This script will help diagnose why closed periods are not showing

Write-Host "╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║         TESTING: Closed Period Dropdown Issue                 ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080"
$apiEndpoint = "/api/v1/fiscal-closing/history"

# Step 1: Check if backend is running
Write-Host "STEP 1: Checking if backend is running..." -ForegroundColor Yellow
Write-Host "------------------------------------------------------" -ForegroundColor Gray
try {
    $healthCheck = Invoke-WebRequest -Uri "$baseUrl/health" -Method GET -ErrorAction Stop -TimeoutSec 5
    Write-Host "✅ Backend is RUNNING" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "❌ Backend is NOT running!" -ForegroundColor Red
    Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    Write-Host "ACTION REQUIRED:" -ForegroundColor Yellow
    Write-Host "   1. Start MySQL database" -ForegroundColor White
    Write-Host "   2. Run: cd backend; go run main.go" -ForegroundColor White
    Write-Host "   3. Or run: cd backend; .\accounting_app.exe" -ForegroundColor White
    Write-Host ""
    exit 1
}

# Step 2: Test fiscal closing history API
Write-Host "STEP 2: Testing Fiscal Closing History API..." -ForegroundColor Yellow
Write-Host "------------------------------------------------------" -ForegroundColor Gray
Write-Host "URL: $baseUrl$apiEndpoint" -ForegroundColor Gray
Write-Host ""

try {
    $response = Invoke-RestMethod -Uri "$baseUrl$apiEndpoint" -Method GET -ErrorAction Stop
    
    Write-Host "✅ API Request SUCCESS" -ForegroundColor Green
    Write-Host ""
    
    # Check response structure
    if ($response.success -eq $true) {
        Write-Host "Response Structure:" -ForegroundColor Cyan
        Write-Host "  - success: $($response.success)" -ForegroundColor White
        
        if ($response.data) {
            $dataCount = $response.data.Count
            Write-Host "  - data count: $dataCount entries" -ForegroundColor White
            Write-Host ""
            
            if ($dataCount -eq 0) {
                Write-Host "❌ PROBLEM FOUND: No closing data in database!" -ForegroundColor Red
                Write-Host ""
                Write-Host "ROOT CAUSE:" -ForegroundColor Yellow
                Write-Host "  The database has NO fiscal year closing entries." -ForegroundColor White
                Write-Host "  This is why the dropdown shows no options." -ForegroundColor White
                Write-Host ""
                Write-Host "SOLUTION:" -ForegroundColor Green
                Write-Host "  1. Open the application" -ForegroundColor White
                Write-Host "  2. Navigate to Period Closing page" -ForegroundColor White
                Write-Host "  3. Select fiscal year end date" -ForegroundColor White
                Write-Host "  4. Click Preview to see closing preview" -ForegroundColor White
                Write-Host "  5. Click Execute Closing to perform closing" -ForegroundColor White
                Write-Host "  6. Test dropdown again - data should appear" -ForegroundColor White
                Write-Host ""
                
                # Check if we can query the database
                Write-Host "ADDITIONAL DIAGNOSTIC:" -ForegroundColor Yellow
                Write-Host "  Running database query to check journal_entries..." -ForegroundColor Gray
                Write-Host ""
                
                # Try to get all journal entries to see if any exist
                try {
                    $allJournals = Invoke-RestMethod -Uri "$baseUrl/api/v1/journals" -Method GET -ErrorAction Stop
                    if ($allJournals.data -and $allJournals.data.Count -gt 0) {
                        Write-Host "  ℹ️  Found $($allJournals.data.Count) journal entries in system" -ForegroundColor Cyan
                        Write-Host "  → Database has data, but no CLOSING entries yet" -ForegroundColor White
                    } else {
                        Write-Host "  ⚠️  No journal entries found in system" -ForegroundColor Yellow
                    }
                } catch {
                    Write-Host "  ℹ️  Cannot check other journal entries" -ForegroundColor Gray
                }
                
            } else {
                Write-Host "✅ FOUND $dataCount closing period(s)!" -ForegroundColor Green
                Write-Host ""
                Write-Host "Closing Periods:" -ForegroundColor Cyan
                Write-Host "------------------------------------------------------" -ForegroundColor Gray
                
                $counter = 1
                foreach ($entry in $response.data) {
                    Write-Host ""
                    Write-Host "[$counter] Closing Entry:" -ForegroundColor White
                    Write-Host "    ID: $($entry.id)" -ForegroundColor Gray
                    Write-Host "    Code: $($entry.code)" -ForegroundColor Gray
                    Write-Host "    Date: $($entry.entry_date)" -ForegroundColor Gray
                    Write-Host "    Description: $($entry.description)" -ForegroundColor Gray
                    if ($entry.total_debit) {
                        Write-Host "    Amount: Rp $($entry.total_debit.ToString('N0'))" -ForegroundColor Gray
                    }
                    $counter++
                }
                
                Write-Host ""
                Write-Host "✅ Data exists in backend!" -ForegroundColor Green
                Write-Host ""
                Write-Host "NEXT STEP:" -ForegroundColor Yellow
                Write-Host "  The problem might be in the FRONTEND." -ForegroundColor White
                Write-Host "  Please check:" -ForegroundColor White
                Write-Host "  1. Open browser DevTools (F12)" -ForegroundColor White
                Write-Host "  2. Go to Network tab" -ForegroundColor White
                Write-Host "  3. Open Balance Sheet modal" -ForegroundColor White
                Write-Host "  4. Click 'Closed Period' dropdown" -ForegroundColor White
                Write-Host "  5. Check if API call is made" -ForegroundColor White
                Write-Host "  6. Check the response in Network tab" -ForegroundColor White
                Write-Host ""
                Write-Host "Expected dropdown options:" -ForegroundColor Cyan
                foreach ($entry in $response.data) {
                    $date = [DateTime]::Parse($entry.entry_date)
                    $formattedDate = $date.ToString("dd/MM/yyyy")
                    Write-Host "  - $formattedDate - $($entry.description)" -ForegroundColor White
                }
            }
        } else {
            Write-Host "⚠️  Response has no 'data' field" -ForegroundColor Yellow
            Write-Host "Full response:" -ForegroundColor Gray
            $response | ConvertTo-Json -Depth 5
        }
    } else {
        Write-Host "❌ API returned success=false" -ForegroundColor Red
        Write-Host "Error: $($response.error)" -ForegroundColor Red
        Write-Host ""
        Write-Host "Full response:" -ForegroundColor Gray
        $response | ConvertTo-Json -Depth 5
    }
    
} catch {
    Write-Host "❌ API Request FAILED" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    
    if ($_.Exception.Response) {
        Write-Host "Status Code: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Gray
    }
    
    Write-Host ""
    Write-Host "POSSIBLE CAUSES:" -ForegroundColor Yellow
    Write-Host "  1. Backend konstanta bug not fixed yet" -ForegroundColor White
    Write-Host "  2. Database connection issue" -ForegroundColor White
    Write-Host "  3. API endpoint not registered" -ForegroundColor White
    Write-Host ""
}

Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                         SUMMARY                                ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""
Write-Host "Test completed at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Gray
Write-Host ""
