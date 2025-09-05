'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Container,
  Heading,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  VStack,
  HStack,
  Button,
  Select,
  Input,
  FormControl,
  FormLabel,
  Card,
  CardBody,
  CardHeader,
  SimpleGrid,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  StatArrow,
  Badge,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Spinner,
  useColorModeValue,
  useToast,
  Text,
  Divider,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
  Progress,
  Icon,
  Tooltip,
} from '@chakra-ui/react';
import { 
  FiDownload, 
  FiRefreshCw, 
  FiTrendingUp, 
  FiTrendingDown,
  FiDollarSign,
  FiPieChart,
  FiBarChart,
  FiActivity,
  FiUsers,
  FiShoppingCart
} from 'react-icons/fi';
import { 
  LineChart, 
  Line, 
  AreaChart, 
  Area, 
  BarChart, 
  Bar, 
  PieChart, 
  Pie, 
  Cell,
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip as RechartsTooltip, 
  Legend, 
  ResponsiveContainer 
} from 'recharts';
import axios from 'axios';

// Types for our comprehensive reports
interface CompanyInfo {
  name: string;
  address: string;
  city: string;
  state: string;
  postal_code: string;
  phone: string;
  email: string;
  website: string;
  tax_number: string;
}

interface BalanceSheetData {
  company: CompanyInfo;
  as_of_date: string;
  currency: string;
  assets: {
    name: string;
    items: Array<{
      account_id: number;
      code: string;
      name: string;
      balance: number;
      category: string;
      level: number;
      is_header: boolean;
    }>;
    subtotals: Array<{
      name: string;
      amount: number;
      category: string;
    }>;
    total: number;
  };
  liabilities: {
    name: string;
    items: Array<{
      account_id: number;
      code: string;
      name: string;
      balance: number;
      category: string;
      level: number;
      is_header: boolean;
    }>;
    subtotals: Array<{
      name: string;
      amount: number;
      category: string;
    }>;
    total: number;
  };
  equity: {
    name: string;
    items: Array<{
      account_id: number;
      code: string;
      name: string;
      balance: number;
      category: string;
      level: number;
      is_header: boolean;
    }>;
    total: number;
  };
  total_assets: number;
  total_equity: number;
  is_balanced: boolean;
  difference: number;
  generated_at: string;
}

interface ProfitLossData {
  company: CompanyInfo;
  start_date: string;
  end_date: string;
  currency: string;
  revenue: {
    name: string;
    items: Array<{
      account_id: number;
      code: string;
      name: string;
      amount: number;
      category: string;
      percentage: number;
    }>;
    subtotal: number;
  };
  cost_of_goods_sold: {
    name: string;
    items: Array<{
      account_id: number;
      code: string;
      name: string;
      amount: number;
      category: string;
      percentage: number;
    }>;
    subtotal: number;
  };
  gross_profit: number;
  gross_profit_margin: number;
  operating_expenses: {
    name: string;
    items: Array<{
      account_id: number;
      code: string;
      name: string;
      amount: number;
      category: string;
      percentage: number;
    }>;
    subtotal: number;
  };
  operating_income: number;
  net_income: number;
  net_income_margin: number;
  generated_at: string;
}

interface SalesSummaryData {
  company: CompanyInfo;
  start_date: string;
  end_date: string;
  currency: string;
  total_revenue: number;
  total_transactions: number;
  average_order_value: number;
  total_customers: number;
  sales_by_period: Array<{
    period: string;
    start_date: string;
    end_date: string;
    amount: number;
    transactions: number;
    growth_rate: number;
  }>;
  sales_by_customer: Array<{
    customer_id: number;
    customer_name: string;
    total_amount: number;
    transaction_count: number;
    average_order: number;
    last_order_date: string;
    first_order_date: string;
  }>;
  sales_by_product: Array<{
    product_id: number;
    product_name: string;
    quantity_sold: number;
    total_amount: number;
    average_price: number;
    transaction_count: number;
  }>;
  top_performers: {
    top_customers: Array<{
      customer_id: number;
      customer_name: string;
      total_amount: number;
      transaction_count: number;
      average_order: number;
    }>;
    top_products: Array<{
      product_id: number;
      product_name: string;
      quantity_sold: number;
      total_amount: number;
      average_price: number;
    }>;
  };
  growth_analysis: {
    month_over_month: number;
    quarter_over_quarter: number;
    year_over_year: number;
    trend_direction: string;
    seasonality_index: number;
  };
  generated_at: string;
}

interface FinancialDashboardData {
  period: {
    start_date: string;
    end_date: string;
  };
  balance_sheet: {
    total_assets: number;
    total_liabilities: number;
    total_equity: number;
    is_balanced: boolean;
    current_assets: number;
    fixed_assets: number;
    current_liabilities: number;
    long_term_liabilities: number;
  };
  profit_loss: {
    total_revenue: number;
    cost_of_goods_sold: number;
    gross_profit: number;
    gross_profit_margin: number;
    operating_expenses: number;
    operating_income: number;
    net_income: number;
    net_income_margin: number;
  };
  cash_flow: {
    beginning_cash: number;
    ending_cash: number;
    net_cash_flow: number;
    operating_cash_flow: number;
    investing_cash_flow: number;
    financing_cash_flow: number;
  };
  sales_summary: {
    total_revenue: number;
    total_transactions: number;
    average_order_value: number;
    total_customers: number;
  };
  purchase_summary: {
    total_purchases: number;
    total_transactions: number;
    average_purchase_value: number;
    total_vendors: number;
  };
  key_ratios: {
    current_ratio: number;
    debt_to_equity: number;
    return_on_assets: number;
    return_on_equity: number;
  };
  generated_at: string;
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Main component
export default function ComprehensiveReportsPage() {
  const [activeTab, setActiveTab] = useState(0);
  const [loading, setLoading] = useState(false);
  const [balanceSheetData, setBalanceSheetData] = useState<BalanceSheetData | null>(null);
  const [profitLossData, setProfitLossData] = useState<ProfitLossData | null>(null);
  const [salesSummaryData, setSalesSummaryData] = useState<SalesSummaryData | null>(null);
  const [dashboardData, setDashboardData] = useState<FinancialDashboardData | null>(null);
  
  // Form states
  const [asOfDate, setAsOfDate] = useState(new Date().toISOString().split('T')[0]);
  const [startDate, setStartDate] = useState(new Date(new Date().getFullYear(), new Date().getMonth(), 1).toISOString().split('T')[0]);
  const [endDate, setEndDate] = useState(new Date().toISOString().split('T')[0]);
  const [groupBy, setGroupBy] = useState('month');
  const [format, setFormat] = useState('json');

  const toast = useToast();
  const bgColor = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');

  // API functions
  const fetchBalanceSheet = async () => {
    setLoading(true);
    try {
      const response = await axios.get(`${API_BASE_URL}/api/reports/comprehensive/balance-sheet`, {
        params: { as_of_date: asOfDate, format },
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (format === 'json') {
        setBalanceSheetData(response.data.data);
      } else {
        // Handle file download
        const blob = new Blob([response.data], { 
          type: format === 'pdf' ? 'application/pdf' : 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' 
        });
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `balance_sheet_${asOfDate}.${format}`;
        link.click();
        window.URL.revokeObjectURL(url);
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch balance sheet data',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const fetchProfitLoss = async () => {
    setLoading(true);
    try {
      const response = await axios.get(`${API_BASE_URL}/api/reports/comprehensive/profit-loss`, {
        params: { start_date: startDate, end_date: endDate, format },
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (format === 'json') {
        setProfitLossData(response.data.data);
      } else {
        // Handle file download
        const blob = new Blob([response.data], { 
          type: format === 'pdf' ? 'application/pdf' : 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' 
        });
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `profit_loss_${startDate}_to_${endDate}.${format}`;
        link.click();
        window.URL.revokeObjectURL(url);
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch profit & loss data',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const fetchSalesSummary = async () => {
    setLoading(true);
    try {
      const response = await axios.get(`${API_BASE_URL}/api/reports/comprehensive/sales-summary`, {
        params: { start_date: startDate, end_date: endDate, group_by: groupBy, format },
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (format === 'json') {
        setSalesSummaryData(response.data.data);
      } else {
        // Handle file download
        const blob = new Blob([response.data], { 
          type: format === 'pdf' ? 'application/pdf' : 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' 
        });
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `sales_summary_${startDate}_to_${endDate}.${format}`;
        link.click();
        window.URL.revokeObjectURL(url);
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch sales summary data',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const fetchDashboard = async () => {
    setLoading(true);
    try {
      const response = await axios.get(`${API_BASE_URL}/api/reports/financial-dashboard`, {
        params: { start_date: startDate, end_date: endDate },
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token')}`,
        },
      });
      setDashboardData(response.data.data);
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch dashboard data',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  // Load dashboard data on component mount
  useEffect(() => {
    fetchDashboard();
  }, []);

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount);
  };

  const formatPercentage = (value: number) => {
    return `${value.toFixed(2)}%`;
  };

  // Chart colors
  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884d8'];

  return (
    <Container maxW="full" p={4}>
      <VStack spacing={6} align="stretch">
        <Box>
          <Heading size="lg" mb={2}>Comprehensive Financial Reports</Heading>
          <Text color="gray.600">
            Advanced financial and operational reporting with detailed analytics and business intelligence
          </Text>
        </Box>

        <Tabs index={activeTab} onChange={setActiveTab} variant="enclosed" colorScheme="blue">
          <TabList>
            <Tab>
              <Icon as={FiActivity} mr={2} />
              Financial Dashboard
            </Tab>
            <Tab>
              <Icon as={FiPieChart} mr={2} />
              Balance Sheet
            </Tab>
            <Tab>
              <Icon as={FiTrendingUp} mr={2} />
              Profit & Loss
            </Tab>
            <Tab>
              <Icon as={FiBarChart} mr={2} />
              Sales Analytics
            </Tab>
          </TabList>

          <TabPanels>
            {/* Financial Dashboard Tab */}
            <TabPanel>
              <VStack spacing={6} align="stretch">
                {/* Dashboard Controls */}
                <Card>
                  <CardHeader>
                    <Heading size="md">Dashboard Controls</Heading>
                  </CardHeader>
                  <CardBody>
                    <HStack spacing={4} wrap="wrap">
                      <FormControl maxW="200px">
                        <FormLabel>Start Date</FormLabel>
                        <Input
                          type="date"
                          value={startDate}
                          onChange={(e) => setStartDate(e.target.value)}
                        />
                      </FormControl>
                      <FormControl maxW="200px">
                        <FormLabel>End Date</FormLabel>
                        <Input
                          type="date"
                          value={endDate}
                          onChange={(e) => setEndDate(e.target.value)}
                        />
                      </FormControl>
                      <Button
                        leftIcon={<FiRefreshCw />}
                        colorScheme="blue"
                        onClick={fetchDashboard}
                        isLoading={loading}
                        alignSelf="flex-end"
                      >
                        Refresh Dashboard
                      </Button>
                    </HStack>
                  </CardBody>
                </Card>

                {dashboardData && (
                  <>
                    {/* Key Metrics Grid */}
                    <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
                      <Card>
                        <CardBody>
                          <Stat>
                            <StatLabel>Total Assets</StatLabel>
                            <StatNumber fontSize="2xl">
                              {formatCurrency(dashboardData.balance_sheet.total_assets)}
                            </StatNumber>
                            <StatHelpText>
                              <Badge colorScheme={dashboardData.balance_sheet.is_balanced ? 'green' : 'red'}>
                                {dashboardData.balance_sheet.is_balanced ? 'Balanced' : 'Unbalanced'}
                              </Badge>
                            </StatHelpText>
                          </Stat>
                        </CardBody>
                      </Card>

                      <Card>
                        <CardBody>
                          <Stat>
                            <StatLabel>Net Income</StatLabel>
                            <StatNumber fontSize="2xl" color={dashboardData.profit_loss.net_income >= 0 ? 'green.500' : 'red.500'}>
                              {formatCurrency(dashboardData.profit_loss.net_income)}
                            </StatNumber>
                            <StatHelpText>
                              <StatArrow type={dashboardData.profit_loss.net_income >= 0 ? 'increase' : 'decrease'} />
                              {formatPercentage(dashboardData.profit_loss.net_income_margin)} margin
                            </StatHelpText>
                          </Stat>
                        </CardBody>
                      </Card>

                      <Card>
                        <CardBody>
                          <Stat>
                            <StatLabel>Total Revenue</StatLabel>
                            <StatNumber fontSize="2xl">
                              {formatCurrency(dashboardData.sales_summary.total_revenue)}
                            </StatNumber>
                            <StatHelpText>
                              {dashboardData.sales_summary.total_transactions} transactions
                            </StatHelpText>
                          </Stat>
                        </CardBody>
                      </Card>

                      <Card>
                        <CardBody>
                          <Stat>
                            <StatLabel>Cash Flow</StatLabel>
                            <StatNumber fontSize="2xl" color={dashboardData.cash_flow.net_cash_flow >= 0 ? 'green.500' : 'red.500'}>
                              {formatCurrency(dashboardData.cash_flow.net_cash_flow)}
                            </StatNumber>
                            <StatHelpText>
                              <StatArrow type={dashboardData.cash_flow.net_cash_flow >= 0 ? 'increase' : 'decrease'} />
                              Net cash flow
                            </StatHelpText>
                          </Stat>
                        </CardBody>
                      </Card>
                    </SimpleGrid>

                    {/* Financial Ratios */}
                    <Card>
                      <CardHeader>
                        <Heading size="md">Key Financial Ratios</Heading>
                      </CardHeader>
                      <CardBody>
                        <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={6}>
                          <Box textAlign="center">
                            <Text fontSize="sm" color="gray.600">Current Ratio</Text>
                            <Text fontSize="2xl" fontWeight="bold" color="blue.500">
                              {dashboardData.key_ratios.current_ratio.toFixed(2)}
                            </Text>
                            <Progress 
                              value={Math.min(dashboardData.key_ratios.current_ratio * 50, 100)} 
                              colorScheme="blue" 
                              size="sm" 
                              mt={2}
                            />
                          </Box>

                          <Box textAlign="center">
                            <Text fontSize="sm" color="gray.600">Debt to Equity</Text>
                            <Text fontSize="2xl" fontWeight="bold" color="orange.500">
                              {dashboardData.key_ratios.debt_to_equity.toFixed(2)}
                            </Text>
                            <Progress 
                              value={Math.min(dashboardData.key_ratios.debt_to_equity * 100, 100)} 
                              colorScheme="orange" 
                              size="sm" 
                              mt={2}
                            />
                          </Box>

                          <Box textAlign="center">
                            <Text fontSize="sm" color="gray.600">Return on Assets</Text>
                            <Text fontSize="2xl" fontWeight="bold" color="green.500">
                              {formatPercentage(dashboardData.key_ratios.return_on_assets)}
                            </Text>
                            <Progress 
                              value={Math.min(Math.max(dashboardData.key_ratios.return_on_assets * 5, 0), 100)} 
                              colorScheme="green" 
                              size="sm" 
                              mt={2}
                            />
                          </Box>

                          <Box textAlign="center">
                            <Text fontSize="sm" color="gray.600">Return on Equity</Text>
                            <Text fontSize="2xl" fontWeight="bold" color="purple.500">
                              {formatPercentage(dashboardData.key_ratios.return_on_equity)}
                            </Text>
                            <Progress 
                              value={Math.min(Math.max(dashboardData.key_ratios.return_on_equity * 5, 0), 100)} 
                              colorScheme="purple" 
                              size="sm" 
                              mt={2}
                            />
                          </Box>
                        </SimpleGrid>
                      </CardBody>
                    </Card>
                  </>
                )}

                {loading && (
                  <Box textAlign="center" py={8}>
                    <Spinner size="xl" color="blue.500" />
                    <Text mt={4} color="gray.600">Loading dashboard data...</Text>
                  </Box>
                )}
              </VStack>
            </TabPanel>

            {/* Balance Sheet Tab */}
            <TabPanel>
              <VStack spacing={6} align="stretch">
                <Card>
                  <CardHeader>
                    <Heading size="md">Balance Sheet Parameters</Heading>
                  </CardHeader>
                  <CardBody>
                    <HStack spacing={4} wrap="wrap">
                      <FormControl maxW="200px">
                        <FormLabel>As of Date</FormLabel>
                        <Input
                          type="date"
                          value={asOfDate}
                          onChange={(e) => setAsOfDate(e.target.value)}
                        />
                      </FormControl>
                      <FormControl maxW="150px">
                        <FormLabel>Format</FormLabel>
                        <Select value={format} onChange={(e) => setFormat(e.target.value)}>
                          <option value="json">View Online</option>
                          <option value="pdf">Download PDF</option>
                          <option value="excel">Download Excel</option>
                        </Select>
                      </FormControl>
                      <Button
                        leftIcon={<FiDownload />}
                        colorScheme="blue"
                        onClick={fetchBalanceSheet}
                        isLoading={loading}
                        alignSelf="flex-end"
                      >
                        Generate Report
                      </Button>
                    </HStack>
                  </CardBody>
                </Card>

                {balanceSheetData && (
                  <Card>
                    <CardHeader>
                      <VStack align="start" spacing={2}>
                        <Heading size="lg">{balanceSheetData.company.name}</Heading>
                        <Heading size="md">Balance Sheet</Heading>
                        <Text color="gray.600">As of {new Date(balanceSheetData.as_of_date).toLocaleDateString()}</Text>
                        {!balanceSheetData.is_balanced && (
                          <Alert status="warning">
                            <AlertIcon />
                            <AlertTitle>Balance Sheet Warning!</AlertTitle>
                            <AlertDescription>
                              Assets and Liabilities+Equity don't match. Difference: {formatCurrency(balanceSheetData.difference)}
                            </AlertDescription>
                          </Alert>
                        )}
                      </VStack>
                    </CardHeader>
                    <CardBody>
                      <TableContainer>
                        <Table variant="simple">
                          <Thead>
                            <Tr>
                              <Th>Account</Th>
                              <Th textAlign="right">Amount</Th>
                            </Tr>
                          </Thead>
                          <Tbody>
                            {/* Assets Section */}
                            <Tr>
                              <Td fontWeight="bold" fontSize="lg" bg="blue.50">ASSETS</Td>
                              <Td textAlign="right" fontWeight="bold" fontSize="lg" bg="blue.50">
                                {formatCurrency(balanceSheetData.assets.total)}
                              </Td>
                            </Tr>
                            {balanceSheetData.assets.items.map((item, index) => (
                              <Tr key={index}>
                                <Td pl={item.level * 4 + 4}>
                                  {item.is_header ? (
                                    <Text fontWeight="semibold">{item.name}</Text>
                                  ) : (
                                    <Text>{item.name}</Text>
                                  )}
                                </Td>
                                <Td textAlign="right" fontWeight={item.is_header ? "semibold" : "normal"}>
                                  {formatCurrency(item.balance)}
                                </Td>
                              </Tr>
                            ))}
                            
                            {/* Liabilities Section */}
                            <Tr>
                              <Td fontWeight="bold" fontSize="lg" bg="orange.50" pt={8}>LIABILITIES</Td>
                              <Td textAlign="right" fontWeight="bold" fontSize="lg" bg="orange.50" pt={8}>
                                {formatCurrency(balanceSheetData.liabilities.total)}
                              </Td>
                            </Tr>
                            {balanceSheetData.liabilities.items.map((item, index) => (
                              <Tr key={index}>
                                <Td pl={item.level * 4 + 4}>
                                  {item.is_header ? (
                                    <Text fontWeight="semibold">{item.name}</Text>
                                  ) : (
                                    <Text>{item.name}</Text>
                                  )}
                                </Td>
                                <Td textAlign="right" fontWeight={item.is_header ? "semibold" : "normal"}>
                                  {formatCurrency(item.balance)}
                                </Td>
                              </Tr>
                            ))}
                            
                            {/* Equity Section */}
                            <Tr>
                              <Td fontWeight="bold" fontSize="lg" bg="green.50" pt={8}>EQUITY</Td>
                              <Td textAlign="right" fontWeight="bold" fontSize="lg" bg="green.50" pt={8}>
                                {formatCurrency(balanceSheetData.equity.total)}
                              </Td>
                            </Tr>
                            {balanceSheetData.equity.items.map((item, index) => (
                              <Tr key={index}>
                                <Td pl={item.level * 4 + 4}>
                                  <Text>{item.name}</Text>
                                </Td>
                                <Td textAlign="right">
                                  {formatCurrency(item.balance)}
                                </Td>
                              </Tr>
                            ))}
                            
                            {/* Total */}
                            <Tr borderTop="2px" borderColor="gray.300">
                              <Td fontWeight="bold" fontSize="lg">TOTAL LIABILITIES & EQUITY</Td>
                              <Td textAlign="right" fontWeight="bold" fontSize="lg">
                                {formatCurrency(balanceSheetData.liabilities.total + balanceSheetData.equity.total)}
                              </Td>
                            </Tr>
                          </Tbody>
                        </Table>
                      </TableContainer>
                    </CardBody>
                  </Card>
                )}
              </VStack>
            </TabPanel>

            {/* Profit & Loss Tab */}
            <TabPanel>
              <VStack spacing={6} align="stretch">
                <Card>
                  <CardHeader>
                    <Heading size="md">Profit & Loss Parameters</Heading>
                  </CardHeader>
                  <CardBody>
                    <HStack spacing={4} wrap="wrap">
                      <FormControl maxW="200px">
                        <FormLabel>Start Date</FormLabel>
                        <Input
                          type="date"
                          value={startDate}
                          onChange={(e) => setStartDate(e.target.value)}
                        />
                      </FormControl>
                      <FormControl maxW="200px">
                        <FormLabel>End Date</FormLabel>
                        <Input
                          type="date"
                          value={endDate}
                          onChange={(e) => setEndDate(e.target.value)}
                        />
                      </FormControl>
                      <FormControl maxW="150px">
                        <FormLabel>Format</FormLabel>
                        <Select value={format} onChange={(e) => setFormat(e.target.value)}>
                          <option value="json">View Online</option>
                          <option value="pdf">Download PDF</option>
                          <option value="excel">Download Excel</option>
                        </Select>
                      </FormControl>
                      <Button
                        leftIcon={<FiDownload />}
                        colorScheme="blue"
                        onClick={fetchProfitLoss}
                        isLoading={loading}
                        alignSelf="flex-end"
                      >
                        Generate Report
                      </Button>
                    </HStack>
                  </CardBody>
                </Card>

                {profitLossData && (
                  <Card>
                    <CardHeader>
                      <VStack align="start" spacing={2}>
                        <Heading size="lg">{profitLossData.company.name}</Heading>
                        <Heading size="md">Profit & Loss Statement</Heading>
                        <Text color="gray.600">
                          For the period {new Date(profitLossData.start_date).toLocaleDateString()} to {new Date(profitLossData.end_date).toLocaleDateString()}
                        </Text>
                      </VStack>
                    </CardHeader>
                    <CardBody>
                      <VStack spacing={6} align="stretch">
                        {/* Key Metrics */}
                        <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
                          <Stat>
                            <StatLabel>Total Revenue</StatLabel>
                            <StatNumber color="green.500">
                              {formatCurrency(profitLossData.revenue.subtotal)}
                            </StatNumber>
                          </Stat>
                          <Stat>
                            <StatLabel>Gross Profit</StatLabel>
                            <StatNumber>
                              {formatCurrency(profitLossData.gross_profit)}
                            </StatNumber>
                            <StatHelpText>
                              {formatPercentage(profitLossData.gross_profit_margin)} margin
                            </StatHelpText>
                          </Stat>
                          <Stat>
                            <StatLabel>Operating Income</StatLabel>
                            <StatNumber>
                              {formatCurrency(profitLossData.operating_income)}
                            </StatNumber>
                          </Stat>
                          <Stat>
                            <StatLabel>Net Income</StatLabel>
                            <StatNumber color={profitLossData.net_income >= 0 ? "green.500" : "red.500"}>
                              {formatCurrency(profitLossData.net_income)}
                            </StatNumber>
                            <StatHelpText>
                              {formatPercentage(profitLossData.net_income_margin)} margin
                            </StatHelpText>
                          </Stat>
                        </SimpleGrid>

                        <Divider />

                        {/* Detailed P&L Table */}
                        <TableContainer>
                          <Table variant="simple">
                            <Thead>
                              <Tr>
                                <Th>Account</Th>
                                <Th textAlign="right">Amount</Th>
                                <Th textAlign="right">% of Revenue</Th>
                              </Tr>
                            </Thead>
                            <Tbody>
                              {/* Revenue Section */}
                              <Tr>
                                <Td fontWeight="bold" fontSize="lg" bg="green.50">REVENUE</Td>
                                <Td textAlign="right" fontWeight="bold" fontSize="lg" bg="green.50">
                                  {formatCurrency(profitLossData.revenue.subtotal)}
                                </Td>
                                <Td textAlign="right" fontWeight="bold" fontSize="lg" bg="green.50">
                                  100.00%
                                </Td>
                              </Tr>
                              {profitLossData.revenue.items.map((item, index) => (
                                <Tr key={index}>
                                  <Td pl={8}>{item.name}</Td>
                                  <Td textAlign="right">{formatCurrency(item.amount)}</Td>
                                  <Td textAlign="right">{formatPercentage(item.percentage)}</Td>
                                </Tr>
                              ))}
                              
                              {/* COGS Section */}
                              <Tr>
                                <Td fontWeight="bold" fontSize="lg" bg="red.50" pt={6}>COST OF GOODS SOLD</Td>
                                <Td textAlign="right" fontWeight="bold" fontSize="lg" bg="red.50" pt={6}>
                                  {formatCurrency(profitLossData.cost_of_goods_sold.subtotal)}
                                </Td>
                                <Td textAlign="right" fontWeight="bold" fontSize="lg" bg="red.50" pt={6}>
                                  {formatPercentage((profitLossData.cost_of_goods_sold.subtotal / profitLossData.revenue.subtotal) * 100)}
                                </Td>
                              </Tr>
                              {profitLossData.cost_of_goods_sold.items.map((item, index) => (
                                <Tr key={index}>
                                  <Td pl={8}>{item.name}</Td>
                                  <Td textAlign="right">{formatCurrency(item.amount)}</Td>
                                  <Td textAlign="right">{formatPercentage(item.percentage)}</Td>
                                </Tr>
                              ))}

                              {/* Gross Profit */}
                              <Tr borderTop="1px" borderColor="gray.300">
                                <Td fontWeight="bold">GROSS PROFIT</Td>
                                <Td textAlign="right" fontWeight="bold">
                                  {formatCurrency(profitLossData.gross_profit)}
                                </Td>
                                <Td textAlign="right" fontWeight="bold">
                                  {formatPercentage(profitLossData.gross_profit_margin)}
                                </Td>
                              </Tr>

                              {/* Operating Expenses */}
                              <Tr>
                                <Td fontWeight="bold" fontSize="lg" bg="orange.50" pt={6}>OPERATING EXPENSES</Td>
                                <Td textAlign="right" fontWeight="bold" fontSize="lg" bg="orange.50" pt={6}>
                                  {formatCurrency(profitLossData.operating_expenses.subtotal)}
                                </Td>
                                <Td textAlign="right" fontWeight="bold" fontSize="lg" bg="orange.50" pt={6}>
                                  {formatPercentage((profitLossData.operating_expenses.subtotal / profitLossData.revenue.subtotal) * 100)}
                                </Td>
                              </Tr>
                              {profitLossData.operating_expenses.items.map((item, index) => (
                                <Tr key={index}>
                                  <Td pl={8}>{item.name}</Td>
                                  <Td textAlign="right">{formatCurrency(item.amount)}</Td>
                                  <Td textAlign="right">{formatPercentage(item.percentage)}</Td>
                                </Tr>
                              ))}

                              {/* Net Income */}
                              <Tr borderTop="2px" borderColor="gray.400">
                                <Td fontWeight="bold" fontSize="lg">NET INCOME</Td>
                                <Td textAlign="right" fontWeight="bold" fontSize="lg" color={profitLossData.net_income >= 0 ? "green.500" : "red.500"}>
                                  {formatCurrency(profitLossData.net_income)}
                                </Td>
                                <Td textAlign="right" fontWeight="bold" fontSize="lg">
                                  {formatPercentage(profitLossData.net_income_margin)}
                                </Td>
                              </Tr>
                            </Tbody>
                          </Table>
                        </TableContainer>
                      </VStack>
                    </CardBody>
                  </Card>
                )}
              </VStack>
            </TabPanel>

            {/* Sales Analytics Tab */}
            <TabPanel>
              <VStack spacing={6} align="stretch">
                <Card>
                  <CardHeader>
                    <Heading size="md">Sales Analytics Parameters</Heading>
                  </CardHeader>
                  <CardBody>
                    <HStack spacing={4} wrap="wrap">
                      <FormControl maxW="200px">
                        <FormLabel>Start Date</FormLabel>
                        <Input
                          type="date"
                          value={startDate}
                          onChange={(e) => setStartDate(e.target.value)}
                        />
                      </FormControl>
                      <FormControl maxW="200px">
                        <FormLabel>End Date</FormLabel>
                        <Input
                          type="date"
                          value={endDate}
                          onChange={(e) => setEndDate(e.target.value)}
                        />
                      </FormControl>
                      <FormControl maxW="150px">
                        <FormLabel>Group By</FormLabel>
                        <Select value={groupBy} onChange={(e) => setGroupBy(e.target.value)}>
                          <option value="day">Daily</option>
                          <option value="week">Weekly</option>
                          <option value="month">Monthly</option>
                          <option value="quarter">Quarterly</option>
                          <option value="year">Yearly</option>
                        </Select>
                      </FormControl>
                      <FormControl maxW="150px">
                        <FormLabel>Format</FormLabel>
                        <Select value={format} onChange={(e) => setFormat(e.target.value)}>
                          <option value="json">View Online</option>
                          <option value="pdf">Download PDF</option>
                          <option value="excel">Download Excel</option>
                        </Select>
                      </FormControl>
                      <Button
                        leftIcon={<FiDownload />}
                        colorScheme="blue"
                        onClick={fetchSalesSummary}
                        isLoading={loading}
                        alignSelf="flex-end"
                      >
                        Generate Report
                      </Button>
                    </HStack>
                  </CardBody>
                </Card>

                {salesSummaryData && (
                  <VStack spacing={6} align="stretch">
                    {/* Sales Summary Stats */}
                    <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={4}>
                      <Card>
                        <CardBody>
                          <Stat>
                            <StatLabel>
                              <HStack>
                                <Icon as={FiDollarSign} color="green.500" />
                                <Text>Total Revenue</Text>
                              </HStack>
                            </StatLabel>
                            <StatNumber color="green.500">
                              {formatCurrency(salesSummaryData.total_revenue)}
                            </StatNumber>
                          </Stat>
                        </CardBody>
                      </Card>

                      <Card>
                        <CardBody>
                          <Stat>
                            <StatLabel>
                              <HStack>
                                <Icon as={FiShoppingCart} color="blue.500" />
                                <Text>Total Transactions</Text>
                              </HStack>
                            </StatLabel>
                            <StatNumber>{salesSummaryData.total_transactions.toLocaleString()}</StatNumber>
                          </Stat>
                        </CardBody>
                      </Card>

                      <Card>
                        <CardBody>
                          <Stat>
                            <StatLabel>
                              <HStack>
                                <Icon as={FiUsers} color="purple.500" />
                                <Text>Total Customers</Text>
                              </HStack>
                            </StatLabel>
                            <StatNumber>{salesSummaryData.total_customers.toLocaleString()}</StatNumber>
                          </Stat>
                        </CardBody>
                      </Card>

                      <Card>
                        <CardBody>
                          <Stat>
                            <StatLabel>Average Order Value</StatLabel>
                            <StatNumber>
                              {formatCurrency(salesSummaryData.average_order_value)}
                            </StatNumber>
                          </Stat>
                        </CardBody>
                      </Card>
                    </SimpleGrid>

                    {/* Sales by Period Chart */}
                    <Card>
                      <CardHeader>
                        <Heading size="md">Sales Trend by Period</Heading>
                      </CardHeader>
                      <CardBody>
                        <Box height="400px">
                          <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={salesSummaryData.sales_by_period}>
                              <CartesianGrid strokeDasharray="3 3" />
                              <XAxis dataKey="period" />
                              <YAxis tickFormatter={(value) => formatCurrency(value)} />
                              <RechartsTooltip formatter={(value) => [formatCurrency(Number(value)), 'Revenue']} />
                              <Legend />
                              <Area 
                                type="monotone" 
                                dataKey="amount" 
                                stroke="#8884d8" 
                                fill="#8884d8" 
                                fillOpacity={0.6}
                                name="Revenue"
                              />
                            </AreaChart>
                          </ResponsiveContainer>
                        </Box>
                      </CardBody>
                    </Card>

                    {/* Top Customers and Products */}
                    <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={4}>
                      <Card>
                        <CardHeader>
                          <Heading size="md">Top Customers</Heading>
                        </CardHeader>
                        <CardBody>
                          <TableContainer>
                            <Table size="sm">
                              <Thead>
                                <Tr>
                                  <Th>Customer</Th>
                                  <Th textAlign="right">Revenue</Th>
                                  <Th textAlign="right">Orders</Th>
                                </Tr>
                              </Thead>
                              <Tbody>
                                {salesSummaryData.top_performers.top_customers.slice(0, 5).map((customer, index) => (
                                  <Tr key={customer.customer_id}>
                                    <Td>
                                      <VStack align="start" spacing={1}>
                                        <Text fontWeight="medium">{customer.customer_name}</Text>
                                        <Text fontSize="sm" color="gray.500">
                                          Avg: {formatCurrency(customer.average_order)}
                                        </Text>
                                      </VStack>
                                    </Td>
                                    <Td textAlign="right" fontWeight="medium">
                                      {formatCurrency(customer.total_amount)}
                                    </Td>
                                    <Td textAlign="right">
                                      {customer.transaction_count}
                                    </Td>
                                  </Tr>
                                ))}
                              </Tbody>
                            </Table>
                          </TableContainer>
                        </CardBody>
                      </Card>

                      <Card>
                        <CardHeader>
                          <Heading size="md">Top Products</Heading>
                        </CardHeader>
                        <CardBody>
                          <TableContainer>
                            <Table size="sm">
                              <Thead>
                                <Tr>
                                  <Th>Product</Th>
                                  <Th textAlign="right">Revenue</Th>
                                  <Th textAlign="right">Qty Sold</Th>
                                </Tr>
                              </Thead>
                              <Tbody>
                                {salesSummaryData.top_performers.top_products.slice(0, 5).map((product, index) => (
                                  <Tr key={product.product_id}>
                                    <Td>
                                      <VStack align="start" spacing={1}>
                                        <Text fontWeight="medium">{product.product_name}</Text>
                                        <Text fontSize="sm" color="gray.500">
                                          Avg: {formatCurrency(product.average_price)}
                                        </Text>
                                      </VStack>
                                    </Td>
                                    <Td textAlign="right" fontWeight="medium">
                                      {formatCurrency(product.total_amount)}
                                    </Td>
                                    <Td textAlign="right">
                                      {product.quantity_sold.toLocaleString()}
                                    </Td>
                                  </Tr>
                                ))}
                              </Tbody>
                            </Table>
                          </TableContainer>
                        </CardBody>
                      </Card>
                    </SimpleGrid>

                    {/* Growth Analysis */}
                    <Card>
                      <CardHeader>
                        <Heading size="md">Growth Analysis</Heading>
                      </CardHeader>
                      <CardBody>
                        <SimpleGrid columns={{ base: 1, md: 3 }} spacing={6}>
                          <Stat>
                            <StatLabel>Month over Month</StatLabel>
                            <StatNumber>
                              <HStack>
                                <StatArrow type={salesSummaryData.growth_analysis.month_over_month >= 0 ? 'increase' : 'decrease'} />
                                <Text>{formatPercentage(Math.abs(salesSummaryData.growth_analysis.month_over_month))}</Text>
                              </HStack>
                            </StatNumber>
                          </Stat>
                          <Stat>
                            <StatLabel>Quarter over Quarter</StatLabel>
                            <StatNumber>
                              <HStack>
                                <StatArrow type={salesSummaryData.growth_analysis.quarter_over_quarter >= 0 ? 'increase' : 'decrease'} />
                                <Text>{formatPercentage(Math.abs(salesSummaryData.growth_analysis.quarter_over_quarter))}</Text>
                              </HStack>
                            </StatNumber>
                          </Stat>
                          <Stat>
                            <StatLabel>Year over Year</StatLabel>
                            <StatNumber>
                              <HStack>
                                <StatArrow type={salesSummaryData.growth_analysis.year_over_year >= 0 ? 'increase' : 'decrease'} />
                                <Text>{formatPercentage(Math.abs(salesSummaryData.growth_analysis.year_over_year))}</Text>
                              </HStack>
                            </StatNumber>
                          </Stat>
                        </SimpleGrid>
                        
                        <Box mt={4}>
                          <Text fontSize="sm" color="gray.600">
                            Trend Direction: 
                            <Badge ml={2} colorScheme={salesSummaryData.growth_analysis.trend_direction === 'UP' ? 'green' : 'red'}>
                              {salesSummaryData.growth_analysis.trend_direction}
                            </Badge>
                          </Text>
                        </Box>
                      </CardBody>
                    </Card>
                  </VStack>
                )}
              </VStack>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </VStack>
    </Container>
  );
}
