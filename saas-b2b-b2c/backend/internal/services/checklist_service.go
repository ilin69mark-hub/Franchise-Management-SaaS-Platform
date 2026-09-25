package services

import (
	"context"
	"errors"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
)

type ChecklistService struct {
	repo     repository.ChecklistRepositoryInterface
	userRepo repository.UserRepositoryInterface
}

func NewChecklistService(repo repository.ChecklistRepositoryInterface) *ChecklistService {
	return &ChecklistService{
		repo: repo,
	}
}

// WithUserRepository добавляет доступ к пользователям: нужен для проверки
// исполнителя чек-листа (REAUDIT-3: assigned_to принимался без проверки, задачу
// можно было назначить сотруднику чужого tenant'а).
func (s *ChecklistService) WithUserRepository(userRepo repository.UserRepositoryInterface) *ChecklistService {
	s.userRepo = userRepo
	return s
}

// ValidateAssignee — исполнитель обязан существовать и быть в том же tenant.
func (s *ChecklistService) ValidateAssignee(ctx context.Context, assignee *uuid.UUID, tenantID *uuid.UUID) error {
	if assignee == nil || *assignee == uuid.Nil {
		return nil
	}
	if s.userRepo == nil {
		return errors.New("assignee validation unavailable")
	}
	target, err := s.userRepo.GetUserByID(ctx, *assignee)
	if err != nil || target == nil {
		return errors.New("assignee not found")
	}
	if target.Role == models.RoleSuperAdmin {
		return nil
	}
	if tenantID == nil || target.TenantID == nil || *target.TenantID != *tenantID {
		return errors.New("assignee must be in same tenant")
	}
	return nil
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

// GetAllChecklists — список в рамках tenant вызывающего (REAUDIT-3).
// Без tenant ответ пустой (fail-closed): глобальный обзор есть только у
// super_admin и идёт через GetAllGlobal.
func (s *ChecklistService) GetAllChecklists(ctx context.Context, tenantID *uuid.UUID, status, priority string) ([]models.Checklist, error) {
	if tenantID == nil || *tenantID == uuid.Nil {
		return []models.Checklist{}, nil
	}
	return s.repo.FindTenant(ctx, *tenantID, status, priority)
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
