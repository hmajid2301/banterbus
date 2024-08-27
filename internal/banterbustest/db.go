package banterbustest

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/pressly/goose/v3"

	// used to connect to sqlite
	_ "modernc.org/sqlite"
)

func CreateDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return db, err
	}

	err = runDBMigrations(db)
	return db, err
}

func runDBMigrations(db *sql.DB) error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get current filename")
	}

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
