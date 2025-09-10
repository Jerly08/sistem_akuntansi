package services

import (
	"bytes"
	"fmt"
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

// GenerateInvoicePDF generates a PDF for a sale invoice
func (p *PDFService) GenerateInvoicePDF(sale *models.Sale) ([]byte, error) {
	// Create new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Company header
	pdf.Cell(190, 10, "INVOICE")
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
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Title
	pdf.Cell(270, 10, "SALES REPORT")
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
	pdf.CellFormat(22, 8, "Sale Code", "1", 0, "C", true, 0, "")
	pdf.CellFormat(22, 8, "Invoice No.", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Customer", "1", 0, "L", true, 0, "")
	pdf.CellFormat(18, 8, "Type", "1", 0, "C", true, 0, "")
	pdf.CellFormat(18, 8, "Status", "1", 0, "C", true, 0, "")
	pdf.CellFormat(44, 8, "Amount", "1", 0, "R", true, 0, "")
	pdf.CellFormat(44, 8, "Paid", "1", 0, "R", true, 0, "")
	pdf.CellFormat(44, 8, "Outstanding", "1", 0, "R", true, 0, "")
	pdf.Ln(8)

	// Table data
	pdf.SetFont("Arial", "", 8)
	pdf.SetFillColor(255, 255, 255)
	
	totalAmount := 0.0
	totalPaid := 0.0
	totalOutstanding := 0.0

	for _, sale := range sales {
		// Check if we need a new page
		if pdf.GetY() > 180 {
			pdf.AddPage()
			// Re-add headers
			pdf.SetFont("Arial", "B", 8)
			pdf.SetFillColor(220, 220, 220)
			pdf.CellFormat(18, 8, "Date", "1", 0, "C", true, 0, "")
			pdf.CellFormat(22, 8, "Sale Code", "1", 0, "C", true, 0, "")
			pdf.CellFormat(22, 8, "Invoice No.", "1", 0, "C", true, 0, "")
			pdf.CellFormat(40, 8, "Customer", "1", 0, "L", true, 0, "")
			pdf.CellFormat(18, 8, "Type", "1", 0, "C", true, 0, "")
			pdf.CellFormat(18, 8, "Status", "1", 0, "C", true, 0, "")
			pdf.CellFormat(44, 8, "Amount", "1", 0, "R", true, 0, "")
			pdf.CellFormat(44, 8, "Paid", "1", 0, "R", true, 0, "")
			pdf.CellFormat(44, 8, "Outstanding", "1", 0, "R", true, 0, "")
			pdf.Ln(8)
			pdf.SetFont("Arial", "", 8)
			pdf.SetFillColor(255, 255, 255)
		}

		// Sale data
		date := sale.Date.Format("02/01/06")
		customerName := "N/A"
		if sale.Customer.ID != 0 {
			customerName = sale.Customer.Name
			// Truncate if too long
			if len(customerName) > 30 {
				customerName = customerName[:27] + "..."
			}
		}

		invoiceNumber := sale.InvoiceNumber
		if invoiceNumber == "" {
			invoiceNumber = "-"
		}

		amount := p.formatRupiah(sale.TotalAmount)
		paid := p.formatRupiah(sale.PaidAmount)
		outstanding := p.formatRupiah(sale.OutstandingAmount)

		pdf.CellFormat(18, 5, date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(22, 5, sale.Code, "1", 0, "L", false, 0, "")
		pdf.CellFormat(22, 5, invoiceNumber, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 5, customerName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(18, 5, sale.Type, "1", 0, "C", false, 0, "")
		pdf.CellFormat(18, 5, sale.Status, "1", 0, "C", false, 0, "")
		pdf.CellFormat(44, 5, amount, "1", 0, "R", false, 0, "")
		pdf.CellFormat(44, 5, paid, "1", 0, "R", false, 0, "")
		pdf.CellFormat(44, 5, outstanding, "1", 0, "R", false, 0, "")
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
	pdf.CellFormat(138, 6, "TOTAL", "1", 0, "R", true, 0, "")
	pdf.CellFormat(44, 6, p.formatRupiah(totalAmount), "1", 0, "R", true, 0, "")
	pdf.CellFormat(44, 6, p.formatRupiah(totalPaid), "1", 0, "R", true, 0, "")
	pdf.CellFormat(44, 6, p.formatRupiah(totalOutstanding), "1", 0, "R", true, 0, "")

	// Statistics
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(270, 6, "SUMMARY STATISTICS")
	pdf.Ln(6)
	
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(135, 5, fmt.Sprintf("Total Sales: %d", len(sales)))
	pdf.Cell(135, 5, fmt.Sprintf("Total Amount: %s", p.formatRupiah(totalAmount)))
	pdf.Ln(5)
	pdf.Cell(135, 5, fmt.Sprintf("Total Paid: %s", p.formatRupiah(totalPaid)))
	pdf.Cell(135, 5, fmt.Sprintf("Total Outstanding: %s", p.formatRupiah(totalOutstanding)))
	pdf.Ln(5)
	
	if len(sales) > 0 {
		avgAmount := totalAmount / float64(len(sales))
		pdf.Cell(135, 5, fmt.Sprintf("Average Sale Amount: %s", p.formatRupiah(avgAmount)))
	}

	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sales report PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// GeneratePaymentReportPDF generates a PDF for payments report
func (p *PDFService) GeneratePaymentReportPDF(payments []models.Payment, startDate, endDate string) ([]byte, error) {
	// Create new PDF document
	pdf := gofpdf.New("L", "mm", "A4", "") // Landscape orientation
	pdf.AddPage()

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

	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Payment header
	pdf.Cell(190, 10, "PAYMENT VOUCHER")
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
