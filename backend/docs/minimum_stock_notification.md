# Minimum Stock Notification System

## Overview
Implementasi sistem notifikasi otomatis ketika stock produk mencapai level minimum.

## Fitur Utama

### 1. **Automatic Stock Monitoring**
- Monitoring otomatis saat stock berubah
- Notifikasi real-time untuk admin dan inventory manager
- Auto-resolve saat stock kembali normal

### 2. **Dashboard Integration**
- Banner alert di dashboard untuk stock minimum
- Badge count di navbar (ikon lonceng)
- Quick stats widget untuk monitoring

### 3. **Deduplication Logic**
- Mencegah notifikasi duplikat untuk produk yang sama
- Update existing notification jika sudah ada
- Track status alert (ACTIVE, RESOLVED, DISMISSED)

## API Endpoints

### Dashboard Endpoints
```
GET /api/v1/dashboard/summary
- Get comprehensive dashboard data including stock alerts

GET /api/v1/dashboard/stock-alerts
- Get stock alerts for banner display

POST /api/v1/dashboard/stock-alerts/:id/dismiss
- Dismiss a specific stock alert

GET /api/v1/dashboard/quick-stats
- Get quick statistics including stock counts
```

### Product Unit Endpoints
```
GET /api/v1/product-units
- Get all product units

GET /api/v1/product-units/:id
- Get specific unit

POST /api/v1/product-units
- Create new unit

PUT /api/v1/product-units/:id
- Update unit

DELETE /api/v1/product-units/:id
- Delete unit
```

## Database Tables

### stock_alerts
```sql
- id (PRIMARY KEY)
- product_id (FK to products)
- alert_type (LOW_STOCK, OUT_OF_STOCK)
- current_stock
- threshold_stock
- status (ACTIVE, RESOLVED, DISMISSED)
- last_alert_at
- created_at, updated_at, deleted_at
```

### notifications (updated)
- Added type: MIN_STOCK
- Priority: HIGH for minimum stock alerts

## Configuration

### Roles dengan akses notifikasi:
- `admin`: Full access
- `inventory_manager`: Full access
- `director`: View only

### Stock Monitoring Triggers:
1. **Product Stock Adjustment** (`/api/v1/products/adjust-stock`)
2. **Stock Opname** (`/api/v1/products/opname`)
3. **Purchase Order Completion**
4. **Sales Order Processing**

## Frontend Integration Guide

### 1. Dashboard Banner Component
```javascript
// Fetch stock alerts for banner
const fetchStockAlerts = async () => {
  const response = await api.get('/api/v1/dashboard/stock-alerts');
  if (response.data.show_banner) {
    showStockAlertBanner(response.data.alerts);
  }
};

// Auto-refresh every 30 seconds
setInterval(fetchStockAlerts, 30000);
```

### 2. Navbar Notification Badge
```javascript
// Get notification count
const getNotificationCount = async () => {
  const response = await api.get('/api/v1/dashboard/summary');
  updateBadgeCount(response.data.min_stock_alerts_count);
};
```

### 3. Alert Banner Styling
```css
.stock-alert-banner {
  background: linear-gradient(135deg, #ff6b6b, #ff8e53);
  color: white;
  padding: 12px 20px;
  display: flex;
  align-items: center;
  animation: slideDown 0.3s ease;
}

.stock-alert-banner.critical {
  background: linear-gradient(135deg, #d32f2f, #ff5252);
}

.stock-alert-banner.warning {
  background: linear-gradient(135deg, #f57c00, #ffb74d);
}
```

## Workflow Diagram

```
Product Stock Change
        ↓
Check if stock ≤ minStock
        ↓
    [Yes]────→ Check existing alert
                      ↓
                [Not exists]────→ Create StockAlert record
                      ↓               ↓
                [Exists]────→    Create/Update Notification
                                      ↓
                                Send to Admin/Inventory Manager
                                      ↓
                                Display in Dashboard & Navbar

Stock Restored (stock > minStock)
        ↓
Resolve StockAlert
        ↓
Mark Notifications as Read
        ↓
Remove from Dashboard Banner
```

## Testing Scenarios

### 1. Test Minimum Stock Alert
```bash
# Adjust stock to below minimum
POST /api/v1/products/adjust-stock
{
  "product_id": 1,
  "quantity": 10,
  "type": "OUT",
  "notes": "Test minimum stock"
}

# Verify notification created
GET /api/v1/notifications?type=MIN_STOCK
```

### 2. Test Auto-Resolve
```bash
# Restore stock above minimum
POST /api/v1/products/adjust-stock
{
  "product_id": 1,
  "quantity": 50,
  "type": "IN",
  "notes": "Restock"
}

# Verify alert resolved
GET /api/v1/dashboard/stock-alerts
```

## Scheduled Jobs (Cron)

Add to cron for periodic checking:
```bash
# Check stock levels every hour
0 * * * * curl -X POST http://localhost:8080/api/v1/internal/stock-check
```

## Performance Considerations

1. **Async Processing**: Stock checks run in goroutines
2. **Database Indexes**: Ensure indexes on:
   - `products.stock`, `products.min_stock`
   - `stock_alerts.status`, `stock_alerts.product_id`
   - `notifications.user_id`, `notifications.type`, `notifications.is_read`
3. **Caching**: Consider caching active alerts for dashboard

## Security Notes

- Only authorized roles can view/dismiss alerts
- Notifications are user-specific
- Audit log for all stock adjustments

## Future Enhancements

1. **Email/SMS Notifications**: Send alerts via email/SMS
2. **Predictive Analytics**: Predict when stock will reach minimum
3. **Auto-Purchase Order**: Generate PO when reorder level reached
4. **Custom Alert Rules**: Per-product or per-category rules
5. **Alert History Report**: Historical data and trends
6. **Mobile Push Notifications**: Real-time mobile alerts
