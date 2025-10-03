package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
)

// SwaggerPatch represents the structure for fixing Swagger documentation
type SwaggerPatch struct {
	OpenAPI string             `json:"openapi"`
	Info    SwaggerInfo        `json:"info"`
	Servers []SwaggerServer    `json:"servers"`
	Paths   map[string]interface{} `json:"paths"`
	Components SwaggerComponents  `json:"components"`
}

type SwaggerInfo struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Version     string          `json:"version"`
	Contact     SwaggerContact  `json:"contact"`
	License     SwaggerLicense  `json:"license"`
}

type SwaggerContact struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Email string `json:"email"`
}

type SwaggerLicense struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type SwaggerServer struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type SwaggerComponents struct {
	SecuritySchemes map[string]interface{} `json:"securitySchemes"`
	Schemas         map[string]interface{} `json:"schemas"`
}

func main() {
	log.Println("🔧 Starting Swagger API fixes...")

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

	// Fix 1: Update paths to match actual route structure
	log.Println("🔧 Fixing route paths...")
	fixRoutePaths(swagger)

	// Fix 2: Add missing CRUD endpoints for contacts
	log.Println("🔧 Adding missing contact endpoints...")
	addMissingContactEndpoints(swagger)

	// Fix 3: Add missing receipt endpoints
	log.Println("🔧 Adding missing receipt endpoints...")
	addReceiptEndpoints(swagger)

	// Fix 4: Enhance payment and sales schemas with tax fields
	log.Println("🔧 Enhancing schemas with tax and payment fields...")
	enhanceSchemas(swagger)

	// Fix 5: Add proper content-type headers
	log.Println("🔧 Adding proper content-type headers...")
	fixContentTypes(swagger)

	// Fix 6: Add security configurations to prevent auth issues
	log.Println("🔧 Adding security configurations...")
	addSecurityConfig(swagger)

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

	log.Println("✅ Swagger fixes applied successfully!")

	// Generate HTML fix for the green Quick Start overlay
	log.Println("🔧 Generating Swagger UI fix for green overlay...")
	generateSwaggerUIFix()

	log.Println("🎉 All fixes completed!")
}

func fixRoutePaths(swagger map[string]interface{}) {
	paths, ok := swagger["paths"].(map[string]interface{})
	if !ok {
		return
	}

	// Fix cash-bank routes
	cashBankRoutes := []string{
		"/api/v1/cash-bank/accounts",
		"/api/v1/cash-bank/accounts/{id}",
		"/api/v1/cash-bank/accounts/{id}/transactions",
		"/api/v1/cash-bank/transactions/deposit",
		"/api/v1/cash-bank/transactions/withdrawal",
		"/api/v1/cash-bank/transactions/transfer",
		"/api/v1/cash-bank/reports/balance-summary",
		"/api/v1/cash-bank/reports/payment-accounts",
	}

	// Add cash-bank routes to paths
	for _, route := range cashBankRoutes {
		if !strings.Contains(route, "{id}") {
			// GET and POST routes
			paths[route] = map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"CashBank"},
					"summary":     "Get cash/bank data",
					"description": "Retrieve cash/bank information",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful operation",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
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
													"$ref": "#/components/schemas/CashBank",
												},
											},
										},
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
						"500": map[string]interface{}{
							"description": "Internal server error",
						},
					},
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
				},
				"post": map[string]interface{}{
					"tags":        []string{"CashBank"},
					"summary":     "Create cash/bank entry",
					"description": "Create new cash/bank entry",
					"requestBody": map[string]interface{}{
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/CashBankCreate",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Created successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ApiResponse",
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Bad request",
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
					"security": []map[string]interface{}{
						{"BearerAuth": []string{}},
					},
				},
			}
		}
	}

	// Fix payment routes
	paymentRoutes := []string{
		"/api/v1/ssot-payments",
		"/api/v1/ssot-payments/{id}",
		"/api/v1/ultra-fast-payments",
	}

	for _, route := range paymentRoutes {
		paths[route] = map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Payments"},
				"summary":     "Get payment data",
				"description": "Retrieve payment information",
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Successful operation",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/ApiResponse",
								},
							},
						},
					},
				},
				"security": []map[string]interface{}{
					{"BearerAuth": []string{}},
				},
			},
		}
	}
}

func addMissingContactEndpoints(swagger map[string]interface{}) {
	paths, ok := swagger["paths"].(map[string]interface{})
	if !ok {
		return
	}

	// Add missing contact endpoints
	contactEndpoints := map[string]interface{}{
		"/api/v1/contacts": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Contacts"},
				"summary":     "Get all contacts",
				"description": "Retrieve a list of all contacts (customers, vendors, employees)",
				"parameters": []map[string]interface{}{
					{
						"name":        "page",
						"in":          "query",
						"description": "Page number",
						"schema": map[string]interface{}{
							"type":    "integer",
							"default": 1,
						},
					},
					{
						"name":        "limit",
						"in":          "query",
						"description": "Number of items per page",
						"schema": map[string]interface{}{
							"type":    "integer",
							"default": 10,
						},
					},
					{
						"name":        "type",
						"in":          "query",
						"description": "Filter by contact type",
						"schema": map[string]interface{}{
							"type": "string",
							"enum": []string{"CUSTOMER", "VENDOR", "EMPLOYEE"},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Successful operation",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
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
												"$ref": "#/components/schemas/Contact",
											},
										},
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
				"tags":        []string{"Contacts"},
				"summary":     "Create new contact",
				"description": "Create a new contact (customer, vendor, or employee)",
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ContactCreate",
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{
						"description": "Contact created successfully",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/ApiResponse",
								},
							},
						},
					},
				},
				"security": []map[string]interface{}{
					{"BearerAuth": []string{}},
				},
			},
		},
	}

	for endpoint, definition := range contactEndpoints {
		paths[endpoint] = definition
	}
}

func addReceiptEndpoints(swagger map[string]interface{}) {
	paths, ok := swagger["paths"].(map[string]interface{})
	if !ok {
		return
	}

	// Add receipt endpoints
	receiptEndpoints := map[string]interface{}{
		"/api/v1/receipts": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Receipts"},
				"summary":     "Get all receipts",
				"description": "Retrieve a list of all receipts",
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Successful operation",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/ApiResponse",
								},
							},
						},
					},
				},
				"security": []map[string]interface{}{
					{"BearerAuth": []string{}},
				},
			},
		},
		"/api/v1/receipts/{id}": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Receipts"},
				"summary":     "Get receipt by ID",
				"description": "Retrieve a specific receipt by its ID",
				"parameters": []map[string]interface{}{
					{
						"name":        "id",
						"in":          "path",
						"description": "Receipt ID",
						"required":    true,
						"schema": map[string]interface{}{
							"type": "integer",
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Successful operation",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/ApiResponse",
								},
							},
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
		},
	}

	for endpoint, definition := range receiptEndpoints {
		paths[endpoint] = definition
	}
}

func enhanceSchemas(swagger map[string]interface{}) {
	definitions, ok := swagger["definitions"].(map[string]interface{})
	if !ok {
		return
	}

	// Enhance Sale schema with tax fields
	if saleSchema, exists := definitions["Sale"].(map[string]interface{}); exists {
		if properties, ok := saleSchema["properties"].(map[string]interface{}); ok {
			properties["ppn_rate"] = map[string]interface{}{
				"description": "PPN tax rate percentage",
				"type":        "number",
				"format":      "double",
				"example":     11.0,
			}
			properties["ppn_amount"] = map[string]interface{}{
				"description": "PPN tax amount",
				"type":        "number",
				"format":      "double",
			}
			properties["other_tax_amount"] = map[string]interface{}{
				"description": "Other tax amount",
				"type":        "number",
				"format":      "double",
			}
		}
	}

	// Enhance SaleCreate schema with tax fields
	if saleCreateSchema, exists := definitions["SaleCreate"].(map[string]interface{}); exists {
		if properties, ok := saleCreateSchema["properties"].(map[string]interface{}); ok {
			properties["ppn_rate"] = map[string]interface{}{
				"description": "PPN tax rate percentage",
				"type":        "number",
				"format":      "double",
				"example":     11.0,
			}
		}
	}

	// Enhance Payment schema with cash bank fields
	if paymentSchema, exists := definitions["Payment"].(map[string]interface{}); exists {
		if properties, ok := paymentSchema["properties"].(map[string]interface{}); ok {
			properties["bank_name"] = map[string]interface{}{
				"description": "Bank name for bank transfers",
				"type":        "string",
			}
			properties["account_number"] = map[string]interface{}{
				"description": "Bank account number",
				"type":        "string",
			}
			properties["payment_method"] = map[string]interface{}{
				"description": "Payment method",
				"type":        "string",
				"enum":        []string{"CASH", "BANK_TRANSFER", "CHEQUE", "CREDIT_CARD", "GIRO"},
			}
		}
	}

	// Enhance PaymentCreate schema
	if paymentCreateSchema, exists := definitions["PaymentCreate"].(map[string]interface{}); exists {
		if properties, ok := paymentCreateSchema["properties"].(map[string]interface{}); ok {
			properties["payment_method"] = map[string]interface{}{
				"description": "Payment method",
				"type":        "string",
				"enum":        []string{"CASH", "BANK_TRANSFER", "CHEQUE", "CREDIT_CARD", "GIRO"},
				"example":     "BANK_TRANSFER",
			}
		}
	}

	// Add Receipt schema
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
			"payment_id": map[string]interface{}{
				"type":        "integer",
				"description": "Associated payment ID",
			},
			"notes": map[string]interface{}{
				"type":        "string",
				"description": "Receipt notes",
			},
		},
		"required": []string{"receipt_number", "date", "amount", "payment_id"},
	}
}

func fixContentTypes(swagger map[string]interface{}) {
	// Ensure all endpoints return proper JSON content-type
	paths, ok := swagger["paths"].(map[string]interface{})
	if !ok {
		return
	}

	for _, pathDef := range paths {
		if pathObj, ok := pathDef.(map[string]interface{}); ok {
			for method, methodDef := range pathObj {
				if methodObj, ok := methodDef.(map[string]interface{}); ok {
					// Add produces field for Swagger 2.0
					if method == "get" || method == "post" || method == "put" || method == "delete" {
						methodObj["produces"] = []string{"application/json"}
						methodObj["consumes"] = []string{"application/json"}
					}
				}
			}
		}
	}
}

func addSecurityConfig(swagger map[string]interface{}) {
	// Ensure security definitions are proper
	swagger["securityDefinitions"] = map[string]interface{}{
		"BearerAuth": map[string]interface{}{
			"type":        "apiKey",
			"name":        "Authorization",
			"in":          "header",
			"description": "JWT token for authentication. Format: Bearer {token}",
		},
	}

	// Add global security
	swagger["security"] = []map[string]interface{}{
		{
			"BearerAuth": []string{},
		},
	}
}

func generateSwaggerUIFix() {
	// Generate a fixed Swagger UI HTML that removes the green overlay
	swaggerHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI - Sistema Akuntansi API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    html, body, #swagger {
      height: 100%;
      margin: 0;
      padding: 0;
    }
    
    /* Remove the green Quick Start overlay */
    .swagger-ui .info .title small.version-stamp,
    .swagger-ui .info .title small,
    .swagger-ui .info .description .markdown p:first-child,
    .info .description .renderedMarkdown p:first-child {
      display: none !important;
    }
    
    /* Hide any green overlay elements */
    [style*="background-color: #1f8443"],
    [style*="background: #1f8443"],
    [class*="quickstart"],
    [class*="quick-start"] {
      display: none !important;
    }
    
    /* Fix any green background elements */
    .swagger-ui .scheme-container {
      background: #ffffff !important;
    }
    
    /* Ensure proper JSON rendering */
    .swagger-ui .response-col_status,
    .swagger-ui .response-col_links,
    .swagger-ui .response-col_description {
      white-space: pre-wrap;
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
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIBundle.presets.standalone
      ],
      layout: "StandaloneLayout",
      deepLinking: true,
      showExtensions: true,
      showCommonExtensions: true,
      defaultModelRendering: 'example',
      defaultModelExpandDepth: 2,
      defaultModelsExpandDepth: 2,
      validatorUrl: null,
      // Remove the green banner/overlay
      plugins: [
        SwaggerUIBundle.plugins.DownloadUrl
      ],
      requestInterceptor: function(request) {
        // Fix JSON parsing issues by ensuring proper content-type
        if (request.headers['Content-Type'] === undefined) {
          request.headers['Content-Type'] = 'application/json';
        }
        if (request.headers['Accept'] === undefined) {
          request.headers['Accept'] = 'application/json';
        }
        return request;
      },
      responseInterceptor: function(response) {
        // Fix JSON parsing by ensuring proper content-type handling
        try {
          if (typeof response.text === 'string' && response.text.length > 0) {
            JSON.parse(response.text);
          }
        } catch (e) {
          console.warn('Response is not valid JSON:', response.text);
        }
        return response;
      }
    });
    
    // Additional fix: Remove any remaining green elements after UI loads
    setTimeout(function() {
      const greenElements = document.querySelectorAll('[style*="background"], [class*="green"], [class*="success"], .info .title small');
      greenElements.forEach(el => {
        if (el.style.backgroundColor && el.style.backgroundColor.includes('green')) {
          el.style.display = 'none';
        }
      });
    }, 1000);
  };
  </script>
</body>
</html>`

	err := ioutil.WriteFile("./docs/index.html", []byte(swaggerHTML), 0644)
	if err != nil {
		log.Printf("❌ Error writing Swagger HTML fix: %v", err)
	} else {
		log.Println("✅ Generated fixed Swagger UI HTML")
	}
}