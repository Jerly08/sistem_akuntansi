package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"app-sistem-akuntansi/models"

	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

// PDFService implements PDFServiceInterface
type PDFService struct{
	db *gorm.DB
}


// NewPDFService creates a new PDF service instance
func NewPDFService(db *gorm.DB) PDFServiceInterface {
	return &PDFService{db: db}
}

// formatRupiah formats a number as Indonesian Rupiah
func (p *PDFService) formatRupiah(amount float64) string {
	// Format number with thousand separators
	amountStr := fmt.Sprintf("%.0f", amount)
	if amount != float64(int64(amount)) {
		amountStr = fmt.Sprintf("%.2f", amount)
	}
	
	// Add thousand separators
	formattedAmount := p.addThousandSeparators(amountStr)
	
	return "Rp " + formattedAmount
}

// getCompanyInfo retrieves company information from settings
func (p *PDFService) getCompanyInfo() (*models.Settings, error) {
	// Return default company info if database is not available
	if p.db == nil {
		return &models.Settings{
			CompanyName:    "PT. Sistem Akuntansi Indonesia",
			CompanyAddress: "Jl. Sudirman Kav. 45-46, Jakarta Pusat 10210, Indonesia",
			CompanyPhone:   "+62-21-5551234",
			CompanyEmail:   "info@sistemakuntansi.co.id",
		}, nil
	}
	
	var settings models.Settings
	err := p.db.First(&settings).Error
	if err != nil {
		// Return default company info if settings not found
		return &models.Settings{
			CompanyName:    "PT. Sistem Akuntansi Indonesia",
			CompanyAddress: "Jl. Sudirman Kav. 45-46, Jakarta Pusat 10210, Indonesia",
			CompanyPhone:   "+62-21-5551234",
			CompanyEmail:   "info@sistemakuntansi.co.id",
		}, nil
	}
	return &settings, nil
}

// addThousandSeparators adds dots as thousand separators for Indonesian currency format
func (p *PDFService) addThousandSeparators(s string) string {
	// Split by decimal point if exists
	parts := strings.Split(s, ".")
	integerPart := parts[0]
	
	// Add thousand separators (dots) to integer part
	if len(integerPart) <= 3 {
		if len(parts) > 1 {
			return integerPart + "," + parts[1]
		}
		return integerPart
	}
	
	var result []string
	for i, digit := range reverse(integerPart) {
		if i > 0 && i%3 == 0 {
			result = append(result, ".")
		}
		result = append(result, string(digit))
	}
	
	formattedInteger := reverse(strings.Join(result, ""))
	
	if len(parts) > 1 {
		return formattedInteger + "," + parts[1]
	}
	return formattedInteger
}

// reverse reverses a string
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// addCompanyLetterhead tries to render company logo as a letterhead at top of the page
func (p *PDFService) addCompanyLetterhead(pdf *gofpdf.Fpdf) {
	settings, err := p.getCompanyInfo()
	if err != nil || settings == nil {
		return
	}
	logo := strings.TrimSpace(settings.CompanyLogo)
	if logo == "" {
		return
	}
	// Map web path "/uploads/..." to local filesystem path "./uploads/..."
	localPath := logo
	if strings.HasPrefix(localPath, "/") {
		localPath = "." + localPath
	}
	// Resolve and ensure file exists
	if _, err := os.Stat(localPath); err != nil {
		// Try joining with working dir uploads/company
		alt := filepath.Clean("./" + strings.TrimPrefix(logo, "/"))
		if _, err2 := os.Stat(alt); err2 != nil {
			return
		}
		localPath = alt
	}
	// Detect image type by magic bytes to avoid invalid JPEG/PNG errors
	imgType := detectImageType(localPath)
	if imgType == "" {
		// Unknown or unsupported image type; skip letterhead to avoid breaking PDF
		return
	}
	// Draw a compact logo at top-left (letterhead style)
	// Get margins and page size
	lm, tm, rm, _ := pdf.GetMargins()
	pageW, _ := pdf.GetPageSize()
	_ = rm // right margin unused in this placement
	
	// Choose a reasonable logo width; larger for landscape pages
	logoW := 35.0
	if pageW > 250 { // landscape A4 ~ 297mm width
		logoW = 40.0
	}
	// Place logo at top-left, slightly below the top margin
	x := lm
	y := tm + 2.0
	// Height=0 preserves aspect ratio
	pdf.ImageOptions(localPath, x, y, logoW, 0, false, gofpdf.ImageOptions{ImageType: imgType, ReadDpi: true}, 0, "")
	
	// Ensure subsequent content starts below the logo area (title printed after this)
	currentY := pdf.GetY()
	// Reserve vertical space at least equal to logo height area
	minY := y + logoW + 4.0
	if currentY < minY {
		pdf.SetY(minY)
	}
}

// detectImageType inspects the file header to determine image type supported by gofpdf (JPG/PNG)
func detectImageType(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	buf := make([]byte, 8)
	if _, err := f.Read(buf); err != nil {
		return ""
	}
	// JPEG SOI marker
	if len(buf) >= 2 && buf[0] == 0xFF && buf[1] == 0xD8 {
		return "JPG"
	}
	// PNG signature
	if len(buf) >= 8 && buf[0] == 0x89 && buf[1] == 0x50 && buf[2] == 0x4E && buf[3] == 0x47 && buf[4] == 0x0D && buf[5] == 0x0A && buf[6] == 0x1A && buf[7] == 0x0A {
		return "PNG"
	}
	return ""
}

// GenerateInvoicePDF generates a PDF for a sale invoice
func (p *PDFService) GenerateInvoicePDF(sale *models.Sale) ([]byte, error) {
	// Create new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Try adding company letterhead/logo
	p.addCompanyLetterhead(pdf)

	// Set font and place document title to the right of the logo
	lm, tm, rm, _ := pdf.GetMargins()
	pageW, _ := pdf.GetPageSize()
	logoW := 35.0
	if pageW > 250 { // landscape width threshold
		logoW = 40.0
	}
	xStart := lm + logoW + 6
	textW := pageW - rm - xStart
	pdf.SetXY(xStart, tm+2)
pdf.SetFont("Arial", "B", 16)
// Title will be drawn below the logo area to avoid overlap
pdf.Ln(0)

	// Get company info from settings
	companyInfo, err := p.getCompanyInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get company info: %v", err)
	}
	
// Company info from settings – align to the right of the logo area
	// Reuse xStart/textW from above (avoid redeclaration in the same scope)
	pdf.SetXY(xStart, tm+2)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(textW, 8, companyInfo.CompanyName)
	pdf.SetFont("Arial", "", 10)
	pdf.Ln(6); pdf.SetX(xStart)
	pdf.Cell(textW, 5, companyInfo.CompanyAddress)
	pdf.Ln(5); pdf.SetX(xStart)
	pdf.Cell(textW, 5, fmt.Sprintf("Phone: %s", companyInfo.CompanyPhone))
	pdf.Ln(5); pdf.SetX(xStart)
	pdf.Cell(textW, 5, fmt.Sprintf("Email: %s", companyInfo.CompanyEmail))
	
// Ensure following content starts below the logo+info block
minY := tm + logoW + 6
if pdf.GetY() < minY { pdf.SetY(minY) }
// Draw title below the logo
pdf.SetX(lm)
pdf.SetFont("Arial", "B", 16)
pdf.Cell(pageW-lm-rm, 10, "INVOICE")
pdf.Ln(10)

	// Invoice details
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(95, 6, fmt.Sprintf("Invoice Number: %s", sale.InvoiceNumber))
	pdf.Cell(95, 6, fmt.Sprintf("Date: %s", sale.Date.Format("02/01/2006")))
	pdf.Ln(6)
	pdf.Cell(95, 6, fmt.Sprintf("Sale Code: %s", sale.Code))
	if !sale.DueDate.IsZero() {
		pdf.Cell(95, 6, fmt.Sprintf("Due Date: %s", sale.DueDate.Format("02/01/2006")))
	}
	pdf.Ln(10)

	// Customer info
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 6, "Bill To:")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 10)
	// Customer info is always loaded, check if ID is set
	if sale.Customer.ID != 0 {
		pdf.Cell(190, 5, sale.Customer.Name)
		pdf.Ln(5)
		if sale.Customer.Address != "" {
			pdf.Cell(190, 5, sale.Customer.Address)
			pdf.Ln(5)
		}
		if sale.Customer.Phone != "" {
			pdf.Cell(190, 5, fmt.Sprintf("Phone: %s", sale.Customer.Phone))
			pdf.Ln(5)
		}
	}
	pdf.Ln(5)

	// Table headers
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(15, 8, "#", "1", 0, "C", true, 0, "")
	pdf.CellFormat(65, 8, "Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(20, 8, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(45, 8, "Unit Price", "1", 0, "R", true, 0, "")
	pdf.CellFormat(45, 8, "Total", "1", 0, "R", true, 0, "")
	pdf.Ln(8)

	// Table data
	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(255, 255, 255)
	
	subtotal := 0.0
	for i, item := range sale.SaleItems {
		// Check if we need a new page
		if pdf.GetY() > 250 {
			pdf.AddPage()
			// Re-add headers
			pdf.SetFont("Arial", "B", 10)
			pdf.SetFillColor(220, 220, 220)
			pdf.CellFormat(15, 8, "#", "1", 0, "C", true, 0, "")
			pdf.CellFormat(65, 8, "Description", "1", 0, "L", true, 0, "")
			pdf.CellFormat(20, 8, "Qty", "1", 0, "C", true, 0, "")
			pdf.CellFormat(45, 8, "Unit Price", "1", 0, "R", true, 0, "")
			pdf.CellFormat(45, 8, "Total", "1", 0, "R", true, 0, "")
			pdf.Ln(8)
			pdf.SetFont("Arial", "", 9)
			pdf.SetFillColor(255, 255, 255)
		}

		// Item data
		itemNumber := strconv.Itoa(i + 1)
		description := "Product"
		if item.Product.ID != 0 {
			description = item.Product.Name
		}

		quantity := strconv.Itoa(int(item.Quantity))
		unitPrice := p.formatRupiah(item.UnitPrice)
		totalPrice := p.formatRupiah(item.TotalPrice)
		
		pdf.CellFormat(15, 6, itemNumber, "1", 0, "C", false, 0, "")
		pdf.CellFormat(65, 6, description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 6, quantity, "1", 0, "C", false, 0, "")
		pdf.CellFormat(45, 6, unitPrice, "1", 0, "R", false, 0, "")
		pdf.CellFormat(45, 6, totalPrice, "1", 0, "R", false, 0, "")
		pdf.Ln(6)

		subtotal += item.TotalPrice
	}

	// Summary section
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 10)
	
	// Subtotal
	pdf.Cell(120, 6, "")
	pdf.Cell(25, 6, "Subtotal:")
	pdf.Cell(45, 6, p.formatRupiah(subtotal))
	pdf.Ln(6)

	// Discount
	if sale.DiscountPercent > 0 {
		discountAmount := subtotal * sale.DiscountPercent / 100
		pdf.Cell(120, 6, "")
		pdf.Cell(25, 6, fmt.Sprintf("Discount (%.1f%%):", sale.DiscountPercent))
		pdf.Cell(45, 6, "-" + p.formatRupiah(discountAmount))
		pdf.Ln(6)
	}

	// Taxes
	if sale.PPNPercent > 0 {
		ppnAmount := (subtotal - (subtotal * sale.DiscountPercent / 100)) * sale.PPNPercent / 100
		pdf.Cell(120, 6, "")
		pdf.Cell(25, 6, fmt.Sprintf("PPN (%.1f%%):", sale.PPNPercent))
		pdf.Cell(45, 6, p.formatRupiah(ppnAmount))
		pdf.Ln(6)
	}

	if sale.PPhPercent > 0 {
		pphAmount := (subtotal - (subtotal * sale.DiscountPercent / 100)) * sale.PPhPercent / 100
		pdf.Cell(120, 6, "")
		pdf.Cell(25, 6, fmt.Sprintf("PPh (%.1f%%):", sale.PPhPercent))
		pdf.Cell(45, 6, "-" + p.formatRupiah(pphAmount))
		pdf.Ln(6)
	}

	// Shipping
	if sale.ShippingCost > 0 {
		pdf.Cell(120, 6, "")
		pdf.Cell(25, 6, "Shipping:")
		pdf.Cell(45, 6, p.formatRupiah(sale.ShippingCost))
		pdf.Ln(6)
	}

	// Total
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(120, 8, "")
	pdf.Cell(25, 8, "TOTAL:")
	pdf.Cell(45, 8, p.formatRupiah(sale.TotalAmount))
	pdf.Ln(10)

	// Payment info
	if sale.PaymentTerms != "" {
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(190, 5, fmt.Sprintf("Payment Terms: %s", sale.PaymentTerms))
		pdf.Ln(5)
	}

	// Notes
	if sale.Notes != "" {
		pdf.Ln(5)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(190, 6, "Notes:")
		pdf.Ln(6)
		pdf.SetFont("Arial", "", 9)
		pdf.MultiCell(190, 4, sale.Notes, "", "", false)
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 4, fmt.Sprintf("Generated on %s", time.Now().Format("02/01/2006 15:04")))

	// Output to buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// GenerateSalesReportPDF generates a PDF for sales report
func (p *PDFService) GenerateSalesReportPDF(sales []models.Sale, startDate, endDate string) ([]byte, error) {
	// Create new PDF document
	pdf := gofpdf.New("L", "mm", "A4", "") // Landscape orientation
	// Standard consistent margins and pagebreak to avoid overflow
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Try adding company letterhead/logo (compact, top-left)
	p.addCompanyLetterhead(pdf)

	// Calculate usable width and header text area to the right of logo
	lm, tm, rm, _ := pdf.GetMargins()
	pageW, pageH := pdf.GetPageSize()
	contentW := pageW - lm - rm
	logoW := 40.0 // landscape default logo width used in header
	xStart := lm + logoW + 6
	textW := pageW - rm - xStart

	// Title placed to the right of the logo
	pdf.SetXY(xStart, tm+2)
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(textW, 10, "SALES REPORT")
	pdf.Ln(10)
	
	// Date range (to the right of logo)
	pdf.SetX(xStart)
	pdf.SetFont("Arial", "", 12)
	if startDate != "" && endDate != "" {
		pdf.Cell(textW, 6, fmt.Sprintf("Period: %s to %s", startDate, endDate))
	} else {
		pdf.Cell(textW, 6, "Period: All Time")
	}
	pdf.Ln(10)
	
	// Report generated info
	pdf.SetX(xStart)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(textW, 5, fmt.Sprintf("Generated on: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.Ln(10)
	
	// Ensure we start the table below the logo area
	if pdf.GetY() < tm+logoW+6 {
		pdf.SetY(tm + logoW + 6)
	}
	pdf.SetX(lm)
	
	// Column widths scaled to content width (keeps margins clean)
	// Tune proportions to ensure long invoice numbers fit nicely
	base := []float64{18, 22, 30, 36, 16, 16, 42, 42, 42}
	var baseSum float64
	for _, b := range base { baseSum += b }
	widths := make([]float64, len(base))
	var accum float64
	for i, b := range base {
		if i == len(base)-1 {
			widths[i] = contentW - accum // avoid rounding drift
		} else {
			w := b * contentW / baseSum
			widths[i] = w
			accum += w
		}
	}
	
	drawHeader := func() {
		pdf.SetFont("Arial", "B", 8)
		pdf.SetFillColor(220, 220, 220)
		pdf.CellFormat(widths[0], 8, "Date", "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[1], 8, "Sale Code", "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[2], 8, "Invoice No.", "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[3], 8, "Customer", "1", 0, "L", true, 0, "")
		pdf.CellFormat(widths[4], 8, "Type", "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[5], 8, "Status", "1", 0, "C", true, 0, "")
		pdf.CellFormat(widths[6], 8, "Amount", "1", 0, "R", true, 0, "")
		pdf.CellFormat(widths[7], 8, "Paid", "1", 0, "R", true, 0, "")
		pdf.CellFormat(widths[8], 8, "Outstanding", "1", 0, "R", true, 0, "")
		pdf.Ln(8)
	}
	
	// Render header initially
	drawHeader()
	
	// Table data
	pdf.SetFont("Arial", "", 8)
	pdf.SetFillColor(255, 255, 255)
	
	totalAmount := 0.0
	totalPaid := 0.0
	totalOutstanding := 0.0
	
	for _, sale := range sales {
		// Add new page when near bottom and redraw header
		if pdf.GetY() > pageH-25 {
			pdf.AddPage()
			p.addCompanyLetterhead(pdf)
			drawHeader()
		}
		// Sale data
		date := sale.Date.Format("02/01/06")
		customerName := "N/A"
		if sale.Customer.ID != 0 {
			customerName = sale.Customer.Name
			if len(customerName) > 30 {
				customerName = customerName[:27] + "..."
			}
		}
		invoiceNumber := sale.InvoiceNumber
		if invoiceNumber == "" { invoiceNumber = "-" }
		amount := p.formatRupiah(sale.TotalAmount)
		paid := p.formatRupiah(sale.PaidAmount)
		outstanding := p.formatRupiah(sale.OutstandingAmount)
		
		pdf.CellFormat(widths[0], 5, date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[1], 5, sale.Code, "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[2], 5, invoiceNumber, "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[3], 5, customerName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[4], 5, sale.Type, "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[5], 5, sale.Status, "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[6], 5, amount, "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[7], 5, paid, "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[8], 5, outstanding, "1", 0, "R", false, 0, "")
		pdf.Ln(5)
		
		// Accumulate totals
		totalAmount += sale.TotalAmount
		totalPaid += sale.PaidAmount
		totalOutstanding += sale.OutstandingAmount
	}
	
	// Summary section
	pdf.Ln(3)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(240, 240, 240)
	leftGroup := widths[0] + widths[1] + widths[2] + widths[3] + widths[4] + widths[5]
	pdf.CellFormat(leftGroup, 6, "TOTAL", "1", 0, "R", true, 0, "")
	pdf.CellFormat(widths[6], 6, p.formatRupiah(totalAmount), "1", 0, "R", true, 0, "")
	pdf.CellFormat(widths[7], 6, p.formatRupiah(totalPaid), "1", 0, "R", true, 0, "")
	pdf.CellFormat(widths[8], 6, p.formatRupiah(totalOutstanding), "1", 0, "R", true, 0, "")
	
	// Statistics
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(contentW, 6, "SUMMARY STATISTICS")
	pdf.Ln(6)
	
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(contentW/2, 5, fmt.Sprintf("Total Sales: %d", len(sales)))
	pdf.Cell(contentW/2, 5, fmt.Sprintf("Total Amount: %s", p.formatRupiah(totalAmount)))
	pdf.Ln(5)
	pdf.Cell(contentW/2, 5, fmt.Sprintf("Total Paid: %s", p.formatRupiah(totalPaid)))
	pdf.Cell(contentW/2, 5, fmt.Sprintf("Total Outstanding: %s", p.formatRupiah(totalOutstanding)))
	pdf.Ln(5)
	if len(sales) > 0 {
		avgAmount := totalAmount / float64(len(sales))
		pdf.Cell(contentW/2, 5, fmt.Sprintf("Average Sale Amount: %s", p.formatRupiah(avgAmount)))
	}

	// Output to buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate sales report PDF: %v", err)
	}
	return buf.Bytes(), nil
}

// GenerateGeneralLedgerPDF generates PDF for general ledger report
func (p *PDFService) GenerateGeneralLedgerPDF(ledgerData interface{}, accountInfo string, startDate, endDate string) ([]byte, error) {
	// Create new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	// Adjust margins/padding for consistent layout
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.SetCellMargin(1.5)
	pdf.AddPage()

	// Try adding company letterhead/logo
	p.addCompanyLetterhead(pdf)

	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Report header
	pdf.Cell(180, 10, "GENERAL LEDGER REPORT")
	pdf.Ln(15)

	// Get company info from settings
	companyInfo, err := p.getCompanyInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get company info: %v", err)
	}
	
	// Company info from settings
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(180, 8, companyInfo.CompanyName)
	pdf.SetFont("Arial", "", 10)
	pdf.Ln(6)
	pdf.Cell(180, 5, companyInfo.CompanyAddress)
	pdf.Ln(5)
	pdf.Cell(180, 5, fmt.Sprintf("Phone: %s", companyInfo.CompanyPhone))
	pdf.Ln(5)
	pdf.Cell(180, 5, fmt.Sprintf("Email: %s", companyInfo.CompanyEmail))
	pdf.Ln(10)

	// Report details
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(90, 6, fmt.Sprintf("Account: %s", accountInfo))
	pdf.Cell(90, 6, fmt.Sprintf("Period: %s to %s", startDate, endDate))
	pdf.Ln(6)
	pdf.Cell(180, 6, fmt.Sprintf("Generated: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.Ln(10)

	// Table headers (fit 180mm content width)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(20, 8, "Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, "Reference", "1", 0, "C", true, 0, "")
	pdf.CellFormat(70, 8, "Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(22, 8, "Debit", "1", 0, "R", true, 0, "")
	pdf.CellFormat(22, 8, "Credit", "1", 0, "R", true, 0, "")
	pdf.CellFormat(21, 8, "Balance", "1", 0, "R", true, 0, "")
	pdf.Ln(8)

	// Process ledger data based on its structure
	pdf.SetFont("Arial", "", 8)
	pdf.SetFillColor(255, 255, 255)
	
	// Normalize input to map[string]interface{} so we can iterate regardless of struct/map input
	var ledgerMap map[string]interface{}
	if m, ok := ledgerData.(map[string]interface{}); ok {
		ledgerMap = m
	} else {
		b, _ := json.Marshal(ledgerData)
		_ = json.Unmarshal(b, &ledgerMap)
	}
	
	if ledgerMap != nil {
		// Handle different possible data structures
		if accounts, exists := ledgerMap["accounts"]; exists {
			// Multiple accounts structure
			if accountsSlice, ok := accounts.([]interface{}); ok {
				for _, account := range accountsSlice {
					if accountMap, ok := account.(map[string]interface{}); ok {
						p.addAccountToLedgerPDF(pdf, accountMap)
					}
				}
			}
		} else if entries, exists := ledgerMap["entries"]; exists {
			// Single account structure
			if entriesSlice, ok := entries.([]interface{}); ok {
				p.addEntriesToLedgerPDF(pdf, entriesSlice, 0.0)
			}
		} else {
			// Fallback: treat entire data as single account
			p.addAccountToLedgerPDF(pdf, ledgerMap)
		}
	} else {
		// Handle simple data structure
	pdf.Cell(180, 6, "No ledger data available")
	pdf.Ln(6)
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(180, 4, fmt.Sprintf("Report generated on %s", time.Now().Format("02/01/2006 15:04")))

	// Output to buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate general ledger PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// addAccountToLedgerPDF adds account data to the PDF
func (p *PDFService) addAccountToLedgerPDF(pdf *gofpdf.Fpdf, accountData map[string]interface{}) {
	accountName := "All Accounts"
	accountCode := ""
	
	if name, exists := accountData["account_name"]; exists {
		if nameStr, ok := name.(string); ok && strings.TrimSpace(nameStr) != "" {
			accountName = nameStr
		}
	}
	if code, exists := accountData["account_code"]; exists {
		if codeStr, ok := code.(string); ok {
			accountCode = codeStr
		}
	}
	
	// Account header
	headerText := accountName
	if accountCode != "" {
		headerText = fmt.Sprintf("%s - %s", accountCode, accountName)
	}
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(180, 8, headerText, "1", 0, "L", true, 0, "")
	pdf.Ln(8)
	
	// Get opening balance
	openingBalance := 0.0
	if opening, exists := accountData["opening_balance"]; exists {
		if openingFloat, ok := opening.(float64); ok {
			openingBalance = openingFloat
		}
	}
	
	// Add opening balance row
	pdf.SetFont("Arial", "I", 8)
	pdf.SetFillColor(250, 250, 250)
	pdf.CellFormat(20, 6, "-", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 6, "-", "1", 0, "C", true, 0, "")
	pdf.CellFormat(70, 6, "Opening Balance", "1", 0, "L", true, 0, "")
	pdf.CellFormat(22, 6, "-", "1", 0, "R", true, 0, "")
	pdf.CellFormat(22, 6, "-", "1", 0, "R", true, 0, "")
	pdf.CellFormat(21, 6, p.formatRupiah(openingBalance), "1", 0, "R", true, 0, "")
	pdf.Ln(6)
	
	// Add entries - support both "entries" and "transactions" keys
	if entries, exists := accountData["entries"]; exists {
		if entriesSlice, ok := entries.([]interface{}); ok {
			p.addEntriesToLedgerPDF(pdf, entriesSlice, openingBalance)
		}
	} else if txs, exists := accountData["transactions"]; exists {
		if entriesSlice, ok := txs.([]interface{}); ok {
			p.addEntriesToLedgerPDF(pdf, entriesSlice, openingBalance)
		}
	}
	
	pdf.Ln(5)
}

// addEntriesToLedgerPDF adds individual entries to the PDF
func (p *PDFService) addEntriesToLedgerPDF(pdf *gofpdf.Fpdf, entries []interface{}, openingBalance float64) {
	pdf.SetFont("Arial", "", 8)
	pdf.SetFillColor(255, 255, 255)
	
	runningBalance := openingBalance
	
	for _, entry := range entries {
		if entryMap, ok := entry.(map[string]interface{}); ok {
			// Extract entry data
			date := "-"
			ref := "-"
			desc := "-"
			debit := 0.0
			credit := 0.0
			
			if dateVal, exists := entryMap["date"]; exists {
				if dateStr, ok := dateVal.(string); ok {
					if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
						date = parsedDate.Format("02/01")
					} else {
						date = dateStr[:10] // Take first 10 chars
					}
				}
			}
			
			if refVal, exists := entryMap["reference"]; exists {
				if refStr, ok := refVal.(string); ok {
					ref = refStr
					if len(ref) > 20 {
						ref = ref[:20] + "..."
					}
				}
			}
			
			if descVal, exists := entryMap["description"]; exists {
				if descStr, ok := descVal.(string); ok {
					desc = descStr
					if len(desc) > 40 {
						desc = desc[:40] + "..."
					}
				}
			}
			
			// Debit amount: support keys "debit" and "debit_amount" and string numbers
			if debitVal, exists := entryMap["debit"]; exists {
				if v, ok := debitVal.(float64); ok { debit = v } else if s, ok := debitVal.(string); ok {
					if f, err := strconv.ParseFloat(s, 64); err == nil { debit = f }
				}
			} else if debitVal, exists := entryMap["debit_amount"]; exists {
				if v, ok := debitVal.(float64); ok { debit = v } else if s, ok := debitVal.(string); ok {
					if f, err := strconv.ParseFloat(s, 64); err == nil { debit = f }
				}
			}
			
			// Credit amount: support keys "credit" and "credit_amount" and string numbers
			if creditVal, exists := entryMap["credit"]; exists {
				if v, ok := creditVal.(float64); ok { credit = v } else if s, ok := creditVal.(string); ok {
					if f, err := strconv.ParseFloat(s, 64); err == nil { credit = f }
				}
			} else if creditVal, exists := entryMap["credit_amount"]; exists {
				if v, ok := creditVal.(float64); ok { credit = v } else if s, ok := creditVal.(string); ok {
					if f, err := strconv.ParseFloat(s, 64); err == nil { credit = f }
				}
			}
			
			// Calculate running balance
			runningBalance += debit - credit
			
			// Add row to PDF
	pdf.CellFormat(20, 6, date, "1", 0, "C", false, 0, "")
	pdf.CellFormat(25, 6, ref, "1", 0, "L", false, 0, "")
	pdf.CellFormat(70, 6, desc, "1", 0, "L", false, 0, "")
	
	if debit > 0 {
		pdf.CellFormat(22, 6, p.formatRupiah(debit), "1", 0, "R", false, 0, "")
	} else {
		pdf.CellFormat(22, 6, "-", "1", 0, "R", false, 0, "")
	}
	
	if credit > 0 {
		pdf.CellFormat(22, 6, p.formatRupiah(credit), "1", 0, "R", false, 0, "")
	} else {
		pdf.CellFormat(22, 6, "-", "1", 0, "R", false, 0, "")
	}
	
	pdf.CellFormat(21, 6, p.formatRupiah(runningBalance), "1", 0, "R", false, 0, "")
			pdf.Ln(6)
			
			// Check if we need a new page
			if pdf.GetY() > 260 {
				pdf.AddPage()
				// Re-add headers
				pdf.SetFont("Arial", "B", 9)
				pdf.SetFillColor(220, 220, 220)
				pdf.CellFormat(20, 8, "Date", "1", 0, "C", true, 0, "")
				pdf.CellFormat(25, 8, "Reference", "1", 0, "C", true, 0, "")
				pdf.CellFormat(70, 8, "Description", "1", 0, "L", true, 0, "")
				pdf.CellFormat(22, 8, "Debit", "1", 0, "R", true, 0, "")
				pdf.CellFormat(22, 8, "Credit", "1", 0, "R", true, 0, "")
				pdf.CellFormat(21, 8, "Balance", "1", 0, "R", true, 0, "")
				pdf.Ln(8)
				pdf.SetFont("Arial", "", 8)
				pdf.SetFillColor(255, 255, 255)
			}
		}
	}
}

// GeneratePaymentReportPDF generates a PDF for payments report
func (p *PDFService) GeneratePaymentReportPDF(payments []models.Payment, startDate, endDate string) ([]byte, error) {
	// Create new PDF document
	pdf := gofpdf.New("L", "mm", "A4", "") // Landscape orientation
	pdf.AddPage()

	// Try adding company letterhead/logo
	p.addCompanyLetterhead(pdf)

	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Title
	pdf.Cell(270, 10, "PAYMENT REPORT")
	pdf.Ln(10)

	// Date range
	pdf.SetFont("Arial", "", 12)
	if startDate != "" && endDate != "" {
		pdf.Cell(270, 6, fmt.Sprintf("Period: %s to %s", startDate, endDate))
	} else {
		pdf.Cell(270, 6, "Period: All Time")
	}
	pdf.Ln(10)

	// Report generated info
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(270, 5, fmt.Sprintf("Generated on: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.Ln(10)

	// Table headers
	pdf.SetFont("Arial", "B", 8)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(18, 8, "Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, "Payment Code", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Contact", "1", 0, "L", true, 0, "")
	pdf.CellFormat(20, 8, "Method", "1", 0, "C", true, 0, "")
	pdf.CellFormat(44, 8, "Amount", "1", 0, "R", true, 0, "")
	pdf.CellFormat(18, 8, "Status", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, "Reference", "1", 0, "L", true, 0, "")
	pdf.CellFormat(55, 8, "Notes", "1", 0, "L", true, 0, "")
	pdf.Ln(8)

	// Table data
	pdf.SetFont("Arial", "", 8)
	pdf.SetFillColor(255, 255, 255)
	
	totalAmount := 0.0
	completedCount := 0
	pendingCount := 0
	failedCount := 0

	for _, payment := range payments {
		// Check if we need a new page
		if pdf.GetY() > 180 {
			pdf.AddPage()
			// Re-add headers
			pdf.SetFont("Arial", "B", 8)
			pdf.SetFillColor(220, 220, 220)
			pdf.CellFormat(18, 8, "Date", "1", 0, "C", true, 0, "")
			pdf.CellFormat(25, 8, "Payment Code", "1", 0, "C", true, 0, "")
			pdf.CellFormat(40, 8, "Contact", "1", 0, "L", true, 0, "")
			pdf.CellFormat(20, 8, "Method", "1", 0, "C", true, 0, "")
			pdf.CellFormat(44, 8, "Amount", "1", 0, "R", true, 0, "")
			pdf.CellFormat(18, 8, "Status", "1", 0, "C", true, 0, "")
			pdf.CellFormat(30, 8, "Reference", "1", 0, "L", true, 0, "")
			pdf.CellFormat(55, 8, "Notes", "1", 0, "L", true, 0, "")
			pdf.Ln(8)
			pdf.SetFont("Arial", "", 8)
			pdf.SetFillColor(255, 255, 255)
		}

		// Payment data
		date := payment.Date.Format("02/01/06")
		contactName := "N/A"
		if payment.Contact.ID != 0 {
			contactName = payment.Contact.Name
			// Truncate if too long
			if len(contactName) > 25 {
				contactName = contactName[:22] + "..."
			}
		}

		method := payment.Method
		if len(method) > 12 {
			method = method[:9] + "..."
		}

		amount := p.formatRupiah(payment.Amount)
		status := payment.Status
		reference := payment.Reference
		if len(reference) > 20 {
			reference = reference[:17] + "..."
		}

		notes := payment.Notes
		if len(notes) > 35 {
			notes = notes[:32] + "..."
		}

		pdf.CellFormat(18, 5, date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 5, payment.Code, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 5, contactName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 5, method, "1", 0, "C", false, 0, "")
		pdf.CellFormat(44, 5, amount, "1", 0, "R", false, 0, "")
		pdf.CellFormat(18, 5, status, "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 5, reference, "1", 0, "L", false, 0, "")
		pdf.CellFormat(55, 5, notes, "1", 0, "L", false, 0, "")
		pdf.Ln(5)

		// Accumulate totals
		totalAmount += payment.Amount
		switch payment.Status {
		case "COMPLETED":
			completedCount++
		case "PENDING":
			pendingCount++
		case "FAILED":
			failedCount++
		}
	}

	// Summary section
	pdf.Ln(3)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(103, 6, "TOTAL", "1", 0, "R", true, 0, "")
	pdf.CellFormat(44, 6, p.formatRupiah(totalAmount), "1", 0, "R", true, 0, "")
	pdf.CellFormat(123, 6, fmt.Sprintf("Count: %d", len(payments)), "1", 0, "L", true, 0, "")

	// Statistics
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(270, 6, "SUMMARY STATISTICS")
	pdf.Ln(6)
	
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(135, 5, fmt.Sprintf("Total Payments: %d", len(payments)))
	pdf.Cell(135, 5, fmt.Sprintf("Total Amount: %s", p.formatRupiah(totalAmount)))
	pdf.Ln(5)
	pdf.Cell(135, 5, fmt.Sprintf("Completed: %d", completedCount))
	pdf.Cell(135, 5, fmt.Sprintf("Pending: %d", pendingCount))
	pdf.Ln(5)
	pdf.Cell(135, 5, fmt.Sprintf("Failed: %d", failedCount))
	
	if len(payments) > 0 {
		avgAmount := totalAmount / float64(len(payments))
		pdf.Cell(135, 5, fmt.Sprintf("Average Payment Amount: %s", p.formatRupiah(avgAmount)))
	}

	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate payment report PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// GeneratePaymentDetailPDF generates a PDF for a single payment detail
func (p *PDFService) GeneratePaymentDetailPDF(payment *models.Payment) ([]byte, error) {
	// Create new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Try adding company letterhead/logo
	p.addCompanyLetterhead(pdf)

	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Payment header removed (no title)
	pdf.Ln(6)

	// Get company info from settings
	companyInfo, err := p.getCompanyInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get company info: %v", err)
	}
	
	// Company info from settings
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(95, 8, companyInfo.CompanyName)
	pdf.SetFont("Arial", "", 10)
	pdf.Ln(6)
	pdf.Cell(95, 5, companyInfo.CompanyAddress)
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Phone: %s", companyInfo.CompanyPhone))
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Email: %s", companyInfo.CompanyEmail))
	pdf.Ln(10)

	// Payment details
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(95, 6, fmt.Sprintf("Payment Code: %s", payment.Code))
	pdf.Cell(95, 6, fmt.Sprintf("Date: %s", payment.Date.Format("02/01/2006")))
	pdf.Ln(6)
	pdf.Cell(95, 6, fmt.Sprintf("Method: %s", payment.Method))
	pdf.Cell(95, 6, fmt.Sprintf("Status: %s", payment.Status))
	pdf.Ln(6)
	if payment.Reference != "" {
		pdf.Cell(190, 6, fmt.Sprintf("Reference: %s", payment.Reference))
		pdf.Ln(6)
	}
	pdf.Ln(5)

	// Contact info
	pdf.SetFont("Arial", "B", 10)
	if payment.Contact.Type == "CUSTOMER" {
		pdf.Cell(190, 6, "Payment From:")
	} else {
		pdf.Cell(190, 6, "Payment To:")
	}
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 10)
	if payment.Contact.ID != 0 {
		pdf.Cell(190, 5, payment.Contact.Name)
		pdf.Ln(5)
		if payment.Contact.Address != "" {
			pdf.Cell(190, 5, payment.Contact.Address)
			pdf.Ln(5)
		}
		if payment.Contact.Phone != "" {
			pdf.Cell(190, 5, fmt.Sprintf("Phone: %s", payment.Contact.Phone))
			pdf.Ln(5)
		}
	}
	pdf.Ln(10)

	// Payment amount section
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(190, 10, "PAYMENT DETAILS", "1", 0, "C", true, 0, "")
	pdf.Ln(10)

	// Amount details
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(95, 8, "Amount:")
	pdf.Cell(95, 8, p.formatRupiah(payment.Amount))
	pdf.Ln(10)

	// Notes section
	if payment.Notes != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(190, 6, "Notes:")
		pdf.Ln(6)
		pdf.SetFont("Arial", "", 9)
		pdf.MultiCell(190, 4, payment.Notes, "", "", false)
		pdf.Ln(5)
	}

	// Signature section
	pdf.Ln(20)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(63, 5, "Prepared by:")
	pdf.Cell(64, 5, "")
	pdf.Cell(63, 5, "Approved by:")
	pdf.Ln(20)
	pdf.Cell(63, 5, "_____________________")
	pdf.Cell(64, 5, "")
	pdf.Cell(63, 5, "_____________________")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 8)
	pdf.Cell(63, 5, "Finance")
	pdf.Cell(64, 5, "")
	pdf.Cell(63, 5, "Manager")

	// Footer
	pdf.Ln(15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 4, fmt.Sprintf("Generated on %s", time.Now().Format("02/01/2006 15:04")))

	// Output to buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate payment detail PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// GenerateReceiptPDF generates PDF for a single purchase receipt
func (p *PDFService) GenerateReceiptPDF(receipt *models.PurchaseReceipt) ([]byte, error) {
	// Create new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Try adding company letterhead/logo
	p.addCompanyLetterhead(pdf)

	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Receipt header
	pdf.Cell(190, 10, "GOODS RECEIPT")
	pdf.Ln(15)

	// Get company info from settings
	companyInfo, err := p.getCompanyInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get company info: %v", err)
	}
	
	// Company info from settings
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(95, 8, companyInfo.CompanyName)
	pdf.SetFont("Arial", "", 10)
	pdf.Ln(6)
	pdf.Cell(95, 5, companyInfo.CompanyAddress)
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Phone: %s", companyInfo.CompanyPhone))
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Email: %s", companyInfo.CompanyEmail))
	pdf.Ln(10)

	// Receipt details
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(95, 6, fmt.Sprintf("Receipt Number: %s", receipt.ReceiptNumber))
	pdf.Cell(95, 6, fmt.Sprintf("Date: %s", receipt.ReceivedDate.Format("02/01/2006")))
	pdf.Ln(6)
	if receipt.Purchase.Code != "" {
		pdf.Cell(95, 6, fmt.Sprintf("Purchase Order: %s", receipt.Purchase.Code))
	}
	pdf.Cell(95, 6, fmt.Sprintf("Status: %s", receipt.Status))
	pdf.Ln(6)
	receiverName := ""
	if receipt.Receiver.FirstName != "" || receipt.Receiver.LastName != "" {
		receiverName = strings.TrimSpace(receipt.Receiver.FirstName + " " + receipt.Receiver.LastName)
	} else if receipt.Receiver.Username != "" {
		receiverName = receipt.Receiver.Username
	}
	
	if receiverName != "" {
		pdf.Cell(190, 6, fmt.Sprintf("Received By: %s", receiverName))
		pdf.Ln(6)
	}
	pdf.Ln(5)

	// Vendor info
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 6, "Vendor Information:")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 10)
	if receipt.Purchase.Vendor.ID != 0 {
		pdf.Cell(190, 5, receipt.Purchase.Vendor.Name)
		pdf.Ln(5)
		if receipt.Purchase.Vendor.Address != "" {
			pdf.Cell(190, 5, receipt.Purchase.Vendor.Address)
			pdf.Ln(5)
		}
		if receipt.Purchase.Vendor.Phone != "" {
			pdf.Cell(190, 5, fmt.Sprintf("Phone: %s", receipt.Purchase.Vendor.Phone))
			pdf.Ln(5)
		}
	}
	pdf.Ln(5)

	// Table headers
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(15, 8, "#", "1", 0, "C", true, 0, "")
	pdf.CellFormat(65, 8, "Product", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 8, "Ordered", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, "Received", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, "Condition", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Notes", "1", 0, "L", true, 0, "")
	pdf.Ln(8)

	// Table data
	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(255, 255, 255)
	
	for i, item := range receipt.ReceiptItems {
		// Check if we need a new page
		if pdf.GetY() > 250 {
			pdf.AddPage()
			// Re-add headers
			pdf.SetFont("Arial", "B", 10)
			pdf.SetFillColor(220, 220, 220)
			pdf.CellFormat(15, 8, "#", "1", 0, "C", true, 0, "")
			pdf.CellFormat(65, 8, "Product", "1", 0, "L", true, 0, "")
			pdf.CellFormat(25, 8, "Ordered", "1", 0, "C", true, 0, "")
			pdf.CellFormat(25, 8, "Received", "1", 0, "C", true, 0, "")
			pdf.CellFormat(25, 8, "Condition", "1", 0, "C", true, 0, "")
			pdf.CellFormat(35, 8, "Notes", "1", 0, "L", true, 0, "")
			pdf.Ln(8)
			pdf.SetFont("Arial", "", 9)
			pdf.SetFillColor(255, 255, 255)
		}

		// Item data
		itemNumber := strconv.Itoa(i + 1)
		productName := "Product"
		orderedQty := "0"
		if item.PurchaseItem.Product.ID != 0 {
			productName = item.PurchaseItem.Product.Name
			orderedQty = strconv.Itoa(item.PurchaseItem.Quantity)
		}

		receivedQty := strconv.Itoa(item.QuantityReceived)
		condition := item.Condition
		notes := item.Notes
		if len(notes) > 20 {
			notes = notes[:17] + "..."
		}
		
		pdf.CellFormat(15, 6, itemNumber, "1", 0, "C", false, 0, "")
		pdf.CellFormat(65, 6, productName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, orderedQty, "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, receivedQty, "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, condition, "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 6, notes, "1", 0, "L", false, 0, "")
		pdf.Ln(6)
	}

	// Notes section
	if receipt.Notes != "" {
		pdf.Ln(10)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(190, 6, "Receipt Notes:")
		pdf.Ln(6)
		pdf.SetFont("Arial", "", 9)
		pdf.MultiCell(190, 4, receipt.Notes, "", "", false)
	}

	// Signature section
	pdf.Ln(20)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(63, 5, "Received by:")
	pdf.Cell(64, 5, "")
	pdf.Cell(63, 5, "Verified by:")
	pdf.Ln(15)
	pdf.Cell(63, 5, "_____________________")
	pdf.Cell(64, 5, "")
	pdf.Cell(63, 5, "_____________________")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 8)
	receiverName = ""
	if receipt.Receiver.FirstName != "" || receipt.Receiver.LastName != "" {
		receiverName = strings.TrimSpace(receipt.Receiver.FirstName + " " + receipt.Receiver.LastName)
	} else if receipt.Receiver.Username != "" {
		receiverName = receipt.Receiver.Username
	}
	
	if receiverName != "" {
		pdf.Cell(63, 5, receiverName)
	} else {
		pdf.Cell(63, 5, "Warehouse Staff")
	}
	pdf.Cell(64, 5, "")
	pdf.Cell(63, 5, "Manager")

	// Footer
	pdf.Ln(15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 4, fmt.Sprintf("Generated on %s", time.Now().Format("02/01/2006 15:04")))

	// Output to buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate receipt PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// GenerateAllReceiptsPDF generates combined PDF for all receipts of a purchase
func (p *PDFService) GenerateAllReceiptsPDF(purchase *models.Purchase, receipts []models.PurchaseReceipt) ([]byte, error) {
	// Create new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Try adding company letterhead/logo
	p.addCompanyLetterhead(pdf)

	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Title
	pdf.Cell(190, 10, "PURCHASE RECEIPTS SUMMARY")
	pdf.Ln(15)

	// Get company info from settings
	companyInfo, err := p.getCompanyInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get company info: %v", err)
	}
	
	// Company info from settings
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(95, 8, companyInfo.CompanyName)
	pdf.SetFont("Arial", "", 10)
	pdf.Ln(6)
	pdf.Cell(95, 5, companyInfo.CompanyAddress)
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Phone: %s", companyInfo.CompanyPhone))
	pdf.Ln(10)

	// Purchase details
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(95, 6, fmt.Sprintf("Purchase Order: %s", purchase.Code))
	pdf.Cell(95, 6, fmt.Sprintf("Date: %s", purchase.Date.Format("02/01/2006")))
	pdf.Ln(6)
	pdf.Cell(95, 6, fmt.Sprintf("Vendor: %s", purchase.Vendor.Name))
	pdf.Cell(95, 6, fmt.Sprintf("Total Receipts: %d", len(receipts)))
	pdf.Ln(10)

	// Receipts summary table
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.Cell(190, 8, "RECEIPTS SUMMARY")
	pdf.Ln(8)
	
	pdf.CellFormat(15, 8, "#", "1", 0, "C", true, 0, "")
	pdf.CellFormat(45, 8, "Receipt Number", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, "Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Received By", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 8, "Status", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Items Count", "1", 0, "C", true, 0, "")
	pdf.Ln(8)

	// Receipts data
	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(255, 255, 255)

	for i, receipt := range receipts {
		itemNumber := strconv.Itoa(i + 1)
		receiptNumber := receipt.ReceiptNumber
		date := receipt.ReceivedDate.Format("02/01/06")
		receivedBy := "N/A"
		if receipt.Receiver.FirstName != "" || receipt.Receiver.LastName != "" {
			receivedBy = strings.TrimSpace(receipt.Receiver.FirstName + " " + receipt.Receiver.LastName)
		} else if receipt.Receiver.Username != "" {
			receivedBy = receipt.Receiver.Username
		}
		if len(receivedBy) > 25 {
			receivedBy = receivedBy[:22] + "..."
		}
		status := receipt.Status
		itemsCount := strconv.Itoa(len(receipt.ReceiptItems))

		pdf.CellFormat(15, 6, itemNumber, "1", 0, "C", false, 0, "")
		pdf.CellFormat(45, 6, receiptNumber, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 6, receivedBy, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, status, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 6, itemsCount, "1", 0, "C", false, 0, "")
		pdf.Ln(6)
	}

	// Add each receipt as separate page
	for _, receipt := range receipts {
		pdf.AddPage()
		
		// Generate individual receipt content (simplified version)
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(190, 10, fmt.Sprintf("Receipt: %s", receipt.ReceiptNumber))
		pdf.Ln(10)
		
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(95, 6, fmt.Sprintf("Date: %s", receipt.ReceivedDate.Format("02/01/2006")))
		pdf.Cell(95, 6, fmt.Sprintf("Status: %s", receipt.Status))
		pdf.Ln(6)
		receiverName := ""
		if receipt.Receiver.FirstName != "" || receipt.Receiver.LastName != "" {
			receiverName = strings.TrimSpace(receipt.Receiver.FirstName + " " + receipt.Receiver.LastName)
		} else if receipt.Receiver.Username != "" {
			receiverName = receipt.Receiver.Username
		}
		
		if receiverName != "" {
			pdf.Cell(190, 6, fmt.Sprintf("Received By: %s", receiverName))
			pdf.Ln(6)
		}
		pdf.Ln(5)

		// Items table for this receipt
		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(220, 220, 220)
		pdf.CellFormat(15, 7, "#", "1", 0, "C", true, 0, "")
		pdf.CellFormat(60, 7, "Product", "1", 0, "L", true, 0, "")
		pdf.CellFormat(20, 7, "Ordered", "1", 0, "C", true, 0, "")
		pdf.CellFormat(20, 7, "Received", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 7, "Condition", "1", 0, "C", true, 0, "")
		pdf.CellFormat(50, 7, "Notes", "1", 0, "L", true, 0, "")
		pdf.Ln(7)

		pdf.SetFont("Arial", "", 8)
		pdf.SetFillColor(255, 255, 255)
		
		for j, item := range receipt.ReceiptItems {
			itemNumber := strconv.Itoa(j + 1)
			productName := "Product"
			orderedQty := "0"
			if item.PurchaseItem.Product.ID != 0 {
				productName = item.PurchaseItem.Product.Name
				if len(productName) > 35 {
					productName = productName[:32] + "..."
				}
				orderedQty = strconv.Itoa(item.PurchaseItem.Quantity)
			}

			receivedQty := strconv.Itoa(item.QuantityReceived)
			condition := item.Condition
			notes := item.Notes
			if len(notes) > 30 {
				notes = notes[:27] + "..."
			}

			pdf.CellFormat(15, 5, itemNumber, "1", 0, "C", false, 0, "")
			pdf.CellFormat(60, 5, productName, "1", 0, "L", false, 0, "")
			pdf.CellFormat(20, 5, orderedQty, "1", 0, "C", false, 0, "")
			pdf.CellFormat(20, 5, receivedQty, "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 5, condition, "1", 0, "C", false, 0, "")
			pdf.CellFormat(50, 5, notes, "1", 0, "L", false, 0, "")
			pdf.Ln(5)
		}

		// Notes for this receipt
		if receipt.Notes != "" {
			pdf.Ln(5)
			pdf.SetFont("Arial", "B", 9)
			pdf.Cell(190, 5, "Notes:")
			pdf.Ln(5)
			pdf.SetFont("Arial", "", 8)
			pdf.MultiCell(190, 4, receipt.Notes, "", "", false)
		}
	}

	// Final summary page
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 10, "COMPLETION SUMMARY")
	pdf.Ln(15)

	// Calculate completion statistics
	totalItems := len(purchase.PurchaseItems)
	totalReceiptItems := 0
	totalReceived := 0
	totalOrdered := 0

	for _, item := range purchase.PurchaseItems {
		totalOrdered += item.Quantity
	}

	for _, receipt := range receipts {
		totalReceiptItems += len(receipt.ReceiptItems)
		for _, item := range receipt.ReceiptItems {
			totalReceived += item.QuantityReceived
		}
	}

	completionRate := float64(totalReceived) / float64(totalOrdered) * 100

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(95, 6, fmt.Sprintf("Purchase Items: %d", totalItems))
	pdf.Cell(95, 6, fmt.Sprintf("Total Ordered: %d", totalOrdered))
	pdf.Ln(6)
	pdf.Cell(95, 6, fmt.Sprintf("Total Receipts: %d", len(receipts)))
	pdf.Cell(95, 6, fmt.Sprintf("Total Received: %d", totalReceived))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Completion Rate: %.1f%%", completionRate))
	pdf.Ln(10)

	// Footer
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 4, fmt.Sprintf("Generated on %s", time.Now().Format("02/01/2006 15:04")))

	// Output to buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate combined receipts PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// ========================================
// Phase 2: Priority Financial Reports PDF Export
// ========================================

// GenerateTrialBalancePDF generates PDF for trial balance
func (p *PDFService) GenerateTrialBalancePDF(trialBalanceData interface{}, asOfDate string) ([]byte, error) {
	// Create new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Try adding company letterhead/logo
	p.addCompanyLetterhead(pdf)

	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Report header
	pdf.Cell(190, 10, "TRIAL BALANCE")
	pdf.Ln(15)

	// Get company info from settings
	companyInfo, err := p.getCompanyInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get company info: %v", err)
	}
	
	// Company info from settings
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(95, 8, companyInfo.CompanyName)
	pdf.SetFont("Arial", "", 10)
	pdf.Ln(6)
	pdf.Cell(95, 5, companyInfo.CompanyAddress)
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Phone: %s", companyInfo.CompanyPhone))
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Email: %s", companyInfo.CompanyEmail))
	pdf.Ln(10)

	// Report details
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 6, fmt.Sprintf("As of: %s", asOfDate))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Generated: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.Ln(15)

	// Table headers
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(25, 8, "Account Code", "1", 0, "C", true, 0, "")
	pdf.CellFormat(75, 8, "Account Name", "1", 0, "L", true, 0, "")
	pdf.CellFormat(45, 8, "Debit Balance", "1", 0, "R", true, 0, "")
	pdf.CellFormat(45, 8, "Credit Balance", "1", 0, "R", true, 0, "")
	pdf.Ln(8)

	// Process trial balance data
	pdf.SetFont("Arial", "", 8)
	pdf.SetFillColor(255, 255, 255)

	totalDebits := 0.0
	totalCredits := 0.0

	// Normalize input to map[string]interface{} so we can iterate regardless of struct/map input
	var tbMap map[string]interface{}
	if m, ok := trialBalanceData.(map[string]interface{}); ok {
		tbMap = m
	} else {
		b, _ := json.Marshal(trialBalanceData)
		_ = json.Unmarshal(b, &tbMap)
	}

	if tbMap != nil {
		// Display accounts
		if accounts, exists := tbMap["accounts"]; exists {
			if accountsSlice, ok := accounts.([]interface{}); ok {
				for _, account := range accountsSlice {
					if accountMap, ok := account.(map[string]interface{}); ok {
						// Check if we need a new page
						if pdf.GetY() > 250 {
							pdf.AddPage()
							// Re-add headers
							pdf.SetFont("Arial", "B", 9)
							pdf.SetFillColor(220, 220, 220)
							pdf.CellFormat(25, 8, "Account Code", "1", 0, "C", true, 0, "")
							pdf.CellFormat(75, 8, "Account Name", "1", 0, "L", true, 0, "")
							pdf.CellFormat(45, 8, "Debit Balance", "1", 0, "R", true, 0, "")
							pdf.CellFormat(45, 8, "Credit Balance", "1", 0, "R", true, 0, "")
							pdf.Ln(8)
							pdf.SetFont("Arial", "", 8)
							pdf.SetFillColor(255, 255, 255)
						}

						accountCode := ""
						accountName := "Unknown Account"
						debitBalance := 0.0
						creditBalance := 0.0

						if code, exists := accountMap["account_code"]; exists {
							if codeStr, ok := code.(string); ok {
								accountCode = codeStr
							}
						}
						if name, exists := accountMap["account_name"]; exists {
							if nameStr, ok := name.(string); ok {
								accountName = nameStr
							}
						}
						if debit, exists := accountMap["debit_balance"]; exists {
							if debitFloat, ok := debit.(float64); ok {
								debitBalance = debitFloat
								totalDebits += debitBalance
							}
						}
						if credit, exists := accountMap["credit_balance"]; exists {
							if creditFloat, ok := credit.(float64); ok {
								creditBalance = creditFloat
								totalCredits += creditBalance
							}
						}

						// Truncate account name if too long
						if len(accountName) > 45 {
							accountName = accountName[:42] + "..."
						}

						pdf.CellFormat(25, 5, accountCode, "1", 0, "C", false, 0, "")
						pdf.CellFormat(75, 5, accountName, "1", 0, "L", false, 0, "")

						// Show debit balance or dash
						if debitBalance != 0 {
							pdf.CellFormat(45, 5, p.formatRupiah(debitBalance), "1", 0, "R", false, 0, "")
						} else {
							pdf.CellFormat(45, 5, "-", "1", 0, "R", false, 0, "")
						}

						// Show credit balance or dash
						if creditBalance != 0 {
							pdf.CellFormat(45, 5, p.formatRupiah(creditBalance), "1", 0, "R", false, 0, "")
						} else {
							pdf.CellFormat(45, 5, "-", "1", 0, "R", false, 0, "")
						}
						pdf.Ln(5)
					}
				}
			}
		}
	} else {
		// Fallback: simple data display
		pdf.Cell(190, 6, "Trial Balance data structure not recognized")
		pdf.Ln(6)
	}

	// Totals section
	pdf.Ln(3)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(100, 6, "TOTAL", "1", 0, "R", true, 0, "")
	pdf.CellFormat(45, 6, p.formatRupiah(totalDebits), "1", 0, "R", true, 0, "")
	pdf.CellFormat(45, 6, p.formatRupiah(totalCredits), "1", 0, "R", true, 0, "")
	pdf.Ln(8)

	// Balance verification
	isBalanced := (totalDebits == totalCredits)
	balanceStatus := "BALANCED"
	if !isBalanced {
		balanceStatus = "NOT BALANCED"
	}

	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.Cell(190, 6, fmt.Sprintf("BALANCE VERIFICATION: %s", balanceStatus))
	pdf.Ln(8)

	if !isBalanced {
		variance := totalDebits - totalCredits
		pdf.SetFont("Arial", "", 9)
		pdf.Cell(190, 5, fmt.Sprintf("Variance: %s", p.formatRupiah(variance)))
		pdf.Ln(5)
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 4, fmt.Sprintf("Report generated on %s", time.Now().Format("02/01/2006 15:04")))

	// Output to buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate trial balance PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// GenerateBalanceSheetPDF generates a PDF for Balance Sheet report
func (p *PDFService) GenerateBalanceSheetPDF(balanceSheetData interface{}, asOfDate string) ([]byte, error) {
	// Create new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Try adding company letterhead/logo
	p.addCompanyLetterhead(pdf)

	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Report header
	pdf.Cell(190, 10, "BALANCE SHEET")
	pdf.Ln(15)
	
	// Get company info from settings
	companyInfo, err := p.getCompanyInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get company info: %v", err)
	}
	
	// Company info from settings
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(95, 8, companyInfo.CompanyName)
	pdf.SetFont("Arial", "", 10)
	pdf.Ln(6)
	pdf.Cell(95, 5, companyInfo.CompanyAddress)
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Phone: %s", companyInfo.CompanyPhone))
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Email: %s", companyInfo.CompanyEmail))
	pdf.Ln(10)
	
	// Report details
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 6, fmt.Sprintf("As of: %s", asOfDate))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Generated: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.Ln(10)
	
	// Process balance sheet data
	if bsMap, ok := balanceSheetData.(map[string]interface{}); ok {
		p.addBalanceSheetSections(pdf, bsMap)
	} else {
		// Try to convert struct -> map[string]interface{} via JSON roundtrip
		if m := p.tryConvertBSDataToMap(balanceSheetData); m != nil {
			p.addBalanceSheetSections(pdf, m)
		} else {
			pdf.Cell(190, 6, "Balance Sheet data not available")
			pdf.Ln(6)
		}
	}
	
	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 4, fmt.Sprintf("Report generated on %s", time.Now().Format("02/01/2006 15:04")))
	
	// Output to buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate balance sheet PDF: %v", err)
	}
	
	return buf.Bytes(), nil
}

// tryConvertBSDataToMap attempts to convert a struct balance sheet into a generic map
func (p *PDFService) tryConvertBSDataToMap(data interface{}) map[string]interface{} {
	b, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}

// addBalanceSheetSections adds balance sheet sections to PDF
func (p *PDFService) addBalanceSheetSections(pdf *gofpdf.Fpdf, bsData map[string]interface{}) {
	pdf.SetFont("Arial", "B", 12)
	
	// Assets section
	if assets, exists := bsData["assets"]; exists {
		pdf.Cell(190, 8, "ASSETS")
		pdf.Ln(8)
		p.addAssetsSections(pdf, assets)
		pdf.Ln(5)
	}
	
	// Liabilities section
	if liabilities, exists := bsData["liabilities"]; exists {
		pdf.Cell(190, 8, "LIABILITIES")
		pdf.Ln(8)
		p.addLiabilitiesSections(pdf, liabilities)
		pdf.Ln(5)
	}
	
	// Equity section
	if equity, exists := bsData["equity"]; exists {
		pdf.Cell(190, 8, "EQUITY")
		pdf.Ln(8)
		p.addEquitySection(pdf, equity)
		pdf.Ln(5)
	}
	
	// Balance verification
	p.addBalanceVerification(pdf, bsData)
}

// addAssetsSections adds assets sections to balance sheet PDF
func (p *PDFService) addAssetsSections(pdf *gofpdf.Fpdf, assets interface{}) {
	pdf.SetFont("Arial", "B", 10)
	
	if assetsMap, ok := assets.(map[string]interface{}); ok {
		// Current Assets
		if currentAssets, exists := assetsMap["current_assets"]; exists {
			pdf.Cell(190, 6, "  Current Assets")
			pdf.Ln(6)
			p.addAccountItems(pdf, currentAssets, "    ")
			
			if total, exists := assetsMap["current_assets_total"]; exists {
				p.addTotalLine(pdf, "  Total Current Assets", total)
			}
		}
		
		// Non-Current Assets
		if nonCurrentAssets, exists := assetsMap["non_current_assets"]; exists {
			pdf.Ln(3)
			pdf.Cell(190, 6, "  Non-Current Assets")
			pdf.Ln(6)
			p.addAccountItems(pdf, nonCurrentAssets, "    ")
			
			if total, exists := assetsMap["non_current_assets_total"]; exists {
				p.addTotalLine(pdf, "  Total Non-Current Assets", total)
			}
		}
		
		// Total Assets
		if totalAssets, exists := assetsMap["total_assets"]; exists {
			pdf.Ln(3)
			pdf.SetFont("Arial", "B", 10)
			pdf.SetFillColor(240, 240, 240)
			pdf.CellFormat(145, 6, "TOTAL ASSETS", "1", 0, "L", true, 0, "")
			if totalFloat, ok := totalAssets.(float64); ok {
				pdf.CellFormat(45, 6, p.formatRupiah(totalFloat), "1", 0, "R", true, 0, "")
			} else {
				pdf.CellFormat(45, 6, fmt.Sprintf("%v", totalAssets), "1", 0, "R", true, 0, "")
			}
			pdf.Ln(6)
		}
	}
}

// addLiabilitiesSections adds liabilities sections to balance sheet PDF
func (p *PDFService) addLiabilitiesSections(pdf *gofpdf.Fpdf, liabilities interface{}) {
	pdf.SetFont("Arial", "B", 10)
	
	if liabilitiesMap, ok := liabilities.(map[string]interface{}); ok {
		// Current Liabilities
		if currentLiabilities, exists := liabilitiesMap["current_liabilities"]; exists {
			pdf.Cell(190, 6, "  Current Liabilities")
			pdf.Ln(6)
			p.addAccountItems(pdf, currentLiabilities, "    ")
			
			if total, exists := liabilitiesMap["current_liabilities_total"]; exists {
				p.addTotalLine(pdf, "  Total Current Liabilities", total)
			}
		}
		
		// Non-Current Liabilities
		if nonCurrentLiabilities, exists := liabilitiesMap["non_current_liabilities"]; exists {
			pdf.Ln(3)
			pdf.Cell(190, 6, "  Non-Current Liabilities")
			pdf.Ln(6)
			p.addAccountItems(pdf, nonCurrentLiabilities, "    ")
			
			if total, exists := liabilitiesMap["non_current_liabilities_total"]; exists {
				p.addTotalLine(pdf, "  Total Non-Current Liabilities", total)
			}
		}
		
		// Total Liabilities
		if totalLiabilities, exists := liabilitiesMap["total_liabilities"]; exists {
			pdf.Ln(3)
			pdf.SetFont("Arial", "B", 10)
			pdf.SetFillColor(240, 240, 240)
			pdf.CellFormat(145, 6, "TOTAL LIABILITIES", "1", 0, "L", true, 0, "")
			if totalFloat, ok := totalLiabilities.(float64); ok {
				pdf.CellFormat(45, 6, p.formatRupiah(totalFloat), "1", 0, "R", true, 0, "")
			} else {
				pdf.CellFormat(45, 6, fmt.Sprintf("%v", totalLiabilities), "1", 0, "R", true, 0, "")
			}
			pdf.Ln(6)
		}
	}
}

// addEquitySection adds equity section to balance sheet PDF
func (p *PDFService) addEquitySection(pdf *gofpdf.Fpdf, equity interface{}) {
	pdf.SetFont("Arial", "B", 10)
	
	if equityMap, ok := equity.(map[string]interface{}); ok {
		p.addAccountItems(pdf, equity, "  ")
		
		// Total Equity
		if totalEquity, exists := equityMap["total_equity"]; exists {
			pdf.Ln(3)
			pdf.SetFont("Arial", "B", 10)
			pdf.SetFillColor(240, 240, 240)
			pdf.CellFormat(145, 6, "TOTAL EQUITY", "1", 0, "L", true, 0, "")
			if totalFloat, ok := totalEquity.(float64); ok {
				pdf.CellFormat(45, 6, p.formatRupiah(totalFloat), "1", 0, "R", true, 0, "")
			} else {
				pdf.CellFormat(45, 6, fmt.Sprintf("%v", totalEquity), "1", 0, "R", true, 0, "")
			}
			pdf.Ln(6)
		}
	}
}

// addAccountItems adds account items to PDF with indentation
func (p *PDFService) addAccountItems(pdf *gofpdf.Fpdf, items interface{}, indent string) {
	pdf.SetFont("Arial", "", 9)

	// Helper to safely get string/float values from a map with multiple possible keys
	getString := func(m map[string]interface{}, keys ...string) string {
		for _, k := range keys {
			if v, ok := m[k]; ok {
				if s, ok := v.(string); ok && s != "" {
					return s
				}
			}
		}
		return ""
	}
	getFloat := func(m map[string]interface{}, keys ...string) float64 {
		for _, k := range keys {
			if v, ok := m[k]; ok {
				if f, ok := v.(float64); ok {
					return f
				}
			}
		}
		return 0.0
	}
	
	if itemsSlice, ok := items.([]interface{}); ok {
		for _, item := range itemsSlice {
			if itemMap, ok := item.(map[string]interface{}); ok {
				// Prefer "account_name" (our SSOT struct tag), fallback to generic "name"
				name := getString(itemMap, "account_name", "name", "AccountName")
				if name == "" {
					name = "Unknown Account"
				}
				// Include account code when available (e.g., "1101 - Kas")
				code := getString(itemMap, "account_code", "code", "AccountCode")
				label := name
				if code != "" {
					label = fmt.Sprintf("%s - %s", code, name)
				}
				// Prefer "amount", but also handle "balance" or "value" if present
				amount := getFloat(itemMap, "amount", "balance", "value")
				
				pdf.Cell(145, 5, fmt.Sprintf("%s%s", indent, label))
				pdf.Cell(45, 5, p.formatRupiah(amount))
				pdf.Ln(5)
			}
		}
	} else if itemsMap, ok := items.(map[string]interface{}); ok {
		// Handle map structure
		for key, value := range itemsMap {
			if key == "items" {
				p.addAccountItems(pdf, value, indent)
			}
		}
	}
}

// addTotalLine adds a total line to the PDF
func (p *PDFService) addTotalLine(pdf *gofpdf.Fpdf, label string, total interface{}) {
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(250, 250, 250)
	pdf.CellFormat(145, 5, label, "1", 0, "L", true, 0, "")
	
	if totalFloat, ok := total.(float64); ok {
		pdf.CellFormat(45, 5, p.formatRupiah(totalFloat), "1", 0, "R", true, 0, "")
	} else {
		pdf.CellFormat(45, 5, fmt.Sprintf("%v", total), "1", 0, "R", true, 0, "")
	}
	pdf.Ln(5)
}

// addBalanceVerification adds balance verification to balance sheet PDF
func (p *PDFService) addBalanceVerification(pdf *gofpdf.Fpdf, bsData map[string]interface{}) {
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	
	isBalanced := false
	balanceDifference := 0.0
	
	if balancedVal, exists := bsData["is_balanced"]; exists {
		if balancedBool, ok := balancedVal.(bool); ok {
			isBalanced = balancedBool
		}
	}
	
	if diffVal, exists := bsData["balance_difference"]; exists {
		if diffFloat, ok := diffVal.(float64); ok {
			balanceDifference = diffFloat
		}
	}
	
	balanceStatus := "BALANCED"
	if !isBalanced {
		balanceStatus = "NOT BALANCED"
	}
	
	pdf.CellFormat(190, 8, fmt.Sprintf("BALANCE VERIFICATION: %s", balanceStatus), "1", 0, "C", true, 0, "")
	pdf.Ln(8)
	
	if !isBalanced {
		pdf.SetFont("Arial", "", 9)
		pdf.Cell(190, 5, fmt.Sprintf("Balance Difference: %s", p.formatRupiah(balanceDifference)))
		pdf.Ln(5)
	}
}

// GenerateProfitLossPDF generates a PDF for Profit & Loss report
func (p *PDFService) GenerateProfitLossPDF(plData interface{}, startDate, endDate string) ([]byte, error) {
	// Create new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Try adding company letterhead/logo
	p.addCompanyLetterhead(pdf)

	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Report header
	pdf.Cell(190, 10, "PROFIT & LOSS STATEMENT")
	pdf.Ln(15)

	// Get company info from settings
	companyInfo, err := p.getCompanyInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get company info: %v", err)
	}
	
	// Company info from settings
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(95, 8, companyInfo.CompanyName)
	pdf.SetFont("Arial", "", 10)
	pdf.Ln(6)
	pdf.Cell(95, 5, companyInfo.CompanyAddress)
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Phone: %s", companyInfo.CompanyPhone))
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Email: %s", companyInfo.CompanyEmail))
	pdf.Ln(10)

	// Report details
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 6, fmt.Sprintf("Period: %s to %s", startDate, endDate))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Generated: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.Ln(10)

	// Process P&L data
	if plMap, ok := plData.(map[string]interface{}); ok {
		p.addProfitLossSections(pdf, plMap)
	} else {
		pdf.Cell(190, 6, "Profit & Loss data not available")
		pdf.Ln(6)
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 4, fmt.Sprintf("Report generated on %s", time.Now().Format("02/01/2006 15:04")))

	// Output to buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate profit & loss PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// GenerateJournalAnalysisPDF generates a PDF for Journal Entry Analysis report
func (p *PDFService) GenerateJournalAnalysisPDF(journalData interface{}, startDate, endDate string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Try adding company letterhead/logo
	p.addCompanyLetterhead(pdf)

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "JOURNAL ENTRY ANALYSIS")
	pdf.Ln(15)

	companyInfo, err := p.getCompanyInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get company info: %v", err)
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(95, 8, companyInfo.CompanyName)
	pdf.SetFont("Arial", "", 10)
	pdf.Ln(6)
	pdf.Cell(95, 5, companyInfo.CompanyAddress)
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Phone: %s", companyInfo.CompanyPhone))
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Email: %s", companyInfo.CompanyEmail))
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 6, fmt.Sprintf("Period: %s to %s", startDate, endDate))
	pdf.Ln(6)
	pdf.Cell(190, 6, fmt.Sprintf("Generated: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.Ln(10)

	// Normalize input to map[string]interface{} (structs are common here)
	var dataMap map[string]interface{}
	if m, ok := journalData.(map[string]interface{}); ok {
		dataMap = m
	} else {
		b, _ := json.Marshal(journalData)
		_ = json.Unmarshal(b, &dataMap)
	}

	if dataMap != nil {
		// Summary grid
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 8, "SUMMARY")
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 10)
		// Use light gray background for grid cells (avoid black fill)
		pdf.SetFillColor(230,230,230)
		
		getNum := func(key string) float64 {
			if v, exists := dataMap[key]; exists {
				switch t := v.(type) {
				case float64:
					return t
				case int:
					return float64(t)
				case int64:
					return float64(t)
				case json.Number:
					f, _ := t.Float64(); return f
				}
			}
			return 0
		}
		
		left := []struct{label string; value float64; currency bool}{
			{"Total Entries", getNum("total_entries"), false},
			{"Posted Entries", getNum("posted_entries"), false},
			{"Draft Entries", getNum("draft_entries"), false},
		}
		right := []struct{label string; value float64; currency bool}{
			{"Reversed Entries", getNum("reversed_entries"), false},
			{"Total Amount", getNum("total_amount"), true},
		}
		
		for i := 0; i < len(left) || i < len(right); i++ {
			if i < len(left) {
				pdf.CellFormat(60, 6, left[i].label, "1", 0, "L", true, 0, "")
				val := strconv.Itoa(int(left[i].value))
				if left[i].currency { val = p.formatRupiah(left[i].value) }
				pdf.CellFormat(35, 6, val, "1", 0, "R", true, 0, "")
			} else { pdf.Cell(95, 6, "") }
			if i < len(right) {
				pdf.CellFormat(60, 6, right[i].label, "1", 0, "L", true, 0, "")
				val := strconv.Itoa(int(right[i].value))
				if right[i].currency { val = p.formatRupiah(right[i].value) }
				pdf.CellFormat(35, 6, val, "1", 1, "R", true, 0, "")
			} else { pdf.Cell(95, 6, ""); pdf.Ln(6) }
		}

		pdf.Ln(8)
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 8, "ENTRIES BY TYPE")
		pdf.Ln(8)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(220,220,220)
		pdf.CellFormat(70, 8, "Source Type", "1", 0, "L", true, 0, "")
		pdf.CellFormat(40, 8, "Count", "1", 0, "C", true, 0, "")
		pdf.CellFormat(40, 8, "Amount", "1", 0, "R", true, 0, "")
		pdf.CellFormat(40, 8, "Percentage", "1", 1, "C", true, 0, "")
		pdf.SetFont("Arial", "", 9)
		
		if list, exists := dataMap["entries_by_type"]; exists {
			if items, ok := list.([]interface{}); ok {
				for _, it := range items {
					if m, ok := it.(map[string]interface{}); ok {
						sType := ""
						count := getNumFrom(m["count"]) 
						amount := getNumFrom(m["total_amount"]) 
						perc := getNumFrom(m["percentage"]) 
						if v, ok := m["source_type"].(string); ok { sType = v }
						pdf.CellFormat(70, 6, sType, "1", 0, "L", false, 0, "")
						pdf.CellFormat(40, 6, strconv.Itoa(int(count)), "1", 0, "C", false, 0, "")
						pdf.CellFormat(40, 6, p.formatRupiah(amount), "1", 0, "R", false, 0, "")
						pdf.CellFormat(40, 6, fmt.Sprintf("%.2f%%", perc), "1", 1, "C", false, 0, "")
					}
				}
			}
		}
	} else {
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(190, 6, "Journal Analysis data not available")
		pdf.Ln(10)
	}

	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 4, fmt.Sprintf("Report generated on %s", time.Now().Format("02/01/2006 15:04")))

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate journal analysis PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// getNumFrom converts various number representations to float64 (helper for PDF tables)
func getNumFrom(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case json.Number:
		f, _ := t.Float64(); return f
	case string:
		// Handle numbers encoded as strings (e.g., from decimal.Decimal JSON)
		if f, err := strconv.ParseFloat(t, 64); err == nil {
			return f
		}
		return 0
	default:
		return 0
	}
}


// renderSSOTFinancialSections renders the financial sections with extracted data
func (p *PDFService) renderSSOTFinancialSections(pdf *gofpdf.Fpdf, data map[string]interface{}) {
	// REVENUE SECTION
	if totalRevenue, exists := data["TotalRevenue"]; exists {
		if revenueFloat, ok := totalRevenue.(float64); ok && revenueFloat > 0 {
			p.addSSOTSection(pdf, "REVENUE", revenueFloat, "Revenue from sales and services")
		}
	}
	
	// COST OF GOODS SOLD SECTION
	if totalCOGS, exists := data["TotalCOGS"]; exists {
		if cogsFloat, ok := totalCOGS.(float64); ok && cogsFloat > 0 {
			p.addSSOTSection(pdf, "COST OF GOODS SOLD", cogsFloat, "Direct costs of producing goods/services")
		}
	}
	
	// GROSS PROFIT
	if grossProfit, exists := data["GrossProfit"]; exists {
		if gpFloat, ok := grossProfit.(float64); ok {
			p.addSSOTTotalLine(pdf, "GROSS PROFIT", gpFloat)
			
			// Add margin if available
			if margin, exists := data["GrossProfitMargin"]; exists {
				if marginFloat, ok := margin.(float64); ok && marginFloat > 0 {
					pdf.SetFont("Arial", "", 9)
					pdf.Cell(190, 5, fmt.Sprintf("Gross Profit Margin: %.2f%%", marginFloat))
					pdf.Ln(8)
				}
			}
		}
	}
	
	// OPERATING EXPENSES
	if totalOpEx, exists := data["TotalOpEx"]; exists {
		if opexFloat, ok := totalOpEx.(float64); ok && opexFloat > 0 {
			p.addSSOTSection(pdf, "OPERATING EXPENSES", opexFloat, "Administrative, selling, and general expenses")
		}
	}
	
	// OPERATING INCOME
	if operatingIncome, exists := data["OperatingIncome"]; exists {
		if oiFloat, ok := operatingIncome.(float64); ok {
			p.addSSOTTotalLine(pdf, "OPERATING INCOME", oiFloat)
			
			if margin, exists := data["OperatingMargin"]; exists {
				if marginFloat, ok := margin.(float64); ok && marginFloat > 0 {
					pdf.SetFont("Arial", "", 9)
					pdf.Cell(190, 5, fmt.Sprintf("Operating Margin: %.2f%%", marginFloat))
					pdf.Ln(8)
				}
			}
		}
	}
	
	// OTHER INCOME/EXPENSES
	p.addOtherIncomeExpenses(pdf, data)
	
	// INCOME BEFORE TAX
	if incomeBeforeTax, exists := data["IncomeBeforeTax"]; exists {
		if ibtFloat, ok := incomeBeforeTax.(float64); ok {
			p.addSSOTTotalLine(pdf, "INCOME BEFORE TAX", ibtFloat)
		}
	}
	
	// TAX EXPENSE
	if taxExpense, exists := data["TaxExpense"]; exists {
		if taxFloat, ok := taxExpense.(float64); ok && taxFloat != 0 {
			pdf.SetFont("Arial", "", 10)
			pdf.SetFillColor(250, 250, 250)
			pdf.CellFormat(140, 6, "Tax Expense", "1", 0, "L", true, 0, "")
			pdf.CellFormat(50, 6, p.formatRupiah(taxFloat), "1", 0, "R", true, 0, "")
			pdf.Ln(6)
		}
	}
	
	// NET INCOME (Final Result)
	if netIncome, exists := data["NetIncome"]; exists {
		if niFloat, ok := netIncome.(float64); ok {
			pdf.SetFont("Arial", "B", 14)
			pdf.SetFillColor(200, 200, 200)
			pdf.CellFormat(140, 10, "NET INCOME", "1", 0, "L", true, 0, "")
			pdf.CellFormat(50, 10, p.formatRupiah(niFloat), "1", 0, "R", true, 0, "")
			pdf.Ln(10)
			
			if margin, exists := data["NetIncomeMargin"]; exists {
				if marginFloat, ok := margin.(float64); ok && marginFloat > 0 {
					pdf.SetFont("Arial", "B", 10)
					pdf.Cell(190, 6, fmt.Sprintf("Net Income Margin: %.2f%%", marginFloat))
					pdf.Ln(8)
				}
			}
		}
	}
	
	// Add financial ratios summary
	p.addFinancialRatiosSummary(pdf, data)
}

// renderSSOTPlaceholder renders placeholder content when data extraction fails
func (p *PDFService) renderSSOTPlaceholder(pdf *gofpdf.Fpdf, ssotData interface{}) {
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "SSOT PROFIT & LOSS REPORT")
	pdf.Ln(8)
	
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 6, "This report is generated from the Single Source of Truth (SSOT) journal system.")
	pdf.Ln(6)
	pdf.Cell(190, 6, "Data includes revenue, expenses, and financial metrics from journal entries.")
	pdf.Ln(10)
	
	// Show that we received some data
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(190, 5, "Status: PDF generation successful - SSOT data structure received")
	pdf.Ln(5)
	pdf.Cell(190, 5, "Note: Detailed financial data parsing is in progress...")
	pdf.Ln(10)
	
	// Add some basic structure
	p.addPlaceholderSections(pdf)
}

// addSSOTSection adds a financial section with title and amount
func (p *PDFService) addSSOTSection(pdf *gofpdf.Fpdf, title string, amount float64, description string) {
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(220, 220, 220)
	pdf.Cell(190, 8, title)
	pdf.Ln(8)
	
	if description != "" {
		pdf.SetFont("Arial", "", 9)
		pdf.Cell(190, 5, description)
		pdf.Ln(5)
	}
	
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(140, 6, fmt.Sprintf("TOTAL %s", strings.ToUpper(title)), "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 6, p.formatRupiah(amount), "1", 0, "R", true, 0, "")
	pdf.Ln(8)
}

// addSSOTTotalLine adds a total line with highlighting
func (p *PDFService) addSSOTTotalLine(pdf *gofpdf.Fpdf, title string, amount float64) {
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(230, 230, 230)
	pdf.CellFormat(140, 8, title, "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 8, p.formatRupiah(amount), "1", 0, "R", true, 0, "")
	pdf.Ln(8)
}

// addOtherIncomeExpenses adds other income and expenses section
func (p *PDFService) addOtherIncomeExpenses(pdf *gofpdf.Fpdf, data map[string]interface{}) {
	otherIncome, hasIncome := data["OtherIncome"]
	otherExpenses, hasExpenses := data["OtherExpenses"]
	
	if hasIncome || hasExpenses {
		pdf.SetFont("Arial", "B", 12)
		pdf.SetFillColor(220, 220, 220)
		pdf.Cell(190, 8, "OTHER INCOME/EXPENSES")
		pdf.Ln(8)
		
		pdf.SetFont("Arial", "", 10)
		if hasIncome {
			if incomeFloat, ok := otherIncome.(float64); ok && incomeFloat != 0 {
				pdf.CellFormat(140, 5, "  Other Income", "1", 0, "L", false, 0, "")
				pdf.CellFormat(50, 5, p.formatRupiah(incomeFloat), "1", 0, "R", false, 0, "")
				pdf.Ln(5)
			}
		}
		
		if hasExpenses {
			if expensesFloat, ok := otherExpenses.(float64); ok && expensesFloat != 0 {
				pdf.CellFormat(140, 5, "  Other Expenses", "1", 0, "L", false, 0, "")
				pdf.CellFormat(50, 5, p.formatRupiah(-expensesFloat), "1", 0, "R", false, 0, "")
				pdf.Ln(5)
			}
		}
		pdf.Ln(3)
	}
}

// addFinancialRatiosSummary adds a summary of key financial ratios
func (p *PDFService) addFinancialRatiosSummary(pdf *gofpdf.Fpdf, data map[string]interface{}) {
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "FINANCIAL RATIOS SUMMARY")
	pdf.Ln(8)
	
	pdf.SetFont("Arial", "", 10)
	
	// First row of ratios
	if grossMargin, exists := data["GrossProfitMargin"]; exists {
		if gm, ok := grossMargin.(float64); ok && gm > 0 {
			pdf.Cell(95, 5, fmt.Sprintf("Gross Profit Margin: %.2f%%", gm))
		} else {
			pdf.Cell(95, 5, "Gross Profit Margin: N/A")
		}
	} else {
		pdf.Cell(95, 5, "Gross Profit Margin: N/A")
	}
	
	if operatingMargin, exists := data["OperatingMargin"]; exists {
		if om, ok := operatingMargin.(float64); ok && om > 0 {
			pdf.Cell(95, 5, fmt.Sprintf("Operating Margin: %.2f%%", om))
		} else {
			pdf.Cell(95, 5, "Operating Margin: N/A")
		}
	} else {
		pdf.Cell(95, 5, "Operating Margin: N/A")
	}
	pdf.Ln(5)
	
	// Second row of ratios
	if ebitdaMargin, exists := data["EBITDAMargin"]; exists {
		if em, ok := ebitdaMargin.(float64); ok && em > 0 {
			pdf.Cell(95, 5, fmt.Sprintf("EBITDA Margin: %.2f%%", em))
		} else {
			pdf.Cell(95, 5, "EBITDA Margin: N/A")
		}
	} else {
		pdf.Cell(95, 5, "EBITDA Margin: N/A")
	}
	
	if netMargin, exists := data["NetIncomeMargin"]; exists {
		if nm, ok := netMargin.(float64); ok && nm > 0 {
			pdf.Cell(95, 5, fmt.Sprintf("Net Income Margin: %.2f%%", nm))
		} else {
			pdf.Cell(95, 5, "Net Income Margin: N/A")
		}
	} else {
		pdf.Cell(95, 5, "Net Income Margin: N/A")
	}
	pdf.Ln(10)
}

// GenerateSSOTProfitLossPDF generates a PDF for SSOT-based Profit & Loss report
func (p *PDFService) GenerateSSOTProfitLossPDF(ssotData interface{}) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Try adding company letterhead/logo
	p.addCompanyLetterhead(pdf)

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "SSOT PROFIT & LOSS REPORT")
	pdf.Ln(15)

	companyInfo, err := p.getCompanyInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get company info: %v", err)
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(95, 8, companyInfo.CompanyName)
	pdf.SetFont("Arial", "", 10)
	pdf.Ln(6)
	pdf.Cell(95, 5, companyInfo.CompanyAddress)
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Phone: %s", companyInfo.CompanyPhone))
	pdf.Ln(5)
	pdf.Cell(95, 5, fmt.Sprintf("Email: %s", companyInfo.CompanyEmail))
	pdf.Ln(10)

	if dataMap, ok := ssotData.(map[string]interface{}); ok && len(dataMap) > 0 {
		p.renderSSOTFinancialSections(pdf, dataMap)
	} else {
		p.renderSSOTPlaceholder(pdf, ssotData)
	}

	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 4, fmt.Sprintf("Report generated on %s", time.Now().Format("02/01/2006 15:04")))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate SSOT Profit & Loss PDF: %v", err)
	}
	return buf.Bytes(), nil
}

// addPlaceholderSections adds placeholder sections when data extraction fails
func (p *PDFService) addPlaceholderSections(pdf *gofpdf.Fpdf) {
	// Add some example sections to show the structure
	sections := []string{"REVENUE", "COST OF GOODS SOLD", "GROSS PROFIT", "OPERATING EXPENSES", "OPERATING INCOME", "NET INCOME"}
	
	for _, section := range sections {
		pdf.SetFont("Arial", "B", 10)
		pdf.SetFillColor(240, 240, 240)
		pdf.CellFormat(140, 6, section, "1", 0, "L", true, 0, "")
		pdf.CellFormat(50, 6, "Processing...", "1", 0, "R", true, 0, "")
		pdf.Ln(6)
	}
}

// addProfitLossSections adds P&L sections to PDF
func (p *PDFService) addProfitLossSections(pdf *gofpdf.Fpdf, plData map[string]interface{}) {
	pdf.SetFont("Arial", "B", 12)
	
	// Process sections if they exist
	if sections, exists := plData["sections"]; exists {
		if sectionsSlice, ok := sections.([]interface{}); ok {
			for _, section := range sectionsSlice {
				if sectionMap, ok := section.(map[string]interface{}); ok {
					p.addPLSection(pdf, sectionMap)
					pdf.Ln(3)
				}
			}
		}
	}
	
	// Add financial metrics summary
	if metrics, exists := plData["financialMetrics"]; exists {
		p.addFinancialMetrics(pdf, metrics)
	}
}

// addPLSection adds a P&L section to PDF
func (p *PDFService) addPLSection(pdf *gofpdf.Fpdf, section map[string]interface{}) {
	sectionName := "Unknown Section"
	sectionTotal := 0.0
	
	if name, exists := section["name"]; exists {
		if nameStr, ok := name.(string); ok {
			sectionName = nameStr
		}
	}
	
	if total, exists := section["total"]; exists {
		if totalFloat, ok := total.(float64); ok {
			sectionTotal = totalFloat
		}
	}
	
	// Section header
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(190, 8, sectionName)
	pdf.Ln(8)
	
	// Add section items
	if items, exists := section["items"]; exists {
		p.addPLSectionItems(pdf, items)
	}
	
	// Add subsections if they exist
	if subsections, exists := section["subsections"]; exists {
		if subsectionsSlice, ok := subsections.([]interface{}); ok {
			for _, subsection := range subsectionsSlice {
				if subsectionMap, ok := subsection.(map[string]interface{}); ok {
					p.addPLSubsection(pdf, subsectionMap)
				}
			}
		}
	}
	
	// Section total
	if !section["is_calculated"].(bool) {
		pdf.SetFont("Arial", "B", 10)
		pdf.SetFillColor(245, 245, 245)
		pdf.CellFormat(145, 6, fmt.Sprintf("Total %s", sectionName), "1", 0, "L", true, 0, "")
		pdf.CellFormat(45, 6, p.formatRupiah(sectionTotal), "1", 0, "R", true, 0, "")
		pdf.Ln(6)
	} else {
		// For calculated sections like Net Income, show with different formatting
		pdf.SetFont("Arial", "B", 11)
		pdf.SetFillColor(230, 230, 230)
		pdf.CellFormat(145, 8, sectionName, "1", 0, "L", true, 0, "")
		pdf.CellFormat(45, 8, p.formatRupiah(sectionTotal), "1", 0, "R", true, 0, "")
		pdf.Ln(8)
	}
}

// addPLSectionItems adds section items to PDF
func (p *PDFService) addPLSectionItems(pdf *gofpdf.Fpdf, items interface{}) {
	pdf.SetFont("Arial", "", 9)
	
	if itemsSlice, ok := items.([]interface{}); ok {
		for _, item := range itemsSlice {
			if itemMap, ok := item.(map[string]interface{}); ok {
				name := "Unknown Item"
				amount := 0.0
				isPercentage := false
				
				if nameVal, exists := itemMap["name"]; exists {
					if nameStr, ok := nameVal.(string); ok {
						name = nameStr
					}
				}
				
				if amountVal, exists := itemMap["amount"]; exists {
					if amountFloat, ok := amountVal.(float64); ok {
						amount = amountFloat
					}
				}
				
				if isPercentageVal, exists := itemMap["is_percentage"]; exists {
					if isPercentageBool, ok := isPercentageVal.(bool); ok {
						isPercentage = isPercentageBool
					}
				}
				
				pdf.Cell(145, 5, fmt.Sprintf("  %s", name))
				if isPercentage {
					pdf.Cell(45, 5, fmt.Sprintf("%.2f%%", amount))
				} else {
					pdf.Cell(45, 5, p.formatRupiah(amount))
				}
				pdf.Ln(5)
			}
		}
	}
}

// addPLSubsection adds a P&L subsection to PDF
func (p *PDFService) addPLSubsection(pdf *gofpdf.Fpdf, subsection map[string]interface{}) {
	subsectionName := "Unknown Subsection"
	subsectionTotal := 0.0
	
	if name, exists := subsection["name"]; exists {
		if nameStr, ok := name.(string); ok {
			subsectionName = nameStr
		}
	}
	
	if total, exists := subsection["total"]; exists {
		if totalFloat, ok := total.(float64); ok {
			subsectionTotal = totalFloat
		}
	}
	
	// Subsection header
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 6, fmt.Sprintf("  %s", subsectionName))
	pdf.Ln(6)
	
	// Add subsection items
	if items, exists := subsection["items"]; exists {
		if itemsSlice, ok := items.([]interface{}); ok {
			pdf.SetFont("Arial", "", 9)
			for _, item := range itemsSlice {
				if itemMap, ok := item.(map[string]interface{}); ok {
					name := "Unknown Item"
					amount := 0.0
					
					if nameVal, exists := itemMap["name"]; exists {
						if nameStr, ok := nameVal.(string); ok {
							name = nameStr
						}
					}
					
					if amountVal, exists := itemMap["amount"]; exists {
						if amountFloat, ok := amountVal.(float64); ok {
							amount = amountFloat
						}
					}
					
					pdf.Cell(145, 4, fmt.Sprintf("    %s", name))
					pdf.Cell(45, 4, p.formatRupiah(amount))
					pdf.Ln(4)
				}
			}
		}
	}
	
	// Subsection total
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(248, 248, 248)
	pdf.CellFormat(145, 5, fmt.Sprintf("  Total %s", subsectionName), "1", 0, "L", true, 0, "")
	pdf.CellFormat(45, 5, p.formatRupiah(subsectionTotal), "1", 0, "R", true, 0, "")
	pdf.Ln(5)
}

// addFinancialMetrics adds financial metrics summary to PDF
func (p *PDFService) addFinancialMetrics(pdf *gofpdf.Fpdf, metrics interface{}) {
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "FINANCIAL METRICS SUMMARY")
	pdf.Ln(10)
	
	if metricsMap, ok := metrics.(map[string]interface{}); ok {
		pdf.SetFont("Arial", "", 9)
		
		// Display key metrics
		metricItems := []string{
			"grossProfit", "grossProfitMargin", "operatingIncome", 
			"operatingMargin", "netIncome", "netIncomeMargin",
		}
		
		metricLabels := map[string]string{
			"grossProfit": "Gross Profit",
			"grossProfitMargin": "Gross Profit Margin",
			"operatingIncome": "Operating Income",
			"operatingMargin": "Operating Margin",
			"netIncome": "Net Income",
			"netIncomeMargin": "Net Income Margin",
		}
		
		for _, metricKey := range metricItems {
			if value, exists := metricsMap[metricKey]; exists {
				label := metricLabels[metricKey]
				if valueFloat, ok := value.(float64); ok {
					pdf.Cell(95, 5, label+":")
					if strings.Contains(metricKey, "Margin") {
						pdf.Cell(95, 5, fmt.Sprintf("%.2f%%", valueFloat))
					} else {
						pdf.Cell(95, 5, p.formatRupiah(valueFloat))
					}
					pdf.Ln(5)
				}
			}
		}
	}
}
