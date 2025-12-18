package external

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// MangaDexClient handles requests to the MangaDex API
type MangaDexClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Debug      bool
}

// NewMangaDexClient creates a new MangaDex API client
func NewMangaDexClient() *MangaDexClient {
	baseURL := os.Getenv("MANGADEX_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.mangadex.org"
	}

	timeoutStr := os.Getenv("MANGADEX_API_TIMEOUT")
	timeout := 15 * time.Second
	if timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
			timeout = time.Duration(t) * time.Second
		}
	}

	debug := os.Getenv("MANGADEX_DEBUG") == "true"
	apiKey := os.Getenv("MANGADEX_API_KEY")

	return &MangaDexClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Debug: debug,
	}
}

// MangaDexMangaResponse represents a manga search response
type MangaDexMangaResponse struct {
	Result   string          `json:"result"`
	Response string          `json:"response"`
	Data     []MangaDexManga `json:"data"`
	Limit    int             `json:"limit"`
	Offset   int             `json:"offset"`
	Total    int             `json:"total"`
}

// MangaDexManga represents a manga entry from MangaDex
type MangaDexManga struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Attributes MangaDexMangaAttributes `json:"attributes"`
}

// MangaDexMangaAttributes contains manga details
type MangaDexMangaAttributes struct {
	Title         map[string]string   `json:"title"`
	AltTitles     []map[string]string `json:"altTitles"`
	Description   map[string]string   `json:"description"`
	Status        string              `json:"status"`
	Year          int                 `json:"year"`
	ContentRating string              `json:"contentRating"`
	Tags          []MangaDexTag       `json:"tags"`
	LastVolume    string              `json:"lastVolume"`
	LastChapter   string              `json:"lastChapter"`
}

// MangaDexTag represents a manga tag/genre
type MangaDexTag struct {
	ID         string                `json:"id"`
	Type       string                `json:"type"`
	Attributes MangaDexTagAttributes `json:"attributes"`
}

// MangaDexTagAttributes contains tag details
type MangaDexTagAttributes struct {
	Name  map[string]string `json:"name"`
	Group string            `json:"group"`
}

// MangaDexChapterFeedResponse represents chapter list response
type MangaDexChapterFeedResponse struct {
	Result   string            `json:"result"`
	Response string            `json:"response"`
	Data     []MangaDexChapter `json:"data"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	Total    int               `json:"total"`
}

// MangaDexChapter represents a chapter entry
type MangaDexChapter struct {
	ID         string                    `json:"id"`
	Type       string                    `json:"type"`
	Attributes MangaDexChapterAttributes `json:"attributes"`
}

// MangaDexChapterAttributes contains chapter details
type MangaDexChapterAttributes struct {
	Volume             *string   `json:"volume"`
	Chapter            *string   `json:"chapter"`
	Title              string    `json:"title"`
	TranslatedLanguage string    `json:"translatedLanguage"`
	PublishAt          time.Time `json:"publishAt"`
	ReadableAt         time.Time `json:"readableAt"`
	Pages              int       `json:"pages"`
	Version            int       `json:"version"`
	ExternalUrl        *string   `json:"externalUrl"` // URL to external site for licensed manga
}

// MangaDexAtHomeResponse represents the at-home server response
type MangaDexAtHomeResponse struct {
	Result  string               `json:"result"`
	BaseUrl string               `json:"baseUrl"`
	Chapter MangaDexChapterFiles `json:"chapter"`
}

// MangaDexChapterFiles contains the image file names
type MangaDexChapterFiles struct {
	Hash      string   `json:"hash"`
	Data      []string `json:"data"`
	DataSaver []string `json:"dataSaver"`
}

// addAuthHeaders adds authentication headers to the request if API key is available
func (c *MangaDexClient) addAuthHeaders(req *http.Request) {
	if c.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	}
	req.Header.Set("User-Agent", "MangaHub/1.0")
}

// SearchMangaByTitle searches for manga by title
func (c *MangaDexClient) SearchMangaByTitle(title string, limit int) (*MangaDexMangaResponse, error) {
	params := url.Values{}
	params.Add("title", title)
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("order[relevance]", "desc")
	params.Add("includes[]", "cover_art")

	url := fmt.Sprintf("%s/manga?%s", c.BaseURL, params.Encode())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.addAuthHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result MangaDexMangaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetMangaByID retrieves a manga by its MangaDex ID
func (c *MangaDexClient) GetMangaByID(mangaID string) (*MangaDexManga, error) {
	params := url.Values{}
	params.Add("includes[]", "cover_art")
	params.Add("includes[]", "author")
	params.Add("includes[]", "artist")

	url := fmt.Sprintf("%s/manga/%s?%s", c.BaseURL, mangaID, params.Encode())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.addAuthHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Result   string        `json:"result"`
		Response string        `json:"response"`
		Data     MangaDexManga `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Data, nil
}

// GetMangaChapterFeed retrieves the chapter list for a manga
func (c *MangaDexClient) GetMangaChapterFeed(mangaID string, limit, offset int, translatedLanguage []string) (*MangaDexChapterFeedResponse, error) {
	params := url.Values{}
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", fmt.Sprintf("%d", offset))
	params.Add("order[chapter]", "asc")
	params.Add("contentRating[]", "safe")
	params.Add("contentRating[]", "suggestive")
	params.Add("contentRating[]", "erotica")

	// Add translated languages
	if len(translatedLanguage) == 0 {
		translatedLanguage = []string{"en"}
	}
	for _, lang := range translatedLanguage {
		params.Add("translatedLanguage[]", lang)
	}

	url := fmt.Sprintf("%s/manga/%s/feed?%s", c.BaseURL, mangaID, params.Encode())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.addAuthHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result MangaDexChapterFeedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetChapterPages retrieves the page URLs for a chapter
func (c *MangaDexClient) GetChapterPages(chapterID string) (*MangaDexAtHomeResponse, error) {
	url := fmt.Sprintf("%s/at-home/server/%s", c.BaseURL, chapterID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.addAuthHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result MangaDexAtHomeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetTitle returns the English title or first available title
func (m *MangaDexManga) GetTitle() string {
	if title, ok := m.Attributes.Title["en"]; ok {
		return title
	}
	// Return first available title
	for _, title := range m.Attributes.Title {
		return title
	}
	return "Unknown"
}

// GetDescription returns the English description or first available description
func (m *MangaDexManga) GetDescription() string {
	if desc, ok := m.Attributes.Description["en"]; ok {
		return desc
	}
	// Return first available description
	for _, desc := range m.Attributes.Description {
		return desc
	}
	return ""
}

// GetGenres returns the list of genre tags
func (m *MangaDexManga) GetGenres() []string {
	genres := []string{}
	for _, tag := range m.Attributes.Tags {
		if tag.Attributes.Group == "genre" {
			if name, ok := tag.Attributes.Name["en"]; ok {
				genres = append(genres, name)
			}
		}
	}
	return genres
}

// GetChapterNumber safely returns the chapter number as a string
func (c *MangaDexChapter) GetChapterNumber() string {
	if c.Attributes.Chapter != nil {
		return *c.Attributes.Chapter
	}
	return "0"
}

// GetVolumeNumber safely returns the volume number as a string
func (c *MangaDexChapter) GetVolumeNumber() string {
	if c.Attributes.Volume != nil {
		return *c.Attributes.Volume
	}
	return ""
}

// BuildPageURL constructs the full URL for a chapter page
func BuildMangaDexPageURL(baseURL, hash, filename string, dataSaver bool) string {
	quality := "data"
	if dataSaver {
		quality = "data-saver"
	}
	return fmt.Sprintf("%s/%s/%s/%s", baseURL, quality, hash, filename)
}

// SearchManga searches for manga on MangaDex by title
func (c *MangaDexClient) SearchManga(title string, limit int) (*MangaDexMangaResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	params := url.Values{}
	params.Add("title", title)
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", "0")
	params.Add("order[relevance]", "desc")
	params.Add("contentRating[]", "safe")
	params.Add("contentRating[]", "suggestive")
	params.Add("contentRating[]", "erotica")
	params.Add("includes[]", "cover_art")

	searchURL := fmt.Sprintf("%s/manga?%s", c.BaseURL, params.Encode())

	if c.Debug {
		fmt.Printf("[MangaDex] Searching manga: %s\n", searchURL)
	}

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	c.addAuthHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result MangaDexMangaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if c.Debug {
		fmt.Printf("[MangaDex] Found %d results for '%s'\n", result.Total, title)
	}

	return &result, nil
}

// ExtractMangaDexID extracts MangaDex ID from various formats
func ExtractMangaDexID(identifier string) (string, error) {
	// Remove any URL prefix
	identifier = strings.TrimSpace(identifier)

	// If it's a full URL, extract the ID
	if strings.Contains(identifier, "mangadex.org") {
		parts := strings.Split(identifier, "/")
		for i, part := range parts {
			if part == "title" && i+1 < len(parts) {
				return parts[i+1], nil
			}
		}
		return "", fmt.Errorf("invalid MangaDex URL")
	}

	// If it has "mangadex-" prefix, remove it
	identifier = strings.TrimPrefix(identifier, "mangadex-")

	// Validate UUID format (MangaDex uses UUIDs)
	if len(identifier) == 36 && strings.Count(identifier, "-") == 4 {
		return identifier, nil
	}

	return "", fmt.Errorf("invalid MangaDex ID format")
}
