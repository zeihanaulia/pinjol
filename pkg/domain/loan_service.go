package domain

import (
	"encoding/base32"
	"fmt"
	"strconv"
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

// CreateLoanFromRequest handles the complete loan creation process from HTTP request
func (s *LoanService) CreateLoanFromRequest(principal int64, annualRate float64, startDateStr string) (*LoanResponse, error) {
	// Parse start date
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, &ValidationError{
			Field:   "start_date",
			Message: "Start date format is invalid",
			Code:    "INVALID_START_DATE_FORMAT",
			Details: map[string]string{
				"provided": startDateStr,
				"required": "YYYY-MM-DD format",
				"example":  "2025-08-31",
			},
		}
	}

	// Default annual rate
	if annualRate == 0 {
		annualRate = 0.10
	}

	req := CreateLoanRequest{
		Principal:  principal,
		AnnualRate: annualRate,
		StartDate:  startDate,
	}

	// Validate request
	if err := s.ValidateCreateLoanRequest(req); err != nil {
		return nil, s.mapValidationError(err, req)
	}

	// Generate unique ID
	id := s.generateLoanID()

	// Create loan
	loan, err := s.CreateLoan(id, req)
	if err != nil {
		return nil, s.mapBusinessError(err, req)
	}

	// Convert to response
	return s.convertLoanToResponse(loan), nil
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

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string
	Message string
	Code    string
	Details map[string]string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// BusinessError represents a business logic error
type BusinessError struct {
	Message string
	Code    string
	Details map[string]string
}

func (e *BusinessError) Error() string {
	return e.Message
}

// Helper methods for error mapping
func (s *LoanService) mapValidationError(err error, req CreateLoanRequest) error {
	errMsg := err.Error()
	switch {
	case req.Principal <= 0:
		return &ValidationError{
			Field:   "principal",
			Message: "Principal amount must be greater than 0",
			Code:    "INVALID_PRINCIPAL",
			Details: map[string]string{
				"provided":  fmt.Sprintf("%d", req.Principal),
				"required":  "Must be a positive integer greater than 0",
				"min_value": "1",
				"max_value": "5000000 (5 million)",
			},
		}
	case req.Principal > 5_000_000:
		return &ValidationError{
			Field:   "principal",
			Message: "Principal amount exceeds maximum allowed limit",
			Code:    "PRINCIPAL_TOO_HIGH",
			Details: map[string]string{
				"provided": fmt.Sprintf("%d", req.Principal),
				"maximum":  "5000000 (5 million)",
				"reason":   "Higher amounts may not result in integral weekly payments",
			},
		}
	case req.AnnualRate < 0:
		return &ValidationError{
			Field:   "annual_rate",
			Message: "Annual rate cannot be negative",
			Code:    "INVALID_ANNUAL_RATE",
			Details: map[string]string{
				"provided":  fmt.Sprintf("%.2f", req.AnnualRate),
				"required":  "Must be a non-negative decimal (0.00 - 0.50)",
				"min_value": "0.00",
				"max_value": "0.50 (50%)",
			},
		}
	case req.AnnualRate > 0.50:
		return &ValidationError{
			Field:   "annual_rate",
			Message: "Annual rate exceeds maximum allowed limit",
			Code:    "ANNUAL_RATE_TOO_HIGH",
			Details: map[string]string{
				"provided": fmt.Sprintf("%.2f", req.AnnualRate),
				"maximum":  "0.50 (50%)",
				"reason":   "Higher rates may not result in integral weekly payments",
			},
		}
	default:
		return &ValidationError{
			Field:   "request",
			Message: errMsg,
			Code:    "INVALID_LOAN_PARAMETERS",
			Details: map[string]string{
				"error": errMsg,
			},
		}
	}
}

func (s *LoanService) mapBusinessError(err error, req CreateLoanRequest) error {
	if err == ErrUnsupportedProduct {
		totalDue := int64(float64(req.Principal) * (1 + req.AnnualRate))
		return &BusinessError{
			Message: "Loan parameters result in non-integral weekly payments",
			Code:    "NON_INTEGRAL_WEEKLY_PAYMENT",
			Details: map[string]string{
				"principal":   fmt.Sprintf("%d", req.Principal),
				"annual_rate": fmt.Sprintf("%.2f", req.AnnualRate),
				"total_due":   fmt.Sprintf("%d", totalDue),
				"required":    "Total due amount must be divisible by 50 (weeks)",
				"suggestion":  "Try adjusting the principal or annual rate",
				"calculation": fmt.Sprintf("%d * (1 + %.2f) = %.0f (not divisible by 50)", req.Principal, req.AnnualRate, float64(totalDue)),
			},
		}
	}
	return &BusinessError{
		Message: "Invalid loan parameters",
		Code:    "INVALID_LOAN_PARAMETERS",
		Details: map[string]string{
			"error": err.Error(),
		},
	}
}

func (s *LoanService) convertLoanToResponse(loan *Loan) *LoanResponse {
	var schedule [50]WeekResponse
	for i, w := range loan.Schedule {
		schedule[i] = WeekResponse{
			Index:  w.Index,
			Amount: w.Amount,
			Paid:   w.Paid,
			PaidAt: w.PaidAt,
		}
	}
	return &LoanResponse{
		ID:          loan.ID,
		Principal:   loan.Principal,
		APR:         loan.APR,
		StartDate:   loan.StartDate,
		WeeklyDue:   loan.WeeklyDue,
		Schedule:    schedule,
		PaidCount:   loan.PaidCount,
		Outstanding: loan.Outstanding,
	}
}

func (s *LoanService) generateLoanID() string {
	timestamp := time.Now().UnixNano()
	encoded := base32.StdEncoding.EncodeToString([]byte(strconv.FormatInt(timestamp, 36)))
	return fmt.Sprintf("loan_%s", encoded[:8])
}
