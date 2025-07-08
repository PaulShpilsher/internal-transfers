package model

import (
	"github.com/shopspring/decimal"
)

// Account represents a bank account with an ID and balance
// monetary values are represented using decimal.Decimal for precision
type Account struct {
	AccountID int64
	Balance   decimal.Decimal
}
