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
  Stack
} from '@chakra-ui/react';
import { FiHome, FiSettings, FiGlobe, FiCalendar, FiDollarSign, FiSave, FiX } from 'react-icons/fi';

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
  invoice_next_number: number;
  quote_prefix: string;
  quote_next_number: number;
  purchase_prefix: string;
  purchase_next_number: number;
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
  
  // Move useColorModeValue to top level to fix hooks order
  const blueColor = useColorModeValue('blue.500', 'blue.300');
  const greenColor = useColorModeValue('green.500', 'green.300');
  const purpleColor = useColorModeValue('purple.500', 'purple.300');

  const fetchSettings = async () => {
    setLoading(true);
    try {
      const response = await api.get(API_ENDPOINTS.SETTINGS);
      if (response.data.success) {
        setSettings(response.data.data);
        setFormData(response.data.data);
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
  }, []);

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
      const resp = await api.post('/settings/company/logo', fd, {
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
          
          <SimpleGrid columns={[1, 1, 2, 3]} spacing={6} width="full">
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
                    <Input
                      type="date"
                      value={formData?.fiscal_year_start || settings?.fiscal_year_start || ''}
                      onChange={(e) => handleFormChange('fiscal_year_start', e.target.value)}
                      variant="filled"
                      _hover={{ bg: 'gray.100' }}
                      _focus={{ bg: 'white', borderColor: 'blue.500' }}
                    />
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
            
            {/* Invoice Settings Card */}
            <Card border="1px" borderColor="gray.200" boxShadow="md">
              <CardHeader>
                <HStack spacing={3}>
                  <Icon as={FiSettings} boxSize={6} color={purpleColor} />
                  <Heading size="md">{t('settings.invoiceSettings')}</Heading>
                </HStack>
              </CardHeader>
              <CardBody>
                <VStack spacing={4} alignItems="start">
                  <FormControl>
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      {t('settings.invoicePrefix')}
                    </FormLabel>
                    <Input
                      value={formData?.invoice_prefix || settings?.invoice_prefix || ''}
                      onChange={(e) => handleFormChange('invoice_prefix', e.target.value)}
                      placeholder="e.g. INV"
                      variant="filled"
                      _hover={{ bg: 'gray.100' }}
                      _focus={{ bg: 'white', borderColor: 'blue.500' }}
                    />
                  </FormControl>
                  <Divider />
                  
                  <FormControl>
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      {t('settings.invoiceNextNumber')}
                    </FormLabel>
                    <NumberInput
                      value={formData?.invoice_next_number ?? settings?.invoice_next_number ?? 1}
                      onChange={(valueString) => handleFormChange('invoice_next_number', parseInt(valueString) || 1)}
                      min={1}
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
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      {t('settings.quotePrefix')}
                    </FormLabel>
                    <Input
                      value={formData?.quote_prefix || settings?.quote_prefix || ''}
                      onChange={(e) => handleFormChange('quote_prefix', e.target.value)}
                      placeholder="e.g. QUO"
                      variant="filled"
                      _hover={{ bg: 'gray.100' }}
                      _focus={{ bg: 'white', borderColor: 'blue.500' }}
                    />
                  </FormControl>
                  <Divider />
                  
                  <FormControl>
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      {t('settings.quoteNextNumber')}
                    </FormLabel>
                    <NumberInput
                      value={formData?.quote_next_number ?? settings?.quote_next_number ?? 1}
                      onChange={(valueString) => handleFormChange('quote_next_number', parseInt(valueString) || 1)}
                      min={1}
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
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      {t('settings.purchasePrefix')}
                    </FormLabel>
                    <Input
                      value={formData?.purchase_prefix || settings?.purchase_prefix || ''}
                      onChange={(e) => handleFormChange('purchase_prefix', e.target.value)}
                      placeholder="e.g. PUR"
                      variant="filled"
                      _hover={{ bg: 'gray.100' }}
                      _focus={{ bg: 'white', borderColor: 'blue.500' }}
                    />
                  </FormControl>
                  <Divider />
                  
                  <FormControl>
                    <FormLabel fontWeight="semibold" color="gray.600" fontSize="sm">
                      {t('settings.purchaseNextNumber')}
                    </FormLabel>
                    <NumberInput
                      value={formData?.purchase_next_number ?? settings?.purchase_next_number ?? 1}
                      onChange={(valueString) => handleFormChange('purchase_next_number', parseInt(valueString) || 1)}
                      min={1}
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
