package database

import (
	"app-sistem-akuntansi/models"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ProductImageFixMigration fixes issues with product image uploads
func ProductImageFixMigration(db *gorm.DB) {
	log.Println("🔧 Starting Product Image Fix Migration...")

	migrationID := "product_image_fix_v1.0"
	
	// Check if this migration has already been run
	var existingMigration models.MigrationRecord
	if err := db.Where("migration_id = ?", migrationID).First(&existingMigration).Error; err == nil {
		log.Printf("✅ Product Image Fix Migration already applied at %v", existingMigration.AppliedAt)
		return
	}

	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("❌ Failed to start product image fix migration transaction: %v", tx.Error)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("❌ Product image fix migration rolled back due to panic: %v", r)
		}
	}()

	var fixesApplied []string

	// Fix 1: Ensure image_path column has correct type and size
	if err := ensureImagePathColumn(tx); err == nil {
		fixesApplied = append(fixesApplied, "Image path column validation")
	}

	// Fix 2: Clean up invalid image paths
	if err := cleanupInvalidImagePaths(tx); err == nil {
		fixesApplied = append(fixesApplied, "Invalid image paths cleanup")
	}

	// Fix 3: Create uploads directory structure
	if err := ensureUploadsDirectory(); err == nil {
		fixesApplied = append(fixesApplied, "Uploads directory structure")
	}

	// Record this migration as completed
	migrationRecord := models.MigrationRecord{
		MigrationID: migrationID,
		Description: fmt.Sprintf("Product image fix migration applied: %v", fixesApplied),
		Version:     "1.0",
		AppliedAt:   time.Now(),
	}

	if err := tx.Create(&migrationRecord).Error; err != nil {
		log.Printf("❌ Failed to record product image fix migration: %v", err)
		tx.Rollback()
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ Failed to commit product image fix migration: %v", err)
		return
	}

	log.Printf("✅ Product Image Fix Migration completed successfully. Applied fixes: %v", fixesApplied)
}

// ensureImagePathColumn ensures the image_path column has the correct specifications
func ensureImagePathColumn(tx *gorm.DB) error {
	log.Println("  🔧 Ensuring image_path column specifications...")
	
	// Check current column specification
	var columnInfo struct {
		ColumnType   string `json:"column_type"`
		IsNullable   string `json:"is_nullable"`
		ColumnDefault string `json:"column_default"`
	}
	
	err := tx.Raw(`
		SELECT column_type, is_nullable, column_default 
		FROM information_schema.columns 
		WHERE table_name = 'products' AND column_name = 'image_path'
	`).Scan(&columnInfo).Error

	if err != nil {
		log.Printf("    ❌ Error checking image_path column: %v", err)
		return err
	}

	log.Printf("    📊 Current image_path column: type=%s, nullable=%s, default=%s", 
		columnInfo.ColumnType, columnInfo.IsNullable, columnInfo.ColumnDefault)

	// Ensure column is adequately sized (VARCHAR(255) should be sufficient)
	if columnInfo.ColumnType != "varchar(255)" {
		log.Println("    🔧 Adjusting image_path column specifications...")
		
		// Modify column to ensure proper size
		err = tx.Exec("ALTER TABLE products MODIFY COLUMN image_path VARCHAR(255) DEFAULT ''").Error
		if err != nil {
			log.Printf("    ❌ Failed to modify image_path column: %v", err)
			return err
		}
		log.Println("    ✅ Modified image_path column to VARCHAR(255)")
	} else {
		log.Println("    ✅ image_path column specifications are correct")
	}

	return nil
}

// cleanupInvalidImagePaths cleans up any invalid or malformed image paths
func cleanupInvalidImagePaths(tx *gorm.DB) error {
	log.Println("  🔧 Cleaning up invalid image paths...")

	// Find products with potentially problematic image paths
	var problematicProducts []struct {
		ID        uint   `json:"id"`
		Code      string `json:"code"`
		Name      string `json:"name"`
		ImagePath string `json:"image_path"`
	}

	err := tx.Raw(`
		SELECT id, code, name, image_path
		FROM products 
		WHERE image_path IS NOT NULL 
		AND image_path != ''
		AND (
			LENGTH(image_path) > 255 
			OR image_path LIKE '%\\\\%' 
			OR image_path LIKE '%//%'
			OR image_path NOT LIKE '/uploads/%'
		)
		LIMIT 100
	`).Scan(&problematicProducts).Error

	if err != nil {
		log.Printf("    ❌ Error finding products with invalid image paths: %v", err)
		return err
	}

	if len(problematicProducts) == 0 {
		log.Println("    ✅ No products with invalid image paths found")
		return nil
	}

	log.Printf("    📊 Found %d products with potentially invalid image paths", len(problematicProducts))

	fixedCount := 0
	for _, product := range problematicProducts {
		var newImagePath string
		
		// Try to fix common issues
		if len(product.ImagePath) > 255 {
			// Path too long - reset to empty
			newImagePath = ""
			log.Printf("    🔧 Resetting overly long path for product %d (%s)", product.ID, product.Code)
		} else {
			// Fix path separators and format
			path := product.ImagePath
			
			// Convert backslashes to forward slashes
			path = strings.ReplaceAll(path, "\\", "/")
			
			// Remove double slashes
			for strings.Contains(path, "//") {
				path = strings.ReplaceAll(path, "//", "/")
			}
			
			// Ensure path starts with /uploads/ if it contains uploads
			if strings.Contains(path, "uploads") && !strings.HasPrefix(path, "/uploads/") {
				if strings.Contains(path, "/uploads/") {
					// Extract from /uploads/ onwards
					index := strings.Index(path, "/uploads/")
					path = path[index:]
				} else if strings.Contains(path, "uploads/") {
					// Add leading slash
					index := strings.Index(path, "uploads/")
					path = "/" + path[index:]
				}
			}
			
			newImagePath = path
		}
		
		// Update the product if the path changed
		if newImagePath != product.ImagePath {
			err := tx.Model(&models.Product{}).
				Where("id = ?", product.ID).
				Update("image_path", newImagePath).Error
			
			if err != nil {
				log.Printf("    ❌ Failed to fix image path for product %d: %v", product.ID, err)
			} else {
				log.Printf("    ✅ Fixed image path for product %d (%s): %s -> %s", 
					product.ID, product.Code, product.ImagePath, newImagePath)
				fixedCount++
			}
		}
	}

	log.Printf("    ✅ Fixed %d out of %d problematic image paths", fixedCount, len(problematicProducts))
	return nil
}

// ensureUploadsDirectory creates the uploads directory structure if it doesn't exist
func ensureUploadsDirectory() error {
	log.Println("  🔧 Ensuring uploads directory structure...")

	directories := []string{
		"./uploads",
		"./uploads/products",
		"./uploads/assets",
		"./uploads/temp",
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("    ❌ Failed to create directory %s: %v", dir, err)
			return err
		}
	}

	log.Println("    ✅ Upload directory structure ensured")
	return nil
}

