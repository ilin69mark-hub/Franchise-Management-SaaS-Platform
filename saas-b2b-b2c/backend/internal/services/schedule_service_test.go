package services

import (
	"context"
	"errors"
	"testing"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockScheduleRepo struct {
	mock.Mock
}

func (m *MockScheduleRepo) CreateEvent(ctx context.Context, event *models.ScheduleEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockScheduleRepo) GetUserEventsByDate(ctx context.Context, userID uuid.UUID, dateStr string) ([]models.ScheduleEvent, error) {
	args := m.Called(ctx, userID, dateStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ScheduleEvent), args.Error(1)
}

func (m *MockScheduleRepo) UpdateEventStatus(ctx context.Context, eventID uuid.UUID, status string) error {
	args := m.Called(ctx, eventID, status)
	return args.Error(0)
}

func (m *MockScheduleRepo) GetEventsByUsers(ctx context.Context, userIDs []uuid.UUID, dateStr string) ([]models.ScheduleEvent, error) {
	args := m.Called(ctx, userIDs, dateStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ScheduleEvent), args.Error(1)
}

func (m *MockScheduleRepo) UpdateEvent(ctx context.Context, eventID uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(ctx, eventID, updates)
	return args.Error(0)
}

func (m *MockScheduleRepo) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

func TestScheduleService_GetUserSchedule_Success(t *testing.T) {
	mockRepo := new(MockScheduleRepo)
	service := NewScheduleService(mockRepo)

	userID := uuid.New()
	expected := []models.ScheduleEvent{
		{ID: uuid.New(), Title: "Meeting"},
	}

	mockRepo.On("GetUserEventsByDate", mock.Anything, userID, "2024-01-15").Return(expected, nil)

	result, err := service.GetUserSchedule(context.Background(), userID, "2024-01-15")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockRepo.AssertExpectations(t)
}

func TestScheduleService_GetUserSchedule_Error(t *testing.T) {
	mockRepo := new(MockScheduleRepo)
	service := NewScheduleService(mockRepo)

	userID := uuid.New()

	mockRepo.On("GetUserEventsByDate", mock.Anything, userID, "2024-01-15").Return(nil, errors.New("db error"))

	result, err := service.GetUserSchedule(context.Background(), userID, "2024-01-15")

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestScheduleService_CreateEvent_Success(t *testing.T) {
	mockRepo := new(MockScheduleRepo)
	service := NewScheduleService(mockRepo)

	event := &models.ScheduleEvent{
		ID:     uuid.New(),
		Title:  "New Event",
		UserID: uuid.New(),
	}

	mockRepo.On("CreateEvent", mock.Anything, event).Return(nil)

	err := service.CreateEvent(context.Background(), event)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestScheduleService_CreateEvent_Error(t *testing.T) {
	mockRepo := new(MockScheduleRepo)
	service := NewScheduleService(mockRepo)

	event := &models.ScheduleEvent{ID: uuid.New(), Title: "New Event"}

	mockRepo.On("CreateEvent", mock.Anything, mock.Anything).Return(errors.New("db error"))

	err := service.CreateEvent(context.Background(), event)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestScheduleService_UpdateEventStatus_Success(t *testing.T) {
	mockRepo := new(MockScheduleRepo)
	service := NewScheduleService(mockRepo)

	eventID := uuid.New()

	mockRepo.On("UpdateEventStatus", mock.Anything, eventID, "completed").Return(nil)

	err := service.UpdateEventStatus(context.Background(), eventID, "completed")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestScheduleService_UpdateEventStatus_Error(t *testing.T) {
	mockRepo := new(MockScheduleRepo)
	service := NewScheduleService(mockRepo)

	eventID := uuid.New()

	mockRepo.On("UpdateEventStatus", mock.Anything, eventID, "completed").Return(errors.New("db error"))

	err := service.UpdateEventStatus(context.Background(), eventID, "completed")

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestScheduleService_GetEventsByUsers_Success(t *testing.T) {
	mockRepo := new(MockScheduleRepo)
	service := NewScheduleService(mockRepo)

	userIDs := []uuid.UUID{uuid.New(), uuid.New()}
	expected := []models.ScheduleEvent{
		{ID: uuid.New(), Title: "Event 1"},
		{ID: uuid.New(), Title: "Event 2"},
	}

	mockRepo.On("GetEventsByUsers", mock.Anything, userIDs, "2024-01-15").Return(expected, nil)

	result, err := service.GetEventsByUsers(context.Background(), userIDs, "2024-01-15")

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestScheduleService_GetEventsByUsers_Error(t *testing.T) {
	mockRepo := new(MockScheduleRepo)
	service := NewScheduleService(mockRepo)

	userIDs := []uuid.UUID{uuid.New()}

	mockRepo.On("GetEventsByUsers", mock.Anything, userIDs, "2024-01-15").Return(nil, errors.New("db error"))

	result, err := service.GetEventsByUsers(context.Background(), userIDs, "2024-01-15")

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestScheduleService_UpdateEvent_Success(t *testing.T) {
	mockRepo := new(MockScheduleRepo)
	service := NewScheduleService(mockRepo)

	eventID := uuid.New()
	updates := map[string]interface{}{"title": "Updated Title"}

	mockRepo.On("UpdateEvent", mock.Anything, eventID, updates).Return(nil)

	err := service.UpdateEvent(context.Background(), eventID, &models.UpdateScheduleEventRequest{Title: "Updated Title"})

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestScheduleService_UpdateEvent_Error(t *testing.T) {
	mockRepo := new(MockScheduleRepo)
	service := NewScheduleService(mockRepo)

	eventID := uuid.New()

	mockRepo.On("UpdateEvent", mock.Anything, eventID, mock.Anything).Return(errors.New("db error"))

	err := service.UpdateEvent(context.Background(), eventID, &models.UpdateScheduleEventRequest{Title: "Test"})

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestScheduleService_DeleteEvent_Success(t *testing.T) {
	mockRepo := new(MockScheduleRepo)
	service := NewScheduleService(mockRepo)

	eventID := uuid.New()

	mockRepo.On("DeleteEvent", mock.Anything, eventID).Return(nil)

	err := service.DeleteEvent(context.Background(), eventID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestScheduleService_DeleteEvent_Error(t *testing.T) {
	mockRepo := new(MockScheduleRepo)
	service := NewScheduleService(mockRepo)

	eventID := uuid.New()

	mockRepo.On("DeleteEvent", mock.Anything, eventID).Return(errors.New("db error"))

	err := service.DeleteEvent(context.Background(), eventID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}