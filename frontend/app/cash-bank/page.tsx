'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import Layout from '@/components/layout/Layout';
import { DataTable } from '@/components/common/DataTable';
import {
  Box,
  Flex,
  Heading,
  Button,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Text,
  SimpleGrid,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  Card,
  CardHeader,
  CardBody,
} from '@chakra-ui/react';
import { FiPlus, FiDollarSign, FiCreditCard } from 'react-icons/fi';

interface BankAccount {
  id: string;
  code: string;
  name: string;
  bankName: string;
  accountNumber: string;
  accountType: string;
  balance: number;
  currency?: string;
}

interface CashAccount {
  id: string;
  code: string;
  name: string;
  location: string;
  balance: number;
  currency?: string;
}

const bankColumns = [
  { header: 'Code', accessor: 'code' as keyof BankAccount },
  { header: 'Account Name', accessor: 'name' as keyof BankAccount },
  { header: 'Bank', accessor: 'bankName' as keyof BankAccount },
  { header: 'Account Number', accessor: 'accountNumber' as keyof BankAccount },
  { header: 'Type', accessor: 'accountType' as keyof BankAccount },
  { 
    header: 'Balance', 
    accessor: ((row: BankAccount) => {
      const currency = row.currency || 'IDR';
      return `${currency} ${row.balance.toLocaleString()}`;
    }) as (row: BankAccount) => React.ReactNode
  },
];

const cashColumns = [
  { header: 'Code', accessor: 'code' as keyof CashAccount },
  { header: 'Account Name', accessor: 'name' as keyof CashAccount },
  { header: 'Location', accessor: 'location' as keyof CashAccount },
  { 
    header: 'Balance', 
    accessor: ((row: CashAccount) => {
      const currency = row.currency || 'IDR';
      return `${currency} ${row.balance.toLocaleString()}`;
    }) as (row: CashAccount) => React.ReactNode
  },
];

const CashBankPage: React.FC = () => {
  const { token } = useAuth();
  const [bankAccounts, setBankAccounts] = useState<BankAccount[]>([]);
  const [cashAccounts, setCashAccounts] = useState<CashAccount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        // Data dummy untuk testing di frontend
        const bankData: BankAccount[] = [
          {
            id: '1',
            code: '1100',
            name: 'Bank BCA Main Account',
            bankName: 'Bank Central Asia',
            accountNumber: '1234567890',
            accountType: 'Checking',
            balance: 125000000,
            currency: 'IDR',
          },
          {
            id: '2',
            code: '1101',
            name: 'Bank Mandiri Operational',
            bankName: 'Bank Mandiri',
            accountNumber: '9876543210',
            accountType: 'Savings',
            balance: 85000000,
            currency: 'IDR',
          },
          {
            id: '3',
            code: '1102',
            name: 'Bank BNI Payroll Account',
            bankName: 'Bank Negara Indonesia',
            accountNumber: '5555666677',
            accountType: 'Payroll',
            balance: 45000000,
            currency: 'IDR',
          },
          {
            id: '4',
            code: '1103',
            name: 'Bank CIMB Investment',
            bankName: 'CIMB Niaga',
            accountNumber: '1111222233',
            accountType: 'Investment',
            balance: 200000000,
            currency: 'IDR',
          },
        ];

        const cashData: CashAccount[] = [
          {
            id: '1',
            code: '1000',
            name: 'Petty Cash - Main Office',
            location: 'Administration Office',
            balance: 5000000,
            currency: 'IDR',
          },
          {
            id: '2',
            code: '1001',
            name: 'Cash Register - Sales Counter',
            location: 'Sales Department',
            balance: 3500000,
            currency: 'IDR',
          },
          {
            id: '3',
            code: '1002',
            name: 'Emergency Cash Fund',
            location: 'Finance Office Safe',
            balance: 10000000,
            currency: 'IDR',
          },
          {
            id: '4',
            code: '1003',
            name: 'Travel Allowance Cash',
            location: 'HR Department',
            balance: 2500000,
            currency: 'IDR',
          },
        ];

        setBankAccounts(bankData);
        setCashAccounts(cashData);
      } catch (err: any) {
        setError(err.message || 'An error occurred while fetching cash & bank data');
      } finally {
        setLoading(false);
      }
    };

    if (token) {
      fetchData();
    }
  }, [token]);

  // Calculate totals
  const totalBankBalance = bankAccounts.reduce((sum, acc) => sum + acc.balance, 0);
  const totalCashBalance = cashAccounts.reduce((sum, acc) => sum + acc.balance, 0);
  const totalBalance = totalBankBalance + totalCashBalance;

  if (loading) {
    return (
<Layout allowedRoles={['admin', 'finance', 'director']}>
        <Box>
          <Text>Loading cash & bank data...</Text>
        </Box>
      </Layout>
    );
  }

  return (
<Layout allowedRoles={['admin', 'finance', 'director']}>
      <Box>
        <Flex justify="space-between" align="center" mb={6}>
          <Heading size="lg">Cash & Bank Management</Heading>
          <Flex gap={2}>
            <Button
              colorScheme="blue"
              leftIcon={<FiCreditCard />}
              onClick={() => console.log('Add bank account')}
            >
              Add Bank Account
            </Button>
            <Button
              colorScheme="green"
              leftIcon={<FiDollarSign />}
              onClick={() => console.log('Add cash account')}
            >
              Add Cash Account
            </Button>
          </Flex>
        </Flex>
        
        {error && (
          <Alert status="error" mb={6}>
            <AlertIcon />
            <AlertTitle mr={2}>Error!</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        
        {/* Summary Cards */}
        <SimpleGrid columns={{ base: 1, md: 3 }} spacing={6} mb={8}>
          <Card>
            <CardBody>
              <Stat>
                <StatLabel>Total Bank Balance</StatLabel>
                <StatNumber color="blue.500">
                  IDR {totalBankBalance.toLocaleString()}
                </StatNumber>
                <StatHelpText>{bankAccounts.length} accounts</StatHelpText>
              </Stat>
            </CardBody>
          </Card>
          
          <Card>
            <CardBody>
              <Stat>
                <StatLabel>Total Cash Balance</StatLabel>
                <StatNumber color="green.500">
                  IDR {totalCashBalance.toLocaleString()}
                </StatNumber>
                <StatHelpText>{cashAccounts.length} accounts</StatHelpText>
              </Stat>
            </CardBody>
          </Card>
          
          <Card>
            <CardBody>
              <Stat>
                <StatLabel>Total Balance</StatLabel>
                <StatNumber color="purple.500">
                  IDR {totalBalance.toLocaleString()}
                </StatNumber>
                <StatHelpText>Combined total</StatHelpText>
              </Stat>
            </CardBody>
          </Card>
        </SimpleGrid>

        {/* Bank Accounts Section */}
        <Box mb={8}>
          <Flex justify="space-between" align="center" mb={4}>
            <Heading size="md">Bank Accounts</Heading>
            <Button size="sm" leftIcon={<FiPlus />} onClick={() => console.log('Add bank account')}>
              Add Bank Account
            </Button>
          </Flex>
          <Box bg="white" borderRadius="lg" overflow="hidden" boxShadow="sm">
            <DataTable<BankAccount>
              columns={bankColumns}
              data={bankAccounts}
              keyField="id"
              searchable={true}
              pagination={true}
              pageSize={5}
            />
          </Box>
        </Box>

        {/* Cash Accounts Section */}
        <Box>
          <Flex justify="space-between" align="center" mb={4}>
            <Heading size="md">Cash Accounts</Heading>
            <Button size="sm" leftIcon={<FiPlus />} onClick={() => console.log('Add cash account')}>
              Add Cash Account
            </Button>
          </Flex>
          <Box bg="white" borderRadius="lg" overflow="hidden" boxShadow="sm">
            <DataTable<CashAccount>
              columns={cashColumns}
              data={cashAccounts}
              keyField="id"
              searchable={true}
              pagination={true}
              pageSize={5}
            />
          </Box>
        </Box>
      </Box>
    </Layout>
  );
};

export default CashBankPage;
