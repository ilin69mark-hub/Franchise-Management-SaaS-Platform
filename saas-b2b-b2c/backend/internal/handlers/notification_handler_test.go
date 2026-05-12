package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type MockNotification struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	Message   string `json:"message"`
	IsRead    bool   `json:"is_read"`
}

type MockNotificationService struct {
	notifications map[string][]MockNotification
}

func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{
		notifications: make(map[string][]MockNotification),
	}
}

type MockNotificationHandler struct {
	svc *MockNotificationService
}

func NewMockNotificationHandler(svc *MockNotificationService) *MockNotificationHandler {
	return &MockNotificationHandler{svc: svc}
}

func (h *MockNotificationHandler) GetMyNotifications(c *gin.Context) {
	val, exists := c.Get("tenantID")
	if !exists || val == "" {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	tenantIDStr, ok := val.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid tenant ID format"})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
		return
	}

	notifications := h.svc.notifications[tenantID.String()]
	c.JSON(http.StatusOK, notifications)
}

func (h *MockNotificationHandler) MarkAsRead(c *gin.Context) {
	id := c.Param("id")
	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	for tenantID, notes := range h.svc.notifications {
		for i, n := range notes {
			if n.ID == uid.String() {
				notes[i].IsRead = true
				h.svc.notifications[tenantID] = notes
				c.JSON(http.StatusOK, gin.H{"message": "Marked as read"})
				return
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
}

func (h *MockNotificationHandler) MarkAllAsRead(c *gin.Context) {
	val, exists := c.Get("tenantID")
	if !exists || val == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID required"})
		return
	}

	tenantIDStr, ok := val.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid tenant ID"})
		return
	}

	tenantID, _ := uuid.Parse(tenantIDStr)

	if notes, exists := h.svc.notifications[tenantID.String()]; exists {
		for i := range notes {
			notes[i].IsRead = true
		}
		h.svc.notifications[tenantID.String()] = notes
	}

	c.JSON(http.StatusOK, gin.H{"message": "All marked as read"})
}

func setupNotificationTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	svc := NewMockNotificationService()
	handler := NewMockNotificationHandler(svc)

	r.GET("/api/v1/notifications", func(c *gin.Context) {
		c.Set("tenantID", "550e8400-e29b-41d4-a716-446655440000")
		handler.GetMyNotifications(c)
	})
	r.POST("/api/v1/notifications/:id/read", func(c *gin.Context) {
		handler.MarkAsRead(c)
	})
	r.POST("/api/v1/notifications/read-all", func(c *gin.Context) {
		c.Set("tenantID", "test-tenant-id")
		handler.MarkAllAsRead(c)
	})

	return r
}

func TestNotificationHandler_GetMyNotifications_NoTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	svc := NewMockNotificationService()
	handler := NewMockNotificationHandler(svc)
	r.GET("/api/v1/notifications", handler.GetMyNotifications)

	req, _ := http.NewRequest("GET", "/api/v1/notifications", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "[]")
}

func TestNotificationHandler_GetMyNotifications_InvalidTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	svc := NewMockNotificationService()
	handler := NewMockNotificationHandler(svc)
	r.GET("/api/v1/notifications", func(c *gin.Context) {
		c.Set("tenantID", "invalid-uuid")
		handler.GetMyNotifications(c)
	})

	req, _ := http.NewRequest("GET", "/api/v1/notifications", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNotificationHandler_GetMyNotifications_Success(t *testing.T) {
	router := setupNotificationTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/notifications", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotificationHandler_MarkAsRead_InvalidID(t *testing.T) {
	router := setupNotificationTestRouter()

	req, _ := http.NewRequest("POST", "/api/v1/notifications/invalid-id/read", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNotificationHandler_MarkAsRead_NotFound(t *testing.T) {
	router := setupNotificationTestRouter()

	req, _ := http.NewRequest("POST", "/api/v1/notifications/"+uuid.New().String()+"/read", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestNotificationHandler_MarkAsRead_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	svc := NewMockNotificationService()
	svc.notifications["test-tenant-id"] = []MockNotification{
		{ID: uuid.New().String(), TenantID: "test-tenant-id", Message: "Test", IsRead: false},
	}
	handler := NewMockNotificationHandler(svc)

	r.POST("/notifications/:id/read", handler.MarkAsRead)

	noteID := svc.notifications["test-tenant-id"][0].ID
	req, _ := http.NewRequest("POST", "/notifications/"+noteID+"/read", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotificationHandler_MarkAllAsRead_NoTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	svc := NewMockNotificationService()
	handler := NewMockNotificationHandler(svc)
	r.POST("/notifications/read-all", handler.MarkAllAsRead)

	req, _ := http.NewRequest("POST", "/notifications/read-all", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNotificationHandler_MarkAllAsRead_Success(t *testing.T) {
	router := setupNotificationTestRouter()

	req, _ := http.NewRequest("POST", "/api/v1/notifications/read-all", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}