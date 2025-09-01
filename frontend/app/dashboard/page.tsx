'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useRouter } from 'next/navigation';
import {
    AdminDashboard,
    FinanceDashboard,
    InventoryManagerDashboard,
    DirectorDashboard,
    EmployeeDashboard
} from '@/components/dashboard';
import DynamicLayout from '@/components/layout/DynamicLayout';
import {
  Flex,
  VStack,
  Spinner,
  Text,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  useToast,
} from '@chakra-ui/react';

// Define the structure of the analytics data
interface DashboardAnalytics {
  totalSales: number;
  totalPurchases: number;
  accountsReceivable: number;
  accountsPayable: number;
  monthlySales: { month: string; value: number }[];
  monthlyPurchases: { month: string; value: number }[];
  cashFlow: { month: string; inflow: number; outflow: number; balance: number }[];
  topAccounts: { name: string; balance: number; type: string }[];
  recentTransactions: any[];
}

export default function DashboardPage() {
  const { user, token } = useAuth();
  const router = useRouter();
  const [analytics, setAnalytics] = useState<DashboardAnalytics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [redirecting, setRedirecting] = useState(false);

  useEffect(() => {
    if (!user || !token) {
        // If user is not authenticated, redirect to login page
        router.push('/login');
        return;
    }

    const fetchAnalytics = async () => {
      try {
        const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/dashboard/analytics`, {
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        });

        if (!res.ok) {
            if (res.status === 401) {
                router.push('/login');
            }
          throw new Error('Gagal memuat data analitik');
        }

        const data = await res.json();
        setAnalytics(data);
      } catch (err) {
        // Use dummy data if backend is not available
        const dummyData: DashboardAnalytics = {
          totalSales: 125000000,
          totalPurchases: 85000000,
          accountsReceivable: 25000000,
          accountsPayable: 15000000,
          monthlySales: [
            { month: 'Jan', value: 8500000 },
            { month: 'Feb', value: 9200000 },
            { month: 'Mar', value: 11800000 },
            { month: 'Apr', value: 10500000 },
            { month: 'May', value: 12200000 },
            { month: 'Jun', value: 15800000 },
            { month: 'Jul', value: 13500000 },
          ],
          monthlyPurchases: [
            { month: 'Jan', value: 6500000 },
            { month: 'Feb', value: 7200000 },
            { month: 'Mar', value: 8800000 },
            { month: 'Apr', value: 8100000 },
            { month: 'May', value: 9200000 },
            { month: 'Jun', value: 11800000 },
            { month: 'Jul', value: 10300000 },
          ],
          cashFlow: [
            { month: 'Jan', inflow: 8500000, outflow: 6500000, balance: 2000000 },
            { month: 'Feb', inflow: 9200000, outflow: 7200000, balance: 2000000 },
            { month: 'Mar', inflow: 11800000, outflow: 8800000, balance: 3000000 },
            { month: 'Apr', inflow: 10500000, outflow: 8100000, balance: 2400000 },
            { month: 'May', inflow: 12200000, outflow: 9200000, balance: 3000000 },
            { month: 'Jun', inflow: 15800000, outflow: 11800000, balance: 4000000 },
            { month: 'Jul', inflow: 13500000, outflow: 10300000, balance: 3200000 },
          ],
          topAccounts: [
            { name: 'Kas', balance: 45000000, type: 'Asset' },
            { name: 'Bank BCA', balance: 125000000, type: 'Asset' },
            { name: 'Piutang Dagang', balance: 25000000, type: 'Asset' },
            { name: 'Persediaan', balance: 75000000, type: 'Asset' },
            { name: 'Utang Dagang', balance: 15000000, type: 'Liability' },
          ],
          recentTransactions: [],
        };
        setAnalytics(dummyData);
        console.warn('Using dummy data for dashboard analytics:', err instanceof Error ? err.message : 'Backend not available');
      } finally {
        setLoading(false);
      }
    };

    // Fetch analytics only for roles that need it
    if (user.role === 'ADMIN' || user.role === 'DIRECTOR') {
        fetchAnalytics();
    } else {
        setLoading(false);
    }
  }, [user, token, router]);

  // Handle unauthorized role redirect
  useEffect(() => {
    if (user && !loading && !['ADMIN', 'FINANCE', 'INVENTORY_MANAGER', 'DIRECTOR', 'EMPLOYEE'].includes(user.role)) {
      setRedirecting(true);
      router.push('/unauthorized');
    }
  }, [user, loading, router]);

  const toast = useToast();

  const renderDashboardByRole = () => {
    if (loading || redirecting) {
      return (
        <Flex justify="center" align="center" minH="60vh">
          <VStack spacing={4}>
            <Spinner size="xl" color="brand.500" thickness="4px" />
            <Text>{redirecting ? 'Mengalihkan...' : 'Memuat dasbor...'}</Text>
          </VStack>
        </Flex>
      );
    }

    if (error) {
      return (
        <Alert status="error" borderRadius="md">
          <AlertIcon />
          <AlertTitle mr={2}>Error!</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      );
    }

    switch (user?.role) {
      case 'ADMIN':
        return <AdminDashboard analytics={analytics} />;
      case 'FINANCE':
        return <FinanceDashboard />;
      case 'INVENTORY_MANAGER':
        return <InventoryManagerDashboard />;
      case 'DIRECTOR':
        return <DirectorDashboard />;
      case 'EMPLOYEE':
        return <EmployeeDashboard />;
      default:
        // Don't call router.push here, it's handled in useEffect
        return (
          <Flex justify="center" align="center" minH="60vh">
            <VStack spacing={4}>
              <Spinner size="xl" color="brand.500" thickness="4px" />
              <Text>Mengalihkan ke halaman yang sesuai...</Text>
            </VStack>
          </Flex>
        );
    }
  };

  return (
<DynamicLayout allowedRoles={['admin', 'finance', 'director', 'inventory_manager', 'employee']}>
      {renderDashboardByRole()}
    </DynamicLayout>
  );
}
