package repository

import (
	"context"
	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
)

type LeadRepositoryInterface interface {
	CreateLead(ctx context.Context, lead *models.Lead) error
	GetLeadsByManager(ctx context.Context, managerID uuid.UUID) ([]models.Lead, error)
	GetLeadByID(ctx context.Context, leadID, managerID uuid.UUID) (*models.Lead, error)
	UpdateLeadStatus(ctx context.Context, leadID uuid.UUID, status string) error
	AddActivity(ctx context.Context, activity *models.LeadActivity) error
	GetLeadActivities(ctx context.Context, leadID uuid.UUID) ([]models.LeadActivity, error)
}

var _ LeadRepositoryInterface = (*LeadRepository)(nil)