package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGoalRepo struct {
	mock.Mock
}

func (m *MockGoalRepo) Create(ctx context.Context, goal *models.Goal) error {
	args := m.Called(ctx, goal)
	return args.Error(0)
}

func (m *MockGoalRepo) GetByID(ctx context.Context, id string) (*models.Goal, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Goal), args.Error(1)
}

func (m *MockGoalRepo) GetByAssigneeAndDate(ctx context.Context, assigneeID string, date time.Time) (*models.Goal, error) {
	args := m.Called(ctx, assigneeID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Goal), args.Error(1)
}

func (m *MockGoalRepo) ListVisibleForUser(ctx context.Context, userID, role, tenantID string) ([]models.Goal, error) {
	args := m.Called(ctx, userID, role, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Goal), args.Error(1)
}

func (m *MockGoalRepo) Update(ctx context.Context, goal *models.Goal) error {
	args := m.Called(ctx, goal)
	return args.Error(0)
}

func (m *MockGoalRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestGoalService_CreateGoal_Success(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	dto := CreateGoalDTO{
		AssigneeID:   uuid.New().String(),
		Role:         "franchise_manager",
		SalesPlan:    1000.0,
		LeadsPlan:    10,
		CallsPlan:    50,
		MeetingsPlan: 5,
		Period:       "month",
		StartDate:    "2024-01-01",
		EndDate:      "2024-01-31",
	}

	ctx := context.WithValue(context.Background(), "role", "super_admin")

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(g *models.Goal) bool {
		return g.Role == "franchise_manager" && g.SalesPlan == 1000.0
	})).Return(nil)

	goal, err := service.CreateGoal(ctx, dto, uuid.New().String(), "tenant-1")

	assert.NoError(t, err)
	assert.NotNil(t, goal)
	mockRepo.AssertExpectations(t)
}

func TestGoalService_CreateGoal_InvalidRole(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	dto := CreateGoalDTO{
		AssigneeID: uuid.New().String(),
		Role:       "dealer",
	}

	ctx := context.WithValue(context.Background(), "role", "invalid_role")

	goal, err := service.CreateGoal(ctx, dto, uuid.New().String(), "tenant-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
	assert.Nil(t, goal)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestGoalService_CreateGoal_InvalidAssigneeID(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	dto := CreateGoalDTO{
		AssigneeID: "invalid-uuid",
		Role:       "dealer",
	}

	ctx := context.WithValue(context.Background(), "role", "super_admin")

	goal, err := service.CreateGoal(ctx, dto, uuid.New().String(), "tenant-1")

	assert.Error(t, err)
	assert.Nil(t, goal)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestGoalService_UpdateGoal_Success(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	goalID := uuid.New().String()
	existingGoal := &models.Goal{SalesPlan: 1000.0}
	dto := UpdateGoalDTO{SalesPlan: 2000.0}

	mockRepo.On("GetByID", mock.Anything, goalID).Return(existingGoal, nil)
	mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	goal, err := service.UpdateGoal(context.Background(), goalID, dto, uuid.New().String(), "tenant-1")

	assert.NoError(t, err)
	assert.NotNil(t, goal)
	mockRepo.AssertExpectations(t)
}

func TestGoalService_UpdateGoal_NotFound(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	goalID := "invalid-goal"
	dto := UpdateGoalDTO{SalesPlan: 2000.0}

	mockRepo.On("GetByID", mock.Anything, goalID).Return(nil, errors.New("not found"))

	goal, err := service.UpdateGoal(context.Background(), goalID, dto, uuid.New().String(), "tenant-1")

	assert.Error(t, err)
	assert.Nil(t, goal)
	mockRepo.AssertExpectations(t)
}

func TestGoalService_GetMyGoal_Success(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	assigneeID := uuid.New().String()
	date := time.Now()
	expectedGoal := &models.Goal{AssigneeID: uuid.MustParse(assigneeID)}

	mockRepo.On("GetByAssigneeAndDate", mock.Anything, assigneeID, date).Return(expectedGoal, nil)

	goal, err := service.GetMyGoal(context.Background(), assigneeID, date)

	assert.NoError(t, err)
	assert.Equal(t, expectedGoal, goal)
	mockRepo.AssertExpectations(t)
}

func TestGoalService_GetMyGoal_NotFound(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	assigneeID := uuid.New().String()
	date := time.Now()

	mockRepo.On("GetByAssigneeAndDate", mock.Anything, assigneeID, date).Return(nil, errors.New("not found"))

	goal, err := service.GetMyGoal(context.Background(), assigneeID, date)

	assert.Error(t, err)
	assert.Nil(t, goal)
	mockRepo.AssertExpectations(t)
}

func TestGoalService_GetVisibleGoals_Success(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	expectedGoals := []models.Goal{
		{Role: "dealer"},
		{Role: "dealer_manager"},
	}

	mockRepo.On("ListVisibleForUser", mock.Anything, "user-1", "dealer", "tenant-1").Return(expectedGoals, nil)

	goals, err := service.GetVisibleGoals(context.Background(), "user-1", "dealer", "tenant-1")

	assert.NoError(t, err)
	assert.Len(t, goals, 2)
	mockRepo.AssertExpectations(t)
}

func TestGoalService_GetVisibleGoals_Error(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	mockRepo.On("ListVisibleForUser", mock.Anything, "user-1", "dealer", "tenant-1").Return(nil, errors.New("db error"))

	goals, err := service.GetVisibleGoals(context.Background(), "user-1", "dealer", "tenant-1")

	assert.Error(t, err)
	assert.Nil(t, goals)
	mockRepo.AssertExpectations(t)
}

func TestGoalService_DeleteGoal_Success(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	goalID := uuid.New().String()

	mockRepo.On("Delete", mock.Anything, goalID).Return(nil)

	err := service.DeleteGoal(context.Background(), goalID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGoalService_DeleteGoal_Error(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	goalID := uuid.New().String()

	mockRepo.On("Delete", mock.Anything, goalID).Return(errors.New("db error"))

	err := service.DeleteGoal(context.Background(), goalID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}