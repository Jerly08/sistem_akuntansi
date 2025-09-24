import axios from 'axios';
import { getAuthHeaders } from '../utils/authTokenUtils';

export interface SSOTGeneralLedgerData {
  company?: CompanyInfo;
  start_date: string;
  end_date: string;
  currency: string;
  account?: AccountInfo;
  entries: GeneralLedgerEntry[];
  opening_balance?: number;
  closing_balance?: number;
  total_debits?: number;
  total_credits?: number;
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

export interface AccountInfo {
  account_id: number;
  account_code: string;
  account_name: string;
  account_type: string;
}

export interface GeneralLedgerEntry {
  journal_id: number;
  entry_number: string;
  entry_date: string;
  description: string;
  reference: string;
  account_code?: string;
  account_name?: string;
  debit_amount: number;
  credit_amount: number;
  running_balance: number;
  status: string;
  source_type?: string;
}

export interface SSOTGeneralLedgerParams {
  account_id?: string;
  start_date: string;
  end_date: string;
  format?: 'json' | 'pdf' | 'excel';
}

class SSOTGeneralLedgerService {
  private baseURL: string;

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
  }

  async generateSSOTGeneralLedger(params: SSOTGeneralLedgerParams): Promise<SSOTGeneralLedgerData> {
    try {
      const queryParams = new URLSearchParams({
        start_date: params.start_date,
        end_date: params.end_date,
        format: params.format || 'json'
      });

      if (params.account_id) {
        queryParams.append('account_id', params.account_id);
      }

      const response = await axios.get(`${this.baseURL}/ssot-reports/general-ledger?${queryParams}`, {
        headers: getAuthHeaders()
      });

      if (response.data.status === 'success') {
        return response.data.data;
      }

      throw new Error(response.data.message || 'Failed to generate general ledger');
    } catch (error: any) {
      console.error('Error generating SSOT general ledger:', error);
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

  getEntryTypeColor(sourceType: string): string {
    const colors: { [key: string]: string } = {
      'SALES': 'green',
      'PURCHASE': 'blue',
      'PAYMENT': 'orange',
      'RECEIPT': 'purple',
      'JOURNAL': 'gray',
      'ADJUSTMENT': 'red'
    };
    return colors[sourceType] || 'gray';
  }

  calculateRunningBalance(entries: GeneralLedgerEntry[]): GeneralLedgerEntry[] {
    let runningBalance = 0;
    return entries.map(entry => {
      runningBalance += entry.debit_amount - entry.credit_amount;
      return { ...entry, running_balance: runningBalance };
    });
  }
}

export const ssotGeneralLedgerService = new SSOTGeneralLedgerService();