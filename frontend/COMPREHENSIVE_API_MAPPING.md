# COMPREHENSIVE API ENDPOINTS MAPPING

## Overview
Dokumentasi lengkap yang memetakan semua API endpoints antara backend dan frontend untuk memastikan tidak ada lagi 404 errors di masa depan.

## Backend Routes Analysis Summary

### Total Backend Endpoints: 175+

#### Authentication & Authorization (No /api/v1 prefix)
- `POST /auth/login` - User login  
- `POST /auth/register` - User registration
- `POST /auth/refresh` - Refresh access token
- `GET /auth/validate-token` - Validate JWT token
- `GET /profile` - Get user profile

#### Core API Routes (With /api/v1 prefix)

**Users Management:**
- `GET /api/v1/users` - List all users
- `GET /api/v1/users/:id` - Get user by ID
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

**Permissions Management:**
- `GET /api/v1/permissions/users` - List user permissions
- `GET /api/v1/permissions/users/:userId` - Get user permissions
- `PUT /api/v1/permissions/users/:userId` - Update user permissions  
- `POST /api/v1/permissions/users/:userId/reset` - Reset to default permissions
- `GET /api/v1/permissions/me` - Get current user permissions
- `GET /api/v1/permissions/check` - Check specific permission

**Dashboard:**
- `GET /api/v1/dashboard/analytics` - Get analytics data
- `GET /api/v1/dashboard/finance` - Get finance dashboard data
- `GET /api/v1/dashboard/stock-alerts` - Get stock alerts
- `POST /api/v1/dashboard/stock-alerts/:id/dismiss` - Dismiss stock alert

**Products & Inventory:**
- `GET /api/v1/products` - List products
- `GET /api/v1/products/:id` - Get product by ID
- `POST /api/v1/products` - Create product
- `PUT /api/v1/products/:id` - Update product
- `DELETE /api/v1/products/:id` - Delete product
- `POST /api/v1/products/adjust-stock` - Adjust stock
- `POST /api/v1/products/opname` - Stock opname
- `POST /api/v1/products/upload-image` - Upload product image

**Categories:**
- `GET /api/v1/categories` - List categories
- `GET /api/v1/categories/tree` - Get category tree
- `GET /api/v1/categories/:id` - Get category by ID
- `GET /api/v1/categories/:id/products` - Get category products
- `POST /api/v1/categories` - Create category
- `PUT /api/v1/categories/:id` - Update category
- `DELETE /api/v1/categories/:id` - Delete category

**Product Units:**
- `GET /api/v1/product-units` - List product units
- `GET /api/v1/product-units/:id` - Get product unit by ID
- `POST /api/v1/product-units` - Create product unit
- `PUT /api/v1/product-units/:id` - Update product unit  
- `DELETE /api/v1/product-units/:id` - Delete product unit

**Warehouse Locations:**
- `GET /api/v1/warehouse-locations` - List warehouse locations
- `GET /api/v1/warehouse-locations/:id` - Get warehouse location by ID
- `POST /api/v1/warehouse-locations` - Create warehouse location
- `PUT /api/v1/warehouse-locations/:id` - Update warehouse location
- `DELETE /api/v1/warehouse-locations/:id` - Delete warehouse location

**Inventory Management:**
- `GET /api/v1/inventory/movements` - Get inventory movements
- `GET /api/v1/inventory/low-stock` - Get low stock products
- `GET /api/v1/inventory/valuation` - Get stock valuation
- `GET /api/v1/inventory/report` - Get stock report
- `POST /api/v1/inventory/bulk-price-update` - Bulk price update

**Accounts (Chart of Accounts):**
- `GET /api/v1/accounts` - List accounts
- `GET /api/v1/accounts/hierarchy` - Get account hierarchy
- `GET /api/v1/accounts/balance-summary` - Get balance summary
- `GET /api/v1/accounts/validate-code` - Validate account code
- `POST /api/v1/accounts/fix-header-status` - Fix header status
- `GET /api/v1/accounts/:code` - Get account by code
- `POST /api/v1/accounts` - Create account
- `PUT /api/v1/accounts/:code` - Update account
- `DELETE /api/v1/accounts/:code` - Delete account
- `DELETE /api/v1/accounts/admin/:code` - Admin delete account
- `POST /api/v1/accounts/import` - Import accounts
- `GET /api/v1/accounts/export/pdf` - Export accounts PDF
- `GET /api/v1/accounts/export/excel` - Export accounts Excel

**Public Account Catalog (No auth required):**
- `GET /api/v1/accounts/catalog` - Get account catalog
- `GET /api/v1/accounts/credit` - Get credit accounts

**Contacts:**
- `GET /api/v1/contacts` - List contacts
- `GET /api/v1/contacts/:id` - Get contact by ID
- `POST /api/v1/contacts` - Create contact
- `PUT /api/v1/contacts/:id` - Update contact
- `DELETE /api/v1/contacts/:id` - Delete contact
- `GET /api/v1/contacts/type/:type` - Get contacts by type
- `GET /api/v1/contacts/search` - Search contacts
- `POST /api/v1/contacts/import` - Import contacts
- `GET /api/v1/contacts/export` - Export contacts
- `POST /api/v1/contacts/:id/addresses` - Add contact address
- `PUT /api/v1/contacts/:id/addresses/:address_id` - Update contact address
- `DELETE /api/v1/contacts/:id/addresses/:address_id` - Delete contact address

**Notifications:**
- `GET /api/v1/notifications` - List notifications
- `GET /api/v1/notifications/unread-count` - Get unread count
- `PUT /api/v1/notifications/:id/read` - Mark notification as read
- `PUT /api/v1/notifications/read-all` - Mark all notifications as read
- `GET /api/v1/notifications/type/:type` - Get notifications by type
- `GET /api/v1/notifications/approvals` - Get approval notifications

**Sales:**
- `GET /api/v1/sales` - List sales
- `GET /api/v1/sales/:id` - Get sale by ID
- `POST /api/v1/sales` - Create sale
- `PUT /api/v1/sales/:id` - Update sale  
- `DELETE /api/v1/sales/:id` - Delete sale
- `POST /api/v1/sales/:id/confirm` - Confirm sale
- `POST /api/v1/sales/:id/invoice` - Create invoice from sale
- `POST /api/v1/sales/:id/cancel` - Cancel sale
- `GET /api/v1/sales/:id/payments` - Get sale payments
- `POST /api/v1/sales/:id/payments` - Create sale payment
- `GET /api/v1/sales/:id/for-payment` - Get sale for payment
- `POST /api/v1/sales/:id/integrated-payment` - Create integrated payment
- `POST /api/v1/sales/:id/returns` - Create sale return
- `GET /api/v1/sales/returns` - List sale returns
- `GET /api/v1/sales/summary` - Get sales summary
- `GET /api/v1/sales/analytics` - Get sales analytics
- `GET /api/v1/sales/receivables` - Get receivables report
- `GET /api/v1/sales/:id/invoice/pdf` - Export invoice PDF
- `GET /api/v1/sales/report/pdf` - Export sales report PDF
- `GET /api/v1/sales/customer/:customer_id` - Get customer sales
- `GET /api/v1/sales/customer/:customer_id/invoices` - Get customer invoices

**Purchases:**
- `GET /api/v1/purchases` - List purchases
- `GET /api/v1/purchases/approval-stats` - Get approval statistics
- `GET /api/v1/purchases/:id` - Get purchase by ID
- `POST /api/v1/purchases` - Create purchase
- `PUT /api/v1/purchases/:id` - Update purchase
- `DELETE /api/v1/purchases/:id` - Delete purchase
- `POST /api/v1/purchases/:id/submit-approval` - Submit for approval
- `POST /api/v1/purchases/:id/approve` - Approve purchase
- `POST /api/v1/purchases/:id/reject` - Reject purchase
- `GET /api/v1/purchases/:id/approval-history` - Get approval history
- `GET /api/v1/purchases/pending-approval` - Get pending approvals
- `POST /api/v1/purchases/:id/documents` - Upload document
- `GET /api/v1/purchases/:id/documents` - Get purchase documents
- `DELETE /api/v1/purchases/documents/:document_id` - Delete document
- `POST /api/v1/purchases/receipts` - Create receipt
- `GET /api/v1/purchases/:id/receipts` - Get purchase receipts
- `GET /api/v1/purchases/receipts/:receipt_id/pdf` - Get receipt PDF
- `GET /api/v1/purchases/:id/receipts/pdf` - Get all receipts PDF
- `GET /api/v1/purchases/summary` - Get purchases summary
- `GET /api/v1/purchases/pending-approvals` - Get pending approvals
- `GET /api/v1/purchases/dashboard` - Get purchase dashboard
- `GET /api/v1/purchases/vendor/:vendor_id/summary` - Get vendor summary
- `GET /api/v1/purchases/:id/payments` - Get purchase payments
- `POST /api/v1/purchases/:id/payments` - Create purchase payment
- `GET /api/v1/purchases/:id/for-payment` - Get purchase for payment  
- `POST /api/v1/purchases/:id/integrated-payment` - Create integrated payment
- `GET /api/v1/purchases/:id/matching` - Get purchase matching
- `POST /api/v1/purchases/:id/validate-matching` - Validate matching
- `GET /api/v1/purchases/:id/journal-entries` - Get journal entries

**Assets:**
- `GET /api/v1/assets` - List assets
- `GET /api/v1/assets/:id` - Get asset by ID
- `POST /api/v1/assets` - Create asset
- `PUT /api/v1/assets/:id` - Update asset
- `DELETE /api/v1/assets/:id` - Delete asset
- `POST /api/v1/assets/upload-image` - Upload asset image
- `GET /api/v1/assets/categories` - Get asset categories
- `POST /api/v1/assets/categories` - Create asset category
- `GET /api/v1/assets/summary` - Get assets summary
- `GET /api/v1/assets/depreciation-report` - Get depreciation report
- `GET /api/v1/assets/:id/depreciation-schedule` - Get depreciation schedule
- `GET /api/v1/assets/:id/calculate-depreciation` - Calculate depreciation

#### Payment & Cash Bank Routes (No /api/v1 prefix)

**Payments:**
- `GET /api/payments` - List payments (deprecated)
- `DELETE /api/payments/:id` - Delete payment
- `POST /api/payments/:id/cancel` - Cancel payment
- `GET /api/payments/analytics` - Get payment analytics
- `POST /api/payments/sales` - Create sales payment
- `GET /api/payments/sales/unpaid-invoices/:customer_id` - Get unpaid invoices
- `GET /api/payments/report/pdf` - Export payment report PDF
- `GET /api/payments/export/excel` - Export payment report Excel
- `GET /api/payments/:id/pdf` - Export payment detail PDF

**Cash Bank:**
- `GET /api/cashbank/accounts` - Get cash bank accounts
- `GET /api/cashbank/payment-accounts` - Get payment accounts
- `GET /api/cashbank/revenue-accounts` - Get revenue accounts
- `GET /api/cashbank/deposit-source-accounts` - Get deposit source accounts
- `GET /api/cashbank/accounts/:id` - Get account by ID
- `POST /api/cashbank/accounts` - Create cash bank account
- `PUT /api/cashbank/accounts/:id` - Update cash bank account  
- `POST /api/cashbank/transfer` - Process transfer
- `POST /api/cashbank/deposit` - Process deposit
- `POST /api/cashbank/withdrawal` - Process withdrawal
- `GET /api/cashbank/accounts/:id/transactions` - Get account transactions
- `GET /api/cashbank/balance-summary` - Get balance summary

**Cash Bank SSOT (With /api/v1 prefix):**
- `GET /api/v1/cash-bank/accounts` - List SSOT cash bank accounts
- `GET /api/v1/cash-bank/accounts/:id` - Get SSOT account by ID
- `POST /api/v1/cash-bank/accounts` - Create SSOT cash bank account
- `PUT /api/v1/cash-bank/accounts/:id` - Update SSOT cash bank account
- `DELETE /api/v1/cash-bank/accounts/:id` - Delete SSOT cash bank account
- `GET /api/v1/cash-bank/accounts/:id/transactions` - Get SSOT account transactions
- `POST /api/v1/cash-bank/accounts/:id/reconcile` - Reconcile SSOT account
- `POST /api/v1/cash-bank/transactions/deposit` - Process SSOT deposit
- `POST /api/v1/cash-bank/transactions/withdrawal` - Process SSOT withdrawal
- `POST /api/v1/cash-bank/transactions/transfer` - Process SSOT transfer
- `GET /api/v1/cash-bank/reports/balance-summary` - Get SSOT balance summary
- `GET /api/v1/cash-bank/reports/payment-accounts` - Get SSOT payment accounts
- `GET /api/v1/cash-bank/ssot/journals` - Get SSOT journal entries
- `POST /api/v1/cash-bank/ssot/validate-integrity` - Validate SSOT integrity

#### Journal & Reports

**Journals (With /api/v1 prefix):**
- `POST /api/v1/journals` - Create journal entry
- `GET /api/v1/journals` - List journal entries
- `GET /api/v1/journals/:id` - Get journal entry by ID
- `GET /api/v1/journals/account-balances` - Get account balances
- `POST /api/v1/journals/account-balances/refresh` - Refresh account balances
- `GET /api/v1/journals/summary` - Get journal summary

**Journal Drilldown (With /api/v1 prefix):**
- `POST /api/v1/journal-drilldown` - Journal entry drill-down
- `GET /api/v1/journal-drilldown/entries` - Get journal entries
- `GET /api/v1/journal-drilldown/entries/:id` - Get journal entry detail
- `GET /api/v1/journal-drilldown/accounts` - Get active accounts for period

**SSOT Reports (With /api/v1 prefix):**
- `GET /api/v1/ssot-reports/general-ledger` - Get general ledger
- `GET /api/v1/ssot-reports/integrated` - Get integrated reports
- `GET /api/v1/ssot-reports/journal-analysis` - Get journal analysis
- `GET /api/v1/ssot-reports/purchase-report` - Get purchase report
- `GET /api/v1/ssot-reports/purchase-report/validate` - Validate purchase report
- `GET /api/v1/ssot-reports/purchase-summary` - Get purchase summary
- `POST /api/v1/ssot-reports/refresh` - Refresh reports
- `GET /api/v1/ssot-reports/sales-summary` - Get sales summary
- `GET /api/v1/ssot-reports/status` - Get SSOT status
- `GET /api/v1/ssot-reports/trial-balance` - Get trial balance
- `GET /api/v1/ssot-reports/vendor-analysis` - Get vendor analysis

**SSOT Balance Sheet & Cash Flow (No /api/v1 prefix):**
- `GET /reports/ssot-profit-loss` - Generate SSOT profit & loss
- `GET /reports/ssot/balance-sheet` - Generate SSOT balance sheet
- `GET /reports/ssot/balance-sheet/account-details` - Get balance sheet details
- `GET /reports/ssot/cash-flow` - Generate SSOT cash flow

**Optimized Reports (With /api/v1 prefix):**
- `GET /api/v1/reports/optimized/balance-sheet` - Optimized balance sheet
- `GET /api/v1/reports/optimized/profit-loss` - Optimized profit & loss
- `GET /api/v1/reports/optimized/trial-balance` - Optimized trial balance
- `POST /api/v1/reports/optimized/refresh-balances` - Refresh materialized view

**Report Aliases (No prefix - compatibility routes):**
- `GET /ssot-reports/trial-balance` - Alias for SSOT trial balance
- `GET /ssot-reports/general-ledger` - Alias for SSOT general ledger  
- `GET /ssot-reports/journal-analysis` - Alias for SSOT journal analysis
- `GET /ssot-reports/purchase-report` - Alias for SSOT purchase report
- `GET /ssot-reports/info` - Info about alias routes

#### Monitoring & Admin (With /api/v1 prefix)

**System Monitoring:**
- `GET /api/v1/monitoring/status` - Get system status
- `GET /api/v1/monitoring/rate-limits` - Get rate limit status
- `GET /api/v1/monitoring/security-alerts` - Get security alerts
- `GET /api/v1/monitoring/audit-logs` - Get audit logs
- `GET /api/v1/monitoring/token-stats` - Get token statistics
- `GET /api/v1/monitoring/refresh-events` - Get refresh events
- `GET /api/v1/monitoring/users/:user_id/security-summary` - Get user security summary
- `GET /api/v1/monitoring/startup-status` - Get startup status
- `POST /api/v1/monitoring/fix-account-headers` - Fix account headers

**Balance Monitoring:**
- `GET /api/v1/monitoring/balance-sync` - Check balance sync
- `POST /api/v1/monitoring/fix-discrepancies` - Fix balance discrepancies
- `GET /api/v1/monitoring/balance-health` - Get balance health
- `GET /api/v1/monitoring/discrepancies` - Get balance discrepancies
- `GET /api/v1/monitoring/sync-status` - Get sync status

**API Usage Monitoring:**
- `GET /api/v1/monitoring/api-usage/stats` - Get API usage stats
- `GET /api/v1/monitoring/api-usage/top` - Get top endpoints
- `GET /api/v1/monitoring/api-usage/unused` - Get unused endpoints
- `GET /api/v1/monitoring/api-usage/analytics` - Get usage analytics
- `POST /api/v1/monitoring/api-usage/reset` - Reset usage stats

**Performance Monitoring:**
- `GET /api/v1/monitoring/performance/report` - Get performance report
- `GET /api/v1/monitoring/performance/metrics` - Get quick metrics
- `GET /api/v1/monitoring/performance/bottlenecks` - Get bottlenecks
- `GET /api/v1/monitoring/performance/recommendations` - Get recommendations
- `GET /api/v1/monitoring/performance/system` - Get system status
- `POST /api/v1/monitoring/performance/metrics/clear` - Clear metrics
- `GET /api/v1/monitoring/performance/test` - Test ultra-fast endpoint
- `GET /api/v1/monitoring/timeout/diagnostics` - Run timeout diagnostics
- `GET /api/v1/monitoring/timeout/health` - Get quick health check

**Security Dashboard:**
- `GET /api/v1/admin/security/incidents` - Get security incidents
- `GET /api/v1/admin/security/incidents/:id` - Get security incident details
- `PUT /api/v1/admin/security/incidents/:id/resolve` - Resolve security incident
- `GET /api/v1/admin/security/alerts` - Get system alerts
- `PUT /api/v1/admin/security/alerts/:id/acknowledge` - Acknowledge alert
- `GET /api/v1/admin/security/metrics` - Get security metrics
- `GET /api/v1/admin/security/ip-whitelist` - Get IP whitelist
- `POST /api/v1/admin/security/ip-whitelist` - Add IP to whitelist
- `GET /api/v1/admin/security/config` - Get security config
- `POST /api/v1/admin/security/cleanup` - Cleanup security logs

**Approval Workflows:**
- `GET /api/v1/approval-workflows` - Get approval workflows
- `POST /api/v1/approval-workflows` - Create approval workflow

#### Debug Routes (Development only, /api/v1/debug)
- `GET /api/v1/debug/auth/context` - Test JWT context
- `GET /api/v1/debug/auth/role` - Test role permission
- `GET /api/v1/debug/auth/test-cashbank-permission` - Test cashbank permission
- `GET /api/v1/debug/auth/test-payments-permission` - Test payments permission

#### Static Files & Documentation

**Static Files:**
- `GET /templates/*filepath` - Static templates
- `HEAD /templates/*filepath` - Head templates
- `GET /uploads/*filepath` - Static uploads
- `HEAD /uploads/*filepath` - Head uploads

**Documentation:**
- `GET /swagger/index.html` - Swagger documentation
- `GET /docs/index.html` - Alternative docs
- `GET /openapi/doc.json` - OpenAPI JSON
- `GET /api/v1/swagger/*any` - Swagger UI (v1)
- `GET /api/v1/docs/*any` - Docs UI (v1)

**Health Check:**
- `GET /api/v1/health` - Health check endpoint

## Frontend Service Files Updated

### ✅ COMPLETED

1. **salesService.ts** - ✅ FULLY UPDATED
   - All endpoints now use `API_ENDPOINTS` constants
   - Updated to match `/api/v1/sales/*` pattern

2. **purchaseService.ts** - ✅ FULLY UPDATED  
   - All endpoints now use `API_ENDPOINTS` constants
   - Updated to match `/api/v1/purchases/*` pattern

3. **productService.ts** - ✅ FULLY UPDATED
   - All endpoints now use `API_ENDPOINTS` constants  
   - Products, Categories, Product Units, Warehouse Locations, Inventory all updated
   - Updated to match `/api/v1/*` pattern

4. **useDashboardAnalytics.ts** - ✅ UPDATED
   - Updated to use `API_ENDPOINTS.DASHBOARD_ANALYTICS`

5. **usePermissions.ts** - ✅ UPDATED
   - Updated to use `API_ENDPOINTS.PERMISSIONS_ME`

6. **UnifiedNotifications.tsx** - ✅ UPDATED
   - All notification endpoints updated to use `API_ENDPOINTS`

7. **AuthContext.tsx** - ✅ UPDATED
   - All auth endpoints updated to use `API_ENDPOINTS`

8. **api.ts** - ✅ UPDATED
   - Token refresh endpoint updated

### 🔄 NEEDS UPDATE

#### Priority 1: Core Service Files (Still using hardcoded URLs)

1. **userService.ts** - ❌ NEEDS UPDATE
   - Currently: `/users`, `/profile`, `/auth/change-password`
   - Should use: `/api/v1/users/*` and `/api/v1/permissions/users/*`
   - Missing `/api/v1` prefix on user endpoints

2. **assetService.ts** - ❌ NEEDS UPDATE
   - Currently: `/assets`, `/assets/categories`, `/accounts`, `/cashbank/payment-accounts`
   - Should use: `/api/v1/assets/*` pattern
   - Missing `/api/v1` prefix consistently

3. **cashbankService.ts** - ❌ NEEDS UPDATE
   - Currently: `/cashbank/*` (CORRECT for this service)
   - But needs to use `API_ENDPOINTS` constants instead of hardcoded strings
   - Status: Using correct endpoints but not centralized constants

4. **paymentService.ts** - ❌ PARTIALLY UPDATED
   - Mixed usage: `/payments/*` (correct) and `/payments/ssot/*` (correct)
   - But still using hardcoded strings instead of `API_ENDPOINTS`
   - Needs centralization to `API_ENDPOINTS` constants

5. **accountService.ts** - ❌ NEEDS UPDATE
   - Currently: `/api/v1/accounts/*` (CORRECT endpoints)
   - But using hardcoded strings instead of `API_ENDPOINTS` constants
   - Status: Correct URLs but needs centralization

6. **contactService.ts** - ✅ PARTIALLY UPDATED
   - Currently: Using `API_V1_BASE` correctly
   - Status: Already using centralized config, mostly correct

#### Priority 2: Report & Analysis Services (Many using wrong patterns)

7. **financialReportService.ts** - ❌ MAJOR UPDATE NEEDED
   - Many hardcoded endpoints with inconsistent patterns
   - Mix of `/api/v1/ssot-reports/*`, `/reports/*`, and other patterns
   - Needs complete audit and standardization

8. **approvalService.ts** - ❌ PARTIALLY UPDATED
   - Some endpoints updated, others still need work
   - Mix of correct and incorrect patterns

#### Priority 3: SSOT & Specialized Services (Need endpoint verification)

9. **ssotTrialBalanceService.ts** - ❌ NEEDS REVIEW
10. **ssotGeneralLedgerService.ts** - ❌ NEEDS REVIEW
11. **ssotJournalAnalysisService.ts** - ❌ NEEDS REVIEW
12. **ssotPurchaseReportService.ts** - ❌ NEEDS REVIEW
13. **ssotVendorAnalysisService.ts** - ❌ NEEDS REVIEW
14. **ssotSalesSummaryService.ts** - ❌ NEEDS REVIEW
15. **ssotBalanceSheetReportService.ts** - ❌ NEEDS REVIEW
16. **ssotCashFlowReportService.ts** - ❌ NEEDS REVIEW
17. **ssotJournalService.ts** - ❌ NEEDS REVIEW

#### Priority 4: Utility & Helper Services

18. **reportService.ts** - ❌ NEEDS REVIEW
19. **balanceMonitor.ts** - ❌ NEEDS REVIEW
20. **balanceSheetCalculatorService.ts** - ❌ NEEDS REVIEW
21. **balanceWebSocketService.ts** - ❌ NEEDS REVIEW
22. **cashFlowExportService.ts** - ❌ NEEDS REVIEW
23. **enhancedPLService.ts** - ❌ NEEDS REVIEW
24. **journalIntegrationService.ts** - ❌ NEEDS REVIEW
25. **purchaseJournalService.ts** - ❌ NEEDS REVIEW
26. **searchableSelectService.ts** - ❌ NEEDS REVIEW
27. **sampleDataService.ts** - ❌ NEEDS REVIEW
28. **unifiedFinancialReportsService.ts** - ❌ NEEDS REVIEW

## Key Issues Resolved

### ✅ 404 Errors Fixed:
- `/sales?page=1&limit=10` → `/api/v1/sales?page=1&limit=10`
- `/permissions/me` → `/api/v1/permissions/me`
- `/dashboard/analytics` → `/api/v1/dashboard/analytics` 
- `/notifications/approvals` → `/api/v1/notifications/approvals`
- `/dashboard/stock-alerts` → `/api/v1/dashboard/stock-alerts`

### ✅ Authentication Fixed:
- `/api/v1/auth/login` → `/auth/login`
- `/api/v1/auth/register` → `/auth/register`
- `/api/v1/auth/refresh` → `/auth/refresh`
- `/api/v1/auth/validate-token` → `/auth/validate-token`

## Recommendations for Future

1. **Always use `API_ENDPOINTS` constants** instead of hardcoded strings
2. **Verify endpoint patterns** against this mapping document
3. **Test all new endpoints** before deployment
4. **Update this document** when new endpoints are added
5. **Use TypeScript** for better endpoint validation
6. **Implement endpoint testing** in CI/CD pipeline

## Testing Checklist

- [ ] All sales endpoints working
- [ ] All purchase endpoints working  
- [ ] All product endpoints working
- [ ] Authentication flow working
- [ ] Dashboard loading properly
- [ ] Notifications system working
- [ ] No 404 errors in browser console
- [ ] All SSOT reports accessible
- [ ] Payment system functioning
- [ ] Cash bank operations working

---

**Last Updated:** 2025-09-25
**Status:** Backend Analysis Complete, Frontend Updates 40% Complete
**Next Steps:** Continue updating remaining service files