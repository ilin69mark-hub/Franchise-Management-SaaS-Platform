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

func TestPlanRepository_Create_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := testdb.MustConnect(t)
	repo := NewPlanRepository(db)

	plan := &models.Plan{
		ID:        uuid.New(),
		Name:      "Basic Plan",
		Price:     29.99,
		MaxSalons: 1,
		MaxUsers:  5,
	}

	err := repo.Create(context.Background(), plan)
	require.NoError(t, err)

	var found models.Plan
	err = db.First(&found, "id = ?", plan.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "Basic Plan", found.Name)
	assert.Equal(t, 29.99, found.Price)
}

func TestPlanRepository_GetByID_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := testdb.MustConnect(t)
	repo := NewPlanRepository(db)

	plan := &models.Plan{
		ID:        uuid.New(),
		Name:      "Premium Plan",
		Price:     99.99,
		MaxSalons: 10,
		MaxUsers:  50,
	}

	err := repo.Create(context.Background(), plan)
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), plan.ID.String())
	require.NoError(t, err)
	assert.Equal(t, "Premium Plan", found.Name)
}

func TestPlanRepository_List_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := testdb.MustConnect(t)
	repo := NewPlanRepository(db)

	plans := []models.Plan{
		{ID: uuid.New(), Name: "Basic", Price: 10, MaxSalons: 1, MaxUsers: 5},
		{ID: uuid.New(), Name: "Pro", Price: 50, MaxSalons: 5, MaxUsers: 20},
		{ID: uuid.New(), Name: "Enterprise", Price: 200, MaxSalons: 100, MaxUsers: 500},
	}

	for _, p := range plans {
		err := repo.Create(context.Background(), &p)
		require.NoError(t, err)
	}

	result, count, err := repo.List(context.Background(), ListOptions{Limit: 10})
	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, int64(3), count)
}

func TestPlanRepository_Update_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := testdb.MustConnect(t)
	repo := NewPlanRepository(db)

	plan := &models.Plan{
		ID:        uuid.New(),
		Name:      "Original",
		Price:     10,
		MaxSalons: 1,
		MaxUsers:  5,
	}

	err := repo.Create(context.Background(), plan)
	require.NoError(t, err)

	plan.Name = "Updated"
	plan.Price = 25

	err = repo.Update(context.Background(), plan)
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), plan.ID.String())
	require.NoError(t, err)
	assert.Equal(t, "Updated", found.Name)
	assert.Equal(t, 25.0, found.Price)
}

func TestPlanRepository_Delete_Postgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := testdb.MustConnect(t)
	repo := NewPlanRepository(db)

	plan := &models.Plan{
		ID:    uuid.New(),
		Name:  "To Delete",
		Price: 10,
	}

	err := repo.Create(context.Background(), plan)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), plan.ID.String())
	require.NoError(t, err)

	_, err = repo.GetByID(context.Background(), plan.ID.String())
	assert.Error(t, err)
}