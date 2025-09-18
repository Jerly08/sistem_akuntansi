/**
 * Deposit Timeout Testing Script
 * Run this in browser console to test deposit operations
 */

console.log('=== Cash Bank Deposit Timeout Test ===');

/**
 * Test deposit operation with timeout monitoring
 */
async function testDepositTimeout(accountId, amount = 100000) {
    console.log('\n🧪 Testing Deposit Timeout');
    console.log('Account ID:', accountId);
    console.log('Amount:', amount);
    
    const token = localStorage.getItem('token');
    if (!token) {
        console.error('❌ No authentication token found');
        return false;
    }
    
    const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
    const startTime = performance.now();
    
    const depositData = {
        account_id: accountId,
        date: new Date().toISOString().split('T')[0],
        amount: amount,
        reference: `TIMEOUT_TEST_${Date.now()}`,
        notes: 'Timeout test deposit'
    };
    
    console.log('📋 Deposit data:', depositData);
    console.log('🚀 Starting deposit operation...');
    
    try {
        const response = await fetch(`${baseURL}/cashbank/deposit`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(depositData)
        });
        
        const endTime = performance.now();
        const duration = endTime - startTime;
        
        console.log(`⏱️ Operation completed in ${Math.round(duration)}ms`);
        console.log('📡 Response status:', response.status, response.statusText);
        
        if (response.ok) {
            const result = await response.json();
            console.log('✅ Deposit successful!');
            console.log('💰 Transaction result:', result);
            
            // Performance analysis
            if (duration > 30000) {
                console.warn('⚠️ SLOW: Operation took longer than 30 seconds');
            } else if (duration > 10000) {
                console.warn('⚠️ MODERATE: Operation took longer than 10 seconds');
            } else {
                console.log('✅ FAST: Operation completed within acceptable time');
            }
            
            return true;
        } else {
            const errorText = await response.text();
            console.error('❌ Deposit failed');
            console.error('Status:', response.status);
            console.error('Error:', errorText);
            
            if (response.status === 408 || response.status === 504) {
                console.error('🔥 TIMEOUT ERROR DETECTED');
            }
            
            return false;
        }
    } catch (error) {
        const endTime = performance.now();
        const duration = endTime - startTime;
        
        console.error('❌ Network/Timeout error after', Math.round(duration), 'ms');
        console.error('Error details:', error.message);
        
        // Detect timeout errors
        if (error.name === 'AbortError' || 
            error.message.includes('timeout') || 
            error.message.includes('exceeded') ||
            duration > 10000) {
            console.error('🔥 TIMEOUT CONFIRMED - This is the root cause!');
            console.log('💡 Recommendations:');
            console.log('1. Backend operation is too slow');
            console.log('2. Database queries need optimization');
            console.log('3. Journal entry creation is blocking the response');
            console.log('4. Consider async processing for journal entries');
        }
        
        return false;
    }
}

/**
 * Find a suitable cash bank account for testing
 */
async function findTestCashBankAccount() {
    console.log('🔍 Looking for cash bank accounts...');
    
    const token = localStorage.getItem('token');
    if (!token) {
        console.error('❌ No authentication token found');
        return null;
    }
    
    const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
    
    try {
        const response = await fetch(`${baseURL}/cashbank/accounts`, {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            console.error('❌ Failed to fetch accounts:', response.status);
            return null;
        }
        
        const accounts = await response.json();
        console.log('📊 Found', accounts.length, 'cash/bank accounts');
        
        if (accounts.length === 0) {
            console.warn('⚠️ No cash/bank accounts found');
            return null;
        }
        
        // Find active account with reasonable balance
        const activeAccount = accounts.find(acc => acc.is_active && acc.balance >= 0);
        
        if (activeAccount) {
            console.log('✅ Selected test account:');
            console.log('Account ID:', activeAccount.id);
            console.log('Name:', activeAccount.name);
            console.log('Type:', activeAccount.type);
            console.log('Balance:', activeAccount.balance);
            console.log('Currency:', activeAccount.currency);
            return activeAccount;
        } else {
            console.warn('⚠️ No suitable account found');
            return accounts[0]; // Return first account anyway
        }
    } catch (error) {
        console.error('❌ Error fetching accounts:', error.message);
        return null;
    }
}

/**
 * Monitor network performance during deposit
 */
async function monitorDepositPerformance(accountId, amount = 50000) {
    console.log('\n📊 Performance Monitoring Test');
    
    const testAccount = await findTestCashBankAccount();
    if (!testAccount) {
        console.error('❌ Cannot run test without account');
        return;
    }
    
    console.log('🎯 Running performance test with varying amounts...');
    
    const testAmounts = [10000, 50000, 100000, 500000, 1000000];
    const results = [];
    
    for (let i = 0; i < testAmounts.length; i++) {
        const amount = testAmounts[i];
        console.log(`\n--- Test ${i + 1}/5: Amount ${amount} ---`);
        
        const startTime = performance.now();
        const success = await testDepositTimeout(testAccount.id, amount);
        const endTime = performance.now();
        const duration = endTime - startTime;
        
        results.push({
            amount: amount,
            duration: Math.round(duration),
            success: success,
            status: duration > 30000 ? 'TIMEOUT' : duration > 10000 ? 'SLOW' : 'OK'
        });
        
        // Wait between tests
        if (i < testAmounts.length - 1) {
            console.log('⏳ Waiting 3 seconds before next test...');
            await new Promise(resolve => setTimeout(resolve, 3000));
        }
    }
    
    console.log('\n📈 PERFORMANCE TEST SUMMARY:');
    console.table(results);
    
    // Analysis
    const timeouts = results.filter(r => r.status === 'TIMEOUT').length;
    const slow = results.filter(r => r.status === 'SLOW').length;
    const fast = results.filter(r => r.status === 'OK').length;
    
    console.log(`\n📊 Results: ${fast} Fast, ${slow} Slow, ${timeouts} Timeouts`);
    
    if (timeouts > 0) {
        console.error('🔥 TIMEOUT ISSUE CONFIRMED');
        console.log('💡 Immediate actions needed:');
        console.log('1. ✅ Frontend timeout increased (implemented)');
        console.log('2. ⚠️ Backend optimization required');
        console.log('3. ⚠️ Database indexing needed');
        console.log('4. ⚠️ Async journal processing recommended');
    } else if (slow > 0) {
        console.warn('⚠️ Performance issues detected - optimization recommended');
    } else {
        console.log('✅ All tests passed within acceptable time');
    }
}

/**
 * Test deposit with simulated high load
 */
async function testDepositUnderLoad() {
    console.log('\n🏋️ Load Testing - Multiple Concurrent Deposits');
    
    const testAccount = await findTestCashBankAccount();
    if (!testAccount) {
        console.error('❌ Cannot run test without account');
        return;
    }
    
    console.log('🎯 Running 3 concurrent deposits...');
    
    const promises = [
        testDepositTimeout(testAccount.id, 25000),
        testDepositTimeout(testAccount.id, 30000), 
        testDepositTimeout(testAccount.id, 35000)
    ];
    
    const startTime = performance.now();
    
    try {
        const results = await Promise.allSettled(promises);
        const endTime = performance.now();
        const totalDuration = endTime - startTime;
        
        console.log(`⏱️ All operations completed in ${Math.round(totalDuration)}ms`);
        
        results.forEach((result, index) => {
            if (result.status === 'fulfilled') {
                console.log(`✅ Deposit ${index + 1}: Success`);
            } else {
                console.error(`❌ Deposit ${index + 1}: Failed - ${result.reason}`);
            }
        });
        
        const successCount = results.filter(r => r.status === 'fulfilled' && r.value === true).length;
        console.log(`\n📊 Concurrent test result: ${successCount}/3 successful`);
        
        if (successCount < 3) {
            console.error('🔥 CONCURRENCY ISSUES DETECTED');
            console.log('This suggests database locking or connection pool issues');
        }
        
    } catch (error) {
        console.error('❌ Load test failed:', error.message);
    }
}

// Export functions for manual use
window.depositTimeoutTest = {
    testDepositTimeout,
    findTestCashBankAccount,
    monitorDepositPerformance,
    testDepositUnderLoad
};

console.log('\n📋 Deposit Timeout Test Tools Loaded');
console.log('Usage:');
console.log('- findTestCashBankAccount()        // Find suitable account');
console.log('- testDepositTimeout(accountId)    // Test single deposit');
console.log('- monitorDepositPerformance()      // Run performance analysis');
console.log('- testDepositUnderLoad()           // Test concurrent deposits');
console.log('\n🚀 Quick start: monitorDepositPerformance()');