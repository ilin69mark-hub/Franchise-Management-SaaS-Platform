package services

import (
	"context"
	"fmt"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
// На границе API принимаем float (фронт шлёт number), внутри конвертим
// в decimal через MoneyFromFloat (Round 2) — см. money.go.
type CreatePlanDTO struct {
	Name      string  `json:"name" binding:"required"`
	Price     float64 `json:"price" binding:"required,gte=0,lte=1000000000000"`
	MaxSalons int     `json:"max_salons" binding:"required,gte=0"`
	MaxUsers  int     `json:"max_users" binding:"required,gte=0"`
}
type UpdatePlanDTO struct {
	Name      *string  `json:"name,omitempty"`
	Price     *float64 `json:"price,omitempty" binding:"omitempty,gte=0,lte=1000000000000"`
	MaxSalons *int     `json:"max_salons,omitempty"`
	MaxUsers  *int     `json:"max_users,omitempty"`
}

// ---------- Реализация ----------
type planService struct {
	repo repository.PlanRepository
	db   *gorm.DB
}

func NewPlanService(r repository.PlanRepository) PlanService {
	var db *gorm.DB
	if provider, ok := r.(interface{ Database() *gorm.DB }); ok {
		db = provider.Database()
	}
	return &planService{repo: r, db: db}
}

func (s *planService) CreatePlan(ctx context.Context, dto CreatePlanDTO) (*models.Plan, error) {
	if dto.Price < 0 || dto.Price > 1e12 {
		return nil, fmt.Errorf("price out of range")
	}
	p := &models.Plan{
		Name:      dto.Name,
		Price:     MoneyFromFloat(dto.Price),
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
func applyPlanUpdate(existing *models.Plan, dto UpdatePlanDTO) error {
	if dto.Name != nil {
		existing.Name = *dto.Name
	}
	if dto.Price != nil {
		if *dto.Price < 0 || *dto.Price > 1e12 {
			return fmt.Errorf("price out of range")
		}
		existing.Price = MoneyFromFloat(*dto.Price)
	}
	if dto.MaxSalons != nil {
		existing.MaxSalons = *dto.MaxSalons
	}
	if dto.MaxUsers != nil {
		existing.MaxUsers = *dto.MaxUsers
	}
	return nil
}

func (s *planService) UpdatePlan(ctx context.Context, id string, dto UpdatePlanDTO) (*models.Plan, error) {
	if s.db != nil {
		if planID, parseErr := uuid.Parse(id); parseErr == nil {
			var existing models.Plan
			err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				if err := lockTenantsUsingPlan(tx, planID); err != nil {
					return err
				}
				if err := tx.First(&existing, "id = ?", planID).Error; err != nil {
					return err
				}
				if err := applyPlanUpdate(&existing, dto); err != nil {
					return err
				}
				return tx.Model(&models.Plan{}).Where("id = ?", planID).Updates(map[string]interface{}{
					"name":       existing.Name,
					"price":      existing.Price,
					"max_salons": existing.MaxSalons,
					"max_users":  existing.MaxUsers,
				}).Error
			})
			if err != nil {
				return nil, err
			}
			return &existing, nil
		}
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := applyPlanUpdate(existing, dto); err != nil {
		return nil, err
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
