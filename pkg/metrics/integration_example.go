// +build ignore

package main

// This file shows how to integrate custom metrics into your Pinjol application
// Add this code to your main.go or handlers to enable business metrics collection

import (
	"net/http"
	"time"

	"pinjol/pkg/metrics"
)

func init() {
	// Initialize metrics during application startup
	metrics.InitMetrics()
}

// Example usage in loan creation handler
func createLoanHandlerWithMetrics(loanRequest LoanRequest) error {
	start := time.Now()

	// Your existing loan creation logic here
	loan, err := createLoan(loanRequest)
	if err != nil {
		metrics.RecordBusinessError("loan_creation_failed", "create_loan")
		return err
	}

	// Record successful loan creation
	metrics.RecordLoanCreated("created", "individual")
	metrics.RecordLoanProcessingTime(time.Since(start))

	// Update active loans gauge
	metrics.ActiveLoans.Inc()

	return nil
}

// Example usage in payment handler
func processPaymentHandlerWithMetrics(paymentRequest PaymentRequest) error {
	start := time.Now()

	// Your existing payment processing logic here
	payment, err := processPayment(paymentRequest)
	if err != nil {
		metrics.RecordBusinessError("payment_failed", "process_payment")
		return err
	}

	// Record successful payment
	metrics.RecordPaymentReceived(payment.Amount, payment.Method, "success")
	metrics.RecordRevenue(payment.Amount)

	// Record database query performance
	metrics.RecordDatabaseQuery("update", "payments", time.Since(start))

	return nil
}

// Example circuit breaker integration
func databaseOperationWithMetrics(operation func() error) error {
	serviceName := "database"

	// Record circuit breaker state (you would get this from your circuit breaker)
	metrics.RecordCircuitBreakerState(serviceName, "closed")

	start := time.Now()
	err := operation()

	if err != nil {
		metrics.CircuitBreakerFailures.WithLabelValues(serviceName).Inc()
		return err
	}

	// Record successful operation
	metrics.RecordDatabaseQuery("operation", "general", time.Since(start))

	return nil
}

// Add metrics endpoint to your Echo instance
func setupMetricsEndpoint(e *echo.Echo) {
	// Expose Prometheus metrics at /metrics
	e.GET("/metrics", echo.WrapHandler(metrics.GetPrometheusHandler()))

	// Custom metrics endpoint with additional info
	e.GET("/metrics/custom", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"application": "pinjol",
			"version":     "1.0.0",
			"metrics_endpoint": "/metrics",
			"custom_metrics": []string{
				"pinjol_loans_created_total",
				"pinjol_payments_received_total",
				"pinjol_revenue_total",
				"pinjol_business_errors_total",
			},
		})
	})
}

// Example middleware for request metrics
func MetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request
			err := next(c)

			// Record request metrics
			status := c.Response().Status
			method := c.Request().Method
			path := c.Request().URL.Path

			// You can add custom request metrics here
			// For example, counting requests by endpoint
			duration := time.Since(start)

			// Log slow requests
			if duration > 2*time.Second {
				// You could add a custom metric for slow requests
				_ = duration // placeholder
			}

			return err
		}
	}
}
