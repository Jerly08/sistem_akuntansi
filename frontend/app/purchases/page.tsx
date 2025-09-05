'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import SimpleLayout from '@/components/layout/SimpleLayout';
import { DataTable } from '@/components/common/DataTable';
import EnhancedPurchaseTable from '@/components/purchase/EnhancedPurchaseTable';
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
  Spinner,
  useToast,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  useDisclosure,
  Select,
  Input,
  FormControl,
  FormLabel,
  Grid,
  GridItem,
  Card,
  CardBody,
  CardHeader,
  Stat,
  StatLabel,
  StatNumber,
  Textarea,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  IconButton,
  Divider,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
  SimpleGrid,
  FormHelperText,
} from '@chakra-ui/react';
import { 
  FiPlus, 
  FiEye, 
  FiEdit, 
  FiTrash2, 
  FiFilter,
  FiRefreshCw,
  FiCheckCircle,
  FiClock,
  FiXCircle,
  FiAlertCircle 
} from 'react-icons/fi';
import purchaseService, { Purchase, PurchaseFilterParams } from '@/services/purchaseService';
import SubmitApprovalButton from '@/components/purchase/SubmitApprovalButton';
import { ApprovalPanel } from '@/components/approval/ApprovalPanel';
import contactService from '@/services/contactService';
import productService, { Product } from '@/services/productService';
import accountService from '@/services/accountService';
import { Account as GLAccount, AccountCatalogItem } from '@/types/account';
import approvalService from '@/services/approvalService';
import { normalizeRole } from '@/utils/roles';
import { useColorModeValue } from '@chakra-ui/react';
import SearchableSelect from '@/components/common/SearchableSelect';
import CurrencyInput from '@/components/common/CurrencyInput';

// Types for form data
interface PurchaseFormData {
  vendor_id: string;
  date: string;
  due_date: string;
  notes: string;
  discount: string;
  
  // Legacy tax field (backward compatibility)
  tax: string;
  
  // Tax additions (Penambahan)
  ppn_rate: string;
  other_tax_additions: string;
  
  // Tax deductions (Pemotongan)
  pph21_rate: string;
  pph23_rate: string;
  other_tax_deductions: string;
  
  // Payment method fields
  payment_method: string;
  bank_account_id: string;
  credit_account_id: string;  // New field for liability account
  payment_reference: string;
  
  items: PurchaseItemFormData[];
}

interface PurchaseItemFormData {
  product_id: string;
  quantity: string;
  unit_price: string;
  discount: string;
  tax: string;
  expense_account_id: string;
}

interface Vendor {
  id: number;
  name: string;
  code: string;
}

interface BankAccount {
  id: number;
  name: string;
  code: string;
  type: string;
  balance?: number;
  currency: string;
}

// Status color mapping
const getStatusColor = (status: string) => {
  switch (status.toLowerCase()) {
    case 'approved':
    case 'completed':
      return 'green';
    case 'draft':
    case 'pending_approval':
      return 'yellow';
    case 'pending':
      return 'blue';
    case 'cancelled':
    case 'rejected':
      return 'red';
    default:
      return 'gray';
  }
};

// Approval status color mapping
const getApprovalStatusColor = (approvalStatus: string) => {
  switch ((approvalStatus || '').toLowerCase()) {
    case 'approved':
      return 'green';
    case 'pending':
      return 'yellow';
    case 'rejected':
      return 'red';
    case 'not_required':
    case 'not_started':
      return 'gray';
    default:
      return 'gray';
  }
};

// Format currency in IDR
const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount);
};

// Format date
const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('id-ID');
};

const columns = [
  { header: 'Purchase #', accessor: 'code' as keyof Purchase },
  { 
    header: 'Vendor', 
    accessor: ((row: Purchase) => {
      return row.vendor?.name || 'N/A';
    }) as (row: Purchase) => React.ReactNode
  },
  { 
    header: 'Date', 
    accessor: ((row: Purchase) => {
      return new Date(row.date).toLocaleDateString('id-ID');
    }) as (row: Purchase) => React.ReactNode
  },
  { 
    header: 'Total', 
    accessor: ((row: Purchase) => {
      return formatCurrency(row.total_amount);
    }) as (row: Purchase) => React.ReactNode
  },
  { 
    header: 'Status', 
    accessor: ((row: Purchase) => (
      <Badge colorScheme={getStatusColor(row.status)} variant="subtle">
        {row.status.replace('_', ' ').toUpperCase()}
      </Badge>
    )) as (row: Purchase) => React.ReactNode
  },
  { 
    header: 'Approval Status', 
    accessor: ((row: Purchase) => (
      <Badge colorScheme={getApprovalStatusColor(row.approval_status)} variant="subtle">
        {row.approval_status.replace('_', ' ').toUpperCase()}
      </Badge>
    )) as (row: Purchase) => React.ReactNode
  },
];

const PurchasesPage: React.FC = () => {
  const { token, user } = useAuth();
  const toast = useToast();
  const { isOpen: isFilterOpen, onOpen: onFilterOpen, onClose: onFilterClose } = useDisclosure();
  
  // Theme colors for enhanced styling - MUST be at top to follow Rules of Hooks
  const bgColor = useColorModeValue('gray.50', 'gray.900');
  const cardBg = useColorModeValue('white', 'gray.800');
  const headingColor = useColorModeValue('gray.800', 'gray.100');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  
  // Pre-calculate all useColorModeValue calls to avoid conditional hook calls
  const textSecondary = useColorModeValue('gray.600', 'gray.400');
  const textPrimary = useColorModeValue('gray.700', 'gray.200');
  const hoverBg = useColorModeValue('gray.50', 'gray.700');
  const hoverBorder = useColorModeValue('gray.300', 'gray.500');
  const buttonBlueBg = useColorModeValue('blue.500', 'blue.600');
  const buttonBlueHover = useColorModeValue('blue.600', 'blue.500');
  const statColors = {
    orange: useColorModeValue('orange.600', 'orange.400'),
    green: useColorModeValue('green.600', 'green.400'),
    red: useColorModeValue('red.600', 'red.400'),
    purple: useColorModeValue('purple.600', 'purple.400'),
    blue: useColorModeValue('blue.600', 'blue.300')
  };
  const statBgColors = {
    orange: useColorModeValue('orange.50', 'orange.900'),
    green: useColorModeValue('green.50', 'green.900'),
    red: useColorModeValue('red.50', 'red.900'),
    purple: useColorModeValue('purple.50', 'purple.900'),
    blue: useColorModeValue('blue.50', 'blue.900')
  };
  const modalBg = useColorModeValue('white', 'gray.900');
  const modalFilterBg = useColorModeValue('blue.50', 'blue.900');
  const modalFilterColor = useColorModeValue('blue.600', 'blue.300');
  const modalHoverBg = useColorModeValue('gray.100', 'gray.700');
  const inputHoverBorder = useColorModeValue('gray.300', 'gray.500');
  const inputFocusBorder = useColorModeValue('blue.500', 'blue.400');
  const inputFocusShadow = `0 0 0 1px ${useColorModeValue('blue.500', 'blue.400')}`;
  const ghostHoverBg = useColorModeValue('gray.50', 'gray.700');
  
  // Additional colors for view modal
  const modalContentBg = useColorModeValue('white', 'gray.800');
  const modalHeaderBg = useColorModeValue('gray.50', 'gray.700');
  const notesBoxBg = useColorModeValue('gray.50', 'gray.600');
  const tableBg = useColorModeValue('white', 'gray.700');
  const tableHeaderBg = useColorModeValue('gray.50', 'gray.600');
  
  // State management
  const [purchases, setPurchases] = useState<Purchase[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState({
    page: 1,
    limit: 10,
    total: 0,
    totalPages: 0,
  });
  
  // Filter state
  const [filters, setFilters] = useState<PurchaseFilterParams>({
    status: '',
    approval_status: '',
    search: '',
    page: 1,
    limit: 10,
  });
  
  // Statistics state
  const [stats, setStats] = useState({
    total: 0,
    pending: 0,
    approved: 0,
    rejected: 0,
    needingApproval: 0,
    totalValue: 0,
    totalApprovedAmount: 0,
  });

  // View and Edit Modal states
  const { isOpen: isViewOpen, onOpen: onViewOpen, onClose: onViewClose } = useDisclosure();
  const { isOpen: isEditOpen, onOpen: onEditOpen, onClose: onEditClose } = useDisclosure();
  const { isOpen: isCreateOpen, onOpen: onCreateOpen, onClose: onCreateClose } = useDisclosure();
  
  const [selectedPurchase, setSelectedPurchase] = useState<Purchase | null>(null);
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [expenseAccounts, setExpenseAccounts] = useState<GLAccount[]>([]);
  const [bankAccounts, setBankAccounts] = useState<BankAccount[]>([]);
  const [creditAccounts, setCreditAccounts] = useState<GLAccount[]>([]);  // New state for liability accounts
  const [loadingExpenseAccounts, setLoadingExpenseAccounts] = useState(false);
  const [defaultExpenseAccountId, setDefaultExpenseAccountId] = useState<number | null>(null);
  const [canListExpenseAccounts, setCanListExpenseAccounts] = useState(true);
  const [loadingBankAccounts, setLoadingBankAccounts] = useState(false);
  const [loadingCreditAccounts, setLoadingCreditAccounts] = useState(false);  // New loading state
  const [formData, setFormData] = useState<PurchaseFormData>({
    vendor_id: '',
    date: new Date().toISOString().split('T')[0],
    due_date: '',
    notes: '',
    discount: '0',
    
    // Legacy tax field
    tax: '0',
    
    // Tax additions (Penambahan)
    ppn_rate: '11',
    other_tax_additions: '0',
    
    // Tax deductions (Pemotongan)
    pph21_rate: '0',
    pph23_rate: '0', 
    other_tax_deductions: '0',
    
    // Payment method fields
    payment_method: 'CREDIT',
    bank_account_id: '',
    credit_account_id: '',  // New field for liability account
    payment_reference: '',
    
    items: []
  });
  const [loadingVendors, setLoadingVendors] = useState(false);
  const [loadingProducts, setLoadingProducts] = useState(false);
  
  // Add Vendor Modal states
  const { isOpen: isAddVendorOpen, onOpen: onAddVendorOpen, onClose: onAddVendorClose } = useDisclosure();
  const [newVendorData, setNewVendorData] = useState({
    name: '',
    code: '',
    email: '',
    phone: '',
    mobile: '',
    address: '',
    pic_name: '',
    external_id: '',
    notes: ''
  });
  const [savingVendor, setSavingVendor] = useState(false);

  // Add Product Modal states
  const { isOpen: isAddProductOpen, onOpen: onAddProductOpen, onClose: onAddProductClose } = useDisclosure();
  const [newProductData, setNewProductData] = useState({
    name: '',
    code: '',
    description: '',
    unit: '',
    purchase_price: '0',
    sale_price: '0',
  });
  const [savingProduct, setSavingProduct] = useState(false);

  // Role-based permissions
  const roleNorm = normalizeRole(user?.role as any);
  const canEdit = roleNorm === 'employee' || roleNorm === 'admin';
  const canDelete = roleNorm === 'admin';

  // Helper function to notify directors
  const notifyDirectors = async (purchase: Purchase) => {
    try {
      // This would typically call a notification service API
      // For now, we'll just log it as the notification is handled by backend
      console.log('Director notification sent for purchase:', purchase.code);
    } catch (err) {
      console.error('Error sending director notification:', err);
    }
  };

  // Handle add new vendor
  const handleAddVendor = async () => {
    if (!newVendorData.name.trim()) {
      toast({
        title: 'Error',
        description: 'Vendor name is required',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    if (!newVendorData.email.trim()) {
      toast({
        title: 'Error',
        description: 'Vendor email is required',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    try {
      setSavingVendor(true);
      
      // Generate unique vendor code if not provided
      let vendorCode = newVendorData.code.trim();
      if (!vendorCode) {
        // Generate code based on name + timestamp + random to ensure uniqueness
        const namePrefix = newVendorData.name.trim().substring(0, 3).toUpperCase().replace(/[^A-Z]/g, 'X');
        const timestamp = Date.now().toString().slice(-6); // Last 6 digits of timestamp
        const random = Math.floor(Math.random() * 1000).toString().padStart(3, '0'); // 3-digit random
        vendorCode = `V${namePrefix}${timestamp}${random}`;
      } else {
        // Check if manually entered code already exists in current vendors list
        const existingVendor = vendors.find(v => v.code.toLowerCase() === vendorCode.toLowerCase());
        if (existingVendor) {
          toast({
            title: 'Error',
            description: `Vendor code "${vendorCode}" already exists. Please use a different code or leave empty for auto-generation.`,
            status: 'error',
            duration: 5000,
            isClosable: true,
          });
          return;
        }
      }
      
      const vendorPayload = {
        ...newVendorData,
        code: vendorCode,
        type: 'VENDOR',
        is_active: true
      };
      
      console.log('Creating vendor with payload:', vendorPayload);
      
      let newVendor;
      try {
        newVendor = await contactService.createContact(token!, vendorPayload);
        console.log('Vendor creation response:', newVendor);
        
        // Check if the response indicates an error (some APIs return error in success response)
        if (newVendor && typeof newVendor === 'object' && 'error' in newVendor) {
          throw new Error(newVendor.error as string || 'Server returned an error');
        }
        
      } catch (createError: any) {
        console.error('API Error creating vendor:', createError);
        throw new Error(
          createError.message || 
          createError.response?.data?.error || 
          'Failed to create vendor: Server error'
        );
      }
      
      // Validate that the new vendor was created successfully
      // Handle different response structures
      let vendorData = newVendor;
      if (newVendor?.data) {
        vendorData = newVendor.data; // If response is wrapped in data object
      }
      
      // Additional checks for undefined response
      if (!newVendor) {
        console.error('Vendor creation returned undefined response');
        throw new Error('Failed to create vendor: Server returned no response. Please try again.');
      }
      
      if (!vendorData || (!vendorData.id && !vendorData.ID)) {
        console.error('Invalid vendor response:', newVendor);
        console.error('Expected vendor data with id field, got:', vendorData);
        throw new Error('Failed to create vendor: Invalid response structure from server. Please check console for details.');
      }
      
      // Use the validated vendor data
      const vendorId = vendorData.id || vendorData.ID;
      const vendorName = vendorData.name || vendorData.Name;
      const finalVendorCode = vendorData.code || vendorData.Code || `V${vendorId}`;
      
      // Add the new vendor to the vendors list
      const formattedVendor = {
        id: vendorId,
        name: vendorName,
        code: finalVendorCode,
      };
      
      console.log('Adding formatted vendor to list:', formattedVendor);
      setVendors(prev => [...prev, formattedVendor]);
      
      // Select the new vendor in the form
      setFormData(prev => ({ ...prev, vendor_id: vendorId.toString() }));
      
      // Reset form and close modal
      setNewVendorData({
        name: '',
        code: '',
        email: '',
        phone: '',
        mobile: '',
        address: '',
        pic_name: '',
        external_id: '',
        notes: ''
      });
      
      onAddVendorClose();
      
      toast({
        title: 'Success',
        description: `Vendor "${vendorName}" created successfully and selected`,
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      
    } catch (err: any) {
      console.error('Error creating vendor:', err);
      toast({
        title: 'Error',
        description: err.response?.data?.error || 'Failed to create vendor',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setSavingVendor(false);
    }
  };

  // Handle add new product
  const handleAddProduct = async () => {
    if (!newProductData.name.trim()) {
      toast({
        title: 'Error',
        description: 'Product name is required',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    if (!newProductData.unit.trim()) {
      toast({
        title: 'Error',
        description: 'Product unit is required',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    try {
      setSavingProduct(true);
      
      const productPayload = {
        name: newProductData.name,
        code: newProductData.code || undefined,
        description: newProductData.description || undefined,
        unit: newProductData.unit,
        purchase_price: parseFloat(newProductData.purchase_price) || 0,
        sale_price: parseFloat(newProductData.sale_price) || 0,
        stock: 0,
        min_stock: 0,
        max_stock: 0,
        reorder_level: 0,
        is_active: true,
        is_service: false,
        taxable: true
      };
      
      const newProduct = await productService.createProduct(productPayload);
      
      // Add the new product to the products list
      setProducts(prev => [...prev, newProduct.data]);
      
      // Select the new product in the form if we have items
      if (formData.items.length > 0) {
        const items = [...formData.items];
        items[0] = { 
          ...items[0], 
          product_id: newProduct.data.id.toString(),
          unit_price: newProduct.data.purchase_price?.toString() || '0'
        };
        setFormData({ ...formData, items });
      } else {
        // Add a new item with the created product
        setFormData({
          ...formData,
          items: [{
            product_id: newProduct.data.id.toString(),
            quantity: '1',
            unit_price: newProduct.data.purchase_price?.toString() || '0',
            discount: '0',
            tax: '0',
            expense_account_id: defaultExpenseAccountId ? defaultExpenseAccountId.toString() : ''
          }]
        });
      }
      
      // Reset form and close modal
      setNewProductData({
        name: '',
        code: '',
        description: '',
        unit: '',
        purchase_price: '0',
        sale_price: '0',
      });
      
      onAddProductClose();
      
      toast({
        title: 'Success',
        description: `Product "${newProduct.data.name}" created successfully and selected`,
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      
    } catch (err: any) {
      console.error('Error creating product:', err);
      toast({
        title: 'Error',
        description: err.response?.data?.error || 'Failed to create product',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setSavingProduct(false);
    }
  };

  // Fetch purchases from API
  const fetchPurchases = async (filterParams: PurchaseFilterParams = filters) => {
    if (!token) return;
    
    try {
      setLoading(true);
      const response = await purchaseService.list(filterParams);
      
      // Ensure response data is an array
      const purchaseData = Array.isArray(response?.data) ? response.data : [];
      
      setPurchases(purchaseData);
      setPagination({
        page: response?.page || 1,
        limit: response?.limit || 10,
        total: response?.total || 0,
        totalPages: response?.total_pages || 0,
      });
      
      // Calculate stats with correct logic for approval status
      // Note: We use purchaseData.length for total since pagination affects response.total
      const totalPurchases = purchaseData.length;
      const pendingApproval = purchaseData.filter(p => {
        const approvalStatus = (p?.approval_status || '').toUpperCase();
        const status = (p?.status || '').toUpperCase();
        // Pending approval includes: PENDING approval status, or purchases requiring approval that haven't been approved/rejected
        return approvalStatus === 'PENDING' || 
               (!!p?.requires_approval && approvalStatus !== 'APPROVED' && approvalStatus !== 'REJECTED' && status !== 'CANCELLED');
      }).length;
      
      const approved = purchaseData.filter(p => (p?.approval_status || '').toUpperCase() === 'APPROVED').length;
      const rejected = purchaseData.filter(p => {
        const approvalStatus = (p?.approval_status || '').toUpperCase();
        const status = (p?.status || '').toUpperCase();
        return approvalStatus === 'REJECTED' || status === 'CANCELLED';
      }).length;
      
      // Calculate total value from current page data
      const totalValue = purchaseData.reduce((sum, p) => {
        const amount = p?.total_amount || 0;
        return sum + (typeof amount === 'number' ? amount : parseFloat(amount) || 0);
      }, 0);
      
      // Fetch purchase summary to get approved amount
      let totalApprovedAmount = 0;
      try {
        const summaryResponse = await purchaseService.getSummary();
        totalApprovedAmount = summaryResponse.total_approved_amount || 0;
      } catch (summaryErr) {
        console.warn('Failed to fetch purchase summary:', summaryErr);
        // Continue without approved amount if summary fetch fails
      }
      
      setStats({
        total: response?.total || totalPurchases, // Use API total if available, otherwise current page count
        pending: pendingApproval,
        approved: approved,
        rejected: rejected,
        needingApproval: pendingApproval, // Same as pending for now
        totalValue: totalValue, // Add total value to stats
        totalApprovedAmount: totalApprovedAmount, // Add approved amount from summary
      });
      
      setError(null);
    } catch (err: any) {
      console.error('Error fetching purchases:', err);
      
      // Set empty state on error
      setPurchases([]);
      setPagination({
        page: 1,
        limit: 10,
        total: 0,
        totalPages: 0,
      });
      
      setStats({
        total: 0,
        pending: 0,
        approved: 0,
        rejected: 0,
        needingApproval: 0,
        totalValue: 0,
        totalApprovedAmount: 0,
      });
      
      const errorMessage = err.response?.data?.message || err.message || 'Failed to fetch purchases';
      setError(errorMessage);
      
      toast({
        title: 'Error',
        description: 'Failed to fetch purchase data. Please check your connection and try again.',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPurchases();
  }, [token]);

  // Handle filter changes
  const handleFilterChange = (newFilters: Partial<PurchaseFilterParams>) => {
    const updatedFilters = { ...filters, ...newFilters, page: 1 };
    setFilters(updatedFilters);
    fetchPurchases(updatedFilters);
  };

  // Handle page change
  const handlePageChange = (page: number) => {
    const updatedFilters = { ...filters, page };
    setFilters(updatedFilters);
    fetchPurchases(updatedFilters);
  };

  // Handle refresh
  const handleRefresh = () => {
    fetchPurchases();
    toast({
      title: 'Refreshed',
      description: 'Purchase data has been refreshed',
      status: 'info',
      duration: 2000,
      isClosable: true,
    });
  };

  // Handle purchase submission for approval
  const handleSubmitForApproval = async (purchaseId: number) => {
    try {
      await purchaseService.submitForApproval(purchaseId);
      await fetchPurchases(); // Refresh data
      toast({
        title: 'Success',
        description: 'Purchase submitted for approval',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (err: any) {
      toast({
        title: 'Error',
        description: err.response?.data?.error || 'Failed to submit purchase for approval',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  // Handle delete purchase
  const handleDelete = async (purchaseId: number) => {
    // Find the purchase to check its status
    const purchaseToDelete = purchases.find(p => p.id === purchaseId);
    const isApproved = purchaseToDelete && (purchaseToDelete.status || '').toUpperCase() === 'APPROVED';
    
    const confirmMessage = isApproved 
      ? 'WARNING: This purchase is APPROVED. Are you sure you want to delete this approved purchase? This action cannot be undone.'
      : 'Are you sure you want to delete this purchase?';
    
    if (!confirm(confirmMessage)) return;
    
    try {
      await purchaseService.delete(purchaseId);
      await fetchPurchases(); // Refresh data
      toast({
        title: 'Success',
        description: `Purchase ${isApproved ? '(APPROVED)' : ''} deleted successfully`,
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (err: any) {
      toast({
        title: 'Error',
        description: err.response?.data?.error || 'Failed to delete purchase',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  // Handle view purchase
  const handleView = async (purchase: Purchase) => {
    try {
      // Fetch detailed purchase data
      const detailResponse = await purchaseService.getById(purchase.id);
      setSelectedPurchase(detailResponse);
      onViewOpen();
    } catch (err: any) {
      toast({
        title: 'Error',
        description: 'Failed to fetch purchase details',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    }
  };

  // Fetch vendors
  const fetchVendors = async () => {
    if (!token) return;
    
    try {
      setLoadingVendors(true);
      
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/contacts?type=VENDOR`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch vendors');
      }

      const vendorsData = await response.json();
      
      // Transform the data to match our Vendor interface
      const formattedVendors = vendorsData.map((vendor: any) => ({
        id: vendor.id,
        name: vendor.name,
        code: vendor.code || `V${vendor.id}`,
      }));
      
      setVendors(formattedVendors);
    } catch (err: any) {
      console.error('Error fetching vendors:', err);
      toast({
        title: 'Error',
        description: 'Failed to fetch vendors',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    } finally {
      setLoadingVendors(false);
    }
  };

  // Fetch products for dropdown
  const fetchProductsList = async () => {
    try {
      setLoadingProducts(true);
      const data = await productService.getProducts({ page: 1, limit: 1000 });
      const list: Product[] = Array.isArray(data?.data) ? data.data : [];
      setProducts(list);
    } catch (err: any) {
      console.error('Error fetching products:', err);
      toast({
        title: 'Error',
        description: 'Failed to fetch products',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    } finally {
      setLoadingProducts(false);
    }
  };

  // Fetch bank accounts for payment method selection
  const fetchBankAccounts = async () => {
    if (!token) return;
    try {
      setLoadingBankAccounts(true);
      
      const response = await fetch('http://localhost:8080/api/v1/cashbank/payment-accounts', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch bank accounts');
      }

      const data = await response.json();
      
      // API returns { success: true, data: [...] }
      // The data array already contains both bank and cash accounts
      const allAccounts = data.data || [];
      
      setBankAccounts(allAccounts);
    } catch (err: any) {
      console.error('Error fetching bank accounts:', err);
      toast({
        title: 'Error',
        description: 'Failed to fetch bank accounts',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    } finally {
      setLoadingBankAccounts(false);
    }
  };

  // Fetch credit accounts (liability) for credit payment method selection
  const fetchCreditAccounts = async () => {
    if (!token) return;
    try {
      setLoadingCreditAccounts(true);
      
      // Try catalog endpoint first for EMPLOYEE role, fallback to regular endpoint
      if (user?.role === 'EMPLOYEE') {
        try {
          const catalogData = await accountService.getAccountCatalog(token, 'LIABILITY');
          const formattedAccounts: GLAccount[] = catalogData.map(item => ({
            id: item.id,
            code: item.code,
            name: item.name,
            type: 'LIABILITY' as const,
            is_active: item.active,
            level: 1,
            is_header: false,
            balance: 0,
            created_at: '',
            updated_at: '',
            description: '',
          }));
          console.log('Formatted credit accounts from catalog:', formattedAccounts);
          setCreditAccounts(formattedAccounts);
          return; // Success, exit early
        } catch (catalogError: any) {
          console.log('Catalog endpoint not available, trying regular endpoint:', catalogError.message);
          // Fall through to try regular endpoint
        }
      }
      
      // Use full account data for other roles or as fallback for EMPLOYEE
      try {
        const data = await accountService.getAccounts(token, 'LIABILITY');
        const list: GLAccount[] = Array.isArray(data) ? data : [];
        console.log('Formatted credit accounts from regular endpoint:', list);
        setCreditAccounts(list);
      } catch (regularError: any) {
        console.error('Regular accounts endpoint also failed:', regularError);
        throw regularError; // Re-throw to be caught by outer catch
      }
    } catch (err: any) {
      console.error('Error fetching credit accounts:', err);
      // If both endpoints fail, fall back to manual entry mode
      setCreditAccounts([]);
      
      // Only show warning for non-EMPLOYEE users or if it's not a permission error
      if (user?.role !== 'EMPLOYEE' || !err.message?.includes('Insufficient permissions')) {
        toast({
          title: 'Limited Access',
          description: 'Unable to load credit accounts list. Credit payment will use default liability account.',
          status: 'warning',
          duration: 5000,
          isClosable: true,
        });
      }
    } finally {
      setLoadingCreditAccounts(false);
    }
  };

  // Fetch expense accounts (GL) for item expense_account_id
  const fetchExpenseAccounts = async () => {
    if (!token) return;
    try {
      setLoadingExpenseAccounts(true);
      
      // Try catalog endpoint first for EMPLOYEE role, fallback to regular endpoint
      if (user?.role === 'EMPLOYEE') {
        try {
          const catalogData = await accountService.getAccountCatalog(token, 'EXPENSE');
          const formattedAccounts: GLAccount[] = catalogData.map(item => ({
            id: item.id,
            code: item.code,
            name: item.name,
            type: 'EXPENSE' as const,
            is_active: item.active,
            level: 1,
            is_header: false,
            balance: 0,
            created_at: '',
            updated_at: '',
            description: '',
          }));
          console.log('Formatted expense accounts from catalog:', formattedAccounts);
          setExpenseAccounts(formattedAccounts);
          setCanListExpenseAccounts(true);
          if (formattedAccounts.length > 0) {
            setDefaultExpenseAccountId(formattedAccounts[0].id as number);
          }
          return; // Success, exit early
        } catch (catalogError: any) {
          console.log('Catalog endpoint not available, trying regular endpoint:', catalogError.message);
          // Fall through to try regular endpoint
        }
      }
      
      // Use full account data for other roles or as fallback for EMPLOYEE
      try {
        const data = await accountService.getAccounts(token, 'EXPENSE');
        const list: GLAccount[] = Array.isArray(data) ? data : [];
        console.log('Formatted expense accounts from regular endpoint:', list);
        setExpenseAccounts(list);
        setCanListExpenseAccounts(true);
        if (list.length > 0) {
          setDefaultExpenseAccountId(list[0].id as number);
        }
      } catch (regularError: any) {
        console.error('Regular accounts endpoint also failed:', regularError);
        throw regularError; // Re-throw to be caught by outer catch
      }
    } catch (err: any) {
      console.error('Error fetching expense accounts:', err);
      // If both endpoints fail, fall back to manual entry mode
      setCanListExpenseAccounts(false);
      setExpenseAccounts([]);
      setDefaultExpenseAccountId(null);
      
      // Only show warning for non-EMPLOYEE users or if it's not a permission error
      if (user?.role !== 'EMPLOYEE' || !err.message?.includes('Insufficient permissions')) {
        toast({
          title: 'Limited Access',
          description: 'Unable to load expense accounts list. You can enter Expense Account ID manually in the items.',
          status: 'warning',
          duration: 5000,
          isClosable: true,
        });
      }
    } finally {
      setLoadingExpenseAccounts(false);
    }
  };

  // Handle edit purchase
  const handleEdit = async (purchase: Purchase) => {
    try {
      // Fetch detailed purchase data for editing
      const detailResponse = await purchaseService.getById(purchase.id);
      setSelectedPurchase(detailResponse);
      
      // Set form data for editing
      setFormData({
        vendor_id: detailResponse.vendor_id?.toString() || '',
        date: detailResponse.date.split('T')[0], // Format for date input
        due_date: detailResponse.due_date ? detailResponse.due_date.split('T')[0] : '',
        notes: detailResponse.notes || '',
        discount: detailResponse.discount?.toString() || '0',
        tax: detailResponse.tax?.toString() || '0',
        // Tax additions (Penambahan)
        ppn_rate: detailResponse.ppn_rate?.toString() || '11',
        other_tax_additions: detailResponse.other_tax_additions?.toString() || '0',
        // Tax deductions (Pemotongan)
        pph21_rate: detailResponse.pph21_rate?.toString() || '0',
        pph23_rate: detailResponse.pph23_rate?.toString() || '0',
        other_tax_deductions: detailResponse.other_tax_deductions?.toString() || '0',
        // Payment method fields
        payment_method: detailResponse.payment_method || 'CREDIT',
        bank_account_id: detailResponse.bank_account_id?.toString() || '',
        credit_account_id: detailResponse.credit_account_id?.toString() || '',
        payment_reference: detailResponse.payment_reference || '',
        items: detailResponse.purchase_items?.map(item => ({
          product_id: item.product_id.toString(),
          quantity: item.quantity.toString(),
          unit_price: item.unit_price.toString(),
          discount: item.discount?.toString() || '0',
          tax: item.tax?.toString() || '0',
          expense_account_id: item.expense_account_id?.toString() || '1'
        })) || [{
          product_id: '2',
          quantity: '1',
          unit_price: '1000',
          discount: '0',
          tax: '0',
          expense_account_id: '1'
        }]
      });
      
    await fetchVendors(); // Load vendors for dropdown
    await fetchProductsList(); // Load products for dropdown
    await fetchExpenseAccounts(); // Load expense accounts for dropdown
    await fetchBankAccounts(); // Load bank accounts for dropdown
    await fetchCreditAccounts(); // Load credit accounts (liability) for dropdown
    onEditOpen();
    } catch (err: any) {
      toast({
        title: 'Error',
        description: 'Failed to fetch purchase details for editing',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    }
  };

  // Handle create new purchase
const handleCreate = async () => {
    // Reset form data
    setFormData({
      vendor_id: '',
      date: new Date().toISOString().split('T')[0], // Today's date
      due_date: '',
      notes: '',
      discount: '0',
      
      // Legacy tax field
      tax: '0',
      
      // Tax additions (Penambahan)
      ppn_rate: '11',
      other_tax_additions: '0',
      
      // Tax deductions (Pemotongan)
      pph21_rate: '0',
      pph23_rate: '0',
      other_tax_deductions: '0',
      
      // Payment method fields
      payment_method: 'CREDIT',
      bank_account_id: '',
      credit_account_id: '',
      payment_reference: '',
      
      items: []
    });
    setSelectedPurchase(null);
    await fetchVendors(); // Load vendors for dropdown
    await fetchProductsList(); // Load products for dropdown
    await fetchExpenseAccounts(); // Load expense accounts for dropdown
    await fetchBankAccounts(); // Load bank accounts for dropdown
    await fetchCreditAccounts(); // Load credit accounts (liability) for dropdown
    onCreateOpen();
  };

  // Handle save for both create and edit
  const handleSave = async () => {
    if (!formData.vendor_id) {
      toast({
        title: 'Error',
        description: 'Please select a vendor',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    if (formData.items.length === 0) {
      toast({
        title: 'Error',
        description: 'Please add at least one item',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    // Validate all items have product, quantity, and expense account
    const invalidItems = formData.items.filter(item => 
      !item.product_id || !item.quantity || !item.expense_account_id
    );

    if (invalidItems.length > 0) {
      toast({
        title: 'Error',
        description: 'Please fill in all required fields for each item',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    try {
      setLoading(true);
      
      // Format the payload with proper tax rates
      const payload = {
        vendor_id: parseInt(formData.vendor_id),
        date: formData.date ? `${formData.date}T00:00:00Z` : new Date().toISOString(),
        due_date: formData.due_date ? `${formData.due_date}T00:00:00Z` : undefined,
        notes: formData.notes,
        discount: parseFloat(formData.discount) || 0,
        // Send proper tax rates (not legacy tax field)
        ppn_rate: parseFloat(formData.ppn_rate) || 11,
        other_tax_additions: parseFloat(formData.other_tax_additions) || 0,
        pph21_rate: parseFloat(formData.pph21_rate) || 0,
        pph23_rate: parseFloat(formData.pph23_rate) || 0,
        other_tax_deductions: parseFloat(formData.other_tax_deductions) || 0,
        // Payment method fields
        payment_method: formData.payment_method,
        bank_account_id: formData.bank_account_id ? parseInt(formData.bank_account_id) : undefined,
        credit_account_id: formData.credit_account_id ? parseInt(formData.credit_account_id) : undefined,
        payment_reference: formData.payment_reference,
        items: formData.items.map(item => ({
          product_id: parseInt(item.product_id),
          quantity: parseFloat(item.quantity),
          unit_price: parseFloat(item.unit_price),
          discount: parseFloat(item.discount) || 0,
          tax: parseFloat(item.tax) || 0,
          expense_account_id: parseInt(item.expense_account_id),
        })),
      };

      let response;
      
      if (selectedPurchase) {
        // Update existing purchase
        response = await purchaseService.update(selectedPurchase.id, payload);
        toast({
          title: 'Success',
          description: `Purchase ${response.code} updated successfully`,
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
        onEditClose();
      } else {
        // Create new purchase
        response = await purchaseService.create(payload);
        toast({
          title: 'Success',
          description: `Purchase ${response.code} created successfully. Use "Submit for Approval" button to submit when ready.`,
          status: 'success',
          duration: 5000,
          isClosable: true,
        });
        onCreateClose();
        
        // NOTE: Purchase is now created as DRAFT - Employee must manually submit for approval
        // This allows Employee to review the purchase details before submitting
      }
      
      // Refresh the list
      await fetchPurchases();
      
    } catch (err: any) {
      console.error('Error saving purchase:', err);
      const errorMessage = err.response?.data?.message || err.response?.data?.error || err.message || 'An error occurred';
      toast({
        title: 'Error',
        description: errorMessage,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };


  // Smart Action Button Logic
  const getActionButtonProps = (purchase: Purchase) => {
    const roleNorm = normalizeRole(user?.role as any);
    const status = (purchase.approval_status || '').toUpperCase();
    const purchaseStatus = (purchase.status || '').toUpperCase();
    
    // Helper function to get current active approval step
    const getCurrentActiveStep = () => {
      // Check if we have approval steps data from backend
      if (purchase.approval_request?.approval_steps) {
        // Find the active step that's pending - this is the most important check
        const activeStep = purchase.approval_request.approval_steps.find(
          step => step.is_active && step.status === 'PENDING'
        );
        
        if (activeStep) {
          return {
            step_name: activeStep.step.step_name,
            approver_role: normalizeRole(activeStep.step.approver_role),
            step_order: activeStep.step.step_order,
            is_escalated: activeStep.step.step_name?.includes('Escalated') || activeStep.step.step_name?.includes('Director') || false
          };
        }
        
        // If no active step found, check for any pending step (fallback)
        const pendingStep = purchase.approval_request.approval_steps.find(
          step => step.status === 'PENDING'
        );
        
        if (pendingStep) {
          return {
            step_name: pendingStep.step.step_name,
            approver_role: normalizeRole(pendingStep.step.approver_role),
            step_order: pendingStep.step.step_order,
            is_escalated: pendingStep.step.step_name?.includes('Escalated') || pendingStep.step.step_name?.includes('Director') || false
          };
        }
      }
      
      // Fallback logic if no approval steps data available
      if (purchaseStatus === 'DRAFT' && roleNorm === 'employee') {
        return { step_name: 'Submit', approver_role: 'employee', step_order: 0, is_escalated: false };
      }
      
      // Enhanced fallback logic based on status and amount
      if (status === 'PENDING' || status === 'NOT_STARTED' || purchaseStatus === 'PENDING_APPROVAL') {
        // Check if this purchase requires director approval based on amount or other criteria
        const requiresDirectorApproval = purchase.total_amount > 25000000;
        
        // Check if purchase has been escalated to director (look for director-related indicators)
        const isEscalatedToDirector = purchase.approval_request?.approval_steps?.some(
          step => normalizeRole(step.step.approver_role) === 'director' && step.status === 'PENDING'
        ) || purchase.approval_request?.current_step_name?.toLowerCase().includes('director');
        
        if (requiresDirectorApproval || isEscalatedToDirector) {
          return { step_name: 'Director Approval', approver_role: 'director', step_order: 2, is_escalated: true };
        }
        
        // Default to finance approval
        return { step_name: 'Finance Approval', approver_role: 'finance', step_order: 1, is_escalated: false };
      }
      
      return null;
    };
    
    const activeStep = getCurrentActiveStep();
    const isUserTurn = activeStep?.approver_role === roleNorm;

    // Completed states
    if (status === 'APPROVED' || status === 'REJECTED' || purchaseStatus === 'CANCELLED') {
      return { text: 'View', icon: <FiEye />, colorScheme: 'gray', variant: 'outline' };
    }

    // User's turn to act
    if (isUserTurn) {
      if (roleNorm === 'employee' && purchaseStatus === 'DRAFT') {
        return { text: 'Submit for Approval', icon: <FiAlertCircle />, colorScheme: 'blue', variant: 'solid' };
      }
      
      // Show appropriate text based on escalation
      const actionText = activeStep?.is_escalated ? 'Action Required (Escalated)' : 'Action Required';
      return { text: actionText, icon: <FiAlertCircle />, colorScheme: 'orange', variant: 'solid' };
    }

    // Waiting for others - show who needs to act
    if (status === 'PENDING' || purchaseStatus === 'PENDING_APPROVAL') {
      if (activeStep) {
        const waitingForRole = activeStep.approver_role === 'finance' ? 'Finance' : 
                              activeStep.approver_role === 'director' ? 'Director' :
                              activeStep.approver_role === 'admin' ? 'Admin' : 'Approval';
        
        const waitingText = activeStep.is_escalated ? `Waiting for ${waitingForRole} (Escalated)` : `Waiting for ${waitingForRole}`;
        return { text: waitingText, icon: <FiClock />, colorScheme: 'blue', variant: 'outline' };
      }
      return { text: 'Review Progress', icon: <FiClock />, colorScheme: 'blue', variant: 'outline' };
    }
    
    return { text: 'View', icon: <FiEye />, colorScheme: 'gray', variant: 'outline' };
  };

  // Action buttons for each row
  const renderActions = (purchase: Purchase) => {
    const actionProps = getActionButtonProps(purchase);
    const roleNorm = normalizeRole(user?.role as any);
    const purchaseStatus = (purchase.status || '').toUpperCase();
    
    return (
      <HStack spacing={2}>
        {/* Smart Single Action Button */}
        <Button
          size="sm"
          variant={actionProps.variant}
          colorScheme={actionProps.colorScheme}
          leftIcon={actionProps.icon}
          onClick={() => {
            // Handle special case for employee submitting draft purchase
            if (roleNorm === 'employee' && purchaseStatus === 'DRAFT' && actionProps.text === 'Submit for Approval') {
              handleSubmitForApproval(purchase.id);
            } else {
              setSelectedPurchase(purchase);
              onViewOpen();
            }
          }}
          fontWeight={actionProps.variant === 'solid' ? 'semibold' : 'medium'}
          _hover={{
            transform: 'translateY(-1px)',
            boxShadow: 'md'
          }}
        >
          {actionProps.text}
        </Button>
        
        {/* Delete button for ADMIN - can delete any status */}
        {normalizeRole(user?.role as any) === 'admin' && (
          <Button
            size="sm"
            colorScheme="red"
            variant="outline"
            leftIcon={<FiTrash2 />}
            onClick={() => handleDelete(purchase.id)}
          >
            Delete
          </Button>
        )}
        
        {/* Delete button removed for DIRECTOR role per requirement */}
      </HStack>
    );
  };

  if (loading) {
    return (
<SimpleLayout allowedRoles={['admin', 'finance', 'inventory_manager', 'employee', 'director']}>
        <Box>
          <Text>Loading purchases...</Text>
        </Box>
      </SimpleLayout>
    );
  }

  return (
    <SimpleLayout allowedRoles={['admin', 'finance', 'inventory_manager', 'employee', 'director']}>
      <Box 
        bg={bgColor}
        minH="100vh"
        p={6}
      >
        <VStack spacing={6} align="stretch">
        {/* Enhanced Header */}
        <Card 
          bg={cardBg}
          borderWidth="1px"
          borderColor={borderColor}
          boxShadow="sm"
          borderRadius="lg"
          mb={2}
        >
          <CardBody p={6}>
            <Flex justify="space-between" align="center">
              <Box>
                <Heading 
                  size="lg" 
                  color={headingColor}
                  fontWeight="bold"
                  mb={1}
                >
                  Purchase Management
                </Heading>
                <Text 
                  fontSize="sm" 
                  color={textSecondary}
                >
                  Manage purchase orders and approvals
                </Text>
              </Box>
              <HStack spacing={3}>
                <Button
                  variant="outline"
                  size="md"
                  leftIcon={<FiFilter />}
                  onClick={onFilterOpen}
                  borderColor={borderColor}
                  color={textPrimary}
                  _hover={{
                    bg: hoverBg,
                    borderColor: hoverBorder,
                    transform: 'translateY(-1px)'
                  }}
                  transition="all 0.2s ease"
                >
                  Filters
                </Button>
                <Button
                  variant="outline"
                  size="md"
                  leftIcon={<FiRefreshCw />}
                  onClick={handleRefresh}
                  isLoading={loading}
                  borderColor={borderColor}
                  color={textPrimary}
                  _hover={{
                    bg: hoverBg,
                    borderColor: hoverBorder,
                    transform: 'translateY(-1px)'
                  }}
                  transition="all 0.2s ease"
                >
                  Refresh
                </Button>
                {/* New Purchase button only for Employee role */}
                {normalizeRole(user?.role as any) === 'employee' && (
                  <Button
                    colorScheme="blue"
                    size="md"
                    leftIcon={<FiPlus />}
                    onClick={handleCreate}
                    bg={buttonBlueBg}
                    color="white"
                    _hover={{
                      bg: buttonBlueHover,
                      transform: 'translateY(-1px)',
                      boxShadow: 'md'
                    }}
                    _active={{
                      transform: 'translateY(0)'
                    }}
                    transition="all 0.2s ease"
                    fontWeight="semibold"
                  >
                    New Purchase
                  </Button>
                )}
              </HStack>
            </Flex>
          </CardBody>
        </Card>

        {/* Enhanced Statistics Cards */}
        <Grid templateColumns="repeat(auto-fit, minmax(280px, 1fr))" gap={6}>
          <Card 
            bg={cardBg}
            borderWidth="1px"
            borderColor={borderColor}
            boxShadow="sm"
            borderRadius="lg"
            _hover={{ 
              boxShadow: 'md',
              transform: 'translateY(-2px)',
              transition: 'all 0.2s ease'
            }}
            transition="all 0.2s ease"
          >
            <CardBody p={6}>
              <Flex align="center" justify="space-between">
                <Stat>
                  <StatLabel 
                    color={textSecondary}
                    fontSize="sm"
                    fontWeight="medium"
                    mb={2}
                  >
                    Total Purchases
                  </StatLabel>
                  <StatNumber 
                    color={headingColor}
                    fontSize="2xl"
                    fontWeight="bold"
                  >
                    {stats.total}
                  </StatNumber>
                </Stat>
                <Box 
                  p={3} 
                  borderRadius="lg"
                  bg={statBgColors.blue}
                  color={statColors.blue}
                >
                  <FiRefreshCw size={20} />
                </Box>
              </Flex>
            </CardBody>
          </Card>
          
          <Card 
            bg={cardBg}
            borderWidth="1px"
            borderColor={borderColor}
            boxShadow="sm"
            borderRadius="lg"
            _hover={{ 
              boxShadow: 'md',
              transform: 'translateY(-2px)',
              transition: 'all 0.2s ease'
            }}
            transition="all 0.2s ease"
          >
            <CardBody p={6}>
              <Flex align="center" justify="space-between">
                <Stat>
                  <StatLabel 
                    color={textSecondary}
                    fontSize="sm"
                    fontWeight="medium"
                    mb={2}
                  >
                    Pending Approval
                  </StatLabel>
                  <StatNumber 
                    color={statColors.orange}
                    fontSize="2xl"
                    fontWeight="bold"
                  >
                    {stats.needingApproval}
                  </StatNumber>
                </Stat>
                <Box 
                  p={3} 
                  borderRadius="lg"
                  bg={statBgColors.orange}
                  color={statColors.orange}
                >
                  <FiClock size={20} />
                </Box>
              </Flex>
            </CardBody>
          </Card>
          
          <Card 
            bg={cardBg}
            borderWidth="1px"
            borderColor={borderColor}
            boxShadow="sm"
            borderRadius="lg"
            _hover={{ 
              boxShadow: 'md',
              transform: 'translateY(-2px)',
              transition: 'all 0.2s ease'
            }}
            transition="all 0.2s ease"
          >
            <CardBody p={6}>
              <Flex align="center" justify="space-between">
                <Stat>
                  <StatLabel 
                    color={textSecondary}
                    fontSize="sm"
                    fontWeight="medium"
                    mb={2}
                  >
                    Approved
                  </StatLabel>
                  <StatNumber 
                    color={statColors.green}
                    fontSize="2xl"
                    fontWeight="bold"
                  >
                    {stats.approved}
                  </StatNumber>
                </Stat>
                <Box 
                  p={3} 
                  borderRadius="lg"
                  bg={statBgColors.green}
                  color={statColors.green}
                >
                  <FiCheckCircle size={20} />
                </Box>
              </Flex>
            </CardBody>
          </Card>
          
          <Card 
            bg={cardBg}
            borderWidth="1px"
            borderColor={borderColor}
            boxShadow="sm"
            borderRadius="lg"
            _hover={{ 
              boxShadow: 'md',
              transform: 'translateY(-2px)',
              transition: 'all 0.2s ease'
            }}
            transition="all 0.2s ease"
          >
            <CardBody p={6}>
              <Flex align="center" justify="space-between">
                <Stat>
                  <StatLabel 
                    color={textSecondary}
                    fontSize="sm"
                    fontWeight="medium"
                    mb={2}
                  >
                    Rejected
                  </StatLabel>
                  <StatNumber 
                    color={statColors.red}
                    fontSize="2xl"
                    fontWeight="bold"
                  >
                    {stats.rejected}
                  </StatNumber>
                </Stat>
                <Box 
                  p={3} 
                  borderRadius="lg"
                  bg={statBgColors.red}
                  color={statColors.red}
                >
                  <FiXCircle size={20} />
                </Box>
              </Flex>
            </CardBody>
          </Card>
          
          <Card 
            bg={cardBg}
            borderWidth="1px"
            borderColor={borderColor}
            boxShadow="sm"
            borderRadius="lg"
            _hover={{ 
              boxShadow: 'md',
              transform: 'translateY(-2px)',
              transition: 'all 0.2s ease'
            }}
            transition="all 0.2s ease"
          >
            <CardBody p={6}>
              <Flex align="center" justify="space-between">
                <Stat>
                  <StatLabel 
                    color={textSecondary}
                    fontSize="sm"
                    fontWeight="medium"
                    mb={2}
                  >
                    Total Value
                  </StatLabel>
                  <StatNumber 
                    color={headingColor}
                    fontSize="lg"
                    fontWeight="bold"
                  >
                    {formatCurrency(stats.totalValue || 0)}
                  </StatNumber>
                </Stat>
                <Box 
                  p={3} 
                  borderRadius="lg"
                  bg={statBgColors.purple}
                  color={statColors.purple}
                >
                  <FiAlertCircle size={20} />
                </Box>
              </Flex>
            </CardBody>
          </Card>
          
          <Card 
            bg={cardBg}
            borderWidth="1px"
            borderColor={borderColor}
            boxShadow="sm"
            borderRadius="lg"
            _hover={{ 
              boxShadow: 'md',
              transform: 'translateY(-2px)',
              transition: 'all 0.2s ease'
            }}
            transition="all 0.2s ease"
          >
            <CardBody p={6}>
              <Flex align="center" justify="space-between">
                <Stat>
                  <StatLabel 
                    color={textSecondary}
                    fontSize="sm"
                    fontWeight="medium"
                    mb={2}
                  >
                    Total Approved Amount
                  </StatLabel>
                  <StatNumber 
                    color={statColors.green}
                    fontSize="lg"
                    fontWeight="bold"
                  >
                    {formatCurrency(stats.totalApprovedAmount || 0)}
                  </StatNumber>
                </Stat>
                <Box 
                  p={3} 
                  borderRadius="lg"
                  bg={statBgColors.green}
                  color={statColors.green}
                >
                  <FiCheckCircle size={20} />
                </Box>
              </Flex>
            </CardBody>
          </Card>
        </Grid>

        {error && (
          <Alert status="error">
            <AlertIcon />
            <AlertTitle>Error!</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {/* Main Data Table */}
        <EnhancedPurchaseTable
          purchases={purchases}
          loading={loading}
          onViewDetails={handleView}
          onEdit={canEdit ? handleEdit : undefined}
          onSubmitForApproval={handleSubmitForApproval}
          onDelete={canDelete ? handleDelete : undefined}
          renderActions={renderActions}
          title="Purchase Transactions"
          formatCurrency={formatCurrency}
          formatDate={formatDate}
          canEdit={canEdit}
          canDelete={canDelete}
          userRole={normalizeRole(user?.role as any)}
        />

        {/* Enhanced Filter Modal */}
        <Modal isOpen={isFilterOpen} onClose={onFilterClose} size="md">
          <ModalOverlay bg="blackAlpha.300" backdropFilter="blur(10px)" />
          <ModalContent 
            bg={cardBg}
            borderWidth="1px"
            borderColor={borderColor}
            borderRadius="lg"
            boxShadow="xl"
          >
            <ModalHeader 
              color={headingColor}
              fontSize="lg"
              fontWeight="bold"
              borderBottom="1px solid"
              borderColor={borderColor}
              pb={4}
            >
              <HStack>
                <Box 
                  p={2} 
                  borderRadius="md"
                  bg={modalFilterBg}
                  color={modalFilterColor}
                >
                  <FiFilter size={16} />
                </Box>
                <Text>Filter Purchases</Text>
              </HStack>
            </ModalHeader>
            <ModalCloseButton 
              color={textSecondary}
              _hover={{
                bg: modalHoverBg
              }}
            />
            <ModalBody py={6}>
              <VStack spacing={5}>
                <FormControl>
                  <FormLabel 
                    fontSize="sm"
                    fontWeight="semibold"
                    color={textPrimary}
                    mb={2}
                  >
                    Search
                  </FormLabel>
                  <Input
                    placeholder="Search by purchase number, vendor..."
                    value={filters.search || ''}
                    onChange={(e) => handleFilterChange({ search: e.target.value })}
                    bg={modalBg}
                    borderColor={borderColor}
                    _hover={{
                      borderColor: inputHoverBorder
                    }}
                    _focus={{
                      borderColor: inputFocusBorder,
                      boxShadow: inputFocusShadow
                    }}
                  />
                </FormControl>
                
                <FormControl>
                  <FormLabel 
                    fontSize="sm"
                    fontWeight="semibold"
                    color={textPrimary}
                    mb={2}
                  >
                    Status
                  </FormLabel>
                  <Select
                    placeholder="All Statuses"
                    value={filters.status || ''}
                    onChange={(e) => handleFilterChange({ status: e.target.value })}
                    bg={modalBg}
                    borderColor={borderColor}
                    _hover={{
                      borderColor: inputHoverBorder
                    }}
                    _focus={{
                      borderColor: inputFocusBorder,
                      boxShadow: inputFocusShadow
                    }}
                  >
                    <option value="draft">Draft</option>
                    <option value="pending_approval">Pending Approval</option>
                    <option value="approved">Approved</option>
                    <option value="cancelled">Cancelled</option>
                  </Select>
                </FormControl>
                
                <FormControl>
                  <FormLabel 
                    fontSize="sm"
                    fontWeight="semibold"
                    color={textPrimary}
                    mb={2}
                  >
                    Approval Status
                  </FormLabel>
                  <Select
                    placeholder="All Approval Statuses"
                    value={filters.approval_status || ''}
                    onChange={(e) => handleFilterChange({ approval_status: e.target.value })}
                    bg={modalBg}
                    borderColor={borderColor}
                    _hover={{
                      borderColor: inputHoverBorder
                    }}
                    _focus={{
                      borderColor: inputFocusBorder,
                      boxShadow: inputFocusShadow
                    }}
                  >
                    <option value="not_required">Not Required</option>
                    <option value="pending">Pending</option>
                    <option value="approved">Approved</option>
                    <option value="rejected">Rejected</option>
                  </Select>
                </FormControl>
              </VStack>
            </ModalBody>
            <ModalFooter 
              borderTop="1px solid"
              borderColor={borderColor}
              pt={4}
            >
              <HStack spacing={3} w="100%">
                <Button 
                  variant="ghost" 
                  onClick={onFilterClose}
                  flex={1}
                  color={textSecondary}
                  _hover={{
                    bg: ghostHoverBg
                  }}
                >
                  Close
                </Button>
                <Button 
                  colorScheme="blue"
                  onClick={() => {
                    setFilters({ page: 1, limit: 10 });
                    fetchPurchases({ page: 1, limit: 10 });
                    onFilterClose();
                  }}
                  flex={1}
                  bg={buttonBlueBg}
                  _hover={{
                    bg: buttonBlueHover
                  }}
                  fontWeight="semibold"
                >
                  Clear Filters
                </Button>
              </HStack>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* View Purchase Modal */}
        <Modal isOpen={isViewOpen} onClose={onViewClose} size="xl">
          <ModalOverlay />
          <ModalContent bg={modalContentBg}>
            <ModalHeader bg={modalHeaderBg} borderBottomWidth={1} borderColor={borderColor}>
              View Purchase - {selectedPurchase?.code}
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              {selectedPurchase && (
                <VStack spacing={6} align="stretch">
                  {/* Show rejection alert for cancelled/rejected purchases */}
                  {(selectedPurchase.status === 'CANCELLED' || selectedPurchase.approval_status === 'REJECTED') && (
                    <Alert status="error" variant="left-accent">
                      <AlertIcon />
                      <VStack align="start" spacing={1}>
                        <AlertTitle>
                          {selectedPurchase.status === 'CANCELLED' ? 'Purchase Dibatalkan' : 'Purchase Ditolak'}
                        </AlertTitle>
                        <AlertDescription>
                          {selectedPurchase.status === 'CANCELLED' 
                            ? 'Purchase ini telah dibatalkan dan tidak dapat diproses lebih lanjut.'
                            : 'Purchase ini ditolak pada proses approval. Lihat detail penolakan di bagian Approval History.'}
                        </AlertDescription>
                      </VStack>
                    </Alert>
                  )}
                  
                  {/* Basic Info */}
                  <SimpleGrid columns={2} spacing={4}>
                    <FormControl>
                      <FormLabel>Purchase Code</FormLabel>
                      <Text fontWeight="medium">{selectedPurchase.code}</Text>
                    </FormControl>
                    
                    <FormControl>
                      <FormLabel>Vendor</FormLabel>
                      <Text fontWeight="medium">{selectedPurchase.vendor?.name || 'N/A'}</Text>
                    </FormControl>
                    
                    <FormControl>
                      <FormLabel>Date</FormLabel>
                      <Text fontWeight="medium">{new Date(selectedPurchase.date).toLocaleDateString('id-ID')}</Text>
                    </FormControl>
                    
                    <FormControl>
                      <FormLabel>Total Amount</FormLabel>
                      <Text fontWeight="medium" color="green.500">{formatCurrency(selectedPurchase.total_amount)}</Text>
                    </FormControl>
                    
                    <FormControl>
                      <FormLabel>Status</FormLabel>
                      <Badge colorScheme={getStatusColor(selectedPurchase.status)} variant="subtle" w="fit-content">
                        {selectedPurchase.status.replace('_', ' ').toUpperCase()}
                      </Badge>
                    </FormControl>
                    
                    <FormControl>
                      <FormLabel>Approval Status</FormLabel>
                      <Badge colorScheme={getApprovalStatusColor(selectedPurchase.approval_status)} variant="subtle" w="fit-content">
                        {selectedPurchase.approval_status.replace('_', ' ').toUpperCase()}
                      </Badge>
                    </FormControl>
                  </SimpleGrid>
                  
                  {/* Payment Information */}
                  {selectedPurchase.payment_method && (
                    <Box>
                      <FormLabel mb={3}>Payment Information</FormLabel>
                      <SimpleGrid columns={3} spacing={4}>
                        <FormControl>
                          <FormLabel fontSize="sm">Payment Method</FormLabel>
                          <Badge 
                            colorScheme={
                              selectedPurchase.payment_method === 'CREDIT' ? 'orange' :
                              selectedPurchase.payment_method === 'CASH' ? 'green' :
                              selectedPurchase.payment_method === 'BANK_TRANSFER' ? 'blue' :
                              selectedPurchase.payment_method === 'CHECK' ? 'purple' : 'gray'
                            } 
                            variant="subtle" 
                            w="fit-content"
                          >
                            {selectedPurchase.payment_method.replace('_', ' ')}
                          </Badge>
                        </FormControl>
                        
                        {selectedPurchase.bank_account_id && (
                          <FormControl>
                            <FormLabel fontSize="sm">Bank Account</FormLabel>
                            <Text fontWeight="medium">
                              {selectedPurchase.bank_account?.name || 'Unknown Account'}
                              {selectedPurchase.bank_account?.code && ` (${selectedPurchase.bank_account.code})`}
                            </Text>
                          </FormControl>
                        )}
                        
                        {selectedPurchase.payment_reference && (
                          <FormControl>
                            <FormLabel fontSize="sm">Payment Reference</FormLabel>
                            <Text fontWeight="medium">{selectedPurchase.payment_reference}</Text>
                          </FormControl>
                        )}
                      </SimpleGrid>
                    </Box>
                  )}
                  
                  {/* Notes */}
                  {selectedPurchase.notes && (
                    <FormControl>
                      <FormLabel>Notes</FormLabel>
                      <Text p={3} bg={notesBoxBg} borderRadius="md">{selectedPurchase.notes}</Text>
                    </FormControl>
                  )}
                  
                  {/* Approval Panel */}
                  <ApprovalPanel 
                    purchaseId={selectedPurchase.id}
                    approvalStatus={selectedPurchase.approval_status}
                    purchaseAmount={selectedPurchase.total_amount}
                    canApprove={(() => {
                      const roleNorm = normalizeRole(user?.role as any);
                      const isDraft = (selectedPurchase.status || '').toUpperCase() === 'DRAFT';
                      const isPending = (selectedPurchase.approval_status || '').toUpperCase() === 'PENDING';
                      const isNotStarted = (selectedPurchase.approval_status || '').toUpperCase() === 'NOT_STARTED';
                      
                      // Admin can always approve
                      if (roleNorm === 'admin') return true;
                      
                      // Finance can approve DRAFT purchases, pending purchases (escalated), or purchases that haven't started approval
                      if (roleNorm === 'finance' && (isDraft || isPending || isNotStarted)) return true;
                      
                      // Director can approve pending purchases (escalated)
                      if (roleNorm === 'director' && isPending) return true;
                      
                      // Check approval steps for other roles
                      const steps: any[] = (selectedPurchase as any)?.approval_steps || [];
                      if (!Array.isArray(steps) || steps.length === 0) return false;
                      const active = steps.find((s: any) => s.is_active && s.status === 'PENDING');
                      const approverRole = active?.step?.approver_role ? normalizeRole(active.step.approver_role) : null;
                      return !!approverRole && approverRole === roleNorm;
                    })()}
                    onApprove={async (comments?: string, requiresDirector?: boolean) => {
                      if (!selectedPurchase) return;
                      try {
                        // Call API to approve with escalation parameter
                        const result = await approvalService.approvePurchase(selectedPurchase.id, { 
                          comments: comments || '',
                          escalate_to_director: requiresDirector || false
                        });
                        
                        // Handle different approval outcomes
                        if (result.escalated) {
                          toast({ 
                            title: 'Approved & Escalated', 
                            description: result.message || 'Purchase approved by Finance and escalated to Director for final approval', 
                            status: 'info', 
                            duration: 5000, 
                            isClosable: true 
                          });
                          
                          // Send notification to directors
                          await notifyDirectors(selectedPurchase);
                        } else {
                          toast({ 
                            title: 'Approved', 
                            description: result.message || 'Purchase approved successfully', 
                            status: 'success', 
                            duration: 3000, 
                            isClosable: true 
                          });
                        }
                        
                        // Refresh purchase data
                        const detailResponse = await purchaseService.getById(selectedPurchase.id);
                        setSelectedPurchase(detailResponse);
                        await fetchPurchases();
                        // Don't close modal - let user see the updated history with comments
                      } catch (err: any) {
                        toast({ 
                          title: 'Error', 
                          description: err.response?.data?.message || err.response?.data?.error || 'Failed to approve', 
                          status: 'error', 
                          duration: 5000, 
                          isClosable: true 
                        });
                      }
                    }}
                    onReject={async (comments: string) => {
                      if (!selectedPurchase) return;
                      if (!comments || comments.trim() === '') {
                        toast({ title: 'Komentar diperlukan', description: 'Mohon isi alasan penolakan.', status: 'warning', duration: 3000, isClosable: true });
                        return;
                      }
                      try {
                        await approvalService.rejectPurchase(selectedPurchase.id, { comments });
                        toast({ title: 'Rejected', description: 'Purchase rejected successfully', status: 'warning', duration: 3000, isClosable: true });
                        const detailResponse = await purchaseService.getById(selectedPurchase.id);
                        setSelectedPurchase(detailResponse);
                        await fetchPurchases();
                        // Don't close modal - let user see the updated history with rejection comments
                      } catch (err: any) {
                        toast({ title: 'Error', description: err.response?.data?.message || 'Failed to reject', status: 'error', duration: 5000, isClosable: true });
                      }
                    }}
                  />

                  {/* Items */}
                  {selectedPurchase.purchase_items && selectedPurchase.purchase_items.length > 0 && (
                    <FormControl>
                      <FormLabel>Purchase Items</FormLabel>
                      <TableContainer>
                        <Table size="sm" bg={tableBg}>
                          <Thead bg={tableHeaderBg}>
                            <Tr>
                              <Th>Product</Th>
                              <Th isNumeric>Quantity</Th>
                              <Th isNumeric>Unit Price</Th>
                              <Th isNumeric>Total</Th>
                            </Tr>
                          </Thead>
                          <Tbody>
                            {selectedPurchase.purchase_items.map((item: any, index: number) => (
                              <Tr key={index}>
                                <Td>{item.product?.name || 'N/A'}</Td>
                                <Td isNumeric>{item.quantity}</Td>
                                <Td isNumeric>{formatCurrency(item.unit_price)}</Td>
                                <Td isNumeric>{formatCurrency(item.quantity * item.unit_price)}</Td>
                              </Tr>
                            ))}
                          </Tbody>
                        </Table>
                      </TableContainer>
                    </FormControl>
                  )}
                </VStack>
              )}
            </ModalBody>
            <ModalFooter>
              <Button onClick={onViewClose}>Close</Button>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* Edit Purchase Modal */}
        <Modal isOpen={isEditOpen} onClose={onEditClose} size="2xl">
          <ModalOverlay />
          <ModalContent bg={modalContentBg}>
            <ModalHeader bg={modalHeaderBg} borderBottomWidth={1} borderColor={borderColor}>
              Edit Purchase - {selectedPurchase?.code}
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <VStack spacing={4} align="stretch">
                <Text fontSize="md" fontWeight="semibold" color={headingColor}>Basic Info</Text>
                <SimpleGrid columns={2} spacing={4}>
                      <FormControl isRequired>
                        <FormLabel>Vendor</FormLabel>
                        <HStack spacing={2}>
                          {loadingVendors ? (
                            <Spinner size="sm" />
                          ) : (
                            <Select
                              placeholder="Select vendor"
                              value={formData.vendor_id}
                              onChange={(e) => setFormData({...formData, vendor_id: e.target.value})}
                              flex={1}
                            >
                              {vendors.map(vendor => (
                                <option key={vendor.id} value={vendor.id}>
                                  {vendor.name} ({vendor.code})
                                </option>
                              ))}
                            </Select>
                          )}
                          <IconButton
                            aria-label="Add new vendor"
                            icon={<FiPlus />}
                            size="sm"
                            colorScheme="green"
                            variant="outline"
                            onClick={onAddVendorOpen}
                            title="Add New Vendor"
                            _hover={{ bg: 'green.50' }}
                          />
                        </HStack>
                      </FormControl>
                  
                  <FormControl isRequired>
                    <FormLabel>Purchase Date</FormLabel>
                    <Input
                      type="date"
                      value={formData.date}
                      onChange={(e) => setFormData({...formData, date: e.target.value})}
                    />
                  </FormControl>
                </SimpleGrid>

                <SimpleGrid columns={2} spacing={4}>
                  <FormControl>
                    <FormLabel>Due Date</FormLabel>
                    <Input
                      type="date"
                      value={formData.due_date}
                      onChange={(e) => setFormData({...formData, due_date: e.target.value})}
                    />
                  </FormControl>

                  <FormControl>
                    <FormLabel>Discount (%)</FormLabel>
                    <NumberInput
                      value={formData.discount}
                      onChange={(value) => setFormData({...formData, discount: value})}
                    >
                      <NumberInputField placeholder="0" />
                    </NumberInput>
                    <FormHelperText>Masukkan persentase diskon atas subtotal (0-100).</FormHelperText>
                  </FormControl>
                </SimpleGrid>

                {!canListExpenseAccounts && (
                  <FormControl>
                    <FormLabel>Default Expense Account ID</FormLabel>
                    <NumberInput min={1} value={defaultExpenseAccountId ?? ''} onChange={(v) => setDefaultExpenseAccountId(isNaN(Number(v)) ? null : Number(v))} maxW="260px">
                      <NumberInputField placeholder="Masukkan Account ID (EXPENSE)" />
                    </NumberInput>
                    <FormHelperText>Karena role Anda tidak bisa melihat daftar akun, isi ID akun beban (EXPENSE) default di sini.</FormHelperText>
                  </FormControl>
                )}
                
                <FormControl>
                  <FormLabel>Notes</FormLabel>
                  <Textarea
                    value={formData.notes}
                    onChange={(e) => setFormData({...formData, notes: e.target.value})}
                    placeholder="Enter any notes or descriptions..."
                    rows={4}
                  />
                </FormControl>

                {/* Purchase Items Section */}
                <Card>
                  <CardHeader pb={3}>
                    <Flex justify="space-between" align="center">
                      <Text fontSize="md" fontWeight="semibold" color={textPrimary}>
                        🛒 Purchase Items
                      </Text>
                      <Button 
                        size="sm" 
                        leftIcon={<FiPlus />} 
                        colorScheme="blue"
                        variant="outline"
                        onClick={() => {
                          setFormData({
                            ...formData,
                            items: [
                              ...formData.items,
                              { product_id: '', quantity: '1', unit_price: '0', discount: '0', tax: '0', expense_account_id: '' }
                            ]
                          });
                        }}
                      >
                        Add Item
                      </Button>
                    </Flex>
                  </CardHeader>
                  <CardBody pt={0}>
                    <Box overflow="visible">
                      <Table size="sm" variant="simple">
                        <Thead bg={tableHeaderBg}>
                          <Tr>
                            <Th fontSize="xs" fontWeight="semibold" color={textSecondary}>Product</Th>
                            <Th fontSize="xs" fontWeight="semibold" color={textSecondary} isNumeric>Qty</Th>
                            <Th fontSize="xs" fontWeight="semibold" color={textSecondary} isNumeric>Unit Price (IDR)</Th>
                            <Th fontSize="xs" fontWeight="semibold" color={textSecondary}>Expense Account</Th>
                            <Th fontSize="xs" fontWeight="semibold" color={textSecondary} isNumeric>Total (IDR)</Th>
                            <Th fontSize="xs" fontWeight="semibold" color={textSecondary} w="60px">Action</Th>
                          </Tr>
                        </Thead>
                        <Tbody>
                          {formData.items.length === 0 ? (
                            <Tr>
                              <Td colSpan={6} textAlign="center" py={8}>
                                <VStack spacing={2}>
                                  <Text fontSize="sm" color={textSecondary}>No items added yet</Text>
                                  <Text fontSize="xs" color={textSecondary}>Click "Add Item" button to start adding purchase items</Text>
                                </VStack>
                              </Td>
                            </Tr>
                          ) : (
                            formData.items.map((item, index) => (
                              <Tr key={index} _hover={{ bg: hoverBg }}>
                                <Td minW="200px">
                                  {loadingProducts ? (
                                    <Flex align="center" justify="center" h="32px">
                                      <Spinner size="sm" />
                                    </Flex>
                                  ) : (
                                    <HStack spacing={2}>
                                      <Select
                                        placeholder="Select product"
                                        value={item.product_id}
                                        onChange={(e) => {
                                          const items = [...formData.items];
                                          items[index] = { ...items[index], product_id: e.target.value };
                                          setFormData({ ...formData, items });
                                        }}
                                        size="sm"
                                        maxW="280px"
                                      >
                                        {products.map((p) => (
                                          <option key={p.id} value={p.id?.toString()}>
                                            {p?.id} - {p?.name || p?.code}
                                          </option>
                                        ))}
                                      </Select>
                                      <IconButton 
                                        aria-label="Add new product"
                                        icon={<FiPlus />}
                                        size="sm"
                                        colorScheme="blue"
                                        variant="outline"
                                        onClick={onAddProductOpen}
                                        title="Add New Product"
                                        _hover={{ bg: 'blue.50' }}
                                      />
                                    </HStack>
                                  )}
                                </Td>
                                <Td isNumeric>
                                  <NumberInput 
                                    size="sm" 
                                    min={1} 
                                    value={item.quantity} 
                                    onChange={(valueString) => {
                                      const items = [...formData.items];
                                      items[index] = { ...items[index], quantity: valueString };
                                      setFormData({ ...formData, items });
                                    }} 
                                    maxW="80px"
                                  >
                                    <NumberInputField textAlign="right" fontSize="sm" />
                                    <NumberInputStepper>
                                      <NumberIncrementStepper />
                                      <NumberDecrementStepper />
                                    </NumberInputStepper>
                                  </NumberInput>
                                </Td>
                                <Td isNumeric>
                                  <Box maxW="160px">
                                    <CurrencyInput
                                      value={parseFloat(item.unit_price) || 0}
                                      onChange={(value) => {
                                        const items = [...formData.items];
                                        items[index] = { ...items[index], unit_price: value.toString() };
                                        setFormData({ ...formData, items });
                                      }}
                                      placeholder="Rp 10.000"
                                      size="sm"
                                      min={0}
                                      showLabel={false}
                                    />
                                  </Box>
                                </Td>
                                <Td minW="240px">
                                  {canListExpenseAccounts ? (
                                    <Box maxW="240px">
                                      <SearchableSelect
                                        options={expenseAccounts.map(acc => ({
                                          id: acc.id!,
                                          code: acc.code,
                                          name: acc.name,
                                          active: acc.is_active
                                        }))}
                                        value={item.expense_account_id}
                                        onChange={(value) => {
                                          const items = [...formData.items];
                                          items[index] = { ...items[index], expense_account_id: value.toString() };
                                          setFormData({ ...formData, items });
                                        }}
                                        placeholder="Pilih akun beban..."
                                        isLoading={loadingExpenseAccounts}
                                        displayFormat={(option) => `${option.code} - ${option.name}`}
                                        size="sm"
                                      />
                                    </Box>
                                  ) : (
                                    <NumberInput 
                                      min={1} 
                                      value={item.expense_account_id || (defaultExpenseAccountId ? defaultExpenseAccountId.toString() : '')} 
                                      onChange={(v) => {
                                        const items = [...formData.items];
                                        items[index] = { ...items[index], expense_account_id: v.toString() };
                                        setFormData({ ...formData, items });
                                      }} 
                                      maxW="240px"
                                      size="sm"
                                    >
                                      <NumberInputField placeholder="Expense Account ID" fontSize="sm" />
                                    </NumberInput>
                                  )}
                                </Td>
                                <Td isNumeric>
                                  <Text fontSize="sm" fontWeight="medium" color="green.600">
                                    {(() => {
                                      const qty = parseFloat(item.quantity || '0');
                                      const price = parseFloat(item.unit_price || '0');
                                      return formatCurrency((isNaN(qty) ? 0 : qty) * (isNaN(price) ? 0 : price));
                                    })()}
                                  </Text>
                                </Td>
                                <Td>
                                  <IconButton
                                    aria-label="Remove item"
                                    size="sm"
                                    colorScheme="red"
                                    variant="ghost"
                                    icon={<FiTrash2 />}
                                    onClick={() => {
                                      const items = [...formData.items];
                                      items.splice(index, 1);
                                      setFormData({ ...formData, items });
                                    }}
                                    _hover={{ bg: 'red.50' }}
                                  />
                                </Td>
                              </Tr>
                            ))
                          )}
                        </Tbody>
                      </Table>
                    </Box>
                    
                    {/* Summary Row */}
                    {formData.items.length > 0 && (
                      <Box mt={4} p={4} bg={statBgColors.blue} borderRadius="md" borderLeft="4px solid" borderLeftColor={statColors.blue}>
                        <Flex justify="space-between" align="center">
                          <Text fontSize="sm" fontWeight="medium" color={textPrimary}>
                            Total Items: {formData.items.length}
                          </Text>
                          <Text fontSize="lg" fontWeight="bold" color={statColors.blue}>
                            Subtotal: {formatCurrency(
                              formData.items.reduce((total, item) => {
                                const qty = parseFloat(item.quantity || '0');
                                const price = parseFloat(item.unit_price || '0');
                                return total + ((isNaN(qty) ? 0 : qty) * (isNaN(price) ? 0 : price));
                              }, 0)
                            )}
                          </Text>
                        </Flex>
                      </Box>
                    )}
                    
                    <FormHelperText mt={3} fontSize="xs">
                      📌 Tambahkan minimal 1 item pembelian. Semua field harus diisi dengan benar.
                    </FormHelperText>
                  </CardBody>
                </Card>

                {/* Tax Configuration Section */}
                <Card>
                  <CardHeader pb={3}>
                    <Text fontSize="md" fontWeight="semibold" color={textPrimary}>
                      💰 Tax Configuration
                    </Text>
                  </CardHeader>
                  <CardBody pt={0}>
                    <VStack spacing={4} align="stretch">
                      {/* Tax Additions (Penambahan) */}
                      <Box>
                        <Text fontSize="sm" fontWeight="medium" color={statColors.green} mb={3}>
                          ➕ Tax Additions (Penambahan)
                        </Text>
                        <SimpleGrid columns={2} spacing={4}>
                          <FormControl>
                            <FormLabel fontSize="sm">PPN Rate (%)</FormLabel>
                            <NumberInput
                              value={formData.ppn_rate}
                              onChange={(value) => setFormData({...formData, ppn_rate: value})}
                              size="sm"
                              min={0}
                              max={100}
                              step={0.1}
                            >
                              <NumberInputField placeholder="11" />
                              <NumberInputStepper>
                                <NumberIncrementStepper />
                                <NumberDecrementStepper />
                              </NumberInputStepper>
                            </NumberInput>
                            <FormHelperText fontSize="xs">Pajak Pertambahan Nilai (default 11%)</FormHelperText>
                          </FormControl>

                          <FormControl>
                            <FormLabel fontSize="sm">Other Tax Additions (%)</FormLabel>
                            <NumberInput
                              value={formData.other_tax_additions}
                              onChange={(value) => setFormData({...formData, other_tax_additions: value})}
                              size="sm"
                              min={0}
                              max={100}
                              step={0.1}
                            >
                              <NumberInputField placeholder="0" />
                              <NumberInputStepper>
                                <NumberIncrementStepper />
                                <NumberDecrementStepper />
                              </NumberInputStepper>
                            </NumberInput>
                            <FormHelperText fontSize="xs">Pajak tambahan lainnya (opsional)</FormHelperText>
                          </FormControl>
                        </SimpleGrid>
                      </Box>

                      <Divider />

                      {/* Tax Deductions (Pemotongan) */}
                      <Box>
                        <Text fontSize="sm" fontWeight="medium" color={statColors.red} mb={3}>
                          ➖ Tax Deductions (Pemotongan)
                        </Text>
                        <SimpleGrid columns={3} spacing={4}>
                          <FormControl>
                            <FormLabel fontSize="sm">PPh 21 Rate (%)</FormLabel>
                            <NumberInput
                              value={formData.pph21_rate}
                              onChange={(value) => setFormData({...formData, pph21_rate: value})}
                              size="sm"
                              min={0}
                              max={100}
                              step={0.1}
                            >
                              <NumberInputField placeholder="0" />
                              <NumberInputStepper>
                                <NumberIncrementStepper />
                                <NumberDecrementStepper />
                              </NumberInputStepper>
                            </NumberInput>
                            <FormHelperText fontSize="xs">Pajak Penghasilan Pasal 21</FormHelperText>
                          </FormControl>

                          <FormControl>
                            <FormLabel fontSize="sm">PPh 23 Rate (%)</FormLabel>
                            <NumberInput
                              value={formData.pph23_rate}
                              onChange={(value) => setFormData({...formData, pph23_rate: value})}
                              size="sm"
                              min={0}
                              max={100}
                              step={0.1}
                            >
                              <NumberInputField placeholder="0" />
                              <NumberInputStepper>
                                <NumberIncrementStepper />
                                <NumberDecrementStepper />
                              </NumberInputStepper>
                            </NumberInput>
                            <FormHelperText fontSize="xs">Pajak Penghasilan Pasal 23</FormHelperText>
                          </FormControl>

                          <FormControl>
                            <FormLabel fontSize="sm">Other Tax Deductions (%)</FormLabel>
                            <NumberInput
                              value={formData.other_tax_deductions}
                              onChange={(value) => setFormData({...formData, other_tax_deductions: value})}
                              size="sm"
                              min={0}
                              max={100}
                              step={0.1}
                            >
                              <NumberInputField placeholder="0" />
                              <NumberInputStepper>
                                <NumberIncrementStepper />
                                <NumberDecrementStepper />
                              </NumberInputStepper>
                            </NumberInput>
                            <FormHelperText fontSize="xs">Potongan pajak lainnya (opsional)</FormHelperText>
                          </FormControl>
                        </SimpleGrid>
                      </Box>

                      {/* Tax Summary Calculation */}
                      {formData.items.length > 0 && (
                        <Box mt={4} p={4} bg={notesBoxBg} borderRadius="md" border="1px solid" borderColor={borderColor}>
                          <VStack spacing={2} align="stretch">
                            <Text fontSize="sm" fontWeight="semibold" color={textPrimary}>Tax Summary:</Text>
                            {(() => {
                              const subtotal = formData.items.reduce((total, item) => {
                                const qty = parseFloat(item.quantity || '0');
                                const price = parseFloat(item.unit_price || '0');
                                const discount = parseFloat(item.discount || '0');
                                const itemSubtotal = (isNaN(qty) ? 0 : qty) * (isNaN(price) ? 0 : price);
                                const lineTotal = itemSubtotal - (isNaN(discount) ? 0 : discount);
                                return total + lineTotal;
                              }, 0);
                              
                              const discount = (parseFloat(formData.discount) || 0) / 100;
                              const discountedSubtotal = subtotal * (1 - discount);
                              
                              const ppnAmount = discountedSubtotal * (parseFloat(formData.ppn_rate) || 0) / 100;
                              const otherAdditions = discountedSubtotal * (parseFloat(formData.other_tax_additions) || 0) / 100;
                              const totalAdditions = ppnAmount + otherAdditions;
                              
                              const pph21Amount = discountedSubtotal * (parseFloat(formData.pph21_rate) || 0) / 100;
                              const pph23Amount = discountedSubtotal * (parseFloat(formData.pph23_rate) || 0) / 100;
                              const otherDeductions = discountedSubtotal * (parseFloat(formData.other_tax_deductions) || 0) / 100;
                              const totalDeductions = pph21Amount + pph23Amount + otherDeductions;
                              
                              const finalTotal = discountedSubtotal + totalAdditions - totalDeductions;
                              
                              return (
                                <SimpleGrid columns={2} spacing={4} fontSize="xs">
                                  <VStack align="start" spacing={1}>
                                    <Text color={textSecondary}>Subtotal: {formatCurrency(subtotal)}</Text>
                                    <Text color={textSecondary}>Discount ({formData.discount}%): -{formatCurrency(subtotal * discount)}</Text>
                                    <Text color={textSecondary}>After Discount: {formatCurrency(discountedSubtotal)}</Text>
                                  </VStack>
                                  
                                  <VStack align="start" spacing={1}>
                                    <Text color={statColors.green}>+ PPN ({formData.ppn_rate}%): {formatCurrency(ppnAmount)}</Text>
                                    {parseFloat(formData.other_tax_additions) > 0 && (
                                      <Text color={statColors.green}>+ Other Additions ({formData.other_tax_additions}%): {formatCurrency(otherAdditions)}</Text>
                                    )}
                                    {parseFloat(formData.pph21_rate) > 0 && (
                                      <Text color={statColors.red}>- PPh 21 ({formData.pph21_rate}%): {formatCurrency(pph21Amount)}</Text>
                                    )}
                                    {parseFloat(formData.pph23_rate) > 0 && (
                                      <Text color={statColors.red}>- PPh 23 ({formData.pph23_rate}%): {formatCurrency(pph23Amount)}</Text>
                                    )}
                                    {parseFloat(formData.other_tax_deductions) > 0 && (
                                      <Text color={statColors.red}>- Other Deductions ({formData.other_tax_deductions}%): {formatCurrency(otherDeductions)}</Text>
                                    )}
                                    <Text fontWeight="bold" color={statColors.blue} borderTop="1px solid" borderColor={borderColor} pt={1}>
                                      Final Total: {formatCurrency(finalTotal)}
                                    </Text>
                                  </VStack>
                                </SimpleGrid>
                              );
                            })()}
                          </VStack>
                        </Box>
                      )}
                    </VStack>
                  </CardBody>
                </Card>

                {/* Payment Method Section */}
                <Card>
                  <CardHeader pb={3}>
                    <Text fontSize="md" fontWeight="semibold" color={textPrimary}>
                      💳 Payment Method
                    </Text>
                  </CardHeader>
                  <CardBody pt={0}>
                    <SimpleGrid columns={2} spacing={4}>
                      <FormControl isRequired>
                        <FormLabel fontSize="sm" fontWeight="medium">Payment Method</FormLabel>
                        <Select
                          value={formData.payment_method}
                          onChange={(e) => setFormData({
                            ...formData, 
                            payment_method: e.target.value,
                            bank_account_id: (e.target.value === 'CASH' || e.target.value === 'CREDIT') ? '' : formData.bank_account_id
                          })}
                          size="sm"
                        >
                          <option value="CREDIT">Credit</option>
                          <option value="CASH">Cash</option>
                          <option value="BANK_TRANSFER">Bank Transfer</option>
                          <option value="CHECK">Check</option>
                        </Select>
                        <FormHelperText fontSize="xs">
                          {formData.payment_method === 'CREDIT' && 'Purchase on credit - payment due later'}
                          {formData.payment_method === 'CASH' && 'Direct cash payment'}
                          {formData.payment_method === 'BANK_TRANSFER' && 'Electronic bank transfer'}
                          {formData.payment_method === 'CHECK' && 'Payment by check'}
                        </FormHelperText>
                      </FormControl>

                      {/* Bank Account dropdown for Bank Transfer, Cash, Check */}
                      {formData.payment_method !== 'CREDIT' && formData.payment_method !== 'CASH' && (
                        <FormControl isRequired>
                          <FormLabel fontSize="sm" fontWeight="medium">
                            Bank Account
                          </FormLabel>
                          <Select
                            value={formData.bank_account_id}
                            onChange={(e) => setFormData({...formData, bank_account_id: e.target.value})}
                            size="sm"
                            disabled={loadingBankAccounts}
                            placeholder={loadingBankAccounts ? 'Loading accounts...' : 'Select bank account'}
                          >
                            {bankAccounts.map((account) => (
                              <option key={account.id} value={account.id.toString()}>
                                {account.name} ({account.code}) - {account.currency} {account.balance?.toLocaleString() || '0'}
                              </option>
                            ))}
                          </Select>
                          <FormHelperText fontSize="xs">
                            Required: Select account for payment processing
                          </FormHelperText>
                        </FormControl>
                      )}
                      
                      {/* Credit Account dropdown for Credit payment */}
                      {formData.payment_method === 'CREDIT' && (
                        <FormControl isRequired>
                          <FormLabel fontSize="sm" fontWeight="medium">
                            Liability Account
                          </FormLabel>
                          <Select
                            value={formData.credit_account_id}
                            onChange={(e) => setFormData({...formData, credit_account_id: e.target.value})}
                            size="sm"
                            disabled={loadingCreditAccounts}
                            placeholder={loadingCreditAccounts ? 'Loading accounts...' : 'Select liability account'}
                          >
                            {creditAccounts.map((account) => (
                              <option key={account.id} value={account.id?.toString()}>
                                {account.code} - {account.name}
                              </option>
                            ))}
                          </Select>
                          <FormHelperText fontSize="xs">
                            Required: Select liability account for tracking credit purchases
                          </FormHelperText>
                        </FormControl>
                      )}
                    </SimpleGrid>

                    {/* Payment Reference (for non-credit and non-cash payments) */}
                    {formData.payment_method !== 'CREDIT' && formData.payment_method !== 'CASH' && (
                      <FormControl mt={4}>
                        <FormLabel fontSize="sm" fontWeight="medium">Payment Reference</FormLabel>
                        <Input
                          value={formData.payment_reference}
                          onChange={(e) => setFormData({...formData, payment_reference: e.target.value})}
                          placeholder={
                            formData.payment_method === 'CHECK' ? 'Check number' :
                            formData.payment_method === 'BANK_TRANSFER' ? 'Transaction ID or reference number' :
                            'Payment reference number'
                          }
                          size="sm"
                        />
                        <FormHelperText fontSize="xs">
                          Optional reference for tracking this payment
                        </FormHelperText>
                      </FormControl>
                    )}
                  </CardBody>
                </Card>

              </VStack>
            </ModalBody>
            <ModalFooter>
              <HStack spacing={3}>
                <Button variant="ghost" onClick={onEditClose}>
                  Cancel
                </Button>
                <Button colorScheme="blue" onClick={handleSave}>
                  Update Purchase
                </Button>
              </HStack>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* Create Purchase Modal */}
        <Modal isOpen={isCreateOpen} onClose={onCreateClose} size="6xl">
          <ModalOverlay />
          <ModalContent maxW="95vw" maxH="95vh" bg={modalContentBg}>
            <ModalHeader bg={modalHeaderBg} borderRadius="md" mx={4} mt={4} mb={2} borderBottomWidth={1} borderColor={borderColor}>
              <HStack>
                <Box w={1} h={6} bg={statColors.blue} borderRadius="full" />
                <Text fontSize="lg" fontWeight="bold" color={statColors.blue}>
                  Create New Purchase
                </Text>
              </HStack>
            </ModalHeader>
            <ModalCloseButton top={6} right={6} />
            <ModalBody overflowY="auto" px={6} pb={2}>
              <VStack spacing={6} align="stretch">
                {/* Basic Information Section */}
                <Card>
                  <CardHeader pb={3}>
                    <Text fontSize="md" fontWeight="semibold" color={textPrimary}>
                      📋 Basic Information
                    </Text>
                  </CardHeader>
                  <CardBody pt={0}>
                    <SimpleGrid columns={3} spacing={4}>
                      <FormControl isRequired>
                        <FormLabel fontSize="sm" fontWeight="medium">Vendor</FormLabel>
                        <HStack spacing={2}>
                          {loadingVendors ? (
                            <Spinner size="sm" />
                          ) : (
                            <Select
                              placeholder="Select vendor"
                              value={formData.vendor_id}
                              onChange={(e) => setFormData({...formData, vendor_id: e.target.value})}
                              size="sm"
                              flex={1}
                            >
                              {vendors.map(vendor => (
                                <option key={vendor.id} value={vendor.id}>
                                  {vendor.name} ({vendor.code})
                                </option>
                              ))}
                            </Select>
                          )}
                          <IconButton
                            aria-label="Add new vendor"
                            icon={<FiPlus />}
                            size="sm"
                            colorScheme="green"
                            variant="outline"
                            onClick={onAddVendorOpen}
                            title="Add New Vendor"
                            _hover={{ bg: 'green.50' }}
                          />
                        </HStack>
                      </FormControl>
                      
                      <FormControl isRequired>
                        <FormLabel fontSize="sm" fontWeight="medium">Purchase Date</FormLabel>
                        <Input
                          type="date"
                          size="sm"
                          value={formData.date}
                          onChange={(e) => setFormData({...formData, date: e.target.value})}
                        />
                      </FormControl>

                      <FormControl>
                        <FormLabel fontSize="sm" fontWeight="medium">Due Date</FormLabel>
                        <Input
                          type="date"
                          size="sm"
                          value={formData.due_date}
                          onChange={(e) => setFormData({...formData, due_date: e.target.value})}
                        />
                      </FormControl>
                    </SimpleGrid>

                    <SimpleGrid columns={2} spacing={4} mt={4}>
                      <FormControl>
                        <FormLabel fontSize="sm" fontWeight="medium">Discount (%)</FormLabel>
                        <NumberInput
                          value={formData.discount}
                          onChange={(value) => setFormData({...formData, discount: value})}
                          size="sm"
                          min={0}
                          max={100}
                        >
                          <NumberInputField placeholder="0" />
                          <NumberInputStepper>
                            <NumberIncrementStepper />
                            <NumberDecrementStepper />
                          </NumberInputStepper>
                        </NumberInput>
                        <FormHelperText fontSize="xs">Masukkan persentase diskon atas subtotal (0-100)</FormHelperText>
                      </FormControl>

                      <FormControl>
                        <FormLabel fontSize="sm" fontWeight="medium">Notes</FormLabel>
                        <Textarea
                          value={formData.notes}
                          onChange={(e) => setFormData({...formData, notes: e.target.value})}
                          placeholder="Enter any notes or descriptions..."
                          rows={3}
                          size="sm"
                          resize="vertical"
                        />
                      </FormControl>
                    </SimpleGrid>
                  </CardBody>
                </Card>

                {/* Purchase Items Section */}
                <Card>
                  <CardHeader pb={3}>
                    <Flex justify="space-between" align="center">
                      <Text fontSize="md" fontWeight="semibold" color={textPrimary}>
                        🛒 Purchase Items
                      </Text>
                      <Button 
                        size="sm" 
                        leftIcon={<FiPlus />} 
                        colorScheme="blue"
                        variant="outline"
                        onClick={() => {
                          setFormData({
                            ...formData,
                            items: [
                              ...formData.items,
                              { product_id: '', quantity: '1', unit_price: '0', discount: '0', tax: '0', expense_account_id: '' }
                            ]
                          });
                        }}
                      >
                        Add Item
                      </Button>
                    </Flex>
                  </CardHeader>
                  <CardBody pt={0}>
                    <Box overflow="visible">
                        <Table size="sm" variant="simple">
                        <Thead bg={tableHeaderBg}>
                          <Tr>
                            <Th fontSize="xs" fontWeight="semibold" color={textSecondary}>Product</Th>
                            <Th fontSize="xs" fontWeight="semibold" color={textSecondary} isNumeric>Qty</Th>
                            <Th fontSize="xs" fontWeight="semibold" color={textSecondary} isNumeric>Unit Price (IDR)</Th>
                            <Th fontSize="xs" fontWeight="semibold" color={textSecondary} isNumeric>Discount (IDR)</Th>
                            <Th fontSize="xs" fontWeight="semibold" color={textSecondary}>Expense Account</Th>
                            <Th fontSize="xs" fontWeight="semibold" color={textSecondary} isNumeric>Line Total (IDR)</Th>
                            <Th fontSize="xs" fontWeight="semibold" color={textSecondary} w="60px">Action</Th>
                          </Tr>
                        </Thead>
                        <Tbody>
                          {formData.items.length === 0 ? (
                            <Tr>
                              <Td colSpan={7} textAlign="center" py={8}>
                                <VStack spacing={2}>
                                  <Text fontSize="sm" color={textSecondary}>No items added yet</Text>
                                  <Text fontSize="xs" color={textSecondary}>Click "Add Item" button to start adding purchase items</Text>
                                </VStack>
                              </Td>
                            </Tr>
                          ) : (
                            formData.items.map((item, index) => (
                              <Tr key={index} _hover={{ bg: hoverBg }}>
                                <Td minW="200px">
                                  {loadingProducts ? (
                                    <Flex align="center" justify="center" h="32px">
                                      <Spinner size="sm" />
                                    </Flex>
                                  ) : (
                                    <HStack spacing={2}>
                                      <Select
                                        placeholder="Select product"
                                        value={item.product_id}
                                        onChange={(e) => {
                                          const items = [...formData.items];
                                          items[index] = { ...items[index], product_id: e.target.value };
                                          setFormData({ ...formData, items });
                                        }}
                                        size="sm"
                                        maxW="280px"
                                      >
                                        {products.map((p) => (
                                          <option key={p.id} value={p.id?.toString()}>
                                            {p?.id} - {p?.name || p?.code}
                                          </option>
                                        ))}
                                      </Select>
                                      <IconButton 
                                        aria-label="Add new product"
                                        icon={<FiPlus />}
                                        size="sm"
                                        colorScheme="blue"
                                        variant="outline"
                                        onClick={onAddProductOpen}
                                        title="Add New Product"
                                        _hover={{ bg: 'blue.50' }}
                                      />
                                    </HStack>
                                  )}
                                </Td>
                                <Td isNumeric>
                                  <NumberInput 
                                    size="sm" 
                                    min={1} 
                                    value={item.quantity} 
                                    onChange={(valueString) => {
                                      const items = [...formData.items];
                                      items[index] = { ...items[index], quantity: valueString };
                                      setFormData({ ...formData, items });
                                    }} 
                                    maxW="80px"
                                  >
                                    <NumberInputField textAlign="right" fontSize="sm" />
                                    <NumberInputStepper>
                                      <NumberIncrementStepper />
                                      <NumberDecrementStepper />
                                    </NumberInputStepper>
                                  </NumberInput>
                                </Td>
                                <Td isNumeric>
                                  <Box maxW="160px">
                                    <CurrencyInput
                                      value={parseFloat(item.unit_price) || 0}
                                      onChange={(value) => {
                                        const items = [...formData.items];
                                        items[index] = { ...items[index], unit_price: value.toString() };
                                        setFormData({ ...formData, items });
                                      }}
                                      placeholder="Rp 10.000"
                                      size="sm"
                                      min={0}
                                      showLabel={false}
                                    />
                                  </Box>
                                </Td>
                                <Td isNumeric>
                                  <Box maxW="140px">
                                    <CurrencyInput
                                      value={parseFloat(item.discount) || 0}
                                      onChange={(value) => {
                                        const items = [...formData.items];
                                        items[index] = { ...items[index], discount: value.toString() };
                                        setFormData({ ...formData, items });
                                      }}
                                      placeholder="Rp 0"
                                      size="sm"
                                      min={0}
                                      showLabel={false}
                                    />
                                  </Box>
                                </Td>
                                <Td minW="240px">
                                  {canListExpenseAccounts ? (
                                    <Box maxW="240px">
                                      <SearchableSelect
                                        options={expenseAccounts.map(acc => ({
                                          id: acc.id!,
                                          code: acc.code,
                                          name: acc.name,
                                          active: acc.is_active
                                        }))}
                                        value={item.expense_account_id}
                                        onChange={(value) => {
                                          const items = [...formData.items];
                                          items[index] = { ...items[index], expense_account_id: value.toString() };
                                          setFormData({ ...formData, items });
                                        }}
                                        placeholder="Pilih akun beban..."
                                        isLoading={loadingExpenseAccounts}
                                        displayFormat={(option) => `${option.code} - ${option.name}`}
                                      />
                                    </Box>
                                  ) : (
                                    <NumberInput 
                                      min={1} 
                                      value={item.expense_account_id || (defaultExpenseAccountId ? defaultExpenseAccountId.toString() : '')} 
                                      onChange={(v) => {
                                        const items = [...formData.items];
                                        items[index] = { ...items[index], expense_account_id: v.toString() };
                                        setFormData({ ...formData, items });
                                      }} 
                                      maxW="240px"
                                      size="sm"
                                    >
                                      <NumberInputField placeholder="Expense Account ID" fontSize="sm" />
                                    </NumberInput>
                                  )}
                                </Td>
                                <Td isNumeric>
                                  <Text fontSize="sm" fontWeight="medium" color={statColors.green}>
                                    {(() => {
                                      const qty = parseFloat(item.quantity || '0');
                                      const price = parseFloat(item.unit_price || '0');
                                      const discount = parseFloat(item.discount || '0');
                                      const subtotal = (isNaN(qty) ? 0 : qty) * (isNaN(price) ? 0 : price);
                                      const lineTotal = subtotal - (isNaN(discount) ? 0 : discount);
                                      return formatCurrency(lineTotal);
                                    })()}
                                  </Text>
                                </Td>
                                <Td>
                                  <IconButton
                                    aria-label="Remove item"
                                    size="sm"
                                    colorScheme="red"
                                    variant="ghost"
                                    icon={<FiTrash2 />}
                                    onClick={() => {
                                      const items = [...formData.items];
                                      items.splice(index, 1);
                                      setFormData({ ...formData, items });
                                    }}
                                    _hover={{ bg: 'red.50' }}
                                  />
                                </Td>
                              </Tr>
                            ))
                          )}
                        </Tbody>
                      </Table>
                    </Box>
                    
                    {/* Summary Row */}
                    {formData.items.length > 0 && (
                      <Box mt={4} p={4} bg={statBgColors.blue} borderRadius="md" borderLeft="4px solid" borderLeftColor={statColors.blue}>
                        <Flex justify="space-between" align="center">
                          <Text fontSize="sm" fontWeight="medium" color={textPrimary}>
                            Total Items: {formData.items.length}
                          </Text>
                          <Text fontSize="lg" fontWeight="bold" color={statColors.blue}>
                            Subtotal: {formatCurrency(
                              formData.items.reduce((total, item) => {
                                const qty = parseFloat(item.quantity || '0');
                                const price = parseFloat(item.unit_price || '0');
                                const discount = parseFloat(item.discount || '0');
                                const subtotal = (isNaN(qty) ? 0 : qty) * (isNaN(price) ? 0 : price);
                                const lineTotal = subtotal - (isNaN(discount) ? 0 : discount);
                                return total + lineTotal;
                              }, 0)
                            )}
                          </Text>
                        </Flex>
                      </Box>
                    )}
                    
                    <FormControl>
                      <FormHelperText mt={3} fontSize="xs">
                        📌 Tambahkan minimal 1 item pembelian. Semua field harus diisi dengan benar.
                      </FormHelperText>
                    </FormControl>
                  </CardBody>
                </Card>

                {/* Tax Configuration Section */}
                <Card>
                  <CardHeader pb={3}>
                    <Text fontSize="md" fontWeight="semibold" color={textPrimary}>
                      💰 Tax Configuration
                    </Text>
                  </CardHeader>
                  <CardBody pt={0}>
                    <VStack spacing={4} align="stretch">
                      {/* Tax Additions (Penambahan) */}
                      <Box>
                        <Text fontSize="sm" fontWeight="medium" color={statColors.green} mb={3}>
                          ➕ Tax Additions (Penambahan)
                        </Text>
                        <SimpleGrid columns={2} spacing={4}>
                          <FormControl>
                            <FormLabel fontSize="sm">PPN Rate (%)</FormLabel>
                            <NumberInput
                              value={formData.ppn_rate}
                              onChange={(value) => setFormData({...formData, ppn_rate: value})}
                              size="sm"
                              min={0}
                              max={100}
                              step={0.1}
                            >
                              <NumberInputField placeholder="11" />
                              <NumberInputStepper>
                                <NumberIncrementStepper />
                                <NumberDecrementStepper />
                              </NumberInputStepper>
                            </NumberInput>
                            <FormHelperText fontSize="xs">Pajak Pertambahan Nilai (default 11%)</FormHelperText>
                          </FormControl>

                          <FormControl>
                            <FormLabel fontSize="sm">Other Tax Additions (%)</FormLabel>
                            <NumberInput
                              value={formData.other_tax_additions}
                              onChange={(value) => setFormData({...formData, other_tax_additions: value})}
                              size="sm"
                              min={0}
                              max={100}
                              step={0.1}
                            >
                              <NumberInputField placeholder="0" />
                              <NumberInputStepper>
                                <NumberIncrementStepper />
                                <NumberDecrementStepper />
                              </NumberInputStepper>
                            </NumberInput>
                            <FormHelperText fontSize="xs">Pajak tambahan lainnya (opsional)</FormHelperText>
                          </FormControl>
                        </SimpleGrid>
                      </Box>

                      <Divider />

                      {/* Tax Deductions (Pemotongan) */}
                      <Box>
                        <Text fontSize="sm" fontWeight="medium" color={statColors.red} mb={3}>
                          ➖ Tax Deductions (Pemotongan)
                        </Text>
                        <SimpleGrid columns={3} spacing={4}>
                          <FormControl>
                            <FormLabel fontSize="sm">PPh 21 Rate (%)</FormLabel>
                            <NumberInput
                              value={formData.pph21_rate}
                              onChange={(value) => setFormData({...formData, pph21_rate: value})}
                              size="sm"
                              min={0}
                              max={100}
                              step={0.1}
                            >
                            <NumberInputField placeholder="2" />
                              <NumberInputStepper>
                                <NumberIncrementStepper />
                                <NumberDecrementStepper />
                              </NumberInputStepper>
                            </NumberInput>
                            <FormHelperText fontSize="xs">PPh 21: 2% jasa konstruksi, 15% dividen/bunga</FormHelperText>
                          </FormControl>

                          <FormControl>
                            <FormLabel fontSize="sm">PPh 23 Rate (%)</FormLabel>
                            <NumberInput
                              value={formData.pph23_rate}
                              onChange={(value) => setFormData({...formData, pph23_rate: value})}
                              size="sm"
                              min={0}
                              max={100}
                              step={0.1}
                            >
                              <NumberInputField placeholder="2" />
                              <NumberInputStepper>
                                <NumberIncrementStepper />
                                <NumberDecrementStepper />
                              </NumberInputStepper>
                            </NumberInput>
                            <FormHelperText fontSize="xs">PPh 23: 2% jasa umum, 15% dividen/bunga/royalti</FormHelperText>
                          </FormControl>

                          <FormControl>
                            <FormLabel fontSize="sm">Other Tax Deductions (%)</FormLabel>
                            <NumberInput
                              value={formData.other_tax_deductions}
                              onChange={(value) => setFormData({...formData, other_tax_deductions: value})}
                              size="sm"
                              min={0}
                              max={100}
                              step={0.1}
                            >
                              <NumberInputField placeholder="0" />
                              <NumberInputStepper>
                                <NumberIncrementStepper />
                                <NumberDecrementStepper />
                              </NumberInputStepper>
                            </NumberInput>
                            <FormHelperText fontSize="xs">Potongan pajak lainnya (opsional)</FormHelperText>
                          </FormControl>
                        </SimpleGrid>
                      </Box>

                      {/* Tax Summary Calculation */}
                      {formData.items.length > 0 && (
                        <Box mt={4} p={4} bg={notesBoxBg} borderRadius="md" border="1px solid" borderColor={borderColor}>
                          <VStack spacing={2} align="stretch">
                            <Text fontSize="sm" fontWeight="semibold" color={textPrimary}>Tax Summary:</Text>
                            {(() => {
                              const subtotal = formData.items.reduce((total, item) => {
                                const qty = parseFloat(item.quantity || '0');
                                const price = parseFloat(item.unit_price || '0');
                                const discount = parseFloat(item.discount || '0');
                                const itemSubtotal = (isNaN(qty) ? 0 : qty) * (isNaN(price) ? 0 : price);
                                const lineTotal = itemSubtotal - (isNaN(discount) ? 0 : discount);
                                return total + lineTotal;
                              }, 0);
                              
                              const discount = (parseFloat(formData.discount) || 0) / 100;
                              const discountedSubtotal = subtotal * (1 - discount);
                              
                              const ppnAmount = discountedSubtotal * (parseFloat(formData.ppn_rate) || 0) / 100;
                              const otherAdditions = discountedSubtotal * (parseFloat(formData.other_tax_additions) || 0) / 100;
                              const totalAdditions = ppnAmount + otherAdditions;
                              
                              const pph21Amount = discountedSubtotal * (parseFloat(formData.pph21_rate) || 0) / 100;
                              const pph23Amount = discountedSubtotal * (parseFloat(formData.pph23_rate) || 0) / 100;
                              const otherDeductions = discountedSubtotal * (parseFloat(formData.other_tax_deductions) || 0) / 100;
                              const totalDeductions = pph21Amount + pph23Amount + otherDeductions;
                              
                              const finalTotal = discountedSubtotal + totalAdditions - totalDeductions;
                              
                              return (
                                <SimpleGrid columns={2} spacing={4} fontSize="xs">
                                  <VStack align="start" spacing={1}>
                                    <Text color={textSecondary}>Subtotal: {formatCurrency(subtotal)}</Text>
                                    <Text color={textSecondary}>Discount ({formData.discount}%): -{formatCurrency(subtotal * discount)}</Text>
                                    <Text color={textSecondary}>After Discount: {formatCurrency(discountedSubtotal)}</Text>
                                  </VStack>
                                  
                                  <VStack align="start" spacing={1}>
                                    <Text color={statColors.green}>+ PPN ({formData.ppn_rate}%): {formatCurrency(ppnAmount)}</Text>
                                    {parseFloat(formData.other_tax_additions) > 0 && (
                                      <Text color={statColors.green}>+ Other Additions ({formData.other_tax_additions}%): {formatCurrency(otherAdditions)}</Text>
                                    )}
                                    {parseFloat(formData.pph21_rate) > 0 && (
                                      <Text color={statColors.red}>- PPh 21 ({formData.pph21_rate}%): {formatCurrency(pph21Amount)}</Text>
                                    )}
                                    {parseFloat(formData.pph23_rate) > 0 && (
                                      <Text color={statColors.red}>- PPh 23 ({formData.pph23_rate}%): {formatCurrency(pph23Amount)}</Text>
                                    )}
                                    {parseFloat(formData.other_tax_deductions) > 0 && (
                                      <Text color={statColors.red}>- Other Deductions ({formData.other_tax_deductions}%): {formatCurrency(otherDeductions)}</Text>
                                    )}
                                    <Text fontWeight="bold" color={statColors.blue} borderTop="1px solid" borderColor={borderColor} pt={1}>
                                      Final Total: {formatCurrency(finalTotal)}
                                    </Text>
                                  </VStack>
                                </SimpleGrid>
                              );
                            })()}
                          </VStack>
                        </Box>
                      )}
                    </VStack>
                  </CardBody>
                </Card>

                {/* Payment Method Section */}
                <Card>
                  <CardHeader pb={3}>
                    <Text fontSize="md" fontWeight="semibold" color={textPrimary}>
                      💳 Payment Method
                    </Text>
                  </CardHeader>
                  <CardBody pt={0}>
                    <SimpleGrid columns={2} spacing={4}>
                      <FormControl isRequired>
                        <FormLabel fontSize="sm" fontWeight="medium">Payment Method</FormLabel>
                        <Select
                          value={formData.payment_method}
                          onChange={(e) => setFormData({
                            ...formData, 
                            payment_method: e.target.value,
                            bank_account_id: (e.target.value === 'CASH' || e.target.value === 'CREDIT') ? '' : formData.bank_account_id
                          })}
                          size="sm"
                        >
                          <option value="CREDIT">Credit</option>
                          <option value="CASH">Cash</option>
                          <option value="BANK_TRANSFER">Bank Transfer</option>
                          <option value="CHECK">Check</option>
                        </Select>
                        <FormHelperText fontSize="xs">
                          {formData.payment_method === 'CREDIT' && 'Purchase on credit - payment due later'}
                          {formData.payment_method === 'CASH' && 'Direct cash payment'}
                          {formData.payment_method === 'BANK_TRANSFER' && 'Electronic bank transfer'}
                          {formData.payment_method === 'CHECK' && 'Payment by check'}
                        </FormHelperText>
                      </FormControl>

                      {/* Bank Account dropdown for Bank Transfer, Cash, Check */}
                      {formData.payment_method !== 'CREDIT' && formData.payment_method !== 'CASH' && (
                        <FormControl isRequired>
                          <FormLabel fontSize="sm" fontWeight="medium">
                            Bank Account
                          </FormLabel>
                          <Select
                            value={formData.bank_account_id}
                            onChange={(e) => setFormData({...formData, bank_account_id: e.target.value})}
                            size="sm"
                            disabled={loadingBankAccounts}
                            placeholder={loadingBankAccounts ? 'Loading accounts...' : 'Select bank account'}
                          >
                            {bankAccounts.map((account) => (
                              <option key={account.id} value={account.id.toString()}>
                                {account.name} ({account.code}) - {account.currency} {account.balance?.toLocaleString() || '0'}
                              </option>
                            ))}
                          </Select>
                          <FormHelperText fontSize="xs">
                            Required: Select account for payment processing
                          </FormHelperText>
                        </FormControl>
                      )}
                      
                      {/* Credit Account dropdown for Credit payment */}
                      {formData.payment_method === 'CREDIT' && (
                        <FormControl isRequired>
                          <FormLabel fontSize="sm" fontWeight="medium">
                            Liability Account
                          </FormLabel>
                          <Select
                            value={formData.credit_account_id}
                            onChange={(e) => setFormData({...formData, credit_account_id: e.target.value})}
                            size="sm"
                            disabled={loadingCreditAccounts}
                            placeholder={loadingCreditAccounts ? 'Loading accounts...' : 'Select liability account'}
                          >
                            {creditAccounts.map((account) => (
                              <option key={account.id} value={account.id?.toString()}>
                                {account.code} - {account.name}
                              </option>
                            ))}
                          </Select>
                          <FormHelperText fontSize="xs">
                            Required: Select liability account for tracking credit purchases
                          </FormHelperText>
                        </FormControl>
                      )}
                    </SimpleGrid>

                    {/* Payment Reference (for non-credit and non-cash payments) */}
                    {formData.payment_method !== 'CREDIT' && formData.payment_method !== 'CASH' && (
                      <FormControl mt={4}>
                        <FormLabel fontSize="sm" fontWeight="medium">Payment Reference</FormLabel>
                        <Input
                          value={formData.payment_reference}
                          onChange={(e) => setFormData({...formData, payment_reference: e.target.value})}
                          placeholder={
                            formData.payment_method === 'CHECK' ? 'Check number' :
                            formData.payment_method === 'BANK_TRANSFER' ? 'Transaction ID or reference number' :
                            'Payment reference number'
                          }
                          size="sm"
                        />
                        <FormHelperText fontSize="xs">
                          Optional reference for tracking this payment
                        </FormHelperText>
                      </FormControl>
                    )}
                  </CardBody>
                </Card>
              </VStack>
            </ModalBody>
            <ModalFooter>
              <HStack spacing={3}>
                <Button variant="ghost" onClick={onCreateClose}>
                  Cancel
                </Button>
                <Button colorScheme="blue" onClick={handleSave}>
                  Create Purchase
                </Button>
              </HStack>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* Add Vendor Modal */}
        <Modal isOpen={isAddVendorOpen} onClose={onAddVendorClose} size="lg">
          <ModalOverlay />
          <ModalContent bg={modalContentBg}>
            <ModalHeader bg={modalHeaderBg} borderBottomWidth={1} borderColor={borderColor}>
              <HStack>
                <Box w={1} h={6} bg={statColors.green} borderRadius="full" />
                <Text fontSize="lg" fontWeight="bold" color={statColors.green}>
                  Add New Vendor
                </Text>
              </HStack>
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <VStack spacing={4} align="stretch">
                <SimpleGrid columns={2} spacing={4}>
                  <FormControl isRequired>
                    <FormLabel fontSize="sm">Vendor Name</FormLabel>
                    <Input
                      size="sm"
                      placeholder="Enter vendor name"
                      value={newVendorData.name}
                      onChange={(e) => setNewVendorData({...newVendorData, name: e.target.value})}
                    />
                  </FormControl>
                  
                  <FormControl>
                    <FormLabel fontSize="sm">Vendor Code</FormLabel>
                    <Input
                      size="sm"
                      placeholder="Auto-generated if empty"
                      value={newVendorData.code}
                      onChange={(e) => setNewVendorData({...newVendorData, code: e.target.value})}
                    />
                  </FormControl>
                </SimpleGrid>
                
                <SimpleGrid columns={2} spacing={4}>
                  <FormControl isRequired>
                    <FormLabel fontSize="sm">Email</FormLabel>
                    <Input
                      size="sm"
                      type="email"
                      placeholder="vendor@company.com"
                      value={newVendorData.email}
                      onChange={(e) => setNewVendorData({...newVendorData, email: e.target.value})}
                    />
                  </FormControl>
                  
                  <FormControl>
                    <FormLabel fontSize="sm">Phone</FormLabel>
                    <Input
                      size="sm"
                      placeholder="Enter phone number"
                      value={newVendorData.phone}
                      onChange={(e) => setNewVendorData({...newVendorData, phone: e.target.value})}
                    />
                  </FormControl>
                </SimpleGrid>
                
                <SimpleGrid columns={2} spacing={4}>
                  <FormControl>
                    <FormLabel fontSize="sm">Mobile</FormLabel>
                    <Input
                      size="sm"
                      placeholder="Enter mobile number"
                      value={newVendorData.mobile}
                      onChange={(e) => setNewVendorData({...newVendorData, mobile: e.target.value})}
                    />
                  </FormControl>
                  
                  <FormControl>
                    <FormLabel fontSize="sm">PIC Name</FormLabel>
                    <Input
                      size="sm"
                      placeholder="Person in charge"
                      value={newVendorData.pic_name}
                      onChange={(e) => setNewVendorData({...newVendorData, pic_name: e.target.value})}
                    />
                  </FormControl>
                </SimpleGrid>
                
                <FormControl>
                  <FormLabel fontSize="sm">Vendor ID</FormLabel>
                  <Input
                    size="sm"
                    placeholder="External vendor ID (optional)"
                    value={newVendorData.external_id}
                    onChange={(e) => setNewVendorData({...newVendorData, external_id: e.target.value})}
                  />
                </FormControl>
                
                <FormControl>
                  <FormLabel fontSize="sm">Address</FormLabel>
                  <Textarea
                    size="sm"
                    placeholder="Enter vendor address"
                    rows={3}
                    value={newVendorData.address}
                    onChange={(e) => setNewVendorData({...newVendorData, address: e.target.value})}
                  />
                </FormControl>
                
                <FormControl>
                  <FormLabel fontSize="sm">Notes</FormLabel>
                  <Textarea
                    size="sm"
                    placeholder="Additional notes (optional)"
                    rows={2}
                    value={newVendorData.notes}
                    onChange={(e) => setNewVendorData({...newVendorData, notes: e.target.value})}
                  />
                </FormControl>
              </VStack>
            </ModalBody>
            <ModalFooter>
              <HStack spacing={3}>
                <Button
                  variant="ghost"
                  onClick={() => {
                    setNewVendorData({
                      name: '',
                      code: '',
                      email: '',
                      phone: '',
                      mobile: '',
                      address: '',
                      pic_name: '',
                      external_id: '',
                      notes: ''
                    });
                    onAddVendorClose();
                  }}
                  disabled={savingVendor}
                >
                  Cancel
                </Button>
                <Button
                  colorScheme="green"
                  onClick={handleAddVendor}
                  isLoading={savingVendor}
                  loadingText="Creating..."
                >
                  Create Vendor
                </Button>
              </HStack>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* Add Product Modal */}
        <Modal isOpen={isAddProductOpen} onClose={onAddProductClose} size="lg">
          <ModalOverlay />
          <ModalContent bg={modalContentBg}>
            <ModalHeader bg={modalHeaderBg} borderBottomWidth={1} borderColor={borderColor}>
              <HStack>
                <Box w={1} h={6} bg={statColors.blue} borderRadius="full" />
                <Text fontSize="lg" fontWeight="bold" color={statColors.blue}>
                  Add New Product
                </Text>
              </HStack>
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <VStack spacing={4} align="stretch">
                <FormControl isRequired>
                  <FormLabel fontSize="sm">Product Name</FormLabel>
                  <Input
                    size="sm"
                    placeholder="Enter product name"
                    value={newProductData.name}
                    onChange={(e) => setNewProductData({ ...newProductData, name: e.target.value })}
                  />
                </FormControl>
                
                <FormControl>
                  <FormLabel fontSize="sm">Product Code</FormLabel>
                  <Input
                    size="sm"
                    placeholder="Enter product code (optional)"
                    value={newProductData.code}
                    onChange={(e) => setNewProductData({ ...newProductData, code: e.target.value })}
                  />
                </FormControl>
                
                <FormControl>
                  <FormLabel fontSize="sm">Description</FormLabel>
                  <Textarea
                    size="sm"
                    placeholder="Enter product description"
                    value={newProductData.description}
                    onChange={(e) => setNewProductData({ ...newProductData, description: e.target.value })}
                  />
                </FormControl>
                
                <SimpleGrid columns={3} spacing={4}>
                  <FormControl isRequired>
                    <FormLabel fontSize="sm">Unit</FormLabel>
                    <Input
                      size="sm"
                      placeholder="e.g., pcs, kg, box"
                      value={newProductData.unit}
                      onChange={(e) => setNewProductData({ ...newProductData, unit: e.target.value })}
                    />
                  </FormControl>
                  
                  <FormControl>
                    <FormLabel fontSize="sm">Purchase Price (IDR)</FormLabel>
                    <CurrencyInput
                      value={parseFloat(newProductData.purchase_price) || 0}
                      onChange={(value) => setNewProductData({ ...newProductData, purchase_price: value.toString() })}
                      placeholder="Rp 10.000"
                      size="sm"
                      min={0}
                      showLabel={false}
                    />
                  </FormControl>
                  
                  <FormControl>
                    <FormLabel fontSize="sm">Sale Price (IDR)</FormLabel>
                    <CurrencyInput
                      value={parseFloat(newProductData.sale_price) || 0}
                      onChange={(value) => setNewProductData({ ...newProductData, sale_price: value.toString() })}
                      placeholder="Rp 15.000"
                      size="sm"
                      min={0}
                      showLabel={false}
                    />
                  </FormControl>
                </SimpleGrid>
              </VStack>
            </ModalBody>
            <ModalFooter>
              <HStack spacing={3} w="100%">
                <Button
                  variant="ghost"
                  onClick={() => {
                    setNewProductData({
                      name: '',
                      code: '',
                      description: '',
                      unit: '',
                      purchase_price: '0',
                      sale_price: '0',
                    });
                    onAddProductClose();
                  }}
                  flex={1}
                >
                  Cancel
                </Button>
                <Button
                  colorScheme="blue"
                  onClick={handleAddProduct}
                  isLoading={savingProduct}
                  loadingText="Creating..."
                  flex={1}
                >
                  Create Product
                </Button>
              </HStack>
            </ModalFooter>
          </ModalContent>
        </Modal>
        </VStack>
      </Box>
    </SimpleLayout>
  );
};

export default PurchasesPage;
