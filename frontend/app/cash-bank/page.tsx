'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useTranslation } from '@/hooks/useTranslation';
import SimpleLayout from '@/components/layout/SimpleLayout';
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
  CardBody,
  Badge,
  useToast,
  useDisclosure,
  IconButton,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  MenuDivider,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  VStack,
  HStack,
  Tooltip,
  useColorModeValue,
} from '@chakra-ui/react';
import { FiPlus, FiDollarSign, FiCreditCard, FiEdit2, FiEye, FiArrowRight, FiTrendingUp, FiTrendingDown, FiMoreVertical, FiTrash2 } from 'react-icons/fi';
import cashbankService, { CashBank, BalanceSummary } from '@/services/cashbankService';
import CashBankForm from '@/components/cashbank/CashBankForm';
import DepositWithdrawalForm from '@/components/cashbank/DepositWithdrawalForm'; // Used for withdrawal only
import DepositFormImproved from '@/components/cashbank/DepositFormImproved'; // New improved form for deposits
import TransferForm from '@/components/cashbank/TransferForm';
import TransactionHistoryModal from '@/components/cashbank/TransactionHistoryModal';

// Table columns for cash and bank accounts with COA standards
const getAccountColumns = (
  t: any,
  onEdit?: (account: CashBank) => void, 
  onView?: (account: CashBank) => void,
  onDeposit?: (account: CashBank) => void,
  onWithdraw?: (account: CashBank) => void,
  onTransfer?: (account: CashBank) => void,
  onDelete?: (account: CashBank) => void,
  textColor?: string,
  mutedTextColor?: string
) => [
  {
    header: t('cashBank.accountCode'),
    accessor: ((row: CashBank) => (
      <Box>
        <Text fontFamily="mono" fontWeight="bold" fontSize="sm" color={textColor}>
          {row.code}
        </Text>
        <Text fontSize="xs" color={mutedTextColor}>
          {row.type} Account
        </Text>
      </Box>
    )) as (row: CashBank) => React.ReactNode
  },
  {
    header: t('cashBank.accountName'),
    accessor: ((row: CashBank) => (
      <Box>
        <Text fontWeight="medium" fontSize="sm" color={textColor}>
          {row.name}
        </Text>
        <Text fontSize="xs" color={mutedTextColor}>
          {row.description || 'No description'}
        </Text>
      </Box>
    )) as (row: CashBank) => React.ReactNode
  },
  {
    header: t('cashBank.accountType'),
    accessor: ((row: CashBank) => (
      <Box textAlign="center">
        <Flex alignItems="center" gap={2} justifyContent="center" mb={1}>
          <Text fontSize="lg">{row.type === 'CASH' ? '💵' : '🏦'}</Text>
          <Badge 
            colorScheme={row.type === 'CASH' ? 'green' : 'blue'}
            variant="solid"
            size="sm"
          >
            {row.type}
          </Badge>
        </Flex>
        <Text fontSize="xs" color={mutedTextColor}>
          {row.type === 'CASH' ? 'Physical Money' : 'Bank Account'}
        </Text>
      </Box>
    )) as (row: CashBank) => React.ReactNode
  },
  {
    header: t('cashBank.glAccount'),
    accessor: ((row: CashBank) => {
      if (row.account && row.account.code && row.account.name) {
        return (
          <Box>
            <Flex alignItems="center" gap={2} mb={1}>
              <Badge size="sm" colorScheme="blue" variant="outline">
                {row.account.code}
              </Badge>
              <Text fontSize="sm" fontWeight="medium" color="blue.600" noOfLines={1}>
                {row.account.name}
              </Text>
            </Flex>
            <Flex alignItems="center" gap={1}>
              <Text fontSize="xs" color="green.600">
                ✅ {t('cashBank.integratedWithCOA')}
              </Text>
              <Badge size="xs" colorScheme="green" variant="subtle">
                {t('cashBank.asset')}
              </Badge>
            </Flex>
          </Box>
        );
      }
      return (
        <Box>
          <Flex alignItems="center" gap={2} mb={1}>
            <Badge size="sm" colorScheme="orange" variant="solid">
              UNLINKED
            </Badge>
            <Text fontSize="sm" color="orange.700" fontWeight="medium">
              No COA Link
            </Text>
          </Flex>
          <Flex alignItems="center" gap={1}>
            <Text fontSize="xs" color="red.500">
              ⚠️ Requires GL Account Setup
            </Text>
          </Flex>
        </Box>
      );
    }) as (row: CashBank) => React.ReactNode
  },
  {
    header: t('cashBank.bankDetails'),
    accessor: ((row: CashBank) => {
      if (row.type === 'BANK') {
        return (
          <Box>
            <Text fontWeight="medium" fontSize="sm" color="blue.600" mb={1}>
              {row.bank_name || 'Unknown Bank'}
            </Text>
            <Text fontSize="xs" color={textColor} fontFamily="mono">
              Account: {row.account_no || 'N/A'}
            </Text>
            <Text fontSize="xs" color={mutedTextColor}>
              Electronic Banking
            </Text>
          </Box>
        );
      }
      return (
        <Box>
          <Text fontSize="sm" color="green.600" fontWeight="medium" mb={1}>
            Cash Storage
          </Text>
          <Text fontSize="xs" color={mutedTextColor}>
            Physical cash management
          </Text>
          <Text fontSize="xs" color={mutedTextColor}>
            Manual transactions
          </Text>
        </Box>
      );
    }) as (row: CashBank) => React.ReactNode
  },
  {
    header: t('cashBank.currentBalance'),
    accessor: ((row: CashBank) => {
      const isNegative = row.balance < 0;
      const balanceColor = isNegative ? 'red.500' : (row.balance > 0 ? 'green.600' : 'gray.500');
      
      return (
        <Box textAlign="right">
          <Text 
            fontWeight="bold" 
            color={balanceColor}
            fontSize="sm"
            fontFamily="mono"
          >
            {row.currency} {Math.abs(row.balance).toLocaleString('id-ID')}
            {isNegative && ' (Dr)'}
          </Text>
          <Text fontSize="xs" color={mutedTextColor} mb={1}>
            {isNegative ? '⚠️ Overdraft' : (row.balance > 0 ? '✅ Credit Balance' : '➜ Zero Balance')}
          </Text>
          <Badge 
            size="xs" 
            colorScheme={isNegative ? 'red' : (row.balance > 0 ? 'green' : 'gray')}
            variant="subtle"
          >
            {isNegative ? 'OVERDRAFT' : (row.balance > 0 ? 'POSITIVE' : 'ZERO')}
          </Badge>
        </Box>
      );
    }) as (row: CashBank) => React.ReactNode
  },
  {
    header: t('cashBank.status'),
    accessor: ((row: CashBank) => (
      <Box textAlign="center">
        <Badge 
          colorScheme={row.is_active ? 'green' : 'red'} 
          mb={2}
          variant="solid"
        >
          {row.is_active ? 'ACTIVE' : 'INACTIVE'}
        </Badge>
        <Text fontSize="xs" color={row.is_active ? 'green.600' : 'red.500'}>
          {row.is_active ? '🟢 Operational' : '🔴 Suspended'}
        </Text>
      </Box>
    )) as (row: CashBank) => React.ReactNode
  },
  {
    header: t('cashBank.actions'),
    accessor: ((row: CashBank) => (
      <Box>
        <Menu>
          <MenuButton
            as={IconButton}
            aria-label="Account actions"
            icon={<FiMoreVertical />}
            size="sm"
            variant="ghost"
            colorScheme="gray"
          />
          <MenuList>
            <MenuItem 
              icon={<FiEye />} 
              onClick={() => onView?.(row)}
              fontSize="sm"
            >
              View Details & Transactions
            </MenuItem>
            <MenuItem 
              icon={<FiEdit2 />} 
              onClick={() => onEdit?.(row)}
              fontSize="sm"
            >
              Edit Account Info
            </MenuItem>
            <MenuDivider />
            <MenuItem 
              icon={<FiTrendingUp />} 
              onClick={() => onDeposit?.(row)}
              color="green.600"
              fontSize="sm"
              isDisabled={!row.is_active}
            >
              Make Deposit
            </MenuItem>
            <MenuItem 
              icon={<FiTrendingDown />} 
              onClick={() => onWithdraw?.(row)}
              color="orange.600"
              fontSize="sm"
              isDisabled={!row.is_active || row.balance <= 0}
            >
              Make Withdrawal
            </MenuItem>
            <MenuItem 
              icon={<FiArrowRight />} 
              onClick={() => onTransfer?.(row)}
              color="blue.600"
              fontSize="sm"
              isDisabled={!row.is_active || row.balance <= 0}
            >
              Transfer Funds
            </MenuItem>
            <MenuDivider />
            <MenuItem 
              icon={<FiTrash2 />} 
              onClick={() => onDelete?.(row)}
              color="red.600"
              fontSize="sm"
              isDisabled={row.balance !== 0}
            >
              Delete Account
            </MenuItem>
          </MenuList>
        </Menu>
        
        {/* Quick Action Buttons */}
        <HStack spacing={1} mt={2} justify="center">
          <Tooltip label="View Details" fontSize="xs">
            <IconButton
              aria-label="View details"
              icon={<FiEye />}
              size="xs"
              variant="ghost"
              colorScheme="gray"
              onClick={() => onView?.(row)}
            />
          </Tooltip>
          
          <Tooltip label="Edit Account" fontSize="xs">
            <IconButton
              aria-label="Edit account"
              icon={<FiEdit2 />}
              size="xs"
              variant="ghost"
              colorScheme="blue"
              onClick={() => onEdit?.(row)}
            />
          </Tooltip>
          
          {row.is_active && (
            <Tooltip label="Deposit" fontSize="xs">
              <IconButton
                aria-label="Make deposit"
                icon={<FiTrendingUp />}
                size="xs"
                variant="ghost"
                colorScheme="green"
                onClick={() => onDeposit?.(row)}
              />
            </Tooltip>
          )}
          
          {row.is_active && row.balance > 0 && (
            <Tooltip label="Transfer" fontSize="xs">
              <IconButton
                aria-label="Transfer funds"
                icon={<FiArrowRight />}
                size="xs"
                variant="ghost"
                colorScheme="orange"
                onClick={() => onTransfer?.(row)}
              />
            </Tooltip>
          )}
        </HStack>
      </Box>
    )) as (row: CashBank) => React.ReactNode
  }
];

const CashBankPage: React.FC = () => {
  const { token } = useAuth();
  const { t } = useTranslation();
  const toast = useToast();
  const { isOpen, onOpen, onClose } = useDisclosure();
  
  // Color mode values
  const bg = useColorModeValue('white', 'gray.800');
  const textColor = useColorModeValue('gray.800', 'white');
  const mutedTextColor = useColorModeValue('gray.500', 'gray.400');
  const modalContentBg = useColorModeValue('white', 'gray.800');
  const modalFooterBg = useColorModeValue('gray.50', 'gray.700');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const readOnlyBg = useColorModeValue('gray.50', 'gray.700');
  
  const [accounts, setAccounts] = useState<CashBank[]>([]);
  const [balanceSummary, setBalanceSummary] = useState<BalanceSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedAccount, setSelectedAccount] = useState<CashBank | null>(null);
  const [formMode, setFormMode] = useState<'create' | 'edit'>('create');
  
  // Withdrawal form states (for withdrawal only)
  const [transactionAccount, setTransactionAccount] = useState<CashBank | null>(null);
  const [transactionMode, setTransactionMode] = useState<'withdrawal'>('withdrawal');
  
  const {
    isOpen: isDetailModalOpen,
    onOpen: onDetailModalOpen,
    onClose: onDetailModalClose
  } = useDisclosure();
  
  const {
    isOpen: isTransactionModalOpen,
    onOpen: onTransactionModalOpen,
    onClose: onTransactionModalClose
  } = useDisclosure();
  
  const {
    isOpen: isTransferModalOpen,
    onOpen: onTransferModalOpen,
    onClose: onTransferModalClose
  } = useDisclosure();
  
  const {
    isOpen: isTransactionHistoryModalOpen,
    onOpen: onTransactionHistoryModalOpen,
    onClose: onTransactionHistoryModalClose
  } = useDisclosure();
  
  const {
    isOpen: isDepositModalOpen,
    onOpen: onDepositModalOpen,
    onClose: onDepositModalClose
  } = useDisclosure();
  
  // Deposit form states
  const [depositAccount, setDepositAccount] = useState<CashBank | null>(null);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const [accountsData, summaryData] = await Promise.all([
        cashbankService.getCashBankAccounts(),
        cashbankService.getBalanceSummary()
      ]);
      
      // Ensure accountsData is an array before setting state
      setAccounts(Array.isArray(accountsData) ? accountsData : []);
      setBalanceSummary(summaryData);
    } catch (err: any) {
      console.error('Error fetching cash bank data:', err);
      const errorMessage = err.response?.data?.details || err.message || 'Failed to fetch cash & bank data';
      setError(errorMessage);
      toast({
        title: 'Error',
        description: errorMessage,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (token) {
      fetchData();
    }
  }, [token]);

  const handleAddAccount = () => {
    setSelectedAccount(null);
    setFormMode('create');
    onOpen();
  };

  const handleEditAccount = (account: CashBank) => {
    setSelectedAccount(account);
    setFormMode('edit');
    onOpen();
  };

  const handleViewAccount = (account: CashBank) => {
    setSelectedAccount(account);
    onDetailModalOpen();
  };

  const handleDeposit = (account: CashBank) => {
    setDepositAccount(account);
    onDepositModalOpen();
  };

  const handleWithdraw = (account: CashBank) => {
    setTransactionAccount(account);
    setTransactionMode('withdrawal');
    onTransactionModalOpen();
  };

  const handleTransfer = (account: CashBank) => {
    setTransactionAccount(account);
    onTransferModalOpen();
  };

  const handleTransactionSuccess = () => {
    fetchData(); // Refresh data after successful transaction
  };

  const handleDelete = async (account: CashBank) => {
    if (account.balance !== 0) {
      toast({
        title: 'Cannot Delete Account',
        description: 'Account must have zero balance before deletion',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
      return;
    }

    if (window.confirm(`Are you sure you want to delete account "${account.name}"? This action cannot be undone.`)) {
      try {
        await cashbankService.deleteCashBankAccount(account.id);
        toast({
          title: 'Account Deleted',
          description: `Account "${account.name}" has been deleted successfully`,
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
        fetchData();
      } catch (error: any) {
        toast({
          title: 'Delete Failed',
          description: error.response?.data?.details || error.message || 'Failed to delete account',
          status: 'error',
          duration: 5000,
          isClosable: true,
        });
      }
    }
  };

  const handleFormSuccess = () => {
    fetchData();
  };

  const handleFixGLLinks = async () => {
    try {
      setLoading(true);
      const result = await cashbankService.fixGLAccountLinks();
      
      toast({
        title: 'GL Account Links Fixed',
        description: `Successfully fixed ${result.fixed_count} cash/bank accounts`,
        status: 'success',
        duration: 5000,
        isClosable: true,
      });
      
      // Refresh data to show updated GL links
      fetchData();
    } catch (error: any) {
      toast({
        title: 'Failed to Fix GL Links',
        description: error.response?.data?.details || error.message || 'An error occurred while fixing GL account links',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  // Separate cash and bank accounts
  const safeAccounts = Array.isArray(accounts) ? accounts : [];
  const cashAccounts = safeAccounts.filter(acc => acc.type === 'CASH' && acc.is_active);
  const bankAccounts = safeAccounts.filter(acc => acc.type === 'BANK' && acc.is_active);
  
  // Get columns with handlers
  const accountColumns = getAccountColumns(
    t,
    handleEditAccount, 
    handleViewAccount, 
    handleDeposit, 
    handleWithdraw, 
    handleTransfer, 
    handleDelete,
    textColor,
    mutedTextColor
  );

  if (loading) {
    return (
      <SimpleLayout allowedRoles={['admin', 'finance', 'director', 'employee', 'inventory_manager']}>
        <Box>
          <Text>{t('common.loading')}</Text>
        </Box>
      </SimpleLayout>
    );
  }

  return (
    <SimpleLayout allowedRoles={['admin', 'finance', 'director', 'employee', 'inventory_manager']}>
      <Box>
        <Flex justify="space-between" align="center" mb={6}>
          <Heading size="lg">{t('cashBank.title')}</Heading>
          <Button
            colorScheme="blue"
            leftIcon={<FiPlus />}
            onClick={handleAddAccount}
          >
            {t('common.add')} {t('accounts.title')}
          </Button>
        </Flex>
        
        {error && (
          <Alert status="error" mb={6}>
            <AlertIcon />
            <AlertTitle mr={2}>Error!</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        
        {/* COA Integration Information */}
        <Alert status="info" mb={6} borderRadius="lg">
          <AlertIcon />
          <Box>
            <AlertTitle>📊 Chart of Accounts Integration</AlertTitle>
            <AlertDescription>
              <Text fontSize="sm" mb={2}>
                All cash and bank accounts are automatically integrated with your Chart of Accounts (COA) for proper financial reporting and audit compliance.
              </Text>
              <Flex gap={4} alignItems="center" flexWrap="wrap">
                <Text fontSize="xs" color="blue.600">
                  🔄 Auto-creates GL accounts when needed
                </Text>
                <Text fontSize="xs" color="green.600">
                  📋 Links to Balance Sheet (Current Assets)
                </Text>
                <Text fontSize="xs" color="purple.600">
                  🔍 Full audit trail via journal entries
                </Text>
                <Button 
                  as="a" 
                  href="/accounts" 
                  target="_blank"
                  size="xs" 
                  variant="outline" 
                  colorScheme="blue"
                  leftIcon={<Text fontSize="xs">🔗</Text>}
                >
                  View Chart of Accounts
                </Button>
              </Flex>
            </AlertDescription>
          </Box>
        </Alert>
        
        {/* COA Integration Status */}
        {safeAccounts.length > 0 && (
          <Card mb={6} borderLeft="4px" borderLeftColor={safeAccounts.some(acc => !acc.account) ? 'orange.400' : 'green.400'}>
            <CardBody>
              <Flex justify="space-between" align="center">
                <Box>
                  <Text fontSize="sm" fontWeight="medium" mb={1}>
                    {safeAccounts.some(acc => !acc.account) ? '⚠️ COA Integration Status' : '✅ COA Integration Status'}
                  </Text>
                  <Text fontSize="xs" color="gray.600">
                    {safeAccounts.filter(acc => acc.account).length} of {safeAccounts.length} accounts linked to COA
                  </Text>
                </Box>
                {safeAccounts.some(acc => !acc.account) && (
                  <Button 
                    size="xs" 
                    colorScheme="orange" 
                    variant="outline"
                    onClick={handleFixGLLinks}
                    isLoading={loading}
                  >
                    Fix Missing Links
                  </Button>
                )}
              </Flex>
            </CardBody>
          </Card>
        )}
        
        {/* Summary Cards */}
        <SimpleGrid columns={{ base: 1, md: 3 }} spacing={6} mb={8}>
          <Card>
            <CardBody>
              <Stat>
                <StatLabel>Total Bank Balance</StatLabel>
                <StatNumber color="blue.500">
                  IDR {balanceSummary?.total_bank?.toLocaleString() || '0'}
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
                  IDR {balanceSummary?.total_cash?.toLocaleString() || '0'}
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
                  IDR {balanceSummary?.total_balance?.toLocaleString() || '0'}
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
            <Button 
              size="sm" 
              leftIcon={<FiCreditCard />} 
              colorScheme="blue"
              onClick={handleAddAccount}
            >
              Add Bank Account
            </Button>
          </Flex>
          {bankAccounts.length > 0 ? (
            <Box bg={bg} borderRadius="lg" overflow="hidden" boxShadow="sm">
              <DataTable<CashBank>
                columns={accountColumns}
                data={bankAccounts}
                keyField="id"
                searchable={true}
                pagination={true}
                pageSize={5}
              />
            </Box>
          ) : (
            <Card>
              <CardBody>
                <Text color={mutedTextColor} textAlign="center" py={8}>
                  No bank accounts found. Click "Add Bank Account" to create one.
                </Text>
              </CardBody>
            </Card>
          )}
        </Box>

        {/* Cash Accounts Section */}
        <Box>
          <Flex justify="space-between" align="center" mb={4}>
            <Heading size="md">Cash Accounts</Heading>
            <Button 
              size="sm" 
              leftIcon={<FiDollarSign />} 
              colorScheme="green"
              onClick={handleAddAccount}
            >
              Add Cash Account
            </Button>
          </Flex>
          {cashAccounts.length > 0 ? (
            <Box bg={bg} borderRadius="lg" overflow="hidden" boxShadow="sm">
              <DataTable<CashBank>
                columns={accountColumns}
                data={cashAccounts}
                keyField="id"
                searchable={true}
                pagination={true}
                pageSize={5}
              />
            </Box>
          ) : (
            <Card>
              <CardBody>
                <Text color={mutedTextColor} textAlign="center" py={8}>
                  No cash accounts found. Click "Add Cash Account" to create one.
                </Text>
              </CardBody>
            </Card>
          )}
        </Box>
      </Box>

      {/* Cash Bank Form Modal */}
      <CashBankForm
        isOpen={isOpen}
        onClose={onClose}
        onSuccess={handleFormSuccess}
        account={selectedAccount}
        mode={formMode}
      />

      {/* Withdrawal Modal - Using legacy form for withdrawal only */}
      <DepositWithdrawalForm
        isOpen={isTransactionModalOpen}
        onClose={onTransactionModalClose}
        onSuccess={handleTransactionSuccess}
        account={transactionAccount}
        mode={transactionMode}
      />

      {/* Transfer Modal */}
      <TransferForm
        isOpen={isTransferModalOpen}
        onClose={onTransferModalClose}
        onSuccess={handleTransactionSuccess}
        sourceAccount={transactionAccount}
      />

      {/* Account Details Modal */}
      <Modal isOpen={isDetailModalOpen} onClose={onDetailModalClose} size="4xl" scrollBehavior="inside">
        <ModalOverlay bg="blackAlpha.600" backdropFilter="blur(10px)" />
        <ModalContent 
          maxH="95vh" 
          maxW={{ base: '95vw', md: '90vw', lg: '70vw' }}
          mx={4}
          bg={modalContentBg}
        >
          {/* Payment Modal Style Header */}
          <ModalHeader 
            bg="blue.500" 
            color="white" 
            borderTopRadius="md"
            py={4}
            px={6}
          >
            <Flex alignItems="center" gap={4}>
              <Box 
                p={3} 
                bg="whiteAlpha.200" 
                borderRadius="full"
                fontSize="2xl"
              >
                {selectedAccount?.type === 'CASH' ? '💵' : '🏦'}
              </Box>
              <VStack align="start" spacing={1}>
                <Text fontSize="xl" fontWeight="bold">
                  {selectedAccount?.name}
                </Text>
                <HStack spacing={3}>
                  <Badge 
                    bg="whiteAlpha.200"
                    color="white"
                    px={2}
                    py={1}
                    borderRadius="md"
                    fontSize="xs"
                    fontWeight="medium"
                  >
                    {selectedAccount?.type}
                  </Badge>
                  <Text fontSize="sm" fontFamily="mono" opacity={0.9}>
                    {selectedAccount?.code}
                  </Text>
                  <Badge 
                    bg={selectedAccount?.is_active ? 'green.500' : 'red.500'}
                    color="white"
                    px={2}
                    py={1}
                    borderRadius="md"
                    fontSize="xs"
                  >
                    {selectedAccount?.is_active ? 'ACTIVE' : 'INACTIVE'}
                  </Badge>
                </HStack>
              </VStack>
            </Flex>
          </ModalHeader>
          <ModalCloseButton color="white" />
          
          <ModalBody p={6}>
            {selectedAccount && (
              <VStack spacing={6} align="stretch">
                {/* Balance Hero Section */}
                <Card 
                  bg="linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)" 
                  color="white" 
                  shadow="lg" 
                  borderRadius="xl"
                >
                  <CardBody p={6} textAlign="center">
                    <VStack spacing={4}>
                      <Text fontSize="md" fontWeight="medium" opacity={0.9}>Current Balance</Text>
                      <Text 
                        fontSize="3xl"
                        fontWeight="bold"
                        fontFamily="mono"
                        letterSpacing="tight"
                      >
                        {selectedAccount.currency} {Math.abs(selectedAccount.balance).toLocaleString('id-ID')}
                      </Text>
                      <Badge 
                        bg={selectedAccount.balance < 0 ? 'red.500' : selectedAccount.balance > 0 ? 'green.500' : 'gray.500'}
                        color="white"
                        px={3}
                        py={1}
                        borderRadius="full"
                        fontSize="sm"
                      >
                        {selectedAccount.balance < 0 ? '⚠️ Overdraft' : 
                         selectedAccount.balance > 0 ? '✅ Positive' : '➖ Zero Balance'}
                      </Badge>
                    </VStack>
                  </CardBody>
                </Card>

                {/* Information Grid - Compact Layout */}
                <SimpleGrid columns={{ base: 1, lg: 2 }} spacing={6}>
                  {/* Left Column */}
                  <VStack spacing={4} align="stretch">
                    {/* Account Information */}
                    <Box>
                      <HStack spacing={3} mb={3}>
                        <Text fontSize="xl">📋</Text>
                        <Text fontWeight="bold" fontSize="lg">Account Information</Text>
                      </HStack>
                      
                      <VStack spacing={3} align="stretch">
                        <Box>
                          <Text fontSize="xs" color={mutedTextColor} mb={1} textTransform="uppercase" letterSpacing="wide">Account Code</Text>
                          <Text fontWeight="bold" fontFamily="mono" bg={readOnlyBg} px={3} py={2} borderRadius="md">
                            {selectedAccount.code}
                          </Text>
                        </Box>
                        <Box>
                          <Text fontSize="xs" color={mutedTextColor} mb={1} textTransform="uppercase" letterSpacing="wide">Currency</Text>
                          <HStack spacing={2}>
                            <Text fontWeight="bold">{selectedAccount.currency}</Text>
                            <Badge colorScheme="blue" variant="subtle" fontSize="xs">
                              Indonesian Rupiah
                            </Badge>
                          </HStack>
                        </Box>
                      </VStack>
                    </Box>

                    {/* Bank Details */}
                    {selectedAccount.type === 'BANK' && (
                      <Box>
                        <HStack spacing={3} mb={3}>
                          <Text fontSize="xl">🏦</Text>
                          <Text fontWeight="bold" fontSize="lg">Bank Details</Text>
                        </HStack>
                        
                        <VStack spacing={3} align="stretch">
                          <Box>
                            <Text fontSize="xs" color={mutedTextColor} mb={1} textTransform="uppercase" letterSpacing="wide">Bank Name</Text>
                            <Text bg={readOnlyBg} px={3} py={2} borderRadius="md" color={selectedAccount.bank_name ? textColor : 'orange.600'}>
                              {selectedAccount.bank_name || 'Bank name not specified'}
                            </Text>
                          </Box>
                          <Box>
                            <Text fontSize="xs" color={mutedTextColor} mb={1} textTransform="uppercase" letterSpacing="wide">Account Number</Text>
                            <Text bg={readOnlyBg} px={3} py={2} borderRadius="md" fontFamily="mono" color={selectedAccount.account_no ? textColor : 'orange.600'}>
                              {selectedAccount.account_no || 'Account number not specified'}
                            </Text>
                          </Box>
                        </VStack>
                      </Box>
                    )}
                  </VStack>

                  {/* Right Column */}
                  <VStack spacing={4} align="stretch">
                    {/* COA Integration */}
                    <Box>
                      <HStack spacing={3} mb={3}>
                        <Text fontSize="xl">📊</Text>
                        <Text fontWeight="bold" fontSize="lg">Chart of Accounts</Text>
                      </HStack>
                      
                      {selectedAccount.account ? (
                        <Box bg="green.50" p={4} borderRadius="lg" borderLeft="4px" borderLeftColor="green.400">
                          <HStack justify="space-between" align="center" mb={2}>
                            <Badge colorScheme="blue" variant="solid" fontSize="sm" px={3} py={1}>
                              {selectedAccount.account.code}
                            </Badge>
                            <Badge colorScheme="green" variant="solid" fontSize="xs" px={2} py={1}>
                              ASSET
                            </Badge>
                          </HStack>
                          <Text fontWeight="bold" mb={1}>
                            {selectedAccount.account.name}
                          </Text>
                          <Text fontSize="sm" color="green.700">
                            ✅ Successfully integrated with Chart of Accounts
                          </Text>
                        </Box>
                      ) : (
                        <Alert status="warning" borderRadius="md" py={3}>
                          <AlertIcon boxSize={4} />
                          <Box>
                            <AlertTitle fontSize="sm">COA Integration Required</AlertTitle>
                            <AlertDescription fontSize="xs">
                              Link this account to your Chart of Accounts.
                            </AlertDescription>
                          </Box>
                        </Alert>
                      )}
                    </Box>

                    {/* Audit Information */}
                    <Box>
                      <HStack spacing={3} mb={3}>
                        <Text fontSize="xl">📅</Text>
                        <Text fontWeight="bold" fontSize="lg">Audit Information</Text>
                      </HStack>
                      
                      <VStack spacing={3} align="stretch">
                        <Box>
                          <Text fontSize="xs" color={mutedTextColor} mb={1} textTransform="uppercase" letterSpacing="wide">Created</Text>
                          <Text fontSize="sm" fontFamily="mono" bg={readOnlyBg} px={3} py={2} borderRadius="md">
                            {new Date(selectedAccount.created_at).toLocaleDateString('id-ID', {
                              weekday: 'long',
                              year: 'numeric',
                              month: 'long',
                              day: 'numeric',
                              hour: '2-digit',
                              minute: '2-digit'
                            })}
                          </Text>
                        </Box>
                        <Box>
                          <Text fontSize="xs" color={mutedTextColor} mb={1} textTransform="uppercase" letterSpacing="wide">Last Updated</Text>
                          <Text fontSize="sm" fontFamily="mono" bg={readOnlyBg} px={3} py={2} borderRadius="md">
                            {new Date(selectedAccount.updated_at).toLocaleDateString('id-ID', {
                              weekday: 'long',
                              year: 'numeric',
                              month: 'long',
                              day: 'numeric',
                              hour: '2-digit',
                              minute: '2-digit'
                            })}
                          </Text>
                        </Box>
                      </VStack>
                    </Box>
                  </VStack>
                </SimpleGrid>

                {/* Quick Actions - Compact */}
                <Box>
                  <HStack spacing={3} mb={4}>
                    <Text fontSize="xl">⚡</Text>
                    <Text fontWeight="bold" fontSize="lg">Quick Actions</Text>
                  </HStack>
                  
                  <SimpleGrid columns={{ base: 2, md: 4 }} spacing={3}>
                    <Button
                      leftIcon={<FiEdit2 />}
                      colorScheme="blue"
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        onDetailModalClose();
                        handleEditAccount(selectedAccount);
                      }}
                    >
                      Edit
                    </Button>
                    
                    <Button
                      leftIcon={<FiEye />}
                      colorScheme="purple"
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        onDetailModalClose();
                        setSelectedAccount(selectedAccount);
                        onTransactionHistoryModalOpen();
                      }}
                    >
                      History
                    </Button>
                    
                    {selectedAccount.is_active && (
                      <Button
                        leftIcon={<FiTrendingUp />}
                        colorScheme="green"
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          onDetailModalClose();
                          handleDeposit(selectedAccount);
                        }}
                      >
                        Deposit
                      </Button>
                    )}
                    
                    {selectedAccount.is_active && selectedAccount.balance > 0 && (
                      <Button
                        leftIcon={<FiArrowRight />}
                        colorScheme="orange"
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          onDetailModalClose();
                          handleTransfer(selectedAccount);
                        }}
                      >
                        Transfer
                      </Button>
                    )}
                  </SimpleGrid>
                </Box>
              </VStack>
            )}
          </ModalBody>

          {/* Modal Footer with Payment Modal Style */}
          <ModalFooter 
            bg={modalFooterBg} 
            borderBottomRadius="md"
            py={4}
            px={6}
            borderTop="1px"
            borderColor={borderColor}
          >
            <HStack spacing={3} width="full" justify="space-between">
              {/* Account Summary Info */}
              {selectedAccount && (
                <Box flex="1" display={{ base: 'none', md: 'block' }}>
                  <HStack spacing={4}>
                    <Text fontSize="sm" color={mutedTextColor}>
                      Balance: <Text as="span" fontWeight="bold" color={selectedAccount.balance < 0 ? "red.600" : selectedAccount.balance > 0 ? "green.600" : "gray.600"}>
                        {selectedAccount.currency} {Math.abs(selectedAccount.balance).toLocaleString('id-ID')}
                      </Text>
                    </Text>
                    <Text fontSize="sm" color={mutedTextColor}>
                      Status: <Text as="span" fontWeight="bold" color={selectedAccount.is_active ? "green.600" : "red.600"}>
                        {selectedAccount.is_active ? 'Active' : 'Inactive'}
                      </Text>
                    </Text>
                  </HStack>
                </Box>
              )}
              
              {/* Action Buttons */}
              <HStack spacing={3}>
                <Button 
                  variant="outline" 
                  onClick={onDetailModalClose}
                  size={{ base: 'sm', md: 'md' }}
                  minW="80px"
                >
                  Close
                </Button>
                <Button
                  colorScheme="blue"
                  leftIcon={<FiEdit2 />}
                  onClick={() => {
                    onDetailModalClose();
                    handleEditAccount(selectedAccount!);
                  }}
                  size={{ base: 'sm', md: 'md' }}
                  minW="120px"
                >
                  Edit Account
                </Button>
              </HStack>
            </HStack>
          </ModalFooter>
        </ModalContent>
      </Modal>
      
      {/* Deposit Modal - New Improved Form */}
      <DepositFormImproved
        isOpen={isDepositModalOpen}
        onClose={onDepositModalClose}
        onSuccess={handleTransactionSuccess}
        account={depositAccount}
      />
      
      {/* Transaction History Modal */}
      <TransactionHistoryModal
        isOpen={isTransactionHistoryModalOpen}
        onClose={onTransactionHistoryModalClose}
        account={selectedAccount}
      />
    </SimpleLayout>
  );
};

export default CashBankPage;
