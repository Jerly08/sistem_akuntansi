package controllers

import (
	"net/http"
	"strconv"
	"time"
	"app-sistem-akuntansi/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


type ProductController struct {
	DB *gorm.DB
}

func NewProductController(db *gorm.DB) *ProductController {
	return &ProductController{DB: db}
}

func (pc *ProductController) GetProducts(c *gin.Context) {
	var products []models.Product
	
	query := pc.DB.Where("is_active = ?", true)
	
	// Add search functionality
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	// Add category filter
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	
	if err := query.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Products retrieved successfully",
		"data":    products,
	})
}

func (pc *ProductController) GetProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := pc.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product retrieved successfully",
		"data":    product,
	})
}

// AdjustStock handles stock adjustments for products
func (pc *ProductController) AdjustStock(c *gin.Context) {
	var input struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity" binding:"required"`
		Type      string `json:"type" binding:"required,oneof=IN OUT"`
		Notes     string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product := models.Product{}
	if err := pc.DB.First(&product, input.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	inventory := models.Inventory{
		ProductID:     input.ProductID,
		Type:          input.Type,
		Quantity:      input.Quantity,
		TransactionDate: time.Now(),
		Notes:         input.Notes,
	}

	if err := pc.DB.Save(&inventory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save inventory adjustment"})
		return
	}

	if input.Type == models.InventoryTypeIn {
		product.Stock += input.Quantity
	} else if input.Type == models.InventoryTypeOut {
		product.Stock -= input.Quantity
	}

	if err := pc.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock adjusted successfully", "product": product})
}

// Opname processes stock opname
func (pc *ProductController) Opname(c *gin.Context) {
	var input struct {
		ProductID uint `json:"product_id" binding:"required"`
		NewStock  int  `json:"new_stock" binding:"required"`
		Notes     string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product := models.Product{}
	if err := pc.DB.First(&product, input.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	inventory := models.Inventory{
		ProductID:     input.ProductID,
		Type:          models.InventoryTypeIn,
		Quantity:      input.NewStock - product.Stock,
		TransactionDate: time.Now(),
		Notes:         input.Notes,
	}

	if err := pc.DB.Save(&inventory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save inventory opname"})
		return
	}

	product.Stock = input.NewStock

	if err := pc.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock opname processed successfully", "product": product})
}

func (pc *ProductController) CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if product code already exists
	var existingProduct models.Product
	if err := pc.DB.Where("code = ?", product.Code).First(&existingProduct).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Product code already exists"})
		return
	}

	if err := pc.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
		"data":    product,
	})
}

func (pc *ProductController) UpdateProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := pc.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var updateData models.Product
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if new code conflicts with existing products
	if updateData.Code != product.Code {
		var existingProduct models.Product
		if err := pc.DB.Where("code = ? AND id != ?", updateData.Code, id).First(&existingProduct).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Product code already exists"})
			return
		}
	}

	if err := pc.DB.Model(&product).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
		"data":    product,
	})
}

func (pc *ProductController) DeleteProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := pc.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Soft delete
	if err := pc.DB.Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product deleted successfully",
	})
}
