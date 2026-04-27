package handlers

import (
	"net/http"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KPIHandler struct {
	kpiSvc   *services.KPIService
	schedSvc *services.ScheduleService
}

func NewKPIHandler(kpiSvc *services.KPIService, schedSvc *services.ScheduleService) *KPIHandler {
	return &KPIHandler{kpiSvc: kpiSvc, schedSvc: schedSvc}
}

// --- KPI Stats ---

func (h *KPIHandler) GetMyStats(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	if user.SalonID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not assigned to salon"})
		return
	}

	stats, err := h.kpiSvc.GetDashboardStats(c.Request.Context(), user.ID, *user.SalonID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *KPIHandler) GetSalonStats(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	salonIDStr := c.Query("salon_id")
	salonID, err := uuid.Parse(salonIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid salon_id"})
		return
	}

	stats, err := h.kpiSvc.GetDashboardStats(c.Request.Context(), user.ID, salonID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *KPIHandler) SetGoal(c *gin.Context) {
	var req models.SetGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, _ := time.Parse("2006-01-02", req.TargetDate)
	goal := models.DailyGoal{
		TargetDate:   date,
		SalesPlan:    req.SalesPlan,
		LeadsPlan:    req.LeadsPlan,
		CallsPlan:    req.CallsPlan,
		MeetingsPlan: req.MeetingsPlan,
	}

	if req.SalonID != "" {
		id, _ := uuid.Parse(req.SalonID)
		goal.SalonID = &id
	}
	if req.UserID != "" {
		id, _ := uuid.Parse(req.UserID)
		goal.UserID = &id
	}

	// Вызываем репозиторий через сервис (добавим метод SetGoal в сервис)
	if err := h.kpiSvc.SetGoal(c.Request.Context(), &goal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Goal set successfully"})
}

func (h *KPIHandler) GetTeamAnalytics(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	period := c.DefaultQuery("period", "day")

	data, err := h.kpiSvc.GetTeamAnalytics(c.Request.Context(), user.ID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// --- Schedule (Manager) ---

func (h *KPIHandler) GetSchedule(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	events, err := h.schedSvc.GetUserSchedule(c.Request.Context(), user.ID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

func (h *KPIHandler) CreateEvent(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	var req models.CreateScheduleEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, _ := time.Parse(time.RFC3339, req.StartTime)

	event := &models.ScheduleEvent{
		UserID:      user.ID,
		SalonID:     *user.SalonID,
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		StartTime:   startTime,
		Priority:    req.Priority,
		Status:      "planned",
	}

	if err := h.schedSvc.CreateEvent(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, event)
}

func (h *KPIHandler) UpdateEventStatus(c *gin.Context) {
	idStr := c.Param("id")
	eventID, _ := uuid.Parse(idStr)

	var req models.UpdateScheduleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.schedSvc.UpdateEventStatus(c.Request.Context(), eventID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
}

// --- Schedule (Dealer Admin) ---

func (h *KPIHandler) GetAllSchedule(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	dateStr := c.Query("date")

	// Получаем ID всех менеджеров
	var managerIDs []uuid.UUID
	db, _ := c.Get("db")
	db.(*gorm.DB).Model(&models.User{}).Where("managed_by = ?", user.ID).Pluck("id", &managerIDs)

	var events []models.ScheduleEvent
	events, err = h.schedSvc.GetEventsByUsers(c.Request.Context(), managerIDs, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

func (h *KPIHandler) CreateEventForManager(c *gin.Context) {
	var req models.CreateScheduleEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	event := &models.ScheduleEvent{
		UserID:      uuid.MustParse(req.UserID), // Берем ID менеджера из тела запроса
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		StartTime:   startTime,
		Priority:    req.Priority,
		Status:      "planned",
		// SalonID нужно заполнить, если он есть в запросе или вычислить
	}

	// Если нужно SalonID, можно взять из базы по UserID или передать в запросе
	// Для простоты оставим без SalonID или добавим поле в Request

	if err := h.schedSvc.CreateEvent(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, event)
}

func (h *KPIHandler) UpdateEvent(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var req models.UpdateScheduleEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.schedSvc.UpdateEvent(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *KPIHandler) DeleteEvent(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.schedSvc.DeleteEvent(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
