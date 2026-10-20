// Package testutil helpers for integration tests that need a real postgres.
package testutil

import (
	"os"
	"testing"

	"github.com/MauricioLopezC/ReservationSystem/internal/database"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OpenTestDB connects to TEST_DATABASE_DSN, recreates the public schema and runs
// the migrations. Tests are skipped when the variable isn't set.
// WARNING: it drops everything in that database, never point it to real data.
func OpenTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set, skipping integration test")
	}

	db, err := database.Open(dsn, logger.Silent)
	require.NoError(t, err)

	require.NoError(t, db.Exec("DROP SCHEMA public CASCADE").Error)
	require.NoError(t, db.Exec("CREATE SCHEMA public").Error)
	require.NoError(t, database.Migrate(db))

	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
	return db
}
