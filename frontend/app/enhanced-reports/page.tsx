'use client';

import React, { useState } from 'react';
import SimpleLayout from '@/components/layout/SimpleLayout';
import { useTranslation } from '@/hooks/useTranslation';
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
  Badge,
  FormControl,
  FormLabel,
  Input,
  Select,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  useDisclosure,
  useColorModeValue
} from '@chakra-ui/react';
import { 
  FiTrendingUp, 
  FiBarChart, 
  FiActivity,
  FiList,
  FiBook,
  FiEye,
  FiRefreshCw
} from 'react-icons/fi';
import EnhancedFinancialReportViewer from '../../src/components/reports/EnhancedFinancialReportViewer';
import { reportService, ReportParameters } from '../../src/services/reportService';

// Sample reports with enhanced drill-down capability
const getEnhancedReports = (t: any) => [
  {
    id: 'profit-loss',
    name: 'Enhanced Profit & Loss',
    description: 'Comprehensive P&L statement with journal entry drill-down capability',
    type: 'FINANCIAL',
    icon: FiTrendingUp,
    requiresDateRange: true
  },
  {
    id: 'balance-sheet',
    name: 'Enhanced Balance Sheet',
    description: 'Interactive balance sheet with clickable line items for journal entry analysis',
    type: 'FINANCIAL', 
    icon: FiBarChart,
    requiresAsOfDate: true
  },
  {
    id: 'trial-balance',
    name: 'Enhanced Trial Balance',
    description: 'Trial balance with account-level drill-down to journal entries',
    type: 'FINANCIAL',
    icon: FiList,
    requiresAsOfDate: true
  },
  {
    id: 'cash-flow',
    name: 'Enhanced Cash Flow',
    description: 'Cash flow statement with transaction-level drill-down (Coming Soon)',
    type: 'FINANCIAL',
    icon: FiActivity,
    requiresDateRange: true,
    disabled: true
  },
  {
    id: 'general-ledger',
    name: 'Enhanced General Ledger',
    description: 'General ledger with detailed journal entry analysis (Coming Soon)',
    type: 'FINANCIAL',
    icon: FiBook,
    requiresDateRange: true,
    disabled: true
  }
];

const EnhancedReportsPage: React.FC = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [selectedReport, setSelectedReport] = useState<any>(null);
  const [reportParams, setReportParams] = useState<ReportParameters>({});
  const [reportData, setReportData] = useState<any>(null);
  const [currentReportType, setCurrentReportType] = useState<string>('');
  const { isOpen, onOpen, onClose } = useDisclosure();
  const toast = useToast();

  // Color mode values
  const cardBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const headingColor = useColorModeValue('gray.700', 'white');
  const textColor = useColorModeValue('gray.800', 'white');
  const descriptionColor = useColorModeValue('gray.600', 'gray.300');

  const enhancedReports = getEnhancedReports(t);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setReportParams(prev => ({ ...prev, [name]: value }));
  };

  const handleGenerateReport = (report: any) => {
    setSelectedReport(report);
    // Set default parameters
    if (report.requiresDateRange) {
      const today = new Date();
      const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
      setReportParams({
        start_date: firstDayOfMonth.toISOString().split('T')[0],
        end_date: today.toISOString().split('T')[0],
        format: 'json'
      });
    } else if (report.requiresAsOfDate) {
      setReportParams({
        as_of_date: new Date().toISOString().split('T')[0],
        format: 'json'
      });
    }
    onOpen();
  };

  const executeReport = async () => {
    if (!selectedReport) return;

    setLoading(true);
    try {
      const result = await reportService.generateReport(selectedReport.id, reportParams);
      
      // Convert API response to format expected by EnhancedFinancialReportViewer
      setReportData(result);
      setCurrentReportType(selectedReport.id.toUpperCase().replace('-', '_'));
      
      toast({
        title: 'Report Generated',
        description: `${selectedReport.name} has been generated successfully.`,
        status: 'success',
        duration: 5000,
        isClosable: true,
      });
      
      onClose();
    } catch (error) {
      console.error('Failed to generate report:', error);
      
      toast({
        title: 'Report Generation Failed',
        description: 'There was an error generating the report. Please try again.',
        status: 'error',
        duration: 8000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const clearReport = () => {
    setReportData(null);
    setCurrentReportType('');
  };

  return (
    <SimpleLayout allowedRoles={['admin', 'finance', 'director', 'inventory_manager']}>
      <Box p={8}>
        <VStack spacing={8} align="stretch">
          {/* Header */}
          <VStack align="start" spacing={4}>
            <Heading as="h1" size="xl" color={headingColor} fontWeight="medium">
              Enhanced Financial Reports
            </Heading>
            <Text color={descriptionColor} fontSize="lg">
              Interactive financial reports with journal entry drill-down functionality. 
              Click on any line item to see the underlying journal entries.
            </Text>
            {reportData && (
              <Button 
                leftIcon={<FiRefreshCw />} 
                onClick={clearReport} 
                variant="outline" 
                colorScheme="blue"
              >
                View All Reports
              </Button>
            )}
          </VStack>
          
          {/* Report Display or Report Grid */}
          {reportData ? (
            <Card border="1px" borderColor={borderColor}>
              <CardBody>
                <EnhancedFinancialReportViewer
                  reportType={currentReportType}
                  reportData={reportData}
                  startDate={reportParams.start_date}
                  endDate={reportParams.end_date}
                  asOfDate={reportParams.as_of_date}
                />
              </CardBody>
            </Card>
          ) : (
            /* Enhanced Reports Grid */
            <SimpleGrid columns={[1, 2, 3]} spacing={6}>
              {enhancedReports.map((report) => (
                <Card
                  key={report.id}
                  bg={cardBg}
                  border="1px"
                  borderColor={borderColor}
                  borderRadius="md"
                  overflow="hidden"
                  _hover={{ shadow: 'md' }}
                  transition="all 0.2s"
                  opacity={report.disabled ? 0.6 : 1}
                  cursor={report.disabled ? 'not-allowed' : 'pointer'}
                >
                  <CardBody p={0}>
                    <VStack spacing={0} align="stretch">
                      {/* Header */}
                      <HStack p={4} align="center" justify="space-between">
                        <Icon as={report.icon} size="24px" color="blue.500" />
                        <HStack spacing={2}>
                          <Badge 
                            colorScheme="green" 
                            variant="solid"
                            fontSize="xs"
                          >
                            ENHANCED
                          </Badge>
                          {report.disabled && (
                            <Badge colorScheme="gray" variant="outline" fontSize="xs">
                              COMING SOON
                            </Badge>
                          )}
                        </HStack>
                      </HStack>
                      
                      {/* Content */}
                      <VStack spacing={3} align="stretch" px={4} pb={4}>
                        <Heading size="md" color={textColor} fontWeight="medium">
                          {report.name}
                        </Heading>
                        <Text 
                          fontSize="sm" 
                          color={descriptionColor} 
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
                          onClick={() => handleGenerateReport(report)}
                          isDisabled={report.disabled}
                          leftIcon={<FiEye />}
                          mt={2}
                        >
                          {report.disabled ? 'Coming Soon' : 'View Enhanced Report'}
                        </Button>
                      </VStack>
                    </VStack>
                  </CardBody>
                </Card>
              ))}
            </SimpleGrid>
          )}
        </VStack>
      </Box>
      
      {/* Report Parameters Modal */}
      <Modal isOpen={isOpen} onClose={onClose} size="md">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>{selectedReport?.name} Parameters</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            {selectedReport && (
              <VStack spacing={4} align="stretch">
                {selectedReport.requiresAsOfDate && (
                  <FormControl isRequired>
                    <FormLabel>As of Date</FormLabel>
                    <Input 
                      type="date" 
                      name="as_of_date" 
                      value={reportParams.as_of_date || ''} 
                      onChange={handleInputChange} 
                    />
                  </FormControl>
                )}
                
                {selectedReport.requiresDateRange && (
                  <>
                    <FormControl isRequired>
                      <FormLabel>Start Date</FormLabel>
                      <Input 
                        type="date" 
                        name="start_date" 
                        value={reportParams.start_date || ''} 
                        onChange={handleInputChange} 
                      />
                    </FormControl>
                    <FormControl isRequired>
                      <FormLabel>End Date</FormLabel>
                      <Input 
                        type="date" 
                        name="end_date" 
                        value={reportParams.end_date || ''} 
                        onChange={handleInputChange} 
                      />
                    </FormControl>
                  </>
                )}

                <Text fontSize="sm" color={descriptionColor}>
                  This enhanced report includes interactive drill-down functionality. 
                  Click on any line item to view underlying journal entries.
                </Text>
              </VStack>
            )}
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onClose}>
              Cancel
            </Button>
            <Button 
              colorScheme="blue" 
              onClick={executeReport} 
              isLoading={loading}
              leftIcon={<FiEye />}
            >
              Generate Enhanced Report
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </SimpleLayout>
  );
};

export default EnhancedReportsPage;
