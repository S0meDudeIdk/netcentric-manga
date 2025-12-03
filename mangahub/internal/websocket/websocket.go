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
	Type      string `json:"type"` // "message", "join", "leave", "user_list"
	Room      string `json:"room,omitempty"`
	Users     []User `json:"users,omitempty"`
}

// User represents user information in the chat
type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// ClientConnection represents a connected WebSocket client
type ClientConnection struct {
	Conn     *websocket.Conn
	UserID   string
	Username string
	Room     string
}

// RoomHub manages room-based WebSocket connections
type RoomHub struct {
	// Map of room ID to room
	Rooms map[string]*ChatRoom

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// ChatRoom represents a chat room for a specific manga
type ChatRoom struct {
	// Room ID (manga ID)
	RoomID string

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

// NewRoomHub creates a new RoomHub instance
func NewRoomHub() *RoomHub {
	return &RoomHub{
		Rooms: make(map[string]*ChatRoom),
	}
}

// GetOrCreateRoom gets an existing room or creates a new one
func (h *RoomHub) GetOrCreateRoom(roomID string) *ChatRoom {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, exists := h.Rooms[roomID]; exists {
		return room
	}

	room := &ChatRoom{
		RoomID:     roomID,
		Clients:    make(map[*websocket.Conn]*ClientConnection),
		Broadcast:  make(chan ChatMessage, 256),
		Register:   make(chan *ClientConnection),
		Unregister: make(chan *websocket.Conn),
	}

	h.Rooms[roomID] = room
	go room.Run()

	log.Printf("Created new chat room: %s", roomID)
	return room
}

// GetRoom gets an existing room
func (h *RoomHub) GetRoom(roomID string) *ChatRoom {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.Rooms[roomID]
}

// Run starts the room's main event loop
func (r *ChatRoom) Run() {
	log.Printf("Chat room %s started", r.RoomID)

	for {
		select {
		case client := <-r.Register:
			r.mu.Lock()
			r.Clients[client.Conn] = client
			clientCount := len(r.Clients)
			r.mu.Unlock()

			log.Printf("[Room %s] Client registered: %s (%s). Total clients: %d", r.RoomID, client.Username, client.UserID, clientCount)

			// Broadcast join message to all clients in room
			joinMsg := ChatMessage{
				UserID:    client.UserID,
				Username:  client.Username,
				Message:   client.Username + " joined the chat",
				Timestamp: time.Now().Unix(),
				Type:      "join",
				Room:      r.RoomID,
			}
			r.Broadcast <- joinMsg

			// Send user list to the new client
			r.sendUserList()

		case conn := <-r.Unregister:
			r.mu.Lock()
			if client, ok := r.Clients[conn]; ok {
				delete(r.Clients, conn)
				conn.Close()
				clientCount := len(r.Clients)
				r.mu.Unlock()

				log.Printf("[Room %s] Client unregistered: %s (%s). Total clients: %d", r.RoomID, client.Username, client.UserID, clientCount)

				// Broadcast leave message to all clients in room
				leaveMsg := ChatMessage{
					UserID:    client.UserID,
					Username:  client.Username,
					Message:   client.Username + " left the chat",
					Timestamp: time.Now().Unix(),
					Type:      "leave",
					Room:      r.RoomID,
				}
				r.Broadcast <- leaveMsg

				// Send updated user list
				r.sendUserList()
			} else {
				r.mu.Unlock()
			}

		case message := <-r.Broadcast:
			r.mu.RLock()
			clients := make(map[*websocket.Conn]*ClientConnection)
			for conn, client := range r.Clients {
				clients[conn] = client
			}
			r.mu.RUnlock()

			// Marshal message to JSON
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("[Room %s] Error marshaling message: %v", r.RoomID, err)
				continue
			}

			// Send to all connected clients in room
			successCount := 0
			for conn := range clients {
				err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err != nil {
					log.Printf("[Room %s] Error setting write deadline: %v", r.RoomID, err)
					r.Unregister <- conn
					continue
				}

				err = conn.WriteMessage(websocket.TextMessage, data)
				if err != nil {
					log.Printf("[Room %s] Error writing message to client: %v", r.RoomID, err)
					r.Unregister <- conn
				} else {
					successCount++
				}
			}

			if message.Type == "message" {
				log.Printf("[Room %s] Broadcasted message from %s to %d clients", r.RoomID, message.Username, successCount)
			}
		}
	}
}

// sendUserList sends the current user list to all clients
func (r *ChatRoom) sendUserList() {
	r.mu.RLock()
	users := make([]User, 0, len(r.Clients))
	seen := make(map[string]bool)

	for _, client := range r.Clients {
		if !seen[client.UserID] {
			users = append(users, User{
				UserID:   client.UserID,
				Username: client.Username,
			})
			seen[client.UserID] = true
		}
	}
	r.mu.RUnlock()

	userListMsg := ChatMessage{
		Type:      "user_list",
		Users:     users,
		Timestamp: time.Now().Unix(),
		Room:      r.RoomID,
	}

	r.Broadcast <- userListMsg
}

// GetClientCount returns the current number of connected clients
func (r *ChatRoom) GetClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Clients)
}

// GetConnectedUsers returns a list of currently connected users
func (r *ChatRoom) GetConnectedUsers() []User {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]User, 0, len(r.Clients))
	seen := make(map[string]bool)

	for _, client := range r.Clients {
		if !seen[client.UserID] {
			users = append(users, User{
				UserID:   client.UserID,
				Username: client.Username,
			})
			seen[client.UserID] = true
		}
	}

	return users
}

// Close closes the room and all connections
func (r *ChatRoom) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("Closing chat room %s...", r.RoomID)

	// Close all client connections
	for conn := range r.Clients {
		conn.Close()
	}

	// Close channels
	close(r.Broadcast)
	close(r.Register)
	close(r.Unregister)

	log.Printf("Chat room %s closed", r.RoomID)
}

// GetGlobalStats returns statistics for all rooms
func (h *RoomHub) GetGlobalStats() (totalRooms int, totalClients int) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	totalRooms = len(h.Rooms)
	for _, room := range h.Rooms {
		totalClients += room.GetClientCount()
	}
	return
}
