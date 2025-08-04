package routes

import (
	"net/http"
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// Controllers
	authController := controllers.NewAuthController(db)
	productController := controllers.NewProductController(db)
	
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
				assets.GET("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Assets endpoint - coming soon"})
				})
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
				inventory.GET("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Inventory endpoint - coming soon"})
				})
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

	// Health check endpoint
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
