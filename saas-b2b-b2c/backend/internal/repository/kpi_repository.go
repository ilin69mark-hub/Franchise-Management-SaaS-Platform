package repository

import (
	"context"
	"franchise-saas-backend/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KPIRepository struct {
	db *gorm.DB
}

func NewKPIRepository(db *gorm.DB) *KPIRepository {
	return &KPIRepository{db: db}
}

func (r *KPIRepository) UpsertGoal(ctx context.Context, goal *models.DailyGoal) error {
	var existing models.DailyGoal
	query := r.db.WithContext(ctx).Where("target_date = ?", goal.TargetDate)

	if goal.UserID != nil {
		query = query.Where("user_id = ?", goal.UserID)
	} else if goal.SalonID != nil {
		query = query.Where("salon_id = ?", goal.SalonID)
	} else {
		return gorm.ErrRecordNotFound // Нужен хоть один ID
	}

	if err := query.First(&existing).Error; err == gorm.ErrRecordNotFound {
		return r.db.Create(goal).Error
	}

	return r.db.Model(&existing).Updates(map[string]interface{}{
		"sales_plan":    goal.SalesPlan,
		"leads_plan":    goal.LeadsPlan,
		"calls_plan":    goal.CallsPlan,
		"meetings_plan": goal.MeetingsPlan,
		"updated_at":    time.Now(),
	}).Error
}

func (r *KPIRepository) GetGoal(ctx context.Context, userID *uuid.UUID, salonID *uuid.UUID, date time.Time) (*models.DailyGoal, error) {
	var goal models.DailyGoal
	query := r.db.WithContext(ctx).Where("target_date = ?", date)

	if userID != nil {
		query = query.Where("user_id = ?", userID)
	} else if salonID != nil {
		query = query.Where("salon_id = ?", salonID)
	}

	err := query.First(&goal).Error
	if err != nil {
		return nil, err
	}
	return &goal, nil
}
