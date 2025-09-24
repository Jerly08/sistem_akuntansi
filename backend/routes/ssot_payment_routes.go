package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupSSOTPaymentRoutes sets up payment routes using SSOT journal integration
func SetupSSOTPaymentRoutes(router *gin.RouterGroup, db *gorm.DB, jwtManager *middleware.JWTManager) {
	// Initialize repositories
	paymentRepo := repositories.NewPaymentRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	salesRepo := repositories.NewSalesRepository(db)
	purchaseRepo := repositories.NewPurchaseRepository(db)

	// Initialize SSOT unified journal service
	unifiedJournalService := services.NewUnifiedJournalService(db)

	// Initialize Enhanced Payment Service with Journal integration
	enhancedPaymentService := services.NewEnhancedPaymentServiceWithJournal(
		db,
		*paymentRepo, // Dereference pointer to get interface implementation
		contactRepo,
		cashBankRepo,
		salesRepo,
		purchaseRepo,
		unifiedJournalService,
	)

	// Initialize SSOT Payment Controller
	ssotPaymentController := controllers.NewSSOTPaymentController(enhancedPaymentService)

	// Initialize permission middleware
	permissionMiddleware := middleware.NewPermissionMiddleware(db)

	// SSOT Payment routes - replaces legacy payment routes
	ssotPayments := router.Group("/payments/ssot")
	ssotPayments.Use(middleware.PaymentRateLimit()) // Apply rate limiting
	if middleware.GlobalAuditLogger != nil {
		ssotPayments.Use(middleware.GlobalAuditLogger.PaymentAuditMiddleware()) // Apply audit logging
	}
	{
		// SSOT Payment CRUD operations with journal integration
		ssotPayments.POST("/receivable", permissionMiddleware.CanCreate("payments"), ssotPaymentController.CreateReceivablePayment)
		ssotPayments.POST("/payable", permissionMiddleware.CanCreate("payments"), ssotPaymentController.CreatePayablePayment)
		ssotPayments.GET("/:id", permissionMiddleware.CanView("payments"), ssotPaymentController.GetPaymentWithJournal)
		ssotPayments.POST("/:id/reverse", permissionMiddleware.CanEdit("payments"), ssotPaymentController.ReversePayment)

		// Journal integration endpoints
		ssotPayments.POST("/preview-journal", permissionMiddleware.CanView("payments"), ssotPaymentController.PreviewPaymentJournal)
		ssotPayments.GET("/:id/balance-updates", permissionMiddleware.CanView("payments"), ssotPaymentController.GetAccountBalanceUpdates)
		
		// Legacy compatibility (deprecated - returns guidance)
		ssotPayments.GET("", permissionMiddleware.CanView("payments"), ssotPaymentController.GetPayments)
	}

	// Mark legacy payment routes as deprecated by adding a redirect route
	legacyPayments := router.Group("/payments")
	{
		legacyPayments.GET("/deprecated-notice", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"notice": "Legacy payment endpoints are deprecated",
				"migration": "Please use /api/v1/payments/ssot/* endpoints for full SSOT journal integration",
				"warnings": []string{
					"Legacy endpoints may cause double posting",
					"SSOT endpoints prevent balance inconsistencies",
					"Full journal audit trail available in SSOT endpoints",
				},
				"available_endpoints": []string{
					"POST /api/v1/payments/ssot/receivable - Create customer payment with journal",
					"POST /api/v1/payments/ssot/payable - Create vendor payment with journal", 
					"GET /api/v1/payments/ssot/:id - Get payment with journal details",
					"POST /api/v1/payments/ssot/:id/reverse - Reverse payment with journal",
					"POST /api/v1/payments/ssot/preview-journal - Preview journal entry",
					"GET /api/v1/payments/ssot/:id/balance-updates - Get account balance updates",
				},
			})
		})
	}
}