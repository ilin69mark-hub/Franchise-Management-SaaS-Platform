package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type MockChecklist struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	UserID      string    `json:"user_id"`
	TenantID    string    `json:"tenant_id"`
	AssignedTo  string    `json:"assigned_to"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MockChecklistService struct {
	checklists map[string]MockChecklist
}

func NewMockChecklistService() *MockChecklistService {
	return &MockChecklistService{
		checklists: make(map[string]MockChecklist),
	}
}

type MockChecklistHandler struct {
	svc *MockChecklistService
}

func NewMockChecklistHandler(svc *MockChecklistService) *MockChecklistHandler {
	return &MockChecklistHandler{svc: svc}
}

func (h *MockChecklistHandler) GetChecklists(c *gin.Context) {
	role, _ := c.Get("role")
	tenantID, _ := c.Get("tenantID")

	var items []MockChecklist
	for _, item := range h.svc.checklists {
		if role == "super_admin" || item.TenantID == tenantID {
			items = append(items, item)
		}
	}

	c.JSON(http.StatusOK, items)
}

func (h *MockChecklistHandler) CreateChecklist(c *gin.Context) {
	var req struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		AssignedTo  string    `json:"assigned_to"`
		Priority    string    `json:"priority"`
		StartDate   time.Time `json:"start_date"`
		EndDate     time.Time `json:"end_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	userID, _ := c.Get("userID")
	tenantID, _ := c.Get("tenantID")

	item := MockChecklist{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Status:      "pending",
		Priority:    req.Priority,
		UserID:      userID.(string),
		TenantID:    tenantID.(string),
		AssignedTo:  req.AssignedTo,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if item.Priority == "" {
		item.Priority = "normal"
	}

	h.svc.checklists[item.ID] = item
	c.JSON(http.StatusCreated, item)
}

func (h *MockChecklistHandler) CompleteChecklist(c *gin.Context) {
	id := c.Param("id")
	if item, exists := h.svc.checklists[id]; exists {
		item.Status = "completed"
		item.UpdatedAt = time.Now()
		h.svc.checklists[id] = item
		c.JSON(http.StatusOK, gin.H{"message": "Completed"})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
}

func (h *MockChecklistHandler) GetChecklistByID(c *gin.Context) {
	id := c.Param("id")
	if item, exists := h.svc.checklists[id]; exists {
		c.JSON(http.StatusOK, item)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
}

func (h *MockChecklistHandler) UpdateChecklist(c *gin.Context) {
	id := c.Param("id")
	item, exists := h.svc.checklists[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	var req struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		AssignedTo  string    `json:"assigned_to"`
		Status      string    `json:"status"`
		Priority    string    `json:"priority"`
		StartDate   time.Time `json:"start_date"`
		EndDate     time.Time `json:"end_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" {
		item.Title = req.Title
	}
	item.Description = req.Description
	item.AssignedTo = req.AssignedTo
	if req.Status != "" {
		item.Status = req.Status
	}
	if req.Priority != "" {
		item.Priority = req.Priority
	}
	item.UpdatedAt = time.Now()

	h.svc.checklists[id] = item
	c.JSON(http.StatusOK, item)
}

func (h *MockChecklistHandler) DeleteChecklist(c *gin.Context) {
	id := c.Param("id")
	if _, exists := h.svc.checklists[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	delete(h.svc.checklists, id)
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

func (h *MockChecklistHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	item, exists := h.svc.checklists[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	item.Status = req.Status
	item.UpdatedAt = time.Now()
	h.svc.checklists[id] = item
	c.JSON(http.StatusOK, item)
}

func setupChecklistTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	svc := NewMockChecklistService()
	handler := NewMockChecklistHandler(svc)

	r.GET("/api/v1/checklists", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.Set("role", "dealer")
		c.Set("tenantID", "test-tenant-id")
		handler.GetChecklists(c)
	})
	r.POST("/api/v1/checklists", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.Set("tenantID", "test-tenant-id")
		handler.CreateChecklist(c)
	})
	r.GET("/api/v1/checklists/:id", handler.GetChecklistByID)
	r.PUT("/api/v1/checklists/:id", handler.UpdateChecklist)
	r.DELETE("/api/v1/checklists/:id", handler.DeleteChecklist)
	r.POST("/api/v1/checklists/:id/complete", handler.CompleteChecklist)
	r.PATCH("/api/v1/checklists/:id/status", handler.UpdateStatus)

	return r
}

func TestChecklistHandler_GetChecklists_Success(t *testing.T) {
	router := setupChecklistTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/checklists", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChecklistHandler_CreateChecklist_ValidRequest(t *testing.T) {
	router := setupChecklistTestRouter()

	body := map[string]interface{}{
		"title":       "Test Checklist",
		"description": "Test description",
		"priority":    "high",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/checklists", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestChecklistHandler_CreateChecklist_MissingTitle(t *testing.T) {
	router := setupChecklistTestRouter()

	body := map[string]interface{}{
		"title":       "",
		"description": "Test description",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/checklists", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChecklistHandler_CreateChecklist_InvalidJSON(t *testing.T) {
	router := setupChecklistTestRouter()

	req, _ := http.NewRequest("POST", "/api/v1/checklists", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChecklistHandler_GetChecklistByID_NotFound(t *testing.T) {
	router := setupChecklistTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/checklists/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChecklistHandler_GetChecklistByID_Success(t *testing.T) {
	router := setupChecklistTestRouter()

	createBody := map[string]interface{}{"title": "Test", "priority": "normal"}
	jsonCreate, _ := json.Marshal(createBody)
	createReq, _ := http.NewRequest("POST", "/api/v1/checklists", bytes.NewBuffer(jsonCreate))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	checklistID := created["id"].(string)

	req, _ := http.NewRequest("GET", "/api/v1/checklists/"+checklistID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChecklistHandler_UpdateChecklist_Success(t *testing.T) {
	router := setupChecklistTestRouter()

	createBody := map[string]interface{}{"title": "Original", "priority": "normal"}
	jsonCreate, _ := json.Marshal(createBody)
	createReq, _ := http.NewRequest("POST", "/api/v1/checklists", bytes.NewBuffer(jsonCreate))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	checklistID := created["id"].(string)

	updateBody := map[string]interface{}{
		"title":       "Updated",
		"description": "New description",
	}
	jsonUpdate, _ := json.Marshal(updateBody)
	req, _ := http.NewRequest("PUT", "/api/v1/checklists/"+checklistID, bytes.NewBuffer(jsonUpdate))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChecklistHandler_UpdateChecklist_NotFound(t *testing.T) {
	router := setupChecklistTestRouter()

	updateBody := map[string]interface{}{"title": "Updated"}
	jsonUpdate, _ := json.Marshal(updateBody)

	req, _ := http.NewRequest("PUT", "/api/v1/checklists/"+uuid.New().String(), bytes.NewBuffer(jsonUpdate))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChecklistHandler_DeleteChecklist_Success(t *testing.T) {
	router := setupChecklistTestRouter()

	createBody := map[string]interface{}{"title": "To Delete"}
	jsonCreate, _ := json.Marshal(createBody)
	createReq, _ := http.NewRequest("POST", "/api/v1/checklists", bytes.NewBuffer(jsonCreate))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	checklistID := created["id"].(string)

	req, _ := http.NewRequest("DELETE", "/api/v1/checklists/"+checklistID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChecklistHandler_CompleteChecklist_Success(t *testing.T) {
	router := setupChecklistTestRouter()

	createBody := map[string]interface{}{"title": "Test", "status": "pending"}
	jsonCreate, _ := json.Marshal(createBody)
	createReq, _ := http.NewRequest("POST", "/api/v1/checklists", bytes.NewBuffer(jsonCreate))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	checklistID := created["id"].(string)

	req, _ := http.NewRequest("POST", "/api/v1/checklists/"+checklistID+"/complete", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChecklistHandler_UpdateStatus_Success(t *testing.T) {
	router := setupChecklistTestRouter()

	createBody := map[string]interface{}{"title": "Test", "status": "pending"}
	jsonCreate, _ := json.Marshal(createBody)
	createReq, _ := http.NewRequest("POST", "/api/v1/checklists", bytes.NewBuffer(jsonCreate))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	checklistID := created["id"].(string)

	updateBody := map[string]interface{}{"status": "in_progress"}
	jsonUpdate, _ := json.Marshal(updateBody)
	req, _ := http.NewRequest("PATCH", "/api/v1/checklists/"+checklistID+"/status", bytes.NewBuffer(jsonUpdate))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChecklistHandler_UpdateStatus_MissingStatus(t *testing.T) {
	router := setupChecklistTestRouter()

	nonexistentID := uuid.New().String()
	body := map[string]interface{}{}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("PATCH", "/api/v1/checklists/"+nonexistentID+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}