/**
 * Test Script for Integrated Payment Endpoint
 * This script tests the /sales/{id}/integrated-payment endpoint
 * Run this in the browser console or as a Node.js script
 */

console.log('=== Integrated Payment Endpoint Test ===');

/**
 * Test the integrated payment endpoint with proper payload formatting
 * @param {number} saleId - The sale ID to create payment for
 * @param {Object} paymentData - Payment data object
 * @param {boolean} dryRun - If true, only validates payload without sending request
 */
async function testIntegratedPayment(saleId, paymentData, dryRun = false) {
    console.log('\n🧪 Testing Integrated Payment Endpoint');
    console.log('Sale ID:', saleId);
    console.log('Payment Data:', paymentData);
    
    // Validate required fields
    const requiredFields = ['amount', 'date', 'method', 'cash_bank_id'];
    const missingFields = requiredFields.filter(field => !paymentData[field]);
    
    if (missingFields.length > 0) {
        console.error('❌ Missing required fields:', missingFields);
        return false;
    }
    
    // Validate field types and formats
    const validationErrors = [];
    
    if (typeof paymentData.amount !== 'number' || paymentData.amount <= 0) {
        validationErrors.push('amount must be a positive number');
    }
    
    if (typeof paymentData.cash_bank_id !== 'number' || paymentData.cash_bank_id <= 0) {
        validationErrors.push('cash_bank_id must be a positive number');
    }
    
    // Validate date format (should be ISO string)
    if (typeof paymentData.date !== 'string' || !isValidISODate(paymentData.date)) {
        validationErrors.push('date must be a valid ISO datetime string');
    }
    
    // Validate payment method
    const validMethods = ['CASH', 'BANK_TRANSFER', 'CHECK', 'CREDIT_CARD', 'DEBIT_CARD', 'OTHER'];
    if (!validMethods.includes(paymentData.method)) {
        validationErrors.push(`method must be one of: ${validMethods.join(', ')}`);
    }
    
    if (validationErrors.length > 0) {
        console.error('❌ Validation errors:', validationErrors);
        return false;
    }
    
    console.log('✅ Payload validation passed');
    
    if (dryRun) {
        console.log('🏁 Dry run completed - actual request not sent');
        return true;
    }
    
    // Get authentication token
    const token = localStorage.getItem('token');
    if (!token) {
        console.error('❌ No authentication token found - please login first');
        return false;
    }
    
    const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
    const endpoint = `${baseURL}/sales/${saleId}/integrated-payment`;
    
    console.log('🚀 Sending request to:', endpoint);
    console.log('📋 Headers:', {
        'Authorization': `Bearer ${token.substring(0, 20)}...`,
        'Content-Type': 'application/json'
    });
    
    try {
        const response = await fetch(endpoint, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(paymentData)
        });
        
        console.log('📡 Response status:', response.status, response.statusText);
        console.log('📡 Response headers:', Object.fromEntries(response.headers.entries()));
        
        if (response.ok) {
            const result = await response.json();
            console.log('✅ Payment created successfully!');
            console.log('💰 Payment result:', result);
            
            // Validate response structure
            if (result.payment && result.sale_payment) {
                console.log('✅ Both payment records created (Payment Management + Sales)');
                console.log('Payment ID:', result.payment.id);
                console.log('Sale Payment ID:', result.sale_payment.id);
            } else {
                console.warn('⚠️ Unexpected response structure:', Object.keys(result));
            }
            
            return true;
        } else {
            const errorText = await response.text();
            console.error('❌ Payment creation failed');
            console.error('Status:', response.status);
            console.error('Error:', errorText);
            
            // Parse error if JSON
            try {
                const errorData = JSON.parse(errorText);
                console.error('Error details:', errorData);
            } catch (e) {
                console.error('Raw error:', errorText);
            }
            
            return false;
        }
    } catch (error) {
        console.error('❌ Network error:', error.message);
        return false;
    }
}

/**
 * Validate ISO date string
 */
function isValidISODate(dateString) {
    const isoRegex = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{3})?Z?$/;
    if (!isoRegex.test(dateString)) {
        return false;
    }
    
    const date = new Date(dateString);
    return !isNaN(date.getTime());
}

/**
 * Create a test payment data object
 */
function createTestPaymentData(amount = 100000) {
    return {
        amount: amount,
        date: new Date().toISOString(),
        method: 'BANK_TRANSFER',
        cash_bank_id: 1, // Assuming first bank account exists
        reference: `TEST_${Date.now()}`,
        notes: 'Test payment created by automated test'
    };
}

/**
 * Find a suitable sale for testing (with outstanding balance)
 */
async function findTestSale() {
    console.log('🔍 Looking for a suitable sale for testing...');
    
    const token = localStorage.getItem('token');
    if (!token) {
        console.error('❌ No authentication token found');
        return null;
    }
    
    const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
    
    try {
        const response = await fetch(`${baseURL}/sales?status=INVOICED&limit=10`, {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            console.error('❌ Failed to fetch sales:', response.status);
            return null;
        }
        
        const result = await response.json();
        console.log('📊 Found', result.data?.length || 0, 'invoiced sales');
        
        if (!result.data || result.data.length === 0) {
            console.warn('⚠️ No invoiced sales found for testing');
            return null;
        }
        
        // Find sale with outstanding balance
        const saleWithBalance = result.data.find(sale => sale.outstanding_amount > 0);
        
        if (saleWithBalance) {
            console.log('✅ Found sale with outstanding balance:');
            console.log('Sale ID:', saleWithBalance.id);
            console.log('Code:', saleWithBalance.code);
            console.log('Outstanding:', saleWithBalance.outstanding_amount);
            console.log('Customer:', saleWithBalance.customer?.name);
            return saleWithBalance;
        } else {
            console.warn('⚠️ All invoiced sales are fully paid');
            return result.data[0]; // Return first sale anyway for testing
        }
    } catch (error) {
        console.error('❌ Error fetching sales:', error.message);
        return null;
    }
}

/**
 * Check if cash/bank accounts are available
 */
async function checkCashBankAccounts() {
    console.log('🏦 Checking available cash/bank accounts...');
    
    const token = localStorage.getItem('token');
    if (!token) {
        console.error('❌ No authentication token found');
        return false;
    }
    
    const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
    
    try {
        const response = await fetch(`${baseURL}/cashbanks/payment-accounts`, {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            console.error('❌ Failed to fetch payment accounts:', response.status);
            return false;
        }
        
        const accounts = await response.json();
        console.log('💳 Available payment accounts:', accounts.length);
        
        if (accounts.length === 0) {
            console.error('❌ No payment accounts configured');
            return false;
        }
        
        accounts.forEach((account, index) => {
            console.log(`Account ${index + 1}:`, {
                id: account.id,
                code: account.code,
                name: account.name,
                type: account.type
            });
        });
        
        return accounts[0]; // Return first account for testing
    } catch (error) {
        console.error('❌ Error fetching accounts:', error.message);
        return false;
    }
}

/**
 * Run full integrated payment test
 */
async function runFullPaymentTest(dryRun = true) {
    console.log('🚀 Starting Full Integrated Payment Test');
    console.log('Mode:', dryRun ? 'DRY RUN' : 'LIVE TEST');
    
    // Step 1: Check authentication
    const token = localStorage.getItem('token');
    if (!token) {
        console.error('❌ Not authenticated - please login first');
        return false;
    }
    console.log('✅ Authentication token found');
    
    // Step 2: Check cash/bank accounts
    const firstAccount = await checkCashBankAccounts();
    if (!firstAccount) {
        console.error('❌ Cannot proceed without payment accounts');
        return false;
    }
    
    // Step 3: Find test sale
    const testSale = await findTestSale();
    if (!testSale) {
        console.error('❌ Cannot proceed without test sale');
        return false;
    }
    
    // Step 4: Create test payment data
    const testAmount = Math.min(50000, testSale.outstanding_amount || 50000);
    const paymentData = {
        amount: testAmount,
        date: new Date().toISOString(),
        method: 'BANK_TRANSFER',
        cash_bank_id: firstAccount.id,
        reference: `INTEG_TEST_${Date.now()}`,
        notes: `Integration test payment for sale ${testSale.code}`
    };
    
    // Step 5: Test payment creation
    const success = await testIntegratedPayment(testSale.id, paymentData, dryRun);
    
    if (success) {
        console.log('🎉 Integration test completed successfully!');
        if (dryRun) {
            console.log('💡 To perform actual payment, run: runFullPaymentTest(false)');
        }
    } else {
        console.log('❌ Integration test failed - check errors above');
    }
    
    return success;
}

/**
 * Test different payment methods
 */
async function testAllPaymentMethods(saleId, cashBankId, dryRun = true) {
    console.log('🎯 Testing all payment methods');
    
    const methods = ['CASH', 'BANK_TRANSFER', 'CHECK', 'CREDIT_CARD', 'DEBIT_CARD', 'OTHER'];
    const baseAmount = 10000;
    
    for (let i = 0; i < methods.length; i++) {
        const method = methods[i];
        console.log(`\n--- Testing method: ${method} ---`);
        
        const paymentData = {
            amount: baseAmount + (i * 1000), // Varying amounts
            date: new Date().toISOString(),
            method: method,
            cash_bank_id: cashBankId,
            reference: `${method}_TEST_${Date.now()}`,
            notes: `Test payment with ${method} method`
        };
        
        const success = await testIntegratedPayment(saleId, paymentData, dryRun);
        console.log(`${method}:`, success ? '✅ PASS' : '❌ FAIL');
    }
}

// Export functions for manual use
window.integratedPaymentTest = {
    runFullPaymentTest,
    testIntegratedPayment,
    findTestSale,
    checkCashBankAccounts,
    createTestPaymentData,
    testAllPaymentMethods
};

console.log('\n📋 Integrated Payment Test Tools Loaded');
console.log('Usage:');
console.log('- runFullPaymentTest(true)  // Dry run test');
console.log('- runFullPaymentTest(false) // Live test (creates actual payment)');
console.log('- findTestSale()            // Find suitable sale for testing');
console.log('- checkCashBankAccounts()   // Check available accounts');
console.log('\n⚠️  IMPORTANT: Dry run mode is enabled by default to prevent accidental payments');
console.log('💡 Run runFullPaymentTest() to start testing');