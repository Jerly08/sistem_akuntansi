'use client';

import React, { useState } from 'react';
import SimpleLayout from '@/components/layout/SimpleLayout';
import { useTranslation } from '@/hooks/useTranslation';
import SalesSummaryModal from '@/components/reports/SalesSummaryModal';
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
  useColorModeValue,
  useDisclosure,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  MenuDivider,
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
  FiDatabase,
  FiFilePlus,
  FiChevronDown
} from 'react-icons/fi';
// Legacy reportService removed - now using SSOT services only
import { ssotBalanceSheetReportService, SSOTBalanceSheetData } from '../../src/services/ssotBalanceSheetReportService';
import { ssotCashFlowReportService, SSOTCashFlowData } from '../../src/services/ssotCashFlowReportService';
import { ssotSalesSummaryService, SSOTSalesSummaryData } from '../../src/services/ssotSalesSummaryService';
// Vendor Analysis removed - replaced with Purchase Report
import { ssotTrialBalanceService, SSOTTrialBalanceData } from '../../src/services/ssotTrialBalanceService';
import { ssotGeneralLedgerService, SSOTGeneralLedgerData } from '../../src/services/ssotGeneralLedgerService';
import { ssotJournalAnalysisService, SSOTJournalAnalysisData } from '../../src/services/ssotJournalAnalysisService';
import { reportService, ReportParameters } from '../../src/services/reportService';
// Import enhanced Balance Sheet export utilities
import { 
  exportAndDownloadCSV, 
  exportAndDownloadPDF 
} from '../../src/utils/balanceSheetExportUtils';
// Import Cash Flow Export Service
import cashFlowExportService from '../../src/services/cashFlowExportService';

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
    id: 'purchase-report',
    name: t('reports.purchaseReport'),
    description: 'Comprehensive purchase analysis with credible vendor transactions, payment history, and performance metrics. Real-time data from SSOT journal integration.',
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
  const summaryBg = useColorModeValue('gray.50', 'gray.700');
  const loadingTextColor = useColorModeValue('gray.700', 'gray.300');
  const previewPeriodTextColor = useColorModeValue('gray.500', 'gray.400');
  
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

  // State untuk SSOT Purchase Report
  const [ssotPROpen, setSSOTPROpen] = useState(false);
  const [ssotPRData, setSSOTPRData] = useState<any>(null);
  const [ssotPRLoading, setSSOTPRLoading] = useState(false);
  const [ssotPRError, setSSOTPRError] = useState<string | null>(null);
  const [ssotPRStartDate, setSSOTPRStartDate] = useState('2025-01-01');
  const [ssotPREndDate, setSSOTPREndDate] = useState('2025-12-31');

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

  // Modal and report generation states
  const { isOpen, onOpen, onClose } = useDisclosure();
  const [selectedReport, setSelectedReport] = useState<any>(null);
  const [reportParams, setReportParams] = useState<ReportParameters>({});
  const [previewReport, setPreviewReport] = useState<any>(null);

  const resetParams = () => {
    setReportParams({});
  };

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
      
      console.log('SSOT Sales Summary Data received:', salesSummaryData);
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

  // Function untuk fetch SSOT Purchase Report
  const fetchSSOTPurchaseReport = async () => {
    setSSOTPRLoading(true);
    setSSOTPRError(null);
    
    try {
      // Get token
      let token = null;
      
      if (typeof window !== 'undefined') {
        token = localStorage.getItem('token');
        
        if (!token) {
          token = localStorage.getItem('authToken') || 
                 sessionStorage.getItem('token') || 
                 sessionStorage.getItem('authToken');
                 
          if (token) {
            localStorage.setItem('token', token);
          }
        }
        
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
        `http://localhost:8080/api/v1/ssot-reports/purchase-report?start_date=${ssotPRStartDate}&end_date=${ssotPREndDate}&format=json`,
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
      
      if (result.success && result.data) {
        console.log('SSOT Purchase Report Data received:', result.data);
        setSSOTPRData(result.data);
        
        toast({
          title: 'Success',
          description: 'Purchase Report generated successfully',
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      } else {
        throw new Error(result.message || 'Failed to fetch purchase report data');
      }
    } catch (error: any) {
      setSSOTPRError(error.message || 'Failed to generate purchase report');
      toast({
        title: 'Error',
        description: error.message || 'Failed to generate purchase report',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setSSOTPRLoading(false);
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
      
      console.log('SSOT Trial Balance Data received:', trialBalanceData);
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
      
      console.log('SSOT General Ledger Data received:', generalLedgerData);
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
      
      console.log('SSOT Journal Analysis Data received:', journalAnalysisData);
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
      
      console.log('SSOT Balance Sheet Data received:', balanceSheetData);
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
      // Get token
      let token = null;
      
      if (typeof window !== 'undefined') {
        token = localStorage.getItem('token');
        
        if (!token) {
          token = localStorage.getItem('authToken') || 
                 sessionStorage.getItem('token') || 
                 sessionStorage.getItem('authToken');
                 
          if (token) {
            localStorage.setItem('token', token);
          }
        }
        
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
        console.log('SSOT P&L Data received:', result.data);
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

  // Enhanced export handlers for Balance Sheet
  const handleEnhancedCSVExport = async (balanceSheetData: SSOTBalanceSheetData) => {
    try {
      exportAndDownloadCSV(balanceSheetData, {
        includeAccountDetails: true,
        companyName: balanceSheetData.company?.name || 'PT. Sistem Akuntansi',
        filename: `balance_sheet_${ssotAsOfDate}.csv`
      });
      
      toast({
        title: 'CSV Export Successful',
        description: 'Enhanced Balance Sheet has been downloaded as CSV',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      toast({
        title: 'CSV Export Failed',
        description: error.message || 'Failed to export CSV',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  const handleEnhancedPDFExport = async (balanceSheetData: SSOTBalanceSheetData) => {
    try {
      exportAndDownloadPDF(balanceSheetData, {
        companyName: balanceSheetData.company?.name || 'PT. Sistem Akuntansi',
        includeAccountDetails: true,
        filename: `balance_sheet_${ssotAsOfDate}.pdf`
      });
      
      toast({
        title: 'PDF Export Successful',
        description: 'Enhanced Balance Sheet has been downloaded as PDF',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      toast({
        title: 'PDF Export Failed',
        description: error.message || 'Failed to export PDF',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  // Cash Flow Export Handlers
  const handleCashFlowCSVExport = async () => {
    if (!ssotCFData || !ssotCFStartDate || !ssotCFEndDate) {
      toast({
        title: 'Export Failed',
        description: 'No Cash Flow data available or missing date parameters',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    setLoading(true);
    try {
      await cashFlowExportService.exportToCSV({
        start_date: ssotCFStartDate,
        end_date: ssotCFEndDate
      });
      
      toast({
        title: 'CSV Export Successful',
        description: 'Cash Flow Statement has been downloaded as CSV',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      toast({
        title: 'CSV Export Failed',
        description: error.message || 'Failed to export CSV',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleCashFlowPDFExport = async () => {
    if (!ssotCFData || !ssotCFStartDate || !ssotCFEndDate) {
      toast({
        title: 'Export Failed',
        description: 'No Cash Flow data available or missing date parameters',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    setLoading(true);
    try {
      await cashFlowExportService.exportToPDF({
        start_date: ssotCFStartDate,
        end_date: ssotCFEndDate
      });
      
      toast({
        title: 'PDF Export Successful',
        description: 'Cash Flow Statement has been downloaded as PDF',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      toast({
        title: 'PDF Export Failed',
        description: error.message || 'Failed to export PDF',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
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
      } else if (report.id === 'purchase-report') {
        setSSOTPROpen(true);
        await fetchSSOTPurchaseReport();
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

  // Quick download function for PDF and CSV
  const handleQuickDownload = async (report: any, format: 'pdf' | 'csv') => {
    setLoading(true);
    
    try {
      // Set default parameters based on report type
      let params: any = { format };
      
      if (report.id === 'balance-sheet' || report.id === 'trial-balance') {
        params.as_of_date = new Date().toISOString().split('T')[0];
      } else if (['profit-loss', 'cash-flow', 'sales-summary', 'purchase-report', 'general-ledger', 'journal-entry-analysis'].includes(report.id)) {
        const today = new Date();
        const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
        params.start_date = firstDayOfMonth.toISOString().split('T')[0];
        params.end_date = today.toISOString().split('T')[0];
        
        if (report.id === 'general-ledger') {
          params.account_id = 'all';
        }
        if (report.id === 'journal-entry-analysis') {
          params.status = 'POSTED';
          params.reference_type = 'ALL';
        }
      }
      
      console.log('Downloading report:', report.id, 'with params:', params);
      
      // Generate and download the report
      const result = await reportService.generateReport(report.id, params);
      
      console.log('Report result type:', typeof result, 'isBlob:', result instanceof Blob);
      console.log('Report result:', result);
      
      if (result instanceof Blob) {
        if (result.size === 0) {
          throw new Error('Empty file received from server');
        }
        
        const fileName = `${report.id}_${format}_${new Date().toISOString().split('T')[0]}.${format}`;
        await reportService.downloadReport(result, fileName);
        
        toast({
          title: 'Download Successful',
          description: `${report.name} has been downloaded as ${format.toUpperCase()}`,
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      } else if (typeof result === 'object' && result !== null) {
        // If it's JSON data, create a manual download
        if (format === 'csv') {
          // Convert JSON to CSV
          const csvContent = convertJSONToCSV(result, report.id);
          const blob = new Blob([csvContent], { type: 'text/csv' });
          const fileName = `${report.id}_${format}_${new Date().toISOString().split('T')[0]}.${format}`;
          await reportService.downloadReport(blob, fileName);
          
          toast({
            title: 'Download Successful',
            description: `${report.name} has been downloaded as CSV`,
            status: 'success',
            duration: 3000,
            isClosable: true,
          });
        } else {
          // For PDF requests, provide specific guidance
          if (format === 'pdf') {
            throw new Error(`PDF export for ${report.name} is not yet implemented. Please use the "View Report" button to access the SSOT version with export options.`);
          } else {
            throw new Error(`${format.toUpperCase()} export is not yet supported for ${report.name}. Please use the "View Report" button for available export options.`);
          }
        }
      } else {
        throw new Error(`Invalid response format from server: ${typeof result}`);
      }
      
    } catch (error) {
      console.error('Quick download failed:', error);
      
      let errorMessage = 'Failed to download report';
      if (error instanceof Error) {
        errorMessage = error.message;
      }
      
      // Provide user-friendly error messages
      if (errorMessage.includes('Unknown report type') || errorMessage.includes('endpoint not found')) {
        errorMessage = `${format.toUpperCase()} export is not yet supported for ${report.name}. Please use the "More..." option for available formats.`;
      }
      
      toast({
        title: 'Download Failed',
        description: errorMessage,
        status: 'error',
        duration: 8000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };
  
  // Helper function to convert JSON to CSV with professional formatting
  const convertJSONToCSV = (data: any, reportType: string): string => {
    try {
      if (!data) return 'No data available';
      
      console.log('Converting to CSV. Report type:', reportType, 'Data:', data);
      
      // Helper function to format currency
      const formatCurrencyForCSV = (amount: number | null | undefined): string => {
        if (amount === null || amount === undefined || isNaN(Number(amount))) {
          return '0';
        }
        return Number(amount).toLocaleString('id-ID');
      };
      
      // Helper function to escape CSV values
      const escapeCSV = (value: string): string => {
        if (value.includes(',') || value.includes('"') || value.includes('\n')) {
          return `"${value.replace(/"/g, '""')}"`;
        }
        return value;
      };
      
      let csvLines: string[] = [];
      
      // Handle Profit & Loss Statement
      if (reportType === 'profit-loss') {
        const companyName = data.company?.name || 'PT. Sistem Akuntansi';
        const period = data.period || `${data.start_date || ''} to ${data.end_date || ''}`;
        
        // Header
        csvLines.push(companyName);
        csvLines.push('PROFIT & LOSS STATEMENT');
        csvLines.push(`Period: ${period}`);
        csvLines.push(''); // Empty line
        
        // Column headers
        csvLines.push('Account,Amount');
        
        // Process sections
        if (data.sections && Array.isArray(data.sections)) {
          data.sections.forEach((section: any) => {
            csvLines.push(`${escapeCSV(section.name || 'Unknown Section')},`);
            
            if (section.items && Array.isArray(section.items)) {
              section.items.forEach((item: any) => {
                const accountName = item.account_code ? `${item.account_code} - ${item.name}` : item.name;
                const amount = item.is_percentage ? `${item.amount}%` : formatCurrencyForCSV(item.amount);
                csvLines.push(`  ${escapeCSV(accountName)},${amount}`);
              });
            }
            
            const totalAmount = formatCurrencyForCSV(section.total);
            csvLines.push(`Total ${escapeCSV(section.name)},${totalAmount}`);
            csvLines.push(''); // Empty line between sections
          });
        }
        
        // Financial metrics if available
        if (data.financialMetrics) {
          csvLines.push('FINANCIAL METRICS,');
          csvLines.push(`Gross Profit,${formatCurrencyForCSV(data.financialMetrics.grossProfit)}`);
          csvLines.push(`Gross Margin,${data.financialMetrics.grossProfitMargin || 0}%`);
          csvLines.push(`Operating Income,${formatCurrencyForCSV(data.financialMetrics.operatingIncome)}`);
          csvLines.push(`Operating Margin,${data.financialMetrics.operatingMargin || 0}%`);
          csvLines.push(`Net Income,${formatCurrencyForCSV(data.financialMetrics.netIncome)}`);
          csvLines.push(`Net Margin,${data.financialMetrics.netIncomeMargin || 0}%`);
        }
      }
      
      // Handle Balance Sheet
      else if (reportType === 'balance-sheet') {
        const companyName = data.company?.name || 'PT. Sistem Akuntansi';
        const asOfDate = data.as_of_date || new Date().toISOString().split('T')[0];
        
        // Header
        csvLines.push(companyName);
        csvLines.push('BALANCE SHEET');
        csvLines.push(`As of: ${asOfDate}`);
        csvLines.push(''); // Empty line
        
        csvLines.push('Account,Amount');
        
        // Summary totals (support both nested SSOT format and flat format)
        csvLines.push('SUMMARY,');
        const totalAssets = (data.assets && data.assets.total_assets !== undefined) ? data.assets.total_assets : (data.total_assets || 0);
        const totalLiabilities = (data.liabilities && data.liabilities.total_liabilities !== undefined) ? data.liabilities.total_liabilities : (data.total_liabilities || 0);
        const totalEquity = (data.equity && data.equity.total_equity !== undefined) ? data.equity.total_equity : (data.total_equity || 0);
        csvLines.push(`Total Assets,${formatCurrencyForCSV(totalAssets)}`);
        csvLines.push(`Total Liabilities,${formatCurrencyForCSV(totalLiabilities)}`);
        csvLines.push(`Total Equity,${formatCurrencyForCSV(totalEquity)}`);
        csvLines.push(`Balanced,${data.is_balanced ? 'Yes' : 'No'}`);
        
        // Detailed breakdown if available
        if (data.assets) {
          csvLines.push('');
          csvLines.push('ASSETS,');
          
          if (data.assets.current_assets?.items) {
            csvLines.push('Current Assets,');
            data.assets.current_assets.items.forEach((item: any) => {
              csvLines.push(`  ${item.account_code} - ${escapeCSV(item.account_name)},${formatCurrencyForCSV(item.amount)}`);
            });
          }
          
          if (data.assets.non_current_assets?.items) {
            csvLines.push('Non-Current Assets,');
            data.assets.non_current_assets.items.forEach((item: any) => {
              csvLines.push(`  ${item.account_code} - ${escapeCSV(item.account_name)},${formatCurrencyForCSV(item.amount)}`);
            });
          }
        }
        
        if (data.liabilities?.current_liabilities?.items) {
          csvLines.push('');
          csvLines.push('LIABILITIES,');
          data.liabilities.current_liabilities.items.forEach((item: any) => {
            csvLines.push(`  ${item.account_code} - ${escapeCSV(item.account_name)},${formatCurrencyForCSV(item.amount)}`);
          });
        }
        
        if (data.equity?.items) {
          csvLines.push('');
          csvLines.push('EQUITY,');
          data.equity.items.forEach((item: any) => {
            csvLines.push(`  ${item.account_code} - ${escapeCSV(item.account_name)},${formatCurrencyForCSV(item.amount)}`);
          });
        }
      }
      
      // Handle Trial Balance
      else if (reportType === 'trial-balance') {
        const companyName = data.company?.name || 'PT. Sistem Akuntansi';
        const reportDate = data.report_date || new Date().toISOString().split('T')[0];
        
        // Header
        csvLines.push(companyName);
        csvLines.push('TRIAL BALANCE');
        csvLines.push(`As of: ${reportDate}`);
        csvLines.push(''); // Empty line
        
        csvLines.push('Account Code,Account Name,Account Type,Debit Balance,Credit Balance');
        
        if (data.accounts && Array.isArray(data.accounts)) {
          data.accounts.forEach((account: any) => {
            csvLines.push([
              escapeCSV(account.account_code || ''),
              escapeCSV(account.account_name || account.name || ''),
              escapeCSV(account.account_type || ''),
              formatCurrencyForCSV(account.debit_balance),
              formatCurrencyForCSV(account.credit_balance)
            ].join(','));
          });
          
          csvLines.push(''); // Empty line
          csvLines.push('TOTALS,,,'+ formatCurrencyForCSV(data.total_debits) + ',' + formatCurrencyForCSV(data.total_credits));
          csvLines.push(`Balanced: ${data.is_balanced ? 'Yes' : 'No'}`);
        }
      }
      
      // Handle Journal Entry Analysis
      else if (reportType === 'journal-entry-analysis') {
        // Backend CSV meta might wrap actual data under data
        const payload = (data && data.data && data.export_ready) ? data.data : data;
        const companyName = payload.company?.name || 'PT. Sistem Akuntansi';
        const period = `${payload.start_date || ''} to ${payload.end_date || ''}`;
        
        csvLines.push(companyName);
        csvLines.push('JOURNAL ENTRY ANALYSIS');
        csvLines.push(`Period: ${period}`);
        csvLines.push('');
        
        // Summary
        csvLines.push('Metric,Value');
        const summaryRows: Array<[string, any]> = [
          ['Total Entries', payload.total_entries],
          ['Posted Entries', payload.posted_entries],
          ['Draft Entries', payload.draft_entries],
          ['Reversed Entries', payload.reversed_entries],
          ['Total Amount', payload.total_amount]
        ];
        summaryRows.forEach(([k,v]) => {
          const val = (k.includes('Amount')) ? formatCurrencyForCSV(v) : (v ?? 0);
          csvLines.push(`${escapeCSV(k)},${val}`);
        });
        
        // Entries by type
        if (Array.isArray(payload.entries_by_type)) {
          csvLines.push('');
          csvLines.push('Entries By Type,,,,');
          csvLines.push('Source Type,Count,Amount,Percentage');
          payload.entries_by_type.forEach((row: any) => {
            csvLines.push([
              escapeCSV(row.source_type || ''),
              row.count ?? 0,
              formatCurrencyForCSV(row.total_amount),
              `${row.percentage ?? 0}%`
            ].join(','));
          });
        }
      }
      
      // Handle other report types with basic structure
      else {
        csvLines.push('FINANCIAL REPORT');
        csvLines.push(`Report Type: ${reportType}`);
        csvLines.push('');
        csvLines.push('Property,Value');
        
        if (typeof data === 'object') {
          Object.entries(data).forEach(([key, value]) => {
            let displayValue = '';
            if (typeof value === 'object' && value !== null) {
              displayValue = JSON.stringify(value);
            } else {
              displayValue = String(value || '');
            }
            csvLines.push(`${escapeCSV(key)},${escapeCSV(displayValue)}`);
          });
        }
      }
      
      // Footer
      csvLines.push('');
      csvLines.push(`Generated on: ${new Date().toLocaleString('id-ID')}`);
      
      const csvContent = csvLines.join('\n');
      console.log('Professional CSV generated, lines:', csvLines.length);
      
      return csvContent;
      
    } catch (error) {
      console.error('Error converting to CSV:', error);
      return `Error converting data to CSV format: ${error instanceof Error ? error.message : 'Unknown error'}`;
    }
  };

  // Format currency for display
  const formatCurrency = (amount: number | null | undefined) => {
    if (amount === null || amount === undefined || isNaN(Number(amount))) {
      return 'Rp 0';
    }
    
    const numericAmount = Number(amount);
    if (!isFinite(numericAmount)) {
      return 'Rp 0';
    }
    
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0
    }).format(numericAmount);
  };

  // Get badge color for entry type
  const getEntryTypeBadgeColor = (entryType: string) => {
    switch (entryType?.toUpperCase()) {
      case 'SALE':
        return 'green';
      case 'PURCHASE':
        return 'orange';
      case 'PAYMENT':
        return 'blue';
      case 'CASH_BANK':
        return 'teal';
      case 'JOURNAL':
        return 'purple';
      default:
        return 'gray';
    }
  };

  const handleGenerateReport = (report: any) => {
    setSelectedReport(report);
    resetParams();
    
    // Set default parameters based on report type
    if (report.id === 'balance-sheet') {
      setReportParams({ as_of_date: new Date().toISOString().split('T')[0], format: 'pdf' });
    } else if (report.id === 'profit-loss' || report.id === 'cash-flow') {
      const today = new Date();
      const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
      setReportParams({
        start_date: firstDayOfMonth.toISOString().split('T')[0],
        end_date: today.toISOString().split('T')[0],
        format: 'pdf'
      });
    } else if (report.id === 'sales-summary' || report.id === 'purchase-summary' || report.id === 'purchase-report') {
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
      const today = new Date();
      const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
      setReportParams({
        account_id: 'all',
        start_date: firstDayOfMonth.toISOString().split('T')[0],
        end_date: today.toISOString().split('T')[0],
        format: 'pdf'
      });
    } else if (report.id === 'journal-entry-analysis') {
      const today = new Date();
      const firstDayOfMonth = new Date(today.getFullYear(), today.getMonth(), 1);
      setReportParams({
        start_date: firstDayOfMonth.toISOString().split('T')[0],
        end_date: today.toISOString().split('T')[0],
        status: 'POSTED',
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
      // Validate required parameters
      if (['profit-loss', 'cash-flow', 'sales-summary', 'purchase-summary', 'purchase-report', 'general-ledger', 'journal-entry-analysis'].includes(selectedReport.id)) {
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
      const result = await reportService.generateReport(selectedReport.id, reportParams);
      
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
        console.error('Unexpected result format:', typeof result, result);
        throw new Error('Invalid response format from server');
      }
      
      onClose();
    } catch (error) {
      console.error('Failed to generate report:', error);
      
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
        <SimpleGrid columns={[1, 2, 3]} spacing={6} position="relative">
          {availableReports.map((report) => (
            <Card
              key={report.id}
              bg={cardBg}
              border="1px"
              borderColor={borderColor}
              borderRadius="md"
              overflow="visible"
              _hover={{ shadow: 'md' }}
              transition="all 0.2s"
              position="relative"
            >
              <CardBody p={0} position="relative">
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
                    <HStack spacing={2} width="full" mt={2} align="flex-start">
                      <Button
                        colorScheme="blue"
                        size="md"
                        flex="1"
                        onClick={() => {
                          // Open SSOT modals for reports that have SSOT integration
                          if (report.id === 'sales-summary') {
                            handleGenerateReport(report); // Use parameters modal first
                          } else if (report.id === 'purchase-report') {
                            setSSOTPROpen(true);
                          } else if (report.id === 'trial-balance') {
                            setSSOTTBOpen(true);
                          } else if (report.id === 'general-ledger') {
                            setSSOTGLOpen(true);
                          } else if (report.id === 'journal-entry-analysis') {
                            setSSOTJAOpen(true);
                          } else if (report.id === 'balance-sheet') {
                            setSSOTBSOpen(true);
                          } else if (report.id === 'cash-flow') {
                            setSSOTCFOpen(true);
                          } else if (report.id === 'profit-loss') {
                            setSSOTPLOpen(true);
                          } else {
                            // Use legacy modal for other reports
                            handleGenerateReport(report);
                          }
                        }}
                        isLoading={loading}
                        leftIcon={<FiEye />}
                      >
                        View Report
                      </Button>
                      <VStack spacing={1} flex="1">
                        <Button
                          colorScheme="red"
                          variant="outline"
                          size="sm"
                          width="full"
                          isLoading={loading}
                          leftIcon={<FiFilePlus />}
                          onClick={() => handleQuickDownload(report, 'pdf')}
                        >
                          PDF
                        </Button>
                        <Button
                          colorScheme="green"
                          variant="outline"
                          size="sm"
                          width="full"
                          isLoading={loading}
                          leftIcon={<FiFileText />}
                          onClick={() => handleQuickDownload(report, 'csv')}
                        >
                          CSV
                        </Button>
                        <Button
                          colorScheme="gray"
                          variant="outline"
                          size="sm"
                          width="full"
                          isLoading={loading}
                          leftIcon={<FiDownload />}
                          onClick={() => handleGenerateReport(report)}
                        >
                          More...
                        </Button>
                      </VStack>
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
                
                {/* Sales Summary Parameters */}
                {selectedReport.id === 'sales-summary' && (
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
                
                {/* Purchase Summary and Purchase Report Parameters */}
                {(selectedReport.id === 'purchase-summary' || selectedReport.id === 'purchase-report') && (
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
                        <option value="MANUAL">Manual</option>
                        <option value="SALES">Sales</option>
                        <option value="PURCHASE">Purchase</option>
                        <option value="PAYMENT">Payment</option>
                        <option value="RECEIPT">Receipt</option>
                      </Select>
                    </FormControl>
                  </>
                )}
                
                {/* Format Selection - Common for all reports */}
                <FormControl>
                  <FormLabel>Output Format</FormLabel>
                  <Select 
                    name="format" 
                    value={reportParams.format || 'pdf'} 
                    onChange={handleInputChange}
                  >
                    {/* For SSOT P&L, only allow PDF and CSV as per requirements */}
                    {selectedReport?.id === 'profit-loss' ? (
                      <>
                        <option value="pdf">PDF</option>
                        <option value="csv">CSV</option>
                      </>
                    ) : (
                      <>
                        <option value="pdf">PDF</option>
                        <option value="csv">CSV</option>
                        <option value="excel">Excel</option>
                        <option value="json">JSON</option>
                      </>
                    )}
                  </Select>
                </FormControl>
              </VStack>
            )}
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onClose}>
              Cancel
            </Button>
            <Button 
              colorScheme="blue" 
              onClick={() => {
                if (selectedReport?.id === 'sales-summary') {
                  // For sales summary, open SSOT modal with date parameters
                  setSSOTSSStartDate(reportParams.start_date || ssotSSStartDate);
                  setSSOTSSEndDate(reportParams.end_date || ssotSSEndDate);
                  setSSOTSSOpen(true);
                  // Auto-fetch the report
                  if (reportParams.start_date && reportParams.end_date) {
                    fetchSSOTSalesSummaryReport();
                  }
                  onClose();
                } else {
                  executeReport();
                }
              }}
              isLoading={loading}
              leftIcon={selectedReport?.id === 'sales-summary' ? <FiEye /> : <FiDownload />}
            >
              {selectedReport?.id === 'sales-summary' ? 'View Report' : 'Generate Report'}
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
                    <Text fontSize="sm" color={descriptionColor}>
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
                <Box textAlign="center" bg={summaryBg} p={4} borderRadius="md">
                  <Heading size="md" color={headingColor}>
                    {ssotPLData.company?.name || 'PT. Sistem Akuntansi'}
                  </Heading>
                  <Text fontSize="lg" fontWeight="semibold" mt={1}>
                    {ssotPLData.title || 'Enhanced Profit and Loss Statement'}
                  </Text>
                  <Text fontSize="sm" color={descriptionColor}>
                    Period: {ssotPLData.period || `${ssotStartDate} - ${ssotEndDate}`}
                  </Text>
                </Box>

                {ssotPLData.message && (
                  <Box bg="blue.50" p={4} borderRadius="md" border="1px" borderColor="blue.200">
                    <Text fontSize="sm" color="blue.800">
                      <strong>Analysis:</strong> {ssotPLData.message}
                    </Text>
                  </Box>
                )}

                {ssotPLData.hasData && ssotPLData.sections && (
                  <VStack spacing={4} align="stretch">
                    {/* Display sections data */}
                    {ssotPLData.sections.map((section: any, index: number) => (
                      <Box key={index} bg={cardBg} p={4} borderRadius="md" border="1px" borderColor={borderColor}>
                        <VStack spacing={3} align="stretch">
                          <HStack justify="space-between">
                            <Text fontSize="md" fontWeight="bold" color={headingColor}>
                              {section.name}
                            </Text>
                            <Text fontSize="md" fontWeight="bold" color={section.total >= 0 ? 'green.600' : 'red.600'}>
                              {formatCurrency(section.total || 0)}
                            </Text>
                          </HStack>
                          
                          {section.items && section.items.length > 0 && (
                            <VStack spacing={1} align="stretch">
                              {section.items.map((item: any, itemIndex: number) => (
                                <HStack key={itemIndex} justify="space-between" pl={4}>
                                  <Text fontSize="sm" color={descriptionColor}>
                                    {item.account_code ? `${item.account_code} - ${item.name}` : item.name}
                                    {item.is_percentage && ' (%)'}
                                  </Text>
                                  <Text fontSize="sm" color={textColor}>
                                    {item.is_percentage ? 
                                      `${item.amount}%` : 
                                      formatCurrency(item.amount || 0)
                                    }
                                  </Text>
                                </HStack>
                              ))}
                            </VStack>
                          )}
                        </VStack>
                      </Box>
                    ))}
                    
                    {/* Financial Metrics Summary */}
                    {ssotPLData.financialMetrics && (
                      <Box bg="green.50" p={4} borderRadius="md" border="1px" borderColor="green.200">
                        <Text fontSize="md" fontWeight="bold" color="green.800" mb={3}>
                          Key Financial Metrics
                        </Text>
                        <SimpleGrid columns={[1, 2]} spacing={3}>
                          <VStack spacing={2}>
                            <HStack justify="space-between" w="full">
                              <Text fontSize="sm" color="green.700">Gross Profit:</Text>
                              <Text fontSize="sm" fontWeight="semibold">
                                {formatCurrency(ssotPLData.financialMetrics.grossProfit || 0)}
                              </Text>
                            </HStack>
                            <HStack justify="space-between" w="full">
                              <Text fontSize="sm" color="green.700">Gross Margin:</Text>
                              <Text fontSize="sm" fontWeight="semibold">
                                {ssotPLData.financialMetrics.grossProfitMargin || 0}%
                              </Text>
                            </HStack>
                            <HStack justify="space-between" w="full">
                              <Text fontSize="sm" color="green.700">Operating Income:</Text>
                              <Text fontSize="sm" fontWeight="semibold">
                                {formatCurrency(ssotPLData.financialMetrics.operatingIncome || 0)}
                              </Text>
                            </HStack>
                          </VStack>
                          <VStack spacing={2}>
                            <HStack justify="space-between" w="full">
                              <Text fontSize="sm" color="green.700">Operating Margin:</Text>
                              <Text fontSize="sm" fontWeight="semibold">
                                {ssotPLData.financialMetrics.operatingMargin || 0}%
                              </Text>
                            </HStack>
                            <HStack justify="space-between" w="full">
                              <Text fontSize="sm" color="green.700">Net Income:</Text>
                              <Text fontSize="sm" fontWeight="semibold">
                                {formatCurrency(ssotPLData.financialMetrics.netIncome || 0)}
                              </Text>
                            </HStack>
                            <HStack justify="space-between" w="full">
                              <Text fontSize="sm" color="green.700">Net Margin:</Text>
                              <Text fontSize="sm" fontWeight="semibold">
                                {ssotPLData.financialMetrics.netIncomeMargin || 0}%
                              </Text>
                            </HStack>
                          </VStack>
                        </SimpleGrid>
                      </Box>
                    )}
                  </VStack>
                )}

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
            <HStack spacing={3}>
              {ssotPLData && !ssotPLLoading && (
                <>
                  <Button
                    colorScheme="red"
                    variant="outline"
                    size="sm"
                    leftIcon={<FiFilePlus />}
                    onClick={() => handleQuickDownload({id: 'profit-loss', name: 'Profit & Loss Statement'}, 'pdf')}
                  >
                    Export PDF
                  </Button>
                  <Button
                    colorScheme="green"
                    variant="outline"
                    size="sm"
                    leftIcon={<FiFileText />}
                    onClick={() => handleQuickDownload({id: 'profit-loss', name: 'Profit & Loss Statement'}, 'csv')}
                  >
                    Export CSV
                  </Button>
                </>
              )}
            </HStack>
            <Button variant="ghost" onClick={() => setSSOTPLOpen(false)}>
              Close
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
                    <Text fontSize="sm" color={descriptionColor}>
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
                <Box textAlign="center" bg={summaryBg} p={4} borderRadius="md">
                  <Heading size="md" color={headingColor}>
                    {ssotBSData.company?.name || 'PT. Sistem Akuntansi'}
                  </Heading>
                  <Text fontSize="lg" fontWeight="semibold" mt={1}>
                    Enhanced Balance Sheet (SSOT)
                  </Text>
                  <Text fontSize="sm" color={descriptionColor}>
                    As of: {ssotBSData.as_of_date ? new Date(ssotBSData.as_of_date).toLocaleDateString('id-ID') : ssotAsOfDate}
                  </Text>
                  {ssotBSData.is_balanced ? (
                    <Badge colorScheme="green" mt={2}>Balanced ✓</Badge>
                  ) : (
                    <Badge colorScheme="red" mt={2}>Not Balanced (Diff: {formatCurrency(ssotBSData.balance_difference || 0)})</Badge>
                  )}
                </Box>

                <SimpleGrid columns={[1, 2, 3]} spacing={4}>
                  <Box bg="green.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="sm" color="green.600">Total Assets</Text>
                    <Text fontSize="xl" fontWeight="bold" color="green.700">
                      {formatCurrency(ssotBSData.assets?.total_assets || ssotBSData.total_assets || 0)}
                    </Text>
                  </Box>
                  
                  <Box bg="orange.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="sm" color="orange.600">Total Liabilities</Text>
                    <Text fontSize="xl" fontWeight="bold" color="orange.700">
                      {formatCurrency(ssotBSData.liabilities?.total_liabilities || ssotBSData.total_liabilities || 0)}
                    </Text>
                  </Box>
                  
                  <Box bg="blue.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="sm" color="blue.600">Total Equity</Text>
                    <Text fontSize="xl" fontWeight="bold" color="blue.700">
                      {formatCurrency(ssotBSData.equity?.total_equity || ssotBSData.total_equity || 0)}
                    </Text>
                  </Box>
                </SimpleGrid>
                
                {/* Display detailed sections if available */}
                {(ssotBSData.assets?.current_assets || ssotBSData.assets?.non_current_assets) && (
                  <VStack spacing={4} align="stretch">
                    <Text fontSize="lg" fontWeight="bold" color={headingColor}>
                      Detailed Breakdown
                    </Text>
                    
                    {/* Assets Section */}
                    <Box bg={cardBg} p={4} borderRadius="md" border="1px" borderColor={borderColor}>
                      <VStack spacing={3} align="stretch">
                        <HStack justify="space-between">
                          <Text fontSize="md" fontWeight="bold" color={headingColor}>ASSETS</Text>
                          <Text fontSize="md" fontWeight="bold" color="green.600">
                            {formatCurrency(ssotBSData.assets?.total_assets || 0)}
                          </Text>
                        </HStack>
                        
                        {ssotBSData.assets?.current_assets?.items && (
                          <VStack spacing={1} align="stretch" pl={4}>
                            <Text fontSize="sm" fontWeight="semibold" color={descriptionColor}>Current Assets</Text>
                            {ssotBSData.assets.current_assets.items.map((item: any, index: number) => (
                              <HStack key={index} justify="space-between" pl={4}>
                                <Text fontSize="sm" color={descriptionColor}>
                                  {item.account_code} - {item.account_name}
                                </Text>
                                <Text fontSize="sm" color={textColor}>
                                  {formatCurrency(item.amount || 0)}
                                </Text>
                              </HStack>
                            ))}
                          </VStack>
                        )}
                        
                        {ssotBSData.assets?.non_current_assets?.items && (
                          <VStack spacing={1} align="stretch" pl={4}>
                            <Text fontSize="sm" fontWeight="semibold" color={descriptionColor}>Non-Current Assets</Text>
                            {ssotBSData.assets.non_current_assets.items.map((item: any, index: number) => (
                              <HStack key={index} justify="space-between" pl={4}>
                                <Text fontSize="sm" color={descriptionColor}>
                                  {item.account_code} - {item.account_name}
                                </Text>
                                <Text fontSize="sm" color={textColor}>
                                  {formatCurrency(item.amount || 0)}
                                </Text>
                              </HStack>
                            ))}
                          </VStack>
                        )}
                      </VStack>
                    </Box>
                    
                    {/* Liabilities Section */}
                    {ssotBSData.liabilities && (
                      <Box bg={cardBg} p={4} borderRadius="md" border="1px" borderColor={borderColor}>
                        <VStack spacing={3} align="stretch">
                          <HStack justify="space-between">
                            <Text fontSize="md" fontWeight="bold" color={headingColor}>LIABILITIES</Text>
                            <Text fontSize="md" fontWeight="bold" color="orange.600">
                              {formatCurrency(ssotBSData.liabilities?.total_liabilities || 0)}
                            </Text>
                          </HStack>
                          
                          {ssotBSData.liabilities?.current_liabilities?.items && (
                            <VStack spacing={1} align="stretch" pl={4}>
                              <Text fontSize="sm" fontWeight="semibold" color={descriptionColor}>Current Liabilities</Text>
                              {ssotBSData.liabilities.current_liabilities.items.map((item: any, index: number) => (
                                <HStack key={index} justify="space-between" pl={4}>
                                  <Text fontSize="sm" color={descriptionColor}>
                                    {item.account_code} - {item.account_name}
                                  </Text>
                                  <Text fontSize="sm" color={textColor}>
                                    {formatCurrency(item.amount || 0)}
                                  </Text>
                                </HStack>
                              ))}
                            </VStack>
                          )}
                        </VStack>
                      </Box>
                    )}
                    
                    {/* Equity Section */}
                    {ssotBSData.equity?.items && (
                      <Box bg={cardBg} p={4} borderRadius="md" border="1px" borderColor={borderColor}>
                        <VStack spacing={3} align="stretch">
                          <HStack justify="space-between">
                            <Text fontSize="md" fontWeight="bold" color={headingColor}>EQUITY</Text>
                            <Text fontSize="md" fontWeight="bold" color="blue.600">
                              {formatCurrency(ssotBSData.equity?.total_equity || 0)}
                            </Text>
                          </HStack>
                          
                          <VStack spacing={1} align="stretch" pl={4}>
                            {ssotBSData.equity.items.map((item: any, index: number) => (
                              <HStack key={index} justify="space-between" pl={4}>
                                <Text fontSize="sm" color={descriptionColor}>
                                  {item.account_code} - {item.account_name}
                                </Text>
                                <Text fontSize="sm" color={textColor}>
                                  {formatCurrency(item.amount || 0)}
                                </Text>
                              </HStack>
                            ))}
                          </VStack>
                        </VStack>
                      </Box>
                    )}
                  </VStack>
                )}
              </VStack>
            )}
          </ModalBody>
          <ModalFooter>
            <HStack spacing={3}>
              {ssotBSData && !ssotBSLoading && (
                <>
                  <Button
                    colorScheme="red"
                    variant="outline"
                    size="sm"
                    leftIcon={<FiFilePlus />}
                    onClick={() => handleEnhancedPDFExport(ssotBSData)}
                  >
                    Export PDF
                  </Button>
                  <Button
                    colorScheme="green"
                    variant="outline"
                    size="sm"
                    leftIcon={<FiFileText />}
                    onClick={() => handleEnhancedCSVExport(ssotBSData)}
                  >
                    Export CSV
                  </Button>
                </>
              )}
            </HStack>
            <Button variant="ghost" onClick={() => setSSOTBSOpen(false)}>
              Close
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* SSOT Cash Flow Modal */}
      <Modal isOpen={ssotCFOpen} onClose={() => setSSOTCFOpen(false)} size="6xl">
        <ModalOverlay />
        <ModalContent bg={modalContentBg}>
          <ModalHeader>
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
            <Box mb={4}>
              <HStack spacing={4} mb={4}>
                <FormControl>
                  <FormLabel>Start Date</FormLabel>
                  <Input 
                    type="date" 
                    value={ssotCFStartDate} 
                    onChange={(e) => setSSOTCFStartDate(e.target.value)} 
                  />
                </FormControl>
                <FormControl>
                  <FormLabel>End Date</FormLabel>
                  <Input 
                    type="date" 
                    value={ssotCFEndDate} 
                    onChange={(e) => setSSOTCFEndDate(e.target.value)} 
                  />
                </FormControl>
                <Button
                  colorScheme="blue"
                  onClick={fetchSSOTCashFlowReport}
                  isLoading={ssotCFLoading}
                  leftIcon={<FiActivity />}
                  size="md"
                  mt={8}
                >
                  Generate Report
                </Button>
              </HStack>
            </Box>

            {ssotCFLoading && (
              <Box textAlign="center" py={8}>
                <VStack spacing={4}>
                  <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
                  <VStack spacing={2}>
                    <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                      Generating SSOT Cash Flow Statement
                    </Text>
                    <Text fontSize="sm" color={descriptionColor}>
                      Analyzing journal entries for cash flow activities...
                    </Text>
                  </VStack>
                </VStack>
              </Box>
            )}

            {ssotCFError && (
              <Box bg="red.50" p={4} borderRadius="md" mb={4}>
                <Text color="red.600" fontWeight="medium">Error: {ssotCFError}</Text>
                <Button
                  mt={2}
                  size="sm"
                  colorScheme="red"
                  variant="outline"
                  onClick={fetchSSOTCashFlowReport}
                >
                  Retry
                </Button>
              </Box>
            )}

            {ssotCFData && !ssotCFLoading && (
              <VStack spacing={6} align="stretch">
                <Box bg={summaryBg} p={4} borderRadius="md" borderWidth="2px" borderColor={borderColor}>
                  <HStack justify="space-between">
                    <VStack align="start" spacing={1}>
                      <Text fontSize="lg" fontWeight="bold" color={headingColor}>
                        Cash Position Summary
                      </Text>
                      <Text fontSize="sm" color={descriptionColor}>
                        {ssotCFData.start_date} to {ssotCFData.end_date}
                      </Text>
                    </VStack>
                    <VStack align="end" spacing={1}>
                      <Text fontSize="sm" color={descriptionColor}>Net Cash Flow</Text>
                      <Text fontSize="xl" fontWeight="bold" color={ssotCFData.net_cash_flow >= 0 ? 'green.600' : 'red.600'}>
                        {formatCurrency(ssotCFData.net_cash_flow)}
                      </Text>
                    </VStack>
                  </HStack>
                </Box>
              </VStack>
            )}
          </ModalBody>
          <ModalFooter>
            <HStack spacing={3}>
              {ssotCFData && !ssotCFLoading && (
                <>
                  <Button
                    colorScheme="red"
                    variant="outline"
                    size="sm"
                    leftIcon={<FiFilePlus />}
                    onClick={() => handleCashFlowPDFExport()}
                    isLoading={loading}
                  >
                    Export PDF
                  </Button>
                  <Button
                    colorScheme="green"
                    variant="outline"
                    size="sm"
                    leftIcon={<FiFileText />}
                    onClick={() => handleCashFlowCSVExport()}
                    isLoading={loading}
                  >
                    Export CSV
                  </Button>
                </>
              )}
            </HStack>
            <Button variant="ghost" onClick={() => setSSOTCFOpen(false)}>
              Close
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* SSOT Purchase Report Modal */}
      <Modal isOpen={ssotPROpen} onClose={() => setSSOTPROpen(false)} size="6xl">
        <ModalOverlay />
        <ModalContent bg={modalContentBg}>
          <ModalHeader>
            <HStack>
              <Icon as={FiShoppingCart} color="blue.500" />
              <VStack align="start" spacing={0}>
                <Text fontSize="lg" fontWeight="bold">
                  Purchase Report (SSOT)
                </Text>
                <Text fontSize="sm" color={previewPeriodTextColor}>
                  {ssotPRStartDate} - {ssotPREndDate} | Credible Purchase Analysis
                </Text>
              </VStack>
            </HStack>
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            <Box mb={4}>
              <HStack spacing={4} mb={4}>
                <FormControl>
                  <FormLabel>Start Date</FormLabel>
                  <Input 
                    type="date" 
                    value={ssotPRStartDate} 
                    onChange={(e) => setSSOTPRStartDate(e.target.value)} 
                  />
                </FormControl>
                <FormControl>
                  <FormLabel>End Date</FormLabel>
                  <Input 
                    type="date" 
                    value={ssotPREndDate} 
                    onChange={(e) => setSSOTPREndDate(e.target.value)} 
                  />
                </FormControl>
                <Button
                  colorScheme="blue"
                  onClick={fetchSSOTPurchaseReport}
                  isLoading={ssotPRLoading}
                  leftIcon={<FiShoppingCart />}
                  size="md"
                  mt={8}
                >
                  Generate Report
                </Button>
              </HStack>
            </Box>

            {ssotPRLoading && (
              <Box textAlign="center" py={8}>
                <VStack spacing={4}>
                  <Spinner size="xl" thickness="4px" speed="0.65s" color="blue.500" />
                  <VStack spacing={2}>
                    <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                      Generating Purchase Report
                    </Text>
                    <Text fontSize="sm" color={descriptionColor}>
                      Analyzing purchase transactions with credible data...
                    </Text>
                  </VStack>
                </VStack>
              </Box>
            )}

            {ssotPRError && (
              <Box bg="red.50" p={4} borderRadius="md" mb={4}>
                <Text color="red.600" fontWeight="medium">Error: {ssotPRError}</Text>
                <Button
                  mt={2}
                  size="sm"
                  colorScheme="red"
                  variant="outline"
                  onClick={fetchSSOTPurchaseReport}
                >
                  Retry
                </Button>
              </Box>
            )}

            {ssotPRData && !ssotPRLoading && (
              <VStack spacing={6} align="stretch">
                {/* Company Header */}
                <Box bg="blue.50" p={4} borderRadius="md">
                  <HStack justify="space-between" align="start">
                    <VStack align="start" spacing={1}>
                      <Text fontSize="lg" fontWeight="bold" color="blue.800">
                        {ssotPRData.company?.name || 'PT. Sistem Akuntansi'}
                      </Text>
                      <Text fontSize="sm" color="blue.600">
                        Purchase Analysis Report
                      </Text>
                    </VStack>
                    <VStack align="end" spacing={1}>
                      <Text fontSize="sm" color="blue.600">
                        Currency: {ssotPRData.currency || 'IDR'}
                      </Text>
                      <Text fontSize="xs" color="blue.500">
                        Generated: {ssotPRData.generated_at ? new Date(ssotPRData.generated_at).toLocaleString('id-ID') : 'N/A'}
                      </Text>
                    </VStack>
                  </HStack>
                </Box>

                {/* Financial Summary */}
                <SimpleGrid columns={[1, 2, 4]} spacing={4}>
                  <Box bg="blue.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="2xl" fontWeight="bold" color="blue.600">
                      {ssotPRData.total_purchases || 0}
                    </Text>
                    <Text fontSize="sm" color="blue.800">Total Purchases</Text>
                  </Box>
                  <Box bg="green.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="2xl" fontWeight="bold" color="green.600">
                      {formatCurrency(ssotPRData.total_amount || 0)}
                    </Text>
                    <Text fontSize="sm" color="green.800">Total Amount</Text>
                  </Box>
                  <Box bg="purple.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="2xl" fontWeight="bold" color="purple.600">
                      {formatCurrency(ssotPRData.total_paid || 0)}
                    </Text>
                    <Text fontSize="sm" color="purple.800">Total Paid</Text>
                  </Box>
                  <Box bg="orange.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="2xl" fontWeight="bold" color="orange.600">
                      {formatCurrency(ssotPRData.outstanding_payables || 0)}
                    </Text>
                    <Text fontSize="sm" color="orange.800">Outstanding</Text>
                  </Box>
                </SimpleGrid>

                {/* Payment Analysis */}
                {ssotPRData.payment_analysis && (
                  <Box bg={cardBg} p={4} borderRadius="md" border="1px" borderColor={borderColor}>
                    <Text fontSize="md" fontWeight="bold" color={headingColor} mb={3}>
                      Payment Method Analysis
                    </Text>
                    <SimpleGrid columns={[1, 2]} spacing={4}>
                      <VStack spacing={2}>
                        <Text fontSize="sm" color={descriptionColor}>Cash Purchases</Text>
                        <Text fontSize="xl" fontWeight="bold" color="green.600">
                          {ssotPRData.payment_analysis.cash_purchases || 0} ({(ssotPRData.payment_analysis.cash_percentage || 0).toFixed(1)}%)
                        </Text>
                        <Text fontSize="sm" color="green.600">
                          {formatCurrency(ssotPRData.payment_analysis.cash_amount || 0)}
                        </Text>
                      </VStack>
                      <VStack spacing={2}>
                        <Text fontSize="sm" color={descriptionColor}>Credit Purchases</Text>
                        <Text fontSize="xl" fontWeight="bold" color="orange.600">
                          {ssotPRData.payment_analysis.credit_purchases || 0} ({(ssotPRData.payment_analysis.credit_percentage || 0).toFixed(1)}%)
                        </Text>
                        <Text fontSize="sm" color="orange.600">
                          {formatCurrency(ssotPRData.payment_analysis.credit_amount || 0)}
                        </Text>
                      </VStack>
                    </SimpleGrid>
                  </Box>
                )}

                {/* Vendor Analysis */}
                {ssotPRData.purchases_by_vendor && ssotPRData.purchases_by_vendor.length > 0 && (
                  <Box>
                    <Text fontSize="md" fontWeight="bold" color={headingColor} mb={4}>
                      Purchases by Vendor ({ssotPRData.purchases_by_vendor.length} vendors)
                    </Text>
                    <VStack spacing={3} align="stretch" maxH="400px" overflow="auto">
                      {ssotPRData.purchases_by_vendor.map((vendor: any, index: number) => (
                        <Box key={index} border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="white" _hover={{ bg: 'gray.50' }}>
                          <SimpleGrid columns={[1, 2, 4]} spacing={2}>
                            <VStack align="start" spacing={1}>
                              <Text fontWeight="bold" fontSize="md" color="gray.800">
                                {vendor.vendor_name || 'Unknown Vendor'}
                              </Text>
                              <Text fontSize="xs" color="gray.600">
                                ID: {vendor.vendor_id || 'N/A'}
                              </Text>
                              <Badge colorScheme={vendor.payment_method === 'CASH' ? 'green' : 'orange'} size="sm">
                                {vendor.payment_method || 'N/A'}
                              </Badge>
                            </VStack>
                            <VStack align="start" spacing={0}>
                              <Text fontSize="sm" color="gray.700">
                                Total: {formatCurrency(vendor.total_amount || 0)}
                              </Text>
                              <Text fontSize="sm" color="green.600">
                                Paid: {formatCurrency(vendor.total_paid || 0)}
                              </Text>
                              <Text fontSize="sm" color="orange.600">
                                Outstanding: {formatCurrency(vendor.outstanding || 0)}
                              </Text>
                            </VStack>
                            <VStack align="start" spacing={0}>
                              <Text fontSize="sm" color="gray.700">
                                Purchases: {vendor.total_purchases || 0}
                              </Text>
                              <Text fontSize="xs" color="gray.600">
                                Status: {vendor.status || 'N/A'}
                              </Text>
                            </VStack>
                            <VStack align="end" spacing={0}>
                              <Text fontSize="xs" color="gray.600">
                                Last Purchase:
                              </Text>
                              <Text fontSize="xs" color="gray.600">
                                {vendor.last_purchase_date ? new Date(vendor.last_purchase_date).toLocaleDateString('id-ID') : 'N/A'}
                              </Text>
                            </VStack>
                          </SimpleGrid>
                        </Box>
                      ))}
                    </VStack>
                  </Box>
                )}

                {/* Tax Analysis */}
                {ssotPRData.tax_analysis && (
                  <Box bg={cardBg} p={4} borderRadius="md" border="1px" borderColor={borderColor}>
                    <Text fontSize="md" fontWeight="bold" color={headingColor} mb={3}>
                      Tax Analysis
                    </Text>
                    <SimpleGrid columns={[1, 3]} spacing={4}>
                      <Box textAlign="center">
                        <Text fontSize="sm" color={descriptionColor}>Taxable Amount</Text>
                        <Text fontSize="lg" fontWeight="bold" color="blue.600">
                          {formatCurrency(ssotPRData.tax_analysis.total_taxable_amount || 0)}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color={descriptionColor}>Tax Amount</Text>
                        <Text fontSize="lg" fontWeight="bold" color="purple.600">
                          {formatCurrency(ssotPRData.tax_analysis.total_tax_amount || 0)}
                        </Text>
                      </Box>
                      <Box textAlign="center">
                        <Text fontSize="sm" color={descriptionColor}>Average Tax Rate</Text>
                        <Text fontSize="lg" fontWeight="bold" color="orange.600">
                          {(ssotPRData.tax_analysis.average_tax_rate || 0).toFixed(2)}%
                        </Text>
                      </Box>
                    </SimpleGrid>
                  </Box>
                )}
              </VStack>
            )}
          </ModalBody>
          <ModalFooter>
            <HStack spacing={3}>
              {ssotPRData && !ssotPRLoading && (
                <>
                  <Button
                    colorScheme="red"
                    variant="outline"
                    size="sm"
                    leftIcon={<FiFilePlus />}
                    onClick={() => handleQuickDownload({id: 'purchase-report', name: 'Purchase Report'}, 'pdf')}
                  >
                    Export PDF
                  </Button>
                  <Button
                    colorScheme="green"
                    variant="outline"
                    size="sm"
                    leftIcon={<FiFileText />}
                    onClick={() => handleQuickDownload({id: 'purchase-report', name: 'Purchase Report'}, 'csv')}
                  >
                    Export CSV
                  </Button>
                </>
              )}
            </HStack>
            <Button variant="ghost" onClick={() => setSSOTPROpen(false)}>
              Close
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* SSOT Sales Summary Modal - Using enhanced component */}
      <SalesSummaryModal
        isOpen={ssotSSOpen}
        onClose={() => setSSOTSSOpen(false)}
        data={ssotSSData}
        isLoading={ssotSSLoading}
        error={ssotSSError}
        startDate={ssotSSStartDate}
        endDate={ssotSSEndDate}
        onDateChange={(newStartDate, newEndDate) => {
          setSSOTSSStartDate(newStartDate);
          setSSOTSSEndDate(newEndDate);
        }}
        onFetch={fetchSSOTSalesSummaryReport}
        onExport={async (format) => {
          try {
            toast({
              title: 'Export ' + format.toUpperCase(),
              description: `Exporting Sales Summary Report as ${format.toUpperCase()}...`,
              status: 'info',
              duration: 2000,
              isClosable: true,
            });
            
            if (format === 'pdf' || format === 'excel') {
              // Try to use the report service for professional export
              try {
                const result = await reportService.generateReport('sales-summary', {
                  start_date: ssotSSStartDate,
                  end_date: ssotSSEndDate,
                  format: format === 'excel' ? 'csv' : 'pdf'
                });
                
                if (result instanceof Blob) {
                  const fileName = `sales-summary-${ssotSSStartDate}-to-${ssotSSEndDate}.${format === 'excel' ? 'csv' : 'pdf'}`;
                  await reportService.downloadReport(result, fileName);
                  
                  toast({
                    title: 'Export Successful',
                    description: `Sales Summary Report exported as ${format.toUpperCase()}`,
                    status: 'success',
                    duration: 3000,
                    isClosable: true,
                  });
                  return;
                }
              } catch (exportError) {
                console.warn('Professional export failed, falling back to JSON:', exportError);
              }
            }
            
            // Fallback: export as JSON/CSV
            if (ssotSSData) {
              let content: string;
              let mimeType: string;
              let extension: string;
              
              if (format === 'excel') {
                // Generate CSV content
                const customers = ssotSSData.sales_by_customer || [];
                const csvHeaders = 'Customer Name,Contact Person,Phone,Email,Total Sales,Order Count,Average Order Value\n';
                const csvRows = customers.map(customer => 
                  `"${customer.customer_name || 'Unnamed Customer'}",` +
                  `"${customer.contact_person || ''}",` +
                  `"${customer.phone || ''}",` +
                  `"${customer.email || ''}",` +
                  `${customer.total_sales || customer.sales_amount || 0},` +
                  `${customer.order_count || customer.orders || 0},` +
                  `${customer.average_order_value || (customer.total_sales / (customer.order_count || 1))}`
                ).join('\n');
                content = csvHeaders + csvRows;
                mimeType = 'text/csv';
                extension = 'csv';
              } else {
                // Generate JSON content
                const reportData = {
                  reportType: 'Sales Summary Report',
                  period: `${ssotSSStartDate} to ${ssotSSEndDate}`,
                  generatedOn: new Date().toISOString(),
                  totalRevenue: ssotSSData.total_revenue || ssotSSData.total_sales || 0,
                  totalCustomers: ssotSSData.total_customers || 0,
                  totalOrders: ssotSSData.total_orders || 0,
                  averageOrderValue: ssotSSData.average_order_value || 0,
                  customers: ssotSSData.sales_by_customer || [],
                  topCustomers: ssotSSData.top_customers || [],
                  salesTrends: ssotSSData.sales_trends || {},
                  company: ssotSSData.company || {}
                };
                content = JSON.stringify(reportData, null, 2);
                mimeType = 'application/json';
                extension = 'json';
              }
              
              const dataBlob = new Blob([content], { type: mimeType });
              const url = URL.createObjectURL(dataBlob);
              const link = document.createElement('a');
              link.href = url;
              link.download = `sales-summary-${ssotSSStartDate}-to-${ssotSSEndDate}.${extension}`;
              link.click();
              URL.revokeObjectURL(url);
              
              toast({
                title: 'Export Successful',
                description: `Sales Summary Report exported as ${extension.toUpperCase()}`,
                status: 'success',
                duration: 3000,
                isClosable: true,
              });
            }
          } catch (error) {
            console.error('Export failed:', error);
            toast({
              title: 'Export Failed',
              description: error instanceof Error ? error.message : 'Failed to export report',
              status: 'error',
              duration: 5000,
              isClosable: true,
            });
          }
        }}
      />

      {/* SSOT Purchase Report Modal */}
      <Modal isOpen={ssotPROpen} onClose={() => setSSOTPROpen(false)} size="6xl">
        <ModalOverlay />
        <ModalContent bg={modalContentBg}>
          <ModalHeader>
            <HStack>
              <Icon as={FiShoppingCart} color="orange.500" />
              <VStack align="start" spacing={0}>
                <Text fontSize="lg" fontWeight="bold">
                  Purchase Report (SSOT)
                </Text>
                <Text fontSize="sm" color={previewPeriodTextColor}>
                  {ssotPRStartDate} - {ssotPREndDate} | SSOT Journal Integration
                </Text>
              </VStack>
            </HStack>
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            <Box mb={4}>
              <HStack spacing={4} mb={4}>
                <FormControl>
                  <FormLabel>Start Date</FormLabel>
                  <Input 
                    type="date" 
                    value={ssotPRStartDate} 
                    onChange={(e) => setSSOTPRStartDate(e.target.value)} 
                  />
                </FormControl>
                <FormControl>
                  <FormLabel>End Date</FormLabel>
                  <Input 
                    type="date" 
                    value={ssotPREndDate} 
                    onChange={(e) => setSSOTPREndDate(e.target.value)} 
                  />
                </FormControl>
                <Button
                  colorScheme="blue"
                  onClick={fetchSSOTPurchaseReport}
                  isLoading={ssotPRLoading}
                  leftIcon={<FiShoppingCart />}
                  size="md"
                  mt={8}
                >
                  Generate Report
                </Button>
              </HStack>
            </Box>

            {ssotPRLoading && (
              <Box textAlign="center" py={8}>
                <VStack spacing={4}>
                  <Spinner size="xl" thickness="4px" speed="0.65s" color="orange.500" />
                  <VStack spacing={2}>
                    <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                      Generating Purchase Report
                    </Text>
                    <Text fontSize="sm" color={descriptionColor}>
                      Analyzing purchase transactions from SSOT journal system...
                    </Text>
                  </VStack>
                </VStack>
              </Box>
            )}

            {ssotPRError && (
              <Box bg="red.50" p={4} borderRadius="md" mb={4}>
                <Text color="red.600" fontWeight="medium">Error: {ssotPRError}</Text>
                <Button
                  mt={2}
                  size="sm"
                  colorScheme="red"
                  variant="outline"
                  onClick={fetchSSOTPurchaseReport}
                >
                  Retry
                </Button>
              </Box>
            )}

            {ssotPRData && !ssotPRLoading && (
              <VStack spacing={6} align="stretch">
                {/* Company Header */}
                {ssotPRData.company && (
                  <Box bg="orange.50" p={4} borderRadius="md">
                    <HStack justify="space-between" align="start">
                      <VStack align="start" spacing={1}>
                        <Text fontSize="lg" fontWeight="bold" color="orange.800">
                          {ssotPRData.company.name || 'Company Name Not Available'}
                        </Text>
                        <Text fontSize="sm" color="orange.600">
                          {ssotPRData.company.address && ssotPRData.company.city ? 
                            `${ssotPRData.company.address}, ${ssotPRData.company.city}` : 
                            'Address not available'
                          }
                        </Text>
                        {ssotPRData.company.phone && (
                          <Text fontSize="sm" color="orange.600">
                            {ssotPRData.company.phone} | {ssotPRData.company.email}
                          </Text>
                        )}
                      </VStack>
                      <VStack align="end" spacing={1}>
                        <Text fontSize="sm" color="orange.600">
                          Currency: {ssotPRData.currency || 'IDR'}
                        </Text>
                        <Text fontSize="xs" color="orange.500">
                          Generated: {ssotPRData.generated_at ? new Date(ssotPRData.generated_at).toLocaleString('id-ID') : 'N/A'}
                        </Text>
                      </VStack>
                    </HStack>
                  </Box>
                )}

                {/* Report Header */}
                <Box textAlign="center" bg={summaryBg} p={4} borderRadius="md">
                  <Heading size="md" color={headingColor}>
                    Purchase Report
                  </Heading>
                  <Text fontSize="sm" color={descriptionColor}>
                    Period: {ssotPRStartDate} - {ssotPREndDate}
                  </Text>
                  <Text fontSize="xs" color={descriptionColor} mt={1}>
                    Generated: {new Date().toLocaleDateString('id-ID')} at {new Date().toLocaleTimeString('id-ID')}
                  </Text>
                </Box>

                {/* Summary Statistics */}
                <SimpleGrid columns={[1, 2, 4]} spacing={4}>
                  <Box bg="orange.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="2xl" fontWeight="bold" color="orange.600">
                      {ssotPRData.total_vendors || 0}
                    </Text>
                    <Text fontSize="sm" color="orange.800">Total Vendors</Text>
                  </Box>
                  <Box bg="blue.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="2xl" fontWeight="bold" color="blue.600">
                      {ssotPRData.active_vendors || 0}
                    </Text>
                    <Text fontSize="sm" color="blue.800">Active Vendors</Text>
                  </Box>
                  <Box bg="red.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="2xl" fontWeight="bold" color="red.600">
                      {ssotPRData.total_purchases || 0}
                    </Text>
                    <Text fontSize="sm" color="red.800">Total Purchases</Text>
                  </Box>
                  <Box bg="green.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="2xl" fontWeight="bold" color="green.600">
                      {formatCurrency(ssotPRData.total_paid || 0)}
                    </Text>
                    <Text fontSize="sm" color="green.800">Total Payments</Text>
                  </Box>
                </SimpleGrid>

                {/* Outstanding Payables */}
                {ssotPRData.outstanding_payables !== undefined && (
                  <Box bg={ssotPRData.outstanding_payables < 0 ? 'green.50' : 'yellow.50'} p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="sm" color={ssotPRData.outstanding_payables < 0 ? 'green.600' : 'yellow.600'} mb={2}>
                      Outstanding Payables Status
                    </Text>
                    <Text fontSize="3xl" fontWeight="bold" color={ssotPRData.outstanding_payables < 0 ? 'green.700' : 'yellow.700'}>
                      {formatCurrency(Math.abs(ssotPRData.outstanding_payables))}
                    </Text>
                    <Text fontSize="sm" color={ssotPRData.outstanding_payables < 0 ? 'green.600' : 'yellow.600'} mt={1}>
                      {ssotPRData.outstanding_payables < 0 ? 'Overpaid (Credit Balance)' : 'Outstanding Amount'}
                    </Text>
                  </Box>
                )}

                {/* Vendors by Performance */}
                {ssotPRData.vendors_by_performance && ssotPRData.vendors_by_performance.length > 0 && (
                  <Box>
                    <Heading size="sm" mb={4} color={headingColor}>
                      Vendors Performance Analysis ({ssotPRData.vendors_by_performance.length} vendors)
                    </Heading>
                    
                    {/* Vendor Table Header */}
                    <Box bg="orange.50" p={3} borderRadius="md" mb={2} border="1px solid" borderColor="orange.200">
                      <SimpleGrid columns={[1, 2, 6]} spacing={2} fontSize="sm" fontWeight="bold" color="orange.800">
                        <Text>Vendor</Text>
                        <Text>Rating & Score</Text>
                        <Text textAlign="right">Purchases</Text>
                        <Text textAlign="right">Payments</Text>
                        <Text textAlign="right">Outstanding</Text>
                        <Text textAlign="right">Avg Pay Days</Text>
                      </SimpleGrid>
                    </Box>
                    
                    {/* Vendor Rows */}
                    <VStack spacing={2} align="stretch">
                      {ssotPRData.vendors_by_performance.map((vendor, index) => (
                        <Box key={index} border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="white" _hover={{ bg: 'gray.50' }}>
                          <SimpleGrid columns={[1, 2, 6]} spacing={2} fontSize="sm">
                            <VStack align="start" spacing={1}>
                              <Text fontWeight="bold" fontSize="md" color="gray.800">
                                {vendor.vendor_name || 'Unnamed Vendor'}
                              </Text>
                              {vendor.vendor_id && (
                                <Text fontSize="xs" color="gray.600">
                                  ID: {vendor.vendor_id}
                                </Text>
                              )}
                            </VStack>
                            <VStack align="start" spacing={1}>
                              <Badge colorScheme={vendor.rating === 'Good' ? 'green' : vendor.rating === 'Fair' ? 'yellow' : 'red'} size="sm">
                                {vendor.rating || 'No Rating'}
                              </Badge>
                              <Text fontSize="xs" color="blue.600">
                                Score: {vendor.payment_score || 0}/100
                              </Text>
                            </VStack>
                            <Text textAlign="right" fontSize="sm" fontWeight="bold" color="orange.600">
                              {formatCurrency(vendor.total_amount || 0)}
                            </Text>
                            <Text textAlign="right" fontSize="sm" fontWeight="bold" color="green.600">
                              {formatCurrency(vendor.total_paid || 0)}
                            </Text>
                            <VStack align="end" spacing={0}>
                              <Text textAlign="right" fontSize="sm" fontWeight="medium" color={vendor.outstanding > 0 ? 'red.600' : vendor.outstanding < 0 ? 'green.600' : 'gray.400'}>
                                {vendor.outstanding !== 0 ? formatCurrency(Math.abs(vendor.outstanding)) : '-'}
                              </Text>
                              {vendor.outstanding < 0 && (
                                <Text fontSize="xs" color="green.500">
                                  (Credit)
                                </Text>
                              )}
                            </VStack>
                            <Text textAlign="right" fontSize="sm" fontWeight="medium" color="blue.600">
                              {vendor.average_payment_days || 0} days
                            </Text>
                          </SimpleGrid>
                        </Box>
                      ))}
                    </VStack>
                    
                    {/* Totals Row */}
                    <Box bg="orange.100" p={3} borderRadius="md" mt={2} border="2px solid" borderColor="orange.300">
                      <SimpleGrid columns={[1, 2, 6]} spacing={2} fontSize="md" fontWeight="bold">
                        <Text color="orange.800">TOTALS:</Text>
                        <Text></Text>
                        <Text textAlign="right" color="orange.700">
                          {formatCurrency(ssotPRData.total_amount || 0)}
                        </Text>
                        <Text textAlign="right" color="green.700">
                          {formatCurrency(ssotPRData.total_paid || 0)}
                        </Text>
                        <Text textAlign="right" color={ssotPRData.outstanding_payables > 0 ? 'red.700' : 'green.700'}>
                          {formatCurrency(Math.abs(ssotPRData.outstanding_payables || 0))}
                        </Text>
                        <Text textAlign="right" color="blue.700">
                          {ssotPRData.vendors_by_performance ? 
                            Math.round(ssotPRData.vendors_by_performance.reduce((sum, v) => sum + (v.average_payment_days || 0), 0) / ssotPRData.vendors_by_performance.length) 
                            : 0} days
                        </Text>
                      </SimpleGrid>
                    </Box>
                  </Box>
                )}

                {/* Top Vendors by Spend */}
                {ssotPRData.top_vendors_by_spend && ssotPRData.top_vendors_by_spend.length > 0 && (
                  <Box>
                    <Heading size="sm" mb={4} color={headingColor}>
                      Top Vendors by Purchase Amount
                    </Heading>
                    <SimpleGrid columns={[1, 2, 3]} spacing={4}>
                      {ssotPRData.top_vendors_by_spend.map((vendor, index) => (
                        <Box key={index} border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="white">
                          <VStack spacing={2}>
                            <Badge colorScheme="orange" size="lg" variant="solid">
                              #{index + 1}
                            </Badge>
                            <Text fontWeight="bold" fontSize="md" color="gray.800" textAlign="center">
                              {vendor.vendor_name || vendor.name}
                            </Text>
                            <Text fontSize="lg" fontWeight="bold" color="orange.600">
                              {formatCurrency(vendor.total_amount || vendor.total_purchases)}
                            </Text>
                            {vendor.percentage && (
                              <Text fontSize="sm" color="gray.600">
                                {vendor.percentage.toFixed(1)}% of total
                              </Text>
                            )}
                          </VStack>
                        </Box>
                      ))}
                    </SimpleGrid>
                  </Box>
                )}

                {/* Payment Analysis */}
                {ssotPRData.payment_analysis && (
                  <Box>
                    <Heading size="sm" mb={4} color={headingColor}>
                      Payment Performance Analysis
                    </Heading>
                    <SimpleGrid columns={[1, 2, 4]} spacing={4}>
                      <Box border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="green.50">
                        <VStack spacing={2}>
                          <Text fontSize="2xl" fontWeight="bold" color="green.600">
                            {ssotPRData.payment_analysis.on_time_payments || 0}
                          </Text>
                          <Text fontSize="sm" color="green.800" textAlign="center">On-Time Payments</Text>
                        </VStack>
                      </Box>
                      <Box border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="yellow.50">
                        <VStack spacing={2}>
                          <Text fontSize="2xl" fontWeight="bold" color="yellow.600">
                            {ssotPRData.payment_analysis.late_payments || 0}
                          </Text>
                          <Text fontSize="sm" color="yellow.800" textAlign="center">Late Payments</Text>
                        </VStack>
                      </Box>
                      <Box border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="red.50">
                        <VStack spacing={2}>
                          <Text fontSize="2xl" fontWeight="bold" color="red.600">
                            {ssotPRData.payment_analysis.overdue_payments || 0}
                          </Text>
                          <Text fontSize="sm" color="red.800" textAlign="center">Overdue Payments</Text>
                        </VStack>
                      </Box>
                      <Box border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="blue.50">
                        <VStack spacing={2}>
                          <Text fontSize="2xl" fontWeight="bold" color="blue.600">
                            {ssotPRData.payment_analysis.payment_efficiency || 0}%
                          </Text>
                          <Text fontSize="sm" color="blue.800" textAlign="center">Payment Efficiency</Text>
                        </VStack>
                      </Box>
                    </SimpleGrid>
                    
                    {/* Average Payment Days */}
                    <Box mt={4} bg="purple.50" p={4} borderRadius="md" textAlign="center">
                      <Text fontSize="sm" color="purple.600" mb={2}>Average Payment Days</Text>
                      <Text fontSize="3xl" fontWeight="bold" color="purple.700">
                        {ssotPRData.payment_analysis.average_payment_days || 0} days
                      </Text>
                    </Box>
                  </Box>
                )}

                {/* Raw Data Fallback */}
                {(!ssotPRData.vendors_by_performance || ssotPRData.vendors_by_performance.length === 0) && !ssotPRData.top_vendors_by_spend && !ssotPRData.payment_analysis && (
                  <Box>
                    <Heading size="sm" mb={2} color={headingColor}>Report Data:</Heading>
                    <Box p={4} bg="gray.50" borderRadius="md" border="1px solid" borderColor="gray.200">
                      <Text fontSize="sm" color="gray.700" whiteSpace="pre-wrap" fontFamily="mono">
                        {JSON.stringify(ssotPRData, null, 2)}
                      </Text>
                    </Box>
                  </Box>
                )}
              </VStack>
            )}
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={() => setSSOTPROpen(false)}>
              Close
            </Button>
            {ssotPRData && !ssotPRLoading && (
              <Button
                leftIcon={<FiDownload />}
                colorScheme="orange"
                onClick={() => {
                  const reportData = {
                    reportType: 'Purchase Report',
                    period: `${ssotPRStartDate} to ${ssotPREndDate}`,
                    generatedOn: new Date().toISOString(),
                    data: ssotPRData
                  };
                  const dataStr = JSON.stringify(reportData, null, 2);
                  const dataBlob = new Blob([dataStr], { type: 'application/json' });
                  const url = URL.createObjectURL(dataBlob);
                  const link = document.createElement('a');
                  link.href = url;
                  link.download = `purchase-report-${ssotPRStartDate}-to-${ssotPREndDate}.json`;
                  link.click();
                  URL.revokeObjectURL(url);
                }}
              >
                Download Report
              </Button>
            )}
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* SSOT Trial Balance Modal */}
      <Modal isOpen={ssotTBOpen} onClose={() => setSSOTTBOpen(false)} size="6xl">
        <ModalOverlay />
        <ModalContent bg={modalContentBg}>
          <ModalHeader>
            <HStack>
              <Icon as={FiList} color="purple.500" />
              <VStack align="start" spacing={0}>
                <Text fontSize="lg" fontWeight="bold">
                  SSOT Trial Balance
                </Text>
                <Text fontSize="sm" color={previewPeriodTextColor}>
                  As of {ssotTBAsOfDate} | SSOT Journal Integration
                </Text>
              </VStack>
            </HStack>
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
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
                  colorScheme="blue"
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
                  <Spinner size="xl" thickness="4px" speed="0.65s" color="purple.500" />
                  <VStack spacing={2}>
                    <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                      Generating SSOT Trial Balance
                    </Text>
                    <Text fontSize="sm" color={descriptionColor}>
                      Calculating account balances from SSOT journal system...
                    </Text>
                  </VStack>
                </VStack>
              </Box>
            )}

            {ssotTBError && (
              <Box bg="red.50" p={4} borderRadius="md" mb={4}>
                <Text color="red.600" fontWeight="medium">Error: {ssotTBError}</Text>
                <Button
                  mt={2}
                  size="sm"
                  colorScheme="red"
                  variant="outline"
                  onClick={fetchSSOTTrialBalanceReport}
                >
                  Retry
                </Button>
              </Box>
            )}

            {ssotTBData && !ssotTBLoading && (
              <VStack spacing={6} align="stretch">
                {/* Company Header */}
                {ssotTBData.company && (
                  <Box bg="purple.50" p={4} borderRadius="md">
                    <HStack justify="space-between" align="start">
                      <VStack align="start" spacing={1}>
                        <Text fontSize="lg" fontWeight="bold" color="purple.800">
                          {ssotTBData.company.name || 'Company Name Not Available'}
                        </Text>
                        <Text fontSize="sm" color="purple.600">
                          {ssotTBData.company.address && ssotTBData.company.city ? 
                            `${ssotTBData.company.address}, ${ssotTBData.company.city}` : 
                            'Address not available'
                          }
                        </Text>
                        {ssotTBData.company.phone && (
                          <Text fontSize="sm" color="purple.600">
                            {ssotTBData.company.phone} | {ssotTBData.company.email}
                          </Text>
                        )}
                      </VStack>
                      <VStack align="end" spacing={1}>
                        <Text fontSize="sm" color="purple.600">
                          Currency: {ssotTBData.currency || 'IDR'}
                        </Text>
                        <Text fontSize="xs" color="purple.500">
                          Generated: {ssotTBData.generated_at ? new Date(ssotTBData.generated_at).toLocaleString('id-ID') : 'N/A'}
                        </Text>
                      </VStack>
                    </HStack>
                  </Box>
                )}

                {/* Report Header */}
                <Box textAlign="center" bg={summaryBg} p={4} borderRadius="md">
                  <Heading size="md" color={headingColor}>
                    Trial Balance Report
                  </Heading>
                  <Text fontSize="sm" color={descriptionColor}>
                    As of: {new Date(ssotTBAsOfDate).toLocaleDateString('id-ID')}
                  </Text>
                  <Text fontSize="xs" color={descriptionColor} mt={1}>
                    Generated: {new Date().toLocaleDateString('id-ID')} at {new Date().toLocaleTimeString('id-ID')}
                  </Text>
                  <HStack justify="center" mt={3}>
                    {ssotTBData.is_balanced ? (
                      <Badge colorScheme="green" size="lg" p={2}>Balanced ✓</Badge>
                    ) : (
                      <Badge colorScheme="red" size="lg" p={2}>Not Balanced ⚠</Badge>
                    )}
                  </HStack>
                </Box>

                {/* Summary Statistics */}
                <SimpleGrid columns={[1, 2, 3]} spacing={4}>
                  <Box bg="green.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="sm" color="green.600">Total Debits</Text>
                    <Text fontSize="2xl" fontWeight="bold" color="green.700">
                      {formatCurrency(ssotTBData.total_debits || 0)}
                    </Text>
                  </Box>
                  <Box bg="red.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="sm" color="red.600">Total Credits</Text>
                    <Text fontSize="2xl" fontWeight="bold" color="red.700">
                      {formatCurrency(ssotTBData.total_credits || 0)}
                    </Text>
                  </Box>
                  <Box bg="blue.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="sm" color="blue.600">Difference</Text>
                    <Text fontSize="2xl" fontWeight="bold" color={Math.abs((ssotTBData.total_debits || 0) - (ssotTBData.total_credits || 0)) === 0 ? 'green.700' : 'red.700'}>
                      {formatCurrency(Math.abs((ssotTBData.total_debits || 0) - (ssotTBData.total_credits || 0)))}
                    </Text>
                  </Box>
                </SimpleGrid>

                {/* Account Details */}
                {ssotTBData.accounts && ssotTBData.accounts.length > 0 && (
                  <Box>
                    <Heading size="sm" mb={4} color={headingColor}>
                      Account Balances ({ssotTBData.accounts.length} accounts)
                    </Heading>
                    
                    {/* Account Table Header */}
                    <Box bg="purple.50" p={3} borderRadius="md" mb={2} border="1px solid" borderColor="purple.200">
                      <SimpleGrid columns={[1, 2, 4]} spacing={2} fontSize="sm" fontWeight="bold" color="purple.800">
                        <Text>Account</Text>
                        <Text>Type</Text>
                        <Text textAlign="right">Debit Balance</Text>
                        <Text textAlign="right">Credit Balance</Text>
                      </SimpleGrid>
                    </Box>
                    
                    {/* Account Rows */}
                    <VStack spacing={1} align="stretch" maxH="400px" overflow="auto">
                      {ssotTBData.accounts.map((account, index) => (
                        <Box key={index} border="1px solid" borderColor="gray.100" borderRadius="sm" p={3} bg="white" _hover={{ bg: 'gray.50' }}>
                          <SimpleGrid columns={[1, 2, 4]} spacing={2} fontSize="sm">
                            <VStack align="start" spacing={0}>
                              <Text fontWeight="medium" color="gray.800">
                                {account.account_name || account.name || 'Unnamed Account'}
                              </Text>
                              {account.account_code && (
                                <Text fontSize="xs" color="gray.600">
                                  Code: {account.account_code}
                                </Text>
                              )}
                            </VStack>
                            <VStack align="start" spacing={0}>
                              {account.account_type && (
                                <Badge colorScheme="blue" size="sm">
                                  {account.account_type}
                                </Badge>
                              )}
                              {account.parent_account && (
                                <Text fontSize="xs" color="gray.500">
                                  Parent: {account.parent_account}
                                </Text>
                              )}
                            </VStack>
                            <Text textAlign="right" fontSize="sm" fontWeight="medium" color={account.debit_balance > 0 ? 'green.600' : 'gray.400'}>
                              {account.debit_balance > 0 ? formatCurrency(account.debit_balance) : '-'}
                            </Text>
                            <Text textAlign="right" fontSize="sm" fontWeight="medium" color={account.credit_balance > 0 ? 'red.600' : 'gray.400'}>
                              {account.credit_balance > 0 ? formatCurrency(account.credit_balance) : '-'}
                            </Text>
                          </SimpleGrid>
                        </Box>
                      ))}
                    </VStack>
                    
                    {/* Totals Row */}
                    <Box bg="purple.100" p={3} borderRadius="md" mt={2} border="2px solid" borderColor="purple.300">
                      <SimpleGrid columns={[1, 2, 4]} spacing={2} fontSize="md" fontWeight="bold">
                        <Text color="purple.800">TOTALS:</Text>
                        <Text></Text>
                        <Text textAlign="right" color="green.700">
                          {formatCurrency(ssotTBData.total_debits || 0)}
                        </Text>
                        <Text textAlign="right" color="red.700">
                          {formatCurrency(ssotTBData.total_credits || 0)}
                        </Text>
                      </SimpleGrid>
                    </Box>
                  </Box>
                )}

                {/* Account Type Summary */}
                {ssotTBData.account_type_summary && (
                  <Box>
                    <Heading size="sm" mb={4} color={headingColor}>
                      Summary by Account Type
                    </Heading>
                    <SimpleGrid columns={[1, 2, 3]} spacing={4}>
                      {Object.entries(ssotTBData.account_type_summary).map(([type, data]) => (
                        <Box key={type} border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="white">
                          <VStack spacing={2}>
                            <Text fontWeight="bold" fontSize="md" color="gray.800">
                              {type}
                            </Text>
                            <VStack spacing={1}>
                              <Text fontSize="sm" color="gray.600">
                                {data.account_count || 0} accounts
                              </Text>
                              <Text fontSize="lg" fontWeight="bold" color="blue.600">
                                {formatCurrency(data.total_balance || 0)}
                              </Text>
                            </VStack>
                          </VStack>
                        </Box>
                      ))}
                    </SimpleGrid>
                  </Box>
                )}

                {/* Raw Data Fallback */}
                {(!ssotTBData.accounts || ssotTBData.accounts.length === 0) && !ssotTBData.account_type_summary && (
                  <Box>
                    <Heading size="sm" mb={2} color={headingColor}>Report Data:</Heading>
                    <Box p={4} bg="gray.50" borderRadius="md" border="1px solid" borderColor="gray.200">
                      <Text fontSize="sm" color="gray.700" whiteSpace="pre-wrap" fontFamily="mono">
                        {JSON.stringify(ssotTBData, null, 2)}
                      </Text>
                    </Box>
                  </Box>
                )}
              </VStack>
            )}
          </ModalBody>
          <ModalFooter>
            <HStack spacing={3}>
              {ssotTBData && !ssotTBLoading && (
                <>
                  <Button
                    colorScheme="red"
                    variant="outline"
                    size="sm"
                    leftIcon={<FiFilePlus />}
                    onClick={() => handleQuickDownload({id: 'trial-balance', name: 'Trial Balance'}, 'pdf')}
                  >
                    Export PDF
                  </Button>
                  <Button
                    colorScheme="green"
                    variant="outline"
                    size="sm"
                    leftIcon={<FiFileText />}
                    onClick={() => handleQuickDownload({id: 'trial-balance', name: 'Trial Balance'}, 'csv')}
                  >
                    Export CSV
                  </Button>
                </>
              )}
            </HStack>
            <Button variant="ghost" onClick={() => setSSOTTBOpen(false)}>
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
              <Icon as={FiBook} color="indigo.500" />
              <VStack align="start" spacing={0}>
                <Text fontSize="lg" fontWeight="bold">
                  SSOT General Ledger
                </Text>
                <Text fontSize="sm" color={previewPeriodTextColor}>
                  {ssotGLStartDate} - {ssotGLEndDate} | SSOT Journal Integration
                </Text>
              </VStack>
            </HStack>
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            <Box mb={4}>
              <HStack spacing={4} mb={4}>
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
                <FormControl>
                  <FormLabel>Account ID (Optional)</FormLabel>
                  <Input 
                    type="text" 
                    value={ssotGLAccountId} 
                    onChange={(e) => setSSOTGLAccountId(e.target.value)}
                    placeholder="Leave empty for all accounts"
                  />
                </FormControl>
                <Button
                  colorScheme="blue"
                  onClick={fetchSSOTGeneralLedgerReport}
                  isLoading={ssotGLLoading}
                  leftIcon={<FiBook />}
                  size="md"
                  mt={8}
                >
                  Generate Report
                </Button>
              </HStack>
            </Box>

            {ssotGLLoading && (
              <Box textAlign="center" py={8}>
                <VStack spacing={4}>
                  <Spinner size="xl" thickness="4px" speed="0.65s" color="indigo.500" />
                  <VStack spacing={2}>
                    <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                      Generating SSOT General Ledger
                    </Text>
                    <Text fontSize="sm" color={descriptionColor}>
                      Retrieving account transactions from SSOT journal system...
                    </Text>
                  </VStack>
                </VStack>
              </Box>
            )}

            {ssotGLError && (
              <Box bg="red.50" p={4} borderRadius="md" mb={4}>
                <Text color="red.600" fontWeight="medium">Error: {ssotGLError}</Text>
                <Button
                  mt={2}
                  size="sm"
                  colorScheme="red"
                  variant="outline"
                  onClick={fetchSSOTGeneralLedgerReport}
                >
                  Retry
                </Button>
              </Box>
            )}

            {ssotGLData && !ssotGLLoading && (
              <VStack spacing={6} align="stretch">
                {/* Company Header */}
                {ssotGLData.company && (
                  <Box bg="blue.50" p={4} borderRadius="md">
                    <HStack justify="space-between" align="start">
                      <VStack align="start" spacing={1}>
                        <Text fontSize="lg" fontWeight="bold" color="blue.800">
                          {ssotGLData.company.name}
                        </Text>
                        <Text fontSize="sm" color="blue.600">
                          {ssotGLData.company.address}, {ssotGLData.company.city}
                        </Text>
                        {ssotGLData.company.phone && (
                          <Text fontSize="sm" color="blue.600">
                            {ssotGLData.company.phone} | {ssotGLData.company.email}
                          </Text>
                        )}
                      </VStack>
                      <VStack align="end" spacing={1}>
                        <Text fontSize="sm" color="blue.600">
                          Currency: {ssotGLData.currency || 'IDR'}
                        </Text>
                        <Text fontSize="xs" color="blue.500">
                          Generated: {ssotGLData.generated_at ? new Date(ssotGLData.generated_at).toLocaleString('id-ID') : 'N/A'}
                        </Text>
                      </VStack>
                    </HStack>
                  </Box>
                )}

                {/* Account Summary */}
                {ssotGLData.account && (
                  <Box border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="white">
                    <HStack justify="space-between" mb={4}>
                      <VStack align="start" spacing={1}>
                        <Text fontSize="lg" fontWeight="bold" color="gray.800">
                          {ssotGLData.account.name || 'All Accounts'}
                        </Text>
                        {ssotGLData.account.code && (
                          <Text fontSize="sm" color="gray.600">
                            Account Code: {ssotGLData.account.code}
                          </Text>
                        )}
                        {ssotGLData.account.type && (
                          <Badge colorScheme="blue" size="sm">
                            {ssotGLData.account.type}
                          </Badge>
                        )}
                      </VStack>
                      <VStack align="end" spacing={1}>
                        <Text fontSize="lg" fontWeight="bold" color={ssotGLData.closing_balance >= 0 ? 'green.600' : 'red.600'}>
                          Closing Balance: {formatCurrency(ssotGLData.closing_balance || 0)}
                        </Text>
                        <Text fontSize="sm" color="gray.600">
                          Opening Balance: {formatCurrency(ssotGLData.opening_balance || 0)}
                        </Text>
                      </VStack>
                    </HStack>

                    {/* Balance Summary Grid */}
                    <SimpleGrid columns={[1, 2, 4]} spacing={4}>
                      <Box bg="green.50" p={3} borderRadius="md" textAlign="center">
                        <Text fontSize="lg" fontWeight="bold" color="green.600">
                          {formatCurrency(ssotGLData.total_debits || 0)}
                        </Text>
                        <Text fontSize="sm" color="green.800">Total Debits</Text>
                      </Box>
                      <Box bg="red.50" p={3} borderRadius="md" textAlign="center">
                        <Text fontSize="lg" fontWeight="bold" color="red.600">
                          {formatCurrency(ssotGLData.total_credits || 0)}
                        </Text>
                        <Text fontSize="sm" color="red.800">Total Credits</Text>
                      </Box>
                      <Box bg="blue.50" p={3} borderRadius="md" textAlign="center">
                        <Text fontSize="lg" fontWeight="bold" color="blue.600">
                          {ssotGLData.transactions ? ssotGLData.transactions.length : 0}
                        </Text>
                        <Text fontSize="sm" color="blue.800">Transactions</Text>
                      </Box>
                      <Box bg="purple.50" p={3} borderRadius="md" textAlign="center">
                        <Text fontSize="lg" fontWeight="bold" color="purple.600">
                          {formatCurrency(Math.abs((ssotGLData.total_debits || 0) - (ssotGLData.total_credits || 0)))}
                        </Text>
                        <Text fontSize="sm" color="purple.800">Net Change</Text>
                      </Box>
                    </SimpleGrid>
                  </Box>
                )}

                {/* Transaction History */}
                {ssotGLData.transactions && ssotGLData.transactions.length > 0 && (
                  <Box>
                    <Heading size="sm" mb={4} color={headingColor}>
                      Transaction History ({ssotGLData.transactions.length} entries)
                    </Heading>
                    
                    {/* Transaction Table Header */}
                    <Box bg="gray.50" p={3} borderRadius="md" mb={2}>
                      <SimpleGrid columns={[1, 2, 6]} spacing={2} fontSize="sm" fontWeight="bold" color="gray.700">
                        <Text>Date</Text>
                        <Text>Description</Text>
                        <Text>Reference</Text>
                        <Text textAlign="right">Debit</Text>
                        <Text textAlign="right">Credit</Text>
                        <Text textAlign="right">Balance</Text>
                      </SimpleGrid>
                    </Box>
                    
                    {/* Transaction Rows */}
                    <VStack spacing={1} align="stretch" maxH="400px" overflow="auto">
                      {ssotGLData.transactions.map((transaction, index) => (
                        <Box key={index} border="1px solid" borderColor="gray.100" borderRadius="sm" p={3} bg="white" _hover={{ bg: 'gray.50' }}>
                          <SimpleGrid columns={[1, 2, 6]} spacing={2} fontSize="sm">
                            <VStack align="start" spacing={0}>
                              <Text fontWeight="medium" color="gray.800">
                                {new Date(transaction.date).toLocaleDateString('id-ID')}
                              </Text>
                              {transaction.journal_code && (
                                <Text fontSize="xs" color="blue.600">
                                  {transaction.journal_code}
                                </Text>
                              )}
                            </VStack>
                            <VStack align="start" spacing={0}>
                              <Text fontSize="sm" color="gray.800" noOfLines={2}>
                                {transaction.description || 'No description'}
                              </Text>
                              {transaction.entry_type && (
                                <Badge size="sm" colorScheme={getEntryTypeBadgeColor(transaction.entry_type)}>
                                  {transaction.entry_type}
                                </Badge>
                              )}
                            </VStack>
                            <Text fontSize="sm" color="gray.600">
                              {transaction.reference || '-'}
                            </Text>
                            <Text textAlign="right" fontSize="sm" fontWeight="medium" color={transaction.debit_amount > 0 ? 'green.600' : 'gray.400'}>
                              {transaction.debit_amount > 0 ? formatCurrency(transaction.debit_amount) : '-'}
                            </Text>
                            <Text textAlign="right" fontSize="sm" fontWeight="medium" color={transaction.credit_amount > 0 ? 'red.600' : 'gray.400'}>
                              {transaction.credit_amount > 0 ? formatCurrency(transaction.credit_amount) : '-'}
                            </Text>
                            <Text textAlign="right" fontSize="sm" fontWeight="bold" color={transaction.balance >= 0 ? 'green.600' : 'red.600'}>
                              {formatCurrency(transaction.balance)}
                            </Text>
                          </SimpleGrid>
                        </Box>
                      ))}
                    </VStack>
                    
                    {ssotGLData.transactions.length === 0 && (
                      <Text textAlign="center" color="gray.500" py={8}>
                        No transactions found for the selected period
                      </Text>
                    )}
                  </Box>
                )}

                {/* Monthly Summary */}
                {ssotGLData.monthly_summary && ssotGLData.monthly_summary.length > 0 && (
                  <Box>
                    <Heading size="sm" mb={4} color={headingColor}>
                      Monthly Summary
                    </Heading>
                    <SimpleGrid columns={[1, 2, 3]} spacing={4}>
                      {ssotGLData.monthly_summary.map((month, index) => (
                        <Box key={index} border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="white">
                          <Text fontWeight="bold" mb={2} color="gray.800">
                            {month.month || `Month ${index + 1}`}
                          </Text>
                          <VStack spacing={1} align="stretch">
                            <HStack justify="space-between">
                              <Text fontSize="sm" color="gray.600">Debits:</Text>
                              <Text fontSize="sm" fontWeight="medium" color="green.600">
                                {formatCurrency(month.total_debits || 0)}
                              </Text>
                            </HStack>
                            <HStack justify="space-between">
                              <Text fontSize="sm" color="gray.600">Credits:</Text>
                              <Text fontSize="sm" fontWeight="medium" color="red.600">
                                {formatCurrency(month.total_credits || 0)}
                              </Text>
                            </HStack>
                            <HStack justify="space-between">
                              <Text fontSize="sm" color="gray.600">Net:</Text>
                              <Text fontSize="sm" fontWeight="bold" color={(month.total_debits || 0) >= (month.total_credits || 0) ? 'green.600' : 'red.600'}>
                                {formatCurrency((month.total_debits || 0) - (month.total_credits || 0))}
                              </Text>
                            </HStack>
                          </VStack>
                        </Box>
                      ))}
                    </SimpleGrid>
                  </Box>
                )}
              </VStack>
            )}
          </ModalBody>
          <ModalFooter>
            <HStack spacing={3}>
              {ssotGLData && !ssotGLLoading && (
                <>
                  <Button
                    colorScheme="red"
                    variant="outline"
                    size="sm"
                    leftIcon={<FiFilePlus />}
                    onClick={() => handleQuickDownload({id: 'general-ledger', name: 'General Ledger'}, 'pdf')}
                  >
                    Export PDF
                  </Button>
                  <Button
                    colorScheme="green"
                    variant="outline"
                    size="sm"
                    leftIcon={<FiFileText />}
                    onClick={() => handleQuickDownload({id: 'general-ledger', name: 'General Ledger'}, 'csv')}
                  >
                    Export CSV
                  </Button>
                </>
              )}
            </HStack>
            <Button variant="ghost" onClick={() => setSSOTGLOpen(false)}>
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
                  {ssotJAStartDate} - {ssotJAEndDate} | SSOT Journal Integration
                </Text>
              </VStack>
            </HStack>
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
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
                  colorScheme="blue"
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
                  <VStack spacing={2}>
                    <Text fontSize="lg" fontWeight="medium" color={loadingTextColor}>
                      Generating SSOT Journal Analysis
                    </Text>
                    <Text fontSize="sm" color={descriptionColor}>
                      Analyzing journal entries from SSOT system...
                    </Text>
                  </VStack>
                </VStack>
              </Box>
            )}

            {ssotJAError && (
              <Box bg="red.50" p={4} borderRadius="md" mb={4}>
                <Text color="red.600" fontWeight="medium">Error: {ssotJAError}</Text>
                <Button
                  mt={2}
                  size="sm"
                  colorScheme="red"
                  variant="outline"
                  onClick={fetchSSOTJournalAnalysisReport}
                >
                  Retry
                </Button>
              </Box>
            )}

            {ssotJAData && !ssotJALoading && (
              <VStack spacing={6} align="stretch">
                {/* Company Header - Show even if empty */}
                <Box bg="teal.50" p={4} borderRadius="md">
                  <HStack justify="space-between" align="start">
                    <VStack align="start" spacing={1}>
                      <Text fontSize="lg" fontWeight="bold" color="teal.800">
                        {ssotJAData.company?.name || 'Company Name Not Available'}
                      </Text>
                      <Text fontSize="sm" color="teal.600">
                        {ssotJAData.company?.address && ssotJAData.company?.city ? 
                          `${ssotJAData.company.address}, ${ssotJAData.company.city}` : 
                          'Address not available'
                        }
                      </Text>
                      <Text fontSize="sm" color="teal.600">
                        {ssotJAData.company?.phone || 'Phone not available'} | {ssotJAData.company?.email || 'Email not available'}
                      </Text>
                    </VStack>
                    <VStack align="end" spacing={1}>
                      <Text fontSize="sm" color="teal.600">
                        Currency: {ssotJAData.currency || 'IDR'}
                      </Text>
                      <Text fontSize="xs" color="teal.500">
                        Generated: {ssotJAData.generated_at ? new Date(ssotJAData.generated_at).toLocaleString('id-ID') : 'N/A'}
                      </Text>
                    </VStack>
                  </HStack>
                </Box>

                {/* Report Summary */}
                <Box textAlign="center" bg={summaryBg} p={4} borderRadius="md">
                  <Heading size="md" color={headingColor}>
                    Journal Entry Analysis Report
                  </Heading>
                  <Text fontSize="sm" color={descriptionColor}>
                    Period: {ssotJAStartDate} - {ssotJAEndDate}
                  </Text>
                  <Text fontSize="xs" color={descriptionColor} mt={1}>
                    Generated: {new Date().toLocaleDateString('id-ID')} at {new Date().toLocaleTimeString('id-ID')}
                  </Text>
                </Box>

                {/* Summary Statistics Grid */}
                <SimpleGrid columns={[1, 2, 4]} spacing={4}>
                  <Box bg="teal.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="2xl" fontWeight="bold" color="teal.600">
                      {ssotJAData.total_entries || 0}
                    </Text>
                    <Text fontSize="sm" color="teal.800">Total Entries</Text>
                  </Box>
                  <Box bg="green.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="2xl" fontWeight="bold" color="green.600">
                      {ssotJAData.posted_entries || 0}
                    </Text>
                    <Text fontSize="sm" color="green.800">Posted Entries</Text>
                  </Box>
                  <Box bg="orange.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="2xl" fontWeight="bold" color="orange.600">
                      {ssotJAData.draft_entries || 0}
                    </Text>
                    <Text fontSize="sm" color="orange.800">Draft Entries</Text>
                  </Box>
                  <Box bg="red.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="2xl" fontWeight="bold" color="red.600">
                      {ssotJAData.reversed_entries || 0}
                    </Text>
                    <Text fontSize="sm" color="red.800">Reversed Entries</Text>
                  </Box>
                </SimpleGrid>

                {/* Total Amount */}
                {ssotJAData.total_amount && (
                  <Box bg="blue.50" p={4} borderRadius="md" textAlign="center">
                    <Text fontSize="sm" color="blue.600" mb={2}>Total Transaction Amount</Text>
                    <Text fontSize="3xl" fontWeight="bold" color="blue.700">
                      {formatCurrency(parseFloat(ssotJAData.total_amount))}
                    </Text>
                  </Box>
                )}

                {/* Entry Type Analysis */}
                {ssotJAData.entries_by_type && ssotJAData.entries_by_type.length > 0 && (
                  <Box>
                    <Heading size="sm" mb={4} color={headingColor}>
                      Entry Type Distribution
                    </Heading>
                    <SimpleGrid columns={[1, 2]} spacing={4}>
                      {ssotJAData.entries_by_type.map((type, index) => (
                        <Box key={index} border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="white">
                          <HStack justify="space-between" align="start">
                            <VStack align="start" spacing={2}>
                              <Badge colorScheme={getEntryTypeBadgeColor(type.source_type)} size="lg">
                                {type.source_type}
                              </Badge>
                              <VStack align="start" spacing={1}>
                                <Text fontSize="lg" fontWeight="bold" color="gray.800">
                                  {type.count} entries
                                </Text>
                                <Text fontSize="sm" color="gray.600">
                                  {type.percentage.toFixed(2)}% of total
                                </Text>
                              </VStack>
                            </VStack>
                            <VStack align="end" spacing={1}>
                              <Text fontSize="lg" fontWeight="bold" color="blue.600">
                                {formatCurrency(parseFloat(type.total_amount))}
                              </Text>
                              <Text fontSize="xs" color="gray.500">
                                Total Amount
                              </Text>
                            </VStack>
                          </HStack>
                          
                          {/* Progress Bar */}
                          <Box mt={3}>
                            <Box bg="gray.200" height="8px" borderRadius="full" overflow="hidden">
                              <Box 
                                bg={`${getEntryTypeBadgeColor(type.source_type)}.400`}
                                height="100%" 
                                width={`${type.percentage}%`}
                                borderRadius="full"
                              />
                            </Box>
                          </Box>
                        </Box>
                      ))}
                    </SimpleGrid>
                  </Box>
                )}

                {/* Compliance Check */}
                {ssotJAData.compliance_check && (
                  <Box>
                    <Heading size="sm" mb={4} color={headingColor}>
                      Compliance Assessment
                    </Heading>
                    <Box border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="white">
                      <SimpleGrid columns={[1, 3]} spacing={4} mb={4}>
                        <VStack>
                          <Text fontSize="2xl" fontWeight="bold" color="blue.600">
                            {ssotJAData.compliance_check.total_checks || 0}
                          </Text>
                          <Text fontSize="sm" color="gray.600" textAlign="center">Total Checks</Text>
                        </VStack>
                        <VStack>
                          <Text fontSize="2xl" fontWeight="bold" color="green.600">
                            {ssotJAData.compliance_check.passed_checks || 0}
                          </Text>
                          <Text fontSize="sm" color="gray.600" textAlign="center">Passed</Text>
                        </VStack>
                        <VStack>
                          <Text fontSize="2xl" fontWeight="bold" color="red.600">
                            {ssotJAData.compliance_check.failed_checks || 0}
                          </Text>
                          <Text fontSize="sm" color="gray.600" textAlign="center">Failed</Text>
                        </VStack>
                      </SimpleGrid>
                      
                      {/* Compliance Score */}
                      <Box textAlign="center" p={3} bg="gray.50" borderRadius="md">
                        <Text fontSize="sm" color="gray.600" mb={1}>Compliance Score</Text>
                        <Text fontSize="2xl" fontWeight="bold" color={ssotJAData.compliance_check.compliance_score >= 80 ? 'green.600' : ssotJAData.compliance_check.compliance_score >= 60 ? 'orange.600' : 'red.600'}>
                          {ssotJAData.compliance_check.compliance_score || 0}%
                        </Text>
                      </Box>
                    </Box>
                  </Box>
                )}

                {/* Data Quality Metrics */}
                {ssotJAData.data_quality_metrics && (
                  <Box>
                    <Heading size="sm" mb={4} color={headingColor}>
                      Data Quality Assessment
                    </Heading>
                    <Box border="1px solid" borderColor="gray.200" borderRadius="md" p={4} bg="white">
                      <SimpleGrid columns={[1, 2, 4]} spacing={4}>
                        <VStack>
                          <Text fontSize="xl" fontWeight="bold" color={ssotJAData.data_quality_metrics.overall_score >= 80 ? 'green.600' : ssotJAData.data_quality_metrics.overall_score >= 60 ? 'orange.600' : 'red.600'}>
                            {ssotJAData.data_quality_metrics.overall_score || 0}%
                          </Text>
                          <Text fontSize="sm" color="gray.600" textAlign="center">Overall Score</Text>
                        </VStack>
                        <VStack>
                          <Text fontSize="xl" fontWeight="bold" color="blue.600">
                            {ssotJAData.data_quality_metrics.completeness_score || 0}%
                          </Text>
                          <Text fontSize="sm" color="gray.600" textAlign="center">Completeness</Text>
                        </VStack>
                        <VStack>
                          <Text fontSize="xl" fontWeight="bold" color="purple.600">
                            {ssotJAData.data_quality_metrics.accuracy_score || 0}%
                          </Text>
                          <Text fontSize="sm" color="gray.600" textAlign="center">Accuracy</Text>
                        </VStack>
                        <VStack>
                          <Text fontSize="xl" fontWeight="bold" color="teal.600">
                            {ssotJAData.data_quality_metrics.consistency_score || 0}%
                          </Text>
                          <Text fontSize="sm" color="gray.600" textAlign="center">Consistency</Text>
                        </VStack>
                      </SimpleGrid>
                    </Box>
                  </Box>
                )}
              </VStack>
            )}
          </ModalBody>
          <ModalFooter>
            <HStack spacing={3}>
              {ssotJAData && !ssotJALoading && (
                <>
                  <Button
                    colorScheme="red"
                    variant="outline"
                    size="sm"
                    leftIcon={<FiFilePlus />}
                    onClick={() => handleQuickDownload({id: 'journal-entry-analysis', name: 'Journal Entry Analysis'}, 'pdf')}
                  >
                    Export PDF
                  </Button>
                  <Button
                    colorScheme="green"
                    variant="outline"
                    size="sm"
                    leftIcon={<FiFileText />}
                    onClick={() => handleQuickDownload({id: 'journal-entry-analysis', name: 'Journal Entry Analysis'}, 'csv')}
                  >
                    Export CSV
                  </Button>
                </>
              )}
            </HStack>
            <Button variant="ghost" onClick={() => setSSOTJAOpen(false)}>
              Close
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
      
    </SimpleLayout>
  );
};

export default ReportsPage;