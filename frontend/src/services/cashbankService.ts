import api from './api';

// Types
export interface CashBank {
  id: number;
  code: string;
  name: string;
  type: 'CASH' | 'BANK';
  account_id: number;
  bank_name?: string;
  account_no?: string;
  currency: string;
  balance: number;
  min_balance: number;
  max_balance: number;
  daily_limit: number;
  monthly_limit: number;
  is_active: boolean;
  is_restricted: boolean;
  user_id: number;
  description?: string;
  created_at: string;
  updated_at: string;
  account?: {
    id: number;
    code: string;
    name: string;
  };
}

export interface CashBankTransaction {
  id: number;
  cash_bank_id: number;
  reference_type: string;
  reference_id: number;
  amount: number;
  balance_after: number;
  transaction_date: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface BalanceSummary {
  total_cash: number;
  total_bank: number;
  total_balance: number;
  by_account: AccountBalance[];
  by_currency: Record<string, number>;
}

export interface AccountBalance {
  account_id: number;
  account_name: string;
  account_type: string;
  balance: number;
  currency: string;
}

export interface CashBankCreateRequest {
  name: string;
  type: 'CASH' | 'BANK';
  account_id?: number;    // GL Account ID from Chart of Accounts
  bank_name?: string;
  account_no?: string;
  currency?: string;
  opening_balance?: number;
  opening_date?: string;
  description?: string;
}

export interface CashBankUpdateRequest {
  name?: string;
  bank_name?: string;
  account_no?: string;
  description?: string;
  is_active?: boolean;
}

export interface TransferRequest {
  from_account_id: number;
  to_account_id: number;
  date: string;
  amount: number;
  exchange_rate?: number;
  reference?: string;
  notes?: string;
}

// Manual journal entry - primarily used for withdrawal transactions
export interface ManualJournalEntry {
  account_id: number;
  description: string;
  debit_amount: number;
  credit_amount: number;
}

export interface DepositRequest {
  account_id: number;
  date: string;
  amount: number;
  reference?: string;
  notes?: string;
  source_account_id?: number; // Revenue account for automatic mode
  journal_entries?: ManualJournalEntry[]; // Deprecated - kept for backward compatibility
}

export interface WithdrawalRequest {
  account_id: number;
  date: string;
  amount: number;
  reference?: string;
  notes?: string;
  journal_entries?: ManualJournalEntry[];
}

export interface TransactionFilter {
  page?: number;
  limit?: number;
  start_date?: string;
  end_date?: string;
}

export interface TransactionResult {
  data: CashBankTransaction[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface ReconciliationRequest {
  date: string;
  statement_balance: number;
  items: ReconciliationItemRequest[];
}

export interface ReconciliationItemRequest {
  transaction_id: number;
  is_cleared: boolean;
  notes?: string;
}

export interface BankReconciliation {
  id: number;
  cash_bank_id: number;
  reconcile_date: string;
  statement_balance: number;
  system_balance: number;
  difference: number;
  status: string;
  user_id: number;
  created_at: string;
  updated_at: string;
}

export interface CashBankTransfer {
  id: number;
  transfer_number: string;
  from_account_id: number;
  to_account_id: number;
  date: string;
  amount: number;
  exchange_rate: number;
  converted_amount: number;
  reference?: string;
  notes?: string;
  status: string;
  user_id: number;
  created_at: string;
  updated_at: string;
}

class CashBankService {
  private readonly baseUrl = '/cashbank';

  // Get all cash and bank accounts
  async getCashBankAccounts(): Promise<CashBank[]> {
    try {
      const response = await api.get(`${this.baseUrl}/accounts`);
      return response.data;
    } catch (error) {
      console.error('Error fetching cash bank accounts:', error);
      throw error;
    }
  }

  // Get account by ID
  async getCashBankById(id: number): Promise<CashBank> {
    try {
      const response = await api.get(`${this.baseUrl}/accounts/${id}`);
      return response.data;
    } catch (error) {
      console.error('Error fetching cash bank account:', error);
      throw error;
    }
  }

  // Create new cash/bank account
  async createCashBankAccount(data: CashBankCreateRequest): Promise<CashBank> {
    try {
      const response = await api.post(`${this.baseUrl}/accounts`, data);
      return response.data;
    } catch (error) {
      console.error('Error creating cash bank account:', error);
      throw error;
    }
  }

  // Update cash/bank account
  async updateCashBankAccount(id: number, data: CashBankUpdateRequest): Promise<CashBank> {
    try {
      const response = await api.put(`${this.baseUrl}/accounts/${id}`, data);
      return response.data;
    } catch (error) {
      console.error('Error updating cash bank account:', error);
      throw error;
    }
  }

  // Delete cash/bank account
  async deleteCashBankAccount(id: number): Promise<void> {
    try {
      await api.delete(`${this.baseUrl}/accounts/${id}`);
    } catch (error) {
      console.error('Error deleting cash bank account:', error);
      throw error;
    }
  }

  // Process transfer
  async processTransfer(data: TransferRequest): Promise<any> {
    try {
      const response = await api.post(`${this.baseUrl}/transfer`, data);
      return response.data;
    } catch (error) {
      console.error('Error processing transfer:', error);
      throw error;
    }
  }

  // Process deposit
  async processDeposit(data: DepositRequest): Promise<CashBankTransaction> {
    try {
      const response = await api.post(`${this.baseUrl}/deposit`, data);
      return response.data;
    } catch (error) {
      console.error('Error processing deposit:', error);
      throw error;
    }
  }

  // Process withdrawal
  async processWithdrawal(data: WithdrawalRequest): Promise<CashBankTransaction> {
    try {
      const response = await api.post(`${this.baseUrl}/withdrawal`, data);
      return response.data;
    } catch (error) {
      console.error('Error processing withdrawal:', error);
      throw error;
    }
  }

  // Get transactions for account
  async getTransactions(accountId: number, filter: TransactionFilter = {}): Promise<TransactionResult> {
    try {
      const params = new URLSearchParams();
      if (filter.page) params.append('page', filter.page.toString());
      if (filter.limit) params.append('limit', filter.limit.toString());
      if (filter.start_date) params.append('start_date', filter.start_date);
      if (filter.end_date) params.append('end_date', filter.end_date);

      const response = await api.get(`${this.baseUrl}/accounts/${accountId}/transactions?${params}`);
      return response.data;
    } catch (error) {
      console.error('Error fetching transactions:', error);
      throw error;
    }
  }

  // Get balance summary
  async getBalanceSummary(): Promise<BalanceSummary> {
    try {
      const response = await api.get(`${this.baseUrl}/balance-summary`);
      return response.data;
    } catch (error) {
      console.error('Error fetching balance summary:', error);
      throw error;
    }
  }

  // Get payment accounts (for dropdowns)
  async getPaymentAccounts(): Promise<CashBank[]> {
    try {
      const response = await api.get(`${this.baseUrl}/payment-accounts`);
      return response.data.data || response.data;
    } catch (error) {
      console.error('Error fetching payment accounts:', error);
      throw error;
    }
  }

  // Get transaction history for an account with filtering
  async getTransactionHistory(accountId: number, filter: TransactionFilter = {}): Promise<TransactionResult> {
    try {
      // Build query parameters
      const params = new URLSearchParams();
      if (filter.page) params.append('page', filter.page.toString());
      if (filter.limit) params.append('limit', filter.limit.toString());
      if (filter.start_date) params.append('start_date', filter.start_date);
      if (filter.end_date) params.append('end_date', filter.end_date);
      
      const response = await api.get(`${this.baseUrl}/accounts/${accountId}/transactions?${params}`);
      return response.data;
    } catch (error) {
      console.error(`Error fetching transaction history for account ${accountId}:`, error);
      throw error;
    }
  }

  // Reconcile bank account
  async reconcileAccount(accountId: number, data: ReconciliationRequest): Promise<BankReconciliation> {
    try {
      const response = await api.post(`${this.baseUrl}/accounts/${accountId}/reconcile`, data);
      return response.data;
    } catch (error) {
      console.error('Error reconciling account:', error);
      throw error;
    }
  }

  // Check GL account links status
  async checkGLAccountLinks(): Promise<any> {
    try {
      const response = await api.get(`${this.baseUrl}/admin/check-gl-links`);
      return response.data;
    } catch (error) {
      console.error('Error checking GL account links:', error);
      throw error;
    }
  }

  // Fix GL account links
  async fixGLAccountLinks(): Promise<any> {
    try {
      const response = await api.post(`${this.baseUrl}/admin/fix-gl-links`);
      return response.data;
    } catch (error) {
      console.error('Error fixing GL account links:', error);
      throw error;
    }
  }

  // Get revenue accounts for deposit form
  async getRevenueAccounts(): Promise<any[]> {
    try {
      const response = await api.get(`${this.baseUrl}/revenue-accounts`);
      return response.data;
    } catch (error) {
      console.error('Error fetching revenue accounts:', error);
      throw error;
    }
  }

  // Get deposit source accounts (Revenue + Equity) for deposit form
  async getDepositSourceAccounts(): Promise<{revenue: any[], equity: any[]}> {
    try {
      const response = await api.get(`${this.baseUrl}/deposit-source-accounts`);
      return response.data.data;
    } catch (error) {
      console.error('Error fetching deposit source accounts:', error);
      throw error;
    }
  }
}

export default new CashBankService();
