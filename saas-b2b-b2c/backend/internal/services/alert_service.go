package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidMessage  = errors.New("message cannot be empty")
	ErrInvalidSeverity = errors.New("invalid severity value")
	ErrForbidden       = errors.New("forbidden")
)

type AlertService struct {
	alertRepo repository.AlertRepositoryInterface
	notifRepo *repository.NotificationRepository
	db        *gorm.DB
}

func NewAlertService(alertRepo repository.AlertRepositoryInterface, notifRepo *repository.NotificationRepository, db *gorm.DB) *AlertService {
	return &AlertService{alertRepo: alertRepo, notifRepo: notifRepo, db: db}
}

func NewAlertServiceWithDB(alertRepo repository.AlertRepositoryInterface, db *gorm.DB) *AlertService {
	return &AlertService{alertRepo: alertRepo, db: db}
}

func (s *AlertService) CreateAlert(ctx context.Context, userID uuid.UUID, title, message string, severity models.AlertSeverity, alertType models.AlertType, link string) (*models.Alert, error) {
	if message == "" {
		return nil, ErrInvalidMessage
	}
	if severity != models.AlertSeverityCritical && severity != models.AlertSeverityWarning && severity != models.AlertSeverityInfo {
		return nil, ErrInvalidSeverity
	}

	alert := &models.Alert{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        alertType,
		Title:       title,
		Description: message,
		Severity:    severity,
		Link:        link,
		CreatedAt:   time.Now(),
	}

	return s.alertRepo.CreateAlert(ctx, alert)
}

func (s *AlertService) GetAlertsByUser(ctx context.Context, userID uuid.UUID) ([]models.Alert, error) {
	return s.alertRepo.GetAlertsByUser(ctx, userID)
}

func (s *AlertService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	alerts, err := s.alertRepo.GetUnreadAlerts(ctx, userID)
	if err != nil {
		return 0, err
	}
	return len(alerts), nil
}

func (s *AlertService) MarkAsRead(ctx context.Context, userID, alertID uuid.UUID) error {
	alert, err := s.alertRepo.GetAlertByID(ctx, alertID)
	if err != nil {
		return err
	}
	if alert == nil {
		return errors.New("alert not found")
	}
	if alert.UserID != userID {
		return ErrForbidden
	}
	return s.alertRepo.MarkAsRead(ctx, alertID)
}

func (s *AlertService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	alerts, err := s.alertRepo.GetUnreadAlerts(ctx, userID)
	if err != nil {
		return err
	}
	if len(alerts) == 0 {
		return nil
	}
	return s.alertRepo.MarkAllAsRead(ctx, userID)
}

func (s *AlertService) DeleteAlert(ctx context.Context, userID, alertID uuid.UUID) error {
	alert, err := s.alertRepo.GetAlertByID(ctx, alertID)
	if err != nil {
		return err
	}
	if alert == nil {
		return errors.New("alert not found")
	}
	if alert.UserID != userID {
		return ErrForbidden
	}
	return s.alertRepo.DeleteAlert(ctx, alertID)
}

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

func (s *AlertService) MarkAsReadNotification(ctx context.Context, userID, alertID uuid.UUID) error {
	return s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", alertID, userID).
		Update("is_read", true).Error
}

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

	var avgConversion float64
	var todayCount int64

	monthAgo := now.AddDate(0, 0, -30)
	s.db.Model(&models.Lead{}).
		Where("salon_id = ? AND created_at > ?", salonID, monthAgo).
		Count(&todayCount)

	if todayCount > 10 {
		avgConversion = 20.0

		todayConversion := 5.0
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