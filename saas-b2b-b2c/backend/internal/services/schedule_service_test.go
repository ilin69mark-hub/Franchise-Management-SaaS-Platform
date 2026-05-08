package services

import (
	"context"
	"errors"
	"testing"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockScheduleRepo struct {
	mock.Mock
}

func NewMockScheduleRepo() *MockScheduleRepo {
	return &MockScheduleRepo{}
}

func (m *MockScheduleRepo) GetUserEventsByDate(ctx context.Context, userID uuid.UUID, dateStr string) ([]models.ScheduleEvent, error) {
	args := m.Called(ctx, userID, dateStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ScheduleEvent), args.Error(1)
}

func (m *MockScheduleRepo) CreateEvent(ctx context.Context, event *models.ScheduleEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockScheduleRepo) UpdateEventStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockScheduleRepo) GetEventsByUsers(ctx context.Context, userIDs []uuid.UUID, dateStr string) ([]models.ScheduleEvent, error) {
	args := m.Called(ctx, userIDs, dateStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ScheduleEvent), args.Error(1)
}

func (m *MockScheduleRepo) UpdateEvent(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockScheduleRepo) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type ScheduleRepoInterface interface {
	GetUserEventsByDate(ctx context.Context, userID uuid.UUID, dateStr string) ([]models.ScheduleEvent, error)
	CreateEvent(ctx context.Context, event *models.ScheduleEvent) error
	UpdateEventStatus(ctx context.Context, id uuid.UUID, status string) error
	GetEventsByUsers(ctx context.Context, userIDs []uuid.UUID, dateStr string) ([]models.ScheduleEvent, error)
	UpdateEvent(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	DeleteEvent(ctx context.Context, id uuid.UUID) error
}

type ScheduleServiceTestable struct {
	repo ScheduleRepoInterface
}

func NewScheduleServiceTestable(repo ScheduleRepoInterface) *ScheduleServiceTestable {
	return &ScheduleServiceTestable{repo: repo}
}

func (s *ScheduleServiceTestable) GetUserSchedule(ctx context.Context, userID uuid.UUID, dateStr string) ([]models.ScheduleEvent, error) {
	return s.repo.GetUserEventsByDate(ctx, userID, dateStr)
}

func (s *ScheduleServiceTestable) CreateEvent(ctx context.Context, event *models.ScheduleEvent) error {
	return s.repo.CreateEvent(ctx, event)
}

func (s *ScheduleServiceTestable) UpdateEventStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.repo.UpdateEventStatus(ctx, id, status)
}

func (s *ScheduleServiceTestable) GetEventsByUsers(ctx context.Context, userIDs []uuid.UUID, dateStr string) ([]models.ScheduleEvent, error) {
	return s.repo.GetEventsByUsers(ctx, userIDs, dateStr)
}

func (s *ScheduleServiceTestable) UpdateEvent(ctx context.Context, id uuid.UUID, data *models.UpdateScheduleEventRequest) error {
	updates := map[string]interface{}{}
	if data.Title != "" {
		updates["title"] = data.Title
	}
	if data.Description != "" {
		updates["description"] = data.Description
	}
	if data.Status != "" {
		updates["status"] = data.Status
	}
	if data.Priority != "" {
		updates["priority"] = data.Priority
	}
	return s.repo.UpdateEvent(ctx, id, updates)
}

func (s *ScheduleServiceTestable) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteEvent(ctx, id)
}

func TestScheduleService_GetUserSchedule(t *testing.T) {
	userID := uuid.New()
	dateStr := "2026-05-07"

	t.Run("success returns events", func(t *testing.T) {
		repo := NewMockScheduleRepo()
		service := NewScheduleServiceTestable(repo)

		events := []models.ScheduleEvent{
			{ID: uuid.New(), UserID: userID, Title: "Meeting with client", Type: "meeting"},
			{ID: uuid.New(), UserID: userID, Title: "Call to lead", Type: "call"},
		}
		repo.On("GetUserEventsByDate", mock.Anything, userID, dateStr).Return(events, nil)

		result, err := service.GetUserSchedule(context.Background(), userID, dateStr)

		require.NoError(t, err)
		require.Len(t, result, 2)
		require.Equal(t, "Meeting with client", result[0].Title)
		repo.AssertExpectations(t)
	})

	t.Run("empty schedule returns empty array", func(t *testing.T) {
		repo := NewMockScheduleRepo()
		service := NewScheduleServiceTestable(repo)

		repo.On("GetUserEventsByDate", mock.Anything, userID, dateStr).Return([]models.ScheduleEvent{}, nil)

		result, err := service.GetUserSchedule(context.Background(), userID, dateStr)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result, 0)
		repo.AssertExpectations(t)
	})

	t.Run("repository error returns error", func(t *testing.T) {
		repo := NewMockScheduleRepo()
		service := NewScheduleServiceTestable(repo)

		repo.On("GetUserEventsByDate", mock.Anything, userID, dateStr).Return(nil, errors.New("db error"))

		result, err := service.GetUserSchedule(context.Background(), userID, dateStr)

		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "db error", err.Error())
		repo.AssertExpectations(t)
	})
}

func TestScheduleService_CreateEvent(t *testing.T) {
	t.Run("success creates event", func(t *testing.T) {
		repo := NewMockScheduleRepo()
		service := NewScheduleServiceTestable(repo)

		event := &models.ScheduleEvent{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			Title:       "New Task",
			Description: "Task description",
			Type:        "task",
		}
		repo.On("CreateEvent", mock.Anything, event).Return(nil)

		err := service.CreateEvent(context.Background(), event)

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("repository error on create", func(t *testing.T) {
		repo := NewMockScheduleRepo()
		service := NewScheduleServiceTestable(repo)

		event := &models.ScheduleEvent{
			ID:     uuid.New(),
			UserID: uuid.New(),
			Title:  "Task",
		}
		repo.On("CreateEvent", mock.Anything, event).Return(errors.New("constraint violation"))

		err := service.CreateEvent(context.Background(), event)

		require.Error(t, err)
		require.Equal(t, "constraint violation", err.Error())
		repo.AssertExpectations(t)
	})
}

func TestScheduleService_UpdateEventStatus(t *testing.T) {
	eventID := uuid.New()

	t.Run("success updates status", func(t *testing.T) {
		repo := NewMockScheduleRepo()
		service := NewScheduleServiceTestable(repo)

		repo.On("UpdateEventStatus", mock.Anything, eventID, "completed").Return(nil)

		err := service.UpdateEventStatus(context.Background(), eventID, "completed")

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("update to cancelled status", func(t *testing.T) {
		repo := NewMockScheduleRepo()
		service := NewScheduleServiceTestable(repo)

		repo.On("UpdateEventStatus", mock.Anything, eventID, "cancelled").Return(nil)

		err := service.UpdateEventStatus(context.Background(), eventID, "cancelled")

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

func TestScheduleService_GetEventsByUsers(t *testing.T) {
	userIDs := []uuid.UUID{uuid.New(), uuid.New()}
	dateStr := "2026-05-07"

	t.Run("success returns combined events", func(t *testing.T) {
		repo := NewMockScheduleRepo()
		service := NewScheduleServiceTestable(repo)

		events := []models.ScheduleEvent{
			{ID: uuid.New(), UserID: userIDs[0], Title: "Event 1"},
			{ID: uuid.New(), UserID: userIDs[1], Title: "Event 2"},
		}
		repo.On("GetEventsByUsers", mock.Anything, userIDs, dateStr).Return(events, nil)

		result, err := service.GetEventsByUsers(context.Background(), userIDs, dateStr)

		require.NoError(t, err)
		require.Len(t, result, 2)
		repo.AssertExpectations(t)
	})

	t.Run("empty result for no events", func(t *testing.T) {
		repo := NewMockScheduleRepo()
		service := NewScheduleServiceTestable(repo)

		repo.On("GetEventsByUsers", mock.Anything, userIDs, dateStr).Return([]models.ScheduleEvent{}, nil)

		result, err := service.GetEventsByUsers(context.Background(), userIDs, dateStr)

		require.NoError(t, err)
		require.Len(t, result, 0)
		repo.AssertExpectations(t)
	})
}

func TestScheduleService_UpdateEvent(t *testing.T) {
	eventID := uuid.New()

	t.Run("success updates title", func(t *testing.T) {
		repo := NewMockScheduleRepo()
		service := NewScheduleServiceTestable(repo)

		repo.On("UpdateEvent", mock.Anything, eventID, mock.Anything).Return(nil)

		err := service.UpdateEvent(context.Background(), eventID, &models.UpdateScheduleEventRequest{
			Title: "Updated Title",
		})

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("success updates multiple fields", func(t *testing.T) {
		repo := NewMockScheduleRepo()
		service := NewScheduleServiceTestable(repo)

		repo.On("UpdateEvent", mock.Anything, eventID, mock.MatchedBy(func(m map[string]interface{}) bool {
			return m["title"] == "New Title" && m["status"] == "in_progress"
		})).Return(nil)

		err := service.UpdateEvent(context.Background(), eventID, &models.UpdateScheduleEventRequest{
			Title:  "New Title",
			Status: "in_progress",
		})

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("repository error returns error", func(t *testing.T) {
		repo := NewMockScheduleRepo()
		service := NewScheduleServiceTestable(repo)

		repo.On("UpdateEvent", mock.Anything, eventID, mock.Anything).Return(errors.New("not found"))

		err := service.UpdateEvent(context.Background(), eventID, &models.UpdateScheduleEventRequest{
			Title: "Test",
		})

		require.Error(t, err)
		require.Equal(t, "not found", err.Error())
		repo.AssertExpectations(t)
	})
}

func TestScheduleService_DeleteEvent(t *testing.T) {
	eventID := uuid.New()

	t.Run("success deletes event", func(t *testing.T) {
		repo := NewMockScheduleRepo()
		service := NewScheduleServiceTestable(repo)

		repo.On("DeleteEvent", mock.Anything, eventID).Return(nil)

		err := service.DeleteEvent(context.Background(), eventID)

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("delete non-existent event returns error", func(t *testing.T) {
		repo := NewMockScheduleRepo()
		service := NewScheduleServiceTestable(repo)

		repo.On("DeleteEvent", mock.Anything, eventID).Return(errors.New("record not found"))

		err := service.DeleteEvent(context.Background(), eventID)

		require.Error(t, err)
		require.Equal(t, "record not found", err.Error())
		repo.AssertExpectations(t)
	})
}