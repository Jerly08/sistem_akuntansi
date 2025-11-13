# Test script to verify the closing history API is working

Write-Host "==================================" -ForegroundColor Cyan
Write-Host "Testing Closing History API" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan

# Set API endpoint
$baseUrl = "http://localhost:8080"
$apiUrl = "$baseUrl/api/v1/fiscal-closing/history"

# First, check if the server is running
Write-Host "`nChecking if server is running..." -ForegroundColor Yellow
try {
    $healthCheck = Invoke-RestMethod -Uri "$baseUrl/api/v1/health" -Method GET -ErrorAction Stop
    Write-Host "✓ Server is running" -ForegroundColor Green
} catch {
    Write-Host "✗ Server is not running. Please start the backend server first." -ForegroundColor Red
    Write-Host "Run: cd backend && go run main.go" -ForegroundColor Yellow
    exit 1
}

# Login first to get JWT token
Write-Host "`nLogging in to get authentication token..." -ForegroundColor Yellow
$loginBody = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.data.token
    Write-Host "✓ Login successful" -ForegroundColor Green
} catch {
    Write-Host "✗ Login failed. Using test token instead." -ForegroundColor Red
    # Use a test token or prompt for one
    $token = "your-test-token-here"
}

# Prepare headers with authentication
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Test the fiscal closing history endpoint
Write-Host "`nTesting /api/v1/fiscal-closing/history endpoint..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri $apiUrl -Method GET -Headers $headers
    
    if ($response.success -eq $true) {
        Write-Host "✓ API call successful" -ForegroundColor Green
        
        if ($response.data -and $response.data.Count -gt 0) {
            Write-Host "✓ Found $($response.data.Count) closed period(s)" -ForegroundColor Green
            Write-Host "`nClosed Periods:" -ForegroundColor Cyan
            
            foreach ($period in $response.data) {
                Write-Host "  - Date: $($period.entry_date) | Code: $($period.code) | Description: $($period.description)" -ForegroundColor White
            }
        } else {
            Write-Host "⚠ No closed periods found in the response" -ForegroundColor Yellow
            Write-Host "This means either:" -ForegroundColor Yellow
            Write-Host "  1. No closing has been performed yet" -ForegroundColor Yellow
            Write-Host "  2. The closing data exists but with different format" -ForegroundColor Yellow
        }
    } else {
        Write-Host "✗ API returned success=false" -ForegroundColor Red
        Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ API call failed with error: $_" -ForegroundColor Red
    Write-Host "Make sure the backend server is running and you're logged in" -ForegroundColor Yellow
}

Write-Host "`n==================================" -ForegroundColor Cyan
Write-Host "Test Complete" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan

# Additional diagnostic: Check what's in the database directly
Write-Host "`nRunning database diagnostic..." -ForegroundColor Yellow
Write-Host "This will check what closing data exists in the database." -ForegroundColor Gray

$runDiagnostic = Read-Host "Do you want to run the database diagnostic? (y/n)"
if ($runDiagnostic -eq 'y') {
    Write-Host "`nRunning check_closing_data.go..." -ForegroundColor Yellow
    Push-Location
    Set-Location "$PSScriptRoot"
    go run cmd/check_closing_data.go
    Pop-Location
}