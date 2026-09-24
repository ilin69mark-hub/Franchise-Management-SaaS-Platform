package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupCORSTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS())
	r.GET("/data", func(c *gin.Context) { c.Status(200) })
	return r
}

func TestCORS_AllowedOriginGetsHeaders(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com")
	r := setupCORSTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/data", nil)
	req.Header.Set("Origin", "https://app.example.com")
	r.ServeHTTP(w, req)
	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"), "доверенный origin получает служебные заголовки")
}

func TestCORS_DisallowedOriginGetsNothing(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com")
	r := setupCORSTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/data", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	r.ServeHTTP(w, req)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Headers"), "чужой origin не должен видеть preflight-подсказки")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCORS_ReleaseModeEmptyEnvDeniesAll(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS())
	r.GET("/data", func(c *gin.Context) { c.Status(200) })
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/data", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), "F11: в release без env localhost запрещён")
	assert.Equal(t, http.StatusOK, w.Code)
}
