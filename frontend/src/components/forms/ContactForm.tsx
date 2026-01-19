'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  FormControl,
  FormLabel,
  FormErrorMessage,
  Input,
  Select,
  Textarea,
  VStack,
  HStack,
  Grid,
  GridItem,
  Switch,
  useToast,
  Divider,
  Text,
} from '@chakra-ui/react';
import { Contact } from '@/types/contact';
import { useTranslation } from '@/hooks/useTranslation';

interface ContactFormProps {
  contact?: Contact | null;
  onSubmit: (contactData: Partial<Contact>) => Promise<void>;
  onCancel: () => void;
  isLoading?: boolean;
}

export default function ContactForm({ 
  contact = null, 
  onSubmit, 
  onCancel, 
  isLoading = false 
}: ContactFormProps) {
  const { t } = useTranslation();
  const toast = useToast();
  const [formData, setFormData] = useState<Partial<Contact>>({
    name: '',
    type: 'CUSTOMER',
    email: '',
    phone: '',
    is_active: true,
    pic_name: '',
    external_id: '',
    address: '',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  // Load contact data for editing
  useEffect(() => {
    if (contact) {
      setFormData({
        name: contact.name || '',
        type: contact.type || 'CUSTOMER',
        email: contact.email || '',
        phone: contact.phone || '',
        is_active: contact.is_active !== undefined ? contact.is_active : true,
        pic_name: contact.pic_name || '',
        external_id: contact.external_id || '',
        address: contact.address || '',
      });
    }
  }, [contact]);

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    // Required fields validation
    if (!formData.name?.trim()) {
      newErrors.name = t('contacts.validation.nameRequired');
    }

    if (!formData.type) {
      newErrors.type = t('contacts.validation.typeRequired');
    }

    // Email validation if provided
    if (formData.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = t('validation.email');
    }

    // Phone validation if provided
    if (formData.phone && formData.phone.length < 10) {
      newErrors.phone = t('contacts.validation.phoneMinLength');
    }


    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      toast({
        title: t('messages.toast.validationError'),
        description: t('messages.toast.validationErrorDesc'),
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    try {
      await onSubmit(formData);
    } catch (error) {
      console.error('Error submitting contact form:', error);
      toast({
        title: t('messages.toast.error'),
        description: error instanceof Error ? error.message : t('contacts.messages.saveFailed'),
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  const handleInputChange = (field: keyof Contact, value: any) => {
    setFormData(prev => {
      const newFormData = { ...prev, [field]: value };
      
      // Clear PIC name when type changes to EMPLOYEE
      if (field === 'type' && value === 'EMPLOYEE') {
        newFormData.pic_name = '';
      }
      
      return newFormData;
    });
    
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }));
    }
  };


  return (
    <Box as="form" onSubmit={handleSubmit}>
      <VStack spacing={6} align="stretch">
        {/* Basic Information */}
        <Box>
          <Text fontSize="lg" fontWeight="semibold" mb={4}>{t('contacts.form.basicInfo')}</Text>
          <Grid templateColumns={{ base: '1fr', md: '1fr 1fr' }} gap={4}>
            <GridItem>
              <FormControl isRequired isInvalid={!!errors.name}>
                <FormLabel>{t('contacts.contactName')}</FormLabel>
                <Input
                  value={formData.name || ''}
                  onChange={(e) => handleInputChange('name', e.target.value)}
                  placeholder={t('common.placeholders.enterName')}
                />
                <FormErrorMessage>{errors.name}</FormErrorMessage>
              </FormControl>
            </GridItem>

            <GridItem>
              <FormControl isRequired isInvalid={!!errors.type}>
                <FormLabel>{t('contacts.type')}</FormLabel>
                <Select
                  value={formData.type || 'CUSTOMER'}
                  onChange={(e) => handleInputChange('type', e.target.value)}
                >
                  <option value="CUSTOMER">{t('contacts.customer')}</option>
                  <option value="VENDOR">{t('contacts.vendor')}</option>
                  <option value="EMPLOYEE">{t('contacts.employee')}</option>
                </Select>
                <FormErrorMessage>{errors.type}</FormErrorMessage>
              </FormControl>
            </GridItem>

            <GridItem>
              <FormControl>
                <FormLabel>{t('contacts.externalId')}</FormLabel>
                <Input
                  value={formData.external_id || ''}
                  onChange={(e) => handleInputChange('external_id', e.target.value)}
                  placeholder={t('contacts.form.enterExternalId')}
                />
              </FormControl>
            </GridItem>

            {/* PIC Name - Only for CUSTOMER and VENDOR */}
            {formData.type !== 'EMPLOYEE' && (
              <GridItem>
                <FormControl>
                  <FormLabel>{t('contacts.picName')}</FormLabel>
                  <Input
                    value={formData.pic_name || ''}
                    onChange={(e) => handleInputChange('pic_name', e.target.value)}
                    placeholder={t('contacts.form.enterPicName')}
                  />
                </FormControl>
              </GridItem>
            )}
          </Grid>
        </Box>

        <Divider />

        {/* Contact Information */}
        <Box>
          <Text fontSize="lg" fontWeight="semibold" mb={4}>{t('contacts.form.contactInfo')}</Text>
          <Grid templateColumns={{ base: '1fr', md: '1fr 1fr' }} gap={4}>
            <GridItem>
              <FormControl isInvalid={!!errors.email}>
                <FormLabel>{t('contacts.email')}</FormLabel>
                <Input
                  type="email"
                  value={formData.email || ''}
                  onChange={(e) => handleInputChange('email', e.target.value)}
                  placeholder={t('common.placeholders.enterEmail')}
                />
                <FormErrorMessage>{errors.email}</FormErrorMessage>
              </FormControl>
            </GridItem>

            <GridItem>
              <FormControl isInvalid={!!errors.phone}>
                <FormLabel>{t('contacts.phone')}</FormLabel>
                <Input
                  value={formData.phone || ''}
                  onChange={(e) => handleInputChange('phone', e.target.value)}
                  placeholder={t('common.placeholders.enterPhone')}
                />
                <FormErrorMessage>{errors.phone}</FormErrorMessage>
              </FormControl>
            </GridItem>

            <GridItem colSpan={{ base: 1, md: 2 }}>
              <FormControl>
                <FormLabel>{t('contacts.address')}</FormLabel>
                <Textarea
                  value={formData.address || ''}
                  onChange={(e) => handleInputChange('address', e.target.value)}
                  placeholder={t('common.placeholders.enterAddress')}
                  rows={3}
                />
              </FormControl>
            </GridItem>
          </Grid>
        </Box>

        <Divider />

        {/* Status */}
        <FormControl display="flex" alignItems="center">
          <FormLabel htmlFor="is-active" mb="0">
            {t('contacts.form.activeStatus')}
          </FormLabel>
          <Switch
            id="is-active"
            isChecked={formData.is_active}
            onChange={(e) => handleInputChange('is_active', e.target.checked)}
            colorScheme="green"
          />
        </FormControl>

        {/* Form Actions */}
        <HStack spacing={4} pt={4}>
          <Button
            type="submit"
            colorScheme="blue"
            isLoading={isLoading}
            loadingText={contact ? t('common.updating') : t('common.creating')}
            flex={1}
          >
            {contact ? t('contacts.editContact') : t('contacts.createContact')}
          </Button>
          <Button
            variant="outline"
            onClick={onCancel}
            isDisabled={isLoading}
            flex={1}
          >
            {t('common.cancel')}
          </Button>
        </HStack>
      </VStack>
    </Box>
  );
}
