'use client';

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { BalanceWebSocketClient, BalanceUpdateData } from '@/services/balanceWebSocketService';
import { useAuth } from '@/contexts/AuthContext';
import { useToast } from '@chakra-ui/react';

interface WebSocketContextType {
  // Connection state
  isConnected: boolean;
  isConnecting: boolean;
  reconnectAttempts: number;
  
  // Balance monitoring
  balanceUpdates: Map<number, BalanceUpdateData>;
  lastUpdateTime: Date | null;
  updateCount: number;
  
  // Connection control
  connect: () => Promise<void>;
  disconnect: () => void;
  toggleConnection: () => void;
  
  // Data management
  clearBalanceUpdates: () => void;
  getAccountBalance: (accountId: number) => BalanceUpdateData | null;
  
  // Event listeners
  onBalanceUpdate: (callback: (data: BalanceUpdateData) => void) => () => void;
  
  // Configuration
  isPaused: boolean;
  togglePause: () => void;
  maxStoredUpdates: number;
  setMaxStoredUpdates: (max: number) => void;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined);

interface WebSocketProviderProps {
  children: React.ReactNode;
  autoConnect?: boolean;
  maxStoredUpdates?: number;
  enableNotifications?: boolean;
}

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({
  children,
  autoConnect = true,
  maxStoredUpdates = 50,
  enableNotifications = true
}) => {
  const { token, user } = useAuth();
  const toast = useToast();
  
  // Connection state
  const [isConnected, setIsConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [reconnectAttempts, setReconnectAttempts] = useState(0);
  const [isPaused, setIsPaused] = useState(false);
  
  // Balance monitoring state
  const [balanceUpdates, setBalanceUpdates] = useState<Map<number, BalanceUpdateData>>(new Map());
  const [lastUpdateTime, setLastUpdateTime] = useState<Date | null>(null);
  const [updateCount, setUpdateCount] = useState(0);
  const [maxUpdates, setMaxUpdates] = useState(maxStoredUpdates);
  
  // WebSocket client
  const [wsClient, setWsClient] = useState<BalanceWebSocketClient | null>(null);
  
  // Event listener callbacks
  const [balanceUpdateListeners, setBalanceUpdateListeners] = useState<((data: BalanceUpdateData) => void)[]>([]);

  // Initialize WebSocket client
  useEffect(() => {
    if (!wsClient) {
      const client = new BalanceWebSocketClient({
        debug: process.env.NODE_ENV === 'development',
        reconnectInterval: 3000,
        maxReconnectAttempts: 5
      });
      setWsClient(client);
    }
  }, []);

  // Auto-connect when conditions are met
  useEffect(() => {
    if (autoConnect && token && wsClient && !isPaused && !isConnected && !isConnecting) {
      connect();
    }
  }, [token, wsClient, isPaused, autoConnect]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (wsClient) {
        wsClient.disconnect();
      }
    };
  }, [wsClient]);

  // Handle balance updates
  const handleBalanceUpdate = useCallback((data: BalanceUpdateData) => {
    if (isPaused) return;

    // Update stored balance data
    setBalanceUpdates(prev => {
      const updated = new Map(prev);
      updated.set(data.account_id, {
        ...data,
        updated_at: new Date().toISOString()
      });
      
      // Limit stored updates
      if (updated.size > maxUpdates) {
        const sortedEntries = Array.from(updated.entries())
          .sort((a, b) => new Date(b[1].updated_at).getTime() - new Date(a[1].updated_at).getTime())
          .slice(0, maxUpdates);
        return new Map(sortedEntries);
      }
      
      return updated;
    });

    setLastUpdateTime(new Date());
    setUpdateCount(prev => prev + 1);

    // Notify listeners
    balanceUpdateListeners.forEach(listener => {
      try {
        listener(data);
      } catch (error) {
        console.error('Error in balance update listener:', error);
      }
    });

    // Show notifications for significant balance changes
    if (enableNotifications && Math.abs(data.balance) > 10000000) { // > 10M IDR
      toast({
        title: 'Significant Balance Update',
        description: `${data.account_code}: ${new Intl.NumberFormat('id-ID', {
          style: 'currency',
          currency: 'IDR',
          minimumFractionDigits: 0
        }).format(data.balance)}`,
        status: 'info',
        duration: 3000,
        isClosable: true,
        position: 'bottom-right',
        size: 'sm'
      });
    }
  }, [isPaused, maxUpdates, balanceUpdateListeners, enableNotifications, toast]);

  // Connection functions
  const connect = useCallback(async () => {
    if (!wsClient || !token || isConnecting || isConnected) return;

    try {
      setIsConnecting(true);
      setReconnectAttempts(0);

      // Set up event listeners
      wsClient.onBalanceUpdate(handleBalanceUpdate);
      
      wsClient.onConnect(() => {
        setIsConnected(true);
        setIsConnecting(false);
        setReconnectAttempts(0);
        
        if (enableNotifications) {
          toast({
            title: 'Real-time Updates Connected',
            description: 'Live balance monitoring is now active',
            status: 'success',
            duration: 3000,
            isClosable: true,
            position: 'bottom-right'
          });
        }
      });

      wsClient.onDisconnect(() => {
        setIsConnected(false);
        setIsConnecting(false);
      });

      await wsClient.connect(token);
    } catch (error) {
      setIsConnecting(false);
      setIsConnected(false);
      console.error('Failed to connect to WebSocket:', error);
      
      toast({
        title: 'Connection Failed',
        description: 'Unable to establish real-time connection',
        status: 'error',
        duration: 5000,
        isClosable: true
      });
    }
  }, [wsClient, token, isConnecting, isConnected, handleBalanceUpdate, enableNotifications, toast]);

  const disconnect = useCallback(() => {
    if (wsClient) {
      wsClient.disconnect();
      setIsConnected(false);
      setIsConnecting(false);
      setReconnectAttempts(0);
    }
  }, [wsClient]);

  const toggleConnection = useCallback(() => {
    if (isConnected) {
      disconnect();
    } else {
      connect();
    }
  }, [isConnected, connect, disconnect]);

  const togglePause = useCallback(() => {
    setIsPaused(prev => {
      const newPaused = !prev;
      if (newPaused && isConnected) {
        disconnect();
      } else if (!newPaused && token) {
        connect();
      }
      return newPaused;
    });
  }, [isConnected, token, connect, disconnect]);

  // Data management functions
  const clearBalanceUpdates = useCallback(() => {
    setBalanceUpdates(new Map());
    setUpdateCount(0);
    setLastUpdateTime(null);
  }, []);

  const getAccountBalance = useCallback((accountId: number): BalanceUpdateData | null => {
    return balanceUpdates.get(accountId) || null;
  }, [balanceUpdates]);

  // Event listener management
  const onBalanceUpdate = useCallback((callback: (data: BalanceUpdateData) => void) => {
    setBalanceUpdateListeners(prev => [...prev, callback]);
    
    // Return cleanup function
    return () => {
      setBalanceUpdateListeners(prev => prev.filter(listener => listener !== callback));
    };
  }, []);

  const setMaxStoredUpdates = useCallback((max: number) => {
    setMaxUpdates(max);
    
    // Trim existing updates if necessary
    setBalanceUpdates(prev => {
      if (prev.size <= max) return prev;
      
      const sortedEntries = Array.from(prev.entries())
        .sort((a, b) => new Date(b[1].updated_at).getTime() - new Date(a[1].updated_at).getTime())
        .slice(0, max);
      return new Map(sortedEntries);
    });
  }, []);

  const contextValue: WebSocketContextType = {
    // Connection state
    isConnected,
    isConnecting,
    reconnectAttempts,
    
    // Balance monitoring
    balanceUpdates,
    lastUpdateTime,
    updateCount,
    
    // Connection control
    connect,
    disconnect,
    toggleConnection,
    
    // Data management
    clearBalanceUpdates,
    getAccountBalance,
    
    // Event listeners
    onBalanceUpdate,
    
    // Configuration
    isPaused,
    togglePause,
    maxStoredUpdates: maxUpdates,
    setMaxStoredUpdates
  };

  return (
    <WebSocketContext.Provider value={contextValue}>
      {children}
    </WebSocketContext.Provider>
  );
};

// Hook to use WebSocket context
export const useWebSocket = (): WebSocketContextType => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
};

// Hook for balance monitoring
export const useBalanceMonitor = () => {
  const {
    balanceUpdates,
    lastUpdateTime,
    updateCount,
    isConnected,
    isConnecting,
    getAccountBalance,
    onBalanceUpdate
  } = useWebSocket();

  return {
    balanceUpdates: Array.from(balanceUpdates.values()).sort(
      (a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
    ),
    lastUpdateTime,
    updateCount,
    isConnected,
    isConnecting,
    getAccountBalance,
    onBalanceUpdate
  };
};

// Hook for connection control
export const useWebSocketConnection = () => {
  const {
    isConnected,
    isConnecting,
    reconnectAttempts,
    connect,
    disconnect,
    toggleConnection,
    isPaused,
    togglePause
  } = useWebSocket();

  return {
    isConnected,
    isConnecting,
    reconnectAttempts,
    connect,
    disconnect,
    toggleConnection,
    isPaused,
    togglePause
  };
};