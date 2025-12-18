package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// HandleWebSocketChat returns a gin handler that upgrades the connection and
// registers the client into the provided ChatHub.
func HandleWebSocketChat(hub *ChatHub, upgrader websocket.Upgrader) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user info from auth middleware
		userID := c.GetString("user_id")
		username := c.GetString("username")

		if userID == "" || username == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		// Get room ID from query parameter (manga ID)
		roomID := c.Query("room")
		if roomID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Room ID required"})
			return
		}

		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}

		// Get or create room
		room := hub.GetOrCreateRoom(roomID)

		// Create client connection
		client := &ClientConnection{
			Conn:     conn,
			UserID:   userID,
			Username: username,
			Room:     roomID,
		}

		// Register client with room
		room.Register <- client

		// Start reading messages from this client
		go room.readWebSocketMessages(client)
	}
}

// GetWebSocketStats returns a gin handler that exposes room/global stats.
func GetWebSocketStats(hub *ChatHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID := c.Query("room")

		if roomID == "" {
			totalRooms, totalClients := hub.GetGlobalStats()
			stats := gin.H{
				"total_rooms":   totalRooms,
				"total_clients": totalClients,
				"status":        "running",
			}
			c.JSON(http.StatusOK, stats)
			return
		}

		room := hub.GetRoom(roomID)
		if room == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
			return
		}

		stats := gin.H{
			"room_id":           roomID,
			"connected_clients": room.GetClientCount(),
			"connected_users":   room.GetConnectedUsers(),
			"status":            "running",
		}

		c.JSON(http.StatusOK, stats)
	}
}
