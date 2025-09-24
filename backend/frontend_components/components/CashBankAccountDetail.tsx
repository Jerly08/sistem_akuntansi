// CashBank Account Detail Component
// Detailed view for individual CashBank account with SSOT integration

import React, { useState, useEffect } from 'react';
import {
  IntegratedAccountResponse,
  ReconciliationData,
  JournalEntriesResponse,
  TransactionHistoryResponse,
  TransactionEntry,
  JournalEntryDetail
} from '../types/cashBankIntegration.types';
import { cashBankIntegrationService } from '../services/cashBankIntegrationService';

interface CashBankAccountDetailProps {
  accountId: number;
  onBack?: () => void;
}

type TabType = 'overview' | 'transactions' | 'journal' | 'reconciliation';

export const CashBankAccountDetail: React.FC<CashBankAccountDetailProps> = ({
  accountId,
  onBack
}) => {
  const [account, setAccount] = useState<IntegratedAccountResponse | null>(null);
  const [reconciliation, setReconciliation] = useState<ReconciliationData | null>(null);
  const [journalEntries, setJournalEntries] = useState<JournalEntriesResponse | null>(null);
  const [transactions, setTransactions] = useState<TransactionHistoryResponse | null>(null);
  
  const [activeTab, setActiveTab] = useState<TabType>('overview');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Pagination states
  const [journalPage, setJournalPage] = useState(1);
  const [transactionPage, setTransactionPage] = useState(1);
  const pageSize = 20;

  // Load account data
  const loadAccountData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const [accountData, reconciliationData] = await Promise.all([
        cashBankIntegrationService.getIntegratedAccount(accountId),
        cashBankIntegrationService.getReconciliation(accountId)
      ]);
      
      setAccount(accountData);
      setReconciliation(reconciliationData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load account data');
    } finally {
      setLoading(false);
    }
  };

  // Load journal entries
  const loadJournalEntries = async (page: number = 1) => {
    try {
      const data = await cashBankIntegrationService.getJournalEntries(accountId, page, pageSize);
      setJournalEntries(data);
    } catch (err) {
      console.error('Failed to load journal entries:', err);
    }
  };

  // Load transaction history
  const loadTransactions = async (page: number = 1) => {
    try {
      const data = await cashBankIntegrationService.getTransactionHistory(accountId, page, pageSize);
      setTransactions(data);
    } catch (err) {
      console.error('Failed to load transactions:', err);
    }
  };

  // Tab change handler
  const handleTabChange = (tab: TabType) => {
    setActiveTab(tab);
    
    if (tab === 'journal' && !journalEntries) {
      loadJournalEntries(1);
      setJournalPage(1);
    } else if (tab === 'transactions' && !transactions) {
      loadTransactions(1);
      setTransactionPage(1);
    }
  };

  // Pagination handlers
  const handleJournalPageChange = (page: number) => {
    setJournalPage(page);
    loadJournalEntries(page);
  };

  const handleTransactionPageChange = (page: number) => {
    setTransactionPage(page);
    loadTransactions(page);
  };

  // Load data on mount
  useEffect(() => {
    loadAccountData();
  }, [accountId]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        <span className="ml-3 text-gray-600">Memuat detail akun...</span>
      </div>
    );
  }

  if (error || !account) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4">
        <div className="flex items-center">
          <div className="text-red-400">
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
            </svg>
          </div>
          <div className="ml-3">
            <p className="text-red-800 font-medium">Gagal memuat detail akun</p>
            <p className="text-red-700">{error}</p>
          </div>
        </div>
        {onBack && (
          <button
            onClick={onBack}
            className="mt-3 bg-red-100 hover:bg-red-200 text-red-800 px-4 py-2 rounded-md text-sm font-medium"
          >
            ← Kembali
          </button>
        )}
      </div>
    );
  }

  const hasVariance = cashBankIntegrationService.hasVariance(account.account.variance);
  const statusColor = cashBankIntegrationService.getReconciliationStatusColor(account.account.reconciliation_status);
  const statusLabel = cashBankIntegrationService.getReconciliationStatusLabel(account.account.reconciliation_status);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          {onBack && (
            <button
              onClick={onBack}
              className="text-gray-600 hover:text-gray-900"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
            </button>
          )}
          <div>
            <div className="flex items-center space-x-3">
              <div className="text-3xl">
                {cashBankIntegrationService.getAccountTypeIcon(account.account.type)}
              </div>
              <div>
                <h1 className="text-2xl font-bold text-gray-900">{account.account.name}</h1>
                <p className="text-gray-600">{account.account.code} • {account.account.type}</p>
              </div>
            </div>
          </div>
        </div>
        
        <div className="text-right">
          <div className="flex items-center space-x-2 mb-1">
            <span className={`px-3 py-1 rounded-full text-sm font-medium border ${statusColor}`}>
              {statusLabel}
            </span>
          </div>
          <div className="text-sm text-gray-600">
            ID: {account.account.id} • Aktif: {account.account.is_active ? 'Ya' : 'Tidak'}
          </div>
        </div>
      </div>

      {/* Balance Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <div className="text-sm text-gray-600 mb-1">Saldo CashBank</div>
          <div className="text-2xl font-bold text-gray-900">
            {cashBankIntegrationService.formatCurrency(account.account.balance)}
          </div>
        </div>
        
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <div className="text-sm text-gray-600 mb-1">Saldo SSOT</div>
          <div className="text-2xl font-bold text-gray-900">
            {cashBankIntegrationService.formatCurrency(account.account.ssot_balance)}
          </div>
        </div>
        
        <div className={`border rounded-lg p-4 ${hasVariance ? 'bg-red-50 border-red-200' : 'bg-green-50 border-green-200'}`}>
          <div className="text-sm text-gray-600 mb-1">Selisih Balance</div>
          <div className={`text-2xl font-bold ${hasVariance ? 'text-red-600' : 'text-green-600'}`}>
            {cashBankIntegrationService.formatCurrency(account.account.variance)}
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          {[
            { id: 'overview', label: 'Overview', icon: '📊' },
            { id: 'transactions', label: 'Transaksi', icon: '💳' },
            { id: 'journal', label: 'Jurnal', icon: '📝' },
            { id: 'reconciliation', label: 'Rekonsiliasi', icon: '🔍' }
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => handleTabChange(tab.id as TabType)}
              className={`py-2 px-1 border-b-2 font-medium text-sm flex items-center space-x-2 ${
                activeTab === tab.id
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              <span>{tab.icon}</span>
              <span>{tab.label}</span>
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'overview' && (
        <OverviewTab account={account} reconciliation={reconciliation} />
      )}

      {activeTab === 'transactions' && (
        <TransactionsTab
          accountId={accountId}
          transactions={transactions}
          currentPage={transactionPage}
          pageSize={pageSize}
          onPageChange={handleTransactionPageChange}
        />
      )}

      {activeTab === 'journal' && (
        <JournalTab
          accountId={accountId}
          journalEntries={journalEntries}
          currentPage={journalPage}
          pageSize={pageSize}
          onPageChange={handleJournalPageChange}
        />
      )}

      {activeTab === 'reconciliation' && reconciliation && (
        <ReconciliationTab reconciliation={reconciliation} />
      )}
    </div>
  );
};

// Overview Tab Component
interface OverviewTabProps {
  account: IntegratedAccountResponse;
  reconciliation: ReconciliationData | null;
}

const OverviewTab: React.FC<OverviewTabProps> = ({ account, reconciliation }) => {
  return (
    <div className="space-y-6">
      {/* Account Info */}
      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Informasi Akun</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <div className="space-y-3">
              <div>
                <label className="text-sm font-medium text-gray-500">Nama Akun</label>
                <p className="text-gray-900">{account.account.name}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-500">Kode Akun</label>
                <p className="text-gray-900">{account.account.code}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-500">Tipe</label>
                <p className="text-gray-900">{account.account.type}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-500">Status</label>
                <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                  account.account.is_active 
                    ? 'bg-green-100 text-green-800' 
                    : 'bg-red-100 text-red-800'
                }`}>
                  {account.account.is_active ? 'Aktif' : 'Tidak Aktif'}
                </span>
              </div>
            </div>
          </div>
          <div>
            <div className="space-y-3">
              <div>
                <label className="text-sm font-medium text-gray-500">Total Jurnal Entries</label>
                <p className="text-gray-900">{account.account.total_journal_entries}</p>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-500">Transaksi Terakhir</label>
                <p className="text-gray-900">
                  {account.account.last_transaction_date 
                    ? cashBankIntegrationService.formatDate(account.account.last_transaction_date)
                    : 'Tidak ada'
                  }
                </p>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-500">Dibuat</label>
                <p className="text-gray-900">
                  {cashBankIntegrationService.formatDateTime(account.account.created_at)}
                </p>
              </div>
              <div>
                <label className="text-sm font-medium text-gray-500">Diperbarui</label>
                <p className="text-gray-900">
                  {cashBankIntegrationService.formatDateTime(account.account.updated_at)}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Recent Transactions */}
      {account.recent_transactions.length > 0 && (
        <div className="bg-white border border-gray-200 rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-semibold text-gray-900">Transaksi Terbaru</h3>
          </div>
          <div className="divide-y divide-gray-200">
            {account.recent_transactions.slice(0, 5).map((transaction, index) => (
              <TransactionRow key={index} transaction={transaction} />
            ))}
          </div>
        </div>
      )}

      {/* Recent Journal Entries */}
      {account.recent_journal_entries.length > 0 && (
        <div className="bg-white border border-gray-200 rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-semibold text-gray-900">Jurnal Entries Terbaru</h3>
          </div>
          <div className="divide-y divide-gray-200">
            {account.recent_journal_entries.slice(0, 5).map((entry, index) => (
              <JournalRow key={index} entry={entry} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

// Transactions Tab Component
interface TransactionsTabProps {
  accountId: number;
  transactions: TransactionHistoryResponse | null;
  currentPage: number;
  pageSize: number;
  onPageChange: (page: number) => void;
}

const TransactionsTab: React.FC<TransactionsTabProps> = ({
  transactions,
  currentPage,
  onPageChange
}) => {
  if (!transactions) {
    return (
      <div className="flex items-center justify-center h-32">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <span className="ml-3 text-gray-600">Memuat transaksi...</span>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="bg-white border border-gray-200 rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold text-gray-900">Riwayat Transaksi</h3>
            <span className="text-sm text-gray-600">
              Total: {transactions.pagination.total} transaksi
            </span>
          </div>
        </div>
        <div className="divide-y divide-gray-200">
          {transactions.transactions.map((transaction) => (
            <TransactionRow key={transaction.id} transaction={transaction} showDetails />
          ))}
        </div>
      </div>

      {/* Pagination */}
      <PaginationControls
        currentPage={currentPage}
        totalPages={transactions.pagination.total_pages}
        onPageChange={onPageChange}
      />
    </div>
  );
};

// Journal Tab Component
interface JournalTabProps {
  accountId: number;
  journalEntries: JournalEntriesResponse | null;
  currentPage: number;
  pageSize: number;
  onPageChange: (page: number) => void;
}

const JournalTab: React.FC<JournalTabProps> = ({
  journalEntries,
  currentPage,
  onPageChange
}) => {
  if (!journalEntries) {
    return (
      <div className="flex items-center justify-center h-32">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <span className="ml-3 text-gray-600">Memuat jurnal entries...</span>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="bg-white border border-gray-200 rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold text-gray-900">Jurnal Entries</h3>
            <span className="text-sm text-gray-600">
              Total: {journalEntries.pagination.total} entries
            </span>
          </div>
        </div>
        <div className="divide-y divide-gray-200">
          {journalEntries.entries.map((entry) => (
            <JournalRow key={entry.id} entry={entry} showDetails />
          ))}
        </div>
      </div>

      {/* Pagination */}
      <PaginationControls
        currentPage={currentPage}
        totalPages={journalEntries.pagination.total_pages}
        onPageChange={onPageChange}
      />
    </div>
  );
};

// Reconciliation Tab Component
interface ReconciliationTabProps {
  reconciliation: ReconciliationData;
}

const ReconciliationTab: React.FC<ReconciliationTabProps> = ({ reconciliation }) => {
  const hasDiscrepancy = reconciliation.has_discrepancy;
  
  return (
    <div className="space-y-6">
      {/* Reconciliation Status */}
      <div className={`border rounded-lg p-6 ${
        hasDiscrepancy ? 'bg-red-50 border-red-200' : 'bg-green-50 border-green-200'
      }`}>
        <div className="flex items-center space-x-3 mb-4">
          <div className={`text-2xl ${hasDiscrepancy ? 'text-red-600' : 'text-green-600'}`}>
            {hasDiscrepancy ? '⚠️' : '✅'}
          </div>
          <div>
            <h3 className={`text-lg font-semibold ${hasDiscrepancy ? 'text-red-800' : 'text-green-800'}`}>
              Status Rekonsiliasi: {hasDiscrepancy ? 'Ada Selisih' : 'Sesuai'}
            </h3>
            <p className={`text-sm ${hasDiscrepancy ? 'text-red-700' : 'text-green-700'}`}>
              Terakhir direkonsiliasi: {cashBankIntegrationService.formatDateTime(reconciliation.last_reconciled_at)}
            </p>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="text-sm font-medium text-gray-500">CashBank Balance</label>
            <p className="text-lg font-semibold text-gray-900">
              {cashBankIntegrationService.formatCurrency(reconciliation.cashbank_balance)}
            </p>
          </div>
          <div>
            <label className="text-sm font-medium text-gray-500">SSOT Balance</label>
            <p className="text-lg font-semibold text-gray-900">
              {cashBankIntegrationService.formatCurrency(reconciliation.ssot_balance)}
            </p>
          </div>
          <div>
            <label className="text-sm font-medium text-gray-500">Selisih</label>
            <p className={`text-lg font-semibold ${hasDiscrepancy ? 'text-red-600' : 'text-green-600'}`}>
              {cashBankIntegrationService.formatCurrency(reconciliation.difference)}
            </p>
          </div>
        </div>
      </div>

      {/* Reconciliation Details */}
      {reconciliation.details && reconciliation.details.length > 0 && (
        <div className="bg-white border border-gray-200 rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-semibold text-gray-900">Detail Rekonsiliasi</h3>
          </div>
          <div className="p-6">
            <div className="space-y-3">
              {reconciliation.details.map((detail, index) => (
                <div key={index} className="flex items-center justify-between py-2">
                  <span className="text-gray-700">{detail}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// Transaction Row Component
interface TransactionRowProps {
  transaction: TransactionEntry;
  showDetails?: boolean;
}

const TransactionRow: React.FC<TransactionRowProps> = ({ transaction, showDetails = false }) => (
  <div className="px-6 py-4">
    <div className="flex items-center justify-between">
      <div>
        <p className="font-medium text-gray-900">{transaction.description}</p>
        {showDetails && (
          <p className="text-sm text-gray-600">
            {transaction.number} • {transaction.type}
          </p>
        )}
        <p className="text-sm text-gray-500">
          {cashBankIntegrationService.formatDateTime(transaction.created_at)}
        </p>
      </div>
      <div className="text-right">
        <p className={`font-medium ${
          transaction.amount >= 0 ? 'text-green-600' : 'text-red-600'
        }`}>
          {cashBankIntegrationService.formatCurrency(transaction.amount)}
        </p>
        {showDetails && (
          <p className="text-sm text-gray-600">
            {transaction.reference_number}
          </p>
        )}
      </div>
    </div>
  </div>
);

// Journal Row Component
interface JournalRowProps {
  entry: JournalEntryDetail;
  showDetails?: boolean;
}

const JournalRow: React.FC<JournalRowProps> = ({ entry, showDetails = false }) => (
  <div className="px-6 py-4">
    <div className="flex items-center justify-between">
      <div>
        <p className="font-medium text-gray-900">{entry.description}</p>
        {showDetails && (
          <p className="text-sm text-gray-600">
            Entry ID: {entry.entry_id} • Line ID: {entry.line_id}
          </p>
        )}
        <p className="text-sm text-gray-500">
          {cashBankIntegrationService.formatDate(entry.entry_date)}
        </p>
      </div>
      <div className="text-right">
        <div className="space-y-1">
          {entry.debit_amount > 0 && (
            <p className="text-sm text-green-600">
              Debit: {cashBankIntegrationService.formatCurrency(entry.debit_amount)}
            </p>
          )}
          {entry.credit_amount > 0 && (
            <p className="text-sm text-red-600">
              Credit: {cashBankIntegrationService.formatCurrency(entry.credit_amount)}
            </p>
          )}
        </div>
        {showDetails && (
          <p className="text-xs text-gray-500 mt-1">
            {entry.status}
          </p>
        )}
      </div>
    </div>
  </div>
);

// Pagination Controls Component
interface PaginationControlsProps {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

const PaginationControls: React.FC<PaginationControlsProps> = ({
  currentPage,
  totalPages,
  onPageChange
}) => {
  if (totalPages <= 1) return null;

  const getPageNumbers = () => {
    const pages = [];
    const maxVisible = 5;
    let start = Math.max(1, currentPage - Math.floor(maxVisible / 2));
    let end = Math.min(totalPages, start + maxVisible - 1);
    
    if (end - start + 1 < maxVisible) {
      start = Math.max(1, end - maxVisible + 1);
    }
    
    for (let i = start; i <= end; i++) {
      pages.push(i);
    }
    
    return pages;
  };

  return (
    <div className="flex items-center justify-between">
      <p className="text-sm text-gray-700">
        Halaman <span className="font-medium">{currentPage}</span> dari{' '}
        <span className="font-medium">{totalPages}</span>
      </p>
      
      <div className="flex items-center space-x-2">
        <button
          onClick={() => onPageChange(currentPage - 1)}
          disabled={currentPage === 1}
          className="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          ← Sebelumnya
        </button>
        
        {getPageNumbers().map((page) => (
          <button
            key={page}
            onClick={() => onPageChange(page)}
            className={`px-3 py-2 text-sm font-medium rounded-md ${
              page === currentPage
                ? 'bg-blue-600 text-white'
                : 'text-gray-700 bg-white border border-gray-300 hover:bg-gray-50'
            }`}
          >
            {page}
          </button>
        ))}
        
        <button
          onClick={() => onPageChange(currentPage + 1)}
          disabled={currentPage === totalPages}
          className="px-3 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Selanjutnya →
        </button>
      </div>
    </div>
  );
};

export default CashBankAccountDetail;