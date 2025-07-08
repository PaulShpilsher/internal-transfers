package model

import "errors"

// Domain-specific errors
var (
	ErrAccountNotFound                = errors.New("account not found")
	ErrSourceAccountNotFound          = errors.New("source account not found")
	ErrDestinationAccountNotFound     = errors.New("destination account not found")
	ErrInsufficientFunds              = errors.New("insufficient funds")
	ErrAccountIDMustBePositive        = errors.New("account id must be a positive number")
	ErrBalanceMustBeNonNegative       = errors.New("balance must be non-negative")
	ErrAccountIDAlreadyExists         = errors.New("account id already exists")
	ErrSourceAndDestinationMustDiffer = errors.New("source and destination accounts must be different")
	ErrAmountMustBePositive           = errors.New("amount must be positive")
	ErrPrecisionTooHigh               = errors.New("precision must be 8 or fewer decimal places")
)
