import axios from 'axios';
import { getAuthHeaders } from '../utils/authTokenUtils';

export interface SSOTTrialBalanceData {
  company?: CompanyInfo;
  as_of_date: string;
  currency: string;
  accounts: TrialBalanceAccount[];
  total_debits: number;
  total_credits: number;
  is_balanced: boolean;
  difference: number;
  generated_at: string;
}

export interface CompanyInfo {
  name: string;
  address: string;
  city: string;
  state: string;
  phone: string;
  email: string;
  tax_number: string;
}

export interface TrialBalanceAccount {
  account_id: number;
  account_code: string;
  account_name: string;
  account_type: string;
  debit_balance: number;
  credit_balance: number;
  normal_balance?: string;
  ssot_balance?: number;
}

export interface SSOTTrialBalanceParams {
  as_of_date?: string;
  format?: 'json' | 'pdf' | 'excel';
}

class SSOTTrialBalanceService {
  private baseURL: string;

  constructor() {
    // Use relative path to work with Next.js rewrites
    this.baseURL = '/api/v1';
  }

  async generateSSOTTrialBalance(params: SSOTTrialBalanceParams = {}): Promise<SSOTTrialBalanceData> {
    try {
      const queryParams = new URLSearchParams();
      if (params.as_of_date) {
        queryParams.append('as_of_date', params.as_of_date);
      }
      queryParams.append('format', params.format || 'json');

      const response = await axios.get(`${this.baseURL}/ssot-reports/trial-balance?${queryParams}`, {
        headers: getAuthHeaders()
      });

      if (response.data.status === 'success') {
        return response.data.data;
      }

      throw new Error(response.data.message || 'Failed to generate trial balance');
    } catch (error: any) {
      console.error('Error generating SSOT trial balance:', error);
      if (error.response?.data?.message) {
        throw new Error(error.response.data.message);
      }
      throw error;
    }
  }

  formatCurrency(value: number): string {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR'
    }).format(value);
  }

  getAccountTypeIcon(accountType: string): string {
    const icons: { [key: string]: string } = {
      'Asset': '💰',
      'Liability': '📋',
      'Equity': '🏦',
      'Revenue': '📈',
      'Expense': '💸',
      'Other': '📊'
    };
    return icons[accountType] || '📊';
  }

  validateBalance(totalDebits: number, totalCredits: number, tolerance: number = 0.01): boolean {
    return Math.abs(totalDebits - totalCredits) < tolerance;
  }
}

export const ssotTrialBalanceService = new SSOTTrialBalanceService();