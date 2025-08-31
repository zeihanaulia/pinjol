package internal

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	// Setup test environment
	os.Exit(m.Run())
}

// TestDatabaseConnectivity tests database connectivity
func TestDatabaseConnectivity(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	err := db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}
}

// TestDatabaseMigration tests database schema migration
func TestDatabaseMigration(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	// Test InitDatabase
	err := InitDatabase(db)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Verify tables exist
	tables := []string{"loans", "loan_schedule"}
	for _, table := range tables {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check table %s: %v", table, err)
		}
		if count == 0 {
			t.Errorf("Table %s was not created", table)
		}
	}

	// Verify indexes exist
	indexes := []string{"idx_loan_schedule_loan_id", "idx_loans_start_date"}
	for _, index := range indexes {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", index).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check index %s: %v", index, err)
		}
		if count == 0 {
			t.Errorf("Index %s was not created", index)
		}
	}
}

// setupTestDatabase creates a test database connection
func setupTestDatabase(t *testing.T) *sql.DB {
	// Use SQLite for simplicity in tests
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Initialize schema
	err = InitDatabase(db)
	if err != nil {
		t.Fatalf("Failed to initialize test database schema: %v", err)
	}

	return db
}
