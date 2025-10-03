'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useTranslation } from '@/hooks/useTranslation';
import SimpleLayout from '@/components/layout/SimpleLayout';
import api from '@/services/api';
import {
  Box,
  VStack,
  HStack,
  Heading,
  Text,
  Card,
  CardBody,
  CardHeader,
  SimpleGrid,
  Icon,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Spinner,
  useColorModeValue,
  Divider,
  Select,
  FormControl,
  FormLabel,
  FormErrorMessage,
  FormHelperText,
  useToast,
  Button,
  ButtonGroup,
  Badge,
  Tooltip,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  useDisclosure,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Stack
} from '@chakra-ui/react';
import { 
  FiSave, 
  FiX, 
  FiSettings, 
  FiDollarSign, 
  FiShoppingCart, 
  FiCreditCard,
  FiInfo,
  FiRefreshCw,
  FiCheck,
  FiActivity
} from 'react-icons/fi';

interface AccountOption {
  id: number;
  code: string;
  name: string;
  type: string;
  category: string;
  is_active: boolean;
}

interface TaxAccountSettings {
  id?: number;
  // Sales accounts
  sales_receivable_account: AccountOption;
  sales_cash_account: AccountOption;
  sales_bank_account: AccountOption;
  sales_revenue_account: AccountOption;
  sales_output_vat_account: AccountOption;
  
  // Purchase accounts
  purchase_payable_account: AccountOption;
  purchase_cash_account: AccountOption;
  purchase_bank_account: AccountOption;
  purchase_input_vat_account: AccountOption;
  purchase_expense_account: AccountOption;
  
  // Optional accounts
  withholding_tax21_account?: AccountOption;
  withholding_tax23_account?: AccountOption;
  withholding_tax25_account?: AccountOption;
  tax_payable_account?: AccountOption;
  inventory_account?: AccountOption;
  cogs_account?: AccountOption;
  
  is_active: boolean;
  apply_to_all_companies: boolean;
  notes: string;
  updated_by_user: {
    id: number;
    name: string;
    username: string;
  };
  created_at: string;
  updated_at: string;
}

interface AccountSuggestions {
  sales: {
    [key: string]: {
      recommended_types: string[];
      recommended_categories: string[];
      suggested_codes: string[];
      description: string;
    }
  };
  purchase: {
    [key: string]: {
      recommended_types: string[];
      recommended_categories: string[];
      suggested_codes: string[];
      description: string;
    }
  };
  tax: {
    [key: string]: {
      recommended_types: string[];
      recommended_categories: string[];
      suggested_codes: string[];
      description: string;
    }
  };
  inventory: {
    [key: string]: {
      recommended_types: string[];
      recommended_categories: string[];
      suggested_codes: string[];
      description: string;
    }
  };
}

const TaxAccountSettingsPage: React.FC = () => {
  const { user } = useAuth();
  const { t } = useTranslation();
  const toast = useToast();
  
  const [settings, setSettings] = useState<TaxAccountSettings | null>(null);
  const [availableAccounts, setAvailableAccounts] = useState<AccountOption[]>([]);
  const [suggestions, setSuggestions] = useState<AccountSuggestions | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasChanges, setHasChanges] = useState(false);
  const [validationErrors, setValidationErrors] = useState<{[key: string]: string}>({});

  // Form state
  const [formData, setFormData] = useState<any>({});

  // Modal for suggestions
  const { isOpen: isSuggestionsOpen, onOpen: onSuggestionsOpen, onClose: onSuggestionsClose } = useDisclosure();

  const blueColor = useColorModeValue('blue.500', 'blue.300');
  const greenColor = useColorModeValue('green.500', 'green.300');
  const purpleColor = useColorModeValue('purple.500', 'purple.300');
  const orangeColor = useColorModeValue('orange.500', 'orange.300');

  // Fetch current settings
  const fetchSettings = async () => {
    setLoading(true);
    try {
      const [settingsRes, accountsRes, suggestionsRes] = await Promise.all([
        api.get('/api/v1/tax-accounts/current'),
        api.get('/api/v1/tax-accounts/accounts'),
        api.get('/api/v1/tax-accounts/suggestions')
      ]);

      if (settingsRes.data.success) {
        setSettings(settingsRes.data.data);
        setFormData(settingsRes.data.data);
      }

      if (accountsRes.data.success) {
        setAvailableAccounts(accountsRes.data.data);
      }

      if (suggestionsRes.data.success) {
        setSuggestions(suggestionsRes.data.data);
      }

      setHasChanges(false);
      setError(null);
    } catch (err: any) {
      console.error('Error fetching tax account settings:', err);
      setError(err.response?.data?.error || err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSettings();
  }, []);

  const handleAccountChange = (fieldName: string, accountId: string) => {
    const accountIdNum = parseInt(accountId);
    const selectedAccount = availableAccounts.find(acc => acc.id === accountIdNum);
    
    if (selectedAccount) {
      setFormData((prev: any) => ({
        ...prev,
        [`${fieldName}_account_id`]: accountIdNum,
        [`${fieldName}_account`]: selectedAccount
      }));
      setHasChanges(true);
      
      // Clear validation error for this field
      setValidationErrors(prev => ({
        ...prev,
        [fieldName]: ''
      }));
    }
  };

  const validateForm = () => {
    const errors: {[key: string]: string} = {};
    
    // Required fields validation
    const requiredFields = [
      'sales_receivable',
      'sales_cash', 
      'sales_bank',
      'sales_revenue',
      'sales_output_vat',
      'purchase_payable',
      'purchase_cash',
      'purchase_bank', 
      'purchase_input_vat',
      'purchase_expense'
    ];

    requiredFields.forEach(field => {
      if (!formData[`${field}_account_id`]) {
        errors[field] = `${field.replace(/_/g, ' ')} account is required`;
      }
    });

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSave = async () => {
    if (!validateForm()) {
      toast({
        title: 'Validation Error',
        description: 'Please select all required accounts',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
      return;
    }

    setSaving(true);
    try {
      const payload = {
        // Sales accounts
        sales_receivable_account_id: formData.sales_receivable_account_id,
        sales_cash_account_id: formData.sales_cash_account_id,
        sales_bank_account_id: formData.sales_bank_account_id,
        sales_revenue_account_id: formData.sales_revenue_account_id,
        sales_output_vat_account_id: formData.sales_output_vat_account_id,
        
        // Purchase accounts
        purchase_payable_account_id: formData.purchase_payable_account_id,
        purchase_cash_account_id: formData.purchase_cash_account_id,
        purchase_bank_account_id: formData.purchase_bank_account_id,
        purchase_input_vat_account_id: formData.purchase_input_vat_account_id,
        purchase_expense_account_id: formData.purchase_expense_account_id,
        
        // Optional accounts
        withholding_tax21_account_id: formData.withholding_tax21_account_id || null,
        withholding_tax23_account_id: formData.withholding_tax23_account_id || null,
        withholding_tax25_account_id: formData.withholding_tax25_account_id || null,
        tax_payable_account_id: formData.tax_payable_account_id || null,
        inventory_account_id: formData.inventory_account_id || null,
        cogs_account_id: formData.cogs_account_id || null,
        
        apply_to_all_companies: true,
        notes: formData.notes || ''
      };

      let response;
      if (settings?.id) {
        // Update existing
        response = await api.put(`/api/v1/tax-accounts/${settings.id}`, payload);
      } else {
        // Create new
        response = await api.post('/api/v1/tax-accounts', payload);
      }

      if (response.data.success) {
        setSettings(response.data.data);
        setFormData(response.data.data);
        setHasChanges(false);
        
        toast({
          title: 'Success',
          description: 'Tax account settings saved successfully',
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      }
    } catch (error: any) {
      toast({
        title: 'Save Error',
        description: error.response?.data?.details || error.message,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    setFormData(settings || {});
    setHasChanges(false);
    setValidationErrors({});
  };

  const handleRefresh = async () => {
    await fetchSettings();
    toast({
      title: 'Refreshed',
      description: 'Settings and accounts list refreshed',
      status: 'success',
      duration: 2000,
      isClosable: true,
    });
  };

  const getAccountOptions = (accountTypes: string[] = [], categories: string[] = []) => {
    let filtered = availableAccounts;
    
    if (accountTypes.length > 0) {
      filtered = filtered.filter(acc => accountTypes.includes(acc.type));
    }
    
    if (categories.length > 0) {
      filtered = filtered.filter(acc => categories.includes(acc.category));
    }
    
    return filtered.sort((a, b) => a.code.localeCompare(b.code));
  };

  const renderAccountSelect = (
    fieldName: string,
    label: string,
    isRequired = true,
    accountTypes: string[] = [],
    categories: string[] = [],
    description?: string
  ) => {
    const currentValue = formData[`${fieldName}_account_id`] || '';
    const hasError = validationErrors[fieldName];
    const options = getAccountOptions(accountTypes, categories);
    
    return (
      <FormControl isRequired={isRequired} isInvalid={!!hasError}>
        <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
          <HStack>
            <Text>{label}</Text>
            {isRequired && <Badge colorScheme="red" size="sm">Required</Badge>}
          </HStack>
        </FormLabel>
        <Select
          value={currentValue}
          onChange={(e) => handleAccountChange(fieldName, e.target.value)}
          placeholder={`Select ${label.toLowerCase()}`}
          variant="filled"
          _hover={{ bg: 'gray.100' }}
          _focus={{ bg: 'white', borderColor: 'blue.500' }}
        >
          {options.map((account) => (
            <option key={account.id} value={account.id}>
              {account.code} - {account.name} ({account.type})
            </option>
          ))}
        </Select>
        {hasError && <FormErrorMessage>{hasError}</FormErrorMessage>}
        {description && !hasError && (
          <FormHelperText fontSize="xs" color="gray.500">
            {description}
          </FormHelperText>
        )}
      </FormControl>
    );
  };

  // Loading state
  if (loading) {
    return (
      <SimpleLayout allowedRoles={['admin']}>
        <Box display="flex" alignItems="center" justifyContent="center" minH="400px">
          <VStack spacing={4}>
            <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
            <Text>Loading tax account settings...</Text>
          </VStack>
        </Box>
      </SimpleLayout>
    );
  }

  return (
    <SimpleLayout allowedRoles={['admin']}>
      <Box>
        <VStack spacing={6} alignItems="start">
          {/* Header */}
          <HStack justify="space-between" width="full">
            <VStack alignItems="start" spacing={2}>
              <Heading as="h1" size="xl">Tax Account Settings</Heading>
              <Text color="gray.600" fontSize="sm">
                Configure account mappings for sales and purchase transactions
              </Text>
            </VStack>
            
            <ButtonGroup spacing={2}>
              {hasChanges && (
                <>
                  <Button
                    colorScheme="green"
                    leftIcon={<FiSave />}
                    onClick={handleSave}
                    isLoading={saving}
                    loadingText="Saving..."
                    size="sm"
                  >
                    Save Changes
                  </Button>
                  <Button
                    variant="outline"
                    leftIcon={<FiX />}
                    onClick={handleCancel}
                    isDisabled={saving}
                    size="sm"
                  >
                    Cancel
                  </Button>
                </>
              )}
              <Button
                variant="outline"
                leftIcon={<FiRefreshCw />}
                onClick={handleRefresh}
                size="sm"
              >
                Refresh
              </Button>
              <Button
                variant="outline"
                leftIcon={<FiInfo />}
                onClick={onSuggestionsOpen}
                size="sm"
              >
                Suggestions
              </Button>
            </ButtonGroup>
          </HStack>

          {/* Status Alert */}
          {settings && (
            <Alert status="success" variant="left-accent">
              <AlertIcon />
              <Box>
                <AlertTitle>Current Configuration Active</AlertTitle>
                <AlertDescription>
                  Last updated: {new Date(settings.updated_at).toLocaleString()} by {settings.updated_by_user.name}
                </AlertDescription>
              </Box>
            </Alert>
          )}
          
          {error && (
            <Alert status="error" width="full">
              <AlertIcon />
              <AlertTitle>Error:</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Main Content */}
          <Tabs width="full" variant="enclosed" colorScheme="blue">
            <TabList>
              <Tab>
                <HStack spacing={2}>
                  <Icon as={FiShoppingCart} />
                  <Text>Sales Accounts</Text>
                </HStack>
              </Tab>
              <Tab>
                <HStack spacing={2}>
                  <Icon as={FiCreditCard} />
                  <Text>Purchase Accounts</Text>
                </HStack>
              </Tab>
              <Tab>
                <HStack spacing={2}>
                  <Icon as={FiDollarSign} />
                  <Text>Tax & Other Accounts</Text>
                </HStack>
              </Tab>
            </TabList>

            <TabPanels>
              {/* Sales Accounts Tab */}
              <TabPanel>
                <SimpleGrid columns={[1, 1, 2]} spacing={6}>
                  <Card>
                    <CardHeader>
                      <HStack spacing={3}>
                        <Icon as={FiShoppingCart} boxSize={5} color={blueColor} />
                        <Heading size="md">Sales Transaction Accounts</Heading>
                      </HStack>
                    </CardHeader>
                    <CardBody>
                      <VStack spacing={4} alignItems="start">
                        {renderAccountSelect(
                          'sales_receivable',
                          'Receivable Account',
                          true,
                          ['ASSET'],
                          ['CURRENT_ASSET'],
                          'Used for credit sales (Piutang Usaha)'
                        )}
                        <Divider />
                        
                        {renderAccountSelect(
                          'sales_cash',
                          'Cash Account',
                          true,
                          ['ASSET'],
                          ['CURRENT_ASSET'],
                          'Used for cash sales'
                        )}
                        <Divider />
                        
                        {renderAccountSelect(
                          'sales_bank',
                          'Bank Account',
                          true,
                          ['ASSET'],
                          ['CURRENT_ASSET'],
                          'Used for bank transfer sales'
                        )}
                      </VStack>
                    </CardBody>
                  </Card>

                  <Card>
                    <CardHeader>
                      <HStack spacing={3}>
                        <Icon as={FiActivity} boxSize={5} color={greenColor} />
                        <Heading size="md">Revenue & Tax Accounts</Heading>
                      </HStack>
                    </CardHeader>
                    <CardBody>
                      <VStack spacing={4} alignItems="start">
                        {renderAccountSelect(
                          'sales_revenue',
                          'Revenue Account',
                          true,
                          ['REVENUE'],
                          ['SALES_REVENUE', 'OPERATING_REVENUE'],
                          'Main revenue account for sales'
                        )}
                        <Divider />
                        
                        {renderAccountSelect(
                          'sales_output_vat',
                          'Output VAT Account',
                          true,
                          ['LIABILITY'],
                          ['CURRENT_LIABILITY'],
                          'PPN Keluaran - for sales tax obligation'
                        )}
                      </VStack>
                    </CardBody>
                  </Card>
                </SimpleGrid>
              </TabPanel>

              {/* Purchase Accounts Tab */}
              <TabPanel>
                <SimpleGrid columns={[1, 1, 2]} spacing={6}>
                  <Card>
                    <CardHeader>
                      <HStack spacing={3}>
                        <Icon as={FiCreditCard} boxSize={5} color={purpleColor} />
                        <Heading size="md">Purchase Transaction Accounts</Heading>
                      </HStack>
                    </CardHeader>
                    <CardBody>
                      <VStack spacing={4} alignItems="start">
                        {renderAccountSelect(
                          'purchase_payable',
                          'Payable Account',
                          true,
                          ['LIABILITY'],
                          ['CURRENT_LIABILITY'],
                          'Used for credit purchases (Hutang Usaha)'
                        )}
                        <Divider />
                        
                        {renderAccountSelect(
                          'purchase_cash',
                          'Cash Account',
                          true,
                          ['ASSET'],
                          ['CURRENT_ASSET'],
                          'Used for cash purchases'
                        )}
                        <Divider />
                        
                        {renderAccountSelect(
                          'purchase_bank',
                          'Bank Account',
                          true,
                          ['ASSET'],
                          ['CURRENT_ASSET'],
                          'Used for bank transfer purchases'
                        )}
                      </VStack>
                    </CardBody>
                  </Card>

                  <Card>
                    <CardHeader>
                      <HStack spacing={3}>
                        <Icon as={FiActivity} boxSize={5} color={orangeColor} />
                        <Heading size="md">Expense & Tax Accounts</Heading>
                      </HStack>
                    </CardHeader>
                    <CardBody>
                      <VStack spacing={4} alignItems="start">
                        {renderAccountSelect(
                          'purchase_input_vat',
                          'Input VAT Account',
                          true,
                          ['ASSET'],
                          ['CURRENT_ASSET'],
                          'PPN Masukan - for claimable purchase tax'
                        )}
                        <Divider />
                        
                        {renderAccountSelect(
                          'purchase_expense',
                          'Default Expense Account',
                          true,
                          ['EXPENSE'],
                          ['OPERATING_EXPENSE', 'ADMINISTRATIVE_EXPENSE'],
                          'Default account for purchase expenses'
                        )}
                      </VStack>
                    </CardBody>
                  </Card>
                </SimpleGrid>
              </TabPanel>

              {/* Tax & Other Accounts Tab */}
              <TabPanel>
                <SimpleGrid columns={[1, 1, 2]} spacing={6}>
                  <Card>
                    <CardHeader>
                      <HStack spacing={3}>
                        <Icon as={FiDollarSign} boxSize={5} color={greenColor} />
                        <Heading size="md">Withholding Tax Accounts</Heading>
                        <Badge colorScheme="gray" size="sm">Optional</Badge>
                      </HStack>
                    </CardHeader>
                    <CardBody>
                      <VStack spacing={4} alignItems="start">
                        {renderAccountSelect(
                          'withholding_tax21',
                          'PPh 21 Account',
                          false,
                          ['ASSET'],
                          ['CURRENT_ASSET'],
                          'For employee income tax withholding'
                        )}
                        <Divider />
                        
                        {renderAccountSelect(
                          'withholding_tax23',
                          'PPh 23 Account',
                          false,
                          ['ASSET'],
                          ['CURRENT_ASSET'],
                          'For vendor service tax withholding'
                        )}
                        <Divider />
                        
                        {renderAccountSelect(
                          'withholding_tax25',
                          'PPh 25 Account',
                          false,
                          ['ASSET'],
                          ['CURRENT_ASSET'],
                          'For installment tax payments'
                        )}
                        <Divider />
                        
                        {renderAccountSelect(
                          'tax_payable',
                          'Tax Payable Account',
                          false,
                          ['LIABILITY'],
                          ['CURRENT_LIABILITY'],
                          'For other tax obligations'
                        )}
                      </VStack>
                    </CardBody>
                  </Card>

                  <Card>
                    <CardHeader>
                      <HStack spacing={3}>
                        <Icon as={FiSettings} boxSize={5} color={purpleColor} />
                        <Heading size="md">Inventory Accounts</Heading>
                        <Badge colorScheme="gray" size="sm">Optional</Badge>
                      </HStack>
                    </CardHeader>
                    <CardBody>
                      <VStack spacing={4} alignItems="start">
                        {renderAccountSelect(
                          'inventory',
                          'Inventory Account',
                          false,
                          ['ASSET'],
                          ['CURRENT_ASSET'],
                          'For inventory/stock management'
                        )}
                        <Divider />
                        
                        {renderAccountSelect(
                          'cogs',
                          'Cost of Goods Sold',
                          false,
                          ['EXPENSE'],
                          ['COST_OF_GOODS_SOLD'],
                          'For recording cost of sold items'
                        )}
                      </VStack>
                    </CardBody>
                  </Card>
                </SimpleGrid>
              </TabPanel>
            </TabPanels>
          </Tabs>

          {/* Current Status */}
          {settings && (
            <Card width="full" bg="gray.50">
              <CardHeader>
                <HStack spacing={2}>
                  <Icon as={FiCheck} color="green.500" />
                  <Heading size="sm">Current Configuration Status</Heading>
                </HStack>
              </CardHeader>
              <CardBody>
                <SimpleGrid columns={[1, 2, 4]} spacing={4}>
                  <VStack spacing={1} alignItems="start">
                    <Text fontSize="xs" color="gray.500">SALES ACCOUNTS</Text>
                    <Text fontSize="sm" fontWeight="bold">
                      {settings.sales_receivable_account?.code} - {settings.sales_receivable_account?.name}
                    </Text>
                    <Text fontSize="sm">
                      {settings.sales_revenue_account?.code} - {settings.sales_revenue_account?.name}
                    </Text>
                  </VStack>
                  <VStack spacing={1} alignItems="start">
                    <Text fontSize="xs" color="gray.500">PURCHASE ACCOUNTS</Text>
                    <Text fontSize="sm" fontWeight="bold">
                      {settings.purchase_payable_account?.code} - {settings.purchase_payable_account?.name}
                    </Text>
                    <Text fontSize="sm">
                      {settings.purchase_input_vat_account?.code} - {settings.purchase_input_vat_account?.name}
                    </Text>
                  </VStack>
                  <VStack spacing={1} alignItems="start">
                    <Text fontSize="xs" color="gray.500">TAX ACCOUNTS</Text>
                    <Text fontSize="sm">
                      Output VAT: {settings.sales_output_vat_account?.code}
                    </Text>
                    <Text fontSize="sm">
                      Input VAT: {settings.purchase_input_vat_account?.code}
                    </Text>
                  </VStack>
                  <VStack spacing={1} alignItems="start">
                    <Text fontSize="xs" color="gray.500">STATUS</Text>
                    <Badge colorScheme={settings.is_active ? 'green' : 'gray'}>
                      {settings.is_active ? 'Active' : 'Inactive'}
                    </Badge>
                    <Text fontSize="xs" color="gray.500">
                      Updated: {new Date(settings.updated_at).toLocaleDateString()}
                    </Text>
                  </VStack>
                </SimpleGrid>
              </CardBody>
            </Card>
          )}
        </VStack>

        {/* Suggestions Modal */}
        <Modal isOpen={isSuggestionsOpen} onClose={onSuggestionsClose} size="6xl">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>Account Configuration Suggestions</ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              {suggestions && (
                <VStack spacing={6} alignItems="start">
                  <Text color="gray.600">
                    Use these recommendations to help configure your tax accounts properly.
                  </Text>
                  
                  <Tabs variant="enclosed" width="full">
                    <TabList>
                      <Tab>Sales</Tab>
                      <Tab>Purchase</Tab>
                      <Tab>Tax</Tab>
                      <Tab>Inventory</Tab>
                    </TabList>
                    
                    <TabPanels>
                      {['sales', 'purchase', 'tax', 'inventory'].map((category) => (
                        <TabPanel key={category}>
                          <SimpleGrid columns={[1, 2]} spacing={4}>
                            {Object.entries(suggestions[category as keyof AccountSuggestions] || {}).map(([key, suggestion]) => (
                              <Card key={key} size="sm" variant="outline">
                                <CardBody>
                                  <VStack spacing={2} alignItems="start">
                                    <Text fontWeight="bold" fontSize="sm">
                                      {key.replace(/_/g, ' ').toUpperCase()}
                                    </Text>
                                    <Text fontSize="xs" color="gray.600">
                                      {suggestion.description}
                                    </Text>
                                    <HStack spacing={2} wrap="wrap">
                                      <Text fontSize="xs">Suggested codes:</Text>
                                      {suggestion.suggested_codes.map((code) => (
                                        <Badge key={code} colorScheme="blue" size="sm">
                                          {code}
                                        </Badge>
                                      ))}
                                    </HStack>
                                    <HStack spacing={2} wrap="wrap">
                                      <Text fontSize="xs">Types:</Text>
                                      {suggestion.recommended_types.map((type) => (
                                        <Badge key={type} colorScheme="gray" size="sm">
                                          {type}
                                        </Badge>
                                      ))}
                                    </HStack>
                                  </VStack>
                                </CardBody>
                              </Card>
                            ))}
                          </SimpleGrid>
                        </TabPanel>
                      ))}
                    </TabPanels>
                  </Tabs>
                </VStack>
              )}
            </ModalBody>
            <ModalFooter>
              <Button onClick={onSuggestionsClose}>Close</Button>
            </ModalFooter>
          </ModalContent>
        </Modal>
      </Box>
    </SimpleLayout>
  );
};

export default TaxAccountSettingsPage;