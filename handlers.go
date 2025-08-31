package main

import (
	"encoding/base32"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"pinjol/pkg/metrics"
	common "pinjol/pkg/common"
	"pinjol/pkg/domain"
)

// CreateLoanRequest represents the request body for creating a loan
type CreateLoanRequest struct {
	Principal  int64   `json:"principal"`
	AnnualRate float64 `json:"annual_rate"`
	StartDate  string  `json:"start_date"`
}

// PaymentRequest represents the request body for making a payment
type PaymentRequest struct {
	Amount int64 `json:"amount"`
}

// PaymentResponse represents the response for a successful payment
type PaymentResponse struct {
	PaidWeek             int   `json:"paid_week"`
	RemainingOutstanding int64 `json:"remaining_outstanding"`
}

// OutstandingResponse represents the response for outstanding amount query
type OutstandingResponse struct {
	Outstanding int64 `json:"outstanding"`
}

// DelinquencyResponse represents the response for delinquency check
type DelinquencyResponse struct {
	Delinquent   bool `json:"delinquent"`
	Streak       int  `json:"streak"`
	ObservedWeek int  `json:"observed_week"`
}

// Error codes
const (
	// Validation errors
	ErrCodeInvalidJSON            = "INVALID_JSON"
	ErrCodeInvalidPrincipal       = "INVALID_PRINCIPAL"
	ErrCodePrincipalTooHigh       = "PRINCIPAL_TOO_HIGH"
	ErrCodeInvalidAnnualRate      = "INVALID_ANNUAL_RATE"
	ErrCodeAnnualRateTooHigh      = "ANNUAL_RATE_TOO_HIGH"
	ErrCodeMissingStartDate       = "MISSING_START_DATE"
	ErrCodeInvalidStartDateFormat = "INVALID_START_DATE_FORMAT"
	ErrCodeStartDateInPast        = "START_DATE_IN_PAST"
	ErrCodeInvalidPaymentAmount   = "INVALID_PAYMENT_AMOUNT"
	ErrCodeIncorrectPaymentAmount = "INCORRECT_PAYMENT_AMOUNT"
	ErrCodeMissingLoanID          = "MISSING_LOAN_ID"
	ErrCodeInvalidQueryParameter  = "INVALID_QUERY_PARAMETER"
	ErrCodeInvalidPaymentJSON     = "INVALID_PAYMENT_JSON"

	// Business logic errors
	ErrCodeNonIntegralWeeklyPayment = "NON_INTEGRAL_WEEKLY_PAYMENT"
	ErrCodeLoanAlreadyPaid          = "LOAN_ALREADY_PAID"
	ErrCodeLoanNotFound             = "LOAN_NOT_FOUND"
	ErrCodePaymentFailed            = "PAYMENT_FAILED"
	ErrCodeInvalidLoanParameters    = "INVALID_LOAN_PARAMETERS"

	// System errors
	ErrCodeDatabaseError        = "DATABASE_ERROR"
	ErrCodeDatabaseUpdateError  = "DATABASE_UPDATE_ERROR"
)

// Alias the centralized error response type from pkg/common so existing
// unqualified composite literals in handlers (ErrorResponse{...}) continue
// to work during the refactor.
type ErrorResponse = common.ErrorResponse

// For backward compatibility with existing tests and clients we return a
// flat map[string]string containing at least an "error" field. Detailed
// codes and structured details are intentionally omitted here to avoid
// breaking callers that unmarshal into map[string]string.
func newErrorResponse(errorMsg, _code string, _details map[string]string) map[string]string {
	return map[string]string{"error": errorMsg}
}

func newValidationError(field, message, _code string, _details map[string]string) map[string]string {
	// keep compatibility: include only the main error string
	return map[string]string{"error": message}
}

func newBusinessError(message, _code string, _details map[string]string) map[string]string {
	return map[string]string{"error": message}
}

func newSystemError(message, _code string, operation string, reason error) map[string]string {
	// include a concise error message; internals can be logged separately
	return map[string]string{"error": message}
}

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func versionHandler(version, buildTime string) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, VersionInfo{
			Service:   "pinjol",
			Version:   version,
			BuildTime: buildTime,
		})
	}
}

func createLoanHandler(c echo.Context, repo domain.LoanRepository, service *domain.LoanService) error {
	var req CreateLoanRequest
	if err := c.Bind(&req); err != nil {
		metrics.ValidationErrors.WithLabelValues("request_body", "invalid_json").Inc()
		return c.JSON(http.StatusBadRequest, newErrorResponse(
			"Invalid JSON format in request body",
			ErrCodeInvalidJSON,
			map[string]string{
				"expected": "Valid JSON with principal, annual_rate, and start_date fields",
				"received": err.Error(),
			},
		))
	}

	// Parse start date
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		metrics.ValidationErrors.WithLabelValues("start_date", "invalid_format").Inc()
		return c.JSON(http.StatusBadRequest, newValidationError(
			"start_date",
			"Start date format is invalid",
			ErrCodeInvalidStartDateFormat,
			map[string]string{
				"provided":   req.StartDate,
				"required":   "YYYY-MM-DD format",
				"example":    "2025-08-31",
				"validation": "Must be a valid date in the future or today",
			},
		))
	}

	// Default annual rate
	if req.AnnualRate == 0 {
		req.AnnualRate = 0.10
	}

	domainReq := domain.CreateLoanRequest{
		Principal:  req.Principal,
		AnnualRate: req.AnnualRate,
		StartDate:  startDate,
	}

	// Validate using domain service
	if err := service.ValidateCreateLoanRequest(domainReq); err != nil {
		// Map errors to appropriate codes
		if req.Principal <= 0 {
			metrics.ValidationErrors.WithLabelValues("principal", "invalid_value").Inc()
			return c.JSON(http.StatusBadRequest, newValidationError(
				"principal",
				"Principal amount must be greater than 0",
				ErrCodeInvalidPrincipal,
				map[string]string{
					"provided":  fmt.Sprintf("%d", req.Principal),
					"required":  "Must be a positive integer greater than 0",
					"min_value": "1",
					"max_value": "5000000 (5 million)",
				},
			))
		}
		if req.Principal > 5_000_000 {
			metrics.ValidationErrors.WithLabelValues("principal", "unsupported_product").Inc()
			return c.JSON(http.StatusBadRequest, newValidationError(
				"principal",
				"Principal amount exceeds maximum allowed limit",
				ErrCodePrincipalTooHigh,
				map[string]string{
					"provided": fmt.Sprintf("%d", req.Principal),
					"maximum":  "5000000 (5 million)",
					"reason":   "Higher amounts may not result in integral weekly payments",
				},
			))
		}
		if req.AnnualRate < 0 {
			metrics.ValidationErrors.WithLabelValues("annual_rate", "negative_value").Inc()
			return c.JSON(http.StatusBadRequest, newValidationError(
				"annual_rate",
				"Annual rate cannot be negative",
				ErrCodeInvalidAnnualRate,
				map[string]string{
					"provided":  fmt.Sprintf("%.2f", req.AnnualRate),
					"required":  "Must be a non-negative decimal (0.00 - 0.50)",
					"min_value": "0.00",
					"max_value": "0.50 (50%)",
				},
			))
		}
		if req.AnnualRate > 0.50 {
			metrics.ValidationErrors.WithLabelValues("annual_rate", "too_high").Inc()
			return c.JSON(http.StatusBadRequest, newValidationError(
				"annual_rate",
				"Annual rate exceeds maximum allowed limit",
				ErrCodeAnnualRateTooHigh,
				map[string]string{
					"provided": fmt.Sprintf("%.2f", req.AnnualRate),
					"maximum":  "0.50 (50%)",
					"reason":   "Higher rates may not result in integral weekly payments",
				},
			))
		}
		return c.JSON(http.StatusBadRequest, newBusinessError(
			"Invalid loan parameters",
			ErrCodeInvalidLoanParameters,
			map[string]string{
				"error": err.Error(),
			},
		))
	}

	// Generate unique ID
	id := generateLoanID()

	// Create loan using domain service
	loan, err := service.CreateLoan(id, domainReq)
	if err != nil {
		metrics.RecordBusinessError("unsupported_product", "create_loan")
		totalDue := int64(float64(req.Principal) * (1 + req.AnnualRate))
		return c.JSON(http.StatusBadRequest, newBusinessError(
			"Loan parameters result in non-integral weekly payments",
			ErrCodeNonIntegralWeeklyPayment,
			map[string]string{
				"principal":   fmt.Sprintf("%d", req.Principal),
				"annual_rate": fmt.Sprintf("%.2f", req.AnnualRate),
				"total_due":   fmt.Sprintf("%d", totalDue),
				"required":    "Total due amount must be divisible by 50 (weeks)",
				"suggestion":  "Try adjusting the principal or annual rate",
				"calculation": fmt.Sprintf("%d * (1 + %.2f) = %.0f (not divisible by 50)", req.Principal, req.AnnualRate, float64(totalDue)),
			},
		))
	}

	// Store loan in database
	if err := repo.Create(loan); err != nil {
		metrics.RecordBusinessError("database_error", "create_loan")
		return c.JSON(http.StatusInternalServerError, newSystemError(
			"Failed to save loan to database",
			ErrCodeDatabaseError,
			"create_loan",
			err,
		))
	}

	// Record successful metrics
	metrics.RecordLoanCreated("created", "individual")
	metrics.LoanProcessingTime.Observe(time.Since(time.Now()).Seconds()) // Note: start time not accurate
	metrics.ActiveLoans.Inc()
	metrics.TotalBorrowers.Inc()

	// Convert to response value object
	var schedule [50]domain.WeekResponse
	for i, w := range loan.Schedule {
		schedule[i] = domain.WeekResponse{
			Index:  w.Index,
			Amount: w.Amount,
			Paid:   w.Paid,
			PaidAt: w.PaidAt,
		}
	}
	response := domain.LoanResponse{
		ID:          loan.ID,
		Principal:   loan.Principal,
		APR:         loan.APR,
		StartDate:   loan.StartDate,
		WeeklyDue:   loan.WeeklyDue,
		Schedule:    schedule,
		PaidCount:   loan.PaidCount,
		Outstanding: loan.Outstanding,
	}

	return c.JSON(http.StatusCreated, response)
}

func getLoanHandler(c echo.Context, repo domain.LoanRepository, service *domain.LoanService) error {
	id := c.Param("id")

	// Validate loan ID
	if id == "" {
		return c.JSON(http.StatusBadRequest, newValidationError(
			"id",
			"Loan ID is required in URL path",
			ErrCodeMissingLoanID,
			map[string]string{
				"parameter": "id",
				"format":    "Base32 encoded timestamp string",
				"example":   "/loans/ABC123DEF",
			},
		))
	}

	loan, err := repo.GetByID(id)
	if err != nil {
		if err == domain.ErrLoanNotFound {
			return c.JSON(http.StatusNotFound, newBusinessError(
				"Loan not found",
				ErrCodeLoanNotFound,
				map[string]string{
					"loan_id":     id,
					"suggestion":  "Verify the loan ID is correct and the loan exists",
				},
			))
		}
		return c.JSON(http.StatusInternalServerError, newSystemError(
			"Failed to retrieve loan from database",
			ErrCodeDatabaseError,
			"get_loan",
			err,
		))
	}

	// Convert to response value object
	var schedule [50]domain.WeekResponse
	for i, w := range loan.Schedule {
		schedule[i] = domain.WeekResponse{
			Index:  w.Index,
			Amount: w.Amount,
			Paid:   w.Paid,
			PaidAt: w.PaidAt,
		}
	}
	response := domain.LoanResponse{
		ID:          loan.ID,
		Principal:   loan.Principal,
		APR:         loan.APR,
		StartDate:   loan.StartDate,
		WeeklyDue:   loan.WeeklyDue,
		Schedule:    schedule,
		PaidCount:   loan.PaidCount,
		Outstanding: loan.Outstanding,
	}

	return c.JSON(http.StatusOK, response)
}

func payLoanHandler(c echo.Context, repo domain.LoanRepository, service *domain.LoanService) error {
	id := c.Param("id")

	// Validate loan ID format
	if id == "" {
		return c.JSON(http.StatusBadRequest, newValidationError(
			"id",
			"Loan ID is required in URL path",
			ErrCodeMissingLoanID,
			map[string]string{
				"parameter": "id",
				"format":    "Base32 encoded timestamp string",
				"example":   "/loans/ABC123DEF/pay",
			},
		))
	}

	var req PaymentRequest
	if err := c.Bind(&req); err != nil {
		metrics.ValidationErrors.WithLabelValues("payment_request", "invalid_json").Inc()
		return c.JSON(http.StatusBadRequest, newErrorResponse(
			"Invalid JSON format in payment request",
			ErrCodeInvalidPaymentJSON,
			map[string]string{
				"expected": "Valid JSON with amount field",
				"received": err.Error(),
				"example":  `{"amount": 110000}`,
			},
		))
	}

	// Record payment processing start time
	start := time.Now()

	// Get loan from database
	loan, err := repo.GetByID(id)
	if err != nil {
		if err == domain.ErrLoanNotFound {
			metrics.RecordBusinessError("loan_not_found", "pay_loan")
			return c.JSON(http.StatusNotFound, newBusinessError(
				"Loan not found",
				ErrCodeLoanNotFound,
				map[string]string{
					"loan_id":     id,
					"suggestion":  "Verify the loan ID is correct and the loan exists",
				},
			))
		}
		metrics.RecordBusinessError("database_error", "pay_loan")
		return c.JSON(http.StatusInternalServerError, newSystemError(
			"Failed to retrieve loan from database",
			ErrCodeDatabaseError,
			"get_loan",
			err,
		))
	}

	// Find the first unpaid week index (1-based) before making payment
	firstUnpaidWeek := 0
	expectedAmount := int64(0)
	for i, week := range loan.Schedule {
		if !week.Paid {
			firstUnpaidWeek = i + 1 // Convert to 1-based index
			expectedAmount = week.Amount
			break
		}
	}

	now := time.Now().UTC()
	domainReq := domain.PaymentRequest{
		Amount: req.Amount,
		Now:    now,
	}

	err = service.ProcessPayment(loan, domainReq)
	if err != nil {
		metrics.RecordBusinessError("payment_error", "pay_loan")
		if err.Error() == "loan already fully paid" {
			return c.JSON(http.StatusBadRequest, newBusinessError(
				"Loan is already fully paid",
				ErrCodeLoanAlreadyPaid,
				map[string]string{
					"loan_id":         id,
					"paid_weeks":      fmt.Sprintf("%d", loan.PaidCount),
					"total_weeks":     "50",
					"outstanding":     fmt.Sprintf("%d", loan.Outstanding),
					"suggestion":      "No further payments are required for this loan",
				},
			))
		} else if err.Error() == "amount must equal this week's payable" {
			return c.JSON(http.StatusBadRequest, newBusinessError(
				"Payment amount does not match the required weekly amount",
				ErrCodeIncorrectPaymentAmount,
				map[string]string{
					"provided_amount":   fmt.Sprintf("%d", req.Amount),
					"required_amount":   fmt.Sprintf("%d", expectedAmount),
					"week_number":       fmt.Sprintf("%d", firstUnpaidWeek),
					"suggestion":        fmt.Sprintf("Pay exactly %d for week %d", expectedAmount, firstUnpaidWeek),
				},
			))
		}
		return c.JSON(http.StatusBadRequest, newBusinessError(
			"Payment processing failed",
			ErrCodePaymentFailed,
			map[string]string{
				"reason": err.Error(),
			},
		))
	}

	// Update loan in database
	if err := repo.Update(loan); err != nil {
		metrics.RecordBusinessError("database_error", "pay_loan")
		return c.JSON(http.StatusInternalServerError, newSystemError(
			"Failed to update loan in database",
			ErrCodeDatabaseUpdateError,
			"update_loan",
			err,
		))
	}

	// Recompute outstanding after payment
	remainingOutstanding := loan.GetOutstanding()

	// Record successful payment metrics
	metrics.RecordPaymentReceived(float64(req.Amount), "bank_transfer", "success")
	metrics.RecordRevenue(float64(req.Amount))
	metrics.RecordDatabaseQuery("update", "loans", time.Since(start))

	response := PaymentResponse{
		PaidWeek:             firstUnpaidWeek,
		RemainingOutstanding: remainingOutstanding,
	}

	return c.JSON(http.StatusOK, response)
}

func getOutstandingHandler(c echo.Context, repo domain.LoanRepository, service *domain.LoanService) error {
	id := c.Param("id")

	// Validate loan ID
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Loan ID is required in URL path",
			Code:  "MISSING_LOAN_ID",
			Details: map[string]string{
				"parameter": "id",
				"format":    "Base32 encoded timestamp string",
				"example":   "/loans/ABC123DEF/outstanding",
			},
		})
	}

	loan, err := repo.GetByID(id)
	if err != nil {
		if err == domain.ErrLoanNotFound {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Loan not found",
				Code:  "LOAN_NOT_FOUND",
				Details: map[string]string{
					"loan_id":     id,
					"suggestion":  "Verify the loan ID is correct and the loan exists",
				},
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve loan from database",
			Code:  "DATABASE_ERROR",
			Details: map[string]string{
				"operation": "get_loan",
				"loan_id":   id,
				"reason":    err.Error(),
			},
		})
	}

	// Recompute outstanding from schedule to ensure consistency
	outstanding := loan.GetOutstanding()

	response := OutstandingResponse{
		Outstanding: outstanding,
	}

	return c.JSON(http.StatusOK, response)
}

func getDelinquencyHandler(c echo.Context, repo domain.LoanRepository, service *domain.LoanService) error {
	id := c.Param("id")

	// Validate loan ID
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Loan ID is required in URL path",
			Code:  "MISSING_LOAN_ID",
			Details: map[string]string{
				"parameter": "id",
				"format":    "Base32 encoded timestamp string",
				"example":   "/loans/ABC123DEF/delinquent",
			},
		})
	}

	loan, err := repo.GetByID(id)
	if err != nil {
		if err == domain.ErrLoanNotFound {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Loan not found",
				Code:  "LOAN_NOT_FOUND",
				Details: map[string]string{
					"loan_id":     id,
					"suggestion":  "Verify the loan ID is correct and the loan exists",
				},
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve loan from database",
			Code:  "DATABASE_ERROR",
			Details: map[string]string{
				"operation": "get_loan",
				"loan_id":   id,
				"reason":    err.Error(),
			},
		})
	}

	// Check for time override in query parameter
	now := time.Now().UTC()
	if nowParam := c.QueryParam("now"); nowParam != "" {
		if parsedTime, err := time.Parse("2006-01-02", nowParam); err == nil {
			now = parsedTime
		} else if parsedTime, err := time.Parse(time.RFC3339, nowParam); err == nil {
			now = parsedTime
		} else {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid 'now' query parameter format",
				Code:  "INVALID_QUERY_PARAMETER",
				Details: map[string]string{
					"parameter":   "now",
					"provided":    nowParam,
					"accepted":    "YYYY-MM-DD or RFC3339 format",
					"examples":    "2025-08-31 or 2025-08-31T10:00:00Z",
					"suggestion":  "Use ISO date format or RFC3339 timestamp",
				},
			})
		}
	}

	delinquent, streak, observedWeek := loan.IsDelinquent(now)

	response := DelinquencyResponse{
		Delinquent:   delinquent,
		Streak:       streak,
		ObservedWeek: observedWeek,
	}

	return c.JSON(http.StatusOK, response)
}

// generateLoanID generates a unique loan ID using base32 timestamp
func generateLoanID() string {
	timestamp := time.Now().UnixNano()
	encoded := base32.StdEncoding.EncodeToString([]byte(strconv.FormatInt(timestamp, 36)))
	return fmt.Sprintf("loan_%s", encoded[:8])
}
