package api

import (
	"encoding/json"
	"log"
	"net"
	"time"
)

// UDPNotification represents a notification from UDP server
type UDPNotification struct {
	Type      string `json:"type"`
	MangaID   string `json:"manga_id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// UDPClient connects to UDP server and receives notifications
type UDPClient struct {
	serverAddr *net.UDPAddr
	localAddr  *net.UDPAddr
	conn       *net.UDPConn
	hub        *SSEHub
	running    bool
}

// NewUDPClient creates a new UDP client
func NewUDPClient(serverAddr string, hub *SSEHub) (*UDPClient, error) {
	// Resolve server address
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, err
	}

	// Create local address for listening
	localAddr, err := net.ResolveUDPAddr("udp", ":0") // Use any available port
	if err != nil {
		return nil, err
	}

	return &UDPClient{
		serverAddr: udpAddr,
		localAddr:  localAddr,
		hub:        hub,
		running:    false,
	}, nil
}

// Start starts listening for UDP notifications
func (c *UDPClient) Start() error {
	// Create UDP connection
	conn, err := net.ListenUDP("udp", c.localAddr)
	if err != nil {
		return err
	}

	c.conn = conn
	c.running = true
	log.Printf("UDP client listening on %s", conn.LocalAddr().String())

	// Register with UDP server
	go c.registerWithServer()

	// Start listening for notifications
	go c.listenForNotifications()

	// Send heartbeat to keep registration alive
	go c.sendHeartbeat()

	return nil
}

// registerWithServer sends registration message to UDP server
func (c *UDPClient) registerWithServer() {
	// UDP server expects simple "REGISTER" string, not JSON
	_, err := c.conn.WriteToUDP([]byte("REGISTER"), c.serverAddr)
	if err != nil {
		log.Printf("Error registering with UDP server: %v", err)
		return
	}

	log.Printf("✅ Registered with UDP Notification Server at %s", c.serverAddr.String())
}

// sendHeartbeat sends periodic heartbeat to maintain registration
func (c *UDPClient) sendHeartbeat() {
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for c.running {
		<-ticker.C

		// UDP server expects simple "PONG" string, not JSON
		_, err := c.conn.WriteToUDP([]byte("PONG"), c.serverAddr)
		if err != nil {
			log.Printf("Error sending heartbeat to UDP server: %v", err)
		}
	}
}

// listenForNotifications receives and processes UDP notifications
func (c *UDPClient) listenForNotifications() {
	defer func() {
		c.running = false
		if c.conn != nil {
			c.conn.Close()
		}
		log.Println("UDP client stopped")
	}()

	buffer := make([]byte, 4096)

	for c.running {
		n, _, err := c.conn.ReadFromUDP(buffer)
		if err != nil {
			if c.running {
				log.Printf("UDP read error: %v", err)
			}
			continue
		}

		message := string(buffer[:n])

		// Ignore heartbeat/control messages
		if message == "PING" || message == "PONG" || message == "REGISTERED" {
			continue
		}

		// Parse notification JSON
		var notification UDPNotification
		err = json.Unmarshal(buffer[:n], &notification)
		if err != nil {
			log.Printf("Error parsing UDP notification: %v (message: %s)", err, message)
			continue
		}

		log.Printf("🔔 UDP Notification: Type=%s, Message=%s",
			notification.Type, notification.Message)

		// Broadcast to SSE clients
		select {
		case c.hub.NotificationBroadcast <- notification:
			log.Printf("✅ Notification forwarded to %d SSE client(s)", len(c.hub.notificationClients))
		default:
			log.Println("⚠️  Warning: Notification broadcast channel full")
		}
	}
}

// Stop stops the UDP client
func (c *UDPClient) Stop() {
	c.running = false
	if c.conn != nil {
		c.conn.Close()
	}
}
