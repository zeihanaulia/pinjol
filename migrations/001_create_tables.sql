-- +migrate Up
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
);

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
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_loan_schedule_loan_id ON loan_schedule(loan_id);
CREATE INDEX IF NOT EXISTS idx_loans_start_date ON loans(start_date);

-- +migrate Down
DROP INDEX IF EXISTS idx_loans_start_date;
DROP INDEX IF EXISTS idx_loan_schedule_loan_id;
DROP TABLE IF EXISTS loan_schedule;
DROP TABLE IF EXISTS loans;
