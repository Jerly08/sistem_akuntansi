package models

import (
	"time"
	"gorm.io/gorm"
)

// ModulePermissionRecord represents a specific permission for modules
type ModulePermissionRecord struct {
	ID          uint           `json:"id" gorm:"primaryKey;table:module_permissions"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Module      string         `json:"module" gorm:"not null;size:50;index"` // accounts, products, contacts, assets, sales, purchases, payments, cash_bank, settings
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
	modules := []string{"accounts", "products", "contacts", "assets", "sales", "purchases", "payments", "cash_bank", "reports", "settings"}
	
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
			} else if module == "settings" {
				// Finance roles need settings access for invoice types and financial configuration
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,  // Can create invoice types
					CanEdit:    true,  // Can edit invoice types
					CanDelete:  false, // Cannot delete settings for safety
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
		// Inventory manager has comprehensive access to inventory and related operations
		coreInventoryModules := []string{"products", "purchases", "sales"}
		supportingModules := []string{"contacts", "assets", "reports"}
		financialSupportModules := []string{"accounts", "payments", "cash_bank"}
		
		for _, module := range modules {
			if contains(coreInventoryModules, module) {
				// Full access to core inventory modules
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    true,
					CanDelete:  false, // Safety: no delete permission
					CanApprove: false, // Purchase approvals handled by finance/director
					CanExport:  true,
				}
			} else if contains(supportingModules, module) {
				// Good access to supporting modules (contacts for vendors/customers, assets for inventory items, reports for analytics)
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,
					CanEdit:    true,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  true, // Can export reports and asset lists
				}
			} else if contains(financialSupportModules, module) {
				// Limited financial access - can create entries for inventory operations but cannot approve
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,  // Can create expense accounts, payments for purchases
					CanEdit:    false, // Cannot edit financial records (safety)
					CanDelete:  false,
					CanApprove: false, // Financial approvals remain with finance team
					CanExport:  true,  // Can export for reporting
				}
			} else {
				// View-only access to other modules
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
			} else if module == "accounts" {
				// Employee needs view access to accounts for purchase form dropdowns
				permissions[module] = &ModulePermission{
					CanView:    true,  // Essential for purchase forms
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: false,
					CanExport:  false,
				}
			} else if module == "purchases" {
				// Employee should be able to create purchases
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,  // Employees can create purchase requests
					CanEdit:    true,  // Can edit their own purchases
					CanDelete:  false, // Cannot delete purchases
					CanApprove: false, // Cannot approve purchases
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
		// Director has view, approve, and limited create/edit access
		for _, module := range modules {
			if module == "purchases" || module == "sales" || module == "payments" || module == "cash_bank" {
				// Directors need create/edit access for purchases to create receipts,
				// and for sales/payments for operational oversight
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,  // ✅ Allow creating for operational modules
					CanEdit:    true,  // ✅ Allow editing for operational modules (needed for receipts)
					CanDelete:  false, // Still no delete access for safety
					CanApprove: true,
					CanExport:  true,
				}
			} else if module == "settings" {
				// Directors need settings access for system configuration and invoice types
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  true,  // Can create invoice types and settings
					CanEdit:    true,  // Can edit system settings
					CanDelete:  false, // Cannot delete settings for safety
					CanApprove: true,
					CanExport:  true,
				}
			} else {
				// For other modules, keep view/approve only access
				permissions[module] = &ModulePermission{
					CanView:    true,
					CanCreate:  false,
					CanEdit:    false,
					CanDelete:  false,
					CanApprove: true,
					CanExport:  true,
				}
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
