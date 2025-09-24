# SSOT Cash Flow Statement API Documentation

## Overview
The SSOT (Single Source of Truth) Cash Flow Statement API provides comprehensive cash flow reporting directly from the unified journal system. This ensures data consistency and real-time accuracy across all financial reports.

## Base URL
```
/api/v1/reports/ssot
```

## Authentication
All endpoints require JWT authentication and appropriate role permissions:
- `finance` - Can access all cash flow reports
- `admin` - Can access all cash flow reports
- `director` - Can access all cash flow reports

## Endpoints

### 1. Generate Cash Flow Statement
**GET** `/cash-flow`

Generates a comprehensive cash flow statement from SSOT journal entries.

#### Query Parameters
| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| start_date | string | Yes | Start date in YYYY-MM-DD format | 2024-01-01 |
| end_date | string | Yes | End date in YYYY-MM-DD format | 2024-12-31 |
| format | string | No | Output format (json, pdf, excel, csv) | json |

#### Example Request
```bash
GET /api/v1/reports/ssot/cash-flow?start_date=2024-01-01&end_date=2024-12-31&format=json
Authorization: Bearer <jwt_token>
```

#### Response Structure
```json
{
  "status": "success",
  "data": {
    "title": "Cash Flow Statement",
    "period": "2024-01-01 - 2024-12-31",
    "company": {
      "name": "PT. Sistem Akuntansi",
      "period": "01/01/2024 - 31/12/2024"
    },
    "sections": [
      {
        "name": "OPERATING ACTIVITIES",
        "total": 4440000.00,
        "items": [
          {
            "name": "Net Income",
            "amount": -8000000.00,
            "account_code": "",
            "type": "base"
          },
          {
            "name": "Adjustments for Non-Cash Items",
            "total": 16900000.00,
            "items": [
              {
                "code": "1201",
                "name": "Piutang Usaha",
                "amount": -2220000.00,
                "type": "increase"
              }
            ]
          },
          {
            "name": "Changes in Working Capital",
            "total": -4460000.00,
            "items": [
              {
                "code": "1301",
                "name": "Persediaan Barang Dagangan",
                "amount": 6000000.00,
                "type": "decrease"
              }
            ]
          }
        ],
        "summary": {
          "net_income": -8000000.00,
          "total_adjustments": 16900000.00,
          "total_working_capital_changes": -4460000.00
        }
      },
      {
        "name": "INVESTING ACTIVITIES",
        "total": 0.00,
        "items": [],
        "summary": {
          "purchase_of_fixed_assets": 0.00,
          "sale_of_fixed_assets": 0.00,
          "purchase_of_investments": 0.00,
          "sale_of_investments": 0.00,
          "intangible_asset_purchases": 0.00,
          "other_investing_activities": 0.00
        }
      },
      {
        "name": "FINANCING ACTIVITIES",
        "total": 10000000.00,
        "items": [
          {
            "name": "Modal Saham",
            "amount": 10000000.00,
            "account_code": "3101",
            "type": "inflow"
          }
        ],
        "summary": {
          "share_capital_increase": 10000000.00,
          "share_capital_decrease": 0.00,
          "long_term_debt_increase": 0.00,
          "long_term_debt_decrease": 0.00,
          "short_term_debt_increase": 0.00,
          "short_term_debt_decrease": 0.00,
          "dividends_paid": 0.00,
          "other_financing_activities": 0.00
        }
      },
      {
        "name": "NET CASH FLOW",
        "total": 14440000.00,
        "is_calculated": true,
        "items": [
          {
            "name": "Cash at Beginning of Period",
            "amount": 0.00
          },
          {
            "name": "Net Cash Flow from Activities",
            "amount": 14440000.00
          },
          {
            "name": "Cash at End of Period",
            "amount": 14440000.00
          }
        ]
      }
    ],
    "enhanced": true,
    "hasData": true,
    "summary": {
      "operating_cash_flow": 4440000.00,
      "investing_cash_flow": 0.00,
      "financing_cash_flow": 10000000.00,
      "net_cash_flow": 14440000.00,
      "cash_at_beginning": 0.00,
      "cash_at_end": 14440000.00
    },
    "cashFlowRatios": {
      "operating_cash_flow_ratio": 0.00,
      "cash_flow_to_debt_ratio": 0.00,
      "free_cash_flow": 0.00,
      "cash_flow_per_share": 0.00
    },
    "start_date": "2024-01-01",
    "end_date": "2024-12-31",
    "generated_at": "2025-09-22T04:36:31Z",
    "account_details": [
      {
        "account_id": 1,
        "account_code": "1101",
        "account_name": "Kas",
        "account_type": "ASSET",
        "debit_total": 14440000.00,
        "credit_total": 0.00,
        "net_balance": 14440000.00
      }
    ],
    "data_source": "SSOT Journal System",
    "message": "Positive cash generation from operations with overall positive net cash flow indicates healthy cash management."
  }
}
```

### 2. Get Cash Flow Summary
**GET** `/cash-flow/summary`

Returns a simplified summary view of the cash flow statement.

#### Query Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| start_date | string | Yes | Start date in YYYY-MM-DD format |
| end_date | string | Yes | End date in YYYY-MM-DD format |

#### Example Response
```json
{
  "status": "success",
  "message": "Cash Flow summary generated successfully",
  "data": {
    "company": {
      "name": "PT. Sistem Akuntansi"
    },
    "period": "2024-01-01 to 2024-12-31",
    "currency": "IDR",
    "activities": {
      "operating": {
        "cash_flow": 4440000.00,
        "net_income": -8000000.00
      },
      "investing": {
        "cash_flow": 0.00
      },
      "financing": {
        "cash_flow": 10000000.00
      }
    },
    "cash_position": {
      "beginning_cash": 0.00,
      "ending_cash": 14440000.00,
      "net_change": 14440000.00
    },
    "ratios": {
      "operating_cash_flow_ratio": 0.00,
      "free_cash_flow": 0.00
    },
    "generated_at": "2025-09-22T04:36:31Z",
    "enhanced": true
  }
}
```

### 3. Validate Cash Flow Statement
**GET** `/cash-flow/validate`

Validates if the cash flow statement balances correctly and provides reconciliation details.

#### Query Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| start_date | string | Yes | Start date in YYYY-MM-DD format |
| end_date | string | Yes | End date in YYYY-MM-DD format |

#### Example Response
```json
{
  "status": "success",
  "message": "Cash Flow validation completed",
  "data": {
    "period": "2024-01-01 to 2024-12-31",
    "is_balanced": true,
    "cash_at_beginning": 0.00,
    "net_cash_flow": 14440000.00,
    "expected_ending_cash": 14440000.00,
    "actual_ending_cash": 14440000.00,
    "difference": 0.00,
    "tolerance": 0.01,
    "validation_status": "PASS",
    "generated_at": "2025-09-22T04:36:31Z"
  }
}
```

## Cash Flow Activities Classification

### Operating Activities
- **Net Income**: Starting point from P&L statement
- **Non-Cash Adjustments**:
  - Depreciation and Amortization
  - Bad Debt Expense
  - Gain/Loss on Asset Disposal
  - Other Non-Cash Items
- **Working Capital Changes**:
  - Accounts Receivable Changes (1201, 1202)
  - Inventory Changes (1301, 130x)
  - Prepaid Expenses Changes (114x, 115x)
  - Accounts Payable Changes (2101)
  - Accrued Liabilities Changes (212x, 213x)

### Investing Activities
- Purchase/Sale of Fixed Assets (12xx, 16xx, 17xx)
- Purchase/Sale of Investments (15xx)
- Intangible Asset Purchases (14xx)
- Other Investing Activities

### Financing Activities
- Share Capital Changes (31xx)
- Long-term Debt Changes (22xx)
- Short-term Debt Changes (211x)
- Dividend Payments
- Other Financing Activities

## Account Code Mapping

The system uses Indonesian Chart of Accounts structure:

| Code Range | Category | Cash Flow Impact |
|------------|----------|------------------|
| 110x | Cash and Bank | Direct cash impact |
| 112x | Accounts Receivable | Operating (working capital) |
| 113x, 130x | Inventory | Operating (working capital) |
| 114x, 115x | Prepaid Expenses | Operating (working capital) |
| 12xx, 16xx, 17xx | Fixed Assets | Investing activities |
| 14xx | Intangible Assets | Investing activities |
| 15xx | Investments | Investing activities |
| 210x | Accounts Payable | Operating (working capital) |
| 211x | Short-term Debt | Financing activities |
| 212x, 213x | Accrued Liabilities | Operating (working capital) |
| 22xx | Long-term Debt | Financing activities |
| 31xx | Share Capital | Financing activities |

## Error Responses

### 400 Bad Request
```json
{
  "status": "error",
  "message": "start_date and end_date are required"
}
```

### 401 Unauthorized
```json
{
  "status": "error",
  "message": "Authentication required"
}
```

### 403 Forbidden
```json
{
  "status": "error",
  "message": "Insufficient permissions. Required roles: finance, admin, director"
}
```

### 500 Internal Server Error
```json
{
  "status": "error",
  "message": "Failed to generate SSOT Cash Flow statement",
  "error": "Database connection failed"
}
```

## Features

- **Real-time Data**: Directly sourced from SSOT journal entries
- **Automatic Categorization**: Accounts automatically classified into appropriate cash flow activities
- **Balance Reconciliation**: Automatic validation of cash flow balance
- **Multiple Formats**: Support for JSON, PDF, Excel, and CSV export
- **Detailed Analysis**: Includes financial ratios and performance metrics
- **Audit Trail**: Complete account details for drilldown analysis
- **Frontend Compatible**: Response format optimized for React components

## Integration Notes

1. **Frontend Integration**: The API response is structured to work seamlessly with React components similar to P&L and Balance Sheet modals.

2. **Data Consistency**: All data comes from the unified journal system, ensuring consistency with P&L and Balance Sheet reports.

3. **Performance**: Optimized queries with proper indexing for fast report generation.

4. **Security**: All endpoints require authentication and proper role-based authorization.

5. **Extensibility**: Easy to add new cash flow categories and account mappings as business needs evolve.

## Testing

Use the test script to verify functionality:
```bash
go run cmd/scripts/test_ssot_cash_flow.go
```

The test covers:
- Basic cash flow generation
- Activities analysis
- Balance reconciliation  
- Ratios calculation
- Data structure validation
- Error handling
- JSON serialization