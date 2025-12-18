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
	MangaService *manga.Service
	UserService  *user.Service
	grpcServer   *grpc.Server
	tcpConn      net.Conn
	tcpMu        sync.Mutex
}

// NewServer creates a new gRPC server
func NewServer(mangaService *manga.Service, userService *user.Service) *Server {
	return &Server{
		MangaService: mangaService,
		UserService:  userService,
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

	update := tcp.ProgressUpdate{
		UserID:    userID,
		MangaID:   mangaID,
		Chapter:   chapter,
		Timestamp: time.Now().Unix(),
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
	log.Printf("gRPC SearchManga called with query: %s", req.Query)

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}

	searchReq := models.MangaSearchRequest{
		Query:  req.Query,
		Limit:  limit,
		Offset: int(req.Offset),
	}

	manga, err := s.MangaService.SearchManga(searchReq)
	if err != nil {
		return &pb.SearchResponse{
			Error: fmt.Sprintf("Failed to search manga: %v", err),
		}, nil
	}

	// Convert to protobuf manga slice
	pbManga := make([]*pb.Manga, len(manga))
	for i := range manga {
		pbManga[i] = modelMangaToPB(&manga[i])
	}

	return &pb.SearchResponse{
		Manga: pbManga,
		Total: int32(len(pbManga)),
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

// Start starts the gRPC server
func (s *Server) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
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
