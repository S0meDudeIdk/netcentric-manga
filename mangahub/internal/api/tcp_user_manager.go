package api

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"
)

// TCPUserConnection represents a single user's TCP connection
type TCPUserConnection struct {
	UserID    string
	Conn      net.Conn
	LastSeen  time.Time
	Connected bool
	mu        sync.Mutex
}

// TCPUserManager manages per-user TCP connections to the TCP server
type TCPUserManager struct {
	tcpServerAddr string
	connections   map[string]*TCPUserConnection
	mu            sync.RWMutex
	hub           *SSEHub
}

// NewTCPUserManager creates a new TCP user connection manager
func NewTCPUserManager(tcpServerAddr string, hub *SSEHub) *TCPUserManager {
	return &TCPUserManager{
		tcpServerAddr: tcpServerAddr,
		connections:   make(map[string]*TCPUserConnection),
		hub:           hub,
	}
}

// ConnectUser establishes a TCP connection for a specific user
func (m *TCPUserManager) ConnectUser(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if user already has a connection
	if existing, exists := m.connections[userID]; exists {
		if existing.Connected {
			log.Printf("User %s already has an active TCP connection", userID)
			return nil
		}
		// Clean up old connection
		existing.Conn.Close()
		delete(m.connections, userID)
	}

	// Establish new TCP connection
	conn, err := net.Dial("tcp", m.tcpServerAddr)
	if err != nil {
		log.Printf("Failed to connect user %s to TCP server: %v", userID, err)
		return err
	}

	userConn := &TCPUserConnection{
		UserID:    userID,
		Conn:      conn,
		LastSeen:  time.Now(),
		Connected: true,
	}

	m.connections[userID] = userConn
	log.Printf("✅ User %s connected to TCP server at %s", userID, m.tcpServerAddr)

	// Start listening for broadcasts for this user
	go m.listenForUserBroadcasts(userConn)

	// Start heartbeat
	go m.sendHeartbeat(userConn)

	return nil
}

// DisconnectUser closes the TCP connection for a specific user
func (m *TCPUserManager) DisconnectUser(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userConn, exists := m.connections[userID]
	if !exists {
		log.Printf("User %s has no TCP connection to disconnect", userID)
		return
	}

	// Send DISCONNECT command to TCP server
	userConn.mu.Lock()
	if userConn.Connected && userConn.Conn != nil {
		userConn.Conn.Write([]byte("DISCONNECT\n"))
		time.Sleep(100 * time.Millisecond) // Give time for message to send
		userConn.Conn.Close()
		userConn.Connected = false
	}
	userConn.mu.Unlock()

	delete(m.connections, userID)
	log.Printf("🔌 User %s disconnected from TCP server", userID)
}

// listenForUserBroadcasts reads progress updates from TCP server for a specific user
func (m *TCPUserManager) listenForUserBroadcasts(userConn *TCPUserConnection) {
	defer func() {
		userConn.mu.Lock()
		userConn.Connected = false
		if userConn.Conn != nil {
			userConn.Conn.Close()
		}
		userConn.mu.Unlock()

		// Remove from connections map
		m.mu.Lock()
		delete(m.connections, userConn.UserID)
		m.mu.Unlock()

		log.Printf("TCP listener stopped for user %s", userConn.UserID)
	}()

	scanner := bufio.NewScanner(userConn.Conn)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line == "PONG" {
			continue
		}

		var update TCPProgressUpdate
		err := json.Unmarshal([]byte(line), &update)
		if err != nil {
			log.Printf("Error parsing TCP progress update for user %s: %v", userConn.UserID, err)
			continue
		}

		log.Printf("📡 TCP Progress Update for user %s: Manga=%s, Chapter=%d",
			userConn.UserID, update.MangaTitle, update.Chapter)

		// Broadcast to SSE clients
		select {
		case m.hub.ProgressBroadcast <- update:
			// Successfully queued
		default:
			log.Println("Warning: Progress broadcast channel full")
		}

		// Update last seen
		userConn.mu.Lock()
		userConn.LastSeen = time.Now()
		userConn.mu.Unlock()
	}

	if err := scanner.Err(); err != nil {
		log.Printf("TCP client error for user %s: %v", userConn.UserID, err)
	}
}

// sendHeartbeat sends periodic PING messages to keep connection alive
func (m *TCPUserManager) sendHeartbeat(userConn *TCPUserConnection) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		userConn.mu.Lock()
		if !userConn.Connected {
			userConn.mu.Unlock()
			return
		}

		_, err := userConn.Conn.Write([]byte("PING\n"))
		if err != nil {
			log.Printf("Failed to send heartbeat for user %s: %v", userConn.UserID, err)
			userConn.Connected = false
			userConn.mu.Unlock()
			return
		}
		userConn.mu.Unlock()
	}
}

// GetConnectedUsers returns the list of currently connected user IDs
func (m *TCPUserManager) GetConnectedUsers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]string, 0, len(m.connections))
	for userID := range m.connections {
		users = append(users, userID)
	}
	return users
}

// DisconnectAll disconnects all users (used during shutdown)
func (m *TCPUserManager) DisconnectAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for userID, userConn := range m.connections {
		userConn.mu.Lock()
		if userConn.Connected && userConn.Conn != nil {
			userConn.Conn.Write([]byte("DISCONNECT\n"))
			time.Sleep(50 * time.Millisecond)
			userConn.Conn.Close()
			userConn.Connected = false
		}
		userConn.mu.Unlock()
		log.Printf("Disconnected user %s during shutdown", userID)
	}

	m.connections = make(map[string]*TCPUserConnection)
	log.Println("All TCP user connections closed")
}
