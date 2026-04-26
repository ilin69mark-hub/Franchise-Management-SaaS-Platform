package handlers // <-- оставляем, остальные хэндлеры тоже в этом пакете

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	// **ВАЖНО**: импортируем папку services под алиасом `service`
	service "franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------
//  План‑хэндлер
// ---------------------------------------------------------------------

type PlanHandler struct {
	svc service.PlanService
}

// NewPlanHandler регистрирует роуты /api/v1/plans.
// r – любой *gin.RouterGroup (protected, admin и т.п.)
func NewPlanHandler(r *gin.RouterGroup, svc service.PlanService) {
	h := &PlanHandler{svc: svc}
	plans := r.Group("/plans")
	{
		plans.POST("", h.Create)       // POST   /plans
		plans.GET("", h.List)          // GET    /plans
		plans.GET("/:id", h.Get)       // GET    /plans/:id
		plans.PATCH("/:id", h.Update)  // PATCH  /plans/:id
		plans.DELETE("/:id", h.Delete) // DELETE /plans/:id
	}
}

// ---------------------------------------------------------------------
//  DTO для ответа (чтобы не возвращать GORM‑модель напрямую)
// ---------------------------------------------------------------------

type PlanResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	MaxSalons int     `json:"max_salons"`
	MaxUsers  int     `json:"max_users"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func toResp(p *models.Plan) PlanResponse {
	return PlanResponse{
		ID:        p.ID.String(),
		Name:      p.Name,
		Price:     p.Price,
		MaxSalons: p.MaxSalons,
		MaxUsers:  p.MaxUsers,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}

// ---------------------------------------------------------------------
//  Handlers
// ---------------------------------------------------------------------

// POST /plans
func (h *PlanHandler) Create(c *gin.Context) {
	var dto service.CreatePlanDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := h.svc.CreatePlan(c.Request.Context(), dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toResp(plan))
}

// GET /plans/:id
func (h *PlanHandler) Get(c *gin.Context) {
	id := c.Param("id")
	plan, err := h.svc.GetPlan(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}
	c.JSON(http.StatusOK, toResp(plan))
}

// GET /plans
// Параметры: search, page, size, sort, desc
func (h *PlanHandler) List(c *gin.Context) {
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	sort := c.DefaultQuery("sort", "created_at")
	desc := c.DefaultQuery("desc", "false") == "true"

	offset := (page - 1) * size
	opts := repository.ListOptions{
		Search:  search,
		OrderBy: sort,
		Desc:    desc,
		Limit:   size,
		Offset:  offset,
	}

	plans, total, err := h.svc.ListPlans(c.Request.Context(), opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := make([]PlanResponse, len(plans))
	for i, p := range plans {
		resp[i] = toResp(p)
	}
	c.JSON(http.StatusOK, gin.H{
		"items": resp,
		"total": total,
	})
}

// PATCH /plans/:id
func (h *PlanHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var dto service.UpdatePlanDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := h.svc.UpdatePlan(c.Request.Context(), id, dto)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toResp(plan))
}

// DELETE /plans/:id
func (h *PlanHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeletePlan(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
