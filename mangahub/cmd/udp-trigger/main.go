// package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"flag"
// 	"fmt"
// 	"net/http"
// 	"time"
// )

// // Simple HTTP client to trigger UDP notifications via the UDP server's HTTP interface
// // This assumes the UDP server exposes an HTTP endpoint for triggering broadcasts

// func main() {
// 	mangaID := flag.String("manga", "one-piece", "Manga ID")
// 	mangaTitle := flag.String("title", "One Piece", "Manga title")
// 	chapter := flag.Int("chapter", 1, "Chapter number")
// 	message := flag.String("message", "", "Custom message (optional)")
// 	notifType := flag.String("type", "chapter_release", "Notification type")
// 	serverHost := flag.String("host", "localhost:8082", "UDP server HTTP trigger port")
// 	flag.Parse()

// 	fmt.Println("=== UDP Notification Trigger ===")
// 	fmt.Printf("Manga: %s (%s)\n", *mangaTitle, *mangaID)
// 	fmt.Printf("Chapter: %d\n", *chapter)
// 	fmt.Printf("Type: %s\n", *notifType)
// 	fmt.Println()

// 	// Build notification message
// 	var notificationMsg string
// 	switch *notifType {
// 	case "chapter_release":
// 		notificationMsg = fmt.Sprintf("New chapter %d released for %s", *chapter, *mangaTitle)
// 	case "manga_update":
// 		if *message == "" {
// 			notificationMsg = fmt.Sprintf("Update for %s", *mangaTitle)
// 		} else {
// 			notificationMsg = *message
// 		}
// 	default:
// 		notificationMsg = *message
// 	}

// 	// Prepare notification payload
// 	notification := map[string]interface{}{
// 		"type":      *notifType,
// 		"manga_id":  *mangaID,
// 		"message":   notificationMsg,
// 		"timestamp": time.Now().Unix(),
// 	}

// 	jsonData, err := json.Marshal(notification)
// 	if err != nil {
// 		fmt.Printf("❌ Error marshaling JSON: %v\n", err)
// 		return
// 	}

// 	// Send to UDP server's HTTP trigger endpoint
// 	url := fmt.Sprintf("http://%s/trigger", *serverHost)
// 	fmt.Printf("📡 Sending to %s...\n", url)

// 	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		fmt.Printf("❌ Error: %v\n", err)
// 		fmt.Println("\n💡 Make sure the UDP server is running with HTTP trigger enabled")
// 		fmt.Println("   Or use the manual test method below:")
// 		printManualTestInstructions()
// 		return
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == 200 {
// 		fmt.Println("✅ Notification triggered successfully!")
// 	} else {
// 		fmt.Printf("⚠️  Server returned status %d\n", resp.StatusCode)
// 	}
// }

// func printManualTestInstructions() {
// 	fmt.Println("\n" + "=".Repeat(60))
// 	fmt.Println("📝 Manual Testing Instructions")
// 	fmt.Println("=".Repeat(60))
// 	fmt.Println("\nSince automated broadcast triggering isn't set up yet, you can:")
// 	fmt.Println("\n1. Test TCP Sync (Works NOW):")
// 	fmt.Println("   - Open 2 CLI clients")
// 	fmt.Println("   - Login with different accounts")
// 	fmt.Println("   - Update progress in one client")
// 	fmt.Println("   - See real-time sync in the other! ✅")
// 	fmt.Println("\n2. Test UDP Notifications (Manual):")
// 	fmt.Println("   - Modify internal/udp/udp.go to add BroadcastFromServer()")
// 	fmt.Println("   - Or integrate into API server when chapters are added")
// 	fmt.Println("   - Or add HTTP trigger endpoint to UDP server")
// 	fmt.Println("\n3. For Demo Purposes:")
// 	fmt.Println("   - Show TCP sync working (it does!)")
// 	fmt.Println("   - Explain UDP architecture (clients registered)")
// 	fmt.Println("   - Show UDP server accepting registrations")
// 	fmt.Println("   - Note: Full UDP broadcast needs API integration")
// }
