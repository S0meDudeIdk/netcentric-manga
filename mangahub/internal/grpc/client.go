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
func (c *Client) SearchManga(ctx context.Context, query string, limit, offset int32, sort string) (*pb.SearchResponse, error) {
	req := &pb.SearchRequest{
		Query:  query,
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
	}

	log.Printf("gRPC Client: Searching manga with query: %s, limit: %d, offset: %d, sort: %s",
		query, limit, offset, sort)

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

// GetLibrary retrieves user's library via gRPC
func (c *Client) GetLibrary(ctx context.Context, userID string) (*pb.LibraryResponse, error) {
	req := &pb.LibraryRequest{
		UserId: userID,
	}

	log.Printf("gRPC Client: Getting library for user %s", userID)

	resp, err := c.client.GetLibrary(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetLibrary RPC failed: %v", err)
	}

	return resp, nil
}

// AddToLibrary adds a manga to user's library via gRPC
func (c *Client) AddToLibrary(ctx context.Context, userID, mangaID, status string) (*pb.AddToLibraryResponse, error) {
	req := &pb.AddToLibraryRequest{
		UserId:  userID,
		MangaId: mangaID,
		Status:  status,
	}

	log.Printf("gRPC Client: Adding manga %s to library for user %s", mangaID, userID)

	resp, err := c.client.AddToLibrary(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AddToLibrary RPC failed: %v", err)
	}

	return resp, nil
}

// RemoveFromLibrary removes a manga from user's library via gRPC
func (c *Client) RemoveFromLibrary(ctx context.Context, userID, mangaID string) (*pb.RemoveFromLibraryResponse, error) {
	req := &pb.RemoveFromLibraryRequest{
		UserId:  userID,
		MangaId: mangaID,
	}

	log.Printf("gRPC Client: Removing manga %s from library for user %s", mangaID, userID)

	resp, err := c.client.RemoveFromLibrary(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("RemoveFromLibrary RPC failed: %v", err)
	}

	return resp, nil
}

// GetLibraryStats retrieves user's library statistics via gRPC
func (c *Client) GetLibraryStats(ctx context.Context, userID string) (*pb.LibraryStatsResponse, error) {
	req := &pb.LibraryStatsRequest{
		UserId: userID,
	}

	log.Printf("gRPC Client: Getting library stats for user %s", userID)

	resp, err := c.client.GetLibraryStats(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetLibraryStats RPC failed: %v", err)
	}

	return resp, nil
}

// RateManga submits a rating for a manga via gRPC
func (c *Client) RateManga(ctx context.Context, userID, mangaID string, rating int32) (*pb.RatingResponse, error) {
	req := &pb.RatingRequest{
		UserId:  userID,
		MangaId: mangaID,
		Rating:  rating,
	}

	log.Printf("gRPC Client: Rating manga %s with %d for user %s", mangaID, rating, userID)

	resp, err := c.client.RateManga(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("RateManga RPC failed: %v", err)
	}

	return resp, nil
}

// GetMangaRatings retrieves rating statistics for a manga via gRPC
func (c *Client) GetMangaRatings(ctx context.Context, mangaID, userID string) (*pb.MangaRatingResponse, error) {
	req := &pb.MangaRatingRequest{
		MangaId: mangaID,
		UserId:  userID,
	}

	log.Printf("gRPC Client: Getting ratings for manga %s", mangaID)

	resp, err := c.client.GetMangaRatings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetMangaRatings RPC failed: %v", err)
	}

	return resp, nil
}

// DeleteRating deletes a user's rating for a manga via gRPC
func (c *Client) DeleteRating(ctx context.Context, userID, mangaID string) (*pb.DeleteRatingResponse, error) {
	req := &pb.DeleteRatingRequest{
		UserId:  userID,
		MangaId: mangaID,
	}

	log.Printf("gRPC Client: Deleting rating for manga %s by user %s", mangaID, userID)

	resp, err := c.client.DeleteRating(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("DeleteRating RPC failed: %v", err)
	}

	return resp, nil
}

// GetUserProfile gets a user's profile via gRPC
func (c *Client) GetUserProfile(ctx context.Context, userID string) (*pb.UserProfileResponse, error) {
	req := &pb.GetUserProfileRequest{
		UserId: userID,
	}

	log.Printf("gRPC Client: Getting profile for user %s", userID)

	resp, err := c.client.GetUserProfile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetUserProfile RPC failed: %v", err)
	}

	return resp, nil
}

// UpdateUserProfile updates a user's profile via gRPC
func (c *Client) UpdateUserProfile(ctx context.Context, userID, username, email string) (*pb.UpdateUserProfileResponse, error) {
	req := &pb.UpdateUserProfileRequest{
		UserId:   userID,
		Username: username,
		Email:    email,
	}

	log.Printf("gRPC Client: Updating profile for user %s", userID)

	resp, err := c.client.UpdateUserProfile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("UpdateUserProfile RPC failed: %v", err)
	}

	return resp, nil
}

// ChangePassword changes a user's password via gRPC
func (c *Client) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) (*pb.ChangePasswordResponse, error) {
	req := &pb.ChangePasswordRequest{
		UserId:      userID,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	log.Printf("gRPC Client: Changing password for user %s", userID)

	resp, err := c.client.ChangePassword(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ChangePassword RPC failed: %v", err)
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
