package main

import (
	"log"
	"mangahub/internal/tcp"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Get ports from environment or use defaults
	tcpPort := os.Getenv("TCP_SERVER_PORT")
	if tcpPort == "" {
		tcpPort = "9001"
	}
	if tcpPort[0] != ':' {
		tcpPort = ":" + tcpPort
	}

	httpPort := os.Getenv("TCP_SERVER_HTTP_PORT")
	if httpPort == "" {
		httpPort = "9010"
	}
	if httpPort[0] != ':' {
		httpPort = ":" + httpPort
	}

	log.Printf("TCP Server starting on port %s", tcpPort)
	log.Printf("TCP HTTP Trigger API starting on port %s", httpPort)

	server := tcp.NewProgressSyncServer(tcpPort)

	// Start HTTP trigger API in background
	go server.StartHTTPTrigger(httpPort)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down TCP server...")
		server.Close()
		os.Exit(0)
	}()

	log.Fatal(server.Start())
}
