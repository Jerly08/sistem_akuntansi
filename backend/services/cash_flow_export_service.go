package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// CashFlowExportService handles export functionality for Cash Flow reports
type CashFlowExportService struct{}

// NewCashFlowExportService creates a new cash flow export service
func NewCashFlowExportService() *CashFlowExportService {
	return &CashFlowExportService{}
}

// ExportToCSV exports cash flow data to CSV format
func (s *CashFlowExportService) ExportToCSV(data *SSOTCashFlowData) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header information
	writer.Write([]string{"Cash Flow Statement"})
	writer.Write([]string{data.Company.Name})
	writer.Write([]string{"Period:", data.StartDate.Format("02/01/2006") + " - " + data.EndDate.Format("02/01/2006")})
	writer.Write([]string{"Generated:", data.GeneratedAt.Format("02/01/2006 15:04")})
	writer.Write([]string{}) // Empty row

	// CSV Headers
	headers := []string{"Activity Type", "Category", "Account Code", "Account Name", "Amount", "Type"}
	writer.Write(headers)

	// Operating Activities
	writer.Write([]string{"OPERATING ACTIVITIES", "", "", "", "", ""})
	
	// Net Income
	writer.Write([]string{
		"Operating",
		"Net Income",
		"",
		"Net Income",
		s.formatAmount(data.OperatingActivities.NetIncome),
		"base",
	})

	// Adjustments
	if len(data.OperatingActivities.Adjustments.Items) > 0 {
		writer.Write([]string{"Operating", "Adjustments for Non-Cash Items", "", "", "", ""})
		for _, item := range data.OperatingActivities.Adjustments.Items {
			writer.Write([]string{
				"Operating",
				"Adjustments",
				item.AccountCode,
				item.AccountName,
				s.formatAmount(item.Amount),
				item.Type,
			})
		}
		writer.Write([]string{
			"Operating",
			"Total Adjustments",
			"",
			"",
			s.formatAmount(data.OperatingActivities.Adjustments.TotalAdjustments),
			"subtotal",
		})
	}

	// Working Capital Changes
	if len(data.OperatingActivities.WorkingCapitalChanges.Items) > 0 {
		writer.Write([]string{"Operating", "Changes in Working Capital", "", "", "", ""})
		for _, item := range data.OperatingActivities.WorkingCapitalChanges.Items {
			writer.Write([]string{
				"Operating",
				"Working Capital",
				item.AccountCode,
				item.AccountName,
				s.formatAmount(item.Amount),
				item.Type,
			})
		}
		writer.Write([]string{
			"Operating",
			"Total Working Capital Changes",
			"",
			"",
			s.formatAmount(data.OperatingActivities.WorkingCapitalChanges.TotalWorkingCapitalChanges),
			"subtotal",
		})
	}

	writer.Write([]string{
		"Operating",
		"NET CASH FROM OPERATING ACTIVITIES",
		"",
		"",
		s.formatAmount(data.OperatingActivities.TotalOperatingCashFlow),
		"total",
	})
	writer.Write([]string{}) // Empty row

	// Investing Activities
	writer.Write([]string{"INVESTING ACTIVITIES", "", "", "", "", ""})
	if len(data.InvestingActivities.Items) > 0 {
		for _, item := range data.InvestingActivities.Items {
			writer.Write([]string{
				"Investing",
				"Investing Activities",
				item.AccountCode,
				item.AccountName,
				s.formatAmount(item.Amount),
				item.Type,
			})
		}
	}
	writer.Write([]string{
		"Investing",
		"NET CASH FROM INVESTING ACTIVITIES",
		"",
		"",
		s.formatAmount(data.InvestingActivities.TotalInvestingCashFlow),
		"total",
	})
	writer.Write([]string{}) // Empty row

	// Financing Activities
	writer.Write([]string{"FINANCING ACTIVITIES", "", "", "", "", ""})
	if len(data.FinancingActivities.Items) > 0 {
		for _, item := range data.FinancingActivities.Items {
			writer.Write([]string{
				"Financing",
				"Financing Activities",
				item.AccountCode,
				item.AccountName,
				s.formatAmount(item.Amount),
				item.Type,
			})
		}
	}
	writer.Write([]string{
		"Financing",
		"NET CASH FROM FINANCING ACTIVITIES",
		"",
		"",
		s.formatAmount(data.FinancingActivities.TotalFinancingCashFlow),
		"total",
	})
	writer.Write([]string{}) // Empty row

	// Summary
	writer.Write([]string{"CASH FLOW SUMMARY", "", "", "", "", ""})
	writer.Write([]string{
		"Summary",
		"Cash at Beginning of Period",
		"",
		"",
		s.formatAmount(data.CashAtBeginning),
		"summary",
	})
	writer.Write([]string{
		"Summary",
		"Net Cash Flow",
		"",
		"",
		s.formatAmount(data.NetCashFlow),
		"summary",
	})
	writer.Write([]string{
		"Summary",
		"Cash at End of Period",
		"",
		"",
		s.formatAmount(data.CashAtEnd),
		"summary",
	})

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to write CSV: %v", err)
	}

	return buf.Bytes(), nil
}

// ExportToPDF exports cash flow data to PDF format
func (s *CashFlowExportService) ExportToPDF(data *SSOTCashFlowData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Title
	pdf.Cell(0, 10, "CASH FLOW STATEMENT")
	pdf.Ln(10)
	
	// Company info
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, data.Company.Name)
	pdf.Ln(8)
	
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Period: %s - %s", 
		data.StartDate.Format("02/01/2006"), 
		data.EndDate.Format("02/01/2006")))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Generated: %s", 
		data.GeneratedAt.Format("02/01/2006 15:04")))
	pdf.Ln(15)

	// Operating Activities
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "OPERATING ACTIVITIES")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 10)
	
	// Net Income
	pdf.Cell(120, 6, "Net Income")
	pdf.Cell(0, 6, s.formatAmountPDF(data.OperatingActivities.NetIncome))
	pdf.Ln(6)
	
	// Adjustments
	if len(data.OperatingActivities.Adjustments.Items) > 0 {
		pdf.Ln(3)
		pdf.SetFont("Arial", "I", 10)
		pdf.Cell(0, 6, "Adjustments for Non-Cash Items:")
		pdf.Ln(8)
		
		pdf.SetFont("Arial", "", 9)
		for _, item := range data.OperatingActivities.Adjustments.Items {
			pdf.Cell(10, 5, "")
			pdf.Cell(90, 5, fmt.Sprintf("%s - %s", item.AccountCode, item.AccountName))
			pdf.Cell(0, 5, s.formatAmountPDF(item.Amount))
			pdf.Ln(5)
		}
		
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(100, 6, "Total Adjustments")
		pdf.Cell(0, 6, s.formatAmountPDF(data.OperatingActivities.Adjustments.TotalAdjustments))
		pdf.Ln(8)
	}

	// Working Capital Changes
	if len(data.OperatingActivities.WorkingCapitalChanges.Items) > 0 {
		pdf.Ln(3)
		pdf.SetFont("Arial", "I", 10)
		pdf.Cell(0, 6, "Changes in Working Capital:")
		pdf.Ln(8)
		
		pdf.SetFont("Arial", "", 9)
		for _, item := range data.OperatingActivities.WorkingCapitalChanges.Items {
			pdf.Cell(10, 5, "")
			pdf.Cell(90, 5, fmt.Sprintf("%s - %s", item.AccountCode, item.AccountName))
			pdf.Cell(0, 5, s.formatAmountPDF(item.Amount))
			pdf.Ln(5)
		}
		
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(100, 6, "Total Working Capital Changes")
		pdf.Cell(0, 6, s.formatAmountPDF(data.OperatingActivities.WorkingCapitalChanges.TotalWorkingCapitalChanges))
		pdf.Ln(8)
	}

	// Operating total
	pdf.Ln(3)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(100, 8, "NET CASH FROM OPERATING ACTIVITIES")
	pdf.Cell(0, 8, s.formatAmountPDF(data.OperatingActivities.TotalOperatingCashFlow))
	pdf.Ln(12)

	// Investing Activities
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "INVESTING ACTIVITIES")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 10)
	if len(data.InvestingActivities.Items) > 0 {
		for _, item := range data.InvestingActivities.Items {
			pdf.Cell(100, 6, fmt.Sprintf("%s - %s", item.AccountCode, item.AccountName))
			pdf.Cell(0, 6, s.formatAmountPDF(item.Amount))
			pdf.Ln(6)
		}
	} else {
		pdf.Cell(0, 6, "No investing activities")
		pdf.Ln(6)
	}

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(100, 8, "NET CASH FROM INVESTING ACTIVITIES")
	pdf.Cell(0, 8, s.formatAmountPDF(data.InvestingActivities.TotalInvestingCashFlow))
	pdf.Ln(12)

	// Financing Activities
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "FINANCING ACTIVITIES")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 10)
	if len(data.FinancingActivities.Items) > 0 {
		for _, item := range data.FinancingActivities.Items {
			pdf.Cell(100, 6, fmt.Sprintf("%s - %s", item.AccountCode, item.AccountName))
			pdf.Cell(0, 6, s.formatAmountPDF(item.Amount))
			pdf.Ln(6)
		}
	} else {
		pdf.Cell(0, 6, "No financing activities")
		pdf.Ln(6)
	}

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(100, 8, "NET CASH FROM FINANCING ACTIVITIES")
	pdf.Cell(0, 8, s.formatAmountPDF(data.FinancingActivities.TotalFinancingCashFlow))
	pdf.Ln(15)

	// Summary
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "NET CASH FLOW")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(100, 6, "Cash at Beginning of Period")
	pdf.Cell(0, 6, s.formatAmountPDF(data.CashAtBeginning))
	pdf.Ln(6)
	
	pdf.Cell(100, 6, "Net Cash Flow from Activities")
	pdf.Cell(0, 6, s.formatAmountPDF(data.NetCashFlow))
	pdf.Ln(6)
	
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(100, 8, "Cash at End of Period")
	pdf.Cell(0, 8, s.formatAmountPDF(data.CashAtEnd))
	pdf.Ln(15)

	// Add footer with generation info
	pdf.SetY(280)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 4, fmt.Sprintf("Generated on %s", time.Now().Format("02/01/2006 15:04:05")))

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}
	return buf.Bytes(), nil
}

// formatAmount formats amount for CSV export
func (s *CashFlowExportService) formatAmount(amount float64) string {
	// Format with thousand separators and 2 decimal places
	return fmt.Sprintf("%.2f", amount)
}

// formatAmountPDF formats amount for PDF export with Indonesian formatting
func (s *CashFlowExportService) formatAmountPDF(amount float64) string {
	// Convert to string with 2 decimal places
	str := fmt.Sprintf("%.2f", amount)
	
	// Add thousand separators
	parts := strings.Split(str, ".")
	intPart := parts[0]
	decPart := parts[1]
	
	// Add commas for thousands
	n := len(intPart)
	if n > 3 {
		var result strings.Builder
		for i, char := range intPart {
			if i > 0 && (n-i)%3 == 0 {
				result.WriteString(",")
			}
			result.WriteRune(char)
		}
		return fmt.Sprintf("%s.%s", result.String(), decPart)
	}
	
	return str
}

// GetCSVFilename generates appropriate filename for CSV export
func (s *CashFlowExportService) GetCSVFilename(data *SSOTCashFlowData) string {
	return fmt.Sprintf("cash_flow_%s_to_%s.csv",
		data.StartDate.Format("2006-01-02"),
		data.EndDate.Format("2006-01-02"))
}

// GetPDFFilename generates appropriate filename for PDF export
func (s *CashFlowExportService) GetPDFFilename(data *SSOTCashFlowData) string {
	return fmt.Sprintf("cash_flow_%s_to_%s.pdf",
		data.StartDate.Format("2006-01-02"),
		data.EndDate.Format("2006-01-02"))
}