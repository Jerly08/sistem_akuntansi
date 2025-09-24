# 🔄 SSOT Journal System Integration dengan Cash-Bank Frontend

## Executive Summary

Dokumen ini merancang integrasi antara **SSOT (Single Source of Truth) Journal System** dengan **CashBank System** untuk memberikan data terpadu pada halaman frontend `http://localhost:3000/cash-bank`.

## 1. OVERVIEW INTEGRASI

### 1.1 Tujuan Integrasi
- ✅ **Unified Data View**: Menampilkan data cashbank beserta journal entries terkait
- ✅ **Real-time Balance**: Sinkronisasi balance antara CashBank dan SSOT Journal
- ✅ **Complete Audit Trail**: Riwayat lengkap transaksi dari perspektif cashbank dan journal
- ✅ **Enhanced Reconciliation**: Rekonsiliasi otomatis antara dua sistem

### 1.2 Arsitektur Current State
```
Frontend (http://localhost:3000/cash-bank)
    ↓ API Calls
CashBankController → CashBankService → CashBankRepository
    ↓ Independent
UnifiedJournalController → UnifiedJournalService → SSOT Tables
```

### 1.3 Arsitektur Target State
```
Frontend (http://localhost:3000/cash-bank)
    ↓ API Calls
CashBankIntegratedController → CashBankIntegratedService
    ↓ Orchestrates
    ├── CashBankService (existing)
    └── UnifiedJournalService (existing)
    ↓ Returns
Integrated Response (CashBank + SSOT Journal Data)
```

## 2. ENDPOINT INTEGRASI YANG DIREKOMENDASIKAN

### 2.1 Enhanced Account Details with Journal Integration
```http
GET /api/cashbank/accounts/:id/integrated
```

**Response Structure:**
```json
{
  "status": "success",
  "data": {
    "account": {
      "id": 1,
      "code": "CB-001",
      "name": "Bank Mandiri",
      "type": "BANK",
      "balance": 5000000.00,
      "currency": "IDR"
    },
    "ssot_balance": 5000000.00,
    "balance_difference": 0.00,
    "recent_transactions": [
      {
        "id": 123,
        "amount": 500000.00,
        "type": "DEPOSIT",
        "date": "2025-09-20T10:00:00Z",
        "journal_entry_id": 456,
        "journal_entry_number": "JE-2025-09-0001"
      }
    ],
    "related_journal_entries": [
      {
        "id": 456,
        "entry_number": "JE-2025-09-0001",
        "description": "Bank deposit from sales payment",
        "total_debit": 500000.00,
        "total_credit": 500000.00,
        "status": "POSTED",
        "lines": [
          {
            "account_id": 1,
            "account_name": "Bank Mandiri",
            "debit_amount": 500000.00,
            "credit_amount": 0.00
          },
          {
            "account_id": 2,
            "account_name": "Piutang Usaha",
            "debit_amount": 0.00,
            "credit_amount": 500000.00
          }
        ]
      }
    ]
  }
}
```

### 2.2 Integrated Summary for All Cash/Bank Accounts
```http
GET /api/cashbank/integrated-summary
```

**Response Structure:**
```json
{
  "status": "success",
  "data": {
    "summary": {
      "total_cash": 2500000.00,
      "total_bank": 15000000.00,
      "total_balance": 17500000.00,
      "total_ssot_balance": 17500000.00,
      "balance_variance": 0.00
    },
    "accounts": [
      {
        "id": 1,
        "name": "Bank Mandiri",
        "type": "BANK",
        "balance": 5000000.00,
        "ssot_balance": 5000000.00,
        "variance": 0.00,
        "last_transaction_date": "2025-09-20T10:00:00Z",
        "total_journal_entries": 25
      }
    ],
    "recent_activities": [
      {
        "type": "JOURNAL_ENTRY",
        "entry_number": "JE-2025-09-0001",
        "description": "Bank deposit",
        "amount": 500000.00,
        "account_name": "Bank Mandiri",
        "created_at": "2025-09-20T10:00:00Z"
      }
    ]
  }
}
```

### 2.3 Balance Reconciliation Endpoint
```http
GET /api/cashbank/accounts/:id/reconciliation
```

**Response Structure:**
```json
{
  "status": "success",
  "data": {
    "account_id": 1,
    "account_name": "Bank Mandiri",
    "cashbank_balance": 5000000.00,
    "ssot_balance": 5000000.00,
    "difference": 0.00,
    "reconciliation_status": "MATCHED",
    "last_reconciled_at": "2025-09-20T10:00:00Z",
    "discrepancies": [],
    "recommendations": []
  }
}
```

### 2.4 Journal Entries for Specific Cash/Bank Account
```http
GET /api/cashbank/accounts/:id/journal-entries
```

**Query Parameters:**
- `start_date` - Filter dari tanggal
- `end_date` - Filter sampai tanggal  
- `page` - Pagination
- `limit` - Items per page

**Response Structure:**
```json
{
  "status": "success",
  "data": {
    "journal_entries": [
      {
        "id": 456,
        "entry_number": "JE-2025-09-0001",
        "entry_date": "2025-09-20",
        "description": "Bank deposit from sales payment",
        "source_type": "CASH_BANK",
        "total_debit": 500000.00,
        "total_credit": 500000.00,
        "status": "POSTED",
        "lines": [
          {
            "account_id": 1,
            "account_code": "1101",
            "account_name": "Bank Mandiri",
            "debit_amount": 500000.00,
            "credit_amount": 0.00
          }
        ]
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 50,
      "total_pages": 3
    }
  }
}
```

## 3. IMPLEMENTASI BACKEND

### 3.1 New Service: CashBankIntegratedService

```go
type CashBankIntegratedService struct {
    db                   *gorm.DB
    cashBankService     *CashBankService
    unifiedJournalService *UnifiedJournalService
    accountRepo         repositories.AccountRepository
}

func (s *CashBankIntegratedService) GetIntegratedAccountDetails(accountID uint) (*IntegratedAccountResponse, error) {
    // 1. Get CashBank account details
    account, err := s.cashBankService.GetCashBankByID(accountID)
    if err != nil {
        return nil, err
    }
    
    // 2. Get related journal entries
    journalEntries, err := s.getJournalEntriesForAccount(account.AccountID)
    if err != nil {
        return nil, err
    }
    
    // 3. Calculate SSOT balance
    ssotBalance, err := s.calculateSSOTBalance(account.AccountID)
    if err != nil {
        return nil, err
    }
    
    // 4. Build integrated response
    return s.buildIntegratedResponse(account, journalEntries, ssotBalance), nil
}
```

### 3.2 New Controller: CashBankIntegratedController

```go
type CashBankIntegratedController struct {
    integratedService *CashBankIntegratedService
}

func (c *CashBankIntegratedController) GetIntegratedAccountDetails(ctx *gin.Context) {
    accountID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
        return
    }
    
    result, err := c.integratedService.GetIntegratedAccountDetails(uint(accountID))
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    ctx.JSON(http.StatusOK, gin.H{
        "status": "success",
        "data":   result,
    })
}
```

### 3.3 Route Registration
Tambahkan routes berikut di `routes/payment_routes.go`:

```go
// Integrated CashBank-SSOT routes
integrated := router.Group("/cashbank/integrated")
{
    integrated.GET("/accounts/:id", permissionMiddleware.CanView("cash_bank"), integratedController.GetIntegratedAccountDetails)
    integrated.GET("/summary", permissionMiddleware.CanView("cash_bank"), integratedController.GetIntegratedSummary)
    integrated.GET("/accounts/:id/reconciliation", permissionMiddleware.CanView("cash_bank"), integratedController.GetReconciliation)
    integrated.GET("/accounts/:id/journal-entries", permissionMiddleware.CanView("cash_bank"), integratedController.GetJournalEntries)
}
```

## 4. IMPLEMENTASI FRONTEND

### 4.1 API Service untuk Integrasi

```typescript
// services/cashBankIntegratedService.ts
export class CashBankIntegratedService {
  async getIntegratedAccountDetails(accountId: number): Promise<IntegratedAccountResponse> {
    const response = await fetch(`/api/cashbank/integrated/accounts/${accountId}`, {
      headers: {
        'Authorization': `Bearer ${this.getToken()}`
      }
    });
    return response.json();
  }
  
  async getIntegratedSummary(): Promise<IntegratedSummaryResponse> {
    const response = await fetch('/api/cashbank/integrated/summary', {
      headers: {
        'Authorization': `Bearer ${this.getToken()}`
      }
    });
    return response.json();
  }
}
```

### 4.2 Enhanced Cash-Bank Page Components

```typescript
// components/CashBank/IntegratedAccountCard.tsx
interface IntegratedAccountCardProps {
  account: IntegratedAccount;
  onViewJournalEntries: (accountId: number) => void;
}

export const IntegratedAccountCard: React.FC<IntegratedAccountCardProps> = ({ 
  account, 
  onViewJournalEntries 
}) => {
  const balanceVariance = account.balance - account.ssot_balance;
  const hasVariance = Math.abs(balanceVariance) > 0.01;
  
  return (
    <Card className="p-6">
      <div className="flex justify-between items-start">
        <div>
          <h3 className="text-lg font-semibold">{account.name}</h3>
          <p className="text-sm text-gray-500">{account.code} • {account.type}</p>
        </div>
        
        <div className="text-right">
          <p className="text-xl font-bold">
            {formatCurrency(account.balance)}
          </p>
          {hasVariance && (
            <p className="text-sm text-orange-600">
              Variance: {formatCurrency(balanceVariance)}
            </p>
          )}
        </div>
      </div>
      
      <div className="mt-4 flex gap-2">
        <Button 
          variant="outline" 
          size="sm"
          onClick={() => onViewJournalEntries(account.id)}
        >
          View Journal Entries
        </Button>
        
        {hasVariance && (
          <Button 
            variant="outline" 
            size="sm" 
            className="text-orange-600 border-orange-600"
          >
            Reconcile
          </Button>
        )}
      </div>
    </Card>
  );
};
```

## 5. DATA FLOW INTEGRASI

```
1. Frontend Request → GET /api/cashbank/integrated/accounts/1
2. CashBankIntegratedController.GetIntegratedAccountDetails()
3. CashBankIntegratedService koordinasi:
   a. CashBankService.GetCashBankByID(1) → CashBank data
   b. UnifiedJournalService.GetJournalEntriesForAccount(GL_ID) → Journal entries
   c. Calculate balance differences dan reconciliation status
4. Aggregate dan return integrated response
5. Frontend render dengan enhanced UI components
```

## 6. BENEFITS INTEGRASI

### 6.1 For Users
- ✅ **Single Dashboard**: Lihat cashbank balance + journal entries di satu tempat
- ✅ **Real-time Reconciliation**: Otomatis detect variance antara systems
- ✅ **Complete Audit Trail**: Track semua transaksi dari kedua perspectives
- ✅ **Enhanced Troubleshooting**: Mudah debug discrepancies

### 6.2 For System
- ✅ **Data Consistency**: Enforce consistency antara CashBank dan SSOT
- ✅ **Performance Optimization**: Aggregate queries untuk reduce API calls
- ✅ **Maintainability**: Clear separation of concerns dengan service layer
- ✅ **Extensibility**: Mudah tambah integrasi dengan modules lain

## 7. IMPLEMENTATION PHASES

### Phase 1: Core Integration (Week 1)
- [ ] Implement CashBankIntegratedService
- [ ] Create basic integrated endpoints
- [ ] Add route registrations
- [ ] Basic frontend integration

### Phase 2: Enhanced Features (Week 2)  
- [ ] Balance reconciliation logic
- [ ] Variance detection dan alerting
- [ ] Enhanced UI components
- [ ] Real-time updates via WebSocket

### Phase 3: Advanced Features (Week 3)
- [ ] Automated reconciliation
- [ ] Advanced reporting
- [ ] Performance optimizations
- [ ] Comprehensive testing

## 8. TESTING STRATEGY

### 8.1 Backend Testing
```go
func TestCashBankIntegration(t *testing.T) {
    // Test integrated service
    service := NewCashBankIntegratedService(db, cashBankService, journalService, accountRepo)
    
    // Test account details integration
    result, err := service.GetIntegratedAccountDetails(1)
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, result.Account.Balance, result.SSOTBalance)
}
```

### 8.2 Frontend Testing
```typescript
describe('CashBank Integration', () => {
  it('should load integrated account details', async () => {
    const service = new CashBankIntegratedService();
    const result = await service.getIntegratedAccountDetails(1);
    
    expect(result.status).toBe('success');
    expect(result.data.account).toBeDefined();
    expect(result.data.related_journal_entries).toBeDefined();
  });
});
```

## 9. MONITORING & ALERTS

### 9.1 Balance Variance Monitoring
- Real-time alerts jika balance difference > threshold
- Daily reconciliation reports
- Automated variance investigation

### 9.2 Performance Monitoring
- API response time tracking
- Database query performance
- Frontend rendering performance

---

**Status**: 📋 Ready for Implementation  
**Priority**: 🔥 High (Critical for system integration)  
**Estimated Effort**: 3-4 weeks  
**Dependencies**: SSOT Journal System, CashBank System