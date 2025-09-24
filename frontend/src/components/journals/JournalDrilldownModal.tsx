'use client';

import React, { useState, useEffect, useMemo } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  ModalCloseButton,
  VStack,
  HStack,
  Box,
  Text,
  Badge,
  Button,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Divider,
  Card,
  CardHeader,
  CardBody,
  Flex,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  Spinner,
  Alert,
  AlertIcon,
  Tooltip,
  IconButton,
  useToast,
  useColorModeValue,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  SimpleGrid
} from '@chakra-ui/react';
import {
  FiActivity,
  FiEye,
  FiEdit,
  FiCheck,
  FiX,
  FiRefreshCw,
  FiClock,
  FiUser,
  FiFileText,
  FiTrendingUp,
  FiAlertCircle
} from 'react-icons/fi';
import { ssotJournalService, SSOTJournalEntry, SSOTJournalLine } from '@/services/ssotJournalService';
import { useBalanceMonitor, useWebSocketConnection } from '@/contexts/WebSocketContext';
import { formatCurrency } from '@/utils/formatters';

interface JournalDrilldownModalProps {
  isOpen: boolean;
  onClose: () => void;
  journalId: number | null;
  onEdit?: (journal: SSOTJournalEntry) => void;
  onPost?: (journal: SSOTJournalEntry) => void;
  onReverse?: (journal: SSOTJournalEntry) => void;
  showRealTimeMonitor?: boolean;
}

interface AccountBalanceInfo {
  account_id: number;
  account_code: string;
  account_name: string;
  current_balance: number;
  balance_type: 'DEBIT' | 'CREDIT';
  last_updated?: string;
}

const JournalDrilldownModal: React.FC<JournalDrilldownModalProps> = ({
  isOpen,
  onClose,
  journalId,
  onEdit,
  onPost,
  onReverse,
  showRealTimeMonitor = true
}) => {
  const [journal, setJournal] = useState<SSOTJournalEntry | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [accountBalances, setAccountBalances] = useState<Map<number, AccountBalanceInfo>>(new Map());
  const [isProcessing, setIsProcessing] = useState(false);

  const toast = useToast();
  
  // Real-time monitoring hooks
  const { balanceUpdates, isConnected, getAccountBalance } = useBalanceMonitor();
  const { connect, disconnect } = useWebSocketConnection();

  // Color mode values
  const cardBg = useColorModeValue('white', 'gray.800');
  const headerBg = useColorModeValue('gray.50', 'gray.700');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const successBg = useColorModeValue('green.50', 'green.900');
  const warningBg = useColorModeValue('orange.50', 'orange.900');
  const errorBg = useColorModeValue('red.50', 'red.900');

  // Load journal data
  const loadJournal = async () => {
    if (!journalId) return;

    try {
      setIsLoading(true);
      setError(null);
      
      const journalData = await ssotJournalService.getJournalEntry(journalId);
      setJournal(journalData);

      // Initialize account balances for all lines
      if (journalData.journal_lines) {
        const balancesMap = new Map<number, AccountBalanceInfo>();
        
        journalData.journal_lines.forEach(line => {
          if (line.account) {
            // Get real-time balance if available
            const realtimeBalance = getAccountBalance(line.account_id);
            
            balancesMap.set(line.account_id, {
              account_id: line.account_id,
              account_code: line.account.code,
              account_name: line.account.name,
              current_balance: realtimeBalance?.balance || 0,
              balance_type: realtimeBalance?.balance_type || 'DEBIT',
              last_updated: realtimeBalance?.updated_at
            });
          }
        });
        
        setAccountBalances(balancesMap);
      }

    } catch (error) {
      console.error('Error loading journal:', error);
      setError(error instanceof Error ? error.message : 'Failed to load journal entry');
    } finally {
      setIsLoading(false);
    }
  };

  // Load journal when modal opens or journalId changes
  useEffect(() => {
    if (isOpen && journalId) {
      loadJournal();
    }
  }, [isOpen, journalId]);

  // Update account balances from real-time updates
  useEffect(() => {
    if (!showRealTimeMonitor || !journal) return;

    journal.journal_lines?.forEach(line => {
      const realtimeBalance = getAccountBalance(line.account_id);
      if (realtimeBalance) {
        setAccountBalances(prev => {
          const updated = new Map(prev);
          const existing = updated.get(line.account_id);
          if (existing) {
            updated.set(line.account_id, {
              ...existing,
              current_balance: realtimeBalance.balance,
              balance_type: realtimeBalance.balance_type,
              last_updated: realtimeBalance.updated_at
            });
          }
          return updated;
        });
      }
    });
  }, [balanceUpdates, journal, showRealTimeMonitor, getAccountBalance]);

  // Get status color
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'POSTED': return 'green';
      case 'DRAFT': return 'yellow';
      case 'REVERSED': return 'red';
      default: return 'gray';
    }
  };

  // Calculate impact on balances
  const calculateBalanceImpact = (line: SSOTJournalLine): { impact: number; description: string } => {
    const debit = line.debit_amount || 0;
    const credit = line.credit_amount || 0;
    
    if (debit > 0) {
      return {
        impact: debit,
        description: `+${formatCurrency(debit)} (Debit)`
      };
    } else if (credit > 0) {
      return {
        impact: -credit,
        description: `-${formatCurrency(credit)} (Credit)`
      };
    }
    
    return { impact: 0, description: 'No impact' };
  };

  // Handle journal actions
  const handlePostJournal = async () => {
    if (!journal) return;

    try {
      setIsProcessing(true);
      const updatedJournal = await ssotJournalService.postJournalEntry(journal.id);
      setJournal(updatedJournal);
      
      toast({
        title: 'Journal Entry Posted',
        description: `Entry ${updatedJournal.entry_number} has been posted successfully`,
        status: 'success',
        duration: 5000,
        isClosable: true
      });

      if (onPost) {
        onPost(updatedJournal);
      }
    } catch (error) {
      toast({
        title: 'Error Posting Journal',
        description: error instanceof Error ? error.message : 'Failed to post journal entry',
        status: 'error',
        duration: 5000,
        isClosable: true
      });
    } finally {
      setIsProcessing(false);
    }
  };

  const handleReverseJournal = async () => {
    if (!journal) return;

    try {
      setIsProcessing(true);
      const reversedJournal = await ssotJournalService.reverseJournalEntry(journal.id, 'Reversed via drilldown modal');
      
      toast({
        title: 'Journal Entry Reversed',
        description: `Entry ${journal.entry_number} has been reversed`,
        status: 'warning',
        duration: 5000,
        isClosable: true
      });

      if (onReverse) {
        onReverse(reversedJournal);
      }

      onClose(); // Close modal after reversing
    } catch (error) {
      toast({
        title: 'Error Reversing Journal',
        description: error instanceof Error ? error.message : 'Failed to reverse journal entry',
        status: 'error',
        duration: 5000,
        isClosable: true
      });
    } finally {
      setIsProcessing(false);
    }
  };

  // Render loading state
  if (isLoading) {
    return (
      <Modal isOpen={isOpen} onClose={onClose} size="xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Loading Journal Entry</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4} py={8}>
              <Spinner size="xl" />
              <Text>Loading journal entry details...</Text>
            </VStack>
          </ModalBody>
        </ModalContent>
      </Modal>
    );
  }

  // Render error state
  if (error) {
    return (
      <Modal isOpen={isOpen} onClose={onClose} size="xl">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Error Loading Journal</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <Alert status="error">
              <AlertIcon />
              <VStack align="start" spacing={2}>
                <Text fontWeight="medium">Failed to load journal entry</Text>
                <Text fontSize="sm">{error}</Text>
              </VStack>
            </Alert>
          </ModalBody>
          <ModalFooter>
            <Button onClick={() => loadJournal()} leftIcon={<FiRefreshCw />} mr={3}>
              Retry
            </Button>
            <Button onClick={onClose}>Close</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    );
  }

  // Render journal details
  if (!journal) {
    return null;
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="6xl">
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>
          <HStack justify="space-between" w="full">
            <VStack align="start" spacing={1}>
              <Text>Journal Entry Details</Text>
              <HStack spacing={3}>
                <Text fontSize="sm" color="gray.500">
                  {journal.entry_number}
                </Text>
                <Badge colorScheme={getStatusColor(journal.status)} variant="solid">
                  {journal.status}
                </Badge>
                {showRealTimeMonitor && (
                  <Badge 
                    colorScheme={isConnected ? 'green' : 'red'} 
                    variant="outline"
                    fontSize="xs"
                  >
                    {isConnected ? 'Live Updates' : 'Offline'}
                  </Badge>
                )}
              </HStack>
            </VStack>

            <HStack spacing={2}>
              {journal.status === 'DRAFT' && onPost && (
                <Button
                  size="sm"
                  colorScheme="green"
                  leftIcon={<FiCheck />}
                  onClick={handlePostJournal}
                  isLoading={isProcessing}
                >
                  Post
                </Button>
              )}
              
              {journal.status === 'POSTED' && onReverse && (
                <Button
                  size="sm"
                  colorScheme="orange"
                  leftIcon={<FiX />}
                  onClick={handleReverseJournal}
                  isLoading={isProcessing}
                >
                  Reverse
                </Button>
              )}
              
              {onEdit && journal.status === 'DRAFT' && (
                <Button
                  size="sm"
                  colorScheme="blue"
                  leftIcon={<FiEdit />}
                  onClick={() => onEdit(journal)}
                >
                  Edit
                </Button>
              )}
            </HStack>
          </HStack>
        </ModalHeader>
        
        <ModalCloseButton />
        
        <ModalBody>
          <Tabs>
            <TabList>
              <Tab>Entry Details</Tab>
              <Tab>Journal Lines</Tab>
              {showRealTimeMonitor && <Tab>Real-time Balances</Tab>}
              <Tab>Audit Trail</Tab>
            </TabList>

            <TabPanels>
              {/* Entry Details Tab */}
              <TabPanel>
                <VStack spacing={6} align="stretch">
                  {/* Basic Information */}
                  <Card>
                    <CardHeader>
                      <Text fontWeight="semibold">Basic Information</Text>
                    </CardHeader>
                    <CardBody>
                      <SimpleGrid columns={[1, 2, 3]} spacing={4}>
                        <Box>
                          <Text fontSize="sm" fontWeight="medium" color="gray.500">Entry Number</Text>
                          <Text fontSize="md">{journal.entry_number}</Text>
                        </Box>
                        
                        <Box>
                          <Text fontSize="sm" fontWeight="medium" color="gray.500">Entry Date</Text>
                          <Text fontSize="md">{new Date(journal.entry_date).toLocaleDateString('id-ID')}</Text>
                        </Box>
                        
                        <Box>
                          <Text fontSize="sm" fontWeight="medium" color="gray.500">Source Type</Text>
                          <Text fontSize="md">{journal.source_type}</Text>
                        </Box>
                        
                        {journal.reference && (
                          <Box>
                            <Text fontSize="sm" fontWeight="medium" color="gray.500">Reference</Text>
                            <Text fontSize="md">{journal.reference}</Text>
                          </Box>
                        )}
                        
                        <Box>
                          <Text fontSize="sm" fontWeight="medium" color="gray.500">Auto Generated</Text>
                          <Badge colorScheme={journal.is_auto_generated ? 'blue' : 'gray'}>
                            {journal.is_auto_generated ? 'Yes' : 'No'}
                          </Badge>
                        </Box>
                        
                        <Box>
                          <Text fontSize="sm" fontWeight="medium" color="gray.500">Balanced</Text>
                          <Badge colorScheme={journal.is_balanced ? 'green' : 'red'}>
                            {journal.is_balanced ? 'Yes' : 'No'}
                          </Badge>
                        </Box>
                      </SimpleGrid>
                      
                      <Box mt={4}>
                        <Text fontSize="sm" fontWeight="medium" color="gray.500">Description</Text>
                        <Text fontSize="md">{journal.description}</Text>
                      </Box>
                      
                      {journal.notes && (
                        <Box mt={4}>
                          <Text fontSize="sm" fontWeight="medium" color="gray.500">Notes</Text>
                          <Text fontSize="md">{journal.notes}</Text>
                        </Box>
                      )}
                    </CardBody>
                  </Card>

                  {/* Financial Summary */}
                  <Card>
                    <CardHeader>
                      <Text fontWeight="semibold">Financial Summary</Text>
                    </CardHeader>
                    <CardBody>
                      <HStack spacing={8}>
                        <Stat>
                          <StatLabel>Total Debit</StatLabel>
                          <StatNumber color="blue.600">
                            {formatCurrency(journal.total_debit)}
                          </StatNumber>
                        </Stat>
                        
                        <Stat>
                          <StatLabel>Total Credit</StatLabel>
                          <StatNumber color="green.600">
                            {formatCurrency(journal.total_credit)}
                          </StatNumber>
                        </Stat>
                        
                        <Stat>
                          <StatLabel>Difference</StatLabel>
                          <StatNumber color={journal.is_balanced ? 'green.600' : 'red.600'}>
                            {formatCurrency(Math.abs(journal.total_debit - journal.total_credit))}
                          </StatNumber>
                          <StatHelpText>
                            {journal.is_balanced ? 'Balanced' : 'Not Balanced'}
                          </StatHelpText>
                        </Stat>
                      </HStack>
                    </CardBody>
                  </Card>
                </VStack>
              </TabPanel>

              {/* Journal Lines Tab */}
              <TabPanel>
                <Card>
                  <CardHeader>
                    <Text fontWeight="semibold">Journal Lines ({journal.journal_lines?.length || 0})</Text>
                  </CardHeader>
                  <CardBody p={0}>
                    <Table variant="simple">
                      <Thead bg={headerBg}>
                        <Tr>
                          <Th>Line #</Th>
                          <Th>Account</Th>
                          <Th>Description</Th>
                          <Th>Debit</Th>
                          <Th>Credit</Th>
                          <Th>Balance Impact</Th>
                        </Tr>
                      </Thead>
                      <Tbody>
                        {journal.journal_lines?.map((line, index) => {
                          const impact = calculateBalanceImpact(line);
                          
                          return (
                            <Tr key={line.id}>
                              <Td>{line.line_number}</Td>
                              <Td>
                                <VStack align="start" spacing={1}>
                                  <Text fontSize="sm" fontWeight="medium">
                                    {line.account?.code} - {line.account?.name}
                                  </Text>
                                  <Badge size="sm" colorScheme="gray">
                                    {line.account?.type}
                                  </Badge>
                                </VStack>
                              </Td>
                              <Td>
                                <Text fontSize="sm">{line.description}</Text>
                              </Td>
                              <Td>
                                <Text fontSize="sm" fontWeight="medium" color="blue.600">
                                  {line.debit_amount ? formatCurrency(line.debit_amount) : '-'}
                                </Text>
                              </Td>
                              <Td>
                                <Text fontSize="sm" fontWeight="medium" color="green.600">
                                  {line.credit_amount ? formatCurrency(line.credit_amount) : '-'}
                                </Text>
                              </Td>
                              <Td>
                                <Text 
                                  fontSize="sm" 
                                  fontWeight="medium"
                                  color={impact.impact >= 0 ? 'blue.600' : 'green.600'}
                                >
                                  {impact.description}
                                </Text>
                              </Td>
                            </Tr>
                          );
                        })}
                      </Tbody>
                    </Table>
                  </CardBody>
                </Card>
              </TabPanel>

              {/* Real-time Balances Tab */}
              {showRealTimeMonitor && (
                <TabPanel>
                  <VStack spacing={4} align="stretch">
                    <Card>
                      <CardHeader>
                        <Flex justify="space-between" align="center">
                          <Text fontWeight="semibold">Current Account Balances</Text>
                          <HStack spacing={2}>
                            <Badge colorScheme={isConnected ? 'green' : 'red'}>
                              {isConnected ? 'Live' : 'Offline'}
                            </Badge>
                            {!isConnected && (
                              <Button size="xs" onClick={connect}>
                                Connect
                              </Button>
                            )}
                          </HStack>
                        </Flex>
                      </CardHeader>
                      <CardBody p={0}>
                        <Table variant="simple">
                          <Thead bg={headerBg}>
                            <Tr>
                              <Th>Account</Th>
                              <Th>Current Balance</Th>
                              <Th>Balance Type</Th>
                              <Th>Last Updated</Th>
                              <Th>Journal Impact</Th>
                            </Tr>
                          </Thead>
                          <Tbody>
                            {Array.from(accountBalances.entries()).map(([accountId, balanceInfo]) => {
                              const journalLine = journal.journal_lines?.find(line => line.account_id === accountId);
                              const impact = journalLine ? calculateBalanceImpact(journalLine) : { impact: 0, description: 'No impact' };
                              
                              return (
                                <Tr key={accountId}>
                                  <Td>
                                    <VStack align="start" spacing={1}>
                                      <Text fontSize="sm" fontWeight="medium">
                                        {balanceInfo.account_code} - {balanceInfo.account_name}
                                      </Text>
                                    </VStack>
                                  </Td>
                                  <Td>
                                    <Text
                                      fontSize="sm"
                                      fontWeight="medium"
                                      color={balanceInfo.current_balance >= 0 ? 'green.600' : 'red.600'}
                                    >
                                      {formatCurrency(balanceInfo.current_balance)}
                                    </Text>
                                  </Td>
                                  <Td>
                                    <Badge colorScheme={balanceInfo.balance_type === 'DEBIT' ? 'blue' : 'green'}>
                                      {balanceInfo.balance_type}
                                    </Badge>
                                  </Td>
                                  <Td>
                                    <Text fontSize="xs" color="gray.500">
                                      {balanceInfo.last_updated
                                        ? new Date(balanceInfo.last_updated).toLocaleString()
                                        : 'Not updated'
                                      }
                                    </Text>
                                  </Td>
                                  <Td>
                                    <Text
                                      fontSize="sm"
                                      color={impact.impact >= 0 ? 'blue.600' : 'green.600'}
                                    >
                                      {impact.description}
                                    </Text>
                                  </Td>
                                </Tr>
                              );
                            })}
                          </Tbody>
                        </Table>
                      </CardBody>
                    </Card>
                  </VStack>
                </TabPanel>
              )}

              {/* Audit Trail Tab */}
              <TabPanel>
                <Card>
                  <CardHeader>
                    <Text fontWeight="semibold">Audit Trail</Text>
                  </CardHeader>
                  <CardBody>
                    <VStack spacing={4} align="stretch">
                      <HStack spacing={4}>
                        <FiUser />
                        <VStack align="start" spacing={0}>
                          <Text fontSize="sm" fontWeight="medium">Created by</Text>
                          <Text fontSize="sm" color="gray.600">User ID: {journal.created_by}</Text>
                          <Text fontSize="xs" color="gray.500">
                            {new Date(journal.created_at).toLocaleString()}
                          </Text>
                        </VStack>
                      </HStack>

                      {journal.posted_by && (
                        <>
                          <Divider />
                          <HStack spacing={4}>
                            <FiCheck />
                            <VStack align="start" spacing={0}>
                              <Text fontSize="sm" fontWeight="medium">Posted by</Text>
                              <Text fontSize="sm" color="gray.600">User ID: {journal.posted_by}</Text>
                              {journal.posted_at && (
                                <Text fontSize="xs" color="gray.500">
                                  {new Date(journal.posted_at).toLocaleString()}
                                </Text>
                              )}
                            </VStack>
                          </HStack>
                        </>
                      )}

                      {journal.reversed_by && (
                        <>
                          <Divider />
                          <HStack spacing={4}>
                            <FiX />
                            <VStack align="start" spacing={0}>
                              <Text fontSize="sm" fontWeight="medium">Reversed by</Text>
                              <Text fontSize="sm" color="gray.600">User ID: {journal.reversed_by}</Text>
                              {journal.reversed_at && (
                                <Text fontSize="xs" color="gray.500">
                                  {new Date(journal.reversed_at).toLocaleString()}
                                </Text>
                              )}
                            </VStack>
                          </HStack>
                        </>
                      )}

                      <Divider />
                      <HStack spacing={4}>
                        <FiClock />
                        <VStack align="start" spacing={0}>
                          <Text fontSize="sm" fontWeight="medium">Last updated</Text>
                          <Text fontSize="xs" color="gray.500">
                            {new Date(journal.updated_at).toLocaleString()}
                          </Text>
                        </VStack>
                      </HStack>
                    </VStack>
                  </CardBody>
                </Card>
              </TabPanel>
            </TabPanels>
          </Tabs>
        </ModalBody>

        <ModalFooter>
          <Button onClick={onClose}>Close</Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default JournalDrilldownModal;