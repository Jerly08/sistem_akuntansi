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
  FiDatabase
} from 'react-icons/fi';
// Legacy reportService removed - now using SSOT services only
import { ssotBalanceSheetReportService, SSOTBalanceSheetData } from '../../src/services/ssotBalanceSheetReportService';
import { ssotCashFlowReportService, SSOTCashFlowData } from '../../src/services/ssotCashFlowReportService';
import { ssotSalesSummaryService, SSOTSalesSummaryData } from '../../src/services/ssotSalesSummaryService';
import { ssotVendorAnalysisService, SSOTVendorAnalysisData } from '../../src/services/ssotVendorAnalysisService';
import { ssotTrialBalanceService, SSOTTrialBalanceData } from '../../src/services/ssotTrialBalanceService';
import { ssotGeneralLedgerService, SSOTGeneralLedgerData } from '../../src/services/ssotGeneralLedgerService';
import { ssotJournalAnalysisService, SSOTJournalAnalysisData } from '../../src/services/ssotJournalAnalysisService';

// Define reports data matching the UI design
const getAvailableReports = (t: any) => [
  {
    id: 'profit-loss',
    name: t('reports.profitLossStatement'),
    description: 'Comprehensive profit and loss statement with enhanced analysis. Automatically integrates journal entry data for accurate revenue, COGS, and expense reporting with detailed financial metrics.',
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
  },
  {
    id: 'journal-entry-analysis',
    name: 'Journal Entry Analysis',
    description: 'Complete analysis of all journal entries showing all transactions with detailed breakdown by accounts, dates, and amounts',
    type: 'FINANCIAL',
    icon: FiDatabase
  }
];


const ReportsPage: React.FC = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  // Legacy states removed - now using SSOT system only
  const toast = useToast();

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
  const rowHoverBg = useColorModeValue('blue.50', 'blue.900');
  
  const availableReports = getAvailableReports(t);
  
  // State untuk SSOT Profit Loss
  const [ssotPLOpen, setSSOTPLOpen] = useState(false);
  const [ssotPLData, setSSOTPLData] = useState<any>(null);
  const [ssotPLLoading, setSSOTPLLoading] = useState(false);
  const [ssotPLError, setSSOTPLError] = useState<string | null>(null);
  const [ssotStartDate, setSSOTStartDate] = useState('2025-01-01');
  const [ssotEndDate, setSSOTEndDate] = useState('2025-12-31');

  // State untuk SSOT Balance Sheet
  const [ssotBSOpen, setSSOTBSOpen] = useState(false);
  const [ssotBSData, setSSOTBSData] = useState<SSOTBalanceSheetData | null>(null);
  const [ssotBSLoading, setSSOTBSLoading] = useState(false);
  const [ssotBSError, setSSOTBSError] = useState<string | null>(null);
  const [ssotAsOfDate, setSSOTAsOfDate] = useState(new Date().toISOString().split('T')[0]);

  // State untuk SSOT Cash Flow
  const [ssotCFOpen, setSSOTCFOpen] = useState(false);
  const [ssotCFData, setSSOTCFData] = useState<SSOTCashFlowData | null>(null);
  const [ssotCFLoading, setSSOTCFLoading] = useState(false);
  const [ssotCFError, setSSOTCFError] = useState<string | null>(null);
  const [ssotCFStartDate, setSSOTCFStartDate] = useState(() => {
    const today = new Date();
    const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
    return firstDayOfMonth.toISOString().split('T')[0];
  });
  const [ssotCFEndDate, setSSOTCFEndDate] = useState(new Date().toISOString().split('T')[0]);

  // State untuk SSOT Sales Summary
  const [ssotSSOpen, setSSOTSSOpen] = useState(false);
  const [ssotSSData, setSSOTSSData] = useState<SSOTSalesSummaryData | null>(null);
  const [ssotSSLoading, setSSOTSSLoading] = useState(false);
  const [ssotSSError, setSSOTSSError] = useState<string | null>(null);
  const [ssotSSStartDate, setSSOTSSStartDate] = useState('2025-01-01');
  const [ssotSSEndDate, setSSOTSSEndDate] = useState('2025-12-31');

  // State untuk SSOT Vendor Analysis
  const [ssotVAOpen, setSSOTVAOpen] = useState(false);
  const [ssotVAData, setSSOTVAData] = useState<SSOTVendorAnalysisData | null>(null);
  const [ssotVALoading, setSSOTVALoading] = useState(false);
  const [ssotVAError, setSSOTVAError] = useState<string | null>(null);
  const [ssotVAStartDate, setSSOTVAStartDate] = useState('2025-01-01');
  const [ssotVAEndDate, setSSOTVAEndDate] = useState('2025-12-31');

  // State untuk SSOT Trial Balance
  const [ssotTBOpen, setSSOTTBOpen] = useState(false);
  const [ssotTBData, setSSOTTBData] = useState<SSOTTrialBalanceData | null>(null);
  const [ssotTBLoading, setSSOTTBLoading] = useState(false);
  const [ssotTBError, setSSOTTBError] = useState<string | null>(null);
  const [ssotTBAsOfDate, setSSOTTBAsOfDate] = useState(new Date().toISOString().split('T')[0]);

  // State untuk SSOT General Ledger
  const [ssotGLOpen, setSSOTGLOpen] = useState(false);
  const [ssotGLData, setSSOTGLData] = useState<SSOTGeneralLedgerData | null>(null);
  const [ssotGLLoading, setSSOTGLLoading] = useState(false);
  const [ssotGLError, setSSOTGLError] = useState<string | null>(null);
  const [ssotGLStartDate, setSSOTGLStartDate] = useState('2025-01-01');
  const [ssotGLEndDate, setSSOTGLEndDate] = useState('2025-12-31');
  const [ssotGLAccountId, setSSOTGLAccountId] = useState<string>('');

  // State untuk SSOT Journal Analysis
  const [ssotJAOpen, setSSOTJAOpen] = useState(false);
  const [ssotJAData, setSSOTJAData] = useState<SSOTJournalAnalysisData | null>(null);
  const [ssotJALoading, setSSOTJALoading] = useState(false);
  const [ssotJAError, setSSOTJAError] = useState<string | null>(null);
  const [ssotJAStartDate, setSSOTJAStartDate] = useState('2025-01-01');
  const [ssotJAEndDate, setSSOTJAEndDate] = useState('2025-12-31');

  // Legacy resetParams removed

  // Function untuk fetch SSOT Sales Summary Report
  const fetchSSOTSalesSummaryReport = async () => {
    setSSOTSSLoading(true);
    setSSOTSSError(null);
    
    try {
      const salesSummaryData = await ssotSalesSummaryService.generateSSOTSalesSummary({
        start_date: ssotSSStartDate,
        end_date: ssotSSEndDate,
        format: 'json'
      });
      
      setSSOTSSData(salesSummaryData);
      
      toast({
        title: 'Success',
        description: 'SSOT Sales Summary generated successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      setSSOTSSError(error.message || 'Failed to generate sales summary');
      toast({
        title: 'Error',
        description: error.message || 'Failed to generate sales summary',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setSSOTSSLoading(false);
    }
  };

  // Function untuk fetch SSOT Vendor Analysis Report
  const fetchSSOTVendorAnalysisReport = async () => {
    setSSOTVALoading(true);
    setSSOTVAError(null);
    
    try {
      const vendorAnalysisData = await ssotVendorAnalysisService.generateSSOTVendorAnalysis({
        start_date: ssotVAStartDate,
        end_date: ssotVAEndDate,
        format: 'json'
      });
      
      setSSOTVAData(vendorAnalysisData);
      
      toast({
        title: 'Success',
        description: 'SSOT Vendor Analysis generated successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      setSSOTVAError(error.message || 'Failed to generate vendor analysis');
      toast({
        title: 'Error',
        description: error.message || 'Failed to generate vendor analysis',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setSSOTVALoading(false);
    }
  };

  // Function untuk fetch SSOT Trial Balance Report
  const fetchSSOTTrialBalanceReport = async () => {
    setSSOTTBLoading(true);
    setSSOTTBError(null);
    
    try {
      const trialBalanceData = await ssotTrialBalanceService.generateSSOTTrialBalance({
        as_of_date: ssotTBAsOfDate,
        format: 'json'
      });
      
      setSSOTTBData(trialBalanceData);
      
      toast({
        title: 'Success',
        description: 'SSOT Trial Balance generated successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      setSSOTTBError(error.message || 'Failed to generate trial balance');
      toast({
        title: 'Error',
        description: error.message || 'Failed to generate trial balance',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setSSOTTBLoading(false);
    }
  };

  // Function untuk fetch SSOT General Ledger Report
  const fetchSSOTGeneralLedgerReport = async () => {
    setSSOTGLLoading(true);
    setSSOTGLError(null);
    
    try {
      const generalLedgerData = await ssotGeneralLedgerService.generateSSOTGeneralLedger({
        start_date: ssotGLStartDate,
        end_date: ssotGLEndDate,
        account_id: ssotGLAccountId || undefined,
        format: 'json'
      });
      
      setSSOTGLData(generalLedgerData);
      
      toast({
        title: 'Success',
        description: 'SSOT General Ledger generated successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      setSSOTGLError(error.message || 'Failed to generate general ledger');
      toast({
        title: 'Error',
        description: error.message || 'Failed to generate general ledger',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setSSOTGLLoading(false);
    }
  };

  // Function untuk fetch SSOT Journal Analysis Report
  const fetchSSOTJournalAnalysisReport = async () => {
    setSSOTJALoading(true);
    setSSOTJAError(null);
    
    try {
      const journalAnalysisData = await ssotJournalAnalysisService.generateSSOTJournalAnalysis({
        start_date: ssotJAStartDate,
        end_date: ssotJAEndDate,
        format: 'json'
      });
      
      setSSOTJAData(journalAnalysisData);
      
      toast({
        title: 'Success',
        description: 'SSOT Journal Analysis generated successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      setSSOTJAError(error.message || 'Failed to generate journal analysis');
      toast({
        title: 'Error',
        description: error.message || 'Failed to generate journal analysis',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setSSOTJALoading(false);
    }
  };

  // Function untuk fetch SSOT Balance Sheet Report
  const fetchSSOTBalanceSheetReport = async () => {
    setSSOTBSLoading(true);
    setSSOTBSError(null);
    
    try {
      const balanceSheetData = await ssotBalanceSheetReportService.generateSSOTBalanceSheet({
        as_of_date: ssotAsOfDate,
        format: 'json'
      });
      
      setSSOTBSData(balanceSheetData);
      
      toast({
        title: 'Success',
        description: 'SSOT Balance Sheet generated successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      
    } catch (error) {
      console.error('Error fetching SSOT Balance Sheet report:', error);
      const errorMessage = error instanceof Error ? error.message : 'An error occurred';
      setSSOTBSError(errorMessage);
      
      toast({
        title: 'Error',
        description: errorMessage,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setSSOTBSLoading(false);
    }
  };

  // Function untuk fetch SSOT Cash Flow Report
  const fetchSSOTCashFlowReport = async () => {
    setSSOTCFLoading(true);
    setSSOTCFError(null);
    
    try {
      const cashFlowData = await ssotCashFlowReportService.generateSSOTCashFlow({
        start_date: ssotCFStartDate,
        end_date: ssotCFEndDate,
        format: 'json'
      });
      
      // Debug: Log the raw response
      console.log('Raw cash flow data received:', cashFlowData);
      console.log('Net cash flow value:', cashFlowData.net_cash_flow, 'Type:', typeof cashFlowData.net_cash_flow);
      console.log('Cash at beginning:', cashFlowData.cash_at_beginning, 'Type:', typeof cashFlowData.cash_at_beginning);
      console.log('Cash at end:', cashFlowData.cash_at_end, 'Type:', typeof cashFlowData.cash_at_end);
      
      setSSOTCFData(cashFlowData as SSOTCashFlowData);
      
      toast({
        title: 'Success',
        description: 'SSOT Cash Flow Statement generated successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      
    } catch (error) {
      console.error('Error fetching SSOT Cash Flow report:', error);
      const errorMessage = error instanceof Error ? error.message : 'An error occurred';
      setSSOTCFError(errorMessage);
      
      toast({
        title: 'Error',
        description: errorMessage,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setSSOTCFLoading(false);
    }
  };

  // Function untuk fetch SSOT P&L Report
  const fetchSSOTPLReport = async () => {
    setSSOTPLLoading(true);
    setSSOTPLError(null);
    
    try {
      // Get token from the same location as reportService for consistency
      // Primary method: get from localStorage with 'token' key
      let token = null;
      
      if (typeof window !== 'undefined') {
        // Method 1: localStorage - try 'token' first (which is what AuthContext uses)
        token = localStorage.getItem('token');
        
        // Fallback to alternative keys if needed
        if (!token) {
          token = localStorage.getItem('authToken') || 
                 sessionStorage.getItem('token') || 
                 sessionStorage.getItem('authToken');
                 
          // If token found with alternative key, store it with the correct key for future use
          if (token) {
            console.log('Token found with alternative key, storing with correct key');
            localStorage.setItem('token', token);
          }
        }
        
        // Method 2: Try cookies as last resort
        if (!token) {
          const cookies = document.cookie.split(';');
          for (let cookie of cookies) {
            const [name, value] = cookie.trim().split('=');
            if (name === 'token' || name === 'authToken' || name === 'access_token') {
              token = value;
              break;
            }
          }
        }
      }
      
      if (!token) {
        throw new Error('Authentication token not found. Please login first.');
      }

      const response = await fetch(
        `http://localhost:8080/api/v1/reports/ssot-profit-loss?start_date=${ssotStartDate}&end_date=${ssotEndDate}&format=json`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }
      );

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      
      if (result.status === 'success' && result.data) {
        setSSOTPLData(result.data);
      } else {
        throw new Error(result.message || 'Failed to fetch P&L data');
      }
    } catch (error) {
      console.error('Error fetching SSOT P&L report:', error);
      setSSOTPLError(error instanceof Error ? error.message : 'An error occurred');
    } finally {
      setSSOTPLLoading(false);
    }
  };

  const handleViewReport = async (report: any) => {
    setLoading(true);
    
    try {
      if (report.id === 'balance-sheet') {
        setSSOTBSOpen(true);
        await fetchSSOTBalanceSheetReport();
      } else if (report.id === 'profit-loss') {
        setSSOTPLOpen(true);
        await fetchSSOTPLReport();
      } else if (report.id === 'cash-flow') {
        setSSOTCFOpen(true);
        await fetchSSOTCashFlowReport();
      } else if (report.id === 'sales-summary') {
        setSSOTSSOpen(true);
        await fetchSSOTSalesSummaryReport();
      } else if (report.id === 'vendor-analysis') {
        setSSOTVAOpen(true);
        await fetchSSOTVendorAnalysisReport();
      } else if (report.id === 'trial-balance') {
        setSSOTTBOpen(true);
        await fetchSSOTTrialBalanceReport();
      } else if (report.id === 'general-ledger') {
        setSSOTGLOpen(true);
        await fetchSSOTGeneralLedgerReport();
      } else if (report.id === 'journal-entry-analysis') {
        setSSOTJAOpen(true);
        await fetchSSOTJournalAnalysisReport();
      }
      
    } catch (error) {
      console.error('Failed to load SSOT report:', error);
      
      toast({
        title: 'Report Load Error',
        description: error instanceof Error ? error.message : 'Failed to load report',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
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
          // Handle both SSOT and legacy BalanceSheetData structures from backend
          if (reportData.company || reportData.sections || reportData.assets) {
            const sections = [];
            
            // Check if it's SSOT Balance Sheet format
            if (reportData.assets && reportData.assets.current_assets && reportData.assets.non_current_assets) {
              // SSOT format with detailed categorization
              const assetItems = [
                ...reportData.assets.current_assets.items || [],
                ...reportData.assets.non_current_assets.items || []
              ];
              
              const liabilityItems = [
                ...reportData.liabilities?.current_liabilities?.items || [],
                ...reportData.liabilities?.non_current_liabilities?.items || []
              ];
              
              sections.push({
                name: 'ASSETS',
                items: assetItems.map((item: any) => ({
                  name: `${item.account_code} - ${item.account_name}`,
                  amount: item.amount
                })),
                total: reportData.assets.total_assets || 0,
                subsections: [
                  {
                    name: 'Current Assets',
                    total: reportData.assets.current_assets.total_current_assets || 0,
                    items: reportData.assets.current_assets.items?.map((item: any) => ({
                      name: `${item.account_code} - ${item.account_name}`,
                      amount: item.amount
                    })) || []
                  },
                  {
                    name: 'Non-Current Assets',
                    total: reportData.assets.non_current_assets.total_non_current_assets || 0,
                    items: reportData.assets.non_current_assets.items?.map((item: any) => ({
                      name: `${item.account_code} - ${item.account_name}`,
                      amount: item.amount
                    })) || []
                  }
                ]
              });
              
              sections.push({
                name: 'LIABILITIES',
                items: liabilityItems.map((item: any) => ({
                  name: `${item.account_code} - ${item.account_name}`,
                  amount: item.amount
                })),
                total: reportData.liabilities?.total_liabilities || 0,
                subsections: [
                  {
                    name: 'Current Liabilities',
                    total: reportData.liabilities?.current_liabilities?.total_current_liabilities || 0,
                    items: reportData.liabilities?.current_liabilities?.items?.map((item: any) => ({
                      name: `${item.account_code} - ${item.account_name}`,
                      amount: item.amount
                    })) || []
                  },
                  {
                    name: 'Non-Current Liabilities',
                    total: reportData.liabilities?.non_current_liabilities?.total_non_current_liabilities || 0,
                    items: reportData.liabilities?.non_current_liabilities?.items?.map((item: any) => ({
                      name: `${item.account_code} - ${item.account_name}`,
                      amount: item.amount
                    })) || []
                  }
                ]
              });
              
              sections.push({
                name: 'EQUITY',
                items: reportData.equity?.items?.map((item: any) => ({
                  name: `${item.account_code} - ${item.account_name}`,
                  amount: item.amount
                })) || [],
                total: reportData.equity?.total_equity || 0
              });
            } else {
              // Legacy format
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
            }
            
            return {
              title: reportData.enhanced ? 'Enhanced Balance Sheet (SSOT)' : 'Balance Sheet',
              period: reportData.as_of_date ? `As of ${new Date(reportData.as_of_date).toLocaleDateString('id-ID')}` : reportData.period || `As of ${new Date().toLocaleDateString('id-ID')}`,
              sections,
              isBalanced: reportData.is_balanced,
              balanceDifference: reportData.balance_difference,
              enhanced: reportData.enhanced || false
            };
          }
          throw new Error('Invalid balance sheet data structure');

        case 'profit-loss':
          // Handle Enhanced ProfitLossData structure from backend
          console.log('Processing enhanced profit-loss data:', reportData);
          
          // Check if it's the enhanced ProfitLossData format (from EnhancedProfitLossService)
          if (reportData.company || reportData.revenue || reportData.cost_of_goods_sold || reportData.generated_at) {
            const sections = [];
            
            // Revenue section - handle enhanced structure
            if (reportData.revenue) {
              const revenueSection = {
                name: 'REVENUE',
                items: [] as any[],
                total: reportData.revenue.total_revenue || 0,
                subsections: [] as any[]
              };
              
              // Sales Revenue subsection
              if (reportData.revenue.sales_revenue && reportData.revenue.sales_revenue.items && reportData.revenue.sales_revenue.items.length > 0) {
                revenueSection.subsections.push({
                  name: 'Sales Revenue',
                  items: reportData.revenue.sales_revenue.items.map((item: any) => ({
                    name: `${item.code || ''} - ${item.name || ''}`,
                    amount: item.amount || 0,
                    accountCode: item.code
                  })),
                  total: reportData.revenue.sales_revenue.subtotal || 0
                });
              }
              
              // Service Revenue subsection
              if (reportData.revenue.service_revenue && reportData.revenue.service_revenue.items && reportData.revenue.service_revenue.items.length > 0) {
                revenueSection.subsections.push({
                  name: 'Service Revenue',
                  items: reportData.revenue.service_revenue.items.map((item: any) => ({
                    name: `${item.code || ''} - ${item.name || ''}`,
                    amount: item.amount || 0,
                    accountCode: item.code
                  })),
                  total: reportData.revenue.service_revenue.subtotal || 0
                });
              }
              
              // Other Revenue subsection
              if (reportData.revenue.other_revenue && reportData.revenue.other_revenue.items && reportData.revenue.other_revenue.items.length > 0) {
                revenueSection.subsections.push({
                  name: 'Other Revenue',
                  items: reportData.revenue.other_revenue.items.map((item: any) => ({
                    name: `${item.code || ''} - ${item.name || ''}`,
                    amount: item.amount || 0,
                    accountCode: item.code
                  })),
                  total: reportData.revenue.other_revenue.subtotal || 0
                });
              }
              
              // If no subsections but has total, create simple revenue items
              if (revenueSection.subsections.length === 0 && reportData.revenue.total_revenue > 0) {
                revenueSection.items.push({
                  name: 'Total Revenue',
                  amount: reportData.revenue.total_revenue,
                  accountCode: '4000' // Generic revenue account
                });
              }
              
              sections.push(revenueSection);
            }
            
            // Cost of Goods Sold section - handle enhanced structure
            if (reportData.cost_of_goods_sold) {
              const cogsSection = {
                name: 'COST OF GOODS SOLD',
                items: [] as any[],
                total: reportData.cost_of_goods_sold.total_cogs || 0,
                subsections: [] as any[]
              };
              
              // Add subsections for detailed COGS breakdown
              if (reportData.cost_of_goods_sold.direct_materials?.items?.length > 0) {
                cogsSection.subsections.push({
                  name: 'Direct Materials',
                  items: reportData.cost_of_goods_sold.direct_materials.items.map((item: any) => ({
                    name: `${item.code || ''} - ${item.name || ''}`,
                    amount: item.amount || 0,
                    accountCode: item.code
                  })),
                  total: reportData.cost_of_goods_sold.direct_materials.subtotal || 0
                });
              }
              
              if (reportData.cost_of_goods_sold.other_cogs?.items?.length > 0) {
                cogsSection.subsections.push({
                  name: 'Other COGS',
                  items: reportData.cost_of_goods_sold.other_cogs.items.map((item: any) => ({
                    name: `${item.code || ''} - ${item.name || ''}`,
                    amount: item.amount || 0,
                    accountCode: item.code
                  })),
                  total: reportData.cost_of_goods_sold.other_cogs.subtotal || 0
                });
              }
              
              // If no subsections but has total, create simple COGS items
              if (cogsSection.subsections.length === 0 && reportData.cost_of_goods_sold.total_cogs > 0) {
                cogsSection.items.push({
                  name: 'Cost of Goods Sold',
                  amount: reportData.cost_of_goods_sold.total_cogs,
                  accountCode: '5101' // Standard COGS account
                });
              }
              
              // Only add COGS section if there's data or total amount
              if (cogsSection.subsections.length > 0 || cogsSection.items.length > 0 || cogsSection.total !== 0) {
                sections.push(cogsSection);
              }
            }
            
            // Gross Profit section with margin
            if (reportData.gross_profit !== undefined) {
              sections.push({
                name: 'GROSS PROFIT',
                items: [
                  { name: 'Gross Profit', amount: reportData.gross_profit || 0 },
                  { name: 'Gross Profit Margin', amount: reportData.gross_profit_margin || 0, isPercentage: true }
                ],
                total: reportData.gross_profit || 0,
                isCalculated: true
              });
            }
            
            // Operating Expenses section - handle enhanced structure
            if (reportData.operating_expenses) {
              const opexSection = {
                name: 'OPERATING EXPENSES',
                items: [] as any[],
                total: reportData.operating_expenses.total_opex || 0,
                subsections: [] as any[]
              };
              
              // Add subsections for detailed operating expenses breakdown
              if (reportData.operating_expenses.administrative?.items?.length > 0) {
                opexSection.subsections.push({
                  name: 'Administrative Expenses',
                  items: reportData.operating_expenses.administrative.items.map((item: any) => ({
                    name: `${item.code || ''} - ${item.name || ''}`,
                    amount: item.amount || 0,
                    accountCode: item.code
                  })),
                  total: reportData.operating_expenses.administrative.subtotal || 0
                });
              }
              
              if (reportData.operating_expenses.selling_marketing?.items?.length > 0) {
                opexSection.subsections.push({
                  name: 'Selling & Marketing Expenses',
                  items: reportData.operating_expenses.selling_marketing.items.map((item: any) => ({
                    name: `${item.code || ''} - ${item.name || ''}`,
                    amount: item.amount || 0,
                    accountCode: item.code
                  })),
                  total: reportData.operating_expenses.selling_marketing.subtotal || 0
                });
              }
              
              if (reportData.operating_expenses.general?.items?.length > 0) {
                opexSection.subsections.push({
                  name: 'General Expenses',
                  items: reportData.operating_expenses.general.items.map((item: any) => ({
                    name: `${item.code || ''} - ${item.name || ''}`,
                    amount: item.amount || 0,
                    accountCode: item.code
                  })),
                  total: reportData.operating_expenses.general.subtotal || 0
                });
              }
              
              // If no subsections but has total, create simple operating expense items
              if (opexSection.subsections.length === 0 && reportData.operating_expenses.total_opex > 0) {
                opexSection.items.push({
                  name: 'Operating Expenses',
                  amount: reportData.operating_expenses.total_opex,
                  accountCode: '6000' // Generic expense account
                });
              }
              
              // Only add section if there's data or total amount
              if (opexSection.subsections.length > 0 || opexSection.items.length > 0 || opexSection.total !== 0) {
                sections.push(opexSection);
              }
            }
            
            // Operating Income and EBITDA section
            if (reportData.operating_income !== undefined) {
              sections.push({
                name: 'OPERATING PERFORMANCE',
                items: [
                  { name: 'Operating Income (EBIT)', amount: reportData.operating_income || 0 },
                  { name: 'Operating Margin', amount: reportData.operating_margin || 0, isPercentage: true },
                  { name: 'EBITDA', amount: reportData.ebitda || 0 },
                  { name: 'EBITDA Margin', amount: reportData.ebitda_margin || 0, isPercentage: true }
                ],
                total: reportData.operating_income || 0,
                isCalculated: true
              });
            }
            
            // Net Income section with comprehensive metrics
            sections.push({
              name: 'NET INCOME',
              items: [
                { name: 'Income Before Tax', amount: reportData.income_before_tax || 0 },
                { name: 'Tax Expense', amount: reportData.tax_expense || 0 },
                { name: 'Net Income', amount: reportData.net_income || 0 },
                { name: 'Net Income Margin', amount: reportData.net_income_margin || 0, isPercentage: true }
              ],
              total: reportData.net_income || 0,
              isCalculated: true
            });

            // Extract period from enhanced data
            let period = `${new Date().toLocaleDateString('id-ID')}`;
            if (reportData.start_date && reportData.end_date) {
              const startDate = new Date(reportData.start_date);
              const endDate = new Date(reportData.end_date);
              period = `${startDate.toLocaleDateString('id-ID')} - ${endDate.toLocaleDateString('id-ID')}`;
            }

            const hasData = sections.some(section => section.items && section.items.length > 0 && section.total !== 0);
            
            return {
              title: 'Enhanced Profit and Loss Statement',
              period,
              sections,
              hasData,
              company: reportData.company,
              financialMetrics: {
                grossProfit: reportData.gross_profit || 0,
                grossProfitMargin: reportData.gross_profit_margin || 0,
                operatingIncome: reportData.operating_income || 0,
                operatingMargin: reportData.operating_margin || 0,
                ebitda: reportData.ebitda || 0,
                ebitdaMargin: reportData.ebitda_margin || 0,
                netIncome: reportData.net_income || 0,
                netIncomeMargin: reportData.net_income_margin || 0
              },
              enhanced: true,
              message: !hasData ? 'No P&L relevant transactions found for this period. The journal entries contain mainly asset purchases, payments, and deposits which affect the Balance Sheet rather than P&L. To generate meaningful P&L data, record sales transactions, operating expenses, and cost of goods sold.' : undefined
            };
          }
          
          // Fallback: Handle legacy ProfitLossStatement structure
          else if (reportData.report_header || reportData.revenue || reportData.total_revenue !== undefined) {
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
              hasData: sections.some(section => section.items && section.items.length > 0),
              enhanced: false
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
          // Handle SSOT Cash Flow Data structure
          if (reportData.sections && Array.isArray(reportData.sections)) {
            // Direct sections format from SSOT Cash Flow controller
            return {
              title: reportData.title || 'Cash Flow Statement (SSOT)',
              period: reportData.period || `${new Date().toLocaleDateString('id-ID')}`,
              sections: reportData.sections.map((section: any) => ({
                name: section.name,
                items: section.items?.map((item: any) => ({
                  name: item.name,
                  amount: item.amount || 0,
                  accountCode: item.account_code,
                  type: item.type,
                  description: item.description
                })) || [],
                total: section.total || 0,
                subsections: section.subsections || [],
                summary: section.summary || {},
                isCalculated: section.is_calculated || false
              })),
              hasData: reportData.hasData !== false,
              summary: reportData.summary || {},
              cashFlowRatios: reportData.cashFlowRatios || {},
              enhanced: reportData.enhanced || false,
              message: reportData.message
            };
          }
          
          // Fallback: Handle legacy cash flow structure
          if (reportData.company && (reportData.operating_activities || reportData.summary)) {
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
              sections,
              hasData: sections.length > 0
            };
          }
          
          // Return empty state if no valid structure
          return {
            title: 'Cash Flow Statement',
            period: `${new Date().toLocaleDateString('id-ID')}`,
            sections: [],
            hasData: false,
            message: 'No cash flow data available for the selected period. Please ensure there are posted transactions in the journal system.'
          };

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

        case 'journal-entry-analysis':
          // Handle Journal Entry Analysis data structure from reportService.generateJournalEntryAnalysis
          const journalEntries = reportData.journal_entries || reportData.entries || reportData.data || [];
          const totalEntries = reportData.total_entries || reportData.total || journalEntries.length || 0;
          
          console.log('Processing journal entry analysis data:', reportData);
          console.log('Journal entries array:', journalEntries);
          
          if (Array.isArray(journalEntries) && journalEntries.length > 0) {
            // Group entries by date for better organization
            const groupedEntries = journalEntries.reduce((acc: any, entry: any) => {
              // Use entry_date from journal entry model
              const date = entry.entry_date || entry.transaction_date || entry.date || entry.created_at || 'Unknown Date';
              let dateKey: string;
              try {
                dateKey = new Date(date).toISOString().split('T')[0];
              } catch {
                dateKey = 'Invalid Date';
              }
              
              if (!acc[dateKey]) {
                acc[dateKey] = [];
              }
              acc[dateKey].push(entry);
              return acc;
            }, {});
            
            const sections = Object.keys(groupedEntries)
              .sort((a, b) => new Date(b).getTime() - new Date(a).getTime()) // Sort by date descending
              .slice(0, 10) // Show only latest 10 dates for preview
              .map(dateKey => {
                const entries = groupedEntries[dateKey];
                const dateFormatted = dateKey === 'Invalid Date' ? 'Invalid Date' : new Date(dateKey).toLocaleDateString('id-ID');
                
                return {
                  name: `Journal Entries - ${dateFormatted}`,
                  items: entries.map((entry: any) => {
                    // Use journal entry model fields
                    const debitAmount = entry.total_debit || entry.debit_amount || entry.debit || 0;
                    const creditAmount = entry.total_credit || entry.credit_amount || entry.credit || 0;
                    
                    // Format the entry display name more professionally
                    const referenceCode = entry.code || entry.reference || `JE-${entry.id}`;
                    const description = entry.description || 'No description';
                    const referenceType = entry.reference_type ? ` [${entry.reference_type}]` : '';
                    
                    return {
                      name: `${referenceCode}${referenceType}`,
                      description: description,
                      amount: debitAmount, // Show debit amount as primary amount
                      debit: debitAmount,
                      credit: creditAmount,
                      reference: referenceCode,
                      date: entry.entry_date || entry.transaction_date || entry.date,
                      status: entry.status || 'DRAFT',
                      referenceType: entry.reference_type || 'MANUAL',
                      isBalanced: entry.is_balanced || false,
                      accountName: entry.account?.name || 'General',
                      accountCode: entry.account?.code || ''
                    };
                  }),
                  total: entries.reduce((sum: number, entry: any) => {
                    const debit = entry.total_debit || entry.debit_amount || entry.debit || 0;
                    return sum + debit; // Sum total debit amounts
                  }, 0),
                  creditTotal: entries.reduce((sum: number, entry: any) => {
                    const credit = entry.total_credit || entry.credit_amount || entry.credit || 0;
                    return sum + credit; // Sum total credit amounts
                  }, 0),
                  entryCount: entries.length,
                  balancedCount: entries.filter((entry: any) => entry.is_balanced).length
                };
              });
            
            // Calculate totals across all sections
            const totalDebit = sections.reduce((sum: number, section: any) => sum + (section.total || 0), 0);
            const totalCredit = sections.reduce((sum: number, section: any) => sum + (section.creditTotal || 0), 0);
            const totalBalancedEntries = sections.reduce((sum: number, section: any) => sum + (section.balancedCount || 0), 0);
            
            return {
              title: 'Journal Entry Analysis Report',
              period: `${new Date(reportData.start_date || Date.now()).toLocaleDateString('id-ID')} - ${new Date(reportData.end_date || Date.now()).toLocaleDateString('id-ID')}`,
              sections,
              hasData: sections.length > 0,
              totalEntries,
              summary: `Analysis of ${totalEntries} journal entries across ${sections.length} transaction dates`,
              financialSummary: {
                totalDebit: totalDebit,
                totalCredit: totalCredit,
                balancedEntries: totalBalancedEntries,
                unbalancedEntries: totalEntries - totalBalancedEntries,
                balanceAccuracy: totalEntries > 0 ? (totalBalancedEntries / totalEntries * 100).toFixed(1) : '0'
              },
              reportMetadata: {
                generatedAt: new Date().toISOString(),
                dateRange: {
                  start: reportData.start_date,
                  end: reportData.end_date
                },
                entriesAnalyzed: totalEntries,
                periodsIncluded: sections.length
              }
            };
          }
          
          // Handle case where journal_entries might be in a different structure
          if (reportData.accounts && Array.isArray(reportData.accounts)) {
            // Similar to general ledger but focused on transactions
            const sections = reportData.accounts
              .filter((account: any) => account.transactions && account.transactions.length > 0)
              .slice(0, 5) // Limit for preview
              .map((account: any) => ({
                name: `${account.account_code || account.code} - ${account.account_name || account.name}`,
                items: (account.transactions || []).map((txn: any) => ({
                  name: `${txn.description || 'Transaction'} (Ref: ${txn.reference || 'N/A'})`,
                  amount: (txn.debit_amount || txn.debit || 0) - (txn.credit_amount || txn.credit || 0),
                  debit: txn.debit_amount || txn.debit || 0,
                  credit: txn.credit_amount || txn.credit || 0,
                  date: txn.date,
                  reference: txn.reference || ''
                })),
                total: account.transactions.reduce((sum: number, txn: any) => {
                  return sum + ((txn.debit_amount || txn.debit || 0) - (txn.credit_amount || txn.credit || 0));
                }, 0)
              }));
            
            return {
              title: 'Journal Entry Analysis',
              period: `${new Date(reportData.start_date || Date.now()).toLocaleDateString('id-ID')} - ${new Date(reportData.end_date || Date.now()).toLocaleDateString('id-ID')}`,
              sections,
              hasData: sections.length > 0,
              summary: `Showing transactions for ${sections.length} accounts`
            };
          }
          
          // If no valid data structure found, return empty state with more debugging info
          console.log('No valid journal entries found. Raw reportData:', reportData);
          
          return {
            title: 'Journal Entry Analysis',
            period: reportData.start_date && reportData.end_date ? 
              `${new Date(reportData.start_date).toLocaleDateString('id-ID')} - ${new Date(reportData.end_date).toLocaleDateString('id-ID')}` :
              `${new Date().toLocaleDateString('id-ID')}`,
            sections: [],
            hasData: false,
            message: `No journal entries found for the selected period. Data received: ${Array.isArray(journalEntries) ? journalEntries.length : 'not an array'} entries. Please check if there are any posted transactions recorded during this period.`,
            totalEntries: totalEntries,
            debug: {
              hasJournalEntries: !!journalEntries,
              isArray: Array.isArray(journalEntries),
              length: Array.isArray(journalEntries) ? journalEntries.length : 'N/A',
              totalEntries,
              dataKeys: Object.keys(reportData || {})
            }
          };


        default:
          throw new Error(`Unsupported report type: ${report.id}`);
      }
    } catch (error) {
      console.error('Failed to convert API data:', error);
      throw error;
    }
  };


  // Simplified error analysis function
  const analyzeReportError = (error: any, report: any) => {
    const errorMessage = error instanceof Error ? error.message : String(error);
    const reportName = report?.name || 'Unknown Report';
    
    // Simplified error patterns for better UX
    const errorPatterns = [
      {
        pattern: /account_id.*required/i,
        title: 'Missing Account Information',
        userMessage: 'This report requires account information. Please try selecting a specific account or contact support.',
        suggestions: [
          'Try refreshing the page and selecting the report again',
          'Check if you have permission to access account data',
          'Contact your system administrator if the problem persists'
        ],
        canRetry: true,
        duration: 8000
      },
      {
        pattern: /start_date.*end_date.*required/i,
        title: 'Date Range Required',
        userMessage: 'This report requires a date range. The system will set default dates automatically.',
        suggestions: [
          'The report will use the current month as default',
          'You can specify custom dates when generating the full report',
        ],
        canRetry: true,
        duration: 6000
      },
      {
        pattern: /no.*data.*found|empty.*result/i,
        title: 'No Data Available',
        userMessage: `No data found for ${reportName}. This may be normal if there are no transactions in the selected period.`,
        suggestions: [
          'Try selecting a different date range',
          'Check if there are any posted transactions in your accounting system',
          'Verify your permissions to access this data'
        ],
        canRetry: true,
        duration: 7000
      },
      {
        pattern: /network.*error|fetch.*failed|connection/i,
        title: 'Connection Problem',
        userMessage: 'Unable to connect to the server. Please check your internet connection.',
        suggestions: [
          'Check your internet connection',
          'Try refreshing the page',
          'Wait a moment and try again'
        ],
        canRetry: true,
        duration: 6000
      },
      {
        pattern: /unauthorized|permission.*denied|access.*denied/i,
        title: 'Access Denied',
        userMessage: 'You do not have permission to access this report.',
        suggestions: [
          'Contact your system administrator to request access',
          'Verify that you are logged in with the correct account',
          'Check if your session has expired'
        ],
        canRetry: false,
        duration: 8000
      },
      {
        pattern: /server.*error|internal.*error|500/i,
        title: 'Server Error',
        userMessage: 'A server error occurred while generating the report. Our technical team has been notified.',
        suggestions: [
          'Try again in a few minutes',
          'Contact support if the problem persists',
          'Use a different report format if available'
        ],
        canRetry: true,
        duration: 8000
      }
    ];
    
    // Find matching error pattern
    const matchedPattern = errorPatterns.find(pattern => 
      pattern.pattern.test(errorMessage)
    );
    
    if (matchedPattern) {
      return {
        title: matchedPattern.title,
        userMessage: matchedPattern.userMessage,
        suggestions: matchedPattern.suggestions,
        canRetry: matchedPattern.canRetry,
        duration: matchedPattern.duration,
        technicalDetails: errorMessage
      };
    }
    
    // Default error handling
    return {
      title: 'Report Generation Error',
      userMessage: `An unexpected error occurred while loading ${reportName}. Please try again.`,
      suggestions: [
        'Refresh the page and try again',
        'Try generating a different report first',
        'Contact support if this error continues'
      ],
      canRetry: true,
      duration: 7000,
      technicalDetails: errorMessage
    };
  };

  // Format currency for display
  const formatCurrency = (amount: number | null | undefined) => {
    // Handle null, undefined, or NaN values
    if (amount === null || amount === undefined || isNaN(Number(amount))) {
      console.warn('formatCurrency received invalid value:', amount, 'Type:', typeof amount);
      return 'Rp 0';
    }
    
    const numericAmount = Number(amount);
    if (!isFinite(numericAmount)) {
      console.warn('formatCurrency received non-finite value:', amount);
      return 'Rp 0';
    }
    
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0
    }).format(numericAmount);
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
        account_id: 'all',
        start_date: firstDayOfMonth.toISOString().split('T')[0],
        end_date: today.toISOString().split('T')[0],
        format: 'pdf'
      });
    } else if (report.id === 'journal-entry-analysis') {
      // Set default start date to first day of current month and end date to today
      const today = new Date();
      const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
      setReportParams({
        start_date: firstDayOfMonth.toISOString().split('T')[0],
        end_date: today.toISOString().split('T')[0],
        status: 'POSTED', // Default to show posted entries
        reference_type: 'ALL',
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
      if (['profit-loss', 'cash-flow', 'sales-summary', 'purchase-summary', 'general-ledger', 'journal-entry-analysis'].includes(selectedReport.id)) {
        if (!reportParams.start_date || !reportParams.end_date) {
          throw new Error('Start date and end date are required for this report');
        }
      }
      
      if (['balance-sheet', 'trial-balance'].includes(selectedReport.id)) {
        if (!reportParams.as_of_date) {
          throw new Error('As of date is required for this report');
        }
      }
      
      if (selectedReport.id === 'general-ledger') {
        if (!reportParams.account_id) {
          throw new Error('Account ID is required for General Ledger report');
        }
      }
      
      // Use unified report generation for all report types
      result = await reportService.generateReport(selectedReport.id, reportParams);
      
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
              <Flex justify="space-between" align="center" w="full">
                <Heading as="h1" size="xl" color={headingColor} fontWeight="medium">
                  Financial Reports
                </Heading>
              </Flex>
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
                          onClick={() => {
                            // Open SSOT modals for new reports
                            if (report.id === 'sales-summary') {
                              setSSOTSSOpen(true);
                            } else if (report.id === 'vendor-analysis') {
                              setSSOTVAOpen(true);
                            } else if (report.id === 'trial-balance') {
                              setSSOTTBOpen(true);
                            } else if (report.id === 'general-ledger') {
                              // Always use SSOT General Ledger
                              setSSOTGLOpen(true);
                            } else if (report.id === 'journal-entry-analysis') {
                              setSSOTJAOpen(true);
                            } else {
                              handleGenerateReport(report);
                            }
                          }}
                          isLoading={loading && selectedReport?.id === report.id}
                          leftIcon={<FiFileText />}
                        >
                          Generate
                        </Button>
                      </HStack>
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
                      <FormLabel>Account Selection</FormLabel>
                      <Select 
                        name="account_id" 
                        value={reportParams.account_id || 'all'} 
                        onChange={handleInputChange}
                      >
                        <option value="all">All Accounts</option>
                        <option value="1101">Cash Account</option>
                        <option value="1102">Bank BCA</option>
                        <option value="1104">Bank Mandiri</option>
                        <option value="1201">Accounts Receivable</option>
                        <option value="2101">Accounts Payable</option>
                        <option value="4101">Sales Revenue</option>
                        <option value="5101">Cost of Goods Sold</option>
                        <option value="6101">Administrative Expenses</option>
                      </Select>
                    </FormControl>
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
                
                {/* Journal Entry Analysis Parameters */}
                {selectedReport.id === 'journal-entry-analysis' && (
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
                      <FormLabel>Status Filter</FormLabel>
                      <Select 
                        name="status" 
                        value={reportParams.status || 'ALL'} 
                        onChange={handleInputChange}
                      >
                        <option value="ALL">All Status</option>
                        <option value="DRAFT">Draft</option>
                        <option value="POSTED">Posted</option>
                        <option value="REVERSED">Reversed</option>
                      </Select>
                    </FormControl>
                    <FormControl>
                      <FormLabel>Reference Type</FormLabel>
                      <Select 
                        name="reference_type" 
                        value={reportParams.reference_type || 'ALL'} 
                        onChange={handleInputChange}
                      >
                        <option value="ALL">All Types</option>
                        <option value="PURCHASE">Purchase</option>
                        <option value="SALE">Sale</option>
                        <option value="PAYMENT">Payment</option>
                        <option value="CASH_BANK">Cash/Bank</option>
                        <option value="MANUAL">Manual</option>
                      </Select>
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
                // Enhanced Error State or Empty State
                <Box textAlign="center" py={8}>
                  <VStack spacing={6}>
                    <Icon as={FiFileText} boxSize={12} color={previewData.error ? errorIconColor : noDataIconColor} />
                    <VStack spacing={3}>
                      <Text fontSize="lg" fontWeight="bold" color={previewData.error ? errorTextColor : noDataTextColor}>
                        {previewData.error ? 'Unable to Load Preview' : 'No Data Available'}
                      </Text>
                      <Text fontSize="md" color={textColor} maxW="lg" textAlign="center" lineHeight="tall">
                        {previewData.message || (previewData.error ? 'There was a problem loading the report data.' : 'No data found for the selected period or criteria.')}
                      </Text>
                      
                      {/* Enhanced error information */}
                      {previewData.suggestions && previewData.suggestions.length > 0 && (
                        <Box bg={summaryBg} p={4} borderRadius="md" maxW="lg" w="full" mt={4}>
                          <Text fontSize="sm" fontWeight="medium" color={textColor} mb={2}>
                            💡 Suggestions to resolve this issue:
                          </Text>
                          <VStack spacing={1} align="start">
                            {previewData.suggestions.map((suggestion: string, index: number) => (
                              <Text key={index} fontSize="sm" color={summaryTextColor} pl={4} position="relative">
                                <Text as="span" position="absolute" left={0}>•</Text>
                                {suggestion}
                              </Text>
                            ))}
                          </VStack>
                        </Box>
                      )}
                      
                      {/* Technical details for debugging (only in development or for admins) */}
                      {previewData.technicalDetails && process.env.NODE_ENV === 'development' && (
                        <Box bg={errorIconColor} p={3} borderRadius="md" maxW="lg" w="full" mt={2}>
                          <Text fontSize="xs" color="white" fontFamily="mono">
                            Technical Details: {previewData.technicalDetails}
                          </Text>
                        </Box>
                      )}
                    </VStack>
                    
                    <HStack spacing={3}>
                      {previewData.canRetry !== false && (
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
                      )}
                      
                      <Button 
                        colorScheme="green" 
                        onClick={() => {
                          onPreviewClose();
                          handleGenerateReport(previewReport);
                        }}
                      >
                        Generate Full Report
                      </Button>
                    </HStack>
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
                        {/* Header row for professional layout */}
                        <HStack py={2} px={4} fontSize="xs" color={summaryTextColor}>
                          <Text flex={2}>Reference</Text>
                          <Text flex={3}>Description</Text>
                          <Text flex={1}>Status</Text>
                          <Text flex={1} textAlign="right">Debit</Text>
                          <Text flex={1} textAlign="right">Credit</Text>
                        </HStack>
                        {section.items?.map((item: any, itemIndex: number) => (
                          <HStack key={itemIndex} py={2} px={4} 
                                 bg={itemIndex % 2 === 0 ? evenRowBg : oddRowBg} 
                                 borderRadius="md"
                                 _hover={{ bg: rowHoverBg }}
                                 transition="background 0.2s">
                            <Text fontSize="sm" color={textColor} flex={2}>
                              {item.name}{item.accountCode ? ` (${item.accountCode})` : ''}
                            </Text>
                            <Text fontSize="sm" color={summaryTextColor} flex={3}>
                              {item.description}
                            </Text>
                            <Text fontSize="sm" color={textColor} flex={1}>
                              {item.status}
                            </Text>
                            <Text fontSize="sm" fontWeight="medium" color="black" minW="120px" textAlign="right" flex={1}>
                              {formatCurrency(item.debit || 0)}
                            </Text>
                            <Text fontSize="sm" fontWeight="medium" color="black" minW="120px" textAlign="right" flex={1}>
                              {formatCurrency(item.credit || 0)}
                            </Text>
                          </HStack>
                        ))}
                        
                        {/* Section Totals */}
                        <HStack justify="space-between" py={3} px={4} 
                               bg={sectionTotalBg} borderRadius="md" 
                               borderTop="2px" borderColor={sectionTotalBorderColor} mt={2}>
                          <Text fontSize="md" fontWeight="bold" color={sectionTotalTextColor} flex={6}>
                            Total {section.name} (Entries: {section.entryCount}{section.balancedCount !== undefined ? `, Balanced: ${section.balancedCount}` : ''})
                          </Text>
                          <Text fontSize="md" fontWeight="bold" color={sectionTotalTextColor} minW="120px" textAlign="right" flex={1}>
                            {formatCurrency(section.total || 0)}
                          </Text>
                          <Text fontSize="md" fontWeight="bold" color={sectionTotalTextColor} minW="120px" textAlign="right" flex={1}>
                            {formatCurrency(section.creditTotal || 0)}
                          </Text>
                        </HStack>
                      </VStack>
                    </Box>
                  ))}
                  
                  {/* Overall Summary for Journal Entry Analysis */}
                  {previewData.financialSummary && (
                    <Box bg={summaryBg} p={4} borderRadius="md" mt={2}>
                      <HStack justify="space-between">
                        <Text fontSize="sm" color={summaryTextColor}>Total Debit</Text>
                        <Text fontSize="sm" fontWeight="bold">{formatCurrency(previewData.financialSummary.totalDebit || 0)}</Text>
                      </HStack>
                      <HStack justify="space-between">
                        <Text fontSize="sm" color={summaryTextColor}>Total Credit</Text>
                        <Text fontSize="sm" fontWeight="bold">{formatCurrency(previewData.financialSummary.totalCredit || 0)}</Text>
                      </HStack>
                      <HStack justify="space-between">
                        <Text fontSize="sm" color={summaryTextColor}>Balanced Entries</Text>
                        <Text fontSize="sm" fontWeight="bold">{previewData.financialSummary.balancedEntries || 0}</Text>
                      </HStack>
                      <HStack justify="space-between">
                        <Text fontSize="sm" color={summaryTextColor}>Unbalanced Entries</Text>
                        <Text fontSize="sm" fontWeight="bold">{previewData.financialSummary.unbalancedEntries || 0}</Text>
                      </HStack>
                      <HStack justify="space-between">
                        <Text fontSize="sm" color={summaryTextColor}>Balance Accuracy</Text>
                        <Text fontSize="sm" fontWeight="bold">{previewData.financialSummary.balanceAccuracy || '0'}%</Text>
                      </HStack>
                      
                    </Box>
                  )}
                  
                  
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
        
        {/* SSOT P&L Modal */}
        <Modal isOpen={ssotPLOpen} onClose={() => setSSOTPLOpen(false)} size="6xl">
          <ModalOverlay />
          <ModalContent bg={modalContentBg}>
            <ModalHeader>
              <HStack>
                <Icon as={FiTrendingUp} color="green.500" />
                <VStack align="start" spacing={0}>
                  <Text fontSize="lg" fontWeight="bold">
                    SSOT Profit & Loss Statement
                  </Text>
                  <Text fontSize="sm" color={previewPeriodTextColor}>
                    Real-time integration with SSOT Journal System
                  </Text>
                </VStack>
              </HStack>
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody pb={6}>
              {/* Date Range Controls */}
              <Box mb={4}>
                <HStack spacing={4} mb={4}>
                  <FormControl>
                    <FormLabel>Start Date</FormLabel>
                    <Input 
                      type="date" 
                      value={ssotStartDate} 
                      onChange={(e) => setSSOTStartDate(e.target.value)} 
                    />
                  </FormControl>
                  <FormControl>
                    <FormLabel>End Date</FormLabel>
                    <Input 
                      type="date" 
                      value={ssotEndDate} 
                      onChange={(e) => setSSOTEndDate(e.target.value)} 
                    />
                  </FormControl>
                  <Button
                    colorScheme="blue"
                    onClick={fetchSSOTPLReport}
                    isLoading={ssotPLLoading}
                    leftIcon={<FiTrendingUp />}
                    size="md"
                    mt={8}
                  >
                    Generate Report
                  </Button>
                </HStack>
              </Box>

              {ssotPLLoading && (
                <Box textAlign="center" py={8}>
                  <VStack spacing={4}>
                    <Spinner size="xl" thickness="4px" speed="0.65s" color="green.500" />
                    <VStack spacing={2}>
                      <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                        Generating SSOT P&L Report
                      </Text>
                      <Text fontSize="sm" color={loadingDescColor}>
                        Fetching real-time data from SSOT journal system...
                      </Text>
                    </VStack>
                  </VStack>
                </Box>
              )}

              {ssotPLError && (
                <Box bg="red.50" p={4} borderRadius="md" mb={4}>
                  <Text color="red.600" fontWeight="medium">Error: {ssotPLError}</Text>
                  <Button
                    mt={2}
                    size="sm"
                    colorScheme="red"
                    variant="outline"
                    onClick={fetchSSOTPLReport}
                  >
                    Retry
                  </Button>
                </Box>
              )}

              {ssotPLData && !ssotPLLoading && (
                <VStack spacing={6} align="stretch">
                  {/* Company Header */}
                  <Box textAlign="center" bg={summaryBg} p={4} borderRadius="md">
                    <Heading size="md" color={headingColor}>
                      {ssotPLData.company?.name || 'PT. Sistem Akuntansi'}
                    </Heading>
                    <Text fontSize="lg" fontWeight="semibold" mt={1}>
                      {ssotPLData.title || 'Enhanced Profit and Loss Statement'}
                    </Text>
                    <Text fontSize="sm" color={summaryTextColor}>
                      Period: {ssotPLData.period || `${ssotStartDate} - ${ssotEndDate}`}
                    </Text>
                  </Box>

                  {/* Financial Metrics Cards */}
                  {ssotPLData.financialMetrics && (
                    <SimpleGrid columns={[1, 2, 4]} spacing={4}>
                      <Box bg="green.50" p={4} borderRadius="md" textAlign="center">
                        <Text fontSize="sm" color="green.600">Gross Profit</Text>
                        <Text fontSize="xl" fontWeight="bold" color="green.700">
                          {formatCurrency(ssotPLData.financialMetrics.grossProfit || 0)}
                        </Text>
                        <Text fontSize="xs" color="green.500">
                          {ssotPLData.financialMetrics.grossProfitMargin?.toFixed(1) || '0'}%
                        </Text>
                      </Box>
                      
                      <Box bg="blue.50" p={4} borderRadius="md" textAlign="center">
                        <Text fontSize="sm" color="blue.600">Operating Income</Text>
                        <Text fontSize="xl" fontWeight="bold" color="blue.700">
                          {formatCurrency(ssotPLData.financialMetrics.operatingIncome || 0)}
                        </Text>
                        <Text fontSize="xs" color="blue.500">
                          {ssotPLData.financialMetrics.operatingMargin?.toFixed(1) || '0'}%
                        </Text>
                      </Box>
                      
                      <Box bg="purple.50" p={4} borderRadius="md" textAlign="center">
                        <Text fontSize="sm" color="purple.600">Net Income</Text>
                        <Text fontSize="xl" fontWeight="bold" color={ssotPLData.financialMetrics.netIncome >= 0 ? "green.700" : "red.700"}>
                          {formatCurrency(ssotPLData.financialMetrics.netIncome || 0)}
                        </Text>
                        <Text fontSize="xs" color="purple.500">
                          {ssotPLData.financialMetrics.netIncomeMargin?.toFixed(1) || '0'}%
                        </Text>
                      </Box>
                      
                      <Box bg="yellow.50" p={4} borderRadius="md" textAlign="center">
                        <Text fontSize="sm" color="yellow.700">EBITDA</Text>
                        <Text fontSize="xl" fontWeight="bold" color="yellow.800">
                          {formatCurrency(ssotPLData.financialMetrics.ebitda || ssotPLData.financialMetrics.operatingIncome || 0)}
                        </Text>
                        <Text fontSize="xs" color="yellow.600">
                          {(ssotPLData.financialMetrics.ebitdaMargin || ssotPLData.financialMetrics.operatingMargin)?.toFixed(1) || '0'}%
                        </Text>
                      </Box>
                    </SimpleGrid>
                  )}

                  {/* P&L Sections */}
                  {ssotPLData.sections?.map((section: any, sectionIndex: number) => (
                    <Box key={sectionIndex}>
                      <HStack justify="space-between" mb={4} borderBottom="2px" borderColor={sectionBorderColor} pb={2}>
                        <Heading size="md" color={headingColor}>
                          {section.name}
                        </Heading>
                        <Text fontSize="lg" fontWeight="bold" color={section.total >= 0 ? "green.600" : "red.600"}>
                          {formatCurrency(section.total || 0)}
                        </Text>
                      </HStack>
                      
                      {section.is_calculated && (
                        <Badge colorScheme="blue" mb={2}>Calculated</Badge>
                      )}

                      <VStack spacing={2} align="stretch">
                        {/* Direct Items */}
                        {section.items?.map((item: any, itemIndex: number) => (
                          <HStack key={itemIndex} justify="space-between" py={2} px={4} 
                                 bg={itemIndex % 2 === 0 ? evenRowBg : oddRowBg} 
                                 borderRadius="md">
                            <HStack>
                              <Text fontSize="sm" fontWeight="medium" color={textColor}>
                                {item.name}
                              </Text>
                              {item.account_code && (
                                <Badge variant="outline" fontSize="xs">
                                  {item.account_code}
                                </Badge>
                              )}
                            </HStack>
                            <Text fontSize="sm" fontWeight="semibold" 
                                 color={item.is_percentage ? "blue.600" : (item.amount >= 0 ? "green.600" : "red.600")}>
                              {item.is_percentage ? `${item.amount.toFixed(1)}%` : formatCurrency(item.amount || 0)}
                            </Text>
                          </HStack>
                        ))}

                        {/* Subsections */}
                        {section.subsections?.map((subsection: any, subIndex: number) => (
                          <Box key={subIndex} bg={summaryBg} p={4} borderRadius="md" mt={2}>
                            <HStack justify="space-between" mb={3}>
                              <Text fontWeight="semibold" color={headingColor}>
                                {subsection.name}
                              </Text>
                              <Text fontWeight="bold" color={subsection.total >= 0 ? "green.600" : "red.600"}>
                                {formatCurrency(subsection.total || 0)}
                              </Text>
                            </HStack>
                            {subsection.items?.map((item: any, itemIndex: number) => (
                              <HStack key={itemIndex} justify="space-between" py={1}>
                                <HStack>
                                  <Text fontSize="xs">{item.name}</Text>
                                  {item.account_code && (
                                    <Badge variant="outline" fontSize="xs">
                                      {item.account_code}
                                    </Badge>
                                  )}
                                </HStack>
                                <Text fontSize="xs" fontWeight="medium">
                                  {formatCurrency(item.amount || 0)}
                                </Text>
                              </HStack>
                            ))}
                          </Box>
                        ))}
                      </VStack>
                    </Box>
                  ))}

                  {/* Analysis Message */}
                  {ssotPLData.message && (
                    <Box bg="blue.50" p={4} borderRadius="md" border="1px" borderColor="blue.200">
                      <Text fontSize="sm" color="blue.800">
                        <strong>Analysis:</strong> {ssotPLData.message}
                      </Text>
                    </Box>
                  )}

                  {/* No Data Message */}
                  {!ssotPLData.hasData && (
                    <Box bg="yellow.50" p={6} borderRadius="md" textAlign="center">
                      <Icon as={FiTrendingUp} boxSize={12} color="yellow.400" mb={4} />
                      <Text fontSize="lg" fontWeight="semibold" color="yellow.800" mb={2}>
                        No Data Available
                      </Text>
                      <Text fontSize="sm" color="yellow.700">
                        No P&L relevant transactions found for this period. The journal entries contain mainly asset purchases, payments, and deposits which affect the Balance Sheet rather than P&L.
                      </Text>
                      <Text fontSize="sm" color="yellow.700" mt={2}>
                        To generate meaningful P&L data, record sales transactions, operating expenses, and cost of goods sold.
                      </Text>
                    </Box>
                  )}
                </VStack>
              )}
            </ModalBody>
            <ModalFooter>
              <Button variant="ghost" mr={3} onClick={() => setSSOTPLOpen(false)}>
                Close
              </Button>
              <Button 
                colorScheme="green" 
                onClick={() => {
                  // Export functionality can be added here
                  toast({
                    title: 'Export Feature',
                    description: 'Export functionality will be implemented next.',
                    status: 'info',
                    duration: 3000,
                    isClosable: true,
                  });
                }}
                leftIcon={<FiDownload />}
              >
                Export PDF
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>
        
        {/* SSOT Balance Sheet Modal */}
        <Modal isOpen={ssotBSOpen} onClose={() => setSSOTBSOpen(false)} size="6xl">
          <ModalOverlay />
          <ModalContent bg={modalContentBg}>
            <ModalHeader>
              <HStack>
                <Icon as={FiBarChart} color="blue.500" />
                <VStack align="start" spacing={0}>
                  <Text fontSize="lg" fontWeight="bold">
                    SSOT Balance Sheet
                  </Text>
                  <Text fontSize="sm" color={previewPeriodTextColor}>
                    Real-time integration with SSOT Journal System
                  </Text>
                </VStack>
              </HStack>
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody pb={6}>
              {/* As Of Date Control */}
              <Box mb={4}>
                <HStack spacing={4} mb={4}>
                  <FormControl>
                    <FormLabel>As Of Date</FormLabel>
                    <Input 
                      type="date" 
                      value={ssotAsOfDate} 
                      onChange={(e) => setSSOTAsOfDate(e.target.value)} 
                    />
                  </FormControl>
                  <Button
                    colorScheme="blue"
                    onClick={fetchSSOTBalanceSheetReport}
                    isLoading={ssotBSLoading}
                    leftIcon={<FiBarChart />}
                    size="md"
                    mt={8}
                  >
                    Generate Report
                  </Button>
                </HStack>
              </Box>

              {ssotBSLoading && (
                <Box textAlign="center" py={8}>
                  <VStack spacing={4}>
                    <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
                    <VStack spacing={2}>
                      <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                        Generating SSOT Balance Sheet
                      </Text>
                      <Text fontSize="sm" color={loadingDescColor}>
                        Fetching real-time data from SSOT journal system...
                      </Text>
                    </VStack>
                  </VStack>
                </Box>
              )}

              {ssotBSError && (
                <Box bg="red.50" p={4} borderRadius="md" mb={4}>
                  <Text color="red.600" fontWeight="medium">Error: {ssotBSError}</Text>
                  <Button
                    mt={2}
                    size="sm"
                    colorScheme="red"
                    variant="outline"
                    onClick={fetchSSOTBalanceSheetReport}
                  >
                    Retry
                  </Button>
                </Box>
              )}

              {ssotBSData && !ssotBSLoading && (
                <VStack spacing={6} align="stretch">
                  {/* Company Header */}
                  <Box textAlign="center" bg={summaryBg} p={4} borderRadius="md">
                    <Heading size="md" color={headingColor}>
                      {ssotBSData.company?.name || 'PT. Sistem Akuntansi'}
                    </Heading>
                    <Text fontSize="lg" fontWeight="semibold" mt={1}>
                      Enhanced Balance Sheet (SSOT)
                    </Text>
                    <Text fontSize="sm" color={summaryTextColor}>
                      As of: {new Date(ssotBSData.as_of_date).toLocaleDateString('id-ID')}
                    </Text>
                    {ssotBSData.is_balanced ? (
                      <Badge colorScheme="green" mt={2}>Balanced ✓</Badge>
                    ) : (
                      <Badge colorScheme="red" mt={2}>Not Balanced (Diff: {formatCurrency(ssotBSData.balance_difference || 0)})</Badge>
                    )}
                  </Box>

                  {/* Balance Summary Cards */}
                  <SimpleGrid columns={[1, 2, 3]} spacing={4}>
                    <Box bg="green.50" p={4} borderRadius="md" textAlign="center">
                      <Text fontSize="sm" color="green.600">Total Assets</Text>
                      <Text fontSize="xl" fontWeight="bold" color="green.700">
                        {formatCurrency(ssotBSData.assets?.total_assets || 0)}
                      </Text>
                    </Box>
                    
                    <Box bg="orange.50" p={4} borderRadius="md" textAlign="center">
                      <Text fontSize="sm" color="orange.600">Total Liabilities</Text>
                      <Text fontSize="xl" fontWeight="bold" color="orange.700">
                        {formatCurrency(ssotBSData.liabilities?.total_liabilities || 0)}
                      </Text>
                    </Box>
                    
                    <Box bg="blue.50" p={4} borderRadius="md" textAlign="center">
                      <Text fontSize="sm" color="blue.600">Total Equity</Text>
                      <Text fontSize="xl" fontWeight="bold" color="blue.700">
                        {formatCurrency(ssotBSData.equity?.total_equity || 0)}
                      </Text>
                    </Box>
                  </SimpleGrid>

                  {/* Assets Section */}
                  {ssotBSData.assets && (
                    <Box>
                      <HStack justify="space-between" mb={4} borderBottom="2px" borderColor={sectionBorderColor} pb={2}>
                        <Heading size="md" color={headingColor}>
                          ASSETS
                        </Heading>
                        <Text fontSize="lg" fontWeight="bold" color="green.600">
                          {formatCurrency(ssotBSData.assets.total_assets || 0)}
                        </Text>
                      </HStack>
                      
                      <VStack spacing={4} align="stretch">
                        {/* Current Assets */}
                        {ssotBSData.assets.current_assets && (
                          <Box bg={summaryBg} p={4} borderRadius="md">
                            <HStack justify="space-between" mb={3}>
                              <Text fontWeight="semibold" color={headingColor}>
                                Current Assets
                              </Text>
                              <Text fontWeight="bold" color="green.600">
                                {formatCurrency(ssotBSData.assets.current_assets.total_current_assets || 0)}
                              </Text>
                            </HStack>
                            {ssotBSData.assets.current_assets.items?.map((item: any, itemIndex: number) => (
                              <HStack key={itemIndex} justify="space-between" py={1}>
                                <HStack>
                                  <Text fontSize="sm">{item.account_code} - {item.account_name}</Text>
                                </HStack>
                                <Text fontSize="sm" fontWeight="medium">
                                  {formatCurrency(item.amount || 0)}
                                </Text>
                              </HStack>
                            ))}
                          </Box>
                        )}
                        
                        {/* Non-Current Assets */}
                        {ssotBSData.assets.non_current_assets && (
                          <Box bg={summaryBg} p={4} borderRadius="md">
                            <HStack justify="space-between" mb={3}>
                              <Text fontWeight="semibold" color={headingColor}>
                                Non-Current Assets
                              </Text>
                              <Text fontWeight="bold" color="green.600">
                                {formatCurrency(ssotBSData.assets.non_current_assets.total_non_current_assets || 0)}
                              </Text>
                            </HStack>
                            {ssotBSData.assets.non_current_assets.items?.map((item: any, itemIndex: number) => (
                              <HStack key={itemIndex} justify="space-between" py={1}>
                                <HStack>
                                  <Text fontSize="sm">{item.account_code} - {item.account_name}</Text>
                                </HStack>
                                <Text fontSize="sm" fontWeight="medium">
                                  {formatCurrency(item.amount || 0)}
                                </Text>
                              </HStack>
                            ))}
                          </Box>
                        )}
                      </VStack>
                    </Box>
                  )}

                  {/* Liabilities Section */}
                  {ssotBSData.liabilities && (
                    <Box>
                      <HStack justify="space-between" mb={4} borderBottom="2px" borderColor={sectionBorderColor} pb={2}>
                        <Heading size="md" color={headingColor}>
                          LIABILITIES
                        </Heading>
                        <Text fontSize="lg" fontWeight="bold" color="orange.600">
                          {formatCurrency(ssotBSData.liabilities.total_liabilities || 0)}
                        </Text>
                      </HStack>
                      
                      <VStack spacing={4} align="stretch">
                        {/* Current Liabilities */}
                        {ssotBSData.liabilities.current_liabilities && (
                          <Box bg={summaryBg} p={4} borderRadius="md">
                            <HStack justify="space-between" mb={3}>
                              <Text fontWeight="semibold" color={headingColor}>
                                Current Liabilities
                              </Text>
                              <Text fontWeight="bold" color="orange.600">
                                {formatCurrency(ssotBSData.liabilities.current_liabilities.total_current_liabilities || 0)}
                              </Text>
                            </HStack>
                            {ssotBSData.liabilities.current_liabilities.items?.map((item: any, itemIndex: number) => (
                              <HStack key={itemIndex} justify="space-between" py={1}>
                                <HStack>
                                  <Text fontSize="sm">{item.account_code} - {item.account_name}</Text>
                                </HStack>
                                <Text fontSize="sm" fontWeight="medium">
                                  {formatCurrency(item.amount || 0)}
                                </Text>
                              </HStack>
                            ))}
                          </Box>
                        )}
                        
                        {/* Non-Current Liabilities */}
                        {ssotBSData.liabilities.non_current_liabilities && (
                          <Box bg={summaryBg} p={4} borderRadius="md">
                            <HStack justify="space-between" mb={3}>
                              <Text fontWeight="semibold" color={headingColor}>
                                Non-Current Liabilities
                              </Text>
                              <Text fontWeight="bold" color="orange.600">
                                {formatCurrency(ssotBSData.liabilities.non_current_liabilities.total_non_current_liabilities || 0)}
                              </Text>
                            </HStack>
                            {ssotBSData.liabilities.non_current_liabilities.items?.map((item: any, itemIndex: number) => (
                              <HStack key={itemIndex} justify="space-between" py={1}>
                                <HStack>
                                  <Text fontSize="sm">{item.account_code} - {item.account_name}</Text>
                                </HStack>
                                <Text fontSize="sm" fontWeight="medium">
                                  {formatCurrency(item.amount || 0)}
                                </Text>
                              </HStack>
                            ))}
                          </Box>
                        )}
                      </VStack>
                    </Box>
                  )}

                  {/* Equity Section */}
                  {ssotBSData.equity && (
                    <Box>
                      <HStack justify="space-between" mb={4} borderBottom="2px" borderColor={sectionBorderColor} pb={2}>
                        <Heading size="md" color={headingColor}>
                          EQUITY
                        </Heading>
                        <Text fontSize="lg" fontWeight="bold" color="blue.600">
                          {formatCurrency(ssotBSData.equity.total_equity || 0)}
                        </Text>
                      </HStack>
                      
                      <VStack spacing={2} align="stretch">
                        {ssotBSData.equity.items?.map((item: any, itemIndex: number) => (
                          <HStack key={itemIndex} justify="space-between" py={2} px={4} 
                                 bg={itemIndex % 2 === 0 ? evenRowBg : oddRowBg} 
                                 borderRadius="md">
                            <HStack>
                              <Text fontSize="sm" fontWeight="medium" color={textColor}>
                                {item.account_code} - {item.account_name}
                              </Text>
                            </HStack>
                            <Text fontSize="sm" fontWeight="semibold" color="blue.600">
                              {formatCurrency(item.amount || 0)}
                            </Text>
                          </HStack>
                        ))}
                      </VStack>
                    </Box>
                  )}
                </VStack>
              )}
            </ModalBody>
            <ModalFooter>
              <Button variant="ghost" mr={3} onClick={() => setSSOTBSOpen(false)}>
                Close
              </Button>
              <Button 
                colorScheme="blue" 
                onClick={() => {
                  // Export functionality can be added here
                  toast({
                    title: 'Export Feature',
                    description: 'Export functionality will be implemented next.',
                    status: 'info',
                    duration: 3000,
                    isClosable: true,
                  });
                }}
                leftIcon={<FiDownload />}
              >
                Export PDF
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* SSOT Cash Flow Modal */}
        <Modal isOpen={ssotCFOpen} onClose={() => setSSOTCFOpen(false)} size="6xl">
          <ModalOverlay />
          <ModalContent bg={modalContentBg}>
            <ModalHeader bg={modalHeaderBg}>
              <HStack>
                <Icon as={FiActivity} color="blue.500" />
                <VStack align="start" spacing={0}>
                  <Text fontSize="lg" fontWeight="bold">
                    SSOT Cash Flow Statement
                  </Text>
                  <Text fontSize="sm" color={previewPeriodTextColor}>
                    {ssotCFStartDate} - {ssotCFEndDate} | SSOT Journal Integration
                  </Text>
                </VStack>
              </HStack>
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody pb={6}>
              <VStack spacing={4} align="stretch">
                {/* Date Range Selector */}
                <Box bg={summaryBg} p={4} borderRadius="md">
                  <HStack spacing={4}>
                    <FormControl flex={1}>
                      <FormLabel fontSize="sm">Start Date</FormLabel>
                      <Input 
                        type="date" 
                        value={ssotCFStartDate} 
                        onChange={(e) => setSSOTCFStartDate(e.target.value)}
                        size="sm"
                      />
                    </FormControl>
                    <FormControl flex={1}>
                      <FormLabel fontSize="sm">End Date</FormLabel>
                      <Input 
                        type="date" 
                        value={ssotCFEndDate} 
                        onChange={(e) => setSSOTCFEndDate(e.target.value)}
                        size="sm"
                      />
                    </FormControl>
                    <Button 
                      colorScheme="blue" 
                      onClick={fetchSSOTCashFlowReport}
                      isLoading={ssotCFLoading}
                      size="sm"
                      alignSelf="flex-end"
                    >
                      Generate
                    </Button>
                  </HStack>
                </Box>

                {/* Loading State */}
                {ssotCFLoading && (
                  <Box textAlign="center" py={8}>
                    <VStack spacing={4}>
                      <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
                      <VStack spacing={2}>
                        <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                          Generating SSOT Cash Flow Statement
                        </Text>
                        <Text fontSize="sm" color={loadingDescColor}>
                          Analyzing journal entries for cash flow activities...
                        </Text>
                      </VStack>
                    </VStack>
                  </Box>
                )}

                {/* Error State */}
                {ssotCFError && !ssotCFLoading && (
                  <Box textAlign="center" py={8}>
                    <VStack spacing={4}>
                      <Icon as={FiActivity} boxSize={12} color={errorIconColor} />
                      <VStack spacing={3}>
                        <Text fontSize="lg" fontWeight="bold" color={errorTextColor}>
                          Unable to Generate Cash Flow Statement
                        </Text>
                        <Text fontSize="md" color={textColor} maxW="lg" textAlign="center">
                          {ssotCFError}
                        </Text>
                        <Button 
                          colorScheme="blue" 
                          variant="outline"
                          onClick={fetchSSOTCashFlowReport}
                        >
                          Retry
                        </Button>
                      </VStack>
                    </VStack>
                  </Box>
                )}

                {/* Cash Flow Data */}
                {ssotCFData && !ssotCFLoading && (
                  <VStack spacing={6} align="stretch">
                    {/* Cash Flow Summary */}
                    <Box bg={sectionTotalBg} p={4} borderRadius="md" borderWidth="2px" borderColor={sectionTotalBorderColor}>
                      <HStack justify="space-between">
                        <VStack align="start" spacing={1}>
                          <Text fontSize="lg" fontWeight="bold" color={sectionTotalTextColor}>
                            Cash Position Summary
                          </Text>
                          <Text fontSize="sm" color={summaryTextColor}>
                            {ssotCFData.start_date} to {ssotCFData.end_date}
                          </Text>
                        </VStack>
                        <VStack align="end" spacing={1}>
                          <Text fontSize="sm" color={summaryTextColor}>Net Cash Flow</Text>
                          <Text fontSize="xl" fontWeight="bold" color={ssotCFData.net_cash_flow >= 0 ? 'green.600' : 'red.600'}>
                            {formatCurrency(ssotCFData.net_cash_flow)}
                          </Text>
                        </VStack>
                      </HStack>
                      
                      <HStack justify="space-between" mt={4}>
                        <VStack>
                          <Text fontSize="sm" color={summaryTextColor}>Cash at Beginning</Text>
                          <Text fontSize="md" fontWeight="semibold">
                            {formatCurrency(ssotCFData.cash_at_beginning)}
                          </Text>
                        </VStack>
                        <VStack>
                          <Text fontSize="sm" color={summaryTextColor}>Cash at End</Text>
                          <Text fontSize="md" fontWeight="semibold">
                            {formatCurrency(ssotCFData.cash_at_end)}
                          </Text>
                        </VStack>
                      </HStack>
                    </Box>

                    {/* Operating Activities */}
                    {ssotCFData.operating_activities && (
                      <Box>
                        <HStack justify="space-between" mb={4} borderBottom="2px" borderColor={sectionBorderColor} pb={2}>
                          <Heading size="md" color={headingColor}>
                            OPERATING ACTIVITIES
                          </Heading>
                          <Text fontSize="lg" fontWeight="bold" color={ssotCFData.operating_activities.total_operating_cash_flow >= 0 ? 'green.600' : 'red.600'}>
                            {formatCurrency(ssotCFData.operating_activities.total_operating_cash_flow)}
                          </Text>
                        </HStack>
                        
                        <VStack spacing={4} align="stretch">
                          {/* Net Income */}
                          <HStack justify="space-between" py={2} px={4} bg={summaryBg} borderRadius="md">
                            <Text fontWeight="medium">Net Income</Text>
                            <Text fontWeight="semibold">
                              {formatCurrency(ssotCFData.operating_activities.net_income)}
                            </Text>
                          </HStack>
                          
                          {/* Non-cash Adjustments */}
                          {ssotCFData.operating_activities.adjustments.items.length > 0 && (
                            <Box>
                              <Text fontWeight="semibold" mb={2} color={headingColor}>Adjustments for Non-Cash Items:</Text>
                              <VStack spacing={1} align="stretch">
                                {ssotCFData.operating_activities.adjustments.items.map((item: any, index: number) => (
                                  <HStack key={index} justify="space-between" py={1} px={2}>
                                    <Text fontSize="sm">{item.account_name} ({item.account_code})</Text>
                                    <Text fontSize="sm" fontWeight="medium">
                                      {formatCurrency(item.amount)}
                                    </Text>
                                  </HStack>
                                ))}
                                <HStack justify="space-between" pt={2} borderTop="1px" borderColor={borderColor}>
                                  <Text fontWeight="semibold">Total Adjustments:</Text>
                                  <Text fontWeight="semibold">
                                    {formatCurrency(ssotCFData.operating_activities.adjustments.total_adjustments)}
                                  </Text>
                                </HStack>
                              </VStack>
                            </Box>
                          )}
                          
                          {/* Working Capital Changes */}
                          {ssotCFData.operating_activities.working_capital_changes.items.length > 0 && (
                            <Box>
                              <Text fontWeight="semibold" mb={2} color={headingColor}>Changes in Working Capital:</Text>
                              <VStack spacing={1} align="stretch">
                                {ssotCFData.operating_activities.working_capital_changes.items.map((item: any, index: number) => (
                                  <HStack key={index} justify="space-between" py={1} px={2}>
                                    <HStack>
                                      <Text fontSize="sm">{item.account_name} ({item.account_code})</Text>
                                      <Text fontSize="xs" color={summaryTextColor}>({item.type})</Text>
                                    </HStack>
                                    <Text fontSize="sm" fontWeight="medium">
                                      {formatCurrency(item.amount)}
                                    </Text>
                                  </HStack>
                                ))}
                                <HStack justify="space-between" pt={2} borderTop="1px" borderColor={borderColor}>
                                  <Text fontWeight="semibold">Total Working Capital Changes:</Text>
                                  <Text fontWeight="semibold">
                                    {formatCurrency(ssotCFData.operating_activities.working_capital_changes.total_working_capital_changes)}
                                  </Text>
                                </HStack>
                              </VStack>
                            </Box>
                          )}
                        </VStack>
                      </Box>
                    )}

                    {/* Investing Activities */}
                    {ssotCFData.investing_activities && ssotCFData.investing_activities.items.length > 0 && (
                      <Box>
                        <HStack justify="space-between" mb={4} borderBottom="2px" borderColor={sectionBorderColor} pb={2}>
                          <Heading size="md" color={headingColor}>
                            INVESTING ACTIVITIES
                          </Heading>
                          <Text fontSize="lg" fontWeight="bold" color={ssotCFData.investing_activities.total_investing_cash_flow >= 0 ? 'green.600' : 'red.600'}>
                            {formatCurrency(ssotCFData.investing_activities.total_investing_cash_flow)}
                          </Text>
                        </HStack>
                        
                        <VStack spacing={1} align="stretch">
                          {ssotCFData.investing_activities.items.map((item: any, index: number) => (
                            <HStack key={index} justify="space-between" py={2} px={4} bg={evenRowBg} borderRadius="md">
                              <HStack>
                                <Text fontSize="sm" fontWeight="medium">{item.account_name} ({item.account_code})</Text>
                                <Text fontSize="xs" color={summaryTextColor}>({item.type})</Text>
                              </HStack>
                              <Text fontSize="sm" fontWeight="semibold">
                                {formatCurrency(item.amount)}
                              </Text>
                            </HStack>
                          ))}
                        </VStack>
                      </Box>
                    )}

                    {/* Financing Activities */}
                    {ssotCFData.financing_activities && ssotCFData.financing_activities.items.length > 0 && (
                      <Box>
                        <HStack justify="space-between" mb={4} borderBottom="2px" borderColor={sectionBorderColor} pb={2}>
                          <Heading size="md" color={headingColor}>
                            FINANCING ACTIVITIES
                          </Heading>
                          <Text fontSize="lg" fontWeight="bold" color={ssotCFData.financing_activities.total_financing_cash_flow >= 0 ? 'green.600' : 'red.600'}>
                            {formatCurrency(ssotCFData.financing_activities.total_financing_cash_flow)}
                          </Text>
                        </HStack>
                        
                        <VStack spacing={1} align="stretch">
                          {ssotCFData.financing_activities.items.map((item: any, index: number) => (
                            <HStack key={index} justify="space-between" py={2} px={4} bg={evenRowBg} borderRadius="md">
                              <HStack>
                                <Text fontSize="sm" fontWeight="medium">{item.account_name} ({item.account_code})</Text>
                                <Text fontSize="xs" color={summaryTextColor}>({item.type})</Text>
                              </HStack>
                              <Text fontSize="sm" fontWeight="semibold">
                                {formatCurrency(item.amount)}
                              </Text>
                            </HStack>
                          ))}
                        </VStack>
                      </Box>
                    )}

                    {/* Cash Flow Ratios */}
                    {ssotCFData.cash_flow_ratios && (
                      <Box bg={summaryBg} p={4} borderRadius="md">
                        <Text fontSize="md" fontWeight="bold" mb={3} color={headingColor}>Financial Ratios</Text>
                        <HStack justify="space-between">
                          <VStack align="start">
                            <Text fontSize="sm" color={summaryTextColor}>Operating Cash Flow Ratio</Text>
                            <Text fontWeight="semibold">{ssotCFData.cash_flow_ratios.operating_cash_flow_ratio.toFixed(2)}</Text>
                          </VStack>
                          <VStack align="end">
                            <Text fontSize="sm" color={summaryTextColor}>Free Cash Flow</Text>
                            <Text fontWeight="semibold">{formatCurrency(ssotCFData.cash_flow_ratios.free_cash_flow)}</Text>
                          </VStack>
                        </HStack>
                      </Box>
                    )}
                    
                    {/* Data Source Info */}
                    <Box bg={summaryBg} p={3} borderRadius="md">
                      <HStack justify="space-between">
                        <VStack align="start" spacing={0}>
                          <Text fontSize="sm" fontWeight="medium" color={headingColor}>Data Source: {ssotCFData.data_source}</Text>
                          <Text fontSize="xs" color={summaryTextColor}>Generated at: {new Date(ssotCFData.generated_at).toLocaleString('id-ID')}</Text>
                        </VStack>
                        <Text fontSize="xs" color={summaryTextColor}>Enhanced: {ssotCFData.enhanced ? 'Yes' : 'No'}</Text>
                      </HStack>
                      {ssotCFData.message && (
                        <Text fontSize="sm" color={summaryTextColor} mt={2} fontStyle="italic">
                          {ssotCFData.message}
                        </Text>
                      )}
                    </Box>
                  </VStack>
                )}
              </VStack>
            </ModalBody>
            <ModalFooter>
              <Button variant="ghost" mr={3} onClick={() => setSSOTCFOpen(false)}>
                Close
              </Button>
              <Button 
                colorScheme="green" 
                onClick={() => {
                  // Export functionality can be added here
                  toast({
                    title: 'Export Feature',
                    description: 'Cash Flow export functionality will be implemented next.',
                    status: 'info',
                    duration: 3000,
                    isClosable: true,
                  });
                }}
                leftIcon={<FiDownload />}
              >
                Export PDF
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* SSOT Sales Summary Modal */}
        <Modal isOpen={ssotSSOpen} onClose={() => setSSOTSSOpen(false)} size="6xl">
          <ModalOverlay />
          <ModalContent bg={modalContentBg}>
            <ModalHeader>
              <HStack>
                <Icon as={FiShoppingCart} color="blue.500" />
                <VStack align="start" spacing={0}>
                  <Text fontSize="lg" fontWeight="bold">
                    SSOT Sales Summary Report
                  </Text>
                  <Text fontSize="sm" color={previewPeriodTextColor}>
                    Comprehensive sales analysis with SSOT Journal integration
                  </Text>
                </VStack>
              </HStack>
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody pb={6}>
              {/* Date Range Controls */}
              <Box mb={4}>
                <HStack spacing={4} mb={4}>
                  <FormControl>
                    <FormLabel>Start Date</FormLabel>
                    <Input 
                      type="date" 
                      value={ssotSSStartDate} 
                      onChange={(e) => setSSOTSSStartDate(e.target.value)} 
                    />
                  </FormControl>
                  <FormControl>
                    <FormLabel>End Date</FormLabel>
                    <Input 
                      type="date" 
                      value={ssotSSEndDate} 
                      onChange={(e) => setSSOTSSEndDate(e.target.value)} 
                    />
                  </FormControl>
                  <Button
                    colorScheme="blue"
                    onClick={fetchSSOTSalesSummaryReport}
                    isLoading={ssotSSLoading}
                    leftIcon={<FiShoppingCart />}
                    size="md"
                    mt={8}
                  >
                    Generate Report
                  </Button>
                </HStack>
              </Box>

              {ssotSSLoading && (
                <Box textAlign="center" py={8}>
                  <VStack spacing={4}>
                    <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
                    <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                      Generating SSOT Sales Summary
                    </Text>
                  </VStack>
                </Box>
              )}

              {ssotSSError && (
                <Box bg="red.50" p={4} borderRadius="md" mb={4}>
                  <Text color="red.600" fontWeight="medium">Error: {ssotSSError}</Text>
                </Box>
              )}

              {ssotSSData && !ssotSSLoading && (
                <VStack spacing={6} align="stretch">
                  <Box bg={summaryBg} p={4} borderRadius="md">
                    <Text fontSize="lg" fontWeight="bold" mb={2}>Sales Overview</Text>
                    <SimpleGrid columns={[1, 2, 4]} spacing={4}>
                      <Box textAlign="center">
                        <Text fontSize="sm" color="gray.600">Total Revenue</Text>
                        <Text fontSize="xl" fontWeight="bold" color="green.600">
                          {formatCurrency(ssotSSData.total_revenue)}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color="gray.600">Net Revenue</Text>
                        <Text fontSize="xl" fontWeight="bold" color="blue.600">
                          {formatCurrency(ssotSSData.net_revenue)}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color="gray.600">Total Sales</Text>
                        <Text fontSize="xl" fontWeight="bold">
                          {formatCurrency(ssotSSData.total_sales)}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color="gray.600">Discounts</Text>
                        <Text fontSize="xl" fontWeight="bold" color="orange.600">
                          {formatCurrency(ssotSSData.total_discounts)}
                        </Text>
                      </Box>
                    </SimpleGrid>
                  </Box>
                </VStack>
              )}
            </ModalBody>
            <ModalFooter>
              <Button variant="ghost" mr={3} onClick={() => setSSOTSSOpen(false)}>
                Close
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* SSOT Vendor Analysis Modal */}
        <Modal isOpen={ssotVAOpen} onClose={() => setSSOTVAOpen(false)} size="6xl">
          <ModalOverlay />
          <ModalContent bg={modalContentBg}>
            <ModalHeader>
              <HStack>
                <Icon as={FiShoppingCart} color="purple.500" />
                <VStack align="start" spacing={0}>
                  <Text fontSize="lg" fontWeight="bold">
                    SSOT Vendor Analysis Report
                  </Text>
                  <Text fontSize="sm" color={previewPeriodTextColor}>
                    Comprehensive vendor performance analysis
                  </Text>
                </VStack>
              </HStack>
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody pb={6}>
              {/* Date Range Controls */}
              <Box mb={4}>
                <HStack spacing={4} mb={4}>
                  <FormControl>
                    <FormLabel>Start Date</FormLabel>
                    <Input 
                      type="date" 
                      value={ssotVAStartDate} 
                      onChange={(e) => setSSOTVAStartDate(e.target.value)} 
                    />
                  </FormControl>
                  <FormControl>
                    <FormLabel>End Date</FormLabel>
                    <Input 
                      type="date" 
                      value={ssotVAEndDate} 
                      onChange={(e) => setSSOTVAEndDate(e.target.value)} 
                    />
                  </FormControl>
                  <Button
                    colorScheme="purple"
                    onClick={fetchSSOTVendorAnalysisReport}
                    isLoading={ssotVALoading}
                    leftIcon={<FiShoppingCart />}
                    size="md"
                    mt={8}
                  >
                    Generate Report
                  </Button>
                </HStack>
              </Box>

              {ssotVALoading && (
                <Box textAlign="center" py={8}>
                  <VStack spacing={4}>
                    <Spinner size="xl" thickness="4px" speed="0.65s" color="purple.500" />
                    <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                      Generating SSOT Vendor Analysis
                    </Text>
                  </VStack>
                </Box>
              )}

              {ssotVAError && (
                <Box bg="red.50" p={4} borderRadius="md" mb={4}>
                  <Text color="red.600" fontWeight="medium">Error: {ssotVAError}</Text>
                </Box>
              )}

              {ssotVAData && !ssotVALoading && (
                <VStack spacing={6} align="stretch">
                  <Box bg={summaryBg} p={4} borderRadius="md">
                    <Text fontSize="lg" fontWeight="bold" mb={2}>Vendor Overview</Text>
                    <SimpleGrid columns={[1, 2, 3]} spacing={4}>
                      <Box textAlign="center">
                        <Text fontSize="sm" color="gray.600">Total Vendors</Text>
                        <Text fontSize="xl" fontWeight="bold">
                          {ssotVAData.total_vendors}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color="gray.600">Total Purchases</Text>
                        <Text fontSize="xl" fontWeight="bold" color="blue.600">
                          {formatCurrency(ssotVAData.total_purchases)}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color="gray.600">Outstanding Payables</Text>
                        <Text fontSize="xl" fontWeight="bold" color="red.600">
                          {formatCurrency(ssotVAData.outstanding_payables)}
                        </Text>
                      </Box>
                    </SimpleGrid>
                  </Box>
                </VStack>
              )}
            </ModalBody>
            <ModalFooter>
              <Button variant="ghost" mr={3} onClick={() => setSSOTVAOpen(false)}>
                Close
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* SSOT Trial Balance Modal */}
        <Modal isOpen={ssotTBOpen} onClose={() => setSSOTTBOpen(false)} size="6xl">
          <ModalOverlay />
          <ModalContent bg={modalContentBg}>
            <ModalHeader>
              <HStack>
                <Icon as={FiList} color="green.500" />
                <VStack align="start" spacing={0}>
                  <Text fontSize="lg" fontWeight="bold">
                    SSOT Trial Balance
                  </Text>
                  <Text fontSize="sm" color={previewPeriodTextColor}>
                    Account balances verification with SSOT Journal
                  </Text>
                </VStack>
              </HStack>
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody pb={6}>
              {/* As Of Date Control */}
              <Box mb={4}>
                <HStack spacing={4} mb={4}>
                  <FormControl>
                    <FormLabel>As Of Date</FormLabel>
                    <Input 
                      type="date" 
                      value={ssotTBAsOfDate} 
                      onChange={(e) => setSSOTTBAsOfDate(e.target.value)} 
                    />
                  </FormControl>
                  <Button
                    colorScheme="green"
                    onClick={fetchSSOTTrialBalanceReport}
                    isLoading={ssotTBLoading}
                    leftIcon={<FiList />}
                    size="md"
                    mt={8}
                  >
                    Generate Report
                  </Button>
                </HStack>
              </Box>

              {ssotTBLoading && (
                <Box textAlign="center" py={8}>
                  <VStack spacing={4}>
                    <Spinner size="xl" thickness="4px" speed="0.65s" color="green.500" />
                    <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                      Generating SSOT Trial Balance
                    </Text>
                  </VStack>
                </Box>
              )}

              {ssotTBError && (
                <Box bg="red.50" p={4} borderRadius="md" mb={4}>
                  <Text color="red.600" fontWeight="medium">Error: {ssotTBError}</Text>
                </Box>
              )}

              {ssotTBData && !ssotTBLoading && (
                <VStack spacing={6} align="stretch">
                  <Box bg={summaryBg} p={4} borderRadius="md">
                    <Text fontSize="lg" fontWeight="bold" mb={2}>Trial Balance Summary</Text>
                    <HStack justify="space-between">
                      <Box textAlign="center">
                        <Text fontSize="sm" color="gray.600">Total Debits</Text>
                        <Text fontSize="xl" fontWeight="bold" color="blue.600">
                          {formatCurrency(ssotTBData.total_debits)}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color="gray.600">Total Credits</Text>
                        <Text fontSize="xl" fontWeight="bold" color="green.600">
                          {formatCurrency(ssotTBData.total_credits)}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color="gray.600">Is Balanced</Text>
                        <Text fontSize="xl" fontWeight="bold" color={ssotTBData.is_balanced ? "green.600" : "red.600"}>
                          {ssotTBData.is_balanced ? "✓" : "✗"}
                        </Text>
                      </Box>
                    </HStack>
                  </Box>
                </VStack>
              )}
            </ModalBody>
            <ModalFooter>
              <Button variant="ghost" mr={3} onClick={() => setSSOTTBOpen(false)}>
                Close
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* SSOT General Ledger Modal */}
        <Modal isOpen={ssotGLOpen} onClose={() => setSSOTGLOpen(false)} size="6xl">
          <ModalOverlay />
          <ModalContent bg={modalContentBg}>
            <ModalHeader>
              <HStack>
                <Icon as={FiBook} color="orange.500" />
                <VStack align="start" spacing={0}>
                  <Text fontSize="lg" fontWeight="bold">
                    SSOT General Ledger
                  </Text>
                  <Text fontSize="sm" color={previewPeriodTextColor}>
                    Detailed transaction records from SSOT Journal
                  </Text>
                </VStack>
              </HStack>
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody pb={6}>
              {/* Date Range and Account Controls */}
              <Box mb={4}>
                <VStack spacing={4} mb={4}>
                  <HStack spacing={4} width="full">
                    <FormControl>
                      <FormLabel>Start Date</FormLabel>
                      <Input 
                        type="date" 
                        value={ssotGLStartDate} 
                        onChange={(e) => setSSOTGLStartDate(e.target.value)} 
                      />
                    </FormControl>
                    <FormControl>
                      <FormLabel>End Date</FormLabel>
                      <Input 
                        type="date" 
                        value={ssotGLEndDate} 
                        onChange={(e) => setSSOTGLEndDate(e.target.value)} 
                      />
                    </FormControl>
                  </HStack>
                  <HStack spacing={4} width="full">
                    <FormControl>
                      <FormLabel>Account ID (Optional)</FormLabel>
                      <Input 
                        placeholder="Leave empty for all accounts" 
                        value={ssotGLAccountId} 
                        onChange={(e) => setSSOTGLAccountId(e.target.value)} 
                      />
                    </FormControl>
                    <Button
                      colorScheme="orange"
                      onClick={fetchSSOTGeneralLedgerReport}
                      isLoading={ssotGLLoading}
                      leftIcon={<FiBook />}
                      size="md"
                      mt={8}
                    >
                      Generate Report
                    </Button>
                  </HStack>
                </VStack>
              </Box>

              {ssotGLLoading && (
                <Box textAlign="center" py={8}>
                  <VStack spacing={4}>
                    <Spinner size="xl" thickness="4px" speed="0.65s" color="orange.500" />
                    <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                      Generating SSOT General Ledger
                    </Text>
                  </VStack>
                </Box>
              )}

              {ssotGLError && (
                <Box bg="red.50" p={4} borderRadius="md" mb={4}>
                  <Text color="red.600" fontWeight="medium">Error: {ssotGLError}</Text>
                </Box>
              )}

              {ssotGLData && !ssotGLLoading && (
                <VStack spacing={6} align="stretch">
                  <Box bg={summaryBg} p={4} borderRadius="md">
                    <Text fontSize="lg" fontWeight="bold" mb={2}>General Ledger Summary</Text>
                    <Text fontSize="sm" color="gray.600">
                      Total Transactions: {ssotGLData.transactions?.length || 0}
                    </Text>
                  </Box>
                </VStack>
              )}
            </ModalBody>
            <ModalFooter>
              <Button variant="ghost" mr={3} onClick={() => setSSOTGLOpen(false)}>
                Close
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* SSOT Journal Analysis Modal */}
        <Modal isOpen={ssotJAOpen} onClose={() => setSSOTJAOpen(false)} size="6xl">
          <ModalOverlay />
          <ModalContent bg={modalContentBg}>
            <ModalHeader>
              <HStack>
                <Icon as={FiDatabase} color="teal.500" />
                <VStack align="start" spacing={0}>
                  <Text fontSize="lg" fontWeight="bold">
                    SSOT Journal Entry Analysis
                  </Text>
                  <Text fontSize="sm" color={previewPeriodTextColor}>
                    Comprehensive journal entries analysis and compliance
                  </Text>
                </VStack>
              </HStack>
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody pb={6}>
              {/* Date Range Controls */}
              <Box mb={4}>
                <HStack spacing={4} mb={4}>
                  <FormControl>
                    <FormLabel>Start Date</FormLabel>
                    <Input 
                      type="date" 
                      value={ssotJAStartDate} 
                      onChange={(e) => setSSOTJAStartDate(e.target.value)} 
                    />
                  </FormControl>
                  <FormControl>
                    <FormLabel>End Date</FormLabel>
                    <Input 
                      type="date" 
                      value={ssotJAEndDate} 
                      onChange={(e) => setSSOTJAEndDate(e.target.value)} 
                    />
                  </FormControl>
                  <Button
                    colorScheme="teal"
                    onClick={fetchSSOTJournalAnalysisReport}
                    isLoading={ssotJALoading}
                    leftIcon={<FiDatabase />}
                    size="md"
                    mt={8}
                  >
                    Generate Report
                  </Button>
                </HStack>
              </Box>

              {ssotJALoading && (
                <Box textAlign="center" py={8}>
                  <VStack spacing={4}>
                    <Spinner size="xl" thickness="4px" speed="0.65s" color="teal.500" />
                    <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                      Generating SSOT Journal Analysis
                    </Text>
                  </VStack>
                </Box>
              )}

              {ssotJAError && (
                <Box bg="red.50" p={4} borderRadius="md" mb={4}>
                  <Text color="red.600" fontWeight="medium">Error: {ssotJAError}</Text>
                </Box>
              )}

              {ssotJAData && !ssotJALoading && (
                <VStack spacing={6} align="stretch">
                  <Box bg={summaryBg} p={4} borderRadius="md">
                    <Text fontSize="lg" fontWeight="bold" mb={2}>Journal Analysis Summary</Text>
                    <SimpleGrid columns={[1, 2, 4]} spacing={4}>
                      <Box textAlign="center">
                        <Text fontSize="sm" color="gray.600">Total Entries</Text>
                        <Text fontSize="xl" fontWeight="bold">
                          {ssotJAData.total_entries}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color="gray.600">Posted Entries</Text>
                        <Text fontSize="xl" fontWeight="bold" color="green.600">
                          {ssotJAData.posted_entries}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color="gray.600">Draft Entries</Text>
                        <Text fontSize="xl" fontWeight="bold" color="yellow.600">
                          {ssotJAData.draft_entries}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color="gray.600">Total Amount</Text>
                        <Text fontSize="xl" fontWeight="bold" color="blue.600">
                          {formatCurrency(ssotJAData.total_amount)}
                        </Text>
                      </Box>
                    </SimpleGrid>
                  </Box>
                </VStack>
              )}
            </ModalBody>
            <ModalFooter>
              <Button variant="ghost" mr={3} onClick={() => setSSOTJAOpen(false)}>
                Close
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>
    </SimpleLayout>
  );
};

export default ReportsPage;
