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

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// Controllers
	authController := controllers.NewAuthController(db)
	productController := controllers.NewProductController(db)
	categoryController := controllers.NewCategoryController(db)
	inventoryController := controllers.NewInventoryController(db)
	assetController := controllers.NewAssetController(db)
	
	// Initialize repositories, services and handlers
	accountRepo := repositories.NewAccountRepository(db)
	exportService := services.NewExportService(accountRepo)
	accountHandler := handlers.NewAccountHandler(accountRepo, exportService)
	
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
	
	// Initialize JWT Manager
	jwtManager := middleware.NewJWTManager(db)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes (no auth required)
		auth := v1.Group("/auth")
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
				products.POST("", middleware.RoleRequired("admin", "inventory_manager"), productController.CreateProduct)
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

			// Account routes (Chart of Accounts)
			accounts := protected.Group("/accounts")
			{
				accounts.GET("", middleware.RoleRequired("admin", "finance"), accountHandler.ListAccounts)
				
				// Get account catalog (minimal EXPENSE data) - accessible by EMPLOYEE for purchases
				accounts.GET("/catalog", middleware.RoleRequired("employee", "admin", "finance", "director"), accountHandler.GetAccountCatalog)
				
				accounts.GET("/hierarchy", middleware.RoleRequired("admin", "finance"), accountHandler.GetAccountHierarchy)
				accounts.GET("/balance-summary", middleware.RoleRequired("admin", "finance"), accountHandler.GetBalanceSummary)
				accounts.GET("/validate-code", middleware.RoleRequired("admin", "finance"), accountHandler.ValidateAccountCode)
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
				contacts.POST("", middleware.RoleRequired("admin", "finance", "inventory_manager"), contactController.CreateContact)
				contacts.PUT("/:id", middleware.RoleRequired("admin", "finance", "inventory_manager"), contactController.UpdateContact)
				contacts.DELETE("/:id", middleware.RoleRequired("admin"), contactController.DeleteContact)
				
				// Advanced operations
				contacts.GET("/type/:type", middleware.RoleRequired("admin", "finance", "inventory_manager", "employee", "director"), contactController.GetContactsByType)
				contacts.GET("/search", middleware.RoleRequired("admin", "finance", "inventory_manager", "employee", "director"), contactController.SearchContacts)
				
				// Import/Export operations
				contacts.POST("/import", middleware.RoleRequired("admin"), contactController.ImportContacts)
				contacts.GET("/export", middleware.RoleRequired("admin", "finance", "inventory_manager"), contactController.ExportContacts)
			}

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
				sales.GET("", middleware.RoleRequired("admin", "finance", "director", "employee"), func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Sales endpoint - coming soon"})
				})
			}

			// Payments routes
			payments := protected.Group("/payments")
			{
				payments.GET("", middleware.RoleRequired("admin", "finance", "director"), func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Payments endpoint - coming soon"})
				})
			}

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
				
				// Reports and calculations
				assets.GET("/summary", middleware.RoleRequired("admin", "finance", "director"), assetController.GetAssetsSummary)
				assets.GET("/depreciation-report", middleware.RoleRequired("admin", "finance", "director"), assetController.GetDepreciationReport)
				assets.GET("/:id/depreciation-schedule", middleware.RoleRequired("admin", "finance"), assetController.GetDepreciationSchedule)
				assets.GET("/:id/calculate-depreciation", middleware.RoleRequired("admin", "finance"), assetController.CalculateCurrentDepreciation)
			}

			// Cash Bank routes
			cashBank := protected.Group("/cash-bank")
			{
				cashBank.GET("", middleware.RoleRequired("admin", "finance", "director"), func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Cash Bank endpoint - coming soon"})
				})
			}

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

			// Reports routes
			reports := protected.Group("/reports")
			{
				reports.GET("", middleware.RoleRequired("admin", "director", "finance"), func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Reports endpoint - coming soon"})
				})
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
}
