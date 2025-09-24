# PSAK-Compliant Financial Reports Documentation

## Overview

Dokumentasi ini menjelaskan implementasi laporan keuangan yang mematuhi Pernyataan Standar Akuntansi Keuangan (PSAK) Indonesia dalam sistem akuntansi. Implementasi ini memastikan bahwa semua laporan keuangan yang dihasilkan sesuai dengan standar akuntansi yang berlaku di Indonesia.

## Table of Contents

1. [PSAK Standards Implemented](#psak-standards-implemented)
2. [Architecture Overview](#architecture-overview)
3. [Financial Reports](#financial-reports)
4. [API Endpoints](#api-endpoints)
5. [Data Structures](#data-structures)
6. [Implementation Details](#implementation-details)
7. [Compliance Checking](#compliance-checking)
8. [Usage Examples](#usage-examples)
9. [Configuration](#configuration)
10. [Testing](#testing)
11. [Future Enhancements](#future-enhancements)

## PSAK Standards Implemented

### 1. PSAK 1 - Penyajian Laporan Keuangan
- **Scope**: Balance Sheet dan Profit & Loss Statement
- **Key Requirements**:
  - Klasifikasi aset lancar dan tidak lancar
  - Klasifikasi liabilitas jangka pendek dan panjang
  - Penyajian ekuitas yang tepat
  - Format laporan laba rugi komprehensif
  - Pengungkapan yang memadai

### 2. PSAK 2 - Laporan Arus Kas
- **Scope**: Cash Flow Statement
- **Key Requirements**:
  - Klasifikasi aktivitas operasi, investasi, dan pendanaan
  - Metode langsung dan tidak langsung
  - Rekonsiliasi kas dan setara kas
  - Pengungkapan transaksi non-kas

### 3. PSAK 14 - Persediaan
- **Scope**: Inventory accounting dalam Balance Sheet dan P&L
- **Key Requirements**:
  - Penilaian persediaan
  - Beban pokok penjualan
  - Penurunan nilai persediaan

### 4. PSAK 16 - Aset Tetap
- **Scope**: Fixed assets dalam Balance Sheet
- **Key Requirements**:
  - Pengakuan dan pengukuran aset tetap
  - Depresiasi dan amortisasi
  - Penurunan nilai aset

### 5. PSAK 23 - Pendapatan dari Kontrak dengan Pelanggan
- **Scope**: Revenue recognition dalam P&L
- **Key Requirements**:
  - Lima langkah pengakuan pendapatan
  - Identifikasi kewajiban pelaksanaan
  - Alokasi harga transaksi

### 6. PSAK 46 - Pajak Penghasilan
- **Scope**: Tax accounting dalam Balance Sheet dan P&L
- **Key Requirements**:
  - Pajak kini dan tangguhan
  - Aset dan liabilitas pajak tangguhan
  - Beban pajak penghasilan

### 7. PSAK 56 - Laba per Saham
- **Scope**: Earnings per share dalam P&L
- **Key Requirements**:
  - Laba per saham dasar
  - Laba per saham dilusian
  - Penyajian dan pengungkapan

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    PSAK Compliance Layer                        │
├─────────────────────────────────────────────────────────────────┤
│  Controllers/                                                   │
│  ├── psak_compliant_report_controller.go                       │
│  │   ├── GetPSAKBalanceSheet()                                 │
│  │   ├── GetPSAKProfitLoss()                                   │
│  │   ├── GetPSAKCashFlow()                                     │
│  │   ├── GetPSAKComplianceSummary()                            │
│  │   └── CheckPSAKCompliance()                                 │
├─────────────────────────────────────────────────────────────────┤
│  Services/                                                      │
│  ├── psak_compliant_report_service.go                          │
│  │   ├── GeneratePSAKBalanceSheet()                            │
│  │   ├── GeneratePSAKProfitLoss()                              │
│  │   ├── GeneratePSAKCashFlow()                                │
│  │   └── PSAK Classification Methods                           │
├─────────────────────────────────────────────────────────────────┤
│  Data Layer (SSOT Integration)                                 │
│  ├── Unified Journal System                                    │
│  ├── Account Balances                                          │
│  └── Transaction Data                                          │
└─────────────────────────────────────────────────────────────────┘
```

## Financial Reports

### 1. PSAK Balance Sheet (Neraca)

#### Structure (sesuai PSAK 1):
- **ASET (Assets)**
  - Aset Lancar (Current Assets)
    - Kas dan Setara Kas
    - Piutang Usaha
    - Persediaan
    - Beban Dibayar Dimuka
  - Aset Tidak Lancar (Non-Current Assets)
    - Aset Tetap
    - Aset Tidak Berwujud
    - Properti Investasi
    - Investasi Jangka Panjang
    - Aset Pajak Tangguhan

- **LIABILITAS (Liabilities)**
  - Liabilitas Jangka Pendek (Current Liabilities)
    - Utang Usaha
    - Beban yang Masih Harus Dibayar
    - Pinjaman Jangka Pendek
    - Bagian Lancar Utang Jangka Panjang
    - Utang Pajak
  - Liabilitas Jangka Panjang (Non-Current Liabilities)
    - Pinjaman Jangka Panjang
    - Liabilitas Pajak Tangguhan
    - Imbalan Kerja
    - Provisi

- **EKUITAS (Equity)**
  - Modal Saham
  - Agio Saham
  - Saldo Laba
  - Komponen Ekuitas Lainnya
  - Saham Treasury

### 2. PSAK Profit & Loss (Laporan Laba Rugi)

#### Structure (sesuai PSAK 1):
- **PENDAPATAN (Revenue)**
  - Penjualan
  - Jasa
  - Pendapatan Bunga
  - Pendapatan Dividen
  - Pendapatan Royalti

- **BEBAN POKOK PENJUALAN (Cost of Sales)**
  - Sesuai PSAK 14 (Persediaan)

- **LABA KOTOR (Gross Profit)**
  - Pendapatan - Beban Pokok Penjualan

- **BEBAN USAHA (Operating Expenses)**
  - Beban Penjualan
  - Beban Umum dan Administrasi

- **LABA USAHA (Operating Profit)**
  - Laba Kotor - Beban Usaha

- **PENDAPATAN DAN BEBAN LAIN-LAIN**
  - Pendapatan Lain-lain
  - Beban Lain-lain

- **LABA SEBELUM PAJAK (Profit Before Tax)**

- **BEBAN PAJAK PENGHASILAN (sesuai PSAK 46)**
  - Beban Pajak Kini
  - Beban Pajak Tangguhan

- **LABA NETO (Net Profit)**

- **PENGHASILAN KOMPREHENSIF LAIN (OCI)**
  - Keuntungan/Kerugian Aktuarial
  - Surplus Revaluasi
  - Selisih Kurs
  - Lindung Nilai Arus Kas

- **LABA PER SAHAM (sesuai PSAK 56)**
  - Laba per Saham Dasar
  - Laba per Saham Dilusian

### 3. PSAK Cash Flow (Laporan Arus Kas)

#### Structure (sesuai PSAK 2):
- **ARUS KAS DARI AKTIVITAS OPERASI**
  - Metode Langsung (Direct Method)
    - Penerimaan kas dari pelanggan
    - Pembayaran kas kepada pemasok
    - Pembayaran kas untuk karyawan
    - Pembayaran bunga
    - Pembayaran pajak
  - Metode Tidak Langsung (Indirect Method)
    - Laba neto
    - Penyesuaian untuk item non-kas
    - Perubahan modal kerja

- **ARUS KAS DARI AKTIVITAS INVESTASI**
  - Pembelian/penjualan aset tetap
  - Investasi dalam sekuritas
  - Pinjaman yang diberikan

- **ARUS KAS DARI AKTIVITAS PENDANAAN**
  - Penerbitan/pelunasan saham
  - Pinjaman dari bank
  - Pembayaran dividen

- **PENGARUH PERUBAHAN KURS VALUTA ASING**

- **KENAIKAN/PENURUNAN KAS DAN SETARA KAS**

## API Endpoints

### Base URL: `/api/v1/reports/psak`

#### 1. Generate PSAK Balance Sheet
```http
POST /balance-sheet
Content-Type: application/json
Authorization: Bearer {token}

{
  "as_of_date": "2024-12-31"
}
```

#### 2. Generate PSAK Profit & Loss
```http
POST /profit-loss
Content-Type: application/json
Authorization: Bearer {token}

{
  "start_date": "2024-01-01",
  "end_date": "2024-12-31"
}
```

#### 3. Generate PSAK Cash Flow
```http
POST /cash-flow
Content-Type: application/json
Authorization: Bearer {token}

{
  "start_date": "2024-01-01",
  "end_date": "2024-12-31",
  "method": "INDIRECT"
}
```

#### 4. Get PSAK Compliance Summary
```http
GET /compliance-summary?as_of_date=2024-12-31&start_date=2024-01-01&end_date=2024-12-31
Authorization: Bearer {token}
```

#### 5. Check PSAK Compliance
```http
POST /check-compliance
Content-Type: application/json
Authorization: Bearer {token}

{
  "report_type": "BALANCE_SHEET",
  "as_of_date": "2024-12-31"
}
```

#### 6. Get PSAK Standards List
```http
GET /standards
Authorization: Bearer {token}
```

## Data Structures

### PSAKBalanceSheetData
```go
type PSAKBalanceSheetData struct {
    CompanyInfo         CompanyInfo           `json:"company_info"`
    ReportingDate       time.Time             `json:"reporting_date"`
    Currency            string                `json:"currency"`
    
    // Assets
    CurrentAssets       PSAKAssetSection      `json:"current_assets"`
    NonCurrentAssets    PSAKAssetSection      `json:"non_current_assets"`
    TotalAssets         decimal.Decimal       `json:"total_assets"`
    
    // Liabilities
    CurrentLiabilities  PSAKLiabilitySection  `json:"current_liabilities"`
    NonCurrentLiabilities PSAKLiabilitySection `json:"non_current_liabilities"`
    TotalLiabilities    decimal.Decimal       `json:"total_liabilities"`
    
    // Equity
    Equity              PSAKEquitySection     `json:"equity"`
    TotalEquity         decimal.Decimal       `json:"total_equity"`
    
    // Validation
    IsBalanced          bool                  `json:"is_balanced"`
    BalanceDifference   decimal.Decimal       `json:"balance_difference"`
    
    // Compliance
    PSAKCompliance      PSAKComplianceInfo    `json:"psak_compliance"`
    GeneratedAt         time.Time             `json:"generated_at"`
}
```

### PSAKProfitLossData
```go
type PSAKProfitLossData struct {
    CompanyInfo         CompanyInfo           `json:"company_info"`
    PeriodStart         time.Time             `json:"period_start"`
    PeriodEnd           time.Time             `json:"period_end"`
    Currency            string                `json:"currency"`
    
    // Revenue
    Revenue             PSAKRevenueSection    `json:"revenue"`
    TotalRevenue        decimal.Decimal       `json:"total_revenue"`
    
    // Cost of Sales
    CostOfSales         PSAKExpenseSection    `json:"cost_of_sales"`
    TotalCostOfSales    decimal.Decimal       `json:"total_cost_of_sales"`
    
    // Profitability
    GrossProfit         decimal.Decimal       `json:"gross_profit"`
    GrossProfitMargin   decimal.Decimal       `json:"gross_profit_margin"`
    OperatingProfit     decimal.Decimal       `json:"operating_profit"`
    OperatingMargin     decimal.Decimal       `json:"operating_margin"`
    ProfitBeforeTax     decimal.Decimal       `json:"profit_before_tax"`
    NetProfit           decimal.Decimal       `json:"net_profit"`
    NetProfitMargin     decimal.Decimal       `json:"net_profit_margin"`
    
    // OCI and EPS
    OtherComprehensiveIncome PSAKOCISection   `json:"other_comprehensive_income"`
    TotalComprehensiveIncome decimal.Decimal  `json:"total_comprehensive_income"`
    BasicEPS            decimal.Decimal       `json:"basic_eps"`
    DilutedEPS          decimal.Decimal       `json:"diluted_eps"`
    WeightedAvgShares   decimal.Decimal       `json:"weighted_avg_shares"`
    
    // Compliance
    PSAKCompliance      PSAKComplianceInfo    `json:"psak_compliance"`
    GeneratedAt         time.Time             `json:"generated_at"`
}
```

### PSAKCashFlowData
```go
type PSAKCashFlowData struct {
    CompanyInfo         CompanyInfo           `json:"company_info"`
    PeriodStart         time.Time             `json:"period_start"`
    PeriodEnd           time.Time             `json:"period_end"`
    Currency            string                `json:"currency"`
    Method              string                `json:"method"` // "DIRECT" or "INDIRECT"
    
    // Cash positions
    BeginningCash       decimal.Decimal       `json:"beginning_cash"`
    EndingCash          decimal.Decimal       `json:"ending_cash"`
    
    // Cash flows
    OperatingActivities PSAKCashFlowSection   `json:"operating_activities"`
    InvestingActivities PSAKCashFlowSection   `json:"investing_activities"`
    FinancingActivities PSAKCashFlowSection   `json:"financing_activities"`
    
    NetOperatingCash    decimal.Decimal       `json:"net_operating_cash"`
    NetInvestingCash    decimal.Decimal       `json:"net_investing_cash"`
    NetFinancingCash    decimal.Decimal       `json:"net_financing_cash"`
    NetCashIncrease     decimal.Decimal       `json:"net_cash_increase"`
    
    // FX effects
    ForeignExchangeEffect decimal.Decimal     `json:"foreign_exchange_effect"`
    
    // Reconciliation (for indirect method)
    Reconciliation      PSAKCashReconciliation `json:"reconciliation,omitempty"`
    
    // Compliance
    PSAKCompliance      PSAKComplianceInfo    `json:"psak_compliance"`
    GeneratedAt         time.Time             `json:"generated_at"`
}
```

### PSAKComplianceInfo
```go
type PSAKComplianceInfo struct {
    StandardsApplied    []string              `json:"standards_applied"`
    ComplianceLevel     string                `json:"compliance_level"` // FULL, PARTIAL, NON_COMPLIANT
    ComplianceScore     decimal.Decimal       `json:"compliance_score"` // 0-100
    NonComplianceIssues []PSAKIssue          `json:"non_compliance_issues,omitempty"`
    Recommendations     []string              `json:"recommendations,omitempty"`
    LastReviewDate      time.Time             `json:"last_review_date"`
    ReviewerNotes       string                `json:"reviewer_notes,omitempty"`
}
```

## Implementation Details

### Account Classification Logic

#### Current Assets Classification (Aset Lancar)
Sesuai PSAK 1 paragraf 66, aset diklasifikasikan sebagai lancar jika:
- Diperkirakan akan direalisasi atau dimaksudkan untuk dijual/digunakan dalam siklus operasi normal
- Dimiliki terutama untuk tujuan diperdagangkan
- Diperkirakan akan direalisasi dalam 12 bulan setelah periode pelaporan
- Berupa kas atau setara kas

```go
func (s *PSAKCompliantReportService) classifyCurrentAssets(balances map[uint64]decimal.Decimal) ([]PSAKBalanceItem, decimal.Decimal) {
    // Implementation berdasarkan mapping account types
    // Contoh: Account codes 1000-1999 = Current Assets
    currentAssets := []PSAKBalanceItem{}
    total := decimal.Zero
    
    for accountID, balance := range balances {
        account := s.getAccountByID(accountID)
        if s.isCurrentAsset(account) {
            item := PSAKBalanceItem{
                AccountID:      accountID,
                AccountCode:    account.Code,
                AccountName:    account.Name,
                Amount:         balance,
                Classification: "LANCAR",
                PSAKReference:  "PSAK 1 paragraf 66",
            }
            currentAssets = append(currentAssets, item)
            total = total.Add(balance)
        }
    }
    
    return currentAssets, total
}
```

#### Revenue Classification (Klasifikasi Pendapatan)
Sesuai PSAK 23, pendapatan diklasifikasikan berdasarkan:
- Penjualan barang
- Penyediaan jasa
- Penggunaan aset entitas oleh pihak lain yang menghasilkan bunga, royalti, dan dividen

```go
func (s *PSAKCompliantReportService) classifyRevenue(activities map[uint64]decimal.Decimal) (PSAKRevenueSection, decimal.Decimal) {
    revenue := PSAKRevenueSection{
        Name: "Pendapatan",
        Items: []PSAKPLItem{},
    }
    total := decimal.Zero
    
    for accountID, activity := range activities {
        account := s.getAccountByID(accountID)
        if s.isRevenueAccount(account) {
            item := PSAKPLItem{
                AccountID:     accountID,
                AccountCode:   account.Code,
                AccountName:   account.Name,
                Amount:        activity,
                PSAKReference: "PSAK 23",
            }
            revenue.Items = append(revenue.Items, item)
            total = total.Add(activity)
        }
    }
    
    revenue.Subtotal = total
    return revenue, total
}
```

### Cash Flow Methods

#### Direct Method Implementation
```go
func (s *PSAKCompliantReportService) generateDirectOperatingCashFlow(startDate, endDate time.Time) PSAKCashFlowSection {
    section := PSAKCashFlowSection{
        Name: "Arus Kas dari Aktivitas Operasi (Metode Langsung)",
        Items: []PSAKCashFlowItem{},
    }
    
    // Penerimaan kas dari pelanggan
    customerReceipts := s.getCashReceiptsFromCustomers(startDate, endDate)
    section.Items = append(section.Items, PSAKCashFlowItem{
        Description:   "Penerimaan kas dari pelanggan",
        Amount:        customerReceipts,
        Type:         "INFLOW",
        Classification: "OPERASI",
        Method:       "LANGSUNG",
        PSAKReference: "PSAK 2 paragraf 18(a)",
    })
    
    // Pembayaran kas kepada pemasok dan karyawan
    supplierPayments := s.getCashPaymentsToSuppliers(startDate, endDate)
    section.Items = append(section.Items, PSAKCashFlowItem{
        Description:   "Pembayaran kas kepada pemasok dan karyawan",
        Amount:        supplierPayments.Neg(),
        Type:         "OUTFLOW",
        Classification: "OPERASI",
        Method:       "LANGSUNG",
        PSAKReference: "PSAK 2 paragraf 18(b)",
    })
    
    // Calculate subtotal
    for _, item := range section.Items {
        section.Subtotal = section.Subtotal.Add(item.Amount)
    }
    
    return section
}
```

#### Indirect Method Implementation
```go
func (s *PSAKCompliantReportService) generateIndirectOperatingCashFlow(startDate, endDate time.Time) PSAKCashFlowSection {
    section := PSAKCashFlowSection{
        Name: "Arus Kas dari Aktivitas Operasi (Metode Tidak Langsung)",
        Items: []PSAKCashFlowItem{},
    }
    
    // Mulai dengan laba neto
    netProfit := s.getNetProfit(startDate, endDate)
    section.Items = append(section.Items, PSAKCashFlowItem{
        Description:   "Laba neto",
        Amount:        netProfit,
        Type:         "ADJUSTMENT",
        Classification: "OPERASI",
        Method:       "TIDAK_LANGSUNG",
        PSAKReference: "PSAK 2 paragraf 20(a)",
    })
    
    // Penyesuaian untuk item non-kas
    depreciation := s.getDepreciationExpense(startDate, endDate)
    section.Items = append(section.Items, PSAKCashFlowItem{
        Description:   "Penyusutan",
        Amount:        depreciation,
        Type:         "ADJUSTMENT",
        Classification: "OPERASI",
        Method:       "TIDAK_LANGSUNG",
        PSAKReference: "PSAK 2 paragraf 20(b)",
    })
    
    // Perubahan modal kerja
    workingCapitalChanges := s.getWorkingCapitalChanges(startDate, endDate)
    for _, change := range workingCapitalChanges {
        section.Items = append(section.Items, change)
    }
    
    // Calculate subtotal
    for _, item := range section.Items {
        section.Subtotal = section.Subtotal.Add(item.Amount)
    }
    
    return section
}
```

## Compliance Checking

### Compliance Scoring System

Sistem scoring menggunakan skala 0-100 berdasarkan:
- **Kelengkapan penyajian** (30%): Semua elemen wajib sesuai PSAK tersaji
- **Klasifikasi yang tepat** (25%): Item diklasifikasikan sesuai standar
- **Pengungkapan memadai** (20%): Informasi tambahan yang diperlukan tersedia
- **Format sesuai standar** (15%): Struktur laporan mengikuti template PSAK
- **Konsistensi** (10%): Konsistensi antar laporan

### Compliance Levels
- **FULL (95-100)**: Fully compliant dengan semua persyaratan PSAK
- **SUBSTANTIAL (80-94)**: Substantially compliant dengan minor issues
- **PARTIAL (60-79)**: Partially compliant dengan beberapa missing elements
- **NON_COMPLIANT (0-59)**: Not compliant dengan major issues

### Sample Compliance Check
```go
func (s *PSAKCompliantReportService) generateBalanceSheetCompliance(bs *PSAKBalanceSheetData) PSAKComplianceInfo {
    compliance := PSAKComplianceInfo{
        StandardsApplied: []string{"PSAK 1", "PSAK 14", "PSAK 16", "PSAK 46"},
        LastReviewDate:   time.Now(),
    }
    
    score := decimal.NewFromInt(100)
    var issues []PSAKIssue
    var recommendations []string
    
    // Check fundamental equation: Assets = Liabilities + Equity
    if !bs.IsBalanced {
        issues = append(issues, PSAKIssue{
            StandardReference: "PSAK 1 fundamental equation",
            IssueDescription:  "Balance sheet does not balance",
            Severity:         "HIGH",
            RecommendedAction: "Review journal entries and account balances",
        })
        score = score.Sub(decimal.NewFromInt(20))
    }
    
    // Check current asset classification
    if bs.CurrentAssets.Subtotal.IsZero() {
        issues = append(issues, PSAKIssue{
            StandardReference: "PSAK 1 paragraf 66",
            IssueDescription:  "No current assets classified",
            Severity:         "MEDIUM",
            RecommendedAction: "Classify assets based on liquidity and maturity",
        })
        score = score.Sub(decimal.NewFromInt(10))
    }
    
    // Set compliance level based on score
    scoreFloat, _ := score.Float64()
    switch {
    case scoreFloat >= 95:
        compliance.ComplianceLevel = "FULL"
    case scoreFloat >= 80:
        compliance.ComplianceLevel = "SUBSTANTIAL"
    case scoreFloat >= 60:
        compliance.ComplianceLevel = "PARTIAL"
    default:
        compliance.ComplianceLevel = "NON_COMPLIANT"
    }
    
    compliance.ComplianceScore = score
    compliance.NonComplianceIssues = issues
    compliance.Recommendations = recommendations
    
    return compliance
}
```

## Usage Examples

### 1. Generate PSAK Balance Sheet
```bash
curl -X POST http://localhost:8080/api/v1/reports/psak/balance-sheet \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "as_of_date": "2024-12-31"
  }'
```

### 2. Generate PSAK Profit & Loss
```bash
curl -X POST http://localhost:8080/api/v1/reports/psak/profit-loss \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "start_date": "2024-01-01",
    "end_date": "2024-12-31"
  }'
```

### 3. Generate PSAK Cash Flow (Direct Method)
```bash
curl -X POST http://localhost:8080/api/v1/reports/psak/cash-flow \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "start_date": "2024-01-01",
    "end_date": "2024-12-31",
    "method": "DIRECT"
  }'
```

### 4. Check PSAK Compliance Summary
```bash
curl -X GET "http://localhost:8080/api/v1/reports/psak/compliance-summary?as_of_date=2024-12-31&start_date=2024-01-01&end_date=2024-12-31" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. Get PSAK Standards List
```bash
curl -X GET http://localhost:8080/api/v1/reports/psak/standards \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Configuration

### Account Mapping Configuration
Untuk implementasi yang lengkap, diperlukan konfigurasi mapping akun ke kategori PSAK:

```yaml
# config/psak_account_mapping.yaml
psak_account_mapping:
  current_assets:
    cash_and_equivalents: ["1010", "1020", "1030"]  # Kas, Bank, Deposito
    trade_receivables: ["1110", "1120"]             # Piutang Usaha, Piutang Lain-lain
    inventories: ["1210", "1220", "1230"]           # Bahan Baku, WIP, Barang Jadi
    prepaid_expenses: ["1310", "1320"]              # Beban Dibayar Dimuka
  
  non_current_assets:
    property_plant_equipment: ["1510", "1520", "1530"]  # Tanah, Bangunan, Mesin
    intangible_assets: ["1610", "1620"]                  # Goodwill, Patent
    long_term_investments: ["1710", "1720"]              # Investasi Saham, Obligasi
  
  current_liabilities:
    trade_payables: ["2010", "2020"]                # Utang Usaha, Utang Lain-lain
    accrued_liabilities: ["2110", "2120"]           # Beban yang Masih Harus Dibayar
    short_term_borrowings: ["2210"]                 # Pinjaman Jangka Pendek
    tax_liabilities: ["2310", "2320"]               # Utang Pajak PPh, PPN
  
  revenue_accounts: ["4010", "4020", "4030"]        # Penjualan, Jasa, Pendapatan Lain
  cogs_accounts: ["5010", "5020"]                   # HPP Barang, HPP Jasa
  operating_expense_accounts: ["6010", "6020", "6030"] # Beban Penjualan, Umum, Adm
```

### Environment Variables
```env
# PSAK Compliance Settings
PSAK_COMPLIANCE_STRICT_MODE=true
PSAK_DEFAULT_CURRENCY=IDR
PSAK_DECIMAL_PLACES=2
PSAK_DATE_FORMAT=2006-01-02

# Chart of Accounts Configuration
COA_MAPPING_FILE=config/psak_account_mapping.yaml
COA_VALIDATION_ENABLED=true

# Compliance Scoring Weights
PSAK_SCORING_PRESENTATION_WEIGHT=30
PSAK_SCORING_CLASSIFICATION_WEIGHT=25
PSAK_SCORING_DISCLOSURE_WEIGHT=20
PSAK_SCORING_FORMAT_WEIGHT=15
PSAK_SCORING_CONSISTENCY_WEIGHT=10
```

## Testing

### Unit Tests
```go
// Test Balance Sheet Compliance
func TestPSAKBalanceSheetCompliance(t *testing.T) {
    service := setupPSAKTestService()
    
    // Test data setup
    testDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
    
    // Generate balance sheet
    bs, err := service.GeneratePSAKBalanceSheet(testDate)
    require.NoError(t, err)
    require.NotNil(t, bs)
    
    // Test compliance
    assert.True(t, bs.IsBalanced)
    assert.Equal(t, "FULL", bs.PSAKCompliance.ComplianceLevel)
    assert.True(t, bs.PSAKCompliance.ComplianceScore.GreaterThanOrEqual(decimal.NewFromInt(95)))
}

// Test P&L PSAK 23 Revenue Recognition
func TestPSAKRevenueClassification(t *testing.T) {
    service := setupPSAKTestService()
    
    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
    
    pl, err := service.GeneratePSAKProfitLoss(startDate, endDate)
    require.NoError(t, err)
    require.NotNil(t, pl)
    
    // Test revenue sections
    assert.NotEmpty(t, pl.Revenue.Items)
    assert.True(t, pl.Revenue.Subtotal.GreaterThan(decimal.Zero))
    
    // Test PSAK 23 compliance
    foundPSAK23 := false
    for _, std := range pl.PSAKCompliance.StandardsApplied {
        if std == "PSAK 23" {
            foundPSAK23 = true
            break
        }
    }
    assert.True(t, foundPSAK23)
}
```

### Integration Tests
```bash
# Run PSAK compliance tests
go test ./tests/psak_compliance_test.go -v

# Run API endpoint tests
go test ./tests/psak_api_test.go -v

# Run full integration test suite
go test ./tests/integration/psak_integration_test.go -v
```

### Manual Testing Checklist

#### Balance Sheet Tests
- [ ] Assets = Liabilities + Equity equation holds
- [ ] Current vs Non-current classification correct
- [ ] All mandatory PSAK 1 elements present
- [ ] Proper Indonesian terminology used
- [ ] Compliance score ≥ 95% for test data

#### P&L Tests
- [ ] Revenue classified per PSAK 23
- [ ] COGS calculated per PSAK 14
- [ ] Tax expense per PSAK 46
- [ ] EPS calculated per PSAK 56
- [ ] OCI properly separated
- [ ] All ratios calculated correctly

#### Cash Flow Tests
- [ ] Both direct and indirect methods work
- [ ] Operating, investing, financing classification correct
- [ ] Cash reconciliation balances
- [ ] Foreign exchange effects handled
- [ ] Complies with PSAK 2 requirements

## Future Enhancements

### Phase 2: Advanced PSAK Features
1. **PSAK 13 - Properti Investasi**
   - Fair value vs cost model
   - Rental income recognition

2. **PSAK 19 - Aset Tidak Berwujud**
   - Amortization schedules
   - Impairment testing

3. **PSAK 24 - Imbalan Kerja**
   - Employee benefit calculations
   - Actuarial valuations

4. **PSAK 48 - Penurunan Nilai Aset**
   - Impairment testing
   - Value in use calculations

### Phase 3: Advanced Reporting Features
1. **Consolidated Statements**
   - Multi-entity reporting
   - Elimination entries

2. **Segment Reporting**
   - Business segment analysis
   - Geographic segment reporting

3. **Interim Reporting**
   - Quarterly statements
   - Year-to-date comparisons

4. **Comparative Reporting**
   - Prior period comparisons
   - Variance analysis

### Phase 4: Digital Integration
1. **XBRL Taxonomy Support**
   - Indonesian XBRL taxonomy mapping
   - Electronic filing preparation

2. **Regulatory Submissions**
   - OJK (Financial Services Authority) formats
   - Bank Indonesia reporting

3. **Audit Trail Enhancement**
   - Detailed supporting schedules
   - Audit working papers generation

4. **Real-time Compliance Monitoring**
   - Continuous compliance checking
   - Alert system for violations

### Phase 5: Advanced Analytics
1. **Compliance Dashboard**
   - Real-time compliance metrics
   - Trend analysis

2. **Benchmarking**
   - Industry comparison
   - Best practice recommendations

3. **Predictive Analytics**
   - Compliance risk assessment
   - Future performance projections

4. **AI-powered Insights**
   - Anomaly detection
   - Automated recommendations

## Conclusion

Implementasi PSAK-compliant financial reports ini menyediakan fondasi yang kuat untuk memenuhi persyaratan standar akuntansi Indonesia. Dengan struktur yang modular dan extensible, sistem ini dapat dikembangkan lebih lanjut untuk mendukung standar PSAK tambahan dan fitur-fitur advanced reporting.

Fitur-fitur utama yang telah diimplementasikan:
- **Compliance dengan 7 standar PSAK utama**
- **Laporan keuangan lengkap sesuai format Indonesia**
- **Sistem scoring dan monitoring compliance**
- **API endpoints yang comprehensive**
- **Integration dengan Single Source of Truth (SSOT)**
- **Audit trail dan dokumentasi lengkap**

Sistem ini siap untuk digunakan dalam produksi dengan beberapa customization untuk mapping akun spesifik perusahaan dan konfigurasi tambahan sesuai kebutuhan bisnis.