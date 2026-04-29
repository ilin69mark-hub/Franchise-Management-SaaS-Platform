package handlers

import (
	"net/http"
	"strconv"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"
	"franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KPIHandler struct {
	db       *gorm.DB
	kpiSvc   *services.KPIService
	schedSvc *services.ScheduleService
	alertSvc *services.AlertService
}

func NewKPIHandler(db *gorm.DB, kpiSvc *services.KPIService, schedSvc *services.ScheduleService, alertSvc *services.AlertService) *KPIHandler {
	return &KPIHandler{db: db, kpiSvc: kpiSvc, schedSvc: schedSvc, alertSvc: alertSvc}
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

// GetDealerSummary - сводка для дилера (все салоны)
func (h *KPIHandler) GetDealerSummary(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	data, err := h.kpiSvc.GetDealerSummary(c.Request.Context(), user.ID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetDealerFinance - финансы для дилера
func (h *KPIHandler) GetDealerFinance(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	data, err := h.kpiSvc.GetDealerFinance(c.Request.Context(), user.ID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetDealerFunnel - воронка для дилера
func (h *KPIHandler) GetDealerFunnel(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	period := c.DefaultQuery("period", "month")

	data, err := h.kpiSvc.GetDealerFunnel(c.Request.Context(), user.ID, period, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetDealerProducts - товары и склад для дилера
func (h *KPIHandler) GetDealerProducts(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	data, err := h.kpiSvc.GetDealerProducts(c.Request.Context(), user.ID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetFranchiserSummary - сводка для франчайзера (все дилеры)
func (h *KPIHandler) GetFranchiserSummary(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	data, err := h.kpiSvc.GetFranchiserSummary(c.Request.Context(), user.ID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetFranchiserNetwork - данные сети для франчайзера
func (h *KPIHandler) GetFranchiserNetwork(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	period := c.DefaultQuery("period", "month")
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	data, err := h.kpiSvc.GetFranchiserNetwork(c.Request.Context(), user.ID, period, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetFranchiserHealth - здоровье сети
func (h *KPIHandler) GetFranchiserHealth(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	data, err := h.kpiSvc.GetFranchiserHealth(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetFranchiserTeam - команда франчайзера
func (h *KPIHandler) GetFranchiserTeam(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	data, err := h.kpiSvc.GetFranchiserTeam(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetTerritorySummary - сводка для территориального менеджера
func (h *KPIHandler) GetTerritorySummary(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	data, err := h.kpiSvc.GetTerritorySummary(c.Request.Context(), user.ID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetTerritoryFunnel - воронка для территориального менеджера
func (h *KPIHandler) GetTerritoryFunnel(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	period := c.DefaultQuery("period", "month")
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	data, err := h.kpiSvc.GetTerritoryFunnel(c.Request.Context(), user.ID, period, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetTerritoryPlanFact - план-факт для территориального менеджера
func (h *KPIHandler) GetTerritoryPlanFact(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	period := c.DefaultQuery("period", "month")
	data, err := h.kpiSvc.GetTerritoryPlanFact(c.Request.Context(), user.ID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetTerritoryCommunications - коммуникации
func (h *KPIHandler) GetTerritoryCommunications(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	data, err := h.kpiSvc.GetTerritoryCommunications(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetTerritoryBenchmarks - бенчмарки
func (h *KPIHandler) GetTerritoryBenchmarks(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	data, err := h.kpiSvc.GetTerritoryBenchmarks(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// === Dealer Tasks Handlers ===

// GetDealerTasks - получить задачи от франчайзера
func (h *KPIHandler) GetDealerTasks(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	data, err := h.kpiSvc.GetDealerTasks(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// UpdateDealerTask - обновить задачу
func (h *KPIHandler) UpdateDealerTask(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	taskID := c.Param("id")
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = h.kpiSvc.UpdateDealerTask(c.Request.Context(), user.ID.String(), taskID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task updated"})
}

// === Dealer Requests Handlers ===

// GetDealerRequests - получить запросы к бренду
func (h *KPIHandler) GetDealerRequests(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	status := c.Query("status")
	data, err := h.kpiSvc.GetDealerRequests(c.Request.Context(), user.ID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// CreateDealerRequest - создать запрос к бренду
func (h *KPIHandler) CreateDealerRequest(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	var req struct {
		Type        string  `json:"type" binding:"required"`
		Description string  `json:"description" binding:"required"`
		Amount     float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = h.kpiSvc.CreateDealerRequest(c.Request.Context(), user.ID, req.Type, req.Description, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Request created"})
}

// === Dealer Marketing Budget Handlers ===

// GetDealerMarketingBudget - получить маркетинговый бюджет
func (h *KPIHandler) GetDealerMarketingBudget(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	quarter := c.DefaultQuery("quarter", "2026-Q2")
	data, err := h.kpiSvc.GetDealerMarketingBudget(c.Request.Context(), user.ID, quarter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// === Dealer Alerts Handlers ===

// GetDealerAlerts - получить алерты дилера
func (h *KPIHandler) GetDealerAlerts(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	data, err := h.kpiSvc.GetDealerAlerts(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// MarkDealerAlertRead - пометить алерт прочитанным
func (h *KPIHandler) MarkDealerAlertRead(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	alertID := c.Param("id")
	var alertUUID uuid.UUID
	alertUUID, err = uuid.Parse(alertID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}
	err = h.kpiSvc.MarkDealerAlertRead(c.Request.Context(), user.ID, alertUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Alert marked as read"})
}

// MarkAllDealerAlertsRead - пометить все алерты прочитанными
func (h *KPIHandler) MarkAllDealerAlertsRead(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	err = h.kpiSvc.MarkAllDealerAlertsRead(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "All alerts marked as read"})
}

// GetFranchiserDealers - получить список дилеров франчайзера
func (h *KPIHandler) GetFranchiserDealers(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	filter := c.DefaultQuery("filter", "all")
	data, err := h.kpiSvc.GetFranchiserDealers(c.Request.Context(), user.ID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetFranchiserRequests - получить запросы от дилеров
func (h *KPIHandler) GetFranchiserRequests(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	filter := c.DefaultQuery("filter", "all")
	data, err := h.kpiSvc.GetFranchiserRequests(c.Request.Context(), user.ID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetFranchiserAlerts - получить алерты для франчайзера
func (h *KPIHandler) GetFranchiserAlerts(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	notifRepo := repository.NewNotificationRepository(h.db)
	notifSvc := services.NewNotificationService(notifRepo)
	
	var tenantID uuid.UUID
	if user.TenantID != nil {
		tenantID = *user.TenantID
	}
	data, err := notifSvc.GetNotifications(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"alerts": data, "unread_count": len(data)})
}

// GetTerritoriesHeatmap - тепловая карта территорий
func (h *KPIHandler) GetTerritoriesHeatmap(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	period := c.DefaultQuery("period", "month")
	data, err := h.kpiSvc.GetTerritoriesHeatmap(c.Request.Context(), user.ID.String(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetManagerDynamics - динамика менеджера
func (h *KPIHandler) GetManagerDynamics(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	managerID := c.Param("id")
	months := c.DefaultQuery("months", "6")
	data, err := h.kpiSvc.GetManagerDynamics(c.Request.Context(), user.ID.String(), managerID, months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetManagerDealers - дилеры менеджера
func (h *KPIHandler) GetManagerDealers(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	managerID := c.Param("id")
	data, err := h.kpiSvc.GetManagerDealers(c.Request.Context(), user.ID.String(), managerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// SetManagerPlans - установить планы менеджеров
func (h *KPIHandler) SetManagerPlans(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	var req struct {
		Quarter string `json:"quarter"`
		Plans   []struct {
			ManagerID   string  `json:"manager_id"`
			PlanAmount  float64 `json:"plan_amount"`
			TargetDealers int   `json:"target_dealers"`
		} `json:"plans"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
err = h.kpiSvc.SetManagerPlans(c.Request.Context(), user.ID.String(), req.Quarter, req.Plans)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Plans saved"})
}

// GetManagerPlans - получить планы менеджеров
func (h *KPIHandler) GetManagerPlans(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	quarter := c.DefaultQuery("quarter", "2026-Q2")
	data, err := h.kpiSvc.GetManagerPlans(c.Request.Context(), user.ID.String(), quarter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetDealersHealth - здоровье дилеров (сегментация)
func (h *KPIHandler) GetDealersHealth(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	period := c.DefaultQuery("period", "quarter")
	data, err := h.kpiSvc.GetDealersHealth(c.Request.Context(), user.ID.String(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetDealersMigration - миграция дилеров
func (h *KPIHandler) GetDealersMigration(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	period := c.DefaultQuery("period", "quarter")
	data, err := h.kpiSvc.GetDealersMigration(c.Request.Context(), user.ID.String(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetSystemIssues - системные проблемы
func (h *KPIHandler) GetSystemIssues(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	status := c.Query("status")
	data, err := h.kpiSvc.GetSystemIssues(c.Request.Context(), user.ID.String(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetDealersGeography - география дилеров
func (h *KPIHandler) GetDealersGeography(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	data, err := h.kpiSvc.GetDealersGeography(c.Request.Context(), user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetMarketingROI - ROI маркетинга
func (h *KPIHandler) GetMarketingROI(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	period := c.DefaultQuery("period", "quarter")
	data, err := h.kpiSvc.GetMarketingROI(c.Request.Context(), user.ID.String(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// MarkFranchiserAlertRead - отметить алерт прочитанным
func (h *KPIHandler) MarkFranchiserAlertRead(c *gin.Context) {
	alertIDStr := c.Param("id")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}
	notifRepo := repository.NewNotificationRepository(h.db)
	notifSvc := services.NewNotificationService(notifRepo)
	err = notifSvc.MarkAsRead(c.Request.Context(), alertID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Alert marked as read"})
}

// MarkAllFranchiserAlertsRead - отметить все алерты прочитанными
func (h *KPIHandler) MarkAllFranchiserAlertsRead(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	notifRepo := repository.NewNotificationRepository(h.db)
	notifSvc := services.NewNotificationService(notifRepo)
	err = notifSvc.MarkAllAsRead(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "All alerts marked as read"})
}

// AssignAlert - назначить ответственного за алерт
func (h *KPIHandler) AssignAlert(c *gin.Context) {
	alertIDStr := c.Param("id")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}
	var req struct {
		AssignTo string `json:"assign_to"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Alert assigned", "alert_id": alertID})
}

// GetAlertSettings - настройки алертов
func (h *KPIHandler) GetAlertSettings(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	data, err := h.kpiSvc.GetAlertSettings(c.Request.Context(), user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// UpdateAlertSettings - обновить настройки алертов
func (h *KPIHandler) UpdateAlertSettings(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	var settings map[string]interface{}
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = h.kpiSvc.UpdateAlertSettings(c.Request.Context(), user.ID.String(), settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Settings updated"})
}

// GetReportData - данные для отчёта
func (h *KPIHandler) GetReportData(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	period := c.DefaultQuery("period", "quarter")
	date := c.DefaultQuery("date", "2026-Q2")
	data, err := h.kpiSvc.GetReportData(c.Request.Context(), user.ID.String(), period, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GeneratePDF - сгенерировать PDF
func (h *KPIHandler) GeneratePDF(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	var req struct {
		Blocks   []string `json:"blocks"`
		Comment  string   `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, err := h.kpiSvc.GeneratePDF(c.Request.Context(), user.ID.String(), req.Blocks, req.Comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// SendReport - отправить отчёт
func (h *KPIHandler) SendReport(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	var req struct {
		ReportID   string   `json:"report_id"`
		Recipients []string `json:"recipients"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = h.kpiSvc.SendReport(c.Request.Context(), user.ID.String(), req.ReportID, req.Recipients)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Report sent"})
}

// GetReportHistory - история отчётов
func (h *KPIHandler) GetReportHistory(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	data, err := h.kpiSvc.GetReportHistory(c.Request.Context(), user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// SaveDraft - сохранить черновик
func (h *KPIHandler) SaveDraft(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	var draft map[string]interface{}
	if err := c.ShouldBindJSON(&draft); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = h.kpiSvc.SaveDraft(c.Request.Context(), user.ID.String(), draft)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Draft saved"})
}

// GetDraft - получить черновик
func (h *KPIHandler) GetDraft(c *gin.Context) {
	user, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	data, err := h.kpiSvc.GetDraft(c.Request.Context(), user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
