package model

import (
	"github.com/shopspring/decimal"
)

type Account struct {
	AccountID int64
	Balance   decimal.Decimal
}
