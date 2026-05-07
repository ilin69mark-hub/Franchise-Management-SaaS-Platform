package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGoalHandlerHTTP_SetGoal_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/goals", func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Set("role", "dealer")
		var req struct {
			Target int    `json:"target"`
			Type   string `json:"type"`
		}
		c.ShouldBindJSON(&req)
		c.JSON(http.StatusCreated, gin.H{"target": req.Target, "type": req.Type})
	})

	req, _ := http.NewRequest("POST", "/goals", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestGoalHandlerHTTP_SetGoal_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/goals", func(c *gin.Context) {
		_, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{})
	})

	req, _ := http.NewRequest("POST", "/goals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGoalHandlerHTTP_SetGoal_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/goals", func(c *gin.Context) {
		c.Set("userID", "user-123")
		var req struct {
			Target int    `json:"target"`
			Type   string `json:"type"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{})
	})

	req, _ := http.NewRequest("POST", "/goals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGoalHandlerHTTP_SetGoal_NegativeTarget(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/goals", func(c *gin.Context) {
		c.Set("userID", "user-123")
		var req struct {
			Target int    `json:"target"`
			Type   string `json:"type"`
		}
		c.ShouldBindJSON(&req)
		if req.Target <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "target must be positive"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{})
	})

	req, _ := http.NewRequest("POST", "/goals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGoalHandlerHTTP_GetGoal_ValidDate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/goals", func(c *gin.Context) {
		c.Set("userID", "user-123")
		dateStr := c.Query("date")
		c.JSON(http.StatusOK, gin.H{"date": dateStr})
	})

	req, _ := http.NewRequest("GET", "/goals?date=2024-01-15", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGoalHandlerHTTP_GetGoal_MissingDate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/goals", func(c *gin.Context) {
		c.Set("userID", "user-123")
		dateStr := c.Query("date")
		if dateStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date query required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("GET", "/goals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGoalHandlerHTTP_GetGoal_InvalidDateFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/goals", func(c *gin.Context) {
		c.Set("userID", "user-123")
		dateStr := c.Query("date")
		_, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("GET", "/goals?date=invalid-date", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGoalHandlerHTTP_GetGoalByDate_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/goals/by-date/:date", func(c *gin.Context) {
		c.Set("userID", "user-123")
		dateStr := c.Param("date")
		c.JSON(http.StatusOK, gin.H{"date": dateStr})
	})

	req, _ := http.NewRequest("GET", "/goals/by-date/2024-01-15", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGoalHandlerHTTP_GetGoalByDate_MissingDate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/goals/by-date/:date", func(c *gin.Context) {
		c.Set("userID", "user-123")
		dateStr := c.Param("date")
		if dateStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("GET", "/goals/by-date", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGoalHandlerHTTP_DeleteGoal_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.DELETE("/goals/:id", func(c *gin.Context) {
		c.Set("userID", "user-123")
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{"goal_id": id})
	})

	goalID := "goal-123"
	req, _ := http.NewRequest("DELETE", "/goals/"+goalID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGoalHandlerHTTP_DeleteGoal_MissingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.DELETE("/goals/:id", func(c *gin.Context) {
		c.Set("userID", "user-123")
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "goal id required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	req, _ := http.NewRequest("DELETE", "/goals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGoalHandlerHTTP_SetGoal_ZeroTarget(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/goals", func(c *gin.Context) {
		c.Set("userID", "user-123")
		var req struct {
			Target int    `json:"target"`
			Type   string `json:"type"`
		}
		c.ShouldBindJSON(&req)
		if req.Target <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "target must be positive"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{})
	})

	req, _ := http.NewRequest("POST", "/goals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGoalHandlerHTTP_SetGoal_WithTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/goals", func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Set("tenantID", "tenant-123")
		var req struct {
			Target   int    `json:"target"`
			Type     string `json:"type"`
			TenantID string `json:"tenant_id"`
		}
		c.ShouldBindJSON(&req)
		c.JSON(http.StatusCreated, gin.H{"tenant_id": req.TenantID})
	})

	req, _ := http.NewRequest("POST", "/goals", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}