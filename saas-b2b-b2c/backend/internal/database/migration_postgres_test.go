package database

import (
	"os"
	"path/filepath"
	"testing"

	"franchise-saas-backend/internal/repository/testdb"

	"github.com/stretchr/testify/require"
)

func TestRunMigrationsFreshPostgres(t *testing.T) {
	if os.Getenv("TEST_DB_DSN") == "" {
		t.Skip("TEST_DB_DSN is not set")
	}
	db := testdb.MustConnectIsolated(t)

	require.NoError(t, runMigrations(db))
	require.NoError(t, runMigrations(db))

	for _, table := range []string{"users", "plans", "tenants", "auth_sessions"} {
		var exists bool
		require.NoError(t, db.Raw("SELECT to_regclass(?) IS NOT NULL", table).Scan(&exists).Error)
		require.True(t, exists, "missing table %s", table)
	}

	upSQL, err := os.ReadFile(filepath.Join("..", "..", "migrations", "021_auth_sessions.up.sql"))
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(upSQL)).Error)
	require.NoError(t, db.Exec(string(upSQL)).Error)

	downSQL, err := os.ReadFile(filepath.Join("..", "..", "migrations", "021_auth_sessions.down.sql"))
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(downSQL)).Error)

	var exists bool
	require.NoError(t, db.Raw("SELECT to_regclass('auth_sessions') IS NOT NULL").Scan(&exists).Error)
	require.False(t, exists)
	require.NoError(t, db.Exec(string(upSQL)).Error)
	require.NoError(t, runMigrations(db))
}
