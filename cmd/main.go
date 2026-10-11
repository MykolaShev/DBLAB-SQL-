package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"freelance-platform/internal/store"
)

func main() {
	connStr := "user=postgres password=secret dbname=freelance sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	uow, err := store.NewUnitOfWork(ctx, db)
	if err != nil {
		log.Fatalf("Failed to start transaction: %v", err)
	}

	// Defer rollback ensures atomic consistency.
	// If uow.Commit() succeeds, calling Rollback() afterward does nothing.
	defer uow.Rollback()

	// Business Logic Scenario: Accept proposal and update project
	
	proposalID := 1
	projectID := 1
	clientID := 2 // ID of the user performing the action

	fmt.Println("Step 1: Accepting proposal and generating contract...")
	err = uow.Proposals().AcceptProposal(ctx, proposalID, clientID, "2026-10-15", "2026-11-15")
	if err != nil {
		log.Fatalf("Error accepting proposal: %v", err)
	}

	fmt.Println("Step 2: Updating project status to 'In Progress'...")
	err = uow.Projects().UpdateStatus(ctx, projectID, "In Progress", clientID)
	if err != nil {
		log.Fatalf("Error updating project status: %v", err)
	}

	// Commit transaction if all operations succeed
	if err := uow.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Println("Success! Transaction committed.")

	// Fetching Read-Only Data (via Views)
	
	// Create a new short-lived transaction for reading
	readUow, _ := store.NewUnitOfWork(ctx, db)
	defer readUow.Rollback()

	projects, err := readUow.Projects().GetActiveProjects(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch projects: %v", err)
	}

	fmt.Printf("Found %d active projects:\n", len(projects))
	for _, p := range projects {
		fmt.Printf("- [%d] %s (Budget: $%.2f, Client: %s)\n", p.ProjectID, p.Title, p.Budget, p.Client)
	}
}