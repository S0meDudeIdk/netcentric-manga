package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// streamProgressUpdates handles SSE endpoint for TCP progress updates
func (s *APIServer) streamProgressUpdates(c *gin.Context) {
	// Create SSE client
	clientID := fmt.Sprintf("progress-%s-%d", uuid.New().String()[:8], time.Now().Unix())
	client := &SSEClient{
		ID:      clientID,
		Channel: make(chan interface{}, 10),
		Context: c,
	}

	// Add client to hub
	s.SSEHub.AddProgressClient(client)

	// Stream SSE
	StreamSSE(c, client, s.SSEHub.RemoveProgressClient)
}

// streamNotifications handles SSE endpoint for UDP notifications
func (s *APIServer) streamNotifications(c *gin.Context) {
	// Create SSE client
	clientID := fmt.Sprintf("notification-%s-%d", uuid.New().String()[:8], time.Now().Unix())
	client := &SSEClient{
		ID:      clientID,
		Channel: make(chan interface{}, 10),
		Context: c,
	}

	// Add client to hub
	s.SSEHub.AddNotificationClient(client)

	// Stream SSE
	StreamSSE(c, client, s.SSEHub.RemoveNotificationClient)
}
