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
  FormHelperText,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  Tooltip
} from '@chakra-ui/react';
import { FiHome, FiSettings, FiGlobe, FiCalendar, FiDollarSign, FiSave, FiX, FiCreditCard, FiTrendingUp, FiLock, FiUnlock, FiCheckCircle, FiInfo } from 'react-icons/fi';
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

interface AccountingPeriod {
  id: number;
  year: number;
  month: number;
  period_name: string;
  is_closed: boolean;
  is_locked: boolean;
  closed_by?: number;
  closed_at?: string;
  start_date: string;
  end_date: string;
}

interface FiscalClosingPreview {
  fiscal_year_end: string;
  total_revenue: number;
  total_expense: number;
  net_income: number;
  retained_earnings_id: number;
  can_close: boolean;
  validation_messages: string[];
  revenue_accounts: Array<{id: number; code: string; name: string; balance: number}>;
  expense_accounts: Array<{id: number; code: string; name: string; balance: number}>;
  closing_entries: Array<{description: string; debit_account: string; credit_account: string; amount: number}>;
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
  
  // Accounting periods
  const [periods, setPeriods] = useState<AccountingPeriod[]>([]);
  const [loadingPeriods, setLoadingPeriods] = useState(false);
  const [closingPeriod, setClosingPeriod] = useState<{year: number; month: number} | null>(null);
  const [reopenPeriod, setReopenPeriod] = useState<{year: number; month: number} | null>(null);
  const [reopenReason, setReopenReason] = useState('');
  
  // Fiscal year closing
  const [fiscalYearEnd, setFiscalYearEnd] = useState('');
  const [fiscalClosingPreview, setFiscalClosingPreview] = useState<FiscalClosingPreview | null>(null);
  const [loadingFiscalPreview, setLoadingFiscalPreview] = useState(false);
  const [showFiscalClosingModal, setShowFiscalClosingModal] = useState(false);
  const [executingFiscalClosing, setExecutingFiscalClosing] = useState(false);
  
  // Move useColorModeValue to top level to fix hooks order
  const blueColor = useColorModeValue('blue.500', 'blue.300');
  const greenColor = useColorModeValue('green.500', 'green.300');
  const purpleColor = useColorModeValue('purple.500', 'purple.300');
  const orangeColor = useColorModeValue('orange.500', 'orange.300');
  const periodBgColor = useColorModeValue('gray.50', 'gray.700');

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
  
  const fetchPeriods = async () => {
    setLoadingPeriods(true);
    try {
      const response = await api.get('/api/v1/periods?limit=6');
      if (response.data.success) {
        setPeriods(response.data.data || []);
      }
    } catch (err: any) {
      console.error('Error fetching periods:', err);
    } finally {
      setLoadingPeriods(false);
    }
  };
  
  const handleClosePeriod = async (year: number, month: number) => {
    try {
      const response = await api.post(`/api/v1/periods/${year}/${month}/close`, {
        notes: 'Closed from settings'
      });
      if (response.data.success) {
        toast({
          title: 'Period Closed',
          description: `${MONTHS[month - 1]} ${year} has been closed successfully`,
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
        setClosingPeriod(null);
        fetchPeriods();
      }
    } catch (err: any) {
      toast({
        title: 'Failed to close period',
        description: err.response?.data?.details || err.message,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };
  
  const handleReopenPeriod = async () => {
    if (!reopenPeriod || !reopenReason.trim()) {
      toast({
        title: 'Reason required',
        description: 'Please provide a reason to reopen the period',
        status: 'warning',
        duration: 3000,
      });
      return;
    }
    
    try {
      const response = await api.post(
        `/api/v1/periods/${reopenPeriod.year}/${reopenPeriod.month}/reopen`,
        { reason: reopenReason }
      );
      if (response.data.success) {
        toast({
          title: 'Period Reopened',
          description: `${MONTHS[reopenPeriod.month - 1]} ${reopenPeriod.year} has been reopened`,
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
        setReopenPeriod(null);
        setReopenReason('');
        fetchPeriods();
      }
    } catch (err: any) {
      toast({
        title: 'Failed to reopen period',
        description: err.response?.data?.details || err.message,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };
  
  const handlePreviewFiscalClosing = async () => {
    if (!fiscalYearEnd) {
      toast({
        title: 'Date required',
        description: 'Please select fiscal year end date',
        status: 'warning',
        duration: 3000,
      });
      return;
    }
    
    setLoadingFiscalPreview(true);
    try {
      const response = await api.get(
        `/api/v1/fiscal-closing/preview?fiscal_year_end=${fiscalYearEnd}`
      );
      if (response.data.success) {
        setFiscalClosingPreview(response.data.data);
        setShowFiscalClosingModal(true);
      }
    } catch (err: any) {
      toast({
        title: 'Failed to preview closing',
        description: err.response?.data?.details || err.message,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoadingFiscalPreview(false);
    }
  };
  
  const handleExecuteFiscalClosing = async () => {
    if (!fiscalClosingPreview) return;
    
    setExecutingFiscalClosing(true);
    try {
      const response = await api.post('/api/v1/fiscal-closing/execute', {
        fiscal_year_end: fiscalYearEnd,
        notes: 'Year-end closing executed from settings'
      });
      if (response.data.success) {
        toast({
          title: 'Fiscal Year Closed!',
          description: 'All revenue and expense accounts have been reset and transferred to retained earnings',
          status: 'success',
          duration: 5000,
          isClosable: true,
        });
        setShowFiscalClosingModal(false);
        setFiscalClosingPreview(null);
        setFiscalYearEnd('');
        fetchPeriods();
      }
    } catch (err: any) {
      toast({
        title: 'Failed to close fiscal year',
        description: err.response?.data?.details || err.message,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setExecutingFiscalClosing(false);
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
    fetchPeriods();
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
            
            {/* Period Closing Management Card */}
            <Card border="1px" borderColor="gray.200" boxShadow="md">
              <CardHeader>
                <HStack spacing={3} justify="space-between" width="full">
                  <HStack spacing={3}>
                    <Icon as={FiCalendar} boxSize={6} color={purpleColor} />
                    <Heading size="md">Period Closing</Heading>
                  </HStack>
                  <Tooltip
                    label={
                      <Box p={2}>
                        <Text fontSize="xs" fontWeight="bold" mb={1}>📅 Monthly Period Closing</Text>
                        <Text fontSize="xs" mb={2}>
                          Lock periode bulanan untuk menjaga integritas data.
                        </Text>
                        <Text fontSize="xs" fontWeight="semibold" mb={1}>Yang Terjadi:</Text>
                        <Text fontSize="xs">• Periode di-lock untuk user biasa</Text>
                        <Text fontSize="xs">• Saldo Revenue/Expense TETAP (tidak direset)</Text>
                        <Text fontSize="xs">• Tidak ada closing entries</Text>
                        <Text fontSize="xs">• Bisa di-reopen dengan reason</Text>
                        <Text fontSize="xs" mt={2} fontStyle="italic">
                          Gunakan ini setiap akhir bulan untuk lock transaksi.
                        </Text>
                      </Box>
                    }
                    placement="left"
                    hasArrow
                    bg="purple.600"
                  >
                    <Box cursor="help">
                      <Icon as={FiInfo} boxSize={4} color="gray.500" _hover={{ color: purpleColor }} />
                    </Box>
                  </Tooltip>
                </HStack>
              </CardHeader>
              <CardBody>
                <VStack spacing={4} alignItems="start">
                  <Text fontSize="sm" color="gray.600" mb={2}>
                    Manage monthly period closing to lock transactions and maintain data integrity.
                  </Text>
                  
                  {loadingPeriods ? (
                    <HStack width="full" justify="center" py={4}>
                      <Spinner size="sm" color="blue.500" />
                      <Text fontSize="xs" color="gray.500">Loading periods...</Text>
                    </HStack>
                  ) : periods.length > 0 ? (
                    <VStack spacing={2} width="full" alignItems="start">
                      {periods.slice(0, 4).map((period) => (
                        <HStack
                          key={period.id}
                          justify="space-between"
                          width="full"
                          p={2}
                          bg={periodBgColor}
                          borderRadius="md"
                        >
                          <HStack spacing={2}>
                            <Tooltip
                              label={
                                period.is_locked
                                  ? "🔒 Hard Locked - Fiscal year-end closed. Cannot reopen easily."
                                  : period.is_closed
                                  ? "🟠 Soft Locked - Monthly closed. Can reopen with reason."
                                  : "🟢 Open - Transactions allowed"
                              }
                              placement="top"
                              hasArrow
                            >
                              <Icon
                                as={period.is_locked ? FiLock : period.is_closed ? FiLock : FiUnlock}
                                color={period.is_locked ? 'red.700' : period.is_closed ? 'red.500' : 'green.500'}
                                boxSize={4}
                                cursor="help"
                              />
                            </Tooltip>
                            <Text fontSize="sm" fontWeight="medium">
                              {MONTHS[period.month - 1]} {period.year}
                            </Text>
                            <Tooltip
                              label={
                                <Box p={1}>
                                  {period.is_locked ? (
                                    <>
                                      <Text fontSize="xs" fontWeight="bold">Hard Locked 🔒</Text>
                                      <Text fontSize="xs">Fiscal year-end closed</Text>
                                      <Text fontSize="xs">Revenue/Expense reset to 0</Text>
                                    </>
                                  ) : period.is_closed ? (
                                    <>
                                      <Text fontSize="xs" fontWeight="bold">Soft Closed 🟠</Text>
                                      <Text fontSize="xs">Monthly closing</Text>
                                      <Text fontSize="xs">Can reopen with approval</Text>
                                    </>
                                  ) : (
                                    <>
                                      <Text fontSize="xs" fontWeight="bold">Open 🟢</Text>
                                      <Text fontSize="xs">Transactions allowed</Text>
                                      <Text fontSize="xs">Can close anytime</Text>
                                    </>
                                  )}
                                </Box>
                              }
                              placement="top"
                              hasArrow
                            >
                              <Text
                                fontSize="xs"
                                px={2}
                                py={1}
                                borderRadius="full"
                                bg={period.is_locked ? 'red.200' : period.is_closed ? 'red.100' : 'green.100'}
                                color={period.is_locked ? 'red.900' : period.is_closed ? 'red.700' : 'green.700'}
                                fontWeight="semibold"
                                cursor="help"
                              >
                                {period.is_locked ? 'Locked' : period.is_closed ? 'Closed' : 'Open'}
                              </Text>
                            </Tooltip>
                          </HStack>
                          
                          {!period.is_locked && (
                            <Button
                              size="xs"
                              colorScheme={period.is_closed ? 'blue' : 'red'}
                              variant="outline"
                              onClick={() =>
                                period.is_closed
                                  ? setReopenPeriod({ year: period.year, month: period.month })
                                  : setClosingPeriod({ year: period.year, month: period.month })
                              }
                            >
                              {period.is_closed ? 'Reopen' : 'Close'}
                            </Button>
                          )}
                        </HStack>
                      ))}
                    </VStack>
                  ) : (
                    <Text fontSize="xs" color="gray.500">
                      No periods available
                    </Text>
                  )}
                </VStack>
              </CardBody>
            </Card>
            
            {/* Fiscal Year-End Closing Card */}
            <Card border="1px" borderColor="gray.200" boxShadow="md">
              <CardHeader>
                <HStack spacing={3} justify="space-between" width="full">
                  <HStack spacing={3}>
                    <Icon as={FiCheckCircle} boxSize={6} color="red.500" />
                    <Heading size="md">Fiscal Year-End Closing</Heading>
                  </HStack>
                  <Tooltip
                    label={
                      <Box p={2}>
                        <Text fontSize="xs" fontWeight="bold" mb={1}>🏁 Fiscal Year-End Closing</Text>
                        <Text fontSize="xs" mb={2}>
                          Tutup buku akhir tahun dengan automated closing entries.
                        </Text>
                        <Text fontSize="xs" fontWeight="semibold" mb={1}>Yang Terjadi:</Text>
                        <Text fontSize="xs" color="green.200">• Revenue → RESET ke 0</Text>
                        <Text fontSize="xs" color="red.200">• Expense → RESET ke 0</Text>
                        <Text fontSize="xs" color="blue.200">• Retained Earnings → BERTAMBAH (Net Income)</Text>
                        <Text fontSize="xs">• Generate closing journal entries</Text>
                        <Text fontSize="xs">• Period HARD LOCK (sulit di-reopen)</Text>
                        <Text fontSize="xs" mt={2} fontWeight="semibold" color="orange.200">
                          ⚠️ PERMANENT ACTION!
                        </Text>
                        <Text fontSize="xs" fontStyle="italic" mt={1}>
                          Hanya dilakukan sekali setahun di akhir fiscal year.
                        </Text>
                      </Box>
                    }
                    placement="left"
                    hasArrow
                    bg="red.600"
                  >
                    <Box cursor="help">
                      <Icon as={FiInfo} boxSize={4} color="gray.500" _hover={{ color: 'red.500' }} />
                    </Box>
                  </Tooltip>
                </HStack>
              </CardHeader>
              <CardBody>
                <VStack spacing={4} alignItems="start">
                  <Alert status="error" borderRadius="md" size="sm">
                    <AlertIcon />
                    <Box>
                      <AlertTitle fontSize="sm">⚠️ Permanent Action</AlertTitle>
                      <AlertDescription fontSize="xs">
                        This will reset ALL revenue and expense accounts to zero and transfer
                        net income to retained earnings. Cannot be easily reversed!
                      </AlertDescription>
                    </Box>
                  </Alert>
                  
                  <Text fontSize="sm" color="gray.600">
                    Close fiscal year and generate automated closing entries.
                  </Text>
                  
                  <FormControl>
                    <HStack spacing={2}>
                      <FormLabel fontSize="sm" fontWeight="semibold" mb={0}>
                        Fiscal Year End Date
                      </FormLabel>
                      <Tooltip
                        label={
                          <Box p={2}>
                            <Text fontSize="xs" mb={1}>
                              Pilih tanggal akhir tahun fiscal Anda.
                            </Text>
                            <Text fontSize="xs">Contoh:</Text>
                            <Text fontSize="xs">• 31 Desember (calendar year)</Text>
                            <Text fontSize="xs">• 31 Maret (fiscal year Apr-Mar)</Text>
                            <Text fontSize="xs">• 30 Juni (fiscal year Jul-Jun)</Text>
                          </Box>
                        }
                        placement="top"
                        hasArrow
                      >
                        <Box cursor="help" display="inline-block">
                          <Icon as={FiInfo} boxSize={3} color="gray.400" />
                        </Box>
                      </Tooltip>
                    </HStack>
                    <Input
                      type="date"
                      value={fiscalYearEnd}
                      onChange={(e) => setFiscalYearEnd(e.target.value)}
                      variant="filled"
                      _hover={{ bg: 'gray.100' }}
                      _focus={{ bg: 'white', borderColor: 'blue.500' }}
                    />
                    <FormHelperText fontSize="xs">
                      Typically December 31 or end of your fiscal year
                    </FormHelperText>
                  </FormControl>
                  
                  <Tooltip
                    label={
                      <Box p={2}>
                        <Text fontSize="xs" fontWeight="bold" mb={1}>🔍 Preview Proses Closing</Text>
                        <Text fontSize="xs" mb={2}>
                          Lihat detail lengkap sebelum execute:
                        </Text>
                        <Text fontSize="xs">• Total Revenue yang akan direset</Text>
                        <Text fontSize="xs">• Total Expense yang akan direset</Text>
                        <Text fontSize="xs">• Net Income yang akan ditransfer</Text>
                        <Text fontSize="xs">• Preview closing journal entries</Text>
                        <Text fontSize="xs">• Validation checks</Text>
                        <Text fontSize="xs" mt={2} fontStyle="italic">
                          ⚠️ Execute hanya bisa dilakukan setelah preview!
                        </Text>
                      </Box>
                    }
                    placement="top"
                    hasArrow
                    isDisabled={!fiscalYearEnd}
                  >
                    <Button
                      colorScheme="red"
                      variant="solid"
                      width="full"
                      onClick={handlePreviewFiscalClosing}
                      isLoading={loadingFiscalPreview}
                      loadingText="Loading Preview..."
                      isDisabled={!fiscalYearEnd}
                    >
                      Preview Year-End Closing
                    </Button>
                  </Tooltip>
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
      
      {/* Close Period Confirmation Modal */}
      <Modal isOpen={!!closingPeriod} onClose={() => setClosingPeriod(null)}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Close Period</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4} alignItems="start">
              <Text>
                Are you sure you want to close{' '}
                <strong>
                  {closingPeriod && MONTHS[closingPeriod.month - 1]} {closingPeriod?.year}
                </strong>
                ?
              </Text>
              <Alert status="warning" borderRadius="md">
                <AlertIcon />
                <Box>
                  <AlertTitle fontSize="sm">After closing:</AlertTitle>
                  <AlertDescription fontSize="xs">
                    • Regular users cannot post new transactions
                    <br />
                    • Period will be locked for data entry
                    <br />
                    • You can reopen if needed
                  </AlertDescription>
                </Box>
              </Alert>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={() => setClosingPeriod(null)}>
              Cancel
            </Button>
            <Button
              colorScheme="red"
              onClick={() =>
                closingPeriod && handleClosePeriod(closingPeriod.year, closingPeriod.month)
              }
            >
              Close Period
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
      
      {/* Reopen Period Modal */}
      <Modal isOpen={!!reopenPeriod} onClose={() => { setReopenPeriod(null); setReopenReason(''); }}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Reopen Period</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4} alignItems="start">
              <Text>
                Reopen{' '}
                <strong>
                  {reopenPeriod && MONTHS[reopenPeriod.month - 1]} {reopenPeriod?.year}
                </strong>
                ?
              </Text>
              <FormControl isRequired>
                <FormLabel fontSize="sm">Reason for reopening</FormLabel>
                <Textarea
                  value={reopenReason}
                  onChange={(e) => setReopenReason(e.target.value)}
                  placeholder="Enter reason for reopening this period..."
                  rows={3}
                />
              </FormControl>
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button
              variant="ghost"
              mr={3}
              onClick={() => {
                setReopenPeriod(null);
                setReopenReason('');
              }}
            >
              Cancel
            </Button>
            <Button
              colorScheme="blue"
              onClick={handleReopenPeriod}
              isDisabled={!reopenReason.trim()}
            >
              Reopen Period
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
      
      {/* Fiscal Year-End Closing Preview Modal */}
      <Modal
        isOpen={showFiscalClosingModal}
        onClose={() => setShowFiscalClosingModal(false)}
        size="xl"
      >
        <ModalOverlay />
        <ModalContent maxW="800px">
          <ModalHeader bg="red.500" color="white">
            🏁 Fiscal Year-End Closing Preview
          </ModalHeader>
          <ModalCloseButton color="white" />
          <ModalBody py={6}>
            {fiscalClosingPreview && (
              <VStack spacing={4} alignItems="start">
                {/* Summary */}
                <Box width="full" p={4} bg="gray.50" borderRadius="md">
                  <Text fontSize="lg" fontWeight="bold" mb={2}>
                    Financial Summary
                  </Text>
                  <SimpleGrid columns={3} spacing={4}>
                    <Box>
                      <Text fontSize="xs" color="gray.600">Total Revenue</Text>
                      <Text fontSize="lg" fontWeight="bold" color="green.600">
                        Rp {fiscalClosingPreview.total_revenue.toLocaleString('id-ID')}
                      </Text>
                    </Box>
                    <Box>
                      <Text fontSize="xs" color="gray.600">Total Expense</Text>
                      <Text fontSize="lg" fontWeight="bold" color="red.600">
                        Rp {fiscalClosingPreview.total_expense.toLocaleString('id-ID')}
                      </Text>
                    </Box>
                    <Box>
                      <Text fontSize="xs" color="gray.600">Net Income</Text>
                      <Text fontSize="lg" fontWeight="bold" color="blue.600">
                        Rp {fiscalClosingPreview.net_income.toLocaleString('id-ID')}
                      </Text>
                    </Box>
                  </SimpleGrid>
                </Box>
                
                {/* Validation Messages */}
                {fiscalClosingPreview.validation_messages.length > 0 && (
                  <Alert status={fiscalClosingPreview.can_close ? 'warning' : 'error'} borderRadius="md">
                    <AlertIcon />
                    <Box>
                      <AlertTitle fontSize="sm">Validation Issues</AlertTitle>
                      <AlertDescription fontSize="xs">
                        {fiscalClosingPreview.validation_messages.map((msg, i) => (
                          <Text key={i}>• {msg}</Text>
                        ))}
                      </AlertDescription>
                    </Box>
                  </Alert>
                )}
                
                {/* What Will Happen */}
                {fiscalClosingPreview.can_close && (
                  <Alert status="info" borderRadius="md">
                    <AlertIcon />
                    <Box>
                      <AlertTitle fontSize="sm">What will happen:</AlertTitle>
                      <AlertDescription fontSize="xs">
                        • {fiscalClosingPreview.revenue_accounts.length} Revenue accounts → Reset to 0
                        <br />
                        • {fiscalClosingPreview.expense_accounts.length} Expense accounts → Reset to 0
                        <br />
                        • Retained Earnings → Increased by Rp {fiscalClosingPreview.net_income.toLocaleString('id-ID')}
                        <br />
                        • Automated journal entry created
                        <br />
                        • Period locked (cannot reopen easily)
                      </AlertDescription>
                    </Box>
                  </Alert>
                )}
                
                {/* Closing Entries Preview */}
                {fiscalClosingPreview.closing_entries && fiscalClosingPreview.closing_entries.length > 0 && (
                  <Box width="full">
                    <Text fontSize="sm" fontWeight="bold" mb={2}>
                      Closing Journal Entries:
                    </Text>
                    <VStack spacing={2} width="full" alignItems="start">
                      {fiscalClosingPreview.closing_entries.map((entry, i) => (
                        <Box
                          key={i}
                          width="full"
                          p={3}
                          bg="white"
                          border="1px"
                          borderColor="gray.200"
                          borderRadius="md"
                        >
                          <Text fontSize="xs" fontWeight="semibold" mb={1}>
                            {entry.description}
                          </Text>
                          <HStack spacing={2} fontSize="xs">
                            <Text>Dr: {entry.debit_account}</Text>
                            <Text>→</Text>
                            <Text>Cr: {entry.credit_account}</Text>
                            <Text fontWeight="bold">
                              Rp {entry.amount.toLocaleString('id-ID')}
                            </Text>
                          </HStack>
                        </Box>
                      ))}
                    </VStack>
                  </Box>
                )}
              </VStack>
            )}
          </ModalBody>
          <ModalFooter>
            <Button
              variant="ghost"
              mr={3}
              onClick={() => setShowFiscalClosingModal(false)}
              isDisabled={executingFiscalClosing}
            >
              Cancel
            </Button>
            <Button
              colorScheme="red"
              onClick={handleExecuteFiscalClosing}
              isLoading={executingFiscalClosing}
              loadingText="Closing..."
              isDisabled={!fiscalClosingPreview?.can_close || executingFiscalClosing}
            >
              Execute Year-End Closing
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

    </SimpleLayout>
  );
};

export default SettingsPage;
