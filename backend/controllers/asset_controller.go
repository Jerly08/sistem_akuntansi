package controllers

import (
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AssetController struct {
	assetService services.AssetServiceInterface
}

type AssetCreateRequest struct {
	Code               string    `json:"code"`
	Name               string    `json:"name" binding:"required"`
	Category           string    `json:"category" binding:"required"`
	Status             string    `json:"status"`
	PurchaseDate       time.Time `json:"purchase_date" binding:"required"`
	PurchasePrice      float64   `json:"purchase_price" binding:"required,gt=0"`
	SalvageValue       float64   `json:"salvage_value"`
	UsefulLife         int       `json:"useful_life" binding:"gt=0"`
	DepreciationMethod string    `json:"depreciation_method"`
	IsActive           bool      `json:"is_active"`
	Notes              string    `json:"notes"`
	Location           string    `json:"location"`
	SerialNumber       string    `json:"serial_number"`
	Condition          string    `json:"condition"`
	AssetAccountID     *uint     `json:"asset_account_id"`
	DepreciationAccountID *uint  `json:"depreciation_account_id"`
}

type AssetUpdateRequest struct {
	Name               string    `json:"name" binding:"required"`
	Category           string    `json:"category" binding:"required"`
	Status             string    `json:"status"`
	PurchaseDate       time.Time `json:"purchase_date" binding:"required"`
	PurchasePrice      float64   `json:"purchase_price" binding:"required,gt=0"`
	SalvageValue       float64   `json:"salvage_value"`
	UsefulLife         int       `json:"useful_life" binding:"gt=0"`
	DepreciationMethod string    `json:"depreciation_method"`
	IsActive           bool      `json:"is_active"`
	Notes              string    `json:"notes"`
	Location           string    `json:"location"`
	SerialNumber       string    `json:"serial_number"`
	Condition          string    `json:"condition"`
	AssetAccountID     *uint     `json:"asset_account_id"`
	DepreciationAccountID *uint  `json:"depreciation_account_id"`
}

func NewAssetController(db *gorm.DB) *AssetController {
	assetRepo := repositories.NewAssetRepository(db)
	assetService := services.NewAssetService(assetRepo)
	
	return &AssetController{
		assetService: assetService,
	}
}

// GetAssets retrieves all assets with optional filtering
func (ac *AssetController) GetAssets(c *gin.Context) {
	assets, err := ac.assetService.GetAllAssets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve assets",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Assets retrieved successfully",
		"data":    assets,
		"count":   len(assets),
	})
}

// GetAsset retrieves a specific asset by ID
func (ac *AssetController) GetAsset(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	asset, err := ac.assetService.GetAssetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Asset not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Asset retrieved successfully",
		"data":    asset,
	})
}

// CreateAsset creates a new asset
func (ac *AssetController) CreateAsset(c *gin.Context) {
	var req AssetCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Convert request to asset model
	asset := &models.Asset{
		Code:                  req.Code,
		Name:                  req.Name,
		Category:              req.Category,
		Status:                req.Status,
		PurchaseDate:          req.PurchaseDate,
		PurchasePrice:         req.PurchasePrice,
		SalvageValue:          req.SalvageValue,
		UsefulLife:            req.UsefulLife,
		DepreciationMethod:    req.DepreciationMethod,
		IsActive:              req.IsActive,
		Notes:                 req.Notes,
		Location:              req.Location,
		SerialNumber:          req.SerialNumber,
		Condition:             req.Condition,
		AssetAccountID:        req.AssetAccountID,
		DepreciationAccountID: req.DepreciationAccountID,
	}

	// Set defaults
	if asset.Status == "" {
		asset.Status = models.AssetStatusActive
	}
	if !asset.IsActive {
		asset.IsActive = true
	}

	err := ac.assetService.CreateAsset(asset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to create asset",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Asset created successfully",
		"data":    asset,
	})
}

// UpdateAsset updates an existing asset
func (ac *AssetController) UpdateAsset(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	var req AssetUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Get existing asset
	existingAsset, err := ac.assetService.GetAssetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Asset not found",
			"details": err.Error(),
		})
		return
	}

	// Update fields
	existingAsset.Name = req.Name
	existingAsset.Category = req.Category
	existingAsset.Status = req.Status
	existingAsset.PurchaseDate = req.PurchaseDate
	existingAsset.PurchasePrice = req.PurchasePrice
	existingAsset.SalvageValue = req.SalvageValue
	existingAsset.UsefulLife = req.UsefulLife
	existingAsset.DepreciationMethod = req.DepreciationMethod
	existingAsset.IsActive = req.IsActive
	existingAsset.Notes = req.Notes
	existingAsset.Location = req.Location
	existingAsset.SerialNumber = req.SerialNumber
	existingAsset.Condition = req.Condition
	existingAsset.AssetAccountID = req.AssetAccountID
	existingAsset.DepreciationAccountID = req.DepreciationAccountID

	err = ac.assetService.UpdateAsset(existingAsset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to update asset",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Asset updated successfully",
		"data":    existingAsset,
	})
}

// DeleteAsset deletes an asset
func (ac *AssetController) DeleteAsset(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	err = ac.assetService.DeleteAsset(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Failed to delete asset",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Asset deleted successfully",
	})
}

// GetAssetsSummary returns summary statistics for assets
func (ac *AssetController) GetAssetsSummary(c *gin.Context) {
	summary, err := ac.assetService.GetAssetsSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get assets summary",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Assets summary retrieved successfully",
		"data":    summary,
	})
}

// GetDepreciationReport returns depreciation report for all assets
func (ac *AssetController) GetDepreciationReport(c *gin.Context) {
	report, err := ac.assetService.GetAssetsForDepreciationReport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate depreciation report",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Depreciation report generated successfully",
		"data":    report,
		"count":   len(report),
	})
}

// GetDepreciationSchedule returns depreciation schedule for a specific asset
func (ac *AssetController) GetDepreciationSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	asset, err := ac.assetService.GetAssetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Asset not found",
			"details": err.Error(),
		})
		return
	}

	schedule, err := ac.assetService.GetDepreciationSchedule(asset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate depreciation schedule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Depreciation schedule generated successfully",
		"data": gin.H{
			"asset":    asset,
			"schedule": schedule,
		},
	})
}

// CalculateCurrentDepreciation calculates current depreciation for an asset
func (ac *AssetController) CalculateCurrentDepreciation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid asset ID"})
		return
	}

	asset, err := ac.assetService.GetAssetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Asset not found",
			"details": err.Error(),
		})
		return
	}

	// Get date parameter (optional, defaults to now)
	dateStr := c.DefaultQuery("as_of_date", "")
	var asOfDate time.Time
	if dateStr != "" {
		asOfDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		asOfDate = time.Now()
	}

	depreciation, err := ac.assetService.CalculateDepreciation(asset, asOfDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to calculate depreciation",
			"details": err.Error(),
		})
		return
	}

	currentBookValue := asset.PurchasePrice - depreciation

	c.JSON(http.StatusOK, gin.H{
		"message": "Depreciation calculated successfully",
		"data": gin.H{
			"asset_id":                    asset.ID,
			"asset_name":                  asset.Name,
			"as_of_date":                  asOfDate.Format("2006-01-02"),
			"purchase_price":              asset.PurchasePrice,
			"salvage_value":               asset.SalvageValue,
			"accumulated_depreciation":    depreciation,
			"current_book_value":          currentBookValue,
			"depreciation_method":         asset.DepreciationMethod,
			"useful_life_years":           asset.UsefulLife,
		},
	})
}
