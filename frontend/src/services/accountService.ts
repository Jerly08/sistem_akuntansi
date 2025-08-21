import { 
  Account, 
  AccountCreateRequest, 
  AccountUpdateRequest,
  AccountImportRequest,
  AccountSummaryResponse,
  AccountCatalogItem,
  ApiResponse,
  ApiError
} from '@/types/account';

// Base API URL - should be moved to environment variables
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

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
    const url = new URL(`${API_BASE_URL}/accounts`);
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

  // Get account catalog (minimal data for EXPENSE accounts) - for EMPLOYEE role
  async getAccountCatalog(token: string, type: string = 'EXPENSE'): Promise<AccountCatalogItem[]> {
    const url = new URL(`${API_BASE_URL}/accounts/catalog`);
    url.searchParams.append('type', type);
    
    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: this.getHeaders(token),
    });
    
    const result: ApiResponse<AccountCatalogItem[]> = await this.handleResponse(response);
    return result.data;
  }

  // Get cash and bank accounts for payment purposes
  async getPaymentAccounts(token: string): Promise<{
    id: number;
    code: string;
    name: string;
    type: string;
    bank_name?: string;
    account_no?: string;
    currency: string;
    balance: number;
  }[]> {
    const response = await fetch(`${API_BASE_URL}/cashbank/payment-accounts`, {
      method: 'GET',
      headers: this.getHeaders(token),
    });
    
    const result = await this.handleResponse<{
      success: boolean;
      data: {
        id: number;
        code: string;
        name: string;
        type: string;
        bank_name?: string;
        account_no?: string;
        currency: string;
        balance: number;
      }[];
      message: string;
    }>(response);
    
    return result.data;
  }

  // Get single account by code
  async getAccount(token: string, code: string): Promise<Account> {
    const response = await fetch(`${API_BASE_URL}/accounts/${code}`, {
      method: 'GET',
      headers: this.getHeaders(token),
    });
    
    const result: ApiResponse<Account> = await this.handleResponse(response);
    return result.data;
  }

  // Create new account
  async createAccount(token: string, accountData: AccountCreateRequest): Promise<Account> {
    const response = await fetch(`${API_BASE_URL}/accounts`, {
      method: 'POST',
      headers: this.getHeaders(token),
      body: JSON.stringify(accountData),
    });
    
    const result: ApiResponse<Account> = await this.handleResponse(response);
    return result.data;
  }

  // Update existing account
  async updateAccount(token: string, code: string, accountData: AccountUpdateRequest): Promise<Account> {
    const response = await fetch(`${API_BASE_URL}/accounts/${code}`, {
      method: 'PUT',
      headers: this.getHeaders(token),
      body: JSON.stringify(accountData),
    });
    
    const result: ApiResponse<Account> = await this.handleResponse(response);
    return result.data;
  }

  // Delete account
  async deleteAccount(token: string, code: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/accounts/${code}`, {
      method: 'DELETE',
      headers: this.getHeaders(token),
    });
    
    await this.handleResponse(response);
  }

  // Get account hierarchy
  async getAccountHierarchy(token: string): Promise<Account[]> {
    const response = await fetch(`${API_BASE_URL}/accounts/hierarchy`, {
      method: 'GET',
      headers: this.getHeaders(token),
    });
    
    const result: ApiResponse<Account[]> = await this.handleResponse(response);
    return result.data;
  }

  // Get balance summary
  async getBalanceSummary(token: string): Promise<AccountSummaryResponse[]> {
    const response = await fetch(`${API_BASE_URL}/accounts/balance-summary`, {
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
    
    const response = await fetch(`${API_BASE_URL}/accounts/import`, {
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
    const response = await fetch(`${API_BASE_URL}/accounts/export/pdf`, {
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
    const response = await fetch(`${API_BASE_URL}/accounts/export/excel`, {
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
  getAccountTypeLabel(type: string, useEnglish: boolean = false): string {
    if (useEnglish) {
      switch (type) {
        case 'ASSET':
          return 'Asset';
        case 'LIABILITY':
          return 'Liability';
        case 'EQUITY':
          return 'Equity';
        case 'REVENUE':
          return 'Revenue';
        case 'EXPENSE':
          return 'Expense';
        default:
          return type;
      }
    }
    
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

  // Validate account code availability
  async validateAccountCode(token: string, code: string, excludeId?: number): Promise<{
    available: boolean;
    message: string;
    existing_account?: {
      id: number;
      code: string;
      name: string;
    };
  }> {
    const url = new URL(`${API_BASE_URL}/accounts/validate-code`);
    url.searchParams.append('code', code);
    if (excludeId) {
      url.searchParams.append('exclude_id', excludeId.toString());
    }

    const response = await fetch(url.toString(), {
      method: 'GET',
      headers: this.getHeaders(token),
    });

    return this.handleResponse(response);
  }

}

export const accountService = new AccountService();
export default accountService;
