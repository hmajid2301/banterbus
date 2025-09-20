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
		t.Fatal("Could not find project root directory (no go.mod found)")
	}

	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("Failed to change to project root %s: %v", projectRoot, err)
	}
	t.Cleanup(func() {
		os.Chdir(originalWd)
	})

	// Use relative path from project root (now that we've changed directory)
	migrationsPath := "internal/store/db/sqlc/migrations"

	// Verify migrations directory exists (using relative path)
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		// Also check absolute path for debugging
		absPath := filepath.Join(projectRoot, migrationsPath)
		if _, err2 := os.Stat(absPath); os.IsNotExist(err2) {
			t.Fatalf("Migrations directory does not exist at relative path '%s' or absolute path '%s'", migrationsPath, absPath)
		}
		t.Fatalf("Migrations directory does not exist at relative path '%s' (current dir: %s)", migrationsPath, projectRoot)
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

	migrator := goosemigrator.New(migrationsPath)
	sqlDB := pgtestdb.New(t, cfg, migrator)

	pgxConfig, err := pgxpool.ParseConfig(getConnectionURL(sqlDB))
	if err != nil {
		t.Fatalf("failed to parse database URL: %v", err)
	}

	pgxConfig.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	loadSeedData(t, pool)

	return pool
}

func findProjectRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	dir := wd
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
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

	ctx := context.Background()
	for _, table := range tables {
		if _, err := pool.Exec(ctx, "DELETE FROM "+table); err != nil {
			t.Logf("Warning: failed to cleanup table %s: %v", table, err)
		}
	}
}

func getConnectionURL(db *sql.DB) string {
	var dbName string
	err := db.QueryRow("SELECT current_database()").Scan(&dbName)
	if err != nil {
		panic("failed to get database name: " + err.Error())
	}

	host := getEnv("BANTERBUS_DB_HOST", "localhost")
	port := getEnv("BANTERBUS_DB_PORT", "5432")
	user := getEnv("BANTERBUS_DB_USER", "postgres")
	password := getEnv("BANTERBUS_DB_PASSWORD", "postgres")

	return "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbName + "?sslmode=disable"
}

func loadSeedData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	seedDataPath := findSeedDataFile(t)
	if seedDataPath == "" {
		return
	}

	seedData, err := os.ReadFile(seedDataPath)
	if err != nil {
		return
	}

	content := string(seedData)
	lines := strings.Split(content, "\n")
	var statements []string
	var currentStatement strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		currentStatement.WriteString(line)
		currentStatement.WriteString("\n")

		if strings.HasSuffix(line, ";") {
			statements = append(statements, currentStatement.String())
			currentStatement.Reset()
		}
	}

	ctx := context.Background()
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		pool.Exec(ctx, stmt)
	}
}

func findSeedDataFile(t *testing.T) string {
	t.Helper()

	candidates := []string{
		"docker/postgres-init/01-seed-data.sql",
		"../docker/postgres-init/01-seed-data.sql",
		"../../docker/postgres-init/01-seed-data.sql",
	}

	for _, path := range candidates {
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
