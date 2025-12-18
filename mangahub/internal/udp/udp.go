package udp

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// ClientInfo tracks client metadata
type ClientInfo struct {
	Addr     net.UDPAddr
	LastSeen time.Time
}

// Simple UDP notifier
type NotificationServer struct {
	Port    string
	Clients map[string]*ClientInfo // key is addr.String()
	conn    *net.UDPConn
	mu      sync.Mutex
}

type Notification struct {
	Type      string `json:"type"`
	MangaID   string `json:"manga_id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

const (
	heartbeatInterval = 30 * time.Second // Send PING every 30s
	clientTimeout     = 90 * time.Second // Remove client if no response for 90s
)

func NewNotificationServer(port string) *NotificationServer {
	return &NotificationServer{
		Port:    port,
		Clients: make(map[string]*ClientInfo),
	}
}

func (s *NotificationServer) Start() error {
	addr, err := net.ResolveUDPAddr("udp", s.Port)
	if err != nil {
		return fmt.Errorf("error resolving UDP address:%w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("error starting UDP server:%w", err)
	}
	s.conn = conn
	defer conn.Close()

	log.Printf("UDP Notification Server listening on %s", s.Port)

	// Start heartbeat checker in background
	go s.startHeartbeat()

	buffer := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		message := string(buffer[:n])
		switch message {
		case "REGISTER":
			s.RegisterClient(*clientAddr)
			log.Printf("Client register: %s", clientAddr.String())

			response := []byte("REGISTERED")
			_, err = conn.WriteToUDP(response, clientAddr)
			if err != nil {
				log.Printf("Error sending acknowlegment")
			}
		case "UNREGISTER":
			s.UnregisterClient(*clientAddr)
			log.Printf("Client unregistered: %s", clientAddr.String())
		case "PONG":
			// Client is alive, update lastSeen
			s.updateClientLastSeen(*clientAddr)
		}
	}
}

func (s *NotificationServer) RegisterClient(clientAddr net.UDPAddr) {
	s.mu.Lock()
	defer s.mu.Unlock()

	addr := clientAddr.String()
	if _, exists := s.Clients[addr]; !exists {
		s.Clients[addr] = &ClientInfo{
			Addr:     clientAddr,
			LastSeen: time.Now(),
		}
	} else {
		// Update lastSeen if already registered
		s.Clients[addr].LastSeen = time.Now()
	}
}

func (s *NotificationServer) UnregisterClient(clientAddr net.UDPAddr) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.Clients, clientAddr.String())
}

// updateClientLastSeen updates the last seen timestamp for a client
func (s *NotificationServer) updateClientLastSeen(clientAddr net.UDPAddr) {
	s.mu.Lock()
	defer s.mu.Unlock()

	addr := clientAddr.String()
	if client, exists := s.Clients[addr]; exists {
		client.LastSeen = time.Now()
	}
}

// startHeartbeat periodically pings clients and removes dead ones
func (s *NotificationServer) startHeartbeat() {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		if s.conn == nil {
			s.mu.Unlock()
			return
		}

		now := time.Now()
		deadClients := []string{}

		// Check for dead clients and send PING to alive ones
		for addr, client := range s.Clients {
			if now.Sub(client.LastSeen) > clientTimeout {
				log.Printf("Client %s timed out (no response for %v), removing...", addr, clientTimeout)
				deadClients = append(deadClients, addr)
			} else {
				// Send PING
				_, err := s.conn.WriteToUDP([]byte("PING"), &client.Addr)
				if err != nil {
					log.Printf("Failed to ping client %s: %v", addr, err)
					deadClients = append(deadClients, addr)
				}
			}
		}

		// Remove dead clients
		for _, addr := range deadClients {
			delete(s.Clients, addr)
		}

		if len(deadClients) > 0 {
			log.Printf("Removed %d dead client(s). Active clients: %d", len(deadClients), len(s.Clients))
		}

		s.mu.Unlock()
	}
}

func (s *NotificationServer) BroadcastNotification(notification Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Clients) == 0 {
		log.Println("No clients registered for notifications")
		return nil
	}

	if s.conn == nil {
		return fmt.Errorf("UDP server connection not initialized")
	}

	if notification.Timestamp == 0 {
		notification.Timestamp = time.Now().Unix()
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("error marshaling notification: %w", err)
	}

	// Send to all registered clients
	failedClients := []string{}
	for addr, client := range s.Clients {
		_, err := s.conn.WriteToUDP(data, &client.Addr)
		if err != nil {
			log.Printf("Error sending notification to %s: %v", addr, err)
			failedClients = append(failedClients, addr)
		}
	}

	// Remove failed clients
	for _, addr := range failedClients {
		delete(s.Clients, addr)
	}

	log.Printf("Notification broadcast to %d clients", len(s.Clients))
	return nil
}

func (s *NotificationServer) SendChapterReleaseNotification(mangaID, mangaTitle string, chapter int) error {
	notification := Notification{
		Type:      "chapter_release",
		MangaID:   mangaID,
		Message:   fmt.Sprintf("New chapter %d release for %s", chapter, mangaTitle),
		Timestamp: time.Now().Unix(),
	}
	return s.BroadcastNotification(notification)
}

func (s *NotificationServer) SendMangaUpdateNotification(mangaID, message string) error {
	notification := Notification{
		Type:      "manga_update",
		MangaID:   mangaID,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}
	return s.BroadcastNotification(notification)
}

func (s *NotificationServer) GetClientCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.Clients)
}

// Close gracefully shuts down the UDP server
func (s *NotificationServer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		log.Println("Closing UDP server...")
		return s.conn.Close()
	}
	return nil
}
