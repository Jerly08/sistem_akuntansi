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
import { FiDownload, FiShoppingCart, FiDollarSign, FiPieChart, FiUsers, FiTrendingUp, FiFilePlus, FiFileText } from 'react-icons/fi';
import { 
  FormControl,
  FormLabel,
  Input
} from '@chakra-ui/react';
import { formatCurrency } from '../../utils/formatters';
import { SSOTPurchaseReportData, VendorPurchaseSummary } from '../../services/ssotPurchaseReportService';

interface PurchaseReportModalProps {
  isOpen: boolean;
  onClose: () => void;
  data: SSOTPurchaseReportData | null;
  isLoading: boolean;
  error: string | null;
  startDate: string;
  endDate: string;
  onDateChange?: (startDate: string, endDate: string) => void;
  onFetch?: () => void;
  onExport?: (format: 'pdf' | 'excel') => void;
}

const PurchaseReportModal: React.FC<PurchaseReportModalProps> = ({
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
  const previewPeriodTextColor = useColorModeValue('gray.500', 'gray.400');
  
  const handleExport = (format: 'pdf' | 'excel') => {
    if (onExport) {
      onExport(format);
    } else {
      if (data) {
        // Fallback: download as JSON
        const reportData = {
          reportType: 'Purchase Report',
          period: `${startDate} to ${endDate}`,
          generatedOn: new Date().toISOString(),
          data: data
        };
        const dataStr = JSON.stringify(reportData, null, 2);
        const dataBlob = new Blob([dataStr], { type: 'application/json' });
        const url = URL.createObjectURL(dataBlob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `purchase-report-${startDate}-to-${endDate}.json`;
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
    
    // Calculate total vendors from purchases_by_vendor array
    const totalVendors = data.purchases_by_vendor ? data.purchases_by_vendor.length : 0;
    
    // Calculate total orders (purchases)
    const totalOrders = data.total_purchases || 0;
    
    // Calculate average order value
    const averageOrderValue = totalOrders > 0 ? (data.total_amount || 0) / totalOrders : 0;
    
    return (
      <Grid templateColumns="repeat(auto-fit, minmax(240px, 1fr))" gap={4} mb={6}>
        <GridItem>
          <Card size="sm">
            <CardBody>
              <Stat>
                <StatLabel>Total Purchase Amount</StatLabel>
                <StatNumber color="green.600">
                  {formatCurrency(data.total_amount || 0)}
                </StatNumber>
                <StatHelpText>
                  <StatArrow type="increase" />
                  Purchases for the period
                </StatHelpText>
              </Stat>
            </CardBody>
          </Card>
        </GridItem>
        
        <GridItem>
          <Card size="sm">
            <CardBody>
              <Stat>
                <StatLabel>Total Vendors</StatLabel>
                <StatNumber color="blue.600">
                  {totalVendors}
                </StatNumber>
                <StatHelpText>
                  <Icon as={FiUsers} />
                  Active vendors
                </StatHelpText>
              </Stat>
            </CardBody>
          </Card>
        </GridItem>
        
        {totalOrders > 0 && (
          <GridItem>
            <Card size="sm">
              <CardBody>
                <Stat>
                  <StatLabel>Total Purchases</StatLabel>
                  <StatNumber color="purple.600">
                    {totalOrders}
                  </StatNumber>
                  <StatHelpText>
                    <Icon as={FiShoppingCart} />
                    Purchase orders
                  </StatHelpText>
                </Stat>
              </CardBody>
            </Card>
          </GridItem>
        )}
        
        {averageOrderValue > 0 && (
          <GridItem>
            <Card size="sm">
              <CardBody>
                <Stat>
                  <StatLabel>Average Purchase Value</StatLabel>
                  <StatNumber color="orange.600">
                    {formatCurrency(averageOrderValue)}
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

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="6xl" scrollBehavior="inside">
      <ModalOverlay />
      <ModalContent bg={modalBg}>
        <ModalHeader>
          <HStack>
            <Icon as={FiShoppingCart} color="orange.500" />
            <VStack align="start" spacing={0}>
              <Text fontSize="lg" fontWeight="bold">
                Purchase Report (SSOT)
              </Text>
              <Text fontSize="sm" color={previewPeriodTextColor}>
                {startDate} - {endDate} | SSOT Journal Integration
              </Text>
            </VStack>
          </HStack>
        </ModalHeader>
        <ModalCloseButton />
        
        <ModalBody pb={6}>
          {/* Date Range Controls - Moved to top like other modals */}
          <Box mb={4}>
            <HStack spacing={4} mb={4} flexWrap="wrap">
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
                whiteSpace="nowrap"
              >
                Generate Report
              </Button>
            </HStack>
          </Box>

          {isLoading && (
            <Box textAlign="center" py={8}>
              <VStack spacing={4}>
                <Spinner size="xl" thickness="4px" speed="0.65s" color="orange.500" />
                <VStack spacing={2}>
                  <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                    Generating Purchase Report
                  </Text>
                  <Text fontSize="sm" color={secondaryTextColor}>
                    Analyzing purchase transactions from SSOT journal system...
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
            <VStack spacing={6} align="stretch">
              {/* Company Header */}
              {data.company && (
                <Box bg="orange.50" p={4} borderRadius="md">
                  <HStack justify="space-between" align="start">
                    <VStack align="start" spacing={1}>
                      <Text fontSize="lg" fontWeight="bold" color="orange.800">
                        {data.company.name || 'PT. Sistem Akuntansi'}
                      </Text>
                      <Text fontSize="sm" color="orange.600">
                        {data.company.address ? (
                          data.company.city ? `${data.company.address}, ${data.company.city}` : data.company.address
                        ) : 'Address not available'}
                      </Text>
                      {data.company.phone && (
                        <Text fontSize="sm" color="orange.600">
                          {data.company.phone} | {data.company.email}
                        </Text>
                      )}
                    </VStack>
                    <VStack align="end" spacing={1}>
                      <Text fontSize="sm" color="orange.600">
                        Currency: {data.currency || 'IDR'}
                      </Text>
                      <Text fontSize="xs" color="orange.500">
                        Generated: {data.generated_at ? new Date(data.generated_at).toLocaleString('id-ID') : new Date().toLocaleString('id-ID')}
                      </Text>
                    </VStack>
                  </HStack>
                </Box>
              )}

              {/* Report Header */}
              <Box textAlign="center" bg={sectionBg} p={4} borderRadius="md">
                <Heading size="md" color={textColor}>
                  Purchase Report
                </Heading>
                <Text fontSize="sm" color={secondaryTextColor}>
                  Period: {startDate} - {endDate}
                </Text>
                <Text fontSize="xs" color={secondaryTextColor} mt={1}>
                  Generated: {new Date().toLocaleDateString('id-ID')} at {new Date().toLocaleTimeString('id-ID')}
                </Text>
              </Box>

              {renderSummaryMetrics()}
              
              {/* Outstanding Payables Status */}
              {data.outstanding_payables !== undefined && (
                <Card>
                  <CardBody>
                    <Flex justify="space-between" align="center" mb={3}>
                      <Heading size="md" color={textColor}>
                        Payables Status
                      </Heading>
                      <Text fontWeight="bold" fontSize="lg" color={data.outstanding_payables < 0 ? 'green.600' : 'orange.600'}>
                        {formatCurrency(Math.abs(data.outstanding_payables))}
                      </Text>
                    </Flex>
                    
                    <SimpleGrid columns={[1, 3]} spacing={4}>
                      <Box textAlign="center" p={3} bg={sectionBg} borderRadius="md">
                        <Text fontSize="sm" color={secondaryTextColor}>Status</Text>
                        <Badge colorScheme={data.outstanding_payables < 0 ? 'green' : 'orange'}>
                          {data.outstanding_payables < 0 ? 'Overpaid (Credit)' : 'Outstanding Amount'}
                        </Badge>
                      </Box>
                      <Box textAlign="center" p={3} bg={sectionBg} borderRadius="md">
                        <Text fontSize="sm" color={secondaryTextColor}>Total Paid</Text>
                        <Text fontWeight="medium" color="green.600">
                          {formatCurrency(data.total_paid || 0)}
                        </Text>
                      </Box>
                      <Box textAlign="center" p={3} bg={sectionBg} borderRadius="md">
                        <Text fontSize="sm" color={secondaryTextColor}>Report Type</Text>
                        <Text fontWeight="medium">SSOT Integration</Text>
                      </Box>
                    </SimpleGrid>
                  </CardBody>
                </Card>
              )}

              {/* Top Vendors Section */}
              {data.purchases_by_vendor && data.purchases_by_vendor.length > 0 && (
                <Box>
                  <Heading size="sm" mb={4} color={textColor}>
                    Top Performing Vendors ({Math.min(data.purchases_by_vendor.length, 6)} vendors)
                  </Heading>
                  <SimpleGrid columns={[1, 2, 3]} spacing={4}>
                    {[...data.purchases_by_vendor]
                      .sort((a, b) => (b.total_amount || 0) - (a.total_amount || 0))
                      .slice(0, 6)
                      .map((vendor: VendorPurchaseSummary, index: number) => (
                        <Box key={index} border="1px" borderColor={borderColor} borderRadius="md" p={4} bg="white">
                          <VStack spacing={3}>
                            <Badge colorScheme="orange" size="lg" variant="solid">
                              #{index + 1}
                            </Badge>
                            <Text fontWeight="bold" fontSize="md" textAlign="center" color="gray.800">
                              {vendor.vendor_name}
                            </Text>
                            <Text fontSize="lg" fontWeight="bold" color="orange.600">
                              {formatCurrency(vendor.total_amount)}
                            </Text>
                            <Text fontSize="sm" color="gray.600">
                              {((vendor.total_amount / (data.total_amount || 1)) * 100).toFixed(1)}% of total
                            </Text>
                            {vendor.total_purchases && (
                              <Text fontSize="xs" color="purple.600">
                                {vendor.total_purchases} purchases
                              </Text>
                            )}
                          </VStack>
                        </Box>
                      ))}
                  </SimpleGrid>
                </Box>
              )}

              {/* Purchases by Vendor Table */}
              {data.purchases_by_vendor && data.purchases_by_vendor.length > 0 && (
                <Box>
                  <Heading size="sm" mb={4} color={textColor}>
                    Purchases by Vendor ({data.purchases_by_vendor.length} vendors)
                  </Heading>
                  
                  {/* Vendor Table Header */}
                  <Box bg="orange.50" p={3} borderRadius="md" mb={2} border="1px solid" borderColor="orange.200">
                    <SimpleGrid columns={[1, 4]} spacing={2} fontSize="sm" fontWeight="bold" color="orange.800">
                      <Text>Vendor</Text>
                      <Text textAlign="right">Total Amount</Text>
                      <Text textAlign="right">Purchases</Text>
                      <Text textAlign="right">Outstanding</Text>
                    </SimpleGrid>
                  </Box>
                  
                  {/* Vendor Rows */}
                  <VStack spacing={2} align="stretch" maxH="400px" overflow="auto">
                    {data.purchases_by_vendor.map((vendor: VendorPurchaseSummary, index: number) => (
                      <Box key={index} border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="white" _hover={{ bg: 'gray.50' }}>
                        <SimpleGrid columns={[1, 4]} spacing={2} fontSize="sm">
                          <VStack align="start" spacing={1}>
                            <Text fontWeight="bold" fontSize="md" color="gray.800">
                              {vendor.vendor_name || 'Unnamed Vendor'}
                            </Text>
                          </VStack>
                          <Text textAlign="right" fontSize="sm" fontWeight="bold" color="orange.600">
                            {formatCurrency(vendor.total_amount || 0)}
                          </Text>
                          <Text textAlign="right" fontSize="sm" fontWeight="medium" color="purple.600">
                            {vendor.total_purchases || 0}
                          </Text>
                          <Text textAlign="right" fontSize="sm" fontWeight="medium" color={vendor.outstanding > 0 ? 'red.600' : vendor.outstanding < 0 ? 'green.600' : 'gray.400'}>
                            {vendor.outstanding !== 0 ? formatCurrency(Math.abs(vendor.outstanding)) : '-'}
                            {vendor.outstanding < 0 && (
                              <Text fontSize="xs" color="green.500" display="block">
                                (Credit)
                              </Text>
                            )}
                          </Text>
                        </SimpleGrid>
                      </Box>
                    ))}
                  </VStack>
                </Box>
              )}

              {/* Payment Analysis */}
              {data.payment_analysis && (
                <Card>
                  <CardHeader>
                    <Heading size="sm">Payment Analysis</Heading>
                  </CardHeader>
                  <CardBody>
                    <SimpleGrid columns={[1, 2, 4]} spacing={4}>
                      <Box textAlign="center">
                        <Text fontSize="sm" color={secondaryTextColor}>Cash Purchases</Text>
                        <Text fontSize="2xl" fontWeight="bold" color="green.600">
                          {data.payment_analysis.cash_purchases || 0}
                        </Text>
                        <Text fontSize="sm" color="green.600">
                          {formatCurrency(data.payment_analysis.cash_amount || 0)}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color={secondaryTextColor}>Credit Purchases</Text>
                        <Text fontSize="2xl" fontWeight="bold" color="orange.600">
                          {data.payment_analysis.credit_purchases || 0}
                        </Text>
                        <Text fontSize="sm" color="orange.600">
                          {formatCurrency(data.payment_analysis.credit_amount || 0)}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color={secondaryTextColor}>Cash Percentage</Text>
                        <Text fontSize="2xl" fontWeight="bold" color="blue.600">
                          {(data.payment_analysis.cash_percentage || 0).toFixed(1)}%
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color={secondaryTextColor}>Average Order Value</Text>
                        <Text fontSize="2xl" fontWeight="bold" color="purple.600">
                          {formatCurrency(data.payment_analysis.average_order_value || 0)}
                        </Text>
                      </Box>
                    </SimpleGrid>
                  </CardBody>
                </Card>
              )}

              {/* Tax Analysis */}
              {data.tax_analysis && (
                <Card>
                  <CardHeader>
                    <Heading size="sm">Tax Analysis</Heading>
                  </CardHeader>
                  <CardBody>
                    <SimpleGrid columns={[1, 3]} spacing={4}>
                      <Box textAlign="center">
                        <Text fontSize="sm" color={secondaryTextColor}>Taxable Amount</Text>
                        <Text fontSize="2xl" fontWeight="bold" color="blue.600">
                          {formatCurrency(data.tax_analysis.total_taxable_amount || 0)}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color={secondaryTextColor}>Tax Amount</Text>
                        <Text fontSize="2xl" fontWeight="bold" color="purple.600">
                          {formatCurrency(data.tax_analysis.total_tax_amount || 0)}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color={secondaryTextColor}>Average Tax Rate</Text>
                        <Text fontSize="2xl" fontWeight="bold" color="orange.600">
                          {(data.tax_analysis.average_tax_rate || 0).toFixed(2)}%
                        </Text>
                      </Box>
                    </SimpleGrid>
                  </CardBody>
                </Card>
              )}
            </VStack>
          )}
        </ModalBody>

        <ModalFooter>
          <HStack spacing={3}>
            {data && !isLoading && (
              <>
                <Button
                  colorScheme="red"
                  variant="outline"
                  size="sm"
                  leftIcon={<FiFilePlus />}
                  onClick={() => handleExport('pdf')}
                >
                  Export PDF
                </Button>
                <Button
                  colorScheme="green"
                  variant="outline"
                  size="sm"
                  leftIcon={<FiFileText />}
                  onClick={() => handleExport('excel')}
                >
                  Export CSV
                </Button>
              </>
            )}
          </HStack>
          <Button variant="ghost" onClick={onClose}>
            Close
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default PurchaseReportModal;