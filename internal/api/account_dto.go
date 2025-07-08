package api

// Package api provides the data transfer objects (DTOs) for account and transaction operations.

// CreateAccountRequest represents the request body for creating a new account.
type CreateAccountRequest struct {
	AccountID      int64  `json:"account_id" validate:"required,gt=0"`
	InitialBalance string `json:"initial_balance" validate:"required"`
}

// GetAccountResponse represents the response body for retrieving an account.
type GetAccountResponse struct {
	AccountID int64  `json:"account_id"`
	Balance   string `json:"balance"`
}

// CreateTransactionRequest represents the request body for transferring funds between accounts.
type CreateTransactionRequest struct {
	SourceAccountID      int64  `json:"source_account_id" validate:"required,gt=0"`
	DestinationAccountID int64  `json:"destination_account_id" validate:"required,gt=0,nefield=SourceAccountID"`
	Amount               string `json:"amount" validate:"required"`
}
