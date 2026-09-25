package handlers

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

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

// allowedRegisterRoles - роли, доступные на публичной саморегистрации.
// ТОЛЬКО franchiser: он создаёт себе tenant. Менеджеры/дилеры/управляющие
// салоном — это внутренние роли: они создаются внутри tenant через POST /users.
// Раньше любой мог зарегистрироваться как franchiser_manager/dealer/salon_manager
// и получить аккаунт с tenant_id = NULL, что давало fail-open доступ (REAUDIT-3).
var allowedRegisterRoles = map[models.Role]bool{
	models.RoleFranchisor: true,
}

// genericRegisterMessage — одинаковый ответ при любом исходе регистрации,
// чтобы по нему нельзя было перечислить существующие email'ы.
const genericRegisterMessage = "If this email is available, the account has been created"

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
			// REAUDIT-3: ответ одинаков для "занят" и прочих ошибок регистрации,
			// иначе 409/500/201 позволяют перечислить email'ы.
			if errors.Is(err, services.ErrEmailTaken) {
				c.JSON(http.StatusAccepted, gin.H{"message": genericRegisterMessage})
				return
			}
			log.Printf("Register failed for email %q: %v", req.Email, err)
			c.JSON(http.StatusAccepted, gin.H{"message": genericRegisterMessage})
			return
		}
	} else {
		_, err := h.service.CreateUser(user, req.Password)
		if err != nil {
			if errors.Is(err, services.ErrEmailTaken) {
				c.JSON(http.StatusAccepted, gin.H{"message": genericRegisterMessage})
				return
			}
			log.Printf("Register failed for email %q: %v", req.Email, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	}

	// ИСПРАВЛЕНО: Передаем user.SalonID
	// REAUDIT-4: успех НЕ выдаёт cookie-сессию. Раньше успех отличался от
	// дублика тремя заголовками Set-Cookie — этого было достаточно, чтобы
	// перечислить email'ы (доказано в аудите). Одинаковый 202 + текст для обоих
	// исходов, дальше пользователь идёт на /login.
	c.JSON(http.StatusAccepted, gin.H{"message": genericRegisterMessage})
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

func clearAuthCookies(c *gin.Context) {
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", "", -1, "/", "", isSecure, true)
	c.SetCookie("refresh_token", "", -1, "/", "", isSecure, true)
	c.SetCookie(middleware.CSRFCookie, "", -1, "/", "", isSecure, false)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Authenticate(c.Request.Context(), req.Email, req.Password, c.ClientIP(), req.CaptchaToken)
	if err != nil {
		// S1: после порога неудач нужен валидный captcha_token.
		if errors.Is(err, services.ErrCaptchaRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "captcha_required"})
			return
		}
		if errors.Is(err, services.ErrUserBlocked) {
			c.JSON(http.StatusForbidden, gin.H{"error": "account is blocked"})
			return
		}
		if errors.Is(err, services.ErrTenantBlocked) {
			c.JSON(http.StatusForbidden, gin.H{"error": "tenant is blocked"})
			return
		}
		// REAUDIT-3: истёкшая подписка = 402 Payment Required.
		if errors.Is(err, services.ErrTenantSubscriptionExpired) {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "subscription expired"})
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

	token, refresh, err := h.service.IssueSession(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
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
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.RefreshToken == "" {
		if cookie, err := c.Cookie("refresh_token"); err == nil {
			req.RefreshToken = cookie
		}
	}
	access, refresh, err := h.service.RefreshTokensWithContext(c.Request.Context(), req.RefreshToken)
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

func (h *AuthHandler) Logout(c *gin.Context) {
	clearAuthCookies(c)

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		if cookie, err := c.Cookie("refresh_token"); err == nil {
			refreshToken = strings.TrimSpace(cookie)
		}
	}
	accessToken := ""
	if header := strings.TrimSpace(c.GetHeader("Authorization")); strings.HasPrefix(strings.ToUpper(header), "BEARER ") {
		accessToken = strings.TrimSpace(header[7:])
	}
	if accessToken == "" {
		if cookie, err := c.Cookie("access_token"); err == nil {
			accessToken = strings.TrimSpace(cookie)
		}
	}

	revoke := func(raw string) (bool, error) {
		if raw == "" {
			return false, nil
		}
		if h.service == nil {
			return true, errors.New("auth service is unavailable")
		}
		return h.service.LogoutToken(c.Request.Context(), raw)
	}
	recognized := false
	if refreshToken != "" {
		var err error
		recognized, err = revoke(refreshToken)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "Logged out locally",
				"warning": "server-side session revocation failed",
			})
			return
		}
	}
	if !recognized && accessToken != "" {
		if _, err := revoke(accessToken); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "Logged out locally",
				"warning": "server-side session revocation failed",
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}
