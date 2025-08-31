-- +migrate Up
CREATE TABLE IF NOT EXISTS borrowers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_borrowers_email ON borrowers(email);
CREATE INDEX IF NOT EXISTS idx_borrowers_created_at ON borrowers(created_at);

-- +migrate Down
DROP INDEX IF EXISTS idx_borrowers_created_at;
DROP INDEX IF EXISTS idx_borrowers_email;
DROP TABLE IF EXISTS borrowers;
