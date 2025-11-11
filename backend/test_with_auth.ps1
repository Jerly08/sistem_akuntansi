# Test Closing History with Authentication
Write-Host "========================================" -ForegroundColor Cyan
Write-Host " Testing Closed Period with Auth" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080"

Write-Host "NOTE: This API requires authentication." -ForegroundColor Yellow
Write-Host "Please test using one of these methods:" -ForegroundColor Yellow
Write-Host ""

Write-Host "METHOD 1: Test via Frontend (Recommended)" -ForegroundColor Green
Write-Host "--------------------------------------" -ForegroundColor Gray
Write-Host "1. Open the application in browser" -ForegroundColor White
Write-Host "2. Login with your credentials" -ForegroundColor White
Write-Host "3. Open DevTools (F12)" -ForegroundColor White
Write-Host "4. Go to Network tab" -ForegroundColor White
Write-Host "5. Navigate to Reports > Balance Sheet" -ForegroundColor White
Write-Host "6. Click 'Closed Period' dropdown" -ForegroundColor White
Write-Host "7. Look for request to:" -ForegroundColor White
Write-Host "   /api/v1/fiscal-closing/history" -ForegroundColor Cyan
Write-Host "8. Check the response:" -ForegroundColor White
Write-Host ""
Write-Host "   If response shows: {success: true, data: []}" -ForegroundColor Yellow
Write-Host "   → NO DATA - You need to perform fiscal year closing first" -ForegroundColor White
Write-Host ""
Write-Host "   If response shows: {success: true, data: [...]}" -ForegroundColor Yellow
Write-Host "   → DATA EXISTS - Problem is in frontend rendering" -ForegroundColor White
Write-Host ""

Write-Host "METHOD 2: Direct Browser Test" -ForegroundColor Green
Write-Host "--------------------------------------" -ForegroundColor Gray
Write-Host "1. Login to the application" -ForegroundColor White
Write-Host "2. Open browser console (F12)" -ForegroundColor White
Write-Host "3. Paste this code:" -ForegroundColor White
Write-Host ""
Write-Host "   fetch('http://localhost:8080/api/v1/fiscal-closing/history', {" -ForegroundColor Cyan
Write-Host "     headers: {" -ForegroundColor Cyan
Write-Host "       'Authorization': 'Bearer ' + localStorage.getItem('token')" -ForegroundColor Cyan
Write-Host "     }" -ForegroundColor Cyan
Write-Host "   })" -ForegroundColor Cyan
Write-Host "   .then(r => r.json())" -ForegroundColor Cyan
Write-Host "   .then(d => console.log('Result:', d))" -ForegroundColor Cyan
Write-Host ""
Write-Host "4. Check console output" -ForegroundColor White
Write-Host ""

Write-Host "METHOD 3: SQL Query (If you have DB access)" -ForegroundColor Green
Write-Host "--------------------------------------" -ForegroundColor Gray
Write-Host "SELECT COUNT(*) as total FROM journal_entries" -ForegroundColor Cyan
Write-Host "WHERE reference_type = 'CLOSING';" -ForegroundColor Cyan
Write-Host ""
Write-Host "If result is 0: You need to perform fiscal year closing" -ForegroundColor White
Write-Host "If result > 0: Data exists, problem is in frontend" -ForegroundColor White
Write-Host ""

Write-Host "QUICK DIAGNOSIS:" -ForegroundColor Yellow
Write-Host "--------------------------------------" -ForegroundColor Gray
Write-Host "The dropdown is empty because either:" -ForegroundColor White
Write-Host "  1. No fiscal year closing has been performed (MOST LIKELY)" -ForegroundColor White
Write-Host "  2. Frontend failed to fetch the data" -ForegroundColor White
Write-Host "  3. Frontend failed to render the dropdown options" -ForegroundColor White
Write-Host ""

Write-Host "SOLUTION:" -ForegroundColor Green
Write-Host "--------------------------------------" -ForegroundColor Gray
Write-Host "1. In the application, go to:" -ForegroundColor White
Write-Host "   Settings > Period Closing" -ForegroundColor Cyan
Write-Host "   (or wherever the fiscal year closing feature is)" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Perform a fiscal year closing:" -ForegroundColor White
Write-Host "   - Select fiscal year end date (e.g., 31/12/2025)" -ForegroundColor White
Write-Host "   - Click 'Preview' to see what will be closed" -ForegroundColor White
Write-Host "   - Click 'Execute Closing' to perform the closing" -ForegroundColor White
Write-Host ""
Write-Host "3. After closing is complete:" -ForegroundColor White
Write-Host "   - Go back to Reports > Balance Sheet" -ForegroundColor White
Write-Host "   - Click 'Closed Period' dropdown" -ForegroundColor White
Write-Host "   - You should now see the closed period in the list" -ForegroundColor White
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test completed at: $(Get-Date -Format 'HH:mm:ss')" -ForegroundColor Gray
