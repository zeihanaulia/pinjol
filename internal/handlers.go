package internal

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"pinjol/pkg/metrics"
	common "pinjol/pkg/common"
	"pinjol/pkg/domain"
)

// VersionInfo represents version information
type VersionInfo struct {
	Service   string `json:"service"`
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
}

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

func HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func VersionHandler(version, buildTime string) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, VersionInfo{
			Service:   "pinjol",
			Version:   version,
			BuildTime: buildTime,
		})
	}
}

func CreateLoanHandler(c echo.Context, repo domain.LoanRepository, service *domain.LoanService) error {
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

	// Delegate to domain service
	response, err := service.CreateLoanFromRequest(req.Principal, req.AnnualRate, req.StartDate)
	if err != nil {
		return handleDomainError(c, err)
	}

	// Store loan in database
	loan := &domain.Loan{
		ID:          response.ID,
		Principal:   response.Principal,
		APR:         response.APR,
		StartDate:   response.StartDate,
		WeeklyDue:   response.WeeklyDue,
		Schedule:    convertResponseScheduleToDomain(response.Schedule),
		PaidCount:   response.PaidCount,
		Outstanding: response.Outstanding,
	}

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
	metrics.ActiveLoans.Inc()
	metrics.TotalBorrowers.Inc()

	return c.JSON(http.StatusCreated, response)
}

func GetLoanHandler(c echo.Context, repo domain.LoanRepository, service *domain.LoanService) error {
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

func PayLoanHandler(c echo.Context, repo domain.LoanRepository, service *domain.LoanService) error {
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

	// Debug: check loan schedule
	if len(loan.Schedule) == 0 {
		return c.JSON(http.StatusInternalServerError, newSystemError(
			"Loan schedule is empty",
			ErrCodeDatabaseError,
			"get_loan",
			fmt.Errorf("loan %s has empty schedule", id),
		))
	}

	// Delegate to domain service
	now := time.Now().UTC()
	response, err := service.ProcessPaymentFromRequest(loan, req.Amount, now)
	if err != nil {
		return handleDomainError(c, err)
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

	// Record successful payment metrics
	metrics.RecordPaymentReceived(float64(req.Amount), "bank_transfer", "success")
	metrics.RecordRevenue(float64(req.Amount))
	metrics.RecordDatabaseQuery("update", "loans", time.Since(start))

	return c.JSON(http.StatusOK, response)
}

func GetOutstandingHandler(c echo.Context, repo domain.LoanRepository, service *domain.LoanService) error {
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

func GetDelinquencyHandler(c echo.Context, repo domain.LoanRepository, service *domain.LoanService) error {
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

// Helper methods for domain error handling
func handleDomainError(c echo.Context, err error) error {
	switch e := err.(type) {
	case *domain.ValidationError:
		metrics.ValidationErrors.WithLabelValues(e.Field, e.Code).Inc()
		return c.JSON(http.StatusBadRequest, newValidationError(
			e.Field,
			e.Message,
			e.Code,
			e.Details,
		))
	case *domain.BusinessError:
		metrics.RecordBusinessError(e.Code, "domain_operation")
		return c.JSON(http.StatusBadRequest, newBusinessError(
			e.Message,
			e.Code,
			e.Details,
		))
	default:
		return c.JSON(http.StatusInternalServerError, newSystemError(
			"Internal server error",
			ErrCodeDatabaseError,
			"domain_operation",
			err,
		))
	}
}

func convertResponseScheduleToDomain(schedule [50]domain.WeekResponse) [50]domain.Week {
	var result [50]domain.Week
	for i, w := range schedule {
		result[i] = domain.Week{
			Index:  w.Index,
			Amount: w.Amount,
			Paid:   w.Paid,
			PaidAt: w.PaidAt,
		}
	}
	return result
}
