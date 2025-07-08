package api

type CreateAccountRequest struct {
	AccountID      int64  `json:"account_id" validate:"required,gt=0"`
	InitialBalance string `json:"initial_balance" validate:"required"`
}

type GetAccountResponse struct {
	AccountID int64  `json:"account_id"`
	Balance   string `json:"balance"`
}
