package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"franchise-saas-backend/internal/models"
)

type MockAuthHandler struct {
	users map[string]*models.User
}

func NewMockAuthHandler() *MockAuthHandler {
	return &MockAuthHandler{
		users: make(map[string]*models.User),
	}
}

func (h *MockAuthHandler) Register(c *gin.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
		return
	}

	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation error", "message": "email and password required"})
		return
	}

	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation error", "message": "password must be at least 6 characters"})
		return
	}

	if _, exists := h.users[req.Email]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}

	user := &models.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Role:      req.Role,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		CreatedAt: time.Now(),
	}
	h.users[req.Email] = user

	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": "mock_jwt_token_" + user.ID.String(),
	})
}

func (h *MockAuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, exists := h.users[req.Email]
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		User:         *user,
		Token:        "access_token_" + uuid.New().String(),
		RefreshToken: "refresh_token_" + uuid.New().String(),
	})
}

func (h *MockAuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	for _, user := range h.users {
		if user.ID.String() == userID.(string) {
			c.JSON(http.StatusOK, gin.H{"user": user})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewMockAuthHandler()

	r.POST("/api/v1/auth/register", handler.Register)
	r.POST("/api/v1/auth/login", handler.Login)
	r.GET("/api/v1/auth/me", handler.Me)

	return r
}

func TestAuthHandler_Register_Success(t *testing.T) {
	router := setupTestRouter()

	body := models.UserRegisterRequest{
		Email:     "newuser@example.com",
		Password:  "password123",
		Role:      "dealer",
		FirstName: "John",
		LastName:  "Doe",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "token")
	assert.NotNil(t, response["user"])
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name      string
		body      interface{}
		wantCode  int
		wantError string
	}{
		{
			name:      "empty email",
			body:      map[string]string{"email": "", "password": "123456"},
			wantCode:  http.StatusBadRequest,
			wantError: "validation",
		},
		{
			name:      "short password",
			body:      map[string]string{"email": "test@example.com", "password": "123"},
			wantCode:  http.StatusBadRequest,
			wantError: "validation",
		},
		{
			name:      "invalid json",
			body:      "not json",
			wantCode:  http.StatusBadRequest,
			wantError: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantError)
		})
	}
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	router := setupTestRouter()

	body1 := map[string]string{"email": "duplicate@example.com", "password": "password123"}
	jsonBody1, _ := json.Marshal(body1)
	req1, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusCreated, w1.Code)

	body2 := map[string]string{"email": "duplicate@example.com", "password": "password123"}
	jsonBody2, _ := json.Marshal(body2)
	req2, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	require.Equal(t, http.StatusConflict, w2.Code)
	assert.Contains(t, w2.Body.String(), "already exists")
}

func TestAuthHandler_Login_Success(t *testing.T) {
	router := setupTestRouter()

	registerBody := map[string]string{"email": "loginuser@example.com", "password": "password123"}
	jsonRegister, _ := json.Marshal(registerBody)
	registerReq, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonRegister))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	router.ServeHTTP(registerW, registerReq)
	require.Equal(t, http.StatusCreated, registerW.Code)

	loginBody := map[string]string{"email": "loginuser@example.com", "password": "password123"}
	jsonLogin, _ := json.Marshal(loginBody)
	loginReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonLogin))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	require.Equal(t, http.StatusOK, loginW.Code)

	var response models.AuthResponse
	err := json.Unmarshal(loginW.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.Token)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "loginuser@example.com", response.User.Email)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	router := setupTestRouter()

	body := map[string]string{"email": "nonexistent@example.com", "password": "password123"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid credentials")
}

func TestAuthHandler_Me_Unauthorized(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}