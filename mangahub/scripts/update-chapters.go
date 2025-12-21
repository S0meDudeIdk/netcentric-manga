package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Get database path
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// Default to data directory - use mangahub.db to match main application
		dbPath = filepath.Join("data", "mangahub.db")
	}

	log.Printf("Opening database at: %s", dbPath)

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Get manga with 0 total_chapters
	rows, err := db.Query(`
		SELECT id, title, total_chapters 
		FROM manga 
		WHERE total_chapters = 0 OR total_chapters IS NULL
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Fatalf("Failed to query manga: %v", err)
	}
	defer rows.Close()

	var mangaList []struct {
		ID            string
		Title         string
		TotalChapters int
	}

	for rows.Next() {
		var m struct {
			ID            string
			Title         string
			TotalChapters int
		}
		if err := rows.Scan(&m.ID, &m.Title, &m.TotalChapters); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		mangaList = append(mangaList, m)
	}

	log.Printf("Found %d manga with 0 or null total_chapters", len(mangaList))

	if len(mangaList) == 0 {
		log.Println("No manga need updating!")
		return
	}

	// For now, we'll just set a default value
	// In a real scenario, you'd want to fetch from MAL API
	// But this requires API keys and rate limiting

	log.Println("Setting default total_chapters to 1 for manga with 0...")

	stmt, err := db.Prepare("UPDATE manga SET total_chapters = ? WHERE id = ?")
	if err != nil {
		log.Fatalf("Failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	updated := 0
	for _, m := range mangaList {
		// Set to 1 as a placeholder (better than 0)
		// You can manually update specific manga later
		_, err := stmt.Exec(1, m.ID)
		if err != nil {
			log.Printf("Failed to update manga %s (%s): %v", m.ID, m.Title, err)
			continue
		}
		updated++
		log.Printf("Updated: %s - %s", m.ID, m.Title)
	}

	log.Printf("Successfully updated %d manga", updated)

	fmt.Println("\n======================================")
	fmt.Println("Note: Manga total_chapters have been set to 1 as a placeholder.")
	fmt.Println("To get accurate chapter counts:")
	fmt.Println("1. Use the sync endpoint: POST /api/v1/manga/sync")
	fmt.Println("2. Or manually update specific manga via the API")
	fmt.Println("======================================")
}
