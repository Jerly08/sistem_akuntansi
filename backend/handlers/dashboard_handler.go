package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	// Add services here if needed
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) GetDashboardAnalytics(c *gin.Context) {
	// Dummy data for now
	analytics := gin.H{
		"totalSales":         125000000,
		"totalPurchases":     85000000,
		"accountsReceivable": 25000000,
		"accountsPayable":    15000000,
		"monthlySales": []gin.H{
			{"month": "Jan", "value": 8500000},
			{"month": "Feb", "value": 9200000},
			{"month": "Mar", "value": 11800000},
			{"month": "Apr", "value": 10500000},
			{"month": "May", "value": 12200000},
			{"month": "Jun", "value": 15800000},
			{"month": "Jul", "value": 13500000},
		},
		"monthlyPurchases": []gin.H{
			{"month": "Jan", "value": 6500000},
			{"month": "Feb", "value": 7200000},
			{"month": "Mar", "value": 8800000},
			{"month": "Apr", "value": 8100000},
			{"month": "May", "value": 9200000},
			{"month": "Jun", "value": 11800000},
			{"month": "Jul", "value": 10300000},
		},
		"cashFlow": []gin.H{
			{"month": "Jan", "inflow": 8500000, "outflow": 6500000, "balance": 2000000},
			{"month": "Feb", "inflow": 9200000, "outflow": 7200000, "balance": 2000000},
			{"month": "Mar", "inflow": 11800000, "outflow": 8800000, "balance": 3000000},
			{"month": "Apr", "inflow": 10500000, "outflow": 8100000, "balance": 2400000},
			{"month": "May", "inflow": 12200000, "outflow": 9200000, "balance": 3000000},
			{"month": "Jun", "inflow": 15800000, "outflow": 11800000, "balance": 4000000},
			{"month": "Jul", "inflow": 13500000, "outflow": 10300000, "balance": 3200000},
		},
		"topAccounts": []gin.H{
			{"name": "Kas", "balance": 45000000, "type": "Asset"},
			{"name": "Bank BCA", "balance": 125000000, "type": "Asset"},
			{"name": "Piutang Dagang", "balance": 25000000, "type": "Asset"},
			{"name": "Persediaan", "balance": 75000000, "type": "Asset"},
			{"name": "Utang Dagang", "balance": 15000000, "type": "Liability"},
		},
		"recentTransactions": []gin.H{},
	}

	c.JSON(http.StatusOK, analytics)
}

func (h *DashboardHandler) GetStockAlerts(c *gin.Context) {
	// Dummy data for now
	alerts := gin.H{
		"data": gin.H{
			"alerts": []gin.H{},
		},
	}
	c.JSON(http.StatusOK, alerts)
}

