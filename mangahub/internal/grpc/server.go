package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mangahub/internal/manga"
	"mangahub/internal/tcp"
	"mangahub/internal/user"
	"mangahub/pkg/models"
	pb "mangahub/proto"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// Server implements the gRPC MangaService
type Server struct {
	pb.UnimplementedMangaServiceServer
	MangaService  *manga.Service
	UserService   *user.Service
	RatingService *manga.RatingService
	grpcServer    *grpc.Server
	tcpConn       net.Conn
	tcpMu         sync.Mutex
}

// NewServer creates a new gRPC server
func NewServer(mangaService *manga.Service, userService *user.Service, ratingService *manga.RatingService) *Server {
	return &Server{
		MangaService:  mangaService,
		UserService:   userService,
		RatingService: ratingService,
	}
}

// ConnectToTCP establishes connection to TCP server for broadcasting
func (s *Server) ConnectToTCP(tcpAddress string) error {
	conn, err := net.Dial("tcp", tcpAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to TCP server: %v", err)
	}
	s.tcpConn = conn
	log.Printf("gRPC server connected to TCP server at %s", tcpAddress)
	return nil
}

// broadcastProgress sends progress update to TCP server
func (s *Server) broadcastProgress(userID, mangaID string, chapter int) {
	if s.tcpConn == nil {
		log.Println("Warning: TCP connection not established, skipping broadcast")
		return
	}

	manga, err := s.MangaService.GetManga(mangaID)
	if err != nil {
		log.Printf("Failed to get manga for TCP broadcast: %v", err)
		return
	}

	user, err := s.UserService.GetProfile(userID)
	if err != nil {
		log.Printf("Failed to get user profile for TCP broadcast: %v", err)
		return
	}

	update := tcp.ProgressUpdate{
		UserID:     userID,
		Username:   user.Username,
		MangaTitle: manga.Title,
		Chapter:    chapter,
		Timestamp:  time.Now().Unix(),
	}

	message, err := json.Marshal(update)
	if err != nil {
		log.Printf("Error marshaling progress update: %v", err)
		return
	}

	message = append(message, '\n')

	s.tcpMu.Lock()
	defer s.tcpMu.Unlock()

	s.tcpConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = s.tcpConn.Write(message)
	if err != nil {
		log.Printf("Error sending progress update to TCP server: %v", err)
		// Try to reconnect on next update
		s.tcpConn = nil
	} else {
		log.Printf("Broadcasted progress update via TCP: User=%s, Manga=%s, Chapter=%d", userID, mangaID, chapter)
	}
}

// Helper function to convert models.Manga to pb.Manga
func modelMangaToPB(m *models.Manga) *pb.Manga {
	if m == nil {
		return nil
	}
	return &pb.Manga{
		Id:              m.ID,
		Title:           m.Title,
		Author:          m.Author,
		Genres:          m.Genres,
		Status:          m.Status,
		TotalChapters:   int32(m.TotalChapters),
		Description:     m.Description,
		CoverUrl:        m.CoverURL,
		PublicationYear: int32(m.PublicationYear),
		Rating:          m.Rating,
		CreatedAt:       m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// GetManga retrieves a single manga by ID
func (s *Server) GetManga(ctx context.Context, req *pb.GetMangaRequest) (*pb.MangaResponse, error) {
	log.Printf("gRPC GetManga called with ID: %s", req.Id)

	manga, err := s.MangaService.GetManga(req.Id)
	if err != nil {
		return &pb.MangaResponse{
			Error: fmt.Sprintf("Failed to get manga: %v", err),
		}, nil
	}

	return &pb.MangaResponse{
		Manga: modelMangaToPB(manga),
	}, nil
}

// SearchManga searches for manga by query
func (s *Server) SearchManga(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	log.Printf("gRPC SearchManga called with query: %s, limit: %d, offset: %d, sort: %s",
		req.Query, req.Limit, req.Offset, req.Sort)

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}

	searchReq := models.MangaSearchRequest{
		Query:  req.Query,
		Limit:  limit,
		Offset: int(req.Offset),
		Sort:   req.Sort,
	}

	manga, err := s.MangaService.SearchManga(searchReq)
	if err != nil {
		return &pb.SearchResponse{
			Error: fmt.Sprintf("Failed to search manga: %v", err),
		}, nil
	}

	// Get total count for pagination
	totalCount, err := s.MangaService.GetMangaCount(searchReq)
	if err != nil {
		log.Printf("Warning: Failed to get manga count: %v", err)
		totalCount = len(manga) // Fallback to current result count
	}

	// Convert to protobuf manga slice
	pbManga := make([]*pb.Manga, len(manga))
	for i := range manga {
		pbManga[i] = modelMangaToPB(&manga[i])
	}

	return &pb.SearchResponse{
		Manga: pbManga,
		Total: int32(totalCount),
	}, nil
}

// UpdateProgress updates user's reading progress
func (s *Server) UpdateProgress(ctx context.Context, req *pb.ProgressRequest) (*pb.ProgressResponse, error) {
	log.Printf("gRPC UpdateProgress called for user: %s, manga: %s", req.UserId, req.MangaId)

	updateReq := models.UpdateProgressRequest{
		MangaID:        req.MangaId,
		CurrentChapter: int(req.CurrentChapter),
		Status:         req.Status,
	}

	err := s.UserService.UpdateProgress(req.UserId, updateReq)
	if err != nil {
		return &pb.ProgressResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to update progress: %v", err),
		}, nil
	}

	// Trigger TCP broadcast for real-time sync
	go s.broadcastProgress(req.UserId, req.MangaId, int(req.CurrentChapter))

	return &pb.ProgressResponse{
		Success: true,
		Message: "Progress updated successfully",
	}, nil
}

// GetLibrary retrieves user's manga library
func (s *Server) GetLibrary(ctx context.Context, req *pb.LibraryRequest) (*pb.LibraryResponse, error) {
	log.Printf("gRPC GetLibrary called for user: %s", req.UserId)

	library, err := s.UserService.GetLibrary(req.UserId)
	if err != nil {
		return &pb.LibraryResponse{
			Error: fmt.Sprintf("Failed to get library: %v", err),
		}, nil
	}

	// Convert models.UserProgress to pb.UserProgress
	convertProgress := func(progressList []models.UserProgress) []*pb.UserProgress {
		result := make([]*pb.UserProgress, len(progressList))
		for i, p := range progressList {
			result[i] = &pb.UserProgress{
				MangaId:        p.MangaID,
				CurrentChapter: int32(p.CurrentChapter),
				Status:         p.Status,
				LastUpdated:    p.LastUpdated.Format("2006-01-02T15:04:05Z07:00"),
				Title:          p.Title,
				Author:         p.Author,
				CoverUrl:       p.CoverURL,
			}
		}
		return result
	}

	return &pb.LibraryResponse{
		Reading:    convertProgress(library.Reading),
		Completed:  convertProgress(library.Completed),
		PlanToRead: convertProgress(library.PlanToRead),
		Dropped:    convertProgress(library.Dropped),
		OnHold:     convertProgress(library.OnHold),
		ReReading:  convertProgress(library.ReReading),
	}, nil
}

// AddToLibrary adds a manga to user's library
func (s *Server) AddToLibrary(ctx context.Context, req *pb.AddToLibraryRequest) (*pb.AddToLibraryResponse, error) {
	log.Printf("gRPC AddToLibrary called for user: %s, manga: %s", req.UserId, req.MangaId)

	addReq := models.AddToLibraryRequest{
		MangaID: req.MangaId,
		Status:  req.Status,
	}

	err := s.UserService.AddToLibrary(req.UserId, addReq)
	if err != nil {
		return &pb.AddToLibraryResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to add to library: %v", err),
		}, nil
	}

	return &pb.AddToLibraryResponse{
		Success: true,
		Message: "Manga added to library successfully",
	}, nil
}

// RemoveFromLibrary removes a manga from user's library
func (s *Server) RemoveFromLibrary(ctx context.Context, req *pb.RemoveFromLibraryRequest) (*pb.RemoveFromLibraryResponse, error) {
	log.Printf("gRPC RemoveFromLibrary called for user: %s, manga: %s", req.UserId, req.MangaId)

	err := s.UserService.RemoveFromLibrary(req.UserId, req.MangaId)
	if err != nil {
		return &pb.RemoveFromLibraryResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to remove from library: %v", err),
		}, nil
	}

	return &pb.RemoveFromLibraryResponse{
		Success: true,
		Message: "Manga removed from library successfully",
	}, nil
}

// GetLibraryStats retrieves user's library statistics
func (s *Server) GetLibraryStats(ctx context.Context, req *pb.LibraryStatsRequest) (*pb.LibraryStatsResponse, error) {
	log.Printf("gRPC GetLibraryStats called for user: %s", req.UserId)

	stats, err := s.UserService.GetLibraryStats(req.UserId)
	if err != nil {
		return &pb.LibraryStatsResponse{
			Error: fmt.Sprintf("Failed to get library stats: %v", err),
		}, nil
	}

	return &pb.LibraryStatsResponse{
		TotalManga:        int32(stats.TotalManga),
		Reading:           int32(stats.Reading),
		Completed:         int32(stats.Completed),
		PlanToRead:        int32(stats.PlanToRead),
		Dropped:           int32(stats.Dropped),
		OnHold:            0, // Not tracked in current model
		ReReading:         0, // Not tracked in current model
		TotalChaptersRead: int32(stats.TotalChapters),
	}, nil
}

// RateManga adds or updates a user's rating for a manga
func (s *Server) RateManga(ctx context.Context, req *pb.RatingRequest) (*pb.RatingResponse, error) {
	log.Printf("gRPC RateManga called for user: %s, manga: %s, rating: %d", req.UserId, req.MangaId, req.Rating)

	err := s.RatingService.RateManga(req.UserId, req.MangaId, int(req.Rating))
	if err != nil {
		return &pb.RatingResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to rate manga: %v", err),
		}, nil
	}

	// Get updated stats
	stats, err := s.RatingService.GetMangaRatingStats(req.MangaId, req.UserId)
	if err != nil {
		// Return success even if we can't get stats
		return &pb.RatingResponse{
			Success: true,
			Message: "Rating submitted successfully",
		}, nil
	}

	return &pb.RatingResponse{
		Success:       true,
		Message:       "Rating submitted successfully",
		AverageRating: stats.AverageRating,
		TotalRatings:  int32(stats.TotalRatings),
	}, nil
}

// GetMangaRatings retrieves rating statistics for a manga
func (s *Server) GetMangaRatings(ctx context.Context, req *pb.MangaRatingRequest) (*pb.MangaRatingResponse, error) {
	log.Printf("gRPC GetMangaRatings called for manga: %s", req.MangaId)

	stats, err := s.RatingService.GetMangaRatingStats(req.MangaId, req.UserId)
	if err != nil {
		return &pb.MangaRatingResponse{
			Error: fmt.Sprintf("Failed to get ratings: %v", err),
		}, nil
	}

	userRating := int32(0)
	if stats.UserRating != nil {
		userRating = int32(*stats.UserRating)
	}

	// Convert rating distribution from map[int]int to map[int32]int32
	ratingDistribution := make(map[int32]int32)
	for rating, count := range stats.RatingDistribution {
		ratingDistribution[int32(rating)] = int32(count)
	}

	return &pb.MangaRatingResponse{
		AverageRating:      stats.AverageRating,
		TotalRatings:       int32(stats.TotalRatings),
		UserRating:         userRating,
		RatingDistribution: ratingDistribution,
	}, nil
}

// DeleteRating removes a user's rating for a manga
func (s *Server) DeleteRating(ctx context.Context, req *pb.DeleteRatingRequest) (*pb.DeleteRatingResponse, error) {
	log.Printf("gRPC DeleteRating called for user: %s, manga: %s", req.UserId, req.MangaId)

	err := s.RatingService.DeleteRating(req.UserId, req.MangaId)
	if err != nil {
		return &pb.DeleteRatingResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to delete rating: %v", err),
		}, nil
	}

	return &pb.DeleteRatingResponse{
		Success: true,
		Message: "Rating deleted successfully",
	}, nil
}

// GetUserProfile retrieves a user's profile
func (s *Server) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.UserProfileResponse, error) {
	log.Printf("gRPC GetUserProfile called for user: %s", req.UserId)

	profile, err := s.UserService.GetProfile(req.UserId)
	if err != nil {
		return &pb.UserProfileResponse{
			Error: fmt.Sprintf("Failed to get user profile: %v", err),
		}, nil
	}

	return &pb.UserProfileResponse{
		Profile: &pb.UserProfile{
			Id:        profile.ID,
			Username:  profile.Username,
			Email:     profile.Email,
			CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// UpdateUserProfile updates a user's profile
func (s *Server) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UpdateUserProfileResponse, error) {
	log.Printf("gRPC UpdateUserProfile called for user: %s", req.UserId)

	profile, err := s.UserService.UpdateProfile(req.UserId, req.Username, req.Email)
	if err != nil {
		return &pb.UpdateUserProfileResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to update profile: %v", err),
		}, nil
	}

	return &pb.UpdateUserProfileResponse{
		Success: true,
		Message: "Profile updated successfully",
		Profile: &pb.UserProfile{
			Id:        profile.ID,
			Username:  profile.Username,
			Email:     profile.Email,
			CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// ChangePassword changes a user's password
func (s *Server) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	log.Printf("gRPC ChangePassword called for user: %s", req.UserId)

	err := s.UserService.ChangePassword(req.UserId, req.OldPassword, req.NewPassword)
	if err != nil {
		return &pb.ChangePasswordResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to change password: %v", err),
		}, nil
	}

	return &pb.ChangePasswordResponse{
		Success: true,
		Message: "Password changed successfully",
	}, nil
}

// Start starts the gRPC server
func (s *Server) Start(port string) error {
	lis, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.grpcServer = grpc.NewServer()

	// Register the MangaService with the gRPC server
	pb.RegisterMangaServiceServer(s.grpcServer, s)

	log.Printf("gRPC server listening on port %s", port)
	log.Println("gRPC MangaService registered with methods:")
	log.Println("  - GetManga(GetMangaRequest) returns (MangaResponse)")
	log.Println("  - SearchManga(SearchRequest) returns (SearchResponse)")
	log.Println("  - UpdateProgress(ProgressRequest) returns (ProgressResponse)")
	log.Println("  - GetLibrary(LibraryRequest) returns (LibraryResponse)")
	log.Println("  - AddToLibrary(AddToLibraryRequest) returns (AddToLibraryResponse)")
	log.Println("  - RemoveFromLibrary(RemoveFromLibraryRequest) returns (RemoveFromLibraryResponse)")
	log.Println("  - GetLibraryStats(LibraryStatsRequest) returns (LibraryStatsResponse)")
	log.Println("  - RateManga(RatingRequest) returns (RatingResponse)")
	log.Println("  - GetMangaRatings(MangaRatingRequest) returns (MangaRatingResponse)")
	log.Println("  - DeleteRating(DeleteRatingRequest) returns (DeleteRatingResponse)")

	return s.grpcServer.Serve(lis)
}

// Stop stops the gRPC server gracefully
func (s *Server) Stop() {
	if s.grpcServer != nil {
		log.Println("Stopping gRPC server...")
		s.grpcServer.GracefulStop()
	}

	// Close TCP connection
	s.tcpMu.Lock()
	defer s.tcpMu.Unlock()
	if s.tcpConn != nil {
		log.Println("Closing TCP connection...")
		s.tcpConn.Close()
		s.tcpConn = nil
	}
}
