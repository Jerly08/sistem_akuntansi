package routes

import (
	"app-sistem-akuntansi/handlers"
	"app-sistem-akuntansi/middleware"
	"github.com/gin-gonic/gin"
)

// SetupAccountRoutes sets up all account-related routes
func SetupAccountRoutes(router *gin.Engine, accountHandler *handlers.AccountHandler, jwtManager *middleware.JWTManager) {
	// Account routes group with authentication
	accounts := router.Group("/api/accounts")
	accounts.Use(jwtManager.AuthRequired())
	
	{
		// List accounts - accessible by ADMIN, FINANCE
		accounts.GET("", middleware.RoleRequired("admin", "finance"), accountHandler.ListAccounts)
		
		// Get account catalog (minimal EXPENSE data) - accessible by EMPLOYEE for purchases
		accounts.GET("/catalog", middleware.RoleRequired("employee", "admin", "finance"), accountHandler.GetAccountCatalog)
		
		// Get account hierarchy - accessible by ADMIN, FINANCE
		accounts.GET("/hierarchy", middleware.RoleRequired("admin", "finance"), accountHandler.GetAccountHierarchy)
		
		// Get balance summary - accessible by ADMIN, FINANCE
		accounts.GET("/balance-summary", middleware.RoleRequired("admin", "finance"), accountHandler.GetBalanceSummary)
		
		// Get single account - accessible by ADMIN, FINANCE
		accounts.GET("/:code", middleware.RoleRequired("admin", "finance"), accountHandler.GetAccount)
		
		// Create account - accessible by ADMIN, FINANCE
		accounts.POST("", middleware.RoleRequired("admin", "finance"), accountHandler.CreateAccount)
		
		// Update account - accessible by ADMIN, FINANCE
		accounts.PUT("/:code", middleware.RoleRequired("admin", "finance"), accountHandler.UpdateAccount)
		
		// Delete account - accessible by ADMIN only
		accounts.DELETE("/:code", middleware.RoleRequired("admin"), accountHandler.DeleteAccount)
		
		// Bulk import accounts - accessible by ADMIN only
		accounts.POST("/import", middleware.RoleRequired("admin"), accountHandler.ImportAccounts)
	}
}
