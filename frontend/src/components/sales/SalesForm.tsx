'use client';

import React, { useState, useEffect, useRef } from 'react';
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
  VStack,
  HStack,
  Box,
  Divider,
  Text,
  IconButton,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
  useToast,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  Switch,
  Badge,
  Flex,
  Card,
  CardHeader,
  CardBody,
  Heading,
  Alert,
  AlertIcon,
  AlertDescription,
  Icon,
  useColorModeValue
} from '@chakra-ui/react';
import CurrencyInput from '@/components/common/CurrencyInput';
import { useForm, useFieldArray } from 'react-hook-form';
import { FiPlus, FiTrash2, FiSave, FiX, FiDollarSign, FiShoppingCart, FiFileText } from 'react-icons/fi';
import salesService, { 
  Sale, 
  SaleCreateRequest, 
  SaleUpdateRequest, 
  SaleItemRequest,
  SaleItemUpdateRequest 
} from '@/services/salesService';
import ErrorHandler from '@/utils/errorHandler';
import { useAuth } from '@/contexts/AuthContext';

interface SalesFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: () => void;
  sale?: Sale | null;
}

interface FormData {
  customer_id: number;
  sales_person_id?: number;
  type: string;
  date: string;
  due_date?: string;
  valid_until?: string;
  currency: string;
  exchange_rate: number;
  discount_percent: number;
  ppn_percent: number;
  pph_percent: number;
  pph_type?: string;
  payment_terms: string;
  payment_method?: string;
  shipping_method?: string;
  shipping_cost: number;
  billing_address?: string;
  shipping_address?: string;
  notes?: string;
  internal_notes?: string;
  reference?: string;
  items: Array<{
    id?: number;
    product_id: number;
    description: string;
    quantity: number;
    unit_price: number;
    discount_percent: number;
    taxable: boolean;
    revenue_account_id?: number;
    delete?: boolean;
  }>;
}

const SalesForm: React.FC<SalesFormProps> = ({
  isOpen,
  onClose,
  onSave,
  sale
}) => {
  const [loading, setLoading] = useState(false);
  const [customers, setCustomers] = useState<any[]>([]);
  const [products, setProducts] = useState<any[]>([]);
  const [salesPersons, setSalesPersons] = useState<any[]>([]);
  const [accounts, setAccounts] = useState<any[]>([]);
  const [loadingData, setLoadingData] = useState(true);
  const toast = useToast();
  const { user, token } = useAuth();
  const modalBodyRef = useRef<HTMLDivElement>(null);
  
  // Color mode values for dark mode support
  const modalBg = useColorModeValue('white', 'gray.800');
  const headerBg = useColorModeValue('blue.50', 'gray.700');
  const headingColor = useColorModeValue('blue.700', 'blue.300');
  const subHeadingColor = useColorModeValue('gray.600', 'gray.300');
  const textColor = useColorModeValue('gray.600', 'gray.400');
  const inputBg = useColorModeValue('gray.50', 'gray.600');
  const inputFocusBg = useColorModeValue('white', 'gray.500');
  const tableBg = useColorModeValue('white', 'gray.700');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const footerBg = useColorModeValue('white', 'gray.800');
  const shadowColor = useColorModeValue('rgba(0, 0, 0, 0.1)', 'rgba(0, 0, 0, 0.3)');
  const alertBg = useColorModeValue('blue.50', 'blue.900');
  const alertBorderColor = useColorModeValue('blue.200', 'blue.700');
  const scrollTrackBg = useColorModeValue('#f7fafc', '#2d3748');
  const scrollThumbBg = useColorModeValue('#cbd5e0', '#4a5568');
  const scrollThumbHoverBg = useColorModeValue('#a0aec0', '#718096');
  
  // Check if user has permission to create/edit sales - using lowercase for consistency
  const userRole = user?.role?.toLowerCase();
  const canCreateSales = userRole === 'finance' || userRole === 'director' || userRole === 'admin';
  const canEditSales = userRole === 'admin' || userRole === 'finance' || userRole === 'director';
  
  // For new sales, check create permission; for editing, check edit permission
  const hasPermission = sale ? canEditSales : canCreateSales;
  
  // If modal is opened but user doesn't have permission, close it and show error
  useEffect(() => {
    if (isOpen && user && !hasPermission) {
      const action = sale ? 'edit' : 'create';
      toast({
        title: 'Access Denied',
        description: `You do not have permission to ${action} sales. Contact your administrator for access.`,
        status: 'error',
        duration: 5000,
      });
      onClose();
    }
  }, [isOpen, user, hasPermission, sale, toast, onClose]);
  
  // Don't render the form if user doesn't have permission
  if (!hasPermission && user) {
    return null;
  }

  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    control,
    formState: { errors }
  } = useForm<FormData>({
    defaultValues: {
      type: 'INVOICE',
      currency: 'IDR',
      exchange_rate: 1,
      discount_percent: 0,
      ppn_percent: 11,
      pph_percent: 0,
      payment_terms: 'NET_30',
      shipping_cost: 0,
      items: [
        {
          product_id: 0,
          description: '',
          quantity: 1,
          unit_price: 0,
          discount_percent: 0,
          taxable: true
        }
      ]
    }
  });

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'items'
  });

  const watchItems = watch('items');
  const watchDiscountPercent = watch('discount_percent');
  const watchPPNPercent = watch('ppn_percent');
  const watchShippingCost = watch('shipping_cost');

  useEffect(() => {
    if (isOpen) {
      loadFormData();
      if (sale) {
        populateFormWithSale(sale);
      } else {
        resetForm();
      }
    }
  }, [isOpen, sale]);

  // Ensure modal body is scrollable when content overflows
  useEffect(() => {
    if (isOpen) {
      // Small delay to ensure DOM is ready
      setTimeout(() => {
        const modalBody = modalBodyRef.current;
        if (modalBody) {
          const isOverflowing = modalBody.scrollHeight > modalBody.clientHeight;
          if (isOverflowing) {
            modalBody.style.overflowY = 'scroll';
          }
        }
      }, 100);
    }
  }, [isOpen, fields.length]);

  const loadFormData = async () => {
    if (!token) {
      toast({
        title: 'Authentication Required',
        description: 'Please login to access this feature.',
        status: 'error',
        duration: 5000,
        isClosable: true
      });
      return;
    }

    setLoadingData(true);
    
    try {
      console.log('SalesForm: Starting to load form data...');
      
      // Load all data concurrently with proper error handling
      const [customersResult, productsResult, salesPersonsResult, accountsResult] = await Promise.allSettled([
        // Load customers
        (async () => {
          console.log('SalesForm: Loading customers...');
          const contactService = await import('@/services/contactService');
          const result = await contactService.default.getContacts(token, 'CUSTOMER');
          console.log('SalesForm: Customers loaded:', result?.length || 0);
          return result;
        })(),
        
        // Load products with fallback for permission errors
        (async () => {
          try {
            console.log('SalesForm: Loading products with token...');
            const productService = await import('@/services/productService');
            const result = await productService.default.getProducts({}, token);
            console.log('SalesForm: Products loaded:', result?.data?.length || 0);
            return result;
          } catch (error: any) {
            console.warn('SalesForm: Failed to load products, using empty list:', error?.message || error);
            // Return empty result for any error
            return { data: [] }; // Empty products array
          }
        })(),
        
        // Load sales persons from contacts (employees)
        (async () => {
          console.log('SalesForm: Loading sales persons (employees)...');
          const contactService = await import('@/services/contactService');
          const result = await contactService.default.getContacts(token, 'EMPLOYEE');
          console.log('SalesForm: Sales persons loaded:', result?.length || 0);
          return result;
        })(),
        
        // Load revenue accounts
        (async () => {
          console.log('SalesForm: Loading revenue accounts...');
          const accountService = await import('@/services/accountService');
          const result = await accountService.default.getAccounts(token, 'REVENUE');
          console.log('SalesForm: Revenue accounts loaded:', result?.length || 0);
          return result;
        })()
      ]);

      // Process customers
      if (customersResult.status === 'fulfilled' && Array.isArray(customersResult.value)) {
        setCustomers(customersResult.value);
      } else {
        console.warn('Failed to load customers:', customersResult.status === 'rejected' ? customersResult.reason : 'No data');
        setCustomers([]);
      }

      // Process products
      if (productsResult.status === 'fulfilled' && productsResult.value?.data && Array.isArray(productsResult.value.data)) {
        setProducts(productsResult.value.data);
      } else {
        console.warn('Failed to load products:', productsResult.status === 'rejected' ? productsResult.reason : 'No data');
        setProducts([]);
      }

      // Process sales persons
      if (salesPersonsResult.status === 'fulfilled' && Array.isArray(salesPersonsResult.value)) {
        const salesPersonsData = salesPersonsResult.value.map(contact => ({
          ...contact,
          name: contact.name || contact.company_name || 'Unknown Employee'
        }));
        setSalesPersons(salesPersonsData);
      } else {
        console.warn('Failed to load sales persons:', salesPersonsResult.status === 'rejected' ? salesPersonsResult.reason : 'No data');
        setSalesPersons([]);
      }

      // Process accounts
      if (accountsResult.status === 'fulfilled' && Array.isArray(accountsResult.value)) {
        setAccounts(accountsResult.value);
      } else {
        console.warn('Failed to load accounts:', accountsResult.status === 'rejected' ? accountsResult.reason : 'No data');
        setAccounts([]);
      }

    } catch (error: any) {
      console.error('Error loading form data:', error);
      toast({
        title: 'Loading Error',
        description: 'Failed to load form data. Please try again.',
        status: 'error',
        duration: 5000,
        isClosable: true
      });
    } finally {
      setLoadingData(false);
    }
  };

  const populateFormWithSale = (saleData: Sale) => {
    reset({
      customer_id: saleData.customer_id,
      sales_person_id: saleData.sales_person_id,
      type: saleData.type,
      date: saleData.date.split('T')[0],
      due_date: saleData.due_date ? saleData.due_date.split('T')[0] : undefined,
      valid_until: saleData.valid_until ? saleData.valid_until.split('T')[0] : undefined,
      currency: saleData.currency,
      exchange_rate: saleData.exchange_rate,
      discount_percent: saleData.discount_percent,
      ppn_percent: saleData.ppn_percent,
      pph_percent: saleData.pph_percent,
      pph_type: saleData.pph_type,
      payment_terms: saleData.payment_terms,
      payment_method: saleData.payment_method,
      shipping_method: saleData.shipping_method,
      shipping_cost: saleData.shipping_cost,
      billing_address: saleData.billing_address,
      shipping_address: saleData.shipping_address,
      notes: saleData.notes,
      internal_notes: saleData.internal_notes,
      reference: saleData.reference,
      items: saleData.sale_items?.map(item => ({
        id: item.id,
        product_id: item.product_id,
        description: item.description || '',
        quantity: item.quantity,
        unit_price: item.unit_price,
        discount_percent: item.discount_percent,
        taxable: item.taxable,
        revenue_account_id: item.revenue_account_id
      })) || []
    });
  };


  const resetForm = () => {
    reset({
      type: 'INVOICE',
      date: new Date().toISOString().split('T')[0],
      currency: 'IDR',
      exchange_rate: 1,
      discount_percent: 0,
      ppn_percent: 11,
      pph_percent: 0,
      payment_terms: 'NET_30',
      shipping_cost: 0,
      items: [
        {
          product_id: 0,
          description: '',
          quantity: 1,
          unit_price: 0,
          discount_percent: 0,
          taxable: true
        }
      ]
    });
  };

  const handleProductChange = (index: number, productId: number) => {
    const product = products.find(p => p.id === parseInt(productId.toString()));
    if (product) {
      setValue(`items.${index}.product_id`, product.id);
      setValue(`items.${index}.description`, product.name);
      setValue(`items.${index}.unit_price`, product.price);
    }
  };

  const calculateLineTotal = (item: any) => {
    const subtotal = item.quantity * item.unit_price;
    const discountAmount = subtotal * (item.discount_percent / 100);
    return subtotal - discountAmount;
  };

  const calculateSubtotal = () => {
    return watchItems.reduce((sum, item) => sum + calculateLineTotal(item), 0);
  };

  const calculateTotal = () => {
    const subtotal = calculateSubtotal();
    const globalDiscount = subtotal * (watchDiscountPercent / 100);
    const afterDiscount = subtotal - globalDiscount;
    const withShipping = afterDiscount + watchShippingCost;
    const ppn = withShipping * (watchPPNPercent / 100);
    return withShipping + ppn;
  };

  const addItem = () => {
    append({
      product_id: 0,
      description: '',
      quantity: 1,
      unit_price: 0,
      discount_percent: 0,
      taxable: true
    });
  };

  const removeItem = (index: number) => {
    if (fields.length > 1) {
      remove(index);
    }
  };

  const onSubmit = async (data: FormData) => {
    try {
      setLoading(true);

      // Validate items - allow items without product_id if products are not available
      const validItems = data.items.filter(item => {
        // If products are available, require product_id
        if (products.length > 0) {
          return item.product_id > 0;
        }
        // If no products available, just check for description and price
        return item.description && item.description.trim() !== '' && item.unit_price > 0;
      });
      
      if (validItems.length === 0) {
        const errorMsg = products.length > 0 
          ? 'At least one item with a selected product is required'
          : 'At least one item with description and price is required';
        ErrorHandler.handleValidationError([errorMsg], toast, 'sales form');
        return;
      }

      // Additional validation
      const validationErrors = salesService.validateSaleData({
        ...data,
        items: validItems.map(item => ({
          product_id: item.product_id,
          description: item.description || '',
          quantity: Math.max(1, Math.floor(item.quantity || 1)), // Ensure positive integer
          unit_price: Math.min(999999999999.99, Math.max(0, item.unit_price || 0)), // Cap to prevent overflow
          discount: Math.min(999999.99, Math.max(0, item.discount_percent || 0)), // Legacy field as flat amount
          discount_percent: Math.min(100, Math.max(0, item.discount_percent || 0)), // New field as percentage
          tax: 0, // Tax will be calculated by backend based on taxable flag
          taxable: item.taxable !== false, // Default to true if not specified
          revenue_account_id: item.revenue_account_id || 0
        }))
      });

      if (validationErrors.length > 0) {
        ErrorHandler.handleValidationError(validationErrors, toast, 'sales form');
        return;
      }

      if (sale) {
      // Update existing sale
      const updateData: SaleUpdateRequest = {
        customer_id: data.customer_id,
        sales_person_id: data.sales_person_id,
        date: data.date ? `${data.date}T00:00:00Z` : undefined, // Convert to ISO datetime format
        due_date: data.due_date ? `${data.due_date}T00:00:00Z` : undefined,
        valid_until: data.valid_until ? `${data.valid_until}T00:00:00Z` : undefined,
          discount_percent: data.discount_percent,
          ppn_percent: data.ppn_percent,
          pph_percent: data.pph_percent,
          pph_type: data.pph_type,
          payment_terms: data.payment_terms,
          payment_method: data.payment_method,
          shipping_method: data.shipping_method,
          shipping_cost: data.shipping_cost,
          billing_address: data.billing_address,
          shipping_address: data.shipping_address,
          notes: data.notes,
          internal_notes: data.internal_notes,
          reference: data.reference,
          items: validItems.map(item => ({
            id: item.id,
            product_id: item.product_id,
            description: item.description,
            quantity: item.quantity,
            unit_price: item.unit_price,
            discount: item.discount_percent || item.discount || 0, // Map to backend field
            tax: 0, // Tax will be calculated by backend based on taxable flag
            revenue_account_id: item.revenue_account_id,
            delete: item.delete || false
          }))
        };

        await salesService.updateSale(sale.id, updateData);
        ErrorHandler.handleSuccess('Sale has been updated successfully', toast, 'update sale');
      } else {
        // Create new sale
        const createData: SaleCreateRequest = {
          customer_id: data.customer_id,
          sales_person_id: data.sales_person_id,
          type: data.type,
          date: `${data.date}T00:00:00Z`, // Convert to ISO datetime format for Go backend
          due_date: data.due_date ? `${data.due_date}T00:00:00Z` : undefined,
          valid_until: data.valid_until ? `${data.valid_until}T00:00:00Z` : undefined,
          currency: data.currency,
          exchange_rate: data.exchange_rate,
          discount_percent: data.discount_percent,
          ppn_percent: data.ppn_percent,
          pph_percent: data.pph_percent,
          pph_type: data.pph_type,
          payment_terms: data.payment_terms,
          payment_method: data.payment_method,
          shipping_method: data.shipping_method,
          shipping_cost: data.shipping_cost,
          billing_address: data.billing_address,
          shipping_address: data.shipping_address,
          notes: data.notes,
          internal_notes: data.internal_notes,
          reference: data.reference,
          items: validItems.map(item => ({
            product_id: item.product_id,
            description: item.description || '',
            quantity: Math.max(1, Math.floor(item.quantity || 1)), // Ensure positive integer
            unit_price: Math.min(999999999999.99, Math.max(0, item.unit_price || 0)), // Cap to prevent overflow
            discount: Math.min(999999.99, Math.max(0, item.discount_percent || 0)), // Legacy field as flat amount
            discount_percent: Math.min(100, Math.max(0, item.discount_percent || 0)), // New field as percentage
            tax: 0, // Tax will be calculated by backend based on taxable flag
            taxable: item.taxable !== false, // Default to true if not specified
            revenue_account_id: item.revenue_account_id || 0
          }))
        };

        await salesService.createSale(createData);
        ErrorHandler.handleSuccess('Sale has been created successfully', toast, 'create sale');
      }

      onSave();
      onClose();
    } catch (error: any) {
      ErrorHandler.handleSaveError('sale', error, toast, !!sale);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  return (
    <Modal 
      isOpen={isOpen} 
      onClose={handleClose} 
      size="6xl" 
      isCentered
      closeOnOverlayClick={false}
      scrollBehavior="inside"
      motionPreset="slideInBottom"
      blockScrollOnMount={false}
    >
      <ModalOverlay 
        bg="blackAlpha.700" 
        backdropFilter="blur(4px)"
        onWheel={(e) => e.stopPropagation()}
      />
      <ModalContent 
        maxH="95vh" 
        minH="80vh"
        mx={4} 
        my={2} 
        borderRadius="xl"
        bg={modalBg}
        shadow="2xl"
        overflow="hidden"
        display="flex"
        flexDirection="column"
        w="full"
        maxW="6xl"
      >
        <ModalHeader 
          bg={headerBg} 
          borderBottomWidth={1} 
          borderColor={borderColor}
          pb={4}
          pt={6}
        >
          <HStack justify="space-between" align="center">
            <Box>
              <Heading size="lg" color={headingColor}>
                {sale ? 'Edit Sale Transaction' : 'Create New Sale'}
              </Heading>
              <Text color={textColor} fontSize="sm" mt={1}>
                {sale ? 'Modify existing sale details and items' : 'Create a new sales transaction with items and pricing'}
              </Text>
            </Box>
            <Badge colorScheme="blue" variant="solid" px={3} py={1} borderRadius="md">
              <Icon as={FiShoppingCart} mr={1} />
              Sale Form
            </Badge>
          </HStack>
        </ModalHeader>
        <ModalCloseButton />

        <form onSubmit={handleSubmit(onSubmit)} style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
          {/* Hidden type field with default value */}
          <input type="hidden" {...register('type')} value="INVOICE" />
          
          <ModalBody 
            ref={modalBodyRef}
            flex="1" 
            overflowY="auto" 
            px={6} 
            py={4}
            pb={8}
            maxH="calc(95vh - 200px)"
            minH="400px"
            sx={{
              // Enable mouse wheel scrolling
              overscrollBehavior: 'contain',
              WebkitOverflowScrolling: 'touch',
              scrollBehavior: 'smooth',
              
              // Force scrollbar to be visible
              overflowY: 'scroll !important',
              
              // Custom scrollbar styles
              '&::-webkit-scrollbar': {
                width: '8px',
                display: 'block',
              },
              '&::-webkit-scrollbar-track': {
                background: scrollTrackBg,
                borderRadius: '4px',
              },
              '&::-webkit-scrollbar-thumb': {
                background: scrollThumbBg,
                borderRadius: '4px',
                '&:hover': {
                  background: scrollThumbHoverBg,
                },
              },
            }}
            onWheel={(e) => {
              // Allow natural scroll behavior
              const target = e.currentTarget;
              const isScrollable = target.scrollHeight > target.clientHeight;
              if (isScrollable) {
                // Let the natural scroll happen
                return;
              }
              e.preventDefault();
              e.stopPropagation();
            }}
          >
            <VStack spacing={6} align="stretch">
              {/* Basic Information */}
              <Box>
                <Heading size="md" mb={4} color={subHeadingColor}>
                  📋 Basic Information
                </Heading>
                <VStack spacing={4}>
                  <HStack w="full" spacing={4}>
                      <FormControl isRequired isInvalid={!!errors.customer_id}>
                        <FormLabel>Customer</FormLabel>
                        <Select
                          {...register('customer_id', {
                            required: 'Customer is required',
                            setValueAs: value => parseInt(value) || 0
                          })}
                          bg={inputBg}
                          _focus={{ bg: inputFocusBg }}
                          isDisabled={loadingData}
                        >
                          <option value="">
                            {loadingData ? 'Loading customers...' : 
                             customers.length === 0 ? 'No customers available' : 'Select customer'}
                          </option>
                          {customers.map(customer => (
                            <option key={customer.id} value={customer.id}>
                              {customer.code} - {customer.name}
                            </option>
                          ))}
                        </Select>
                        <FormErrorMessage>{errors.customer_id?.message}</FormErrorMessage>
                      </FormControl>

                      <FormControl>
                        <FormLabel>Sales Person</FormLabel>
                        <Select
                          {...register('sales_person_id', {
                            setValueAs: value => value ? parseInt(value) : undefined
                          })}
                          bg={inputBg}
                          _focus={{ bg: inputFocusBg }}
                          isDisabled={loadingData}
                        >
                          <option value="">
                            {loadingData ? 'Loading sales persons...' : 
                             salesPersons.length === 0 ? 'No sales persons available' : 'Select sales person'}
                          </option>
                          {salesPersons.map(person => (
                            <option key={person.id} value={person.id}>
                              {person.name}
                            </option>
                          ))}
                        </Select>
                      </FormControl>
                  </HStack>
                  
                  <HStack w="full" spacing={4}>
                      <FormControl isRequired isInvalid={!!errors.date}>
                        <FormLabel>Date</FormLabel>
                        <Input
                          type="date"
                          {...register('date', {
                            required: 'Date is required'
                          })}
                          bg={inputBg}
                          _focus={{ bg: inputFocusBg }}
                        />
                        <FormErrorMessage>{errors.date?.message}</FormErrorMessage>
                      </FormControl>

                      <FormControl>
                        <FormLabel>Due Date</FormLabel>
                        <Input
                          type="date"
                          {...register('due_date')}
                          bg={inputBg}
                          _focus={{ bg: inputFocusBg }}
                        />
                      </FormControl>

                      <FormControl>
                        <FormLabel>Valid Until</FormLabel>
                        <Input
                          type="date"
                          {...register('valid_until')}
                          bg={inputBg}
                          _focus={{ bg: inputFocusBg }}
                        />
                      </FormControl>
                  </HStack>
                </VStack>
              </Box>

              <Divider />

              {/* Items Section */}
              <Box>
                <Flex justify="space-between" align="center" mb={4}>
                  <Heading size="md" color={subHeadingColor}>
                    🛍️ Sale Items
                  </Heading>
                  <Button
                    size="sm"
                    colorScheme="blue"
                    leftIcon={<FiPlus />}
                    onClick={addItem}
                  >
                    Add Item
                  </Button>
                </Flex>

                {/* Products warning if empty */}
                {!loadingData && products.length === 0 && (
                  <Alert status="warning" mb={4} borderRadius="md" bg={alertBg} borderColor={alertBorderColor}>
                    <AlertIcon />
                    <AlertDescription fontSize="sm">
                      Products are not available. You can still create sales by manually entering product information in the description field.
                    </AlertDescription>
                  </Alert>
                )}

                <Box 
                  overflowX="auto" 
                  border="1px" 
                  borderColor={borderColor} 
                  borderRadius="md"
                  bg={tableBg}
                  shadow="sm"
                  css={{
                    '&::-webkit-scrollbar': {
                      height: '8px',
                    },
                    '&::-webkit-scrollbar-track': {
                      background: scrollTrackBg,
                      borderRadius: '4px',
                    },
                    '&::-webkit-scrollbar-thumb': {
                      background: scrollThumbBg,
                      borderRadius: '4px',
                    },
                    '&::-webkit-scrollbar-thumb:hover': {
                      background: scrollThumbHoverBg,
                    },
                  }}
                >
                  <Table variant="simple" size="sm">
                    <Thead>
                      <Tr>
                        <Th>Product</Th>
                        <Th>Description</Th>
                        <Th>Qty</Th>
                        <Th>Unit Price</Th>
                        <Th>Discount %</Th>
                        <Th>Taxable</Th>
                        <Th>Total</Th>
                        <Th>Action</Th>
                      </Tr>
                    </Thead>
                    <Tbody>
                      {fields.map((field, index) => (
                        <Tr key={field.id}>
                          <Td>
                            <Select
                              size="sm"
                              {...register(`items.${index}.product_id`, {
                                required: 'Product is required',
                                setValueAs: value => parseInt(value) || 0
                              })}
                              onChange={(e) => handleProductChange(index, parseInt(e.target.value))}
                              bg={inputBg}
                              _focus={{ bg: inputFocusBg }}
                            >
                              <option value="">Select product</option>
                              {products.map(product => (
                                <option key={product.id} value={product.id}>
                                  {product.code} - {product.name}
                                </option>
                              ))}
                            </Select>
                          </Td>
                          <Td>
                            <Input
                              size="sm"
                              {...register(`items.${index}.description`)}
                              placeholder="Item description"
                              bg={inputBg}
                              _focus={{ bg: inputFocusBg }}
                            />
                          </Td>
                          <Td>
                            <NumberInput size="sm" min={1}>
                              <NumberInputField
                                {...register(`items.${index}.quantity`, {
                                  required: 'Quantity is required',
                                  min: 1,
                                  setValueAs: value => parseInt(value) || 1
                                })}
                              />
                              <NumberInputStepper>
                                <NumberIncrementStepper />
                                <NumberDecrementStepper />
                              </NumberInputStepper>
                            </NumberInput>
                          </Td>
                          <Td>
                            <CurrencyInput
                              value={watchItems[index]?.unit_price || 0}
                              onChange={(value) => setValue(`items.${index}.unit_price`, value)}
                              placeholder="Rp 10.000"
                              size="sm"
                              min={0}
                              showLabel={false}
                            />
                          </Td>
                          <Td>
                            <NumberInput size="sm" min={0} max={100}>
                              <NumberInputField
                                {...register(`items.${index}.discount_percent`, {
                                  setValueAs: value => parseFloat(value) || 0
                                })}
                              />
                            </NumberInput>
                          </Td>
                          <Td>
                            <Switch
                              size="sm"
                              {...register(`items.${index}.taxable`)}
                            />
                          </Td>
                          <Td>
                            <Text fontSize="sm" fontWeight="medium">
                              {salesService.formatCurrency(calculateLineTotal(watchItems[index] || {}))}
                            </Text>
                          </Td>
                          <Td>
                            <IconButton
                              size="sm"
                              colorScheme="red"
                              variant="ghost"
                              icon={<FiTrash2 />}
                              onClick={() => removeItem(index)}
                              isDisabled={fields.length === 1}
                              aria-label="Remove item"
                            />
                          </Td>
                        </Tr>
                      ))}
                    </Tbody>
                  </Table>
                </Box>
              </Box>

              <Divider />

              {/* Pricing & Taxes */}
              <Box>
                <Heading size="md" mb={4} color={subHeadingColor}>
                  💰 Pricing & Taxes
                </Heading>
                <HStack w="full" spacing={4}>
                    <FormControl>
                      <FormLabel>Global Discount (%)</FormLabel>
                      <NumberInput min={0} max={100}>
                        <NumberInputField
                          {...register('discount_percent', {
                            setValueAs: value => parseFloat(value) || 0
                          })}
                          bg={inputBg}
                          _focus={{ bg: inputFocusBg }}
                        />
                      </NumberInput>
                    </FormControl>

                    <FormControl>
                      <FormLabel>PPN (%)</FormLabel>
                      <NumberInput min={0} max={100}>
                        <NumberInputField
                          {...register('ppn_percent', {
                            setValueAs: value => parseFloat(value) || 0
                          })}
                          bg={inputBg}
                          _focus={{ bg: inputFocusBg }}
                        />
                      </NumberInput>
                    </FormControl>

                    <FormControl>
                      <FormLabel>Shipping Cost</FormLabel>
                      <NumberInput min={0}>
                        <NumberInputField
                          {...register('shipping_cost', {
                            setValueAs: value => parseFloat(value) || 0
                          })}
                          bg={inputBg}
                          _focus={{ bg: inputFocusBg }}
                        />
                      </NumberInput>
                    </FormControl>
                </HStack>

                
                {/* Total Calculation Alert */}
                {calculateSubtotal() > 0 && (
                  <Alert status="info" borderRadius="lg" mt={4} bg={alertBg} borderColor={alertBorderColor}>
                    <AlertIcon color="blue.500" />
                    <AlertDescription fontSize="sm">
                      <VStack align="stretch" spacing={3} w="full">
                        <HStack justify="space-between">
                          <Text color={textColor}><strong>Subtotal:</strong></Text>
                          <Text fontWeight="medium" color={textColor}>
                            {salesService.formatCurrency(calculateSubtotal())}
                          </Text>
                        </HStack>
                        <HStack justify="space-between">
                          <Text color={textColor}><strong>Total Amount:</strong></Text>
                          <Text fontSize="lg" fontWeight="bold" color="blue.600">
                            {salesService.formatCurrency(calculateTotal())}
                          </Text>
                        </HStack>
                      </VStack>
                    </AlertDescription>
                  </Alert>
                )}
              </Box>

              <Divider />

              {/* Additional Information */}
              <Box>
                <Heading size="md" mb={4} color={subHeadingColor}>
                  📝 Additional Information
                </Heading>
                <VStack spacing={4}>
                  <HStack w="full" spacing={4}>
                      <FormControl>
                        <FormLabel>Payment Terms</FormLabel>
                        <Select 
                          {...register('payment_terms')}
                          bg={inputBg}
                          _focus={{ bg: inputFocusBg }}
                        >
                          <option value="COD">COD (Cash on Delivery)</option>
                          <option value="NET_15">NET 15</option>
                          <option value="NET_30">NET 30</option>
                          <option value="NET_60">NET 60</option>
                          <option value="NET_90">NET 90</option>
                        </Select>
                      </FormControl>

                      <FormControl>
                        <FormLabel>Reference</FormLabel>
                        <Input
                          {...register('reference')}
                          placeholder="External reference number"
                          bg={inputBg}
                          _focus={{ bg: inputFocusBg }}
                        />
                      </FormControl>
                  </HStack>

                  <FormControl>
                    <FormLabel>Notes</FormLabel>
                    <Textarea
                      {...register('notes')}
                      placeholder="Customer-visible notes"
                      rows={3}
                      bg={inputBg}
                      _focus={{ bg: inputFocusBg }}
                    />
                  </FormControl>

                  <FormControl>
                    <FormLabel>Internal Notes</FormLabel>
                    <Textarea
                      {...register('internal_notes')}
                      placeholder="Internal notes (not visible to customer)"
                      rows={3}
                      bg={inputBg}
                      _focus={{ bg: inputFocusBg }}
                    />
                  </FormControl>
                </VStack>
              </Box>
            </VStack>
          </ModalBody>

          <ModalFooter 
            position="sticky"
            bottom={0}
            borderTopWidth={2} 
            borderColor={borderColor} 
            bg={footerBg}
            boxShadow={`0 -4px 12px ${shadowColor}`}
            px={6}
            py={4}
            mt={6}
            flexShrink={0}
            zIndex={10}
          >
            <HStack justify="space-between" spacing={4} w="full">
              {/* Left side - Form info */}
              <HStack spacing={2}>
                <Text fontSize="sm" color={textColor}>
                  {loadingData ? 'Loading...' : `${fields.length} item${fields.length !== 1 ? 's' : ''}`}
                </Text>
                {calculateSubtotal() > 0 && (
                  <Text fontSize="sm" color="blue.600" fontWeight="medium">
                    Total: {salesService.formatCurrency(calculateTotal())}
                  </Text>
                )}
              </HStack>
              
              {/* Right side - Action buttons */}
              <HStack spacing={3}>
                <Button
                  leftIcon={<FiX />}
                  onClick={handleClose}
                  variant="outline"
                  size="lg"
                  isDisabled={loading}
                  colorScheme="gray"
                  minW="120px"
                >
                  Cancel
                </Button>
                <Button
                  leftIcon={loading ? undefined : <FiSave />}
                  type="submit"
                  colorScheme="blue"
                  size="lg"
                  isLoading={loading}
                  loadingText={sale ? "Updating..." : "Creating..."}
                  minW="150px"
                  shadow="md"
                  _hover={{
                    shadow: "lg",
                    transform: "translateY(-1px)",
                  }}
                  _active={{
                    transform: "translateY(0)",
                  }}
                >
                  {sale ? 'Update Sale' : 'Create Sale'}
                </Button>
              </HStack>
            </HStack>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};

export default SalesForm;
