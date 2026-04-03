package repository

import (
    "context"
    "franchise-saas-backend/internal/models"

    "gorm.io/gorm"
)

type TenantRepository struct {
    db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
    return &TenantRepository{db: db}
}

// FindAll - получить все сети
func (r *TenantRepository) FindAll(ctx context.Context) ([]models.Tenant, error) {
    var tenants []models.Tenant
    err := r.db.WithContext(ctx).Find(&tenants).Error
    return tenants, err
}

// FindByID - найти сеть по ID
func (r *TenantRepository) FindByID(ctx context.Context, id string) (*models.Tenant, error) {
    var tenant models.Tenant
    err := r.db.WithContext(ctx).First(&tenant, "id = ?", id).Error
    if err != nil {
        return nil, err
    }
    return &tenant, nil
}

// CreateTenant - создать сеть (для AuthService)
func (r *TenantRepository) CreateTenant(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error) {
    if err := r.db.WithContext(ctx).Create(tenant).Error; err != nil {
        return nil, err
    }
    return tenant, nil
}
