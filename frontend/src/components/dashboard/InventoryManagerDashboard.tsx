'use client';
import React from 'react';
import { Box, Flex, Heading, Text } from '@chakra-ui/react';

export const InventoryManagerDashboard = () => (
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
  </Box>
);
