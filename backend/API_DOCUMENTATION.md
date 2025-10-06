# 🚀 API Documentation - Sistema Akuntansi

Dokumentasi lengkap semua API endpoints dengan methods, parameters, dan response formats.

## 📋 Table of Contents
- [Base Configuration](#-base-configuration)
- [Authentication](#-authentication)
- [Master Data](#-master-data)
- [Transactions](#-transactions)
- [Reports & Analytics](#-reports--analytics)
- [Administration](#-administration)
- [WebSocket](#-websocket)
- [Error Handling](#-error-handling)
- [Rate Limiting](#-rate-limiting)

## 🔧 Base Configuration

### Base URLs
```
Development: http://localhost:8080/api/v1
Production:  https://your-domain.com/api/v1
```

### Headers
```http
Content-Type: application/json
Accept: application/json
Authorization: Bearer <jwt_token>
```

### Standard Response Format
```json
{
  "success": true,
  "data": {},
  "metadata": {
    "total": 100,
    "page": 1,
    "limit": 20,
    "has_more": true
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Standard Error Format
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": {}
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## 🔐 Authentication

### POST /auth/login
Login user dan generate JWT tokens.

**Request Body:**
```json
{
  "username": "admin",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "username": "admin",
      "role": "ADMIN",
      "permissions": ["view_all", "create_all", "edit_all"]
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "expires_in": 5400
    }
  }
}
```

**Error Responses:**
- `401` - Invalid credentials
- `429` - Too many login attempts
- `500` - Server error

### POST /auth/refresh
Refresh access token menggunakan refresh token.

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 5400
  }
}
```

### GET /auth/validate-token
Validate current access token.

**Headers:** `Authorization: Bearer <token>`

**Response:**
```json
{
  "success": true,
  "data": {
    "valid": true,
    "user_id": 1,
    "expires_in": 3600,
    "permissions": ["view_sales", "create_sales"]
  }
}
```

### GET /profile
Get current user profile.

**Headers:** `Authorization: Bearer <token>`

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "admin",
    "role": "ADMIN",
    "email": "admin@company.com",
    "last_login": "2024-01-15T10:30:00Z",
    "permissions": {
      "sales": ["view", "create", "edit", "delete"],
      "reports": ["view", "export"]
    }
  }
}
```

## 📊 Master Data

### Chart of Accounts

#### GET /accounts
Get list of accounts dengan pagination dan filtering.

**Query Parameters:**
- `page` (int, default: 1) - Page number
- `limit` (int, default: 20) - Items per page  
- `search` (string) - Search by code atau name
- `account_type` (string) - Filter by account type
- `is_header` (boolean) - Filter header accounts
- `is_active` (boolean, default: true) - Filter active accounts

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "code": "1001",
      "name": "Cash in Hand",
      "account_type": "ASSET",
      "parent_id": null,
      "parent_code": null,
      "is_header": false,
      "is_active": true,
      "balance": 1500000.00,
      "level": 0,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "metadata": {
    "total": 150,
    "page": 1,
    "limit": 20,
    "total_pages": 8
  }
}
```

#### POST /accounts
Create new account.

**Request Body:**
```json
{
  "code": "1002",
  "name": "Cash in Bank - BCA",
  "account_type": "ASSET",
  "parent_code": "1000",
  "description": "Bank Central Asia checking account"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "code": "1002",
    "name": "Cash in Bank - BCA",
    "account_type": "ASSET",
    "parent_id": 1,
    "balance": 0.00,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Error Responses:**
- `400` - Invalid account data
- `409` - Account code already exists
- `403` - Insufficient permissions

#### GET /accounts/:code
Get specific account by code.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "code": "1001",
    "name": "Cash in Hand",
    "account_type": "ASSET",
    "balance": 1500000.00,
    "parent": {
      "code": "1000",
      "name": "Current Assets"
    },
    "children": [
      {
        "code": "1001.01",
        "name": "Petty Cash"
      }
    ]
  }
}
```

#### PUT /accounts/:code
Update existing account.

**Request Body:**
```json
{
  "name": "Updated Account Name",
  "description": "Updated description",
  "is_active": true
}
```

#### DELETE /accounts/:code
Delete account (soft delete).

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Account deleted successfully"
  }
}
```

#### GET /accounts/hierarchy
Get accounts dalam hierarchical structure.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "code": "1000",
      "name": "ASSETS",
      "is_header": true,
      "children": [
        {
          "code": "1001",
          "name": "Cash in Hand",
          "balance": 1500000.00,
          "children": []
        }
      ]
    }
  ]
}
```

### Contacts Management

#### GET /contacts
Get list of contacts dengan filtering.

**Query Parameters:**
- `page`, `limit` - Pagination
- `search` - Search by name atau email
- `type` - Filter: `CUSTOMER`, `VENDOR`, `EMPLOYEE`
- `status` - Filter: `ACTIVE`, `INACTIVE`

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "PT. Customer ABC",
      "type": "CUSTOMER",
      "email": "contact@customerabc.com",
      "phone": "+62-21-1234567",
      "status": "ACTIVE",
      "credit_limit": 5000000.00,
      "outstanding_balance": 1200000.00,
      "addresses": [
        {
          "id": 1,
          "type": "BILLING",
          "address": "Jl. Sudirman No. 123",
          "city": "Jakarta",
          "postal_code": "12190"
        }
      ]
    }
  ]
}
```

#### POST /contacts
Create new contact.

**Request Body:**
```json
{
  "name": "PT. New Customer",
  "type": "CUSTOMER",
  "email": "info@newcustomer.com",
  "phone": "+62-21-9876543",
  "tax_number": "01.234.567.8-901.000",
  "credit_limit": 3000000.00,
  "payment_terms": 30,
  "addresses": [
    {
      "type": "BILLING",
      "address": "Jl. Thamrin No. 456",
      "city": "Jakarta",
      "state": "DKI Jakarta",
      "postal_code": "10350",
      "country": "Indonesia"
    }
  ]
}
```

#### GET /contacts/type/:type
Get contacts by type (CUSTOMER, VENDOR, EMPLOYEE).

#### POST /contacts/:id/addresses
Add address to existing contact.

### Products & Inventory

#### GET /products
Get products dengan inventory information.

**Query Parameters:**
- `search` - Search by name atau SKU
- `category_id` - Filter by category
- `low_stock` (boolean) - Show only low stock items
- `include_inactive` (boolean) - Include inactive products

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "sku": "PRD001",
      "name": "Product Name",
      "description": "Product description",
      "category": {
        "id": 1,
        "name": "Category Name"
      },
      "unit": {
        "id": 1,
        "name": "Piece",
        "symbol": "pcs"
      },
      "price": 150000.00,
      "cost": 100000.00,
      "stock": {
        "current_stock": 50,
        "reserved_stock": 5,
        "available_stock": 45,
        "min_stock": 10,
        "max_stock": 100
      },
      "is_active": true
    }
  ]
}
```

#### POST /products
Create new product.

**Request Body:**
```json
{
  "sku": "PRD002",
  "name": "New Product",
  "description": "Product description",
  "category_id": 1,
  "unit_id": 1,
  "price": 200000.00,
  "cost": 150000.00,
  "min_stock": 5,
  "max_stock": 50,
  "is_active": true
}
```

#### POST /products/adjust-stock
Adjust product stock.

**Request Body:**
```json
{
  "product_id": 1,
  "adjustment_type": "INCREASE", // INCREASE, DECREASE
  "quantity": 10,
  "reason": "Stock replenishment",
  "reference": "PO001"
}
```

## 💰 Transactions

### Sales Management

#### GET /sales
Get sales list dengan filtering.

**Query Parameters:**
- `page`, `limit` - Pagination
- `start_date`, `end_date` - Date range filter
- `customer_id` - Filter by customer
- `status` - Filter: `DRAFT`, `CONFIRMED`, `INVOICED`, `PAID`
- `search` - Search by sales number atau customer name

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "sales_number": "SO-2024-001",
      "date": "2024-01-15",
      "customer": {
        "id": 1,
        "name": "PT. Customer ABC"
      },
      "status": "CONFIRMED",
      "subtotal": 2500000.00,
      "tax_amount": 250000.00,
      "total": 2750000.00,
      "paid_amount": 0.00,
      "remaining_amount": 2750000.00,
      "items_count": 3,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

#### POST /sales
Create new sales order.

**Request Body:**
```json
{
  "customer_id": 1,
  "date": "2024-01-15",
  "due_date": "2024-02-14",
  "reference": "Quotation Q001",
  "notes": "Sales order notes",
  "items": [
    {
      "product_id": 1,
      "quantity": 5,
      "unit_price": 500000.00,
      "discount": 0.00,
      "description": "Product description"
    },
    {
      "product_id": 2,
      "quantity": 2,
      "unit_price": 750000.00,
      "discount": 50000.00
    }
  ]
}
```

#### GET /sales/:id
Get detailed sales order.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "sales_number": "SO-2024-001",
    "date": "2024-01-15",
    "due_date": "2024-02-14",
    "status": "CONFIRMED",
    "customer": {
      "id": 1,
      "name": "PT. Customer ABC",
      "email": "contact@customerabc.com"
    },
    "subtotal": 2500000.00,
    "discount_total": 50000.00,
    "tax_amount": 250000.00,
    "total": 2750000.00,
    "items": [
      {
        "id": 1,
        "product": {
          "id": 1,
          "sku": "PRD001",
          "name": "Product Name"
        },
        "quantity": 5,
        "unit_price": 500000.00,
        "discount": 0.00,
        "total": 2500000.00
      }
    ],
    "payments": [
      {
        "id": 1,
        "amount": 1000000.00,
        "date": "2024-01-16",
        "method": "BANK_TRANSFER",
        "reference": "TRF123456"
      }
    ],
    "journal_entries": [
      {
        "id": 1,
        "journal_code": "JE-2024-001",
        "status": "POSTED"
      }
    ]
  }
}
```

#### POST /sales/:id/confirm
Confirm sales order (allocate stock).

#### POST /sales/:id/invoice
Generate invoice dari sales order.

#### POST /sales/:id/payments
Record payment untuk sales.

**Request Body:**
```json
{
  "amount": 1000000.00,
  "payment_date": "2024-01-16",
  "payment_method": "BANK_TRANSFER",
  "account_id": 2,
  "reference": "TRF123456",
  "notes": "Partial payment"
}
```

### Purchase Management

#### GET /purchases
Get purchase orders.

**Query Parameters:**
Similar to sales dengan additional:
- `vendor_id` - Filter by vendor
- `approval_status` - Filter: `PENDING`, `APPROVED`, `REJECTED`

#### POST /purchases
Create purchase order.

**Request Body:**
```json
{
  "vendor_id": 2,
  "date": "2024-01-15",
  "delivery_date": "2024-01-20",
  "reference": "RFQ-001",
  "items": [
    {
      "product_id": 1,
      "quantity": 10,
      "unit_price": 95000.00,
      "description": "Raw material"
    }
  ]
}
```

#### POST /purchases/:id/submit-approval
Submit purchase for approval.

#### POST /purchases/:id/approve
Approve purchase order (approver only).

**Request Body:**
```json
{
  "notes": "Approved by finance director"
}
```

#### POST /purchases/:id/reject
Reject purchase order.

**Request Body:**
```json
{
  "reason": "Budget exceeded",
  "notes": "Please revise quantities"
}
```

#### GET /purchases/pending-approval
Get purchases pending approval (for approvers).

### Payment Processing

#### GET /payments
Get payments list.

**Query Parameters:**
- `type` - Filter: `RECEIVABLE`, `PAYABLE`
- `status` - Filter: `PENDING`, `COMPLETED`, `CANCELLED`
- `account_id` - Filter by payment account

#### POST /payments
Create new payment.

**Request Body:**
```json
{
  "type": "RECEIVABLE",
  "customer_id": 1,
  "sales_ids": [1, 2],
  "amount": 1500000.00,
  "payment_date": "2024-01-16",
  "payment_method": "BANK_TRANSFER",
  "account_id": 2,
  "reference": "TRF789012",
  "notes": "Payment for multiple invoices"
}
```

#### GET /payments/unpaid-invoices/:customer_id
Get unpaid invoices untuk customer.

#### GET /payments/unpaid-bills/:vendor_id
Get unpaid bills untuk vendor.

## 📈 Reports & Analytics

### Financial Reports

#### GET /reports/balance-sheet
Generate Balance Sheet.

**Query Parameters:**
- `as_of_date` (required) - Balance sheet date
- `comparative` (boolean) - Include comparative period
- `format` - Response format: `json`, `pdf`, `excel`

**Response (JSON):**
```json
{
  "success": true,
  "data": {
    "report_header": {
      "report_type": "BALANCE_SHEET",
      "company_name": "Your Company",
      "as_of_date": "2024-12-31",
      "generated_at": "2024-01-15T10:30:00Z"
    },
    "assets": {
      "current_assets": {
        "items": [
          {
            "account_code": "1001",
            "account_name": "Cash in Hand",
            "amount": 1500000.00
          }
        ],
        "total": 15000000.00
      },
      "non_current_assets": {
        "items": [],
        "total": 25000000.00
      },
      "total": 40000000.00
    },
    "liabilities": {
      "current_liabilities": {
        "total": 8000000.00
      },
      "non_current_liabilities": {
        "total": 12000000.00
      },
      "total": 20000000.00
    },
    "equity": {
      "total": 20000000.00
    },
    "is_balanced": true
  }
}
```

#### GET /reports/ssot-profit-loss
Generate Profit & Loss Statement dari SSOT Journal System.

**Query Parameters:**
- `start_date` (required)
- `end_date` (required)
- `format` - `json`, `pdf`, `excel`
- `comparative` (boolean)

**Response:**
```json
{
  "success": true,
  "data": {
    "report_header": {
      "report_type": "PROFIT_LOSS",
      "period": "January 1, 2024 - December 31, 2024"
    },
    "revenue": {
      "sales_revenue": {
        "items": [
          {
            "account_code": "4001",
            "account_name": "Product Sales",
            "amount": 50000000.00
          }
        ],
        "total": 52000000.00
      },
      "total": 52000000.00
    },
    "cost_of_goods_sold": {
      "total": 31200000.00
    },
    "gross_profit": 20800000.00,
    "gross_profit_margin": 40.00,
    "operating_expenses": {
      "total": 15600000.00
    },
    "operating_income": 5200000.00,
    "operating_margin": 10.00,
    "net_income": 4680000.00,
    "net_income_margin": 9.00
  }
}
```

#### GET /ssot-reports/trial-balance
Generate Trial Balance dari SSOT system.

**Query Parameters:**
- `as_of_date` (required)
- `show_zero_balances` (boolean, default: false)
- `format`

#### GET /reports/ssot/cash-flow
Generate Cash Flow Statement.

**Query Parameters:**
- `start_date` (required)
- `end_date` (required)
- `method` - `DIRECT` atau `INDIRECT`
- `format`

### Operational Reports

#### GET /reports/sales-summary
Sales summary dan analytics.

**Query Parameters:**
- `start_date`, `end_date`
- `customer_id` (optional)
- `product_id` (optional)
- `group_by` - `DAY`, `WEEK`, `MONTH`

#### GET /ssot-reports/purchase-report
Purchase analysis report.

#### GET /reports/inventory-report
Inventory valuation dan movement report.

**Query Parameters:**
- `as_of_date`
- `location_id` (optional)
- `category_id` (optional)
- `valuation_method` - `FIFO`, `LIFO`, `AVERAGE`

### Journal Entry Analysis

#### POST /journal-drilldown
Get detailed journal entries dengan filtering.

**Request Body:**
```json
{
  "start_date": "2024-01-01",
  "end_date": "2024-12-31",
  "account_codes": ["1001", "4001"],
  "account_ids": [1, 15],
  "report_type": "PROFIT_LOSS",
  "line_item_name": "Sales Revenue",
  "min_amount": 1000000.00,
  "status": "POSTED",
  "page": 1,
  "limit": 50
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "journal_entries": [
      {
        "id": 1,
        "journal_code": "JE-2024-001",
        "date": "2024-01-15",
        "description": "Sales Invoice SO-2024-001",
        "reference": "SO-2024-001",
        "reference_type": "SALES",
        "total_amount": 2750000.00,
        "status": "POSTED",
        "items": [
          {
            "account_code": "1201",
            "account_name": "Accounts Receivable",
            "debit_amount": 2750000.00,
            "credit_amount": 0.00
          },
          {
            "account_code": "4001", 
            "account_name": "Product Sales",
            "debit_amount": 0.00,
            "credit_amount": 2500000.00
          }
        ]
      }
    ],
    "summary": {
      "total_entries": 156,
      "total_debit": 125000000.00,
      "total_credit": 125000000.00,
      "date_range": {
        "start": "2024-01-01",
        "end": "2024-12-31"
      }
    }
  }
}
```

## 🔧 Administration

### User Management

#### GET /users
Get users list (admin only).

#### POST /users
Create new user (admin only).

**Request Body:**
```json
{
  "username": "newuser",
  "password": "securepassword",
  "email": "user@company.com",
  "role": "FINANCE",
  "is_active": true,
  "permissions": {
    "sales": ["view", "create"],
    "reports": ["view", "export"]
  }
}
```

### Permissions Management

#### GET /permissions/users/:userId
Get user permissions.

#### PUT /permissions/users/:userId
Update user permissions.

**Request Body:**
```json
{
  "permissions": {
    "sales": ["view", "create", "edit"],
    "purchases": ["view", "approve"],
    "reports": ["view", "export"],
    "accounts": ["view"]
  }
}
```

### System Monitoring

#### GET /health
Basic health check.

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-15T10:30:00Z",
    "version": "1.0.0",
    "uptime": "25h 30m 15s"
  }
}
```

#### GET /monitoring/status
Detailed system status.

**Response:**
```json
{
  "success": true,
  "data": {
    "system": {
      "status": "healthy",
      "memory_usage": "512MB",
      "cpu_usage": "15%"
    },
    "database": {
      "status": "connected",
      "connections": {
        "active": 5,
        "idle": 15,
        "max": 100
      },
      "query_performance": "good"
    },
    "services": {
      "balance_sync": "running",
      "websocket": "active",
      "reports": "ready"
    }
  }
}
```

#### GET /admin/balance-health/check
Check balance system health (admin only).

**Response:**
```json
{
  "success": true,
  "data": {
    "overall_status": "HEALTHY",
    "total_accounts": 150,
    "accounts_with_mismatch": 0,
    "total_balance_checked": 150,
    "sync_status": "UP_TO_DATE",
    "last_sync": "2024-01-15T10:25:00Z",
    "mismatches": []
  }
}
```

#### POST /admin/balance-health/auto-heal
Auto-heal balance mismatches (admin only).

#### GET /monitoring/performance/metrics
Get performance metrics.

**Response:**
```json
{
  "success": true,
  "data": {
    "api_performance": {
      "average_response_time": "45ms",
      "slowest_endpoints": [
        {
          "endpoint": "/api/v1/reports/balance-sheet",
          "avg_time": "850ms",
          "call_count": 25
        }
      ]
    },
    "database_performance": {
      "avg_query_time": "15ms",
      "slow_queries": 2,
      "connection_pool_usage": "65%"
    }
  }
}
```

## 🔌 WebSocket

### Connection
```javascript
const ws = new WebSocket('ws://localhost:8080/ws?token=your_jwt_token');
```

### Message Types

**Balance Update:**
```json
{
  "type": "balance_update",
  "data": {
    "account_id": 1,
    "account_code": "1001",
    "old_balance": 1500000.00,
    "new_balance": 1750000.00,
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

**System Notification:**
```json
{
  "type": "system_notification",
  "data": {
    "level": "INFO",
    "title": "System Maintenance",
    "message": "Scheduled maintenance in 30 minutes",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

**Approval Notification:**
```json
{
  "type": "approval_notification",
  "data": {
    "type": "PURCHASE_APPROVAL",
    "id": 123,
    "message": "Purchase PO-2024-001 requires your approval",
    "priority": "HIGH"
  }
}
```

## ❌ Error Handling

### HTTP Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `422` - Validation Error
- `429` - Rate Limited
- `500` - Internal Server Error

### Error Codes
- `AUTH_001` - Invalid credentials
- `AUTH_002` - Token expired
- `AUTH_003` - Insufficient permissions
- `BAL_001` - Balance mismatch detected
- `VAL_001` - Validation failed
- `BIZ_001` - Business rule violation
- `DB_001` - Database error
- `SYS_001` - System error

### Sample Error Responses

**Validation Error:**
```json
{
  "success": false,
  "error": {
    "code": "VAL_001",
    "message": "Validation failed",
    "details": {
      "field_errors": {
        "amount": "Amount must be greater than zero",
        "date": "Date is required"
      }
    }
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Business Rule Violation:**
```json
{
  "success": false,
  "error": {
    "code": "BIZ_001",
    "message": "Insufficient stock",
    "details": {
      "product_id": 1,
      "requested": 10,
      "available": 5
    }
  }
}
```

## 🚦 Rate Limiting

### Limits by Endpoint Type
- **Authentication**: 10 requests/minute
- **General API**: 100 requests/minute  
- **Reports**: 30 requests/minute
- **Admin**: 50 requests/minute

### Rate Limit Headers
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642248600
```

### Rate Limit Exceeded Response
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests",
    "details": {
      "limit": 100,
      "window": "60s",
      "retry_after": 45
    }
  }
}
```

## 📝 Usage Examples

### Complete Sales Flow
```bash
# 1. Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# 2. Create Sales Order
curl -X POST http://localhost:8080/api/v1/sales \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 1,
    "date": "2024-01-15",
    "items": [
      {
        "product_id": 1,
        "quantity": 5,
        "unit_price": 500000
      }
    ]
  }'

# 3. Confirm Sales Order
curl -X POST http://localhost:8080/api/v1/sales/1/confirm \
  -H "Authorization: Bearer $TOKEN"

# 4. Generate Invoice
curl -X POST http://localhost:8080/api/v1/sales/1/invoice \
  -H "Authorization: Bearer $TOKEN"

# 5. Record Payment
curl -X POST http://localhost:8080/api/v1/sales/1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 2750000,
    "payment_date": "2024-01-16",
    "payment_method": "BANK_TRANSFER",
    "account_id": 2,
    "reference": "TRF123456"
  }'
```

### Generate Reports
```bash
# Balance Sheet PDF
curl -X GET "http://localhost:8080/api/v1/reports/balance-sheet?as_of_date=2024-12-31&format=pdf" \
  -H "Authorization: Bearer $TOKEN" \
  --output balance_sheet.pdf

# Profit & Loss JSON
curl -X GET "http://localhost:8080/api/v1/reports/ssot-profit-loss?start_date=2024-01-01&end_date=2024-12-31" \
  -H "Authorization: Bearer $TOKEN"

# Trial Balance
curl -X GET "http://localhost:8080/api/v1/ssot-reports/trial-balance?as_of_date=2024-12-31" \
  -H "Authorization: Bearer $TOKEN"
```

### Journal Analysis
```bash
# Drill-down analysis
curl -X POST http://localhost:8080/api/v1/journal-drilldown \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "start_date": "2024-01-01",
    "end_date": "2024-12-31",
    "account_codes": ["4001"],
    "report_type": "PROFIT_LOSS",
    "page": 1,
    "limit": 50
  }'
```

---

## 🔄 Versioning & Changelog

### Current Version: v1.0

### API Changes
- **v1.0**: Initial release with full functionality
- Enhanced reporting dengan SSOT integration
- Real-time balance updates via WebSocket
- Comprehensive permission system

### Upcoming Changes (v1.1)
- GraphQL endpoint support
- Bulk operations API
- Advanced filtering pada reports
- API rate limiting per user

---

**📚 Related Documentation:**
- [User Guide](README_COMPREHENSIVE.md) - General usage guide
- [Features Documentation](FEATURES.md) - Detailed feature explanations  
- [Technical Guide](TECHNICAL_GUIDE.md) - Technical implementation details

**🔗 Interactive Documentation:**
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **OpenAPI Spec**: http://localhost:8080/openapi/doc.json

---

**📞 API Support:**
For API-specific questions atau issues:
1. Check Swagger documentation first
2. Review error codes dan messages
3. Enable debug logging untuk detailed information
4. Contact technical team dengan complete request/response examples