package services

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/utils"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

type ExportService interface {
	ExportAccountsPDF(ctx context.Context) ([]byte, error)
	ExportAccountsExcel(ctx context.Context) ([]byte, error)
}

type ExportServiceImpl struct {
	accountRepo repositories.AccountRepository
}

func NewExportService(accountRepo repositories.AccountRepository) ExportService {
	return &ExportServiceImpl{
		accountRepo: accountRepo,
	}
}

// ExportAccountsPDF exports accounts to PDF format
func (s *ExportServiceImpl) ExportAccountsPDF(ctx context.Context) ([]byte, error) {
	// Get all accounts from database
	accounts, err := s.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, utils.NewInternalError("Failed to fetch accounts", err)
	}

	// Create new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 16)
	
	// Title
	pdf.Cell(190, 10, "Chart of Accounts")
	pdf.Ln(15)

	// Company info (you can get this from database)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 5, fmt.Sprintf("Generated on: %s", time.Now().Format("2006-01-02 15:04:05")))
	pdf.Ln(10)

	// Table headers
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(30, 8, "Code", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 8, "Name", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, "Type", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Balance", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, "Status", "1", 0, "C", true, 0, "")
	pdf.Ln(8)

	// Table data
	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(255, 255, 255)
	
	for _, account := range accounts {
		// Check if we need a new page
		if pdf.GetY() > 270 {
			pdf.AddPage()
			// Re-add headers
			pdf.SetFont("Arial", "B", 10)
			pdf.SetFillColor(220, 220, 220)
			pdf.CellFormat(30, 8, "Code", "1", 0, "C", true, 0, "")
			pdf.CellFormat(60, 8, "Name", "1", 0, "C", true, 0, "")
			pdf.CellFormat(25, 8, "Type", "1", 0, "C", true, 0, "")
			pdf.CellFormat(35, 8, "Balance", "1", 0, "C", true, 0, "")
			pdf.CellFormat(30, 8, "Status", "1", 0, "C", true, 0, "")
			pdf.Ln(8)
			pdf.SetFont("Arial", "", 9)
			pdf.SetFillColor(255, 255, 255)
		}

		// Account data
		status := "Active"
		if !account.IsActive {
			status = "Inactive"
		}

		balance := fmt.Sprintf("%.2f", account.Balance)
		
		pdf.CellFormat(30, 6, account.Code, "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 6, account.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, account.Type, "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 6, balance, "1", 0, "R", false, 0, "")
		pdf.CellFormat(30, 6, status, "1", 0, "C", false, 0, "")
		pdf.Ln(6)
	}

	// Output to buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, utils.NewInternalError("Failed to generate PDF", err)
	}

	return buf.Bytes(), nil
}

// ExportAccountsExcel exports accounts to Excel format
func (s *ExportServiceImpl) ExportAccountsExcel(ctx context.Context) ([]byte, error) {
	// Get all accounts from database
	accounts, err := s.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, utils.NewInternalError("Failed to fetch accounts", err)
	}

	// Create new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := "Chart of Accounts"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, utils.NewInternalError("Failed to create Excel sheet", err)
	}

	// Set active sheet
	f.SetActiveSheet(index)

	// Set title
	f.SetCellValue(sheetName, "A1", "Chart of Accounts")
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("Generated on: %s", time.Now().Format("2006-01-02 15:04:05")))

	// Headers
	headers := []string{"Code", "Name", "Type", "Category", "Balance", "Status", "Description", "Created At"}
	for i, header := range headers {
		cell := string(rune('A'+i)) + "4"
		f.SetCellValue(sheetName, cell, header)
	}

	// Style for headers
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D3D3D3"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return nil, utils.NewInternalError("Failed to create Excel style", err)
	}

	// Apply style to headers
	f.SetCellStyle(sheetName, "A4", "H4", headerStyle)

	// Data style
	dataStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return nil, utils.NewInternalError("Failed to create data style", err)
	}

	// Fill data
	for i, account := range accounts {
		row := i + 5 // Start from row 5 (after headers)
		
		status := "Active"
		if !account.IsActive {
			status = "Inactive"
		}

		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), account.Code)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), account.Name)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), account.Type)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), account.Category)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(row), account.Balance)
		f.SetCellValue(sheetName, "F"+strconv.Itoa(row), status)
		f.SetCellValue(sheetName, "G"+strconv.Itoa(row), account.Description)
		f.SetCellValue(sheetName, "H"+strconv.Itoa(row), account.CreatedAt.Format("2006-01-02"))

		// Apply style to data row
		cellRange := "A" + strconv.Itoa(row) + ":H" + strconv.Itoa(row)
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "H"+strconv.Itoa(row), dataStyle)
		_ = cellRange
	}

	// Auto-fit columns
	cols := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
	for _, col := range cols {
		f.SetColWidth(sheetName, col, col, 15)
	}
	
	// Make name column wider
	f.SetColWidth(sheetName, "B", "B", 25)
	f.SetColWidth(sheetName, "G", "G", 30) // Description column

	// Delete default Sheet1 if it exists
	if f.GetSheetName(0) == "Sheet1" {
		f.DeleteSheet("Sheet1")
	}

	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, utils.NewInternalError("Failed to write Excel file", err)
	}

	return buf.Bytes(), nil
}
