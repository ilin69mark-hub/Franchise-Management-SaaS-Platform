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
	"github.com/stretchr/testify/require"
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

func (m *MockGoalRepo) GetUserTenant(ctx context.Context, userID string) (*uuid.UUID, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	tid := args.Get(0).(uuid.UUID)
	return &tid, args.Error(1)
}

func TestGoalService_CreateGoal_Success(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	dto := CreateGoalDTO{
		AssigneeID:   uuid.New().String(),
		Role:         string(models.RoleFranchisorManager),
		SalesPlan:    1000.0,
		LeadsPlan:    10,
		CallsPlan:    50,
		MeetingsPlan: 5,
		Period:       "month",
		StartDate:    "2024-01-01",
		EndDate:      "2024-01-31",
	}

	ctx := context.WithValue(context.Background(), "role", string(models.RoleSuperAdmin))

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(g *models.Goal) bool {
		return g.Role == string(models.RoleFranchisorManager) && g.SalesPlan == 1000.0
	})).Return(nil)

	goal, err := service.CreateGoal(ctx, dto, uuid.New().String(), "tenant-1")

	assert.NoError(t, err)
	assert.NotNil(t, goal)
	mockRepo.AssertExpectations(t)
}

func TestGoalService_CreateGoal_FranchiserManagerAssignment(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	dto := CreateGoalDTO{
		AssigneeID: uuid.New().String(),
		Role:       string(models.RoleDealerManager),
		SalesPlan:  500.0,
		Period:     "month",
		StartDate:  "2024-03-01",
		EndDate:    "2024-03-31",
	}

	ctx := context.WithValue(context.Background(), "role", string(models.RoleFranchisorManager))

	tenant := uuid.New()
	mockRepo.On("GetUserTenant", mock.Anything, dto.AssigneeID).Return(tenant, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(g *models.Goal) bool {
		return g.Role == string(models.RoleDealerManager)
	})).Return(nil)

	goal, err := service.CreateGoal(ctx, dto, uuid.New().String(), tenant.String())

	assert.NoError(t, err)
	assert.NotNil(t, goal)
	mockRepo.AssertExpectations(t)
}

func TestGoalService_CreateGoal_CtxRoleAsModelRole(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	dto := CreateGoalDTO{
		AssigneeID: uuid.New().String(),
		Role:       string(models.RoleDealer),
		SalesPlan:  750.0,
		Period:     "month",
		StartDate:  "2024-04-01",
		EndDate:    "2024-04-30",
	}

	ctx := context.WithValue(context.Background(), "role", models.RoleSuperAdmin)

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(g *models.Goal) bool {
		return g.Role == string(models.RoleDealer)
	})).Return(nil)

	goal, err := service.CreateGoal(ctx, dto, uuid.New().String(), "tenant-1")

	assert.NoError(t, err)
	assert.NotNil(t, goal)
	mockRepo.AssertExpectations(t)
}

func TestCanAssignRoles(t *testing.T) {
	tests := []struct {
		name     string
		assigner string
		assignee string
		expected bool
	}{
		{"super_admin can assign franchiser", string(models.RoleSuperAdmin), string(models.RoleFranchisor), true},
		{"super_admin can assign franchiser_manager", string(models.RoleSuperAdmin), string(models.RoleFranchisorManager), true},
		{"super_admin can assign dealer", string(models.RoleSuperAdmin), string(models.RoleDealer), true},
		{"super_admin can assign salon_manager", string(models.RoleSuperAdmin), string(models.RoleDealerManager), true},
		{"franchiser can assign franchiser_manager", string(models.RoleFranchisor), string(models.RoleFranchisorManager), true},
		{"franchiser can assign dealer", string(models.RoleFranchisor), string(models.RoleDealer), true},
		{"franchiser can assign salon_manager", string(models.RoleFranchisor), string(models.RoleDealerManager), true},
		{"franchiser_manager can assign dealer", string(models.RoleFranchisorManager), string(models.RoleDealer), true},
		{"franchiser_manager can assign salon_manager", string(models.RoleFranchisorManager), string(models.RoleDealerManager), true},
		{"dealer can assign salon_manager", string(models.RoleDealer), string(models.RoleDealerManager), true},
		{"salon_manager cannot assign anyone", string(models.RoleDealerManager), string(models.RoleDealer), false},
		{"franchiser cannot assign super_admin", string(models.RoleFranchisor), string(models.RoleSuperAdmin), false},
		{"dealer cannot assign franchiser_manager", string(models.RoleDealer), string(models.RoleFranchisorManager), false},
		{"unknown assigner denied", "unknown_role", string(models.RoleDealer), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, canAssign(tt.assigner, tt.assignee))
		})
	}
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

	goal, err := service.UpdateGoal(context.Background(), goalID, dto, uuid.New().String(), "tenant-1", "super_admin")

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

	goal, err := service.UpdateGoal(context.Background(), goalID, dto, uuid.New().String(), "tenant-1", "super_admin")

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

	mockRepo.On("GetByID", mock.Anything, goalID).Return(&models.Goal{}, nil)
	mockRepo.On("Delete", mock.Anything, goalID).Return(nil)

	err := service.DeleteGoal(context.Background(), goalID, uuid.New().String(), "", "super_admin")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGoalService_DeleteGoal_ForbiddenCrossTenant(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	goalID := uuid.New().String()
	assigner := uuid.New()
	otherTenant := uuid.New()
	goalTenant := uuid.New()
	assert.NotEqual(t, otherTenant.String(), goalTenant.String())
	existing := &models.Goal{AssignerID: assigner, AssigneeID: uuid.New(), Role: "salon_manager", TenantID: &goalTenant}
	mockRepo.On("GetByID", mock.Anything, goalID).Return(existing, nil)

	err := service.DeleteGoal(context.Background(), goalID, uuid.New().String(), otherTenant.String(), "dealer")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
	mockRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestGoalService_UpdateGoal_ForbiddenCrossTenant(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	goalID := uuid.New().String()
	goalTenant := uuid.New()
	otherTenant := uuid.New()
	existing := &models.Goal{AssignerID: uuid.New(), AssigneeID: uuid.New(), Role: "salon_manager", TenantID: &goalTenant}
	mockRepo.On("GetByID", mock.Anything, goalID).Return(existing, nil)

	dto := UpdateGoalDTO{SalesPlan: 2000.0}
	_, err := service.UpdateGoal(context.Background(), goalID, dto, uuid.New().String(), otherTenant.String(), "dealer")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
	mockRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestGoalService_DeleteGoal_Error(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	goalID := uuid.New().String()

	mockRepo.On("GetByID", mock.Anything, goalID).Return(&models.Goal{}, nil)
	mockRepo.On("Delete", mock.Anything, goalID).Return(errors.New("db error"))

	err := service.DeleteGoal(context.Background(), goalID, uuid.New().String(), "", "super_admin")

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGoalService_CreateGoal_CrossTenantForbidden(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	dto := CreateGoalDTO{
		AssigneeID: uuid.New().String(),
		Role:       string(models.RoleDealer),
		SalesPlan:  500.0,
		Period:     "month",
		StartDate:  "2024-03-01",
		EndDate:    "2024-03-31",
	}

	ctx := context.WithValue(context.Background(), "role", string(models.RoleFranchisorManager))

	mockRepo.On("GetUserTenant", mock.Anything, dto.AssigneeID).Return(uuid.New(), nil)

	goal, err := service.CreateGoal(ctx, dto, uuid.New().String(), uuid.New().String())

	require.Error(t, err)
	require.Contains(t, err.Error(), "forbidden")
	require.Nil(t, goal)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestGoalService_CreateGoal_UnknownAssigneeRejected(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	dto := CreateGoalDTO{
		AssigneeID: uuid.New().String(),
		Role:       string(models.RoleDealer),
		SalesPlan:  500.0,
		Period:     "month",
		StartDate:  "2024-03-01",
		EndDate:    "2024-03-31",
	}

	ctx := context.WithValue(context.Background(), "role", string(models.RoleDealer))

	mockRepo.On("GetUserTenant", mock.Anything, dto.AssigneeID).Return(nil, errors.New("not found"))

	goal, err := service.CreateGoal(ctx, dto, uuid.New().String(), uuid.New().String())

	require.Error(t, err)
	require.Nil(t, goal)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestGoalService_CreateGoal_DuplicateMaps409(t *testing.T) {
	mockRepo := new(MockGoalRepo)
	service := NewGoalService(mockRepo)

	tenant := uuid.New()
	assignee := uuid.New().String()
	dto := CreateGoalDTO{
		AssigneeID: assignee,
		Role:       string(models.RoleDealer),
		SalesPlan:  500.0,
		Period:     "month",
		StartDate:  "2024-03-01",
		EndDate:    "2024-03-31",
	}

	ctx := context.WithValue(context.Background(), "role", string(models.RoleFranchisorManager))

	mockRepo.On("GetUserTenant", mock.Anything, assignee).Return(tenant, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New(`duplicate key value violates unique constraint "idx_goals_assignee_period_dates"`))

	goal, err := service.CreateGoal(ctx, dto, uuid.New().String(), tenant.String())

	require.ErrorIs(t, err, ErrGoalExists)
	require.Nil(t, goal)
}
