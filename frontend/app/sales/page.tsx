'use client';

import React, { useState, useEffect } from 'react';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { DataTable } from '@/components/common/DataTable';
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
  Badge,
  IconButton,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  MenuDivider,
  useToast,
  useDisclosure,
  Spinner,
  Alert,
  AlertIcon,
  Select,
  VStack,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  SimpleGrid
} from '@chakra-ui/react';
import {
  FiPlus,
  FiSearch,
  FiMoreVertical,
  FiEdit,
  FiEye,
  FiDollarSign,
  FiTrash2,
  FiDownload,
  FiRefreshCw,
  FiFileText,
  FiCheck,
  FiX,
  FiTrendingUp,
  FiBarChart3,
  FiSend
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
  
  const { isOpen: isFormOpen, onOpen: onFormOpen, onClose: onFormClose } = useDisclosure();
  const { isOpen: isPaymentOpen, onOpen: onPaymentOpen, onClose: onPaymentClose } = useDisclosure();
  
  const canCreate = user?.role === 'ADMIN' || user?.role === 'FINANCE' || user?.role === 'DIRECTOR';
  const canEdit = user?.role === 'ADMIN' || user?.role === 'FINANCE' || user?.role === 'DIRECTOR';
  const canDelete = user?.role === 'ADMIN';
  const canExport = user?.role === 'ADMIN' || user?.role === 'FINANCE' || user?.role === 'DIRECTOR';

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
    loadSales({ search: searchTerm, page: 1 });
  };

  // Handle filter change
  const handleFilterChange = (key: keyof SalesFilter, value: string) => {
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
          await salesService.confirmSale(sale.id);
          message = 'Sale has been confirmed and invoiced successfully';
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

  // Get status badge color
  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'paid': return 'green';
      case 'invoiced': return 'blue';
      case 'confirmed': return 'purple';
      case 'overdue': return 'red';
      case 'draft': return 'gray';
      case 'cancelled': return 'red';
      default: return 'gray';
    }
  };


  // Define columns for the table
  const columns = [
    {
      header: 'Code',
      accessor: (row: Sale) => (
        <Text fontWeight="medium" color="blue.600">{row.code}</Text>
      )
    },
    {
      header: 'Invoice #',
      accessor: (row: Sale) => row.invoice_number || '-'
    },
    {
      header: 'Customer',
      accessor: (row: Sale) => row.customer?.name || 'N/A'
    },
    {
      header: 'Date',
      accessor: (row: Sale) => salesService.formatDate(row.date)
    },
    {
      header: 'Total',
      accessor: (row: Sale) => (
        <Text fontWeight="medium">{salesService.formatCurrency(row.total_amount)}</Text>
      )
    },
    {
      header: 'Outstanding',
      accessor: (row: Sale) => (
        <Text 
          fontWeight="medium" 
          color={row.outstanding_amount > 0 ? 'orange.600' : 'green.600'}
        >
          {salesService.formatCurrency(row.outstanding_amount)}
        </Text>
      )
    },
    {
      header: 'Status',
      accessor: (row: Sale) => (
        <Badge colorScheme={getStatusColor(row.status)} variant="subtle">
          {salesService.getStatusLabel(row.status)}
        </Badge>
      )
    },
    {
      header: 'Actions',
      accessor: (row: Sale) => (
        <Menu>
          <MenuButton as={IconButton} icon={<FiMoreVertical />} variant="ghost" size="sm" />
          <MenuList>
            <MenuItem icon={<FiEye />} onClick={() => window.open(`/sales/${row.id}`, '_blank')}>
              View Details
            </MenuItem>
            {row.status === 'DRAFT' && canEdit && (
              <MenuItem icon={<FiEdit />} onClick={() => handleSaleAction(row)}>
                Edit
              </MenuItem>
            )}
            {row.status === 'DRAFT' && canEdit && (
              <MenuItem icon={<FiCheck />} onClick={() => handleSaleStatusAction(row, 'confirm')}>Confirm & Invoice</MenuItem>
            )}
            {row.status === 'INVOICED' && row.outstanding_amount > 0 && canEdit && (
              <MenuItem icon={<FiDollarSign />} onClick={() => handlePayment(row)}>
                Record Payment
              </MenuItem>
            )}
            {canEdit && row.status !== 'PAID' && row.status !== 'CANCELLED' && (
              <MenuItem icon={<FiX />} onClick={() => handleSaleStatusAction(row, 'cancel')}>Cancel Sale</MenuItem>
            )}
            <MenuItem 
              icon={<FiDownload />} 
              onClick={() => salesService.downloadInvoicePDF(row.id, row.invoice_number || row.code)}
            >
              Download Invoice
            </MenuItem>
            {canDelete && (
              <>
                <MenuDivider />
                <MenuItem icon={<FiTrash2 />} color="red.500" onClick={() => handleDelete(row)}>
                  Delete
                </MenuItem>
              </>
            )}
          </MenuList>
        </Menu>
      )
    }
  ];

  return (
    <Layout allowedRoles={['ADMIN', 'FINANCE', 'DIRECTOR', 'EMPLOYEE', 'INVENTORY_MANAGER']}>
      <VStack spacing={6} align="stretch">
        {/* Summary Cards */}
        {state.summary && (
          <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>Total Sales</StatLabel>
                  <StatNumber>{state.summary.total_sales}</StatNumber>
                  <StatHelpText>This period</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>Total Revenue</StatLabel>
                  <StatNumber>{salesService.formatCurrency(state.summary.total_amount)}</StatNumber>
                  <StatHelpText>Gross revenue</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>Outstanding</StatLabel>
                  <StatNumber color="orange.500">
                    {salesService.formatCurrency(state.summary.total_outstanding)}
                  </StatNumber>
                  <StatHelpText>Unpaid invoices</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>Avg Order Value</StatLabel>
                  <StatNumber>{salesService.formatCurrency(state.summary.avg_order_value)}</StatNumber>
                  <StatHelpText>Per transaction</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </SimpleGrid>
        )}

        {/* Header */}
        <Flex justify="space-between" align="center">
          <Box>
            <Heading as="h1" size="xl" mb={2}>Sales Management</Heading>
            <Text color="gray.600">Manage your sales transactions and invoices</Text>
          </Box>
          <HStack spacing={3}>
            <Button
              leftIcon={<FiRefreshCw />}
              variant="ghost"
              onClick={() => { loadSales(); loadSalesSummary(); }}
              isLoading={state.loading}
            >
              Refresh
            </Button>
            {canExport && (
              <Button
                leftIcon={<FiDownload />}
                colorScheme="green"
                variant="outline"
                onClick={handleBulkExport}
              >
                Export Report
              </Button>
            )}
            {/* Show Create Sale button for authorized users or as fallback */}
            {(canCreate || !user) && (
              <Button 
                leftIcon={<FiPlus />} 
                colorScheme="blue" 
                size="lg"
                onClick={() => handleSaleAction()}
              >
                Create Sale
              </Button>
            )}
          </HStack>
        </Flex>

        {/* Search and Filters */}
        <Card>
          <CardBody>
            <HStack spacing={4} wrap="wrap">
              <InputGroup maxW="400px">
                <InputLeftElement pointerEvents="none">
                  <FiSearch color="gray.300" />
                </InputLeftElement>
                <Input 
                  placeholder="Search by invoice, customer, or code..."
                  value={state.filter.search}
                  onChange={(e) => handleSearch(e.target.value)}
                />
              </InputGroup>
              
              <Select 
                maxW="200px" 
                placeholder="All Status"
                value={state.filter.status}
                onChange={(e) => handleFilterChange('status', e.target.value)}
              >
                <option value="DRAFT">Draft</option>
                {/* Remove CONFIRMED status as it's no longer used */}
                <option value="INVOICED">Invoiced</option>
                <option value="PAID">Paid</option>
                <option value="OVERDUE">Overdue</option>
                <option value="CANCELLED">Cancelled</option>
              </Select>
              
              <Input
                type="date"
                maxW="200px"
                placeholder="Start Date"
                value={state.filter.start_date}
                onChange={(e) => handleFilterChange('start_date', e.target.value)}
              />
              
              <Input
                type="date"
                maxW="200px"
                placeholder="End Date"
                value={state.filter.end_date}
                onChange={(e) => handleFilterChange('end_date', e.target.value)}
              />
            </HStack>
          </CardBody>
        </Card>

        {/* Error Alert */}
        {state.error && (
          <Alert status="error">
            <AlertIcon />
            {state.error}
          </Alert>
        )}

        {/* Sales Table */}
        <Card>
          <CardHeader>
            <Flex justify="space-between" align="center">
              <Heading size="md">Sales Transactions ({state.sales?.length || 0})</Heading>
            </Flex>
          </CardHeader>
          <CardBody>
            {state.loading ? (
              <Flex justify="center" py={10}>
                <Spinner size="lg" />
              </Flex>
            ) : (
              <DataTable 
                columns={columns} 
                data={state.sales || []} 
                keyField="id"
                searchable={false}
                pagination={true}
                pageSize={10}
              />
            )}
          </CardBody>
        </Card>
      </VStack>

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
    </Layout>
  );
};

export default SalesPage;

