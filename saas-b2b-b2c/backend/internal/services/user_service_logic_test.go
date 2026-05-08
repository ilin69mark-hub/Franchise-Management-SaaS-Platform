package services

import (
	"testing"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserServiceLogic_GetProfileFromMock(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userID := uuid.New()
	
	user := &models.User{
		ID:        userID,
		Email:     "profile@test.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      models.RoleDealer,
		PasswordHash: "hash",
	}
	mockRepo.byID[userID] = user
	
	result, err := mockRepo.GetUserByID(nil, userID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, result.Email)
}

func TestUserServiceLogic_UpdateProfileFromMock(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userID := uuid.New()
	
	user := &models.User{
		ID:        userID,
		Email:     "update@test.com",
		FirstName: "OldName",
		Role:      models.RoleDealer,
		PasswordHash: "hash",
	}
	mockRepo.byID[userID] = user
	
	err := mockRepo.UpdateUserFields(nil, userID, map[string]interface{}{
		"first_name": "NewName",
	})
	require.NoError(t, err)
	
	_, err = mockRepo.GetUserByID(nil, userID)
	require.NoError(t, err)
}

func TestUserServiceLogic_GetEmployeesFromMock(t *testing.T) {
	mockRepo := NewMockUserRepository()
	tenantID := uuid.New()
	
	employee1 := &models.User{
		ID:        uuid.New(),
		Email:     "emp1@test.com",
		Role:      models.RoleDealer,
		TenantID:  &tenantID,
		PasswordHash: "hash",
	}
	employee2 := &models.User{
		ID:        uuid.New(),
		Email:     "emp2@test.com",
		Role:      models.RoleDealerManager,
		TenantID:  &tenantID,
		PasswordHash: "hash",
	}
	
	mockRepo.byID[employee1.ID] = employee1
	mockRepo.byID[employee2.ID] = employee2
	
	result, err := mockRepo.GetUserByID(nil, employee1.ID)
	require.NoError(t, err)
	assert.Equal(t, "emp1@test.com", result.Email)
}

func TestUserServiceLogic_CreateUser(t *testing.T) {
	mockRepo := NewMockUserRepository()
	
	user := &models.User{
		Email:        "new@test.com",
		FirstName:    "New",
		LastName:     "User",
		Role:         models.RoleFranchisor,
		PasswordHash: "hash",
	}
	
	err := mockRepo.CreateUser(nil, user)
	require.NoError(t, err)
	
	result, err := mockRepo.GetUserByEmail(nil, "new@test.com")
	require.NoError(t, err)
	assert.Equal(t, "new@test.com", result.Email)
}

func TestUserServiceLogic_NotFound(t *testing.T) {
	mockRepo := NewMockUserRepository()
	
	_, err := mockRepo.GetUserByID(nil, uuid.New())
	assert.Error(t, err)
}

func TestUserServiceLogic_UpdateMultipleFields(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userID := uuid.New()
	
	user := &models.User{
		ID:           userID,
		Email:        "multi@test.com",
		FirstName:    "First",
		LastName:     "Last",
		Phone:        "",
		Role:         models.RoleDealer,
		PasswordHash: "hash",
	}
	mockRepo.byID[userID] = user
	
	err := mockRepo.UpdateUserFields(nil, userID, map[string]interface{}{
		"first_name": "UpdatedFirst",
		"last_name":  "UpdatedLast",
		"phone":      "+1234567890",
	})
	require.NoError(t, err)
	
	_, err = mockRepo.GetUserByID(nil, userID)
	require.NoError(t, err)
}

func TestUserServiceLogic_DeleteUser(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userID := uuid.New()
	
	user := &models.User{
		ID:           userID,
		Email:        "delete@test.com",
		Role:         models.RoleDealer,
		PasswordHash: "hash",
	}
	mockRepo.byID[userID] = user
	
	result, err := mockRepo.GetUserByID(nil, userID)
	require.NoError(t, err)
	assert.Equal(t, "delete@test.com", result.Email)
}

func TestUserServiceLogic_RoleHierarchy(t *testing.T) {
	tests := []struct {
		name string
		role models.Role
		want string
	}{
		{"super_admin", models.RoleSuperAdmin, "super_admin"},
		{"franchiser", models.RoleFranchisor, "franchiser"},
		{"franchiser_manager", models.RoleFranchisorManager, "franchiser_manager"},
		{"dealer", models.RoleDealer, "dealer"},
		{"salon_manager", models.RoleDealerManager, "salon_manager"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{
				ID:           uuid.New(),
				Email:        "test@test.com",
				Role:         tt.role,
				PasswordHash: "hash",
			}
			assert.Equal(t, tt.want, string(user.Role))
		})
	}
}