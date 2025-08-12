package main

import (
	"fmt"
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"app-sistem-akuntansi/models"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Connected to database successfully")
	
	// Check existing users
	var users []models.User
	result := db.Find(&users)
	if result.Error != nil {
		log.Fatalf("Error querying users: %v", result.Error)
	}

	fmt.Printf("Found %d users:\n", len(users))
	for _, user := range users {
		fmt.Printf("ID: %d, Email: %s, Role: %s, Active: %v\n", 
			user.ID, user.Email, user.Role, user.IsActive)
	}

	// If no admin user exists, create one
	var adminUser models.User
	err = db.Where("email = ? AND role = ?", "admin@example.com", "admin").First(&adminUser).Error
	if err == gorm.ErrRecordNotFound {
		fmt.Println("Creating admin user...")
		
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}
		
		newAdmin := models.User{
			Username:  "admin",
			Email:     "admin@example.com",
			Password:  string(hashedPassword),
			FirstName: "Admin",
			LastName:  "User",
			Role:      "admin",
			IsActive:  true,
		}
		
		err = db.Create(&newAdmin).Error
		if err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}
		
		fmt.Println("Admin user created successfully!")
	} else if err != nil {
		log.Fatalf("Error checking admin user: %v", err)
	} else {
		fmt.Println("Admin user already exists")
	}

	fmt.Println("User check completed")
}
