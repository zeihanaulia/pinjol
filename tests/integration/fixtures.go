package integration

import (
	"database/sql"
	"time"
)

// TestLoan represents test loan data
type TestLoan struct {
	ID          string
	Principal   int64
	APR         float64
	StartDate   time.Time
	WeeklyDue   int64
	PaidCount   int
	Outstanding int64
}

// SeedLoans inserts test loan data into database
func SeedLoans(db *sql.DB) ([]TestLoan, error) {
	loans := []TestLoan{
		{
			ID:          "test-loan-001",
			Principal:   5_000_000,
			APR:         0.10,
			StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			WeeklyDue:   110_000,
			PaidCount:   0,
			Outstanding: 5_500_000,
		},
		{
			ID:          "test-loan-002",
			Principal:   3_000_000,
			APR:         0.08,
			StartDate:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			WeeklyDue:   66_000,
			PaidCount:   5,
			Outstanding: 3_300_000 - (5 * 66_000),
		},
		{
			ID:          "test-loan-003",
			Principal:   1_000_000,
			APR:         0.05,
			StartDate:   time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			WeeklyDue:   26_000,
			PaidCount:   50, // Fully paid
			Outstanding: 0,
		},
	}

	// Insert loans
	for _, loan := range loans {
		_, err := db.Exec(`
			INSERT INTO loans (id, principal, apr, start_date, weekly_due, paid_count, outstanding)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			loan.ID, loan.Principal, loan.APR, loan.StartDate.Format(time.RFC3339), loan.WeeklyDue, loan.PaidCount, loan.Outstanding)
		if err != nil {
			return nil, err
		}

		// Insert schedule (simplified - only insert paid weeks for testing)
		for i := 0; i < 50; i++ {
			paid := i < loan.PaidCount
			var paidAt *time.Time
			if paid {
				t := loan.StartDate.AddDate(0, 0, i*7)
				paidAt = &t
			}

			_, err = db.Exec(`
				INSERT INTO loan_schedule (loan_id, week_index, amount, paid, paid_at)
				VALUES (?, ?, ?, ?, ?)`,
				loan.ID, i+1, loan.WeeklyDue, paid, paidAt)
			if err != nil {
				return nil, err
			}
		}
	}

	return loans, nil
}

// CleanUp removes all test data
func CleanUp(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM loan_schedule")
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM loans")
	if err != nil {
		return err
	}

	return nil
}
