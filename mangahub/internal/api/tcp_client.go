package api

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"time"
)

// TCPProgressUpdate represents a progress update from TCP server
type TCPProgressUpdate struct {
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	MangaTitle string `json:"manga_title"`
	Chapter    int    `json:"chapter"`
	Timestamp  int64  `json:"timestamp"`
}

// TCPClient connects to TCP server and receives progress broadcasts
type TCPClient struct {
	addr      string
	conn      net.Conn
	hub       *SSEHub
	connected bool
}

// NewTCPClient creates a new TCP client
func NewTCPClient(addr string, hub *SSEHub) *TCPClient {
	return &TCPClient{
		addr:      addr,
		hub:       hub,
		connected: false,
	}
}

// Connect establishes connection to TCP server and starts listening
func (c *TCPClient) Connect() error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}

	c.conn = conn
	c.connected = true
	log.Printf("Connected to TCP Progress Sync Server at %s", c.addr)

	// Start listening for broadcasts
	go c.listenForBroadcasts()

	return nil
}

// listenForBroadcasts reads progress updates from TCP server and forwards to SSE hub
func (c *TCPClient) listenForBroadcasts() {
	defer func() {
		c.connected = false
		if c.conn != nil {
			c.conn.Close()
		}
		log.Println("TCP client disconnected")
	}()

	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var update TCPProgressUpdate
		err := json.Unmarshal([]byte(line), &update)
		if err != nil {
			log.Printf("Error parsing TCP progress update: %v", err)
			continue
		}

		log.Printf("📡 TCP Progress Update: User=%s, Manga=%s, Chapter=%d",
			update.Username, update.MangaTitle, update.Chapter)

		// Broadcast to SSE clients
		select {
		case c.hub.ProgressBroadcast <- update:
			// Successfully queued
		default:
			log.Println("Warning: Progress broadcast channel full")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("TCP client error: %v", err)
	}
}

// Reconnect attempts to reconnect to TCP server
func (c *TCPClient) Reconnect() {
	for {
		if c.connected {
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("Attempting to reconnect to TCP server...")
		err := c.Connect()
		if err != nil {
			log.Printf("Failed to reconnect to TCP server: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("Successfully reconnected to TCP server")
		return
	}
}

// Close closes the TCP connection
func (c *TCPClient) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.connected = false
	}
}
