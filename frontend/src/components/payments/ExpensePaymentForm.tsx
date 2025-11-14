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
  VStack,
  HStack,
  FormControl,
  FormLabel,
  FormErrorMessage,
  FormHelperText,
  Input,
  Select,
  Textarea,
  NumberInput,
  NumberInputField,
  Box,
  Text,
  Divider,
  useToast,
  Alert,
  AlertIcon,
  AlertDescription,
  useColorModeValue,
  Tooltip,
  Icon,
  Spinner,
  Badge,
} from '@chakra-ui/react';
import { useForm, Controller } from 'react-hook-form';
import { 
  FiCreditCard, 
  FiDollarSign, 
  FiCalendar, 
  FiFileText, 
  FiInfo,
  FiBookOpen,
  FiBriefcase
} from 'react-icons/fi';
import paymentService from '@/services/paymentService';
import accountService from '@/services/accountService';
import cashbankService, { CashBank } from '@/services/cashbankService';
import { AccountCatalogItem } from '@/types/account';
import CurrencyInput from '@/components/common/CurrencyInput';
import { useAuth } from '@/contexts/AuthContext';

interface ExpensePaymentFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess?: () => void;
}

interface ExpensePaymentFormData {
  expense_account_id: number;
  cash_bank_id: number;
  date: string;
  amount: number;
  method: string;
  reference: string;
  notes: string;
  description: string;
}

const PAYMENT_METHODS = [
  { value: 'CASH', label: 'Cash', icon: FiDollarSign },
  { value: 'BANK_TRANSFER', label: 'Bank Transfer', icon: FiCreditCard },
  { value: 'CHECK', label: 'Check', icon: FiFileText },
  { value: 'CREDIT_CARD', label: 'Credit Card', icon: FiCreditCard },
  { value: 'DEBIT_CARD', label: 'Debit Card', icon: FiCreditCard },
  { value: 'OTHER', label: 'Other', icon: FiFileText },
];

const ExpensePaymentForm: React.FC<ExpensePaymentFormProps> = ({
  isOpen,
  onClose,
  onSuccess,
}) => {
  const { token } = useAuth();
  const toast = useToast();
  
  // Color mode values
  const bgColor = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const textColor = useColorModeValue('gray.600', 'gray.300');
  const labelColor = useColorModeValue('gray.700', 'gray.200');
  
  const [loading, setLoading] = useState(false);
  const [expenseAccounts, setExpenseAccounts] = useState<AccountCatalogItem[]>([]);
  const [liabilityAccounts, setLiabilityAccounts] = useState<AccountCatalogItem[]>([]);
  const [cashBanks, setCashBanks] = useState<CashBank[]>([]);
  const [loadingAccounts, setLoadingAccounts] = useState(true);

  const {
    control,
    register,
    handleSubmit,
    formState: { errors },
    reset,
    watch,
  } = useForm<ExpensePaymentFormData>({
    defaultValues: {
      date: new Date().toISOString().split('T')[0],
      amount: 0,
      method: 'CASH',
      reference: '',
      notes: '',
      description: '',
    },
  });

  const selectedAccountId = watch('expense_account_id');
  const selectedAccount = [...expenseAccounts, ...liabilityAccounts].find(
    acc => acc.id === Number(selectedAccountId)
  );

  // Load accounts and cash banks
  useEffect(() => {
    if (isOpen && token) {
      loadData();
    }
  }, [isOpen, token]);

  const loadData = async () => {
    setLoadingAccounts(true);
    try {
      // Load expense accounts
      const expenseData = await accountService.getAccountCatalog(token, 'EXPENSE');
      setExpenseAccounts(expenseData.filter(acc => acc.active));

      // Load liability accounts
      const liabilityData = await accountService.getAccountCatalog(token, 'LIABILITY');
      setLiabilityAccounts(liabilityData.filter(acc => acc.active));

      // Load cash/bank accounts that can be used for payments
      const cashBankData = await cashbankService.getPaymentAccounts();
      setCashBanks(cashBankData || []);
    } catch (error) {
      console.error('Error loading data:', error);
      toast({
        title: 'Error loading accounts',
        description: 'Failed to load account data. Please try again.',
        status: 'error',
        duration: 3000,
      });
    } finally {
      setLoadingAccounts(false);
    }
  };

  const onSubmit = async (data: ExpensePaymentFormData) => {
    setLoading(true);
    try {
      // Create expense payment request
      const paymentRequest = {
        expense_account_id: Number(data.expense_account_id),
        cash_bank_id: Number(data.cash_bank_id),
        date: data.date,
        amount: data.amount,
        method: data.method,
        reference: data.reference || '',
        notes: data.notes || '',
        description: data.description || `Payment for ${selectedAccount?.name || 'Expense'}`,
        auto_create_journal: true,
      };

      // Call the expense payment endpoint
      await paymentService.createExpensePayment(paymentRequest);

      toast({
        title: 'Success',
        description: 'Expense payment has been created successfully',
        status: 'success',
        duration: 3000,
      });

      handleClose();
      if (onSuccess) onSuccess();
    } catch (error: any) {
      console.error('Error creating expense payment:', error);
      toast({
        title: 'Error creating payment',
        description: error.message || 'Failed to create expense payment',
        status: 'error',
        duration: 5000,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} size="xl" closeOnOverlayClick={false}>
      <ModalOverlay />
      <ModalContent bg={bgColor}>
        <ModalHeader>
          <HStack>
            <Icon as={FiBriefcase} />
            <Text>Create Expense Payment</Text>
          </HStack>
        </ModalHeader>
        <ModalCloseButton />
        
        <form onSubmit={handleSubmit(onSubmit)}>
          <ModalBody>
            <VStack spacing={4}>
              {/* Info Alert */}
              <Alert status="info" borderRadius="md">
                <AlertIcon />
                <AlertDescription fontSize="sm">
                  Use this form to record direct expense payments or liability settlements from your cash/bank accounts
                </AlertDescription>
              </Alert>

              {loadingAccounts ? (
                <Box py={8} textAlign="center" width="full">
                  <Spinner size="lg" />
                  <Text mt={2} color={textColor}>Loading accounts...</Text>
                </Box>
              ) : (
                <>
                  {/* Expense/Liability Account Selection */}
                  <FormControl isInvalid={!!errors.expense_account_id} isRequired>
                    <FormLabel color={labelColor}>
                      <HStack spacing={1}>
                        <Icon as={FiBookOpen} />
                        <Text>Expense/Liability Account</Text>
                        <Tooltip label="Select the expense or liability account from COA">
                          <Icon as={FiInfo} boxSize={3} color={textColor} />
                        </Tooltip>
                      </HStack>
                    </FormLabel>
                    <Controller
                      name="expense_account_id"
                      control={control}
                      rules={{ required: 'Please select an account' }}
                      render={({ field }) => (
                        <Select
                          {...field}
                          placeholder="Select account from COA"
                          borderColor={borderColor}
                        >
                          {expenseAccounts.length > 0 && (
                            <>
                              <option disabled style={{ fontWeight: 'bold', color: '#666' }}>
                                ── Expense Accounts ──
                              </option>
                              {expenseAccounts.map(account => (
                                <option key={account.id} value={account.id}>
                                  {account.code} - {account.name}
                                </option>
                              ))}
                            </>
                          )}
                          {liabilityAccounts.length > 0 && (
                            <>
                              <option disabled style={{ fontWeight: 'bold', color: '#666' }}>
                                ── Liability Accounts ──
                              </option>
                              {liabilityAccounts.map(account => (
                                <option key={account.id} value={account.id}>
                                  {account.code} - {account.name}
                                </option>
                              ))}
                            </>
                          )}
                        </Select>
                      )}
                    />
                    <FormErrorMessage>{errors.expense_account_id?.message}</FormErrorMessage>
                  </FormControl>

                  {/* Cash/Bank Account */}
                  <FormControl isInvalid={!!errors.cash_bank_id} isRequired>
                    <FormLabel color={labelColor}>
                      <HStack spacing={1}>
                        <Icon as={FiCreditCard} />
                        <Text>Payment From (Cash/Bank)</Text>
                        <Tooltip label="Select the cash or bank account to pay from">
                          <Icon as={FiInfo} boxSize={3} color={textColor} />
                        </Tooltip>
                      </HStack>
                    </FormLabel>
                    <Controller
                      name="cash_bank_id"
                      control={control}
                      rules={{ required: 'Please select a cash/bank account' }}
                      render={({ field }) => (
                        <Select {...field} placeholder="Select cash/bank account" borderColor={borderColor}>
                          {cashBanks.map(cb => (
                            <option key={cb.id} value={cb.id}>
                              {cb.name} ({cb.type}) - Balance: {new Intl.NumberFormat('id-ID', {
                                style: 'currency',
                                currency: 'IDR',
                                minimumFractionDigits: 0,
                              }).format(cb.balance || 0)}
                            </option>
                          ))}
                        </Select>
                      )}
                    />
                    <FormErrorMessage>{errors.cash_bank_id?.message}</FormErrorMessage>
                  </FormControl>

                  <HStack width="full" spacing={4}>
                    {/* Payment Date */}
                    <FormControl isInvalid={!!errors.date} isRequired flex={1}>
                      <FormLabel color={labelColor}>
                        <HStack spacing={1}>
                          <Icon as={FiCalendar} />
                          <Text>Payment Date</Text>
                        </HStack>
                      </FormLabel>
                      <Input
                        type="date"
                        {...register('date', { required: 'Payment date is required' })}
                        borderColor={borderColor}
                      />
                      <FormErrorMessage>{errors.date?.message}</FormErrorMessage>
                    </FormControl>

                    {/* Payment Method */}
                    <FormControl isInvalid={!!errors.method} isRequired flex={1}>
                      <FormLabel color={labelColor}>Payment Method</FormLabel>
                      <Controller
                        name="method"
                        control={control}
                        rules={{ required: 'Payment method is required' }}
                        render={({ field }) => (
                          <Select {...field} borderColor={borderColor}>
                            {PAYMENT_METHODS.map(method => (
                              <option key={method.value} value={method.value}>
                                {method.label}
                              </option>
                            ))}
                          </Select>
                        )}
                      />
                      <FormErrorMessage>{errors.method?.message}</FormErrorMessage>
                    </FormControl>
                  </HStack>

                  {/* Amount */}
                  <FormControl isInvalid={!!errors.amount} isRequired>
                    <FormLabel color={labelColor}>
                      <HStack spacing={1}>
                        <Icon as={FiDollarSign} />
                        <Text>Amount</Text>
                      </HStack>
                    </FormLabel>
                    <Controller
                      name="amount"
                      control={control}
                      rules={{
                        required: 'Amount is required',
                        min: { value: 0.01, message: 'Amount must be greater than 0' },
                      }}
                      render={({ field }) => (
                        <CurrencyInput
                          value={field.value}
                          onChange={field.onChange}
                          placeholder="Enter payment amount"
                        />
                      )}
                    />
                    <FormErrorMessage>{errors.amount?.message}</FormErrorMessage>
                  </FormControl>

                  {/* Reference Number */}
                  <FormControl>
                    <FormLabel color={labelColor}>
                      <HStack spacing={1}>
                        <Icon as={FiFileText} />
                        <Text>Reference Number</Text>
                      </HStack>
                    </FormLabel>
                    <Input
                      {...register('reference')}
                      placeholder="e.g., Invoice number, receipt number"
                      borderColor={borderColor}
                    />
                    <FormHelperText color={textColor}>
                      Optional: External reference number for this payment
                    </FormHelperText>
                  </FormControl>

                  {/* Description */}
                  <FormControl>
                    <FormLabel color={labelColor}>Description</FormLabel>
                    <Input
                      {...register('description')}
                      placeholder="Brief description of the expense"
                      borderColor={borderColor}
                    />
                  </FormControl>

                  {/* Notes */}
                  <FormControl>
                    <FormLabel color={labelColor}>Notes</FormLabel>
                    <Textarea
                      {...register('notes')}
                      placeholder="Additional notes or details"
                      borderColor={borderColor}
                      rows={3}
                    />
                  </FormControl>
                </>
              )}

              {/* Display selected account info */}
              {selectedAccount && (
                <Box
                  width="full"
                  p={3}
                  borderRadius="md"
                  bg={useColorModeValue('gray.50', 'gray.700')}
                  borderWidth={1}
                  borderColor={borderColor}
                >
                  <HStack justify="space-between">
                    <Text fontSize="sm" color={textColor}>
                      Selected Account:
                    </Text>
                    <Badge colorScheme="blue">
                      {selectedAccount.code} - {selectedAccount.name}
                    </Badge>
                  </HStack>
                </Box>
              )}
            </VStack>
          </ModalBody>

          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={handleClose} isDisabled={loading}>
              Cancel
            </Button>
            <Button
              type="submit"
              colorScheme="blue"
              isLoading={loading}
              isDisabled={loadingAccounts}
              leftIcon={<FiDollarSign />}
            >
              Create Payment
            </Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};

export default ExpensePaymentForm;