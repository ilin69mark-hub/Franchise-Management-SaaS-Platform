package repository

import (
	"context"
	"time"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
)

type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	UpdateUserFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	FindUsersByTenantID(ctx context.Context, tenantID uuid.UUID) ([]models.User, error)
	FindAllGlobal(ctx context.Context) ([]models.User, error)
	FindUserByIDAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*models.User, error)
	CountByRole(ctx context.Context, role string) (int64, error)
}

type TenantRepositoryInterface interface {
	CreateTenant(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error)
	GetTenantByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	GetTenantBySlug(ctx context.Context, slug string) (*models.Tenant, error)
	UpdateTenant(ctx context.Context, tenant *models.Tenant) error
	DeleteTenant(ctx context.Context, id uuid.UUID) error
	FindAll(ctx context.Context) ([]models.Tenant, error)
	FindByStatus(ctx context.Context, status string) ([]models.Tenant, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	CountUsersByTenantID(ctx context.Context, tenantID uuid.UUID) (int64, error)
}

type AuditLogRepositoryInterface interface {
	FindByFilters(ctx context.Context, tenantID *uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]models.AuditLog, error)
	CountByFilters(ctx context.Context, tenantID *uuid.UUID, startDate, endDate *time.Time) (int64, error)
}

type GoalRepositoryInterface interface {
	CreateGoal(ctx context.Context, goal *models.Goal) (*models.Goal, error)
	GetGoalByID(ctx context.Context, id uuid.UUID) (*models.Goal, error)
	GetGoalsByUser(ctx context.Context, userID uuid.UUID) ([]models.Goal, error)
	GetGoalsByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.Goal, error)
	UpdateGoal(ctx context.Context, goal *models.Goal) error
	DeleteGoal(ctx context.Context, id uuid.UUID) error
}

type AlertRepositoryInterface interface {
	CreateAlert(ctx context.Context, alert *models.Alert) (*models.Alert, error)
	GetAlertByID(ctx context.Context, id uuid.UUID) (*models.Alert, error)
	GetAlertsByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.Alert, error)
	GetAlertsByUser(ctx context.Context, userID uuid.UUID) ([]models.Alert, error)
	GetUnreadAlerts(ctx context.Context, userID uuid.UUID) ([]models.Alert, error)
	MarkAsRead(ctx context.Context, alertID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	DeleteAlert(ctx context.Context, id uuid.UUID) error
}

type SalonRepositoryInterface interface {
	CreateSalon(ctx context.Context, salon *models.Salon) (*models.Salon, error)
	GetSalonByID(ctx context.Context, id uuid.UUID) (*models.Salon, error)
	GetSalonsByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.Salon, error)
	GetSalonsByDealer(ctx context.Context, dealerID uuid.UUID) ([]models.Salon, error)
	UpdateSalon(ctx context.Context, salon *models.Salon) error
	DeleteSalon(ctx context.Context, id uuid.UUID) error
}