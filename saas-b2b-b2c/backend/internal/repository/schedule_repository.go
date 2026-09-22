package repository

import (
	"context"
	"time"

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
	q := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if dateStr != "" {
		if d, err := time.Parse("2006-01-02", dateStr); err == nil {
			ds := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
			de := ds.AddDate(0, 0, 1)
			q = q.Where("start_time >= ? AND start_time < ?", ds, de)
		} else {
			q = q.Where("DATE(start_time) = ?", dateStr)
		}
	}
	err := q.Order("start_time asc").Find(&events).Error
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
	q := r.db.WithContext(ctx).Where("user_id IN ?", userIDs)
	if dateStr != "" {
		if d, err := time.Parse("2006-01-02", dateStr); err == nil {
			ds := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
			de := ds.AddDate(0, 0, 1)
			q = q.Where("start_time >= ? AND start_time < ?", ds, de)
		} else {
			q = q.Where("DATE(start_time) = ?", dateStr)
		}
	}
	err := q.Order("start_time asc").Find(&events).Error
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
