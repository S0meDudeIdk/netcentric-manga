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

type ProgressSyncServer struct {
	Port        string
	Connections map[string]net.Conn
	Broadcast   chan ProgressUpdate
	mu          sync.Mutex
}
type ProgressUpdate struct {
	UserID    string `json:"user_id"`
	MangaID   string `json:"manga_id"`
	Chapter   int    `json:"chapter"`
	Timestamp int64  `json:"timestamp"`
}

func NewProgressSyncServer(port string) *ProgressSyncServer {
	return &ProgressSyncServer{
		Port:        port,
		Connections: make(map[string]net.Conn),
		Broadcast:   make(chan ProgressUpdate),
	}
}

func (s *ProgressSyncServer) Start() error {
	listener, err := net.Listen("tcp", s.Port)
	if err != nil {
		return fmt.Errorf("error starting tcp server: %w", err)
	}
	defer listener.Close()

	log.Println("TCP Server listening on", s.Port)

	go s.handleBroadcast()

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
	s.Connections[addr] = conn
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

		var update ProgressUpdate
		if err := json.Unmarshal([]byte(message), &update); err != nil {
			log.Printf("Error parsing message from %s: %v", addr, err)
			continue
		}

		if update.Timestamp == 0 {
			update.Timestamp = time.Now().Unix()
		}

		log.Printf("Received progress update from %s: User=%s, Manga=%s, Chapter=%d", addr, update.UserID, update.MangaID, update.Chapter)

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
		}

		message = append(message, '\n')

		s.mu.Lock()
		defer s.mu.Unlock()

		for addr, conn := range s.Connections {
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			_, err := conn.Write(message)
			if err != nil {
				log.Printf("Error sending to client %s: %v", addr, err)
				conn.Close()
				delete(s.Connections, addr)
			}
		}

		log.Printf("Broadcasted update to %d clients", len(s.Connections))
	}
}

func (s *ProgressSyncServer) Close() {
	close(s.Broadcast)

	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, conn := range s.Connections {
		log.Printf("Closing connection to %s", addr)
		conn.Close()
	}
}
