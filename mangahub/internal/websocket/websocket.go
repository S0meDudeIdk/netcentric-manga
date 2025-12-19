package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Re-export the gorilla websocket Upgrader type so callers can avoid importing
// the gorilla package directly when they already import this internal package.
type Upgrader = websocket.Upgrader

const (
	// Maximum number of messages to keep in history per room
	maxHistorySize = 1000
)

// Message represents a WebSocket message (chat, notification, progress update, etc.)
type Message struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"` // "message", "join", "leave", "user_list", "notification", "progress_update"
	Room      string `json:"room,omitempty"`
	Users     []User `json:"users,omitempty"`
	MangaID   string `json:"manga_id,omitempty"` // For notifications and progress updates
	Chapter   int    `json:"chapter,omitempty"`  // For progress updates
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

// ChatHub manages room-based WebSocket connections
type ChatHub struct {
	// Map of room ID to room
	Rooms map[string]*ChatRoom

	// Global notification channel for broadcasting to all clients
	NotificationChannel chan Message

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
	Broadcast chan Message

	// Register channel for new clients
	Register chan *ClientConnection

	// Unregister channel for disconnecting clients
	Unregister chan *websocket.Conn

	// Message history (up to 1000 messages)
	MessageHistory []Message

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewChatHub creates a new ChatHub instance
func NewChatHub() *ChatHub {
	hub := &ChatHub{
		Rooms:               make(map[string]*ChatRoom),
		NotificationChannel: make(chan Message, 100),
	}
	// Start global notification broadcaster
	go hub.broadcastNotifications()
	return hub
}

// GetOrCreateRoom gets an existing room or creates a new one
func (h *ChatHub) GetOrCreateRoom(roomID string) *ChatRoom {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, exists := h.Rooms[roomID]; exists {
		return room
	}

	room := &ChatRoom{
		RoomID:         roomID,
		Clients:        make(map[*websocket.Conn]*ClientConnection),
		Broadcast:      make(chan Message, 256),
		Register:       make(chan *ClientConnection),
		Unregister:     make(chan *websocket.Conn),
		MessageHistory: make([]Message, 0, maxHistorySize),
	}

	h.Rooms[roomID] = room
	go room.Run()

	log.Printf("Created new chat room: %s", roomID)
	return room
}

// GetRoom gets an existing room
func (h *ChatHub) GetRoom(roomID string) *ChatRoom {
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
			joinMsg := Message{
				UserID:    client.UserID,
				Username:  client.Username,
				Message:   client.Username + " joined the chat",
				Timestamp: time.Now().Unix(),
				Type:      "join",
				Room:      r.RoomID,
			}
			select {
			case r.Broadcast <- joinMsg:
				// Queued successfully
			default:
				log.Printf("[Room %s] Warning: broadcast channel full, dropping join message for %s", r.RoomID, client.Username)
			}

			// Send message history to new client
			r.sendMessageHistory(client)

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
				leaveMsg := Message{
					UserID:    client.UserID,
					Username:  client.Username,
					Message:   client.Username + " left the chat",
					Timestamp: time.Now().Unix(),
					Type:      "leave",
					Room:      r.RoomID,
				}
				select {
				case r.Broadcast <- leaveMsg:
					// Queued successfully
				default:
					log.Printf("[Room %s] Warning: broadcast channel full, dropping leave message for %s", r.RoomID, client.Username)
				}

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

			// Log broadcast results based on message type
			switch message.Type {
			case "message":
				log.Printf("[Room %s] Broadcasted message from %s to %d clients", r.RoomID, message.Username, successCount)
				// Add message to history
				r.addToHistory(message)
			case "progress_update":
				log.Printf("[Room %s] Broadcasted progress update from %s (chapter %d) to %d clients", r.RoomID, message.Username, message.Chapter, successCount)
			case "notification":
				log.Printf("[Room %s] Broadcasted notification to %d clients: %s", r.RoomID, successCount, message.Message)
			case "join", "leave":
				log.Printf("[Room %s] Broadcasted %s event for %s to %d clients", r.RoomID, message.Type, message.Username, successCount)
			case "user_list":
				log.Printf("[Room %s] Broadcasted user list (%d users) to %d clients", r.RoomID, len(message.Users), successCount)
			}
		}
	}
}

// ReadWebSocketMessages reads messages from a WebSocket client and broadcasts
// them to the room. This is intentionally package-level so main can delegate.
func (r *ChatRoom) readWebSocketMessages(client *ClientConnection) {
	defer func() {
		r.Unregister <- client.Conn
	}()

	// Ping interval and timeouts
	const (
		pongWait   = 120 * time.Second
		pingPeriod = (pongWait * 9) / 10
		writeWait  = 10 * time.Second
	)

	// Configure read settings with extended deadline
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Start ping ticker to keep connection alive
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	// Start goroutine to send periodic pings
	go func() {
		for range pingTicker.C {
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	// Read messages loop
	for {
		var msg Message
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Reset read deadline on every message
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))

		if msg.Type == "user_list" {
			r.sendUserList()
			continue
		}

		// Set message metadata
		msg.UserID = client.UserID
		msg.Username = client.Username
		msg.Timestamp = time.Now().Unix()
		msg.Type = "message"
		msg.Room = client.Room

		// Validate message
		if msg.Message == "" || len(msg.Message) > 1000 {
			log.Printf("Invalid message from %s: empty or too long", client.Username)
			continue
		}

		// Broadcast message to all clients in room (non-blocking)
		select {
		case r.Broadcast <- msg:
			// Message queued successfully
		default:
			log.Printf("[Room %s] Warning: broadcast channel full, dropping message from %s", r.RoomID, client.Username)
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

	userListMsg := Message{
		Type:      "user_list",
		Users:     users,
		Timestamp: time.Now().Unix(),
		Room:      r.RoomID,
	}

	select {
	case r.Broadcast <- userListMsg:
		// Queued successfully
	default:
		log.Printf("[Room %s] Warning: broadcast channel full, dropping user list update", r.RoomID)
	}
}

// sendMessageHistory sends the message history to a newly connected client
func (r *ChatRoom) sendMessageHistory(client *ClientConnection) {
	r.mu.RLock()
	history := make([]Message, len(r.MessageHistory))
	copy(history, r.MessageHistory)
	r.mu.RUnlock()

	// Send each historical message to the client
	for _, msg := range history {
		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("[Room %s] Error marshaling history message: %v", r.RoomID, err)
			continue
		}

		client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		err = client.Conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("[Room %s] Error sending history to client: %v", r.RoomID, err)
			break
		}
	}

	if len(history) > 0 {
		log.Printf("[Room %s] Sent %d messages from history to %s", r.RoomID, len(history), client.Username)
	}
}

// addToHistory adds a message to the room's history, maintaining max size
func (r *ChatRoom) addToHistory(message Message) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.MessageHistory = append(r.MessageHistory, message)

	// Keep only the last maxHistorySize messages
	if len(r.MessageHistory) > maxHistorySize {
		r.MessageHistory = r.MessageHistory[len(r.MessageHistory)-maxHistorySize:]
	}
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
func (h *ChatHub) GetGlobalStats() (totalRooms int, totalClients int) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	totalRooms = len(h.Rooms)
	for _, room := range h.Rooms {
		totalClients += room.GetClientCount()
	}
	return
}

// broadcastNotifications listens for global notifications and broadcasts only to the global-notifications room
func (h *ChatHub) broadcastNotifications() {
	for notification := range h.NotificationChannel {
		h.mu.RLock()
		// Only broadcast to the special "global-notifications" room
		if room, exists := h.Rooms["global-notifications"]; exists {
			select {
			case room.Broadcast <- notification:
				// Broadcast success - actual result logged in Run()
			default:
				log.Printf("Warning: notification channel full for global-notifications room")
			}
		} else {
			log.Printf("Warning: global-notifications room not found, notification not sent")
		}
		h.mu.RUnlock()
	}
}

// BroadcastNotification sends a notification to all connected WebSocket clients
func (h *ChatHub) BroadcastNotification(mangaID, notificationType, message string) {
	notification := Message{
		Type:      "notification",
		MangaID:   mangaID,
		Message:   message,
		Timestamp: time.Now().Unix(),
		UserID:    "system",
		Username:  "System",
	}

	select {
	case h.NotificationChannel <- notification:
		// Queued successfully - actual broadcast logged in Run()
	default:
		log.Println("Warning: notification channel full, dropping notification")
	}
}

// BroadcastProgressUpdate sends a progress update to all connected WebSocket clients in the manga's chat room
func (h *ChatHub) BroadcastProgressUpdate(userID, username, mangaID string, chapter int) {
	// Get the specific manga's chat room if it exists
	h.mu.RLock()
	room := h.Rooms[mangaID]
	h.mu.RUnlock()

	if room == nil {
		log.Printf("No chat room found for manga %s, progress update not broadcasted to WebSocket", mangaID)
		return
	}

	// Create progress update message
	progressMsg := Message{
		Type:      "progress_update",
		UserID:    userID,
		Username:  username,
		MangaID:   mangaID,
		Chapter:   chapter,
		Timestamp: time.Now().Unix(),
		Room:      mangaID,
	}

	// Send through the room's broadcast channel (proper pattern)
	select {
	case room.Broadcast <- progressMsg:
		// Queued successfully - actual broadcast logged in Run()
	default:
		log.Printf("Warning: broadcast channel full for room %s, dropping progress update", mangaID)
	}
}
