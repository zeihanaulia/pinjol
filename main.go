package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"

	"pinjol/pkg/logging"
	"pinjol/pkg/monitoring"
	"pinjol/pkg/profiling"
	"pinjol/pkg/metrics"
	"pinjol/pkg/domain"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "cli":
			runCLI()
			return
		case "db-init":
			runDBInit()
			return
		}
	}
	mainServer()
}

func runDBInit() {
	dbPath := getEnv("DATABASE_PATH", "./pinjol.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := InitDatabase(db); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	log.Printf("database initialized successfully at %s", dbPath)
}

func mainServer() {
	port := getEnv("PORT", "8080")
	env := getEnv("APP_ENV", "dev")
	dbPath := getEnv("DATABASE_PATH", "./pinjol.db")

	// Configure structured logging for Loki
	logConfig := logging.Config{
		Level:       getEnv("LOG_LEVEL", "info"),
		Format:      getEnv("LOG_FORMAT", "json"), // Always use JSON for Loki
		Output:      getEnv("LOG_OUTPUT", "stdout"),
		AddSource:   env == "dev",
		ServiceName: "pinjol",
	}

	logger := logging.NewLogger(logConfig)

	// Initialize metrics
	metrics.InitMetrics()

	// Initialize database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Configure database connection
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Initialize schema
	if err := InitDatabase(db); err != nil {
		logger.Error("failed to initialize database", "err", err)
		os.Exit(1)
	}

	// Create repository
	repo := NewSQLiteLoanRepository(db)

	// Create domain repository adapter
	domainRepo := NewDomainLoanRepositoryAdapter(repo)

	// Create domain service
	service := domain.NewLoanService()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database", "err", err)
		os.Exit(1)
	}

	logger.Info("database initialized successfully", "path", dbPath)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover(), middleware.RequestID())
	e.Use(logging.RequestLogger(logger))

	// Setup profiling endpoints
	profiling.SetupProfiling(e)

	// Setup monitoring endpoints
	monitoring.SetupMonitoring(e, db, version)

	e.GET("/healthz", healthHandler)
	e.GET("/version", versionHandler(version, buildTime))

	// Metrics endpoint
	e.GET("/metrics", echo.WrapHandler(metrics.GetPrometheusHandler()))

	// Loan endpoints with repository injection
	e.POST("/loans", func(c echo.Context) error { return createLoanHandler(c, domainRepo, service) })
	e.GET("/loans/:id", func(c echo.Context) error { return getLoanHandler(c, domainRepo, service) })
	e.POST("/loans/:id/pay", func(c echo.Context) error { return payLoanHandler(c, domainRepo, service) })
	e.GET("/loans/:id/outstanding", func(c echo.Context) error { return getOutstandingHandler(c, domainRepo, service) })
	e.GET("/loans/:id/delinquent", func(c echo.Context) error { return getDelinquencyHandler(c, domainRepo, service) })

	addr := fmt.Sprintf(":%s", port)
	go func() {
		if err := e.Start(addr); err != nil {
			logger.Error("server stopped", "err", err)
		}
	}()
	logger.Info("pinjol service started", "env", env, "addr", addr, "db", dbPath)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown failed", "err", err)
	}
}
