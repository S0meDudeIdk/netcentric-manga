package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mangahub/internal/udp"
)

func main() {
	port := flag.String("port", ":8081", "UDP server listen address (host:port)")
	flag.Parse()

	server := udp.NewNotificationServer(*port)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("UDP server stopped with error: %v", err)
		}
	}()

	log.Printf("UDP Notification Server declared on %s", *port)

	// wait for termination signal for a graceful exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutdown signal received, exiting...")
	// allow small time for logs / cleanup (no Stop() defined)
	time.Sleep(200 * time.Millisecond)
}
