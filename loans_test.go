package main

import (
	"testing"
	"time"
)

func TestNewLoan(t *testing.T) {
	tests := []struct {
		name        string
		principal   int64
		apr         float64
		startDate   time.Time
		expectError error
		weeklyDue   int64
		outstanding int64
	}{
		{
			name:        "valid default loan",
			principal:   5_000_000,
			apr:         0.10,
			startDate:   time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC),
			expectError: nil,
			weeklyDue:   110_000,
			outstanding: 5_500_000,
		},
		{
			name:        "negative principal",
			principal:   -1_000_000,
			apr:         0.10,
			startDate:   time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC),
			expectError: ErrInvalidRequest,
		},
		{
			name:        "negative rate",
			principal:   5_000_000,
			apr:         -0.05,
			startDate:   time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC),
			expectError: ErrInvalidRequest,
		},
		{
			name:        "unsupported product - not divisible by 50",
			principal:   1_234_567,
			apr:         0.10,
			startDate:   time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC),
			expectError: ErrUnsupportedProduct,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loan, err := NewLoan("test-id", tt.principal, tt.apr, tt.startDate)

			if tt.expectError != nil {
				if err != tt.expectError {
					t.Fatalf("expected error %v, got %v", tt.expectError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if loan.WeeklyDue != tt.weeklyDue {
				t.Errorf("expected weekly due %d, got %d", tt.weeklyDue, loan.WeeklyDue)
			}

			if loan.Outstanding != tt.outstanding {
				t.Errorf("expected outstanding %d, got %d", tt.outstanding, loan.Outstanding)
			}

			// Verify schedule
			if len(loan.Schedule) != 50 {
				t.Errorf("expected 50 weeks in schedule, got %d", len(loan.Schedule))
			}

			for i, week := range loan.Schedule {
				if week.Index != i+1 {
					t.Errorf("week %d has incorrect index %d", i, week.Index)
				}
				if week.Amount != tt.weeklyDue {
					t.Errorf("week %d has incorrect amount %d, expected %d", i, week.Amount, tt.weeklyDue)
				}
				if week.Paid {
					t.Errorf("week %d should not be paid initially", i)
				}
			}
		})
	}
}

func TestLoanGetOutstanding(t *testing.T) {
	loan, err := NewLoan("test", 5_000_000, 0.10, time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("failed to create loan: %v", err)
	}

	tests := []struct {
		name            string
		paidWeeks       []int
		expectedBalance int64
	}{
		{
			name:            "no payments",
			paidWeeks:       []int{},
			expectedBalance: 5_500_000,
		},
		{
			name:            "one payment",
			paidWeeks:       []int{0},
			expectedBalance: 5_390_000,
		},
		{
			name:            "two payments",
			paidWeeks:       []int{0, 1},
			expectedBalance: 5_280_000,
		},
		{
			name:            "49 payments",
			paidWeeks:       make([]int, 49),
			expectedBalance: 110_000,
		},
	}

	for i := range tests[3].paidWeeks {
		tests[3].paidWeeks[i] = i
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset loan state
			for i := range loan.Schedule {
				loan.Schedule[i].Paid = false
				loan.Schedule[i].PaidAt = nil
			}

			// Mark specified weeks as paid
			for _, weekIndex := range tt.paidWeeks {
				loan.Schedule[weekIndex].Paid = true
				now := time.Now()
				loan.Schedule[weekIndex].PaidAt = &now
			}

			outstanding := loan.GetOutstanding()
			if outstanding != tt.expectedBalance {
				t.Errorf("expected outstanding %d, got %d", tt.expectedBalance, outstanding)
			}
		})
	}
}

func TestLoanMakePayment(t *testing.T) {
	loan, err := NewLoan("test", 5_000_000, 0.10, time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("failed to create loan: %v", err)
	}

	now := time.Date(2025, 8, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		amount      int64
		expectError error
		setup       func()
	}{
		{
			name:        "valid first payment",
			amount:      110_000,
			expectError: nil,
			setup:       func() {},
		},
		{
			name:        "wrong amount",
			amount:      100_000,
			expectError: ErrWrongAmount,
			setup:       func() {},
		},
		{
			name:        "all weeks paid",
			amount:      110_000,
			expectError: ErrAlreadyPaid,
			setup: func() {
				for i := range loan.Schedule {
					loan.Schedule[i].Paid = true
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset loan state
			for i := range loan.Schedule {
				loan.Schedule[i].Paid = false
				loan.Schedule[i].PaidAt = nil
			}
			loan.PaidCount = 0
			loan.Outstanding = 5_500_000

			tt.setup()

			err := loan.MakePayment(tt.amount, now)

			if tt.expectError != nil {
				if err != tt.expectError {
					t.Fatalf("expected error %v, got %v", tt.expectError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify first week is paid
			if !loan.Schedule[0].Paid {
				t.Error("first week should be marked as paid")
			}

			if loan.Schedule[0].PaidAt == nil {
				t.Error("first week should have paid_at timestamp")
			}

			if loan.PaidCount != 1 {
				t.Errorf("expected paid count 1, got %d", loan.PaidCount)
			}
		})
	}
}

func TestLoanWeekIndexAt(t *testing.T) {
	startDate := time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC)
	loan, err := NewLoan("test", 5_000_000, 0.10, startDate)
	if err != nil {
		t.Fatalf("failed to create loan: %v", err)
	}

	tests := []struct {
		name          string
		now           time.Time
		expectedWeek  int
	}{
		{
			name:         "before start",
			now:          time.Date(2025, 8, 14, 0, 0, 0, 0, time.UTC),
			expectedWeek: 1,
		},
		{
			name:         "week 1 day 0",
			now:          time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC),
			expectedWeek: 1,
		},
		{
			name:         "week 1 day 6",
			now:          time.Date(2025, 8, 21, 0, 0, 0, 0, time.UTC),
			expectedWeek: 1,
		},
		{
			name:         "week 2 day 0",
			now:          time.Date(2025, 8, 22, 0, 0, 0, 0, time.UTC),
			expectedWeek: 2,
		},
		{
			name:         "week 3 day 0",
			now:          time.Date(2025, 8, 29, 0, 0, 0, 0, time.UTC),
			expectedWeek: 3,
		},
		{
			name:         "far future - should cap at 50",
			now:          time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedWeek: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			week := loan.WeekIndexAt(tt.now)
			if week != tt.expectedWeek {
				t.Errorf("expected week %d, got %d", tt.expectedWeek, week)
			}
		})
	}
}

func TestLoanIsDelinquent(t *testing.T) {
	startDate := time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)
	loan, err := NewLoan("test", 5_000_000, 0.10, startDate)
	if err != nil {
		t.Fatalf("failed to create loan: %v", err)
	}

	tests := []struct {
		name               string
		now                time.Time
		paidWeeks          []int
		expectedDelinquent bool
		expectedStreak     int
		expectedObserved   int
	}{
		{
			name:               "week 1 - not delinquent",
			now:                time.Date(2025, 8, 7, 0, 0, 0, 0, time.UTC),
			paidWeeks:          []int{},
			expectedDelinquent: false,
			expectedStreak:     0,
			expectedObserved:   1,
		},
		{
			name:               "week 2 - not delinquent (only week 1 unpaid)",
			now:                time.Date(2025, 8, 14, 0, 0, 0, 0, time.UTC),
			paidWeeks:          []int{},
			expectedDelinquent: false,
			expectedStreak:     0,
			expectedObserved:   2,
		},
		{
			name:               "week 3 - delinquent (weeks 1,2 unpaid)",
			now:                time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC),
			paidWeeks:          []int{},
			expectedDelinquent: true,
			expectedStreak:     2,
			expectedObserved:   3,
		},
		{
			name:               "week 3 - not delinquent after catch-up",
			now:                time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC),
			paidWeeks:          []int{0, 1},
			expectedDelinquent: false,
			expectedStreak:     0,
			expectedObserved:   3,
		},
		{
			name:               "future start date - not delinquent",
			now:                time.Date(2025, 7, 31, 0, 0, 0, 0, time.UTC),
			paidWeeks:          []int{},
			expectedDelinquent: false,
			expectedStreak:     0,
			expectedObserved:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset loan state
			for i := range loan.Schedule {
				loan.Schedule[i].Paid = false
				loan.Schedule[i].PaidAt = nil
			}

			// Mark specified weeks as paid
			for _, weekIndex := range tt.paidWeeks {
				loan.Schedule[weekIndex].Paid = true
				now := time.Now()
				loan.Schedule[weekIndex].PaidAt = &now
			}

			delinquent, streak, observed := loan.IsDelinquent(tt.now)

			if delinquent != tt.expectedDelinquent {
				t.Errorf("expected delinquent %v, got %v", tt.expectedDelinquent, delinquent)
			}

			if streak != tt.expectedStreak {
				t.Errorf("expected streak %d, got %d", tt.expectedStreak, streak)
			}

			if observed != tt.expectedObserved {
				t.Errorf("expected observed week %d, got %d", tt.expectedObserved, observed)
			}
		})
	}
}

// TestPropertyFullPayment verifies that 50 payments equals total due
func TestPropertyFullPayment(t *testing.T) {
	loan, err := NewLoan("test", 5_000_000, 0.10, time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("failed to create loan: %v", err)
	}

	totalDue := int64(5_500_000)
	totalPaid := int64(0)
	now := time.Now()

	// Pay all 50 weeks
	for i := 0; i < 50; i++ {
		err := loan.MakePayment(loan.WeeklyDue, now)
		if err != nil {
			t.Fatalf("payment %d failed: %v", i+1, err)
		}
		totalPaid += loan.WeeklyDue
	}

	// Verify properties
	if totalPaid != totalDue {
		t.Errorf("total paid %d != total due %d", totalPaid, totalDue)
	}

	outstanding := loan.GetOutstanding()
	if outstanding != 0 {
		t.Errorf("expected outstanding 0, got %d", outstanding)
	}

	// Verify cannot pay more
	err = loan.MakePayment(loan.WeeklyDue, now)
	if err != ErrAlreadyPaid {
		t.Errorf("expected ErrAlreadyPaid, got %v", err)
	}
}
