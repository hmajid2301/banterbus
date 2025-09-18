package banterbustest

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"math/big"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/goosemigrator"
	"github.com/pressly/goose/v3"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

// getConfig returns the pgtestdb configuration based on environment
func getConfig() pgtestdb.Config {
	// Check for CI environment variables first
	host := os.Getenv("BANTERBUS_DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("BANTERBUS_DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("BANTERBUS_DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("BANTERBUS_DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	database := os.Getenv("BANTERBUS_DB_NAME")
	if database == "" {
		database = "banterbus" // Use banterbus database which has migrations applied
	}

	return pgtestdb.Config{
		DriverName: "pgx",
		User:       user,
		Password:   password,
		Host:       host,
		Port:       port,
		Database:   database,
		Options:    "sslmode=disable",
	}
}

// getMigrator returns a configured goose migrator
func getMigrator() *goosemigrator.GooseMigrator {
	// Check for explicit migrations path from environment first (for CI)
	migrationsDir := os.Getenv("BANTERBUS_MIGRATIONS_DIR")
	if migrationsDir == "" {
		// Find project root by looking for go.mod
		projectRoot := findProjectRoot()
		migrationsDir = path.Join(projectRoot, "internal", "store", "db", "sqlc", "migrations")
	}

	return goosemigrator.New(migrationsDir)
}

// findProjectRoot looks for go.mod to find the project root
func findProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic("failed to get working directory")
	}

	// Walk up the directory tree looking for go.mod
	for {
		if _, err := os.Stat(path.Join(wd, "go.mod")); err == nil {
			return wd
		}

		parent := path.Dir(wd)
		if parent == wd {
			// Reached root directory
			panic("could not find project root (go.mod not found)")
		}
		wd = parent
	}
}

// getSeedDataScript returns the path to the seed data SQL script
func getSeedDataScript() string {
	// Check for explicit seed data path from environment first (for CI)
	seedDataPath := os.Getenv("BANTERBUS_SEED_DATA_PATH")
	if seedDataPath == "" {
		// Find project root and construct path
		projectRoot := findProjectRoot()
		seedDataPath = path.Join(projectRoot, "docker", "postgres-init", "01-seed-data.sql")
	}

	return seedDataPath
}

// CustomMigrator wraps the goose migrator and adds seed data functionality
type CustomMigrator struct {
	*goosemigrator.GooseMigrator
	seedDataPath string
}

// Migrate runs the goose migrations and then applies seed data
func (cm *CustomMigrator) Migrate(ctx context.Context, db *sql.DB, config pgtestdb.Config) error {
	// First run the goose migrations
	if err := cm.GooseMigrator.Migrate(ctx, db, config); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Then apply seed data
	if cm.seedDataPath != "" {
		seedData, err := os.ReadFile(cm.seedDataPath)
		if err != nil {
			return fmt.Errorf("failed to read seed data from %s: %w", cm.seedDataPath, err)
		}

		// Execute seed data SQL
		if _, err := db.ExecContext(ctx, string(seedData)); err != nil {
			return fmt.Errorf("failed to execute seed data: %w", err)
		}
	}

	return nil
}

// Hash returns the combined hash of migrations and seed data
func (cm *CustomMigrator) Hash() (string, error) {
	// Get the migrations hash
	migrationsHash, err := cm.GooseMigrator.Hash()
	if err != nil {
		return "", err
	}

	// If we have seed data, include it in the hash
	if cm.seedDataPath != "" {
		seedData, err := os.ReadFile(cm.seedDataPath)
		if err != nil {
			return "", fmt.Errorf("failed to read seed data for hash: %w", err)
		}
		// Simple hash combination - in production you might want something more sophisticated
		return fmt.Sprintf("%s-%x", migrationsHash, len(seedData)), nil
	}

	return migrationsHash, nil
}

// NewDB returns a test database connection using pgtestdb with migrations and seed data
func NewDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// For now, fall back to a simpler approach that directly sets up the test environment
	// This avoids the pgtestdb template hash issues we're seeing

	config := getConfig()

	// Build connection URL directly to the postgres database (not banterbus)
	baseDbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?%s",
		config.User, config.Password, config.Host, config.Port, config.Options)

	// Create a unique test database name with proper sanitization
	sanitizedName := strings.ToLower(t.Name())
	sanitizedName = strings.ReplaceAll(sanitizedName, "/", "_")
	sanitizedName = strings.ReplaceAll(sanitizedName, " ", "_")
	sanitizedName = strings.ReplaceAll(sanitizedName, ",", "_")
	sanitizedName = strings.ReplaceAll(sanitizedName, "'", "_")
	sanitizedName = strings.ReplaceAll(sanitizedName, "\"", "_")
	sanitizedName = strings.ReplaceAll(sanitizedName, "-", "_")
	sanitizedName = strings.ReplaceAll(sanitizedName, ".", "_")

	// Truncate if too long and add a hash for uniqueness
	if len(sanitizedName) > 30 {
		hash := sha256.Sum256([]byte(sanitizedName))
		sanitizedName = fmt.Sprintf("%s_%x", sanitizedName[:20], hash[:4])
	}

	// Add timestamp and random for uniqueness (PostgreSQL DB names max 63 chars)
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(9999))
	testDBName := fmt.Sprintf("bt_%s_%d_%s_%d", sanitizedName, time.Now().Unix(), randomNum.String(), os.Getpid())

	// Final length check
	if len(testDBName) > 63 {
		hash := sha256.Sum256([]byte(testDBName))
		testDBName = fmt.Sprintf("bt_%x", hash[:15])
	}

	// Connect to postgres database to create test database
	basePool, err := pgxpool.New(context.Background(), baseDbURL)
	if err != nil {
		t.Fatalf("failed to connect to postgres database: %v", err)
	}
	defer basePool.Close()

	// Create test database
	_, err = basePool.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		t.Fatalf("failed to create test database %s: %v", testDBName, err)
	}

	// Register cleanup to drop the test database
	t.Cleanup(func() {
		// Recreate connection to postgres database for cleanup
		cleanupPool, err := pgxpool.New(context.Background(), baseDbURL)
		if err != nil {
			t.Logf("failed to connect for cleanup: %v", err)
			return
		}
		defer cleanupPool.Close()

		// Drop the test database
		_, err = cleanupPool.Exec(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
		if err != nil {
			t.Logf("failed to drop test database %s: %v", testDBName, err)
		}
	})

	// Connect to the test database
	testDbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
		config.User, config.Password, config.Host, config.Port, testDBName, config.Options)

	// Configure pgxpool with UUID support
	pgxConfig, err := pgxpool.ParseConfig(testDbURL)
	if err != nil {
		t.Fatalf("failed to parse test database URL: %v", err)
	}

	pgxConfig.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}

	// Verify connection works
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	// Apply migrations manually using goose
	if err := applyMigrationsAndSeedData(context.Background(), pool); err != nil {
		t.Fatalf("failed to apply migrations and seed data: %v", err)
	}

	// Register cleanup for the pool
	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

// applyMigrationsAndSeedData applies migrations and seed data to the test database
func applyMigrationsAndSeedData(ctx context.Context, pool *pgxpool.Pool) error {
	// Get a standard database connection for goose
	dbURL := pool.Config().ConnString()
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection for migrations: %w", err)
	}
	defer db.Close()

	// Apply migrations using goose
	migrationsDir := os.Getenv("BANTERBUS_MIGRATIONS_DIR")
	if migrationsDir == "" {
		projectRoot := findProjectRoot()
		migrationsDir = path.Join(projectRoot, "internal", "store", "db", "sqlc", "migrations")
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Apply seed data
	seedDataPath := getSeedDataScript()
	if seedDataPath != "" {
		seedData, err := os.ReadFile(seedDataPath)
		if err != nil {
			return fmt.Errorf("failed to read seed data from %s: %w", seedDataPath, err)
		}

		// Execute seed data SQL using the pgx pool
		if _, err := pool.Exec(ctx, string(seedData)); err != nil {
			return fmt.Errorf("failed to execute seed data: %w", err)
		}
	}

	return nil
}

// cleanupTestData removes all dynamic data but keeps seed data
func cleanupTestData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()

	// Delete in order to respect foreign key constraints
	// The challenge is that rooms.host_player references players.id,
	// but rooms_players.player_id also references players.id
	// So we need to delete rooms and rooms_players before players
	tables := []string{
		"fibbing_it_scores",
		"fibbing_it_votes",
		"fibbing_it_answers",
		"fibbing_it_player_roles",
		"fibbing_it_rounds",
		"game_state",
		"rooms_players", // Delete room-player relationships first
		"rooms",         // Delete rooms (which reference players as host)
		"players",       // Finally delete players
	}

	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: failed to cleanup table %s: %v", table, err)
		}
	}
}

// CreateDB is an alias for NewDB for backward compatibility
func CreateDB(ctx context.Context) (*pgxpool.Pool, error) {
	panic("CreateDB is deprecated, use NewDB(t) in your tests instead")
}

// RemoveDB is no longer needed with pgtestdb as cleanup is automatic
func RemoveDB(ctx context.Context, pool *pgxpool.Pool) error {
	// pgtestdb handles cleanup automatically via t.Cleanup()
	// Just close the pool here
	pool.Close()
	return nil
}
