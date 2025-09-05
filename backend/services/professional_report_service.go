package services

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// ProfessionalReportService provides enhanced reporting capabilities with
// professional PDF and Excel exports for financial and operational reports
type ProfessionalReportService struct {
	db              *gorm.DB
	accountRepo     repositories.AccountRepository
	salesRepo       *repositories.SalesRepository
	purchaseRepo    *repositories.PurchaseRepository
	productRepo     *repositories.ProductRepository
	contactRepo     repositories.ContactRepository
	paymentRepo     *repositories.PaymentRepository
	cashBankRepo    *repositories.CashBankRepository
	companyProfile  *models.CompanyProfile
}

// ReportTheme contains styling configurations for the reports
type ReportTheme struct {
	PrimaryColor       [3]int
	SecondaryColor     [3]int
	HeaderColor        [3]int
	FooterColor        [3]int
	AlternateRowColor  [3]int
	HeaderFont         string
	BodyFont           string
	Currency           string
	LogoPath           string
	ShowPageNumbers    bool
	PageNumberFormat   string
	DateFormat         string
	DecimalPlaces      int
	ThousandsSeparator string
}

// ReportDataProvider is a flexible data structure that can be used for both
// financial statements and tabular reports
type ReportDataProvider struct {
	Title           string
	Subtitle        string
	PeriodText      string
	HeaderData      map[string]string
	FooterText      string
	ColumnHeaders   []string
	ColumnWidths    []float64
	ColumnAlignments []string
	Rows            [][]interface{}
	SummaryData     map[string]interface{}
	GroupData       []GroupedReportData
	ChartData       []ChartData
	Notes           []string
}

// GroupedReportData represents grouped data for reports
type GroupedReportData struct {
	GroupName     string
	GroupTotal    float64
	GroupSubtotals map[string]float64
	Items         []map[string]interface{}
}

// ChartData represents data for generating charts in reports
type ChartData struct {
	ChartType string // "bar", "line", "pie"
	Title     string
	Labels    []string
	Series    []ChartSeries
}

// ChartSeries represents a series of data points for charts
type ChartSeries struct {
	Name   string
	Values []float64
	Color  string
}

// DefaultTheme returns the default report theme
func DefaultTheme() ReportTheme {
	return ReportTheme{
		PrimaryColor:      [3]int{0, 123, 255},  // Blue
		SecondaryColor:    [3]int{40, 167, 69},  // Green
		HeaderColor:       [3]int{52, 58, 64},   // Dark gray
		FooterColor:       [3]int{233, 236, 239}, // Light gray
		AlternateRowColor: [3]int{248, 249, 250}, // Very light gray
		HeaderFont:        "Arial",
		BodyFont:          "Arial",
		Currency:          "Rp",
		ShowPageNumbers:   true,
		PageNumberFormat:  "Page %d of %d",
		DateFormat:        "02 January 2006",
		DecimalPlaces:     2,
		ThousandsSeparator: ".",
	}
}

// NewProfessionalReportService creates a new professional report service
func NewProfessionalReportService(
	db *gorm.DB,
	accountRepo repositories.AccountRepository,
	salesRepo *repositories.SalesRepository,
	purchaseRepo *repositories.PurchaseRepository,
	productRepo *repositories.ProductRepository,
	contactRepo repositories.ContactRepository,
	paymentRepo *repositories.PaymentRepository,
	cashBankRepo *repositories.CashBankRepository,
) *ProfessionalReportService {
	service := &ProfessionalReportService{
		db:           db,
		accountRepo:  accountRepo,
		salesRepo:    salesRepo,
		purchaseRepo: purchaseRepo,
		productRepo:  productRepo,
		contactRepo:  contactRepo,
		paymentRepo:  paymentRepo,
		cashBankRepo: cashBankRepo,
	}
	
	// Load company profile
	service.loadCompanyProfile()
	
	return service
}

// loadCompanyProfile loads the company profile for report headers
func (prs *ProfessionalReportService) loadCompanyProfile() {
	var profile models.CompanyProfile
	if err := prs.db.First(&profile).Error; err != nil {
		// Create default profile if none exists
		profile = models.CompanyProfile{
			Name:     "Your Company Name",
			Address:  "Company Address",
			City:     "City",
			State:    "State",
			Country:  "Indonesia",
			PostalCode: "12345",
			Phone:    "+62-21-1234567",
			Email:    "contact@company.com",
			Website:  "www.company.com",
			Currency: "IDR",
			FiscalYearStart: "01-01", // January 1st
		}
		prs.db.Create(&profile)
	}
	prs.companyProfile = &profile
}

// FormatCurrency formats a number as currency
func (prs *ProfessionalReportService) FormatCurrency(amount float64) string {
	// Format with thousand separators and 2 decimal places
	intPart := int(amount)
	decPart := int(amount*100) % 100
	
	// Format integer part with thousand separators
	strInt := strconv.Itoa(intPart)
	var result string
	
	for i := len(strInt) - 1; i >= 0; i-- {
		if (len(strInt)-i-1)%3 == 0 && i < len(strInt)-1 {
			result = "." + result
		}
		result = string(strInt[i]) + result
	}
	
	// Add decimal part
	if decPart > 0 {
		result = fmt.Sprintf("%s,%02d", result, decPart)
	}
	
	return fmt.Sprintf("%s %s", prs.companyProfile.Currency, result)
}

// GenerateBalanceSheetPDF generates a professional balance sheet PDF
func (prs *ProfessionalReportService) GenerateBalanceSheetPDF(asOfDate time.Time) ([]byte, error) {
	theme := DefaultTheme()
	
	// Get all accounts with their balances
	ctx := context.Background()
	accounts, err := prs.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}
	
	// Prepare data structure for balance sheet
	var assets []models.Account
	var liabilities []models.Account
	var equity []models.Account
	
	var totalAssets float64
	var totalLiabilities float64
	var totalEquity float64
	
	// Categorize accounts and calculate totals
	for _, account := range accounts {
		balance := prs.calculateAccountBalance(account.ID, asOfDate)
		
		account.Balance = balance
		
		switch account.Type {
		case models.AccountTypeAsset:
			assets = append(assets, account)
			totalAssets += balance
		case models.AccountTypeLiability:
			liabilities = append(liabilities, account)
			totalLiabilities += balance
		case models.AccountTypeEquity:
			equity = append(equity, account)
			totalEquity += balance
		}
	}
	
	// Sort accounts by code
	sort.Slice(assets, func(i, j int) bool { return assets[i].Code < assets[j].Code })
	sort.Slice(liabilities, func(i, j int) bool { return liabilities[i].Code < liabilities[j].Code })
	sort.Slice(equity, func(i, j int) bool { return equity[i].Code < equity[j].Code })
	
	// Create PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Set up fonts
	pdf.SetFont(theme.HeaderFont, "B", 16)
	
	// Company header
	pdf.Cell(190, 10, prs.companyProfile.Name)
	pdf.Ln(7)
	pdf.SetFont(theme.BodyFont, "", 10)
	pdf.Cell(190, 5, prs.companyProfile.Address)
	pdf.Ln(5)
	pdf.Cell(190, 5, fmt.Sprintf("%s, %s %s", prs.companyProfile.City, prs.companyProfile.State, prs.companyProfile.PostalCode))
	pdf.Ln(5)
	pdf.Cell(190, 5, fmt.Sprintf("Tel: %s | Email: %s", prs.companyProfile.Phone, prs.companyProfile.Email))
	pdf.Ln(10)
	
	// Report title
	pdf.SetFont(theme.HeaderFont, "B", 14)
	pdf.Cell(190, 10, "BALANCE SHEET")
	pdf.Ln(7)
	pdf.SetFont(theme.BodyFont, "", 10)
	pdf.Cell(190, 6, fmt.Sprintf("As of %s", asOfDate.Format("January 2, 2006")))
	pdf.Ln(7)
	pdf.Cell(190, 6, fmt.Sprintf("Currency: %s", prs.companyProfile.Currency))
	pdf.Ln(10)
	
	// Assets section
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "ASSETS", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	
	// Reset fill color for data rows
	pdf.SetFillColor(255, 255, 255)
	
	// Current Assets
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(190, 7, "Current Assets")
	pdf.Ln(7)
	
	pdf.SetFont(theme.BodyFont, "", 9)
	var currentAssets float64
	for _, asset := range assets {
		if asset.Category == models.CategoryCurrentAsset && !asset.IsHeader {
			pdf.CellFormat(120, 6, "  "+asset.Name, "0", 0, "L", false, 0, "")
			pdf.CellFormat(70, 6, prs.FormatCurrency(asset.Balance), "0", 1, "R", false, 0, "")
			currentAssets += asset.Balance
		}
	}
	
	pdf.SetFont(theme.BodyFont, "B", 9)
	pdf.Cell(120, 6, "Total Current Assets")
	pdf.CellFormat(70, 6, prs.FormatCurrency(currentAssets), "T", 1, "R", false, 0, "")
	pdf.Ln(5)
	
	// Fixed Assets
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(190, 7, "Fixed Assets")
	pdf.Ln(7)
	
	pdf.SetFont(theme.BodyFont, "", 9)
	var fixedAssets float64
	for _, asset := range assets {
		if asset.Category == models.CategoryFixedAsset && !asset.IsHeader {
			pdf.CellFormat(120, 6, "  "+asset.Name, "0", 0, "L", false, 0, "")
			pdf.CellFormat(70, 6, prs.FormatCurrency(asset.Balance), "0", 1, "R", false, 0, "")
			fixedAssets += asset.Balance
		}
	}
	
	pdf.SetFont(theme.BodyFont, "B", 9)
	pdf.Cell(120, 6, "Total Fixed Assets")
	pdf.CellFormat(70, 6, prs.FormatCurrency(fixedAssets), "T", 1, "R", false, 0, "")
	pdf.Ln(5)
	
	// Total Assets
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(120, 6, "TOTAL ASSETS")
	pdf.SetFillColor(theme.SecondaryColor[0], theme.SecondaryColor[1], theme.SecondaryColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(70, 6, prs.FormatCurrency(totalAssets), "1", 1, "R", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(10)
	
	// Liabilities section
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "LIABILITIES", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	
	// Reset fill color for data rows
	pdf.SetFillColor(255, 255, 255)
	
	// Current Liabilities
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(190, 7, "Current Liabilities")
	pdf.Ln(7)
	
	pdf.SetFont(theme.BodyFont, "", 9)
	var currentLiabilities float64
	for _, liability := range liabilities {
		if liability.Category == models.CategoryCurrentLiability && !liability.IsHeader {
			pdf.CellFormat(120, 6, "  "+liability.Name, "0", 0, "L", false, 0, "")
			pdf.CellFormat(70, 6, prs.FormatCurrency(liability.Balance), "0", 1, "R", false, 0, "")
			currentLiabilities += liability.Balance
		}
	}
	
	pdf.SetFont(theme.BodyFont, "B", 9)
	pdf.Cell(120, 6, "Total Current Liabilities")
	pdf.CellFormat(70, 6, prs.FormatCurrency(currentLiabilities), "T", 1, "R", false, 0, "")
	pdf.Ln(5)
	
	// Long-term Liabilities
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(190, 7, "Long-term Liabilities")
	pdf.Ln(7)
	
	pdf.SetFont(theme.BodyFont, "", 9)
	var longTermLiabilities float64
	for _, liability := range liabilities {
		if liability.Category == models.CategoryLongTermLiability && !liability.IsHeader {
			pdf.CellFormat(120, 6, "  "+liability.Name, "0", 0, "L", false, 0, "")
			pdf.CellFormat(70, 6, prs.FormatCurrency(liability.Balance), "0", 1, "R", false, 0, "")
			longTermLiabilities += liability.Balance
		}
	}
	
	pdf.SetFont(theme.BodyFont, "B", 9)
	pdf.Cell(120, 6, "Total Long-term Liabilities")
	pdf.CellFormat(70, 6, prs.FormatCurrency(longTermLiabilities), "T", 1, "R", false, 0, "")
	pdf.Ln(5)
	
	// Total Liabilities
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(120, 6, "TOTAL LIABILITIES")
	pdf.CellFormat(70, 6, prs.FormatCurrency(totalLiabilities), "T", 1, "R", false, 0, "")
	pdf.Ln(10)
	
	// Equity section
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "EQUITY", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	
	// Reset fill color for data rows
	pdf.SetFillColor(255, 255, 255)
	
	// Equity items
	pdf.SetFont(theme.BodyFont, "", 9)
	for _, eq := range equity {
		if !eq.IsHeader {
			pdf.CellFormat(120, 6, "  "+eq.Name, "0", 0, "L", false, 0, "")
			pdf.CellFormat(70, 6, prs.FormatCurrency(eq.Balance), "0", 1, "R", false, 0, "")
		}
	}
	
	// Total Equity
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(120, 6, "TOTAL EQUITY")
	pdf.CellFormat(70, 6, prs.FormatCurrency(totalEquity), "T", 1, "R", false, 0, "")
	pdf.Ln(10)
	
	// Total Liabilities & Equity
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(120, 6, "TOTAL LIABILITIES & EQUITY")
	pdf.SetFillColor(theme.SecondaryColor[0], theme.SecondaryColor[1], theme.SecondaryColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(70, 6, prs.FormatCurrency(totalLiabilities+totalEquity), "1", 1, "R", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	
	// Add footer
	_, pageHeight := pdf.GetPageSize()
	pdf.SetY(pageHeight - 20)
	pdf.SetFont(theme.BodyFont, "I", 8)
	pdf.CellFormat(190, 10, fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006 15:04:05")), "T", 1, "C", false, 0, "")
	
	if theme.ShowPageNumbers {
		pdf.SetY(pageHeight - 10)
		pdf.CellFormat(190, 10, fmt.Sprintf("Page %d of %d", pdf.PageNo(), pdf.PageCount()), "", 0, "C", false, 0, "")
	}
	
	// Output to buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}
	
	return buf.Bytes(), nil
}

// GenerateBalanceSheetCSV generates a professional balance sheet CSV file
func (prs *ProfessionalReportService) GenerateBalanceSheetCSV(asOfDate time.Time) ([]byte, error) {
	// Get all accounts with their balances
	ctx := context.Background()
	accounts, err := prs.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}
	
	// Prepare data structure for balance sheet
	var assets []models.Account
	var liabilities []models.Account
	var equity []models.Account
	
	var totalAssets float64
	var totalLiabilities float64
	var totalEquity float64
	
	// Categorize accounts and calculate totals
	for _, account := range accounts {
		balance := prs.calculateAccountBalance(account.ID, asOfDate)
		
		account.Balance = balance
		
		switch account.Type {
		case models.AccountTypeAsset:
			assets = append(assets, account)
			totalAssets += balance
		case models.AccountTypeLiability:
			liabilities = append(liabilities, account)
			totalLiabilities += balance
		case models.AccountTypeEquity:
			equity = append(equity, account)
			totalEquity += balance
		}
	}
	
	// Sort accounts by code
	sort.Slice(assets, func(i, j int) bool { return assets[i].Code < assets[j].Code })
	sort.Slice(liabilities, func(i, j int) bool { return liabilities[i].Code < liabilities[j].Code })
	sort.Slice(equity, func(i, j int) bool { return equity[i].Code < equity[j].Code })
	
	// Create CSV content
	var buf bytes.Buffer
	
	// Write company header
	buf.WriteString(fmt.Sprintf("%s\n", prs.companyProfile.Name))
	buf.WriteString(fmt.Sprintf("%s\n", prs.companyProfile.Address))
	buf.WriteString(fmt.Sprintf("%s, %s %s\n", prs.companyProfile.City, prs.companyProfile.State, prs.companyProfile.PostalCode))
	buf.WriteString(fmt.Sprintf("Tel: %s | Email: %s\n", prs.companyProfile.Phone, prs.companyProfile.Email))
	buf.WriteString("\n")
	
	// Write report title
	buf.WriteString("BALANCE SHEET\n")
	buf.WriteString(fmt.Sprintf("As of %s\n", asOfDate.Format("January 2, 2006")))
	buf.WriteString(fmt.Sprintf("Currency: %s\n", prs.companyProfile.Currency))
	buf.WriteString("\n")
	
	// Write CSV headers
	buf.WriteString("Section,Account Code,Account Name,Amount\n")
	
	// ASSETS SECTION
	buf.WriteString(fmt.Sprintf("ASSETS,,,%s\n", prs.formatCSVAmount(0)))
	
	// Current Assets
	buf.WriteString("Current Assets,,,\n")
	var currentAssets float64
	for _, asset := range assets {
		if asset.Category == models.CategoryCurrentAsset && !asset.IsHeader {
			buf.WriteString(fmt.Sprintf(",\"%s\",\"%s\",%s\n", asset.Code, asset.Name, prs.formatCSVAmount(asset.Balance)))
			currentAssets += asset.Balance
		}
	}
	buf.WriteString(fmt.Sprintf(",,,Total Current Assets,%s\n", prs.formatCSVAmount(currentAssets)))
	buf.WriteString("\n")
	
	// Fixed Assets
	buf.WriteString("Fixed Assets,,,\n")
	var fixedAssets float64
	for _, asset := range assets {
		if asset.Category == models.CategoryFixedAsset && !asset.IsHeader {
			buf.WriteString(fmt.Sprintf(",\"%s\",\"%s\",%s\n", asset.Code, asset.Name, prs.formatCSVAmount(asset.Balance)))
			fixedAssets += asset.Balance
		}
	}
	buf.WriteString(fmt.Sprintf(",,,Total Fixed Assets,%s\n", prs.formatCSVAmount(fixedAssets)))
	buf.WriteString("\n")
	
	// Total Assets
	buf.WriteString(fmt.Sprintf(",,,TOTAL ASSETS,%s\n", prs.formatCSVAmount(totalAssets)))
	buf.WriteString("\n")
	
	// LIABILITIES SECTION
	buf.WriteString("LIABILITIES,,,\n")
	
	// Current Liabilities
	buf.WriteString("Current Liabilities,,,\n")
	var currentLiabilities float64
	for _, liability := range liabilities {
		if liability.Category == models.CategoryCurrentLiability && !liability.IsHeader {
			buf.WriteString(fmt.Sprintf(",\"%s\",\"%s\",%s\n", liability.Code, liability.Name, prs.formatCSVAmount(liability.Balance)))
			currentLiabilities += liability.Balance
		}
	}
	buf.WriteString(fmt.Sprintf(",,,Total Current Liabilities,%s\n", prs.formatCSVAmount(currentLiabilities)))
	buf.WriteString("\n")
	
	// Long-term Liabilities
	buf.WriteString("Long-term Liabilities,,,\n")
	var longTermLiabilities float64
	for _, liability := range liabilities {
		if liability.Category == models.CategoryLongTermLiability && !liability.IsHeader {
			buf.WriteString(fmt.Sprintf(",\"%s\",\"%s\",%s\n", liability.Code, liability.Name, prs.formatCSVAmount(liability.Balance)))
			longTermLiabilities += liability.Balance
		}
	}
	buf.WriteString(fmt.Sprintf(",,,Total Long-term Liabilities,%s\n", prs.formatCSVAmount(longTermLiabilities)))
	buf.WriteString("\n")
	
	// Total Liabilities
	buf.WriteString(fmt.Sprintf(",,,TOTAL LIABILITIES,%s\n", prs.formatCSVAmount(totalLiabilities)))
	buf.WriteString("\n")
	
	// EQUITY SECTION
	buf.WriteString("EQUITY,,,\n")
	for _, eq := range equity {
		if !eq.IsHeader {
			buf.WriteString(fmt.Sprintf(",\"%s\",\"%s\",%s\n", eq.Code, eq.Name, prs.formatCSVAmount(eq.Balance)))
		}
	}
	buf.WriteString(fmt.Sprintf(",,,TOTAL EQUITY,%s\n", prs.formatCSVAmount(totalEquity)))
	buf.WriteString("\n")
	
	// Total Liabilities & Equity
	buf.WriteString(fmt.Sprintf(",,,TOTAL LIABILITIES & EQUITY,%s\n", prs.formatCSVAmount(totalLiabilities+totalEquity)))
	buf.WriteString("\n")
	
	// Footer
	buf.WriteString(fmt.Sprintf("Generated on %s\n", time.Now().Format("January 2, 2006 15:04:05")))
	
	return buf.Bytes(), nil
}

// GenerateBalanceSheetExcel generates a professional balance sheet Excel file
func (prs *ProfessionalReportService) GenerateBalanceSheetExcel(asOfDate time.Time) ([]byte, error) {
	// Get all accounts with their balances
	ctx := context.Background()
	accounts, err := prs.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}
	
	// Prepare data structure for balance sheet
	var assets []models.Account
	var liabilities []models.Account
	var equity []models.Account
	
	var totalAssets float64
	var totalLiabilities float64
	var totalEquity float64
	
	// Categorize accounts and calculate totals
	for _, account := range accounts {
		balance := prs.calculateAccountBalance(account.ID, asOfDate)
		
		account.Balance = balance
		
		switch account.Type {
		case models.AccountTypeAsset:
			assets = append(assets, account)
			totalAssets += balance
		case models.AccountTypeLiability:
			liabilities = append(liabilities, account)
			totalLiabilities += balance
		case models.AccountTypeEquity:
			equity = append(equity, account)
			totalEquity += balance
		}
	}
	
	// Sort accounts by code
	sort.Slice(assets, func(i, j int) bool { return assets[i].Code < assets[j].Code })
	sort.Slice(liabilities, func(i, j int) bool { return liabilities[i].Code < liabilities[j].Code })
	sort.Slice(equity, func(i, j int) bool { return equity[i].Code < equity[j].Code })
	
	// Create Excel file
	f := excelize.NewFile()
	sheetName := "Balance Sheet"
	
	// Create sheet and make it active
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create Excel sheet: %v", err)
	}
	f.SetActiveSheet(index)
	
	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 15)
	f.SetColWidth(sheetName, "B", "B", 40)
	f.SetColWidth(sheetName, "C", "C", 20)
	f.SetColWidth(sheetName, "D", "D", 20)
	
	// Company header
	f.SetCellValue(sheetName, "A1", prs.companyProfile.Name)
	f.SetCellValue(sheetName, "A2", prs.companyProfile.Address)
	f.SetCellValue(sheetName, "A3", fmt.Sprintf("%s, %s %s", prs.companyProfile.City, prs.companyProfile.State, prs.companyProfile.PostalCode))
	f.SetCellValue(sheetName, "A4", fmt.Sprintf("Tel: %s | Email: %s", prs.companyProfile.Phone, prs.companyProfile.Email))
	
	// Company header style
	companyStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   14,
			Family: "Arial",
		},
	})
	f.SetCellStyle(sheetName, "A1", "A1", companyStyle)
	
	// Report title
	f.SetCellValue(sheetName, "A6", "BALANCE SHEET")
	f.SetCellValue(sheetName, "A7", fmt.Sprintf("As of %s", asOfDate.Format("January 2, 2006")))
	f.SetCellValue(sheetName, "A8", fmt.Sprintf("Currency: %s", prs.companyProfile.Currency))
	
	// Title style
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   14,
			Family: "Arial",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	f.SetCellStyle(sheetName, "A6", "D6", titleStyle)
	f.MergeCell(sheetName, "A6", "D6")
	
	// Subtitle style
	subtitleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Arial",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	f.SetCellStyle(sheetName, "A7", "D7", subtitleStyle)
	f.MergeCell(sheetName, "A7", "D7")
	f.SetCellStyle(sheetName, "A8", "D8", subtitleStyle)
	f.MergeCell(sheetName, "A8", "D8")
	
	// Section header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  "FFFFFF",
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4169E1"}, // Royal blue
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	
	// Sub-section header style
	subHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   11,
			Family: "Arial",
		},
	})
	
	// Total style with border on top
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   10,
			Family: "Arial",
		},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
		},
	})
	
	// Grand total style
	grandTotalStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   11,
			Color:  "FFFFFF",
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"2E8B57"}, // SeaGreen
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "right",
		},
	})
	
	// Current row
	row := 10
	
	// ASSETS SECTION
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "ASSETS")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerStyle)
	row++
	
	// Current Assets
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Current Assets")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), subHeaderStyle)
	row++
	
	var currentAssets float64
	for _, asset := range assets {
		if asset.Category == models.CategoryCurrentAsset && !asset.IsHeader {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), asset.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), asset.Balance)
			currentAssets += asset.Balance
			row++
		}
	}
	
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Total Current Assets")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), currentAssets)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), totalStyle)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), totalStyle)
	row += 2
	
	// Fixed Assets
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Fixed Assets")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), subHeaderStyle)
	row++
	
	var fixedAssets float64
	for _, asset := range assets {
		if asset.Category == models.CategoryFixedAsset && !asset.IsHeader {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), asset.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), asset.Balance)
			fixedAssets += asset.Balance
			row++
		}
	}
	
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Total Fixed Assets")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), fixedAssets)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), totalStyle)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), totalStyle)
	row += 2
	
	// Total Assets
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "TOTAL ASSETS")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), totalAssets)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), grandTotalStyle)
	row += 2
	
	// LIABILITIES SECTION
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "LIABILITIES")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerStyle)
	row++
	
	// Current Liabilities
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Current Liabilities")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), subHeaderStyle)
	row++
	
	var currentLiabilities float64
	for _, liability := range liabilities {
		if liability.Category == models.CategoryCurrentLiability && !liability.IsHeader {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), liability.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), liability.Balance)
			currentLiabilities += liability.Balance
			row++
		}
	}
	
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Total Current Liabilities")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), currentLiabilities)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), totalStyle)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), totalStyle)
	row += 2
	
	// Long-term Liabilities
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Long-term Liabilities")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), subHeaderStyle)
	row++
	
	var longTermLiabilities float64
	for _, liability := range liabilities {
		if liability.Category == models.CategoryLongTermLiability && !liability.IsHeader {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), liability.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), liability.Balance)
			longTermLiabilities += liability.Balance
			row++
		}
	}
	
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Total Long-term Liabilities")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), longTermLiabilities)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), totalStyle)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), totalStyle)
	row += 2
	
	// Total Liabilities
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "TOTAL LIABILITIES")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), totalLiabilities)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), totalStyle)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), totalStyle)
	row += 2
	
	// EQUITY SECTION
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "EQUITY")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerStyle)
	row++
	
	for _, eq := range equity {
		if !eq.IsHeader {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), eq.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), eq.Balance)
			row++
		}
	}
	
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "TOTAL EQUITY")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), totalEquity)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), totalStyle)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), totalStyle)
	row += 2
	
	// Total Liabilities & Equity
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "TOTAL LIABILITIES & EQUITY")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), totalLiabilities+totalEquity)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), grandTotalStyle)
	
	// Footer
	row += 3
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006 15:04:05")))
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	
	// Format the currency cells
	currencyStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 44, // Currency format with thousand separators
	})
	
	// Apply currency style to all amount cells
	for i := 11; i < row; i++ {
		cellRef := fmt.Sprintf("D%d", i)
		f.SetCellStyle(sheetName, cellRef, cellRef, currencyStyle)
	}
	
	// Delete the default sheet
	f.DeleteSheet("Sheet1")
	
	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %v", err)
	}
	
	return buf.Bytes(), nil
}

// GenerateProfitLossJSON generates a professional profit and loss statement in JSON format
func (prs *ProfessionalReportService) GenerateProfitLossJSON(startDate, endDate time.Time) (map[string]interface{}, error) {
	// Get all accounts with their balances
	ctx := context.Background()
	accounts, err := prs.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}

	// Prepare data structure for profit & loss
	var revenues []models.Account
	var expenses []models.Account

	var totalRevenue float64
	var totalExpenses float64

	// Categorize accounts and calculate totals
	for _, account := range accounts {
		balance := prs.calculateAccountBalanceForPeriod(account.ID, startDate, endDate)

		// Skip accounts with no activity
		if balance == 0 {
			continue
		}

		account.Balance = balance

		switch account.Type {
		case models.AccountTypeRevenue:
			revenues = append(revenues, account)
			totalRevenue += balance
		case models.AccountTypeExpense:
			expenses = append(expenses, account)
			totalExpenses += balance
		}
	}

	// Sort accounts by code
	sort.Slice(revenues, func(i, j int) bool { return revenues[i].Code < revenues[j].Code })
	sort.Slice(expenses, func(i, j int) bool { return expenses[i].Code < expenses[j].Code })

	// Calculate gross profit and net income
	grossProfit := totalRevenue - totalExpenses
	netIncome := grossProfit // Simplified

	// Categorize revenues and expenses
	var operatingRevenue, otherRevenue float64
	var operatingExpenses, otherExpenses float64
	var operatingRevenueAccounts, otherRevenueAccounts []map[string]interface{}
	var operatingExpenseAccounts, otherExpenseAccounts []map[string]interface{}

	for _, revenue := range revenues {
		accountData := map[string]interface{}{
			"id": revenue.ID,
			"code": revenue.Code,
			"name": revenue.Name,
			"balance": revenue.Balance,
			"formatted_balance": prs.FormatCurrency(revenue.Balance),
		}
		if revenue.Category == models.CategoryOperatingRevenue {
			operatingRevenue += revenue.Balance
			operatingRevenueAccounts = append(operatingRevenueAccounts, accountData)
		} else {
			otherRevenue += revenue.Balance
			otherRevenueAccounts = append(otherRevenueAccounts, accountData)
		}
	}

	for _, expense := range expenses {
		accountData := map[string]interface{}{
			"id": expense.ID,
			"code": expense.Code,
			"name": expense.Name,
			"balance": expense.Balance,
			"formatted_balance": prs.FormatCurrency(expense.Balance),
		}
		if expense.Category == models.CategoryOperatingExpense {
			operatingExpenses += expense.Balance
			operatingExpenseAccounts = append(operatingExpenseAccounts, accountData)
		} else {
			otherExpenses += expense.Balance
			otherExpenseAccounts = append(otherExpenseAccounts, accountData)
		}
	}

	// Build JSON response
	jsonResponse := map[string]interface{}{
		"report_title": "Profit & Loss Statement",
		"company": map[string]interface{}{
			"name": prs.companyProfile.Name,
			"address": prs.companyProfile.Address,
			"city": prs.companyProfile.City,
			"state": prs.companyProfile.State,
			"postal_code": prs.companyProfile.PostalCode,
			"phone": prs.companyProfile.Phone,
			"email": prs.companyProfile.Email,
			"currency": prs.companyProfile.Currency,
		},
		"period": map[string]interface{}{
			"start_date": startDate.Format("2006-01-02"),
			"end_date": endDate.Format("2006-01-02"),
			"description": fmt.Sprintf("For the period %s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")),
		},
		"revenue": map[string]interface{}{
			"operating": map[string]interface{}{
				"accounts": operatingRevenueAccounts,
				"total": operatingRevenue,
				"formatted_total": prs.FormatCurrency(operatingRevenue),
			},
			"other": map[string]interface{}{
				"accounts": otherRevenueAccounts,
				"total": otherRevenue,
				"formatted_total": prs.FormatCurrency(otherRevenue),
			},
			"total": totalRevenue,
			"formatted_total": prs.FormatCurrency(totalRevenue),
		},
		"expenses": map[string]interface{}{
			"operating": map[string]interface{}{
				"accounts": operatingExpenseAccounts,
				"total": operatingExpenses,
				"formatted_total": prs.FormatCurrency(operatingExpenses),
			},
			"other": map[string]interface{}{
				"accounts": otherExpenseAccounts,
				"total": otherExpenses,
				"formatted_total": prs.FormatCurrency(otherExpenses),
			},
			"total": totalExpenses,
			"formatted_total": prs.FormatCurrency(totalExpenses),
		},
		"summary": map[string]interface{}{
			"gross_profit": grossProfit,
			"formatted_gross_profit": prs.FormatCurrency(grossProfit),
			"net_income": netIncome,
			"formatted_net_income": prs.FormatCurrency(netIncome),
		},
		"generated_at": time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}

	return jsonResponse, nil
}

// GenerateProfitLossPDF generates a professional profit and loss statement PDF
func (prs *ProfessionalReportService) GenerateProfitLossPDF(startDate, endDate time.Time) ([]byte, error) {
	theme := DefaultTheme()
	
	// Get all accounts with their balances
	ctx := context.Background()
	accounts, err := prs.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}
	
	// Prepare data structure for profit & loss
	var revenues []models.Account
	var expenses []models.Account
	
	var totalRevenue float64
	var totalExpenses float64
	
	// Categorize accounts and calculate totals
	for _, account := range accounts {
		balance := prs.calculateAccountBalanceForPeriod(account.ID, startDate, endDate)
		
		// Skip accounts with no activity
		if balance == 0 {
			continue
		}
		
		account.Balance = balance
		
		switch account.Type {
		case models.AccountTypeRevenue:
			revenues = append(revenues, account)
			totalRevenue += balance
		case models.AccountTypeExpense:
			expenses = append(expenses, account)
			totalExpenses += balance
		}
	}
	
	// Sort accounts by code
	sort.Slice(revenues, func(i, j int) bool { return revenues[i].Code < revenues[j].Code })
	sort.Slice(expenses, func(i, j int) bool { return expenses[i].Code < expenses[j].Code })
	
	// Calculate gross profit and net income
	grossProfit := totalRevenue - totalExpenses
	netIncome := grossProfit // Simplified
	
	// Create PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Set up fonts
	pdf.SetFont(theme.HeaderFont, "B", 16)
	
	// Company header
	pdf.Cell(190, 10, prs.companyProfile.Name)
	pdf.Ln(7)
	pdf.SetFont(theme.BodyFont, "", 10)
	pdf.Cell(190, 5, prs.companyProfile.Address)
	pdf.Ln(5)
	pdf.Cell(190, 5, fmt.Sprintf("%s, %s %s", prs.companyProfile.City, prs.companyProfile.State, prs.companyProfile.PostalCode))
	pdf.Ln(5)
	pdf.Cell(190, 5, fmt.Sprintf("Tel: %s | Email: %s", prs.companyProfile.Phone, prs.companyProfile.Email))
	pdf.Ln(10)
	
	// Report title
	pdf.SetFont(theme.HeaderFont, "B", 14)
	pdf.Cell(190, 10, "PROFIT & LOSS STATEMENT")
	pdf.Ln(7)
	pdf.SetFont(theme.BodyFont, "", 10)
	pdf.Cell(190, 6, fmt.Sprintf("For the period %s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	pdf.Ln(7)
	pdf.Cell(190, 6, fmt.Sprintf("Currency: %s", prs.companyProfile.Currency))
	pdf.Ln(10)
	
	// Revenue section
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "REVENUE", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	
	// Reset fill color for data rows
	pdf.SetFillColor(255, 255, 255)
	
	// Operating Revenue
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(190, 7, "Operating Revenue")
	pdf.Ln(7)
	
	pdf.SetFont(theme.BodyFont, "", 9)
	var operatingRevenue float64
	for _, revenue := range revenues {
		if revenue.Category == models.CategoryOperatingRevenue && !revenue.IsHeader {
			pdf.CellFormat(120, 6, "  "+revenue.Name, "0", 0, "L", false, 0, "")
			pdf.CellFormat(70, 6, prs.FormatCurrency(revenue.Balance), "0", 1, "R", false, 0, "")
			operatingRevenue += revenue.Balance
		}
	}
	
	pdf.SetFont(theme.BodyFont, "B", 9)
	pdf.Cell(120, 6, "Total Operating Revenue")
	pdf.CellFormat(70, 6, prs.FormatCurrency(operatingRevenue), "T", 1, "R", false, 0, "")
	pdf.Ln(5)
	
	// Other Revenue
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(190, 7, "Other Revenue")
	pdf.Ln(7)
	
	pdf.SetFont(theme.BodyFont, "", 9)
	var otherRevenue float64
	for _, revenue := range revenues {
	if revenue.Category == models.CategoryOtherIncome && !revenue.IsHeader {
			pdf.CellFormat(120, 6, "  "+revenue.Name, "0", 0, "L", false, 0, "")
			pdf.CellFormat(70, 6, prs.FormatCurrency(revenue.Balance), "0", 1, "R", false, 0, "")
			otherRevenue += revenue.Balance
		}
	}
	
	pdf.SetFont(theme.BodyFont, "B", 9)
	pdf.Cell(120, 6, "Total Other Revenue")
	pdf.CellFormat(70, 6, prs.FormatCurrency(otherRevenue), "T", 1, "R", false, 0, "")
	pdf.Ln(5)
	
	// Total Revenue
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(120, 6, "TOTAL REVENUE")
	pdf.SetFillColor(theme.SecondaryColor[0], theme.SecondaryColor[1], theme.SecondaryColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(70, 6, prs.FormatCurrency(totalRevenue), "1", 1, "R", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(10)
	
	// Expenses section
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "EXPENSES", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	
	// Reset fill color for data rows
	pdf.SetFillColor(255, 255, 255)
	
	// Operating Expenses
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(190, 7, "Operating Expenses")
	pdf.Ln(7)
	
	pdf.SetFont(theme.BodyFont, "", 9)
	var operatingExpenses float64
	for _, expense := range expenses {
		if expense.Category == models.CategoryOperatingExpense && !expense.IsHeader {
			pdf.CellFormat(120, 6, "  "+expense.Name, "0", 0, "L", false, 0, "")
			pdf.CellFormat(70, 6, prs.FormatCurrency(expense.Balance), "0", 1, "R", false, 0, "")
			operatingExpenses += expense.Balance
		}
	}
	
	pdf.SetFont(theme.BodyFont, "B", 9)
	pdf.Cell(120, 6, "Total Operating Expenses")
	pdf.CellFormat(70, 6, prs.FormatCurrency(operatingExpenses), "T", 1, "R", false, 0, "")
	pdf.Ln(5)
	
	// Other Expenses
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(190, 7, "Other Expenses")
	pdf.Ln(7)
	
	pdf.SetFont(theme.BodyFont, "", 9)
	var otherExpenses float64
	for _, expense := range expenses {
		if expense.Category == models.CategoryOtherExpense && !expense.IsHeader {
			pdf.CellFormat(120, 6, "  "+expense.Name, "0", 0, "L", false, 0, "")
			pdf.CellFormat(70, 6, prs.FormatCurrency(expense.Balance), "0", 1, "R", false, 0, "")
			otherExpenses += expense.Balance
		}
	}
	
	pdf.SetFont(theme.BodyFont, "B", 9)
	pdf.Cell(120, 6, "Total Other Expenses")
	pdf.CellFormat(70, 6, prs.FormatCurrency(otherExpenses), "T", 1, "R", false, 0, "")
	pdf.Ln(5)
	
	// Total Expenses
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(120, 6, "TOTAL EXPENSES")
	pdf.SetFillColor(theme.SecondaryColor[0], theme.SecondaryColor[1], theme.SecondaryColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(70, 6, prs.FormatCurrency(totalExpenses), "1", 1, "R", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(10)
	
	// Summary section
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "SUMMARY", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)
	
	// Gross Profit
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(120, 6, "Gross Profit")
	pdf.CellFormat(70, 6, prs.FormatCurrency(grossProfit), "0", 1, "R", false, 0, "")
	pdf.Ln(5)
	
	// Net Income
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(120, 6, "NET INCOME")
	pdf.SetFillColor(theme.PrimaryColor[0], theme.PrimaryColor[1], theme.PrimaryColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(70, 6, prs.FormatCurrency(netIncome), "1", 1, "R", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	
	// Add footer
	_, pageHeight := pdf.GetPageSize()
	pdf.SetY(pageHeight - 20)
	pdf.SetFont(theme.BodyFont, "I", 8)
	pdf.CellFormat(190, 10, fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006 15:04:05")), "T", 1, "C", false, 0, "")
	
	if theme.ShowPageNumbers {
		pdf.SetY(pageHeight - 10)
		pdf.CellFormat(190, 10, fmt.Sprintf("Page %d of %d", pdf.PageNo(), pdf.PageCount()), "", 0, "C", false, 0, "")
	}
	
	// Output to buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}
	
	return buf.Bytes(), nil
}

// GenerateProfitLossExcel generates a professional profit and loss Excel file
func (prs *ProfessionalReportService) GenerateProfitLossExcel(startDate, endDate time.Time) ([]byte, error) {
	// Get all accounts with their balances
	ctx := context.Background()
	accounts, err := prs.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}
	
	// Prepare data structure for profit & loss
	var revenues []models.Account
	var expenses []models.Account
	
	var totalRevenue float64
	var totalExpenses float64
	
	// Categorize accounts and calculate totals
	for _, account := range accounts {
		balance := prs.calculateAccountBalanceForPeriod(account.ID, startDate, endDate)
		
		// Skip accounts with no activity
		if balance == 0 {
			continue
		}
		
		account.Balance = balance
		
		switch account.Type {
		case models.AccountTypeRevenue:
			revenues = append(revenues, account)
			totalRevenue += balance
		case models.AccountTypeExpense:
			expenses = append(expenses, account)
			totalExpenses += balance
		}
	}
	
	// Sort accounts by code
	sort.Slice(revenues, func(i, j int) bool { return revenues[i].Code < revenues[j].Code })
	sort.Slice(expenses, func(i, j int) bool { return expenses[i].Code < expenses[j].Code })
	
	// Calculate gross profit and net income
	grossProfit := totalRevenue - totalExpenses
	netIncome := grossProfit // Simplified
	
	// Create Excel file
	f := excelize.NewFile()
	sheetName := "Profit & Loss"
	
	// Create sheet and make it active
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create Excel sheet: %v", err)
	}
	f.SetActiveSheet(index)
	
	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 15)
	f.SetColWidth(sheetName, "B", "B", 40)
	f.SetColWidth(sheetName, "C", "C", 20)
	f.SetColWidth(sheetName, "D", "D", 20)
	
	// Company header
	f.SetCellValue(sheetName, "A1", prs.companyProfile.Name)
	f.SetCellValue(sheetName, "A2", prs.companyProfile.Address)
	f.SetCellValue(sheetName, "A3", fmt.Sprintf("%s, %s %s", prs.companyProfile.City, prs.companyProfile.State, prs.companyProfile.PostalCode))
	f.SetCellValue(sheetName, "A4", fmt.Sprintf("Tel: %s | Email: %s", prs.companyProfile.Phone, prs.companyProfile.Email))
	
	// Company header style
	companyStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   14,
			Family: "Arial",
		},
	})
	f.SetCellStyle(sheetName, "A1", "A1", companyStyle)
	
	// Report title
	f.SetCellValue(sheetName, "A6", "PROFIT & LOSS STATEMENT")
	f.SetCellValue(sheetName, "A7", fmt.Sprintf("For the period %s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	f.SetCellValue(sheetName, "A8", fmt.Sprintf("Currency: %s", prs.companyProfile.Currency))
	
	// Title style
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   14,
			Family: "Arial",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	f.SetCellStyle(sheetName, "A6", "D6", titleStyle)
	f.MergeCell(sheetName, "A6", "D6")
	
	// Subtitle style
	subtitleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Arial",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	f.SetCellStyle(sheetName, "A7", "D7", subtitleStyle)
	f.MergeCell(sheetName, "A7", "D7")
	f.SetCellStyle(sheetName, "A8", "D8", subtitleStyle)
	f.MergeCell(sheetName, "A8", "D8")
	
	// Section header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  "FFFFFF",
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4169E1"}, // Royal blue
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	
	// Sub-section header style
	subHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   11,
			Family: "Arial",
		},
	})
	
	// Total style with border on top
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   10,
			Family: "Arial",
		},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
		},
	})
	
	// Grand total style
	grandTotalStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   11,
			Color:  "FFFFFF",
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"2E8B57"}, // SeaGreen
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "right",
		},
	})
	
	// Current row
	row := 10
	
	// REVENUE SECTION
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "REVENUE")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerStyle)
	row++
	
	// Operating Revenue
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Operating Revenue")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), subHeaderStyle)
	row++
	
	var operatingRevenue float64
	for _, revenue := range revenues {
		if revenue.Category == models.CategoryOperatingRevenue && !revenue.IsHeader {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), revenue.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), revenue.Balance)
			operatingRevenue += revenue.Balance
			row++
		}
	}
	
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Total Operating Revenue")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), operatingRevenue)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), totalStyle)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), totalStyle)
	row += 2
	
	// Other Revenue
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Other Revenue")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), subHeaderStyle)
	row++
	
	var otherRevenue float64
	for _, revenue := range revenues {
	if revenue.Category == models.CategoryOtherIncome && !revenue.IsHeader {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), revenue.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), revenue.Balance)
			otherRevenue += revenue.Balance
			row++
		}
	}
	
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Total Other Revenue")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), otherRevenue)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), totalStyle)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), totalStyle)
	row += 2
	
	// Total Revenue
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "TOTAL REVENUE")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), totalRevenue)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), grandTotalStyle)
	row += 2
	
	// EXPENSES SECTION
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "EXPENSES")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerStyle)
	row++
	
	// Operating Expenses
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Operating Expenses")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), subHeaderStyle)
	row++
	
	var operatingExpenses float64
	for _, expense := range expenses {
		if expense.Category == models.CategoryOperatingExpense && !expense.IsHeader {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), expense.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), expense.Balance)
			operatingExpenses += expense.Balance
			row++
		}
	}
	
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Total Operating Expenses")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), operatingExpenses)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), totalStyle)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), totalStyle)
	row += 2
	
	// Other Expenses
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Other Expenses")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), subHeaderStyle)
	row++
	
	var otherExpenses float64
	for _, expense := range expenses {
		if expense.Category == models.CategoryOtherExpense && !expense.IsHeader {
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), expense.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), expense.Balance)
			otherExpenses += expense.Balance
			row++
		}
	}
	
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Total Other Expenses")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), otherExpenses)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), totalStyle)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), totalStyle)
	row += 2
	
	// Total Expenses
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "TOTAL EXPENSES")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), totalExpenses)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), grandTotalStyle)
	row += 2
	
	// SUMMARY SECTION
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "SUMMARY")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerStyle)
	row++
	
	// Gross Profit
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Gross Profit")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), grossProfit)
	row += 2
	
	// Net Income
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "NET INCOME")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), netIncome)
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), grandTotalStyle)
	
	// Footer
	row += 3
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006 15:04:05")))
	f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	
	// Format the currency cells
	currencyStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 44, // Currency format with thousand separators
	})
	
	// Apply currency style to all amount cells
	for i := 11; i < row; i++ {
		cellRef := fmt.Sprintf("D%d", i)
		f.SetCellStyle(sheetName, cellRef, cellRef, currencyStyle)
	}
	
	// Delete the default sheet
	f.DeleteSheet("Sheet1")
	
	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %v", err)
	}
	
	return buf.Bytes(), nil
}

// Helper methods

// calculateAccountBalance calculates the account balance as of a specific date
func (prs *ProfessionalReportService) calculateAccountBalance(accountID uint, asOfDate time.Time) float64 {
	// Get account starting balance (opening balance)
	var account models.Account
	if err := prs.db.First(&account, accountID).Error; err != nil {
		return 0
	}
	
	// Calculate balance from journal entries up to asOfDate
	var totalDebits, totalCredits float64
	prs.db.Table("journal_entries").
		Joins("JOIN journals ON journal_entries.journal_id = journals.id").
		Where("journal_entries.account_id = ? AND journals.date <= ? AND journals.status = 'POSTED'", accountID, asOfDate).
		Select("COALESCE(SUM(journal_entries.debit_amount), 0) as total_debits, COALESCE(SUM(journal_entries.credit_amount), 0) as total_credits").
		Row().Scan(&totalDebits, &totalCredits)
	
	// Calculate balance based on account normal balance type
	switch account.Type {
	case models.AccountTypeAsset, models.AccountTypeExpense:
		// Assets and Expenses: Debit increases, Credit decreases
		return account.Balance + totalDebits - totalCredits
	case models.AccountTypeLiability, models.AccountTypeEquity, models.AccountTypeRevenue:
		// Liabilities, Equity, Revenue: Credit increases, Debit decreases
		return account.Balance + totalCredits - totalDebits
	default:
		return account.Balance
	}
}

// calculateAccountBalanceForPeriod calculates the account balance for a specific period
func (prs *ProfessionalReportService) calculateAccountBalanceForPeriod(accountID uint, startDate, endDate time.Time) float64 {
	// Calculate balance from journal entries for the specific period only
	var account models.Account
	if err := prs.db.First(&account, accountID).Error; err != nil {
		return 0
	}
	
	var totalDebits, totalCredits float64
	prs.db.Table("journal_entries").
		Joins("JOIN journals ON journal_entries.journal_id = journals.id").
		Where("journal_entries.account_id = ? AND journals.date BETWEEN ? AND ? AND journals.status = 'POSTED'", accountID, startDate, endDate).
		Select("COALESCE(SUM(journal_entries.debit_amount), 0) as total_debits, COALESCE(SUM(journal_entries.credit_amount), 0) as total_credits").
		Row().Scan(&totalDebits, &totalCredits)
	
	// For P&L accounts, we want the period activity
	switch account.Type {
	case models.AccountTypeRevenue:
		// Revenue: Credit is positive
		return totalCredits - totalDebits
	case models.AccountTypeExpense:
		// Expenses: Debit is positive
		return totalDebits - totalCredits
	default:
		// For balance sheet accounts, return cumulative balance
		return prs.calculateAccountBalance(accountID, endDate)
	}
}

// GenerateSalesSummaryPDF creates a sales summary report in PDF format
func (prs *ProfessionalReportService) GenerateSalesSummaryPDF(startDate, endDate time.Time, groupBy string) ([]byte, error) {
	theme := DefaultTheme()
	
	// Query sales data
	var sales []models.Sale
	if err := prs.db.Preload("Customer").
		Preload("SaleItems").
		Preload("SaleItems.Product").
		Where("sales.date BETWEEN ? AND ?", startDate, endDate).
		Find(&sales).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch sales data: %v", err)
	}
	
	// Calculate totals
	var totalSales float64
	salesByCustomer := make(map[uint]float64)
	salesByProduct := make(map[uint]struct {
		Name    string
		Quantity int64
		Amount  float64
	})
	salesByStatus := make(map[string]int)
	salesByPeriod := make(map[string]float64)
	
	for _, sale := range sales {
		totalSales += sale.TotalAmount
		
		// Sales by customer
		salesByCustomer[sale.CustomerID] += sale.TotalAmount
		
		// Sales by status
		salesByStatus[sale.Status]++
		
		// Sales by period
		var period string
		switch groupBy {
		case "month":
			period = sale.Date.Format("2006-01")
		case "quarter":
			quarter := (sale.Date.Month()-1)/3 + 1
			period = fmt.Sprintf("%d-Q%d", sale.Date.Year(), quarter)
		case "year":
			period = sale.Date.Format("2006")
		default:
			period = sale.Date.Format("2006-01-02")
		}
		salesByPeriod[period] += sale.TotalAmount
		
		// Sales by product
		for _, item := range sale.SaleItems {
			productInfo := salesByProduct[item.ProductID]
			productInfo.Name = item.Product.Name
			productInfo.Quantity += int64(item.Quantity)
			productInfo.Amount += item.LineTotal
			salesByProduct[item.ProductID] = productInfo
		}
	}
	
	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Set up fonts
	pdf.SetFont(theme.HeaderFont, "B", 16)
	
	// Company header
	pdf.Cell(190, 10, prs.companyProfile.Name)
	pdf.Ln(7)
	pdf.SetFont(theme.BodyFont, "", 10)
	pdf.Cell(190, 5, prs.companyProfile.Address)
	pdf.Ln(5)
	pdf.Cell(190, 5, fmt.Sprintf("%s, %s %s", prs.companyProfile.City, prs.companyProfile.State, prs.companyProfile.PostalCode))
	pdf.Ln(5)
	pdf.Cell(190, 5, fmt.Sprintf("Tel: %s | Email: %s", prs.companyProfile.Phone, prs.companyProfile.Email))
	pdf.Ln(10)
	
	// Report title
	pdf.SetFont(theme.HeaderFont, "B", 14)
	pdf.Cell(190, 10, "SALES SUMMARY REPORT")
	pdf.Ln(7)
	pdf.SetFont(theme.BodyFont, "", 10)
	pdf.Cell(190, 6, fmt.Sprintf("For the period %s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	pdf.Ln(7)
	pdf.Cell(190, 6, fmt.Sprintf("Currency: %s", prs.companyProfile.Currency))
	pdf.Ln(10)
	
	// Sales Summary Section
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "SALES SUMMARY", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)
	
	// Key metrics
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(100, 6, "Total Sales:")
	pdf.CellFormat(90, 6, prs.FormatCurrency(totalSales), "0", 1, "R", false, 0, "")
	
	pdf.Cell(100, 6, "Total Transactions:")
	pdf.CellFormat(90, 6, fmt.Sprintf("%d", len(sales)), "0", 1, "R", false, 0, "")
	
	avgOrderValue := 0.0
	if len(sales) > 0 {
		avgOrderValue = totalSales / float64(len(sales))
	}
	pdf.Cell(100, 6, "Average Order Value:")
	pdf.CellFormat(90, 6, prs.FormatCurrency(avgOrderValue), "0", 1, "R", false, 0, "")
	pdf.Ln(10)
	
	// Sales by Period
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "SALES BY PERIOD", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)
	
	// Column headers
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.CellFormat(95, 6, "Period", "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, "Amount", "1", 1, "R", false, 0, "")
	
	// Sales data by period
	pdf.SetFont(theme.BodyFont, "", 9)
	for period, amount := range salesByPeriod {
		pdf.CellFormat(95, 6, period, "1", 0, "L", false, 0, "")
		pdf.CellFormat(95, 6, prs.FormatCurrency(amount), "1", 1, "R", false, 0, "")
	}
	pdf.Ln(10)
	
	// Top Customers
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "TOP CUSTOMERS", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)
	
	// Column headers
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.CellFormat(95, 6, "Customer", "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, "Total Sales", "1", 1, "R", false, 0, "")
	
	// Convert map to slice for sorting
	type CustomerSales struct {
		ID        uint
		Name      string
		TotalSales float64
	}
	
	var customerSalesSlice []CustomerSales
	for customerID, amount := range salesByCustomer {
		var customerName string
		for _, sale := range sales {
			if sale.CustomerID == customerID {
				customerName = sale.Customer.Name
				break
			}
		}
		customerSalesSlice = append(customerSalesSlice, CustomerSales{
			ID:        customerID,
			Name:      customerName,
			TotalSales: amount,
		})
	}
	
	// Sort by total sales (descending)
	sort.Slice(customerSalesSlice, func(i, j int) bool {
		return customerSalesSlice[i].TotalSales > customerSalesSlice[j].TotalSales
	})
	
	// Take top 5 customers
	topCustomers := customerSalesSlice
	if len(topCustomers) > 5 {
		topCustomers = topCustomers[:5]
	}
	
	// Display top customers
	pdf.SetFont(theme.BodyFont, "", 9)
	for _, cs := range topCustomers {
		pdf.CellFormat(95, 6, cs.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(95, 6, prs.FormatCurrency(cs.TotalSales), "1", 1, "R", false, 0, "")
	}
	pdf.Ln(10)
	
	// Top Products
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "TOP PRODUCTS", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)
	
	// Column headers
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.CellFormat(75, 6, "Product", "1", 0, "L", false, 0, "")
	pdf.CellFormat(40, 6, "Quantity", "1", 0, "C", false, 0, "")
	pdf.CellFormat(75, 6, "Amount", "1", 1, "R", false, 0, "")
	
	// Convert map to slice for sorting
	type ProductSales struct {
		ID        uint
		Name      string
		Quantity  int64
		Amount    float64
	}
	
	var productSalesSlice []ProductSales
	for productID, info := range salesByProduct {
		productSalesSlice = append(productSalesSlice, ProductSales{
			ID:       productID,
			Name:     info.Name,
			Quantity: info.Quantity,
			Amount:   info.Amount,
		})
	}
	
	// Sort by total sales amount (descending)
	sort.Slice(productSalesSlice, func(i, j int) bool {
		return productSalesSlice[i].Amount > productSalesSlice[j].Amount
	})
	
	// Take top 5 products
	topProducts := productSalesSlice
	if len(topProducts) > 5 {
		topProducts = topProducts[:5]
	}
	
	// Display top products
	pdf.SetFont(theme.BodyFont, "", 9)
	for _, ps := range topProducts {
		pdf.CellFormat(75, 6, ps.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("%d", ps.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(75, 6, prs.FormatCurrency(ps.Amount), "1", 1, "R", false, 0, "")
	}
	
	// Add footer
	_, pageHeight := pdf.GetPageSize()
	pdf.SetY(pageHeight - 20)
	pdf.SetFont(theme.BodyFont, "I", 8)
	pdf.CellFormat(190, 10, fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006 15:04:05")), "T", 1, "C", false, 0, "")
	
	if theme.ShowPageNumbers {
		pdf.SetY(pageHeight - 10)
		pdf.CellFormat(190, 10, fmt.Sprintf("Page %d of %d", pdf.PageNo(), pdf.PageCount()), "", 0, "C", false, 0, "")
	}
	
	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}
	
	return buf.Bytes(), nil
}

// GenerateSalesSummaryExcel creates a sales summary report in Excel format
func (prs *ProfessionalReportService) GenerateSalesSummaryExcel(startDate, endDate time.Time, groupBy string) ([]byte, error) {
	// Query sales data
	var sales []models.Sale
	if err := prs.db.Preload("Customer").
		Preload("SaleItems").
		Preload("SaleItems.Product").
		Where("sales.date BETWEEN ? AND ?", startDate, endDate).
		Find(&sales).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch sales data: %v", err)
	}
	
	// Calculate totals
	var totalSales float64
	salesByCustomer := make(map[uint]float64)
	salesByProduct := make(map[uint]struct {
		Name    string
		Quantity int64
		Amount  float64
	})
	salesByStatus := make(map[string]int)
	salesByPeriod := make(map[string]float64)
	
	for _, sale := range sales {
		totalSales += sale.TotalAmount
		
		// Sales by customer
		salesByCustomer[sale.CustomerID] += sale.TotalAmount
		
		// Sales by status
		salesByStatus[sale.Status]++
		
		// Sales by period
		var period string
		switch groupBy {
		case "month":
			period = sale.Date.Format("2006-01")
		case "quarter":
			quarter := (sale.Date.Month()-1)/3 + 1
			period = fmt.Sprintf("%d-Q%d", sale.Date.Year(), quarter)
		case "year":
			period = sale.Date.Format("2006")
		default:
			period = sale.Date.Format("2006-01-02")
		}
		salesByPeriod[period] += sale.TotalAmount
		
		// Sales by product
		for _, item := range sale.SaleItems {
			productInfo := salesByProduct[item.ProductID]
			productInfo.Name = item.Product.Name
			productInfo.Quantity += int64(item.Quantity)
			productInfo.Amount += item.LineTotal
			salesByProduct[item.ProductID] = productInfo
		}
	}
	
	// Create Excel file
	f := excelize.NewFile()
	
	// Add Summary sheet
	summarySheet := "Sales Summary"
	index, err := f.NewSheet(summarySheet)
	if err != nil {
		return nil, fmt.Errorf("failed to create Excel sheet: %v", err)
	}
	f.SetActiveSheet(index)
	
	// Set column widths
	f.SetColWidth(summarySheet, "A", "A", 20)
	f.SetColWidth(summarySheet, "B", "B", 40)
	f.SetColWidth(summarySheet, "C", "C", 15)
	f.SetColWidth(summarySheet, "D", "D", 20)
	
	// Company header
	f.SetCellValue(summarySheet, "A1", prs.companyProfile.Name)
	f.SetCellValue(summarySheet, "A2", prs.companyProfile.Address)
	f.SetCellValue(summarySheet, "A3", fmt.Sprintf("%s, %s %s", prs.companyProfile.City, prs.companyProfile.State, prs.companyProfile.PostalCode))
	f.SetCellValue(summarySheet, "A4", fmt.Sprintf("Tel: %s | Email: %s", prs.companyProfile.Phone, prs.companyProfile.Email))
	
	// Company header style
	companyStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   14,
			Family: "Arial",
		},
	})
	f.SetCellStyle(summarySheet, "A1", "A1", companyStyle)
	
	// Report title
	f.SetCellValue(summarySheet, "A6", "SALES SUMMARY REPORT")
	f.SetCellValue(summarySheet, "A7", fmt.Sprintf("For the period %s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	f.SetCellValue(summarySheet, "A8", fmt.Sprintf("Currency: %s", prs.companyProfile.Currency))
	
	// Title style
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   14,
			Family: "Arial",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	f.SetCellStyle(summarySheet, "A6", "D6", titleStyle)
	f.MergeCell(summarySheet, "A6", "D6")
	
	// Subtitle style
	subtitleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Arial",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	f.SetCellStyle(summarySheet, "A7", "D7", subtitleStyle)
	f.MergeCell(summarySheet, "A7", "D7")
	f.SetCellStyle(summarySheet, "A8", "D8", subtitleStyle)
	f.MergeCell(summarySheet, "A8", "D8")
	
	// Section header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  "FFFFFF",
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4169E1"}, // Royal blue
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	
	// Table header style
	tableHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   10,
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"E0E0E0"}, // Light gray
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	
	// Data style
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	
	// Currency style
	currencyStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 44, // Currency format with thousand separators
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "right",
		},
	})
	
	// Current row
	row := 10
	
	// Key Metrics Section
	f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), "SALES SUMMARY")
	f.MergeCell(summarySheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(summarySheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerStyle)
	row++
	
	// Key metrics
	f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), "Total Sales:")
	f.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), totalSales)
	f.SetCellStyle(summarySheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), currencyStyle)
	row++
	
	f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), "Total Transactions:")
	f.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), len(sales))
	row++
	
	avgOrderValue := 0.0
	if len(sales) > 0 {
		avgOrderValue = totalSales / float64(len(sales))
	}
	f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), "Average Order Value:")
	f.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), avgOrderValue)
	f.SetCellStyle(summarySheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), currencyStyle)
	row += 2
	
	// Sales by Period Section
	f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), "SALES BY PERIOD")
	f.MergeCell(summarySheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(summarySheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerStyle)
	row++
	
	// Column headers
	f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), "Period")
	f.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), "Amount")
	f.SetCellStyle(summarySheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), tableHeaderStyle)
	row++
	
	// Convert map to slice for sorting
	type PeriodSales struct {
		Period string
		Amount float64
	}
	
	var periodSalesSlice []PeriodSales
	for period, amount := range salesByPeriod {
		periodSalesSlice = append(periodSalesSlice, PeriodSales{
			Period: period,
			Amount: amount,
		})
	}
	
	// Sort by period
	sort.Slice(periodSalesSlice, func(i, j int) bool {
		return periodSalesSlice[i].Period < periodSalesSlice[j].Period
	})
	
	// Sales data by period
	for _, ps := range periodSalesSlice {
		f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), ps.Period)
		f.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), ps.Amount)
		f.SetCellStyle(summarySheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), dataStyle)
		f.SetCellStyle(summarySheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), currencyStyle)
		row++
	}
	row++
	
	// Add "By Customer" sheet
	customerSheet := "By Customer"
	f.NewSheet(customerSheet)
	
	// Title
	f.SetCellValue(customerSheet, "A1", "SALES BY CUSTOMER")
	f.MergeCell(customerSheet, "A1", "C1")
	f.SetCellStyle(customerSheet, "A1", "C1", titleStyle)
	
	// Set column widths
	f.SetColWidth(customerSheet, "A", "A", 10)
	f.SetColWidth(customerSheet, "B", "B", 40)
	f.SetColWidth(customerSheet, "C", "C", 20)
	
	// Column headers
	f.SetCellValue(customerSheet, "A3", "Rank")
	f.SetCellValue(customerSheet, "B3", "Customer")
	f.SetCellValue(customerSheet, "C3", "Total Sales")
	f.SetCellStyle(customerSheet, "A3", "C3", tableHeaderStyle)
	
	// Convert map to slice for sorting
	type CustomerSales struct {
		ID        uint
		Name      string
		TotalSales float64
	}
	
	var customerSalesSlice []CustomerSales
	for customerID, amount := range salesByCustomer {
		var customerName string
		for _, sale := range sales {
			if sale.CustomerID == customerID {
				customerName = sale.Customer.Name
				break
			}
		}
		customerSalesSlice = append(customerSalesSlice, CustomerSales{
			ID:        customerID,
			Name:      customerName,
			TotalSales: amount,
		})
	}
	
	// Sort by total sales (descending)
	sort.Slice(customerSalesSlice, func(i, j int) bool {
		return customerSalesSlice[i].TotalSales > customerSalesSlice[j].TotalSales
	})
	
	// Customer data
	for i, cs := range customerSalesSlice {
		row := i + 4 // Start from row 4
		f.SetCellValue(customerSheet, fmt.Sprintf("A%d", row), i+1) // Rank
		f.SetCellValue(customerSheet, fmt.Sprintf("B%d", row), cs.Name)
		f.SetCellValue(customerSheet, fmt.Sprintf("C%d", row), cs.TotalSales)
		f.SetCellStyle(customerSheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), dataStyle)
		f.SetCellStyle(customerSheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), dataStyle)
		f.SetCellStyle(customerSheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), currencyStyle)
	}
	
	// Add "By Product" sheet
	productSheet := "By Product"
	f.NewSheet(productSheet)
	
	// Title
	f.SetCellValue(productSheet, "A1", "SALES BY PRODUCT")
	f.MergeCell(productSheet, "A1", "D1")
	f.SetCellStyle(productSheet, "A1", "D1", titleStyle)
	
	// Set column widths
	f.SetColWidth(productSheet, "A", "A", 10)
	f.SetColWidth(productSheet, "B", "B", 40)
	f.SetColWidth(productSheet, "C", "C", 15)
	f.SetColWidth(productSheet, "D", "D", 20)
	
	// Column headers
	f.SetCellValue(productSheet, "A3", "Rank")
	f.SetCellValue(productSheet, "B3", "Product")
	f.SetCellValue(productSheet, "C3", "Quantity")
	f.SetCellValue(productSheet, "D3", "Amount")
	f.SetCellStyle(productSheet, "A3", "D3", tableHeaderStyle)
	
	// Convert map to slice for sorting
	type ProductSales struct {
		ID        uint
		Name      string
		Quantity  int64
		Amount    float64
	}
	
	var productSalesSlice []ProductSales
	for productID, info := range salesByProduct {
		productSalesSlice = append(productSalesSlice, ProductSales{
			ID:       productID,
			Name:     info.Name,
			Quantity: info.Quantity,
			Amount:   info.Amount,
		})
	}
	
	// Sort by total sales amount (descending)
	sort.Slice(productSalesSlice, func(i, j int) bool {
		return productSalesSlice[i].Amount > productSalesSlice[j].Amount
	})
	
	// Product data
	for i, ps := range productSalesSlice {
		row := i + 4 // Start from row 4
		f.SetCellValue(productSheet, fmt.Sprintf("A%d", row), i+1) // Rank
		f.SetCellValue(productSheet, fmt.Sprintf("B%d", row), ps.Name)
		f.SetCellValue(productSheet, fmt.Sprintf("C%d", row), ps.Quantity)
		f.SetCellValue(productSheet, fmt.Sprintf("D%d", row), ps.Amount)
		f.SetCellStyle(productSheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), dataStyle)
		f.SetCellStyle(productSheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), dataStyle)
		f.SetCellStyle(productSheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), dataStyle)
		f.SetCellStyle(productSheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), currencyStyle)
	}
	
	// Add "Transactions" sheet with detailed information
	transactionSheet := "Transactions"
	f.NewSheet(transactionSheet)
	
	// Title
	f.SetCellValue(transactionSheet, "A1", "SALES TRANSACTIONS")
	f.MergeCell(transactionSheet, "A1", "G1")
	f.SetCellStyle(transactionSheet, "A1", "G1", titleStyle)
	
	// Set column widths
	f.SetColWidth(transactionSheet, "A", "A", 15) // Invoice/Code
	f.SetColWidth(transactionSheet, "B", "B", 15) // Date
	f.SetColWidth(transactionSheet, "C", "C", 30) // Customer
	f.SetColWidth(transactionSheet, "D", "D", 15) // Amount
	f.SetColWidth(transactionSheet, "E", "E", 15) // Tax
	f.SetColWidth(transactionSheet, "F", "F", 15) // Total
	f.SetColWidth(transactionSheet, "G", "G", 15) // Status
	
	// Column headers
	f.SetCellValue(transactionSheet, "A3", "Invoice No")
	f.SetCellValue(transactionSheet, "B3", "Date")
	f.SetCellValue(transactionSheet, "C3", "Customer")
	f.SetCellValue(transactionSheet, "D3", "Subtotal")
	f.SetCellValue(transactionSheet, "E3", "Tax")
	f.SetCellValue(transactionSheet, "F3", "Total")
	f.SetCellValue(transactionSheet, "G3", "Status")
	f.SetCellStyle(transactionSheet, "A3", "G3", tableHeaderStyle)
	
	// Sort sales by date (newest first)
	sort.Slice(sales, func(i, j int) bool {
		return sales[i].Date.After(sales[j].Date)
	})
	
	// Sales transactions
	for i, sale := range sales {
		row := i + 4 // Start from row 4
		f.SetCellValue(transactionSheet, fmt.Sprintf("A%d", row), sale.Code)
		f.SetCellValue(transactionSheet, fmt.Sprintf("B%d", row), sale.Date.Format("2006-01-02"))
		f.SetCellValue(transactionSheet, fmt.Sprintf("C%d", row), sale.Customer.Name)
		f.SetCellValue(transactionSheet, fmt.Sprintf("D%d", row), sale.Subtotal)
		f.SetCellValue(transactionSheet, fmt.Sprintf("E%d", row), sale.TotalTax)
		f.SetCellValue(transactionSheet, fmt.Sprintf("F%d", row), sale.TotalAmount)
		f.SetCellValue(transactionSheet, fmt.Sprintf("G%d", row), sale.Status)
		
		// Apply styles
		f.SetCellStyle(transactionSheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), dataStyle)
		f.SetCellStyle(transactionSheet, fmt.Sprintf("D%d", row), fmt.Sprintf("F%d", row), currencyStyle)
		f.SetCellStyle(transactionSheet, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), dataStyle)
	}
	
	// Delete the default sheet
	f.DeleteSheet("Sheet1")
	
	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %v", err)
	}
	
	return buf.Bytes(), nil
}

// GeneratePurchaseSummaryPDF creates a purchase summary report in PDF format
func (prs *ProfessionalReportService) GeneratePurchaseSummaryPDF(startDate, endDate time.Time, groupBy string) ([]byte, error) {
	theme := DefaultTheme()
	
	// Query purchase data
	var purchases []models.Purchase
	if err := prs.db.Preload("Vendor").
		Preload("PurchaseItems").
		Preload("PurchaseItems.Product").
		Where("purchases.date BETWEEN ? AND ?", startDate, endDate).
		Find(&purchases).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch purchase data: %v", err)
	}
	
	// Calculate totals
	var totalPurchases float64
	purchasesByVendor := make(map[uint]float64)
	purchasesByProduct := make(map[uint]struct {
		Name    string
		Quantity int
		Amount  float64
	})
	purchasesByStatus := make(map[string]int)
	purchasesByPeriod := make(map[string]float64)
	
	for _, purchase := range purchases {
		totalPurchases += purchase.TotalAmount
		
		// Purchases by vendor
		purchasesByVendor[purchase.VendorID] += purchase.TotalAmount
		
		// Purchases by status
		purchasesByStatus[purchase.Status]++
		
		// Purchases by period
		var period string
		switch groupBy {
		case "month":
			period = purchase.Date.Format("2006-01")
		case "quarter":
			quarter := (purchase.Date.Month()-1)/3 + 1
			period = fmt.Sprintf("%d-Q%d", purchase.Date.Year(), quarter)
		case "year":
			period = purchase.Date.Format("2006")
		default:
			period = purchase.Date.Format("2006-01-02")
		}
		purchasesByPeriod[period] += purchase.TotalAmount
		
		// Purchases by product
		for _, item := range purchase.PurchaseItems {
			productInfo := purchasesByProduct[item.ProductID]
			productInfo.Name = item.Product.Name
			productInfo.Quantity += item.Quantity
			productInfo.Amount += item.TotalPrice
			purchasesByProduct[item.ProductID] = productInfo
		}
	}
	
	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Set up fonts
	pdf.SetFont(theme.HeaderFont, "B", 16)
	
	// Company header
	pdf.Cell(190, 10, prs.companyProfile.Name)
	pdf.Ln(7)
	pdf.SetFont(theme.BodyFont, "", 10)
	pdf.Cell(190, 5, prs.companyProfile.Address)
	pdf.Ln(5)
	pdf.Cell(190, 5, fmt.Sprintf("%s, %s %s", prs.companyProfile.City, prs.companyProfile.State, prs.companyProfile.PostalCode))
	pdf.Ln(5)
	pdf.Cell(190, 5, fmt.Sprintf("Tel: %s | Email: %s", prs.companyProfile.Phone, prs.companyProfile.Email))
	pdf.Ln(10)
	
	// Report title
	pdf.SetFont(theme.HeaderFont, "B", 14)
	pdf.Cell(190, 10, "PURCHASE SUMMARY REPORT")
	pdf.Ln(7)
	pdf.SetFont(theme.BodyFont, "", 10)
	pdf.Cell(190, 6, fmt.Sprintf("For the period %s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	pdf.Ln(7)
	pdf.Cell(190, 6, fmt.Sprintf("Currency: %s", prs.companyProfile.Currency))
	pdf.Ln(10)
	
	// Purchase Summary Section
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "PURCHASE SUMMARY", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)
	
	// Key metrics
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(100, 6, "Total Purchases:")
	pdf.CellFormat(90, 6, prs.FormatCurrency(totalPurchases), "0", 1, "R", false, 0, "")
	
	pdf.Cell(100, 6, "Total Transactions:")
	pdf.CellFormat(90, 6, fmt.Sprintf("%d", len(purchases)), "0", 1, "R", false, 0, "")
	
	avgOrderValue := 0.0
	if len(purchases) > 0 {
		avgOrderValue = totalPurchases / float64(len(purchases))
	}
	pdf.Cell(100, 6, "Average Order Value:")
	pdf.CellFormat(90, 6, prs.FormatCurrency(avgOrderValue), "0", 1, "R", false, 0, "")
	pdf.Ln(10)
	
	// Purchases by Period
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "PURCHASES BY PERIOD", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)
	
	// Column headers
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.CellFormat(95, 6, "Period", "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, "Amount", "1", 1, "R", false, 0, "")
	
	// Purchases data by period
	pdf.SetFont(theme.BodyFont, "", 9)
	for period, amount := range purchasesByPeriod {
		pdf.CellFormat(95, 6, period, "1", 0, "L", false, 0, "")
		pdf.CellFormat(95, 6, prs.FormatCurrency(amount), "1", 1, "R", false, 0, "")
	}
	pdf.Ln(10)
	
	// Top Vendors
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "TOP VENDORS", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)
	
	// Column headers
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.CellFormat(95, 6, "Vendor", "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, "Total Purchases", "1", 1, "R", false, 0, "")
	
	// Convert map to slice for sorting
	type VendorPurchases struct {
		ID            uint
		Name          string
		TotalPurchases float64
	}
	
	var vendorPurchasesSlice []VendorPurchases
	for vendorID, amount := range purchasesByVendor {
		var vendorName string
		for _, purchase := range purchases {
			if purchase.VendorID == vendorID {
				vendorName = purchase.Vendor.Name
				break
			}
		}
		vendorPurchasesSlice = append(vendorPurchasesSlice, VendorPurchases{
			ID:             vendorID,
			Name:           vendorName,
			TotalPurchases: amount,
		})
	}
	
	// Sort by total purchases (descending)
	sort.Slice(vendorPurchasesSlice, func(i, j int) bool {
		return vendorPurchasesSlice[i].TotalPurchases > vendorPurchasesSlice[j].TotalPurchases
	})
	
	// Take top 5 vendors
	topVendors := vendorPurchasesSlice
	if len(topVendors) > 5 {
		topVendors = topVendors[:5]
	}
	
	// Display top vendors
	pdf.SetFont(theme.BodyFont, "", 9)
	for _, vp := range topVendors {
		pdf.CellFormat(95, 6, vp.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(95, 6, prs.FormatCurrency(vp.TotalPurchases), "1", 1, "R", false, 0, "")
	}
	pdf.Ln(10)
	
	// Top Products
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "TOP PRODUCTS", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)
	
	// Column headers
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.CellFormat(75, 6, "Product", "1", 0, "L", false, 0, "")
	pdf.CellFormat(40, 6, "Quantity", "1", 0, "C", false, 0, "")
	pdf.CellFormat(75, 6, "Amount", "1", 1, "R", false, 0, "")
	
	// Convert map to slice for sorting
	type ProductPurchases struct {
		ID        uint
		Name      string
		Quantity  int
		Amount    float64
	}
	
	var productPurchasesSlice []ProductPurchases
	for productID, info := range purchasesByProduct {
		productPurchasesSlice = append(productPurchasesSlice, ProductPurchases{
			ID:       productID,
			Name:     info.Name,
			Quantity: info.Quantity,
			Amount:   info.Amount,
		})
	}
	
	// Sort by total purchase amount (descending)
	sort.Slice(productPurchasesSlice, func(i, j int) bool {
		return productPurchasesSlice[i].Amount > productPurchasesSlice[j].Amount
	})
	
	// Take top 5 products
	topProducts := productPurchasesSlice
	if len(topProducts) > 5 {
		topProducts = topProducts[:5]
	}
	
	// Display top products
	pdf.SetFont(theme.BodyFont, "", 9)
	for _, pp := range topProducts {
		pdf.CellFormat(75, 6, pp.Name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("%d", pp.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(75, 6, prs.FormatCurrency(pp.Amount), "1", 1, "R", false, 0, "")
	}
	
	// Add footer
	_, pageHeight := pdf.GetPageSize()
	pdf.SetY(pageHeight - 20)
	pdf.SetFont(theme.BodyFont, "I", 8)
	pdf.CellFormat(190, 10, fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006 15:04:05")), "T", 1, "C", false, 0, "")
	
	if theme.ShowPageNumbers {
		_, pageHeight := pdf.GetPageSize()
		pdf.SetY(pageHeight - 10)
		pdf.CellFormat(190, 10, fmt.Sprintf("Page %d of %d", pdf.PageNo(), pdf.PageCount()), "", 0, "C", false, 0, "")
	}
	
	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}
	
	return buf.Bytes(), nil
}

// GeneratePurchaseSummaryExcel creates a purchase summary report in Excel format
func (prs *ProfessionalReportService) GeneratePurchaseSummaryExcel(startDate, endDate time.Time, groupBy string) ([]byte, error) {
	// Query purchase data
	var purchases []models.Purchase
	if err := prs.db.Preload("Vendor").
		Preload("PurchaseItems").
		Preload("PurchaseItems.Product").
		Where("purchases.date BETWEEN ? AND ?", startDate, endDate).
		Find(&purchases).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch purchase data: %v", err)
	}
	
	// Calculate totals
	var totalPurchases float64
	purchasesByVendor := make(map[uint]float64)
	purchasesByProduct := make(map[uint]struct {
		Name    string
		Quantity int
		Amount  float64
	})
	purchasesByStatus := make(map[string]int)
	purchasesByPeriod := make(map[string]float64)
	
	for _, purchase := range purchases {
		totalPurchases += purchase.TotalAmount
		
		// Purchases by vendor
		purchasesByVendor[purchase.VendorID] += purchase.TotalAmount
		
		// Purchases by status
		purchasesByStatus[purchase.Status]++
		
		// Purchases by period
		var period string
		switch groupBy {
		case "month":
			period = purchase.Date.Format("2006-01")
		case "quarter":
			quarter := (purchase.Date.Month()-1)/3 + 1
			period = fmt.Sprintf("%d-Q%d", purchase.Date.Year(), quarter)
		case "year":
			period = purchase.Date.Format("2006")
		default:
			period = purchase.Date.Format("2006-01-02")
		}
		purchasesByPeriod[period] += purchase.TotalAmount
		
		// Purchases by product
		for _, item := range purchase.PurchaseItems {
			productInfo := purchasesByProduct[item.ProductID]
			productInfo.Name = item.Product.Name
			productInfo.Quantity += item.Quantity
			productInfo.Amount += item.TotalPrice
			purchasesByProduct[item.ProductID] = productInfo
		}
	}
	
	// Create Excel file
	f := excelize.NewFile()
	
	// Add Summary sheet
	summarySheet := "Purchase Summary"
	index, err := f.NewSheet(summarySheet)
	if err != nil {
		return nil, fmt.Errorf("failed to create Excel sheet: %v", err)
	}
	f.SetActiveSheet(index)
	
	// Set column widths
	f.SetColWidth(summarySheet, "A", "A", 20)
	f.SetColWidth(summarySheet, "B", "B", 40)
	f.SetColWidth(summarySheet, "C", "C", 15)
	f.SetColWidth(summarySheet, "D", "D", 20)
	
	// Company header
	f.SetCellValue(summarySheet, "A1", prs.companyProfile.Name)
	f.SetCellValue(summarySheet, "A2", prs.companyProfile.Address)
	f.SetCellValue(summarySheet, "A3", fmt.Sprintf("%s, %s %s", prs.companyProfile.City, prs.companyProfile.State, prs.companyProfile.PostalCode))
	f.SetCellValue(summarySheet, "A4", fmt.Sprintf("Tel: %s | Email: %s", prs.companyProfile.Phone, prs.companyProfile.Email))
	
	// Company header style
	companyStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   14,
			Family: "Arial",
		},
	})
	f.SetCellStyle(summarySheet, "A1", "A1", companyStyle)
	
	// Report title
	f.SetCellValue(summarySheet, "A6", "PURCHASE SUMMARY REPORT")
	f.SetCellValue(summarySheet, "A7", fmt.Sprintf("For the period %s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	f.SetCellValue(summarySheet, "A8", fmt.Sprintf("Currency: %s", prs.companyProfile.Currency))
	
	// Title style
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   14,
			Family: "Arial",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	f.SetCellStyle(summarySheet, "A6", "D6", titleStyle)
	f.MergeCell(summarySheet, "A6", "D6")
	
	// Subtitle style
	subtitleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Arial",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	f.SetCellStyle(summarySheet, "A7", "D7", subtitleStyle)
	f.MergeCell(summarySheet, "A7", "D7")
	f.SetCellStyle(summarySheet, "A8", "D8", subtitleStyle)
	f.MergeCell(summarySheet, "A8", "D8")
	
	// Section header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  "FFFFFF",
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4169E1"}, // Royal blue
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	
	// Table header style
	tableHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   10,
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"E0E0E0"}, // Light gray
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	
	// Data style
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	
	// Currency style
	currencyStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 44, // Currency format with thousand separators
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "right",
		},
	})
	
	// Current row
	row := 10
	
	// Key Metrics Section
	f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), "PURCHASE SUMMARY")
	f.MergeCell(summarySheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(summarySheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerStyle)
	row++
	
	// Key metrics
	f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), "Total Purchases:")
	f.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), totalPurchases)
	f.SetCellStyle(summarySheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), currencyStyle)
	row++
	
	f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), "Total Transactions:")
	f.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), len(purchases))
	row++
	
	avgOrderValue := 0.0
	if len(purchases) > 0 {
		avgOrderValue = totalPurchases / float64(len(purchases))
	}
	f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), "Average Order Value:")
	f.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), avgOrderValue)
	f.SetCellStyle(summarySheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), currencyStyle)
	row += 2
	
	// Purchases by Period Section
	f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), "PURCHASES BY PERIOD")
	f.MergeCell(summarySheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	f.SetCellStyle(summarySheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerStyle)
	row++
	
	// Column headers
	f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), "Period")
	f.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), "Amount")
	f.SetCellStyle(summarySheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), tableHeaderStyle)
	row++
	
	// Convert map to slice for sorting
	type PeriodPurchases struct {
		Period string
		Amount float64
	}
	
	var periodPurchasesSlice []PeriodPurchases
	for period, amount := range purchasesByPeriod {
		periodPurchasesSlice = append(periodPurchasesSlice, PeriodPurchases{
			Period: period,
			Amount: amount,
		})
	}
	
	// Sort by period
	sort.Slice(periodPurchasesSlice, func(i, j int) bool {
		return periodPurchasesSlice[i].Period < periodPurchasesSlice[j].Period
	})
	
	// Purchases data by period
	for _, pp := range periodPurchasesSlice {
		f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), pp.Period)
		f.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), pp.Amount)
		f.SetCellStyle(summarySheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), dataStyle)
		f.SetCellStyle(summarySheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), currencyStyle)
		row++
	}
	row++
	
	// Add "By Vendor" sheet
	vendorSheet := "By Vendor"
	f.NewSheet(vendorSheet)
	
	// Title
	f.SetCellValue(vendorSheet, "A1", "PURCHASES BY VENDOR")
	f.MergeCell(vendorSheet, "A1", "C1")
	f.SetCellStyle(vendorSheet, "A1", "C1", titleStyle)
	
	// Set column widths
	f.SetColWidth(vendorSheet, "A", "A", 10)
	f.SetColWidth(vendorSheet, "B", "B", 40)
	f.SetColWidth(vendorSheet, "C", "C", 20)
	
	// Column headers
	f.SetCellValue(vendorSheet, "A3", "Rank")
	f.SetCellValue(vendorSheet, "B3", "Vendor")
	f.SetCellValue(vendorSheet, "C3", "Total Purchases")
	f.SetCellStyle(vendorSheet, "A3", "C3", tableHeaderStyle)
	
	// Convert map to slice for sorting
	type VendorPurchases struct {
		ID            uint
		Name          string
		TotalPurchases float64
	}
	
	var vendorPurchasesSlice []VendorPurchases
	for vendorID, amount := range purchasesByVendor {
		var vendorName string
		for _, purchase := range purchases {
			if purchase.VendorID == vendorID {
				vendorName = purchase.Vendor.Name
				break
			}
		}
		vendorPurchasesSlice = append(vendorPurchasesSlice, VendorPurchases{
			ID:            vendorID,
			Name:          vendorName,
			TotalPurchases: amount,
		})
	}
	
	// Sort by total purchases (descending)
	sort.Slice(vendorPurchasesSlice, func(i, j int) bool {
		return vendorPurchasesSlice[i].TotalPurchases > vendorPurchasesSlice[j].TotalPurchases
	})
	
	// Vendor data
	for i, vp := range vendorPurchasesSlice {
		row := i + 4 // Start from row 4
		f.SetCellValue(vendorSheet, fmt.Sprintf("A%d", row), i+1) // Rank
		f.SetCellValue(vendorSheet, fmt.Sprintf("B%d", row), vp.Name)
		f.SetCellValue(vendorSheet, fmt.Sprintf("C%d", row), vp.TotalPurchases)
		f.SetCellStyle(vendorSheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), dataStyle)
		f.SetCellStyle(vendorSheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), dataStyle)
		f.SetCellStyle(vendorSheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), currencyStyle)
	}
	
	// Add "By Product" sheet
	productSheet := "By Product"
	f.NewSheet(productSheet)
	
	// Title
	f.SetCellValue(productSheet, "A1", "PURCHASES BY PRODUCT")
	f.MergeCell(productSheet, "A1", "D1")
	f.SetCellStyle(productSheet, "A1", "D1", titleStyle)
	
	// Set column widths
	f.SetColWidth(productSheet, "A", "A", 10)
	f.SetColWidth(productSheet, "B", "B", 40)
	f.SetColWidth(productSheet, "C", "C", 15)
	f.SetColWidth(productSheet, "D", "D", 20)
	
	// Column headers
	f.SetCellValue(productSheet, "A3", "Rank")
	f.SetCellValue(productSheet, "B3", "Product")
	f.SetCellValue(productSheet, "C3", "Quantity")
	f.SetCellValue(productSheet, "D3", "Amount")
	f.SetCellStyle(productSheet, "A3", "D3", tableHeaderStyle)
	
	// Convert map to slice for sorting
	type ProductPurchases struct {
		ID        uint
		Name      string
		Quantity  int
		Amount    float64
	}
	
	var productPurchasesSlice []ProductPurchases
	for productID, info := range purchasesByProduct {
		productPurchasesSlice = append(productPurchasesSlice, ProductPurchases{
			ID:       productID,
			Name:     info.Name,
			Quantity: info.Quantity,
			Amount:   info.Amount,
		})
	}
	
	// Sort by total purchase amount (descending)
	sort.Slice(productPurchasesSlice, func(i, j int) bool {
		return productPurchasesSlice[i].Amount > productPurchasesSlice[j].Amount
	})
	
	// Product data
	for i, pp := range productPurchasesSlice {
		row := i + 4 // Start from row 4
		f.SetCellValue(productSheet, fmt.Sprintf("A%d", row), i+1) // Rank
		f.SetCellValue(productSheet, fmt.Sprintf("B%d", row), pp.Name)
		f.SetCellValue(productSheet, fmt.Sprintf("C%d", row), pp.Quantity)
		f.SetCellValue(productSheet, fmt.Sprintf("D%d", row), pp.Amount)
		f.SetCellStyle(productSheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), dataStyle)
		f.SetCellStyle(productSheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), dataStyle)
		f.SetCellStyle(productSheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), dataStyle)
		f.SetCellStyle(productSheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), currencyStyle)
	}
	
	// Add "Transactions" sheet with detailed information
	transactionSheet := "Transactions"
	f.NewSheet(transactionSheet)
	
	// Title
	f.SetCellValue(transactionSheet, "A1", "PURCHASE TRANSACTIONS")
	f.MergeCell(transactionSheet, "A1", "G1")
	f.SetCellStyle(transactionSheet, "A1", "G1", titleStyle)
	
	// Set column widths
	f.SetColWidth(transactionSheet, "A", "A", 15) // PO Number
	f.SetColWidth(transactionSheet, "B", "B", 15) // Date
	f.SetColWidth(transactionSheet, "C", "C", 30) // Vendor
	f.SetColWidth(transactionSheet, "D", "D", 15) // Amount
	f.SetColWidth(transactionSheet, "E", "E", 15) // Tax
	f.SetColWidth(transactionSheet, "F", "F", 15) // Total
	f.SetColWidth(transactionSheet, "G", "G", 15) // Status
	
	// Column headers
	f.SetCellValue(transactionSheet, "A3", "PO Number")
	f.SetCellValue(transactionSheet, "B3", "Date")
	f.SetCellValue(transactionSheet, "C3", "Vendor")
	f.SetCellValue(transactionSheet, "D3", "Subtotal")
	f.SetCellValue(transactionSheet, "E3", "Tax")
	f.SetCellValue(transactionSheet, "F3", "Total")
	f.SetCellValue(transactionSheet, "G3", "Status")
	f.SetCellStyle(transactionSheet, "A3", "G3", tableHeaderStyle)
	
	// Sort purchases by date (newest first)
	sort.Slice(purchases, func(i, j int) bool {
		return purchases[i].Date.After(purchases[j].Date)
	})
	
	// Purchase transactions
	for i, purchase := range purchases {
		row := i + 4 // Start from row 4
		f.SetCellValue(transactionSheet, fmt.Sprintf("A%d", row), purchase.Code)
		f.SetCellValue(transactionSheet, fmt.Sprintf("B%d", row), purchase.Date.Format("2006-01-02"))
		f.SetCellValue(transactionSheet, fmt.Sprintf("C%d", row), purchase.Vendor.Name)
		f.SetCellValue(transactionSheet, fmt.Sprintf("D%d", row), purchase.NetBeforeTax)
		f.SetCellValue(transactionSheet, fmt.Sprintf("E%d", row), purchase.TaxAmount)
		f.SetCellValue(transactionSheet, fmt.Sprintf("F%d", row), purchase.TotalAmount)
		f.SetCellValue(transactionSheet, fmt.Sprintf("G%d", row), purchase.Status)
		
		// Apply styles
		f.SetCellStyle(transactionSheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), dataStyle)
		f.SetCellStyle(transactionSheet, fmt.Sprintf("D%d", row), fmt.Sprintf("F%d", row), currencyStyle)
		f.SetCellStyle(transactionSheet, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), dataStyle)
	}
	
	// Delete the default sheet
	f.DeleteSheet("Sheet1")
	
	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %v", err)
	}
	
	return buf.Bytes(), nil
}

// GenerateCashFlowStatementPDF generates a cash flow statement in PDF format
func (prs *ProfessionalReportService) GenerateCashFlowStatementPDF(startDate, endDate time.Time) ([]byte, error) {
	theme := DefaultTheme()
	
	// Get cash and bank account balances
	beginningBalance := prs.calculateCashBalance(startDate.AddDate(0, 0, -1))
	endingBalance := prs.calculateCashBalance(endDate)
	
	// Get cash flow data
	operatingActivities := prs.calculateOperatingCashFlows(startDate, endDate)
	investingActivities := prs.calculateInvestingCashFlows(startDate, endDate)
	financingActivities := prs.calculateFinancingCashFlows(startDate, endDate)
	
	// Calculate net cash flows
	var netOperating, netInvesting, netFinancing, netCashFlow float64
	
	for _, item := range operatingActivities {
		if item.Category == "INFLOW" {
			netOperating += item.Amount
		} else {
			netOperating -= item.Amount
		}
	}
	
	for _, item := range investingActivities {
		if item.Category == "INFLOW" {
			netInvesting += item.Amount
		} else {
			netInvesting -= item.Amount
		}
	}
	
	for _, item := range financingActivities {
		if item.Category == "INFLOW" {
			netFinancing += item.Amount
		} else {
			netFinancing -= item.Amount
		}
	}
	
	netCashFlow = netOperating + netInvesting + netFinancing
	
	// Create PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Set up fonts
	pdf.SetFont(theme.HeaderFont, "B", 16)
	
	// Company header
	pdf.Cell(190, 10, prs.companyProfile.Name)
	pdf.Ln(7)
	pdf.SetFont(theme.BodyFont, "", 10)
	pdf.Cell(190, 5, prs.companyProfile.Address)
	pdf.Ln(5)
	pdf.Cell(190, 5, fmt.Sprintf("%s, %s %s", prs.companyProfile.City, prs.companyProfile.State, prs.companyProfile.PostalCode))
	pdf.Ln(5)
	pdf.Cell(190, 5, fmt.Sprintf("Tel: %s | Email: %s", prs.companyProfile.Phone, prs.companyProfile.Email))
	pdf.Ln(10)
	
	// Report title
	pdf.SetFont(theme.HeaderFont, "B", 14)
	pdf.Cell(190, 10, "CASH FLOW STATEMENT")
	pdf.Ln(7)
	pdf.SetFont(theme.BodyFont, "", 10)
	pdf.Cell(190, 6, fmt.Sprintf("For the period %s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	pdf.Ln(7)
	pdf.Cell(190, 6, fmt.Sprintf("Currency: %s", prs.companyProfile.Currency))
	pdf.Ln(10)
	
	// Beginning balance
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(120, 6, "Cash and Cash Equivalents, Beginning")
	pdf.CellFormat(70, 6, prs.FormatCurrency(beginningBalance), "0", 1, "R", false, 0, "")
	pdf.Ln(5)
	
	// Operating Activities
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "CASH FLOWS FROM OPERATING ACTIVITIES", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)
	
	// Operating activities items
	pdf.SetFont(theme.BodyFont, "", 9)
	for _, item := range operatingActivities {
		pdf.Cell(120, 6, "  "+item.Description)
		if item.Category == "INFLOW" {
			pdf.CellFormat(70, 6, prs.FormatCurrency(item.Amount), "0", 1, "R", false, 0, "")
		} else {
			pdf.CellFormat(70, 6, "(" + prs.FormatCurrency(item.Amount) + ")", "0", 1, "R", false, 0, "")
		}
	}
	
	// Net cash from operating activities
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(120, 6, "Net Cash from Operating Activities")
	pdf.CellFormat(70, 6, prs.FormatCurrency(netOperating), "T", 1, "R", false, 0, "")
	pdf.Ln(10)
	
	// Investing Activities
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "CASH FLOWS FROM INVESTING ACTIVITIES", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)
	
	// Investing activities items
	pdf.SetFont(theme.BodyFont, "", 9)
	for _, item := range investingActivities {
		pdf.Cell(120, 6, "  "+item.Description)
		if item.Category == "INFLOW" {
			pdf.CellFormat(70, 6, prs.FormatCurrency(item.Amount), "0", 1, "R", false, 0, "")
		} else {
			pdf.CellFormat(70, 6, "(" + prs.FormatCurrency(item.Amount) + ")", "0", 1, "R", false, 0, "")
		}
	}
	
	// Net cash from investing activities
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(120, 6, "Net Cash from Investing Activities")
	pdf.CellFormat(70, 6, prs.FormatCurrency(netInvesting), "T", 1, "R", false, 0, "")
	pdf.Ln(10)
	
	// Financing Activities
	pdf.SetFillColor(theme.HeaderColor[0], theme.HeaderColor[1], theme.HeaderColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont(theme.HeaderFont, "B", 11)
	pdf.CellFormat(190, 8, "CASH FLOWS FROM FINANCING ACTIVITIES", "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)
	
	// Financing activities items
	pdf.SetFont(theme.BodyFont, "", 9)
	for _, item := range financingActivities {
		pdf.Cell(120, 6, "  "+item.Description)
		if item.Category == "INFLOW" {
			pdf.CellFormat(70, 6, prs.FormatCurrency(item.Amount), "0", 1, "R", false, 0, "")
		} else {
			pdf.CellFormat(70, 6, "(" + prs.FormatCurrency(item.Amount) + ")", "0", 1, "R", false, 0, "")
		}
	}
	
	// Net cash from financing activities
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(120, 6, "Net Cash from Financing Activities")
	pdf.CellFormat(70, 6, prs.FormatCurrency(netFinancing), "T", 1, "R", false, 0, "")
	pdf.Ln(10)
	
	// Net increase/decrease in cash
	pdf.SetFont(theme.BodyFont, "B", 10)
	pdf.Cell(120, 6, "Net Increase/(Decrease) in Cash")
	pdf.CellFormat(70, 6, prs.FormatCurrency(netCashFlow), "T", 1, "R", false, 0, "")
	pdf.Ln(5)
	
	// Ending balance
	pdf.SetFont(theme.BodyFont, "B", 11)
	pdf.Cell(120, 6, "Cash and Cash Equivalents, Ending")
	pdf.SetFillColor(theme.SecondaryColor[0], theme.SecondaryColor[1], theme.SecondaryColor[2])
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(70, 6, prs.FormatCurrency(endingBalance), "1", 1, "R", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	
	// Add footer
	_, pageHeight := pdf.GetPageSize()
	pdf.SetY(pageHeight - 20)
	pdf.SetFont(theme.BodyFont, "I", 8)
	pdf.CellFormat(190, 10, fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006 15:04:05")), "T", 1, "C", false, 0, "")
	
	if theme.ShowPageNumbers {
		_, pageHeight := pdf.GetPageSize()
		pdf.SetY(pageHeight - 10)
		pdf.CellFormat(190, 10, fmt.Sprintf("Page %d of %d", pdf.PageNo(), pdf.PageCount()), "", 0, "C", false, 0, "")
	}
	
	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}
	
	return buf.Bytes(), nil
}

// Helper methods for cash flow

// calculateCashBalance calculates total cash and bank balance at a specific date
func (prs *ProfessionalReportService) calculateCashBalance(asOfDate time.Time) float64 {
	ctx := context.Background()
	
	// Get all cash and bank accounts
	accounts, err := prs.accountRepo.FindAll(ctx)
	if err != nil {
		return 0
	}
	
	var totalCash float64
	for _, account := range accounts {
		// Check if account is cash or bank account (by code prefix or type)
		if prs.isCashOrBankAccount(account) {
			balance := prs.calculateAccountBalance(account.ID, asOfDate)
			totalCash += balance
		}
	}
	
	return totalCash
}

// isCashOrBankAccount checks if an account is a cash or bank account
func (prs *ProfessionalReportService) isCashOrBankAccount(account models.Account) bool {
	// Common cash/bank account codes start with 1-1 (assets) and contain cash/bank keywords
	cashKeywords := []string{"cash", "bank", "checking", "savings", "petty"}
	accountNameLower := strings.ToLower(account.Name)
	accountCodeLower := strings.ToLower(account.Code)
	
	// Check if account type is asset and contains cash/bank keywords
	if account.Type != models.AccountTypeAsset {
		return false
	}
	
	for _, keyword := range cashKeywords {
		if strings.Contains(accountNameLower, keyword) || strings.Contains(accountCodeLower, keyword) {
			return true
		}
	}
	
	// Also check if account code starts with typical cash/bank prefixes
	if strings.HasPrefix(account.Code, "1-1-001") || strings.HasPrefix(account.Code, "1-1-002") {
		return true
	}
	
	return false
}

// calculateOperatingCashFlows calculates cash flows from operating activities
func (prs *ProfessionalReportService) calculateOperatingCashFlows(startDate, endDate time.Time) []CashFlowItem {
	var items []CashFlowItem
	
	// For a real implementation, you would calculate these from actual data
	// Here we provide some example data for demonstration
	
	// Get net income (simplified calculation)
	var totalRevenue float64
	var totalExpenses float64
	
	ctx := context.Background()
	accounts, _ := prs.accountRepo.FindAll(ctx)
	
	for _, account := range accounts {
		balance := prs.calculateAccountBalanceForPeriod(account.ID, startDate, endDate)
		
		switch account.Type {
		case models.AccountTypeRevenue:
			totalRevenue += balance
		case models.AccountTypeExpense:
			totalExpenses += balance
		}
	}
	
	netIncome := totalRevenue - totalExpenses
	
	// Add net income as the first item
	items = append(items, CashFlowItem{
		Description: "Net Income",
		Amount:      math.Abs(netIncome),
		Category:        prs.getFlowType(netIncome),
	})
	
	// Add adjustments
	// In a real implementation, these would be calculated from actual transactions
	items = append(items, CashFlowItem{
		Description: "Depreciation and Amortization",
		Amount:      10000,
		Category:        "INFLOW", // Non-cash expense, so it's added back
	})
	
	items = append(items, CashFlowItem{
		Description: "Increase in Accounts Receivable",
		Amount:      5000,
		Category:        "OUTFLOW", // Increase in receivables uses cash
	})
	
	items = append(items, CashFlowItem{
		Description: "Decrease in Inventory",
		Amount:      3000,
		Category:        "INFLOW", // Decrease in inventory generates cash
	})
	
	items = append(items, CashFlowItem{
		Description: "Increase in Accounts Payable",
		Amount:      7000,
		Category:        "INFLOW", // Increase in payables preserves cash
	})
	
	return items
}

// calculateInvestingCashFlows calculates cash flows from investing activities
func (prs *ProfessionalReportService) calculateInvestingCashFlows(startDate, endDate time.Time) []CashFlowItem {
	var items []CashFlowItem
	
	// For a real implementation, you would calculate these from actual data
	// Here we provide some example data for demonstration
	
	// Purchase of equipment
	items = append(items, CashFlowItem{
		Description: "Purchase of Equipment",
		Amount:      25000,
		Category:        "OUTFLOW",
	})
	
	// Purchase of investments
	items = append(items, CashFlowItem{
		Description: "Purchase of Investments",
		Amount:      15000,
		Category:        "OUTFLOW",
	})
	
	// Sale of investments
	items = append(items, CashFlowItem{
		Description: "Sale of Investments",
		Amount:      8000,
		Category:        "INFLOW",
	})
	
	return items
}

// calculateFinancingCashFlows calculates cash flows from financing activities
func (prs *ProfessionalReportService) calculateFinancingCashFlows(startDate, endDate time.Time) []CashFlowItem {
	var items []CashFlowItem
	
	// For a real implementation, you would calculate these from actual data
	// Here we provide some example data for demonstration
	
	// Proceeds from loans
	items = append(items, CashFlowItem{
		Description: "Proceeds from Long-term Debt",
		Amount:      50000,
		Category:        "INFLOW",
	})
	
	// Repayment of loans
	items = append(items, CashFlowItem{
		Description: "Repayment of Short-term Debt",
		Amount:      10000,
		Category:        "OUTFLOW",
	})
	
	// Payment of dividends
	items = append(items, CashFlowItem{
		Description: "Payment of Dividends",
		Amount:      5000,
		Category:        "OUTFLOW",
	})
	
	return items
}

// Helper method to determine flow type based on amount
func (prs *ProfessionalReportService) getFlowType(amount float64) string {
	if amount >= 0 {
		return "INFLOW"
	}
	return "OUTFLOW"
}

// formatCSVAmount formats amount for CSV output
func (prs *ProfessionalReportService) formatCSVAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

// GenerateProfitLossCSV generates a professional profit and loss statement CSV file
func (prs *ProfessionalReportService) GenerateProfitLossCSV(startDate, endDate time.Time) ([]byte, error) {
	// Get all accounts with their balances
	ctx := context.Background()
	accounts, err := prs.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}
	
	// Prepare data structure for profit & loss
	var revenues []models.Account
	var expenses []models.Account
	
	var totalRevenue float64
	var totalExpenses float64
	
	// Categorize accounts and calculate totals
	for _, account := range accounts {
		balance := prs.calculateAccountBalanceForPeriod(account.ID, startDate, endDate)
		
		// Skip accounts with no activity
		if balance == 0 {
			continue
		}
		
		account.Balance = balance
		
		switch account.Type {
		case models.AccountTypeRevenue:
			revenues = append(revenues, account)
			totalRevenue += balance
		case models.AccountTypeExpense:
			expenses = append(expenses, account)
			totalExpenses += balance
		}
	}
	
	// Sort accounts by code
	sort.Slice(revenues, func(i, j int) bool { return revenues[i].Code < revenues[j].Code })
	sort.Slice(expenses, func(i, j int) bool { return expenses[i].Code < expenses[j].Code })
	
	// Calculate gross profit and net income
	grossProfit := totalRevenue - totalExpenses
	netIncome := grossProfit // Simplified
	
	// Create CSV content
	var buf bytes.Buffer
	
	// Write company header
	buf.WriteString(fmt.Sprintf("%s\n", prs.companyProfile.Name))
	buf.WriteString(fmt.Sprintf("%s\n", prs.companyProfile.Address))
	buf.WriteString(fmt.Sprintf("%s, %s %s\n", prs.companyProfile.City, prs.companyProfile.State, prs.companyProfile.PostalCode))
	buf.WriteString(fmt.Sprintf("Tel: %s | Email: %s\n", prs.companyProfile.Phone, prs.companyProfile.Email))
	buf.WriteString("\n")
	
	// Write report title
	buf.WriteString("PROFIT & LOSS STATEMENT\n")
	buf.WriteString(fmt.Sprintf("For the period %s to %s\n", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	buf.WriteString(fmt.Sprintf("Currency: %s\n", prs.companyProfile.Currency))
	buf.WriteString("\n")
	
	// Write CSV headers
	buf.WriteString("Section,Account Code,Account Name,Amount\n")
	
	// REVENUE SECTION
	buf.WriteString("REVENUE,,,\n")
	
	// Operating Revenue
	buf.WriteString("Operating Revenue,,,\n")
	var operatingRevenue float64
	for _, revenue := range revenues {
		if revenue.Category == models.CategoryOperatingRevenue && !revenue.IsHeader {
			buf.WriteString(fmt.Sprintf(",\"%s\",\"%s\",%s\n", revenue.Code, revenue.Name, prs.formatCSVAmount(revenue.Balance)))
			operatingRevenue += revenue.Balance
		}
	}
	buf.WriteString(fmt.Sprintf(",,,Total Operating Revenue,%s\n", prs.formatCSVAmount(operatingRevenue)))
	buf.WriteString("\n")
	
	// Other Revenue
	buf.WriteString("Other Revenue,,,\n")
	var otherRevenue float64
	for _, revenue := range revenues {
	if revenue.Category == models.CategoryOtherIncome && !revenue.IsHeader {
			buf.WriteString(fmt.Sprintf(",\"%s\",\"%s\",%s\n", revenue.Code, revenue.Name, prs.formatCSVAmount(revenue.Balance)))
			otherRevenue += revenue.Balance
		}
	}
	buf.WriteString(fmt.Sprintf(",,,Total Other Revenue,%s\n", prs.formatCSVAmount(otherRevenue)))
	buf.WriteString("\n")
	
	// Total Revenue
	buf.WriteString(fmt.Sprintf(",,,TOTAL REVENUE,%s\n", prs.formatCSVAmount(totalRevenue)))
	buf.WriteString("\n")
	
	// EXPENSES SECTION
	buf.WriteString("EXPENSES,,,\n")
	
	// Operating Expenses
	buf.WriteString("Operating Expenses,,,\n")
	var operatingExpenses float64
	for _, expense := range expenses {
		if expense.Category == models.CategoryOperatingExpense && !expense.IsHeader {
			buf.WriteString(fmt.Sprintf(",\"%s\",\"%s\",%s\n", expense.Code, expense.Name, prs.formatCSVAmount(expense.Balance)))
			operatingExpenses += expense.Balance
		}
	}
	buf.WriteString(fmt.Sprintf(",,,Total Operating Expenses,%s\n", prs.formatCSVAmount(operatingExpenses)))
	buf.WriteString("\n")
	
	// Other Expenses
	buf.WriteString("Other Expenses,,,\n")
	var otherExpenses float64
	for _, expense := range expenses {
		if expense.Category == models.CategoryOtherExpense && !expense.IsHeader {
			buf.WriteString(fmt.Sprintf(",\"%s\",\"%s\",%s\n", expense.Code, expense.Name, prs.formatCSVAmount(expense.Balance)))
			otherExpenses += expense.Balance
		}
	}
	buf.WriteString(fmt.Sprintf(",,,Total Other Expenses,%s\n", prs.formatCSVAmount(otherExpenses)))
	buf.WriteString("\n")
	
	// Total Expenses
	buf.WriteString(fmt.Sprintf(",,,TOTAL EXPENSES,%s\n", prs.formatCSVAmount(totalExpenses)))
	buf.WriteString("\n")
	
	// SUMMARY SECTION
	buf.WriteString("SUMMARY,,,\n")
	buf.WriteString(fmt.Sprintf(",,,Gross Profit,%s\n", prs.formatCSVAmount(grossProfit)))
	buf.WriteString(fmt.Sprintf(",,,NET INCOME,%s\n", prs.formatCSVAmount(netIncome)))
	buf.WriteString("\n")
	
	// Footer
	buf.WriteString(fmt.Sprintf("Generated on %s\n", time.Now().Format("January 2, 2006 15:04:05")))
	
	return buf.Bytes(), nil
}

// GenerateSalesSummaryCSV creates a sales summary report in CSV format
func (prs *ProfessionalReportService) GenerateSalesSummaryCSV(startDate, endDate time.Time, groupBy string) ([]byte, error) {
	// Query sales data
	var sales []models.Sale
	if err := prs.db.Preload("Customer").
		Preload("SaleItems").
		Preload("SaleItems.Product").
		Where("sales.date BETWEEN ? AND ?", startDate, endDate).
		Find(&sales).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch sales data: %v", err)
	}
	
	// Calculate totals
	var totalSales float64
	salesByCustomer := make(map[uint]float64)
	salesByProduct := make(map[uint]struct {
		Name    string
		Quantity int64
		Amount  float64
	})
	salesByPeriod := make(map[string]float64)
	
	for _, sale := range sales {
		totalSales += sale.TotalAmount
		
		// Sales by customer
		salesByCustomer[sale.CustomerID] += sale.TotalAmount
		
		// Sales by period
		var period string
		switch groupBy {
		case "month":
			period = sale.Date.Format("2006-01")
		case "quarter":
			quarter := (sale.Date.Month()-1)/3 + 1
			period = fmt.Sprintf("%d-Q%d", sale.Date.Year(), quarter)
		case "year":
			period = sale.Date.Format("2006")
		default:
			period = sale.Date.Format("2006-01-02")
		}
		salesByPeriod[period] += sale.TotalAmount
		
		// Sales by product
		for _, item := range sale.SaleItems {
			productInfo := salesByProduct[item.ProductID]
			productInfo.Name = item.Product.Name
			productInfo.Quantity += int64(item.Quantity)
			productInfo.Amount += item.LineTotal
			salesByProduct[item.ProductID] = productInfo
		}
	}
	
	// Create CSV content
	var buf bytes.Buffer
	
	// Write company header
	buf.WriteString(fmt.Sprintf("%s\n", prs.companyProfile.Name))
	buf.WriteString(fmt.Sprintf("%s\n", prs.companyProfile.Address))
	buf.WriteString(fmt.Sprintf("%s, %s %s\n", prs.companyProfile.City, prs.companyProfile.State, prs.companyProfile.PostalCode))
	buf.WriteString(fmt.Sprintf("Tel: %s | Email: %s\n", prs.companyProfile.Phone, prs.companyProfile.Email))
	buf.WriteString("\n")
	
	// Write report title
	buf.WriteString("SALES SUMMARY REPORT\n")
	buf.WriteString(fmt.Sprintf("For the period %s to %s\n", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	buf.WriteString(fmt.Sprintf("Currency: %s\n", prs.companyProfile.Currency))
	buf.WriteString("\n")
	
	// Sales Summary Section
	buf.WriteString("SALES SUMMARY\n")
	buf.WriteString(fmt.Sprintf("Total Sales,%s\n", prs.formatCSVAmount(totalSales)))
	buf.WriteString(fmt.Sprintf("Total Transactions,%d\n", len(sales)))
	
	avgOrderValue := 0.0
	if len(sales) > 0 {
		avgOrderValue = totalSales / float64(len(sales))
	}
	buf.WriteString(fmt.Sprintf("Average Order Value,%s\n", prs.formatCSVAmount(avgOrderValue)))
	buf.WriteString("\n")
	
	// Sales by Period Section
	buf.WriteString("SALES BY PERIOD\n")
	buf.WriteString("Period,Amount\n")
	
	// Convert map to slice for sorting
	type PeriodSales struct {
		Period string
		Amount float64
	}
	
	var periodSalesSlice []PeriodSales
	for period, amount := range salesByPeriod {
		periodSalesSlice = append(periodSalesSlice, PeriodSales{
			Period: period,
			Amount: amount,
		})
	}
	
	// Sort by period
	sort.Slice(periodSalesSlice, func(i, j int) bool {
		return periodSalesSlice[i].Period < periodSalesSlice[j].Period
	})
	
	// Sales data by period
	for _, ps := range periodSalesSlice {
		buf.WriteString(fmt.Sprintf("\"%s\",%s\n", ps.Period, prs.formatCSVAmount(ps.Amount)))
	}
	buf.WriteString("\n")
	
	// Top Customers Section
	buf.WriteString("TOP CUSTOMERS\n")
	buf.WriteString("Rank,Customer,Total Sales\n")
	
	// Convert map to slice for sorting
	type CustomerSales struct {
		ID        uint
		Name      string
		TotalSales float64
	}
	
	var customerSalesSlice []CustomerSales
	for customerID, amount := range salesByCustomer {
		var customerName string
		for _, sale := range sales {
			if sale.CustomerID == customerID {
				customerName = sale.Customer.Name
				break
			}
		}
		customerSalesSlice = append(customerSalesSlice, CustomerSales{
			ID:        customerID,
			Name:      customerName,
			TotalSales: amount,
		})
	}
	
	// Sort by total sales (descending)
	sort.Slice(customerSalesSlice, func(i, j int) bool {
		return customerSalesSlice[i].TotalSales > customerSalesSlice[j].TotalSales
	})
	
	// Customer data
	for i, cs := range customerSalesSlice {
		buf.WriteString(fmt.Sprintf("%d,\"%s\",%s\n", i+1, cs.Name, prs.formatCSVAmount(cs.TotalSales)))
	}
	buf.WriteString("\n")
	
	// Top Products Section
	buf.WriteString("TOP PRODUCTS\n")
	buf.WriteString("Rank,Product,Quantity,Amount\n")
	
	// Convert map to slice for sorting
	type ProductSales struct {
		ID        uint
		Name      string
		Quantity  int64
		Amount    float64
	}
	
	var productSalesSlice []ProductSales
	for productID, info := range salesByProduct {
		productSalesSlice = append(productSalesSlice, ProductSales{
			ID:       productID,
			Name:     info.Name,
			Quantity: info.Quantity,
			Amount:   info.Amount,
		})
	}
	
	// Sort by total sales amount (descending)
	sort.Slice(productSalesSlice, func(i, j int) bool {
		return productSalesSlice[i].Amount > productSalesSlice[j].Amount
	})
	
	// Product data
	for i, ps := range productSalesSlice {
		buf.WriteString(fmt.Sprintf("%d,\"%s\",%d,%s\n", i+1, ps.Name, ps.Quantity, prs.formatCSVAmount(ps.Amount)))
	}
	buf.WriteString("\n")
	
	// Footer
	buf.WriteString(fmt.Sprintf("Generated on %s\n", time.Now().Format("January 2, 2006 15:04:05")))
	
	return buf.Bytes(), nil
}

// GeneratePurchaseSummaryCSV creates a purchase summary report in CSV format
func (prs *ProfessionalReportService) GeneratePurchaseSummaryCSV(startDate, endDate time.Time, groupBy string) ([]byte, error) {
	// Query purchase data
	var purchases []models.Purchase
	if err := prs.db.Preload("Vendor").
		Preload("PurchaseItems").
		Preload("PurchaseItems.Product").
		Where("purchases.date BETWEEN ? AND ?", startDate, endDate).
		Find(&purchases).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch purchase data: %v", err)
	}
	
	// Calculate totals
	var totalPurchases float64
	purchasesByVendor := make(map[uint]float64)
	purchasesByProduct := make(map[uint]struct {
		Name    string
		Quantity int
		Amount  float64
	})
	purchasesByPeriod := make(map[string]float64)
	
	for _, purchase := range purchases {
		totalPurchases += purchase.TotalAmount
		
		// Purchases by vendor
		purchasesByVendor[purchase.VendorID] += purchase.TotalAmount
		
		// Purchases by period
		var period string
		switch groupBy {
		case "month":
			period = purchase.Date.Format("2006-01")
		case "quarter":
			quarter := (purchase.Date.Month()-1)/3 + 1
			period = fmt.Sprintf("%d-Q%d", purchase.Date.Year(), quarter)
		case "year":
			period = purchase.Date.Format("2006")
		default:
			period = purchase.Date.Format("2006-01-02")
		}
		purchasesByPeriod[period] += purchase.TotalAmount
		
		// Purchases by product
		for _, item := range purchase.PurchaseItems {
			productInfo := purchasesByProduct[item.ProductID]
			productInfo.Name = item.Product.Name
			productInfo.Quantity += item.Quantity
			productInfo.Amount += item.TotalPrice
			purchasesByProduct[item.ProductID] = productInfo
		}
	}
	
	// Create CSV content
	var buf bytes.Buffer
	
	// Write company header
	buf.WriteString(fmt.Sprintf("%s\n", prs.companyProfile.Name))
	buf.WriteString(fmt.Sprintf("%s\n", prs.companyProfile.Address))
	buf.WriteString(fmt.Sprintf("%s, %s %s\n", prs.companyProfile.City, prs.companyProfile.State, prs.companyProfile.PostalCode))
	buf.WriteString(fmt.Sprintf("Tel: %s | Email: %s\n", prs.companyProfile.Phone, prs.companyProfile.Email))
	buf.WriteString("\n")
	
	// Write report title
	buf.WriteString("PURCHASE SUMMARY REPORT\n")
	buf.WriteString(fmt.Sprintf("For the period %s to %s\n", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	buf.WriteString(fmt.Sprintf("Currency: %s\n", prs.companyProfile.Currency))
	buf.WriteString("\n")
	
	// Purchase Summary Section
	buf.WriteString("PURCHASE SUMMARY\n")
	buf.WriteString(fmt.Sprintf("Total Purchases,%s\n", prs.formatCSVAmount(totalPurchases)))
	buf.WriteString(fmt.Sprintf("Total Transactions,%d\n", len(purchases)))
	
	avgOrderValue := 0.0
	if len(purchases) > 0 {
		avgOrderValue = totalPurchases / float64(len(purchases))
	}
	buf.WriteString(fmt.Sprintf("Average Order Value,%s\n", prs.formatCSVAmount(avgOrderValue)))
	buf.WriteString("\n")
	
	// Purchases by Period Section
	buf.WriteString("PURCHASES BY PERIOD\n")
	buf.WriteString("Period,Amount\n")
	
	// Convert map to slice for sorting
	type PeriodPurchases struct {
		Period string
		Amount float64
	}
	
	var periodPurchasesSlice []PeriodPurchases
	for period, amount := range purchasesByPeriod {
		periodPurchasesSlice = append(periodPurchasesSlice, PeriodPurchases{
			Period: period,
			Amount: amount,
		})
	}
	
	// Sort by period
	sort.Slice(periodPurchasesSlice, func(i, j int) bool {
		return periodPurchasesSlice[i].Period < periodPurchasesSlice[j].Period
	})
	
	// Purchases data by period
	for _, pp := range periodPurchasesSlice {
		buf.WriteString(fmt.Sprintf("\"%s\",%s\n", pp.Period, prs.formatCSVAmount(pp.Amount)))
	}
	buf.WriteString("\n")
	
	// Top Vendors Section
	buf.WriteString("TOP VENDORS\n")
	buf.WriteString("Rank,Vendor,Total Purchases\n")
	
	// Convert map to slice for sorting
	type VendorPurchases struct {
		ID            uint
		Name          string
		TotalPurchases float64
	}
	
	var vendorPurchasesSlice []VendorPurchases
	for vendorID, amount := range purchasesByVendor {
		var vendorName string
		for _, purchase := range purchases {
			if purchase.VendorID == vendorID {
				vendorName = purchase.Vendor.Name
				break
			}
		}
		vendorPurchasesSlice = append(vendorPurchasesSlice, VendorPurchases{
			ID:            vendorID,
			Name:          vendorName,
			TotalPurchases: amount,
		})
	}
	
	// Sort by total purchases (descending)
	sort.Slice(vendorPurchasesSlice, func(i, j int) bool {
		return vendorPurchasesSlice[i].TotalPurchases > vendorPurchasesSlice[j].TotalPurchases
	})
	
	// Vendor data
	for i, vp := range vendorPurchasesSlice {
		buf.WriteString(fmt.Sprintf("%d,\"%s\",%s\n", i+1, vp.Name, prs.formatCSVAmount(vp.TotalPurchases)))
	}
	buf.WriteString("\n")
	
	// Top Products Section
	buf.WriteString("TOP PRODUCTS\n")
	buf.WriteString("Rank,Product,Quantity,Amount\n")
	
	// Convert map to slice for sorting
	type ProductPurchases struct {
		ID        uint
		Name      string
		Quantity  int
		Amount    float64
	}
	
	var productPurchasesSlice []ProductPurchases
	for productID, info := range purchasesByProduct {
		productPurchasesSlice = append(productPurchasesSlice, ProductPurchases{
			ID:       productID,
			Name:     info.Name,
			Quantity: info.Quantity,
			Amount:   info.Amount,
		})
	}
	
	// Sort by total purchase amount (descending)
	sort.Slice(productPurchasesSlice, func(i, j int) bool {
		return productPurchasesSlice[i].Amount > productPurchasesSlice[j].Amount
	})
	
	// Product data
	for i, pp := range productPurchasesSlice {
		buf.WriteString(fmt.Sprintf("%d,\"%s\",%d,%s\n", i+1, pp.Name, pp.Quantity, prs.formatCSVAmount(pp.Amount)))
	}
	buf.WriteString("\n")
	
	// Footer
	buf.WriteString(fmt.Sprintf("Generated on %s\n", time.Now().Format("January 2, 2006 15:04:05")))
	
	return buf.Bytes(), nil
}

