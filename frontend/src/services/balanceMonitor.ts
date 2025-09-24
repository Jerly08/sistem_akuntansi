// Client-side WebSocket Balance Monitor
// Usage:
// const monitor = new BalanceMonitor(API_WS_BASE);
// const stop = monitor.start([1,2,3], (event) => { /* on balance refresh */ });
// stop();

export type BalanceEvent = {
  type: 'BALANCE_REFRESHED' | 'CONNECTED';
  accounts?: number[];
  updated_at: string;
};

export class BalanceMonitor {
  private ws: WebSocket | null = null;
  private baseUrl: string;

  constructor(baseUrl: string = '') {
    this.baseUrl = baseUrl; // e.g., ws://localhost:8080 or empty to use relative
  }

  start(accountIds: number[] = [], onEvent?: (evt: BalanceEvent) => void) {
    const params = new URLSearchParams();
    if (accountIds.length > 0) params.set('accounts', accountIds.join(','));

    const wsUrl = this.buildWsUrl(`/api/v1/journals/account-balances/ws${params.toString() ? '?' + params.toString() : ''}`);
    this.ws = new WebSocket(wsUrl);

    const handleMessage = (e: MessageEvent) => {
      try {
        const data: BalanceEvent = JSON.parse(e.data);
        onEvent?.(data);
      } catch {}
    };

    const handleOpen = () => {
      // no-op for now
    };

    const handleClose = () => {
      // no-op: caller may restart
    };

    this.ws.addEventListener('message', handleMessage);
    this.ws.addEventListener('open', handleOpen);
    this.ws.addEventListener('close', handleClose);

    // return stop function
    return () => {
      if (this.ws) {
        this.ws.removeEventListener('message', handleMessage);
        this.ws.removeEventListener('open', handleOpen);
        this.ws.removeEventListener('close', handleClose);
        this.ws.close();
        this.ws = null;
      }
    };
  }

  private buildWsUrl(path: string) {
    if (this.baseUrl) return this.ensureWsBase(this.baseUrl) + path;
    const { protocol, host } = window.location;
    const wsProto = protocol === 'https:' ? 'wss:' : 'ws:';
    return `${wsProto}//${host}${path}`;
  }

  private ensureWsBase(url: string) {
    if (url.startsWith('http')) return url.replace(/^http/, 'ws');
    if (url.startsWith('ws')) return url;
    return 'ws://' + url;
  }
}