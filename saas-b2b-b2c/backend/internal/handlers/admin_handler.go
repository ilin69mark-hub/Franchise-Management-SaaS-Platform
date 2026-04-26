package handlers

import (
	"franchise-saas-backend/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	service *services.AdminService
}

func NewAdminHandler(service *services.AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

// GetDashboardStats - статистика
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.service.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetAllTenants - список сетей
func (h *AdminHandler) GetAllTenants(c *gin.Context) {
	tenants, err := h.service.GetAllTenants(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tenants)
}

// GetTenantsPaymentStatus - статус оплат
func (h *AdminHandler) GetTenantsPaymentStatus(c *gin.Context) {
	data, err := h.service.GetTenantsPaymentStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// CreateTenant - создать сеть
func (h *AdminHandler) CreateTenant(c *gin.Context) {
	var req struct {
		Name   string     `json:"name" binding:"required"`
		PlanID *uuid.UUID `json:"plan_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenant, err := h.service.CreateTenant(req.Name, req.PlanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tenant)
}

// UpdateTenant - обновить сеть
func (h *AdminHandler) UpdateTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := uuid.Parse(idStr)

	var req struct {
		Name      string     `json:"name"`
		PlanID    *uuid.UUID `json:"plan_id"`
		PaidUntil *time.Time `json:"paid_until"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateTenant(id, req.Name, req.PlanID, req.PaidUntil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Updated"})
}

// BlockTenant - заблокировать
func (h *AdminHandler) BlockTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := uuid.Parse(idStr)
	if err := h.service.BlockTenant(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Blocked"})
}

// DeleteTenant - удалить
func (h *AdminHandler) DeleteTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := uuid.Parse(idStr)
	if err := h.service.DeleteTenant(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// === PLANS ===

func (h *AdminHandler) GetAllPlans(c *gin.Context) {
	plans, err := h.service.GetAllPlans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plans)
}

func (h *AdminHandler) CreatePlan(c *gin.Context) {
	var req struct {
		Name     string  `json:"name" binding:"required"`
		Price    float64 `json:"price" binding:"required"`
		MaxUsers int     `json:"max_users"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := h.service.CreatePlan(req.Name, req.Price, req.MaxUsers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, plan)
}

// UpdatePlan - обновить тариф (НОВОЕ)
func (h *AdminHandler) UpdatePlan(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := uuid.Parse(idStr)

	var req struct {
		Name     string  `json:"name" binding:"required"`
		Price    float64 `json:"price" binding:"required"`
		MaxUsers int     `json:"max_users"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan, err := h.service.UpdatePlan(id, req.Name, req.Price, req.MaxUsers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plan)
}

// DeletePlan - удалить тариф (НОВОЕ)
func (h *AdminHandler) DeletePlan(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := uuid.Parse(idStr)
	if err := h.service.DeletePlan(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// === INVOICES ===

// CreateInvoice - создать счет
func (h *AdminHandler) CreateInvoice(c *gin.Context) {
	var req struct {
		TenantID    uuid.UUID `json:"tenant_id" binding:"required"`
		Amount      float64   `json:"amount" binding:"required"`
		Description string    `json:"description"`
		DueDate     time.Time `json:"due_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inv, err := h.service.CreateInvoice(req.TenantID, req.Amount, req.Description, req.DueDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, inv)
}

// PayInvoice - оплатить счет
func (h *AdminHandler) PayInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := uuid.Parse(idStr)

	if err := h.service.MarkInvoicePaid(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invoice paid successfully"})
}

// === ЭТАП 2: АНАЛИТИКА ===

// GetAnalytics возвращает DAU/MAU и Воронку
func (h *AdminHandler) GetAnalytics(c *gin.Context) {
	data, err := h.service.GetAnalyticsData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// === ЭТАП 3: Продукт и KPI ===

// GetProductAnalytics - возвращает KPI и продуктовые метрики
func (h *AdminHandler) GetProductAnalytics(c *gin.Context) {
	data, err := h.service.GetProductAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// === ЭТАП 4: Риски и Health Score ===

// GetRisks - список сетей в зоне риска
func (h *AdminHandler) GetRisks(c *gin.Context) {
	risks, err := h.service.GetRisksAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, risks)
}

// === ЭТАП 5: Unit Economics ===

// GetUnitEconomics - LTV, CAC, Ratio
func (h *AdminHandler) GetUnitEconomics(c *gin.Context) {
	data, err := h.service.GetUnitEconomics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// UpdateMarketingSpend - ввод расходов на рекламу
func (h *AdminHandler) UpdateMarketingSpend(c *gin.Context) {
	var req struct {
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateMarketingSpend(req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Marketing spend updated"})
}
