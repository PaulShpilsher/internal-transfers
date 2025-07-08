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
