'use client';

import React, { useState, useEffect, useRef } from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  Card,
  CardBody,
  CardHeader,
  Badge,
  Button,
  IconButton,
  Tooltip,
  useToast,
  Grid,
  GridItem,
  Spinner,
  Alert,
  AlertIcon,
  useColorModeValue,
  Flex,
  Stat,
  StatLabel,
  StatNumber,
  StatHelpText,
  Divider,
} from '@chakra-ui/react';
import {
  FiActivity,
  FiWifi,
  FiWifiOff,
  FiRefreshCw,
  FiPause,
  FiPlay,
  FiX,
} from 'react-icons/fi';
import { BalanceWebSocketClient, BalanceUpdateData } from '@/services/balanceWebSocketService';
import { useAuth } from '@/contexts/AuthContext';
import { formatCurrency } from '@/utils/formatters';

interface BalanceMonitorProps {
  autoConnect?: boolean;
  showControls?: boolean;
  maxDisplayItems?: number;
  showConnectionStatus?: boolean;
}

interface DisplayedBalance {
  account_id: number;
  account_code: string;
  account_name: string;
  balance: number;
  balance_type: 'DEBIT' | 'CREDIT';
  last_updated: Date;
}

const BalanceMonitor: React.FC<BalanceMonitorProps> = ({
  autoConnect = true,
  showControls = true,
  maxDisplayItems = 8,
  showConnectionStatus = true
}) => {
  const { token, user } = useAuth();
  const toast = useToast();
  const [isConnected, setIsConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [balances, setBalances] = useState<Map<number, DisplayedBalance>>(new Map());
  const [lastUpdateTime, setLastUpdateTime] = useState<Date | null>(null);
  const [updateCount, setUpdateCount] = useState(0);
  const clientRef = useRef<BalanceWebSocketClient | null>(null);

  // Color mode values
  const cardBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const headerBg = useColorModeValue('gray.50', 'gray.700');
  const textColor = useColorModeValue('gray.800', 'white');
  const mutedColor = useColorModeValue('gray.500', 'gray.400');

  // Initialize WebSocket connection
  useEffect(() => {
    if (autoConnect && token && !isPaused) {
      connectToService();
    }

    return () => {
      disconnectFromService();
    };
  }, [token, autoConnect, isPaused]);

  const connectToService = async () => {
    if (!token || clientRef.current?.isConnected()) return;

    try {
      setIsConnecting(true);
      
      if (!clientRef.current) {
        clientRef.current = new BalanceWebSocketClient({
          debug: process.env.NODE_ENV === 'development',
          reconnectInterval: 3000,
          maxReconnectAttempts: 5
        });
      }

      // Set up event listeners
      clientRef.current.onBalanceUpdate(handleBalanceUpdate);
      clientRef.current.onConnect(() => {
        setIsConnected(true);
        setIsConnecting(false);
        toast({
          title: 'Balance Monitor Connected',
          description: 'Real-time balance updates are now active',
          status: 'success',
          duration: 3000,
          isClosable: true,
          position: 'bottom-right',
        });
      });

      clientRef.current.onDisconnect(() => {
        setIsConnected(false);
        setIsConnecting(false);
      });

      // Connect
      await clientRef.current.connect(token);
    } catch (error) {
      setIsConnecting(false);
      setIsConnected(false);
      console.error('Failed to connect to balance monitor:', error);
      
      toast({
        title: 'Connection Failed',
        description: 'Unable to connect to real-time balance updates',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  const disconnectFromService = () => {
    if (clientRef.current) {
      clientRef.current.disconnect();
      clientRef.current = null;
    }
    setIsConnected(false);
    setIsConnecting(false);
  };

  const handleBalanceUpdate = (data: BalanceUpdateData) => {
    if (isPaused) return;

    const newBalance: DisplayedBalance = {
      account_id: data.account_id,
      account_code: data.account_code,
      account_name: data.account_name,
      balance: data.balance,
      balance_type: data.balance_type,
      last_updated: new Date()
    };

    setBalances(prev => {
      const updated = new Map(prev);
      updated.set(data.account_id, newBalance);
      
      // Keep only the most recent items if we exceed maxDisplayItems
      if (updated.size > maxDisplayItems) {
        const sortedEntries = Array.from(updated.entries())
          .sort((a, b) => b[1].last_updated.getTime() - a[1].last_updated.getTime())
          .slice(0, maxDisplayItems);
        
        return new Map(sortedEntries);
      }
      
      return updated;
    });

    setLastUpdateTime(new Date());
    setUpdateCount(prev => prev + 1);

    // Show brief notification for significant balance changes
    if (Math.abs(data.balance) > 1000000) { // > 1M IDR
      toast({
        title: 'Significant Balance Update',
        description: `${data.account_code}: ${formatCurrency(data.balance)}`,
        status: 'info',
        duration: 2000,
        isClosable: true,
        position: 'bottom-right',
        size: 'sm'
      });
    }
  };

  const toggleConnection = () => {
    if (isConnected) {
      disconnectFromService();
    } else {
      connectToService();
    }
  };

  const togglePause = () => {
    setIsPaused(!isPaused);
    if (!isPaused && isConnected) {
      disconnectFromService();
    } else if (isPaused && token) {
      connectToService();
    }
  };

  const clearBalances = () => {
    setBalances(new Map());
    setUpdateCount(0);
    setLastUpdateTime(null);
  };

  const getConnectionStatus = () => {
    if (isConnecting) return { text: 'Connecting...', color: 'yellow', icon: Spinner };
    if (isConnected && !isPaused) return { text: 'Live', color: 'green', icon: FiWifi };
    if (isPaused) return { text: 'Paused', color: 'orange', icon: FiPause };
    return { text: 'Disconnected', color: 'red', icon: FiWifiOff };
  };

  const status = getConnectionStatus();
  const balanceArray = Array.from(balances.values()).sort(
    (a, b) => b.last_updated.getTime() - a.last_updated.getTime()
  );

  return (
    <Card bg={cardBg} border="1px" borderColor={borderColor} shadow="sm">
      <CardHeader pb={2}>
        <Flex justify="space-between" align="center">
          <HStack spacing={3}>
            <FiActivity size="20px" color="blue.500" />
            <VStack align="start" spacing={0}>
              <Text fontSize="md" fontWeight="semibold" color={textColor}>
                Real-time Balance Monitor
              </Text>
              {showConnectionStatus && (
                <HStack spacing={2}>
                  <Badge
                    colorScheme={status.color}
                    variant="subtle"
                    display="flex"
                    alignItems="center"
                    gap={1}
                    fontSize="xs"
                  >
                    {status.icon === Spinner ? (
                      <Spinner size="xs" />
                    ) : (
                      <status.icon size={10} />
                    )}
                    {status.text}
                  </Badge>
                  {lastUpdateTime && (
                    <Text fontSize="xs" color={mutedColor}>
                      Last update: {lastUpdateTime.toLocaleTimeString()}
                    </Text>
                  )}
                </HStack>
              )}
            </VStack>
          </HStack>

          {showControls && (
            <HStack spacing={2}>
              <Tooltip label={isPaused ? 'Resume updates' : 'Pause updates'}>
                <IconButton
                  aria-label={isPaused ? 'Resume' : 'Pause'}
                  icon={isPaused ? <FiPlay /> : <FiPause />}
                  size="sm"
                  variant="ghost"
                  onClick={togglePause}
                />
              </Tooltip>

              <Tooltip label={isConnected ? 'Disconnect' : 'Connect'}>
                <IconButton
                  aria-label={isConnected ? 'Disconnect' : 'Connect'}
                  icon={isConnected ? <FiWifiOff /> : <FiWifi />}
                  size="sm"
                  variant="ghost"
                  onClick={toggleConnection}
                  isLoading={isConnecting}
                />
              </Tooltip>

              <Tooltip label="Clear display">
                <IconButton
                  aria-label="Clear"
                  icon={<FiX />}
                  size="sm"
                  variant="ghost"
                  onClick={clearBalances}
                />
              </Tooltip>

              <Tooltip label="Refresh connection">
                <IconButton
                  aria-label="Refresh"
                  icon={<FiRefreshCw />}
                  size="sm"
                  variant="ghost"
                  onClick={() => {
                    disconnectFromService();
                    setTimeout(() => connectToService(), 500);
                  }}
                  isDisabled={isConnecting}
                />
              </Tooltip>
            </HStack>
          )}
        </Flex>
      </CardHeader>

      <CardBody pt={0}>
        {!isConnected && !isConnecting && (
          <Alert status="info" size="sm" borderRadius="md">
            <AlertIcon />
            <Text fontSize="sm">
              {token 
                ? 'Not connected to real-time updates. Click the connect button to enable live monitoring.'
                : 'Please log in to enable real-time balance monitoring.'
              }
            </Text>
          </Alert>
        )}

        {balanceArray.length === 0 && isConnected && (
          <Alert status="info" size="sm" borderRadius="md">
            <AlertIcon />
            <Text fontSize="sm">
              Connected and waiting for balance updates. Balances will appear here when transactions occur.
            </Text>
          </Alert>
        )}

        {balanceArray.length > 0 && (
          <>
            <HStack justify="space-between" mb={3}>
              <Text fontSize="sm" fontWeight="medium" color={textColor}>
                Recent Balance Updates ({balanceArray.length})
              </Text>
              <Text fontSize="xs" color={mutedColor}>
                {updateCount} updates received
              </Text>
            </HStack>

            <Grid templateColumns="repeat(auto-fit, minmax(280px, 1fr))" gap={3}>
              {balanceArray.map((balance) => (
                <GridItem key={balance.account_id}>
                  <Box
                    p={3}
                    borderRadius="md"
                    border="1px"
                    borderColor={borderColor}
                    bg={headerBg}
                    transition="all 0.2s"
                    _hover={{ shadow: 'sm' }}
                  >
                    <VStack align="stretch" spacing={2}>
                      <HStack justify="space-between" align="start">
                        <VStack align="start" spacing={0} flex={1} minW={0}>
                          <Text fontSize="sm" fontWeight="semibold" color={textColor} isTruncated>
                            {balance.account_code}
                          </Text>
                          <Text fontSize="xs" color={mutedColor} isTruncated>
                            {balance.account_name}
                          </Text>
                        </VStack>
                        <Badge
                          colorScheme={balance.balance_type === 'DEBIT' ? 'blue' : 'green'}
                          variant="subtle"
                          fontSize="xs"
                        >
                          {balance.balance_type}
                        </Badge>
                      </HStack>

                      <Divider />

                      <HStack justify="space-between">
                        <Text
                          fontSize="md"
                          fontWeight="bold"
                          color={balance.balance >= 0 ? 'green.600' : 'red.600'}
                          fontFamily="mono"
                        >
                          {formatCurrency(balance.balance)}
                        </Text>
                        <Text fontSize="xs" color={mutedColor}>
                          {balance.last_updated.toLocaleTimeString()}
                        </Text>
                      </HStack>
                    </VStack>
                  </Box>
                </GridItem>
              ))}
            </Grid>
          </>
        )}
      </CardBody>
    </Card>
  );
};

export default BalanceMonitor;