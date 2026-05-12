package services

import (
	"context"
	"testing"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAnalyticsRepo struct {
	mock.Mock
}

func (m *MockAnalyticsRepo) CalculateDashboardStats(ctx context.Context, userID *uuid.UUID, salonID *uuid.UUID, isManager bool) (*models.DashboardStatsResponse, error) {
	args := m.Called(ctx, userID, salonID, isManager)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardStatsResponse), args.Error(1)
}

func (m *MockAnalyticsRepo) CalculateDashboardMain(ctx context.Context, userID uuid.UUID, dateStr string) (*models.DashboardMainResponse, error) {
	args := m.Called(ctx, userID, dateStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardMainResponse), args.Error(1)
}

func (m *MockAnalyticsRepo) CalculateDashboardFunnel(ctx context.Context, userID uuid.UUID, dateStr string) (*models.DashboardFunnelResponse, error) {
	args := m.Called(ctx, userID, dateStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DashboardFunnelResponse), args.Error(1)
}

func (m *MockAnalyticsRepo) CalculateTerritoryPlanFact(ctx context.Context, userID uuid.UUID, period string) (*models.TerritoryPlanFactResponse, error) {
	args := m.Called(ctx, userID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TerritoryPlanFactResponse), args.Error(1)
}

func (m *MockAnalyticsRepo) CalculateManagerTargets(ctx context.Context, userID uuid.UUID, dateStr string) (*models.ManagerTargetsResponse, error) {
	args := m.Called(ctx, userID, dateStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ManagerTargetsResponse), args.Error(1)
}

func TestKPIService_GetDashboardStats_WithAnalytics(t *testing.T) {
	mockKpi := new(MockKPIRepositoryForTest)
	mockSched := new(MockScheduleRepositoryForTest)
	mockAnalytics := new(MockAnalyticsRepo)

	service := NewKPIServiceWithAnalytics(nil, mockKpi, mockSched, mockAnalytics)

	userID := uuid.New()
	salonID := uuid.New()
	expected := &models.DashboardStatsResponse{Date: "2024-01-15"}

	mockAnalytics.On("CalculateDashboardStats", mock.Anything, &userID, &salonID, true).Return(expected, nil)

	resp, err := service.GetDashboardStats(context.Background(), userID, salonID, true)

	assert.NoError(t, err)
	assert.Equal(t, expected, resp)
	mockAnalytics.AssertExpectations(t)
}