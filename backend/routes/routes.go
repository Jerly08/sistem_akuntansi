package routes

import (
	"net/http"
	"os"
	"strings"
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/handlers"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Environment detection helper
func getEnvironment() string {
	env := strings.ToLower(os.Getenv("ENV"))
	if env == "" {
		env = strings.ToLower(os.Getenv("GO_ENV"))
	}
	if env == "" {
		env = strings.ToLower(os.Getenv("ENVIRONMENT"))
	}
	if env == "" {
		env = "development" // default
	}
	return env
}

// Check if development features should be enabled
func isDevelopmentMode() bool {
	env := getEnvironment()
	return env == "development" || env == "dev" || env == "local"
}

// Check if debug routes should be enabled (requires explicit flag)
func shouldEnableDebugRoutes() bool {
	return os.Getenv("ENABLE_DEBUG_ROUTES") == "true" && isDevelopmentMode()
}

func SetupRoutes(r *gin.Engine, db *gorm.DB, startupService *services.StartupService) {
	// Controllers
	authController := controllers.NewAuthController(db)
	userController := controllers.NewUserController(db)
	permissionController := controllers.NewPermissionController(db)
	productController := controllers.NewProductController(db)
	categoryController := controllers.NewCategoryController(db)
	unitController := controllers.NewProductUnitController(db)
	inventoryController := controllers.NewInventoryController(db)
	assetController := controllers.NewAssetController(db)
	debugController := controllers.NewDebugController()
	monitoringController := controllers.NewMonitoringController()
	
	// Initialize repositories, services and handlers
	accountRepo := repositories.NewAccountRepository(db)
	exportService := services.NewExportService(accountRepo)
	accountHandler := handlers.NewAccountHandler(accountRepo, exportService)
	
	// Initialize startup handler for startup service monitoring
	startupHandler := handlers.NewStartupHandler(startupService)
	
	// Contact repositories, services and controllers
	contactRepo := repositories.NewContactRepository(db)
	contactService := services.NewContactService(contactRepo)
	contactController := controllers.NewContactController(contactService)
	
	// Notification repositories, services and handlers
	notificationRepo := repositories.NewNotificationRepository(db)
	notificationService := services.NewNotificationService(db, notificationRepo)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	
	// Purchase repositories, services and controllers
	purchaseRepo := repositories.NewPurchaseRepository(db)
	productRepo := repositories.NewProductRepository(db)
	approvalService := services.NewApprovalService(db)
	// Initialize services needed for purchase service
	journalRepo := repositories.NewJournalEntryRepository(db)
	pdfService := services.NewPDFService()
	purchaseService := services.NewPurchaseService(
		db,
		purchaseRepo,
		productRepo, 
		contactRepo,
		accountRepo,
		approvalService,
		nil, // journal service - can be nil for now
		journalRepo,
		pdfService,
	)
	purchaseController := controllers.NewPurchaseController(purchaseService)
	// Handlers that depend on services
	purchaseApprovalHandler := handlers.NewPurchaseApprovalHandler(purchaseService, approvalService)
	
	// Initialize security middleware
	middleware.InitAuditLogger(db)       // Initialize audit logging
	middleware.InitTokenMonitor(db)      // Initialize token monitoring
	
	// Initialize Security controller for security dashboard
	securityController := controllers.NewSecurityController(db)
	
	// Initialize JWT Manager
	jwtManager := middleware.NewJWTManager(db)
	
	// Initialize Permission Middleware
	permMiddleware := middleware.NewPermissionMiddleware(db)
	
	// 🔒 Initialize Enhanced Security Middleware
	enhancedSecurity := middleware.NewEnhancedSecurityMiddleware(db)
	
	// 🎛️ Apply global security middleware
	r.Use(enhancedSecurity.SecurityHeaders())     // Security headers pada semua requests
	r.Use(enhancedSecurity.RequestMonitoring())   // Monitor semua requests untuk threats
	
	// Initialize Stock Monitoring service and Dashboard controller
	stockMonitoringService := services.NewStockMonitoringService(db, notificationService)
	dashboardController := controllers.NewDashboardController(db, stockMonitoringService)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// 🔐 Auth routes (minimal public access)
		auth := v1.Group("/auth")
		auth.Use(middleware.AuthRateLimit()) // Apply auth rate limiting
		auth.Use(enhancedSecurity.SecurityHeaders()) // Extra security for auth endpoints
		{
			auth.POST("/login", authController.Login)
			// 🔒 PRODUCTION: Disable register endpoint in production
			if isDevelopmentMode() || os.Getenv("ALLOW_REGISTRATION") == "true" {
				auth.POST("/register", authController.Register)
			}
			auth.POST("/refresh", authController.RefreshToken)
			
			// Token validation endpoint (requires auth)
			auth.GET("/validate-token", jwtManager.AuthRequired(), authController.ValidateToken)
		}

		// 🔒 SECURITY: Secure debug routes (development only dengan multiple security layers)
		if shouldEnableDebugRoutes() {
			debugAuth := v1.Group("/debug")
			debugAuth.Use(enhancedSecurity.EnvironmentGate("development", "dev")) // ✅ Environment restriction
			debugAuth.Use(enhancedSecurity.IPWhitelist())                        // ✅ IP whitelisting
			debugAuth.Use(jwtManager.AuthRequired())                            // ✅ Authentication required
			debugAuth.Use(middleware.RoleRequired("admin"))                     // ✅ Admin only access
			debugAuth.Use(middleware.RateLimit())                               // ✅ Rate limiting
			{
				// 🔍 Read-only endpoints untuk debugging (safe)
				debugAuth.GET("/contacts", contactController.GetContacts)
				debugAuth.GET("/contacts/:id", contactController.GetContact)
				debugAuth.GET("/contacts/type/:type", contactController.GetContactsByType)
				debugAuth.GET("/contacts/search", contactController.SearchContacts)
				
				// 📊 Debug system information
				// debugAuth.GET("/system/info", debugController.GetSystemInfo)
				// debugAuth.GET("/database/health", debugController.GetDatabaseHealth)
				
				// ⚠️  ALL WRITE OPERATIONS COMPLETELY REMOVED FOR SECURITY
				// No CREATE/UPDATE/DELETE operations allowed in any debug route
			}
		}

		// Protected routes (auth required)
		protected := v1.Group("")
		protected.Use(jwtManager.AuthRequired())
		{
			// Profile routes
			protected.GET("/profile", authController.Profile)

			// User management routes (admin only)
			users := protected.Group("/users")
			{
				users.GET("", middleware.RoleRequired("admin"), userController.GetUsers)
				users.GET("/:id", middleware.RoleRequired("admin"), userController.GetUser)
				users.POST("", middleware.RoleRequired("admin"), userController.CreateUser)
				users.PUT("/:id", middleware.RoleRequired("admin"), userController.UpdateUser)
				users.DELETE("/:id", middleware.RoleRequired("admin"), userController.DeleteUser)
			}
			
			// Permission management routes
			permissions := protected.Group("/permissions")
			{
				// Admin only routes
				permissions.GET("/users", middleware.RoleRequired("admin"), permissionController.GetAllUsersPermissions)
				permissions.GET("/users/:userId", middleware.RoleRequired("admin"), permissionController.GetUserPermissions)
				permissions.PUT("/users/:userId", middleware.RoleRequired("admin"), permissionController.UpdateUserPermissions)
				permissions.POST("/users/:userId/reset", middleware.RoleRequired("admin"), permissionController.ResetToDefaultPermissions)
				
				// Self permission routes (any authenticated user)
				permissions.GET("/me", permissionController.GetMyPermissions) // User can get their own permissions
				permissions.GET("/check", permissionController.CheckUserPermission) // User can check their own permission
			}

			// Dashboard routes
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/analytics", middleware.RoleRequired("admin", "finance", "director"), dashboardController.GetAnalytics)
				dashboard.GET("/summary", middleware.RoleRequired("admin", "finance", "director", "inventory_manager", "employee"), dashboardController.GetDashboardSummary)
				dashboard.GET("/quick-stats", middleware.RoleRequired("admin", "finance", "director", "inventory_manager", "employee"), dashboardController.GetQuickStats)
				dashboard.GET("/stock-alerts", middleware.RoleRequired("admin", "inventory_manager", "director"), dashboardController.GetStockAlertsBanner)
				dashboard.POST("/stock-alerts/:id/dismiss", middleware.RoleRequired("admin", "inventory_manager"), dashboardController.DismissStockAlert)
			}

			// 📦 Product routes with enhanced permission checks dan inventory monitoring
			products := protected.Group("/products")
		products.Use(enhancedSecurity.RequestMonitoring()) // 📊 Enhanced monitoring
		// if middleware.GlobalAuditLogger != nil {
		//	products.Use(middleware.GlobalAuditLogger.InventoryAuditMiddleware()) // 📋 Inventory audit
		// }
			{
				// Basic CRUD operations dengan enhanced security
				products.GET("", permMiddleware.CanView("products"), productController.GetProducts)
				products.GET("/:id", permMiddleware.CanView("products"), productController.GetProduct)
				products.POST("", permMiddleware.CanCreate("products"), productController.CreateProduct)
				products.PUT("/:id", permMiddleware.CanEdit("products"), productController.UpdateProduct)
				products.DELETE("/:id", permMiddleware.CanDelete("products"), productController.DeleteProduct)
				
				// 📊 Critical inventory operations dengan extra monitoring
				products.POST("/adjust-stock", permMiddleware.CanEdit("products"), enhancedSecurity.RequestMonitoring(), productController.AdjustStock)
				products.POST("/opname", permMiddleware.CanEdit("products"), enhancedSecurity.RequestMonitoring(), productController.Opname)
				products.POST("/upload-image", permMiddleware.CanEdit("products"), productController.UploadProductImage)
			}

			// Category routes
			categories := protected.Group("/categories")
			{
				categories.GET("", middleware.RoleRequired("admin", "inventory_manager", "employee", "director", "finance"), categoryController.GetCategories)
				categories.GET("/tree", middleware.RoleRequired("admin", "inventory_manager", "employee", "director", "finance"), categoryController.GetCategoryTree)
				categories.GET("/:id", middleware.RoleRequired("admin", "inventory_manager", "employee", "director", "finance"), categoryController.GetCategory)
				categories.GET("/:id/products", middleware.RoleRequired("admin", "inventory_manager", "employee", "director", "finance"), categoryController.GetCategoryProducts)
				categories.POST("", middleware.RoleRequired("admin", "inventory_manager"), categoryController.CreateCategory)
				categories.PUT("/:id", middleware.RoleRequired("admin", "inventory_manager"), categoryController.UpdateCategory)
				categories.DELETE("/:id", middleware.RoleRequired("admin"), categoryController.DeleteCategory)
			}

			// Product Units routes
			units := protected.Group("/product-units")
			{
				units.GET("", middleware.RoleRequired("admin", "inventory_manager", "employee", "director", "finance"), unitController.GetProductUnits)
				units.GET("/:id", middleware.RoleRequired("admin", "inventory_manager", "employee", "director", "finance"), unitController.GetProductUnit)
				units.POST("", middleware.RoleRequired("admin"), unitController.CreateProductUnit)
				units.PUT("/:id", middleware.RoleRequired("admin"), unitController.UpdateProductUnit)
				units.DELETE("/:id", middleware.RoleRequired("admin"), unitController.DeleteProductUnit)
			}

			// 📊 Account routes (Chart of Accounts) dengan enhanced security
			accounts := protected.Group("/accounts")
		accounts.Use(enhancedSecurity.RequestMonitoring()) // 📊 Enhanced monitoring
		// if middleware.GlobalAuditLogger != nil {
		//	accounts.Use(middleware.GlobalAuditLogger.AccountAuditMiddleware()) // 📋 Financial audit
		// }
			{
				accounts.GET("", permMiddleware.CanView("accounts"), accountHandler.ListAccounts)
				
				// Get account catalog (minimal EXPENSE data) - accessible by EMPLOYEE for purchases
				accounts.GET("/catalog", permMiddleware.CanView("accounts"), accountHandler.GetAccountCatalog)
				
				accounts.GET("/hierarchy", permMiddleware.CanView("accounts"), accountHandler.GetAccountHierarchy)
				accounts.GET("/balance-summary", permMiddleware.CanView("accounts"), accountHandler.GetBalanceSummary)
				accounts.GET("/validate-code", permMiddleware.CanView("accounts"), accountHandler.ValidateAccountCode)
				
				// Fix account header status
				accounts.POST("/fix-header-status", middleware.RoleRequired("admin"), accountHandler.FixAccountHeaderStatus)
				
				accounts.GET("/:code", permMiddleware.CanView("accounts"), accountHandler.GetAccount)
				accounts.POST("", permMiddleware.CanCreate("accounts"), accountHandler.CreateAccount)
				accounts.PUT("/:code", permMiddleware.CanEdit("accounts"), accountHandler.UpdateAccount)
				accounts.DELETE("/:code", permMiddleware.CanDelete("accounts"), accountHandler.DeleteAccount)
				accounts.POST("/import", permMiddleware.CanCreate("accounts"), accountHandler.ImportAccounts)
				
				// Export routes
				accounts.GET("/export/pdf", permMiddleware.CanExport("accounts"), accountHandler.ExportAccountsPDF)
				accounts.GET("/export/excel", permMiddleware.CanExport("accounts"), accountHandler.ExportAccountsExcel)
			}

			// 📞 Contact routes with enhanced permission checks dan audit logging
			contacts := protected.Group("/contacts")
		contacts.Use(enhancedSecurity.RequestMonitoring()) // 📊 Enhanced monitoring
		// if middleware.GlobalAuditLogger != nil {
		//	contacts.Use(middleware.GlobalAuditLogger.ContactAuditMiddleware()) // 📋 Audit logging
		// }
			{
				// Basic CRUD operations dengan enhanced security
				contacts.GET("", permMiddleware.CanView("contacts"), contactController.GetContacts)
				contacts.GET("/:id", permMiddleware.CanView("contacts"), contactController.GetContact)
				contacts.POST("", permMiddleware.CanCreate("contacts"), contactController.CreateContact)
				contacts.PUT("/:id", permMiddleware.CanEdit("contacts"), contactController.UpdateContact)
				contacts.DELETE("/:id", permMiddleware.CanDelete("contacts"), contactController.DeleteContact)
				
				// Advanced operations
				contacts.GET("/type/:type", permMiddleware.CanView("contacts"), contactController.GetContactsByType)
				contacts.GET("/search", permMiddleware.CanView("contacts"), contactController.SearchContacts)
				
				// Import/Export operations
				contacts.POST("/import", permMiddleware.CanCreate("contacts"), contactController.ImportContacts)
				contacts.GET("/export", permMiddleware.CanExport("contacts"), contactController.ExportContacts)
				
				// Address management
				contacts.POST("/:id/addresses", permMiddleware.CanEdit("contacts"), contactController.AddContactAddress)
				contacts.PUT("/:id/addresses/:address_id", permMiddleware.CanEdit("contacts"), contactController.UpdateContactAddress)
				contacts.DELETE("/:id/addresses/:address_id", permMiddleware.CanEdit("contacts"), contactController.DeleteContactAddress)
			}

			// Sales repositories, services and controllers
			salesRepo := repositories.NewSalesRepository(db)
			productRepo := repositories.NewProductRepository(db)
			// Note: pdfService is already initialized earlier for purchase service
		salesService := services.NewSalesService(db, salesRepo, productRepo, contactRepo, accountRepo, nil, pdfService)
			salesController := controllers.NewSalesController(salesService)

			// Notification routes
			notifs := protected.Group("/notifications")
			{
				notifs.GET("", middleware.RoleRequired("admin", "finance", "director", "inventory_manager", "employee"), notificationHandler.GetNotifications)
				notifs.GET("/unread-count", middleware.RoleRequired("admin", "finance", "director", "inventory_manager", "employee"), notificationHandler.GetUnreadCount)
				notifs.PUT("/:id/read", middleware.RoleRequired("admin", "finance", "director", "inventory_manager", "employee"), notificationHandler.MarkNotificationAsRead)
				notifs.PUT("/read-all", middleware.RoleRequired("admin", "finance", "director", "inventory_manager", "employee"), notificationHandler.MarkAllNotificationsAsRead)
				notifs.GET("/type/:type", middleware.RoleRequired("admin", "finance", "director", "inventory_manager", "employee"), notificationHandler.GetNotificationsByType)
				notifs.GET("/approvals", middleware.RoleRequired("admin", "finance", "director", "inventory_manager", "employee"), notificationHandler.GetApprovalNotifications)
			}

			// Sales routes with permission checks
			sales := protected.Group("/sales")
			{
				// Basic CRUD operations
				sales.GET("", permMiddleware.CanView("sales"), salesController.GetSales)
				sales.GET("/:id", permMiddleware.CanView("sales"), salesController.GetSale)
				sales.POST("", permMiddleware.CanCreate("sales"), salesController.CreateSale)
				sales.PUT("/:id", permMiddleware.CanEdit("sales"), salesController.UpdateSale)
				sales.DELETE("/:id", permMiddleware.CanDelete("sales"), salesController.DeleteSale)

				// Status management
				sales.POST("/:id/confirm", middleware.RoleRequired("admin", "finance", "director"), salesController.ConfirmSale)
				sales.POST("/:id/invoice", middleware.RoleRequired("admin", "finance", "director"), salesController.InvoiceSale)
				sales.POST("/:id/cancel", middleware.RoleRequired("admin", "finance", "director"), salesController.CancelSale)

				// Payment management
				sales.GET("/:id/payments", middleware.RoleRequired("admin", "finance", "director", "employee"), salesController.GetSalePayments)
				sales.POST("/:id/payments", middleware.RoleRequired("admin", "finance", "director"), salesController.CreateSalePayment)

				// Returns management
				sales.POST("/:id/returns", middleware.RoleRequired("admin", "finance", "director"), salesController.CreateSaleReturn)
				sales.GET("/returns", middleware.RoleRequired("admin", "finance", "director"), salesController.GetSaleReturns)

				// Analytics and reporting
				sales.GET("/summary", middleware.RoleRequired("admin", "finance", "director", "employee"), salesController.GetSalesSummary)
				sales.GET("/analytics", middleware.RoleRequired("admin", "finance", "director"), salesController.GetSalesAnalytics)
				sales.GET("/receivables", middleware.RoleRequired("admin", "finance", "director"), salesController.GetReceivablesReport)

				// PDF exports
				sales.GET("/:id/invoice/pdf", middleware.RoleRequired("admin", "finance", "director"), salesController.ExportSaleInvoicePDF)
				sales.GET("/report/pdf", middleware.RoleRequired("admin", "finance", "director"), salesController.ExportSalesReportPDF)

				// Customer portal
				sales.GET("/customer/:customer_id", middleware.RoleRequired("admin", "finance", "director"), salesController.GetCustomerSales)
				sales.GET("/customer/:customer_id/invoices", middleware.RoleRequired("admin", "finance", "director"), salesController.GetCustomerInvoices)
			}

	// Initialize Payment repositories, services and controllers
	paymentRepo := repositories.NewPaymentRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	paymentService := services.NewPaymentService(db, paymentRepo, salesRepo, purchaseRepo, cashBankRepo, accountRepo, contactRepo)
	paymentController := controllers.NewPaymentController(paymentService)
	cashBankService := services.NewCashBankService(db, cashBankRepo, accountRepo)
	cashBankController := controllers.NewCashBankController(cashBankService)
	
	// Initialize Balance Monitoring service and controller
	balanceMonitoringService := services.NewBalanceMonitoringService(db)
	balanceMonitoringController := controllers.NewBalanceMonitoringController(balanceMonitoringService)
			
			// Setup Payment routes (including cash bank routes with GL fix functionality)
			SetupPaymentRoutes(protected, paymentController, cashBankController, cashBankService, jwtManager, db)

			// Purchases routes
			purchases := protected.Group("/purchases")
			{
				// Basic CRUD operations
				purchases.GET("", middleware.RoleRequired("admin", "finance", "inventory_manager", "employee", "director"), purchaseController.GetPurchases)
				// Approval statistics (must be defined before parameterized "/:id" route)
				purchases.GET("/approval-stats", middleware.RoleRequired("admin", "finance", "director"), purchaseApprovalHandler.GetApprovalStats)
				purchases.GET("/:id", middleware.RoleRequired("admin", "finance", "inventory_manager", "employee", "director"), purchaseController.GetPurchase)
				purchases.POST("", middleware.RoleRequired("admin", "finance", "inventory_manager", "employee", "director"), purchaseController.CreatePurchase)
				purchases.PUT("/:id", middleware.RoleRequired("admin", "finance", "inventory_manager"), purchaseController.UpdatePurchase)
				purchases.DELETE("/:id", middleware.RoleRequired("admin"), purchaseController.DeletePurchase)
				
				// Approval operations
				purchases.POST("/:id/submit-approval", middleware.RoleRequired("admin", "finance", "inventory_manager", "employee", "director"), purchaseController.SubmitForApproval)
				purchases.POST("/:id/approve", middleware.RoleRequired("admin", "finance", "director"), purchaseController.ApprovePurchase)
				purchases.POST("/:id/reject", middleware.RoleRequired("admin", "finance", "director"), purchaseController.RejectPurchase)
				// Approval history endpoint (used by frontend)
				purchases.GET("/:id/approval-history", purchaseApprovalHandler.GetApprovalHistory)
				// Pending approvals (singular path for frontend compatibility)
				purchases.GET("/pending-approval", middleware.RoleRequired("admin", "finance", "director"), purchaseApprovalHandler.GetPurchasesForApproval)
				
				// Document management
				purchases.POST("/:id/documents", middleware.RoleRequired("admin", "finance", "inventory_manager", "director"), purchaseController.UploadDocument)
				purchases.GET("/:id/documents", middleware.RoleRequired("admin", "finance", "inventory_manager", "director"), purchaseController.GetPurchaseDocuments)
				purchases.DELETE("/documents/:document_id", middleware.RoleRequired("admin", "finance"), purchaseController.DeleteDocument)
				
				// Receipt operations
				purchases.POST("/receipts", middleware.RoleRequired("admin", "inventory_manager", "director"), purchaseController.CreatePurchaseReceipt)
				purchases.GET("/:id/receipts", middleware.RoleRequired("admin", "finance", "inventory_manager", "director"), purchaseController.GetPurchaseReceipts)
				
				// Analytics and reporting
				purchases.GET("/summary", middleware.RoleRequired("admin", "finance", "director", "employee"), purchaseController.GetPurchasesSummary)
				purchases.GET("/pending-approvals", middleware.RoleRequired("admin", "finance", "director"), purchaseController.GetPendingApprovals)
				purchases.GET("/dashboard", middleware.RoleRequired("admin", "finance", "inventory_manager", "director"), purchaseController.GetPurchaseDashboard)
				purchases.GET("/vendor/:vendor_id/summary", middleware.RoleRequired("admin", "finance"), purchaseController.GetVendorPurchaseSummary)
				
				// Three-way matching
				purchases.GET("/:id/matching", middleware.RoleRequired("admin", "finance"), purchaseController.GetPurchaseMatching)
				purchases.POST("/:id/validate-matching", middleware.RoleRequired("admin", "finance"), purchaseController.ValidateThreeWayMatching)
			}

			// Expenses routes
			expenses := protected.Group("/expenses")
			{
				expenses.GET("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Expenses endpoint - coming soon"})
				})
			}

			// Assets routes
			assets := protected.Group("/assets")
			{
				// Basic CRUD operations
				assets.GET("", middleware.RoleRequired("admin", "finance", "director"), assetController.GetAssets)
				assets.GET("/:id", middleware.RoleRequired("admin", "finance", "director"), assetController.GetAsset)
				assets.POST("", middleware.RoleRequired("admin"), assetController.CreateAsset)
				assets.PUT("/:id", middleware.RoleRequired("admin"), assetController.UpdateAsset)
				assets.DELETE("/:id", middleware.RoleRequired("admin"), assetController.DeleteAsset)
				assets.POST("/upload-image", middleware.RoleRequired("admin"), assetController.UploadAssetImage)
				
				// Reports and calculations
				assets.GET("/summary", middleware.RoleRequired("admin", "finance", "director"), assetController.GetAssetsSummary)
				assets.GET("/depreciation-report", middleware.RoleRequired("admin", "finance", "director"), assetController.GetDepreciationReport)
				assets.GET("/:id/depreciation-schedule", middleware.RoleRequired("admin", "finance"), assetController.GetDepreciationSchedule)
				assets.GET("/:id/calculate-depreciation", middleware.RoleRequired("admin", "finance"), assetController.CalculateCurrentDepreciation)
			}

		// Note: CashBank routes are already set up via SetupPaymentRoutes

			// Inventory routes
			inventory := protected.Group("/inventory")
			{
				inventory.GET("/movements", inventoryController.GetInventoryMovements)
				inventory.GET("/low-stock", inventoryController.GetLowStockProducts)
				inventory.GET("/valuation", inventoryController.GetStockValuation)
				inventory.GET("/report", middleware.RoleRequired("admin", "inventory_manager"), inventoryController.GetStockReport)
				inventory.POST("/bulk-price-update", middleware.RoleRequired("admin", "inventory_manager"), inventoryController.BulkPriceUpdate)
			}

			// Approval workflows routes
			workflows := protected.Group("/approval-workflows")
			{
				workflows.GET("", purchaseApprovalHandler.GetApprovalWorkflows)
				workflows.POST("", middleware.RoleRequired("admin"), purchaseApprovalHandler.CreateApprovalWorkflow)
			}

			// Initialize Report services and controller
			reportService := services.NewReportService(db, accountRepo, salesRepo, purchaseRepo, productRepo, contactRepo, paymentRepo, cashBankRepo)
			professionalService := services.NewProfessionalReportService(db, accountRepo, salesRepo, purchaseRepo, productRepo, contactRepo, paymentRepo, cashBankRepo)
			standardizedService := services.NewStandardizedReportService(db, accountRepo, salesRepo, purchaseRepo, productRepo, contactRepo, paymentRepo, cashBankRepo)
			reportController := controllers.NewReportController(reportService, professionalService, standardizedService)
			
			// Initialize Financial Report services and controller
			// Note: journalRepo is already initialized earlier for purchase service
			financialReportService := services.NewFinancialReportService(db, accountRepo, journalRepo)
			financialReportController := controllers.NewFinancialReportController(financialReportService)
			
			// Setup Settings routes
			SetupSettingsRoutes(protected, db)
			
			// Setup Report routes - Using single consolidated report controller
			SetupReportRoutes(protected, reportController)
			
			// Setup Financial Report routes (enhanced endpoints under /reports/enhanced)
			SetupFinancialReportRoutes(protected, financialReportController)
			
			// NOTE: Unified report routes are commented out to avoid duplicate registrations
			// The main report routes already handle all necessary endpoints at /api/v1/reports/*
			// If unified reports are needed in the future, they should use a different path
			// like /api/v1/unified-reports to avoid conflicts
			
			// // Setup Unified Report Controller and Routes
			// balanceSheetService := services.NewStandardizedReportService(db, accountRepo, salesRepo, purchaseRepo, productRepo, contactRepo, paymentRepo, cashBankRepo)
			// profitLossService := services.NewEnhancedProfitLossService(db, accountRepo)
			// cashFlowService := services.NewStandardizedReportService(db, accountRepo, salesRepo, purchaseRepo, productRepo, contactRepo, paymentRepo, cashBankRepo)
			// 
			// unifiedReportController := controllers.NewUnifiedReportController(
			// 	db,
			// 	accountRepo,
			// 	salesRepo,
			// 	purchaseRepo,
			// 	contactRepo,
			// 	productRepo,
			// 	reportService,
			// 	balanceSheetService,
			// 	profitLossService,
			// 	cashFlowService,
			// )
			// 
			// // Register unified report routes
			// RegisterUnifiedReportRoutes(r, unifiedReportController, jwtManager)
			// RegisterUnifiedReportMiddleware(r)
			
			// Setup Unified Financial Report Routes (at /api/unified-reports - different path)
			SetupUnifiedReportRoutes(r, db)

			// Monitoring routes (admin only)
			monitoring := protected.Group("/monitoring")
			monitoring.Use(middleware.RoleRequired("admin")) // Only admins can access monitoring
			{
				// System monitoring
				monitoring.GET("/status", monitoringController.GetSystemSecurityStatus)
				monitoring.GET("/rate-limits", monitoringController.GetRateLimitStatus)
				monitoring.GET("/security-alerts", monitoringController.GetSecurityAlerts)

				// Audit logging
				monitoring.GET("/audit-logs", monitoringController.GetAuditLogs)

				// Token monitoring
				monitoring.GET("/token-stats", monitoringController.GetTokenStats)
				monitoring.GET("/refresh-events", monitoringController.GetRecentRefreshEvents)

				// User-specific monitoring
				monitoring.GET("/users/:user_id/security-summary", monitoringController.GetUserSecuritySummary)
				
				// Startup service monitoring
				monitoring.GET("/startup-status", startupHandler.GetStartupStatus)
				monitoring.POST("/fix-account-headers", startupHandler.TriggerAccountHeaderFix)
				
				// Balance monitoring routes
				monitoring.GET("/balance-sync", balanceMonitoringController.CheckBalanceSync)
				monitoring.POST("/fix-discrepancies", balanceMonitoringController.FixBalanceDiscrepancies)
				monitoring.GET("/balance-health", balanceMonitoringController.GetBalanceHealth)
				monitoring.GET("/discrepancies", balanceMonitoringController.GetBalanceDiscrepancies)
				monitoring.GET("/sync-status", balanceMonitoringController.GetSyncStatus)
			}
			
			// 🔒 Security Dashboard routes (admin only) 
			security := protected.Group("/admin/security")
			security.Use(middleware.RoleRequired("admin")) // Only admins can access security dashboard
			security.Use(enhancedSecurity.RequestMonitoring()) // Enhanced monitoring for security routes
			{
				// Security Incident Management
				security.GET("/incidents", securityController.GetSecurityIncidents)
				security.GET("/incidents/:id", securityController.GetSecurityIncident)
				security.PUT("/incidents/:id/resolve", securityController.ResolveSecurityIncident)
				
				// System Alerts Management
				security.GET("/alerts", securityController.GetSystemAlerts)
				security.PUT("/alerts/:id/acknowledge", securityController.AcknowledgeAlert)
				
				// Security Metrics & Analytics
				security.GET("/metrics", securityController.GetSecurityMetrics)
				
				// IP Whitelist Management
				security.GET("/ip-whitelist", securityController.GetIPWhitelist)
				security.POST("/ip-whitelist", securityController.AddIPToWhitelist)
				
				// Security Configuration
				security.GET("/config", securityController.GetSecurityConfig)
				
				// Maintenance Operations
				security.POST("/cleanup", securityController.CleanupSecurityLogs)
			}
		}
	}

	// Static files (templates and uploads)
	r.Static("/templates", "./templates")
	r.Static("/uploads", "./uploads")

	// Health check endpoint
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Debug routes for token validation testing
	debug := v1.Group("/debug")
	{
		// Route with JWT middleware to test context
		debugWithAuth := debug.Group("/auth")
		debugWithAuth.Use(jwtManager.AuthRequired())
		{
			debugWithAuth.GET("/context", debugController.TestJWTContext)
			
			// This checks the role directly
			debugWithAuth.GET("/role", debugController.TestRolePermission)
			
			// This uses the RoleRequired middleware
			debugWithAuth.GET("/admin-only", middleware.RoleRequired("admin"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "You have admin role!"})
			})
			
			debugWithAuth.GET("/finance-only", middleware.RoleRequired("finance"), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "You have finance role!"})
			})
		}
	}
}
