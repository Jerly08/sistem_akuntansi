package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
)

func main() {
	// Connect to database using existing function
	db := database.ConnectDB()

	// Create account repository
	accountRepo := repositories.NewAccountRepository(db)

	// Get hierarchy data
	ctx := context.Background()
	accounts, err := accountRepo.GetHierarchy(ctx)
	if err != nil {
		log.Fatalf("Failed to get hierarchy: %v", err)
	}

	fmt.Printf("Total accounts in hierarchy: %d\n", len(accounts))
	fmt.Println("\nHierarchy structure:")
	
	// Convert to JSON for pretty printing
	jsonData, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Save to file for detailed inspection
	err = os.WriteFile("hierarchy_debug.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	fmt.Println("Hierarchy data saved to hierarchy_debug.json")
	
	// Print summary
	fmt.Println("\nSummary:")
	for i, acc := range accounts {
		fmt.Printf("%d. %s - %s (Level %d, Children: %d)\n", 
			i+1, acc.Code, acc.Name, acc.Level, len(acc.Children))
		printChildren(acc.Children, 1)
	}
}

func printChildren(children []models.Account, level int) {
	for _, child := range children {
		indent := ""
		for i := 0; i < level; i++ {
			indent += "  "
		}
		
		fmt.Printf("%s- %s - %s (Level %d)\n", indent, child.Code, child.Name, child.Level)
		
		if len(child.Children) > 0 {
			printChildren(child.Children, level+1)
		}
	}
}
