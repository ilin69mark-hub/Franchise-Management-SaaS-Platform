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

// checklistOwnedByCaller — REAUDIT-4: авторизация на ЗАПИСЬ строго по
// владению (создатель или исполнитель). Раньше правило "тот же tenant" давало
// любому сотруднику сети менять/завершать/удалять чужие чеклисты (доказано
// живым прогоном). Tenant-скоуп остаётся только для ЧТЕНИЯ списка.
func checklistOwnedByCaller(user *models.User, chk *models.Checklist) bool {
	if user == nil || chk == nil {
		return false
	}
	if user.Role == models.RoleSuperAdmin {
		return true
	}
	if chk.UserID == user.ID {
		return true
	}
	// исполнитель — только если он действительно в той же сети
	if chk.AssignedTo != nil && *chk.AssignedTo == user.ID {
		if chk.TenantID == nil || user.TenantID == nil || *chk.TenantID == *user.TenantID {
			return true
		}
	}
	return false
}

// checklistReadableByCaller — чтение: владелец, исполнитель или та же сеть.
func checklistReadableByCaller(user *models.User, chk *models.Checklist) bool {
	if checklistOwnedByCaller(user, chk) {
		return true
	}
	if user == nil || chk == nil || user.Role == models.RoleSuperAdmin {
		return false
	}
	return chk.TenantID != nil && user.TenantID != nil && *chk.TenantID == *user.TenantID
}

func (h *ChecklistHandler) GetChecklists(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	status := c.Query("status")
	priority := c.Query("priority")

	var items []models.Checklist

	if currentUser.Role == models.RoleSuperAdmin {
		items, err = h.service.GetAllGlobal(c.Request.Context(), status, priority)
	} else {
		items, err = h.service.GetAllChecklists(c.Request.Context(), currentUser.TenantID, status, priority)
	}

	if err != nil {
		log.Printf("ERROR GetChecklists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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

	// REAUDIT-3: исполнитель обязан быть в том же tenant.
	if err := h.service.ValidateAssignee(c.Request.Context(), req.AssignedTo, currentUser.TenantID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
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
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	existing, err := h.service.GetChecklistByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	if !checklistOwnedByCaller(user, existing) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if err := h.service.CompleteChecklist(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Completed"})
}

func (h *ChecklistHandler) GetChecklistByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	item, err := h.service.GetChecklistByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	if !checklistReadableByCaller(user, item) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
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

	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	// REAUDIT-4: изменение — это запись, читать могут все в сети, писать — только
	// создатель/исполнитель (иначе коллега переписывал чужие задачи, live 200).
	if !checklistOwnedByCaller(user, existing) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// REAUDIT-3: новый исполнитель — тоже внутри tenant (проверяем до мутации).
	if req.AssignedTo != nil {
		if err := h.service.ValidateAssignee(c.Request.Context(), req.AssignedTo, user.TenantID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
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
	// W1: recurrence раньше молча терялся при обновлении.
	if req.Recurrence != "" {
		existing.Recurrence = req.Recurrence
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
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	existing, err := h.service.GetChecklistByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	if !checklistOwnedByCaller(user, existing) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
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

	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	if !checklistOwnedByCaller(user, existing) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
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
