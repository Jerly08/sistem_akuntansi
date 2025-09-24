# SSOT High Priority Integration Test Script
# - Login
# - Create manual journal (balanced)
# - Post journal
# - Refresh balances
# Note: fill in credentials below before running.

param(
  [string]$BaseUrl = "http://localhost:8080",
  [string]$Username = "{{YOUR_USERNAME}}",
  [string]$Password = "{{YOUR_PASSWORD}}"
)

$ErrorActionPreference = 'Stop'

Write-Host "[1/5] Logging in..." -ForegroundColor Cyan
$loginBody = @{ username = $Username; password = $Password } | ConvertTo-Json
$loginResp = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/login" -Method Post -ContentType 'application/json' -Body $loginBody
$token = $loginResp.data.access_token
if (-not $token) { throw "Login failed. Please check credentials." }
$headers = @{ Authorization = "Bearer $token" }

Write-Host "[2/5] Creating a balanced manual journal entry..." -ForegroundColor Cyan
$today = (Get-Date).ToString('yyyy-MM-dd')
$createBody = @{
  source_type = "MANUAL"
  entry_date  = "$today"
  description = "High Priority Test Entry"
  reference   = "HP-TEST-$(Get-Date -Format yyyyMMddHHmmss)"
  auto_post   = $false
  created_by  = 1
  lines = @(
    @{ account_id = 1; description = "Debit Line";  debit_amount = 1000.00; credit_amount = 0.00 },
    @{ account_id = 2; description = "Credit Line"; debit_amount = 0.00;   credit_amount = 1000.00 }
  )
} | ConvertTo-Json -Depth 5

$createResp = Invoke-RestMethod -Uri "$BaseUrl/api/v1/journals" -Headers $headers -Method Post -ContentType 'application/json' -Body $createBody
$journalId = $createResp.id
if (-not $journalId) { throw "Failed to create journal entry" }
Write-Host "Created journal ID: $journalId" -ForegroundColor Green

Write-Host "[3/5] Posting the journal entry..." -ForegroundColor Cyan
$postResp = Invoke-RestMethod -Uri "$BaseUrl/api/v1/journals/$journalId/post" -Headers $headers -Method Put
Write-Host "Post response: $($postResp.message)" -ForegroundColor Green

Write-Host "[4/5] Refreshing account balances..." -ForegroundColor Cyan
$refreshResp = Invoke-RestMethod -Uri "$BaseUrl/api/v1/journals/account-balances/refresh" -Headers $headers -Method Post
Write-Host "Refresh response: $($refreshResp.message)" -ForegroundColor Green

Write-Host "[5/5] Fetching account balances..." -ForegroundColor Cyan
$balances = Invoke-RestMethod -Uri "$BaseUrl/api/v1/journals/account-balances" -Headers $headers -Method Get
$balances.data | Select-Object -First 5 | ForEach-Object { $_ | ConvertTo-Json }

Write-Host "Done." -ForegroundColor Green