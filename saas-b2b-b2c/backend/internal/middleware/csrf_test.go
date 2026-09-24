package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCSRFRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CSRF())
	r.POST("/mut", func(c *gin.Context) { c.Status(200) })
	r.GET("/safe", func(c *gin.Context) { c.Status(200) })
	return r
}

func TestCSRF_MissingTokenForbidden(t *testing.T) {
	r := setupCSRFRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/mut", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "sess"})
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "CSRF")
}

func TestCSRF_MismatchForbidden(t *testing.T) {
	r := setupCSRFRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/mut", nil)
	req.AddCookie(&http.Cookie{Name: CSRFCookie, Value: "aaa"})
	req.Header.Set(CSRFHeader, "bbb")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRF_MatchAllowed(t *testing.T) {
	r := setupCSRFRouter()
	tok, err := GenerateCSRFToken()
	require.NoError(t, err)
	require.NotEmpty(t, tok)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/mut", nil)
	req.AddCookie(&http.Cookie{Name: CSRFCookie, Value: tok})
	req.Header.Set(CSRFHeader, tok)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRF_BearerExempt(t *testing.T) {
	r := setupCSRFRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/mut", nil)
	req.Header.Set("Authorization", "Bearer api-token")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRF_SafeMethodsPass(t *testing.T) {
	r := setupCSRFRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/safe", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRF_GenerateUnique(t *testing.T) {
	a, err := GenerateCSRFToken()
	require.NoError(t, err)
	b, err := GenerateCSRFToken()
	require.NoError(t, err)
	require.NotEqual(t, a, b)
	require.GreaterOrEqual(t, len(a), 32)
}
