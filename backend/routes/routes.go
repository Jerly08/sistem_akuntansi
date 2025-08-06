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
				products.GET("", productController.GetProducts)
				products.GET("/:id", productController.GetProduct)
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
				categories.GET("", categoryController.GetCategories)
				categories.GET("/tree", categoryController.GetCategoryTree)
				categories.GET("/:id", categoryController.GetCategory)
				categories.GET("/:id/products", categoryController.GetCategoryProducts)
				categories.POST("", middleware.RoleRequired("admin", "inventory_manager"), categoryController.CreateCategory)
				categories.PUT("/:id", middleware.RoleRequired("admin", "inventory_manager"), categoryController.UpdateCategory)
				categories.DELETE("/:id", middleware.RoleRequired("admin"), categoryController.DeleteCategory)
			}

			// Account routes (Chart of Accounts)
			accounts := protected.Group("/accounts")
			{
				accounts.GET("", middleware.RoleRequired("admin", "finance"), accountHandler.ListAccounts)
				accounts.GET("/hierarchy", middleware.RoleRequired("admin", "finance"), accountHandler.GetAccountHierarchy)
				accounts.GET("/balance-summary", middleware.RoleRequired("admin", "finance"), accountHandler.GetBalanceSummary)
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
				contacts.GET("", middleware.RoleRequired("admin", "finance", "inventory_manager"), contactController.GetContacts)
				contacts.GET("/:id", middleware.RoleRequired("admin", "finance", "inventory_manager"), contactController.GetContact)
				contacts.POST("", middleware.RoleRequired("admin", "finance", "inventory_manager"), contactController.CreateContact)
				contacts.PUT("/:id", middleware.RoleRequired("admin", "finance", "inventory_manager"), contactController.UpdateContact)
				contacts.DELETE("/:id", middleware.RoleRequired("admin"), contactController.DeleteContact)
				
				// Advanced operations
				contacts.GET("/type/:type", middleware.RoleRequired("admin", "finance", "inventory_manager"), contactController.GetContactsByType)
				contacts.GET("/search", middleware.RoleRequired("admin", "finance", "inventory_manager"), contactController.SearchContacts)
				
				// Import/Export operations
				contacts.POST("/import", middleware.RoleRequired("admin"), contactController.ImportContacts)
				contacts.GET("/export", middleware.RoleRequired("admin", "finance", "inventory_manager"), contactController.ExportContacts)
			}

			// Sales routes
			sales := protected.Group("/sales")
			{
				sales.GET("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Sales endpoint - coming soon"})
				})
			}

			// Purchases routes
			purchases := protected.Group("/purchases")
			{
				purchases.GET("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Purchases endpoint - coming soon"})
				})
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
				assets.GET("", middleware.RoleRequired("admin", "finance"), assetController.GetAssets)
				assets.GET("/:id", middleware.RoleRequired("admin", "finance"), assetController.GetAsset)
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
				cashBank.GET("", func(c *gin.Context) {
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
