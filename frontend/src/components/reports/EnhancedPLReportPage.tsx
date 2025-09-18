'use client';

import React, { useState, useCallback } from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  Button,
  Card,
  CardBody,
  useDisclosure,
  useToast,
  FormControl,
  FormLabel,
  Input,
  Grid,
  GridItem,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  StatArrow,
  Badge,
  Divider,
} from '@chakra-ui/react';
import { FiTrendingUp, FiDownload, FiEye, FiDatabase } from 'react-icons/fi';
import { formatCurrency } from '@/utils/formatters';
import { reportService } from '@/services/reportService';
import { enhancedPLService } from '@/services/enhancedPLService';
import { journalIntegrationService } from '@/services/journalIntegrationService';
import EnhancedProfitLossModal from './EnhancedProfitLossModal';
import { JournalDrilldownModal } from './JournalDrilldownModal';

interface EnhancedPLData {
  title: string;
  period: string;
  company: any;
  enhanced: boolean;
  sections: any[];
  financialMetrics: {
    grossProfit: number;
    grossProfitMargin: number;
    operatingIncome: number;
    operatingMargin: number;
    ebitda: number;
    ebitdaMargin: number;
    netIncome: number;
    netIncomeMargin: number;
  };
}

interface JournalDrilldownRequest {
  account_codes?: string[];
  account_ids?: number[];
  start_date: string;
  end_date: string;
  report_type?: string;
  line_item_name?: string;
  min_amount?: number;
  max_amount?: number;
  transaction_types?: string[];
  page: number;
  limit: number;
}

const EnhancedPLReportPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [plData, setPLData] = useState<EnhancedPLData | null>(null);
  const [reportParams, setReportParams] = useState({
    start_date: '',
    end_date: '',
  });
  const [drilldownRequest, setDrilldownRequest] = useState<JournalDrilldownRequest | null>(null);
  
  const { isOpen: isPLModalOpen, onOpen: onPLModalOpen, onClose: onPLModalClose } = useDisclosure();
  const { isOpen: isDrilldownModalOpen, onOpen: onDrilldownModalOpen, onClose: onDrilldownModalClose } = useDisclosure();
  const toast = useToast();

  // Set default date range (current month)
  React.useEffect(() => {
    const today = new Date();
    const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
    setReportParams({
      start_date: firstDayOfMonth.toISOString().split('T')[0],
      end_date: today.toISOString().split('T')[0],
    });
  }, []);

  const generateEnhancedPL = async () => {
    if (!reportParams.start_date || !reportParams.end_date) {
      toast({
        title: 'Missing Parameters',
        description: 'Please provide start date and end date',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    setLoading(true);
    try {
      console.log('Generating Enhanced P&L with params:', reportParams);
      
      // Generate enhanced P&L using the journal integration service
      const enhancedData = await enhancedPLService.generateEnhancedPLFromJournals({
        start_date: reportParams.start_date,
        end_date: reportParams.end_date,
        format: 'json'
      });

      console.log('Enhanced P&L data received:', enhancedData);

      // Convert to the format expected by EnhancedProfitLossModal
      const formattedData: EnhancedPLData = {
        title: 'Enhanced Profit and Loss Statement',
        period: `${new Date(reportParams.start_date).toLocaleDateString()} - ${new Date(reportParams.end_date).toLocaleDateString()}`,
        company: enhancedData.company || { name: 'Your Company' },
        enhanced: true,
        sections: formatPLSections(enhancedData),
        financialMetrics: {
          grossProfit: enhancedData.gross_profit || 0,
          grossProfitMargin: enhancedData.gross_profit_margin || 0,
          operatingIncome: enhancedData.operating_income || 0,
          operatingMargin: enhancedData.operating_margin || 0,
          ebitda: enhancedData.ebitda || 0,
          ebitdaMargin: enhancedData.ebitda_margin || 0,
          netIncome: enhancedData.net_income || 0,
          netIncomeMargin: enhancedData.net_income_margin || 0,
        },
      };

      setPLData(formattedData);
      onPLModalOpen();

    } catch (error) {
      console.error('Error generating enhanced P&L:', error);
      toast({
        title: 'Generation Failed',
        description: error instanceof Error ? error.message : 'Failed to generate enhanced P&L report',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const formatPLSections = (data: any) => {
    const sections = [];

    // Revenue section
    if (data.revenue) {
      const revenueSection = {
        name: 'REVENUE',
        items: [],
        total: data.revenue.total_revenue || 0,
        subsections: []
      };

      // Add subsections for different revenue types
      if (data.revenue.sales_revenue?.items?.length > 0) {
        revenueSection.subsections.push({
          name: 'Sales Revenue',
          items: data.revenue.sales_revenue.items.map((item: any) => ({
            name: `${item.code} - ${item.name}`,
            amount: item.amount,
            accountCode: item.code
          })),
          total: data.revenue.sales_revenue.subtotal
        });
      }

      if (data.revenue.service_revenue?.items?.length > 0) {
        revenueSection.subsections.push({
          name: 'Service Revenue',
          items: data.revenue.service_revenue.items.map((item: any) => ({
            name: `${item.code} - ${item.name}`,
            amount: item.amount,
            accountCode: item.code
          })),
          total: data.revenue.service_revenue.subtotal
        });
      }

      sections.push(revenueSection);
    }

    // Cost of Goods Sold section
    if (data.cost_of_goods_sold) {
      const cogsSection = {
        name: 'COST OF GOODS SOLD',
        items: [],
        total: data.cost_of_goods_sold.total_cogs || 0,
        subsections: []
      };

      if (data.cost_of_goods_sold.direct_materials?.items?.length > 0) {
        cogsSection.subsections.push({
          name: 'Direct Materials',
          items: data.cost_of_goods_sold.direct_materials.items.map((item: any) => ({
            name: `${item.code} - ${item.name}`,
            amount: item.amount,
            accountCode: item.code
          })),
          total: data.cost_of_goods_sold.direct_materials.subtotal
        });
      }

      sections.push(cogsSection);
    }

    // Operating Expenses section
    if (data.operating_expenses) {
      const opexSection = {
        name: 'OPERATING EXPENSES',
        items: [],
        total: data.operating_expenses.total_opex || 0,
        subsections: []
      };

      if (data.operating_expenses.administrative?.items?.length > 0) {
        opexSection.subsections.push({
          name: 'Administrative Expenses',
          items: data.operating_expenses.administrative.items.map((item: any) => ({
            name: `${item.code} - ${item.name}`,
            amount: item.amount,
            accountCode: item.code
          })),
          total: data.operating_expenses.administrative.subtotal
        });
      }

      sections.push(opexSection);
    }

    // Calculated sections
    if (data.gross_profit !== undefined) {
      sections.push({
        name: 'GROSS PROFIT',
        items: [
          { name: 'Gross Profit', amount: data.gross_profit },
          { name: 'Gross Profit Margin', amount: data.gross_profit_margin, isPercentage: true }
        ],
        total: data.gross_profit,
        isCalculated: true
      });
    }

    sections.push({
      name: 'NET INCOME',
      items: [
        { name: 'Operating Income', amount: data.operating_income || 0 },
        { name: 'EBITDA', amount: data.ebitda || 0 },
        { name: 'Net Income', amount: data.net_income || 0 },
        { name: 'Net Income Margin', amount: data.net_income_margin || 0, isPercentage: true }
      ],
      total: data.net_income || 0,
      isCalculated: true
    });

    return sections;
  };

  // Handle journal drilldown from P&L modal
  const handleJournalDrilldown = useCallback((itemName: string, accountCode?: string, amount?: number) => {
    console.log('Journal drilldown requested:', { itemName, accountCode, amount });
    
    if (!reportParams.start_date || !reportParams.end_date) {
      toast({
        title: 'Invalid Date Range',
        description: 'Please ensure the P&L report has valid start and end dates',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    const drilldownReq: JournalDrilldownRequest = {
      start_date: reportParams.start_date,
      end_date: reportParams.end_date,
      report_type: 'PROFIT_LOSS',
      line_item_name: itemName,
      page: 1,
      limit: 50,
    };

    // Add account filter if available
    if (accountCode) {
      drilldownReq.account_codes = [accountCode];
    }

    // Add amount filter if available
    if (amount !== undefined && amount > 0) {
      drilldownReq.min_amount = Math.max(0, amount * 0.01); // Allow 1% variance
      drilldownReq.max_amount = amount * 1.01;
    }

    setDrilldownRequest(drilldownReq);
    onDrilldownModalOpen();
  }, [reportParams, toast, onDrilldownModalOpen]);

  // Handle P&L export
  const handlePLExport = useCallback(async (format: 'pdf' | 'excel') => {
    if (!reportParams.start_date || !reportParams.end_date) return;

    try {
      const result = await reportService.generateProfessionalReport('profit-loss', {
        ...reportParams,
        format: format === 'excel' ? 'csv' : 'pdf'
      });

      if (result instanceof Blob) {
        const fileName = `profit-loss-${new Date().toISOString().split('T')[0]}.${format === 'excel' ? 'csv' : 'pdf'}`;
        await reportService.downloadReport(result, fileName);
        
        toast({
          title: 'Export Successful',
          description: `P&L report exported as ${format.toUpperCase()}`,
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      }
    } catch (error) {
      toast({
        title: 'Export Failed',
        description: error instanceof Error ? error.message : 'Failed to export report',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  }, [reportParams, toast]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setReportParams(prev => ({ ...prev, [name]: value }));
  };

  return (
    <Box p={8}>
      <VStack spacing={8} align="stretch">
        {/* Header */}
        <Box>
          <HStack spacing={4} align="center">
            <Box p={3} bg="blue.100" borderRadius="md">
              <FiTrendingUp size="24px" color="blue.600" />
            </Box>
            <VStack align="start" spacing={1}>
              <Text fontSize="2xl" fontWeight="bold" color="gray.700">
                Enhanced Profit & Loss Report
              </Text>
              <Text fontSize="md" color="gray.600">
                Generate comprehensive P&L statements with journal entry drilldown capabilities
              </Text>
            </VStack>
          </HStack>
        </Box>

        {/* Parameters Card */}
        <Card>
          <CardBody>
            <VStack spacing={4} align="stretch">
              <Text fontSize="lg" fontWeight="semibold">Report Parameters</Text>
              <Grid templateColumns="repeat(2, 1fr)" gap={4}>
                <GridItem>
                  <FormControl isRequired>
                    <FormLabel>Start Date</FormLabel>
                    <Input
                      type="date"
                      name="start_date"
                      value={reportParams.start_date}
                      onChange={handleInputChange}
                    />
                  </FormControl>
                </GridItem>
                <GridItem>
                  <FormControl isRequired>
                    <FormLabel>End Date</FormLabel>
                    <Input
                      type="date"
                      name="end_date"
                      value={reportParams.end_date}
                      onChange={handleInputChange}
                    />
                  </FormControl>
                </GridItem>
              </Grid>
            </VStack>
          </CardBody>
        </Card>

        {/* Action Buttons */}
        <HStack spacing={4}>
          <Button
            colorScheme="blue"
            size="lg"
            leftIcon={<FiEye />}
            onClick={generateEnhancedPL}
            isLoading={loading}
            loadingText="Generating..."
          >
            Generate Enhanced P&L
          </Button>
          <Button
            variant="outline"
            size="lg"
            leftIcon={<FiDownload />}
            onClick={() => handlePLExport('pdf')}
            isDisabled={loading}
          >
            Export PDF
          </Button>
          <Button
            variant="outline"
            size="lg"
            leftIcon={<FiDatabase />}
            onClick={() => handlePLExport('excel')}
            isDisabled={loading}
          >
            Export CSV
          </Button>
        </HStack>

        {/* Quick Stats Preview */}
        {plData?.financialMetrics && (
          <Grid templateColumns="repeat(4, 1fr)" gap={6}>
            <GridItem>
              <Card>
                <CardBody>
                  <Stat>
                    <StatLabel>Gross Profit</StatLabel>
                    <StatNumber color={plData.financialMetrics.grossProfit >= 0 ? 'green.600' : 'red.600'}>
                      {formatCurrency(plData.financialMetrics.grossProfit)}
                    </StatNumber>
                    <StatHelpText>
                      <StatArrow type={plData.financialMetrics.grossProfitMargin >= 0 ? 'increase' : 'decrease'} />
                      {plData.financialMetrics.grossProfitMargin.toFixed(2)}%
                    </StatHelpText>
                  </Stat>
                </CardBody>
              </Card>
            </GridItem>
            <GridItem>
              <Card>
                <CardBody>
                  <Stat>
                    <StatLabel>Operating Income</StatLabel>
                    <StatNumber color={plData.financialMetrics.operatingIncome >= 0 ? 'green.600' : 'red.600'}>
                      {formatCurrency(plData.financialMetrics.operatingIncome)}
                    </StatNumber>
                    <StatHelpText>
                      <StatArrow type={plData.financialMetrics.operatingMargin >= 0 ? 'increase' : 'decrease'} />
                      {plData.financialMetrics.operatingMargin.toFixed(2)}%
                    </StatHelpText>
                  </Stat>
                </CardBody>
              </Card>
            </GridItem>
            <GridItem>
              <Card>
                <CardBody>
                  <Stat>
                    <StatLabel>EBITDA</StatLabel>
                    <StatNumber color={plData.financialMetrics.ebitda >= 0 ? 'green.600' : 'red.600'}>
                      {formatCurrency(plData.financialMetrics.ebitda)}
                    </StatNumber>
                    <StatHelpText>
                      <StatArrow type={plData.financialMetrics.ebitdaMargin >= 0 ? 'increase' : 'decrease'} />
                      {plData.financialMetrics.ebitdaMargin.toFixed(2)}%
                    </StatHelpText>
                  </Stat>
                </CardBody>
              </Card>
            </GridItem>
            <GridItem>
              <Card>
                <CardBody>
                  <Stat>
                    <StatLabel>Net Income</StatLabel>
                    <StatNumber color={plData.financialMetrics.netIncome >= 0 ? 'green.600' : 'red.600'}>
                      {formatCurrency(plData.financialMetrics.netIncome)}
                    </StatNumber>
                    <StatHelpText>
                      <StatArrow type={plData.financialMetrics.netIncomeMargin >= 0 ? 'increase' : 'decrease'} />
                      {plData.financialMetrics.netIncomeMargin.toFixed(2)}%
                    </StatHelpText>
                  </Stat>
                </CardBody>
              </Card>
            </GridItem>
          </Grid>
        )}

        {/* Features */}
        <Card>
          <CardBody>
            <VStack spacing={4} align="stretch">
              <Text fontSize="lg" fontWeight="semibold">Features</Text>
              <VStack spacing={3} align="start">
                <HStack>
                  <Badge colorScheme="green">Enhanced</Badge>
                  <Text fontSize="sm">Journal entry-based calculations for accurate financial metrics</Text>
                </HStack>
                <HStack>
                  <Badge colorScheme="blue">Interactive</Badge>
                  <Text fontSize="sm">Click any line item to drill down to supporting journal entries</Text>
                </HStack>
                <HStack>
                  <Badge colorScheme="purple">Comprehensive</Badge>
                  <Text fontSize="sm">Includes gross profit, operating income, EBITDA, and net income with margins</Text>
                </HStack>
                <HStack>
                  <Badge colorScheme="orange">Export Ready</Badge>
                  <Text fontSize="sm">Export to PDF or CSV with detailed formatting</Text>
                </HStack>
              </VStack>
            </VStack>
          </CardBody>
        </Card>
      </VStack>

      {/* Enhanced P&L Modal */}
      {plData && (
        <EnhancedProfitLossModal
          isOpen={isPLModalOpen}
          onClose={onPLModalClose}
          data={plData}
          onJournalDrilldown={handleJournalDrilldown}
          onExport={handlePLExport}
        />
      )}

      {/* Journal Drilldown Modal */}
      {drilldownRequest && (
        <JournalDrilldownModal
          isOpen={isDrilldownModalOpen}
          onClose={onDrilldownModalClose}
          drilldownRequest={drilldownRequest}
          title={`Journal Entries: ${drilldownRequest.line_item_name || 'Selected Item'}`}
        />
      )}
    </Box>
  );
};

export default EnhancedPLReportPage;