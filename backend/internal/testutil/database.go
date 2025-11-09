package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/mahendrakalkura/saas-blueprint/internal/config"
	"github.com/mahendrakalkura/saas-blueprint/internal/database"
	_ "github.com/lib/pq"
)

// TestDB represents a test database connection
type TestDB struct {
	DB     *database.DB
	Config *config.DatabaseConfig
}

// SetupTestDB creates a test database and returns a connection
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	cfg := &config.DatabaseConfig{
		Host:            getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:            getEnvOrDefault("TEST_DB_PORT", "5432"),
		User:            getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password:        getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		Name:            getEnvOrDefault("TEST_DB_NAME", "postgres"),
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
	}

	// Create a unique test database name
	testDBName := fmt.Sprintf("test_saas_%s", t.Name())
	testDBName = sanitizeDBName(testDBName)

	// Connect to default database to create test database
	defaultDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name,
	)

	defaultDB, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		t.Fatalf("Failed to connect to default database: %v", err)
	}
	defer defaultDB.Close()

	// Drop test database if exists
	_, _ = defaultDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))

	// Create test database
	_, err = defaultDB.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Connect to test database
	testCfg := *cfg
	testCfg.Name = testDBName

	db, err := database.New(&testCfg)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := runMigrations(db.DB, "../../../db/migrations"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return &TestDB{
		DB:     db,
		Config: &testCfg,
	}
}

// TeardownTestDB drops the test database and closes the connection
func (tdb *TestDB) TeardownTestDB(t *testing.T) {
	t.Helper()

	dbName := tdb.Config.Name

	// Close connection
	tdb.DB.Close()

	// Connect to default database to drop test database
	defaultDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		tdb.Config.Host, tdb.Config.Port, tdb.Config.User, tdb.Config.Password,
	)

	defaultDB, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		t.Logf("Warning: Failed to connect to default database: %v", err)
		return
	}
	defer defaultDB.Close()

	// Drop test database
	_, err = defaultDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		t.Logf("Warning: Failed to drop test database: %v", err)
	}
}

// CleanTables truncates all tables in the test database
func (tdb *TestDB) CleanTables(t *testing.T, tables ...string) {
	t.Helper()

	for _, table := range tables {
		_, err := tdb.DB.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}
}

// runMigrations runs SQL migration files
func runMigrations(db *sql.DB, migrationsPath string) error {
	// For simplicity, we'll just create the essential tables manually
	// In production, you'd use a migration tool like golang-migrate
	schema := `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			is_active BOOLEAN DEFAULT TRUE,
			email_verified BOOLEAN DEFAULT FALSE,
			email_verification_token VARCHAR(255),
			email_verification_expires_at TIMESTAMP,
			password_reset_token VARCHAR(255),
			password_reset_expires_at TIMESTAMP,
			avatar_url VARCHAR(500),
			mfa_enabled BOOLEAN DEFAULT FALSE,
			mfa_secret VARCHAR(255),
			mfa_backup_codes TEXT[],
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			refresh_token VARCHAR(255) UNIQUE NOT NULL,
			refresh_token_expires_at TIMESTAMP NOT NULL,
			user_agent VARCHAR(500),
			ip_address VARCHAR(50),
			revoked BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token);
	`

	_, err := db.Exec(schema)
	return err
}

// sanitizeDBName removes invalid characters from database name
func sanitizeDBName(name string) string {
	// Replace invalid characters with underscore
	result := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			result += string(c)
		} else {
			result += "_"
		}
	}
	// Ensure it starts with a letter
	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		result = "t_" + result
	}
	// Limit length
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
