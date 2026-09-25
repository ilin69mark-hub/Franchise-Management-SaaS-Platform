package services

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"franchise-saas-backend/internal/cache"
	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository/testdb"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func postgresTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Getenv("TEST_DB_DSN") == "" {
		t.Skip("TEST_DB_DSN is not set")
	}
	db := testdb.MustConnect(t)
	require.NoError(t, testdb.Cleanup(db))
	return db
}

func TestRefreshRotationConcurrentPostgres(t *testing.T) {
	viper.Set("jwt_secret", "postgres-refresh-secret")
	defer viper.Set("jwt_secret", "")
	db := postgresTestDB(t)
	oldRedis := cache.Client
	cache.Client = nil
	defer func() { cache.Client = oldRedis }()
	tenantID := uuid.New()
	require.NoError(t, db.Create(&models.Tenant{ID: tenantID, Name: "refresh", Status: "active", Timezone: "UTC", MaxUsers: 10, GracePeriodDays: 7}).Error)
	user := &models.User{ID: uuid.New(), Email: "refresh-" + uuid.NewString() + "@test.com", PasswordHash: "hash", Role: models.RoleFranchisor, Status: "active", AuthVersion: 1, TenantID: &tenantID}
	require.NoError(t, db.Create(user).Error)
	svc := NewAuthService(db)
	_, refresh, err := svc.IssueSession(context.Background(), user)
	require.NoError(t, err)
	var session models.AuthSession
	require.NoError(t, db.Where("user_id = ?", user.ID).First(&session).Error)

	const workers = 8
	start := make(chan struct{})
	results := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, _, refreshErr := svc.RefreshTokensWithContext(context.Background(), refresh)
			results <- refreshErr
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	successes := 0
	failures := make([]string, 0, workers)
	for refreshErr := range results {
		if refreshErr == nil {
			successes++
		} else {
			failures = append(failures, refreshErr.Error())
		}
	}
	t.Logf("refresh failures: %v", failures)
	require.Equal(t, 1, successes)
	require.NoError(t, db.First(&session, "id = ?", session.ID).Error)
	require.NotNil(t, session.RevokedAt)
	_, err = svc.ResolveAccessIdentity(context.Background(), user.ID, session.ID, 1)
	require.Error(t, err)
}

func TestCreateEmployeeQuotaConcurrentPostgres(t *testing.T) {
	db := postgresTestDB(t)
	tenantID := uuid.New()
	require.NoError(t, db.Create(&models.Tenant{ID: tenantID, Name: "quota", Status: "active", Timezone: "UTC", MaxUsers: 2, GracePeriodDays: 7}).Error)
	owner := &models.User{ID: uuid.New(), Email: "quota-owner-" + uuid.NewString() + "@test.com", PasswordHash: "hash", Role: models.RoleFranchisor, Status: "active", AuthVersion: 1, TenantID: &tenantID}
	require.NoError(t, db.Create(owner).Error)
	svc := NewUserService(db)
	const workers = 12
	var successes atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, err := svc.CreateEmployee(models.CreateEmployeeRequest{
				Email:     "quota-" + uuid.NewString() + "@test.com",
				Password:  "Forensic-Quota-123!",
				FirstName: "Quota",
				LastName:  uuid.NewString(),
				Role:      models.RoleDealer,
			}, tenantID, owner.ID, string(models.RoleFranchisor))
			if err == nil {
				successes.Add(1)
			}
		}(i)
	}
	wg.Wait()
	require.Equal(t, int64(1), successes.Load())
	var count int64
	require.NoError(t, db.Model(&models.User{}).Where("tenant_id = ?", tenantID).Count(&count).Error)
	require.Equal(t, int64(2), count)
}
