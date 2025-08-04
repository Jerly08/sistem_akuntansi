package database

import (
	"log"
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
	
	// Seed Permissions
	seedPermissions(db)
	
	// Seed Role Permissions
	seedRolePermissions(db)

	log.Println("Database seeding completed successfully")
}

func seedUsers(db *gorm.DB) {
	// Seed all users for all roles
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	allUsers := []models.User{
		{
			Username:  "admin",
			Email:     "admin@company.com",
			Password:  string(hashedPassword),
			Role:      "admin",
			FirstName: "Admin",
			LastName:  "User",
			IsActive:  true,
		},
		{
			Username:  "finance",
			Email:     "finance@company.com",
			Password:  string(hashedPassword),
			Role:      "finance",
			FirstName: "Finance",
			LastName:  "User",
			IsActive:  true,
		},
		{
			Username:  "inventory",
			Email:     "inventory@company.com",
			Password:  string(hashedPassword),
			Role:      "inventory_manager",
			FirstName: "Inventory",
			LastName:  "User",
			IsActive:  true,
		},
		{
			Username:  "director",
			Email:     "director@company.com",
			Password:  string(hashedPassword),
			Role:      "director",
			FirstName: "Director",
			LastName:  "User",
			IsActive:  true,
		},
		{
			Username:  "operational",
			Email:     "operational@company.com",
			Password:  string(hashedPassword),
			Role:      "operational_user",
			FirstName: "Operational",
			LastName:  "User",
			IsActive:  true,
		},
		{
			Username:  "auditor",
			Email:     "auditor@company.com",
			Password:  string(hashedPassword),
			Role:      "auditor",
			FirstName: "Auditor",
			LastName:  "User",
			IsActive:  true,
		},
		{
			Username:  "employee",
			Email:     "employee@company.com",
			Password:  string(hashedPassword),
			Role:      "employee",
			FirstName: "Employee",
			LastName:  "User",
			IsActive:  true,
		},
	}
	
	// Add users if not existing
	for _, user := range allUsers {
		var existingUser models.User
		if err := db.Where("username = ?", user.Username).First(&existingUser).Error; err != nil {
			db.Create(&user)
		}
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

func seedPermissions(db *gorm.DB) {
	// Check if permissions already exist
	var count int64
	db.Model(&models.Permission{}).Count(&count)
	if count > 0 {
		return
	}

	permissions := []models.Permission{
		// User permissions
		{Name: "users:read", Resource: "users", Action: "read", Description: "View users"},
		{Name: "users:create", Resource: "users", Action: "create", Description: "Create users"},
		{Name: "users:update", Resource: "users", Action: "update", Description: "Update users"},
		{Name: "users:delete", Resource: "users", Action: "delete", Description: "Delete users"},
		{Name: "users:manage", Resource: "users", Action: "manage", Description: "Full user management"},

		// Account permissions
		{Name: "accounts:read", Resource: "accounts", Action: "read", Description: "View accounts"},
		{Name: "accounts:create", Resource: "accounts", Action: "create", Description: "Create accounts"},
		{Name: "accounts:update", Resource: "accounts", Action: "update", Description: "Update accounts"},
		{Name: "accounts:delete", Resource: "accounts", Action: "delete", Description: "Delete accounts"},

		// Transaction permissions
		{Name: "transactions:read", Resource: "transactions", Action: "read", Description: "View transactions"},
		{Name: "transactions:create", Resource: "transactions", Action: "create", Description: "Create transactions"},
		{Name: "transactions:update", Resource: "transactions", Action: "update", Description: "Update transactions"},
		{Name: "transactions:delete", Resource: "transactions", Action: "delete", Description: "Delete transactions"},

		// Product permissions
		{Name: "products:read", Resource: "products", Action: "read", Description: "View products"},
		{Name: "products:create", Resource: "products", Action: "create", Description: "Create products"},
		{Name: "products:update", Resource: "products", Action: "update", Description: "Update products"},
		{Name: "products:delete", Resource: "products", Action: "delete", Description: "Delete products"},

		// Sales permissions
		{Name: "sales:read", Resource: "sales", Action: "read", Description: "View sales"},
		{Name: "sales:create", Resource: "sales", Action: "create", Description: "Create sales"},
		{Name: "sales:update", Resource: "sales", Action: "update", Description: "Update sales"},
		{Name: "sales:delete", Resource: "sales", Action: "delete", Description: "Delete sales"},

		// Purchase permissions
		{Name: "purchases:read", Resource: "purchases", Action: "read", Description: "View purchases"},
		{Name: "purchases:create", Resource: "purchases", Action: "create", Description: "Create purchases"},
		{Name: "purchases:update", Resource: "purchases", Action: "update", Description: "Update purchases"},
		{Name: "purchases:delete", Resource: "purchases", Action: "delete", Description: "Delete purchases"},

		// Report permissions
		{Name: "reports:read", Resource: "reports", Action: "read", Description: "View reports"},
		{Name: "reports:create", Resource: "reports", Action: "create", Description: "Create reports"},
		{Name: "reports:update", Resource: "reports", Action: "update", Description: "Update reports"},
		{Name: "reports:delete", Resource: "reports", Action: "delete", Description: "Delete reports"},

		// Contact permissions
		{Name: "contacts:read", Resource: "contacts", Action: "read", Description: "View contacts"},
		{Name: "contacts:create", Resource: "contacts", Action: "create", Description: "Create contacts"},
		{Name: "contacts:update", Resource: "contacts", Action: "update", Description: "Update contacts"},
		{Name: "contacts:delete", Resource: "contacts", Action: "delete", Description: "Delete contacts"},

		// Asset permissions
		{Name: "assets:read", Resource: "assets", Action: "read", Description: "View assets"},
		{Name: "assets:create", Resource: "assets", Action: "create", Description: "Create assets"},
		{Name: "assets:update", Resource: "assets", Action: "update", Description: "Update assets"},
		{Name: "assets:delete", Resource: "assets", Action: "delete", Description: "Delete assets"},

		// Budget permissions
		{Name: "budgets:read", Resource: "budgets", Action: "read", Description: "View budgets"},
		{Name: "budgets:create", Resource: "budgets", Action: "create", Description: "Create budgets"},
		{Name: "budgets:update", Resource: "budgets", Action: "update", Description: "Update budgets"},
		{Name: "budgets:delete", Resource: "budgets", Action: "delete", Description: "Delete budgets"},
	}

	for _, permission := range permissions {
		db.Create(&permission)
	}
}

func seedRolePermissions(db *gorm.DB) {
	// Check if role permissions already exist
	var count int64
	db.Model(&models.RolePermission{}).Count(&count)
	if count > 0 {
		return
	}

	// Get all permissions
	var permissions []models.Permission
	db.Find(&permissions)

	permissionMap := make(map[string]uint)
	for _, perm := range permissions {
		permissionMap[perm.Name] = perm.ID
	}

	// Define role permissions
	rolePermissions := map[string][]string{
		"admin": {
			"users:read", "users:create", "users:update", "users:delete", "users:manage",
			"accounts:read", "accounts:create", "accounts:update", "accounts:delete",
			"transactions:read", "transactions:create", "transactions:update", "transactions:delete",
			"products:read", "products:create", "products:update", "products:delete",
			"sales:read", "sales:create", "sales:update", "sales:delete",
			"purchases:read", "purchases:create", "purchases:update", "purchases:delete",
			"reports:read", "reports:create", "reports:update", "reports:delete",
			"contacts:read", "contacts:create", "contacts:update", "contacts:delete",
			"assets:read", "assets:create", "assets:update", "assets:delete",
			"budgets:read", "budgets:create", "budgets:update", "budgets:delete",
		},
		"finance": {
			"accounts:read", "accounts:create", "accounts:update",
			"transactions:read", "transactions:create", "transactions:update",
			"sales:read", "sales:create", "sales:update",
			"purchases:read", "purchases:create", "purchases:update",
			"reports:read", "reports:create",
			"contacts:read", "contacts:update",
			"assets:read", "assets:update",
			"budgets:read", "budgets:create", "budgets:update",
		},
		"director": {
			"users:read",
			"accounts:read",
			"transactions:read",
			"products:read",
			"sales:read",
			"purchases:read",
			"reports:read", "reports:create",
			"contacts:read",
			"assets:read",
			"budgets:read", "budgets:create", "budgets:update",
		},
		"inventory_manager": {
			"products:read", "products:create", "products:update",
			"sales:read", "sales:create", "sales:update",
			"purchases:read", "purchases:create", "purchases:update",
			"contacts:read", "contacts:create", "contacts:update",
		},
		"employee": {
			"products:read",
			"sales:read", "sales:create",
			"contacts:read",
		},
		"auditor": {
			"users:read",
			"accounts:read",
			"transactions:read",
			"products:read",
			"sales:read",
			"purchases:read",
			"reports:read",
			"contacts:read",
			"assets:read",
			"budgets:read",
		},
		"operational_user": {
			"transactions:create", "transactions:read",
			"sales:create", "sales:read",
			"purchases:create", "purchases:read",
			"products:read",
			"contacts:read",
		},
	}

	// Create role permissions
	for role, perms := range rolePermissions {
		for _, permName := range perms {
			if permID, exists := permissionMap[permName]; exists {
				rolePermission := models.RolePermission{
					Role:         role,
					PermissionID: permID,
				}
				db.Create(&rolePermission)
			}
		}
	}
}
