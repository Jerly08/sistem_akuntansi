package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	log.Println("🔧 Starting comprehensive Swagger fixes...")

	// Read the current swagger.json file
	swaggerPath := "./docs/swagger.json"
	swaggerData, err := ioutil.ReadFile(swaggerPath)
	if err != nil {
		log.Printf("❌ Error reading swagger.json: %v", err)
		return
	}

	var swagger map[string]interface{}
	if err := json.Unmarshal(swaggerData, &swagger); err != nil {
		log.Printf("❌ Error parsing swagger.json: %v", err)
		return
	}

	// Fix 1: Update basePath to correct API version
	log.Println("🔧 Fixing basePath to /api/v1...")
	swagger["basePath"] = "/api/v1"

	// Fix 2: Fix all path references
	log.Println("🔧 Fixing path references...")
	fixPathReferences(swagger)

	// Fix 3: Add proper security definitions and authentication
	log.Println("🔧 Adding proper security definitions...")
	addSecurityDefinitions(swagger)

	// Fix 4: Add missing cash-bank routes
	log.Println("🔧 Adding missing cash-bank routes...")
	addCashBankRoutes(swagger)

	// Fix 5: Add receipt endpoints
	log.Println("🔧 Adding receipt endpoints...")
	addReceiptEndpoints(swagger)

	// Fix 6: Enhance schemas with business fields
	log.Println("🔧 Enhancing schemas with business fields...")
	enhanceBusinessSchemas(swagger)

	// Fix 7: Add complete contact CRUD
	log.Println("🔧 Adding complete contact CRUD...")
	addCompleteContactCRUD(swagger)

	// Fix 8: Fix content types and JSON parsing
	log.Println("🔧 Fixing content types and JSON parsing...")
	fixContentTypesAndParsing(swagger)

	// Write the fixed swagger file
	fixedSwagger, err := json.MarshalIndent(swagger, "", "  ")
	if err != nil {
		log.Printf("❌ Error marshaling fixed swagger: %v", err)
		return
	}

	if err := ioutil.WriteFile(swaggerPath, fixedSwagger, 0644); err != nil {
		log.Printf("❌ Error writing fixed swagger: %v", err)
		return
	}

	// Fix 9: Generate proper Swagger UI without green overlay
	log.Println("🔧 Generating fixed Swagger UI...")
	generateFixedSwaggerUI()

	log.Println("✅ All comprehensive fixes applied successfully!")
}

func fixPathReferences(swagger map[string]interface{}) {
	paths, ok := swagger["paths"].(map[string]interface{})
	if !ok {
		return
	}

	// Convert all paths to have proper /api/v1 prefix removed since basePath is now /api/v1
	newPaths := make(map[string]interface{})
	
	for path, definition := range paths {
		// Remove /api/v1 prefix if it exists, since basePath will handle it
		newPath := strings.TrimPrefix(path, "/api/v1")
		if newPath == "" {
			newPath = "/"
		}
		newPaths[newPath] = definition
	}
	
	swagger["paths"] = newPaths
}

func addSecurityDefinitions(swagger map[string]interface{}) {
	swagger["securityDefinitions"] = map[string]interface{}{
		"BearerAuth": map[string]interface{}{
			"type":        "apiKey",
			"name":        "Authorization",
			"in":          "header",
			"description": "JWT token for authentication. Format: Bearer {token}",
		},
	}

	swagger["security"] = []map[string]interface{}{
		{
			"BearerAuth": []string{},
		},
	}
}

func addCashBankRoutes(swagger map[string]interface{}) {
	paths, ok := swagger["paths"].(map[string]interface{})
	if !ok {
		return
	}

	// Add cash-bank routes
	cashBankRoutes := map[string]interface{}{
		"/cash-bank/accounts": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"CashBank"},
				"summary":     "Get all cash/bank accounts",
				"description": "Retrieve a list of all cash and bank accounts",
				"produces":    []string{"application/json"},
				"parameters": []map[string]interface{}{
					{
						"name":        "page",
						"in":          "query",
						"description": "Page number",
						"type":        "integer",
						"default":     1,
					},
					{
						"name":        "limit",
						"in":          "query",
						"description": "Number of items per page",
						"type":        "integer",
						"default":     10,
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Successful operation",
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"status": map[string]interface{}{
									"type":    "string",
									"example": "success",
								},
								"data": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"$ref": "#/definitions/CashBank",
									},
								},
							},
						},
					},
					"401": map[string]interface{}{
						"description": "Unauthorized",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/ErrorResponse",
						},
					},
				},
				"security": []map[string]interface{}{
					{"BearerAuth": []string{}},
				},
			},
			"post": map[string]interface{}{
				"tags":        []string{"CashBank"},
				"summary":     "Create cash/bank account",
				"description": "Create a new cash or bank account",
				"consumes":    []string{"application/json"},
				"produces":    []string{"application/json"},
				"parameters": []map[string]interface{}{
					{
						"in":       "body",
						"name":     "body",
						"required": true,
						"schema": map[string]interface{}{
							"$ref": "#/definitions/CashBankCreate",
						},
					},
				},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{
						"description": "Cash/Bank account created successfully",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/ApiResponse",
						},
					},
					"400": map[string]interface{}{
						"description": "Bad request",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/ErrorResponse",
						},
					},
				},
				"security": []map[string]interface{}{
					{"BearerAuth": []string{}},
				},
			},
		},
		"/cash-bank/accounts/{id}": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"CashBank"},
				"summary":     "Get cash/bank account by ID",
				"description": "Retrieve a specific cash/bank account by its ID",
				"produces":    []string{"application/json"},
				"parameters": []map[string]interface{}{
					{
						"name":        "id",
						"in":          "path",
						"description": "Cash/Bank account ID",
						"required":    true,
						"type":        "integer",
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Successful operation",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/CashBank",
						},
					},
					"404": map[string]interface{}{
						"description": "Cash/Bank account not found",
					},
				},
				"security": []map[string]interface{}{
					{"BearerAuth": []string{}},
				},
			},
			"put": map[string]interface{}{
				"tags":        []string{"CashBank"},
				"summary":     "Update cash/bank account",
				"description": "Update an existing cash/bank account",
				"consumes":    []string{"application/json"},
				"produces":    []string{"application/json"},
				"parameters": []map[string]interface{}{
					{
						"name":        "id",
						"in":          "path",
						"description": "Cash/Bank account ID",
						"required":    true,
						"type":        "integer",
					},
					{
						"in":       "body",
						"name":     "body",
						"required": true,
						"schema": map[string]interface{}{
							"$ref": "#/definitions/CashBankUpdate",
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Cash/Bank account updated successfully",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/ApiResponse",
						},
					},
				},
				"security": []map[string]interface{}{
					{"BearerAuth": []string{}},
				},
			},
			"delete": map[string]interface{}{
				"tags":        []string{"CashBank"},
				"summary":     "Delete cash/bank account",
				"description": "Delete a cash/bank account",
				"produces":    []string{"application/json"},
				"parameters": []map[string]interface{}{
					{
						"name":        "id",
						"in":          "path",
						"description": "Cash/Bank account ID",
						"required":    true,
						"type":        "integer",
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Cash/Bank account deleted successfully",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/ApiResponse",
						},
					},
				},
				"security": []map[string]interface{}{
					{"BearerAuth": []string{}},
				},
			},
		},
	}

	for path, definition := range cashBankRoutes {
		paths[path] = definition
	}
}

func addReceiptEndpoints(swagger map[string]interface{}) {
	paths, ok := swagger["paths"].(map[string]interface{})
	if !ok {
		return
	}

	receiptRoutes := map[string]interface{}{
		"/receipts": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Receipts"},
				"summary":     "Get all receipts",
				"description": "Retrieve a list of all receipts",
				"produces":    []string{"application/json"},
				"parameters": []map[string]interface{}{
					{
						"name":        "page",
						"in":          "query",
						"description": "Page number",
						"type":        "integer",
						"default":     1,
					},
					{
						"name":        "limit",
						"in":          "query",
						"description": "Number of items per page",
						"type":        "integer",
						"default":     10,
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Successful operation",
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"status": map[string]interface{}{
									"type":    "string",
									"example": "success",
								},
								"data": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"$ref": "#/definitions/Receipt",
									},
								},
							},
						},
					},
				},
				"security": []map[string]interface{}{
					{"BearerAuth": []string{}},
				},
			},
			"post": map[string]interface{}{
				"tags":        []string{"Receipts"},
				"summary":     "Create receipt",
				"description": "Create a new receipt",
				"consumes":    []string{"application/json"},
				"produces":    []string{"application/json"},
				"parameters": []map[string]interface{}{
					{
						"in":       "body",
						"name":     "body",
						"required": true,
						"schema": map[string]interface{}{
							"$ref": "#/definitions/ReceiptCreate",
						},
					},
				},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{
						"description": "Receipt created successfully",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/ApiResponse",
						},
					},
				},
				"security": []map[string]interface{}{
					{"BearerAuth": []string{}},
				},
			},
		},
		"/receipts/{id}": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Receipts"},
				"summary":     "Get receipt by ID",
				"description": "Retrieve a specific receipt by its ID",
				"produces":    []string{"application/json"},
				"parameters": []map[string]interface{}{
					{
						"name":        "id",
						"in":          "path",
						"description": "Receipt ID",
						"required":    true,
						"type":        "integer",
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Successful operation",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/Receipt",
						},
					},
					"404": map[string]interface{}{
						"description": "Receipt not found",
					},
				},
				"security": []map[string]interface{}{
					{"BearerAuth": []string{}},
				},
			},
			"put": map[string]interface{}{
				"tags":        []string{"Receipts"},
				"summary":     "Update receipt",
				"description": "Update an existing receipt",
				"consumes":    []string{"application/json"},
				"produces":    []string{"application/json"},
				"parameters": []map[string]interface{}{
					{
						"name":        "id",
						"in":          "path",
						"description": "Receipt ID",
						"required":    true,
						"type":        "integer",
					},
					{
						"in":       "body",
						"name":     "body",
						"required": true,
						"schema": map[string]interface{}{
							"$ref": "#/definitions/ReceiptUpdate",
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Receipt updated successfully",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/ApiResponse",
						},
					},
				},
				"security": []map[string]interface{}{
					{"BearerAuth": []string{}},
				},
			},
			"delete": map[string]interface{}{
				"tags":        []string{"Receipts"},
				"summary":     "Delete receipt",
				"description": "Delete a receipt",
				"produces":    []string{"application/json"},
				"parameters": []map[string]interface{}{
					{
						"name":        "id",
						"in":          "path",
						"description": "Receipt ID",
						"required":    true,
						"type":        "integer",
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Receipt deleted successfully",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/ApiResponse",
						},
					},
				},
				"security": []map[string]interface{}{
					{"BearerAuth": []string{}},
				},
			},
		},
	}

	for path, definition := range receiptRoutes {
		paths[path] = definition
	}
}

func enhanceBusinessSchemas(swagger map[string]interface{}) {
	definitions, ok := swagger["definitions"].(map[string]interface{})
	if !ok {
		return
	}

	// Add Receipt schemas
	definitions["Receipt"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "integer",
				"description": "Receipt ID",
			},
			"receipt_number": map[string]interface{}{
				"type":        "string",
				"description": "Receipt number",
			},
			"payment_id": map[string]interface{}{
				"type":        "integer",
				"description": "Associated payment ID",
			},
			"date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "Receipt date",
			},
			"amount": map[string]interface{}{
				"type":        "number",
				"format":      "double",
				"description": "Receipt amount",
			},
			"notes": map[string]interface{}{
				"type":        "string",
				"description": "Receipt notes",
			},
			"created_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Creation timestamp",
			},
			"updated_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
			},
		},
		"required": []string{"receipt_number", "payment_id", "date", "amount"},
	}

	definitions["ReceiptCreate"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"payment_id": map[string]interface{}{
				"type":        "integer",
				"description": "Associated payment ID",
			},
			"date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "Receipt date",
			},
			"amount": map[string]interface{}{
				"type":        "number",
				"format":      "double",
				"description": "Receipt amount",
			},
			"notes": map[string]interface{}{
				"type":        "string",
				"description": "Receipt notes",
			},
		},
		"required": []string{"payment_id", "date", "amount"},
	}

	definitions["ReceiptUpdate"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"description": "Receipt date",
			},
			"amount": map[string]interface{}{
				"type":        "number",
				"format":      "double",
				"description": "Receipt amount",
			},
			"notes": map[string]interface{}{
				"type":        "string",
				"description": "Receipt notes",
			},
		},
	}

	// Enhance Sale schemas with PPN fields
	if saleSchema, exists := definitions["Sale"].(map[string]interface{}); exists {
		if properties, ok := saleSchema["properties"].(map[string]interface{}); ok {
			properties["ppn_rate"] = map[string]interface{}{
				"type":        "number",
				"format":      "double",
				"description": "PPN tax rate percentage",
				"example":     11.0,
			}
			properties["ppn_amount"] = map[string]interface{}{
				"type":        "number",
				"format":      "double",
				"description": "PPN tax amount",
			}
			properties["other_tax_amount"] = map[string]interface{}{
				"type":        "number",
				"format":      "double",
				"description": "Other tax amount",
			}
		}
	}

	if saleCreateSchema, exists := definitions["SaleCreate"].(map[string]interface{}); exists {
		if properties, ok := saleCreateSchema["properties"].(map[string]interface{}); ok {
			properties["ppn_rate"] = map[string]interface{}{
				"type":        "number",
				"format":      "double",
				"description": "PPN tax rate percentage",
				"example":     11.0,
			}
		}
	}

	// Enhance Payment schemas
	if paymentSchema, exists := definitions["Payment"].(map[string]interface{}); exists {
		if properties, ok := paymentSchema["properties"].(map[string]interface{}); ok {
			properties["payment_method"] = map[string]interface{}{
				"type":        "string",
				"description": "Payment method",
				"enum":        []string{"CASH", "BANK_TRANSFER", "CHEQUE", "CREDIT_CARD", "GIRO"},
			}
			properties["bank_name"] = map[string]interface{}{
				"type":        "string",
				"description": "Bank name for bank transfers",
			}
			properties["account_number"] = map[string]interface{}{
				"type":        "string",
				"description": "Bank account number",
			}
		}
	}

	if paymentCreateSchema, exists := definitions["PaymentCreate"].(map[string]interface{}); exists {
		if properties, ok := paymentCreateSchema["properties"].(map[string]interface{}); ok {
			properties["payment_method"] = map[string]interface{}{
				"type":        "string",
				"description": "Payment method",
				"enum":        []string{"CASH", "BANK_TRANSFER", "CHEQUE", "CREDIT_CARD", "GIRO"},
				"example":     "BANK_TRANSFER",
			}
		}
	}
}

func addCompleteContactCRUD(swagger map[string]interface{}) {
	paths, ok := swagger["paths"].(map[string]interface{})
	if !ok {
		return
	}

	// Enhance existing contacts endpoint
	if contactsPath, exists := paths["/contacts"].(map[string]interface{}); exists {
		// Add GET method parameters for filtering
		if getMethod, exists := contactsPath["get"].(map[string]interface{}); exists {
			getMethod["parameters"] = []map[string]interface{}{
				{
					"name":        "page",
					"in":          "query",
					"description": "Page number",
					"type":        "integer",
					"default":     1,
				},
				{
					"name":        "limit",
					"in":          "query",
					"description": "Number of items per page",
					"type":        "integer",
					"default":     10,
				},
				{
					"name":        "type",
					"in":          "query",
					"description": "Filter by contact type",
					"type":        "string",
					"enum":        []string{"CUSTOMER", "VENDOR", "EMPLOYEE"},
				},
				{
					"name":        "is_active",
					"in":          "query",
					"description": "Filter by active status",
					"type":        "boolean",
				},
			}
		}
	}
}

func fixContentTypesAndParsing(swagger map[string]interface{}) {
	paths, ok := swagger["paths"].(map[string]interface{})
	if !ok {
		return
	}

	// Ensure all endpoints have proper content types
	for _, pathDef := range paths {
		if pathObj, ok := pathDef.(map[string]interface{}); ok {
			for method, methodDef := range pathObj {
				if methodObj, ok := methodDef.(map[string]interface{}); ok {
					// Add produces and consumes to all methods
					if method == "get" || method == "post" || method == "put" || method == "delete" {
						if _, exists := methodObj["produces"]; !exists {
							methodObj["produces"] = []string{"application/json"}
						}
						if method == "post" || method == "put" {
							if _, exists := methodObj["consumes"]; !exists {
								methodObj["consumes"] = []string{"application/json"}
							}
						}
						
						// Ensure security is applied to all protected endpoints
						if _, exists := methodObj["security"]; !exists {
							methodObj["security"] = []map[string]interface{}{
								{"BearerAuth": []string{}},
							}
						}
					}
				}
			}
		}
	}
}

func generateFixedSwaggerUI() {
	swaggerHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Sistema Akuntansi API - Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    html, body, #swagger {
      height: 100%;
      margin: 0;
      padding: 0;
    }
    
    /* Remove all green elements and quick start banners */
    .swagger-ui .info .title small.version-stamp,
    .swagger-ui .info .title small,
    .swagger-ui .info .description .markdown p:first-child,
    .info .description .renderedMarkdown p:first-child,
    [style*="background-color: #1f8443"],
    [style*="background: #1f8443"],
    [style*="background-color: green"],
    [style*="background: green"],
    [class*="quickstart"],
    [class*="quick-start"],
    .swagger-ui .scheme-container {
      display: none !important;
    }
    
    /* Ensure clean appearance */
    .swagger-ui {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    }
    
    .swagger-ui .topbar {
      background-color: #f8f9fa;
      border-bottom: 1px solid #e9ecef;
    }
    
    /* Improve JSON display */
    .swagger-ui .response-col_status,
    .swagger-ui .response-col_links,
    .swagger-ui .response-col_description {
      white-space: pre-wrap;
      word-break: break-word;
    }
    
    /* Custom authentication styling */
    .swagger-ui .auth-wrapper {
      margin-bottom: 20px;
    }
  </style>
</head>
<body>
  <div id="swagger"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
  window.onload = function() {
    window.ui = SwaggerUIBundle({
      url: '/docs/swagger.json',
      dom_id: '#swagger',
      deepLinking: true,
      showExtensions: true,
      showCommonExtensions: true,
      defaultModelRendering: 'example',
      defaultModelExpandDepth: 2,
      defaultModelsExpandDepth: 2,
      validatorUrl: null,
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIBundle.presets.standalone
      ],
      layout: "StandaloneLayout",
      requestInterceptor: function(request) {
        // Ensure proper content-type headers
        if (!request.headers['Content-Type'] && (request.method === 'POST' || request.method === 'PUT')) {
          request.headers['Content-Type'] = 'application/json';
        }
        if (!request.headers['Accept']) {
          request.headers['Accept'] = 'application/json';
        }
        
        // Add CORS headers if needed
        request.headers['Access-Control-Allow-Origin'] = '*';
        
        return request;
      },
      responseInterceptor: function(response) {
        // Ensure proper JSON parsing
        if (response.headers['content-type'] && response.headers['content-type'].includes('application/json')) {
          try {
            if (typeof response.text === 'string' && response.text.length > 0) {
              // Validate JSON
              JSON.parse(response.text);
            }
          } catch (e) {
            console.warn('Response is not valid JSON:', response.text);
            // Try to fix common JSON issues
            if (response.text) {
              try {
                response.text = response.text.replace(/([{,]\s*)([a-zA-Z_$][a-zA-Z0-9_$]*)\s*:/g, '$1"$2":');
                JSON.parse(response.text);
              } catch (e2) {
                console.error('Could not fix JSON:', e2);
              }
            }
          }
        }
        return response;
      }
    });
    
    // Remove any remaining green elements after a delay
    setTimeout(function() {
      const elementsToHide = [
        '[style*="background-color: green"]',
        '[style*="background: green"]',
        '[class*="quickstart"]',
        '[class*="quick-start"]',
        '.info .title small'
      ];
      
      elementsToHide.forEach(selector => {
        const elements = document.querySelectorAll(selector);
        elements.forEach(el => {
          el.style.display = 'none';
        });
      });
    }, 2000);
  };
  </script>
</body>
</html>`

	err := ioutil.WriteFile("./docs/index.html", []byte(swaggerHTML), 0644)
	if err != nil {
		log.Printf("❌ Error writing Swagger HTML: %v", err)
	} else {
		log.Println("✅ Generated fixed Swagger UI HTML")
	}
}