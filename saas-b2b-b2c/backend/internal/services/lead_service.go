package services

import (
	"context"
	"fmt" // Добавлен импорт для форматирования ошибок
	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
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
		Budget:          req.Budget,
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

func (s *LeadService) UpdateStatus(ctx context.Context, managerID, leadID uuid.UUID, status string) error {
	// Проверяем, что лид принадлежит этому менеджеру (безопасность)
	_, err := s.repo.GetLeadByID(ctx, leadID, managerID)
	if err != nil {
		return err // Ошибка доступа или лид не найден
	}
	return s.repo.UpdateLeadStatus(ctx, leadID, status)
}

func (s *LeadService) AddActivity(ctx context.Context, userID, leadID uuid.UUID, req models.AddLeadActivityRequest) error {
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
