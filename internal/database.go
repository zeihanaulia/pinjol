package internal

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"pinjol/pkg/domain"
)

// SQLiteLoanRepository implements domain.LoanRepository using SQLite
type SQLiteLoanRepository struct {
	db *sql.DB
}

// NewSQLiteLoanRepository creates a new SQLite repository
func NewSQLiteLoanRepository(db *sql.DB) *SQLiteLoanRepository {
	return &SQLiteLoanRepository{db: db}
}

// Create inserts a new loan into the database
func (r *SQLiteLoanRepository) Create(loan *domain.Loan) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert loan
	_, err = tx.Exec(`
		INSERT INTO loans (id, principal, apr, start_date, weekly_due, paid_count, outstanding)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		loan.ID, loan.Principal, loan.APR, loan.StartDate.Format(time.RFC3339), loan.WeeklyDue, loan.PaidCount, loan.Outstanding)
	if err != nil {
		return fmt.Errorf("failed to insert loan: %w", err)
	}

	// Insert schedule
	for _, week := range loan.Schedule {
		var paidAt *time.Time
		if week.PaidAt != nil {
			paidAt = week.PaidAt
		}
		_, err = tx.Exec(`
			INSERT INTO loan_schedule (loan_id, week_index, amount, paid, paid_at)
			VALUES (?, ?, ?, ?, ?)`,
			loan.ID, week.Index, week.Amount, week.Paid, paidAt)
		if err != nil {
			return fmt.Errorf("failed to insert schedule: %w", err)
		}
	}

	return tx.Commit()
}

// GetByID retrieves a loan by ID
func (r *SQLiteLoanRepository) GetByID(id string) (*domain.Loan, error) {
	loan := &domain.Loan{}
	var startDateStr string

	err := r.db.QueryRow(`
		SELECT id, principal, apr, start_date, weekly_due, paid_count, outstanding
		FROM loans WHERE id = ?`, id).Scan(
		&loan.ID, &loan.Principal, &loan.APR, &startDateStr, &loan.WeeklyDue, &loan.PaidCount, &loan.Outstanding)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrLoanNotFound
		}
		return nil, fmt.Errorf("failed to get loan: %w", err)
	}

	loan.StartDate, err = time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start date: %w", err)
	}

	// Load schedule
	rows, err := r.db.Query(`
		SELECT week_index, amount, paid, paid_at
		FROM loan_schedule WHERE loan_id = ? ORDER BY week_index`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		var week domain.Week
		var paidAt *time.Time
		err := rows.Scan(&week.Index, &week.Amount, &week.Paid, &paidAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		if paidAt != nil {
			week.PaidAt = paidAt
		}
		if i < 50 {
			loan.Schedule[i] = week
			i++
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule rows: %w", err)
	}

	// Ensure we loaded exactly 50 weeks
	// if i != 50 {
	// 	return nil, fmt.Errorf("expected 50 schedule weeks, got %d", i)
	// }

	return loan, nil
}

// Update updates a loan in the database
func (r *SQLiteLoanRepository) Update(loan *domain.Loan) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update loan
	_, err = tx.Exec(`
		UPDATE loans SET paid_count = ?, outstanding = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, loan.PaidCount, loan.Outstanding, loan.ID)
	if err != nil {
		return fmt.Errorf("failed to update loan: %w", err)
	}

	// Update schedule
	for _, week := range loan.Schedule {
		var paidAt *time.Time
		if week.PaidAt != nil {
			paidAt = week.PaidAt
		}
		_, err = tx.Exec(`
			UPDATE loan_schedule SET paid = ?, paid_at = ?, updated_at = CURRENT_TIMESTAMP
			WHERE loan_id = ? AND week_index = ?`, week.Paid, paidAt, loan.ID, week.Index)
		if err != nil {
			return fmt.Errorf("failed to update schedule: %w", err)
		}
	}

	return tx.Commit()
}

// List retrieves all loans from the database
func (r *SQLiteLoanRepository) List() ([]*domain.Loan, error) {
	rows, err := r.db.Query(`
		SELECT id, principal, apr, start_date, weekly_due, paid_count, outstanding
		FROM loans ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to list loans: %w", err)
	}
	defer rows.Close()

	var loans []*domain.Loan
	for rows.Next() {
		loan := &domain.Loan{}
		var startDateStr string
		err := rows.Scan(&loan.ID, &loan.Principal, &loan.APR, &startDateStr, &loan.WeeklyDue, &loan.PaidCount, &loan.Outstanding)
		if err != nil {
			return nil, fmt.Errorf("failed to scan loan: %w", err)
		}
		loan.StartDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start date: %w", err)
		}
		loans = append(loans, loan)
	}

	return loans, nil
}

// Delete removes a loan from the database
func (r *SQLiteLoanRepository) Delete(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete schedule first (foreign key constraint)
	_, err = tx.Exec(`DELETE FROM loan_schedule WHERE loan_id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	// Delete loan
	_, err = tx.Exec(`DELETE FROM loans WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete loan: %w", err)
	}

	return tx.Commit()
}

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
