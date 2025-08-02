package database

import (
	"log"
	"time"
	"app-sistem-akuntansi/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedData(db *gorm.DB) {
	log.Println("Starting database seeding...")

	// Seed Users
	seedUsers(db)
	
	// Seed Chart of Accounts
	seedAccounts(db)
	
	// Seed Contacts
	seedContacts(db)
	
	// Seed Product Categories
	seedProductCategories(db)
	
	// Seed Products
	seedProducts(db)
	
	// Seed Expense Categories
	seedExpenseCategories(db)
	
	// Seed Cash & Bank accounts
	seedCashBankAccounts(db)
	
	// Seed Company Profile
	seedCompanyProfile(db)
	
	// Seed Report Templates
	seedReportTemplates(db)

	log.Println("Database seeding completed successfully")
}

func seedUsers(db *gorm.DB) {
	// Check if users already exist
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	users := []models.User{
		{
			Username:  "admin",
			Email:     "admin@company.com",
			Password:  string(hashedPassword),
			Role:      "admin",
			FirstName: "System",
			LastName:  "Administrator",
			IsActive:  true,
		},
		{
			Username:  "finance",
			Email:     "finance@company.com",
			Password:  string(hashedPassword),
			Role:      "finance",
			FirstName: "Finance",
			LastName:  "Manager",
			IsActive:  true,
		},
		{
			Username:  "inventory",
			Email:     "inventory@company.com",
			Password:  string(hashedPassword),
			Role:      "inventory_manager",
			FirstName: "Inventory",
			LastName:  "Manager",
			IsActive:  true,
		},
		{
			Username:  "director",
			Email:     "director@company.com",
			Password:  string(hashedPassword),
			Role:      "director",
			FirstName: "Company",
			LastName:  "Director",
			IsActive:  true,
		},
	}

	for _, user := range users {
		db.Create(&user)
	}
}

func seedAccounts(db *gorm.DB) {
	// Check if accounts already exist
	var count int64
	db.Model(&models.Account{}).Count(&count)
	if count > 0 {
		return
	}

	accounts := []models.Account{
		// ASSETS
		{Code: "1000", Name: "ASET", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, IsHeader: true, Level: 1},
		{Code: "1100", Name: "ASET LANCAR", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, IsHeader: true, Level: 2},
		{Code: "1101", Name: "Kas", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3},
		{Code: "1102", Name: "Bank BCA", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3},
		{Code: "1103", Name: "Bank Mandiri", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3},
		{Code: "1201", Name: "Piutang Usaha", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3},
		{Code: "1301", Name: "Persediaan Barang Dagangan", Type: models.AccountTypeAsset, Category: models.CategoryCurrentAsset, Level: 3},
		
		// FIXED ASSETS
		{Code: "1500", Name: "ASET TETAP", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, IsHeader: true, Level: 2},
		{Code: "1501", Name: "Peralatan Kantor", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3},
		{Code: "1502", Name: "Kendaraan", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3},
		{Code: "1503", Name: "Bangunan", Type: models.AccountTypeAsset, Category: models.CategoryFixedAsset, Level: 3},

		// LIABILITIES
		{Code: "2000", Name: "KEWAJIBAN", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, IsHeader: true, Level: 1},
		{Code: "2100", Name: "KEWAJIBAN LANCAR", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, IsHeader: true, Level: 2},
		{Code: "2101", Name: "Utang Usaha", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3},
		{Code: "2102", Name: "Utang Pajak", Type: models.AccountTypeLiability, Category: models.CategoryCurrentLiability, Level: 3},

		// EQUITY
		{Code: "3000", Name: "EKUITAS", Type: models.AccountTypeEquity, Category: models.CategoryEquity, IsHeader: true, Level: 1},
		{Code: "3101", Name: "Modal Pemilik", Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 2},
		{Code: "3201", Name: "Laba Ditahan", Type: models.AccountTypeEquity, Category: models.CategoryEquity, Level: 2},

		// REVENUE
		{Code: "4000", Name: "PENDAPATAN", Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, IsHeader: true, Level: 1},
		{Code: "4101", Name: "Pendapatan Penjualan", Type: models.AccountTypeRevenue, Category: models.CategoryOperatingRevenue, Level: 2},
		{Code: "4201", Name: "Pendapatan Lain-lain", Type: models.AccountTypeRevenue, Category: models.CategoryOtherRevenue, Level: 2},

		// EXPENSES
		{Code: "5000", Name: "BEBAN", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, IsHeader: true, Level: 1},
		{Code: "5101", Name: "Harga Pokok Penjualan", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2},
		{Code: "5201", Name: "Beban Gaji", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2},
		{Code: "5202", Name: "Beban Listrik", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2},
		{Code: "5203", Name: "Beban Telepon", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2},
		{Code: "5204", Name: "Beban Transportasi", Type: models.AccountTypeExpense, Category: models.CategoryOperatingExpense, Level: 2},
	}

	for _, account := range accounts {
		db.Create(&account)
	}
}

func seedContacts(db *gorm.DB) {
	// Check if contacts already exist
	var count int64
	db.Model(&models.Contact{}).Count(&count)
	if count > 0 {
		return
	}

	contacts := []models.Contact{
		// Customers
		{
			Code:         "CUST001",
			Name:         "PT Maju Sejahtera",
			Type:         models.ContactTypeCustomer,
			Category:     models.CategoryWholesale,
			Email:        "contact@majusejahtera.com",
			Phone:        "021-1234567",
			CreditLimit:  50000000,
			PaymentTerms: 30,
		},
		{
			Code:         "CUST002",
			Name:         "CV Berkah Jaya",
			Type:         models.ContactTypeCustomer,
			Category:     models.CategoryRetail,
			Email:        "info@berkahjaya.com",
			Phone:        "021-2345678",
			CreditLimit:  25000000,
			PaymentTerms: 15,
		},
		
		// Vendors
		{
			Code:         "VEND001",
			Name:         "PT Supplier Utama",
			Type:         models.ContactTypeVendor,
			Category:     models.CategoryDistributor,
			Email:        "sales@supplierutama.com",
			Phone:        "021-3456789",
			PaymentTerms: 30,
		},
		{
			Code:         "VEND002",
			Name:         "UD Barang Lengkap",
			Type:         models.ContactTypeVendor,
			Category:     models.CategoryWholesale,
			Email:        "order@baranglengkap.com",
			Phone:        "021-4567890",
			PaymentTerms: 21,
		},
	}

	for _, contact := range contacts {
		db.Create(&contact)
	}
}

func seedProductCategories(db *gorm.DB) {
	// Check if product categories already exist
	var count int64
	db.Model(&models.ProductCategory{}).Count(&count)
	if count > 0 {
		return
	}

	categories := []models.ProductCategory{
		{Code: "CAT001", Name: "Elektronik", Description: "Produk elektronik dan gadget"},
		{Code: "CAT002", Name: "Furniture", Description: "Perabotan dan furniture kantor"},
		{Code: "CAT003", Name: "Alat Tulis", Description: "Alat tulis dan perlengkapan kantor"},
		{Code: "CAT004", Name: "Komputer", Description: "Komputer dan aksesoris"},
	}

	for _, category := range categories {
		db.Create(&category)
	}
}

func seedProducts(db *gorm.DB) {
	// Check if products already exist
	var count int64
	db.Model(&models.Product{}).Count(&count)
	if count > 0 {
		return
	}

	// Get first category for relation
	var category models.ProductCategory
	db.First(&category)

	products := []models.Product{
		{
			Code:          "PRD001",
			Name:          "Laptop Dell XPS 13",
			Description:   "Laptop Dell XPS 13 inch with Intel Core i7",
			CategoryID:    &category.ID,
			Brand:         "Dell",
			Unit:          "pcs",
			PurchasePrice: 12000000,
			SalePrice:     15000000,
			Stock:         10,
			MinStock:      5,
			MaxStock:      50,
			ReorderLevel:  8,
			SKU:           "DELL-XPS13-I7",
			IsActive:      true,
		},
		{
			Code:          "PRD002",
			Name:          "Mouse Wireless Logitech",
			Description:   "Mouse wireless Logitech MX Master 3",
			CategoryID:    &category.ID,
			Brand:         "Logitech",
			Unit:          "pcs",
			PurchasePrice: 800000,
			SalePrice:     1200000,
			Stock:         25,
			MinStock:      10,
			MaxStock:      100,
			ReorderLevel:  15,
			SKU:           "LOG-MX3-WL",
			IsActive:      true,
		},
		{
			Code:          "PRD003",
			Name:          "Kertas A4 80gsm",
			Description:   "Kertas A4 80gsm per rim (500 lembar)",
			CategoryID:    &category.ID,
			Unit:          "rim",
			PurchasePrice: 45000,
			SalePrice:     65000,
			Stock:         100,
			MinStock:      20,
			MaxStock:      500,
			ReorderLevel:  30,
			SKU:           "PAPER-A4-80G",
			IsActive:      true,
		},
	}

	for _, product := range products {
		db.Create(&product)
	}
}

func seedExpenseCategories(db *gorm.DB) {
	// Check if expense categories already exist
	var count int64
	db.Model(&models.ExpenseCategory{}).Count(&count)
	if count > 0 {
		return
	}

	categories := []models.ExpenseCategory{
		{Code: "EXP001", Name: "Operasional", Description: "Biaya operasional harian"},
		{Code: "EXP002", Name: "Marketing", Description: "Biaya pemasaran dan promosi"},
		{Code: "EXP003", Name: "Administrasi", Description: "Biaya administrasi dan umum"},
		{Code: "EXP004", Name: "Transportasi", Description: "Biaya transportasi dan perjalanan"},
	}

	for _, category := range categories {
		db.Create(&category)
	}
}

func seedCashBankAccounts(db *gorm.DB) {
	// Check if cash bank accounts already exist
	var count int64
	db.Model(&models.CashBank{}).Count(&count)
	if count > 0 {
		return
	}

	cashBanks := []models.CashBank{
		{
			Code:     "CASH001",
			Name:     "Kas Besar",
			Type:     models.CashBankTypeCash,
			Balance:  5000000,
			IsActive: true,
		},
		{
			Code:     "BANK001",
			Name:     "Bank BCA - Operasional",
			Type:     models.CashBankTypeBank,
			Balance:  50000000,
			IsActive: true,
		},
		{
			Code:     "BANK002",
			Name:     "Bank Mandiri - Payroll",
			Type:     models.CashBankTypeBank,
			Balance:  25000000,
			IsActive: true,
		},
	}

	for _, cashBank := range cashBanks {
		db.Create(&cashBank)
	}
}

func seedCompanyProfile(db *gorm.DB) {
	// Check if company profile already exists
	var count int64
	db.Model(&models.CompanyProfile{}).Count(&count)
	if count > 0 {
		return
	}

	company := models.CompanyProfile{
		Name:            "PT Contoh Perusahaan",
		LegalName:       "PT Contoh Perusahaan Tbk",
		TaxNumber:       "01.234.567.8-901.000",
		RegistrationNumber: "AHU-123456789",
		Industry:        "Perdagangan",
		Address:         "Jl. Contoh No. 123, Jakarta Selatan",
		City:           "Jakarta",
		State:          "DKI Jakarta",
		PostalCode:     "12345",
		Country:        "Indonesia",
		Phone:          "021-12345678",
		Email:          "info@perusahaan.com",
		Website:        "www.perusahaan.com",
		FiscalYearStart: "01-01",
		Currency:       "IDR",
		IsActive:       true,
	}

	db.Create(&company)
}

func seedReportTemplates(db *gorm.DB) {
	// Check if report templates already exist
	var count int64
	db.Model(&models.ReportTemplate{}).Count(&count)
	if count > 0 {
		return
	}

	// Get first user for relation
	var user models.User
	db.First(&user)

	templates := []models.ReportTemplate{
		{
			Name:        "Neraca Standar",
			Type:        models.ReportTypeBalanceSheet,
			Description: "Template neraca standar dengan format Indonesia",
			Template:    `{"sections":["ASET","KEWAJIBAN","EKUITAS"],"format":"standard"}`,
			IsDefault:   true,
			IsActive:    true,
			UserID:      user.ID,
		},
		{
			Name:        "Laporan Laba Rugi",
			Type:        models.ReportTypeIncomeStatement,
			Description: "Template laporan laba rugi dengan format Indonesia",
			Template:    `{"sections":["PENDAPATAN","BEBAN","LABA_BERSIH"],"format":"standard"}`,
			IsDefault:   true,
			IsActive:    true,
			UserID:      user.ID,
		},
		{
			Name:        "Neraca Saldo",
			Type:        models.ReportTypeTrialBalance,
			Description: "Template neraca saldo untuk semua akun",
			Template:    `{"columns":["code","name","debit","credit"],"format":"detailed"}`,
			IsDefault:   true,
			IsActive:    true,
			UserID:      user.ID,
		},
	}

	for _, template := range templates {
		db.Create(&template)
	}
}
