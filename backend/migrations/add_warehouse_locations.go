package main

import (
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/models"

	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db, err := config.InitDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Starting warehouse locations migration...")

	// Create warehouse locations table
	err = db.AutoMigrate(&models.WarehouseLocation{})
	if err != nil {
		log.Fatal("Failed to create warehouse_locations table:", err)
	}
	fmt.Println("✓ Created warehouse_locations table")

	// Add warehouse_location_id column to products table if it doesn't exist
	if !db.Migrator().HasColumn(&models.Product{}, "warehouse_location_id") {
		err = db.Migrator().AddColumn(&models.Product{}, "warehouse_location_id")
		if err != nil {
			log.Fatal("Failed to add warehouse_location_id column to products table:", err)
		}
		fmt.Println("✓ Added warehouse_location_id column to products table")
	} else {
		fmt.Println("✓ warehouse_location_id column already exists in products table")
	}

	// Create default warehouse locations
	createDefaultWarehouseLocations(db)

	fmt.Println("Migration completed successfully!")
}

func createDefaultWarehouseLocations(db *gorm.DB) {
	defaultLocations := []models.WarehouseLocation{
		{
			Code:        "WH-001",
			Name:        "Main Warehouse",
			Description: "Primary storage facility",
			Address:     "Jl. Gudang Utama No. 1",
			IsActive:    true,
		},
		{
			Code:        "WH-002",
			Name:        "Storage Room A",
			Description: "Small items storage",
			Address:     "Jl. Gudang Utama No. 2",
			IsActive:    true,
		},
		{
			Code:        "WH-003",
			Name:        "Cold Storage",
			Description: "Temperature controlled storage",
			Address:     "Jl. Gudang Utama No. 3",
			IsActive:    true,
		},
	}

	for _, location := range defaultLocations {
		// Check if location already exists
		var existing models.WarehouseLocation
		if err := db.Where("code = ?", location.Code).First(&existing).Error; err == gorm.ErrRecordNotFound {
			// Create new location
			if err := db.Create(&location).Error; err != nil {
				log.Printf("Failed to create warehouse location %s: %v", location.Code, err)
			} else {
				fmt.Printf("✓ Created warehouse location: %s - %s\n", location.Code, location.Name)
			}
		} else {
			fmt.Printf("✓ Warehouse location already exists: %s - %s\n", location.Code, location.Name)
		}
	}
}
