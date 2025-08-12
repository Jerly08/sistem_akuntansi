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
import SearchableSelect from '@/components/common/SearchableSelect';

// Types for form data
interface PurchaseFormData {
  vendor_id: string;
  date: string;
  due_date: string;
  notes: string;
  discount: string;
  tax: string;
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
  }).format(amount);
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
  });

  // View and Edit Modal states
  const { isOpen: isViewOpen, onOpen: onViewOpen, onClose: onViewClose } = useDisclosure();
  const { isOpen: isEditOpen, onOpen: onEditOpen, onClose: onEditClose } = useDisclosure();
  const { isOpen: isCreateOpen, onOpen: onCreateOpen, onClose: onCreateClose } = useDisclosure();
  
  const [selectedPurchase, setSelectedPurchase] = useState<Purchase | null>(null);
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [expenseAccounts, setExpenseAccounts] = useState<GLAccount[]>([]);
  const [loadingExpenseAccounts, setLoadingExpenseAccounts] = useState(false);
  const [defaultExpenseAccountId, setDefaultExpenseAccountId] = useState<number | null>(null);
  const [canListExpenseAccounts, setCanListExpenseAccounts] = useState(true);
  const [formData, setFormData] = useState<PurchaseFormData>({
    vendor_id: '',
    date: new Date().toISOString().split('T')[0],
    due_date: '',
    notes: '',
    discount: '0',
    tax: '0',
    items: []
  });
  const [loadingVendors, setLoadingVendors] = useState(false);
  const [loadingProducts, setLoadingProducts] = useState(false);

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
      
      setStats({
        total: response?.total || totalPurchases, // Use API total if available, otherwise current page count
        pending: pendingApproval,
        approved: approved,
        rejected: rejected,
        needingApproval: pendingApproval, // Same as pending for now
        totalValue: totalValue, // Add total value to stats
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
      
      const response = await fetch('http://localhost:8080/api/v1/contacts?type=VENDOR', {
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
      tax: '0',
      items: []
    });
    setSelectedPurchase(null);
    await fetchVendors(); // Load vendors for dropdown
    await fetchProductsList(); // Load products for dropdown
    await fetchExpenseAccounts(); // Load expense accounts for dropdown
    onCreateOpen();
  };

  // Handle save purchase (both create and edit)
  const handleSave = async () => {
    try {
      // Validation
      if (!formData.vendor_id) {
        toast({
          title: 'Validation Error',
          description: 'Please select a vendor',
          status: 'error',
          duration: 3000,
          isClosable: true,
        });
        return;
      }

      if (!formData.date) {
        toast({
          title: 'Validation Error',
          description: 'Please select a purchase date',
          status: 'error',
          duration: 3000,
          isClosable: true,
        });
        return;
      }

      if (formData.items.length === 0) {
        toast({
          title: 'Validation Error',
          description: 'Please add at least one purchase item',
          status: 'error',
          duration: 3000,
          isClosable: true,
        });
        return;
      }

      // Per-item validation to avoid sending invalid product_id/values
      for (let i = 0; i < formData.items.length; i++) {
        const it = formData.items[i];
        const pid = parseInt(it.product_id);
        const qty = parseInt(it.quantity);
        const price = parseFloat(it.unit_price);
        const expId = it.expense_account_id ? parseInt(it.expense_account_id) : (defaultExpenseAccountId ?? 0);

        if (!pid || isNaN(pid) || pid <= 0) {
          toast({
            title: 'Validation Error',
            description: `Item #${i + 1}: Please select a valid product`,
            status: 'error',
            duration: 3500,
            isClosable: true,
          });
          return;
        }
        if (!qty || isNaN(qty) || qty <= 0) {
          toast({
            title: 'Validation Error',
            description: `Item #${i + 1}: Quantity must be greater than 0`,
            status: 'error',
            duration: 3500,
            isClosable: true,
          });
          return;
        }
        if (isNaN(price) || price < 0) {
          toast({
            title: 'Validation Error',
            description: `Item #${i + 1}: Unit price must be a valid number (>= 0)`,
            status: 'error',
            duration: 3500,
            isClosable: true,
          });
          return;
        }
        if (!expId || isNaN(expId) || expId <= 0) {
          toast({
            title: 'Validation Error',
            description: `Item #${i + 1}: Expense account is required`,
            status: 'error',
            duration: 3500,
            isClosable: true,
          });
          return;
        }
      }

      // Prepare purchase data according to API requirements
      const purchaseData = {
        vendor_id: parseInt(formData.vendor_id),
        date: new Date(formData.date).toISOString(),
        due_date: formData.due_date ? new Date(formData.due_date).toISOString() : undefined,
        notes: formData.notes,
        discount: parseFloat(formData.discount) || 0,
        tax: parseFloat(formData.tax) || 0,
        items: formData.items.map((item) => ({
          product_id: parseInt(item.product_id),
          quantity: parseInt(item.quantity),
          unit_price: parseFloat(item.unit_price),
          discount: parseFloat(item.discount) || 0,
          tax: parseFloat(item.tax) || 0,
          expense_account_id: item.expense_account_id ? parseInt(item.expense_account_id) : (defaultExpenseAccountId ?? undefined),
        }))
      };

      if (selectedPurchase) {
        // Update existing purchase
        await purchaseService.update(selectedPurchase.id, purchaseData);
        toast({
          title: 'Success',
          description: 'Purchase updated successfully',
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
        onEditClose();
      } else {
        // Create new purchase
        await purchaseService.create(purchaseData);
        toast({
          title: 'Success',
          description: 'Purchase created successfully',
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
        onCreateClose();
      }
      
      await fetchPurchases(); // Refresh data
    } catch (err: any) {
      console.error('Save error:', err);
      
      let errorMessage = `Failed to ${selectedPurchase ? 'update' : 'create'} purchase`;
      
      // Handle specific error cases
      if (err.response?.status === 404) {
        errorMessage = 'Purchase API endpoint not found. The purchase management feature is not yet fully implemented on the backend.';
      } else if (err.response?.data?.error) {
        errorMessage = err.response.data.error;
      } else if (err.response?.data?.message) {
        errorMessage = err.response.data.message;
      } else if (err.message) {
        errorMessage = err.message;
      }
      
      toast({
        title: 'API Not Available',
        description: errorMessage,
        status: 'warning',
        duration: 8000,
        isClosable: true,
      });
    }
  };

  // Smart Action Button Logic
  const getActionButtonProps = (purchase: Purchase) => {
    const roleNorm = normalizeRole(user?.role as any);
    const status = (purchase.approval_status || '').toUpperCase();
    
    // Check active step (simplified version, should be more robust in real-world)
    // This is a placeholder logic. For a real app, the active step should come from the API.
    const getActiveStep = () => {
      if (status === 'PENDING') {
        // A more robust check would look at the last history item
        if (purchase.total_amount > 25000000) return 'director';
        return 'finance';
      }
      if (status === 'NOT_STARTED' || status === 'DRAFT') return 'finance';
      return null;
    };
    
    const activeStep = getActiveStep();
    const isUserTurn = activeStep === roleNorm;

    if (status === 'APPROVED' || status === 'REJECTED' || status === 'CANCELLED') {
      return { text: 'View', icon: <FiEye />, colorScheme: 'gray', variant: 'outline' };
    }

    if (isUserTurn) {
      return { text: 'Action Required', icon: <FiAlertCircle />, colorScheme: 'orange', variant: 'solid' };
    }

    if (status === 'PENDING') {
      return { text: 'Review Progress', icon: <FiClock />, colorScheme: 'blue', variant: 'outline' };
    }
    
    return { text: 'View', icon: <FiEye />, colorScheme: 'gray', variant: 'outline' };
  };

  // Action buttons for each row
  const renderActions = (purchase: Purchase) => {
    const actionProps = getActionButtonProps(purchase);
    
    return (
      <HStack spacing={2}>
        {/* Smart Single Action Button */}
        <Button
          size="sm"
          variant={actionProps.variant}
          colorScheme={actionProps.colorScheme}
          leftIcon={actionProps.icon}
          onClick={() => {
            setSelectedPurchase(purchase);
            onViewOpen();
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
        {user?.role === 'ADMIN' && (
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
        
        {/* Delete button for DIRECTOR - only DRAFT purchases */}
        {user?.role === 'DIRECTOR' && (purchase.status || '').toUpperCase() === 'DRAFT' && (
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
      </HStack>
    );
  };

  if (loading) {
    return (
<Layout allowedRoles={['admin', 'finance', 'inventory_manager', 'employee', 'director']}>
        <Box>
          <Text>Loading purchases...</Text>
        </Box>
      </Layout>
    );
  }

  return (
<Layout allowedRoles={['admin', 'finance', 'inventory_manager', 'employee', 'director']}>
      <VStack spacing={6} align="stretch">
        {/* Header */}
        <Flex justify="space-between" align="center">
          <Heading size="lg">Purchase Management</Heading>
          <HStack spacing={3}>
            <Button
              variant="outline"
              leftIcon={<FiFilter />}
              onClick={onFilterOpen}
            >
              Filters
            </Button>
            <Button
              variant="outline"
              leftIcon={<FiRefreshCw />}
              onClick={handleRefresh}
              isLoading={loading}
            >
              Refresh
            </Button>
            {/* New Purchase button only for Employee role */}
            {normalizeRole(user?.role as any) === 'employee' && (
              <Button
                colorScheme="blue"
                leftIcon={<FiPlus />}
                onClick={handleCreate}
              >
                New Purchase
              </Button>
            )}
          </HStack>
        </Flex>

        {/* Statistics Cards */}
        <Grid templateColumns="repeat(auto-fit, minmax(250px, 1fr))" gap={4}>
          <Card>
            <CardBody>
              <Stat>
                <StatLabel>Total Purchases</StatLabel>
                <StatNumber>{stats.total}</StatNumber>
              </Stat>
            </CardBody>
          </Card>
          
          <Card>
            <CardBody>
              <Stat>
                <StatLabel>Pending Approval</StatLabel>
                <StatNumber color="orange.500">
                  <HStack>
                    <FiClock />
                    <Text>{stats.needingApproval}</Text>
                  </HStack>
                </StatNumber>
              </Stat>
            </CardBody>
          </Card>
          
          <Card>
            <CardBody>
              <Stat>
                <StatLabel>Approved</StatLabel>
                <StatNumber color="green.500">
                  <HStack>
                    <FiCheckCircle />
                    <Text>{stats.approved}</Text>
                  </HStack>
                </StatNumber>
              </Stat>
            </CardBody>
          </Card>
          
          <Card>
            <CardBody>
              <Stat>
                <StatLabel>Rejected</StatLabel>
                <StatNumber color="red.500">
                  <HStack>
                    <FiXCircle />
                    <Text>{stats.rejected}</Text>
                  </HStack>
                </StatNumber>
              </Stat>
            </CardBody>
          </Card>
          
          <Card>
            <CardBody>
              <Stat>
                <StatLabel>Total Value</StatLabel>
                <StatNumber fontSize="sm">
                  {formatCurrency(stats.totalValue || 0)}
                </StatNumber>
              </Stat>
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
        <Card>
          <CardBody p={0}>
            <DataTable<Purchase>
              columns={columns}
              data={purchases}
              keyField="id"
              title="Purchase Transactions"
              actions={renderActions}
              searchable={true}
              pagination={true}
              pageSize={pagination.limit}
              totalPages={pagination.totalPages}
              currentPage={pagination.page}
              onPageChange={handlePageChange}
              isLoading={loading}
            />
          </CardBody>
        </Card>

        {/* Filter Modal */}
        <Modal isOpen={isFilterOpen} onClose={onFilterClose} size="md">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>Filter Purchases</ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <VStack spacing={4}>
                <FormControl>
                  <FormLabel>Search</FormLabel>
                  <Input
                    placeholder="Search by purchase number, vendor..."
                    value={filters.search || ''}
                    onChange={(e) => handleFilterChange({ search: e.target.value })}
                  />
                </FormControl>
                
                <FormControl>
                  <FormLabel>Status</FormLabel>
                  <Select
                    placeholder="All Statuses"
                    value={filters.status || ''}
                    onChange={(e) => handleFilterChange({ status: e.target.value })}
                  >
                    <option value="draft">Draft</option>
                    <option value="pending_approval">Pending Approval</option>
                    <option value="approved">Approved</option>
                    <option value="cancelled">Cancelled</option>
                  </Select>
                </FormControl>
                
                <FormControl>
                  <FormLabel>Approval Status</FormLabel>
                  <Select
                    placeholder="All Approval Statuses"
                    value={filters.approval_status || ''}
                    onChange={(e) => handleFilterChange({ approval_status: e.target.value })}
                  >
                    <option value="not_required">Not Required</option>
                    <option value="pending">Pending</option>
                    <option value="approved">Approved</option>
                    <option value="rejected">Rejected</option>
                  </Select>
                </FormControl>
              </VStack>
            </ModalBody>
            <Box p={6}>
              <HStack spacing={3}>
                <Button variant="ghost" onClick={onFilterClose} flex={1}>
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
                >
                  Clear Filters
                </Button>
              </HStack>
            </Box>
          </ModalContent>
        </Modal>

        {/* View Purchase Modal */}
        <Modal isOpen={isViewOpen} onClose={onViewClose} size="xl">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>
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
                  
                  {/* Notes */}
                  {selectedPurchase.notes && (
                    <FormControl>
                      <FormLabel>Notes</FormLabel>
                      <Text p={3} bg="gray.50" borderRadius="md">{selectedPurchase.notes}</Text>
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
                        <Table size="sm">
                          <Thead>
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
          <ModalContent>
            <ModalHeader>
              Edit Purchase - {selectedPurchase?.code}
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <VStack spacing={4} align="stretch">
                <Text fontSize="md" fontWeight="semibold">Basic Info</Text>
                <SimpleGrid columns={2} spacing={4}>
                  <FormControl isRequired>
                    <FormLabel>Vendor</FormLabel>
                    {loadingVendors ? (
                      <Spinner size="sm" />
                    ) : (
                      <Select
                        placeholder="Select vendor"
                        value={formData.vendor_id}
                        onChange={(e) => setFormData({...formData, vendor_id: e.target.value})}
                      >
                        {vendors.map(vendor => (
                          <option key={vendor.id} value={vendor.id}>
                            {vendor.name} ({vendor.code})
                          </option>
                        ))}
                      </Select>
                    )}
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
                      <Text fontSize="md" fontWeight="semibold" color="gray.700">
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
                        <Thead bg="gray.50">
                          <Tr>
                            <Th fontSize="xs" fontWeight="semibold" color="gray.600">Product</Th>
                            <Th fontSize="xs" fontWeight="semibold" color="gray.600" isNumeric>Qty</Th>
                            <Th fontSize="xs" fontWeight="semibold" color="gray.600" isNumeric>Unit Price (IDR)</Th>
                            <Th fontSize="xs" fontWeight="semibold" color="gray.600">Expense Account</Th>
                            <Th fontSize="xs" fontWeight="semibold" color="gray.600" isNumeric>Total (IDR)</Th>
                            <Th fontSize="xs" fontWeight="semibold" color="gray.600" w="60px">Action</Th>
                          </Tr>
                        </Thead>
                        <Tbody>
                          {formData.items.length === 0 ? (
                            <Tr>
                              <Td colSpan={6} textAlign="center" py={8}>
                                <VStack spacing={2}>
                                  <Text fontSize="sm" color="gray.500">No items added yet</Text>
                                  <Text fontSize="xs" color="gray.400">Click "Add Item" button to start adding purchase items</Text>
                                </VStack>
                              </Td>
                            </Tr>
                          ) : (
                            formData.items.map((item, index) => (
                              <Tr key={index} _hover={{ bg: 'gray.50' }}>
                                <Td minW="200px">
                                  {loadingProducts ? (
                                    <Flex align="center" justify="center" h="32px">
                                      <Spinner size="sm" />
                                    </Flex>
                                  ) : (
                                    <Select
                                      placeholder="Select product"
                                      value={item.product_id}
                                      onChange={(e) => {
                                        const items = [...formData.items];
                                        items[index] = { ...items[index], product_id: e.target.value };
                                        setFormData({ ...formData, items });
                                      }}
                                      size="sm"
                                    >
                                      {products.map((p) => (
                                        <option key={p.id} value={p.id?.toString()}>
                                          {p?.id} - {p?.name || p?.code}
                                        </option>
                                      ))}
                                    </Select>
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
                                  <NumberInput 
                                    size="sm" 
                                    min={0} 
                                    precision={0} 
                                    value={item.unit_price} 
                                    onChange={(valueString) => {
                                      const items = [...formData.items];
                                      items[index] = { ...items[index], unit_price: valueString };
                                      setFormData({ ...formData, items });
                                    }} 
                                    maxW="160px"
                                    clampValueOnBlur={false}
                                  >
                                    <NumberInputField 
                                      textAlign="right" 
                                      fontSize="sm" 
                                      placeholder="Masukkan harga"
                                      inputMode="numeric"
                                      bg="white"
                                      borderColor="gray.200"
                                      _hover={{ borderColor: 'gray.300' }}
                                      _focus={{ borderColor: 'blue.400', boxShadow: '0 0 0 1px var(--chakra-colors-blue-400)' }}
                                    />
                                    <NumberInputStepper>
                                      <NumberIncrementStepper />
                                      <NumberDecrementStepper />
                                    </NumberInputStepper>
                                  </NumberInput>
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
                      <Box mt={4} p={4} bg="blue.50" borderRadius="md" borderLeft="4px solid" borderLeftColor="blue.400">
                        <Flex justify="space-between" align="center">
                          <Text fontSize="sm" fontWeight="medium" color="gray.700">
                            Total Items: {formData.items.length}
                          </Text>
                          <Text fontSize="lg" fontWeight="bold" color="blue.600">
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
          <ModalContent maxW="90vw" maxH="90vh">
            <ModalHeader bg="blue.50" borderRadius="md" mb={4}>
              <HStack>
                <Box w={1} h={6} bg="blue.500" borderRadius="full" />
                <Text fontSize="lg" fontWeight="bold" color="blue.700">
                  Create New Purchase
                </Text>
              </HStack>
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody overflowY="visible" px={6}>
              <VStack spacing={6} align="stretch">
                {/* Basic Information Section */}
                <Card>
                  <CardHeader pb={3}>
                    <Text fontSize="md" fontWeight="semibold" color="gray.700">
                      📋 Basic Information
                    </Text>
                  </CardHeader>
                  <CardBody pt={0}>
                    <SimpleGrid columns={3} spacing={4}>
                      <FormControl isRequired>
                        <FormLabel fontSize="sm" fontWeight="medium">Vendor</FormLabel>
                        {loadingVendors ? (
                          <Spinner size="sm" />
                        ) : (
                          <Select
                            placeholder="Select vendor"
                            value={formData.vendor_id}
                            onChange={(e) => setFormData({...formData, vendor_id: e.target.value})}
                            size="sm"
                          >
                            {vendors.map(vendor => (
                              <option key={vendor.id} value={vendor.id}>
                                {vendor.name} ({vendor.code})
                              </option>
                            ))}
                          </Select>
                        )}
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
                      <Text fontSize="md" fontWeight="semibold" color="gray.700">
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
                        <Thead bg="gray.50">
                          <Tr>
                            <Th fontSize="xs" fontWeight="semibold" color="gray.600">Product</Th>
                            <Th fontSize="xs" fontWeight="semibold" color="gray.600" isNumeric>Qty</Th>
                            <Th fontSize="xs" fontWeight="semibold" color="gray.600" isNumeric>Unit Price (IDR)</Th>
                            <Th fontSize="xs" fontWeight="semibold" color="gray.600">Expense Account</Th>
                            <Th fontSize="xs" fontWeight="semibold" color="gray.600" isNumeric>Total (IDR)</Th>
                            <Th fontSize="xs" fontWeight="semibold" color="gray.600" w="60px">Action</Th>
                          </Tr>
                        </Thead>
                        <Tbody>
                          {formData.items.length === 0 ? (
                            <Tr>
                              <Td colSpan={6} textAlign="center" py={8}>
                                <VStack spacing={2}>
                                  <Text fontSize="sm" color="gray.500">No items added yet</Text>
                                  <Text fontSize="xs" color="gray.400">Click "Add Item" button to start adding purchase items</Text>
                                </VStack>
                              </Td>
                            </Tr>
                          ) : (
                            formData.items.map((item, index) => (
                              <Tr key={index} _hover={{ bg: 'gray.50' }}>
                                <Td minW="200px">
                                  {loadingProducts ? (
                                    <Flex align="center" justify="center" h="32px">
                                      <Spinner size="sm" />
                                    </Flex>
                                  ) : (
                                    <Select
                                      placeholder="Select product"
                                      value={item.product_id}
                                      onChange={(e) => {
                                        const items = [...formData.items];
                                        items[index] = { ...items[index], product_id: e.target.value };
                                        setFormData({ ...formData, items });
                                      }}
                                      size="sm"
                                    >
                                      {products.map((p) => (
                                        <option key={p.id} value={p.id?.toString()}>
                                          {p?.id} - {p?.name || p?.code}
                                        </option>
                                      ))}
                                    </Select>
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
                                  <NumberInput 
                                    size="sm" 
                                    min={0} 
                                    precision={0} 
                                    value={item.unit_price} 
                                    onChange={(valueString) => {
                                      const items = [...formData.items];
                                      items[index] = { ...items[index], unit_price: valueString };
                                      setFormData({ ...formData, items });
                                    }} 
                                    maxW="160px"
                                    clampValueOnBlur={false}
                                  >
                                    <NumberInputField 
                                      textAlign="right" 
                                      fontSize="sm" 
                                      placeholder="Masukkan harga"
                                      inputMode="numeric"
                                      bg="white"
                                      borderColor="gray.200"
                                      _hover={{ borderColor: 'gray.300' }}
                                      _focus={{ borderColor: 'blue.400', boxShadow: '0 0 0 1px var(--chakra-colors-blue-400)' }}
                                    />
                                    <NumberInputStepper>
                                      <NumberIncrementStepper />
                                      <NumberDecrementStepper />
                                    </NumberInputStepper>
                                  </NumberInput>
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
                      <Box mt={4} p={4} bg="blue.50" borderRadius="md" borderLeft="4px solid" borderLeftColor="blue.400">
                        <Flex justify="space-between" align="center">
                          <Text fontSize="sm" fontWeight="medium" color="gray.700">
                            Total Items: {formData.items.length}
                          </Text>
                          <Text fontSize="lg" fontWeight="bold" color="blue.600">
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
                    
                    <FormControl>
                      <FormHelperText mt={3} fontSize="xs">
                        📌 Tambahkan minimal 1 item pembelian. Semua field harus diisi dengan benar.
                      </FormHelperText>
                    </FormControl>
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
      </VStack>
    </Layout>
  );
};

export default PurchasesPage;
