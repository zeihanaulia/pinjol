package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"pinjol/pkg/common"
	"pinjol/pkg/domain"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestServer() *echo.Echo {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	// Initialize schema
	if err := InitDatabase(db); err != nil {
		panic(err)
	}

	// Create repository
	repo := NewSQLiteLoanRepository(db)

	// Create domain service
	service := domain.NewLoanService()

	e := echo.New()
	e.POST("/loans", func(c echo.Context) error { return createLoanHandler(c, repo, service) })
	e.GET("/loans/:id", func(c echo.Context) error { return getLoanHandler(c, repo, service) })
	e.POST("/loans/:id/pay", func(c echo.Context) error { return payLoanHandler(c, repo, service) })
	e.GET("/loans/:id/outstanding", func(c echo.Context) error { return getOutstandingHandler(c, repo, service) })
	e.GET("/loans/:id/delinquent", func(c echo.Context) error { return getDelinquencyHandler(c, repo, service) })
	return e
}

func TestCreateLoanAPI(t *testing.T) {
	e := setupTestServer()

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "valid loan creation",
			body:           `{"principal": 5000000, "annual_rate": 0.10, "start_date": "2025-08-15"}`,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body string) {
				var loan domain.LoanResponse
				if err := json.Unmarshal([]byte(body), &loan); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if loan.WeeklyDue != 110_000 {
					t.Errorf("expected weekly due 110000, got %d", loan.WeeklyDue)
				}
				if loan.Outstanding != 5_500_000 {
					t.Errorf("expected outstanding 5500000, got %d", loan.Outstanding)
				}
			},
		},
		{
			name:           "invalid principal",
			body:           `{"principal": -1000000, "annual_rate": 0.10, "start_date": "2025-08-15"}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				if err := json.Unmarshal([]byte(body), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp["error"] != "Principal amount must be greater than 0" {
					t.Errorf("expected error 'Principal amount must be greater than 0', got %q", resp["error"])
				}
			},
		},
		{
			name:           "unsupported product",
			body:           `{"principal": 5000001, "annual_rate": 0.10, "start_date": "2025-08-15"}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				if err := json.Unmarshal([]byte(body), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp["error"] != "Principal amount exceeds maximum allowed limit" {
					t.Errorf("expected error 'Principal amount exceeds maximum allowed limit', got %q", resp["error"])
				}
			},
		},
		{
			name:           "invalid date format",
			body:           `{"principal": 5000000, "annual_rate": 0.10, "start_date": "invalid-date"}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				if err := json.Unmarshal([]byte(body), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp["error"] != "Start date format is invalid" {
					t.Errorf("expected error 'Start date format is invalid', got %q", resp["error"])
				}
			},
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

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.String())
			}
		})
	}
}

func TestPaymentAPI(t *testing.T) {
	e := setupTestServer()

	// Create a loan first
	createReq := httptest.NewRequest(http.MethodPost, "/loans", 
		strings.NewReader(`{"principal": 5000000, "annual_rate": 0.10, "start_date": "2025-08-15"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var loan domain.LoanResponse
	if err := json.Unmarshal(createRec.Body.Bytes(), &loan); err != nil {
		t.Fatalf("failed to unmarshal loan: %v", err)
	}

	tests := []struct {
		name           string
		loanID         string
		body           string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "valid payment",
			loanID:         loan.ID,
			body:           `{"amount": 110000}`,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp PaymentResponse
				if err := json.Unmarshal([]byte(body), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.PaidWeek != 1 {
					t.Errorf("expected paid week 1, got %d", resp.PaidWeek)
				}
				if resp.RemainingOutstanding != 5_390_000 {
					t.Errorf("expected remaining outstanding 5390000, got %d", resp.RemainingOutstanding)
				}
			},
		},
		{
			name:           "wrong amount",
			loanID:         loan.ID,
			body:           `{"amount": 100000}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				if err := json.Unmarshal([]byte(body), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp["error"] != "Payment amount does not match the required weekly amount" {
					t.Errorf("expected error 'Payment amount does not match the required weekly amount', got %q", resp["error"])
				}
			},
		},
		{
			name:           "loan not found",
			loanID:         "nonexistent",
			body:           `{"amount": 110000}`,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				if err := json.Unmarshal([]byte(body), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp["error"] != "Loan not found" {
					t.Errorf("expected error 'Loan not found', got %q", resp["error"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/loans/"+tt.loanID+"/pay", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.String())
			}
		})
	}
}

func TestOutstandingAPI(t *testing.T) {
	e := setupTestServer()

	// Create a loan first
	createReq := httptest.NewRequest(http.MethodPost, "/loans", 
		strings.NewReader(`{"principal": 5000000, "annual_rate": 0.10, "start_date": "2025-08-15"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var loan domain.Loan
	if err := json.Unmarshal(createRec.Body.Bytes(), &loan); err != nil {
		t.Fatalf("failed to unmarshal loan: %v", err)
	}

	tests := []struct {
		name                string
		loanID              string
		expectedStatus      int
		expectedOutstanding int64
		checkError          func(t *testing.T, body string)
	}{
		{
			name:                "valid outstanding check",
			loanID:              loan.ID,
			expectedStatus:      http.StatusOK,
			expectedOutstanding: 5_500_000,
		},
		{
			name:           "loan not found",
			loanID:         "nonexistent",
			expectedStatus: http.StatusNotFound,
			checkError: func(t *testing.T, body string) {
				var resp common.ErrorResponse
				if err := json.Unmarshal([]byte(body), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Error != "Loan not found" {
					t.Errorf("expected error 'Loan not found', got %q", resp.Error)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/loans/"+tt.loanID+"/outstanding", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp OutstandingResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Outstanding != tt.expectedOutstanding {
					t.Errorf("expected outstanding %d, got %d", tt.expectedOutstanding, resp.Outstanding)
				}
			} else if tt.checkError != nil {
				tt.checkError(t, rec.Body.String())
			}
		})
	}
}

func TestDelinquencyAPI(t *testing.T) {
	e := setupTestServer()

	// Create a loan first
	createReq := httptest.NewRequest(http.MethodPost, "/loans", 
		strings.NewReader(`{"principal": 5000000, "annual_rate": 0.10, "start_date": "2025-08-01"}`))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)

	var loan domain.Loan
	if err := json.Unmarshal(createRec.Body.Bytes(), &loan); err != nil {
		t.Fatalf("failed to unmarshal loan: %v", err)
	}

	tests := []struct {
		name               string
		loanID             string
		nowParam           string
		expectedStatus     int
		expectedDelinquent bool
		expectedObserved   int
		checkError         func(t *testing.T, body string)
	}{
		{
			name:               "delinquent check with time override",
			loanID:             loan.ID,
			nowParam:           "?now=2025-08-15",
			expectedStatus:     http.StatusOK,
			expectedDelinquent: true,
			expectedObserved:   3,
		},
		{
			name:           "loan not found",
			loanID:         "nonexistent",
			nowParam:       "",
			expectedStatus: http.StatusNotFound,
			checkError: func(t *testing.T, body string) {
				var resp common.ErrorResponse
				if err := json.Unmarshal([]byte(body), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Error != "Loan not found" {
					t.Errorf("expected error 'Loan not found', got %q", resp.Error)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/loans/" + tt.loanID + "/delinquent" + tt.nowParam
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp DelinquencyResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Delinquent != tt.expectedDelinquent {
					t.Errorf("expected delinquent %v, got %v", tt.expectedDelinquent, resp.Delinquent)
				}
				if resp.ObservedWeek != tt.expectedObserved {
					t.Errorf("expected observed week %d, got %d", tt.expectedObserved, resp.ObservedWeek)
				}
			} else if tt.checkError != nil {
				tt.checkError(t, rec.Body.String())
			}
		})
	}
}
