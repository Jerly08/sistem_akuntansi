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
  FiDollarSign,
  FiShoppingCart,
  FiTrendingUp,
  FiPlus,
} from 'react-icons/fi';

export const FinanceDashboard = () => {
  const router = useRouter();
  
  return (
    <Box>
      <Heading as="h2" size="xl" mb={6} color="gray.800">
        Dasbor Keuangan
      </Heading>
    
    <Flex gap={4} flexWrap="wrap" mt={4}>
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Heading as="h3" size="sm" mb={2}>Invoice Perlu Dibayar</Heading>
        <Text fontSize="2xl" fontWeight="bold">12</Text>
      </Box>
      
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Heading as="h3" size="sm" mb={2}>Invoice Belum Lunas</Heading>
        <Text fontSize="2xl" fontWeight="bold">8</Text>
      </Box>
      
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Heading as="h3" size="sm" mb={2}>Jurnal Perlu di-Posting</Heading>
        <Text fontSize="2xl" fontWeight="bold">4</Text>
      </Box>
      
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Heading as="h3" size="sm" mb={2}>Rekonsiliasi Bank</Heading>
        <Text>Rekonsiliasi terakhir: 2 hari lalu</Text>
      </Box>
    </Flex>

    {/* Quick Access Section */}
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
            leftIcon={<FiDollarSign />}
            colorScheme="green"
            variant="outline"
            onClick={() => router.push('/sales')}
            size="md"
          >
            Tambah Penjualan
          </Button>
          <Button
            leftIcon={<FiShoppingCart />}
            colorScheme="orange"
            variant="outline"
            onClick={() => router.push('/purchases')}
            size="md"
          >
            Tambah Pembelian
          </Button>
          <Button
            leftIcon={<FiTrendingUp />}
            colorScheme="blue"
            variant="outline"
            onClick={() => router.push('/payments')}
            size="md"
          >
            Kelola Pembayaran
          </Button>
        </HStack>
      </CardBody>
    </Card>
  </Box>
  );
};
