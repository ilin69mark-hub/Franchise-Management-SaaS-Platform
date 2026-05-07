package repository

import (
	"context"
	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) FindAll(ctx context.Context) ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := r.db.WithContext(ctx).Find(&tenants).Error
	return tenants, err
}

func (r *TenantRepository) CreateTenant(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error) {
	if err := r.db.WithContext(ctx).Create(tenant).Error; err != nil {
		return nil, err
	}
	return tenant, nil
}

func (r *TenantRepository) GetTenantByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.db.WithContext(ctx).First(&tenant, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) GetTenantBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.db.WithContext(ctx).First(&tenant, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	return r.GetTenantByID(ctx, id)
}

func (r *TenantRepository) UpdateTenant(ctx context.Context, tenant *models.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

func (r *TenantRepository) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Tenant{}, "id = ?", id).Error
}

func (r *TenantRepository) FindByStatus(ctx context.Context, status string) ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&tenants).Error
	return tenants, err
}

func (r *TenantRepository) CountUsersByTenantID(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}