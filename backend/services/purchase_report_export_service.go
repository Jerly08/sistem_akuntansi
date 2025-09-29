package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

// PurchaseReportExportService handles export functionality for Purchase Report
type PurchaseReportExportService struct{ db *gorm.DB }

// NewPurchaseReportExportService creates a new purchase report export service
func NewPurchaseReportExportService(db *gorm.DB) *PurchaseReportExportService {
	return &PurchaseReportExportService{db: db}
}

// getCompanyInfo from settings with defaults
func (s *PurchaseReportExportService) getCompanyInfo() *models.Settings {
	if s.db == nil {
		return &models.Settings{CompanyName: "PT. Sistem Akuntansi Indonesia"}
	}
	var settings models.Settings
	if err := s.db.First(&settings).Error; err != nil {
		return &models.Settings{CompanyName: "PT. Sistem Akuntansi Indonesia"}
	}
	return &settings
}

// ExportToCSV exports purchase report to CSV bytes (optional helper)
func (s *PurchaseReportExportService) ExportToCSV(data *PurchaseReportData) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Header
	w.Write([]string{"Purchase Report"})
	w.Write([]string{data.Company.Name})
	w.Write([]string{"Period:", data.StartDate.Format("2006-01-02"), "to", data.EndDate.Format("2006-01-02")})
	w.Write([]string{"Generated:", data.GeneratedAt.In(time.Local).Format("2006-01-02 15:04")})
	w.Write([]string{})

	// Summary
	w.Write([]string{"SUMMARY"})
	w.Write([]string{"Total Purchases", fmt.Sprintf("%d", data.TotalPurchases)})
	w.Write([]string{"Completed Purchases", fmt.Sprintf("%d", data.CompletedPurchases)})
	w.Write([]string{"Total Amount", fmt.Sprintf("%.2f", data.TotalAmount)})
	w.Write([]string{"Total Paid", fmt.Sprintf("%.2f", data.TotalPaid)})
	w.Write([]string{"Outstanding Payables", fmt.Sprintf("%.2f", data.OutstandingPayables)})
	w.Write([]string{})

	// Purchases by vendor (top 20)
	if len(data.PurchasesByVendor) > 0 {
		w.Write([]string{"PURCHASES BY VENDOR (Top 20)"})
		w.Write([]string{"Vendor ID", "Vendor Name", "Total Purchases", "Total Amount", "Total Paid", "Outstanding", "Last Purchase", "Payment Method", "Status"})
		limit := len(data.PurchasesByVendor)
		if limit > 20 { limit = 20 }
		for i := 0; i < limit; i++ {
			v := data.PurchasesByVendor[i]
			w.Write([]string{
				fmt.Sprintf("%d", v.VendorID), v.VendorName,
				fmt.Sprintf("%d", v.TotalPurchases),
				fmt.Sprintf("%.2f", v.TotalAmount),
				fmt.Sprintf("%.2f", v.TotalPaid),
				fmt.Sprintf("%.2f", v.Outstanding),
				v.LastPurchaseDate.Format("2006-01-02"),
				v.PaymentMethod, v.Status,
			})
		}
		w.Write([]string{})
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("failed to write CSV: %v", err)
	}
	return buf.Bytes(), nil
}

// ExportToPDF exports purchase report to PDF bytes (invoice-like)
func (s *PurchaseReportExportService) ExportToPDF(data *PurchaseReportData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	lm, tm, rm, _ := pdf.GetMargins()
	pageW, _ := pdf.GetPageSize()
	contentW := pageW - lm - rm

	// Company settings for consistent letterhead
	settings := s.getCompanyInfo()

	// Try to render real logo at top-left
	logoW := 35.0
	logoPath := strings.TrimSpace(settings.CompanyLogo)
	logoDrawn := false
	if logoPath != "" {
		if strings.HasPrefix(logoPath, "/") { logoPath = "." + logoPath }
		if _, err := os.Stat(logoPath); err != nil {
			alt := filepath.Clean("./" + strings.TrimPrefix(settings.CompanyLogo, "/"))
			if _, err2 := os.Stat(alt); err2 == nil { logoPath = alt } else { logoPath = "" }
		}
		if logoPath != "" {
			if imgType := detectImageType(logoPath); imgType != "" {
				pdf.ImageOptions(logoPath, lm, tm, logoW, 0, false, gofpdf.ImageOptions{ImageType: imgType, ReadDpi: true}, 0, "")
				logoDrawn = true
			}
		}
	}
	if !logoDrawn {
		pdf.SetDrawColor(220, 220, 220)
		pdf.SetFillColor(248, 249, 250)
		pdf.SetLineWidth(0.3)
		pdf.Rect(lm, tm, logoW, logoW, "FD")
		pdf.SetFont("Arial", "B", 16)
		pdf.SetTextColor(120, 120, 120)
		pdf.SetXY(lm+8, tm+19)
		pdf.CellFormat(19, 8, "</>", "", 0, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	}

	// Company info (right-aligned)
	companyName := strings.TrimSpace(settings.CompanyName)
	if companyName == "" { companyName = data.Company.Name }
	pdf.SetFont("Arial", "B", 12)
	w := pdf.GetStringWidth(companyName)
	pdf.SetXY(pageW-rm-w, tm)
	pdf.Cell(w, 6, companyName)
	pdf.SetFont("Arial", "", 9)
	addr := strings.TrimSpace(settings.CompanyAddress)
	if addr == "" { addr = strings.TrimSpace(data.Company.Address) }
	if addr != "" {
		pdf.SetXY(pageW-rm-pdf.GetStringWidth(addr), tm+8)
		pdf.Cell(0, 4, addr)
	}
	phoneVal := strings.TrimSpace(settings.CompanyPhone)
	if phoneVal == "" { phoneVal = strings.TrimSpace(data.Company.Phone) }
	if phoneVal != "" {
		phone := fmt.Sprintf("Phone: %s", phoneVal)
		pdf.SetXY(pageW-rm-pdf.GetStringWidth(phone), tm+14)
		pdf.Cell(0, 4, phone)
	}
	emailVal := strings.TrimSpace(settings.CompanyEmail)
	if emailVal == "" { emailVal = strings.TrimSpace(data.Company.Email) }
	if emailVal != "" {
		email := fmt.Sprintf("Email: %s", emailVal)
		pdf.SetXY(pageW-rm-pdf.GetStringWidth(email), tm+20)
		pdf.Cell(0, 4, email)
	}

	// Separator
	pdf.SetDrawColor(238, 238, 238)
	pdf.SetLineWidth(0.2)
	pdf.Line(lm, tm+45, pageW-rm, tm+45)

	// Title and period
	pdf.SetY(tm + 55)
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(51, 51, 51)
	pdf.Cell(contentW, 10, "PURCHASE REPORT")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(contentW, 6, fmt.Sprintf("Period: %s to %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")))
	pdf.Ln(6)
	pdf.Cell(contentW, 6, fmt.Sprintf("Generated: %s", data.GeneratedAt.In(time.Local).Format("2006-01-02 15:04")))
	pdf.Ln(10)

	// Summary block
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "SUMMARY")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(90, 6, "Total Purchases", "1", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, fmt.Sprintf("%d", data.TotalPurchases), "1", 1, "R", false, 0, "")
	pdf.CellFormat(90, 6, "Completed Purchases", "1", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, fmt.Sprintf("%d", data.CompletedPurchases), "1", 1, "R", false, 0, "")
	pdf.CellFormat(90, 6, "Total Amount", "1", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, formatRupiahSimple(data.TotalAmount), "1", 1, "R", false, 0, "")
	pdf.CellFormat(90, 6, "Total Paid", "1", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, formatRupiahSimple(data.TotalPaid), "1", 1, "R", false, 0, "")
	pdf.CellFormat(90, 6, "Outstanding Payables", "1", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, formatRupiahSimple(data.OutstandingPayables), "1", 1, "R", false, 0, "")
	pdf.Ln(6)

	// Top vendors table (limit to fit one page)
	if len(data.PurchasesByVendor) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "TOP VENDORS")
		pdf.Ln(8)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(220, 220, 220)
		pdf.CellFormat(60, 7, "Vendor", "1", 0, "L", true, 0, "")
		pdf.CellFormat(25, 7, "Orders", "1", 0, "C", true, 0, "")
		pdf.CellFormat(35, 7, "Amount", "1", 0, "R", true, 0, "")
		pdf.CellFormat(35, 7, "Paid", "1", 0, "R", true, 0, "")
		pdf.CellFormat(35, 7, "Outstanding", "1", 1, "R", true, 0, "")
		pdf.SetFont("Arial", "", 9)
		limit := len(data.PurchasesByVendor)
		if limit > 12 { limit = 12 }
		for i := 0; i < limit; i++ {
			v := data.PurchasesByVendor[i]
			name := v.VendorName
			if len(name) > 30 { name = name[:27] + "..." }
			pdf.CellFormat(60, 6, name, "1", 0, "L", false, 0, "")
			pdf.CellFormat(25, 6, fmt.Sprintf("%d", v.TotalPurchases), "1", 0, "C", false, 0, "")
			pdf.CellFormat(35, 6, formatRupiahSimple(v.TotalAmount), "1", 0, "R", false, 0, "")
			pdf.CellFormat(35, 6, formatRupiahSimple(v.TotalPaid), "1", 0, "R", false, 0, "")
			pdf.CellFormat(35, 6, formatRupiahSimple(v.Outstanding), "1", 1, "R", false, 0, "")
		}
	}

	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, fmt.Errorf("failed to generate purchase report PDF: %v", err)
	}
return out.Bytes(), nil
}


// formatRupiahSimple formats a number as Indonesian Rupiah (no decimals)
func formatRupiahSimple(amount float64) string {
	// Round to integer and format with thousand separators using simple logic
	s := fmt.Sprintf("%.0f", amount)
	if s == "0" { return "Rp 0" }
	var parts []string
	for i, r := range reverseString(s) {
		if i > 0 && i%3 == 0 { parts = append(parts, ".") }
		parts = append(parts, string(r))
	}
	formatted := reverseString(strings.Join(parts, ""))
	return "Rp " + formatted
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
