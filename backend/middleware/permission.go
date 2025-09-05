package middleware

import (
	"net/http"
	"app-sistem-akuntansi/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PermissionMiddleware struct {
	db *gorm.DB
}

func NewPermissionMiddleware(db *gorm.DB) *PermissionMiddleware {
	return &PermissionMiddleware{db: db}
}

// CheckModulePermission checks if user has specific permission for a module
func (pm *PermissionMiddleware) CheckModulePermission(module string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by JWT middleware)
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Convert user_id to uint
		var userID uint
		switch v := userIDInterface.(type) {
		case float64:
			userID = uint(v)
		case int:
			userID = uint(v)
		case uint:
			userID = v
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
			c.Abort()
			return
		}

		// Get user role from context
		roleInterface, _ := c.Get("role")
		role := ""
		if roleStr, ok := roleInterface.(string); ok {
			role = roleStr
		}

		// Check if user has specific permission in database
		var permission models.ModulePermissionRecord
		err := pm.db.Where("user_id = ? AND module = ?", userID, module).First(&permission).Error

		hasPermission := false
		
		if err == nil {
			// Permission record found, check specific action
			switch action {
			case "view":
				hasPermission = permission.CanView
			case "create":
				hasPermission = permission.CanCreate
			case "edit":
				hasPermission = permission.CanEdit
			case "delete":
				hasPermission = permission.CanDelete
			case "approve":
				hasPermission = permission.CanApprove
			case "export":
				hasPermission = permission.CanExport
			default:
				hasPermission = false
			}
		} else if err == gorm.ErrRecordNotFound {
			// No custom permission, use default based on role
			defaultPerms := models.GetDefaultPermissions(role)
			if modPerm, ok := defaultPerms[module]; ok {
				switch action {
				case "view":
					hasPermission = modPerm.CanView
				case "create":
					hasPermission = modPerm.CanCreate
				case "edit":
					hasPermission = modPerm.CanEdit
				case "delete":
					hasPermission = modPerm.CanDelete
				case "approve":
					hasPermission = modPerm.CanApprove
				case "export":
					hasPermission = modPerm.CanExport
				}
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You don't have permission to " + action + " " + module,
				"required_permission": action,
				"module": module,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Convenience methods for common permissions
func (pm *PermissionMiddleware) CanView(module string) gin.HandlerFunc {
	return pm.CheckModulePermission(module, "view")
}

func (pm *PermissionMiddleware) CanCreate(module string) gin.HandlerFunc {
	return pm.CheckModulePermission(module, "create")
}

func (pm *PermissionMiddleware) CanEdit(module string) gin.HandlerFunc {
	return pm.CheckModulePermission(module, "edit")
}

func (pm *PermissionMiddleware) CanDelete(module string) gin.HandlerFunc {
	return pm.CheckModulePermission(module, "delete")
}

func (pm *PermissionMiddleware) CanApprove(module string) gin.HandlerFunc {
	return pm.CheckModulePermission(module, "approve")
}

func (pm *PermissionMiddleware) CanExport(module string) gin.HandlerFunc {
	return pm.CheckModulePermission(module, "export")
}
