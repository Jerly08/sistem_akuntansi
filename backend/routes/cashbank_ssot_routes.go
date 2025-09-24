package routes

import (
	"app-sistem-akuntansi/handlers"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupCashBankSSOTRoutes sets up all cash-bank routes with SSOT integration
func SetupCashBankSSOTRoutes(v1 *gin.RouterGroup, db *gorm.DB, jwtManager *middleware.JWTManager) {
	// Initialize repositories
	accountRepo := repositories.NewAccountRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	
	// Initialize services  
	cashBankService := services.NewCashBankService(db, cashBankRepo, accountRepo)
	
	// Initialize handler
	cashBankHandler := handlers.NewCashBankHandler(cashBankService)
	
	// Initialize Permission Middleware
	permMiddleware := middleware.NewPermissionMiddleware(db)
	
	// Cash-Bank routes with SSOT integration
	cashBankGroup := v1.Group("/cash-bank")
	cashBankGroup.Use(jwtManager.AuthRequired())
	{
		// Account Management
		accounts := cashBankGroup.Group("/accounts")
		{
			accounts.GET("", permMiddleware.CanView("cash_bank"), cashBankHandler.GetCashBankAccounts)
			accounts.GET("/:id", permMiddleware.CanView("cash_bank"), cashBankHandler.GetCashBankByID)
			accounts.POST("", permMiddleware.CanCreate("cash_bank"), cashBankHandler.CreateCashBankAccount)
			accounts.PUT("/:id", permMiddleware.CanEdit("cash_bank"), cashBankHandler.UpdateCashBankAccount)
			accounts.DELETE("/:id", permMiddleware.CanDelete("cash_bank"), cashBankHandler.DeleteCashBankAccount)
			
			// Transaction History
			accounts.GET("/:id/transactions", permMiddleware.CanView("cash_bank"), cashBankHandler.GetTransactions)
			
			// Bank Reconciliation
			accounts.POST("/:id/reconcile", permMiddleware.CanEdit("cash_bank"), cashBankHandler.ReconcileAccount)
		}
		
		// Transaction Processing (all with SSOT journal integration)
		transactions := cashBankGroup.Group("/transactions")
		{
			transactions.POST("/deposit", permMiddleware.CanCreate("cash_bank"), cashBankHandler.ProcessDeposit)
			transactions.POST("/withdrawal", permMiddleware.CanCreate("cash_bank"), cashBankHandler.ProcessWithdrawal)
			transactions.POST("/transfer", permMiddleware.CanCreate("cash_bank"), cashBankHandler.ProcessTransfer)
		}
		
		// Reporting and Summary
		reports := cashBankGroup.Group("/reports")
		{
			reports.GET("/balance-summary", permMiddleware.CanView("cash_bank"), cashBankHandler.GetBalanceSummary)
			reports.GET("/payment-accounts", permMiddleware.CanView("cash_bank"), cashBankHandler.GetPaymentAccounts)
		}
		
		// SSOT Integration and Validation
		ssot := cashBankGroup.Group("/ssot")
		{
			ssot.GET("/journals", permMiddleware.CanView("reports"), cashBankHandler.GetSSOTJournalEntries)
			ssot.POST("/validate-integrity", permMiddleware.CanView("reports"), cashBankHandler.ValidateIntegrity)
		}
	}
}