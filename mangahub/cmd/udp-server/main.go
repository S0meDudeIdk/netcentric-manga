package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"mangahub/internal/udp"
)

var server *udp.NotificationServer

func main() {
	// Get ports from environment or use defaults (fixed ports, no flags)
	udpPort := os.Getenv("UDP_SERVER_PORT")
	if udpPort == "" {
		udpPort = "9002"
	}
	if udpPort[0] != ':' {
		udpPort = ":" + udpPort
	}

	httpPort := os.Getenv("UDP_SERVER_HTTP_PORT")
	if httpPort == "" {
		httpPort = "9020"
	}
	if httpPort[0] != ':' {
		httpPort = ":" + httpPort
	}

	log.Printf("UDP Server starting on port %s", udpPort)
	log.Printf("UDP HTTP Trigger API starting on port %s", httpPort)

	server = udp.NewNotificationServer(udpPort)

	// Start UDP server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("UDP server stopped with error: %v", err)
		}
	}()

	log.Printf("UDP Notification Server listening on %s", udpPort)

	// Start HTTP trigger API
	go server.StartHTTPTrigger(httpPort)

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
