package banterbustest

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/goosemigrator"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

func NewDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	originalWd, _ := os.Getwd()
	projectRoot := findProjectRoot(t)
	if projectRoot == "" {
		t.Fatal("Could not find project root")
	}

	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}
	t.Cleanup(func() { os.Chdir(originalWd) })

	migrationsPath := "internal/store/db/sqlc/migrations"
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		t.Fatalf("Migrations directory not found: %s", migrationsPath)
	}

	cfg := pgtestdb.Config{
		DriverName: "pgx",
		User:       getEnv("BANTERBUS_DB_USER", "postgres"),
		Password:   getEnv("BANTERBUS_DB_PASSWORD", "postgres"),
		Host:       getEnv("BANTERBUS_DB_HOST", "localhost"),
		Port:       getEnv("BANTERBUS_DB_PORT", "5432"),
		Database:   "postgres",
		Options:    "sslmode=disable",
	}

	sqlDB := pgtestdb.New(t, cfg, goosemigrator.New(migrationsPath))

	pgxConfig, err := pgxpool.ParseConfig(buildConnectionURL(sqlDB, cfg))
	if err != nil {
		t.Fatalf("failed to parse database URL: %v", err)
	}

	pgxConfig.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(t.Context(), pgxConfig)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}

	t.Cleanup(pool.Close)
	loadSeedData(t, pool)

	return pool
}

func findProjectRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	for dir := wd; ; {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func CleanupData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	tables := []string{
		"fibbing_it_scores",
		"fibbing_it_votes",
		"fibbing_it_answers",
		"fibbing_it_player_roles",
		"fibbing_it_rounds",
		"game_state",
		"rooms_players",
		"rooms",
		"players",
	}

	ctx := t.Context()
	for _, table := range tables {
		if _, err := pool.Exec(ctx, "DELETE FROM "+table); err != nil {
			t.Logf("Warning: failed to cleanup table %s: %v", table, err)
		}
	}
}

func buildConnectionURL(db *sql.DB, cfg pgtestdb.Config) string {
	var dbName string
	if err := db.QueryRow("SELECT current_database()").Scan(&dbName); err != nil {
		panic("failed to get database name: " + err.Error())
	}

	return "postgres://" + cfg.User + ":" + cfg.Password + "@" + cfg.Host + ":" + cfg.Port + "/" + dbName + "?sslmode=disable"
}

func loadSeedData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	seedPath := findSeedDataFile(t)
	if seedPath == "" {
		return
	}

	data, err := os.ReadFile(seedPath)
	if err != nil {
		return
	}

	statements := parseSQLStatements(string(data))
	ctx := t.Context()
	for _, stmt := range statements {
		if stmt = strings.TrimSpace(stmt); stmt != "" {
			pool.Exec(ctx, stmt)
		}
	}
}

func parseSQLStatements(content string) []string {
	var statements []string
	var current strings.Builder

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		if strings.HasSuffix(line, ";") {
			statements = append(statements, current.String())
			current.Reset()
		}
	}

	return statements
}

func findSeedDataFile(t *testing.T) string {
	t.Helper()

	for _, path := range []string{
		"docker/postgres-init/01-seed-data.sql",
		"../docker/postgres-init/01-seed-data.sql",
		"../../docker/postgres-init/01-seed-data.sql",
	} {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
