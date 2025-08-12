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
  VStack,
  HStack,
  Grid,
  GridItem,
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
  Flex
} from '@chakra-ui/react';
import { useForm, useFieldArray } from 'react-hook-form';
import { FiPlus, FiTrash2, FiSave } from 'react-icons/fi';
import salesService, { 
  Sale, 
  SaleCreateRequest, 
  SaleUpdateRequest, 
  SaleItemCreateRequest,
  SaleItemUpdateRequest 
} from '@/services/salesService';

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
  const toast = useToast();

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
      type: 'SALE',
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

  const loadFormData = async () => {
    try {
      // Load customers, products, sales persons, and accounts
      // These would typically come from their respective services
      setCustomers([
        { id: 1, name: 'PT ABC Corp', code: 'CUST001' },
        { id: 2, name: 'CV XYZ Ltd', code: 'CUST002' },
        { id: 3, name: 'Toko Makmur', code: 'CUST003' }
      ]);

      setProducts([
        { id: 1, name: 'Product A', code: 'PROD001', price: 100000 },
        { id: 2, name: 'Product B', code: 'PROD002', price: 150000 },
        { id: 3, name: 'Service C', code: 'SERV001', price: 200000 }
      ]);

      setSalesPersons([
        { id: 1, name: 'John Doe', email: 'john@company.com' },
        { id: 2, name: 'Jane Smith', email: 'jane@company.com' }
      ]);

      setAccounts([
        { id: 1, name: 'Sales Revenue', code: '4000' },
        { id: 2, name: 'Service Revenue', code: '4100' }
      ]);
    } catch (error) {
      console.error('Error loading form data:', error);
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
      type: 'SALE',
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

      // Validate items
      const validItems = data.items.filter(item => item.product_id > 0);
      if (validItems.length === 0) {
        toast({
          title: 'Validation Error',
          description: 'At least one item is required',
          status: 'error',
          duration: 3000
        });
        return;
      }

      if (sale) {
        // Update existing sale
        const updateData: SaleUpdateRequest = {
          customer_id: data.customer_id,
          sales_person_id: data.sales_person_id,
          date: new Date(data.date),
          due_date: data.due_date ? new Date(data.due_date) : undefined,
          valid_until: data.valid_until ? new Date(data.valid_until) : undefined,
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
            discount_percent: item.discount_percent,
            taxable: item.taxable,
            revenue_account_id: item.revenue_account_id,
            delete: item.delete || false
          }))
        };

        await salesService.updateSale(sale.id, updateData);
        toast({
          title: 'Sale Updated',
          description: 'Sale has been updated successfully',
          status: 'success',
          duration: 3000
        });
      } else {
        // Create new sale
        const createData: SaleCreateRequest = {
          customer_id: data.customer_id,
          sales_person_id: data.sales_person_id,
          type: data.type,
          date: new Date(data.date),
          due_date: data.due_date ? new Date(data.due_date) : undefined,
          valid_until: data.valid_until ? new Date(data.valid_until) : undefined,
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
            description: item.description,
            quantity: item.quantity,
            unit_price: item.unit_price,
            discount_percent: item.discount_percent,
            taxable: item.taxable,
            revenue_account_id: item.revenue_account_id
          }))
        };

        await salesService.createSale(createData);
        toast({
          title: 'Sale Created',
          description: 'Sale has been created successfully',
          status: 'success',
          duration: 3000
        });
      }

      onSave();
      onClose();
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.message || `Failed to ${sale ? 'update' : 'create'} sale`,
        status: 'error',
        duration: 5000
      });
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} size="6xl">
      <ModalOverlay bg="blackAlpha.600" />
      <ModalContent maxH="95vh" mx={4}>
        <ModalHeader bg="blue.50" borderBottomWidth={1} borderColor="gray.200">
          <VStack align="start" spacing={1}>
            <Text fontSize="xl" fontWeight="bold" color="blue.700">
              {sale ? 'Edit Sale Transaction' : 'Create New Sale'}
            </Text>
            <Text fontSize="sm" color="gray.600">
              {sale ? 'Modify existing sale details and items' : 'Create a new sales transaction with items and pricing'}
            </Text>
          </VStack>
        </ModalHeader>
        <ModalCloseButton />

        <form onSubmit={handleSubmit(onSubmit)}>
          <ModalBody overflowY="auto">
            <VStack spacing={6} align="stretch">
              {/* Basic Information */}
              <Box>
                <Text fontSize="lg" fontWeight="bold" mb={4}>Basic Information</Text>
                <Grid templateColumns="repeat(3, 1fr)" gap={4}>
                  <GridItem>
                    <FormControl isRequired isInvalid={!!errors.customer_id}>
                      <FormLabel>Customer</FormLabel>
                      <Select
                        {...register('customer_id', {
                          required: 'Customer is required',
                          setValueAs: value => parseInt(value) || 0
                        })}
                      >
                        <option value="">Select customer</option>
                        {customers.map(customer => (
                          <option key={customer.id} value={customer.id}>
                            {customer.code} - {customer.name}
                          </option>
                        ))}
                      </Select>
                      <FormErrorMessage>{errors.customer_id?.message}</FormErrorMessage>
                    </FormControl>
                  </GridItem>

                  <GridItem>
                    <FormControl>
                      <FormLabel>Sales Person</FormLabel>
                      <Select
                        {...register('sales_person_id', {
                          setValueAs: value => value ? parseInt(value) : undefined
                        })}
                      >
                        <option value="">Select sales person</option>
                        {salesPersons.map(person => (
                          <option key={person.id} value={person.id}>
                            {person.name}
                          </option>
                        ))}
                      </Select>
                    </FormControl>
                  </GridItem>

                  <GridItem>
                    <FormControl isRequired isInvalid={!!errors.type}>
                      <FormLabel>Type</FormLabel>
                      <Select
                        {...register('type', {
                          required: 'Sale type is required'
                        })}
                      >
                        <option value="QUOTATION">Quotation</option>
                        <option value="ORDER">Order</option>
                        <option value="INVOICE">Invoice</option>
                        <option value="SALE">Sale</option>
                      </Select>
                      <FormErrorMessage>{errors.type?.message}</FormErrorMessage>
                    </FormControl>
                  </GridItem>

                  <GridItem>
                    <FormControl isRequired isInvalid={!!errors.date}>
                      <FormLabel>Date</FormLabel>
                      <Input
                        type="date"
                        {...register('date', {
                          required: 'Date is required'
                        })}
                      />
                      <FormErrorMessage>{errors.date?.message}</FormErrorMessage>
                    </FormControl>
                  </GridItem>

                  <GridItem>
                    <FormControl>
                      <FormLabel>Due Date</FormLabel>
                      <Input
                        type="date"
                        {...register('due_date')}
                      />
                    </FormControl>
                  </GridItem>

                  <GridItem>
                    <FormControl>
                      <FormLabel>Valid Until</FormLabel>
                      <Input
                        type="date"
                        {...register('valid_until')}
                      />
                    </FormControl>
                  </GridItem>
                </Grid>
              </Box>

              <Divider />

              {/* Items Section */}
              <Box>
                <Flex justify="space-between" align="center" mb={4}>
                  <Text fontSize="lg" fontWeight="bold">Items</Text>
                  <Button
                    size="sm"
                    colorScheme="blue"
                    leftIcon={<FiPlus />}
                    onClick={addItem}
                  >
                    Add Item
                  </Button>
                </Flex>

                <TableContainer>
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
                            <NumberInput size="sm" min={0}>
                              <NumberInputField
                                {...register(`items.${index}.unit_price`, {
                                  required: 'Unit price is required',
                                  min: 0,
                                  setValueAs: value => parseFloat(value) || 0
                                })}
                              />
                            </NumberInput>
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
                </TableContainer>
              </Box>

              <Divider />

              {/* Pricing & Taxes */}
              <Box>
                <Text fontSize="lg" fontWeight="bold" mb={4}>Pricing & Taxes</Text>
                <Grid templateColumns="repeat(3, 1fr)" gap={4}>
                  <GridItem>
                    <FormControl>
                      <FormLabel>Global Discount (%)</FormLabel>
                      <NumberInput min={0} max={100}>
                        <NumberInputField
                          {...register('discount_percent', {
                            setValueAs: value => parseFloat(value) || 0
                          })}
                        />
                      </NumberInput>
                    </FormControl>
                  </GridItem>

                  <GridItem>
                    <FormControl>
                      <FormLabel>PPN (%)</FormLabel>
                      <NumberInput min={0} max={100}>
                        <NumberInputField
                          {...register('ppn_percent', {
                            setValueAs: value => parseFloat(value) || 0
                          })}
                        />
                      </NumberInput>
                    </FormControl>
                  </GridItem>

                  <GridItem>
                    <FormControl>
                      <FormLabel>Shipping Cost</FormLabel>
                      <NumberInput min={0}>
                        <NumberInputField
                          {...register('shipping_cost', {
                            setValueAs: value => parseFloat(value) || 0
                          })}
                        />
                      </NumberInput>
                    </FormControl>
                  </GridItem>
                </Grid>

                <Box mt={4} p={4} bg="gray.50" borderRadius="md">
                  <Grid templateColumns="repeat(2, 1fr)" gap={4}>
                    <Box>
                      <Text fontSize="sm" color="gray.600">Subtotal:</Text>
                      <Text fontWeight="medium">{salesService.formatCurrency(calculateSubtotal())}</Text>
                    </Box>
                    <Box>
                      <Text fontSize="sm" color="gray.600">Total Amount:</Text>
                      <Text fontSize="lg" fontWeight="bold" color="blue.500">
                        {salesService.formatCurrency(calculateTotal())}
                      </Text>
                    </Box>
                  </Grid>
                </Box>
              </Box>

              <Divider />

              {/* Additional Information */}
              <Box>
                <Text fontSize="lg" fontWeight="bold" mb={4}>Additional Information</Text>
                <Grid templateColumns="repeat(2, 1fr)" gap={4}>
                  <GridItem>
                    <FormControl>
                      <FormLabel>Payment Terms</FormLabel>
                      <Select {...register('payment_terms')}>
                        <option value="COD">COD (Cash on Delivery)</option>
                        <option value="NET_15">NET 15</option>
                        <option value="NET_30">NET 30</option>
                        <option value="NET_60">NET 60</option>
                        <option value="NET_90">NET 90</option>
                      </Select>
                    </FormControl>
                  </GridItem>

                  <GridItem>
                    <FormControl>
                      <FormLabel>Reference</FormLabel>
                      <Input
                        {...register('reference')}
                        placeholder="External reference number"
                      />
                    </FormControl>
                  </GridItem>

                  <GridItem colSpan={2}>
                    <FormControl>
                      <FormLabel>Notes</FormLabel>
                      <Textarea
                        {...register('notes')}
                        placeholder="Customer-visible notes"
                        rows={3}
                      />
                    </FormControl>
                  </GridItem>

                  <GridItem colSpan={2}>
                    <FormControl>
                      <FormLabel>Internal Notes</FormLabel>
                      <Textarea
                        {...register('internal_notes')}
                        placeholder="Internal notes (not visible to customer)"
                        rows={3}
                      />
                    </FormControl>
                  </GridItem>
                </Grid>
              </Box>
            </VStack>
          </ModalBody>

          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={handleClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              colorScheme="blue"
              isLoading={loading}
              loadingText={sale ? "Updating..." : "Creating..."}
              leftIcon={<FiSave />}
            >
              {sale ? 'Update Sale' : 'Create Sale'}
            </Button>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};

export default SalesForm;
