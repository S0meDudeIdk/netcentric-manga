package main

import (
	"log"
	api "mangahub/internal/api"
	"mangahub/pkg/database"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	// Try current directory first, then parent directory
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../../.env"); err != nil {
			log.Println("Warning: .env file not found, using environment variables or defaults")
		} else {
			log.Println("Loaded environment variables from ../../.env file")
		}
	} else {
		log.Println("Loaded environment variables from .env file")
	}

	// Initialize database
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Print current working directory for debugging
	if cwd, err := os.Getwd(); err == nil {
		log.Printf("Current working directory: %s", cwd)
	}

	// Create and start server
	// Note: Manga data is now fetched from external APIs (MAL/Jikan, MangaDex)
	// Local manga.json is no longer used
	server := api.NewAPIServer()

	log.Println("MangaHub API Server starting...")
	log.Printf("Server configuration:")
	log.Printf("  - Port: %s", server.Port)
	log.Printf("  - Gin Mode: %s", os.Getenv("GIN_MODE"))
	log.Printf("  - CORS Origins: %s", os.Getenv("CORS_ALLOW_ORIGINS"))
	log.Printf("  - Rate Limit: %s requests/min", os.Getenv("RATE_LIMIT_REQUESTS_PER_MINUTE"))

	// Log MAL API configuration
	if server.MALClient.IsConfigured() {
		log.Printf("  - MyAnimeList: Official API (Client ID configured) ✓")
	} else {
		log.Printf("  - MyAnimeList: Jikan API (unofficial) - Configure MAL_CLIENT_ID for official API")
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
