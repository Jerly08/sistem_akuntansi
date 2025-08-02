'use client';
import React from 'react';
import { Box, Flex, Heading, Text } from '@chakra-ui/react';

export const FinanceDashboard = () => (
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
  </Box>
);
