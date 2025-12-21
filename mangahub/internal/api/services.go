package api

import (
	"encoding/json"
	"log"
	grpcClient "mangahub/internal/grpc"
	"mangahub/internal/udp"
	"net/http"
	"os"
	"strings"
	"time"
)

// connectToGRPCServer establishes connection to gRPC server
func (s *APIServer) connectToGRPCServer() {
	grpcAddr := os.Getenv("GRPC_SERVER_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:9003" // Default gRPC server address
	}

	maxRetries := 5
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		if i == 0 {
			log.Printf("Attempting to connect to gRPC server at %s...", grpcAddr)
		}

		client, err := grpcClient.NewClient(grpcAddr)
		if err != nil {
			if i < maxRetries-1 {
				log.Printf("Failed to connect to gRPC server (attempt %d/%d): %v", i+1, maxRetries, err)
				time.Sleep(retryDelay)
				continue
			} else {
				log.Printf("WARNING: gRPC server not available. Some features may be limited. Error: %v", err)
				return
			}
		}

		s.GRPCClient = client
		log.Printf("Successfully connected to gRPC server at %s", grpcAddr)
		return
	}

	// gRPC server is optional - continue without it
	log.Printf("INFO: Running without gRPC server connection.")
}
func (s *APIServer) initializeTCP() {
	// Configure TCP server URL
	tcpServerHost := os.Getenv("TCP_SERVER_ADDR")
	if tcpServerHost == "TCP_SERVER_HTTP_ADDR" {
		tcpServerHost = "http://localhost:9010" // Default TCP server HTTP trigger API
	}
	s.tcpServerURL = tcpServerHost
	log.Printf("TCP Server HTTP API configured at %s", s.tcpServerURL)
}

// triggerTCPBroadcast sends a progress update to the standalone TCP server via HTTP
func (s *APIServer) triggerTCPBroadcast(userID, userName, mangaID string, chapter int) {
	if s.tcpServerURL == "" || s.httpClient == nil {
		log.Println("TCP server not configured")
		return
	}

	// Get manga title from ID
	manga, err := s.MangaService.GetManga(mangaID)
	if err != nil {
		log.Printf("Failed to get manga for TCP broadcast: %v", err)
		return
	}

	// Create progress update
	type ProgressUpdate struct {
		UserID     string `json:"user_id"`
		Username   string `json:"username"`
		MangaTitle string `json:"manga_title"`
		Chapter    int    `json:"chapter"`
		Timestamp  int64  `json:"timestamp"`
	}

	update := ProgressUpdate{
		UserID:     userID,
		Username:   userName,
		MangaTitle: manga.Title,
		Chapter:    chapter,
		Timestamp:  time.Now().Unix(),
	}

	// Marshal to JSON
	data, err := json.Marshal(update)
	if err != nil {
		log.Printf("Failed to marshal progress update: %v", err)
		return
	}

	// Send POST request to TCP server's HTTP trigger
	resp, err := s.httpClient.Post(
		s.tcpServerURL+"/trigger",
		"application/json",
		strings.NewReader(string(data)),
	)
	if err != nil {
		log.Printf("Failed to trigger TCP broadcast: %v (Is TCP server running on %s?)", err, s.tcpServerURL)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("TCP server returned status %d", resp.StatusCode)
		return
	}

	log.Printf("Successfully triggered TCP broadcast: User=%s, Manga=%s, Chapter=%d", userID, manga.Title, chapter)
}

func (s *APIServer) initializeUDP() {
	// Initialize HTTP client for UDP server communication
	udpServerHost := os.Getenv("UDP_SERVER_HTTP_ADDR")
	if udpServerHost == "" {
		udpServerHost = "http://udp-server:9020" // Default to Docker service name
	}
	s.udpServerURL = udpServerHost
	log.Printf("UDP Server HTTP API configured at %s", s.udpServerURL)
}

// triggerUDPNotification sends a notification to the standalone UDP server via HTTP
// Overloaded to accept individual parameters for convenience
func (s *APIServer) triggerUDPNotification(userID, notifType, message string) {
	if s.udpServerURL == "" || s.httpClient == nil {
		log.Println("UDP server not configured")
		return
	}

	notification := udp.Notification{
		Type:      notifType,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	// Marshal notification to JSON
	data, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal notification: %v", err)
		return
	}

	// Send POST request to UDP server's HTTP trigger
	resp, err := s.httpClient.Post(
		s.udpServerURL+"/trigger",
		"application/json",
		strings.NewReader(string(data)),
	)
	if err != nil {
		log.Printf("Failed to trigger UDP notification: %v (Is UDP server running on %s?)", err, s.udpServerURL)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("UDP server returned status %d", resp.StatusCode)
		return
	}

	log.Printf("Successfully triggered UDP notification: %s - %s", notification.Type, notification.Message)
}

// initializeTCPUserManager creates the TCP user connection manager
func (s *APIServer) initializeTCPUserManager() {
	tcpAddr := os.Getenv("TCP_SERVER_ADDR")
	if tcpAddr == "" {
		tcpAddr = "127.0.0.1:9001" // Default TCP progress sync server (use IPv4)
	}

	s.TCPUserManager = NewTCPUserManager(tcpAddr, s.SSEHub)
	log.Printf("✅ TCP User Manager initialized - Server: %s", tcpAddr)
	log.Println("Users will connect to TCP server on login and disconnect on logout")
}

// initializeUDPClient connects to UDP server to receive notifications
func (s *APIServer) initializeUDPClient() {
	udpAddr := os.Getenv("UDP_SERVER_ADDR")
	if udpAddr == "" {
		udpAddr = "udp-server:9002" // Default to Docker service name
	}

	var err error
	s.UDPClient, err = NewUDPClient(udpAddr, s.SSEHub)
	if err != nil {
		log.Printf("Warning: Failed to create UDP client: %v", err)
		log.Println("UDP notifications will be unavailable")
		return
	}

	err = s.UDPClient.Start()
	if err != nil {
		log.Printf("Warning: Failed to start UDP client: %v", err)
		log.Println("UDP notifications will be unavailable")
		return
	}

	log.Println("✅ Connected to UDP Notification Server - Push notifications enabled")
}
