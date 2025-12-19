package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type ProgressUpdate struct {
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	MangaTitle string `json:"manga_title"`
	Chapter    int    `json:"chapter"`
	Timestamp  int64  `json:"timestamp"`
}

func (c *Client) listenTCPUpdates() {
	if c.tcpConn == nil {
		return
	}

	scanner := bufio.NewScanner(c.tcpConn)
	for scanner.Scan() {
		message := scanner.Text()

		var update ProgressUpdate
		if err := json.Unmarshal([]byte(message), &update); err != nil {
			continue
		}

		// Display progress update
		c.DisplayProgressUpdate(update)
	}

	if err := scanner.Err(); err != nil {
		// Connection lost
		c.tcpEnabled = false
		c.tcpConn = nil
	}
}

func (c *Client) ConnectTCP() {
	conn, err := net.DialTimeout("tcp", tcpAddr, 3*time.Second)
	if err != nil {
		fmt.Println(colorYellow + "⚠️  TCP sync unavailable (server offline)" + colorReset)
		c.tcpEnabled = false
		return
	}

	c.tcpConn = conn
	c.tcpEnabled = true
	fmt.Println(colorGreen + "✅ Connected to real-time sync server" + colorReset)

	// Start listening for updates in background
	go c.listenTCPUpdates()

	// Start keep-alive pings
	go c.tcpKeepAlive()
}

// tcpKeepAlive sends periodic PING messages to prevent timeout
func (c *Client) tcpKeepAlive() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !c.tcpEnabled || c.tcpConn == nil {
			return
		}

		// Send PING keep-alive message
		c.tcpConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_, err := c.tcpConn.Write([]byte("PING\n"))
		if err != nil {
			// Connection lost
			c.tcpEnabled = false
			c.tcpConn = nil
			return
		}
	}
}

// DisplayProgressUpdate formats and displays a TCP progress update
func (c *Client) DisplayProgressUpdate(update ProgressUpdate) {
	// Only show updates from other users
	if update.UserID != c.UserID {
		fmt.Printf("\n%s🔔 User update: %s is reading '%s' at chapter %d%s\n",
			colorCyan, update.Username, update.MangaTitle, update.Chapter, colorReset)
	}
}
