package api

import (
	"log"
	"mangahub/pkg/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Register endpoint
func (s *APIServer) register(c *gin.Context) {
	var req models.UserRegistration
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := s.UserService.Register(req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			log.Printf("Registration error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login endpoint
func (s *APIServer) login(c *gin.Context) {
	var req models.UserLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := s.UserService.Login(req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		} else {
			log.Printf("Login error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Establish TCP connection for this user
	if s.TCPUserManager != nil {
		go func() {
			err := s.TCPUserManager.ConnectUser(response.User.ID)
			if err != nil {
				log.Printf("Warning: Failed to establish TCP connection for user %s: %v", response.User.ID, err)
			}
		}()
	}

	c.JSON(http.StatusOK, response)
}

// Logout endpoint - disconnects user's TCP connection
func (s *APIServer) logout(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Disconnect user's TCP connection
	if s.TCPUserManager != nil {
		s.TCPUserManager.DisconnectUser(userID.(string))
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
