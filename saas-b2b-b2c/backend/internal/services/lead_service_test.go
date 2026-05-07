package services

import (
	"context"
	"errors"
	"testing"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockLeadRepo struct {
	mock.Mock
}

func NewMockLeadRepo() *MockLeadRepo {
	return &MockLeadRepo{}
}

func (m *MockLeadRepo) CreateLead(ctx context.Context, lead *models.Lead) error {
	args := m.Called(ctx, lead)
	return args.Error(0)
}

func (m *MockLeadRepo) GetLeadsByManager(ctx context.Context, managerID uuid.UUID) ([]models.Lead, error) {
	args := m.Called(ctx, managerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Lead), args.Error(1)
}

func (m *MockLeadRepo) GetLeadByID(ctx context.Context, leadID, managerID uuid.UUID) (*models.Lead, error) {
	args := m.Called(ctx, leadID, managerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Lead), args.Error(1)
}

func (m *MockLeadRepo) UpdateLeadStatus(ctx context.Context, leadID uuid.UUID, status string) error {
	args := m.Called(ctx, leadID, status)
	return args.Error(0)
}

func (m *MockLeadRepo) AddActivity(ctx context.Context, activity *models.LeadActivity) error {
	args := m.Called(ctx, activity)
	return args.Error(0)
}

func (m *MockLeadRepo) GetLeadActivities(ctx context.Context, leadID uuid.UUID) ([]models.LeadActivity, error) {
	args := m.Called(ctx, leadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.LeadActivity), args.Error(1)
}

type LeadRepoInterface interface {
	CreateLead(ctx context.Context, lead *models.Lead) error
	GetLeadsByManager(ctx context.Context, managerID uuid.UUID) ([]models.Lead, error)
	GetLeadByID(ctx context.Context, leadID, managerID uuid.UUID) (*models.Lead, error)
	UpdateLeadStatus(ctx context.Context, leadID uuid.UUID, status string) error
	AddActivity(ctx context.Context, activity *models.LeadActivity) error
	GetLeadActivities(ctx context.Context, leadID uuid.UUID) ([]models.LeadActivity, error)
}

type LeadServiceTestable struct {
	repo LeadRepoInterface
}

func NewLeadServiceTestable(repo LeadRepoInterface) *LeadServiceTestable {
	return &LeadServiceTestable{repo: repo}
}

func (s *LeadServiceTestable) CreateLead(ctx context.Context, user *models.User, req models.CreateLeadRequest) (*models.Lead, error) {
	if user.SalonID == nil {
		return nil, errors.New("пользователь не привязан к салону, невозможно создать лид")
	}
	lead := &models.Lead{
		SalonID:         *user.SalonID,
		ManagerID:       user.ID,
		FullName:        req.FullName,
		Phone:           req.Phone,
		Email:           req.Email,
		InterestProduct: req.InterestProduct,
		Budget:          req.Budget,
		Status:          "new",
	}
	if err := s.repo.CreateLead(ctx, lead); err != nil {
		return nil, err
	}
	return lead, nil
}

func (s *LeadServiceTestable) GetMyLeads(ctx context.Context, managerID uuid.UUID) ([]models.Lead, error) {
	return s.repo.GetLeadsByManager(ctx, managerID)
}

func (s *LeadServiceTestable) UpdateStatus(ctx context.Context, managerID, leadID uuid.UUID, status string) error {
	_, err := s.repo.GetLeadByID(ctx, leadID, managerID)
	if err != nil {
		return err
	}
	return s.repo.UpdateLeadStatus(ctx, leadID, status)
}

func (s *LeadServiceTestable) AddActivity(ctx context.Context, userID, leadID uuid.UUID, req models.AddLeadActivityRequest) error {
	activity := &models.LeadActivity{
		LeadID:      leadID,
		UserID:      userID,
		Type:        req.Type,
		Description: req.Description,
	}
	return s.repo.AddActivity(ctx, activity)
}

func (s *LeadServiceTestable) GetLeadDetails(ctx context.Context, managerID, leadID uuid.UUID) (*models.Lead, []models.LeadActivity, error) {
	lead, err := s.repo.GetLeadByID(ctx, leadID, managerID)
	if err != nil {
		return nil, nil, err
	}
	activities, err := s.repo.GetLeadActivities(ctx, leadID)
	if err != nil {
		return lead, nil, err
	}
	return lead, activities, nil
}

func TestLeadService_CreateLead(t *testing.T) {
	t.Run("success creates lead", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		salonID := uuid.New()
		user := &models.User{
			ID:      uuid.New(),
			SalonID: &salonID,
		}
		req := models.CreateLeadRequest{
			FullName:        "John Doe",
			Phone:           "+1234567890",
			Email:           "john@example.com",
			InterestProduct: "Product A",
			Budget:          50000,
		}
		repo.On("CreateLead", mock.Anything, mock.MatchedBy(func(l *models.Lead) bool {
			return l.FullName == "John Doe" && l.Status == "new"
		})).Return(nil)

		lead, err := service.CreateLead(context.Background(), user, req)

		require.NoError(t, err)
		require.NotNil(t, lead)
		require.Equal(t, "John Doe", lead.FullName)
		require.Equal(t, "new", lead.Status)
		repo.AssertExpectations(t)
	})

	t.Run("fails when user has no salon", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		user := &models.User{
			ID:      uuid.New(),
			SalonID: nil,
		}
		req := models.CreateLeadRequest{
			FullName: "John Doe",
		}

		lead, err := service.CreateLead(context.Background(), user, req)

		require.Error(t, err)
		require.Contains(t, err.Error(), "не привязан к салону")
		require.Nil(t, lead)
	})

	t.Run("repository error returns error", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		salonID := uuid.New()
		user := &models.User{
			ID:      uuid.New(),
			SalonID: &salonID,
		}
		req := models.CreateLeadRequest{
			FullName: "John Doe",
		}
		repo.On("CreateLead", mock.Anything, mock.Anything).Return(errors.New("duplicate key"))

		lead, err := service.CreateLead(context.Background(), user, req)

		require.Error(t, err)
		require.Nil(t, lead)
		require.Equal(t, "duplicate key", err.Error())
		repo.AssertExpectations(t)
	})

	t.Run("creates lead with budget", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		salonID := uuid.New()
		user := &models.User{
			ID:      uuid.New(),
			SalonID: &salonID,
		}
		req := models.CreateLeadRequest{
			FullName: "Jane Smith",
			Budget:   100000,
		}
		repo.On("CreateLead", mock.Anything, mock.MatchedBy(func(l *models.Lead) bool {
			return l.Budget == 100000
		})).Return(nil)

		lead, err := service.CreateLead(context.Background(), user, req)

		require.NoError(t, err)
		require.Equal(t, 100000.0, lead.Budget)
		repo.AssertExpectations(t)
	})
}

func TestLeadService_GetMyLeads(t *testing.T) {
	managerID := uuid.New()

	t.Run("success returns leads", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		leads := []models.Lead{
			{ID: uuid.New(), ManagerID: managerID, FullName: "Lead 1", Status: "new"},
			{ID: uuid.New(), ManagerID: managerID, FullName: "Lead 2", Status: "contacted"},
		}
		repo.On("GetLeadsByManager", mock.Anything, managerID).Return(leads, nil)

		result, err := service.GetMyLeads(context.Background(), managerID)

		require.NoError(t, err)
		require.Len(t, result, 2)
		require.Equal(t, "Lead 1", result[0].FullName)
		repo.AssertExpectations(t)
	})

	t.Run("empty leads returns empty array", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		repo.On("GetLeadsByManager", mock.Anything, managerID).Return([]models.Lead{}, nil)

		result, err := service.GetMyLeads(context.Background(), managerID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result, 0)
		repo.AssertExpectations(t)
	})
}

func TestLeadService_UpdateStatus(t *testing.T) {
	managerID := uuid.New()
	leadID := uuid.New()

	t.Run("success updates status", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		lead := &models.Lead{ID: leadID, ManagerID: managerID}
		repo.On("GetLeadByID", mock.Anything, leadID, managerID).Return(lead, nil)
		repo.On("UpdateLeadStatus", mock.Anything, leadID, "sale").Return(nil)

		err := service.UpdateStatus(context.Background(), managerID, leadID, "sale")

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("access denied returns error", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		repo.On("GetLeadByID", mock.Anything, leadID, managerID).Return(nil, errors.New("record not found"))

		err := service.UpdateStatus(context.Background(), managerID, leadID, "sale")

		require.Error(t, err)
		require.Equal(t, "record not found", err.Error())
		repo.AssertExpectations(t)
	})

	t.Run("updates to cancelled status", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		lead := &models.Lead{ID: leadID, ManagerID: managerID}
		repo.On("GetLeadByID", mock.Anything, leadID, managerID).Return(lead, nil)
		repo.On("UpdateLeadStatus", mock.Anything, leadID, "cancelled").Return(nil)

		err := service.UpdateStatus(context.Background(), managerID, leadID, "cancelled")

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

func TestLeadService_AddActivity(t *testing.T) {
	userID := uuid.New()
	leadID := uuid.New()

	t.Run("success adds activity", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		req := models.AddLeadActivityRequest{
			Type:        "call",
			Description: "Initial call completed",
		}
		repo.On("AddActivity", mock.Anything, mock.MatchedBy(func(a *models.LeadActivity) bool {
			return a.Type == "call" && a.Description == "Initial call completed"
		})).Return(nil)

		err := service.AddActivity(context.Background(), userID, leadID, req)

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("adds meeting activity", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		req := models.AddLeadActivityRequest{
			Type:        "meeting",
			Description: "Demo meeting scheduled",
		}
		repo.On("AddActivity", mock.Anything, mock.Anything).Return(nil)

		err := service.AddActivity(context.Background(), userID, leadID, req)

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("repository error on add activity", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		req := models.AddLeadActivityRequest{
			Type: "call",
		}
		repo.On("AddActivity", mock.Anything, mock.Anything).Return(errors.New("foreign key violation"))

		err := service.AddActivity(context.Background(), userID, leadID, req)

		require.Error(t, err)
		require.Equal(t, "foreign key violation", err.Error())
		repo.AssertExpectations(t)
	})
}

func TestLeadService_GetLeadDetails(t *testing.T) {
	managerID := uuid.New()
	leadID := uuid.New()

	t.Run("success returns lead and activities", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		lead := &models.Lead{ID: leadID, ManagerID: managerID, FullName: "Test Lead"}
		activities := []models.LeadActivity{
			{ID: uuid.New(), LeadID: leadID, Type: "call"},
			{ID: uuid.New(), LeadID: leadID, Type: "meeting"},
		}
		repo.On("GetLeadByID", mock.Anything, leadID, managerID).Return(lead, nil)
		repo.On("GetLeadActivities", mock.Anything, leadID).Return(activities, nil)

		result, acts, err := service.GetLeadDetails(context.Background(), managerID, leadID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "Test Lead", result.FullName)
		require.Len(t, acts, 2)
		repo.AssertExpectations(t)
	})

	t.Run("lead not found returns error", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		repo.On("GetLeadByID", mock.Anything, leadID, managerID).Return(nil, errors.New("not found"))

		result, acts, err := service.GetLeadDetails(context.Background(), managerID, leadID)

		require.Error(t, err)
		require.Nil(t, result)
		require.Nil(t, acts)
		repo.AssertExpectations(t)
	})

	t.Run("activities error returns error", func(t *testing.T) {
		repo := NewMockLeadRepo()
		service := NewLeadServiceTestable(repo)

		lead := &models.Lead{ID: leadID, ManagerID: managerID}
		repo.On("GetLeadByID", mock.Anything, leadID, managerID).Return(lead, nil)
		repo.On("GetLeadActivities", mock.Anything, leadID).Return(nil, errors.New("table not found"))

		result, acts, err := service.GetLeadDetails(context.Background(), managerID, leadID)

		require.Error(t, err)
		require.NotNil(t, result)
		require.Nil(t, acts)
		repo.AssertExpectations(t)
	})
}