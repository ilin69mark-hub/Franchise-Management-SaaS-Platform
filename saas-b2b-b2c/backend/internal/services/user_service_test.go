package services

import (
	"errors"
	"testing"

	"franchise-saas-backend/internal/mocks"
	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUserService_GetProfile(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)
	userID := uuid.New()

	tests := []struct {
		name      string
		userID    uuid.UUID
		setupMock func()
		wantUser  bool
		wantErr   bool
	}{
		{
			name:   "success - user profile found",
			userID: userID,
			setupMock: func() {
				user := &models.User{
					ID:        userID,
					Email:     "test@example.com",
					FirstName: "John",
					LastName:  "Doe",
					Role:      models.RoleDealer,
				}
				mockRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil).Once()
			},
			wantUser: true,
			wantErr:  false,
		},
		{
			name:   "failure - user not found",
			userID: uuid.New(),
			setupMock: func() {
				mockRepo.On("GetUserByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, gorm.ErrRecordNotFound).Once()
			},
			wantUser: false,
			wantErr:  true,
		},
		{
			name:   "failure - database error",
			userID: uuid.New(),
			setupMock: func() {
				mockRepo.On("GetUserByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("database error")).Once()
			},
			wantUser: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.setupMock()

			user, err := service.GetProfile(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdateProfile(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)
	userID := uuid.New()

	tests := []struct {
		name      string
		userID    uuid.UUID
		req       models.UserUpdateRequest
		setupMock func()
		wantErr   bool
	}{
		{
			name:   "success - profile updated",
			userID: userID,
			req:    models.UserUpdateRequest{FirstName: "Jane", LastName: "Smith"},
			setupMock: func() {
				mockRepo.On("UpdateUserFields", mock.Anything, userID, mock.Anything).Return(nil).Once()
				user := &models.User{ID: userID, FirstName: "Jane", LastName: "Smith"}
				mockRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "failure - user not found",
			userID: uuid.New(),
			req:    models.UserUpdateRequest{FirstName: "Jane"},
			setupMock: func() {
				mockRepo.On("UpdateUserFields", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.Anything).Return(gorm.ErrRecordNotFound).Once()
			},
			wantErr: true,
		},
		{
			name:   "failure - database error",
			userID: uuid.New(),
			req:    models.UserUpdateRequest{FirstName: "Jane"},
			setupMock: func() {
				mockRepo.On("UpdateUserFields", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.Anything).Return(errors.New("database error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.setupMock()

			user, err := service.UpdateProfile(tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetEmployees(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)
	tenantID := uuid.New()

	tests := []struct {
		name      string
		tenantID  uuid.UUID
		setupMock func()
		wantCount int
		wantErr   bool
	}{
		{
			name:     "success - employees found",
			tenantID: tenantID,
			setupMock: func() {
				employees := []models.User{
					{ID: uuid.New(), Email: "emp1@example.com", Role: models.RoleDealer},
					{ID: uuid.New(), Email: "emp2@example.com", Role: models.RoleDealer},
				}
				mockRepo.On("FindUsersByTenantID", mock.Anything, tenantID).Return(employees, nil).Once()
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:     "success - no employees",
			tenantID: tenantID,
			setupMock: func() {
				mockRepo.On("FindUsersByTenantID", mock.Anything, tenantID).Return([]models.User{}, nil).Once()
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:     "failure - database error",
			tenantID: uuid.New(),
			setupMock: func() {
				mockRepo.On("FindUsersByTenantID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("database error")).Once()
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.setupMock()

			employees, err := service.GetEmployees(tt.tenantID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, employees, tt.wantCount)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetAllUsersGlobal(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantCount int
		wantErr   bool
	}{
		{
			name: "success - all users retrieved",
			setupMock: func() {
				users := []models.User{
					{ID: uuid.New(), Email: "admin@example.com", Role: models.RoleSuperAdmin},
					{ID: uuid.New(), Email: "franchiser@example.com", Role: models.RoleFranchisor},
					{ID: uuid.New(), Email: "dealer@example.com", Role: models.RoleDealer},
				}
				mockRepo.On("FindAllGlobal", mock.Anything).Return(users, nil).Once()
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "failure - database error",
			setupMock: func() {
				mockRepo.On("FindAllGlobal", mock.Anything).Return(nil, errors.New("database error")).Once()
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.setupMock()

			users, err := service.GetAllUsersGlobal()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, users, tt.wantCount)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_ChangePassword(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)

	tests := []struct {
		name        string
		userID      uuid.UUID
		setupMock   func()
		wantErr     bool
		errContains string
	}{
		{
			name:   "failure - user not found",
			userID: uuid.New(),
			setupMock: func() {
				mockRepo.On("GetUserByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, gorm.ErrRecordNotFound).Once()
			},
			wantErr:     true,
			errContains: "record not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.setupMock()

			err := service.ChangePassword(tt.userID, "oldpass-is-long-1", "newpass-is-long-12")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_CreateEmployee(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)

	tests := []struct {
		name        string
		req         models.CreateEmployeeRequest
		tenantID    uuid.UUID
		creatorID   uuid.UUID
		creatorRole string
		setupMock   func()
		wantErr     bool
		errMessage  string
	}{
		{
			name: "success - employee created by super_admin",
			req: models.CreateEmployeeRequest{
				Email:     "newemployee@example.com",
				Password:  "secure-password-12",
				Role:      models.RoleDealer,
				FirstName: "New",
				LastName:  "Employee",
			},
			tenantID:    uuid.New(),
			creatorID:   uuid.New(),
			creatorRole: string(models.RoleSuperAdmin),
			setupMock: func() {
				mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "failure - permission denied - creating franchiser",
			req: models.CreateEmployeeRequest{
				Email:    "franchiser@example.com",
				Password: "secure-password-12",
				Role:     models.RoleFranchisor,
			},
			tenantID:    uuid.New(),
			creatorID:   uuid.New(),
			creatorRole: string(models.RoleDealer),
			setupMock:   func() {},
			wantErr:     true,
			errMessage:  "permission denied",
		},
		{
			name: "failure - franchiser role requires tenant_id",
			req: models.CreateEmployeeRequest{
				Email:    "franchiser@example.com",
				Password: "secure-password-12",
				Role:     models.RoleFranchisor,
			},
			tenantID:    uuid.Nil,
			creatorID:   uuid.New(),
			creatorRole: string(models.RoleSuperAdmin),
			setupMock:   func() {},
			wantErr:     true,
			errMessage:  "tenant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			tt.setupMock()

			user, err := service.CreateEmployee(tt.req, tt.tenantID, tt.creatorID, tt.creatorRole)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdateEmployee_Success(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)

	userID := uuid.New()
	tenantID := uuid.New()
	req := models.UpdateEmployeeRequest{
		FirstName: "John",
		LastName:  "Doe",
		Role:      models.RoleDealerManager,
	}

	mockRepo.On("FindUserByIDAndTenant", mock.Anything, userID, tenantID).Return(&models.User{ID: userID, Role: models.RoleDealerManager}, nil)
	mockRepo.On("UpdateUserFields", mock.Anything, userID, mock.Anything).Return(nil)
	mockRepo.On("GetUserByID", mock.Anything, userID).Return(&models.User{ID: userID, FirstName: "John"}, nil)

	user, err := service.UpdateEmployee(userID, tenantID, req, string(models.RoleDealer))

	assert.NoError(t, err)
	assert.NotNil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateEmployee_NotFound(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)

	userID := uuid.New()
	tenantID := uuid.New()
	req := models.UpdateEmployeeRequest{FirstName: "John"}

	mockRepo.On("FindUserByIDAndTenant", mock.Anything, userID, tenantID).Return(nil, errors.New("not found"))

	user, err := service.UpdateEmployee(userID, tenantID, req, string(models.RoleFranchisor))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateEmployee_InvalidRole(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)

	userID := uuid.New()
	tenantID := uuid.New()
	req := models.UpdateEmployeeRequest{Role: models.RoleSuperAdmin}

	// S8: super_admin грузит цель напрямую через GetUserByID.
	mockRepo.On("GetUserByID", mock.Anything, userID).Return(&models.User{ID: userID}, nil)

	user, err := service.UpdateEmployee(userID, tenantID, req, string(models.RoleSuperAdmin))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
	assert.Nil(t, user)
	mockRepo.AssertNotCalled(t, "UpdateUserFields")
}

func TestUserService_DeleteEmployee_Success(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)

	userID := uuid.New()
	tenantID := uuid.New()

	mockRepo.On("FindUserByIDAndTenant", mock.Anything, userID, tenantID).Return(&models.User{ID: userID, Role: models.RoleDealer}, nil)
	mockRepo.On("DeleteUser", mock.Anything, userID).Return(nil)

	err := service.DeleteEmployee(userID, tenantID, string(models.RoleFranchisor))

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_DeleteEmployee_NotFound(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)

	userID := uuid.New()
	tenantID := uuid.New()

	mockRepo.On("FindUserByIDAndTenant", mock.Anything, userID, tenantID).Return(nil, errors.New("not found"))

	err := service.DeleteEmployee(userID, tenantID, string(models.RoleFranchisor))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateEmployee_EscalationDenied(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)

	userID := uuid.New()
	tenantID := uuid.New()
	req := models.UpdateEmployeeRequest{Role: models.RoleFranchisorManager}

	mockRepo.On("FindUserByIDAndTenant", mock.Anything, userID, tenantID).Return(&models.User{ID: userID, Role: models.RoleDealerManager}, nil)

	user, err := service.UpdateEmployee(userID, tenantID, req, string(models.RoleDealer))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
	assert.Nil(t, user)
	mockRepo.AssertNotCalled(t, "UpdateUserFields", mock.Anything, mock.Anything, mock.Anything)
}

func TestUserService_DeleteEmployee_EscalationDenied(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)

	userID := uuid.New()
	tenantID := uuid.New()

	mockRepo.On("FindUserByIDAndTenant", mock.Anything, userID, tenantID).Return(&models.User{ID: userID, Role: models.RoleDealer}, nil)

	err := service.DeleteEmployee(userID, tenantID, string(models.RoleDealerManager))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
	mockRepo.AssertNotCalled(t, "DeleteUser", mock.Anything, mock.Anything)
}

func TestUserService_ChangePassword_ShortRejected(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)

	err := service.ChangePassword(uuid.New(), "oldpass-is-long-1", "short")
	require.Error(t, err)
	require.Contains(t, err.Error(), "12..72")
	mockRepo.AssertNotCalled(t, "GetUserByID", mock.Anything, mock.Anything)
}

func TestUserService_SuperAdminManagesAnyTenant(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)

	userID := uuid.New()
	tenantID := uuid.New()
	target := &models.User{ID: userID, TenantID: &tenantID, Role: models.RoleDealer}

	mockRepo.On("GetUserByID", mock.Anything, userID).Return(target, nil)
	mockRepo.On("UpdateUserFields", mock.Anything, userID, mock.Anything).Return(nil)

	updated, err := service.UpdateEmployee(userID, uuid.Nil, models.UpdateEmployeeRequest{FirstName: "SA"}, string(models.RoleSuperAdmin))

	require.NoError(t, err, "S8: super_admin без сети управляет любым tenant")
	require.NotNil(t, updated)
	mockRepo.AssertExpectations(t)
}

func TestUserService_SuperAdminDeletesAnyTenant(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()
	service := NewUserServiceWithInterface(mockRepo, nil)

	userID := uuid.New()
	tenantID := uuid.New()
	target := &models.User{ID: userID, TenantID: &tenantID, Role: models.RoleDealer}

	mockRepo.On("GetUserByID", mock.Anything, userID).Return(target, nil)
	mockRepo.On("DeleteUser", mock.Anything, userID).Return(nil)

	require.NoError(t, service.DeleteEmployee(userID, uuid.Nil, string(models.RoleSuperAdmin)))
	mockRepo.AssertExpectations(t)
}
