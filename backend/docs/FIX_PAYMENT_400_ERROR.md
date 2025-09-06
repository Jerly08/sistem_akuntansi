# Fix: Sales-Payment Integration 400 Error

## Problem Summary
Payment creation through Sales Management was failing with a **400 Bad Request** error when calling `/api/v1/sales/{id}/integrated-payment`.

## Root Cause
**Field Name Mismatch** between frontend and backend expectations:

### Frontend was sending:
```json
{
  "sale_id": 13,               // ❌ Not expected by backend
  "amount": 208125,            // ✅ Correct
  "payment_date": "2025-09-06T00:00:00.000Z",  // ❌ Backend expects "date"
  "payment_method": "BANK_TRANSFER",           // ❌ Backend expects "method"
  "reference": "test005",      // ✅ Correct
  "notes": "test005",         // ✅ Correct
  "cash_bank_id": 2,          // ✅ Correct
  "account_id": 2             // ❌ Not expected by backend
}
```

### Backend expects:
```go
type IntegratedPaymentRequest struct {
    Amount     float64   `json:"amount" binding:"required,min=0"`
    Date       time.Time `json:"date" binding:"required"`
    Method     string    `json:"method" binding:"required"`
    CashBankID uint      `json:"cash_bank_id" binding:"required"`
    Reference  string    `json:"reference"`
    Notes      string    `json:"notes"`
}
```

## Solution Applied

### 1. Fixed Frontend Field Mapping
**File**: `frontend/src/services/salesService.ts`

**Before**:
```typescript
const backendData = {
  sale_id: saleId,
  amount: data.amount,
  payment_date: data.payment_date,      // ❌ Wrong field name
  payment_method: data.payment_method,  // ❌ Wrong field name
  reference: data.reference,
  notes: data.notes,
  cash_bank_id: data.cash_bank_id,
  account_id: data.account_id           // ❌ Not expected
};
```

**After**:
```typescript
const backendData = {
  amount: data.amount,                  // ✅ Correct
  date: data.payment_date,             // ✅ Fixed: maps to backend "date"
  method: data.payment_method,         // ✅ Fixed: maps to backend "method"
  cash_bank_id: data.cash_bank_id,    // ✅ Correct
  reference: data.reference || '',     // ✅ Correct with fallback
  notes: data.notes || ''             // ✅ Correct with fallback
};
```

### 2. Enhanced Backend Error Logging
**File**: `backend/controllers/sales_controller.go`

Added detailed error messages to help with future debugging:
```go
if err := c.ShouldBindJSON(&request); err != nil {
    log.Printf("Payment creation validation error for sale %d: %v", id, err)
    c.JSON(http.StatusBadRequest, gin.H{
        "error": "Invalid request data",
        "details": err.Error(),
        "expected_fields": map[string]string{
            "amount": "number (required, min=0)",
            "date": "datetime string (required, ISO format)",
            "method": "string (required)",
            "cash_bank_id": "number (required)",
            "reference": "string (optional)",
            "notes": "string (optional)",
        },
    })
    return
}
```

### 3. Added Request Logging
Added logging for successful request parsing to help with future debugging:
```go
log.Printf("Received integrated payment request for sale %d: amount=%.2f, method=%s, cash_bank_id=%d", 
    id, request.Amount, request.Method, request.CashBankID)
```

## Testing the Fix

### 1. Rebuild Frontend
```bash
# In the frontend directory
npm run build
# or for development
npm run dev
```

### 2. Restart Backend
The Go backend should automatically reload, or restart it manually.

### 3. Test Payment Creation
1. Go to Sales Management
2. Find an invoiced sale
3. Click "Record Payment"
4. Fill in the payment details
5. Submit

### Expected Result
- ✅ Payment should be created successfully
- ✅ Success message: "Payment has been recorded successfully and will appear in Payment Management"
- ✅ Payment should appear in Payment Management module
- ✅ Backend logs should show successful payment creation

## Verification Commands

### Check Recent Payments in Database:
```sql
SELECT p.*, c.name as customer_name 
FROM payments p 
JOIN contacts c ON p.contact_id = c.id 
ORDER BY p.created_at DESC 
LIMIT 5;
```

### Check Debug Endpoint (Admin only):
```bash
GET /api/v1/payments/debug/recent
```

## Files Modified
1. `frontend/src/services/salesService.ts` - Fixed field mapping
2. `backend/controllers/sales_controller.go` - Enhanced error logging

## Related Documentation
- `backend/docs/PAYMENT_INTEGRATION_TESTING.md` - Testing guide
- `backend/scripts/debug_payment_integration.sql` - Database debugging queries

## Status
✅ **FIXED**: Field mapping corrected
✅ **ENHANCED**: Better error logging added
✅ **READY**: For testing and verification
