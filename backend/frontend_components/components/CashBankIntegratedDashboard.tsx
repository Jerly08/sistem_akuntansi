// CashBank SSOT Integrated Dashboard Component
// Main dashboard component for CashBank integration with SSOT Journal system

import React, { useState, useEffect } from 'react';
import { 
  IntegratedSummaryResponse, 
  IntegratedAccountSummary,
  ReconciliationStatus
} from '../types/cashBankIntegration.types';
import { cashBankIntegrationService } from '../services/cashBankIntegrationService';
import SummaryStatCard from './SummaryStatCard';

interface CashBankIntegratedDashboardProps {
  onAccountSelect?: (accountId: number) => void;
}

export const CashBankIntegratedDashboard: React.FC<CashBankIntegratedDashboardProps> = ({
  onAccountSelect
}) => {
  const [summary, setSummary] = useState<IntegratedSummaryResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);

  // Load integrated summary data
  const loadSummary = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await cashBankIntegrationService.getIntegratedSummary();
      setSummary(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load summary');
    } finally {
      setLoading(false);
    }
  };

  // Refresh data
  const handleRefresh = async () => {
    setRefreshing(true);
    await loadSummary();
    setRefreshing(false);
  };

  // Load data on component mount
  useEffect(() => {
    loadSummary();
  }, []);

  // Auto-refresh every 5 minutes
  useEffect(() => {
    const interval = setInterval(loadSummary, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        <span className="ml-3 text-gray-600">Memuat data integrasi...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4">
        <div className="flex items-center">
          <div className="text-red-400">
            <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
            </svg>
          </div>
          <div className="ml-3">
            <p className="text-red-800 font-medium">Gagal memuat data</p>
            <p className="text-red-700">{error}</p>
          </div>
        </div>
        <button
          onClick={() => loadSummary()}
          className="mt-3 bg-red-100 hover:bg-red-200 text-red-800 px-4 py-2 rounded-md text-sm font-medium"
        >
          Coba Lagi
        </button>
      </div>
    );
  }

  if (!summary) {
    return (
      <div className="text-center py-8">
        <p className="text-gray-500">Tidak ada data untuk ditampilkan</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header with refresh button */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Cash & Bank Terintegrasi</h1>
          <p className="text-gray-600">Dashboard terpadu Cash/Bank dengan SSOT Journal</p>
        </div>
        <button
          onClick={handleRefresh}
          disabled={refreshing}
          className="bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white px-4 py-2 rounded-lg flex items-center space-x-2"
        >
          <svg 
            className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} 
            fill="none" 
            stroke="currentColor" 
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <span>{refreshing ? 'Memperbarui...' : 'Refresh'}</span>
        </button>
      </div>

      {/* Summary Cards - colored like Sales Management */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <SummaryStatCard
          title="Total Cash"
          value={cashBankIntegrationService.formatCurrency(summary.summary.total_cash)}
          icon={<span className="text-2xl">💵</span>}
          color="green"
        />
        <SummaryStatCard
          title="Total Bank"
          value={cashBankIntegrationService.formatCurrency(summary.summary.total_bank)}
          icon={<span className="text-2xl">🏦</span>}
          color="blue"
        />
        <SummaryStatCard
          title="Total Balance"
          value={cashBankIntegrationService.formatCurrency(summary.summary.total_balance)}
          icon={<span className="text-2xl">💰</span>}
          color="purple"
        />
        <SummaryStatCard
          title="SSOT Balance"
          value={cashBankIntegrationService.formatCurrency(summary.summary.total_ssot_balance)}
          icon={<span className="text-2xl">📊</span>}
          color="indigo"
        />
      </div>

      {/* Variance Alert */}
      {summary.summary.variance_count > 0 && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <div className="flex items-center">
            <div className="text-yellow-400">
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <p className="text-yellow-800 font-medium">
                Peringatan: {summary.summary.variance_count} akun memiliki selisih balance
              </p>
              <p className="text-yellow-700">
                Total selisih: {cashBankIntegrationService.formatCurrency(summary.summary.balance_variance)}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Sync Status */}
      <div className="bg-white border border-gray-200 rounded-lg p-4">
        <h3 className="text-lg font-semibold text-gray-900 mb-3">Status Sinkronisasi</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-blue-600">{summary.sync_status.total_accounts}</div>
            <div className="text-sm text-gray-600">Total Akun</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-green-600">{summary.sync_status.synced_accounts}</div>
            <div className="text-sm text-gray-600">Tersinkron</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-red-600">{summary.sync_status.variance_accounts}</div>
            <div className="text-sm text-gray-600">Ada Selisih</div>
          </div>
        </div>
        <div className="mt-3 text-sm text-gray-600 text-center">
          Terakhir disinkron: {cashBankIntegrationService.formatDateTime(summary.sync_status.last_sync_at)}
        </div>
      </div>

      {/* Accounts List */}
      <div className="bg-white border border-gray-200 rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900">Akun Cash & Bank</h3>
        </div>
        <div className="divide-y divide-gray-200">
          {summary.accounts.map((account) => (
            <AccountCard
              key={account.id}
              account={account}
              onSelect={onAccountSelect}
            />
          ))}
        </div>
      </div>

      {/* Recent Activities */}
      {summary.recent_activities.length > 0 && (
        <div className="bg-white border border-gray-200 rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-semibold text-gray-900">Aktivitas Terbaru</h3>
          </div>
          <div className="divide-y divide-gray-200">
            {summary.recent_activities.slice(0, 5).map((activity, index) => (
              <div key={index} className="px-6 py-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium text-gray-900">{activity.description}</p>
                    <p className="text-sm text-gray-600">
                      {activity.account_name} • {activity.number}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="font-medium text-gray-900">
                      {cashBankIntegrationService.formatCurrency(activity.amount)}
                    </p>
                    <p className="text-sm text-gray-600">
                      {cashBankIntegrationService.formatDate(activity.created_at)}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};


// Account Card Component
interface AccountCardProps {
  account: IntegratedAccountSummary;
  onSelect?: (accountId: number) => void;
}

const AccountCard: React.FC<AccountCardProps> = ({ account, onSelect }) => {
  const hasVariance = cashBankIntegrationService.hasVariance(account.variance);
  const statusColor = cashBankIntegrationService.getReconciliationStatusColor(account.reconciliation_status);
  const statusLabel = cashBankIntegrationService.getReconciliationStatusLabel(account.reconciliation_status);

  return (
    <div className="px-6 py-4 hover:bg-gray-50">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <div className="text-2xl">
            {cashBankIntegrationService.getAccountTypeIcon(account.type)}
          </div>
          <div>
            <div className="flex items-center space-x-2">
              <h4 className="font-semibold text-gray-900">{account.name}</h4>
              <span className={`px-2 py-1 rounded-full text-xs font-medium border ${statusColor}`}>
                {statusLabel}
              </span>
            </div>
            <p className="text-sm text-gray-600">{account.code} • {account.type}</p>
            <p className="text-xs text-gray-500">
              {account.total_journal_entries} jurnal entries
              {account.last_transaction_date && (
                <> • Transaksi terakhir: {cashBankIntegrationService.formatDate(account.last_transaction_date)}</>
              )}
            </p>
          </div>
        </div>
        
        <div className="text-right">
          <div className="font-semibold text-gray-900">
            {cashBankIntegrationService.formatCurrency(account.balance)}
          </div>
          <div className="text-sm text-gray-600">
            SSOT: {cashBankIntegrationService.formatCurrency(account.ssot_balance)}
          </div>
          {hasVariance && (
            <div className="text-sm text-red-600 font-medium">
              Selisih: {cashBankIntegrationService.formatCurrency(account.variance)}
            </div>
          )}
        </div>
      </div>
      
      {onSelect && (
        <div className="mt-3">
          <button
            onClick={() => onSelect(account.id)}
            className="text-blue-600 hover:text-blue-800 text-sm font-medium"
          >
            Lihat Detail →
          </button>
        </div>
      )}
    </div>
  );
};

export default CashBankIntegratedDashboard;