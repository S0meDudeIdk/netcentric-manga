package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mangahub/pkg/database"
	"mangahub/pkg/models"
	"os"
	"time"
)

// LoadMangaData loads manga data from JSON file into the database
func LoadMangaData(filepath string) error {
	// Open file
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file content
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	var mangaList []models.Manga
	if err := json.Unmarshal(data, &mangaList); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Get database connection
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Insert manga data
	for _, manga := range mangaList {
		// Set genres JSON
		if err := manga.SetGenres(manga.Genres); err != nil {
			log.Printf("Error setting genres for manga %s: %v", manga.ID, err)
			continue
		}

		// Set created time
		manga.CreatedAt = time.Now()

		// Insert into database
		_, err := db.Exec(`
			INSERT OR REPLACE INTO manga 
			(id, title, author, genres, status, total_chapters, description, cover_url, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			manga.ID, manga.Title, manga.Author, manga.GenresJSON,
			manga.Status, manga.TotalChapters, manga.Description,
			manga.CoverURL, manga.CreatedAt)

		if err != nil {
			log.Printf("Error inserting manga %s: %v", manga.ID, err)
			continue
		}

		log.Printf("Loaded manga: %s", manga.Title)
	}

	log.Printf("Successfully loaded %d manga entries", len(mangaList))
	return nil
}

// ValidateJSON checks if a string is valid JSON
func ValidateJSON(data string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(data), &js) == nil
}

// GenerateID creates a simple ID from a title
func GenerateID(title string) string {
	// Simple implementation - in production, use a more robust method
	id := ""
	for _, char := range title {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			id += string(char)
		} else if char == ' ' || char == '-' {
			id += "-"
		}
	}
	return id
}

// Contains checks if a slice contains a string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from a slice
func RemoveDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
