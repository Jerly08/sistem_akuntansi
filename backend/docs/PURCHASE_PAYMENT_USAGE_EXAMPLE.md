# Purchase Payment Integration - Usage Example

## Implementation in Purchase Management Page

Here's how to integrate the Purchase Payment functionality into your Purchase Management page:

### 1. Import Required Components and Services

```tsx
import React, { useState, useEffect } from 'react';
import EnhancedPurchaseTable from '@/components/purchase/EnhancedPurchaseTable';
import PurchasePaymentForm from '@/components/purchase/PurchasePaymentForm';
import purchaseService, { Purchase } from '@/services/purchaseService';
// ... other imports

interface CashBank {
  id: number;
  name: string;
  account_code: string;
  balance: number;
}
```

### 2. Add State Management

```tsx
const PurchaseManagementPage = () => {
  const [purchases, setPurchases] = useState<Purchase[]>([]);
  const [cashBanks, setCashBanks] = useState<CashBank[]>([]);
  const [selectedPurchaseForPayment, setSelectedPurchaseForPayment] = useState<Purchase | null>(null);
  const [isPaymentModalOpen, setIsPaymentModalOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  // Load data on component mount
  useEffect(() => {
    loadPurchases();
    loadCashBanks();
  }, []);

  const loadPurchases = async () => {
    setLoading(true);
    try {
      const result = await purchaseService.list({});
      setPurchases(result.data);
    } catch (error) {
      console.error('Error loading purchases:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadCashBanks = async () => {
    try {
      // You'll need to implement this in your cashBankService
      // const banks = await cashBankService.list();
      // setCashBanks(banks);
      
      // Mock data for example
      setCashBanks([
        { id: 1, name: 'Bank BCA', account_code: '1101-001', balance: 50000000 },
        { id: 2, name: 'Bank Mandiri', account_code: '1101-002', balance: 25000000 },
      ]);
    } catch (error) {
      console.error('Error loading cash banks:', error);
    }
  };
```

### 3. Add Handler Functions

```tsx
  // Handler for Record Payment button click
  const handleRecordPayment = (purchase: Purchase) => {
    setSelectedPurchaseForPayment(purchase);
    setIsPaymentModalOpen(true);
  };

  // Handler for successful payment recording
  const handlePaymentSuccess = (result: any) => {
    console.log('Payment recorded successfully:', result);
    
    // Refresh purchases list to show updated amounts
    loadPurchases();
    
    // You might want to show a detailed success message
    // or redirect to Payment Management to see the payment
  };

  // Close payment modal
  const handleClosePaymentModal = () => {
    setIsPaymentModalOpen(false);
    setSelectedPurchaseForPayment(null);
  };

  // Other existing handlers...
  const handleViewDetails = (purchase: Purchase) => {
    // Your view details implementation
  };

  const handleEdit = (purchase: Purchase) => {
    // Your edit implementation
  };

  const handleDelete = (purchaseId: number) => {
    // Your delete implementation
  };

  const handleSubmitForApproval = (purchaseId: number) => {
    // Your submit for approval implementation
  };
```

### 4. Add Helper Functions

```tsx
  const formatCurrency = (amount: number): string => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount);
  };

  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleDateString('id-ID', {
      day: '2-digit',
      month: 'long',
      year: 'numeric',
    });
  };
```

### 5. Render Components

```tsx
  return (
    <Box>
      <VStack spacing={6} align="stretch">
        {/* Your existing filters, buttons, etc. */}
        
        {/* Enhanced Purchase Table with Payment Integration */}
        <EnhancedPurchaseTable
          purchases={purchases}
          loading={loading}
          onViewDetails={handleViewDetails}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onSubmitForApproval={handleSubmitForApproval}
          onRecordPayment={handleRecordPayment} // NEW: Payment handler
          formatCurrency={formatCurrency}
          formatDate={formatDate}
          canEdit={true}
          canDelete={true}
          userRole="admin" // or get from your auth context
        />

        {/* Purchase Payment Modal */}
        <PurchasePaymentForm
          isOpen={isPaymentModalOpen}
          onClose={handleClosePaymentModal}
          purchase={selectedPurchaseForPayment}
          onSuccess={handlePaymentSuccess}
          cashBanks={cashBanks}
        />
      </VStack>
    </Box>
  );
};

export default PurchaseManagementPage;
```

## 6. Backend Routes Setup

Make sure to add the new routes to your backend router:

```go
// In purchase_routes.go
func SetupPurchaseRoutes(router *gin.RouterGroup, purchaseController *controllers.PurchaseController) {
    // ... existing routes

    // Purchase Payment Integration routes
    router.GET("/:id/for-payment", purchaseController.GetPurchaseForPayment)
    router.POST("/:id/integrated-payment", purchaseController.CreateIntegratedPayment)
    router.GET("/:id/payments", purchaseController.GetPurchasePayments)
}
```

## 7. Database Migration

Run the migration to add the required tables and fields:

```sql
-- Run this migration: 011_purchase_payment_integration.sql
```

## Features Implemented

✅ **Record Payment Button** - Only appears for APPROVED credit purchases with outstanding amounts
✅ **Payment Form** - Comprehensive form with validation
✅ **Cross-Reference Tracking** - Links purchases to payment management
✅ **Outstanding Amount Tracking** - Updates paid and outstanding amounts automatically  
✅ **Status Updates** - Changes status to PAID when fully paid
✅ **Integration with Payment Management** - Payments appear in both systems
✅ **Validation** - Prevents overpayment and invalid data
✅ **User Feedback** - Toast notifications for success/error states

## Workflow

1. **Create Purchase** with payment_method: "CREDIT"
2. **Approve Purchase** → Status becomes "APPROVED", Outstanding = Total Amount
3. **Record Payment** via "Record Payment" button → Opens payment form
4. **Submit Payment** → Creates records in both Purchase and Payment Management
5. **View in Payment Management** → Payment appears with proper cross-reference
6. **Track Outstanding** → Outstanding amount decreases, status updates when fully paid

## Testing Checklist

- [ ] Record Payment button only shows for APPROVED CREDIT purchases
- [ ] Payment form validates amount against outstanding balance
- [ ] Payment successfully creates records in both systems
- [ ] Outstanding amount updates correctly after payment
- [ ] Status changes to PAID when fully paid
- [ ] Payment appears in Payment Management list
- [ ] Cross-reference data is correct in database
- [ ] Toast notifications show appropriate messages

This integration provides a seamless experience for managing vendor payments while maintaining data consistency across both Purchase and Payment Management systems.