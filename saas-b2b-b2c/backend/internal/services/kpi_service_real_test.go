package services

import (
	"context"
	"testing"
	"time"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockKPIRepositoryForTest struct {
	mock.Mock
}

func (m *MockKPIRepositoryForTest) UpsertGoal(ctx context.Context, goal *models.DailyGoal) error {
	args := m.Called(ctx, goal)
	return args.Error(0)
}

func (m *MockKPIRepositoryForTest) GetGoal(ctx context.Context, userID *uuid.UUID, salonID *uuid.UUID, date time.Time) (*models.DailyGoal, error) {
	args := m.Called(ctx, userID, salonID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DailyGoal), args.Error(1)
}

type MockScheduleRepositoryForTest struct {
	mock.Mock
}

func (m *MockScheduleRepositoryForTest) CreateEvent(ctx context.Context, event *models.ScheduleEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockScheduleRepositoryForTest) GetUserEventsByDate(ctx context.Context, userID uuid.UUID, dateStr string) ([]models.ScheduleEvent, error) {
	args := m.Called(ctx, userID, dateStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ScheduleEvent), args.Error(1)
}

func (m *MockScheduleRepositoryForTest) UpdateEventStatus(ctx context.Context, eventID uuid.UUID, status string) error {
	args := m.Called(ctx, eventID, status)
	return args.Error(0)
}

func (m *MockScheduleRepositoryForTest) GetEventsByUsers(ctx context.Context, userIDs []uuid.UUID, dateStr string) ([]models.ScheduleEvent, error) {
	args := m.Called(ctx, userIDs, dateStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ScheduleEvent), args.Error(1)
}

func (m *MockScheduleRepositoryForTest) UpdateEvent(ctx context.Context, eventID uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(ctx, eventID, updates)
	return args.Error(0)
}

func (m *MockScheduleRepositoryForTest) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

func TestKPIService_SetGoal_WithRealService(t *testing.T) {
	mockKpi := new(MockKPIRepositoryForTest)
	mockSched := new(MockScheduleRepositoryForTest)
	
	// Используем nil для DB - сервис будет работать без него
	service := NewKPIService(nil, mockKpi, mockSched)

	goal := &models.DailyGoal{
		UserID:      ptrToUUIDPtr(uuid.New()),
		SalesPlan:   100000,
		LeadsPlan:   20,
		CallsPlan:   50,
		MeetingsPlan: 10,
		TargetDate:  time.Now(),
	}

	mockKpi.On("UpsertGoal", mock.Anything, goal).Return(nil)

	err := service.SetGoal(context.Background(), goal)

	assert.NoError(t, err)
	mockKpi.AssertExpectations(t)
}

func TestKPIService_SetGoal_Error(t *testing.T) {
	mockKpi := new(MockKPIRepositoryForTest)
	mockSched := new(MockScheduleRepositoryForTest)
	
	service := NewKPIService(nil, mockKpi, mockSched)

	goal := &models.DailyGoal{
		UserID:      ptrToUUIDPtr(uuid.New()),
		SalesPlan:   100000,
		TargetDate:  time.Now(),
	}

	mockKpi.On("UpsertGoal", mock.Anything, mock.Anything).Return(assert.AnError)

	err := service.SetGoal(context.Background(), goal)

	assert.Error(t, err)
	mockKpi.AssertExpectations(t)
}

func ptrToUUIDPtr(id uuid.UUID) *uuid.UUID {
	return &id
}