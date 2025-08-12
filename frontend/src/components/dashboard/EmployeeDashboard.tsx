'use client';
import React from 'react';
import { useRouter } from 'next/navigation';
import { 
  Box, 
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
  FiUser,
  FiPlus,
  FiFileText,
} from 'react-icons/fi';

export const EmployeeDashboard = () => {
  const router = useRouter();
  
  return (
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

    {/* Quick Access Section - Employee has no access to sales/purchases/payments */}
    <Card mt={6}>
      <CardHeader>
        <Heading size="md" display="flex" alignItems="center">
          <Icon as={FiPlus} mr={2} color="blue.500" />
          Akses Cepat
        </Heading>
      </CardHeader>
      <CardBody>
        <Text mb={4} color="gray.600">
          Sebagai karyawan, Anda dapat mengakses profil dan melihat laporan yang tersedia.
        </Text>
        <HStack spacing={4} flexWrap="wrap">
          <Button
            leftIcon={<FiUser />}
            colorScheme="blue"
            variant="outline"
            onClick={() => router.push('/profile')}
            size="md"
          >
            Profil Saya
          </Button>
          <Button
            leftIcon={<FiFileText />}
            colorScheme="gray"
            variant="outline"
            onClick={() => router.push('/reports')}
            size="md"
          >
            Lihat Laporan
          </Button>
        </HStack>
      </CardBody>
    </Card>
  </Box>
  );
};
