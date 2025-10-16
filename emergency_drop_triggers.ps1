# Emergency: Drop all problematic triggers
Write-Host ""
Write-Host "============================================" -ForegroundColor Red
Write-Host "EMERGENCY: Dropping All Problematic Triggers" -ForegroundColor Red
Write-Host "============================================" -ForegroundColor Red
Write-Host ""

# Load environment variables
$envFile = ".env"
if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            $key = $matches[1].Trim()
            $value = $matches[2].Trim()
            [System.Environment]::SetEnvironmentVariable($key, $value, "Process")
        }
    }
}

# Database connection details
$DB_HOST = $env:DB_HOST
$DB_PORT = $env:DB_PORT
$DB_USER = $env:DB_USER
$DB_PASSWORD = $env:DB_PASSWORD
$DB_NAME = $env:DB_NAME

if (-not $DB_HOST) { $DB_HOST = "localhost" }
if (-not $DB_PORT) { $DB_PORT = "5432" }
if (-not $DB_USER) { $DB_USER = "postgres" }
if (-not $DB_NAME) { $DB_NAME = "sistem_akuntansi" }

Write-Host "Database: ${DB_NAME}@${DB_HOST}:${DB_PORT}" -ForegroundColor Cyan
Write-Host ""

# Set PGPASSWORD environment variable
$env:PGPASSWORD = $DB_PASSWORD

# Run SQL script
Write-Host "Executing emergency trigger drop..." -ForegroundColor Yellow
$sqlFile = "backend\scripts\emergency_drop_all_triggers.sql"

if (Test-Path $sqlFile) {
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $sqlFile
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "✅ All triggers dropped successfully!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Next steps:" -ForegroundColor Yellow
        Write-Host "  1. Restart backend (it will run migration automatically)" -ForegroundColor White
        Write-Host "  2. Test invoice creation again" -ForegroundColor White
        Write-Host ""
    } else {
        Write-Host ""
        Write-Host "❌ Error dropping triggers" -ForegroundColor Red
        Write-Host "Error code: $LASTEXITCODE" -ForegroundColor Red
        Write-Host ""
    }
} else {
    Write-Host "❌ SQL file not found: $sqlFile" -ForegroundColor Red
}

# Clear password from environment
Remove-Item Env:\PGPASSWORD -ErrorAction SilentlyContinue

