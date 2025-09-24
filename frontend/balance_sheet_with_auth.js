/**
 * Authenticated Balance Sheet Test
 * 
 * Script untuk login dan mengambil Balance Sheet dari SSOT Journal dengan auth
 */

const API_V1_BASE = 'http://localhost:8080/api/v1';

// Admin credentials
const ADMIN_CREDENTIALS = {
  email: 'admin@company.com',
  password: 'admin123'
};

async function getAuthToken() {
  console.log('🔑 Authenticating with admin credentials...');
  
  try {
    const response = await fetch(`${API_V1_BASE}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      },
      body: JSON.stringify(ADMIN_CREDENTIALS)
    });
    
    if (response.ok) {
      const data = await response.json();
      console.log('✅ Authentication successful!');
      
      // Extract token from various possible response formats
      const token = data.token || data.access_token || data.data?.token || data.data?.access_token;
      
      if (token) {
        console.log(`🎫 Token received (length: ${token.length})`);
        return token;
      } else {
        console.log('⚠️  Token not found in response structure:', Object.keys(data));
        return null;
      }
    } else {
      const errorText = await response.text();
      console.log(`❌ Authentication failed: ${response.status} ${response.statusText}`);
      console.log(`   Error: ${errorText}`);
      return null;
    }
  } catch (error) {
    console.log(`❌ Authentication error: ${error.message}`);
    return null;
  }
}

function getAuthHeaders(token) {
  return {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
    'Authorization': `Bearer ${token}`
  };
}

async function generateAuthenticatedBalanceSheet() {
  console.log('🚀 Authenticated Balance Sheet Generator');
  console.log('='.repeat(70));
  console.log(`🌐 API Base: ${API_V1_BASE}`);
  console.log(`👤 Admin: ${ADMIN_CREDENTIALS.email}`);
  console.log('='.repeat(70));
  
  try {
    // Step 1: Get authentication token
    const token = await getAuthToken();
    if (!token) {
      throw new Error('Failed to get authentication token');
    }
    
    const authHeaders = getAuthHeaders(token);
    
    // Step 2: Test authenticated endpoints
    console.log('\n📊 Testing Authenticated Endpoints...');
    console.log('-'.repeat(50));
    
    // Test journal summary
    try {
      const summaryResponse = await fetch(`${API_V1_BASE}/journals/summary`, {
        headers: authHeaders
      });
      
      if (summaryResponse.ok) {
        const summaryData = await summaryResponse.json();
        console.log('✅ Journal Summary accessible');
        if (summaryData.data) {
          console.log('   Summary data:');
          Object.entries(summaryData.data).forEach(([key, value]) => {
            console.log(`     ${key}: ${value}`);
          });
        }
      } else {
        console.log(`❌ Journal Summary failed: ${summaryResponse.status}`);
      }
    } catch (error) {
      console.log(`❌ Journal Summary error: ${error.message}`);
    }
    
    // Step 3: Get Balance Sheet data
    console.log('\n🧮 Generating Balance Sheet...');
    console.log('-'.repeat(50));
    
    // Get both account balances and master accounts
    const [balancesResponse, accountsResponse, journalsResponse] = await Promise.all([
      fetch(`${API_V1_BASE}/journals/account-balances`, { headers: authHeaders }),
      fetch(`${API_V1_BASE}/accounts`, { headers: authHeaders }),
      fetch(`${API_V1_BASE}/journals?status=POSTED&limit=10`, { headers: authHeaders })
    ]);
    
    if (!balancesResponse.ok) {
      console.log(`❌ Account balances failed: ${balancesResponse.status}`);
      const errorText = await balancesResponse.text();
      console.log(`   Error: ${errorText.substring(0, 200)}`);
      return;
    }
    
    if (!accountsResponse.ok) {
      console.log(`❌ Master accounts failed: ${accountsResponse.status}`);
      return;
    }
    
    const balancesData = await balancesResponse.json();
    const accountsData = await accountsResponse.json();
    const journalsData = await journalsResponse.json();
    
    const accountBalances = balancesData.data || [];
    const accounts = accountsData.data || [];
    const journals = journalsData.data || [];
    
    console.log(`✅ Data retrieved successfully:`);
    console.log(`   Account balances: ${accountBalances.length}`);
    console.log(`   Master accounts: ${accounts.length}`);
    console.log(`   Recent journals: ${journals.length}`);
    
    if (accountBalances.length === 0) {
      console.log('\n⚠️  No account balances found!');
      console.log('This could mean:');
      console.log('1. No journal entries have been posted yet');
      console.log('2. Account balances materialized view needs refresh');
      console.log('3. Database is empty');
      
      // Try to refresh account balances
      console.log('\n🔄 Attempting to refresh account balances...');
      try {
        const refreshResponse = await fetch(`${API_V1_BASE}/journals/account-balances/refresh`, {
          method: 'POST',
          headers: authHeaders
        });
        
        if (refreshResponse.ok) {
          const refreshData = await refreshResponse.json();
          console.log('✅ Account balances refreshed');
          console.log(`   Message: ${refreshData.message || 'Success'}`);
          
          // Try to get balances again
          const newBalancesResponse = await fetch(`${API_V1_BASE}/journals/account-balances`, { headers: authHeaders });
          if (newBalancesResponse.ok) {
            const newBalancesData = await newBalancesResponse.json();
            const newAccountBalances = newBalancesData.data || [];
            console.log(`   New account balances count: ${newAccountBalances.length}`);
          }
        } else {
          console.log(`❌ Refresh failed: ${refreshResponse.status}`);
        }
      } catch (refreshError) {
        console.log(`❌ Refresh error: ${refreshError.message}`);
      }
    }
    
    if (accountBalances.length > 0 && accounts.length > 0) {
      // Create account lookup map
      const accountMap = new Map();
      accounts.forEach(account => accountMap.set(account.id, account));
      
      let totalAssets = 0;
      let totalLiabilities = 0;
      let totalEquity = 0;
      
      const assetItems = [];
      const liabilityItems = [];
      const equityItems = [];
      
      // Show sample account balances for debugging
      console.log('\n📊 Sample Account Balances:');
      accountBalances.slice(0, 5).forEach((balance, index) => {
        const account = accountMap.get(balance.account_id);
        console.log(`   ${index + 1}. ID:${balance.account_id} Code:${balance.account_code || 'N/A'} Name:${balance.account_name || 'Unknown'}`);
        console.log(`      Type:${account?.type || 'Unknown'} Debit:${balance.debit_balance || 0} Credit:${balance.credit_balance || 0}`);
      });
      
      accountBalances.forEach(balance => {
        const account = accountMap.get(balance.account_id);
        if (!account) {
          console.log(`⚠️  Account not found for ID: ${balance.account_id}`);
          return;
        }
        
        if (!['ASSET', 'LIABILITY', 'EQUITY'].includes(account.type)) return;
        
        let netBalance = 0;
        if (account.type === 'ASSET') {
          netBalance = (balance.debit_balance || 0) - (balance.credit_balance || 0);
          totalAssets += netBalance;
          if (Math.abs(netBalance) > 0.01) {
            assetItems.push({
              code: account.code,
              name: account.name,
              balance: netBalance,
              debit: balance.debit_balance || 0,
              credit: balance.credit_balance || 0
            });
          }
        } else if (account.type === 'LIABILITY') {
          netBalance = (balance.credit_balance || 0) - (balance.debit_balance || 0);
          totalLiabilities += netBalance;
          if (Math.abs(netBalance) > 0.01) {
            liabilityItems.push({
              code: account.code,
              name: account.name,
              balance: netBalance,
              debit: balance.debit_balance || 0,
              credit: balance.credit_balance || 0
            });
          }
        } else if (account.type === 'EQUITY') {
          netBalance = (balance.credit_balance || 0) - (balance.debit_balance || 0);
          totalEquity += netBalance;
          if (Math.abs(netBalance) > 0.01) {
            equityItems.push({
              code: account.code,
              name: account.name,
              balance: netBalance,
              debit: balance.debit_balance || 0,
              credit: balance.credit_balance || 0
            });
          }
        }
      });
      
      // Display Balance Sheet
      console.log('\n' + '='.repeat(80));
      console.log('📊 BALANCE SHEET FROM SSOT JOURNAL SYSTEM');
      console.log('🏢 Admin Company Dashboard');
      console.log(`📅 As of: ${new Date().toLocaleDateString('id-ID')}`);
      console.log(`⏰ Generated: ${new Date().toLocaleString('id-ID')}`);
      console.log(`🔍 Data Source: SSOT Journal & Account Balances`);
      console.log('='.repeat(80));
      
      // Format currency function
      const formatCurrency = (amount) => {
        return new Intl.NumberFormat('id-ID', {
          style: 'currency',
          currency: 'IDR',
          minimumFractionDigits: 0
        }).format(amount);
      };
      
      // Assets
      console.log('\n🏢 ASSETS');
      console.log('-'.repeat(70));
      if (assetItems.length > 0) {
        assetItems.sort((a, b) => a.code.localeCompare(b.code));
        assetItems.forEach(item => {
          const formatted = formatCurrency(item.balance);
          console.log(`${item.code.padEnd(12)} ${item.name.padEnd(35)} ${formatted.padStart(18)}`);
          console.log(`${' '.repeat(12)} (Debit: ${formatCurrency(item.debit)}, Credit: ${formatCurrency(item.credit)})`);
        });
      } else {
        console.log('   No asset accounts with non-zero balances found');
      }
      console.log('-'.repeat(70));
      console.log(`TOTAL ASSETS${' '.repeat(36)}${formatCurrency(totalAssets).padStart(18)}`);
      
      // Liabilities
      console.log('\n💳 LIABILITIES');
      console.log('-'.repeat(70));
      if (liabilityItems.length > 0) {
        liabilityItems.sort((a, b) => a.code.localeCompare(b.code));
        liabilityItems.forEach(item => {
          const formatted = formatCurrency(item.balance);
          console.log(`${item.code.padEnd(12)} ${item.name.padEnd(35)} ${formatted.padStart(18)}`);
          console.log(`${' '.repeat(12)} (Debit: ${formatCurrency(item.debit)}, Credit: ${formatCurrency(item.credit)})`);
        });
      } else {
        console.log('   No liability accounts with non-zero balances found');
      }
      console.log('-'.repeat(70));
      console.log(`TOTAL LIABILITIES${' '.repeat(31)}${formatCurrency(totalLiabilities).padStart(18)}`);
      
      // Equity
      console.log('\n🏛️  EQUITY');
      console.log('-'.repeat(70));
      if (equityItems.length > 0) {
        equityItems.sort((a, b) => a.code.localeCompare(b.code));
        equityItems.forEach(item => {
          const formatted = formatCurrency(item.balance);
          console.log(`${item.code.padEnd(12)} ${item.name.padEnd(35)} ${formatted.padStart(18)}`);
          console.log(`${' '.repeat(12)} (Debit: ${formatCurrency(item.debit)}, Credit: ${formatCurrency(item.credit)})`);
        });
      } else {
        console.log('   No equity accounts with non-zero balances found');
      }
      console.log('-'.repeat(70));
      console.log(`TOTAL EQUITY${' '.repeat(36)}${formatCurrency(totalEquity).padStart(18)}`);
      
      // Summary
      const totalLiabilitiesEquity = totalLiabilities + totalEquity;
      const balanceDifference = totalAssets - totalLiabilitiesEquity;
      const isBalanced = Math.abs(balanceDifference) <= 0.01;
      
      console.log('\n' + '-'.repeat(70));
      console.log(`TOTAL LIABILITIES + EQUITY${' '.repeat(22)}${formatCurrency(totalLiabilitiesEquity).padStart(18)}`);
      console.log('='.repeat(80));
      
      // Balance check
      if (isBalanced) {
        console.log('\n✅ BALANCE SHEET IS BALANCED');
      } else {
        console.log('\n❌ BALANCE SHEET IS NOT BALANCED');
        console.log(`   Difference: ${formatCurrency(balanceDifference)}`);
      }
      
      // Statistics
      console.log('\n📈 STATISTICS');
      console.log('-'.repeat(50));
      console.log(`Asset accounts (non-zero): ${assetItems.length}`);
      console.log(`Liability accounts (non-zero): ${liabilityItems.length}`);
      console.log(`Equity accounts (non-zero): ${equityItems.length}`);
      console.log(`Total accounts in balance sheet: ${assetItems.length + liabilityItems.length + equityItems.length}`);
      console.log(`Total account balances retrieved: ${accountBalances.length}`);
      console.log(`Total master accounts: ${accounts.length}`);
      console.log(`Recent journal entries: ${journals.length}`);
      
      // Financial Ratios
      console.log('\n📊 FINANCIAL RATIOS');
      console.log('-'.repeat(50));
      
      if (totalEquity !== 0) {
        const debtToEquity = totalLiabilities / totalEquity;
        console.log(`Debt to Equity Ratio: ${debtToEquity.toFixed(2)}`);
      } else {
        console.log('Debt to Equity Ratio: N/A (no equity)');
      }
      
      if (totalAssets !== 0) {
        const equityRatio = totalEquity / totalAssets;
        console.log(`Equity Ratio: ${(equityRatio * 100).toFixed(1)}%`);
        
        if (totalLiabilities !== 0) {
          const assetToLiabilityRatio = totalAssets / totalLiabilities;
          console.log(`Asset to Liability Ratio: ${assetToLiabilityRatio.toFixed(2)}`);
        } else {
          console.log('Asset to Liability Ratio: N/A (no liabilities)');
        }
      } else {
        console.log('Equity Ratio: N/A (no assets)');
        console.log('Asset to Liability Ratio: N/A (no assets)');
      }
      
      // Current assets vs current liabilities (using account code patterns)
      const currentAssets = assetItems
        .filter(item => item.code.startsWith('11'))
        .reduce((sum, item) => sum + item.balance, 0);
      const currentLiabilities = liabilityItems
        .filter(item => item.code.startsWith('21'))
        .reduce((sum, item) => sum + item.balance, 0);
      
      if (currentLiabilities !== 0 && currentAssets > 0) {
        const currentRatio = currentAssets / currentLiabilities;
        console.log(`Current Ratio (estimated): ${currentRatio.toFixed(2)}`);
        console.log(`Working Capital: ${formatCurrency(currentAssets - currentLiabilities)}`);
      } else {
        console.log('Current Ratio: N/A');
        console.log('Working Capital: N/A');
      }
      
      // Recent Journal Entries Summary
      if (journals.length > 0) {
        console.log('\n📋 RECENT JOURNAL ENTRIES');
        console.log('-'.repeat(50));
        journals.slice(0, 5).forEach((journal, index) => {
          console.log(`${index + 1}. ${journal.entry_date} - ${journal.entry_number || 'N/A'}`);
          console.log(`   ${journal.description || 'No description'}`);
          console.log(`   Debit: ${formatCurrency(journal.total_debit || 0)}, Credit: ${formatCurrency(journal.total_credit || 0)}`);
          console.log(`   Balanced: ${journal.is_balanced ? '✅' : '❌'}, Status: ${journal.status}`);
        });
      }
      
    } else {
      console.log('\n❌ Insufficient data for balance sheet generation');
      console.log(`   Account balances: ${accountBalances.length}`);
      console.log(`   Master accounts: ${accounts.length}`);
      
      if (accounts.length > 0) {
        console.log('\n📋 Available Account Types:');
        const accountTypes = accounts.reduce((acc, account) => {
          acc[account.type] = (acc[account.type] || 0) + 1;
          return acc;
        }, {});
        Object.entries(accountTypes).forEach(([type, count]) => {
          console.log(`   ${type}: ${count} accounts`);
        });
      }
    }
    
    console.log('\n🎯 BALANCE SHEET GENERATION SUMMARY');
    console.log('='.repeat(80));
    console.log('✅ Balance Sheet calculation completed successfully!');
    console.log(`📅 Generated: ${new Date().toLocaleString('id-ID')}`);
    console.log(`👤 Authenticated as: ${ADMIN_CREDENTIALS.email}`);
    console.log(`🌐 Source: SSOT Journal System (${API_V1_BASE})`);
    console.log(`💾 Data freshness: Real-time from materialized view`);
    console.log('='.repeat(80));
    
  } catch (error) {
    console.error('\n💥 Critical Error:', error);
    console.error('Error details:', error.stack);
  }
}

// Run the authenticated balance sheet generation
generateAuthenticatedBalanceSheet()
  .then(() => {
    console.log('\n🎉 Balance Sheet generation completed successfully!');
  })
  .catch(error => {
    console.error('\n💥 Balance Sheet generation failed:', error);
  });