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
import salesService, { Sale, SalePaymentRequest } from '@/services/salesService';

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
  const [loading, setLoading] = useState(false);
  const [accounts, setAccounts] = useState<any[]>([]);
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
    try {
      // This would typically load cash/bank accounts
      // For now, using mock data
      setAccounts([
        { id: 1, name: 'Cash', code: '1000' },
        { id: 2, name: 'Bank - BCA', code: '1001' },
        { id: 3, name: 'Bank - BNI', code: '1002' }
      ]);
    } catch (error) {
      console.error('Error loading accounts:', error);
    }
  };

  const onSubmit = async (data: PaymentFormData) => {
    if (!sale) return;

    try {
      setLoading(true);

      const paymentData: SalePaymentRequest = {
        date: data.date,
        amount: data.amount,
        method: data.method,
        reference: data.reference,
        account_id: data.account_id,
        cash_bank_id: data.cash_bank_id,
        notes: data.notes
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

            <VStack spacing={4} align="stretch">
              <HStack spacing={4}>
                <FormControl isRequired isInvalid={!!errors.date}>
                  <FormLabel>Payment Date</FormLabel>
                  <Input
                    type="date"
                    {...register('date', {
                      required: 'Payment date is required'
                    })}
                  />
                  <FormErrorMessage>{errors.date?.message}</FormErrorMessage>
                </FormControl>

                <FormControl isRequired isInvalid={!!errors.amount}>
                  <FormLabel>Amount</FormLabel>
                  <NumberInput min={0} max={sale?.outstanding_amount}>
                    <NumberInputField
                      {...register('amount', {
                        required: 'Amount is required',
                        min: { value: 0.01, message: 'Amount must be greater than 0' },
                        max: {
                          value: sale?.outstanding_amount || 0,
                          message: 'Amount cannot exceed outstanding amount'
                        }
                      })}
                    />
                    <NumberInputStepper>
                      <NumberIncrementStepper />
                      <NumberDecrementStepper />
                    </NumberInputStepper>
                  </NumberInput>
                  <FormErrorMessage>{errors.amount?.message}</FormErrorMessage>
                  {watchAmount > (sale?.outstanding_amount || 0) && (
                    <Text fontSize="sm" color="red.500" mt={1}>
                      Amount exceeds outstanding balance
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
                  <FormLabel>Account</FormLabel>
                  <Select
                    {...register('account_id', {
                      required: 'Account is required',
                      setValueAs: value => parseInt(value) || 0
                    })}
                  >
                    <option value="">Select account</option>
                    {accounts.map(account => (
                      <option key={account.id} value={account.id}>
                        {account.code} - {account.name}
                      </option>
                    ))}
                  </Select>
                  <FormErrorMessage>{errors.account_id?.message}</FormErrorMessage>
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
