package main

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitDatabase(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize database
	err = InitDatabase(db)
	if err != nil {
		t.Fatalf("InitDatabase failed: %v", err)
	}

	// Check if loans table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='loans'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Loans table not created: %v", err)
	}
	if tableName != "loans" {
		t.Errorf("Expected table name 'loans', got '%s'", tableName)
	}

	// Check if loan_schedule table exists
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='loan_schedule'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Loan_schedule table not created: %v", err)
	}
	if tableName != "loan_schedule" {
		t.Errorf("Expected table name 'loan_schedule', got '%s'", tableName)
	}

	// Check if indexes exist
	var indexName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_loan_schedule_loan_id'").Scan(&indexName)
	if err != nil {
		t.Fatalf("Index idx_loan_schedule_loan_id not created: %v", err)
	}

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_loans_start_date'").Scan(&indexName)
	if err != nil {
		t.Fatalf("Index idx_loans_start_date not created: %v", err)
	}

	// Test inserting a sample loan to ensure schema is correct
	_, err = db.Exec(`INSERT INTO loans (id, principal, apr, start_date, weekly_due, outstanding) VALUES (?, ?, ?, ?, ?, ?)`,
		"test-loan-1", 1000000, 0.1, "2025-09-01", 25000, 1000000)
	if err != nil {
		t.Fatalf("Failed to insert into loans table: %v", err)
	}

	// Test inserting a sample schedule
	_, err = db.Exec(`INSERT INTO loan_schedule (loan_id, week_index, amount) VALUES (?, ?, ?)`,
		"test-loan-1", 1, 25000)
	if err != nil {
		t.Fatalf("Failed to insert into loan_schedule table: %v", err)
	}
}
