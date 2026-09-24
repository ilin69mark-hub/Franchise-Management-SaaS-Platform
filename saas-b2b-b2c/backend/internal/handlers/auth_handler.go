package handlers

import (
	"errors"
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

// allowedRegisterRoles - роли, которые пользователь может выбрать при
// самостоятельной регистрации. Ни super_admin, ни другие неизвестные роли
// недопустимы (иначе это privilege escalation).
var allowedRegisterRoles = map[models.Role]bool{
	models.RoleFranchisor:        true,
	models.RoleFranchisorManager: true,
	models.RoleDealer:            true,
	models.RoleDealerManager:     true,
}

func isAllowedRegisterRole(role models.Role) bool {
	return allowedRegisterRoles[role]
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Validation Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	if req.Role == "" {
		req.Role = models.RoleFranchisor
	}

	if !isAllowedRegisterRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
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
	token, refresh, err := h.service.GenerateTokens(user.ID, user.Email, user.Role, user.TenantID, user.SalonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}
	setAuthCookies(c, token, refresh)

	c.JSON(http.StatusCreated, models.AuthResponse{
		User:         *user,
		Token:        token,
		RefreshToken: refresh,
	})
}

func setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	// httpOnly cookie для access — защита от XSS (localStorage остаётся для совместимости, но cookie — primary)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", accessToken, 86400, "/", "", isSecure, true)
	c.SetCookie("refresh_token", refreshToken, 604800, "/", "", isSecure, true)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Authenticate(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserBlocked) {
			c.JSON(http.StatusForbidden, gin.H{"error": "account is blocked"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// ИСПРАВЛЕНО: Передаем user.SalonID (чтобы токен содержал актуальный салон)
	token, refresh, err := h.service.GenerateTokens(user.ID, user.Email, user.Role, user.TenantID, user.SalonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}
	setAuthCookies(c, token, refresh)

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

	// fallback на cookie если тело пустое
	if req.RefreshToken == "" {
		if cookie, err := c.Cookie("refresh_token"); err == nil {
			req.RefreshToken = cookie
		}
	}
	access, refresh, err := h.service.RefreshTokens(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	setAuthCookies(c, access, refresh)

	c.JSON(http.StatusOK, gin.H{
		"token":         access,
		"refresh_token": refresh,
	})
}

// Logout - отзывает refresh token (по jti) и чистит cookie + отзывает access jti если передан
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.RefreshToken == "" {
		if cookie, err := c.Cookie("refresh_token"); err == nil {
			req.RefreshToken = cookie
		}
	}
	// отозвать access jti из Authorization/cookie для мгновенной инвалидации
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		_ = h.service.RevokeAccessToken(authHeader)
	} else if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
		_ = h.service.RevokeAccessToken("Bearer " + cookie)
	}
	if req.RefreshToken != "" {
		_ = h.service.Logout(req.RefreshToken)
	}
	// clear cookies
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}
