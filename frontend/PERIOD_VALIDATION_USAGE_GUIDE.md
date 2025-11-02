# Frontend Period Validation - Usage Guide

## Overview

Panduan lengkap implementasi period validation error handling di frontend dengan UX yang informatif untuk user.

---

## Files Created

1. **`src/hooks/usePeriodValidation.ts`** - Custom hook untuk handle period errors
2. **`src/components/periods/ReopenPeriodDialog.tsx`** - Dialog untuk reopen period
3. **Backend Test Script:** `backend/cmd/scripts/test_period_validation.go`

---

## Quick Start

### Step 1: Import Hook dan Dialog

```typescript
// In your component (e.g., SaleForm.tsx, PurchaseForm.tsx)
import { usePeriodValidation } from '@/hooks/usePeriodValidation';
import { ReopenPeriodDialog } from '@/components/periods/ReopenPeriodDialog';
import api from '@/services/api'; // Your axios instance
```

### Step 2: Setup Hook dalam Component

```typescript
function SaleForm() {
  const {
    handlePeriodError,
    reopenPeriod,
    reopenDialogOpen,
    periodToReopen,
    isReopening,
    closeReopenDialog,
  } = usePeriodValidation({
    onReopenSuccess: (period) => {
      console.log(`Period ${period} reopened, retrying transaction...`);
      // Optionally retry the original request
      handleSubmit();
    },
  });

  // Your existing states...
  const [formData, setFormData] = useState({...});
  const [loading, setLoading] = useState(false);

  // ... rest of component
}
```

### Step 3: Wrap Submit Handler

```typescript
const handleSubmit = async () => {
  setLoading(true);

  try {
    // Your normal API call
    const response = await api.post('/sales', formData);
    
    toast.success('Sale created successfully');
    navigate('/sales');
    
  } catch (error: any) {
    // ✅ Handle period validation error
    const wasHandled = handlePeriodError(error);
    
    if (!wasHandled) {
      // Handle other errors
      toast.error(error?.response?.data?.message || 'Failed to create sale');
    }
  } finally {
    setLoading(false);
  }
};
```

### Step 4: Add Reopen Dialog to JSX

```typescript
return (
  <>
    {/* Your form JSX */}
    <form onSubmit={handleSubmit}>
      {/* ... form fields ... */}
    </form>

    {/* Period Reopen Dialog */}
    {periodToReopen && (
      <ReopenPeriodDialog
        open={reopenDialogOpen}
        period={periodToReopen.period}
        year={periodToReopen.year}
        month={periodToReopen.month}
        onClose={closeReopenDialog}
        onReopen={(year, month, reason) => reopenPeriod(year, month, reason, api)}
        isLoading={isReopening}
      />
    )}
  </>
);
```

---

## Complete Example: Sale Form

```typescript
// components/sales/SaleForm.tsx
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'react-toastify';
import {
  Box,
  Button,
  Card,
  CardContent,
  TextField,
  CircularProgress,
} from '@mui/material';
import { Save } from '@mui/icons-material';

import { usePeriodValidation } from '@/hooks/usePeriodValidation';
import { ReopenPeriodDialog } from '@/components/periods/ReopenPeriodDialog';
import api from '@/services/api';

interface SaleFormData {
  customer_id: number;
  date: string;
  due_date: string;
  items: Array<{
    product_id: number;
    quantity: number;
    unit_price: number;
  }>;
}

export const SaleForm: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<SaleFormData>({
    customer_id: 0,
    date: new Date().toISOString().split('T')[0],
    due_date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
    items: [],
  });

  // ✅ Setup period validation hook
  const {
    handlePeriodError,
    reopenPeriod,
    reopenDialogOpen,
    periodToReopen,
    isReopening,
    closeReopenDialog,
  } = usePeriodValidation({
    onReopenSuccess: (period) => {
      toast.success(`Period ${period} reopened. You can now submit again.`);
      // Optionally auto-retry
      // handleSubmit();
    },
    onReopenError: (error) => {
      console.error('Failed to reopen period:', error);
    },
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const response = await api.post('/sales', {
        ...formData,
        type: 'INVOICE',
      });

      toast.success('Sale created successfully!');
      navigate('/sales');
      
    } catch (error: any) {
      console.error('Sale creation error:', error);

      // ✅ Handle period validation error
      const wasHandled = handlePeriodError(error);

      if (!wasHandled) {
        // Handle other types of errors
        const errorMessage = 
          error?.response?.data?.message || 
          error?.response?.data?.error ||
          'Failed to create sale';
        
        toast.error(errorMessage);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Card>
        <CardContent>
          <form onSubmit={handleSubmit}>
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              
              {/* Date Field - This will be validated */}
              <TextField
                label="Transaction Date"
                type="date"
                value={formData.date}
                onChange={(e) => setFormData({ ...formData, date: e.target.value })}
                InputLabelProps={{ shrink: true }}
                required
                helperText="Ensure period is open for this date"
              />

              <TextField
                label="Due Date"
                type="date"
                value={formData.due_date}
                onChange={(e) => setFormData({ ...formData, due_date: e.target.value })}
                InputLabelProps={{ shrink: true }}
                required
              />

              {/* Other form fields... */}

              <Box sx={{ display: 'flex', gap: 2, justifyContent: 'flex-end' }}>
                <Button
                  variant="outlined"
                  onClick={() => navigate('/sales')}
                  disabled={loading}
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  variant="contained"
                  startIcon={loading ? <CircularProgress size={20} /> : <Save />}
                  disabled={loading}
                >
                  {loading ? 'Saving...' : 'Save Sale'}
                </Button>
              </Box>
            </Box>
          </form>
        </CardContent>
      </Card>

      {/* ✅ Reopen Period Dialog */}
      {periodToReopen && (
        <ReopenPeriodDialog
          open={reopenDialogOpen}
          period={periodToReopen.period}
          year={periodToReopen.year}
          month={periodToReopen.month}
          onClose={closeReopenDialog}
          onReopen={(year, month, reason) => reopenPeriod(year, month, reason, api)}
          isLoading={isReopening}
        />
      )}
    </>
  );
};
```

---

## Error Messages (Bahasa Indonesia)

Hook `usePeriodValidation` akan menampilkan pesan error yang informatif:

### 1. PERIOD_CLOSED
```
❌ Tidak dapat membuat transaksi: Periode 2025-01 sudah ditutup.
💡 Anda dapat membuka kembali periode ini jika memiliki permission.

[Dialog Reopen akan muncul]
```

### 2. PERIOD_LOCKED
```
❌ Tidak dapat membuat transaksi: Periode 2025-12 telah dikunci secara permanen 
   (fiscal year-end closing).
💡 Periode sudah dikunci permanen. Hubungi administrator untuk bantuan.
```

### 3. DATE_TOO_OLD
```
❌ Tanggal transaksi terlalu lama (lebih dari 2 tahun). 
   Periode tidak dapat dibuat otomatis.
💡 Silakan gunakan tanggal yang sesuai atau buat periode secara manual.
```

### 4. DATE_TOO_FUTURE
```
❌ Tanggal transaksi terlalu jauh ke depan (lebih dari 7 hari). 
   Gunakan tanggal yang lebih dekat.
💡 Silakan gunakan tanggal yang sesuai atau buat periode secara manual.
```

---

## User Flow

### Scenario 1: Normal Transaction (OPEN Period)
```
User → Fill form with date 2025-12-15
     → Click "Save"
     → ✅ Success: Transaction created
```

### Scenario 2: Transaction to CLOSED Period
```
User → Fill form with date 2025-01-15
     → Click "Save"
     → ❌ Error Toast: "Periode 2025-01 sudah ditutup"
     → 🔓 Dialog appears: "Buka Kembali Periode?"
     
     IF user has permission:
       → User enters reason: "Need to add correction entry"
       → Click "Buka Periode"
       → ✅ Success: Period reopened
       → User can retry submission
     
     ELSE:
       → User sees: "You don't have permission to reopen periods"
```

### Scenario 3: Transaction to LOCKED Period
```
User → Fill form with date 2024-12-31
     → Click "Save"
     → ❌ Error Toast: "Periode 2024-12 telah dikunci secara permanen"
     → 💬 Message: "Hubungi administrator untuk bantuan"
     → ❌ No reopen option (LOCKED is permanent)
```

---

## API Client Setup

Make sure your axios instance is configured properly:

```typescript
// services/api.ts
import axios from 'axios';

const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token interceptor
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export default api;
```

---

## Testing with Backend Script

### Step 1: Update credentials in test script

```go
// backend/cmd/scripts/test_period_validation.go
const (
    baseURL = "http://localhost:8080/api/v1"
)

func getAuthToken() string {
    loginData := map[string]interface{}{
        "email":    "your-email@example.com",  // ← Update this
        "password": "your-password",            // ← Update this
    }
    // ...
}
```

### Step 2: Run the test script

```bash
cd backend/cmd/scripts
go run test_period_validation.go
```

### Expected Output:

```
🧪 Period Validation Testing Suite
===================================

✅ Authenticated successfully

📋 Running Test Scenarios...
----------------------------

🟢 Test 1: Create Sale to OPEN Period
   ✅ SUCCESS: Sale created to open period

🔒 Closing Period: 2025-01
   ✅ Period 2025-01 closed successfully

🔴 Test 2: Create Sale to CLOSED Period
   ✅ CORRECT: Transaction blocked
   Error Code: PERIOD_CLOSED
   Details: cannot post to 2025-01 period (status: CLOSED)
   Period: 2025-01

🔓 Reopening Period: 2025-01
   ✅ Period 2025-01 reopened successfully

🟢 Test 3: Create Sale to REOPENED Period
   ✅ SUCCESS: Sale created to reopened period

📊 Test Summary
===============
✅ PASSED - Open Period - Create Sale (Status: 200)
✅ PASSED - Closed Period - Create Sale (Status: 403)
✅ PASSED - Reopened Period - Create Sale (Status: 200)

📈 Results: 7 passed, 0 failed (Total: 7)
🎉 All tests passed!
```

---

## Advanced Usage

### Custom Error Messages

```typescript
const { handlePeriodError, getErrorMessage } = usePeriodValidation({
  showToast: false, // Disable automatic toast
});

// Manual error handling
try {
  await api.post('/sales', data);
} catch (error: any) {
  if (isPeriodValidationError(error)) {
    const message = getErrorMessage(error.response.data);
    // Custom UI feedback
    setCustomError(message);
  }
}
```

### Disable Auto-Reopen Dialog

```typescript
const { handlePeriodError } = usePeriodValidation({
  // Don't show reopen dialog automatically
  onReopenSuccess: undefined,
});

// Manual reopen
const handleManualReopen = async () => {
  const success = await reopenPeriod(2025, 1, "Manual reopen", api);
  if (success) {
    // Retry logic
  }
};
```

---

## Permission Check

Before showing reopen dialog, check if user has permission:

```typescript
// In your auth context or user service
export function useAuth() {
  const user = useSelector((state) => state.auth.user);
  
  const hasPermission = (resource: string, action: string) => {
    // Check user permissions
    return user?.permissions?.[resource]?.[action] === true;
  };

  const canReopenPeriods = () => {
    return (
      user?.role === 'admin' ||
      user?.role === 'finance' ||
      hasPermission('periods', 'reopen')
    );
  };

  return { user, hasPermission, canReopenPeriods };
}

// In your component
const { canReopenPeriods } = useAuth();

// Only show reopen option if user has permission
{canReopenPeriods() && periodToReopen && (
  <ReopenPeriodDialog ... />
)}
```

---

## Best Practices

### 1. Always Show Informative Error
```typescript
// ✅ Good
toast.error(
  <div>
    <strong>Periode {period} sudah ditutup</strong>
    <p>Hubungi finance team untuk membuka periode</p>
  </div>
);

// ❌ Bad
toast.error("Error 403");
```

### 2. Provide Context
```typescript
// ✅ Good
helperText="Ensure the period for this date is open"

// ❌ Bad
helperText="Enter date"
```

### 3. Log for Debugging
```typescript
catch (error: any) {
  console.error('Period validation error:', {
    code: error?.response?.data?.code,
    period: error?.response?.data?.period,
    details: error?.response?.data?.details,
  });
  
  handlePeriodError(error);
}
```

### 4. Handle Edge Cases
```typescript
// What if API is down?
const wasHandled = handlePeriodError(error);

if (!wasHandled) {
  // Network error, server error, etc.
  if (error.code === 'ECONNABORTED') {
    toast.error('Request timeout. Please try again.');
  } else if (!error.response) {
    toast.error('Network error. Check your connection.');
  } else {
    toast.error('Something went wrong. Please try again.');
  }
}
```

---

## Summary

✅ **Created:**
- Custom hook `usePeriodValidation` untuk handle errors
- Reopen dialog component dengan UX yang baik
- Test script untuk backend validation

✅ **Features:**
- Informative error messages (Bahasa Indonesia)
- Auto-detect period errors
- Reopen dialog dengan audit trail
- Permission-aware UI
- Retry mechanism after reopen

✅ **Integration:**
- Works with any form (Sales, Purchase, Payments, etc.)
- Minimal changes to existing code
- Toast notifications dengan custom layout
- Material-UI components

**Next: Test thoroughly dengan user acceptance testing!**
