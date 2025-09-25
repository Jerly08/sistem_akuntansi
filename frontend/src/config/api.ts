// API Configuration
export const API_BASE_URL = (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080') + '/api/v1';
// For local development, we use relative URLs since Next.js handles rewrites
// This prevents duplication of /api/v1 in the URL
export const API_V1_BASE = `/api/v1`;

// API Endpoints - Base URL already includes /api/v1, so these are relative to that
export const API_ENDPOINTS = {
  // Auth
  LOGIN: '/auth/login',
  REGISTER: '/auth/register',
  REFRESH: '/auth/refresh',
  PROFILE: '/profile',
  
  // Products
  PRODUCTS: '/products',
  CATEGORIES: '/categories',
  
  // Notifications
  NOTIFICATIONS: '/notifications',
  NOTIFICATIONS_UNREAD_COUNT: '/notifications/unread-count',
  NOTIFICATIONS_MARK_READ: (id: number) => `/notifications/${id}/read`,
  
  // Dashboard
  DASHBOARD_ANALYTICS: '/dashboard/analytics',
  DASHBOARD_STOCK_ALERTS: '/dashboard/stock-alerts',
  
  // Purchases
  PURCHASES: '/purchases',
  PURCHASE_APPROVAL: (id: number) => `/purchases/${id}/approve`,
  PURCHASE_REJECT: (id: number) => `/purchases/${id}/reject`,
  
  // Contacts
  CONTACTS: '/contacts',
  
  // Accounts
  ACCOUNTS: '/accounts',
  ACCOUNTS_HIERARCHY: '/accounts/hierarchy',
  
  // Cash & Bank
  CASHBANK: '/cashbank',
  CASHBANK_ACCOUNTS: '/cashbank/accounts',
  CASHBANK_PAYMENT_ACCOUNTS: '/cashbank/payment-accounts',
  CASHBANK_BALANCE_SUMMARY: '/cashbank/balance-summary',
  CASHBANK_TRANSFER: '/cashbank/transfer',
  CASHBANK_DEPOSIT: '/cashbank/deposit',
  CASHBANK_WITHDRAWAL: '/cashbank/withdrawal',
  
  // Documentation
  SWAGGER: '/swagger/index.html',
  DOCS: '/docs/index.html',
  
  // Health Check
  HEALTH: '/health',
};

export default API_BASE_URL;
