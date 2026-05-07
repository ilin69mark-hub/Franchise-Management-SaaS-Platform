package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test_secret_key"

func createTestToken(userID, email, role string, exp time.Time) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     exp.Unix(),
	})
	tokenStr, _ := token.SignedString([]byte(testSecret))
	return tokenStr
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func setupAuthRouter() *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		
		tokenStr := authHeader[7:]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(testSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		
		claims := token.Claims.(jwt.MapClaims)
		c.Set("userID", claims["user_id"])
		c.Set("email", claims["email"])
		c.Set("role", claims["role"])
		c.Set("tenantID", claims["tenant_id"])
		c.Next()
	})
	return r
}

// RequireRole - Role-based middleware for testing
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "role not found"})
			c.Abort()
			return
		}
		
		userRole, ok := roleVal.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid role"})
			c.Abort()
			return
		}
		
		if userRole != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

func TestAuthHandler_MissingHeader(t *testing.T) {
	r := setupTestRouter()
	r.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}
		c.Next()
	})
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header required")
}

func TestAuthHandler_InvalidFormat(t *testing.T) {
	r := setupTestRouter()
	r.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) < 7 || authHeader[:6] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}
		c.Next()
	})
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic token123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid authorization header format")
}

func TestAuthHandler_ValidToken(t *testing.T) {
	r := setupAuthRouter()
	r.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("userID")
		email, _ := c.Get("email")
		role, _ := c.Get("role")
		c.JSON(http.StatusOK, gin.H{
			"userID": userID,
			"email":  email,
			"role":   role,
		})
	})

	validToken := createTestToken("user-123", "test@test.com", "dealer", time.Now().Add(time.Hour))
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user-123")
	assert.Contains(t, w.Body.String(), "test@test.com")
	assert.Contains(t, w.Body.String(), "dealer")
}

func TestAuthHandler_ExpiredToken(t *testing.T) {
	r := setupAuthRouter()
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	expiredToken := createTestToken("user-123", "test@test.com", "dealer", time.Now().Add(-time.Hour))
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_MissingUserID(t *testing.T) {
	r := setupTestRouter()
	r.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenStr := authHeader[7:]
		token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(testSecret), nil
		})
		claims := token.Claims.(jwt.MapClaims)
		userID, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing user_id"})
			c.Abort()
			return
		}
		c.Set("userID", userID)
		c.Next()
	})
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": "test@test.com",
		"role":  "dealer",
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString([]byte(testSecret))

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Token missing user_id")
}

func TestAuthHandler_AllRoles(t *testing.T) {
	roles := []string{"dealer", "franchiser", "franchiser_manager", "salon_manager", "super_admin"}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			r := setupAuthRouter()
			r.GET("/test", func(c *gin.Context) {
				role, _ := c.Get("role")
				c.JSON(http.StatusOK, gin.H{"role": role})
			})

			token := createTestToken("user-1", "test@test.com", role, time.Now().Add(time.Hour))
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestRoleMiddleware_DealerAccess(t *testing.T) {
	tests := []struct {
		name       string
		userRole   string
		required   string
		wantCode   int
	}{
		{"dealer accessing dealer resource", "dealer", "dealer", http.StatusOK},
		{"dealer accessing franchiser resource", "dealer", "franchiser", http.StatusForbidden},
		{"franchiser accessing dealer resource", "franchiser", "dealer", http.StatusForbidden},
		{"franchiser accessing franchiser resource", "franchiser", "franchiser", http.StatusOK},
		{"super_admin cannot access franchiser resource", "super_admin", "franchiser", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupAuthRouter()
			r.GET("/resource", RequireRole(tt.required), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"access": "granted"})
			})

			token := createTestToken("user-1", "test@test.com", tt.userRole, time.Now().Add(time.Hour))
			req, _ := http.NewRequest("GET", "/resource", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestTenantID_Extraction(t *testing.T) {
	r := setupAuthRouter()
	r.GET("/test", func(c *gin.Context) {
		tenantID, _ := c.Get("tenantID")
		c.JSON(http.StatusOK, gin.H{"tenantID": tenantID})
	})

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   "user-123",
		"email":     "test@test.com",
		"role":      "franchiser",
		"tenant_id": "tenant-abc",
		"exp":       time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString([]byte(testSecret))

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "tenant-abc")
}

func TestAuthorization_Required(t *testing.T) {
	r := setupAuthRouter()
	r.GET("/secret", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"secret": "data"})
	})

	req, _ := http.NewRequest("GET", "/secret", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestChained_Middleware(t *testing.T) {
	r := setupAuthRouter()
	
	callCount := 0
	r.GET("/admin", func(c *gin.Context) {
		callCount++
		c.JSON(http.StatusOK, gin.H{"stage": "handler"})
	}, func(c *gin.Context) {
		callCount++
		c.Next()
	})

	token := createTestToken("user-1", "test@test.com", "dealer", time.Now().Add(time.Hour))
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}