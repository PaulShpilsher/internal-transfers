package api

type CreateAccountRequest struct {
	AccountID      int64  `json:"account_id" validate:"required,gt=0"`
	InitialBalance string `json:"initial_balance" validate:"required"`
}

type GetAccountResponse struct {
	AccountID int64  `json:"account_id"`
	Balance   string `json:"balance"`
}

type CreateTransactionRequest struct {
	SourceAccountID      int64  `json:"source_account_id" validate:"required,gt=0"`
	DestinationAccountID int64  `json:"destination_account_id" validate:"required,gt=0,nefield=SourceAccountID"`
	Amount               string `json:"amount" validate:"required"`
}
