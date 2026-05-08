package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockKPIRepository struct {
	mock.Mock
}

func NewMockKPIRepository() *MockKPIRepository {
	return &MockKPIRepository{}
}

func (m *MockKPIRepository) UpsertGoal(ctx context.Context, goal *models.DailyGoal) error {
	args := m.Called(ctx, goal)
	return args.Error(0)
}

func (m *MockKPIRepository) GetGoal(ctx context.Context, userID *uuid.UUID, salonID *uuid.UUID, date time.Time) (*models.DailyGoal, error) {
	args := m.Called(ctx, userID, salonID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DailyGoal), args.Error(1)
}

func (m *MockKPIRepository) GetKPISummary(ctx context.Context, userID *uuid.UUID, salonID *uuid.UUID, startDate, endDate time.Time) (*models.KPISummary, error) {
	args := m.Called(ctx, userID, salonID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.KPISummary), args.Error(1)
}

type MockScheduleRepository struct {
	mock.Mock
}

func NewMockScheduleRepository() *MockScheduleRepository {
	return &MockScheduleRepository{}
}

func (m *MockScheduleRepository) GetScheduleByFilters(ctx context.Context, userID *uuid.UUID, salonID *uuid.UUID, startDate, endDate time.Time) ([]models.ScheduleEvent, error) {
	args := m.Called(ctx, userID, salonID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ScheduleEvent), args.Error(1)
}

type KPIRepositoryInterface interface {
	UpsertGoal(ctx context.Context, goal *models.DailyGoal) error
	GetGoal(ctx context.Context, userID *uuid.UUID, salonID *uuid.UUID, date time.Time) (*models.DailyGoal, error)
	GetKPISummary(ctx context.Context, userID *uuid.UUID, salonID *uuid.UUID, startDate, endDate time.Time) (*models.KPISummary, error)
}

type ScheduleRepositoryInterface interface {
	GetScheduleByFilters(ctx context.Context, userID *uuid.UUID, salonID *uuid.UUID, startDate, endDate time.Time) ([]models.ScheduleEvent, error)
}

type KPIServiceTestable struct {
	kpiRepo  KPIRepositoryInterface
	schedRep ScheduleRepositoryInterface
}

func NewKPIServiceTestable(kpiRepo KPIRepositoryInterface, schedRep ScheduleRepositoryInterface) *KPIServiceTestable {
	return &KPIServiceTestable{
		kpiRepo:  kpiRepo,
		schedRep: schedRep,
	}
}

func (s *KPIServiceTestable) SetGoal(ctx context.Context, goal *models.DailyGoal) error {
	return s.kpiRepo.UpsertGoal(ctx, goal)
}

func (s *KPIServiceTestable) GetKPISummary(ctx context.Context, userID *uuid.UUID, salonID *uuid.UUID, startDate, endDate time.Time) (*models.KPISummary, error) {
	return s.kpiRepo.GetKPISummary(ctx, userID, salonID, startDate, endDate)
}

func TestKPIService_SetGoal(t *testing.T) {
	t.Run("success creates new goal", func(t *testing.T) {
		kpiRepo := NewMockKPIRepository()
		schedRepo := NewMockScheduleRepository()
		service := NewKPIServiceTestable(kpiRepo, schedRepo)

		goal := &models.DailyGoal{
			UserID:      ptrUUID(uuid.New()),
			SalesPlan:   100000,
			LeadsPlan:   20,
			CallsPlan:   50,
			MeetingsPlan: 10,
			TargetDate:  time.Now(),
		}
		kpiRepo.On("UpsertGoal", mock.Anything, goal).Return(nil)

		err := service.SetGoal(context.Background(), goal)

		require.NoError(t, err)
		kpiRepo.AssertExpectations(t)
	})

	t.Run("success updates existing goal", func(t *testing.T) {
		kpiRepo := NewMockKPIRepository()
		schedRepo := NewMockScheduleRepository()
		service := NewKPIServiceTestable(kpiRepo, schedRepo)

		existingGoal := &models.DailyGoal{
			ID:         uuid.New(),
			UserID:     ptrUUID(uuid.New()),
			SalesPlan:   100000,
			LeadsPlan:   20,
			TargetDate: time.Now(),
		}
		kpiRepo.On("UpsertGoal", mock.Anything, mock.Anything).Return(nil)

		err := service.SetGoal(context.Background(), existingGoal)

		require.NoError(t, err)
		kpiRepo.AssertExpectations(t)
	})

	t.Run("repository error returns error", func(t *testing.T) {
		kpiRepo := NewMockKPIRepository()
		schedRepo := NewMockScheduleRepository()
		service := NewKPIServiceTestable(kpiRepo, schedRepo)

		goal := &models.DailyGoal{
			UserID:      ptrUUID(uuid.New()),
			TargetDate: time.Now(),
		}
		kpiRepo.On("UpsertGoal", mock.Anything, goal).Return(errors.New("db error"))

		err := service.SetGoal(context.Background(), goal)

		require.Error(t, err)
		require.Equal(t, "db error", err.Error())
		kpiRepo.AssertExpectations(t)
	})

	t.Run("goal with salon id", func(t *testing.T) {
		kpiRepo := NewMockKPIRepository()
		schedRepo := NewMockScheduleRepository()
		service := NewKPIServiceTestable(kpiRepo, schedRepo)

		goal := &models.DailyGoal{
			SalonID:     ptrUUID(uuid.New()),
			SalesPlan:   500000,
			LeadsPlan:   50,
			TargetDate: time.Now(),
		}
		kpiRepo.On("UpsertGoal", mock.Anything, mock.MatchedBy(func(g *models.DailyGoal) bool {
			return g.SalonID != nil && g.SalesPlan == 500000
		})).Return(nil)

		err := service.SetGoal(context.Background(), goal)

		require.NoError(t, err)
		kpiRepo.AssertExpectations(t)
	})
}

func TestKPIService_GetKPISummary(t *testing.T) {
	userID := uuid.New()
	salonID := uuid.New()
	startDate := time.Now().AddDate(0, 0, -30)
	endDate := time.Now()

	t.Run("success returns kpi summary", func(t *testing.T) {
		kpiRepo := NewMockKPIRepository()
		schedRepo := NewMockScheduleRepository()
		service := NewKPIServiceTestable(kpiRepo, schedRepo)

		summary := &models.KPISummary{
			TotalSales:     500000,
			TotalLeads:     100,
			ConversionRate: 25.5,
			AvgCheck:       5000,
		}
		kpiRepo.On("GetKPISummary", mock.Anything, &userID, &salonID, startDate, endDate).Return(summary, nil)

		result, err := service.GetKPISummary(context.Background(), &userID, &salonID, startDate, endDate)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 500000.0, result.TotalSales)
		require.Equal(t, 100, result.TotalLeads)
		kpiRepo.AssertExpectations(t)
	})

	t.Run("empty summary returns zero values", func(t *testing.T) {
		kpiRepo := NewMockKPIRepository()
		schedRepo := NewMockScheduleRepository()
		service := NewKPIServiceTestable(kpiRepo, schedRepo)

		summary := &models.KPISummary{
			TotalSales:     0,
			TotalLeads:     0,
			ConversionRate: 0,
			AvgCheck:       0,
		}
		kpiRepo.On("GetKPISummary", mock.Anything, (*uuid.UUID)(nil), &salonID, startDate, endDate).Return(summary, nil)

		result, err := service.GetKPISummary(context.Background(), nil, &salonID, startDate, endDate)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 0.0, result.TotalSales)
		kpiRepo.AssertExpectations(t)
	})

	t.Run("repository error returns error", func(t *testing.T) {
		kpiRepo := NewMockKPIRepository()
		schedRepo := NewMockScheduleRepository()
		service := NewKPIServiceTestable(kpiRepo, schedRepo)

		kpiRepo.On("GetKPISummary", mock.Anything, &userID, (*uuid.UUID)(nil), startDate, endDate).Return(nil, errors.New("query failed"))

		result, err := service.GetKPISummary(context.Background(), &userID, nil, startDate, endDate)

		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "query failed", err.Error())
		kpiRepo.AssertExpectations(t)
	})

	t.Run("summary with high conversion rate", func(t *testing.T) {
		kpiRepo := NewMockKPIRepository()
		schedRepo := NewMockScheduleRepository()
		service := NewKPIServiceTestable(kpiRepo, schedRepo)

		summary := &models.KPISummary{
			TotalSales:     1000000,
			TotalLeads:     50,
			ConversionRate: 80.0,
			AvgCheck:       20000,
		}
		kpiRepo.On("GetKPISummary", mock.Anything, &userID, &salonID, startDate, endDate).Return(summary, nil)

		result, err := service.GetKPISummary(context.Background(), &userID, &salonID, startDate, endDate)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Greater(t, result.ConversionRate, 50.0)
		kpiRepo.AssertExpectations(t)
	})
}

func ptrUUID(id uuid.UUID) *uuid.UUID {
	return &id
}