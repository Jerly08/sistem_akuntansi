// API Configuration
export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

// API Endpoints
export const API_ENDPOINTS = {
  // Auth
  LOGIN: '/api/v1/auth/login',
  REGISTER: '/api/v1/auth/register',
  REFRESH: '/api/v1/auth/refresh',
  PROFILE: '/api/v1/profile',
  
  // Products
  PRODUCTS: '/api/v1/products',
  CATEGORIES: '/api/v1/categories',
  
  // Notifications
  NOTIFICATIONS: '/api/v1/notifications',
  NOTIFICATIONS_UNREAD_COUNT: '/api/v1/notifications/unread-count',
  NOTIFICATIONS_MARK_READ: (id: number) => `/api/v1/notifications/${id}/read`,
  
  // Dashboard
  DASHBOARD_ANALYTICS: '/api/v1/dashboard/analytics',
  DASHBOARD_STOCK_ALERTS: '/api/v1/dashboard/stock-alerts',
  
  // Purchases
  PURCHASES: '/api/v1/purchases',
  PURCHASE_APPROVAL: (id: number) => `/api/v1/purchases/${id}/approve`,
  PURCHASE_REJECT: (id: number) => `/api/v1/purchases/${id}/reject`,
  
  // Contacts
  CONTACTS: '/api/v1/contacts',
  
  // Accounts
  ACCOUNTS: '/api/v1/accounts',
  ACCOUNTS_HIERARCHY: '/api/v1/accounts/hierarchy',
  
  // Cash & Bank
  CASHBANK: '/api/v1/cashbank',
  CASHBANK_ACCOUNTS: '/api/v1/cashbank/accounts',
  CASHBANK_PAYMENT_ACCOUNTS: '/api/v1/cashbank/payment-accounts',
  CASHBANK_BALANCE_SUMMARY: '/api/v1/cashbank/balance-summary',
  CASHBANK_TRANSFER: '/api/v1/cashbank/transfer',
  CASHBANK_DEPOSIT: '/api/v1/cashbank/deposit',
  CASHBANK_WITHDRAWAL: '/api/v1/cashbank/withdrawal',
};

export default API_BASE_URL;
