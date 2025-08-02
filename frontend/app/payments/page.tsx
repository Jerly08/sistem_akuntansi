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

interface Payment {
  id: string;
  paymentNumber: string;
  vendorName: string;
  date: string;
  amount: number;
  paymentMethod: string;
  reference: string;
  status: string;
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
  { header: 'Payment #', accessor: 'paymentNumber' as keyof Payment },
  { header: 'Vendor', accessor: 'vendorName' as keyof Payment },
  {
    header: 'Date',
    accessor: ((row: Payment) => {
      return new Date(row.date).toLocaleDateString();
    }) as (row: Payment) => React.ReactNode
  },
  {
    header: 'Amount',
    accessor: ((row: Payment) => {
      return `$${row.amount.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
    }) as (row: Payment) => React.ReactNode
  },
  {
    header: 'Method',
    accessor: 'paymentMethod' as keyof Payment
  },
  {
    header: 'Status',
    accessor: ((row: Payment) => (
      <Badge colorScheme={getStatusColor(row.status)} variant="subtle">
        {row.status}
      </Badge>
    )) as (row: Payment) => React.ReactNode
  },
];

const PaymentsPage: React.FC = () => {
  const { token } = useAuth();
  const [payments, setPayments] = useState<Payment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

useEffect(() => {
 const fetchPayments = async () => {
 try {
 // Data dummy untuk testing di frontend
 const data: Payment[] = [
 {
 id: '1',
 paymentNumber: 'PAY-2025-001',
 vendorName: 'CV Sumber Rejeki',
 date: '2025-07-20',
 amount: 5000000,
 paymentMethod: 'Bank Transfer',
 reference: 'TRF20250720001',
 status: 'Completed',
 },
 {
 id: '2',
 paymentNumber: 'PAY-2025-002',
 vendorName: 'PT Maju Jaya',
 date: '2025-07-18',
 amount: 7500000,
 paymentMethod: 'Cash',
 reference: 'CSH20250718002',
 status: 'Pending',
 },
 {
 id: '3',
 paymentNumber: 'PAY-2025-003',
 vendorName: 'PT Global Tech',
 date: '2025-07-19',
 amount: 10500000,
 paymentMethod: 'Credit Card',
 reference: 'CRD20250719003',
 status: 'Completed',
 },
         {
           id: '4',
           paymentNumber: 'PAY-2025-004',
           vendorName: 'Toko Elektronik Sejati',
           date: '2025-07-21',
           amount: 12000000,
           paymentMethod: 'Bank Transfer',
           reference: 'TRF20250721004',
           status: 'In Transit',
         },
         {
           id: '5',
           paymentNumber: 'PAY-2025-005',
           vendorName: 'PT Indah Karya',
           date: '2025-07-22',
           amount: 8750000,
           paymentMethod: 'Cash',
           reference: 'CSH20250722005',
           status: 'Completed',
         },
         {
           id: '6',
           paymentNumber: 'PAY-2025-006',
           vendorName: 'CV Berkah Jaya',
           date: '2025-07-23',
           amount: 6200000,
           paymentMethod: 'Bank Transfer',
           reference: 'TRF20250723006',
           status: 'Pending',
         },
         {
           id: '7',
           paymentNumber: 'PAY-2025-007',
           vendorName: 'PT Solusi Digital',
           date: '2025-07-24',
           amount: 15500000,
           paymentMethod: 'Credit Card',
           reference: 'CRD20250724007',
           status: 'Completed',
         },
         {
           id: '8',
           paymentNumber: 'PAY-2025-008',
           vendorName: 'CV Sumber Rejeki',
           date: '2025-07-25',
           amount: 4300000,
           paymentMethod: 'Bank Transfer',
           reference: 'TRF20250725008',
           status: 'Cancelled',
         },
 ];

 setPayments(data);
 } catch (err: any) {
 setError(err.message || 'An error occurred while fetching payments');
 } finally {
 setLoading(false);
 }
 };

 if (token) {
 fetchPayments();
 }
}, [token]);

  // Action buttons for each row
  const renderActions = (payment: Payment) => (
    <Flex gap={2}>
      <Button
        size="sm"
        variant="outline"
        leftIcon={<FiEye />}
        onClick={() => console.log('View payment:', payment.id)}
      >
        View
      </Button>
      <Button
        size="sm"
        variant="outline"
        leftIcon={<FiEdit />}
        onClick={() => console.log('Edit payment:', payment.id)}
      >
        Edit
      </Button>
      <Button
        size="sm"
        colorScheme="red"
        variant="outline"
        leftIcon={<FiTrash2 />}
        onClick={() => console.log('Delete payment:', payment.id)}
      >
        Delete
      </Button>
    </Flex>
  );

  if (loading) {
    return (
      <Layout allowedRoles={['ADMIN', 'FINANCE', 'INVENTORY_MANAGER']}>
        <Box>
          <Text>Loading payments...</Text>
        </Box>
      </Layout>
    );
  }

  return (
    <Layout allowedRoles={['ADMIN', 'FINANCE', 'INVENTORY_MANAGER']}>
      <Box>
        <Flex justify="space-between" align="center" mb={6}>
          <Heading size="lg">Payments</Heading>
          <Button
            colorScheme="brand"
            leftIcon={<FiPlus />}
            onClick={() => console.log('Create new payment')}
          >
            New Payment
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
          <DataTable<Payment>
            columns={columns}
            data={payments}
            keyField="id"
            title="All Payments"
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

export default PaymentsPage;
