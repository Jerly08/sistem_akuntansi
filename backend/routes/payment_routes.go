package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupPaymentRoutes(router *gin.RouterGroup, paymentController *controllers.PaymentController, cashBankController *controllers.CashBankController, cashBankService *services.CashBankService, jwtManager *middleware.JWTManager, db *gorm.DB) {
	// Initialize fix controller for GL account linking
	fixCashBankController := controllers.NewFixCashBankController(db, cashBankService)
	// Initialize permission middleware
	permissionMiddleware := middleware.NewPermissionMiddleware(db)
	
	// Initialize CashBank validation middleware and services for Phase 1 sync
	accountingService := services.NewCashBankAccountingService(db)
	validationService := services.NewCashBankValidationService(db, accountingService)
	// Get repositories for enhanced service
	cashBankRepo := repositories.NewCashBankRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	enhancedCashBankService := services.NewCashBankEnhancedService(db, cashBankRepo, accountRepo)
	cashBankValidationMiddleware := middleware.NewCashBankValidationMiddleware(validationService, enhancedCashBankService)
	
	// Payment routes
	payment := router.Group("/payments")
	payment.Use(middleware.PaymentRateLimit()) // Apply rate limiting to all payment endpoints
	if middleware.GlobalAuditLogger != nil {
		payment.Use(middleware.GlobalAuditLogger.PaymentAuditMiddleware()) // Apply audit logging
	}
	{
		// Payment routes with permission-based restrictions
		payment.GET("", permissionMiddleware.CanView("payments"), paymentController.GetPayments)
		payment.GET("/:id", permissionMiddleware.CanView("payments"), paymentController.GetPaymentByID)
		payment.POST("/receivable", permissionMiddleware.CanCreate("payments"), paymentController.CreateReceivablePayment)
		payment.POST("/payable", permissionMiddleware.CanCreate("payments"), paymentController.CreatePayablePayment)
		payment.POST("/:id/cancel", permissionMiddleware.CanEdit("payments"), paymentController.CancelPayment)
		payment.DELETE("/:id", middleware.RoleRequired("admin"), paymentController.DeletePayment)
		payment.GET("/unpaid-invoices/:customer_id", permissionMiddleware.CanView("payments"), paymentController.GetUnpaidInvoices)
		payment.GET("/unpaid-bills/:vendor_id", permissionMiddleware.CanView("payments"), paymentController.GetUnpaidBills)
		payment.GET("/summary", permissionMiddleware.CanView("payments"), paymentController.GetPaymentSummary)
		payment.GET("/analytics", permissionMiddleware.CanView("payments"), paymentController.GetPaymentAnalytics)
		
		// Sales integration routes
		payment.POST("/sales", permissionMiddleware.CanCreate("payments"), paymentController.CreateSalesPayment)
		payment.GET("/sales/unpaid-invoices/:customer_id", permissionMiddleware.CanView("payments"), paymentController.GetSalesUnpaidInvoices)
		
		// Debug routes (admin only)
		payment.GET("/debug/recent", middleware.RoleRequired("admin"), paymentController.GetRecentPayments)
		
		// Export routes
		payment.GET("/report/pdf", permissionMiddleware.CanExport("payments"), paymentController.ExportPaymentReportPDF)
		payment.GET("/export/excel", permissionMiddleware.CanExport("payments"), paymentController.ExportPaymentReportExcel)
		payment.GET("/:id/pdf", permissionMiddleware.CanExport("payments"), paymentController.ExportPaymentDetailPDF)
	}
	
	// Cash & Bank routes
	cashbank := router.Group("/cashbank")
	{
		// Account management
		cashbank.GET("/accounts", permissionMiddleware.CanView("cash_bank"), cashBankController.GetAccounts)
		
		// Payment accounts endpoint - specifically for payment form dropdowns
		cashbank.GET("/payment-accounts", permissionMiddleware.CanView("cash_bank"), cashBankController.GetPaymentAccounts)
		
		// Revenue accounts endpoint - for deposit form source account dropdown
		cashbank.GET("/revenue-accounts", permissionMiddleware.CanView("cash_bank"), cashBankController.GetRevenueAccounts)
		
		cashbank.GET("/accounts/:id", permissionMiddleware.CanView("cash_bank"), cashBankController.GetAccountByID)
cashbank.POST("/accounts", permissionMiddleware.CanCreate("cash_bank"), cashBankController.CreateAccount)
cashbank.PUT("/accounts/:id", permissionMiddleware.CanEdit("cash_bank"), cashBankController.UpdateAccount)
cashbank.DELETE("/accounts/:id", permissionMiddleware.CanDelete("cash_bank"), cashBankController.DeleteAccount)
		
		// Transactions
		cashbank.POST("/transfer", permissionMiddleware.CanCreate("cash_bank"), cashBankController.ProcessTransfer)
		cashbank.POST("/deposit", permissionMiddleware.CanCreate("cash_bank"), cashBankController.ProcessDeposit)
		cashbank.POST("/withdrawal", permissionMiddleware.CanCreate("cash_bank"), cashBankController.ProcessWithdrawal)
		cashbank.GET("/accounts/:id/transactions", permissionMiddleware.CanView("cash_bank"), cashBankController.GetTransactions)
		
		// Reports
		cashbank.GET("/balance-summary", permissionMiddleware.CanView("cash_bank"), cashBankController.GetBalanceSummary)
		cashbank.POST("/accounts/:id/reconcile", permissionMiddleware.CanEdit("cash_bank"), cashBankController.ReconcileAccount)
		
		// Admin operations - GL Account linking fixes (keep role-based for admin-only operations)
		cashbank.GET("/admin/check-gl-links", middleware.RoleRequired("admin"), fixCashBankController.CheckCashBankGLLinks)
		cashbank.POST("/admin/fix-gl-links", middleware.RoleRequired("admin"), fixCashBankController.FixCashBankGLLinks)
	}
	
	// 🚀 Phase 1: CashBank-COA Synchronization Routes
	// Add validation middleware routes to router for health checks and sync management
	cashBankValidationMiddleware.AddRoutes(router)
	
	// Apply validation middleware to cashbank operations for automatic sync checking
	cashbank.Use(cashBankValidationMiddleware.ValidateCashBankSync())
}
