'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import Layout from '@/components/layout/Layout';
import {
  Box,
  VStack,
  HStack,
  Heading,
  Text,
  Card,
  CardBody,
  CardHeader,
  SimpleGrid,
  Icon,
  Badge,
  Button,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Spinner,
  useColorModeValue,
  Divider
} from '@chakra-ui/react';
import { FiHome, FiDollarSign, FiCalendar, FiSettings, FiEdit } from 'react-icons/fi';

interface SystemSettings {
  companyName: string;
  companyAddress: string;
  companyPhone: string;
  companyEmail: string;
  currency: string;
  dateFormat: string;
  fiscalYearStart: string;
}

const SettingsPage: React.FC = () => {
  const { user } = useAuth();
  const [settings, setSettings] = useState<SystemSettings | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Move useColorModeValue to top level to fix hooks order
  const blueColor = useColorModeValue('blue.500', 'blue.300');
  const greenColor = useColorModeValue('green.500', 'green.300');

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        // Data dummy untuk testing di frontend
        const data: SystemSettings = {
          companyName: 'PT. Sistem Akuntansi Indonesia',
          companyAddress: 'Jl. Sudirman Kav. 45-46, Jakarta Pusat 10210, Indonesia',
          companyPhone: '+62-21-5551234',
          companyEmail: 'info@sistemakuntansi.co.id',
          currency: 'IDR',
          dateFormat: 'DD/MM/YYYY',
          fiscalYearStart: 'January 1st',
        };
        
        setSettings(data);
      } catch (err: any) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchSettings();
  }, []);

  // Loading state - moved after all hooks
  if (loading) {
    return (
      <Layout allowedRoles={['ADMIN']}>
        <Box>
          <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
          <Text ml={4}>Loading settings...</Text>
        </Box>
      </Layout>
    );
  }

  return (
    <Layout allowedRoles={['ADMIN']}>
      <Box>
        <VStack spacing={6} alignItems="start">
          <HStack justify="space-between" width="full">
            <Heading as="h1" size="xl">System Settings</Heading>
            <Button
              colorScheme="blue"
              leftIcon={<FiEdit />}
              onClick={() => console.log('Edit settings')}
            >
              Edit Settings
            </Button>
          </HStack>
          
          {error && (
            <Alert status="error" width="full">
              <AlertIcon />
              <AlertTitle>Error:</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          
          <SimpleGrid columns={[1, 1, 2]} spacing={6} width="full">
            {/* Company Information Card */}
            <Card border="1px" borderColor="gray.200" boxShadow="md">
              <CardHeader>
                <HStack spacing={3}>
                  <Icon as={FiHome} boxSize={6} color={blueColor} />
                  <Heading size="md">Company Information</Heading>
                </HStack>
              </CardHeader>
              <CardBody>
                <VStack spacing={4} alignItems="start">
                  <Box>
                    <Text fontWeight="semibold" color="gray.600" fontSize="sm">Company Name</Text>
                    <Text fontSize="md">{settings?.companyName}</Text>
                  </Box>
                  <Divider />
                  <Box>
                    <Text fontWeight="semibold" color="gray.600" fontSize="sm">Address</Text>
                    <Text fontSize="md">{settings?.companyAddress}</Text>
                  </Box>
                  <Divider />
                  <Box>
                    <Text fontWeight="semibold" color="gray.600" fontSize="sm">Phone</Text>
                    <Text fontSize="md">{settings?.companyPhone}</Text>
                  </Box>
                  <Divider />
                  <Box>
                    <Text fontWeight="semibold" color="gray.600" fontSize="sm">Email</Text>
                    <Text fontSize="md">{settings?.companyEmail}</Text>
                  </Box>
                </VStack>
              </CardBody>
            </Card>

            {/* Financial Settings Card */}
            <Card border="1px" borderColor="gray.200" boxShadow="md">
              <CardHeader>
                <HStack spacing={3}>
                  <Icon as={FiDollarSign} boxSize={6} color={greenColor} />
                  <Heading size="md">System Configuration</Heading>
                </HStack>
              </CardHeader>
              <CardBody>
                <VStack spacing={4} alignItems="start">
                  <Box>
                    <Text fontWeight="semibold" color="gray.600" fontSize="sm">Date Format</Text>
                    <Text fontSize="md">{settings?.dateFormat}</Text>
                  </Box>
                  <Divider />
                  <Box>
                    <Text fontWeight="semibold" color="gray.600" fontSize="sm">Fiscal Year Start</Text>
                    <HStack>
                      <Icon as={FiCalendar} boxSize={4} color="gray.500" />
                      <Text fontSize="md">{settings?.fiscalYearStart}</Text>
                    </HStack>
                  </Box>
                </VStack>
              </CardBody>
            </Card>
          </SimpleGrid>
        </VStack>
      </Box>
    </Layout>
  );
};

export default SettingsPage;
