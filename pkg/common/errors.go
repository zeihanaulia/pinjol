package common

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	// Details contains optional structured fields (key/value) used by handlers
	// for providing actionable debugging information to clients.
	Details map[string]string `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err error, code string) ErrorResponse {
	return ErrorResponse{
		Error: err.Error(),
		Code:  code,
	}
}

// LogError logs an error with stack trace
func LogError(logger *slog.Logger, err error, message string) {
	logger.Error(message,
		"error", err,
		"stack", string(debug.Stack()),
	)
}

// HandleError handles HTTP errors with proper logging and response
func HandleError(logger *slog.Logger, c echo.Context, err error, statusCode int, message string) error {
	LogError(logger, err, message)

	response := NewErrorResponse(err, http.StatusText(statusCode))
	return c.JSON(statusCode, response)
}

// RecoverMiddleware creates a panic recovery middleware
func RecoverMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					logger.Error("panic recovered",
						"error", err,
						"stack", string(debug.Stack()),
					)

					response := NewErrorResponse(err, "INTERNAL_ERROR")
					c.JSON(http.StatusInternalServerError, response)
				}
			}()
			return next(c)
		}
	}
}
