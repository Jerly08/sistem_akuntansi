'use client';

import React, { useState, useEffect } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  Button,
  FormControl,
  FormLabel,
  Input,
  Textarea,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  Alert,
  AlertIcon,
  useToast,
  Box,
  Text,
  VStack,
  Flex,
  Badge,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  Select,
  HStack,
  Divider,
} from '@chakra-ui/react';
import { CashBank } from '@/services/cashbankService';
import cashbankService from '@/services/cashbankService';
import { useAuth } from '@/contexts/AuthContext';

interface Account {
  id: number;
  code: string;
  name: string;
  type: string;
  category: string;
  is_active: boolean;
  balance: number;
}

interface DepositFormImprovedProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
  account: CashBank | null;
}

interface DepositRequestImproved {
  account_id: number;
  date: string;
  amount: number;
  reference: string;
  notes: string;
  source_account_id?: number; // Optional revenue account ID
}

const DepositFormImproved: React.FC<DepositFormImprovedProps> = ({
  isOpen,
  onClose,
  onSuccess,
  account
}) => {
  const { token } = useAuth();
  const [formData, setFormData] = useState({
    date: new Date().toISOString().split('T')[0],
    amount: 0,
    reference: '',
    notes: '',
    source_account_id: ''
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [revenueAccounts, setRevenueAccounts] = useState<Account[]>([]);
  const [loadingAccounts, setLoadingAccounts] = useState(false);
  const toast = useToast();

  // Load revenue accounts when form opens
  useEffect(() => {
    if (isOpen && token) {
      loadRevenueAccounts();
    }
  }, [isOpen, token]);

  const loadRevenueAccounts = async () => {
    try {
      setLoadingAccounts(true);
      const accounts = await cashbankService.getRevenueAccounts();
      setRevenueAccounts(accounts);
    } catch (error) {
      console.error('Error loading revenue accounts:', error);
      // Set default revenue accounts as fallback
      setRevenueAccounts([
        { id: 0, code: '4900', name: 'Other Income (Default)', type: 'REVENUE', category: 'OTHER_REVENUE', is_active: true, balance: 0 }
      ]);
    } finally {
      setLoadingAccounts(false);
    }
  };

  const handleInputChange = (field: string, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
    setError(null);
  };

  const handleSubmit = async () => {
    try {
      setLoading(true);
      setError(null);

      // Basic validation
      if (!account) {
        throw new Error('No account selected');
      }

      if (formData.amount <= 0) {
        throw new Error('Amount must be greater than zero');
      }

      // Prepare request data
      const requestData: any = {
        account_id: account.id,
        date: formData.date,
        amount: formData.amount,
        reference: formData.reference,
        notes: formData.notes,
      };

      // Add source account ID if selected (not default)
      if (formData.source_account_id && formData.source_account_id !== '' && parseInt(formData.source_account_id) > 0) {
        requestData.source_account_id = parseInt(formData.source_account_id);
      }

      // Use cashbankService instead of direct fetch
      await cashbankService.processDeposit(requestData);
      
      // If we get here, it was successful
      const selectedRevenue = revenueAccounts.find(acc => acc.id === parseInt(formData.source_account_id || '0'));
      const sourceAccountName = selectedRevenue ? selectedRevenue.name : 'Other Income (Default)';
      
      toast({
        title: 'Deposit Successful! 💰',
        description: (
          <Box>
            <Text fontSize="sm" fontWeight="bold">{account.currency} {formData.amount.toLocaleString('id-ID')} deposited to {account.name}</Text>
            <Text fontSize="xs" color="gray.200" mt={1}>
              ✅ Debit: {account.name} (+{formData.amount.toLocaleString('id-ID')})
            </Text>
            <Text fontSize="xs" color="gray.200">
              ✅ Credit: {sourceAccountName} (+{formData.amount.toLocaleString('id-ID')})
            </Text>
            <Text fontSize="xs" color="green.200" mt={1} fontWeight="bold">
              📊 Double-entry balanced automatically!
            </Text>
          </Box>
        ),
        status: 'success',
        duration: 5000,
        isClosable: true,
      });

      onSuccess();
      onClose();
    } catch (err: any) {
      console.error('Error processing deposit:', err);
      setError(err.message || 'Failed to process deposit');
      toast({
        title: 'Deposit Failed',
        description: err.message || 'Failed to process deposit',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setError(null);
    setFormData({
      date: new Date().toISOString().split('T')[0],
      amount: 0,
      reference: '',
      notes: '',
      source_account_id: ''
    });
    onClose();
  };

  if (!account) return null;

  const newBalance = account.balance + formData.amount;
  const selectedRevenue = revenueAccounts.find(acc => acc.id === parseInt(formData.source_account_id || '0'));
  const sourceAccountName = selectedRevenue ? selectedRevenue.name : 'Other Income (Default)';

  return (
    <Modal isOpen={isOpen} onClose={handleClose} size="xl" scrollBehavior="inside">
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>
          <Flex alignItems="center" gap={3}>
            <Text fontSize="lg">💰</Text>
            <Box>
              <Text fontSize="lg" fontWeight="bold">
                Make Deposit
              </Text>
              <Text fontSize="sm" color="gray.500" fontFamily="mono">
                {account.name} ({account.code})
              </Text>
            </Box>
            <Badge colorScheme="green" variant="solid" fontSize="xs">
              AUTOMATIC MODE
            </Badge>
          </Flex>
        </ModalHeader>
        
        <ModalBody>
          {error && (
            <Alert status="error" mb={4}>
              <AlertIcon />
              {error}
            </Alert>
          )}

          <VStack spacing={6} align="stretch">
            {/* Account Information */}
            <Box>
              <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={3}>
                💼 Account Information
              </Text>
              <Box bg="gray.50" p={4} borderRadius="md">
                <Flex justify="space-between" align="center" mb={2}>
                  <Box>
                    <Text fontSize="sm" color="gray.600" mb={1}>Account Name</Text>
                    <Text fontWeight="medium">{account.name}</Text>
                  </Box>
                  <Badge colorScheme={account.type === 'CASH' ? 'green' : 'blue'}>
                    {account.type}
                  </Badge>
                </Flex>
                
                <Stat>
                  <StatLabel>Current Balance</StatLabel>
                  <StatNumber 
                    color={account.balance < 0 ? 'red.500' : 'green.600'}
                    fontFamily="mono"
                  >
                    {account.currency} {Math.abs(account.balance).toLocaleString('id-ID')}
                    {account.balance < 0 && ' (Dr)'}
                  </StatNumber>
                  <StatHelpText>
                    {account.balance < 0 ? '⚠️ Overdraft' : 
                     account.balance > 0 ? '✅ Available' : '➖ Zero Balance'}
                  </StatHelpText>
                </Stat>
              </Box>
            </Box>

            {/* Transaction Form */}
            <Box>
              <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={3}>
                📝 Transaction Details
              </Text>
              
              <VStack spacing={4} align="stretch">
                <FormControl isRequired>
                  <FormLabel>Transaction Date</FormLabel>
                  <Input
                    type="date"
                    value={formData.date}
                    onChange={(e) => handleInputChange('date', e.target.value)}
                  />
                </FormControl>

                <FormControl isRequired>
                  <FormLabel>Amount ({account.currency})</FormLabel>
                  <NumberInput
                    value={formData.amount}
                    onChange={(_, value) => handleInputChange('amount', value || 0)}
                    min={0}
                    precision={2}
                  >
                    <NumberInputField />
                    <NumberInputStepper>
                      <NumberIncrementStepper />
                      <NumberDecrementStepper />
                    </NumberInputStepper>
                  </NumberInput>
                </FormControl>

                <FormControl>
                  <FormLabel>Reference</FormLabel>
                  <Input
                    value={formData.reference}
                    onChange={(e) => handleInputChange('reference', e.target.value)}
                    placeholder="e.g., Receipt #123, Bank slip, etc."
                  />
                </FormControl>

                <FormControl>
                  <FormLabel>Notes</FormLabel>
                  <Textarea
                    value={formData.notes}
                    onChange={(e) => handleInputChange('notes', e.target.value)}
                    placeholder="Optional transaction notes"
                    rows={3}
                  />
                </FormControl>
              </VStack>
            </Box>

            {/* Revenue Account Selection */}
            <Box>
              <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={3}>
                📊 Revenue Account Selection
              </Text>
              <Text fontSize="sm" color="gray.600" mb={3}>
                Select the revenue account to be credited. Leave blank to use default "Other Income" account.
              </Text>
              
              <FormControl>
                <FormLabel>Credit Account (Revenue Source)</FormLabel>
                <Select
                  value={formData.source_account_id}
                  onChange={(e) => handleInputChange('source_account_id', e.target.value)}
                  placeholder="Use default 'Other Income' account"
                  isDisabled={loadingAccounts}
                >
                  {revenueAccounts.map((account) => (
                    <option key={account.id} value={account.id}>
                      {account.code} - {account.name}
                      {account.balance > 0 && ` (Balance: ${account.balance.toLocaleString('id-ID')})`}
                    </option>
                  ))}
                </Select>
                {loadingAccounts && (
                  <Text fontSize="xs" color="gray.500" mt={1}>
                    Loading revenue accounts...
                  </Text>
                )}
              </FormControl>
            </Box>

            {/* Double-Entry Preview */}
            {formData.amount > 0 && (
              <Box>
                <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={3}>
                  📈 Double-Entry Preview
                </Text>
                <Box bg="blue.50" p={4} borderRadius="md" border="1px solid" borderColor="blue.200">
                  <VStack spacing={3} align="stretch">
                    <HStack justify="space-between">
                      <Text fontSize="sm" fontWeight="bold" color="blue.800">
                        Journal Entry (Auto-Generated):
                      </Text>
                      <Badge colorScheme="blue" variant="solid" fontSize="xs">
                        BALANCED
                      </Badge>
                    </HStack>
                    
                    <Divider />
                    
                    {/* Debit Entry */}
                    <HStack justify="space-between" align="center">
                      <HStack spacing={2}>
                        <Badge colorScheme="green" variant="outline" fontSize="xs">DR</Badge>
                        <Box>
                          <Text fontSize="sm" fontWeight="medium">{account.name}</Text>
                          <Text fontSize="xs" color="gray.600" fontFamily="mono">{account.code}</Text>
                        </Box>
                      </HStack>
                      <Text fontSize="sm" fontWeight="bold" fontFamily="mono" color="green.600">
                        +{formData.amount.toLocaleString('id-ID')}
                      </Text>
                    </HStack>
                    
                    {/* Credit Entry */}
                    <HStack justify="space-between" align="center">
                      <HStack spacing={2}>
                        <Badge colorScheme="orange" variant="outline" fontSize="xs">CR</Badge>
                        <Box>
                          <Text fontSize="sm" fontWeight="medium">{sourceAccountName}</Text>
                          <Text fontSize="xs" color="gray.600" fontFamily="mono">
                            {selectedRevenue?.code || '4900'}
                          </Text>
                        </Box>
                      </HStack>
                      <Text fontSize="sm" fontWeight="bold" fontFamily="mono" color="orange.600">
                        +{formData.amount.toLocaleString('id-ID')}
                      </Text>
                    </HStack>
                    
                    <Divider />
                    
                    <HStack justify="space-between">
                      <Text fontSize="sm" fontWeight="bold">Total Balance:</Text>
                      <Text fontSize="sm" fontWeight="bold" color="green.600">
                        DR {formData.amount.toLocaleString('id-ID')} = CR {formData.amount.toLocaleString('id-ID')} ✅
                      </Text>
                    </HStack>
                  </VStack>
                </Box>
              </Box>
            )}

            {/* Balance Preview */}
            {formData.amount > 0 && (
              <Box>
                <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={3}>
                  💡 Balance Preview
                </Text>
                <Alert status={newBalance < 0 ? 'warning' : 'success'} borderRadius="md">
                  <AlertIcon />
                  <Box>
                    <Text fontSize="sm" fontWeight="medium">
                      New balance after deposit:
                    </Text>
                    <Text fontSize="lg" fontWeight="bold" fontFamily="mono" mt={1}>
                      {account.currency} {Math.abs(newBalance).toLocaleString('id-ID')}
                      {newBalance < 0 && ' (Dr)'}
                    </Text>
                    {newBalance < 0 && (
                      <Text fontSize="xs" color="orange.600" mt={1}>
                        ⚠️ This will result in an overdraft
                      </Text>
                    )}
                  </Box>
                </Alert>
              </Box>
            )}
          </VStack>
        </ModalBody>

        <ModalFooter>
          <Button variant="ghost" mr={3} onClick={handleClose} isDisabled={loading}>
            Cancel
          </Button>
          <Button
            colorScheme="green"
            onClick={handleSubmit}
            isLoading={loading}
            loadingText="Processing deposit..."
            isDisabled={formData.amount <= 0}
          >
            Process Deposit
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default DepositFormImproved;
