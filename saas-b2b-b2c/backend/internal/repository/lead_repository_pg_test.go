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

func TestLeadRepository_CreateLead_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := testdb.MustConnect(t)
	repo := NewLeadRepository(db)

	lead := &models.Lead{
		ID:        uuid.New(),
		SalonID:   uuid.New(),
		ManagerID: uuid.New(),
		FullName:  "John Doe",
		Phone:     "+1234567890",
		Email:     "john@example.com",
		Status:    "new",
	}

	err := repo.CreateLead(context.Background(), lead)
	require.NoError(t, err)

	var found models.Lead
	err = db.First(&found, "id = ?", lead.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "John Doe", found.FullName)
	assert.Equal(t, "new", found.Status)
}

func TestLeadRepository_GetLeadsByManager_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := testdb.MustConnect(t)
	repo := NewLeadRepository(db)

	managerID := uuid.New()
	salonID := uuid.New()

	leads := []models.Lead{
		{ID: uuid.New(), SalonID: salonID, ManagerID: managerID, FullName: "Lead 1", Status: "new"},
		{ID: uuid.New(), SalonID: salonID, ManagerID: managerID, FullName: "Lead 2", Status: "contact"},
		{ID: uuid.New(), SalonID: salonID, ManagerID: uuid.New(), FullName: "Other Lead", Status: "new"},
	}

	for _, lead := range leads {
		err := repo.CreateLead(context.Background(), &lead)
		require.NoError(t, err)
	}

	result, err := repo.GetLeadsByManager(context.Background(), managerID)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestLeadRepository_UpdateLeadStatus_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := testdb.MustConnect(t)
	repo := NewLeadRepository(db)

	lead := models.Lead{
		ID:        uuid.New(),
		SalonID:   uuid.New(),
		ManagerID: uuid.New(),
		FullName:  "Test Lead",
		Status:    "new",
	}

	err := repo.CreateLead(context.Background(), &lead)
	require.NoError(t, err)

	err = repo.UpdateLeadStatus(context.Background(), lead.ID, "contact")
	require.NoError(t, err)

	found, err := repo.GetLeadByID(context.Background(), lead.ID, lead.ManagerID)
	require.NoError(t, err)
	assert.Equal(t, "contact", found.Status)
}

func TestLeadRepository_AddActivity_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := testdb.MustConnect(t)
	repo := NewLeadRepository(db)

	lead := models.Lead{
		ID:        uuid.New(),
		SalonID:   uuid.New(),
		ManagerID: uuid.New(),
		FullName:  "Test Lead",
		Status:    "new",
	}
	err := repo.CreateLead(context.Background(), &lead)
	require.NoError(t, err)

	activity := &models.LeadActivity{
		ID:          uuid.New(),
		LeadID:      lead.ID,
		UserID:      lead.ManagerID,
		Type:        "call",
		Description: "Initial call",
	}

	err = repo.AddActivity(context.Background(), activity)
	require.NoError(t, err)

	activities, err := repo.GetLeadActivities(context.Background(), lead.ID)
	require.NoError(t, err)
	assert.Len(t, activities, 1)
	assert.Equal(t, "call", activities[0].Type)
}