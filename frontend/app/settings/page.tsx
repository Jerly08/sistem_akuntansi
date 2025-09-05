'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useTranslation } from '@/hooks/useTranslation';
import SimpleLayout from '@/components/layout/SimpleLayout';
import SettingsEditModal from '@/components/settings/SettingsEditModal';
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
  Badge,
  Button,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Spinner,
  useColorModeValue,
  Divider,
  Select,
  useToast
} from '@chakra-ui/react';
import { FiHome, FiDollarSign, FiCalendar, FiSettings, FiEdit, FiGlobe } from 'react-icons/fi';

interface SystemSettings {
  id?: number;
  company_name: string;
  company_address: string;
  company_phone: string;
  company_email: string;
  company_website?: string;
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
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  
  // Move useColorModeValue to top level to fix hooks order
  const blueColor = useColorModeValue('blue.500', 'blue.300');
  const greenColor = useColorModeValue('green.500', 'green.300');
  const purpleColor = useColorModeValue('purple.500', 'purple.300');

  const fetchSettings = async () => {
    setLoading(true);
    try {
      const response = await api.get('/settings');
      if (response.data.success) {
        setSettings(response.data.data);
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
        await api.put('/settings', { language: newLanguage });
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

  const handleSettingsUpdate = (updatedSettings: SystemSettings) => {
    setSettings(updatedSettings);
    // If language was changed, update it in the UI
    if (updatedSettings.language && updatedSettings.language !== language) {
      setLanguage(updatedSettings.language);
    }
  };

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

  return (
<SimpleLayout allowedRoles={['admin']}>
      <Box>
        <VStack spacing={6} alignItems="start">
          <HStack justify="space-between" width="full">
            <Heading as="h1" size="xl">{t('settings.title')}</Heading>
            <Button
              colorScheme="blue"
              leftIcon={<FiEdit />}
              onClick={() => setIsEditModalOpen(true)}
            >
              {t('settings.editSettings')}
            </Button>
          </HStack>
          
          {error && (
            <Alert status="error" width="full">
              <AlertIcon />
              <AlertTitle>{t('common.error')}:</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          
          <SimpleGrid columns={[1, 1, 2]} spacing={6} width="full">
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
                  <Box>
                    <Text fontWeight="semibold" color="gray.600" fontSize="sm">{t('settings.companyName')}</Text>
                    <Text fontSize="md">{settings?.company_name}</Text>
                  </Box>
                  <Divider />
                  <Box>
                    <Text fontWeight="semibold" color="gray.600" fontSize="sm">{t('settings.address')}</Text>
                    <Text fontSize="md">{settings?.company_address}</Text>
                  </Box>
                  <Divider />
                  <Box>
                    <Text fontWeight="semibold" color="gray.600" fontSize="sm">{t('settings.phone')}</Text>
                    <Text fontSize="md">{settings?.company_phone}</Text>
                  </Box>
                  <Divider />
                  <Box>
                    <Text fontWeight="semibold" color="gray.600" fontSize="sm">{t('settings.email')}</Text>
                    <Text fontSize="md">{settings?.company_email}</Text>
                  </Box>
                  {settings?.tax_number && (
                    <>
                      <Divider />
                      <Box>
                        <Text fontWeight="semibold" color="gray.600" fontSize="sm">{t('settings.taxNumber')}</Text>
                        <Text fontSize="md">{settings?.tax_number}</Text>
                      </Box>
                    </>
                  )}
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
                  <Box>
                    <Text fontWeight="semibold" color="gray.600" fontSize="sm">{t('settings.currency')}</Text>
                    <Text fontSize="md">{settings?.currency}</Text>
                  </Box>
                  <Divider />
                  <Box>
                    <Text fontWeight="semibold" color="gray.600" fontSize="sm">{t('settings.dateFormat')}</Text>
                    <Text fontSize="md">{settings?.date_format}</Text>
                  </Box>
                  <Divider />
                  <Box>
                    <Text fontWeight="semibold" color="gray.600" fontSize="sm">{t('settings.fiscalYearStart')}</Text>
                    <HStack>
                      <Icon as={FiCalendar} boxSize={4} color="gray.500" />
                      <Text fontSize="md">{settings?.fiscal_year_start}</Text>
                    </HStack>
                  </Box>
                  <Divider />
                  <Box>
                    <Text fontWeight="semibold" color="gray.600" fontSize="sm">{t('settings.defaultTaxRate')}</Text>
                    <Text fontSize="md">{settings?.default_tax_rate}%</Text>
                  </Box>
                  <Divider />
                  <Box width="full">
                    <HStack mb={2}>
                      <Icon as={FiGlobe} boxSize={4} color="gray.600" />
                      <Text fontWeight="semibold" color="gray.600" fontSize="sm">
                        {t('settings.language')}
                      </Text>
                    </HStack>
                    <Select 
                      value={language} 
                      onChange={(e) => handleLanguageChange(e.target.value)}
                      size="md"
                      borderColor="gray.300"
                      _hover={{ borderColor: 'gray.400' }}
                      _focus={{ borderColor: purpleColor, boxShadow: `0 0 0 1px ${purpleColor}` }}
                    >
                      <option value="id">{t('settings.indonesian')}</option>
                      <option value="en">{t('settings.english')}</option>
                    </Select>
                  </Box>
                </VStack>
              </CardBody>
            </Card>
          </SimpleGrid>
        </VStack>
      </Box>

      {/* Settings Edit Modal */}
      <SettingsEditModal
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        settings={settings}
        onUpdate={handleSettingsUpdate}
      />
    </SimpleLayout>
  );
};

export default SettingsPage;
