# Payment Integration Guide - Sales & Payment Management

## Overview

Sistem pembayaran telah diintegrasikan antara Sales Management dan Payment Management untuk memberikan workflow pembayaran nyicil yang optimal dan centralized tracking.

## Architecture

### Dua Sistem Payment:

1. **Sales Payment (Direct)** - Pembayaran langsung dari Sales Module
2. **Payment Management (Comprehensive)** - Sistem payment dengan tracking lengkap

### Integration Features:

- ✅ Cross-reference antara kedua sistem
- ✅ Automatic journal entries untuk kedua sistem  
- ✅ Sales status update otomatis
- ✅ Unified reporting
- ✅ Payment allocation ke multiple invoices

## API Endpoints

### Sales Integration Endpoints

#### 1. Get Sale for Payment
```
GET /api/sales/{id}/for-payment
```
Response:
```json
{
  "sale_id": 1,
  "invoice_number": "INV/2025/01/0001", 
  "customer": {
    "id": 1,
    "name": "PT Global Tech",
    "type": "CUSTOMER"
  },
  "total_amount": 1665000,
  "paid_amount": 0,
  "outstanding_amount": 1665000,
  "status": "INVOICED",
  "date": "2025-01-06",
  "due_date": "2025-02-05",
  "can_receive_payment": true,
  "payment_url_suggestion": "/api/sales/1/integrated-payment"
}
```

#### 2. Create Integrated Payment (Sales → Payment Management)
```
POST /api/sales/{id}/integrated-payment
```
Request:
```json
{
  "amount": 500000,
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
    "code": "RCV/2025/01/0001",
    "amount": 500000,
    "status": "COMPLETED"
  },
  "updated_sale": {
    "id": 1,
    "status": "INVOICED",
    "paid_amount": 500000,
    "outstanding_amount": 1165000
  },
  "message": "Payment created successfully via Payment Management"
}
```

### Payment Management Integration Endpoints

#### 1. Create Sales Payment
```
POST /api/payments/sales
```
Request:
```json
{
  "sale_id": 1,
  "amount": 665000,
  "date": "2025-01-20T10:00:00Z", 
  "method": "Cash",
  "cash_bank_id": 2,
  "reference": "CASH001",
  "notes": "Cicilan kedua"
}
```

#### 2. Get Unpaid Invoices for Customer
```
GET /api/payments/sales/unpaid-invoices/{customer_id}
```

## Frontend Implementation Guide

### 1. Sales Payment Form Enhancement

```jsx
// Enhanced Payment Form Component
const EnhancedPaymentForm = ({ sale, onSuccess }) => {
  const [paymentData, setPaymentData] = useState({
    amount: '',
    date: new Date().toISOString().split('T')[0],
    method: 'Bank Transfer',
    cash_bank_id: '',
    reference: '',
    notes: ''
  });

  const handleIntegratedPayment = async () => {
    try {
      const response = await fetch(`/api/sales/${sale.id}/integrated-payment`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(paymentData)
      });
      
      const result = await response.json();
      onSuccess(result);
    } catch (error) {
      console.error('Payment failed:', error);
    }
  };

  return (
    <form onSubmit={handleIntegratedPayment}>
      {/* Form fields */}
      <div className="payment-info">
        <p>Outstanding: Rp {sale.outstanding_amount.toLocaleString()}</p>
        <p>This payment will be recorded in Payment Management system</p>
      </div>
    </form>
  );
};
```

### 2. Payment History Component

```jsx
const UnifiedPaymentHistory = ({ saleId }) => {
  const [payments, setPayments] = useState([]);

  useEffect(() => {
    // Fetch both sales payments and payment management data
    Promise.all([
      fetch(`/api/sales/${saleId}/payments`),
      fetch(`/api/payments?sale_id=${saleId}`)
    ]).then(([salesPayments, paymentMgmt]) => {
      // Combine and display unified payment history
    });
  }, [saleId]);

  return (
    <div className="payment-history">
      {payments.map(payment => (
        <div key={`${payment.type}-${payment.id}`} className="payment-item">
          <span className="badge">{payment.source}</span>
          <span>{payment.date}</span>
          <span>Rp {payment.amount.toLocaleString()}</span>
          <span>{payment.method}</span>
        </div>
      ))}
    </div>
  );
};
```

### 3. Payment Dashboard Integration

```jsx
const PaymentDashboard = () => {
  const [unifiedReport, setUnifiedReport] = useState(null);

  const fetchUnifiedReport = async (startDate, endDate) => {
    const response = await fetch(
      `/api/reports/unified-payment?start_date=${startDate}&end_date=${endDate}&format=json`
    );
    const data = await response.json();
    setUnifiedReport(data);
  };

  return (
    <div className="payment-dashboard">
      <div className="summary-cards">
        <Card title="Total Payments" value={unifiedReport?.total_amount} />
        <Card title="Payment Count" value={unifiedReport?.total_count} />
        <Card title="Sales Payments" value={unifiedReport?.sales_payments_count} />
        <Card title="Payment Mgmt" value={unifiedReport?.payment_mgmt_count} />
      </div>
      
      <div className="charts">
        <MethodBreakdownChart data={unifiedReport?.by_method} />
        <CustomerAnalysisChart data={unifiedReport?.by_customer} />
        <DailyTrendChart data={unifiedReport?.daily_breakdown} />
      </div>
    </div>
  );
};
```

## Workflow Rekomendasi

### Untuk Pembayaran Nyicil:

1. **Create Invoice** di Sales Management
2. **Use Integrated Payment** untuk semua pembayaran:
   - Cicilan 1: Rp 500.000 → via `/api/sales/1/integrated-payment`
   - Cicilan 2: Rp 665.000 → via `/api/sales/1/integrated-payment`  
   - Cicilan 3: Rp 500.000 → via `/api/sales/1/integrated-payment`
3. **Track Payments** di Payment Management dashboard
4. **Generate Unified Reports** untuk analytics lengkap

### Benefits:

- ✅ **Centralized Payment Tracking** - semua payment tercatat di Payment Management
- ✅ **Automatic Sales Status Update** - invoice status terupdate otomatis
- ✅ **Cross-Reference Data** - data tersinkronisasi antara kedua sistem
- ✅ **Comprehensive Reporting** - analytics gabungan dari kedua sistem
- ✅ **Journal Entries** - entri jurnal otomatis untuk accounting

## Database Schema Changes

### SalePayment Model - Added PaymentID field:
```sql
ALTER TABLE sale_payments ADD COLUMN payment_id INT NULL;
ALTER TABLE sale_payments ADD INDEX idx_payment_id (payment_id);
```

### Cross-Reference Tracking:
- SalePayment.PaymentID → Payment.ID (cross-reference)
- Payment dibuat dengan allocation ke Sales ID
- Journal entries dibuat untuk kedua sistem

## Testing Scenarios

### 1. Full Payment Flow:
1. Create invoice (Outstanding: Rp 1.665.000)
2. Create payment Rp 500.000 via integrated endpoint
3. Verify: Outstanding = Rp 1.165.000, Status = INVOICED
4. Create payment Rp 1.165.000 via integrated endpoint  
5. Verify: Outstanding = Rp 0, Status = PAID

### 2. Cross-Reference Verification:
1. Create payment via sales integration
2. Verify SalePayment record created with PaymentID
3. Verify Payment record created with allocation
4. Verify journal entries created for both

### 3. Unified Reporting:
1. Create payments via both systems
2. Generate unified payment report
3. Verify totals combine both sources correctly

## Error Handling

### Common Errors:
- Payment amount exceeds outstanding amount
- Sale not in INVOICED status
- Invalid cash/bank account
- Customer mismatch

### Frontend Error Handling:
```jsx
const handlePaymentError = (error) => {
  if (error.message.includes('exceeds outstanding')) {
    showError('Jumlah pembayaran melebihi saldo terutang');
  } else if (error.message.includes('not invoiced')) {
    showError('Invoice belum dibuat. Silakan buat invoice terlebih dahulu');  
  } else {
    showError('Pembayaran gagal. Silakan coba lagi');
  }
};
```

## Migration Guide

### Untuk sistem yang sudah ada:
1. Deploy backend changes dengan migration PaymentID field
2. Update frontend payment forms untuk menggunakan integrated endpoints
3. Migrate existing payment history display untuk show unified data
4. Update reporting dashboard untuk include unified payment report
5. Training user untuk workflow yang baru

Integrasi ini memberikan workflow pembayaran yang lebih komprehensif sambil menjaga backward compatibility dengan sistem yang sudah ada.
