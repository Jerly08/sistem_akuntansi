'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import Layout from '@/components/layout/Layout';
import {
  Box,
  Container,
  Heading,
  Text,
  SimpleGrid,
  Card,
  CardBody,
  Button,
  Badge,
  VStack,
  HStack,
  Icon,
  useColorModeValue,
  Spinner,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Flex,
  useToast
} from '@chakra-ui/react';
import { FiFileText, FiBarChart, FiTrendingUp, FiShoppingCart, FiDollarSign } from 'react-icons/fi';

interface Report {
  id: string;
  name: string;
  description: string;
  type: string;
}

const reportsList: Report[] = [
  {
    id: 'profit-loss',
    name: 'Profit and Loss Statement',
    description: 'Summarizes revenues, costs, and expenses incurred during a specific period.',
    type: 'Financial',
  },
  {
    id: 'balance-sheet',
    name: 'Balance Sheet',
    description: 'Reports a company\'s assets, liabilities, and shareholder equity at a specific point in time.',
    type: 'Financial',
  },
  {
    id: 'cash-flow',
    name: 'Cash Flow Statement',
    description: 'Measures how well a company manages its cash position.',
    type: 'Financial',
  },
  {
    id: 'sales-summary',
    name: 'Sales Summary Report',
    description: 'Provides a summary of sales transactions over a period.',
    type: 'Operational',
  },
  {
    id: 'purchase-summary',
    name: 'Purchase Summary Report',
    description: 'Provides a summary of purchase transactions over a period.',
    type: 'Operational',
  },
];

const ReportsPage: React.FC = () => {
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleGenerateReport = async (reportId: string) => {
    try {
      setLoading(true);
      const response = await fetch(`http://localhost:8080/api/v1/reports/${reportId}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to generate report');
      }

      const reportData = await response.json();
      // For now, just show an alert with the report title
      alert(`Generated report: ${reportData.title}`);
      console.log('Report data:', reportData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const icons = [FiDollarSign, FiBarChart, FiTrendingUp, FiFileText, FiShoppingCart];
  const toast = useToast();

  if (loading) {
    return (
      <Layout allowedRoles={['ADMIN', 'FINANCE', 'DIRECTOR']}>
        <Box>
          <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
          <Text ml={4}>Loading reports...</Text>
        </Box>
      </Layout>
    );
  }

  return (
    <Layout allowedRoles={['ADMIN', 'FINANCE', 'DIRECTOR']}>
      <Box>
        <VStack spacing={6} alignItems="start">
          <Heading as="h1" size="xl">Financial Reports</Heading>
          
          {error && (
            <Alert status="error">
              <AlertIcon />
              <AlertTitle>Error:</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          
          <SimpleGrid columns={[1, 2, 3]} spacing={8} width="full">
            {reportsList.map((report, index) => (
              <Card key={report.id} border="1px" borderColor="gray.200" boxShadow="md">
                <CardBody>
                  <HStack spacing={4} mb={4}>
                    <Icon as={icons[index % icons.length]} boxSize={6} color={useColorModeValue('blue.500', 'blue.300')} />
                    <Heading size="md">{report.name}</Heading>
                    <Badge colorScheme="teal">{report.type}</Badge>
                  </HStack>
                  <Text mb={4}>{report.description}</Text>
                  <Button
                    onClick={() => {
                      handleGenerateReport(report.id);
                      toast({
                        title: 'Generating Report',
                        description: `The ${report.name} is being generated.`,
                        status: 'info',
                        duration: 3000,
                        isClosable: true,
                      });
                    }}
                    colorScheme="blue"
                    width="full"
                    isLoading={loading}
                  >
                    Generate Report
                  </Button>
                </CardBody>
              </Card>
            ))}
          </SimpleGrid>
        </VStack>
      </Box>
    </Layout>
  );
};

export default ReportsPage;
