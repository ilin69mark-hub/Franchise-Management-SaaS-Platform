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

type MockChecklistRepo struct {
	mock.Mock
}

func (m *MockChecklistRepo) FindAllGlobal(ctx context.Context, status, priority string, isArchive bool) ([]models.Checklist, error) {
	args := m.Called(ctx, status, priority, isArchive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Checklist), args.Error(1)
}

func (m *MockChecklistRepo) FindUserTasks(ctx context.Context, userID uuid.UUID, status, priority string, isArchive bool) ([]models.Checklist, error) {
	args := m.Called(ctx, userID, status, priority, isArchive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Checklist), args.Error(1)
}

func (m *MockChecklistRepo) FindAll(ctx context.Context, status, priority string) ([]models.Checklist, error) {
	args := m.Called(ctx, status, priority)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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

func (m *MockChecklistRepo) GetTaskStats() (int64, int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func TestChecklistService_GetUserChecklists_SuperAdmin(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	user := &models.User{ID: uuid.New(), Role: models.RoleSuperAdmin}
	expected := []models.Checklist{{ID: uuid.New(), Title: "Global Task"}}

	mockRepo.On("FindAllGlobal", mock.Anything, "pending", "", false).Return(expected, nil)

	result, err := service.GetUserChecklists(context.Background(), user, "pending", "", false)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_GetUserChecklists_RegularUser(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	user := &models.User{ID: uuid.New(), Role: models.RoleDealer}
	expected := []models.Checklist{{ID: uuid.New(), Title: "My Task"}}

	mockRepo.On("FindUserTasks", mock.Anything, user.ID, "pending", "", false).Return(expected, nil)

	result, err := service.GetUserChecklists(context.Background(), user, "pending", "", false)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_GetUserChecklists_Error(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	user := &models.User{ID: uuid.New(), Role: models.RoleSuperAdmin}

	mockRepo.On("FindAllGlobal", mock.Anything, "pending", "", false).Return(nil, errors.New("db error"))

	result, err := service.GetUserChecklists(context.Background(), user, "pending", "", false)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_GetAllGlobal_Success(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	expected := []models.Checklist{{ID: uuid.New(), Title: "Task 1"}}

	mockRepo.On("FindAllGlobal", mock.Anything, "pending", "high", false).Return(expected, nil)

	result, err := service.GetAllGlobal(context.Background(), "pending", "high")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_GetAllChecklists_Success(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	expected := []models.Checklist{{ID: uuid.New(), Title: "Task 1"}}

	mockRepo.On("FindAll", mock.Anything, "pending", "high").Return(expected, nil)

	result, err := service.GetAllChecklists(context.Background(), "pending", "high")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_GetChecklistByID_Success(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	checklistID := uuid.New()
	expected := &models.Checklist{ID: checklistID, Title: "Test Task"}

	mockRepo.On("GetChecklistByID", mock.Anything, checklistID).Return(expected, nil)

	result, err := service.GetChecklistByID(context.Background(), checklistID)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_GetChecklistByID_NotFound(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	checklistID := uuid.New()

	mockRepo.On("GetChecklistByID", mock.Anything, checklistID).Return(nil, errors.New("record not found"))

	result, err := service.GetChecklistByID(context.Background(), checklistID)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_CreateChecklist_Success(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	checklist := &models.Checklist{
		Title:       "New Task",
		Description: "Test description",
	}

	mockRepo.On("CreateChecklist", mock.Anything, mock.MatchedBy(func(c *models.Checklist) bool {
		return c.Title == "New Task" && c.CreatedAt.IsZero() == false
	})).Return(nil)

	err := service.CreateChecklist(context.Background(), checklist)

	assert.NoError(t, err)
	assert.False(t, checklist.CreatedAt.IsZero())
	assert.False(t, checklist.UpdatedAt.IsZero())
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_CreateChecklist_Error(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	checklist := &models.Checklist{Title: "New Task"}

	mockRepo.On("CreateChecklist", mock.Anything, mock.Anything).Return(errors.New("db error"))

	err := service.CreateChecklist(context.Background(), checklist)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_UpdateChecklist_Success(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	checklist := &models.Checklist{ID: uuid.New(), Title: "Updated Task"}

	mockRepo.On("UpdateChecklist", mock.Anything, mock.MatchedBy(func(c *models.Checklist) bool {
		return c.Title == "Updated Task" && c.UpdatedAt.IsZero() == false
	})).Return(nil)

	err := service.UpdateChecklist(context.Background(), checklist)

	assert.NoError(t, err)
	assert.False(t, checklist.UpdatedAt.IsZero())
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_UpdateChecklist_Error(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	checklist := &models.Checklist{ID: uuid.New(), Title: "Updated Task"}

	mockRepo.On("UpdateChecklist", mock.Anything, mock.Anything).Return(errors.New("db error"))

	err := service.UpdateChecklist(context.Background(), checklist)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_UpdateStatus_Success(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	checklistID := uuid.New()
	existingChk := &models.Checklist{ID: checklistID, Status: "pending", UpdatedAt: time.Now().Add(-time.Hour)}

	mockRepo.On("GetChecklistByID", mock.Anything, checklistID).Return(existingChk, nil)
	mockRepo.On("UpdateChecklist", mock.Anything, mock.MatchedBy(func(c *models.Checklist) bool {
		return c.Status == "completed"
	})).Return(nil)

	err := service.UpdateStatus(context.Background(), checklistID, "completed")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_UpdateStatus_NotFound(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	checklistID := uuid.New()

	mockRepo.On("GetChecklistByID", mock.Anything, checklistID).Return(nil, errors.New("not found"))

	err := service.UpdateStatus(context.Background(), checklistID, "completed")

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_CompleteChecklist(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	checklistID := uuid.New()
	existingChk := &models.Checklist{ID: checklistID, Status: "pending"}

	mockRepo.On("GetChecklistByID", mock.Anything, checklistID).Return(existingChk, nil)
	mockRepo.On("UpdateChecklist", mock.Anything, mock.Anything).Return(nil)

	err := service.CompleteChecklist(context.Background(), checklistID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_DeleteChecklist_Success(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	checklistID := uuid.New()

	mockRepo.On("DeleteChecklist", mock.Anything, checklistID).Return(nil)

	err := service.DeleteChecklist(context.Background(), checklistID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestChecklistService_DeleteChecklist_Error(t *testing.T) {
	mockRepo := new(MockChecklistRepo)
	service := NewChecklistService(mockRepo)

	checklistID := uuid.New()

	mockRepo.On("DeleteChecklist", mock.Anything, checklistID).Return(errors.New("db error"))

	err := service.DeleteChecklist(context.Background(), checklistID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}