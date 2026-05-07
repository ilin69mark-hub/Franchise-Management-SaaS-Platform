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

type MockChecklistRepo struct {
	mock.Mock
}

func NewMockChecklistRepo() *MockChecklistRepo {
	return &MockChecklistRepo{}
}

func (m *MockChecklistRepo) FindAllGlobal(ctx context.Context, status, priority string, isArchive bool) ([]models.Checklist, error) {
	args := m.Called(ctx, status, priority, isArchive)
	return args.Get(0).([]models.Checklist), args.Error(1)
}

func (m *MockChecklistRepo) FindUserTasks(ctx context.Context, userID uuid.UUID, status, priority string, isArchive bool) ([]models.Checklist, error) {
	args := m.Called(ctx, userID, status, priority, isArchive)
	return args.Get(0).([]models.Checklist), args.Error(1)
}

func (m *MockChecklistRepo) FindAll(ctx context.Context, status, priority string) ([]models.Checklist, error) {
	args := m.Called(ctx, status, priority)
	return args.Get(0).([]models.Checklist), args.Error(1)
}

func (m *MockChecklistRepo) GetChecklistByID(ctx context.Context, id uuid.UUID) (*models.Checklist, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Checklist), args.Error(1)
}

func (m *MockChecklistRepo) CreateChecklist(ctx context.Context, chk *models.Checklist) error {
	args := m.Called(ctx, chk)
	return args.Error(0)
}

func (m *MockChecklistRepo) UpdateChecklist(ctx context.Context, chk *models.Checklist) error {
	args := m.Called(ctx, chk)
	return args.Error(0)
}

func (m *MockChecklistRepo) DeleteChecklist(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type ChecklistRepoInterface interface {
	FindAllGlobal(ctx context.Context, status, priority string, isArchive bool) ([]models.Checklist, error)
	FindUserTasks(ctx context.Context, userID uuid.UUID, status, priority string, isArchive bool) ([]models.Checklist, error)
	FindAll(ctx context.Context, status, priority string) ([]models.Checklist, error)
	GetChecklistByID(ctx context.Context, id uuid.UUID) (*models.Checklist, error)
	CreateChecklist(ctx context.Context, chk *models.Checklist) error
	UpdateChecklist(ctx context.Context, chk *models.Checklist) error
	DeleteChecklist(ctx context.Context, id uuid.UUID) error
}

type ChecklistServiceTestable struct {
	repo ChecklistRepoInterface
}

func NewChecklistServiceTestable(repo ChecklistRepoInterface) *ChecklistServiceTestable {
	return &ChecklistServiceTestable{repo: repo}
}

func (s *ChecklistServiceTestable) GetUserChecklists(ctx context.Context, user *models.User, status, priority string, isArchive bool) ([]models.Checklist, error) {
	if user.Role == models.RoleSuperAdmin {
		return s.repo.FindAllGlobal(ctx, status, priority, isArchive)
	}
	return s.repo.FindUserTasks(ctx, user.ID, status, priority, isArchive)
}

func (s *ChecklistServiceTestable) GetAllGlobal(ctx context.Context, status, priority string) ([]models.Checklist, error) {
	return s.repo.FindAllGlobal(ctx, status, priority, false)
}

func (s *ChecklistServiceTestable) GetAllChecklists(ctx context.Context, status, priority string) ([]models.Checklist, error) {
	return s.repo.FindAll(ctx, status, priority)
}

func (s *ChecklistServiceTestable) GetChecklistByID(ctx context.Context, id uuid.UUID) (*models.Checklist, error) {
	return s.repo.GetChecklistByID(ctx, id)
}

func (s *ChecklistServiceTestable) CreateChecklist(ctx context.Context, chk *models.Checklist) error {
	chk.CreatedAt = time.Now()
	chk.UpdatedAt = time.Now()
	return s.repo.CreateChecklist(ctx, chk)
}

func (s *ChecklistServiceTestable) UpdateChecklist(ctx context.Context, chk *models.Checklist) error {
	chk.UpdatedAt = time.Now()
	return s.repo.UpdateChecklist(ctx, chk)
}

func (s *ChecklistServiceTestable) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	chk, err := s.repo.GetChecklistByID(ctx, id)
	if err != nil {
		return err
	}
	chk.Status = status
	return s.repo.UpdateChecklist(ctx, chk)
}

func (s *ChecklistServiceTestable) CompleteChecklist(ctx context.Context, id uuid.UUID) error {
	return s.UpdateStatus(ctx, id, "completed")
}

func (s *ChecklistServiceTestable) DeleteChecklist(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteChecklist(ctx, id)
}

func TestChecklistService_GetUserChecklists(t *testing.T) {
	t.Run("superadmin sees all global", func(t *testing.T) {
		repo := NewMockChecklistRepo()
		service := NewChecklistServiceTestable(repo)

		user := &models.User{ID: uuid.New(), Role: models.RoleSuperAdmin}
		checklists := []models.Checklist{{ID: uuid.New(), Title: "Global Task"}}
		repo.On("FindAllGlobal", mock.Anything, "pending", "high", false).Return(checklists, nil)

		result, err := service.GetUserChecklists(context.Background(), user, "pending", "high", false)

		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Equal(t, "Global Task", result[0].Title)
		repo.AssertExpectations(t)
	})

	t.Run("regular user sees own tasks", func(t *testing.T) {
		repo := NewMockChecklistRepo()
		service := NewChecklistServiceTestable(repo)

		userID := uuid.New()
		user := &models.User{ID: userID, Role: models.RoleDealer}
		checklists := []models.Checklist{{ID: uuid.New(), Title: "My Task"}}
		repo.On("FindUserTasks", mock.Anything, userID, "pending", "", false).Return(checklists, nil)

		result, err := service.GetUserChecklists(context.Background(), user, "pending", "", false)

		require.NoError(t, err)
		require.Len(t, result, 1)
		repo.AssertExpectations(t)
	})
}

func TestChecklistService_GetChecklistByID(t *testing.T) {
	t.Run("success returns checklist", func(t *testing.T) {
		repo := NewMockChecklistRepo()
		service := NewChecklistServiceTestable(repo)

		id := uuid.New()
		checklist := &models.Checklist{ID: id, Title: "Test Task"}
		repo.On("GetChecklistByID", mock.Anything, id).Return(checklist, nil)

		result, err := service.GetChecklistByID(context.Background(), id)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "Test Task", result.Title)
		repo.AssertExpectations(t)
	})

	t.Run("not found returns error", func(t *testing.T) {
		repo := NewMockChecklistRepo()
		service := NewChecklistServiceTestable(repo)

		id := uuid.New()
		repo.On("GetChecklistByID", mock.Anything, id).Return(nil, errors.New("not found"))

		result, err := service.GetChecklistByID(context.Background(), id)

		require.Error(t, err)
		require.Nil(t, result)
		repo.AssertExpectations(t)
	})
}

func TestChecklistService_CreateChecklist(t *testing.T) {
	t.Run("success creates checklist", func(t *testing.T) {
		repo := NewMockChecklistRepo()
		service := NewChecklistServiceTestable(repo)

		chk := &models.Checklist{Title: "New Task", Status: "pending"}
		repo.On("CreateChecklist", mock.Anything, mock.Anything).Return(nil)

		err := service.CreateChecklist(context.Background(), chk)

		require.NoError(t, err)
		require.NotZero(t, chk.CreatedAt)
		repo.AssertExpectations(t)
	})

	t.Run("repo error returns error", func(t *testing.T) {
		repo := NewMockChecklistRepo()
		service := NewChecklistServiceTestable(repo)

		chk := &models.Checklist{Title: "Task"}
		repo.On("CreateChecklist", mock.Anything, mock.Anything).Return(errors.New("db error"))

		err := service.CreateChecklist(context.Background(), chk)

		require.Error(t, err)
		require.Equal(t, "db error", err.Error())
		repo.AssertExpectations(t)
	})
}

func TestChecklistService_UpdateChecklist(t *testing.T) {
	t.Run("success updates checklist", func(t *testing.T) {
		repo := NewMockChecklistRepo()
		service := NewChecklistServiceTestable(repo)

		chk := &models.Checklist{ID: uuid.New(), Title: "Updated"}
		repo.On("UpdateChecklist", mock.Anything, mock.Anything).Return(nil)

		err := service.UpdateChecklist(context.Background(), chk)

		require.NoError(t, err)
		require.NotZero(t, chk.UpdatedAt)
		repo.AssertExpectations(t)
	})
}

func TestChecklistService_UpdateStatus(t *testing.T) {
	t.Run("success changes status", func(t *testing.T) {
		repo := NewMockChecklistRepo()
		service := NewChecklistServiceTestable(repo)

		id := uuid.New()
		chk := &models.Checklist{ID: id, Status: "pending"}
		repo.On("GetChecklistByID", mock.Anything, id).Return(chk, nil)
		repo.On("UpdateChecklist", mock.Anything, mock.Anything).Return(nil)

		err := service.UpdateStatus(context.Background(), id, "in_progress")

		require.NoError(t, err)
		require.Equal(t, "in_progress", chk.Status)
		repo.AssertExpectations(t)
	})

	t.Run("not found returns error", func(t *testing.T) {
		repo := NewMockChecklistRepo()
		service := NewChecklistServiceTestable(repo)

		id := uuid.New()
		repo.On("GetChecklistByID", mock.Anything, id).Return(nil, errors.New("not found"))

		err := service.UpdateStatus(context.Background(), id, "completed")

		require.Error(t, err)
		repo.AssertExpectations(t)
	})
}

func TestChecklistService_CompleteChecklist(t *testing.T) {
	t.Run("success marks as completed", func(t *testing.T) {
		repo := NewMockChecklistRepo()
		service := NewChecklistServiceTestable(repo)

		id := uuid.New()
		chk := &models.Checklist{ID: id, Status: "in_progress"}
		repo.On("GetChecklistByID", mock.Anything, id).Return(chk, nil)
		repo.On("UpdateChecklist", mock.Anything, mock.Anything).Return(nil)

		err := service.CompleteChecklist(context.Background(), id)

		require.NoError(t, err)
		require.Equal(t, "completed", chk.Status)
		repo.AssertExpectations(t)
	})
}

func TestChecklistService_DeleteChecklist(t *testing.T) {
	t.Run("success deletes checklist", func(t *testing.T) {
		repo := NewMockChecklistRepo()
		service := NewChecklistServiceTestable(repo)

		id := uuid.New()
		repo.On("DeleteChecklist", mock.Anything, id).Return(nil)

		err := service.DeleteChecklist(context.Background(), id)

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("repo error returns error", func(t *testing.T) {
		repo := NewMockChecklistRepo()
		service := NewChecklistServiceTestable(repo)

		id := uuid.New()
		repo.On("DeleteChecklist", mock.Anything, id).Return(errors.New("cannot delete"))

		err := service.DeleteChecklist(context.Background(), id)

		require.Error(t, err)
		repo.AssertExpectations(t)
	})
}

func TestChecklistService_GetAllGlobal(t *testing.T) {
	t.Run("success returns global checklists", func(t *testing.T) {
		repo := NewMockChecklistRepo()
		service := NewChecklistServiceTestable(repo)

		checklists := []models.Checklist{{ID: uuid.New(), Title: "Global"}}
		repo.On("FindAllGlobal", mock.Anything, "pending", "high", false).Return(checklists, nil)

		result, err := service.GetAllGlobal(context.Background(), "pending", "high")

		require.NoError(t, err)
		require.Len(t, result, 1)
		repo.AssertExpectations(t)
	})
}

func TestChecklistService_GetAllChecklists(t *testing.T) {
	t.Run("success returns all checklists", func(t *testing.T) {
		repo := NewMockChecklistRepo()
		service := NewChecklistServiceTestable(repo)

		checklists := []models.Checklist{{ID: uuid.New(), Title: "Task1"}}
		repo.On("FindAll", mock.Anything, "pending", "medium").Return(checklists, nil)

		result, err := service.GetAllChecklists(context.Background(), "pending", "medium")

		require.NoError(t, err)
		require.Len(t, result, 1)
		repo.AssertExpectations(t)
	})
}