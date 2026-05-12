package services

import (
	"context"
	"errors"
	"testing"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPlanRepo struct {
	mock.Mock
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

func TestPlanService_CreatePlan_Success(t *testing.T) {
	mockRepo := new(MockPlanRepo)
	service := NewPlanService(mockRepo)

	dto := CreatePlanDTO{
		Name:      "Premium",
		Price:     100.0,
		MaxSalons: 5,
		MaxUsers:  10,
	}

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *models.Plan) bool {
		return p.Name == "Premium" && p.Price == 100.0
	})).Return(nil)

	plan, err := service.CreatePlan(context.Background(), dto)

	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, "Premium", plan.Name)
	mockRepo.AssertExpectations(t)
}

func TestPlanService_CreatePlan_Error(t *testing.T) {
	mockRepo := new(MockPlanRepo)
	service := NewPlanService(mockRepo)

	dto := CreatePlanDTO{Name: "Premium", Price: 100.0}

	mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	plan, err := service.CreatePlan(context.Background(), dto)

	assert.Error(t, err)
	assert.Nil(t, plan)
	mockRepo.AssertExpectations(t)
}

func TestPlanService_GetPlan_Success(t *testing.T) {
	mockRepo := new(MockPlanRepo)
	service := NewPlanService(mockRepo)

	planID := "test-plan-id"
	expected := &models.Plan{Name: "Basic"}

	mockRepo.On("GetByID", mock.Anything, planID).Return(expected, nil)

	plan, err := service.GetPlan(context.Background(), planID)

	assert.NoError(t, err)
	assert.Equal(t, expected, plan)
	mockRepo.AssertExpectations(t)
}

func TestPlanService_GetPlan_NotFound(t *testing.T) {
	mockRepo := new(MockPlanRepo)
	service := NewPlanService(mockRepo)

	mockRepo.On("GetByID", mock.Anything, "invalid").Return(nil, errors.New("not found"))

	plan, err := service.GetPlan(context.Background(), "invalid")

	assert.Error(t, err)
	assert.Nil(t, plan)
	mockRepo.AssertExpectations(t)
}

func TestPlanService_ListPlans_Success(t *testing.T) {
	mockRepo := new(MockPlanRepo)
	service := NewPlanService(mockRepo)

	expected := []*models.Plan{
		{Name: "Basic"},
		{Name: "Premium"},
	}

	mockRepo.On("List", mock.Anything, repository.ListOptions{}).Return(expected, int64(2), nil)

	plans, total, err := service.ListPlans(context.Background(), repository.ListOptions{})

	assert.NoError(t, err)
	assert.Len(t, plans, 2)
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestPlanService_UpdatePlan_Success(t *testing.T) {
	mockRepo := new(MockPlanRepo)
	service := NewPlanService(mockRepo)

	planID := "test-plan-id"
	dto := UpdatePlanDTO{Name: ptrToString("Updated")}

	mockRepo.On("GetByID", mock.Anything, planID).Return(&models.Plan{}, nil)
	mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	plan, err := service.UpdatePlan(context.Background(), planID, dto)

	assert.NoError(t, err)
	assert.NotNil(t, plan)
	mockRepo.AssertExpectations(t)
}

func TestPlanService_UpdatePlan_NotFound(t *testing.T) {
	mockRepo := new(MockPlanRepo)
	service := NewPlanService(mockRepo)

	dto := UpdatePlanDTO{Name: ptrToString("Updated")}

	mockRepo.On("GetByID", mock.Anything, "invalid").Return(nil, errors.New("not found"))

	plan, err := service.UpdatePlan(context.Background(), "invalid", dto)

	assert.Error(t, err)
	assert.Nil(t, plan)
	mockRepo.AssertExpectations(t)
}

func TestPlanService_DeletePlan_Success(t *testing.T) {
	mockRepo := new(MockPlanRepo)
	service := NewPlanService(mockRepo)

	planID := "test-plan-id"

	mockRepo.On("GetByID", mock.Anything, planID).Return(&models.Plan{}, nil)
	mockRepo.On("Delete", mock.Anything, planID).Return(nil)

	err := service.DeletePlan(context.Background(), planID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestPlanService_DeletePlan_Error(t *testing.T) {
	mockRepo := new(MockPlanRepo)
	service := NewPlanService(mockRepo)

	mockRepo.On("GetByID", mock.Anything, "invalid").Return(nil, errors.New("not found"))

	err := service.DeletePlan(context.Background(), "invalid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func ptrToString(s string) *string {
	return &s
}