package main

import (
	"log"
	"mangahub/internal/grpc"
	"mangahub/internal/manga"
	"mangahub/internal/user"
	"mangahub/pkg/database"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../../.env"); err != nil {
			log.Println("Warning: .env file not found, using environment variables or defaults")
		} else {
			log.Println("Loaded environment variables from ../../.env file")
		}
	} else {
		log.Println("Loaded environment variables from .env file")
	}

	// Initialize database
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Get gRPC server port from environment or use default
	grpcPort := os.Getenv("GRPC_SERVER_PORT")
	if grpcPort == "" {
		grpcPort = "9001"
	}

	// Create services
	mangaService := manga.NewService()
	userService := user.NewService()

	// Create gRPC server
	grpcServer := grpc.NewServer(mangaService, userService)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start gRPC server in a goroutine
	go func() {
		log.Printf("Starting gRPC MangaService server on port %s...", grpcPort)
		if err := grpcServer.Start(grpcPort); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Println("Received interrupt signal, shutting down...")

	// Stop gRPC server
	grpcServer.Stop()

	log.Println("gRPC server stopped gracefully")
}
