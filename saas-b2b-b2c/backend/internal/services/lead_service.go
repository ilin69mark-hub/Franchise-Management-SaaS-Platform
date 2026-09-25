package services

import (
	"context"
	"errors"
	"fmt" // Добавлен импорт для форматирования ошибок
	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LeadService struct {
	repo repository.LeadRepositoryInterface
}

func NewLeadService(repo repository.LeadRepositoryInterface) *LeadService {
	return &LeadService{repo: repo}
}

// CreateLead создает нового лида.
// Мы изменили входной параметр: вместо просто managerID передаем весь объект User.
// Это нужно, чтобы получить SalonID (салон, к которому привязан менеджер).
func (s *LeadService) CreateLead(ctx context.Context, user *models.User, req models.CreateLeadRequest) (*models.Lead, error) {

	// 1. Проверка безопасности: у менеджера должен быть назначен салон.
	// Если user.SalonID равен nil, база данных выдаст ошибку при попытке вставки,
	// поэтому мы перехватываем это заранее и выдаем понятное сообщение.
	if user.SalonID == nil {
		return nil, fmt.Errorf("пользователь не привязан к салону, невозможно создать лид")
	}

	// 2. Заполняем модель лида
	lead := &models.Lead{
		SalonID:         *user.SalonID, // Автоматически берем ID салона из профиля менеджера
		ManagerID:       user.ID,       // ID менеджера, который создает лид
		FullName:        req.FullName,
		Phone:           req.Phone,
		Email:           req.Email,
		InterestProduct: req.InterestProduct,
		Budget:          decimal.NewFromFloat(req.Budget),
		Status:          "new", // Статус по умолчанию для новых клиентов
	}

	// 3. Сохраняем в базе
	if err := s.repo.CreateLead(ctx, lead); err != nil {
		return nil, err
	}
	return lead, nil
}

func (s *LeadService) GetMyLeads(ctx context.Context, managerID uuid.UUID) ([]models.Lead, error) {
	return s.repo.GetLeadsByManager(ctx, managerID)
}

// REAUDIT-4: типизированные ошибки статуса, чтобы хендлер отдавал 400, а не 500.
var (
	ErrInvalidLeadStatus     = errors.New("invalid lead status")
	ErrInvalidLeadTransition = errors.New("invalid lead status transition")
)

// validLeadStatuses — REAUDIT-4: домен статусов лида. Раньше принималась любая
// строка, а KPI считали "выручку" по набору разных статусов в разных отчётах.
var validLeadStatuses = map[string]bool{
	"new": true, "contact": true, "meeting": true, "wait": true,
	"sale": true, "paid": true, "contract": true,
	"archive": true, "cancelled": true, "other": true,
}

// leadStatusTransitions — куда можно переходить (обратный ход запрещён).
var leadStatusTransitions = map[string]map[string]bool{
	"new":       {"contact": true, "meeting": true, "archive": true, "cancelled": true, "other": true},
	"contact":   {"meeting": true, "wait": true, "sale": true, "archive": true, "cancelled": true, "other": true},
	"meeting":   {"wait": true, "sale": true, "archive": true, "cancelled": true, "other": true},
	"wait":      {"contact": true, "meeting": true, "sale": true, "archive": true, "cancelled": true, "other": true},
	"sale":      {"paid": true, "contract": true, "archive": true, "other": true},
	"paid":      {"archive": true},
	"contract":  {"paid": true, "archive": true},
	"cancelled": {"new": true},
	"archive":   {},
	"other":     {"archive": true},
}

func (s *LeadService) UpdateStatus(ctx context.Context, managerID, leadID uuid.UUID, status string) error {
	// Проверяем, что лид принадлежит этому менеджеру (безопасность)
	lead, err := s.repo.GetLeadByID(ctx, leadID, managerID)
	if err != nil {
		return err // Ошибка доступа или лид не найден
	}
	normalized := strings.ToLower(strings.TrimSpace(status))
	if !validLeadStatuses[normalized] {
		return fmt.Errorf("%w: %s", ErrInvalidLeadStatus, status)
	}
	current := strings.ToLower(strings.TrimSpace(lead.Status))
	if allowed, known := leadStatusTransitions[current]; known && current != normalized {
		if !allowed[normalized] {
			return fmt.Errorf("%w: %s -> %s", ErrInvalidLeadTransition, current, normalized)
		}
	}
	return s.repo.UpdateLeadStatus(ctx, leadID, normalized)
}

func (s *LeadService) AddActivity(ctx context.Context, userID, leadID uuid.UUID, req models.AddLeadActivityRequest) error {
	// RE-AUDIT: заметки только в свои лиды (раньше — запись в лид чужой сети).
	if _, err := s.repo.GetLeadByID(ctx, leadID, userID); err != nil {
		return errors.New("lead not found or access denied")
	}
	activity := &models.LeadActivity{
		LeadID:      leadID,
		UserID:      userID,
		Type:        req.Type,
		Description: req.Description,
	}
	return s.repo.AddActivity(ctx, activity)
}

func (s *LeadService) GetLeadDetails(ctx context.Context, managerID, leadID uuid.UUID) (*models.Lead, []models.LeadActivity, error) {
	lead, err := s.repo.GetLeadByID(ctx, leadID, managerID)
	if err != nil {
		return nil, nil, err
	}
	activities, err := s.repo.GetLeadActivities(ctx, leadID)
	if err != nil {
		return lead, nil, err
	}
	return lead, activities, nil
}
