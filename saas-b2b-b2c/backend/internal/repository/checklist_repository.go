package repository

import (
    "context"
    "time"

    "franchise-saas-backend/internal/models"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type ChecklistRepository struct {
    db *gorm.DB
}

func NewChecklistRepository(db *gorm.DB) *ChecklistRepository {
    return &ChecklistRepository{db: db}
}

// FindAllGlobal для Супер-Админа
func (r *ChecklistRepository) FindAllGlobal(ctx context.Context, status, priority string, isArchive bool) ([]models.Checklist, error) {
    var checklists []models.Checklist
    query := r.db.WithContext(ctx)

    query = applyFilters(query, status, priority, isArchive)

    err := query.Order("created_at desc").Find(&checklists).Error
    return checklists, err
}

// FindUserTasks находит задачи конкретного пользователя (исполнитель или создатель)
func (r *ChecklistRepository) FindUserTasks(ctx context.Context, userID uuid.UUID, status, priority string, isArchive bool) ([]models.Checklist, error) {
    var checklists []models.Checklist
    query := r.db.WithContext(ctx)

    // Фильтр принадлежности: Я создатель ИЛИ Я исполнитель
    query = query.Where("user_id = ? OR assigned_to = ?", userID, userID)

    query = applyFilters(query, status, priority, isArchive)

    err := query.Order("created_at desc").Find(&checklists).Error
    return checklists, err
}

// applyFilters - вспомогательная функция для фильтрации
func applyFilters(query *gorm.DB, status, priority string, isArchive bool) *gorm.DB {
    now := time.Now()

    if isArchive {
        // АРХИВ: (Статус completed) ИЛИ (Статус pending И время вышло)
        // Формула: (status = 'completed') OR (status = 'pending' AND end_date < NOW)
        query = query.Where(
            query.Where("status = ?", "completed").
                Or("status = ? AND end_date < ?", "pending", now),
        )
    } else {
        // АКТИВНЫЕ: Статус pending И (время НЕ вышло ИЛИ время не указано)
        // Формула: status = 'pending' AND (end_date IS NULL OR end_date >= NOW)
        query = query.Where("status = ? AND (end_date IS NULL OR end_date >= ?)", "pending", now)
    }

    // Дополнительные фильтры из запроса
    if status != "" {
        query = query.Where("status = ?", status)
    }
    if priority != "" {
        query = query.Where("priority = ?", priority)
    }

    return query
}

// FindAll (старый метод, оставим для совместимости)
func (r *ChecklistRepository) FindAll(ctx context.Context, status, priority string) ([]models.Checklist, error) {
    var checklists []models.Checklist
    query := r.db.WithContext(ctx)

    if status != "" {
        query = query.Where("status = ?", status)
    }
    if priority != "" {
        query = query.Where("priority = ?", priority)
    }

    err := query.Order("created_at desc").Find(&checklists).Error
    return checklists, err
}

func (r *ChecklistRepository) GetChecklistByID(ctx context.Context, id uuid.UUID) (*models.Checklist, error) {
    var checklist models.Checklist
    err := r.db.WithContext(ctx).First(&checklist, id).Error
    if err != nil {
        return nil, err
    }
    return &checklist, nil
}

func (r *ChecklistRepository) CreateChecklist(ctx context.Context, checklist *models.Checklist) error {
    return r.db.WithContext(ctx).Create(checklist).Error
}

func (r *ChecklistRepository) UpdateChecklist(ctx context.Context, checklist *models.Checklist) error {
    return r.db.WithContext(ctx).Save(checklist).Error
}

func (r *ChecklistRepository) DeleteChecklist(ctx context.Context, id uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&models.Checklist{}, id).Error
}

// GetTaskStats возвращает общее кол-во задач и кол-во выполненных
func (r *ChecklistRepository) GetTaskStats() (total int64, completed int64, err error) {
    err = r.db.Model(&models.Checklist{}).Count(&total).Error
    if err != nil {
        return 0, 0, err
    }
    err = r.db.Model(&models.Checklist{}).Where("status = ?", "completed").Count(&completed).Error
    return total, completed, err
}
