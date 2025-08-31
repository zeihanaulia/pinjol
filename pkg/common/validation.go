package common

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidatePhone validates Indonesian phone number format
func ValidatePhone(phone string) error {
	// Remove spaces, dashes, etc.
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "+", "")

	// Indonesian phone number patterns
	patterns := []string{
		`^62[0-9]{8,12}$`,  // +62xxxxxxxxx
		`^0[0-9]{8,12}$`,    // 0xxxxxxxxx
		`^[0-9]{8,12}$`,     // xxxxxxxxx
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, cleaned); matched {
			return nil
		}
	}

	return fmt.Errorf("invalid phone number format")
}

// ValidateAmount validates monetary amount
func ValidateAmount(amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if amount > 1_000_000_000 { // 1 billion limit
		return fmt.Errorf("amount exceeds maximum limit")
	}
	return nil
}

// ValidateDate validates date string format
func ValidateDate(dateStr string) error {
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}
	return nil
}

// SanitizeString removes potentially harmful characters
func SanitizeString(input string) string {
	// Remove null bytes and other control characters
	return strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, input)
}
