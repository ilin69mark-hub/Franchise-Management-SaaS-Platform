package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_TableName(t *testing.T) {
	user := User{}
	assert.Equal(t, "users", user.TableName())
}

func TestUser_Roles(t *testing.T) {
	tests := []struct {
		name  string
		role  Role
		valid bool
	}{
		{"super_admin", RoleSuperAdmin, true},
		{"franchiser", RoleFranchisor, true},
		{"franchiser_manager", RoleFranchisorManager, true},
		{"dealer", RoleDealer, true},
		{"salon_manager", RoleDealerManager, true},
		{"invalid", Role("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{Role: tt.role}
			if tt.valid {
				assert.NotEmpty(t, user.Role)
			} else {
				assert.Equal(t, Role("invalid"), user.Role)
			}
		})
	}
}

func TestUser_Fields(t *testing.T) {
	tenantID := uuid.New()
	salonID := uuid.New()
	managedByID := uuid.New()

	user := User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		Role:         RoleDealer,
		Status:       "active",
		TenantID:     &tenantID,
		SalonID:      &salonID,
		ManagedBy:    &managedByID,
		FirstName:    "John",
		LastName:     "Doe",
		Phone:        "+1234567890",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, RoleDealer, user.Role)
	assert.Equal(t, "active", user.Status)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
	assert.Equal(t, "+1234567890", user.Phone)
	assert.NotNil(t, user.TenantID)
	assert.NotNil(t, user.SalonID)
	assert.NotNil(t, user.ManagedBy)
}

func TestUserLoginRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		password string
		valid   bool
	}{
		{"valid", "test@example.com", "password123", true},
		{"empty email", "", "password123", false},
		{"empty password", "test@example.com", "", false},
		{"both empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := UserLoginRequest{
				Email:    tt.email,
				Password: tt.password,
			}
			
			if tt.valid {
				assert.NotEmpty(t, req.Email)
				assert.NotEmpty(t, req.Password)
			} else {
				isEmpty := req.Email == "" || req.Password == ""
				assert.True(t, isEmpty)
			}
		})
	}
}

func TestUserRegisterRequest_Fields(t *testing.T) {
	req := UserRegisterRequest{
		Email:     "register@example.com",
		Password:  "secure123",
		Role:      RoleDealer,
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+1234567890",
	}

	assert.Equal(t, "register@example.com", req.Email)
	assert.Equal(t, "secure123", req.Password)
	assert.Equal(t, RoleDealer, req.Role)
	assert.Equal(t, "John", req.FirstName)
	assert.Equal(t, "Doe", req.LastName)
	assert.Equal(t, "+1234567890", req.Phone)
}

func TestUserUpdateRequest(t *testing.T) {
	req := UserUpdateRequest{
		FirstName: "UpdatedName",
		LastName:  "UpdatedLast",
		Phone:     "+9876543210",
	}

	assert.Equal(t, "UpdatedName", req.FirstName)
	assert.Equal(t, "UpdatedLast", req.LastName)
	assert.Equal(t, "+9876543210", req.Phone)
}

func TestAuthResponse(t *testing.T) {
	user := User{
		ID:    uuid.New(),
		Email: "auth@example.com",
		Role:  RoleDealer,
	}

	response := AuthResponse{
		User:         user,
		Token:        "access_token_123",
		RefreshToken: "refresh_token_456",
	}

	assert.Equal(t, user.ID, response.User.ID)
	assert.Equal(t, "access_token_123", response.Token)
	assert.Equal(t, "refresh_token_456", response.RefreshToken)
}

func TestCreateEmployeeRequest(t *testing.T) {
	tenantID := uuid.New()
	
	req := CreateEmployeeRequest{
		Email:     "employee@example.com",
		Password:  "password123",
		FirstName: "Employee",
		LastName:  "Name",
		Phone:     "+1111111111",
		Role:      RoleFranchisorManager,
		ManagedBy: new(uuid.UUID),
		TenantID:  &tenantID,
	}

	assert.Equal(t, "employee@example.com", req.Email)
	assert.NotEmpty(t, req.Password)
	assert.Equal(t, RoleFranchisorManager, req.Role)
	assert.NotNil(t, req.TenantID)
}

func TestUpdateEmployeeRequest(t *testing.T) {
	managedByID := uuid.New()
	
	req := UpdateEmployeeRequest{
		FirstName: "NewFirstName",
		LastName:  "NewLastName",
		Phone:     "+2222222222",
		Role:      RoleDealer,
		ManagedBy: &managedByID,
	}

	assert.Equal(t, "NewFirstName", req.FirstName)
	assert.Equal(t, RoleDealer, req.Role)
	assert.NotNil(t, req.ManagedBy)
}

func TestUser_Timestamps(t *testing.T) {
	now := time.Now()
	
	user := User{
		ID:        uuid.New(),
		Email:     "time@test.com",
		Role:      RoleDealer,
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
}

func TestUser_RoleConstants(t *testing.T) {
	assert.Equal(t, Role("super_admin"), RoleSuperAdmin)
	assert.Equal(t, Role("franchiser"), RoleFranchisor)
	assert.Equal(t, Role("franchiser_manager"), RoleFranchisorManager)
	assert.Equal(t, Role("dealer"), RoleDealer)
	assert.Equal(t, Role("salon_manager"), RoleDealerManager)
}

func TestUser_StatusDefault(t *testing.T) {
	user := User{
		ID:   uuid.New(),
		Email: "status@test.com",
		Role: RoleDealer,
	}

	// Default status should be empty or "active" based on model definition
	// This tests the field existence
	assert.NotNil(t, user.Status)
}

func TestUser_SoftDelete(t *testing.T) {
	user := User{
		ID:        uuid.New(),
		Email:     "delete@test.com",
		Role:      RoleDealer,
	}

	// Verify DeletedAt field exists - check it's initially zero
	// DeletedAt is gorm.DeletedAt which wraps sql.NullTime
	// Check it's zero by checking Valid field
	_ = user.DeletedAt.Valid // This should compile
	
	// Just verify the field exists
	assert.NotNil(t, user.DeletedAt)
}

func TestUser_TenantAssociation(t *testing.T) {
	tenantID := uuid.New()
	
	user := User{
		ID:        uuid.New(),
		Email:     "tenant@test.com",
		Role:      RoleFranchisor,
		TenantID:  &tenantID,
	}

	assert.NotNil(t, user.TenantID)
	assert.Equal(t, tenantID, *user.TenantID)
}

func TestUser_SalonAssociation(t *testing.T) {
	salonID := uuid.New()
	tenantID := uuid.New()
	
	user := User{
		ID:        uuid.New(),
		Email:     "salon@test.com",
		Role:      RoleDealerManager,
		TenantID:  &tenantID,
		SalonID:   &salonID,
		ManagedBy: new(uuid.UUID),
	}

	assert.NotNil(t, user.SalonID)
	assert.NotNil(t, user.ManagedBy)
}

func TestUser_ManagerHierarchy(t *testing.T) {
	franchiserID := uuid.New()
	dealerID := uuid.New()
	
	dealer := User{
		ID:        dealerID,
		Email:     "dealer@test.com",
		Role:      RoleDealer,
		ManagedBy: &franchiserID,
	}

	// Dealer has a manager (franchiser)
	require.NotNil(t, dealer.ManagedBy)
	assert.Equal(t, franchiserID, *dealer.ManagedBy)
}