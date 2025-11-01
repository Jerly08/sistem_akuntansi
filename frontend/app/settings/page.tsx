'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useTranslation } from '@/hooks/useTranslation';
import SimpleLayout from '@/components/layout/SimpleLayout';
import api from '@/services/api';
import { API_ENDPOINTS } from '@/config/api';
import { getImageUrl } from '@/utils/imageUrl';
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
  Input,
  Textarea,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  FormControl,
  FormLabel,
  FormErrorMessage,
  useToast,
  Button,
  ButtonGroup,
  Image,
  Stack,
  InputGroup,
  InputLeftElement,
  FormHelperText
} from '@chakra-ui/react';
import { FiHome, FiSettings, FiGlobe, FiCalendar, FiDollarSign, FiSave, FiX, FiCreditCard, FiTrendingUp } from 'react-icons/fi';
import Link from 'next/link';

// Helper for converting fiscal year formats between backend and date input
const MONTHS = [
  'January','February','March','April','May','June','July','August','September','October','November','December'
];

function monthDayStringToISO(src: string): string {
  if (!src) return '';
  const lower = src.trim().toLowerCase();
  // Month name + day (e.g., "january 1")
  const monthIdx = MONTHS.findIndex(m => lower.startsWith(m.toLowerCase()));
  if (monthIdx >= 0) {
    const dayMatch = lower.match(/(\d{1,2})/);
    const day = Math.min(Math.max(parseInt(dayMatch?.[1] || '1', 10), 1), 31);
    const year = new Date().getFullYear();
    const mm = String(monthIdx + 1).padStart(2, '0');
    const dd = String(day).padStart(2, '0');
    return `${year}-${mm}-${dd}`;
  }
  // Fallback: try parsing any other date-like string
  const d = new Date(src);
  if (!isNaN(d.getTime())) {
    const yyyy = d.getFullYear();
    const mm = String(d.getMonth() + 1).padStart(2, '0');
    const dd = String(d.getDate()).padStart(2, '0');
    return `${yyyy}-${mm}-${dd}`;
  }
  return '';
}

function isoToMonthDayString(iso: string): string {
  if (!iso) return '';
  const parts = iso.split('-');
  if (parts.length < 3) return '';
  const m = parseInt(parts[1], 10);
  const d = parseInt(parts[2], 10);
  if (!m || !d || m < 1 || m > 12) return '';
  return `${MONTHS[m - 1]} ${d}`;
}

// Format an ISO date string to a selected display format
function formatDateISO(iso: string, format: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  if (isNaN(d.getTime())) return '';
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, '0');
  const dd = String(d.getDate()).padStart(2, '0');
  switch (format) {
    case 'DD/MM/YYYY':
      return `${dd}/${mm}/${yyyy}`;
    case 'MM/DD/YYYY':
      return `${mm}/${dd}/${yyyy}`;
    case 'DD-MM-YYYY':
      return `${dd}-${mm}-${yyyy}`;
    default:
      return `${yyyy}-${mm}-${dd}`; // 'YYYY-MM-DD'
  }
}

// Compute current fiscal year range from a selected month-day (ISO)
function computeFiscalRange(currentISO: string): { startISO: string; endISO: string } | null {
  if (!currentISO) return null;
  const base = new Date(currentISO);
  if (isNaN(base.getTime())) return null;
  const month = base.getMonth();
  const day = base.getDate();

  const today = new Date();
  const thisYearStart = new Date(Date.UTC(today.getFullYear(), month, day));
  let start = thisYearStart;
  if (today.getTime() < thisYearStart.getTime()) {
    // fiscal year started last year
    start = new Date(Date.UTC(today.getFullYear() - 1, month, day));
  }
  // End date: day before next year's start
  const nextStart = new Date(Date.UTC(start.getUTCFullYear() + 1, month, day));
  const end = new Date(nextStart.getTime() - 24 * 60 * 60 * 1000);

  const toISO = (dt: Date) => `${dt.getUTCFullYear()}-${String(dt.getUTCMonth() + 1).padStart(2, '0')}-${String(dt.getUTCDate()).padStart(2, '0')}`;
  return { startISO: toISO(start), endISO: toISO(end) };
}

interface TaxAccountSettings {
  sales_receivable_account?: { id: number; code: string; name: string };
  sales_revenue_account?: { id: number; code: string; name: string };
  purchase_payable_account?: { id: number; code: string; name: string };
  withholding_tax21_account?: { id: number; code: string; name: string };
  withholding_tax23_account?: { id: number; code: string; name: string };
  withholding_tax25_account?: { id: number; code: string; name: string };
  inventory_account?: { id: number; code: string; name: string };
  cogs_account?: { id: number; code: string; name: string };
  is_active?: boolean;
  updated_at?: string;
}

interface SystemSettings {
  id?: number;
  company_name: string;
  company_address: string;
  company_phone: string;
  company_email: string;
  company_website?: string;
  company_logo?: string;
  tax_number?: string;
  currency: string;
  date_format: string;
  fiscal_year_start: string;
  default_tax_rate: number;
  language: string;
  timezone: string;
  thousand_separator: string;
  decimal_separator: string;
  decimal_places: number;
  invoice_prefix: string;
  invoice_next_number?: number; // removed from API – keep optional for backward compatibility
  quote_prefix: string;
  quote_next_number?: number; // removed from API – keep optional for backward compatibility
  purchase_prefix: string;
  purchase_next_number?: number; // removed from API – keep optional for backward compatibility
  // New: Payment prefixes
  payment_receivable_prefix?: string;
  payment_payable_prefix?: string;
  email_notifications: boolean;
  smtp_host?: string;
  smtp_port?: number;
  smtp_username?: string;
  smtp_from?: string;
}

const SettingsPage: React.FC = () => {
  const { user } = useAuth();
  const { t, language, setLanguage } = useTranslation();
  const toast = useToast();
  const [settings, setSettings] = useState<SystemSettings | null>(null);
  const [formData, setFormData] = useState<SystemSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);
  const [uploadingLogo, setUploadingLogo] = useState(false);
  // Local UI state for date input binding (ISO format required by <input type="date">)
  const [fiscalStartISO, setFiscalStartISO] = useState<string>('');
  // Tax account settings
  const [taxAccountSettings, setTaxAccountSettings] = useState<TaxAccountSettings | null>(null);
  const [loadingTaxAccounts, setLoadingTaxAccounts] = useState(false);
  
  // Move useColorModeValue to top level to fix hooks order
  const blueColor = useColorModeValue('blue.500', 'blue.300');
  const greenColor = useColorModeValue('green.500', 'green.300');
  const purpleColor = useColorModeValue('purple.500', 'purple.300');
  const orangeColor = useColorModeValue('orange.500', 'orange.300');

  const fetchTaxAccountSettings = async () => {
    setLoadingTaxAccounts(true);
    try {
      const response = await api.get('/api/v1/tax-accounts/current');
      if (response.data.success) {
        setTaxAccountSettings(response.data.data);
      }
    } catch (err: any) {
      console.error('Error fetching tax account settings:', err);
      // Don't show error if tax accounts are not configured yet
    } finally {
      setLoadingTaxAccounts(false);
    }
  };

  const fetchSettings = async () => {
    setLoading(true);
    try {
      const response = await api.get(API_ENDPOINTS.SETTINGS);
      if (response.data.success) {
        setSettings(response.data.data);
        setFormData(response.data.data);
        // derive ISO date for UI
        setFiscalStartISO(monthDayStringToISO(response.data.data?.fiscal_year_start));
        setHasChanges(false);
        // Sync language from settings
        if (response.data.data.language && response.data.data.language !== language) {
          setLanguage(response.data.data.language);
        }
      }
    } catch (err: any) {
      console.error('Error fetching settings:', err);
      setError(err.response?.data?.error || err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSettings();
    fetchTaxAccountSettings();
  }, []);

  // Keep ISO date in sync when settings are populated later
  useEffect(() => {
    if (settings) {
      setFiscalStartISO(prev => prev || monthDayStringToISO(settings.fiscal_year_start));
    }
  }, [settings]);

  const handleLanguageChange = async (newLanguage: string) => {
    if (newLanguage === 'id' || newLanguage === 'en') {
      setLanguage(newLanguage);
      // Update language in backend
      try {
        await api.put(API_ENDPOINTS.SETTINGS_UPDATE, { language: newLanguage });
        toast({
          title: t('settings.languageChanged'),
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      } catch (error) {
        console.error('Error updating language:', error);
      }
    }
  };

  const handleFormChange = (field: keyof SystemSettings, value: any) => {
    setFormData(prev => {
      if (!prev) {
        // If formData is null, create new object with current value
        return {
          ...settings,
          [field]: value
        } as SystemSettings;
      }
      return {
        ...prev,
        [field]: value
      };
    });
    setHasChanges(true);
  };

  const handleSave = async () => {
    if (!formData || !hasChanges) return;
    
    setSaving(true);
    try {
      const response = await api.put(API_ENDPOINTS.SETTINGS_UPDATE, formData);
      if (response.data.success) {
        setSettings(response.data.data);
        setFormData(response.data.data);
        setHasChanges(false);
        
        // Update language if changed
        if (response.data.data.language !== language) {
          setLanguage(response.data.data.language);
        }
        
        toast({
          title: t('settings.updateSuccess') || 'Settings updated successfully',
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      }
    } catch (error: any) {
      toast({
        title: t('settings.updateError') || 'Failed to update settings',
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
    setFormData(settings);
    setHasChanges(false);
  };

  const handleLogoUpload = async (file: File) => {
    if (!file) return;
    setUploadingLogo(true);
    try {
      const fd = new FormData();
      fd.append('image', file);
      const resp = await api.post('/api/v1/settings/company/logo', fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      const newPath: string | undefined = resp.data?.path;
      if (newPath) {
        setSettings(prev => prev ? { ...prev, company_logo: newPath } : prev);
        setFormData(prev => prev ? { ...prev, company_logo: newPath } : prev);
        toast({ title: 'Logo updated', status: 'success', duration: 2500, isClosable: true });
      } else {
        toast({ title: 'Upload succeeded but no path returned', status: 'warning', duration: 3000, isClosable: true });
      }
    } catch (err: any) {
      toast({ title: 'Failed to upload logo', description: err.response?.data?.error || err.message, status: 'error', duration: 5000, isClosable: true });
    } finally {
      setUploadingLogo(false);
    }
  };


  // Add safety check for formData initialization
  useEffect(() => {
    if (settings && !formData) {
      setFormData(settings);
    }
  }, [settings, formData]);

  // Loading state - moved after all hooks
  if (loading) {
    return (
<SimpleLayout allowedRoles={['admin']}>
      <Box>
          <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
          <Text ml={4}>{t('common.loading')}</Text>
        </Box>
      </SimpleLayout>
    );
  }

  // Don't render form until settings are loaded
  if (!settings) {
    return (
      <SimpleLayout allowedRoles={['admin']}>
        <Box>
          <Text>No settings data available</Text>
        </Box>
      </SimpleLayout>
    );
  }

  return (
<SimpleLayout allowedRoles={['admin']}>
      <Box>
        <VStack spacing={6} alignItems="start">
          <HStack justify="space-between" width="full">
            <Heading as="h1" size="xl">{t('settings.title')}</Heading>
            {hasChanges && (
              <ButtonGroup>
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
              </ButtonGroup>
            )}
          </HStack>
          
          {error && (
            <Alert status="error" width="full">
              <AlertIcon />
              <AlertTitle>{t('common.error')}:</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          
          <SimpleGrid columns={[1, 1, 2, 2]} spacing={6} width="full">
            {/* Company Information Card */}
            <Card border="1px" borderColor="gray.200" boxShadow="md">
              <CardHeader>
                <HStack spacing={3}>
                  <Icon as={FiHome} boxSize={6} color={blueColor} />
                  <Heading size="md">{t('settings.companyInfo')}</Heading>
                </HStack>
              </CardHeader>
              <CardBody>
                <VStack spacing={4} alignItems="start">
                  {/* Company Logo */}
                  <FormControl>
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      Company Logo
                    </FormLabel>
                    <Stack direction={{ base: 'column', md: 'row' }} spacing={4} align="center">
                      <Image
                        src={getImageUrl(formData?.company_logo || settings?.company_logo || '') || undefined}
                        alt="Company Logo"
                        boxSize="80px"
                        objectFit="contain"
                        borderRadius="md"
                        fallbackStrategy="onError"
                      />
                      <Input
                        type="file"
                        accept="image/*"
                        onChange={(e) => {
                          const f = e.target.files?.[0];
                          if (f) handleLogoUpload(f);
                        }}
                        isDisabled={uploadingLogo}
                      />
                      {uploadingLogo && <Spinner size="sm" />}
                    </Stack>
                  </FormControl>
                  <Divider />
                  
                  <FormControl>
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      {t('settings.companyName')}
                    </FormLabel>
                    <Input
                      value={formData?.company_name || settings?.company_name || ''}
                      onChange={(e) => handleFormChange('company_name', e.target.value)}
                      placeholder="Enter company name"
                      variant="filled"
                      _hover={{ bg: 'gray.100' }}
                      _focus={{ bg: 'white', borderColor: 'blue.500' }}
                    />
                  </FormControl>
                  <Divider />
                  
                  <FormControl>
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      {t('settings.address')}
                    </FormLabel>
                    <Textarea
                      value={formData?.company_address || settings?.company_address || ''}
                      onChange={(e) => handleFormChange('company_address', e.target.value)}
                      placeholder="Enter company address"
                      rows={3}
                      variant="filled"
                      _hover={{ bg: 'gray.100' }}
                      _focus={{ bg: 'white', borderColor: 'blue.500' }}
                    />
                  </FormControl>
                  <Divider />
                  
                  <FormControl>
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      {t('settings.phone')}
                    </FormLabel>
                    <Input
                      value={formData?.company_phone || settings?.company_phone || ''}
                      onChange={(e) => handleFormChange('company_phone', e.target.value)}
                      placeholder="Enter phone number"
                      type="tel"
                      variant="filled"
                      _hover={{ bg: 'gray.100' }}
                      _focus={{ bg: 'white', borderColor: 'blue.500' }}
                    />
                  </FormControl>
                  <Divider />
                  
                  <FormControl>
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      {t('settings.email')}
                    </FormLabel>
                    <Input
                      value={formData?.company_email || settings?.company_email || ''}
                      onChange={(e) => handleFormChange('company_email', e.target.value)}
                      placeholder="Enter email address"
                      type="email"
                      variant="filled"
                      _hover={{ bg: 'gray.100' }}
                      _focus={{ bg: 'white', borderColor: 'blue.500' }}
                    />
                  </FormControl>
                  <Divider />
                  
                  <FormControl>
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      {t('settings.taxNumber')}
                    </FormLabel>
                    <Input
                      value={formData?.tax_number || settings?.tax_number || ''}
                      onChange={(e) => handleFormChange('tax_number', e.target.value)}
                      placeholder="Enter tax number (optional)"
                      variant="filled"
                      _hover={{ bg: 'gray.100' }}
                      _focus={{ bg: 'white', borderColor: 'blue.500' }}
                    />
                  </FormControl>
                </VStack>
              </CardBody>
            </Card>

            {/* System Configuration Card */}
            <Card border="1px" borderColor="gray.200" boxShadow="md">
              <CardHeader>
                <HStack spacing={3}>
                  <Icon as={FiDollarSign} boxSize={6} color={greenColor} />
                  <Heading size="md">{t('settings.systemConfig')}</Heading>
                </HStack>
              </CardHeader>
              <CardBody>
                <VStack spacing={4} alignItems="start">
                  <FormControl>
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      {t('settings.dateFormat')}
                    </FormLabel>
                    <Select
                      value={formData?.date_format || settings?.date_format || 'YYYY-MM-DD'}
                      onChange={(e) => handleFormChange('date_format', e.target.value)}
                      variant="filled"
                      _hover={{ bg: 'gray.100' }}
                      _focus={{ bg: 'white', borderColor: 'blue.500' }}
                    >
                      <option value="YYYY-MM-DD">YYYY-MM-DD</option>
                      <option value="DD/MM/YYYY">DD/MM/YYYY</option>
                      <option value="MM/DD/YYYY">MM/DD/YYYY</option>
                      <option value="DD-MM-YYYY">DD-MM-YYYY</option>
                    </Select>
                  </FormControl>
                  <Divider />
                  
                  <FormControl>
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      {t('settings.fiscalYearStart')}
                    </FormLabel>
                    <InputGroup>
                      <InputLeftElement pointerEvents="none">
                        <Icon as={FiCalendar} color="gray.500" />
                      </InputLeftElement>
                      <Input
                        type="date"
                        value={fiscalStartISO}
                        onChange={(e) => {
                          const iso = e.target.value;
                          setFiscalStartISO(iso);
                          // store normalized form for backend
                          handleFormChange('fiscal_year_start', isoToMonthDayString(iso));
                        }}
                        variant="filled"
                        _hover={{ bg: 'gray.100' }}
                        _focus={{ bg: 'white', borderColor: 'blue.500' }}
                      />
                    </InputGroup>
                    <FormHelperText fontSize="xs" color="gray.600">
                      {language === 'id'
                        ? 'Hanya hari dan bulan yang digunakan. Tahun akan ditentukan otomatis untuk periode fiskal.'
                        : 'Only day and month are used. The year is determined automatically for the fiscal period.'}
                    </FormHelperText>
                    {(() => {
                      const fmt = formData?.date_format || settings?.date_format || 'YYYY-MM-DD';
                      const range = computeFiscalRange(fiscalStartISO);
                      if (!range) return null;
                      return (
                        <Text mt={1} fontSize="xs" color="gray.700">
                          {language === 'id' ? 'Periode fiskal saat ini:' : 'Current fiscal period:'} {formatDateISO(range.startISO, fmt)} — {formatDateISO(range.endISO, fmt)}
                        </Text>
                      );
                    })()}
                  </FormControl>
                  <Divider />
                  
                  <FormControl>
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      {t('settings.defaultTaxRate')}
                    </FormLabel>
                    <NumberInput
                      value={formData?.default_tax_rate ?? settings?.default_tax_rate ?? 0}
                      onChange={(valueString) => handleFormChange('default_tax_rate', parseFloat(valueString) || 0)}
                      min={0}
                      max={100}
                      precision={2}
                    >
                      <NumberInputField 
                        variant="filled"
                        _hover={{ bg: 'gray.100' }}
                        _focus={{ bg: 'white', borderColor: 'blue.500' }}
                      />
                      <NumberInputStepper>
                        <NumberIncrementStepper />
                        <NumberDecrementStepper />
                      </NumberInputStepper>
                    </NumberInput>
                  </FormControl>
                  <Divider />
                  
                  <FormControl>
                    <HStack mb={2}>
                      <Icon as={FiGlobe} boxSize={4} color="gray.600" />
                      <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm" mb={0}>
                        {t('settings.language')}
                      </FormLabel>
                    </HStack>
                    <Select 
                      value={formData?.language || settings?.language || language} 
                      onChange={(e) => {
                        handleFormChange('language', e.target.value);
                      }}
                      variant="filled"
                      _hover={{ bg: 'gray.100' }}
                      _focus={{ bg: 'white', borderColor: 'blue.500' }}
                    >
                      <option value="id">{t('settings.indonesian')}</option>
                      <option value="en">{t('settings.english')}</option>
                    </Select>
                  </FormControl>
                </VStack>
              </CardBody>
            </Card>
            
            {/* Tax Account Configuration Card */}
            <Card border="1px" borderColor="gray.200" boxShadow="md">
              <CardHeader>
                <HStack spacing={3}>
                  <Icon as={FiCreditCard} boxSize={6} color={orangeColor} />
                  <Heading size="md">{t('settings.taxAccountConfig')}</Heading>
                </HStack>
              </CardHeader>
              <CardBody>
                <VStack spacing={4} alignItems="start">
                  <Text fontSize="sm" color="gray.600" mb={2}>
                    Configure account mappings for sales and purchase transactions, including tax accounts and payment methods.
                  </Text>
                  
                  <Alert status="info" variant="left-accent" size="sm">
                    <AlertIcon />
                    <AlertDescription fontSize="sm">
                      Set up proper account mappings to ensure accurate financial reporting and tax calculations.
                    </AlertDescription>
                  </Alert>
                  
                  {loadingTaxAccounts ? (
                    <HStack width="full" justify="center" py={4}>
                      <Spinner size="sm" color="blue.500" />
                      <Text fontSize="xs" color="gray.500">Loading configuration...</Text>
                    </HStack>
                  ) : taxAccountSettings ? (
                    <VStack spacing={3} width="full" alignItems="start">
                      <HStack justify="space-between" width="full">
                        <VStack alignItems="start" spacing={1}>
                          <Text fontSize="sm" fontWeight="semibold" color="gray.700">
                            Withholding Tax Accounts
                          </Text>
                          <VStack spacing={1} alignItems="start" mt={2}>
                            {taxAccountSettings.withholding_tax21_account ? (
                              <HStack spacing={1}>
                                <Icon as={FiDollarSign} color="green.500" boxSize={3} />
                                <Text fontSize="xs" color="green.600" fontWeight="medium">
                                  PPh 21: {taxAccountSettings.withholding_tax21_account.code} - {taxAccountSettings.withholding_tax21_account.name}
                                </Text>
                              </HStack>
                            ) : (
                              <Text fontSize="xs" color="gray.400">PPh 21: Not configured</Text>
                            )}
                            {taxAccountSettings.withholding_tax23_account ? (
                              <HStack spacing={1}>
                                <Icon as={FiDollarSign} color="green.500" boxSize={3} />
                                <Text fontSize="xs" color="green.600" fontWeight="medium">
                                  PPh 23: {taxAccountSettings.withholding_tax23_account.code} - {taxAccountSettings.withholding_tax23_account.name}
                                </Text>
                              </HStack>
                            ) : (
                              <Text fontSize="xs" color="gray.400">PPh 23: Not configured</Text>
                            )}
                            {taxAccountSettings.withholding_tax25_account ? (
                              <HStack spacing={1}>
                                <Icon as={FiDollarSign} color="green.500" boxSize={3} />
                                <Text fontSize="xs" color="green.600" fontWeight="medium">
                                  PPh 25: {taxAccountSettings.withholding_tax25_account.code} - {taxAccountSettings.withholding_tax25_account.name}
                                </Text>
                              </HStack>
                            ) : (
                              <Text fontSize="xs" color="gray.400">PPh 25: Not configured</Text>
                            )}
                          </VStack>
                        </VStack>
                        <Icon as={FiDollarSign} color="orange.500" boxSize={4} />
                      </HStack>
                      
                      <Divider />
                      
                      <HStack justify="space-between" width="full">
                        <VStack alignItems="start" spacing={1}>
                          <Text fontSize="sm" fontWeight="semibold" color="gray.700">
                            Inventory & COGS
                          </Text>
                          <VStack spacing={1} alignItems="start" mt={2}>
                            {taxAccountSettings.inventory_account ? (
                              <HStack spacing={1}>
                                <Icon as={FiTrendingUp} color="blue.500" boxSize={3} />
                                <Text fontSize="xs" color="blue.600" fontWeight="medium">
                                  Inventory: {taxAccountSettings.inventory_account.code} - {taxAccountSettings.inventory_account.name}
                                </Text>
                              </HStack>
                            ) : (
                              <Text fontSize="xs" color="gray.400">Inventory: Not configured</Text>
                            )}
                            {taxAccountSettings.cogs_account ? (
                              <HStack spacing={1}>
                                <Icon as={FiTrendingUp} color="blue.500" boxSize={3} />
                                <Text fontSize="xs" color="blue.600" fontWeight="medium">
                                  COGS: {taxAccountSettings.cogs_account.code} - {taxAccountSettings.cogs_account.name}
                                </Text>
                              </HStack>
                            ) : (
                              <Text fontSize="xs" color="gray.400">COGS: Not configured</Text>
                            )}
                          </VStack>
                        </VStack>
                        <Icon as={FiTrendingUp} color="blue.500" boxSize={4} />
                      </HStack>
                    </VStack>
                  ) : (
                    <VStack spacing={3} width="full" alignItems="start">
                      <Text fontSize="xs" color="gray.500">
                        No tax account configuration found. Click the button below to configure withholding tax accounts.
                      </Text>
                    </VStack>
                  )}
                  
                  <Link href="/settings/tax-accounts">
                    <Button
                      colorScheme="orange"
                      variant="outline"
                      size="sm"
                      leftIcon={<FiSettings />}
                      width="full"
                      _hover={{ bg: 'orange.50' }}
                    >
                      Configure Tax Accounts
                    </Button>
                  </Link>
                </VStack>
              </CardBody>
            </Card>
          </SimpleGrid>
        </VStack>
      </Box>

    </SimpleLayout>
  );
};

export default SettingsPage;
