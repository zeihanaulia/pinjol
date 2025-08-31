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

// ProcessPaymentFromRequest handles the complete payment process from HTTP request
func (s *LoanService) ProcessPaymentFromRequest(loan *Loan, amount int64, now time.Time) (*PaymentResponse, error) {
	req := PaymentRequest{
		Amount: amount,
		Now:    now,
	}

	// Validate payment request
	if err := s.ValidatePaymentRequest(req); err != nil {
		return nil, s.mapPaymentValidationError(err, req)
	}

	// Find the first unpaid week before making payment
	firstUnpaidWeek := 0
	expectedAmount := int64(0)
	for i, week := range loan.Schedule {
		if !week.Paid {
			firstUnpaidWeek = i + 1 // Convert to 1-based index
			expectedAmount = week.Amount
			break
		}
	}

	// Debug: check if we found unpaid week
	if firstUnpaidWeek == 0 {
		// Count total weeks and paid weeks for debugging
		totalWeeks := len(loan.Schedule)
		paidWeeks := 0
		for _, week := range loan.Schedule {
			if week.Paid {
				paidWeeks++
			}
		}
		// If no unpaid week found, this might indicate all weeks are paid or schedule is empty
		return nil, &BusinessError{
			Message: "No unpaid weeks found",
			Code:    "NO_UNPAID_WEEKS",
			Details: map[string]string{
				"loan_id":     loan.ID,
				"total_weeks": fmt.Sprintf("%d", totalWeeks),
				"paid_weeks":  fmt.Sprintf("%d", paidWeeks),
			},
		}
	}

	// Process the payment
	err := s.ProcessPayment(loan, req)
	if err != nil {
		return nil, s.mapPaymentBusinessError(err, loan, req, firstUnpaidWeek, expectedAmount)
	}

	// Recompute outstanding after payment
	remainingOutstanding := loan.GetOutstanding()

	return &PaymentResponse{
		PaidWeek:             firstUnpaidWeek,
		RemainingOutstanding: remainingOutstanding,
	}, nil
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

func (s *LoanService) mapPaymentValidationError(err error, req PaymentRequest) error {
	errMsg := err.Error()
	if errMsg == "payment amount must be greater than 0" {
		return &ValidationError{
			Field:   "amount",
			Message: "Payment amount must be greater than 0",
			Code:    "INVALID_PAYMENT_AMOUNT",
			Details: map[string]string{
				"provided":  fmt.Sprintf("%d", req.Amount),
				"required":  "Must be a positive integer greater than 0",
				"min_value": "1",
			},
		}
	}
	return &ValidationError{
		Field:   "amount",
		Message: errMsg,
		Code:    "INVALID_PAYMENT_AMOUNT",
		Details: map[string]string{
			"error": errMsg,
		},
	}
}

func (s *LoanService) mapPaymentBusinessError(err error, loan *Loan, req PaymentRequest, firstUnpaidWeek int, expectedAmount int64) error {
	errMsg := err.Error()
	switch errMsg {
	case "loan already fully paid":
		return &BusinessError{
			Message: "Loan is already fully paid",
			Code:    "LOAN_ALREADY_PAID",
			Details: map[string]string{
				"loan_id":         loan.ID,
				"paid_weeks":      fmt.Sprintf("%d", loan.PaidCount),
				"total_weeks":     "50",
				"outstanding":     fmt.Sprintf("%d", loan.Outstanding),
				"suggestion":      "No further payments are required for this loan",
			},
		}
	case "amount must equal this week's payable":
		return &BusinessError{
			Message: "Payment amount does not match the required weekly amount",
			Code:    "INCORRECT_PAYMENT_AMOUNT",
			Details: map[string]string{
				"provided_amount":   fmt.Sprintf("%d", req.Amount),
				"required_amount":   fmt.Sprintf("%d", expectedAmount),
				"week_number":       fmt.Sprintf("%d", firstUnpaidWeek),
				"suggestion":        fmt.Sprintf("Pay exactly %d for week %d", expectedAmount, firstUnpaidWeek),
			},
		}
	default:
		return &BusinessError{
			Message: "Payment processing failed",
			Code:    "PAYMENT_FAILED",
			Details: map[string]string{
				"reason": errMsg,
			},
		}
	}
}

func (s *LoanService) generateLoanID() string {
	timestamp := time.Now().UnixNano()
	encoded := base32.StdEncoding.EncodeToString([]byte(strconv.FormatInt(timestamp, 36)))
	return fmt.Sprintf("loan_%s", encoded[:8])
}
