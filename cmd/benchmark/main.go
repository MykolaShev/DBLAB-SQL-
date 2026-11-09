package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"freelance-platform/internal/nosql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	TotalRecords = 10000 // Number of records to generate
	TotalReads   = 5000  // Number of random readings during the test
)

func main() {
	godotenv.Load()

	// 1. Connect to PostgreSQL
	pgDB, err := sql.Open("postgres", os.Getenv("DB_CONN_STR"))
	if err != nil {
		log.Fatalf("Postgres connection failed: %v", err)
	}
	defer pgDB.Close()

	// 2. Connect to MongoDB
	ctx := context.Background()
	mongoStore, err := nosql.NewMongoStore(ctx, os.Getenv("MONGO_URI"), "freelance", "freelancer_profiles_view")
	if err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	defer mongoStore.Close(ctx)

	fmt.Println("--- Starting Data Seeding ---")
	seedData(ctx, pgDB, mongoStore)

	fmt.Println("\n--- Starting Benchmark ---")
	benchmarkSQL(pgDB)
	benchmarkMongo(mongoStore)

	fmt.Println("\n--- Testing Redis (Key-Value NoSQL) ---")
	redisStore, err := nosql.NewRedisStore(ctx, os.Getenv("REDIS_ADDR"))
	if err != nil {
		log.Printf("Redis connection failed: %v", err)
	} else {
		defer redisStore.Close()
		
		// Simulating "Top-3 projects"
		topProjects := []string{"E-commerce App", "Mobile Game", "CRM System"}
		
		fmt.Println("[Redis] Caching top projects...")
		err = redisStore.CacheTopProjects(ctx, "top_10_projects", topProjects, 5*time.Minute)
		if err != nil {
			log.Printf("Failed to cache: %v", err)
		}

		var cachedProjects []string
		err = redisStore.GetCachedProjects(ctx, "top_10_projects", &cachedProjects)
		if err != nil {
			log.Printf("Failed to get cache: %v", err)
		} else {
			fmt.Printf("[Redis] Successfully retrieved from cache: %v\n", cachedProjects)
		}
	}
}

// seedData generates mock data and inserts it into both databases
func seedData(ctx context.Context, pgDB *sql.DB, mongoStore *nosql.MongoStore) {
	// Check if already seeded to save time on multiple runs
	var count int
	pgDB.QueryRow("SELECT COUNT(*) FROM users WHERE username LIKE 'bench_user_%'").Scan(&count)
	if count >= TotalRecords {
		fmt.Printf("Data already seeded (%d records found).\n", count)
		return
	}

	fmt.Printf("Seeding %d records... This might take a minute.\n", TotalRecords)
	
	// Transaction for PostgreSQL to speed up inserts
	tx, err := pgDB.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	
	// Protection from manual inserts
	_, err = tx.Exec(`SELECT setval('users_id_seq', COALESCE((SELECT MAX(id) FROM users), 1))`)
	if err != nil {
		log.Printf("Warning: Failed to sync sequence: %v", err)
	}
	
	for i := 1; i <= TotalRecords; i++ {
		username := fmt.Sprintf("bench_user_%d", i)
		email := fmt.Sprintf("bench%d@email.com", i)
		bio := fmt.Sprintf("I am a software engineer number %d with great skills.", i)
		
		// PostgreSQL Insert
		var userID int
		err := tx.QueryRow(
			"INSERT INTO users (role_id, username, email, password_hash) VALUES (2, $1, $2, 'hash') RETURNING id",
			username, email,
		).Scan(&userID)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed PG insert at index %d: %v", i, err)
		}

		_, err = tx.Exec(
			"INSERT INTO freelancer_profiles (user_id, first_name, last_name, bio, hourly_rate) VALUES ($1, 'Name', 'Lastname', $2, 35.00)",
			userID, bio,
		)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed PG profile insert at index %d: %v", i, err)
		}

		// MongoDB Insert
		mongoProfile := nosql.FreelancerProfileDoc{
			UserID:     userID,
			Username:   username,
			Bio:        bio,
			HourlyRate: 35.00,
			Skills:     []string{"Go", "PostgreSQL", "Docker"},
		}
		mongoStore.InsertProfile(ctx, mongoProfile)
	}
	tx.Commit()
	fmt.Println("Seeding completed successfully!")
}

// benchmarkSQL reads profiles from PostgreSQL using JOINs
func benchmarkSQL(db *sql.DB) {
	start := time.Now()
	var wg sync.WaitGroup

	// Fetch max ID to generate random IDs
	var maxID int
	db.QueryRow("SELECT MAX(id) FROM users").Scan(&maxID)

	for i := 0; i < TotalReads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			randomID := rand.Intn(maxID) + 1
			
			// Complex JOIN simulation
			query := `
				SELECT u.username, fp.bio, fp.hourly_rate 
				FROM users u
				JOIN freelancer_profiles fp ON u.id = fp.user_id
				WHERE u.id = $1`
				
			var username, bio string
			var rate float64
			db.QueryRow(query, randomID).Scan(&username, &bio, &rate)
		}()
	}
	wg.Wait()
	duration := time.Since(start)
	
	fmt.Printf("[PostgreSQL] %d concurrent reads completed in %v\n", TotalReads, duration)
	fmt.Printf("[PostgreSQL] Average latency: %v per request\n", duration/time.Duration(TotalReads))
}

// benchmarkMongo reads profiles from MongoDB directly by indexed user_id
func benchmarkMongo(store *nosql.MongoStore) {
	start := time.Now()
	var wg sync.WaitGroup
	ctx := context.Background()

	// Hardcoding maxID logic for simplicity based on generated rows
	maxID := TotalRecords

	for i := 0; i < TotalReads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			randomID := rand.Intn(maxID) + 1
			store.GetProfileByID(ctx, randomID)
		}()
	}
	wg.Wait()
	duration := time.Since(start)
	
	fmt.Printf("[MongoDB]    %d concurrent reads completed in %v\n", TotalReads, duration)
	fmt.Printf("[MongoDB]    Average latency: %v per request\n", duration/time.Duration(TotalReads))
}