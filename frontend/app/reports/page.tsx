'use client';

import React, { useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import Layout from '@/components/layout/Layout';
import {
  Box,
  Heading,
  Text,
  SimpleGrid,
  Button,
  VStack,
  HStack,
  useToast,
  Card,
  CardBody,
  Icon,
  Flex,
  Badge,
  useColorModeValue
} from '@chakra-ui/react';
import { 
  FiFileText, 
  FiBarChart, 
  FiTrendingUp, 
  FiShoppingCart, 
  FiActivity
} from 'react-icons/fi';

// Define reports data matching the UI design
const availableReports = [
  {
    id: 'profit-loss',
    name: 'Profit and Loss Statement',
    description: 'Comprehensive profit and loss statements provides a detailed view of financial transactions on a specific period',
    type: 'FINANCIAL',
    icon: FiTrendingUp
  },
  {
    id: 'balance-sheet',
    name: 'Balance Sheet',
    description: 'Provides a company\'s assets, liabilities, and shareholders\' equity at a specific point in time',
    type: 'FINANCIAL', 
    icon: FiBarChart
  },
  {
    id: 'cash-flow',
    name: 'Cash Flow Statement',
    description: 'Measures how well a company generates cash to pay its debt obligations and fund its operating expenditures',
    type: 'FINANCIAL',
    icon: FiActivity
  },
  {
    id: 'sales-summary',
    name: 'Sales Summary Report',
    description: 'Provides a summary of sales transactions over a period',
    type: 'OPERATIONAL',
    icon: FiShoppingCart
  },
  {
    id: 'purchase-summary',
    name: 'Purchase Summary Report',
    description: 'Provides a summary of purchase transactions over a period',
    type: 'OPERATIONAL',
    icon: FiShoppingCart
  }
];

interface Report {
  id: string;
  name: string;
  description: string;
  type: string;
  icon: any;
}

const ReportsPage: React.FC = () => {
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const toast = useToast();

  const handleGenerateReport = (reportId: string) => {
    toast({
      title: 'Report Generation',
      description: `Generating ${reportId} report...`,
      status: 'info',
      duration: 3000,
      isClosable: true,
    });
  };

  return (
    <Layout allowedRoles={['admin', 'finance', 'director']}>
      <Box p={8}>
        <VStack spacing={8} align="stretch">
          <Heading as="h1" size="xl" color="gray.700" fontWeight="medium">
            Financial Reports
          </Heading>
          
          <SimpleGrid columns={[1, 2, 3]} spacing={6}>
            {availableReports.map((report) => (
              <Card
                key={report.id}
                bg="white"
                border="1px"
                borderColor="gray.200"
                borderRadius="md"
                overflow="hidden"
                _hover={{ shadow: 'md' }}
                transition="all 0.2s"
              >
                <CardBody p={0}>
                  <VStack spacing={0} align="stretch">
                    {/* Icon and Badge Header */}
                    <Flex p={4} align="center" justify="space-between">
                      <Icon as={report.icon} size="24px" color="blue.500" />
                      <Badge 
                        colorScheme="green" 
                        variant="solid"
                        fontSize="xs"
                        px={2}
                        py={1}
                        borderRadius="md"
                      >
                        {report.type}
                      </Badge>
                    </Flex>
                    
                    {/* Content */}
                    <VStack spacing={3} align="stretch" px={4} pb={4}>
                      <Heading size="md" color="gray.800" fontWeight="medium">
                        {report.name}
                      </Heading>
                      <Text 
                        fontSize="sm" 
                        color="gray.600" 
                        lineHeight="1.4"
                        noOfLines={3}
                      >
                        {report.description}
                      </Text>
                      
                      {/* Action Button */}
                      <Button
                        colorScheme="blue"
                        size="md"
                        width="full"
                        mt={2}
                        onClick={() => handleGenerateReport(report.id)}
                        isLoading={loading}
                      >
                        Generate Report
                      </Button>
                    </VStack>
                  </VStack>
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
