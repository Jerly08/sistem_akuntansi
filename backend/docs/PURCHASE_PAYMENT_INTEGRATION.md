# Purchase-Payment Integration Guide

## Overview
Integrasi sistem pembayaran antara Purchase Management dan Payment Management untuk tracking hutang vendor yang optimal.

## Current Status
- ✅ Payment Management memiliki `createPayablePayment` function
- ✅ Endpoint `GET /api/payments/unpaid-bills/{vendor_id}` tersedia
- ❌ Purchase Management belum terintegrasi dengan Payment Management
- ❌ Tidak ada cross-reference tracking untuk purchase payments

## Required Integration Endpoints

### 1. Get Purchase for Payment
```
GET /api/purchases/{id}/for-payment
```
Response:
```json
{
  "purchase_id": 1,
  "bill_number": "PO/2025/01/0001", 
  "vendor": {
    "id": 1,
    "name": "PT Epson Indonesia",
    "type": "VENDOR"
  },
  "total_amount": 3385000,
  "paid_amount": 0,
  "outstanding_amount": 3385000,
  "status": "APPROVED",
  "payment_method": "CREDIT",
  "date": "2025-11-30",
  "due_date": "2025-12-30",
  "can_receive_payment": true,
  "payment_url_suggestion": "/api/purchases/1/integrated-payment"
}
```

### 2. Create Integrated Payment (Purchase → Payment Management)
```
POST /api/purchases/{id}/integrated-payment
```
Request:
```json
{
  "amount": 1000000,
  "date": "2025-01-06T10:00:00Z",
  "method": "Bank Transfer",
  "cash_bank_id": 1,
  "reference": "TRF001",
  "notes": "Cicilan pertama"
}
```

Response:
```json
{
  "payment": {
    "id": 1,
    "code": "PAY/2025/01/0001",
    "amount": 1000000,
    "status": "COMPLETED"
  },
  "updated_purchase": {
    "id": 1,
    "status": "APPROVED",
    "paid_amount": 1000000,
    "outstanding_amount": 2385000
  },
  "message": "Payment created successfully via Payment Management"
}
```

## Database Schema Changes Required

### Add Payment Tracking Fields to Purchases
```sql
-- Migration: Add payment tracking to purchases table
ALTER TABLE purchases 
ADD COLUMN IF NOT EXISTS paid_amount DECIMAL(15,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS outstanding_amount DECIMAL(15,2) DEFAULT 0;

-- Initialize outstanding_amount = total_amount for credit purchases
UPDATE purchases 
SET outstanding_amount = total_amount 
WHERE payment_method = 'CREDIT' AND outstanding_amount = 0;
```

### Purchase Payments Cross-Reference Table
```sql
CREATE TABLE purchase_payments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    purchase_id INT NOT NULL,
    payment_number VARCHAR(50),
    date DATETIME,
    amount DECIMAL(15,2),
    method VARCHAR(20),
    reference VARCHAR(100),
    notes TEXT,
    cash_bank_id INT,
    user_id INT NOT NULL,
    payment_id INT, -- Cross-reference to payments table
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (purchase_id) REFERENCES purchases(id),
    FOREIGN KEY (payment_id) REFERENCES payments(id),
    INDEX idx_purchase_payments_purchase_id (purchase_id),
    INDEX idx_purchase_payments_payment_id (payment_id)
);
```

### Payment Allocations Enhancement
```sql
-- Enhance payment_allocations to support bills
ALTER TABLE payment_allocations 
ADD COLUMN IF NOT EXISTS bill_id INT,
ADD INDEX idx_payment_allocations_bill_id (bill_id);
```

## Implementation Steps

### 1. Backend Controller Enhancement
```go
// In purchase_controller.go
func (c *PurchaseController) GetPurchaseForPayment(ctx *gin.Context) {
    purchaseID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
    // ... implementation
}

func (c *PurchaseController) CreateIntegratedPayment(ctx *gin.Context) {
    // Similar to sales integrated payment
    // Call payment service to create payment + allocation
    // Update purchase paid_amount and outstanding_amount
}
```

### 2. Frontend Service Enhancement
```typescript
// In purchaseService.ts
async createIntegratedPayment(purchaseId: number, data: PurchasePaymentRequest): Promise<any> {
  const backendData = {
    amount: data.amount,
    date: data.payment_date,
    method: data.payment_method,
    cash_bank_id: data.cash_bank_id,
    reference: data.reference || '',
    notes: data.notes || ''
  };
  const response = await api.post(`/purchases/${purchaseId}/integrated-payment`, backendData);
  return response.data;
}
```

## Workflow Rekomendasi

### Untuk Purchase CREDIT:
1. **Create Purchase** dengan `payment_method: "CREDIT"`
2. **Approve Purchase** → Status: APPROVED, Outstanding = Total Amount
3. **Record Payment** via integrated endpoint:
   - Cicilan 1: Rp 1.000.000
   - Cicilan 2: Rp 1.385.000
   - Cicilan 3: Rp 1.000.000
4. **Track di Payment Management** untuk monitoring cash flow

### Untuk Purchase CASH/TRANSFER:
1. **Create Purchase** dengan `payment_method: "CASH"` atau `"TRANSFER"`
2. **Approve Purchase** → Langsung paid, tidak perlu Payment Management
3. **Journal Entry** otomatis: Debit Expense/Asset, Credit Cash/Bank

## Benefits Integrasi

✅ **Unified Payment Tracking** - Semua payment (receivable + payable) di satu tempat
✅ **Cash Flow Management** - Monitor uang masuk dan keluar secara terpusat  
✅ **Accounts Payable Tracking** - Track hutang vendor dengan akurat
✅ **Cross-Reference Data** - Link antara purchase dan payment
✅ **Comprehensive Reporting** - Report gabungan cash flow
✅ **Journal Entries** - Otomatis untuk accounting integrity

## Testing Scenarios

### 1. Credit Purchase Payment Flow:
1. Create purchase CREDIT (Outstanding: Rp 3.385.000)
2. Create payment Rp 1.000.000 via integrated endpoint
3. Verify: Outstanding = Rp 2.385.000, Status = APPROVED
4. Create payment Rp 2.385.000 via integrated endpoint  
5. Verify: Outstanding = Rp 0, Status = PAID

### 2. Cross-Reference Verification:
1. Create payment via purchase integration
2. Verify PurchasePayment record created with PaymentID
3. Verify Payment record created with bill allocation
4. Verify journal entries created for both

This integration will provide comprehensive payment management for both sales (receivable) and purchase (payable) transactions.