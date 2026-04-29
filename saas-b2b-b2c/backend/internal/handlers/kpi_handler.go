package handlers

import (
	"net/http"
	"strconv"
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
	alertSvc *services.AlertService
}

func NewKPIHandler(kpiSvc *services.KPIService, schedSvc *services.ScheduleService, alertSvc *services.AlertService) *KPIHandler {
	return &KPIHandler{kpiSvc: kpiSvc, schedSvc: schedSvc, alertSvc: alertSvc}
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

// --- Dashboard Main (Salon Manager) ---

func (h *KPIHandler) GetDashboardMain(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	if user.SalonID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not assigned to salon"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	data, err := h.kpiSvc.GetDashboardMain(c.Request.Context(), user.ID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// TopBar - верхняя панель дашборда
func (h *KPIHandler) GetTopBar(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	if user.SalonID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not assigned to salon"})
		return
	}

	type TopBarResponse struct {
		SalonName       string  `json:"salon_name"`
		PlanPercent     int     `json:"plan_percent"`
		AvgCheck        float64 `json:"avg_check"`
		PrepaymentsSum  float64 `json:"prepayments_sum"`
		AlertsCount     int     `json:"alerts_count"`
	}

	resp := TopBarResponse{
		SalonName: "Мой салон",
		PlanPercent: 0,
		AvgCheck: 0,
		PrepaymentsSum: 0,
		AlertsCount: 0,
	}

	data, err := h.kpiSvc.GetDashboardMain(c.Request.Context(), user.ID, time.Now().Format("2006-01-02"))
	if err == nil {
		resp.SalonName = "Салон №1"
		resp.PlanPercent = data.PlanPercent
		resp.AvgCheck = data.AvgCheck
		resp.PrepaymentsSum = data.PrepaymentsSum
		resp.AlertsCount = len(data.PendingPayments)
	}

	c.JSON(http.StatusOK, resp)
}

// GetDashboardFunnel - воронка продаж
func (h *KPIHandler) GetDashboardFunnel(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	if user.SalonID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not assigned to salon"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	data, err := h.kpiSvc.GetDashboardFunnel(c.Request.Context(), user.ID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetDashboardTeam - команда (рейтинг продавцов)
func (h *KPIHandler) GetDashboardTeam(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	if user.SalonID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not assigned to salon"})
		return
	}

	period := c.DefaultQuery("period", "month")
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	data, err := h.kpiSvc.GetDashboardTeam(c.Request.Context(), user.ID, period, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetSalesRepHistory - история продавца
func (h *KPIHandler) GetSalesRepHistory(c *gin.Context) {
	idStr := c.Param("id")
	managerID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manager ID"})
		return
	}

	months := 6
	if m := c.Query("months"); m != "" {
		if parsed, _ := strconv.Atoi(m); parsed > 0 && parsed <= 12 {
			months = parsed
		}
	}

	data, err := h.kpiSvc.GetSalesRepHistory(c.Request.Context(), managerID, months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetDashboardProducts - товары
func (h *KPIHandler) GetDashboardProducts(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	if user.SalonID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not assigned to salon"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	data, err := h.kpiSvc.GetDashboardProducts(c.Request.Context(), user.ID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetAlerts - получить алерты
func (h *KPIHandler) GetAlerts(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	// Сначала генерируем алерты
	h.alertSvc.GenerateAlertsForUser(c.Request.Context(), user.ID)

	alerts, count, err := h.alertSvc.GetUnreadAlerts(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type AlertResponse struct {
		Alerts     []models.Notification `json:"alerts"`
		UnreadCount int                   `json:"unread_count"`
	}

	c.JSON(http.StatusOK, AlertResponse{Alerts: alerts, UnreadCount: count})
}

// MarkAlertRead - пометить алерт прочитанным
func (h *KPIHandler) MarkAlertRead(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	alertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	if err := h.alertSvc.MarkAsRead(c.Request.Context(), user.ID, alertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert marked as read"})
}

// GetManagerTargets - директивы от дилера
func (h *KPIHandler) GetManagerTargets(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	if user.SalonID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not assigned to salon"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	data, err := h.kpiSvc.GetManagerTargets(c.Request.Context(), user.ID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
