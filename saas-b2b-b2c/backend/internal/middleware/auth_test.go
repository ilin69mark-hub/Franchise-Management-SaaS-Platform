package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthMiddleware_BearerToken_ParsesFullToken проверяет, что "Bearer <token>"
// корректно разбирается и токен не обрезается (первый символ не теряется).
func TestAuthMiddleware_BearerToken_ParsesFullToken(t *testing.T) {
	viper.Set("jwt_secret", testSecret)
	defer viper.Set("jwt_secret", "")

	userID := "11111111-1111-1111-1111-111111111111"
	token := createTestToken(userID, "bearer@test.com", "dealer", time.Now().Add(time.Hour))

	r := setupTestRouter()
	r.Use(AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"userID": c.GetString("userID"),
			"email":  c.GetString("email"),
			"role":   c.GetString("role"),
		})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), userID)
	assert.Contains(t, w.Body.String(), "bearer@test.com")
}

// TestAuthMiddleware_BearerToken_CaseInsensitive принимает схему в нижнем регистре.
func TestAuthMiddleware_BearerToken_CaseInsensitive(t *testing.T) {
	viper.Set("jwt_secret", testSecret)
	defer viper.Set("jwt_secret", "")

	userID := "22222222-2222-2222-2222-222222222222"
	token := createTestToken(userID, "lower@test.com", "dealer", time.Now().Add(time.Hour))

	r := setupTestRouter()
	r.Use(AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"userID": c.GetString("userID")})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), userID)
}

// TestAuthMiddleware_InvalidFormat отвергает неподходящую схему.
func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	viper.Set("jwt_secret", testSecret)
	defer viper.Set("jwt_secret", "")

	r := setupTestRouter()
	r.Use(AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRequireRole_AllowsAndRejects проверяет role-guard middleware.
func TestRequireRole_AllowsAndRejects(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// REAUDIT-3: роль без tenant = 403 (fail-closed), поэтому в тестах tenant
	// проставляется явно; кейсы без tenant вынесены отдельно.
	newRouter := func(role string, allowed ...string) *gin.Engine {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			if role != "" {
				c.Set("role", role)
				if role != "super_admin" {
					c.Set("tenantID", "11111111-1111-1111-1111-111111111111")
				}
			}
			c.Next()
		})
		r.GET("/resource", RequireRole(allowed...), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"access": "granted"})
		})
		return r
	}

	tests := []struct {
		name     string
		role     string
		allowed  []string
		wantCode int
	}{
		{"allows matching role", "dealer", []string{"dealer", "salon_manager"}, http.StatusOK},
		{"allows second role", "salon_manager", []string{"dealer", "salon_manager"}, http.StatusOK},
		{"rejects wrong role", "dealer", []string{"franchiser", "super_admin"}, http.StatusForbidden},
		{"rejects super_admin when not listed", "super_admin", []string{"franchiser"}, http.StatusForbidden},
		{"rejects empty role", "", []string{"dealer"}, http.StatusForbidden},
		{"allows super_admin without tenant", "super_admin", []string{"super_admin"}, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRouter(tt.role, tt.allowed...)
			req, _ := http.NewRequest("GET", "/resource", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantCode == http.StatusForbidden {
				assert.Contains(t, w.Body.String(), "forbidden")
			}
		})
	}
}

// TestRequireRole_NilTenantForbidden — REAUDIT-3: роль без tenant = 403.
func TestRequireRole_NilTenantForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", "dealer")
		c.Next()
	})
	r.GET("/resource", RequireRole("dealer"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"access": "granted"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/resource", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code, "tenantless caller must be rejected")
}
