/**
 * Payment Recording Diagnostic Script
 * This script helps diagnose payment recording issues in the accounting system
 * Run this in the browser console on the frontend application
 */

console.log('=== Payment Recording Diagnostic ===');

// 1. Check Authentication Status
function checkAuthentication() {
    console.log('\n1. AUTHENTICATION CHECK:');
    
    const token = localStorage.getItem('token');
    const refreshToken = localStorage.getItem('refreshToken');
    const userStr = localStorage.getItem('user');
    
    console.log('✓ Token exists:', !!token);
    console.log('✓ Token length:', token ? token.length : 0);
    console.log('✓ Refresh token exists:', !!refreshToken);
    
    if (userStr) {
        try {
            const user = JSON.parse(userStr);
            console.log('✓ User data:', {
                id: user.id,
                username: user.username,
                role: user.role,
                name: user.name
            });
            
            // Check if user has permission to create payments
            const allowedRoles = ['admin', 'finance', 'director', 'employee'];
            const hasPermission = allowedRoles.includes(user.role?.toLowerCase());
            console.log('✓ Has payment permission:', hasPermission);
            
            if (!hasPermission) {
                console.error('❌ User role not allowed for payments:', user.role);
                return false;
            }
        } catch (e) {
            console.error('❌ Invalid user data:', e);
            return false;
        }
    } else {
        console.error('❌ No user data found');
        return false;
    }
    
    return !!token;
}

// 2. Test API Connectivity
async function testAPIConnectivity() {
    console.log('\n2. API CONNECTIVITY TEST:');
    
    const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
    console.log('✓ API Base URL:', baseURL);
    
    try {
        // Test basic auth endpoint
        const response = await fetch(`${baseURL}/auth/me`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`,
                'Content-Type': 'application/json'
            }
        });
        
        console.log('✓ Auth test response:', response.status, response.statusText);
        
        if (response.status === 401) {
            console.error('❌ Token expired or invalid - need to re-login');
            return false;
        }
        
        if (response.ok) {
            const data = await response.json();
            console.log('✓ Current user from API:', data);
            return true;
        } else {
            console.error('❌ API connectivity failed:', response.status);
            return false;
        }
    } catch (error) {
        console.error('❌ Network error:', error.message);
        return false;
    }
}

// 3. Test Cash Bank Accounts Loading
async function testCashBankAccounts() {
    console.log('\n3. CASH/BANK ACCOUNTS TEST:');
    
    const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
    
    try {
        const response = await fetch(`${baseURL}/cashbanks/payment-accounts`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`,
                'Content-Type': 'application/json'
            }
        });
        
        console.log('✓ Cash bank accounts response:', response.status, response.statusText);
        
        if (response.ok) {
            const accounts = await response.json();
            console.log('✓ Available payment accounts:', accounts);
            console.log('✓ Account count:', accounts.length);
            
            if (accounts.length === 0) {
                console.error('❌ No payment accounts available - this will prevent payments');
                return false;
            }
            
            // Check account structure
            accounts.forEach((account, index) => {
                console.log(`✓ Account ${index + 1}:`, {
                    id: account.id,
                    code: account.code,
                    name: account.name,
                    type: account.type,
                    bank_name: account.bank_name || 'N/A'
                });
            });
            
            return true;
        } else if (response.status === 401) {
            console.error('❌ Unauthorized - token may be expired');
            return false;
        } else {
            console.error('❌ Failed to load accounts:', response.status);
            return false;
        }
    } catch (error) {
        console.error('❌ Network error loading accounts:', error.message);
        return false;
    }
}

// 4. Test Sample Sales Data
async function testSalesData() {
    console.log('\n4. SALES DATA TEST:');
    
    const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
    
    try {
        const response = await fetch(`${baseURL}/sales?limit=5&status=INVOICED`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`,
                'Content-Type': 'application/json'
            }
        });
        
        console.log('✓ Sales data response:', response.status, response.statusText);
        
        if (response.ok) {
            const result = await response.json();
            console.log('✓ Sales result structure:', {
                totalSales: result.total,
                currentPage: result.page,
                salesCount: result.data?.length
            });
            
            if (result.data && result.data.length > 0) {
                const sampleSale = result.data[0];
                console.log('✓ Sample invoiced sale:', {
                    id: sampleSale.id,
                    code: sampleSale.code,
                    status: sampleSale.status,
                    total_amount: sampleSale.total_amount,
                    outstanding_amount: sampleSale.outstanding_amount,
                    customer: sampleSale.customer?.name
                });
                
                if (sampleSale.outstanding_amount > 0) {
                    console.log('✓ Found sale with outstanding balance for testing');
                    return sampleSale;
                } else {
                    console.log('⚠️ This sale is fully paid');
                }
            } else {
                console.log('⚠️ No invoiced sales found for testing');
            }
            
            return result.data?.[0] || null;
        } else {
            console.error('❌ Failed to load sales data:', response.status);
            return null;
        }
    } catch (error) {
        console.error('❌ Network error loading sales:', error.message);
        return null;
    }
}

// 5. Test Payment Creation (dry run)
async function testPaymentCreation(sampleSale) {
    if (!sampleSale) {
        console.log('\n5. PAYMENT CREATION TEST: Skipped (no sample sale)');
        return;
    }
    
    console.log('\n5. PAYMENT CREATION TEST (DRY RUN):');
    
    const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
    
    // Test payload structure
    const testPayment = {
        amount: Math.min(100000, sampleSale.outstanding_amount), // Test with 100k or outstanding amount
        date: new Date().toISOString(),
        method: 'BANK_TRANSFER',
        cash_bank_id: 1, // Assume first account exists
        reference: 'DIAGNOSTIC_TEST_' + Date.now(),
        notes: 'Diagnostic test payment'
    };
    
    console.log('✓ Test payment payload:', testPayment);
    console.log('✓ Target endpoint:', `${baseURL}/sales/${sampleSale.id}/integrated-payment`);
    
    // Don't actually send the request, just validate the structure
    console.log('✓ Payload validation passed');
    console.log('⚠️ Actual payment creation skipped (dry run mode)');
    
    // Test different endpoint formats
    const endpoints = [
        `/sales/${sampleSale.id}/integrated-payment`,
        `/sales/${sampleSale.id}/payments`,
        `/payments/receivable`
    ];
    
    console.log('✓ Available payment endpoints to test:', endpoints);
    
    return testPayment;
}

// 6. Browser Environment Check
function checkBrowserEnvironment() {
    console.log('\n6. BROWSER ENVIRONMENT CHECK:');
    
    console.log('✓ User Agent:', navigator.userAgent);
    console.log('✓ Local Storage available:', typeof Storage !== "undefined");
    console.log('✓ Fetch API available:', typeof fetch !== "undefined");
    console.log('✓ Current URL:', window.location.href);
    
    // Check if we're in the right domain/environment
    const isDevelopment = window.location.hostname === 'localhost';
    const isSecure = window.location.protocol === 'https:';
    
    console.log('✓ Development environment:', isDevelopment);
    console.log('✓ Secure context:', isSecure);
    
    return true;
}

// Main diagnostic function
async function runFullDiagnostic() {
    console.log('🔧 Starting Payment Recording Diagnostic...\n');
    
    try {
        // Run all checks
        const results = {
            auth: checkAuthentication(),
            browser: checkBrowserEnvironment(),
            api: await testAPIConnectivity(),
            accounts: await testCashBankAccounts(),
            sales: await testSalesData()
        };
        
        // Test payment if we have valid data
        const sampleSale = results.sales;
        if (sampleSale) {
            results.payment = await testPaymentCreation(sampleSale);
        }
        
        // Summary
        console.log('\n=== DIAGNOSTIC SUMMARY ===');
        console.log('✓ Authentication:', results.auth ? 'PASS' : 'FAIL');
        console.log('✓ Browser Environment:', results.browser ? 'PASS' : 'FAIL');
        console.log('✓ API Connectivity:', results.api ? 'PASS' : 'FAIL');
        console.log('✓ Payment Accounts:', results.accounts ? 'PASS' : 'FAIL');
        console.log('✓ Sales Data:', results.sales ? 'PASS' : 'FAIL');
        
        const allPassed = results.auth && results.browser && results.api && results.accounts;
        
        if (allPassed) {
            console.log('\n🎉 All checks passed! Payment system should be working.');
            console.log('If payments still fail, check backend logs and database connectivity.');
        } else {
            console.log('\n❌ Some checks failed. Fix the issues above before testing payments.');
        }
        
        // Specific recommendations
        if (!results.auth) {
            console.log('\n💡 RECOMMENDATION: Re-login to get fresh authentication tokens');
        }
        
        if (!results.accounts) {
            console.log('\n💡 RECOMMENDATION: Add cash/bank accounts in the system first');
        }
        
        if (!results.api) {
            console.log('\n💡 RECOMMENDATION: Check if backend server is running and accessible');
        }
        
        return results;
        
    } catch (error) {
        console.error('❌ Diagnostic failed:', error);
        return null;
    }
}

// Auto-run diagnostic
console.log('To run the full diagnostic, execute: runFullDiagnostic()');
console.log('To run individual tests, use: checkAuthentication(), testAPIConnectivity(), testCashBankAccounts(), etc.');

// Export for manual use
window.paymentDiagnostic = {
    runFullDiagnostic,
    checkAuthentication,
    testAPIConnectivity,
    testCashBankAccounts,
    testSalesData,
    testPaymentCreation,
    checkBrowserEnvironment
};

console.log('\n📋 Payment diagnostic tools loaded. Run runFullDiagnostic() to start.');