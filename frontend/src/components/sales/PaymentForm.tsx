'use client';

import React, { useState, useEffect } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  Button,
  FormControl,
  FormLabel,
  FormErrorMessage,
  Input,
  Select,
  Textarea,
  VStack,
  HStack,
  Text,
  Box,
  Divider,
  useToast,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper
} from '@chakra-ui/react';
import { useForm } from 'react-hook-form';
import { useAuth } from '@/contexts/AuthContext';
import salesService, { Sale, SalePaymentRequest } from '@/services/salesService';
import cashbankService from '@/services/cashbankService';

interface PaymentFormProps {
  isOpen: boolean;
  onClose: () => void;
  sale: Sale | null;
  onSave: () => void;
}

interface PaymentFormData {
  date: string;
  amount: number;
  method: string;
  reference: string;
  account_id: number;
  cash_bank_id?: number;
  notes: string;
}

const PaymentForm: React.FC<PaymentFormProps> = ({
  isOpen,
  onClose,
  sale,
  onSave
}) => {
  const { token, user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [accounts, setAccounts] = useState<any[]>([]);
  const [paymentHistory, setPaymentHistory] = useState<any[]>([]);
  const [accountsLoading, setAccountsLoading] = useState(false);
  const toast = useToast();

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors }
  } = useForm<PaymentFormData>();

  const watchAmount = watch('amount');

  useEffect(() => {
    if (sale && isOpen) {
      // Set default values
      setValue('date', new Date().toISOString().split('T')[0]);
      setValue('amount', sale.outstanding_amount || 0);
      setValue('method', 'BANK_TRANSFER');
      setValue('account_id', 0);
      setValue('reference', '');
      setValue('notes', '');
      
      // Load accounts for payment
      loadAccounts();
    }
  }, [sale, isOpen, setValue]);

  const loadAccounts = async () => {
    // Check if user is authenticated and has required permissions
    if (!token || !user) {
      toast({
        title: 'Authentication Required',
        description: 'Please log in to access payment accounts.',
        status: 'error',
        duration: 5000
      });
      return;
    }

    // Check if user has permission to view accounts (based on RBAC)
    const allowedRoles = ['ADMIN', 'FINANCE', 'DIRECTOR', 'EMPLOYEE'];
    if (!allowedRoles.includes(user.role)) {
      toast({
        title: 'Access Denied',
        description: 'You do not have permission to view payment accounts.',
        status: 'error',
        duration: 5000
      });
      return;
    }

    try {
      setAccountsLoading(true);
      
      // Use cashbank service to get payment accounts (cash and bank accounts)
      const paymentAccounts = await cashbankService.getPaymentAccounts();
      setAccounts(paymentAccounts || []);

      if (paymentAccounts.length === 0) {
        toast({
          title: 'No Payment Accounts',
          description: 'No cash or bank accounts available for payments. Please contact your administrator.',
          status: 'warning',
          duration: 5000
        });
      }
      
    } catch (error: any) {
      console.error('Error loading payment accounts:', error);
      
      // Set empty accounts if service fails
      setAccounts([]);
      
      // Provide more specific error messages based on the error
      let errorMessage = 'Could not load payment accounts. Please contact your administrator.';
      
      if (error.message?.includes('403') || error.message?.includes('Forbidden')) {
        errorMessage = 'You do not have permission to view payment accounts.';
      } else if (error.message?.includes('401') || error.message?.includes('Unauthorized')) {
        errorMessage = 'Your session has expired. Please log in again.';
      } else if (error.message?.includes('Network')) {
        errorMessage = 'Network error. Please check your connection and try again.';
      }
      
      toast({
        title: 'Error Loading Payment Accounts',
        description: errorMessage,
        status: 'error',
        duration: 5000
      });
    } finally {
      setAccountsLoading(false);
    }
  };

  const onSubmit = async (data: PaymentFormData) => {
    if (!sale) return;

    try {
      setLoading(true);

      // Validate required fields
      if (!data.account_id || data.account_id === 0) {
        toast({
          title: 'Validation Error',
          description: 'Please select a payment account',
          status: 'error',
          duration: 3000
        });
        return;
      }

      // Convert date to proper ISO datetime format for backend
      const paymentDateTime = new Date(data.date).toISOString();
      
      const paymentData: SalePaymentRequest = {
        payment_date: paymentDateTime, // Send full datetime in ISO format
        amount: data.amount,
        payment_method: data.method, // Use correct field name
        reference: data.reference || '', // Ensure it's not undefined
        cash_bank_id: data.account_id, // Use the selected account ID as cash_bank_id
        notes: data.notes || '' // Ensure it's not undefined
      };

      await salesService.createSalePayment(sale.id, paymentData);

      toast({
        title: 'Payment Recorded',
        description: 'Payment has been recorded successfully',
        status: 'success',
        duration: 3000
      });

      reset();
      onSave();
      onClose();
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.message || 'Failed to record payment',
        status: 'error',
        duration: 5000
      });
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  const paymentMethods = [
    { value: 'CASH', label: 'Cash' },
    { value: 'BANK_TRANSFER', label: 'Bank Transfer' },
    { value: 'CHECK', label: 'Check' },
    { value: 'CREDIT_CARD', label: 'Credit Card' },
    { value: 'DEBIT_CARD', label: 'Debit Card' },
    { value: 'OTHER', label: 'Other' }
  ];

  return (
    <Modal isOpen={isOpen} onClose={handleClose} size="lg">
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Record Payment</ModalHeader>
        <ModalCloseButton />
        
        <form onSubmit={handleSubmit(onSubmit)}>
          <ModalBody>
            {sale && (
              <Box mb={6} p={4} bg="gray.50" borderRadius="md">
                <VStack align="stretch" spacing={2}>
                  <HStack justify="space-between">
                    <Text fontSize="sm" color="gray.600">Sale Code:</Text>
                    <Text fontWeight="medium">{sale.code}</Text>
                  </HStack>
                  <HStack justify="space-between">
                    <Text fontSize="sm" color="gray.600">Invoice Number:</Text>
                    <Text fontWeight="medium">{sale.invoice_number || 'N/A'}</Text>
                  </HStack>
                  <HStack justify="space-between">
                    <Text fontSize="sm" color="gray.600">Customer:</Text>
                    <Text fontWeight="medium">{sale.customer?.name || 'N/A'}</Text>
                  </HStack>
                  <Divider />
                  <HStack justify="space-between">
                    <Text fontSize="sm" color="gray.600">Total Amount:</Text>
                    <Text fontWeight="bold">
                      {salesService.formatCurrency(sale.total_amount)}
                    </Text>
                  </HStack>
                  <HStack justify="space-between">
                    <Text fontSize="sm" color="gray.600">Paid Amount:</Text>
                    <Text>{salesService.formatCurrency(sale.paid_amount)}</Text>
                  </HStack>
                  <HStack justify="space-between">
                    <Text fontSize="sm" color="gray.600">Outstanding:</Text>
                    <Text fontWeight="bold" color="orange.600">
                      {salesService.formatCurrency(sale.outstanding_amount)}
                    </Text>
                  </HStack>
                </VStack>
              </Box>
            )}

            {/* Payment History Section */}
            {sale?.sale_payments && sale.sale_payments.length > 0 && (
              <Box mb={4} p={4} bg="blue.50" borderRadius="md" borderLeft="4px" borderColor="blue.400">
                <Text fontSize="sm" fontWeight="bold" color="blue.700" mb={2}>
                  📋 Previous Payments
                </Text>
                <VStack spacing={2} align="stretch">
                  {sale.sale_payments.slice(-3).map((payment, index) => (
                    <HStack key={index} justify="space-between" fontSize="sm">
                      <Text color="gray.600">
                        {salesService.formatDate(payment.date)} • {payment.method}
                        {payment.reference && ` • Ref: ${payment.reference}`}
                      </Text>
                      <Text fontWeight="medium" color="green.600">
                        +{salesService.formatCurrency(payment.amount)}
                      </Text>
                    </HStack>
                  ))}
                  {sale.sale_payments.length > 3 && (
                    <Text fontSize="xs" color="gray.500" textAlign="center">
                      ... and {sale.sale_payments.length - 3} more payments
                    </Text>
                  )}
                </VStack>
              </Box>
            )}

            <VStack spacing={4} align="stretch">
              <HStack spacing={4}>
                <FormControl isRequired isInvalid={!!errors.date}>
                  <FormLabel>Payment Date *</FormLabel>
                  <Input
                    type="date"
                    max={new Date().toISOString().split('T')[0]} // Prevent future dates
                    {...register('date', {
                      required: 'Payment date is required',
                      validate: {
                        notFuture: (value) => {
                          const today = new Date();
                          const inputDate = new Date(value);
                          return inputDate <= today || 'Payment date cannot be in the future';
                        }
                      }
                    })}
                  />
                  <FormErrorMessage>{errors.date?.message}</FormErrorMessage>
                </FormControl>

                <FormControl isRequired isInvalid={!!errors.amount}>
                  <FormLabel>Amount *</FormLabel>
                  <NumberInput 
                    min={0.01} 
                    max={sale?.outstanding_amount}
                    precision={2}
                    step={0.01}
                  >
                    <NumberInputField
                      placeholder="0.00"
                      {...register('amount', {
                        required: 'Amount is required',
                        min: { value: 0.01, message: 'Amount must be greater than 0' },
                        max: {
                          value: sale?.outstanding_amount || 0,
                          message: 'Amount cannot exceed outstanding amount'
                        },
                        validate: {
                          notZero: (value) => value > 0 || 'Amount must be greater than zero',
                          hasDecimals: (value) => {
                            const decimals = value.toString().split('.')[1];
                            return !decimals || decimals.length <= 2 || 'Maximum 2 decimal places allowed';
                          }
                        }
                      })}
                    />
                    <NumberInputStepper>
                      <NumberIncrementStepper />
                      <NumberDecrementStepper />
                    </NumberInputStepper>
                  </NumberInput>
                  
                  {/* Quick Amount Selection Buttons */}
                  <HStack spacing={2} mt={2}>
                    <Button
                      size="xs"
                      variant="outline"
                      onClick={() => setValue('amount', (sale?.outstanding_amount || 0) * 0.25)}
                      disabled={!sale?.outstanding_amount}
                    >
                      25%
                    </Button>
                    <Button
                      size="xs"
                      variant="outline"
                      onClick={() => setValue('amount', (sale?.outstanding_amount || 0) * 0.5)}
                      disabled={!sale?.outstanding_amount}
                    >
                      50%
                    </Button>
                    <Button
                      size="xs"
                      variant="outline"
                      onClick={() => setValue('amount', (sale?.outstanding_amount || 0) * 0.75)}
                      disabled={!sale?.outstanding_amount}
                    >
                      75%
                    </Button>
                    <Button
                      size="xs"
                      variant="solid"
                      colorScheme="blue"
                      onClick={() => setValue('amount', sale?.outstanding_amount || 0)}
                      disabled={!sale?.outstanding_amount}
                    >
                      Full Payment
                    </Button>
                  </HStack>
                  
                  <FormErrorMessage>{errors.amount?.message}</FormErrorMessage>
                  {watchAmount > (sale?.outstanding_amount || 0) && (
                    <Text fontSize="sm" color="red.500" mt={1}>
                      ⚠️ Amount exceeds outstanding balance of {salesService.formatCurrency(sale?.outstanding_amount || 0)}
                    </Text>
                  )}
                  {watchAmount > 0 && watchAmount <= (sale?.outstanding_amount || 0) && (
                    <Text fontSize="sm" color="green.600" mt={1}>
                      ✓ Remaining balance: {salesService.formatCurrency((sale?.outstanding_amount || 0) - watchAmount)}
                    </Text>
                  )}
                  {watchAmount === (sale?.outstanding_amount || 0) && watchAmount > 0 && (
                    <Text fontSize="sm" color="blue.600" mt={1} fontWeight="medium">
                      🎉 This will fully pay the invoice!
                    </Text>
                  )}
                </FormControl>
              </HStack>

              <HStack spacing={4}>
                <FormControl isRequired isInvalid={!!errors.method}>
                  <FormLabel>Payment Method</FormLabel>
                  <Select
                    {...register('method', {
                      required: 'Payment method is required'
                    })}
                  >
                    <option value="">Select payment method</option>
                    {paymentMethods.map(method => (
                      <option key={method.value} value={method.value}>
                        {method.label}
                      </option>
                    ))}
                  </Select>
                  <FormErrorMessage>{errors.method?.message}</FormErrorMessage>
                </FormControl>

                <FormControl isRequired isInvalid={!!errors.account_id}>
                  <FormLabel>Account *</FormLabel>
                  <Select
                    {...register('account_id', {
                      required: 'Account is required',
                      setValueAs: value => parseInt(value) || 0
                    })}
                    disabled={accountsLoading || accounts.length === 0}
                  >
                    {accountsLoading ? (
                      <option value="">Loading accounts...</option>
                    ) : accounts.length === 0 ? (
                      <option value="">No accounts available</option>
                    ) : (
                      <>
                        <option value="">Select payment account</option>
                        {accounts.map(account => (
                          <option key={account.id} value={account.id}>
                            {account.type === 'BANK' && account.bank_name 
                              ? `${account.code} - ${account.name} (${account.bank_name} - ${account.account_no})`
                              : `${account.code} - ${account.name} (${account.type})`
                            }
                          </option>
                        ))}
                      </>
                    )}
                  </Select>
                  <FormErrorMessage>{errors.account_id?.message}</FormErrorMessage>
                  {accounts.length === 0 && !accountsLoading && (
                    <Text fontSize="xs" color="orange.500" mt={1}>
                      ⚠️ No payment accounts loaded. Contact your administrator if this persists.
                    </Text>
                  )}
                </FormControl>
              </HStack>

              <FormControl>
                <FormLabel>Reference Number</FormLabel>
                <Input
                  placeholder="Transaction reference number"
                  {...register('reference')}
                />
              </FormControl>

              <FormControl>
                <FormLabel>Notes</FormLabel>
                <Textarea
                  placeholder="Additional notes about this payment"
                  {...register('notes')}
                />
              </FormControl>
            </VStack>
          </ModalBody>

          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={handleClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              colorScheme="blue"
              isLoading={loading}
              loadingText="Recording Payment..."
            >
              Record Payment
            </Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};

export default PaymentForm;
