/**
 * Direct Balance Sheet Test
 * 
 * Script untuk mengecek Balance Sheet dari SSOT Journal dengan direct API calls
 */

// Function to get auth headers (mocked - replace with actual token if needed)
function getAuthHeaders() {
  return {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
    // Add authorization header if needed
    // 'Authorization': 'Bearer your-token-here'
  };
}

// API base URL (adjust if different)
const API_V1_BASE = process.env.NEXT_PUBLIC_API_V1_BASE || 'http://localhost:8080/api/v1';

async function testDirectBalanceSheet() {
  console.log('🚀 Testing Balance Sheet with Direct API Calls');
  console.log('='.repeat(70));
  console.log(`🌐 API Base: ${API_V1_BASE}`);
  
  try {
    // Step 1: Test API connectivity
    console.log('\n📡 Step 1: Testing API Connectivity...');
    
    try {
      const response = await fetch(`${API_V1_BASE}/journals/summary`, {
        headers: getAuthHeaders(),
      });
      
      if (response.ok) {
        const summaryData = await response.json();
        console.log('✅ API connected successfully!');
        console.log(`   Status: ${response.status} ${response.statusText}`);
        console.log(`   Response structure:`, Object.keys(summaryData));
        
        if (summaryData.data) {
          console.log('   Summary data:');
          Object.entries(summaryData.data).forEach(([key, value]) => {
            console.log(`     ${key}: ${value}`);
          });
        }
      } else {
        console.log(`❌ API connection failed: ${response.status} ${response.statusText}`);
        const errorText = await response.text();
        console.log(`   Error response: ${errorText.substring(0, 200)}...`);
      }
    } catch (error) {
      console.log('❌ API connectivity test failed:', error.message);
    }

    // Step 2: Test Account Balances endpoint
    console.log('\n💰 Step 2: Testing Account Balances...');
    
    try {
      const response = await fetch(`${API_V1_BASE}/journals/account-balances`, {
        headers: getAuthHeaders(),
      });
      
      if (response.ok) {
        const balancesData = await response.json();
        console.log('✅ Account balances endpoint accessible!');
        console.log(`   Response structure:`, Object.keys(balancesData));
        
        const balances = balancesData.data || [];
        console.log(`   Retrieved ${balances.length} account balances`);
        
        if (balances.length > 0) {
          console.log('\n📊 Sample Account Balances:');
          balances.slice(0, 5).forEach((balance, index) => {
            console.log(`   ${index + 1}. ${balance.account_code || 'N/A'} - ${balance.account_name || 'Unknown'}`);
            console.log(`      Debit: ${balance.debit_balance || 0}, Credit: ${balance.credit_balance || 0}`);
          });
        }
      } else {
        console.log(`❌ Account balances endpoint failed: ${response.status} ${response.statusText}`);
      }
    } catch (error) {
      console.log('❌ Account balances test failed:', error.message);
    }

    // Step 3: Test Accounts endpoint
    console.log('\n🏦 Step 3: Testing Accounts Endpoint...');
    
    try {
      const response = await fetch(`${API_V1_BASE}/accounts`, {
        headers: getAuthHeaders(),
      });
      
      if (response.ok) {
        const accountsData = await response.json();
        console.log('✅ Accounts endpoint accessible!');
        
        const accounts = accountsData.data || [];
        console.log(`   Retrieved ${accounts.length} master accounts`);
        
        if (accounts.length > 0) {
          const balanceSheetAccounts = accounts.filter(acc => 
            ['ASSET', 'LIABILITY', 'EQUITY'].includes(acc.type)
          );
          console.log(`   Balance Sheet Accounts: ${balanceSheetAccounts.length}`);
          
          const assetAccounts = balanceSheetAccounts.filter(acc => acc.type === 'ASSET');
          const liabilityAccounts = balanceSheetAccounts.filter(acc => acc.type === 'LIABILITY');
          const equityAccounts = balanceSheetAccounts.filter(acc => acc.type === 'EQUITY');
          
          console.log(`     - Assets: ${assetAccounts.length} accounts`);
          console.log(`     - Liabilities: ${liabilityAccounts.length} accounts`);
          console.log(`     - Equity: ${equityAccounts.length} accounts`);
        }
      } else {
        console.log(`❌ Accounts endpoint failed: ${response.status} ${response.statusText}`);
      }
    } catch (error) {
      console.log('❌ Accounts endpoint test failed:', error.message);
    }

    // Step 4: Test Journal Entries
    console.log('\n📋 Step 4: Testing Journal Entries...');
    
    try {
      const response = await fetch(`${API_V1_BASE}/journals?status=POSTED&limit=5`, {
        headers: getAuthHeaders(),
      });
      
      if (response.ok) {
        const journalsData = await response.json();
        console.log('✅ Journal entries endpoint accessible!');
        
        const entries = journalsData.data || [];
        console.log(`   Retrieved ${entries.length} recent entries`);
        
        if (entries.length > 0) {
          console.log('\n📝 Recent Journal Entries:');
          entries.forEach((entry, index) => {
            console.log(`   ${index + 1}. ${entry.entry_date || 'N/A'} - ${entry.entry_number || 'N/A'}`);
            console.log(`      ${entry.description || 'No description'}`);
            console.log(`      Debit: ${entry.total_debit || 0}, Credit: ${entry.total_credit || 0}`);
          });
        }
      } else {
        console.log(`❌ Journal entries endpoint failed: ${response.status} ${response.statusText}`);
      }
    } catch (error) {
      console.log('❌ Journal entries test failed:', error.message);
    }

    // Step 5: Generate Balance Sheet
    console.log('\n🧮 Step 5: Generating Balance Sheet...');
    
    try {
      // Get both account balances and master accounts
      const [balancesResponse, accountsResponse] = await Promise.all([
        fetch(`${API_V1_BASE}/journals/account-balances`, { headers: getAuthHeaders() }),
        fetch(`${API_V1_BASE}/accounts`, { headers: getAuthHeaders() })
      ]);
      
      if (balancesResponse.ok && accountsResponse.ok) {
        const balancesData = await balancesResponse.json();
        const accountsData = await accountsResponse.json();
        
        const accountBalances = balancesData.data || [];
        const accounts = accountsData.data || [];
        
        console.log('✅ Starting balance sheet calculation...');
        console.log(`   Account balances: ${accountBalances.length}`);
        console.log(`   Master accounts: ${accounts.length}`);
        
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
          
          accountBalances.forEach(balance => {
            const account = accountMap.get(balance.account_id);
            if (!account) return;
            
            if (!['ASSET', 'LIABILITY', 'EQUITY'].includes(account.type)) return;
            
            let netBalance = 0;
            if (account.type === 'ASSET') {
              netBalance = (balance.debit_balance || 0) - (balance.credit_balance || 0);
              totalAssets += netBalance;
              if (Math.abs(netBalance) > 0.01) {
                assetItems.push({
                  code: account.code,
                  name: account.name,
                  balance: netBalance
                });
              }
            } else if (account.type === 'LIABILITY') {
              netBalance = (balance.credit_balance || 0) - (balance.debit_balance || 0);
              totalLiabilities += netBalance;
              if (Math.abs(netBalance) > 0.01) {
                liabilityItems.push({
                  code: account.code,
                  name: account.name,
                  balance: netBalance
                });
              }
            } else if (account.type === 'EQUITY') {
              netBalance = (balance.credit_balance || 0) - (balance.debit_balance || 0);
              totalEquity += netBalance;
              if (Math.abs(netBalance) > 0.01) {
                equityItems.push({
                  code: account.code,
                  name: account.name,
                  balance: netBalance
                });
              }
            }
          });
          
          // Display Balance Sheet
          console.log('\n' + '='.repeat(70));
          console.log('📊 BALANCE SHEET FROM SSOT JOURNAL SYSTEM');
          console.log(`📅 As of: ${new Date().toLocaleDateString('id-ID')}`);
          console.log(`⏰ Generated: ${new Date().toLocaleString('id-ID')}`);
          console.log('='.repeat(70));
          
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
          console.log('-'.repeat(60));
          if (assetItems.length > 0) {
            // Sort assets by account code
            assetItems.sort((a, b) => a.code.localeCompare(b.code));
            assetItems.forEach(item => {
              const formatted = formatCurrency(item.balance);
              console.log(`${item.code.padEnd(10)} ${item.name.padEnd(30)} ${formatted.padStart(18)}`);
            });
          } else {
            console.log('   No asset accounts with non-zero balances found');
          }
          console.log('-'.repeat(60));
          console.log(`TOTAL ASSETS${' '.repeat(30)}${formatCurrency(totalAssets).padStart(18)}`);
          
          // Liabilities
          console.log('\n💳 LIABILITIES');
          console.log('-'.repeat(60));
          if (liabilityItems.length > 0) {
            // Sort liabilities by account code
            liabilityItems.sort((a, b) => a.code.localeCompare(b.code));
            liabilityItems.forEach(item => {
              const formatted = formatCurrency(item.balance);
              console.log(`${item.code.padEnd(10)} ${item.name.padEnd(30)} ${formatted.padStart(18)}`);
            });
          } else {
            console.log('   No liability accounts with non-zero balances found');
          }
          console.log('-'.repeat(60));
          console.log(`TOTAL LIABILITIES${' '.repeat(25)}${formatCurrency(totalLiabilities).padStart(18)}`);
          
          // Equity
          console.log('\n🏛️  EQUITY');
          console.log('-'.repeat(60));
          if (equityItems.length > 0) {
            // Sort equity by account code
            equityItems.sort((a, b) => a.code.localeCompare(b.code));
            equityItems.forEach(item => {
              const formatted = formatCurrency(item.balance);
              console.log(`${item.code.padEnd(10)} ${item.name.padEnd(30)} ${formatted.padStart(18)}`);
            });
          } else {
            console.log('   No equity accounts with non-zero balances found');
          }
          console.log('-'.repeat(60));
          console.log(`TOTAL EQUITY${' '.repeat(30)}${formatCurrency(totalEquity).padStart(18)}`);
          
          // Summary
          const totalLiabilitiesEquity = totalLiabilities + totalEquity;
          const balanceDifference = totalAssets - totalLiabilitiesEquity;
          const isBalanced = Math.abs(balanceDifference) <= 0.01;
          
          console.log('\n' + '-'.repeat(60));
          console.log(`TOTAL LIABILITIES + EQUITY${' '.repeat(16)}${formatCurrency(totalLiabilitiesEquity).padStart(18)}`);
          console.log('='.repeat(70));
          
          // Balance check
          if (isBalanced) {
            console.log('\n✅ BALANCE SHEET IS BALANCED');
          } else {
            console.log('\n❌ BALANCE SHEET IS NOT BALANCED');
            console.log(`   Difference: ${formatCurrency(balanceDifference)}`);
          }
          
          // Statistics
          console.log('\n📈 STATISTICS');
          console.log('-'.repeat(40));
          console.log(`Asset accounts (non-zero): ${assetItems.length}`);
          console.log(`Liability accounts (non-zero): ${liabilityItems.length}`);
          console.log(`Equity accounts (non-zero): ${equityItems.length}`);
          console.log(`Total accounts in balance sheet: ${assetItems.length + liabilityItems.length + equityItems.length}`);
          
          // Financial Ratios
          console.log('\n📊 FINANCIAL RATIOS');
          console.log('-'.repeat(40));
          
          if (totalEquity !== 0) {
            const debtToEquity = totalLiabilities / totalEquity;
            console.log(`Debt to Equity Ratio: ${debtToEquity.toFixed(2)}`);
          } else {
            console.log('Debt to Equity Ratio: N/A (no equity)');
          }
          
          if (totalAssets !== 0) {
            const equityRatio = totalEquity / totalAssets;
            console.log(`Equity Ratio: ${(equityRatio * 100).toFixed(1)}%`);
            const assetToLiabilityRatio = totalLiabilities !== 0 ? totalAssets / totalLiabilities : 'N/A';
            console.log(`Asset to Liability Ratio: ${typeof assetToLiabilityRatio === 'number' ? assetToLiabilityRatio.toFixed(2) : assetToLiabilityRatio}`);
          } else {
            console.log('Equity Ratio: N/A (no assets)');
            console.log('Asset to Liability Ratio: N/A (no assets)');
          }
          
          // Current assets vs current liabilities (if identifiable)
          const currentAssets = assetItems.filter(item => item.code.startsWith('11')).reduce((sum, item) => sum + item.balance, 0);
          const currentLiabilities = liabilityItems.filter(item => item.code.startsWith('21')).reduce((sum, item) => sum + item.balance, 0);
          
          if (currentLiabilities !== 0 && currentAssets > 0) {
            const currentRatio = currentAssets / currentLiabilities;
            console.log(`Current Ratio (estimated): ${currentRatio.toFixed(2)}`);
          } else {
            console.log('Current Ratio: N/A');
          }
          
        } else {
          console.log('❌ Insufficient data for balance sheet generation');
          console.log(`   Account balances: ${accountBalances.length}`);
          console.log(`   Master accounts: ${accounts.length}`);
        }
      } else {
        console.log('❌ Failed to fetch required data for balance sheet');
        if (!balancesResponse.ok) {
          console.log(`   Account balances error: ${balancesResponse.status}`);
        }
        if (!accountsResponse.ok) {
          console.log(`   Master accounts error: ${accountsResponse.status}`);
        }
      }
      
    } catch (error) {
      console.log('❌ Balance sheet generation failed:', error.message);
      console.error('Error details:', error);
    }

    // Final Summary
    console.log('\n🎯 TEST SUMMARY');
    console.log('='.repeat(70));
    console.log('✅ Direct Balance Sheet Test Completed');
    console.log(`📅 Test completed: ${new Date().toLocaleString('id-ID')}`);
    console.log(`🌐 API Base URL: ${API_V1_BASE}`);
    console.log('='.repeat(70));
    
  } catch (error) {
    console.error('\n💥 Critical Error in Direct Balance Sheet Test:', error);
  }
}

// Run the test
testDirectBalanceSheet()
  .then(() => {
    console.log('\n🎉 Test completed successfully!');
  })
  .catch(error => {
    console.error('\n💥 Test failed:', error);
  });