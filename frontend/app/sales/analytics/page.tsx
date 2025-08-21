'use client';

import React, { useState, useEffect } from 'react';
import Layout from '@/components/layout/Layout';
import SalesChart from '@/components/sales/SalesChart';
import ReceivablesTable from '@/components/sales/ReceivablesTable';
import salesService, { SalesAnalytics, ReceivablesReport, SalesSummary } from '@/services/salesService';
import {
  Box,
  Heading,
  Text,
  VStack,
  HStack,
  Card,
  CardHeader,
  CardBody,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  SimpleGrid,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  Button,
  Select,
  Input,
  Flex,
  Spinner,
  Alert,
  AlertIcon,
  useToast
} from '@chakra-ui/react';
import { FiRefreshCw, FiDownload, FiTrendingUp, FiDollarSign, FiFileText, FiUsers } from 'react-icons/fi';

const SalesAnalyticsPage: React.FC = () => {
  const toast = useToast();
  
  const [analytics, setAnalytics] = useState<SalesAnalytics | null>(null);
  const [receivables, setReceivables] = useState<ReceivablesReport | null>(null);
  const [summary, setSummary] = useState<SalesSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Analytics filters
  const [period, setPeriod] = useState('monthly');
  const [year, setYear] = useState('2024');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');

  // Load all analytics data
  const loadAnalyticsData = async () => {
    try {
      setLoading(true);
      setError(null);

      const [analyticsData, receivablesData, summaryData] = await Promise.all([
        salesService.getSalesAnalytics(period, year),
        salesService.getReceivablesReport(),
        salesService.getSalesSummary(startDate || undefined, endDate || undefined)
      ]);

      setAnalytics(analyticsData);
      setReceivables(receivablesData);
      setSummary(summaryData);
    } catch (error: any) {
      setError(error.response?.data?.message || 'Failed to load analytics data');
      toast({
        title: 'Error loading analytics',
        description: error.response?.data?.message || 'Failed to load analytics data',
        status: 'error',
        duration: 3000
      });
    } finally {
      setLoading(false);
    }
  };

  // Initial load
  useEffect(() => {
    loadAnalyticsData();
  }, [period, year]);

  // Refresh with date filters
  const handleDateFilterRefresh = () => {
    loadAnalyticsData();
  };

  // Export reports
  const handleExportReport = async () => {
    try {
      await salesService.downloadSalesReportPDF(
        startDate || undefined, 
        endDate || undefined
      );
      toast({
        title: 'Report downloaded',
        description: 'Sales report has been downloaded successfully',
        status: 'success',
        duration: 3000
      });
    } catch (error: any) {
      toast({
        title: 'Export failed',
        description: error.response?.data?.message || 'Failed to export report',
        status: 'error',
        duration: 3000
      });
    }
  };

  if (loading) {
    return (
      <Layout allowedRoles={['admin', 'finance', 'director']}>
        <Flex justify="center" align="center" minH="400px">
          <Spinner size="xl" />
        </Flex>
      </Layout>
    );
  }

  if (error) {
    return (
      <Layout allowedRoles={['admin', 'finance', 'director']}>
        <Alert status="error">
          <AlertIcon />
          {error}
        </Alert>
      </Layout>
    );
  }

  return (
    <Layout allowedRoles={['admin', 'finance', 'director']}>
      <VStack spacing={6} align="stretch">
        {/* Header */}
        <Flex justify="space-between" align="center">
          <Box>
            <Heading as="h1" size="xl" mb={2}>Sales Analytics</Heading>
            <Text color="gray.600">Comprehensive sales performance analysis and reporting</Text>
          </Box>
          <HStack spacing={3}>
            <Button
              leftIcon={<FiRefreshCw />}
              variant="ghost"
              onClick={loadAnalyticsData}
              isLoading={loading}
            >
              Refresh
            </Button>
            <Button
              leftIcon={<FiDownload />}
              colorScheme="blue"
              onClick={handleExportReport}
            >
              Export Report
            </Button>
          </HStack>
        </Flex>

        {/* Filters */}
        <Card>
          <CardBody>
            <VStack spacing={4} align="stretch">
              <Text fontWeight="medium">Analytics Filters</Text>
              <HStack spacing={4} wrap="wrap">
                <Box>
                  <Text fontSize="sm" mb={2}>Period</Text>
                  <Select value={period} onChange={(e) => setPeriod(e.target.value)} maxW="150px">
                    <option value="monthly">Monthly</option>
                    <option value="quarterly">Quarterly</option>
                    <option value="yearly">Yearly</option>
                  </Select>
                </Box>
                <Box>
                  <Text fontSize="sm" mb={2}>Year</Text>
                  <Select value={year} onChange={(e) => setYear(e.target.value)} maxW="120px">
                    <option value="2024">2024</option>
                    <option value="2023">2023</option>
                    <option value="2022">2022</option>
                  </Select>
                </Box>
                <Box>
                  <Text fontSize="sm" mb={2}>Start Date (for summary)</Text>
                  <Input
                    type="date"
                    value={startDate}
                    onChange={(e) => setStartDate(e.target.value)}
                    maxW="200px"
                  />
                </Box>
                <Box>
                  <Text fontSize="sm" mb={2}>End Date (for summary)</Text>
                  <Input
                    type="date"
                    value={endDate}
                    onChange={(e) => setEndDate(e.target.value)}
                    maxW="200px"
                  />
                </Box>
                <Box alignSelf="end">
                  <Button onClick={handleDateFilterRefresh} colorScheme="blue" size="sm">
                    Apply Filters
                  </Button>
                </Box>
              </HStack>
            </VStack>
          </CardBody>
        </Card>

        {/* Summary Cards */}
        {summary && (
          <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>
                    <HStack>
                      <Box p={2} bg="blue.50" borderRadius="md" color="blue.500">
                        <FiFileText />
                      </Box>
                      <Text>Total Sales</Text>
                    </HStack>
                  </StatLabel>
                  <StatNumber>{summary.total_sales}</StatNumber>
                  <StatHelpText>Transactions</StatHelpText>
                </Stat>
              </CardBody>
            </Card>

            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>
                    <HStack>
                      <Box p={2} bg="green.50" borderRadius="md" color="green.500">
                        <FiDollarSign />
                      </Box>
                      <Text>Total Revenue</Text>
                    </HStack>
                  </StatLabel>
                  <StatNumber>{salesService.formatCurrency(summary.total_amount)}</StatNumber>
                  <StatHelpText>Gross revenue</StatHelpText>
                </Stat>
              </CardBody>
            </Card>

            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>
                    <HStack>
                      <Box p={2} bg="orange.50" borderRadius="md" color="orange.500">
                        <FiTrendingUp />
                      </Box>
                      <Text>Outstanding</Text>
                    </HStack>
                  </StatLabel>
                  <StatNumber color="orange.500">
                    {salesService.formatCurrency(summary.total_outstanding)}
                  </StatNumber>
                  <StatHelpText>Pending payments</StatHelpText>
                </Stat>
              </CardBody>
            </Card>

            <Card>
              <CardBody>
                <Stat>
                  <StatLabel>
                    <HStack>
                      <Box p={2} bg="purple.50" borderRadius="md" color="purple.500">
                        <FiUsers />
                      </Box>
                      <Text>Avg Order Value</Text>
                    </HStack>
                  </StatLabel>
                  <StatNumber>{salesService.formatCurrency(summary.avg_order_value)}</StatNumber>
                  <StatHelpText>Per transaction</StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </SimpleGrid>
        )}

        {/* Main Content Tabs */}
        <Tabs variant="enclosed" colorScheme="blue">
          <TabList>
            <Tab>Sales Charts</Tab>
            <Tab>Receivables Analysis</Tab>
            <Tab>Top Customers</Tab>
          </TabList>

          <TabPanels>
            {/* Sales Charts Tab */}
            <TabPanel px={0}>
              <SalesChart analytics={analytics} />
            </TabPanel>

            {/* Receivables Tab */}
            <TabPanel px={0}>
              <ReceivablesTable />
            </TabPanel>

            {/* Top Customers Tab */}
            <TabPanel px={0}>
              <Card>
                <CardHeader>
                  <Heading size="md">Top Customers by Revenue</Heading>
                </CardHeader>
                <CardBody>
                  {summary?.top_customers && summary.top_customers.length > 0 ? (
                    <VStack spacing={4} align="stretch">
                      {summary.top_customers.map((customer, index) => (
                        <Flex key={customer.customer_id} justify="space-between" align="center" p={4} bg="gray.50" borderRadius="md">
                          <HStack>
                            <Box
                              w={8}
                              h={8}
                              bg="blue.500"
                              color="white"
                              borderRadius="full"
                              display="flex"
                              alignItems="center"
                              justifyContent="center"
                              fontSize="sm"
                              fontWeight="bold"
                            >
                              {index + 1}
                            </Box>
                            <VStack align="start" spacing={0}>
                              <Text fontWeight="medium">{customer.customer_name}</Text>
                              <Text fontSize="sm" color="gray.600">
                                {customer.total_orders} orders
                              </Text>
                            </VStack>
                          </HStack>
                          <VStack align="end" spacing={0}>
                            <Text fontWeight="bold" color="green.600">
                              {salesService.formatCurrency(customer.total_amount)}
                            </Text>
                            <Text fontSize="sm" color="gray.600">
                              Avg: {salesService.formatCurrency(customer.total_amount / customer.total_orders)}
                            </Text>
                          </VStack>
                        </Flex>
                      ))}
                    </VStack>
                  ) : (
                    <Text textAlign="center" color="gray.500" py={10}>
                      No customer data available for the selected period
                    </Text>
                  )}
                </CardBody>
              </Card>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </VStack>
    </Layout>
  );
};

export default SalesAnalyticsPage;
