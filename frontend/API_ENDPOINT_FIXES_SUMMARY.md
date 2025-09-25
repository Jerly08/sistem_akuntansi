# API Endpoint Updates Summary

## Issue Description
The frontend application was making requests to API endpoints with incorrect URL prefixes, causing 404 errors. The issue was that frontend was requesting endpoints without proper `/api/v1` prefixes or using incorrect prefixes.

## Root Cause
- Frontend code was directly concatenating `API_BASE_URL` with incomplete endpoint paths
- API endpoints were not consistent with the Swagger documentation 
- Some endpoints used `/api/v1` prefix while others didn't, causing confusion

## Changes Made

### 1. Updated `src/config/api.ts`
**Purpose**: Centralized API endpoint configuration to match Swagger documentation

**Key Changes**:
- Updated all endpoint paths to match exact Swagger documentation
- Authentication endpoints: `/auth/*` (no `/api/v1` prefix)
- Dashboard endpoints: `/dashboard/*` (no `/api/v1` prefix)  
- Notification endpoints: `/api/v1/notifications/*` (with `/api/v1` prefix)
- Cash & Bank endpoints: `/api/cashbank/*` (no `/api/v1` prefix)
- Payment endpoints: `/api/payments/*` (no `/api/v1` prefix)
- Added new endpoint constants for all Swagger endpoints

**Before**:
```typescript
LOGIN: '/api/v1/auth/login',
DASHBOARD_ANALYTICS: '/api/v1/dashboard/analytics',
CASHBANK_ACCOUNTS: '/api/v1/cashbank/accounts',
```

**After**:
```typescript
LOGIN: '/auth/login',
DASHBOARD_ANALYTICS: '/dashboard/analytics', 
CASHBANK_ACCOUNTS: '/api/cashbank/accounts',
```

### 2. Updated `src/contexts/AuthContext.tsx`
**Purpose**: Use correct authentication endpoints from API configuration

**Changes**:
- Added import for `API_ENDPOINTS` from config
- Updated all auth endpoints to use constants:
  - `VALIDATE_TOKEN`: `/auth/validate-token`
  - `LOGIN`: `/auth/login`
  - `REGISTER`: `/auth/register`
  - `REFRESH`: `/auth/refresh`

### 3. Updated `src/services/api.ts`
**Purpose**: Use correct refresh endpoint in token refresh interceptor

**Changes**:
- Added import for `API_ENDPOINTS` from config
- Updated refresh endpoint to use `API_ENDPOINTS.REFRESH`

### 4. Updated `src/hooks/useDashboardAnalytics.ts`
**Purpose**: Use correct dashboard analytics endpoint

**Changes**:
- Added import for `API_ENDPOINTS` from config
- Updated endpoint to use `API_ENDPOINTS.DASHBOARD_ANALYTICS`

### 5. Updated `src/hooks/usePermissions.ts`
**Purpose**: Use correct permissions endpoint

**Changes**:
- Added import for `API_ENDPOINTS` from config
- Updated endpoint to use `API_ENDPOINTS.PERMISSIONS_ME`

### 6. Updated `src/components/notification/UnifiedNotifications.tsx`
**Purpose**: Use correct notification endpoints

**Changes**:
- Updated import to use `API_ENDPOINTS` from config
- Updated all notification endpoints:
  - `NOTIFICATIONS_APPROVALS`
  - `NOTIFICATIONS_BY_TYPE()` function
  - `DASHBOARD_STOCK_ALERTS`
  - `NOTIFICATIONS_MARK_READ()` function

### 7. Updated `src/services/salesService.ts`
**Purpose**: Fix sales API endpoints causing 404 errors

**Changes**:
- Added import for `API_ENDPOINTS` from config
- Updated all sales endpoints to use proper `/api/v1/sales` prefix:
  - `getSales()`: now uses `API_ENDPOINTS.SALES`
  - `getSale()`: now uses `API_ENDPOINTS.SALES_BY_ID()`
  - `createSale()`: now uses `API_ENDPOINTS.SALES`
  - `updateSale()`: now uses `API_ENDPOINTS.SALES_BY_ID()`
  - `deleteSale()`: now uses `API_ENDPOINTS.SALES_BY_ID()`
  - All payment, return, and analytics endpoints updated
  - Fixed the main issue: `/sales?page=1&limit=10` → `/api/v1/sales?page=1&limit=10`

## API Endpoint Mapping (Based on Swagger)

### Authentication Endpoints (No /api/v1 prefix)
- `POST /auth/login` - User login
- `POST /auth/register` - User registration  
- `POST /auth/refresh` - Refresh access token
- `GET /auth/validate-token` - Validate JWT token
- `GET /profile` - Get user profile

### Dashboard Endpoints (No /api/v1 prefix)  
- `GET /dashboard/analytics` - Get analytics data
- `GET /dashboard/finance` - Get finance dashboard data

### Notification Endpoints (With /api/v1 prefix)
- `GET /api/v1/notifications/approvals` - Get approval notifications
- `GET /api/v1/notifications/type/{type}` - Get notifications by type
- `PUT /api/v1/notifications/{id}/read` - Mark notification as read

### Permissions Endpoints (With /api/v1 prefix)
- `GET /api/v1/permissions/me` - Get user permissions

### Sales Endpoints (With /api/v1 prefix)
- `GET /api/v1/sales` - Get sales list
- `GET /api/v1/sales/{id}` - Get sale by ID
- `POST /api/v1/sales` - Create sale
- `PUT /api/v1/sales/{id}` - Update sale
- `DELETE /api/v1/sales/{id}` - Delete sale
- `POST /api/v1/sales/{id}/confirm` - Confirm sale
- `POST /api/v1/sales/{id}/invoice` - Create invoice from sale
- `POST /api/v1/sales/{id}/cancel` - Cancel sale
- `GET /api/v1/sales/{id}/payments` - Get sale payments
- `POST /api/v1/sales/{id}/payments` - Create sale payment
- `POST /api/v1/sales/{id}/integrated-payment` - Create integrated payment
- `GET /api/v1/sales/summary` - Get sales summary
- `GET /api/v1/sales/analytics` - Get sales analytics
- `GET /api/v1/sales/receivables` - Get receivables report

### Cash & Bank Endpoints (No /api/v1 prefix)
- `GET /api/cashbank/accounts` - Get cash and bank accounts
- `GET /api/cashbank/balance-summary` - Get balance summary
- `POST /api/cashbank/deposit` - Process deposit
- `POST /api/cashbank/withdrawal` - Process withdrawal

### Payment Endpoints (No /api/v1 prefix)  
- `GET /api/payments` - Get payments list
- `GET /api/payments/analytics` - Get payment analytics
- `GET /api/payments/summary` - Get payment summary

## Environment Configuration
- `.env.local` properly configured with:
  ```env
  NEXT_PUBLIC_API_URL=http://localhost:8080
  ```

## Expected Results
- No more 404 errors for:
  - `/permissions/me` → Now uses `/api/v1/permissions/me`
  - `/dashboard/analytics` → Now uses `/dashboard/analytics` 
  - `/notifications/approvals` → Now uses `/api/v1/notifications/approvals`
  - `/dashboard/stock-alerts` → Now uses `/api/v1/dashboard/stock-alerts`
  - Authentication endpoints now use `/auth/*` instead of `/api/v1/auth/*`

## Testing
- Frontend should successfully connect to backend API
- Authentication flow should work properly
- Dashboard data should load without 404 errors
- Notification system should work correctly
- All API calls should use proper endpoints matching Swagger documentation

## Notes
- All changes maintain backward compatibility where possible
- Centralized configuration makes future updates easier
- Endpoints now exactly match the Swagger documentation at `http://localhost:8080/swagger/index.html`