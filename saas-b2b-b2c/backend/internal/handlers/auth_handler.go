package handlers

import (
	"errors"
	"log"
	"net/http"

	"franchise-saas-backend/internal/middleware"
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
			// F6: дубль email — 409 без текста SQL (enumeration/разведка схемы);
			// остальное — generic 500, детали в лог.
			if errors.Is(err, services.ErrEmailTaken) {
				c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
				return
			}
			log.Printf("Register failed for email %q: %v", req.Email, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	} else {
		_, err := h.service.CreateUser(user, req.Password)
		if err != nil {
			if errors.Is(err, services.ErrEmailTaken) {
				c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
				return
			}
			log.Printf("Register failed for email %q: %v", req.Email, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
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

	// F7: токенов в теле больше нет — только httpOnly cookie (+ читаемый csrf_token).
	// XSS больше не может украсть сессию из localStorage/тела ответа.
	c.JSON(http.StatusCreated, gin.H{"user": *user})
}

// setAuthCookies — сессия только в cookie: access/refresh httpOnly,
// csrf_token читаемый (double-submit для мутаций, см. middleware.CSRF).
func setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", accessToken, 86400, "/", "", isSecure, true)
	c.SetCookie("refresh_token", refreshToken, 604800, "/", "", isSecure, true)
	if csrf, err := middleware.GenerateCSRFToken(); err == nil {
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(middleware.CSRFCookie, csrf, 604800, "/", "", isSecure, false)
	} else {
		log.Printf("setAuthCookies: CSRF generation failed: %v", err)
	}
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
		// F9: lockout — 429 c Retry-After, без различия "нет юзера/неверный пароль".
		if errors.Is(err, services.ErrTooManyAttempts) {
			c.Header("Retry-After", "900")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many login attempts, try again later"})
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

	// F7: без токенов в теле (см. Register).
	c.JSON(http.StatusOK, gin.H{"user": *user})
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

	// F7: без токенов в теле — сессия только в cookie.
	c.JSON(http.StatusOK, gin.H{"message": "refreshed"})
}

// GetCSRF - bootstrap читаемой csrf_token cookie для cookie-сессий
// (например, сессия жива, а csrf cookie потеряна). Требует авторизации.
func (h *AuthHandler) GetCSRF(c *gin.Context) {
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	csrf, err := middleware.GenerateCSRFToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate CSRF token"})
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.CSRFCookie, csrf, 604800, "/", "", isSecure, false)
	c.JSON(http.StatusOK, gin.H{"message": "csrf refreshed"})
}

// Logout - отзывает refresh token (по jti) и чистит cookie + отзывает access jti если передан.
// F1: ошибки отзыва больше не глотаются — клиент должен знать,
// если серверная инвалидация не удалась (иначе logout-театр при падении Redis).
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
	var revokeErrs []string
	// отозвать access jti из Authorization/cookie для мгновенной инвалидации
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if err := h.service.RevokeAccessToken(authHeader); err != nil {
			log.Printf("Logout: RevokeAccessToken failed: %v", err)
			revokeErrs = append(revokeErrs, "access revocation failed")
		}
	} else if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
		if err := h.service.RevokeAccessToken("Bearer " + cookie); err != nil {
			log.Printf("Logout: RevokeAccessToken failed: %v", err)
			revokeErrs = append(revokeErrs, "access revocation failed")
		}
	}
	if req.RefreshToken != "" {
		if err := h.service.Logout(req.RefreshToken); err != nil {
			log.Printf("Logout: refresh revocation failed: %v", err)
			revokeErrs = append(revokeErrs, "refresh revocation failed")
		}
	}
	// clear cookies всегда — локальный выход гарантирован.
	// F7: Secure обязан совпадать с установленным (иначе https-cookie не сотрётся).
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", "", -1, "/", "", isSecure, true)
	c.SetCookie("refresh_token", "", -1, "/", "", isSecure, true)
	c.SetCookie(middleware.CSRFCookie, "", -1, "/", "", isSecure, false)

	if len(revokeErrs) > 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "Logged out locally",
			"warning": "server-side revocation failed, session may still be usable until expiry",
			"errors":  revokeErrs,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}
