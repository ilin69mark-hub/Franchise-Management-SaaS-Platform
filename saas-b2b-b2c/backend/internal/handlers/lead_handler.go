package handlers

import (
	"log"
	"net/http"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LeadHandler struct {
	service *services.LeadService
}

func NewLeadHandler(service *services.LeadService) *LeadHandler {
	return &LeadHandler{service: service}
}

// GetMyLeads получает всех лидов текущего менеджера
func (h *LeadHandler) GetMyLeads(c *gin.Context) {
	currentUser := getCurrentUser(c)

	leads, err := h.service.GetMyLeads(c.Request.Context(), currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, leads)
}

// CreateLead создает нового лида
func (h *LeadHandler) CreateLead(c *gin.Context) {
	// Получаем текущего пользователя
	currentUser := getCurrentUser(c)

	// ОТЛАДКА: Выводим, что пришло в заголовке
	log.Printf("DEBUG CreateLead: User ID: %s, Role: %s, SalonID: %v", currentUser.ID, currentUser.Role, currentUser.SalonID)

	var req models.CreateLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Передаем currentUser целиком
	lead, err := h.service.CreateLead(c.Request.Context(), currentUser, req)
	if err != nil {
		log.Printf("ERROR CreateLead: %v", err) // Логируем ошибку
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, lead)
}

// GetLeadByID детальная информация по лиду + история
func (h *LeadHandler) GetLeadByID(c *gin.Context) {
	currentUser := getCurrentUser(c)
	idStr := c.Param("id")
	leadID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	lead, activities, err := h.service.GetLeadDetails(c.Request.Context(), currentUser.ID, leadID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lead not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"lead":       lead,
		"activities": activities,
	})
}

// UpdateLeadStatus обновляет статус (Kanban drag-n-drop)
func (h *LeadHandler) UpdateLeadStatus(c *gin.Context) {
	currentUser := getCurrentUser(c)
	idStr := c.Param("id")
	leadID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req models.UpdateLeadStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateStatus(c.Request.Context(), currentUser.ID, leadID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
}

// AddActivity добавляет заметку/звонок в историю
func (h *LeadHandler) AddActivity(c *gin.Context) {
	currentUser := getCurrentUser(c)
	idStr := c.Param("id")
	leadID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req models.AddLeadActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddActivity(c.Request.Context(), currentUser.ID, leadID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Activity added"})
}
