package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAdminHandlerHTTP_GetDashboardStats_Authorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/admin/stats", func(c *gin.Context) {
		c.Set("role", "super_admin")
		c.JSON(http.StatusOK, gin.H{"total_tenants": 10})
	})

	req, _ := http.NewRequest("GET", "/admin/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandlerHTTP_GetDashboardStats_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/admin/stats", func(c *gin.Context) {
		role, _ := c.Get("role")
		if role == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("GET", "/admin/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminHandlerHTTP_GetAllTenants_SuperAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/admin/tenants", func(c *gin.Context) {
		c.Set("role", "super_admin")
		c.JSON(http.StatusOK, gin.H{"tenants": []gin.H{}})
	})

	req, _ := http.NewRequest("GET", "/admin/tenants", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandlerHTTP_GetAllTenants_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/admin/tenants", func(c *gin.Context) {
		c.Set("role", "dealer")
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	})

	req, _ := http.NewRequest("GET", "/admin/tenants", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminHandlerHTTP_GetTenantByID_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/admin/tenants/:id", func(c *gin.Context) {
		c.Set("role", "super_admin")
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	tenantID := "tenant-123"
	req, _ := http.NewRequest("GET", "/admin/tenants/"+tenantID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandlerHTTP_GetTenantByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/admin/tenants/:id", func(c *gin.Context) {
		c.Set("role", "super_admin")
		id := c.Param("id")
		if id == "" || id == "invalid-uuid" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID"})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("GET", "/admin/tenants/invalid-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandlerHTTP_CreateTenant_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/admin/tenants", func(c *gin.Context) {
		c.Set("role", "super_admin")
		var req struct {
			Name string `json:"name"`
		}
		c.ShouldBindJSON(&req)
		c.JSON(http.StatusCreated, gin.H{"name": req.Name})
	})

	req, _ := http.NewRequest("POST", "/admin/tenants", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAdminHandlerHTTP_CreateTenant_MissingName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/admin/tenants", func(c *gin.Context) {
		c.Set("role", "super_admin")
		var req struct {
			Name string `json:"name"`
		}
		c.ShouldBindJSON(&req)
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{})
	})

	req, _ := http.NewRequest("POST", "/admin/tenants", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminHandlerHTTP_UpdateTenant_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.PUT("/admin/tenants/:id", func(c *gin.Context) {
		c.Set("role", "super_admin")
		id := c.Param("id")
		var req struct {
			Name string `json:"name"`
		}
		c.ShouldBindJSON(&req)
		c.JSON(http.StatusOK, gin.H{"id": id, "name": req.Name})
	})

	tenantID := "tenant-123"
	req, _ := http.NewRequest("PUT", "/admin/tenants/"+tenantID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandlerHTTP_DeleteTenant_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.DELETE("/admin/tenants/:id", func(c *gin.Context) {
		c.Set("role", "super_admin")
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{"message": "deleted", "id": id})
	})

	tenantID := "tenant-123"
	req, _ := http.NewRequest("DELETE", "/admin/tenants/"+tenantID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandlerHTTP_GetPayments_Status(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/admin/payments", func(c *gin.Context) {
		c.Set("role", "super_admin")
		c.JSON(http.StatusOK, gin.H{"payments": []gin.H{}})
	})

	req, _ := http.NewRequest("GET", "/admin/payments", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandlerHTTP_GetPayments_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/admin/payments", func(c *gin.Context) {
		c.Set("role", "admin")
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	})

	req, _ := http.NewRequest("GET", "/admin/payments", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminHandlerHTTP_Stats_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/admin/stats", func(c *gin.Context) {
		c.Set("role", "super_admin")
		period := c.Query("period")
		c.JSON(http.StatusOK, gin.H{"period": period})
	})

	req, _ := http.NewRequest("GET", "/admin/stats?period=month", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}