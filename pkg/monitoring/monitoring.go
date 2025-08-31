package monitoring

import (
	"context"
	"database/sql"
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
)

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Uptime    time.Duration     `json:"uptime"`
	Services  map[string]string `json:"services"`
}

// Metrics represents system and application metrics
type Metrics struct {
	Timestamp   time.Time         `json:"timestamp"`
	GoVersion   string            `json:"go_version"`
	Goroutines  int               `json:"goroutines"`
	Memory      MemoryStats       `json:"memory"`
	Database    DatabaseStats     `json:"database"`
	Application ApplicationStats  `json:"application"`
}

// MemoryStats represents Go runtime memory statistics
type MemoryStats struct {
	Alloc        uint64 `json:"alloc_bytes"`
	TotalAlloc   uint64 `json:"total_alloc_bytes"`
	Sys          uint64 `json:"sys_bytes"`
	Lookups      uint64 `json:"lookups"`
	Mallocs      uint64 `json:"mallocs"`
	Frees        uint64 `json:"frees"`
	HeapAlloc    uint64 `json:"heap_alloc_bytes"`
	HeapSys      uint64 `json:"heap_sys_bytes"`
	HeapIdle     uint64 `json:"heap_idle_bytes"`
	HeapInuse    uint64 `json:"heap_inuse_bytes"`
	HeapReleased uint64 `json:"heap_released_bytes"`
	HeapObjects  uint64 `json:"heap_objects"`
	StackInuse   uint64 `json:"stack_inuse_bytes"`
	StackSys     uint64 `json:"stack_sys_bytes"`
	GCSys        uint64 `json:"gc_sys_bytes"`
	NextGC       uint64 `json:"next_gc_bytes"`
	LastGC       uint64 `json:"last_gc_timestamp"`
	NumGC        uint32 `json:"num_gc"`
}

// DatabaseStats represents database connection statistics
type DatabaseStats struct {
	MaxOpenConnections int `json:"max_open_connections"`
	OpenConnections    int `json:"open_connections"`
	InUse              int `json:"in_use"`
	Idle               int `json:"idle"`
	WaitCount          int64 `json:"wait_count"`
	WaitDuration       time.Duration `json:"wait_duration"`
	MaxIdleClosed      int64 `json:"max_idle_closed"`
	MaxLifetimeClosed  int64 `json:"max_lifetime_closed"`
}

// ApplicationStats represents application-specific statistics
type ApplicationStats struct {
	TotalLoans      int64 `json:"total_loans"`
	ActiveLoans     int64 `json:"active_loans"`
	OverdueLoans    int64 `json:"overdue_loans"`
	TotalPayments   int64 `json:"total_payments"`
	TotalRevenue    float64 `json:"total_revenue"`
}

// Monitor holds monitoring configuration and state
type Monitor struct {
	db           *sql.DB
	startTime    time.Time
	version      string
	loanRepo     interface{} // We'll use interface{} to avoid circular imports
}

// NewMonitor creates a new monitoring instance
func NewMonitor(db *sql.DB, version string) *Monitor {
	return &Monitor{
		db:        db,
		startTime: time.Now(),
		version:   version,
	}
}

// SetLoanRepository sets the loan repository for application metrics
func (m *Monitor) SetLoanRepository(repo interface{}) {
	m.loanRepo = repo
}

// HealthHandler returns a health check endpoint handler
func (m *Monitor) HealthHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		status := &HealthStatus{
			Status:    "healthy",
			Timestamp: time.Now(),
			Version:   m.version,
			Uptime:    time.Since(m.startTime),
			Services:  make(map[string]string),
		}

		// Check database health
		if err := m.db.PingContext(c.Request().Context()); err != nil {
			status.Status = "unhealthy"
			status.Services["database"] = "unhealthy"
		} else {
			status.Services["database"] = "healthy"
		}

		statusCode := http.StatusOK
		if status.Status == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		}

		return c.JSON(statusCode, status)
	}
}

// MetricsHandler returns a metrics endpoint handler
func (m *Monitor) MetricsHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		metrics := &Metrics{
			Timestamp: time.Now(),
			GoVersion: runtime.Version(),
			Goroutines: runtime.NumGoroutine(),
			Memory:    m.getMemoryStats(),
			Database:  m.getDatabaseStats(),
		}

		// Get application stats if repository is available
		if m.loanRepo != nil {
			metrics.Application = m.getApplicationStats(c.Request().Context())
		}

		return c.JSON(http.StatusOK, metrics)
	}
}

// ReadinessHandler returns a readiness check endpoint handler
func (m *Monitor) ReadinessHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Check if database is ready
		if err := m.db.PingContext(c.Request().Context()); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status": "not ready",
				"reason": "database not available",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"status": "ready",
		})
	}
}

// getMemoryStats collects Go runtime memory statistics
func (m *Monitor) getMemoryStats() MemoryStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return MemoryStats{
		Alloc:        memStats.Alloc,
		TotalAlloc:   memStats.TotalAlloc,
		Sys:          memStats.Sys,
		Lookups:      memStats.Lookups,
		Mallocs:      memStats.Mallocs,
		Frees:        memStats.Frees,
		HeapAlloc:    memStats.HeapAlloc,
		HeapSys:      memStats.HeapSys,
		HeapIdle:     memStats.HeapIdle,
		HeapInuse:    memStats.HeapInuse,
		HeapReleased: memStats.HeapReleased,
		HeapObjects:  memStats.HeapObjects,
		StackInuse:   memStats.StackInuse,
		StackSys:     memStats.StackSys,
		GCSys:        memStats.GCSys,
		NextGC:       memStats.NextGC,
		LastGC:       memStats.LastGC,
		NumGC:        memStats.NumGC,
	}
}

// getDatabaseStats collects database connection statistics
func (m *Monitor) getDatabaseStats() DatabaseStats {
	stats := m.db.Stats()

	return DatabaseStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}
}

// getApplicationStats collects application-specific statistics
func (m *Monitor) getApplicationStats(ctx context.Context) ApplicationStats {
	// This would typically query the loan repository
	// For now, return placeholder values
	return ApplicationStats{
		TotalLoans:     0,
		ActiveLoans:    0,
		OverdueLoans:   0,
		TotalPayments:  0,
		TotalRevenue:   0.0,
	}
}

// SetupMonitoring sets up monitoring endpoints on the Echo instance
func SetupMonitoring(e *echo.Echo, db *sql.DB, version string) *Monitor {
	monitor := NewMonitor(db, version)

	// Health check endpoints
	e.GET("/healthz", monitor.HealthHandler())
	e.GET("/readyz", monitor.ReadinessHandler())

	// Metrics endpoint
	e.GET("/metrics", monitor.MetricsHandler())

	return monitor
}
