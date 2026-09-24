package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"
	"franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type capturePlanService struct {
	mock.Mock
	lastOpts repository.ListOptions
}

func (m *capturePlanService) CreatePlan(ctx context.Context, dto services.CreatePlanDTO) (*models.Plan, error) {
	args := m.Called(ctx, dto)
	return nil, args.Error(1)
}

func (m *capturePlanService) GetPlan(ctx context.Context, id string) (*models.Plan, error) {
	args := m.Called(ctx, id)
	return nil, args.Error(1)
}

func (m *capturePlanService) ListPlans(ctx context.Context, opts repository.ListOptions) ([]*models.Plan, int64, error) {
	m.lastOpts = opts
	return []*models.Plan{}, 0, nil
}

func (m *capturePlanService) UpdatePlan(ctx context.Context, id string, dto services.UpdatePlanDTO) (*models.Plan, error) {
	args := m.Called(ctx, id, dto)
	return nil, args.Error(1)
}

func (m *capturePlanService) DeletePlan(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// RE-AUDIT: сырой sort уходил в SQL ORDER BY; size без потолка.
func TestPlanList_SortInjectionNeutralized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(capturePlanService)
	r := gin.New()
	g := r.Group("/api/v1")
	g.Use(func(c *gin.Context) {
		c.Set("role", "super_admin")
		c.Next()
	})
	NewPlanHandler(g, svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/plans?sort=(CASE+WHEN+1%3D1+THEN+name+ELSE+price+END)", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "created_at", svc.lastOpts.OrderBy, "вне whitelist — дефолт")
}

func TestPlanList_SizeCapped(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(capturePlanService)
	r := gin.New()
	g := r.Group("/api/v1")
	g.Use(func(c *gin.Context) {
		c.Set("role", "super_admin")
		c.Next()
	})
	NewPlanHandler(g, svc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/plans?size=1000000&page=0", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, 100, svc.lastOpts.Limit, "потолок страницы")
	require.Equal(t, 0, svc.lastOpts.Offset, "page<1 → первая страница")
}
