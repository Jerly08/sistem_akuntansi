'use client';

import React, { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import Layout from '@/components/layout/Layout';
import Table from '@/components/common/Table';
import {
  Box,
  Flex,
  Heading,
  Button,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  useToast,
} from '@chakra-ui/react';
import { FiPlus, FiEdit, FiTrash2 } from 'react-icons/fi';
import AccountForm from '@/components/accounts/AccountForm';

// Define the Account type based on the Prisma schema
interface Account {
  id: string;
  code: string;
  name: string;
  description?: string;
  type: 'ASSET' | 'LIABILITY' | 'EQUITY' | 'REVENUE' | 'EXPENSE';
  subType?: string;
  parentAccountId?: string;
  active: boolean;
  balance: number;
  createdAt: string;
  updatedAt: string;
}

const AccountsPage = () => {
  const { token } = useAuth();
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState<Partial<Account> | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Fetch accounts from API
const fetchAccounts = async () => {
  // Data dummy untuk testing di frontend
  setAccounts([
    {
      id: '1',
      code: '1000',
      name: 'Cash',
      type: 'ASSET',
      active: true,
      balance: 50000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '2',
      code: '1100',
      name: 'Bank Account',
      type: 'ASSET',
      active: true,
      balance: 125000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '3',
      code: '1200',
      name: 'Accounts Receivable',
      type: 'ASSET',
      active: true,
      balance: 18000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '4',
      code: '1300',
      name: 'Inventory',
      type: 'ASSET',
      active: true,
      balance: 35000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '5',
      code: '1500',
      name: 'Fixed Assets',
      type: 'ASSET',
      active: true,
      balance: 250000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '6',
      code: '1600',
      name: 'Accumulated Depreciation',
      type: 'ASSET',
      active: true,
      balance: -50000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '7',
      code: '2000',
      name: 'Accounts Payable',
      type: 'LIABILITY',
      active: true,
      balance: 9200.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '8',
      code: '2100',
      name: 'Accrued Expenses',
      type: 'LIABILITY',
      active: true,
      balance: 3500.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '9',
      code: '2200',
      name: 'Wages Payable',
      type: 'LIABILITY',
      active: true,
      balance: 8500.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '10',
      code: '2300',
      name: 'Taxes Payable',
      type: 'LIABILITY',
      active: true,
      balance: 2800.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '11',
      code: '2500',
      name: 'Long-term Debt',
      type: 'LIABILITY',
      active: true,
      balance: 75000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '12',
      code: '3000',
      name: "Owner's Equity",
      type: 'EQUITY',
      active: true,
      balance: 200000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '13',
      code: '3100',
      name: 'Retained Earnings',
      type: 'EQUITY',
      active: true,
      balance: 85000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '14',
      code: '3200',
      name: 'Common Stock',
      type: 'EQUITY',
      active: true,
      balance: 100000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '15',
      code: '4000',
      name: 'Sales Revenue',
      type: 'REVENUE',
      active: true,
      balance: 150000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '16',
      code: '4100',
      name: 'Service Revenue',
      type: 'REVENUE',
      active: true,
      balance: 25000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '17',
      code: '4200',
      name: 'Interest Income',
      type: 'REVENUE',
      active: true,
      balance: 1500.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '18',
      code: '5000',
      name: 'Cost of Goods Sold',
      type: 'EXPENSE',
      active: true,
      balance: 75000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '19',
      code: '5100',
      name: 'Salaries Expense',
      type: 'EXPENSE',
      active: true,
      balance: 60000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '20',
      code: '5200',
      name: 'Rent Expense',
      type: 'EXPENSE',
      active: true,
      balance: 18000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '21',
      code: '5300',
      name: 'Utilities Expense',
      type: 'EXPENSE',
      active: true,
      balance: 4500.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '22',
      code: '5400',
      name: 'Depreciation Expense',
      type: 'EXPENSE',
      active: true,
      balance: 12000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '23',
      code: '5500',
      name: 'Supplies Expense',
      type: 'EXPENSE',
      active: true,
      balance: 3500.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
    {
      id: '24',
      code: '5600',
      name: 'Advertising Expense',
      type: 'EXPENSE',
      active: true,
      balance: 8000.00,
      createdAt: '2025-01-01',
      updatedAt: '2025-08-01',
    },
  ]);
  setIsLoading(false);
  };

  // Load accounts on component mount
  useEffect(() => {
    if (token) {
      fetchAccounts();
    }
  }, [token]);

  // Handle form submission for create/update
  const handleSubmit = async (accountData: Partial<Account>) => {
    setIsSubmitting(true);
    setError(null);
    
    try {
      const url = accountData.id
        ? `/api/accounts/${accountData.id}`
        : '/api/accounts';
        
      const method = accountData.id ? 'PUT' : 'POST';
      
      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(accountData),
      });
      
      if (!response.ok) {
        throw new Error(`Failed to ${accountData.id ? 'update' : 'create'} account`);
      }
      
      // Refresh accounts list
      fetchAccounts();
      
      // Close modal
      setIsModalOpen(false);
      setSelectedAccount(null);
    } catch (err) {
      setError(`Error ${accountData.id ? 'updating' : 'creating'} account. Please try again.`);
      console.error('Error submitting account:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle account deletion
  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this account?')) {
      return;
    }
    
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`/api/accounts/${id}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      
      if (!response.ok) {
        throw new Error('Failed to delete account');
      }
      
      // Refresh accounts list
      fetchAccounts();
    } catch (err) {
      setError('Error deleting account. Please try again.');
      console.error('Error deleting account:', err);
    } finally {
      setIsLoading(false);
    }
  };

  // Open modal for creating a new account
  const handleCreate = () => {
    setSelectedAccount(null);
    setIsModalOpen(true);
  };

  // Open modal for editing an existing account
  const handleEdit = (account: Account) => {
    setSelectedAccount(account);
    setIsModalOpen(true);
  };

  // Table columns definition
  const columns = [
    { header: 'Code', accessor: 'code' },
    { header: 'Name', accessor: 'name' },
    { header: 'Type', accessor: 'type' },
    { header: 'Balance', accessor: (account: Account) => `$${account.balance.toFixed(2)}` },
    { header: 'Status', accessor: (account: Account) => (account.active ? 'Active' : 'Inactive') },
  ];

  const toast = useToast();
  const { isOpen, onOpen, onClose } = useDisclosure();

  // Action buttons for each row
  const renderActions = (account: Account) => (
    <>
      <Button
        size="sm"
        variant="outline"
        leftIcon={<FiEdit />}
        onClick={() => handleEdit(account)}
      >
        Edit
      </Button>
      <Button
        size="sm"
        colorScheme="red"
        variant="outline"
        leftIcon={<FiTrash2 />}
        onClick={() => handleDelete(account.id)}
      >
        Delete
      </Button>
    </>
  );

  return (
    <Layout allowedRoles={['ADMIN', 'FINANCE']}>
      <Box>
        <Flex justify="space-between" align="center" mb={6}>
          <Heading size="lg">Chart of Accounts</Heading>
          <Button
            colorScheme="brand"
            leftIcon={<FiPlus />}
            onClick={handleCreate}
          >
            Add Account
          </Button>
        </Flex>
        
        {error && (
          <Alert status="error" mb={4}>
            <AlertIcon />
            <AlertTitle mr={2}>Error!</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        
        <Table<Account>
          columns={columns}
          data={accounts}
          keyField="id"
          title="Accounts"
          actions={renderActions}
          isLoading={isLoading}
        />
        
        <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} size="lg">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>
              {selectedAccount?.id ? 'Edit Account' : 'Create Account'}
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <AccountForm
                account={selectedAccount || undefined}
                parentAccounts={accounts.filter(a => a.id !== selectedAccount?.id)}
                onSubmit={handleSubmit}
                onCancel={() => setIsModalOpen(false)}
                isSubmitting={isSubmitting}
              />
            </ModalBody>
          </ModalContent>
        </Modal>
      </Box>
    </Layout>
  );
};

export default AccountsPage; 