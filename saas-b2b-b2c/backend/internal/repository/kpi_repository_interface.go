package repository

import (
	"context"
	"franchise-saas-backend/internal/models"
	"time"

	"github.com/google/uuid"
)

type KPIRepositoryInterface interface {
	UpsertGoal(ctx context.Context, goal *models.DailyGoal) error
	GetGoal(ctx context.Context, userID *uuid.UUID, salonID *uuid.UUID, date time.Time) (*models.DailyGoal, error)
}

var _ KPIRepositoryInterface = (*KPIRepository)(nil)