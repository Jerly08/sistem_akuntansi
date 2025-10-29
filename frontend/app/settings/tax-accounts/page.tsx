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
  FiActivity,
  FiLock
} from 'react-icons/fi';

interface AccountOption {
  id: number;
  code: string;
  name: string;
  type: string;
  category: string;
  is_active: boolean;
  is_system_critical?: boolean; // Lock critical accounts from modification
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

interface TaxConfig {
  id?: number;
  config_name: string;
  description: string;
  
  // Sales tax rates
  sales_ppn_rate: number;
  sales_pph21_rate: number;
  sales_pph23_rate: number;
  sales_other_tax_rate: number;
  
  // Purchase tax rates
  purchase_ppn_rate: number;
  purchase_pph21_rate: number;
  purchase_pph23_rate: number;
  purchase_pph25_rate: number;
  purchase_other_tax_rate: number;
  
  // Additional settings
  shipping_taxable: boolean;
  discount_before_tax: boolean;
  rounding_method: string;
  
  is_active: boolean;
  is_default: boolean;
  notes: string;
  updated_by_user?: {
    id: number;
    name: string;
    username: string;
  };
  created_at?: string;
  updated_at?: string;
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

  // Critical accounts that should not be changed
  const isCriticalAccount = (fieldName: string): boolean => {
    const criticalAccounts = [
      'sales_receivable',   // 1201 - Used in ALL credit sales
      'sales_revenue',      // 4101 - Used in ALL sales
      'sales_output_vat',   // 2103 - Tax regulation (DJP)
      'purchase_payable',   // 2001 - Used in ALL credit purchases
    ];
    return criticalAccounts.includes(fieldName);
  };

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
        const settingsData = settingsRes.data.data;
        setSettings(settingsData);
        
        // Helper function to parse account JSON strings
        const parseAccount = (accountData: any) => {
          if (!accountData) return null;
          if (typeof accountData === 'string') {
            try {
              return JSON.parse(accountData);
            } catch (e) {
              console.error('Failed to parse account data:', e);
              return null;
            }
          }
          return accountData; // Already an object
        };
        
        // Transform the settings data to match form field names
        const formDataTransformed = {
          ...settingsData,
          // Parse account objects from JSON strings
          sales_receivable_account: parseAccount(settingsData.sales_receivable_account),
          sales_cash_account: parseAccount(settingsData.sales_cash_account),
          sales_bank_account: parseAccount(settingsData.sales_bank_account),
          sales_revenue_account: parseAccount(settingsData.sales_revenue_account),
          sales_output_vat_account: parseAccount(settingsData.sales_output_vat_account),
          purchase_payable_account: parseAccount(settingsData.purchase_payable_account),
          purchase_cash_account: parseAccount(settingsData.purchase_cash_account),
          purchase_bank_account: parseAccount(settingsData.purchase_bank_account),
          purchase_input_vat_account: parseAccount(settingsData.purchase_input_vat_account),
          purchase_expense_account: parseAccount(settingsData.purchase_expense_account),
          withholding_tax21_account: parseAccount(settingsData.withholding_tax21_account),
          withholding_tax23_account: parseAccount(settingsData.withholding_tax23_account),
          withholding_tax25_account: parseAccount(settingsData.withholding_tax25_account),
          tax_payable_account: parseAccount(settingsData.tax_payable_account),
          inventory_account: parseAccount(settingsData.inventory_account),
          cogs_account: parseAccount(settingsData.cogs_account),
          
          // Sales account IDs
          sales_receivable_account_id: parseAccount(settingsData.sales_receivable_account)?.id,
          sales_cash_account_id: parseAccount(settingsData.sales_cash_account)?.id,
          sales_bank_account_id: parseAccount(settingsData.sales_bank_account)?.id,
          sales_revenue_account_id: parseAccount(settingsData.sales_revenue_account)?.id,
          sales_output_vat_account_id: parseAccount(settingsData.sales_output_vat_account)?.id,
          
          // Purchase account IDs
          purchase_payable_account_id: parseAccount(settingsData.purchase_payable_account)?.id,
          purchase_cash_account_id: parseAccount(settingsData.purchase_cash_account)?.id,
          purchase_bank_account_id: parseAccount(settingsData.purchase_bank_account)?.id,
          purchase_input_vat_account_id: parseAccount(settingsData.purchase_input_vat_account)?.id,
          purchase_expense_account_id: parseAccount(settingsData.purchase_expense_account)?.id,
          
          // Optional account IDs
          withholding_tax21_account_id: parseAccount(settingsData.withholding_tax21_account)?.id,
          withholding_tax23_account_id: parseAccount(settingsData.withholding_tax23_account)?.id,
          withholding_tax25_account_id: parseAccount(settingsData.withholding_tax25_account)?.id,
          tax_payable_account_id: parseAccount(settingsData.tax_payable_account)?.id,
          inventory_account_id: parseAccount(settingsData.inventory_account)?.id,
          cogs_account_id: parseAccount(settingsData.cogs_account)?.id
        };
        
        setFormData(formDataTransformed);
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
      const fieldValue = formData[`${field}_account_id`];
      
      if (!fieldValue) {
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
    if (settings) {
      // Helper function to parse account JSON strings
      const parseAccount = (accountData: any) => {
        if (!accountData) return null;
        if (typeof accountData === 'string') {
          try {
            return JSON.parse(accountData);
          } catch (e) {
            console.error('Failed to parse account data:', e);
            return null;
          }
        }
        return accountData; // Already an object
      };
      
      // Transform the settings data to match form field names
      const formDataTransformed = {
        ...settings,
        // Parse account objects from JSON strings
        sales_receivable_account: parseAccount(settings.sales_receivable_account),
        sales_cash_account: parseAccount(settings.sales_cash_account),
        sales_bank_account: parseAccount(settings.sales_bank_account),
        sales_revenue_account: parseAccount(settings.sales_revenue_account),
        sales_output_vat_account: parseAccount(settings.sales_output_vat_account),
        purchase_payable_account: parseAccount(settings.purchase_payable_account),
        purchase_cash_account: parseAccount(settings.purchase_cash_account),
        purchase_bank_account: parseAccount(settings.purchase_bank_account),
        purchase_input_vat_account: parseAccount(settings.purchase_input_vat_account),
        purchase_expense_account: parseAccount(settings.purchase_expense_account),
        withholding_tax21_account: parseAccount(settings.withholding_tax21_account),
        withholding_tax23_account: parseAccount(settings.withholding_tax23_account),
        withholding_tax25_account: parseAccount(settings.withholding_tax25_account),
        tax_payable_account: parseAccount(settings.tax_payable_account),
        inventory_account: parseAccount(settings.inventory_account),
        cogs_account: parseAccount(settings.cogs_account),
        
        // Sales account IDs
        sales_receivable_account_id: parseAccount(settings.sales_receivable_account)?.id,
        sales_cash_account_id: parseAccount(settings.sales_cash_account)?.id,
        sales_bank_account_id: parseAccount(settings.sales_bank_account)?.id,
        sales_revenue_account_id: parseAccount(settings.sales_revenue_account)?.id,
        sales_output_vat_account_id: parseAccount(settings.sales_output_vat_account)?.id,
        
        // Purchase account IDs
        purchase_payable_account_id: parseAccount(settings.purchase_payable_account)?.id,
        purchase_cash_account_id: parseAccount(settings.purchase_cash_account)?.id,
        purchase_bank_account_id: parseAccount(settings.purchase_bank_account)?.id,
        purchase_input_vat_account_id: parseAccount(settings.purchase_input_vat_account)?.id,
        purchase_expense_account_id: parseAccount(settings.purchase_expense_account)?.id,
        
        // Optional account IDs
        withholding_tax21_account_id: parseAccount(settings.withholding_tax21_account)?.id,
        withholding_tax23_account_id: parseAccount(settings.withholding_tax23_account)?.id,
        withholding_tax25_account_id: parseAccount(settings.withholding_tax25_account)?.id,
        tax_payable_account_id: parseAccount(settings.tax_payable_account)?.id,
        inventory_account_id: parseAccount(settings.inventory_account)?.id,
        cogs_account_id: parseAccount(settings.cogs_account)?.id
      };
      
      setFormData(formDataTransformed);
    } else {
      setFormData({});
    }
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
    
    // Check if critical from field name OR from selected account's database flag
    const isCritical = isCriticalAccount(fieldName);
    const selectedAccount = availableAccounts.find(acc => acc.id === currentValue);
    const isSelectedCritical = selectedAccount?.is_system_critical === true;
    
    return (
      <FormControl isRequired={isRequired} isInvalid={!!hasError}>
        <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
          <HStack>
            <Text>{label}</Text>
            {isRequired && <Badge colorScheme="red" size="sm">Required</Badge>}
            {(isCritical || isSelectedCritical) && (
              <HStack spacing={1}>
                <Icon as={FiLock} color="red.500" boxSize={3} />
                <Badge colorScheme="red" size="sm">LOCKED</Badge>
              </HStack>
            )}
          </HStack>
        </FormLabel>
        <Select
          value={currentValue}
          onChange={(e) => handleAccountChange(fieldName, e.target.value)}
          placeholder={`Select ${label.toLowerCase()}`}
          variant="filled"
          isDisabled={isCritical || isSelectedCritical}
          bg={(isCritical || isSelectedCritical) ? 'gray.200' : undefined}
          cursor={(isCritical || isSelectedCritical) ? 'not-allowed' : undefined}
          opacity={(isCritical || isSelectedCritical) ? 0.7 : 1}
          _hover={{ bg: (isCritical || isSelectedCritical) ? 'gray.200' : 'gray.100' }}
          _focus={{ bg: 'white', borderColor: 'blue.500' }}
        >
          {options.map((account) => (
            <option key={account.id} value={account.id}>
              {account.code} - {account.name} ({account.type})
            </option>
          ))}
        </Select>
        {hasError && <FormErrorMessage>{hasError}</FormErrorMessage>}
        {(isCritical || isSelectedCritical) && !hasError && (
          <FormHelperText fontSize="xs" color="red.600" fontWeight="medium">
            <HStack spacing={1}>
              <Icon as={FiInfo} />
              <Text>
                🔒 This account is system critical and locked to ensure data integrity. 
                Changing it would break journal entries and reports.
                {isSelectedCritical && ' (Locked in database)'}
              </Text>
            </HStack>
          </FormHelperText>
        )}
        {description && !hasError && !isCritical && !isSelectedCritical && (
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
              <Heading as="h1" size="xl">Withholding Tax Account Settings</Heading>
              <Text color="gray.600" fontSize="sm">
                Configure withholding tax accounts for PPh 21, PPh 23, PPh 25, and other tax obligations
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

          {/* Main Content - Withholding Tax Accounts Only */}
          <Box width="full">
                <Box width="full">
            <Card>
              <CardHeader>
                <HStack spacing={3}>
                  <Icon as={FiDollarSign} boxSize={5} color={greenColor} />
                  <Heading size="md">Withholding Tax Accounts</Heading>
                  <Badge colorScheme="gray" size="sm">Optional</Badge>
                </HStack>
              </CardHeader>
              <CardBody>
                <SimpleGrid columns={[1, 1, 2]} spacing={6}>
                  {renderAccountSelect(
                    'withholding_tax21',
                    'PPh 21 Account',
                    false,
                    ['ASSET'],
                    ['CURRENT_ASSET'],
                    'For employee income tax withholding'
                  )}
                  
                  {renderAccountSelect(
                    'withholding_tax23',
                    'PPh 23 Account',
                    false,
                    ['ASSET'],
                    ['CURRENT_ASSET'],
                    'For vendor service tax withholding'
                  )}
                  
                  {renderAccountSelect(
                    'withholding_tax25',
                    'PPh 25 Account',
                    false,
                    ['ASSET'],
                    ['CURRENT_ASSET'],
                    'For installment tax payments'
                  )}
                  
                  {renderAccountSelect(
                    'tax_payable',
                    'Tax Payable Account',
                    false,
                    ['LIABILITY'],
                    ['CURRENT_LIABILITY'],
                    'For other tax obligations'
                  )}
                </SimpleGrid>
              </CardBody>
            </Card>
          </Box>

          {/* Current Status */}
          {settings && (
            <Card width="full" bg="gray.50">
              <CardHeader>
                <HStack spacing={2}>
                  <Icon as={FiCheck} color="green.500" />
                  <Heading size="sm">Current Withholding Tax Configuration</Heading>
                </HStack>
              </CardHeader>
              <CardBody>
                <SimpleGrid columns={[1, 2, 3]} spacing={6}>
                  <VStack spacing={2} alignItems="start">
                    <Text fontSize="xs" color="gray.500" fontWeight="bold">PPh 21 (Employee Tax)</Text>
                    {settings.withholding_tax21_account ? (
                      <HStack spacing={2}>
                        <Badge colorScheme="green">Configured</Badge>
                        <Text fontSize="sm">
                          {settings.withholding_tax21_account?.code} - {settings.withholding_tax21_account?.name}
                        </Text>
                      </HStack>
                    ) : (
                      <HStack spacing={2}>
                        <Badge colorScheme="gray">Not Set</Badge>
                        <Text fontSize="xs" color="gray.500">Optional</Text>
                      </HStack>
                    )}
                  </VStack>

                  <VStack spacing={2} alignItems="start">
                    <Text fontSize="xs" color="gray.500" fontWeight="bold">PPh 23 (Vendor Tax)</Text>
                    {settings.withholding_tax23_account ? (
                      <HStack spacing={2}>
                        <Badge colorScheme="green">Configured</Badge>
                        <Text fontSize="sm">
                          {settings.withholding_tax23_account?.code} - {settings.withholding_tax23_account?.name}
                        </Text>
                      </HStack>
                    ) : (
                      <HStack spacing={2}>
                        <Badge colorScheme="gray">Not Set</Badge>
                        <Text fontSize="xs" color="gray.500">Optional</Text>
                      </HStack>
                    )}
                  </VStack>

                  <VStack spacing={2} alignItems="start">
                    <Text fontSize="xs" color="gray.500" fontWeight="bold">PPh 25 (Installment Tax)</Text>
                    {settings.withholding_tax25_account ? (
                      <HStack spacing={2}>
                        <Badge colorScheme="green">Configured</Badge>
                        <Text fontSize="sm">
                          {settings.withholding_tax25_account?.code} - {settings.withholding_tax25_account?.name}
                        </Text>
                      </HStack>
                    ) : (
                      <HStack spacing={2}>
                        <Badge colorScheme="gray">Not Set</Badge>
                        <Text fontSize="xs" color="gray.500">Optional</Text>
                      </HStack>
                    )}
                  </VStack>

                  <VStack spacing={2} alignItems="start">
                    <Text fontSize="xs" color="gray.500" fontWeight="bold">Tax Payable</Text>
                    {settings.tax_payable_account ? (
                      <HStack spacing={2}>
                        <Badge colorScheme="green">Configured</Badge>
                        <Text fontSize="sm">
                          {settings.tax_payable_account?.code} - {settings.tax_payable_account?.name}
                        </Text>
                      </HStack>
                    ) : (
                      <HStack spacing={2}>
                        <Badge colorScheme="gray">Not Set</Badge>
                        <Text fontSize="xs" color="gray.500">Optional</Text>
                      </HStack>
                    )}
                  </VStack>

                  <VStack spacing={2} alignItems="start">
                    <Text fontSize="xs" color="gray.500" fontWeight="bold">Configuration Status</Text>
                    <Badge colorScheme={settings.is_active ? 'green' : 'gray'} fontSize="sm">
                      {settings.is_active ? 'Active' : 'Inactive'}
                    </Badge>
                    <Text fontSize="xs" color="gray.500">
                      Last updated: {new Date(settings.updated_at).toLocaleDateString()}
                    </Text>
                  </VStack>

                  <VStack spacing={2} alignItems="start">
                    <Text fontSize="xs" color="gray.500" fontWeight="bold">Updated By</Text>
                    <Text fontSize="sm">{settings.updated_by_user?.name || 'System'}</Text>
                    <Text fontSize="xs" color="gray.500">
                      {new Date(settings.updated_at).toLocaleTimeString()}
                    </Text>
                  </VStack>
                </SimpleGrid>
              </CardBody>
            </Card>
          )}
          </Box>
        </VStack>

        {/* Suggestions Modal */}
        <Modal isOpen={isSuggestionsOpen} onClose={onSuggestionsClose} size="4xl">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>
              <HStack spacing={3}>
                <Icon as={FiInfo} color="blue.500" />
                <Text>Withholding Tax Account Suggestions</Text>
              </HStack>
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              {suggestions && suggestions.tax && (
                <VStack spacing={6} alignItems="start" width="full">
                  <Alert status="info" variant="left-accent">
                    <AlertIcon />
                    <Box>
                      <AlertTitle fontSize="sm">Account Configuration Guide</AlertTitle>
                      <AlertDescription fontSize="xs">
                        Use these recommendations to help configure your withholding tax accounts. All accounts are optional.
                      </AlertDescription>
                    </Box>
                  </Alert>
                  
                  <SimpleGrid columns={[1, 1, 2]} spacing={4} width="full">
                    {Object.entries(suggestions.tax || {}).map(([key, suggestion]) => (
                      <Card key={key} size="sm" variant="outline" borderWidth="2px">
                        <CardHeader pb={2}>
                          <HStack spacing={2}>
                            <Icon as={FiDollarSign} color="green.500" />
                            <Text fontWeight="bold" fontSize="md">
                              {key.replace(/_/g, ' ').toUpperCase()}
                            </Text>
                          </HStack>
                        </CardHeader>
                        <CardBody pt={0}>
                          <VStack spacing={3} alignItems="start">
                            <Text fontSize="sm" color="gray.600">
                              {suggestion.description}
                            </Text>
                            <Divider />
                            <Box width="full">
                              <Text fontSize="xs" fontWeight="semibold" mb={1}>Suggested Account Codes:</Text>
                              <HStack spacing={2} wrap="wrap">
                                {suggestion.suggested_codes.map((code) => (
                                  <Badge key={code} colorScheme="blue" fontSize="xs" px={2} py={1}>
                                    {code}
                                  </Badge>
                                ))}
                              </HStack>
                            </Box>
                            <Box width="full">
                              <Text fontSize="xs" fontWeight="semibold" mb={1}>Account Types:</Text>
                              <HStack spacing={2} wrap="wrap">
                                {suggestion.recommended_types.map((type) => (
                                  <Badge key={type} colorScheme="purple" fontSize="xs" px={2} py={1}>
                                    {type}
                                  </Badge>
                                ))}
                              </HStack>
                            </Box>
                            <Box width="full">
                              <Text fontSize="xs" fontWeight="semibold" mb={1}>Categories:</Text>
                              <HStack spacing={2} wrap="wrap">
                                {suggestion.recommended_categories.map((cat) => (
                                  <Badge key={cat} colorScheme="gray" fontSize="xs" px={2} py={1}>
                                    {cat}
                                  </Badge>
                                ))}
                              </HStack>
                            </Box>
                          </VStack>
                        </CardBody>
                      </Card>
                    ))}
                  </SimpleGrid>
                  
                  <Alert status="warning" variant="left-accent">
                    <AlertIcon />
                    <Box>
                      <AlertDescription fontSize="xs">
                        💡 <strong>Note:</strong> These are optional configurations. Only configure the tax accounts that are relevant to your business operations.
                      </AlertDescription>
                    </Box>
                  </Alert>
                </VStack>
              )}
            </ModalBody>
            <ModalFooter>
              <Button colorScheme="blue" onClick={onSuggestionsClose}>Close</Button>
            </ModalFooter>
          </ModalContent>
        </Modal>
      </Box>
    </SimpleLayout>
  );
};

export default TaxAccountSettingsPage;