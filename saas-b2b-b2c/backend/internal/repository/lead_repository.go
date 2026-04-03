package repository

import (
	"context"
	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeadRepository struct {
	db *gorm.DB
}

func NewLeadRepository(db *gorm.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

// CreateLead создает нового лида
func (r *LeadRepository) CreateLead(ctx context.Context, lead *models.Lead) error {
	return r.db.WithContext(ctx).Create(lead).Error
}

// GetLeadsByManager получает всех лидов конкретного менеджера
func (r *LeadRepository) GetLeadsByManager(ctx context.Context, managerID uuid.UUID) ([]models.Lead, error) {
	var leads []models.Lead
	err := r.db.WithContext(ctx).
		Where("manager_id = ?", managerID).
		Order("created_at desc").
		Find(&leads).Error
	return leads, err
}

// GetLeadByID получает лида по ID (проверяя владельца)
func (r *LeadRepository) GetLeadByID(ctx context.Context, leadID, managerID uuid.UUID) (*models.Lead, error) {
	var lead models.Lead
	err := r.db.WithContext(ctx).
		Where("id = ? AND manager_id = ?", leadID, managerID).
		First(&lead).Error
	if err != nil {
		return nil, err
	}
	return &lead, nil
}

// UpdateLeadStatus обновляет статус лида
func (r *LeadRepository) UpdateLeadStatus(ctx context.Context, leadID uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.Lead{}).
		Where("id = ?", leadID).
		Update("status", status).Error
}

// AddActivity добавляет запись в историю взаимодействий
func (r *LeadRepository) AddActivity(ctx context.Context, activity *models.LeadActivity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

// GetLeadActivities получает историю по лиду
func (r *LeadRepository) GetLeadActivities(ctx context.Context, leadID uuid.UUID) ([]models.LeadActivity, error) {
	var activities []models.LeadActivity
	err := r.db.WithContext(ctx).
		Where("lead_id = ?", leadID).
		Order("created_at desc").
		Find(&activities).Error
	return activities, err
}
