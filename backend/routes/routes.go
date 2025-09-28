package routes

import (
	"log"
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
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	// Initialize Stock Monitoring service and Dashboard controller
	stockMonitoringService := services.NewStockMonitoringService(db, notificationService)
	notificationHandler := handlers.NewNotificationHandler(notificationService, stockMonitoringService)
	dashboardController := controllers.NewDashboardController(db, stockMonitoringService)
	
	// Update ProductController with stockMonitoringService
	productController := controllers.NewProductController(db, stockMonitoringService)
	
	// Initialize WarehouseLocationController
	warehouseLocationController := controllers.NewWarehouseLocationController(db)
	
	// Initialize SSOT Unified Journal Service first (needed by purchase service)
	unifiedJournalService := services.NewUnifiedJournalService(db)
	
	// Purchase repositories, services and controllers
	purchaseRepo := repositories.NewPurchaseRepository(db)
	productRepo := repositories.NewProductRepository(db)
	approvalService := services.NewApprovalService(db)
	// Initialize services needed for purchase service
	journalRepo := repositories.NewJournalEntryRepository(db)
	pdfService := services.NewPDFService(db)
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
		unifiedJournalService, // Add unified journal service for SSOT integration
	)
	// Handlers that depend on services (purchaseController will be initialized later)
	purchaseApprovalHandler := handlers.NewPurchaseApprovalHandler(purchaseService, approvalService)
	
	// Initialize security middleware
	middleware.InitAuditLogger(db)       // Initialize audit logging
	middleware.InitTokenMonitor(db)      // Initialize token monitoring
	
	// Initialize Security controller for security dashboard
	securityController := controllers.NewSecurityController(db)
	
	// Initialize Journal Drilldown controller
	journalDrilldownController := controllers.NewJournalDrilldownController(db)
	
	// Journal Entry controller removed - migrated to SSOT unified system
	
	// Initialize SSOT Unified Journal Controller (service already initialized above)
	unifiedJournalController := controllers.NewUnifiedJournalController(unifiedJournalService)
	
	// Initialize JWT Manager
	jwtManager := middleware.NewJWTManager(db)
	
	// Initialize Permission Middleware
	permMiddleware := middleware.NewPermissionMiddleware(db)
	
	// 🔒 Initialize Enhanced Security Middleware
	enhancedSecurity := middleware.NewEnhancedSecurityMiddleware(db)
	
	// 🎛️ Apply global security middleware
	r.Use(enhancedSecurity.SecurityHeaders())     // Security headers pada semua requests
	r.Use(enhancedSecurity.RequestMonitoring())   // Monitor semua requests untuk threats
	r.Use(middleware.APIUsageMiddleware())        // 📊 Track API usage for optimization
	

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
		
		// 📊 Journal Entry Drilldown routes at API level (accessible by finance, admin, director)
		// These routes are at /api/v1/journal-drilldown to match frontend expectations
		journalDrilldownAPI := v1.Group("/journal-drilldown")
		journalDrilldownAPI.Use(jwtManager.AuthRequired())
		{
			// Main drill-down endpoint for POST requests with detailed filtering
			journalDrilldownAPI.POST("", permMiddleware.CanView("reports"), journalDrilldownController.GetJournalDrilldown)
			
			// Alternative GET endpoint for simpler URL-based filtering  
			journalDrilldownAPI.GET("/entries", permMiddleware.CanView("reports"), journalDrilldownController.GetJournalDrilldownByParams)
			
			// Get detailed information for a specific journal entry
			journalDrilldownAPI.GET("/entries/:id", permMiddleware.CanView("reports"), journalDrilldownController.GetJournalEntryDetail)
			
			// Get accounts that have activity in a period (useful for filters)
			journalDrilldownAPI.GET("/accounts", permMiddleware.CanView("reports"), journalDrilldownController.GetAccountsForPeriod)
		}
		
		// 📋 Journal Entry Management routes (accessible by finance, admin, director)
		journalEntriesAPI := v1.Group("/journal-entries")
		journalEntriesAPI.Use(jwtManager.AuthRequired())
		{
			// CRUD operations for journal entries
			
			// Summary endpoint - MISSING ROUTE ADDED
		}
		
		// 📊 SSOT Journal System routes (NEW - unified journal management)
		unifiedJournals := v1.Group("/journals")
		unifiedJournals.Use(jwtManager.AuthRequired())
		{
			// Main CRUD operations
			unifiedJournals.POST("", permMiddleware.CanCreate("reports"), unifiedJournalController.CreateJournalEntry)
			unifiedJournals.GET("", permMiddleware.CanView("reports"), unifiedJournalController.GetJournalEntries)
			unifiedJournals.GET("/:id", permMiddleware.CanView("reports"), unifiedJournalController.GetJournalEntry)
			
			// Status operations - removed unused endpoints
			
			// Balance management
			unifiedJournals.GET("/account-balances", permMiddleware.CanView("reports"), unifiedJournalController.GetAccountBalances)
			unifiedJournals.POST("/account-balances/refresh", permMiddleware.CanEdit("reports"), unifiedJournalController.RefreshAccountBalances)
			
			// Summary and reporting
			unifiedJournals.GET("/summary", permMiddleware.CanView("reports"), unifiedJournalController.GetJournalSummary)
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

		// Public account catalog endpoints for purchase forms (minimal permission required)
		// These endpoints are needed for dropdown population in purchase forms
		v1.GET("/accounts/catalog", accountHandler.GetAccountCatalog)
		v1.GET("/accounts/credit", accountHandler.GetAccountCatalog)
		
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
				dashboard.GET("/analytics", permMiddleware.CanView("reports"), dashboardController.GetAnalytics)
				dashboard.GET("/finance", middleware.RoleRequired("admin", "finance"), dashboardController.GetFinanceDashboardData)
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

			// 📂 Category routes with enhanced permission checks
			categories := protected.Group("/categories")
			{
				categories.GET("", permMiddleware.CanView("products"), categoryController.GetCategories)
				categories.GET("/tree", permMiddleware.CanView("products"), categoryController.GetCategoryTree)
				categories.GET("/:id", permMiddleware.CanView("products"), categoryController.GetCategory)
				categories.GET("/:id/products", permMiddleware.CanView("products"), categoryController.GetCategoryProducts)
				categories.POST("", permMiddleware.CanCreate("products"), categoryController.CreateCategory)
				categories.PUT("/:id", permMiddleware.CanEdit("products"), categoryController.UpdateCategory)
				categories.DELETE("/:id", permMiddleware.CanDelete("products"), categoryController.DeleteCategory)
			}

			// 📏 Product Units routes with enhanced permission checks
			units := protected.Group("/product-units")
			{
				units.GET("", permMiddleware.CanView("products"), unitController.GetProductUnits)
				units.GET("/:id", permMiddleware.CanView("products"), unitController.GetProductUnit)
				units.POST("", permMiddleware.CanCreate("products"), unitController.CreateProductUnit)
				units.PUT("/:id", permMiddleware.CanEdit("products"), unitController.UpdateProductUnit)
				units.DELETE("/:id", permMiddleware.CanDelete("products"), unitController.DeleteProductUnit)
			}

			// 🏢 Warehouse Location routes with enhanced permission checks
			warehouseLocations := protected.Group("/warehouse-locations")
			{
				warehouseLocations.GET("", permMiddleware.CanView("products"), warehouseLocationController.GetWarehouseLocations)
				warehouseLocations.GET("/:id", permMiddleware.CanView("products"), warehouseLocationController.GetWarehouseLocation)
				warehouseLocations.POST("", permMiddleware.CanCreate("products"), warehouseLocationController.CreateWarehouseLocation)
				warehouseLocations.PUT("/:id", permMiddleware.CanEdit("products"), warehouseLocationController.UpdateWarehouseLocation)
				warehouseLocations.DELETE("/:id", permMiddleware.CanDelete("products"), warehouseLocationController.DeleteWarehouseLocation)
			}

			// 📊 Account routes (Chart of Accounts) dengan enhanced security
			accounts := protected.Group("/accounts")
	accounts.Use(enhancedSecurity.RequestMonitoring()) // 📊 Enhanced monitoring
	// if middleware.GlobalAuditLogger != nil {
	//	accounts.Use(middleware.GlobalAuditLogger.AccountAuditMiddleware()) // 📋 Financial audit
	// }
			{
				accounts.GET("", permMiddleware.CanView("accounts"), accountHandler.ListAccounts)
				
				accounts.GET("/hierarchy", permMiddleware.CanView("accounts"), accountHandler.GetAccountHierarchy)
				accounts.GET("/balance-summary", permMiddleware.CanView("accounts"), accountHandler.GetBalanceSummary)
				accounts.GET("/validate-code", permMiddleware.CanView("accounts"), accountHandler.ValidateAccountCode)
				
				// Fix account header status
				accounts.POST("/fix-header-status", middleware.RoleRequired("admin"), accountHandler.FixAccountHeaderStatus)
				
				accounts.GET("/:code", permMiddleware.CanView("accounts"), accountHandler.GetAccount)
				accounts.POST("", permMiddleware.CanCreate("accounts"), accountHandler.CreateAccount)
				accounts.PUT("/:code", permMiddleware.CanEdit("accounts"), accountHandler.UpdateAccount)
				accounts.DELETE("/:code", permMiddleware.CanDelete("accounts"), accountHandler.DeleteAccount)
				// Admin-only delete with cascade options
				accounts.DELETE("/admin/:code", middleware.RoleRequired("admin"), accountHandler.AdminDeleteAccount)
				accounts.POST("/import", permMiddleware.CanCreate("accounts"), accountHandler.ImportAccounts)
				
			// Export routes
			accounts.GET("/export/pdf", permMiddleware.CanExport("accounts"), accountHandler.ExportAccountsPDF)
			accounts.GET("/export/excel", permMiddleware.CanExport("accounts"), accountHandler.ExportAccountsExcel)
		}
		
// 📊 COA Display routes (V2 with proper balance display)
			coaDisplayServiceV2 := services.NewCOADisplayServiceV2(db)
			coaControllerV2 := controllers.NewCOAControllerV2(coaDisplayServiceV2)
			postedCOAController := controllers.NewCOAPostedController(db)
			
			coadisplay := protected.Group("/coa-display")
			{
				coadisplay.GET("", permMiddleware.CanView("accounts"), coaControllerV2.GetCOAWithDisplay)
				coadisplay.GET("/:id", permMiddleware.CanView("accounts"), coaControllerV2.GetCOAByID)
				coadisplay.GET("/by-type", permMiddleware.CanView("accounts"), coaControllerV2.GetCOABalancesByType)
				coadisplay.GET("/specific", permMiddleware.CanView("accounts"), coaControllerV2.GetSpecificAccounts)
				coadisplay.GET("/sales-related", permMiddleware.CanView("accounts"), coaControllerV2.GetSalesRelatedAccounts)
			}

			// 🔒 SSOT Posted-only COA endpoint for frontend
			coaPosted := protected.Group("/coa")
			{
				coaPosted.GET("/posted-balances", permMiddleware.CanView("accounts"), postedCOAController.GetPostedBalances)
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
			
// Initialize services for Sales V2
			coaService := services.NewCOAService(db)
			
			// Initialize Sales Journal Service V2 (clean implementation)
salesJournalServiceV2 := services.NewSalesJournalServiceV2(db, journalRepo, coaService)
			
			// Initialize Stock Service (can be nil if not available)
			stockService := services.NewStockService(db)
			
			// Initialize Settings Service for sales code generation
			settingsService := services.NewSettingsService(db)
			
			// Initialize Sales Service V2 (clean implementation with proper status-based journal posting)
			salesServiceV2 := services.NewSalesServiceV2(db, salesRepo, salesJournalServiceV2, stockService, notificationService, settingsService)

	// Initialize Payment repositories, services and controllers
	paymentRepo := repositories.NewPaymentRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	paymentService := services.NewPaymentService(db, paymentRepo, salesRepo, purchaseRepo, cashBankRepo, accountRepo, contactRepo)
	paymentController := controllers.NewPaymentController(paymentService)
	cashBankService := services.NewCashBankService(db, cashBankRepo, accountRepo)
	accountService := services.NewAccountService(accountRepo)
	cashBankController := controllers.NewCashBankController(cashBankService, accountService)

	// Initialize additional Sales Payment services required by SalesController
unifiedSalesPaymentService := services.NewUnifiedSalesPaymentService(db)
	
	// Initialize SalesController with new V2 Sales Service (inject pdfService)
	salesController := controllers.NewSalesController(salesServiceV2, paymentService, unifiedSalesPaymentService, pdfService)
	
	// Initialize PurchaseController with PaymentService integration (moved here after paymentService is available)
	purchaseController := controllers.NewPurchaseController(purchaseService, paymentService)

			// 🔔 Notification routes (accessible by all authenticated users)
			notifs := protected.Group("/notifications")
			{
				// Notification routes are generally accessible to all authenticated users since they're personal
				notifs.GET("", notificationHandler.GetNotifications)
				notifs.GET("/unread-count", notificationHandler.GetUnreadCount)
				notifs.PUT("/:id/read", notificationHandler.MarkNotificationAsRead)
				notifs.PUT("/read-all", notificationHandler.MarkAllNotificationsAsRead)
				notifs.GET("/type/:type", notificationHandler.GetNotificationsByType)
				notifs.GET("/approvals", notificationHandler.GetApprovalNotifications)
			}

			// Sales routes with permission checks
				sales := protected.Group("/sales")
			{
				// Basic CRUD operations
				sales.GET("", permMiddleware.CanView("sales"), salesController.GetSales)
				sales.GET("/:id", permMiddleware.CanView("sales"), salesController.GetSale)
				// Validate stock for sales create form
				sales.POST("/validate-stock", permMiddleware.CanCreate("sales"), salesController.ValidateSaleStock)
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
				
				// Integrated Payment Management routes
				sales.GET("/:id/for-payment", middleware.RoleRequired("admin", "finance", "director"), salesController.GetSaleForPayment)
				sales.POST("/:id/integrated-payment", middleware.RoleRequired("admin", "finance", "director"), salesController.CreateIntegratedPayment)

				// Returns management
				sales.POST("/:id/returns", middleware.RoleRequired("admin", "finance", "director"), salesController.CreateSaleReturn)
				sales.GET("/returns", middleware.RoleRequired("admin", "finance", "director"), salesController.GetSaleReturns)

				// Analytics and reporting
				sales.GET("/summary", middleware.RoleRequired("admin", "finance", "director", "employee"), salesController.GetSalesSummary)
				sales.GET("/analytics", middleware.RoleRequired("admin", "finance", "director"), salesController.GetSalesAnalytics)
				sales.GET("/receivables", middleware.RoleRequired("admin", "finance", "director"), salesController.GetReceivablesReport)

				// PDF exports
				sales.GET("/:id/invoice/pdf", permMiddleware.CanExport("sales"), salesController.ExportSaleInvoicePDF)
				sales.GET("/report/pdf", permMiddleware.CanExport("sales"), salesController.ExportSalesReportPDF)

				// Customer portal
				sales.GET("/customer/:customer_id", middleware.RoleRequired("admin", "finance", "director"), salesController.GetCustomerSales)
				sales.GET("/customer/:customer_id/invoices", middleware.RoleRequired("admin", "finance", "director"), salesController.GetCustomerInvoices)
			}

	// Initialize Balance Monitoring service and controller
	balanceMonitoringService := services.NewBalanceMonitoringService(db)
	balanceMonitoringController := controllers.NewBalanceMonitoringController(balanceMonitoringService)
	
	// Initialize API Usage Monitoring controller
	apiUsageController := controllers.NewAPIUsageController()
	
	// Initialize Performance Monitoring controller
	performanceController := controllers.NewPerformanceController(db)
			
			// ⚠️  DEPRECATED: Setup legacy Payment routes (including cash bank routes with GL fix functionality)
			// These routes may cause double posting - use SSOT routes instead
			// 🔒 PRODUCTION GUARD: Only enable legacy routes in development with explicit flag
			if os.Getenv("ENABLE_LEGACY_PAYMENT_ROUTES") == "true" && isDevelopmentMode() {
				log.Printf("⚠️ WARNING: Legacy payment routes enabled - may cause conflicts with SalesJournalServiceV2")
				SetupPaymentRoutes(protected, paymentController, cashBankController, cashBankService, jwtManager, db)
			} else {
				log.Printf("✅ Legacy payment routes disabled - using SalesJournalServiceV2 consistent flow only")
			}
			
			// ✅ NEW: Setup SSOT Payment routes with journal integration (prevents double posting)
			SetupSSOTPaymentRoutes(protected, db, jwtManager)

			// 📄 Export-only compatibility routes for Payments (safe, read-only)
			// These endpoints restore PDF/Excel exports without enabling legacy write routes
			paymentExports := protected.Group("/payments")
			{
				paymentExports.GET("/report/pdf", permMiddleware.CanExport("payments"), paymentController.ExportPaymentReportPDF)
				paymentExports.GET("/export/excel", permMiddleware.CanExport("payments"), paymentController.ExportPaymentReportExcel)
				paymentExports.GET("/:id/pdf", permMiddleware.CanExport("payments"), paymentController.ExportPaymentDetailPDF)
			}
			
			// ⚡ ULTRA-FAST: Setup Ultra-Fast Payment routes with minimal operations
			ultraFastRoutes := NewUltraFastPaymentRoutes(db)
			ultraFastRoutes.SetupUltraFastPaymentRoutes(r)
			
			// 🔄 Setup CashBank SSOT Integration routes (NEW - Phase 1 Implementation)
			// This provides unified view of CashBank data integrated with SSOT Journal system
			SetupCashBankIntegratedRoutes(protected, db, jwtManager)
			
			// 💰 Setup NEW Cash-Bank routes with SSOT integration
			SetupCashBankSSOTRoutes(v1, db, jwtManager)

			// 💰 Purchases routes with enhanced permission checks
			purchases := protected.Group("/purchases")
	purchases.Use(enhancedSecurity.RequestMonitoring()) // 📊 Enhanced monitoring
			{
				// Basic CRUD operations dengan enhanced security
				purchases.GET("", permMiddleware.CanView("purchases"), purchaseController.GetPurchases)
				// Approval statistics (must be defined before parameterized "/:id" route)
				purchases.GET("/approval-stats", permMiddleware.CanApprove("purchases"), purchaseApprovalHandler.GetApprovalStats)
				purchases.GET("/:id", permMiddleware.CanView("purchases"), purchaseController.GetPurchase)
				purchases.POST("", permMiddleware.CanCreate("purchases"), purchaseController.CreatePurchase)
				purchases.PUT("/:id", permMiddleware.CanEdit("purchases"), purchaseController.UpdatePurchase)
				purchases.DELETE("/:id", permMiddleware.CanDelete("purchases"), purchaseController.DeletePurchase)
				
				// Approval operations dengan permission checks
				purchases.POST("/:id/submit-approval", permMiddleware.CanCreate("purchases"), purchaseController.SubmitForApproval)
				purchases.POST("/:id/approve", permMiddleware.CanApprove("purchases"), purchaseController.ApprovePurchase)
				purchases.POST("/:id/reject", permMiddleware.CanApprove("purchases"), purchaseController.RejectPurchase)
				// Approval history endpoint (accessible by those who can view purchases)
				purchases.GET("/:id/approval-history", permMiddleware.CanView("purchases"), purchaseApprovalHandler.GetApprovalHistory)
				// Pending approvals (for those who can approve)
				purchases.GET("/pending-approval", permMiddleware.CanApprove("purchases"), purchaseApprovalHandler.GetPurchasesForApproval)
				
				// Document management dengan permission checks
				purchases.POST("/:id/documents", permMiddleware.CanEdit("purchases"), purchaseController.UploadDocument)
				purchases.GET("/:id/documents", permMiddleware.CanView("purchases"), purchaseController.GetPurchaseDocuments)
				purchases.DELETE("/documents/:document_id", permMiddleware.CanDelete("purchases"), purchaseController.DeleteDocument)
				
				// Receipt operations dengan permission checks
				purchases.POST("/receipts", permMiddleware.CanEdit("purchases"), purchaseController.CreatePurchaseReceipt)
				purchases.GET("/:id/receipts", permMiddleware.CanView("purchases"), purchaseController.GetPurchaseReceipts)
				
				// Receipt PDF exports dengan permission checks
				purchases.GET("/receipts/:receipt_id/pdf", permMiddleware.CanExport("purchases"), purchaseController.GetReceiptPDF)
				purchases.GET("/:id/receipts/pdf", permMiddleware.CanExport("purchases"), purchaseController.GetAllReceiptsPDF)
				
				// Analytics and reporting dengan permission checks
				purchases.GET("/summary", permMiddleware.CanView("purchases"), purchaseController.GetPurchasesSummary)
				purchases.GET("/pending-approvals", permMiddleware.CanApprove("purchases"), purchaseController.GetPendingApprovals)
				purchases.GET("/dashboard", permMiddleware.CanView("purchases"), purchaseController.GetPurchaseDashboard)
				purchases.GET("/vendor/:vendor_id/summary", permMiddleware.CanView("purchases"), purchaseController.GetVendorPurchaseSummary)
				
				// Payment management (similar to sales payment management)
				purchases.GET("/:id/payments", middleware.RoleRequired("admin", "finance", "director", "employee"), purchaseController.GetPurchasePayments)
				purchases.POST("/:id/payments", middleware.RoleRequired("admin", "finance", "director"), purchaseController.CreatePurchasePayment)
				
				// Integrated Payment Management routes  
				purchases.GET("/:id/for-payment", middleware.RoleRequired("admin", "finance", "director"), purchaseController.GetPurchaseForPayment)
				purchases.POST("/:id/integrated-payment", middleware.RoleRequired("admin", "finance", "director"), purchaseController.CreateIntegratedPayment)
				
				// Three-way matching dengan permission checks
				purchases.GET("/:id/matching", permMiddleware.CanView("purchases"), purchaseController.GetPurchaseMatching)
				purchases.POST("/:id/validate-matching", permMiddleware.CanApprove("purchases"), purchaseController.ValidateThreeWayMatching)
				
				// Journal entries integration dengan SSOT Journal System
				purchases.GET("/:id/journal-entries", permMiddleware.CanView("reports"), purchaseController.GetPurchaseJournalEntries)
			}

			// Expenses routes - REMOVED: No implementation yet

			// 🏢 Assets routes with enhanced permission checks dan audit logging
			assets := protected.Group("/assets")
			assets.Use(enhancedSecurity.RequestMonitoring()) // 📊 Enhanced monitoring
			{
				// Basic CRUD operations dengan enhanced security
				assets.GET("", permMiddleware.CanView("assets"), assetController.GetAssets)
				assets.GET("/:id", permMiddleware.CanView("assets"), assetController.GetAsset)
				assets.POST("", permMiddleware.CanCreate("assets"), assetController.CreateAsset)
				assets.PUT("/:id", permMiddleware.CanEdit("assets"), assetController.UpdateAsset)
				assets.DELETE("/:id", permMiddleware.CanDelete("assets"), assetController.DeleteAsset)
				assets.POST("/upload-image", permMiddleware.CanEdit("assets"), assetController.UploadAssetImage)
				
				// Manual capitalization endpoint
				assets.POST("/:id/capitalize", permMiddleware.CanEdit("assets"), assetController.CapitalizeAsset)
				
				// Asset categories management
				assets.GET("/categories", permMiddleware.CanView("assets"), assetController.GetAssetCategories)
				assets.POST("/categories", permMiddleware.CanCreate("assets"), assetController.CreateAssetCategory)
				
				// 📊 Reports and calculations dengan permission checks
				assets.GET("/summary", permMiddleware.CanView("assets"), assetController.GetAssetsSummary)
				assets.GET("/depreciation-report", permMiddleware.CanView("assets"), assetController.GetDepreciationReport)
				assets.GET("/:id/depreciation-schedule", permMiddleware.CanView("assets"), assetController.GetDepreciationSchedule)
				assets.GET("/:id/calculate-depreciation", permMiddleware.CanView("assets"), assetController.CalculateCurrentDepreciation)
				
				// Export routes - REMOVED: Not implemented yet
			}

		// Note: CashBank routes are already set up via SetupPaymentRoutes

			// 📦 Inventory routes with enhanced permission checks
			inventory := protected.Group("/inventory")
			{
				// Basic inventory operations - accessible by those who can view products
				inventory.GET("/movements", permMiddleware.CanView("products"), inventoryController.GetInventoryMovements)
				inventory.GET("/low-stock", permMiddleware.CanView("products"), inventoryController.GetLowStockProducts)
				inventory.GET("/valuation", permMiddleware.CanView("products"), inventoryController.GetStockValuation)
				inventory.GET("/report", permMiddleware.CanExport("products"), inventoryController.GetStockReport)
				inventory.POST("/bulk-price-update", permMiddleware.CanEdit("products"), inventoryController.BulkPriceUpdate)
			}

			// Approval workflows routes
			workflows := protected.Group("/approval-workflows")
			{
				workflows.GET("", purchaseApprovalHandler.GetApprovalWorkflows)
				workflows.POST("", middleware.RoleRequired("admin"), purchaseApprovalHandler.CreateApprovalWorkflow)
			}

			// ✅ CONSOLIDATED: Use only Enhanced Report Service as primary service with caching
			cacheService := services.NewReportCacheService()
			enhancedReportService := services.NewEnhancedReportService(db, accountRepo, salesRepo, purchaseRepo, productRepo, contactRepo, paymentRepo, cashBankRepo, cacheService)
			
			// 🔧 SIMPLIFIED: Use only Enhanced Report Service and Controller with PDF service
			enhancedReportController := controllers.NewEnhancedReportController(db)
			
			// Setup Settings routes
			SetupSettingsRoutes(protected, db)
			
			// ✅ CONSOLIDATED ROUTES: Use only Enhanced Report Routes - UNDER V1
			RegisterEnhancedReportRoutes(v1, enhancedReportController, jwtManager)

			// 🚀 SSOT REPORT INTEGRATION ROUTES: Single Source of Truth integration with all financial reports
			RegisterSSOTReportRoutesInMain(v1, db, unifiedJournalService, enhancedReportService, jwtManager)

			// ⚡ OPTIMIZED FINANCIAL REPORTS: Ultra-fast reports using materialized view
			SetupOptimizedReportsRoutes(r, db)
			
			// 🔧 COMPATIBILITY ROUTES: Add root-level aliases for SSOT reports
			// This provides backward compatibility for frontend requests to /ssot-reports/*
			// These routes redirect to the proper /api/v1/ssot-reports/* endpoints
			ssotAliasGroup := r.Group("/ssot-reports")
			ssotAliasGroup.Use(jwtManager.AuthRequired())
			ssotAliasGroup.Use(middleware.RoleRequired("finance", "admin", "director", "auditor"))
			{
				// Initialize SSOT controllers for direct access using existing services
				ssotAliasReportIntegrationService := services.NewSSOTReportIntegrationService(
					db,
					unifiedJournalService,
					enhancedReportService,
				)
				ssotAliasReportController := controllers.NewSSOTReportIntegrationController(ssotAliasReportIntegrationService, db)
				
				// Route aliases that mirror the v1 endpoints
				ssotAliasGroup.GET("/trial-balance", ssotAliasReportController.GetSSOTTrialBalance)
				ssotAliasGroup.GET("/general-ledger", ssotAliasReportController.GetSSOTGeneralLedger)
				ssotAliasGroup.GET("/journal-analysis", ssotAliasReportController.GetSSOTJournalAnalysis)
				
				// Purchase report alias (already working, but add for consistency)
				purchaseReportController := controllers.NewSSOTPurchaseReportController(db)
				ssotAliasGroup.GET("/purchase-report", purchaseReportController.GetPurchaseReport)
				
				// Info endpoint explaining the alias routes
				ssotAliasGroup.GET("/info", func(c *gin.Context) {
					c.JSON(200, gin.H{
						"status":  "success",
						"message": "SSOT Reports Compatibility Routes",
						"note":    "These are alias routes for backward compatibility",
						"recommendation": "Use /api/v1/ssot-reports/* for new implementations",
						"available_endpoints": []string{
							"/ssot-reports/trial-balance",
							"/ssot-reports/general-ledger", 
							"/ssot-reports/journal-analysis",
							"/ssot-reports/purchase-report",
						},
						"proper_api_endpoints": []string{
							"/api/v1/ssot-reports/trial-balance",
							"/api/v1/ssot-reports/general-ledger",
							"/api/v1/ssot-reports/journal-analysis", 
							"/api/v1/ssot-reports/purchase-report",
						},
					})
				})
			}

			// 📊 SSOT Profit & Loss Controller - Direct P&L endpoint for frontend
			ssotPLController := controllers.NewSSOTProfitLossController(db)
			
			// SSOT Report routes for frontend integration
			ssotReports := v1.Group("/reports")
			ssotReports.Use(jwtManager.AuthRequired())
			ssotReports.Use(middleware.RoleRequired("finance", "admin", "director"))
			{
				// Main P&L endpoint for frontend - matches the format expected by EnhancedProfitLossModal
				ssotReports.GET("/ssot-profit-loss", ssotPLController.GetSSOTProfitLoss)
			}

			// 📊 SSOT Balance Sheet Controller - Direct Balance Sheet endpoint for frontend
			ssotBSController := controllers.NewSSOTBalanceSheetController(db)
			
			// 💰 SSOT Cash Flow Controller - Direct Cash Flow endpoint for frontend
			ssotCFController := controllers.NewSSOTCashFlowController(db)
			
			// Balance Sheet Report routes for frontend integration
			ssotBSReports := v1.Group("/reports/ssot")
			ssotBSReports.Use(jwtManager.AuthRequired())
			ssotBSReports.Use(middleware.RoleRequired("finance", "admin", "director"))
			{
				// Main Balance Sheet endpoint for frontend - matches the format expected by SSOT Balance Sheet Modal
				ssotBSReports.GET("/balance-sheet", ssotBSController.GenerateSSOTBalanceSheet)
				ssotBSReports.GET("/balance-sheet/account-details", ssotBSController.GetSSOTBalanceSheetAccountDetails)
				ssotBSReports.GET("/balance-sheet/validate", ssotBSController.ValidateSSOTBalanceSheet)
				ssotBSReports.GET("/balance-sheet/comparison", ssotBSController.GetSSOTBalanceSheetComparison)
				
				// 💰 Cash Flow Statement endpoints for frontend - matches the format expected by SSOT Cash Flow Modal
				ssotBSReports.GET("/cash-flow", ssotCFController.GetSSOTCashFlow)
				ssotBSReports.GET("/cash-flow/summary", ssotCFController.GetSSOTCashFlowSummary)
				ssotBSReports.GET("/cash-flow/validate", ssotCFController.ValidateSSOTCashFlow)
			}

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
				
				// 📊 API Usage monitoring routes
				monitoring.GET("/api-usage/stats", apiUsageController.GetAPIUsageStats)
				monitoring.GET("/api-usage/top", apiUsageController.GetTopEndpoints)
				monitoring.GET("/api-usage/unused", apiUsageController.GetUnusedEndpoints)
				monitoring.GET("/api-usage/analytics", apiUsageController.GetUsageAnalytics)
				monitoring.POST("/api-usage/reset", apiUsageController.ResetUsageStats)
				
				// 🏁 Performance monitoring routes (admin only)
				monitoring.GET("/performance/report", performanceController.GetPerformanceReport)
				monitoring.GET("/performance/metrics", performanceController.GetQuickMetrics)
				monitoring.GET("/performance/bottlenecks", performanceController.GetBottlenecks)
				monitoring.GET("/performance/recommendations", performanceController.GetRecommendations)
				monitoring.GET("/performance/system", performanceController.GetSystemStatus)
				monitoring.POST("/performance/metrics/clear", performanceController.ClearMetrics)
				monitoring.GET("/performance/test", performanceController.TestUltraFastEndpoint)
				
				// 🚑 Payment timeout diagnostic routes (critical for troubleshooting)
				monitoring.GET("/timeout/diagnostics", performanceController.RunTimeoutDiagnostics)
				monitoring.GET("/timeout/health", performanceController.GetQuickHealthCheck)
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


	// Swagger documentation endpoint (accessible in development or when ENABLE_SWAGGER=true)
	if isDevelopmentMode() || os.Getenv("ENABLE_SWAGGER") == "true" {
		// Serve filtered swagger.json from a neutral path to avoid wildcard conflicts
		r.StaticFile("/openapi/doc.json", "./docs/swagger.json")
		
		// Swagger UI endpoints pointing to the filtered doc.json
		v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi/doc.json")))
		v1.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi/doc.json")))
		
		// ROOT-LEVEL SWAGGER ROUTES: Add root-level routes for browser compatibility
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi/doc.json")))
		r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi/doc.json")))
	}

	// Debug routes for development only
	if gin.Mode() == gin.DebugMode {
		debug := v1.Group("/debug")
		{
			// Minimal debug routes for development
			debugWithAuth := debug.Group("/auth")
			debugWithAuth.Use(jwtManager.AuthRequired())
			{
				debugWithAuth.GET("/context", debugController.TestJWTContext)
				debugWithAuth.GET("/role", debugController.TestRolePermission)
				
				// Essential permission tests only
				debugWithAuth.GET("/test-cashbank-permission", permMiddleware.CanView("cash_bank"), debugController.TestCashBankPermission)
				debugWithAuth.GET("/test-payments-permission", permMiddleware.CanView("payments"), debugController.TestPaymentsPermission)
			}
		}
	}
}
