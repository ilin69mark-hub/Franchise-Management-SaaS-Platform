package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"franchise-saas-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
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

func TestSecurity_JWT_ValidAccepted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	claims := jwt.MapClaims{"user_id": "00000000-0000-0000-0000-000000000001", "role": "dealer", "exp": time.Now().Add(time.Hour).Unix()}
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
	claims := jwt.MapClaims{"user_id": "00000000-0000-0000-0000-000000000001", "exp": time.Now().Add(time.Hour).Unix()}
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

func TestSecurity_RateLimit_6thRequest429(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// do not set TrustedProxies so ClientIP uses RemoteAddr
	r.Use(middleware.RateLimit(5, time.Minute))
	r.GET("/", func(c *gin.Context) { c.Status(200) })
	var lastCode int
	for i := 0; i < 6; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
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
	for i := 0; i < 6; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		req.Header.Set("X-Forwarded-For", "9.9.9.9")
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
	claims := jwt.MapClaims{"user_id": "00000000-0000-0000-0000-000000000001", "exp": time.Now().Add(time.Hour).Unix()}
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
	// set revoked via cache mock — if redis not available IsTokenRevoked returns false, so we test via refresh flow indirectly
	// This test asserts that token with jti can be revoked via service and then rejected (uses in-memory cache not available, so skip if no redis)
	t.Skip("requires redis — manual check: RevokeRefreshToken + IsTokenRevoked")
}
