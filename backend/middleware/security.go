package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"app-sistem-akuntansi/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SecurityManager struct {
	DB           *gorm.DB
	rateLimiters map[string]*RateLimiter
	mutex        sync.RWMutex
}

type RateLimiter struct {
	Requests  int
	Window    time.Duration
	requests  map[string][]time.Time
	mutex     sync.RWMutex
	LastClean time.Time
}

func NewSecurityManager(db *gorm.DB) *SecurityManager {
	return &SecurityManager{
		DB:           db,
		rateLimiters: make(map[string]*RateLimiter),
	}
}

// SecurityHeaders adds security headers to all responses
func (sm *SecurityManager) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		// HSTS header for HTTPS
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"media-src 'self'; " +
			"object-src 'none'; " +
			"frame-src 'none';"
		c.Header("Content-Security-Policy", csp)
		
		c.Next()
	}
}

// RateLimit implements rate limiting per IP address
func (sm *SecurityManager) RateLimit(requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		endpoint := c.Request.URL.Path
		
		// Create rate limiter key
		key := clientIP + ":" + endpoint
		
		sm.mutex.Lock()
		limiter, exists := sm.rateLimiters[key]
		if !exists {
			limiter = &RateLimiter{
				Requests:  requests,
				Window:    window,
				requests:  make(map[string][]time.Time),
				LastClean: time.Now(),
			}
			sm.rateLimiters[key] = limiter
		}
		sm.mutex.Unlock()
		
		if !limiter.Allow(clientIP) {
			// Log rate limit violation
			sm.logRateLimitViolation(clientIP, endpoint)
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
				"code":  "RATE_LIMIT_EXCEEDED",
				"retry_after": int(window.Seconds()),
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// AuthRateLimit implements stricter rate limiting for auth endpoints
func (sm *SecurityManager) AuthRateLimit() gin.HandlerFunc {
	return sm.RateLimit(5, 15*time.Minute) // 5 attempts per 15 minutes
}

// StrictRateLimit implements very strict rate limiting for sensitive endpoints
func (sm *SecurityManager) StrictRateLimit() gin.HandlerFunc {
	return sm.RateLimit(10, 1*time.Hour) // 10 requests per hour
}

// Allow checks if a request is allowed based on rate limiting
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	
	// Clean old requests periodically
	if now.Sub(rl.LastClean) > rl.Window {
		rl.cleanOldRequests(now)
		rl.LastClean = now
	}
	
	// Get requests for this IP
	requests := rl.requests[ip]
	
	// Remove expired requests
	validRequests := make([]time.Time, 0)
	for _, reqTime := range requests {
		if now.Sub(reqTime) < rl.Window {
			validRequests = append(validRequests, reqTime)
		}
	}
	
	// Check if limit exceeded
	if len(validRequests) >= rl.Requests {
		return false
	}
	
	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[ip] = validRequests
	
	return true
}

func (rl *RateLimiter) cleanOldRequests(now time.Time) {
	for ip, requests := range rl.requests {
		validRequests := make([]time.Time, 0)
		for _, reqTime := range requests {
			if now.Sub(reqTime) < rl.Window {
				validRequests = append(validRequests, reqTime)
			}
		}
		
		if len(validRequests) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = validRequests
		}
	}
}

// CORS middleware with secure defaults
func (sm *SecurityManager) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Define allowed origins (configure based on your needs)
		allowedOrigins := []string{
			"http://localhost:3000",  // Development frontend
			"https://your-domain.com", // Production frontend
		}
		
		// Check if origin is allowed
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}
		
		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// RequestLogging logs all requests for security monitoring
func (sm *SecurityManager) RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		
		// Skip logging for health checks and static assets
		if shouldSkipLogging(path) {
			c.Next()
			return
		}
		
		c.Next()
		
		// Log after request
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		
		// Log suspicious activities
		if statusCode >= 400 || latency > 10*time.Second {
			sm.logSuspiciousActivity(c, statusCode, latency)
		}
	}
}

// IP Whitelist middleware (optional, for admin endpoints)
func (sm *SecurityManager) IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		for _, allowedIP := range allowedIPs {
			if clientIP == allowedIP || strings.HasPrefix(clientIP, allowedIP) {
				c.Next()
				return
			}
		}
		
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied from this IP",
			"code":  "IP_NOT_ALLOWED",
		})
		c.Abort()
	}
}

// Input Sanitization middleware
func (sm *SecurityManager) InputSanitization() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for common attack patterns in URL
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		
		suspiciousPatterns := []string{
			"<script", "</script>", "javascript:", "onerror=", "onload=",
			"../", "..\\", "/etc/passwd", "/proc/", "cmd.exe",
			"SELECT * FROM", "DROP TABLE", "INSERT INTO", "DELETE FROM",
			"' OR '1'='1", "' OR 1=1", "UNION SELECT",
		}
		
		for _, pattern := range suspiciousPatterns {
			if strings.Contains(strings.ToLower(path), strings.ToLower(pattern)) ||
				strings.Contains(strings.ToLower(query), strings.ToLower(pattern)) {
				
				sm.logSecurityViolation(c, "SUSPICIOUS_INPUT", pattern)
				
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid request format",
					"code":  "INVALID_INPUT",
				})
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

// Block User Agent middleware (block suspicious user agents)
func (sm *SecurityManager) BlockSuspiciousUserAgents() gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := c.Request.Header.Get("User-Agent")
		
		// Block empty user agents
		if userAgent == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User agent required",
				"code":  "USER_AGENT_MISSING",
			})
			c.Abort()
			return
		}
		
		// Block known malicious user agents
		blockedPatterns := []string{
			"sqlmap", "nmap", "nikto", "dirb", "gobuster",
			"masscan", "zap", "burp", "w3af", "skipfish",
		}
		
		userAgentLower := strings.ToLower(userAgent)
		for _, pattern := range blockedPatterns {
			if strings.Contains(userAgentLower, pattern) {
				sm.logSecurityViolation(c, "BLOCKED_USER_AGENT", userAgent)
				
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Access denied",
					"code":  "BLOCKED_USER_AGENT",
				})
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

// Helper functions

func (sm *SecurityManager) logRateLimitViolation(ip, endpoint string) {
	rateLimitRecord := models.RateLimitRecord{
		IPAddress:   ip,
		Endpoint:    endpoint,
		Attempts:    1,
		WindowStart: time.Now(),
	}
	
	// Update or create rate limit record
	var existingRecord models.RateLimitRecord
	if err := sm.DB.Where("ip_address = ? AND endpoint = ?", ip, endpoint).First(&existingRecord).Error; err != nil {
		// Create new record
		sm.DB.Create(&rateLimitRecord)
	} else {
		// Update existing record
		sm.DB.Model(&existingRecord).UpdateColumn("attempts", gorm.Expr("attempts + ?", 1))
	}
}

func (sm *SecurityManager) logSuspiciousActivity(c *gin.Context, statusCode int, latency time.Duration) {
	// Create audit log for suspicious activity
	auditLog := models.AuditLog{
		Action:      "SUSPICIOUS_ACTIVITY",
		TableName:   "security_monitor",
		RecordID:    0,
		NewValues:   fmt.Sprintf(`{"path":"%s","method":"%s","status_code":%d,"latency_ms":%d,"user_agent":"%s"}`,
			c.Request.URL.Path, c.Request.Method, statusCode, latency.Milliseconds(), c.Request.Header.Get("User-Agent")),
		IPAddress:   c.ClientIP(),
		UserAgent:   c.Request.Header.Get("User-Agent"),
	}
	
	// Try to get user ID if authenticated
	if userID, exists := c.Get("user_id"); exists {
		auditLog.UserID = userID.(uint)
	}
	
	sm.DB.Create(&auditLog)
}

func (sm *SecurityManager) logSecurityViolation(c *gin.Context, violationType, details string) {
	auditLog := models.AuditLog{
		Action:      "SECURITY_VIOLATION",
		TableName:   "security_monitor",
		RecordID:    0,
		NewValues:   fmt.Sprintf(`{"violation_type":"%s","details":"%s","path":"%s","method":"%s"}`,
			violationType, details, c.Request.URL.Path, c.Request.Method),
		IPAddress:   c.ClientIP(),
		UserAgent:   c.Request.Header.Get("User-Agent"),
	}
	
	sm.DB.Create(&auditLog)
}

func shouldSkipLogging(path string) bool {
	skipPaths := []string{
		"/health",
		"/ping",
		"/favicon.ico",
		"/robots.txt",
	}
	
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	
	// Skip static assets
	if strings.HasPrefix(path, "/static/") ||
		strings.HasPrefix(path, "/assets/") ||
		strings.HasSuffix(path, ".css") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".png") ||
		strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".ico") {
		return true
	}
	
	return false
}
