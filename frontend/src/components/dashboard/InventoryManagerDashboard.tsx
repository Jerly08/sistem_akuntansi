'use client';
import React from 'react';
import { useRouter } from 'next/navigation';
import { 
  Box, 
  Flex, 
  Heading, 
  Text, 
  Card,
  CardHeader,
  CardBody,
  Button,
  HStack,
  Icon
} from '@chakra-ui/react';
import {
  FiPackage,
  FiPlus,
  FiBarChart,
} from 'react-icons/fi';

export const InventoryManagerDashboard = () => {
  const router = useRouter();
  
  return (
    <Box>
      <Heading as="h2" size="xl" mb={6} color="gray.800">
        Dasbor Inventaris
      </Heading>
    
    <Flex gap={4} flexWrap="wrap" mt={4}>
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Heading as="h3" size="sm" mb={2}>Nilai Total Inventaris</Heading>
        <Text fontSize="2xl" fontWeight="bold">Rp 2.253.450.750</Text>
      </Box>
      
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Heading as="h3" size="sm" mb={2}>Stok Menipis</Heading>
        <Text fontSize="2xl" fontWeight="bold">5 item</Text>
      </Box>
      
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Heading as="h3" size="sm" mb={2}>Perputaran Stok</Heading>
        <Text fontSize="2xl" fontWeight="bold">4.5</Text>
      </Box>
    </Flex>

    {/* Quick Access Section - Inventory Manager has no access to sales/purchases/payments */}
    <Card mt={6}>
      <CardHeader>
        <Heading size="md" display="flex" alignItems="center">
          <Icon as={FiPlus} mr={2} color="blue.500" />
          Akses Cepat
        </Heading>
      </CardHeader>
      <CardBody>
        <HStack spacing={4} flexWrap="wrap">
          <Button
            leftIcon={<FiPackage />}
            colorScheme="purple"
            variant="outline"
            onClick={() => router.push('/products')}
            size="md"
          >
            Kelola Produk
          </Button>
          <Button
            leftIcon={<FiBarChart />}
            colorScheme="teal"
            variant="outline"
            onClick={() => router.push('/reports')}
            size="md"
          >
            Laporan Inventaris
          </Button>
        </HStack>
      </CardBody>
    </Card>
  </Box>
  );
};
