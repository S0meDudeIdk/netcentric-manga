package protocol

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type ChatMessage struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"` // "message", "join", "leave", "user_list"
	Room      string `json:"room,omitempty"`
	Users     []User `json:"users,omitempty"`
}

type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// joinChatHub connects to a manga-specific chat room via WebSocket
func (c *Client) JoinChatHub(mangaID, mangaTitle string) {
	// Disconnect from previous room if connected
	if c.wsConn != nil {
		c.wsConn.Close()
		c.wsConn = nil
		c.wsEnabled = false
	}

	fmt.Printf("\n%s💬 Joining Chat Hub: %s%s\n", colorCyan, mangaTitle, colorReset)

	// Build WebSocket URL with room and token
	wsURL := fmt.Sprintf("ws://localhost:8080/api/v1/ws/chat?token=%s&room=%s", c.Token, mangaID)

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Printf("%s❌ Failed to connect to chat: %s%s\n", colorRed, err.Error(), colorReset)
		return
	}

	c.wsConn = conn
	c.wsEnabled = true
	c.currentRoom = mangaID

	// Set up pong handler to respond to server pings
	conn.SetPongHandler(func(appData string) error {
		return nil
	})

	fmt.Printf("%s✅ Connected to chat hub!%s\n", colorGreen, colorReset)
	fmt.Println(colorYellow + "\nChat Commands:" + colorReset)
	fmt.Println("  - Type a message and press Enter to send")
	fmt.Println("  - Type '/exit' to leave the chat")
	fmt.Println("  - Type '/users' to see online users")
	fmt.Println()

	// Start listening for messages in background
	go c.ListenWebSocketMessages()

	// Start chat loop
	c.ChatLoop()
}

// listenWebSocketMessages listens for incoming WebSocket messages
func (c *Client) ListenWebSocketMessages() {
	if c.wsConn == nil {
		return
	}

	// Set ping handler to automatically respond with pong
	c.wsConn.SetPingHandler(func(appData string) error {
		err := c.wsConn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(10*time.Second))
		if err != nil {
			return err
		}
		return nil
	})

	for {
		var msg ChatMessage
		err := c.wsConn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("\n%s❌ Connection lost: %s%s\n", colorRed, err.Error(), colorReset)
			}
			c.wsEnabled = false
			return
		}

		c.DisplayChatMessage(msg)
	}
}

// displayChatMessage formats and displays a chat message
func (c *Client) DisplayChatMessage(msg ChatMessage) {
	msgType := msg.Type
	username := msg.Username
	userID := msg.UserID
	message := msg.Message
	timestamp := msg.Timestamp

	// Format timestamp
	t := time.Unix(int64(timestamp), 0)
	timeStr := t.Format("15:04:05")

	switch msgType {
	case "message":
		// Different color for own messages
		if userID == c.UserID {
			fmt.Printf("[%s] %s%s%s: %s\n", timeStr, colorGreen, username, colorReset, message)
		} else {
			fmt.Printf("[%s] %s%s%s: %s\n", timeStr, colorCyan, username, colorReset, message)
		}

	case "join":
		fmt.Printf("%s→ %s joined the chat%s\n", colorYellow, username, colorReset)

	case "leave":
		fmt.Printf("%s← %s left the chat%s\n", colorYellow, username, colorReset)

	case "user_list":
		fmt.Printf("\n%s👥 Online Users (%d):%s\n", colorCyan, len(msg.Users), colorReset)
		for _, user := range msg.Users {
			if user.UserID == c.UserID {
				fmt.Printf("  %s• %s (you)%s\n", colorGreen, user.Username, colorReset)
			} else {
				fmt.Printf("  • %s\n", user.Username)
			}
		}
	}
}

// chatLoop handles user input in the chat
func (c *Client) ChatLoop() {
	chatScanner := bufio.NewScanner(os.Stdin)

	for c.wsEnabled {
		fmt.Print(colorGreen + "> " + colorReset)

		if !chatScanner.Scan() {
			break
		}

		input := strings.TrimSpace(chatScanner.Text())

		if input == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(input, "/") {
			c.HandleChatCommand(input)
			continue
		}

		// Send message
		c.SendChatMessage(input)
	}
}

// handleChatCommand processes chat commands
func (c *Client) HandleChatCommand(command string) {
	switch command {
	case "/exit":
		fmt.Printf("%s← Leaving chat hub...%s\n", colorYellow, colorReset)
		if c.wsConn != nil {
			c.wsConn.Close()
			c.wsConn = nil
		}
		c.wsEnabled = false
		c.currentRoom = ""

	case "/users":
		// Request user list
		msg := ChatMessage{
			Type: "get_users",
		}
		if err := c.wsConn.WriteJSON(msg); err != nil {
			fmt.Printf("%s❌ Failed to request user list%s\n", colorRed, colorReset)
		}

	default:
		fmt.Printf("%s❌ Unknown command. Available: /exit, /users%s\n", colorRed, colorReset)
	}
}

// sendChatMessage sends a message to the chat room
func (c *Client) SendChatMessage(message string) {
	if !c.wsEnabled || c.wsConn == nil {
		fmt.Printf("%s❌ Not connected to chat%s\n", colorRed, colorReset)
		return
	}

	// Validate message length
	if len(message) > 1000 {
		fmt.Printf("%s❌ Message too long (max 1000 characters)%s\n", colorRed, colorReset)
		return
	}

	msg := ChatMessage{
		Type:    "message",
		Message: message,
		Room:    c.currentRoom,
	}

	err := c.wsConn.WriteJSON(msg)
	if err != nil {
		fmt.Printf("%s❌ Failed to send message: %s%s\n", colorRed, err.Error(), colorReset)
		c.wsEnabled = false
	}
}
