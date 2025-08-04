package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/utils"
	"github.com/gin-gonic/gin"
)

// AccountHandler handles account-related operations
type AccountHandler struct {
	repo repositories.AccountRepository
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(repo repositories.AccountRepository) *AccountHandler {
	return &AccountHandler{
		repo: repo,
	}
}

// CreateAccount creates a new account
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req models.AccountCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appError := utils.NewBadRequestError("Invalid request payload")
		c.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	account, err := h.repo.Create(c.Request.Context(), &req)
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to create account", err)
			c.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": account})
}

// GetAccount gets a single account by code
func (h *AccountHandler) GetAccount(c *gin.Context) {
	code := c.Param("code")
	
	account, err := h.repo.FindByCode(c.Request.Context(), code)
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to get account", err)
			c.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": account})
}

// UpdateAccount updates an account
func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	code := c.Param("code")
	var req models.AccountUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appError := utils.NewBadRequestError("Invalid request payload")
		c.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	account, err := h.repo.Update(c.Request.Context(), code, &req)
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to update account", err)
			c.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": account})
}

// DeleteAccount deletes an account
func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	code := c.Param("code")

	err := h.repo.Delete(c.Request.Context(), code)
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to delete account", err)
			c.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

// ListAccounts lists all accounts with optional filtering
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	accountType := c.Query("type")
	
	var accounts []models.Account
	var err error
	
	if accountType != "" {
		accounts, err = h.repo.FindByType(c.Request.Context(), accountType)
	} else {
		accounts, err = h.repo.FindAll(c.Request.Context())
	}
	
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to retrieve accounts", err)
			c.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": accounts, "count": len(accounts)})
}

// GetAccountHierarchy gets account hierarchy tree
func (h *AccountHandler) GetAccountHierarchy(c *gin.Context) {
	accounts, err := h.repo.GetHierarchy(c.Request.Context())
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to get account hierarchy", err)
			c.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": accounts})
}

// GetBalanceSummary gets balance summary by account type
func (h *AccountHandler) GetBalanceSummary(c *gin.Context) {
	summary, err := h.repo.GetBalanceSummary(c.Request.Context())
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to get balance summary", err)
			c.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// ImportAccounts handles bulk import of accounts from CSV
func (h *AccountHandler) ImportAccounts(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		appError := utils.NewBadRequestError("No file uploaded")
		c.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		appError := utils.NewBadRequestError("Failed to read CSV file")
		c.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	if len(records) < 2 {
		appError := utils.NewBadRequestError("CSV file must contain header and at least one data row")
		c.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	// Skip header row
	var importRequests []models.AccountImportRequest
	for i, record := range records[1:] {
		if len(record) < 3 {
			appError := utils.NewBadRequestError(fmt.Sprintf("Row %d: insufficient columns", i+2))
			c.JSON(appError.StatusCode, appError.ToErrorResponse(""))
			return
		}

		openingBalance := 0.0
		if len(record) > 4 && record[4] != "" {
			if balance, err := strconv.ParseFloat(record[4], 64); err == nil {
				openingBalance = balance
			}
		}

		importReq := models.AccountImportRequest{
			Code:           strings.TrimSpace(record[0]),
			Name:           strings.TrimSpace(record[1]),
			Type:           models.AccountType(strings.TrimSpace(strings.ToUpper(record[2]))),
			Description:    "",
			OpeningBalance: openingBalance,
		}

		if len(record) > 3 && record[3] != "" {
			importReq.ParentCode = strings.TrimSpace(record[3])
		}
		if len(record) > 5 && record[5] != "" {
			importReq.Description = strings.TrimSpace(record[5])
		}

		importRequests = append(importRequests, importReq)
	}

	err = h.repo.BulkImport(c.Request.Context(), importRequests)
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to import accounts", err)
			c.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Accounts imported successfully",
		"count":   len(importRequests),
	})
}

