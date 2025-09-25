// API Configuration
export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || '';
// For local development, we use relative URLs since Next.js handles rewrites
// This allows Next.js proxy to handle the backend communication
export const API_V1_BASE = `/api/v1`;

// COMPREHENSIVE API Endpoints - Based on Backend Routes Analysis
export const API_ENDPOINTS = {
  // Authentication (with /api/v1 prefix - corrected based on actual backend routes)
  LOGIN: '/api/v1/auth/login',
  REGISTER: '/api/v1/auth/register', 
  REFRESH: '/api/v1/auth/refresh',
  VALIDATE_TOKEN: '/api/v1/auth/validate-token',
  PROFILE: '/api/v1/profile',
  
  // Products (with /api/v1 prefix)
  PRODUCTS: '/api/v1/products',
  CATEGORIES: '/api/v1/categories',
  
  // Notifications (with /api/v1 prefix)
  NOTIFICATIONS: '/api/v1/notifications',
  NOTIFICATIONS_UNREAD_COUNT: '/api/v1/notifications/unread-count',
  NOTIFICATIONS_MARK_READ: (id: number) => `/api/v1/notifications/${id}/read`,
  NOTIFICATIONS_APPROVALS: '/api/v1/notifications/approvals',
  NOTIFICATIONS_BY_TYPE: (type: string) => `/api/v1/notifications/type/${type}`,
  
  // Dashboard (no /api/v1 prefix based on Swagger)
  DASHBOARD_ANALYTICS: '/dashboard/analytics',
  DASHBOARD_FINANCE: '/dashboard/finance',
  DASHBOARD_STOCK_ALERTS: '/api/v1/dashboard/stock-alerts', // Keep v1 for this one
  
  // Permissions (with /api/v1 prefix)
  PERMISSIONS_ME: '/api/v1/permissions/me',
  
  // Purchases (with /api/v1 prefix)
  PURCHASES: '/api/v1/purchases',
  PURCHASES_BY_ID: (id: number) => `/api/v1/purchases/${id}`,
  PURCHASES_APPROVAL_STATS: '/api/v1/purchases/approval-stats',
  PURCHASES_SUBMIT_APPROVAL: (id: number) => `/api/v1/purchases/${id}/submit-approval`,
  PURCHASES_APPROVE: (id: number) => `/api/v1/purchases/${id}/approve`,
  PURCHASES_REJECT: (id: number) => `/api/v1/purchases/${id}/reject`,
  PURCHASES_APPROVAL_HISTORY: (id: number) => `/api/v1/purchases/${id}/approval-history`,
  PURCHASES_PENDING_APPROVAL: '/api/v1/purchases/pending-approval',
  PURCHASES_DOCUMENTS: (id: number) => `/api/v1/purchases/${id}/documents`,
  PURCHASES_DELETE_DOCUMENT: (documentId: number) => `/api/v1/purchases/documents/${documentId}`,
  PURCHASES_RECEIPTS: '/api/v1/purchases/receipts',
  PURCHASES_RECEIPTS_BY_ID: (id: number) => `/api/v1/purchases/${id}/receipts`,
  PURCHASES_RECEIPT_PDF: (receiptId: number) => `/api/v1/purchases/receipts/${receiptId}/pdf`,
  PURCHASES_ALL_RECEIPTS_PDF: (id: number) => `/api/v1/purchases/${id}/receipts/pdf`,
  PURCHASES_SUMMARY: '/api/v1/purchases/summary',
  PURCHASES_PENDING_APPROVALS: '/api/v1/purchases/pending-approvals',
  PURCHASES_DASHBOARD: '/api/v1/purchases/dashboard',
  PURCHASES_VENDOR_SUMMARY: (vendorId: number) => `/api/v1/purchases/vendor/${vendorId}/summary`,
  PURCHASES_PAYMENTS: (id: number) => `/api/v1/purchases/${id}/payments`,
  PURCHASES_FOR_PAYMENT: (id: number) => `/api/v1/purchases/${id}/for-payment`,
  PURCHASES_INTEGRATED_PAYMENT: (id: number) => `/api/v1/purchases/${id}/integrated-payment`,
  PURCHASES_MATCHING: (id: number) => `/api/v1/purchases/${id}/matching`,
  PURCHASES_VALIDATE_MATCHING: (id: number) => `/api/v1/purchases/${id}/validate-matching`,
  PURCHASES_JOURNAL_ENTRIES: (id: number) => `/api/v1/purchases/${id}/journal-entries`,
  
  // Assets (with /api/v1 prefix)
  ASSETS: '/api/v1/assets',
  ASSETS_BY_ID: (id: number) => `/api/v1/assets/${id}`,
  ASSETS_UPLOAD_IMAGE: '/api/v1/assets/upload-image',
  ASSETS_CATEGORIES: '/api/v1/assets/categories',
  ASSETS_SUMMARY: '/api/v1/assets/summary',
  ASSETS_DEPRECIATION_REPORT: '/api/v1/assets/depreciation-report',
  ASSETS_DEPRECIATION_SCHEDULE: (id: number) => `/api/v1/assets/${id}/depreciation-schedule`,
  ASSETS_CALCULATE_DEPRECIATION: (id: number) => `/api/v1/assets/${id}/calculate-depreciation`,
  
  // Approval Workflows (with /api/v1 prefix)
  APPROVAL_WORKFLOWS: '/api/v1/approval-workflows',
  
  // Contacts (with /api/v1 prefix)
  CONTACTS: '/api/v1/contacts',
  
  // Sales (with /api/v1 prefix)
  SALES: '/api/v1/sales',
  SALES_BY_ID: (id: number) => `/api/v1/sales/${id}`,
  SALES_CONFIRM: (id: number) => `/api/v1/sales/${id}/confirm`,
  SALES_INVOICE: (id: number) => `/api/v1/sales/${id}/invoice`,
  SALES_CANCEL: (id: number) => `/api/v1/sales/${id}/cancel`,
  SALES_PAYMENTS: (id: number) => `/api/v1/sales/${id}/payments`,
  SALES_FOR_PAYMENT: (id: number) => `/api/v1/sales/${id}/for-payment`,
  SALES_INTEGRATED_PAYMENT: (id: number) => `/api/v1/sales/${id}/integrated-payment`,
  SALES_RETURNS: (id: number) => `/api/v1/sales/${id}/returns`,
  SALES_ALL_RETURNS: '/api/v1/sales/returns',
  SALES_SUMMARY: '/api/v1/sales/summary',
  SALES_ANALYTICS: '/api/v1/sales/analytics',
  SALES_RECEIVABLES: '/api/v1/sales/receivables',
  SALES_INVOICE_PDF: (id: number) => `/api/v1/sales/${id}/invoice/pdf`,
  SALES_REPORT_PDF: '/api/v1/sales/report/pdf',
  SALES_CUSTOMER: (customerId: number) => `/api/v1/sales/customer/${customerId}`,
  SALES_CUSTOMER_INVOICES: (customerId: number) => `/api/v1/sales/customer/${customerId}/invoices`,
  
  // Accounts (with /api/v1 prefix)
  ACCOUNTS: '/api/v1/accounts',
  ACCOUNTS_HIERARCHY: '/api/v1/accounts/hierarchy',
  ACCOUNTS_BALANCE_SUMMARY: '/api/v1/accounts/balance-summary',
  ACCOUNTS_VALIDATE_CODE: '/api/v1/accounts/validate-code',
  ACCOUNTS_FIX_HEADER_STATUS: '/api/v1/accounts/fix-header-status',
  ACCOUNTS_BY_CODE: (code: string) => `/api/v1/accounts/${code}`,
  ACCOUNTS_ADMIN_DELETE: (code: string) => `/api/v1/accounts/admin/${code}`,
  ACCOUNTS_IMPORT: '/api/v1/accounts/import',
  ACCOUNTS_EXPORT_PDF: '/api/v1/accounts/export/pdf',
  ACCOUNTS_EXPORT_EXCEL: '/api/v1/accounts/export/excel',
  ACCOUNTS_CATALOG: '/api/v1/accounts/catalog', // Public
  ACCOUNTS_CREDIT: '/api/v1/accounts/credit', // Public
  
  // Products (with /api/v1 prefix)
  PRODUCTS: '/api/v1/products',
  PRODUCTS_BY_ID: (id: number) => `/api/v1/products/${id}`,
  PRODUCTS_ADJUST_STOCK: '/api/v1/products/adjust-stock',
  PRODUCTS_OPNAME: '/api/v1/products/opname',
  PRODUCTS_UPLOAD_IMAGE: '/api/v1/products/upload-image',
  
  // Categories (with /api/v1 prefix)
  CATEGORIES: '/api/v1/categories',
  CATEGORIES_TREE: '/api/v1/categories/tree',
  CATEGORIES_BY_ID: (id: number) => `/api/v1/categories/${id}`,
  CATEGORIES_PRODUCTS: (id: number) => `/api/v1/categories/${id}/products`,
  
  // Product Units (with /api/v1 prefix)
  PRODUCT_UNITS: '/api/v1/product-units',
  PRODUCT_UNITS_BY_ID: (id: number) => `/api/v1/product-units/${id}`,
  
  // Warehouse Locations (with /api/v1 prefix)
  WAREHOUSE_LOCATIONS: '/api/v1/warehouse-locations',
  WAREHOUSE_LOCATIONS_BY_ID: (id: number) => `/api/v1/warehouse-locations/${id}`,
  
  // Inventory (with /api/v1 prefix)
  INVENTORY_MOVEMENTS: '/api/v1/inventory/movements',
  INVENTORY_LOW_STOCK: '/api/v1/inventory/low-stock',
  INVENTORY_VALUATION: '/api/v1/inventory/valuation',
  INVENTORY_REPORT: '/api/v1/inventory/report',
  INVENTORY_BULK_PRICE_UPDATE: '/api/v1/inventory/bulk-price-update',
  
  // Users (with /api/v1 prefix)
  USERS: '/api/v1/users',
  USERS_BY_ID: (id: number) => `/api/v1/users/${id}`,
  
  // Permissions (with /api/v1 prefix)  
  PERMISSIONS_USERS: '/api/v1/permissions/users',
  PERMISSIONS_USER_BY_ID: (userId: number) => `/api/v1/permissions/users/${userId}`,
  PERMISSIONS_USER_RESET: (userId: number) => `/api/v1/permissions/users/${userId}/reset`,
  PERMISSIONS_ME: '/api/v1/permissions/me',
  PERMISSIONS_CHECK: '/api/v1/permissions/check',
  
  // Cash & Bank (no /api/v1 prefix based on analysis)
  CASHBANK: '/api/cashbank',
  CASHBANK_ACCOUNTS: '/api/cashbank/accounts',
  CASHBANK_ACCOUNT_BY_ID: (id: number) => `/api/cashbank/accounts/${id}`,
  CASHBANK_ACCOUNT_TRANSACTIONS: (id: number) => `/api/cashbank/accounts/${id}/transactions`,
  CASHBANK_PAYMENT_ACCOUNTS: '/api/cashbank/payment-accounts',
  CASHBANK_REVENUE_ACCOUNTS: '/api/cashbank/revenue-accounts',
  CASHBANK_DEPOSIT_SOURCE_ACCOUNTS: '/api/cashbank/deposit-source-accounts',
  CASHBANK_BALANCE_SUMMARY: '/api/cashbank/balance-summary',
  CASHBANK_TRANSFER: '/api/cashbank/transfer',
  CASHBANK_DEPOSIT: '/api/cashbank/deposit',
  CASHBANK_WITHDRAWAL: '/api/cashbank/withdrawal',
  
  // Cash Bank SSOT Routes (with /api/v1 prefix)
  CASH_BANK_SSOT_ACCOUNTS: '/api/v1/cash-bank/accounts',
  CASH_BANK_SSOT_ACCOUNT_BY_ID: (id: number) => `/api/v1/cash-bank/accounts/${id}`,
  CASH_BANK_SSOT_ACCOUNT_TRANSACTIONS: (id: number) => `/api/v1/cash-bank/accounts/${id}/transactions`,
  CASH_BANK_SSOT_ACCOUNT_RECONCILE: (id: number) => `/api/v1/cash-bank/accounts/${id}/reconcile`,
  CASH_BANK_SSOT_DEPOSIT: '/api/v1/cash-bank/transactions/deposit',
  CASH_BANK_SSOT_WITHDRAWAL: '/api/v1/cash-bank/transactions/withdrawal',
  CASH_BANK_SSOT_TRANSFER: '/api/v1/cash-bank/transactions/transfer',
  CASH_BANK_SSOT_BALANCE_SUMMARY: '/api/v1/cash-bank/reports/balance-summary',
  CASH_BANK_SSOT_PAYMENT_ACCOUNTS: '/api/v1/cash-bank/reports/payment-accounts',
  CASH_BANK_SSOT_JOURNALS: '/api/v1/cash-bank/ssot/journals',
  CASH_BANK_SSOT_VALIDATE: '/api/v1/cash-bank/ssot/validate-integrity',
  
  // Admin
  ADMIN_CHECK_CASHBANK_GL: '/api/admin/check-cashbank-gl-links',
  ADMIN_FIX_CASHBANK_GL: '/api/admin/fix-cashbank-gl-links',
  
  // Balance Monitoring
  MONITORING_BALANCE_HEALTH: '/api/monitoring/balance-health',
  MONITORING_BALANCE_SYNC: '/api/monitoring/balance-sync',
  MONITORING_DISCREPANCIES: '/api/monitoring/discrepancies',
  MONITORING_FIX_DISCREPANCIES: '/api/monitoring/fix-discrepancies',
  MONITORING_SYNC_STATUS: '/api/monitoring/sync-status',
  
  // API Usage Monitoring  
  MONITORING_API_ANALYTICS: '/monitoring/api-usage/analytics',
  MONITORING_API_STATS: '/monitoring/api-usage/stats',
  MONITORING_API_TOP: '/monitoring/api-usage/top',
  MONITORING_API_UNUSED: '/monitoring/api-usage/unused',
  MONITORING_API_RESET: '/monitoring/api-usage/reset',
  
  // Payments (no /api/v1 prefix based on Swagger)
  PAYMENTS: '/api/payments',
  PAYMENTS_ANALYTICS: '/api/payments/analytics', 
  PAYMENTS_SUMMARY: '/api/payments/summary',
  PAYMENTS_UNPAID_BILLS: (vendorId: number) => `/api/payments/unpaid-bills/${vendorId}`,
  PAYMENTS_UNPAID_INVOICES: (customerId: number) => `/api/payments/unpaid-invoices/${customerId}`,
  PAYMENTS_EXPORT_EXCEL: '/api/payments/export/excel',
  PAYMENTS_REPORT_PDF: '/api/payments/report/pdf',
  PAYMENTS_BY_ID: (id: number) => `/api/payments/${id}`,
  PAYMENTS_CANCEL: (id: number) => `/api/payments/${id}/cancel`,
  PAYMENTS_PDF: (id: number) => `/api/payments/${id}/pdf`,
  
  // Payment Integration
  PAYMENTS_ACCOUNT_BALANCES: '/api/payments/account-balances/real-time',
  PAYMENTS_REFRESH_BALANCES: '/api/payments/account-balances/refresh',
  PAYMENTS_ENHANCED: '/api/payments/enhanced-with-journal',
  PAYMENTS_INTEGRATION_METRICS: '/api/payments/integration-metrics',
  PAYMENTS_JOURNAL_ENTRIES: '/api/payments/journal-entries',
  PAYMENTS_PREVIEW_JOURNAL: '/api/payments/preview-journal',
  PAYMENTS_ACCOUNT_UPDATES: (id: number) => `/api/payments/${id}/account-updates`,
  PAYMENTS_REVERSE: (id: number) => `/api/payments/${id}/reverse`,
  PAYMENTS_WITH_JOURNAL: (id: number) => `/api/payments/${id}/with-journal`,
  
  // Security (with /api/v1 prefix)
  SECURITY_ALERTS: '/api/v1/admin/security/alerts',
  SECURITY_ALERT_ACKNOWLEDGE: (id: number) => `/api/v1/admin/security/alerts/${id}/acknowledge`,
  SECURITY_CLEANUP: '/api/v1/admin/security/cleanup',
  SECURITY_CONFIG: '/api/v1/admin/security/config',
  SECURITY_INCIDENTS: '/api/v1/admin/security/incidents',
  SECURITY_INCIDENT_BY_ID: (id: number) => `/api/v1/admin/security/incidents/${id}`,
  SECURITY_INCIDENT_RESOLVE: (id: number) => `/api/v1/admin/security/incidents/${id}/resolve`,
  SECURITY_IP_WHITELIST: '/api/v1/admin/security/ip-whitelist',
  SECURITY_METRICS: '/api/v1/admin/security/metrics',
  
  // Journal (with /api/v1 prefix)
  JOURNALS: '/api/v1/journals',
  JOURNALS_ACCOUNT_BALANCES: '/api/v1/journals/account-balances',
  JOURNALS_REFRESH_BALANCES: '/api/v1/journals/account-balances/refresh',
  JOURNALS_SUMMARY: '/api/v1/journals/summary',
  JOURNALS_BY_ID: (id: number) => `/api/v1/journals/${id}`,
  
  // Journal Drilldown (no /api/v1 prefix based on Swagger)
  JOURNAL_DRILLDOWN: '/journal-drilldown',
  JOURNAL_DRILLDOWN_ACCOUNTS: '/journal-drilldown/accounts',
  JOURNAL_DRILLDOWN_ENTRIES: '/journal-drilldown/entries',
  JOURNAL_DRILLDOWN_ENTRY_BY_ID: (id: number) => `/journal-drilldown/entries/${id}`,
  
  // Optimized Reports (with /api/v1 prefix)
  REPORTS_OPTIMIZED_BALANCE_SHEET: '/api/v1/reports/optimized/balance-sheet',
  REPORTS_OPTIMIZED_PROFIT_LOSS: '/api/v1/reports/optimized/profit-loss',
  REPORTS_OPTIMIZED_TRIAL_BALANCE: '/api/v1/reports/optimized/trial-balance',
  REPORTS_OPTIMIZED_REFRESH_BALANCES: '/api/v1/reports/optimized/refresh-balances',
  
  // SSOT Reports (with /api/v1 prefix)
  SSOT_REPORTS_GENERAL_LEDGER: '/api/v1/ssot-reports/general-ledger',
  SSOT_REPORTS_INTEGRATED: '/api/v1/ssot-reports/integrated',
  SSOT_REPORTS_JOURNAL_ANALYSIS: '/api/v1/ssot-reports/journal-analysis',
  SSOT_REPORTS_PURCHASE_REPORT: '/api/v1/ssot-reports/purchase-report',
  SSOT_REPORTS_PURCHASE_VALIDATE: '/api/v1/ssot-reports/purchase-report/validate',
  SSOT_REPORTS_PURCHASE_SUMMARY: '/api/v1/ssot-reports/purchase-summary',
  SSOT_REPORTS_REFRESH: '/api/v1/ssot-reports/refresh',
  SSOT_REPORTS_SALES_SUMMARY: '/api/v1/ssot-reports/sales-summary',
  SSOT_REPORTS_STATUS: '/api/v1/ssot-reports/status',
  SSOT_REPORTS_TRIAL_BALANCE: '/api/v1/ssot-reports/trial-balance',
  SSOT_REPORTS_VENDOR_ANALYSIS: '/api/v1/ssot-reports/vendor-analysis',
  
  // SSOT Balance Sheet & Cash Flow (no /api/v1 prefix based on Swagger)
  REPORTS_SSOT_PROFIT_LOSS: '/reports/ssot-profit-loss',
  REPORTS_SSOT_BALANCE_SHEET: '/reports/ssot/balance-sheet',
  REPORTS_SSOT_BALANCE_SHEET_DETAILS: '/reports/ssot/balance-sheet/account-details',
  REPORTS_SSOT_CASH_FLOW: '/reports/ssot/cash-flow',
  
  // Journal Drilldown (with /api/v1 prefix)
  JOURNAL_DRILLDOWN: '/api/v1/journal-drilldown',
  JOURNAL_DRILLDOWN_ENTRIES: '/api/v1/journal-drilldown/entries',
  JOURNAL_DRILLDOWN_ENTRY_BY_ID: (id: number) => `/api/v1/journal-drilldown/entries/${id}`,
  JOURNAL_DRILLDOWN_ACCOUNTS: '/api/v1/journal-drilldown/accounts',
  
  // Monitoring & Admin (with /api/v1 prefix) 
  MONITORING_STATUS: '/api/v1/monitoring/status',
  MONITORING_RATE_LIMITS: '/api/v1/monitoring/rate-limits',
  MONITORING_SECURITY_ALERTS: '/api/v1/monitoring/security-alerts',
  MONITORING_AUDIT_LOGS: '/api/v1/monitoring/audit-logs',
  MONITORING_TOKEN_STATS: '/api/v1/monitoring/token-stats',
  MONITORING_REFRESH_EVENTS: '/api/v1/monitoring/refresh-events',
  MONITORING_USER_SECURITY: (userId: number) => `/api/v1/monitoring/users/${userId}/security-summary`,
  MONITORING_STARTUP_STATUS: '/api/v1/monitoring/startup-status',
  MONITORING_FIX_ACCOUNT_HEADERS: '/api/v1/monitoring/fix-account-headers',
  MONITORING_BALANCE_SYNC: '/api/v1/monitoring/balance-sync',
  MONITORING_FIX_DISCREPANCIES: '/api/v1/monitoring/fix-discrepancies',
  MONITORING_BALANCE_HEALTH: '/api/v1/monitoring/balance-health',
  MONITORING_DISCREPANCIES: '/api/v1/monitoring/discrepancies',
  MONITORING_SYNC_STATUS: '/api/v1/monitoring/sync-status',
  MONITORING_API_USAGE_STATS: '/api/v1/monitoring/api-usage/stats',
  MONITORING_API_USAGE_TOP: '/api/v1/monitoring/api-usage/top',
  MONITORING_API_USAGE_UNUSED: '/api/v1/monitoring/api-usage/unused',
  MONITORING_API_USAGE_ANALYTICS: '/api/v1/monitoring/api-usage/analytics',
  MONITORING_API_USAGE_RESET: '/api/v1/monitoring/api-usage/reset',
  MONITORING_PERFORMANCE_REPORT: '/api/v1/monitoring/performance/report',
  MONITORING_PERFORMANCE_METRICS: '/api/v1/monitoring/performance/metrics',
  MONITORING_PERFORMANCE_BOTTLENECKS: '/api/v1/monitoring/performance/bottlenecks',
  MONITORING_PERFORMANCE_RECOMMENDATIONS: '/api/v1/monitoring/performance/recommendations',
  MONITORING_PERFORMANCE_SYSTEM: '/api/v1/monitoring/performance/system',
  MONITORING_PERFORMANCE_CLEAR: '/api/v1/monitoring/performance/metrics/clear',
  MONITORING_PERFORMANCE_TEST: '/api/v1/monitoring/performance/test',
  MONITORING_TIMEOUT_DIAGNOSTICS: '/api/v1/monitoring/timeout/diagnostics',
  MONITORING_TIMEOUT_HEALTH: '/api/v1/monitoring/timeout/health',
  
  // Debug Routes (development only, /api/v1/debug)
  DEBUG_AUTH_CONTEXT: '/api/v1/debug/auth/context',
  DEBUG_AUTH_ROLE: '/api/v1/debug/auth/role',
  DEBUG_CASHBANK_PERMISSION: '/api/v1/debug/auth/test-cashbank-permission',
  DEBUG_PAYMENTS_PERMISSION: '/api/v1/debug/auth/test-payments-permission',
  
  // Static Files
  TEMPLATES: (filepath: string) => `/templates/${filepath}`,
  UPLOADS: (filepath: string) => `/uploads/${filepath}`,
  
  // Documentation - Standardized
  SWAGGER: '/swagger/index.html',
  DOCS: '/docs/index.html',
  OPENAPI_DOC: '/openapi/doc.json',
  
  // Health Check
  HEALTH: '/api/v1/health',
};

export default API_BASE_URL;
