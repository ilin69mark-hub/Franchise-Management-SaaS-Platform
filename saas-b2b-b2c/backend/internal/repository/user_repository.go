package repository

import (
    "context"
    "franchise-saas-backend/internal/models"

    "github.com/google/uuid"
    "time"
	"gorm.io/gorm"
)

type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) GetUsersByTenantAndRole(ctx context.Context, tenantID uuid.UUID, role models.Role) ([]models.User, error) {
    var users []models.User
    err := r.db.WithContext(ctx).
        Where("tenant_id = ? AND role = ?", tenantID, role).
        Find(&users).Error
    return users, err
}

func (r *UserRepository) UpdateUserFields(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
    return r.db.WithContext(ctx).
        Model(&models.User{}).
        Where("id = ?", userID).
        Updates(updates).Error
}

// FindAllGlobal - НОВЫЙ МЕТОД для Супер-Админа (без фильтрации по тенанту)
func (r *UserRepository) FindAllGlobal(ctx context.Context) ([]models.User, error) {
    var users []models.User
    if err := r.db.WithContext(ctx).Order("created_at desc").Find(&users).Error; err != nil {
        return nil, err
    }
    return users, nil
}

// FindUsersByTenantID - получить всех сотрудников сети (существующий)
func (r *UserRepository) FindUsersByTenantID(ctx context.Context, tenantID uuid.UUID) ([]models.User, error) {
    var users []models.User
    if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&users).Error; err != nil {
        return nil, err
    }
    return users, nil
}

// FindUserByIDAndTenant - найти сотрудника по ID внутри конкретной сети (проверка прав)
func (r *UserRepository) FindUserByIDAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*models.User, error) {
    var user models.User
    if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}

// DeleteUser - удалить пользователя
func (r *UserRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
    return r.db.WithContext(ctx).Delete(&models.User{}, userID).Error
}

// === ЭТАП 2: АНАЛИТИКА ===

// CreateLog сохраняет запись о действии пользователя
func (r *UserRepository) CreateLog(log *models.UserLog) error {
    return r.db.Create(log).Error
}

// GetDAU считает уникальных юзеров за период
func (r *UserRepository) GetDAU(start, end time.Time) (int64, error) {
    var count int64
    err := r.db.Model(&models.UserLog{}).
        Select("COUNT(DISTINCT user_id)").
        Where("created_at BETWEEN ? AND ?", start, end).
        Scan(&count).Error
    return count, err
}

// GetMAU считает активных за последние 30 дней
func (r *UserRepository) GetMAU() (int64, error) {
    var count int64
    thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
    err := r.db.Model(&models.UserLog{}).
        Select("COUNT(DISTINCT user_id)").
        Where("created_at >= ?", thirtyDaysAgo).
        Scan(&count).Error
    return count, err
}

// GetRegistrationsStats группирует регистрации по дням
func (r *UserRepository) GetRegistrationsStats(days int) ([]map[string]interface{}, error) {
    var results []map[string]interface{}
    startDate := time.Now().AddDate(0, 0, -days)

    rows, err := r.db.Table("users").
        Select("DATE(created_at) as date, COUNT(*) as count").
        Where("created_at >= ?", startDate).
        Group("DATE(created_at)").
        Order("date ASC").
        Rows()

    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var date string
        var count int64
        if err := rows.Scan(&date, &count); err != nil {
            return nil, err
        }
        results = append(results, map[string]interface{}{"date": date, "count": count})
    }
    return results, nil
}
