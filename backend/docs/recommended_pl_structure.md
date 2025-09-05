# Rekomendasi Struktur Profit & Loss Statement

## 1. STRUKTUR P&L YANG DISARANKAN

```
REVENUE
├── Sales Revenue                    Rp 500,000,000
├── Service Revenue                  Rp 100,000,000
└── Other Income                     Rp  50,000,000
TOTAL REVENUE                        Rp 650,000,000

COST OF GOODS SOLD (COGS)
├── Direct Materials                 Rp 200,000,000
├── Direct Labor                     Rp  50,000,000
├── Manufacturing Overhead           Rp  50,000,000
└── TOTAL COGS                       Rp 300,000,000

GROSS PROFIT                         Rp 350,000,000
Gross Profit Margin                  53.85%

OPERATING EXPENSES
├── Administrative Expenses          Rp  75,000,000
├── Selling & Marketing Expenses     Rp  25,000,000
├── General Expenses                 Rp  50,000,000
├── Depreciation & Amortization      Rp  15,000,000
└── TOTAL OPERATING EXPENSES         Rp 165,000,000

OPERATING INCOME (EBIT)              Rp 185,000,000
Operating Margin                     28.46%

OTHER INCOME & EXPENSES
├── Interest Income                  Rp   5,000,000
├── Interest Expense                 (Rp 10,000,000)
├── Other Non-Operating Income       Rp   2,000,000
└── NET OTHER EXPENSES               (Rp  3,000,000)

INCOME BEFORE TAX                    Rp 182,000,000

INCOME TAX EXPENSE                   (Rp 40,000,000)
Tax Rate                             21.98%

NET INCOME                           Rp 142,000,000
Net Profit Margin                    21.85%

EBITDA                              Rp 200,000,000
EBITDA Margin                       30.77%
```

## 2. PERBAIKAN KODE BACKEND

### A. Update Account Categories untuk P&L

```go
// Enhanced P&L Categories
const (
    // Revenue Categories
    CategorySalesRevenue        = "SALES_REVENUE"
    CategoryServiceRevenue      = "SERVICE_REVENUE"
    CategoryOtherOperatingRev   = "OTHER_OPERATING_REVENUE"
    CategoryNonOperatingRevenue = "NON_OPERATING_REVENUE"
    CategoryInterestIncome      = "INTEREST_INCOME"
    
    // COGS Categories  
    CategoryDirectMaterials     = "DIRECT_MATERIALS"
    CategoryDirectLabor         = "DIRECT_LABOR"
    CategoryManufacturingOH     = "MANUFACTURING_OVERHEAD"
    CategoryFreightIn           = "FREIGHT_IN"
    
    // Operating Expense Categories
    CategoryAdminExpense        = "ADMINISTRATIVE_EXPENSE"
    CategorySellingExpense      = "SELLING_MARKETING_EXPENSE"
    CategoryGeneralExpense      = "GENERAL_EXPENSE"
    CategoryDepreciationExp     = "DEPRECIATION_AMORTIZATION"
    
    // Non-Operating Categories
    CategoryInterestExpense     = "INTEREST_EXPENSE"
    CategoryOtherNonOpExp      = "OTHER_NON_OPERATING_EXPENSE"
    CategoryTaxExpense         = "INCOME_TAX_EXPENSE"
)
```

### B. Enhanced P&L Data Structure

```go
type EnhancedProfitLossData struct {
    Company     CompanyInfo    `json:"company"`
    StartDate   time.Time      `json:"start_date"`
    EndDate     time.Time      `json:"end_date"`
    Currency    string         `json:"currency"`
    
    // Revenue Section
    Revenue struct {
        SalesRevenue    PLSection `json:"sales_revenue"`
        ServiceRevenue  PLSection `json:"service_revenue"`
        OtherRevenue    PLSection `json:"other_revenue"`
        TotalRevenue    float64   `json:"total_revenue"`
    } `json:"revenue"`
    
    // COGS Section
    CostOfGoodsSold struct {
        DirectMaterials     PLSection `json:"direct_materials"`
        DirectLabor         PLSection `json:"direct_labor"`
        ManufacturingOH     PLSection `json:"manufacturing_overhead"`
        OtherCOGS          PLSection `json:"other_cogs"`
        TotalCOGS          float64   `json:"total_cogs"`
    } `json:"cost_of_goods_sold"`
    
    // Profitability Metrics
    GrossProfit       float64 `json:"gross_profit"`
    GrossProfitMargin float64 `json:"gross_profit_margin"`
    
    // Operating Expenses
    OperatingExpenses struct {
        Administrative  PLSection `json:"administrative"`
        SellingMarketing PLSection `json:"selling_marketing"`
        General         PLSection `json:"general"`
        Depreciation    PLSection `json:"depreciation"`
        TotalOpex       float64   `json:"total_operating_expenses"`
    } `json:"operating_expenses"`
    
    // Operating Performance
    OperatingIncome   float64 `json:"operating_income"`
    OperatingMargin   float64 `json:"operating_margin"`
    EBITDA           float64 `json:"ebitda"`
    EBITDAMargin     float64 `json:"ebitda_margin"`
    
    // Non-Operating Items
    OtherIncomeExpense struct {
        InterestIncome   PLSection `json:"interest_income"`
        InterestExpense  PLSection `json:"interest_expense"`
        OtherIncome      PLSection `json:"other_income"`
        OtherExpense     PLSection `json:"other_expense"`
        NetOtherIncome   float64   `json:"net_other_income"`
    } `json:"other_income_expense"`
    
    // Tax and Final Result
    IncomeBeforeTax   float64 `json:"income_before_tax"`
    TaxExpense        float64 `json:"tax_expense"`
    TaxRate          float64 `json:"tax_rate"`
    NetIncome        float64 `json:"net_income"`
    NetIncomeMargin  float64 `json:"net_income_margin"`
    
    GeneratedAt      time.Time `json:"generated_at"`
}
```

## 3. PERBAIKAN FRONTEND DISPLAY

### A. Component Structure
```tsx
function EnhancedProfitLossStatement({ data }: { data: EnhancedProfitLossData }) {
    return (
        <div className="profit-loss-statement">
            {/* Revenue Section */}
            <PLSection title="REVENUE" data={data.revenue} />
            
            {/* COGS Section */}
            <PLSection title="COST OF GOODS SOLD" data={data.cost_of_goods_sold} />
            
            {/* Gross Profit */}
            <PLMetric 
                label="GROSS PROFIT" 
                amount={data.gross_profit}
                percentage={data.gross_profit_margin}
                highlighted={true}
            />
            
            {/* Operating Expenses */}
            <PLSection title="OPERATING EXPENSES" data={data.operating_expenses} />
            
            {/* Operating Income */}
            <PLMetric 
                label="OPERATING INCOME (EBIT)" 
                amount={data.operating_income}
                percentage={data.operating_margin}
                highlighted={true}
            />
            
            {/* Other Income/Expense */}
            <PLSection title="OTHER INCOME & EXPENSES" data={data.other_income_expense} />
            
            {/* Pre-tax Income */}
            <PLMetric 
                label="INCOME BEFORE TAX" 
                amount={data.income_before_tax}
                highlighted={true}
            />
            
            {/* Tax */}
            <PLMetric 
                label="INCOME TAX EXPENSE" 
                amount={-data.tax_expense}
                percentage={data.tax_rate}
            />
            
            {/* Net Income */}
            <PLMetric 
                label="NET INCOME" 
                amount={data.net_income}
                percentage={data.net_income_margin}
                highlighted={true}
                final={true}
            />
        </div>
    );
}
```

## 4. KEY FINANCIAL RATIOS

### A. Profitability Ratios
- **Gross Profit Margin** = (Gross Profit / Total Revenue) × 100%
- **Operating Margin** = (Operating Income / Total Revenue) × 100%  
- **EBITDA Margin** = (EBITDA / Total Revenue) × 100%
- **Net Profit Margin** = (Net Income / Total Revenue) × 100%

### B. Additional Metrics
- **Revenue Growth** = ((Current Period Revenue - Previous Period Revenue) / Previous Period Revenue) × 100%
- **Operating Leverage** = % Change in Operating Income / % Change in Revenue
- **Tax Rate** = (Tax Expense / Income Before Tax) × 100%

## 5. BUSINESS RULES VALIDATION

### A. Data Validation Rules
1. **Revenue must be positive or zero**
2. **COGS cannot exceed Sales Revenue**
3. **Tax rate should be reasonable (0-50%)**
4. **All percentages should sum logically**

### B. Account Mapping Rules
```go
func (ers *EnhancedReportService) isOperatingRevenue(category string) bool {
    operatingCategories := []string{
        CategorySalesRevenue,
        CategoryServiceRevenue,
        CategoryOtherOperatingRev,
    }
    return contains(operatingCategories, category)
}

func (ers *EnhancedReportService) isCOGS(category string) bool {
    cogsCategories := []string{
        CategoryDirectMaterials,
        CategoryDirectLabor,
        CategoryManufacturingOH,
        CategoryFreightIn,
    }
    return contains(cogsCategories, category)
}

func (ers *EnhancedReportService) isOperatingExpense(category string) bool {
    opexCategories := []string{
        CategoryAdminExpense,
        CategorySellingExpense,
        CategoryGeneralExpense,
        CategoryDepreciationExp,
    }
    return contains(opexCategories, category)
}
```

## 6. REPORT FORMATTING STANDARDS

### A. Number Formatting
- **Currency**: IDR format with proper thousand separators
- **Percentages**: Show 2 decimal places
- **Negative numbers**: Show in parentheses (Rp 10,000,000)

### B. Color Coding
- **Positive values**: Green or black
- **Negative values**: Red 
- **Key metrics**: Bold and highlighted
- **Subtotals**: Underlined

## KESIMPULAN

Laporan P&L Anda sudah memiliki dasar yang baik, namun perlu perbaikan pada:
1. **Separasi COGS dari Operating Expenses** 
2. **Penambahan key metrics** (Gross Profit, Operating Income, EBITDA)
3. **Kategori akun yang lebih spesifik**
4. **Tax expense handling**
5. **Financial ratios calculation**

Implementasi rekomendasi ini akan memberikan laporan P&L yang lebih akurat dan sesuai dengan standar akuntansi internasional.
