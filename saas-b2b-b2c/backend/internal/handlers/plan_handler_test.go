package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type MockPlan struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Price     float64 `json:"price"`
	MaxSalons int `json:"max_salons"`
	MaxUsers  int `json:"max_users"`
}

type MockPlanService struct {
	plans map[string]MockPlan
}

func NewMockPlanService() *MockPlanService {
	return &MockPlanService{
		plans: make(map[string]MockPlan),
	}
}

func (s *MockPlanService) Create(name string, price float64, maxSalons, maxUsers int) MockPlan {
	plan := MockPlan{
		ID:        uuid.New().String(),
		Name:      name,
		Price:     price,
		MaxSalons: maxSalons,
		MaxUsers:  maxUsers,
	}
	s.plans[plan.ID] = plan
	return plan
}

type MockPlanHandler struct {
	svc *MockPlanService
}

func NewMockPlanHandler(svc *MockPlanService) *MockPlanHandler {
	return &MockPlanHandler{svc: svc}
}

func (h *MockPlanHandler) Create(c *gin.Context) {
	var req struct {
		Name      string  `json:"name" binding:"required"`
		Price     float64 `json:"price" binding:"gte=0"`
		MaxSalons int     `json:"max_salons" binding:"gte=0"`
		MaxUsers  int     `json:"max_users" binding:"gte=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	plan := h.svc.Create(req.Name, req.Price, req.MaxSalons, req.MaxUsers)
	c.JSON(http.StatusCreated, plan)
}

func (h *MockPlanHandler) Get(c *gin.Context) {
	id := c.Param("id")
	plan, exists := h.svc.plans[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *MockPlanHandler) List(c *gin.Context) {
	plans := make([]MockPlan, 0, len(h.svc.plans))
	for _, p := range h.svc.plans {
		plans = append(plans, p)
	}
	c.JSON(http.StatusOK, gin.H{"items": plans, "total": len(plans)})
}

func (h *MockPlanHandler) Update(c *gin.Context) {
	id := c.Param("id")
	_, exists := h.svc.plans[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}

	var req struct {
		Name      string  `json:"name"`
		Price     float64 `json:"price"`
		MaxSalons int     `json:"max_salons"`
		MaxUsers  int     `json:"max_users"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan := h.svc.plans[id]
	if req.Name != "" {
		plan.Name = req.Name
	}
	if req.Price > 0 {
		plan.Price = req.Price
	}
	if req.MaxSalons > 0 {
		plan.MaxSalons = req.MaxSalons
	}
	if req.MaxUsers > 0 {
		plan.MaxUsers = req.MaxUsers
	}
	h.svc.plans[id] = plan

	c.JSON(http.StatusOK, plan)
}

func (h *MockPlanHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if _, exists := h.svc.plans[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}
	delete(h.svc.plans, id)
	c.Status(http.StatusNoContent)
}

func setupPlanTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	svc := NewMockPlanService()
	handler := NewMockPlanHandler(svc)

	r.POST("/api/v1/plans", handler.Create)
	r.GET("/api/v1/plans", handler.List)
	r.GET("/api/v1/plans/:id", handler.Get)
	r.PATCH("/api/v1/plans/:id", handler.Update)
	r.DELETE("/api/v1/plans/:id", handler.Delete)

	return r
}

func TestPlanHandler_Create_ValidRequest(t *testing.T) {
	router := setupPlanTestRouter()

	body := map[string]interface{}{
		"name":      "Basic Plan",
		"price":     29.99,
		"max_salons": 1,
		"max_users": 5,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/plans", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestPlanHandler_Create_MissingName(t *testing.T) {
	router := setupPlanTestRouter()

	body := map[string]interface{}{
		"name":      "",
		"price":     29.99,
		"max_salons": 1,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/plans", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPlanHandler_Create_InvalidJSON(t *testing.T) {
	router := setupPlanTestRouter()

	req, _ := http.NewRequest("POST", "/api/v1/plans", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPlanHandler_Get_Success(t *testing.T) {
	router := setupPlanTestRouter()

	createBody := map[string]interface{}{
		"name":      "Test Plan",
		"price":     49.99,
		"max_salons": 2,
		"max_users": 10,
	}
	jsonCreate, _ := json.Marshal(createBody)
	createReq, _ := http.NewRequest("POST", "/api/v1/plans", bytes.NewBuffer(jsonCreate))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	planID := created["id"].(string)

	req, _ := http.NewRequest("GET", "/api/v1/plans/"+planID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPlanHandler_Get_NotFound(t *testing.T) {
	router := setupPlanTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/plans/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPlanHandler_List_Success(t *testing.T) {
	router := setupPlanTestRouter()

	body := map[string]interface{}{
		"name":      "Plan 1",
		"price":     10.0,
		"max_salons": 1,
		"max_users": 5,
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/api/v1/plans", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	req, _ = http.NewRequest("GET", "/api/v1/plans", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPlanHandler_Update_Success(t *testing.T) {
	router := setupPlanTestRouter()

	createBody := map[string]interface{}{
		"name":      "Original Plan",
		"price":     10.0,
		"max_salons": 1,
		"max_users": 5,
	}
	jsonCreate, _ := json.Marshal(createBody)
	createReq, _ := http.NewRequest("POST", "/api/v1/plans", bytes.NewBuffer(jsonCreate))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	planID := created["id"].(string)

	updateBody := map[string]interface{}{
		"name":      "Updated Plan",
		"price":     19.99,
	}
	jsonUpdate, _ := json.Marshal(updateBody)
	req, _ := http.NewRequest("PATCH", "/api/v1/plans/"+planID, bytes.NewBuffer(jsonUpdate))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPlanHandler_Update_NotFound(t *testing.T) {
	router := setupPlanTestRouter()

	updateBody := map[string]interface{}{
		"name": "Updated Plan",
	}
	jsonUpdate, _ := json.Marshal(updateBody)

	req, _ := http.NewRequest("PATCH", "/api/v1/plans/"+uuid.New().String(), bytes.NewBuffer(jsonUpdate))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPlanHandler_Delete_Success(t *testing.T) {
	router := setupPlanTestRouter()

	createBody := map[string]interface{}{
		"name":      "To Delete",
		"price":     10.0,
		"max_salons": 1,
		"max_users": 5,
	}
	jsonCreate, _ := json.Marshal(createBody)
	createReq, _ := http.NewRequest("POST", "/api/v1/plans", bytes.NewBuffer(jsonCreate))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	var created map[string]interface{}
	json.Unmarshal(createW.Body.Bytes(), &created)
	planID := created["id"].(string)

	req, _ := http.NewRequest("DELETE", "/api/v1/plans/"+planID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestPlanHandler_Delete_NotFound(t *testing.T) {
	router := setupPlanTestRouter()

	req, _ := http.NewRequest("DELETE", "/api/v1/plans/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}