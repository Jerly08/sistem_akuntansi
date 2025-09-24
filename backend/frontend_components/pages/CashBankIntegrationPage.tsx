// CashBank Integration Page
// Main page component that combines dashboard and account detail views for CashBank-SSOT integration

import React, { useState } from 'react';
import CashBankIntegratedDashboard from '../components/CashBankIntegratedDashboard';
import CashBankAccountDetail from '../components/CashBankAccountDetail';

type ViewMode = 'dashboard' | 'account-detail';

interface ViewState {
  mode: ViewMode;
  selectedAccountId?: number;
}

export const CashBankIntegrationPage: React.FC = () => {
  const [viewState, setViewState] = useState<ViewState>({ mode: 'dashboard' });

  // Handle account selection from dashboard
  const handleAccountSelect = (accountId: number) => {
    setViewState({
      mode: 'account-detail',
      selectedAccountId: accountId
    });
  };

  // Handle back to dashboard
  const handleBackToDashboard = () => {
    setViewState({ mode: 'dashboard' });
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Page Header */}
      <div className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="py-4">
            <nav className="flex items-center space-x-2 text-sm text-gray-500">
              <button
                onClick={handleBackToDashboard}
                className={`hover:text-gray-700 ${
                  viewState.mode === 'dashboard' ? 'text-blue-600 font-medium' : ''
                }`}
              >
                Cash & Bank
              </button>
              {viewState.mode === 'account-detail' && (
                <>
                  <span className="text-gray-300">/</span>
                  <span className="text-gray-700">Detail Akun</span>
                </>
              )}
            </nav>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {viewState.mode === 'dashboard' && (
          <CashBankIntegratedDashboard onAccountSelect={handleAccountSelect} />
        )}
        
        {viewState.mode === 'account-detail' && viewState.selectedAccountId && (
          <CashBankAccountDetail
            accountId={viewState.selectedAccountId}
            onBack={handleBackToDashboard}
          />
        )}
      </div>

      {/* Footer */}
      <div className="bg-white border-t border-gray-200 mt-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between text-sm text-gray-500">
            <div>
              <span>Cash & Bank Integration dengan SSOT Journal System</span>
            </div>
            <div className="flex items-center space-x-4">
              <span>🔄 Auto-refresh setiap 5 menit</span>
              <span>📊 Real-time reconciliation</span>
              <span>🔍 Audit trail lengkap</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CashBankIntegrationPage;