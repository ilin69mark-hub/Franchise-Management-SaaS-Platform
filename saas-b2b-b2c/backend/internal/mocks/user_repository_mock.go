package mocks

import (
	"context"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{}
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserFields(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error {
	args := m.Called(ctx, id, fields)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) FindUsersByTenantID(ctx context.Context, tenantID uuid.UUID) ([]models.User, error) {
	args := m.Called(ctx, tenantID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	if args.Get(0) == nil {
		return []models.User{}, nil
	}
	return args.Get(0).([]models.User), nil
}

func (m *MockUserRepository) FindAllGlobal(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	if args.Get(0) == nil {
		return []models.User{}, nil
	}
	return args.Get(0).([]models.User), nil
}

func (m *MockUserRepository) FindUserByIDAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) CountByRole(ctx context.Context, role string) (int64, error) {
	args := m.Called(ctx, role)
	return args.Get(0).(int64), args.Error(1)
}