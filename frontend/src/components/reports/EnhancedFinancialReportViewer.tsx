'use client';

import React from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  Heading,
  Card,
  CardBody,
  CardHeader,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  SimpleGrid,
  Badge,
  Alert,
  AlertIcon,
  Divider,
  Icon,
  Progress,
  useColorModeValue,
} from '@chakra-ui/react';
import {
  FiTrendingUp,
  FiTrendingDown,
  FiDollarSign,
  FiPieChart,
  FiBarChart,
  FiActivity,
} from 'react-icons/fi';
import financialReportService, {
  ProfitLossStatement,
  BalanceSheet,
  CashFlowStatement,
  TrialBalance,
  GeneralLedger,
  FinancialRatios,
  FinancialHealthScore,
} from '../../services/financialReportService';
import { useJournalDrilldown } from '@/hooks/useJournalDrilldown';
import JournalDrilldownModal from './JournalDrilldownModal';
import JournalDrilldownButton from './JournalDrilldownButton';
import { formatCurrency } from '@/utils/formatters';

interface EnhancedFinancialReportViewerProps {
  reportType: string;
  reportData: any;
  startDate?: string;
  endDate?: string;
  asOfDate?: string;
}

const EnhancedFinancialReportViewer: React.FC<EnhancedFinancialReportViewerProps> = ({
  reportType,
  reportData,
  startDate,
  endDate,
  asOfDate,
}) => {
  const journalDrilldown = useJournalDrilldown();
  const rowHoverBg = useColorModeValue('gray.50', 'gray.700');

  if (!reportData) {
    return (
      <Alert status="info">
        <AlertIcon />
        No report data to display. Please generate a report first.
      </Alert>
    );
  }

  const renderReportContent = () => {
    switch (reportType) {
      case 'PROFIT_LOSS':
        return (
          <EnhancedProfitLossView 
            data={reportData as ProfitLossStatement} 
            journalDrilldown={journalDrilldown}
            startDate={startDate}
            endDate={endDate}
          />
        );
      case 'BALANCE_SHEET':
        return (
          <EnhancedBalanceSheetView 
            data={reportData as BalanceSheet} 
            journalDrilldown={journalDrilldown}
            asOfDate={asOfDate}
          />
        );
      case 'CASH_FLOW':
        return (
          <EnhancedCashFlowView 
            data={reportData as CashFlowStatement} 
            journalDrilldown={journalDrilldown}
            startDate={startDate}
            endDate={endDate}
          />
        );
      case 'TRIAL_BALANCE':
        return (
          <EnhancedTrialBalanceView 
            data={reportData as TrialBalance} 
            journalDrilldown={journalDrilldown}
            asOfDate={asOfDate}
          />
        );
      case 'GENERAL_LEDGER':
        return (
          <EnhancedGeneralLedgerView 
            data={reportData as GeneralLedger} 
            journalDrilldown={journalDrilldown}
            startDate={startDate}
            endDate={endDate}
          />
        );
      default:
        return (
          <Alert status="warning">
            <AlertIcon />
            Report type "{reportType}" is not supported for viewing.
          </Alert>
        );
    }
  };

  return (
    <Box>
      {renderReportContent()}
      
      {/* Journal Drilldown Modal */}
      {journalDrilldown.drilldownRequest && (
        <JournalDrilldownModal
          isOpen={journalDrilldown.isOpen}
          onClose={journalDrilldown.closeDrilldown}
          drilldownRequest={journalDrilldown.drilldownRequest}
          title={journalDrilldown.title}
        />
      )}
    </Box>
  );
};

// Enhanced Profit & Loss Statement Viewer with Drill-down
const EnhancedProfitLossView: React.FC<{ 
  data: ProfitLossStatement; 
  journalDrilldown: any;
  startDate?: string;
  endDate?: string;
}> = ({ data, journalDrilldown, startDate, endDate }) => {
  const getVarianceColor = (variance: number) => {
    if (variance > 0) return 'green.500';
    if (variance < 0) return 'red.500';
    return 'gray.500';
  };

  const handleDrillDown = (accountCodes: string[], lineItemName: string, amount?: number) => {
    if (!startDate || !endDate) {
      console.warn('Start date and end date are required for drill-down');
      return;
    }

    journalDrilldown.drillDownProfitLoss(
      lineItemName,
      accountCodes,
      [],
      startDate,
      endDate
    );
  };

  const renderAccountLineItems = (items: any[], title: string, sectionType: 'revenue' | 'expense') => (
    <Card mb={4}>
      <CardHeader>
        <Heading size="md">{title}</Heading>
      </CardHeader>
      <CardBody>
        <TableContainer>
          <Table size="sm" variant="simple">
            <Thead>
              <Tr>
                <Th>{title}</Th>
                <Th isNumeric>Amount</Th>
                {data.comparative && <Th isNumeric>Previous</Th>}
                {data.comparative && <Th isNumeric>Variance</Th>}
                <Th width="50px">Actions</Th>
              </Tr>
            </Thead>
            <Tbody>
              {items.map((item, index) => (
                <Tr key={index} _hover={{ bg: 'gray.50' }}>
                  <Td fontSize="sm">
                    <Text fontWeight="medium">
                      {item.accountCode} - {item.accountName}
                    </Text>
                  </Td>
                  <Td isNumeric fontSize="sm">
                    <Text 
                      fontWeight="semibold"
                      color={sectionType === 'revenue' ? 'green.600' : 'red.600'}
                    >
                      {formatCurrency(item.balance)}
                    </Text>
                  </Td>
                  {data.comparative && (
                    <>
                      <Td isNumeric fontSize="sm">
                        {formatCurrency(0)} {/* Previous period data */}
                      </Td>
                      <Td isNumeric fontSize="sm">
                        <Text color={getVarianceColor(item.balance)}>
                          {formatCurrency(item.balance)}
                        </Text>
                      </Td>
                    </>
                  )}
                  <Td>
                    <JournalDrilldownButton
                      onClick={() => handleDrillDown([item.accountCode], `${item.accountCode} - ${item.accountName}`, item.balance)}
                      label={`View journal entries for ${item.accountName}`}
                    />
                  </Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        </TableContainer>
      </CardBody>
    </Card>
  );

  return (
    <VStack spacing={6} align="stretch">
      {/* Header */}
      <Card>
        <CardHeader>
          <VStack align="start" spacing={2}>
            <Heading size="lg" display="flex" alignItems="center">
              <Icon as={FiTrendingUp} mr={2} />
              {data.reportHeader.reportTitle}
            </Heading>
            <Text color="gray.600">
              {startDate && endDate ? 
                `${new Date(startDate).toLocaleDateString()} - ${new Date(endDate).toLocaleDateString()}` :
                financialReportService.formatDateRange(data.reportHeader.startDate, data.reportHeader.endDate)
              }
            </Text>
            <Text fontSize="sm" color="gray.500">
              Generated on {financialReportService.formatDate(data.reportHeader.generatedAt)}
            </Text>
          </VStack>
        </CardHeader>
      </Card>

      {/* Key Metrics with Drill-down */}
      <SimpleGrid columns={{ base: 2, md: 4 }} spacing={4}>
        <Card>
          <CardBody>
            <Stat>
              <HStack justify="space-between">
                <VStack align="start" spacing={1}>
                  <StatLabel>Total Revenue</StatLabel>
                  <StatNumber color="green.500">
                    {formatCurrency(data.totalRevenue)}
                  </StatNumber>
                  {data.comparative && (
                    <StatHelpText>
                      <Icon as={FiTrendingUp} /> 
                      {financialReportService.formatPercentage(5.2)} vs previous
                    </StatHelpText>
                  )}
                </VStack>
                <JournalDrilldownButton
                  variant="eye"
                  onClick={() => handleDrillDown(['4000', '4001', '4100'], 'Total Revenue', data.totalRevenue)}
                  label="View revenue journal entries"
                />
              </HStack>
            </Stat>
          </CardBody>
        </Card>
        
        <Card>
          <CardBody>
            <Stat>
              <HStack justify="space-between">
                <VStack align="start" spacing={1}>
                  <StatLabel>Gross Profit</StatLabel>
                  <StatNumber color="blue.500">
                    {formatCurrency(data.grossProfit)}
                  </StatNumber>
                  <StatHelpText>
                    {financialReportService.formatPercentage((data.grossProfit / data.totalRevenue) * 100, 1)} margin
                  </StatHelpText>
                </VStack>
                <JournalDrilldownButton
                  variant="eye"
                  onClick={() => handleDrillDown(['4000', '5000'], 'Gross Profit Analysis', data.grossProfit)}
                  label="Analyze gross profit components"
                />
              </HStack>
            </Stat>
          </CardBody>
        </Card>

        <Card>
          <CardBody>
            <Stat>
              <HStack justify="space-between">
                <VStack align="start" spacing={1}>
                  <StatLabel>Total Expenses</StatLabel>
                  <StatNumber color="red.500">
                    {formatCurrency(data.totalExpenses)}
                  </StatNumber>
                </VStack>
                <JournalDrilldownButton
                  variant="eye"
                  onClick={() => handleDrillDown(['6000', '6001', '6100'], 'Total Expenses', data.totalExpenses)}
                  label="View expense journal entries"
                />
              </HStack>
            </Stat>
          </CardBody>
        </Card>

        <Card>
          <CardBody>
            <Stat>
              <HStack justify="space-between">
                <VStack align="start" spacing={1}>
                  <StatLabel>Net Income</StatLabel>
                  <StatNumber color={data.netIncome >= 0 ? 'green.500' : 'red.500'}>
                    {formatCurrency(data.netIncome)}
                  </StatNumber>
                  <StatHelpText>
                    {financialReportService.formatPercentage((data.netIncome / data.totalRevenue) * 100, 1)} margin
                  </StatHelpText>
                </VStack>
                <JournalDrilldownButton
                  variant="eye"
                  onClick={() => handleDrillDown([], 'Net Income Analysis', data.netIncome)}
                  label="Analyze net income components"
                />
              </HStack>
            </Stat>
          </CardBody>
        </Card>
      </SimpleGrid>

      {/* Revenue Section */}
      {data.revenue && data.revenue.length > 0 && (
        renderAccountLineItems(data.revenue, 'Revenue', 'revenue')
      )}

      {/* COGS Section */}
      {data.costOfGoodsSold && data.costOfGoodsSold.length > 0 && (
        renderAccountLineItems(data.costOfGoodsSold, 'Cost of Goods Sold', 'expense')
      )}

      {/* Operating Expenses Section */}
      {data.operatingExpenses && data.operatingExpenses.length > 0 && (
        renderAccountLineItems(data.operatingExpenses, 'Operating Expenses', 'expense')
      )}

      {/* Other Income/Expenses Section */}
      {data.otherIncomeExpenses && data.otherIncomeExpenses.length > 0 && (
        renderAccountLineItems(data.otherIncomeExpenses, 'Other Income & Expenses', 'expense')
      )}
    </VStack>
  );
};

// Enhanced Balance Sheet View with Drill-down
const EnhancedBalanceSheetView: React.FC<{ 
  data: BalanceSheet; 
  journalDrilldown: any;
  asOfDate?: string;
}> = ({ data, journalDrilldown, asOfDate }) => {
  const handleDrillDown = (accountCodes: string[], lineItemName: string, amount?: number) => {
    const dateToUse = asOfDate || new Date().toISOString().split('T')[0];
    
    journalDrilldown.drillDownBalanceSheet(
      lineItemName,
      accountCodes,
      [],
      dateToUse
    );
  };

  const renderBalanceSheetSection = (items: any[], title: string, sectionType: 'asset' | 'liability' | 'equity') => (
    <Card mb={4}>
      <CardHeader>
        <Heading size="md">{title}</Heading>
      </CardHeader>
      <CardBody>
        <TableContainer>
          <Table size="sm" variant="simple">
            <Thead>
              <Tr>
                <Th>{title}</Th>
                <Th isNumeric>Amount</Th>
                <Th width="50px">Actions</Th>
              </Tr>
            </Thead>
            <Tbody>
              {items.map((item, index) => (
                <Tr key={index} _hover={{ bg: 'gray.50' }}>
                  <Td fontSize="sm">
                    <Text fontWeight="medium">
                      {item.accountCode} - {item.accountName}
                    </Text>
                  </Td>
                  <Td isNumeric fontSize="sm">
                    <Text 
                      fontWeight="semibold"
                      color={
                        sectionType === 'asset' ? 'green.600' : 
                        sectionType === 'liability' ? 'red.600' : 
                        'blue.600'
                      }
                    >
                      {formatCurrency(item.balance)}
                    </Text>
                  </Td>
                  <Td>
                    <JournalDrilldownButton
                      onClick={() => handleDrillDown([item.accountCode], `${item.accountCode} - ${item.accountName}`, item.balance)}
                      label={`View journal entries for ${item.accountName}`}
                    />
                  </Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        </TableContainer>
      </CardBody>
    </Card>
  );

  return (
    <VStack spacing={6} align="stretch">
      {/* Header */}
      <Card>
        <CardHeader>
          <VStack align="start" spacing={2}>
            <Heading size="lg" display="flex" alignItems="center">
              <Icon as={FiPieChart} mr={2} />
              {data.reportHeader.reportTitle}
            </Heading>
            <Text color="gray.600">
              As of {asOfDate ? new Date(asOfDate).toLocaleDateString() : new Date().toLocaleDateString()}
            </Text>
            <Text fontSize="sm" color="gray.500">
              Generated on {financialReportService.formatDate(data.reportHeader.generatedAt)}
            </Text>
          </VStack>
        </CardHeader>
      </Card>

      {/* Assets Section */}
      <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6}>
        <VStack spacing={4}>
          <Heading size="lg" alignSelf="start" color="green.600">Assets</Heading>
          
          {/* Current Assets */}
          {data.currentAssets && data.currentAssets.length > 0 && (
            renderBalanceSheetSection(data.currentAssets, 'Current Assets', 'asset')
          )}

          {/* Non-Current Assets */}
          {data.nonCurrentAssets && data.nonCurrentAssets.length > 0 && (
            renderBalanceSheetSection(data.nonCurrentAssets, 'Non-Current Assets', 'asset')
          )}
        </VStack>

        <VStack spacing={4}>
          <Heading size="lg" alignSelf="start" color="red.600">Liabilities & Equity</Heading>
          
          {/* Current Liabilities */}
          {data.currentLiabilities && data.currentLiabilities.length > 0 && (
            renderBalanceSheetSection(data.currentLiabilities, 'Current Liabilities', 'liability')
          )}

          {/* Non-Current Liabilities */}
          {data.nonCurrentLiabilities && data.nonCurrentLiabilities.length > 0 && (
            renderBalanceSheetSection(data.nonCurrentLiabilities, 'Non-Current Liabilities', 'liability')
          )}

          {/* Equity */}
          {data.equity && data.equity.length > 0 && (
            renderBalanceSheetSection(data.equity, 'Equity', 'equity')
          )}
        </VStack>
      </SimpleGrid>

      {/* Balance Verification */}
      <Card>
        <CardBody>
          <HStack justify="space-between" align="center">
            <VStack align="start">
              <Text fontWeight="bold" fontSize="lg">Balance Verification</Text>
              <Text fontSize="sm" color="gray.600">
                Assets should equal Liabilities + Equity
              </Text>
            </VStack>
            <VStack align="end">
              <HStack spacing={4}>
                <Text>Total Assets: <Text as="span" fontWeight="bold" color="green.600">{formatCurrency(data.totalAssets)}</Text></Text>
                <Text>Total Liab. + Equity: <Text as="span" fontWeight="bold" color="red.600">{formatCurrency(data.totalLiabilitiesAndEquity)}</Text></Text>
              </HStack>
              <Badge colorScheme={Math.abs(data.totalAssets - data.totalLiabilitiesAndEquity) < 0.01 ? 'green' : 'red'}>
                {Math.abs(data.totalAssets - data.totalLiabilitiesAndEquity) < 0.01 ? 'BALANCED' : 'UNBALANCED'}
              </Badge>
            </VStack>
          </HStack>
        </CardBody>
      </Card>
    </VStack>
  );
};

// Enhanced Trial Balance View with Drill-down
const EnhancedTrialBalanceView: React.FC<{ 
  data: TrialBalance; 
  journalDrilldown: any;
  asOfDate?: string;
}> = ({ data, journalDrilldown, asOfDate }) => {
  const handleDrillDown = (accountCode: string, accountName: string, balance?: number) => {
    const dateToUse = asOfDate || new Date().toISOString().split('T')[0];
    const startDate = new Date(dateToUse);
    startDate.setFullYear(startDate.getFullYear() - 1); // One year range

    journalDrilldown.drillDownAccount(
      accountCode,
      accountName,
      startDate.toISOString().split('T')[0],
      dateToUse,
      'TRIAL_BALANCE'
    );
  };

  return (
    <VStack spacing={6} align="stretch">
      {/* Header */}
      <Card>
        <CardHeader>
          <VStack align="start" spacing={2}>
            <Heading size="lg" display="flex" alignItems="center">
              <Icon as={FiBarChart} mr={2} />
              {data.reportHeader.reportTitle}
            </Heading>
            <Text color="gray.600">
              As of {asOfDate ? new Date(asOfDate).toLocaleDateString() : new Date().toLocaleDateString()}
            </Text>
            <Text fontSize="sm" color="gray.500">
              Generated on {financialReportService.formatDate(data.reportHeader.generatedAt)}
            </Text>
          </VStack>
        </CardHeader>
      </Card>

      {/* Trial Balance Table */}
      <Card>
        <CardBody>
          <TableContainer>
            <Table size="sm" variant="simple">
              <Thead>
                <Tr>
                  <Th>Account Code</Th>
                  <Th>Account Name</Th>
                  <Th isNumeric>Debit</Th>
                  <Th isNumeric>Credit</Th>
                  <Th width="50px">Actions</Th>
                </Tr>
              </Thead>
              <Tbody>
                {data.accounts.map((account, index) => (
                  <Tr key={index} _hover={{ bg: 'gray.50' }}>
                    <Td fontSize="sm">
                      <Text fontFamily="mono" fontWeight="medium">
                        {account.accountCode}
                      </Text>
                    </Td>
                    <Td fontSize="sm">
                      <Text fontWeight="medium">{account.accountName}</Text>
                    </Td>
                    <Td isNumeric fontSize="sm">
                      <Text color={account.debitBalance > 0 ? 'green.600' : undefined}>
                        {account.debitBalance > 0 ? formatCurrency(account.debitBalance) : '-'}
                      </Text>
                    </Td>
                    <Td isNumeric fontSize="sm">
                      <Text color={account.creditBalance > 0 ? 'red.600' : undefined}>
                        {account.creditBalance > 0 ? formatCurrency(account.creditBalance) : '-'}
                      </Text>
                    </Td>
                    <Td>
                      <JournalDrilldownButton
                        onClick={() => handleDrillDown(account.accountCode, account.accountName, account.debitBalance || account.creditBalance)}
                        label={`View journal entries for ${account.accountName}`}
                      />
                    </Td>
                  </Tr>
                ))}
              </Tbody>
              <Thead>
                <Tr bg="gray.100">
                  <Th colSpan={2} fontSize="md" fontWeight="bold">TOTALS</Th>
                  <Th isNumeric fontSize="md" fontWeight="bold" color="green.600">
                    {formatCurrency(data.totalDebits)}
                  </Th>
                  <Th isNumeric fontSize="md" fontWeight="bold" color="red.600">
                    {formatCurrency(data.totalCredits)}
                  </Th>
                  <Th></Th>
                </Tr>
              </Thead>
            </Table>
          </TableContainer>
        </CardBody>
      </Card>

      {/* Balance Verification */}
      <Card>
        <CardBody>
          <HStack justify="center" align="center" spacing={6}>
            <VStack>
              <Text fontSize="lg" fontWeight="bold">Balance Check</Text>
              <Badge 
                size="lg" 
                colorScheme={Math.abs(data.totalDebits - data.totalCredits) < 0.01 ? 'green' : 'red'}
              >
                {Math.abs(data.totalDebits - data.totalCredits) < 0.01 ? 'BALANCED' : 'UNBALANCED'}
              </Badge>
            </VStack>
            <Text fontSize="sm" color="gray.600">
              Difference: {formatCurrency(Math.abs(data.totalDebits - data.totalCredits))}
            </Text>
          </HStack>
        </CardBody>
      </Card>
    </VStack>
  );
};

// Placeholder for Enhanced Cash Flow View
const EnhancedCashFlowView: React.FC<{ 
  data: CashFlowStatement; 
  journalDrilldown: any;
  startDate?: string;
  endDate?: string;
}> = ({ data, journalDrilldown, startDate, endDate }) => {
  return (
    <Alert status="info">
      <AlertIcon />
      Enhanced Cash Flow View with journal drilldown is coming soon!
    </Alert>
  );
};

// Placeholder for Enhanced General Ledger View
const EnhancedGeneralLedgerView: React.FC<{ 
  data: GeneralLedger; 
  journalDrilldown: any;
  startDate?: string;
  endDate?: string;
}> = ({ data, journalDrilldown, startDate, endDate }) => {
  return (
    <Alert status="info">
      <AlertIcon />
      Enhanced General Ledger View with journal drilldown is coming soon!
    </Alert>
  );
};

export default EnhancedFinancialReportViewer;
