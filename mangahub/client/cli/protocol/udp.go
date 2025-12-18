package protocol

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Notification struct {
	Type      string `json:"type"`
	MangaID   string `json:"manga_id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// connectUDP attempts to connect to the UDP server for notifications
func (c *Client) ConnectUDP() {
	// Resolve UDP server address
	serverAddr, err := net.ResolveUDPAddr("udp", udpAddr)
	if err != nil {
		fmt.Println(colorYellow + "⚠️  UDP notifications unavailable (server offline)" + colorReset)
		c.udpEnabled = false
		return
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println(colorYellow + "⚠️  UDP notifications unavailable (connection failed)" + colorReset)
		c.udpEnabled = false
		return
	}

	c.udpConn = conn

	// Send REGISTER message to server
	_, err = conn.Write([]byte("REGISTER"))
	if err != nil {
		fmt.Println(colorYellow + "⚠️  Failed to register for notifications" + colorReset)
		conn.Close()
		c.udpEnabled = false
		return
	}

	// Wait for acknowledgment with timeout
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil || string(buffer[:n]) != "REGISTERED" {
		fmt.Println(colorYellow + "⚠️  UDP registration failed" + colorReset)
		conn.Close()
		c.udpEnabled = false
		return
	}

	c.udpEnabled = true
	fmt.Println(colorGreen + "✅ Connected to notification server" + colorReset)

	// Start listening for notifications in background
	go c.listenUDPNotifications()
}

// listenUDPNotifications listens for chapter release and manga update notifications
func (c *Client) listenUDPNotifications() {
	if c.udpConn == nil {
		return
	}

	buffer := make([]byte, 2048)
	for {
		// Remove read deadline for continuous listening
		c.udpConn.SetReadDeadline(time.Time{})

		n, err := c.udpConn.Read(buffer)
		if err != nil {
			// Connection closed or error occurred
			c.udpEnabled = false
			c.udpConn = nil
			return
		}

		message := string(buffer[:n])

		// Handle PING from server (heartbeat check)
		if message == "PING" {
			// Respond with PONG to indicate we're alive
			_, err := c.udpConn.Write([]byte("PONG"))
			if err != nil {
				fmt.Println(colorRed + "⚠️  Failed to respond to heartbeat" + colorReset)
			}
			continue
		}

		// Parse notification
		var notification Notification
		if err := json.Unmarshal(buffer[:n], &notification); err != nil {
			continue
		}

		// Display notification to user
		c.DisplayNotification(notification)
	}
}

// displayNotification formats and displays a UDP notification
func (c *Client) DisplayNotification(notification Notification) {
	notifType := notification.Type
	message := notification.Message
	mangaID := notification.MangaID

	switch notifType {
	case "chapter_release":
		fmt.Printf("\n%s🔔 NEW CHAPTER! %s (Manga: %s)%s\n",
			colorCyan, message, mangaID, colorReset)
	case "manga_update":
		fmt.Printf("\n%s📢 UPDATE: %s (Manga: %s)%s\n",
			colorYellow, message, mangaID, colorReset)
	default:
		fmt.Printf("\n%s📬 Notification: %s%s\n",
			colorBlue, message, colorReset)
	}
}
