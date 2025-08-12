package main

import (
	"context"
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
)

func main() {
	// Connect to database
	db := database.ConnectDB()

	// Create account repository
	accountRepo := repositories.NewAccountRepository(db)

	fmt.Println("=== DEBUGGING HIERARCHY BUILDING LOGIC ===")

	// Get ALL accounts first
	var allAccounts []models.Account
	if err := db.Order("code").Find(&allAccounts).Error; err != nil {
		log.Fatalf("Failed to get all accounts: %v", err)
	}

	fmt.Printf("Total accounts loaded: %d\n", len(allAccounts))
	for i, acc := range allAccounts {
		parentInfo := "ROOT"
		if acc.ParentID != nil {
			parentInfo = fmt.Sprintf("Parent ID: %d", *acc.ParentID)
		}
		fmt.Printf("[%d] ID: %d, Code: %s, Name: %s, %s\n", i, acc.ID, acc.Code, acc.Name, parentInfo)
	}

	fmt.Println("\n=== Building Account Map ===")
	accountMap := make(map[uint]*models.Account)
	for i := range allAccounts {
		allAccounts[i].Children = []models.Account{}
		accountMap[allAccounts[i].ID] = &allAccounts[i]
		fmt.Printf("Added to map: ID %d -> %s (%s)\n", allAccounts[i].ID, allAccounts[i].Code, allAccounts[i].Name)
	}

	fmt.Println("\n=== Building Hierarchy Structure ===")
	var rootAccounts []*models.Account
	for i := range allAccounts {
		account := &allAccounts[i]
		if account.ParentID != nil {
			if parent, exists := accountMap[*account.ParentID]; exists {
				fmt.Printf("Adding %s (%d) as child of %s (%d)\n", account.Code, account.ID, parent.Code, parent.ID)
				parent.Children = append(parent.Children, *account)
			} else {
				fmt.Printf("ORPHAN FOUND: %s (%d) has parent ID %d but parent not found in map\n", account.Code, account.ID, *account.ParentID)
				rootAccounts = append(rootAccounts, account)
			}
		} else {
			fmt.Printf("ROOT FOUND: %s (%d)\n", account.Code, account.ID)
			rootAccounts = append(rootAccounts, account)
		}
	}

	fmt.Printf("\nTotal root accounts: %d\n", len(rootAccounts))
	for i, root := range rootAccounts {
		fmt.Printf("Root[%d]: %s (%d) - Children: %d\n", i, root.Code, root.ID, len(root.Children))
		for j, child := range root.Children {
			fmt.Printf("  Child[%d]: %s (%d) - Children: %d\n", j, child.Code, child.ID, len(child.Children))
		}
	}

	fmt.Println("\n=== Using Repository GetHierarchy ===")
	ctx := context.Background()
	accounts, err := accountRepo.GetHierarchy(ctx)
	if err != nil {
		log.Fatalf("Failed to get hierarchy: %v", err)
	}

	fmt.Printf("Repository returned %d root accounts:\n", len(accounts))
	for i, acc := range accounts {
		fmt.Printf("Root[%d]: %s (%d) - Children: %d\n", i, acc.Code, acc.ID, len(acc.Children))
		for j, child := range acc.Children {
			fmt.Printf("  Child[%d]: %s (%d) - Children: %d\n", j, child.Code, child.ID, len(child.Children))
			for k, grandchild := range child.Children {
				fmt.Printf("    Grandchild[%d]: %s (%d)\n", k, grandchild.Code, grandchild.ID)
			}
		}
	}
}
