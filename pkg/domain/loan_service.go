package domain

import (
	"fmt"
	"time"
)

// LoanService handles business logic for loans
type LoanService struct{}

// NewLoanService creates a new loan service
func NewLoanService() *LoanService {
	return &LoanService{}
}

// CreateLoanRequest represents the request to create a loan
type CreateLoanRequest struct {
	Principal  int64
	AnnualRate float64
	StartDate  time.Time
}

// ValidateCreateLoanRequest validates the create loan request
func (s *LoanService) ValidateCreateLoanRequest(req CreateLoanRequest) error {
	if req.Principal <= 0 {
		return fmt.Errorf("principal must be greater than 0")
	}
	if req.Principal > 5_000_000 {
		return fmt.Errorf("principal exceeds maximum allowed limit")
	}
	if req.AnnualRate < 0 {
		return fmt.Errorf("annual rate cannot be negative")
	}
	if req.AnnualRate > 0.50 {
		return fmt.Errorf("annual rate exceeds maximum allowed limit")
	}
	return nil
}

// CreateLoan creates a new loan
func (s *LoanService) CreateLoan(id string, req CreateLoanRequest) (*Loan, error) {
	if err := s.ValidateCreateLoanRequest(req); err != nil {
		return nil, err
	}
	return NewLoan(id, req.Principal, req.AnnualRate, req.StartDate)
}

// PaymentRequest represents the request to make a payment
type PaymentRequest struct {
	Amount int64
	Now    time.Time
}

// ValidatePaymentRequest validates the payment request
func (s *LoanService) ValidatePaymentRequest(req PaymentRequest) error {
	if req.Amount <= 0 {
		return fmt.Errorf("payment amount must be greater than 0")
	}
	return nil
}

// ProcessPayment processes a payment on a loan
func (s *LoanService) ProcessPayment(loan *Loan, req PaymentRequest) error {
	if err := s.ValidatePaymentRequest(req); err != nil {
		return err
	}
	return loan.MakePayment(req.Amount, req.Now)
}

// Value objects for responses

// LoanResponse represents the loan data for response
type LoanResponse struct {
	ID          string
	Principal   int64
	APR         float64
	StartDate   time.Time
	WeeklyDue   int64
	Schedule    [50]WeekResponse
	PaidCount   int
	Outstanding int64
}

// WeekResponse represents a week in the schedule
type WeekResponse struct {
	Index  int
	Amount int64
	Paid   bool
	PaidAt *time.Time
}

// PaymentResponse represents the payment result
type PaymentResponse struct {
	PaidWeek             int
	RemainingOutstanding int64
}

// OutstandingResponse represents the outstanding amount
type OutstandingResponse struct {
	Outstanding int64
}

// DelinquencyResponse represents the delinquency status
type DelinquencyResponse struct {
	Delinquent   bool
	Streak       int
	ObservedWeek int
}
