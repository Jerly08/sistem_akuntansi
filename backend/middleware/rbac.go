package middleware

import (
	"net/http"
	"strings"

	"app-sistem-akuntansi/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RBACManager struct {
	DB *gorm.DB
}

func NewRBACManager(db *gorm.DB) *RBACManager {
	return &RBACManager{DB: db}
}

// RequireRole checks if user has one of the required roles
func (rm *RBACManager) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found",
				"code":  "ROLE_NOT_FOUND",
			})
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		for _, role := range roles {
			if roleStr == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient role permissions",
			"code":  "INSUFFICIENT_ROLE",
			"required_roles": roles,
			"user_role": roleStr,
		})
		c.Abort()
	}
}

// RequirePermission checks if user has specific permission
func (rm *RBACManager) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found",
				"code":  "ROLE_NOT_FOUND",
			})
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		permissionName := resource + ":" + action

		// Check if user has the required permission
		hasPermission, err := rm.hasPermission(roleStr, permissionName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error checking permissions",
				"code":  "PERMISSION_CHECK_ERROR",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"code":  "INSUFFICIENT_PERMISSION",
				"required_permission": permissionName,
				"user_role": roleStr,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermissions checks if user has all specified permissions
func (rm *RBACManager) RequirePermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found",
				"code":  "ROLE_NOT_FOUND",
			})
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		
		for _, permission := range permissions {
			hasPermission, err := rm.hasPermission(roleStr, permission)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Error checking permissions",
					"code":  "PERMISSION_CHECK_ERROR",
				})
				c.Abort()
				return
			}

			if !hasPermission {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Insufficient permissions",
					"code":  "INSUFFICIENT_PERMISSION",
					"required_permissions": permissions,
					"missing_permission": permission,
					"user_role": roleStr,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireAnyPermission checks if user has at least one of the specified permissions
func (rm *RBACManager) RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found",
				"code":  "ROLE_NOT_FOUND",
			})
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		
		for _, permission := range permissions {
			hasPermission, err := rm.hasPermission(roleStr, permission)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Error checking permissions",
					"code":  "PERMISSION_CHECK_ERROR",
				})
				c.Abort()
				return
			}

			if hasPermission {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
			"code":  "INSUFFICIENT_PERMISSION",
			"required_any_of": permissions,
			"user_role": roleStr,
		})
		c.Abort()
	}
}

// RequireOwnershipOrRole checks if user owns the resource or has required role
func (rm *RBACManager) RequireOwnershipOrRole(userIDField string, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found",
				"code":  "USER_ID_NOT_FOUND",
			})
			c.Abort()
			return
		}

		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found",
				"code":  "ROLE_NOT_FOUND",
			})
			c.Abort()
			return
		}

		currentUserID := userID.(uint)
		roleStr := userRole.(string)

		// Check if user has one of the required roles
		for _, role := range roles {
			if roleStr == role {
				c.Next()
				return
			}
		}

		// Check ownership - get resource user ID from request
		resourceUserIDStr := c.Param(userIDField)
		if resourceUserIDStr == "" {
			resourceUserIDStr = c.Query(userIDField)
		}

		if resourceUserIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Resource user ID not provided",
				"code":  "RESOURCE_USER_ID_MISSING",
			})
			c.Abort()
			return
		}

		// Convert string to uint (simple conversion for demo)
		var resourceUserID uint
		if resourceUserIDStr != "" {
			// In production, use proper conversion with error handling
			resourceUserID = uint(parseUint(resourceUserIDStr))
		}

		if currentUserID == resourceUserID {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied: not owner and insufficient role",
			"code":  "ACCESS_DENIED",
		})
		c.Abort()
	}
}

// AdminOnly restricts access to admin users only
func (rm *RBACManager) AdminOnly() gin.HandlerFunc {
	return rm.RequireRole(models.RoleAdmin)
}

// FinanceOrAdmin restricts access to finance users and admins
func (rm *RBACManager) FinanceOrAdmin() gin.HandlerFunc {
	return rm.RequireRole(models.RoleAdmin, models.RoleFinance)
}

// DirectorOrAdmin restricts access to director and admin users
func (rm *RBACManager) DirectorOrAdmin() gin.HandlerFunc {
	return rm.RequireRole(models.RoleAdmin, models.RoleDirector)
}

// InventoryManagerOrAdmin restricts access to inventory manager and admin users
func (rm *RBACManager) InventoryManagerOrAdmin() gin.HandlerFunc {
	return rm.RequireRole(models.RoleAdmin, models.RoleInventoryManager)
}


// Custom permission checkers for specific resources

// CanReadFinancialData checks if user can read financial data
func (rm *RBACManager) CanReadFinancialData() gin.HandlerFunc {
	return rm.RequireAnyPermission(
		"accounts:read",
		"transactions:read",
		"reports:read",
	)
}

// CanModifyFinancialData checks if user can modify financial data
func (rm *RBACManager) CanModifyFinancialData() gin.HandlerFunc {
	return rm.RequireAnyPermission(
		"accounts:update",
		"transactions:create",
		"transactions:update",
	)
}

// CanAccessReports checks if user can access reports
func (rm *RBACManager) CanAccessReports() gin.HandlerFunc {
	return rm.RequireRole(
		models.RoleAdmin,
		models.RoleDirector,
		models.RoleFinance,
	)
}

// CanManageUsers checks if user can manage other users
func (rm *RBACManager) CanManageUsers() gin.HandlerFunc {
	return rm.RequirePermission("users", "manage")
}

// CanManageInventory checks if user can manage inventory
func (rm *RBACManager) CanManageInventory() gin.HandlerFunc {
	return rm.RequireRole(
		models.RoleAdmin,
		models.RoleInventoryManager,
	)
}

// Helper functions

func (rm *RBACManager) hasPermission(role, permissionName string) (bool, error) {
	var count int64
	
	// Split permission name to get resource and action
	parts := strings.Split(permissionName, ":")
	if len(parts) != 2 {
		return false, nil
	}

	resource := parts[0]
	action := parts[1]

	// Check if role has the specific permission
	err := rm.DB.Table("role_permissions rp").
		Joins("JOIN permissions p ON rp.permission_id = p.id").
		Where("rp.role = ? AND p.resource = ? AND p.action = ?", role, resource, action).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Simplified uint parsing (replace with proper implementation)
func parseUint(s string) int {
	// Simple conversion - in production use strconv.ParseUint with error handling
	result := 0
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		}
	}
	return result
}

// Permission constants for easier reference
const (
	// User permissions
	PermissionUsersRead   = "users:read"
	PermissionUsersCreate = "users:create"
	PermissionUsersUpdate = "users:update"
	PermissionUsersDelete = "users:delete"
	PermissionUsersManage = "users:manage"

	// Account permissions
	PermissionAccountsRead   = "accounts:read"
	PermissionAccountsCreate = "accounts:create"
	PermissionAccountsUpdate = "accounts:update"
	PermissionAccountsDelete = "accounts:delete"

	// Transaction permissions
	PermissionTransactionsRead   = "transactions:read"
	PermissionTransactionsCreate = "transactions:create"
	PermissionTransactionsUpdate = "transactions:update"
	PermissionTransactionsDelete = "transactions:delete"

	// Product permissions
	PermissionProductsRead   = "products:read"
	PermissionProductsCreate = "products:create"
	PermissionProductsUpdate = "products:update"
	PermissionProductsDelete = "products:delete"

	// Sales permissions
	PermissionSalesRead   = "sales:read"
	PermissionSalesCreate = "sales:create"
	PermissionSalesUpdate = "sales:update"
	PermissionSalesDelete = "sales:delete"

	// Purchase permissions
	PermissionPurchasesRead   = "purchases:read"
	PermissionPurchasesCreate = "purchases:create"
	PermissionPurchasesUpdate = "purchases:update"
	PermissionPurchasesDelete = "purchases:delete"

	// Report permissions
	PermissionReportsRead   = "reports:read"
	PermissionReportsCreate = "reports:create"
	PermissionReportsUpdate = "reports:update"
	PermissionReportsDelete = "reports:delete"

	// Budget permissions
	PermissionBudgetsRead   = "budgets:read"
	PermissionBudgetsCreate = "budgets:create"
	PermissionBudgetsUpdate = "budgets:update"
	PermissionBudgetsDelete = "budgets:delete"

	// Asset permissions
	PermissionAssetsRead   = "assets:read"
	PermissionAssetsCreate = "assets:create"
	PermissionAssetsUpdate = "assets:update"
	PermissionAssetsDelete = "assets:delete"

	// Contact permissions
	PermissionContactsRead   = "contacts:read"
	PermissionContactsCreate = "contacts:create"
	PermissionContactsUpdate = "contacts:update"
	PermissionContactsDelete = "contacts:delete"
)
