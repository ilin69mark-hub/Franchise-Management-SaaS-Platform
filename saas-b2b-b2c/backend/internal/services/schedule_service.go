package services

import (
	"context"
	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"
	"time"

	"github.com/google/uuid"
)

type ScheduleService struct {
	repo repository.ScheduleRepositoryInterface
}

func NewScheduleService(repo repository.ScheduleRepositoryInterface) *ScheduleService {
	return &ScheduleService{repo: repo}
}

func (s *ScheduleService) GetUserSchedule(ctx context.Context, userID uuid.UUID, dateStr string) ([]models.ScheduleEvent, error) {
	return s.repo.GetUserEventsByDate(ctx, userID, dateStr)
}

func (s *ScheduleService) CreateEvent(ctx context.Context, event *models.ScheduleEvent) error {
	return s.repo.CreateEvent(ctx, event)
}

func (s *ScheduleService) UpdateEventStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.repo.UpdateEventStatus(ctx, id, status)
}

// --- Новые методы для Дилера (Управление командой) ---

// GetEventsByUsers возвращает события для списка менеджеров (для сводного календаря)
func (s *ScheduleService) GetEventsByUsers(ctx context.Context, userIDs []uuid.UUID, dateStr string) ([]models.ScheduleEvent, error) {
	return s.repo.GetEventsByUsers(ctx, userIDs, dateStr)
}

// UpdateEvent обновляет поля задачи (название, время, статус и т.д.)
func (s *ScheduleService) UpdateEvent(ctx context.Context, id uuid.UUID, data *models.UpdateScheduleEventRequest) error {
	updates := map[string]interface{}{}

	if data.Title != "" {
		updates["title"] = data.Title
	}
	if data.Description != "" {
		updates["description"] = data.Description
	}
	if data.Status != "" {
		updates["status"] = data.Status
	}
	if data.Priority != "" {
		updates["priority"] = data.Priority
	}

	// Парсинг времени, если передано
	if data.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, data.StartTime); err == nil {
			updates["start_time"] = t
		}
	}
	if data.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, data.EndTime); err == nil {
			updates["end_time"] = t
		}
	}

	return s.repo.UpdateEvent(ctx, id, updates)
}

// DeleteEvent удаляет задачу
func (s *ScheduleService) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteEvent(ctx, id)
}
