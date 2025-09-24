# CashBank Integration Frontend Components

Komponen-komponen React untuk integrasi CashBank dengan SSOT Journal System.

## Overview

Frontend components ini menyediakan interface untuk menampilkan dan mengelola data integrasi antara sistem CashBank dan SSOT Journal. Komponen-komponen ini dirancang dengan TypeScript dan menggunakan Tailwind CSS untuk styling.

## Structure

```
frontend_components/
├── components/
│   ├── CashBankIntegratedDashboard.tsx    # Dashboard utama
│   └── CashBankAccountDetail.tsx          # Detail akun dengan tab views
├── pages/
│   └── CashBankIntegrationPage.tsx        # Main page wrapper
├── types/
│   └── cashBankIntegration.types.ts       # TypeScript interfaces
├── services/
│   └── cashBankIntegrationService.ts      # API service layer
├── hooks/
│   └── useCashBankIntegration.ts          # Custom React hook
└── README.md
```

## Components

### 1. CashBankIntegratedDashboard

Dashboard utama yang menampilkan:
- **Summary Cards**: Total cash, bank, balance, dan SSOT balance
- **Variance Alerts**: Peringatan jika ada selisih balance
- **Sync Status**: Status sinkronisasi dan statistik
- **Accounts List**: Daftar akun dengan status reconciliation
- **Recent Activities**: Aktivitas transaksi terbaru

#### Props
```typescript
interface CashBankIntegratedDashboardProps {
  onAccountSelect?: (accountId: number) => void;
}
```

#### Features
- Auto-refresh setiap 5 menit
- Real-time variance detection
- Interactive account selection
- Loading states dan error handling

### 2. CashBankAccountDetail

Komponen detail akun dengan tab-based interface:

#### Tabs
- **Overview**: Informasi akun, transaksi terbaru, jurnal entries terbaru
- **Transactions**: Riwayat transaksi dengan pagination
- **Journal**: Jurnal entries dengan pagination  
- **Reconciliation**: Status rekonsiliasi dan detail selisih

#### Props
```typescript
interface CashBankAccountDetailProps {
  accountId: number;
  onBack?: () => void;
}
```

#### Features
- Lazy loading untuk tab content
- Pagination untuk data yang besar
- Balance comparison (CashBank vs SSOT)
- Reconciliation status tracking

### 3. CashBankIntegrationPage

Main page component yang menggabungkan dashboard dan detail views.

#### Features
- Navigation breadcrumbs
- View state management
- Responsive layout
- Footer dengan status information

## Services

### CashBankIntegrationService

Service layer untuk komunikasi dengan backend API:

#### Methods
- `getIntegratedSummary()`: Mengambil summary data
- `getIntegratedAccount(id)`: Detail akun terintegrasi
- `getReconciliation(id)`: Data rekonsiliasi
- `getJournalEntries(id, page, size)`: Jurnal entries dengan pagination
- `getTransactionHistory(id, page, size)`: Riwayat transaksi

#### Utilities
- `formatCurrency()`: Format mata uang
- `formatDate()`: Format tanggal
- `formatDateTime()`: Format tanggal dan waktu
- `hasVariance()`: Cek apakah ada selisih
- `getReconciliationStatusColor()`: Warna status
- `getAccountTypeIcon()`: Icon berdasarkan tipe akun

## Custom Hook

### useCashBankIntegration

React hook untuk state management dan caching:

#### Features
- **State Management**: Centralized state untuk summary dan account data
- **Caching**: Intelligent caching dengan TTL
- **Auto Refresh**: Configurable auto-refresh
- **Error Handling**: Comprehensive error states
- **Loading States**: Granular loading indicators

#### Usage
```typescript
const {
  state,
  loadSummary,
  loadAccount,
  refreshAll,
  hasVarianceAccounts
} = useCashBankIntegration({
  autoRefresh: true,
  refreshInterval: 300000, // 5 minutes
  enableCaching: true,
  cacheTimeout: 600000 // 10 minutes
});
```

## TypeScript Interfaces

### Core Types
- `IntegratedSummaryResponse`: Response summary data
- `IntegratedAccountResponse`: Response detail akun
- `ReconciliationData`: Data rekonsiliasi
- `JournalEntriesResponse`: Response jurnal entries
- `TransactionHistoryResponse`: Response riwayat transaksi

### Enums
- `ReconciliationStatus`: Status rekonsiliasi
- `AccountType`: Tipe akun (cash/bank)

## Installation & Setup

### Prerequisites
- React 18+
- TypeScript 4.5+
- Tailwind CSS 3.0+

### Integration Steps

1. **Copy Components**
   ```bash
   # Copy all files ke project React Anda
   cp -r frontend_components/* src/
   ```

2. **Install Dependencies**
   ```bash
   npm install
   # atau
   yarn install
   ```

3. **Configure API Endpoints**
   ```typescript
   // Ubah BASE_URL di cashBankIntegrationService.ts
   const BASE_URL = 'http://your-backend-url/api/cashbank';
   ```

4. **Setup Authentication**
   ```typescript
   // Sesuaikan getAuthToken() dengan auth system Anda
   const getAuthToken = (): string => {
     return localStorage.getItem('auth_token') || '';
   };
   ```

5. **Add to Your App**
   ```typescript
   import CashBankIntegrationPage from './pages/CashBankIntegrationPage';
   
   // Di router Anda
   <Route path="/cash-bank" component={CashBankIntegrationPage} />
   ```

## Usage Examples

### Basic Dashboard
```typescript
import CashBankIntegratedDashboard from './components/CashBankIntegratedDashboard';

function MyPage() {
  const handleAccountSelect = (accountId: number) => {
    // Handle account selection
    console.log('Selected account:', accountId);
  };

  return (
    <CashBankIntegratedDashboard onAccountSelect={handleAccountSelect} />
  );
}
```

### Account Detail with Hook
```typescript
import CashBankAccountDetail from './components/CashBankAccountDetail';
import useCashBankIntegration from './hooks/useCashBankIntegration';

function AccountPage({ accountId }: { accountId: number }) {
  const { loadAccount, getAccountFromCache } = useCashBankIntegration();
  
  useEffect(() => {
    loadAccount(accountId);
  }, [accountId]);

  return <CashBankAccountDetail accountId={accountId} />;
}
```

### Custom Hook Usage
```typescript
function MyDashboard() {
  const {
    state,
    refreshAll,
    hasVarianceAccounts,
    getAccountsWithVariance
  } = useCashBankIntegration({
    autoRefresh: true,
    refreshInterval: 60000 // 1 minute
  });

  const varianceAccounts = getAccountsWithVariance();

  return (
    <div>
      {hasVarianceAccounts() && (
        <div className="alert alert-warning">
          {varianceAccounts.length} akun memiliki selisih balance
        </div>
      )}
      {/* Rest of component */}
    </div>
  );
}
```

## Customization

### Styling
Komponen menggunakan Tailwind CSS classes. Untuk customization:

1. **Colors**: Ubah color palette di `tailwind.config.js`
2. **Spacing**: Sesuaikan spacing dengan design system Anda
3. **Typography**: Customize font dan text sizes

### API Integration
Sesuaikan service layer di `cashBankIntegrationService.ts`:

```typescript
// Custom API client
const apiClient = axios.create({
  baseURL: process.env.REACT_APP_API_URL,
  headers: {
    'Content-Type': 'application/json'
  }
});
```

### Internationalization
Untuk multi-bahasa, replace hard-coded strings:

```typescript
// Before
<h1>Cash & Bank Terintegrasi</h1>

// After
<h1>{t('cashbank.integrated.title')}</h1>
```

## Error Handling

### API Errors
```typescript
try {
  const data = await cashBankIntegrationService.getIntegratedSummary();
  // Handle success
} catch (error) {
  if (error.response?.status === 401) {
    // Handle authentication
  } else if (error.response?.status === 403) {
    // Handle authorization
  } else {
    // Handle other errors
  }
}
```

### Network Errors
Components automatically handle:
- Loading states
- Network timeouts
- Server errors
- Retry mechanisms

## Performance Optimization

### Caching Strategy
- Summary data: 10 menit TTL
- Account data: 10 menit TTL  
- Paginated data: No caching (real-time)

### Lazy Loading
- Tab content dimuat saat pertama kali diakses
- Images dan assets di-lazy load
- Pagination untuk data besar

### Memory Management
- Cleanup intervals saat component unmount
- Cache eviction untuk memory optimization
- AbortController untuk cancel requests

## Testing

### Unit Tests
```typescript
// Example test untuk service
import { cashBankIntegrationService } from './services/cashBankIntegrationService';

describe('CashBankIntegrationService', () => {
  test('should format currency correctly', () => {
    expect(cashBankIntegrationService.formatCurrency(1000000))
      .toBe('Rp 1.000.000');
  });
});
```

### Integration Tests
```typescript
// Example test untuk component
import { render, screen } from '@testing-library/react';
import CashBankIntegratedDashboard from './CashBankIntegratedDashboard';

test('renders dashboard with summary cards', async () => {
  render(<CashBankIntegratedDashboard />);
  
  expect(screen.getByText('Total Cash')).toBeInTheDocument();
  expect(screen.getByText('Total Bank')).toBeInTheDocument();
});
```

## Troubleshooting

### Common Issues

1. **CORS Errors**
   ```
   Solution: Configure backend CORS untuk frontend domain
   ```

2. **Authentication Issues**
   ```typescript
   // Check token validity
   const token = localStorage.getItem('auth_token');
   if (!token || isTokenExpired(token)) {
     // Redirect to login
   }
   ```

3. **Performance Issues**
   ```typescript
   // Reduce auto-refresh frequency
   const hook = useCashBankIntegration({
     refreshInterval: 600000 // 10 minutes
   });
   ```

4. **Memory Leaks**
   ```typescript
   // Ensure cleanup in useEffect
   useEffect(() => {
     const interval = setInterval(loadData, 60000);
     return () => clearInterval(interval);
   }, []);
   ```

## Contributing

1. Follow TypeScript strict mode
2. Add JSDoc comments untuk public methods
3. Write unit tests untuk new features
4. Update documentation
5. Use semantic commit messages

## License

MIT License - sesuai dengan project utama.

---

## Changelog

### v1.0.0
- Initial release dengan basic integration
- Dashboard dan detail components
- Custom hook dengan caching
- API service layer
- TypeScript interfaces
- Documentation

### Future Enhancements
- Real-time WebSocket updates
- Export data functionality  
- Advanced filtering dan search
- Mobile responsive improvements
- Dark mode support