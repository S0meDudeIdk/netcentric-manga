package main

import (
	"context"
	"log"
	"mangahub/internal/grpc"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../../.env"); err != nil {
			log.Println("Warning: .env file not found, using defaults")
		}
	}

	// Get gRPC server address
	grpcPort := os.Getenv("GRPC_SERVER_PORT")
	if grpcPort == "" {
		grpcPort = "9001"
	}
	address := "localhost:" + grpcPort

	// Create gRPC client
	log.Printf("Connecting to gRPC server at %s...", address)
	client, err := grpc.NewClient(address)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer client.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test 1: Ping the server
	log.Println("\n=== Test 1: Ping Server ===")
	if err := client.Ping(ctx); err != nil {
		log.Printf("Ping failed: %v", err)
	} else {
		log.Println("Ping successful!")
	}

	// Test 2: Get a manga by ID
	log.Println("\n=== Test 2: Get Manga by ID ===")
	manga, err := client.GetManga(ctx, "1")
	if err != nil {
		log.Printf("GetManga failed: %v", err)
	} else {
		log.Printf("GetManga successful: %+v", manga)
	}

	// Test 3: Search manga
	log.Println("\n=== Test 3: Search Manga ===")
	searchResult, err := client.SearchManga(ctx, "naruto", 5)
	if err != nil {
		log.Printf("SearchManga failed: %v", err)
	} else {
		log.Printf("SearchManga successful: %+v", searchResult)
	}

	// Test 4: Update progress
	log.Println("\n=== Test 4: Update Progress ===")
	progressResult, err := client.UpdateProgress(ctx, "user123", "1", 10, "reading")
	if err != nil {
		log.Printf("UpdateProgress failed: %v", err)
	} else {
		log.Printf("UpdateProgress successful: %+v", progressResult)
	}

	log.Println("\n=== All gRPC tests completed ===")
}
