import React, { useState, useCallback, useMemo } from 'react';
import { 
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogActions,
} from '@mui/material';
import {
  Button,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Alert,
  CircularProgress,
  Box,
  Typography,
  InputAdornment,
} from '@mui/material';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { format } from 'date-fns';
import { fastPaymentService } from '../services/fastPaymentService';

interface FastPaymentFormProps {
  open: boolean;
  onClose: () => void;
  onSuccess: (response: any) => void;
  saleData: {
    sale_id: number;
    invoice_number: string;
    customer: {
      name: string;
    };
    total_amount: number;
    outstanding_amount: number;
  };
}

interface PaymentFormData {
  amount: number;
  payment_date: Date;
  method: string;
  cash_bank_id: number;
  reference: string;
  notes: string;
}

const PAYMENT_METHODS = [
  { value: 'BANK_TRANSFER', label: 'Bank Transfer' },
  { value: 'CASH', label: 'Cash' },
  { value: 'CHECK', label: 'Check' },
  { value: 'CREDIT_CARD', label: 'Credit Card' },
];

// Mock cash bank accounts - in real app, this would come from API
const CASH_BANK_ACCOUNTS = [
  { id: 1, name: 'BCA - Main Account', code: 'BCA-001' },
  { id: 2, name: 'Mandiri - Operational', code: 'MDR-001' },
  { id: 3, name: 'Cash - Petty Cash', code: 'CSH-001' },
];

const FastPaymentForm: React.FC<FastPaymentFormProps> = ({
  open,
  onClose,
  onSuccess,
  saleData,
}) => {
  const [formData, setFormData] = useState<PaymentFormData>({
    amount: saleData?.outstanding_amount || 0,
    payment_date: new Date(),
    method: 'BANK_TRANSFER',
    cash_bank_id: 1,
    reference: saleData?.invoice_number || '',
    notes: '',
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [validationErrors, setValidationErrors] = useState<any>({});
  const [isValidating, setIsValidating] = useState(false);

  // Debounced validation
  const validateFormAsync = useCallback(
    async (data: PaymentFormData) => {
      if (!data.amount || data.amount <= 0) return;

      setIsValidating(true);
      try {
        const payload = {
          sale_id: saleData.sale_id,
          amount: data.amount,
          payment_date: format(data.payment_date, 'yyyy-MM-dd'),
          method: data.method,
          cash_bank_id: data.cash_bank_id,
          reference: data.reference,
          notes: data.notes,
        };

        await fastPaymentService.validatePayment(payload);
        setValidationErrors({});
      } catch (err: any) {
        console.log('Validation warning:', err.message); // Don't show validation errors aggressively
      } finally {
        setIsValidating(false);
      }
    },
    [saleData.sale_id]
  );

  // Fast form validation
  const isFormValid = useMemo(() => {
    return (
      formData.amount > 0 &&
      formData.amount <= saleData.outstanding_amount &&
      formData.payment_date &&
      formData.method &&
      formData.cash_bank_id &&
      formData.reference.length > 0
    );
  }, [formData, saleData.outstanding_amount]);

  // Handle form changes with minimal re-renders
  const handleChange = useCallback((field: keyof PaymentFormData, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));

    // Clear specific field errors
    if (validationErrors[field]) {
      setValidationErrors(prev => ({
        ...prev,
        [field]: null,
      }));
    }
  }, [validationErrors]);

  // Handle form submission with optimistic UI
  const handleSubmit = async (useAsync = true) => {
    if (!isFormValid) return;

    setLoading(true);
    setError(null);

    try {
      const payload = {
        sale_id: saleData.sale_id,
        amount: formData.amount,
        payment_date: format(formData.payment_date, 'yyyy-MM-dd'),
        method: formData.method,
        cash_bank_id: formData.cash_bank_id,
        reference: formData.reference,
        notes: formData.notes,
      };

      // Use async processing for better UX
      const response = useAsync
        ? await fastPaymentService.recordPaymentAsync(payload)
        : await fastPaymentService.recordPaymentFast(payload);

      // Immediate success feedback
      onSuccess(response);
      onClose();

      // Reset form
      setFormData({
        amount: 0,
        payment_date: new Date(),
        method: 'BANK_TRANSFER',
        cash_bank_id: 1,
        reference: '',
        notes: '',
      });

    } catch (err: any) {
      console.error('Payment submission error:', err);
      setError(err.message || 'Failed to record payment. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  // Calculate remaining amount for validation
  const remainingAmount = saleData.outstanding_amount - formData.amount;

  return (
    <LocalizationProvider dateAdapter={AdapterDateFns}>
      <Dialog 
        open={open} 
        onClose={onClose} 
        maxWidth="sm" 
        fullWidth
        PaperProps={{
          sx: { minHeight: '500px' } // Consistent size to prevent layout shifts
        }}
      >
        <DialogHeader>
          <DialogTitle>
            🚀 Fast Payment Recording
            {loading && (
              <CircularProgress 
                size={20} 
                sx={{ ml: 2 }} 
                color="primary"
              />
            )}
          </DialogTitle>
        </DialogHeader>

        <DialogContent>
          {/* Sale Info - Minimal display */}
          <Box sx={{ mb: 3, p: 2, bgcolor: 'grey.50', borderRadius: 1 }}>
            <Typography variant="subtitle2" color="text.secondary">
              Invoice: {saleData.invoice_number} | Customer: {saleData.customer.name}
            </Typography>
            <Typography variant="h6" color="primary">
              Outstanding: Rp {saleData.outstanding_amount.toLocaleString('id-ID')}
            </Typography>
          </Box>

          {/* Error Display */}
          {error && (
            <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
              {error}
            </Alert>
          )}

          {/* Payment Form - Optimized fields only */}
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            {/* Amount - Most important field first */}
            <TextField
              label="Payment Amount"
              type="number"
              value={formData.amount}
              onChange={(e) => handleChange('amount', parseFloat(e.target.value) || 0)}
              fullWidth
              required
              inputProps={{ 
                min: 0, 
                max: saleData.outstanding_amount,
                step: 0.01
              }}
              InputProps={{
                startAdornment: <InputAdornment position="start">Rp</InputAdornment>,
              }}
              helperText={
                formData.amount > saleData.outstanding_amount
                  ? 'Amount exceeds outstanding balance'
                  : remainingAmount > 0
                  ? `Remaining: Rp ${remainingAmount.toLocaleString('id-ID')}`
                  : formData.amount === saleData.outstanding_amount
                  ? '✅ Full payment'
                  : ''
              }
              error={formData.amount > saleData.outstanding_amount}
            />

            {/* Payment Date */}
            <DatePicker
              label="Payment Date"
              value={formData.payment_date}
              onChange={(date) => handleChange('payment_date', date || new Date())}
              renderInput={(params) => (
                <TextField {...params} required fullWidth />
              )}
              maxDate={new Date()}
            />

            {/* Payment Method & Account - Single Row */}
            <Box sx={{ display: 'flex', gap: 2 }}>
              <FormControl required sx={{ flex: 1 }}>
                <InputLabel>Payment Method</InputLabel>
                <Select
                  value={formData.method}
                  onChange={(e) => handleChange('method', e.target.value)}
                  label="Payment Method"
                >
                  {PAYMENT_METHODS.map((method) => (
                    <MenuItem key={method.value} value={method.value}>
                      {method.label}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>

              <FormControl required sx={{ flex: 1 }}>
                <InputLabel>Account</InputLabel>
                <Select
                  value={formData.cash_bank_id}
                  onChange={(e) => handleChange('cash_bank_id', e.target.value)}
                  label="Account"
                >
                  {CASH_BANK_ACCOUNTS.map((account) => (
                    <MenuItem key={account.id} value={account.id}>
                      {account.name}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Box>

            {/* Reference - Auto-filled */}
            <TextField
              label="Reference"
              value={formData.reference}
              onChange={(e) => handleChange('reference', e.target.value)}
              fullWidth
              required
              placeholder="Payment reference number"
            />

            {/* Notes - Optional */}
            <TextField
              label="Notes"
              value={formData.notes}
              onChange={(e) => handleChange('notes', e.target.value)}
              fullWidth
              multiline
              rows={2}
              placeholder="Additional notes (optional)"
            />
          </Box>

          {/* Validation Indicator */}
          {isValidating && (
            <Box sx={{ display: 'flex', alignItems: 'center', mt: 1 }}>
              <CircularProgress size={16} />
              <Typography variant="caption" sx={{ ml: 1 }}>
                Validating...
              </Typography>
            </Box>
          )}
        </DialogContent>

        <DialogActions sx={{ p: 3, gap: 1 }}>
          <Button onClick={onClose} disabled={loading}>
            Cancel
          </Button>
          
          {/* Fast Processing Option */}
          <Button
            onClick={() => handleSubmit(false)}
            disabled={!isFormValid || loading}
            variant="outlined"
            startIcon={loading ? <CircularProgress size={16} /> : null}
          >
            Record (Sync)
          </Button>

          {/* Async Processing - Default and Recommended */}
          <Button
            onClick={() => handleSubmit(true)}
            disabled={!isFormValid || loading}
            variant="contained"
            startIcon={loading ? <CircularProgress size={16} /> : null}
          >
            Record Payment
          </Button>
        </DialogActions>
      </Dialog>
    </LocalizationProvider>
  );
};

export default FastPaymentForm;