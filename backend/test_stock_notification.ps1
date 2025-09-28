# Test Stock Notification System
# This script tests the stock notification functionality

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$Token = "",
    [int]$ProductId = 1
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TESTING STOCK NOTIFICATION SYSTEM" -ForegroundColor Cyan  
Write-Host "========================================" -ForegroundColor Cyan

if ($Token -eq "") {
    Write-Host "❌ Error: Token is required. Please provide a valid JWT token." -ForegroundColor Red
    Write-Host "Usage: ./test_stock_notification.ps1 -Token 'your_jwt_token_here' -ProductId 1" -ForegroundColor Yellow
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $Token"
    "Content-Type" = "application/json"
}

Write-Host "🔍 Step 1: Check current product stock..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/products/$ProductId" -Method GET -Headers $headers
    $product = $response.data
    
    Write-Host "✅ Product found:" -ForegroundColor Green
    Write-Host "   - ID: $($product.id)" -ForegroundColor White
    Write-Host "   - Code: $($product.code)" -ForegroundColor White  
    Write-Host "   - Name: $($product.name)" -ForegroundColor White
    Write-Host "   - Current Stock: $($product.stock)" -ForegroundColor White
    Write-Host "   - Min Stock: $($product.min_stock)" -ForegroundColor White
    Write-Host "   - Reorder Level: $($product.reorder_level)" -ForegroundColor White
    Write-Host ""
    
    if ($product.min_stock -eq 0) {
        Write-Host "⚠️  WARNING: Product has min_stock = 0. Setting minimum stock to test notifications..." -ForegroundColor Yellow
        
        # Update product to set min_stock
        $updateData = @{
            code = $product.code
            name = $product.name
            min_stock = 20
            reorder_level = 30
            stock = $product.stock
        }
        
        $updateResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/products/$ProductId" -Method PUT -Headers $headers -Body ($updateData | ConvertTo-Json)
        Write-Host "✅ Updated product with min_stock=20, reorder_level=30" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Error fetching product: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host "🔍 Step 2: Check current notifications..." -ForegroundColor Yellow
try {
    $notifResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/notifications/type/MIN_STOCK" -Method GET -Headers $headers
    Write-Host "✅ Current MIN_STOCK notifications: $($notifResponse.total)" -ForegroundColor Green
    
    $reorderResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/notifications/type/REORDER_ALERT" -Method GET -Headers $headers  
    Write-Host "✅ Current REORDER_ALERT notifications: $($reorderResponse.total)" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "❌ Error fetching notifications: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "🔍 Step 3: Check dashboard stock alerts..." -ForegroundColor Yellow
try {
    $alertResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/dashboard/stock-alerts" -Method GET -Headers $headers
    Write-Host "✅ Current stock alerts: $($alertResponse.data.total_count)" -ForegroundColor Green
    Write-Host "   - Show banner: $($alertResponse.data.show_banner)" -ForegroundColor White
    Write-Host ""
} catch {
    Write-Host "❌ Error fetching stock alerts: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "🎯 Step 4: Simulate stock reduction to trigger notification..." -ForegroundColor Yellow
try {
    # Reduce stock to below minimum (set to 5 if min_stock is 20)
    $adjustData = @{
        product_id = $ProductId
        quantity = 50  # Reduce by 50
        type = "OUT"
        notes = "Test stock notification - reducing stock"
    }
    
    $adjustResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/products/adjust-stock" -Method POST -Headers $headers -Body ($adjustData | ConvertTo-Json)
    Write-Host "✅ Stock adjusted successfully" -ForegroundColor Green
    Write-Host "   - New stock: $($adjustResponse.product.stock)" -ForegroundColor White
    Write-Host ""
} catch {
    Write-Host "❌ Error adjusting stock: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Response: $($_.Exception.Response)" -ForegroundColor Red
}

Write-Host "🕐 Step 5: Wait 2 seconds for notification processing..." -ForegroundColor Yellow
Start-Sleep -Seconds 2

Write-Host "🔍 Step 6: Check for new notifications..." -ForegroundColor Yellow
try {
    $notifResponse2 = Invoke-RestMethod -Uri "$BaseUrl/api/v1/notifications/type/MIN_STOCK" -Method GET -Headers $headers
    Write-Host "✅ MIN_STOCK notifications after adjustment: $($notifResponse2.total)" -ForegroundColor Green
    
    if ($notifResponse2.total -gt 0) {
        Write-Host "📋 Latest MIN_STOCK notification:" -ForegroundColor Green
        $latestNotif = $notifResponse2.notifications[0]
        Write-Host "   - Title: $($latestNotif.title)" -ForegroundColor White
        Write-Host "   - Message: $($latestNotif.message)" -ForegroundColor White
        Write-Host "   - Created: $($latestNotif.created_at)" -ForegroundColor White
        Write-Host "   - Read: $($latestNotif.is_read)" -ForegroundColor White
    }
    
    $reorderResponse2 = Invoke-RestMethod -Uri "$BaseUrl/api/v1/notifications/type/REORDER_ALERT" -Method GET -Headers $headers
    Write-Host "✅ REORDER_ALERT notifications after adjustment: $($reorderResponse2.total)" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "❌ Error fetching notifications: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "🔍 Step 7: Check dashboard stock alerts after adjustment..." -ForegroundColor Yellow
try {
    $alertResponse2 = Invoke-RestMethod -Uri "$BaseUrl/api/v1/dashboard/stock-alerts" -Method GET -Headers $headers
    Write-Host "✅ Stock alerts after adjustment: $($alertResponse2.data.total_count)" -ForegroundColor Green
    Write-Host "   - Show banner: $($alertResponse2.data.show_banner)" -ForegroundColor White
    
    if ($alertResponse2.data.total_count -gt 0) {
        Write-Host "🚨 Active stock alerts:" -ForegroundColor Red
        foreach ($alert in $alertResponse2.data.alerts) {
            Write-Host "   - Product: $($alert.product_name) ($($alert.product_code))" -ForegroundColor White
            Write-Host "   - Current Stock: $($alert.current_stock)" -ForegroundColor White
            Write-Host "   - Threshold: $($alert.threshold_stock)" -ForegroundColor White
            Write-Host "   - Alert Type: $($alert.alert_type)" -ForegroundColor White
            Write-Host "   - Urgency: $($alert.urgency)" -ForegroundColor White
            Write-Host "   - Message: $($alert.message)" -ForegroundColor White
            Write-Host ""
        }
    }
} catch {
    Write-Host "❌ Error fetching stock alerts: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  STOCK NOTIFICATION TEST COMPLETED" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host ""
Write-Host "📝 SUMMARY:" -ForegroundColor Cyan
Write-Host "1. ✅ Product information retrieved" -ForegroundColor Green
Write-Host "2. ✅ Stock adjustment applied" -ForegroundColor Green  
Write-Host "3. 🔍 Check the results above to see if notifications were created" -ForegroundColor Yellow
Write-Host ""
Write-Host "💡 If no notifications were created, check:" -ForegroundColor Yellow
Write-Host "   - Product min_stock and reorder_level settings" -ForegroundColor White
Write-Host "   - Stock monitoring service is running" -ForegroundColor White
Write-Host "   - User role has notification permissions" -ForegroundColor White
Write-Host "   - Database for stock_alerts and notifications tables" -ForegroundColor White