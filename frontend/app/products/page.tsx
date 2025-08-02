'use client';

import React, { useState } from 'react';
import Layout from '@/components/layout/Layout';
import {
  Box,
  Heading,
  Text,
  Button,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Badge,
  Card,
  CardHeader,
  CardBody,
  Flex,
  Input,
  InputGroup,
  InputLeftElement,
  HStack,
  useColorModeValue,
} from '@chakra-ui/react';
import { FiPlus, FiSearch, FiEdit, FiTrash2 } from 'react-icons/fi';

// Dummy data for products
const dummyProducts = [
  {
    id: 'PRD-001',
    name: 'Laptop Dell XPS 13',
    category: 'Electronics',
    stock: 25,
    price: 999.99,
    cost: 750.00,
    status: 'active',
    supplier: 'Dell Inc.',
    lastUpdated: '2024-01-30'
  },
  {
    id: 'PRD-002',
    name: 'Office Chair Ergonomic',
    category: 'Furniture',
    stock: 50,
    price: 299.99,
    cost: 180.00,
    status: 'active',
    supplier: 'Herman Miller',
    lastUpdated: '2024-01-29'
  },
  {
    id: 'PRD-003', 
    name: 'Wireless Mouse Logitech',
    category: 'Electronics',
    stock: 0,
    price: 49.99,
    cost: 25.00,
    status: 'out_of_stock',
    supplier: 'Logitech',
    lastUpdated: '2024-01-28'
  },
  {
    id: 'PRD-004',
    name: 'A4 Copy Paper (500 sheets)',
    category: 'Office Supplies',
    stock: 120,
    price: 12.99,
    cost: 8.50,
    status: 'active',
    supplier: 'Staples',
    lastUpdated: '2024-01-30'
  },
  {
    id: 'PRD-005',
    name: 'Standing Desk Converter',
    category: 'Furniture',
    stock: 8,
    price: 199.99,
    cost: 120.00,
    status: 'low_stock',
    supplier: 'Varidesk',
    lastUpdated: '2024-01-27'
  }
];

const getStatusColor = (status: string) => {
  switch (status) {
    case 'active': return 'green';
    case 'low_stock': return 'yellow';
    case 'out_of_stock': return 'red';
    default: return 'gray';
  }
};

const getStatusText = (status: string) => {
  switch (status) {
    case 'active': return 'Active';
    case 'low_stock': return 'Low Stock';
    case 'out_of_stock': return 'Out of Stock';
    default: return status;
  }
};

export default function ProductsPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [products] = useState(dummyProducts);
  
  const filteredProducts = products.filter(product =>
    product.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    product.category.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <Layout allowedRoles={['ADMIN', 'INVENTORY_MANAGER']}>
      <Box>
        {/* Header */}
        <Flex justify="space-between" align="center" mb={6}>
          <Box>
            <Heading as="h1" size="xl" mb={2}>Product Master</Heading>
            <Text color="gray.600">Manage your products and stock inventory</Text>
          </Box>
          <Button leftIcon={<FiPlus />} colorScheme="brand" size="lg">
            Add Product
          </Button>
        </Flex>

        {/* Search and Filters */}
        <Card mb={6}>
          <CardBody>
            <HStack spacing={4}>
              <InputGroup maxW="400px">
                <InputLeftElement pointerEvents="none">
                  <FiSearch color="gray.300" />
                </InputLeftElement>
                <Input
                  placeholder="Search products..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </InputGroup>
            </HStack>
          </CardBody>
        </Card>

        {/* Products Table */}
        <Card>
          <CardHeader>
            <Flex justify="space-between" align="center">
              <Heading size="md">Products ({filteredProducts.length})</Heading>
            </Flex>
          </CardHeader>
          <CardBody>
            <Table variant="simple">
              <Thead>
                <Tr>
                  <Th>Product ID</Th>
                  <Th>Name</Th>
                  <Th>Category</Th>
                  <Th>Stock</Th>
                  <Th>Price</Th>
                  <Th>Cost</Th>
                  <Th>Status</Th>
                  <Th>Actions</Th>
                </Tr>
              </Thead>
              <Tbody>
                {filteredProducts.map((product) => (
                  <Tr key={product.id}>
                    <Td fontWeight="medium">{product.id}</Td>
                    <Td>
                      <Box>
                        <Text fontWeight="medium">{product.name}</Text>
                        <Text fontSize="sm" color="gray.500">{product.supplier}</Text>
                      </Box>
                    </Td>
                    <Td>{product.category}</Td>
                    <Td>
                      <Text color={product.stock <= 10 ? 'red.500' : 'inherit'}>
                        {product.stock} units
                      </Text>
                    </Td>
                    <Td fontWeight="semibold">${product.price}</Td>
                    <Td>${product.cost}</Td>
                    <Td>
                      <Badge colorScheme={getStatusColor(product.status)} variant="subtle">
                        {getStatusText(product.status)}
                      </Badge>
                    </Td>
                    <Td>
                      <HStack spacing={2}>
                        <Button size="sm" variant="ghost" leftIcon={<FiEdit />}> 
                          Edit
                        </Button>
                        <Button size="sm" variant="ghost" colorScheme="red" leftIcon={<FiTrash2 />}>
                          Delete
                        </Button>
                      </HStack>
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
          </CardBody>
        </Card>
      </Box>
    </Layout>
  );
}
