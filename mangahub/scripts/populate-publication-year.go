// UNUSED FILE - COMMENTED OUT
// This script was never executed or imported anywhere in the codebase
// The functions fetchMangaDexYear and fetchMALYear are placeholders that always return 0

/*
package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open database
	db, err := sql.Open("sqlite3", "../../data/mangahub.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Get all manga without publication_year
	rows, err := db.Query(`
		SELECT id, title FROM manga
		WHERE publication_year IS NULL OR publication_year = 0
		ORDER BY title
	`)
	if err != nil {
		log.Fatalf("Failed to query manga: %v", err)
	}
	defer rows.Close()

	updated := 0
	skipped := 0

	for rows.Next() {
		var id, title string
		if err := rows.Scan(&id, &title); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Try to extract year from MangaDex or MAL API
		year := 0

		if strings.HasPrefix(id, "md-") {
			// MangaDex ID
			mangadexID := strings.TrimPrefix(id, "md-")
			year = fetchMangaDexYear(mangadexID)
		} else if strings.HasPrefix(id, "mal-") {
			// MAL ID
			malID := strings.TrimPrefix(id, "mal-")
			year = fetchMALYear(malID)
		}

		if year > 0 {
			_, err := db.Exec("UPDATE manga SET publication_year = ? WHERE id = ?", year, id)
			if err != nil {
				log.Printf("Failed to update %s: %v", title, err)
				skipped++
			} else {
				log.Printf("✓ Updated %s: %d", title, year)
				updated++
			}
			// Rate limiting
			time.Sleep(200 * time.Millisecond)
		} else {
			log.Printf("⚠ Skipped %s: No year found", title)
			skipped++
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Updated: %d\n", updated)
	fmt.Printf("Skipped: %d\n", skipped)
}

func fetchMangaDexYear(mangadexID string) int {
	// This is a placeholder - you would implement actual API call here
	// For now, just return 0
	return 0
}

func fetchMALYear(malID string) int {
	// This is a placeholder - you would implement actual API call here
	// For now, just return 0
	return 0
}
*/
