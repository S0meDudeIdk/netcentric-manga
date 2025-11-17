package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "mangahub/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client represents a gRPC client for MangaService
type Client struct {
	conn   *grpc.ClientConn
	client pb.MangaServiceClient
}

// NewClient creates a new gRPC client
func NewClient(address string) (*Client, error) {
	// Set up connection options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Dial the gRPC server
	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %v", err)
	}

	log.Printf("Connected to gRPC server at %s", address)

	// Create the gRPC client using the generated code
	client := pb.NewMangaServiceClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// GetManga retrieves a manga by ID via gRPC
func (c *Client) GetManga(ctx context.Context, id string) (*pb.MangaResponse, error) {
	req := &pb.GetMangaRequest{
		Id: id,
	}

	log.Printf("gRPC Client: Getting manga with ID: %s", id)

	// Call the generated gRPC method
	resp, err := c.client.GetManga(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetManga RPC failed: %v", err)
	}

	return resp, nil
}

// SearchManga searches for manga via gRPC
func (c *Client) SearchManga(ctx context.Context, query string, limit int32) (*pb.SearchResponse, error) {
	req := &pb.SearchRequest{
		Query: query,
		Limit: limit,
	}

	log.Printf("gRPC Client: Searching manga with query: %s", query)

	// Call the generated gRPC method
	resp, err := c.client.SearchManga(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("SearchManga RPC failed: %v", err)
	}

	return resp, nil
}

// UpdateProgress updates reading progress via gRPC
func (c *Client) UpdateProgress(ctx context.Context, userID, mangaID string, currentChapter int32, status string) (*pb.ProgressResponse, error) {
	req := &pb.ProgressRequest{
		UserId:         userID,
		MangaId:        mangaID,
		CurrentChapter: currentChapter,
		Status:         status,
	}

	log.Printf("gRPC Client: Updating progress for user %s, manga %s", userID, mangaID)

	// Call the generated gRPC method
	resp, err := c.client.UpdateProgress(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("UpdateProgress RPC failed: %v", err)
	}

	return resp, nil
}

// Close closes the gRPC client connection
func (c *Client) Close() error {
	if c.conn != nil {
		log.Println("Closing gRPC client connection")
		return c.conn.Close()
	}
	return nil
}

// Ping checks if the gRPC server is reachable
func (c *Client) Ping(ctx context.Context) error {
	// Simple health check
	log.Println("gRPC Client: Pinging server")

	if c.conn == nil {
		return fmt.Errorf("no connection established")
	}

	state := c.conn.GetState()
	log.Printf("gRPC connection state: %v", state)

	return nil
}
