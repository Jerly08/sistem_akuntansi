/**
 * Live Test Balance Sheet Calculator
 * 
 * Script untuk menguji Balance Sheet Calculator dengan data aktual dari SSOT Journal
 */

// Import yang diperlukan (menggunakan CommonJS untuk compatibility)
const { ssotJournalService } = require('./src/services/ssotJournalService');
const { accountService } = require('./src/services/accountService');

async function testBalanceSheetLive() {
  console.log('🚀 Testing Balance Sheet Calculator with Live SSOT Journal Data');
  console.log('='.repeat(70));
  
  try {
    // Step 1: Test SSOT Journal Service connectivity
    console.log('\n📡 Step 1: Testing SSOT Journal Service Connection...');
    
    try {
      const journalSummary = await ssotJournalService.getJournalSummary();
      console.log('✅ SSOT Journal Service connected successfully!');
      console.log(`   Total Entries: ${journalSummary.total_entries}`);
      console.log(`   Posted Entries: ${journalSummary.posted_entries}`);
      console.log(`   Total Debit: ${journalSummary.total_debit}`);
      console.log(`   Total Credit: ${journalSummary.total_credit}`);
    } catch (error) {
      console.log('❌ SSOT Journal Service connection failed:', error.message);
    }

    // Step 2: Test Account Balances
    console.log('\n💰 Step 2: Testing Account Balances...');
    
    try {
      const accountBalances = await ssotJournalService.getAccountBalances();
      console.log(`✅ Retrieved ${accountBalances.length} account balances`);
      
      if (accountBalances.length > 0) {
        console.log('\n📊 Sample Account Balances:');
        accountBalances.slice(0, 5).forEach(balance => {
          console.log(`   ${balance.account_code} - ${balance.account_name}: Debit ${balance.debit_balance}, Credit ${balance.credit_balance}`);
        });
      }
    } catch (error) {
      console.log('❌ Account Balances retrieval failed:', error.message);
    }

    // Step 3: Test Account Service
    console.log('\n🏦 Step 3: Testing Account Service...');
    
    try {
      const accounts = await accountService.getAccounts();
      console.log(`✅ Retrieved ${accounts.data?.length || 0} master accounts`);
      
      if (accounts.data && accounts.data.length > 0) {
        const balanceSheetAccounts = accounts.data.filter(acc => 
          ['ASSET', 'LIABILITY', 'EQUITY'].includes(acc.type)
        );
        console.log(`   Balance Sheet Accounts: ${balanceSheetAccounts.length}`);
        
        const assetAccounts = balanceSheetAccounts.filter(acc => acc.type === 'ASSET');
        const liabilityAccounts = balanceSheetAccounts.filter(acc => acc.type === 'LIABILITY');
        const equityAccounts = balanceSheetAccounts.filter(acc => acc.type === 'EQUITY');
        
        console.log(`   - Assets: ${assetAccounts.length} accounts`);
        console.log(`   - Liabilities: ${liabilityAccounts.length} accounts`);
        console.log(`   - Equity: ${equityAccounts.length} accounts`);
      }
    } catch (error) {
      console.log('❌ Account Service connection failed:', error.message);
    }

    // Step 4: Test Recent Journal Entries
    console.log('\n📋 Step 4: Testing Recent Journal Entries...');
    
    try {
      const recentEntries = await ssotJournalService.getJournalEntries({
        status: 'POSTED',
        limit: 5
      });
      
      console.log(`✅ Retrieved ${recentEntries.data?.length || 0} recent entries`);
      
      if (recentEntries.data && recentEntries.data.length > 0) {
        console.log('\n📝 Recent Journal Entries:');
        recentEntries.data.forEach(entry => {
          console.log(`   ${entry.entry_date} - ${entry.entry_number}: ${entry.description}`);
          console.log(`     Debit: ${entry.total_debit}, Credit: ${entry.total_credit}, Balanced: ${entry.is_balanced ? '✅' : '❌'}`);
        });
      }
    } catch (error) {
      console.log('❌ Recent Journal Entries retrieval failed:', error.message);
    }

    // Step 5: Manual Balance Sheet Calculation
    console.log('\n🧮 Step 5: Manual Balance Sheet Calculation...');
    
    try {
      // Get data for manual calculation
      const accountBalances = await ssotJournalService.getAccountBalances();
      const accounts = await accountService.getAccounts();
      
      if (accountBalances.length > 0 && accounts.data && accounts.data.length > 0) {
        console.log('✅ Starting manual balance sheet calculation...');
        
        // Create account lookup map
        const accountMap = new Map();
        accounts.data.forEach(account => accountMap.set(account.id, account));
        
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
            netBalance = balance.debit_balance - balance.credit_balance;
            totalAssets += netBalance;
            if (Math.abs(netBalance) > 0.01) {
              assetItems.push({
                code: account.code,
                name: account.name,
                balance: netBalance
              });
            }
          } else if (account.type === 'LIABILITY') {
            netBalance = balance.credit_balance - balance.debit_balance;
            totalLiabilities += netBalance;
            if (Math.abs(netBalance) > 0.01) {
              liabilityItems.push({
                code: account.code,
                name: account.name,
                balance: netBalance
              });
            }
          } else if (account.type === 'EQUITY') {
            netBalance = balance.credit_balance - balance.debit_balance;
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
        
        // Display results
        console.log('\n' + '='.repeat(70));
        console.log('📊 BALANCE SHEET FROM SSOT JOURNAL DATA');
        console.log(`As of: ${new Date().toISOString().split('T')[0]}`);
        console.log('='.repeat(70));
        
        // Assets
        console.log('\n🏢 ASSETS');
        console.log('-'.repeat(50));
        if (assetItems.length > 0) {
          assetItems.forEach(item => {
            const formattedBalance = new Intl.NumberFormat('id-ID', {
              style: 'currency',
              currency: 'IDR',
              minimumFractionDigits: 0
            }).format(item.balance);
            console.log(`${item.code.padEnd(8)} ${item.name.padEnd(25)} ${formattedBalance.padStart(15)}`);
          });
        } else {
          console.log('No asset accounts found');
        }
        console.log('-'.repeat(50));
        const formattedTotalAssets = new Intl.NumberFormat('id-ID', {
          style: 'currency',
          currency: 'IDR',
          minimumFractionDigits: 0
        }).format(totalAssets);
        console.log(`TOTAL ASSETS${' '.repeat(26)}${formattedTotalAssets.padStart(15)}`);
        
        // Liabilities
        console.log('\n💳 LIABILITIES');
        console.log('-'.repeat(50));
        if (liabilityItems.length > 0) {
          liabilityItems.forEach(item => {
            const formattedBalance = new Intl.NumberFormat('id-ID', {
              style: 'currency',
              currency: 'IDR',
              minimumFractionDigits: 0
            }).format(item.balance);
            console.log(`${item.code.padEnd(8)} ${item.name.padEnd(25)} ${formattedBalance.padStart(15)}`);
          });
        } else {
          console.log('No liability accounts found');
        }
        console.log('-'.repeat(50));
        const formattedTotalLiabilities = new Intl.NumberFormat('id-ID', {
          style: 'currency',
          currency: 'IDR',
          minimumFractionDigits: 0
        }).format(totalLiabilities);
        console.log(`TOTAL LIABILITIES${' '.repeat(19)}${formattedTotalLiabilities.padStart(15)}`);
        
        // Equity
        console.log('\n🏛️  EQUITY');
        console.log('-'.repeat(50));
        if (equityItems.length > 0) {
          equityItems.forEach(item => {
            const formattedBalance = new Intl.NumberFormat('id-ID', {
              style: 'currency',
              currency: 'IDR',
              minimumFractionDigits: 0
            }).format(item.balance);
            console.log(`${item.code.padEnd(8)} ${item.name.padEnd(25)} ${formattedBalance.padStart(15)}`);
          });
        } else {
          console.log('No equity accounts found');
        }
        console.log('-'.repeat(50));
        const formattedTotalEquity = new Intl.NumberFormat('id-ID', {
          style: 'currency',
          currency: 'IDR',
          minimumFractionDigits: 0
        }).format(totalEquity);
        console.log(`TOTAL EQUITY${' '.repeat(24)}${formattedTotalEquity.padStart(15)}`);
        
        // Summary
        const totalLiabilitiesEquity = totalLiabilities + totalEquity;
        const balanceDifference = totalAssets - totalLiabilitiesEquity;
        const isBalanced = Math.abs(balanceDifference) <= 0.01;
        
        console.log('\n' + '-'.repeat(50));
        const formattedTotalLiabilitiesEquity = new Intl.NumberFormat('id-ID', {
          style: 'currency',
          currency: 'IDR',
          minimumFractionDigits: 0
        }).format(totalLiabilitiesEquity);
        console.log(`TOTAL LIABILITIES + EQUITY${' '.repeat(8)}${formattedTotalLiabilitiesEquity.padStart(15)}`);
        console.log('='.repeat(70));
        
        // Balance check
        if (isBalanced) {
          console.log('✅ BALANCE SHEET IS BALANCED');
        } else {
          console.log('❌ BALANCE SHEET IS NOT BALANCED');
          const formattedDifference = new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            minimumFractionDigits: 0
          }).format(balanceDifference);
          console.log(`   Difference: ${formattedDifference}`);
        }
        
        // Statistics
        console.log('\n📈 Statistics:');
        console.log(`   Asset accounts: ${assetItems.length}`);
        console.log(`   Liability accounts: ${liabilityItems.length}`);
        console.log(`   Equity accounts: ${equityItems.length}`);
        console.log(`   Total accounts: ${assetItems.length + liabilityItems.length + equityItems.length}`);
        
        // Simple ratios
        if (totalEquity !== 0) {
          const debtToEquity = totalLiabilities / totalEquity;
          console.log(`   Debt to Equity Ratio: ${debtToEquity.toFixed(2)}`);
        }
        
        if (totalAssets !== 0) {
          const equityRatio = totalEquity / totalAssets;
          console.log(`   Equity Ratio: ${(equityRatio * 100).toFixed(1)}%`);
        }
        
      } else {
        console.log('❌ Insufficient data for balance sheet calculation');
      }
    } catch (error) {
      console.log('❌ Manual balance sheet calculation failed:', error.message);
      console.error('Error details:', error);
    }

    // Step 6: System Check Summary
    console.log('\n🔍 Step 6: System Check Summary...');
    console.log('='.repeat(70));
    console.log('✅ Balance Sheet Calculator Live Test Completed!');
    console.log(`📅 Test Date: ${new Date().toLocaleString('id-ID')}`);
    console.log('='.repeat(70));
    
  } catch (error) {
    console.error('\n❌ Critical Error in Balance Sheet Live Test:', error);
  }
}

// Run the test
if (require.main === module) {
  testBalanceSheetLive()
    .then(() => {
      console.log('\n🎉 Live test completed successfully!');
      process.exit(0);
    })
    .catch(error => {
      console.error('\n💥 Live test failed:', error);
      process.exit(1);
    });
}

module.exports = { testBalanceSheetLive };