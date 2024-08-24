package banterbustest

import (
	"context"
	"database/sql"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"

	// used to connect to sqlite
	_ "modernc.org/sqlite"
)

func CreateDB(ctx context.Context, t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	assert.NoError(t, err)

	err = runDBMigrations(t, db)
	assert.NoError(t, err)

	return db
}

func runDBMigrations(t *testing.T, db *sql.DB) error {
	_, filename, _, ok := runtime.Caller(0)
	assert.True(t, ok)
	dir := path.Join(path.Dir(filename), "..")
	migrationFolder := filepath.Join(dir, "../")

	fs := os.DirFS(migrationFolder)
	goose.SetBaseFS(fs)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	if err := goose.Up(db, "db/migrations"); err != nil {
		return err
	}
	return nil
}
