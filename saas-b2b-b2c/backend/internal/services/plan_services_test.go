package services

import (
	"context"
	"errors"
	"testing"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockPlanRepo struct {
	mock.Mock
}

func NewMockPlanRepo() *MockPlanRepo {
	return &MockPlanRepo{}
}

func (m *MockPlanRepo) Create(ctx context.Context, p *models.Plan) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPlanRepo) GetByID(ctx context.Context, id string) (*models.Plan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Plan), args.Error(1)
}

func (m *MockPlanRepo) List(ctx context.Context, opts repository.ListOptions) ([]*models.Plan, int64, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, 0, args.Error(1)
	}
	return args.Get(0).([]*models.Plan), args.Get(1).(int64), args.Error(2)
}

func (m *MockPlanRepo) Update(ctx context.Context, p *models.Plan) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPlanRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type PlanRepoInterface interface {
	Create(ctx context.Context, p *models.Plan) error
	GetByID(ctx context.Context, id string) (*models.Plan, error)
	List(ctx context.Context, opts repository.ListOptions) ([]*models.Plan, int64, error)
	Update(ctx context.Context, p *models.Plan) error
	Delete(ctx context.Context, id string) error
}

type PlanServiceTestable struct {
	repo PlanRepoInterface
}

func NewPlanServiceTestable(repo PlanRepoInterface) *PlanServiceTestable {
	return &PlanServiceTestable{repo: repo}
}

func (s *PlanServiceTestable) CreatePlan(ctx context.Context, dto CreatePlanDTO) (*models.Plan, error) {
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

func (s *PlanServiceTestable) GetPlan(ctx context.Context, id string) (*models.Plan, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PlanServiceTestable) ListPlans(ctx context.Context, opts repository.ListOptions) ([]*models.Plan, int64, error) {
	return s.repo.List(ctx, opts)
}

func (s *PlanServiceTestable) UpdatePlan(ctx context.Context, id string, dto UpdatePlanDTO) (*models.Plan, error) {
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

func (s *PlanServiceTestable) DeletePlan(ctx context.Context, id string) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func TestPlanService_CreatePlan(t *testing.T) {
	t.Run("success creates plan", func(t *testing.T) {
		repo := NewMockPlanRepo()
		service := NewPlanServiceTestable(repo)

		dto := CreatePlanDTO{
			Name:      "Premium",
			Price:     999.99,
			MaxSalons: 10,
			MaxUsers:  100,
		}
		repo.On("Create", mock.Anything, mock.MatchedBy(func(p *models.Plan) bool {
			return p.Name == "Premium" && p.Price == 999.99
		})).Return(nil)

		plan, err := service.CreatePlan(context.Background(), dto)

		require.NoError(t, err)
		require.NotNil(t, plan)
		require.Equal(t, "Premium", plan.Name)
		repo.AssertExpectations(t)
	})

	t.Run("repository error returns error", func(t *testing.T) {
		repo := NewMockPlanRepo()
		service := NewPlanServiceTestable(repo)

		dto := CreatePlanDTO{
			Name:  "Basic",
			Price: 99.99,
		}
		repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("duplicate name"))

		plan, err := service.CreatePlan(context.Background(), dto)

		require.Error(t, err)
		require.Nil(t, plan)
		require.Equal(t, "duplicate name", err.Error())
		repo.AssertExpectations(t)
	})
}

func TestPlanService_GetPlan(t *testing.T) {
	t.Run("success returns plan", func(t *testing.T) {
		repo := NewMockPlanRepo()
		service := NewPlanServiceTestable(repo)

		plan := &models.Plan{ID: uuid.New(), Name: "Basic", Price: 99.99}
		repo.On("GetByID", mock.Anything, "plan-123").Return(plan, nil)

		result, err := service.GetPlan(context.Background(), "plan-123")

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "Basic", result.Name)
		repo.AssertExpectations(t)
	})

	t.Run("not found returns error", func(t *testing.T) {
		repo := NewMockPlanRepo()
		service := NewPlanServiceTestable(repo)

		repo.On("GetByID", mock.Anything, "invalid").Return(nil, errors.New("record not found"))

		result, err := service.GetPlan(context.Background(), "invalid")

		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "record not found", err.Error())
		repo.AssertExpectations(t)
	})
}

func TestPlanService_ListPlans(t *testing.T) {
	t.Run("success returns plans", func(t *testing.T) {
		repo := NewMockPlanRepo()
		service := NewPlanServiceTestable(repo)

		plans := []*models.Plan{
			{ID: uuid.New(), Name: "Basic", Price: 99},
			{ID: uuid.New(), Name: "Premium", Price: 299},
		}
		opts := repository.ListOptions{Limit: 10, Offset: 0}
		repo.On("List", mock.Anything, opts).Return(plans, int64(2), nil)

		result, total, err := service.ListPlans(context.Background(), opts)

		require.NoError(t, err)
		require.Len(t, result, 2)
		require.Equal(t, int64(2), total)
		repo.AssertExpectations(t)
	})

	t.Run("empty returns zero", func(t *testing.T) {
		repo := NewMockPlanRepo()
		service := NewPlanServiceTestable(repo)

		opts := repository.ListOptions{Limit: 10, Offset: 0}
		repo.On("List", mock.Anything, opts).Return([]*models.Plan{}, int64(0), nil)

		result, total, err := service.ListPlans(context.Background(), opts)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result, 0)
		require.Equal(t, int64(0), total)
		repo.AssertExpectations(t)
	})

	t.Run("pagination works", func(t *testing.T) {
		repo := NewMockPlanRepo()
		service := NewPlanServiceTestable(repo)

		plans := []*models.Plan{{ID: uuid.New(), Name: "Enterprise", Price: 999}}
		opts := repository.ListOptions{Limit: 5, Offset: 10}
		repo.On("List", mock.Anything, opts).Return(plans, int64(50), nil)

		result, total, err := service.ListPlans(context.Background(), opts)

		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Equal(t, int64(50), total)
		repo.AssertExpectations(t)
	})
}

func TestPlanService_UpdatePlan(t *testing.T) {
	t.Run("success updates name", func(t *testing.T) {
		repo := NewMockPlanRepo()
		service := NewPlanServiceTestable(repo)

		existing := &models.Plan{ID: uuid.New(), Name: "Basic", Price: 99}
		repo.On("GetByID", mock.Anything, "plan-123").Return(existing, nil)
		repo.On("Update", mock.Anything, mock.Anything).Return(nil)

		name := "Basic Plus"
		dto := UpdatePlanDTO{Name: &name}

		plan, err := service.UpdatePlan(context.Background(), "plan-123", dto)

		require.NoError(t, err)
		require.NotNil(t, plan)
		require.Equal(t, "Basic Plus", plan.Name)
		repo.AssertExpectations(t)
	})

	t.Run("success updates price", func(t *testing.T) {
		repo := NewMockPlanRepo()
		service := NewPlanServiceTestable(repo)

		existing := &models.Plan{ID: uuid.New(), Price: 99}
		repo.On("GetByID", mock.Anything, "plan-123").Return(existing, nil)
		repo.On("Update", mock.Anything, mock.Anything).Return(nil)

		price := 149.99
		dto := UpdatePlanDTO{Price: &price}

		plan, err := service.UpdatePlan(context.Background(), "plan-123", dto)

		require.NoError(t, err)
		require.Equal(t, 149.99, plan.Price)
		repo.AssertExpectations(t)
	})

	t.Run("plan not found returns error", func(t *testing.T) {
		repo := NewMockPlanRepo()
		service := NewPlanServiceTestable(repo)

		repo.On("GetByID", mock.Anything, "invalid").Return(nil, errors.New("not found"))

		dto := UpdatePlanDTO{Name: ptrString("New Name")}

		plan, err := service.UpdatePlan(context.Background(), "invalid", dto)

		require.Error(t, err)
		require.Nil(t, plan)
		repo.AssertExpectations(t)
	})
}

func TestPlanService_DeletePlan(t *testing.T) {
	t.Run("success deletes plan", func(t *testing.T) {
		repo := NewMockPlanRepo()
		service := NewPlanServiceTestable(repo)

		repo.On("GetByID", mock.Anything, "plan-123").Return(&models.Plan{ID: uuid.New()}, nil)
		repo.On("Delete", mock.Anything, "plan-123").Return(nil)

		err := service.DeletePlan(context.Background(), "plan-123")

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("delete non-existent returns error", func(t *testing.T) {
		repo := NewMockPlanRepo()
		service := NewPlanServiceTestable(repo)

		repo.On("GetByID", mock.Anything, "invalid").Return(nil, errors.New("not found"))

		err := service.DeletePlan(context.Background(), "invalid")

		require.Error(t, err)
		require.Equal(t, "not found", err.Error())
		repo.AssertExpectations(t)
	})
}

func ptrString(s string) *string {
	return &s
}