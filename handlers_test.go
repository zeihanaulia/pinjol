package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func setupTestServer() *echo.Echo {
	e := echo.New()
	e.POST("/loans", createLoanHandler)
	e.GET("/loans/:id", getLoanHandler)
	e.POST("/loans/:id/pay", payLoanHandler)
	e.GET("/loans/:id/outstanding", getOutstandingHandler)
	e.GET("/loans/:id/delinquent", getDelinquencyHandler)
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
				var loan Loan
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
				json.Unmarshal([]byte(body), &resp)
				if resp["error"] != "invalid request" {
					t.Errorf("expected error 'invalid request', got %q", resp["error"])
				}
			},
		},
		{
			name:           "unsupported product",
			body:           `{"principal": 5000001, "annual_rate": 0.10, "start_date": "2025-08-15"}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				json.Unmarshal([]byte(body), &resp)
				if resp["error"] != "weekly equal amount not integral" {
					t.Errorf("expected error 'weekly equal amount not integral', got %q", resp["error"])
				}
			},
		},
		{
			name:           "invalid date format",
			body:           `{"principal": 5000000, "annual_rate": 0.10, "start_date": "invalid-date"}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body string) {
				var resp map[string]string
				json.Unmarshal([]byte(body), &resp)
				if resp["error"] != "invalid request" {
					t.Errorf("expected error 'invalid request', got %q", resp["error"])
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

	var loan Loan
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
				json.Unmarshal([]byte(body), &resp)
				if resp["error"] != "amount must equal this week's payable" {
					t.Errorf("expected error 'amount must equal this week's payable', got %q", resp["error"])
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
				json.Unmarshal([]byte(body), &resp)
				if resp["error"] != "loan not found" {
					t.Errorf("expected error 'loan not found', got %q", resp["error"])
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

	var loan Loan
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
				var resp map[string]string
				json.Unmarshal([]byte(body), &resp)
				if resp["error"] != "loan not found" {
					t.Errorf("expected error 'loan not found', got %q", resp["error"])
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

	var loan Loan
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
				var resp map[string]string
				json.Unmarshal([]byte(body), &resp)
				if resp["error"] != "loan not found" {
					t.Errorf("expected error 'loan not found', got %q", resp["error"])
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
