package repository

import (
	"context"
	"time"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------
//
//	GoalRepository – интерфейс
//
// ---------------------------------------------------------------------
type GoalRepository interface {
	// Создать (или обновить) цель
	Create(ctx context.Context, g *models.Goal) error

	// Получить цель, назначенную конкретному пользователю на конкретную дату
	GetByAssigneeAndDate(ctx context.Context, assigneeID string, date time.Time) (*models.Goal, error)

	// Список целей, которые видит текущий пользователь
	// role – роль текущего пользователя (super_admin, franchise_manager, dealer, dealer_manager, salon_manager)
	// tenantID – id tenant'а (может быть пустым, тогда ищем по всему DB)
	ListVisibleForUser(ctx context.Context, userID, role, tenantID string) ([]models.Goal, error)

	// Удалить цель по ID
	Delete(ctx context.Context, id string) error
}

// ---------------------------------------------------------------------
//
//	Реализация
//
// ---------------------------------------------------------------------
type goalRepo struct{ db *gorm.DB }

func NewGoalRepository(db *gorm.DB) GoalRepository { return &goalRepo{db: db} }

func (r *goalRepo) Create(ctx context.Context, g *models.Goal) error {
	return r.db.WithContext(ctx).Create(g).Error
}

// ---------------------------------------------------------------
//
//	GetByAssigneeAndDate
//
// ---------------------------------------------------------------
func (r *goalRepo) GetByAssigneeAndDate(ctx context.Context, assigneeID string, date time.Time) (*models.Goal, error) {
	var g models.Goal
	assigneeUUID, err := uuid.Parse(assigneeID)
	if err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).
		Where("assignee_id = ? AND target_date = ?", assigneeUUID, date).
		First(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

// ---------------------------------------------------------------
//
//	ListVisibleForUser – что может увидеть текущий пользователь
//
// ---------------------------------------------------------------
func (r *goalRepo) ListVisibleForUser(ctx context.Context, userID, role, tenantID string) ([]models.Goal, error) {
	var goals []models.Goal

	// -------------------------------------------------
	// 1️⃣ Super‑admin (и franchise_manager) – может видеть «все» цели,
	//    но если указан tenantID – только цели внутри этого tenant.
	// -------------------------------------------------
	if role == "super_admin" || role == "franchise_manager" {
		query := r.db.WithContext(ctx).Model(&models.Goal{})
		if tenantID != "" {
			tid, err := uuid.Parse(tenantID)
			if err != nil {
				// Некорректный tenant_id в JWT – возвращаем ошибку,
				// а не «panic», чтобы обработчик мог отдать 500 с пояснением.
				return nil, err
			}
			query = query.Where("tenant_id = ?", tid)
		}
		if err := query.Find(&goals).Error; err != nil {
			return nil, err
		}
		return goals, nil
	}

	// -------------------------------------------------
	// 2️⃣ Обычные роли (franchiser, franchiser_manager, dealer, dealer_manager, salon_manager)
	//    – показываем только цели, где пользователь является получателем
	//      ИЛИ назначителем.
	// -------------------------------------------------
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	query := r.db.WithContext(ctx).
		Where(
			r.db.Where("assignee_id = ?", uid).
				Or("assigner_id = ?", uid),
		)

	// Если в JWT присутствует tenant_id – добавляем фильтр,
	// иначе ищем по всей базе.
	if tenantID != "" {
		tid, err := uuid.Parse(tenantID)
		if err != nil {
			return nil, err
		}
		query = query.Where("tenant_id = ?", tid)
	}

	if err := query.Find(&goals).Error; err != nil {
		return nil, err
	}
	return goals, nil
}

// ---------------------------------------------------------------
//
//	Delete
//
// ---------------------------------------------------------------
func (r *goalRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Goal{}).Error
}
