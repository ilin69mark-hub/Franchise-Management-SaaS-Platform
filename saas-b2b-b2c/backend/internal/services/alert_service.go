package services

import (
	"context"
	"encoding/json"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AlertService struct {
	notifRepo *repository.NotificationRepository
	db        *gorm.DB
}

func NewAlertService(notifRepo *repository.NotificationRepository, db *gorm.DB) *AlertService {
	return &AlertService{notifRepo: notifRepo, db: db}
}

// GetUnreadAlerts - получить непрочитанные алерты
func (s *AlertService) GetUnreadAlerts(ctx context.Context, userID uuid.UUID) ([]models.Notification, int, error) {
	var alerts []models.Notification
	var count int64

	s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Order("created_at DESC").
		Find(&alerts)
	s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count)

	return alerts, int(count), nil
}

// MarkAsRead - пометить алерт прочитанным
func (s *AlertService) MarkAsRead(ctx context.Context, userID, alertID uuid.UUID) error {
	return s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", alertID, userID).
		Update("is_read", true).Error
}

// GenerateAlertsForUser - сгенерировать алерты для пользователя
func (s *AlertService) GenerateAlertsForUser(ctx context.Context, userID uuid.UUID) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	if user.SalonID == nil {
		return nil
	}
	salonID := *user.SalonID

	now := time.Now()

	// === 1. Просроченные замеры (> 3 дней) ===
	threeDaysAgo := now.AddDate(0, 0, -3)
	var overdueMeasurements []models.Lead
	s.db.Where("salon_id = ? AND status = ? AND updated_at < ?", salonID, "meeting", threeDaysAgo).
		Find(&overdueMeasurements)

	for _, lead := range overdueMeasurements {
		alert := models.Notification{
			ID:        uuid.New(),
			UserID:    &userID,
			TenantID:  salonID,
			Type:      models.NotificationTypeSystem,
			Title:     "Просроченный замер",
			Message:   "Замер для " + lead.FullName + " просрочен более 3 дней",
			IsRead:    false,
			Data:      mustJson(map[string]string{"lead_id": lead.ID.String(), "type": "overdue_measurement"}),
			CreatedAt: now,
		}
		s.db.Create(&alert)
	}

	// === 2. Брошенные КП (> 5 дней) ===
	fiveDaysAgo := now.AddDate(0, 0, -5)
	var abandonedKP []models.Lead
	s.db.Where("salon_id = ? AND status = ? AND updated_at < ?", salonID, "wait", fiveDaysAgo).
		Find(&abandonedKP)

	for _, lead := range abandonedKP {
		alert := models.Notification{
			ID:        uuid.New(),
			UserID:    &userID,
			TenantID:  salonID,
			Type:      models.NotificationTypeSystem,
			Title:     "Брошенное КП",
			Message:   "КП для " + lead.FullName + " без ответа более 5 дней",
			IsRead:    false,
			Data:      mustJson(map[string]string{"lead_id": lead.ID.String(), "type": "abandoned_kp"}),
			CreatedAt: now,
		}
		s.db.Create(&alert)
	}

	// === 3. Падение конверсии (> 20% к среднему) ===
	var avgConversion float64
	var todayCount int64

	// Средняя конверсия за последние 30 дней
	monthAgo := now.AddDate(0, 0, -30)
	s.db.Model(&models.Lead{}).
		Where("salon_id = ? AND created_at > ?", salonID, monthAgo).
		Count(&todayCount)

	if todayCount > 10 {
		avgConversion = 20.0 // Заглушка - нужно реальный расчет

		todayConversion := 5.0 // Заглушка - получить из реальных данных
		drop := ((avgConversion - todayConversion) / avgConversion) * 100

		if drop > 20 {
			alert := models.Notification{
				ID:        uuid.New(),
				UserID:    &userID,
				TenantID:  salonID,
				Type:      models.NotificationTypeSystem,
				Title:     "Падение конверсии",
				Message:   "Конверсия упала на " + formatFloat(drop) + "% относительно среднего",
				IsRead:    false,
				Data:      mustJson(map[string]string{"type": "conversion_drop"}),
				CreatedAt: now,
			}
			s.db.Create(&alert)
		}
	}

	// === 4. Падение трафика (> 30% к среднему) ===
	var avgLeads, todayLeads int64
	s.db.Model(&models.Lead{}).
		Where("salon_id = ? AND created_at > ?", salonID, monthAgo).
		Count(&avgLeads)

	yesterday := now.AddDate(0, 0, -1)
	s.db.Model(&models.Lead{}).
		Where("salon_id = ? AND DATE(created_at) = ?", salonID, yesterday.Format("2006-01-02")).
		Count(&todayLeads)

	if avgLeads > 0 {
		avgDailyLeads := float64(avgLeads) / 30
		drop := ((avgDailyLeads - float64(todayLeads)) / avgDailyLeads) * 100

		if drop > 30 {
			alert := models.Notification{
				ID:        uuid.New(),
				UserID:    &userID,
				TenantID:  salonID,
				Type:      models.NotificationTypeSystem,
				Title:     "Падение трафика",
				Message:   "Трафик упал на " + formatFloat(drop) + "% относительно среднего",
				IsRead:    false,
				Data:      mustJson(map[string]string{"type": "traffic_drop"}),
				CreatedAt: now,
			}
			s.db.Create(&alert)
		}
	}

	return nil
}

func mustJson(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func formatFloat(v float64) string {
	return string(rune(int(v)+'0')) + "%"
}