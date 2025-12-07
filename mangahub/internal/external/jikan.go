package external

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// getJikanBaseURL returns the Jikan API base URL from environment or default
func getJikanBaseURL() string {
	baseURL := os.Getenv("JIKAN_API_BASE_URL")
	if baseURL == "" {
		return "https://api.jikan.moe/v4" // Default
	}
	return baseURL
}

// getJikanRateLimit returns the rate limit duration from environment or default
func getJikanRateLimit() time.Duration {
	rateLimitStr := os.Getenv("JIKAN_RATE_LIMIT_SECONDS")
	if rateLimitStr != "" {
		if seconds, err := strconv.Atoi(rateLimitStr); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return time.Second // Default: 1 request per second
}

var lastRequestTime time.Time

// JikanClient handles requests to the Jikan API (MyAnimeList unofficial API)
type JikanClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewJikanClient creates a new Jikan API client
func NewJikanClient() *JikanClient {
	return &JikanClient{
		BaseURL: getJikanBaseURL(),
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// JikanMangaResponse represents the response from Jikan API for manga search
type JikanMangaResponse struct {
	Data       []JikanManga    `json:"data"`
	Pagination JikanPagination `json:"pagination"`
}

// JikanManga represents a manga entry from Jikan API
type JikanManga struct {
	MalID          int                  `json:"mal_id"`
	URL            string               `json:"url"`
	Images         JikanImages          `json:"images"`
	Approved       bool                 `json:"approved"`
	Titles         []JikanTitle         `json:"titles"`
	Title          string               `json:"title"`
	TitleEnglish   string               `json:"title_english"`
	TitleJapanese  string               `json:"title_japanese"`
	Type           string               `json:"type"`
	Chapters       int                  `json:"chapters"`
	Volumes        int                  `json:"volumes"`
	Status         string               `json:"status"`
	Publishing     bool                 `json:"publishing"`
	Published      JikanDateRange       `json:"published"`
	Score          float64              `json:"score"`
	Scored         float64              `json:"scored"`
	ScoredBy       int                  `json:"scored_by"`
	Rank           int                  `json:"rank"`
	Popularity     int                  `json:"popularity"`
	Members        int                  `json:"members"`
	Favorites      int                  `json:"favorites"`
	Synopsis       string               `json:"synopsis"`
	Background     string               `json:"background"`
	Authors        []JikanAuthor        `json:"authors"`
	Serializations []JikanSerialization `json:"serializations"`
	Genres         []JikanGenre         `json:"genres"`
	Themes         []JikanTheme         `json:"themes"`
	Demographics   []JikanDemographic   `json:"demographics"`
}

// JikanImages contains image URLs
type JikanImages struct {
	JPG  JikanImageSet `json:"jpg"`
	WebP JikanImageSet `json:"webp"`
}

// JikanImageSet contains different image sizes
type JikanImageSet struct {
	ImageURL      string `json:"image_url"`
	SmallImageURL string `json:"small_image_url"`
	LargeImageURL string `json:"large_image_url"`
}

// JikanTitle represents alternative titles
type JikanTitle struct {
	Type  string `json:"type"`
	Title string `json:"title"`
}

// JikanDateRange represents a date range
type JikanDateRange struct {
	From   string    `json:"from"`
	To     string    `json:"to"`
	Prop   JikanProp `json:"prop"`
	String string    `json:"string"`
}

// JikanProp contains date properties
type JikanProp struct {
	From JikanDate `json:"from"`
	To   JikanDate `json:"to"`
}

// JikanDate represents a date
type JikanDate struct {
	Day   int `json:"day"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

// JikanAuthor represents a manga author
type JikanAuthor struct {
	MalID int    `json:"mal_id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

// JikanSerialization represents where manga is serialized
type JikanSerialization struct {
	MalID int    `json:"mal_id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

// JikanGenre represents a genre
type JikanGenre struct {
	MalID int    `json:"mal_id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

// JikanTheme represents a theme
type JikanTheme struct {
	MalID int    `json:"mal_id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

// JikanDemographic represents a demographic
type JikanDemographic struct {
	MalID int    `json:"mal_id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

// JikanPagination contains pagination info
type JikanPagination struct {
	LastVisiblePage int                  `json:"last_visible_page"`
	HasNextPage     bool                 `json:"has_next_page"`
	CurrentPage     int                  `json:"current_page"`
	Items           JikanPaginationItems `json:"items"`
}

// JikanPaginationItems contains item counts
type JikanPaginationItems struct {
	Count   int `json:"count"`
	Total   int `json:"total"`
	PerPage int `json:"per_page"`
}

// respectRateLimit ensures we don't exceed API rate limits
func (c *JikanClient) respectRateLimit() {
	rateLimit := getJikanRateLimit()
	elapsed := time.Since(lastRequestTime)
	if elapsed < rateLimit {
		time.Sleep(rateLimit - elapsed)
	}
	lastRequestTime = time.Now()
}

// SearchManga searches for manga by query
func (c *JikanClient) SearchManga(query string, page int, limit int) (*JikanMangaResponse, error) {
	return c.SearchMangaWithSort(query, page, limit, "popularity", "asc")
}

// SearchMangaWithSort searches for manga by query with sorting options
func (c *JikanClient) SearchMangaWithSort(query string, page int, limit int, orderBy string, sort string) (*JikanMangaResponse, error) {
	c.respectRateLimit()

	if limit <= 0 {
		limit = 10
	}
	if limit > 25 {
		limit = 25 // Jikan API max is 25
	}
	if page <= 0 {
		page = 1
	}

	params := url.Values{}
	params.Add("q", query)
	params.Add("page", fmt.Sprintf("%d", page))
	params.Add("limit", fmt.Sprintf("%d", limit))

	if orderBy != "" {
		params.Add("order_by", orderBy)
	}
	if sort != "" {
		params.Add("sort", sort)
	}

	url := fmt.Sprintf("%s/manga?%s", c.BaseURL, params.Encode())

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manga: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result JikanMangaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetMangaByID retrieves a specific manga by MAL ID
func (c *JikanClient) GetMangaByID(malID int) (*JikanManga, error) {
	c.respectRateLimit()

	url := fmt.Sprintf("%s/manga/%d", c.BaseURL, malID)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manga: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data JikanManga `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Data, nil
}

// GetTopManga retrieves top manga from MAL
func (c *JikanClient) GetTopManga(page int, limit int) (*JikanMangaResponse, error) {
	return c.GetMangaWithSort(page, limit, "", "")
}

// GetMangaWithSort retrieves manga with sorting options
// orderBy: "mal_id", "title", "start_date", "end_date", "chapters", "volumes", "score", "scored_by", "rank", "popularity", "members", "favorites"
// sort: "asc" or "desc"
func (c *JikanClient) GetMangaWithSort(page int, limit int, orderBy string, sort string) (*JikanMangaResponse, error) {
	c.respectRateLimit()

	if limit <= 0 {
		limit = 10
	}
	if limit > 25 {
		limit = 25
	}
	if page <= 0 {
		page = 1
	}

	params := url.Values{}
	params.Add("page", fmt.Sprintf("%d", page))
	params.Add("limit", fmt.Sprintf("%d", limit))

	// Add sorting parameters if provided
	if orderBy != "" {
		params.Add("order_by", orderBy)
	}
	if sort != "" {
		params.Add("sort", sort)
	}

	// Use /manga endpoint instead of /top/manga for more flexibility with sorting
	endpoint := "/manga"
	if orderBy == "" && sort == "" {
		// Use /top/manga for default behavior (sorted by popularity)
		endpoint = "/top/manga"
	}

	url := fmt.Sprintf("%s%s?%s", c.BaseURL, endpoint, params.Encode())

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manga: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result JikanMangaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetMangaRecommendations gets manga recommendations by MAL ID
func (c *JikanClient) GetMangaRecommendations(malID int) ([]JikanManga, error) {
	c.respectRateLimit()

	url := fmt.Sprintf("%s/manga/%d/recommendations", c.BaseURL, malID)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recommendations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			Entry JikanManga `json:"entry"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	recommendations := make([]JikanManga, len(result.Data))
	for i, item := range result.Data {
		recommendations[i] = item.Entry
	}

	return recommendations, nil
}
