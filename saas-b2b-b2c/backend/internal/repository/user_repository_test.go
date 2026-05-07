package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"franchise-saas-backend/internal/models"
)

// MockDB - мок для GORM
type MockDB struct {
	users map[uuid.UUID]*models.User
}

func NewMockDB() *MockDB {
	return &MockDB{
		users: make(map[uuid.UUID]*models.User),
	}
}

func (m *MockDB) Create(user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockDB) First(user *models.User, id uuid.UUID) error {
	if u, ok := m.users[id]; ok {
		*user = *u
		return nil
	}
	return gorm.ErrRecordNotFound
}

func (m *MockDB) Where(email string) *MockQuery {
	return &MockQuery{db: m, email: email}
}

func (m *MockDB) Find(users *[]models.User) error {
	for _, u := range m.users {
		*users = append(*users, *u)
	}
	return nil
}

func (m *MockDB) Model(model interface{}) *MockDB {
	return m
}

func (m *MockDB) Updates(updates map[string]interface{}) error {
	return nil
}

type MockQuery struct {
	db    *MockDB
	email string
}

func (q *MockQuery) First(user *models.User) error {
	for _, u := range q.db.users {
		if u.Email == q.email {
			*user = *u
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (q *MockQuery) Find(users *[]models.User) error {
	for _, u := range q.db.users {
		if u.Email == q.email || q.email == "" {
			*users = append(*users, *u)
		}
	}
	return nil
}

// UserRepositoryMock - мок репозиторий
type UserRepositoryMock struct {
	db    *MockDB
	users map[string]*models.User
}

func NewUserRepositoryMock() *UserRepositoryMock {
	mockDB := NewMockDB()
	return &UserRepositoryMock{
		db:    mockDB,
		users: make(map[string]*models.User),
	}
}

func (r *UserRepositoryMock) CreateUser(ctx context.Context, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	r.users[user.Email] = user
	return nil
}

func (r *UserRepositoryMock) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if user, ok := r.users[email]; ok {
		return user, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *UserRepositoryMock) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *UserRepositoryMock) GetUsersByTenantAndRole(ctx context.Context, tenantID uuid.UUID, role models.Role) ([]models.User, error) {
	var result []models.User
	for _, user := range r.users {
		if user.TenantID != nil && *user.TenantID == tenantID && user.Role == role {
			result = append(result, *user)
		}
	}
	return result, nil
}

func (r *UserRepositoryMock) UpdateUserFields(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
	for _, user := range r.users {
		if user.ID == userID {
			if fn, ok := updates["first_name"]; ok {
				user.FirstName = fn.(string)
			}
			if ln, ok := updates["last_name"]; ok {
				user.LastName = ln.(string)
			}
			if r, ok := updates["role"]; ok {
				user.Role = r.(models.Role)
			}
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (r *UserRepositoryMock) FindAllGlobal(ctx context.Context) ([]models.User, error) {
	var result []models.User
	for _, user := range r.users {
		result = append(result, *user)
	}
	return result, nil
}

// Tests using mock
func TestUserRepositoryMock_CreateUser(t *testing.T) {
	repo := NewUserRepositoryMock()
	ctx := context.Background()

	user := &models.User{
		Email:     "test@example.com",
		FirstName: "John",
		Role:      models.RoleDealer,
	}

	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.NotEmpty(t, repo.users)
}

func TestUserRepositoryMock_GetUserByEmail(t *testing.T) {
	repo := NewUserRepositoryMock()
	ctx := context.Background()

	// Create user first
	user := &models.User{
		ID:    uuid.New(),
		Email: "search@test.com",
		Role:  models.RoleDealer,
	}
	repo.CreateUser(ctx, user)

	tests := []struct {
		name      string
		email     string
		wantFound bool
	}{
		{"existing", "search@test.com", true},
		{"not found", "notfound@test.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetUserByEmail(ctx, tt.email)
			if tt.wantFound {
				require.NoError(t, err)
				assert.Equal(t, tt.email, user.Email)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestUserRepositoryMock_GetUserByID(t *testing.T) {
	repo := NewUserRepositoryMock()
	ctx := context.Background()

	testUserID := uuid.New()
	user := &models.User{
		ID:    testUserID,
		Email: "byid@test.com",
		Role:  models.RoleDealer,
	}
	repo.CreateUser(ctx, user)

	tests := []struct {
		name      string
		id        uuid.UUID
		wantFound bool
	}{
		{"exists", testUserID, true},
		{"not exists", uuid.New(), false},
		{"zero UUID", uuid.Nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetUserByID(ctx, tt.id)
			if tt.wantFound {
				require.NoError(t, err)
				assert.Equal(t, tt.id, user.ID)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestUserRepositoryMock_GetUsersByTenantAndRole(t *testing.T) {
	repo := NewUserRepositoryMock()
	ctx := context.Background()

	tenantID := uuid.New()
	otherTenantID := uuid.New()

	users := []*models.User{
		{ID: uuid.New(), Email: "dealer1@test.com", Role: models.RoleDealer, TenantID: &tenantID},
		{ID: uuid.New(), Email: "dealer2@test.com", Role: models.RoleDealer, TenantID: &tenantID},
		{ID: uuid.New(), Email: "manager@test.com", Role: models.RoleFranchisorManager, TenantID: &tenantID},
		{ID: uuid.New(), Email: "other@test.com", Role: models.RoleDealer, TenantID: &otherTenantID},
	}

	for _, u := range users {
		repo.CreateUser(ctx, u)
	}

	tests := []struct {
		name      string
		tenantID  uuid.UUID
		role      models.Role
		wantCount int
	}{
		{"dealers in tenant", tenantID, models.RoleDealer, 2},
		{"managers in tenant", tenantID, models.RoleFranchisorManager, 1},
		{"other tenant", otherTenantID, models.RoleDealer, 1},
		{"no results", tenantID, models.RoleSuperAdmin, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetUsersByTenantAndRole(ctx, tt.tenantID, tt.role)
			require.NoError(t, err)
			assert.Len(t, result, tt.wantCount)
		})
	}
}

func TestUserRepositoryMock_UpdateUserFields(t *testing.T) {
	repo := NewUserRepositoryMock()
	ctx := context.Background()

	testUserID := uuid.New()
	user := &models.User{
		ID:        testUserID,
		Email:     "update@test.com",
		FirstName: "Original",
		Role:      models.RoleDealer,
	}
	repo.CreateUser(ctx, user)

	tests := []struct {
		name    string
		updates map[string]interface{}
		wantErr bool
	}{
		{
			name:    "update first name",
			updates: map[string]interface{}{"first_name": "Updated"},
			wantErr: false,
		},
		{
			name:    "update multiple",
			updates: map[string]interface{}{"first_name": "New", "last_name": "Last"},
			wantErr: false,
		},
		{
			name:    "update role",
			updates: map[string]interface{}{"role": models.RoleFranchisorManager},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateUserFields(ctx, testUserID, tt.updates)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				user, _ := repo.GetUserByID(ctx, testUserID)
				if fn, ok := tt.updates["first_name"]; ok {
					assert.Equal(t, fn, user.FirstName)
				}
			}
		})
	}
}

func TestUserRepositoryMock_FindAllGlobal(t *testing.T) {
	repo := NewUserRepositoryMock()
	ctx := context.Background()

	tenant1 := uuid.New()

	users := []*models.User{
		{ID: uuid.New(), Email: "user1@test.com", Role: models.RoleDealer, TenantID: &tenant1},
		{ID: uuid.New(), Email: "user2@test.com", Role: models.RoleFranchisor},
		{ID: uuid.New(), Email: "user3@test.com", Role: models.RoleDealerManager},
	}

	for _, u := range users {
		repo.CreateUser(ctx, u)
	}

	result, err := repo.FindAllGlobal(ctx)
	require.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestUserRepositoryMock_BatchOperations(t *testing.T) {
	repo := NewUserRepositoryMock()
	ctx := context.Background()

	tenantID := uuid.New()

	users := []*models.User{
		{ID: uuid.New(), Email: "batch1@test.com", Role: models.RoleDealer, TenantID: &tenantID},
		{ID: uuid.New(), Email: "batch2@test.com", Role: models.RoleDealer, TenantID: &tenantID},
		{ID: uuid.New(), Email: "batch3@test.com", Role: models.RoleDealer, TenantID: &tenantID},
	}

	for _, u := range users {
		repo.CreateUser(ctx, u)
	}

	result, err := repo.GetUsersByTenantAndRole(ctx, tenantID, models.RoleDealer)
	require.NoError(t, err)
	assert.Len(t, result, 3)

	for _, u := range result {
		repo.UpdateUserFields(ctx, u.ID, map[string]interface{}{"role": models.RoleFranchisorManager})
	}

	managers, err := repo.GetUsersByTenantAndRole(ctx, tenantID, models.RoleFranchisorManager)
	require.NoError(t, err)
	assert.Len(t, managers, 3)
}

func TestUserRepositoryMock_RoleHierarchy(t *testing.T) {
	repo := NewUserRepositoryMock()
	ctx := context.Background()

	franchiserID := uuid.New()
	dealerID := uuid.New()
	salonManagerID := uuid.New()

	// Create hierarchy
	franchiser := &models.User{
		ID:       franchiserID,
		Email:    "franchiser@test.com",
		Role:     models.RoleFranchisor,
		TenantID: new(uuid.UUID),
	}
	dealer := &models.User{
		ID:        dealerID,
		Email:     "dealer@test.com",
		Role:      models.RoleDealer,
		TenantID:  franchiser.TenantID,
		ManagedBy: &franchiserID,
	}
	salonManager := &models.User{
		ID:        salonManagerID,
		Email:     "manager@test.com",
		Role:      models.RoleDealerManager,
		TenantID:  franchiser.TenantID,
		ManagedBy: &dealerID,
	}

	repo.CreateUser(ctx, franchiser)
	repo.CreateUser(ctx, dealer)
	repo.CreateUser(ctx, salonManager)

	// Verify hierarchy
	fUser, err := repo.GetUserByID(ctx, franchiserID)
	require.NoError(t, err)
	assert.Equal(t, models.RoleFranchisor, fUser.Role)

	dUser, err := repo.GetUserByID(ctx, dealerID)
	require.NoError(t, err)
	assert.Equal(t, &franchiserID, dUser.ManagedBy)

	sUser, err := repo.GetUserByID(ctx, salonManagerID)
	require.NoError(t, err)
	assert.Equal(t, &dealerID, sUser.ManagedBy)
}

func TestUserRepositoryMock_Validation(t *testing.T) {
	repo := NewUserRepositoryMock()
	ctx := context.Background()

	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
	}{
		{
			name: "valid dealer",
			user: &models.User{
				Email: "dealer@test.com",
				Role:  models.RoleDealer,
			},
			wantErr: false,
		},
		{
			name: "valid franchiser",
			user: &models.User{
				Email: "franchiser@test.com",
				Role:  models.RoleFranchisor,
			},
			wantErr: false,
		},
		{
			name: "valid salon manager",
			user: &models.User{
				Email: "salon@test.com",
				Role:  models.RoleDealerManager,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateUser(ctx, tt.user)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}