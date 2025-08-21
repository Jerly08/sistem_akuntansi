'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import {
  Container,
  VStack,
  HStack,
  Heading,
  Text,
  Badge,
  Button,
  Card,
  CardHeader,
  CardBody,
  Divider,
  SimpleGrid,
  Box,
  Spinner,
  Alert,
  AlertIcon,
  useToast,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Flex
} from '@chakra-ui/react';
import { 
  FiArrowLeft,
  FiEdit,
  FiPrinter,
  FiMail,
  FiCheckCircle,
  FiXCircle,
  FiClock,
  FiDollarSign
} from 'react-icons/fi';
import Link from 'next/link';
import { Sale } from '@/services/salesService';
import salesService from '@/services/salesService';
import { useAuth } from '@/contexts/AuthContext';

export default function SaleDetailPage() {
  const params = useParams();
  const id = params?.id as string;
  const toast = useToast();
  const { user } = useAuth();

  const [sale, setSale] = useState<Sale | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [downloadLoading, setDownloadLoading] = useState(false);

  useEffect(() => {
    if (id) {
      loadSale();
    }
  }, [id]);

  const loadSale = async () => {
    try {
      setLoading(true);
      setError(null);
      const saleData = await salesService.getSale(parseInt(id));
      setSale(saleData);
    } catch (error: any) {
      setError(error.message || 'Failed to load sale details');
      toast({
        title: 'Error',
        description: 'Failed to load sale details',
        status: 'error',
        duration: 5000,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleStatusUpdate = async (status: string) => {
    if (!sale) return;

    try {
      setActionLoading(true);
      
      let result;
      switch (status) {
        case 'CONFIRMED':
          result = await salesService.confirmSale(sale.id);
          break;
        case 'INVOICED':
          result = await salesService.createInvoiceFromSale(sale.id);
          break;
        case 'CANCELLED':
          result = await salesService.cancelSale(sale.id);
          break;
        default:
          throw new Error('Invalid status');
      }

      toast({
        title: 'Success',
        description: `Sale ${status.toLowerCase()} successfully`,
        status: 'success',
        duration: 3000,
      });

      // Refresh sale data
      await loadSale();
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || `Failed to ${status.toLowerCase()} sale`,
        status: 'error',
        duration: 5000,
      });
    } finally {
      setActionLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status?.toUpperCase()) {
      case 'DRAFT': return 'gray';
      case 'PENDING': return 'yellow';
      case 'CONFIRMED': return 'green';
      case 'INVOICED': return 'blue';
      case 'PAID': return 'teal';
      case 'CANCELLED': return 'red';
      case 'OVERDUE': return 'red';
      default: return 'gray';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'CONFIRMED': return <FiCheckCircle />;
      case 'INVOICED': return <FiMail />;
      case 'PAID': return <FiDollarSign />;
      case 'CANCELLED': return <FiXCircle />;
      case 'PENDING': return <FiClock />;
      default: return <FiClock />;
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0
    }).format(amount);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('id-ID', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  const handleDownloadPDF = async () => {
    if (!sale?.invoice_number) {
      toast({
        title: 'Error',
        description: 'This sale does not have an invoice number. Please create invoice first.',
        status: 'error',
        duration: 5000,
      });
      return;
    }

    try {
      setDownloadLoading(true);
      await salesService.downloadInvoicePDF(sale.id, sale.invoice_number);
      
      toast({
        title: 'Success',
        description: 'Invoice PDF downloaded successfully',
        status: 'success',
        duration: 3000,
      });
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || 'Failed to download PDF',
        status: 'error',
        duration: 5000,
      });
    } finally {
      setDownloadLoading(false);
    }
  };

  if (loading) {
    return (
      <Container maxW="7xl" py={6}>
        <VStack spacing={6}>
          <Spinner size="xl" />
          <Text>Loading sale details...</Text>
        </VStack>
      </Container>
    );
  }

  if (error || !sale) {
    return (
      <Container maxW="7xl" py={6}>
        <Alert status="error">
          <AlertIcon />
          {error || 'Sale not found'}
        </Alert>
      </Container>
    );
  }

  return (
    <Container maxW="7xl" py={6}>
      <VStack spacing={6} align="stretch">
        {/* Header */}
        <HStack justify="space-between" align="center">
          <HStack spacing={4}>
            <Link href="/sales">
              <Button variant="ghost" leftIcon={<FiArrowLeft />}>
                Back to Sales
              </Button>
            </Link>
            <VStack align="start" spacing={1}>
              <Heading size="lg">Sale #{sale.code}</Heading>
              <HStack>
                <Badge 
                  colorScheme={getStatusColor(sale.status)} 
                  variant="solid"
                  display="flex"
                  alignItems="center"
                  gap={1}
                >
                  {getStatusIcon(sale.status)}
                  {sale.status}
                </Badge>
                <Text fontSize="sm" color="gray.500">
                  Created: {formatDate(sale.date)}
                </Text>
              </HStack>
            </VStack>
          </HStack>

          <HStack spacing={3}>
            <Button
              leftIcon={<FiEdit />}
              colorScheme="blue"
              variant="outline"
              as={Link}
              href={`/sales/${sale.id}/edit`}
              isDisabled={sale.status !== 'DRAFT'}
            >
              Edit
            </Button>
            {sale.invoice_number && (
              <Button 
                leftIcon={<FiPrinter />} 
                variant="outline"
                onClick={handleDownloadPDF}
                isLoading={downloadLoading}
              >
                Download PDF
              </Button>
            )}
          </HStack>
        </HStack>

        <Tabs variant="enclosed">
          <TabList>
            <Tab>Details</Tab>
            <Tab>Items</Tab>
            <Tab>History</Tab>
          </TabList>

          <TabPanels>
            {/* Details Tab */}
            <TabPanel>
              <SimpleGrid columns={{ base: 1, md: 2 }} spacing={6}>
                {/* Customer Information */}
                <Card>
                  <CardHeader>
                    <Heading size="md">Customer Information</Heading>
                  </CardHeader>
                  <CardBody>
                    <VStack spacing={3} align="stretch">
                      <Box>
                        <Text fontWeight="medium" mb={1}>Customer</Text>
                        <Text>{sale.customer?.name || 'N/A'}</Text>
                      </Box>
                      {sale.customer?.address && (
                        <Box>
                          <Text fontWeight="medium" mb={1}>Address</Text>
                          <Text fontSize="sm" color="gray.600">{sale.customer.address}</Text>
                        </Box>
                      )}
                      {sale.customer?.phone && (
                        <Box>
                          <Text fontWeight="medium" mb={1}>Phone</Text>
                          <Text fontSize="sm">{sale.customer.phone}</Text>
                        </Box>
                      )}
                    </VStack>
                  </CardBody>
                </Card>

                {/* Sale Information */}
                <Card>
                  <CardHeader>
                    <Heading size="md">Sale Information</Heading>
                  </CardHeader>
                  <CardBody>
                    <VStack spacing={3} align="stretch">
                      <Box>
                        <Text fontWeight="medium" mb={1}>Sales Person</Text>
                        <Text>{sale.sales_person?.name || 'N/A'}</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="medium" mb={1}>Due Date</Text>
                        <Text>{sale.due_date ? formatDate(sale.due_date) : 'N/A'}</Text>
                      </Box>
                      <Box>
                        <Text fontWeight="medium" mb={1}>Payment Terms</Text>
                        <Text>{sale.payment_terms || 'N/A'}</Text>
                      </Box>
                      {sale.notes && (
                        <Box>
                          <Text fontWeight="medium" mb={1}>Notes</Text>
                          <Text fontSize="sm" color="gray.600">{sale.notes}</Text>
                        </Box>
                      )}
                    </VStack>
                  </CardBody>
                </Card>
              </SimpleGrid>

              {/* Financial Summary */}
              <Card mt={6}>
                <CardHeader>
                  <Heading size="md">Financial Summary</Heading>
                </CardHeader>
                <CardBody>
                  <SimpleGrid columns={{ base: 1, md: 4 }} spacing={6}>
                    <Box>
                      <Text fontWeight="medium" mb={1}>Subtotal</Text>
                      <Text fontSize="xl">{formatCurrency(sale.sub_total || sale.subtotal || 0)}</Text>
                    </Box>
                    <Box>
                      <Text fontWeight="medium" mb={1}>Discount</Text>
                      <Text fontSize="xl" color="orange.500">
                        -{formatCurrency(sale.discount || sale.discount_amount || 0)}
                      </Text>
                    </Box>
                    <Box>
                      <Text fontWeight="medium" mb={1}>Tax</Text>
                      <Text fontSize="xl">{formatCurrency(sale.tax || sale.total_tax || 0)}</Text>
                    </Box>
                    <Box>
                      <Text fontWeight="medium" mb={1}>Total</Text>
                      <Text fontSize="2xl" fontWeight="bold" color="green.600">
                        {formatCurrency(sale.total_amount || 0)}
                      </Text>
                    </Box>
                  </SimpleGrid>
                </CardBody>
              </Card>

              {/* Action Buttons */}
              <Card mt={6}>
                <CardBody>
                  <HStack spacing={4} wrap="wrap">
                    {sale.status === 'DRAFT' && (
                      <Button
                        colorScheme="green"
                        onClick={() => handleStatusUpdate('CONFIRMED')}
                        isLoading={actionLoading}
                        leftIcon={<FiCheckCircle />}
                      >
                        Confirm Sale
                      </Button>
                    )}
                    
                    {sale.status === 'CONFIRMED' && (
                      <Button
                        colorScheme="blue"
                        onClick={() => handleStatusUpdate('INVOICED')}
                        isLoading={actionLoading}
                        leftIcon={<FiMail />}
                      >
                        Create Invoice
                      </Button>
                    )}
                    
                    {['DRAFT', 'PENDING', 'CONFIRMED'].includes(sale.status) && (
                      <Button
                        colorScheme="red"
                        variant="outline"
                        onClick={() => handleStatusUpdate('CANCELLED')}
                        isLoading={actionLoading}
                        leftIcon={<FiXCircle />}
                      >
                        Cancel Sale
                      </Button>
                    )}
                  </HStack>
                </CardBody>
              </Card>
            </TabPanel>

            {/* Items Tab */}
            <TabPanel>
              <Card>
                <CardHeader>
                  <Heading size="md">Sale Items</Heading>
                </CardHeader>
                <CardBody>
                  {((sale.items && sale.items.length > 0) || (sale.sale_items && sale.sale_items.length > 0)) ? (
                    <VStack spacing={4} align="stretch">
                      {(sale.items || sale.sale_items || []).map((item, index) => (
                        <Box key={index} p={4} border="1px" borderColor="gray.200" borderRadius="md">
                          <Flex justify="space-between" align="center">
                            <VStack align="start" spacing={1}>
                              <Text fontWeight="medium">{item.product?.name || 'Unknown Product'}</Text>
                              <Text fontSize="sm" color="gray.600">
                                Qty: {item.quantity} × {formatCurrency(item.unit_price)}
                              </Text>
                              {item.description && (
                                <Text fontSize="xs" color="gray.500">
                                  {item.description}
                                </Text>
                              )}
                            </VStack>
                            <Text fontWeight="medium">{formatCurrency(item.total_price || item.line_total || 0)}</Text>
                          </Flex>
                        </Box>
                      ))}
                    </VStack>
                  ) : (
                    <Text color="gray.500">No items found</Text>
                  )}
                </CardBody>
              </Card>
            </TabPanel>


            {/* History Tab */}
            <TabPanel>
              <Card>
                <CardHeader>
                  <Heading size="md">Sale History</Heading>
                </CardHeader>
                <CardBody>
                  <Text color="gray.500">History tracking will be implemented here</Text>
                </CardBody>
              </Card>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </VStack>
    </Container>
  );
}
