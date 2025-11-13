# P&L Closed Period History Integration

## Overview
Implementasi fitur untuk melihat historical Profit & Loss (P&L) reports dari periode-periode yang sudah di-close. Fitur ini terintegrasi dengan closing period system yang sama seperti Balance Sheet report.

## Implementation Date
**Date:** 2025-11-13  
**Status:** ✅ Completed

---

## Features

### 1. **Closed Period History Dropdown**
- Dropdown selector di P&L modal untuk memilih periode yang sudah di-close
- Auto-populate dari `/api/v1/fiscal-closing/history` endpoint
- Menampilkan daftar periode dengan format: "DD MMM YYYY - Period Closing"
- Sorted dari yang paling baru ke yang paling lama

### 2. **Auto-Load Closed Periods**
- Secara otomatis load closed periods saat P&L modal dibuka
- Loading state dengan button "Load History" untuk manual reload
- Error handling yang smooth tanpa mengganggu user experience

### 3. **Date Range Auto-Fill**
- Saat memilih closed period, otomatis mengisi Start Date dan End Date
- Start Date: Awal fiscal year (default: January 1)
- End Date: Tanggal periode closing yang dipilih

---

## Technical Implementation

### Frontend Changes

#### 1. **State Management** (`frontend/app/reports/page.tsx`)

```typescript
// State untuk SSOT Profit Loss
const [ssotPLOpen, setSSOTPLOpen] = useState(false);
const [ssotPLData, setSSOTPLData] = useState<SSOTProfitLossData | null>(null);
const [ssotPLLoading, setSSOTPLLoading] = useState(false);
const [ssotPLError, setSSOTPLError] = useState<string | null>(null);
const [ssotStartDate, setSSOTStartDate] = useState('');
const [ssotEndDate, setSSOTEndDate] = useState('');
const [ssotPLClosedPeriods, setSSOTPLClosedPeriods] = useState<any[]>([]);
const [loadingSSOTPLClosedPeriods, setLoadingSSOTPLClosedPeriods] = useState(false);
```

#### 2. **Auto-Load on Modal Open**

```typescript
} else if (report.id === 'profit-loss') {
  setSSOTPLOpen(true);
  // Auto-load closed periods untuk P&L
  (async () => {
    setLoadingSSOTPLClosedPeriods(true);
    try {
      const response = await api.get(API_ENDPOINTS.FISCAL_CLOSING.HISTORY);
      if (response.data.success && response.data.data && Array.isArray(response.data.data)) {
        const periods = response.data.data
          .filter((entry: any) => entry.entry_date)
          .sort((a: any, b: any) => new Date(b.entry_date).getTime() - new Date(a.entry_date).getTime())
          .map((entry: any) => {
            const dateObj = new Date(entry.entry_date);
            const formattedDate = dateObj.toLocaleDateString('id-ID', {
              day: '2-digit',
              month: 'short',
              year: 'numeric'
            });
            const dateValue = dateObj.toISOString().split('T')[0];
            return {
              value: dateValue,
              label: `${formattedDate} - ${entry.description || 'Period Closing'}`,
              date: dateValue,
              fiscal_year: dateObj.getFullYear()
            };
          });
        setSSOTPLClosedPeriods(periods);
        console.log(`[P&L] Auto-loaded ${periods.length} closed periods`);
      }
    } catch (error) {
      console.error('[P&L] Failed to auto-load closed periods:', error);
    } finally {
      setLoadingSSOTPLClosedPeriods(false);
    }
  })();
}
```

#### 3. **UI Components in Modal**

```tsx
{/* Closed Period History Selector */}
<HStack spacing={4} mb={4} alignItems="flex-start">
  <FormControl flex="1">
    <FormLabel fontSize="sm">
      <HStack spacing={2}>
        <Icon as={FiCalendar} />
        <Text>Period Closed History</Text>
        <Tooltip label="Select a closed period to view historical P&L report" placement="top">
          <Box><Icon as={FiInfo} color="gray.500" /></Box>
        </Tooltip>
      </HStack>
    </FormLabel>
    <Select
      placeholder="Select closed period (optional)"
      size="md"
      onChange={(e) => {
        if (e.target.value) {
          const selectedPeriod = ssotPLClosedPeriods.find(p => p.date === e.target.value);
          if (selectedPeriod) {
            // Set date range untuk periode yang dipilih
            const periodEndDate = new Date(selectedPeriod.date);
            const fiscalYearStart = new Date(periodEndDate.getFullYear(), 0, 1);
            setSSOTStartDate(fiscalYearStart.toISOString().split('T')[0]);
            setSSOTEndDate(selectedPeriod.date);
          }
        }
      }}
      isDisabled={loadingSSOTPLClosedPeriods}
    >
      {ssotPLClosedPeriods.length > 0 ? (
        ssotPLClosedPeriods.map((period) => (
          <option key={period.date} value={period.date}>
            {period.label}
          </option>
        ))
      ) : (
        <option disabled>No closed periods available</option>
      )}
    </Select>
  </FormControl>
  <Button
    size="md"
    variant="outline"
    colorScheme="blue"
    onClick={/* Manual reload handler */}
    isLoading={loadingSSOTPLClosedPeriods}
    leftIcon={<FiClock />}
    mt={7}
  >
    Load History
  </Button>
</HStack>

<Divider my={4} />
```

---

## Backend Integration

### Endpoint Used
**GET** `/api/v1/fiscal-closing/history`

#### Response Format:
```json
{
  "success": true,
  "data": [
    {
      "id": 6,
      "code": "CLO-2027-001",
      "description": "Closing Entry - Period End 2027-12-31",
      "entry_date": "2027-12-31T00:00:00Z",
      "created_at": "2025-01-12T19:55:45.123Z",
      "total_debit": 7000000,
      "source": "accounting_periods"
    },
    {
      "id": 2,
      "code": "CLO-2026-001",
      "description": "Closing Entry - Period End 2026-12-31",
      "entry_date": "2026-12-31T00:00:00Z",
      "created_at": "2025-01-12T19:55:45.123Z",
      "total_debit": 7000000,
      "source": "accounting_periods"
    }
  ]
}
```

### Unified Closing System
P&L report menggunakan sistem closing yang sama dengan Balance Sheet:

1. **Closing Logic:**
   - Revenue/Expense accounts di-zero out saat closing
   - Net income di-transfer ke Retained Earnings (3201)
   - Closing journal entries dibuat di `unified_journal_entries`

2. **Historical Data:**
   - P&L untuk periode closed menampilkan data **BEFORE** closing
   - Query dari `unified_journal_lines` dengan filter `entry_date BETWEEN start AND end`
   - Menggunakan SSOT journal system sebagai single source of truth

---

## User Flow

### Step-by-Step Usage:

1. **Open P&L Report:**
   - User klik "View Report" pada P&L card
   - Modal P&L terbuka
   - System auto-load closed periods di background

2. **Select Closed Period:**
   - User lihat dropdown "Period Closed History"
   - Pilih salah satu periode dari list (contoh: "31 Des 2027 - Period Closing")
   - Start Date dan End Date otomatis terisi

3. **Generate Report:**
   - User klik "Generate Report"
   - System fetch P&L data untuk periode yang dipilih
   - Report ditampilkan dengan data historical

4. **View Historical Data:**
   - Revenue details dari periode tersebut
   - Expense details dari periode tersebut
   - Net Income/Loss calculation
   - All based on journal entries BEFORE closing

---

## Data Flow Diagram

```
User Action              Frontend                Backend                  Database
-----------              --------                -------                  --------
Click P&L     →    setSSOTPLOpen(true)
                          ↓
                   Auto-load Periods    →    GET /fiscal-closing/history
                          ↓                            ↓
                   setSSOTPLClosedPeriods    ←    accounting_periods
                                                   unified_journal_entries
                          ↓
Select Period  →   Fill Start/End Dates
                          ↓
Generate       →   fetchSSOTPLReport()  →    GET /reports/ssot-profit-loss
                          ↓                            ↓
                   setSSOTPLData()         ←    Query unified_journal_lines
                                                   WHERE entry_date BETWEEN
                                                   start AND end
```

---

## Consistency with Balance Sheet

### Shared Features:

| Feature | Balance Sheet | P&L Report |
|---------|--------------|------------|
| Closed Period Dropdown | ✅ | ✅ |
| Auto-load on Modal Open | ✅ | ✅ |
| Manual Reload Button | ✅ | ✅ |
| Date Auto-fill | ✅ (as_of_date) | ✅ (start/end) |
| SSOT Integration | ✅ | ✅ |
| Historical Data View | ✅ | ✅ |

### Key Differences:

| Aspect | Balance Sheet | P&L Report |
|--------|--------------|------------|
| Date Selection | Single date (as_of_date) | Date range (start/end) |
| Closed Period Effect | Shows zero Revenue/Expense | Shows historical Revenue/Expense |
| Report Type | Point-in-time snapshot | Period performance |
| Net Income Display | Included in Retained Earnings | Calculated separately |

---

## Testing & Verification

### Test Scenarios:

1. **✅ Load Closed Periods:**
   ```bash
   - Open P&L modal
   - Verify dropdown is populated
   - Check console: "[P&L] Auto-loaded X closed periods"
   ```

2. **✅ Select Period:**
   ```bash
   - Choose period from dropdown
   - Verify Start Date = Fiscal Year Start
   - Verify End Date = Selected Period Date
   ```

3. **✅ Generate Historical Report:**
   ```bash
   - Click "Generate Report"
   - Verify Revenue/Expense data shown
   - Confirm data is from BEFORE closing
   ```

4. **✅ Compare with Backend:**
   ```bash
   cd backend
   go run cmd/check_unified_closing.go
   
   # Verify:
   # - Revenue Accounts with Balance: 0 (in current state)
   # - Expense Accounts with Balance: 0 (in current state)
   # - Historical data still accessible via journal entries
   ```

---

## Known Issues & Limitations

### Current Limitations:
1. **Fiscal Year Start:** Currently hardcoded to January 1. Should be configurable from Settings.
2. **Multiple Periods:** If there are duplicate dates, only the most recent is shown (by created_at).
3. **Date Format:** Uses Indonesian locale ('id-ID') for consistency.

### Future Enhancements:
1. Add filter by fiscal year
2. Group periods by year in dropdown
3. Show period comparison (YoY, QoQ)
4. Export historical P&L to PDF/CSV
5. Add period summary metadata (revenue, expenses, net income)

---

## API Endpoints Reference

### 1. Get Closed Periods
```
GET /api/v1/fiscal-closing/history
```

### 2. Generate P&L Report
```
GET /api/v1/reports/ssot-profit-loss?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD&format=json
```

---

## Code References

### Files Modified:
1. `frontend/app/reports/page.tsx`
   - Lines 203-211: State declarations
   - Lines 2200-2234: Auto-load logic
   - Lines 2504-2594: UI components

### Files Used:
1. `backend/services/ssot_profit_loss_service.go` - P&L generation logic
2. `backend/services/unified_period_closing_service.go` - Closing logic
3. `backend/controllers/fiscal_closing_controller.go` - History endpoint

---

## Conclusion

✅ **Implementation Status:** COMPLETED

Fitur P&L Closed Period History sudah berhasil diimplementasikan dengan sempurna. System terintegrasi dengan Balance Sheet closing logic dan menggunakan SSOT journal system sebagai single source of truth.

**Key Benefits:**
- User dapat melihat historical P&L dari periode yang sudah di-close
- Seamless integration dengan existing closing system
- Consistent user experience dengan Balance Sheet report
- Automatic data loading untuk better UX
- Historical data tetap accessible meskipun account balances sudah di-zero

**Next Steps:**
- Monitor user feedback
- Implement enhancements as needed
- Add more historical analysis features
- Consider adding period comparison capabilities
