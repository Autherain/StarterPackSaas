package database

import (
	"fmt"
	"os"
	"time"

	"testing"

	"github.com/autherain/test/internal/assert"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()

	dsn := os.Getenv("TEST_DB_DSN")

	if dsn == "" {
		t.Fatal("TEST_DB_DSN environment variable must be set in the format user:pass@localhost:port/db")
	}

	schemaName := fmt.Sprintf("test_schema_%d", time.Now().UnixNano())
	dsn = fmt.Sprintf("%s?search_path=%s", dsn, schemaName)

	db, err := New(dsn)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		defer db.Close()

		_, err = db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
		if err != nil {
			t.Error(err)
		}
	})

	_, err = db.Exec(fmt.Sprintf("CREATE SCHEMA %s", schemaName))
	if err != nil {
		t.Fatal(err)
	}

	err = db.MigrateUp()
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func TestNew(t *testing.T) {
	t.Run("Creates DB connection pool", func(t *testing.T) {
		dsn := os.Getenv("TEST_DB_DSN")

		if dsn == "" {
			t.Fatal("TEST_DB_DSN environment variable must be set in the format user:pass@localhost:port/db")
		}

		db, err := New(dsn)
		assert.Nil(t, err)
		assert.NotNil(t, db)
		assert.NotNil(t, db.DB)
		defer db.Close()

		err = db.Ping()
		assert.Nil(t, err)

		assert.Equal(t, 25, db.Stats().MaxOpenConnections)
	})

	t.Run("Fails with invalid DSN", func(t *testing.T) {
		dsn := "fake_user:fake_pass@localhost:5432/fake_db"

		db, err := New(dsn)
		assert.NotNil(t, err)
		assert.Nil(t, db)
	})
}

func TestMigrateUp(t *testing.T) {
	t.Run("Applies all up migrations", func(t *testing.T) {
		db := newTestDB(t)

		err := db.MigrateUp()
		assert.Nil(t, err)

		var version int
		err = db.Get(&version, "SELECT version FROM schema_migrations LIMIT 1")
		assert.Nil(t, err)
		assert.True(t, version > 0)
	})
}
