package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupCashBankRoutes sets up all cash bank related routes
func SetupCashBankRoutes(router *gin.Engine, db *gorm.DB, jwtManager *middleware.JWTManager) {
	// Initialize repositories and services
	accountRepo := repositories.NewAccountRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	cashBankService := services.NewCashBankService(db, cashBankRepo, accountRepo)
	cashBankController := controllers.NewCashBankController(cashBankService)
	fixCashBankController := controllers.NewFixCashBankController(db, cashBankService)

	// Cash Bank routes group with authentication
	cashBank := router.Group("/api/v1/cashbank")
	cashBank.Use(jwtManager.AuthRequired())
	
	{
		// Basic CRUD operations
		cashBank.GET("/accounts", middleware.RoleRequired("admin", "finance", "director", "employee"), cashBankController.GetAccounts)
		
		// Payment accounts endpoint - specifically for payment form dropdowns
		cashBank.GET("/payment-accounts", middleware.RoleRequired("admin", "finance", "director", "employee"), cashBankController.GetPaymentAccounts)
		
		cashBank.GET("/accounts/:id", middleware.RoleRequired("admin", "finance", "director"), cashBankController.GetAccountByID)
		cashBank.POST("/accounts", middleware.RoleRequired("admin", "finance"), cashBankController.CreateAccount)
		cashBank.PUT("/accounts/:id", middleware.RoleRequired("admin", "finance"), cashBankController.UpdateAccount)
		
		// Transaction operations
		cashBank.POST("/transfer", middleware.RoleRequired("admin", "finance", "director"), cashBankController.ProcessTransfer)
		cashBank.POST("/deposit", middleware.RoleRequired("admin", "finance", "director"), cashBankController.ProcessDeposit)
		cashBank.POST("/withdrawal", middleware.RoleRequired("admin", "finance", "director"), cashBankController.ProcessWithdrawal)
		
		// Reports and analytics
		cashBank.GET("/accounts/:id/transactions", middleware.RoleRequired("admin", "finance", "director"), cashBankController.GetTransactions)
		cashBank.GET("/balance-summary", middleware.RoleRequired("admin", "finance", "director"), cashBankController.GetBalanceSummary)
		cashBank.POST("/accounts/:id/reconcile", middleware.RoleRequired("admin", "finance"), cashBankController.ReconcileAccount)
		
		// Admin operations - GL Account linking fixes
		cashBank.GET("/admin/check-gl-links", middleware.RoleRequired("admin"), fixCashBankController.CheckCashBankGLLinks)
		cashBank.POST("/admin/fix-gl-links", middleware.RoleRequired("admin"), fixCashBankController.FixCashBankGLLinks)
	}
}
