package routes

import (
	"context"
	"strconv"
	"time"

	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupFastPaymentRoutes sets up lightweight, high-performance payment routes
func SetupFastPaymentRoutes(router *gin.RouterGroup, db *gorm.DB, jwtManager *middleware.JWTManager) {
	// Initialize repositories
	paymentRepo := repositories.NewPaymentRepository(db)
	salesRepo := repositories.NewSalesRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)

	// Initialize lightweight payment service
	lightweightPaymentService := services.NewLightweightPaymentService(
		db,
		*paymentRepo,
		*salesRepo,
		contactRepo,
		*cashBankRepo,
	)

	// Initialize fast payment controller
	fastPaymentController := controllers.NewFastPaymentController(lightweightPaymentService)

	// Initialize permission middleware
	permissionMiddleware := middleware.NewPermissionMiddleware(db)

	// Fast payment routes - optimized for performance
	fastPayments := router.Group("/payments/fast")
	fastPayments.Use(middleware.PaymentRateLimit()) // Apply rate limiting
	if middleware.GlobalAuditLogger != nil {
		fastPayments.Use(middleware.GlobalAuditLogger.PaymentAuditMiddleware()) // Apply audit logging
	}
	{
		// Core fast payment operations
		fastPayments.POST("/record", 
			jwtManager.AuthRequired(), 
			permissionMiddleware.CanCreate("payments"), 
			fastPaymentController.RecordPaymentFast)

		fastPayments.POST("/record-async", 
			jwtManager.AuthRequired(), 
			permissionMiddleware.CanCreate("payments"), 
			fastPaymentController.RecordPaymentWithAsyncJournal)

		// Validation endpoint (no authentication required for better UX)
		fastPayments.POST("/validate", fastPaymentController.ValidatePayment)

		// Status checking
		fastPayments.GET("/status/:id", 
			jwtManager.AuthRequired(), 
			permissionMiddleware.CanView("payments"), 
			fastPaymentController.GetPaymentStatus)

		// Health check (public endpoint for monitoring)
		fastPayments.GET("/health", fastPaymentController.HealthCheck)
	}

	// Sales-specific fast payment endpoints
	salesFastPayments := router.Group("/sales/fast-payment")
	salesFastPayments.Use(jwtManager.AuthRequired())
	salesFastPayments.Use(middleware.PaymentRateLimit())
	{
		// Direct sales payment recording
		salesFastPayments.POST("/:id/record", 
			permissionMiddleware.CanCreate("payments"), 
			func(c *gin.Context) {
				// Extract sale ID from URL and inject into request body
				saleID := c.Param("id")
				
				var req services.LightweightPaymentRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{
						"success": false,
						"error":   "Invalid request format",
						"details": err.Error(),
					})
					return
				}

				// Parse sale ID from URL
				if saleID != "" {
					if id, err := strconv.ParseUint(saleID, 10, 32); err == nil {
						req.SaleID = uint(id)
					}
				}

				// Set user ID from JWT
				if userID, exists := c.Get("user_id"); exists {
					req.UserID = userID.(uint)
				}

				// Use fast payment service
				ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
				defer cancel()

				response, err := lightweightPaymentService.RecordPaymentWithAsyncJournal(ctx, &req)
				if err != nil {
					c.JSON(500, gin.H{
						"success": false,
						"error":   "Payment recording failed",
						"details": err.Error(),
					})
					return
				}

				c.JSON(200, response)
			})
	}
}