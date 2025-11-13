package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	mangaID := flag.String("manga", "one-piece", "Manga ID")
	mangaTitle := flag.String("title", "One Piece", "Manga title")
	chapter := flag.Int("chapter", 1, "Chapter number")
	message := flag.String("message", "", "Custom message (optional)")
	notifType := flag.String("type", "chapter_release", "Notification type: chapter_release or manga_update")
	flag.Parse()

	fmt.Println("=== UDP Notification Broadcast Test ===")
	fmt.Printf("Manga: %s (%s)\n", *mangaTitle, *mangaID)
	fmt.Printf("Chapter: %d\n", *chapter)
	fmt.Printf("Type: %s\n", *notifType)
	fmt.Println()

	// Build notification message
	var notificationMsg string
	switch *notifType {
	case "chapter_release":
		notificationMsg = fmt.Sprintf("New chapter %d released for %s", *chapter, *mangaTitle)
	case "manga_update":
		if *message == "" {
			notificationMsg = fmt.Sprintf("Update for %s", *mangaTitle)
		} else {
			notificationMsg = *message
		}
	default:
		notificationMsg = *message
	}

	// Prepare notification payload
	notification := map[string]interface{}{
		"type":      *notifType,
		"manga_id":  *mangaID,
		"message":   notificationMsg,
		"timestamp": time.Now().Unix(),
	}

	jsonData, err := json.Marshal(notification)
	if err != nil {
		log.Fatalf("❌ Error marshaling JSON: %v", err)
	}

	// Send to UDP server's HTTP trigger endpoint
	url := "http://localhost:8082/trigger"
	fmt.Printf("📡 Triggering broadcast via %s...\n", url)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Error: %v\n", err)
		log.Println("\n⚠️  Make sure the UDP server is running:")
		log.Println("   go run ./cmd/udp-server")
		log.Println("\n   The UDP server now includes an HTTP trigger on port 8082")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		
		fmt.Println("✅ Notification broadcasted successfully!")
		if clients, ok := result["clients"].(float64); ok {
			fmt.Printf("📢 Sent to %d registered client(s)\n", int(clients))
		}
	} else {
		fmt.Printf("⚠️  Server returned status %d\n", resp.StatusCode)
	}

	fmt.Println("\n💡 Tip: Make sure CLI clients are logged in to receive notifications!")
}
