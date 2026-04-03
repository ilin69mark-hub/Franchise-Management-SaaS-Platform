package services

import (
	"context"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChecklistService struct {
	repo *repository.ChecklistRepository
}

func NewChecklistService(db *gorm.DB) *ChecklistService {
	return &ChecklistService{
		repo: repository.NewChecklistRepository(db),
	}
}

// GetUserChecklists - основная точка входа для получения задач
func (s *ChecklistService) GetUserChecklists(ctx context.Context, user *models.User, status, priority string, isArchive bool) ([]models.Checklist, error) {
	// Суперадмин видит всё
	if user.Role == models.RoleSuperAdmin {
		return s.repo.FindAllGlobal(ctx, status, priority, isArchive)
	}

	// Остальные - только свои (где они создатель или исполнитель)
	return s.repo.FindUserTasks(ctx, user.ID, status, priority, isArchive)
}

func (s *ChecklistService) GetAllGlobal(ctx context.Context, status, priority string) ([]models.Checklist, error) {
	return s.repo.FindAllGlobal(ctx, status, priority, false)
}

func (s *ChecklistService) GetAllChecklists(ctx context.Context, status, priority string) ([]models.Checklist, error) {
	return s.repo.FindAll(ctx, status, priority)
}

func (s *ChecklistService) GetChecklistByID(ctx context.Context, id uuid.UUID) (*models.Checklist, error) {
	return s.repo.GetChecklistByID(ctx, id)
}

func (s *ChecklistService) CreateChecklist(ctx context.Context, chk *models.Checklist) error {
	chk.CreatedAt = time.Now()
	chk.UpdatedAt = time.Now()
	return s.repo.CreateChecklist(ctx, chk)
}

func (s *ChecklistService) UpdateChecklist(ctx context.Context, chk *models.Checklist) error {
	chk.UpdatedAt = time.Now()
	return s.repo.UpdateChecklist(ctx, chk)
}

// UpdateStatus - новый метод для смены статуса (в работе / выполнено)
func (s *ChecklistService) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	chk, err := s.repo.GetChecklistByID(ctx, id)
	if err != nil {
		return err
	}

	// Бизнес-логика: нельзя менять статус у уже завершенного?
	// Или можно вернуть в работу? Пока разрешим менять.

	chk.Status = status
	chk.UpdatedAt = time.Now()
	return s.repo.UpdateChecklist(ctx, chk)
}

func (s *ChecklistService) CompleteChecklist(ctx context.Context, id uuid.UUID) error {
	// Используем новый метод
	return s.UpdateStatus(ctx, id, "completed")
}

func (s *ChecklistService) DeleteChecklist(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteChecklist(ctx, id)
}
