-- +migrate Up
CREATE TABLE IF NOT EXISTS loans (
    id TEXT PRIMARY KEY,
    principal INTEGER NOT NULL,
    apr REAL NOT NULL,
    start_date TEXT NOT NULL,
    weekly_due INTEGER NOT NULL,
    paid_count INTEGER NOT NULL DEFAULT 0,
    outstanding INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS payments (
    loan_id TEXT NOT NULL,
    week_index INTEGER NOT NULL,
    amount INTEGER NOT NULL,
    paid BOOLEAN NOT NULL DEFAULT FALSE,
    paid_at TEXT,
    PRIMARY KEY (loan_id, week_index),
    FOREIGN KEY (loan_id) REFERENCES loans(id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS loans;
