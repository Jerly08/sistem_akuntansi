// TypeScript interfaces for CashBank SSOT Integration
// These interfaces match the backend service responses

export interface IntegratedAccountDetail {
  id: number;
  account_id: number;
  code: string;
  name: string;
  type: 'CASH' | 'BANK';
  balance: string;
  ssot_balance: string;
  variance: string;
  reconciliation_status: ReconciliationStatus;
  last_transaction_date?: string;
  total_journal_entries: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface IntegratedTransaction {
  id: number;
  amount: string;
  type: string;
  number: string;
  created_at: string;
  description: string;
  balance_after: string;
  reference_number: string;
  journal_entry_id?: number;
  journal_entry_number?: string;
}

export interface IntegratedJournalEntry {
  id: number;
  entry_number: string;
  entry_date: string;
  description: string;
  source_type: string;
  total_debit: string;
  total_credit: string;
  status: 'DRAFT' | 'POSTED' | 'REVERSED' | 'CANCELLED';
  lines: IntegratedJournalLine[];
  created_at: string;
}

export interface IntegratedJournalLine {
  id: number;
  account_id: number;
  account_code: string;
  account_name: string;
  description: string;
  debit_amount: string;
  credit_amount: string;
}

export interface IntegratedAccountResponse {
  account: IntegratedAccountDetail;
  recent_transactions: IntegratedTransaction[];
  recent_journal_entries: IntegratedJournalEntry[];
  last_synced_at: string;
}

export interface IntegratedBalanceSummary {
  total_cash: string;
  total_bank: string;
  total_balance: string;
  total_ssot_balance: string;
  balance_variance: string;
  variance_count: number;
}

export interface IntegratedAccountSummary {
  id: number;
  name: string;
  code: string;
  type: 'CASH' | 'BANK';
  balance: string;
  ssot_balance: string;
  variance: string;
  last_transaction_date?: string;
  total_journal_entries: number;
  reconciliation_status: 'MATCHED' | 'MINOR_VARIANCE' | 'VARIANCE';
}

export interface IntegratedActivity {
  type: string;
  id: number;
  number: string;
  description: string;
  amount: string;
  account_name: string;
  created_at: string;
}

export interface SyncStatus {
  last_sync_at: string;
  sync_status: string;
  total_accounts: number;
  synced_accounts: number;
  variance_accounts: number;
}

export interface IntegratedSummaryResponse {
  summary: IntegratedBalanceSummary;
  accounts: IntegratedAccountSummary[];
  recent_activities: IntegratedActivity[];
  sync_status: SyncStatus;
}

export interface ReconciliationData {
  account_id: number;
  account_name: string;
  cashbank_balance: string;
  ssot_balance: string;
  difference: string;
  has_discrepancy: boolean;
  reconciliation_status: ReconciliationStatus;
  last_reconciled_at: string;
  details: string[];
  recommendations: string[];
}

export interface JournalEntriesResponse {
  entries: IntegratedJournalEntry[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}

export interface TransactionHistoryResponse {
  transactions: IntegratedTransaction[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}

// API Response wrapper
export interface ApiResponse<T> {
  status: 'success' | 'error';
  data: T;
  message?: string;
}

// Additional types for account detail page
export interface JournalEntryDetail {
  id: number;
  entry_id: number;
  line_id: number;
  entry_date: string;
  description: string;
  debit_amount: string;
  credit_amount: string;
  status: JournalStatus;
}

export interface TransactionEntry {
  id: number;
  number: string;
  type: string;
  amount: string;
  description: string;
  reference_number: string;
  created_at: string;
}

// Utility types
export type ReconciliationStatus = 'MATCHED' | 'MINOR_VARIANCE' | 'VARIANCE';
export type AccountType = 'CASH' | 'BANK';
export type JournalStatus = 'DRAFT' | 'POSTED' | 'REVERSED' | 'CANCELLED';
