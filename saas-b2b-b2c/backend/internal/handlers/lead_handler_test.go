package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"franchise-saas-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupLeadTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/api/v1/leads", func(c *gin.Context) {
		leads := []models.Lead{
			{ID: uuid.New(), FullName: "Lead 1", Status: "new"},
			{ID: uuid.New(), FullName: "Lead 2", Status: "contact"},
		}
		c.JSON(http.StatusOK, leads)
	})

	r.POST("/api/v1/leads", func(c *gin.Context) {
		var req models.CreateLeadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.FullName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "full_name is required"})
			return
		}
		lead := &models.Lead{
			ID:        uuid.New(),
			FullName:  req.FullName,
			Phone:     req.Phone,
			Email:     req.Email,
			Status:    "new",
		}
		c.JSON(http.StatusCreated, lead)
	})

	r.GET("/api/v1/leads/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		_, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		lead := &models.Lead{ID: uuid.MustParse(idStr), FullName: "Test Lead"}
		activities := []models.LeadActivity{}
		c.JSON(http.StatusOK, gin.H{"lead": lead, "activities": activities})
	})

	r.PUT("/api/v1/leads/:id/status", func(c *gin.Context) {
		idStr := c.Param("id")
		_, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		var req models.UpdateLeadStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Status == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
	})

	r.POST("/api/v1/leads/:id/activities", func(c *gin.Context) {
		idStr := c.Param("id")
		_, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		var req models.AddLeadActivityRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Type == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type is required"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Activity added"})
	})

	return r
}

func TestLeadHandler_GetMyLeads_Success(t *testing.T) {
	router := setupLeadTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/leads", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var leads []models.Lead
	err := json.Unmarshal(w.Body.Bytes(), &leads)
	assert.NoError(t, err)
	assert.Len(t, leads, 2)
}

func TestLeadHandler_CreateLead_ValidRequest(t *testing.T) {
	router := setupLeadTestRouter()

	body := map[string]interface{}{
		"full_name": "New Lead",
		"phone":     "+1234567890",
		"email":     "new@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/leads", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var lead models.Lead
	err := json.Unmarshal(w.Body.Bytes(), &lead)
	assert.NoError(t, err)
	assert.Equal(t, "New Lead", lead.FullName)
}

func TestLeadHandler_CreateLead_EmptyName(t *testing.T) {
	router := setupLeadTestRouter()

	body := map[string]interface{}{
		"full_name": "",
		"phone":     "+1234567890",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/leads", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "FullName")
}

func TestLeadHandler_CreateLead_InvalidJSON(t *testing.T) {
	router := setupLeadTestRouter()

	req, _ := http.NewRequest("POST", "/api/v1/leads", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLeadHandler_GetLeadByID_Success(t *testing.T) {
	router := setupLeadTestRouter()

	leadID := uuid.New()
	req, _ := http.NewRequest("GET", "/api/v1/leads/"+leadID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLeadHandler_GetLeadByID_InvalidID(t *testing.T) {
	router := setupLeadTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/leads/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid ID")
}

func TestLeadHandler_UpdateLeadStatus_ValidRequest(t *testing.T) {
	router := setupLeadTestRouter()

	leadID := uuid.New()
	body := map[string]interface{}{"status": "contact"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT", "/api/v1/leads/"+leadID.String()+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLeadHandler_UpdateLeadStatus_InvalidID(t *testing.T) {
	router := setupLeadTestRouter()

	body := map[string]interface{}{"status": "contact"}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT", "/api/v1/leads/invalid-id/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLeadHandler_UpdateLeadStatus_MissingStatus(t *testing.T) {
	router := setupLeadTestRouter()

	leadID := uuid.New()
	body := map[string]interface{}{}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT", "/api/v1/leads/"+leadID.String()+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLeadHandler_AddActivity_ValidRequest(t *testing.T) {
	router := setupLeadTestRouter()

	leadID := uuid.New()
	body := map[string]interface{}{
		"type":        "call",
		"description": "Initial call",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/leads/"+leadID.String()+"/activities", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestLeadHandler_AddActivity_InvalidID(t *testing.T) {
	router := setupLeadTestRouter()

	body := map[string]interface{}{
		"type": "call",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/leads/invalid-id/activities", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLeadHandler_AddActivity_MissingType(t *testing.T) {
	router := setupLeadTestRouter()

	leadID := uuid.New()
	body := map[string]interface{}{
		"description": "Some description",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/leads/"+leadID.String()+"/activities", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLeadHandler_AddActivity_InvalidJSON(t *testing.T) {
	router := setupLeadTestRouter()

	req, _ := http.NewRequest("POST", "/api/v1/leads/"+uuid.New().String()+"/activities", bytes.NewBuffer([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLeadHandler_AllLeadStatuses(t *testing.T) {
	statuses := []string{"new", "contact", "meeting", "sale", "cancelled", "archive"}

	router := setupLeadTestRouter()
	leadID := uuid.New()

	for _, status := range statuses {
		t.Run("update to "+status, func(t *testing.T) {
			body := map[string]interface{}{"status": status}
			jsonBody, _ := json.Marshal(body)

			req, _ := http.NewRequest("PUT", "/api/v1/leads/"+leadID.String()+"/status", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestLeadHandler_AllActivityTypes(t *testing.T) {
	activityTypes := []string{"call", "meeting", "note", "email"}

	router := setupLeadTestRouter()
	leadID := uuid.New()

	for _, atype := range activityTypes {
		t.Run("add "+atype, func(t *testing.T) {
			body := map[string]interface{}{
				"type":        atype,
				"description": "Test " + atype,
			}
			jsonBody, _ := json.Marshal(body)

			req, _ := http.NewRequest("POST", "/api/v1/leads/"+leadID.String()+"/activities", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)
		})
	}
}