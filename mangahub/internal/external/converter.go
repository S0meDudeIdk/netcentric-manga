package external

import (
	"fmt"
	"mangahub/pkg/models"
	"strings"
	"time"
)

// ConvertJikanToManga converts a Jikan manga to our internal Manga model
func ConvertJikanToManga(jikan *JikanManga) *models.Manga {
	// Extract genres
	genres := make([]string, 0)
	for _, g := range jikan.Genres {
		genres = append(genres, g.Name)
	}
	for _, t := range jikan.Themes {
		genres = append(genres, t.Name)
	}
	for _, d := range jikan.Demographics {
		genres = append(genres, d.Name)
	}

	// Get author name
	author := "Unknown"
	if len(jikan.Authors) > 0 {
		author = jikan.Authors[0].Name
	}

	// Determine status
	status := "unknown"
	switch strings.ToLower(jikan.Status) {
	case "finished":
		status = "completed"
	case "publishing":
		status = "ongoing"
	case "on hiatus":
		status = "on_hold"
	case "discontinued":
		status = "dropped"
	default:
		status = "ongoing"
	}

	// Get publication year
	publicationYear := 0
	if jikan.Published.Prop.From.Year > 0 {
		publicationYear = jikan.Published.Prop.From.Year
	}

	// Create manga with just the numeric ID
	manga := &models.Manga{
		ID:              fmt.Sprintf("%d", jikan.MalID),
		Title:           jikan.Title,
		Author:          author,
		Genres:          genres,
		Status:          status,
		TotalChapters:   jikan.Chapters,
		Description:     jikan.Synopsis,
		CoverURL:        jikan.Images.JPG.LargeImageURL,
		PublicationYear: publicationYear,
		Rating:          jikan.Score,
		CreatedAt:       time.Now(),
	}

	// Set genres JSON
	if err := manga.SetGenres(genres); err == nil {
		// Genres set successfully
	}

	return manga
}

// ConvertJikanListToManga converts a list of Jikan manga to our internal format
func ConvertJikanListToManga(jikanList []JikanManga) []*models.Manga {
	mangaList := make([]*models.Manga, len(jikanList))
	for i, jikan := range jikanList {
		mangaList[i] = ConvertJikanToManga(&jikan)
	}
	return mangaList
}
