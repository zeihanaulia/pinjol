package metrics

import (
	"net/http"
	"time"

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

// GetPrometheusHandler returns the Prometheus metrics handler
func GetPrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// InitMetrics initializes all metrics (call this during application startup)
func InitMetrics() {
	// This function ensures all metrics are registered
	// Prometheus client library automatically registers metrics when created
}
