package controllers

import (
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

	// Convert role to uppercase for frontend compatibility
	userWithUppercaseRole := tokens.User
	userWithUppercaseRole.Role = convertRoleToUppercase(userWithUppercaseRole.Role)

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"token":   tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"user":    userWithUppercaseRole,
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
	
	// Convert role to uppercase for frontend compatibility
	userWithUppercaseRole := tokens.User
	userWithUppercaseRole.Role = convertRoleToUppercase(userWithUppercaseRole.Role)
	
	// Return response in format expected by frontend
	c.JSON(http.StatusOK, gin.H{
		"token":        tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"user":         userWithUppercaseRole,
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

	// Convert role to uppercase for frontend compatibility
	userWithUppercaseRole := tokens.User
	userWithUppercaseRole.Role = convertRoleToUppercase(userWithUppercaseRole.Role)

	// Update tokens with uppercase role
	updatedTokens := *tokens
	updatedTokens.User = userWithUppercaseRole

	c.JSON(http.StatusOK, gin.H{
		"token":        updatedTokens.AccessToken,
		"refreshToken": updatedTokens.RefreshToken,
		"user":         userWithUppercaseRole,
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

	// Convert role to uppercase for frontend compatibility
	user.Role = convertRoleToUppercase(user.Role)

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile retrieved successfully",
		"data":    user,
	})
}
