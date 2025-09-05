# Unified Financial Reporting System

## Overview

The Unified Financial Reporting System is a comprehensive solution that consolidates all financial reporting features into a single, cohesive service architecture. This system provides a unified interface for generating various financial reports, dashboards, and analytical data while maintaining high performance and consistency across all reports.

## Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────────┐
│                     Unified Financial Reporting System          │
├─────────────────────────────────────────────────────────────────┤
│  Controllers Layer                                              │
│  ├── UnifiedFinancialReportController                          │
│  └── Legacy Controllers (for backward compatibility)           │
├─────────────────────────────────────────────────────────────────┤
│  Services Layer                                                 │
│  └── UnifiedFinancialReportService                             │
├─────────────────────────────────────────────────────────────────┤
│  Repositories Layer                                             │
│  ├── AccountRepository                                          │
│  ├── JournalEntryRepository                                     │
│  ├── SaleRepository                                             │
│  ├── PurchaseRepository                                         │
│  ├── CashBankRepository                                         │
│  └── PaymentRepository                                          │
├─────────────────────────────────────────────────────────────────┤
│  Database Layer                                                 │
│  └── GORM with PostgreSQL/MySQL                                │
└─────────────────────────────────────────────────────────────────┘
```

### Key Features

- **Comprehensive Financial Statements**: Profit & Loss, Balance Sheet, Cash Flow
- **Accounting Reports**: Trial Balance, General Ledger
- **Operational Analytics**: Sales Summary, Vendor Analysis
- **Real-time Dashboard**: Key financial metrics and ratios
- **Comparative Analysis**: Multi-period comparison capabilities
- **Batch Report Generation**: Generate all reports simultaneously
- **Parameter Validation**: Built-in validation for all parameters
- **Legacy Compatibility**: Backward compatibility with existing endpoints

## Services

### UnifiedFinancialReportService

The core service that handles all financial report generation with the following methods:

#### Financial Statements
- `GenerateComprehensiveProfitLoss()` - Comprehensive P&L with COGS separation
- `GenerateComprehensiveBalanceSheet()` - Balance sheet with category grouping
- `GenerateComprehensiveCashFlow()` - Cash flow with detailed activity analysis

#### Accounting Reports
- `GenerateComprehensiveTrialBalance()` - Trial balance with optional zero balances
- `GenerateComprehensiveGeneralLedger()` - Detailed account ledger with running balances

#### Operational Reports
- `GenerateComprehensiveSalesSummary()` - Sales analytics with customer/product insights
- `GenerateComprehensiveVendorAnalysis()` - Vendor performance and payment analysis

## API Endpoints

### Base URL: `/api/unified-reports`

#### Financial Statements

##### Profit & Loss Statement
```
GET /profit-loss
Parameters:
  - start_date (required): YYYY-MM-DD
  - end_date (required): YYYY-MM-DD  
  - comparative (optional): boolean, default false
  
Example: /api/unified-reports/profit-loss?start_date=2024-01-01&end_date=2024-12-31&comparative=true
```

##### Balance Sheet
```
GET /balance-sheet
Parameters:
  - as_of_date (optional): YYYY-MM-DD, default today
  - comparative (optional): boolean, default false
  
Example: /api/unified-reports/balance-sheet?as_of_date=2024-12-31&comparative=false
```

##### Cash Flow Statement
```
GET /cash-flow
Parameters:
  - start_date (required): YYYY-MM-DD
  - end_date (required): YYYY-MM-DD
  
Example: /api/unified-reports/cash-flow?start_date=2024-01-01&end_date=2024-12-31
```

#### Accounting Reports

##### Trial Balance
```
GET /trial-balance
Parameters:
  - as_of_date (optional): YYYY-MM-DD, default today
  - show_zero (optional): boolean, default false
  
Example: /api/unified-reports/trial-balance?as_of_date=2024-12-31&show_zero=true
```

##### General Ledger
```
GET /general-ledger/:account_id
Parameters:
  - account_id (required): uint, path parameter
  - start_date (required): YYYY-MM-DD
  - end_date (required): YYYY-MM-DD
  
Example: /api/unified-reports/general-ledger/1?start_date=2024-01-01&end_date=2024-12-31
```

#### Operational Reports

##### Sales Summary
```
GET /sales-summary
Parameters:
  - start_date (required): YYYY-MM-DD
  - end_date (required): YYYY-MM-DD
  
Example: /api/unified-reports/sales-summary?start_date=2024-01-01&end_date=2024-12-31
```

##### Vendor Analysis
```
GET /vendor-analysis
Parameters:
  - start_date (required): YYYY-MM-DD
  - end_date (required): YYYY-MM-DD
  
Example: /api/unified-reports/vendor-analysis?start_date=2024-01-01&end_date=2024-12-31
```

#### Dashboard & Utilities

##### Financial Dashboard
```
GET /dashboard
Parameters:
  - start_date (optional): YYYY-MM-DD, default first day of current month
  - end_date (optional): YYYY-MM-DD, default today
  
Example: /api/unified-reports/dashboard?start_date=2024-01-01&end_date=2024-12-31
```

##### Available Reports Metadata
```
GET /available
Example: /api/unified-reports/available
```

##### Batch Report Generation
```
GET /all
Parameters:
  - start_date (required): YYYY-MM-DD
  - end_date (required): YYYY-MM-DD
  - as_of_date (optional): YYYY-MM-DD, default today
  
Example: /api/unified-reports/all?start_date=2024-01-01&end_date=2024-12-31
```

##### Parameter Validation
```
GET /validate
Parameters:
  - start_date (optional): YYYY-MM-DD
  - end_date (optional): YYYY-MM-DD
  - as_of_date (optional): YYYY-MM-DD
  
Example: /api/unified-reports/validate?start_date=2024-01-01&end_date=2024-12-31
```

##### System Documentation
```
GET /docs
Example: /api/unified-reports/docs
```

##### Health Check
```
GET /health
Example: /api/unified-reports/health
```

## Response Format

All endpoints return a standardized JSON response:

```json
{
  "status": "success|error|partial_success",
  "message": "Description of the response",
  "data": { /* Report data */ },
  "errors": [ /* Array of errors (if any) */ ]
}
```

## Report Data Structures

### Profit & Loss Statement
```json
{
  "header": {
    "company_name": "Company Name",
    "report_name": "Profit & Loss Statement",
    "period": "2024-01-01 to 2024-12-31",
    "generated_at": "2024-01-15T10:00:00Z"
  },
  "revenue": {
    "accounts": [...],
    "total": 100000.00
  },
  "cost_of_goods_sold": {
    "accounts": [...],
    "total": 60000.00
  },
  "gross_profit": 40000.00,
  "operating_expenses": {
    "accounts": [...],
    "total": 25000.00
  },
  "operating_income": 15000.00,
  "other_income": {
    "accounts": [...],
    "total": 2000.00
  },
  "other_expenses": {
    "accounts": [...],
    "total": 1000.00
  },
  "net_income": 16000.00,
  "comparative_data": { /* if comparative=true */ }
}
```

### Balance Sheet
```json
{
  "header": {
    "company_name": "Company Name",
    "report_name": "Balance Sheet",
    "as_of_date": "2024-12-31",
    "generated_at": "2024-01-15T10:00:00Z"
  },
  "assets": {
    "categories": [
      {
        "name": "Current Assets",
        "accounts": [...],
        "total": 50000.00
      }
    ],
    "total": 150000.00
  },
  "liabilities": {
    "categories": [...],
    "total": 80000.00
  },
  "equity": {
    "categories": [...],
    "total": 70000.00
  },
  "total_assets": 150000.00,
  "total_liabilities": 80000.00,
  "total_equity": 70000.00,
  "is_balanced": true
}
```

### Dashboard
```json
{
  "period": {
    "start_date": "2024-01-01",
    "end_date": "2024-12-31"
  },
  "financial_overview": {
    "total_revenue": 100000.00,
    "total_expenses": 84000.00,
    "gross_profit": 40000.00,
    "net_income": 16000.00,
    "total_assets": 150000.00,
    "total_liabilities": 80000.00,
    "total_equity": 70000.00,
    "cash_position": 25000.00,
    "is_balanced": true
  },
  "profitability_metrics": {
    "gross_profit_margin": 40.0,
    "net_profit_margin": 16.0,
    "return_on_assets": 10.67,
    "return_on_equity": 22.86
  },
  "liquidity_ratios": {
    "current_ratio": 2.5,
    "quick_ratio": 2.0,
    "cash_ratio": 1.25,
    "working_capital": 30000.00
  },
  "leverage_ratios": {
    "debt_to_assets": 53.33,
    "debt_to_equity": 1.14,
    "equity_multiplier": 2.14
  }
}
```

## Integration Guide

### Quick Setup

```go
package main

import (
    "app-sistem-akuntansi/integration"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func main() {
    // Initialize database and router
    db := setupDatabase()
    router := gin.Default()
    
    // Quick setup of unified financial reporting
    integration, err := integration.SetupUnifiedFinancialReporting(db, router)
    if err != nil {
        log.Fatalf("Failed to setup unified financial reporting: %v", err)
    }
    
    // Get system summary
    summary := integration.GetIntegrationSummary()
    log.Printf("System Summary: %+v", summary)
    
    // Perform health check
    healthy, details := integration.PerformHealthCheck()
    log.Printf("System Health: %t, Details: %+v", healthy, details)
    
    // Start server
    router.Run(":8080")
    
    // Graceful shutdown
    defer integration.Shutdown()
}
```

### Manual Setup

```go
func manualSetup() {
    // Initialize database and router
    db := setupDatabase()
    router := gin.Default()
    
    // Create integration instance
    integration := integration.NewUnifiedFinancialIntegration(db, router)
    
    // Initialize the system
    if err := integration.Initialize(); err != nil {
        log.Fatalf("Failed to initialize: %v", err)
    }
    
    // Add custom middleware or routes if needed
    // integration.Router.Use(customMiddleware())
    // integration.Router.GET("/custom-endpoint", customHandler)
    
    router.Run(":8080")
}
```

## Usage Examples

### Generate Profit & Loss Statement

```bash
curl -X GET "http://localhost:8080/api/unified-reports/profit-loss?start_date=2024-01-01&end_date=2024-12-31&comparative=true"
```

### Generate Financial Dashboard

```bash
curl -X GET "http://localhost:8080/api/unified-reports/dashboard?start_date=2024-01-01&end_date=2024-12-31"
```

### Generate All Reports (Batch)

```bash
curl -X GET "http://localhost:8080/api/unified-reports/all?start_date=2024-01-01&end_date=2024-12-31"
```

### Check System Health

```bash
curl -X GET "http://localhost:8080/api/unified-reports/health"
```

## Performance Considerations

### Optimization Features

1. **Concurrent Report Generation**: Batch report generation uses goroutines
2. **Database Connection Pooling**: Efficient database connection management
3. **Caching Strategy**: Results can be cached for frequently requested periods
4. **Selective Data Loading**: Only loads necessary data for each report type
5. **Prepared Statements**: Uses prepared statements for better performance

### Recommended Configurations

```go
// Database configuration for optimal performance
db.Config.ConnMaxLifetime = time.Hour
db.Config.MaxIdleConns = 10
db.Config.MaxOpenConns = 100
```

## Error Handling

The system provides comprehensive error handling:

### Common Error Types
- **Validation Errors**: Invalid date formats, missing parameters
- **Database Errors**: Connection issues, query failures
- **Business Logic Errors**: Unbalanced accounts, missing data
- **System Errors**: Memory issues, timeout errors

### Error Response Format
```json
{
  "status": "error",
  "message": "Error description",
  "error": "Detailed error message",
  "code": "ERROR_CODE"
}
```

## Security Considerations

### Access Control
- API endpoints should be protected with authentication
- Role-based access control for different report types
- Rate limiting to prevent abuse

### Data Security
- Sensitive financial data encryption
- Audit logging for all report generations
- Input validation and sanitization

## Monitoring and Logging

### Built-in Monitoring
- Health check endpoints
- Performance metrics
- Error tracking
- Request logging

### Log Formats
```
2024-01-15 10:00:00 - [GET] "/api/unified-reports/profit-loss" 200 1.234s "User-Agent"
```

## Testing

### Unit Tests
Run unit tests for individual components:
```bash
go test ./services/unified_financial_report_service_test.go
go test ./controllers/unified_financial_report_controller_test.go
```

### Integration Tests
Run integration tests for the complete system:
```bash
go test ./integration/unified_financial_integration_test.go
```

### API Tests
Test API endpoints:
```bash
# Test all endpoints
curl -X GET "http://localhost:8080/api/unified-reports/available"
```

## Troubleshooting

### Common Issues

1. **Database Connection Errors**
   - Check database configuration
   - Verify connection strings
   - Test database connectivity

2. **Report Generation Failures**
   - Validate date parameters
   - Check data availability
   - Review error logs

3. **Performance Issues**
   - Monitor database query performance
   - Check memory usage
   - Review concurrent request handling

### Debug Mode

Enable debug logging:
```go
gin.SetMode(gin.DebugMode)
```

## Migration Guide

### From Legacy Reports

1. **Update Import Statements**
   ```go
   // Old
   import "app-sistem-akuntansi/controllers"
   
   // New
   import "app-sistem-akuntansi/integration"
   ```

2. **Update Route Initialization**
   ```go
   // Old
   controllers.SetupReportRoutes(router, db)
   
   // New
   integration.SetupUnifiedFinancialReporting(db, router)
   ```

3. **Update API Calls**
   ```bash
   # Old
   GET /api/reports/profit-loss
   
   # New (recommended)
   GET /api/unified-reports/profit-loss
   
   # Legacy compatibility still available
   GET /api/reports/profit-loss
   ```

## Contributing

### Development Setup

1. Clone the repository
2. Install dependencies: `go mod tidy`
3. Set up database
4. Run tests: `go test ./...`
5. Start development server: `go run main.go`

### Code Style

Follow standard Go conventions:
- Use `gofmt` for formatting
- Follow naming conventions
- Add comprehensive comments
- Write unit tests for new features

### Submitting Changes

1. Create feature branch
2. Implement changes with tests
3. Update documentation
4. Submit pull request

## Support

For support and questions:
- Check the documentation
- Review error logs
- Contact development team
- File issues in the repository

## License

This system is part of the accounting application and follows the same license terms.

---

## Version History

### v1.0.0 (Current)
- Initial release of unified financial reporting system
- All major financial reports implemented
- Dashboard and analytics features
- Legacy compatibility maintained
- Comprehensive documentation

### Future Releases
- Performance optimizations
- Additional report types
- Enhanced dashboard features
- Export capabilities (PDF, Excel)
- Advanced filtering options
