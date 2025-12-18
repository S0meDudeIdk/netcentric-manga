package websocket

import (
  "log"
  "net/http"
  "time"

  "github.com/gin-gonic/gin"
  "github.com/gorilla/websocket"
)

// HandleWebSocketChat returns a gin handler that upgrades the connection and
// registers the client into the provided RoomHub.
func HandleWebSocketChat(hub *RoomHub, upgrader websocket.Upgrader) gin.HandlerFunc {
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
    go ReadWebSocketMessages(client, room)
  }
}

// ReadWebSocketMessages reads messages from a WebSocket client and broadcasts
// them to the room. This is intentionally package-level so main can delegate.
func ReadWebSocketMessages(client *ClientConnection, room *ChatRoom) {
  defer func() {
    room.Unregister <- client.Conn
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
    var msg ChatMessage
    err := client.Conn.ReadJSON(&msg)
    if err != nil {
      if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
        log.Printf("WebSocket error: %v", err)
      }
      break
    }

    // Reset read deadline on every message
    client.Conn.SetReadDeadline(time.Now().Add(pongWait))

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

    // Broadcast message to all clients in room
    room.Broadcast <- msg
  }
}

// GetWebSocketStats returns a gin handler that exposes room/global stats.
func GetWebSocketStats(hub *RoomHub) gin.HandlerFunc {
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
