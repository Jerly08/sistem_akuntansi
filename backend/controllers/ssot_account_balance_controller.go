package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SSOTAccountBalanceRow is a lightweight DTO for SSOT-derived balances
type SSOTAccountBalanceRow struct {
	AccountID   uint    `json:"account_id"`
	AccountCode string  `json:"account_code"`
	AccountName string  `json:"account_name"`
	AccountType string  `json:"account_type"`
	DebitTotal  float64 `json:"debit_total"`
	CreditTotal float64 `json:"credit_total"`
	NetBalance  float64 `json:"net_balance"`
}

// SSOTAccountBalanceController exposes SSOT account balances for COA sync
type SSOTAccountBalanceController struct {
	db *gorm.DB
}

func NewSSOTAccountBalanceController(db *gorm.DB) *SSOTAccountBalanceController {
	return &SSOTAccountBalanceController{db: db}
}

// GetSSOTAccountBalances returns net balances per account from SSOT (posted up to as_of_date)
// GET /api/v1/ssot-reports/account-balances?as_of_date=YYYY-MM-DD
func (ctl *SSOTAccountBalanceController) GetSSOTAccountBalances(c *gin.Context) {
	asOf := c.DefaultQuery("as_of_date", time.Now().Format("2006-01-02"))
	if _, err := time.Parse("2006-01-02", asOf); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid as_of_date, expected YYYY-MM-DD"})
		return
	}

	query := `
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			COALESCE(SUM(ujl.debit_amount), 0) as debit_total,
			COALESCE(SUM(ujl.credit_amount), 0) as credit_total,
			CASE 
				WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
					COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
				ELSE 
					COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
			END as net_balance
		FROM accounts a
		LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
		LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		WHERE ((uje.status = 'POSTED' AND uje.entry_date <= ?) OR uje.status IS NULL)
		  AND COALESCE(a.is_header, false) = false
		GROUP BY a.id, a.code, a.name, a.type
		ORDER BY a.code
	`

	rows := []SSOTAccountBalanceRow{}
	if err := ctl.db.Raw(query, asOf).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query SSOT balances", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    rows,
		"as_of":   asOf,
		"source":  "SSOT",
	})
}