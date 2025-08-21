package controllers

import (
	"net/http"
	"strconv"
	"app-sistem-akuntansi/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	db *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{db: db}
}

// GetUsers retrieves users with optional filters
func (uc *UserController) GetUsers(c *gin.Context) {
	var users []models.User
	query := uc.db.Model(&models.User{})

	// Apply filters
	if status := c.Query("status"); status != "" {
		if status == "active" {
			query = query.Where("is_active = ?", true)
		} else if status == "inactive" {
			query = query.Where("is_active = ?", false)
		}
	}

	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit

	// Get total count
	var total int64
	query.Count(&total)

	// Apply pagination
	query = query.Offset(offset).Limit(limit)

	// Execute query (exclude sensitive fields)
	if err := query.Select("id, username, email, role, first_name, last_name, phone, address, department, position, hire_date, is_active, last_login_at, created_at, updated_at").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Calculate full name for each user
	for i := range users {
		fullName := ""
		if users[i].FirstName != "" {
			fullName = users[i].FirstName
		}
		if users[i].LastName != "" {
			if fullName != "" {
				fullName += " " + users[i].LastName
			} else {
				fullName = users[i].LastName
			}
		}
		if fullName == "" {
			fullName = users[i].Username
		}
		// Add full_name to the response (we'll use a map)
	}

	// Convert to response format with full_name
	response := make([]map[string]interface{}, len(users))
	for i, user := range users {
		fullName := ""
		if user.FirstName != "" {
			fullName = user.FirstName
		}
		if user.LastName != "" {
			if fullName != "" {
				fullName += " " + user.LastName
			} else {
				fullName = user.LastName
			}
		}
		if fullName == "" {
			fullName = user.Username
		}

		response[i] = map[string]interface{}{
			"id":            user.ID,
			"username":      user.Username,
			"email":         user.Email,
			"role":          user.Role,
			"first_name":    user.FirstName,
			"last_name":     user.LastName,
			"full_name":     fullName,
			"phone":         user.Phone,
			"address":       user.Address,
			"department":    user.Department,
			"position":      user.Position,
			"hire_date":     user.HireDate,
			"is_active":     user.IsActive,
			"last_login_at": user.LastLoginAt,
			"created_at":    user.CreatedAt,
			"updated_at":    user.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
		"pagination": gin.H{
			"page":     page,
			"limit":    limit,
			"total":    total,
			"pages":    (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetUser retrieves a single user by ID
func (uc *UserController) GetUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := uc.db.Select("id, username, email, role, first_name, last_name, phone, address, department, position, hire_date, is_active, last_login_at, created_at, updated_at").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Calculate full name
	fullName := ""
	if user.FirstName != "" {
		fullName = user.FirstName
	}
	if user.LastName != "" {
		if fullName != "" {
			fullName += " " + user.LastName
		} else {
			fullName = user.LastName
		}
	}
	if fullName == "" {
		fullName = user.Username
	}

	response := map[string]interface{}{
		"id":            user.ID,
		"username":      user.Username,
		"email":         user.Email,
		"role":          user.Role,
		"first_name":    user.FirstName,
		"last_name":     user.LastName,
		"full_name":     fullName,
		"phone":         user.Phone,
		"address":       user.Address,
		"department":    user.Department,
		"position":      user.Position,
		"hire_date":     user.HireDate,
		"is_active":     user.IsActive,
		"last_login_at": user.LastLoginAt,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}
