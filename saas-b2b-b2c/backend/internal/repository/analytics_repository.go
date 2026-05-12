package repository

import (
	"context"
	"franchise-saas-backend/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AnalyticsRepositoryInterface interface {
	CalculateDashboardStats(ctx context.Context, userID *uuid.UUID, salonID *uuid.UUID, isManager bool) (*models.DashboardStatsResponse, error)
	CalculateDashboardMain(ctx context.Context, userID uuid.UUID, dateStr string) (*models.DashboardMainResponse, error)
	CalculateDashboardFunnel(ctx context.Context, userID uuid.UUID, dateStr string) (*models.DashboardFunnelResponse, error)
	CalculateTerritoryPlanFact(ctx context.Context, userID uuid.UUID, period string) (*models.TerritoryPlanFactResponse, error)
	CalculateManagerTargets(ctx context.Context, userID uuid.UUID, dateStr string) (*models.ManagerTargetsResponse, error)
}

type AnalyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func calcPercent(plan, fact float64) models.KPIItem {
	if plan == 0 {
		return models.KPIItem{}
	}
	return models.KPIItem{
		Plan:    plan,
		Fact:    fact,
		Percent: int((fact / plan) * 100),
	}
}

func (r *AnalyticsRepository) CalculateDashboardStats(ctx context.Context, userID *uuid.UUID, salonID *uuid.UUID, isManager bool) (*models.DashboardStatsResponse, error) {
	today := time.Now()
	todayStr := today.Format("2006-01-02")

	resp := &models.DashboardStatsResponse{Date: todayStr}

	var goal *models.DailyGoal
	query := r.db.WithContext(ctx).Where("target_date = ?", today)
	if userID != nil {
		query = query.Where("user_id = ?", userID)
	} else if salonID != nil {
		query = query.Where("salon_id = ?", salonID)
	}
	if err := query.First(&goal).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if goal == nil {
		goal = &models.DailyGoal{}
	}

	var salesFact float64
	qSales := r.db.Model(&models.Lead{}).Where("status = ?", "sale")
	if isManager && userID != nil {
		qSales = qSales.Where("manager_id = ?", *userID)
	} else if salonID != nil {
		qSales = qSales.Where("salon_id = ?", *salonID)
	}
	qSales.Where("DATE(created_at) = ?", todayStr).Select("COALESCE(SUM(budget), 0)").Scan(&salesFact)

	resp.Sales = calcPercent(goal.SalesPlan, salesFact)
	resp.Leads = models.KPIItem{Plan: float64(goal.LeadsPlan)}
	resp.Calls = models.KPIItem{Plan: float64(goal.CallsPlan)}
	resp.Meetings = models.KPIItem{Plan: float64(goal.MeetingsPlan)}

	return resp, nil
}

func (r *AnalyticsRepository) CalculateDashboardMain(ctx context.Context, userID uuid.UUID, dateStr string) (*models.DashboardMainResponse, error) {
	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = parsed
		}
	}

	resp := &models.DashboardMainResponse{}

	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	if user.SalonID == nil {
		return nil, gorm.ErrRecordNotFound
	}
	salonID := *user.SalonID

	firstOfMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())

	var plan float64
	r.db.Model(&models.Goal{}).
		Where("assignee_id = ? AND target_date >= ?", userID, firstOfMonth).
		Select("COALESCE(SUM(sales_plan), 0)").Scan(&plan)

	var fact float64
	r.db.Model(&models.Lead{}).
		Where("salon_id = ? AND status = ? AND created_at >= ?", salonID, "sale", firstOfMonth).
		Select("COALESCE(SUM(budget), 0)").Scan(&fact)

	resp.Plan = plan
	resp.Fact = fact
	resp.PlanPercent = calcPercent(plan, fact).Percent

	return resp, nil
}

func (r *AnalyticsRepository) CalculateDashboardFunnel(ctx context.Context, userID uuid.UUID, dateStr string) (*models.DashboardFunnelResponse, error) {
	return &models.DashboardFunnelResponse{Stages: []models.FunnelStage{}}, nil
}

func (r *AnalyticsRepository) CalculateTerritoryPlanFact(ctx context.Context, userID uuid.UUID, period string) (*models.TerritoryPlanFactResponse, error) {
	return &models.TerritoryPlanFactResponse{}, nil
}

func (r *AnalyticsRepository) CalculateManagerTargets(ctx context.Context, userID uuid.UUID, dateStr string) (*models.ManagerTargetsResponse, error) {
	return &models.ManagerTargetsResponse{}, nil
}

var _ AnalyticsRepositoryInterface = (*AnalyticsRepository)(nil)