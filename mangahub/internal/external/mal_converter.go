package external

import (
	"fmt"
	"mangahub/pkg/models"
	"strings"
	"time"
)

// ConvertMALToManga converts official MAL API manga to our internal Manga model
func ConvertMALToManga(mal *MALMangaNode) *models.Manga {
	// Extract genres
	genres := make([]string, 0)
	for _, g := range mal.Genres {
		genres = append(genres, g.Name)
	}

	// Get author name
	author := "Unknown"
	if len(mal.Authors) > 0 {
		firstName := mal.Authors[0].Node.FirstName
		lastName := mal.Authors[0].Node.LastName
		author = strings.TrimSpace(firstName + " " + lastName)
		if author == "" {
			author = "Unknown"
		}
	}

	// Get cover URL
	coverURL := ""
	if mal.MainPicture.Large != "" {
		coverURL = mal.MainPicture.Large
	} else if mal.MainPicture.Medium != "" {
		coverURL = mal.MainPicture.Medium
	}

	// Convert status
	status := "unknown"
	switch strings.ToLower(mal.Status) {
	case "finished":
		status = "completed"
	case "currently_publishing":
		status = "ongoing"
	case "not_yet_published":
		status = "upcoming"
	case "on_hiatus":
		status = "hiatus"
	default:
		if mal.Status != "" {
			status = strings.ToLower(mal.Status)
		}
	}

	// Get publication year from start date
	publicationYear := 0
	if mal.StartDate != "" {
		if t, err := time.Parse("2006-01-02", mal.StartDate); err == nil {
			publicationYear = t.Year()
		}
	}

	// Use English title if available, otherwise use main title
	title := mal.Title
	if mal.AlternativeTitles.En != "" {
		title = mal.AlternativeTitles.En
	}

	return &models.Manga{
		ID:              fmt.Sprintf("mal-%d", mal.ID),
		Title:           title,
		Author:          author,
		Genres:          genres,
		Status:          status,
		TotalChapters:   mal.NumChapters,
		Description:     mal.Synopsis,
		CoverURL:        coverURL,
		PublicationYear: publicationYear,
		Rating:          mal.Mean,
		CreatedAt:       time.Now(),
	}
}

// ConvertMALListToManga converts a list of official MAL manga to internal format
func ConvertMALListToManga(malList []MALMangaNode) []*models.Manga {
	manga := make([]*models.Manga, 0, len(malList))
	for i := range malList {
		manga = append(manga, ConvertMALToManga(&malList[i]))
	}
	return manga
}
