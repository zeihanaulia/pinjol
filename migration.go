package main

import (
	"database/sql"
	"fmt"
)

// InitDatabase initializes the database schema
func InitDatabase(db *sql.DB) error {
	// Create loans table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS loans (
			id TEXT PRIMARY KEY,
			principal INTEGER NOT NULL,
			apr REAL NOT NULL,
			start_date TEXT NOT NULL,
			weekly_due INTEGER NOT NULL,
			paid_count INTEGER NOT NULL DEFAULT 0,
			outstanding INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("failed to create loans table: %w", err)
	}

	// Create loan_schedule table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS loan_schedule (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			loan_id TEXT NOT NULL,
			week_index INTEGER NOT NULL,
			amount INTEGER NOT NULL,
			paid BOOLEAN NOT NULL DEFAULT FALSE,
			paid_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (loan_id) REFERENCES loans(id) ON DELETE CASCADE,
			UNIQUE(loan_id, week_index)
		)`)
	if err != nil {
		return fmt.Errorf("failed to create loan_schedule table: %w", err)
	}

	// Create indexes for better performance
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_loan_schedule_loan_id ON loan_schedule(loan_id)`)
	if err != nil {
		return fmt.Errorf("failed to create index on loan_schedule: %w", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_loans_start_date ON loans(start_date)`)
	if err != nil {
		return fmt.Errorf("failed to create index on loans start_date: %w", err)
	}

	return nil
}
