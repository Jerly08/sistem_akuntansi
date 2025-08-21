package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/config"
	"github.com/golang-jwt/jwt/v5"
)

type EnhancedClaims struct {
	UserID      uint     `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	SessionID   string   `json:"session_id"`
	DeviceInfo  string   `json:"device_info"`
	IPAddress   string   `json:"ip_address"`
	TokenType   string   `json:"token_type"` // access or refresh
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run debug_jwt.go <jwt_token>")
	}

	tokenString := os.Args[1]

	// Load config to get JWT secret
	cfg := config.LoadConfig()

	// Parse token
	claims := &EnhancedClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		log.Printf("Error parsing token: %v", err)
		return
	}

	if !token.Valid {
		log.Printf("Token is not valid")
		return
	}

	// Print all claims
	fmt.Printf("=== JWT TOKEN CLAIMS ===\n")
	fmt.Printf("Token Valid: %v\n", token.Valid)
	fmt.Printf("User ID: %d\n", claims.UserID)
	fmt.Printf("Username: %s\n", claims.Username)
	fmt.Printf("Email: %s\n", claims.Email)
	fmt.Printf("Role: %s\n", claims.Role)
	fmt.Printf("Session ID: %s\n", claims.SessionID)
	fmt.Printf("Device Info: %s\n", claims.DeviceInfo)
	fmt.Printf("IP Address: %s\n", claims.IPAddress)
	fmt.Printf("Token Type: %s\n", claims.TokenType)
	fmt.Printf("Permissions: %v\n", claims.Permissions)
	fmt.Printf("Issued At: %v\n", claims.IssuedAt.Time)
	fmt.Printf("Expires At: %v\n", claims.ExpiresAt.Time)
	fmt.Printf("Not Before: %v\n", claims.NotBefore.Time)
	fmt.Printf("Issuer: %s\n", claims.Issuer)
	fmt.Printf("Subject: %s\n", claims.Subject)
	fmt.Printf("ID: %s\n", claims.ID)
	fmt.Printf("========================\n")

	// Check if token is expired
	if claims.ExpiresAt.Time.Before(jwt.TimeFunc()) {
		fmt.Printf("⚠️ TOKEN IS EXPIRED!\n")
	} else {
		fmt.Printf("✅ Token is still valid\n")
	}
}
