import React, { useState } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Box,
  Text,
  VStack,
  HStack,
  Button,
  SimpleGrid,
  Badge,
  Flex,
  useColorModeValue,
  Grid,
  GridItem,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  StatArrow,
  Card,
  CardBody,
  CardHeader,
  Heading,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  useToast,
  Spinner,
  Icon
} from '@chakra-ui/react';
import { FiDownload, FiShoppingCart, FiDollarSign, FiPieChart, FiUsers, FiTrendingUp } from 'react-icons/fi';
import { 
  FormControl,
  FormLabel,
  Input
} from '@chakra-ui/react';
import { formatCurrency } from '../../utils/formatters';
import { SSOTSalesSummaryData } from '../../services/ssotSalesSummaryService';

interface SalesSummaryModalProps {
  isOpen: boolean;
  onClose: () => void;
  data: SSOTSalesSummaryData | null;
  isLoading: boolean;
  error: string | null;
  startDate: string;
  endDate: string;
  onDateChange?: (startDate: string, endDate: string) => void;
  onFetch?: () => void;
  onExport?: (format: 'pdf' | 'excel') => void;
}

const SalesSummaryModal: React.FC<SalesSummaryModalProps> = ({
  isOpen,
  onClose,
  data,
  isLoading,
  error,
  startDate,
  endDate,
  onDateChange,
  onFetch,
  onExport
}) => {
  const [activeTab, setActiveTab] = useState<'summary'>('summary');
  const toast = useToast();
  
  // Color mode values
  const modalBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const sectionBg = useColorModeValue('gray.50', 'gray.700');
  const textColor = useColorModeValue('gray.800', 'white');
  const secondaryTextColor = useColorModeValue('gray.600', 'gray.300');
  const loadingTextColor = useColorModeValue('gray.700', 'gray.300');
  
  const handleExport = (format: 'pdf' | 'excel') => {
    if (onExport) {
      onExport(format);
    } else {
      if (data) {
        // Fallback: download as JSON
        const reportData = {
          reportType: 'Sales Summary',
          period: `${startDate} to ${endDate}`,
          generatedOn: new Date().toISOString(),
          data: data
        };
        const dataStr = JSON.stringify(reportData, null, 2);
        const dataBlob = new Blob([dataStr], { type: 'application/json' });
        const url = URL.createObjectURL(dataBlob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `sales-summary-${startDate}-to-${endDate}.json`;
        link.click();
        URL.revokeObjectURL(url);
      }
      
      toast({
        title: 'Export Feature',
        description: `${format.toUpperCase()} export will be implemented soon`,
        status: 'info',
        duration: 3000,
        isClosable: true,
      });
    }
  };

  const renderSummaryMetrics = () => {
    if (!data) return null;
    
    return (
      <Grid templateColumns="repeat(auto-fit, minmax(240px, 1fr))" gap={4} mb={6}>
        <GridItem>
          <Card size="sm">
            <CardBody>
              <Stat>
                <StatLabel>Total Revenue</StatLabel>
                <StatNumber color="green.600">
                  {formatCurrency(data.total_revenue || data.total_sales || 0)}
                </StatNumber>
                <StatHelpText>
                  <StatArrow type="increase" />
                  Sales for the period
                </StatHelpText>
              </Stat>
            </CardBody>
          </Card>
        </GridItem>
        
        <GridItem>
          <Card size="sm">
            <CardBody>
              <Stat>
                <StatLabel>Total Customers</StatLabel>
                <StatNumber color="blue.600">
                  {data.total_customers || (data.sales_by_customer?.length) || 0}
                </StatNumber>
                <StatHelpText>
                  <Icon as={FiUsers} />
                  Active customers
                </StatHelpText>
              </Stat>
            </CardBody>
          </Card>
        </GridItem>
        
        {data.total_orders && (
          <GridItem>
            <Card size="sm">
              <CardBody>
                <Stat>
                  <StatLabel>Total Orders</StatLabel>
                  <StatNumber color="purple.600">
                    {data.total_orders}
                  </StatNumber>
                  <StatHelpText>
                    <Icon as={FiShoppingCart} />
                    Orders processed
                  </StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </GridItem>
        )}
        
        {data.average_order_value && (
          <GridItem>
            <Card size="sm">
              <CardBody>
                <Stat>
                  <StatLabel>Average Order Value</StatLabel>
                  <StatNumber color="orange.600">
                    {formatCurrency(data.average_order_value)}
                  </StatNumber>
                  <StatHelpText>
                    <Icon as={FiTrendingUp} />
                    Per order average
                  </StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </GridItem>
        )}
      </Grid>
    );
  };

  // Note: This function is currently unused as the Customers tab has been removed
  const renderCustomersTab = () => {
    if (!data) return null;

    const customers = data.sales_by_customer || [];
    const topCustomers = data.top_customers || [];

    return (
      <VStack spacing={6} align="stretch">
        {customers.length > 0 && (
          <Card>
            <CardHeader>
              <Heading size="sm">Sales by Customer</Heading>
            </CardHeader>
            <CardBody>
              <Table size="sm">
                <Thead>
                  <Tr>
                    <Th>Customer</Th>
                    <Th>Contact</Th>
                    <Th isNumeric>Total Sales</Th>
                    <Th isNumeric>Orders</Th>
                    <Th isNumeric>Avg Order</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {customers.map((customer: any, index: number) => (
                    <Tr key={index}>
                      <Td>
                        <VStack align="start" spacing={1}>
                          <Text fontWeight="medium">
                            {customer.customer_name || customer.name || 'Unnamed Customer'}
                          </Text>
                          {customer.customer_code && (
                            <Text fontSize="xs" color="gray.500">
                              Code: {customer.customer_code}
                            </Text>
                          )}
                          {customer.customer_type && (
                            <Badge colorScheme="blue" size="sm">
                              {customer.customer_type}
                            </Badge>
                          )}
                        </VStack>
                      </Td>
                      <Td>
                        <VStack align="start" spacing={0}>
                          {customer.contact_person && (
                            <Text fontSize="sm">{customer.contact_person}</Text>
                          )}
                          {customer.phone && (
                            <Text fontSize="xs" color="gray.500">{customer.phone}</Text>
                          )}
                          {customer.email && (
                            <Text fontSize="xs" color="blue.500">{customer.email}</Text>
                          )}
                        </VStack>
                      </Td>
                      <Td isNumeric>
                        <Text fontWeight="bold" color="green.600">
                          {formatCurrency(customer.total_sales || 0)}
                        </Text>
                      </Td>
                      <Td isNumeric>
                        <Text color="purple.600">
                          {customer.transaction_count || 0}
                        </Text>
                      </Td>
                      <Td isNumeric>
                        <Text color="orange.600">
                          {formatCurrency(
                            customer.average_transaction ||
                            (customer.total_sales && customer.transaction_count > 0 
                              ? customer.total_sales / customer.transaction_count 
                              : 0)
                          )}
                        </Text>
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            </CardBody>
          </Card>
        )}

        {topCustomers.length > 0 && (
          <Card>
            <CardHeader>
              <Heading size="sm">Top Performing Customers</Heading>
            </CardHeader>
            <CardBody>
              <SimpleGrid columns={[1, 2, 3]} spacing={4}>
                {topCustomers.map((customer: any, index: number) => (
                  <Box key={index} border="1px" borderColor={borderColor} borderRadius="md" p={4}>
                    <VStack spacing={3}>
                      <Badge colorScheme="gold" size="lg" variant="solid">
                        #{index + 1}
                      </Badge>
                      <Text fontWeight="bold" fontSize="md" textAlign="center">
                        {customer.customer_name || customer.name}
                      </Text>
                      <Text fontSize="lg" fontWeight="bold" color="green.600">
                        {formatCurrency(customer.total_amount || customer.total_sales)}
                      </Text>
                      {customer.percentage && (
                        <Text fontSize="sm" color="gray.500">
                          {customer.percentage.toFixed(1)}% of total
                        </Text>
                      )}
                      {customer.order_count && (
                        <Text fontSize="xs" color="purple.500">
                          {customer.order_count} orders
                        </Text>
                      )}
                    </VStack>
                  </Box>
                ))}
              </SimpleGrid>
            </CardBody>
          </Card>
        )}

        {customers.length === 0 && topCustomers.length === 0 && (
          <Box textAlign="center" py={8}>
            <Text color={secondaryTextColor}>
              No customer data available for this period
            </Text>
          </Box>
        )}
      </VStack>
    );
  };

  // Note: This function is currently unused as the Analysis tab has been removed
  const renderAnalysisTab = () => {
    if (!data?.sales_trends) {
      return (
        <Box textAlign="center" py={8}>
          <Text color={secondaryTextColor}>
            Sales analysis not available for this report format
          </Text>
        </Box>
      );
    }

    const trends = data.sales_trends;
    
    return (
      <VStack spacing={6} align="stretch">
        <Card>
          <CardHeader>
            <Heading size="sm">Sales Performance Analysis</Heading>
          </CardHeader>
          <CardBody>
            <SimpleGrid columns={[1, 2, 4]} spacing={4}>
              {trends.growth_rate !== undefined && (
                <Box p={4} bg={trends.growth_rate >= 0 ? 'green.50' : 'red.50'} borderRadius="md" textAlign="center">
                  <Text fontSize="2xl" fontWeight="bold" color={trends.growth_rate >= 0 ? 'green.600' : 'red.600'}>
                    {trends.growth_rate > 0 ? '+' : ''}{trends.growth_rate.toFixed(1)}%
                  </Text>
                  <Text fontSize="sm" color={trends.growth_rate >= 0 ? 'green.800' : 'red.800'}>
                    Growth Rate
                  </Text>
                </Box>
              )}
              
              {trends.best_performing_period && (
                <Box p={4} bg="blue.50" borderRadius="md" textAlign="center">
                  <Text fontSize="md" fontWeight="bold" color="blue.600">
                    {trends.best_performing_period}
                  </Text>
                  <Text fontSize="sm" color="blue.800">
                    Best Period
                  </Text>
                </Box>
              )}
              
              {trends.recurring_customers !== undefined && (
                <Box p={4} bg="purple.50" borderRadius="md" textAlign="center">
                  <Text fontSize="2xl" fontWeight="bold" color="purple.600">
                    {trends.recurring_customers}
                  </Text>
                  <Text fontSize="sm" color="purple.800">
                    Recurring Customers
                  </Text>
                </Box>
              )}
              
              {trends.new_customers !== undefined && (
                <Box p={4} bg="teal.50" borderRadius="md" textAlign="center">
                  <Text fontSize="2xl" fontWeight="bold" color="teal.600">
                    {trends.new_customers}
                  </Text>
                  <Text fontSize="sm" color="teal.800">
                    New Customers
                  </Text>
                </Box>
              )}
            </SimpleGrid>
          </CardBody>
        </Card>
        
        <Card>
          <CardHeader>
            <Heading size="sm">Key Insights</Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={3} align="stretch">
              <Text fontSize="sm">
                📈 <strong>Sales Performance:</strong> {trends.growth_rate && trends.growth_rate > 0 ? 'Your sales are growing positively' : 'Sales performance shows room for improvement'}
              </Text>
              <Text fontSize="sm">
                👥 <strong>Customer Base:</strong> {trends.recurring_customers || 0} recurring customers indicate strong customer loyalty
              </Text>
              <Text fontSize="sm">
                🎯 <strong>Market Expansion:</strong> {trends.new_customers || 0} new customers shows business growth potential
              </Text>
            </VStack>
          </CardBody>
        </Card>
      </VStack>
    );
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="6xl" scrollBehavior="inside">
      <ModalOverlay />
      <ModalContent bg={modalBg} maxH="90vh">
        <ModalHeader borderBottom="1px" borderColor={borderColor}>
          <VStack align="stretch" spacing={2}>
            <HStack justify="space-between">
              <VStack align="start" spacing={0}>
                <HStack>
                  <Icon as={FiShoppingCart} color="blue.500" />
                  <Text fontSize="xl" fontWeight="bold">
                    Sales Summary Report
                  </Text>
                </HStack>
                <Text fontSize="sm" color={secondaryTextColor}>
                  Period: {startDate} - {endDate}
                </Text>
                {data?.company && (
                  <Text fontSize="sm" color={secondaryTextColor}>
                    {data.company.name}
                  </Text>
                )}
              </VStack>
              <Badge colorScheme="blue" variant="solid">
                SSOT Integration
              </Badge>
            </HStack>
            
            <HStack spacing={1}>
              <Button
                size="sm"
                variant="solid"
                leftIcon={<FiDollarSign />}
                disabled
              >
                Summary
              </Button>
            </HStack>
          </VStack>
        </ModalHeader>
        <ModalCloseButton />

        <ModalBody py={6}>
          {/* Date Range Controls */}
          <Box mb={4}>
            <HStack spacing={4} mb={4}>
              <FormControl>
                <FormLabel>Start Date</FormLabel>
                <Input 
                  type="date" 
                  value={startDate} 
                  onChange={(e) => onDateChange && onDateChange(e.target.value, endDate)} 
                />
              </FormControl>
              <FormControl>
                <FormLabel>End Date</FormLabel>
                <Input 
                  type="date" 
                  value={endDate} 
                  onChange={(e) => onDateChange && onDateChange(startDate, e.target.value)} 
                />
              </FormControl>
              <Button
                colorScheme="blue"
                onClick={onFetch}
                isLoading={isLoading}
                leftIcon={<FiShoppingCart />}
                size="md"
                mt={8}
              >
                Generate Report
              </Button>
            </HStack>
          </Box>

          {isLoading && (
            <Box textAlign="center" py={8}>
              <VStack spacing={4}>
                <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
                <VStack spacing={2}>
                  <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                    Generating Sales Summary Report
                  </Text>
                  <Text fontSize="sm" color={secondaryTextColor}>
                    Analyzing sales transactions from SSOT journal system...
                  </Text>
                </VStack>
              </VStack>
            </Box>
          )}

          {error && (
            <Box bg="red.50" p={4} borderRadius="md" mb={4}>
              <Text color="red.600" fontWeight="medium">Error: {error}</Text>
              <Button
                mt={2}
                size="sm"
                colorScheme="red"
                variant="outline"
                onClick={onFetch}
              >
                Retry
              </Button>
            </Box>
          )}

          {data && !isLoading && (
            <>
              {/* Company Header */}
              {data.company && (
                <Box bg={sectionBg} p={4} borderRadius="md" mb={6}>
                  <HStack justify="space-between" align="start">
                    <VStack align="start" spacing={1}>
                      <Text fontSize="lg" fontWeight="bold" color={textColor}>
                        {data.company.name || 'Company Name'}
                      </Text>
                      <Text fontSize="sm" color={secondaryTextColor}>
                        {data.company.address && data.company.city ? 
                          `${data.company.address}, ${data.company.city}` : 
                          'Address not available'
                        }
                      </Text>
                      {data.company.phone && (
                        <Text fontSize="sm" color={secondaryTextColor}>
                          {data.company.phone} | {data.company.email}
                        </Text>
                      )}
                    </VStack>
                    <VStack align="end" spacing={1}>
                      <Text fontSize="sm" color={secondaryTextColor}>
                        Currency: {data.currency || 'IDR'}
                      </Text>
                      <Text fontSize="xs" color={secondaryTextColor}>
                        Generated: {data.generated_at ? new Date(data.generated_at).toLocaleString('id-ID') : new Date().toLocaleString('id-ID')}
                      </Text>
                    </VStack>
                  </HStack>
                </Box>
              )}

              <VStack spacing={4} align="stretch">
                {renderSummaryMetrics()}
                
                {/* Period Summary */}
                <Card>
                  <CardBody>
                    <Flex justify="space-between" align="center" mb={3}>
                      <Heading size="md" color={textColor}>
                        Sales Performance
                      </Heading>
                      <Text fontWeight="bold" fontSize="lg" color="green.600">
                        {formatCurrency(data.total_revenue || data.total_sales || 0)}
                      </Text>
                    </Flex>
                    
                    <SimpleGrid columns={[1, 3]} spacing={4}>
                      <Box textAlign="center" p={3} bg={sectionBg} borderRadius="md">
                        <Text fontSize="sm" color={secondaryTextColor}>Period</Text>
                        <Text fontWeight="medium">{startDate} to {endDate}</Text>
                      </Box>
                      <Box textAlign="center" p={3} bg={sectionBg} borderRadius="md">
                        <Text fontSize="sm" color={secondaryTextColor}>Report Type</Text>
                        <Text fontWeight="medium">SSOT Integration</Text>
                      </Box>
                      <Box textAlign="center" p={3} bg={sectionBg} borderRadius="md">
                        <Text fontSize="sm" color={secondaryTextColor}>Status</Text>
                        <Badge colorScheme="green">Active</Badge>
                      </Box>
                    </SimpleGrid>
                  </CardBody>
                </Card>
              </VStack>
            </>
          )}
        </ModalBody>

        <ModalFooter borderTop="1px" borderColor={borderColor}>
          <HStack spacing={3}>
            <Button
              leftIcon={<FiDownload />}
              size="sm"
              variant="outline"
              onClick={() => handleExport('pdf')}
              isDisabled={isLoading || !data}
            >
              Export PDF
            </Button>
            <Button
              leftIcon={<FiDownload />}
              size="sm"
              variant="outline"
              onClick={() => handleExport('excel')}
              isDisabled={isLoading || !data}
            >
              Export Excel
            </Button>
            <Button onClick={onClose} size="sm">
              Close
            </Button>
          </HStack>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default SalesSummaryModal;