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

	// Test 1: Get a manga by ID
	log.Println("\n=== Test 1: Get Manga by ID ===")
	manga, err := client.GetManga(ctx, "md-da45f161-d568-4b32-bac1-5eda0cdecaea")
	if err != nil {
		log.Printf("GetManga failed: %v", err)
	} else {
		log.Printf("GetManga successful: Title=%s, Author=%s", manga.Manga.Title, manga.Manga.Author)
	}

	// Test 2: Search manga with relevant sorting
	log.Println("\n=== Test 2: Search Manga (Relevant) ===")
	searchResult, err := client.SearchManga(ctx, "one piece", 5, 0, "relevant")
	if err != nil {
		log.Printf("SearchManga failed: %v", err)
	} else {
		log.Printf("SearchManga successful: Found %d manga", len(searchResult.Manga))
		for i, m := range searchResult.Manga {
			log.Printf("  %d. %s (by %s)", i+1, m.Title, m.Author)
		}
	}

	// Test 3: Search manga with newest sorting
	log.Println("\n=== Test 3: Search Manga (Newest) ===")
	newestResult, err := client.SearchManga(ctx, "", 5, 0, "newest")
	if err != nil {
		log.Printf("SearchManga (newest) failed: %v", err)
	} else {
		log.Printf("SearchManga (newest) successful: Found %d manga", len(newestResult.Manga))
		for i, m := range newestResult.Manga {
			log.Printf("  %d. %s", i+1, m.Title)
		}
	}

	// Test 4: Get library
	log.Println("\n=== Test 4: Get Library ===")
	testUserID := "test_user_123"
	libraryResult, err := client.GetLibrary(ctx, testUserID)
	if err != nil {
		log.Printf("GetLibrary failed: %v", err)
	} else {
		log.Printf("GetLibrary successful: Reading=%d, Completed=%d, PlanToRead=%d",
			len(libraryResult.Reading), len(libraryResult.Completed), len(libraryResult.PlanToRead))
	}

	// Test 5: Add manga to library
	log.Println("\n=== Test 5: Add to Library ===")
	addResult, err := client.AddToLibrary(ctx, testUserID, "md-da45f161-d568-4b32-bac1-5eda0cdecaea", "reading")
	if err != nil {
		log.Printf("AddToLibrary failed: %v", err)
	} else {
		log.Printf("AddToLibrary successful: %s", addResult.Message)
	}

	// Test 6: Update progress
	log.Println("\n=== Test 6: Update Progress ===")
	progressResult, err := client.UpdateProgress(ctx, testUserID, "md-da45f161-d568-4b32-bac1-5eda0cdecaea", 10, "reading")
	if err != nil {
		log.Printf("UpdateProgress failed: %v", err)
	} else {
		log.Printf("UpdateProgress successful: %s", progressResult.Message)
	}

	// Test 7: Rate manga
	log.Println("\n=== Test 7: Rate Manga ===")
	ratingResult, err := client.RateManga(ctx, testUserID, "md-da45f161-d568-4b32-bac1-5eda0cdecaea", 5)
	if err != nil {
		log.Printf("RateManga failed: %v", err)
	} else {
		log.Printf("RateManga successful: Average=%.2f, Total=%d",
			ratingResult.AverageRating, ratingResult.TotalRatings)
	}

	// Test 8: Get manga ratings
	log.Println("\n=== Test 8: Get Manga Ratings ===")
	ratingsResult, err := client.GetMangaRatings(ctx, "md-da45f161-d568-4b32-bac1-5eda0cdecaea", testUserID)
	if err != nil {
		log.Printf("GetMangaRatings failed: %v", err)
	} else {
		log.Printf("GetMangaRatings successful: Average=%.2f, Total=%d, UserRating=%d",
			ratingsResult.AverageRating, ratingsResult.TotalRatings, ratingsResult.UserRating)
	}

	// Test 9: Get library stats
	log.Println("\n=== Test 9: Get Library Stats ===")
	statsResult, err := client.GetLibraryStats(ctx, testUserID)
	if err != nil {
		log.Printf("GetLibraryStats failed: %v", err)
	} else {
		log.Printf("GetLibraryStats successful: Total=%d, Reading=%d, Completed=%d, Chapters=%d",
			statsResult.TotalManga, statsResult.Reading, statsResult.Completed, statsResult.TotalChaptersRead)
	}

	// Test 10: Remove from library
	log.Println("\n=== Test 10: Remove from Library ===")
	removeResult, err := client.RemoveFromLibrary(ctx, testUserID, "md-da45f161-d568-4b32-bac1-5eda0cdecaea")
	if err != nil {
		log.Printf("RemoveFromLibrary failed: %v", err)
	} else {
		log.Printf("RemoveFromLibrary successful: %s", removeResult.Message)
	}

	// Test 11: Get user profile
	log.Println("\n=== Test 11: Get User Profile ===")
	profileResult, err := client.GetUserProfile(ctx, testUserID)
	if err != nil {
		log.Printf("GetUserProfile failed: %v", err)
	} else if profileResult.Error != "" {
		log.Printf("GetUserProfile error: %s", profileResult.Error)
	} else {
		log.Printf("GetUserProfile successful: Username=%s, Email=%s",
			profileResult.Profile.Username, profileResult.Profile.Email)
	}

	// Test 12: Update user profile
	log.Println("\n=== Test 12: Update User Profile ===")
	updateProfileResult, err := client.UpdateUserProfile(ctx, testUserID, "test_user_updated", "updated@test.com")
	if err != nil {
		log.Printf("UpdateUserProfile failed: %v", err)
	} else if updateProfileResult.Error != "" {
		log.Printf("UpdateUserProfile error: %s", updateProfileResult.Error)
	} else {
		log.Printf("UpdateUserProfile successful: %s", updateProfileResult.Message)
	}

	// Test 13: Change password
	log.Println("\n=== Test 13: Change Password ===")
	changePasswordResult, err := client.ChangePassword(ctx, testUserID, "oldPassword123", "newPassword456")
	if err != nil {
		log.Printf("ChangePassword failed: %v", err)
	} else if changePasswordResult.Error != "" {
		log.Printf("ChangePassword error: %s", changePasswordResult.Error)
	} else {
		log.Printf("ChangePassword successful: %s", changePasswordResult.Message)
	}

	log.Println("\n=== All gRPC tests completed ===")
}
