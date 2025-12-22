package tcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type ClientInfo struct {
	Conn     net.Conn
	LastSeen time.Time
}

type ProgressSyncServer struct {
	Port        string
	Connections map[string]*ClientInfo
	Broadcast   chan ProgressUpdate
	mu          sync.Mutex
}

type ProgressUpdate struct {
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	MangaTitle string `json:"manga_title"`
	Chapter    int    `json:"chapter"`
	Timestamp  int64  `json:"timestamp"`
}

func NewProgressSyncServer(port string) *ProgressSyncServer {
	return &ProgressSyncServer{
		Port:        port,
		Connections: make(map[string]*ClientInfo),
		Broadcast:   make(chan ProgressUpdate, 100),
	}
}

func (s *ProgressSyncServer) Start() error {
	// Bind to 0.0.0.0 to accept connections from all network interfaces
	bindAddr := "0.0.0.0" + s.Port
	listener, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return fmt.Errorf("error starting tcp server: %w", err)
	}
	defer listener.Close()

	log.Println("TCP Server listening on", bindAddr)

	go s.handleBroadcast()
	go s.startHealthCheck()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connections:", err)
			continue
		}

		go s.handleTCPClient(conn)
	}
}

func (s *ProgressSyncServer) handleTCPClient(conn net.Conn) {
	defer conn.Close()

	addr := conn.RemoteAddr().String()
	log.Printf("New connection from %s", addr)

	s.mu.Lock()
	s.Connections[addr] = &ClientInfo{
		Conn:     conn,
		LastSeen: time.Now(),
	}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.Connections, addr)
		s.mu.Unlock()
		log.Printf("Client %s disconnected", addr)
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()

		// Handle PING keep-alive messages
		if message == "PING" {
			s.mu.Lock()
			if client, exists := s.Connections[addr]; exists {
				client.LastSeen = time.Now()
				// Respond with PONG
				client.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				client.Conn.Write([]byte("PONG\n"))
			}
			s.mu.Unlock()
			continue
		}

		var update ProgressUpdate
		if err := json.Unmarshal([]byte(message), &update); err != nil {
			log.Printf("Error parsing message from %s: %v", addr, err)
			continue
		}

		// Ignore empty/keep-alive messages (no manga title)
		if update.MangaTitle == "" {
			// Still update last seen
			s.mu.Lock()
			if client, exists := s.Connections[addr]; exists {
				client.LastSeen = time.Now()
			}
			s.mu.Unlock()
			continue
		}

		if update.Timestamp == 0 {
			update.Timestamp = time.Now().Unix()
		}

		// Update last seen
		s.mu.Lock()
		if client, exists := s.Connections[addr]; exists {
			client.LastSeen = time.Now()
		}
		s.mu.Unlock()

		log.Printf("Received progress update from %s: User=%s, Manga=%s, Chapter=%d", addr, update.UserID, update.MangaTitle, update.Chapter)

		s.Broadcast <- update
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from client %s: %v", addr, err)
	}
}

func (s *ProgressSyncServer) handleBroadcast() {
	for update := range s.Broadcast {
		message, err := json.Marshal(update)
		if err != nil {
			log.Printf("Error marshaling update: %v", err)
			continue
		}

		message = append(message, '\n')

		s.mu.Lock()
		failedClients := []string{}
		for addr, client := range s.Connections {
			client.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			_, err := client.Conn.Write(message)
			if err != nil {
				log.Printf("Error sending to client %s: %v", addr, err)
				client.Conn.Close()
				failedClients = append(failedClients, addr)
			}
		}
		// Remove failed clients
		for _, addr := range failedClients {
			delete(s.Connections, addr)
		}
		log.Printf("Broadcasted update to %d clients", len(s.Connections)-len(failedClients))
		s.mu.Unlock()
	}
}

func (s *ProgressSyncServer) Close() {
	close(s.Broadcast)

	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, client := range s.Connections {
		log.Printf("Closing connection to %s", addr)
		client.Conn.Close()
	}
}

// startHealthCheck periodically checks for dead connections
func (s *ProgressSyncServer) startHealthCheck() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		deadClients := []string{}

		for addr, client := range s.Connections {
			if now.Sub(client.LastSeen) > 120*time.Second {
				log.Printf("Client %s timed out, removing...", addr)
				client.Conn.Close()
				deadClients = append(deadClients, addr)
			}
		}

		for _, addr := range deadClients {
			delete(s.Connections, addr)
		}

		if len(deadClients) > 0 {
			log.Printf("Removed %d dead client(s). Active: %d", len(deadClients), len(s.Connections))
		}
		s.mu.Unlock()
	}
}

// GetClientCount returns the current number of connected clients
func (s *ProgressSyncServer) GetClientCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.Connections)
}

// TriggerBroadcast sends a progress update to all connected clients (called via HTTP)
func (s *ProgressSyncServer) TriggerBroadcast(update ProgressUpdate) error {
	if update.Timestamp == 0 {
		update.Timestamp = time.Now().Unix()
	}

	select {
	case s.Broadcast <- update:
		return nil
	default:
		return fmt.Errorf("broadcast channel full")
	}
}
