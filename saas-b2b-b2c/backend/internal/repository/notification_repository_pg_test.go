package repository

import (
	"context"
	"testing"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository/testdb"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationRepository_Create_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := testdb.MustConnect(t)
	repo := NewNotificationRepository(db)

	notification := &models.Notification{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Message:   "Test notification",
		IsRead:    false,
	}

	err := repo.Create(context.Background(), notification)
	require.NoError(t, err)

	var found models.Notification
	err = db.First(&found, "id = ?", notification.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "Test notification", found.Message)
	assert.False(t, found.IsRead)
}

func TestNotificationRepository_GetByTenant_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := testdb.MustConnect(t)
	repo := NewNotificationRepository(db)

	tenantID := uuid.New()

	notifications := []models.Notification{
		{ID: uuid.New(), TenantID: tenantID, Message: "Notif 1", IsRead: false},
		{ID: uuid.New(), TenantID: tenantID, Message: "Notif 2", IsRead: true},
		{ID: uuid.New(), TenantID: uuid.New(), Message: "Other tenant", IsRead: false},
	}

	for _, n := range notifications {
		err := repo.Create(context.Background(), &n)
		require.NoError(t, err)
	}

	result, err := repo.GetByTenant(context.Background(), tenantID, 10)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestNotificationRepository_MarkAsRead_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := testdb.MustConnect(t)
	repo := NewNotificationRepository(db)

	notification := models.Notification{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Message:   "Test",
		IsRead:    false,
	}

	err := repo.Create(context.Background(), &notification)
	require.NoError(t, err)

	err = repo.MarkAsRead(context.Background(), notification.ID)
	require.NoError(t, err)

	var found models.Notification
	err = db.First(&found, "id = ?", notification.ID).Error
	require.NoError(t, err)
	assert.True(t, found.IsRead)
}

func TestNotificationRepository_MarkAllAsRead_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := testdb.MustConnect(t)
	repo := NewNotificationRepository(db)

	tenantID := uuid.New()

	notifications := []models.Notification{
		{ID: uuid.New(), TenantID: tenantID, Message: "Notif 1", IsRead: false},
		{ID: uuid.New(), TenantID: tenantID, Message: "Notif 2", IsRead: false},
		{ID: uuid.New(), TenantID: uuid.New(), Message: "Other", IsRead: false},
	}

	for _, n := range notifications {
		err := repo.Create(context.Background(), &n)
		require.NoError(t, err)
	}

	err := repo.MarkAllAsRead(context.Background(), tenantID)
	require.NoError(t, err)

	result, err := repo.GetByTenant(context.Background(), tenantID, 10)
	require.NoError(t, err)
	for _, n := range result {
		assert.True(t, n.IsRead)
	}
}