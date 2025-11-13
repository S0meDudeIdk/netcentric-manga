package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"mangahub/internal/udp"
)

var server *udp.NotificationServer

func main() {
	port := flag.String("port", ":8081", "UDP server listen address (host:port)")
	httpPort := flag.String("http", ":8082", "HTTP trigger API port")
	flag.Parse()

	server = udp.NewNotificationServer(*port)

	// Start UDP server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("UDP server stopped with error: %v", err)
		}
	}()

	log.Printf("UDP Notification Server listening on %s", *port)

	// Start HTTP trigger API
	go startHTTPTrigger(*httpPort)

	// wait for termination signal for a graceful exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutdown signal received, closing server...")
	if err := server.Close(); err != nil {
		log.Printf("Error closing server: %v", err)
	}
	log.Println("Server shut down gracefully")
}

// startHTTPTrigger starts a simple HTTP server for triggering broadcasts
func startHTTPTrigger(port string) {
	http.HandleFunc("/trigger", triggerHandler)
	http.HandleFunc("/health", healthHandler)
	
	log.Printf("HTTP trigger API listening on %s", port)
	log.Printf("  POST %s/trigger - Trigger UDP broadcast", port)
	log.Printf("  GET  %s/health  - Health check", port)
	
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Printf("HTTP trigger server error: %v", err)
	}
}

// triggerHandler handles broadcast trigger requests
func triggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var notification udp.Notification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Broadcast the notification
	if err := server.BroadcastNotification(notification); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Notification broadcasted",
		"clients": server.GetClientCount(),
	})
}

// healthHandler handles health check requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "healthy",
		"registered_clients": server.GetClientCount(),
	})
}
