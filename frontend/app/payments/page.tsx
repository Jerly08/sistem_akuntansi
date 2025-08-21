'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import Layout from '@/components/layout/Layout';
import { DataTable } from '@/components/common/DataTable';
import {
  Box,
  Flex,
  Heading,
  Button,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Badge,
  Text,
  HStack,
  VStack,
  Select,
  Input,
  InputGroup,
  InputLeftElement,
  FormControl,
  FormLabel,
  IconButton,
  Tooltip,
  Spinner,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  useDisclosure,
  AlertDialog,
  AlertDialogBody,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogContent,
  AlertDialogOverlay,
  Card,
  CardHeader,
  CardBody,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  SimpleGrid,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  MenuDivider,
  useToast,
} from '@chakra-ui/react';
import {
  FiPlus, 
  FiEye, 
  FiEdit, 
  FiTrash2, 
  FiFilter, 
  FiRefreshCw, 
  FiDownload, 
  FiDollarSign, 
  FiFilePlus,
  FiSearch,
  FiMoreVertical,
  FiFileText,
  FiChevronDown
} from 'react-icons/fi';
import paymentService, { Payment, PaymentFilters, PaymentResult, PaymentCreateRequest } from '@/services/paymentService';
import AdvancedPaymentForm from '@/components/payments/AdvancedPaymentForm';
import PaymentDetailModal from '@/components/payments/PaymentDetailModal';
import { exportPaymentsToPDF, exportPaymentDetailToPDF, PDFExportOptions } from '@/utils/pdfExport';
import ExportButton from '@/components/common/ExportButton';

// Status type for filtering
type PaymentStatusType = 'ALL' | 'PENDING' | 'COMPLETED' | 'FAILED';

// Payment method type for filtering
type PaymentMethodType = 'ALL' | 'CASH' | 'BANK_TRANSFER' | 'CHECK' | 'CREDIT_CARD' | 'DEBIT_CARD' | 'OTHER';

// Pagination settings
const ITEMS_PER_PAGE = 10;

// Date formatter for display
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('id-ID');
};

// Currency formatter - fixed to match sales format
const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount);
};

const PaymentsPage: React.FC = () => {
  const { token, user } = useAuth();
  const toast = useToast();
  const [payments, setPayments] = useState<Payment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showFilters, setShowFilters] = useState(false);
  const [pagination, setPagination] = useState({
    current: 1,
    total: 1,
    totalItems: 0
  });
  const [summary, setSummary] = useState<any>(null);
  
  // Filter states
  const [filters, setFilters] = useState<PaymentFilters>({
    page: 1,
    limit: ITEMS_PER_PAGE
  });
  const [statusFilter, setStatusFilter] = useState<PaymentStatusType>('ALL');
  const [methodFilter, setMethodFilter] = useState<PaymentMethodType>('ALL');
  const [startDate, setStartDate] = useState<string>('');
  const [endDate, setEndDate] = useState<string>('');
  
  // State for modals
  const [showPaymentForm, setShowPaymentForm] = useState(false);
  const [selectedPayment, setSelectedPayment] = useState<Payment | null>(null);
  const [showConfirmDelete, setShowConfirmDelete] = useState(false);
  const [showPaymentDetail, setShowPaymentDetail] = useState(false);
  const [formLoading, setFormLoading] = useState(false);

  // Permission checks
  const canCreate = user?.role === 'ADMIN' || user?.role === 'FINANCE' || user?.role === 'DIRECTOR';
  const canEdit = user?.role === 'ADMIN' || user?.role === 'FINANCE' || user?.role === 'DIRECTOR';
  const canDelete = user?.role === 'ADMIN';
  const canExport = user?.role === 'ADMIN' || user?.role === 'FINANCE' || user?.role === 'DIRECTOR';

  // New Payment handler
  const handleNewPayment = () => {
    setSelectedPayment(null);
    setShowPaymentForm(true);
  };
  
  // Edit Payment handler
  const handleEditPayment = (payment: Payment) => {
    setSelectedPayment(payment);
    setShowPaymentForm(true);
  };
  
  // View Payment handler
  const handleViewPayment = (payment: Payment) => {
    setSelectedPayment(payment);
    setShowPaymentDetail(true);
  };
  
  // Delete Payment handler
  const handleDeletePayment = (payment: Payment) => {
    setSelectedPayment(payment);
    setShowConfirmDelete(true);
  };

  const columns = [
  {
    header: 'Payment #',
    accessor: (row: Payment) => (
      <Text fontWeight="medium" color="blue.600">{row.code}</Text>
    )
  },
  { 
    header: 'Contact',
    accessor: (row: Payment) => row.contact?.name || '-'
  },
  {
    header: 'Date',
    accessor: (row: Payment) => formatDate(row.date)
  },
  {
    header: 'Amount',
    accessor: (row: Payment) => (
      <Text fontWeight="medium">{formatCurrency(row.amount)}</Text>
    )
  },
  {
    header: 'Method',
    accessor: (row: Payment) => paymentService.getMethodDisplayName(row.method)
  },
  {
    header: 'Status',
    accessor: (row: Payment) => (
      <Badge colorScheme={paymentService.getStatusColorScheme(row.status)} variant="subtle">
        {row.status}
      </Badge>
    )
  },
  {
    header: 'Actions',
    accessor: (row: Payment) => (
      <Menu>
        <MenuButton as={IconButton} icon={<FiMoreVertical />} variant="ghost" size="sm" />
        <MenuList>
          <MenuItem icon={<FiEye />} onClick={() => handleViewPayment(row)}>
            View Details
          </MenuItem>
          {canEdit && (
            <MenuItem icon={<FiEdit />} onClick={() => handleEditPayment(row)}>
              Edit
            </MenuItem>
          )}
          <MenuItem 
            icon={<FiFilePlus />} 
            onClick={() => paymentService.downloadPaymentDetailPDF(row.id, row.code)}
          >
            Export PDF
          </MenuItem>
          {canDelete && (
            <>
              <MenuDivider />
              <MenuItem icon={<FiTrash2 />} color="red.500" onClick={() => handleDeletePayment(row)}>
                Delete
              </MenuItem>
            </>
          )}
        </MenuList>
      </Menu>
    )
  }
];

// Load payments data
const loadPayments = async (newFilters?: Partial<PaymentFilters>) => {
  try {
    setLoading(true);
    setError(null);
    
    const currentFilters = newFilters ? { ...filters, ...newFilters } : filters;
    
    // Prepare filters for API request
    const apiFilters: PaymentFilters = {
      page: currentFilters.page,
      limit: currentFilters.limit
    };
    
    // Add status filter if selected
    if (statusFilter !== 'ALL') {
      apiFilters.status = statusFilter;
    }
    
    // Add method filter if selected
    if (methodFilter !== 'ALL') {
      apiFilters.method = methodFilter;
    }
    
    // Add date filters if selected
    if (startDate) {
      apiFilters.start_date = startDate;
    }
    
    if (endDate) {
      apiFilters.end_date = endDate;
    }
    
    // Make API call
    const result = await paymentService.getPayments(apiFilters);
    
    // Update state with results
    setPayments(result?.data || []);
    setFilters({ ...currentFilters, page: result?.page || currentFilters.page });
    setPagination({
      current: result?.page || 1,
      total: result?.total_pages || 1,
      totalItems: result?.total || 0
    });
    
  } catch (err: any) {
    console.error('Error fetching payments:', err);
    setError(err.message || 'An error occurred while fetching payments');
    setPayments([]);
    
    toast({
      title: 'Error loading payments',
      description: err.message || 'Failed to load payments data',
      status: 'error',
      duration: 3000
    });
  } finally {
    setLoading(false);
  }
};

// Load payment summary
const loadPaymentSummary = async () => {
  try {
    const totalAmount = payments.reduce((sum, payment) => sum + payment.amount, 0);
    const completedPayments = payments.filter(p => p.status === 'COMPLETED');
    const pendingPayments = payments.filter(p => p.status === 'PENDING');
    const completedAmount = completedPayments.reduce((sum, payment) => sum + payment.amount, 0);
    const pendingAmount = pendingPayments.reduce((sum, payment) => sum + payment.amount, 0);
    
    setSummary({
      total_payments: payments.length,
      total_amount: totalAmount,
      completed_amount: completedAmount,
      pending_amount: pendingAmount,
      completed_count: completedPayments.length,
      pending_count: pendingPayments.length,
      avg_payment_value: payments.length > 0 ? totalAmount / payments.length : 0
    });
  } catch (error) {
    console.error('Failed to calculate payment summary:', error);
  }
};

// Initial load
useEffect(() => {
  if (token) {
    loadPayments();
  }
}, [token]);

// Update summary when payments change
useEffect(() => {
  if (payments.length > 0) {
    loadPaymentSummary();
  }
}, [payments]);

// Handle filters change
useEffect(() => {
  if (token) {
    loadPayments();
  }
}, [statusFilter, methodFilter, startDate, endDate]);

// Handle page change
const handlePageChange = (page: number) => {
  setFilters(prev => ({
    ...prev,
    page: page
  }));
};

// Handle search
const handleSearch = (searchTerm: string) => {
  loadPayments({ search: searchTerm, page: 1 } as any);
};

// Handle filter change
const handleFilterChange = (key: string, value: string) => {
  switch(key) {
    case 'status':
      setStatusFilter(value as PaymentStatusType);
      break;
    case 'method':
      setMethodFilter(value as PaymentMethodType);
      break;
    case 'start_date':
      setStartDate(value);
      break;
    case 'end_date':
      setEndDate(value);
      break;
  }
};

// Apply filters
const applyFilters = () => {
  loadPayments({ page: 1 });
};

// Reset filters
const resetFilters = () => {
  setStatusFilter('ALL');
  setMethodFilter('ALL');
  setStartDate('');
  setEndDate('');
  setFilters({
    page: 1,
    limit: ITEMS_PER_PAGE
  });
  loadPayments({ page: 1 });
};

  // Handle delete payment
  const handleDelete = async (payment: Payment) => {
    if (!window.confirm('Are you sure you want to delete this payment?')) return;
    
    try {
      await paymentService.deletePayment(payment.id);
      toast({
        title: 'Success',
        description: 'Payment has been deleted successfully',
        status: 'success',
        duration: 3000
      });
      loadPayments();
    } catch (error: any) {
      toast({
        title: 'Error deleting payment',
        description: error.message || 'Failed to delete payment',
        status: 'error',
        duration: 3000
      });
    }
  };
  
  // Confirm delete
  const confirmDeletePayment = async () => {
    if (!selectedPayment) return;
    
    try {
      await paymentService.deletePayment(selectedPayment.id);
      toast({
        title: 'Success',
        description: 'Payment has been deleted successfully',
        status: 'success',
        duration: 3000
      });
      loadPayments();
      setShowConfirmDelete(false);
      setSelectedPayment(null);
    } catch (error: any) {
      console.error('Error deleting payment:', error);
      toast({
        title: 'Error deleting payment',
        description: error.message || 'Failed to delete payment',
        status: 'error',
        duration: 3000
      });
    }
  };
  
  // Export payments to Excel handler
  const handleExportPayments = async () => {
    try {
      // Prepare current filters for export
      const exportFilters: PaymentFilters = { ...filters };
      if (statusFilter !== 'ALL') exportFilters.status = statusFilter;
      if (methodFilter !== 'ALL') exportFilters.method = methodFilter;
      if (startDate) exportFilters.start_date = startDate;
      if (endDate) exportFilters.end_date = endDate;
      
      // Call export API
      const blob = await paymentService.exportPayments(exportFilters);
      
      // Create download link
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `payments-export-${new Date().toISOString().split('T')[0]}.xlsx`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (error: any) {
      console.error('Error exporting payments:', error);
      setError(error.message || 'Failed to export payments');
    }
  };

  // Export payments to PDF handler
  const handleExportPaymentsPDF = async () => {
    try {
      setError(null);
      
      // Prepare PDF export options
      const pdfOptions: PDFExportOptions = {
        title: 'Laporan Pembayaran',
        subtitle: 'Daftar Transaksi Pembayaran',
        companyName: 'PT. Sistem Akuntansi',
        companyAddress: 'Jakarta, Indonesia',
        includeFilters: true,
        statusFilter: statusFilter,
        methodFilter: methodFilter,
        startDate: startDate,
        endDate: endDate
      };
      
      // Export PDF with current data
      exportPaymentsToPDF(payments, pdfOptions);
      
    } catch (error: any) {
      console.error('Error exporting payments to PDF:', error);
      setError(error.message || 'Failed to export payments to PDF');
    }
  };
  
  // Handle bulk export (like sales page)
  const handleBulkExport = async () => {
    try {
      // Prepare filters for export
      const exportStatus = statusFilter !== 'ALL' ? statusFilter : undefined;
      const exportMethod = methodFilter !== 'ALL' ? methodFilter : undefined;
      
      // Use backend PDF export
      await paymentService.downloadPaymentReportPDF(
        startDate || undefined,
        endDate || undefined,
        exportStatus,
        exportMethod
      );
      
      toast({
        title: 'Success',
        description: 'Payment report has been downloaded',
        status: 'success',
        duration: 3000
      });
    } catch (error: any) {
      console.error('Error exporting payments to PDF:', error);
      toast({
        title: 'Export failed',
        description: 'Failed to export payment report',
        status: 'error',
        duration: 3000
      });
    }
  };
  
  // Refresh data handler
  const handleRefreshData = () => {
    loadPayments();
    loadPaymentSummary();
  };

  // Handle form save
  const handleFormSave = () => {
    setShowPaymentForm(false);
    setSelectedPayment(null);
    loadPayments();
  };

  // Handle form cancel
  const handleFormCancel = () => {
    setShowPaymentForm(false);
    setSelectedPayment(null);
    setError(null);
  };
  
  if (loading && (!payments || payments.length === 0)) {
    return (
      <Layout allowedRoles={['ADMIN', 'FINANCE', 'DIRECTOR']}>
        <Box display="flex" justifyContent="center" alignItems="center" height="400px">
          <Spinner size="xl" thickness="4px" speed="0.65s" color="brand.500" />
          <Text ml={4} fontSize="lg">Loading payments...</Text>
        </Box>
      </Layout>
    );
  }

  return (
    <Layout allowedRoles={['ADMIN', 'FINANCE', 'DIRECTOR']}>
      <VStack spacing={6} align="stretch">
        {/* Summary Cards */}
        {summary && (
          <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>Total Payments</StatLabel>
                  <StatNumber>{summary.total_payments}</StatNumber>
                  <StatHelpText>This period</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>Total Amount</StatLabel>
                  <StatNumber>{formatCurrency(summary.total_amount)}</StatNumber>
                  <StatHelpText>Gross amount</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>Completed</StatLabel>
                  <StatNumber color="green.500">
                    {formatCurrency(summary.completed_amount)}
                  </StatNumber>
                  <StatHelpText>{summary.completed_count} payments</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>Avg Payment Value</StatLabel>
                  <StatNumber>{formatCurrency(summary.avg_payment_value)}</StatNumber>
                  <StatHelpText>Per transaction</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </SimpleGrid>
        )}

        {/* Header */}
        <Flex justify="space-between" align="center">
          <Box>
            <Heading as="h1" size="xl" mb={2}>Payment Management</Heading>
            <Text color="gray.600">Manage your payment transactions</Text>
          </Box>
          <HStack spacing={3}>
            <Button
              leftIcon={<FiRefreshCw />}
              variant="ghost"
              size="md"
              onClick={handleRefreshData}
              isLoading={loading}
            >
              Refresh
            </Button>
            {canExport && (
              <Menu>
                <MenuButton 
                  as={Button} 
                  leftIcon={<FiDownload />}
                  colorScheme="green"
                  variant="outline"
                  size="md"
                >
                  Export Report
                </MenuButton>
                <MenuList>
                  <MenuItem icon={<FiFileText />} onClick={handleBulkExport}>
                    Export PDF Report
                  </MenuItem>
                  <MenuItem icon={<FiDownload />} onClick={async () => {
                    try {
                      // Prepare current filters for export
                      const exportFilters: any = {};
                      if (statusFilter !== 'ALL') exportFilters.status = statusFilter;
                      if (methodFilter !== 'ALL') exportFilters.method = methodFilter;
                      if (startDate) exportFilters.start_date = startDate;
                      if (endDate) exportFilters.end_date = endDate;
                      
                      // Call Excel export API
                      await paymentService.downloadPaymentReportExcel(
                        startDate || undefined,
                        endDate || undefined,
                        exportFilters.status,
                        exportFilters.method
                      );
                      
                      toast({
                        title: 'Success',
                        description: 'Payment report has been downloaded as Excel',
                        status: 'success',
                        duration: 3000
                      });
                    } catch (error: any) {
                      console.error('Error exporting payments to Excel:', error);
                      toast({
                        title: 'Export failed',
                        description: 'Failed to export payment report as Excel',
                        status: 'error',
                        duration: 3000
                      });
                    }
                  }}>
                    Export Excel Report
                  </MenuItem>
                </MenuList>
              </Menu>
            )}
            {canCreate && (
              <Button 
                leftIcon={<FiPlus />} 
                colorScheme="blue"
                size="md"
                onClick={handleNewPayment}
              >
                Create Payment
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
                  placeholder="Search by payment code or contact..."
                  onChange={(e) => handleSearch(e.target.value)}
                />
              </InputGroup>
              
              <Select 
                maxW="200px" 
                placeholder="All Status"
                value={statusFilter}
                onChange={(e) => handleFilterChange('status', e.target.value)}
              >
                <option value="PENDING">Pending</option>
                <option value="COMPLETED">Completed</option>
                <option value="FAILED">Failed</option>
              </Select>
              
              <Select
                maxW="200px"
                placeholder="All Methods"
                value={methodFilter}
                onChange={(e) => handleFilterChange('method', e.target.value)}
              >
                <option value="CASH">Cash</option>
                <option value="BANK_TRANSFER">Bank Transfer</option>
                <option value="CHECK">Check</option>
                <option value="CREDIT_CARD">Credit Card</option>
                <option value="DEBIT_CARD">Debit Card</option>
              </Select>
              
              <Input
                type="date"
                maxW="200px"
                placeholder="Start Date"
                value={startDate}
                onChange={(e) => handleFilterChange('start_date', e.target.value)}
              />
              
              <Input
                type="date"
                maxW="200px"
                placeholder="End Date"
                value={endDate}
                onChange={(e) => handleFilterChange('end_date', e.target.value)}
              />
            </HStack>
          </CardBody>
        </Card>

        {/* Error Alert */}
        {error && (
          <Alert status="error">
            <AlertIcon />
            {error}
          </Alert>
        )}

        {/* Payments Table */}
        <Card>
          <CardHeader>
            <Flex justify="space-between" align="center">
              <Heading size="md">Payment Transactions ({payments?.length || 0})</Heading>
            </Flex>
          </CardHeader>
          <CardBody>
            {loading ? (
              <Flex justify="center" py={10}>
                <Spinner size="lg" />
              </Flex>
            ) : (
              <DataTable 
                columns={columns} 
                data={payments || []} 
                keyField="id"
                searchable={false}
                pagination={true}
                pageSize={ITEMS_PER_PAGE}
              />
            )}
          </CardBody>
        </Card>
      </VStack>

      {/* Payment Form Modal */}
      <AdvancedPaymentForm
        isOpen={showPaymentForm}
        onClose={handleFormCancel}
        type="receivable" // Default to receivable, could be made dynamic
        onSuccess={handleFormSave}
        preSelectedContact={selectedPayment ? { 
          id: selectedPayment.contact_id,
          name: selectedPayment.contact?.name || 'Unknown'
        } : undefined}
      />

      {/* Delete Confirmation Dialog */}
      <AlertDialog
        isOpen={showConfirmDelete}
        leastDestructiveRef={undefined}
        onClose={() => setShowConfirmDelete(false)}
      >
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader fontSize="lg" fontWeight="bold">
              Delete Payment
            </AlertDialogHeader>

            <AlertDialogBody>
              Are you sure you want to delete this payment?
              {selectedPayment && (
                <Box mt={3} p={3} bg="red.50" borderRadius="md">
                  <Text fontSize="sm" fontWeight="bold">Payment: {selectedPayment.code}</Text>
                  <Text fontSize="sm">Amount: {formatCurrency(selectedPayment.amount)}</Text>
                  <Text fontSize="sm" color="red.600">This action cannot be undone.</Text>
                </Box>
              )}
            </AlertDialogBody>

            <AlertDialogFooter>
              <Button onClick={() => setShowConfirmDelete(false)}>
                Cancel
              </Button>
              <Button colorScheme="red" onClick={confirmDeletePayment} ml={3}>
                Delete
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>
      {/* Payment Detail Modal */}
      <PaymentDetailModal
        payment={selectedPayment}
        isOpen={showPaymentDetail}
        onClose={() => {
          setShowPaymentDetail(false);
          setSelectedPayment(null);
        }}
      />
    </Layout>
  );
};

export default PaymentsPage;
