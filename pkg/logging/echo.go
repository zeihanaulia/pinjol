package logging

import (
	"time"

	"github.com/labstack/echo/v4"
)

// EchoMiddleware returns an Echo middleware for structured logging
func EchoMiddleware(logger *Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			// Log the request
			logger.RequestLogger(
				c.Request().Method,
				c.Request().URL.Path,
				c.Response().Status,
				time.Since(start).Milliseconds(),
				c.Response().Header().Get(echo.HeaderXRequestID),
			)

			return err
		}
	}
}

// RequestLogger logs HTTP requests with Loki-compatible format
func RequestLogger(logger *Logger) echo.MiddlewareFunc {
	return EchoMiddleware(logger)
}
