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
  Flex,
  Badge,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  FormControl,
  FormLabel,
  Input,
  Select,
  Spinner,
  useDisclosure,
  useColorModeValue,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription
} from '@chakra-ui/react';
import { 
  FiFileText, 
  FiBarChart, 
  FiTrendingUp, 
  FiShoppingCart, 
  FiActivity,
  FiDownload,
  FiEye,
  FiList,
  FiBook,
  FiSearch
} from 'react-icons/fi';
import { reportService, ReportParameters } from '../../src/services/reportService';
import { useJournalDrilldown } from '../../src/hooks/useJournalDrilldown';
import JournalDrilldownModal from '../../src/components/reports/JournalDrilldownModal';
import JournalDrilldownButton from '../../src/components/reports/JournalDrilldownButton';
import { formatCurrency } from '../../src/utils/formatters';

// Define reports data matching the UI design
const getAvailableReports = (t: any) => [
  {
    id: 'profit-loss',
    name: t('reports.profitLossStatement'),
    description: t('reports.description.profitLoss'),
    type: 'FINANCIAL',
    icon: FiTrendingUp
  },
  {
    id: 'balance-sheet',
    name: t('reports.balanceSheet'),
    description: t('reports.description.balanceSheet'),
    type: 'FINANCIAL', 
    icon: FiBarChart
  },
  {
    id: 'cash-flow',
    name: t('reports.cashFlowStatement'),
    description: t('reports.description.cashFlow'),
    type: 'FINANCIAL',
    icon: FiActivity
  },
  {
    id: 'sales-summary',
    name: t('reports.salesSummaryReport'),
    description: t('reports.description.salesSummary'),
    type: 'OPERATIONAL',
    icon: FiShoppingCart
  },
  {
    id: 'vendor-analysis',
    name: t('reports.vendorAnalysisReport'),
    description: t('reports.description.vendorAnalysis'),
    type: 'OPERATIONAL',
    icon: FiShoppingCart
  },
  {
    id: 'trial-balance',
    name: t('reports.trialBalance'),
    description: t('reports.description.trialBalance') || 'Summary of all account balances to ensure debits equal credits and verify accounting equation',
    type: 'FINANCIAL',
    icon: FiList
  },
  {
    id: 'general-ledger',
    name: t('reports.generalLedger'),
    description: t('reports.description.generalLedger'),
    type: 'FINANCIAL',
    icon: FiBook
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
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [selectedReport, setSelectedReport] = useState<any>(null);
  const [reportParams, setReportParams] = useState<ReportParameters>({});
  const [previewData, setPreviewData] = useState<any>(null);
  const [previewReport, setPreviewReport] = useState<any>(null);
  const { isOpen, onOpen, onClose } = useDisclosure();
  const { isOpen: isPreviewOpen, onOpen: onPreviewOpen, onClose: onPreviewClose } = useDisclosure();
  const toast = useToast();
  const journalDrilldown = useJournalDrilldown();

  // Color mode values
  const cardBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const headingColor = useColorModeValue('gray.700', 'white');
  const textColor = useColorModeValue('gray.800', 'white');
  const descriptionColor = useColorModeValue('gray.600', 'gray.300');
  const modalContentBg = useColorModeValue('white', 'gray.800');
  const modalHeaderBg = useColorModeValue('white', 'gray.800');
  const sectionBorderColor = useColorModeValue('gray.200', 'gray.600');
  const evenRowBg = useColorModeValue('gray.50', 'gray.700');
  const oddRowBg = useColorModeValue('white', 'gray.800');
  const sectionTotalBg = useColorModeValue('blue.50', 'blue.900');
  const sectionTotalBorderColor = useColorModeValue('blue.200', 'blue.700');
  const sectionTotalTextColor = useColorModeValue('blue.700', 'blue.200');
  const summaryBg = useColorModeValue('gray.50', 'gray.700');
  const summaryTextColor = useColorModeValue('gray.500', 'gray.400');
  const loadingTextColor = useColorModeValue('gray.700', 'gray.300');
  const loadingDescColor = useColorModeValue('gray.500', 'gray.400');
  const errorIconColor = useColorModeValue('red.400', 'red.300');
  const errorTextColor = useColorModeValue('red.600', 'red.300');
  const noDataIconColor = useColorModeValue('gray.400', 'gray.500');
  const noDataTextColor = useColorModeValue('gray.500', 'gray.400');
  const previewPeriodTextColor = useColorModeValue('gray.500', 'gray.400');
  
  const availableReports = getAvailableReports(t);

  const resetParams = () => {
    setReportParams({});
  };


  // Helper function to handle journal drill-down for different report types
  const handleJournalDrilldown = (itemName: string, accountCode?: string, amount?: number) => {
    if (!previewReport || !previewData) return;

    const reportId = previewReport.id;
    
    // Extract date parameters from the last used parameters or current preview data
    let startDate: string = '';
    let endDate: string = '';
    let asOfDate: string = '';

    if (reportId === 'balance-sheet' || reportId === 'trial-balance') {
      asOfDate = new Date().toISOString().split('T')[0]; // Default to today
    } else {
      // For P&L, cash flow, etc.
      const today = new Date();
      const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
      startDate = firstDayOfMonth.toISOString().split('T')[0];
      endDate = today.toISOString().split('T')[0];
    }

    // Call appropriate drill-down method based on report type
    switch (reportId) {
      case 'profit-loss':
        journalDrilldown.drillDownProfitLoss(
          itemName,
          accountCode ? [accountCode] : [],
          [],
          startDate,
          endDate
        );
        break;
      case 'balance-sheet':
        journalDrilldown.drillDownBalanceSheet(
          itemName,
          accountCode ? [accountCode] : [],
          [],
          asOfDate
        );
        break;
      case 'trial-balance':
        if (accountCode) {
          journalDrilldown.drillDownAccount(
            accountCode,
            itemName,
            new Date(new Date(asOfDate).getTime() - 365 * 24 * 60 * 60 * 1000).toISOString().split('T')[0], // One year ago
            asOfDate,
            'TRIAL_BALANCE'
          );
        }
        break;
      case 'general-ledger':
        if (accountCode) {
          journalDrilldown.drillDownAccount(
            accountCode,
            itemName,
            startDate,
            endDate,
            'GENERAL_LEDGER'
          );
        }
        break;
      case 'cash-flow':
        journalDrilldown.drillDownCashFlow(
          itemName,
          accountCode ? [accountCode] : [],
          [],
          startDate,
          endDate,
          ['CASH_BANK', 'PAYMENT', 'DEPOSIT', 'WITHDRAWAL']
        );
        break;
      default:
        // Generic drill-down for other reports
        journalDrilldown.openDrilldown({
          start_date: startDate || new Date(new Date().getTime() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
          end_date: endDate || new Date().toISOString().split('T')[0],
          line_item_name: itemName,
          account_codes: accountCode ? [accountCode] : undefined,
          report_type: reportId.toUpperCase(),
          page: 1,
          limit: 20
        }, `Journal Entries - ${itemName}`);
    }
  };

  const handleViewReport = async (report: any) => {
    setLoading(true);
    setPreviewReport(report);
    
    try {
      // Set default parameters for quick view
      let quickViewParams: ReportParameters = { format: 'json' };
      
      if (report.id === 'balance-sheet') {
        quickViewParams = { 
          as_of_date: new Date().toISOString().split('T')[0],
          format: 'json'
        };
      } else if (report.id === 'profit-loss') {
        const today = new Date();
        const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
        quickViewParams = {
          start_date: firstDayOfMonth.toISOString().split('T')[0],
          end_date: today.toISOString().split('T')[0],
          format: 'json'
        };
      } else if (report.id === 'cash-flow') {
        const today = new Date();
        const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
        quickViewParams = {
          start_date: firstDayOfMonth.toISOString().split('T')[0],
          end_date: today.toISOString().split('T')[0],
          format: 'json'
        };
      } else if (report.id === 'sales-summary' || report.id === 'vendor-analysis') {
        const today = new Date();
        const thirtyDaysAgo = new Date(today);
        thirtyDaysAgo.setDate(today.getDate() - 30);
        quickViewParams = {
          start_date: thirtyDaysAgo.toISOString().split('T')[0],
          end_date: today.toISOString().split('T')[0],
          group_by: 'month',
          format: 'json'
        };
      } else if (report.id === 'trial-balance') {
        quickViewParams = { 
          as_of_date: new Date().toISOString().split('T')[0],
          format: 'json'
        };
      } else if (report.id === 'general-ledger') {
        const today = new Date();
        const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
        quickViewParams = {
          start_date: firstDayOfMonth.toISOString().split('T')[0],
          end_date: today.toISOString().split('T')[0],
          format: 'json'
        };
      }
      
      // Get real preview data from API
      const previewData = await reportService.generateReportPreview(report.id, quickViewParams);
      setPreviewData(convertApiDataToPreviewFormat(previewData, report));
      
      // Open preview modal
      onPreviewOpen();
      
    } catch (error) {
      console.error('Failed to load report preview:', error);
      
      // Set error state for preview
      setPreviewData({ 
        error: true, 
        message: error instanceof Error ? error.message : 'Failed to load report data from server'
      });
      
      toast({
        title: 'Preview Error',
        description: `Unable to load preview: ${error instanceof Error ? error.message : 'Server connection failed'}`,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
      
      // Still open preview to show error state
      onPreviewOpen();
    } finally {
      setLoading(false);
    }
  };

  // Convert API response data to preview format - Now handles real data from UnifiedReportController
  const convertApiDataToPreviewFormat = (apiData: any, report: any) => {
    if (!apiData) {
      throw new Error('No data received from API');
    }

    // Handle standardized response structure from UnifiedReportController
    let reportData = apiData;
    if (apiData.data) {
      reportData = apiData.data; // Extract from StandardReportResponse wrapper
    }

    try {
      // Handle different report types based on real API response structure
      switch (report.id) {
        case 'balance-sheet':
          // Handle BalanceSheetData structure from backend
          if (reportData.company || reportData.sections || reportData.assets) {
            const sections = [];
            
            if (reportData.assets) {
              sections.push({
                name: 'ASSETS',
                items: reportData.assets.items || [],
                total: reportData.assets.total || 0
              });
            }
            
            if (reportData.liabilities) {
              sections.push({
                name: 'LIABILITIES', 
                items: reportData.liabilities.items || [],
                total: reportData.liabilities.total || 0
              });
            }
            
            if (reportData.equity) {
              sections.push({
                name: 'EQUITY',
                items: reportData.equity.items || [],
                total: reportData.equity.total || 0
              });
            }
            
            return {
              title: 'Balance Sheet',
              period: reportData.period || `As of ${new Date().toLocaleDateString('id-ID')}`,
              sections
            };
          }
          throw new Error('Invalid balance sheet data structure');

        case 'profit-loss':
          // Handle ProfitLossStatement structure from backend
          console.log('Processing profit-loss data:', reportData);
          
          // Check if it's the new ProfitLossStatement format
          if (reportData.report_header || reportData.revenue || reportData.total_revenue !== undefined) {
            const sections = [];
            
            // Revenue section - handle array format from FinancialReportService
            if (reportData.revenue && Array.isArray(reportData.revenue)) {
              sections.push({
                name: 'REVENUE',
                items: reportData.revenue.map((item: any) => ({
                  name: `${item.account_code || ''} - ${item.account_name || ''}`,
                  amount: item.balance || 0
                })),
                total: reportData.total_revenue || 0
              });
            }
            
            // Cost of Goods Sold section - handle array format
            if (reportData.cost_of_goods_sold && Array.isArray(reportData.cost_of_goods_sold)) {
              sections.push({
                name: 'COST OF GOODS SOLD',
                items: reportData.cost_of_goods_sold.map((item: any) => ({
                  name: `${item.account_code || ''} - ${item.account_name || ''}`,
                  amount: item.balance || 0
                })),
                total: reportData.total_cogs || 0
              });
            }
            
            // Gross Profit section
            if (reportData.gross_profit !== undefined) {
              sections.push({
                name: 'GROSS PROFIT',
                items: [{ name: 'Gross Profit', amount: reportData.gross_profit || 0 }],
                total: reportData.gross_profit || 0
              });
            }
            
            // Operating Expenses section - handle array format
            if (reportData.expenses && Array.isArray(reportData.expenses)) {
              sections.push({
                name: 'OPERATING EXPENSES',
                items: reportData.expenses.map((item: any) => ({
                  name: `${item.account_code || ''} - ${item.account_name || ''}`,
                  amount: item.balance || 0
                })),
                total: reportData.total_expenses || 0
              });
            }
            
            // Net Income section
            sections.push({
              name: 'NET INCOME',
              items: [{ name: 'Net Income', amount: reportData.net_income || 0 }],
              total: reportData.net_income || 0
            });

            // Extract period from report header or generate default
            let period = `${new Date().toLocaleDateString('id-ID')}`;
            if (reportData.report_header) {
              const startDate = new Date(reportData.report_header.start_date || Date.now());
              const endDate = new Date(reportData.report_header.end_date || Date.now());
              period = `${startDate.toLocaleDateString('id-ID')} - ${endDate.toLocaleDateString('id-ID')}`;
            }

            return {
              title: reportData.report_header?.report_title || 'Profit and Loss Statement',
              period,
              sections,
              hasData: sections.some(section => section.items && section.items.length > 0)
            };
          }
          
          // If no valid data structure found, return empty state with error info
          return {
            title: 'Profit and Loss Statement',
            period: `${new Date().toLocaleDateString('id-ID')}`,
            sections: [],
            hasData: false,
            error: true,
            message: 'No financial data available for the selected period. Please check if there are any journal entries posted during this period.'
          };

        case 'cash-flow':
          if (reportData.company && reportData.operating_activities) {
            const sections = [];
            
            if (reportData.operating_activities) {
              sections.push({
                name: 'OPERATING ACTIVITIES',
                items: reportData.operating_activities.items || [],
                total: reportData.operating_activities.total || 0
              });
            }
            
            if (reportData.investing_activities) {
              sections.push({
                name: 'INVESTING ACTIVITIES',
                items: reportData.investing_activities.items || [],
                total: reportData.investing_activities.total || 0
              });
            }
            
            if (reportData.financing_activities) {
              sections.push({
                name: 'FINANCING ACTIVITIES',
                items: reportData.financing_activities.items || [],
                total: reportData.financing_activities.total || 0
              });
            }

            return {
              title: 'Cash Flow Statement',
              period: `${new Date(reportData.start_date || Date.now()).toLocaleDateString('id-ID')} - ${new Date(reportData.end_date || Date.now()).toLocaleDateString('id-ID')}`,
              sections
            };
          }
          throw new Error('Invalid cash flow data structure');

        case 'trial-balance':
          // Handle TrialBalanceData structure from UnifiedReportController
          if (reportData.accounts && Array.isArray(reportData.accounts)) {
            return {
              title: 'Trial Balance',
              period: reportData.period || `As of ${new Date().toLocaleDateString('id-ID')}`,
              sections: [
                {
                  name: 'ACCOUNTS',
                  items: reportData.accounts.map((account: any) => ({
                    name: `${account.account_code || account.code || ''} - ${account.account_name || account.name || ''}`,
                    amount: (account.debit_balance || 0) - (account.credit_balance || 0),
                    debit: account.debit_balance || 0,
                    credit: account.credit_balance || 0
                  })),
                  total: reportData.total_debits || 0,
                  totalDebits: reportData.total_debits || 0,
                  totalCredits: reportData.total_credits || 0,
                  isBalanced: reportData.is_balanced || false
                }
              ],
              isBalanced: reportData.is_balanced || false,
              totalDebits: reportData.total_debits || 0,
              totalCredits: reportData.total_credits || 0,
              hasData: reportData.accounts.length > 0
            };
          }
          
          // If no accounts data, return empty state  
          return {
            title: 'Trial Balance',
            period: reportData.period || `As of ${new Date().toLocaleDateString('id-ID')}`,
            sections: [],
            hasData: false,
            message: 'No accounts found for trial balance'
          };

        case 'general-ledger':
          // Handle both possible data structures from different backend endpoints
          // Check if accounts field exists (from both endpoints)
          if (reportData.accounts && Array.isArray(reportData.accounts)) {
            return {
              title: 'General Ledger',
              period: `${new Date(reportData.start_date || Date.now()).toLocaleDateString('id-ID')} - ${new Date(reportData.end_date || Date.now()).toLocaleDateString('id-ID')}`,
              sections: reportData.accounts.map((account: any) => {
                // Handle different field names for account properties
                const accountCode = account.account_code || account.code || '';
                const accountName = account.account_name || account.name || '';
                
                // Handle different field names for transactions
                const transactions = account.transactions || account.entries || [];
                
                // Handle different field names for balance
                const closingBalance = account.closing_balance || account.ending_balance || 
                                       account.closingBalance || account.endingBalance || 0;
                
                return {
                  name: `${accountCode} - ${accountName}`,
                  items: Array.isArray(transactions) ? transactions.map((txn: any) => ({
                    name: txn.description || 'Transaction',
                    amount: (txn.debit_amount || txn.debit || 0) - (txn.credit_amount || txn.credit || 0),
                    debit: txn.debit_amount || txn.debit || 0,
                    credit: txn.credit_amount || txn.credit || 0,
                    date: txn.date,
                    reference: txn.reference || ''
                  })) : [],
                  total: closingBalance,
                  openingBalance: account.opening_balance || account.openingBalance || 0,
                  totalDebits: account.total_debits || account.totalDebits || 0,
                  totalCredits: account.total_credits || account.totalCredits || 0
                };
              }),
              hasData: reportData.accounts.length > 0
            };
          }
          
          // If no accounts but account_count is 0, return empty state
          if (reportData.account_count === 0) {
            return {
              title: 'General Ledger',
              period: `${new Date(reportData.start_date || Date.now()).toLocaleDateString('id-ID')} - ${new Date(reportData.end_date || Date.now()).toLocaleDateString('id-ID')}`,
              sections: [],
              hasData: false,
              message: 'No transactions found for the selected period'
            };
          }
          
          // Log the structure for debugging
          console.error('General ledger data structure:', reportData);
          throw new Error('Invalid general ledger data structure - accounts field missing or invalid');

        case 'sales-summary':
          // Handle SalesSummaryData structure from UnifiedReportController
          const salesByPeriod = reportData.sales_by_period || [];
          const totalRevenue = reportData.total_revenue || 0;
          
          return {
            title: 'Sales Summary Report',
            period: reportData.period || `${new Date().toLocaleDateString('id-ID')}`,
            sections: [
              {
                name: 'SALES BY PERIOD',
                items: Array.isArray(salesByPeriod) ? salesByPeriod.map((period: any) => ({
                  name: period.period || 'Unknown Period',
                  amount: period.amount || 0
                })) : [],
                total: totalRevenue
              }
            ],
            hasData: salesByPeriod && salesByPeriod.length > 0,
            message: (!salesByPeriod || salesByPeriod.length === 0) ? 'No sales data available for the selected period' : undefined
          };

        case 'vendor-analysis':
          // Handle VendorAnalysisData structure from UnifiedReportController
          const purchasesByPeriod = reportData.purchases_by_period || [];
          const totalPurchases = reportData.total_purchases || 0;
          
          return {
            title: 'Vendor Analysis Report',
            period: reportData.period || `${new Date().toLocaleDateString('id-ID')}`,
            sections: [
              {
                name: 'PURCHASES BY PERIOD',
                items: Array.isArray(purchasesByPeriod) ? purchasesByPeriod.map((period: any) => ({
                  name: period.period || 'Unknown Period',
                  amount: period.amount || 0
                })) : [],
                total: totalPurchases
              }
            ],
            hasData: purchasesByPeriod && purchasesByPeriod.length > 0,
            message: (!purchasesByPeriod || purchasesByPeriod.length === 0) ? 'No purchase data available for the selected period' : undefined
          };

        default:
          throw new Error(`Unsupported report type: ${report.id}`);
      }
    } catch (error) {
      console.error('Failed to convert API data:', error);
      throw error;
    }
  };


  // Format currency for display
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0
    }).format(amount);
  };

  const handleGenerateReport = (report: any) => {
    setSelectedReport(report);
    resetParams();
    
    // Set default parameters based on report type
    if (report.id === 'balance-sheet') {
      setReportParams({ as_of_date: new Date().toISOString().split('T')[0], format: 'pdf' });
    } else if (report.id === 'profit-loss' || report.id === 'cash-flow') {
      // Set default start date to first day of current month and end date to today
      const today = new Date();
      const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
      setReportParams({
        start_date: firstDayOfMonth.toISOString().split('T')[0],
        end_date: today.toISOString().split('T')[0],
        format: 'pdf'
      });
    } else if (report.id === 'sales-summary' || report.id === 'purchase-summary' || report.id === 'vendor-analysis') {
      // Set default start date to 30 days ago and end date to today
      const today = new Date();
      const thirtyDaysAgo = new Date(today);
      thirtyDaysAgo.setDate(today.getDate() - 30);
      setReportParams({
        start_date: thirtyDaysAgo.toISOString().split('T')[0],
        end_date: today.toISOString().split('T')[0],
        group_by: 'month',
        format: 'pdf'
      });
    } else if (report.id === 'trial-balance') {
      setReportParams({ as_of_date: new Date().toISOString().split('T')[0], format: 'pdf' });
    } else if (report.id === 'general-ledger') {
      // Set default start date to first day of current month and end date to today
      const today = new Date();
      const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
      setReportParams({
        start_date: firstDayOfMonth.toISOString().split('T')[0],
        end_date: today.toISOString().split('T')[0],
        format: 'pdf'
      });
    }
    
    onOpen();
  };
  
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setReportParams(prev => ({ ...prev, [name]: value }));
  };
  

  const executeReport = async () => {
    if (!selectedReport) return;
    
    setLoading(true);
    try {
      let result;
      
      // Validate required parameters
      if (['profit-loss', 'cash-flow', 'sales-summary', 'purchase-summary', 'general-ledger'].includes(selectedReport.id)) {
        if (!reportParams.start_date || !reportParams.end_date) {
          throw new Error('Start date and end date are required for this report');
        }
      }
      
      if (['balance-sheet', 'trial-balance'].includes(selectedReport.id)) {
        if (!reportParams.as_of_date) {
          throw new Error('As of date is required for this report');
        }
      }
      
      // Use professional report service for specific reports
      if (["balance-sheet", "profit-loss", "cash-flow", "sales-summary", "purchase-summary"].includes(selectedReport.id)) {
        result = await reportService.generateProfessionalReport(selectedReport.id, reportParams);
      } else {
        result = await reportService.generateReport(selectedReport.id, reportParams);
      }
      
      // Verify result is valid before attempting download
      if (!result) {
        throw new Error('No data received from server');
      }
      
      // Check if result is a Blob (for file downloads)
      if (result instanceof Blob) {
        if (result.size === 0) {
          throw new Error('Empty file received from server');
        }
        
        // Handle the result - Download file
        const fileName = `${selectedReport.id}_report_${new Date().toISOString().split('T')[0]}.${reportParams.format}`;
        await reportService.downloadReport(result, fileName);
        toast({
          title: 'Report Downloaded',
          description: `${selectedReport.name} has been downloaded successfully.`,
          status: 'success',
          duration: 5000,
          isClosable: true,
        });
      } else {
        // If not a Blob, probably an error or unexpected data format
        console.error('Unexpected result format:', typeof result, result);
        throw new Error('Invalid response format from server');
      }
      
      onClose();
    } catch (error) {
      console.error('Failed to generate report:', error);
      
      // More detailed error message
      let errorMessage = 'Unknown error occurred';
      if (error instanceof Error) {
        errorMessage = error.message;
      } else if (typeof error === 'string') {
        errorMessage = error;
      }
      
      toast({
        title: 'Report Generation Failed',
        description: errorMessage,
        status: 'error',
        duration: 8000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <SimpleLayout allowedRoles={['admin', 'finance', 'director', 'inventory_manager']}>
      <Box p={8}>
        <VStack spacing={8} align="stretch">
          <VStack align="start" spacing={4}>
            <Heading as="h1" size="xl" color={headingColor} fontWeight="medium">
              Financial Reports
            </Heading>
            
            {/* Journal Drilldown Feature Banner */}
            <Alert 
              status="info" 
              variant="left-accent" 
              borderRadius="md"
              bg={useColorModeValue('blue.50', 'blue.900')}
              borderColor={useColorModeValue('blue.200', 'blue.600')}
            >
              <AlertIcon color={useColorModeValue('blue.500', 'blue.300')} />
              <Box>
                <AlertTitle fontSize="sm" mb={1}>
                  🆕 New Feature: Journal Entry Drill-down! 
                </AlertTitle>
                <Text fontSize="sm">
                  🔍 Try the <strong>"Try Journal Drilldown"</strong> button on financial reports below, or 
                  📊 visit <strong>Enhanced Reports</strong> in the sidebar for full interactive experience.
                  Click any line item in reports to see underlying journal entries!
                </Text>
              </Box>
            </Alert>
          </VStack>
          
          {/* Financial Reports Grid */}
          <SimpleGrid columns={[1, 2, 3]} spacing={6}>
            {availableReports.map((report) => (
              <Card
                key={report.id}
                bg={cardBg}
                border="1px"
                borderColor={borderColor}
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
                        colorScheme={report.type === 'FINANCIAL' ? 'green' : 'blue'} 
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
                      
                      {/* Action Buttons */}
                      <HStack spacing={2} width="full" mt={2}>
                        <Button
                          colorScheme="gray"
                          variant="outline"
                          size="md"
                          flex="1"
                          onClick={() => handleViewReport(report)}
                          isLoading={loading}
                          leftIcon={<FiEye />}
                        >
                          View
                        </Button>
                        <Button
                          colorScheme="blue"
                          size="md"
                          flex="1"
                          onClick={() => handleGenerateReport(report)}
                          isLoading={loading && selectedReport?.id === report.id}
                          leftIcon={<FiFileText />}
                        >
                          Generate
                        </Button>
                      </HStack>
                      
                      {/* Demo Journal Drilldown Button */}
                      {report.type === 'FINANCIAL' && (
                        <HStack spacing={2} width="full" mt={2} pt={2} borderTop="1px" borderColor={borderColor}>
                          <Text fontSize="xs" color={descriptionColor} flex="1">
                            📊 New Feature:
                          </Text>
                          <Button
                            colorScheme="purple"
                            variant="outline"
                            size="sm"
                            onClick={() => {
                              // Demo journal drilldown dengan data sample
                              const today = new Date();
                              const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
                              
                              if (report.id === 'balance-sheet' || report.id === 'trial-balance') {
                                journalDrilldown.drillDownBalanceSheet(
                                  `Demo ${report.name}`,
                                  ['1000', '2000', '3000'], // Sample account codes
                                  [],
                                  today.toISOString().split('T')[0]
                                );
                              } else {
                                journalDrilldown.drillDownProfitLoss(
                                  `Demo ${report.name}`,
                                  ['4000', '5000', '6000'], // Sample account codes
                                  [],
                                  firstDayOfMonth.toISOString().split('T')[0],
                                  today.toISOString().split('T')[0]
                                );
                              }
                            }}
                            leftIcon={<FiSearch />}
                          >
                            Try Journal Drilldown
                          </Button>
                        </HStack>
                      )}
                    </VStack>
                  </VStack>
                </CardBody>
              </Card>
            ))}
          </SimpleGrid>
        </VStack>
      </Box>
      
      {/* Report Parameters Modal */}
      <Modal isOpen={isOpen} onClose={onClose} size="md">
        <ModalOverlay />
        <ModalContent bg={modalContentBg}>
          <ModalHeader>{selectedReport?.name}</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            {selectedReport && (
              <VStack spacing={4} align="stretch">
                {/* Balance Sheet Parameters */}
                {selectedReport.id === 'balance-sheet' && (
                  <>
                    <FormControl isRequired>
                      <FormLabel>As of Date</FormLabel>
                      <Input 
                        type="date" 
                        name="as_of_date" 
                        value={reportParams.as_of_date || ''} 
                        onChange={handleInputChange} 
                      />
                    </FormControl>
                  </>
                )}
                
                {/* Profit & Loss and Cash Flow Parameters */}
                {(selectedReport.id === 'profit-loss' || selectedReport.id === 'cash-flow') && (
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
                
                {/* Sales Summary, Purchase Summary, and Vendor Analysis Parameters */}
                {(selectedReport.id === 'sales-summary' || selectedReport.id === 'purchase-summary' || selectedReport.id === 'vendor-analysis') && (
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
                    <FormControl>
                      <FormLabel>Group By</FormLabel>
                      <Select 
                        name="group_by" 
                        value={reportParams.group_by || 'month'} 
                        onChange={handleInputChange}
                      >
                        <option value="month">Month</option>
                        <option value="quarter">Quarter</option>
                        <option value="year">Year</option>
                      </Select>
                    </FormControl>
                  </>
                )}
                
                {/* Trial Balance Parameters */}
                {selectedReport.id === 'trial-balance' && (
                  <>
                    <FormControl isRequired>
                      <FormLabel>As of Date</FormLabel>
                      <Input 
                        type="date" 
                        name="as_of_date" 
                        value={reportParams.as_of_date || ''} 
                        onChange={handleInputChange} 
                      />
                    </FormControl>
                  </>
                )}
                
                {/* General Ledger Parameters */}
                {selectedReport.id === 'general-ledger' && (
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
                
                {/* Format selection for all reports */}
                <FormControl>
                  <FormLabel>Format</FormLabel>
                  <Select 
                    name="format" 
                    value={reportParams.format || 'pdf'} 
                    onChange={handleInputChange}
                  >
                    <option value="pdf">Download as PDF</option>
                    <option value="csv">Download as CSV</option>
                  </Select>
                </FormControl>
              </VStack>
            )}
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onClose} isDisabled={loading}>
              Cancel
            </Button>
            <Button 
              colorScheme="blue" 
              onClick={executeReport} 
              isLoading={loading}
              leftIcon={<FiDownload />}
            >
              Download
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* Report Preview Modal */}
      <Modal isOpen={isPreviewOpen} onClose={onPreviewClose} size="6xl">
        <ModalOverlay />
        <ModalContent bg={modalContentBg}>
          <ModalHeader>
            <HStack>
              <Icon as={previewReport?.icon || FiFileText} color="blue.500" />
              <VStack align="start" spacing={0}>
                <Text fontSize="lg" fontWeight="bold">
                  {previewData?.title || previewReport?.name}
                </Text>
                <Text fontSize="sm" color={previewPeriodTextColor}>
                  {previewData?.period}
                </Text>
              </VStack>
            </HStack>
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            {loading ? (
              <Box textAlign="center" py={8}>
                <VStack spacing={4}>
                  <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
                  <VStack spacing={2}>
                    <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                      Generating Report Preview
                    </Text>
                    <Text fontSize="sm" color={loadingDescColor}>
                      Please wait while we fetch real data from the database...
                    </Text>
                  </VStack>
                </VStack>
              </Box>
            ) : previewData ? (
              previewData.error || !previewData.hasData ? (
                // Error State or Empty State
                <Box textAlign="center" py={12}>
                  <VStack spacing={4}>
                    <Icon as={FiFileText} boxSize={12} color={previewData.error ? errorIconColor : noDataIconColor} />
                    <VStack spacing={2}>
                      <Text fontSize="lg" fontWeight="medium" color={previewData.error ? errorTextColor : noDataTextColor}>
                        {previewData.error ? 'Unable to Load Preview' : 'No Data Available'}
                      </Text>
                      <Text fontSize="sm" color={summaryTextColor} maxW="md" textAlign="center">
                        {previewData.message || (previewData.error ? 'There was a problem loading the report data. Please try again or generate the full report.' : 'No data found for the selected period or criteria.')}
                      </Text>
                    </VStack>
                    <Button 
                      colorScheme="blue" 
                      variant="outline"
                      onClick={() => {
                        onPreviewClose();
                        if (previewData.error) {
                          handleViewReport(previewReport);
                        } else {
                          handleGenerateReport(previewReport);
                        }
                      }}
                    >
                      {previewData.error ? 'Retry Preview' : 'Generate Full Report'}
                    </Button>
                  </VStack>
                </Box>
              ) : (
                <VStack spacing={6} align="stretch">
                  {previewData.sections?.map((section: any, sectionIndex: number) => (
                    <Box key={sectionIndex}>
                      <Heading size="md" color={headingColor} mb={4} borderBottom="2px" borderColor={sectionBorderColor} pb={2}>
                        {section.name}
                      </Heading>
                      <VStack spacing={2} align="stretch">
                        {section.items?.map((item: any, itemIndex: number) => (
                          <HStack key={itemIndex} justify="space-between" py={2} px={4} 
                                 bg={itemIndex % 2 === 0 ? evenRowBg : oddRowBg} 
                                 borderRadius="md"
                                 _hover={{ bg: useColorModeValue('blue.50', 'blue.900') }}
                                 transition="background 0.2s">
                            <HStack spacing={3} flex={1}>
                              <Text fontSize="sm" color={textColor} flex={1}>
                                {item.name}
                              </Text>
                              <Text fontSize="sm" fontWeight="medium" 
                                    color={item.amount >= 0 ? "black" : "red.500"} minW="120px" textAlign="right">
                                {formatCurrency(item.amount)}
                              </Text>
                            </HStack>
                            <Box>
                              <JournalDrilldownButton
                                onClick={() => {
                                  // Extract account code from item name if available (e.g., "1000 - Cash" or "4000-Revenue")
                                  const accountCodeMatch = item.name.match(/^([0-9A-Z-]+)\s*[-\s]/);
                                  const accountCode = accountCodeMatch ? accountCodeMatch[1] : undefined;
                                  handleJournalDrilldown(item.name, accountCode, item.amount);
                                }}
                                size="xs"
                                label={`View journal entries for ${item.name}`}
                                variant="search"
                              />
                            </Box>
                          </HStack>
                        ))}
                        
                        {/* Section Total */}
                        <HStack justify="space-between" py={3} px={4} 
                               bg={sectionTotalBg} borderRadius="md" 
                               borderTop="2px" borderColor={sectionTotalBorderColor} mt={2}>
                          <HStack spacing={3} flex={1}>
                            <Text fontSize="md" fontWeight="bold" color={sectionTotalTextColor} flex={1}>
                              Total {section.name}
                            </Text>
                            <Text fontSize="md" fontWeight="bold" 
                                  color={section.total >= 0 ? sectionTotalTextColor : "red.500"} minW="120px" textAlign="right">
                              {formatCurrency(section.total)}
                            </Text>
                          </HStack>
                          <Box>
                            <JournalDrilldownButton
                              onClick={() => {
                                // For section totals, we use the section name for drill-down
                                handleJournalDrilldown(`Total ${section.name}`, undefined, section.total);
                              }}
                              size="xs"
                              label={`View journal entries for Total ${section.name}`}
                              variant="eye"
                            />
                          </Box>
                        </HStack>
                      </VStack>
                    </Box>
                  ))}
                  
                  {/* Report Summary */}
                  <Box bg={summaryBg} p={4} borderRadius="md" mt={4}>
                    <Text fontSize="xs" color={summaryTextColor} textAlign="center">
                      This is a preview of the report. For detailed and up-to-date information, 
                      please generate the full report using the "Generate" button.
                    </Text>
                  </Box>
                </VStack>
              )
            ) : (
              <Box textAlign="center" py={8}>
                <VStack spacing={3}>
                  <Icon as={FiFileText} boxSize={8} color={noDataIconColor} />
                  <Text color={noDataTextColor}>No preview data available</Text>
                  <Text fontSize="sm" color={summaryTextColor}>
                    Please try again or generate the full report
                  </Text>
                </VStack>
              </Box>
            )}
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onPreviewClose}>
              Close
            </Button>
            <Button 
              colorScheme="blue" 
              onClick={() => {
                onPreviewClose();
                handleGenerateReport(previewReport);
              }}
              leftIcon={<FiDownload />}
            >
              Generate Full Report
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* Journal Drilldown Modal */}
      {journalDrilldown.drilldownRequest && (
        <JournalDrilldownModal
          isOpen={journalDrilldown.isOpen}
          onClose={journalDrilldown.closeDrilldown}
          drilldownRequest={journalDrilldown.drilldownRequest}
          title={journalDrilldown.title}
        />
      )}
    </SimpleLayout>
  );
};

export default ReportsPage;
