package udp

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

// Simple UDP notifier
type NotificationServer struct {
	Port    string
	Clients []net.UDPAddr
}
type Notification struct {
	Type      string `json:"type"`
	MangaID   string `json:"manga_id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

func NewNotificationServer(port string) *NotificationServer {
	return &NotificationServer{
		Port:    port,
		Clients: make([]net.UDPAddr, 0),
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
	defer conn.Close()

	log.Printf("UDP Notification Server listening on %s", s.Port)

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
		}
	}
}

func (s *NotificationServer) RegisterClient(clientAddr net.UDPAddr) {
	for _, client := range s.Clients {
		if client.String() == clientAddr.String() {
			return
		}
	}
	s.Clients = append(s.Clients, clientAddr)
}

func (s *NotificationServer) UnregisterClient(clientAddr net.UDPAddr) {
	for i, client := range s.Clients {
		if client.String() == clientAddr.String() {
			s.Clients = append(s.Clients[:i], s.Clients[i+1:]...)
			return
		}
	}
}

func (s *NotificationServer) BroadcastNotification(notification Notification) error {
	if len(s.Clients) == 0 {
		log.Println("No clients registerd for notifications")
		return nil
	}

	if notification.Timestamp == 0 {
		notification.Timestamp = time.Now().Unix()
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("error marshaling notification: %w", err)
	}

	addr, err := net.ResolveUDPAddr("udp", s.Port)
	if err != nil {
		return fmt.Errorf("error resolving UDP address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("error creating UDP connection: %w", err)
	}
	defer conn.Close()

	// Send to all registered clients
	failedClients := []int{}
	for i, client := range s.Clients {
		_, err := conn.WriteToUDP(data, &client)
		if err != nil {
			log.Printf("Error sending notification to %s: %v", client.String(), err)
			failedClients = append(failedClients, i)
		}
	}

	// Remove failed clients
	for i := len(failedClients) - 1; i >= 0; i-- {
		idx := failedClients[i]
		s.Clients = append(s.Clients[:idx], s.Clients[idx+1:]...)
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
	return len(s.Clients)
}
