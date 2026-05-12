package repository

import (
	"context"
	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
)

type ChecklistRepositoryInterface interface {
	FindAllGlobal(ctx context.Context, status, priority string, isArchive bool) ([]models.Checklist, error)
	FindUserTasks(ctx context.Context, userID uuid.UUID, status, priority string, isArchive bool) ([]models.Checklist, error)
	FindAll(ctx context.Context, status, priority string) ([]models.Checklist, error)
	GetChecklistByID(ctx context.Context, id uuid.UUID) (*models.Checklist, error)
	CreateChecklist(ctx context.Context, checklist *models.Checklist) error
	UpdateChecklist(ctx context.Context, checklist *models.Checklist) error
	DeleteChecklist(ctx context.Context, id uuid.UUID) error
	GetTaskStats() (total int64, completed int64, err error)
}

var _ ChecklistRepositoryInterface = (*ChecklistRepository)(nil)