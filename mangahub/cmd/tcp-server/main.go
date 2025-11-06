package main

import (
	"log"
	"mangahub/internal/tcp"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	port := ":9000"

	server := tcp.NewProgressSyncServer(port)

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
