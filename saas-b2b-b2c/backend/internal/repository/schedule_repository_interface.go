package repository

import (
	"context"
	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
)

type ScheduleRepositoryInterface interface {
	CreateEvent(ctx context.Context, event *models.ScheduleEvent) error
	GetUserEventsByDate(ctx context.Context, userID uuid.UUID, dateStr string) ([]models.ScheduleEvent, error)
	UpdateEventStatus(ctx context.Context, eventID uuid.UUID, status string) error
	GetEventsByUsers(ctx context.Context, userIDs []uuid.UUID, dateStr string) ([]models.ScheduleEvent, error)
	UpdateEvent(ctx context.Context, eventID uuid.UUID, updates map[string]interface{}) error
	DeleteEvent(ctx context.Context, eventID uuid.UUID) error
}

var _ ScheduleRepositoryInterface = (*ScheduleRepository)(nil)