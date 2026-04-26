package services

import (
	"context"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"
)

// ---------- Интерфейс ----------
type PlanService interface {
	CreatePlan(ctx context.Context, dto CreatePlanDTO) (*models.Plan, error)
	GetPlan(ctx context.Context, id string) (*models.Plan, error)
	ListPlans(ctx context.Context, opts repository.ListOptions) ([]*models.Plan, int64, error)
	UpdatePlan(ctx context.Context, id string, dto UpdatePlanDTO) (*models.Plan, error)
	DeletePlan(ctx context.Context, id string) error
}

// ---------- DTO ----------
type CreatePlanDTO struct {
	Name      string  `json:"name" binding:"required"`
	Price     float64 `json:"price" binding:"required,gte=0"`
	MaxSalons int     `json:"max_salons" binding:"required,gte=0"`
	MaxUsers  int     `json:"max_users" binding:"required,gte=0"`
}
type UpdatePlanDTO struct {
	Name      *string  `json:"name,omitempty"`
	Price     *float64 `json:"price,omitempty"`
	MaxSalons *int     `json:"max_salons,omitempty"`
	MaxUsers  *int     `json:"max_users,omitempty"`
}

// ---------- Реализация ----------
type planService struct{ repo repository.PlanRepository }

func NewPlanService(r repository.PlanRepository) PlanService { return &planService{repo: r} }

func (s *planService) CreatePlan(ctx context.Context, dto CreatePlanDTO) (*models.Plan, error) {
	p := &models.Plan{
		Name:      dto.Name,
		Price:     dto.Price,
		MaxSalons: dto.MaxSalons,
		MaxUsers:  dto.MaxUsers,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}
func (s *planService) GetPlan(ctx context.Context, id string) (*models.Plan, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *planService) ListPlans(ctx context.Context, opts repository.ListOptions) ([]*models.Plan, int64, error) {
	return s.repo.List(ctx, opts)
}
func (s *planService) UpdatePlan(ctx context.Context, id string, dto UpdatePlanDTO) (*models.Plan, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dto.Name != nil {
		existing.Name = *dto.Name
	}
	if dto.Price != nil {
		existing.Price = *dto.Price
	}
	if dto.MaxSalons != nil {
		existing.MaxSalons = *dto.MaxSalons
	}
	if dto.MaxUsers != nil {
		existing.MaxUsers = *dto.MaxUsers
	}
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}
func (s *planService) DeletePlan(ctx context.Context, id string) error {
	// проверяем, существует ли запись
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
