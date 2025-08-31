package domain

import "errors"

var (
	// ErrInvalidRequest represents invalid request data
	ErrInvalidRequest = errors.New("invalid request")
	
	// ErrUnsupportedProduct represents an unsupported loan product
	ErrUnsupportedProduct = errors.New("weekly equal amount not integral")
	
	// ErrLoanNotFound represents a loan that doesn't exist
	ErrLoanNotFound = errors.New("loan not found")
	
	// ErrAlreadyPaid represents an attempt to pay when all weeks are paid
	ErrAlreadyPaid = errors.New("loan already fully paid")
	
	// ErrWrongAmount represents a payment with incorrect amount
	ErrWrongAmount = errors.New("amount must equal this week's payable")
)
