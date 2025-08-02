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
} from '@chakra-ui/react';
import { FiPlus, FiEye, FiEdit, FiTrash2 } from 'react-icons/fi';

interface PurchaseTransaction {
  id: string;
  purchaseNumber: string;
  vendorName: string;
  date: string;
  total: number;
  status: string;
  description?: string;
}

// Status color mapping
const getStatusColor = (status: string) => {
  switch (status.toLowerCase()) {
    case 'completed':
      return 'green';
    case 'pending':
      return 'yellow';
    case 'in transit':
      return 'blue';
    case 'cancelled':
      return 'red';
    default:
      return 'gray';
  }
};

const columns = [
  { header: 'Purchase #', accessor: 'purchaseNumber' as keyof PurchaseTransaction },
  { header: 'Vendor', accessor: 'vendorName' as keyof PurchaseTransaction },
  { 
    header: 'Date', 
    accessor: ((row: PurchaseTransaction) => {
      return new Date(row.date).toLocaleDateString();
    }) as (row: PurchaseTransaction) => React.ReactNode
  },
  { 
    header: 'Total', 
    accessor: ((row: PurchaseTransaction) => {
      return `$${row.total.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
    }) as (row: PurchaseTransaction) => React.ReactNode
  },
  { 
    header: 'Status', 
    accessor: ((row: PurchaseTransaction) => (
      <Badge colorScheme={getStatusColor(row.status)} variant="subtle">
        {row.status}
      </Badge>
    )) as (row: PurchaseTransaction) => React.ReactNode
  },
];

const PurchasesPage: React.FC = () => {
  const { token } = useAuth();
  const [purchases, setPurchases] = useState<PurchaseTransaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

useEffect(() => {
 const fetchPurchases = async () => {
 try {
  // Data dummy untuk testing di frontend
  const data: PurchaseTransaction[] = [
    {
      id: '1',
      purchaseNumber: 'PO-2025-001',
      vendorName: 'CV Sumber Rejeki',
      date: '2025-07-15',
      total: 15000000,
      status: 'Approved',
      description: 'Purchase of office supplies'
    },
    {
      id: '2',
      purchaseNumber: 'PO-2025-002',
      vendorName: 'PT Maju Jaya',
      date: '2025-07-16',
      total: 23500000,
      status: 'Pending',
      description: 'Purchase of electronic equipment'
    },
    {
      id: '3',
      purchaseNumber: 'PO-2025-003',
      vendorName: 'Toko Elektronik Sejati',
      date: '2025-07-17',
      total: 4700000,
      status: 'Completed',
      description: 'Purchase of cables and accessories'
    },
    {
      id: '4',
      purchaseNumber: 'PO-2025-004',
      vendorName: 'PT Global Tech',
      date: '2025-07-18',
      total: 7300000,
      status: 'In Transit',
      description: 'Purchase of network equipment'
    },
    {
      id: '5',
      purchaseNumber: 'PO-2025-005',
      vendorName: 'PT Indah Karya',
      date: '2025-07-19',
      total: 18200000,
      status: 'Approved',
      description: 'Purchase of furniture and fixtures'
    },
    {
      id: '6',
      purchaseNumber: 'PO-2025-006',
      vendorName: 'CV Berkah Jaya',
      date: '2025-07-20',
      total: 3450000,
      status: 'Pending',
      description: 'Purchase of cleaning supplies'
    },
    {
      id: '7',
      purchaseNumber: 'PO-2025-007',
      vendorName: 'PT Solusi Digital',
      date: '2025-07-21',
      total: 12800000,
      status: 'Completed',
      description: 'Purchase of software licenses'
    },
    {
      id: '8',
      purchaseNumber: 'PO-2025-008',
      vendorName: 'CV Sumber Rejeki',
      date: '2025-07-22',
      total: 5600000,
      status: 'Cancelled',
      description: 'Purchase of raw materials'
    },
  ];

  setPurchases(data);
 } catch (err: any) {
 setError(err.message || 'An error occurred while fetching purchases');
 } finally {
 setLoading(false);
 }
 };

 if (token) {
 fetchPurchases();
 }
}, [token]);

  // Action buttons for each row
  const renderActions = (purchase: PurchaseTransaction) => (
    <Flex gap={2}>
      <Button
        size="sm"
        variant="outline"
        leftIcon={<FiEye />}
        onClick={() => console.log('View purchase:', purchase.id)}
      >
        View
      </Button>
      <Button
        size="sm"
        variant="outline"
        leftIcon={<FiEdit />}
        onClick={() => console.log('Edit purchase:', purchase.id)}
      >
        Edit
      </Button>
      <Button
        size="sm"
        colorScheme="red"
        variant="outline"
        leftIcon={<FiTrash2 />}
        onClick={() => console.log('Delete purchase:', purchase.id)}
      >
        Delete
      </Button>
    </Flex>
  );

  if (loading) {
    return (
      <Layout allowedRoles={['ADMIN', 'FINANCE', 'INVENTORY_MANAGER']}>
        <Box>
          <Text>Loading purchases...</Text>
        </Box>
      </Layout>
    );
  }

  return (
    <Layout allowedRoles={['ADMIN', 'FINANCE', 'INVENTORY_MANAGER']}>
      <Box>
        <Flex justify="space-between" align="center" mb={6}>
          <Heading size="lg">Purchase Transactions</Heading>
          <Button
            colorScheme="brand"
            leftIcon={<FiPlus />}
            onClick={() => console.log('Create new purchase')}
          >
            New Purchase
          </Button>
        </Flex>
        
        {error && (
          <Alert status="error" mb={4}>
            <AlertIcon />
            <AlertTitle mr={2}>Error!</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        
        <Box bg="white" borderRadius="lg" overflow="hidden" boxShadow="sm">
          <DataTable<PurchaseTransaction>
            columns={columns}
            data={purchases}
            keyField="id"
            title="All Purchases"
            actions={renderActions}
            searchable={true}
            pagination={true}
            pageSize={10}
          />
        </Box>
      </Box>
    </Layout>
  );
};

export default PurchasesPage;
