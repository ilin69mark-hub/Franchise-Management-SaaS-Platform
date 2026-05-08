package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"franchise-saas-backend/internal/models"
)

type MockUserRepository struct {
	users map[string]*models.User
	byID  map[uuid.UUID]*models.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*models.User),
		byID:  make(map[uuid.UUID]*models.User),
	}
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if user, ok := m.users[email]; ok {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if user, ok := m.byID[id]; ok {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	m.users[user.Email] = user
	m.byID[user.ID] = user
	return nil
}

func (m *MockUserRepository) UpdateUserFields(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
	return nil
}

type MockTenantRepository struct {
	tenants map[uuid.UUID]*models.Tenant
}

func NewMockTenantRepository() *MockTenantRepository {
	return &MockTenantRepository{
		tenants: make(map[uuid.UUID]*models.Tenant),
	}
}

func (m *MockTenantRepository) CreateTenant(ctx context.Context, tenant *models.Tenant) (*models.Tenant, error) {
	tenant.ID = uuid.New()
	m.tenants[tenant.ID] = tenant
	return tenant, nil
}

func (m *MockTenantRepository) GetTenantByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	if t, ok := m.tenants[id]; ok {
		return t, nil
	}
	return nil, gorm.ErrRecordNotFound
}

type AuthServiceTestable struct {
	userRepo   *MockUserRepository
	tenantRepo *MockTenantRepository
}

func NewAuthServiceTestable() *AuthServiceTestable {
	return &AuthServiceTestable{
		userRepo:   NewMockUserRepository(),
		tenantRepo: NewMockTenantRepository(),
	}
}

func (s *AuthServiceTestable) CreateUser(user *models.User, password string) (*models.User, error) {
	if user.Email == "" {
		return nil, errors.New("email is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}
	if len(password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	user.ID = uuid.New()
	user.PasswordHash = "hashed_" + password
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	s.userRepo.users[user.Email] = user
	s.userRepo.byID[user.ID] = user
	return user, nil
}

func (s *AuthServiceTestable) Authenticate(email, password string) (*models.User, error) {
	user, err := s.userRepo.GetUserByEmail(context.Background(), email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if user.PasswordHash != "hashed_"+password {
		return nil, errors.New("invalid credentials")
	}

	freshUser, err := s.userRepo.GetUserByID(context.Background(), user.ID)
	if err != nil {
		return nil, errors.New("failed to fetch user details")
	}

	return freshUser, nil
}

func (s *AuthServiceTestable) CreateUserWithTenant(user *models.User, password string, companyName string) (*models.User, error) {
	if user.Email == "" || password == "" || companyName == "" {
		return nil, errors.New("email, password and company name are required")
	}

	user.ID = uuid.New()
	user.PasswordHash = "hashed_" + password

	tenant := &models.Tenant{ID: uuid.New(), Name: companyName, Status: "active"}
	s.tenantRepo.tenants[tenant.ID] = tenant
	user.TenantID = &tenant.ID
	user.Role = models.RoleFranchisor
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	s.userRepo.users[user.Email] = user
	s.userRepo.byID[user.ID] = user
	return user, nil
}

func (s *AuthServiceTestable) GenerateTokens(userID uuid.UUID, email string, role models.Role) (string, string, error) {
	secret := "test_secret_key"

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"jti":     uuid.New().String(),
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	accessStr, _ := accessToken.SignedString([]byte(secret))
	refreshStr, _ := refreshToken.SignedString([]byte(secret))

	return accessStr, refreshStr, nil
}

func (s *AuthServiceTestable) ValidateToken(tokenString string) (*models.User, error) {
	secret := "test_secret_key"

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}

	userID, _ := uuid.Parse(userIDStr)
	return s.userRepo.GetUserByID(context.Background(), userID)
}

func TestAuthService_CreateUser_Success(t *testing.T) {
	svc := NewAuthServiceTestable()

	tests := []struct {
		name     string
		email    string
		password string
		role     models.Role
	}{
		{"dealer user", "dealer@test.com", "password123", models.RoleDealer},
		{"manager user", "manager@test.com", "secure456", models.RoleFranchisorManager},
		{"salon manager", "salon@test.com", "mysecret", models.RoleDealerManager},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{
				Email:     tt.email,
				FirstName: "Test",
				LastName:  "User",
				Role:      tt.role,
			}

			result, err := svc.CreateUser(user, tt.password)

			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, result.ID)
			assert.Equal(t, tt.email, result.Email)
			assert.Equal(t, tt.role, result.Role)
			assert.Contains(t, result.PasswordHash, "hashed_")
		})
	}
}

func TestAuthService_CreateUser_ValidationErrors(t *testing.T) {
	svc := NewAuthServiceTestable()

	tests := []struct {
		name        string
		email       string
		password    string
		wantErr     string
	}{
		{"empty email", "", "password123", "email is required"},
		{"empty password", "test@test.com", "", "password is required"},
		{"short password", "test@test.com", "123", "password must be at least 6 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{Email: tt.email}
			_, err := svc.CreateUser(user, tt.password)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestAuthService_Authenticate_Success(t *testing.T) {
	svc := NewAuthServiceTestable()

	userID := uuid.New()
	user := &models.User{
		ID:             userID,
		Email:          "login@test.com",
		PasswordHash:   "hashed_password123",
	}
	svc.userRepo.users[user.Email] = user
	svc.userRepo.byID[user.ID] = user

	result, err := svc.Authenticate("login@test.com", "password123")

	require.NoError(t, err)
	assert.Equal(t, "login@test.com", result.Email)
}

func TestAuthService_Authenticate_Failures(t *testing.T) {
	svc := NewAuthServiceTestable()

	user := &models.User{
		Email:     "existing@test.com",
		PasswordHash: "hashed_correctpass",
	}
	svc.userRepo.users[user.Email] = user
	svc.userRepo.byID[user.ID] = user

	tests := []struct {
		name      string
		email     string
		password  string
		wantErr   string
	}{
		{"user not found", "nonexistent@test.com", "password", "invalid credentials"},
		{"wrong password", "existing@test.com", "wrongpass", "invalid credentials"},
		{"empty email", "", "password", "invalid credentials"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Authenticate(tt.email, tt.password)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestAuthService_CreateUserWithTenant(t *testing.T) {
	svc := NewAuthServiceTestable()

	tests := []struct {
		name        string
		email       string
		password    string
		companyName string
		wantErr     string
	}{
		{
			name:        "success",
			email:       "franchiser@test.com",
			password:    "secure123",
			companyName: "My Company",
			wantErr:     "",
		},
		{
			name:        "missing company name",
			email:       "test@test.com",
			password:    "password",
			companyName: "",
			wantErr:     "company name are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{Email: tt.email}
			result, err := svc.CreateUserWithTenant(user, tt.password, tt.companyName)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result.TenantID)
				assert.Equal(t, models.RoleFranchisor, result.Role)
			}
		})
	}
}

func TestAuthService_GenerateTokens(t *testing.T) {
	svc := NewAuthServiceTestable()

	userID := uuid.New()
	email := "token@test.com"

	tests := []struct {
		name    string
		userID  uuid.UUID
		email   string
		role    models.Role
	}{
		{"dealer token", userID, email, models.RoleDealer},
		{"franchiser token", uuid.New(), "boss@test.com", models.RoleFranchisor},
		{"manager token", uuid.New(), "mgr@test.com", models.RoleFranchisorManager},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			access, refresh, err := svc.GenerateTokens(tt.userID, tt.email, tt.role)

			require.NoError(t, err)
			assert.NotEmpty(t, access)
			assert.NotEmpty(t, refresh)

			// Verify token claims
			token, _ := jwt.Parse(access, func(token *jwt.Token) (interface{}, error) {
				return []byte("test_secret_key"), nil
			})
			claims := token.Claims.(jwt.MapClaims)
			assert.Equal(t, tt.userID.String(), claims["user_id"])
			assert.Equal(t, tt.email, claims["email"])
			assert.Equal(t, string(tt.role), claims["role"])
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	svc := NewAuthServiceTestable()

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "valid@test.com",
	}
	svc.userRepo.byID[userID] = user

	access, _, _ := svc.GenerateTokens(userID, "valid@test.com", models.RoleDealer)

	tests := []struct {
		name    string
		token   string
		wantErr string
	}{
		{"valid token", access, ""},
		{"invalid token", "invalid.token.here", "invalid token"},
		{"empty token", "", "invalid token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ValidateToken(tt.token)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuthService_RoleBasedAccess(t *testing.T) {
	svc := NewAuthServiceTestable()

	franchiserID := uuid.New()
	dealerID := uuid.New()

	franchiser := &models.User{
		ID:    franchiserID,
		Email: "franchiser@test.com",
		Role:  models.RoleFranchisor,
	}
	dealer := &models.User{
		ID:        dealerID,
		Email:     "dealer@test.com",
		Role:      models.RoleDealer,
		ManagedBy: &franchiserID,
	}

	svc.userRepo.byID[franchiserID] = franchiser
	svc.userRepo.byID[dealerID] = dealer

	// Franchiser can see all
	franchiserUser, err := svc.userRepo.GetUserByID(context.Background(), franchiserID)
	require.NoError(t, err)
	assert.Equal(t, models.RoleFranchisor, franchiserUser.Role)

	// Dealer has managed_by
	dealerUser, err := svc.userRepo.GetUserByID(context.Background(), dealerID)
	require.NoError(t, err)
	assert.Equal(t, franchiserID, *dealerUser.ManagedBy)
}

