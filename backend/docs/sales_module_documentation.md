# Sales Module Documentation

## Overview

The Sales Module is a comprehensive system for managing sales transactions, including quotations, orders, invoices, payments, and returns. It integrates with accounting, inventory, and customer management systems.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Controllers   │────│    Services     │────│  Repositories   │
│                 │    │                 │    │                 │
│ sales_controller│    │ sales_service   │    │sales_repository │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Models      │    │   Middleware    │    │    Database     │
│                 │    │                 │    │                 │
│ Sale, SaleItem  │    │ RBAC, JWT, etc  │    │  PostgreSQL/    │
│ SalePayment,etc │    │                 │    │  MySQL          │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Core Models

### Sale
Main sales transaction entity containing:
- Basic Information: Code, Customer, Type, Date, Status
- Financial: Amounts, Discounts, Taxes, Payment Terms
- Relationships: Items, Payments, Returns

### SaleItem
Individual line items within a sale:
- Product details and pricing
- Quantities and discounts
- Tax calculations
- Accounting allocations

### SalePayment
Payment records against sales:
- Payment amounts and methods
- References and notes
- Accounting entries

### SaleReturn
Return/refund transactions:
- Return items and quantities
- Credit notes
- Approval workflows

## Key Features

### 1. Sales Lifecycle Management
- **Draft** → **Confirmed** → **Invoiced** → **Paid**
- Status validation and business rules
- Automated number generation

### 2. Financial Calculations
- Item-level and order-level discounts
- Multi-tax support (PPN, PPh)
- Currency and exchange rate handling
- Real-time total calculations

### 3. Payment Management
- Multiple payment methods
- Partial payment support
- Outstanding balance tracking
- Payment allocation

### 4. Returns & Refunds
- Selective item returns
- Credit note generation
- Inventory restoration
- Approval workflows

### 5. Reporting & Analytics
- Sales summaries and analytics
- Customer reports
- Receivables aging
- PDF generation

## API Endpoints

### Basic CRUD Operations
```http
GET    /api/sales                    # List sales with filters
POST   /api/sales                    # Create new sale
GET    /api/sales/{id}               # Get sale details
PUT    /api/sales/{id}               # Update sale
DELETE /api/sales/{id}               # Delete sale (draft only)
```

### Status Management
```http
POST   /api/sales/{id}/confirm       # Confirm sale
POST   /api/sales/{id}/invoice       # Create invoice
POST   /api/sales/{id}/cancel        # Cancel sale
```

### Payments
```http
GET    /api/sales/{id}/payments      # List payments
POST   /api/sales/{id}/payments      # Record payment
```

### Returns
```http
POST   /api/sales/{id}/returns       # Create return
GET    /api/sales-returns            # List all returns
```

### Reports
```http
GET    /api/sales-reports/summary    # Sales summary
GET    /api/sales-reports/analytics  # Sales analytics
GET    /api/sales-reports/receivables # Receivables report
GET    /api/sales/{id}/invoice/pdf   # Invoice PDF
```

## Request/Response Examples

### Create Sale
```json
{
  "customer_id": 1,
  "sales_person_id": 6,
  "type": "INVOICE",
  "date": "2024-08-17T00:00:00Z",
  "due_date": "2024-09-16T00:00:00Z",
  "discount_percent": 5.0,
  "ppn_percent": 11.0,
  "payment_terms": "NET30",
  "notes": "Rush order",
  "items": [
    {
      "product_id": 1,
      "quantity": 2,
      "unit_price": 100000.00,
      "discount_percent": 0,
      "taxable": true
    }
  ]
}
```

### Sale Response
```json
{
  "id": 1,
  "code": "INV-2024-0001",
  "customer_id": 1,
  "type": "INVOICE",
  "status": "DRAFT",
  "date": "2024-08-17T00:00:00Z",
  "due_date": "2024-09-16T00:00:00Z",
  "subtotal": 200000.00,
  "discount_amount": 10000.00,
  "taxable_amount": 190000.00,
  "ppn": 20900.00,
  "total_amount": 210900.00,
  "outstanding_amount": 210900.00,
  "customer": {
    "id": 1,
    "name": "PT. Example Client",
    "email": "client@example.com"
  },
  "sale_items": [
    {
      "id": 1,
      "product_id": 1,
      "quantity": 2,
      "unit_price": 100000.00,
      "line_total": 200000.00,
      "ppn_amount": 20900.00,
      "final_amount": 200000.00
    }
  ]
}
```

## Business Rules

### Sale Creation
1. Customer must exist and be active
2. All products must exist with sufficient stock
3. Sales person (if specified) must exist and be active
4. Credit limit validation for customer

### Status Transitions
1. **DRAFT** → **CONFIRMED**: Validates items and updates inventory
2. **CONFIRMED** → **INVOICED**: Generates invoice number and creates accounting entries
3. **INVOICED** → **PAID**: Records payments and updates status when fully paid
4. **Any Status** → **CANCELLED**: Reverses inventory and accounting entries

### Payment Rules
1. Payments only allowed on invoiced sales
2. Payment amount cannot exceed outstanding balance
3. Multiple partial payments supported
4. Automatic status update to PAID when fully paid

### Return Rules
1. Returns only allowed on invoiced/paid sales
2. Return quantity cannot exceed sold quantity
3. Returns can be partial or full
4. Credit notes generated for approved returns

## Data Validation

### Required Fields
- Customer ID
- Sale Type (QUOTATION, ORDER, INVOICE)
- Date
- At least one sale item

### Business Validations
- Customer credit limit check
- Stock availability validation
- Sales person existence validation
- Accounting account validation

### Data Integrity
- Referential integrity constraints
- Amount calculation validation
- Status transition validation
- Audit trail maintenance

## Error Handling

### Common Error Codes
- `400`: Validation errors (missing required fields, invalid data)
- `404`: Resource not found (sale, customer, product)
- `409`: Business rule violation (insufficient stock, credit limit exceeded)
- `422`: Invalid status transition
- `500`: Internal server error

### Error Response Format
```json
{
  "error": "Credit limit exceeded",
  "message": "Customer credit limit exceeded. Available: 500000.00, Required: 750000.00",
  "code": "CREDIT_LIMIT_EXCEEDED",
  "details": {
    "customer_id": 1,
    "available_credit": 500000.00,
    "required_amount": 750000.00
  }
}
```

## Integration Points

### Accounting Integration
- Automatic journal entries for sales
- Revenue recognition
- Tax accounting (PPN/PPh)
- Accounts receivable management

### Inventory Integration
- Stock reservation and allocation
- Inventory updates on confirmation
- COGS calculation
- Stock restoration on cancellation

### Customer Management
- Credit limit checking
- Customer history tracking
- Payment behavior analysis

## Performance Considerations

### Database Optimization
- Proper indexing on frequently queried fields
- Pagination for large result sets
- Efficient joins and preloading
- Query optimization for reporting

### Caching Strategy
- Cache frequently accessed customer data
- Cache product information
- Cache calculated totals where appropriate

### Concurrent Access
- Optimistic locking for sale updates
- Transaction management for critical operations
- Race condition prevention for code generation

## Security

### Authentication & Authorization
- JWT-based authentication required for all endpoints
- RBAC implementation with roles:
  - `sales`: Basic sales operations
  - `sales_admin`: Advanced operations (cancel, returns)
  - `sales_reports`: Reporting access

### Data Protection
- Sensitive financial data encryption
- Audit logging for all changes
- Input validation and sanitization
- SQL injection prevention

### Rate Limiting
- API rate limits per user/IP
- Special limits for resource-intensive operations
- Abuse detection and prevention

## Monitoring & Logging

### Audit Trail
- All CRUD operations logged
- Status change tracking
- Payment recording history
- User action attribution

### Performance Monitoring
- Response time tracking
- Database query performance
- Error rate monitoring
- Resource usage tracking

### Business Metrics
- Sales volume tracking
- Conversion rate monitoring
- Customer behavior analysis
- Revenue recognition tracking

## Deployment Considerations

### Environment Setup
- Database migration scripts
- Initial data seeding
- Configuration management
- Environment-specific settings

### Backup & Recovery
- Regular database backups
- Point-in-time recovery capability
- Data retention policies
- Disaster recovery procedures

## Troubleshooting

### Common Issues

1. **Foreign Key Violations**
   - Issue: Sales person ID doesn't exist
   - Solution: Create user or set sales_person_id to null

2. **Calculation Discrepancies**
   - Issue: Totals don't match expected values
   - Solution: Run data integrity fix script

3. **Status Transition Errors**
   - Issue: Invalid status changes
   - Solution: Verify business rules and current status

4. **Performance Issues**
   - Issue: Slow query responses
   - Solution: Check indexes, optimize queries, implement caching

### Diagnostic Commands
```bash
# Check data integrity
go run scripts/fix_sales_data_integrity.go

# Validate sales codes
go run scripts/check_sales_codes.go

# Create missing sales person
go run scripts/create_sales_person.go
```

## Future Enhancements

### Planned Features
1. Multi-warehouse support
2. Advanced pricing rules
3. Subscription billing
4. Integration with external payment gateways
5. Advanced analytics and forecasting

### Technical Improvements
1. GraphQL API support
2. Real-time notifications
3. Event-driven architecture
4. Advanced caching strategies
5. Microservices architecture

## Support & Maintenance

### Regular Maintenance Tasks
1. Data integrity checks (weekly)
2. Performance monitoring (daily)
3. Backup verification (daily)
4. Security audit (monthly)

### Contact Information
- Development Team: dev-team@company.com
- System Administrator: sysadmin@company.com
- Business Owner: business@company.com

---

*Last Updated: August 17, 2024*
*Version: 1.0*
