package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "cli" {
		runCLI()
		return
	}
	mainServer()
}

func mainServer() {
	port := getEnv("PORT", "8080")
	env := getEnv("APP_ENV", "dev")

	var log *slog.Logger
	if env == "prod" {
		log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	} else {
		log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover(), middleware.RequestID())
	e.Use(LogMiddleware(log))

	e.GET("/healthz", healthHandler)
	e.GET("/version", versionHandler(version, buildTime))

	// Loan endpoints
	e.POST("/loans", createLoanHandler)
	e.GET("/loans/:id", getLoanHandler)
	e.POST("/loans/:id/pay", payLoanHandler)
	e.GET("/loans/:id/outstanding", getOutstandingHandler)
	e.GET("/loans/:id/delinquent", getDelinquencyHandler)

	addr := fmt.Sprintf(":%s", port)
	go func() {
		if err := e.Start(addr); err != nil {
			log.Error("server stopped", "err", err)
		}
	}()
	log.Info("pinjol service started", "env", env, "addr", addr)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown failed", "err", err)
	}
}
