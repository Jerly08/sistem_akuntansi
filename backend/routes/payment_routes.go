package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

func SetupPaymentRoutes(router *gin.RouterGroup, paymentController *controllers.PaymentController, cashBankController *controllers.CashBankController, cashBankService *services.CashBankService, jwtManager *middleware.JWTManager) {
	// Payment routes
	payment := router.Group("/payments")
	payment.Use(middleware.PaymentRateLimit()) // Apply rate limiting to all payment endpoints
	if middleware.GlobalAuditLogger != nil {
		payment.Use(middleware.GlobalAuditLogger.PaymentAuditMiddleware()) // Apply audit logging
	}
	{
		// Payment routes with appropriate role restrictions
		payment.GET("", middleware.RoleRequired("admin", "finance", "director"), paymentController.GetPayments)
		payment.GET("/:id", middleware.RoleRequired("admin", "finance", "director"), paymentController.GetPaymentByID)
		payment.POST("/receivable", middleware.RoleRequired("admin", "finance"), paymentController.CreateReceivablePayment)
		payment.POST("/payable", middleware.RoleRequired("admin", "finance"), paymentController.CreatePayablePayment)
		payment.POST("/:id/cancel", middleware.RoleRequired("admin", "finance"), paymentController.CancelPayment)
		payment.GET("/unpaid-invoices/:customer_id", middleware.RoleRequired("admin", "finance", "director"), paymentController.GetUnpaidInvoices)
		payment.GET("/unpaid-bills/:vendor_id", middleware.RoleRequired("admin", "finance", "director"), paymentController.GetUnpaidBills)
		payment.GET("/summary", middleware.RoleRequired("admin", "finance", "director"), paymentController.GetPaymentSummary)
		payment.GET("/analytics", middleware.RoleRequired("admin", "finance", "director"), paymentController.GetPaymentAnalytics)
		
		// Export routes
		payment.GET("/report/pdf", middleware.RoleRequired("admin", "finance", "director"), paymentController.ExportPaymentReportPDF)
		payment.GET("/export/excel", middleware.RoleRequired("admin", "finance", "director"), paymentController.ExportPaymentReportExcel)
		payment.GET("/:id/pdf", middleware.RoleRequired("admin", "finance", "director"), paymentController.ExportPaymentDetailPDF)
	}
	
	// Cash & Bank routes
	cashbank := router.Group("/cashbank")
	{
		// Account management
		cashbank.GET("/accounts", middleware.RoleRequired("admin", "finance", "director", "employee"), cashBankController.GetAccounts)
		
		// Payment accounts endpoint - specifically for payment form dropdowns
		cashbank.GET("/payment-accounts", middleware.RoleRequired("admin", "finance", "director", "employee"), cashBankController.GetPaymentAccounts)
		
		cashbank.GET("/accounts/:id", middleware.RoleRequired("admin", "finance", "director"), cashBankController.GetAccountByID)
		cashbank.POST("/accounts", middleware.RoleRequired("admin", "finance"), cashBankController.CreateAccount)
		cashbank.PUT("/accounts/:id", middleware.RoleRequired("admin", "finance"), cashBankController.UpdateAccount)
		
		// Transactions
		cashbank.POST("/transfer", middleware.RoleRequired("admin", "finance", "director"), cashBankController.ProcessTransfer)
		cashbank.POST("/deposit", middleware.RoleRequired("admin", "finance", "director"), cashBankController.ProcessDeposit)
		cashbank.POST("/withdrawal", middleware.RoleRequired("admin", "finance", "director"), cashBankController.ProcessWithdrawal)
		cashbank.GET("/accounts/:id/transactions", middleware.RoleRequired("admin", "finance", "director"), cashBankController.GetTransactions)
		
		// Reports
		cashbank.GET("/balance-summary", middleware.RoleRequired("admin", "finance", "director"), cashBankController.GetBalanceSummary)
		cashbank.POST("/accounts/:id/reconcile", middleware.RoleRequired("admin", "finance"), cashBankController.ReconcileAccount)
	}
}
