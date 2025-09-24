# SSOT Reports Test Script
# Test all new SSOT report endpoints

$baseURL = "http://localhost:5000/api/v1"
$token = ""  # Will be set after login

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "   SSOT Reports Integration Test     " -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# Function to display test results
function Show-TestResult {
    param(
        [string]$TestName,
        [bool]$Success,
        [string]$Details = ""
    )
    
    if ($Success) {
        Write-Host "✓" -ForegroundColor Green -NoNewline
        Write-Host " $TestName" -ForegroundColor White
        if ($Details) {
            Write-Host "  $Details" -ForegroundColor Gray
        }
    } else {
        Write-Host "✗" -ForegroundColor Red -NoNewline
        Write-Host " $TestName" -ForegroundColor White
        if ($Details) {
            Write-Host "  Error: $Details" -ForegroundColor Yellow
        }
    }
}

# 1. Login first
Write-Host "1. Authenticating..." -ForegroundColor Yellow
$loginBody = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseURL/auth/login" `
        -Method POST `
        -Body $loginBody `
        -ContentType "application/json"
    
    $token = $loginResponse.data.access_token
    Show-TestResult "Authentication" $true "Token obtained successfully"
} catch {
    Show-TestResult "Authentication" $false $_.Exception.Message
    Write-Host ""
    Write-Host "Cannot proceed without authentication. Exiting..." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "2. Testing SSOT Report Endpoints..." -ForegroundColor Yellow
Write-Host ""

# Headers with token
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Test dates
$startDate = "2025-01-01"
$endDate = "2025-12-31"
$asOfDate = Get-Date -Format "yyyy-MM-dd"

# Test 1: Sales Summary Report
Write-Host "Testing Sales Summary Report..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseURL/ssot-reports/sales-summary?start_date=$startDate&end_date=$endDate" `
        -Method GET `
        -Headers $headers
    
    if ($response.status -eq "success") {
        Show-TestResult "Sales Summary Report" $true "Generated successfully"
        if ($response.data) {
            Write-Host "  Total Revenue: $($response.data.total_revenue)" -ForegroundColor Gray
            Write-Host "  Net Revenue: $($response.data.net_revenue)" -ForegroundColor Gray
        }
    } else {
        Show-TestResult "Sales Summary Report" $false $response.message
    }
} catch {
    Show-TestResult "Sales Summary Report" $false $_.Exception.Message
}

Write-Host ""

# Test 2: Vendor Analysis Report
Write-Host "Testing Vendor Analysis Report..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseURL/ssot-reports/vendor-analysis?start_date=$startDate&end_date=$endDate" `
        -Method GET `
        -Headers $headers
    
    if ($response.status -eq "success") {
        Show-TestResult "Vendor Analysis Report" $true "Generated successfully"
        if ($response.data) {
            Write-Host "  Total Vendors: $($response.data.total_vendors)" -ForegroundColor Gray
            Write-Host "  Total Purchases: $($response.data.total_purchases)" -ForegroundColor Gray
        }
    } else {
        Show-TestResult "Vendor Analysis Report" $false $response.message
    }
} catch {
    Show-TestResult "Vendor Analysis Report" $false $_.Exception.Message
}

Write-Host ""

# Test 3: Trial Balance
Write-Host "Testing Trial Balance..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseURL/ssot-reports/trial-balance?as_of_date=$asOfDate" `
        -Method GET `
        -Headers $headers
    
    if ($response.status -eq "success") {
        Show-TestResult "Trial Balance" $true "Generated successfully"
        if ($response.data) {
            Write-Host "  Total Debits: $($response.data.total_debits)" -ForegroundColor Gray
            Write-Host "  Total Credits: $($response.data.total_credits)" -ForegroundColor Gray
            Write-Host "  Is Balanced: $($response.data.is_balanced)" -ForegroundColor Gray
        }
    } else {
        Show-TestResult "Trial Balance" $false $response.message
    }
} catch {
    Show-TestResult "Trial Balance" $false $_.Exception.Message
}

Write-Host ""

# Test 4: General Ledger
Write-Host "Testing General Ledger..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseURL/ssot-reports/general-ledger?start_date=$startDate&end_date=$endDate" `
        -Method GET `
        -Headers $headers
    
    if ($response.status -eq "success") {
        Show-TestResult "General Ledger" $true "Generated successfully"
        if ($response.data) {
            Write-Host "  Total Entries: $($response.data.entries.Count)" -ForegroundColor Gray
        }
    } else {
        Show-TestResult "General Ledger" $false $response.message
    }
} catch {
    Show-TestResult "General Ledger" $false $_.Exception.Message
}

Write-Host ""

# Test 5: Journal Entry Analysis
Write-Host "Testing Journal Entry Analysis..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseURL/ssot-reports/journal-analysis?start_date=$startDate&end_date=$endDate" `
        -Method GET `
        -Headers $headers
    
    if ($response.status -eq "success") {
        Show-TestResult "Journal Entry Analysis" $true "Generated successfully"
        if ($response.data) {
            Write-Host "  Total Entries: $($response.data.total_entries)" -ForegroundColor Gray
            Write-Host "  Posted Entries: $($response.data.posted_entries)" -ForegroundColor Gray
            Write-Host "  Draft Entries: $($response.data.draft_entries)" -ForegroundColor Gray
        }
    } else {
        Show-TestResult "Journal Entry Analysis" $false $response.message
    }
} catch {
    Show-TestResult "Journal Entry Analysis" $false $_.Exception.Message
}

Write-Host ""

# Test 6: Integrated Reports (All reports at once)
Write-Host "Testing Integrated Reports..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseURL/ssot-reports/integrated?start_date=$startDate&end_date=$endDate" `
        -Method GET `
        -Headers $headers
    
    if ($response.status -eq "success") {
        Show-TestResult "Integrated Reports" $true "All reports generated successfully"
        if ($response.data) {
            Write-Host "  Reports included:" -ForegroundColor Gray
            if ($response.data.profit_loss) { Write-Host "    ✓ Profit & Loss" -ForegroundColor Green }
            if ($response.data.balance_sheet) { Write-Host "    ✓ Balance Sheet" -ForegroundColor Green }
            if ($response.data.cash_flow) { Write-Host "    ✓ Cash Flow" -ForegroundColor Green }
            if ($response.data.sales_summary) { Write-Host "    ✓ Sales Summary" -ForegroundColor Green }
            if ($response.data.vendor_analysis) { Write-Host "    ✓ Vendor Analysis" -ForegroundColor Green }
            if ($response.data.trial_balance) { Write-Host "    ✓ Trial Balance" -ForegroundColor Green }
            if ($response.data.general_ledger) { Write-Host "    ✓ General Ledger" -ForegroundColor Green }
            if ($response.data.journal_analysis) { Write-Host "    ✓ Journal Analysis" -ForegroundColor Green }
        }
    } else {
        Show-TestResult "Integrated Reports" $false $response.message
    }
} catch {
    Show-TestResult "Integrated Reports" $false $_.Exception.Message
}

Write-Host ""

# Test 7: Report Status Endpoint
Write-Host "Testing Report Status..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseURL/ssot-reports/status" `
        -Method GET `
        -Headers $headers
    
    if ($response.status -eq "success") {
        Show-TestResult "Report Status" $true "Status retrieved successfully"
        if ($response.data.available_reports) {
            Write-Host "  Available Reports: $($response.data.available_reports -join ', ')" -ForegroundColor Gray
        }
    } else {
        Show-TestResult "Report Status" $false $response.message
    }
} catch {
    Show-TestResult "Report Status" $false $_.Exception.Message
}

Write-Host ""
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "        Test Complete                " -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Note: If any tests failed, check:" -ForegroundColor Yellow
Write-Host "1. Backend server is running" -ForegroundColor White
Write-Host "2. Database has SSOT journal entries" -ForegroundColor White
Write-Host "3. All required services are initialized" -ForegroundColor White
Write-Host "4. User has proper permissions" -ForegroundColor White