# Sales-Payment Management Integration Testing Guide

## Overview
This guide helps you verify that payments created through Sales Management appear correctly in Payment Management.

## Current Integration Status
✅ **CONFIRMED**: Sales Management is using the correct integrated endpoint (`/sales/:id/integrated-payment`)
✅ **CONFIRMED**: Backend integration is properly implemented
❓ **TO VERIFY**: Payment records are appearing in Payment Management frontend

## Testing Steps

### Step 1: Test Payment Creation through Sales Management

1. **Create an Invoice**:
   - Go to Sales Management
   - Create a new sale and confirm it (status should be "INVOICED")
   - Note the invoice number and outstanding amount

2. **Record a Payment**:
   - In Sales Management, click "Record Payment" for the invoice
   - Fill in the payment details:
     - Date: Today's date
     - Amount: Part or full payment amount
     - Method: Bank Transfer
     - Bank Account: Select any available account
     - Reference: Test reference number
     - Notes: "Test payment integration"
   - Submit the payment

3. **Expected Results**:
   - Success message: "Payment has been recorded successfully and will appear in Payment Management"
   - Sale status should update (PAID if full payment, or remain INVOICED if partial)
   - Outstanding amount should decrease

### Step 2: Verify in Payment Management

1. **Check Payment Management List**:
   - Navigate to Payment Management module
   - Look for the payment you just created
   - The payment should appear with:
     - Customer name matching the sale
     - Amount matching what you entered
     - Status: COMPLETED
     - Method: Bank Transfer (or whatever you selected)
     - Notes: Should include invoice reference

2. **If Payment is Missing**:
   - Check the debug endpoint: `GET /api/payments/debug/recent`
   - This will show all recent payments in the database
   - Look for your payment in the results

### Step 3: Database Verification

Run the debug SQL script to verify data integrity:

```sql
-- Run this query in your MySQL/database client
-- File: backend/scripts/debug_payment_integration.sql

-- Check payments table
SELECT 'RECENT PAYMENTS' as check_type;
SELECT 
    p.id, p.code, p.contact_id, c.name as customer_name,
    p.date, p.amount, p.method, p.status, p.notes, p.created_at
FROM payments p
JOIN contacts c ON p.contact_id = c.id
WHERE p.created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)
ORDER BY p.created_at DESC;

-- Check cross-references
SELECT 'CROSS-REFERENCES' as check_type;
SELECT 
    p.id as payment_id, p.code as payment_code,
    sp.id as sale_payment_id, sp.sale_id,
    s.invoice_number, s.code as sale_code,
    p.amount as payment_amount, sp.amount as sale_payment_amount
FROM payments p
LEFT JOIN sale_payments sp ON sp.payment_id = p.id
LEFT JOIN sales s ON sp.sale_id = s.id
WHERE p.created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY)
ORDER BY p.created_at DESC;
```

### Step 4: Frontend Debugging

If payments are in the database but not showing in Payment Management frontend:

1. **Check Network Tab**:
   - Open browser DevTools → Network tab
   - Refresh Payment Management page
   - Look for the API call to `/api/payments`
   - Verify the response contains your payment

2. **Check Console for Errors**:
   - Look for JavaScript errors in Console tab
   - Common issues: Authentication errors, filtering issues

3. **Test Direct API Call**:
   ```bash
   # Test the payments API directly
   curl -H "Authorization: Bearer YOUR_TOKEN" \
        http://localhost:8080/api/payments?limit=20
   ```

## Troubleshooting Common Issues

### Issue 1: Payment Not Created in Database
**Symptoms**: Payment creation fails, error message in Sales Management
**Solution**: Check backend logs for transaction errors

### Issue 2: Payment Created but Not in Payment Management List
**Symptoms**: Payment exists in database but not in frontend list
**Causes**: 
- Frontend filtering issues
- Authentication problems
- API pagination issues
**Solution**: Use debug endpoint, check network requests

### Issue 3: Cross-Reference Missing
**Symptoms**: Payment exists but no link to sale
**Solution**: Check `sale_payments` table for cross-reference records

## API Endpoints for Testing

### Debug Endpoints (Admin Only)
- `GET /api/payments/debug/recent` - Recent payments for debugging
- `GET /api/payments?limit=50` - All payments with increased limit

### Regular Endpoints
- `GET /api/payments` - Standard payment list
- `GET /api/payments/{id}` - Single payment details
- `POST /api/sales/{id}/integrated-payment` - Create integrated payment

## Expected Database Schema

### payments table
```sql
CREATE TABLE payments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    code VARCHAR(20) UNIQUE NOT NULL,
    contact_id INT NOT NULL,
    user_id INT NOT NULL,
    date DATETIME,
    amount DECIMAL(15,2) DEFAULT 0,
    method VARCHAR(20),
    reference VARCHAR(50),
    status VARCHAR(20),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### sale_payments table (Cross-reference)
```sql
CREATE TABLE sale_payments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    sale_id INT NOT NULL,
    payment_number VARCHAR(50),
    date DATETIME,
    amount DECIMAL(15,2),
    method VARCHAR(20),
    reference VARCHAR(100),
    notes TEXT,
    cash_bank_id INT,
    user_id INT NOT NULL,
    payment_id INT, -- Cross-reference to payments table
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### payment_allocations table
```sql
CREATE TABLE payment_allocations (
    id INT PRIMARY KEY AUTO_INCREMENT,
    payment_id INT NOT NULL,
    invoice_id INT,
    bill_id INT,
    allocated_amount DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Success Criteria

✅ **Payment created in payments table**
✅ **Cross-reference created in sale_payments table**
✅ **Payment allocation created in payment_allocations table**
✅ **Sale outstanding amount updated**
✅ **Payment appears in Payment Management frontend**
✅ **Payment shows correct customer, amount, and status**

## Contact

If you continue to have issues after following this guide:
1. Run the SQL debug script and share the results
2. Check the debug API endpoint results
3. Review backend logs during payment creation
4. Verify frontend network requests in DevTools
