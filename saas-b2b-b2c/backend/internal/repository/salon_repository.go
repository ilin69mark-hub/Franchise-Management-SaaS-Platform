package repository

import (
	"context"
	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalonRepository struct {
	db *gorm.DB
}

func NewSalonRepository(db *gorm.DB) *SalonRepository {
	return &SalonRepository{db: db}
}

// CreateSalon создает новый салон
func (r *SalonRepository) CreateSalon(ctx context.Context, salon *models.Salon) error {
	return r.db.WithContext(ctx).Create(salon).Error
}

// GetSalonsByTenant получает все салоны сети
func (r *SalonRepository) GetSalonsByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.Salon, error) {
	var salons []models.Salon
	// Убрали Preload, чтобы избежать ошибки 500
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&salons).Error
	return salons, err
}

// UpdateUserSalon назначает пользователя на салон
func (r *SalonRepository) UpdateUserSalon(ctx context.Context, userID, salonID uuid.UUID) error {
	// Используем UpdateColumn, чтобы обойти возможные проблемы с хуками или пустыми значениями
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("salon_id", salonID).Error
}

// UpdateSalon обновляет данные салона
func (r *SalonRepository) UpdateSalon(ctx context.Context, salon *models.Salon) error {
	return r.db.WithContext(ctx).Save(salon).Error
}

// DeleteSalon удаляет салон
func (r *SalonRepository) DeleteSalon(ctx context.Context, salonID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Salon{}, "id = ?", salonID).Error
}
