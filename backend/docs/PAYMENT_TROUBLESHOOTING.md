# Payment Recording Troubleshooting Guide

This guide helps diagnose and fix payment recording failures in the accounting system.

## Common Issues and Solutions

### 1. Authentication Issues

**Symptoms:**
- HTTP 401 Unauthorized errors
- "Session expired" messages
- Payment form doesn't load accounts

**Solutions:**

#### A. Check User Role Permissions
The payment system has role-based access control. Check if the current user has the right permissions:

```javascript
// In browser console
const user = JSON.parse(localStorage.getItem('user') || '{}');
console.log('Current user role:', user.role);

// Allowed roles for payment creation:
const allowedRoles = ['admin', 'finance', 'director', 'employee'];
const hasPermission = allowedRoles.includes(user.role?.toLowerCase());
console.log('Has payment permission:', hasPermission);
```

#### B. Token Refresh Issues
```javascript
// Check token validity
const token = localStorage.getItem('token');
const refreshToken = localStorage.getItem('refreshToken');
console.log('Token exists:', !!token, 'Refresh token exists:', !!refreshToken);

// If tokens are invalid, clear and re-login
if (!token || !refreshToken) {
    localStorage.clear();
    window.location.href = '/login';
}
```

### 2. Cash/Bank Account Loading Issues

**Symptoms:**
- "No payment accounts available" error
- Empty dropdown in payment form
- Account loading spinner never finishes

**Solutions:**

#### A. Verify Cash Bank Service
```javascript
// Test cashbank service directly
fetch('/api/cashbanks/payment-accounts', {
    headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
    }
}).then(response => {
    console.log('CashBank response status:', response.status);
    return response.json();
}).then(data => {
    console.log('Payment accounts:', data);
    if (data.length === 0) {
        console.error('No payment accounts configured!');
    }
}).catch(error => {
    console.error('CashBank loading error:', error);
});
```

#### B. Add Missing Cash/Bank Accounts
If no payment accounts exist, add them through the admin panel:
1. Go to Master Data → Cash & Bank
2. Add at least one cash or bank account
3. Ensure accounts are marked as "Active"
4. Set correct account codes and names

### 3. Payment Form Validation Issues

**Symptoms:**
- Form validation errors
- Required field messages
- Amount validation problems

**Solutions:**

#### A. Amount Validation Fix
```typescript
// In PaymentForm component, ensure amount validation allows decimals
const handleAmountChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const inputValue = e.target.value;
    
    // Allow numbers, dots, commas for currency formatting
    const allowedCharsRegex = /^[Rp\d.,\s]*$/;
    if (!allowedCharsRegex.test(inputValue)) {
        return;
    }
    
    const numericValue = parseRupiah(inputValue);
    
    // Validate against outstanding amount
    if (sale && numericValue > sale.outstanding_amount) {
        toast({
            title: 'Amount Exceeds Outstanding',
            description: `Maximum amount is ${salesService.formatCurrency(sale.outstanding_amount)}`,
            status: 'warning',
            duration: 3000
        });
        return;
    }
    
    setValue('amount', numericValue);
    setDisplayAmount(formatRupiah(numericValue));
};
```

#### B. Date Validation Fix
```typescript
// Ensure date is not in the future
const validatePaymentDate = (dateString: string): boolean => {
    const paymentDate = new Date(dateString);
    const today = new Date();
    today.setHours(23, 59, 59, 999); // End of today
    
    return paymentDate <= today;
};
```

### 4. Backend Payment Processing Issues

**Symptoms:**
- HTTP 500 Internal Server Error
- Payment recorded but journal entries missing
- Database constraint errors

**Solutions:**

#### A. Check Backend Logs
Look for errors in the Go backend logs:
```bash
# In your backend directory
go run main.go 2>&1 | grep -i error
# Or check your log files
tail -f logs/app.log | grep -i payment
```

#### B. Database Connection Issues
Ensure database connection is working:
```sql
-- Test database connectivity
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM sales WHERE status = 'INVOICED';
SELECT COUNT(*) FROM cashbanks WHERE is_active = true;
```

#### C. Journal Entry Account Configuration
Ensure Chart of Accounts is properly configured:
```sql
-- Check for required account types
SELECT code, name, type FROM accounts 
WHERE type IN ('ASSET', 'LIABILITY', 'RECEIVABLE') 
AND is_active = true;
```

### 5. Frontend API Call Issues

**Symptoms:**
- Network errors
- Timeout errors
- CORS issues

**Solutions:**

#### A. API Base URL Configuration
```typescript
// Check if API URL is correctly configured
const API_URL = process.env.NEXT_PUBLIC_API_URL;
console.log('API Base URL:', API_URL);

// Common configurations:
// Development: http://localhost:8080/api
// Production: https://your-domain.com/api
```

#### B. Request Timeout Issues
```typescript
// Increase timeout for payment operations
const api = axios.create({
    baseURL: API_BASE_URL,
    timeout: 30000, // 30 seconds for payment operations
    headers: {
        'Content-Type': 'application/json',
    },
});
```

### 6. Payment Endpoint Issues

**Symptoms:**
- 404 Not Found on payment endpoints
- Wrong endpoint being called
- Payload format mismatch

**Solutions:**

#### A. Verify Endpoint URLs
```javascript
// Test different payment endpoints
const endpoints = [
    '/sales/{id}/integrated-payment',  // Integrated payment (recommended)
    '/sales/{id}/payments',            // Sales payment only
    '/payments/receivable'             // Payment management only
];

// Use integrated payment for best results
```

#### B. Correct Payload Format
```typescript
// For integrated payment endpoint
const paymentData = {
    amount: 100000,                    // Required: numeric amount
    date: '2025-01-15T10:30:00Z',     // Required: ISO datetime
    method: 'BANK_TRANSFER',          // Required: payment method
    cash_bank_id: 1,                  // Required: bank account ID
    reference: 'REF123',              // Optional: reference number
    notes: 'Payment notes'            // Optional: notes
};
```

## Diagnostic Tools

### 1. Browser Console Diagnostic
Use the diagnostic script provided in `payment_diagnostic.js`:

```javascript
// Load the diagnostic script and run
runFullDiagnostic();
```

### 2. Network Tab Debugging
1. Open browser Developer Tools (F12)
2. Go to Network tab
3. Try to record a payment
4. Look for failed requests (red entries)
5. Check request/response details

### 3. Console Log Debugging
Enable detailed logging in the payment form:

```typescript
// Add to PaymentForm component
const onSubmit = async (data: PaymentFormData) => {
    console.log('Payment form submission:', data);
    
    try {
        const paymentData = {
            payment_date: new Date(data.date).toISOString(),
            amount: data.amount,
            payment_method: data.method,
            reference: data.reference || '',
            cash_bank_id: data.account_id,
            notes: data.notes || ''
        };
        
        console.log('Sending payment data:', paymentData);
        console.log('Target URL:', `/sales/${sale.id}/integrated-payment`);
        
        const result = await salesService.createIntegratedPayment(sale.id, paymentData);
        console.log('Payment creation result:', result);
        
        // ... rest of success handling
    } catch (error) {
        console.error('Payment creation error:', error);
        console.log('Error details:', {
            status: error.response?.status,
            message: error.response?.data?.message,
            data: error.response?.data
        });
        
        // ... error handling
    }
};
```

## Step-by-Step Debugging Process

### 1. Run Diagnostic Script
```javascript
runFullDiagnostic();
```

### 2. Check Authentication
```javascript
checkAuthentication();
```

### 3. Test API Connectivity
```javascript
testAPIConnectivity();
```

### 4. Verify Account Loading
```javascript
testCashBankAccounts();
```

### 5. Test with Sample Data
```javascript
testSalesData();
```

### 6. Monitor Network Requests
- Open DevTools → Network tab
- Try payment recording
- Look for failed requests

### 7. Check Backend Logs
```bash
# If using systemd
sudo journalctl -u accounting-backend -f

# If running directly
tail -f /path/to/your/logs/app.log
```

## Quick Fixes Checklist

- [ ] User is logged in with valid token
- [ ] User role allows payment creation
- [ ] Cash/bank accounts are configured and active
- [ ] Backend server is running and accessible
- [ ] Database connection is working
- [ ] Chart of accounts is properly set up
- [ ] No browser console errors
- [ ] Network requests are successful
- [ ] Payment amounts are within valid ranges
- [ ] Payment dates are not in the future

## Common Error Messages and Solutions

| Error Message | Cause | Solution |
|---------------|-------|----------|
| "Session expired" | Invalid/expired token | Re-login to get fresh token |
| "No payment accounts" | No cash/bank accounts | Add accounts in Master Data |
| "Amount exceeds outstanding" | Payment > outstanding balance | Check invoice amounts |
| "Unauthorized" | Wrong user role/permissions | Check user role settings |
| "Network Error" | Backend not accessible | Check backend server status |
| "Validation Error" | Missing required fields | Fill all required form fields |

## Contact Support

If issues persist after following this guide:

1. Collect diagnostic information using the script
2. Check browser console for errors
3. Export browser network logs
4. Check backend logs
5. Document the exact steps to reproduce the issue
6. Contact your system administrator or development team

---

**Note**: This troubleshooting guide covers the most common payment recording issues. For custom implementations or specific business requirements, additional debugging may be needed.