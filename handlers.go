package main

import (
	"encoding/base32"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// In-memory storage for loans (in production, use a database)
var (
	loans = make(map[string]*Loan)
	mu    sync.RWMutex
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

// OutstandingResponse represents the response for outstanding amount
type OutstandingResponse struct {
	Outstanding int64 `json:"outstanding"`
}

// DelinquencyResponse represents the response for delinquency status
type DelinquencyResponse struct {
	Delinquent   bool `json:"delinquent"`
	Streak       int  `json:"streak"`
	ObservedWeek int  `json:"observed_week"`
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

func createLoanHandler(c echo.Context) error {
	var req CreateLoanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": ErrInvalidRequest.Error()})
	}

	// Default annual rate
	if req.AnnualRate == 0 {
		req.AnnualRate = 0.10
	}

	// Parse start date
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": ErrInvalidRequest.Error()})
	}

	// Generate unique ID (base36 timestamp)
	id := generateLoanID()

	// Create loan
	loan, err := NewLoan(id, req.Principal, req.AnnualRate, startDate)
	if err != nil {
		if err == ErrUnsupportedProduct {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{"error": ErrInvalidRequest.Error()})
	}

	// Store loan
	mu.Lock()
	loans[id] = loan
	mu.Unlock()

	return c.JSON(http.StatusCreated, loan)
}

func getLoanHandler(c echo.Context) error {
	id := c.Param("id")

	mu.RLock()
	loan, exists := loans[id]
	mu.RUnlock()

	if !exists {
		return c.JSON(http.StatusNotFound, map[string]string{"error": ErrLoanNotFound.Error()})
	}

	return c.JSON(http.StatusOK, loan)
}

func payLoanHandler(c echo.Context) error {
	id := c.Param("id")

	var req PaymentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": ErrInvalidRequest.Error()})
	}

	mu.Lock()
	loan, exists := loans[id]
	if !exists {
		mu.Unlock()
		return c.JSON(http.StatusNotFound, map[string]string{"error": ErrLoanNotFound.Error()})
	}

	// Find the first unpaid week index (1-based) before making payment
	firstUnpaidWeek := 0
	for i, week := range loan.Schedule {
		if !week.Paid {
			firstUnpaidWeek = i + 1 // Convert to 1-based index
			break
		}
	}

	now := time.Now().UTC()
	err := loan.MakePayment(req.Amount, now)
	if err != nil {
		mu.Unlock()
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Recompute outstanding after payment
	remainingOutstanding := loan.GetOutstanding()
	mu.Unlock()

	response := PaymentResponse{
		PaidWeek:             firstUnpaidWeek,
		RemainingOutstanding: remainingOutstanding,
	}

	return c.JSON(http.StatusOK, response)
}

func getOutstandingHandler(c echo.Context) error {
	id := c.Param("id")

	mu.RLock()
	loan, exists := loans[id]
	mu.RUnlock()

	if !exists {
		return c.JSON(http.StatusNotFound, map[string]string{"error": ErrLoanNotFound.Error()})
	}

	// Recompute outstanding from schedule to ensure consistency
	outstanding := loan.GetOutstanding()

	response := OutstandingResponse{
		Outstanding: outstanding,
	}

	return c.JSON(http.StatusOK, response)
}

func getDelinquencyHandler(c echo.Context) error {
	id := c.Param("id")

	mu.RLock()
	loan, exists := loans[id]
	mu.RUnlock()

	if !exists {
		return c.JSON(http.StatusNotFound, map[string]string{"error": ErrLoanNotFound.Error()})
	}

	// Check for time override in query parameter
	now := time.Now().UTC()
	if nowParam := c.QueryParam("now"); nowParam != "" {
		if parsedTime, err := time.Parse("2006-01-02", nowParam); err == nil {
			now = parsedTime
		} else if parsedTime, err := time.Parse(time.RFC3339, nowParam); err == nil {
			now = parsedTime
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
