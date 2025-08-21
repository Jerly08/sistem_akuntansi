package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitRecord represents a rate limit entry
type RateLimitEntry struct {
	Count      int
	WindowStart time.Time
	BlockedUntil *time.Time
}

// RateLimiter manages rate limiting
type RateLimiter struct {
	mu      sync.RWMutex
	entries map[string]*RateLimitEntry
	limit   int
	window  time.Duration
	blockDuration time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration, blockDuration time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*RateLimitEntry),
		limit:   limit,
		window:  window,
		blockDuration: blockDuration,
	}
	
	// Start cleanup goroutine
	go rl.cleanup()
	
	return rl
}

// IsAllowed checks if a request from the given key is allowed
func (rl *RateLimiter) IsAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]

	if !exists {
		rl.entries[key] = &RateLimitEntry{
			Count:      1,
			WindowStart: now,
		}
		return true
	}

	// Check if currently blocked
	if entry.BlockedUntil != nil && now.Before(*entry.BlockedUntil) {
		return false
	}

	// Reset window if expired
	if now.Sub(entry.WindowStart) > rl.window {
		entry.Count = 1
		entry.WindowStart = now
		entry.BlockedUntil = nil
		return true
	}

	// Increment count
	entry.Count++

	// Check if limit exceeded
	if entry.Count > rl.limit {
		blockUntil := now.Add(rl.blockDuration)
		entry.BlockedUntil = &blockUntil
		return false
	}

	return true
}

// GetRemainingRequests returns the number of remaining requests in the current window
func (rl *RateLimiter) GetRemainingRequests(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	entry, exists := rl.entries[key]
	if !exists {
		return rl.limit
	}

	now := time.Now()
	if now.Sub(entry.WindowStart) > rl.window {
		return rl.limit
	}

	remaining := rl.limit - entry.Count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// cleanup removes expired entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		
		for key, entry := range rl.entries {
			// Remove entries that are old and not blocked
			if now.Sub(entry.WindowStart) > rl.window*2 && 
			   (entry.BlockedUntil == nil || now.After(*entry.BlockedUntil)) {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Global rate limiters for different endpoints
var (
	paymentRateLimiter = NewRateLimiter(100, time.Minute, 5*time.Minute)      // 100 requests per minute
	authRateLimiter    = NewRateLimiter(10, time.Minute, 10*time.Minute)      // 10 auth attempts per minute
	generalRateLimiter = NewRateLimiter(200, time.Minute, 2*time.Minute)      // 200 requests per minute
)

// PaymentRateLimit middleware for payment endpoints
func PaymentRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := getClientKey(c)
		
		if !paymentRateLimiter.IsAllowed(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Payment rate limit exceeded. Please try again later.",
				"code":  "PAYMENT_RATE_LIMIT_EXCEEDED",
				"retry_after": "5 minutes",
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining := paymentRateLimiter.GetRemainingRequests(key)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", paymentRateLimiter.limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Window", "60")

		c.Next()
	}
}

// AuthRateLimit middleware for authentication endpoints
func AuthRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := getClientKey(c)
		
		if !authRateLimiter.IsAllowed(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Authentication rate limit exceeded. Please try again later.",
				"code":  "AUTH_RATE_LIMIT_EXCEEDED",
				"retry_after": "10 minutes",
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining := authRateLimiter.GetRemainingRequests(key)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", authRateLimiter.limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Window", "60")

		c.Next()
	}
}

// GeneralRateLimit middleware for general endpoints
func GeneralRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := getClientKey(c)
		
		if !generalRateLimiter.IsAllowed(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
				"code":  "RATE_LIMIT_EXCEEDED",
				"retry_after": "2 minutes",
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining := generalRateLimiter.GetRemainingRequests(key)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", generalRateLimiter.limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Window", "60")

		c.Next()
	}
}

// getClientKey generates a key for rate limiting based on IP and user
func getClientKey(c *gin.Context) string {
	// Try to get user ID from context first
	userID := c.GetUint("user_id")
	if userID != 0 {
		return fmt.Sprintf("user_%d", userID)
	}

	// Fall back to IP address
	clientIP := c.ClientIP()
	return fmt.Sprintf("ip_%s", clientIP)
}

// GetRateLimitStatus returns current rate limit status for debugging
func GetRateLimitStatus(c *gin.Context) gin.H {
	key := getClientKey(c)
	
	return gin.H{
		"payment_remaining": paymentRateLimiter.GetRemainingRequests(key),
		"auth_remaining":    authRateLimiter.GetRemainingRequests(key),
		"general_remaining": generalRateLimiter.GetRemainingRequests(key),
		"client_key":        key,
	}
}
