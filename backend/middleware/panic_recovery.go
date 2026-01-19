package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

// RecoverPanic returns a middleware that recovers from panics and returns a proper error response
// This ensures that panics don't crash the server and provide meaningful error information
func RecoverPanic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stackTrace := string(debug.Stack())
				
				// Log the panic with full details
				log.Printf("❌ PANIC RECOVERED in %s %s", c.Request.Method, c.Request.URL.Path)
				log.Printf("❌ Panic error: %v", err)
				log.Printf("❌ Stack trace:\n%s", stackTrace)
				
				// Get user ID if available
				userID := c.GetUint("user_id")
				
				// Build error response
				errorResponse := gin.H{
					"error":     "Internal server error",
					"details":   fmt.Sprintf("An unexpected error occurred: %v", err),
					"timestamp": time.Now().Format(time.RFC3339),
					"path":      c.Request.URL.Path,
					"method":    c.Request.Method,
				}
				
				// Add user_id if available
				if userID > 0 {
					errorResponse["user_id"] = userID
				}
				
				// Return 500 error with details
				c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
			}
		}()
		
		c.Next()
	}
}

// RecoverPanicWithLogger returns a middleware that recovers from panics with custom logging
func RecoverPanicWithLogger(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stackTrace := string(debug.Stack())
				
				// Log the panic with full details
				if logger != nil {
					logger.Printf("❌ PANIC RECOVERED in %s %s", c.Request.Method, c.Request.URL.Path)
					logger.Printf("❌ Panic error: %v", err)
					logger.Printf("❌ Stack trace:\n%s", stackTrace)
				} else {
					log.Printf("❌ PANIC RECOVERED in %s %s", c.Request.Method, c.Request.URL.Path)
					log.Printf("❌ Panic error: %v", err)
					log.Printf("❌ Stack trace:\n%s", stackTrace)
				}
				
				// Get user ID if available
				userID := c.GetUint("user_id")
				
				// Build error response
				errorResponse := gin.H{
					"error":     "Internal server error",
					"details":   fmt.Sprintf("An unexpected error occurred: %v", err),
					"timestamp": time.Now().Format(time.RFC3339),
					"path":      c.Request.URL.Path,
					"method":    c.Request.Method,
				}
				
				// Add user_id if available
				if userID > 0 {
					errorResponse["user_id"] = userID
				}
				
				// Return 500 error with details
				c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
			}
		}()
		
		c.Next()
	}
}
