'use client';

import React from 'react';
import { useRouter } from 'next/navigation';
import {
  Box,
  SimpleGrid,
  Card,
  CardHeader,
  CardBody,
  Heading,
  Text,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  StatArrow,
  Flex,
  Icon,
  Button,
  HStack,
} from '@chakra-ui/react';
import {
  FiTrendingUp,
  FiTrendingDown,
  FiDollarSign,
  FiShoppingCart,
  FiActivity,
  FiBarChart2,
  FiPieChart,
  FiUsers,
  FiPlus,
} from 'react-icons/fi';
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart as RechartsPieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';

// This would be passed as a prop from the main dashboard page
interface DashboardAnalytics {
  totalSales: number;
  totalPurchases: number;
  accountsReceivable: number;
  accountsPayable: number;
  monthlySales: { month: string; value: number }[];
  monthlyPurchases: { month: string; value: number }[];
  cashFlow: { month: string; inflow: number; outflow: number; balance: number }[];
  topAccounts: { name: string; balance: number; type: string }[];
  recentTransactions: any[]; // Define a proper type later
}

interface AdminDashboardProps {
  analytics: DashboardAnalytics | null;
}

const StatCard = ({ icon, title, stat, change, changeType }) => (
  <Card>
    <CardHeader display="flex" flexDirection="row" alignItems="center" justifyContent="space-between" pb={2}>
      <Stat>
        <StatLabel color="gray.500">{title}</StatLabel>
        <StatNumber fontSize="2xl" fontWeight="bold">{stat}</StatNumber>
        <StatHelpText>
          <StatArrow type={changeType === 'increase' ? 'increase' : 'decrease'} />
          {change}
        </StatHelpText>
      </Stat>
      <Flex
        w={12}
        h={12}
        align="center"
        justify="center"
        borderRadius="full"
        bg={`${changeType === 'increase' ? 'green' : 'red'}.100`}
      >
        <Icon as={icon} color={`${changeType === 'increase' ? 'green' : 'red'}.500`} w={6} h={6} />
      </Flex>
    </CardHeader>
  </Card>
);

export const AdminDashboard: React.FC<AdminDashboardProps> = ({ analytics }) => {
  const router = useRouter();
  
  if (!analytics) {
    return <Box>Loading analytics...</Box>;
  }

  const formatCurrency = (value: number) =>
    new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(value);
  
  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#A28bF4'];

  // Format data for charts
  const salesPurchaseData = analytics.monthlySales.map((sale, index) => ({
    month: sale.month,
    sales: sale.value,
    purchases: analytics.monthlyPurchases[index]?.value || 0,
  }));

  const topAccountsData = analytics.topAccounts.map((account, index) => ({
    name: account.name,
    value: account.balance,
    fill: COLORS[index % COLORS.length],
  }));

  return (
    <Box>
      <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} spacing={6} mb={6}>
        <StatCard
          icon={FiDollarSign}
          title="Total Pendapatan"
          stat={formatCurrency(analytics.totalSales)}
          change="+20.1%"
          changeType="increase"
        />
        <StatCard
          icon={FiShoppingCart}
          title="Total Pembelian"
          stat={formatCurrency(analytics.totalPurchases)}
          change="+18.3%"
          changeType="increase"
        />
        <StatCard
          icon={FiTrendingUp}
          title="Piutang Usaha"
          stat={formatCurrency(analytics.accountsReceivable)}
          change="+5.2%"
          changeType="increase"
        />
        <StatCard
          icon={FiTrendingDown}
          title="Utang Usaha"
          stat={formatCurrency(analytics.accountsPayable)}
          change="+3.1%"
          changeType="increase"
        />
      </SimpleGrid>

      {/* Quick Access Section */}
      <Card mb={6}>
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
              onClick={() => router.push('/payments')}
              size="md"
            >
              Kelola Pembayaran
            </Button>
          </HStack>
        </CardBody>
      </Card>

      <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6}>
        <Card>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiActivity} mr={2} color="blue.500" />
              Tinjauan Penjualan & Pembelian
            </Heading>
          </CardHeader>
          <CardBody>
<ResponsiveContainer width="100%" height={300}>
              <LineChart data={salesPurchaseData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="month" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Line type="monotone" dataKey="sales" stroke="#8884d8" activeDot={{ r: 8 }} />
                <Line type="monotone" dataKey="purchases" stroke="#82ca9d" />
              </LineChart>
            </ResponsiveContainer>
          </CardBody>
        </Card>

        <Card>
          <CardHeader>
            <Heading size="md" display="flex" alignItems="center">
              <Icon as={FiPieChart} mr={2} color="green.500" />
              Akun Teratas
            </Heading>
          </CardHeader>
          <CardBody>
<ResponsiveContainer width="100%" height={300}>
              <RechartsPieChart>
                <Pie data={topAccountsData} innerRadius={60} outerRadius={80} fill="#8884d8" dataKey="value" label>
                  {
                    topAccountsData.map((entry, index) => <Cell key={`cell-${index}`} fill={entry.fill} />)
                  }
                </Pie>
                <Tooltip />
                <Legend />
              </RechartsPieChart>
            </ResponsiveContainer>
          </CardBody>
        </Card>
      </SimpleGrid>

      {/* Cash Flow Chart */}
      <Card mt={6}>
        <CardHeader>
          <Heading size="md" display="flex" alignItems="center">
            <Icon as={FiBarChart2} mr={2} color="purple.500" />
            Arus Kas Bulanan
          </Heading>
        </CardHeader>
        <CardBody>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={analytics.cashFlow} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="month" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar dataKey="inflow" fill="#00C49F" name="Arus Masuk" />
              <Bar dataKey="outflow" fill="#FF8042" name="Arus Keluar" />
              <Bar dataKey="balance" fill="#0088FE" name="Saldo" />
            </BarChart>
          </ResponsiveContainer>
        </CardBody>
      </Card>
    </Box>
  );
};

