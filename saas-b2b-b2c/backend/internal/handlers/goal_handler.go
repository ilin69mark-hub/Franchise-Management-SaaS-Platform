package handlers

import (
	"net/http"
	"time"

	"franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type GoalHandler struct {
	svc services.GoalService
}

func NewGoalHandler(s services.GoalService) *GoalHandler {
	return &GoalHandler{svc: s}
}

// POST /goals – создать/обновить цель
func (h *GoalHandler) SetGoal(c *gin.Context) {
	var dto services.CreateGoalDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// получаем assigner (текущий пользователь) и tenant из контекста
	assignerID, _ := c.Get("userID")
	tenantID, _ := c.Get("tenantID")

	goal, err := h.svc.CreateGoal(c.Request.Context(), dto, assignerID.(string), tenantID.(string))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, goal)
}

// GET /goals?date=2024-09-01 – план текущего пользователя на дату
func (h *GoalHandler) GetGoal(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date query required"})
		return
	}
	date, _ := time.Parse("2006-01-02", dateStr)

	userID, _ := c.Get("userID")
	goal, err := h.svc.GetMyGoal(c.Request.Context(), userID.(string), date)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}
	c.JSON(http.StatusOK, goal)
}

// GET /goals/by-date/:date – план текущего пользователя на конкретную дату
func (h *GoalHandler) GetGoalByDate(c *gin.Context) {
	dateStr := c.Param("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date required"})
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
		return
	}

	userID, _ := c.Get("userID")
	goal, err := h.svc.GetMyGoal(c.Request.Context(), userID.(string), date)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}
	c.JSON(http.StatusOK, goal)
}

// GET /goals/visible – список целей, которые видит пользователь
func (h *GoalHandler) GetVisibleGoals(c *gin.Context) {
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")
	tenantID, _ := c.Get("tenantID")

	list, err := h.svc.GetVisibleGoals(c.Request.Context(),
		userID.(string), role.(string), tenantID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// DELETE /goals/:id – удалить цель (по желанию)
func (h *GoalHandler) DeleteGoal(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeleteGoal(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
