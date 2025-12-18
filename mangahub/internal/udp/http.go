package udp

import (
	"encoding/json"
	"log"
	"net/http"
)

// startHTTPTrigger starts a simple HTTP server for triggering broadcasts
func (s *NotificationServer) StartHTTPTrigger(port string) {
	http.HandleFunc("/trigger", s.triggerHandler)
	http.HandleFunc("/health", s.healthHandler)

	log.Printf("HTTP trigger API listening on %s", port)
	log.Printf("  POST %s/trigger - Trigger UDP broadcast", port)
	log.Printf("  GET  %s/health  - Health check", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Printf("HTTP trigger server error: %v", err)
	}
}

// triggerHandler handles broadcast trigger requests
func (s *NotificationServer) triggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var notification Notification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Broadcast the notification
	if err := s.BroadcastNotification(notification); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Notification broadcasted",
		"clients": s.GetClientCount(),
	})
}

// healthHandler handles health check requests
func (s *NotificationServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":             "healthy",
		"registered_clients": s.GetClientCount(),
	})
}
