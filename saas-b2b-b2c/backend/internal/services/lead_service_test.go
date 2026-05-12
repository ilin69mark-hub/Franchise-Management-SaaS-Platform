package services

import (
	"context"
	"errors"
	"testing"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockLeadRepository struct {
	mock.Mock
}

func (m *MockLeadRepository) CreateLead(ctx context.Context, lead *models.Lead) error {
	args := m.Called(ctx, lead)
	return args.Error(0)
}

func (m *MockLeadRepository) GetLeadsByManager(ctx context.Context, managerID uuid.UUID) ([]models.Lead, error) {
	args := m.Called(ctx, managerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Lead), args.Error(1)
}

func (m *MockLeadRepository) GetLeadByID(ctx context.Context, leadID, managerID uuid.UUID) (*models.Lead, error) {
	args := m.Called(ctx, leadID, managerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Lead), args.Error(1)
}

func (m *MockLeadRepository) UpdateLeadStatus(ctx context.Context, leadID uuid.UUID, status string) error {
	args := m.Called(ctx, leadID, status)
	return args.Error(0)
}

func (m *MockLeadRepository) AddActivity(ctx context.Context, activity *models.LeadActivity) error {
	args := m.Called(ctx, activity)
	return args.Error(0)
}

func (m *MockLeadRepository) GetLeadActivities(ctx context.Context, leadID uuid.UUID) ([]models.LeadActivity, error) {
	args := m.Called(ctx, leadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.LeadActivity), args.Error(1)
}

func TestLeadService_CreateLead_Success(t *testing.T) {
	mockRepo := new(MockLeadRepository)
	service := NewLeadService(mockRepo)

	user := &models.User{
		ID:     uuid.New(),
		SalonID: ptrToUUID(uuid.New()),
	}
	req := models.CreateLeadRequest{
		FullName: "Test Lead",
		Phone:    "+1234567890",
		Email:    "test@example.com",
	}

	mockRepo.On("CreateLead", mock.Anything, mock.MatchedBy(func(lead *models.Lead) bool {
		return lead.FullName == "Test Lead" && lead.Status == "new"
	})).Return(nil)

	lead, err := service.CreateLead(context.Background(), user, req)

	assert.NoError(t, err)
	assert.Equal(t, "Test Lead", lead.FullName)
	assert.Equal(t, "new", lead.Status)
	mockRepo.AssertExpectations(t)
}

func TestLeadService_CreateLead_NoSalonID(t *testing.T) {
	mockRepo := new(MockLeadRepository)
	service := NewLeadService(mockRepo)

	user := &models.User{
		ID:     uuid.New(),
		SalonID: nil,
	}
	req := models.CreateLeadRequest{
		FullName: "Test Lead",
	}

	lead, err := service.CreateLead(context.Background(), user, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "не привязан к салону")
	assert.Nil(t, lead)
	mockRepo.AssertNotCalled(t, "CreateLead")
}

func TestLeadService_CreateLead_RepoError(t *testing.T) {
	mockRepo := new(MockLeadRepository)
	service := NewLeadService(mockRepo)

	user := &models.User{
		ID:     uuid.New(),
		SalonID: ptrToUUID(uuid.New()),
	}
	req := models.CreateLeadRequest{
		FullName: "Test Lead",
	}

	mockRepo.On("CreateLead", mock.Anything, mock.Anything).Return(errors.New("database error"))

	lead, err := service.CreateLead(context.Background(), user, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, lead)
	mockRepo.AssertExpectations(t)
}

func TestLeadService_GetMyLeads_Success(t *testing.T) {
	mockRepo := new(MockLeadRepository)
	service := NewLeadService(mockRepo)

	managerID := uuid.New()
	expectedLeads := []models.Lead{
		{ID: uuid.New(), FullName: "Lead 1", Status: "new"},
		{ID: uuid.New(), FullName: "Lead 2", Status: "contact"},
	}

	mockRepo.On("GetLeadsByManager", mock.Anything, managerID).Return(expectedLeads, nil)

	leads, err := service.GetMyLeads(context.Background(), managerID)

	assert.NoError(t, err)
	assert.Len(t, leads, 2)
	assert.Equal(t, "Lead 1", leads[0].FullName)
	mockRepo.AssertExpectations(t)
}

func TestLeadService_GetMyLeads_Empty(t *testing.T) {
	mockRepo := new(MockLeadRepository)
	service := NewLeadService(mockRepo)

	managerID := uuid.New()

	mockRepo.On("GetLeadsByManager", mock.Anything, managerID).Return([]models.Lead{}, nil)

	leads, err := service.GetMyLeads(context.Background(), managerID)

	assert.NoError(t, err)
	assert.Len(t, leads, 0)
	mockRepo.AssertExpectations(t)
}

func TestLeadService_GetMyLeads_Error(t *testing.T) {
	mockRepo := new(MockLeadRepository)
	service := NewLeadService(mockRepo)

	managerID := uuid.New()

	mockRepo.On("GetLeadsByManager", mock.Anything, managerID).Return(nil, errors.New("database error"))

	leads, err := service.GetMyLeads(context.Background(), managerID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, leads)
	mockRepo.AssertExpectations(t)
}

func TestLeadService_UpdateStatus_Success(t *testing.T) {
	mockRepo := new(MockLeadRepository)
	service := NewLeadService(mockRepo)

	managerID := uuid.New()
	leadID := uuid.New()

	mockRepo.On("GetLeadByID", mock.Anything, leadID, managerID).Return(&models.Lead{ID: leadID}, nil)
	mockRepo.On("UpdateLeadStatus", mock.Anything, leadID, "contact").Return(nil)

	err := service.UpdateStatus(context.Background(), managerID, leadID, "contact")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLeadService_UpdateStatus_LeadNotFound(t *testing.T) {
	mockRepo := new(MockLeadRepository)
	service := NewLeadService(mockRepo)

	managerID := uuid.New()
	leadID := uuid.New()

	mockRepo.On("GetLeadByID", mock.Anything, leadID, managerID).Return(nil, errors.New("record not found"))

	err := service.UpdateStatus(context.Background(), managerID, leadID, "contact")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
	mockRepo.AssertExpectations(t)
}

func TestLeadService_UpdateStatus_RepoError(t *testing.T) {
	mockRepo := new(MockLeadRepository)
	service := NewLeadService(mockRepo)

	managerID := uuid.New()
	leadID := uuid.New()

	mockRepo.On("GetLeadByID", mock.Anything, leadID, managerID).Return(&models.Lead{ID: leadID}, nil)
	mockRepo.On("UpdateLeadStatus", mock.Anything, leadID, "contact").Return(errors.New("database error"))

	err := service.UpdateStatus(context.Background(), managerID, leadID, "contact")

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLeadService_AddActivity_Success(t *testing.T) {
	mockRepo := new(MockLeadRepository)
	service := NewLeadService(mockRepo)

	userID := uuid.New()
	leadID := uuid.New()
	req := models.AddLeadActivityRequest{
		Type:        "call",
		Description: "Initial call",
	}

	mockRepo.On("AddActivity", mock.Anything, mock.MatchedBy(func(a *models.LeadActivity) bool {
		return a.LeadID == leadID && a.UserID == userID && a.Type == "call"
	})).Return(nil)

	err := service.AddActivity(context.Background(), userID, leadID, req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLeadService_AddActivity_RepoError(t *testing.T) {
	mockRepo := new(MockLeadRepository)
	service := NewLeadService(mockRepo)

	userID := uuid.New()
	leadID := uuid.New()
	req := models.AddLeadActivityRequest{
		Type:        "call",
		Description: "Initial call",
	}

	mockRepo.On("AddActivity", mock.Anything, mock.Anything).Return(errors.New("database error"))

	err := service.AddActivity(context.Background(), userID, leadID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestLeadService_GetLeadDetails_Success(t *testing.T) {
	mockRepo := new(MockLeadRepository)
	service := NewLeadService(mockRepo)

	managerID := uuid.New()
	leadID := uuid.New()
	expectedLead := &models.Lead{ID: leadID, FullName: "Test Lead", Status: "new"}
	expectedActivities := []models.LeadActivity{
		{ID: uuid.New(), LeadID: leadID, Type: "call"},
	}

	mockRepo.On("GetLeadByID", mock.Anything, leadID, managerID).Return(expectedLead, nil)
	mockRepo.On("GetLeadActivities", mock.Anything, leadID).Return(expectedActivities, nil)

	lead, activities, err := service.GetLeadDetails(context.Background(), managerID, leadID)

	assert.NoError(t, err)
	assert.Equal(t, expectedLead, lead)
	assert.Len(t, activities, 1)
	mockRepo.AssertExpectations(t)
}

func TestLeadService_GetLeadDetails_LeadNotFound(t *testing.T) {
	mockRepo := new(MockLeadRepository)
	service := NewLeadService(mockRepo)

	managerID := uuid.New()
	leadID := uuid.New()

	mockRepo.On("GetLeadByID", mock.Anything, leadID, managerID).Return(nil, errors.New("record not found"))

	lead, activities, err := service.GetLeadDetails(context.Background(), managerID, leadID)

	assert.Error(t, err)
	assert.Nil(t, lead)
	assert.Nil(t, activities)
	mockRepo.AssertExpectations(t)
}

func TestLeadService_GetLeadDetails_ActivitiesError(t *testing.T) {
	mockRepo := new(MockLeadRepository)
	service := NewLeadService(mockRepo)

	managerID := uuid.New()
	leadID := uuid.New()
	expectedLead := &models.Lead{ID: leadID, FullName: "Test Lead"}

	mockRepo.On("GetLeadByID", mock.Anything, leadID, managerID).Return(expectedLead, nil)
	mockRepo.On("GetLeadActivities", mock.Anything, leadID).Return(nil, errors.New("database error"))

	_, activities, err := service.GetLeadDetails(context.Background(), managerID, leadID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	assert.Nil(t, activities)
	mockRepo.AssertExpectations(t)
}

func ptrToUUID(id uuid.UUID) *uuid.UUID {
	return &id
}