// CashBank SSOT Integration API Service
// This service handles all API calls for the integrated CashBank-SSOT system

import {
  IntegratedSummaryResponse,
  IntegratedAccountResponse,
  ReconciliationData,
  JournalEntriesResponse,
  TransactionHistoryResponse,
  ApiResponse
} from '../types/cashBankIntegration.types';

class CashBankIntegrationService {
  private baseURL: string;

  constructor(baseURL: string = '/api/v1') {
    this.baseURL = baseURL;
  }

  // Get authentication token from localStorage or other secure storage
  private getAuthToken(): string | null {
    // In a real app, this would be retrieved from secure storage
    return localStorage.getItem('auth_token');
  }

  // Create authenticated headers
  private getAuthHeaders(): Record<string, string> {
    const token = this.getAuthToken();
    return {
      'Content-Type': 'application/json',
      ...(token && { 'Authorization': `Bearer ${token}` })
    };
  }

  // Handle API errors
  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ message: 'Network error' }));
      throw new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`);
    }
    
    const data = await response.json();
    return data as T;
  }

  // Get integrated summary of all cash/bank accounts
  async getIntegratedSummary(): Promise<IntegratedSummaryResponse> {
    const response = await fetch(`${this.baseURL}/cashbank/integrated/summary`, {
      method: 'GET',
      headers: this.getAuthHeaders(),
    });
    
    const apiResponse = await this.handleResponse<ApiResponse<IntegratedSummaryResponse>>(response);
    return apiResponse.data;
  }

  // Get integrated account details with SSOT data
  async getIntegratedAccountDetails(accountId: number): Promise<IntegratedAccountResponse> {
    const response = await fetch(`${this.baseURL}/cashbank/integrated/accounts/${accountId}`, {
      method: 'GET',
      headers: this.getAuthHeaders(),
    });
    
    const apiResponse = await this.handleResponse<ApiResponse<IntegratedAccountResponse>>(response);
    return apiResponse.data;
  }

  // Get account reconciliation data
  async getAccountReconciliation(accountId: number): Promise<ReconciliationData> {
    const response = await fetch(`${this.baseURL}/cashbank/integrated/accounts/${accountId}/reconciliation`, {
      method: 'GET',
      headers: this.getAuthHeaders(),
    });
    
    const apiResponse = await this.handleResponse<ApiResponse<ReconciliationData>>(response);
    return apiResponse.data;
  }

  // Get journal entries for account
  async getAccountJournalEntries(
    accountId: number,
    params?: {
      start_date?: string;
      end_date?: string;
      page?: number;
      limit?: number;
    }
  ): Promise<JournalEntriesResponse> {
    const searchParams = new URLSearchParams();
    
    if (params?.start_date) searchParams.set('start_date', params.start_date);
    if (params?.end_date) searchParams.set('end_date', params.end_date);
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.limit) searchParams.set('limit', params.limit.toString());

    const queryString = searchParams.toString();
    const url = `${this.baseURL}/cashbank/integrated/accounts/${accountId}/journal-entries${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
      method: 'GET',
      headers: this.getAuthHeaders(),
    });
    
    const apiResponse = await this.handleResponse<ApiResponse<JournalEntriesResponse>>(response);
    return apiResponse.data;
  }

  // Get transaction history for account
  async getAccountTransactionHistory(
    accountId: number,
    params?: {
      page?: number;
      limit?: number;
    }
  ): Promise<TransactionHistoryResponse> {
    const searchParams = new URLSearchParams();
    
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.limit) searchParams.set('limit', params.limit.toString());

    const queryString = searchParams.toString();
    const url = `${this.baseURL}/cashbank/integrated/accounts/${accountId}/transactions${queryString ? '?' + queryString : ''}`;
    
    const response = await fetch(url, {
      method: 'GET',
      headers: this.getAuthHeaders(),
    });
    
    const apiResponse = await this.handleResponse<ApiResponse<TransactionHistoryResponse>>(response);
    return apiResponse.data;
  }

  // Utility methods for formatting and display

  // Format currency values
  formatCurrency(amount: string | number, currency: string = 'IDR'): string {
    const numAmount = typeof amount === 'string' ? parseFloat(amount) : amount;
    
    if (currency === 'IDR') {
      return new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
      }).format(numAmount);
    }
    
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency,
    }).format(numAmount);
  }

  // Format date strings
  formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('id-ID', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  }

  // Format datetime strings
  formatDateTime(dateString: string): string {
    return new Date(dateString).toLocaleString('id-ID', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  // Get reconciliation status color
  getReconciliationStatusColor(status: string): string {
    switch (status) {
      case 'MATCHED':
        return 'text-green-600 bg-green-50 border-green-200';
      case 'MINOR_VARIANCE':
        return 'text-yellow-600 bg-yellow-50 border-yellow-200';
      case 'VARIANCE':
        return 'text-red-600 bg-red-50 border-red-200';
      default:
        return 'text-gray-600 bg-gray-50 border-gray-200';
    }
  }

  // Get reconciliation status label
  getReconciliationStatusLabel(status: string): string {
    switch (status) {
      case 'MATCHED':
        return 'Sesuai';
      case 'MINOR_VARIANCE':
        return 'Selisih Kecil';
      case 'VARIANCE':
        return 'Ada Selisih';
      default:
        return 'Tidak Diketahui';
    }
  }

  // Get account type icon
  getAccountTypeIcon(type: string): string {
    switch (type) {
      case 'CASH':
        return '💵';
      case 'BANK':
        return '🏦';
      default:
        return '💰';
    }
  }

  // Check if account has variance
  hasVariance(variance: string): boolean {
    const numVariance = parseFloat(variance);
    return Math.abs(numVariance) > 0.01; // More than 1 cent difference
  }

  // Get variance severity
  getVarianceSeverity(variance: string): 'low' | 'medium' | 'high' {
    const numVariance = Math.abs(parseFloat(variance));
    
    if (numVariance <= 100) return 'low';
    if (numVariance <= 1000) return 'medium';
    return 'high';
  }

  // Method aliases for backward compatibility with components
  async getIntegratedAccount(accountId: number): Promise<IntegratedAccountResponse> {
    return this.getIntegratedAccountDetails(accountId);
  }

  async getReconciliation(accountId: number): Promise<ReconciliationData> {
    return this.getAccountReconciliation(accountId);
  }

  async getJournalEntries(
    accountId: number,
    page: number = 1,
    pageSize: number = 20
  ): Promise<JournalEntriesResponse> {
    return this.getAccountJournalEntries(accountId, { page, limit: pageSize });
  }

  async getTransactionHistory(
    accountId: number,
    page: number = 1,
    pageSize: number = 20
  ): Promise<TransactionHistoryResponse> {
    return this.getAccountTransactionHistory(accountId, { page, limit: pageSize });
  }
}

// Export singleton instance
export const cashBankIntegrationService = new CashBankIntegrationService();
export default CashBankIntegrationService;