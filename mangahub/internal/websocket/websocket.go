package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ChatMessage represents a chat message
type ChatMessage struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"` // "message", "join", "leave"
}

// ClientConnection represents a connected WebSocket client
type ClientConnection struct {
	Conn     *websocket.Conn
	UserID   string
	Username string
}

// ChatHub manages WebSocket connections and message broadcasting
type ChatHub struct {
	// Registered clients mapped by connection
	Clients map[*websocket.Conn]*ClientConnection

	// Broadcast channel for messages
	Broadcast chan ChatMessage

	// Register channel for new clients
	Register chan *ClientConnection

	// Unregister channel for disconnecting clients
	Unregister chan *websocket.Conn

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewChatHub creates a new ChatHub instance
func NewChatHub() *ChatHub {
	return &ChatHub{
		Clients:    make(map[*websocket.Conn]*ClientConnection),
		Broadcast:  make(chan ChatMessage, 256),
		Register:   make(chan *ClientConnection),
		Unregister: make(chan *websocket.Conn),
	}
}

// Run starts the hub's main event loop
func (h *ChatHub) Run() {
	log.Println("WebSocket Chat Hub started")

	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.Conn] = client
			clientCount := len(h.Clients)
			h.mu.Unlock()

			log.Printf("Client registered: %s (%s). Total clients: %d", client.Username, client.UserID, clientCount)

			// Broadcast join message to all clients
			joinMsg := ChatMessage{
				UserID:    client.UserID,
				Username:  client.Username,
				Message:   client.Username + " joined the chat",
				Timestamp: time.Now().Unix(),
				Type:      "join",
			}
			h.Broadcast <- joinMsg

		case conn := <-h.Unregister:
			h.mu.Lock()
			if client, ok := h.Clients[conn]; ok {
				delete(h.Clients, conn)
				conn.Close()
				clientCount := len(h.Clients)
				h.mu.Unlock()

				log.Printf("Client unregistered: %s (%s). Total clients: %d", client.Username, client.UserID, clientCount)

				// Broadcast leave message to all clients
				leaveMsg := ChatMessage{
					UserID:    client.UserID,
					Username:  client.Username,
					Message:   client.Username + " left the chat",
					Timestamp: time.Now().Unix(),
					Type:      "leave",
				}
				h.Broadcast <- leaveMsg
			} else {
				h.mu.Unlock()
			}

		case message := <-h.Broadcast:
			h.mu.RLock()
			clients := make(map[*websocket.Conn]*ClientConnection)
			for conn, client := range h.Clients {
				clients[conn] = client
			}
			h.mu.RUnlock()

			// Marshal message to JSON
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			// Send to all connected clients
			successCount := 0
			for conn := range clients {
				err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err != nil {
					log.Printf("Error setting write deadline: %v", err)
					h.Unregister <- conn
					continue
				}

				err = conn.WriteMessage(websocket.TextMessage, data)
				if err != nil {
					log.Printf("Error writing message to client: %v", err)
					h.Unregister <- conn
				} else {
					successCount++
				}
			}

			if message.Type == "message" {
				log.Printf("Broadcasted message from %s to %d clients", message.Username, successCount)
			}
		}
	}
}

// GetClientCount returns the current number of connected clients
func (h *ChatHub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Clients)
}

// GetConnectedUsers returns a list of currently connected usernames
func (h *ChatHub) GetConnectedUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0, len(h.Clients))
	seen := make(map[string]bool)

	for _, client := range h.Clients {
		if !seen[client.Username] {
			users = append(users, client.Username)
			seen[client.Username] = true
		}
	}

	return users
}

// Close closes the hub and all connections
func (h *ChatHub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Println("Closing WebSocket Chat Hub...")

	// Close all client connections
	for conn := range h.Clients {
		conn.Close()
	}

	// Close channels
	close(h.Broadcast)
	close(h.Register)
	close(h.Unregister)

	log.Println("WebSocket Chat Hub closed")
}
