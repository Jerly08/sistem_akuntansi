'use client';
import React from 'react';
import { Box, Heading, Text } from '@chakra-ui/react';

export const EmployeeDashboard = () => (
  <Box>
    <Heading as="h2" size="xl" mb={6} color="gray.800">
      Dasbor Saya
    </Heading>
    
    <Box bg="white" p={6} borderRadius="lg" boxShadow="sm" mt={4}>
      <Heading as="h3" size="md" mb={4}>Pengumuman Perusahaan</Heading>
      <Text>
        Selamat datang di sistem akuntansi yang baru! Mohon untuk membiasakan diri dengan antarmuka yang tersedia.
      </Text>
    </Box>
  </Box>
);
