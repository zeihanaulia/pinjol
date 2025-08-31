package metrics

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Business Metrics for Pinjol Application
var (
	// Loan Metrics
	LoansCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pinjol_loans_created_total",
		Help: "Total number of loans created",
	}, []string{"status", "borrower_type"})

	LoansApproved = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pinjol_loans_approved_total",
		Help: "Total number of loans approved",
	}, []string{"amount_range", "duration"})

	LoansRejected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pinjol_loans_rejected_total",
		Help: "Total number of loans rejected",
	})

	ActiveLoans = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pinjol_loans_active",
		Help: "Number of currently active loans",
	})

	OverdueLoans = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pinjol_loans_overdue",
		Help: "Number of overdue loans",
	})

	// Payment Metrics
	PaymentsReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pinjol_payments_received_total",
		Help: "Total number of payments received",
	}, []string{"payment_method", "status"})

	PaymentAmount = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "pinjol_payment_amount",
		Help: "Payment amount distribution",
		Buckets: prometheus.LinearBuckets(100, 100, 20), // 100, 200, 300, ... 2000
	})

	// Revenue Metrics
	TotalRevenue = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pinjol_revenue_total",
		Help: "Total revenue generated",
	})

	MonthlyRevenue = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pinjol_revenue_monthly",
		Help: "Current month revenue",
	})

	// User/Borrower Metrics
	TotalBorrowers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pinjol_borrowers_total",
		Help: "Total number of registered borrowers",
	})

	ActiveBorrowers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pinjol_borrowers_active",
		Help: "Number of active borrowers (with loans)",
	})

	// Performance Metrics
	LoanProcessingTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "pinjol_loan_processing_duration_seconds",
		Help: "Time taken to process loan applications",
		Buckets: prometheus.DefBuckets,
	})

	DatabaseQueryTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "pinjol_db_query_duration_seconds",
		Help: "Database query execution time",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation", "table"})

	// Error Metrics
	BusinessErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pinjol_business_errors_total",
		Help: "Total number of business logic errors",
	}, []string{"error_type", "operation"})

	ValidationErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pinjol_validation_errors_total",
		Help: "Total number of validation errors",
	}, []string{"field", "error_type"})

	// Circuit Breaker Metrics
	CircuitBreakerState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pinjol_circuit_breaker_state",
		Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
	}, []string{"service", "state"})

	CircuitBreakerFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pinjol_circuit_breaker_failures_total",
		Help: "Total number of circuit breaker failures",
	}, []string{"service"})

	// Cache Metrics (if using cache)
	CacheHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pinjol_cache_hits_total",
		Help: "Total number of cache hits",
	}, []string{"cache_type"})

	CacheMisses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pinjol_cache_misses_total",
		Help: "Total number of cache misses",
	}, []string{"cache_type"})

	// External API Metrics
	ExternalAPICalls = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pinjol_external_api_calls_total",
		Help: "Total number of external API calls",
	}, []string{"service", "method", "status"})

	ExternalAPILatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "pinjol_external_api_duration_seconds",
		Help: "External API call duration",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "method"})

	// RED Metrics for HTTP Requests (Rate, Errors, Duration)
	HTTPRequestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "endpoint", "status_code"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "HTTP request duration in seconds",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
	}, []string{"method", "endpoint"})

	HTTPRequestsInFlight = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Number of HTTP requests currently being processed",
	}, []string{"method", "endpoint"})
)

// RecordLoanCreated records a loan creation event
func RecordLoanCreated(status, borrowerType string) {
	LoansCreated.WithLabelValues(status, borrowerType).Inc()
}

// RecordLoanApproved records a loan approval event
func RecordLoanApproved(amount float64, duration int) {
	amountRange := getAmountRange(amount)
	durationStr := getDurationRange(duration)
	LoansApproved.WithLabelValues(amountRange, durationStr).Inc()
}

// RecordPaymentReceived records a payment event
func RecordPaymentReceived(amount float64, method, status string) {
	PaymentsReceived.WithLabelValues(method, status).Inc()
	PaymentAmount.Observe(amount)
}

// RecordRevenue records revenue
func RecordRevenue(amount float64) {
	TotalRevenue.Add(amount)
}

// RecordBusinessError records a business logic error
func RecordBusinessError(errorType, operation string) {
	BusinessErrors.WithLabelValues(errorType, operation).Inc()
}

// RecordDatabaseQuery records database query performance
func RecordDatabaseQuery(operation, table string, duration time.Duration) {
	DatabaseQueryTime.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordCircuitBreakerState records circuit breaker state change
func RecordCircuitBreakerState(service, state string) {
	var stateValue float64
	switch state {
	case "closed":
		stateValue = 0
	case "open":
		stateValue = 1
	case "half-open":
		stateValue = 2
	}
	CircuitBreakerState.WithLabelValues(service, state).Set(stateValue)
}

// Helper functions for categorization
func getAmountRange(amount float64) string {
	switch {
	case amount < 1000:
		return "0-999"
	case amount < 5000:
		return "1000-4999"
	case amount < 10000:
		return "5000-9999"
	case amount < 50000:
		return "10000-49999"
	default:
		return "50000+"
	}
}

func getDurationRange(duration int) string {
	switch {
	case duration <= 3:
		return "1-3_months"
	case duration <= 6:
		return "4-6_months"
	case duration <= 12:
		return "7-12_months"
	case duration <= 24:
		return "13-24_months"
	default:
		return "24+_months"
	}
}

// RecordHTTPRequest records an HTTP request with method, endpoint, status code, and duration
func RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	HTTPRequestTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordHTTPRequestStart records the start of an HTTP request (increments in-flight counter)
func RecordHTTPRequestStart(method, endpoint string) {
	HTTPRequestsInFlight.WithLabelValues(method, endpoint).Inc()
}

// RecordHTTPRequestEnd records the end of an HTTP request (decrements in-flight counter)
func RecordHTTPRequestEnd(method, endpoint string) {
	HTTPRequestsInFlight.WithLabelValues(method, endpoint).Dec()
}

// HTTPMetricsMiddleware returns an Echo middleware that records RED metrics
func HTTPMetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			method := c.Request().Method
			path := c.Request().URL.Path

			// Normalize path for metrics (remove IDs and parameters)
			endpoint := normalizeEndpoint(path)

			// Record request start
			RecordHTTPRequestStart(method, endpoint)

			// Process request
			err := next(c)

			// Record request end
			duration := time.Since(start)
			statusCode := getStatusCode(c, err)
			RecordHTTPRequest(method, endpoint, statusCode, duration)
			RecordHTTPRequestEnd(method, endpoint)

			return err
		}
	}
}

// normalizeEndpoint normalizes the endpoint path for metrics
func normalizeEndpoint(path string) string {
	// Replace IDs with placeholders
	if strings.HasPrefix(path, "/loans/") {
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			// Replace loan ID with placeholder
			if parts[2] != "" && parts[2] != " " {
				parts[2] = ":id"
			}
			path = strings.Join(parts, "/")
		}
	}
	return path
}

// getStatusCode extracts the status code from the response or error
func getStatusCode(c echo.Context, err error) string {
	if err != nil {
		// Check if it's an HTTP error
		if he, ok := err.(*echo.HTTPError); ok {
			return fmt.Sprintf("%d", he.Code)
		}
		// Default to 500 for other errors
		return "500"
	}
	return fmt.Sprintf("%d", c.Response().Status)
}

// GetPrometheusHandler returns the Prometheus metrics handler
func GetPrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// InitMetrics initializes all metrics (call this during application startup)
func InitMetrics() {
	// This function ensures all metrics are registered
	// Prometheus client library automatically registers metrics when created
}
