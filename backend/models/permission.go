package models

import (
	"time"
	"gorm.io/gorm"
)

// ModulePermissionRecord represents a specific permission for modules
type ModulePermissionRecord struct {
	ID          uint           `json:"id" gorm:"primaryKey;table:module_permissions"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Module      string         `json:"module" gorm:"not null;size:50;index"` // accounts, products, contacts, assets, sales, purchases, payments, cash_bank
	CanView     bool           `json:"can_view" gorm:"default:false"`
	CanCreate   bool           `json:"can_create" gorm:"default:false"`
	CanEdit     bool           `json:"can_edit" gorm:"default:false"`
	CanDelete   bool           `json:"can_delete" gorm:"default:false"`
	CanApprove  bool           `json:"can_approve" gorm:"default:false"`
	CanExport   bool           `json:"can_export" gorm:"default:false"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// UserPermission is a simplified structure for API responses
type UserPermission struct {
	UserID      uint                    `json:"user_id"`
	Username    string                  `json:"username"`
	Email       string                  `json:"email"`
	Role        string                  `json:"role"`
	Permissions map[string]*ModulePermission `json:"permissions"`
}

// ModulePermission represents permissions for a specific module
type ModulePermission struct {
	CanView    bool `json:"can_view"`
	CanCreate  bool `json:"can_create"`
	CanEdit    bool `json:"can_edit"`
	CanDelete  bool `json:"can_delete"`
	CanApprove bool `json:"can_approve"`
	CanExport  bool `json:"can_export"`
}

// GetDefaultPermissions returns default permissions based on role
func GetDefaultPermissions(role string) map[string]*ModulePermission {
	permissions := make(map[string]*ModulePermission)
	modules := []string{"accounts", "products", "contacts", "assets", "sales", "purchases", "payments", "cash_bank"}
	
	switch role {
	case "admin":
		// Admin has full access to everything
		for _, module := range modules {
			permissions[module] = &ModulePermission{
				CanView:    true,
				CanCreate:  true,
				CanEdit:    true,
				CanDelete:  true,
				CanApprove: true,
				CanExport:  true,
			}
		}
	case "finance", "finance_manager":
		// Finance and Finance Manager have full access to financial modules
		financialModules := []string{"accounts", "payments", "cash_bank", "sales", "purchases"}
		for _, module := range modules {
			if contains(financialModules, module) {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    true,
					CanDelete:  false,
					CanApprove: true,
					CanExport:  true,
				}
			} else {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
				}
			}
		}
	case "inventory_manager":
		// Inventory manager has access to inventory-related modules
		inventoryModules := []string{"products", "purchases", "sales"}
		for _, module := range modules {
			if contains(inventoryModules, module) {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    true,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  true,
				}
			} else if module == "contacts" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    true,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
				}
			} else {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
				}
			}
		}
	case "employee":
		// Employee has limited access
		for _, module := range modules {
			if module == "contacts" || module == "products" {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
				}
			} else {
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
				}
			}
		}
	case "director":
		// Director has view and approve access
		for _, module := range modules {
			permissions[module] = &ModulePermission{
				CanView:    true,
				CanCreate:  false,
				CanEdit:    false,
				CanDelete:  false,
				CanApprove: true,
				CanExport:  true,
			}
		}
	default:
		// Default no permissions
		for _, module := range modules {
			permissions[module] = &ModulePermission{
				CanView:    false,
				CanCreate:  false,
				CanEdit:    false,
				CanDelete:  false,
				CanApprove: false,
				CanExport:  false,
			}
		}
	}
	
	return permissions
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
