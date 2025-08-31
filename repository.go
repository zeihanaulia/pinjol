package main

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

	// Parse start date
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

	// Initialize schedule array
	schedule := [50]domain.Week{}
	for rows.Next() {
		var week domain.Week
		var paidAt *time.Time
		err := rows.Scan(&week.Index, &week.Amount, &week.Paid, &paidAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule row: %w", err)
		}
		week.PaidAt = paidAt
		if week.Index >= 1 && week.Index <= 50 {
			schedule[week.Index-1] = week
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule rows: %w", err)
	}

	loan.Schedule = schedule
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
		UPDATE loans SET principal = ?, apr = ?, start_date = ?, weekly_due = ?, paid_count = ?, outstanding = ?
		WHERE id = ?`,
		loan.Principal, loan.APR, loan.StartDate.Format(time.RFC3339), loan.WeeklyDue, loan.PaidCount, loan.Outstanding, loan.ID)
	if err != nil {
		return fmt.Errorf("failed to update loan: %w", err)
	}

	// Update payments
	for _, week := range loan.Schedule {
		var paidAt *time.Time
		if week.PaidAt != nil {
			paidAt = week.PaidAt
		}
		_, err = tx.Exec(`
			UPDATE loan_schedule SET amount = ?, paid = ?, paid_at = ?
			WHERE loan_id = ? AND week_index = ?`,
			week.Amount, week.Paid, paidAt, loan.ID, week.Index)
		if err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}
	}

	return tx.Commit()
}

// List lists all loans
func (r *SQLiteLoanRepository) List() ([]*domain.Loan, error) {
	rows, err := r.db.Query(`
		SELECT id, principal, apr, start_date, weekly_due, paid_count, outstanding
		FROM loans`)
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
			return nil, fmt.Errorf("failed to parse start_date: %w", err)
		}
		loans = append(loans, loan)
	}

	return loans, nil
}

// Delete deletes a loan from the database
func (r *SQLiteLoanRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM loans WHERE id = ?`, id)
	return err
}


