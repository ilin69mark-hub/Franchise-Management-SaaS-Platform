package repository

import (
	"context"
	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ScheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) CreateEvent(ctx context.Context, event *models.ScheduleEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *ScheduleRepository) GetUserEventsByDate(ctx context.Context, userID uuid.UUID, dateStr string) ([]models.ScheduleEvent, error) {
	var events []models.ScheduleEvent
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("DATE(start_time) = ?", dateStr).
		Order("start_time asc").
		Find(&events).Error
	return events, err
}

func (r *ScheduleRepository) UpdateEventStatus(ctx context.Context, eventID uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.ScheduleEvent{}).
		Where("id = ?", eventID).
		Update("status", status).Error
}

// GetEventsByUsers - получение событий списка пользователей
func (r *ScheduleRepository) GetEventsByUsers(ctx context.Context, userIDs []uuid.UUID, dateStr string) ([]models.ScheduleEvent, error) {
	var events []models.ScheduleEvent
	err := r.db.WithContext(ctx).
		Where("user_id IN ?", userIDs).
		Where("DATE(start_time) = ?", dateStr).
		Order("start_time asc").
		Find(&events).Error
	return events, err
}

// UpdateEvent - универсальное обновление через map (исправляет ошибку типа)
func (r *ScheduleRepository) UpdateEvent(ctx context.Context, eventID uuid.UUID, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&models.ScheduleEvent{}).
		Where("id = ?", eventID).
		Updates(updates).Error
}

// DeleteEvent - удаление
func (r *ScheduleRepository) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&models.ScheduleEvent{}, eventID).Error
}
