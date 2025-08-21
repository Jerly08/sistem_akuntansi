package routes

import (
	"net/http"
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/handlers"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, startupService *services.StartupService) {
	// Controllers
	authController := controllers.NewAuthController(db)
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
	notificationService := services.NewNotificationService(notificationRepo)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	
	// Purchase repositories, services and controllers
	purchaseRepo := repositories.NewPurchaseRepository(db)
	productRepo := repositories.NewProductRepository(db)
	approvalService := services.NewApprovalService(db)
	purchaseService := services.NewPurchaseService(
		purchaseRepo,
		productRepo, 
		contactRepo,
		accountRepo,
		approvalService,
		nil, // journal service - can be nil for now
		nil, // pdf service - can be nil for now
	)
	purchaseController := controllers.NewPurchaseController(purchaseService)
	// Handlers that depend on services
	purchaseApprovalHandler := handlers.NewPurchaseApprovalHandler(purchaseService, approvalService)
	
	// Initialize security middleware
	middleware.InitAuditLogger(db)       // Initialize audit logging
	middleware.InitTokenMonitor(db)      // Initialize token monitoring
	
	// Initialize JWT Manager
	jwtManager := middleware.NewJWTManager(db)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes (no auth required)
		auth := v1.Group("/auth")
		auth.Use(middleware.AuthRateLimit()) // Apply auth rate limiting
		{
			auth.POST("/login", authController.Login)
			auth.POST("/register", authController.Register)
			auth.POST("/refresh", authController.RefreshToken)
		}

		// Protected routes (auth required)
		protected := v1.Group("")
		protected.Use(jwtManager.AuthRequired())
		{
			// Profile routes
			protected.GET("/profile", authController.Profile)

			// Product routes
			products := protected.Group("/products")
			{
				products.GET("", middleware.RoleRequired("admin", "inventory_manager", "employee", "director"), productController.GetProducts)
				products.GET("/:id", middleware.RoleRequired("admin", "inventory_manager", "employee", "director"), productController.GetProduct)
products.POST("", middleware.RoleRequired("admin", "inventory_manager", "employee"), productController.CreateProduct)
				products.PUT("/:id", middleware.RoleRequired("admin", "inventory_manager"), productController.UpdateProduct)
				products.DELETE("/:id", middleware.RoleRequired("admin"), productController.DeleteProduct)
				products.POST("/adjust-stock", middleware.RoleRequired("admin", "inventory_manager"), productController.AdjustStock)
				products.POST("/opname", middleware.RoleRequired("admin", "inventory_manager"), productController.Opname)
				products.POST("/upload-image", middleware.RoleRequired("admin", "inventory_manager"), productController.UploadProductImage)
			}

			// Category routes
			categories := protected.Group("/categories")
			{
				categories.GET("", middleware.RoleRequired("admin", "inventory_manager", "employee", "director"), categoryController.GetCategories)
				categories.GET("/tree", middleware.RoleRequired("admin", "inventory_manager", "employee", "director"), categoryController.GetCategoryTree)
				categories.GET("/:id", middleware.RoleRequired("admin", "inventory_manager", "employee", "director"), categoryController.GetCategory)
				categories.GET("/:id/products", middleware.RoleRequired("admin", "inventory_manager", "employee", "director"), categoryController.GetCategoryProducts)
				categories.POST("", middleware.RoleRequired("admin", "inventory_manager"), categoryController.CreateCategory)
				categories.PUT("/:id", middleware.RoleRequired("admin", "inventory_manager"), categoryController.UpdateCategory)
				categories.DELETE("/:id", middleware.RoleRequired("admin"), categoryController.DeleteCategory)
			}

			// Product Units routes
			units := protected.Group("/product-units")
			{
				units.GET("", middleware.RoleRequired("admin", "inventory_manager", "employee", "director"), unitController.GetProductUnits)
				units.GET("/:id", middleware.RoleRequired("admin", "inventory_manager", "employee", "director"), unitController.GetProductUnit)
				units.POST("", middleware.RoleRequired("admin"), unitController.CreateProductUnit)
				units.PUT("/:id", middleware.RoleRequired("admin"), unitController.UpdateProductUnit)
				units.DELETE("/:id", middleware.RoleRequired("admin"), unitController.DeleteProductUnit)
			}

			// Account routes (Chart of Accounts)
			accounts := protected.Group("/accounts")
			{
				accounts.GET("", middleware.RoleRequired("admin", "finance"), accountHandler.ListAccounts)
				
				// Get account catalog (minimal EXPENSE data) - accessible by EMPLOYEE for purchases
				accounts.GET("/catalog", middleware.RoleRequired("employee", "admin", "finance", "director"), accountHandler.GetAccountCatalog)
				
				accounts.GET("/hierarchy", middleware.RoleRequired("admin", "finance"), accountHandler.GetAccountHierarchy)
				accounts.GET("/balance-summary", middleware.RoleRequired("admin", "finance"), accountHandler.GetBalanceSummary)
				accounts.GET("/validate-code", middleware.RoleRequired("admin", "finance"), accountHandler.ValidateAccountCode)
				
				// Fix account header status
				accounts.POST("/fix-header-status", middleware.RoleRequired("admin"), accountHandler.FixAccountHeaderStatus)
				
				accounts.GET("/:code", middleware.RoleRequired("admin", "finance"), accountHandler.GetAccount)
				accounts.POST("", middleware.RoleRequired("admin", "finance"), accountHandler.CreateAccount)
				accounts.PUT("/:code", middleware.RoleRequired("admin", "finance"), accountHandler.UpdateAccount)
				accounts.DELETE("/:code", middleware.RoleRequired("admin"), accountHandler.DeleteAccount)
				accounts.POST("/import", middleware.RoleRequired("admin"), accountHandler.ImportAccounts)
				
				// Export routes
				accounts.GET("/export/pdf", middleware.RoleRequired("admin", "finance"), accountHandler.ExportAccountsPDF)
				accounts.GET("/export/excel", middleware.RoleRequired("admin", "finance"), accountHandler.ExportAccountsExcel)
			}

			// Contact routes
			contacts := protected.Group("/contacts")
			{
				// Basic CRUD operations
				contacts.GET("", middleware.RoleRequired("admin", "finance", "inventory_manager", "employee", "director"), contactController.GetContacts)
				contacts.GET("/:id", middleware.RoleRequired("admin", "finance", "inventory_manager", "employee", "director"), contactController.GetContact)
contacts.POST("", middleware.RoleRequired("admin", "finance", "inventory_manager", "employee"), contactController.CreateContact)
				contacts.PUT("/:id", middleware.RoleRequired("admin", "finance", "inventory_manager"), contactController.UpdateContact)
				contacts.DELETE("/:id", middleware.RoleRequired("admin"), contactController.DeleteContact)
				
				// Advanced operations
				contacts.GET("/type/:type", middleware.RoleRequired("admin", "finance", "inventory_manager", "employee", "director"), contactController.GetContactsByType)
				contacts.GET("/search", middleware.RoleRequired("admin", "finance", "inventory_manager", "employee", "director"), contactController.SearchContacts)
				
				// Import/Export operations
				contacts.POST("/import", middleware.RoleRequired("admin"), contactController.ImportContacts)
				contacts.GET("/export", middleware.RoleRequired("admin", "finance", "inventory_manager"), contactController.ExportContacts)
			}

			// Sales repositories, services and controllers
			salesRepo := repositories.NewSalesRepository(db)
			productRepo := repositories.NewProductRepository(db)
			pdfService := services.NewPDFService()
		salesService := services.NewSalesService(salesRepo, productRepo, contactRepo, accountRepo, nil, pdfService)
			salesController := controllers.NewSalesController(salesService)

			// Notification routes
			notifs := protected.Group("/notifications")
			{
				notifs.GET("", notificationHandler.GetNotifications)
				notifs.GET("/unread-count", notificationHandler.GetUnreadCount)
				notifs.PUT("/:id/read", notificationHandler.MarkNotificationAsRead)
				notifs.PUT("/read-all", notificationHandler.MarkAllNotificationsAsRead)
				notifs.GET("/type/:type", notificationHandler.GetNotificationsByType)
				notifs.GET("/approvals", notificationHandler.GetApprovalNotifications)
			}

			// Sales routes
			sales := protected.Group("/sales")
			{
				// Basic CRUD operations
				sales.GET("", middleware.RoleRequired("admin", "finance", "director", "employee", "inventory_manager"), salesController.GetSales)
				sales.GET("/:id", middleware.RoleRequired("admin", "finance", "director", "employee", "inventory_manager"), salesController.GetSale)
				sales.POST("", middleware.RoleRequired("admin", "finance", "director"), salesController.CreateSale)
				sales.PUT("/:id", middleware.RoleRequired("admin", "finance", "director"), salesController.UpdateSale)
				sales.DELETE("/:id", middleware.RoleRequired("admin"), salesController.DeleteSale)

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
				sales.GET("/summary", middleware.RoleRequired("admin", "finance", "director"), salesController.GetSalesSummary)
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
			
			// Setup Payment routes
			SetupPaymentRoutes(protected, paymentController, cashBankController, cashBankService, jwtManager)

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
				purchases.GET("/summary", middleware.RoleRequired("admin", "finance", "director"), purchaseController.GetPurchasesSummary)
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

			// Initialize Report service and controller
			reportService := services.NewReportService(db, accountRepo, salesRepo, purchaseRepo, productRepo, contactRepo, paymentRepo, cashBankRepo)
			reportController := controllers.NewReportController(reportService)
			
			// Setup Report routes
			SetupReportRoutes(protected, reportController)

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
