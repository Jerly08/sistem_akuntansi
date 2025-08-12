'use client';
import React from 'react';
import { useRouter } from 'next/navigation';
import { 
  Box, 
  Flex, 
  Heading, 
  Text, 
  Button, 
  Card,
  CardHeader,
  CardBody,
  HStack,
  Icon
} from '@chakra-ui/react';
import {
  FiDollarSign,
  FiShoppingCart,
  FiTrendingUp,
  FiPlus,
} from 'react-icons/fi';

export const DirectorDashboard = () => {
  const router = useRouter();
  
  return (
    <Box>
      <Heading as="h2" size="xl" mb={6} color="gray.800">
        Dasbor Direktur
      </Heading>
    
    <Box 
      p={4} 
      mb={4} 
      bg="orange.100" 
      borderLeft="4px" 
      borderColor="orange.500" 
      borderRadius="md"
      color="orange.800"
    >
      <Text fontWeight="bold" mb={2}>Butuh Persetujuan</Text>
      <Text mb={3}>
        Terdapat <Text as="strong">3</Text> transaksi penjualan bernilai besar menunggu persetujuan Anda.
      </Text>
      <Button colorScheme="orange" size="sm">
        Lihat Persetujuan
      </Button>
    </Box>
    
    <Flex gap={4} flexWrap="wrap" mt={4}>
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Heading as="h3" size="sm" mb={2}>Margin Laba Kotor</Heading>
        <Text fontSize="2xl" fontWeight="bold">45.2%</Text>
      </Box>
      
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Heading as="h3" size="sm" mb={2}>Margin Laba Bersih</Heading>
        <Text fontSize="2xl" fontWeight="bold">15.8%</Text>
      </Box>
      
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Heading as="h3" size="sm" mb={2}>Pertumbuhan Pendapatan</Heading>
        <Text fontSize="2xl" fontWeight="bold">+12% YoY</Text>
      </Box>
      
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Heading as="h3" size="sm" mb={2}>Kesehatan Arus Kas</Heading>
        <Text fontSize="2xl" fontWeight="bold" color="green.600">Sehat</Text>
      </Box>
    </Flex>

    {/* Quick Access Section - Director has access to all modules */}
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
