package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// UpdateSwaggerDocs dynamically updates the generated Swagger documentation
// This allows us to override compile-time annotations with runtime configuration
func UpdateSwaggerDocs() {
	swaggerConfig := GetSwaggerConfig()
	
	// Path to the generated swagger.json
	docsPath := filepath.Join("docs", "swagger.json")
	
	// Check if swagger.json exists
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		log.Printf("Swagger docs not found at %s, skipping dynamic update", docsPath)
		return
	}
	
	// Read the existing swagger.json
	data, err := os.ReadFile(docsPath)
	if err != nil {
		log.Printf("Failed to read swagger.json: %v", err)
		return
	}
	
	// Parse the JSON
	var swaggerDoc map[string]interface{}
	if err := json.Unmarshal(data, &swaggerDoc); err != nil {
		log.Printf("Failed to parse swagger.json: %v", err)
		return
	}
	
	// Update the dynamic fields
	swaggerDoc["host"] = swaggerConfig.Host
	swaggerDoc["schemes"] = []string{swaggerConfig.Scheme}
	
	// Update info section if it exists
	if info, ok := swaggerDoc["info"].(map[string]interface{}); ok {
		info["title"] = swaggerConfig.Title
		info["description"] = swaggerConfig.Description
	}
	
	// Update basePath
	swaggerDoc["basePath"] = swaggerConfig.BasePath
	
	// Marshal back to JSON
	updatedData, err := json.MarshalIndent(swaggerDoc, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal updated swagger.json: %v", err)
		return
	}
	
	// Write back to file
	if err := os.WriteFile(docsPath, updatedData, 0644); err != nil {
		log.Printf("Failed to write updated swagger.json: %v", err)
		return
	}
	
	// Also update the go docs if they exist
	updateSwaggerGoDocs(swaggerConfig)
	
	log.Printf("✅ Swagger docs updated dynamically: %s", swaggerConfig.GetSwaggerURL())
}

// updateSwaggerGoDocs updates the generated docs.go file with dynamic values
func updateSwaggerGoDocs(config *SwaggerConfig) {
	docsGoPath := filepath.Join("docs", "docs.go")
	
	if _, err := os.Stat(docsGoPath); os.IsNotExist(err) {
		return // docs.go doesn't exist, skip
	}
	
	// Read the docs.go file
	data, err := os.ReadFile(docsGoPath)
	if err != nil {
		log.Printf("Failed to read docs.go: %v", err)
		return
	}
	
	content := string(data)
	
	// Replace the host, schemes, and other dynamic values
	content = replaceInSwaggerDoc(content, `\"host\":`, fmt.Sprintf(`\"host\": \"%s\"`, config.Host))
	content = replaceInSwaggerDoc(content, `\"schemes\":`, fmt.Sprintf(`\"schemes\": [\"%s\"]`, config.Scheme))
	
	// Write back
	if err := os.WriteFile(docsGoPath, []byte(content), 0644); err != nil {
		log.Printf("Failed to write updated docs.go: %v", err)
	}
}

// replaceInSwaggerDoc replaces a field in the swagger doc string
func replaceInSwaggerDoc(content, field, replacement string) string {
	lines := strings.Split(content, "\\n")
	for i, line := range lines {
		if strings.Contains(line, field) {
			// Find the start and end of the field value
			start := strings.Index(line, field)
			if start != -1 {
				// Find the comma or closing brace
				end := strings.Index(line[start:], ",")
				if end == -1 {
					end = strings.Index(line[start:], "}")
				}
				if end != -1 {
					end += start
					lines[i] = line[:start] + replacement + "," + line[end+1:]
				}
			}
			break
		}
	}
	return strings.Join(lines, "\\n")
}

// PrintSwaggerInfo prints helpful Swagger configuration information
func PrintSwaggerInfo() {
	cfg := LoadConfig()
	swaggerConfig := GetSwaggerConfig()
	
	fmt.Printf("🚀 Swagger Configuration:\n")
	fmt.Printf("   Environment: %s\n", cfg.Environment)
	fmt.Printf("   Swagger URL: %s\n", swaggerConfig.GetSwaggerURL())
	fmt.Printf("   API Base URL: %s\n", swaggerConfig.GetAPIBaseURL())
	fmt.Printf("   Host: %s\n", swaggerConfig.Host)
	fmt.Printf("   Scheme: %s\n", swaggerConfig.Scheme)
	
	// Print CORS origins
	origins := GetAllowedOrigins(cfg)
	fmt.Printf("   CORS Origins: %v\n", origins)
	
	// Print environment variable hints
	if cfg.Environment == "production" {
		fmt.Printf("\n💡 Production Environment Variables:\n")
		fmt.Printf("   SWAGGER_HOST: Set your production domain (e.g., api.yourdomain.com)\n")
		fmt.Printf("   SWAGGER_SCHEME: https (recommended for production)\n")
		fmt.Printf("   ALLOWED_ORIGINS: Your frontend URL(s)\n")
		fmt.Printf("   DOMAIN or APP_URL: Your main domain\n")
		fmt.Printf("   ENABLE_HTTPS: true (recommended for production)\n")
	} else {
		fmt.Printf("\n💡 Development Mode - Dynamic Configuration Active\n")
		fmt.Printf("   To override: Set SWAGGER_HOST, SWAGGER_SCHEME, ALLOWED_ORIGINS\n")
	}
	fmt.Printf("\n")
}