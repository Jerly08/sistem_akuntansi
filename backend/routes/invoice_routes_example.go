package routes

import (
	"github.com/gin-gonic/gin"
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

// SetupInvoiceRoutes registers all invoice-related routes
// This is an example of how to integrate invoice routes with settings
func SetupInvoiceRoutes(protected *gin.RouterGroup, db *gorm.DB) {
	// Initialize repositories
	contactRepo := repositories.NewContactRepository(db)
	productRepo := repositories.NewProductRepository(db)
	
	// Initialize services
	invoiceService := services.NewInvoiceServiceFull(db, contactRepo, productRepo)
	quoteService := services.NewQuoteServiceFull(db, contactRepo, productRepo)
	
	// Initialize controllers
	invoiceController := controllers.NewInvoiceController(invoiceService)
	quoteController := controllers.NewQuoteController(quoteService, invoiceService)
	
	// Invoice routes
	invoices := protected.Group("/invoices")
	{
		invoices.GET("", invoiceController.GetInvoices)
		invoices.GET("/:id", invoiceController.GetInvoice)
		invoices.POST("", invoiceController.CreateInvoice)
		invoices.PUT("/:id", invoiceController.UpdateInvoice)
		invoices.DELETE("/:id", invoiceController.DeleteInvoice)
		
		// Utility endpoints
		invoices.POST("/generate-code", invoiceController.GenerateInvoiceCode)
		invoices.POST("/format-currency", invoiceController.FormatCurrency)
	}
	
	// Quote routes
	quotes := protected.Group("/quotes")
	{
		quotes.GET("", quoteController.GetQuotes)
		quotes.GET("/:id", quoteController.GetQuote)
		quotes.POST("", quoteController.CreateQuote)
		quotes.PUT("/:id", quoteController.UpdateQuote)
		quotes.DELETE("/:id", quoteController.DeleteQuote)
		
		// Utility endpoints
		quotes.POST("/generate-code", quoteController.GenerateQuoteCode)
		quotes.POST("/format-currency", quoteController.FormatCurrency)
		quotes.POST("/:id/convert-to-invoice", quoteController.ConvertToInvoice)
	}
}