'use client';
import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Box, 
  Flex, 
  Heading, 
  Text, 
  Card,
  CardHeader,
  CardBody,
  Button,
  HStack,
  Icon,
  Spinner,
  Alert,
  AlertIcon,
  Badge
} from '@chakra-ui/react';
import {
  FiDollarSign,
  FiShoppingCart,
  FiTrendingUp,
  FiPlus,
  FiBarChart2,
  FiAlertCircle,
  FiCheckCircle,
  FiClock
} from 'react-icons/fi';
import api from '../../services/api';
import { API_ENDPOINTS } from '@/config/api';

interface FinanceDashboardData {
  invoices_pending_payment: number;
  invoices_not_paid: number;
  journals_need_posting: number;
  bank_reconciliation: {
    last_reconciled: string | null;
    days_ago: number;
    status: 'up_to_date' | 'recent' | 'needs_attention' | 'never_reconciled';
  };
  outstanding_receivables: number;
  outstanding_payables: number;
  cash_bank_balance: number;
}

export const FinanceDashboard = () => {
  const router = useRouter();
  const [data, setData] = useState<FinanceDashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchFinanceDashboardData();
  }, []);

  const fetchFinanceDashboardData = async () => {
    try {
      setLoading(true);
      const response = await api.get(API_ENDPOINTS.DASHBOARD_FINANCE);
      setData(response.data.data);
    } catch (error: any) {
      console.error('Error fetching finance dashboard data:', error);
      setError(error.response?.data?.error || 'Failed to load dashboard data');
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);
  };

  const getReconciliationStatus = (status: string) => {
    switch (status) {
      case 'up_to_date':
        return { color: 'green', icon: FiCheckCircle, text: 'Up to date' };
      case 'recent':
        return { color: 'yellow', icon: FiClock, text: 'Recent' };
      case 'needs_attention':
        return { color: 'red', icon: FiAlertCircle, text: 'Needs attention' };
      case 'never_reconciled':
        return { color: 'gray', icon: FiAlertCircle, text: 'Never reconciled' };
      default:
        return { color: 'gray', icon: FiClock, text: 'Unknown' };
    }
  };

  const getReconciliationMessage = () => {
    if (!data?.bank_reconciliation) return 'No reconciliation data';
    
    const { days_ago, status } = data.bank_reconciliation;
    
    if (status === 'never_reconciled') {
      return 'Bank belum pernah direkonsiliasi';
    }
    
    if (days_ago === 0) {
      return 'Rekonsiliasi terakhir: Hari ini';
    } else if (days_ago === 1) {
      return 'Rekonsiliasi terakhir: 1 hari lalu';
    } else {
      return `Rekonsiliasi terakhir: ${days_ago} hari lalu`;
    }
  };

  if (loading) {
    return (
      <Box>
        <Heading as="h2" size="xl" mb={6} color="gray.800">
          Dasbor Keuangan
        </Heading>
        <Flex justify="center" align="center" h="200px">
          <Spinner size="xl" color="blue.500" />
        </Flex>
      </Box>
    );
  }

  if (error) {
    return (
      <Box>
        <Heading as="h2" size="xl" mb={6} color="gray.800">
          Dasbor Keuangan
        </Heading>
        <Alert status="error">
          <AlertIcon />
          {error}
        </Alert>
      </Box>
    );
  }

  if (!data) {
    return (
      <Box>
        <Heading as="h2" size="xl" mb={6} color="gray.800">
          Dasbor Keuangan
        </Heading>
        <Alert status="info">
          <AlertIcon />
          No data available
        </Alert>
      </Box>
    );
  }

  const reconciliationStatus = getReconciliationStatus(data.bank_reconciliation.status);
  
  return (
    <Box>
      <Heading as="h2" size="xl" mb={6} color="gray.800">
        Dasbor Keuangan
      </Heading>
    
    <Flex gap={4} flexWrap="wrap" mt={4}>
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Heading as="h3" size="sm" mb={2} color="orange.600">Invoice Perlu Dibayar</Heading>
        <Text fontSize="2xl" fontWeight="bold" color="orange.500">{data.invoices_pending_payment}</Text>
        <Text fontSize="sm" color="gray.500" mt={1}>
          {formatCurrency(data.outstanding_receivables)}
        </Text>
      </Box>
      
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Heading as="h3" size="sm" mb={2} color="red.600">Invoice Belum Lunas</Heading>
        <Text fontSize="2xl" fontWeight="bold" color="red.500">{data.invoices_not_paid}</Text>
        <Text fontSize="sm" color="gray.500" mt={1}>
          {formatCurrency(data.outstanding_payables)}
        </Text>
      </Box>
      
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Heading as="h3" size="sm" mb={2} color="blue.600">Jurnal Perlu di-Posting</Heading>
        <Text fontSize="2xl" fontWeight="bold" color="blue.500">{data.journals_need_posting}</Text>
        <Text fontSize="sm" color="gray.500" mt={1}>Jurnal draft</Text>
      </Box>
      
      <Box bg="white" p={4} borderRadius="lg" boxShadow="sm" flex="1" minW="200px">
        <Flex align="center" mb={2}>
          <Heading as="h3" size="sm" color="purple.600">Rekonsiliasi Bank</Heading>
          <Badge 
            ml={2} 
            colorScheme={reconciliationStatus.color} 
            variant="subtle"
            display="flex"
            alignItems="center"
            gap={1}
          >
            <Icon as={reconciliationStatus.icon} boxSize={3} />
            {reconciliationStatus.text}
          </Badge>
        </Flex>
        <Text fontSize="sm" color="gray.700">{getReconciliationMessage()}</Text>
        <Text fontSize="sm" color="gray.500" mt={1}>
          Saldo Kas & Bank: {formatCurrency(data.cash_bank_balance)}
        </Text>
      </Box>
    </Flex>

    {/* Quick Access Section */}
    <Card mt={6}>
      <CardHeader>
        <Heading size="md" display="flex" alignItems="center">
          <Icon as={FiPlus} mr={2} color="blue.500" />
          Akses Cepat
        </Heading>
      </CardHeader>
      <CardBody>
        <HStack spacing={4} flexWrap="wrap">
          <Button
            leftIcon={<FiDollarSign />}
            colorScheme="green"
            variant="outline"
            onClick={() => router.push('/sales')}
            size="md"
          >
            Tambah Penjualan
          </Button>
          <Button
            leftIcon={<FiShoppingCart />}
            colorScheme="orange"
            variant="outline"
            onClick={() => router.push('/purchases')}
            size="md"
          >
            Tambah Pembelian
          </Button>
          <Button
            leftIcon={<FiTrendingUp />}
            colorScheme="blue"
            variant="outline"
            onClick={() => router.push('/cash-bank')}
            size="md"
          >
            Kelola Kas & Bank
          </Button>
          <Button
            leftIcon={<FiBarChart2 />}
            colorScheme="purple"
            variant="outline"
            onClick={() => router.push('/reports')}
            size="md"
          >
            Laporan Keuangan
          </Button>
        </HStack>
      </CardBody>
    </Card>
  </Box>
  );
};
