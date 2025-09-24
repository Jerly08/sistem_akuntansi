// useCashBankIntegration Hook
// Custom React hook for managing CashBank-SSOT integration state and operations

import { useState, useEffect, useCallback, useRef } from 'react';
import {
  IntegratedSummaryResponse,
  IntegratedAccountResponse,
  ReconciliationData,
  JournalEntriesResponse,
  TransactionHistoryResponse,
  IntegratedAccountSummary
} from '../types/cashBankIntegration.types';
import { cashBankIntegrationService } from '../services/cashBankIntegrationService';

interface CashBankIntegrationState {
  // Summary data
  summary: IntegratedSummaryResponse | null;
  summaryLoading: boolean;
  summaryError: string | null;
  
  // Account data cache
  accountsCache: Map<number, IntegratedAccountResponse>;
  reconciliationCache: Map<number, ReconciliationData>;
  
  // Loading states
  accountLoading: Map<number, boolean>;
  reconciliationLoading: Map<number, boolean>;
  
  // Errors
  accountErrors: Map<number, string>;
  reconciliationErrors: Map<number, string>;
  
  // Refresh states
  isRefreshing: boolean;
  lastRefresh: Date | null;
  
  // Real-time updates
  autoRefreshEnabled: boolean;
  refreshInterval: number; // in milliseconds
}

interface UseCashBankIntegrationOptions {
  autoRefresh?: boolean;
  refreshInterval?: number;
  enableCaching?: boolean;
  cacheTimeout?: number; // in milliseconds
}

interface UseCashBankIntegrationReturn {
  // State
  state: CashBankIntegrationState;
  
  // Summary operations
  loadSummary: () => Promise<void>;
  refreshSummary: () => Promise<void>;
  
  // Account operations
  loadAccount: (accountId: number, forceRefresh?: boolean) => Promise<IntegratedAccountResponse | null>;
  loadReconciliation: (accountId: number, forceRefresh?: boolean) => Promise<ReconciliationData | null>;
  
  // Journal & Transaction operations
  loadJournalEntries: (accountId: number, page?: number, pageSize?: number) => Promise<JournalEntriesResponse>;
  loadTransactionHistory: (accountId: number, page?: number, pageSize?: number) => Promise<TransactionHistoryResponse>;
  
  // Cache operations
  clearCache: () => void;
  clearAccountCache: (accountId: number) => void;
  
  // Refresh operations
  refreshAll: () => Promise<void>;
  toggleAutoRefresh: () => void;
  
  // Utility functions
  getAccountFromCache: (accountId: number) => IntegratedAccountResponse | null;
  getReconciliationFromCache: (accountId: number) => ReconciliationData | null;
  hasVarianceAccounts: () => boolean;
  getTotalVariance: () => number;
  getAccountsWithVariance: () => IntegratedAccountSummary[];
}

export const useCashBankIntegration = (
  options: UseCashBankIntegrationOptions = {}
): UseCashBankIntegrationReturn => {
  const {
    autoRefresh = true,
    refreshInterval = 5 * 60 * 1000, // 5 minutes
    enableCaching = true,
    cacheTimeout = 10 * 60 * 1000 // 10 minutes
  } = options;

  // State initialization
  const [state, setState] = useState<CashBankIntegrationState>({
    summary: null,
    summaryLoading: true,
    summaryError: null,
    accountsCache: new Map(),
    reconciliationCache: new Map(),
    accountLoading: new Map(),
    reconciliationLoading: new Map(),
    accountErrors: new Map(),
    reconciliationErrors: new Map(),
    isRefreshing: false,
    lastRefresh: null,
    autoRefreshEnabled: autoRefresh,
    refreshInterval
  });

  // Refs for cleanup
  const refreshIntervalRef = useRef<NodeJS.Timeout>();
  const abortControllerRef = useRef<AbortController>();

  // Cache timestamps for TTL
  const cacheTimestamps = useRef<Map<string, Date>>(new Map());

  // Helper function to check cache validity
  const isCacheValid = (key: string): boolean => {
    if (!enableCaching) return false;
    const timestamp = cacheTimestamps.current.get(key);
    if (!timestamp) return false;
    return Date.now() - timestamp.getTime() < cacheTimeout;
  };

  // Helper function to set cache timestamp
  const setCacheTimestamp = (key: string) => {
    cacheTimestamps.current.set(key, new Date());
  };

  // Load summary data
  const loadSummary = useCallback(async () => {
    if (!isCacheValid('summary')) {
      setState(prev => ({ ...prev, summaryLoading: true, summaryError: null }));
    }

    try {
      const summary = await cashBankIntegrationService.getIntegratedSummary();
      setState(prev => ({
        ...prev,
        summary,
        summaryLoading: false,
        summaryError: null,
        lastRefresh: new Date()
      }));
      setCacheTimestamp('summary');
    } catch (error) {
      setState(prev => ({
        ...prev,
        summaryLoading: false,
        summaryError: error instanceof Error ? error.message : 'Failed to load summary'
      }));
    }
  }, [enableCaching, cacheTimeout]);

  // Refresh summary (always fresh)
  const refreshSummary = useCallback(async () => {
    setState(prev => ({ ...prev, isRefreshing: true }));
    
    // Clear cache for fresh data
    cacheTimestamps.current.delete('summary');
    
    await loadSummary();
    
    setState(prev => ({ ...prev, isRefreshing: false }));
  }, [loadSummary]);

  // Load account data
  const loadAccount = useCallback(async (
    accountId: number, 
    forceRefresh = false
  ): Promise<IntegratedAccountResponse | null> => {
    const cacheKey = `account-${accountId}`;
    
    // Return cached data if valid
    if (!forceRefresh && isCacheValid(cacheKey)) {
      const cached = state.accountsCache.get(accountId);
      if (cached) return cached;
    }

    // Set loading state
    setState(prev => ({
      ...prev,
      accountLoading: new Map(prev.accountLoading).set(accountId, true),
      accountErrors: new Map(prev.accountErrors.set(accountId, ''))
    }));

    try {
      const account = await cashBankIntegrationService.getIntegratedAccount(accountId);
      
      setState(prev => ({
        ...prev,
        accountsCache: new Map(prev.accountsCache).set(accountId, account),
        accountLoading: new Map(prev.accountLoading).set(accountId, false)
      }));
      
      setCacheTimestamp(cacheKey);
      return account;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to load account';
      
      setState(prev => ({
        ...prev,
        accountLoading: new Map(prev.accountLoading).set(accountId, false),
        accountErrors: new Map(prev.accountErrors).set(accountId, errorMessage)
      }));
      
      return null;
    }
  }, [state.accountsCache, enableCaching, cacheTimeout]);

  // Load reconciliation data
  const loadReconciliation = useCallback(async (
    accountId: number, 
    forceRefresh = false
  ): Promise<ReconciliationData | null> => {
    const cacheKey = `reconciliation-${accountId}`;
    
    // Return cached data if valid
    if (!forceRefresh && isCacheValid(cacheKey)) {
      const cached = state.reconciliationCache.get(accountId);
      if (cached) return cached;
    }

    // Set loading state
    setState(prev => ({
      ...prev,
      reconciliationLoading: new Map(prev.reconciliationLoading).set(accountId, true),
      reconciliationErrors: new Map(prev.reconciliationErrors.set(accountId, ''))
    }));

    try {
      const reconciliation = await cashBankIntegrationService.getReconciliation(accountId);
      
      setState(prev => ({
        ...prev,
        reconciliationCache: new Map(prev.reconciliationCache).set(accountId, reconciliation),
        reconciliationLoading: new Map(prev.reconciliationLoading).set(accountId, false)
      }));
      
      setCacheTimestamp(cacheKey);
      return reconciliation;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to load reconciliation';
      
      setState(prev => ({
        ...prev,
        reconciliationLoading: new Map(prev.reconciliationLoading).set(accountId, false),
        reconciliationErrors: new Map(prev.reconciliationErrors).set(accountId, errorMessage)
      }));
      
      return null;
    }
  }, [state.reconciliationCache, enableCaching, cacheTimeout]);

  // Load journal entries (no caching for paginated data)
  const loadJournalEntries = useCallback(async (
    accountId: number,
    page = 1,
    pageSize = 20
  ): Promise<JournalEntriesResponse> => {
    return await cashBankIntegrationService.getJournalEntries(accountId, page, pageSize);
  }, []);

  // Load transaction history (no caching for paginated data)
  const loadTransactionHistory = useCallback(async (
    accountId: number,
    page = 1,
    pageSize = 20
  ): Promise<TransactionHistoryResponse> => {
    return await cashBankIntegrationService.getTransactionHistory(accountId, page, pageSize);
  }, []);

  // Clear all caches
  const clearCache = useCallback(() => {
    setState(prev => ({
      ...prev,
      accountsCache: new Map(),
      reconciliationCache: new Map(),
      accountErrors: new Map(),
      reconciliationErrors: new Map()
    }));
    cacheTimestamps.current.clear();
  }, []);

  // Clear specific account cache
  const clearAccountCache = useCallback((accountId: number) => {
    setState(prev => {
      const newAccountsCache = new Map(prev.accountsCache);
      const newReconciliationCache = new Map(prev.reconciliationCache);
      const newAccountErrors = new Map(prev.accountErrors);
      const newReconciliationErrors = new Map(prev.reconciliationErrors);
      
      newAccountsCache.delete(accountId);
      newReconciliationCache.delete(accountId);
      newAccountErrors.delete(accountId);
      newReconciliationErrors.delete(accountId);
      
      return {
        ...prev,
        accountsCache: newAccountsCache,
        reconciliationCache: newReconciliationCache,
        accountErrors: newAccountErrors,
        reconciliationErrors: newReconciliationErrors
      };
    });
    
    cacheTimestamps.current.delete(`account-${accountId}`);
    cacheTimestamps.current.delete(`reconciliation-${accountId}`);
  }, []);

  // Refresh all data
  const refreshAll = useCallback(async () => {
    setState(prev => ({ ...prev, isRefreshing: true }));
    
    // Clear all cache
    clearCache();
    
    // Reload summary
    await loadSummary();
    
    setState(prev => ({ ...prev, isRefreshing: false }));
  }, [clearCache, loadSummary]);

  // Toggle auto refresh
  const toggleAutoRefresh = useCallback(() => {
    setState(prev => ({
      ...prev,
      autoRefreshEnabled: !prev.autoRefreshEnabled
    }));
  }, []);

  // Utility functions
  const getAccountFromCache = useCallback((accountId: number): IntegratedAccountResponse | null => {
    return state.accountsCache.get(accountId) || null;
  }, [state.accountsCache]);

  const getReconciliationFromCache = useCallback((accountId: number): ReconciliationData | null => {
    return state.reconciliationCache.get(accountId) || null;
  }, [state.reconciliationCache]);

  const hasVarianceAccounts = useCallback((): boolean => {
    return state.summary?.summary.variance_count > 0 || false;
  }, [state.summary]);

  const getTotalVariance = useCallback((): number => {
    return state.summary?.summary.balance_variance || 0;
  }, [state.summary]);

  const getAccountsWithVariance = useCallback((): IntegratedAccountSummary[] => {
    return state.summary?.accounts.filter(account => 
      cashBankIntegrationService.hasVariance(account.variance)
    ) || [];
  }, [state.summary]);

  // Set up auto refresh
  useEffect(() => {
    if (state.autoRefreshEnabled && refreshInterval > 0) {
      refreshIntervalRef.current = setInterval(() => {
        loadSummary();
      }, refreshInterval);
    } else if (refreshIntervalRef.current) {
      clearInterval(refreshIntervalRef.current);
    }

    return () => {
      if (refreshIntervalRef.current) {
        clearInterval(refreshIntervalRef.current);
      }
    };
  }, [state.autoRefreshEnabled, refreshInterval, loadSummary]);

  // Initial load
  useEffect(() => {
    loadSummary();
    
    // Cleanup function
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, [loadSummary]);

  return {
    state,
    loadSummary,
    refreshSummary,
    loadAccount,
    loadReconciliation,
    loadJournalEntries,
    loadTransactionHistory,
    clearCache,
    clearAccountCache,
    refreshAll,
    toggleAutoRefresh,
    getAccountFromCache,
    getReconciliationFromCache,
    hasVarianceAccounts,
    getTotalVariance,
    getAccountsWithVariance
  };
};

export default useCashBankIntegration;