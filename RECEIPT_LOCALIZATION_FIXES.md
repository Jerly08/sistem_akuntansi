# Receipt PDF Localization Fixes

## Issue Description
The system settings show language is set to English, but the sales receipt PDF was still generating in Indonesian, showing "KWITANSI" instead of "RECEIPT".

## Root Cause
The `generateSaleReceiptPDF` function in `services/pdf_service.go` was using hardcoded Indonesian text instead of respecting the system language configuration.

## Solution Applied

### 1. Enhanced Localization Dictionary
**File**: `backend/utils/localization.go`

Added receipt-specific translations for both Indonesian and English:

```go
// Indonesian translations (lines 151-157)
"receipt":                   "KWITANSI",
"received_from":             "Sudah Terima Dari",
"amount_in_words":           "Banyaknya Uang",
"for_payment":               "Untuk Pembayaran",
"amount_rp":                 "Jumlah Rp.",
"received_by":               "Diterima oleh",

// English translations (lines 292-298)
"receipt":                   "RECEIPT",
"received_from":             "Received From",
"amount_in_words":           "Amount in Words",
"for_payment":               "For Payment of",
"amount_rp":                 "Amount Rp.",
"received_by":               "Received by",
```

### 2. Updated PDF Generation Function
**File**: `backend/services/pdf_service.go` (function `generateSaleReceiptPDF`)

#### Key Changes:

1. **Added localization helper** (lines 2883-2888):
```go
language := utils.GetUserLanguageFromSettings(p.db)
loc := func(key, fallback string) string {
    t := utils.T(key, language)
    if t == key { return fallback }
    return t
}
```

2. **Dynamic title** (lines 2893-2899):
```go
receiptTitle := loc("receipt", "RECEIPT")
pdf.CellFormat(contentW, 14, receiptTitle, "", 0, "C", false, 0, "")
// Dynamic underline width based on title
titleWidth := pdf.GetStringWidth(receiptTitle)
cx := lm + (contentW-titleWidth)/2
pdf.Line(cx, pdf.GetY(), cx+titleWidth, pdf.GetY())
```

3. **Localized field labels**:
   - "Received From" / "Sudah Terima Dari" 
   - "Amount in Words" / "Banyaknya Uang"
   - "For Payment of" / "Untuk Pembayaran"
   - "Amount Rp." / "Jumlah Rp."
   - "Received by" / "Diterima oleh"

## How It Works

### Language Detection Flow:
1. System reads language setting from `settings` table (`Settings.Language` field)
2. PDF service calls `utils.GetUserLanguageFromSettings(p.db)` to get current language
3. Each text element uses `loc(key, fallback)` to get appropriate translation
4. Falls back to English if translation not found, then to provided fallback

### System Settings Integration:
- Language setting stored in database: `settings.language` ('id' or 'en')
- Default language: Indonesian ('id') as per model definition
- Frontend shows language dropdown in System Settings page
- Changes take effect immediately for new PDF generations

## Testing
Added test functionality in `test_pdf_formatting.go` to generate both sales report and receipt PDFs for verification.

## Verification Steps
1. **Change language to English**: Go to System Settings → Set Language to "English"
2. **Generate receipt**: Go to Sales → Select paid sale → Create Receipt
3. **Verify title**: Should show "RECEIPT" instead of "KWITANSI"
4. **Verify labels**: All field labels should be in English
5. **Test Indonesian**: Change language back to Indonesian to verify backward compatibility

## Files Modified
- `backend/utils/localization.go` - Added receipt translations
- `backend/services/pdf_service.go` - Updated generateSaleReceiptPDF function
- `backend/test_pdf_formatting.go` - Added receipt testing

## Database Schema
The `settings` table already includes the `language` field:
```sql
language VARCHAR DEFAULT 'id'
```

No database changes required - the system was ready for localization.