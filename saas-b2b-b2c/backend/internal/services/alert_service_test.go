package services

import (
	"context"
	"errors"
	"testing"

	"franchise-saas-backend/internal/mocks"
	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAlertService_CreateAlert(t *testing.T) {
	userID := uuid.New()

	t.Run("successful alert creation", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("CreateAlert", mock.Anything, mock.MatchedBy(func(a *models.Alert) bool {
			return a.Title == "Test Alert" && a.Description == "This is a test message" && a.Severity == models.AlertSeverityWarning
		})).Return(&models.Alert{
			ID:          uuid.New(),
			UserID:      userID,
			Title:       "Test Alert",
			Description: "This is a test message",
			Severity:    models.AlertSeverityWarning,
		}, nil)

		alert, err := service.CreateAlert(context.Background(), userID, "Test Alert", "This is a test message", models.AlertSeverityWarning, models.AlertTypeOverdueMeasurement, "/test")

		require.NoError(t, err)
		require.NotNil(t, alert)
		alertRepo.AssertExpectations(t)
	})

	t.Run("empty message validation error", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alert, err := service.CreateAlert(context.Background(), userID, "Test Alert", "", models.AlertSeverityWarning, models.AlertTypeOverdueMeasurement, "")

		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidMessage)
		require.Nil(t, alert)
	})

	t.Run("invalid severity error", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alert, err := service.CreateAlert(context.Background(), userID, "Test Alert", "Test message", models.AlertSeverity("invalid"), models.AlertTypeOverdueMeasurement, "")

		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidSeverity)
		require.Nil(t, alert)
	})

	t.Run("critical severity accepted", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("CreateAlert", mock.Anything, mock.Anything).Return(&models.Alert{
			ID:       uuid.New(),
			UserID:   userID,
			Severity: models.AlertSeverityCritical,
		}, nil)

		alert, err := service.CreateAlert(context.Background(), userID, "Critical Alert", "Critical message", models.AlertSeverityCritical, models.AlertTypeConversionDrop, "/critical")

		require.NoError(t, err)
		require.NotNil(t, alert)
		alertRepo.AssertExpectations(t)
	})

	t.Run("info severity accepted", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("CreateAlert", mock.Anything, mock.Anything).Return(&models.Alert{
			ID:       uuid.New(),
			UserID:   userID,
			Severity: models.AlertSeverityInfo,
		}, nil)

		alert, err := service.CreateAlert(context.Background(), userID, "Info Alert", "Info message", models.AlertSeverityInfo, models.AlertTypeTrafficDrop, "")

		require.NoError(t, err)
		require.NotNil(t, alert)
		alertRepo.AssertExpectations(t)
	})
}

func TestAlertService_GetAlertsByUser(t *testing.T) {
	userID := uuid.New()

	t.Run("success returns array of alerts", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alerts := []models.Alert{
			{ID: uuid.New(), UserID: userID, Title: "Alert 1"},
			{ID: uuid.New(), UserID: userID, Title: "Alert 2"},
			{ID: uuid.New(), UserID: userID, Title: "Alert 3"},
		}
		alertRepo.On("GetAlertsByUser", mock.Anything, userID).Return(alerts, nil)

		result, err := service.GetAlertsByUser(context.Background(), userID)

		require.NoError(t, err)
		require.Len(t, result, 3)
		alertRepo.AssertExpectations(t)
	})

	t.Run("empty list returns empty array", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetAlertsByUser", mock.Anything, userID).Return([]models.Alert{}, nil)

		result, err := service.GetAlertsByUser(context.Background(), userID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result, 0)
		alertRepo.AssertExpectations(t)
	})

	t.Run("repository returns error", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetAlertsByUser", mock.Anything, userID).Return(nil, errors.New("db error"))

		result, err := service.GetAlertsByUser(context.Background(), userID)

		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "db error", err.Error())
		alertRepo.AssertExpectations(t)
	})
}

func TestAlertService_MarkAsRead(t *testing.T) {
	userID := uuid.New()
	alertID := uuid.New()

	t.Run("success marks alert as read", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetAlertByID", mock.Anything, alertID).Return(&models.Alert{
			ID:     alertID,
			UserID: userID,
		}, nil)
		alertRepo.On("MarkAsRead", mock.Anything, alertID).Return(nil)

		err := service.MarkAsRead(context.Background(), userID, alertID)

		require.NoError(t, err)
		alertRepo.AssertExpectations(t)
	})

	t.Run("forbidden when user does not own alert", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)
		otherUserID := uuid.New()

		alertRepo.On("GetAlertByID", mock.Anything, alertID).Return(&models.Alert{
			ID:     alertID,
			UserID: otherUserID,
		}, nil)

		err := service.MarkAsRead(context.Background(), userID, alertID)

		require.Error(t, err)
		require.Equal(t, ErrForbidden, err)
		alertRepo.AssertExpectations(t)
	})

	t.Run("alert not found returns error", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetAlertByID", mock.Anything, alertID).Return(nil, errors.New("not found"))

		err := service.MarkAsRead(context.Background(), userID, alertID)

		require.Error(t, err)
		require.Equal(t, "not found", err.Error())
		alertRepo.AssertExpectations(t)
	})

	t.Run("mark read repository error", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetAlertByID", mock.Anything, alertID).Return(&models.Alert{
			ID:     alertID,
			UserID: userID,
		}, nil)
		alertRepo.On("MarkAsRead", mock.Anything, alertID).Return(errors.New("db error"))

		err := service.MarkAsRead(context.Background(), userID, alertID)

		require.Error(t, err)
		require.Equal(t, "db error", err.Error())
		alertRepo.AssertExpectations(t)
	})

	t.Run("get alert returns nil returns error", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetAlertByID", mock.Anything, alertID).Return(nil, nil)

		err := service.MarkAsRead(context.Background(), userID, alertID)

		require.Error(t, err)
		require.Equal(t, "alert not found", err.Error())
		alertRepo.AssertExpectations(t)
	})
}

func TestAlertService_MarkAllAsRead(t *testing.T) {
	userID := uuid.New()

	t.Run("success marks all unread as read", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetUnreadAlerts", mock.Anything, userID).Return([]models.Alert{
			{ID: uuid.New(), UserID: userID},
			{ID: uuid.New(), UserID: userID},
		}, nil)
		alertRepo.On("MarkAllAsRead", mock.Anything, userID).Return(nil)

		err := service.MarkAllAsRead(context.Background(), userID)

		require.NoError(t, err)
		alertRepo.AssertExpectations(t)
	})

	t.Run("no unread alerts does not fail", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetUnreadAlerts", mock.Anything, userID).Return([]models.Alert{}, nil)

		err := service.MarkAllAsRead(context.Background(), userID)

		require.NoError(t, err)
		alertRepo.AssertExpectations(t)
	})

	t.Run("repository error on get unread", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetUnreadAlerts", mock.Anything, userID).Return(nil, errors.New("db error"))

		err := service.MarkAllAsRead(context.Background(), userID)

		require.Error(t, err)
		require.Equal(t, "db error", err.Error())
		alertRepo.AssertExpectations(t)
	})

	t.Run("repository error on mark all read", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetUnreadAlerts", mock.Anything, userID).Return([]models.Alert{{ID: uuid.New()}}, nil)
		alertRepo.On("MarkAllAsRead", mock.Anything, userID).Return(errors.New("update failed"))

		err := service.MarkAllAsRead(context.Background(), userID)

		require.Error(t, err)
		require.Equal(t, "update failed", err.Error())
		alertRepo.AssertExpectations(t)
	})
}

func TestAlertService_GetUnreadCount(t *testing.T) {
	userID := uuid.New()

	t.Run("returns count when there are unread alerts", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alerts := []models.Alert{
			{ID: uuid.New()},
			{ID: uuid.New()},
			{ID: uuid.New()},
		}
		alertRepo.On("GetUnreadAlerts", mock.Anything, userID).Return(alerts, nil)

		count, err := service.GetUnreadCount(context.Background(), userID)

		require.NoError(t, err)
		require.Equal(t, 3, count)
		alertRepo.AssertExpectations(t)
	})

	t.Run("returns 0 when no unread alerts", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetUnreadAlerts", mock.Anything, userID).Return([]models.Alert{}, nil)

		count, err := service.GetUnreadCount(context.Background(), userID)

		require.NoError(t, err)
		require.Equal(t, 0, count)
		alertRepo.AssertExpectations(t)
	})

	t.Run("returns 0 when nil returned", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetUnreadAlerts", mock.Anything, userID).Return(nil, nil)

		count, err := service.GetUnreadCount(context.Background(), userID)

		require.NoError(t, err)
		require.Equal(t, 0, count)
		alertRepo.AssertExpectations(t)
	})

	t.Run("returns error on repository failure", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetUnreadAlerts", mock.Anything, userID).Return(nil, errors.New("db error"))

		count, err := service.GetUnreadCount(context.Background(), userID)

		require.Error(t, err)
		require.Equal(t, 0, count)
		require.Equal(t, "db error", err.Error())
		alertRepo.AssertExpectations(t)
	})
}

func TestAlertService_DeleteAlert(t *testing.T) {
	userID := uuid.New()
	alertID := uuid.New()

	t.Run("success deletes own alert", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetAlertByID", mock.Anything, alertID).Return(&models.Alert{
			ID:     alertID,
			UserID: userID,
		}, nil)
		alertRepo.On("DeleteAlert", mock.Anything, alertID).Return(nil)

		err := service.DeleteAlert(context.Background(), userID, alertID)

		require.NoError(t, err)
		alertRepo.AssertExpectations(t)
	})

	t.Run("forbidden when deleting someone elses alert", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)
		otherUserID := uuid.New()

		alertRepo.On("GetAlertByID", mock.Anything, alertID).Return(&models.Alert{
			ID:     alertID,
			UserID: otherUserID,
		}, nil)

		err := service.DeleteAlert(context.Background(), userID, alertID)

		require.Error(t, err)
		require.Equal(t, ErrForbidden, err)
		alertRepo.AssertExpectations(t)
	})

	t.Run("returns error when alert not found", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetAlertByID", mock.Anything, alertID).Return(nil, errors.New("not found"))

		err := service.DeleteAlert(context.Background(), userID, alertID)

		require.Error(t, err)
		require.Equal(t, "not found", err.Error())
		alertRepo.AssertExpectations(t)
	})

	t.Run("get alert returns nil returns error", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetAlertByID", mock.Anything, alertID).Return(nil, nil)

		err := service.DeleteAlert(context.Background(), userID, alertID)

		require.Error(t, err)
		require.Equal(t, "alert not found", err.Error())
		alertRepo.AssertExpectations(t)
	})

	t.Run("delete repository error", func(t *testing.T) {
		alertRepo := mocks.NewMockAlertRepository()
		service := NewAlertServiceWithDB(alertRepo, nil)

		alertRepo.On("GetAlertByID", mock.Anything, alertID).Return(&models.Alert{
			ID:     alertID,
			UserID: userID,
		}, nil)
		alertRepo.On("DeleteAlert", mock.Anything, alertID).Return(errors.New("delete failed"))

		err := service.DeleteAlert(context.Background(), userID, alertID)

		require.Error(t, err)
		require.Equal(t, "delete failed", err.Error())
		alertRepo.AssertExpectations(t)
	})
}