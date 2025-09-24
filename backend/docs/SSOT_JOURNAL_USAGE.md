# Single Source of Truth (SSOT) Journal System - Usage Guide

## Overview

The SSOT Journal System provides a unified approach to managing journal entries in the accounting application. This system consolidates all journal operations into a single, consistent interface while maintaining data integrity and performance.

## Features

- ✅ **Unified Journal Entry Management**: Single interface for all journal operations
- ✅ **Automatic Balance Validation**: Built-in validation ensures all entries are balanced
- ✅ **Real-time Account Balances**: Materialized view for optimized balance queries
- ✅ **Event Logging**: Complete audit trail of all journal operations
- ✅ **Transaction Factory Pattern**: Consistent handling across different source types
- ✅ **Auto-generated Entry Numbers**: Sequential numbering with customizable formats
- ✅ **Reversing Entries**: Support for journal reversals
- ✅ **Status Management**: Draft → Posted workflow
- ✅ **Comprehensive API**: REST endpoints for all operations

## Database Schema

### Core Tables

1. **`unified_journal_ledger`**: Main journal entries
2. **`unified_journal_lines`**: Individual debit/credit lines
3. **`journal_event_log`**: Complete audit trail
4. **`account_balances`**: Materialized view for real-time balances

## Usage Examples

### 1. Running the Migration Test

```bash
# Test the SSOT migration
cd backend
go run cmd/scripts/test_ssot_migration.go
```

### 2. Creating a Journal Entry

```go
package main

import (
    "time"
    "app-sistem-akuntansi/services"
    "github.com/shopspring/decimal"
)

func createJournalExample(service *services.UnifiedJournalService) {
    request := &services.JournalEntryRequest{
        SourceType:  "manual",
        Reference:   "JV-2024-001",
        EntryDate:   time.Now(),
        Description: "Sample Journal Entry",
        Lines: []services.JournalLineRequest{
            {
                AccountID:    1, // Cash
                Description:  "Cash receipt",
                DebitAmount:  decimal.NewFromFloat(1000.00),
                CreditAmount: decimal.Zero,
            },
            {
                AccountID:    2, // Sales
                Description:  "Sales revenue",
                DebitAmount:  decimal.Zero,
                CreditAmount: decimal.NewFromFloat(1000.00),
            },
        },
        AutoPost:  true,
        CreatedBy: 1,
    }
    
    response, err := service.CreateJournalEntry(request)
    if err != nil {
        log.Printf("Error creating journal entry: %v", err)
        return
    }
    
    log.Printf("Created journal entry: %s", response.EntryNumber)
}
```

### 3. REST API Examples

#### Create Journal Entry
```bash
curl -X POST http://localhost:8080/api/v1/journals \
  -H "Content-Type: application/json" \
  -d '{
    "source_type": "manual",
    "reference": "JV-2024-001",
    "entry_date": "2024-01-15T00:00:00Z",
    "description": "Sample journal entry",
    "lines": [
      {
        "account_id": 1,
        "description": "Cash receipt",
        "debit_amount": "1000.00",
        "credit_amount": "0"
      },
      {
        "account_id": 2,
        "description": "Sales revenue",
        "debit_amount": "0",
        "credit_amount": "1000.00"
      }
    ],
    "auto_post": true,
    "created_by": 1
  }'
```

#### Get Journal Entries with Filters
```bash
# Get all journal entries
curl http://localhost:8080/api/v1/journals

# Get journal entries with filters
curl "http://localhost:8080/api/v1/journals?status=posted&date_from=2024-01-01&limit=10"

# Get specific journal entry
curl http://localhost:8080/api/v1/journals/1
```

#### Post a Draft Entry
```bash
curl -X PUT http://localhost:8080/api/v1/journals/1/post
```

#### Reverse a Posted Entry
```bash
curl -X POST http://localhost:8080/api/v1/journals/1/reverse \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Reversing incorrect entry"
  }'
```

#### Get Account Balances
```bash
# Get current account balances
curl http://localhost:8080/api/v1/journals/account-balances

# Refresh materialized view
curl -X POST http://localhost:8080/api/v1/journals/account-balances/refresh
```

## Integration Guide

### 1. Initialize the Service

```go
// In your main application setup
func setupJournalService(db *gorm.DB) *services.UnifiedJournalService {
    return services.NewUnifiedJournalService(db)
}

// Register the controller
func setupRoutes(router *gin.Engine, db *gorm.DB) {
    journalService := setupJournalService(db)
    journalController := controllers.NewUnifiedJournalController(journalService)
    journalController.RegisterRoutes(router)
}
```

### 2. Using in Existing Services

```go
// Example: Sales invoice service creating journal entries
type SalesService struct {
    db             *gorm.DB
    journalService *services.UnifiedJournalService
}

func (s *SalesService) CreateInvoice(invoice *Invoice) error {
    // Create invoice record
    if err := s.db.Create(invoice).Error; err != nil {
        return err
    }
    
    // Create corresponding journal entry
    journalReq := &services.JournalEntryRequest{
        SourceType:  "sales_invoice",
        SourceID:    &invoice.ID,
        Reference:   invoice.Number,
        EntryDate:   invoice.Date,
        Description: fmt.Sprintf("Sales Invoice %s", invoice.Number),
        Lines: []services.JournalLineRequest{
            {
                AccountID:    s.getAccountsReceivableID(),
                Description:  "Accounts Receivable",
                DebitAmount:  invoice.Total,
                CreditAmount: decimal.Zero,
            },
            {
                AccountID:    s.getSalesAccountID(),
                Description:  "Sales Revenue",
                DebitAmount:  decimal.Zero,
                CreditAmount: invoice.Total,
            },
        },
        AutoPost:  true,
        CreatedBy: invoice.CreatedBy,
    }
    
    _, err := s.journalService.CreateJournalEntry(journalReq)
    return err
}
```

## API Endpoints Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/journals` | Create new journal entry |
| GET | `/api/v1/journals` | List journal entries (with filters) |
| GET | `/api/v1/journals/{id}` | Get specific journal entry |
| PUT | `/api/v1/journals/{id}/post` | Post a draft entry |
| POST | `/api/v1/journals/{id}/reverse` | Create reversing entry |
| GET | `/api/v1/journals/account-balances` | Get account balances |
| POST | `/api/v1/journals/account-balances/refresh` | Refresh balance view |
| GET | `/api/v1/journals/summary` | Get journal statistics |

## Request/Response Formats

### Journal Entry Request
```json
{
  "source_type": "manual|sales_invoice|purchase_invoice|payment|receipt|adjustment|reversal",
  "source_id": 123,
  "reference": "JV-2024-001",
  "entry_date": "2024-01-15T00:00:00Z",
  "description": "Journal entry description",
  "lines": [
    {
      "account_id": 1,
      "description": "Line description",
      "debit_amount": "100.00",
      "credit_amount": "0"
    }
  ],
  "auto_post": true,
  "created_by": 1
}
```

### Journal Entry Response
```json
{
  "id": 1,
  "entry_number": "JE-2024-000001",
  "status": "posted",
  "total_debit": "100.00",
  "total_credit": "100.00",
  "is_balanced": true,
  "lines": [
    {
      "id": 1,
      "line_number": 1,
      "account_id": 1,
      "description": "Line description",
      "debit_amount": "100.00",
      "credit_amount": "0"
    }
  ],
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

## Validation Rules

1. **Balance Requirement**: Total debits must equal total credits
2. **Minimum Lines**: At least 2 journal lines required
3. **Amount Validation**: 
   - Each line must have either debit OR credit amount (not both)
   - Amounts must be positive
   - Zero amounts not allowed
4. **Account Validation**: All account IDs must be valid
5. **Status Workflow**: Only draft entries can be posted
6. **Reversal Rules**: Only posted entries can be reversed

## Performance Considerations

1. **Materialized View**: Account balances are updated via materialized view
2. **Indexing**: Optimized indexes for common query patterns
3. **Partitioning**: Ready for table partitioning by date ranges
4. **Batch Operations**: Support for bulk journal entry creation

## Monitoring and Maintenance

### Refresh Account Balances
```sql
REFRESH MATERIALIZED VIEW account_balances;
```

### Check System Health
```sql
-- Check for unbalanced entries
SELECT * FROM unified_journal_ledger WHERE NOT is_balanced;

-- Check event log
SELECT * FROM journal_event_log ORDER BY created_at DESC LIMIT 10;
```

## Migration from Legacy System

The SSOT system includes migration scripts to move data from the existing fragmented journal tables. See `backend/migrations/020_create_unified_journal_ssot.sql` for the complete migration.

## Troubleshooting

### Common Issues

1. **Unbalanced Entries**: Check validation logic in your client code
2. **Missing Entry Numbers**: Ensure triggers are installed correctly
3. **Performance**: Refresh materialized view if balance queries are slow
4. **Permissions**: Check database user permissions for materialized views

### Test Migration
```bash
go run backend/cmd/scripts/test_ssot_migration.go
```

This comprehensive guide should help you understand and use the SSOT Journal System effectively in your accounting application.