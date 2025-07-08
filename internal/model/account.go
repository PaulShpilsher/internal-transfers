package model

import (
	"errors"

	"github.com/shopspring/decimal"
)

type Account struct {
	AccountID int64
	Balance   decimal.Decimal
}

// Domain-specific error for not found
var ErrAccountNotFound = errors.New("account not found")
var ErrSourceAccountNotFound = errors.New("source account not found")
var ErrDestinationAccountNotFound = errors.New("destination account not found")
var ErrInsufficientFunds = errors.New("insufficient funds")
var ErrAccountIDMustBePositive = errors.New("account id must be a positive number")
var ErrBalanceMustBeNonNegative = errors.New("balance must be non-negative")
var ErrAccountIDAlreadyExists = errors.New("account id already exists")
var ErrSourceAndDestinationMustDiffer = errors.New("source and destination accounts must be different")
var ErrAmountMustBePositive = errors.New("amount must be positive")
var ErrPrecisionTooHigh = errors.New("precision must be 8 or fewer decimal places")
