'use client';

import React, { useState, useEffect, useCallback } from 'react';
import {
  Box,
  VStack,
  HStack,
  Button,
  Text,
  Badge,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Code,
  useToast,
  Divider,
} from '@chakra-ui/react';
import { FiPlay, FiStop, FiRefreshCw } from 'react-icons/fi';
import { BalanceMonitor, BalanceEvent } from '@/services/balanceMonitor';
import { useAuth } from '@/contexts/AuthContext';

// Demo component showing how to use real-time balance monitoring
export const BalanceMonitorDemo: React.FC = () => {
  const { token } = useAuth();
  const toast = useToast();
  const [monitor] = useState(() => new BalanceMonitor());
  const [isConnected, setIsConnected] = useState(false);
  const [events, setEvents] = useState<BalanceEvent[]>([]);
  const [stopMonitoring, setStopMonitoring] = useState<(() => void) | null>(null);

  const handleEvent = useCallback((event: BalanceEvent) => {
    console.log('Balance event received:', event);
    setEvents(prev => [event, ...prev.slice(0, 9)]); // Keep last 10 events
    
    if (event.type === 'CONNECTED') {
      setIsConnected(true);
      toast({
        title: 'WebSocket Connected',
        description: 'Real-time balance monitoring is active',
        status: 'success',
        duration: 3000,
      });
    } else if (event.type === 'BALANCE_REFRESHED') {
      toast({
        title: 'Balance Refreshed',
        description: `Account balances have been updated at ${new Date(event.updated_at).toLocaleTimeString()}`,
        status: 'info',
        duration: 2000,
      });
    }
  }, [toast]);

  const startMonitoring = useCallback(() => {
    if (!token) {
      toast({
        title: 'Authentication Required',
        description: 'Please login to use real-time monitoring',
        status: 'warning',
      });
      return;
    }

    const stopFn = monitor.start([1, 2, 3], handleEvent); // Monitor accounts 1, 2, 3
    setStopMonitoring(() => stopFn);
  }, [monitor, handleEvent, token, toast]);

  const stopMonitoringFn = useCallback(() => {
    if (stopMonitoring) {
      stopMonitoring();
      setStopMonitoring(null);
      setIsConnected(false);
      toast({
        title: 'Monitoring Stopped',
        description: 'Real-time balance monitoring disconnected',
        status: 'info',
      });
    }
  }, [stopMonitoring, toast]);

  const testRefresh = useCallback(async () => {
    if (!token) return;
    
    try {
      const response = await fetch('/api/v1/journals/account-balances/refresh', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      
      if (!response.ok) throw new Error('Refresh failed');
      
      toast({
        title: 'Manual Refresh Triggered',
        description: 'Account balances refresh has been triggered',
        status: 'success',
      });
    } catch (error) {
      console.error('Manual refresh error:', error);
      toast({
        title: 'Refresh Failed',
        description: 'Failed to trigger balance refresh',
        status: 'error',
      });
    }
  }, [token, toast]);

  useEffect(() => {
    return () => {
      if (stopMonitoring) {
        stopMonitoring();
      }
    };
  }, [stopMonitoring]);

  return (
    <Box p={4} maxW="800px" mx="auto">
      <VStack spacing={6} align="stretch">
        <Box>
          <Text fontSize="xl" fontWeight="bold" mb={2}>
            🔄 Real-time Balance Monitor Demo
          </Text>
          <Text color="gray.600">
            This demo shows WebSocket-based real-time monitoring of account balance updates.
          </Text>
        </Box>

        <HStack spacing={4}>
          <Button
            leftIcon={<FiPlay />}
            colorScheme="green"
            onClick={startMonitoring}
            isDisabled={!!stopMonitoring}
          >
            Start Monitoring
          </Button>
          
          <Button
            leftIcon={<FiStop />}
            colorScheme="red"
            onClick={stopMonitoringFn}
            isDisabled={!stopMonitoring}
          >
            Stop Monitoring
          </Button>
          
          <Button
            leftIcon={<FiRefreshCw />}
            colorScheme="blue"
            onClick={testRefresh}
            isDisabled={!token}
          >
            Test Manual Refresh
          </Button>
        </HStack>

        <Alert status={isConnected ? 'success' : 'warning'}>
          <AlertIcon />
          <AlertTitle>{isConnected ? 'Connected' : 'Disconnected'}</AlertTitle>
          <AlertDescription>
            WebSocket connection is {isConnected ? 'active' : 'inactive'}
          </AlertDescription>
        </Alert>

        <Divider />

        <Box>
          <Text fontSize="lg" fontWeight="semibold" mb={3}>
            📊 Recent Events ({events.length})
          </Text>
          
          {events.length === 0 ? (
            <Text color="gray.500" fontStyle="italic">
              No events yet. Start monitoring and trigger a balance refresh to see events.
            </Text>
          ) : (
            <VStack spacing={2} align="stretch">
              {events.map((event, index) => (
                <Box key={index} p={3} bg="gray.50" borderRadius="md">
                  <HStack justify="space-between" mb={2}>
                    <Badge
                      colorScheme={event.type === 'CONNECTED' ? 'green' : 'blue'}
                    >
                      {event.type}
                    </Badge>
                    <Text fontSize="xs" color="gray.600">
                      {new Date(event.updated_at).toLocaleString()}
                    </Text>
                  </HStack>
                  <Code fontSize="xs" display="block" p={2} bg="white">
                    {JSON.stringify(event, null, 2)}
                  </Code>
                </Box>
              ))}
            </VStack>
          )}
        </Box>

        <Alert status="info">
          <AlertIcon />
          <Box>
            <AlertTitle>How it works:</AlertTitle>
            <AlertDescription>
              <Text fontSize="sm" mt={2}>
                1. Click "Start Monitoring" to connect to the WebSocket<br />
                2. The system will send real-time events when account balances are refreshed<br />
                3. Click "Test Manual Refresh" to trigger a balance refresh and see the event<br />
                4. Events are also triggered when journal entries are posted
              </Text>
            </AlertDescription>
          </Box>
        </Alert>
      </VStack>
    </Box>
  );
};

export default BalanceMonitorDemo;