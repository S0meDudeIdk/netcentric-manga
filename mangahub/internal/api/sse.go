package api

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SSEClient represents a connected SSE client
type SSEClient struct {
	ID      string
	Channel chan interface{}
	Context *gin.Context
}

// SSEHub manages SSE connections and broadcasts
type SSEHub struct {
	// Progress sync clients (TCP bridge)
	progressClients map[string]*SSEClient
	progressMutex   sync.RWMutex

	// Notification clients (UDP bridge)
	notificationClients map[string]*SSEClient
	notificationMutex   sync.RWMutex

	// Channels for broadcasting
	ProgressBroadcast     chan interface{}
	NotificationBroadcast chan interface{}
}

// NewSSEHub creates a new SSE hub
func NewSSEHub() *SSEHub {
	hub := &SSEHub{
		progressClients:       make(map[string]*SSEClient),
		notificationClients:   make(map[string]*SSEClient),
		ProgressBroadcast:     make(chan interface{}, 100),
		NotificationBroadcast: make(chan interface{}, 100),
	}

	// Start broadcast goroutines
	go hub.runProgressBroadcaster()
	go hub.runNotificationBroadcaster()

	return hub
}

// AddProgressClient adds a new progress sync client
func (h *SSEHub) AddProgressClient(client *SSEClient) {
	h.progressMutex.Lock()
	defer h.progressMutex.Unlock()
	h.progressClients[client.ID] = client
	log.Printf("SSE Progress client connected: %s (Total: %d)", client.ID, len(h.progressClients))
}

// RemoveProgressClient removes a progress sync client
func (h *SSEHub) RemoveProgressClient(clientID string) {
	h.progressMutex.Lock()
	defer h.progressMutex.Unlock()
	if client, exists := h.progressClients[clientID]; exists {
		close(client.Channel)
		delete(h.progressClients, clientID)
		log.Printf("SSE Progress client disconnected: %s (Remaining: %d)", clientID, len(h.progressClients))
	}
}

// AddNotificationClient adds a new notification client
func (h *SSEHub) AddNotificationClient(client *SSEClient) {
	h.notificationMutex.Lock()
	defer h.notificationMutex.Unlock()
	h.notificationClients[client.ID] = client
	log.Printf("SSE Notification client connected: %s (Total: %d)", client.ID, len(h.notificationClients))
}

// RemoveNotificationClient removes a notification client
func (h *SSEHub) RemoveNotificationClient(clientID string) {
	h.notificationMutex.Lock()
	defer h.notificationMutex.Unlock()
	if client, exists := h.notificationClients[clientID]; exists {
		close(client.Channel)
		delete(h.notificationClients, clientID)
		log.Printf("SSE Notification client disconnected: %s (Remaining: %d)", clientID, len(h.notificationClients))
	}
}

// runProgressBroadcaster handles broadcasting progress updates to all connected clients
func (h *SSEHub) runProgressBroadcaster() {
	for update := range h.ProgressBroadcast {
		h.progressMutex.RLock()
		for clientID, client := range h.progressClients {
			select {
			case client.Channel <- update:
				// Successfully sent
			case <-time.After(1 * time.Second):
				log.Printf("Timeout sending to progress client %s", clientID)
			}
		}
		h.progressMutex.RUnlock()
	}
}

// runNotificationBroadcaster handles broadcasting notifications to all connected clients
func (h *SSEHub) runNotificationBroadcaster() {
	for notification := range h.NotificationBroadcast {
		h.notificationMutex.RLock()
		for clientID, client := range h.notificationClients {
			select {
			case client.Channel <- notification:
				// Successfully sent
			case <-time.After(1 * time.Second):
				log.Printf("Timeout sending to notification client %s", clientID)
			}
		}
		h.notificationMutex.RUnlock()
	}
}

// StreamSSE handles SSE streaming for a client
func StreamSSE(c *gin.Context, client *SSEClient, removeFunc func(string)) {
	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Get client context for cancellation
	ctx := c.Request.Context()

	// Send initial connection message
	c.SSEvent("connected", map[string]string{
		"client_id": client.ID,
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	})
	c.Writer.Flush()

	// Keep-alive ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			removeFunc(client.ID)
			return

		case data, ok := <-client.Channel:
			if !ok {
				// Channel closed
				return
			}

			// Marshal data to JSON
			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Printf("Error marshaling SSE data: %v", err)
				continue
			}

			// Send SSE message
			c.SSEvent("message", string(jsonData))
			c.Writer.Flush()

		case <-ticker.C:
			// Send keep-alive ping
			c.SSEvent("ping", map[string]string{
				"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
			})
			c.Writer.Flush()
		}
	}
}
