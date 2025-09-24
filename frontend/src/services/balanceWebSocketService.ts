/**
 * WebSocket service for real-time balance monitoring
 * Connects to the backend WebSocket endpoint for balance updates
 */

export interface BalanceUpdateData {
  account_id: number;
  account_code: string;
  account_name: string;
  balance: number;
  balance_type: 'DEBIT' | 'CREDIT';
  updated_at: string;
  transaction_id?: number;
  journal_entry_id?: number;
}

export interface BalanceWebSocketOptions {
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  debug?: boolean;
}

export class BalanceWebSocketClient {
  private ws: WebSocket | null = null;
  private token: string | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectInterval = 3000;
  private debug = false;
  private isConnecting = false;
  private listeners: ((data: BalanceUpdateData) => void)[] = [];
  private connectionListeners: (() => void)[] = [];
  private disconnectionListeners: (() => void)[] = [];

  constructor(options?: BalanceWebSocketOptions) {
    this.maxReconnectAttempts = options?.maxReconnectAttempts ?? 5;
    this.reconnectInterval = options?.reconnectInterval ?? 3000;
    this.debug = options?.debug ?? false;
  }

  /**
   * Connect to the WebSocket server
   */
  async connect(token: string): Promise<void> {
    if (this.isConnecting || this.isConnected()) {
      return;
    }

    this.token = token;
    this.isConnecting = true;

    return new Promise((resolve, reject) => {
      try {
        // Use appropriate protocol based on current page protocol
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.hostname}:8080/ws/balance?token=${encodeURIComponent(token)}`;
        
        if (this.debug) {
          console.log('[BalanceWS] Connecting to:', wsUrl);
          console.log('[BalanceWS] Token length:', token?.length || 0);
        }

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
          this.isConnecting = false;
          this.reconnectAttempts = 0;
          
          if (this.debug) {
            console.log('[BalanceWS] Connected successfully');
          }

          // Notify connection listeners
          this.connectionListeners.forEach(listener => listener());
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const data: BalanceUpdateData = JSON.parse(event.data);
            
            if (this.debug) {
              console.log('[BalanceWS] Received update:', data);
            }

            // Notify all listeners
            this.listeners.forEach(listener => listener(data));
          } catch (error) {
            console.error('[BalanceWS] Error parsing message:', error);
          }
        };

        this.ws.onclose = (event) => {
          this.isConnecting = false;
          
          if (this.debug) {
            console.log('[BalanceWS] Connection closed:', {
              code: event.code,
              reason: event.reason,
              wasClean: event.wasClean,
              timestamp: new Date().toISOString()
            });
            
            // Log specific close codes for debugging
            switch(event.code) {
              case 1006:
                console.warn('[BalanceWS] Abnormal closure - possible server issue or network problem');
                break;
              case 1000:
                console.log('[BalanceWS] Normal closure');
                break;
              case 1001:
                console.warn('[BalanceWS] Going away');
                break;
              case 1002:
                console.error('[BalanceWS] Protocol error');
                break;
              case 1003:
                console.error('[BalanceWS] Unsupported data');
                break;
              case 1004:
                console.error('[BalanceWS] Reserved');
                break;
              case 1005:
                console.log('[BalanceWS] No status received');
                break;
              case 1011:
                console.error('[BalanceWS] Internal server error');
                break;
              default:
                console.warn(`[BalanceWS] Unexpected close code: ${event.code}`);
            }
          }

          // Notify disconnection listeners
          this.disconnectionListeners.forEach(listener => listener());

          // Auto-reconnect logic with exponential backoff
          if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts - 1), 10000);
            
            if (this.debug) {
              console.log(`[BalanceWS] Reconnecting... (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts}) in ${delay}ms`);
            }

            setTimeout(() => {
              if (this.token) {
                this.connect(this.token).catch(error => {
                  console.error('[BalanceWS] Reconnection failed:', error);
                });
              }
            }, delay);
          } else if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.warn('[BalanceWS] Max reconnection attempts reached');
          }
        };

        this.ws.onerror = (error) => {
          this.isConnecting = false;
          console.error('[BalanceWS] WebSocket error:', {
            error,
            readyState: this.ws?.readyState,
            url: wsUrl,
            timestamp: new Date().toISOString(),
            tokenProvided: !!token,
            tokenLength: token?.length || 0
          });
          
          // Provide more specific error context
          let errorMessage = 'WebSocket connection failed';
          if (this.ws?.readyState === WebSocket.CONNECTING) {
            errorMessage = 'WebSocket connection failed - server may be unavailable or authentication failed';
          } else if (this.ws?.readyState === WebSocket.OPEN) {
            errorMessage = 'WebSocket error on open connection - connection may be unstable';
          } else if (this.ws?.readyState === WebSocket.CLOSED) {
            errorMessage = 'WebSocket connection closed unexpectedly - check server status';
          }
          
          reject(new Error(errorMessage));
        };

      } catch (error) {
        this.isConnecting = false;
        reject(error);
      }
    });
  }

  /**
   * Disconnect from the WebSocket server
   */
  disconnect(): void {
    if (this.ws) {
      this.ws.close(1000, 'Manual disconnect');
      this.ws = null;
    }
    this.token = null;
    this.reconnectAttempts = 0;
    this.isConnecting = false;

    if (this.debug) {
      console.log('[BalanceWS] Disconnected');
    }
  }

  /**
   * Check if the WebSocket is connected
   */
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }

  /**
   * Register a callback for balance updates
   */
  onBalanceUpdate(callback: (data: BalanceUpdateData) => void): void {
    this.listeners.push(callback);
  }

  /**
   * Remove a balance update callback
   */
  removeBalanceUpdateListener(callback: (data: BalanceUpdateData) => void): void {
    const index = this.listeners.indexOf(callback);
    if (index > -1) {
      this.listeners.splice(index, 1);
    }
  }

  /**
   * Register a callback for connection events
   */
  onConnect(callback: () => void): void {
    this.connectionListeners.push(callback);
  }

  /**
   * Register a callback for disconnection events
   */
  onDisconnect(callback: () => void): void {
    this.disconnectionListeners.push(callback);
  }

  /**
   * Send a ping to keep connection alive (optional)
   */
  ping(): void {
    if (this.isConnected()) {
      this.ws?.send(JSON.stringify({ type: 'ping' }));
    }
  }

  /**
   * Get connection status information
   */
  getStatus() {
    return {
      connected: this.isConnected(),
      connecting: this.isConnecting,
      reconnectAttempts: this.reconnectAttempts,
      maxReconnectAttempts: this.maxReconnectAttempts,
      hasToken: !!this.token,
      listenerCount: this.listeners.length
    };
  }
}

/**
 * Default singleton instance
 */
let defaultClient: BalanceWebSocketClient | null = null;

/**
 * Get the default WebSocket client instance
 */
export function getDefaultBalanceClient(): BalanceWebSocketClient {
  if (!defaultClient) {
    defaultClient = new BalanceWebSocketClient({ debug: process.env.NODE_ENV === 'development' });
  }
  return defaultClient;
}

/**
 * Utility hook for React components (if using React hooks)
 * Note: React import needed when using this hook
 */
export function useBalanceWebSocket(token: string | null) {
  // This would need React import in the component that uses this hook
  // const client = getDefaultBalanceClient();
  // React.useEffect(() => { ... }, [token, client]);
  // return client;
  
  // For now, return static client
  return getDefaultBalanceClient();
}

export default BalanceWebSocketClient;