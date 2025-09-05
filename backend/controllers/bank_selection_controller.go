package controllers

import (
	"fmt"
	"net/http"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BankSelectionController struct {
	cashBankRepo *repositories.CashBankRepository
}

func NewBankSelectionController(db *gorm.DB) *BankSelectionController {
	return &BankSelectionController{
		cashBankRepo: repositories.NewCashBankRepository(db),
	}
}

// GetBankAccountsForPurchase returns active bank accounts suitable for purchase payments
func (c *BankSelectionController) GetBankAccountsForPurchase(ctx *gin.Context) {
	// Get all active cash/bank accounts
	cashBanks, err := c.cashBankRepo.FindAll()
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch bank accounts",
		})
		return
	}
	
	// Filter and transform to simple format for dropdown
	type BankOption struct {
		ID       uint    `json:"id"`
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		Type     string  `json:"type"`
		Balance  float64 `json:"balance"`
		Currency string  `json:"currency"`
	}
	
	var bankOptions []BankOption
	for _, bank := range cashBanks {
		// Filter only active bank accounts
		if bank.IsActive && bank.Type == "BANK" {
			bankOptions = append(bankOptions, BankOption{
				ID:       bank.ID,
				Code:     bank.Code,
				Name:     bank.Name,
				Type:     bank.Type,
				Balance:  bank.Balance,
				Currency: bank.Currency,
			})
		}
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"data": bankOptions,
	})
}

// GetCashAccountsForPurchase returns cash accounts suitable for cash purchases
func (c *BankSelectionController) GetCashAccountsForPurchase(ctx *gin.Context) {
	cashBanks, err := c.cashBankRepo.FindAll()
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch cash accounts",
		})
		return
	}
	
	// Transform to simple format
	type CashOption struct {
		ID      uint    `json:"id"`
		Code    string  `json:"code"`
		Name    string  `json:"name"`
		Balance float64 `json:"balance"`
	}
	
	var cashOptions []CashOption
	for _, cash := range cashBanks {
		// Filter only active cash accounts
		if cash.IsActive && cash.Type == "CASH" {
			cashOptions = append(cashOptions, CashOption{
				ID:      cash.ID,
				Code:    cash.Code,
				Name:    cash.Name,
				Balance: cash.Balance,
			})
		}
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"data": cashOptions,
	})
}

// GetAllPaymentAccounts returns both cash and bank accounts
func (c *BankSelectionController) GetAllPaymentAccounts(ctx *gin.Context) {
	cashBanks, err := c.cashBankRepo.FindAll()
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch payment accounts",
		})
		return
	}
	
	// Group by type for better UX
	bankAccounts := []models.CashBank{}
	cashAccounts := []models.CashBank{}
	
	for _, account := range cashBanks {
		// Only include active accounts
		if account.IsActive {
			if account.Type == "BANK" {
				bankAccounts = append(bankAccounts, account)
			} else if account.Type == "CASH" {
				cashAccounts = append(cashAccounts, account)
			}
		}
	}
	
	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"bank_accounts": bankAccounts,
			"cash_accounts": cashAccounts,
		},
	})
}

// ValidateBankAccountForPurchase validates if a bank account can be used for purchase
func (c *BankSelectionController) ValidateBankAccountForPurchase(ctx *gin.Context) {
	bankIDStr := ctx.Param("bank_id")
	amountStr := ctx.Query("amount")
	
	if bankIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Bank account ID is required",
		})
		return
	}
	
	// Parse bank ID
	var bankID uint
	if _, err := fmt.Sscanf(bankIDStr, "%d", &bankID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid bank account ID",
		})
		return
	}
	
	// Get bank account
	cashBank, err := c.cashBankRepo.FindByID(bankID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Bank account not found",
		})
		return
	}
	
	validation := gin.H{
		"valid": true,
		"bank":  cashBank,
	}
	
	// Check balance if amount provided
	if amountStr != "" {
		var amount float64
		if _, err := fmt.Sscanf(amountStr, "%f", &amount); err == nil {
			if cashBank.Balance < amount {
				validation["valid"] = false
				validation["reason"] = "Insufficient balance"
				validation["available_balance"] = cashBank.Balance
				validation["required_amount"] = amount
			}
		}
	}
	
	ctx.JSON(http.StatusOK, validation)
}
