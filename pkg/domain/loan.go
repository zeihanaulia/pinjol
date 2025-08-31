package domain

import (
	"math"
	"time"
)

// Loan represents a billing loan with flat interest
type Loan struct {
	ID          string
	Principal   int64
	APR         float64
	StartDate   time.Time
	WeeklyDue   int64
	Schedule    [50]Week
	PaidCount   int
	Outstanding int64
}

// Week represents a single week in the payment schedule
type Week struct {
	Index  int
	Amount int64
	Paid   bool
	PaidAt *time.Time
}

// NewLoan creates a new loan with the specified parameters
func NewLoan(id string, principal int64, apr float64, startDate time.Time) (*Loan, error) {
	if principal <= 0 {
		return nil, ErrInvalidRequest
	}
	if apr < 0 {
		return nil, ErrInvalidRequest
	}

	// Calculate total due with flat interest
	totalDue := int64(math.Round(float64(principal) * (1 + apr)))

	// Check if weekly amount would be integral
	if totalDue%50 != 0 {
		return nil, ErrUnsupportedProduct
	}

	weeklyDue := totalDue / 50

	loan := &Loan{
		ID:          id,
		Principal:   principal,
		APR:         apr,
		StartDate:   startDate,
		WeeklyDue:   weeklyDue,
		PaidCount:   0,
		Outstanding: totalDue,
	}

	// Initialize schedule
	for i := 0; i < 50; i++ {
		loan.Schedule[i] = Week{
			Index:  i + 1,
			Amount: weeklyDue,
			Paid:   false,
		}
	}

	return loan, nil
}

// GetOutstanding recomputes and returns the outstanding amount
func (l *Loan) GetOutstanding() int64 {
	outstanding := int64(0)
	for _, week := range l.Schedule {
		if !week.Paid {
			outstanding += week.Amount
		}
	}
	l.Outstanding = outstanding
	return outstanding
}

// MakePayment processes a payment for the oldest unpaid week
func (l *Loan) MakePayment(amount int64, now time.Time) error {
	// Find the first unpaid week
	firstUnpaidIndex := -1
	for i, week := range l.Schedule {
		if !week.Paid {
			firstUnpaidIndex = i
			break
		}
	}

	// Check if all weeks are already paid
	if firstUnpaidIndex == -1 {
		return ErrAlreadyPaid
	}

	// Check if payment amount matches the required amount
	if amount != l.Schedule[firstUnpaidIndex].Amount {
		return ErrWrongAmount
	}

	// Process the payment
	l.Schedule[firstUnpaidIndex].Paid = true
	l.Schedule[firstUnpaidIndex].PaidAt = &now
	l.PaidCount++
	l.Outstanding -= amount

	return nil
}

// WeekIndexAt returns the week index (1-50) for the given time
func (l *Loan) WeekIndexAt(now time.Time) int {
	if now.Before(l.StartDate) {
		return 1
	}

	daysSinceStart := int(now.Sub(l.StartDate).Hours() / 24)
	weekIndex := (daysSinceStart / 7) + 1

	if weekIndex > 50 {
		return 50
	}
	return weekIndex
}

// IsDelinquent checks if the loan is delinquent based on the latest two scheduled weeks.
func (l *Loan) IsDelinquent(now time.Time) (bool, int, int) {
	observedWeek := l.WeekIndexAt(now)

	// If we're in the first two weeks, cannot be delinquent
	if observedWeek < 3 {
		return false, 0, observedWeek
	}

	// Check if the latest two scheduled weeks are both unpaid
	week1Index := observedWeek - 2 // 0-based index for array
	week2Index := observedWeek - 1 // 0-based index for array

	if week1Index >= 0 && week1Index < 50 && week2Index >= 0 && week2Index < 50 {
		week1Unpaid := !l.Schedule[week1Index].Paid
		week2Unpaid := !l.Schedule[week2Index].Paid

		if week1Unpaid && week2Unpaid {
			return true, 2, observedWeek
		}
	}

	return false, 0, observedWeek
}
