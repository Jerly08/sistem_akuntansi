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
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalCloseButton,
  useToast,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Input,
  InputGroup,
  InputLeftElement,
  HStack,
  Select,
  Text,
  Badge,
  Spinner,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  MenuDivider,
} from '@chakra-ui/react';
import { FiPlus, FiEdit, FiTrash2, FiDownload, FiSearch } from 'react-icons/fi';
import AccountForm from '@/components/accounts/AccountForm';
import AccountTreeView from '@/components/accounts/AccountTreeView';
import { Account, AccountCreateRequest, AccountUpdateRequest } from '@/types/account';
import accountService from '@/services/accountService';

const AccountsPage = () => {
  const { token } = useAuth();
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [hierarchyAccounts, setHierarchyAccounts] = useState<Account[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tabIndex, setTabIndex] = useState(0);
  const [searchTerm, setSearchTerm] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const toast = useToast();

  // Fetch accounts from API
  const fetchAccounts = async () => {
    if (!token) return;
    
    setIsLoading(true);
    try {
      const accountsData = await accountService.getAccounts(token);
      setAccounts(accountsData);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to load accounts');
      console.error('Error fetching accounts:', err);
    } finally {
      setIsLoading(false);
    }
  };

  // Get account hierarchy
  const fetchHierarchy = async () => {
    if (!token) return;
    
    try {
      const hierarchyData = await accountService.getAccountHierarchy(token);
      setHierarchyAccounts(hierarchyData);
    } catch (err: any) {
      console.error('Error fetching hierarchy:', err);
    }
  };

  // Load accounts on component mount
  useEffect(() => {
    if (token) {
      fetchAccounts();
      fetchHierarchy();
    }
  }, [token]);

  // Handle form submission for create/update
  const handleSubmit = async (accountData: AccountCreateRequest | AccountUpdateRequest) => {
    console.log('handleSubmit called with data:', accountData);
    console.log('selectedAccount:', selectedAccount);
    
    setIsSubmitting(true);
    setError(null);
    
    try {
      if (selectedAccount) {
        // Update existing account
        console.log('Updating account with code:', selectedAccount.code);
        console.log('Update data:', accountData);
        const result = await accountService.updateAccount(token!, selectedAccount.code, accountData as AccountUpdateRequest);
        console.log('Update result:', result);
        toast({
          title: 'Account updated',
          description: 'Account has been updated successfully.',
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      } else {
        // Create new account
        console.log('Creating new account with data:', accountData);
        const result = await accountService.createAccount(token!, accountData as AccountCreateRequest);
        console.log('Create result:', result);
        toast({
          title: 'Account created',
          description: 'New account has been created successfully.',
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      }
      
      // Refresh accounts list
      fetchAccounts();
      fetchHierarchy();
      
      // Close modal
      setIsModalOpen(false);
      setSelectedAccount(null);
    } catch (err: any) {
      const errorMessage = err.message || `Error ${selectedAccount ? 'updating' : 'creating'} account`;
      console.error('Submit error:', err);
      setError(errorMessage);
      toast({
        title: 'Error',
        description: errorMessage,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle account deletion
  const handleDelete = async (account: Account) => {
    console.log('Delete account:', account); // Debug log
    if (!window.confirm(`Are you sure you want to delete account "${account.name}"?`)) {
      return;
    }
    
    try {
      await accountService.deleteAccount(token!, account.code);
      toast({
        title: 'Account deleted',
        description: 'Account has been deleted successfully.',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      
      // Refresh accounts list
      fetchAccounts();
      fetchHierarchy();
    } catch (err: any) {
      const errorMessage = err.message || 'Error deleting account';
      toast({
        title: 'Error',
        description: errorMessage,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  // Handle download template
  const handleDownloadTemplate = async () => {
    try {
      const blob = await accountService.downloadTemplate();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.style.display = 'none';
      a.href = url;
      a.download = 'accounts_import_template.csv';
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (err: any) {
      toast({
        title: 'Download failed',
        description: err.message || 'Failed to download template',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  // Handle download PDF
  const handleDownloadPDF = async () => {
    try {
      const blob = await accountService.exportAccountsPDF(token!);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.style.display = 'none';
      a.href = url;
      a.download = `chart_of_accounts_${new Date().toISOString().split('T')[0]}.pdf`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      toast({
        title: 'Download successful',
        description: 'Chart of Accounts PDF has been downloaded successfully.',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (err: any) {
      toast({
        title: 'Download failed',
        description: err.message || 'Failed to download PDF',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  // Handle download Excel
  const handleDownloadExcel = async () => {
    try {
      const blob = await accountService.exportAccountsExcel(token!);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.style.display = 'none';
      a.href = url;
      a.download = `chart_of_accounts_${new Date().toISOString().split('T')[0]}.xlsx`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      toast({
        title: 'Download successful',
        description: 'Chart of Accounts Excel has been downloaded successfully.',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (err: any) {
      toast({
        title: 'Download failed',
        description: err.message || 'Failed to download Excel',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  // Open modal for creating a new account
  const handleCreate = () => {
    setSelectedAccount(null);
    setIsModalOpen(true);
  };

  // Open modal for editing an existing account
  const handleEdit = (account: Account) => {
    console.log('Edit account:', account); // Debug log
    setSelectedAccount(account);
    setIsModalOpen(true);
  };

  // Filter accounts based on search and type
  const filteredAccounts = accounts.filter(account => {
    const matchesSearch = account.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         account.code.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesType = !typeFilter || account.type === typeFilter;
    return matchesSearch && matchesType;
  });

  // Table columns definition
  const columns = [
    { header: 'Code', accessor: 'code' },
    { header: 'Name', accessor: 'name' },
    { 
      header: 'Type', 
      accessor: (account: Account) => (
        <Badge colorScheme={accountService.getAccountTypeColor(account.type)}>
          {accountService.getAccountTypeLabel(account.type)}
        </Badge>
      )
    },
    { 
      header: 'Balance', 
      accessor: (account: Account) => (
        <Text color={account.balance >= 0 ? 'green.600' : 'red.600'}>
          {accountService.formatBalance(account.balance)}
        </Text>
      )
    },
    { 
      header: 'Status', 
      accessor: (account: Account) => (
        <Badge colorScheme={account.is_active ? 'green' : 'gray'}>
          {account.is_active ? 'Active' : 'Inactive'}
        </Badge>
      )
    },
  ];

  // Action buttons for each row
  const renderActions = (account: Account) => (
    <HStack spacing={2}>
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
        onClick={() => handleDelete(account)}
        isDisabled={!account.is_active}
      >
        Delete
      </Button>
    </HStack>
  );

  return (
    <Layout allowedRoles={['ADMIN', 'FINANCE']}>
      <Box>
        <Flex justify="space-between" align="center" mb={6}>
          <Heading size="lg">Chart of Accounts</Heading>
          <HStack spacing={3}>
            <Menu>
              <MenuButton as={Button} variant="outline" leftIcon={<FiDownload />}>
                Download
              </MenuButton>
              <MenuList>
                <MenuItem onClick={handleDownloadPDF}>
                  Download PDF
                </MenuItem>
                <MenuItem onClick={handleDownloadExcel}>
                  Download Excel
                </MenuItem>
                <MenuDivider />
                <MenuItem onClick={handleDownloadTemplate}>
                  Download CSV Template
                </MenuItem>
              </MenuList>
            </Menu>
            <Button
              colorScheme="brand"
              leftIcon={<FiPlus />}
              onClick={handleCreate}
            >
              Add Account
            </Button>
          </HStack>
        </Flex>
        
        {error && (
          <Alert status="error" mb={4}>
            <AlertIcon />
            <AlertTitle mr={2}>Error!</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <Tabs index={tabIndex} onChange={setTabIndex}>
          <TabList>
            <Tab>List View</Tab>
            <Tab>Tree View</Tab>
          </TabList>

          <TabPanels>
            <TabPanel px={0}>
              {/* Search and Filter */}
              <HStack spacing={4} mb={4}>
                <InputGroup maxW="300px">
                  <InputLeftElement pointerEvents="none">
                    <FiSearch color="gray.300" />
                  </InputLeftElement>
                  <Input
                    placeholder="Search accounts..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                  />
                </InputGroup>
                <Select
                  placeholder="Filter by type"
                  maxW="200px"
                  value={typeFilter}
                  onChange={(e) => setTypeFilter(e.target.value)}
                >
                  <option value="ASSET">Asset</option>
                  <option value="LIABILITY">Liability</option>
                  <option value="EQUITY">Equity</option>
                  <option value="REVENUE">Revenue</option>
                  <option value="EXPENSE">Expense</option>
                </Select>
              </HStack>

              {isLoading ? (
                <Flex justify="center" py={10}>
                  <Spinner size="lg" />
                </Flex>
              ) : filteredAccounts.length === 0 ? (
                <Box textAlign="center" py={10}>
                  <Text color="gray.500" mb={4}>
                    {accounts.length === 0 ? 'No accounts found. Try creating one!' : 'No accounts match your search criteria.'}
                  </Text>
                  {accounts.length === 0 && (
                    <Button colorScheme="brand" onClick={handleCreate}>
                      Create First Account
                    </Button>
                  )}
                </Box>
              ) : (
                <Table<Account>
                  columns={columns}
                  data={filteredAccounts}
                  keyField="id"
                  title="Accounts"
                  actions={renderActions}
                  isLoading={isLoading}
                />
              )}
            </TabPanel>

            <TabPanel px={0}>
              {isLoading ? (
                <Flex justify="center" py={10}>
                  <Spinner size="lg" />
                </Flex>
              ) : (
                <AccountTreeView
                  accounts={hierarchyAccounts}
                  onEdit={handleEdit}
                  onDelete={handleDelete}
                  showActions={true}
                  showBalance={true}
                />
              )}
            </TabPanel>
          </TabPanels>
        </Tabs>
        
        <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} size="lg">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>
              {selectedAccount ? 'Edit Account' : 'Create Account'}
            </ModalHeader>
            <ModalCloseButton />
            <ModalBody pb={6}>
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

