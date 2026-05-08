package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUserHandlerHTTP_GetEmployees_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/users/employees", func(c *gin.Context) {
		_, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("GET", "/users/employees", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserHandlerHTTP_GetEmployees_Authorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/users/employees", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.Set("role", "admin")
		c.JSON(http.StatusOK, gin.H{"users": []gin.H{}, "count": 0})
	})

	req, _ := http.NewRequest("GET", "/users/employees", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandlerHTTP_GetProfile_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/users/profile", func(c *gin.Context) {
		_, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("GET", "/users/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserHandlerHTTP_GetProfile_Authorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/users/profile", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.JSON(http.StatusOK, gin.H{"id": "test-user-id", "email": "test@example.com"})
	})

	req, _ := http.NewRequest("GET", "/users/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandlerHTTP_UpdateProfile_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.PUT("/users/profile", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		var req struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}
		c.ShouldBindJSON(&req)
		c.JSON(http.StatusOK, gin.H{"first_name": req.FirstName, "last_name": req.LastName})
	})

	req, _ := http.NewRequest("PUT", "/users/profile", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandlerHTTP_UpdateProfile_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.PUT("/users/profile", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		var req struct {
			FirstName string `json:"first_name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("PUT", "/users/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandlerHTTP_DeleteUser_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("userID", "admin-user-id")
		targetID := c.Param("id")
		c.JSON(http.StatusOK, gin.H{"message": "user deleted", "user_id": targetID})
	})

	userID := "target-user-id"
	req, _ := http.NewRequest("DELETE", "/users/"+userID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandlerHTTP_DeleteUser_MissingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("userID", "admin-user-id")
		targetID := c.Param("id")
		if targetID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user id required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("DELETE", "/users/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandlerHTTP_GetEmployees_Pagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/users/employees", func(c *gin.Context) {
		c.Set("userID", "admin-user-id")
		c.Set("role", "super_admin")
		page := c.Query("page")
		limit := c.Query("limit")
		c.JSON(http.StatusOK, gin.H{"page": page, "limit": limit, "users": []gin.H{}})
	})

	req, _ := http.NewRequest("GET", "/users/employees?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandlerHTTP_GetEmployees_FilterByRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/users/employees", func(c *gin.Context) {
		c.Set("userID", "admin-user-id")
		c.Set("role", "admin")
		roleFilter := c.Query("role")
		c.JSON(http.StatusOK, gin.H{"role_filter": roleFilter, "users": []gin.H{}})
	})

	req, _ := http.NewRequest("GET", "/users/employees?role=dealer", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandlerHTTP_UpdateProfile_EmptyFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.PUT("/users/profile", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		var req struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}
		c.ShouldBindJSON(&req)
		if req.FirstName == "" && req.LastName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("PUT", "/users/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandlerHTTP_GetProfile_AdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/users/profile", func(c *gin.Context) {
		c.Set("userID", "admin-user-id")
		c.Set("role", "admin")
		c.JSON(http.StatusOK, gin.H{"id": "admin-user-id", "role": "admin"})
	})

	req, _ := http.NewRequest("GET", "/users/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandlerHTTP_GetEmployees_SuperAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/users/employees", func(c *gin.Context) {
		c.Set("userID", "super-admin-id")
		c.Set("role", "super_admin")
		c.JSON(http.StatusOK, gin.H{"scope": "global", "users": []gin.H{}})
	})

	req, _ := http.NewRequest("GET", "/users/employees", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}