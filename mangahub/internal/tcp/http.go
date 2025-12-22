package tcp

import (
	"encoding/json"
	"log"
	"net/http"
)

// HTTP handlers for triggering broadcasts

func (s *ProgressSyncServer) triggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var update ProgressUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Trigger broadcast
	if err := s.TriggerBroadcast(update); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Progress update broadcasted",
		"clients": s.GetClientCount(),
	})
}

func (s *ProgressSyncServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":            "healthy",
		"connected_clients": s.GetClientCount(),
	})
}

func (s *ProgressSyncServer) StartHTTPTrigger(port string) {
	http.HandleFunc("/trigger", s.triggerHandler)
	http.HandleFunc("/health", s.healthHandler)

	// Bind to 0.0.0.0 to accept connections from all network interfaces
	bindAddr := "0.0.0.0" + port
	log.Printf("HTTP trigger API listening on %s", bindAddr)
	log.Printf("  POST %s/trigger - Trigger TCP broadcast", port)
	log.Printf("  GET  %s/health  - Health check", port)

	if err := http.ListenAndServe(bindAddr, nil); err != nil {
		log.Printf("HTTP trigger server error: %v", err)
	}
}
