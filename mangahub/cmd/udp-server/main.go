package main

import (
	"flag"
	"log"
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
	go server.StartHTTPTrigger(*httpPort)

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
