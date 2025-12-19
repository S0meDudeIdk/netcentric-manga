package protocol

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	baseURL = "http://localhost:8080"
	tcpAddr = "localhost:9001"
	udpAddr = "localhost:9002"
	apiURL  = baseURL + "/api/v1"
	// baseURL = "http://10.11.240.116:8080"
	// tcpAddr = "10.11.240.116:9001"
	// udpAddr = "10.11.240.116:9002"
)

// Color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

// Client represents the MangaHub CLI client
type Client struct {
	Token       string
	Username    string
	Email       string
	UserID      string
	scanner     *bufio.Scanner
	tcpConn     net.Conn
	tcpEnabled  bool
	udpConn     *net.UDPConn
	udpEnabled  bool
	wsConn      *websocket.Conn
	wsEnabled   bool
	currentRoom string
}

func NewClient() *Client {
	return &Client{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// Manga represents a manga entry
type Manga struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Author        string    `json:"author"`
	Genres        []string  `json:"genres"`
	Status        string    `json:"status"`
	TotalChapters int       `json:"total_chapters"`
	Description   string    `json:"description"`
	CoverURL      string    `json:"cover_url"`
	CreatedAt     time.Time `json:"created_at"`
}

// UserProgress represents user's reading progress
type UserProgress struct {
	UserID         string    `json:"user_id"`
	MangaID        string    `json:"manga_id"`
	CurrentChapter int       `json:"current_chapter"`
	Status         string    `json:"status"`
	LastUpdated    time.Time `json:"last_updated"`
}

func (c *Client) ShowWelcome() {
	fmt.Println(colorCyan + "╔════════════════════════════════════════╗")
	fmt.Println("║                                        ║")
	fmt.Println("║         MangaHub CLI Client            ║")
	fmt.Println("║         v1.0.0                         ║")
	fmt.Println("║                                        ║")
	fmt.Println("╚════════════════════════════════════════╝" + colorReset)
	fmt.Println()
}

func (c *Client) MainMenu() {
	for {
		if c.Token == "" {
			c.AuthMenu()
		} else {
			c.UserMenu()
		}
	}
}

func (c *Client) AuthMenu() {
	fmt.Println(colorYellow + "\n📚 Authentication Menu" + colorReset)
	fmt.Println("1. Login")
	fmt.Println("2. Register")
	fmt.Println("3. Exit")
	fmt.Print("\nSelect an option: ")

	choice := c.readInput()
	fmt.Println()

	switch choice {
	case "1":
		c.Login()
	case "2":
		c.Register()
	case "3":
		fmt.Println(colorGreen + "Goodbye! 👋" + colorReset)
		os.Exit(0)
	default:
		fmt.Println(colorRed + "❌ Invalid option" + colorReset)
	}
}

func (c *Client) UserMenu() {
	fmt.Println(colorYellow + "\n📚 Main Menu" + colorReset)
	fmt.Printf(colorCyan+"Logged in as: %s (%s)\n"+colorReset, c.Username, c.Email)

	// Show TCP sync status
	if c.tcpEnabled {
		fmt.Printf(colorGreen + "📡 Real-time sync: ENABLED\n" + colorReset)
	} else {
		fmt.Printf(colorYellow + "📡 Real-time sync: OFFLINE\n" + colorReset)
	}

	// Show UDP notification status
	if c.udpEnabled {
		fmt.Printf(colorGreen + "🔔 Notifications: ENABLED\n" + colorReset)
	} else {
		fmt.Printf(colorYellow + "🔔 Notifications: OFFLINE\n" + colorReset)
	}

	// Show WebSocket status
	if c.wsEnabled {
		fmt.Printf(colorGreen+"💬 Chat: CONNECTED (Room: %s)\n"+colorReset, c.currentRoom)
	} else {
		fmt.Printf(colorYellow + "💬 Chat: DISCONNECTED\n" + colorReset)
	}

	fmt.Println("\n1. Browse Manga")
	fmt.Println("2. Search Manga")
	fmt.Println("3. Search MyAnimeList")
	fmt.Println("4. My Library")
	fmt.Println("5. Get Recommendations")
	fmt.Println("6. Join General Chat")
	fmt.Println("7. Logout")
	fmt.Print("\nSelect an option: ")

	choice := c.readInput()
	fmt.Println()

	switch choice {
	case "1":
		c.BrowseManga()
	case "2":
		c.SearchManga()
	case "3":
		c.SearchMAL()
	case "4":
		c.MyLibrary()
	case "5":
		c.GetRecommendations()
	case "6":
		c.JoinChatHub(baseURL, "general", "General Chat")
	case "7":
		c.Logout()
	default:
		fmt.Println(colorRed + "❌ Invalid option" + colorReset)
	}
}

func (c *Client) Login() {
	fmt.Println(colorCyan + "🔐 Login" + colorReset)
	fmt.Print("Email: ")
	email := c.readInput()
	fmt.Print("Password: ")
	password := c.readInput()

	data := map[string]string{
		"email":    email,
		"password": password,
	}

	resp, err := c.makeRequest("POST", apiURL+"/auth/login", data, false)
	if err != nil {
		fmt.Println(colorRed + "❌ Login failed: " + err.Error() + colorReset)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		fmt.Println(colorRed + "❌ Error parsing response" + colorReset)
		return
	}

	if token, ok := result["token"].(string); ok {
		c.Token = token
		if user, ok := result["user"].(map[string]interface{}); ok {
			c.Username = user["username"].(string)
			c.Email = user["email"].(string)
			if id, ok := user["id"].(string); ok {
				c.UserID = id
			}
		}
		fmt.Println(colorGreen + "✅ Login successful!" + colorReset)

		// Try to connect to TCP server for real-time sync
		c.ConnectTCP()

		// Try to connect to UDP server for notifications
		c.ConnectUDP()
	} else {
		fmt.Println(colorRed + "❌ Login failed" + colorReset)
	}
}

func (c *Client) Register() {
	fmt.Println(colorCyan + "📝 Register" + colorReset)
	fmt.Println(colorYellow + "\nPassword Requirements:" + colorReset)
	fmt.Println("  • Minimum 6 characters")
	fmt.Println("  • Username: 3-30 characters")
	fmt.Println()

	fmt.Print("Username: ")
	username := c.readInput()

	// Validate username length
	if len(username) < 3 || len(username) > 30 {
		fmt.Println(colorRed + "❌ Username must be between 3 and 30 characters" + colorReset)
		return
	}

	fmt.Print("Email: ")
	email := c.readInput()

	// Basic email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		fmt.Println(colorRed + "❌ Please enter a valid email address" + colorReset)
		return
	}

	fmt.Print("Password (min 6 characters): ")
	password := c.readInput()

	// Validate password length
	if len(password) < 6 {
		fmt.Println(colorRed + "❌ Password must be at least 6 characters" + colorReset)
		return
	}

	data := map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}

	resp, err := c.makeRequest("POST", apiURL+"/auth/register", data, false)
	if err != nil {
		// Parse validation errors from the API
		errMsg := err.Error()
		if strings.Contains(errMsg, "Password") && strings.Contains(errMsg, "min") {
			fmt.Println(colorRed + "❌ Password must be at least 6 characters" + colorReset)
		} else if strings.Contains(errMsg, "Username") && strings.Contains(errMsg, "min") {
			fmt.Println(colorRed + "❌ Username must be at least 3 characters" + colorReset)
		} else if strings.Contains(errMsg, "already exists") {
			fmt.Println(colorRed + "❌ User with this email or username already exists" + colorReset)
		} else {
			fmt.Println(colorRed + "❌ Registration failed: " + errMsg + colorReset)
		}
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		fmt.Println(colorRed + "❌ Error parsing response" + colorReset)
		return
	}

	fmt.Println(colorGreen + "✅ Registration successful! You can now login." + colorReset)
}

func (c *Client) BrowseManga() {
	fmt.Println(colorCyan + "📖 Browse Popular Manga" + colorReset)
	fmt.Print("How many results? (default 10): ")
	limitStr := c.readInput()
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	url := fmt.Sprintf("%s/manga/popular?limit=%d", apiURL, limit)
	resp, err := c.makeRequest("GET", url, nil, true)
	if err != nil {
		fmt.Println(colorRed + "❌ Error: " + err.Error() + colorReset)
		return
	}

	var result struct {
		Manga []Manga `json:"manga"`
		Count int     `json:"count"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		fmt.Println(colorRed + "❌ Error parsing response" + colorReset)
		return
	}

	fmt.Printf("\n%s📚 Found %d manga:%s\n\n", colorGreen, result.Count, colorReset)
	for i, manga := range result.Manga {
		c.DisplayManga(i+1, manga)
	}

	fmt.Print("\nEnter manga number to view details (or press Enter to return): ")
	choice := c.readInput()
	if choice != "" {
		if idx, err := strconv.Atoi(choice); err == nil && idx > 0 && idx <= len(result.Manga) {
			c.ViewMangaDetails(result.Manga[idx-1])
		}
	}
}

func (c *Client) SearchManga() {
	fmt.Println(colorCyan + "🔍 Search Manga" + colorReset)
	fmt.Print("Enter search query: ")
	query := c.readInput()

	if query == "" {
		return
	}

	url := fmt.Sprintf("%s/manga?query=%s", apiURL, query)
	resp, err := c.makeRequest("GET", url, nil, true)
	if err != nil {
		fmt.Println(colorRed + "❌ Error: " + err.Error() + colorReset)
		return
	}

	var result struct {
		Manga []Manga `json:"manga"`
		Count int     `json:"count"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		fmt.Println(colorRed + "❌ Error parsing response" + colorReset)
		return
	}

	fmt.Printf("\n%s🔍 Found %d manga matching '%s':%s\n\n", colorGreen, result.Count, query, colorReset)
	for i, manga := range result.Manga {
		c.DisplayManga(i+1, manga)
	}

	fmt.Print("\nEnter manga number to view details (or press Enter to return): ")
	choice := c.readInput()
	if choice != "" {
		if idx, err := strconv.Atoi(choice); err == nil && idx > 0 && idx <= len(result.Manga) {
			c.ViewMangaDetails(result.Manga[idx-1])
		}
	}
}

func (c *Client) SearchMAL() {
	fmt.Println(colorCyan + "🌐 Search MyAnimeList" + colorReset)
	fmt.Print("Enter search query: ")
	query := c.readInput()

	if query == "" {
		return
	}

	url := fmt.Sprintf("%s/manga/mal/search?q=%s", apiURL, query)
	resp, err := c.makeRequest("GET", url, nil, false) // MAL search is public
	if err != nil {
		fmt.Println(colorRed + "❌ Error: " + err.Error() + colorReset)
		return
	}

	var result struct {
		Data  []Manga `json:"data"`
		Total int     `json:"total"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		fmt.Println(colorRed + "❌ Error parsing response" + colorReset)
		return
	}

	fmt.Printf("\n%s🌐 Found %d manga on MyAnimeList matching '%s':%s\n\n", colorGreen, result.Total, query, colorReset)
	for i, manga := range result.Data {
		c.DisplayManga(i+1, manga)
	}

	fmt.Print("\nEnter manga number to view details (or press Enter to return): ")
	choice := c.readInput()
	if choice != "" {
		if idx, err := strconv.Atoi(choice); err == nil && idx > 0 && idx <= len(result.Data) {
			c.ViewMALMangaDetails(result.Data[idx-1])
		}
	}
}

func (c *Client) ViewMALMangaDetails(manga Manga) {
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Printf("%s📖 %s%s\n", colorCyan, manga.Title, colorReset)
	fmt.Printf("%s🌐 Source: MyAnimeList%s\n", colorYellow, colorReset)
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("%s✍️  Author:%s %s\n", colorYellow, colorReset, manga.Author)
	fmt.Printf("%s📊 Status:%s %s\n", colorYellow, colorReset, manga.Status)
	fmt.Printf("%s📚 Chapters:%s %d\n", colorYellow, colorReset, manga.TotalChapters)
	fmt.Printf("%s🏷️  Genres:%s %s\n", colorYellow, colorReset, strings.Join(manga.Genres, ", "))
	fmt.Printf("\n%s📝 Description:%s\n%s\n", colorYellow, colorReset, manga.Description)
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println("\nPress Enter to return...")
	c.readInput()
}

func (c *Client) ViewMangaDetails(manga Manga) {
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Printf("%s📖 %s%s\n", colorCyan, manga.Title, colorReset)
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("%s✍️  Author:%s %s\n", colorYellow, colorReset, manga.Author)
	fmt.Printf("%s📊 Status:%s %s\n", colorYellow, colorReset, manga.Status)
	fmt.Printf("%s📚 Chapters:%s %d\n", colorYellow, colorReset, manga.TotalChapters)
	fmt.Printf("%s🏷️  Genres:%s %s\n", colorYellow, colorReset, strings.Join(manga.Genres, ", "))
	fmt.Printf("\n%s📝 Description:%s\n%s\n", colorYellow, colorReset, manga.Description)
	fmt.Println(strings.Repeat("═", 60))

	fmt.Println("\n1. Add to Library")
	fmt.Println("2. Join Chat Hub")
	fmt.Println("3. Back")
	fmt.Print("\nSelect an option: ")

	choice := c.readInput()
	switch choice {
	case "1":
		c.AddToLibrary(manga.ID)
	case "2":
		c.JoinChatHub(baseURL, manga.ID, manga.Title)
	}
}

func (c *Client) AddToLibrary(mangaID string) {
	fmt.Println("\nSelect status:")
	fmt.Println("1. Reading")
	fmt.Println("2. Plan to Read")
	fmt.Println("3. Completed")
	fmt.Println("4. Dropped")
	fmt.Print("\nChoice: ")

	choice := c.readInput()
	statusMap := map[string]string{
		"1": "reading",
		"2": "plan_to_read",
		"3": "completed",
		"4": "dropped",
	}

	status, ok := statusMap[choice]
	if !ok {
		fmt.Println(colorRed + "❌ Invalid status" + colorReset)
		return
	}

	data := map[string]string{
		"manga_id": mangaID,
		"status":   status,
	}

	_, err := c.makeRequest("POST", apiURL+"/users/library", data, true)
	if err != nil {
		fmt.Println(colorRed + "❌ Error: " + err.Error() + colorReset)
		return
	}

	fmt.Println(colorGreen + "✅ Added to library!" + colorReset)
}

func (c *Client) MyLibrary() {
	fmt.Println(colorCyan + "📚 My Library" + colorReset)

	resp, err := c.makeRequest("GET", apiURL+"/users/library", nil, true)
	if err != nil {
		fmt.Println(colorRed + "❌ Error: " + err.Error() + colorReset)
		return
	}

	var library map[string][]UserProgress
	if err := json.Unmarshal(resp, &library); err != nil {
		fmt.Println(colorRed + "❌ Error parsing response" + colorReset)
		return
	}

	categories := []struct {
		Name   string
		Color  string
		Status string
	}{
		{"📖 Reading", colorGreen, "reading"},
		{"✅ Completed", colorBlue, "completed"},
		{"📋 Plan to Read", colorYellow, "plan_to_read"},
		{"❌ Dropped", colorRed, "dropped"},
	}

	for _, cat := range categories {
		if items, ok := library[cat.Status]; ok && len(items) > 0 {
			fmt.Printf("\n%s%s (%d)%s\n", cat.Color, cat.Name, len(items), colorReset)
			for i, item := range items {
				fmt.Printf("  %d. %s (Chapter %d)\n", i+1, item.MangaID, item.CurrentChapter)
			}
		}
	}

	fmt.Println("\n1. Update Reading Progress")
	fmt.Println("2. View Library Stats")
	fmt.Println("3. Back")
	fmt.Print("\nSelect an option: ")

	choice := c.readInput()
	switch choice {
	case "1":
		c.UpdateProgress()
	case "2":
		c.ViewLibraryStats()
	}
}

func (c *Client) UpdateProgress() {
	fmt.Print("\nManga ID: ")
	mangaID := c.readInput()
	fmt.Print("Current Chapter: ")
	chapterStr := c.readInput()
	chapter, _ := strconv.Atoi(chapterStr)

	fmt.Println("\nSelect status:")
	fmt.Println("1. Reading")
	fmt.Println("2. Completed")
	fmt.Print("\nChoice: ")
	choice := c.readInput()

	status := "reading"
	if choice == "2" {
		status = "completed"
	}

	data := map[string]interface{}{
		"manga_id":        mangaID,
		"current_chapter": chapter,
		"status":          status,
	}

	_, err := c.makeRequest("PUT", apiURL+"/users/progress", data, true)
	if err != nil {
		fmt.Println(colorRed + "❌ Error: " + err.Error() + colorReset)
		return
	}

	fmt.Println(colorGreen + "✅ Progress updated!" + colorReset)

	// Sync progress to TCP server for real-time updates
	if c.tcpEnabled {
		// c.SyncProgress(mangaID, chapter)
		fmt.Println(colorCyan + "📡 Progress synced to other clients" + colorReset)
	}
}

func (c *Client) ViewLibraryStats() {
	resp, err := c.makeRequest("GET", apiURL+"/users/library/stats", nil, true)
	if err != nil {
		fmt.Println(colorRed + "❌ Error: " + err.Error() + colorReset)
		return
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(resp, &stats); err != nil {
		fmt.Println(colorRed + "❌ Error parsing response" + colorReset)
		return
	}

	fmt.Println("\n" + strings.Repeat("═", 40))
	fmt.Printf("%s📊 Library Statistics%s\n", colorCyan, colorReset)
	fmt.Println(strings.Repeat("═", 40))
	for key, value := range stats {
		fmt.Printf("%s: %v\n", key, value)
	}
	fmt.Println(strings.Repeat("═", 40))
}

func (c *Client) GetRecommendations() {
	fmt.Println(colorCyan + "💡 Recommendations" + colorReset)

	resp, err := c.makeRequest("GET", apiURL+"/users/recommendations?limit=5", nil, true)
	if err != nil {
		fmt.Println(colorRed + "❌ Error: " + err.Error() + colorReset)
		return
	}

	var result struct {
		Recommendations []Manga `json:"recommendations"`
		Count           int     `json:"count"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		fmt.Println(colorRed + "❌ Error parsing response" + colorReset)
		return
	}

	fmt.Printf("\n%s💡 We recommend these %d manga for you:%s\n\n", colorGreen, result.Count, colorReset)
	for i, manga := range result.Recommendations {
		c.DisplayManga(i+1, manga)
	}

	fmt.Print("\nPress Enter to continue...")
	c.readInput()
}

func (c *Client) DisplayManga(num int, manga Manga) {
	fmt.Printf("%s%d. %s%s\n", colorCyan, num, manga.Title, colorReset)
	fmt.Printf("   %s✍️  %s | 📚 %d chapters | 🏷️  %s%s\n",
		colorYellow, manga.Author, manga.TotalChapters, strings.Join(manga.Genres, ", "), colorReset)
}

func (c *Client) Logout() {
	c.Token = ""
	c.Username = ""
	c.Email = ""
	c.UserID = ""

	// Disconnect from TCP server
	if c.tcpConn != nil {
		c.tcpConn.Close()
		c.tcpConn = nil
		c.tcpEnabled = false
	}

	// Disconnect from UDP server
	if c.udpConn != nil {
		// Send UNREGISTER message
		c.udpConn.Write([]byte("UNREGISTER"))
		c.udpConn.Close()
		c.udpConn = nil
		c.udpEnabled = false
	}

	// Disconnect from WebSocket
	if c.wsConn != nil {
		c.wsConn.Close()
		c.wsConn = nil
		c.wsEnabled = false
		c.currentRoom = ""
	}

	fmt.Println(colorGreen + "✅ Logged out successfully" + colorReset)
}

func (c *Client) readInput() string {
	c.scanner.Scan()
	return strings.TrimSpace(c.scanner.Text())
}

func (c *Client) makeRequest(method, url string, data interface{}, auth bool) ([]byte, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if auth && c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			if errMsg, ok := errResp["error"].(string); ok {
				return nil, fmt.Errorf("%s", errMsg)
			}
		}
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	return respBody, nil
}
