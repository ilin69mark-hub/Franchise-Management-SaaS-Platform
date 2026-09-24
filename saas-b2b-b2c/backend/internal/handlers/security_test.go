package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"franchise-saas-backend/internal/middleware"
	"franchise-saas-backend/internal/mocks"
	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// helpers for JWT tests
func jwtSecretForTest() string {
	secret := viper.GetString("jwt_secret")
	if secret == "" {
		viper.Set("jwt_secret", "test-secret-for-security-tests-32bytes!")
		secret = viper.GetString("jwt_secret")
	}
	return secret
}

func signToken(claims jwt.MapClaims, method jwt.SigningMethod) string {
	secret := jwtSecretForTest()
	t := jwt.NewWithClaims(method, claims)
	s, _ := t.SignedString([]byte(secret))
	return s
}

func TestSecurity_JWT_NoneAlgorithmRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// alg none token
	claims := jwt.MapClaims{"user_id": "00000000-0000-0000-0000-000000000001", "exp": time.Now().Add(time.Hour).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	noneStr, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+noneStr)
	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/", func(c *gin.Context) { c.Status(200) })
	r.ServeHTTP(w, c.Request)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "none alg must be rejected")
}

func TestSecurity_JWT_ExpiredRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	claims := jwt.MapClaims{"user_id": "00000000-0000-0000-0000-000000000001", "exp": time.Now().Add(-time.Hour).Unix()}
	str := signToken(claims, jwt.SigningMethodHS256)
	w := httptest.NewRecorder()
	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/", func(c *gin.Context) { c.Status(200) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+str)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "expired must be rejected")
}

func TestSecurity_JWT_MissingJtiRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// F12: токен без jti нельзя отозвать — должен быть отвергнут.
	claims := jwt.MapClaims{"user_id": "00000000-0000-0000-0000-000000000001", "role": "dealer", "exp": time.Now().Add(time.Hour).Unix()}
	str := signToken(claims, jwt.SigningMethodHS256)
	w := httptest.NewRecorder()
	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/", func(c *gin.Context) { c.Status(200) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+str)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "missing jti must be rejected")
	assert.Contains(t, w.Body.String(), "jti")
}

func TestSecurity_JWT_ValidAccepted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	claims := jwt.MapClaims{"user_id": "00000000-0000-0000-0000-000000000001", "role": "dealer", "jti": "00000000-0000-0000-0000-000000000099", "exp": time.Now().Add(time.Hour).Unix()}
	str := signToken(claims, jwt.SigningMethodHS256)
	w := httptest.NewRecorder()
	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/", func(c *gin.Context) { c.Status(200) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+str)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecurity_AuthMiddleware_TrimSpaceBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	claims := jwt.MapClaims{"user_id": "00000000-0000-0000-0000-000000000001", "jti": "00000000-0000-0000-0000-000000000098", "exp": time.Now().Add(time.Hour).Unix()}
	str := signToken(claims, jwt.SigningMethodHS256)
	w := httptest.NewRecorder()
	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/", func(c *gin.Context) { c.Status(200) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer   "+str+"   ")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "bearer with extra spaces must be trimmed")
}

// testClientIP — уникальный IP на запуск: глобальный in-memory лимитер общий
// на процесс, фиксированный IP флапает при повторах.
var testClientIPSeq int64

func nextTestClientIP() string {
	n := atomic.AddInt64(&testClientIPSeq, 1)
	return fmt.Sprintf("10.%d.%d.1:1234", (n/250)%250+1, n%250+1)
}

func TestSecurity_RateLimit_6thRequest429(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// do not set TrustedProxies so ClientIP uses RemoteAddr
	r.Use(middleware.RateLimit(5, time.Minute))
	r.GET("/", func(c *gin.Context) { c.Status(200) })
	var lastCode int
	clientIP := nextTestClientIP()
	for i := 0; i < 6; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = clientIP
		r.ServeHTTP(w, req)
		lastCode = w.Code
	}
	assert.Equal(t, http.StatusTooManyRequests, lastCode, "6th request must be 429")
}

func TestSecurity_RateLimit_XFFIgnoredWhenTrusted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	_ = r.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12"})
	r.Use(middleware.RateLimit(5, time.Minute))
	r.GET("/", func(c *gin.Context) { c.Status(200) })
	// attacker tries XFF random
	clientIP := nextTestClientIP()
	// Ключ лимитера здесь — сам XFF (прокси доверенные): уникален на прогон.
	xff := nextTestClientIP()
	for i := 0; i < 6; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = clientIP
		req.Header.Set("X-Forwarded-For", xff)
		r.ServeHTTP(w, req)
		if i < 5 {
			require.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code, "XFF must be ignored when trusted, 6th still 429")
		}
	}
}

func TestSecurity_CookieFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	claims := jwt.MapClaims{"user_id": "00000000-0000-0000-0000-000000000001", "jti": "00000000-0000-0000-0000-000000000097", "exp": time.Now().Add(time.Hour).Unix()}
	str := signToken(claims, jwt.SigningMethodHS256)
	w := httptest.NewRecorder()
	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/", func(c *gin.Context) { c.Status(200) })
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: str})
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "cookie fallback must work")
}

func TestSecurity_JWT_RevokedRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// F1: отзыв работает и без Redis (instance-local fallback) — скип убран.
	secret := jwtSecretForTest()
	_ = secret
	svc := newTestAuthServiceForSecurity()
	access, refresh, err := svc.GenerateTokens(uuid.New(), "revoked@test.com", "dealer", nil, nil)
	require.NoError(t, err)
	require.NoError(t, svc.Logout(refresh))
	require.NoError(t, svc.RevokeAccessToken("Bearer "+access))

	r := gin.New()
	r.Use(middleware.AuthMiddleware())
	r.GET("/", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "revoked access must be rejected")

	_, _, err = svc.RefreshTokens(refresh)
	require.Error(t, err, "revoked refresh must not rotate")
}

func TestSecurity_Refresh_ReuseDenied(t *testing.T) {
	// F1: повторное использование refresh после ротации — отказ (раньше без Redis проходило).
	userID := uuid.New()
	svc := newTestAuthServiceForSecurityWithUser(userID)
	_, refresh, err := svc.GenerateTokens(userID, "reuse@test.com", "dealer", nil, nil)
	require.NoError(t, err)
	_, _, err = svc.RefreshTokens(refresh)
	require.NoError(t, err)
	_, _, err = svc.RefreshTokens(refresh)
	require.Error(t, err, "refresh reuse must be denied")
}

// --- F1/F12 helpers: real AuthService with mocked user repo (no Redis needed) ---
func newTestAuthServiceForSecurity() *services.AuthService {
	jwtSecretForTest()
	repo := mocks.NewMockUserRepository()
	return services.NewAuthServiceWithInterface(repo, nil)
}

func newTestAuthServiceForSecurityWithUser(id uuid.UUID) *services.AuthService {
	jwtSecretForTest()
	repo := mocks.NewMockUserRepository()
	repo.On("GetUserByID", mock.Anything, id).Return(&models.User{ID: id, Email: "reuse@test.com", Role: models.RoleDealer}, nil)
	return services.NewAuthServiceWithInterface(repo, nil)
}
