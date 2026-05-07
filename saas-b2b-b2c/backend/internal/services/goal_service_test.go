package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockGoalRepo struct {
	mock.Mock
}

func NewMockGoalRepo() *MockGoalRepo {
	return &MockGoalRepo{}
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

type GoalRepoInterface interface {
	Create(ctx context.Context, goal *models.Goal) error
	GetByID(ctx context.Context, id string) (*models.Goal, error)
	GetByAssigneeAndDate(ctx context.Context, assigneeID string, date time.Time) (*models.Goal, error)
	ListVisibleForUser(ctx context.Context, userID, role, tenantID string) ([]models.Goal, error)
	Update(ctx context.Context, goal *models.Goal) error
	Delete(ctx context.Context, id string) error
}

type GoalServiceTestable struct {
	repo GoalRepoInterface
}

func NewGoalServiceTestable(repo GoalRepoInterface) *GoalServiceTestable {
	return &GoalServiceTestable{repo: repo}
}

func (s *GoalServiceTestable) CreateGoal(ctx context.Context, dto CreateGoalDTO, assignerID, tenantID string) (*models.Goal, error) {
	assignerRole, _ := ctx.Value("role").(string)
	if !canAssign(assignerRole, dto.Role) {
		return nil, errors.New("you are not allowed to assign a goal to this role")
	}

	assigneeUUID, err := uuid.Parse(dto.AssigneeID)
	if err != nil {
		return nil, err
	}

	period := models.PeriodDay
	if dto.Period == "week" || dto.Period == "month" || dto.Period == "year" {
		period = models.GoalPeriod(dto.Period)
	}

	var startDate, endDate, targetDate time.Time

	if dto.StartDate != "" {
		startDate, _ = time.Parse("2006-01-02", dto.StartDate)
	}
	if dto.EndDate != "" {
		endDate, _ = time.Parse("2006-01-02", dto.EndDate)
	}
	if startDate.IsZero() && !endDate.IsZero() {
		startDate = endDate
	}
	if !startDate.IsZero() && endDate.IsZero() {
		endDate = startDate
	}
	targetDate = endDate

	goal := &models.Goal{
		AssignerID:    uuid.MustParse(assignerID),
		AssigneeID:   assigneeUUID,
		Role:        dto.Role,
		SalesPlan:    dto.SalesPlan,
		LeadsPlan:   dto.LeadsPlan,
		CallsPlan:   dto.CallsPlan,
		MeetingsPlan: dto.MeetingsPlan,
		Period:     period,
		StartDate:  startDate,
		EndDate:   endDate,
		TargetDate: targetDate,
	}
	if tenantID != "" {
		tid, _ := uuid.Parse(tenantID)
		goal.TenantID = &tid
	}
	if err := s.repo.Create(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func (s *GoalServiceTestable) UpdateGoal(ctx context.Context, id string, dto UpdateGoalDTO, assignerID, tenantID string) (*models.Goal, error) {
	goal, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("goal not found")
	}

	if dto.SalesPlan > 0 {
		goal.SalesPlan = dto.SalesPlan
	}
	if dto.LeadsPlan > 0 {
		goal.LeadsPlan = dto.LeadsPlan
	}
	if dto.CallsPlan > 0 {
		goal.CallsPlan = dto.CallsPlan
	}
	if dto.MeetingsPlan > 0 {
		goal.MeetingsPlan = dto.MeetingsPlan
	}
	if dto.Period != "" {
		goal.Period = models.GoalPeriod(dto.Period)
	}
	if dto.StartDate != "" {
		goal.StartDate, _ = time.Parse("2006-01-02", dto.StartDate)
	}
	if dto.EndDate != "" {
		goal.EndDate, _ = time.Parse("2006-01-02", dto.EndDate)
	}

	if err := s.repo.Update(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func (s *GoalServiceTestable) GetMyGoal(ctx context.Context, assigneeID string, date time.Time) (*models.Goal, error) {
	return s.repo.GetByAssigneeAndDate(ctx, assigneeID, date)
}

func (s *GoalServiceTestable) GetVisibleGoals(ctx context.Context, userID, role, tenantID string) ([]models.Goal, error) {
	return s.repo.ListVisibleForUser(ctx, userID, role, tenantID)
}

func (s *GoalServiceTestable) DeleteGoal(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// Тесты на canAssign (вспомогательная функция)
func TestCanAssign(t *testing.T) {
	tests := []struct {
		name         string
		assignerRole string
		assigneeRole string
		expected     bool
	}{
		{"super_admin to franchise_manager", "super_admin", "franchise_manager", true},
		{"franchiser to franchise_manager", "franchiser", "franchise_manager", true},
		{"franchiser to dealer", "franchiser", "dealer", true},
		{"franchiser to salon_manager", "franchiser", "salon_manager", true},
		{"dealer to salon_manager", "dealer", "salon_manager", true},
		{"dealer_manager to salon_manager", "dealer_manager", "salon_manager", true},
		{"franchise_manager to dealer", "franchise_manager", "dealer", true},
		{"franchise_manager to salon_manager", "franchise_manager", "salon_manager", false},
		{"dealer to franchise_manager", "dealer", "franchise_manager", false},
		{"salon_manager to anyone", "salon_manager", "dealer", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canAssign(tt.assignerRole, tt.assigneeRole)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestGoalService_CreateGoal(t *testing.T) {
	t.Run("success creates goal with day period", func(t *testing.T) {
		repo := NewMockGoalRepo()
		service := NewGoalServiceTestable(repo)

		ctx := context.WithValue(context.Background(), "role", "franchiser")
		dto := CreateGoalDTO{
			AssigneeID: uuid.New().String(),
			Role:       "salon_manager",
			SalesPlan:  100000,
			LeadsPlan:   20,
			Period:      "day",
			StartDate:  "2026-05-01",
			EndDate:    "2026-05-01",
		}
		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		goal, err := service.CreateGoal(ctx, dto, uuid.New().String(), "")

		require.NoError(t, err)
		require.NotNil(t, goal)
		require.Equal(t, models.PeriodDay, goal.Period)
		repo.AssertExpectations(t)
	})

	t.Run("success creates goal with month period", func(t *testing.T) {
		repo := NewMockGoalRepo()
		service := NewGoalServiceTestable(repo)

		ctx := context.WithValue(context.Background(), "role", "super_admin")
		dto := CreateGoalDTO{
			AssigneeID: uuid.New().String(),
			Role:        "franchise_manager",
			SalesPlan:   500000,
			Period:      "month",
			StartDate:  "2026-05-01",
			EndDate:    "2026-05-31",
		}
		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		goal, err := service.CreateGoal(ctx, dto, uuid.New().String(), "")

		require.NoError(t, err)
		require.NotNil(t, goal)
		require.Equal(t, models.GoalPeriod("month"), goal.Period)
		repo.AssertExpectations(t)
	})

	t.Run("forbidden role assignment returns error", func(t *testing.T) {
		repo := NewMockGoalRepo()
		service := NewGoalServiceTestable(repo)

		// salon_manager не может назначать цели франчайзеру
		ctx := context.WithValue(context.Background(), "role", "salon_manager")
		dto := CreateGoalDTO{
			AssigneeID: uuid.New().String(),
			Role:       "franchiser",
		}

		goal, err := service.CreateGoal(ctx, dto, uuid.New().String(), "")

		require.Error(t, err)
		require.Contains(t, err.Error(), "not allowed")
		require.Nil(t, goal)
	})

	t.Run("invalid assignee uuid returns error", func(t *testing.T) {
		repo := NewMockGoalRepo()
		service := NewGoalServiceTestable(repo)

		ctx := context.WithValue(context.Background(), "role", "franchiser")
		dto := CreateGoalDTO{
			AssigneeID: "invalid-uuid",
			Role:       "salon_manager",
		}

		goal, err := service.CreateGoal(ctx, dto, uuid.New().String(), "")

		require.Error(t, err)
		require.Nil(t, goal)
	})

	t.Run("repository error returns error", func(t *testing.T) {
		repo := NewMockGoalRepo()
		service := NewGoalServiceTestable(repo)

		ctx := context.WithValue(context.Background(), "role", "dealer")
		dto := CreateGoalDTO{
			AssigneeID: uuid.New().String(),
			Role:       "salon_manager",
		}
		repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("duplicate key"))

		goal, err := service.CreateGoal(ctx, dto, uuid.New().String(), "")

		require.Error(t, err)
		require.Nil(t, goal)
		require.Equal(t, "duplicate key", err.Error())
		repo.AssertExpectations(t)
	})
}

func TestGoalService_UpdateGoal(t *testing.T) {
	goalID := uuid.New().String()

	t.Run("success updates sales plan", func(t *testing.T) {
		repo := NewMockGoalRepo()
		service := NewGoalServiceTestable(repo)

		existingGoal := &models.Goal{
			ID:        uuid.MustParse(goalID),
			SalesPlan: 100000,
		}
		repo.On("GetByID", mock.Anything, goalID).Return(existingGoal, nil)
		repo.On("Update", mock.Anything, mock.Anything).Return(nil)

		dto := UpdateGoalDTO{
			SalesPlan: 150000,
		}

		goal, err := service.UpdateGoal(context.Background(), goalID, dto, uuid.New().String(), "")

		require.NoError(t, err)
		require.NotNil(t, goal)
		require.Equal(t, 150000.0, goal.SalesPlan)
		repo.AssertExpectations(t)
	})

	t.Run("goal not found returns error", func(t *testing.T) {
		repo := NewMockGoalRepo()
		service := NewGoalServiceTestable(repo)

		repo.On("GetByID", mock.Anything, goalID).Return(nil, errors.New("not found"))

		dto := UpdateGoalDTO{
			SalesPlan: 100000,
		}

		goal, err := service.UpdateGoal(context.Background(), goalID, dto, uuid.New().String(), "")

		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
		require.Nil(t, goal)
		repo.AssertExpectations(t)
	})

	t.Run("updates multiple fields", func(t *testing.T) {
		repo := NewMockGoalRepo()
		service := NewGoalServiceTestable(repo)

		existingGoal := &models.Goal{
			ID:         uuid.MustParse(goalID),
			SalesPlan:  100000,
			LeadsPlan:  10,
		}
		repo.On("GetByID", mock.Anything, goalID).Return(existingGoal, nil)
		repo.On("Update", mock.Anything, mock.Anything).Return(nil)

		dto := UpdateGoalDTO{
			SalesPlan:  200000,
			LeadsPlan:  25,
			Period:     "month",
		}

		goal, err := service.UpdateGoal(context.Background(), goalID, dto, uuid.New().String(), "")

		require.NoError(t, err)
		require.Equal(t, 200000.0, goal.SalesPlan)
		require.Equal(t, 25, goal.LeadsPlan)
		require.Equal(t, models.GoalPeriod("month"), goal.Period)
		repo.AssertExpectations(t)
	})
}

func TestGoalService_GetMyGoal(t *testing.T) {
	assigneeID := uuid.New().String()
	date := time.Now()

	t.Run("success returns goal", func(t *testing.T) {
		repo := NewMockGoalRepo()
		service := NewGoalServiceTestable(repo)

		goal := &models.Goal{
			ID:         uuid.New(),
			AssigneeID: uuid.MustParse(assigneeID),
			SalesPlan:  50000,
		}
		repo.On("GetByAssigneeAndDate", mock.Anything, assigneeID, date).Return(goal, nil)

		result, err := service.GetMyGoal(context.Background(), assigneeID, date)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 50000.0, result.SalesPlan)
		repo.AssertExpectations(t)
	})

	t.Run("not found returns error", func(t *testing.T) {
		repo := NewMockGoalRepo()
		service := NewGoalServiceTestable(repo)

		repo.On("GetByAssigneeAndDate", mock.Anything, assigneeID, date).Return(nil, errors.New("not found"))

		result, err := service.GetMyGoal(context.Background(), assigneeID, date)

		require.Error(t, err)
		require.Nil(t, result)
		repo.AssertExpectations(t)
	})
}

func TestGoalService_GetVisibleGoals(t *testing.T) {
	t.Run("success returns visible goals", func(t *testing.T) {
		repo := NewMockGoalRepo()
		service := NewGoalServiceTestable(repo)

		goals := []models.Goal{
			{ID: uuid.New(), SalesPlan: 100000},
			{ID: uuid.New(), SalesPlan: 200000},
		}
		repo.On("ListVisibleForUser", mock.Anything, "user123", "franchiser", "tenant123").Return(goals, nil)

		result, err := service.GetVisibleGoals(context.Background(), "user123", "franchiser", "tenant123")

		require.NoError(t, err)
		require.Len(t, result, 2)
		repo.AssertExpectations(t)
	})

	t.Run("empty result returns empty array", func(t *testing.T) {
		repo := NewMockGoalRepo()
		service := NewGoalServiceTestable(repo)

		repo.On("ListVisibleForUser", mock.Anything, "user123", "salon_manager", "").Return([]models.Goal{}, nil)

		result, err := service.GetVisibleGoals(context.Background(), "user123", "salon_manager", "")

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result, 0)
		repo.AssertExpectations(t)
	})
}

func TestGoalService_DeleteGoal(t *testing.T) {
	goalID := "goal-123"

	t.Run("success deletes goal", func(t *testing.T) {
		repo := NewMockGoalRepo()
		service := NewGoalServiceTestable(repo)

		repo.On("Delete", mock.Anything, goalID).Return(nil)

		err := service.DeleteGoal(context.Background(), goalID)

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("delete non-existent returns error", func(t *testing.T) {
		repo := NewMockGoalRepo()
		service := NewGoalServiceTestable(repo)

		repo.On("Delete", mock.Anything, goalID).Return(errors.New("record not found"))

		err := service.DeleteGoal(context.Background(), goalID)

		require.Error(t, err)
		require.Equal(t, "record not found", err.Error())
		repo.AssertExpectations(t)
	})
}