package handlers

import (
	"log"
	"net/http"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Validation Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role == "" {
		req.Role = models.RoleFranchisor
	}

	user := &models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Role:      req.Role,
	}

	if user.Role == models.RoleFranchisor {
		companyName := user.FirstName + " " + user.LastName
		_, err := h.service.CreateUserWithTenant(user, req.Password, companyName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		_, err := h.service.CreateUser(user, req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// ИСПРАВЛЕНО: Передаем user.SalonID
	token, refresh, _ := h.service.GenerateTokens(user.ID, user.Email, user.Role, user.TenantID, user.SalonID)

	c.JSON(http.StatusCreated, models.AuthResponse{
		User:         *user,
		Token:        token,
		RefreshToken: refresh,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Authenticate(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// ИСПРАВЛЕНО: Передаем user.SalonID (чтобы токен содержал актуальный салон)
	token, refresh, _ := h.service.GenerateTokens(user.ID, user.Email, user.Role, user.TenantID, user.SalonID)

	c.JSON(http.StatusOK, models.AuthResponse{
		User:         *user,
		Token:        token,
		RefreshToken: refresh,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	access, refresh, err := h.service.RefreshTokens(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         access,
		"refresh_token": refresh,
	})
}
