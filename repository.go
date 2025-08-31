package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// LoanRepository defines the interface for loan data persistence
type LoanRepository interface {
	Create(loan *Loan) error
	GetByID(id string) (*Loan, error)
	Update(loan *Loan) error
	List() ([]*Loan, error)
	Delete(id string) error
}

// SQLiteLoanRepository implements LoanRepository using SQLite
type SQLiteLoanRepository struct {
	db *sql.DB
}

// NewSQLiteLoanRepository creates a new SQLite repository
func NewSQLiteLoanRepository(db *sql.DB) *SQLiteLoanRepository {
	return &SQLiteLoanRepository{db: db}
}

// Create inserts a new loan into the database
func (r *SQLiteLoanRepository) Create(loan *Loan) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert loan
	_, err = tx.Exec(`
		INSERT INTO loans (id, principal, apr, start_date, weekly_due, paid_count, outstanding)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		loan.ID, loan.Principal, loan.APR, loan.StartDate, loan.WeeklyDue, loan.PaidCount, loan.Outstanding)
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
			return fmt.Errorf("failed to insert schedule for week %d: %w", week.Index, err)
		}
	}

	return tx.Commit()
}

// GetByID retrieves a loan by ID
func (r *SQLiteLoanRepository) GetByID(id string) (*Loan, error) {
	// Get loan
	var loan Loan
	var startDateStr string
	err := r.db.QueryRow(`
		SELECT id, principal, apr, start_date, weekly_due, paid_count, outstanding
		FROM loans WHERE id = ?`, id).Scan(
		&loan.ID, &loan.Principal, &loan.APR, &startDateStr, &loan.WeeklyDue, &loan.PaidCount, &loan.Outstanding)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrLoanNotFound
		}
		return nil, fmt.Errorf("failed to get loan: %w", err)
	}

	// Parse start date
	loan.StartDate, err = time.Parse("2006-01-02 15:04:05Z07:00", startDateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start date: %w", err)
	}

	// Get schedule
	rows, err := r.db.Query(`
		SELECT week_index, amount, paid, paid_at
		FROM loan_schedule WHERE loan_id = ? ORDER BY week_index`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}
	defer rows.Close()

	// Initialize schedule array
	var schedule [50]Week
	for rows.Next() {
		var week Week
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
	loan.Schedule = schedule

	return &loan, nil
}

// Update updates an existing loan in the database
func (r *SQLiteLoanRepository) Update(loan *Loan) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update loan
	_, err = tx.Exec(`
		UPDATE loans SET paid_count = ?, outstanding = ? WHERE id = ?`,
		loan.PaidCount, loan.Outstanding, loan.ID)
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
			UPDATE loan_schedule SET paid = ?, paid_at = ? WHERE loan_id = ? AND week_index = ?`,
			week.Paid, paidAt, loan.ID, week.Index)
		if err != nil {
			return fmt.Errorf("failed to update schedule for week %d: %w", week.Index, err)
		}
	}

	return tx.Commit()
}

// List returns all loans (for admin purposes)
func (r *SQLiteLoanRepository) List() ([]*Loan, error) {
	rows, err := r.db.Query(`
		SELECT id, principal, apr, start_date, weekly_due, paid_count, outstanding
		FROM loans ORDER BY start_date DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to list loans: %w", err)
	}
	defer rows.Close()

	var loans []*Loan
	for rows.Next() {
		var loan Loan
		var startDateStr string
		err := rows.Scan(&loan.ID, &loan.Principal, &loan.APR, &startDateStr, &loan.WeeklyDue, &loan.PaidCount, &loan.Outstanding)
		if err != nil {
			return nil, fmt.Errorf("failed to scan loan row: %w", err)
		}

		loan.StartDate, err = time.Parse("2006-01-02T15:04:05Z07:00", startDateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start date: %w", err)
		}

		loans = append(loans, &loan)
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
	_, err = tx.Exec("DELETE FROM loan_schedule WHERE loan_id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	// Delete loan
	result, err := tx.Exec("DELETE FROM loans WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete loan: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrLoanNotFound
	}

	return tx.Commit()
}
