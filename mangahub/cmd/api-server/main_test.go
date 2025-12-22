// package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"mangahub/pkg/database"
// 	"mangahub/pkg/models"
// 	"net/http"
// 	"net/http/httptest"
// 	"os"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/suite"
// )

// // APITestSuite defines the test suite
// type APITestSuite struct {
// 	suite.Suite
// 	server *APIServer
// 	router *gin.Engine
// 	token  string
// 	userID string
// }

// // SetupSuite runs before all tests
// func (suite *APITestSuite) SetupSuite() {
// 	// Set test environment
// 	gin.SetMode(gin.TestMode)
// 	os.Setenv("GIN_MODE", "test")

// 	// Initialize test database
// 	err := database.InitDatabase()
// 	assert.NoError(suite.T(), err)

// 	// Create test server
// 	suite.server = NewAPIServer()
// 	suite.router = suite.server.Router

// 	// Create test user and get token
// 	suite.createTestUser()
// }

// // TearDownSuite runs after all tests
// func (suite *APITestSuite) TearDownSuite() {
// 	database.Close()
// }

// // createTestUser creates a test user and gets authentication token
// func (suite *APITestSuite) createTestUser() {
// 	// Register test user
// 	registerReq := models.UserRegistration{
// 		Username: "testuser",
// 		Email:    "test@example.com",
// 		Password: "testpassword123",
// 	}

// 	reqBody, _ := json.Marshal(registerReq)
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
// 	req.Header.Set("Content-Type", "application/json")

// 	suite.router.ServeHTTP(w, req)

// 	if w.Code == http.StatusCreated {
// 		var response models.AuthResponse
// 		json.Unmarshal(w.Body.Bytes(), &response)
// 		suite.token = response.Token
// 		suite.userID = response.User.ID
// 	} else {
// 		// User might already exist, try login
// 		loginReq := models.UserLogin{
// 			Email:    "test@example.com",
// 			Password: "testpassword123",
// 		}

// 		reqBody, _ := json.Marshal(loginReq)
// 		w := httptest.NewRecorder()
// 		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
// 		req.Header.Set("Content-Type", "application/json")

// 		suite.router.ServeHTTP(w, req)

// 		var response models.LoginResponse
// 		json.Unmarshal(w.Body.Bytes(), &response)
// 		suite.token = response.Token
// 		suite.userID = response.User.ID
// 	}
// }

// // makeAuthenticatedRequest makes a request with authentication
// func (suite *APITestSuite) makeAuthenticatedRequest(method, url string, body interface{}) *httptest.ResponseRecorder {
// 	var reqBody *bytes.Buffer
// 	if body != nil {
// 		jsonBody, _ := json.Marshal(body)
// 		reqBody = bytes.NewBuffer(jsonBody)
// 	} else {
// 		reqBody = bytes.NewBuffer([]byte{})
// 	}

// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest(method, url, reqBody)
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+suite.token)

// 	suite.router.ServeHTTP(w, req)
// 	return w
// }

// // TestHealthCheck tests the health check endpoint
// func (suite *APITestSuite) TestHealthCheck() {
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/health", nil)
// 	suite.router.ServeHTTP(w, req)

// 	assert.Equal(suite.T(), http.StatusOK, w.Code)

// 	var response map[string]interface{}
// 	err := json.Unmarshal(w.Body.Bytes(), &response)
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), "ok", response["status"])
// }

// // TestUserRegistration tests user registration
// func (suite *APITestSuite) TestUserRegistration() {
// 	registerReq := models.UserRegistration{
// 		Username: "newuser",
// 		Email:    "newuser@example.com",
// 		Password: "password123",
// 	}

// 	reqBody, _ := json.Marshal(registerReq)
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
// 	req.Header.Set("Content-Type", "application/json")

// 	suite.router.ServeHTTP(w, req)

// 	assert.Equal(suite.T(), http.StatusCreated, w.Code)

// 	var response models.AuthResponse
// 	err := json.Unmarshal(w.Body.Bytes(), &response)
// 	assert.NoError(suite.T(), err)
// 	assert.NotEmpty(suite.T(), response.Token)
// 	assert.Equal(suite.T(), "newuser", response.User.Username)
// }

// // TestUserLogin tests user login
// func (suite *APITestSuite) TestUserLogin() {
// 	loginReq := models.UserLogin{
// 		Email:    "test@example.com",
// 		Password: "testpassword123",
// 	}

// 	reqBody, _ := json.Marshal(loginReq)
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
// 	req.Header.Set("Content-Type", "application/json")

// 	suite.router.ServeHTTP(w, req)

// 	assert.Equal(suite.T(), http.StatusOK, w.Code)

// 	var response models.LoginResponse
// 	err := json.Unmarshal(w.Body.Bytes(), &response)
// 	assert.NoError(suite.T(), err)
// 	assert.NotEmpty(suite.T(), response.Token)
// }

// // TestGetProfile tests getting user profile
// func (suite *APITestSuite) TestGetProfile() {
// 	w := suite.makeAuthenticatedRequest("GET", "/api/v1/users/profile", nil)

// 	assert.Equal(suite.T(), http.StatusOK, w.Code)

// 	var response models.UserResponse
// 	err := json.Unmarshal(w.Body.Bytes(), &response)
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), "testuser", response.Username)
// }

// // TestSearchManga tests manga search
// func (suite *APITestSuite) TestSearchManga() {
// 	w := suite.makeAuthenticatedRequest("GET", "/api/v1/manga?query=One Piece", nil)

// 	assert.Equal(suite.T(), http.StatusOK, w.Code)

// 	var response map[string]interface{}
// 	err := json.Unmarshal(w.Body.Bytes(), &response)
// 	assert.NoError(suite.T(), err)
// 	assert.Contains(suite.T(), response, "manga")
// 	assert.Contains(suite.T(), response, "count")
// }

// // TestGetManga tests getting specific manga
// func (suite *APITestSuite) TestGetManga() {
// 	w := suite.makeAuthenticatedRequest("GET", "/api/v1/manga/one-piece", nil)

// 	assert.Equal(suite.T(), http.StatusOK, w.Code)

// 	var response models.Manga
// 	err := json.Unmarshal(w.Body.Bytes(), &response)
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), "one-piece", response.ID)
// 	assert.Equal(suite.T(), "One Piece", response.Title)
// }

// // TestAddToLibrary tests adding manga to library
// func (suite *APITestSuite) TestAddToLibrary() {
// 	addReq := models.AddToLibraryRequest{
// 		MangaID: "one-piece",
// 		Status:  "reading",
// 	}

// 	w := suite.makeAuthenticatedRequest("POST", "/api/v1/users/library", addReq)

// 	assert.Equal(suite.T(), http.StatusOK, w.Code)

// 	var response map[string]interface{}
// 	err := json.Unmarshal(w.Body.Bytes(), &response)
// 	assert.NoError(suite.T(), err)
// 	assert.Contains(suite.T(), response, "message")
// }

// // TestUpdateProgress tests updating reading progress
// func (suite *APITestSuite) TestUpdateProgress() {
// 	// First add manga to library
// 	addReq := models.AddToLibraryRequest{
// 		MangaID: "naruto",
// 		Status:  "reading",
// 	}
// 	suite.makeAuthenticatedRequest("POST", "/api/v1/users/library", addReq)

// 	// Then update progress
// 	updateReq := models.UpdateProgressRequest{
// 		MangaID:        "naruto",
// 		CurrentChapter: 50,
// 		Status:         "reading",
// 	}

// 	w := suite.makeAuthenticatedRequest("PUT", "/api/v1/users/progress", updateReq)

// 	assert.Equal(suite.T(), http.StatusOK, w.Code)

// 	var response map[string]interface{}
// 	err := json.Unmarshal(w.Body.Bytes(), &response)
// 	assert.NoError(suite.T(), err)
// 	assert.Contains(suite.T(), response, "message")
// }

// // TestGetLibrary tests getting user library
// func (suite *APITestSuite) TestGetLibrary() {
// 	w := suite.makeAuthenticatedRequest("GET", "/api/v1/users/library", nil)

// 	assert.Equal(suite.T(), http.StatusOK, w.Code)

// 	var response models.UserLibrary
// 	err := json.Unmarshal(w.Body.Bytes(), &response)
// 	assert.NoError(suite.T(), err)
// 	// Library should have the manga we added in previous tests
// }

// // TestGetLibraryStats tests getting library statistics
// func (suite *APITestSuite) TestGetLibraryStats() {
// 	w := suite.makeAuthenticatedRequest("GET", "/api/v1/users/library/stats", nil)

// 	assert.Equal(suite.T(), http.StatusOK, w.Code)

// 	var response models.LibraryStatsResponse
// 	err := json.Unmarshal(w.Body.Bytes(), &response)
// 	assert.NoError(suite.T(), err)
// 	assert.GreaterOrEqual(suite.T(), response.TotalManga, 0)
// }

// // TestGetGenres tests getting all genres
// func (suite *APITestSuite) TestGetGenres() {
// 	w := suite.makeAuthenticatedRequest("GET", "/api/v1/manga/genres", nil)

// 	assert.Equal(suite.T(), http.StatusOK, w.Code)

// 	var response map[string]interface{}
// 	err := json.Unmarshal(w.Body.Bytes(), &response)
// 	assert.NoError(suite.T(), err)
// 	assert.Contains(suite.T(), response, "genres")
// 	assert.Contains(suite.T(), response, "count")
// }

// // TestUnauthorizedAccess tests access without authentication
// func (suite *APITestSuite) TestUnauthorizedAccess() {
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/api/v1/users/profile", nil)
// 	suite.router.ServeHTTP(w, req)

// 	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
// }

// // TestInvalidManga tests accessing non-existent manga
// func (suite *APITestSuite) TestInvalidManga() {
// 	w := suite.makeAuthenticatedRequest("GET", "/api/v1/manga/non-existent", nil)

// 	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
// }

// // TestRateLimit tests rate limiting (simplified)
// func (suite *APITestSuite) TestRateLimit() {
// 	// This is a simplified test - in a real scenario, you'd make many requests quickly
// 	w := suite.makeAuthenticatedRequest("GET", "/api/v1/manga", nil)
// 	assert.NotEqual(suite.T(), http.StatusTooManyRequests, w.Code)
// }

// // TestInvalidJSON tests invalid JSON input
// func (suite *APITestSuite) TestInvalidJSON() {
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer([]byte("invalid json")))
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+suite.token)

// 	suite.router.ServeHTTP(w, req)

// 	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
// }

// // TestSuite runs the test suite
// func TestAPISuite(t *testing.T) {
// 	suite.Run(t, new(APITestSuite))
// }