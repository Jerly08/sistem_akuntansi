# Payment-SSOT Journal Integration Architecture

## Overview

Integrasi Payment System dengan SSOT (Single Source of Truth) Journal System untuk menyediakan tracking transaksi yang unified, account balance yang real-time, dan audit trail yang lengkap.

## Architecture Design

```
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│   Payment Frontend  │    │   Payment Backend   │    │   SSOT Journal      │
│   (localhost:3000/  │───►│   Controllers &     │───►│   System            │
│    payments)        │    │   Services          │    │                     │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
                                      │                           │
                                      ▼                           ▼
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│   Payment Models    │    │  Payment Journal    │    │  Account Balances   │
│   & Database        │◄───│  Factory            │───►│  (Materialized      │
│                     │    │                     │    │   View)             │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
```

## Core Components

### 1. Payment Journal Factory
**File**: `backend/services/payment_journal_factory.go`

```go
type PaymentJournalFactory struct {
    journalService *UnifiedJournalService
    accountResolver *AccountResolver
}

// CreatePaymentJournalEntry creates journal entry for payment transactions
func (pjf *PaymentJournalFactory) CreatePaymentJournalEntry(payment *Payment) (*SSOTJournalEntry, error) {
    if payment.Method == "RECEIVABLE" {
        return pjf.createReceivableJournalEntry(payment)
    }
    return pjf.createPayableJournalEntry(payment)
}
```

### 2. Enhanced Payment Service
**Integration Points**:
- Payment creation → Journal entry generation
- Payment status sync → Journal entry status
- Account balance updates via SSOT materialized view

### 3. Frontend Integration
**New Features**:
- Display journal entry reference for each payment
- Real-time account balance from SSOT system  
- Journal entry drilldown from payment details

## Integration Flow

### Payment Creation Flow
```
1. User creates payment via frontend form
   ↓
2. PaymentService.CreatePayment()
   ↓
3. PaymentJournalFactory.CreatePaymentJournalEntry()
   ↓
4. UnifiedJournalService.CreateJournalEntry()
   ↓
5. SSOT Journal Entry created in unified_journal_ledger
   ↓
6. Account balances auto-updated via materialized view triggers
   ↓
7. Payment status updated based on journal entry status
   ↓
8. Frontend refreshes with updated data
```

## Journal Entry Patterns

### Receivable Payment (Customer pays invoice)
```
Dr. Cash/Bank Account        1,000,000
    Cr. Accounts Receivable             1,000,000
```

### Payable Payment (Company pays vendor bill)
```
Dr. Accounts Payable         1,000,000
    Cr. Cash/Bank Account               1,000,000
```

### Multi-Currency Payment
```
Dr. Cash/Bank Account (IDR)  14,000,000
Dr. Exchange Rate Loss          100,000  
    Cr. Accounts Receivable (USD)       1,000,000 (@ 14,100)
```

## Database Schema Updates

### Payment Model Enhancement
```sql
ALTER TABLE payments 
ADD COLUMN journal_entry_id BIGINT REFERENCES unified_journal_ledger(id),
ADD INDEX idx_payments_journal_entry (journal_entry_id);
```

### SSOT Journal Enhancement  
```sql
-- Add payment-specific source types
UPDATE unified_journal_ledger 
SET source_type = 'PAYMENT_RECEIVABLE' 
WHERE source_type = 'PAYMENT' AND EXISTS (
    SELECT 1 FROM payments p 
    WHERE p.id = source_id AND p.method = 'RECEIVABLE'
);
```

## API Endpoints

### Enhanced Payment Endpoints

#### Create Payment with Journal Integration
```
POST /api/payments/enhanced-with-journal
```

**Request**:
```json
{
  "contact_id": 1,
  "amount": 1000000,
  "date": "2025-01-20",
  "method": "RECEIVABLE",
  "cash_bank_id": 1,
  "reference": "INV-001-PAYMENT",
  "notes": "Payment for Invoice INV-001",
  "auto_create_journal": true
}
```

**Response**:
```json
{
  "payment": {
    "id": 123,
    "code": "RCV-2025-001",
    "amount": 1000000,
    "status": "COMPLETED",
    "journal_entry_id": 456
  },
  "journal_entry": {
    "id": 456,
    "entry_number": "JE-PAYMENT-2025-001",
    "status": "POSTED",
    "total_debit": 1000000,
    "total_credit": 1000000,
    "is_balanced": true
  },
  "account_updates": [
    {
      "account_id": 1101,
      "account_code": "CASH",
      "old_balance": 5000000,
      "new_balance": 6000000,
      "change": 1000000
    },
    {
      "account_id": 1201,
      "account_code": "ACCOUNTS_RECEIVABLE", 
      "old_balance": 10000000,
      "new_balance": 9000000,
      "change": -1000000
    }
  ]
}
```

#### Get Payment with Journal Details
```
GET /api/payments/{id}/with-journal
```

**Response**:
```json
{
  "payment": {
    "id": 123,
    "code": "RCV-2025-001",
    "amount": 1000000
  },
  "journal_entry": {
    "id": 456,
    "entry_number": "JE-PAYMENT-2025-001",
    "lines": [
      {
        "account_code": "1101-CASH",
        "account_name": "Kas Perusahaan",
        "debit_amount": 1000000,
        "credit_amount": 0
      },
      {
        "account_code": "1201-AR", 
        "account_name": "Piutang Usaha",
        "debit_amount": 0,
        "credit_amount": 1000000
      }
    ]
  }
}
```

### SSOT Journal Integration Endpoints

#### Get Account Balance Real-time
```
GET /api/journals/account-balances/real-time
```

#### Get Payment Journal Entries
```
GET /api/journals/payment-entries?payment_id={id}
```

## Frontend Integration

### Enhanced Payment Form
```tsx
// Enhanced form with journal entry preview
const PaymentFormWithJournal = ({ onSubmit }) => {
  const [journalPreview, setJournalPreview] = useState(null);

  const handleAmountChange = async (amount) => {
    // Preview journal entry before submission
    const preview = await paymentService.previewJournalEntry({
      amount,
      method: formData.method,
      contact_id: formData.contact_id
    });
    setJournalPreview(preview);
  };

  return (
    <Form>
      {/* Payment form fields */}
      
      {/* Journal Entry Preview */}
      {journalPreview && (
        <JournalEntryPreview 
          entry={journalPreview}
          title="Journal Entry (Preview)"
        />
      )}
    </Form>
  );
};
```

### Payment Dashboard with Account Balances
```tsx
const PaymentDashboardWithSSOT = () => {
  const [accountBalances, setAccountBalances] = useState([]);

  useEffect(() => {
    // Subscribe to real-time balance updates via WebSocket
    const ws = new WebSocket('/api/journals/account-balances/ws');
    ws.onmessage = (event) => {
      const updatedBalances = JSON.parse(event.data);
      setAccountBalances(updatedBalances);
    };
  }, []);

  return (
    <Dashboard>
      {/* Payment summary cards */}
      
      {/* Real-time account balances */}
      <AccountBalancePanel balances={accountBalances} />
      
      {/* Payment transactions with journal references */}
      <PaymentTransactionsWithJournal />
    </Dashboard>
  );
};
```

## Performance Considerations

### Database Optimizations
```sql
-- Optimize payment-journal queries
CREATE INDEX CONCURRENTLY idx_unified_journal_source_payment 
ON unified_journal_ledger(source_type, source_id) 
WHERE source_type IN ('PAYMENT_RECEIVABLE', 'PAYMENT_PAYABLE');

-- Optimize account balance materialized view refresh
CREATE INDEX CONCURRENTLY idx_account_balances_updated_accounts
ON account_balances(last_updated DESC, account_id)
WHERE current_balance != 0;
```

### Caching Strategy
```go
// Cache frequently accessed account balances
type BalanceCache struct {
    cache   map[uint64]CachedBalance
    timeout time.Duration
    mutex   sync.RWMutex
}

func (bc *BalanceCache) GetAccountBalance(accountID uint64) (decimal.Decimal, error) {
    bc.mutex.RLock()
    if cached, exists := bc.cache[accountID]; exists && !cached.IsExpired() {
        bc.mutex.RUnlock()
        return cached.Balance, nil
    }
    bc.mutex.RUnlock()
    
    // Fetch from materialized view if not in cache
    return bc.fetchAndCache(accountID)
}
```

## Error Handling

### Transaction Rollback Strategy
```go
func (pjf *PaymentJournalFactory) CreatePaymentWithJournal(payment *Payment) error {
    return pjf.db.Transaction(func(tx *gorm.DB) error {
        // Step 1: Create payment record
        if err := tx.Create(payment).Error; err != nil {
            return fmt.Errorf("failed to create payment: %w", err)
        }

        // Step 2: Create journal entry
        journalEntry, err := pjf.createJournalEntry(tx, payment)
        if err != nil {
            return fmt.Errorf("failed to create journal entry: %w", err)
        }

        // Step 3: Update payment with journal reference
        payment.JournalEntryID = &journalEntry.ID
        if err := tx.Save(payment).Error; err != nil {
            return fmt.Errorf("failed to link payment to journal: %w", err)
        }

        return nil
    })
}
```

### Frontend Error Handling
```tsx
const handlePaymentSubmission = async (paymentData) => {
  try {
    const result = await paymentService.createPaymentWithJournal(paymentData);
    
    // Success: show confirmation with journal details
    showSuccessNotification({
      title: 'Payment Created Successfully',
      message: `Journal Entry ${result.journal_entry.entry_number} has been posted`,
      journalDetails: result.journal_entry
    });
    
  } catch (error) {
    if (error.code === 'JOURNAL_UNBALANCED') {
      showErrorDialog({
        title: 'Journal Entry Error',
        message: 'Unable to create balanced journal entry. Please check account configuration.',
        technicalDetails: error.details
      });
    } else if (error.code === 'INSUFFICIENT_BALANCE') {
      showWarningDialog({
        title: 'Insufficient Balance',
        message: 'This payment will create negative balance in cash account. Continue?',
        onConfirm: () => submitPaymentWithForce(paymentData)
      });
    } else {
      showGenericError(error);
    }
  }
};
```

## Testing Strategy

### Integration Tests
```go
func TestPaymentJournalIntegration(t *testing.T) {
    // Test 1: Receivable payment creates correct journal entry
    payment := createTestReceivablePayment(1000000)
    
    factory := NewPaymentJournalFactory(db, journalService)
    journalEntry, err := factory.CreatePaymentJournalEntry(payment)
    
    assert.NoError(t, err)
    assert.Equal(t, models.SSOTSourceTypePayment, journalEntry.SourceType)
    assert.Equal(t, payment.ID, *journalEntry.SourceID)
    assert.True(t, journalEntry.IsBalanced)
    
    // Verify journal lines
    assert.Len(t, journalEntry.Lines, 2)
    
    // Verify account balance updates
    cashBalance := getAccountBalance(CASH_ACCOUNT_ID)
    arBalance := getAccountBalance(AR_ACCOUNT_ID)
    
    assert.Equal(t, decimal.NewFromInt(1000000), cashBalance.Delta)
    assert.Equal(t, decimal.NewFromInt(-1000000), arBalance.Delta)
}
```

### Frontend Integration Tests
```tsx
describe('Payment-Journal Integration', () => {
  test('should display journal entry reference after payment creation', async () => {
    render(<PaymentForm />);
    
    // Fill form
    fireEvent.change(screen.getByLabelText('Amount'), { target: { value: '1000000' } });
    fireEvent.click(screen.getByText('Create Payment'));
    
    // Wait for API response
    await waitFor(() => {
      expect(screen.getByText(/Journal Entry: JE-PAYMENT-/)).toBeInTheDocument();
    });
    
    // Verify journal details are displayed
    expect(screen.getByText('Dr. Cash Account: Rp 1,000,000')).toBeInTheDocument();
    expect(screen.getByText('Cr. Accounts Receivable: Rp 1,000,000')).toBeInTheDocument();
  });
});
```

## Monitoring & Analytics

### Business Intelligence Integration
```sql
-- Payment-Journal Analytics View
CREATE VIEW payment_journal_analytics AS
SELECT 
    p.id as payment_id,
    p.code as payment_code,
    p.amount as payment_amount,
    p.method as payment_method,
    p.date as payment_date,
    p.status as payment_status,
    
    j.id as journal_id,
    j.entry_number as journal_number,
    j.status as journal_status,
    j.posted_at as journal_posted_at,
    
    -- Account balance changes
    COALESCE(cash_line.debit_amount - cash_line.credit_amount, 0) as cash_impact,
    COALESCE(ar_line.credit_amount - ar_line.debit_amount, 0) as receivable_impact
    
FROM payments p
LEFT JOIN unified_journal_ledger j ON p.journal_entry_id = j.id
LEFT JOIN unified_journal_lines cash_line ON j.id = cash_line.journal_id 
    AND cash_line.account_id IN (SELECT id FROM accounts WHERE type = 'CASH')
LEFT JOIN unified_journal_lines ar_line ON j.id = ar_line.journal_id 
    AND ar_line.account_id IN (SELECT id FROM accounts WHERE type = 'ACCOUNTS_RECEIVABLE');
```

### Dashboard Metrics
```tsx
const PaymentJournalMetrics = () => {
  const [metrics, setMetrics] = useState(null);

  useEffect(() => {
    // Fetch real-time metrics
    paymentService.getJournalIntegrationMetrics().then(setMetrics);
  }, []);

  return (
    <MetricsDashboard>
      <MetricCard 
        title="Payments with Journal Entries" 
        value={`${metrics?.journal_coverage_rate}%`}
        trend={metrics?.journal_coverage_trend}
      />
      <MetricCard 
        title="Journal Entry Creation Success Rate" 
        value={`${metrics?.journal_success_rate}%`}
        alert={metrics?.journal_success_rate < 98}
      />
      <MetricCard 
        title="Account Balance Accuracy" 
        value={metrics?.balance_accuracy_score}
        description="Based on journal vs materialized view comparison"
      />
    </MetricsDashboard>
  );
};
```

## Migration Plan

### Phase 1: Foundation (Week 1-2)
1. ✅ SSOT Journal System already implemented
2. 🔄 Create PaymentJournalFactory service
3. 🔄 Update Payment models with journal_entry_id reference
4. 🔄 Create database migrations

### Phase 2: Backend Integration (Week 3-4) 
1. 🔄 Implement enhanced payment service with journal integration
2. 🔄 Update payment controllers to use journal factory
3. 🔄 Add journal entry endpoints for payments
4. 🔄 Implement error handling and rollback mechanisms

### Phase 3: Frontend Integration (Week 5-6)
1. 🔄 Update payment forms to show journal entry preview
2. 🔄 Add journal entry details to payment dashboard
3. 🔄 Implement real-time account balance display
4. 🔄 Update payment details modal with journal information

### Phase 4: Testing & Optimization (Week 7-8)
1. 🔄 Comprehensive integration testing
2. 🔄 Performance optimization and caching
3. 🔄 User acceptance testing
4. 🔄 Documentation and training materials

## Success Metrics

### Technical KPIs
- ✅ **100% Journal Coverage**: All payments have corresponding journal entries
- ✅ **<200ms Response Time**: Payment creation with journal entry
- ✅ **99.9% Data Consistency**: Account balances match between payment and journal systems
- ✅ **Zero Failed Transactions**: All payment-journal operations succeed or rollback completely

### Business KPIs  
- ✅ **Real-time Reporting**: Account balances update immediately after payment
- ✅ **Complete Audit Trail**: Full traceability from payment to journal to account balance
- ✅ **Simplified Reconciliation**: Automated account balance reconciliation
- ✅ **Enhanced User Experience**: Users can see journal impact of their payment actions

---

Arsitektur ini menyediakan integrasi yang seamless antara Payment System dan SSOT Journal System, memastikan data consistency, audit trail yang lengkap, dan user experience yang optimal.