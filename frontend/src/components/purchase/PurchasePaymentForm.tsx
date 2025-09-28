'use client';

import React, { useState, useEffect } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Button,
  FormControl,
  FormLabel,
  Input,
  Select,
  Textarea,
  VStack,
  HStack,
  Alert,
  AlertIcon,
  Text,
  Divider,
  Box,
  useToast,
  Spinner,
} from '@chakra-ui/react';
import { Purchase, PurchasePaymentRequest } from '@/services/purchaseService';
import purchaseService from '@/services/purchaseService';

interface CashBank {
  id: number;
  name: string;
  account_code: string;
  balance: number;
}

interface PurchasePaymentFormProps {
  isOpen: boolean;
  onClose: () => void;
  purchase: Purchase | null;
  onSuccess?: (result: any) => void;
  cashBanks: CashBank[];
}

const PurchasePaymentForm: React.FC<PurchasePaymentFormProps> = ({
  isOpen,
  onClose,
  purchase,
  onSuccess,
  cashBanks = [],
}) => {
  const toast = useToast();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<PurchasePaymentRequest>({
    amount: 0,
    payment_date: new Date().toISOString().split('T')[0],
    payment_method: 'Bank Transfer',
    cash_bank_id: undefined,
    reference: '',
    notes: '',
  });
  const [displayAmount, setDisplayAmount] = useState('0');

  // Format number to Rupiah display
  const formatRupiah = (value: number | string): string => {
    const numValue = typeof value === 'string' ? parseFloat(value) || 0 : value;
    return new Intl.NumberFormat('id-ID').format(numValue);
  };

  // Parse Rupiah string to number
  const parseRupiah = (value: string): number => {
    // Remove Rp prefix, spaces, and convert Indonesian decimal format
    const cleanValue = value
      .replace(/^Rp\s*/, '') // Remove "Rp " prefix
      .replace(/\./g, '') // Remove thousand separators (dots)
      .replace(/,/, '.'); // Convert comma to decimal point
    return parseFloat(cleanValue) || 0;
  };

  // Reset form when modal opens with new purchase
  useEffect(() => {
    if (isOpen && purchase) {
      const defaultAmount = purchase.outstanding_amount || 0;
      setFormData({
        amount: defaultAmount,
        payment_date: new Date().toISOString().split('T')[0],
        payment_method: 'Bank Transfer',
        cash_bank_id: cashBanks.length > 0 ? cashBanks[0].id : undefined,
        reference: '',
        notes: `Payment for purchase ${purchase.code}`,
      });
      setDisplayAmount(formatRupiah(defaultAmount));
    }
  }, [isOpen, purchase, cashBanks]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!purchase) return;

    // Enhanced Validation
    if (!formData.amount || formData.amount <= 0) {
      toast({
        title: 'Validation Error',
        description: 'Payment amount must be greater than zero',
        status: 'error',
        duration: 4000,
        isClosable: true,
      });
      return;
    }

    // Strict validation: prevent exceeding outstanding amount
    if (formData.amount > (purchase.outstanding_amount || 0)) {
      const maxAmount = purchase.outstanding_amount || 0;
      toast({
        title: 'Payment Amount Too High ⚠️',
        description: `Payment amount ${formatCurrency(formData.amount)} exceeds outstanding balance ${formatCurrency(maxAmount)}. Maximum allowed: ${formatCurrency(maxAmount)}`,
        status: 'error',
        duration: 6000,
        isClosable: true,
      });
      return;
    }

    // Additional validation for minimum payment amount (optional: can be removed if not needed)
    const minPayment = 1000; // Rp 1,000 minimum
    if (formData.amount < minPayment) {
      toast({
        title: 'Minimum Payment Required',
        description: `Payment amount must be at least ${formatCurrency(minPayment)}`,
        status: 'error',
        duration: 4000,
        isClosable: true,
      });
      return;
    }

    if (!formData.cash_bank_id) {
      toast({
        title: 'Validation Error',
        description: 'Please select a cash/bank account',
        status: 'error',
        duration: 4000,
        isClosable: true,
      });
      return;
    }

    // Balance validation - prevent payments when insufficient balance
    const selectedAccount = cashBanks.find(account => account.id === formData.cash_bank_id);
    if (selectedAccount) {
      if (selectedAccount.balance <= 0) {
        toast({
          title: 'Insufficient Balance ⚠️',
          description: `Cannot process payment. The selected account "${selectedAccount.name}" has zero or negative balance (${formatCurrency(selectedAccount.balance)}).`,
          status: 'error',
          duration: 8000,
          isClosable: true,
        });
        return;
      }
      
      if (formData.amount > selectedAccount.balance) {
        toast({
          title: 'Insufficient Balance ⚠️',
          description: (
            `Payment amount ${formatCurrency(formData.amount)} exceeds available balance ${formatCurrency(selectedAccount.balance)} ` +
            `in account "${selectedAccount.name}". ` +
            `Please reduce the payment amount or select a different account.`
          ),
          status: 'error',
          duration: 10000,
          isClosable: true,
        });
        return;
      }
    }

    setLoading(true);
    try {
      // Use the new Payment Management integration endpoint
      const result = await purchaseService.createPurchasePayment(purchase.id, formData);
      
      // Avoid duplicate success toasts.
      // If a parent onSuccess handler is provided, let the parent show the toast.
      if (onSuccess) {
        onSuccess(result);
      } else {
        toast({
          title: 'Payment Recorded Successfully! 🎉',
          description: 'Payment has been recorded via Payment Management and will appear in both Purchase and Payment systems',
          status: 'success',
          duration: 5000,
          isClosable: true,
        });
      }

      onClose();
    } catch (error: any) {
      console.error('Error recording payment:', error);
      
      // Check if this is a timeout error
      const isTimeoutError = error.code === 'ECONNABORTED' || error.message?.includes('timeout');
      
      // Check if this is an insufficient balance error
      const isInsufficientBalance = error.response?.data?.error_type === 'INSUFFICIENT_BALANCE' || 
                                   error.response?.data?.error?.includes('Saldo rekening tidak mencukupi') ||
                                   error.message?.includes('insufficient balance');
      
      if (isTimeoutError) {
        toast({
          title: 'Request Timeout ⏱️',
          description: 'The payment is taking longer than expected to process. This might be due to system load. Please check the payment status in a few moments or try again.',
          status: 'warning',
          duration: 10000,
          isClosable: true,
        });
      } else if (isInsufficientBalance) {
        const availableBalance = error.response?.data?.details?.match(/Available: ([\d\.]+)/);
        const requestedAmount = error.response?.data?.requested_amount || formData.amount;
        
        let errorMessage = 'Saldo rekening bank yang dipilih tidak mencukupi untuk melakukan pembayaran ini.';
        if (availableBalance) {
          errorMessage += ` Saldo tersedia: ${formatCurrency(parseFloat(availableBalance[1]))}, yang dibutuhkan: ${formatCurrency(requestedAmount)}.`;
        }
        
        toast({
          title: 'Saldo Tidak Mencukupi ⚠️',
          description: errorMessage,
          status: 'error',
          duration: 8000,
          isClosable: true,
        });
      } else {
        // Generic error handling
        let errorTitle = 'Payment Failed';
        let errorDescription = 'Failed to record payment. Please try again.';
        
        if (error.response?.data?.message) {
          errorDescription = error.response.data.message;
        } else if (error.response?.data?.error) {
          errorDescription = error.response.data.error;
        } else if (error.message) {
          errorDescription = error.message;
        }
        
        // Add specific guidance based on error type
        if (error.response?.status === 401) {
          errorTitle = 'Authentication Error';
          errorDescription = 'Your session has expired. Please login again.';
        } else if (error.response?.status === 403) {
          errorTitle = 'Permission Error';
          errorDescription = 'You do not have permission to record payments.';
        } else if (error.response?.status === 500) {
          errorTitle = 'Server Error';
          errorDescription = 'A server error occurred. Please try again later or contact support.';
        }
        
        toast({
          title: errorTitle,
          description: errorDescription,
          status: 'error',
          duration: 8000,
          isClosable: true,
        });
      }
    } finally {
      setLoading(false);
    }
  };

  // Handle amount input change with Rupiah formatting
  const handleAmountChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const inputValue = e.target.value;
    
    // Only allow numbers, dots, commas, spaces, and "Rp"
    const allowedCharsRegex = /^[Rp\d.,\s]*$/;
    if (!allowedCharsRegex.test(inputValue)) {
      return; // Ignore invalid characters
    }
    
    const numericValue = parseRupiah(inputValue);
    
    // Validate max amount
    const maxAmount = purchase?.outstanding_amount || 0;
    if (numericValue > maxAmount) {
      toast({
        title: 'Amount Exceeds Outstanding Balance',
        description: `Maximum payment amount is ${formatCurrency(maxAmount)}`,
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      // Still allow the input but show validation message
    }
    
    // Update form value
    setFormData(prev => ({ ...prev, amount: numericValue }));
    
    // Update display value
    setDisplayAmount(formatRupiah(numericValue));
  };

  const handleChange = (field: keyof PurchasePaymentRequest) => (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>
  ) => {
    let value: string | number = e.target.value;
    
    // Special handling for cash_bank_id field
    if (field === 'cash_bank_id') {
      value = parseFloat(e.target.value) || 0;
    }
    
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
  };

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

  if (!purchase) return null;

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="lg">
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Record Payment</ModalHeader>
        <ModalCloseButton />
        <form onSubmit={handleSubmit}>
          <ModalBody>
            <VStack spacing={4} align="stretch">
              {/* Purchase Information */}
              <Box p={4} bg="gray.50" borderRadius="md">
                <Text fontSize="sm" fontWeight="bold" color="gray.600" mb={2}>
                  PURCHASE INFORMATION
                </Text>
                <HStack justify="space-between">
                  <Box>
                    <Text fontSize="sm" color="gray.600">Purchase #</Text>
                    <Text fontWeight="bold">{purchase.code}</Text>
                  </Box>
                  <Box>
                    <Text fontSize="sm" color="gray.600">Vendor</Text>
                    <Text fontWeight="bold">{purchase.vendor?.name}</Text>
                  </Box>
                  <Box>
                    <Text fontSize="sm" color="gray.600">Date</Text>
                    <Text fontWeight="bold">{formatDate(purchase.date)}</Text>
                  </Box>
                </HStack>
                <Divider my={2} />
                <HStack justify="space-between">
                  <Box>
                    <Text fontSize="sm" color="gray.600">Total Amount</Text>
                    <Text fontWeight="bold">{formatCurrency(purchase.total_amount)}</Text>
                  </Box>
                  <Box>
                    <Text fontSize="sm" color="gray.600">Paid Amount</Text>
                    <Text fontWeight="bold">{formatCurrency(purchase.paid_amount || 0)}</Text>
                  </Box>
                  <Box>
                    <Text fontSize="sm" color="red.600">Outstanding</Text>
                    <Text fontWeight="bold" color="red.600">
                      {formatCurrency(purchase.outstanding_amount || 0)}
                    </Text>
                  </Box>
                </HStack>
              </Box>

              <Alert status="info">
                <AlertIcon />
                <Text fontSize="sm">
                  This payment will be recorded in both Purchase and Payment Management systems.
                  {loading && (
                    <><br />⏳ <strong>Processing payment...</strong> This may take up to 30 seconds due to complex accounting operations.</>
                  )}
                </Text>
              </Alert>

              {/* Payment Form */}
              <FormControl isRequired>
                <FormLabel>Payment Amount *</FormLabel>
                <Input
                  placeholder="Rp 0"
                  value={`Rp ${displayAmount}`}
                  onChange={handleAmountChange}
                  textAlign="right"
                  fontWeight="medium"
                  fontSize="md"
                  pl={8}
                />
                
                {/* Quick Payment Amount Buttons */}
                <HStack spacing={2} mt={3} flexWrap="wrap">
                  <Text fontSize="sm" color="gray.600" minW="fit-content">Quick Select:</Text>
                  <Button
                    size="xs"
                    variant="outline"
                    colorScheme="blue"
                    onClick={() => {
                      const amount = Math.round((purchase.outstanding_amount || 0) * 0.25);
                      setFormData(prev => ({ ...prev, amount }));
                      setDisplayAmount(formatRupiah(amount));
                    }}
                    disabled={!purchase?.outstanding_amount}
                  >
                    25%
                  </Button>
                  <Button
                    size="xs"
                    variant="outline"
                    colorScheme="blue"
                    onClick={() => {
                      const amount = Math.round((purchase.outstanding_amount || 0) * 0.5);
                      setFormData(prev => ({ ...prev, amount }));
                      setDisplayAmount(formatRupiah(amount));
                    }}
                    disabled={!purchase?.outstanding_amount}
                  >
                    50%
                  </Button>
                  <Button
                    size="xs"
                    variant="outline"
                    colorScheme="orange"
                    onClick={() => {
                      const amount = Math.round((purchase.outstanding_amount || 0) * 0.8);
                      setFormData(prev => ({ ...prev, amount }));
                      setDisplayAmount(formatRupiah(amount));
                    }}
                    disabled={!purchase?.outstanding_amount}
                  >
                    80%
                  </Button>
                  <Button
                    size="xs"
                    variant="solid"
                    colorScheme="green"
                    onClick={() => {
                      const amount = purchase.outstanding_amount || 0;
                      setFormData(prev => ({ ...prev, amount }));
                      setDisplayAmount(formatRupiah(amount));
                    }}
                    disabled={!purchase?.outstanding_amount}
                  >
                    100% Full Pay
                  </Button>
                </HStack>
                
                {/* Payment Amount Info */}
                {formData.amount > 0 && (
                  <Text fontSize="sm" color="gray.600" mt={2}>
                    💰 Payment: <Text as="span" fontWeight="bold" color="green.600">
                      {formatCurrency(formData.amount)}
                    </Text>
                    {formData.amount < (purchase.outstanding_amount || 0) && (
                      <Text as="span" color="orange.500">
                        {' • '} Remaining: {formatCurrency((purchase.outstanding_amount || 0) - formData.amount)}
                      </Text>
                    )}
                    {formData.amount === (purchase.outstanding_amount || 0) && (
                      <Text as="span" color="green.500">
                        {' • '} ✅ Full Payment
                      </Text>
                    )}
                  </Text>
                )}
                
                {/* Validation Messages */}
                {formData.amount > (purchase.outstanding_amount || 0) && (
                  <Text fontSize="sm" color="red.500" mt={1}>
                    ⚠️ Amount exceeds outstanding balance of {formatCurrency(purchase.outstanding_amount || 0)}
                  </Text>
                )}
                {formData.amount > 0 && formData.amount <= (purchase.outstanding_amount || 0) && formData.amount < (purchase.outstanding_amount || 0) && (
                  <Text fontSize="sm" color="blue.600" mt={1}>
                    ✓ Partial payment - Remaining balance: {formatCurrency((purchase.outstanding_amount || 0) - formData.amount)}
                  </Text>
                )}
                {formData.amount === (purchase.outstanding_amount || 0) && formData.amount > 0 && (
                  <Text fontSize="sm" color="green.600" mt={1} fontWeight="medium">
                    🎉 This will fully pay the purchase!
                  </Text>
                )}
              </FormControl>

              <FormControl isRequired>
                <FormLabel>Payment Date</FormLabel>
                <Input
                  type="date"
                  value={formData.payment_date}
                  onChange={handleChange('payment_date')}
                />
              </FormControl>

              <FormControl isRequired>
                <FormLabel>Payment Method</FormLabel>
                <Select
                  value={formData.payment_method}
                  onChange={handleChange('payment_method')}
                >
                  <option value="Bank Transfer">Bank Transfer</option>
                  <option value="Cash">Cash</option>
                  <option value="Check">Check</option>
                  <option value="Other">Other</option>
                </Select>
              </FormControl>

              <FormControl isRequired>
                <FormLabel>
                  {formData.payment_method === 'Cash' ? 'Cash Account' : 'Bank Account'}
                </FormLabel>
                <Select
                  value={formData.cash_bank_id || ''}
                  onChange={handleChange('cash_bank_id')}
                  placeholder={`Select ${formData.payment_method === 'Cash' ? 'cash' : 'bank'} account`}
                >
                  {cashBanks
                    .filter(account => {
                      // Filter based on payment method
                      if (formData.payment_method === 'Cash') {
                        return account.account_code?.toUpperCase().includes('CASH') || 
                               account.name?.toUpperCase().includes('CASH') ||
                               (account as any).type === 'CASH';
                      } else {
                        return account.account_code?.toUpperCase().includes('BANK') || 
                               account.name?.toUpperCase().includes('BANK') ||
                               (account as any).type === 'BANK' ||
                               (!account.account_code?.toUpperCase().includes('CASH') && 
                                !account.name?.toUpperCase().includes('CASH'));
                      }
                    })
                    .map((account) => {
                      const isInsufficientBalance = account.balance < formData.amount;
                      const isZeroBalance = account.balance <= 0;
                      const balanceStatus = isZeroBalance ? ' ❌ NO BALANCE' : 
                                           isInsufficientBalance ? ' ⚠️ INSUFFICIENT' : 
                                           ' ✅';
                      
                      return (
                        <option 
                          key={account.id} 
                          value={account.id}
                          style={{ 
                            color: isZeroBalance ? '#e53e3e' : 
                                   isInsufficientBalance ? '#dd6b20' : 
                                   '#38a169',
                            backgroundColor: isZeroBalance ? '#fed7d7' : 
                                           isInsufficientBalance ? '#feebc8' : 
                                           '#f0fff4'
                          }}
                        >
                          {account.name} ({account.account_code}) - {formatCurrency(account.balance)}{balanceStatus}
                        </option>
                      );
                    })}
                  {/* If no filtered accounts, show all accounts */}
                  {cashBanks.filter(account => {
                    if (formData.payment_method === 'Cash') {
                      return account.account_code?.toUpperCase().includes('CASH') || 
                             account.name?.toUpperCase().includes('CASH') ||
                             (account as any).type === 'CASH';
                    } else {
                      return account.account_code?.toUpperCase().includes('BANK') || 
                             account.name?.toUpperCase().includes('BANK') ||
                             (account as any).type === 'BANK' ||
                             (!account.account_code?.toUpperCase().includes('CASH') && 
                              !account.name?.toUpperCase().includes('CASH'));
                    }
                  }).length === 0 && cashBanks.map((account) => (
                    <option key={account.id} value={account.id}>
                      {account.name} ({account.account_code}) - {formatCurrency(account.balance)}
                    </option>
                  ))}
                </Select>
                
                {/* Real-time Balance Warning */}
                {formData.cash_bank_id && (() => {
                  const selectedAccount = cashBanks.find(account => account.id === formData.cash_bank_id);
                  if (!selectedAccount) return null;
                  
                  const isZeroBalance = selectedAccount.balance <= 0;
                  const isInsufficientBalance = formData.amount > selectedAccount.balance;
                  const remainingBalance = selectedAccount.balance - formData.amount;
                  
                  if (isZeroBalance) {
                    return (
                      <Text fontSize="sm" color="red.500" mt={2} fontWeight="medium">
                        ❌ <strong>No Balance Available</strong><br/>
                        Account "{selectedAccount.name}" has {formatCurrency(selectedAccount.balance)} balance. Cannot process any payments.
                      </Text>
                    );
                  }
                  
                  if (isInsufficientBalance && formData.amount > 0) {
                    return (
                      <Text fontSize="sm" color="orange.500" mt={2} fontWeight="medium">
                        ⚠️ <strong>Insufficient Balance</strong><br/>
                        Available: {formatCurrency(selectedAccount.balance)} | Required: {formatCurrency(formData.amount)} | 
                        Short by: {formatCurrency(formData.amount - selectedAccount.balance)}
                      </Text>
                    );
                  }
                  
                  if (formData.amount > 0 && remainingBalance >= 0) {
                    return (
                      <Text fontSize="sm" color="green.600" mt={2}>
                        ✅ <strong>Sufficient Balance</strong><br/>
                        Available: {formatCurrency(selectedAccount.balance)} | After payment: {formatCurrency(remainingBalance)}
                      </Text>
                    );
                  }
                  
                  return (
                    <Text fontSize="sm" color="blue.600" mt={2}>
                      💰 <strong>Account Balance:</strong> {formatCurrency(selectedAccount.balance)}
                    </Text>
                  );
                })()}
              </FormControl>

              <FormControl>
                <FormLabel>Reference</FormLabel>
                <Input
                  value={formData.reference}
                  onChange={handleChange('reference')}
                  placeholder="Transfer reference, check number, etc."
                />
              </FormControl>

              <FormControl>
                <FormLabel>Notes</FormLabel>
                <Textarea
                  value={formData.notes}
                  onChange={handleChange('notes')}
                  placeholder="Additional notes (optional)"
                  rows={3}
                />
              </FormControl>
            </VStack>
          </ModalBody>

          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onClose} disabled={loading}>
              Cancel
            </Button>
            <Button 
              colorScheme="green" 
              type="submit" 
              disabled={loading || (() => {
                // Disable if no account selected
                if (!formData.cash_bank_id) return true;
                
                // Disable if amount is invalid
                if (!formData.amount || formData.amount <= 0) return true;
                
                // Check balance availability
                const selectedAccount = cashBanks.find(account => account.id === formData.cash_bank_id);
                if (selectedAccount) {
                  // Disable if zero balance or insufficient balance
                  if (selectedAccount.balance <= 0 || formData.amount > selectedAccount.balance) {
                    return true;
                  }
                }
                
                return false;
              })()}
              leftIcon={loading ? <Spinner size="sm" /> : undefined}
              loadingText="Processing Payment..."
              isLoading={loading}
            >
              {(() => {
                if (loading) return 'Processing Payment...';
                
                // Check for balance issues
                const selectedAccount = cashBanks.find(account => account.id === formData.cash_bank_id);
                if (selectedAccount && formData.cash_bank_id) {
                  if (selectedAccount.balance <= 0) {
                    return 'No Balance Available';
                  }
                  if (formData.amount > selectedAccount.balance) {
                    return 'Insufficient Balance';
                  }
                }
                
                return 'Record Payment';
              })()} 
            </Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};

export default PurchasePaymentForm;