package handlers

import (
	"log"
	"net/http"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChecklistHandler struct {
	service *services.ChecklistService
}

func NewChecklistHandler(service *services.ChecklistService) *ChecklistHandler {
	return &ChecklistHandler{service: service}
}

func (h *ChecklistHandler) GetChecklists(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	status := c.Query("status")
	priority := c.Query("priority")

	log.Printf("DEBUG GetChecklists: User %s (Role: %s)", currentUser.Email, currentUser.Role)

	var items []models.Checklist

	if currentUser.Role == models.RoleSuperAdmin {
		log.Println("DEBUG GetChecklists: Path -> SUPER ADMIN (Global)")
		items, err = h.service.GetAllGlobal(c.Request.Context(), status, priority)
	} else {
		log.Println("DEBUG GetChecklists: Path -> STANDARD USER")
		items, err = h.service.GetAllChecklists(c.Request.Context(), status, priority)
	}

	if err != nil {
		log.Printf("ERROR GetChecklists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("DEBUG GetChecklists: Returning %d tasks", len(items))
	c.JSON(http.StatusOK, items)
}

func (h *ChecklistHandler) CreateChecklist(c *gin.Context) {
	var req models.ChecklistCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	item := models.Checklist{
		Title:       req.Title,
		Description: req.Description,
		UserID:      currentUser.ID,
		TenantID:    currentUser.TenantID,
		AssignedTo:  req.AssignedTo,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Recurrence:  req.Recurrence,
		Status:      "pending",
		Priority:    req.Priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if item.Priority == "" {
		item.Priority = "normal"
	}

	if err := h.service.CreateChecklist(c.Request.Context(), &item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *ChecklistHandler) CompleteChecklist(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := uuid.Parse(idStr)
	if err := h.service.CompleteChecklist(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Completed"})
}

func (h *ChecklistHandler) GetChecklistByID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := uuid.Parse(idStr)
	item, err := h.service.GetChecklistByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *ChecklistHandler) UpdateChecklist(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := uuid.Parse(idStr)
	var req models.ChecklistUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existing, err := h.service.GetChecklistByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	if req.Title != "" {
		existing.Title = req.Title
	}
	existing.Description = req.Description
	existing.AssignedTo = req.AssignedTo
	existing.StartDate = req.StartDate
	existing.EndDate = req.EndDate
	if req.Status != "" {
		existing.Status = req.Status
	}
	if req.Priority != "" {
		existing.Priority = req.Priority
	}
	existing.UpdatedAt = time.Now()

	if err := h.service.UpdateChecklist(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}

func (h *ChecklistHandler) DeleteChecklist(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := uuid.Parse(idStr)
	if err := h.service.DeleteChecklist(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// UpdateStatus - обновляет только статус задачи (для KPI и дашбордов)
func (h *ChecklistHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем текущую задачу
	existing, err := h.service.GetChecklistByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	// Обновляем статус и время
	existing.Status = req.Status
	existing.UpdatedAt = time.Now()

	if err := h.service.UpdateChecklist(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, existing)
}
