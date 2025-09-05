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
  RadioGroup,
  Radio,
  HStack,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  Alert,
  AlertIcon,
  AlertDescription,
  AlertTitle,
  useToast,
  Box,
  Checkbox,
  Text,
  VStack,
  Divider,
  Card,
  CardBody,
  SimpleGrid,
  Tab,
  Tabs,
  TabList,
  TabPanels,
  TabPanel,
  Badge,
  useColorModeValue,
} from '@chakra-ui/react';
import { useForm, Controller } from 'react-hook-form';
import { FiUser, FiDollarSign, FiCalendar, FiCreditCard, FiFileText, FiSettings } from 'react-icons/fi';
import { CashBank, CashBankCreateRequest, CashBankUpdateRequest } from '@/services/cashbankService';
import cashbankService from '@/services/cashbankService';
import accountService from '@/services/accountService';
import { useAuth } from '@/contexts/AuthContext';
import { Account } from '@/types/account';

interface CashBankFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
  account?: CashBank | null;
  mode: 'create' | 'edit';
}

interface CashBankFormData {
  name: string;
  type: 'CASH' | 'BANK';
  bank_name?: string;
  account_no?: string;
  currency: string;
  opening_balance?: number;
  opening_date?: string;
  description?: string;
  account_id?: number;
}

const CashBankForm: React.FC<CashBankFormProps> = ({
  isOpen,
  onClose,
  onSuccess,
  account,
  mode
}) => {
  const { token } = useAuth();
  const [loading, setLoading] = useState(false);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [autoCreateGL, setAutoCreateGL] = useState(true);
  const [activeTab, setActiveTab] = useState(0);
  const toast = useToast();

  // Color mode values
  const modalContentBg = useColorModeValue('white', 'gray.800');
  const modalHeaderBg = useColorModeValue('blue.500', 'blue.600');
  const modalFooterBg = useColorModeValue('gray.50', 'gray.700');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const textColor = useColorModeValue('gray.700', 'gray.300');
  const mutedTextColor = useColorModeValue('gray.500', 'gray.400');
  const readOnlyBg = useColorModeValue('gray.50', 'gray.700');
  const cardInfoBg = useColorModeValue('blue.50', 'blue.900');
  const cardInfoTextColor = useColorModeValue('blue.700', 'blue.200');
  const cardInfoBorderColor = useColorModeValue('blue.200', 'blue.700');
  const cardSuccessBg = useColorModeValue('green.50', 'green.900');
  const cardSuccessTextColor = useColorModeValue('green.700', 'green.200');
  const cardSuccessBorderColor = useColorModeValue('green.200', 'green.700');
  const warningTextColor = useColorModeValue('orange.500', 'orange.300');
  const scrollbarTrackColor = useColorModeValue('#f1f1f1', '#2d3748');
  const scrollbarThumbColor = useColorModeValue('#c1c1c1', '#4a5568');
  const scrollbarThumbHoverColor = useColorModeValue('#a8a8a8', '#718096');

  const {
    control,
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<CashBankFormData>({
    defaultValues: {
      name: '',
      type: 'CASH',
      bank_name: '',
      account_no: '',
      currency: 'IDR',
      opening_balance: 0,
      opening_date: new Date().toISOString().split('T')[0],
      description: '',
      account_id: undefined
    }
  });

  const watchedType = watch('type');

  // Load available GL accounts (Asset accounts for cash/bank)
  useEffect(() => {
    const loadAccounts = async () => {
      if (token && isOpen && mode === 'create') {
        try {
          const allAccounts = await accountService.getAccounts(token, 'ASSET');
          // Filter only asset accounts that are not headers (can be used for cash/bank)
          const assetAccounts = allAccounts.filter(acc => 
            acc.type === 'ASSET' && 
            !acc.is_header && 
            acc.is_active &&
            (acc.category === 'CURRENT_ASSET' || !acc.category)
          );
          setAccounts(assetAccounts);
        } catch (error) {
          console.error('Error loading accounts:', error);
        }
      }
    };
    
    loadAccounts();
  }, [token, isOpen, mode]);

  // Initialize form data when editing or opening modal
  useEffect(() => {
    if (isOpen) {
      if (mode === 'edit' && account) {
        reset({
          name: account.name,
          type: account.type,
          bank_name: account.bank_name || '',
          account_no: account.account_no || '',
          currency: account.currency,
          description: account.description || '',
          account_id: account.account_id
        });
        setAutoCreateGL(false); // Disable auto-create when editing
      } else if (mode === 'create') {
        reset({
          name: '',
          type: 'CASH',
          bank_name: '',
          account_no: '',
          currency: 'IDR',
          opening_balance: 0,
          opening_date: new Date().toISOString().split('T')[0],
          description: '',
          account_id: undefined
        });
        setAutoCreateGL(true); // Enable auto-create for new accounts
      }
      setActiveTab(0); // Reset to first tab
    }
  }, [isOpen, mode, account, reset]);

  const onSubmit = async (data: CashBankFormData) => {
    try {
      setLoading(true);

      // Basic Validation
      if (!data.name.trim()) {
        throw new Error('Account name is required');
      }

      if (data.type === 'BANK' && !data.bank_name?.trim()) {
        throw new Error('Bank name is required for bank accounts');
      }

      // COA Integration Validation (only for create mode)
      if (mode === 'create' && !autoCreateGL && !data.account_id) {
        throw new Error('Please select a GL account from Chart of Accounts or enable auto-create option');
      }

      // Prepare data for submission
      const submitData = { ...data };
      
      // If auto-create is enabled, don't send account_id (backend will create)
      if (autoCreateGL) {
        delete submitData.account_id;
      }

      if (mode === 'create') {
        await cashbankService.createCashBankAccount(submitData as CashBankCreateRequest);
        toast({
          title: 'Success',
          description: `${data.type === 'CASH' ? 'Cash' : 'Bank'} account created successfully`,
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      } else if (mode === 'edit' && account) {
        const updateData: CashBankUpdateRequest = {
          name: data.name,
          bank_name: data.bank_name,
          account_no: data.account_no,
          description: data.description
        };
        await cashbankService.updateCashBankAccount(account.id, updateData);
        toast({
          title: 'Success',
          description: 'Account updated successfully',
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      }

      onSuccess();
      onClose();
    } catch (err: any) {
      console.error('Error saving account:', err);
      toast({
        title: 'Error',
        description: err.response?.data?.details || err.message || 'Failed to save account',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    reset();
    setActiveTab(0);
    onClose();
  };

  return (
    <Modal 
      isOpen={isOpen} 
      onClose={handleClose} 
      size="4xl"
      scrollBehavior="inside"
    >
      <ModalOverlay bg="blackAlpha.600" backdropFilter="blur(10px)" />
      <ModalContent 
        maxH="95vh" 
        maxW={{ base: '95vw', md: '90vw', lg: '70vw' }}
        mx={4}
        bg={modalContentBg}
      >
        <ModalHeader 
          bg={modalHeaderBg} 
          color="white" 
          borderTopRadius="md"
          py={4}
        >
          {mode === 'create' ? 'Create Cash/Bank Account' : 'Edit Account'}
        </ModalHeader>
        <ModalCloseButton color="white" />

        <form onSubmit={handleSubmit(onSubmit)}>
          <ModalBody 
            overflowY="auto"
            maxH="70vh"
            css={{
              '&::-webkit-scrollbar': {
                width: '8px',
              },
              '&::-webkit-scrollbar-track': {
                background: scrollbarTrackColor,
                borderRadius: '4px',
              },
              '&::-webkit-scrollbar-thumb': {
                background: scrollbarThumbColor,
                borderRadius: '4px',
                '&:hover': {
                  background: scrollbarThumbHoverColor,
                },
              },
              scrollBehavior: 'smooth',
              willChange: 'scroll-position',
              transform: 'translateZ(0)',
              WebkitOverflowScrolling: 'touch',
            }}
          >
            <Tabs index={activeTab} onChange={setActiveTab}>
              <TabList>
                <Tab>Account Details</Tab>
                {mode === 'create' && <Tab>COA Integration</Tab>}
              </TabList>

              <TabPanels>
                {/* Account Details Tab */}
                <TabPanel px={0}>
                  <VStack spacing={6} align="stretch">
                    {/* Account Type Selection */}
                    <Card>
                      <CardBody>
                        <FormControl isRequired>
                          <FormLabel>
                            <HStack>
                              <FiSettings />
                              <Text>Account Type</Text>
                            </HStack>
                          </FormLabel>
                          <Controller
                            name="type"
                            control={control}
                            rules={{ required: 'Account type is required' }}
                            render={({ field }) => (
                              <RadioGroup
                                {...field}
                                isDisabled={mode === 'edit'}
                              >
                                <HStack spacing={8}>
                                  <Radio value="CASH">
                                    <HStack>
                                      <Text fontSize="2xl">💵</Text>
                                      <Box>
                                        <Text>Cash Account</Text>
                                        <Text fontSize="xs" color={mutedTextColor}>Physical money management</Text>
                                      </Box>
                                    </HStack>
                                  </Radio>
                                  <Radio value="BANK">
                                    <HStack>
                                      <Text fontSize="2xl">🏦</Text>
                                      <Box>
                                        <Text>Bank Account</Text>
                                        <Text fontSize="xs" color={mutedTextColor}>Electronic banking</Text>
                                      </Box>
                                    </HStack>
                                  </Radio>
                                </HStack>
                              </RadioGroup>
                            )}
                          />
                          {mode === 'edit' && (
                            <Text fontSize="xs" color={warningTextColor} mt={2}>
                              ⚠️ Account type cannot be changed after creation
                            </Text>
                          )}
                        </FormControl>
                      </CardBody>
                    </Card>

                    {/* Basic Account Information */}
                    <Card>
                      <CardBody>
                        <Text fontWeight="bold" mb={4} color={textColor}>
                          📋 Basic Information
                        </Text>
                        
                        <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4}>
                          <FormControl isRequired isInvalid={!!errors.name}>
                            <FormLabel>
                              <HStack>
                                <FiFileText />
                                <Text>Account Name</Text>
                              </HStack>
                            </FormLabel>
                            <Input
                              {...register('name', {
                                required: 'Account name is required',
                                minLength: { value: 2, message: 'Account name must be at least 2 characters' }
                              })}
                              placeholder={watchedType === 'CASH' ? 'e.g., Petty Cash - Main Office' : 'e.g., BCA Main Account'}
                            />
                            <FormErrorMessage>{errors.name?.message}</FormErrorMessage>
                          </FormControl>

                          <FormControl>
                            <FormLabel>
                              <HStack>
                                <FiDollarSign />
                                <Text>Currency</Text>
                              </HStack>
                            </FormLabel>
                            <Input
                              value="Indonesian Rupiah (IDR)"
                              isReadOnly
                              bg={readOnlyBg}
                              color={textColor}
                              fontWeight="medium"
                            />
                            <Text fontSize="xs" color={mutedTextColor} mt={1}>
                              Currency is automatically set to Indonesian Rupiah
                            </Text>
                          </FormControl>
                        </SimpleGrid>

                        <FormControl mt={4}>
                          <FormLabel>Description</FormLabel>
                          <Textarea
                            {...register('description')}
                            placeholder="Optional description for this account"
                            rows={3}
                            resize="vertical"
                          />
                        </FormControl>
                      </CardBody>
                    </Card>

                    {/* Bank Details (only for BANK type) */}
                    {watchedType === 'BANK' && (
                      <Card>
                        <CardBody>
                          <Text fontWeight="bold" mb={4} color={textColor}>
                            🏦 Bank Details
                          </Text>
                          
                          <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4}>
                            <FormControl isRequired isInvalid={!!errors.bank_name}>
                              <FormLabel>
                                <HStack>
                                  <FiCreditCard />
                                  <Text>Bank Name</Text>
                                </HStack>
                              </FormLabel>
                              <Input
                                {...register('bank_name', {
                                  required: watchedType === 'BANK' ? 'Bank name is required for bank accounts' : false
                                })}
                                placeholder="e.g., Bank Central Asia"
                              />
                              <FormErrorMessage>{errors.bank_name?.message}</FormErrorMessage>
                            </FormControl>

                            <FormControl>
                              <FormLabel>Account Number</FormLabel>
                              <Input
                                {...register('account_no')}
                                placeholder="e.g., 1234567890"
                              />
                              <Text fontSize="xs" color={mutedTextColor} mt={1}>
                                Optional: Bank account number for reference
                              </Text>
                            </FormControl>
                          </SimpleGrid>
                        </CardBody>
                      </Card>
                    )}

                    {/* Opening Balance (only for create mode) */}
                    {mode === 'create' && (
                      <Card>
                        <CardBody>
                          <Text fontWeight="bold" mb={4} color={textColor}>
                            💰 Initial Setup
                          </Text>
                          
                          <SimpleGrid columns={{ base: 1, md: 2 }} spacing={4}>
                            <FormControl>
                              <FormLabel>
                                <HStack>
                                  <FiDollarSign />
                                  <Text>Opening Balance</Text>
                                </HStack>
                              </FormLabel>
                              <Controller
                                name="opening_balance"
                                control={control}
                                render={({ field }) => (
                                  <NumberInput
                                    {...field}
                                    min={0}
                                    precision={2}
                                    onChange={(_, value) => field.onChange(value || 0)}
                                  >
                                    <NumberInputField />
                                    <NumberInputStepper>
                                      <NumberIncrementStepper />
                                      <NumberDecrementStepper />
                                    </NumberInputStepper>
                                  </NumberInput>
                                )}
                              />
                              <Text fontSize="xs" color={mutedTextColor} mt={1}>
                                Initial balance when creating this account
                              </Text>
                            </FormControl>

                            <FormControl isRequired isInvalid={!!errors.opening_date}>
                              <FormLabel>
                                <HStack>
                                  <FiCalendar />
                                  <Text>Opening Date</Text>
                                </HStack>
                              </FormLabel>
                              <Input
                                type="date"
                                {...register('opening_date', {
                                  required: 'Opening date is required',
                                  validate: {
                                    notFuture: (value) => {
                                      const today = new Date();
                                      const inputDate = new Date(value);
                                      return inputDate <= today || 'Opening date cannot be in the future';
                                    },
                                  },
                                })}
                              />
                              <FormErrorMessage>{errors.opening_date?.message}</FormErrorMessage>
                            </FormControl>
                          </SimpleGrid>
                        </CardBody>
                      </Card>
                    )}

                    {/* Next Step for COA Integration */}
                    {mode === 'create' && (
                      <Card bg={cardInfoBg} borderColor={cardInfoBorderColor}>
                        <CardBody>
                          <HStack justify="space-between" align="center">
                            <Box>
                              <Text fontWeight="bold" color={cardInfoTextColor} mb={1}>
                                📊 Chart of Accounts Integration
                              </Text>
                              <Text fontSize="sm" color={cardInfoTextColor}>
                                Configure how this account links to your Chart of Accounts
                              </Text>
                            </Box>
                            <Button
                              size="sm"
                              colorScheme="blue"
                              variant="outline"
                              onClick={() => setActiveTab(1)}
                            >
                              Configure
                            </Button>
                          </HStack>
                        </CardBody>
                      </Card>
                    )}
                  </VStack>
                </TabPanel>

                {/* COA Integration Tab */}
                {mode === 'create' && (
                  <TabPanel px={0}>
                    <VStack spacing={6} align="stretch">
                      {/* COA Integration Info */}
                      <Alert status="info" borderRadius="md">
                        <AlertIcon />
                        <Box>
                          <AlertTitle>Chart of Accounts Integration</AlertTitle>
                          <AlertDescription>
                            Each cash/bank account must be linked to a GL account in your Chart of Accounts. 
                            This ensures proper financial reporting and audit trails.
                          </AlertDescription>
                        </Box>
                      </Alert>

                      {/* Auto-create Option */}
                      <Card>
                        <CardBody>
                          <VStack align="stretch" spacing={4}>
                            <Checkbox 
                              isChecked={autoCreateGL}
                              onChange={(e) => setAutoCreateGL(e.target.checked)}
                              colorScheme="blue"
                              size="lg"
                            >
                              <VStack align="start" spacing={1} ml={3}>
                                <Text fontWeight="bold" color={cardInfoTextColor}>
                                  🔄 Auto-create GL Account (Recommended)
                                </Text>
                                <Text fontSize="sm" color={textColor}>
                                  Let the system automatically create and link a GL account
                                </Text>
                              </VStack>
                            </Checkbox>
                            
                            {autoCreateGL ? (
                              <Alert status="success" borderRadius="md">
                                <AlertIcon />
                                <VStack align="start" spacing={2}>
                                  <Text fontSize="sm" fontWeight="medium">
                                    System will automatically create:
                                  </Text>
                                  <VStack align="start" spacing={1} pl={4}>
                                    <Text fontSize="xs" color="green.700">
                                      • GL Code: {watchedType === 'CASH' ? '1100-series' : '1110-series'} (Current Asset)
                                    </Text>
                                    <Text fontSize="xs" color="green.700">
                                      • GL Name: Same as cash/bank account name
                                    </Text>
                                    <Text fontSize="xs" color="green.700">
                                      • Category: CURRENT_ASSET for proper balance sheet reporting
                                    </Text>
                                  </VStack>
                                </VStack>
                              </Alert>
                            ) : (
                              <FormControl isRequired>
                                <FormLabel fontSize="sm">
                                  🎯 Select Existing GL Account (Asset Type Only)
                                </FormLabel>
                                <Select
                                  placeholder="Choose GL account from Chart of Accounts..."
                                  {...register('account_id', {
                                    setValueAs: value => value ? parseInt(value) : undefined
                                  })}
                                  size="sm"
                                >
                                  {accounts.map((acc) => (
                                    <option key={acc.id} value={acc.id}>
                                      [{acc.code}] {acc.name} - Balance: {accountService.formatBalance(acc.balance)}
                                    </option>
                                  ))}
                                </Select>
                                <Text fontSize="xs" color={mutedTextColor} mt={1}>
                                  Only active Asset accounts (Current Asset category) are shown. 
                                  Need more accounts? Visit <strong>Chart of Accounts</strong> page to create them.
                                </Text>
                              </FormControl>
                            )}
                          </VStack>
                        </CardBody>
                      </Card>

                      {/* COA Benefits */}
                      <Card bg={cardSuccessBg} borderColor={cardSuccessBorderColor}>
                        <CardBody>
                          <Text fontWeight="bold" color={cardSuccessTextColor} mb={3}>
                            ✅ Benefits of COA Integration
                          </Text>
                          <VStack align="start" spacing={2}>
                            <HStack>
                              <Text fontSize="sm" color={cardSuccessTextColor}>
                                📈 Automatic journal entries for all transactions
                              </Text>
                            </HStack>
                            <HStack>
                              <Text fontSize="sm" color={cardSuccessTextColor}>
                                📊 Proper balance sheet and income statement reporting
                              </Text>
                            </HStack>
                            <HStack>
                              <Text fontSize="sm" color={cardSuccessTextColor}>
                                🔍 Full audit trail and compliance readiness
                              </Text>
                            </HStack>
                            <HStack>
                              <Text fontSize="sm" color={cardSuccessTextColor}>
                                🔄 Real-time synchronization between cash/bank and GL
                              </Text>
                            </HStack>
                          </VStack>
                        </CardBody>
                      </Card>
                    </VStack>
                  </TabPanel>
                )}
              </TabPanels>
            </Tabs>
          </ModalBody>

          <ModalFooter 
            bg={modalFooterBg} 
            borderBottomRadius="md"
            py={4}
            px={6}
            borderTop="1px"
            borderColor={borderColor}
          >
            <HStack spacing={3} width="full" justify="flex-end">
              <Button 
                variant="outline" 
                onClick={handleClose}
                size={{ base: 'sm', md: 'md' }}
                minW="80px"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                colorScheme="blue"
                isLoading={isSubmitting || loading}
                loadingText={mode === 'create' ? 'Creating...' : 'Updating...'}
                size={{ base: 'sm', md: 'md' }}
                minW="140px"
                leftIcon={watchedType === 'CASH' ? <FiDollarSign /> : <FiCreditCard />}
              >
                {mode === 'create' ? 'Create Account' : 'Update Account'}
              </Button>
            </HStack>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};

export default CashBankForm;
