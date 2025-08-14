package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// E2) RFC3339 with timezone offset in ?now=
func TestDelinquencyWithTimezone(t *testing.T) {
	e := setupTestServer()

	// Create loan with start=2025-08-01
	createReq := httptest.NewRequest(http.MethodPost, "/loans", 
		strings.NewReader(`{"principal": 5000000, "annual_rate": 0.10, "start_date": "2025-08-01"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var loan Loan
	if err := json.Unmarshal(createRec.Body.Bytes(), &loan); err != nil {
		t.Fatalf("failed to unmarshal loan: %v", err)
	}

	// Test with timezone offset - should normalize to UTC and detect delinquency
	// 2025-08-15T00:00:00+07:00 = 2025-08-14T17:00:00Z, which is still week 2
	// Let's use 2025-08-16T00:00:00+07:00 = 2025-08-15T17:00:00Z, which is week 3
	req := httptest.NewRequest(http.MethodGet, "/loans/"+loan.ID+"/delinquent?now=2025-08-16T00:00:00%2B07:00", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp DelinquencyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Should be delinquent (week 3 with no payments)
	if !resp.Delinquent {
		t.Errorf("expected delinquent true, got false")
	}
}

// E4) Paying when start date is in the future (API test)
func TestFutureStartDatePaymentAPI(t *testing.T) {
	e := setupTestServer()

	// Create loan with future start date
	createReq := httptest.NewRequest(http.MethodPost, "/loans", 
		strings.NewReader(`{"principal": 5000000, "annual_rate": 0.10, "start_date": "2099-01-01"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var loan Loan
	if err := json.Unmarshal(createRec.Body.Bytes(), &loan); err != nil {
		t.Fatalf("failed to unmarshal loan: %v", err)
	}

	// Make payment immediately - should succeed
	payReq := httptest.NewRequest(http.MethodPost, "/loans/"+loan.ID+"/pay", 
		strings.NewReader(`{"amount": 110000}`))
	payReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	payRec := httptest.NewRecorder()
	e.ServeHTTP(payRec, payReq)

	if payRec.Code != http.StatusOK {
		t.Errorf("expected early payment to succeed, got status %d", payRec.Code)
	}

	// Check delinquency - should be false
	delinqReq := httptest.NewRequest(http.MethodGet, "/loans/"+loan.ID+"/delinquent", nil)
	delinqRec := httptest.NewRecorder()
	e.ServeHTTP(delinqRec, delinqReq)

	var delinqResp DelinquencyResponse
	if err := json.Unmarshal(delinqRec.Body.Bytes(), &delinqResp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if delinqResp.Delinquent {
		t.Errorf("expected delinquent false when now < start, got true")
	}
}

// E5) Wrong amount after some payments
func TestWrongAmountAfterPayments(t *testing.T) {
	e := setupTestServer()

	// Create loan
	createReq := httptest.NewRequest(http.MethodPost, "/loans", 
		strings.NewReader(`{"principal": 5000000, "annual_rate": 0.10, "start_date": "2025-08-15"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var loan Loan
	if err := json.Unmarshal(createRec.Body.Bytes(), &loan); err != nil {
		t.Fatalf("failed to unmarshal loan: %v", err)
	}

	// Pay week 1 correctly
	payReq1 := httptest.NewRequest(http.MethodPost, "/loans/"+loan.ID+"/pay", 
		strings.NewReader(`{"amount": 110000}`))
	payReq1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	payRec1 := httptest.NewRecorder()
	e.ServeHTTP(payRec1, payReq1)

	if payRec1.Code != http.StatusOK {
		t.Fatalf("first payment failed with status %d", payRec1.Code)
	}

	// Try wrong amount for week 2
	payReq2 := httptest.NewRequest(http.MethodPost, "/loans/"+loan.ID+"/pay", 
		strings.NewReader(`{"amount": 120000}`))
	payReq2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	payRec2 := httptest.NewRecorder()
	e.ServeHTTP(payRec2, payReq2)

	if payRec2.Code != http.StatusBadRequest {
		t.Errorf("expected wrong amount to fail with 400, got %d", payRec2.Code)
	}

	var errorResp map[string]string
	json.Unmarshal(payRec2.Body.Bytes(), &errorResp)
	if errorResp["error"] != "amount must equal this week's payable" {
		t.Errorf("expected specific error message, got %q", errorResp["error"])
	}

	// Then pay correct amount
	payReq3 := httptest.NewRequest(http.MethodPost, "/loans/"+loan.ID+"/pay", 
		strings.NewReader(`{"amount": 110000}`))
	payReq3.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	payRec3 := httptest.NewRecorder()
	e.ServeHTTP(payRec3, payReq3)

	if payRec3.Code != http.StatusOK {
		t.Errorf("correct payment should succeed, got status %d", payRec3.Code)
	}

	var payResp PaymentResponse
	json.Unmarshal(payRec3.Body.Bytes(), &payResp)
	if payResp.PaidWeek != 2 {
		t.Errorf("expected paid week 2, got %d", payResp.PaidWeek)
	}
}

// E6) Already paid (51st payment)
func TestAlreadyFullyPaid(t *testing.T) {
	e := setupTestServer()

	// Create loan
	createReq := httptest.NewRequest(http.MethodPost, "/loans", 
		strings.NewReader(`{"principal": 5000000, "annual_rate": 0.10, "start_date": "2025-08-15"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var loan Loan
	if err := json.Unmarshal(createRec.Body.Bytes(), &loan); err != nil {
		t.Fatalf("failed to unmarshal loan: %v", err)
	}

	// Pay all 50 weeks
	for i := 0; i < 50; i++ {
		payReq := httptest.NewRequest(http.MethodPost, "/loans/"+loan.ID+"/pay", 
			strings.NewReader(`{"amount": 110000}`))
		payReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		payRec := httptest.NewRecorder()
		e.ServeHTTP(payRec, payReq)

		if payRec.Code != http.StatusOK {
			t.Fatalf("payment %d failed with status %d", i+1, payRec.Code)
		}
	}

	// Try 51st payment
	payReq51 := httptest.NewRequest(http.MethodPost, "/loans/"+loan.ID+"/pay", 
		strings.NewReader(`{"amount": 110000}`))
	payReq51.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	payRec51 := httptest.NewRecorder()
	e.ServeHTTP(payRec51, payReq51)

	if payRec51.Code != http.StatusBadRequest {
		t.Errorf("expected 51st payment to fail with 400, got %d", payRec51.Code)
	}

	var errorResp map[string]string
	json.Unmarshal(payRec51.Body.Bytes(), &errorResp)
	if errorResp["error"] != "loan already fully paid" {
		t.Errorf("expected 'loan already fully paid', got %q", errorResp["error"])
	}
}

// E7) Default annual rate when omitted
func TestDefaultAnnualRate(t *testing.T) {
	e := setupTestServer()

	// Create loan without annual_rate
	createReq := httptest.NewRequest(http.MethodPost, "/loans", 
		strings.NewReader(`{"principal": 5000000, "start_date": "2025-08-01"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected loan creation to succeed, got status %d", createRec.Code)
	}

	var loan Loan
	if err := json.Unmarshal(createRec.Body.Bytes(), &loan); err != nil {
		t.Fatalf("failed to unmarshal loan: %v", err)
	}

	// Should use default 10% rate
	if loan.WeeklyDue != 110_000 {
		t.Errorf("expected weekly due 110000 with default rate, got %d", loan.WeeklyDue)
	}
	if loan.Outstanding != 5_500_000 {
		t.Errorf("expected outstanding 5500000 with default rate, got %d", loan.Outstanding)
	}
}

// E8) Invalid ?now= parsing 
func TestInvalidNowParsing(t *testing.T) {
	e := setupTestServer()

	// Create loan
	createReq := httptest.NewRequest(http.MethodPost, "/loans", 
		strings.NewReader(`{"principal": 5000000, "annual_rate": 0.10, "start_date": "2025-08-01"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var loan Loan
	if err := json.Unmarshal(createRec.Body.Bytes(), &loan); err != nil {
		t.Fatalf("failed to unmarshal loan: %v", err)
	}

	// Test with invalid ?now= parameter - should ignore and use server time
	req := httptest.NewRequest(http.MethodGet, "/loans/"+loan.ID+"/delinquent?now=not-a-date", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 despite invalid now param, got %d", rec.Code)
	}

	var resp DelinquencyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Should have used server time and calculated correct observed week
	if resp.ObservedWeek == 0 {
		t.Errorf("expected valid observed week despite invalid now param, got %d", resp.ObservedWeek)
	}
}

// E9) Invalid create payloads  
func TestInvalidCreatePayloads(t *testing.T) {
	e := setupTestServer()

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "negative principal",
			body:           `{"principal": -1, "annual_rate": 0.10, "start_date": "2025-08-01"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request",
		},
		{
			name:           "bad date format",
			body:           `{"principal": 5000000, "annual_rate": 0.10, "start_date": "2025-13-40"}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/loans", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			var errorResp map[string]string
			json.Unmarshal(rec.Body.Bytes(), &errorResp)
			if errorResp["error"] != tt.expectedError {
				t.Errorf("expected error %q, got %q", tt.expectedError, errorResp["error"])
			}
		})
	}
}
