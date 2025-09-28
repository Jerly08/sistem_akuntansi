'use client';

import React, { useState, useEffect } from 'react';
import SimpleLayout from '@/components/layout/SimpleLayout';
import ProtectedModule from '@/components/common/ProtectedModule';
import { useAuth } from '@/contexts/AuthContext';
import { useModulePermissions } from '@/hooks/usePermissions';
import EnhancedSalesTable from '@/components/sales/EnhancedSalesTable';
import EnhancedStatsCards from '@/components/sales/EnhancedStatsCards';
import SalesForm from '@/components/sales/SalesForm';
import PaymentForm from '@/components/sales/PaymentForm';
import salesService, { Sale, SalesFilter } from '@/services/salesService';
import {
  Box,
  Heading,
  Text,
  Button,
  Flex,
  HStack,
  Input,
  InputGroup,
  InputLeftElement,
  Card,
  CardHeader,
  CardBody,
  useToast,
  useDisclosure,
  Alert,
  AlertIcon,
  Select,
  VStack,
  useColorModeValue,
  Tooltip,
  IconButton,
  Icon,
} from '@chakra-ui/react';
import {
  FiPlus,
  FiSearch,
  FiDownload,
  FiRefreshCw,
  FiFilter,
  FiBarChart,
} from 'react-icons/fi';
import {
  handleLoadingError,
  handleDeleteError,
  handleSuccess
} from '@/utils/errorHandler';

interface SalesPageState {
  sales: Sale[];
  loading: boolean;
  error: string | null;
  filter: SalesFilter;
  selectedSale: Sale | null;
  summary: any;
}

const SalesPage: React.FC = () => {
  const { user } = useAuth();
  const toast = useToast();
  const { canCreate, canEdit, canDelete, canExport } = useModulePermissions('sales');
  
  const { isOpen: isFormOpen, onOpen: onFormOpen, onClose: onFormClose } = useDisclosure();
  const { isOpen: isPaymentOpen, onOpen: onPaymentOpen, onClose: onPaymentClose } = useDisclosure();
  
  // Theme-aware colors
  const headingColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const subheadingColor = useColorModeValue('gray.600', 'var(--text-secondary)');
  const tableBg = useColorModeValue('white', 'var(--bg-secondary)');
  const textColor = useColorModeValue('gray.600', 'var(--text-secondary)');

  const [state, setState] = useState<SalesPageState>({
    sales: [],
    loading: true,
    error: null,
    filter: {
      page: 1,
      limit: 10,
      search: '',
      status: '',
      start_date: '',
      end_date: ''
    },
    selectedSale: null,
    summary: null
  });

  // Load sales data
  const loadSales = async (newFilter?: Partial<SalesFilter>) => {
    try {
      setState(prev => ({ ...prev, loading: true, error: null }));
      
      const filter = newFilter ? { ...state.filter, ...newFilter } : state.filter;
      const result = await salesService.getSales(filter);
      
      setState(prev => ({
        ...prev,
        sales: result.data,
        filter: { ...filter, page: result.page },
        loading: false
      }));
    } catch (error: any) {
      setState(prev => ({
        ...prev,
        error: error.response?.data?.message || 'Failed to load sales data',
        loading: false
      }));
      
      toast({
        title: 'Error loading sales',
        description: error.response?.data?.message || 'Failed to load sales data',
        status: 'error',
        duration: 3000
      });
    }
  };

  // Load sales summary
  const loadSalesSummary = async () => {
    try {
      const summary = await salesService.getSalesSummary();
      setState(prev => ({ ...prev, summary }));
    } catch (error) {
      console.error('Failed to load sales summary:', error);
    }
  };

  // Initial load
  useEffect(() => {
    loadSales();
    loadSalesSummary();
  }, []);

  // Handle search
  const handleSearch = (searchTerm: string) => {
    setState(prev => ({ ...prev, filter: { ...prev.filter, search: searchTerm } }));
    loadSales({ search: searchTerm, page: 1 });
  };

  // Handle filter change
  const handleFilterChange = (key: keyof SalesFilter, value: string) => {
    setState(prev => ({ ...prev, filter: { ...prev.filter, [key]: value } }));
    loadSales({ [key]: value, page: 1 });
  };

  // Handle create/edit sale
  const handleSaleAction = (sale?: Sale) => {
    setState(prev => ({ ...prev, selectedSale: sale || null }));
    onFormOpen();
  };

  // Handle payment
  const handlePayment = (sale: Sale) => {
    setState(prev => ({ ...prev, selectedSale: sale }));
    onPaymentOpen();
  };

  // Handle delete sale
  const handleDelete = async (sale: Sale) => {
    if (!window.confirm('Are you sure you want to delete this sale?')) return;
    
    try {
      await salesService.deleteSale(sale.id);
      handleSuccess('Sale has been deleted successfully', toast, 'delete sale');
      loadSales();
      loadSalesSummary();
    } catch (error: any) {
      handleDeleteError('sale', error, toast);
    }
  };

  // Handle sale status actions
  const handleSaleStatusAction = async (sale: Sale, action: 'confirm' | 'cancel') => {
    try {
      let message = '';
      
      switch (action) {
        case 'confirm':
          // FIXED: Use createInvoiceFromSale to directly create invoice (DRAFT -> INVOICED)
          // This will create journal entries and set proper accounting impact
          await salesService.createInvoiceFromSale(sale.id);
          message = 'Sale has been invoiced successfully! Journal entries have been created.';
          break;
        case 'cancel':
          const reason = window.prompt('Please provide a reason for cancellation:');
          if (!reason) return;
          await salesService.cancelSale(sale.id, reason);
          message = 'Sale has been cancelled';
          break;
      }
      
      handleSuccess(message, toast, action + ' sale');
      loadSales();
      loadSalesSummary();
    } catch (error: any) {
      toast({
        title: `Error ${action}ing sale`,
        description: error.response?.data?.message || `Failed to ${action} sale`,
        status: 'error',
        duration: 3000
      });
    }
  };

  // Handle bulk export
  const handleBulkExport = async () => {
    try {
      await salesService.downloadSalesReportPDF(
        state.filter.start_date || undefined,
        state.filter.end_date || undefined
      );
      handleSuccess('Sales report has been downloaded', toast, 'export report');
    } catch (error: any) {
      toast({
        title: 'Export failed',
        description: 'Failed to export sales report',
        status: 'error',
        duration: 3000
      });
    }
  };

  // Handle form save
  const handleFormSave = () => {
    onFormClose();
    loadSales();
    loadSalesSummary();
  };

  // Handle payment save
  const handlePaymentSave = () => {
    onPaymentClose();
    loadSales();
    loadSalesSummary();
  };

  // Handle view details
  const handleViewDetails = (sale: Sale) => {
    window.open(`/sales/${sale.id}`, '_blank');
  };

  // Handle download invoice
  const handleDownloadInvoice = (sale: Sale) => {
    salesService.downloadInvoicePDF(sale.id, sale.invoice_number || sale.code);
  };

  return (
    <ProtectedModule module="sales">
      <SimpleLayout>
      <Box>
        {/* Header */}
        <Flex justify="space-between" align="center" mb={6} wrap="wrap" gap={4}>
          <VStack align="start" spacing={1}>
            <Heading size="xl" color={headingColor} fontWeight="600">
              Sales Management
            </Heading>
            <Text color={subheadingColor} fontSize="md">
              Manage your sales transactions and invoices
            </Text>
          </VStack>
          
          <HStack spacing={3}>
            <Tooltip label="Refresh Data">
              <IconButton
                aria-label="Refresh"
                icon={<FiRefreshCw />}
                variant="ghost"
                onClick={() => { loadSales(); loadSalesSummary(); }}
                isLoading={state.loading}
              />
            </Tooltip>
            
            {canExport && (
              <Button
                leftIcon={<FiDownload />}
                colorScheme="green"
                variant="outline"
                size="md"
                onClick={handleBulkExport}
              >
                Export Report
              </Button>
            )}
            
            {(canCreate || !user) && (
              <Button 
                leftIcon={<FiPlus />} 
                colorScheme="blue" 
                size="md"
                px={6}
                fontWeight="medium"
                onClick={() => handleSaleAction()}
                _hover={{ 
                  transform: 'translateY(-1px)',
                  boxShadow: 'lg'
                }}
              >
                Create Sale
              </Button>
            )}
          </HStack>
        </Flex>

        {/* Summary Cards */}
        {state.summary && (
          <Box mb={6}>
            <EnhancedStatsCards 
              stats={state.summary} 
              formatCurrency={salesService.formatCurrency}
            />
          </Box>
        )}

        {/* Search and Filters */}
        <Card mb={6}>
          <CardBody>
            <Flex gap={4} align="end" wrap="wrap">
              <Box flex="1" minW="300px">
                <Text fontSize="sm" fontWeight="medium" mb={2} color={textColor}>
                  Search Transactions
                </Text>
                <InputGroup>
                  <InputLeftElement pointerEvents="none">
                    <FiSearch color={textColor} />
                  </InputLeftElement>
                  <Input
                    placeholder="Search by invoice, customer, or code..."
                    value={state.filter.search}
                    onChange={(e) => handleSearch(e.target.value)}
                    bg={tableBg}
                  />
                </InputGroup>
              </Box>
              
              <Box minW="180px">
                <Text fontSize="sm" fontWeight="medium" mb={2} color={textColor}>
                  Filter by Status
                </Text>
                <Select 
                  placeholder="All Status"
                  value={state.filter.status}
                  onChange={(e) => handleFilterChange('status', e.target.value)}
                  bg={tableBg}
                >
                  <option value="DRAFT">Draft</option>
                  <option value="INVOICED">Invoiced</option>
                  <option value="PAID">Paid</option>
                  <option value="OVERDUE">Overdue</option>
                  <option value="CANCELLED">Cancelled</option>
                </Select>
              </Box>
              
              <Box minW="160px">
                <Text fontSize="sm" fontWeight="medium" mb={2} color={textColor}>
                  Start Date
                </Text>
                <Input
                  type="date"
                  value={state.filter.start_date}
                  onChange={(e) => handleFilterChange('start_date', e.target.value)}
                  bg={tableBg}
                />
              </Box>
              
              <Box minW="160px">
                <Text fontSize="sm" fontWeight="medium" mb={2} color={textColor}>
                  End Date
                </Text>
                <Input
                  type="date"
                  value={state.filter.end_date}
                  onChange={(e) => handleFilterChange('end_date', e.target.value)}
                  bg={tableBg}
                />
              </Box>
              
              <Button
                leftIcon={<FiFilter />}
                variant="outline"
                onClick={() => {
                  setState(prev => ({
                    ...prev,
                    filter: {
                      ...prev.filter,
                      search: '',
                      status: '',
                      start_date: '',
                      end_date: ''
                    }
                  }));
                  loadSales({
                    search: '',
                    status: '',
                    start_date: '',
                    end_date: '',
                    page: 1
                  });
                }}
              >
                Clear Filters
              </Button>
            </Flex>
          </CardBody>
        </Card>

        {/* Error Alert */}
        {state.error && (
          <Alert status="error" mb={6}>
            <AlertIcon />
            {state.error}
          </Alert>
        )}

        {/* Enhanced Sales Table */}
        <EnhancedSalesTable
          sales={state.sales || []}
          loading={state.loading}
          onViewDetails={handleViewDetails}
          onEdit={canEdit ? (sale) => handleSaleAction(sale) : undefined}
          onConfirm={canEdit ? (sale) => handleSaleStatusAction(sale, 'confirm') : undefined}
          onCancel={canEdit ? (sale) => handleSaleStatusAction(sale, 'cancel') : undefined}
          onPayment={canEdit ? handlePayment : undefined}
          onDelete={canDelete ? handleDelete : undefined}
          onDownloadInvoice={handleDownloadInvoice}
          formatCurrency={salesService.formatCurrency}
          formatDate={salesService.formatDate}
          getStatusLabel={salesService.getStatusLabel}
          canEdit={canEdit}
          canDelete={canDelete}
        />

      </Box>
      
      {/* Sales Form Modal */}
      <SalesForm
        isOpen={isFormOpen}
        onClose={onFormClose}
        onSave={handleFormSave}
        sale={state.selectedSale}
      />

      {/* Payment Form Modal */}
      <PaymentForm
        isOpen={isPaymentOpen}
        onClose={onPaymentClose}
        onSave={handlePaymentSave}
        sale={state.selectedSale}
      />
      </SimpleLayout>
    </ProtectedModule>
  );
};

export default SalesPage;

