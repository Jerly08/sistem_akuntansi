# Panduan Perbaikan Laporan Laba Rugi

## Ringkasan Masalah
Laporan Laba Rugi menampilkan data yang tidak akurat karena:
1. **Account categorization yang salah** - Account 5101 tidak dikenali sebagai COGS
2. **Ketidaksesuaian antara account balance dan journal entries** - Balance ada di tabel accounts tapi tidak ada di journal_lines
3. **Logic perhitungan yang tidak tepat** dalam financial report service

## ✅ Solusi yang Telah Diimplementasi

### 1. Database Fixes (Sudah Selesai)
```sql
-- Fix account categories
UPDATE accounts SET category = 'COST_OF_GOODS_SOLD' WHERE code = '5101';

-- Create missing journal entries for account balances
-- (Sudah dijalankan via fix_profit_loss.go)
```

### 2. Enhanced Financial Report Service
File: `services/enhanced_financial_report_service.go` telah dibuat dengan:
- ✅ **Proper COGS categorization** dengan multiple checks
- ✅ **Enhanced balance calculation** yang menggabungkan journal entries dan account balances
- ✅ **Automatic account category validation**
- ✅ **Fallback mechanism** jika journal entries tidak ada

## 🚀 Implementasi untuk Production

### Step 1: Update Controller
Modifikasi `controllers/report_controller.go` untuk menggunakan enhanced service:

```go
// Add enhanced service to controller
type ReportController struct {
    reportService        *services.ReportService
    financialReportService services.FinancialReportService
    enhancedReportService services.EnhancedFinancialReportService // ← Add this
    professionalService  *services.ProfessionalReportService
    standardizedService  *services.StandardizedReportService
}

// Update GetProfitLoss method
func (rc *ReportController) GetProfitLoss(c *gin.Context) {
    // ... existing validation code ...

    // Use enhanced service instead of regular one
    pnl, err := rc.enhancedReportService.GenerateEnhancedProfitLossStatement(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"status": "success", "data": pnl})
}
```

### Step 2: Update Dependency Injection
Modifikasi main.go atau di tempat controller di-initialize:

```go
// Initialize enhanced service
enhancedReportService := services.NewEnhancedFinancialReportService(db, accountRepo, journalRepo)

// Pass to controller
reportController := controllers.NewReportController(
    reportService,
    financialReportService,
    enhancedReportService, // ← Add this parameter
    professionalService,
    standardizedService,
)
```

### Step 3: Add New Endpoint (Optional)
Tambahkan endpoint baru untuk enhanced report:

```go
// Di routes/report_routes.go
reports.GET("/enhanced/profit-loss", reportController.GetEnhancedProfitLoss)
```

## 🔍 Verifikasi Hasil

### Test Endpoint
```bash
curl -X GET "http://localhost:8080/api/reports/profit-loss?start_date=2025-01-01&end_date=2025-12-31&format=json"
```

### Expected Results
```json
{
  "status": "success",
  "data": {
    "total_revenue": 20000000.00,
    "total_cogs": 32400000.00,
    "gross_profit": -12400000.00,
    "total_expenses": 5000000.00,
    "net_income": -17400000.00,
    "revenue": [
      {
        "account_code": "4900",
        "account_name": "Other Income", 
        "balance": 20000000.00
      }
    ],
    "cogs": [
      {
        "account_code": "5101",
        "account_name": "Harga Pokok Penjualan",
        "balance": 32400000.00
      }
    ],
    "expenses": [
      {
        "account_code": "5000", 
        "account_name": "EXPENSES",
        "balance": 5000000.00
      }
    ]
  }
}
```

## ⚠️ Catatan Penting

1. **Data Consistency**: Pastikan semua journal entries future dibuat dengan benar
2. **Account Categories**: Review semua account categories untuk memastikan mapping yang tepat:
   - Revenue accounts: `OPERATING_REVENUE`, `OTHER_INCOME`, dll
   - COGS accounts: `COST_OF_GOODS_SOLD`, `DIRECT_MATERIAL`, dll  
   - Operating Expenses: `OPERATING_EXPENSE`, `ADMINISTRATIVE_EXPENSE`, dll

3. **Balance Synchronization**: 
   - Gunakan journal entries sebagai single source of truth
   - Account balance di tabel `accounts` sebaiknya di-update secara otomatis dari journal entries

## 🔄 Maintenance

### Regular Checks
1. **Monthly**: Verifikasi consistency antara account balances dan journal entries
2. **Quarterly**: Review account categorization
3. **Yearly**: Audit full P&L calculation logic

### Monitoring
- Monitor performa query enhanced service
- Log any fallback ke account balance (should be minimal)
- Alert jika ada ketidaksesuaian significant

## 📋 Files Created/Modified

### New Files:
- ✅ `services/enhanced_financial_report_service.go` - Enhanced service dengan logic perbaikan
- ✅ `fix_profit_loss.go` - Script untuk memperbaiki data existing  
- ✅ `debug_profit_loss.go` - Script untuk debug dan testing
- ✅ `fix_profit_loss_data.sql` - SQL script untuk perbaikan manual

### Files to Modify:
- `controllers/report_controller.go` - Update untuk gunakan enhanced service
- `main.go` - Add enhanced service ke dependency injection
- `routes/report_routes.go` - Optional: Add enhanced endpoint

---

**Status**: ✅ Database sudah diperbaiki, service enhancement sudah siap untuk implementasi
**Next Action**: Implement enhanced service di controller dan test di frontend
