'use client';

import React, { useState } from 'react';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { DataTable } from '@/components/common/DataTable';
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
  CardBody
} from '@chakra-ui/react';
import { FiPlus, FiSearch } from 'react-icons/fi';

// Dummy data for sales transactions
const dummySales = [
  {
    id: 'SALE-001',
    invoiceNumber: 'INV-2024-001',
    customerName: 'Client A',
    date: '2024-01-30',
    total: 5500,
    status: 'paid'
  },
  {
    id: 'SALE-002',
    invoiceNumber: 'INV-2024-002',
    customerName: 'Client B',
    date: '2024-01-29',
    total: 8900,
    status: 'pending'
  },
  {
    id: 'SALE-003',
    invoiceNumber: 'INV-2024-003',
    customerName: 'Client C',
    date: '2024-01-28',
    total: 3200,
    status: 'overdue'
  },
  {
    id: 'SALE-004',
    invoiceNumber: 'INV-2024-004',
    customerName: 'Client A',
    date: '2024-01-27',
    total: 12500,
    status: 'paid'
  },
  {
    id: 'SALE-005',
    invoiceNumber: 'INV-2024-005',
    customerName: 'Client D',
    date: '2024-01-26',
    total: 7500,
    status: 'draft'
  }
];

const columns = [
  { header: 'Invoice #', accessor: 'invoiceNumber' },
  { header: 'Customer', accessor: 'customerName' },
  { header: 'Date', accessor: 'date' },
  { header: 'Total', accessor: 'total' },
  { header: 'Status', accessor: 'status' },
];

const SalesPage: React.FC = () => {
  const { user } = useAuth();
  const canCreate = user?.role === 'ADMIN' || user?.role === 'FINANCE' || user?.role === 'DIRECTOR';
  const [sales, setSales] = useState(dummySales);

  return (
<Layout allowedRoles={['admin', 'finance', 'director', 'employee', 'inventory_manager']}>
      <Box>
        {/* Header */}
        <Flex justify="space-between" align="center" mb={6}>
          <Box>
            <Heading as="h1" size="xl" mb={2}>Sales</Heading>
            <Text color="gray.600">Manage your sales transactions</Text>
          </Box>
          {canCreate && (
            <Button leftIcon={<FiPlus />} colorScheme="brand" size="lg">
              Create Invoice
            </Button>
          )}
        </Flex>

        {/* Search and Filters */}
        <Card mb={6}>
          <CardBody>
            <HStack spacing={4}>
              <InputGroup maxW="400px">
                <InputLeftElement pointerEvents="none">
                  <FiSearch color="gray.300" />
                </InputLeftElement>
                <Input placeholder="Search by invoice # or customer..." />
              </InputGroup>
            </HStack>
          </CardBody>
        </Card>

        {/* Sales Table */}
        <Card>
          <CardHeader>
            <Heading size="md">Sales Transactions ({sales.length})</Heading>
          </CardHeader>
          <CardBody>
            <DataTable columns={columns} data={sales} keyField="id" />
          </CardBody>
        </Card>
      </Box>
    </Layout>
  );
};

export default SalesPage;

