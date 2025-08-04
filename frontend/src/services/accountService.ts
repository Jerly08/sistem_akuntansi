import { 
  Account, 
  AccountCreateRequest, 
  AccountUpdateRequest,
  AccountImportRequest,
  AccountSummaryResponse,
  ApiResponse,
  ApiError
} from '@/types/account';

// Base API URL - should be moved to environment variables
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

class AccountService {
  private getHeaders(token?: string): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };
    
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
    
    return headers;
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      let errorData: ApiError;
      try {
        errorData = await response.json();
      } catch {
        errorData = {
          error: 'Network error',
          code: 'NETWORK_ERROR',
        };
      }
      throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
    }
    
    return response.json();
  }

  // Get all accounts
  async getAccounts(token: string, type?: string): Promise<Account[]> {
    const url = new URL(`${API_BASE_URL}/api/v1/accounts`);
    if (type) {
      url.searchParams.append('type', type);
    }
    
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: this.getHeaders(token),
    });
    
    const result: ApiResponse<Account[]> = await this.handleResponse(response);
    return result.data;
  }

  // Get single account by code
  async getAccount(token: string, code: string): Promise<Account> {
    const response = await fetch(`${API_BASE_URL}/api/v1/accounts/${code}`, {
      method: 'GET',
      headers: this.getHeaders(token),
    });
    
    const result: ApiResponse<Account> = await this.handleResponse(response);
    return result.data;
  }

  // Create new account
  async createAccount(token: string, accountData: AccountCreateRequest): Promise<Account> {
    const response = await fetch(`${API_BASE_URL}/api/v1/accounts`, {
      method: 'POST',
      headers: this.getHeaders(token),
      body: JSON.stringify(accountData),
    });
    
    const result: ApiResponse<Account> = await this.handleResponse(response);
    return result.data;
  }

  // Update existing account
  async updateAccount(token: string, code: string, accountData: AccountUpdateRequest): Promise<Account> {
    const response = await fetch(`${API_BASE_URL}/api/v1/accounts/${code}`, {
      method: 'PUT',
      headers: this.getHeaders(token),
      body: JSON.stringify(accountData),
    });
    
    const result: ApiResponse<Account> = await this.handleResponse(response);
    return result.data;
  }

  // Delete account
  async deleteAccount(token: string, code: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/api/v1/accounts/${code}`, {
      method: 'DELETE',
      headers: this.getHeaders(token),
    });
    
    await this.handleResponse(response);
  }

  // Get account hierarchy
  async getAccountHierarchy(token: string): Promise<Account[]> {
    const response = await fetch(`${API_BASE_URL}/api/v1/accounts/hierarchy`, {
      method: 'GET',
      headers: this.getHeaders(token),
    });
    
    const result: ApiResponse<Account[]> = await this.handleResponse(response);
    return result.data;
  }

  // Get balance summary
  async getBalanceSummary(token: string): Promise<AccountSummaryResponse[]> {
    const response = await fetch(`${API_BASE_URL}/api/v1/accounts/balance-summary`, {
      method: 'GET',
      headers: this.getHeaders(token),
    });
    
    const result: ApiResponse<AccountSummaryResponse[]> = await this.handleResponse(response);
    return result.data;
  }

  // Bulk import accounts
  async importAccounts(token: string, file: File): Promise<{ message: string; count: number }> {
    const formData = new FormData();
    formData.append('file', file);
    
    const headers: HeadersInit = {};
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
    
    const response = await fetch(`${API_BASE_URL}/api/v1/accounts/import`, {
      method: 'POST',
      headers,
      body: formData,
    });
    
    return this.handleResponse(response);
  }

  // Download import template
  async downloadTemplate(): Promise<Blob> {
    const response = await fetch(`${API_BASE_URL}/templates/accounts_import_template.csv`, {
      method: 'GET',
    });
    
    if (!response.ok) {
      throw new Error('Failed to download template');
    }
    
    return response.blob();
  }

  // Export accounts to PDF
  async exportAccountsPDF(token: string): Promise<Blob> {
    const response = await fetch(`${API_BASE_URL}/api/v1/accounts/export/pdf`, {
      method: 'GET',
      headers: this.getHeaders(token),
    });
    
    if (!response.ok) {
      let errorData: ApiError;
      try {
        errorData = await response.json();
      } catch {
        errorData = {
          error: 'Network error',
          code: 'NETWORK_ERROR',
        };
      }
      throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
    }
    
    return response.blob();
  }

  // Export accounts to Excel
  async exportAccountsExcel(token: string): Promise<Blob> {
    const response = await fetch(`${API_BASE_URL}/api/v1/accounts/export/excel`, {
      method: 'GET',
      headers: this.getHeaders(token),
    });
    
    if (!response.ok) {
      let errorData: ApiError;
      try {
        errorData = await response.json();
      } catch {
        errorData = {
          error: 'Network error',
          code: 'NETWORK_ERROR',
        };
      }
      throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
    }
    
    return response.blob();
  }

  // Helper: Format balance for display
  formatBalance(balance: number, currency = 'IDR'): string {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(balance);
  }

  // Helper: Get account type color
  getAccountTypeColor(type: string): string {
    switch (type) {
      case 'ASSET':
        return 'green';
      case 'LIABILITY':
        return 'red';
      case 'EQUITY':
        return 'blue';
      case 'REVENUE':
        return 'purple';
      case 'EXPENSE':
        return 'orange';
      default:
        return 'gray';
    }
  }

  // Helper: Get account type label
  getAccountTypeLabel(type: string): string {
    switch (type) {
      case 'ASSET':
        return 'Aktiva';
      case 'LIABILITY':
        return 'Kewajiban';
      case 'EQUITY':
        return 'Modal';
      case 'REVENUE':
        return 'Pendapatan';
      case 'EXPENSE':
        return 'Beban';
      default:
        return type;
    }
  }
}

export const accountService = new AccountService();
export default accountService;
