package controllers

import (
	"log"
	"net/http"
	"strings"
	"time"
	"app-sistem-akuntansi/middleware"
	"app-sistem-akuntansi/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController struct {
	DB *gorm.DB
}

func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{DB: db}
}

// Helper function to convert role to uppercase for frontend
func convertRoleToUppercase(role string) string {
	switch role {
	case "admin":
		return "ADMIN"
	case "finance":
		return "FINANCE"
	case "director":
		return "DIRECTOR"
	case "inventory_manager":
		return "INVENTORY_MANAGER"
	case "employee":
		return "EMPLOYEE"
	default:
		return strings.ToUpper(role)
	}
}

// Helper function to convert role to lowercase for backend
func convertRoleToLowercase(role string) string {
	switch role {
	case "ADMIN":
		return "admin"
	case "FINANCE":
		return "finance"
	case "DIRECTOR":
		return "director"
	case "INVENTORY_MANAGER":
		return "inventory_manager"
	case "EMPLOYEE":
		return "employee"
	default:
		return strings.ToLower(role)
	}
}

func (ac *AuthController) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := ac.DB.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Set default role if not provided
	if req.Role == "" {
		req.Role = "employee"
	}

	// Create user
	user := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
		IsActive:  true,
	}

	if err := ac.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Initialize JWT Manager
	jw := middleware.NewJWTManager(ac.DB)

	// Generate token pair
	tokens, err := jw.GenerateTokenPair(user, "Web Browser", c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"access_token":   tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user":    tokens.User,
	})
}

func (ac *AuthController) Login(c *gin.Context) {
	// Support both simple and enhanced login requests
	var simpleReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	
	var enhancedReq models.EnhancedLoginRequest
	
	// Try to bind as simple request first (for frontend compatibility)
	if err := c.ShouldBindJSON(&simpleReq); err != nil {
		// If simple binding fails, try enhanced request
		if err := c.ShouldBindJSON(&enhancedReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}
	}
	
	// Debug logging
	if simpleReq.Email != "" {
		log.Printf("Login attempt - Email: %s, Password length: %d", simpleReq.Email, len(simpleReq.Password))
	} else {
		log.Printf("Login attempt - Identifier: %s, Password length: %d", enhancedReq.EmailOrUsername, len(enhancedReq.Password))
	}
	
	// Determine the identifier and password
	var identifier, password, deviceInfo string
	if simpleReq.Email != "" {
		identifier = simpleReq.Email
		password = simpleReq.Password
		deviceInfo = "Web Browser"
	} else {
		identifier = enhancedReq.EmailOrUsername
		password = enhancedReq.Password
		deviceInfo = enhancedReq.DeviceInfo
	}

	// Find user by email or username
	var user models.User
	if err := ac.DB.Where("username = ? OR email = ?", identifier, identifier).First(&user).Error; err != nil {
		ac.logAuthAttempt(identifier, false, models.FailureReasonInvalidCredentials, c.ClientIP(), c.Request.UserAgent())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		ac.logAuthAttempt(identifier, false, models.FailureReasonAccountDisabled, c.ClientIP(), c.Request.UserAgent())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is deactivated"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		ac.logAuthAttempt(identifier, false, models.FailureReasonInvalidCredentials, c.ClientIP(), c.Request.UserAgent())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Initialize JWT Manager
	jw := middleware.NewJWTManager(ac.DB)

	// Generate token pair
	tokens, err := jw.GenerateTokenPair(user, deviceInfo, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	ac.logAuthAttempt(identifier, true, "", c.ClientIP(), c.Request.UserAgent())
	
	// Return response in format expected by frontend
	c.JSON(http.StatusOK, gin.H{
		"access_token":        tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user":         tokens.User,
		"message":      "Login successful",
	})
}

// Log authentication attempts
func (ac *AuthController) logAuthAttempt(identifier string, success bool, reason, ipAddress, userAgent string) {
	authAttempt := models.AuthAttempt{
		Email:         identifier,
		Username:      identifier,
		Success:       success,
		FailureReason: reason,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		AttemptedAt:   time.Now(),
	}
	
	ac.DB.Create(&authAttempt)
}

func (ac *AuthController) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jw := middleware.NewJWTManager(ac.DB)
	tokens, err := jw.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":        tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user":         tokens.User,
		"message":      "Token refreshed successfully",
	})
}

func (ac *AuthController) Profile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	var user models.User
	if err := ac.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile retrieved successfully",
		"data":    user,
	})
}

// ValidateToken validates if the current token is valid and active
func (ac *AuthController) ValidateToken(c *gin.Context) {
	// If we reach this point, it means the JWT middleware has already validated the token
	// and set the user context, so the token is valid
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token - user ID not found",
			"code":  "INVALID_TOKEN",
			"valid": false,
		})
		return
	}
	
	// Double-check that the user still exists and is active
	var user models.User
	if err := ac.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
			"code":  "USER_NOT_FOUND",
			"valid": false,
		})
		return
	}
	
	// Check if user is still active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User account is disabled",
			"code":  "ACCOUNT_DISABLED",
			"valid": false,
		})
		return
	}
	
	// Token is valid
	c.JSON(http.StatusOK, gin.H{
		"message": "Token is valid",
		"valid":   true,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}
