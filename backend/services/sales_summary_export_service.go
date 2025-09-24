package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// SalesSummaryExportService handles export functionality for Sales Summary reports
type SalesSummaryExportService struct{}

// NewSalesSummaryExportService creates a new sales summary export service
func NewSalesSummaryExportService() *SalesSummaryExportService {
	return &SalesSummaryExportService{}
}

// ExportToCSV exports sales summary data to CSV format
func (s *SalesSummaryExportService) ExportToCSV(data *SalesSummaryData) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header information
	writer.Write([]string{"Sales Summary Report"})
	company := data.Company.Name
	if company == "" {
		company = "PT. Sistem Akuntansi Indonesia"
	}
	writer.Write([]string{company})
	writer.Write([]string{"Period:", data.StartDate.Format("2006-01-02"), "to", data.EndDate.Format("2006-01-02")})
	writer.Write([]string{"Generated:", data.GeneratedAt.In(time.Local).Format("2006-01-02 15:04")})
	writer.Write([]string{}) // empty row

	// Summary metrics
	writer.Write([]string{"SUMMARY"})
	writer.Write([]string{"Total Revenue", s.formatAmount(data.TotalRevenue)})
	writer.Write([]string{"Total Transactions", fmt.Sprintf("%d", data.TotalTransactions)})
	writer.Write([]string{"Average Order Value", s.formatAmount(data.AverageOrderValue)})
	writer.Write([]string{"Total Customers", fmt.Sprintf("%d", data.TotalCustomers)})
	writer.Write([]string{})

	// Sales by period
	if len(data.SalesByPeriod) > 0 {
		writer.Write([]string{"SALES BY PERIOD"})
		writer.Write([]string{"Period", "Start Date", "End Date", "Amount", "Transactions", "Growth %"})
		for _, p := range data.SalesByPeriod {
			writer.Write([]string{
				p.Period,
				p.StartDate.Format("2006-01-02"),
				p.EndDate.Format("2006-01-02"),
				s.formatAmount(p.Amount),
				fmt.Sprintf("%d", p.Transactions),
				fmt.Sprintf("%.2f", p.GrowthRate),
			})
		}
		writer.Write([]string{})
	}

	// Sales by customer (top 20)
	if len(data.SalesByCustomer) > 0 {
		writer.Write([]string{"SALES BY CUSTOMER (Top 20)"})
		writer.Write([]string{"Customer ID", "Customer Name", "Total Amount", "Transactions", "Avg. Order"})
		limit := len(data.SalesByCustomer)
		if limit > 20 {
			limit = 20
		}
		for i := 0; i < limit; i++ {
			c := data.SalesByCustomer[i]
			writer.Write([]string{
				fmt.Sprintf("%d", c.CustomerID),
				s.csvSafe(c.CustomerName),
				s.formatAmount(c.TotalAmount),
				fmt.Sprintf("%d", c.TransactionCount),
				s.formatAmount(c.AverageOrder),
			})
		}
		writer.Write([]string{})
	}

	// Sales by product (top 20)
	if len(data.SalesByProduct) > 0 {
		writer.Write([]string{"SALES BY PRODUCT (Top 20)"})
		writer.Write([]string{"Product ID", "Product Name", "Qty Sold", "Total Amount", "Avg. Price"})
		limit := len(data.SalesByProduct)
		if limit > 20 {
			limit = 20
		}
		for i := 0; i < limit; i++ {
			p := data.SalesByProduct[i]
			writer.Write([]string{
				fmt.Sprintf("%d", p.ProductID),
				s.csvSafe(p.ProductName),
				fmt.Sprintf("%d", p.QuantitySold),
				s.formatAmount(p.TotalAmount),
				s.formatAmount(p.AveragePrice),
			})
		}
		writer.Write([]string{})
	}

	// Sales by status
	if len(data.SalesByStatus) > 0 {
		writer.Write([]string{"SALES BY STATUS"})
		writer.Write([]string{"Status", "Count", "Amount", "Percentage %"})
		for _, st := range data.SalesByStatus {
			writer.Write([]string{
				st.Status,
				fmt.Sprintf("%d", st.Count),
				s.formatAmount(st.Amount),
				fmt.Sprintf("%.2f", st.Percentage),
			})
		}
		writer.Write([]string{})
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to write CSV: %v", err)
	}

	return buf.Bytes(), nil
}

// ExportToPDF exports sales summary data to PDF format
func (s *SalesSummaryExportService) ExportToPDF(data *SalesSummaryData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "SALES SUMMARY REPORT")
	pdf.Ln(10)

	// Company & Period
	company := data.Company.Name
	if company == "" {
		company = "PT. Sistem Akuntansi Indonesia"
	}
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 7, company)
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Period: %s to %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Generated: %s", data.GeneratedAt.In(time.Local).Format("2006-01-02 15:04")))
	pdf.Ln(10)

	// Summary metrics block
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "SUMMARY")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(90, 6, "Total Revenue", "1", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, s.formatAmount(data.TotalRevenue), "1", 1, "R", false, 0, "")
	pdf.CellFormat(90, 6, "Total Transactions", "1", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, fmt.Sprintf("%d", data.TotalTransactions), "1", 1, "R", false, 0, "")
	pdf.CellFormat(90, 6, "Average Order Value", "1", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, s.formatAmount(data.AverageOrderValue), "1", 1, "R", false, 0, "")
	pdf.CellFormat(90, 6, "Total Customers", "1", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, fmt.Sprintf("%d", data.TotalCustomers), "1", 1, "R", false, 0, "")
	pdf.Ln(6)

	// Sales by period (limit rows to fit a page)
	if len(data.SalesByPeriod) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "SALES BY PERIOD")
		pdf.Ln(8)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(220, 220, 220)
		pdf.CellFormat(40, 7, "Period", "1", 0, "C", true, 0, "")
		pdf.CellFormat(35, 7, "Start Date", "1", 0, "C", true, 0, "")
		pdf.CellFormat(35, 7, "End Date", "1", 0, "C", true, 0, "")
		pdf.CellFormat(40, 7, "Amount", "1", 0, "C", true, 0, "")
		pdf.CellFormat(40, 7, "Transactions", "1", 1, "C", true, 0, "")
		pdf.SetFont("Arial", "", 9)
		limit := len(data.SalesByPeriod)
		if limit > 12 {
			limit = 12
		}
		for i := 0; i < limit; i++ {
			p := data.SalesByPeriod[i]
			pdf.CellFormat(40, 6, p.Period, "1", 0, "L", false, 0, "")
			pdf.CellFormat(35, 6, p.StartDate.Format("2006-01-02"), "1", 0, "C", false, 0, "")
			pdf.CellFormat(35, 6, p.EndDate.Format("2006-01-02"), "1", 0, "C", false, 0, "")
			pdf.CellFormat(40, 6, s.formatAmount(p.Amount), "1", 0, "R", false, 0, "")
			pdf.CellFormat(40, 6, fmt.Sprintf("%d", p.Transactions), "1", 1, "C", false, 0, "")
		}
		pdf.Ln(4)
	}

	// Sales by customer (top 10)
	if len(data.SalesByCustomer) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "TOP CUSTOMERS")
		pdf.Ln(8)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(220, 220, 220)
		pdf.CellFormat(20, 7, "ID", "1", 0, "C", true, 0, "")
		pdf.CellFormat(90, 7, "Customer", "1", 0, "C", true, 0, "")
		pdf.CellFormat(40, 7, "Amount", "1", 0, "C", true, 0, "")
		pdf.CellFormat(40, 7, "Orders", "1", 1, "C", true, 0, "")
		pdf.SetFont("Arial", "", 9)
		limit := len(data.SalesByCustomer)
		if limit > 10 {
			limit = 10
		}
		for i := 0; i < limit; i++ {
			c := data.SalesByCustomer[i]
			name := c.CustomerName
			if len(name) > 40 {
				name = name[:37] + "..."
			}
			pdf.CellFormat(20, 6, fmt.Sprintf("%d", c.CustomerID), "1", 0, "C", false, 0, "")
			pdf.CellFormat(90, 6, name, "1", 0, "L", false, 0, "")
			pdf.CellFormat(40, 6, s.formatAmount(c.TotalAmount), "1", 0, "R", false, 0, "")
			pdf.CellFormat(40, 6, fmt.Sprintf("%d", c.TransactionCount), "1", 1, "C", false, 0, "")
		}
	}

	// Output to buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}

	return buf.Bytes(), nil
}

// GetCSVFilename generates appropriate filename for CSV export
func (s *SalesSummaryExportService) GetCSVFilename(data *SalesSummaryData) string {
	return fmt.Sprintf("sales_summary_%s_to_%s.csv",
		data.StartDate.Format("2006-01-02"),
		data.EndDate.Format("2006-01-02"))
}

// GetPDFFilename generates appropriate filename for PDF export
func (s *SalesSummaryExportService) GetPDFFilename(data *SalesSummaryData) string {
	return fmt.Sprintf("sales_summary_%s_to_%s.pdf",
		data.StartDate.Format("2006-01-02"),
		data.EndDate.Format("2006-01-02"))
}

// formatAmount formats float as string with thousand separators (basic)
func (s *SalesSummaryExportService) formatAmount(amount float64) string {
	// Use dot as thousand separator and comma as decimal separator for readability
	str := fmt.Sprintf("%.2f", amount)
	parts := strings.Split(str, ".")
	intPart := parts[0]
	decPart := parts[1]
	var out []rune
	for i, r := range reverseRunes(intPart) {
		if i != 0 && i%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, r)
	}
	// reverse back
	rev := reverseRunes(string(out))
	return fmt.Sprintf("%s.%s", rev, decPart)
}

func (s *SalesSummaryExportService) csvSafe(val string) string {
	val = strings.ReplaceAll(val, "\n", " ")
	return val
}

func reverseRunes(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
