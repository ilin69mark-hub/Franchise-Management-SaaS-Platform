package mocks

import (
	"context"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockGoalRepository struct {
	mock.Mock
}

func NewMockGoalRepository() *MockGoalRepository {
	return &MockGoalRepository{}
}

func (m *MockGoalRepository) CreateGoal(ctx context.Context, goal *models.Goal) (*models.Goal, error) {
	args := m.Called(ctx, goal)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Goal), args.Error(1)
}

func (m *MockGoalRepository) GetGoalByID(ctx context.Context, id uuid.UUID) (*models.Goal, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Goal), args.Error(1)
}

func (m *MockGoalRepository) GetGoalsByUser(ctx context.Context, userID uuid.UUID) ([]models.Goal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Goal), args.Error(1)
}

func (m *MockGoalRepository) GetGoalsByTenant(ctx context.Context, tenantID uuid.UUID) ([]models.Goal, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]models.Goal), args.Error(1)
}

func (m *MockGoalRepository) UpdateGoal(ctx context.Context, goal *models.Goal) error {
	args := m.Called(ctx, goal)
	return args.Error(0)
}

func (m *MockGoalRepository) DeleteGoal(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}