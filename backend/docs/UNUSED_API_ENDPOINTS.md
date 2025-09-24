# Unused API Endpoints Analysis

## Summary
Based on comprehensive analysis of both backend API definitions and frontend usage patterns, the following endpoints are identified as unused or underused.

## Unused/Underused Endpoints

### 1. Journal Entry Auto-generation
- `POST /journal-entries/auto-generate/purchase`
- `POST /journal-entries/auto-generate/sale`
**Status**: Not found in frontend usage

### 2. Journal Entry Operations  
- `POST /journal-entries/{id}/post`
- `POST /journal-entries/{id}/reverse`
- `GET /journal-entries/summary`
**Status**: Not found in frontend usage

### 3. Admin Operations
- `GET /api/admin/check-cashbank-gl-links`
- `POST /api/admin/fix-cashbank-gl-links`
**Status**: Not found in frontend usage

### 4. Balance Monitoring (All endpoints)
- `GET /api/monitoring/balance-health`
- `GET /api/monitoring/balance-sync`  
- `GET /api/monitoring/discrepancies`
- `POST /api/monitoring/fix-discrepancies`
- `GET /api/monitoring/sync-status`
**Status**: Not found in frontend usage

### 5. Payment Debug/Analytics
- `GET /api/payments/debug/recent`
- `GET /api/payments/analytics`
- `GET /api/payments/export/excel`
- `GET /api/payments/report/pdf`
- `GET /api/payments/{id}/pdf`
**Status**: Limited or no frontend usage

### 6. Enhanced Reports (All endpoints)
- `GET /api/reports/enhanced/financial-metrics`
- `GET /api/reports/enhanced/profit-loss`  
- `GET /api/reports/enhanced/profit-loss-comparison`
**Status**: Frontend uses `/reports/*` not `/reports/enhanced/*`

### 7. Security Dashboard (All endpoints)
- `GET /api/v1/admin/security/alerts`
- `PUT /api/v1/admin/security/alerts/{id}/acknowledge`
- `POST /api/v1/admin/security/cleanup`
- `GET /api/v1/admin/security/config`
- `GET /api/v1/admin/security/incidents`
- `GET /api/v1/admin/security/incidents/{id}`
- `PUT /api/v1/admin/security/incidents/{id}/resolve`
- `GET /api/v1/admin/security/ip-whitelist`
- `POST /api/v1/admin/security/ip-whitelist`
- `GET /api/v1/admin/security/metrics`
**Status**: Not found in frontend usage

### 8. Account Journal Entries
- `GET /accounts/{account_id}/journal-entries`
**Status**: Not found in frontend usage

## Recommendations

1. **Remove from Swagger**: All endpoints listed above
2. **Keep Backend Code**: Some endpoints may be used by external integrations or admin tools
3. **Frontend Integration**: Consider if any of these features should be added to frontend
4. **API Documentation**: Update Swagger to reflect only actively used endpoints

## Frontend-Backend Endpoint Mapping

### Used Endpoints (Keep in Swagger):
- `/auth/*` ✅
- `/accounts/*` (except account journal entries) ✅
- `/products/*` ✅
- `/purchases/*` ✅
- `/sales/*` ✅
- `/contacts/*` ✅
- `/cashbank/accounts/*` ✅
- `/cashbank/payment-accounts` ✅
- `/cashbank/balance-summary` ✅
- `/cashbank/transfer` ✅
- `/cashbank/deposit` ✅
- `/cashbank/withdrawal` ✅
- `/reports/*` (standard reports) ✅
- `/journal-entries` (basic CRUD) ✅
- `/journal-drilldown/*` ✅
- `/dashboard/*` ✅
- `/notifications/*` ✅

### Unused Endpoints (Remove from Swagger):
All endpoints listed in the "Unused/Underused Endpoints" section above.